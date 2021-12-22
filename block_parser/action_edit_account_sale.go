package block_parser

import (
	"das_database/dao"
	"das_database/timer"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/witness"
)

func (b *BlockParser) ActionEditAccountSale(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountSaleCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersionTx err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version edit account sale tx")
		return
	}

	log.Info("ActionEditAccountSale:", req.TxHash)

	builder, err := witness.AccountSaleCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountSaleCellDataBuilderFromTx err: %s", err.Error())
		return
	}

	account := builder.Account()
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(account))
	description := builder.Description()
	priceCkb, _ := builder.Price()
	startedAt, _ := builder.StartedAt()

	tokenInfo := timer.GetTokenPriceInfo(timer.TokenIdCkb)
	priceUsd := tokenInfo.GetPriceUsd(priceCkb)
	tradeInfo := dao.TableTradeInfo{
		BlockNumber:    req.BlockNumber,
		Outpoint:       common.OutPoint2String(req.TxHash, uint(builder.Index)),
		AccountId:      accountId,
		Account:        account,
		Description:    description,
		StartedAt:      startedAt,
		BlockTimestamp: req.BlockTimestamp,
		PriceCkb:       priceCkb,
		PriceUsd:       priceUsd,
		Status:         dao.AccountStatusOnSale,
	}
	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      accountId,
		Account:        account,
		Action:         common.DasActionEditAccountSale,
		ServiceType:    dao.ServiceTypeTransaction,
		ChainType:      tradeInfo.OwnerChainType,
		Address:        tradeInfo.OwnerAddress,
		Capacity:       0,
		Outpoint:       common.OutPoint2String(req.TxHash, uint(builder.Index)),
		BlockTimestamp: req.BlockTimestamp,
	}

	log.Info("ActionEditAccountSale:", transactionInfo.Account)

	if err := b.dbDao.EditAccountSale(tradeInfo, transactionInfo); err != nil {
		resp.Err = fmt.Errorf("EditAccountSale err: %s", err.Error())
		return
	}

	return
}
