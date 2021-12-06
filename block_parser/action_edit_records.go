package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/DeAccountSystems/das-lib/witness"
	"github.com/scorpiotzh/toolib"
	"strconv"
)

func (b *BlockParser) ActionEditRecords(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version edit records tx")
		return
	}

	log.Info("ActionEditRecords:", req.BlockNumber, req.TxHash)

	accBuilder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	var recordsInfos []dao.TableRecordsInfo
	account := accBuilder.Account
	recordList := accBuilder.RecordList()
	for _, v := range recordList {
		recordsInfos = append(recordsInfos, dao.TableRecordsInfo{
			Account: account,
			Key:     v.Key,
			Type:    v.Type,
			Label:   v.Label,
			Value:   v.Value,
			Ttl:     strconv.FormatUint(uint64(v.TTL), 10),
		})
	}
	accountInfo := dao.TableAccountInfo{
		BlockNumber: req.BlockNumber,
		Outpoint:    common.OutPoint2String(req.TxHash, uint(accBuilder.Index)),
		Account:     account,
	}
	_, _, _, mCT, _, mA := core.FormatDasLockToHexAddress(req.Tx.Outputs[accBuilder.Index].Lock.Args)
	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		Account:        account,
		Action:         common.DasActionEditRecords,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      mCT,
		Address:        mA,
		Capacity:       0,
		Outpoint:       common.OutPoint2String(req.TxHash, uint(accBuilder.Index)),
		BlockTimestamp: req.BlockTimestamp,
	}

	log.Info("ActionEditRecords:", account, transactionInfo.Address)

	if err := b.dbDao.CreateRecordsInfos(accountInfo, recordsInfos, transactionInfo); err != nil {
		log.Error("CreateRecordsInfos err:", err.Error(), toolib.JsonString(transactionInfo))
		resp.Err = fmt.Errorf("CreateRecordsInfos err: %s", err.Error())
	}

	return
}
