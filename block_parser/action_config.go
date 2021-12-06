package block_parser

import (
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
)

func (b *BlockParser) ActionConfigCell(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	configContract, err := core.GetDasContractInfo(common.DasContractNameConfigCellType)
	if err != nil {
		resp.Err = fmt.Errorf("GetDasContractInfo err: %s", err.Error())
		return
	} else if configContract.ContractTypeId != req.Tx.Outputs[0].Type.CodeHash {
		log.Warn("not current version config cell")
		return
	}

	log.Info("ActionConfigCell:", req.TxHash)
	// config cell 更新，重新同步 config cell out point
	if err = b.dasCore.AsyncDasConfigCell(); err != nil {
		resp.Err = fmt.Errorf("AsyncDasConfigCell err: %s", err.Error())
		return
	}
	return
}
