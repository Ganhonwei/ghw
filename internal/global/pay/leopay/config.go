package leopay

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
	OrderNo   string `json:"order_no"`
	OrderTime string `json:"order_time"`
	OrderData string `json:"order_data"`
	ErrMsg    string `json:"err_msg"`
	ErrCode   string `json:"err_code"`
	Status    string `json:"status"`
}

type PayOrderRespondData struct {
	MerchantId string `json:"pay_memberid"`
	OrderId    string `json:"pay_orderid"`
	Amount     string `json:"pay_amount"`
	PayUrl     string `json:"pay_url"`
	TradeId    string `json:"transaction_id"`
}

type PayWithdrawRespond struct {
	Status    string `json:"status"`
	// ErrCode   string `json:"err_code"`
	ErrMsg    string `json:"err_msg"`
	TerraceNo string `json:"terraceNo"`
}

type PayWithdrawRespondData struct {
	MerchantId  string `json:"memberId"`
	OrderId     string `json:"orderId"`
	Amount      string `json:"amount"`
	Model       string `json:"model"`
	TradeId     string `json:"tradeId"`
	OrderStatus string `json:"orderStatus"`
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
	Message string     `json:"msg"`
	Data    PayMessage `json:"data"`
}

type PayMessage struct {
	OrderId     string `json:"orderId"`
	MemberId    string `json:"memberId"`
	Amount      string `json:"amount"`
	TradeId     string `json:"tradeId"`
	PayTime     int64  `json:"payTime"`
	OrderStatus string `json:"orderStatus"`
}

type PayBalanceRespond struct {
	Code    int            `json:"code"`
	Message string         `json:"msg"`
	Data    BalanceMessage `json:"data"`
}

type BalanceMessage struct {
	Balance string `json:"balance"`
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
		// InspectOrderUrl:    "https://pay_api.leopay.lol/pay/query",
		// InspectWithdrawUrl: "https://pay_api.leopay.lol/payout/query",
		// QueryUrl:           "https://pay_api.leopay.lol/payout/querybalance",
	}
	service.Regist(Config.GetChannelId(), Config)
	service.RegistCallback(Config)
}
