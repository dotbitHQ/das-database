package block_parser

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"das_database/dao"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/witness"
	"math/big"
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
	log.Info("ActionUpdateDeviceKeyList err:", req.BlockNumber, req.TxHash)

	builder, err := witness.WebAuthnKeyListDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("WebAuthnKeyListDataBuilderFromTx err: %s", err.Error())
		return
	}
	log.Info("args: ", common.Bytes2Hex(req.Tx.Outputs[0].Lock.Args))
	ownerHex, _, err := b.dasCore.Daf().ArgsToHex(req.Tx.Outputs[0].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}
	var masterCidPk1 dao.TableCidPk

	masterCidPk1.Cid = common.Bytes2Hex(ownerHex.AddressPayload[:10])
	masterCidPk1.Pk = common.Bytes2Hex(ownerHex.AddressPayload[10:])
	masterCidPk1.Outpoint = common.OutPoint2String(req.TxHash, 0)
	//todo GetWebAuthnSignByWitnessArgs
	webauthnSignLv, err := witness.GetWebAuthnSignLvByWitness0(req.Tx.Witnesses[0])
	if err != nil {
		resp.Err = fmt.Errorf("GetWebAuthnSignLvByWitness0 err:%s", err.Error())
		return
	}

	if webauthnSignLv.PkIndex == 255 {
		masterCidPk1.OriginPk = webauthnSignLv.PubKey
	}

	var pubKey ecdsa.PublicKey
	pubKey.Curve = elliptic.P256()
	pubKey.X = new(big.Int).SetBytes(common.Hex2Bytes(webauthnSignLv.PubKey)[:32])
	pubKey.Y = new(big.Int).SetBytes(common.Hex2Bytes(webauthnSignLv.PubKey)[32:])
	signAddrPk1 := common.CaculatePk1(&pubKey)
	keyList := witness.ConvertToWebauthnKeyList(builder.DeviceKeyListCellData.Keys())

	//var authorize []dao.TableAuthorize
	//update master cid1, pk1, originPk
	//update slave cid1 pk1
	//a [a,b] => [a,b,c]
	var slaveCidPks []dao.TableCidPk
	var slaveCidPksSign dao.TableCidPk
	var authorize []dao.TableAuthorize
	for i := 0; i < len(keyList); i++ {
		var slaveCidPk dao.TableCidPk
		cid1 := keyList[i].Cid
		pk1 := keyList[i].PubKey
		//非master
		if cid1 != masterCidPk1.Cid {
			slaveCidPk.Cid = keyList[i].Cid
			slaveCidPk.Pk = keyList[i].PubKey
			if common.Bytes2Hex(signAddrPk1) == pk1 {
				slaveCidPk.OriginPk = webauthnSignLv.PubKey
				slaveCidPksSign = slaveCidPk
			} else {
				slaveCidPks = append(slaveCidPks, slaveCidPk)
			}

		}
		authorize = append(authorize, dao.TableAuthorize{
			MasterAlgId:    common.DasAlgorithmIdWebauthn,
			MasterSubAlgId: common.DasAlgorithmId(7),
			MasterCid:      masterCidPk1.Cid,
			MasterPk:       masterCidPk1.Pk,
			SlaveAlgId:     common.DasAlgorithmId(keyList[i].MinAlgId),
			SlaveSubAlgId:  common.DasAlgorithmId(keyList[i].SubAlgId),
			SlaveCid:       keyList[i].Cid,
			SlavePk:        keyList[i].PubKey,
			Outpoint:       common.OutPoint2String(req.TxHash, 0),
		})
	}
	if err = b.dbDao.UpdateAuthorizeByMaster(authorize, masterCidPk1, slaveCidPksSign, slaveCidPks); err != nil {
		resp.Err = fmt.Errorf("UpdateAuthorizeByMaster err: %s", err.Error())
		return
	}
	return
}
