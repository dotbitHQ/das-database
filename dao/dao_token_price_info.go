package dao

import (
	"fmt"
	"github.com/shopspring/decimal"
	"gorm.io/gorm/clause"
	"time"
)

type TableTokenPriceInfo struct {
	Id            uint64          `json:"id" gorm:"column:id;primaryKey;type:bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT ''"`
	TokenId       string          `json:"token_id" gorm:"column:token_id;uniqueIndex:uk_token_id;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	GeckoId       string          `json:"gecko_id" gorm:"column:gecko_id;uniqueIndex:uk_gecko_id;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT 'the id from coingecko'"`
	ChainType     int             `json:"chain_type" gorm:"column:chain_type;index:k_ct_c;type:smallint(6) NOT NULL DEFAULT '0' COMMENT ''"`
	Contract      string          `json:"contact" gorm:"column:contract;index:k_ct_c;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	Name          string          `json:"name" gorm:"column:name;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT 'the name of token'"`
	Symbol        string          `json:"symbol" gorm:"column:symbol;index:k_symbol;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT 'the symbol of token'"`
	Decimals      int32           `json:"decimals" gorm:"column:decimals;type:smallint(6) NOT NULL DEFAULT '0' COMMENT ''"`
	Price         decimal.Decimal `json:"price" gorm:"column:price;type:decimal(50, 8) NOT NULL DEFAULT '0.00000000' COMMENT ''"`
	Logo          string          `json:"logo" gorm:"column:logo;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	Change24h     decimal.Decimal `json:"change_24_h" gorm:"column:change_24_h;type:decimal(50, 8) NOT NULL DEFAULT '0.00000000' COMMENT ''"`
	Vol24h        decimal.Decimal `json:"vol_24_h" gorm:"column:vol_24_h;type:decimal(50, 8) NOT NULL DEFAULT '0.00000000' COMMENT ''"`
	MarketCap     decimal.Decimal `json:"market_cap" gorm:"column:market_cap;type:decimal(50, 8) NOT NULL DEFAULT '0.00000000' COMMENT ''"`
	LastUpdatedAt int64           `json:"last_updated_at" gorm:"column:last_updated_at;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	Status        int             `json:"status" gorm:"column:status;type:smallint(6) NOT NULL DEFAULT '0' COMMENT '0: normal 1: banned'"`
	CreatedAt     time.Time       `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''"`
	UpdatedAt     time.Time       `json:"updated_at" gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
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
