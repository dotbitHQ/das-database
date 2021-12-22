package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/scorpiotzh/toolib"
)

func (b *BlockParser) ActionTransferPayment(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	dasLock, err := core.GetDasContractInfo(common.DasContractNameDispatchCellType)
	if err != nil {
		resp.Err = fmt.Errorf("GetDasContractInfo err: %s", err.Error())
		return
	}

	balanceType, err := core.GetDasContractInfo(common.DasContractNameBalanceCellType)
	if err != nil {
		resp.Err = fmt.Errorf("GetDasContractInfo err: %s", err.Error())
		return
	}

	log.Info("ActionTransferPayment:", req.TxHash)

	res, err := b.ckbClient.GetTxByHashOnChain(req.Tx.Inputs[0].PreviousOutput.TxHash)
	if err != nil {
		resp.Err = fmt.Errorf("GetTxByHashOnChain err: %s", err.Error())
		return
	}
	cellOutput := res.Transaction.Outputs[req.Tx.Inputs[0].PreviousOutput.Index]
	if !dasLock.IsSameTypeId(cellOutput.Lock.CodeHash) {
		log.Warn("ActionTransferPayment: das lock not match", req.TxHash)
		return
	}
	if cellOutput.Type != nil && !balanceType.IsSameTypeId(cellOutput.Type.CodeHash) {
		log.Warn("ActionTransferPayment: balance type not match", req.TxHash)
		return
	}

	_, _, oCT, _, oA, _ := core.FormatDasLockToHexAddress(cellOutput.Lock.Args)
	tx := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		Action:         common.DasActionTransfer,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      oCT,
		Address:        oA,
		Capacity:       req.Tx.Outputs[0].Capacity,
		Outpoint:       common.OutPoint2String(req.TxHash, 0),
		BlockTimestamp: req.BlockTimestamp,
	}
	if err := b.dbDao.CreateTransactionInfo2(tx); err != nil {
		log.Error("CreateTransactionInfo err:", err.Error(), toolib.JsonString(tx))
		resp.Err = fmt.Errorf("ActionTransferPayment err: %s", err.Error())
		return
	}

	return
}
