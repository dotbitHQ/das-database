package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/witness"
	"strconv"
)

func (b *BlockParser) ActionEditDidCellRecords(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameDidCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version account cross chain tx")
		return
	}
	log.Info("ActionEditDidCellRecords:", req.BlockNumber, req.TxHash, req.Action)

	txDidEntity, err := witness.TxToDidEntity(req.Tx)
	if err != nil {
		resp.Err = fmt.Errorf("witness.TxToDidEntity err: %s", err.Error())
		return
	}

	var didCellData witness.DidCellData
	if err := didCellData.BysToObj(req.Tx.OutputsData[txDidEntity.Outputs[0].Target.Index]); err != nil {
		resp.Err = fmt.Errorf("didCellData.BysToObj err: %s", err.Error())
		return
	}

	account := didCellData.Account
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(account))
	var recordsInfos []dao.TableRecordsInfo
	recordList := txDidEntity.Outputs[0].DidCellWitnessDataV0.Records
	for _, v := range recordList {
		recordsInfos = append(recordsInfos, dao.TableRecordsInfo{
			AccountId: accountId,
			Account:   account,
			Key:       v.Key,
			Type:      v.Type,
			Label:     v.Label,
			Value:     v.Value,
			Ttl:       strconv.FormatUint(uint64(v.TTL), 10),
		})
	}
	oldDidCellOutpoint := common.OutPointStruct2String(req.Tx.Inputs[txDidEntity.Inputs[0].Target.Index].PreviousOutput)
	var didCellInfo dao.TableDidCellInfo
	didCellInfo.AccountId = accountId
	didCellInfo.BlockNumber = req.BlockNumber
	didCellInfo.Outpoint = common.OutPoint2String(req.Tx.Hash.Hex(), uint(txDidEntity.Outputs[0].Target.Index))
	if err := b.dbDao.CreateDidCellRecordsInfos(oldDidCellOutpoint, didCellInfo, recordsInfos); err != nil {
		log.Error("CreateDidCellRecordsInfos err:", err.Error())
		resp.Err = fmt.Errorf("CreateDidCellRecordsInfos err: %s", err.Error())
	}

	return
}

func (b *BlockParser) ActionEditDidCellOwner(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameDidCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version account cross chain tx")
		return
	}
	log.Info("ActionEditDidCellOwner:", req.BlockNumber, req.TxHash, req.Action)
	//transfer：获取output里didcell的args
	//renew：获取input里的didcell的args（更新t_did_cell 的expired_at）
	didEntity, err := witness.TxToOneDidEntity(req.Tx, witness.SourceTypeOutputs)
	if err != nil {
		resp.Err = fmt.Errorf("TxToOneDidEntity err: %s", err.Error())
		return
	}
	var didCellData witness.DidCellData
	if err := didCellData.BysToObj(req.Tx.OutputsData[didEntity.Target.Index]); err != nil {
		resp.Err = fmt.Errorf("didCellData.BysToObj err: %s", err.Error())
		return
	}
	didCellArgs := common.Bytes2Hex(req.Tx.Outputs[didEntity.Target.Index].Lock.Args)
	account := didCellData.Account
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(account))
	didCellInfo := dao.TableDidCellInfo{
		BlockNumber:  req.BlockNumber,
		Outpoint:     common.OutPoint2String(req.TxHash, uint(didEntity.Target.Index)),
		AccountId:    accountId,
		Args:         didCellArgs,
		LockCodeHash: req.Tx.Outputs[didEntity.Target.Index].Lock.CodeHash.Hex(),
	}

	oldOutpoint := common.OutPointStruct2String(req.Tx.Inputs[0].PreviousOutput)
	if err := b.dbDao.EditDidCellOwner(oldOutpoint, didCellInfo); err != nil {
		log.Error("EditDidCellOwner err:", err.Error())
		resp.Err = fmt.Errorf("EditDidCellOwner err: %s", err.Error())
	}
	return
}

func (b *BlockParser) ActionDidCellRecycle(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameDidCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version account cross chain tx")
		return
	}
	log.Info("ActionAccountCrossChain:", req.BlockNumber, req.TxHash, req.Action)
	didEntity, err := witness.TxToOneDidEntity(req.Tx, witness.SourceTypeOutputs)
	if err != nil {
		resp.Err = fmt.Errorf("TxToOneDidEntity err: %s", err.Error())
		return
	}
	var didCellData witness.DidCellData
	if err := didCellData.BysToObj(req.Tx.OutputsData[didEntity.Target.Index]); err != nil {
		resp.Err = fmt.Errorf("didCellData.BysToObj err: %s", err.Error())
		return
	}
	account := didCellData.Account
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(account))

	oldOutpoint := common.OutPointStruct2String(req.Tx.Inputs[0].PreviousOutput)
	if err := b.dbDao.DidCellRecycle(oldOutpoint, accountId); err != nil {
		log.Error("DidCellRecycle err:", err.Error())
		resp.Err = fmt.Errorf("DidCellRecycle err: %s", err.Error())
	}
	return
}
