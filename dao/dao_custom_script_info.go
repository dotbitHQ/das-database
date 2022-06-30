package dao

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type TableCustomScriptInfo struct {
	Id             uint64    `json:"id" gorm:"column:id;primaryKey;type:bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT ''"`
	BlockNumber    uint64    `json:"block_number" gorm:"column:block_number;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	Outpoint       string    `json:"outpoint" gorm:"column:outpoint;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT 'Hash-Index'"`
	BlockTimestamp uint64    `json:"block_timestamp" gorm:"column:block_timestamp;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	AccountId      string    `json:"account_id" gorm:"column:account_id;uniqueIndex:uk_account_id;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
}

const (
	TableNameCustomScriptInfo = "t_custom_script_info"
)

func (t *TableCustomScriptInfo) TableName() string {
	return TableNameCustomScriptInfo
}

func (d *DbDao) UpdateCustomScript(cs TableCustomScriptInfo, accountCellOutpoint string) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(TableAccountInfo{}).Where("account_id=?", cs.AccountId).
			Updates(map[string]interface{}{
				"outpoint": accountCellOutpoint,
			}).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.Insert{
			Modifier: "IGNORE",
		}).Create(&cs).Error; err != nil {
			return err
		}

		if err := tx.Model(TableCustomScriptInfo{}).Where("account_id=?", cs.AccountId).
			Updates(map[string]interface{}{
				"block_number":    cs.BlockNumber,
				"outpoint":        cs.Outpoint,
				"block_timestamp": cs.BlockTimestamp,
			}).Error; err != nil {
			return err
		}

		return nil
	})
}
