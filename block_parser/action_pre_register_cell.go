package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/witness"
)

func (b *BlockParser) ActionPreRegister(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNamePreAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version pre register tx")
		return
	}
	log.Info("ActionPreRegister:", req.BlockNumber, req.TxHash)

	preBuilder, err := witness.PreAccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("PreAccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	log.Info("ActionPreRegister:", preBuilder.Account)

	refundLock, _ := preBuilder.RefundLock()
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(preBuilder.Account))
	var transactionInfo = dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      accountId,
		Account:        preBuilder.Account,
		Action:         common.DasActionPreRegister,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      common.ChainTypeCkb,
		Address:        common.Bytes2Hex(refundLock.Args().RawData()), // refund lock(register itself)
		Outpoint:       common.OutPoint2String(req.TxHash, 0),
		Capacity:       req.Tx.Outputs[0].Capacity,
		BlockTimestamp: req.BlockTimestamp,
	}
	if err := b.dbDao.CreateTransactionInfo(transactionInfo); err != nil {
		log.Error("CreateTransactionInfo err:", err.Error(), req.TxHash, req.BlockNumber)
		resp.Err = fmt.Errorf("CreateTransactionInfo err: %s", err.Error())
		return
	}

	return
}
