package timer

import (
	"encoding/json"
	"fmt"
	"github.com/parnurzeal/gorequest"
	"github.com/shopspring/decimal"
	"net/http"
	"strings"
	"time"
)

type GeckoTokenInfo struct {
	Id              string `json:"id"`
	Symbol          string `json:"symbol"`
	Name            string `json:"name"`
	ContractAddress string `json:"contract_address"`
	Links           struct {
		Homepage []string `json:"homepage"`
	} `json:"links"`
	Image struct {
		Large string `json:"large"`
	} `json:"image"`
	Cny           decimal.Decimal `json:"cny"`
	Price         decimal.Decimal `json:"price"`
	Change24h     decimal.Decimal `json:"change_24_h"`
	Vol24h        decimal.Decimal `json:"vol_24_h"`
	MarketCap     decimal.Decimal `json:"market_cap"`
	LastUpdatedAt int64           `json:"last_updated_at"`
}

func GetTokenPrice(ids []string) ([]GeckoTokenInfo, error) {
	idsStr := strings.Join(ids, ",")
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd,cny", idsStr)
	url = fmt.Sprintf("%s&include_market_cap=true&include_24hr_vol=true&include_24hr_change=true&include_last_updated_at=true", url)
	fmt.Println(url)
	resp, body, errs := gorequest.New().Timeout(time.Second*30).Get(url).Retry(3, time.Second*2).End()
	if len(errs) > 0 {
		return nil, fmt.Errorf("GetTokenPrice api err:%v", errs)
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetTokenPrice api status code:%d", resp.StatusCode)
	}
	//fmt.Println(body)
	var res map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(body), &res); err != nil {
		return nil, err
	}
	list := mapToList(res)
	return list, nil
}

func mapToList(res map[string]map[string]interface{}) []GeckoTokenInfo {
	list := make([]GeckoTokenInfo, 0)
	doParse := func(id, key string) decimal.Decimal {
		if v, ok := res[id][key].(float64); ok {
			return decimal.NewFromFloat(v)
		}
		return decimal.NewFromFloat(0)
	}
	for k, _ := range res {
		gti := GeckoTokenInfo{}
		gti.Id = k
		gti.Cny = doParse(k, "cny")
		gti.Price = doParse(k, "usd")
		gti.MarketCap = doParse(k, "usd_market_cap")
		gti.Vol24h = doParse(k, "usd_24h_vol")
		gti.Change24h = doParse(k, "usd_24h_change")
		if v, ok := res[k]["last_updated_at"].(float64); ok {
			gti.LastUpdatedAt = int64(v)
		}
		list = append(list, gti)
	}
	return list
}

func GetTokenPriceNew(ids []string) ([]GeckoTokenInfo, error) {
	var symbols []string
	symbols = append(symbols, "BTCUSDT", "CKBUSDT", "ETHUSDT", "TRXUSDT", "BNBUSDT", "MATICUSDT")

	symbolStr := ""
	for _, v := range symbols {
		symbolStr += "%22" + v + "%22,"
	}
	symbolStr = strings.Trim(symbolStr, ",")
	url := fmt.Sprintf("https://api1.binance.com/api/v3/ticker/price?symbols=\\[%s\\]", symbolStr)
	fmt.Println(url)

	var res []TokenPriceNew
	resp, _, errs := gorequest.New().Timeout(time.Second*30).Get(url).Retry(3, time.Second*2).EndStruct(&res)
	if len(errs) > 0 {
		return nil, fmt.Errorf("GetTokenPrice api err:%v", errs)
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetTokenPrice api status code:%d", resp.StatusCode)
	}
	//fmt.Println(body)
	//fmt.Println(res)
	list := TokenPriceNewToList(res)

	return list, nil
}

type TokenPriceNew struct {
	Symbol string          `json:"symbol"`
	Price  decimal.Decimal `json:"price"`
}

var TokenIdMap = map[string]string{
	"CKBUSDT":   "nervos-network",
	"BTCUSDT":   "bitcoin",
	"ETHUSDT":   "ethereum",
	"BNBUSDT":   "binancecoin",
	"TRXUSDT":   "tron",
	"MATICUSDT": "matic-network",
}

func TokenPriceNewToList(res []TokenPriceNew) []GeckoTokenInfo {
	list := make([]GeckoTokenInfo, 0)
	for _, v := range res {
		gti := GeckoTokenInfo{}
		gti.Id = TokenIdMap[v.Symbol]
		gti.Price = v.Price
		gti.LastUpdatedAt = time.Now().Unix()
		list = append(list, gti)
	}
	return list
}
