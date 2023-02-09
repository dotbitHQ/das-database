package http_server

import (
	"context"
	"das_database/block_parser"
	"das_database/dao"
	"das_database/http_server/handle"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/mylog"
	"net/http"
)

var (
	log = mylog.NewLogger("http_server", mylog.LevelDebug)
)

type HttpServer struct {
	address string
	engine  *gin.Engine
	h       *handle.HttpHandle
	srv     *http.Server
	ctx     context.Context
}

type HttpServerParams struct {
	Address string
	DbDao   *dao.DbDao
	Ctx     context.Context
	DasCore *core.DasCore
	Bp      *block_parser.BlockParser
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
		}),
		ctx: p.Ctx,
	}
	return &hs, nil
}

func (h *HttpServer) Run() {
	v1 := h.engine.Group("v1")
	{
		v1.POST("/latest/block/number", h.h.IsLatestBlockNumber) // check if the newest height
		//v1.POST("/parser/transaction", h.h.ParserTransaction)
		v1.POST("/snapshot/permissions/info", h.h.SnapshotPermissionsInfo)
		v1.POST("/snapshot/address/accounts", h.h.SnapshotAddressAccounts)
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
