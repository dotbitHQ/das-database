package block_parser

import (
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
)

func (b *BlockParser) ActionCreateIncome(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameIncomeCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if isCV {
		log.Warn("not current version create income tx")
		return
	}

	log.Info("ActionCreateIncome:", req.BlockNumber, req.TxHash)

	return
}
