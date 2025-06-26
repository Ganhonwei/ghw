package jypay

import (
	"fmt"
	"goserver/internal/global/pay/service"
	"strings"
)

var (
	WalletMap map[string]string
)

type JYPayConfig struct {
	Merchant           string // 商户号
	HamcRSA            string // hamc秘钥
	RSAPubRSA          string // RSA公钥
	RSAPriRSA          string // RSA私钥
	PayOrderUrl        string // 代收地址
	WithDrawUrl        string // 代付地址
	PayNotifyUrl       string // 支付通知地址
	WithDrawNotifyUrl  string // 提现通知地址
	InspectOrderUrl    string // 查单接口
	InspectWithdrawUrl string // 代付查询
	QueryUrl           string // 余额查询
}

type JyOrderRespond struct {
	Code    int       `json:"code"` // 状态码
	Message string    `json:"msg"`  // 返回的请求信息
	Data    JyPayData `json:"data"`
}

// 代收回调
type JyRechargeCallback struct {
	MerchantCode    string `json:"merchant_code"`     //商户号
	OrderId         string `json:"order_no"`          //平台订单号
	MerchantOrderNo string `json:"merchant_order_no"` //商户订单号
	Amount          string `json:"amount"`            //金额
	Status          int32  `json:"status"`            //订单状态
	ErrorMessage    string `json:"error_message"`     //错误信息
	Timestamp       int64  `json:"timestamp"`         //时间
	Sign            string `json:"sign"`              //签名
}

type JyPayData struct {
	OrderNo     string  `json:"orderNo"` //系统生成的平台订单号
	MerOrderNo  string  `json:"merOrderNo"`
	MerNo       string  `json:"merNo"` // 商户号
	SubCode     string  `json:"subCode"`
	SubMsg      string  `json:"subMsg"`
	Status      int32   `json:"status"`
	OrderData   string  `json:"orderData" ` // 支付二维码
	Common      string  `json:"common"`
	OrderAmount string  `json:"orderAmount"`
	PayAmount   float64 `json:"payAmount"`
	BusiCode    string  `json:"busiCode"`
	PayTime     string  `json:"payTime"`
	Sign        string  `json:"sign"` // 签名
}

type JyBalanceRespond struct {
	Code    int            `json:"code"`
	Message string         `json:"msg"`
	Data    BalanceMessage `json:"data"`
}

type BalanceMessage struct {
	List []AccountMessage `json:"list"`
	Sign string           `json:"sign"`
}

type AccountMessage struct {
	AccountNo     string  `json:"accountNo"`
	Currency      string  `json:"currency"`
	Balance       float64 `json:"balance"`
	FrozenBalance float64 `json:"frozenBalance"`
}

func InitConfig(appid, hamcRSA, serverPubRSA, clientPriRSA, notifyUrl, withdrawNotifyUrl, payUrl, payoutUrl string) {
	Config = &JYPayConfig{
		Merchant:           appid,
		HamcRSA:            hamcRSA,
		RSAPubRSA:          serverPubRSA,
		RSAPriRSA:          clientPriRSA,
		PayOrderUrl:        fmt.Sprintf("%s%s", payUrl, "/payin/createOrder"),
		WithDrawUrl:        fmt.Sprintf("%s%s", payoutUrl, "/payout/singleOrder"),
		PayNotifyUrl:       fmt.Sprintf("%s%s", notifyUrl, "/jypay"),
		WithDrawNotifyUrl:  fmt.Sprintf("%s%s", withdrawNotifyUrl, "/jypay"),
		InspectOrderUrl:    fmt.Sprintf("%s%s", payUrl, "/payin/orderQuery"),
		InspectWithdrawUrl: fmt.Sprintf("%s%s", payoutUrl, "/payout/singleQuery"),
		QueryUrl:           fmt.Sprintf("%s%s", payoutUrl, "/payout/balanceQuery"),
	}
	service.Regist(service.MJL_JY, Config)
	InitWallet()
}

func InitWallet() {
	WalletMap = map[string]string{
		"bkash": "baksh",
		// "upay":  "upay",
		"nagad": "ngand",
	}
}

func checkWallet(wallet string) bool {
	noSpaces := strings.ReplaceAll(wallet, " ", "")
	str := strings.ToLower(noSpaces)
	return WalletMap[str] != ""
}

func getWallet(wallet string) string {
	noSpaces := strings.ReplaceAll(wallet, " ", "")
	str := strings.ToLower(noSpaces)
	return WalletMap[str]
}
