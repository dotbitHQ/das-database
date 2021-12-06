package http_server

import (
	"context"
	"das_database/block_parser"
	"das_database/chain/chain_ckb"
	"das_database/dao"
	"das_database/http_server/handle"
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
	Address   string
	DbDao     *dao.DbDao
	Ctx       context.Context
	CkbClient *chain_ckb.Client
	Bp        *block_parser.BlockParser
}

func Initialize(p HttpServerParams) (*HttpServer, error) {
	hs := HttpServer{
		address: p.Address,
		engine:  gin.New(),
		h: handle.Initialize(handle.HttpHandleParams{
			DbDao:     p.DbDao,
			CkbClient: p.CkbClient,
			Ctx:       p.Ctx,
			Bp:        p.Bp,
		}),
		ctx: p.Ctx,
	}
	return &hs, nil
}

func (h *HttpServer) Run() {
	v1 := h.engine.Group("v1")
	{
		v1.POST("/latest/block/number", h.h.IsLatestBlockNumber) // check if the newest height
		v1.POST("/parser/transaction", h.h.ParserTransaction)
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
