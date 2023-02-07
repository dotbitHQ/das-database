package snapshot

import (
	"context"
	"das_database/dao"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/scorpiotzh/mylog"
	"sync"
)

var log = mylog.NewLogger("snapshot", mylog.LevelDebug)

type ToolSnapshot struct {
	Ctx            context.Context
	Wg             *sync.WaitGroup
	DbDao          *dao.DbDao
	DasCore        *core.DasCore
	ConcurrencyNum uint64
	ConfirmNum     uint64

	currentBlockNumber uint64
	parserType         dao.ParserType
}
