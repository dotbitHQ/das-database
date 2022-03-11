package dao

import (
	"github.com/DeAccountSystems/das-lib/common"
	"time"
)

type TableRebateInfo struct {
	Id               uint64           `json:"id" gorm:"column:id;primaryKey;type:bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT ''"`
	BlockNumber      uint64           `json:"block_number" gorm:"column:block_number;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	Outpoint         string           `json:"outpoint" gorm:"column:outpoint;uniqueIndex:uk_o_rt;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	InviteeId        string           `json:"invitee_id" gorm:"column:invitee_id;index:k_invitee_id;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT 'account id of invitee'"`
	InviteeAccount   string           `json:"invitee_account" gorm:"column:invitee_account;index:k_invitee_account;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	InviteeChainType common.ChainType `json:"invitee_chain_type" gorm:"column:invitee_chain_type;index:k_ict_ia;type:smallint(6) NOT NULL DEFAULT '0' COMMENT ''"`
	InviteeAddress   string           `json:"invitee_address" gorm:"column:invitee_address;index:k_ict_ia;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	RewardType       int              `json:"reward_type" gorm:"column:reward_type;uniqueIndex:uk_o_rt;type:smallint(6) NOT NULL DEFAULT '0' COMMENT '1: invite 2: channel'"`
	Reward           uint64           `json:"reward" gorm:"column:reward;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT 'reward amount'"`
	Action           string           `json:"action" gorm:"column:action;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	ServiceType      int              `json:"service_type" gorm:"column:service_type;type:smallint(6) NOT NULL DEFAULT '0' COMMENT '1: register 2: trade'"`
	InviterArgs      string           `json:"inviter_args" gorm:"column:inviter_args;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	InviterId        string           `json:"inviter_id" gorm:"column:inviter_id;index:k_inviter_id;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT 'account id of inviter'"`
	InviterAccount   string           `json:"inviter_account" gorm:"column:inviter_account;index:k_inviter_account;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT 'inviter account'"`
	InviterChainType common.ChainType `json:"inviter_chain_type" gorm:"column:inviter_chain_type;type:smallint(6) NOT NULL DEFAULT '0' COMMENT ''"`
	InviterAddress   string           `json:"inviter_address" gorm:"column:inviter_address;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT 'address of inviter'"`
	BlockTimestamp   uint64           `json:"block_timestamp" gorm:"column:block_timestamp;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	CreatedAt        time.Time        `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''"`
	UpdatedAt        time.Time        `json:"updated_at" gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
}

const (
	TableNameRebateInfo = "t_rebate_info"
	RewardTypeInviter   = 0
	RewardTypeChannel   = 1
)

func (t *TableRebateInfo) TableName() string {
	return TableNameRebateInfo
}
