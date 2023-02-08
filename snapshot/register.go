package snapshot

import (
	"das_database/dao"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
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

func (t *ToolSnapshot) addSubAccountRegister(info dao.TableSnapshotTxInfo, tx *types.Transaction) error {
	log.Info("addSubAccountRegister:", info.Action, info.Hash)

	// todo
	return nil
}
