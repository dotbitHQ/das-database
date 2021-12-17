package block_parser

import (
	"das_database/dao"
	"das_database/timer"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/DeAccountSystems/das-lib/molecule"
	"github.com/DeAccountSystems/das-lib/witness"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"strconv"
	"strings"
)

func (b *BlockParser) ActionBuyAccount(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersionTx err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version buy account tx")
		return
	}

	log.Info("ActionBuyAccount:", req.TxHash)

	// add income cell infos
	incomeContract, err := core.GetDasContractInfo(common.DasContractNameIncomeCellType)
	if err != nil {
		resp.Err = fmt.Errorf("GetDasContractInfo err: %s", err.Error())
		return
	}
	var incomeCellInfos []dao.TableIncomeCellInfo
	for i, v := range req.Tx.Outputs {
		if v.Type == nil {
			continue
		}
		if v.Type.CodeHash.Hex() == incomeContract.ContractTypeId.Hex() {
			incomeCellInfos = append(incomeCellInfos, dao.TableIncomeCellInfo{
				BlockNumber:    req.BlockNumber,
				Action:         common.DasActionBuyAccount,
				Outpoint:       common.OutPoint2String(req.TxHash, uint(i)),
				Capacity:       v.Capacity,
				BlockTimestamp: req.BlockTimestamp,
				Status:         dao.IncomeCellStatusUnMerge,
			})
		}
	}

	// sale cell
	res, err := b.ckbClient.GetTxByHashOnChain(req.Tx.Inputs[1].PreviousOutput.TxHash)
	if err != nil {
		resp.Err = fmt.Errorf("GetTxByHashOnChain err: %s", err.Error())
		return
	}
	saleBuilder, err := witness.AccountSaleCellDataBuilderFromTx(res.Transaction, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountSaleCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	priceBuyCkb, err := saleBuilder.Price()
	if err != nil {
		resp.Err = fmt.Errorf("get price err: %s", err.Error())
		return
	}
	// account cell
	accBuilder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	account := accBuilder.Account
	// inviter channel
	salePrice, _ := decimal.NewFromString(fmt.Sprintf("%d", priceBuyCkb))
	rebateList, err := b.getRebateInfoList(salePrice, account, &req)
	if err != nil {
		resp.Err = fmt.Errorf("getRebateInfoList err: %s", err.Error())
		return
	}

	oID, _, oCT, _, oA, _ := core.FormatDasLockToHexAddress(req.Tx.Outputs[0].Lock.Args)
	accountInfo := dao.TableAccountInfo{
		BlockNumber:        req.BlockNumber,
		Outpoint:           common.OutPoint2String(req.TxHash, uint(accBuilder.Index)),
		Account:            account,
		OwnerChainType:     oCT,
		Owner:              oA,
		OwnerAlgorithmId:   oID,
		ManagerChainType:   oCT,
		Manager:            oA,
		ManagerAlgorithmId: oID,
		Status:             dao.AccountStatusNormal,
	}
	transactionInfoBuy := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		Account:        account,
		Action:         common.DasActionBuyAccount,
		ServiceType:    dao.ServiceTypeTransaction,
		ChainType:      oCT,
		Address:        oA,
		Capacity:       priceBuyCkb,
		Outpoint:       common.OutPoint2String(req.TxHash, 0),
		BlockTimestamp: req.BlockTimestamp,
	}
	_, _, oCT, _, oA, _ = core.FormatDasLockToHexAddress(res.Transaction.Outputs[req.Tx.Inputs[1].PreviousOutput.Index].Lock.Args)
	transactionInfoSale := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		Account:        account,
		Action:         dao.DasActionSaleAccount,
		ServiceType:    dao.ServiceTypeTransaction,
		ChainType:      oCT,
		Address:        oA,
		Outpoint:       common.OutPoint2String(req.TxHash, 1),
		BlockTimestamp: req.BlockTimestamp,
	}
	for i := 1; i < len(req.Tx.Outputs); i++ {
		_, _, oCT, _, oA, _ = core.FormatDasLockToHexAddress(req.Tx.Outputs[i].Lock.Args)
		if transactionInfoSale.ChainType == oCT && strings.EqualFold(transactionInfoSale.Address, oA) {
			transactionInfoSale.Capacity = req.Tx.Outputs[i].Capacity
			break
		}
	}
	tokenInfo := timer.GetTokenPriceInfo(timer.TokenIdCkb)
	tradeDealInfo := dao.TableTradeDealInfo{
		BlockNumber:    req.BlockNumber,
		Outpoint:       transactionInfoBuy.Outpoint,
		Account:        account,
		DealType:       dao.DealTypeSale,
		SellChainType:  transactionInfoSale.ChainType,
		SellAddress:    transactionInfoSale.Address,
		BuyChainType:   transactionInfoBuy.ChainType,
		BuyAddress:     transactionInfoBuy.Address,
		PriceCkb:       transactionInfoBuy.Capacity,
		PriceUsd:       tokenInfo.GetPriceUsd(transactionInfoBuy.Capacity),
		BlockTimestamp: req.BlockTimestamp,
	}
	var recordsInfos []dao.TableRecordsInfo
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

	log.Info("ActionBuyAccount:", account, len(rebateList))

	if err := b.dbDao.BuyAccount(incomeCellInfos, accountInfo, tradeDealInfo, transactionInfoBuy, transactionInfoSale, rebateList, recordsInfos); err != nil {
		log.Error("BuyAccount err:", err.Error(), toolib.JsonString(transactionInfoBuy), toolib.JsonString(transactionInfoSale))
		resp.Err = fmt.Errorf("BuyAccount err: %s", err.Error())
		return
	}

	return
}

func (b *BlockParser) getRebateInfoList(salePrice decimal.Decimal, account string, req *FuncTransactionHandleReq) ([]dao.TableRebateInfo, error) {
	var list []dao.TableRebateInfo
	actionDataBuilder, err := witness.ActionDataBuilderFromTx(req.Tx)
	if err != nil {
		return list, fmt.Errorf("ActionDataBuilderFromTx err: %s", err.Error())
	}

	builder, err := witness.ConfigCellDataBuilderByTypeArgs(req.Tx, common.ConfigCellTypeArgsProfitRate)
	if err != nil {
		return list, fmt.Errorf("ConfigCellDataBuilderByTypeArgs err: %s", err.Error())
	}
	saleBuyerInviter, _ := builder.ProfitRateSaleBuyerInviter()
	saleBuyerChannel, _ := builder.ProfitRateSaleBuyerChannel()
	decInviter := decimal.NewFromInt(int64(saleBuyerInviter)).Div(decimal.NewFromInt(common.PercentRateBase))
	decChannel := decimal.NewFromInt(int64(saleBuyerChannel)).Div(decimal.NewFromInt(common.PercentRateBase))

	inviterScript, _ := actionDataBuilder.ActionBuyAccountInviterScript()
	channelScript, _ := actionDataBuilder.ActionBuyAccountChannelScript()
	if inviterScript == nil {
		tmp := molecule.ScriptDefault()
		inviterScript = &tmp
	}
	if channelScript == nil {
		tmp := molecule.ScriptDefault()
		channelScript = &tmp
	}
	_, _, oCT, _, oA, _ := core.FormatDasLockToHexAddress(inviterScript.Args().RawData())
	if oA == "" {
		oCT = common.ChainTypeCkb
		oA = common.Bytes2Hex(inviterScript.Args().RawData())
	}
	list = append(list, dao.TableRebateInfo{
		BlockNumber:      req.BlockNumber,
		Outpoint:         common.OutPoint2String(req.TxHash, 0),
		InviteeAccount:   account,
		InviteeChainType: 0,
		InviteeAddress:   "",
		RewardType:       dao.RewardTypeInviter,
		Reward:           salePrice.Mul(decInviter).BigInt().Uint64(),
		Action:           common.DasActionBuyAccount,
		ServiceType:      dao.ServiceTypeTransaction,
		InviterArgs:      common.Bytes2Hex(inviterScript.Args().RawData()),
		InviterAccount:   "",
		InviterChainType: oCT,
		InviterAddress:   oA,
		BlockTimestamp:   req.BlockTimestamp,
	})
	_, _, oCT, _, oA, _ = core.FormatDasLockToHexAddress(channelScript.Args().RawData())
	if oA == "" {
		oCT = common.ChainTypeCkb
		oA = common.Bytes2Hex(channelScript.Args().RawData())
	}
	list = append(list, dao.TableRebateInfo{
		BlockNumber:      req.BlockNumber,
		Outpoint:         common.OutPoint2String(req.TxHash, 0),
		InviteeAccount:   account,
		InviteeChainType: 0,
		InviteeAddress:   "",
		RewardType:       dao.RewardTypeChannel,
		Reward:           salePrice.Mul(decChannel).BigInt().Uint64(),
		Action:           common.DasActionBuyAccount,
		ServiceType:      dao.ServiceTypeTransaction,
		InviterArgs:      common.Bytes2Hex(channelScript.Args().RawData()),
		InviterAccount:   "",
		InviterChainType: oCT,
		InviterAddress:   oA,
		BlockTimestamp:   req.BlockTimestamp,
	})
	return list, nil
}
