package dao

import (
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
	"time"
)

type TableSnapshotPermissionsInfo struct {
	Id                 uint64                `json:"id" gorm:"column:id; primaryKey; type:bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '';"`
	BlockNumber        uint64                `json:"block_number" gorm:"column:block_number; index:k_block_number; type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '';"`
	AccountId          string                `json:"account_id" gorm:"column:account_id; uniqueIndex:uk_account_id; type:varchar(255) NOT NULL DEFAULT '' COMMENT '';"`
	Hash               string                `json:"hash" gorm:"column:hash; uniqueIndex:uk_account_id; index:k_hash; type:varchar(255) NOT NULL DEFAULT '' COMMENT '';"`
	Account            string                `json:"account" gorm:"column:account; type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '';"`
	BlockTimestamp     uint64                `json:"block_timestamp" gorm:"column:block_timestamp; type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '';"`
	Owner              string                `json:"owner" gorm:"column:owner; index:k_owner; type:varchar(255) NOT NULL DEFAULT '' COMMENT '';"`
	Manager            string                `json:"manager" gorm:"column:manager; index:k_manager; type:varchar(255) NOT NULL DEFAULT '' COMMENT '';"`
	OwnerAlgorithmId   common.DasAlgorithmId `json:"owner_algorithm_id" gorm:"column:owner_algorithm_id; type:SMALLINT(6) NOT NULL DEFAULT '0' COMMENT '';"`
	ManagerAlgorithmId common.DasAlgorithmId `json:"manager_algorithm_id" gorm:"column:manager_algorithm_id; type:SMALLINT(6) NOT NULL DEFAULT '0' COMMENT '';"`
	OwnerBlockNumber   uint64                `json:"owner_block_number" gorm:"column:owner_block_number; type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '';"`
	ManagerBlockNumber uint64                `json:"manager_block_number" gorm:"column:manager_block_number; type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '';"`
	Status             AccountStatus         `json:"status" gorm:"column:status; type:SMALLINT(6) NOT NULL DEFAULT '0' COMMENT '';"`
	ExpiredAt          uint64                `json:"expired_at" gorm:"column:expired_at; type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '';"`
	CreatedAt          time.Time             `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''"`
	UpdatedAt          time.Time             `json:"updated_at" gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
}

const (
	TableNameSnapshotPermissionsInfo = "t_snapshot_permissions_info"
)

func (t *TableSnapshotPermissionsInfo) TableName() string {
	return TableNameSnapshotPermissionsInfo
}

type RoleType string

const (
	RoleTypeOwner   RoleType = "owner"
	RoleTypeManager RoleType = "manager"
)

func (d *DbDao) CreateSnapshotPermissions(list []TableSnapshotPermissionsInfo) error {
	if len(list) == 0 {
		return nil
	}

	var accountIds []string
	var mapNewPermissions = make(map[string]*TableSnapshotPermissionsInfo)
	for i, v := range list {
		accountIds = append(accountIds, v.AccountId)
		mapNewPermissions[v.AccountId] = &list[i]
		if v.Status == AccountStatusRecycle {
			list[i].OwnerBlockNumber = v.BlockNumber
			list[i].ManagerBlockNumber = v.BlockNumber
		}
	}

	oldPermissions, err := d.GetPreSnapshotPermissionsByAccountIds(accountIds, list[0].BlockNumber)
	if err != nil {
		return fmt.Errorf("GetPreSnapshotPermissionsByAccountIds err:%s", err.Error())
	}
	var updatePermissions []TableSnapshotPermissionsInfo
	for i, v := range oldPermissions {
		needUpdate := false
		newPermissions := mapNewPermissions[v.AccountId]

		if newPermissions.Status != AccountStatusRecycle && v.Status != AccountStatusRecycle && !strings.EqualFold(newPermissions.Owner, v.Owner) {
			oldPermissions[i].OwnerBlockNumber = newPermissions.BlockNumber
			needUpdate = true
		}
		if newPermissions.Status != AccountStatusRecycle && !strings.EqualFold(newPermissions.Manager, v.Manager) {
			oldPermissions[i].ManagerBlockNumber = newPermissions.BlockNumber
			needUpdate = true
		}
		if newPermissions.Status == AccountStatusRecycle && v.Status != AccountStatusRecycle {
			oldPermissions[i].OwnerBlockNumber = newPermissions.BlockNumber
			oldPermissions[i].ManagerBlockNumber = newPermissions.BlockNumber
			needUpdate = true
		}
		if needUpdate {
			updatePermissions = append(updatePermissions, oldPermissions[i])
		}
	}

	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Insert{
			Modifier: "IGNORE",
		}).Create(&list).Error; err != nil {
			return err
		}
		for _, v := range updatePermissions {
			if err := tx.Model(TableSnapshotPermissionsInfo{}).
				Where("id=?", v.Id).
				Updates(map[string]interface{}{
					"owner_block_number":   v.OwnerBlockNumber,
					"manager_block_number": v.ManagerBlockNumber,
				}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (d *DbDao) GetPreSnapshotPermissionsByAccountIds(accountIds []string, blockNumber uint64) (list []TableSnapshotPermissionsInfo, err error) {
	//sql := fmt.Sprintf(" SELECT id,account_id,`owner`,manager FROM %s WHERE id IN( SELECT MAX(id) AS id FROM %s WHERE account_id IN(?) AND block_number<? GROUP BY `account_id` )",
	//	TableNameSnapshotPermissionsInfo, TableNameSnapshotPermissionsInfo)

	sql := fmt.Sprintf(" SELECT b.id,b.account_id,b.`owner`,b.manager FROM ( SELECT MAX(id) AS id FROM %s WHERE account_id IN(?) AND block_number<? GROUP BY `account_id` )a JOIN %s b ON b.id=a.id ",
		TableNameSnapshotPermissionsInfo, TableNameSnapshotPermissionsInfo)

	err = d.db.Raw(sql, accountIds, blockNumber).Find(&list).Error
	return
}

func (d *DbDao) GetSnapshotPermissionsInfo(accountId string, blockNumber uint64) (info TableSnapshotPermissionsInfo, err error) {
	err = d.db.Where("account_id=? AND block_number<=?",
		accountId, blockNumber).Order("block_number DESC,id DESC").Limit(1).Find(&info).Error
	return
}

func (d *DbDao) GetRecycleInfo(accountId string, startBlockNumber, endBlockNumber uint64) (info TableSnapshotPermissionsInfo, err error) {
	err = d.db.Where("account_id=? AND block_number>=? AND block_number<=? AND `status`=?",
		accountId, startBlockNumber, endBlockNumber, AccountStatusRecycle).Limit(1).Find(&info).Error
	return
}

func (d *DbDao) GetSnapshotAddressAccounts(addressHex string, roleType RoleType, blockNumber uint64, limit, offset int) (list []TableSnapshotPermissionsInfo, err error) {
	switch roleType {
	case RoleTypeOwner:
		err = d.db.Select("account_id,account").
			Where("owner=? AND block_number<=? AND (owner_block_number=0 OR owner_block_number>?)",
				addressHex, blockNumber, blockNumber).
			Group("account_id,account").
			Order("account").
			Limit(limit).Offset(offset).Find(&list).Error
	case RoleTypeManager:
		err = d.db.Select("account_id,account").
			Where("manager=? AND block_number<=? AND (manager_block_number=0 OR manager_block_number>?)",
				addressHex, blockNumber, blockNumber).
			Group("account_id,account").
			Order("account").
			Limit(limit).Offset(offset).Find(&list).Error
	}

	return
}

func (d *DbDao) GetSnapshotAddressAccountsTotal(addressHex string, roleType RoleType, blockNumber uint64) (count int64, err error) {
	switch roleType {
	case RoleTypeOwner:
		err = d.db.Model(&TableSnapshotPermissionsInfo{}).
			Where("owner=? AND block_number<=? AND (owner_block_number=0 OR owner_block_number>?)",
				addressHex, blockNumber, blockNumber).
			Group("account_id,account").Count(&count).Error
	case RoleTypeManager:
		err = d.db.Model(&TableSnapshotPermissionsInfo{}).
			Where("manager=? AND block_number<=? AND (manager_block_number=0 OR manager_block_number>?)",
				addressHex, blockNumber, blockNumber).
			Group("account_id,account").Count(&count).Error
	}

	return
}
