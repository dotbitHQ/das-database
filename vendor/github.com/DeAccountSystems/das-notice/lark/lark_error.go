package lark

import (
	"fmt"
	"time"
)

//=================== payment error ===================

type ParamsDoOrderNotRefundedNotify struct {
	Count      int64
	WebhookUrl string
}

func DoOrderNotRefundedNotifyLark(p ParamsDoOrderNotRefundedNotify) {
	msg := fmt.Sprintf("未退款订单数：%d", p.Count)
	err := sendLarkTextNotify(p.WebhookUrl, "", msg)
	if err != nil {
		log.Error("DoOrderNotRefundedNotifyLark err: ", err.Error())
	}
}

type ParamsDoErrorHandleNotify struct {
	FuncName   string
	KeyInfo    string
	ErrInfo    string
	WebhookUrl string
}

func getLarkTextNotifyStr(funcName, keyInfo, errInfo string) string {
	msg := fmt.Sprintf(`方法名：%s
关键信息：%s
错误信息：%s`, funcName, keyInfo, errInfo)
	return msg
}

func DoHedgeDepositNotifyLark(p ParamsDoErrorHandleNotify) {
	msg := getLarkTextNotifyStr(p.FuncName, p.KeyInfo, p.ErrInfo)
	err := sendLarkTextNotify(p.WebhookUrl, "对冲失败", msg)
	if err != nil {
		log.Error("DoHedgeDepositNotifyLark err: ", err.Error())
	}
}
func DoUpdateOrderNotifyLark(p ParamsDoErrorHandleNotify) {
	msg := getLarkTextNotifyStr(p.FuncName, p.KeyInfo, p.ErrInfo)
	err := sendLarkTextNotify(p.WebhookUrl, "更新订单CkbHash失败", msg)
	if err != nil {
		log.Error("DoUpdateOrderNotifyLark err: ", err.Error())
	}
}
func DoPayCallbackNotifyLark(p ParamsDoErrorHandleNotify) {
	msg := getLarkTextNotifyStr(p.FuncName, p.KeyInfo, p.ErrInfo)
	err := sendLarkTextNotify(p.WebhookUrl, "支付回调发交易失败", msg)
	if err != nil {
		log.Error("DoPayCallbackNotifyLark err: ", err.Error())
	}
}
func DoNodeMonitorNotifyLark(p ParamsDoErrorHandleNotify) {
	msg := getLarkTextNotifyStr(p.FuncName, p.KeyInfo, p.ErrInfo)
	err := sendLarkTextNotify(p.WebhookUrl, "节点监控", msg)
	if err != nil {
		log.Error("DoNodeMonitorNotifyLark err: ", err.Error())
	}
}

//=================== register error ===================

type ParamsDoOrderTxRejectedNotify struct {
	Account    string
	Status     int
	SinceMin   float64
	WebhookUrl string
}

func DoPreRegisterNotifyLark(p ParamsDoErrorHandleNotify) {
	msg := getLarkTextNotifyStr(p.FuncName, p.KeyInfo, p.ErrInfo)
	err := sendLarkTextNotify(p.WebhookUrl, "发预注册交易失败", msg)
	if err != nil {
		log.Error("DoPreRegisterNotifyLark err: ", err.Error())
	}
}

func DoOrderTxRejectedNotifyLark(p ParamsDoOrderTxRejectedNotify) {
	action := "申请注册"
	if p.Status == 3 {
		action = "预注册"
	}
	msg := `> 账号：%s
> 步骤: %s
> 时间：%.2f 分钟前`
	msg = fmt.Sprintf(msg, p.Account, action, p.SinceMin)
	err := sendLarkTextNotify(p.WebhookUrl, "Rejected 交易监听", msg)
	if err != nil {
		log.Error("DoOrderTxRejectedNotifyLark err: ", err.Error())
	}
}

//=================== parser error ===================

type ParamsDoBlockParserNotify struct {
	Action      string
	BlockNumber uint64
	Hash        string
	WebhookUrl  string
}

func DoBlockParserNotifyLark(p ParamsDoBlockParserNotify) {
	msg := `> 高度：%d
> 步骤：%s
> 时间：%s
> 交易哈希：%s`
	msg = fmt.Sprintf(msg, p.BlockNumber, p.Action, time.Now().Format("2006-01-02 15:04:05"), p.Hash)
	err := sendLarkTextNotify(p.WebhookUrl, "区块监听", msg)
	if err != nil {
		log.Error("DoBlockParserNotifyLark err: ", err.Error())
	}
}
