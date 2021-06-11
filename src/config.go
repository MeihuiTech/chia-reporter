package main

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
	"io/ioutil"
)

type Config struct {
	Dsn string
	RpcHost string
	FullNodeRpcPort uint
	HarvesterRpcPort uint
	WalletRpcPort uint
	WalletId uint
	PrivateCert string
	PrivateKey string
	CaCert string
	SyncBlocks bool
	IgnoreGormNotFoundError bool
}

func NewConfig(ctx *cli.Context) (*Config, error) {
	var config Config
	viper.SetConfigType("json") // REQUIRED if the config file does not have the extension in the name
	configFile := ctx.String("config")
	if configFile != "" {
		configBuffer, err := ioutil.ReadFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("error read config file: %v", err)
		} else {
			err = viper.ReadConfig(bytes.NewBuffer(configBuffer))
			if err != nil {
				return nil, fmt.Errorf("error load config: %v", err)
			}
		}
	} else {
		viper.SetConfigName("config")                 // name of config file (without extension)
		viper.AddConfigPath("/etc/chia-reporter/")  // path to look for the config file in
		viper.AddConfigPath("$HOME/.chia-reporter") // call multiple times to add many search paths
		viper.AddConfigPath(".")                      // optionally look for config in the working directory
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				return nil, fmt.Errorf("Config file not found \n")
			} else {
				return nil, fmt.Errorf("Config file was found but another error was produced: %s \n", err)
			}
		}
	}

	config.IgnoreGormNotFoundError = viper.GetBool("ignore_gorm_not_found_error")
	config.RpcHost = viper.GetString("rpc_host")
	config.FullNodeRpcPort = viper.GetUint("full_node_rpc_port")
	config.WalletRpcPort = viper.GetUint("wallet_rpc_port")
	config.HarvesterRpcPort = viper.GetUint("harvester_rpc_port")
	config.WalletId = viper.GetUint("wallet_id")
	config.PrivateCert = viper.GetString("private_cert")
	config.PrivateKey = viper.GetString("private_key")
	config.CaCert = viper.GetString("ca_cert")
	config.SyncBlocks = viper.GetBool("sync_blocks")
	config.Dsn = viper.GetString("dsn")

	if config.RpcHost == "" {
		return nil, fmt.Errorf("error config: rpc_host can not be empty")
	}
	if config.FullNodeRpcPort == 0 {
		return nil, fmt.Errorf("error config: full_node_rpc_port can not be empty")
	}
	if config.WalletRpcPort == 0 {
		return nil, fmt.Errorf("error config: wallet_rpc_port can not be empty")
	}
	if config.HarvesterRpcPort == 0 {
		return nil, fmt.Errorf("error config: harvester_rpc_port can not be empty")
	}
	if config.PrivateCert == "" {
		return nil, fmt.Errorf("error config: private_cert can not be empty")
	}
	if config.PrivateKey == "" {
		return nil, fmt.Errorf("error config: private_key can not be empty")
	}
	if config.CaCert == "" {
		return nil, fmt.Errorf("error config: ca_cert can not be empty")
	}
	if config.Dsn == "" {
		return nil, fmt.Errorf("error config: dsn can not be empty")
	}
	if config.WalletId == 0 {
		config.WalletId = 1
	}

	return &config, nil
}
