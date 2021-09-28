package main

import (
	"fmt"
	"github.com/polynetwork/polygon-relayer/cmd"
	"github.com/polynetwork/polygon-relayer/config"
	"github.com/polynetwork/polygon-relayer/db"
	"github.com/polynetwork/polygon-relayer/log"
	"github.com/urfave/cli"
	"os"
	"runtime"
)

var (
	logLevelFlag = cli.UintFlag{
		Name:  "loglevel",
		Usage: "Set the log level to `<level>` (0~6). 0:Trace 1:Debug 2:Info 3:Warn 4:Error 5:Fatal 6:MaxLevel",
		Value: 1,
	}

	configPathFlag = cli.StringFlag{
		Name:  "cliconfig",
		Usage: "Server config file `<path>`",
		Value: config.DEFAULT_CONFIG_FILE_NAME,
	}
	txFlag = cli.StringFlag{
		Name:  "tx",
		Usage: "tx",
	}
)

func setupApp() *cli.App {
	app := cli.NewApp()
	app.Usage = "polygon db tools"
	app.Action = startServer
	app.Version = "1.0.0"
	app.Copyright = "Copyright in 2019 The Ontology Authors"
	app.Flags = []cli.Flag{
		logLevelFlag,
		configPathFlag,
	}
	app.Commands = []cli.Command{}
	app.Before = func(context *cli.Context) error {
		runtime.GOMAXPROCS(runtime.NumCPU())
		return nil
	}
	return app
}

func startServer(ctx *cli.Context) {
	tx := ctx.GlobalString(cmd.GetFlagName(txFlag))
	log.Info("qqqqqqtx:", tx)
	if tx != "" {
		ConfigPath := ctx.GlobalString(cmd.GetFlagName(cmd.ConfigPathFlag))
		servConfig := config.NewServiceConfig(ConfigPath)
		if servConfig == nil {
			log.Errorf("startServer - create config failed!")
			return
		}
		var boltDB *db.BoltDB
		var err error
		if servConfig.BoltDbPath != "" {
			boltDB, err = db.NewBoltDB(servConfig.BoltDbPath)
			if err != nil {
				log.Fatalf("db.NewWaitingDB error:%s", err)
				return
			}
		}
		err = boltDB.DeleteBridgeTransactions(tx)
		if err != nil {
			log.Errorf("DeleteBridgeTransactions %s err %w", tx, err)
		} else {
			log.Info("success DeleteBridgeTransactions %s", tx)
		}
	}

}

func main() {
	if err := setupApp().Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
