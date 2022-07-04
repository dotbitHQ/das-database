package block_parser

import (
	"das_database/dao"
	"das_database/timer"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/molecule"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"strconv"
	"strings"
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
	ownerHex, managerHex, err := b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[accBuilder.Index].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}
	accountInfo := dao.TableAccountInfo{
		BlockNumber:        req.BlockNumber,
		Outpoint:           common.OutPoint2String(req.TxHash, uint(accBuilder.Index)),
		AccountId:          accBuilder.AccountId,
		Account:            accBuilder.Account,
		Status:             dao.AccountStatusOnSale,
		OwnerAlgorithmId:   ownerHex.DasAlgorithmId,
		OwnerChainType:     ownerHex.ChainType,
		Owner:              ownerHex.AddressHex,
		ManagerAlgorithmId: managerHex.DasAlgorithmId,
		ManagerChainType:   managerHex.ChainType,
		Manager:            managerHex.AddressHex,
	}
	tokenInfo := timer.GetTokenPriceInfo(timer.TokenIdCkb)
	priceUsd := tokenInfo.GetPriceUsd(builder.Price)

	ownerHex, _, err = b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[builder.Index].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}

	tradeInfo := dao.TableTradeInfo{
		BlockNumber:      req.BlockNumber,
		Outpoint:         common.OutPoint2String(req.TxHash, uint(builder.Index)),
		AccountId:        accountInfo.AccountId,
		Account:          accountInfo.Account,
		OwnerAlgorithmId: ownerHex.DasAlgorithmId,
		OwnerChainType:   ownerHex.ChainType,
		OwnerAddress:     ownerHex.AddressHex,
		Description:      builder.Description,
		StartedAt:        builder.StartedAt * 1e3,
		PriceCkb:         builder.Price,
		PriceUsd:         priceUsd,
		ProfitRate:       builder.BuyerInviterProfitRate,
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

	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(builder.Account))
	tokenInfo := timer.GetTokenPriceInfo(timer.TokenIdCkb)
	priceUsd := tokenInfo.GetPriceUsd(builder.Price)

	oHex, _, err := b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[0].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}

	tradeInfo := dao.TableTradeInfo{
		BlockNumber:    req.BlockNumber,
		Outpoint:       common.OutPoint2String(req.TxHash, uint(builder.Index)),
		AccountId:      accountId,
		Account:        builder.Account,
		Description:    builder.Description,
		StartedAt:      builder.StartedAt,
		BlockTimestamp: req.BlockTimestamp,
		PriceCkb:       builder.Price,
		PriceUsd:       priceUsd,
		ProfitRate:     builder.BuyerInviterProfitRate,
		Status:         dao.AccountStatusOnSale,
	}
	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      accountId,
		Account:        tradeInfo.Account,
		Action:         common.DasActionEditAccountSale,
		ServiceType:    dao.ServiceTypeTransaction,
		ChainType:      oHex.ChainType,
		Address:        oHex.AddressHex,
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

func (b *BlockParser) ActionCancelAccountSale(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersionTx err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version cancel account sale tx")
		return
	}
	log.Info("ActionCancelAccountSale:", req.TxHash)

	builder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}

	accountInfo := dao.TableAccountInfo{
		BlockNumber: req.BlockNumber,
		Outpoint:    common.OutPoint2String(req.TxHash, uint(builder.Index)),
		AccountId:   builder.AccountId,
		Account:     builder.Account,
		Status:      dao.AccountStatusNormal,
	}

	ownerHex, _, err := b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[0].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}
	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      builder.AccountId,
		Account:        builder.Account,
		Action:         common.DasActionCancelAccountSale,
		ServiceType:    dao.ServiceTypeTransaction,
		ChainType:      ownerHex.ChainType,
		Address:        ownerHex.AddressHex,
		Capacity:       req.Tx.Outputs[1].Capacity,
		Outpoint:       common.OutPoint2String(req.TxHash, 1),
		BlockTimestamp: req.BlockTimestamp,
	}

	log.Info("ActionCancelAccountSale:", transactionInfo.Account)

	if err := b.dbDao.CancelAccountSale(accountInfo, transactionInfo); err != nil {
		resp.Err = fmt.Errorf("CancelAccountSale err: %s", err.Error())
		return
	}

	return
}

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
		if incomeContract.IsSameTypeId(v.Type.CodeHash) {
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
	res, err := b.dasCore.Client().GetTransaction(b.ctx, req.Tx.Inputs[1].PreviousOutput.TxHash)
	if err != nil {
		resp.Err = fmt.Errorf("GetTransaction err: %s", err.Error())
		return
	}
	builder, err := witness.AccountSaleCellDataBuilderFromTx(res.Transaction, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountSaleCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	// account cell
	accBuilder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	account := accBuilder.Account
	accountId := accBuilder.AccountId
	// inviter channel
	salePrice, _ := decimal.NewFromString(fmt.Sprintf("%d", builder.Price))
	rebateList, err := b.getRebateInfoList(salePrice, account, builder.BuyerInviterProfitRate, &req)
	if err != nil {
		resp.Err = fmt.Errorf("getRebateInfoList err: %s", err.Error())
		return
	}

	ownerHex, managerHex, err := b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[0].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}
	accountInfo := dao.TableAccountInfo{
		BlockNumber:        req.BlockNumber,
		Outpoint:           common.OutPoint2String(req.TxHash, uint(accBuilder.Index)),
		AccountId:          accountId,
		Account:            account,
		OwnerChainType:     ownerHex.ChainType,
		Owner:              ownerHex.AddressHex,
		OwnerAlgorithmId:   ownerHex.DasAlgorithmId,
		ManagerChainType:   managerHex.ChainType,
		Manager:            managerHex.AddressHex,
		ManagerAlgorithmId: managerHex.DasAlgorithmId,
		Status:             dao.AccountStatusNormal,
	}
	transactionInfoBuy := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      accountId,
		Account:        account,
		Action:         common.DasActionBuyAccount,
		ServiceType:    dao.ServiceTypeTransaction,
		ChainType:      ownerHex.ChainType,
		Address:        ownerHex.AddressHex,
		Capacity:       builder.Price,
		Outpoint:       common.OutPoint2String(req.TxHash, 0),
		BlockTimestamp: req.BlockTimestamp,
	}
	ownerHex, _, err = b.dasCore.Daf().ArgsToHex(res.Transaction.Outputs[req.Tx.Inputs[1].PreviousOutput.Index].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}
	transactionInfoSale := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      accountId,
		Account:        account,
		Action:         dao.DasActionSaleAccount,
		ServiceType:    dao.ServiceTypeTransaction,
		ChainType:      ownerHex.ChainType,
		Address:        ownerHex.AddressHex,
		Outpoint:       common.OutPoint2String(req.TxHash, 1),
		BlockTimestamp: req.BlockTimestamp,
	}
	for i := 1; i < len(req.Tx.Outputs); i++ {
		ownerHex, _, err = b.dasCore.Daf().ScriptToHex(req.Tx.Outputs[i].Lock)
		if err != nil {
			resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
			return
		}
		if transactionInfoSale.ChainType == ownerHex.ChainType && strings.EqualFold(transactionInfoSale.Address, ownerHex.AddressHex) {
			transactionInfoSale.Capacity = req.Tx.Outputs[i].Capacity
			break
		}
	}
	tokenInfo := timer.GetTokenPriceInfo(timer.TokenIdCkb)
	tradeDealInfo := dao.TableTradeDealInfo{
		BlockNumber:    req.BlockNumber,
		Outpoint:       transactionInfoBuy.Outpoint,
		AccountId:      accountId,
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
	recordList := accBuilder.Records
	for _, v := range recordList {
		recordsInfos = append(recordsInfos, dao.TableRecordsInfo{
			AccountId: accountId,
			Account:   account,
			Key:       v.Key,
			Type:      v.Type,
			Label:     v.Label,
			Value:     v.Value,
			Ttl:       strconv.FormatUint(uint64(v.TTL), 10),
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

func (b *BlockParser) getRebateInfoList(salePrice decimal.Decimal, account string, profitRate uint32, req *FuncTransactionHandleReq) ([]dao.TableRebateInfo, error) {
	var list []dao.TableRebateInfo
	actionDataBuilder, err := witness.ActionDataBuilderFromTx(req.Tx)
	if err != nil {
		return list, fmt.Errorf("ActionDataBuilderFromTx err: %s", err.Error())
	}

	builder, err := b.dasCore.ConfigCellDataBuilderByTypeArgs(common.ConfigCellTypeArgsProfitRate)
	if err != nil {
		return list, fmt.Errorf("ConfigCellDataBuilderByTypeArgs err: %s", err.Error())
	}
	saleBuyerChannel, _ := builder.ProfitRateSaleBuyerChannel()
	decInviter := decimal.NewFromInt(int64(profitRate)).Div(decimal.NewFromInt(common.PercentRateBase))
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
	inviterHex, _, err := b.dasCore.Daf().ScriptToHex(molecule.MoleculeScript2CkbScript(inviterScript))
	if err != nil {
		return list, fmt.Errorf("ScriptToHex err: %s", err.Error())
	}
	channelHex, _, err := b.dasCore.Daf().ScriptToHex(molecule.MoleculeScript2CkbScript(channelScript))
	if err != nil {
		return list, fmt.Errorf("ScriptToHex err: %s", err.Error())
	}
	inviteeId := common.Bytes2Hex(common.GetAccountIdByAccount(account))
	list = append(list, dao.TableRebateInfo{
		BlockNumber:      req.BlockNumber,
		Outpoint:         common.OutPoint2String(req.TxHash, 0),
		InviteeId:        inviteeId,
		InviteeAccount:   account,
		InviteeChainType: 0,
		InviteeAddress:   "",
		RewardType:       dao.RewardTypeInviter,
		Reward:           salePrice.Mul(decInviter).BigInt().Uint64(),
		Action:           common.DasActionBuyAccount,
		ServiceType:      dao.ServiceTypeTransaction,
		InviterArgs:      common.Bytes2Hex(inviterScript.Args().RawData()),
		InviterAccount:   "",
		InviterChainType: inviterHex.ChainType,
		InviterAddress:   inviterHex.AddressHex,
		BlockTimestamp:   req.BlockTimestamp,
	})

	list = append(list, dao.TableRebateInfo{
		BlockNumber:      req.BlockNumber,
		Outpoint:         common.OutPoint2String(req.TxHash, 0),
		InviteeId:        inviteeId,
		InviteeAccount:   account,
		InviteeChainType: 0,
		InviteeAddress:   "",
		RewardType:       dao.RewardTypeChannel,
		Reward:           salePrice.Mul(decChannel).BigInt().Uint64(),
		Action:           common.DasActionBuyAccount,
		ServiceType:      dao.ServiceTypeTransaction,
		InviterArgs:      common.Bytes2Hex(channelScript.Args().RawData()),
		InviterAccount:   "",
		InviterChainType: channelHex.ChainType,
		InviterAddress:   channelHex.AddressHex,
		BlockTimestamp:   req.BlockTimestamp,
	})
	return list, nil
}
