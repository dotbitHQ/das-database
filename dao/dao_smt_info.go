package dao

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type TableSmtInfo struct {
	Id              uint64    `json:"id" gorm:"column:id;primary_key;AUTO_INCREMENT"`
	BlockNumber     uint64    `json:"block_number" gorm:"column:block_number"`
	Outpoint        string    `json:"outpoint" gorm:"column:outpoint"`
	AccountId       string    `json:"account_id" gorm:"column:account_id"`
	ParentAccountId string    `json:"parent_account_id" gorm:"column:parent_account_id"`
	Account         string    `json:"account" gorm:"column:account"`
	LeafDataHash    string    `json:"leaf_data_hash" gorm:"column:leaf_data_hash"`
	CreatedAt       time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"column:updated_at"`
}

const (
	TableNameSmtInfo = "t_smt_info"
)

func (t *TableSmtInfo) TableName() string {
	return TableNameSmtInfo
}

func (d *DbDao) CreateSubAccount(incomeCellInfos []TableIncomeCellInfo, accountInfos []TableAccountInfo,
	smtInfos []TableSmtInfo, transactionInfos []TableTransactionInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if len(incomeCellInfos) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns([]string{
					"action", "capacity", "status",
				}),
			}).Create(&incomeCellInfos).Error; err != nil {
				return err
			}
		}

		if len(accountInfos) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns([]string{
					"block_number", "outpoint",
					"owner_chain_type", "owner", "owner_algorithm_id",
					"manager_chain_type", "manager", "manager_algorithm_id",
					"registered_at", "expired_at", "status",
					"enable_sub_account", "renew_sub_account_price", "nonce",
				}),
			}).Create(&accountInfos).Error; err != nil {
				return err
			}
		}

		if len(smtInfos) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns([]string{
					"block_number", "outpoint", "leaf_data_hash",
				}),
			}).Create(&smtInfos).Error; err != nil {
				return err
			}
		}

		if len(transactionInfos) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns([]string{
					"account_id", "account", "service_type",
					"chain_type", "address", "capacity", "status",
				}),
			}).Create(&transactionInfos).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
