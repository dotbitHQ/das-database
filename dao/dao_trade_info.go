package dao

import (
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type TableTradeInfo struct {
	Id               uint64                `json:"id" gorm:"column:id;primary_key;AUTO_INCREMENT"`
	BlockNumber      uint64                `json:"block_number" gorm:"column:block_number"`
	Outpoint         string                `json:"outpoint" gorm:"column:outpoint"`
	AccountId        string                `json:"account_id" gorm:"account_id"`
	Account          string                `json:"account" gorm:"column:account"`
	OwnerAlgorithmId common.DasAlgorithmId `json:"owner_algorithm_id" gorm:"column:owner_algorithm_id"`
	OwnerChainType   common.ChainType      `json:"owner_chain_type" gorm:"column:owner_chain_type"`
	OwnerAddress     string                `json:"owner_address" gorm:"column:owner_address"`
	Description      string                `json:"description" gorm:"column:description"`
	StartedAt        uint64                `json:"started_at" gorm:"column:started_at"`
	BlockTimestamp   uint64                `json:"block_timestamp" gorm:"column:block_timestamp"`
	PriceCkb         uint64                `json:"price_ckb" gorm:"column:price_ckb"`
	PriceUsd         decimal.Decimal       `json:"price_usd" gorm:"column:price_usd"`
	Status           AccountStatus         `json:"status" gorm:"column:status"` // 0: normal 1: on sale 2: on auction
	CreatedAt        time.Time             `json:"created_at" gorm:"column:created_at"`
	UpdatedAt        time.Time             `json:"updated_at" gorm:"column:updated_at"`
}

const (
	TableNameTradeInfo = "t_trade_info"
)

func (t *TableTradeInfo) TableName() string {
	return TableNameTradeInfo
}

type RecordTotal struct {
	Total int `json:"total" gorm:"column:total"`
}

type SaleAccount struct {
	Account string `json:"account" gorm:"column:account"`
}

func (d *DbDao) StartAccountSale(accountInfo TableAccountInfo, tradeInfo TableTradeInfo, transactionInfo TableTransactionInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Select("block_number", "outpoint", "status").
			Where("account_id = ?", accountInfo.AccountId).
			Updates(accountInfo).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{
				"block_number", "outpoint", "owner_algorithm_id", "owner_chain_type", "owner_address",
				"description", "started_at", "block_timestamp", "price_ckb", "price_usd", "status",
			}),
		}).Create(&tradeInfo).Error; err != nil {
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

func (d *DbDao) EditAccountSale(tradeInfo TableTradeInfo, transactionInfo TableTransactionInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Select("block_number", "outpoint", "description", "block_timestamp", "price_ckb", "price_usd").
			Where("account_id = ?", tradeInfo.AccountId).
			Updates(tradeInfo).Error; err != nil {
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

func (d *DbDao) CancelAccountSale(accountInfo TableAccountInfo, transactionInfo TableTransactionInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Select("block_number", "outpoint", "status").
			Where("account_id = ?", accountInfo.AccountId).
			Updates(accountInfo).Error; err != nil {
			return err
		}

		if err := tx.Where("account_id = ?", accountInfo.AccountId).Delete(&TableTradeInfo{}).Error; err != nil {
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

func (d *DbDao) BuyAccount(incomeCellInfos []TableIncomeCellInfo, accountInfo TableAccountInfo, dealInfo TableTradeDealInfo, transactionInfoBuy, transactionInfoSale TableTransactionInfo, rebateInfos []TableRebateInfo, recordsInfos []TableRecordsInfo) error {
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

		if err := tx.Select("block_number", "outpoint", "owner_chain_type", "owner", "owner_algorithm_id", "manager_chain_type", "manager", "manager_algorithm_id", "status").
			Where("account_id = ?", accountInfo.AccountId).
			Updates(accountInfo).Error; err != nil {
			return err
		}

		if err := tx.Where("account_id = ?", accountInfo.AccountId).Delete(&TableTradeInfo{}).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{
				"account_id", "account", "deal_type", "sell_chain_type", "sell_address",
				"buy_chain_type", "buy_address", "price_ckb", "price_usd",
			}),
		}).Create(&dealInfo).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{
				"account_id", "account", "service_type",
				"chain_type", "address", "capacity", "status",
			}),
		}).Create(&transactionInfoBuy).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{
				"account_id", "account", "service_type",
				"chain_type", "address", "capacity", "status",
			}),
		}).Create(&transactionInfoSale).Error; err != nil {
			return err
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
