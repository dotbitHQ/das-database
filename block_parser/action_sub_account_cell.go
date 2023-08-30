package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/molecule"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	feeOwner, _, err := b.dasCore.Daf().ScriptToHex(req.Tx.Outputs[len(req.Tx.Outputs)-1].Lock)
	if err != nil {
		resp.Err = fmt.Errorf("ScriptToHex err: %s", err.Error())
		return
	}
	if feeOwner.AddressHex != ownerHex.AddressHex {
		transactionInfo.Capacity = 0
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
	var renewBuilderMap = make(map[string]*witness.SubAccountNew)
	var editBuilderMap = make(map[string]*witness.SubAccountNew)
	var recycleBuilderMap = make(map[string]*witness.SubAccountNew)
	var approvalBuilderMap = make(map[string]*witness.SubAccountNew)
	for k, v := range builderMap {
		switch v.Action {
		case common.SubActionCreate:
			createBuilderMap[k] = v
		case common.SubActionRenew:
			renewBuilderMap[k] = v
		case common.SubActionEdit:
			editBuilderMap[k] = v
		case common.SubActionRecycle:
			recycleBuilderMap[k] = v
		case common.SubActionCreateApproval, common.SubActionDelayApproval,
			common.SubActionRevokeApproval, common.SubActionFullfillApproval:
			approvalBuilderMap[k] = v
		default:
			resp.Err = fmt.Errorf("unknow sub-action [%s]", v.Action)
			return
		}
	}

	if err := b.actionUpdateSubAccountForRecycle(req, recycleBuilderMap); err != nil {
		resp.Err = fmt.Errorf("recycle sub-account err: %s", err.Error())
		return
	}

	if err := b.actionUpdateSubAccountForCreate(req, createBuilderMap); err != nil {
		resp.Err = fmt.Errorf("create sub-account err: %s", err.Error())
		return
	}

	if err := b.actionUpdateSubAccountForRenew(req, renewBuilderMap); err != nil {
		resp.Err = fmt.Errorf("create err: %s", err.Error())
		return
	}

	if err := b.actionUpdateSubAccountForEdit(req, editBuilderMap); err != nil {
		resp.Err = fmt.Errorf("edit sub-account err: %s", err.Error())
		return
	}

	if err := b.actionUpdateSubAccountForApproval(req, approvalBuilderMap); err != nil {
		resp.Err = fmt.Errorf("approval err: %s", err.Error())
		return
	}
	return
}

func (b *BlockParser) actionUpdateSubAccountForRecycle(req FuncTransactionHandleReq, recycleBuilderMap map[string]*witness.SubAccountNew) error {
	var subAccIds []string
	var smtInfos []dao.TableSmtInfo
	var txs []dao.TableTransactionInfo
	outpoint := common.OutPoint2String(req.TxHash, 0)

	indexTx := uint(0)
	for _, builder := range recycleBuilderMap {
		subAccIds = append(subAccIds, builder.SubAccountData.AccountId)
		value, err := builder.CurrentSubAccountData.ToH256()
		if err != nil {
			return err
		}
		smtInfo := dao.TableSmtInfo{
			BlockNumber:  req.BlockNumber,
			Outpoint:     outpoint,
			AccountId:    builder.SubAccountData.AccountId,
			LeafDataHash: common.Bytes2Hex(value),
		}
		smtInfos = append(smtInfos, smtInfo)
		oHex, _, err := b.dasCore.Daf().ScriptToHex(builder.SubAccountData.Lock)
		if err != nil {
			return fmt.Errorf("ScriptToHex err: %s", err.Error())
		}
		outpointTx := common.OutPoint2String(req.TxHash, indexTx)
		txInfo := dao.TableTransactionInfo{
			BlockNumber:    req.BlockNumber,
			AccountId:      builder.SubAccountData.AccountId,
			Account:        builder.SubAccountData.Account(),
			Action:         common.DasActionRecycleExpiredAccount,
			ServiceType:    dao.ServiceTypeRegister,
			ChainType:      oHex.ChainType,
			Address:        oHex.AddressHex,
			Capacity:       0,
			Outpoint:       outpointTx,
			BlockTimestamp: req.BlockTimestamp,
		}
		txs = append(txs, txInfo)
		indexTx++
	}

	if err := b.dbDao.RecycleSubAccount(subAccIds, smtInfos, txs); err != nil {
		return fmt.Errorf("RecycleSubAccount err: %s", err.Error())
	}

	return nil
}

func (b *BlockParser) actionUpdateSubAccountForCreate(req FuncTransactionHandleReq, createBuilderMap map[string]*witness.SubAccountNew) error {
	if len(createBuilderMap) == 0 {
		return nil
	}
	// check sub_account config custom-script-args or not
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
	var records []dao.TableRecordsInfo
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
			OwnerSubAid:          ownerHex.DasSubAlgorithmId,
			ManagerChainType:     managerHex.ChainType,
			Manager:              managerHex.AddressHex,
			ManagerAlgorithmId:   managerHex.DasAlgorithmId,
			ManagerSubAid:        managerHex.DasSubAlgorithmId,
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
		value, err := v.SubAccountData.ToH256()
		if err != nil {
			return fmt.Errorf("SubAccountData.ToH256() err: %s", err.Error())
		}
		smtInfos = append(smtInfos, dao.TableSmtInfo{
			BlockNumber:     req.BlockNumber,
			Outpoint:        subAccountCellOutpoint,
			AccountId:       v.SubAccountData.AccountId,
			ParentAccountId: parentAccountId,
			LeafDataHash:    common.Bytes2Hex(value),
		})
		capacity += (v.SubAccountData.ExpiredAt - v.SubAccountData.RegisteredAt) / uint64(common.OneYearSec) * newPrice
		for _, record := range v.SubAccountData.Records {
			records = append(records, dao.TableRecordsInfo{
				AccountId:       v.SubAccountData.AccountId,
				ParentAccountId: parentAccountId,
				Account:         v.Account,
				Key:             record.Key,
				Type:            record.Type,
				Label:           record.Label,
				Value:           record.Value,
				Ttl:             strconv.FormatUint(uint64(record.TTL), 10),
			})
		}

	}

	ownerHex, _, err := b.dasCore.Daf().ScriptToHex(req.Tx.Outputs[len(req.Tx.Outputs)-1].Lock)
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

	if err := b.dbDao.Transaction(func(tx *gorm.DB) error {
		if len(subAccountIds) > 0 {
			if err := tx.Where("account_id IN(?)", subAccountIds).
				Delete(&dao.TableRecordsInfo{}).Error; err != nil {
				return err
			}
		}
		if len(records) > 0 {
			if err := tx.Create(&records).Error; err != nil {
				return err
			}
		}
		if len(accountInfos) > 0 {
			if err := tx.Clauses(clause.Insert{
				Modifier: "IGNORE",
			}).Create(&accountInfos).Error; err != nil {
				return err
			}
		}

		if len(smtInfos) > 0 {
			if err := tx.Where("account_id IN(?)", subAccountIds).
				Delete(&dao.TableSmtInfo{}).Error; err != nil {
				return err
			}
			if err := tx.Clauses(clause.Insert{
				Modifier: "IGNORE",
			}).Create(&smtInfos).Error; err != nil {
				return err
			}
		}

		if err := tx.Clauses(clause.Insert{
			Modifier: "IGNORE",
		}).Create(&transactionInfo).Error; err != nil {
			return err
		}

		for _, v := range createBuilderMap {
			if v.EditKey != common.EditKeyCustomRule {
				continue
			}
			if len(v.EditValue) != 28 {
				return fmt.Errorf("edit_key: %s edit_value: %s is invalid", v.EditKey, common.Bytes2Hex(v.EditValue))
			}
			providerId := common.Bytes2Hex(v.EditValue[:20])
			price, err := molecule.Bytes2GoU64(v.EditValue[20:])
			if err != nil {
				return err
			}
			if err := tx.Clauses(clause.Insert{
				Modifier: "IGNORE",
			}).Create(&dao.TableSubAccountAutoMintStatement{
				BlockNumber:       req.BlockNumber,
				TxHash:            req.TxHash,
				WitnessIndex:      v.Index,
				ParentAccountId:   parentAccountId,
				ServiceProviderId: providerId,
				Price:             decimal.NewFromInt(int64(price)),
				BlockTimestamp:    req.BlockTimestamp,
				TxType:            dao.SubAccountAutoMintTxTypeIncome,
				SubAction:         v.Action,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("UpdateSubAccountForCreate err: %s", err.Error())
	}
	return nil
}

func (b *BlockParser) actionUpdateSubAccountForRenew(req FuncTransactionHandleReq, renewBuilderMap map[string]*witness.SubAccountNew) error {
	if len(renewBuilderMap) == 0 {
		return nil
	}
	// check sub_account config custom-script-args or not
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
	renewPriceConfig, err := builderConfig.RenewSubAccountPrice()
	if err != nil {
		return fmt.Errorf("RenewSubAccountPrice err: %s", err.Error())
	}

	var capacity uint64
	var parentAccount string
	var smtInfos []dao.TableSmtInfo
	var accountInfos []dao.TableAccountInfo

	for _, v := range renewBuilderMap {
		if parentAccount == "" {
			parentAccount = v.Account[strings.Index(v.Account, ".")+1:]
		}

		subAcc, err := b.dbDao.GetAccountInfoByAccountId(v.CurrentSubAccountData.AccountId)
		if err != nil {
			return err
		}
		if subAcc.Id == 0 {
			return fmt.Errorf("account: [%s] no exist", v.Account)
		}

		renewPrice := uint64(0)
		switch v.EditKey {
		case common.EditKeyManual:
			renewPrice = (v.CurrentSubAccountData.ExpiredAt - subAcc.ExpiredAt) / uint64(common.OneYearSec) * renewPriceConfig
		case common.EditKeyCustomRule:
			renewPrice, err = molecule.Bytes2GoU64(v.EditValue[28:])
			if err != nil {
				return err
			}
		}
		capacity += renewPrice

		accountInfos = append(accountInfos, dao.TableAccountInfo{
			Id:          subAcc.Id,
			BlockNumber: req.BlockNumber,
			Outpoint:    common.OutPoint2String(req.TxHash, 0),
			Nonce:       v.CurrentSubAccountData.Nonce,
			ExpiredAt:   v.CurrentSubAccountData.ExpiredAt,
		})

		value, err := v.CurrentSubAccountData.ToH256()
		if err != nil {
			return fmt.Errorf("CurrentSubAccountData.ToH256() err: %s", err.Error())
		}
		smtInfos = append(smtInfos, dao.TableSmtInfo{
			BlockNumber:  req.BlockNumber,
			Outpoint:     subAccountCellOutpoint,
			AccountId:    v.CurrentSubAccountData.AccountId,
			LeafDataHash: common.Bytes2Hex(value),
		})
	}

	ownerHex, _, err := b.dasCore.Daf().ScriptToHex(req.Tx.Outputs[len(req.Tx.Outputs)-1].Lock)
	if err != nil {
		return fmt.Errorf("ArgsToHex err: %s", err.Error())
	}

	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      parentAccountId,
		Account:        parentAccount,
		Action:         common.DasActionRenewSubAccount,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      ownerHex.ChainType,
		Address:        ownerHex.AddressHex,
		Capacity:       capacity,
		Outpoint:       subAccountCellOutpoint,
		BlockTimestamp: req.BlockTimestamp,
	}

	if err := b.dbDao.Transaction(func(tx *gorm.DB) error {
		for i := range accountInfos {
			accountInfo := accountInfos[i]
			if err := tx.Where("id=?", accountInfo.Id).Updates(&accountInfo).Error; err != nil {
				return err
			}
		}

		for i := range smtInfos {
			smtInfo := smtInfos[i]
			if err := tx.Where("account_id = ?", smtInfo.AccountId).Updates(&smtInfo).Error; err != nil {
				return err
			}
		}

		if err := tx.Clauses(clause.Insert{
			Modifier: "IGNORE",
		}).Create(&transactionInfo).Error; err != nil {
			return err
		}

		for _, v := range renewBuilderMap {
			if v.EditKey != common.EditKeyCustomRule {
				continue
			}
			//expiredAt, _ := molecule.Bytes2GoU64(v.EditValue[:8])
			providerId := common.Bytes2Hex(v.EditValue[8:28])
			price, err := molecule.Bytes2GoU64(v.EditValue[28:])
			if err != nil {
				return err
			}
			if err := tx.Create(&dao.TableSubAccountAutoMintStatement{
				BlockNumber:       req.BlockNumber,
				TxHash:            req.TxHash,
				WitnessIndex:      v.Index,
				ParentAccountId:   parentAccountId,
				ServiceProviderId: providerId,
				Price:             decimal.NewFromInt(int64(price)),
				BlockTimestamp:    req.BlockTimestamp,
				TxType:            dao.SubAccountAutoMintTxTypeIncome,
				SubAction:         common.SubActionRenew,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("UpdateSubAccountForRenew err: %s", err.Error())
	}
	return nil
}

func (b *BlockParser) actionUpdateSubAccountForEdit(req FuncTransactionHandleReq, editBuilderMap map[string]*witness.SubAccountNew) error {
	if len(editBuilderMap) == 0 {
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

		value, err := builder.CurrentSubAccountData.ToH256()
		if err != nil {
			return fmt.Errorf("CurrentSubAccountData.ToH256() err: %s", err.Error())
		}
		smtInfo := dao.TableSmtInfo{
			BlockNumber:  req.BlockNumber,
			Outpoint:     outpoint,
			AccountId:    builder.SubAccountData.AccountId,
			LeafDataHash: common.Bytes2Hex(value),
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
			accountInfo.ManagerSubAid = oHex.DasSubAlgorithmId
			accountInfo.OwnerChainType = oHex.ChainType
			accountInfo.Owner = oHex.AddressHex
			accountInfo.ManagerAlgorithmId = mHex.DasAlgorithmId
			accountInfo.ManagerSubAid = mHex.DasSubAlgorithmId
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
			accountInfo.ManagerSubAid = mHex.DasSubAlgorithmId
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

func (b *BlockParser) actionUpdateSubAccountForApproval(req FuncTransactionHandleReq, approvalBuilderMap map[string]*witness.SubAccountNew) error {
	if len(approvalBuilderMap) == 0 {
		return nil
	}

	contractSub, err := core.GetDasContractInfo(common.DASContractNameSubAccountCellType)
	if err != nil {
		return fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	var subAccountCellOutpoint string
	for i, v := range req.Tx.Outputs {
		if v.Type != nil && contractSub.IsSameTypeId(v.Type.CodeHash) {
			subAccountCellOutpoint = common.OutPoint2String(req.TxHash, uint(i))
		}
	}

	indexTx := uint(0)
	var parentAccount string
	var txs []dao.TableTransactionInfo
	var smtInfos []dao.TableSmtInfo
	var accountInfos []map[string]interface{}

	for _, v := range approvalBuilderMap {
		if parentAccount == "" {
			parentAccount = v.Account[strings.Index(v.Account, ".")+1:]
		}

		value, err := v.CurrentSubAccountData.ToH256()
		if err != nil {
			return fmt.Errorf("CurrentSubAccountData.ToH256() err: %s", err.Error())
		}
		smtInfos = append(smtInfos, dao.TableSmtInfo{
			BlockNumber:  req.BlockNumber,
			Outpoint:     subAccountCellOutpoint,
			AccountId:    v.CurrentSubAccountData.AccountId,
			LeafDataHash: common.Bytes2Hex(value),
		})

		oHex, _, err := b.dasCore.Daf().ScriptToHex(v.CurrentSubAccountData.Lock)
		if err != nil {
			return fmt.Errorf("ScriptToHex err: %s", err.Error())
		}
		outpointTx := common.OutPoint2String(req.TxHash, indexTx)
		txInfo := dao.TableTransactionInfo{
			BlockNumber:    req.BlockNumber,
			AccountId:      v.SubAccountData.AccountId,
			Account:        v.SubAccountData.Account(),
			Action:         v.Action,
			ServiceType:    dao.ServiceTypeSubAccount,
			ChainType:      oHex.ChainType,
			Address:        oHex.AddressHex,
			Outpoint:       outpointTx,
			BlockTimestamp: req.BlockTimestamp,
		}
		txs = append(txs, txInfo)
		indexTx++

		accountInfo := map[string]interface{}{
			"block_number": req.BlockNumber,
			"outpoint":     subAccountCellOutpoint,
			"account_id":   v.CurrentSubAccountData.AccountId,
			"nonce":        v.CurrentSubAccountData.Nonce,
			"action":       v.Action,
		}

		switch v.Action {
		case common.SubActionCreateApproval:
			accountInfo["status"] = uint8(dao.AccountStatusApproval)
		case common.SubActionRevokeApproval:
			accountInfo["status"] = uint8(dao.AccountStatusNormal)
		case common.SubActionFullfillApproval:
			approval := v.CurrentSubAccountData.AccountApproval
			switch approval.Action {
			case witness.AccountApprovalActionTransfer:
				owner, manager, err := b.dasCore.Daf().ScriptToHex(approval.Params.Transfer.ToLock)
				if err != nil {
					return err
				}
				accountInfo["status"] = uint8(dao.AccountStatusNormal)
				accountInfo["owner"] = owner.AddressHex
				accountInfo["owner_chain_type"] = owner.ChainType
				accountInfo["owner_algorithm_id"] = owner.DasAlgorithmId
				accountInfo["manager"] = manager.AddressHex
				accountInfo["manager_chain_type"] = manager.ChainType
				accountInfo["manager_algorithm_id"] = manager.DasAlgorithmId
			}
		}
		accountInfos = append(accountInfos, accountInfo)
	}
	if err := b.dbDao.ApprovalSubAccount(accountInfos, smtInfos, txs); err != nil {
		return err
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
			OwnerSubAid:          ownerHex.DasSubAlgorithmId,
			ManagerChainType:     managerHex.ChainType,
			Manager:              managerHex.AddressHex,
			ManagerAlgorithmId:   managerHex.DasAlgorithmId,
			ManagerSubAid:        managerHex.DasSubAlgorithmId,
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
		value, err := v.SubAccountData.ToH256()
		if err != nil {
			resp.Err = fmt.Errorf("SubAccountData.ToH256() err: %s", err.Error())
			return
		}
		smtInfos = append(smtInfos, dao.TableSmtInfo{
			BlockNumber:     req.BlockNumber,
			Outpoint:        common.OutPoint2String(req.TxHash, 1),
			AccountId:       v.SubAccountData.AccountId,
			ParentAccountId: parentAccountId,
			LeafDataHash:    common.Bytes2Hex(value),
		})
		capacity += (v.SubAccountData.ExpiredAt - v.SubAccountData.RegisteredAt) / uint64(common.OneYearSec) * newPrice
	}

	ownerHex, _, err := b.dasCore.Daf().ScriptToHex(req.Tx.Outputs[len(req.Tx.Outputs)-1].Lock)
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

	if err := b.actionUpdateSubAccountForEdit(req, builderMap); err != nil {
		resp.Err = fmt.Errorf("edit err: %s", err.Error())
		return
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

func (b *BlockParser) ActionCollectSubAccountChannelProfit(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DASContractNameSubAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		return
	}
	log.Info("ActionCollectSubAccountChannelProfit:", req.BlockNumber, req.TxHash)

	parentAccountId := common.Bytes2Hex(req.Tx.Outputs[0].Type.Args)

	if err := b.dbDao.Transaction(func(tx *gorm.DB) error {
		for i := 1; i < len(req.Tx.Outputs)-1; i++ {
			providerId := common.Bytes2Hex(req.Tx.Outputs[i].Lock.Args)
			price := req.Tx.Outputs[i].Capacity

			latest := &dao.TableSubAccountAutoMintStatement{}
			err := tx.Where("service_provider_id = ? AND parent_account_id = ? AND tx_type = ?", providerId, parentAccountId, dao.SubAccountAutoMintTxTypeExpenditure).Order("id desc").First(latest).Error
			if err != nil && err != gorm.ErrRecordNotFound {
				return err
			}

			rows, err := tx.Model(&dao.TableSubAccountAutoMintStatement{}).Where("service_provider_id = ? AND parent_account_id = ? AND block_number > ? AND tx_type = ?", providerId, parentAccountId, latest.BlockNumber, dao.SubAccountAutoMintTxTypeIncome).Rows()
			if err != nil {
				return err
			}

			var latestBlockNumber uint64
			var priceIncome decimal.Decimal
			for rows.Next() {
				tsas := &dao.TableSubAccountAutoMintStatement{}
				if err := tx.ScanRows(rows, tsas); err != nil {
					_ = rows.Close()
					return err
				}
				priceIncome = priceIncome.Add(tsas.Price)
				if priceIncome.IntPart() > int64(price) {
					_ = rows.Close()
					return fmt.Errorf("data exception priceIncome.IntPart(): %d > int64(price): %d", priceIncome.IntPart(), int64(price))
				}
				if priceIncome.Equal(decimal.NewFromInt(int64(price))) {
					latestBlockNumber = tsas.BlockNumber
				}
			}
			_ = rows.Close()

			if err := tx.Clauses(clause.Insert{
				Modifier: "IGNORE",
			}).Create(&dao.TableSubAccountAutoMintStatement{
				BlockNumber:       latestBlockNumber,
				TxHash:            req.TxHash,
				ParentAccountId:   parentAccountId,
				ServiceProviderId: providerId,
				Price:             decimal.NewFromInt(int64(price)),
				BlockTimestamp:    req.BlockTimestamp,
				TxType:            dao.SubAccountAutoMintTxTypeExpenditure,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		resp.Err = fmt.Errorf("transaction err: %s", err.Error())
	}
	return
}

func (b *BlockParser) ActionConfigSubAccount(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	isCV, index, err := CurrentVersionTx(req.Tx, common.DASContractNameSubAccountCellType)
	if err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warnf("not current version %s tx", common.DASContractNameSubAccountCellType)
		return
	}
	log.Info("ActionConfigSubAccount:", req.BlockNumber, req.TxHash)

	parentAccountId := common.Bytes2Hex(req.Tx.Outputs[index].Type.Args)

	if err := b.dbDao.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("account_id=?", parentAccountId).Delete(&dao.RuleConfig{}).Error; err != nil {
			return err
		}

		accountInfo := &dao.TableAccountInfo{}
		if err := tx.Where("account_id=?", parentAccountId).First(accountInfo).Error; err != nil {
			return err
		}

		if err := tx.Create(&dao.RuleConfig{
			Account:        accountInfo.Account,
			AccountId:      accountInfo.AccountId,
			TxHash:         req.TxHash,
			BlockNumber:    req.BlockNumber,
			BlockTimestamp: req.BlockTimestamp,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&dao.TableAccountInfo{}).Where("account_id=?", parentAccountId).Updates(map[string]interface{}{
			"outpoint": common.OutPoint2String(req.TxHash, 0),
		}).Error
	}); err != nil {
		resp.Err = fmt.Errorf("ActionConfigSubAccount err: %s", err.Error())
		return
	}
	return
}
