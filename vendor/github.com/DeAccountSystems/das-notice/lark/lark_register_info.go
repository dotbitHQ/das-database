package lark

import (
	"fmt"
	"github.com/scorpiotzh/mylog"
	"time"
)

var log = mylog.NewLogger("lark", mylog.LevelDebug)

type ParamsDoRegisterInfoNotify struct {
	AccountNum    int64
	OwnerNum      int64
	ApplyNum      int64
	PreNum        int64
	ProNum        int64
	ConfirmProNum int64
	WebhookUrl    string
}

func DoRegisterInfoNotifyLark(p ParamsDoRegisterInfoNotify) {
	msg := `- 已注册账号数：%d
- 已注册owner总数：%d
- 申请注册中数量：%d
- 预注册数中量：%d
- 提案中数量：%d
- 确认提案中数量：%d
- 时间：%s`
	msg = fmt.Sprintf(msg, p.AccountNum, p.OwnerNum, p.ApplyNum, p.PreNum, p.ProNum, p.ConfirmProNum, time.Now().Format("2006-01-02 15:04:05"))
	err := sendLarkTextNotify(p.WebhookUrl, "注册信息统计", msg)
	if err != nil {
		log.Error("DoRegisterInfoNotifyLark err: ", err.Error())
	}
}

type ParamsDoNormalCellNotify struct {
	Count      int
	Capacity   string
	WebhookUrl string
}

func DoNormalCellNotifyLark(p ParamsDoNormalCellNotify) {
	msg := `- 数量：%d
- Capacity: %s
- 时间：%s`
	msg = fmt.Sprintf(msg, p.Count, p.Capacity, time.Now().Format("2006-01-02 15:04:05"))
	err := sendLarkTextNotify(p.WebhookUrl, "剩余零散的NormalCell", msg)
	if err != nil {
		log.Error("DoNormalCellNotifyLark err: ", err.Error())
	}
}
