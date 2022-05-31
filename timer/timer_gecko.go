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

//
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

//
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
	url := fmt.Sprintf("https://api1.binance.com/api/v3/ticker/price?symbols=[%s]", symbolStr)
	fmt.Println(url)

	var res []TokenPriceNew
	resp, body, errs := gorequest.New().Timeout(time.Second*30).Get(url).Retry(3, time.Second*2).End()
	if len(errs) > 0 {
		return nil, fmt.Errorf("GetTokenPrice api err:%v", errs)
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetTokenPrice api status code:%d", resp.StatusCode)
	}
	fmt.Println(body)
	//fmt.Println(res)
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
	var rate = Rate{"人民币", "CNY", "¥", 0, 1.5}
	url := "https://sp0.baidu.com/8aQDcjqpAAV3otqbppnN2DJv/api.php?query=1美元等于多少人民币&co=&resource_id=5293&t=1587039033404&cardId=5293&ie=utf8&oe=gbk&cb=op_aladdin_callback&format=json&tn=baidu&alr=1&cb=jQuery1102038387806309445316_1587037695932&_=1587037695933"
	res, body, errs := gorequest.New().Timeout(10 * time.Second).Get(url).End()
	if errs != nil {
		return nil, fmt.Errorf("http req err: %v", errs)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http req err: %v", res.StatusCode)
	}

	indexI := strings.Index(body, "{")
	indexJ := strings.LastIndex(body, ")")
	if indexI > 0 && indexJ > 0 {
		body = body[indexI:indexJ]
	}
	var result ResultData
	if err := json.Unmarshal([]byte(body), &result); err != nil {
		return nil, fmt.Errorf("json.Unmarshal err: %s", err.Error())
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
