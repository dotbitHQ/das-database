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

var ApiUrl = "http://127.0.0.1:8118/v1"

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
		Account:     "7aaaaaaa.bit",
		BlockNumber: 3593828,
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
				CoinType: "195",
				ChainId:  "",
				Key:      "41a2ac25bf43680c05abe82c7b1bcc1a779cff8d5d",
			},
		},
		BlockNumber: 1941502,
		RoleType:    "manager",
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
	fmt.Println(toolib.JsonString(&data))
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
