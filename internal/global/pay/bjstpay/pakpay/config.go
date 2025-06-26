package pakpay

import (
	"fmt"
	"goserver/internal/global/pay/service"
)

var ChannelCode map[string]string = map[string]string{
	"JAZZCASH":  "PKR_Jazzcash",
	"EASYPAISA": "PKR_Easypaisa",
}

var BankCode map[string]string = map[string]string{
	"JAZZCASH":                            "JAZZCASH",
	"EASYPAISA":                           "EASYPAISA",
	"ALBARAKA ISLAMIC BANK":               "ALBARAKA_ISLAMIC_BANK",
	"ALLIED BANK LIMITED":                 "ALLIED_BANK_LIMITED",
	"APNA MICRO FINANCE BANK":             "APNA_MICRO_FINANCE_BANK",
	"ASKARI BANK LIMITED":                 "ASKARI_BANK_LIMITED",
	"BANK AL HABIB LIMITED":               "BANK_AL_HABIB_LIMITED",
	"BANK ALFALAH LIMITED":                "BANK_ALFALAH_LIMITED",
	"BANK ISLAMI PAKISTAN LIMITED":        "BANK_ISLAMI_PAKISTAN_LIMITED",
	"BANK OF KHYBER":                      "BANK_OF_KHYBER",
	"CITI BANK NA":                        "CITI_BANK_NA",
	"DUBAI ISLAMIC BANK PAKISTAN LIMITED": "DUBAI_ISLAMIC_BANK_PAKISTAN_LIMITED",
	"FAYSAL BANK LIMITED":                 "FAYSAL_BANK_LIMITED",
	"FINCA MICRO FINANCE BANK":            "FINCA_MICRO_FINANCE_BANK",
	"FIRST WOMEN BANK LIMITED":            "FIRST_WOMEN_BANK_LIMITED",
	"HABIB BANK LIMITED":                  "HABIB_BANK_LIMITED",
	"HABIB METROPOLITAN BANK LIMITED":     "HABIB_METROPOLITAN_BANK_LIMITED",
	"JS BANK LIMITED":                     "JS_BANK_LIMITED",
	"MCB BANK LIMITED":                    "MCB_BANK_LIMITED",
	"MCB ISLAMIC":                         "MCB_ISLAMIC",
	"MEEZAN BANK":                         "MEEZAN_BANK",
	"NATIONAL BANK OF PAKISTAN":           "NATIONAL_BANK_OF_PAKISTAN",
	"NRSP MICRO FINANCE BANK":             "NRSP_MICRO_FINANCE_BANK",
	"NIB BANK LIMITED":                    "NIB_BANK_LIMITED",
	"SAMBA BANK LIMITED":                  "SAMBA_BANK_LIMITED",
	"SILK BANK LIMITED":                   "SILK_BANK_LIMITED",
	"SINDH BANK LIMITED":                  "SINDH_BANK_LIMITED",
	"SONERI BANK LIMITED":                 "SONERI_BANK_LIMITED",
	"STANDARD CHARTERED BANK LTD":         "STANDARD_CHARTERED_BANK_LTD",
	"SUMMIT BANK LIMITED":                 "SUMMIT_BANK_LIMITED",
	"TELENOR MICRO FINANCE BANK":          "TELENOR_MICRO_FINANCE_BANK",
	"THE BANK OF PUNJAB":                  "THE_BANK_OF_PUNJAB",
	"U MICRO FINANCE BANK":                "U_MICRO_FINANCE_BANK",
	"UNITED BANK LIMITED":                 "UNITED_BANK_LIMITED",
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
	Code int                 `json:"code"`
	Msg  string              `json:"msg"`
	Data PayOrderRespondData `json:"data"`
}

type PayOrderRespondData struct {
	MerchantId string `json:"pay_memberid"`
	OrderId    string `json:"pay_orderid"`
	Amount     string `json:"pay_amount"`
	PayUrl     string `json:"pay_url"`
	TradeId    string `json:"transaction_id"`
}

type PayWithdrawRespond struct {
	Code int                    `json:"code"`
	Msg  string                 `json:"msg"`
	Data PayWithdrawRespondData `json:"data"`
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
		Merchant:           merchant,
		AppId:              appid,
		ServerPubRSA:       serverPubRSA,
		ClientPriRSA:       clientPriRSA,
		MD5Key:             md5key,
		PayOrderUrl:        payurl,
		WithDrawUrl:        withdrawurl,
		PayNotifyUrl:       fmt.Sprintf("%s%s", notifyUrl, "/"+Config.GetChannelName()),
		WithDrawNotifyUrl:  fmt.Sprintf("%s%s", withdrawNotifyUrl, "/"+Config.GetChannelName()),
		InspectOrderUrl:    "https://pay_api.pakpay.lol/pay/query",
		InspectWithdrawUrl: "https://pay_api.pakpay.lol/payout/query",
		QueryUrl:           "https://pay_api.pakpay.lol/payout/querybalance",
	}
	service.Regist(Config.GetChannelId(), Config)
	service.RegistCallback(Config)
}
