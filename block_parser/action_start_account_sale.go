package block_parser

import (
	"das_database/dao"
	"das_database/timer"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/DeAccountSystems/das-lib/witness"
)

func (b *BlockParser) ActionStartAccountSale(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountSaleCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersionTx err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version start account sale tx")
		return
	}

	log.Info("ActionStartAccountSale:", req.TxHash)

	builder, err := witness.AccountSaleCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountSaleCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	accBuilder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	accountInfo := dao.TableAccountInfo{
		BlockNumber: req.BlockNumber,
		Outpoint:    common.OutPoint2String(req.TxHash, uint(accBuilder.Index)),
		AccountId:   accBuilder.AccountId,
		Account:     accBuilder.Account,
		Status:      dao.AccountStatusOnSale,
	}

	salePrice, _ := builder.Price()
	startedAt, _ := builder.StartedAt()
	tokenInfo := timer.GetTokenPriceInfo(timer.TokenIdCkb)
	salePriceUsd := tokenInfo.GetPriceUsd(salePrice)
	account := builder.Account()
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(account))
	oID, _, oCT, _, oA, _ := core.FormatDasLockToHexAddress(req.Tx.Outputs[builder.Index].Lock.Args)
	tradeInfo := dao.TableTradeInfo{
		BlockNumber:      req.BlockNumber,
		Outpoint:         common.OutPoint2String(req.TxHash, uint(builder.Index)),
		AccountId:        accountId,
		Account:          account,
		OwnerAlgorithmId: oID,
		OwnerChainType:   oCT,
		OwnerAddress:     oA,
		Description:      builder.Description(),
		StartedAt:        startedAt * 1e3,
		PriceCkb:         salePrice,
		PriceUsd:         salePriceUsd,
		BlockTimestamp:   req.BlockTimestamp,
		Status:           dao.AccountStatusOnSale,
	}
	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      tradeInfo.AccountId,
		Account:        tradeInfo.Account,
		Action:         common.DasActionStartAccountSale,
		ServiceType:    dao.ServiceTypeTransaction,
		ChainType:      tradeInfo.OwnerChainType,
		Address:        tradeInfo.OwnerAddress,
		Capacity:       req.Tx.Outputs[builder.Index].Capacity,
		Outpoint:       common.OutPoint2String(req.TxHash, uint(builder.Index)),
		BlockTimestamp: req.BlockTimestamp,
	}

	log.Info("ActionStartAccountSale:", transactionInfo.Account)

	if err = b.dbDao.StartAccountSale(accountInfo, tradeInfo, transactionInfo); err != nil {
		resp.Err = fmt.Errorf("StartAccountSale err: %s", err.Error())
		return
	}

	return
}
