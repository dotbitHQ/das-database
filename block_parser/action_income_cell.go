package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
)

func (b *BlockParser) ActionCreateIncome(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameIncomeCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if isCV {
		log.Warn("not current version create income tx")
		return
	}
	log.Info("ActionCreateIncome:", req.BlockNumber, req.TxHash)

	return
}

func (b *BlockParser) ActionConsolidateIncome(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	incomeContract, err := core.GetDasContractInfo(common.DasContractNameIncomeCellType)
	if err != nil {
		resp.Err = fmt.Errorf("GetDasContractInfo err: %s", err.Error())
		return
	}
	dasContract, err := core.GetDasContractInfo(common.DasContractNameDispatchCellType)
	if err != nil {
		resp.Err = fmt.Errorf("GetDasContractInfo err: %s", err.Error())
		return
	}
	log.Info("ActionConsolidateIncome:", req.TxHash)

	var inputsOutpoints []string
	var incomeCellInfos []dao.TableIncomeCellInfo
	var transactionInfos []dao.TableTransactionInfo

	for _, v := range req.Tx.Inputs {
		inputsOutpoints = append(inputsOutpoints, common.OutPoint2String(v.PreviousOutput.TxHash.Hex(), v.PreviousOutput.Index))
	}

	for i, v := range req.Tx.Outputs {
		if dasContract.IsSameTypeId(v.Lock.CodeHash) {
			ownerHex, _, err := b.dasCore.Daf().ArgsToHex(v.Lock.Args)
			if err != nil {
				resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
				return
			}
			transactionInfos = append(transactionInfos, dao.TableTransactionInfo{
				BlockNumber:    req.BlockNumber,
				Action:         common.DasActionConsolidateIncome,
				ServiceType:    dao.ServiceTypeTransaction,
				ChainType:      ownerHex.ChainType,
				Address:        ownerHex.AddressHex,
				Capacity:       v.Capacity,
				Outpoint:       common.OutPoint2String(req.TxHash, uint(i)),
				BlockTimestamp: req.BlockTimestamp,
			})
		} else if v.Type != nil && incomeContract.IsSameTypeId(v.Type.CodeHash) {
			incomeCellInfos = append(incomeCellInfos, dao.TableIncomeCellInfo{
				BlockNumber:    req.BlockNumber,
				Action:         common.DasActionConsolidateIncome,
				Outpoint:       common.OutPoint2String(req.TxHash, uint(i)),
				Capacity:       v.Capacity,
				BlockTimestamp: req.BlockTimestamp,
				Status:         dao.IncomeCellStatusUnMerge,
			})
		}
	}

	if err = b.dbDao.ConsolidateIncome(inputsOutpoints, incomeCellInfos, transactionInfos); err != nil {
		log.Error("ConsolidateIncome err: ", err.Error())
		resp.Err = fmt.Errorf("ConsolidateIncome err: %s", err.Error())
		return
	}

	return
}
