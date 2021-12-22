package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
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
	oID, _, oCT, _, oA, _ := core.FormatDasLockToHexAddress(req.Tx.Outputs[0].Lock.Args)

	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(account))
	reverseInfo := dao.TableReverseInfo{
		BlockNumber:    req.BlockNumber,
		BlockTimestamp: req.BlockTimestamp,
		Outpoint:       common.OutPoint2String(req.TxHash, 0),
		AlgorithmId:    oID,
		ChainType:      oCT,
		Address:        oA,
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
		ChainType:      oCT,
		Address:        oA,
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
