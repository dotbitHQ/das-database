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

func (b *BlockParser) ActionMakeOffer(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DASContractNameOfferCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version make offer record tx")
		return
	}
	log.Info("ActionMakeOffer:", req.BlockNumber, req.TxHash)

	builder, err := witness.OfferCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("OfferCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	ownerHex, _, err := b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[builder.Index].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}

	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(builder.Account))
	offerInfo := dao.TableOfferInfo{
		BlockNumber:    req.BlockNumber,
		Outpoint:       common.OutPoint2String(req.TxHash, uint(builder.Index)),
		AccountId:      accountId,
		Account:        builder.Account,
		AlgorithmId:    ownerHex.DasAlgorithmId,
		ChainType:      ownerHex.ChainType,
		Address:        ownerHex.AddressHex,
		BlockTimestamp: req.BlockTimestamp,
		Price:          builder.Price,
		Message:        builder.Message,
		InviterArgs:    common.Bytes2Hex(builder.InviterLock.Args().RawData()),
		ChannelArgs:    common.Bytes2Hex(builder.ChannelLock.Args().RawData()),
	}
	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      accountId,
		Account:        builder.Account,
		Action:         common.DasActionMakeOffer,
		ServiceType:    dao.ServiceTypeTransaction,
		ChainType:      ownerHex.ChainType,
		Address:        ownerHex.AddressHex,
		Capacity:       req.Tx.Outputs[builder.Index].Capacity,
		Outpoint:       common.OutPoint2String(req.TxHash, uint(builder.Index)),
		BlockTimestamp: req.BlockTimestamp,
	}

	log.Info("ActionMakeOffer:", builder.Account)

	if err = b.dbDao.MakeOffer(offerInfo, transactionInfo); err != nil {
		resp.Err = fmt.Errorf("MakeOffer err: %s", err.Error())
		return
	}

	return
}

func (b *BlockParser) ActionEditOffer(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DASContractNameOfferCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version edit offer record tx")
		return
	}
	log.Info("ActionEditOffer:", req.BlockNumber, req.TxHash)

	oldBuilder, err := witness.OfferCellDataBuilderFromTx(req.Tx, common.DataTypeOld)
	if err != nil {
		resp.Err = fmt.Errorf("OfferCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	oldOutpoint := common.OutPoint2String(req.Tx.Inputs[oldBuilder.Index].PreviousOutput.TxHash.Hex(), uint(oldBuilder.Index))
	builder, err := witness.OfferCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("OfferCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	ownerHex, _, err := b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[builder.Index].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}

	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(builder.Account))
	offerInfo := dao.TableOfferInfo{
		BlockNumber:    req.BlockNumber,
		Outpoint:       common.OutPoint2String(req.TxHash, uint(builder.Index)),
		AccountId:      accountId,
		Account:        builder.Account,
		BlockTimestamp: req.BlockTimestamp,
		Price:          builder.Price,
		Message:        builder.Message,
	}
	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      accountId,
		Account:        builder.Account,
		ServiceType:    dao.ServiceTypeTransaction,
		ChainType:      ownerHex.ChainType,
		Address:        ownerHex.AddressHex,
		Outpoint:       common.OutPoint2String(req.TxHash, uint(builder.Index)),
		BlockTimestamp: req.BlockTimestamp,
	}
	if oldBuilder.Price <= builder.Price {
		transactionInfo.Action = dao.DasActionEditOfferAdd
		transactionInfo.Capacity = builder.Price - oldBuilder.Price
	} else {
		transactionInfo.Action = dao.DasActionEditOfferSub
		transactionInfo.Capacity = oldBuilder.Price - builder.Price
	}

	log.Info("ActionEditOffer:", builder.Account)

	if err = b.dbDao.EditOffer(oldOutpoint, offerInfo, transactionInfo); err != nil {
		resp.Err = fmt.Errorf("EditOffer err: %s", err.Error())
		return
	}

	return
}

func (b *BlockParser) ActionCancelOffer(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	res, err := b.dasCore.Client().GetTransaction(b.ctx, req.Tx.Inputs[0].PreviousOutput.TxHash)
	if err != nil {
		resp.Err = fmt.Errorf("GetTransaction err: %s", err.Error())
		return
	}

	if isCV, err := isCurrentVersionTx(res.Transaction, common.DASContractNameOfferCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version cancel offer record tx")
		return
	}
	log.Info("ActionCancelOffer:", req.BlockNumber, req.TxHash)

	oldBuilderMap, err := witness.OfferCellDataBuilderMapFromTx(req.Tx, common.DataTypeOld)
	if err != nil {
		resp.Err = fmt.Errorf("OfferCellDataBuilderMapFromTx err: %s", err.Error())
		return
	}
	var oldOutpoints []string
	for _, v := range oldBuilderMap {
		oldOutpoint := common.OutPoint2String(req.Tx.Inputs[v.Index].PreviousOutput.TxHash.Hex(), uint(v.Index))
		oldOutpoints = append(oldOutpoints, oldOutpoint)
	}
	var account string
	if len(oldBuilderMap) > 0 {
		account = oldBuilderMap[common.OutPoint2String(req.TxHash, 0)].Account
	}
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(account))
	ownerHex, _, err := b.dasCore.Daf().ArgsToHex(res.Transaction.Outputs[req.Tx.Inputs[0].PreviousOutput.Index].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}
	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      accountId,
		Account:        account,
		Action:         common.DasActionCancelOffer,
		ServiceType:    dao.ServiceTypeTransaction,
		ChainType:      ownerHex.ChainType,
		Address:        ownerHex.AddressHex,
		Capacity:       req.Tx.OutputsCapacity(),
		Outpoint:       common.OutPoint2String(req.TxHash, 0),
		BlockTimestamp: req.BlockTimestamp,
	}

	if err = b.dbDao.CancelOffer(oldOutpoints, transactionInfo); err != nil {
		resp.Err = fmt.Errorf("CancelOffer err: %s", err.Error())
		return
	}

	return
}

func (b *BlockParser) ActionAcceptOffer(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	res, err := b.dasCore.Client().GetTransaction(b.ctx, req.Tx.Inputs[0].PreviousOutput.TxHash)
	if err != nil {
		resp.Err = fmt.Errorf("GetTransaction err: %s", err.Error())
		return
	}
	if isCV, err := isCurrentVersionTx(res.Transaction, common.DASContractNameOfferCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version accept offer tx")
		return
	}
	log.Info("ActionAcceptOffer:", req.BlockNumber, req.TxHash)

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
				Action:         common.DasActionAcceptOffer,
				Outpoint:       common.OutPoint2String(req.TxHash, uint(i)),
				Capacity:       v.Capacity,
				BlockTimestamp: req.BlockTimestamp,
				Status:         dao.IncomeCellStatusUnMerge,
			})
		}
	}

	// res account cell
	resAccount, err := b.dasCore.Client().GetTransaction(b.ctx, req.Tx.Inputs[1].PreviousOutput.TxHash)
	if err != nil {
		resp.Err = fmt.Errorf("GetTransaction err: %s", err.Error())
		return
	}

	// offer cell
	offerBuilder, err := witness.OfferCellDataBuilderFromTx(req.Tx, common.DataTypeOld)
	if err != nil {
		resp.Err = fmt.Errorf("OfferCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	offerOutpoint := common.OutPointStruct2String(req.Tx.Inputs[offerBuilder.Index].PreviousOutput)

	// buyer account cell
	buyerBuilder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}

	// inviter channel
	offerPrice, _ := decimal.NewFromString(fmt.Sprintf("%d", offerBuilder.Price))
	rebateList, err := b.getOfferRebateInfoList(offerPrice, buyerBuilder.Account, &req)
	if err != nil {
		resp.Err = fmt.Errorf("getRebateInfoList err: %s", err.Error())
		return
	}

	ownerHex, managerHex, err := b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[buyerBuilder.Index].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}
	accountInfo := dao.TableAccountInfo{
		BlockNumber:        req.BlockNumber,
		Outpoint:           common.OutPoint2String(req.TxHash, uint(buyerBuilder.Index)),
		AccountId:          buyerBuilder.AccountId,
		Account:            buyerBuilder.Account,
		OwnerChainType:     ownerHex.ChainType,
		Owner:              ownerHex.AddressHex,
		OwnerAlgorithmId:   ownerHex.DasAlgorithmId,
		ManagerChainType:   managerHex.ChainType,
		Manager:            managerHex.AddressHex,
		ManagerAlgorithmId: managerHex.DasAlgorithmId,
		Status:             buyerBuilder.Status,
	}
	transactionInfoBuy := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      buyerBuilder.AccountId,
		Account:        buyerBuilder.Account,
		Action:         dao.DasActionOfferAccepted,
		ServiceType:    dao.ServiceTypeTransaction,
		ChainType:      ownerHex.ChainType,
		Address:        ownerHex.AddressHex,
		Capacity:       0,
		Outpoint:       common.OutPoint2String(req.TxHash, 0),
		BlockTimestamp: req.BlockTimestamp,
	}
	ownerHex, _, err = b.dasCore.Daf().ArgsToHex(resAccount.Transaction.Outputs[req.Tx.Inputs[1].PreviousOutput.Index].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}
	transactionInfoSale := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      buyerBuilder.AccountId,
		Account:        buyerBuilder.Account,
		Action:         common.DasActionAcceptOffer,
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
		Outpoint:       transactionInfoSale.Outpoint,
		AccountId:      buyerBuilder.AccountId,
		Account:        buyerBuilder.Account,
		DealType:       dao.DealTypeOffer,
		SellChainType:  transactionInfoSale.ChainType,
		SellAddress:    transactionInfoSale.Address,
		BuyChainType:   transactionInfoBuy.ChainType,
		BuyAddress:     transactionInfoBuy.Address,
		PriceCkb:       offerBuilder.Price,
		PriceUsd:       tokenInfo.GetPriceUsd(offerBuilder.Price),
		BlockTimestamp: req.BlockTimestamp,
	}
	var recordsInfos []dao.TableRecordsInfo
	recordList := buyerBuilder.Records
	for _, v := range recordList {
		recordsInfos = append(recordsInfos, dao.TableRecordsInfo{
			AccountId: buyerBuilder.AccountId,
			Account:   buyerBuilder.Account,
			Key:       v.Key,
			Type:      v.Type,
			Label:     v.Label,
			Value:     v.Value,
			Ttl:       strconv.FormatUint(uint64(v.TTL), 10),
		})
	}

	log.Info("ActionAcceptOffer:", buyerBuilder.AccountId, len(rebateList))

	if err = b.dbDao.AcceptOffer(incomeCellInfos, accountInfo, offerOutpoint, tradeDealInfo, transactionInfoBuy, transactionInfoSale, rebateList, recordsInfos); err != nil {
		log.Error("AcceptOffer err:", err.Error(), toolib.JsonString(transactionInfoBuy), toolib.JsonString(transactionInfoSale))
		resp.Err = fmt.Errorf("AcceptOffer err: %s", err.Error())
		return
	}

	return
}

func (b *BlockParser) getOfferRebateInfoList(salePrice decimal.Decimal, account string, req *FuncTransactionHandleReq) ([]dao.TableRebateInfo, error) {
	var list []dao.TableRebateInfo
	offerCellBuilder, err := witness.OfferCellDataBuilderFromTx(req.Tx, common.DataTypeOld)
	if err != nil {
		return list, fmt.Errorf("OfferCellDataBuilderMapFromTx err: %s", err.Error())
	}

	builder, err := b.dasCore.ConfigCellDataBuilderByTypeArgs(common.ConfigCellTypeArgsProfitRate)
	if err != nil {
		return list, fmt.Errorf("ConfigCellDataBuilderByTypeArgs err: %s", err.Error())
	}
	saleBuyerInviter, _ := builder.ProfitRateSaleBuyerInviter()
	saleBuyerChannel, _ := builder.ProfitRateSaleBuyerChannel()
	decInviter := decimal.NewFromInt(int64(saleBuyerInviter)).Div(decimal.NewFromInt(common.PercentRateBase))
	decChannel := decimal.NewFromInt(int64(saleBuyerChannel)).Div(decimal.NewFromInt(common.PercentRateBase))

	inviterScript := offerCellBuilder.InviterLock
	channelScript := offerCellBuilder.ChannelLock
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
		Action:           common.DasActionAcceptOffer,
		ServiceType:      dao.ServiceTypeTransaction,
		InviterArgs:      common.Bytes2Hex(inviterScript.Args().RawData()),
		InviterAccount:   "",
		InviterChainType: inviterHex.ChainType,
		InviterAddress:   inviterHex.AddressHex,
		BlockTimestamp:   req.BlockTimestamp,
	})
	channelHex, _, err := b.dasCore.Daf().ScriptToHex(molecule.MoleculeScript2CkbScript(channelScript))
	if err != nil {
		return list, fmt.Errorf("ScriptToHex err: %s", err.Error())
	}
	list = append(list, dao.TableRebateInfo{
		BlockNumber:      req.BlockNumber,
		Outpoint:         common.OutPoint2String(req.TxHash, 0),
		InviteeId:        inviteeId,
		InviteeAccount:   account,
		InviteeChainType: 0,
		InviteeAddress:   "",
		RewardType:       dao.RewardTypeChannel,
		Reward:           salePrice.Mul(decChannel).BigInt().Uint64(),
		Action:           common.DasActionAcceptOffer,
		ServiceType:      dao.ServiceTypeTransaction,
		InviterArgs:      common.Bytes2Hex(channelScript.Args().RawData()),
		InviterAccount:   "",
		InviterChainType: channelHex.ChainType,
		InviterAddress:   channelHex.AddressHex,
		BlockTimestamp:   req.BlockTimestamp,
	})
	return list, nil
}
