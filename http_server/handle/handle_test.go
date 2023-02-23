package handle

import (
	"das_database/dao"
	"das_database/http_server/api_code"
	"fmt"
	"testing"
)

func TestHistory(t *testing.T) {
	var h HttpHandle
	db, err := dao.NewGormDataBase("", "", "", "", 100, 200)
	if err != nil {
		t.Fatal(err)
	}
	h.dbDao = dao.NewDbDao(db)
	req := ReqSnapshotRegisterHistory{StartTime: "2023-02-10"}
	var apiResp api_code.ApiResp
	if err := h.doSnapshotRegisterHistory(&req, &apiResp); err != nil {
		t.Fatal(err)
	}
	fmt.Println(apiResp.Data)
}
