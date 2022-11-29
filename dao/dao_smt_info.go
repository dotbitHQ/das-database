package dao

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type TableSmtInfo struct {
	Id              uint64    `json:"id" gorm:"column:id;primaryKey;type:bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT ''"`
	BlockNumber     uint64    `json:"block_number" gorm:"column:block_number;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	Outpoint        string    `json:"outpoint" gorm:"column:outpoint;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	AccountId       string    `json:"account_id" gorm:"column:account_id;uniqueIndex:uk_account_id;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	ParentAccountId string    `json:"parent_account_id" gorm:"column:parent_account_id;index:k_parent_account_id;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	LeafDataHash    string    `json:"leaf_data_hash" gorm:"column:leaf_data_hash;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	CreatedAt       time.Time `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
}

const (
	TableNameSmtInfo = "t_smt_info"
)

func (t *TableSmtInfo) TableName() string {
	return TableNameSmtInfo
}

func (d *DbDao) CreateSubAccount(subAccountIds []string, accountInfos []TableAccountInfo, smtInfos []TableSmtInfo, transactionInfo TableTransactionInfo, parentAccountInfo TableAccountInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if len(subAccountIds) > 0 {
			if err := tx.Where(" account_id IN(?) ", subAccountIds).
				Delete(&TableRecordsInfo{}).Error; err != nil {
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

		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{
				"account_id", "account", "service_type",
				"chain_type", "address", "capacity", "status",
			}),
		}).Create(&transactionInfo).Error; err != nil {
			return err
		}

		if parentAccountInfo.AccountId != "" {
			if err := tx.Select("block_number", "outpoint").
				Where("account_id = ?", parentAccountInfo.AccountId).
				Updates(parentAccountInfo).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (d *DbDao) UpdateSubAccountForCreate(subAccountIds []string, accountInfos []TableAccountInfo, smtInfos []TableSmtInfo, transactionInfo TableTransactionInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if len(subAccountIds) > 0 {
			if err := tx.Where("account_id IN(?)", subAccountIds).
				Delete(&TableRecordsInfo{}).Error; err != nil {
				return err
			}
		}
		if len(accountInfos) > 0 {
			if err := tx.Clauses(clause.Insert{
				Modifier: "IGNORE",
			}).Create(&accountInfos).Error; err != nil {
				return err
			}
		}

		if len(smtInfos) > 0 {
			if err := tx.Clauses(clause.Insert{
				Modifier: "IGNORE",
			}).Create(&smtInfos).Error; err != nil {
				return err
			}
		}

		if err := tx.Clauses(clause.Insert{
			Modifier: "IGNORE",
		}).Create(&transactionInfo).Error; err != nil {
			return err
		}

		return nil
	})
}

func (d *DbDao) EditOwnerSubAccount(accountInfo TableAccountInfo, smtInfo TableSmtInfo, transactionInfo TableTransactionInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Select("block_number", "outpoint",
			"owner_chain_type", "owner", "owner_algorithm_id",
			"manager_chain_type", "manager", "manager_algorithm_id", "nonce").
			Where("account_id = ?", accountInfo.AccountId).
			Updates(accountInfo).Error; err != nil {
			return err
		}

		if err := tx.Select("block_number", "outpoint", "leaf_data_hash").
			Where("account_id = ?", accountInfo.AccountId).
			Updates(&smtInfo).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{
				"account_id", "account", "service_type",
				"chain_type", "address", "capacity", "status",
			}),
		}).Create(&transactionInfo).Error; err != nil {
			return err
		}

		if err := tx.Where("account_id = ?", accountInfo.AccountId).Delete(&TableRecordsInfo{}).Error; err != nil {
			return err
		}

		return nil
	})
}
func (d *DbDao) EditManagerSubAccount(accountInfo TableAccountInfo, smtInfo TableSmtInfo, transactionInfo TableTransactionInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Select("block_number", "outpoint",
			"manager_chain_type", "manager", "manager_algorithm_id", "nonce").
			Where("account_id = ?", accountInfo.AccountId).
			Updates(accountInfo).Error; err != nil {
			return err
		}

		if err := tx.Select("block_number", "outpoint", "leaf_data_hash").
			Where("account_id = ?", accountInfo.AccountId).
			Updates(&smtInfo).Error; err != nil {
			return err
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
func (d *DbDao) EditRecordsSubAccount(accountInfo TableAccountInfo, smtInfo TableSmtInfo, transactionInfo TableTransactionInfo, recordsInfos []TableRecordsInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Select("block_number", "outpoint", "nonce").
			Where("account_id = ?", accountInfo.AccountId).
			Updates(accountInfo).Error; err != nil {
			return err
		}

		if err := tx.Select("block_number", "outpoint", "leaf_data_hash").
			Where("account_id = ?", accountInfo.AccountId).
			Updates(&smtInfo).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{
				"account_id", "account", "service_type",
				"chain_type", "address", "capacity", "status",
			}),
		}).Create(&transactionInfo).Error; err != nil {
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

		return nil
	})
}

func (d *DbDao) RenewSubAccount(accountInfos []TableAccountInfo, smtInfos []TableSmtInfo, transactionInfos []TableTransactionInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if len(accountInfos) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns([]string{
					"block_number", "outpoint", "expired_at", "nonce",
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

func (d *DbDao) RecycleSubAccount(accountIds []string, transactionInfos []TableTransactionInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("account_id IN(?)", accountIds).Delete(&TableAccountInfo{}).Error; err != nil {
			return err
		}

		if err := tx.Where("account_id IN(?)", accountIds).Delete(&TableSmtInfo{}).Error; err != nil {
			return err
		}

		if err := tx.Where("account_id IN(?)", accountIds).Delete(&TableRecordsInfo{}).Error; err != nil {
			return err
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
