package config

import (
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/http_api/logger"
	"github.com/fsnotify/fsnotify"
	"github.com/scorpiotzh/toolib"
)

var (
	Cfg CfgServer
	log = logger.NewLogger("config", logger.LevelDebug)
)

func InitCfg(configFilePath string) error {
	if configFilePath == "" {
		configFilePath = "../config/config.yaml"
	}
	log.Info("read from config：", configFilePath)
	if err := toolib.UnmarshalYamlFile(configFilePath, &Cfg); err != nil {
		return fmt.Errorf("UnmarshalYamlFile err:%s", err.Error())
	}
	log.Info("config file：", toolib.JsonString(Cfg))
	return nil
}

func AddCfgFileWatcher(configFilePath string) (*fsnotify.Watcher, error) {
	if configFilePath == "" {
		configFilePath = "../config/config.yaml"
	}
	return toolib.AddFileWatcher(configFilePath, func() {
		log.Info("update config file：", configFilePath)
		if err := toolib.UnmarshalYamlFile(configFilePath, &Cfg); err != nil {
			log.Error("UnmarshalYamlFile err:", err.Error())
		}
		log.Info("new config file：", toolib.JsonString(Cfg))
	})
}

type CfgServer struct {
	Server struct {
		Net            common.DasNetType `json:"net" yaml:"net"`
		HttpServerAddr string            `json:"http_server_addr" yaml:"http_server_addr"`
		FixCharset     bool              `json:"fix_charset" yaml:"fix_charset"`
		NotExit        bool              `json:"not_exit" yaml:"not_exit"`
	} `json:"server" yaml:"server"`
	Notice struct {
		WebhookLarkErr string `json:"webhook_lark_err" yaml:"webhook_lark_err"`
		SentryDsn      string `json:"sentry_dsn" yaml:"sentry_dsn"`
	} `json:"notice" yaml:"notice"`
	Chain struct {
		CkbUrl             string `json:"ckb_url" yaml:"ckb_url"`
		IndexUrl           string `json:"index_url" yaml:"index_url"`
		CurrentBlockNumber uint64 `json:"current_block_number" yaml:"current_block_number"`
		ConfirmNum         uint64 `json:"confirm_num" yaml:"confirm_num"`
		ConcurrencyNum     uint64 `json:"concurrency_num" yaml:"concurrency_num"`
	} `json:"chain" yaml:"chain"`
	Origins  []string `json:"origins"`
	Snapshot struct {
		Open           bool   `json:"open" yaml:"open"`
		ConcurrencyNum uint64 `json:"concurrency_num" yaml:"concurrency_num"`
		ConfirmNum     uint64 `json:"confirm_num" yaml:"confirm_num"`
		SnapshotNum    int    `json:"snapshot_num" yaml:"snapshot_num"`
	} `json:"snapshot" yaml:"snapshot"`
	DB struct {
		Mysql DbMysql `json:"mysql" yaml:"mysql"`
	} `json:"db" yaml:"db"`
	Cache struct {
		Redis struct {
			Addr     string `json:"addr" yaml:"addr"`
			Password string `json:"password" yaml:"password"`
			DbNum    int    `json:"db_num" yaml:"db_num"`
		} `json:"redis" yaml:"redis"`
	} `json:"cache" yaml:"cache"`
}

type DbMysql struct {
	Addr        string `json:"addr" yaml:"addr"`
	User        string `json:"user" yaml:"user"`
	Password    string `json:"password" yaml:"password"`
	DbName      string `json:"db_name" yaml:"db_name"`
	MaxOpenConn int    `json:"max_open_conn" yaml:"max_open_conn"`
	MaxIdleConn int    `json:"max_idle_conn" yaml:"max_idle_conn"`
}

func PriceToCKB(price, quote, years uint64) (total uint64) {
	log.Info("PriceToCKB:", price, quote, years)
	if price > quote {
		total = price / quote * common.OneCkb * years
	} else {
		total = price * common.OneCkb / quote * years
	}
	log.Info("PriceToCKB:", price, quote, total)
	return
}
