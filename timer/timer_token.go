package timer

import (
	"das_database/dao"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"strings"
	"sync"
)

const (
	TokenIdCkb  = "ckb_ckb"
	TokenIdEth  = "eth_eth"
	TokenIdTron = "tron_trx"
)

var (
	tokenLock sync.RWMutex
	mapToken  map[string]dao.TableTokenPriceInfo
)

func (p *ParserTimer) updateTokenMap() {
	list, err := p.DbDao.SearchTokenPriceInfoList()
	if err != nil {
		log.Errorf("SearchTokenPriceInfoList err:%s", err.Error())
	}

	tokenLock.Lock()
	defer tokenLock.Unlock()
	mapToken = make(map[string]dao.TableTokenPriceInfo, 0)
	for i, v := range list {
		mapToken[v.TokenId] = list[i]
	}
}

func GetTokenPriceInfo(tokenId string) dao.TableTokenPriceInfo {
	tokenLock.RLock()
	defer tokenLock.RUnlock()
	t, _ := mapToken[tokenId]
	return t
}

func GetTokenPriceInfoList() map[string]dao.TableTokenPriceInfo {
	tokenLock.Lock()
	defer tokenLock.Unlock()
	return mapToken
}

func (p *ParserTimer) updateTokenPriceInfoList() {
	var geckoIds []string
	if list, err := p.DbDao.SearchTokenPriceInfoList(); err != nil {
		log.Error("SearchTokenPriceInfoList err:", err.Error())
	} else {
		for _, v := range list {
			geckoIds = append(geckoIds, v.GeckoId)
		}
	}

	if list, err := GetTokenPriceNew(geckoIds); err != nil {
		log.Error("GetTokenPriceNew err:", err.Error())
	} else {
		var tokenList []dao.TableTokenPriceInfo
		for _, v := range list {
			tokenList = append(tokenList, dao.TableTokenPriceInfo{
				GeckoId:       strings.ToLower(v.Id),
				Price:         v.Price,
				Change24h:     v.Change24h,
				Vol24h:        v.Vol24h,
				MarketCap:     v.MarketCap,
				LastUpdatedAt: v.LastUpdatedAt,
			})
		}
		if err := p.DbDao.UpdateTokenPriceInfoList(tokenList); err != nil {
			log.Error("UpdateTokenPriceInfoList err:", err.Error())
		}
	}
}

//
func (p *ParserTimer) updateUSDRate() {
	rate, err := GetCnyRate()
	if err != nil {
		log.Error("GetCnyRate err: ", err.Error())
	}
	log.Info("updateUSDRate:", toolib.JsonString(&rate))
	if rate != nil && rate.Value > 0 {
		dec := decimal.NewFromInt(1).DivRound(decimal.NewFromFloat(rate.Value), 4)
		if err = p.DbDao.UpdateCNYToUSDRate([]string{"wx_cny"}, dec); err != nil {
			log.Errorf("UpdateCNYToUSDRate err:%s", err)
		}
	}
}
