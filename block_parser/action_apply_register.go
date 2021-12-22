package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/scorpiotzh/toolib"
)

func (b *BlockParser) ActionApplyRegister(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameApplyRegisterCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version apply register tx")
		return
	}

	log.Info("ActionApplyRegister:", req.BlockNumber, req.TxHash)

	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		Action:         common.DasActionApplyRegister,
		ServiceType:    dao.ServiceTypeRegister,
		Outpoint:       common.OutPoint2String(req.TxHash, 0),
		ChainType:      common.ChainTypeCkb,
		Address:        common.Bytes2Hex(req.Tx.Outputs[0].Lock.Args),
		Capacity:       req.Tx.Outputs[0].Capacity,
		BlockTimestamp: req.BlockTimestamp,
	}

	if err := b.dbDao.CreateTransactionInfo(transactionInfo); err != nil {
		log.Error("CreateTransactionInfo err:", err.Error(), toolib.JsonString(transactionInfo))
		resp.Err = fmt.Errorf("CreateTransactionInfo err: %s", err.Error())
		return
	}

	return
}
