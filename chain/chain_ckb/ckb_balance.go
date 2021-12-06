package chain_ckb

import (
	"fmt"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/collector"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/transaction"
	"github.com/nervosnetwork/ckb-sdk-go/types"
)

const (
	Min61Ckb = uint64(6100000000)
)

func (c *Client) GetNormalLiveCell(addr string, limit uint64) ([]*indexer.LiveCell, uint64, error) {
	parseAddr, err := address.Parse(addr)
	if err != nil {
		return nil, 0, fmt.Errorf("address.Parse err: %s", err.Error())
	}
	searchKey := &indexer.SearchKey{
		Script: &types.Script{
			CodeHash: types.HexToHash(transaction.SECP256K1_BLAKE160_SIGHASH_ALL_TYPE_HASH),
			HashType: types.HashTypeType,
			Args:     parseAddr.Script.Args,
		},
		ScriptType: indexer.ScriptTypeLock,
		Filter: &indexer.CellsFilter{
			OutputDataLenRange: &[2]uint64{0, 1},
		},
	}
	co := collector.NewLiveCellCollector(c.client, searchKey, indexer.SearchOrderAsc, indexer.SearchLimit, "")
	iterator, err := co.Iterator()
	if err != nil {
		return nil, 0, fmt.Errorf("iterator err:%s", err.Error())
	}
	var cells []*indexer.LiveCell
	total := uint64(0)
	for iterator.HasNext() {
		liveCell, err := iterator.CurrentItem()
		if err != nil {
			return nil, 0, fmt.Errorf("CurrentItem err:%s", err.Error())
		}
		//fmt.Println(liveCell)
		cells = append(cells, liveCell)
		total += liveCell.Output.Capacity
		if limit > 0 && (total == limit || total-limit > Min61Ckb) { // limit 为转账金额+手续费
			break
		}
		if err = iterator.Next(); err != nil {
			return nil, 0, fmt.Errorf("next err:%s", err.Error())
		}
	}
	return cells, total, nil
}

func (c *Client) GetBalance(addr string) (uint64, error) {
	parseAddr, err := address.Parse(addr)
	if err != nil {
		return 0, fmt.Errorf("address.Parse err: %s", err.Error())
	}
	searchKey := &indexer.SearchKey{
		Script: &types.Script{
			CodeHash: types.HexToHash(transaction.SECP256K1_BLAKE160_SIGHASH_ALL_TYPE_HASH),
			HashType: types.HashTypeType,
			Args:     parseAddr.Script.Args,
		},
		ScriptType: indexer.ScriptTypeLock,
		Filter: &indexer.CellsFilter{
			OutputDataLenRange: &[2]uint64{0, 1},
		},
	}
	res, err := c.client.GetCellsCapacity(c.ctx, searchKey)
	if err != nil {
		return 0, fmt.Errorf("GetCellsCapacity err: %s", err.Error())
	}
	return res.Capacity, nil
}
