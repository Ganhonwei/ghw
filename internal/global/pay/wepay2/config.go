package wepay2

import (
	"fmt"
	"goserver/internal/global/pay/service"
)

type PayConfig struct {
	Merchant           string // 商户号
	AppId              string // 应用ID
	ServerPubRSA       string // 服务端公钥
	ClientPriRSA       string // 客户端私钥
	PayMD5Key          string // 代收md5key
	WithdrawMD5Key     string // 代付md5key
	PayOrderUrl        string // 代收地址
	WithDrawUrl        string // 代付地址
	PayNotifyUrl       string // 支付通知地址
	WithDrawNotifyUrl  string // 提现通知地址
	InspectOrderUrl    string // 查单接口
	InspectWithdrawUrl string // 代付查询
	QueryUrl           string // 余额查询
	PayType            string // 通道编码
}

type PayOrderRespond struct {
	RespCode    string `json:"respCode"`
	TradeMsg    string `json:"tradeMsg"`
	SignType    string `json:"signType"`
	Sign        string `json:"sign"`
	MchId       string `json:"mchId"`
	MchOrderNo  string `json:"mchOrderNo"`
	OriAmount   string `json:"oriAmount"`
	TradeAmount string `json:"tradeAmount"`
	OrderDate   string `json:"orderDate"`
	OrderNo     string `json:"orderNo"`
	TradeResult string `json:"tradeResult"`
	PayInfo     string `json:"payInfo"`
}

type PayOrderRespondData struct {
	MerchantId      string `json:"merchantId"`
	MerchantOrderId string `json:"merchantOrderId"`
	OrderId         string `json:"orderId"`
	Amount          string `json:"amount"`
	State           string `json:"state"`
	PayUrl          string `json:"payUrl"`
	ChannelId       string `json:"channelId"`
	Attach          string `json:"attach"`
	Sign            string `json:"sign"`
}

type PayWithdrawRespond struct {
	RespCode       string `json:"respCode"`
	ErrorMsg       string `json:"errorMsg"`
	SignType       string `json:"signType"`
	Sign           string `json:"sign"`
	MchId          string `json:"mchId"`
	MerTransferId  string `json:"merTransferId"`
	TransferAmount string `json:"transferAmount"`
	ApplyDate      string `json:"applyDate"`
	TradeNo        string `json:"tradeNo"`
	TradeResult    string `json:"tradeResult"`
}

type PayWithdrawRespondData struct {
	MerchantId      string `json:"merchantId"`
	MerchantOrderId string `json:"merchantOrderId"`
	OrderId         string `json:"orderId"`
	Amount          string `json:"amount"`
	State           string `json:"state"`
	ChannelId       string `json:"channelId"`
	FailReason      string `json:"failReason"`
	Attach          string `json:"attach"`
	MerchantDeduct  string `json:"merchantDeduct"`
	Sign            string `json:"sign"`
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

func InitConfig(payurl, withdrawurl, merchant, appid, serverPubRSA, clientPriRSA, pay_md5key, withdraw_md5key, notifyUrl, withdrawNotifyUrl, pay_type string) {
	Config = &PayConfig{
		Merchant:          merchant,
		AppId:             appid,
		ServerPubRSA:      serverPubRSA,
		ClientPriRSA:      clientPriRSA,
		PayMD5Key:         pay_md5key,
		WithdrawMD5Key:    withdraw_md5key,
		PayOrderUrl:       payurl,
		WithDrawUrl:       withdrawurl,
		PayNotifyUrl:      fmt.Sprintf("%s%s", notifyUrl, "/"+Config.GetChannelName()),
		WithDrawNotifyUrl: fmt.Sprintf("%s%s", withdrawNotifyUrl, "/"+Config.GetChannelName()),
		PayType:           pay_type,
		// InspectOrderUrl:    fmt.Sprintf("%s%s", url, "/pay/query"),
		// InspectWithdrawUrl: fmt.Sprintf("%s%s", url, "/transfer/query"),
		// QueryUrl:           fmt.Sprintf("%s%s", url, "/balance"),
	}
	service.Regist(Config.GetChannelId(), Config)
	service.RegistCallback(Config)
}
