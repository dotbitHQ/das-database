package block_parser

import (
	"das_database/config"
	"das_database/dao"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"strconv"
)

func (b *BlockParser) ActionEditDidCellRecords(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameDidCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version edit didcell records tx")
		return
	}
	log.Info("ActionEditDidCellRecords:", req.BlockNumber, req.TxHash, req.Action)

	txDidEntity, err := witness.TxToDidEntity(req.Tx)
	if err != nil {
		resp.Err = fmt.Errorf("witness.TxToDidEntity err: %s", err.Error())
		return
	}

	account, _, err := witness.GetAccountAndExpireFromDidCellData(req.Tx.OutputsData[txDidEntity.Outputs[0].Target.Index])
	if err != nil {
		resp.Err = fmt.Errorf("config.GetDidCellAccountAndExpire err: %s", err.Error())
		return
	}

	//account := didCellData.Account
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

	mode := address.Mainnet
	if config.Cfg.Server.Net != common.DasNetTypeMainNet {
		mode = address.Testnet
	}
	anyLockAddr, err := address.ConvertScriptToAddress(mode, req.Tx.Outputs[txDidEntity.Outputs[0].Target.Index].Lock)
	if err != nil {
		resp.Err = fmt.Errorf("address.ConvertScriptToAddress err: %s", err.Error())
		return
	}
	txInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      accountId,
		Account:        account,
		Action:         common.DidCellActionEditRecords,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      common.ChainTypeAnyLock,
		Address:        anyLockAddr,
		Capacity:       0,
		Outpoint:       didCellInfo.Outpoint,
		BlockTimestamp: req.BlockTimestamp,
	}

	if err := b.dbDao.CreateDidCellRecordsInfos(oldDidCellOutpoint, didCellInfo, recordsInfos, txInfo); err != nil {
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
		log.Warn("not current version edit didcell owner")
		return
	}
	log.Info("ActionEditDidCellOwner:", req.BlockNumber, req.TxHash, req.Action)

	didEntity, err := witness.TxToOneDidEntity(req.Tx, witness.SourceTypeOutputs)
	if err != nil {
		resp.Err = fmt.Errorf("TxToOneDidEntity err: %s", err.Error())
		return
	}

	account, _, err := witness.GetAccountAndExpireFromDidCellData(req.Tx.OutputsData[didEntity.Target.Index])
	if err != nil {
		resp.Err = fmt.Errorf("config.GetDidCellAccountAndExpire err: %s", err.Error())
		return
	}

	didCellArgs := common.Bytes2Hex(req.Tx.Outputs[didEntity.Target.Index].Lock.Args)
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(account))
	didCellInfo := dao.TableDidCellInfo{
		BlockNumber:  req.BlockNumber,
		Outpoint:     common.OutPoint2String(req.TxHash, uint(didEntity.Target.Index)),
		AccountId:    accountId,
		Args:         didCellArgs,
		LockCodeHash: req.Tx.Outputs[didEntity.Target.Index].Lock.CodeHash.Hex(),
	}

	oldOutpoint := common.OutPointStruct2String(req.Tx.Inputs[0].PreviousOutput)

	mode := address.Mainnet
	if config.Cfg.Server.Net != common.DasNetTypeMainNet {
		mode = address.Testnet
	}
	anyLockAddr, err := address.ConvertScriptToAddress(mode, req.Tx.Outputs[didEntity.Target.Index].Lock)
	if err != nil {
		resp.Err = fmt.Errorf("address.ConvertScriptToAddress err: %s", err.Error())
		return
	}
	txInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      accountId,
		Account:        account,
		Action:         common.DidCellActionEditOwner,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      common.ChainTypeAnyLock,
		Address:        anyLockAddr,
		Capacity:       0,
		Outpoint:       didCellInfo.Outpoint,
		BlockTimestamp: req.BlockTimestamp,
	}

	if err := b.dbDao.EditDidCellOwner(oldOutpoint, didCellInfo, txInfo); err != nil {
		log.Error("EditDidCellOwner err:", err.Error())
		resp.Err = fmt.Errorf("EditDidCellOwner err: %s", err.Error())
	}
	return
}

func (b *BlockParser) ActionDidCellRecycle(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	didEntity, err := witness.TxToOneDidEntity(req.Tx, witness.SourceTypeInputs)
	if err != nil {
		resp.Err = fmt.Errorf("TxToOneDidEntity err: %s", err.Error())
		return
	}
	preTx, err := b.dasCore.Client().GetTransaction(b.ctx, req.Tx.Inputs[didEntity.Target.Index].PreviousOutput.TxHash)
	if err != nil {
		resp.Err = fmt.Errorf("GetTransaction err: %s", err.Error())
		return
	}

	if isCV, err := isCurrentVersionTx(preTx.Transaction, common.DasContractNameDidCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version didcell recycle")
		return
	}
	log.Info("ActionDidCellRecycle:", req.BlockNumber, req.TxHash, req.Action)
	account, _, err := witness.GetAccountAndExpireFromDidCellData(preTx.Transaction.OutputsData[req.Tx.Inputs[didEntity.Target.Index].PreviousOutput.Index])
	if err != nil {
		resp.Err = fmt.Errorf("config.GetDidCellAccountAndExpire err: %s", err.Error())
		return
	}

	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(account))
	oldOutpoint := common.OutPointStruct2String(req.Tx.Inputs[0].PreviousOutput)

	mode := address.Mainnet
	if config.Cfg.Server.Net != common.DasNetTypeMainNet {
		mode = address.Testnet
	}
	anyLockAddr, err := address.ConvertScriptToAddress(mode, preTx.Transaction.Outputs[req.Tx.Inputs[didEntity.Target.Index].PreviousOutput.Index].Lock)
	if err != nil {
		resp.Err = fmt.Errorf("address.ConvertScriptToAddress err: %s", err.Error())
		return
	}
	txInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      accountId,
		Account:        account,
		Action:         common.DidCellActionRecycle,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      common.ChainTypeAnyLock,
		Address:        anyLockAddr,
		Capacity:       0,
		Outpoint:       common.OutPoint2String(req.TxHash, 0),
		BlockTimestamp: req.BlockTimestamp,
	}

	if err := b.dbDao.DidCellRecycle(oldOutpoint, accountId, txInfo); err != nil {
		log.Error("DidCellRecycle err:", err.Error())
		resp.Err = fmt.Errorf("DidCellRecycle err: %s", err.Error())
	}
	return
}
