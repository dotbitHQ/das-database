package chain_ckb

import (
	"fmt"
	"github.com/nervosnetwork/ckb-sdk-go/types"
)

func (c *Client) GetTipBlockNumber() (uint64, error) {
	if blockNumber, err := c.client.GetTipBlockNumber(c.ctx); err != nil {
		return 0, fmt.Errorf("GetTipBlockNumber err:%s", err.Error())
	} else {
		return blockNumber, nil
	}
}

func (c *Client) GetBlockByNumber(blockNumber uint64) (*types.Block, error) {
	return c.client.GetBlockByNumber(c.ctx, blockNumber)
}
