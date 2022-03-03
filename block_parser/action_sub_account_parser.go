package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/DeAccountSystems/das-lib/witness"
	"github.com/nervosnetwork/ckb-sdk-go/crypto/blake2b"
	"strconv"
)

func (b *BlockParser) ActionEnableSubAccount(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DASContractNameSubAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version enable sub account tx")
		return
	}

	log.Info("ActionEnableSubAccount:", req.BlockNumber, req.TxHash)

	builder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	_, _, oCT, _, oA, _ := core.FormatDasLockToHexAddress(req.Tx.Outputs[builder.Index].Lock.Args)
	outpoint := common.OutPoint2String(req.TxHash, 1)

	accountInfo := dao.TableAccountInfo{
		BlockNumber:          req.BlockNumber,
		Outpoint:             outpoint,
		AccountId:            builder.AccountId,
		Account:              builder.Account,
		EnableSubAccount:     dao.AccountEnableStatusOn,
		RenewSubAccountPrice: builder.RenewSubAccountPrice,
	}
	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      builder.AccountId,
		Account:        builder.Account,
		Action:         common.DasActionEnableSubAccount,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      oCT,
		Address:        oA,
		Capacity:       req.Tx.Outputs[builder.Index].Capacity,
		Outpoint:       outpoint,
		BlockTimestamp: req.BlockTimestamp,
	}

	log.Info("ActionEnableSubAccount:", builder.Account)

	if err = b.dbDao.EnableSubAccount(accountInfo, transactionInfo); err != nil {
		resp.Err = fmt.Errorf("EnableSubAccount err: %s", err.Error())
		return
	}

	return
}

func (b *BlockParser) ActionCreateSubAccount(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DASContractNameSubAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version create sub account tx")
		return
	}

	log.Info("ActionCreateSubAccount:", req.BlockNumber, req.TxHash)

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
				Action:         common.DasActionCreateSubAccount,
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

	subAccountMap, err := witness.SubAccountDataBuilderMapFromTx(req.Tx)
	if err != nil {
		resp.Err = fmt.Errorf("SubAccountDataBuilderMapFromTx err: %s", err.Error())
		return
	}

	log.Info("ActionCreateSubAccount:", len(subAccountMap))

	var accountInfos []dao.TableAccountInfo
	var smtInfos []dao.TableSmtInfo
	var transactionInfos []dao.TableTransactionInfo
	for _, v := range subAccountMap {
		oID, mID, oCT, mCT, oA, mA := core.FormatDasLockToHexAddress(v.SubAccount.Lock.Args)
		outpoint := common.OutPoint2String(req.TxHash, 1)

		accountInfos = append(accountInfos, dao.TableAccountInfo{
			BlockNumber:          req.BlockNumber,
			Outpoint:             outpoint,
			AccountId:            v.SubAccount.AccountId,
			ParentAccountId:      builder.AccountId,
			Account:              v.Account,
			OwnerChainType:       oCT,
			Owner:                oA,
			OwnerAlgorithmId:     oID,
			ManagerChainType:     mCT,
			Manager:              mA,
			ManagerAlgorithmId:   mID,
			Status:               dao.AccountStatus(v.SubAccount.Status),
			EnableSubAccount:     dao.EnableSubAccount(v.SubAccount.EnableSubAccount),
			RenewSubAccountPrice: v.SubAccount.RenewSubAccountPrice,
			Nonce:                v.SubAccount.Nonce,
			RegisteredAt:         v.SubAccount.RegisteredAt,
			ExpiredAt:            v.SubAccount.ExpiredAt,
			ConfirmProposalHash:  req.TxHash,
		})
		bys, _ := blake2b.Blake256(v.MoleculeSubAccount.AsSlice())
		smtInfos = append(smtInfos, dao.TableSmtInfo{
			BlockNumber:     req.BlockNumber,
			Outpoint:        outpoint,
			AccountId:       v.SubAccount.AccountId,
			ParentAccountId: builder.AccountId,
			Account:         v.Account,
			LeafDataHash:    common.Bytes2Hex(bys),
		})
		transactionInfos = append(transactionInfos, dao.TableTransactionInfo{
			BlockNumber:    req.BlockNumber,
			AccountId:      v.SubAccount.AccountId,
			Account:        v.Account,
			Action:         common.DasActionCreateSubAccount,
			ServiceType:    dao.ServiceTypeRegister,
			ChainType:      oCT,
			Address:        oA,
			Capacity:       req.Tx.Outputs[1].Capacity,
			Outpoint:       outpoint,
			BlockTimestamp: req.BlockTimestamp,
		})
	}

	if err = b.dbDao.CreateSubAccount(incomeCellInfos, accountInfos, smtInfos, transactionInfos); err != nil {
		resp.Err = fmt.Errorf("CreateSubAccount err: %s", err.Error())
		return
	}

	return
}

func (b *BlockParser) ActionEditSubAccount(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DASContractNameSubAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version edit sub account tx")
		return
	}

	log.Info("ActionEditSubAccount:", req.BlockNumber, req.TxHash)

	builder, err := witness.SubAccountDataBuilderFromTx(req.Tx)
	if err != nil {
		resp.Err = fmt.Errorf("SubAccountDataBuilderFromTx err: %s", err.Error())
		return
	}
	oID, mID, oCT, mCT, oA, mA := core.FormatDasLockToHexAddress(builder.SubAccount.Lock.Args)
	outpoint := common.OutPoint2String(req.TxHash, 0)

	accountInfo := dao.TableAccountInfo{
		BlockNumber:          req.BlockNumber,
		Outpoint:             outpoint,
		AccountId:            builder.SubAccount.AccountId,
		OwnerChainType:       oCT,
		Owner:                oA,
		OwnerAlgorithmId:     oID,
		ManagerChainType:     mCT,
		Manager:              mA,
		ManagerAlgorithmId:   mID,
		Status:               dao.AccountStatus(builder.SubAccount.Status),
		EnableSubAccount:     dao.EnableSubAccount(builder.SubAccount.EnableSubAccount),
		RenewSubAccountPrice: builder.SubAccount.RenewSubAccountPrice,
		Nonce:                builder.SubAccount.Nonce,
		RegisteredAt:         builder.SubAccount.RegisteredAt,
		ExpiredAt:            builder.SubAccount.ExpiredAt,
	}
	bys, _ := blake2b.Blake256(builder.MoleculeSubAccount.AsSlice())
	smtInfo := dao.TableSmtInfo{
		BlockNumber:  req.BlockNumber,
		Outpoint:     outpoint,
		AccountId:    builder.SubAccount.AccountId,
		LeafDataHash: common.Bytes2Hex(bys),
	}
	transactionInfo := dao.TableTransactionInfo{
		BlockNumber:    req.BlockNumber,
		AccountId:      builder.SubAccount.AccountId,
		Account:        builder.Account,
		Action:         common.DasActionEditSubAccount,
		ServiceType:    dao.ServiceTypeRegister,
		ChainType:      oCT,
		Address:        oA,
		Capacity:       req.Tx.Outputs[0].Capacity,
		Outpoint:       outpoint,
		BlockTimestamp: req.BlockTimestamp,
	}

	subAccount := builder.ConvertToSubAccount()
	var recordsInfos []dao.TableRecordsInfo
	switch string(builder.EditKey) {
	case "lock":
		oID, mID, oCT, mCT, oA, mA = core.FormatDasLockToHexAddress(subAccount.Lock.Args)
		accountInfo.OwnerAlgorithmId = oID
		accountInfo.OwnerChainType = oCT
		accountInfo.Owner = oA
		accountInfo.ManagerAlgorithmId = mID
		accountInfo.ManagerChainType = mCT
		accountInfo.Manager = mA
	case "expired_at":
		accountInfo.ExpiredAt = subAccount.ExpiredAt
	case "status":
		accountInfo.Status = dao.AccountStatus(subAccount.Status)
	case "records":
		for _, v := range subAccount.Records {
			recordsInfos = append(recordsInfos, dao.TableRecordsInfo{
				AccountId: builder.SubAccount.AccountId,
				Account:   builder.Account,
				Key:       v.Key,
				Type:      v.Type,
				Label:     v.Label,
				Value:     v.Value,
				Ttl:       strconv.FormatUint(uint64(v.TTL), 10),
			})
		}
	case "enable_sub_account":
		accountInfo.EnableSubAccount = dao.EnableSubAccount(subAccount.EnableSubAccount)
	case "renew_sub_account_price":
		accountInfo.RenewSubAccountPrice = subAccount.RenewSubAccountPrice
	}

	log.Info("ActionEditSubAccount:", builder.Account)

	if err = b.dbDao.EditSubAccount(accountInfo, smtInfo, transactionInfo, recordsInfos); err != nil {
		resp.Err = fmt.Errorf("EditSubAccount err: %s", err.Error())
		return
	}

	return
}

func (b *BlockParser) ActionRenewSubAccount(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {

	return
}

func (b *BlockParser) ActionRecycleSubAccount(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {

	return
}
