package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/DeAccountSystems/das-lib/molecule"
	"github.com/DeAccountSystems/das-lib/witness"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"strconv"
)

func (b *BlockParser) ActionConfirmProposal(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version confirm proposal tx")
		return
	}
	log.Info("ActionConfirmProposal:", req.BlockNumber, req.TxHash)

	// add income cell infos
	incomeContract, err := core.GetDasContractInfo(common.DasContractNameIncomeCellType)
	if err != nil {
		resp.Err = fmt.Errorf("GetDasContractInfo err: %s", err.Error())
		return
	}
	var incomeCellInfos []dao.TableIncomeCellInfo
	for i, v := range req.Tx.Outputs {
		if v.Type == nil {
			continue
		}
		if incomeContract.IsSameTypeId(v.Type.CodeHash) {
			incomeCellInfos = append(incomeCellInfos, dao.TableIncomeCellInfo{
				BlockNumber:    req.BlockNumber,
				Action:         common.DasActionConfirmProposal,
				Outpoint:       common.OutPoint2String(req.TxHash, uint(i)),
				Capacity:       v.Capacity,
				BlockTimestamp: req.BlockTimestamp,
				Status:         dao.IncomeCellStatusUnMerge,
			})
		}
	}

	preMap, err := witness.PreAccountCellDataBuilderMapFromTx(req.Tx, common.DataTypeOld) //new account
	if err != nil {
		resp.Err = fmt.Errorf("PreAccountCellDataBuilderMapFromTx err: %s", err.Error())
		return
	}
	accMap, err := witness.AccountCellDataBuilderMapFromTx(req.Tx, common.DataTypeNew) //old+new account
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderMapFromTx err: %s", err.Error())
		return
	}

	log.Info("ActionConfirmProposal:", len(preMap), len(accMap))

	var accountInfos []dao.TableAccountInfo
	var transactionInfos []dao.TableTransactionInfo
	var rebateInfos []dao.TableRebateInfo
	// account basic store fee
	configCell, err := b.dasCore.ConfigCellDataBuilderByTypeArgsList(common.ConfigCellTypeArgsAccount, common.ConfigCellTypeArgsProfitRate)
	if err != nil {
		resp.Err = fmt.Errorf("ConfigCellDataBuilderByTypeArgs err: %s", err.Error())
		return
	}
	//basicCapacity, _ := configCell.BasicCapacity()
	// rebate rate
	profitRateInviter, _ := configCell.ProfitRateInviter()
	profitRateChannel, _ := configCell.ProfitRateChannel()

	for _, v := range accMap {
		ownerHex, managerHex, err := b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[v.Index].Lock.Args)
		if err != nil {
			resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
			return
		}
		accountInfos = append(accountInfos, dao.TableAccountInfo{
			BlockNumber:         req.BlockNumber,
			Outpoint:            common.OutPoint2String(req.TxHash, uint(v.Index)),
			AccountId:           v.AccountId,
			Account:             v.Account,
			OwnerChainType:      ownerHex.ChainType,
			Owner:               ownerHex.AddressHex,
			OwnerAlgorithmId:    ownerHex.DasAlgorithmId,
			ManagerChainType:    managerHex.ChainType,
			Manager:             managerHex.AddressHex,
			ManagerAlgorithmId:  managerHex.DasAlgorithmId,
			Status:              dao.AccountStatus(v.Status),
			RegisteredAt:        v.RegisteredAt,
			ExpiredAt:           v.ExpiredAt,
			ConfirmProposalHash: req.TxHash,
		})

		if preAcc, ok := preMap[v.Account]; ok {
			preTx, err := b.ckbClient.Client().GetTransaction(b.ctx, req.Tx.Inputs[preAcc.Index].PreviousOutput.TxHash)
			if err != nil {
				resp.Err = fmt.Errorf("pretx GetTransaction err: %s", err.Error())
				return
			}

			transactionInfos = append(transactionInfos, dao.TableTransactionInfo{
				BlockNumber:    req.BlockNumber,
				AccountId:      v.AccountId,
				Account:        v.Account,
				Action:         common.DasActionConfirmProposal,
				ServiceType:    dao.ServiceTypeRegister,
				ChainType:      ownerHex.ChainType,
				Address:        ownerHex.AddressHex,
				Capacity:       req.Tx.Outputs[v.Index].Capacity,
				Outpoint:       common.OutPoint2String(req.TxHash, uint(v.Index)),
				BlockTimestamp: req.BlockTimestamp,
			})

			argsStr, _ := preAcc.OwnerLockArgsStr()
			inviteeHex, _, err := b.dasCore.Daf().ArgsToHex(common.Hex2Bytes(argsStr))
			if err != nil {
				resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
				return
			}
			inviterId, _ := preAcc.InviterId()
			accLen := uint64(len([]byte(preAcc.Account))) * common.OneCkb

			basicCapacity, _ := configCell.BasicCapacityFromOwnerDasAlgorithmId(argsStr)
			log.Info("ActionConfirmProposal:", basicCapacity, profitRateInviter, profitRateChannel)

			preCapacity := preTx.Transaction.Outputs[req.Tx.Inputs[preAcc.Index].PreviousOutput.Index].Capacity - basicCapacity - accLen // 扣除存储费，账号长度
			capacity, _ := decimal.NewFromString(fmt.Sprintf("%d", preCapacity))

			inviterLock, _ := preAcc.InviterLock()
			if inviterLock == nil {
				log.Warn("pre InviterLock nil:", req.BlockNumber, req.TxHash, preAcc.Account)
				tmp := molecule.ScriptDefault()
				inviterLock = &tmp
			}
			inviterHex, _, err := b.dasCore.Daf().ScriptToHex(molecule.MoleculeScript2CkbScript(inviterLock))
			if err != nil {
				resp.Err = fmt.Errorf("ScriptToHex err: %s", err.Error())
				return
			}
			inviteeId := common.Bytes2Hex(common.GetAccountIdByAccount(preAcc.Account))
			rebateInfos = append(rebateInfos, dao.TableRebateInfo{
				BlockNumber:      req.BlockNumber,
				Outpoint:         common.OutPoint2String(req.TxHash, uint(v.Index)),
				InviteeId:        inviteeId,
				InviteeAccount:   preAcc.Account,
				InviteeChainType: inviteeHex.ChainType,
				InviteeAddress:   inviteeHex.AddressHex,
				RewardType:       dao.RewardTypeInviter,
				Reward:           capacity.Div(decimal.NewFromInt(common.PercentRateBase)).Mul(decimal.NewFromInt(int64(profitRateInviter))).BigInt().Uint64(),
				Action:           common.DasActionConfirmProposal,
				ServiceType:      dao.ServiceTypeRegister,
				InviterArgs:      common.Bytes2Hex(inviterLock.Args().RawData()),
				InviterId:        inviterId,
				InviterChainType: inviterHex.ChainType,
				InviterAddress:   inviterHex.AddressHex,
				BlockTimestamp:   req.BlockTimestamp,
			})

			channelLock, _ := preAcc.ChannelLock()
			if channelLock == nil {
				log.Warn("pre ChannelLock nil:", req.BlockNumber, req.TxHash, preAcc.Account)
				tmp := molecule.ScriptDefault()
				channelLock = &tmp
			}
			channelHex, _, err := b.dasCore.Daf().ScriptToHex(molecule.MoleculeScript2CkbScript(channelLock))
			if err != nil {
				resp.Err = fmt.Errorf("ScriptToHex err: %s", err.Error())
				return
			}
			rebateInfos = append(rebateInfos, dao.TableRebateInfo{
				BlockNumber:      req.BlockNumber,
				Outpoint:         common.OutPoint2String(req.TxHash, uint(v.Index)),
				InviteeId:        inviteeId,
				InviteeAccount:   preAcc.Account,
				InviteeChainType: inviteeHex.ChainType,
				InviteeAddress:   inviteeHex.AddressHex,
				RewardType:       dao.RewardTypeChannel,
				Reward:           capacity.Div(decimal.NewFromInt(common.PercentRateBase)).Mul(decimal.NewFromInt(int64(profitRateChannel))).BigInt().Uint64(),
				Action:           common.DasActionConfirmProposal,
				ServiceType:      dao.ServiceTypeRegister,
				InviterArgs:      common.Bytes2Hex(channelLock.Args().RawData()),
				InviterChainType: channelHex.ChainType,
				InviterAddress:   channelHex.AddressHex,
				BlockTimestamp:   req.BlockTimestamp,
			})
		}
	}

	if err = b.dbDao.ConfirmProposal(incomeCellInfos, accountInfos, transactionInfos, rebateInfos); err != nil {
		log.Error("ConfirmProposal err:", err.Error(), req.TxHash, req.BlockNumber)
		resp.Err = fmt.Errorf("ConfirmProposal err: %s ", err.Error())
		return
	}

	return
}

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
	recordList := accBuilder.RecordList()
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
		Account:     account,
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
	}

	log.Info("ActionEditManager:", account, managerHex.DasAlgorithmId, managerHex.ChainType, managerHex.AddressHex, transactionInfo.Address)

	if err := b.dbDao.EditManager(accountInfo, transactionInfo); err != nil {
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

	if err := b.dbDao.RenewAccount(inputsOutpoints, incomeCellInfos, accountInfo, transactionInfo); err != nil {
		log.Error("RenewAccount err:", err.Error(), toolib.JsonString(transactionInfo))
		resp.Err = fmt.Errorf("RenewAccount err: %s", err.Error())
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
	res, err := b.ckbClient.GetTxByHashOnChain(req.Tx.Inputs[oldBuilder.Index].PreviousOutput.TxHash)
	if err != nil {
		resp.Err = fmt.Errorf("GetTxByHashOnChain err: %s", err.Error())
		return
	}

	oldHex, _, err := b.dasCore.Daf().ArgsToHex(res.Transaction.Outputs[oldBuilder.Index].Lock.Args)
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
	accountInfo := dao.TableAccountInfo{
		BlockNumber:        req.BlockNumber,
		Outpoint:           common.OutPoint2String(req.TxHash, uint(builder.Index)),
		AccountId:          accountId,
		Account:            account,
		OwnerChainType:     oHex.ChainType,
		Owner:              oHex.AddressHex,
		OwnerAlgorithmId:   oHex.DasAlgorithmId,
		ManagerChainType:   mHex.ChainType,
		Manager:            mHex.AddressHex,
		ManagerAlgorithmId: mHex.DasAlgorithmId,
	}
	var recordsInfos []dao.TableRecordsInfo
	recordList := builder.RecordList()
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

	if err := b.dbDao.TransferAccount(accountInfo, transactionInfo, recordsInfos); err != nil {
		log.Error("TransferAccount err:", err.Error(), toolib.JsonString(transactionInfo))
		resp.Err = fmt.Errorf("TransferAccount err: %s", err.Error())
	}

	return
}

func (b *BlockParser) ActionRecycleExpiredAccount(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	log.Info("ActionRecycleExpiredAccount:", req.BlockNumber, req.TxHash)
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

	accBuilder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	status := dao.AccountStatusOnLock
	if req.Action == common.DasActionUnlockAccountForCrossChain {
		status = dao.AccountStatusNormal
	}

	ownerHex, managerHex, err := b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[0].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}
	accountInfo := dao.TableAccountInfo{
		BlockNumber:        req.BlockNumber,
		Outpoint:           common.OutPoint2String(req.TxHash, 0),
		AccountId:          accBuilder.AccountId,
		OwnerChainType:     ownerHex.ChainType,
		Owner:              ownerHex.AddressHex,
		OwnerAlgorithmId:   ownerHex.DasAlgorithmId,
		ManagerChainType:   managerHex.ChainType,
		Manager:            managerHex.AddressHex,
		ManagerAlgorithmId: managerHex.DasAlgorithmId,
		Status:             status,
	}
	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      accBuilder.AccountId,
		Account:        accBuilder.Account,
		Action:         req.Action,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      ownerHex.ChainType,
		Address:        ownerHex.AddressHex,
		Capacity:       0,
		Outpoint:       common.OutPoint2String(req.TxHash, 0),
		BlockTimestamp: req.BlockTimestamp,
	}

	if err = b.dbDao.AccountCrossChain(accountInfo, transactionInfo); err != nil {
		log.Error("AccountCrossChain err:", err.Error(), req.TxHash, req.BlockNumber)
		resp.Err = fmt.Errorf("AccountCrossChain err: %s ", err.Error())
		return
	}

	return
}
