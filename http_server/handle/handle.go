package handle

import (
	"context"
	"das_database/block_parser"
	"das_database/dao"
	"das_database/http_server/api_code"
	"fmt"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/http_api/logger"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"net/http"
)

var (
	log = logger.NewLogger("http_handle", logger.LevelDebug)
)

type HttpHandle struct {
	ctx     context.Context
	dbDao   *dao.DbDao
	dasCore *core.DasCore
	bp      *block_parser.BlockParser
	red     *redis.Client
}

type HttpHandleParams struct {
	DbDao   *dao.DbDao
	DasCore *core.DasCore
	Ctx     context.Context
	Bp      *block_parser.BlockParser
	Red     *redis.Client
}

func Initialize(p HttpHandleParams) *HttpHandle {
	hh := HttpHandle{
		dbDao:   p.DbDao,
		dasCore: p.DasCore,
		ctx:     p.Ctx,
		bp:      p.Bp,
		red:     p.Red,
	}
	return &hh
}

func GetClientIp(ctx *gin.Context) string {
	return fmt.Sprintf("%v", ctx.Request.Header.Get("X-Real-IP"))
}

func (h *HttpHandle) IsLatestBlockNumber(ctx *gin.Context) {
	log.Info("IsLatestBlockNumber", GetClientIp(ctx))

	blockNumber, err := h.dasCore.Client().GetTipBlockNumber(h.ctx)
	if err != nil {
		log.Error("GetTipBlockNumber err: ", err.Error())
		ctx.JSON(http.StatusOK, api_code.ApiRespErr(api_code.ApiCodeBlockError, "search block number err"))
		return
	}

	ctx.JSON(http.StatusOK, api_code.ApiRespOKData(map[string]interface{}{
		"blockNumber":         blockNumber,
		"isLatestBlockNumber": block_parser.IsLatestBlockNumber,
	}))
}

type ParserTransactionData struct {
	TxHash string `json:"txHash"`
}

func (h *HttpHandle) ParserTransaction(ctx *gin.Context) {
	var transactionData ParserTransactionData
	if err := ctx.ShouldBindJSON(&transactionData); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error())
		ctx.JSON(http.StatusOK, api_code.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid"))
		return
	}

	log.Info("ParserTransaction", transactionData.TxHash, GetClientIp(ctx))

	tx, err := h.dasCore.Client().GetTransaction(h.ctx, types.HexToHash(transactionData.TxHash))
	if err != nil {
		log.Error("GetTransaction err: ", err.Error())
		ctx.JSON(http.StatusOK, api_code.ApiRespErr(api_code.ApiCodeBlockError, "search transaction err"))
		return
	}
	header, err := h.dasCore.Client().GetHeader(h.ctx, types.HexToHash(tx.TxStatus.BlockHash.Hex()))
	if err != nil {
		log.Error("GetHeader err: ", err.Error())
		ctx.JSON(http.StatusOK, api_code.ApiRespErr(api_code.ApiCodeBlockError, "search header err"))
		return
	}

	builder, err := witness.ActionDataBuilderFromTx(tx.Transaction)
	if err != nil {
		log.Error("ActionDataBuilderFromTx err: ", err.Error())
		ctx.JSON(http.StatusOK, api_code.ApiRespErr(api_code.ApiCodeBlockError, "builder from tx err"))
		return
	}

	if handle, ok := h.bp.GetMapTransactionHandle(builder.Action); ok {
		resp := handle(block_parser.FuncTransactionHandleReq{
			DbDao:          h.dbDao,
			Tx:             tx.Transaction,
			TxHash:         transactionData.TxHash,
			BlockNumber:    header.Number,
			BlockTimestamp: header.Timestamp,
			Action:         builder.Action,
		})
		if resp.Err != nil {
			log.Error("action handle resp:", builder.Action, header.Number, transactionData, resp.Err.Error())
			ctx.JSON(http.StatusOK, api_code.ApiRespErr(api_code.ApiCodeBlockError, "action handle err"))
			return
		}
	}

	ctx.JSON(http.StatusOK, api_code.ApiRespOK())
}
