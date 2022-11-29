package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/witness"
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
		EnableSubAccount:     builder.EnableSubAccount,
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

func (b *BlockParser) ActionUpdateSubAccount(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DASContractNameSubAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version edit sub account tx")
		return
	}
	log.Info("ActionUpdateSubAccount:", req.BlockNumber, req.TxHash)

	var subAccountNewBuilder witness.SubAccountNewBuilder
	builderMap, err := subAccountNewBuilder.SubAccountNewMapFromTx(req.Tx)
	if err != nil {
		resp.Err = fmt.Errorf("SubAccountBuilderMapFromTx err: %s", err.Error())
		return
	}

	var createBuilderMap = make(map[string]*witness.SubAccountNew)
	var editBuilderMap = make(map[string]*witness.SubAccountNew)
	for k, v := range builderMap {
		switch v.Action {
		case common.SubActionCreate:
			createBuilderMap[k] = v
		case common.SubActionEdit:
			editBuilderMap[k] = v
		default:
			resp.Err = fmt.Errorf("unknow sub-action [%s]", v.Action)
			return
		}
	}
	if err := b.actionUpdateSubAccountForCreate(req, createBuilderMap); err != nil {
		resp.Err = fmt.Errorf("create err: %s", err.Error())
		return
	}

	if err := b.actionUpdateSubAccountForEdit(req, createBuilderMap); err != nil {
		resp.Err = fmt.Errorf("edit err: %s", err.Error())
		return
	}

	return
}

func (b *BlockParser) actionUpdateSubAccountForCreate(req FuncTransactionHandleReq, createBuilderMap map[string]*witness.SubAccountNew) error {
	if len(createBuilderMap) > 0 {
		return nil
	}
	// check sub-account config custom-script-args or not
	contractSub, err := core.GetDasContractInfo(common.DASContractNameSubAccountCellType)
	if err != nil {
		return fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}

	var subAccountCellOutpoint, parentAccountId string
	for i, v := range req.Tx.Outputs {
		if v.Type != nil && contractSub.IsSameTypeId(v.Type.CodeHash) {
			parentAccountId = common.Bytes2Hex(v.Type.Args)
			subAccountCellOutpoint = common.OutPoint2String(req.TxHash, uint(i))
		}
	}

	builderConfig, err := b.dasCore.ConfigCellDataBuilderByTypeArgs(common.ConfigCellTypeArgsSubAccount)
	if err != nil {
		return fmt.Errorf("ConfigCellDataBuilderByTypeArgs err: %s", err.Error())
	}
	newPrice, err := builderConfig.NewSubAccountPrice()
	if err != nil {
		return fmt.Errorf("NewSubAccountPrice err: %s", err.Error())
	}

	var accountInfos []dao.TableAccountInfo
	var subAccountIds []string
	var smtInfos []dao.TableSmtInfo
	var capacity uint64
	var parentAccount string

	for _, v := range createBuilderMap {
		ownerHex, managerHex, err := b.dasCore.Daf().ArgsToHex(v.SubAccountData.Lock.Args)
		if err != nil {
			return fmt.Errorf("ArgsToHex err: %s", err.Error())
		}
		accountInfos = append(accountInfos, dao.TableAccountInfo{
			BlockNumber:          req.BlockNumber,
			Outpoint:             common.OutPoint2String(req.TxHash, 0),
			AccountId:            v.SubAccountData.AccountId,
			ParentAccountId:      parentAccountId,
			Account:              v.Account,
			OwnerChainType:       ownerHex.ChainType,
			Owner:                ownerHex.AddressHex,
			OwnerAlgorithmId:     ownerHex.DasAlgorithmId,
			ManagerChainType:     managerHex.ChainType,
			Manager:              managerHex.AddressHex,
			ManagerAlgorithmId:   managerHex.DasAlgorithmId,
			Status:               v.SubAccountData.Status,
			EnableSubAccount:     v.SubAccountData.EnableSubAccount,
			RenewSubAccountPrice: v.SubAccountData.RenewSubAccountPrice,
			Nonce:                v.SubAccountData.Nonce,
			RegisteredAt:         v.SubAccountData.RegisteredAt,
			ExpiredAt:            v.SubAccountData.ExpiredAt,
			ConfirmProposalHash:  req.TxHash,
		})
		parentAccount = v.Account[strings.Index(v.Account, ".")+1:]
		subAccountIds = append(subAccountIds, v.SubAccountData.AccountId)
		smtInfos = append(smtInfos, dao.TableSmtInfo{
			BlockNumber:     req.BlockNumber,
			Outpoint:        subAccountCellOutpoint,
			AccountId:       v.SubAccountData.AccountId,
			ParentAccountId: parentAccountId,
			LeafDataHash:    common.Bytes2Hex(v.SubAccountData.ToH256()),
		})
		capacity += (v.SubAccountData.ExpiredAt - v.SubAccountData.RegisteredAt) / uint64(common.OneYearSec) * newPrice
	}

	ownerHex, _, err := b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[len(req.Tx.Outputs)-1].Lock.Args)
	if err != nil {
		return fmt.Errorf("ArgsToHex err: %s", err.Error())
	}

	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      parentAccountId,
		Account:        parentAccount,
		Action:         common.DasActionCreateSubAccount,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      ownerHex.ChainType,
		Address:        ownerHex.AddressHex,
		Capacity:       capacity,
		Outpoint:       subAccountCellOutpoint,
		BlockTimestamp: req.BlockTimestamp,
	}

	if err = b.dbDao.UpdateSubAccountForCreate(subAccountIds, accountInfos, smtInfos, transactionInfo); err != nil {
		return fmt.Errorf("UpdateSubAccountForCreate err: %s", err.Error())
	}

	return nil
}

func (b *BlockParser) actionUpdateSubAccountForEdit(req FuncTransactionHandleReq, editBuilderMap map[string]*witness.SubAccountNew) error {
	if len(editBuilderMap) > 0 {
		return nil
	}

	var index uint
	for _, builder := range editBuilderMap {
		ownerHex, _, err := b.dasCore.Daf().ArgsToHex(builder.SubAccountData.Lock.Args)
		if err != nil {
			return fmt.Errorf("ArgsToHex err: %s", err.Error())
		}
		outpoint := common.OutPoint2String(req.TxHash, 0)
		accountInfo := dao.TableAccountInfo{
			BlockNumber: req.BlockNumber,
			Outpoint:    outpoint,
			AccountId:   builder.SubAccountData.AccountId,
			Nonce:       builder.CurrentSubAccountData.Nonce,
		}

		smtInfo := dao.TableSmtInfo{
			BlockNumber:  req.BlockNumber,
			Outpoint:     outpoint,
			AccountId:    builder.SubAccountData.AccountId,
			LeafDataHash: common.Bytes2Hex(builder.CurrentSubAccountData.ToH256()),
		}
		transactionInfo := dao.TableTransactionInfo{
			BlockNumber:    req.BlockNumber,
			AccountId:      builder.SubAccountData.AccountId,
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

		switch builder.EditKey {
		case common.EditKeyOwner:
			oHex, mHex, err := b.dasCore.Daf().ArgsToHex(builder.EditLockArgs)
			if err != nil {
				return fmt.Errorf("ArgsToHex err: %s", err.Error())
			}
			accountInfo.OwnerAlgorithmId = oHex.DasAlgorithmId
			accountInfo.OwnerChainType = oHex.ChainType
			accountInfo.Owner = oHex.AddressHex
			accountInfo.ManagerAlgorithmId = mHex.DasAlgorithmId
			accountInfo.ManagerChainType = mHex.ChainType
			accountInfo.Manager = mHex.AddressHex
			if err = b.dbDao.EditOwnerSubAccount(accountInfo, smtInfo, transactionInfo); err != nil {
				return fmt.Errorf("EditOwnerSubAccount err: %s", err.Error())
			}
		case common.EditKeyManager:
			_, mHex, err := b.dasCore.Daf().ArgsToHex(builder.EditLockArgs)
			if err != nil {
				return fmt.Errorf("ArgsToHex err: %s", err.Error())
			}
			accountInfo.ManagerAlgorithmId = mHex.DasAlgorithmId
			accountInfo.ManagerChainType = mHex.ChainType
			accountInfo.Manager = mHex.AddressHex
			if err = b.dbDao.EditManagerSubAccount(accountInfo, smtInfo, transactionInfo); err != nil {
				return fmt.Errorf("EditManagerSubAccount err: %s", err.Error())
			}
		case common.EditKeyRecords:
			var recordsInfos []dao.TableRecordsInfo
			for _, v := range builder.EditRecords {
				recordsInfos = append(recordsInfos, dao.TableRecordsInfo{
					AccountId:       builder.SubAccountData.AccountId,
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
				return fmt.Errorf("EditRecordsSubAccount err: %s", err.Error())
			}
		}
	}

	return nil
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

	var subAccountNewBuilder witness.SubAccountNewBuilder
	builderMap, err := subAccountNewBuilder.SubAccountNewMapFromTx(req.Tx)
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
		ownerHex, managerHex, err := b.dasCore.Daf().ArgsToHex(v.SubAccountData.Lock.Args)
		if err != nil {
			resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
			return
		}

		accountInfos = append(accountInfos, dao.TableAccountInfo{
			BlockNumber:          req.BlockNumber,
			Outpoint:             common.OutPoint2String(req.TxHash, 0),
			AccountId:            v.SubAccountData.AccountId,
			ParentAccountId:      parentAccountId,
			Account:              v.Account,
			OwnerChainType:       ownerHex.ChainType,
			Owner:                ownerHex.AddressHex,
			OwnerAlgorithmId:     ownerHex.DasAlgorithmId,
			ManagerChainType:     managerHex.ChainType,
			Manager:              managerHex.AddressHex,
			ManagerAlgorithmId:   managerHex.DasAlgorithmId,
			Status:               v.SubAccountData.Status,
			EnableSubAccount:     v.SubAccountData.EnableSubAccount,
			RenewSubAccountPrice: v.SubAccountData.RenewSubAccountPrice,
			Nonce:                v.SubAccountData.Nonce,
			RegisteredAt:         v.SubAccountData.RegisteredAt,
			ExpiredAt:            v.SubAccountData.ExpiredAt,
			ConfirmProposalHash:  req.TxHash,
		})
		parentAccount = v.Account[strings.Index(v.Account, ".")+1:]
		subAccountIds = append(subAccountIds, v.SubAccountData.AccountId)
		smtInfos = append(smtInfos, dao.TableSmtInfo{
			BlockNumber:     req.BlockNumber,
			Outpoint:        common.OutPoint2String(req.TxHash, 1),
			AccountId:       v.SubAccountData.AccountId,
			ParentAccountId: parentAccountId,
			LeafDataHash:    common.Bytes2Hex(v.SubAccountData.ToH256()),
		})
		capacity += (v.SubAccountData.ExpiredAt - v.SubAccountData.RegisteredAt) / uint64(common.OneYearSec) * newPrice
	}

	ownerHex, _, err := b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[len(req.Tx.Outputs)-1].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}

	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      parentAccountId,
		Account:        parentAccount,
		Action:         common.DasActionCreateSubAccount,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      ownerHex.ChainType,
		Address:        ownerHex.AddressHex,
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

	var subAccountNewBuilder witness.SubAccountNewBuilder
	builderMap, err := subAccountNewBuilder.SubAccountNewMapFromTx(req.Tx)
	if err != nil {
		resp.Err = fmt.Errorf("SubAccountBuilderMapFromTx err: %s", err.Error())
		return
	}

	var index uint
	for _, builder := range builderMap {
		ownerHex, _, err := b.dasCore.Daf().ArgsToHex(builder.SubAccountData.Lock.Args)
		if err != nil {
			resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
			return
		}
		outpoint := common.OutPoint2String(req.TxHash, 0)

		accountInfo := dao.TableAccountInfo{
			BlockNumber: req.BlockNumber,
			Outpoint:    outpoint,
			AccountId:   builder.SubAccountData.AccountId,
			Nonce:       builder.CurrentSubAccountData.Nonce,
		}
		smtInfo := dao.TableSmtInfo{
			BlockNumber:  req.BlockNumber,
			Outpoint:     outpoint,
			AccountId:    builder.SubAccountData.AccountId,
			LeafDataHash: common.Bytes2Hex(builder.CurrentSubAccountData.ToH256()),
		}
		transactionInfo := dao.TableTransactionInfo{
			BlockNumber:    req.BlockNumber,
			AccountId:      builder.SubAccountData.AccountId,
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

		switch builder.EditKey {
		case common.EditKeyOwner:
			oHex, mHex, err := b.dasCore.Daf().ArgsToHex(builder.EditLockArgs)
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
			if err = b.dbDao.EditOwnerSubAccount(accountInfo, smtInfo, transactionInfo); err != nil {
				resp.Err = fmt.Errorf("EditOwnerSubAccount err: %s", err.Error())
			}
		case common.EditKeyManager:
			_, mHex, err := b.dasCore.Daf().ArgsToHex(builder.EditLockArgs)
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
			for _, v := range builder.EditRecords {
				recordsInfos = append(recordsInfos, dao.TableRecordsInfo{
					AccountId:       builder.SubAccountData.AccountId,
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
	accountCellOutpoint := common.OutPoint2String(req.TxHash, uint(builder.Index))
	ownerHex, _, err := b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[builder.Index].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}

	cs := dao.TableCustomScriptInfo{
		BlockNumber:    req.BlockNumber,
		Outpoint:       common.OutPoint2String(req.TxHash, 1),
		BlockTimestamp: req.BlockTimestamp,
		AccountId:      builder.AccountId,
	}

	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      builder.AccountId,
		Account:        builder.Account,
		Action:         common.DasActionConfigSubAccountCustomScript,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      ownerHex.ChainType,
		Address:        ownerHex.AddressHex,
		Capacity:       0,
		Outpoint:       common.OutPoint2String(req.TxHash, 0),
		BlockTimestamp: req.BlockTimestamp,
	}

	if err = b.dbDao.UpdateCustomScript(cs, accountCellOutpoint, transactionInfo); err != nil {
		resp.Err = fmt.Errorf("UpdateAccountOutpoint err: %s", err.Error())
	}

	return
}

func (b *BlockParser) ActionCollectSubAccountProfit(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DASContractNameSubAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		return
	}
	log.Info("ActionCollectSubAccountProfit:", req.BlockNumber, req.TxHash)

	accBuilder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeDep)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}

	var txs []dao.TableTransactionInfo
	if len(req.Tx.Outputs) >= 2 {
		ownerHex, _, err := b.dasCore.Daf().ScriptToHex(req.Tx.Outputs[1].Lock)
		if err != nil {
			resp.Err = fmt.Errorf("ScriptToHex err: %s", err.Error())
			return
		}
		txs = append(txs, dao.TableTransactionInfo{
			BlockNumber:    req.BlockNumber,
			AccountId:      accBuilder.AccountId,
			Account:        accBuilder.Account,
			Action:         req.Action,
			ServiceType:    dao.ServiceTypeRegister,
			ChainType:      ownerHex.ChainType,
			Address:        ownerHex.AddressHex,
			Capacity:       req.Tx.Outputs[1].Capacity,
			Outpoint:       common.OutPoint2String(req.TxHash, 1),
			BlockTimestamp: req.BlockTimestamp,
		})
	}
	if len(req.Tx.Outputs) >= 3 {
		ownerHex, _, err := b.dasCore.Daf().ScriptToHex(req.Tx.Outputs[2].Lock)
		if err != nil {
			resp.Err = fmt.Errorf("ScriptToHex err: %s", err.Error())
			return
		}
		txs = append(txs, dao.TableTransactionInfo{
			BlockNumber:    req.BlockNumber,
			AccountId:      accBuilder.AccountId,
			Account:        accBuilder.Account,
			Action:         req.Action,
			ServiceType:    dao.ServiceTypeRegister,
			ChainType:      ownerHex.ChainType,
			Address:        ownerHex.AddressHex,
			Capacity:       req.Tx.Outputs[2].Capacity,
			Outpoint:       common.OutPoint2String(req.TxHash, 2),
			BlockTimestamp: req.BlockTimestamp,
		})
	}

	if err := b.dbDao.CreateTxs(txs); err != nil {
		resp.Err = fmt.Errorf("CreateTxs err: %s", err.Error())
		return
	}

	return
}
