package timer

import (
	"encoding/json"
	"fmt"
	"github.com/parnurzeal/gorequest"
	"github.com/shopspring/decimal"
	"net/http"
	"strconv"
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

// https://binance-docs.github.io/apidocs/spot/cn/#8ff46b58de
func GetTokenPriceNew(ids []string) ([]GeckoTokenInfo, error) {
	var symbols []string
	symbols = append(symbols, "BTCUSDT", "CKBUSDT", "ETHUSDT", "TRXUSDT", "BNBUSDT", "DOGEUSDT", "POLUSDT")

	symbolStr := ""
	for _, v := range symbols {
		symbolStr += "%22" + v + "%22,"
	}
	symbolStr = strings.Trim(symbolStr, ",")
	url := fmt.Sprintf("https://api1.binance.com/api/v3/ticker/price?symbols=[%s]", symbolStr)
	log.Info("GetTokenPriceNew:", url)

	var res []TokenPriceNew
	resp, body, errs := gorequest.New().Timeout(time.Second*30).Get(url).Retry(3, time.Second*2).End()
	if len(errs) > 0 {
		return nil, fmt.Errorf("GetTokenPrice api err:%v", errs)
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetTokenPrice api status code:%d", resp.StatusCode)
	}
	log.Info("GetTokenPriceNew:", body)
	if err := json.Unmarshal([]byte(body), &res); err != nil {
		return nil, err
	}
	list := TokenPriceNewToList(res)

	return list, nil
}

type TokenPriceNew struct {
	Symbol string          `json:"symbol"`
	Price  decimal.Decimal `json:"price"`
}

var TokenIdMap = map[string][]string{
	"CKBUSDT":   {"ckb_ckb", "ckb_ccc"},
	"BTCUSDT":   {"btc_btc"},
	"ETHUSDT":   {"eth_eth"},
	"BNBUSDT":   {"bsc_bnb"},
	"TRXUSDT":   {"tron_trx"},
	"MATICUSDT": {"polygon_matic"},
	"DOGEUSDT":  {"doge_doge"},
	"POLUSDT":   {"polygon_pol"},
}

func TokenPriceNewToList(res []TokenPriceNew) []GeckoTokenInfo {
	list := make([]GeckoTokenInfo, 0)
	for _, v := range res {
		now := time.Now().Unix()
		for _, tokenId := range TokenIdMap[v.Symbol] {
			list = append(list, GeckoTokenInfo{
				Id:            tokenId,
				Price:         v.Price,
				LastUpdatedAt: now,
			})
		}
	}
	return list
}

type Rate struct {
	Title      string  `gorm:"-" json:"-"`
	Name       string  `gorm:"column:name" json:"name"`
	Symbol     string  `gorm:"column:symbol" json:"symbol"`
	Value      float64 `gorm:"column:value" json:"value"`
	CheckValue float64 `gorm:"-" json:"-"`
}

type ResultData struct {
	Result []struct {
		DisplayData struct {
			ResultData struct {
				TplData struct {
					Value string `json:"money2_num"`
				} `json:"tplData"`
			} `json:"resultData"`
		} `json:"DisplayData"`
	} `json:"Result"`
}

func GetCnyRate() (*Rate, error) {
	var rate = Rate{"", "CNY", "¥", 0, 1.5}
	url := "https://sp1.baidu.com/8aQDcjqpAAV3otqbppnN2DJv/api.php?query=1%E7%BE%8E%E5%85%83%E7%AD%89%E4%BA%8E%E5%A4%9A%E5%B0%91%E4%BA%BA%E6%B0%91%E5%B8%81&resource_id=5293&alr=1"

	var result ResultData
	res, _, errs := gorequest.New().Timeout(10 * time.Second).Get(url).EndStruct(&result)
	if errs != nil {
		return nil, fmt.Errorf("http req err: %v", errs)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http req err: %v", res.StatusCode)
	}
	if len(result.Result) > 0 {
		valueStr := strings.Replace(result.Result[0].DisplayData.ResultData.TplData.Value, ",", "", -1)
		if value, err := strconv.ParseFloat(valueStr, 64); err != nil {
			return nil, fmt.Errorf("json.Unmarshal err: %s", err.Error())
		} else {
			rate.Value = value
		}
	}
	return &rate, nil
}
