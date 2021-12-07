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
		if err := tx.Select("block_number", "manager", "manager_chain_type", "manager_algorithm_id", "owner", "owner_algorithm_id", "owner_chain_type", "outpoint", "status").Where("account = ?", accountInfo.Account).
			Updates(accountInfo).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{"block_number", "outpoint", "owner_algorithm_id", "owner_chain_type", "owner_address", "description", "started_at", "block_timestamp", "price_ckb", "price_usd", "status"}),
		}).Create(&tradeInfo).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{"block_number", "block_timestamp", "capacity"}),
		}).Create(&transactionInfo).Error; err != nil {
			return err
		}

		return nil
	})
}

func (d *DbDao) EditAccountSale(tradeInfo TableTradeInfo, transactionInfo TableTransactionInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Select("block_number", "outpoint", "description", "block_timestamp", "price_ckb", "price_usd").
			Where("account = ?", tradeInfo.Account).Updates(tradeInfo).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{"block_number", "block_timestamp", "capacity"}),
		}).Create(&transactionInfo).Error; err != nil {
			return err
		}

		return nil
	})
}

func (d *DbDao) CancelAccountSale(accountInfo TableAccountInfo, transactionInfo TableTransactionInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Select("block_number", "outpoint", "status").Where("account = ?", accountInfo.Account).
			Updates(accountInfo).Error; err != nil {
			return err
		}

		if err := tx.Where("account = ?", accountInfo.Account).Delete(&TableTradeInfo{}).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{"block_number", "block_timestamp", "capacity"}),
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
				DoUpdates: clause.AssignmentColumns([]string{"block_number", "capacity", "block_timestamp"}),
			}).Create(&incomeCellInfos).Error; err != nil {
				return err
			}
		}

		if err := tx.Select("block_number", "outpoint", "owner_chain_type", "owner", "owner_algorithm_id", "manager_chain_type", "manager", "manager_algorithm_id", "status").
			Where("account = ?", accountInfo.Account).Updates(accountInfo).Error; err != nil {
			return err
		}

		if err := tx.Where("account = ?", accountInfo.Account).Delete(&TableTradeInfo{}).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{"block_number", "sell_chain_type", "sell_address", "buy_chain_type", "buy_address", "price_ckb", "price_usd"}),
		}).Create(&dealInfo).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{"block_number", "block_timestamp", "capacity"}),
		}).Create(&transactionInfoBuy).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{"block_number", "block_timestamp", "capacity"}),
		}).Create(&transactionInfoSale).Error; err != nil {
			return err
		}

		if len(rebateInfos) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns([]string{"block_number", "invitee_account", "invitee_chain_type", "invitee_address", "reward", "block_timestamp"}),
			}).Create(&rebateInfos).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("account = ?", accountInfo.Account).Delete(&TableRecordsInfo{}).Error; err != nil {
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
