package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/molecule"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/shopspring/decimal"
	"strconv"
)

func (b *BlockParser) ActionPropose(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameProposalCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version proposal tx")
		return
	}
	log.Info("ActionPropose:", req.BlockNumber, req.TxHash)

	preAccMap, err := witness.PreAccountCellDataBuilderMapFromTx(req.Tx, common.DataTypeDep)
	if err != nil {
		resp.Err = fmt.Errorf("PreAccountCellDataBuilderMapFromTx err: %s", err.Error())
		return
	}
	log.Info("ActionPropose:", len(preAccMap))

	proBuilder, err := witness.ProposalCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("ProposalCellDataBuilderFromTx err: %s", err.Error())
		return
	}

	var transactionInfos []dao.TableTransactionInfo
	for _, v := range preAccMap {
		accountId := common.Bytes2Hex(common.GetAccountIdByAccount(v.Account))
		transactionInfos = append(transactionInfos, dao.TableTransactionInfo{
			BlockNumber:    req.BlockNumber,
			AccountId:      accountId,
			Account:        v.Account,
			Action:         common.DasActionPropose,
			ServiceType:    dao.ServiceTypeRegister,
			ChainType:      common.ChainTypeCkb,
			Address:        common.Bytes2Hex(proBuilder.ProposalCellData.ProposerLock().Args().RawData()),
			Capacity:       req.Tx.Outputs[0].Capacity,
			Outpoint:       common.OutPoint2String(req.TxHash, uint(v.Index)),
			BlockTimestamp: req.BlockTimestamp,
		})
	}

	if err = b.dbDao.CreateTransactionInfoList(transactionInfos); err != nil {
		log.Error("CreateTransactionInfoList err:", err.Error(), req.TxHash, req.BlockNumber)
		resp.Err = fmt.Errorf("CreateTransactionInfoList err: %s ", err.Error())
		return
	}

	return
}

func (b *BlockParser) ActionConfirmProposal(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version confirm proposal tx")
		return
	}
	log.Info("ActionConfirmProposal:", req.BlockNumber, req.TxHash)

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
				Action:         common.DasActionConfirmProposal,
				Outpoint:       common.OutPoint2String(req.TxHash, uint(i)),
				Capacity:       v.Capacity,
				BlockTimestamp: req.BlockTimestamp,
				Status:         dao.IncomeCellStatusUnMerge,
			})
		}
	}

	preMap, err := witness.PreAccountCellDataBuilderMapFromTx(req.Tx, common.DataTypeOld) //new account
	if err != nil {
		resp.Err = fmt.Errorf("PreAccountCellDataBuilderMapFromTx err: %s", err.Error())
		return
	}
	accMap, err := witness.AccountIdCellDataBuilderFromTx(req.Tx, common.DataTypeNew) //old+new account
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderMapFromTx err: %s", err.Error())
		return
	}

	log.Info("ActionConfirmProposal:", len(preMap), len(accMap))

	var accountInfos []dao.TableAccountInfo
	var transactionInfos []dao.TableTransactionInfo
	var rebateInfos []dao.TableRebateInfo
	var records []dao.TableRecordsInfo
	var recordAccountIds []string
	var cidPks []dao.TableCidPk
	// account basic store fee
	configCell, err := b.dasCore.ConfigCellDataBuilderByTypeArgsList(common.ConfigCellTypeArgsAccount, common.ConfigCellTypeArgsProfitRate)
	if err != nil {
		resp.Err = fmt.Errorf("ConfigCellDataBuilderByTypeArgs err: %s", err.Error())
		return
	}
	//basicCapacity, _ := configCell.BasicCapacity()
	// rebate rate
	profitRateInviter, _ := configCell.ProfitRateInviter()
	profitRateChannel, _ := configCell.ProfitRateChannel()

	for _, v := range accMap {
		ownerHex, managerHex, err := b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[v.Index].Lock.Args)
		if err != nil {
			resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
			return
		}
		if ownerHex.DasAlgorithmId == common.DasAlgorithmIdWebauthn {
			cidPks = append(cidPks, dao.TableCidPk{
				Cid: common.Bytes2Hex(ownerHex.AddressPayload[:10]),
				Pk:  common.Bytes2Hex(ownerHex.AddressPayload[10:]),
			})
		}

		// charset
		charsetNum := common.ConvertAccountCharsToCharsetNum(v.AccountChars)
		accountInfos = append(accountInfos, dao.TableAccountInfo{
			BlockNumber:         req.BlockNumber,
			Outpoint:            common.OutPoint2String(req.TxHash, uint(v.Index)),
			AccountId:           v.AccountId,
			Account:             v.Account,
			OwnerChainType:      ownerHex.ChainType,
			Owner:               ownerHex.AddressHex,
			OwnerAlgorithmId:    ownerHex.DasAlgorithmId,
			OwnerSubAid:         ownerHex.DasSubAlgorithmId,
			ManagerChainType:    managerHex.ChainType,
			Manager:             managerHex.AddressHex,
			ManagerAlgorithmId:  managerHex.DasAlgorithmId,
			ManagerSubAid:       managerHex.DasSubAlgorithmId,
			Status:              v.Status,
			RegisteredAt:        v.RegisteredAt,
			ExpiredAt:           v.ExpiredAt,
			ConfirmProposalHash: req.TxHash,
			CharsetNum:          charsetNum,
		})

		if preAcc, ok := preMap[v.Account]; ok {
			preTx, err := b.dasCore.Client().GetTransaction(b.ctx, req.Tx.Inputs[preAcc.Index].PreviousOutput.TxHash)
			if err != nil {
				resp.Err = fmt.Errorf("GetTransaction err: %s", err.Error())
				return
			}

			transactionInfos = append(transactionInfos, dao.TableTransactionInfo{
				BlockNumber:    req.BlockNumber,
				AccountId:      v.AccountId,
				Account:        v.Account,
				Action:         common.DasActionConfirmProposal,
				ServiceType:    dao.ServiceTypeRegister,
				ChainType:      ownerHex.ChainType,
				Address:        ownerHex.AddressHex,
				Capacity:       req.Tx.Outputs[v.Index].Capacity,
				Outpoint:       common.OutPoint2String(req.TxHash, uint(v.Index)),
				BlockTimestamp: req.BlockTimestamp,
			})

			argsStr := preAcc.OwnerLockArgs
			inviteeHex, _, err := b.dasCore.Daf().ArgsToHex(common.Hex2Bytes(argsStr))
			if err != nil {
				resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
				return
			}
			inviterId := preAcc.InviterId
			accLen := uint64(len([]byte(preAcc.Account))) * common.OneCkb

			basicCapacity, _ := configCell.BasicCapacityFromOwnerDasAlgorithmId(argsStr)
			log.Info("ActionConfirmProposal:", basicCapacity, profitRateInviter, profitRateChannel)

			preCapacity := preTx.Transaction.Outputs[req.Tx.Inputs[preAcc.Index].PreviousOutput.Index].Capacity - basicCapacity - accLen // 扣除存储费，账号长度
			capacity, _ := decimal.NewFromString(fmt.Sprintf("%d", preCapacity))

			inviterLock := preAcc.InviterLock
			if inviterLock == nil {
				log.Warn("InviterLock nil:", req.BlockNumber, req.TxHash, preAcc.Account)
				tmp := molecule.ScriptDefault()
				inviterLock = &tmp
			}
			inviterHex, _, err := b.dasCore.Daf().ScriptToHex(molecule.MoleculeScript2CkbScript(inviterLock))
			if err != nil {
				resp.Err = fmt.Errorf("ScriptToHex err: %s", err.Error())
				return
			}
			inviteeId := common.Bytes2Hex(common.GetAccountIdByAccount(preAcc.Account))
			rebateInfos = append(rebateInfos, dao.TableRebateInfo{
				BlockNumber:      req.BlockNumber,
				Outpoint:         common.OutPoint2String(req.TxHash, uint(v.Index)),
				InviteeId:        inviteeId,
				InviteeAccount:   preAcc.Account,
				InviteeChainType: inviteeHex.ChainType,
				InviteeAddress:   inviteeHex.AddressHex,
				RewardType:       dao.RewardTypeInviter,
				Reward:           capacity.Div(decimal.NewFromInt(common.PercentRateBase)).Mul(decimal.NewFromInt(int64(profitRateInviter))).BigInt().Uint64(),
				Action:           common.DasActionConfirmProposal,
				ServiceType:      dao.ServiceTypeRegister,
				InviterArgs:      common.Bytes2Hex(inviterLock.Args().RawData()),
				InviterId:        inviterId,
				InviterChainType: inviterHex.ChainType,
				InviterAddress:   inviterHex.AddressHex,
				BlockTimestamp:   req.BlockTimestamp,
			})

			channelLock := preAcc.ChannelLock
			if channelLock == nil {
				log.Warn("ChannelLock nil:", req.BlockNumber, req.TxHash, preAcc.Account)
				tmp := molecule.ScriptDefault()
				channelLock = &tmp
			}
			channelHex, _, err := b.dasCore.Daf().ScriptToHex(molecule.MoleculeScript2CkbScript(channelLock))
			if err != nil {
				resp.Err = fmt.Errorf("ScriptToHex err: %s", err.Error())
				return
			}
			rebateInfos = append(rebateInfos, dao.TableRebateInfo{
				BlockNumber:      req.BlockNumber,
				Outpoint:         common.OutPoint2String(req.TxHash, uint(v.Index)),
				InviteeId:        inviteeId,
				InviteeAccount:   preAcc.Account,
				InviteeChainType: inviteeHex.ChainType,
				InviteeAddress:   inviteeHex.AddressHex,
				RewardType:       dao.RewardTypeChannel,
				Reward:           capacity.Div(decimal.NewFromInt(common.PercentRateBase)).Mul(decimal.NewFromInt(int64(profitRateChannel))).BigInt().Uint64(),
				Action:           common.DasActionConfirmProposal,
				ServiceType:      dao.ServiceTypeRegister,
				InviterArgs:      common.Bytes2Hex(channelLock.Args().RawData()),
				InviterChainType: channelHex.ChainType,
				InviterAddress:   channelHex.AddressHex,
				BlockTimestamp:   req.BlockTimestamp,
			})
			recordAccountIds = append(recordAccountIds, v.AccountId)
			for _, record := range v.Records {
				records = append(records, dao.TableRecordsInfo{
					AccountId: v.AccountId,
					Account:   v.Account,
					Key:       record.Key,
					Type:      record.Type,
					Label:     record.Label,
					Value:     record.Value,
					Ttl:       strconv.FormatUint(uint64(record.TTL), 10),
				})
			}
		}
	}

	// did cell
	_, req.TxDidCellMap, err = b.dasCore.TxToDidCellEntityAndAction(req.Tx)
	if err != nil {
		resp.Err = fmt.Errorf("TxToDidCellEntityAndAction err: %s", err.Error())
		return
	}
	txDidEntityWitness, err := witness.GetDidEntityFromTx(req.Tx)
	if err != nil {
		resp.Err = fmt.Errorf("witness.GetDidEntityFromTx err: %s", err.Error())
		return
	}
	var didCellList []dao.TableDidCellInfo
	var didCellRecords []dao.TableRecordsInfo
	var didCellTxs []dao.TableTransactionInfo
	for k, v := range req.TxDidCellMap.Outputs {
		_, cellDataNew, err := v.GetDataInfo()
		if err != nil {
			resp.Err = fmt.Errorf("GetDataInfo new err: %s[%s]", err.Error(), k)
			return
		}
		acc := cellDataNew.Account
		accId := common.Bytes2Hex(common.GetAccountIdByAccount(acc))
		tmp := dao.TableDidCellInfo{
			BlockNumber:  req.BlockNumber,
			Outpoint:     common.OutPointStruct2String(v.OutPoint),
			AccountId:    accId,
			Account:      acc,
			Args:         common.Bytes2Hex(v.Lock.Args),
			LockCodeHash: v.Lock.CodeHash.Hex(),
			ExpiredAt:    cellDataNew.ExpireAt,
		}
		didCellList = append(didCellList, tmp)
		if w, yes := txDidEntityWitness.Outputs[v.Index]; yes {
			for _, r := range w.DidCellWitnessDataV0.Records {
				didCellRecords = append(didCellRecords, dao.TableRecordsInfo{
					AccountId: accId,
					Account:   acc,
					Key:       r.Key,
					Type:      r.Type,
					Label:     r.Label,
					Value:     r.Value,
					Ttl:       strconv.FormatUint(uint64(r.TTL), 10),
				})
			}
		}
		txAddress, err := address.ConvertScriptToAddress(b.dasCore.GetCkbAddressMode(), v.Lock)
		if err != nil {
			resp.Err = fmt.Errorf("ConvertScriptToAddress err: %s", err.Error())
			return
		}
		cellCapacity := req.Tx.Outputs[v.Index].Capacity
		if accItem, ok := accMap[accId]; ok {
			cellCapacity += req.Tx.Outputs[accItem.Index].Capacity
		}
		didCellTxs = append(didCellTxs, dao.TableTransactionInfo{
			BlockNumber:    req.BlockNumber,
			AccountId:      accId,
			Account:        acc,
			Action:         common.DasActionConfirmProposal,
			ServiceType:    dao.ServiceTypeRegister,
			ChainType:      common.ChainTypeAnyLock,
			Address:        txAddress,
			Capacity:       cellCapacity,
			Outpoint:       common.OutPoint2String(req.TxHash, uint(v.Index)),
			BlockTimestamp: req.BlockTimestamp,
		})
	}
	if len(didCellList) > 0 {
		records = didCellRecords
		transactionInfos = didCellTxs
	}

	if err = b.dbDao.ConfirmProposal(incomeCellInfos, accountInfos, transactionInfos, rebateInfos, records, recordAccountIds, cidPks, didCellList); err != nil {
		log.Error("ConfirmProposal err:", err.Error(), req.TxHash, req.BlockNumber)
		resp.Err = fmt.Errorf("ConfirmProposal err: %s ", err.Error())
		return
	}

	return
}
