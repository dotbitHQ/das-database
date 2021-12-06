package lark

import (
	"fmt"
	"github.com/shopspring/decimal"
	"time"
)

type ParamsDoOrderNotify struct {
	WxKey        string
	Action       string
	Account      string
	OrderId      string
	PayChainType uint
	PayType      string
	PayAmount    decimal.Decimal
	actionName   string
	payName      string
	Hash         string
	WebhookUrl   string
}

// DoOrderNotifyLark 订单通知
func DoOrderNotifyLark(p ParamsDoOrderNotify) {
	switch p.Action {
	case "apply_register":
		p.actionName = "注册下单"
	case "renew_account":
		p.actionName = "续费下单"
	default:
		p.actionName = "未知 Action"
	}
	switch p.PayChainType {
	case 0:
		p.payName = "CKB支付"
		p.PayAmount = p.PayAmount.DivRound(decimal.New(1, 8), 8)
	case 1:
		p.payName = "ETH支付"
		p.PayAmount = p.PayAmount.DivRound(decimal.New(1, 18), 18)
	case 2:
		p.payName = "BTC支付"
		p.PayAmount = p.PayAmount.DivRound(decimal.New(1, 8), 8)
	case 3:
		p.payName = "TRON支付"
		p.PayAmount = p.PayAmount.DivRound(decimal.New(1, 6), 6)
	case 4:
		p.payName = "微信支付"
		p.PayAmount = p.PayAmount.DivRound(decimal.New(1, 2), 2)
	default:
		p.payName = "未知支付方式"
	}
	msg := `> 账号：%s
> 订单号：%s
> 支付方式：%s
> 订单金额：%s
> 时间：%s
> 交易哈希：%s`
	msg = fmt.Sprintf(msg, p.Account, p.OrderId, p.payName, p.PayAmount.String(), time.Now().Format("2006-01-02 15:04:05"), p.Hash)
	err := sendLarkTextNotify(p.WebhookUrl, p.actionName, msg)
	if err != nil {
		log.Error("DoOrderNotifyLark err: ", err.Error())
	}
}

type ParamsDoRegisterNotify struct {
	WxKey      string
	Action     string
	Account    string
	OrderId    string
	Hash       string
	ActionName string
	WebhookUrl string
}

// DoRegisterNotifyLark 注册通知
func DoRegisterNotifyLark(p ParamsDoRegisterNotify) {
	switch p.Action {
	case "apply_register":
		p.ActionName = "申请注册"
	case "pre_register":
		p.ActionName = "预注册"
	case "propose":
		p.ActionName = "提案"
	case "confirm_proposal":
		p.ActionName = `确认提案`
	default:
		p.ActionName = "未知 Action"
	}
	msg := `> 注册账号：%s
> 订单号：%s
> 时间：%s
> 交易哈希：%s`
	msg = fmt.Sprintf(msg, p.Account, p.OrderId, time.Now().Format("2006-01-02 15:04:05"), p.Hash)
	err := sendLarkTextNotify(p.WebhookUrl, p.ActionName, msg)
	if err != nil {
		log.Error("DoRegisterNotifyLark err: ", err.Error())
	}
}
