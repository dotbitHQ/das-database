package block_parser

import (
	"context"
	"das_database/config"
	"das_database/dao"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/dotbitHQ/das-notice/lark"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/mylog"
	"sync"
	"sync/atomic"
	"time"
)

var log = mylog.NewLogger("block_parser", mylog.LevelDebug)
var IsLatestBlockNumber bool

type BlockParser struct {
	dasCore              *core.DasCore
	mapTransactionHandle map[common.DasAction]FuncTransactionHandle
	currentBlockNumber   uint64
	dbDao                *dao.DbDao
	concurrencyNum       uint64
	confirmNum           uint64
	ctx                  context.Context
	wg                   *sync.WaitGroup
}

type ParamsBlockParser struct {
	DasCore            *core.DasCore
	CurrentBlockNumber uint64
	DbDao              *dao.DbDao
	ConcurrencyNum     uint64
	ConfirmNum         uint64
	Ctx                context.Context
	Wg                 *sync.WaitGroup
}

func NewBlockParser(p ParamsBlockParser) (*BlockParser, error) {
	bp := BlockParser{
		dasCore:            p.DasCore,
		currentBlockNumber: p.CurrentBlockNumber,
		dbDao:              p.DbDao,
		concurrencyNum:     p.ConcurrencyNum,
		confirmNum:         p.ConfirmNum,
		ctx:                p.Ctx,
		wg:                 p.Wg,
	}
	bp.registerTransactionHandle()
	if err := bp.initCurrentBlockNumber(); err != nil {
		return nil, fmt.Errorf("initCurrentBlockNumber err: %s", err.Error())
	}
	return &bp, nil
}

func (b *BlockParser) GetMapTransactionHandle(action common.DasAction) (FuncTransactionHandle, bool) {
	handler, ok := b.mapTransactionHandle[action]
	return handler, ok
}

func (b *BlockParser) initCurrentBlockNumber() error {
	if block, err := b.dbDao.FindBlockInfo(); err != nil {
		return err
	} else if block.Id > 0 {
		b.currentBlockNumber = block.BlockNumber
	}
	return nil
}

func (b *BlockParser) RunParser() {
	atomic.AddUint64(&b.currentBlockNumber, 1)
	b.wg.Add(1)
	go func() {
		for {
			select {
			default:
				// get the new height and compare with current height
				latestBlockNumber, err := b.dasCore.Client().GetTipBlockNumber(b.ctx)
				if err != nil {
					log.Error("get latest block number err:", err.Error())
				} else {
					// async
					if b.concurrencyNum > 1 && b.currentBlockNumber < (latestBlockNumber-b.confirmNum-b.concurrencyNum) {
						nowTime := time.Now()
						if err = b.parserConcurrencyMode(); err != nil {
							log.Error("parserConcurrencyMode err:", err.Error(), b.currentBlockNumber)
						}
						log.Warn("parserConcurrencyMode time:", time.Since(nowTime).Seconds())
					} else if b.currentBlockNumber < (latestBlockNumber - b.confirmNum) { // check rollback
						nowTime := time.Now()
						if err = b.parserSubMode(); err != nil {
							log.Error("parserSubMode err:", err.Error(), b.currentBlockNumber)
						}
						log.Warn("parserSubMode time:", time.Since(nowTime).Seconds())
					} else {
						log.Info("RunParser:", IsLatestBlockNumber, b.currentBlockNumber, latestBlockNumber)
						IsLatestBlockNumber = true
						time.Sleep(time.Second * 10)
					}
					time.Sleep(time.Millisecond * 300)
				}
			case <-b.ctx.Done():
				b.wg.Done()
				return
			}
		}
	}()
}

// subscribe mode
func (b *BlockParser) parserSubMode() error {
	log.Info("parserSubMode:", b.currentBlockNumber)
	block, err := b.dasCore.Client().GetBlockByNumber(b.ctx, b.currentBlockNumber)
	if err != nil {
		return fmt.Errorf("GetBlockByNumber err: %s", err.Error())
	} else {
		blockHash := block.Header.Hash.Hex()
		parentHash := block.Header.ParentHash.Hex()
		log.Info("parserSubMode:", b.currentBlockNumber, blockHash, parentHash)
		// block fork check
		if fork, err := b.checkFork(parentHash); err != nil {
			return fmt.Errorf("checkFork err: %s", err.Error())
		} else if fork {
			log.Warn("CheckFork is true:", b.currentBlockNumber, blockHash, parentHash)
			atomic.AddUint64(&b.currentBlockNumber, ^uint64(0))
		} else if err = b.parsingBlockData(block); err != nil {
			return fmt.Errorf("parsingBlockData err: %s", err.Error())
		} else {
			if err = b.dbDao.CreateBlockInfo(b.currentBlockNumber, blockHash, parentHash); err != nil {
				return fmt.Errorf("CreateBlockInfo err: %s", err.Error())
			} else {
				atomic.AddUint64(&b.currentBlockNumber, 1)
			}
			if err = b.dbDao.DeleteBlockInfo(b.currentBlockNumber - 20); err != nil {
				return fmt.Errorf("DeleteBlockInfo err: %s", err.Error())
			}
		}
	}
	return nil
}

// rollback checking
func (b *BlockParser) checkFork(parentHash string) (bool, error) {
	block, err := b.dbDao.FindBlockInfoByBlockNumber(b.currentBlockNumber - 1)
	if err != nil {
		return false, err
	}
	if block.Id == 0 {
		return false, nil
	} else if block.BlockHash != parentHash {
		log.Warn("CheckFork:", b.currentBlockNumber, parentHash, block.BlockHash)
		return true, nil
	}
	return false, nil
}

func (b *BlockParser) parsingBlockData(block *types.Block) error {
	for _, tx := range block.Transactions {
		txHash := tx.Hash.Hex()
		blockNumber := block.Header.Number
		blockTimestamp := block.Header.Timestamp
		log.Info("parsingBlockData txHash:", txHash)

		if builder, err := witness.ActionDataBuilderFromTx(tx); err != nil {
			log.Warn("ActionDataBuilderFromTx err:", err.Error())
		} else {
			if handle, ok := b.mapTransactionHandle[builder.Action]; ok {
				// transaction parse by action
				resp := handle(FuncTransactionHandleReq{
					DbDao:          b.dbDao,
					Tx:             tx,
					TxHash:         txHash,
					BlockNumber:    blockNumber,
					BlockTimestamp: blockTimestamp,
					Action:         builder.Action,
				})
				if resp.Err != nil {
					log.Error("action handle resp:", builder.Action, blockNumber, txHash, resp.Err.Error())
					// notify
					msg := "> Block number：%d\n> Action：%s\n> Timestamp：%s\n> Transaction hash：%s"
					msg = fmt.Sprintf(msg, blockNumber, builder.Action, time.Now().Format("2006-01-02 15:04:05"), txHash)
					err = lark.SendLarkTextNotify(config.Cfg.Notice.WebhookLarkErr, "DasDatabase block parser", msg)
					if err != nil {
						log.Error("SendLarkTextNotify err:", err.Error())
					}
					return resp.Err
				}
			}

		}
	}
	return nil
}

func (b *BlockParser) parserConcurrencyMode() error {
	log.Info("parserConcurrencyMode:", b.currentBlockNumber, b.concurrencyNum)
	for i := uint64(0); i < b.concurrencyNum; i++ {
		block, err := b.dasCore.Client().GetBlockByNumber(b.ctx, b.currentBlockNumber)
		if err != nil {
			return fmt.Errorf("GetBlockByNumber err: %s [%d]", err.Error(), b.currentBlockNumber)
		}
		blockHash := block.Header.Hash.Hex()
		parentHash := block.Header.ParentHash.Hex()
		log.Info("parserConcurrencyMode:", b.currentBlockNumber, blockHash, parentHash)

		if err = b.parsingBlockData(block); err != nil {
			return fmt.Errorf("parsingBlockData err: %s", err.Error())
		} else {
			if err = b.dbDao.CreateBlockInfo(b.currentBlockNumber, blockHash, parentHash); err != nil {
				return fmt.Errorf("CreateBlockInfo err: %s", err.Error())
			} else {
				atomic.AddUint64(&b.currentBlockNumber, 1)
			}
		}
	}
	if err := b.dbDao.DeleteBlockInfo(b.currentBlockNumber - 20); err != nil {
		return fmt.Errorf("DeleteBlockInfo err: %s", err.Error())
	}
	return nil
}
