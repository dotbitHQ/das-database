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
	oID, mID, oCT, mCT, oA, mA := core.FormatDasLockToHexAddress(req.Tx.Outputs[accBuilder.Index].Lock.Args)

	accountInfo := dao.TableAccountInfo{
		BlockNumber:        req.BlockNumber,
		Outpoint:           common.OutPoint2String(req.TxHash, uint(accBuilder.Index)),
		Account:            accBuilder.Account,
		Status:             dao.AccountStatusOnSale,
		OwnerAlgorithmId:   oID,
		OwnerChainType:     oCT,
		Owner:              oA,
		ManagerAlgorithmId: mID,
		ManagerChainType:   mCT,
		Manager:            mA,
	}

	salePrice, _ := builder.Price()
	startedAt, _ := builder.StartedAt()
	tokenInfo := timer.GetTokenPriceInfo(timer.TokenIdCkb)
	salePriceUsd := tokenInfo.GetPriceUsd(salePrice)
	oID, _, oCT, _, oA, _ = core.FormatDasLockToHexAddress(req.Tx.Outputs[builder.Index].Lock.Args)
	tradeInfo := dao.TableTradeInfo{
		BlockNumber:      req.BlockNumber,
		Outpoint:         common.OutPoint2String(req.TxHash, uint(builder.Index)),
		Account:          builder.Account(),
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
