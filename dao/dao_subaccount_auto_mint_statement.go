package dao

import (
	"github.com/shopspring/decimal"
	"time"
)

type SubAccountAutoMintTxType int

const (
	SubAccountAutoMintTxTypeIncome      SubAccountAutoMintTxType = 1
	SubAccountAutoMintTxTypeExpenditure SubAccountAutoMintTxType = 2
)

type TableSubAccountAutoMintStatement struct {
	Id                uint64                   `json:"id" gorm:"column:id; primaryKey; type:bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '';"`
	BlockNumber       uint64                   `json:"block_number" gorm:"column:block_number; index:k_block_number; type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '';"`
	TxHash            string                   `json:"tx_hash" gorm:"column:tx_hash; uniqueIndex:uk_tx_wi; index:idx_hash; type:varchar(255) NOT NULL DEFAULT '' COMMENT '';"`
	WitnessIndex      int                      `json:"witness_index" gorm:"column:witness_index; uniqueIndex:uk_tx_wi; type:int(11) NOT NULL DEFAULT '0' COMMENT '';"`
	ParentAccountId   string                   `json:"parent_account_id" gorm:"column:parent_account_id; type:varchar(255) NOT NULL DEFAULT '' COMMENT '';"`
	ServiceProviderId string                   `json:"service_provider_id" gorm:"column:service_provider_id; index:idx_service_provider_id; type:varchar(255) NOT NULL DEFAULT '' COMMENT '';"`
	Price             decimal.Decimal          `json:"price" gorm:"column:price; type:decimal(60,2) NOT NULL DEFAULT '0' COMMENT '';"`
	BlockTimestamp    uint64                   `json:"block_timestamp" gorm:"column:block_timestamp; type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '';"`
	TxType            SubAccountAutoMintTxType `json:"tx_type" gorm:"column:tx_type; type:int(11) NOT NULL DEFAULT '0' COMMENT '1: income, 2: expenditure';"`
	CreatedAt         time.Time                `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''"`
	UpdatedAt         time.Time                `json:"updated_at" gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
}

func (t *TableSubAccountAutoMintStatement) TableName() string {
	return "t_sub_account_auto_mint_statement"
}
