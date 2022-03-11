package dao

import (
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type TableTradeInfo struct {
	Id               uint64                `json:"id" gorm:"column:id;primaryKey;type:bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT ''"`
	BlockNumber      uint64                `json:"block_number" gorm:"column:block_number;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	Outpoint         string                `json:"outpoint" gorm:"column:outpoint;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci  NOT NULL DEFAULT '' COMMENT ''"`
	AccountId        string                `json:"account_id" gorm:"account_id;uniqueIndex:uk_account_id;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci  NOT NULL DEFAULT '' COMMENT 'hash of account'"`
	Account          string                `json:"account" gorm:"column:account;index:k_account;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci  NOT NULL DEFAULT '' COMMENT ''"`
	OwnerAlgorithmId common.DasAlgorithmId `json:"owner_algorithm_id" gorm:"column:owner_algorithm_id;type:smallint(6) NOT NULL DEFAULT '0' COMMENT ''"`
	OwnerChainType   common.ChainType      `json:"owner_chain_type" gorm:"column:owner_chain_type;index:k_oct_oa;type:smallint(6) NOT NULL DEFAULT '0' COMMENT ''"`
	OwnerAddress     string                `json:"owner_address" gorm:"column:owner_address;index:k_oct_oa;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci  NOT NULL DEFAULT '' COMMENT ''"`
	Description      string                `json:"description" gorm:"column:description;type:varchar(2048) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT ''"`
	StartedAt        uint64                `json:"started_at" gorm:"column:started_at;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	BlockTimestamp   uint64                `json:"block_timestamp" gorm:"column:block_timestamp;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	PriceCkb         uint64                `json:"price_ckb" gorm:"column:price_ckb;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	PriceUsd         decimal.Decimal       `json:"price_usd" gorm:"column:price_usd;type:decimal(50, 8) NOT NULL DEFAULT '0.00000000' COMMENT ''"`
	ProfitRate       uint32                `json:"profit_rate" gorm:"column:profit_rate;type:int(11) unsigned NOT NULL DEFAULT '100' COMMENT ''"`
	Status           AccountStatus         `json:"status" gorm:"column:status;type:smallint(6) NOT NULL DEFAULT '0' COMMENT '0: normal 1: on sale 2: on auction'"`
	CreatedAt        time.Time             `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''"`
	UpdatedAt        time.Time             `json:"updated_at" gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
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
		if err := tx.Select("block_number", "manager", "manager_chain_type", "manager_algorithm_id", "owner", "owner_algorithm_id", "owner_chain_type", "outpoint", "status").
			Where("account_id = ?", accountInfo.AccountId).
			Updates(accountInfo).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{
				"block_number", "outpoint", "owner_algorithm_id", "owner_chain_type", "owner_address",
				"description", "started_at", "block_timestamp", "price_ckb", "price_usd", "profit_rate", "status",
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
		if err := tx.Select("block_number", "outpoint", "description", "block_timestamp", "price_ckb", "price_usd", "profit_rate").
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
