package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/DeAccountSystems/das-lib/witness"
	"github.com/scorpiotzh/toolib"
)

func (b *BlockParser) ActionRenewAccount(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version renew account tx")
		return
	}

	log.Info("ActionRenewAccount:", req.BlockNumber, req.TxHash)

	incomeContract, err := core.GetDasContractInfo(common.DasContractNameIncomeCellType)
	if err != nil {
		resp.Err = fmt.Errorf("GetDasContractInfo err: %s", err.Error())
		return
	}

	var inputsOutpoints []string
	var incomeCellInfos []dao.TableIncomeCellInfo
	for _, v := range req.Tx.Inputs {
		inputsOutpoints = append(inputsOutpoints, common.OutPoint2String(v.PreviousOutput.TxHash.Hex(), v.PreviousOutput.Index))
	}
	renewCapacity := uint64(0)
	for i, v := range req.Tx.Outputs {
		if v.Type == nil {
			continue
		}
		if incomeContract.IsSameTypeId(v.Type.CodeHash) {
			renewCapacity = v.Capacity
			incomeCellInfos = append(incomeCellInfos, dao.TableIncomeCellInfo{
				BlockNumber:    req.BlockNumber,
				Action:         common.DasActionRenewAccount,
				Outpoint:       common.OutPoint2String(req.TxHash, uint(i)),
				Capacity:       v.Capacity,
				BlockTimestamp: req.BlockTimestamp,
				Status:         dao.IncomeCellStatusUnMerge,
			})
		}
	}

	builder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}

	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(builder.Account))
	accountInfo := dao.TableAccountInfo{
		BlockNumber: req.BlockNumber,
		Outpoint:    common.OutPoint2String(req.TxHash, uint(builder.Index)),
		AccountId:   accountId,
		Account:     builder.Account,
		ExpiredAt:   builder.ExpiredAt,
	}
	_, _, oChainType, _, oAddress, _ := core.FormatDasLockToHexAddress(req.Tx.Outputs[builder.Index].Lock.Args)
	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      accountId,
		Account:        builder.Account,
		Action:         common.DasActionRenewAccount,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      oChainType,
		Address:        oAddress,
		Capacity:       renewCapacity,
		Outpoint:       common.OutPoint2String(req.TxHash, uint(builder.Index)),
		BlockTimestamp: req.BlockTimestamp,
	}

	log.Info("ActionRenewAccount:", builder.Account, builder.ExpiredAt, transactionInfo.Capacity)

	if err := b.dbDao.RenewAccount(inputsOutpoints, incomeCellInfos, accountInfo, transactionInfo); err != nil {
		log.Error("RenewAccount err:", err.Error(), toolib.JsonString(transactionInfo))
		resp.Err = fmt.Errorf("RenewAccount err: %s", err.Error())
	}

	return
}
