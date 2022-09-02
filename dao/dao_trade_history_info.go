package dao

import (
	"github.com/dotbitHQ/das-lib/common"
	"github.com/shopspring/decimal"
	"time"
)

type TableTradeHistoryInfo struct {
	Id               uint64                `json:"id" gorm:"column:id;primaryKey;type:bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT ''"`
	BlockNumber      uint64                `json:"block_number" gorm:"column:block_number;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	Outpoint         string                `json:"outpoint" gorm:"column:outpoint; uniqueIndex:uk_outpoint; type:varchar(255) NOT NULL DEFAULT '' COMMENT ''"`
	AccountId        string                `json:"account_id" gorm:"account_id;index:k_account_id;type:varchar(255) NOT NULL DEFAULT '' COMMENT 'hash of account'"`
	Account          string                `json:"account" gorm:"column:account;index:k_account;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci  NOT NULL DEFAULT '' COMMENT ''"`
	OwnerAlgorithmId common.DasAlgorithmId `json:"owner_algorithm_id" gorm:"column:owner_algorithm_id;type:smallint(6) NOT NULL DEFAULT '0' COMMENT ''"`
	OwnerChainType   common.ChainType      `json:"owner_chain_type" gorm:"column:owner_chain_type;index:k_oct_oa;type:smallint(6) NOT NULL DEFAULT '0' COMMENT ''"`
	OwnerAddress     string                `json:"owner_address" gorm:"column:owner_address;index:k_oct_oa;type:varchar(255) NOT NULL DEFAULT '' COMMENT ''"`
	Description      string                `json:"description" gorm:"column:description;type:varchar(2048) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT ''"`
	StartedAt        uint64                `json:"started_at" gorm:"column:started_at;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	BlockTimestamp   uint64                `json:"block_timestamp" gorm:"column:block_timestamp;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	PriceCkb         uint64                `json:"price_ckb" gorm:"column:price_ckb;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	PriceUsd         decimal.Decimal       `json:"price_usd" gorm:"column:price_usd;type:decimal(50, 8) NOT NULL DEFAULT '0' COMMENT ''"`
	ProfitRate       uint32                `json:"profit_rate" gorm:"column:profit_rate;type:int(11) unsigned NOT NULL DEFAULT '100' COMMENT ''"`
	Status           uint8                 `json:"status" gorm:"column:status;type:smallint(6) NOT NULL DEFAULT '0' COMMENT '0: normal 1: on sale 2: on auction'"`
	CreatedAt        time.Time             `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''"`
	UpdatedAt        time.Time             `json:"updated_at" gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
}

const (
	TableNameTradeHistoryInfo = "t_trade_history_info"
)

func (t *TableTradeHistoryInfo) TableName() string {
	return TableNameTradeHistoryInfo
}
