package universalpay

import (
	"goserver/internal/global/pay/service"
)

type UniversalpayConfig struct {
	Appid             string // 服务代码
	PrirsaClient      string // 客户端私钥
	PubrsaClient      string // 客户端公钥
	PubrsaServer      string // 服务端公钥
	PayNotifyUrl      string
	WithdrawNotifyUrl string
}

func InitConfig(appid, prirsaClient, pubrsaClient, pubrsaServer, payNotifyUrl, withdrawNotifyUrl string) {
	Config = &UniversalpayConfig{
		Appid:             appid,
		PrirsaClient:      prirsaClient,
		PubrsaClient:      pubrsaClient,
		PubrsaServer:      pubrsaServer,
		PayNotifyUrl:      payNotifyUrl,
		WithdrawNotifyUrl: withdrawNotifyUrl,
	}
	service.Regist(service.UNIVERSALPAY, Config)
}

type PayOrderRespond struct {
	PlatformOrderNo string `json:"platformOrderNo"`
	OrderNo         string `json:"orderNo"`
	Status          int    `json:"status"`
	LinkUrl         string `json:"linkUrl"`
}

type PayInspectRespond struct {
	OrderNo         string  `json:"orderNo"`
	PlatformOrderNo string  `json:"platformOrderNo"`
	Status          int     `json:"status"` // 状态 0.支付失败 1.支付成功 2.待支付 3.支付中
	Amount          float64 `json:"amount"`
	Fee             float64 `json:"fee"`
	LinkUrl         string  `json:"linkUrl"`
}

type WithdrawOrderRespond struct {
	PlatformOrderNo string `json:"platformOrderNo"`
	OrderNo         string `json:"orderNo"`
	Status          int    `json:"status"`
}

type WithdrawInspectRespond struct {
	PlatformOrderNo string  `json:"platformOrderNo"`
	OrderNo         string  `json:"orderNo"`
	UtrNo           string  `json:"utrNo"`
	Status          int     `json:"status"` // 状态 0.支付失败 1.支付成功 2.待支付 3.支付中
	Amount          float64 `json:"amount"`
	Fee             float64 `json:"fee"`
}

type BalanceRespond struct {
	TotalAmount   float64 `json:"totalAmount"`
	DeductAmount  float64 `json:"deductAmount"`
	PaymentAmount float64 `json:"paymentAmount"`
	FrozenAmount  float64 `json:"frozenAmount"`
}

type PayRechargeCallback struct {
	PlatformOrderNo string `json:"platformOrderNo"`
	OrderNo         string `json:"orderNo"`
	Status          string `json:"status"` // 状态 0.支付失败 1.支付成功 2.待支付 3.支付中
	Amount          string `json:"amount"`
	Fee             string `json:"fee"`
	LinkUrl         string `json:"linkUrl"`
	UtrNo           string `json:"utrNo"`
}

type WithdrawCallback struct {
	PlatformOrderNo string `json:"platformOrderNo"`
	OrderNo         string `json:"orderNo"`
	Status          string `json:"status"` // 状态 0.支付失败 1.支付成功 2.待支付 3.支付中
	Amount          string `json:"amount"`
	Fee             string `json:"fee"`
	UtrNo           string `json:"utrNo"`
	ErrorMsg        string `json:"errorMsg"`
}
