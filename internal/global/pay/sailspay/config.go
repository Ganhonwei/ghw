package sailspay

import (
	"fmt"
	"goserver/internal/global/pay/service"
)

type SailspayConfig struct {
	MachId            string // 商户号
	Md5Key            string // MD5秘钥
	PayOrderUrl       string // 代收地址
	WithDrawUrl       string // 代付地址
	PayNotifyUrl      string // 支付通知地址
	WithDrawNotifyUrl string // 提现通知地址
}

// 下单响应
type OrderRespond struct {
	Code       int    `json:"code"`
	Msg        string `json:"msg"`
	PaymentUrl string `json:"paymentUrl"`
	PayOrderId string `json:"payOrderId"`
}

// 提现响应
type WithdrawRespond struct {
	Code       int    `json:"code"`
	Msg        string `json:"msg"`
	PayOrderId string `json:"payOrderId"`
}

// 充值完成响应
type RechargeCallback struct {
	Amount     int    `json:"amount"`     // 客户实际支付金额，单位分
	MerchantId int    `json:"merchantId"` // 商户ID
	OrderId    string `json:"orderId"`    // 商户请求代收时提交的订单号
	PayOrderId string `json:"payOrderId"` // 商户请求代收时，在返回结果中返回的支付网关订单号
	Status     int    `json:"status"`     // 订单状态，1:支付成功，2:支付失败，只有在成功或失败时才会回调。除了1支付成功外，其它均为支付失败。
	Timestamp  int64  `json:"timestamp"`  // 时间戳，精确到毫秒
	Sign       string `json:"sign"`       // 签名
}

// 代付回调
type WithdrawCallback struct {
	Amount     int    `json:"amount"`     // 支付金额，单位分
	MerchantId int    `json:"merchantId"` // 商户ID
	OrderId    string `json:"orderId"`    // 商户请求代收时提交的订单号
	PayOrderId string `json:"payOrderId"` // 商户请求代收时，在返回结果中返回的支付网关订单号
	Status     int    `json:"status"`     // 订单状态，** 1:代付成功，2:代付失败，只有成功或失败时才会推送 **
	Timestamp  int64  `json:"timestamp"`  // 时间戳，精确到毫秒
	Sign       string `json:"sign"`       // 签名
}

func InitConfig(url, appid, md5key, notifyUrl, withdrawNotifyUrl string) {
	Config = &SailspayConfig{
		MachId:            appid,
		Md5Key:            md5key,
		PayOrderUrl:       fmt.Sprintf("%s%s", url, "/payment/createOrder"),
		WithDrawUrl:       fmt.Sprintf("%s%s", url, "/payout/createOrder"),
		PayNotifyUrl:      fmt.Sprintf("%s%s", notifyUrl, "/sailspay"),
		WithDrawNotifyUrl: fmt.Sprintf("%s%s", withdrawNotifyUrl, "/sailspay"),
	}
	service.Regist(service.SAILSPAY, Config)
}
