package dao

import (
	"github.com/DeAccountSystems/das-lib/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type TableOfferInfo struct {
	Id             uint64                `json:"id" gorm:"column:id;primary_key;AUTO_INCREMENT"`
	BlockNumber    uint64                `json:"block_number" gorm:"column:block_number"`
	Outpoint       string                `json:"outpoint" gorm:"column:outpoint"`
	AccountId      string                `json:"account_id" gorm:"account_id"`
	Account        string                `json:"account" gorm:"column:account"`
	AlgorithmId    common.DasAlgorithmId `json:"algorithm_id" gorm:"column:algorithm_id"`
	ChainType      common.ChainType      `json:"chain_type" gorm:"column:chain_type"`
	Address        string                `json:"address" gorm:"column:address"`
	BlockTimestamp uint64                `json:"block_timestamp" gorm:"column:block_timestamp"`
	Price          uint64                `json:"price" gorm:"column:price"`
	Message        string                `json:"message" gorm:"column:message"`
	InviterArgs    string                `json:"inviter_args" gorm:"column:inviter_args"`
	ChannelArgs    string                `json:"channel_args" gorm:"column:channel_args"`
	CreatedAt      time.Time             `json:"created_at" gorm:"column:created_at"`
	UpdatedAt      time.Time             `json:"updated_at" gorm:"column:updated_at"`
}

const (
	TableNameOfferInfo = "t_offer_info"
)

func (t *TableOfferInfo) TableName() string {
	return TableNameOfferInfo
}

func (d *DbDao) MakeOffer(offerInfo TableOfferInfo, transactionInfo TableTransactionInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{
				"account_id", "account", "algorithm_id", "chain_type", "address",
				"price", "message", "inviter_args", "channel_args",
			}),
		}).Create(&offerInfo).Error; err != nil {
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

func (d *DbDao) EditOffer(oldOutpoint string, offerInfo TableOfferInfo, transactionInfo TableTransactionInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Select(
			"block_number", "outpoint", "account_id", "account",
			"block_timestamp", "price", "message").
			Where("outpoint = ?", oldOutpoint).Updates(offerInfo).Error; err != nil {
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

func (d *DbDao) CancelOffer(oldOutpoints []string, transactionInfo TableTransactionInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("outpoint IN ?", oldOutpoints).Delete(&TableOfferInfo{}).Error; err != nil {
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

func (d *DbDao) AcceptOffer(incomeCellInfos []TableIncomeCellInfo, accountInfo TableAccountInfo, offerOutpoint string, tradeDealInfo TableTradeDealInfo, transactionInfoBuy, transactionInfoSale TableTransactionInfo, rebateInfos []TableRebateInfo, recordsInfos []TableRecordsInfo) error {
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
			Where("account_id = ?", accountInfo.AccountId).Updates(accountInfo).Error; err != nil {
			return err
		}

		if err := tx.Where("outpoint = ?", offerOutpoint).Delete(&TableOfferInfo{}).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{
				"account_id", "account", "deal_type", "sell_chain_type", "sell_address",
				"buy_chain_type", "buy_address", "price_ckb", "price_usd",
			}),
		}).Create(&tradeDealInfo).Error; err != nil {
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
