package dao

import (
	"das-multi-device/tables"
	"fmt"
	"gorm.io/gorm/clause"
	"time"
)

const (
	TableNameCidPk = "t_cid_pk"
)

type IsEnableAuthorize = uint8

const (
	EnableAuthorizeOff IsEnableAuthorize = 0
	EnableAuthorizeOn  IsEnableAuthorize = 1
)

type TableCidPk struct {
	Id              uint64            `json:"id" gorm:"column:id;primaryKey;type:bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT ''"`
	Cid             string            `json:"cid" gorm:"column:cid;uniqueIndex:uk_cid; type:varchar(255) NOT NULL DEFAULT '0'"`
	Pk              string            `json:"pk" gorm:"column:pk; type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '';"`
	OriginPk        string            `json:"origin_pk" gorm:"column:origin_pk; type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT ''"`
	EnableAuthorize IsEnableAuthorize `json:"enable_authorize" gorm:"column:enable_authorize; type:tinyint NOT NULL DEFAULT '0'"`
	Outpoint        string            `json:"outpoint" gorm:"column:outpoint;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL"`
	CreatedAt       time.Time         `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''""`
	UpdatedAt       time.Time         `json:"updated_at" gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
}

func (t *TableCidPk) TableName() string {
	return TableNameCidPk
}

//insert cid pk
func (d *DbDao) InsertCidPk(data []TableCidPk) (err error) {
	if len(data) == 0 {
		return fmt.Errorf("data is empty")
	}
	if err := d.db.Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{
			"enable_authorize", "outpoint",
		}),
	}).Create(&data).Error; err != nil {
		return err
	}
	return
}

func (d *DbDao) GetCidPk(cid1 string) (cidpk tables.TableCidPk, err error) {
	err = d.db.Where("`cid`= ? ", cid1).Find(&cidpk).Error
	return
}
