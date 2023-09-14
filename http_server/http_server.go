package http_server

import (
	"context"
	"das_database/block_parser"
	"das_database/config"
	"das_database/dao"
	"das_database/http_server/api_code"
	"das_database/http_server/handle"
	"encoding/json"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/http_api/logger"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/scorpiotzh/toolib"
	"net/http"
	"time"
)

var (
	log = logger.NewLogger("http_server", logger.LevelDebug)
)

type HttpServer struct {
	address string
	engine  *gin.Engine
	h       *handle.HttpHandle
	srv     *http.Server
	ctx     context.Context
	red     *redis.Client
}

type HttpServerParams struct {
	Address string
	DbDao   *dao.DbDao
	Ctx     context.Context
	DasCore *core.DasCore
	Bp      *block_parser.BlockParser
	Red     *redis.Client
}

func Initialize(p HttpServerParams) (*HttpServer, error) {
	hs := HttpServer{
		address: p.Address,
		engine:  gin.New(),
		h: handle.Initialize(handle.HttpHandleParams{
			DbDao:   p.DbDao,
			DasCore: p.DasCore,
			Ctx:     p.Ctx,
			Bp:      p.Bp,
			Red:     p.Red,
		}),
		ctx: p.Ctx,
		red: p.Red,
	}
	return &hs, nil
}

func (h *HttpServer) Run() {
	shortDataTime, lockTime, shortExpireTime := time.Minute, time.Second*30, time.Second*5
	cacheHandle := toolib.MiddlewareCacheByRedis(h.red, false, shortDataTime, lockTime, shortExpireTime, respHandle)

	if len(config.Cfg.Origins) > 0 {
		toolib.AllowOriginList = append(toolib.AllowOriginList, config.Cfg.Origins...)
	}

	h.engine.Use(toolib.MiddlewareCors())
	h.engine.POST("", cacheHandle, h.h.JasonRpcHandle)
	h.engine.Use(sentrygin.New(sentrygin.Options{
		Repanic: true,
	}))
	v1 := h.engine.Group("v1")
	{
		v1.POST("/latest/block/number", h.h.IsLatestBlockNumber)
		v1.POST("/snapshot/progress", cacheHandle, h.h.SnapshotProgress)
		v1.POST("/snapshot/permissions/info", cacheHandle, h.h.SnapshotPermissionsInfo)
		v1.POST("/snapshot/address/accounts", cacheHandle, h.h.SnapshotAddressAccounts)
		v1.POST("/snapshot/register/history", cacheHandle, h.h.SnapshotRegisterHistory)
	}

	h.srv = &http.Server{
		Addr:    h.address,
		Handler: h.engine,
	}
	go func() {
		if err := h.srv.ListenAndServe(); err != nil {
			log.Error("http_server run err:", err)
		}
	}()
}

func (h *HttpServer) Shutdown() {
	log.Warn("http server Shutdown ... ")
	if h.srv != nil {
		if err := h.srv.Shutdown(h.ctx); err != nil {
			log.Error("http server Shutdown err:", err.Error())
		}
	}
}

func respHandle(c *gin.Context, res string, err error) {
	if err != nil {
		log.Error("respHandle err:", err.Error())
		c.AbortWithStatusJSON(http.StatusOK, api_code.ApiRespErr(api_code.ApiCodeError500, err.Error()))
	} else if res != "" {
		var respMap map[string]interface{}
		_ = json.Unmarshal([]byte(res), &respMap)
		c.AbortWithStatusJSON(http.StatusOK, respMap)
	}
}
