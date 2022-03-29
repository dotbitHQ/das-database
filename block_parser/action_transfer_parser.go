package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/scorpiotzh/toolib"
)

func (b *BlockParser) ActionTransferBalance(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	resp = b.actionTransferParser(req, "ActionTransferBalance:", dao.DasActionTransferBalance, dao.ServiceTypeTransaction)
	return
}

func (b *BlockParser) ActionOrderRefund(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	resp = b.actionTransferParser(req, "ActionOrderRefund:", dao.DasActionOrderRefund, dao.ServiceTypeRegister)
	return
}

func (b *BlockParser) actionTransferParser(req FuncTransactionHandleReq, funcName, action string, serviceType int) (resp FuncTransactionHandleResp) {
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

	log.Info(funcName, req.TxHash)

	var transactionInfos []dao.TableTransactionInfo
	for i, v := range req.Tx.Outputs {
		if v.Lock.CodeHash.Hex() != dasLock.ContractTypeId.Hex() {
			continue
		}
		if v.Type != nil && v.Type.CodeHash.Hex() != dasBalance.ContractTypeId.Hex() {
			continue
		}

		_, _, oCT, _, oA, _ := core.FormatDasLockToHexAddress(v.Lock.Args)
		transactionInfos = append(transactionInfos, dao.TableTransactionInfo{
			BlockNumber:    req.BlockNumber,
			Action:         action,
			ServiceType:    serviceType,
			ChainType:      oCT,
			Address:        oA,
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
