package dao

import (
	"fmt"
	"github.com/scorpiotzh/mylog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
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
		&ReverseSmtInfo{},
		&TableCidPk{},
		&TableAuthorize{},
	); err != nil {
		return nil, err
	}

	if len(tokenList) > 0 {
		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Clauses(clause.Insert{
				Modifier: "IGNORE",
			}).Create(&tokenList).Error; err != nil {
				return err
			}
			for i := range tokenList {
				if err := tx.Model(TableTokenPriceInfo{}).
					Where("token_id=?", tokenList[i].TokenId).
					Updates(map[string]interface{}{
						"chain_type":      tokenList[i].ChainType,
						"name":            tokenList[i].Name,
						"symbol":          tokenList[i].Symbol,
						"decimals":        tokenList[i].Decimals,
						"logo":            tokenList[i].Logo,
						"last_updated_at": tokenList[i].LastUpdatedAt,
					}).Error; err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}

	return &DbDao{db: db}, nil
}

var tokenList = []TableTokenPriceInfo{
	{
		TokenId:       "ckb_ckb",
		ChainType:     0,
		GeckoId:       "nervos-network",
		Name:          "Nervos Network",
		Symbol:        "CKB",
		Decimals:      8,
		Logo:          "https://app.did.id/images/components/portal-wallet.svg",
		LastUpdatedAt: time.Now().Unix(),
	},
	{
		TokenId:       "eth_eth",
		ChainType:     1,
		GeckoId:       "ethereum",
		Name:          "Ethereum",
		Symbol:        "ETH",
		Decimals:      18,
		Logo:          "https://app.did.id/images/components/ethereum.svg",
		LastUpdatedAt: time.Now().Unix(),
	},
	{
		TokenId:       "btc_btc",
		ChainType:     2,
		GeckoId:       "bitcoin",
		Name:          "Bitcoin",
		Symbol:        "BTC",
		Decimals:      8,
		Logo:          "https://app.did.id/images/components/bitcoin.svg",
		LastUpdatedAt: time.Now().Unix(),
	},
	{
		TokenId:       "tron_trx",
		ChainType:     3,
		GeckoId:       "tron",
		Name:          "TRON",
		Symbol:        "TRX",
		Decimals:      6,
		Logo:          "https://app.did.id/images/components/tron.svg",
		LastUpdatedAt: time.Now().Unix(),
	},
	{
		TokenId:       "bsc_bnb",
		ChainType:     1,
		GeckoId:       "binancecoin",
		Name:          "Binance",
		Symbol:        "BNB",
		Decimals:      18,
		Logo:          "https://app.did.id/images/components/binance-smart-chain.svg",
		LastUpdatedAt: time.Now().Unix(),
	},
	{
		TokenId:       "polygon_matic",
		ChainType:     1,
		GeckoId:       "matic-network",
		Name:          "Polygon",
		Symbol:        "MATIC",
		Decimals:      18,
		Logo:          "https://app.did.id/images/components/polygon.svg",
		LastUpdatedAt: time.Now().Unix(),
	},
	{
		TokenId:       "doge_doge",
		ChainType:     7,
		GeckoId:       "doge_doge",
		Name:          "Dogecoin",
		Symbol:        "doge",
		Decimals:      8,
		Logo:          "https://app.did.id/images/components/doge.svg",
		LastUpdatedAt: time.Now().Unix(),
	},
}

func (d *DbDao) Transaction(fn func(tx *gorm.DB) error) error {
	return d.db.Transaction(fn)
}
