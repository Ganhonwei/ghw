package ramapay

import (
	"fmt"
	"goserver/internal/global/pay/service"
)

type RamaPayConfig struct {
	Merchant          string // 商户号
	ServerPubRSA      string // 服务端公钥
	ClientPriRSA      string // 客户端私钥
	PayOrderUrl       string // 代收地址
	WithDrawUrl       string // 代付地址
	PayNotifyUrl      string // 支付通知地址
	WithDrawNotifyUrl string // 提现通知地址
}

type RamaOrderRespond struct {
	Code    int    // 状态码
	Message string // 返回的请求信息
	Data    RamaPayData
}

// 代收回调
type RamaRechargeCallback struct {
	MerchantCode    string `json:"merchant_code"`     //商户号
	OrderId         string `json:"order_no"`          //平台订单号
	MerchantOrderNo string `json:"merchant_order_no"` //商户订单号
	Amount          string `json:"amount"`            //金额
	Status          int32  `json:"status"`            //订单状态
	ErrorMessage    string `json:"error_message"`     //错误信息
	Timestamp       int64  `json:"timestamp"`         //时间
	Sign            string `json:"sign"`              //签名
}

type RamaPayData struct {
	Order           string `json:"order_no"` //系统生成的平台订单号
	MerchantOrderNo string `json:"merchant_order_no"`
	MerchantCode    string `json:"merchant_code"` // 商户号
	Amount          string `json:"amount" `       // 金额
	PayLink         string `json:"pay_link" `     // 支付二维码
	Sign            string `json:"sign" `         // 签名
}

func InitConfig(url, appid, serverPubRSA, clientPriRSA, notifyUrl, withdrawNotifyUrl string) {
	Config = &RamaPayConfig{
		Merchant:          appid,
		ServerPubRSA:      serverPubRSA,
		ClientPriRSA:      clientPriRSA,
		PayOrderUrl:       fmt.Sprintf("%s%s", url, "/payment-gateway/payment"),
		WithDrawUrl:       fmt.Sprintf("%s%s", url, "/disbursement/cash"),
		PayNotifyUrl:      fmt.Sprintf("%s%s", notifyUrl, "/rama"),
		WithDrawNotifyUrl: fmt.Sprintf("%s%s", withdrawNotifyUrl, "/rama"),
	}
	service.Regist(service.RAMAPAY, Config)
}
