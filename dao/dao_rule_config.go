package dao

import (
	"time"
)

type RuleConfig struct {
	Id             int64     `gorm:"column:id;AUTO_INCREMENT" json:"id"`
	Account        string    `gorm:"column:account;type:varchar(255);comment:账号;NOT NULL" json:"account"`
	AccountId      string    `gorm:"uniqueIndex:idx_acc_id,column:account_id;type:varchar(255);comment:账号id;NOT NULL" json:"account_id"`
	TxHash         string    `gorm:"index:idx_tx_hash,column:tx_hash;type:varchar(255);comment:交易hash;NOT NULL" json:"tx_hash"`
	BlockNumber    uint64    `json:"block_number" gorm:"column:block_number;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	BlockTimestamp uint64    `json:"block_timestamp" gorm:"column:block_timestamp;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	CreatedAt      time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL" json:"updated_at"`
}

func (m *RuleConfig) TableName() string {
	return "t_rule_config"
}
