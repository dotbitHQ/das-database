package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/nervosnetwork/ckb-sdk-go/types"
)

/*
* das database parser map handle
 */
func (b *BlockParser) registerTransactionHandle() {
	b.mapTransactionHandle = make(map[string]FuncTransactionHandle)
	b.mapTransactionHandle[dao.DasActionTransferBalance] = b.ActionTransferBalance
	b.mapTransactionHandle[common.DasActionConfig] = b.ActionConfigCell
	b.mapTransactionHandle[common.DasActionWithdrawFromWallet] = b.ActionWithdrawFromWallet
	b.mapTransactionHandle[common.DasActionTransfer] = b.ActionTransferPayment
	b.mapTransactionHandle[common.DasActionCreateIncome] = b.ActionCreateIncome
	b.mapTransactionHandle[common.DasActionConsolidateIncome] = b.ActionConsolidateIncome

	b.mapTransactionHandle[common.DasActionStartAccountSale] = b.ActionStartAccountSale
	b.mapTransactionHandle[common.DasActionEditAccountSale] = b.ActionEditAccountSale
	b.mapTransactionHandle[common.DasActionCancelAccountSale] = b.ActionCancelAccountSale
	b.mapTransactionHandle[common.DasActionBuyAccount] = b.ActionBuyAccount

	b.mapTransactionHandle[common.DasActionApplyRegister] = b.ActionApplyRegister
	b.mapTransactionHandle[common.DasActionPreRegister] = b.ActionPreRegister
	b.mapTransactionHandle[common.DasActionPropose] = b.ActionPropose
	b.mapTransactionHandle[common.DasActionExtendPropose] = b.ActionPropose
	b.mapTransactionHandle[common.DasActionConfirmProposal] = b.ActionConfirmProposal
	b.mapTransactionHandle[common.DasActionEditRecords] = b.ActionEditRecords
	b.mapTransactionHandle[common.DasActionEditManager] = b.ActionEditManager
	b.mapTransactionHandle[common.DasActionRenewAccount] = b.ActionRenewAccount
	b.mapTransactionHandle[common.DasActionTransferAccount] = b.ActionTransferAccount

	b.mapTransactionHandle[common.DasActionDeclareReverseRecord] = b.ActionDeclareReverseRecord
	b.mapTransactionHandle[common.DasActionRedeclareReverseRecord] = b.ActionRedeclareReverseRecord
	b.mapTransactionHandle[common.DasActionRetractReverseRecord] = b.ActionRetractReverseRecord

	b.mapTransactionHandle[common.DasActionMakeOffer] = b.ActionMakeOffer
	b.mapTransactionHandle[common.DasActionEditOffer] = b.ActionEditOffer
	b.mapTransactionHandle[common.DasActionCancelOffer] = b.ActionCancelOffer
	b.mapTransactionHandle[common.DasActionAcceptOffer] = b.ActionAcceptOffer
}

func isCurrentVersionTx(tx *types.Transaction, name common.DasContractName) (bool, error) {
	contract, err := core.GetDasContractInfo(name)
	if err != nil {
		return false, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	isCV := false
	for _, v := range tx.Outputs {
		if v.Type == nil {
			continue
		}
		if contract.IsSameTypeId(v.Type.CodeHash) {
			isCV = true
			break
		}
	}
	return isCV, nil
}

type FuncTransactionHandleReq struct {
	DbDao          *dao.DbDao
	Tx             *types.Transaction
	TxHash         string
	BlockNumber    uint64
	BlockTimestamp uint64
	Action         common.DasAction
}

type FuncTransactionHandleResp struct {
	ActionName string
	Err        error
}

type FuncTransactionHandle func(FuncTransactionHandleReq) FuncTransactionHandleResp
