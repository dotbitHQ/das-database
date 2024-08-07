package handle

import (
	"das_database/dao"
	"das_database/http_server/api_code"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

type ReqSnapshotAddressAccounts struct {
	core.ChainTypeAddress
	BlockNumber uint64       `json:"block_number"`
	RoleType    dao.RoleType `json:"role_type"`
	Pagination
}

type RespSnapshotAddressAccounts struct {
	Total    int64                    `json:"total"`
	Accounts []SnapshotAddressAccount `json:"accounts"`
}

type SnapshotAddressAccount struct {
	Account string `json:"account"`
}

func (h *HttpHandle) JsonRpcSnapshotAddressAccounts(p json.RawMessage, apiResp *http_api.ApiResp) {
	var req []ReqSnapshotAddressAccounts
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

	if err = h.doSnapshotAddressAccounts(&req[0], apiResp); err != nil {
		log.Error("doSnapshotAddressAccounts err:", err.Error())
	}
}

func (h *HttpHandle) SnapshotAddressAccounts(ctx *gin.Context) {
	var (
		funcName = "SnapshotAddressAccounts"
		req      ReqSnapshotAddressAccounts
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

	if err = h.doSnapshotAddressAccounts(&req, &apiResp); err != nil {
		log.Error("doSnapshotAddressAccounts err:", err.Error(), funcName)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doSnapshotAddressAccounts(req *ReqSnapshotAddressAccounts, apiResp *http_api.ApiResp) error {
	var resp RespSnapshotAddressAccounts
	resp.Accounts = make([]SnapshotAddressAccount, 0)

	addrHex, err := req.ChainTypeAddress.FormatChainTypeAddress(h.dasCore.NetType(), false)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "Invalid key info parameter")
		return nil
	}

	// check
	//ok, err := h.checkSnapshotProgress(req.BlockNumber, apiResp)
	//if apiResp.ErrNo != api_code.ApiCodeSuccess {
	//	return err
	//} else if !ok {
	//	apiResp.ApiRespErr(api_code.ApiCodeSnapshotBehindSchedule, "Snapshot behind schedule")
	//	return nil
	//}

	// snapshot
	list, err := h.dbDao.GetSnapshotAddressAccounts(addrHex.AddressHex, req.RoleType, req.BlockNumber, req.GetLimit(), req.GetOffset())
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "Failed to query historical account holding")
		return fmt.Errorf("GetSnapshotAddressAccounts err: %s", err.Error())
	}
	for _, v := range list {
		resp.Accounts = append(resp.Accounts, SnapshotAddressAccount{Account: v.Account})
	}

	total, err := h.dbDao.GetSnapshotAddressAccountsTotal(addrHex.AddressHex, req.RoleType, req.BlockNumber)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "Failed to query historical account holding")
		return fmt.Errorf("GetSnapshotAddressAccountsTotal err: %s", err.Error())
	}
	resp.Total = total

	apiResp.ApiRespOK(resp)
	return nil
}
