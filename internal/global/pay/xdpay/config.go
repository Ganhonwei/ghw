package xdpay

import (
	"fmt"
	"goserver/internal/global/pay/service"
)

type XDPayConfig struct {
	MachId            string // 商户号
	PayCode           string // 代收通道
	WithdrawCode      string // 代付通道
	Md5Key            string // 支付MD5秘钥
	PayOrderUrl       string // 代收地址
	WithDrawUrl       string // 代付地址
	PayNotifyUrl      string // 支付通知地址
	WithDrawNotifyUrl string // 提现通知地址
}

// 下单响应
type XDOrderRespond struct {
	Code    int          `json:"code"`
	Success bool         `json:"success"`
	Msg     string       `json:"msg"`
	Desc    string       `json:"desc"`
	Data    *XDOrderData `json:"data"`
}
type XDOrderData struct {
	Url string `json:"url"`
}

// 提现响应
type XDWithdrawRespond struct {
	Code    int             `json:"code"`
	Success bool            `json:"success"`
	Msg     string          `json:"msg"`
	Desc    string          `json:"desc"`
	Data    *XDWithdrawData `json:"data"`
}
type XDWithdrawData struct {
	PlatOrderId string `json:"platOrderId"`
}

// 回调
type OrderCallback struct {
	PlatOrderId string `json:"platOrderId"` // 平台订单号
	OrderId     string `json:"orderId"`     // 商户订单号
	Amount      int    `json:"amount"`      // 实际支付金额
	Status      int    `json:"status"`      // 交易状态
	Reverse     bool   `json:"reverse"`     //是否反转订单
	Remark      string `json:"remark"`      // 付款失败原因
	Sign        string `json:"sign"`        // 签名
}

func InitConfig(url, appid, paycode, withdrawcode, md5key, notifyUrl, withdrawNotifyUrl string) {
	Config = &XDPayConfig{
		MachId:            appid,
		PayCode:           paycode,
		WithdrawCode:      withdrawcode,
		Md5Key:            md5key,
		PayOrderUrl:       fmt.Sprintf("%s%s", url, "/collect/create"),
		WithDrawUrl:       fmt.Sprintf("%s%s", url, "/pay/create"),
		PayNotifyUrl:      fmt.Sprintf("%s%s", notifyUrl, "/xdpay"),
		WithDrawNotifyUrl: fmt.Sprintf("%s%s", withdrawNotifyUrl, "/xdpay"),
	}
	service.Regist(service.XDPAY, Config)
}
