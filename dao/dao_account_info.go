package dao

import (
	"github.com/DeAccountSystems/das-lib/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type TableAccountInfo struct {
	Id                   uint64                `json:"id" gorm:"column:id;primary_key;AUTO_INCREMENT"`
	BlockNumber          uint64                `json:"block_number" gorm:"column:block_number"`
	Outpoint             string                `json:"outpoint" gorm:"column:outpoint"`
	AccountId            string                `json:"account_id" gorm:"column:account_id"`
	ParentAccountId      string                `json:"parent_account_id" gorm:"column:parent_account_id"`
	Account              string                `json:"account" gorm:"column:account"`
	OwnerChainType       common.ChainType      `json:"owner_chain_type" gorm:"column:owner_chain_type"`
	Owner                string                `json:"owner" gorm:"column:owner"`
	OwnerAlgorithmId     common.DasAlgorithmId `json:"owner_algorithm_id" gorm:"column:owner_algorithm_id"`
	ManagerChainType     common.ChainType      `json:"manager_chain_type" gorm:"column:manager_chain_type"`
	Manager              string                `json:"manager" gorm:"column:manager"`
	ManagerAlgorithmId   common.DasAlgorithmId `json:"manager_algorithm_id" gorm:"column:manager_algorithm_id"`
	Status               AccountStatus         `json:"status" gorm:"column:status"`
	EnableSubAccount     EnableSubAccount      `json:"enable_sub_account" gorm:"column:enable_sub_account"`
	RenewSubAccountPrice uint64                `json:"renew_sub_account_price" gorm:"column:renew_sub_account_price"`
	Nonce                uint64                `json:"nonce" gorm:"column:nonce"`
	RegisteredAt         uint64                `json:"registered_at" gorm:"column:registered_at"`
	ExpiredAt            uint64                `json:"expired_at" gorm:"column:expired_at"`
	ConfirmProposalHash  string                `json:"confirm_proposal_hash" gorm:"column:confirm_proposal_hash"`
	CreatedAt            time.Time             `json:"created_at" gorm:"column:created_at"`
	UpdatedAt            time.Time             `json:"updated_at" gorm:"column:updated_at"`
}

type AccountStatus int
type EnableSubAccount int

const (
	AccountStatusNotOpenRegister AccountStatus = -1
	AccountStatusNormal          AccountStatus = 0
	AccountStatusOnSale          AccountStatus = 1
	AccountStatusOnAuction       AccountStatus = 2
	TableNameAccountInfo                       = "t_account_info"
)
const (
	AccountEnableStatusOff EnableSubAccount = 0
	AccountEnableStatusOn  EnableSubAccount = 1
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

func (d *DbDao) ConfirmProposal(incomeCellInfos []TableIncomeCellInfo, accountInfos []TableAccountInfo, transactionInfos []TableTransactionInfo, rebateInfos []TableRebateInfo) error {
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
