package dao

import (
	"github.com/dotbitHQ/das-lib/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type TableAccountInfo struct {
	Id                   uint64                `json:"id" gorm:"column:id;primaryKey;type:bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT ''"`
	BlockNumber          uint64                `json:"block_number" gorm:"column:block_number;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	Outpoint             string                `json:"outpoint" gorm:"column:outpoint;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT 'Hash-Index'"`
	AccountId            string                `json:"account_id" gorm:"column:account_id;uniqueIndex:uk_account_id;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT 'hash of account'"`
	ParentAccountId      string                `json:"parent_account_id" gorm:"column:parent_account_id;index:k_parent_account_id;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	Account              string                `json:"account" gorm:"column:account;index:account;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	OwnerChainType       common.ChainType      `json:"owner_chain_type" gorm:"column:owner_chain_type;index:k_oct_o;type:smallint(6) NOT NULL DEFAULT '0' COMMENT ''"`
	Owner                string                `json:"owner" gorm:"column:owner;index:k_oct_o;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT 'owner address'"`
	OwnerAlgorithmId     common.DasAlgorithmId `json:"owner_algorithm_id" gorm:"column:owner_algorithm_id;type:smallint(6) NOT NULL DEFAULT '0' COMMENT ''"`
	ManagerChainType     common.ChainType      `json:"manager_chain_type" gorm:"column:manager_chain_type;index:k_mct_m;type:smallint(6) NOT NULL DEFAULT '0' COMMENT ''"`
	Manager              string                `json:"manager" gorm:"column:manager;index:k_mct_m;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT 'manager address'"`
	ManagerAlgorithmId   common.DasAlgorithmId `json:"manager_algorithm_id" gorm:"column:manager_algorithm_id;type:smallint(6) NOT NULL DEFAULT '0' COMMENT ''"`
	Status               uint8                 `json:"status" gorm:"column:status;type:smallint(6) NOT NULL DEFAULT '0' COMMENT ''"`
	EnableSubAccount     uint8                 `json:"enable_sub_account" gorm:"column:enable_sub_account;type:smallint(6) NOT NULL DEFAULT '0' COMMENT ''"`
	RenewSubAccountPrice uint64                `json:"renew_sub_account_price" gorm:"column:renew_sub_account_price;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	Nonce                uint64                `json:"nonce" gorm:"column:nonce;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	RegisteredAt         uint64                `json:"registered_at" gorm:"column:registered_at; index:k_registered_at; type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	ExpiredAt            uint64                `json:"expired_at" gorm:"column:expired_at;index:k_expired_at;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	ConfirmProposalHash  string                `json:"confirm_proposal_hash" gorm:"column:confirm_proposal_hash;index:k_confirm_proposal_hash;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	CharsetNum           uint64                `json:"charset_num" gorm:"column:charset_num; index:k_charset_num; type: bigint(20) unsigned NOT NULL DEFAULT '0'; "`
	CreatedAt            time.Time             `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''"`
	UpdatedAt            time.Time             `json:"updated_at" gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
}

type AccountStatus uint8
type EnableSubAccount uint8

const (
	AccountStatusNormal    AccountStatus = 0
	AccountStatusOnSale    AccountStatus = 1
	AccountStatusOnAuction AccountStatus = 2
	AccountStatusOnLock    AccountStatus = 3

	AccountEnableStatusOff EnableSubAccount = 0
	AccountEnableStatusOn  EnableSubAccount = 1

	TableNameAccountInfo = "t_account_info"
)

func (t *TableAccountInfo) TableName() string {
	return TableNameAccountInfo
}

func (d *DbDao) EditManager(accountInfo TableAccountInfo, transactionInfo TableTransactionInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Select("block_number", "outpoint", "manager_chain_type", "manager", "manager_algorithm_id").
			Where("account_id = ?", accountInfo.AccountId).Updates(accountInfo).Error; err != nil {
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

func (d *DbDao) TransferAccount(accountInfo TableAccountInfo, transactionInfo TableTransactionInfo, recordsInfos []TableRecordsInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Select("block_number", "outpoint", "owner_chain_type", "owner", "owner_algorithm_id", "manager_chain_type", "manager", "manager_algorithm_id").
			Where("account_id = ?", accountInfo.AccountId).
			Updates(accountInfo).Error; err != nil {
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

func (d *DbDao) ConfirmProposal(incomeCellInfos []TableIncomeCellInfo, accountInfos []TableAccountInfo, transactionInfos []TableTransactionInfo, rebateInfos []TableRebateInfo, records []TableRecordsInfo, recordAccountIds []string) error {
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
				}),
			}).Create(&accountInfos).Error; err != nil {
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

		if len(rebateInfos) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns([]string{
					"invitee_id", "invitee_account", "invitee_chain_type", "invitee_address",
					"reward", "action", "service_type", "inviter_args",
					"inviter_id", "inviter_account", "inviter_chain_type", "inviter_address",
				}),
			}).Create(&rebateInfos).Error; err != nil {
				return err
			}
		}

		if len(recordAccountIds) > 0 {
			if err := tx.Where("account_id IN(?)", recordAccountIds).Delete(&TableRecordsInfo{}).Error; err != nil {
				return err
			}
		}
		if len(records) > 0 {
			if err := tx.Create(&records).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (d *DbDao) EnableSubAccount(accountInfo TableAccountInfo, transactionInfo TableTransactionInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Select("block_number", "outpoint", "enable_sub_account", "renew_sub_account_price").
			Where("account_id = ?", accountInfo.AccountId).
			Updates(accountInfo).Error; err != nil {
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

func (d *DbDao) ForceRecoverAccountStatus(oldStatus uint8, accountInfo TableAccountInfo, transactionInfo TableTransactionInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if oldStatus == 1 {
			if err := tx.Where("account_id=?", transactionInfo.AccountId).Delete(&TableTradeInfo{}).Error; err != nil {
				return err
			}
		}

		if err := tx.Select("block_number", "outpoint", "status").
			Where("account_id = ?", accountInfo.AccountId).
			Updates(accountInfo).Error; err != nil {
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

func (d *DbDao) GetAccountInfoByParentAccountId(parentAccountId string) (accountInfos []TableAccountInfo, err error) {
	err = d.db.Where("parent_account_id=?", parentAccountId).Find(&accountInfos).Error
	return
}

func (d *DbDao) RecycleExpiredAccount(accountInfo TableAccountInfo, transactionInfo TableTransactionInfo, accountId string, enableSubAccount uint8) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Select("block_number", "outpoint").
			Where("account_id=?", accountInfo.AccountId).
			Updates(accountInfo).Error; err != nil {
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

		if err := tx.Where("account_id=?", accountId).Delete(&TableAccountInfo{}).Error; err != nil {
			return err
		}

		if err := tx.Where("account_id=?", accountId).Delete(&TableRecordsInfo{}).Error; err != nil {
			return err
		}

		if enableSubAccount == 1 {
			if err := tx.Where("parent_account_id=?", accountId).Delete(&TableAccountInfo{}).Error; err != nil {
				return err
			}

			if err := tx.Where("parent_account_id=?", accountId).Delete(&TableRecordsInfo{}).Error; err != nil {
				return err
			}

			if err := tx.Where("parent_account_id=?", accountId).Delete(&TableSmtInfo{}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (d *DbDao) AccountCrossChain(accountInfo TableAccountInfo, transactionInfo TableTransactionInfo, isTrans bool) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Select("block_number", "outpoint",
			"owner_chain_type", "owner", "owner_algorithm_id", "manager_chain_type", "manager", "manager_algorithm_id", "status").
			Where("account_id = ?", accountInfo.AccountId).
			Updates(accountInfo).Error; err != nil {
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

		if isTrans {
			if err := tx.Where("account_id = ?", accountInfo.AccountId).Delete(&TableRecordsInfo{}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (d *DbDao) GetNeedFixCharsetAccountList() (list []TableAccountInfo, err error) {
	err = d.db.Where("parent_account_id='' AND charset_num=0 AND account_id!='0x0000000000000000000000000000000000000000' ").Limit(200).Find(&list).Error
	return
}

func (d *DbDao) UpdateAccountCharsetNum(accCharset map[string]uint64) error {
	if len(accCharset) == 0 {
		return nil
	}

	return d.db.Transaction(func(tx *gorm.DB) error {
		for k, v := range accCharset {
			if err := tx.Model(TableAccountInfo{}).
				Where("account_id=? AND charset_num=0", k).
				Updates(map[string]interface{}{
					"charset_num": v,
				}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
