package snapshot

import (
	"das_database/config"
	"das_database/dao"
	"das_database/notify"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"golang.org/x/sync/errgroup"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

func (t *ToolSnapshot) initCurrentBlockNumber() error {
	if block, err := t.DbDao.FindBlockInfo(t.parserType); err != nil {
		return err
	} else if block.Id > 0 {
		t.currentBlockNumber = block.BlockNumber
	} else if t.DasCore.NetType() == common.DasNetTypeMainNet {
		t.currentBlockNumber = 4872287
	} else {
		t.currentBlockNumber = 1927285
	}
	return nil
}

func (t *ToolSnapshot) RunTxSnapshot() {

	atomic.AddUint64(&t.currentBlockNumber, 1)
	t.Wg.Add(1)
	go func() {
		for {
			select {
			default:
				latestBlockNumber, err := t.DasCore.Client().GetTipBlockNumber(t.Ctx)
				if err != nil {
					log.Error("GetTipBlockNumber err:", err.Error())
				} else {
					if t.ConcurrencyNum > 1 && t.currentBlockNumber < (latestBlockNumber-t.ConfirmNum-t.ConcurrencyNum) {
						nowTime := time.Now()
						if err = t.parserConcurrencyMode(); err != nil {
							log.Error("parserConcurrencyMode err:", err.Error(), t.currentBlockNumber)
							if t.errTxCount < 100 {
								t.errTxCount++
								if err = notify.SendLarkTextNotify(config.Cfg.Notice.WebhookLarkErr, "RunTxSnapshot parserConcurrencyMode", err.Error()); err != nil {
									log.Error("SendLarkTextNotify err: %s", err.Error())
								}
							}
						} else {
							t.errTxCount = 0
						}
						log.Warn("parserConcurrencyMode time:", time.Since(nowTime).Seconds())
					} else if t.currentBlockNumber < (latestBlockNumber - t.ConfirmNum) {
						nowTime := time.Now()
						if err = t.parserMode(); err != nil {
							log.Error("parserMode err:", err.Error(), t.currentBlockNumber)
							if t.errTxCount < 100 {
								t.errTxCount++
								if err = notify.SendLarkTextNotify(config.Cfg.Notice.WebhookLarkErr, "RunTxSnapshot parserMode", err.Error()); err != nil {
									log.Error("SendLarkTextNotify err: %s", err.Error())
								}
							}
						} else {
							t.errTxCount = 0
						}
						log.Warn("parserMode time:", time.Since(nowTime).Seconds())
					} else {
						log.Info("RunParser:", t.currentBlockNumber, latestBlockNumber)
						time.Sleep(time.Second * 10)
					}
					time.Sleep(time.Millisecond * 300)
				}
			case <-t.Ctx.Done():
				t.Wg.Done()
				return
			}
		}
	}()
}

type blockList []*types.Block

func (b blockList) Len() int {
	return len(b)
}
func (b blockList) Less(i, j int) bool {
	return b[i].Header.Number > b[j].Header.Number
}

func (b blockList) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (t *ToolSnapshot) parserConcurrencyMode() error {
	log.Info("parserConcurrencyMode:", t.currentBlockNumber, t.ConcurrencyNum)

	blockLock := &sync.Mutex{}
	blockInfoList := make([]dao.TableBlockInfo, 0)
	blockListTmp := make([]*types.Block, 0)

	errGroup := &errgroup.Group{}
	for i := uint64(0); i < t.ConcurrencyNum; i++ {
		blockNumber := t.currentBlockNumber + i
		errGroup.Go(func() error {
			block, err := t.DasCore.Client().GetBlockByNumber(t.Ctx, blockNumber)
			if err != nil {
				return fmt.Errorf("GetBlockByNumber err: %s [%d]", err, blockNumber)
			}
			blockHash := block.Header.Hash.Hex()
			parentHash := block.Header.ParentHash.Hex()

			blockInfo := dao.TableBlockInfo{
				ParserType:  t.parserType,
				BlockNumber: blockNumber,
				BlockHash:   blockHash,
				ParentHash:  parentHash,
			}

			blockLock.Lock()
			blockInfoList = append(blockInfoList, blockInfo)
			blockListTmp = append(blockListTmp, block)
			blockLock.Unlock()
			return nil
		})
	}

	if err := errGroup.Wait(); err != nil {
		return err
	}
	sort.Sort(blockList(blockListTmp))

	for i := range blockListTmp {
		if err := t.parsingBlockData(blockListTmp[i]); err != nil {
			return fmt.Errorf("parsingBlockData err:%s", err.Error())
		}
	}

	if err := t.DbDao.CreateBlockInfoList(blockInfoList); err != nil {
		return fmt.Errorf("CreateBlockInfoList err:%s", err.Error())
	}

	atomic.AddUint64(&t.currentBlockNumber, t.ConcurrencyNum)

	if err := t.DbDao.DeleteBlockInfo(t.parserType, t.currentBlockNumber-20); err != nil {
		return fmt.Errorf("DeleteBlockInfo err: %s", err.Error())
	}
	return nil
}

func (t *ToolSnapshot) parserMode() error {
	log.Info("parserMode:", t.currentBlockNumber)
	block, err := t.DasCore.Client().GetBlockByNumber(t.Ctx, t.currentBlockNumber)
	if err != nil {
		return fmt.Errorf("GetBlockByNumber err: %s", err.Error())
	} else {
		blockHash := block.Header.Hash.Hex()
		parentHash := block.Header.ParentHash.Hex()
		log.Info("parserSubMode:", t.currentBlockNumber, blockHash, parentHash)
		// block fork check
		if fork, err := t.checkFork(parentHash); err != nil {
			return fmt.Errorf("checkFork err: %s", err.Error())
		} else if fork {
			log.Warn("CheckFork is true:", t.currentBlockNumber, blockHash, parentHash)
			atomic.AddUint64(&t.currentBlockNumber, ^uint64(0))
		} else if err = t.parsingBlockData(block); err != nil {
			return fmt.Errorf("parsingBlockData err: %s", err.Error())
		} else {
			if err = t.DbDao.CreateBlockInfo(t.parserType, t.currentBlockNumber, blockHash, parentHash); err != nil {
				return fmt.Errorf("CreateBlockInfo err: %s", err.Error())
			} else {
				atomic.AddUint64(&t.currentBlockNumber, 1)
			}
			if err = t.DbDao.DeleteBlockInfo(t.parserType, t.currentBlockNumber-20); err != nil {
				return fmt.Errorf("DeleteBlockInfo err: %s", err.Error())
			}
		}
	}
	return nil
}

func (t *ToolSnapshot) checkFork(parentHash string) (bool, error) {
	block, err := t.DbDao.FindBlockInfoByBlockNumber(t.parserType, t.currentBlockNumber-1)
	if err != nil {
		return false, err
	}
	if block.Id == 0 {
		return false, nil
	} else if block.BlockHash != parentHash {
		log.Warn("CheckFork:", t.currentBlockNumber, parentHash, block.BlockHash)
		return true, nil
	}
	return false, nil
}

func (t *ToolSnapshot) checkContractCodeHash(tx *types.Transaction) (bool, error) {
	isSelf := false
	for _, v := range tx.Outputs {
		core.DasContractMap.Range(func(key, value interface{}) bool {
			item, ok := value.(*core.DasContractInfo)
			if !ok {
				return true
			}
			log.Info(item.ContractName, item.ContractTypeId.String(), v.Lock.CodeHash.String())
			if item.IsSameTypeId(v.Lock.CodeHash) {
				isSelf = true
				return false
			}
			if v.Type != nil && item.IsSameTypeId(v.Type.CodeHash) {
				isSelf = true
				return false
			}
			return true
		})
		if isSelf {
			return true, nil
		}
	}
	return false, nil
}

func (t *ToolSnapshot) parsingBlockData(block *types.Block) error {
	if err := config.CheckContractVersion(t.DasCore, t.Cancel); err != nil {
		return err
	}
	for _, tx := range block.Transactions {
		txHash := tx.Hash.Hex()
		blockNumber := block.Header.Number
		blockTimestamp := block.Header.Timestamp
		//log.Info("parsingBlockData txHash:", txHash)

		if builder, err := witness.ActionDataBuilderFromTx(tx); err != nil {
			//log.Warn("ActionDataBuilderFromTx err:", err.Error())
		} else {
			log.Info("parsingBlockData action:", builder.Action, txHash)
			if ok, err := t.checkContractCodeHash(tx); err != nil {
				return fmt.Errorf("checkContractCodeHash err: %s", err.Error())
			} else if ok {
				info := dao.TableSnapshotTxInfo{
					BlockNumber:    blockNumber,
					Hash:           txHash,
					Action:         builder.Action,
					BlockTimestamp: blockTimestamp,
				}
				if err := t.DbDao.CreateSnapshotTxInfo(info); err != nil {
					return fmt.Errorf("CreateSnapshotTxInfo err: %s", err.Error())
				}
			} else {
				log.Warn("parsingBlockData give up:", blockNumber, txHash, builder.Action, blockTimestamp)
			}
		}
	}
	return nil
}
