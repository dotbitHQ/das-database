package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/molecule"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/nervosnetwork/ckb-sdk-go/crypto/blake2b"
	"gorm.io/gorm"
)

func (b *BlockParser) ActionReverseRecordRoot(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameReverseRecordRootCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err)
		return
	} else if !isCV {
		return
	}
	log.Info("ActionReverseRecordRoot:", req.BlockNumber, req.TxHash)

	txReverseSmtRecord := make([]*witness.ReverseSmtRecord, 0)
	if err := witness.ParseFromTx(req.Tx, common.ActionDataTypeReverseSmt, &txReverseSmtRecord); err != nil {
		resp.Err = err
		return
	}

	smtRecords := make([]*dao.ReverseSmtInfo, 0)
	for idx, v := range txReverseSmtRecord {
		nonce := molecule.GoU32ToMoleculeU32(v.PrevNonce + 1)
		valBs := make([]byte, 0)
		valBs = append(valBs, nonce.RawData()...)
		valBs = append(valBs, v.NextAccount...)

		smtValBlake256, err := blake2b.Blake256(valBs)
		if err != nil {
			resp.Err = err
			return
		}
		algorithmId := common.DasAlgorithmId(v.SignType)
		address := common.FormatAddressPayload(v.Address, algorithmId)

		outpoint := common.OutPoint2String(req.TxHash, uint(idx))
		smtRecord := &dao.ReverseSmtInfo{
			RootHash:     common.Bytes2Hex(v.NextRoot),
			BlockNumber:  req.BlockNumber,
			AlgorithmID:  uint8(algorithmId),
			Outpoint:     outpoint,
			Address:      address,
			LeafDataHash: common.Bytes2Hex(smtValBlake256),
		}
		smtRecords = append(smtRecords, smtRecord)
	}

	if err := b.dbDao.Transaction(func(tx *gorm.DB) error {
		for _, v := range smtRecords {
			err := tx.Where("algorithm_id=? and address=?", v.AlgorithmID, v.Address).Delete(&dao.ReverseSmtInfo{}).Error
			if err != nil && err != gorm.ErrRecordNotFound {
				return err
			}
			if err := tx.Create(v).Error; err != nil {
				return err
			}
		}
		for idx, v := range txReverseSmtRecord {
			outpoint := common.OutPoint2String(req.TxHash, uint(idx))
			accountId := common.Bytes2Hex(common.GetAccountIdByAccount(v.NextAccount))
			algorithmId := common.DasAlgorithmId(v.SignType)
			address := common.FormatAddressPayload(v.Address, algorithmId)
			reverseInfo := &dao.TableReverseInfo{
				BlockNumber:    req.BlockNumber,
				BlockTimestamp: req.BlockTimestamp,
				Outpoint:       outpoint,
				AlgorithmId:    algorithmId,
				ChainType:      algorithmId.ToChainType(),
				Address:        address,
				Account:        v.NextAccount,
				AccountId:      accountId,
				ReverseType:    dao.ReverseTypeSmt,
			}
			switch v.Action {
			case witness.ReverseSmtRecordActionUpdate:
				if v.PrevAccount != "" {
					if err := tx.Where("address=? and reverse_type=?", address, dao.ReverseTypeSmt).Delete(&dao.TableReverseInfo{}).Error; err != nil {
						return err
					}
				}
				if err := tx.Create(reverseInfo).Error; err != nil {
					return err
				}
			case witness.ReverseSmtRecordActionRemove:
				if err := tx.Where("address=? and reverse_type=?", address, dao.ReverseTypeSmt).Delete(&dao.TableReverseInfo{}).Error; err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		resp.Err = err
		return
	}
	return
}
