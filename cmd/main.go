package main

import (
	"context"
	"das_database/block_parser"
	"das_database/config"
	"das_database/dao"
	"das_database/http_server"
	"das_database/snapshot"
	"das_database/timer"
	"fmt"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/nervosnetwork/ckb-sdk-go/rpc"
	"github.com/scorpiotzh/mylog"
	"github.com/scorpiotzh/toolib"
	"github.com/urfave/cli/v2"
	"os"
	"sync"
	"time"
)

var (
	log               = mylog.NewLogger("main", mylog.LevelDebug)
	exit              = make(chan struct{})
	ctxServer, cancel = context.WithCancel(context.Background())
	wgServer          = sync.WaitGroup{}
)

func main() {
	log.Debugf("server start：")
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Load configuration from `FILE`",
			},
		},
		Action: runServer,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func runServer(ctx *cli.Context) error {
	// config
	configFilePath := ctx.String("config")
	if err := config.InitCfg(configFilePath); err != nil {
		return err
	}

	// config update
	watcher, err := config.AddCfgFileWatcher(configFilePath)
	if err != nil {
		return err
	}

	// db
	cfgMysql := config.Cfg.DB.Mysql
	db, err := dao.NewGormDataBase(cfgMysql.Addr, cfgMysql.User, cfgMysql.Password, cfgMysql.DbName, cfgMysql.MaxOpenConn, cfgMysql.MaxIdleConn)
	if err != nil {
		return fmt.Errorf("NewGormDataBase err:%s", err.Error())
	}
	dbDao, err := dao.Initialize(db)
	if err != nil {
		return fmt.Errorf("Initialize err:%s ", err.Error())
	}
	log.Info("db ok")

	// ckb node
	ckbClient, err := rpc.DialWithIndexer(config.Cfg.Chain.CkbUrl, config.Cfg.Chain.IndexUrl)
	if err != nil {
		return fmt.Errorf("DialWithIndexer err: %s", err.Error())
	}
	log.Info("ckb node ok")

	// das core
	env := core.InitEnv(config.Cfg.Server.Net)
	opts := []core.DasCoreOption{
		core.WithClient(ckbClient),
		core.WithDasContractArgs(env.ContractArgs),
		core.WithDasContractCodeHash(env.ContractCodeHash),
		core.WithDasNetType(config.Cfg.Server.Net),
		core.WithTHQCodeHash(env.THQCodeHash),
	}
	dc := core.NewDasCore(ctxServer, &wgServer, opts...)
	dc.InitDasContract(env.MapContract)
	if err := dc.InitDasConfigCell(); err != nil {
		return fmt.Errorf("InitDasConfigCell err: %s", err.Error())
	}
	if err := dc.InitDasSoScript(); err != nil {
		return fmt.Errorf("InitDasSoScript err: %s", err.Error())
	}
	dc.RunAsyncDasContract(time.Minute * 5)   // contract outpoint
	dc.RunAsyncDasConfigCell(time.Minute * 3) // config cell outpoint
	dc.RunAsyncDasSoScript(time.Minute * 7)   // so
	log.Info("contract ok")

	// block parser
	bp, err := block_parser.NewBlockParser(block_parser.ParamsBlockParser{
		DasCore:            dc,
		CurrentBlockNumber: config.Cfg.Chain.CurrentBlockNumber,
		DbDao:              dbDao,
		ConcurrencyNum:     config.Cfg.Chain.ConcurrencyNum,
		ConfirmNum:         config.Cfg.Chain.ConfirmNum,
		Ctx:                ctxServer,
		Cancel:             cancel,
		Wg:                 &wgServer,
	})
	if err != nil {
		return fmt.Errorf("NewBlockParser err: %s", err.Error())
	}
	bp.RunParser()

	// timer
	parserTimer := timer.ParserTimer{
		DbDao:   dbDao,
		Ctx:     ctxServer,
		Wg:      &wgServer,
		DasCore: dc,
	}
	parserTimer.RunUpdateTokenPrice()
	parserTimer.RunFixCharset()
	log.Info("parser timer ok")

	// snapshot
	toolSnapshot := snapshot.ToolSnapshot{
		Ctx:            ctxServer,
		Cancel:         cancel,
		Wg:             &wgServer,
		DbDao:          dbDao,
		DasCore:        dc,
		ConcurrencyNum: config.Cfg.Snapshot.ConcurrencyNum,
		ConfirmNum:     config.Cfg.Snapshot.ConfirmNum,
	}
	if err := toolSnapshot.Run(config.Cfg.Snapshot.Open); err != nil {
		return fmt.Errorf("toolSnapshot.Run err: %s", err.Error())
	}

	// cache
	red, err := toolib.NewRedisClient(config.Cfg.Cache.Redis.Addr, config.Cfg.Cache.Redis.Password, config.Cfg.Cache.Redis.DbNum)
	if err != nil {
		log.Error("NewRedisClient err:", err.Error())
	}

	// http server
	hs, err := http_server.Initialize(http_server.HttpServerParams{
		Address: config.Cfg.Server.HttpServerAddr,
		DbDao:   dbDao,
		Ctx:     ctxServer,
		DasCore: dc,
		Bp:      bp,
		Red:     red,
	})
	if err != nil {
		return fmt.Errorf("http server Initialize err:%s", err.Error())
	}
	hs.Run()

	// quit monitor
	toolib.ExitMonitoring(func(sig os.Signal) {
		log.Warn("ExitMonitoring:", sig.String())
		if watcher != nil {
			log.Warn("close watcher ... ")
			_ = watcher.Close()
		}
		hs.Shutdown()
		cancel()
		wgServer.Wait()
		exit <- struct{}{}
	})

	<-exit
	log.Warn("success exit server. bye bye!")
	return nil
}
