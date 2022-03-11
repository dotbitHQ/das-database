package dao

import (
	"github.com/DeAccountSystems/das-lib/common"
	"gorm.io/gorm/clause"
	"time"
)

type TableTransactionInfo struct {
	Id             uint64           `json:"id" gorm:"column:id;primaryKey;type:bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT ''"`
	BlockNumber    uint64           `json:"block_number" gorm:"column:block_number;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	AccountId      string           `json:"account_id" gorm:"account_id;index:k_ai_a;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT 'hash of account'"`
	Account        string           `json:"account" gorm:"column:account;index:k_a_a;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	Action         string           `json:"action" gorm:"column:action;index:k_ct_a_a;index:k_a_a;index:k_ai_a;uniqueIndex:uk_a_o;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	ServiceType    int              `json:"service_type" gorm:"column:service_type;type:smallint(6) NOT NULL DEFAULT '0' COMMENT '1: register 2: trade'"`
	ChainType      common.ChainType `json:"chain_type" gorm:"column:chain_type;index:k_ct_a_a;index:k_ct_a;type:smallint(6) NOT NULL DEFAULT '0' COMMENT ''"`
	Address        string           `json:"address" gorm:"column:address;index:k_ct_a_a;index:k_ct_a;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	Capacity       uint64           `json:"capacity" gorm:"column:capacity;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	Outpoint       string           `json:"outpoint" gorm:"column:outpoint;index:k_outpoint;uniqueIndex:uk_a_o;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	BlockTimestamp uint64           `json:"block_timestamp" gorm:"column:block_timestamp;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	Status         int              `json:"status" gorm:"column:status;type:smallint(6) NOT NULL DEFAULT '0' COMMENT '0: normal -1: rejected'"`
	CreatedAt      time.Time        `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''"`
	UpdatedAt      time.Time        `json:"updated_at" gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
}

const (
	TableNameTransactionInfo = "t_transaction_info"

	ServiceTypeRegister    = 1
	ServiceTypeTransaction = 2
)

func (t *TableTransactionInfo) TableName() string {
	return TableNameTransactionInfo
}

type TxAction = int

const (
	ActionUndefined          TxAction = 99
	ActionWithdrawFromWallet TxAction = 0
	ActionConsolidateIncome  TxAction = 1 // merge reward
	ActionStartAccountSale   TxAction = 2
	ActionEditAccountSale    TxAction = 3
	ActionCancelAccountSale  TxAction = 4
	ActionBuyAccount         TxAction = 5
	ActionSaleAccount        TxAction = 6
	ActionTransferBalance    TxAction = 7 // active balance

	ActionDeclareReverseRecord   TxAction = 8  // set reverse records
	ActionRedeclareReverseRecord TxAction = 9  // edit reverse records
	ActionRetractReverseRecord   TxAction = 10 // delete reverse records

	ActionDasBalanceTransfer TxAction = 11 // transfer or use DAS's balance to register DAS account
	ActionEditRecords        TxAction = 12 // edit records
	ActionTransferAccount    TxAction = 13 // edit owner
	ActionEditManager        TxAction = 14 // edit manager
	ActionRenewAccount       TxAction = 15
	ActionAcceptOffer        TxAction = 16 // taker
	ActionOfferAccepted      TxAction = 17 // maker
	ActionEditOfferAdd       TxAction = 18 // add offer price
	ActionEditOfferSub       TxAction = 19 // sub offer price

	DasActionTransferBalance = "transfer_balance"
	DasActionSaleAccount     = "sale_account"
	DasActionOfferAccepted   = "offer_accepted"
	DasActionEditOfferAdd    = "offer_edit_add"
	DasActionEditOfferSub    = "offer_edit_sub"
)

func FormatTxAction(action string) TxAction {
	switch action {
	case common.DasActionWithdrawFromWallet:
		return ActionWithdrawFromWallet
	case common.DasActionConsolidateIncome:
		return ActionConsolidateIncome
	case common.DasActionStartAccountSale:
		return ActionStartAccountSale
	case common.DasActionEditAccountSale:
		return ActionEditAccountSale
	case common.DasActionCancelAccountSale:
		return ActionCancelAccountSale
	case common.DasActionBuyAccount:
		return ActionBuyAccount
	case DasActionSaleAccount:
		return ActionSaleAccount
	case DasActionTransferBalance:
		return ActionTransferBalance
	case common.DasActionDeclareReverseRecord:
		return ActionDeclareReverseRecord
	case common.DasActionRedeclareReverseRecord:
		return ActionRedeclareReverseRecord
	case common.DasActionRetractReverseRecord:
		return ActionRetractReverseRecord
	case common.DasActionTransfer:
		return ActionDasBalanceTransfer
	case common.DasActionEditRecords:
		return ActionEditRecords
	case common.DasActionTransferAccount:
		return ActionTransferAccount
	case common.DasActionEditManager:
		return ActionEditManager
	case common.DasActionRenewAccount:
		return ActionRenewAccount
	case common.DasActionAcceptOffer:
		return ActionAcceptOffer
	case DasActionOfferAccepted:
		return ActionOfferAccepted
	case DasActionEditOfferAdd:
		return ActionEditOfferAdd
	case DasActionEditOfferSub:
		return ActionEditOfferSub
	}

	return ActionUndefined
}

func FormatActionType(actionType TxAction) string {
	switch actionType {
	case ActionWithdrawFromWallet:
		return common.DasActionWithdrawFromWallet
	case ActionConsolidateIncome:
		return common.DasActionConsolidateIncome
	case ActionStartAccountSale:
		return common.DasActionStartAccountSale
	case ActionEditAccountSale:
		return common.DasActionEditAccountSale
	case ActionCancelAccountSale:
		return common.DasActionCancelAccountSale
	case ActionBuyAccount:
		return common.DasActionBuyAccount
	case ActionSaleAccount:
		return DasActionSaleAccount
	case ActionTransferBalance:
		return DasActionTransferBalance
	case ActionDeclareReverseRecord:
		return common.DasActionDeclareReverseRecord
	case ActionRedeclareReverseRecord:
		return common.DasActionRedeclareReverseRecord
	case ActionRetractReverseRecord:
		return common.DasActionRetractReverseRecord
	case ActionDasBalanceTransfer:
		return common.DasActionTransfer
	case ActionEditRecords:
		return common.DasActionEditRecords
	case ActionTransferAccount:
		return common.DasActionTransferAccount
	case ActionEditManager:
		return common.DasActionEditManager
	case ActionRenewAccount:
		return common.DasActionRenewAccount
	case ActionAcceptOffer:
		return common.DasActionAcceptOffer
	case ActionOfferAccepted:
		return DasActionOfferAccepted
	case ActionEditOfferAdd:
		return DasActionEditOfferAdd
	case ActionEditOfferSub:
		return DasActionEditOfferSub
	}

	return ""
}

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
