package handle

import (
	"das_database/dao"
	"das_database/http_server/api_code"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/http_api"
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

func (h *HttpHandle) JsonRpcSnapshotPermissionsInfo(p json.RawMessage, apiResp *http_api.ApiResp) {
	var req []ReqSnapshotPermissionsInfo
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

	if err = h.doSnapshotPermissionsInfo(&req[0], apiResp); err != nil {
		log.Error("doSnapshotPermissionsInfo err:", err.Error())
	}
}

func (h *HttpHandle) SnapshotPermissionsInfo(ctx *gin.Context) {
	var (
		funcName = "SnapshotPermissionsInfo"
		req      ReqSnapshotPermissionsInfo
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

	if err = h.doSnapshotPermissionsInfo(&req, &apiResp); err != nil {
		log.Error("doSnapshotPermissionsInfo err:", err.Error(), funcName)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) checkSnapshotProgress(blockNumber uint64, apiResp *http_api.ApiResp) (bool, error) {
	txS, err := h.dbDao.GetTxSnapshotSchedule()
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "Failed to query snapshot progress")
		return false, fmt.Errorf("GetTxSnapshotSchedule err: %s", err.Error())
	} else if txS.Id > 0 {
		if txS.BlockNumber >= blockNumber {
			return true, nil
		}
	}
	return false, nil
}

func (h *HttpHandle) doSnapshotPermissionsInfo(req *ReqSnapshotPermissionsInfo, apiResp *http_api.ApiResp) error {
	var resp RespSnapshotPermissionsInfo

	if req.Account == "" || !strings.HasSuffix(req.Account, common.DasAccountSuffix) {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "Invalid account parameter")
		return nil
	}

	// check
	//ok, err := h.checkSnapshotProgress(req.BlockNumber, apiResp)
	//if apiResp.ErrNo != http_api.ApiCodeSuccess {
	//	return err
	//} else if !ok {
	//	apiResp.ApiRespErr(http_api.ApiCodeSnapshotBehindSchedule, "Snapshot behind schedule")
	//	return nil
	//}

	// snapshot
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))
	info, err := h.dbDao.GetSnapshotPermissionsInfo(accountId, req.BlockNumber)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "Failed to find account permission information")
		return fmt.Errorf("GetSnapshotPermissionsInfo err: %s", err.Error())
	}
	if info.Id == 0 {
		apiResp.ApiRespErr(api_code.ApiCodeAccountPermissionsDoNotExist, "Account permissions do not exist")
		return nil
	}
	if info.Status == dao.AccountStatusRecycle {
		apiResp.ApiRespErr(api_code.ApiCodeAccountHasBeenRecycled, "Account has been recycled")
		return nil
	}
	if info.Status == dao.AccountStatusOnLock {
		apiResp.ApiRespErr(api_code.ApiCodeAccountCrossChain, "Account cross-chain")
		return nil
	}

	// check expired or not
	builder, err := h.dasCore.ConfigCellDataBuilderByTypeArgsList(
		common.ConfigCellTypeArgsAccount,
	)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "Failed to get ExpirationGracePeriod")
		return fmt.Errorf("ConfigCellDataBuilderByTypeArgsList err: %s", err.Error())
	}
	expirationGracePeriod, _ := builder.ExpirationGracePeriod()

	expiredAt := info.ExpiredAt + uint64(expirationGracePeriod)
	block, err := h.dasCore.Client().GetBlockByNumber(h.ctx, req.BlockNumber)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "Failed to get block")
		return fmt.Errorf("GetBlockByNumber err: %s", err.Error())
	}
	blockTimestamp := block.Header.Timestamp / 1000
	log.Info("ExpirationGracePeriod:", expirationGracePeriod, info.ExpiredAt, expiredAt, blockTimestamp)
	if expiredAt < blockTimestamp {
		apiResp.ApiRespErr(api_code.ApiCodeAccountExpired, "Account expired")
		return nil
	}

	// sub account
	if count := strings.Count(info.Account, "."); count > 1 {
		acc, err := h.dbDao.GetAccountInfoByAccountId(info.AccountId)
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeDbError, "Failed to find account information")
			return fmt.Errorf("GetAccountInfoByAccountId err: %s", err.Error())
		}
		if acc.ParentAccountId != "" {
			recycleInfo, err := h.dbDao.GetRecycleInfo(acc.ParentAccountId, info.BlockNumber, req.BlockNumber)
			if err != nil {
				apiResp.ApiRespErr(http_api.ApiCodeDbError, "Failed to find parent account permissions information")
				return fmt.Errorf("GetRecycleInfo err: %s", err.Error())
			}
			if recycleInfo.Id > 0 {
				apiResp.ApiRespErr(api_code.ApiCodeAccountHasBeenRecycled, "Parent account has been recycled")
				return nil
			}
		}
	}

	resp.Account = req.Account
	resp.AccountId = accountId
	resp.BlockNumber = info.BlockNumber
	resp.OwnerAlgorithmId = info.OwnerAlgorithmId
	resp.ManagerAlgorithmId = info.ManagerAlgorithmId

	if info.OwnerAlgorithmId == common.DasAlgorithmIdAnyLock {
		resp.Owner = info.Owner
	} else {
		owner, err := h.dasCore.Daf().HexToNormal(core.DasAddressHex{
			DasAlgorithmId:    info.OwnerAlgorithmId,
			DasSubAlgorithmId: info.OwnerSubAid,
			AddressHex:        info.Owner,
			IsMulti:           false,
			ChainType:         info.OwnerAlgorithmId.ToChainType(),
		})
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeError500, "HexToNormal Err")
			return fmt.Errorf("HexToNormal err: %s", err.Error())
		}
		resp.Owner = owner.AddressNormal
	}

	if info.ManagerAlgorithmId == common.DasAlgorithmIdAnyLock {
		resp.Manager = info.Manager
	} else {
		manager, err := h.dasCore.Daf().HexToNormal(core.DasAddressHex{
			DasAlgorithmId:    info.ManagerAlgorithmId,
			DasSubAlgorithmId: info.ManagerSubAid,
			AddressHex:        info.Manager,
			IsMulti:           false,
			ChainType:         info.ManagerAlgorithmId.ToChainType(),
		})
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeError500, "HexToNormal Err")
			return fmt.Errorf("HexToNormal err: %s", err.Error())
		}
		resp.Manager = manager.AddressNormal
	}

	log.Info("doSnapshotPermissionsInfo:", resp)
	apiResp.ApiRespOK(resp)
	return nil
}
