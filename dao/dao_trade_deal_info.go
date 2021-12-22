package dao

import (
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/shopspring/decimal"
	"time"
)

type TableTradeDealInfo struct {
	Id             uint64           `json:"id" gorm:"column:id;primary_key;AUTO_INCREMENT"`
	BlockNumber    uint64           `json:"block_number" gorm:"column:block_number"`
	Outpoint       string           `json:"outpoint" gorm:"column:outpoint"`
	AccountId      string           `json:"account_id" gorm:"account_id"`
	Account        string           `json:"account" gorm:"column:account"`
	DealType       int              `json:"deal_type" gorm:"column:deal_type"` // 0:transaction 1:auction 2:offer
	SellChainType  common.ChainType `json:"sell_chain_type" gorm:"column:sell_chain_type"`
	SellAddress    string           `json:"sell_address" gorm:"column:sell_address"`
	BuyChainType   common.ChainType `json:"buy_chain_type" gorm:"column:buy_chain_type"`
	BuyAddress     string           `json:"buy_address" gorm:"column:buy_address"`
	PriceCkb       uint64           `json:"price_ckb" gorm:"column:price_ckb"`
	PriceUsd       decimal.Decimal  `json:"price_usd" gorm:"column:price_usd"`
	BlockTimestamp uint64           `json:"block_timestamp" gorm:"column:block_timestamp"`
	CreatedAt      time.Time        `json:"created_at" gorm:"column:created_at"`
	UpdatedAt      time.Time        `json:"updated_at" gorm:"column:updated_at"`
}

const (
	TableNameTradeDealInfo = "t_trade_deal_info"
	DealTypeSale           = 0
	DealTypeAuction        = 1
	DealTypeOffer          = 2
)

func (t *TableTradeDealInfo) TableName() string {
	return TableNameTradeDealInfo
}
