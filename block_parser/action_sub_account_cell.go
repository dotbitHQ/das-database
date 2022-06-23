package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/DeAccountSystems/das-lib/witness"
	"strconv"
	"strings"
)

func (b *BlockParser) ActionEnableSubAccount(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DASContractNameSubAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version enable sub account tx")
		return
	}

	log.Info("ActionEnableSubAccount:", req.BlockNumber, req.TxHash)

	builder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	ownerHex, _, err := b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[builder.Index].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}

	accountInfo := dao.TableAccountInfo{
		BlockNumber:          req.BlockNumber,
		Outpoint:             common.OutPoint2String(req.TxHash, 0),
		AccountId:            builder.AccountId,
		EnableSubAccount:     dao.AccountEnableStatusOn,
		RenewSubAccountPrice: builder.RenewSubAccountPrice,
	}
	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      builder.AccountId,
		Account:        builder.Account,
		Action:         common.DasActionEnableSubAccount,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      ownerHex.ChainType,
		Address:        ownerHex.AddressHex,
		Capacity:       req.Tx.Outputs[1].Capacity,
		Outpoint:       common.OutPoint2String(req.TxHash, 1),
		BlockTimestamp: req.BlockTimestamp,
	}

	if err = b.dbDao.EnableSubAccount(accountInfo, transactionInfo); err != nil {
		resp.Err = fmt.Errorf("EnableSubAccount err: %s", err.Error())
		return
	}

	return
}

func (b *BlockParser) ActionCreateSubAccount(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DASContractNameSubAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version create sub account tx")
		return
	}
	log.Info("ActionCreateSubAccount:", req.BlockNumber, req.TxHash)

	// check sub-account config custom-script-args or not
	contractSub, err := core.GetDasContractInfo(common.DASContractNameSubAccountCellType)
	if err != nil {
		resp.Err = fmt.Errorf("GetDasContractInfo err: %s", err.Error())
		return
	}
	contractAcc, err := core.GetDasContractInfo(common.DasContractNameAccountCellType)
	if err != nil {
		resp.Err = fmt.Errorf("GetDasContractInfo err: %s", err.Error())
		return
	}
	var subAccountCellOutpoint, parentAccountId, accountCellOutpoint string
	for i, v := range req.Tx.Outputs {
		if v.Type != nil && contractSub.IsSameTypeId(v.Type.CodeHash) {
			parentAccountId = common.Bytes2Hex(v.Type.Args)
			subAccountCellOutpoint = common.OutPoint2String(req.TxHash, uint(i))
		}
		if v.Type != nil && contractAcc.IsSameTypeId(v.Type.CodeHash) {
			accountCellOutpoint = common.OutPoint2String(req.TxHash, uint(i))
		}
	}
	var parentAccountInfo dao.TableAccountInfo
	if accountCellOutpoint != "" {
		parentAccountInfo = dao.TableAccountInfo{
			BlockNumber: req.BlockNumber,
			Outpoint:    accountCellOutpoint,
			AccountId:   parentAccountId,
		}
	}

	builderMap, err := witness.SubAccountBuilderMapFromTx(req.Tx)
	if err != nil {
		resp.Err = fmt.Errorf("SubAccountBuilderMapFromTx err: %s", err.Error())
		return
	}

	builderConfig, err := b.dasCore.ConfigCellDataBuilderByTypeArgs(common.ConfigCellTypeArgsSubAccount)
	if err != nil {
		resp.Err = fmt.Errorf("ConfigCellDataBuilderByTypeArgs err: %s", err.Error())
		return
	}
	newPrice, err := builderConfig.NewSubAccountPrice()
	if err != nil {
		resp.Err = fmt.Errorf("NewSubAccountPrice err: %s", err.Error())
		return
	}

	var accountInfos []dao.TableAccountInfo
	var subAccountIds []string
	var smtInfos []dao.TableSmtInfo
	var capacity uint64
	var parentAccount string
	for _, v := range builderMap {
		ownerHex, managerHex, err := b.dasCore.Daf().ArgsToHex(v.SubAccount.Lock.Args)
		if err != nil {
			resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
			return
		}

		accountInfos = append(accountInfos, dao.TableAccountInfo{
			BlockNumber:          req.BlockNumber,
			Outpoint:             common.OutPoint2String(req.TxHash, 0),
			AccountId:            v.SubAccount.AccountId,
			ParentAccountId:      parentAccountId,
			Account:              v.Account,
			OwnerChainType:       ownerHex.ChainType,
			Owner:                ownerHex.AddressHex,
			OwnerAlgorithmId:     ownerHex.DasAlgorithmId,
			ManagerChainType:     managerHex.ChainType,
			Manager:              managerHex.AddressHex,
			ManagerAlgorithmId:   managerHex.DasAlgorithmId,
			Status:               dao.AccountStatus(v.SubAccount.Status),
			EnableSubAccount:     dao.EnableSubAccount(v.SubAccount.EnableSubAccount),
			RenewSubAccountPrice: v.SubAccount.RenewSubAccountPrice,
			Nonce:                v.SubAccount.Nonce,
			RegisteredAt:         v.SubAccount.RegisteredAt,
			ExpiredAt:            v.SubAccount.ExpiredAt,
			ConfirmProposalHash:  req.TxHash,
		})
		parentAccount = v.Account[strings.Index(v.Account, ".")+1:]
		subAccountIds = append(subAccountIds, v.SubAccount.AccountId)
		smtInfos = append(smtInfos, dao.TableSmtInfo{
			BlockNumber:     req.BlockNumber,
			Outpoint:        common.OutPoint2String(req.TxHash, 1),
			AccountId:       v.SubAccount.AccountId,
			ParentAccountId: parentAccountId,
			LeafDataHash:    common.Bytes2Hex(v.SubAccount.ToH256()),
		})
		capacity += (v.SubAccount.ExpiredAt - v.SubAccount.RegisteredAt) / uint64(common.OneYearSec) * newPrice
	}

	dasLock, err := core.GetDasContractInfo(common.DasContractNameDispatchCellType)
	if err != nil {
		resp.Err = fmt.Errorf("GetDasContractInfo err: %s", err.Error())
		return
	}

	res, err := b.dasCore.Client().GetTransaction(b.ctx, req.Tx.Inputs[len(req.Tx.Inputs)-1].PreviousOutput.TxHash)
	if err != nil {
		resp.Err = fmt.Errorf("GetTransaction err: %s", err.Error())
		return
	}

	output := res.Transaction.Outputs[req.Tx.Inputs[len(req.Tx.Inputs)-1].PreviousOutput.Index]
	oCT, oA := common.ChainTypeCkb, common.Bytes2Hex(output.Lock.Args)
	if dasLock.IsSameTypeId(output.Lock.CodeHash) {
		tmpHex, _, err := b.dasCore.Daf().ArgsToHex(output.Lock.Args)
		if err != nil {
			resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
			return
		}
		oCT, oA = tmpHex.ChainType, tmpHex.AddressHex
	}
	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      parentAccountId,
		Account:        parentAccount,
		Action:         common.DasActionCreateSubAccount,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      oCT,
		Address:        oA,
		Capacity:       capacity,
		Outpoint:       subAccountCellOutpoint,
		BlockTimestamp: req.BlockTimestamp,
	}

	if err = b.dbDao.CreateSubAccount(subAccountIds, accountInfos, smtInfos, transactionInfo, parentAccountInfo); err != nil {
		resp.Err = fmt.Errorf("CreateSubAccount err: %s", err.Error())
		return
	}

	return
}

func (b *BlockParser) ActionEditSubAccount(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DASContractNameSubAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version edit sub account tx")
		return
	}
	log.Info("ActionEditSubAccount:", req.BlockNumber, req.TxHash)

	builderMap, err := witness.SubAccountBuilderMapFromTx(req.Tx)
	if err != nil {
		resp.Err = fmt.Errorf("SubAccountBuilderMapFromTx err: %s", err.Error())
		return
	}

	var index uint
	for _, builder := range builderMap {
		ownerHex, _, err := b.dasCore.Daf().ArgsToHex(builder.SubAccount.Lock.Args)
		if err != nil {
			resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
			return
		}
		outpoint := common.OutPoint2String(req.TxHash, 0)

		accountInfo := dao.TableAccountInfo{
			BlockNumber: req.BlockNumber,
			Outpoint:    outpoint,
			AccountId:   builder.SubAccount.AccountId,
			Nonce:       builder.CurrentSubAccount.Nonce,
		}
		smtInfo := dao.TableSmtInfo{
			BlockNumber:  req.BlockNumber,
			Outpoint:     outpoint,
			AccountId:    builder.SubAccount.AccountId,
			LeafDataHash: common.Bytes2Hex(builder.CurrentSubAccount.ToH256()),
		}
		transactionInfo := dao.TableTransactionInfo{
			BlockNumber:    req.BlockNumber,
			AccountId:      builder.SubAccount.AccountId,
			Account:        builder.Account,
			Action:         common.DasActionEditSubAccount,
			ServiceType:    dao.ServiceTypeRegister,
			ChainType:      ownerHex.ChainType,
			Address:        ownerHex.AddressHex,
			Capacity:       0,
			Outpoint:       common.OutPoint2String(outpoint, index),
			BlockTimestamp: req.BlockTimestamp,
		}
		index++

		subAccount, err := builder.ConvertToEditValue()
		if err != nil {
			resp.Err = fmt.Errorf("ConvertToEditValue err: %s", err.Error())
			return
		}
		switch string(builder.EditKey) {
		case common.EditKeyOwner:
			oHex, mHex, err := b.dasCore.Daf().ArgsToHex(common.Hex2Bytes(subAccount.LockArgs))
			if err != nil {
				resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
				return
			}
			accountInfo.OwnerAlgorithmId = oHex.DasAlgorithmId
			accountInfo.OwnerChainType = oHex.ChainType
			accountInfo.Owner = oHex.AddressHex
			accountInfo.ManagerAlgorithmId = mHex.DasAlgorithmId
			accountInfo.ManagerChainType = mHex.ChainType
			accountInfo.Manager = mHex.AddressHex
			transactionInfo.ChainType = oHex.ChainType
			transactionInfo.Address = oHex.AddressHex
			if err = b.dbDao.EditOwnerSubAccount(accountInfo, smtInfo, transactionInfo); err != nil {
				resp.Err = fmt.Errorf("EditOwnerSubAccount err: %s", err.Error())
			}
		case common.EditKeyManager:
			_, mHex, err := b.dasCore.Daf().ArgsToHex(common.Hex2Bytes(subAccount.LockArgs))
			if err != nil {
				resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
				return
			}
			accountInfo.ManagerAlgorithmId = mHex.DasAlgorithmId
			accountInfo.ManagerChainType = mHex.ChainType
			accountInfo.Manager = mHex.AddressHex
			if err = b.dbDao.EditManagerSubAccount(accountInfo, smtInfo, transactionInfo); err != nil {
				resp.Err = fmt.Errorf("EditManagerSubAccount err: %s", err.Error())
			}
		case common.EditKeyRecords:
			var recordsInfos []dao.TableRecordsInfo
			for _, v := range subAccount.Records {
				recordsInfos = append(recordsInfos, dao.TableRecordsInfo{
					AccountId:       builder.SubAccount.AccountId,
					ParentAccountId: common.Bytes2Hex(req.Tx.Outputs[0].Type.Args),
					Account:         builder.Account,
					Key:             v.Key,
					Type:            v.Type,
					Label:           v.Label,
					Value:           v.Value,
					Ttl:             strconv.FormatUint(uint64(v.TTL), 10),
				})
			}
			if err = b.dbDao.EditRecordsSubAccount(accountInfo, smtInfo, transactionInfo, recordsInfos); err != nil {
				resp.Err = fmt.Errorf("EditRecordsSubAccount err: %s", err.Error())
				return
			}
		}
	}

	return
}

func (b *BlockParser) ActionRenewSubAccount(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	log.Info("ActionRenewSubAccount:", req.BlockNumber, req.TxHash)
	return
}

/*func (b *BlockParser) ActionRenewSubAccount(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DASContractNameSubAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version renew sub account tx")
		return
	}
	log.Info("ActionRenewSubAccount:", req.BlockNumber, req.TxHash)

	builderMap, err := witness.SubAccountBuilderMapFromTx(req.Tx)
	if err != nil {
		resp.Err = fmt.Errorf("SubAccountBuilderMapFromTx err: %s", err.Error())
		return
	}

	builderConfig, err := b.dasCore.ConfigCellDataBuilderByTypeArgs(common.ConfigCellTypeArgsSubAccount)
	if err != nil {
		resp.Err = fmt.Errorf("ConfigCellDataBuilderByTypeArgs err: %s", err.Error())
		return
	}
	renewPrice, err := builderConfig.RenewSubAccountPrice()
	if err != nil {
		resp.Err = fmt.Errorf("RenewSubAccountPrice err: %s", err.Error())
		return
	}

	var accountInfos []dao.TableAccountInfo
	var smtInfos []dao.TableSmtInfo
	var transactionInfos []dao.TableTransactionInfo
	var index uint
	for _, builder := range builderMap {
		oHex, _, err := b.dasCore.Daf().ArgsToHex(builder.SubAccount.Lock.Args)
		if err != nil {
			resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
			return
		}
		outpoint := common.OutPoint2String(req.TxHash, 0)

		accountInfo := dao.TableAccountInfo{
			BlockNumber: req.BlockNumber,
			Outpoint:    outpoint,
			AccountId:   builder.SubAccount.AccountId,
			Nonce:       builder.CurrentSubAccount.Nonce,
		}
		smtInfo := dao.TableSmtInfo{
			BlockNumber:  req.BlockNumber,
			Outpoint:     outpoint,
			AccountId:    builder.SubAccount.AccountId,
			LeafDataHash: common.Bytes2Hex(builder.CurrentSubAccount.ToH256()),
		}
		transactionInfo := dao.TableTransactionInfo{
			BlockNumber:    req.BlockNumber,
			AccountId:      builder.SubAccount.AccountId,
			Account:        builder.Account,
			Action:         common.DasActionRenewSubAccount,
			ServiceType:    dao.ServiceTypeRegister,
			ChainType:      oHex.ChainType,
			Address:        oHex.AddressHex,
			Outpoint:       common.OutPoint2String(outpoint, index),
			BlockTimestamp: req.BlockTimestamp,
		}
		index++

		subAccount, err := builder.ConvertToEditValue()
		if err != nil {
			resp.Err = fmt.Errorf("ConvertToEditValue err: %s", err.Error())
			return
		}
		switch string(builder.EditKey) {
		case common.EditKeyExpiredAt:
			accountInfo.ExpiredAt = subAccount.ExpiredAt
			accountInfos = append(accountInfos, accountInfo)
			smtInfos = append(smtInfos, smtInfo)
			transactionInfo.Capacity = (subAccount.ExpiredAt - builder.SubAccount.ExpiredAt) / uint64(common.OneYearSec) * renewPrice
			transactionInfos = append(transactionInfos, transactionInfo)
		}
	}

	if err = b.dbDao.RenewSubAccount(accountInfos, smtInfos, transactionInfos); err != nil {
		resp.Err = fmt.Errorf("RenewSubAccount err: %s", err.Error())
		return
	}

	return
}*/

func (b *BlockParser) ActionRecycleSubAccount(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	log.Info("ActionRecycleSubAccount:", req.BlockNumber, req.TxHash)
	return
}

/*func (b *BlockParser) ActionRecycleSubAccount(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DASContractNameSubAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version recycle sub account tx")
		return
	}
	log.Info("ActionRecycleSubAccount:", req.BlockNumber, req.TxHash)

	builderMap, err := witness.SubAccountBuilderMapFromTx(req.Tx)
	if err != nil {
		resp.Err = fmt.Errorf("SubAccountBuilderMapFromTx err: %s", err.Error())
		return
	}

	var accountIds []string
	var transactionInfos []dao.TableTransactionInfo
	var index uint
	for _, builder := range builderMap {
		// if expired time greater than three months ago, then reject the recycle of sub_account.
		if builder.SubAccount.ExpiredAt > uint64(time.Now().Add(-time.Hour*24*90).Unix()) {
			resp.Err = fmt.Errorf("not yet arrived expired time: %d", builder.SubAccount.ExpiredAt)
			return
		}
		oHex, _, err := b.dasCore.Daf().ArgsToHex(builder.SubAccount.Lock.Args)
		if err != nil {
			resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
			return
		}
		outpoint := common.OutPoint2String(req.TxHash, 0)

		accountIds = append(accountIds, builder.SubAccount.AccountId)
		transactionInfos = append(transactionInfos, dao.TableTransactionInfo{
			BlockNumber:    req.BlockNumber,
			AccountId:      builder.SubAccount.AccountId,
			Account:        builder.Account,
			Action:         common.DasActionRecycleSubAccount,
			ServiceType:    dao.ServiceTypeRegister,
			ChainType:      oHex.ChainType,
			Address:        oHex.AddressHex,
			Capacity:       req.Tx.Outputs[0].Capacity,
			Outpoint:       common.OutPoint2String(outpoint, index),
			BlockTimestamp: req.BlockTimestamp,
		})
		index++
	}

	if err = b.dbDao.RecycleSubAccount(accountIds, transactionInfos); err != nil {
		resp.Err = fmt.Errorf("RecycleSubAccount err: %s", err.Error())
		return
	}

	return
}*/

func (b *BlockParser) ActionSubAccountCrossChain(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	log.Info("ActionSubAccountCrossChain:", req.BlockNumber, req.TxHash, req.Action)
	return
}

func (b *BlockParser) ActionConfigSubAccountCreatingScript(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DASContractNameSubAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		return
	}
	log.Info("ActionConfigSubAccountCreatingScript:", req.BlockNumber, req.TxHash)

	// update account cell outpoint
	builder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("witness.AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	outpoint := common.OutPoint2String(req.TxHash, uint(builder.Index))
	if err := b.dbDao.UpdateAccountOutpoint(builder.AccountId, outpoint); err != nil {
		resp.Err = fmt.Errorf("UpdateAccountOutpoint err: %s", err.Error())
	}

	return
}
