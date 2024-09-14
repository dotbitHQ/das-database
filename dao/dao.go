package dao

import (
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/http_api/logger"
	"github.com/shopspring/decimal"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

var log = logger.NewLogger("dao", logger.LevelDebug)

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
		&RuleConfig{},
		&TableSubAccountAutoMintStatement{},
		&TableCidPk{},
		&TableAuthorize{},
		&ApprovalInfo{},
		&TableDidCellInfo{},
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
						"chain_type":   tokenList[i].ChainType,
						"name":         tokenList[i].Name,
						"symbol":       tokenList[i].Symbol,
						"decimals":     tokenList[i].Decimals,
						"logo":         tokenList[i].Logo,
						"coin_type":    tokenList[i].CoinType,
						"display_name": tokenList[i].DisplayName,
						"icon":         tokenList[i].Icon,
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
		CoinType:      common.CoinTypeCKB,
		GeckoId:       "nervos-network",
		Name:          "Nervos Network",
		Symbol:        "CKB",
		Decimals:      8,
		Logo:          "https://app.did.id/images/components/portal-wallet.svg",
		LastUpdatedAt: time.Now().Unix(),
		DisplayName:   ".bit Balance",
		Icon:          "dotbit-balance",
	},
	{
		TokenId:       "eth_eth",
		ChainType:     1,
		CoinType:      common.CoinTypeEth,
		GeckoId:       "ethereum",
		Name:          "Ethereum",
		Symbol:        "ETH",
		Decimals:      18,
		Logo:          "https://app.did.id/images/components/ethereum.svg",
		LastUpdatedAt: time.Now().Unix(),
		DisplayName:   "Ethereum",
		Icon:          "ethereum",
	},
	{
		TokenId:       "btc_btc",
		ChainType:     2,
		CoinType:      "0",
		GeckoId:       "bitcoin",
		Name:          "Bitcoin",
		Symbol:        "BTC",
		Decimals:      8,
		Logo:          "https://app.did.id/images/components/bitcoin.svg",
		LastUpdatedAt: time.Now().Unix(),
		DisplayName:   "Bitcoin",
		Icon:          "",
	},
	{
		TokenId:       "tron_trx",
		ChainType:     3,
		CoinType:      common.CoinTypeTrx,
		GeckoId:       "tron",
		Name:          "TRON",
		Symbol:        "TRX",
		Decimals:      6,
		Logo:          "https://app.did.id/images/components/tron.svg",
		LastUpdatedAt: time.Now().Unix(),
		DisplayName:   "TRON",
		Icon:          "tron",
	},
	{
		TokenId:       "bsc_bnb",
		ChainType:     1,
		CoinType:      common.CoinTypeBSC,
		GeckoId:       "binancecoin",
		Name:          "Binance",
		Symbol:        "BNB",
		Decimals:      18,
		Logo:          "https://app.did.id/images/components/binance-smart-chain.svg",
		LastUpdatedAt: time.Now().Unix(),
		DisplayName:   "Binance",
		Icon:          "binance-smart-chain",
	},
	{
		TokenId:       "polygon_matic",
		ChainType:     1,
		CoinType:      common.CoinTypeMatic,
		GeckoId:       "matic-network",
		Name:          "Polygon",
		Symbol:        "MATIC",
		Decimals:      18,
		Logo:          "https://app.did.id/images/components/polygon.svg",
		LastUpdatedAt: time.Now().Unix(),
		DisplayName:   "Polygon",
		Icon:          "polygon",
	},
	{
		TokenId:       "doge_doge",
		ChainType:     7,
		CoinType:      common.CoinTypeDogeCoin,
		GeckoId:       "doge_doge",
		Name:          "Dogecoin",
		Symbol:        "doge",
		Decimals:      8,
		Logo:          "https://app.did.id/images/components/doge.svg",
		LastUpdatedAt: time.Now().Unix(),
		DisplayName:   "Dogecoin",
		Icon:          "dogecoin",
	},
	{
		TokenId:       "eth_erc20_usdt",
		ChainType:     1,
		CoinType:      common.CoinTypeEth,
		GeckoId:       "eth_erc20_usdt",
		Name:          "ERC20-USDT",
		Symbol:        "ERC20-USDT",
		Decimals:      6,
		Logo:          "",
		Price:         decimal.NewFromInt(1),
		LastUpdatedAt: time.Now().Unix(),
		DisplayName:   "Ethereum",
		Icon:          "ethereum",
	},
	{
		TokenId:       "bsc_bep20_usdt",
		ChainType:     1,
		CoinType:      common.CoinTypeBSC,
		GeckoId:       "bsc_bep20_usdt",
		Name:          "BEP20-USDT",
		Symbol:        "BEP20-USDT",
		Decimals:      6,
		Logo:          "",
		Price:         decimal.NewFromInt(1),
		LastUpdatedAt: time.Now().Unix(),
		DisplayName:   "Binance",
		Icon:          "binance-smart-chain",
	},
	{
		TokenId:       "tron_trc20_usdt",
		ChainType:     3,
		CoinType:      common.CoinTypeTrx,
		GeckoId:       "tron_trc20_usdt",
		Name:          "TRC20-USDT",
		Symbol:        "TRC20-USDT",
		Decimals:      6,
		Logo:          "",
		Price:         decimal.NewFromInt(1),
		LastUpdatedAt: time.Now().Unix(),
		DisplayName:   "TRON",
		Icon:          "tron",
	},
	{
		TokenId:       "stripe_usd",
		ChainType:     99,
		GeckoId:       "stripe_usd",
		Name:          "USD",
		Symbol:        "USD",
		Decimals:      2,
		Logo:          "",
		Price:         decimal.NewFromInt(1),
		LastUpdatedAt: time.Now().Unix(),
		DisplayName:   "by Stripe",
		Icon:          "stripe",
	},
	{
		TokenId:       "did_point",
		ChainType:     98,
		CoinType:      common.CoinTypeCKB,
		GeckoId:       "did_point",
		Name:          "DIDCredits",
		Symbol:        "Credits",
		Decimals:      6,
		Logo:          "",
		Price:         decimal.NewFromInt(1),
		LastUpdatedAt: time.Now().Unix(),
		DisplayName:   "DIDCredits",
		Icon:          "didpoint",
	},
	{
		TokenId:       "ckb_ccc",
		ChainType:     0,
		CoinType:      common.CoinTypeCKB,
		GeckoId:       "ckb_ccc",
		Name:          "Nervos Network",
		Symbol:        "CKB",
		Decimals:      8,
		LastUpdatedAt: time.Now().Unix(),
		DisplayName:   "CKB",
		Icon:          "ckbccc",
	},
}

func (d *DbDao) Transaction(fn func(tx *gorm.DB) error) error {
	return d.db.Transaction(fn)
}
