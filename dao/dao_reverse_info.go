package dao

import (
	"github.com/DeAccountSystems/das-lib/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TableReverseInfo struct {
	Id             uint64                `json:"id" gorm:"column:id"`
	BlockNumber    uint64                `json:"block_number" gorm:"column:block_number"`
	BlockTimestamp uint64                `json:"block_timestamp" gorm:"column:block_timestamp"`
	Outpoint       string                `json:"outpoint" gorm:"column:outpoint"`
	AlgorithmId    common.DasAlgorithmId `json:"algorithm_id" gorm:"column:algorithm_id"`
	ChainType      common.ChainType      `json:"chain_type" gorm:"column:chain_type"`
	Address        string                `json:"address" gorm:"column:address"`
	Account        string                `json:"account" gorm:"column:account"`
	Capacity       uint64                `json:"capacity" gorm:"column:capacity"`
}

const (
	TableNameReverseInfo = "t_reverse_info"
)

func (t *TableReverseInfo) TableName() string {
	return TableNameReverseInfo
}

func (d *DbDao) DeclareReverseRecord(reverseInfo TableReverseInfo, txInfo TableTransactionInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{"block_number", "block_timestamp", "capacity"}),
		}).Create(&reverseInfo).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{"block_number", "block_timestamp", "capacity"}),
		}).Create(&txInfo).Error; err != nil {
			return err
		}

		return nil
	})
}

func (d *DbDao) RedeclareReverseRecord(lastOutpoint string, reverseInfo TableReverseInfo, txInfo TableTransactionInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Where(" outpoint=? ", lastOutpoint).Delete(TableReverseInfo{}).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{"block_number", "block_timestamp", "capacity"}),
		}).Create(&reverseInfo).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{"block_number", "block_timestamp", "capacity"}),
		}).Create(&txInfo).Error; err != nil {
			return err
		}

		return nil
	})
}

func (d *DbDao) RetractReverseRecord(listOutpoint []string, txInfo TableTransactionInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Where(" outpoint IN(?) ", listOutpoint).Delete(TableReverseInfo{}).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{"block_number", "block_timestamp", "capacity"}),
		}).Create(&txInfo).Error; err != nil {
			return err
		}

		return nil
	})
}
