package dao

import (
	"gorm.io/gorm"
	"time"
)

const (
	TableNameDidCellInfo = "t_did_cell_info"
)

type TableDidCellInfo struct {
	Id           uint64    `json:"id" gorm:"column:id;primaryKey;type:bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT ''"`
	BlockNumber  uint64    `json:"block_number" gorm:"column:block_number;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	Outpoint     string    `json:"outpoint" gorm:"column:outpoint;uniqueIndex:uk_op;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' "`
	AccountId    string    `json:"account_id" gorm:"column:account_id;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT 'hash of account'"`
	Account      string    `json:"account" gorm:"column:account;index:account;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	Args         string    `json:"args" gorm:"column:args;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' "`
	LockCodeHash string    `json:"lock_code_hash" gorm:"column:lock_code_hash;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' "`
	ExpiredAt    uint64    `json:"expired_at" gorm:"column:expired_at;index:k_expired_at;type:bigint(20) unsigned NOT NULL DEFAULT '0' "`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP "`
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP "`
}

func (t *TableDidCellInfo) TableName() string {
	return TableNameDidCellInfo
}

func (d *DbDao) CreateDidCellRecordsInfos(outpoint string, didCellInfo TableDidCellInfo, recordsInfos []TableRecordsInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("account_id = ?", didCellInfo.AccountId).Delete(&TableRecordsInfo{}).Error; err != nil {
			return err
		}

		if len(recordsInfos) > 0 {
			if err := tx.Create(&recordsInfos).Error; err != nil {
				return err
			}
		}

		if err := tx.Select("outpoint", "expired_at", "block_number").
			Where("outpoint = ?", outpoint).
			Updates(didCellInfo).Error; err != nil {
			return err
		}
		return nil
	})
}

func (d *DbDao) EditDidCellOwner(outpoint string, didCellInfo TableDidCellInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Select("outpoint", "block_number", "args", "lock_code_hash").
			Where("outpoint = ?", outpoint).
			Updates(didCellInfo).Error; err != nil {
			return err
		}
		return nil

	})
}

func (d *DbDao) DidCellRecycle(outpoint string) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("outpoint = ? ", outpoint).Delete(&TableRecordsInfo{}).Error; err != nil {
			return err
		}
		return nil
	})
}
