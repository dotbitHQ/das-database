package dao

import (
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/shopspring/decimal"
	"time"
)

type TableTradeDealInfo struct {
	Id             uint64           `json:"id" gorm:"column:id;primaryKey;type:bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT ''"`
	BlockNumber    uint64           `json:"block_number" gorm:"column:block_number;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	Outpoint       string           `json:"outpoint" gorm:"column:outpoint;uniqueIndex:uk_outpoint;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	AccountId      string           `json:"account_id" gorm:"account_id;index:k_account_id;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT 'hash of account'"`
	Account        string           `json:"account" gorm:"column:account;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	DealType       int              `json:"deal_type" gorm:"column:deal_type;type:smallint(6) NOT NULL DEFAULT '0' COMMENT '0: sale 1: auction'"`
	SellChainType  common.ChainType `json:"sell_chain_type" gorm:"column:sell_chain_type;index:k_sct_sa;type:int(11) NOT NULL DEFAULT '0' COMMENT ''"`
	SellAddress    string           `json:"sell_address" gorm:"column:sell_address;index:k_sct_sa;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	BuyChainType   common.ChainType `json:"buy_chain_type" gorm:"column:buy_chain_type;index:k_bct_ba;type:int(11) NOT NULL DEFAULT '0' COMMENT ''"`
	BuyAddress     string           `json:"buy_address" gorm:"column:buy_address;index:k_bct_ba;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	PriceCkb       uint64           `json:"price_ckb" gorm:"column:price_ckb;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'price in CKB'"`
	PriceUsd       decimal.Decimal  `json:"price_usd" gorm:"column:price_usd;type:decimal(50, 8) NOT NULL DEFAULT '0.00000000' COMMENT 'price in dollar'"`
	BlockTimestamp uint64           `json:"block_timestamp" gorm:"column:block_timestamp;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	CreatedAt      time.Time        `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''"`
	UpdatedAt      time.Time        `json:"updated_at" gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
}

const (
	TableNameTradeDealInfo = "t_trade_deal_info"
	DealTypeSale           = 0
	DealTypeAuction        = 1
	DealTypeOffer          = 2
)

func (t *TableTradeDealInfo) TableName() string {
	return TableNameTradeDealInfo
}
