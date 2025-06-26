package letspay

import (
	"fmt"
	"goserver/internal/global/pay/service"
)

type LetsPayConfig struct {
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
type LetsPayOrderRespond struct {
	RetCode   string `json:"retCode"`
	PayUrl    string `json:"payUrl"`
	OrderNo   string `json:"orderNo"`
	PlatOrder string `json:"platOrder"`
	Code      string `json:"code"`
	RetMsg    string `json:"retMsg"`
}

// 提现下单响应
type LetsPayWithdrawRespond struct {
	RetCode    string `json:"retCode"`
	MchTransNo string `json:"mchTransNo"`
	PlatOrder  string `json:"platOrder"`
	Status     string `json:"status"`
	RetMsg     string `json:"retMsg"`
}

// 充值完成响应
type LetsPayRechargeCallback struct {
	MchId       string `json:"mchId"`
	OrderNo     string `json:"orderNo"`
	Product     string `json:"product"`
	Amount      int    `json:"amount"`
	PaySuccTime int64  `json:"paySuccTime"`
	Status      int    `json:"status"`
	Sign        string `json:"sign"`
}

// 核单响应
type LetsPayInspectRespond struct {
	MchId   string `json:"mchId"`
	OrderNo string `json:"orderNo"`
	Amount  int    `json:"amount"`
	Status  int    `json:"status"`
	Sign    string `json:"sign"`
	RetCode string `json:"retCode"`
}

// 代付核单响应
type LetsPayInspectWdRespond struct {
	MchId      string `json:"mchId"`
	MchTransNo string `json:"mchTransNo"`
	Amount     int    `json:"amount"`
	Status     int    `json:"status"`
	Sign       string `json:"sign"`
	RetCode    string `json:"retCode"`
}

func InitConfig(url, checkUrl, appid, md5key, notifyUrl, withdrawNotifyUrl string) {
	Config = &LetsPayConfig{
		Appid:              appid,
		Md5Key:             md5key,
		PayOrderUrl:        fmt.Sprintf("%s%s", url, "/apipay"),
		WithDrawUrl:        fmt.Sprintf("%s%s", url, "/apitrans"),
		PayNotifyUrl:       fmt.Sprintf("%s%s", notifyUrl, "/letspay"),
		WithDrawNotifyUrl:  fmt.Sprintf("%s%s", withdrawNotifyUrl, "/letspay"),
		InspectOrderUrl:    fmt.Sprintf("%s%s", checkUrl, "/qpayorder"),
		InspectWithdrawUrl: fmt.Sprintf("%s%s", checkUrl, "/qtransorder"),
		QueryUrl:           fmt.Sprintf("%s%s", checkUrl, "/qaccount"),
	}
	service.Regist(service.LETSPAY, Config)
}
