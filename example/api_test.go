package example

import (
	"das_database/dao"
	"das_database/http_server/api_code"
	"das_database/http_server/handle"
	"fmt"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/parnurzeal/gorequest"
	"github.com/scorpiotzh/toolib"
	"golang.org/x/sync/errgroup"
	"testing"
	"time"
)

var ApiUrl = "https://snapshot-api.did.id/v1"

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
	req := handle.ReqSnapshotRegisterHistory{StartTime: "2023-02-10"}
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

func TestPage(t *testing.T) {
	page := handle.Pagination{
		Page: 1,
		Size: 20000,
	}
	fmt.Println(page.GetLimit(), page.GetOffset())
	page.SetMaxSize(20000)
	fmt.Println(page.GetLimit(), page.GetOffset())

	startTime := "2023-02-10"
	loc, _ := time.LoadLocation("Local")
	theTime, err := time.ParseInLocation("2006-01-02", startTime, loc)
	if err != nil {
		fmt.Println(err.Error())
	}
	theTimestamp := uint64(theTime.Unix())
	fmt.Println(theTimestamp)
}

func TestFixSnapshot(t *testing.T) {
	db, err := dao.NewGormDataBase("", "", "", "", 100, 200)
	if err != nil {
		t.Fatal(err)
	}

	var list []dao.TableSnapshotPermissionsInfo
	if err := db.Where(" `status`=99 ").Find(&list).Error; err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(list))
	errG := &errgroup.Group{}

	var ch = make(chan dao.TableSnapshotPermissionsInfo, 20)

	for i := 0; i < 20; i++ {
		errG.Go(func() error {
			for v := range ch {
				t.Log(v.AccountId, v.BlockNumber)
				var old dao.TableSnapshotPermissionsInfo
				if err := db.Where("account_id=? AND block_number<?", v.AccountId, v.BlockNumber).
					Order("block_number DESC").Limit(1).Find(&old).Error; err != nil {
					t.Log(err.Error(), v.AccountId, v.BlockNumber)
					continue
				}
				if err := db.Model(dao.TableSnapshotPermissionsInfo{}).
					Where("id IN(?)", []uint64{v.Id, old.Id}).
					Updates(map[string]interface{}{
						"owner_block_number":   v.BlockNumber,
						"manager_block_number": v.BlockNumber,
					}).Error; err != nil {
					t.Log(err.Error(), v.AccountId, v.BlockNumber)
				}
			}
			return nil
		})
	}
	errG.Go(func() error {
		for i := range list {
			ch <- list[i]
		}
		close(ch)
		return nil
	})

	if err := errG.Wait(); err != nil {
		t.Error(err.Error())
	}
	t.Log("OK")
}
