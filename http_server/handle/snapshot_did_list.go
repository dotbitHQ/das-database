package handle

import (
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

type ReqSnapshotDidList struct {
	core.ChainTypeAddress
	BlockNumber   uint64 `json:"block_number"`
	AccountLength uint64 `json:"account_length"`
}

type RespSnapshotDidList struct {
	Total    int           `json:"total"`
	Accounts []SnapshotDid `json:"accounts"`
}

type SnapshotDid struct {
	Account string `json:"account"`
}

func (h *HttpHandle) JsonRpcSnapshotDidList(p json.RawMessage, apiResp *http_api.ApiResp) {
	var req []ReqSnapshotDidList
	err := json.Unmarshal(p, &req)
	if err != nil {
		log.Error("json.Unmarshal err:", err.Error())
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		return
	}
	if len(req) != 1 {
		log.Error("len(req) is :", len(req))
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		return
	}

	if err = h.doSnapshotDidList(&req[0], apiResp); err != nil {
		log.Error("doSnapshotDidList err:", err.Error())
	}
}

func (h *HttpHandle) SnapshotDidList(ctx *gin.Context) {
	var (
		funcName = "SnapshotDidList"
		req      ReqSnapshotDidList
		apiResp  http_api.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName)
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, toolib.JsonString(req))

	if err = h.doSnapshotDidList(&req, &apiResp); err != nil {
		log.Error("doSnapshotDidList err:", err.Error(), funcName)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doSnapshotDidList(req *ReqSnapshotDidList, apiResp *http_api.ApiResp) error {
	var resp RespSnapshotDidList
	resp.Accounts = make([]SnapshotDid, 0)

	addrHex, err := req.ChainTypeAddress.FormatChainTypeAddress(h.dasCore.NetType(), false)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "Invalid key info parameter")
		return nil
	}
	log.Info("doSnapshotDidList:", addrHex.AddressHex, addrHex.DasAlgorithmId)

	// snapshot
	list, err := h.dbDao.GetSnapshotDidList(addrHex.AddressHex, req.BlockNumber, req.AccountLength)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "Failed to query did list")
		return fmt.Errorf("GetSnapshotDidList err: %s", err.Error())
	}
	for _, v := range list {
		_, accLen, err := common.GetDotBitAccountLength(v.Account)
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeError500, err.Error())
			return fmt.Errorf("GetDotBitAccountLength err: %s", err.Error())
		}
		if uint64(accLen) != req.AccountLength {
			continue
		}
		resp.Accounts = append(resp.Accounts, SnapshotDid{Account: v.Account})
	}
	resp.Total = len(list)

	apiResp.ApiRespOK(resp)
	return nil
}
