package timer

import (
	"context"
	"das_database/chain/chain_ckb"
	"das_database/dao"
	"github.com/scorpiotzh/mylog"
	"sync"
	"time"
)

var log = mylog.NewLogger("timer", mylog.LevelDebug)

type ParserTimer struct {
	dbDao     *dao.DbDao
	ctx       context.Context
	wg        *sync.WaitGroup
	ckbClient *chain_ckb.Client
}

type ParserTimerParam struct {
	DbDao     *dao.DbDao
	Ctx       context.Context
	Wg        *sync.WaitGroup
	CkbClient *chain_ckb.Client
}

func NewParserTimer(p ParserTimerParam) *ParserTimer {
	var t ParserTimer
	t.dbDao = p.DbDao
	t.ctx = p.Ctx
	t.wg = p.Wg
	t.ckbClient = p.CkbClient
	return &t
}

func (p *ParserTimer) RunUpdateTokenPrice() {
	p.updateTokenMap()

	tickerToken := time.NewTicker(time.Second * 180)
	tickerUSD := time.NewTicker(time.Second * 300)

	p.wg.Add(1)
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
			case <-p.ctx.Done():
				p.wg.Done()
				return
			}
		}
	}()
}

func (p *ParserTimer) RunDailyRegister() {
	tickerRegister := time.NewTicker(time.Hour * 24)
	if registerInfo, err := p.dbDao.GetLastRegisterInfo(); err != nil {
		log.Error("GetLastRegisterInfo err:", err.Error())
	} else if registerInfo.Id == 0 {
		now := time.Now()
		origin := time.Date(2021, 7, 22, 0, 0, 0, 0, time.Local)
		nowUnix := now.Unix()
		originUnix := origin.Unix()

		days := (nowUnix - originUnix) / (3600 * 24)
		for i := int64(0); i < days; i++ {
			registeredAt := origin.Format("2006-01-02")
			p.dailyRegister(registeredAt)
			origin = origin.Add(time.Hour * 24)
		}
	}

	p.wg.Add(1)
	go func() {
		for {
			select {
			case <-tickerRegister.C:
				log.Info("RunDailyRegister start ...", time.Now().Format("2006-01-02 15:04:05"))
				registeredAt := time.Now().Add(-time.Hour * 24).Format("2006-01-02")
				p.dailyRegister(registeredAt)
				log.Info("RunDailyRegister end ...", time.Now().Format("2006-01-02 15:04:05"))
			case <-p.ctx.Done():
				p.wg.Done()
				return
			}
		}
	}()
}
