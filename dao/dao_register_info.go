package dao

import (
	"gorm.io/gorm/clause"
	"time"
)

type TableRegisterInfo struct {
	Id              uint64    `json:"id" gorm:"column:id;primaryKey;type:bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT ''"`
	RegisterDate    string    `json:"register_date" gorm:"column:register_date;uniqueIndex:uk_register_date;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '2022-01-02'"`
	TotalAccount    int       `json:"total_account" gorm:"column:total_account;type:smallint(6) NOT NULL DEFAULT '0' COMMENT ''"`
	TotalSubAccount int       `json:"total_sub_account" gorm:"column:total_sub_account;type:smallint(6) NOT NULL DEFAULT '0' COMMENT ''"`
	TotalOwner      int       `json:"total_owner" gorm:"column:total_owner;type:smallint(6) NOT NULL DEFAULT '0' COMMENT 'independent owner'"`
	One             int       `json:"one" gorm:"column:one;type:smallint(6) NOT NULL DEFAULT '0' COMMENT ''"`
	Two             int       `json:"two" gorm:"column:two;type:smallint(6) NOT NULL DEFAULT '0' COMMENT ''"`
	Three           int       `json:"three" gorm:"column:three;type:smallint(6) NOT NULL DEFAULT '0' COMMENT ''"`
	Four            int       `json:"four" gorm:"column:four;type:smallint(6) NOT NULL DEFAULT '0' COMMENT ''"`
	FiveAndMore     int       `json:"five_and_more" gorm:"column:five_and_more;type:smallint(6) NOT NULL DEFAULT '0' COMMENT ''"`
	RegisterDetail  string    `json:"register_detail" gorm:"column:register_detail;type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '{\"4\":1,\"5\":10,\"sub\":4}'"`
	CreatedAt       time.Time `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
}

const (
	TableNameRegisterInfo = "t_register_info"
)

func (t *TableRegisterInfo) TableName() string {
	return TableNameRegisterInfo
}

func (d *DbDao) CreateRegisterInfo(registerInfo TableRegisterInfo) error {
	return d.db.Clauses(clause.Insert{
		Modifier: "IGNORE",
	}).Create(&registerInfo).Error
}
