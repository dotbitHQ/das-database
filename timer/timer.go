package timer

import (
	"context"
	"das_database/dao"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/dotbitHQ/das-lib/http_api/logger"
	"sync"
	"time"
)

var log = logger.NewLogger("timer", logger.LevelDebug)

type ParserTimer struct {
	DbDao   *dao.DbDao
	Ctx     context.Context
	Wg      *sync.WaitGroup
	DasCore *core.DasCore
}

func (p *ParserTimer) RunUpdateTokenPrice() {
	p.updateTokenMap()

	tickerToken := time.NewTicker(time.Second * 180)
	//tickerUSD := time.NewTicker(time.Second * 300)

	p.Wg.Add(1)
	go func() {
		defer http_api.RecoverPanic()
		for {
			select {
			case <-tickerToken.C:
				log.Debug("RunUpdateTokenPriceList start ...", time.Now().Format("2006-01-02 15:04:05"))
				p.updateTokenPriceInfoList()
				p.updateTokenMap()
				log.Debug("RunUpdateTokenPriceList end ...", time.Now().Format("2006-01-02 15:04:05"))
			//case <-tickerUSD.C:
			//	log.Info("RunUpdateUSDRate start ...", time.Now().Format("2006-01-02 15:04:05"))
			//	p.updateUSDRate()
			//	log.Info("RunUpdateUSDRate end ...", time.Now().Format("2006-01-02 15:04:05"))
			case <-p.Ctx.Done():
				p.Wg.Done()
				return
			}
		}
	}()
}
