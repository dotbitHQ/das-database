package dao

import (
	"github.com/DeAccountSystems/das-lib/common"
	"gorm.io/gorm/clause"
	"time"
)

type TableTransactionInfo struct {
	Id             uint64           `json:"id" gorm:"column:id;primary_key;AUTO_INCREMENT"`
	BlockNumber    uint64           `json:"block_number" gorm:"column:block_number"`
	AccountId      string           `json:"account_id" gorm:"account_id"`
	Account        string           `json:"account" gorm:"column:account"`
	Action         string           `json:"action" gorm:"column:action"`
	ServiceType    int              `json:"service_type" gorm:"column:service_type"` // 1: register 2: trade',
	ChainType      common.ChainType `json:"chain_type" gorm:"column:chain_type"`
	Address        string           `json:"address" gorm:"column:address"`
	Capacity       uint64           `json:"capacity" gorm:"column:capacity"`
	Outpoint       string           `json:"outpoint" gorm:"column:outpoint"`
	BlockTimestamp uint64           `json:"block_timestamp" gorm:"column:block_timestamp"`
	Status         int              `json:"status" gorm:"column:status"` // 0: normal -1: rejected
	CreatedAt      time.Time        `json:"created_at" gorm:"column:created_at"`
	UpdatedAt      time.Time        `json:"updated_at" gorm:"column:updated_at"`
}

const (
	TableNameTransactionInfo = "t_transaction_info"

	ServiceTypeRegister    = 1
	ServiceTypeTransaction = 2
)

func (t *TableTransactionInfo) TableName() string {
	return TableNameTransactionInfo
}

const (
	DasActionTransferBalance = "transfer_balance"
	DasActionSaleAccount     = "sale_account"
	DasActionOfferAccepted   = "offer_accepted"
	DasActionEditOfferAdd    = "offer_edit_add"
	DasActionEditOfferSub    = "offer_edit_sub"
	DasActionOrderRefund     = "order_refund"
)

func (d *DbDao) CreateTransactionInfo(transactionInfo TableTransactionInfo) error {
	return d.db.Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{
			"account_id", "account", "service_type",
			"chain_type", "address", "capacity", "status",
		}),
	}).Create(&transactionInfo).Error
}

func (d *DbDao) CreateTransactionInfoList(transactionInfos []TableTransactionInfo) error {
	if len(transactionInfos) > 0 {
		return d.db.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{
				"account_id", "account", "service_type",
				"chain_type", "address", "capacity", "status",
			}),
		}).Create(&transactionInfos).Error
	}

	return nil
}

func (d *DbDao) FindTransactionInfoByAccountAction(account, action string) (transactionInfo TableTransactionInfo, err error) {
	err = d.db.Where("account = ? AND action = ?", account, action).Limit(1).Find(&transactionInfo).Error
	return
}
