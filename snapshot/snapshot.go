package snapshot

import (
	"context"
	"das_database/dao"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/mylog"
	"golang.org/x/sync/errgroup"
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
}

type FuncTransactionHandle func(info dao.TableSnapshotTxInfo, tx *types.Transaction) error

func (t *ToolSnapshot) registerTransactionHandle() {
	t.mapTransactionHandle = make(map[string][]FuncTransactionHandle)

	t.mapTransactionHandle[common.DasActionTransferAccount] = []FuncTransactionHandle{t.addAccountPermissions}
	t.mapTransactionHandle[common.DasActionEditManager] = []FuncTransactionHandle{t.addAccountPermissions}
	t.mapTransactionHandle[common.DasActionBuyAccount] = []FuncTransactionHandle{t.addAccountPermissions}
	t.mapTransactionHandle[common.DasActionAcceptOffer] = []FuncTransactionHandle{t.addAccountPermissions}
	t.mapTransactionHandle[common.DasActionUnlockAccountForCrossChain] = []FuncTransactionHandle{t.addAccountPermissions}
	t.mapTransactionHandle[common.DasActionStartAccountSale] = []FuncTransactionHandle{t.addAccountPermissions}
	t.mapTransactionHandle[common.DasActionLockAccountForCrossChain] = []FuncTransactionHandle{t.addAccountPermissions}

	t.mapTransactionHandle[common.DasActionConfirmProposal] = []FuncTransactionHandle{t.addAccountPermissionsByDasActionConfirmProposal, t.addAccountRegisterByDasActionConfirmProposal}
	t.mapTransactionHandle[common.DasActionRecycleExpiredAccount] = []FuncTransactionHandle{t.addAccountPermissionsByDasActionRecycleExpiredAccount}

	t.mapTransactionHandle[common.DasActionCreateSubAccount] = []FuncTransactionHandle{t.addSubAccountPermissions, t.addSubAccountRegister}
	t.mapTransactionHandle[common.DasActionEditSubAccount] = []FuncTransactionHandle{t.addSubAccountPermissions}
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
					log.Error("runDataSnapshot err:%s", err.Error())
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
	}

	// get parser list
	list, err := t.DbDao.GetTxSnapshotByBlockNumber(currentBlockNumber)
	if err != nil {
		return fmt.Errorf("GetTxSnapshotByBlockNumber err: %s", err.Error())
	}

	// parser
	ch := make(chan dao.TableSnapshotTxInfo, 10)
	errGroup := &errgroup.Group{}
	errGroup.Go(func() error {
		for i := range list {
			ch <- list[i]
		}
		close(ch)
		return nil
	})

	errGroup.Go(func() error {
		for v := range ch {
			if er := t.doDataSnapshotParser(v); er != nil {
				return fmt.Errorf("doDataSnapshotParser err: %s", er.Error())
			}
		}
		return nil
	})
	if err = errGroup.Wait(); err != nil {
		return err
	}

	// update
	if err := t.DbDao.UpdateTxSnapshotSchedule(list[len(list)-1].BlockNumber - 1); err != nil {
		return fmt.Errorf("UpdateTxSnapshotSchedule err: %s", err.Error())
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
