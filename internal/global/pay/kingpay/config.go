package kingpay

import (
	"fmt"
	"goserver/internal/global/pay/service"
)

type KingPayConfig struct {
	Appid             string // 商户号
	ApplicationId     string // 应用Id
	ServerPubRSA      string // 服务端公钥
	ClientPriRSA      string // 客户端私钥
	PayOrderUrl       string // 代收地址
	WithDrawUrl       string // 代付地址
	PayNotifyUrl      string // 支付通知地址
	WithDrawNotifyUrl string // 提现通知地址
}

// 下单&&提现响应
type KingOrderRespond struct {
	Code       string `json:"code"`    //响应状态码
	Message    string `json:"message"` //响应结果描述
	Timestamp  int64  `json:"timestamp"`
	Sign       string `json:"sign"`
	OrderId    string `json:"orderId"`
	OutOrderId string `json:"outOrderId"`
	Status     int32  `json:"status"`
	StatusDesc string `json:"statusDesc"`
	PayUrl     string `json:"payUrl"`
	ExtendInfo string `json:"extendInfo"`
}

// 代收回调
type KingRechargeCallback struct {
	Timestamp  int64  `json:"timestamp"`  //时间戳
	OrderId    string `json:"orderId"`    //平台订单号
	OutOrderId string `json:"outOrderId"` //商户订单号
	Status     int32  `json:"status"`     //状态
	StatusDesc string `json:"statusDesc"` //状态描述
	Amount     string `json:"amount"`     //金额
	TradeFee   string `json:"tradeFee"`   //手续费
	PayTime    int64  `json:"payTime"`    //支付时间
	Sign       string `json:"sign"`       //签名
}

// 代付回调
type KingWithdrawCallback struct {
	Timestamp      int64  `json:"timestamp"`      //时间戳
	OrderId        string `json:"orderId"`        //平台订单号
	OutOrderId     string `json:"outOrderId"`     //商户订单号
	Status         int32  `json:"status"`         //状态
	StatusDesc     string `json:"statusDesc"`     //状态描述
	WithdrawAmount string `json:"withdrawAmount"` //金额
	TransferTime   int64  `json:"transferTime"`   //支付时间
	Sign           string `json:"sign"`           //签名
}

// 支付回调响应
type KingPayResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func InitConfig(url, appid, applicationId, serverPubRSA, clientPriRSA, notifyUrl, withdrawNotifyUrl string) {
	Config = &KingPayConfig{
		Appid:             appid,
		ApplicationId:     applicationId,
		ServerPubRSA:      serverPubRSA,
		ClientPriRSA:      clientPriRSA,
		PayOrderUrl:       fmt.Sprintf("%s%s", url, "/trade/preorder"),
		WithDrawUrl:       fmt.Sprintf("%s%s", url, "/trade/withdraw"),
		PayNotifyUrl:      fmt.Sprintf("%s%s", notifyUrl, "/king"),
		WithDrawNotifyUrl: fmt.Sprintf("%s%s", withdrawNotifyUrl, "/king"),
	}
	service.Regist(service.KINGPAY, Config)
}
