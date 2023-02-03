package dao

import (
	"gorm.io/gorm/clause"
	"time"
)

type TableSnapshotTxInfo struct {
	Id             uint64    `json:"id" gorm:"column:id; primaryKey; type:bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '';"`
	BlockNumber    uint64    `json:"block_number" gorm:"column:block_number; type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '';"`
	Hash           string    `json:"hash" gorm:"column:hash; uniqueIndex:uk_hash; type:varchar(255) NOT NULL DEFAULT '' COMMENT '';"`
	Action         string    `json:"action" gorm:"column:action; index:k_action; type:varchar(255) NOT NULL DEFAULT '' COMMENT '';"`
	BlockTimestamp uint64    `json:"block_timestamp" gorm:"column:block_timestamp; type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '';"`
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
}

const (
	TableNameSnapshotTxInfo = "t_snapshot_tx_info"
)

func (t *TableSnapshotTxInfo) TableName() string {
	return TableNameSnapshotTxInfo
}

func (d *DbDao) CreateSnapshotTxInfo(info TableSnapshotTxInfo) error {
	if err := d.db.Clauses(clause.Insert{
		Modifier: "IGNORE",
	}).Create(&info).Error; err != nil {
		return err
	}
	return nil
}
