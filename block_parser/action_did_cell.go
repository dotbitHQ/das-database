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
	var recordsInfos []dao.TableRecordsInfo
	account := didCellData.Account
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(account))
	recordList := didEntity.DidCellWitnessDataV0.Records
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
	log.Info("ActionEditDidRecords:", account)

	if err := b.dbDao.CreateDidCellRecordsInfos(accountId, recordsInfos); err != nil {
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
	didCellArgs := common.Bytes2Hex(req.Tx.Outputs[didEntity.Target.Index].Lock.Args)
	account := didCellData.Account
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(account))
	didCellInfo := dao.TableDidCellInfo{
		BlockNumber: req.BlockNumber,
		Outpoint:    common.OutPoint2String(req.TxHash, 0),
		AccountId:   accountId,
		Args:        didCellArgs,
	}
	if err := b.dbDao.EditDidCellOwner(didCellInfo); err != nil {
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
	didCellArgs := common.Bytes2Hex(req.Tx.Outputs[didEntity.Target.Index].Lock.Args)
	account := didCellData.Account
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(account))
	var didCellInfo dao.TableDidCellInfo
	didCellInfo.Args = didCellArgs
	didCellInfo.AccountId = accountId
	if err := b.dbDao.DidCellRecycle(didCellInfo); err != nil {
		log.Error("DidCellRecycle err:", err.Error())
		resp.Err = fmt.Errorf("DidCellRecycle err: %s", err.Error())
	}
	return
}
