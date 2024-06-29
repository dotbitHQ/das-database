package snapshot

import (
	"das_database/config"
	"das_database/dao"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/types"
)

func (t *ToolSnapshot) addAccountPermissionsForDidCell(info dao.TableSnapshotTxInfo, tx *types.Transaction) error {
	log.Info("addAccountPermissionsForDidCell:", info.Action, info.Hash)
	var list []dao.TableSnapshotPermissionsInfo

	_, res, err := t.DasCore.TxToDidCellEntityAndAction(tx)
	if err != nil {
		return fmt.Errorf("TxToDidCellEntityAndAction err: %s", err.Error())
	}
	if len(res.Inputs) != len(res.Outputs) {
		return fmt.Errorf("len(res.Inputs) != len(res.Outputs)")
	}

	for k, v := range res.Inputs {
		n, ok := res.Outputs[k]
		if !ok {
			return fmt.Errorf("not exist did cell in outputs: %s", k)
		}
		_, cellDataNew, err := n.GetDataInfo()
		if err != nil {
			return fmt.Errorf("GetDataInfo err: %s", err.Error())
		}
		account := cellDataNew.Account
		accountId := common.Bytes2Hex(common.GetAccountIdByAccount(account))
		if !v.Lock.Equals(n.Lock) {
			mode := address.Mainnet
			if config.Cfg.Server.Net != common.DasNetTypeMainNet {
				mode = address.Testnet
			}
			addr, err := address.ConvertScriptToAddress(mode, n.Lock)
			if err != nil {
				return fmt.Errorf("address.ConvertScriptToAddress err: %s", err.Error())
			}

			tmp := dao.TableSnapshotPermissionsInfo{
				BlockNumber:        info.BlockNumber,
				AccountId:          accountId,
				Hash:               info.Hash,
				Account:            account,
				BlockTimestamp:     info.BlockTimestamp,
				Owner:              addr,
				Manager:            addr,
				OwnerAlgorithmId:   common.DasAlgorithmIdAnyLock,
				ManagerAlgorithmId: common.DasAlgorithmIdAnyLock,
				Status:             dao.AccountStatusOnUpgrade,
				ExpiredAt:          cellDataNew.ExpireAt,
			}
			list = append(list, tmp)
		}
	}

	if err := t.DbDao.CreateSnapshotPermissions(list); err != nil {
		return fmt.Errorf("CreateSnapshotPermissions err: %s", err.Error())
	}
	return nil
}

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
			ExpiredAt:          v.ExpiredAt,
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

		if v.Status == common.AccountStatusOnUpgrade {
			_, res, err := t.DasCore.TxToDidCellEntityAndAction(tx)
			if err != nil {
				return fmt.Errorf("TxToDidCellEntityAndAction err: %s", err.Error())
			}
			tmp.OwnerAlgorithmId = common.DasAlgorithmIdAnyLock
			tmp.ManagerAlgorithmId = common.DasAlgorithmIdAnyLock
			tmp.Status = dao.AccountStatusOnUpgrade
			for k, c := range res.Outputs {
				mode := address.Mainnet
				if config.Cfg.Server.Net != common.DasNetTypeMainNet {
					mode = address.Testnet
				}
				addr, err := address.ConvertScriptToAddress(mode, c.Lock)
				if err != nil {
					return fmt.Errorf("address.ConvertScriptToAddress err: %s[%s]", err.Error(), k)
				}
				tmp.Owner = addr
				tmp.Manager = addr
			}
		}
		list = append(list, tmp)
	}

	if err := t.DbDao.CreateSnapshotPermissions(list); err != nil {
		return fmt.Errorf("CreateSnapshotPermissions err: %s", err.Error())
	}
	return nil
}

func (t *ToolSnapshot) addSubAccountPermissionsByDasActionCreateSubAccount(info dao.TableSnapshotTxInfo, tx *types.Transaction) error {
	log.Info("addSubAccountPermissions:", info.Action, info.Hash)

	contractSub, err := core.GetDasContractInfo(common.DASContractNameSubAccountCellType)
	if err != nil {
		return fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	var parentAccountId string
	for _, v := range tx.Outputs {
		if v.Type != nil && contractSub.IsSameTypeId(v.Type.CodeHash) {
			parentAccountId = common.Bytes2Hex(v.Type.Args)
			break
		}
	}

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
			ParentAccountId:    parentAccountId,
			Hash:               info.Hash,
			Account:            v.Account,
			BlockTimestamp:     info.BlockTimestamp,
			Owner:              owner.AddressHex,
			Manager:            manager.AddressHex,
			OwnerAlgorithmId:   owner.DasAlgorithmId,
			ManagerAlgorithmId: manager.DasAlgorithmId,
			ExpiredAt:          v.CurrentSubAccountData.ExpiredAt,
		}
		list = append(list, tmp)
	}

	if err := t.DbDao.CreateSnapshotPermissions(list); err != nil {
		return fmt.Errorf("CreateSnapshotPermissions err: %s", err.Error())
	}

	return nil
}

func (t *ToolSnapshot) addSubAccountPermissionsByDasActionEditSubAccount(info dao.TableSnapshotTxInfo, tx *types.Transaction) error {
	log.Info("addSubAccountPermissions:", info.Action, info.Hash)

	contractSub, err := core.GetDasContractInfo(common.DASContractNameSubAccountCellType)
	if err != nil {
		return fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	var parentAccountId string
	for _, v := range tx.Outputs {
		if v.Type != nil && contractSub.IsSameTypeId(v.Type.CodeHash) {
			parentAccountId = common.Bytes2Hex(v.Type.Args)
			break
		}
	}

	var sanb witness.SubAccountNewBuilder
	mapSubAcc, err := sanb.SubAccountNewMapFromTx(tx)
	if err != nil {
		return fmt.Errorf("SubAccountNewMapFromTx err: %s", err.Error())
	}
	var list []dao.TableSnapshotPermissionsInfo
	for k, v := range mapSubAcc {
		if v.EditKey != common.EditKeyOwner && v.EditKey != common.EditKeyManager {
			continue
		}

		owner, manager, err := t.DasCore.Daf().ArgsToHex(v.CurrentSubAccountData.Lock.Args)
		if err != nil {
			return fmt.Errorf("ArgsToHex err: %s", err.Error())
		}
		tmp := dao.TableSnapshotPermissionsInfo{
			BlockNumber:        info.BlockNumber,
			AccountId:          k,
			ParentAccountId:    parentAccountId,
			Hash:               info.Hash,
			Account:            v.Account,
			BlockTimestamp:     info.BlockTimestamp,
			Owner:              owner.AddressHex,
			Manager:            manager.AddressHex,
			OwnerAlgorithmId:   owner.DasAlgorithmId,
			ManagerAlgorithmId: manager.DasAlgorithmId,
			ExpiredAt:          v.CurrentSubAccountData.ExpiredAt,
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

	contractSub, err := core.GetDasContractInfo(common.DASContractNameSubAccountCellType)
	if err != nil {
		return fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	var parentAccountId string
	for _, v := range tx.Outputs {
		if v.Type != nil && contractSub.IsSameTypeId(v.Type.CodeHash) {
			parentAccountId = common.Bytes2Hex(v.Type.Args)
			break
		}
	}

	var sanb witness.SubAccountNewBuilder
	mapSubAcc, err := sanb.SubAccountNewMapFromTx(tx)
	if err != nil {
		return fmt.Errorf("SubAccountNewMapFromTx err: %s", err.Error())
	}
	var list []dao.TableSnapshotPermissionsInfo
	for k, v := range mapSubAcc {
		if info.Action == common.DasActionUpdateSubAccount && v.Action == common.SubActionEdit && v.EditKey != common.EditKeyOwner && v.EditKey != common.EditKeyManager {
			continue
		}
		var owner, manager core.DasAddressHex
		if v.Action == common.SubActionRecycle {
			owner, manager, err = t.DasCore.Daf().ArgsToHex(v.SubAccountData.Lock.Args)
			if err != nil {
				return fmt.Errorf("ArgsToHex err: %s", err.Error())
			}
		} else {
			owner, manager, err = t.DasCore.Daf().ArgsToHex(v.CurrentSubAccountData.Lock.Args)
			if err != nil {
				return fmt.Errorf("ArgsToHex err: %s", err.Error())
			}
		}
		tmp := dao.TableSnapshotPermissionsInfo{
			BlockNumber:        info.BlockNumber,
			AccountId:          k,
			ParentAccountId:    parentAccountId,
			Hash:               info.Hash,
			Account:            v.Account,
			BlockTimestamp:     info.BlockTimestamp,
			Owner:              owner.AddressHex,
			Manager:            manager.AddressHex,
			OwnerAlgorithmId:   owner.DasAlgorithmId,
			ManagerAlgorithmId: manager.DasAlgorithmId,
			ExpiredAt:          v.CurrentSubAccountData.ExpiredAt,
		}
		if v.Action == common.SubActionRecycle {
			tmp.Status = dao.AccountStatusRecycle
		}
		list = append(list, tmp)
	}

	if err := t.DbDao.CreateSnapshotPermissions(list); err != nil {
		return fmt.Errorf("CreateSnapshotPermissions err: %s", err.Error())
	}

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
			ExpiredAt:          v.ExpiredAt,
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
			ExpiredAt:          v.ExpiredAt,
		}
		list = append(list, tmp)
	}

	if err := t.DbDao.CreateSnapshotPermissions(list); err != nil {
		return fmt.Errorf("CreateSnapshotPermissions err: %s", err.Error())
	}
	return nil
}
