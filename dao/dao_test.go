package dao

import (
	"das_database/config"
	"fmt"
	"github.com/shopspring/decimal"
	"testing"
)

func getInit() (*DbDao, error) {
	if err := config.InitCfg("../config/config.yaml"); err != nil {
		return nil, fmt.Errorf("InitCfg err: %s", err)
	}
	cfgMysql := config.Cfg.DB.Mysql
	db, err := NewGormDataBase(cfgMysql.Addr, cfgMysql.User, cfgMysql.Password, cfgMysql.DbName, cfgMysql.MaxOpenConn, cfgMysql.MaxIdleConn)
	if err != nil {
		return nil, fmt.Errorf("NewGormDataBase err:%s", err.Error())
	}
	dbDao, err := Initialize(db)
	if err != nil {
		return nil, fmt.Errorf("Initialize err:%s ", err.Error())
	}
	return dbDao, nil
}

func TestInsertOrUpdate(t *testing.T) {
	dbDao, err := getInit()
	if err != nil {
		t.Fatal(err)
	}
	incomeCellInfo := TableIncomeCellInfo{
		BlockNumber:    5718189,
		Action:         "create_income",
		Outpoint:       "0xe9116d651c371662b6e29e2102422e23f90656b8619df82c48b782ff4db43a37_2",
		Capacity:       40000000000,
		BlockTimestamp: 1635320117861,
		Status:         0,
	}
	err = dbDao.CreateIncomeCellInfo(incomeCellInfo)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(incomeCellInfo)
}

func TestInsertListOrUpdate(t *testing.T) {
	dbDao, err := getInit()
	if err != nil {
		t.Fatal(err)
	}
	incomeCellInfos := []TableIncomeCellInfo{
		{
			BlockNumber:    57181892,
			Action:         "create_income",
			Outpoint:       "0xe9116d651c371662b6e29e2102422e23f90656b8619df82c48b782ff4db43a37_2",
			Capacity:       400000000002,
			BlockTimestamp: 16353201178612,
			Status:         0,
		},
		{
			BlockNumber:    5718189,
			Action:         "create_income",
			Outpoint:       "0xe9116d651c371662b6e29e2102422e23f90656b8619df82c48b782ff4db43a37_3",
			Capacity:       40000000000,
			BlockTimestamp: 1635320117861,
			Status:         0,
		},
	}
	err = dbDao.CreateIncomeCellInfoList(incomeCellInfos)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(incomeCellInfos)
}

func TestDelete(t *testing.T) {
	dbDao, err := getInit()
	if err != nil {
		t.Fatal(err)
	}
	err = dbDao.DeleteIncomeCellInfo()
	if err != nil {
		t.Fatal(err)
	}
	// gorm.ErrMissingWhereClause
}

func TestDeleteWhere(t *testing.T) {
	dbDao, err := getInit()
	if err != nil {
		t.Fatal(err)
	}
	err = dbDao.DeleteIncomeCellInfoByOutpoint("0xe9116d651c371662b6e29e2102422e23f90656b8619df82c48b782ff4db43a37_2")
	if err != nil {
		t.Fatal(err)
	}
}

func TestSave(t *testing.T) {
	dbDao, err := getInit()
	if err != nil {
		t.Fatal(err)
	}
	incomeCellInfo := TableIncomeCellInfo{
		BlockNumber:    57181892,
		Action:         "create_income",
		Outpoint:       "0xe9116d651c371662b6e29e2102422e23f90656b8619df82c48b782ff4db43a37_2",
		Capacity:       400000000002,
		BlockTimestamp: 16353201178612,
		Status:         0,
	}
	err = dbDao.SaveIncomeCellInfo(incomeCellInfo)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdateWhere(t *testing.T) {
	dbDao, err := getInit()
	if err != nil {
		t.Fatal(err)
	}
	err = dbDao.UpdateIncomeCellInfoMerged([]string{"0xe9116d651c371662b6e29e2102422e23f90656b8619df82c48b782ff4db43a37_2", "0xe9116d651c371662b6e29e2102422e23f90656b8619df82c48b782ff4db43a37_3"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdatesWhere(t *testing.T) {
	dbDao, err := getInit()
	if err != nil {
		t.Fatal(err)
	}
	incomeCellInfo := TableIncomeCellInfo{
		BlockNumber:    5718189,
		Action:         "create_income",
		Outpoint:       "0xe9116d651c371662b6e29e2102422e23f90656b8619df82c48b782ff4db43a37_2",
		Capacity:       40000000000,
		BlockTimestamp: 1635320117861,
		Status:         0,
	}
	err = dbDao.UpdatesIncomeCellInfo(incomeCellInfo)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFirstWhere(t *testing.T) {
	dbDao, err := getInit()
	if err != nil {
		t.Fatal(err)
	}
	incomeCellInfo, err := dbDao.FirstIncomeCellInfoByOutpoint("0xe9116d651c371662b6e29e2102422e23f90656b8619df82c48b782ff4db43a37_2")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(incomeCellInfo)
}

func TestFindWhere(t *testing.T) {
	dbDao, err := getInit()
	if err != nil {
		t.Fatal(err)
	}
	incomeCellInfo, err := dbDao.FindIncomeCellInfoListByOutpoint("0xe9116d651c371662b6e29e2102422e23f90656b8619df82c48b782ff4db43a37_2")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(incomeCellInfo)
}

func TestTransaction(t *testing.T) {
	dbDao, err := getInit()
	if err != nil {
		t.Fatal(err)
	}
	outpoints := []string{"0x5687b6b801d9214f11d36bfd88edefef7d428941cd48deb3dcc0581307bd7ff5-0", "0xb53374c1f187696dc64fcbe6656b3e5915d7649ddca6c65136334655b57d3375-0"}
	incomeCellInfos := []TableIncomeCellInfo{
		{
			BlockNumber:    57181892,
			Action:         "consolidate_income",
			Outpoint:       "0x5687b6b801d9214f11d36bfd88edefef7d428941cd48deb3dcc0581307bd7ff5-0",
			Capacity:       13375050000,
			BlockTimestamp: 1635320117861,
		},
	}
	transactionInfos := []TableTransactionInfo{
		{
			BlockNumber:    57181892,
			Account:        "linux.bit",
			Action:         "consolidate_income",
			ServiceType:    1,
			ChainType:      0,
			Address:        "0x78100ef7dfa4485dd1118b0d1bca3ef65d742ce0",
			Capacity:       40000000000,
			Outpoint:       "0x5687b6b801d9214f11d36bfd88edefef7d428941cd48deb3dcc0581307bd7ff5-0",
			BlockTimestamp: 1635320117861,
		},
	}
	err = dbDao.ConsolidateIncome(outpoints, incomeCellInfos, transactionInfos)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(incomeCellInfos, "\n")
	t.Log(transactionInfos)
}

func TestUpdateCNYToUSDRate(t *testing.T) {
	dbDao, err := getInit()
	if err != nil {
		t.Fatal(err)
	}
	err = dbDao.UpdateCNYToUSDRate([]string{"ckb_ckb", "eth_eth"}, decimal.NewFromInt(80))
	if err != nil {
		t.Fatal(err)
	}
}

func TestAutoMigrate(t *testing.T) {
	dbDao, err := getInit()
	if err != nil {
		t.Fatal(err)
	}
	err = dbDao.db.AutoMigrate(
		&TableAccountInfo{},
		&TableBlockInfo{},
		&TableIncomeCellInfo{},
		&TableOfferInfo{},
		&TableRebateInfo{},
		&TableRecordsInfo{},
		&TableReverseInfo{},
		&TableSmtInfo{},
		&TableTokenPriceInfo{},
		&TableTradeDealInfo{},
		&TableTradeInfo{},
		&TableTransactionInfo{},
	)
	if err != nil {
		t.Fatal(err)
	}
}
