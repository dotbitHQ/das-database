package block_parser

import (
	"context"
	"das_database/config"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/nervosnetwork/ckb-sdk-go/rpc"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/toolib"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
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

func TestAccount(t *testing.T) {
	acc := "tzh03.00acc2022042902.bit"
	fmt.Println(acc[strings.Index(acc, ".")+1:])
}

func TestCheckContractVersion(t *testing.T) {
	dc, err := getNewDasCoreTestnet2()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	var wg sync.WaitGroup
	bp, err := NewBlockParser(ParamsBlockParser{
		DasCore:            dc,
		CurrentBlockNumber: config.Cfg.Chain.CurrentBlockNumber,
		DbDao:              nil,
		ConcurrencyNum:     config.Cfg.Chain.ConcurrencyNum,
		ConfirmNum:         config.Cfg.Chain.ConfirmNum,
		Ctx:                ctx,
		Wg:                 &wg,
	})
	if err != nil {
		t.Fatal(err)
	}

	wg.Add(2)
	go func() {
		ticker := time.NewTicker(time.Second * 2)
		for {
			select {
			case <-ticker.C:
				fmt.Println("checkContractVersion")
				if err := bp.checkContractVersion(); err != nil {
					t.Log(err)
				}
			case <-ctx.Done():
				wg.Done()
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Second * 2)
		count := 0
		for {
			select {
			case <-ticker.C:
				if count == 3 {
					core.ContractStatusMapTestNet[common.DasContractNameAccountCellType] = common.ContractStatus{Version: "1.6.0"}
				}
				fmt.Println("count:", count)
				count++

			case <-ctx.Done():
				wg.Done()
			}
		}
	}()
	fmt.Println(1111222)
	wg.Wait()
	toolib.ExitMonitoring(func(sig os.Signal) {
		log.Warn("ExitMonitoring:", sig.String())
	})
	fmt.Println(1111222)
	time.Sleep(time.Second * 5)
}

func getClientTestnet2() (rpc.Client, error) {
	ckbUrl := "https://testnet.ckb.dev/"
	indexerUrl := "https://testnet.ckb.dev/"
	return rpc.DialWithIndexer(ckbUrl, indexerUrl)
}

func getNewDasCoreTestnet2() (*core.DasCore, error) {
	client, err := getClientTestnet2()
	if err != nil {
		return nil, err
	}

	env := core.InitEnvOpt(common.DasNetTypeTestnet2,
		common.DasContractNameConfigCellType,
		//common.DasContractNameAccountCellType,
		//common.DasContractNameDispatchCellType,
		//common.DasContractNameBalanceCellType,
		common.DasContractNameAlwaysSuccess,
		common.DasContractNameIncomeCellType,
		//common.DASContractNameSubAccountCellType,
		//common.DasContractNamePreAccountCellType,
	)
	var wg sync.WaitGroup
	ops := []core.DasCoreOption{
		core.WithClient(client),
		core.WithDasContractArgs(env.ContractArgs),
		core.WithDasContractCodeHash(env.ContractCodeHash),
		core.WithDasNetType(common.DasNetTypeTestnet2),
		core.WithTHQCodeHash(env.THQCodeHash),
	}
	dc := core.NewDasCore(context.Background(), &wg, ops...)
	// contract
	dc.InitDasContract(env.MapContract)
	// config cell
	if err = dc.InitDasConfigCell(); err != nil {
		return nil, err
	}
	// so script
	if err = dc.InitDasSoScript(); err != nil {
		return nil, err
	}
	return dc, nil
}
