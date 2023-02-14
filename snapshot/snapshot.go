package snapshot

import (
	"context"
	"das_database/config"
	"das_database/dao"
	"das_database/notify"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/mylog"
	"sync"
	"time"
)

var log = mylog.NewLogger("snapshot", mylog.LevelDebug)

type ToolSnapshot struct {
	Ctx            context.Context
	Wg             *sync.WaitGroup
	DbDao          *dao.DbDao
	DasCore        *core.DasCore
	ConcurrencyNum uint64
	ConfirmNum     uint64

	currentBlockNumber   uint64
	parserType           dao.ParserType
	mapTransactionHandle map[common.DasAction][]FuncTransactionHandle
	errTxCount           int
	errSnapshotCount     int
}

type FuncTransactionHandle func(info dao.TableSnapshotTxInfo, tx *types.Transaction) error

func (t *ToolSnapshot) init() error {
	t.parserType = dao.ParserTypeSnapshot
	t.registerTransactionHandle()
	if err := t.initCurrentBlockNumber(); err != nil {
		return fmt.Errorf("initCurrentBlockNumber err: %s", err.Error())
	}
	return nil
}
func (t *ToolSnapshot) Run(open bool) error {
	if !open {
		return nil
	}
	if err := t.init(); err != nil {
		return fmt.Errorf("ToolSnapshot init err: %s", err.Error())
	}
	t.RunTxSnapshot()

	t.RunDataSnapshot()
	return nil
}

func (t *ToolSnapshot) registerTransactionHandle() {
	t.mapTransactionHandle = make(map[string][]FuncTransactionHandle)

	t.mapTransactionHandle[common.DasActionTransferAccount] = []FuncTransactionHandle{t.addAccountPermissions}
	t.mapTransactionHandle[common.DasActionEditManager] = []FuncTransactionHandle{t.addAccountPermissions}
	t.mapTransactionHandle[common.DasActionBuyAccount] = []FuncTransactionHandle{t.addAccountPermissions}
	t.mapTransactionHandle[common.DasActionCancelAccountSale] = []FuncTransactionHandle{t.addAccountPermissions}
	t.mapTransactionHandle[common.DasActionAcceptOffer] = []FuncTransactionHandle{t.addAccountPermissions}
	t.mapTransactionHandle[common.DasActionStartAccountSale] = []FuncTransactionHandle{t.addAccountPermissions}
	t.mapTransactionHandle[common.DasActionConfirmProposal] = []FuncTransactionHandle{t.addAccountPermissionsByDasActionConfirmProposal, t.addAccountRegisterByDasActionConfirmProposal}
	t.mapTransactionHandle[common.DasActionRenewAccount] = []FuncTransactionHandle{t.addAccountPermissions}
	t.mapTransactionHandle[common.DasActionForceRecoverAccountStatus] = []FuncTransactionHandle{t.addAccountPermissions}

	t.mapTransactionHandle[common.DasActionUnlockAccountForCrossChain] = []FuncTransactionHandle{t.addAccountPermissions}
	t.mapTransactionHandle[common.DasActionLockAccountForCrossChain] = []FuncTransactionHandle{t.addAccountPermissions}
	t.mapTransactionHandle[common.DasActionRecycleExpiredAccount] = []FuncTransactionHandle{t.addAccountPermissionsByDasActionRecycleExpiredAccount}

	t.mapTransactionHandle[common.DasActionCreateSubAccount] = []FuncTransactionHandle{t.addSubAccountPermissionsByDasActionCreateSubAccount, t.addSubAccountRegisterByDasActionCreateSubAccount}
	t.mapTransactionHandle[common.DasActionEditSubAccount] = []FuncTransactionHandle{t.addSubAccountPermissionsByDasActionEditSubAccount}
	t.mapTransactionHandle[common.DasActionUpdateSubAccount] = []FuncTransactionHandle{t.addSubAccountPermissions, t.addSubAccountRegister}
}

func (t *ToolSnapshot) RunDataSnapshot() {
	t.Wg.Add(1)
	tickerParser := time.NewTicker(time.Second * 10)
	go func() {
		for {
			select {
			case <-tickerParser.C:
				if err := t.runDataSnapshot(); err != nil {
					log.Error("runDataSnapshot err:", err.Error())
					if t.errSnapshotCount < 100 {
						t.errSnapshotCount++
						if err = notify.SendLarkTextNotify(config.Cfg.Notice.WebhookLarkErr, "runDataSnapshot", err.Error()); err != nil {
							log.Error("SendLarkTextNotify err: %s", err.Error())
						}
					}
				} else {
					t.errSnapshotCount = 0
				}
			case <-t.Ctx.Done():
				t.Wg.Done()
				return
			}
		}
	}()
}

func (t *ToolSnapshot) runDataSnapshot() error {
	// get currentBlockNumber
	currentBlockNumber := uint64(0)
	txS, err := t.DbDao.GetTxSnapshotSchedule()
	if err != nil {
		return fmt.Errorf("GetTxSnapshotSchedule err: %s", err.Error())
	} else if txS.Id > 0 {
		currentBlockNumber = txS.BlockNumber
	} else {
		if err := t.DbDao.InitSnapshotSchedule(); err != nil {
			return fmt.Errorf("InitSnapshotSchedule err: %s", err.Error())
		}
	}
	log.Info("runDataSnapshot:", currentBlockNumber)

	// get parser list
	list, err := t.DbDao.GetTxSnapshotByBlockNumber(currentBlockNumber)
	if err != nil {
		return fmt.Errorf("GetTxSnapshotByBlockNumber err: %s", err.Error())
	}

	// parser
	for _, v := range list {
		if er := t.doDataSnapshotParser(v); er != nil {
			return fmt.Errorf("doDataSnapshotParser err: %s", er.Error())
		}
		if v.BlockNumber > currentBlockNumber {
			currentBlockNumber = v.BlockNumber
			// update
			if err := t.DbDao.UpdateTxSnapshotSchedule(currentBlockNumber); err != nil {
				return fmt.Errorf("UpdateTxSnapshotSchedule err: %s", err.Error())
			}
		}
	}

	return nil
}

func (t *ToolSnapshot) doDataSnapshotParser(info dao.TableSnapshotTxInfo) error {
	res, err := t.DasCore.Client().GetTransaction(t.Ctx, types.HexToHash(info.Hash))
	if err != nil {
		return fmt.Errorf("GetTransaction err: %s", err.Error())
	}

	if handleList, ok := t.mapTransactionHandle[info.Action]; ok {
		for _, handle := range handleList {
			if err := handle(info, res.Transaction); err != nil {
				return fmt.Errorf("handle err: %s", err.Error())
			}
		}
	}

	return nil
}
