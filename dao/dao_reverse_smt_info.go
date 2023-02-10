package dao

import (
	"time"
)

// ReverseSmtInfo current reverse info
type ReverseSmtInfo struct {
	ID           uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	RootHash     string    `gorm:"column:root_hash;NOT NULL"` // SMT根节点hash
	BlockNumber  uint64    `gorm:"column:block_number;default:0;NOT NULL"`
	Outpoint     string    `gorm:"column:outpoint;NOT NULL"` // 设置反解的地址
	Address      string    `gorm:"column:address;NOT NULL"`
	LeafDataHash string    `gorm:"column:leaf_data_hash;NOT NULL"` // SMT叶子节点hash
	CreatedAt    time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP;NOT NULL"`
	UpdatedAt    time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP;NOT NULL"`
}

const TableNameReverseSmtInfo = "t_reverse_smt_info"

func (m *ReverseSmtInfo) TableName() string {
	return TableNameReverseSmtInfo
}
