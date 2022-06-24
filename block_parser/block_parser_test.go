package block_parser

import (
	"context"
	"das_database/config"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/nervosnetwork/ckb-sdk-go/rpc"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"sync/atomic"
	"testing"
)

func TestBlockNumber(t *testing.T) {
	blockNumber := uint64(0)
	fmt.Println(^uint64(0))
	blockNumber2 := atomic.AddUint64(&blockNumber, 1)
	fmt.Println(blockNumber, blockNumber2)
	blockNumber2 = atomic.AddUint64(&blockNumber, ^uint64(0))
	fmt.Println(blockNumber, blockNumber2)
}

func getCkbClient() (rpc.Client, error) {
	if err := config.InitCfg("../config/config.yaml"); err != nil {
		panic(fmt.Errorf("InitCfg err: %s", err))
	}
	return rpc.DialWithIndexer(config.Cfg.Chain.CkbUrl, config.Cfg.Chain.IndexUrl)
}

func TestBuyAccount(t *testing.T) {
	c, err := getCkbClient()
	if err != nil {
		t.Fatal(err)
	}
	res, err := c.GetTransaction(context.Background(), types.HexToHash("0xd1ec867a25b7982ac95e13d129ecd44b0a5ca5b459363c4909fbfa07d1ccf28b"))
	if err != nil {
		t.Fatal(err)
	}
	actionDataBuilder, err := witness.ActionDataBuilderFromTx(res.Transaction)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(actionDataBuilder.Action)
	s, _ := actionDataBuilder.ActionBuyAccountInviterScript()
	fmt.Println(common.Bytes2Hex(s.Args().RawData()))
	IncomeCellBuilder, err := witness.IncomeCellDataBuilderFromTx(res.Transaction, common.DataTypeNew)
	if err != nil {
		t.Fatal(err)
	}
	list := IncomeCellBuilder.Records()
	fmt.Println(list)
}
