package handle

import (
	"das_database/http_server/api_code"
	"das_database/prometheus"
	"fmt"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
	"time"
)

func (h *HttpHandle) JasonRpcHandle(ctx *gin.Context) {
	var (
		req      api_code.JsonRequest
		resp     api_code.JsonResponse
		apiResp  http_api.ApiResp
		clientIp = GetClientIp(ctx)
	)
	resp.Result = &apiResp

	start := time.Now()

	defer func() {
		prometheus.Tools.Metrics.Api().WithLabelValues(req.Method, "200", fmt.Sprint(apiResp.ErrNo), apiResp.ErrMsg).Observe(time.Since(start).Seconds())
	}()

	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		log.Error("ShouldBindJSON err:", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, resp)
		return
	}

	resp.ID, resp.JsonRpc = req.ID, req.JsonRpc
	log.Info("JasonRpcHandle:", req.Method, clientIp, toolib.JsonString(req))

	switch req.Method {
	case api_code.MethodSnapshotPermissionsInfo:
		h.JsonRpcSnapshotPermissionsInfo(req.Params, &apiResp)
	case api_code.MethodSnapshotAddressAccounts:
		h.JsonRpcSnapshotAddressAccounts(req.Params, &apiResp)
	case api_code.MethodSnapshotRegisterHistory:
		h.JsonRpcSnapshotRegisterHistory(req.Params, &apiResp)
	case api_code.MethodSnapshotDidList:
		h.JsonRpcSnapshotDidList(req.Params, &apiResp)
	default:
		log.Error("method not exist:", req.Method)
		apiResp.ApiRespErr(api_code.ApiCodeMethodNotExist, fmt.Sprintf("method [%s] not exits", req.Method))
	}
	ctx.JSON(http.StatusOK, resp)
	return
}
