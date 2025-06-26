package ai2pay

import (
	"fmt"
	"goserver/internal/global/pay/service"
)

var BankCode map[string]string = map[string]string{
	"JAZZCASH":                  "JAZZCASH",
	"EASYPAISA":                 "EASYPAISA",
	"HABIB BANK LIMITED":        "HBL",
	"UNITED BANK LIMITED":       "UBL",
	"NATIONAL BANK OF PAKISTAN": "NBOP",
	"MCB BANK LIMITED":          "MBL",
	"BANK ALFALAH LIMITED":      "AAL",
	"MEEZAN BANK":               "MEZ",
	"FAYSAL BANK LIMITED":       "FBL",
	"ASKARI BANK LIMITED":       "ASK",
}

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
	Mer_no       string `json:"mer_no"`
	Order_no     string `json:"order_no"`
	Order_amount string `json:"order_amount"`
	Status       string `json:"status"`
	Status_mes   string `json:"status_mes"`
	Order_data   string `json:"order_data"`
	Note_info    string `json:"note_info"`
	Sys_no       string `json:"sys_no"`
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
	Status    string `json:"status"`
	StatusMes string `json:"status_mes"`
	SysNo     string `json:"sys_no"`
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
	MerNo              string `json:"mer_no"`
	CheckStatus        string `json:"checkstatus"`
	ResultStatus       string `json:"resultstatus"`        // 查询结果，实际订单以这个状态为准
	OrderAmount        string `json:"order_amount"`        // 订单金额
	OrderRealityAmount string `json:"order_realityamount"` // 实际支付金额
	SysNo              string `json:"sys_no"`              // 三方单号
}

type PayBalanceRespond struct {
	MerNo       string         `json:"mer_no"`
	CheckStatus string         `json:"checkstatus"`
	StatusMes   string         `json:"status_mes"`
	Currency    BalanceMessage `json:"currency"`
}

type BalanceMessage struct {
	Balance string `json:"PKR"`
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
