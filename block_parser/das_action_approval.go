package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/witness"
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

	accBuilder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	resp.Err = b.dbDao.UpdateAccountInfo(accBuilder.AccountId, map[string]interface{}{
		"status": dao.AccountStatusApproval,
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

	accBuilder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	resp.Err = b.dbDao.UpdateAccountInfo(accBuilder.AccountId, map[string]interface{}{
		"status": dao.AccountStatusNormal,
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

	accBuilder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
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
		resp.Err = b.dbDao.UpdateAccountInfo(accBuilder.AccountId, map[string]interface{}{
			"status":               dao.AccountStatusNormal,
			"owner":                owner.AddressHex,
			"owner_chain_type":     owner.ChainType,
			"owner_algorithm_id":   owner.DasAlgorithmId,
			"manager":              manager.AddressHex,
			"manager_chain_type":   manager.ChainType,
			"manager_algorithm_id": manager.DasAlgorithmId,
		})
	}
	return
}
