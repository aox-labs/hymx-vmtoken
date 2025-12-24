package main

import (
	"github.com/aox-labs/hymx-vmtoken/vmtoken/basic"
	"github.com/aox-labs/hymx-vmtoken/vmtoken/crosschain"
	"github.com/aox-labs/hymx-vmtoken/vmtoken/schema"
	"github.com/gin-gonic/gin"
	"github.com/hymatrix/hymx/common"
	"github.com/hymatrix/hymx/node"
	"github.com/hymatrix/hymx/server"
	"github.com/inconshreveable/log15"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
	"os"
	"os/signal"
	"syscall"
)

var log = common.NewLog(Name + "-" + Version)

func main() {
	cli.VersionFlag = flagVersion

	app := &cli.App{
		Name:     Name,
		Version:  Version,
		Flags:    flags,
		Commands: cmds,
		Action:   action,
	}

	if err := app.Run(os.Args); err != nil {
		log.Error("run server failed", "err", err)
	}
}

func action(c *cli.Context) error {
	// viper configuration
	// notice: viper only for yaml file, cmd flags use urfave
	configPath := c.String("config")
	if configPath == "" {
		configPath = DefaultConfig
	}
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return run(c)
}

func run(c *cli.Context) (err error) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	// node config
	port, ginMode, redisURL, arweaveURL, hymxURL, bundler, nodeInfo, err := LoadNodeConfig()
	if err != nil {
		return err
	}

	gin.SetMode(ginMode)
	if ginMode == "release" {
		log15.Root().SetHandler(log15.LvlFilterHandler(log15.LvlInfo, log15.StderrHandler))
	}

	node := node.New(bundler, redisURL, arweaveURL, hymxURL, nodeInfo, nil)
	// chainkit := chainkit.New(node, LoadChainkitConfig())

	s := server.New(node, nil)
	// mount vm token variants
	s.Mount(schema.VmTokenBasicModuleFormat, basic.Spawn)
	s.Mount(schema.VmTokenCrossChainModuleFormat, crosschain.Spawn)

	s.Run(port)

	log.Info("server is running", "wallet", bundler.Address, "port", port)

	<-signals
	s.Close()

	return nil
}
