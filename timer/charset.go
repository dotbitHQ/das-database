package timer

import (
	"das_database/config"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"time"
)

func (p *ParserTimer) RunFixCharset() {
	if config.Cfg.Server.FixCharset {
		tickerCharset := time.NewTicker(time.Second * 10)
		p.Wg.Add(1)
		go func() {
			ok := false
			for {
				select {
				case <-tickerCharset.C:
					ok = p.doFixCharset()
				case <-p.Ctx.Done():
					p.Wg.Done()
					return
				}
				if ok {
					break
				}
			}
		}()
	}
}

func (p *ParserTimer) doFixCharset() bool {
	list, err := p.DbDao.GetNeedFixCharsetAccountList()
	if err != nil {
		log.Error("GetNeedFixCharsetAccountList err: %s", err.Error())
		return false
	}
	if len(list) == 0 {
		log.Info("doFixCharset ok")
		return true
	}

	var hashList = make(map[string]struct{})
	for _, v := range list {
		hashList[v.ConfirmProposalHash] = struct{}{}
	}

	var accCharset = make(map[string]uint64)
	for k, _ := range hashList {
		tx, err := p.DasCore.Client().GetTransaction(p.Ctx, types.HexToHash(k))
		if err != nil {
			log.Error("GetTransaction err: %s", err.Error())
			continue
		}
		accMap, err := witness.AccountIdCellDataBuilderFromTx(tx.Transaction, common.DataTypeNew)
		if err != nil {
			log.Error("AccountIdCellDataBuilderFromTx err: %s", err.Error())
			continue
		}
		for _, v := range accMap {
			charsetNum := common.ConvertAccountCharsToCharsetNum(v.AccountChars)
			accCharset[v.AccountId] = charsetNum
		}
	}
	//
	if err := p.DbDao.UpdateAccountCharsetNum(accCharset); err != nil {
		log.Error("UpdateAccountCharsetNum err: %s", err.Error())
	}

	return false
}
