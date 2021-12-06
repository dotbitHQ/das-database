package dao

import (
	"fmt"
	"github.com/shopspring/decimal"
	"gorm.io/gorm/clause"
	"time"
)

type TableTokenPriceInfo struct {
	Id            uint64          `json:"id" gorm:"column:id;primary_key;AUTO_INCREMENT"`
	TokenId       string          `json:"token_id" gorm:"column:token_id"`
	GeckoId       string          `json:"gecko_id" gorm:"column:gecko_id"`
	ChainType     int             `json:"chain_type" gorm:"column:chain_type"`
	Contract      string          `json:"contact" gorm:"column:contract"`
	Name          string          `json:"name" gorm:"column:name"`
	Symbol        string          `json:"symbol" gorm:"column:symbol"`
	Decimals      int32           `json:"decimals" gorm:"column:decimals"`
	Price         decimal.Decimal `json:"price" gorm:"column:price"`
	Logo          string          `json:"logo" gorm:"column:logo"`
	Change24h     decimal.Decimal `json:"change_24_h" gorm:"column:change_24_h"`
	Vol24h        decimal.Decimal `json:"vol_24_h" gorm:"column:vol_24_h"`
	MarketCap     decimal.Decimal `json:"market_cap" gorm:"column:market_cap"`
	LastUpdatedAt int64           `json:"last_updated_at" gorm:"column:last_updated_at"`
	Status        int             `json:"status" gorm:"column:status"`
	CreatedAt     time.Time       `json:"created_at" gorm:"column:created_at"`
	UpdatedAt     time.Time       `json:"updated_at" gorm:"column:updated_at"`
}

const (
	TableNameTokenPriceInfo = "t_token_price_info"
)

func (t *TableTokenPriceInfo) TableName() string {
	return TableNameTokenPriceInfo
}

func (t *TableTokenPriceInfo) GetPriceUsd(price uint64) decimal.Decimal {
	decPrice, _ := decimal.NewFromString(fmt.Sprintf("%d", price))
	return t.Price.Mul(decPrice).DivRound(decimal.New(1, t.Decimals), 6)
}

func (d *DbDao) SearchTokenPriceInfoList() (tokenPriceInfos []TableTokenPriceInfo, err error) {
	err = d.db.Order("id DESC").Find(&tokenPriceInfos).Error
	return
}

func (d *DbDao) UpdateTokenPriceInfoList(tokenList []TableTokenPriceInfo) error {
	return d.db.Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{"price", "change_24_h", "vol_24_h", "market_cap", "last_updated_at"}),
	}).Create(&tokenList).Error
}

func (d *DbDao) UpdateCNYToUSDRate(tokenIds []string, price decimal.Decimal) error {
	return d.db.Select("price", "last_updated_at").Where("token_id IN ?", tokenIds).Updates(TableTokenPriceInfo{
		Price:         price,
		LastUpdatedAt: time.Now().Unix(),
	}).Error
}
