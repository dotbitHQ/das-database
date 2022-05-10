package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/witness"
)

func (b *BlockParser) ActionAccountCrossChain(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version account cross chain tx")
		return
	}
	log.Info("ActionAccountCrossChain:", req.BlockNumber, req.TxHash, req.Action)

	accBuilder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	status := dao.AccountStatusOnLock
	if req.Action == common.DasActionUnlockAccountForCrossChain {
		status = dao.AccountStatusNormal
	}

	ownerHex, managerHex, err := b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[0].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}
	accountInfo := dao.TableAccountInfo{
		BlockNumber:        req.BlockNumber,
		Outpoint:           common.OutPoint2String(req.TxHash, 0),
		AccountId:          accBuilder.AccountId,
		OwnerChainType:     ownerHex.ChainType,
		Owner:              ownerHex.AddressHex,
		OwnerAlgorithmId:   ownerHex.DasAlgorithmId,
		ManagerChainType:   managerHex.ChainType,
		Manager:            managerHex.AddressHex,
		ManagerAlgorithmId: managerHex.DasAlgorithmId,
		Status:             status,
	}
	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      accBuilder.AccountId,
		Account:        accBuilder.Account,
		Action:         req.Action,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      ownerHex.ChainType,
		Address:        ownerHex.AddressHex,
		Capacity:       0,
		Outpoint:       common.OutPoint2String(req.TxHash, 0),
		BlockTimestamp: req.BlockTimestamp,
	}

	if err = b.dbDao.AccountCrossChain(accountInfo, transactionInfo); err != nil {
		log.Error("AccountCrossChain err:", err.Error(), req.TxHash, req.BlockNumber)
		resp.Err = fmt.Errorf("AccountCrossChain err: %s ", err.Error())
		return
	}

	return
}
