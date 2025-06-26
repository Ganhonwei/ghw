package mepay

import (
	"fmt"
	"goserver/internal/global/pay/service"
)

type PayConfig struct {
	Merchant           string // 商户号
	AppId              string // 应用ID
	ServerPubRSA       string // 服务端公钥
	ClientPriRSA       string // 客户端私钥
	MD5Key             string // md5key
	PayOrderUrl        string // 代收地址
	WithDrawUrl        string // 代付地址
	PayNotifyUrl       string // 支付通知地址
	WithDrawNotifyUrl  string // 提现通知地址
	InspectOrderUrl    string // 查单接口
	InspectWithdrawUrl string // 代付查询
	QueryUrl           string // 余额查询
}

type PayOrderRespond struct {
	Code int                 `json:"code"`
	Msg  string              `json:"msg"`
	Data PayOrderRespondData `json:"data"`
	Sign string              `json:"sign"`
}

type PayOrderRespondData struct {
	PayOrderId     string `json:"payOrderId"`
	MchOrderNo     string `json:"mchOrderNo"`
	OrderState     int    `json:"orderState"`
	IsvPayDataType string `json:"isvPayDataType"`
	PayDataType    string `json:"payDataType"`
	PayData        string `json:"payData"`
	ErrCode        string `json:"errCode"`
	ErrMsg         string `json:"errMsg"`
}

type PayWithdrawRespond struct {
	Code int                    `json:"code"`
	Msg  string                 `json:"msg"`
	Sign string                 `json:"sign"`
	Data PayWithdrawRespondData `json:"data"`
}

type PayWithdrawRespondData struct {
	TransferId     string `json:"transferId"`
	MchOrderNo     string `json:"mchOrderNo"`
	State          int    `json:"state"`
	ChannelOrderNo string `json:"channelOrderNo"`
	ErrCode        string `json:"errCode"`
	ErrMsg         string `json:"errMsg"`
}

// 代收回调
type PayRechargeCallback struct {
	MerchantCode    string `json:"merchant_code"`     //商户号
	OrderId         string `json:"order_no"`          //平台订单号
	MerchantOrderNo string `json:"merchant_order_no"` //商户订单号
	Amount          string `json:"amount"`            //金额
	Status          int32  `json:"status"`            //订单状态
	ErrorMessage    string `json:"error_message"`     //错误信息
	Timestamp       int64  `json:"timestamp"`         //时间
	Sign            string `json:"sign"`              //签名
}

// 核单响应
type PayInspectRespond struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    PayMessage `json:"data"`
}

type PayMessage struct {
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

type PayBalanceRespond struct {
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

func InitConfig(payurl, withdrawurl, merchant, appid, serverPubRSA, clientPriRSA, md5key, notifyUrl, withdrawNotifyUrl string) {
	Config = &PayConfig{
		Merchant:          merchant,
		AppId:             appid,
		ServerPubRSA:      serverPubRSA,
		ClientPriRSA:      clientPriRSA,
		MD5Key:            md5key,
		PayOrderUrl:       payurl,
		WithDrawUrl:       withdrawurl,
		PayNotifyUrl:      fmt.Sprintf("%s%s", notifyUrl, "/"+Config.GetChannelName()),
		WithDrawNotifyUrl: fmt.Sprintf("%s%s", withdrawNotifyUrl, "/"+Config.GetChannelName()),
		// InspectOrderUrl:    fmt.Sprintf("%s%s", url, "/pay/query"),
		// InspectWithdrawUrl: fmt.Sprintf("%s%s", url, "/transfer/query"),
		// QueryUrl:           fmt.Sprintf("%s%s", url, "/balance"),
	}
	service.Regist(Config.GetChannelId(), Config)
	service.RegistCallback(Config)
}
