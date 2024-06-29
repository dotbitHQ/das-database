package block_parser

import (
	"bytes"
	"das_database/dao"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/witness"
	"strconv"
)

func (b *BlockParser) DidCellActionUpdate(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	log.Info("DidCellActionUpdate:", req.BlockNumber, req.TxHash, req.Action)

	if len(req.TxDidCellMap.Inputs) != len(req.TxDidCellMap.Outputs) {
		resp.Err = fmt.Errorf("len(req.TxDidCellMap.Inputs)!=len(req.TxDidCellMap.Outputs)")
		return
	}
	txDidEntityWitness, err := witness.GetDidEntityFromTx(req.Tx)
	if err != nil {
		resp.Err = fmt.Errorf("witness.GetDidEntityFromTx err: %s", err.Error())
		return
	}

	var oldOutpointList []string
	var list []dao.TableDidCellInfo
	var accountIds []string
	var records []dao.TableRecordsInfo
	var txList []dao.TableTransactionInfo

	for k, v := range req.TxDidCellMap.Inputs {
		_, cellDataOld, err := v.GetDataInfo()
		if err != nil {
			resp.Err = fmt.Errorf("GetDataInfo old err: %s[%s]", err.Error(), k)
			return
		}
		n, ok := req.TxDidCellMap.Outputs[k]
		if !ok {
			resp.Err = fmt.Errorf("TxDidCellMap diff err: %s[%s]", err.Error(), k)
			return
		}
		_, cellDataNew, err := n.GetDataInfo()
		if err != nil {
			resp.Err = fmt.Errorf("GetDataInfo new err: %s[%s]", err.Error(), k)
			return
		}
		account := cellDataOld.Account
		accountId := common.Bytes2Hex(common.GetAccountIdByAccount(account))

		oldOutpoint := common.OutPointStruct2String(v.OutPoint)
		oldOutpointList = append(oldOutpointList, oldOutpoint)

		tmp := dao.TableDidCellInfo{
			BlockNumber:  req.BlockNumber,
			Outpoint:     common.OutPointStruct2String(n.OutPoint),
			AccountId:    accountId,
			Account:      account,
			Args:         common.Bytes2Hex(n.Lock.Args),
			LockCodeHash: n.Lock.CodeHash.Hex(),
			ExpiredAt:    cellDataNew.ExpireAt,
		}
		list = append(list, tmp)
		addrOld, err := v.GetLockAddress(b.dasCore.NetType())
		if err != nil {
			resp.Err = fmt.Errorf("GetLockAddress err: %s", err.Error())
			return
		}
		if !v.Lock.Equals(n.Lock) {
			txList = append(txList, dao.TableTransactionInfo{
				BlockNumber:    req.BlockNumber,
				AccountId:      accountId,
				Account:        account,
				Action:         common.DidCellActionEditOwner,
				ServiceType:    dao.ServiceTypeRegister,
				ChainType:      common.ChainTypeAnyLock,
				Address:        addrOld,
				Capacity:       0,
				Outpoint:       common.OutPoint2String(req.TxHash, uint(v.Index)),
				BlockTimestamp: req.BlockTimestamp,
			})
		}
		if cellDataOld.ExpireAt != cellDataNew.ExpireAt {
			txList = append(txList, dao.TableTransactionInfo{
				BlockNumber:    req.BlockNumber,
				AccountId:      accountId,
				Account:        account,
				Action:         common.DidCellActionRenew,
				ServiceType:    dao.ServiceTypeRegister,
				ChainType:      common.ChainTypeAnyLock,
				Address:        addrOld,
				Capacity:       0,
				Outpoint:       common.OutPoint2String(req.TxHash, uint(v.Index)),
				BlockTimestamp: req.BlockTimestamp,
			})
		}

		if bytes.Compare(cellDataOld.WitnessHash, cellDataNew.WitnessHash) != 0 {
			txList = append(txList, dao.TableTransactionInfo{
				BlockNumber:    req.BlockNumber,
				AccountId:      accountId,
				Account:        account,
				Action:         common.DidCellActionEditRecords,
				ServiceType:    dao.ServiceTypeRegister,
				ChainType:      common.ChainTypeAnyLock,
				Address:        addrOld,
				Capacity:       0,
				Outpoint:       common.OutPoint2String(req.TxHash, uint(v.Index)),
				BlockTimestamp: req.BlockTimestamp,
			})

			accountIds = append(accountIds, accountId)
			if w, yes := txDidEntityWitness.Outputs[n.Index]; yes {
				for _, r := range w.DidCellWitnessDataV0.Records {
					records = append(records, dao.TableRecordsInfo{
						AccountId:       accountId,
						ParentAccountId: "",
						Account:         account,
						Key:             r.Key,
						Type:            r.Type,
						Label:           r.Label,
						Value:           r.Value,
						Ttl:             strconv.FormatUint(uint64(r.TTL), 10),
					})
				}
			}
		}
	}

	if err := b.dbDao.DidCellUpdateList(oldOutpointList, list, accountIds, records, txList); err != nil {
		resp.Err = fmt.Errorf("DidCellUpdateList err: %s", err.Error())
		return
	}
	return
}

func (b *BlockParser) DidCellActionRecycle(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	log.Info("DidCellActionRecycle:", req.BlockNumber, req.TxHash, req.Action)

	var oldOutpointList []string
	var accountIds []string
	var txList []dao.TableTransactionInfo
	for k, v := range req.TxDidCellMap.Inputs {
		oldOutpoint := common.OutPointStruct2String(v.OutPoint)
		oldOutpointList = append(oldOutpointList, oldOutpoint)

		_, cellData, err := v.GetDataInfo()
		if err != nil {
			resp.Err = fmt.Errorf("GetDataInfo err: %s[%s]", err.Error(), k)
			return
		}
		account := cellData.Account
		accountId := common.Bytes2Hex(common.GetAccountIdByAccount(account))
		accountIds = append(accountIds, accountId)

		anyLockAddr, err := v.GetLockAddress(b.dasCore.NetType())
		if err != nil {
			resp.Err = fmt.Errorf("GetLockAddress err: %s[%s]", err.Error(), k)
			return
		}

		txInfo := dao.TableTransactionInfo{
			BlockNumber:    req.BlockNumber,
			AccountId:      accountId,
			Account:        account,
			Action:         common.DidCellActionRecycle,
			ServiceType:    dao.ServiceTypeRegister,
			ChainType:      common.ChainTypeAnyLock,
			Address:        anyLockAddr,
			Capacity:       0,
			Outpoint:       common.OutPoint2String(req.TxHash, uint(v.Index)),
			BlockTimestamp: req.BlockTimestamp,
		}
		txList = append(txList, txInfo)
	}

	if err := b.dbDao.DidCellRecycleList(oldOutpointList, accountIds, txList); err != nil {
		resp.Err = fmt.Errorf("DidCellRecycleList err: %s", err.Error())
		return
	}
	return
}
