package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/scorpiotzh/toolib"
)

func (b *BlockParser) ActionTransferParser(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	log.Info("ActionTransferParser:", req.BlockNumber, req.TxHash, req.Action)

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
		_, _, oCT, _, oA, _ := core.FormatDasLockToHexAddress(v.Lock.Args)
		transactionInfos = append(transactionInfos, dao.TableTransactionInfo{
			BlockNumber:    req.BlockNumber,
			Action:         req.Action,
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
