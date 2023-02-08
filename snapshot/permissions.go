package snapshot

import (
	"das_database/dao"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/nervosnetwork/ckb-sdk-go/types"
)

func (t *ToolSnapshot) addAccountPermissions(info dao.TableSnapshotTxInfo, tx *types.Transaction) error {
	log.Info("addAccountPermissions:", info.Action, info.Hash)

	mapAcc, err := witness.AccountIdCellDataBuilderFromTx(tx, common.DataTypeNew)
	if err != nil {
		return fmt.Errorf("AccountIdCellDataBuilderFromTx err: %s", err.Error())
	}
	var list []dao.TableSnapshotPermissionsInfo
	for k, v := range mapAcc {
		owner, manager, err := t.DasCore.Daf().ArgsToHex(tx.Outputs[v.Index].Lock.Args)
		if err != nil {
			return fmt.Errorf("ArgsToHex err: %s", err.Error())
		}
		tmp := dao.TableSnapshotPermissionsInfo{
			BlockNumber:        info.BlockNumber,
			AccountId:          k,
			Hash:               info.Hash,
			Account:            v.Account,
			BlockTimestamp:     info.BlockTimestamp,
			Owner:              owner.AddressHex,
			Manager:            manager.AddressHex,
			OwnerAlgorithmId:   owner.DasAlgorithmId,
			ManagerAlgorithmId: manager.DasAlgorithmId,
			Status:             dao.AccountStatusNormal,
		}
		if info.Action == common.DasActionStartAccountSale {
			tmp.Status = dao.AccountStatusOnSale
		} else if info.Action == common.DasActionLockAccountForCrossChain {
			tmp.Status = dao.AccountStatusOnLock
			if owner.AddressHex == "0x0000000000000000000000000000000000000000" {
				log.Info("cross black address:", info.Action, info.Hash)

				refOutpoint := tx.Inputs[0].PreviousOutput
				res, err := t.DasCore.Client().GetTransaction(t.Ctx, refOutpoint.TxHash)
				if err != nil {
					return fmt.Errorf("GetTransaction err: %s", err.Error())
				}
				owner, manager, err = t.DasCore.Daf().ArgsToHex(res.Transaction.Outputs[refOutpoint.Index].Lock.Args)
				if err != nil {
					return fmt.Errorf("ArgsToHex err: %s", err.Error())
				}
				tmp.Owner = owner.AddressHex
				tmp.OwnerAlgorithmId = owner.DasAlgorithmId
				tmp.Manager = manager.AddressHex
				tmp.ManagerAlgorithmId = manager.DasAlgorithmId
			}
		}
		list = append(list, tmp)
	}

	if err := t.DbDao.CreateSnapshotPermissions(list); err != nil {
		return fmt.Errorf("CreateSnapshotPermissions err: %s", err.Error())
	}
	return nil
}

func (t *ToolSnapshot) addSubAccountPermissions(info dao.TableSnapshotTxInfo, tx *types.Transaction) error {
	log.Info("addSubAccountPermissions:", info.Action, info.Hash)

	var sanb witness.SubAccountNewBuilder
	mapSubAcc, err := sanb.SubAccountNewMapFromTx(tx)
	if err != nil {
		return fmt.Errorf("SubAccountNewMapFromTx err: %s", err.Error())
	}
	var list []dao.TableSnapshotPermissionsInfo
	for k, v := range mapSubAcc {
		owner, manager, err := t.DasCore.Daf().ArgsToHex(v.CurrentSubAccountData.Lock.Args)
		if err != nil {
			return fmt.Errorf("ArgsToHex err: %s", err.Error())
		}
		tmp := dao.TableSnapshotPermissionsInfo{
			BlockNumber:        info.BlockNumber,
			AccountId:          k,
			Hash:               info.Hash,
			Account:            v.Account,
			BlockTimestamp:     info.BlockTimestamp,
			Owner:              owner.AddressHex,
			Manager:            manager.AddressHex,
			OwnerAlgorithmId:   owner.DasAlgorithmId,
			ManagerAlgorithmId: manager.DasAlgorithmId,
		}
		list = append(list, tmp)
	}

	if err := t.DbDao.CreateSnapshotPermissions(list); err != nil {
		return fmt.Errorf("CreateSnapshotPermissions err: %s", err.Error())
	}

	// todo
	return nil
}

func (t *ToolSnapshot) addAccountPermissionsByDasActionConfirmProposal(info dao.TableSnapshotTxInfo, tx *types.Transaction) error {
	log.Info("addAccountPermissionsByDasActionConfirmProposal:", info.Action, info.Hash)

	mapOldAcc, err := witness.AccountIdCellDataBuilderFromTx(tx, common.DataTypeOld)
	if err != nil {
		return fmt.Errorf("AccountIdCellDataBuilderFromTx err: %s", err.Error())
	}

	mapNewAcc, err := witness.AccountIdCellDataBuilderFromTx(tx, common.DataTypeNew)
	if err != nil {
		return fmt.Errorf("AccountIdCellDataBuilderFromTx err: %s", err.Error())
	}

	var list []dao.TableSnapshotPermissionsInfo
	for k, v := range mapNewAcc {
		if _, ok := mapOldAcc[k]; ok {
			continue
		}

		owner, manager, err := t.DasCore.Daf().ArgsToHex(tx.Outputs[v.Index].Lock.Args)
		if err != nil {
			return fmt.Errorf("ArgsToHex err: %s", err.Error())
		}
		tmp := dao.TableSnapshotPermissionsInfo{
			BlockNumber:        info.BlockNumber,
			AccountId:          k,
			Hash:               info.Hash,
			Account:            v.Account,
			BlockTimestamp:     info.BlockTimestamp,
			Owner:              owner.AddressHex,
			Manager:            manager.AddressHex,
			OwnerAlgorithmId:   owner.DasAlgorithmId,
			ManagerAlgorithmId: manager.DasAlgorithmId,
			Status:             dao.AccountStatusNormal,
		}
		list = append(list, tmp)
	}

	if err := t.DbDao.CreateSnapshotPermissions(list); err != nil {
		return fmt.Errorf("CreateSnapshotPermissions err: %s", err.Error())
	}

	return nil
}

func (t *ToolSnapshot) addAccountPermissionsByDasActionRecycleExpiredAccount(info dao.TableSnapshotTxInfo, tx *types.Transaction) error {
	log.Info("addAccountPermissionsByDasActionRecycleExpiredAccount:", info.Action, info.Hash)

	mapOldAcc, err := witness.AccountIdCellDataBuilderFromTx(tx, common.DataTypeOld)
	if err != nil {
		return fmt.Errorf("AccountIdCellDataBuilderFromTx err: %s", err.Error())
	}

	mapNewAcc, err := witness.AccountIdCellDataBuilderFromTx(tx, common.DataTypeNew)
	if err != nil {
		return fmt.Errorf("AccountIdCellDataBuilderFromTx err: %s", err.Error())
	}

	var list []dao.TableSnapshotPermissionsInfo
	for k, v := range mapOldAcc {
		if _, ok := mapNewAcc[k]; ok {
			continue
		}
		refOutpoint := tx.Inputs[v.Index].PreviousOutput
		res, err := t.DasCore.Client().GetTransaction(t.Ctx, refOutpoint.TxHash)
		if err != nil {
			return fmt.Errorf("GetTransaction err: %s", err.Error())
		}
		owner, manager, err := t.DasCore.Daf().ArgsToHex(res.Transaction.Outputs[refOutpoint.Index].Lock.Args)
		if err != nil {
			return fmt.Errorf("ArgsToHex err: %s", err.Error())
		}

		tmp := dao.TableSnapshotPermissionsInfo{
			BlockNumber:        info.BlockNumber,
			AccountId:          k,
			Hash:               info.Hash,
			Account:            v.Account,
			BlockTimestamp:     info.BlockTimestamp,
			Owner:              owner.AddressHex,
			Manager:            manager.AddressHex,
			OwnerAlgorithmId:   owner.DasAlgorithmId,
			ManagerAlgorithmId: manager.DasAlgorithmId,
			Status:             dao.AccountStatusRecycle,
		}
		list = append(list, tmp)
	}

	if err := t.DbDao.CreateSnapshotPermissions(list); err != nil {
		return fmt.Errorf("CreateSnapshotPermissions err: %s", err.Error())
	}
	return nil
}
