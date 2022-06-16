package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
)

func (b *BlockParser) ActionDeclareReverseRecord(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameReverseRecordCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version declare reverse record tx")
		return
	}
	log.Info("ActionDeclareReverseRecord:", req.BlockNumber, req.TxHash)

	account := string(req.Tx.OutputsData[0])
	oHex, _, err := b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[0].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}

	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(account))
	reverseInfo := dao.TableReverseInfo{
		BlockNumber:    req.BlockNumber,
		BlockTimestamp: req.BlockTimestamp,
		Outpoint:       common.OutPoint2String(req.TxHash, 0),
		AlgorithmId:    oHex.DasAlgorithmId,
		ChainType:      oHex.ChainType,
		Address:        oHex.AddressHex,
		Account:        account,
		AccountId:      accountId,
		Capacity:       req.Tx.Outputs[0].Capacity,
	}

	txInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      accountId,
		Account:        account,
		Action:         common.DasActionDeclareReverseRecord,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      oHex.ChainType,
		Address:        oHex.AddressHex,
		Capacity:       req.Tx.Outputs[0].Capacity,
		Outpoint:       common.OutPoint2String(req.TxHash, 0),
		BlockTimestamp: req.BlockTimestamp,
	}

	if err := b.dbDao.DeclareReverseRecord(reverseInfo, txInfo); err != nil {
		resp.Err = fmt.Errorf("DeclareReverseRecord err: %s", err.Error())
		return
	}

	return
}

func (b *BlockParser) ActionRedeclareReverseRecord(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameReverseRecordCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version redeclare reverse record tx")
		return
	}
	log.Info("ActionDeclareReverseRecord:", req.BlockNumber, req.TxHash)

	lastOutpoint := common.OutPointStruct2String(req.Tx.Inputs[0].PreviousOutput)
	account := string(req.Tx.OutputsData[0])

	ownerHex, _, err := b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[0].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}

	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(account))
	reverseInfo := dao.TableReverseInfo{
		BlockNumber:    req.BlockNumber,
		BlockTimestamp: req.BlockTimestamp,
		Outpoint:       common.OutPoint2String(req.TxHash, 0),
		AlgorithmId:    ownerHex.DasAlgorithmId,
		ChainType:      ownerHex.ChainType,
		Address:        ownerHex.AddressHex,
		AccountId:      accountId,
		Account:        account,
		Capacity:       req.Tx.Outputs[0].Capacity,
	}

	txInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      accountId,
		Account:        account,
		Action:         common.DasActionRedeclareReverseRecord,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      ownerHex.ChainType,
		Address:        ownerHex.AddressHex,
		Capacity:       req.Tx.Outputs[0].Capacity,
		Outpoint:       common.OutPoint2String(req.TxHash, 0),
		BlockTimestamp: req.BlockTimestamp,
	}

	if err := b.dbDao.RedeclareReverseRecord(lastOutpoint, reverseInfo, txInfo); err != nil {
		resp.Err = fmt.Errorf("RedeclareReverseRecord err: %s", err.Error())
		return
	}

	return
}

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

	ownerHex, _, err := b.dasCore.Daf().ArgsToHex(res.Transaction.Outputs[req.Tx.Inputs[0].PreviousOutput.Index].Lock.Args)
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
