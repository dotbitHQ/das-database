package timer

import (
	"context"
	"das_database/config"
	"das_database/dao"
	"encoding/json"
	"fmt"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"sync"
	"testing"
)

func getInit() (*dao.DbDao, error) {
	if err := config.InitCfg("../config/config.yaml"); err != nil {
		return nil, fmt.Errorf("InitCfg err: %s", err)
	}
	cfgMysql := config.Cfg.DB.Mysql
	db, err := dao.NewGormDataBase(cfgMysql.Addr, cfgMysql.User, cfgMysql.Password, cfgMysql.DbName, cfgMysql.MaxOpenConn, cfgMysql.MaxIdleConn)
	if err != nil {
		return nil, fmt.Errorf("NewGormDataBase err:%s", err.Error())
	}
	dbDao, err := dao.Initialize(db, cfgMysql.LogMode)
	if err != nil {
		return nil, fmt.Errorf("Initialize err:%s ", err.Error())
	}
	return dbDao, nil
}

func TestGetTokenInfo(t *testing.T) {
	dbDao, err := getInit()
	if err != nil {
		t.Fatal(err)
	}
	parserTimer := NewParserTimer(ParserTimerParam{
		DbDao: dbDao,
		Ctx:   context.Background(),
		Wg:    &sync.WaitGroup{},
	})
	if err = parserTimer.Run(); err != nil {
		t.Fatal(err)
	}
	tokenInfo := GetTokenPriceInfo(TokenIdCkb)
	fmt.Println(toolib.JsonString(tokenInfo))
}

func TestGetTokenPrice(t *testing.T) {
	ids := []string{
		"ethereum",
	}
	list, err := GetTokenPrice(ids)
	b, _ := json.Marshal(list)
	fmt.Println(string(b), err)
	if len(list) > 0 && list[0].Cny.Cmp(decimal.Zero) == 1 && list[0].Price.Cmp(decimal.Zero) == 1 {
		dec := list[0].Price.DivRound(list[0].Cny, 4)
		fmt.Println(dec.String())
	}
}

func TestGetTokenPriceBinance(t *testing.T) {
	fmt.Println(GetTokenPriceNew(nil))
}
