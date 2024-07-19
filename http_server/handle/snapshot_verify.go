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

type ReqSnapshotVerify struct {
	core.ChainTypeAddress
	Message            string `json:"message"`
	Signature          string `json:"signature"`
	PasskeySignAddress string `json:"passkey_sign_address"`
}

type RespSnapshotVerify struct {
	Verified bool `json:"verified"`
}

func (h *HttpHandle) JsonRpcSnapshotVerify(p json.RawMessage, apiResp *http_api.ApiResp) {
	var req []ReqSnapshotVerify
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

	if err = h.doSnapshotVerify(&req[0], apiResp); err != nil {
		log.Error("doSnapshotVerify err:", err.Error())
	}
}

func (h *HttpHandle) SnapshotVerify(ctx *gin.Context) {
	var (
		funcName = "SnapshotVerify"
		req      ReqSnapshotVerify
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

	if err = h.doSnapshotVerify(&req, &apiResp); err != nil {
		log.Error("doSnapshotVerify err:", err.Error(), funcName)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doSnapshotVerify(req *ReqSnapshotVerify, apiResp *http_api.ApiResp) error {
	var resp RespSnapshotVerify

	addrHex, err := req.ChainTypeAddress.FormatChainTypeAddress(h.dasCore.NetType(), false)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "Invalid key info")
		return nil
	}
	if addrHex.DasAlgorithmId == common.DasAlgorithmIdAnyLock {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "Invalid key info")
		return nil
	}

	signature := req.Signature
	addressHex := addrHex.AddressHex
	if addrHex.DasAlgorithmId == common.DasAlgorithmIdWebauthn {
		signAddressHex, err := h.dasCore.Daf().NormalToHex(core.DasAddressNormal{
			ChainType:     common.ChainTypeWebauthn,
			AddressNormal: req.PasskeySignAddress,
		})
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "Invalid key info")
			return nil
		}
		addressHex = signAddressHex.AddressHex
		loginAddrHex := core.DasAddressHex{
			DasAlgorithmId:    common.DasAlgorithmIdWebauthn,
			DasSubAlgorithmId: common.DasWebauthnSubAlgorithmIdES256,
			AddressHex:        addrHex.AddressHex,
			AddressPayload:    common.Hex2Bytes(addrHex.AddressHex),
			ChainType:         common.ChainTypeWebauthn,
		}
		signAddrHex := core.DasAddressHex{
			DasAlgorithmId:    common.DasAlgorithmIdWebauthn,
			DasSubAlgorithmId: common.DasWebauthnSubAlgorithmIdES256,
			AddressHex:        signAddressHex.AddressHex,
			AddressPayload:    common.Hex2Bytes(signAddressHex.AddressHex),
			ChainType:         common.ChainTypeWebauthn,
		}
		idx, err := h.dasCore.GetIdxOfKeylist(loginAddrHex, signAddrHex)
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "Invalid key info")
			return fmt.Errorf("GetIdxOfKeylist err: %s", err.Error())
		}
		if idx == -1 {
			apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "Invalid key info")
			return fmt.Errorf("GetIdxOfKeylist idx: %d", idx)
		}
		h.dasCore.AddPkIndexForSignMsg(&signature, idx)
	}

	resp.Verified, _, err = http_api.VerifySignature(addrHex.DasAlgorithmId, req.Message, signature, addressHex)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeSignError, "VerifySignature err")
		return fmt.Errorf("VerifySignature err: %s", err.Error())
	}

	apiResp.ApiRespOK(resp)
	return nil
}
