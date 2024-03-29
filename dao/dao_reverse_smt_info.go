package dao

import (
	"time"
)

// ReverseSmtInfo current reverse info
type ReverseSmtInfo struct {
	ID           uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	RootHash     string    `gorm:"column:root_hash;NOT NULL"` // SMT根节点hash
	BlockNumber  uint64    `gorm:"column:block_number;default:0;NOT NULL;index:idx_address_blk_outpoint,priority:2"`
	AlgorithmID  uint8     `gorm:"column:algorithm_id;type:smallint(6) NOT NULL DEFAULT '0'"`
	Outpoint     string    `gorm:"column:outpoint;NOT NULL;index:idx_address_blk_outpoint,priority:3"` // 设置反解的地址
	Address      string    `gorm:"column:address;NOT NULL;index:idx_address_blk_outpoint,priority:1"`
	LeafDataHash string    `gorm:"column:leaf_data_hash;NOT NULL"` // SMT叶子节点hash
	CreatedAt    time.Time `gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''"`
	UpdatedAt    time.Time `gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
}

const TableNameReverseSmtInfo = "t_reverse_smt_info"

func (m *ReverseSmtInfo) TableName() string {
	return TableNameReverseSmtInfo
}
