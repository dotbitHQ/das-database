package dao

import (
	"gorm.io/gorm/clause"
	"time"
)

type TableBlockInfo struct {
	Id          uint64     `json:"id" gorm:"column:id; primaryKey; type:bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '';"`
	ParserType  ParserType `json:"parser_type" gorm:"column:parser_type; uniqueIndex:uk_pt_bn; type:smallint(6) NOT NULL DEFAULT '0' COMMENT '';"`
	BlockNumber uint64     `json:"block_number" gorm:"column:block_number; uniqueIndex:uk_pt_bn; type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '';"`
	BlockHash   string     `json:"block_hash" gorm:"column:block_hash; type:varchar(255) NOT NULL DEFAULT '' COMMENT '';"`
	ParentHash  string     `json:"parent_hash" gorm:"column:parent_hash; type:varchar(255) NOT NULL DEFAULT '' COMMENT '';"`
	CreatedAt   time.Time  `json:"created_at" gorm:"column:created_at; type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '';"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"column:updated_at; type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '';"`
}

const (
	TableNameBlockInfo = "t_block_info"
)

func (t *TableBlockInfo) TableName() string {
	return TableNameBlockInfo
}

type ParserType int

const (
	ParserTypeDAS        = 99
	ParserTypeSubAccount = 98 // das-sub-account
	ParserTypeSnapshot   = 97

	ParserTypeCKB     = 0
	ParserTypeETH     = 1
	ParserTypeTRON    = 3
	ParserTypeBSC     = 5
	ParserTypePOLYGON = 6
)

func (d *DbDao) CreateBlockInfo(parserType ParserType, blockNumber uint64, blockHash, parentHash string) error {
	return d.db.Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{"block_hash", "parent_hash"}),
	}).Create(&TableBlockInfo{
		ParserType:  parserType,
		BlockNumber: blockNumber,
		BlockHash:   blockHash,
		ParentHash:  parentHash,
	}).Error
}

func (d *DbDao) DeleteBlockInfo(parserType ParserType, blockNumber uint64) error {
	return d.db.Where("parser_type=? AND block_number<?", parserType, blockNumber).Delete(&TableBlockInfo{}).Error
}

func (d *DbDao) FindBlockInfo(parserType ParserType) (blockInfo TableBlockInfo, err error) {
	err = d.db.Where("parser_type=?", parserType).
		Order("block_number DESC").Limit(1).Find(&blockInfo).Error
	return
}

func (d *DbDao) FindBlockInfoByBlockNumber(parserType ParserType, blockNumber uint64) (blockInfo TableBlockInfo, err error) {
	err = d.db.Where("parser_type=? AND block_number=?", parserType, blockNumber).Limit(1).Find(&blockInfo).Error
	return
}

func (d *DbDao) CreateBlockInfoList(list []TableBlockInfo) error {
	return d.db.Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{"block_hash", "parent_hash"}),
	}).Create(&list).Error
}
