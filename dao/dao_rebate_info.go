package dao

import (
	"github.com/DeAccountSystems/das-lib/common"
	"time"
)

type TableRebateInfo struct {
	Id               uint64           `json:"id" gorm:"column:id;primary_key;AUTO_INCREMENT"`
	BlockNumber      uint64           `json:"block_number" gorm:"column:block_number"`
	Outpoint         string           `json:"outpoint" gorm:"column:outpoint"`
	InviteeAccount   string           `json:"invitee_account" gorm:"column:invitee_account"`
	InviteeChainType common.ChainType `json:"invitee_chain_type" gorm:"column:invitee_chain_type"`
	InviteeAddress   string           `json:"invitee_address" gorm:"column:invitee_address"`
	RewardType       int              `json:"reward_type" gorm:"column:reward_type"`               // 1: inviter 2: channel
	Reward           uint64           `json:"reward" gorm:"column:reward"`
	Action           string           `json:"action" gorm:"column:action"`
	ServiceType      int              `json:"service_type" gorm:"column:service_type"`             // 1: register 2: trade
	InviterArgs      string           `json:"inviter_args" gorm:"column:inviter_args"`
	InviterId        string           `json:"inviter_id" gorm:"column:inviter_id"`
	InviterAccount   string           `json:"inviter_account" gorm:"column:inviter_account"`
	InviterChainType common.ChainType `json:"inviter_chain_type" gorm:"column:inviter_chain_type"`
	InviterAddress   string           `json:"inviter_address" gorm:"column:inviter_address"`
	BlockTimestamp   uint64           `json:"block_timestamp" gorm:"column:block_timestamp"`
	CreatedAt        time.Time        `json:"created_at" gorm:"column:created_at"`
	UpdatedAt        time.Time        `json:"updated_at" gorm:"column:updated_at"`
}

const (
	TableNameRebateInfo = "t_rebate_info"
	RewardTypeInviter   = 0
	RewardTypeChannel   = 1
)

func (t *TableRebateInfo) TableName() string {
	return TableNameRebateInfo
}
