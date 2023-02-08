package snapshot

import (
	"das_database/dao"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/nervosnetwork/ckb-sdk-go/types"
)

func (t *ToolSnapshot) addAccountRegisterByDasActionConfirmProposal(info dao.TableSnapshotTxInfo, tx *types.Transaction) error {
	log.Info("addAccountRegister:", info.Action, info.Hash)

	mapOldAcc, err := witness.AccountIdCellDataBuilderFromTx(tx, common.DataTypeOld)
	if err != nil {
		return fmt.Errorf("AccountIdCellDataBuilderFromTx err: %s", err.Error())
	}

	mapNewAcc, err := witness.AccountIdCellDataBuilderFromTx(tx, common.DataTypeNew)
	if err != nil {
		return fmt.Errorf("AccountIdCellDataBuilderFromTx err: %s", err.Error())
	}

	var list []dao.TableSnapshotRegisterInfo
	for k, v := range mapNewAcc {
		if _, ok := mapOldAcc[k]; ok {
			continue
		}

		owner, _, err := t.DasCore.Daf().ArgsToHex(tx.Outputs[v.Index].Lock.Args)
		if err != nil {
			return fmt.Errorf("ArgsToHex err: %s", err.Error())
		}
		tmp := dao.TableSnapshotRegisterInfo{
			BlockNumber:      info.BlockNumber,
			AccountId:        k,
			ParentAccountId:  "",
			Hash:             info.Hash,
			Account:          v.Account,
			BlockTimestamp:   info.BlockTimestamp,
			Owner:            owner.AddressHex,
			RegisteredAt:     v.RegisteredAt,
			OwnerAlgorithmId: owner.DasAlgorithmId,
		}
		list = append(list, tmp)
	}

	if err := t.DbDao.CreateSnapshotRegister(list); err != nil {
		return fmt.Errorf("CreateSnapshotRegister err: %s", err.Error())
	}

	return nil
}

func (t *ToolSnapshot) addSubAccountRegisterByDasActionCreateSubAccount(info dao.TableSnapshotTxInfo, tx *types.Transaction) error {
	log.Info("addSubAccountRegister:", info.Action, info.Hash)

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
	var list []dao.TableSnapshotRegisterInfo
	for k, v := range mapSubAcc {
		owner, _, err := t.DasCore.Daf().ArgsToHex(v.CurrentSubAccountData.Lock.Args)
		if err != nil {
			return fmt.Errorf("ArgsToHex err: %s", err.Error())
		}
		tmp := dao.TableSnapshotRegisterInfo{
			BlockNumber:      info.BlockNumber,
			AccountId:        k,
			ParentAccountId:  parentAccountId,
			Hash:             info.Hash,
			Account:          v.CurrentSubAccountData.Account(),
			BlockTimestamp:   info.BlockTimestamp,
			Owner:            owner.AddressHex,
			RegisteredAt:     v.CurrentSubAccountData.RegisteredAt,
			OwnerAlgorithmId: owner.DasAlgorithmId,
		}
		list = append(list, tmp)
	}

	if err := t.DbDao.CreateSnapshotRegister(list); err != nil {
		return fmt.Errorf("CreateSnapshotRegister err: %s", err.Error())
	}
	return nil
}

func (t *ToolSnapshot) addSubAccountRegister(info dao.TableSnapshotTxInfo, tx *types.Transaction) error {
	log.Info("addSubAccountRegister:", info.Action, info.Hash)

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
	var list []dao.TableSnapshotRegisterInfo
	for k, v := range mapSubAcc {
		if info.Action == common.DasActionUpdateSubAccount && v.Action != common.SubActionCreate {
			continue
		}
		owner, _, err := t.DasCore.Daf().ArgsToHex(v.CurrentSubAccountData.Lock.Args)
		if err != nil {
			return fmt.Errorf("ArgsToHex err: %s", err.Error())
		}
		tmp := dao.TableSnapshotRegisterInfo{
			BlockNumber:      info.BlockNumber,
			AccountId:        k,
			ParentAccountId:  parentAccountId,
			Hash:             info.Hash,
			Account:          v.CurrentSubAccountData.Account(),
			BlockTimestamp:   info.BlockTimestamp,
			Owner:            owner.AddressHex,
			RegisteredAt:     v.CurrentSubAccountData.RegisteredAt,
			OwnerAlgorithmId: owner.DasAlgorithmId,
		}
		list = append(list, tmp)
	}

	if err := t.DbDao.CreateSnapshotRegister(list); err != nil {
		return fmt.Errorf("CreateSnapshotRegister err: %s", err.Error())
	}
	return nil
}
