package example

import (
	"das_database/http_server/api_code"
	"das_database/http_server/handle"
	"fmt"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/parnurzeal/gorequest"
	"github.com/scorpiotzh/toolib"
	"testing"
)

var ApiUrl = "https://test-snapshot-api.did.id/v1"

func TestSnapshotProgress(t *testing.T) {
	url := ApiUrl + "/snapshot/progress"

	var req handle.ReqSnapshotProgress
	var data handle.RespSnapshotProgress

	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
}

func TestSnapshotPermissionsInfo(t *testing.T) {
	url := ApiUrl + "/snapshot/permissions/info"

	req := handle.ReqSnapshotPermissionsInfo{
		Account:     "test.20230216.bit",
		BlockNumber: 8357751,
	}
	var data handle.RespSnapshotPermissionsInfo

	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
}

func TestSnapshotAddressAccounts(t *testing.T) {
	url := ApiUrl + "/snapshot/address/accounts"
	req := handle.ReqSnapshotAddressAccounts{
		ChainTypeAddress: core.ChainTypeAddress{
			Type: "blockchain",
			KeyInfo: core.KeyInfo{
				CoinType: "60",
				ChainId:  "",
				Key:      "0xc9f53b1d85356b60453f867610888d89a0b667ad",
			},
		},
		BlockNumber: 8357751,
		RoleType:    "manager",
		Pagination:  handle.Pagination{Page: 1, Size: 10},
	}
	var data handle.RespSnapshotAddressAccounts

	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
}

func TestSnapshotRegisterHistory(t *testing.T) {
	url := ApiUrl + "/snapshot/register/history"
	req := handle.ReqSnapshotRegisterHistory{StartTimestamp: 0}
	var data handle.RespSnapshotRegisterHistory

	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Println(data.Result))
}

func doReq(url string, req, data interface{}) error {
	var resp api_code.ApiResp
	resp.Data = &data

	_, _, errs := gorequest.New().Post(url).SendStruct(&req).EndStruct(&resp)
	if errs != nil {
		return fmt.Errorf("%v", errs)
	}
	if resp.ErrNo != api_code.ApiCodeSuccess {
		return fmt.Errorf("%d - %s", resp.ErrNo, resp.ErrMsg)
	}
	return nil
}
