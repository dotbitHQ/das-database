package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/witness"
)

func (b *BlockParser) ActionPropose(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameProposalCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version proposal tx")
		return
	}

	log.Info("ActionPropose:", req.BlockNumber, req.TxHash)

	preAccMap, err := witness.PreAccountCellDataBuilderMapFromTx(req.Tx, common.DataTypeDep)
	if err != nil {
		resp.Err = fmt.Errorf("PreAccountCellDataBuilderMapFromTx err: %s", err.Error())
		return
	}

	log.Info("ActionPropose:", len(preAccMap))

	proBuilder, err := witness.ProposalCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("ProposalCellDataBuilderFromTx err: %s", err.Error())
		return
	}

	var transactionInfos []dao.TableTransactionInfo
	for _, v := range preAccMap {
		transactionInfos = append(transactionInfos, dao.TableTransactionInfo{
			BlockNumber:    req.BlockNumber,
			Account:        v.Account,
			Action:         common.DasActionPropose,
			ServiceType:    dao.ServiceTypeRegister,
			ChainType:      common.ChainTypeCkb,
			Address:        common.Bytes2Hex(proBuilder.ProposalCellData.ProposerLock().Args().RawData()),
			Capacity:       req.Tx.Outputs[0].Capacity,
			Outpoint:       common.OutPoint2String(req.TxHash, uint(v.Index)),
			BlockTimestamp: req.BlockTimestamp,
		})
	}

	if err = b.dbDao.CreateTransactionInfoList(transactionInfos); err != nil {
		log.Error("CreateTransactionInfoList err:", err.Error(), req.TxHash, req.BlockNumber)
		resp.Err = fmt.Errorf("CreateTransactionInfoList err: %s ", err.Error())
		return
	}

	return
}
