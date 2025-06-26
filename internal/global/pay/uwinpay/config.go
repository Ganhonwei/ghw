package uwinpay

import (
	"fmt"
	"goserver/internal/global/pay/service"
)

type UWinPayConfig struct {
	Merchant           string // 商户号
	ServerPubRSA       string // 服务端公钥
	ClientPriRSA       string // 客户端私钥
	PayOrderUrl        string // 代收地址
	WithDrawUrl        string // 代付地址
	PayNotifyUrl       string // 支付通知地址
	WithDrawNotifyUrl  string // 提现通知地址
	InspectOrderUrl    string // 查单接口
	InspectWithdrawUrl string // 代付查询
	QueryUrl           string // 余额查询
}

type UWinOrderRespond struct {
	Code    int    // 状态码
	Message string // 返回的请求信息
	Data    UWinPayData
}

// 代收回调
type UWinRechargeCallback struct {
	MerchantCode    string `json:"merchant_code"`     //商户号
	OrderId         string `json:"order_no"`          //平台订单号
	MerchantOrderNo string `json:"merchant_order_no"` //商户订单号
	Amount          string `json:"amount"`            //金额
	Status          int32  `json:"status"`            //订单状态
	ErrorMessage    string `json:"error_message"`     //错误信息
	Timestamp       int64  `json:"timestamp"`         //时间
	Sign            string `json:"sign"`              //签名
}

type UWinPayData struct {
	Order           string `json:"order_no"` //系统生成的平台订单号
	MerchantOrderNo string `json:"merchant_order_no"`
	MerchantCode    string `json:"merchant_code"` // 商户号
	Amount          string `json:"amount" `       // 金额
	PayLink         string `json:"pay_link" `     // 支付二维码
	Sign            string `json:"sign" `         // 签名
}

// 核单响应
type UWinPayInspectRespond struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    UWinPayMessage `json:"data"`
}

type UWinPayMessage struct {
	OrderNo         string `json:"order_no"`
	MerchantOrderNo string `json:"merchant_order_no"`
	MerchantCode    string `json:"merchant_code"`
	Amount          string `json:"amount"`
	Status          int    `json:"status"`
	AccountCode     string `json:"account_code"`
	PayerName       string `json:"payer_name"`
	Timestamp       int    `json:"timestamp"`
	Sign            int    `json:"sign"`
}

type UWinBalanceRespond struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    BalanceMessage `json:"data"`
}

type BalanceMessage struct {
	MerchantCode      string `json:"merchant_code"`
	AvailableBalance  string `json:"available_balance"`
	UnsettledAmount   string `json:"unsettled_amount"`
	WithdrawingAmount string `json:"withdrawing_amount"`
	FrozenAmount      string `json:"frozen_amount"`
	Sign              string `json:"sign"`
}

func InitConfig(url, appid, serverPubRSA, clientPriRSA, notifyUrl, withdrawNotifyUrl string) {
	Config = &UWinPayConfig{
		Merchant:           appid,
		ServerPubRSA:       serverPubRSA,
		ClientPriRSA:       clientPriRSA,
		PayOrderUrl:        fmt.Sprintf("%s%s", url, "/pay"),
		WithDrawUrl:        fmt.Sprintf("%s%s", url, "/transfer"),
		PayNotifyUrl:       fmt.Sprintf("%s%s", notifyUrl, "/uwin"),
		WithDrawNotifyUrl:  fmt.Sprintf("%s%s", withdrawNotifyUrl, "/uwin"),
		InspectOrderUrl:    fmt.Sprintf("%s%s", url, "/pay/query"),
		InspectWithdrawUrl: fmt.Sprintf("%s%s", url, "/transfer/query"),
		QueryUrl:           fmt.Sprintf("%s%s", url, "/balance"),
	}
	service.Regist(service.UWINPAY, Config)
}
