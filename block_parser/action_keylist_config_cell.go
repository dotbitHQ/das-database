package block_parser

import (
	"das_database/dao"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/witness"
)

func (b *BlockParser) ActionCreateDeviceKeyList(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasKeyListCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version transfer account tx")
		return
	}
	log.Info("ActionCreateDeviceKeyList:", req.BlockNumber, req.TxHash)

	builder, err := witness.WebAuthnKeyListDataBuilderFromTx(req.Tx, common.DataTypeNew)
	//add cidpk
	var cidPk []dao.TableCidPk
	keyList := witness.ConvertToWebauthnKeyList(builder.DeviceKeyListCellData.Keys())
	if len(keyList) == 0 {
		resp.Err = fmt.Errorf("ConvertToWebauthnKeyList err: %s", err.Error())
		return
	}
	//var cidPk []dao.TableCidPk
	cidPk = append([]dao.TableCidPk{}, dao.TableCidPk{
		Cid:             keyList[0].Cid,
		Pk:              keyList[0].PubKey,
		EnableAuthorize: dao.EnableAuthorizeOn,
		Outpoint:        common.OutPoint2String(req.TxHash, uint(builder.Index)),
	})
	if err := b.dbDao.InsertCidPk(cidPk); err != nil {
		resp.Err = fmt.Errorf("InsertCidPk err: %s", err.Error())
		return
	}
	return
}

// add and delete deviceKey
func (b *BlockParser) ActionUpdateDeviceKeyList(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasKeyListCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		log.Warn("not current version transfer account tx")
		return
	}
	log.Info("ActionUpdateDeviceKeyList:", req.BlockNumber, req.TxHash)

	builder, err := witness.WebAuthnKeyListDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("WebAuthnKeyListDataBuilderFromTx err: %s", err.Error())
		return
	}
	keyList := witness.ConvertToWebauthnKeyList(builder.DeviceKeyListCellData.Keys())
	var master witness.WebauthnKey
	var authorize []dao.TableAuthorize
	var cidPks []dao.TableCidPk
	for i := 0; i < len(keyList); i++ {
		var cidPk dao.TableCidPk
		cidPk.Cid = keyList[i].Cid
		cidPk.Pk = keyList[i].PubKey
		if i == 0 {
			master.MinAlgId = keyList[0].MinAlgId
			master.SubAlgId = keyList[0].SubAlgId
			master.Cid = keyList[0].Cid
			master.PubKey = keyList[0].PubKey
			cidPk.Outpoint = common.OutPoint2String(req.TxHash, 0)
			oringinPubkey, _ := witness.GetWebAuthnPubkeyByWitness0(req.Tx.Witnesses[0])
			cidPk.OriginPk = common.Bytes2Hex(oringinPubkey)
			cidPks = append(cidPks, cidPk)
		} else {
			res, err := b.dbDao.GetCidPk(keyList[i].Cid)
			if err != nil {
				resp.Err = fmt.Errorf("GetCidPk err:%s", err.Error())
				return
			}
			if res.Id == 0 {
				cidPks = append(cidPks, cidPk)
			}
		}

		authorize = append(authorize, dao.TableAuthorize{
			MasterAlgId:    common.DasAlgorithmId(master.MinAlgId),
			MasterSubAlgId: common.DasAlgorithmId(master.SubAlgId),
			MasterCid:      master.Cid,
			MasterPk:       master.PubKey,
			SlaveAlgId:     common.DasAlgorithmId(keyList[i].MinAlgId),
			SlaveSubAlgId:  common.DasAlgorithmId(keyList[i].SubAlgId),
			SlaveCid:       keyList[i].Cid,
			SlavePk:        keyList[i].PubKey,
			Outpoint:       common.OutPoint2String(req.TxHash, 0),
		})

	}
	if err = b.dbDao.UpdateAuthorizeByMaster(authorize, cidPks); err != nil {
		resp.Err = fmt.Errorf("UpdateAuthorizeByMaster err:%s", err.Error())
		return
	}
	return
}
