package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/DeAccountSystems/das-lib/molecule"
	"github.com/DeAccountSystems/das-lib/witness"
	"github.com/shopspring/decimal"
)

func (b *BlockParser) ActionConfirmProposal(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version confirm proposal tx")
		return
	}

	log.Info("ActionConfirmProposal:", req.BlockNumber, req.TxHash)

	preMap, err := witness.PreAccountCellDataBuilderMapFromTx(req.Tx, common.DataTypeOld) //new account
	if err != nil {
		resp.Err = fmt.Errorf("PreAccountCellDataBuilderMapFromTx err: %s", err.Error())
		return
	}
	accMap, err := witness.AccountCellDataBuilderMapFromTx(req.Tx, common.DataTypeNew) //old+new account
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderMapFromTx err: %s", err.Error())
		return
	}

	log.Info("ActionConfirmProposal:", len(preMap), len(accMap))

	var accountInfos []dao.TableAccountInfo
	var transactionInfos []dao.TableTransactionInfo
	var rebateInfos []dao.TableRebateInfo
	// 账号基础存储费
	configCell, err := b.dasCore.ConfigCellDataBuilderByTypeArgs(common.ConfigCellTypeArgsAccount)
	if err != nil {
		resp.Err = fmt.Errorf("ConfigCellDataBuilderByTypeArgs err: %s", err.Error())
		return
	}
	basicCapacity, _ := configCell.BasicCapacity()
	// 反佣比例
	configCellRate, err := b.dasCore.ConfigCellDataBuilderByTypeArgs(common.ConfigCellTypeArgsProfitRate)
	if err != nil {
		resp.Err = fmt.Errorf("ConfigCellDataBuilderByTypeArgs err: %s", err.Error())
		return
	}
	profitRateInviter, _ := configCellRate.ProfitRateInviter()
	profitRateChannel, _ := configCellRate.ProfitRateChannel()

	log.Info("ActionConfirmProposal:", basicCapacity, profitRateInviter, profitRateChannel)

	for _, v := range accMap {
		oID, mID, oCT, mCT, oA, mA := core.FormatDasLockToHexAddress(req.Tx.Outputs[v.Index].Lock.Args)
		accountInfos = append(accountInfos, dao.TableAccountInfo{
			BlockNumber:         req.BlockNumber,
			Outpoint:            common.OutPoint2String(req.TxHash, uint(v.Index)),
			AccountId:           v.AccountId,
			Account:             v.Account,
			OwnerChainType:      oCT,
			Owner:               oA,
			OwnerAlgorithmId:    oID,
			ManagerChainType:    mCT,
			Manager:             mA,
			ManagerAlgorithmId:  mID,
			Status:              dao.AccountStatus(v.Status),
			RegisteredAt:        v.RegisteredAt,
			ExpiredAt:           v.ExpiredAt,
			ConfirmProposalHash: req.TxHash,
		})

		if preAcc, ok := preMap[v.Account]; ok {
			preTx, err := b.ckbClient.Client().GetTransaction(b.ctx, req.Tx.Inputs[preAcc.Index].PreviousOutput.TxHash)
			if err != nil {
				resp.Err = fmt.Errorf("pretx GetTransaction err: %s", err.Error())
				return
			}

			transactionInfos = append(transactionInfos, dao.TableTransactionInfo{
				BlockNumber:    req.BlockNumber,
				Account:        v.Account,
				Action:         common.DasActionConfirmProposal,
				ServiceType:    dao.ServiceTypeRegister,
				ChainType:      oCT,
				Address:        oA,
				Capacity:       req.Tx.Outputs[v.Index].Capacity,
				Outpoint:       common.OutPoint2String(req.TxHash, uint(v.Index)),
				BlockTimestamp: req.BlockTimestamp,
			})

			argsStr, _ := preAcc.OwnerLockArgsStr()
			_, _, inviteeOCT, _, inviteeOA, _ := core.FormatDasLockToHexAddress(common.Hex2Bytes(argsStr))
			inviterId, _ := preAcc.InviterId()
			accLen := uint64(len([]byte(preAcc.Account))) * common.OneCkb
			preCapacity := preTx.Transaction.Outputs[req.Tx.Inputs[preAcc.Index].PreviousOutput.Index].Capacity - basicCapacity - accLen // 扣除存储费，账号长度
			capacity, _ := decimal.NewFromString(fmt.Sprintf("%d", preCapacity))

			inviterLock, _ := preAcc.InviterLock()
			if inviterLock == nil {
				log.Warn("pre InviterLock nil:", req.BlockNumber, req.TxHash, preAcc.Account)
				tmp := molecule.ScriptDefault()
				inviterLock = &tmp
			}
			_, _, inviterOCT, _, inviterOA, _ := core.FormatDasLockToHexAddress(inviterLock.Args().RawData())
			rebateInfos = append(rebateInfos, dao.TableRebateInfo{
				BlockNumber:      req.BlockNumber,
				Outpoint:         common.OutPoint2String(req.TxHash, uint(v.Index)),
				InviteeAccount:   preAcc.Account,
				InviteeChainType: inviteeOCT,
				InviteeAddress:   inviteeOA,
				RewardType:       dao.RewardTypeInviter,
				Reward:           capacity.Div(decimal.NewFromInt(common.PercentRateBase)).Mul(decimal.NewFromInt(int64(profitRateInviter))).BigInt().Uint64(),
				Action:           common.DasActionConfirmProposal,
				ServiceType:      dao.ServiceTypeRegister,
				InviterArgs:      common.Bytes2Hex(inviterLock.Args().RawData()),
				InviterId:        inviterId,
				InviterChainType: inviterOCT,
				InviterAddress:   inviterOA,
				BlockTimestamp:   req.BlockTimestamp,
			})

			channelLock, _ := preAcc.ChannelLock()
			if channelLock == nil {
				log.Warn("pre ChannelLock nil:", req.BlockNumber, req.TxHash, preAcc.Account)
				tmp := molecule.ScriptDefault()
				channelLock = &tmp
			}
			_, _, channelOCT, _, channelOA, _ := core.FormatDasLockToHexAddress(channelLock.Args().RawData())
			rebateInfos = append(rebateInfos, dao.TableRebateInfo{
				BlockNumber:      req.BlockNumber,
				Outpoint:         common.OutPoint2String(req.TxHash, uint(v.Index)),
				InviteeAccount:   preAcc.Account,
				InviteeChainType: inviteeOCT,
				InviteeAddress:   inviteeOA,
				RewardType:       dao.RewardTypeChannel,
				Reward:           capacity.Div(decimal.NewFromInt(common.PercentRateBase)).Mul(decimal.NewFromInt(int64(profitRateChannel))).BigInt().Uint64(),
				Action:           common.DasActionConfirmProposal,
				ServiceType:      dao.ServiceTypeRegister,
				InviterArgs:      common.Bytes2Hex(channelLock.Args().RawData()),
				InviterChainType: channelOCT,
				InviterAddress:   channelOA,
				BlockTimestamp:   req.BlockTimestamp,
			})
		}
	}


	if err = b.dbDao.ConfirmProposal(accountInfos, transactionInfos, rebateInfos); err != nil {
		log.Error("ConfirmProposal err:", err.Error(), req.TxHash, req.BlockNumber)
		resp.Err = fmt.Errorf("ConfirmProposal err: %s ", err.Error())
		return
	}

	return
}
