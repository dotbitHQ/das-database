package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/scorpiotzh/toolib"
)

func (b *BlockParser) ActionBalanceCells(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	log.Info("ActionBalanceCells:", req.BlockNumber, req.TxHash, req.Action)

	dasLock, err := core.GetDasContractInfo(common.DasContractNameDispatchCellType)
	if err != nil {
		resp.Err = fmt.Errorf("GetDasContractInfo err: %s", err.Error())
		return
	}
	dasBalance, err := core.GetDasContractInfo(common.DasContractNameBalanceCellType)
	if err != nil {
		resp.Err = fmt.Errorf("GetDasContractInfo err: %s", err.Error())
		return
	}
	serviceType := dao.ServiceTypeRegister
	if req.Action == dao.DasActionTransferBalance {
		serviceType = dao.ServiceTypeTransaction
	}

	var transactionInfos []dao.TableTransactionInfo
	for i, v := range req.Tx.Outputs {
		if v.Lock.CodeHash.Hex() != dasLock.ContractTypeId.Hex() {
			continue
		}
		if req.Action == dao.DasActionTransferBalance && v.Type != nil && v.Type.CodeHash.Hex() != dasBalance.ContractTypeId.Hex() {
			continue
		}
		oldHex, _, err := b.dasCore.Daf().ArgsToHex(v.Lock.Args)
		if err != nil {
			resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
			return
		}
		transactionInfos = append(transactionInfos, dao.TableTransactionInfo{
			BlockNumber:    req.BlockNumber,
			Action:         req.Action,
			ServiceType:    serviceType,
			ChainType:      oldHex.ChainType,
			Address:        oldHex.AddressHex,
			Capacity:       v.Capacity,
			Outpoint:       common.OutPoint2String(req.TxHash, uint(i)),
			BlockTimestamp: req.BlockTimestamp,
		})
	}

	if err = b.dbDao.CreateTransactionInfoList(transactionInfos); err != nil {
		log.Error("CreateTransactionInfoList err: ", err.Error(), toolib.JsonString(transactionInfos))
		resp.Err = fmt.Errorf("CreateTransactionInfoList err: %s", err.Error())
		return
	}

	return
}

func (b *BlockParser) ActionBalanceCell(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	log.Info("ActionBalanceCell:", req.BlockNumber, req.TxHash, req.Action)

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
	serviceType := dao.ServiceTypeRegister
	if req.Action == common.DasActionWithdrawFromWallet {
		serviceType = dao.ServiceTypeTransaction
	}

	res, err := b.dasCore.Client().GetTransaction(b.ctx, req.Tx.Inputs[0].PreviousOutput.TxHash)
	if err != nil {
		resp.Err = fmt.Errorf("GetTransaction err: %s", err.Error())
		return
	}
	output := res.Transaction.Outputs[req.Tx.Inputs[0].PreviousOutput.Index]
	if !dasLock.IsSameTypeId(output.Lock.CodeHash) {
		log.Warn("ActionBalanceCell: das lock not match", req.TxHash)
		return
	}
	if output.Type != nil && !balanceType.IsSameTypeId(output.Type.CodeHash) {
		log.Warn("ActionBalanceCell: balance type not match", req.TxHash)
		return
	}

	oHex, _, err := b.dasCore.Daf().ArgsToHex(output.Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}
	tx := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		Action:         req.Action,
		ServiceType:    serviceType,
		ChainType:      oHex.ChainType,
		Address:        oHex.AddressHex,
		Capacity:       req.Tx.Outputs[0].Capacity,
		Outpoint:       common.OutPoint2String(req.TxHash, 0),
		BlockTimestamp: req.BlockTimestamp,
	}
	if err := b.dbDao.CreateTransactionInfo(tx); err != nil {
		log.Error("CreateTransactionInfo err:", err.Error(), toolib.JsonString(tx))
		resp.Err = fmt.Errorf("WithdrawFromWallet err: %s", err.Error())
		return
	}

	return
}
