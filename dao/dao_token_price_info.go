package dao

import (
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"time"
)

type TableTokenPriceInfo struct {
	Id            uint64          `json:"id" gorm:"column:id;primaryKey;type:bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT ''"`
	TokenId       string          `json:"token_id" gorm:"column:token_id;uniqueIndex:uk_token_id;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	GeckoId       string          `json:"gecko_id" gorm:"column:gecko_id;uniqueIndex:uk_gecko_id;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT 'the id from coingecko'"`
	ChainType     int             `json:"chain_type" gorm:"column:chain_type;index:k_ct_c;type:smallint(6) NOT NULL DEFAULT '0' COMMENT ''"`
	CoinType      common.CoinType `json:"coin_type" gorm:"column:coin_type; type:varchar(255) NOT NULL DEFAULT '' COMMENT '';"`
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
	Icon          string          `json:"icon" gorm:"column:icon; type:varchar(255) NOT NULL DEFAULT '' COMMENT '';"`
	DisplayName   string          `json:"display_name" gorm:"column:display_name; type:varchar(255) NOT NULL DEFAULT '' COMMENT '';"`
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
	if len(tokenList) == 0 {
		return nil
	}
	return d.db.Transaction(func(tx *gorm.DB) error {
		for i, _ := range tokenList {
			if err := d.db.Model(TableTokenPriceInfo{}).
				Where("token_id=?", tokenList[i].TokenId).
				Updates(map[string]interface{}{
					"price":           tokenList[i].Price,
					"last_updated_at": tokenList[i].LastUpdatedAt,
				}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (d *DbDao) UpdateCNYToUSDRate(tokenIds []string, price decimal.Decimal) error {
	return d.db.Select("price", "last_updated_at").
		Where("token_id IN ?", tokenIds).
		Updates(TableTokenPriceInfo{
			Price:         price,
			LastUpdatedAt: time.Now().Unix(),
		}).Error
}
