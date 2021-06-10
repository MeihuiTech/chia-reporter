package main

import (
	"github.com/urfave/cli"
	"log"
	"os"
)

var vSyncCommand = cli.Command{
	Name: "sync",
	Usage: "read chia block records and calc total/daily won blocks group by farmer",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Value: "",
			Usage: "set config file(json format)",
		},
	},
	Action: func(c *cli.Context) error {
		return SyncAction(c)
	},
}

var vExportCommand = cli.Command{
	Name: "export",
	Usage: "export farmer data to a server",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Value: "",
			Usage: "set config file(json format)",
		},
	},
	Action: func(c *cli.Context) error {
		return ExportAction(c)
	},
}

func main() {
	local := []cli.Command{
		vSyncCommand,
		vExportCommand,
	}

	app := &cli.App{
		Name:     "chia-blocks-sync",
		Version:  "1.0",
		Commands: local,
		Flags:    []cli.Flag{},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
		return
	}
}
