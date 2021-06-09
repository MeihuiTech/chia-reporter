package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
)

var vRunCmd = cli.Command{
	Name: "run",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "config",
			Value: "",
			Usage: "set config file(json format)",
		},
	},
	Action: func(c *cli.Context) error {
		return runCmd(c)
	},
}

func runCmd(ctx *cli.Context) error {
	viper.SetConfigType("json") // REQUIRED if the config file does not have the extension in the name
	config := ctx.String("config")
	if config != "" {
		configBuffer, err := ioutil.ReadFile(config)
		if err != nil {
			return fmt.Errorf("error read config file: %v", err)
		} else {
			err = viper.ReadConfig(bytes.NewBuffer(configBuffer))
			if err != nil {
				return fmt.Errorf("error load config: %v", err)
			}
		}
	} else {
		viper.SetConfigName("config") // name of config file (without extension)
		viper.AddConfigPath("/etc/chia-block-sync/")   // path to look for the config file in
		viper.AddConfigPath("$HOME/.chia-block-sync")  // call multiple times to add many search paths
		viper.AddConfigPath(".")               // optionally look for config in the working directory
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				panic(fmt.Errorf("Config file not found \n"))
			} else {
				panic(fmt.Errorf("Config file was found but another error was produced: %s \n", err))
			}
		}
	}


	dsn := viper.GetString("dsn")
	if dsn == "" {
		return fmt.Errorf("error config: dsn can not be empty")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to open db connection: %v \n", err)
	}

	err = db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").AutoMigrate(&ChiaBlockRecord{})
	if err != nil {
		return fmt.Errorf("error migrate db: %v", err)
	}

	err = db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='矿工累计出块数量'").AutoMigrate(&ChiaTotalFarmerBlocks{})
	if err != nil {
		return fmt.Errorf("error migrate db: %v", err)
	}

	err = db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='矿工每天出块数量'").AutoMigrate(&ChiaDailyFarmerBlocks{})
	if err != nil {
		return fmt.Errorf("error migrate db: %v", err)
	}

	err = db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='区块同步高度'").AutoMigrate(&ChiaBlockSyncHeight{})
	if err != nil {
		return fmt.Errorf("error migrate db: %v", err)
	}

	host := viper.GetString("rpc_host")
	port := viper.GetUint("rpc_port")
	privateCert := viper.GetString("private_cert")
	privateKey := viper.GetString("private_key")
	caCert := viper.GetString("ca_cert")
	syncBlocks := viper.GetBool("sync_blocks")

	if host == "" {
		return fmt.Errorf("error config: rpc_host can not be empty")
	}
	if port == 0 {
		return fmt.Errorf("error config: rpc_port can not be empty")
	}
	if privateCert == "" {
		return fmt.Errorf("error config: private_cert can not be empty")
	}
	if privateKey == "" {
		return fmt.Errorf("error config: private_key can not be empty")
	}
	if caCert == "" {
		return fmt.Errorf("error config: ca_cert can not be empty")
	}

	syncChannel := make(chan int)
	signalChannel := make(chan os.Signal)
	defer close(signalChannel)
	defer close(syncChannel)
	go SyncBlocks(context.Background(), syncChannel, host, port, privateCert, privateKey, caCert, syncBlocks, db)

	signal.Notify(signalChannel, os.Interrupt)
	select {
	case sig := <-signalChannel:
		fmt.Printf("Got %s signal. Aborting...\n", sig)
	case code := <-syncChannel:
		fmt.Printf("sync goroutine exit with code: %d\n", code)
	}
	return nil
}
func main() {
	local := []cli.Command{
		vRunCmd,
	}

	app := &cli.App{
		Name:     "chia blocks sync",
		Usage:    "chia-block-sync run",
		Version:  "1.0",
		Commands: local,
		Flags:    []cli.Flag{},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
		return
	}
}
