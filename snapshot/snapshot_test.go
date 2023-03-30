package snapshot

import (
	"context"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/nervosnetwork/ckb-sdk-go/rpc"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"sort"
	"sync"
	"testing"
)

func TestCheckContractCodeHash(t *testing.T) {
	ctxServer, _ := context.WithCancel(context.Background())
	wgServer := sync.WaitGroup{}

	ckbUrl := "https://testnet.ckb.dev/"
	ckbClient, err := rpc.DialWithIndexer(ckbUrl, ckbUrl)
	if err != nil {
		t.Fatal(err)
	}

	env := core.InitEnv(common.DasNetTypeTestnet2)
	opts := []core.DasCoreOption{
		core.WithClient(ckbClient),
		core.WithDasContractArgs(env.ContractArgs),
		core.WithDasContractCodeHash(env.ContractCodeHash),
		core.WithDasNetType(common.DasNetTypeTestnet2),
		core.WithTHQCodeHash(env.THQCodeHash),
	}
	dc := core.NewDasCore(ctxServer, &wgServer, opts...)
	dc.InitDasContract(env.MapContract)

	res, err := dc.Client().GetTransaction(ctxServer, types.HexToHash("0x97697e27b8690a9cf10f297150f1305e5305f8c1d514619196829e8efceb856b"))
	if err != nil {
		t.Fatal(err)
	}
	var s ToolSnapshot
	fmt.Println(s.checkContractCodeHash(res.Transaction))
}

func TestBlockList(t *testing.T) {
	blockListTmp := make([]*types.Block, 0)
	blockListTmp = append(blockListTmp, &types.Block{Header: &types.Header{Number: 5007343}})
	blockListTmp = append(blockListTmp, &types.Block{Header: &types.Header{Number: 5007335}})
	blockListTmp = append(blockListTmp, &types.Block{Header: &types.Header{Number: 5007498}})
	sort.Sort(blockList(blockListTmp))
	for i := range blockListTmp {
		fmt.Println(blockListTmp[i].Header.Number)
	}
}
