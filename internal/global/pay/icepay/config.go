package icepay

import (
	"fmt"
	"goserver/internal/global/pay/service"
)

type IcePayConfig struct {
	Appid              string // 商户号
	Md5Key             string // 支付MD5秘钥
	PayOrderUrl        string // 代收地址
	WithDrawUrl        string // 代付地址
	PayNotifyUrl       string // 支付通知地址
	WithDrawNotifyUrl  string // 提现通知地址
	InspectOrderUrl    string // 查单接口
	InspectWithdrawUrl string // 代付查询
	QueryUrl           string // 余额查询
}

// 下单响应
type IcePayOrderRespond struct {
	Code       int    `json:"code"`
	Msg        string `json:"msg"`
	PaymentUrl string `json:"paymentUrl"`
	PayOrderId string `json:"payOrderId"`
}

// 提现下单响应
type IcePayWithdrawRespond struct {
	Code       int    `json:"code"`
	Msg        string `json:"msg"`
	PayOrderId string `json:"payOrderId"`
}

// 充值完成响应
type IcePayRechargeCallback struct {
	Amount     int    `json:"amount"`
	MerchantId int    `json:"merchantId"`
	OrderId    string `json:"orderId"`
	PayOrderId string `json:"payOrderId"`
	Status     int    `json:"status"`
	Timestamp  int64  `json:"timestamp"`
	Sign       string `json:"sign"`
}

// 核单响应
type InspectRespond struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data InspectData `json:"data"`
}

type InspectData struct {
	Amount     int    `json:"amount"`
	MerchantId string `json:"merchantId"`
	OrderId    string `json:"orderId"`
	PayOrderId string `json:"payOrderId"`
	Status     int    `json:"status"`
	StatusDesc string `json:"statusDesc"`
	Timestamp  int    `json:"timestamp"`
	Sign       string `json:"sign"`
}

func InitConfig(url, checkUrl, appid, md5key, notifyUrl, withdrawNotifyUrl string) {
	Config = &IcePayConfig{
		Appid:              appid,
		Md5Key:             md5key,
		PayOrderUrl:        fmt.Sprintf("%s%s", url, "/api/payment/createOrder"),
		WithDrawUrl:        fmt.Sprintf("%s%s", url, "/api/payout/createOrder"),
		PayNotifyUrl:       fmt.Sprintf("%s%s", notifyUrl, "/icepay"),
		WithDrawNotifyUrl:  fmt.Sprintf("%s%s", withdrawNotifyUrl, "/icepay"),
		InspectOrderUrl:    fmt.Sprintf("%s%s", checkUrl, "/api/payment/status"),
		InspectWithdrawUrl: fmt.Sprintf("%s%s", checkUrl, "/api/payout/status"),
		QueryUrl:           fmt.Sprintf("%s%s", url, "/api/payout/balance"),
	}
	service.Regist(service.ICEPAY, Config)
}
