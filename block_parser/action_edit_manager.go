package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/witness"
	"github.com/scorpiotzh/toolib"
)

func (b *BlockParser) ActionEditManager(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version edit manager tx")
		return
	}

	log.Info("ActionEditManager:", req.BlockNumber, req.TxHash)

	accBuilder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	account := accBuilder.Account
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(account))
	ownerHex, managerHex, err := b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[accBuilder.Index].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}
	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      accountId,
		Account:        account,
		Action:         common.DasActionEditManager,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      ownerHex.ChainType,
		Address:        ownerHex.AddressHex,
		Capacity:       0,
		Outpoint:       common.OutPoint2String(req.TxHash, uint(accBuilder.Index)),
		BlockTimestamp: req.BlockTimestamp,
	}
	accountInfo := dao.TableAccountInfo{
		BlockNumber:        req.BlockNumber,
		Outpoint:           common.OutPoint2String(req.TxHash, uint(accBuilder.Index)),
		Account:            account,
		AccountId:          accountId,
		ManagerChainType:   managerHex.ChainType,
		Manager:            managerHex.AddressHex,
		ManagerAlgorithmId: managerHex.DasAlgorithmId,
	}

	log.Info("ActionEditManager:", account, managerHex.DasAlgorithmId, managerHex.ChainType, managerHex.AddressHex, transactionInfo.Address)

	if err := b.dbDao.EditManager(accountInfo, transactionInfo); err != nil {
		log.Error("EditManager err:", err.Error(), toolib.JsonString(transactionInfo))
		resp.Err = fmt.Errorf("EditManager err: %s", err.Error())
	}

	return
}
