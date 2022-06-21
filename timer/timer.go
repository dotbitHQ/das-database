package timer

import (
	"context"
	"das_database/dao"
	"github.com/scorpiotzh/mylog"
	"sync"
	"time"
)

var log = mylog.NewLogger("timer", mylog.LevelDebug)

type ParserTimer struct {
	DbDao *dao.DbDao
	Ctx   context.Context
	Wg    *sync.WaitGroup
}

func (p *ParserTimer) RunUpdateTokenPrice() {
	p.updateTokenMap()

	tickerToken := time.NewTicker(time.Second * 180)
	tickerUSD := time.NewTicker(time.Second * 300)

	p.Wg.Add(1)
	go func() {
		for {
			select {
			case <-tickerToken.C:
				log.Info("RunUpdateTokenPriceList start ...", time.Now().Format("2006-01-02 15:04:05"))
				p.updateTokenPriceInfoList()
				p.updateTokenMap()
				log.Info("RunUpdateTokenPriceList end ...", time.Now().Format("2006-01-02 15:04:05"))
			case <-tickerUSD.C:
				log.Info("RunUpdateUSDRate start ...", time.Now().Format("2006-01-02 15:04:05"))
				p.updateUSDRate()
				log.Info("RunUpdateUSDRate end ...", time.Now().Format("2006-01-02 15:04:05"))
			case <-p.Ctx.Done():
				p.Wg.Done()
				return
			}
		}
	}()
}
