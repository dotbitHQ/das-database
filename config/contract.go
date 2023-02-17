package config

import (
	"context"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
)

var contractNames = []common.DasContractName{
	common.DasContractNameApplyRegisterCellType,
	common.DasContractNamePreAccountCellType,
	common.DasContractNameProposalCellType,
	common.DasContractNameConfigCellType,
	common.DasContractNameAccountCellType,
	common.DasContractNameAccountSaleCellType,
	common.DASContractNameSubAccountCellType,
	common.DASContractNameOfferCellType,
	common.DasContractNameBalanceCellType,
	common.DasContractNameIncomeCellType,
	common.DasContractNameReverseRecordCellType,
	common.DASContractNameEip712LibCellType,
}

func CheckContractVersion(dasCore *core.DasCore, cancel context.CancelFunc) error {
	if dasCore == nil {
		return fmt.Errorf("dasCore is nil")
	}
	for _, v := range contractNames {
		defaultVersion, chainVersion, err := dasCore.CheckContractVersion(v)
		if err != nil {
			if err == core.ErrContractMajorVersionDiff {
				log.Errorf("contract[%s] version diff, chain[%s], service[%s].", v, chainVersion, defaultVersion)
				log.Error("Please update the service. [https://github.com/dotbitHQ/das-database]")
				if cancel != nil && !Cfg.Server.NotExit {
					cancel()
				}
				return err
			}
			return fmt.Errorf("CheckContractVersion err: %s", err.Error())
		}
	}
	return nil
}
