package dao

import (
	"das_sub_account/tables"
	"errors"
	"github.com/dotbitHQ/das-lib/common"
	"gorm.io/gorm"
	"time"
)

type ApprovalStatus int

const (
	ApprovalStatusEnable  ApprovalStatus = 1
	ApprovalStatusFulFill ApprovalStatus = 2
	ApprovalStatusRevoke  ApprovalStatus = 3
)

type ApprovalInfo struct {
	ID               uint64                `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	BlockNumber      uint64                `gorm:"column:block_number;default:0;NOT NULL"`
	RefOutpoint      string                `gorm:"column:ref_outpoint;index:idx_ref_outpoint;NOT NULL"` // Hash-Index
	Outpoint         string                `gorm:"column:outpoint;index:idx_outpoint;NOT NULL"`         // Hash-Index
	Account          string                `gorm:"column:account;NOT NULL"`
	AccountID        string                `gorm:"column:account_id;index:idx_account_id;NOT NULL"`
	ParentAccountID  string                `gorm:"column:parent_account_id;index:idx_parent_account_id;NOT NULL"`
	Platform         string                `gorm:"column:platform;NOT NULL"` // platform address
	OwnerAlgorithmID common.DasAlgorithmId `gorm:"column:owner_algorithm_id;default:0;NOT NULL"`
	Owner            string                `gorm:"column:owner;index:idx_owner;NOT NULL"` // owner address
	ToAlgorithmID    common.DasAlgorithmId `gorm:"column:to_algorithm_id;default:0;NOT NULL"`
	To               string                `gorm:"column:to;index:idx_to;NOT NULL"`           // to address
	ProtectedUntil   uint64                `gorm:"column:protected_until;default:0;NOT NULL"` // 授权的不可撤销时间
	SealedUntil      uint64                `gorm:"column:sealed_until;default:0;NOT NULL"`    // 授权开放时间
	MaxDelayCount    uint8                 `gorm:"column:max_delay_count;default:0;NOT NULL"` // 可推迟次数
	PostponedCount   int                   `gorm:"column:postponed_count;default:0;NOT NULL"` // 推迟过的次数
	Status           ApprovalStatus        `gorm:"column:status;default:0;NOT NULL"`          // 0-default 1-开启授权 2:完成授权 3-撤销授权
	CreatedAt        time.Time             `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL" json:"created_at"`
	UpdatedAt        time.Time             `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL" json:"updated_at"`
}

func (m *ApprovalInfo) TableName() string {
	return "t_approval_info"
}

func (d *DbDao) CreateAccountApproval(info ApprovalInfo) (err error) {
	err = d.db.Create(&info).Error
	return
}

func (d *DbDao) UpdateAccountApproval(id uint64, info map[string]interface{}) (err error) {
	err = d.db.Model(&ApprovalInfo{}).Where("id=?", id).Updates(info).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

func (d *DbDao) GetAccountPendingApproval(accountId string) (approval ApprovalInfo, err error) {
	err = d.db.Where("account_id=? and status=?", accountId, tables.ApprovalStatusEnable).Order("id desc").First(&approval).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

func (d *DbDao) GetPendingApprovalByAccIdAndPlatform(accountId, platform string) (approval ApprovalInfo, err error) {
	err = d.db.Where("account_id=? and platform=? and status=?", accountId, platform, ApprovalStatusEnable).Order("id desc").First(&approval).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}
