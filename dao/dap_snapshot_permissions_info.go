package dao

import "time"

type TableSnapshotPermissionsInfo struct {
	Id                 uint64    `json:"id" gorm:"column:id; primaryKey; type:bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '';"`
	BlockNumber        uint64    `json:"block_number" gorm:"column:block_number; type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '';"`
	AccountId          string    `json:"account_id" gorm:"column:account_id; uniqueIndex:uk_account_id; type:varchar(255) NOT NULL DEFAULT '' COMMENT '';"`
	Outpoint           string    `json:"outpoint" gorm:"column:outpoint; uniqueIndex:uk_account_id; type:varchar(255) NOT NULL DEFAULT '' COMMENT '';"`
	Account            string    `json:"account" gorm:"column:account; type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '';"`
	BlockTimestamp     uint64    `json:"block_timestamp" gorm:"column:block_timestamp; type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '';"`
	Owner              string    `json:"owner" gorm:"column:owner; index:k_owner; type:varchar(255) NOT NULL DEFAULT '' COMMENT '';"`
	Manager            string    `json:"manager" gorm:"column:manager; index:k_manager; type:varchar(255) NOT NULL DEFAULT '' COMMENT '';"`
	OwnerAlgorithmId   int       `json:"owner_algorithm_id" gorm:"column:owner_algorithm_id; type:SMALLINT(6) NOT NULL DEFAULT '0' COMMENT '';"`
	ManagerAlgorithmId int       `json:"manager_algorithm_id" gorm:"column:manager_algorithm_id; type:SMALLINT(6) NOT NULL DEFAULT '0' COMMENT '';"`
	CreatedAt          time.Time `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
}

const (
	TableNameSnapshotPermissionsInfo = "t_snapshot_permissions_info"
)

func (t *TableSnapshotPermissionsInfo) TableName() string {
	return TableNameSnapshotPermissionsInfo
}
