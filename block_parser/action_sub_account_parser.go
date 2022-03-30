package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/DeAccountSystems/das-lib/witness"
	"strconv"
	"time"
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
	_, _, oCT, _, oA, _ := core.FormatDasLockToHexAddress(req.Tx.Outputs[builder.Index].Lock.Args)

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
		ChainType:      oCT,
		Address:        oA,
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
				Action:         common.DasActionCreateSubAccount,
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

	builderMap, err := witness.SubAccountBuilderMapFromTx(req.Tx)
	if err != nil {
		resp.Err = fmt.Errorf("SubAccountBuilderMapFromTx err: %s", err.Error())
		return
	}

	var accountInfos []dao.TableAccountInfo
	var smtInfos []dao.TableSmtInfo
	var transactionInfos []dao.TableTransactionInfo
	var index uint
	for _, v := range builderMap {
		oID, mID, oCT, mCT, oA, mA := core.FormatDasLockToHexAddress(v.SubAccount.Lock.Args)

		accountInfos = append(accountInfos, dao.TableAccountInfo{
			BlockNumber:          req.BlockNumber,
			Outpoint:             common.OutPoint2String(req.TxHash, 0),
			AccountId:            v.SubAccount.AccountId,
			ParentAccountId:      builder.AccountId,
			Account:              v.Account,
			OwnerChainType:       oCT,
			Owner:                oA,
			OwnerAlgorithmId:     oID,
			ManagerChainType:     mCT,
			Manager:              mA,
			ManagerAlgorithmId:   mID,
			Status:               dao.AccountStatus(v.SubAccount.Status),
			EnableSubAccount:     dao.EnableSubAccount(v.SubAccount.EnableSubAccount),
			RenewSubAccountPrice: v.SubAccount.RenewSubAccountPrice,
			Nonce:                v.SubAccount.Nonce,
			RegisteredAt:         v.SubAccount.RegisteredAt,
			ExpiredAt:            v.SubAccount.ExpiredAt,
			ConfirmProposalHash:  req.TxHash,
		})
		smtInfos = append(smtInfos, dao.TableSmtInfo{
			BlockNumber:     req.BlockNumber,
			Outpoint:        common.OutPoint2String(req.TxHash, 1),
			AccountId:       v.SubAccount.AccountId,
			ParentAccountId: builder.AccountId,
			LeafDataHash:    common.Bytes2Hex(v.SubAccount.ToH256()),
		})
		transactionInfos = append(transactionInfos, dao.TableTransactionInfo{
			BlockNumber:    req.BlockNumber,
			AccountId:      v.SubAccount.AccountId,
			Account:        v.Account,
			Action:         common.DasActionCreateSubAccount,
			ServiceType:    dao.ServiceTypeRegister,
			ChainType:      oCT,
			Address:        oA,
			Capacity:       req.Tx.Outputs[1].Capacity,
			Outpoint:       common.OutPoint2String(common.OutPoint2String(req.TxHash, 1), index),
			BlockTimestamp: req.BlockTimestamp,
		})
		index++
	}
	accountInfo := dao.TableAccountInfo{
		BlockNumber: req.BlockNumber,
		AccountId:   builder.AccountId,
		Outpoint:    common.OutPoint2String(req.TxHash, 0),
	}

	if err = b.dbDao.CreateSubAccount(incomeCellInfos, accountInfos, smtInfos, transactionInfos, accountInfo); err != nil {
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
		_, _, chainType, _, address, _ := core.FormatDasLockToHexAddress(builder.SubAccount.Lock.Args)
		outpoint := common.OutPoint2String(req.TxHash, 0)

		accountInfo := dao.TableAccountInfo{
			BlockNumber: req.BlockNumber,
			Outpoint:    outpoint,
			AccountId:   builder.SubAccount.AccountId,
			Nonce:       builder.SubAccount.Nonce,
		}
		smtInfo := dao.TableSmtInfo{
			BlockNumber:  req.BlockNumber,
			Outpoint:     outpoint,
			AccountId:    builder.SubAccount.AccountId,
			LeafDataHash: common.Bytes2Hex(builder.SubAccount.ToH256()),
		}
		transactionInfo := dao.TableTransactionInfo{
			BlockNumber:    req.BlockNumber,
			AccountId:      builder.SubAccount.AccountId,
			Account:        builder.Account,
			Action:         common.DasActionEditSubAccount,
			ServiceType:    dao.ServiceTypeRegister,
			ChainType:      chainType,
			Address:        address,
			Capacity:       req.Tx.Outputs[0].Capacity,
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
			oID, _, oCT, _, oA, _ := core.FormatDasLockToHexAddress(common.Hex2Bytes(subAccount.LockArgs))
			accountInfo.OwnerAlgorithmId = oID
			accountInfo.OwnerChainType = oCT
			accountInfo.Owner = oA
			transactionInfo.ChainType = oCT
			transactionInfo.Address = oA
			if err = b.dbDao.EditOwnerSubAccount(accountInfo, smtInfo, transactionInfo); err != nil {
				resp.Err = fmt.Errorf("EditOwnerSubAccount err: %s", err.Error())
			}
		case common.EditKeyManager:
			_, mID, _, mCT, _, mA := core.FormatDasLockToHexAddress(common.Hex2Bytes(subAccount.LockArgs))
			accountInfo.ManagerAlgorithmId = mID
			accountInfo.ManagerChainType = mCT
			accountInfo.Manager = mA
			if err = b.dbDao.EditManagerSubAccount(accountInfo, smtInfo, transactionInfo); err != nil {
				resp.Err = fmt.Errorf("EditManagerSubAccount err: %s", err.Error())
			}
		case common.EditKeyRecords:
			var recordsInfos []dao.TableRecordsInfo
			for _, v := range subAccount.Records {
				recordsInfos = append(recordsInfos, dao.TableRecordsInfo{
					AccountId: builder.SubAccount.AccountId,
					Account:   builder.Account,
					Key:       v.Key,
					Type:      v.Type,
					Label:     v.Label,
					Value:     v.Value,
					Ttl:       strconv.FormatUint(uint64(v.TTL), 10),
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

	var accountIds []string
	var accountInfos []dao.TableAccountInfo
	var smtInfos []dao.TableSmtInfo
	var transactionInfos []dao.TableTransactionInfo
	var index uint
	for _, builder := range builderMap {
		_, _, oCT, _, oA, _ := core.FormatDasLockToHexAddress(builder.SubAccount.Lock.Args)
		outpoint := common.OutPoint2String(req.TxHash, 0)

		accountInfo := dao.TableAccountInfo{
			BlockNumber: req.BlockNumber,
			Outpoint:    outpoint,
			AccountId:   builder.SubAccount.AccountId,
			Nonce:       builder.SubAccount.Nonce,
		}
		smtInfo := dao.TableSmtInfo{
			BlockNumber:  req.BlockNumber,
			Outpoint:     outpoint,
			AccountId:    builder.SubAccount.AccountId,
			LeafDataHash: common.Bytes2Hex(builder.SubAccount.ToH256()),
		}
		transactionInfo := dao.TableTransactionInfo{
			BlockNumber:    req.BlockNumber,
			AccountId:      builder.SubAccount.AccountId,
			Account:        builder.Account,
			Action:         common.DasActionRenewSubAccount,
			ServiceType:    dao.ServiceTypeRegister,
			ChainType:      oCT,
			Address:        oA,
			Capacity:       req.Tx.Outputs[0].Capacity,
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
			accountIds = append(accountIds, builder.SubAccount.AccountId)
			accountInfo.ExpiredAt = subAccount.ExpiredAt
			accountInfos = append(accountInfos, accountInfo)
			smtInfos = append(smtInfos, smtInfo)
			transactionInfos = append(transactionInfos, transactionInfo)
		}
	}

	if err = b.dbDao.RenewSubAccount(accountIds, accountInfos, smtInfos, transactionInfos); err != nil {
		resp.Err = fmt.Errorf("RenewSubAccount err: %s", err.Error())
		return
	}

	return
}

func (b *BlockParser) ActionRecycleSubAccount(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
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

		_, _, oCT, _, oA, _ := core.FormatDasLockToHexAddress(builder.SubAccount.Lock.Args)
		outpoint := common.OutPoint2String(req.TxHash, 0)

		accountIds = append(accountIds, builder.SubAccount.AccountId)
		transactionInfos = append(transactionInfos, dao.TableTransactionInfo{
			BlockNumber:    req.BlockNumber,
			AccountId:      builder.SubAccount.AccountId,
			Account:        builder.Account,
			Action:         common.DasActionRecycleSubAccount,
			ServiceType:    dao.ServiceTypeRegister,
			ChainType:      oCT,
			Address:        oA,
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
}
