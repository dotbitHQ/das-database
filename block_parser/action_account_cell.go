package block_parser

import (
	"bytes"
	"das_database/dao"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/scorpiotzh/toolib"
	"strconv"
)

func (b *BlockParser) ActionEditRecords(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version edit records tx")
		return
	}
	log.Info("ActionEditRecords:", req.BlockNumber, req.TxHash)

	accBuilder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	var recordsInfos []dao.TableRecordsInfo
	account := accBuilder.Account
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(account))
	recordList := accBuilder.Records
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
	accountInfo := dao.TableAccountInfo{
		BlockNumber: req.BlockNumber,
		Outpoint:    common.OutPoint2String(req.TxHash, uint(accBuilder.Index)),
		AccountId:   accountId,
	}
	_, mHex, err := b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[accBuilder.Index].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}

	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      accountId,
		Account:        account,
		Action:         common.DasActionEditRecords,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      mHex.ChainType,
		Address:        mHex.AddressHex,
		Capacity:       0,
		Outpoint:       common.OutPoint2String(req.TxHash, uint(accBuilder.Index)),
		BlockTimestamp: req.BlockTimestamp,
	}

	log.Info("ActionEditRecords:", account, transactionInfo.Address)

	if err := b.dbDao.CreateRecordsInfos(accountInfo, recordsInfos, transactionInfo); err != nil {
		log.Error("CreateRecordsInfos err:", err.Error(), toolib.JsonString(transactionInfo))
		resp.Err = fmt.Errorf("CreateRecordsInfos err: %s", err.Error())
	}

	return
}

func (b *BlockParser) ActionEditManager(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version edit manager tx")
		return
	}
	log.Info("ActionEditManager:", req.BlockNumber, req.TxHash)

	accBuilder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	account := accBuilder.Account
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(account))
	ownerHex, managerHex, err := b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[accBuilder.Index].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}
	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      accountId,
		Account:        account,
		Action:         common.DasActionEditManager,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      ownerHex.ChainType,
		Address:        ownerHex.AddressHex,
		Capacity:       0,
		Outpoint:       common.OutPoint2String(req.TxHash, uint(accBuilder.Index)),
		BlockTimestamp: req.BlockTimestamp,
	}
	accountInfo := dao.TableAccountInfo{
		BlockNumber:        req.BlockNumber,
		Outpoint:           common.OutPoint2String(req.TxHash, uint(accBuilder.Index)),
		Account:            account,
		AccountId:          accountId,
		ManagerChainType:   managerHex.ChainType,
		Manager:            managerHex.AddressHex,
		ManagerAlgorithmId: managerHex.DasAlgorithmId,
		ManagerSubAid:      managerHex.DasSubAlgorithmId,
	}
	var cidPk dao.TableCidPk
	if managerHex.DasAlgorithmId == common.DasAlgorithmIdWebauthn {
		cidPk.Cid = common.Bytes2Hex(managerHex.AddressPayload[:10])
		cidPk.Pk = common.Bytes2Hex(managerHex.AddressPayload[10:])
	}

	log.Info("ActionEditManager:", account, managerHex.DasAlgorithmId, managerHex.ChainType, managerHex.AddressHex, transactionInfo.Address)

	if err := b.dbDao.EditManager(accountInfo, transactionInfo, cidPk); err != nil {
		log.Error("EditManager err:", err.Error(), toolib.JsonString(transactionInfo))
		resp.Err = fmt.Errorf("EditManager err: %s", err.Error())
	}

	return
}

func (b *BlockParser) ActionRenewAccount(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version renew account tx")
		return
	}
	log.Info("ActionRenewAccount:", req.BlockNumber, req.TxHash)

	incomeContract, err := core.GetDasContractInfo(common.DasContractNameIncomeCellType)
	if err != nil {
		resp.Err = fmt.Errorf("GetDasContractInfo err: %s", err.Error())
		return
	}

	var inputsOutpoints []string
	var incomeCellInfos []dao.TableIncomeCellInfo
	for _, v := range req.Tx.Inputs {
		inputsOutpoints = append(inputsOutpoints, common.OutPoint2String(v.PreviousOutput.TxHash.Hex(), v.PreviousOutput.Index))
	}
	renewCapacity := uint64(0)
	for i, v := range req.Tx.Outputs {
		if v.Type == nil {
			continue
		}
		if incomeContract.IsSameTypeId(v.Type.CodeHash) {
			renewCapacity = v.Capacity
			incomeCellInfos = append(incomeCellInfos, dao.TableIncomeCellInfo{
				BlockNumber:    req.BlockNumber,
				Action:         common.DasActionRenewAccount,
				Outpoint:       common.OutPoint2String(req.TxHash, uint(i)),
				Capacity:       v.Capacity,
				BlockTimestamp: req.BlockTimestamp,
				Status:         dao.IncomeCellStatusUnMerge,
			})
		}
	}

	builder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}

	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(builder.Account))
	accountInfo := dao.TableAccountInfo{
		BlockNumber: req.BlockNumber,
		Outpoint:    common.OutPoint2String(req.TxHash, uint(builder.Index)),
		AccountId:   accountId,
		Account:     builder.Account,
		ExpiredAt:   builder.ExpiredAt,
	}

	ownerHex, _, err := b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[builder.Index].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}
	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      accountId,
		Account:        builder.Account,
		Action:         common.DasActionRenewAccount,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      ownerHex.ChainType,
		Address:        ownerHex.AddressHex,
		Capacity:       renewCapacity,
		Outpoint:       common.OutPoint2String(req.TxHash, uint(builder.Index)),
		BlockTimestamp: req.BlockTimestamp,
	}

	log.Info("ActionRenewAccount:", builder.Account, builder.ExpiredAt, transactionInfo.Capacity)

	//renew did cell
	var didCellInfo dao.TableDidCellInfo
	var oldDidCellOutpoint string
	if didCellAction, err := b.dasCore.TxToDidCellAction(req.Tx); err != nil {
		if err != core.ErrorNotExistDidCell {
			log.Error("TxToDidCellAction err :", err.Error())
		}
	} else if didCellAction == common.DidCellActionRenew {
		txDidEntity, err := witness.TxToDidEntity(req.Tx)
		if err != nil {
			resp.Err = fmt.Errorf("witness.TxToDidEntity err: %s", err.Error())
		} else {
			oldDidCellOutpoint = common.OutPointStruct2String(req.Tx.Inputs[txDidEntity.Inputs[0].Target.Index].PreviousOutput)
			didCellInfo.Outpoint = common.OutPoint2String(req.Tx.Hash.Hex(), uint(txDidEntity.Outputs[0].Target.Index))
			didCellInfo.ExpiredAt = accountInfo.ExpiredAt
			didCellInfo.BlockNumber = accountInfo.BlockNumber
		}
	}

	if err := b.dbDao.RenewAccount(inputsOutpoints, incomeCellInfos, accountInfo, transactionInfo, oldDidCellOutpoint, didCellInfo); err != nil {
		log.Error("RenewAccount err:", err.Error(), toolib.JsonString(transactionInfo))
		resp.Err = fmt.Errorf("RenewAccount err: %s", err.Error())
	}

	return
}

func (b *BlockParser) ActionBidExpiredAccountAuction(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version transfer account tx")
		return
	}
	log.Info("BidExpiredAccountAuction:", req.BlockNumber, req.TxHash)

	builder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	account := builder.Account
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(account))
	oHex, mHex, err := b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[builder.Index].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}

	res, err := b.dasCore.Client().GetTransaction(b.ctx, req.Tx.Inputs[builder.Index].PreviousOutput.TxHash)
	if err != nil {
		resp.Err = fmt.Errorf("GetTransaction err: %s", err.Error())
		return
	}

	oldHex, _, err := b.dasCore.Daf().ArgsToHex(res.Transaction.Outputs[req.Tx.Inputs[builder.Index].PreviousOutput.Index].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}
	transactionInfos := make([]dao.TableTransactionInfo, 0)

	transactionInfos = append(transactionInfos, dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      accountId,
		Account:        account,
		Action:         common.DasActionBidExpiredAccountAuction,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      oldHex.ChainType,
		Address:        oldHex.AddressHex,
		Capacity:       res.Transaction.Outputs[req.Tx.Inputs[builder.Index].PreviousOutput.Index].Capacity,
		Outpoint:       common.OutPoint2String(req.TxHash, uint(builder.Index)),
		BlockTimestamp: req.BlockTimestamp,
	})

	accountInfo := dao.TableAccountInfo{
		BlockNumber:        req.BlockNumber,
		Outpoint:           common.OutPoint2String(req.TxHash, uint(builder.Index)),
		AccountId:          accountId,
		Account:            account,
		OwnerChainType:     oHex.ChainType,
		Owner:              oHex.AddressHex,
		OwnerAlgorithmId:   oHex.DasAlgorithmId,
		OwnerSubAid:        oHex.DasSubAlgorithmId,
		ManagerChainType:   mHex.ChainType,
		Manager:            mHex.AddressHex,
		ManagerAlgorithmId: mHex.DasAlgorithmId,
		ManagerSubAid:      mHex.DasSubAlgorithmId,
		ExpiredAt:          builder.ExpiredAt,
		RegisteredAt:       builder.RegisteredAt,
		Status:             builder.Status,
	}
	log.Info("ActionBidExpiredAccountAuction:", accountInfo)

	var recordsInfos []dao.TableRecordsInfo

	recordList := builder.Records
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

	if err := b.dbDao.BidExpiredAccountAuction(accountInfo, recordsInfos, transactionInfos); err != nil {
		log.Error("ActionBidExpiredAccountAuction err:", err.Error(), toolib.JsonString(accountInfo))
		resp.Err = fmt.Errorf("ActionBidExpiredAccountAuction err: %s", err.Error())
	}
	return
}

func (b *BlockParser) ActionTransferAccount(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version transfer account tx")
		return
	}
	log.Info("ActionTransferAccount:", req.BlockNumber, req.TxHash)

	builder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	account := builder.Account
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(account))

	oHex, mHex, err := b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[builder.Index].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}
	oldBuilder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeOld)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	res, err := b.dasCore.Client().GetTransaction(b.ctx, req.Tx.Inputs[oldBuilder.Index].PreviousOutput.TxHash)
	if err != nil {
		resp.Err = fmt.Errorf("GetTransaction err: %s", err.Error())
		return
	}

	oldHex, _, err := b.dasCore.Daf().ArgsToHex(res.Transaction.Outputs[req.Tx.Inputs[oldBuilder.Index].PreviousOutput.Index].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}
	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      accountId,
		Account:        account,
		Action:         common.DasActionTransferAccount,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      oldHex.ChainType,
		Address:        oldHex.AddressHex,
		Capacity:       0,
		Outpoint:       common.OutPoint2String(req.TxHash, uint(builder.Index)),
		BlockTimestamp: req.BlockTimestamp,
	}

	// transfer accCell to didCell (for passkey)
	if builder.Status == common.AccountStatusOnUpgrade {
		txDidEntity, err := witness.TxToDidEntity(req.Tx)
		if err != nil {
			resp.Err = fmt.Errorf("witness.TxToDidEntity err: %s", err.Error())
			return
		}

		if len(txDidEntity.Inputs) == 0 && len(txDidEntity.Outputs) == 1 {
			didCellArgs := common.Bytes2Hex(req.Tx.Outputs[txDidEntity.Outputs[0].Target.Index].Lock.Args)
			accountInfo := dao.TableAccountInfo{
				BlockNumber: req.BlockNumber,
				Outpoint:    common.OutPoint2String(req.TxHash, uint(builder.Index)),
				AccountId:   builder.AccountId,
				Status:      builder.Status,
			}

			didCellInfo := dao.TableDidCellInfo{
				BlockNumber:  req.BlockNumber,
				Outpoint:     common.OutPoint2String(req.TxHash, uint(txDidEntity.Outputs[0].Target.Index)),
				AccountId:    builder.AccountId,
				Account:      builder.Account,
				Args:         didCellArgs,
				ExpiredAt:    builder.ExpiredAt,
				LockCodeHash: req.Tx.Outputs[txDidEntity.Outputs[0].Target.Index].Lock.CodeHash.Hex(),
			}

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
			if err := b.dbDao.TransferAccountToDid(accountInfo, didCellInfo, transactionInfo, recordsInfos); err != nil {
				log.Error("TransferAccountToDid err:", err.Error(), toolib.JsonString(transactionInfo))
				resp.Err = fmt.Errorf("TransferAccount err: %s", err.Error())
			}
			return
		}
	}

	accountInfo := dao.TableAccountInfo{
		BlockNumber:        req.BlockNumber,
		Outpoint:           common.OutPoint2String(req.TxHash, uint(builder.Index)),
		AccountId:          accountId,
		Account:            account,
		OwnerChainType:     oHex.ChainType,
		Owner:              oHex.AddressHex,
		OwnerAlgorithmId:   oHex.DasAlgorithmId,
		OwnerSubAid:        oHex.DasSubAlgorithmId,
		ManagerChainType:   mHex.ChainType,
		Manager:            mHex.AddressHex,
		ManagerAlgorithmId: mHex.DasAlgorithmId,
		ManagerSubAid:      mHex.DasSubAlgorithmId,
	}

	var cidPk dao.TableCidPk
	if oHex.DasAlgorithmId == common.DasAlgorithmIdWebauthn {
		cidPk.Cid = common.Bytes2Hex(oHex.AddressPayload[:10])
		cidPk.Pk = common.Bytes2Hex(oHex.AddressPayload[10:])
	}

	var recordsInfos []dao.TableRecordsInfo
	recordList := builder.Records
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

	log.Info("ActionTransferAccount:", account, oHex.DasAlgorithmId, oHex.ChainType, oHex.AddressHex, mHex.DasAlgorithmId, mHex.ChainType, mHex.AddressHex, transactionInfo.Address)

	if err := b.dbDao.TransferAccount(accountInfo, transactionInfo, recordsInfos, cidPk); err != nil {
		log.Error("TransferAccount err:", err.Error(), toolib.JsonString(transactionInfo))
		resp.Err = fmt.Errorf("TransferAccount err: %s", err.Error())
	}

	return
}

func (b *BlockParser) ActionForceRecoverAccountStatus(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version force recover account status tx")
		return
	}
	log.Info("ActionForceRecoverAccountStatus:", req.BlockNumber, req.TxHash)

	oldBuilder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeOld)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	builder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	oHex, _, err := b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[builder.Index].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}

	accountInfo := dao.TableAccountInfo{
		BlockNumber: req.BlockNumber,
		Outpoint:    common.OutPoint2String(req.TxHash, uint(builder.Index)),
		AccountId:   builder.AccountId,
		Status:      builder.Status,
	}
	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      builder.AccountId,
		Account:        builder.Account,
		Action:         common.DasActionForceRecoverAccountStatus,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      oHex.ChainType,
		Address:        oHex.AddressHex,
		Capacity:       req.Tx.OutputsCapacity() - req.Tx.Outputs[0].Capacity,
		Outpoint:       common.OutPoint2String(req.TxHash, 0),
		BlockTimestamp: req.BlockTimestamp,
	}

	log.Info("ActionForceRecoverAccountStatus:", builder.Account, oldBuilder.Status, builder.Status)

	if err = b.dbDao.ForceRecoverAccountStatus(oldBuilder.Status, accountInfo, transactionInfo); err != nil {
		resp.Err = fmt.Errorf("ForceRecoverAccountStatus err: %s", err.Error())
		return
	}

	return
}

func (b *BlockParser) ActionRecycleExpiredAccount(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version recycle expired account tx")
		return
	}
	log.Info("ActionRecycleExpiredAccount:", req.BlockNumber, req.TxHash)

	var builder *witness.AccountCellDataBuilder
	builderMap, err := witness.AccountCellDataBuilderMapFromTx(req.Tx, common.DataTypeOld)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderMapFromTx err: %s", err.Error())
		return
	}
	for _, v := range builderMap {
		if v.Index == 1 {
			builder = v
		}
	}
	if builder == nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilder is nil")
		return
	}

	res, err := b.dasCore.Client().GetTransaction(b.ctx, req.Tx.Inputs[1].PreviousOutput.TxHash)
	if err != nil {
		resp.Err = fmt.Errorf("GetTransaction err: %s", err.Error())
		return
	}
	oHex, _, err := b.dasCore.Daf().ArgsToHex(res.Transaction.Outputs[req.Tx.Inputs[1].PreviousOutput.Index].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}
	oArgs, err := b.dasCore.Daf().HexToArgs(oHex, oHex)
	if err != nil {
		resp.Err = fmt.Errorf("HexToArgs err: %s", err.Error())
		return
	}
	var oCapacity uint64
	for _, output := range req.Tx.Outputs[1:] {
		if bytes.EqualFold(oArgs, output.Lock.Args) {
			oCapacity += output.Capacity
		}
	}

	accountId, err := common.OutputDataToAccountId(req.Tx.OutputsData[0])
	if err != nil {
		resp.Err = fmt.Errorf("OutputDataToAccountId err: %s", err.Error())
		return
	}
	accountInfo := dao.TableAccountInfo{
		BlockNumber: req.BlockNumber,
		Outpoint:    common.OutPoint2String(req.TxHash, 0),
		AccountId:   common.Bytes2Hex(accountId),
	}
	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      builder.AccountId,
		Account:        builder.Account,
		Action:         common.DasActionRecycleExpiredAccount,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      oHex.ChainType,
		Address:        oHex.AddressHex,
		Capacity:       oCapacity,
		Outpoint:       common.OutPoint2String(req.TxHash, 0),
		BlockTimestamp: req.BlockTimestamp,
	}

	log.Info("ActionRecycleExpiredAccount:", builder.Account, oHex.DasAlgorithmId, oHex.ChainType, oHex.AddressHex)

	if err = b.dbDao.RecycleExpiredAccount(accountInfo, transactionInfo, builder.AccountId, builder.EnableSubAccount); err != nil {
		resp.Err = fmt.Errorf("RecycleExpiredAccount err: %s", err.Error())
		return
	}

	return
}

func (b *BlockParser) ActionAccountCrossChain(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version account cross chain tx")
		return
	}
	log.Info("ActionAccountCrossChain:", req.BlockNumber, req.TxHash, req.Action)

	builder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	ownerHex, managerHex, err := b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[0].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}
	var isTrans bool
	if req.Action == common.DasActionUnlockAccountForCrossChain {
		res, err := b.dasCore.Client().GetTransaction(b.ctx, req.Tx.Inputs[0].PreviousOutput.TxHash)
		if err != nil {
			resp.Err = fmt.Errorf("GetTransaction err: %s", err.Error())
			return
		}
		args := res.Transaction.Outputs[req.Tx.Inputs[0].PreviousOutput.Index].Lock.Args
		if !bytes.EqualFold(args, req.Tx.Outputs[0].Lock.Args) {
			isTrans = true
		}
	}

	accountInfo := dao.TableAccountInfo{
		BlockNumber:        req.BlockNumber,
		Outpoint:           common.OutPoint2String(req.TxHash, 0),
		AccountId:          builder.AccountId,
		OwnerChainType:     ownerHex.ChainType,
		Owner:              ownerHex.AddressHex,
		OwnerAlgorithmId:   ownerHex.DasAlgorithmId,
		OwnerSubAid:        ownerHex.DasSubAlgorithmId,
		ManagerChainType:   managerHex.ChainType,
		Manager:            managerHex.AddressHex,
		ManagerAlgorithmId: managerHex.DasAlgorithmId,
		ManagerSubAid:      managerHex.DasSubAlgorithmId,
		Status:             builder.Status,
	}
	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      builder.AccountId,
		Account:        builder.Account,
		Action:         req.Action,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      ownerHex.ChainType,
		Address:        ownerHex.AddressHex,
		Capacity:       0,
		Outpoint:       common.OutPoint2String(req.TxHash, 0),
		BlockTimestamp: req.BlockTimestamp,
	}

	if err = b.dbDao.AccountCrossChain(accountInfo, transactionInfo, isTrans); err != nil {
		log.Error("AccountCrossChain err:", err.Error(), req.TxHash, req.BlockNumber)
		resp.Err = fmt.Errorf("AccountCrossChain err: %s ", err.Error())
		return
	}

	return
}

func (b *BlockParser) ActionAccountUpgrade(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version account cross chain tx")
		return
	}
	fmt.Println("aaaaaaaaaaaaaaaaa")
	log.Info("ActionAccountCrossChain:", req.BlockNumber, req.TxHash, req.Action)

	builder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	ownerHex, _, err := b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[0].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}
	didEntity, err := witness.TxToOneDidEntity(req.Tx, witness.SourceTypeOutputs)
	if err != nil {
		resp.Err = fmt.Errorf("TxToOneDidEntity err: %s", err.Error())
		return
	}
	didCellArgs := common.Bytes2Hex(req.Tx.Outputs[didEntity.Target.Index].Lock.Args)
	accountInfo := dao.TableAccountInfo{
		BlockNumber: req.BlockNumber,
		Outpoint:    common.OutPoint2String(req.TxHash, 0),
		AccountId:   builder.AccountId,
		Status:      builder.Status,
	}

	didCellInfo := dao.TableDidCellInfo{
		BlockNumber:  req.BlockNumber,
		Outpoint:     common.OutPoint2String(req.TxHash, 0),
		AccountId:    builder.AccountId,
		Args:         didCellArgs,
		LockCodeHash: req.Tx.Outputs[didEntity.Target.Index].Lock.CodeHash.Hex(),
	}

	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      builder.AccountId,
		Account:        builder.Account,
		Action:         req.Action,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      ownerHex.ChainType,
		Address:        ownerHex.AddressHex,
		Capacity:       0,
		Outpoint:       common.OutPoint2String(req.TxHash, 0),
		BlockTimestamp: req.BlockTimestamp,
	}

	var recordsInfos []dao.TableRecordsInfo
	recordList := didEntity.DidCellWitnessDataV0.Records
	for _, v := range recordList {
		recordsInfos = append(recordsInfos, dao.TableRecordsInfo{
			AccountId: builder.AccountId,
			Account:   builder.Account,
			Key:       v.Key,
			Type:      v.Type,
			Label:     v.Label,
			Value:     v.Value,
			Ttl:       strconv.FormatUint(uint64(v.TTL), 10),
		})
	}
	if err = b.dbDao.AccountUpgrade(accountInfo, didCellInfo, transactionInfo, recordsInfos); err != nil {
		log.Error("AccountCrossChain err:", err.Error(), req.TxHash, req.BlockNumber)
		resp.Err = fmt.Errorf("AccountCrossChain err: %s ", err.Error())
		return
	}
	return
}
