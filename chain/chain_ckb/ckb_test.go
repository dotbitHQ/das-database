package chain_ckb

import (
	"context"
	"das_database/config"
	"encoding/hex"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/toolib"
	"testing"
	"time"
)

func getCkbClient() (*Client, error) {
	if err := config.InitCfg("../config/config.yaml"); err != nil {
		panic(fmt.Errorf("InitCfg err: %s", err))
	}
	return NewClient(context.Background(), config.Cfg.Chain.CkbUrl, config.Cfg.Chain.IndexUrl)
}

func TestSearch(t *testing.T) {
	c, err := getCkbClient()
	if err != nil {
		t.Fatal(err)
	}

	configCellArgs := "0xbc502a34a430e3e167c82a24db6f9237b15ebf35"
	configCellTypeCodeHash := "0x00000000000000000000000000000000000000000000000000545950455f4944"
	applyArgs := "0xc78fa9066af1624e600ccfb21df9546f900b2afe5d7940d91aefc115653f90d9"
	key := indexer.SearchKey{
		Script:     common.GetNormalLockScript(configCellArgs),
		ScriptType: indexer.ScriptTypeLock,
		Filter: &indexer.CellsFilter{
			Script: common.GetScript(configCellTypeCodeHash, applyArgs),
		},
	}
	res, err := c.client.GetCells(context.Background(), &key, indexer.SearchOrderDesc, 100, "")
	if err != nil {
		t.Fatal(err)
	}
	log.Info("res:", len(res.Objects))
	for _, v := range res.Objects {
		log.Info(v.TxIndex, v.OutPoint.TxHash)
	}
}

func Test712(t *testing.T) {
	c, err := getCkbClient()
	if err != nil {
		t.Fatal(err)
	}
	key := indexer.SearchKey{
		Script:     common.GetScript("0x9376c3b5811942960a846691e16e477cf43d7c7fa654067c9948dfcd09a32137", "0x04017fc487590b24639e32750f26de1513c6c0ca4804017fc487590b24639e32750f26de1513c6c0ca48"),
		ScriptType: indexer.ScriptTypeLock,
		Filter: &indexer.CellsFilter{
			OutputDataLenRange: &[2]uint64{0, 2},
			//Script: &types.Script{
			//	CodeHash: types.HexToHash("0x"),
			//	HashType: types.HashTypeType,
			//	Args:     nil,
			//}, //GetScript("0x4ff58f2c76b4ac26fdf675aa82541e02e4cf896279c6d6982d17b959788b2f0c", "0x"),
		},
	}
	res, err := c.client.GetCells(context.Background(), &key, indexer.SearchOrderDesc, 100, "")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(res))
}

func TestConfigCell(t *testing.T) {

	fmt.Println(common.Hex2Bytes(""), types.HexToHash("0x").Bytes())
	c, err := getCkbClient()
	if err != nil {
		t.Fatal(err)
	}
	key := indexer.SearchKey{
		Script:     common.GetNormalLockScript("0xbc502a34a430e3e167c82a24db6f9237b15ebf35"),
		ScriptType: indexer.ScriptTypeLock,
		Filter: &indexer.CellsFilter{
			Script: common.GetScript("0x030ac2acd9c016f9a4ab13d52c244d23aaea636e0cbd386ec660b79974946517", ""),
		},
	}
	res, err := c.client.GetCells(context.Background(), &key, indexer.SearchOrderDesc, 100, "")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(res))
}

func TestGetTxs(t *testing.T) {
	client, err := getCkbClient()
	if err != nil {
		t.Fatal(err)
	}
	key := indexer.SearchKey{
		Script:     common.GetScript("0xbf43c3602455798c1a61a596e0d95278864c552fafe231c063b3fabf97a8febc", "0x26e9aa4899003b28de08c00a6e946d422c18bbba"),
		ScriptType: indexer.ScriptTypeLock,
		Filter: &indexer.CellsFilter{
			BlockRange: &[2]uint64{5377315, 5377442},
		},
	}
	list, err := client.Client().GetTransactions(context.Background(), &key, indexer.SearchOrderDesc, 1000, "")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(list.Objects))
	var mapTx = make(map[string]types.Hash)
	for _, v := range list.Objects {
		mapTx[v.TxHash.Hex()] = v.TxHash
	}
	fmt.Println(len(mapTx))
	count := 0
	for k, v := range mapTx {
		tx, err := client.Client().GetTransaction(context.Background(), v)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			if len(tx.Transaction.Outputs) > 0 {
				if "6d91285768e7c96f1cea0173c8167ada2cfeabe8" == hex.EncodeToString(tx.Transaction.Outputs[0].Lock.Args) {
					count++
					fmt.Println(count, k, string(tx.Transaction.OutputsData[0]))
				}
			}
		}
	}

}

// 余额
func TestBalance(t *testing.T) {
	client, err := getCkbClient()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("111")
	fmt.Println(client.GetBalance("ckt1qyq27z6pccncqlaamnh8ttapwn260egnt67ss2cwvz"))
	fmt.Println(client.GetNormalLiveCell("ckt1qyq27z6pccncqlaamnh8ttapwn260egnt67ss2cwvz", 50999994254))
}

func TestHeightCell(t *testing.T) {
	codeHash := "0x96248cdefb09eed910018a847cfb51ad044c2d7db650112931760e3ef34a7e9a"
	client, err := getCkbClient()
	if err != nil {
		t.Fatal(err)
	}
	searchKey := &indexer.SearchKey{
		Script:     common.GetScript(codeHash, "0x01"),
		ScriptType: indexer.ScriptTypeType,
	}
	res, err := client.Client().GetCells(context.Background(), searchKey, indexer.SearchOrderDesc, 100, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range res.Objects {
		fmt.Println(v.OutPoint.TxHash.Hex(), v.OutPoint.Index)
	}
}

func TestGetBlockByNumber(t *testing.T) {
	client, err := getCkbClient()
	if err != nil {
		t.Fatal(err)
	}
	block, err := client.GetBlockByNumber(1)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(block.Header.Timestamp)
	fmt.Println(time.Now().UnixNano() / 1e6)
	fmt.Println(time.Unix(int64(block.Header.Timestamp/1e3), 0).String())
}

func TestTx(t *testing.T) {
	hash := "0x5e594f15662fc75fe01fd67c76cee02c79dea2e6573509e5408af5114afd459e"
	client, err := getCkbClient()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := client.Client().GetTransaction(context.Background(), types.HexToHash(hash))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(tx))
}

func TestGenerateAddress(t *testing.T) {
	script := &types.Script{
		CodeHash: types.HexToHash("0x58c5f491aba6d61678b7cf7edf4910b1f5e00ec0cde2f42e0abb4fd9aff25a63"),
		HashType: types.HashTypeType,
		Args:     common.Hex2Bytes("0xa266e3226426af7f30ae133fc0fdcdd761e69aac"),
	}
	tnAddress, err := address.ConvertScriptToAddress(address.Testnet, script)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(tnAddress)

	tnAddress, err = address.ConvertScriptToFullAddress(address.Testnet, script)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(tnAddress)

	//ckt1qpvvtay34wndv9nckl8hah6fzzcltcqwcrx79apwp2a5lkd07fdxxqdzvm3jyepx4alnptsn8lq0mnwhv8nf4tq2t44mg
	//ckt1q3vvtay34wndv9nckl8hah6fzzcltcqwcrx79apwp2a5lkd07fdx8gnxuv3xgf400uc2uyelcr7um4mpu6d2cu6sy2x
}

func TestGenerateDASAddress(t *testing.T) {
	script := &types.Script{
		CodeHash: types.HexToHash("0x326df166e3f0a900a0aee043e31a4dea0f01ea3307e6e235f09d1b4220b75fbd"),
		HashType: types.HashTypeType,
		Args:     common.Hex2Bytes("0x03a266e3226426af7f30ae133fc0fdcdd761e69aac03a266e3226426af7f30ae133fc0fdcdd761e69aac"),
	}
	tnAddress, err := address.ConvertScriptToAddress(address.Testnet, script)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(tnAddress)

	script = &types.Script{
		CodeHash: types.HexToHash("0x326df166e3f0a900a0aee043e31a4dea0f01ea3307e6e235f09d1b4220b75fbd"),
		HashType: types.HashTypeType,
		Args:     common.Hex2Bytes("0x05a266e3226426af7f30ae133fc0fdcdd761e69aac05a266e3226426af7f30ae133fc0fdcdd761e69aac"),
	}
	tnAddress, err = address.ConvertScriptToFullAddress(address.Testnet, script)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(tnAddress)

	//ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qgr5fnwxgnyy6hh7v9wzvluplwd6as7dx4vqw3xdcezvsn27les4cfnls8aehtkre564s9tursd
	//ckt1qsexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6pdzvm3jyepx4alnptsn8lq0mnwhv8nf4tq95fnwxgnyy6hh7v9wzvluplwd6as7dx4vr368uu
}

func TestGenerateShortAddress(t *testing.T) {
	tnAddress, err := address.GenerateShortAddress(address.Testnet)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(tnAddress.Address, tnAddress.LockArgs, tnAddress.PrivateKey)

	mnAddress, err := address.GenerateShortAddress(address.Mainnet)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(mnAddress.Address, mnAddress.LockArgs, mnAddress.PrivateKey)
}
