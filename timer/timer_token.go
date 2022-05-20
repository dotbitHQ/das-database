package timer

import (
	"das_database/dao"
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
	list, err := p.dbDao.SearchTokenPriceInfoList()
	if err != nil {
		log.Errorf("doUpdateTokenMap SearchTokenPriceInfoList err:%s", err.Error())
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
	if list, err := p.dbDao.SearchTokenPriceInfoList(); err != nil {
		log.Error("updateTokenPriceInfoList SearchTokenPriceInfoList err:", err.Error())
	} else {
		for _, v := range list {
			geckoIds = append(geckoIds, v.GeckoId)
		}
	}

	if list, err := GetTokenPriceNew(geckoIds); err != nil {
		log.Error("updateTokenPriceInfoList GetTokenPrice err:", err.Error())
	} else {
		var tokenList []dao.TableTokenPriceInfo
		for _, v := range list {
			if v.Price.Cmp(decimal.Zero) == 1 {
				tokenList = append(tokenList, dao.TableTokenPriceInfo{
					GeckoId:       strings.ToLower(v.Id),
					Price:         v.Price,
					Change24h:     v.Change24h,
					Vol24h:        v.Vol24h,
					MarketCap:     v.MarketCap,
					LastUpdatedAt: v.LastUpdatedAt,
				})
			}
		}
		if err := p.dbDao.UpdateTokenPriceInfoList(tokenList); err != nil {
			log.Error("updateTokenPriceInfoList UpdateTokenPriceInfoList err:", err.Error())
		}
	}
}

func (p *ParserTimer) updateUSDRate() {
	ids := []string{
		"ethereum",
	}
	list, err := GetTokenPrice(ids)
	if err != nil {
		log.Errorf("updateUSDRate GetTokenPrice err:%s", err.Error())
	}
	if len(list) > 0 && list[0].Cny.Cmp(decimal.Zero) == 1 && list[0].Price.Cmp(decimal.Zero) == 1 {
		dec := list[0].Price.DivRound(list[0].Cny, 4)
		if err = p.dbDao.UpdateCNYToUSDRate([]string{"wx_cny"}, dec); err != nil {
			log.Errorf("updateUSDRate UpdateCNYToUSDRate err:%s", err)
		}
	}
}
