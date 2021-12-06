package chain_ckb

import (
	"fmt"
	"github.com/nervosnetwork/ckb-sdk-go/types"
)

func (c *Client) GetTxByHashOnChain(txHash types.Hash) (*types.TransactionWithStatus, error) {
	if res, err := c.client.GetTransaction(c.ctx, txHash); err != nil {
		return nil, fmt.Errorf("GetTransaction err:%s", err.Error())
	} else {
		return res, nil
	}
}

func (c *Client) GetHeaderByHashOnChain(blockHash types.Hash) (*types.Header, error) {
	if res, err := c.client.GetHeader(c.ctx, blockHash); err != nil {
		return nil, fmt.Errorf("GetHeader err:%s", err.Error())
	} else {
		return res, nil
	}
}
