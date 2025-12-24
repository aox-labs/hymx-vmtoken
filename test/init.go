package test

import (
	"context"
	"fmt"
	"github.com/aox-labs/hymx-vmtoken/basic"
	"github.com/aox-labs/hymx-vmtoken/crosschain"
	"github.com/aox-labs/hymx-vmtoken/schema"
	"github.com/everFinance/goether"
	"github.com/gin-gonic/gin"
	"github.com/hymatrix/hymx/node"
	nodeSchema "github.com/hymatrix/hymx/node/schema"
	hymxSchema "github.com/hymatrix/hymx/schema"
	"github.com/hymatrix/hymx/sdk"
	"github.com/hymatrix/hymx/server"
	registrySchema "github.com/hymatrix/hymx/vmm/core/registry/schema"
	"github.com/inconshreveable/log15"
	"github.com/permadao/goar"
	goarSchema "github.com/permadao/goar/schema"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"io"
)

var (
	// config
	port, ginMode, redisURL, arweaveURL, hymxURL, prvKey string

	nodeInfo *nodeSchema.Info
	signer   *goether.Signer
	bundler  *goar.Bundler
	hysdk    *sdk.SDK

	tAXPid, registryPid string
)

func init() {
	fmt.Println("init...")
	// log disabled
	log15.Root().SetHandler(log15.DiscardHandler())

	loadConfig()
	cleanDB()
	initNode()
	initTokenAndRegistry()
	fmt.Println("Initialization successful. Starting test...")
	fmt.Printf("\n\n")
}

func loadConfig() {
	viper.SetConfigFile("./config.yaml")
	viper.SetConfigType("yaml")
	viper.ReadInConfig()

	port = viper.GetString("port")
	ginMode = viper.GetString("ginMode")
	redisURL = viper.GetString("redisURL")
	arweaveURL = viper.GetString("arweaveURL")
	hymxURL = viper.GetString("hymxURL")
	prvKey = viper.GetString("prvKey")

	signer, _ = goether.NewSigner(prvKey)
	bundler, _ = goar.NewBundler(signer)
	hysdk = sdk.NewFromBundler(hymxURL, bundler)

	nodeInfo = &nodeSchema.Info{
		Protocol:    hymxSchema.DataProtocol,
		Variant:     hymxSchema.Variant,
		NodeVersion: nodeSchema.NodeVersion,
		JoinNetwork: viper.GetBool("joinNetwork"),
		Node: registrySchema.Node{
			AccId: bundler.Address,
			Name:  viper.GetString("nodeName"),
			Desc:  viper.GetString("nodeDesc"),
			URL:   viper.GetString("nodeURL"),
		},
	}
}

func cleanDB() {
	redisOpt, err := redis.ParseURL(redisURL)
	if err != nil {
		panic(err)
	}
	rdb := redis.NewClient(redisOpt)
	err = rdb.FlushDB(context.Background()).Err()
	if err != nil {
		panic(err)
	}
}

func initNode() {
	gin.SetMode(ginMode)
	// log disabled
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard

	node := node.New(bundler, redisURL, arweaveURL, hymxURL, nodeInfo, nil)
	s := server.New(node, nil)

	// mount your vm here.....
	s.Mount(schema.VmTokenBasicModuleFormat, basic.Spawn)
	s.Mount(schema.VmTokenCrossChainModuleFormat, crosschain.Spawn)

	s.Run(port)
}

func initTokenAndRegistry() {
	res, err := hysdk.SpawnAndWait(
		TAxModule,
		hysdk.GetAddress(),
		[]goarSchema.Tag{})
	if err != nil {
		panic(err)
	}
	tAXPid = res.Id

	res, err = hysdk.SpawnAndWait(
		RegistryModule,
		hysdk.GetAddress(),
		[]goarSchema.Tag{
			{Name: "Token-Pid", Value: tAXPid},
			{Name: "Name", Value: nodeInfo.Node.Name},
			{Name: "Desc", Value: nodeInfo.Node.Desc},
			{Name: "URL", Value: nodeInfo.Node.URL},
		})
	if err != nil {
		panic(err)
	}
	registryPid = res.Id

	fmt.Println("tAX and registry initialization complete")
	fmt.Println("tAXPid", tAXPid)
	fmt.Println("registryPid", registryPid)
	fmt.Println()
}
