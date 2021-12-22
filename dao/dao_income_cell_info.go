package dao

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type TableIncomeCellInfo struct {
	Id             uint64    `json:"id" gorm:"column:id;primary_key;AUTO_INCREMENT"`
	BlockNumber    uint64    `json:"block_number" gorm:"column:block_number"`
	Action         string    `json:"action" gorm:"column:action"`
	Outpoint       string    `json:"outpoint" gorm:"column:outpoint"`
	Capacity       uint64    `json:"capacity" gorm:"column:capacity"`
	BlockTimestamp uint64    `json:"block_timestamp" gorm:"column:block_timestamp"`
	Status         int       `json:"status" gorm:"column:status"`
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"column:updated_at"`
}

const (
	TableNameIncomeCellInfo = "t_income_cell_info"
	IncomeCellStatusUnMerge = 0
	IncomeCellStatusMerged  = 1
)

func (t *TableIncomeCellInfo) TableName() string {
	return TableNameIncomeCellInfo
}

func (d *DbDao) CreateIncomeCellInfo(incomeCellInfo TableIncomeCellInfo) error {
	return d.db.Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{"block_number", "capacity", "block_timestamp"}),
	}).Create(&incomeCellInfo).Error
}

func (d *DbDao) CreateIncomeCellInfoList(incomeCellInfos []TableIncomeCellInfo) error {
	return d.db.Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{"block_number", "capacity", "block_timestamp"}),
	}).Create(&incomeCellInfos).Error
}

func (d *DbDao) DeleteIncomeCellInfo() error {
	return d.db.Delete(&TableIncomeCellInfo{}).Error
}

func (d *DbDao) DeleteIncomeCellInfoByOutpoint(outpoint string) error {
	return d.db.Where("outpoint = ?", outpoint).Delete(&TableIncomeCellInfo{}).Error
}

func (d *DbDao) SaveIncomeCellInfo(incomeCellInfo TableIncomeCellInfo) error {
	return d.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "outpoint"}},
		DoUpdates: clause.AssignmentColumns([]string{"block_number", "capacity", "block_timestamp"}),
	}).Save(&incomeCellInfo).Error
}

func (d *DbDao) UpdateIncomeCellInfoMerged(outpoint []string) error {
	return d.db.Model(&TableIncomeCellInfo{}).Where("outpoint IN ?", outpoint).Update("status", IncomeCellStatusMerged).Error
}

func (d *DbDao) UpdatesIncomeCellInfo(incomeCellInfo TableIncomeCellInfo) error {
	return d.db.Select("block_number", "capacity", "block_timestamp").Where("outpoint = ?", incomeCellInfo.Outpoint).Updates(incomeCellInfo).Error
}

func (d *DbDao) FirstIncomeCellInfoByOutpoint(outpoint string) (incomeCellInfo TableIncomeCellInfo, err error) {
	err = d.db.Where("outpoint = ?", outpoint).Limit(1).Find(&incomeCellInfo).Error
	return
}

func (d *DbDao) FindIncomeCellInfoListByOutpoint(outpoint string) (incomeCellInfo []TableIncomeCellInfo, err error) {
	err = d.db.Where("outpoint = ?", outpoint).Find(&incomeCellInfo).Error
	return
}

func (d *DbDao) ConsolidateIncome(outpoints []string, incomeCellInfos []TableIncomeCellInfo, transactionInfos []TableTransactionInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("outpoint IN ?", outpoints).Delete(&TableIncomeCellInfo{}).Error; err != nil {
			return err
		}

		if len(incomeCellInfos) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns([]string{
					"action", "capacity", "status",
				}),
			}).Create(&incomeCellInfos).Error; err != nil {
				return err
			}
		}

		if len(transactionInfos) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns([]string{
					"account_id", "account", "service_type",
					"chain_type", "address", "capacity", "status",
				}),
			}).Create(&transactionInfos).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (d *DbDao) RenewAccount(outpoints []string, incomeCellInfos []TableIncomeCellInfo, accountInfo TableAccountInfo, transactionInfo TableTransactionInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("outpoint IN ?", outpoints).Delete(&TableIncomeCellInfo{}).Error; err != nil {
			return err
		}

		if len(incomeCellInfos) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns([]string{
					"action", "capacity", "status",
				}),
			}).Create(&incomeCellInfos).Error; err != nil {
				return err
			}
		}

		if err := tx.Select("block_number", "outpoint", "expired_at").
			Where("account_id = ?", accountInfo.AccountId).
			Updates(accountInfo).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{
				"account_id", "account", "service_type",
				"chain_type", "address", "capacity", "status",
			}),
		}).Create(&transactionInfo).Error; err != nil {
			return err
		}

		return nil
	})
}
