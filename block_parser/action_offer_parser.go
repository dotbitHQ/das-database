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
	oID, _, oCT, _, oA, _ := core.FormatDasLockToHexAddress(req.Tx.Outputs[builder.Index].Lock.Args)

	offerInfo := dao.TableOfferInfo{
		BlockNumber:    req.BlockNumber,
		Outpoint:       common.OutPoint2String(req.TxHash, uint(builder.Index)),
		Account:        builder.Account,
		AlgorithmId:    oID,
		ChainType:      oCT,
		Address:        oA,
		BlockTimestamp: req.BlockTimestamp,
		Price:          builder.Price,
		Message:        builder.Message,
		InviterArgs:    common.Bytes2Hex(builder.InviterLock.Args().RawData()),
		ChannelArgs:    common.Bytes2Hex(builder.ChannelLock.Args().RawData()),
	}
	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		Account:        builder.Account,
		Action:         common.DasActionMakeOffer,
		ServiceType:    dao.ServiceTypeTransaction,
		ChainType:      oCT,
		Address:        oA,
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
	_, _, oCT, _, oA, _ := core.FormatDasLockToHexAddress(req.Tx.Outputs[builder.Index].Lock.Args)

	offerInfo := dao.TableOfferInfo{
		BlockNumber:    req.BlockNumber,
		Outpoint:       common.OutPoint2String(req.TxHash, uint(builder.Index)),
		Account:        builder.Account,
		BlockTimestamp: req.BlockTimestamp,
		Price:          builder.Price,
		Message:        builder.Message,
	}
	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		Account:        builder.Account,
		ServiceType:    dao.ServiceTypeTransaction,
		ChainType:      oCT,
		Address:        oA,
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
	res, err := b.ckbClient.GetTxByHashOnChain(req.Tx.Inputs[0].PreviousOutput.TxHash)
	if err != nil {
		resp.Err = fmt.Errorf("GetTxByHashOnChain err: %s", err.Error())
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
	_, _, oCT, _, oA, _ := core.FormatDasLockToHexAddress(res.Transaction.Outputs[req.Tx.Inputs[0].PreviousOutput.Index].Lock.Args)
	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		Account:        account,
		Action:         common.DasActionCancelOffer,
		ServiceType:    dao.ServiceTypeTransaction,
		ChainType:      oCT,
		Address:        oA,
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
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
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
		if v.Type.CodeHash.Hex() == incomeContract.ContractTypeId.Hex() {
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

	// seller account cell
	sellerBuilder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeOld)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	res, err := b.ckbClient.GetTxByHashOnChain(req.Tx.Inputs[sellerBuilder.Index].PreviousOutput.TxHash)
	if err != nil {
		resp.Err = fmt.Errorf("GetTxByHashOnChain err: %s", err.Error())
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

	oID, _, oCT, _, oA, _ := core.FormatDasLockToHexAddress(req.Tx.Outputs[buyerBuilder.Index].Lock.Args)
	accountInfo := dao.TableAccountInfo{
		BlockNumber:        req.BlockNumber,
		Outpoint:           common.OutPoint2String(req.TxHash, uint(buyerBuilder.Index)),
		Account:            buyerBuilder.Account,
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
		Account:        buyerBuilder.Account,
		Action:         dao.DasActionOfferAccepted,
		ServiceType:    dao.ServiceTypeTransaction,
		ChainType:      oCT,
		Address:        oA,
		Capacity:       0,
		Outpoint:       common.OutPoint2String(req.TxHash, 0),
		BlockTimestamp: req.BlockTimestamp,
	}
	_, _, oCT, _, oA, _ = core.FormatDasLockToHexAddress(res.Transaction.Outputs[req.Tx.Inputs[sellerBuilder.Index].PreviousOutput.Index].Lock.Args)
	transactionInfoSale := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		Account:        buyerBuilder.Account,
		Action:         common.DasActionAcceptOffer,
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
		Outpoint:       transactionInfoSale.Outpoint,
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
	recordList := buyerBuilder.RecordList()
	for _, v := range recordList {
		recordsInfos = append(recordsInfos, dao.TableRecordsInfo{
			Account: buyerBuilder.Account,
			Key:     v.Key,
			Type:    v.Type,
			Label:   v.Label,
			Value:   v.Value,
			Ttl:     strconv.FormatUint(uint64(v.TTL), 10),
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

	builder, err := witness.ConfigCellDataBuilderByTypeArgs(req.Tx, common.ConfigCellTypeArgsProfitRate)
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
		Action:           common.DasActionAcceptOffer,
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
		Action:           common.DasActionAcceptOffer,
		ServiceType:      dao.ServiceTypeTransaction,
		InviterArgs:      common.Bytes2Hex(channelScript.Args().RawData()),
		InviterAccount:   "",
		InviterChainType: oCT,
		InviterAddress:   oA,
		BlockTimestamp:   req.BlockTimestamp,
	})
	return list, nil
}
