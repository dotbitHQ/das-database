package handle

import (
	"das_database/dao"
	"das_database/http_server/api_code"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
	"strings"
)

type ReqSnapshotPermissionsInfo struct {
	Account     string `json:"account"`
	BlockNumber uint64 `json:"block_number"`
}

type RespSnapshotPermissionsInfo struct {
	Account            string                `json:"account"`
	AccountId          string                `json:"account_id"`
	BlockNumber        uint64                `json:"block_number"`
	Owner              string                `json:"owner"`
	OwnerAlgorithmId   common.DasAlgorithmId `json:"owner_algorithm_id"`
	Manager            string                `json:"manager"`
	ManagerAlgorithmId common.DasAlgorithmId `json:"manager_algorithm_id"`
}

func (h *HttpHandle) JsonRpcSnapshotPermissionsInfo(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqSnapshotPermissionsInfo
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

	if err = h.doSnapshotPermissionsInfo(&req[0], apiResp); err != nil {
		log.Error("doSnapshotPermissionsInfo err:", err.Error())
	}
}

func (h *HttpHandle) SnapshotPermissionsInfo(ctx *gin.Context) {
	var (
		funcName = "SnapshotPermissionsInfo"
		req      ReqSnapshotPermissionsInfo
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

	if err = h.doSnapshotPermissionsInfo(&req, &apiResp); err != nil {
		log.Error("doSnapshotPermissionsInfo err:", err.Error(), funcName)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doSnapshotPermissionsInfo(req *ReqSnapshotPermissionsInfo, apiResp *api_code.ApiResp) error {
	var resp RespSnapshotPermissionsInfo

	if req.Account == "" || !strings.HasSuffix(req.Account, common.DasAccountSuffix) {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "Invalid account parameter")
		return nil
	}
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))
	info, err := h.dbDao.GetSnapshotPermissionsInfo(accountId, req.BlockNumber)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "Failed to find account permission information")
		return fmt.Errorf("GetSnapshotPermissionsInfo err: %s", err.Error())
	}
	if info.Id == 0 {
		apiResp.ApiRespErr(api_code.ApiCodeAccountPermissionsDoNotExist, "Account permissions do not exist")
		return nil
	}
	if info.Status == dao.AccountStatusRecycle {
		apiResp.ApiRespErr(api_code.ApiCodeAccountHasBeenRevoked, "Account has been revoked")
		return nil
	}
	if info.Status == dao.AccountStatusOnLock {
		apiResp.ApiRespErr(api_code.ApiCodeAccountIsCrossChained, "Account is cross-chained")
		return nil
	}

	// sub account
	if count := strings.Count(info.Account, "."); count > 1 {
		acc, err := h.dbDao.GetAccountInfoByAccountId(info.AccountId)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeDbError, "Failed to find account information")
			return fmt.Errorf("GetAccountInfoByAccountId err: %s", err.Error())
		}
		if acc.ParentAccountId != "" {
			recycleInfo, err := h.dbDao.GetRecycleInfo(acc.ParentAccountId, info.BlockNumber, req.BlockNumber)
			if err != nil {
				apiResp.ApiRespErr(api_code.ApiCodeDbError, "Failed to find parent account permissions information")
				return fmt.Errorf("GetRecycleInfo err: %s", err.Error())
			}
			if recycleInfo.Id > 0 {
				apiResp.ApiRespErr(api_code.ApiCodeParentAccountIsRecycled, "Parent account is recycled")
				return nil
			}
		}
	}

	owner, err := h.dasCore.Daf().HexToNormal(core.DasAddressHex{
		DasAlgorithmId: info.OwnerAlgorithmId,
		AddressHex:     info.Owner,
		IsMulti:        false,
		ChainType:      info.OwnerAlgorithmId.ToChainType(),
	})
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "HexToNormal Err")
		return fmt.Errorf("HexToNormal err: %s", err.Error())
	}
	manager, err := h.dasCore.Daf().HexToNormal(core.DasAddressHex{
		DasAlgorithmId: info.ManagerAlgorithmId,
		AddressHex:     info.Manager,
		IsMulti:        false,
		ChainType:      info.ManagerAlgorithmId.ToChainType(),
	})
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "HexToNormal Err")
		return fmt.Errorf("HexToNormal err: %s", err.Error())
	}
	resp.Account = req.Account
	resp.AccountId = accountId
	resp.BlockNumber = req.BlockNumber
	resp.Owner = owner.AddressNormal
	resp.OwnerAlgorithmId = info.OwnerAlgorithmId
	resp.Manager = manager.AddressNormal
	resp.ManagerAlgorithmId = info.ManagerAlgorithmId

	apiResp.ApiRespOK(resp)
	return nil
}
