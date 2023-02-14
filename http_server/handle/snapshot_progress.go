package handle

import (
	"das_database/http_server/api_code"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

type ReqSnapshotProgress struct {
	BlockNumber uint64 `json:"block_number"`
}

type RespSnapshotProgress struct {
	BlockNumber uint64 `json:"block_number"`
}

func (h *HttpHandle) JsonRpcSnapshotProgress(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqSnapshotProgress
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

	if err = h.doSnapshotProgress(&req[0], apiResp); err != nil {
		log.Error("doSnapshotProgress err:", err.Error())
	}
}

func (h *HttpHandle) SnapshotProgress(ctx *gin.Context) {
	var (
		funcName = "SnapshotProgress"
		req      ReqSnapshotProgress
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

	if err = h.doSnapshotProgress(&req, &apiResp); err != nil {
		log.Error("doSnapshotProgress err:", err.Error(), funcName)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doSnapshotProgress(req *ReqSnapshotProgress, apiResp *api_code.ApiResp) error {
	var resp RespSnapshotProgress

	txS, err := h.dbDao.GetTxSnapshotSchedule()
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "Failed to query snapshot progress")
		return fmt.Errorf("GetTxSnapshotSchedule err: %s", err.Error())
	} else if txS.Id > 0 {
		resp.BlockNumber = txS.BlockNumber
	}
	log.Info("doSnapshotProgress:", resp.BlockNumber)

	apiResp.ApiRespOK(resp)
	return nil
}
