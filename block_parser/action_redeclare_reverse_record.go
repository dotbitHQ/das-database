package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
)

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
