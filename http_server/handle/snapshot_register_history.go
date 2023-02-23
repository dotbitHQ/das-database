package handle

import (
	"das_database/http_server/api_code"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
	"sort"
	"strings"
	"time"
)

type ReqSnapshotRegisterHistory struct {
	StartTime string `json:"start_time"`
}

type RespSnapshotRegisterHistory struct {
	Result string `json:"result"`
}

func (h *HttpHandle) JsonRpcSnapshotRegisterHistory(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqSnapshotRegisterHistory
	err := json.Unmarshal(p, &req)
	if err != nil {
		log.Error("json.Unmarshal err:", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return
	}
	if len(req) != 1 {
		log.Error("len(req) is :", len(req))
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return
	}

	if err = h.doSnapshotRegisterHistory(&req[0], apiResp); err != nil {
		log.Error("doSnapshotRegisterHistory err:", err.Error())
	}
}

func (h *HttpHandle) SnapshotRegisterHistory(ctx *gin.Context) {
	var (
		funcName = "SnapshotRegisterHistory"
		req      ReqSnapshotRegisterHistory
		apiResp  api_code.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName)
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, toolib.JsonString(req))

	if err = h.doSnapshotRegisterHistory(&req, &apiResp); err != nil {
		log.Error("doSnapshotRegisterHistory err:", err.Error(), funcName)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

type registerInfo struct {
	count4     uint64
	count5     uint64
	countAll   uint64
	countOwner uint64
}

func (h *HttpHandle) doSnapshotRegisterHistory(req *ReqSnapshotRegisterHistory, apiResp *api_code.ApiResp) error {
	var resp RespSnapshotRegisterHistory

	theTime, err := time.ParseInLocation("2006-01-02", req.StartTime, time.Local)
	if err != nil {
		fmt.Println(err.Error())
	}
	theTimestamp := uint64(theTime.Unix())
	log.Info("theTimestamp:", theTimestamp)

	page := Pagination{
		Page:    1,
		Size:    10000,
		maxSize: 10000,
	}

	var res = make(map[string]registerInfo)
	var owner = make(map[string]struct{})

	for {
		list, err := h.dbDao.GetRegisterHistory(page.GetLimit(), page.GetOffset())
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeDbError, "Failed to query history info")
			return fmt.Errorf("GetRegisterHistory err: %s", err.Error())
		}
		page.Page++
		for _, v := range list {
			_, length, _ := common.GetDotBitAccountLength(v.Account)
			tm := time.Unix(int64(v.RegisteredAt), 0)
			registeredAt := tm.Format("2006-01-02")
			var tmp registerInfo
			if item, ok := res[registeredAt]; ok {
				tmp.count4 = item.count4
				tmp.count5 = item.count5
				tmp.countAll = item.countAll
				tmp.countOwner = item.countOwner
			}
			if length == 4 {
				tmp.count4++
			} else if length >= 5 {
				tmp.count5++
			}
			tmp.countAll++

			if _, ok := owner[strings.ToLower(v.Owner)]; !ok {
				tmp.countOwner++
				owner[strings.ToLower(v.Owner)] = struct{}{}
			}
			if v.RegisteredAt < theTimestamp {
				continue
			}
			res[registeredAt] = tmp
		}
		if len(list) == 0 || len(list) < page.GetLimit() {
			break
		}
	}
	log.Info("res:", len(res))

	var strList []string
	for k, v := range res {
		str := "%s,%d,%d,%d,%d,\n"
		strList = append(strList, fmt.Sprintf(str, k, v.count4, v.count5, v.countAll, v.countOwner))
	}
	sort.Strings(strList)
	resp.Result = "Date,4Digit,5Digit,DailyNewCount,DailyNewOwner,\n"
	for _, v := range strList {
		resp.Result += v
	}

	apiResp.ApiRespOK(resp)
	return nil
}
