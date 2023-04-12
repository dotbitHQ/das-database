package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"gorm.io/gorm"
)

func (b *BlockParser) ActionConfigSubAccount(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	isCV, txIndex, err := CurrentVersionTx(req.Tx, common.DASContractNameSubAccountCellType)
	if err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warnf("not current version %s tx", common.DASContractNameSubAccountCellType)
		return
	}
	log.Info("ActionConfigSubAccount:", req.BlockNumber, req.TxHash)

	parentAccountId := common.Bytes2Hex(req.Tx.Outputs[txIndex].Type.Args)

	if err := b.dbDao.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("account_id=?", parentAccountId).Delete(&dao.RuleConfig{}).Error; err != nil {
			return err
		}

		accountInfo := &dao.TableAccountInfo{}
		if err := tx.Where("account_id=?", parentAccountId).First(accountInfo).Error; err != nil {
			return err
		}

		return tx.Create(dao.RuleConfig{
			Account:        accountInfo.Account,
			AccountId:      accountInfo.AccountId,
			TxHash:         req.TxHash,
			BlockNumber:    req.BlockNumber,
			BlockTimestamp: req.BlockTimestamp,
		}).Error
	}); err != nil {
		resp.Err = fmt.Errorf("ActionConfigSubAccount err: %s", err.Error())
		return
	}
	return
}
