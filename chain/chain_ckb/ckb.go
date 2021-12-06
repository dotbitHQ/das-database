package chain_ckb

import (
	"context"
	"fmt"
	"github.com/nervosnetwork/ckb-sdk-go/rpc"
	"github.com/scorpiotzh/mylog"
)

var (
	log = mylog.NewLogger("chain_ckb", mylog.LevelDebug)
)

type Client struct {
	ckbUrl     string
	indexerUrl string
	client     rpc.Client
	ctx        context.Context
}

func NewClient(ctx context.Context, ckbUrl, indexerUrl string) (*Client, error) {
	rpcClient, err := rpc.DialWithIndexer(ckbUrl, indexerUrl)
	if err != nil {
		return nil, fmt.Errorf("init ckb client err:%s", err.Error())
	}
	return &Client{
		ckbUrl:     ckbUrl,
		indexerUrl: indexerUrl,
		client:     rpcClient,
		ctx:        ctx,
	}, nil
}

func (c *Client) Client() rpc.Client {
	return c.client
}
