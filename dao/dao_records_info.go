package dao

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type TableRecordsInfo struct {
	Id        uint64    `json:"id" gorm:"column:id;primary_key;AUTO_INCREMENT"`
	AccountId string    `json:"account_id" gorm:"account_id"`
	Account   string    `json:"account" gorm:"column:account"`
	Key       string    `json:"key" gorm:"column:key"`
	Type      string    `json:"type" gorm:"column:type"`
	Label     string    `json:"label" gorm:"column:label"`
	Value     string    `json:"value" gorm:"column:value"`
	Ttl       string    `json:"ttl" gorm:"column:ttl"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
}

const (
	TableNameRecordsInfo = "t_records_info"
)

func (t *TableRecordsInfo) TableName() string {
	return TableNameRecordsInfo
}

func (d *DbDao) CreateRecordsInfos(accountInfo TableAccountInfo, recordsInfos []TableRecordsInfo, transactionInfo TableTransactionInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Select("block_number", "outpoint").
			Where("account = ?", accountInfo.Account).Updates(accountInfo).Error; err != nil {
			return err
		}

		if err := tx.Where("account = ?", transactionInfo.Account).Delete(&TableRecordsInfo{}).Error; err != nil {
			return err
		}

		if len(recordsInfos) > 0 {
			if err := tx.Create(&recordsInfos).Error; err != nil {
				return err
			}
		}

		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{"block_number", "block_timestamp", "capacity"}),
		}).Create(&transactionInfo).Error; err != nil {
			return err
		}

		return nil
	})
}

func (d *DbDao) CreateRecordsInfos2(accountInfo TableAccountInfo, recordsInfos []TableRecordsInfo, transactionInfo TableTransactionInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Select("block_number", "outpoint").
			Where("account_id = ?", accountInfo.AccountId).
			Updates(accountInfo).Error; err != nil {
			return err
		}

		if err := tx.Where("account_id = ?", accountInfo.AccountId).Delete(&TableRecordsInfo{}).Error; err != nil {
			return err
		}

		if len(recordsInfos) > 0 {
			if err := tx.Create(&recordsInfos).Error; err != nil {
				return err
			}
		}

		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{
				"account_id", "account", "service_type",
				"chain_type", "address", "capacity", "status",
			}),
		}).Create(&transactionInfo).Error; err != nil {
			return err
		}

		return nil
	})
}
