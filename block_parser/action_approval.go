package block_parser

import (
	"das_database/config"
	"das_database/dao"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/witness"
	"gorm.io/gorm"
)

func (b *BlockParser) DasActionCreateApproval(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version edit records tx")
		return
	}
	log.Info("DasActionCreateApproval:", req.BlockNumber, req.TxHash)

	refOutpoint, outpoint, err := b.getOutpoint(req, common.DasContractNameAccountCellType)
	if err != nil {
		resp.Err = fmt.Errorf("getOutpoint err: %s", err.Error())
		return
	}

	accBuilder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}

	accountInfo, err := b.dbDao.GetAccountInfoByAccountId(accBuilder.AccountId)
	if err != nil {
		resp.Err = fmt.Errorf("GetAccountInfoByAccountId err: %s", err.Error())
		return
	}

	transfer := accBuilder.AccountApproval.Params.Transfer
	dasf := core.DasAddressFormat{DasNetType: config.Cfg.Server.Net}
	toHex, _, err := dasf.ScriptToHex(transfer.ToLock)
	if err != nil {
		resp.Err = fmt.Errorf("ScriptToHex err: %s", err.Error())
		return
	}
	toNormal, err := dasf.HexToNormal(toHex)
	if err != nil {
		resp.Err = fmt.Errorf("HexToNormal err: %s", err.Error())
		return
	}

	platformHex, _, err := dasf.ScriptToHex(transfer.PlatformLock)
	if err != nil {
		resp.Err = fmt.Errorf("ScriptToHex err: %s", err.Error())
		return
	}

	approval := &dao.ApprovalInfo{
		BlockNumber:      req.BlockNumber,
		RefOutpoint:      refOutpoint,
		Outpoint:         outpoint,
		Account:          accBuilder.Account,
		AccountID:        accBuilder.AccountId,
		Platform:         platformHex.AddressHex,
		OwnerAlgorithmID: accountInfo.OwnerAlgorithmId,
		Owner:            accountInfo.Owner,
		ToAlgorithmID:    toHex.DasAlgorithmId,
		To:               toNormal.AddressNormal,
		ProtectedUntil:   transfer.ProtectedUntil,
		SealedUntil:      transfer.SealedUntil,
		MaxDelayCount:    transfer.DelayCountRemain,
		Status:           dao.ApprovalStatusEnable,
	}

	resp.Err = b.dbDao.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&dao.TableAccountInfo{}).Where("account_id=?", accountInfo.AccountId).Updates(map[string]interface{}{
			"outpoint":     outpoint,
			"block_number": req.BlockNumber,
			"status":       dao.AccountStatusApproval,
		}).Error; err != nil {
			return err
		}
		return tx.Create(approval).Error
	})
	return
}

func (b *BlockParser) DasActionDelayApproval(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version edit records tx")
		return
	}
	log.Info("DasActionDelayApproval:", req.BlockNumber, req.TxHash)

	refOutpoint, outpoint, err := b.getOutpoint(req, common.DasContractNameAccountCellType)
	if err != nil {
		resp.Err = fmt.Errorf("getOutpoint err: %s", err.Error())
		return
	}

	accBuilder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	resp.Err = b.dbDao.UpdateAccountInfo(accBuilder.AccountId, map[string]interface{}{
		"outpoint":     common.OutPoint2String(req.TxHash, 0),
		"block_number": req.BlockNumber,
	})

	transfer := accBuilder.AccountApproval.Params.Transfer

	approval, err := b.dbDao.GetAccountPendingApproval(accBuilder.AccountId)
	if err != nil {
		resp.Err = fmt.Errorf("GetAccountApprovalByOutpoint err: %s", err.Error())
		return
	}
	if approval.ID == 0 {
		resp.Err = fmt.Errorf("approval not found")
		return
	}
	approval.SealedUntil = transfer.SealedUntil
	approval.PostponedCount++

	resp.Err = b.dbDao.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&dao.TableAccountInfo{}).Where("account_id=?", accBuilder.AccountId).Updates(map[string]interface{}{
			"outpoint":     outpoint,
			"block_number": req.BlockNumber,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&dao.ApprovalInfo{}).Where("id=?", approval.ID).Updates(map[string]interface{}{
			"outpoint":        outpoint,
			"ref_outpoint":    refOutpoint,
			"sealed_until":    approval.SealedUntil,
			"postponed_count": approval.PostponedCount,
		}).Error
	})
	return
}

func (b *BlockParser) DasActionRevokeApproval(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version edit records tx")
		return
	}
	log.Info("DasActionRevokeApproval:", req.BlockNumber, req.TxHash)

	refOutpoint, outpoint, err := b.getOutpoint(req, common.DasContractNameAccountCellType)
	if err != nil {
		resp.Err = fmt.Errorf("getOutpoint err: %s", err.Error())
		return
	}

	accBuilder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeOld)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}

	approval, err := b.dbDao.GetAccountPendingApproval(accBuilder.AccountId)
	if err != nil {
		resp.Err = fmt.Errorf("GetAccountApprovalByOutpoint err: %s", err.Error())
		return
	}
	if approval.ID == 0 {
		resp.Err = fmt.Errorf("approval not found")
		return
	}
	resp.Err = b.dbDao.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&dao.TableAccountInfo{}).Where("account_id=?", accBuilder.AccountId).Updates(map[string]interface{}{
			"outpoint":     outpoint,
			"block_number": req.BlockNumber,
			"status":       dao.AccountStatusNormal,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&dao.ApprovalInfo{}).Where("id=?", approval.ID).Updates(map[string]interface{}{
			"outpoint":     outpoint,
			"ref_outpoint": refOutpoint,
			"status":       dao.ApprovalStatusRevoke,
		}).Error
	})
	return
}

func (b *BlockParser) DasActionFulfillApproval(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version edit records tx")
		return
	}
	log.Info("DasActionFulfillApproval:", req.BlockNumber, req.TxHash)

	refOutpoint, outpoint, err := b.getOutpoint(req, common.DasContractNameAccountCellType)
	if err != nil {
		resp.Err = fmt.Errorf("getOutpoint err: %s", err.Error())
		return
	}

	accBuilder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeOld)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}

	approval := accBuilder.AccountApproval
	switch approval.Action {
	case witness.AccountApprovalActionTransfer:
		owner, manager, err := b.dasCore.Daf().ScriptToHex(approval.Params.Transfer.ToLock)
		if err != nil {
			resp.Err = fmt.Errorf("ScriptToHex err: %s", err.Error())
			return
		}

		approvalInfo, err := b.dbDao.GetAccountPendingApproval(accBuilder.AccountId)
		if err != nil {
			resp.Err = fmt.Errorf("GetAccountApprovalByOutpoint err: %s", err.Error())
			return
		}
		if approvalInfo.ID == 0 {
			resp.Err = fmt.Errorf("approval not found")
			return
		}

		resp.Err = b.dbDao.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&dao.TableAccountInfo{}).Where("account_id=?", accBuilder.AccountId).Updates(map[string]interface{}{
				"outpoint":             common.OutPoint2String(req.TxHash, 0),
				"block_number":         req.BlockNumber,
				"status":               dao.AccountStatusNormal,
				"owner":                owner.AddressHex,
				"owner_chain_type":     owner.ChainType,
				"owner_algorithm_id":   owner.DasAlgorithmId,
				"manager":              manager.AddressHex,
				"manager_chain_type":   manager.ChainType,
				"manager_algorithm_id": manager.DasAlgorithmId,
			}).Error; err != nil {
				return err
			}
			if err := tx.Where("account_id = ?", accBuilder.AccountId).Delete(&dao.TableRecordsInfo{}).Error; err != nil {
				return err
			}
			return tx.Model(&dao.ApprovalInfo{}).Where("id=?", approvalInfo.ID).Updates(map[string]interface{}{
				"outpoint":     outpoint,
				"ref_outpoint": refOutpoint,
				"status":       dao.ApprovalStatusFulFill,
			}).Error
		})
	}
	return
}
