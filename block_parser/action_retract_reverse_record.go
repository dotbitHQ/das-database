package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
)

func (b *BlockParser) ActionRetractReverseRecord(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {

	res, err := b.ckbClient.Client().GetTransaction(b.ctx, req.Tx.Inputs[0].PreviousOutput.TxHash)
	if err != nil {
		resp.Err = fmt.Errorf("GetTransaction err: %s", err.Error())
		return
	}
	if isCV, err := isCurrentVersionTx(res.Transaction, common.DasContractNameReverseRecordCellType); err != nil {
		resp.Err = fmt.Errorf("isisCurrentVersionTx err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version retract reverse record tx")
		return
	}

	log.Info("ActionRetractReverseRecord:", req.BlockNumber, req.TxHash)

	var listOutpoint []string
	for _, v := range req.Tx.Inputs {
		listOutpoint = append(listOutpoint, common.OutPointStruct2String(v.PreviousOutput))
	}

	ownerHex, _, err := b.dasCore.Daf().ArgsToHex(res.Transaction.Outputs[0].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}
	txInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      "",
		Account:        "",
		Action:         common.DasActionRetractReverseRecord,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      ownerHex.ChainType,
		Address:        ownerHex.AddressHex,
		Capacity:       req.Tx.OutputsCapacity(),
		Outpoint:       common.OutPoint2String(req.TxHash, 0),
		BlockTimestamp: req.BlockTimestamp,
	}

	if err := b.dbDao.RetractReverseRecord(listOutpoint, txInfo); err != nil {
		resp.Err = fmt.Errorf("RetractReverseRecord err: %s", err.Error())
		return
	}

	return
}
