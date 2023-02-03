package snapshot

import (
	"das_database/dao"
	"fmt"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/nervosnetwork/ckb-sdk-go/types"
)

func (t *ToolSnapshot) RunTxSnapshot() {
	// todo
}

func (t *ToolSnapshot) parserConcurrencyMode() error {
	// todo
	return nil
}

func (t *ToolSnapshot) parserMode() error {
	// todo
	return nil
}

func (t *ToolSnapshot) checkContractVersion() error {
	// todo
	return nil
}

func (t *ToolSnapshot) checkContractCodeHash(tx *types.Transaction) (bool, error) {
	// todo
	return true, nil
}

func (t *ToolSnapshot) parsingBlockData(block *types.Block) error {
	if err := t.checkContractVersion(); err != nil {
		return fmt.Errorf("checkContractVersion err: %s", err.Error())
	}
	for _, tx := range block.Transactions {
		txHash := tx.Hash.Hex()
		blockNumber := block.Header.Number
		blockTimestamp := block.Header.Timestamp
		log.Info("parsingBlockData txHash:", txHash)

		if builder, err := witness.ActionDataBuilderFromTx(tx); err != nil {
			log.Warn("ActionDataBuilderFromTx err:", err.Error())
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
