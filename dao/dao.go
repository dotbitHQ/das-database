package dao

import (
	"das_database/config"
	"fmt"
	"github.com/scorpiotzh/mylog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var log = mylog.NewLogger("dao", mylog.LevelDebug)

type DbDao struct {
	db *gorm.DB
}

func NewDbDao(db *gorm.DB) *DbDao {
	return &DbDao{db: db}
}

func NewGormDataBase(addr, user, password, dbName string, maxOpenConn, maxIdleConn int) (*gorm.DB, error) {
	conn := "%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local"
	dsn := fmt.Sprintf(conn, user, password, addr, dbName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("gorm open :%v", err)
	}
	db = db.Debug()
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("gorm db :%v", err)
	}

	sqlDB.SetMaxOpenConns(maxOpenConn)
	sqlDB.SetMaxIdleConns(maxIdleConn)

	return db, nil
}

func Initialize(db *gorm.DB) (*DbDao, error) {
	if db.Migrator().HasIndex(&TableBlockInfo{}, "uk_block_number") {
		log.Info("HasIndex: uk_block_number")
		if err := db.Migrator().DropIndex(&TableBlockInfo{}, "uk_block_number"); err != nil {
			return nil, fmt.Errorf("DropIndex err: %s", err.Error())
		}
	}
	// AutoMigrate will create tables, missing foreign keys, constraints, columns and indexes.
	// It will change existing column’s type if its size, precision, nullable changed.
	// It WON’T delete unused columns to protect your data.
	if err := db.AutoMigrate(
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
		&TableCustomScriptInfo{},
		&TableTradeHistoryInfo{},
		&TableSnapshotTxInfo{},
		&TableSnapshotPermissionsInfo{},
		&TableSnapshotRegisterInfo{},
	); err != nil {
		return nil, err
	}

	var tokenList []TableTokenPriceInfo
	for _, v := range config.Cfg.GeckoIds {
		if tokenInfo, ok := geckoIds[v]; ok {
			tokenList = append(tokenList, tokenInfo)
		}
	}
	if len(tokenList) > 0 {
		if err := db.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{"chain_type", "name", "symbol", "decimals", "logo"}),
		}).Create(&tokenList).Error; err != nil {
			return nil, err
		}
	}

	return &DbDao{db: db}, nil
}

var geckoIds = map[string]TableTokenPriceInfo{
	"nervos-network": {
		TokenId:   "ckb_ckb",
		GeckoId:   "nervos-network",
		ChainType: 0,
		Name:      "Nervos Network",
		Symbol:    "CKB",
		Decimals:  8,
		Logo:      "https://app.did.id/images/components/portal-wallet.svg",
	},
	"ethereum": {
		TokenId:   "eth_eth",
		GeckoId:   "ethereum",
		ChainType: 1,
		Name:      "Ethereum",
		Symbol:    "ETH",
		Decimals:  18,
		Logo:      "https://app.did.id/images/components/ethereum.svg",
	},
	"bitcoin": {
		TokenId:   "btc_btc",
		GeckoId:   "bitcoin",
		ChainType: 2,
		Name:      "Bitcoin",
		Symbol:    "BTC",
		Decimals:  8,
		Logo:      "https://app.did.id/images/components/bitcoin.svg",
	},
	"tron": {
		TokenId:   "tron_trx",
		GeckoId:   "tron",
		ChainType: 3,
		Name:      "TRON",
		Symbol:    "TRX",
		Decimals:  6,
		Logo:      "https://app.did.id/images/components/tron.svg",
	},
	"_wx_cny_": {
		TokenId:   "wx_cny",
		GeckoId:   "_wx_cny_",
		ChainType: 4,
		Name:      "WeChat Pay",
		Symbol:    "¥",
		Decimals:  2,
		Logo:      "https://app.did.id/images/components/wechat_pay.png",
	},
	"binancecoin": {
		TokenId:   "bsc_bnb",
		GeckoId:   "binancecoin",
		ChainType: 5,
		Name:      "Binance",
		Symbol:    "BNB",
		Decimals:  18,
		Logo:      "https://app.did.id/images/components/binance-smart-chain.svg",
	},
	"matic-network": {
		TokenId:   "polygon_matic",
		GeckoId:   "matic-network",
		ChainType: 1,
		Name:      "Polygon",
		Symbol:    "MATIC",
		Decimals:  18,
		Logo:      "https://app.did.id/images/components/polygon.svg",
	},
}
