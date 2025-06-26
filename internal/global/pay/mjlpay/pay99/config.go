package pay99

import (
	"fmt"
	"goserver/internal/global/pay/service"
	"strings"
)

var (
	WalletMap map[string]string
)

type Pay99Config struct {
	Merchant           string // 商户号
	HamcRSA            string // hamc秘钥
	Key                string // 秘钥key
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

type _99OrderRespond struct {
	// Code    int       `json:"code"` // 状态码
	// Message string    `json:"msg"`  // 返回的请求信息
	// Data    JyPayData `json:"data"`
	Result string     `json:"result"`
	Data   _99PayData `json:"data"`
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

type _99PayData struct {
	// OrderNo     string  `json:"orderNo"` //系统生成的平台订单号
	// MerOrderNo  string  `json:"merOrderNo"`
	// MerNo       string  `json:"merNo"` // 商户号
	// SubCode     string  `json:"subCode"`
	// SubMsg      string  `json:"subMsg"`
	// Status      int32   `json:"status"`
	// OrderData   string  `json:"orderData" ` // 支付二维码
	// Common      string  `json:"common"`
	// OrderAmount string  `json:"orderAmount"`
	// PayAmount   float64 `json:"payAmount"`
	// BusiCode    string  `json:"busiCode"`
	// PayTime     string  `json:"payTime"`
	// Sign        string  `json:"sign"` // 签名
	Payment     PaymentData `json:"payment"`
	Payment_url string      `json:"payment_url"`
}

type PaymentData struct {
	From                string `json:"from"`
	From_name           string `json:"from_name"`
	From_bank_name      string `json:"from_bank_name"`
	To                  string `json:"to"`
	To_name             string `json:"to_name"`
	To_bank_name        string `json:"to_bank_name"`
	Payment_status      string `json:"payment_status"`
	Payment_fail_reason string `json:"payment_fail_reason"`
	Final_amount        string `json:"final_amount"`
	Currency            string `json:"currency"`
	Transaction_type    string `json:"transaction_type"`
	Payment_uid         string `json:"payment_uid"`
	Channel             string `json:"channel"`
}

type _99BalanceRespond struct {
	Result string         `json:"result"`
	Data   _99BalanceData `json:"data"`
	// Code    int            `json:"code"`
	// Message string         `json:"msg"`
	// Data    BalanceMessage `json:"data"`
}

type _99BalanceData struct {
	Id            int    `json:"id"`
	Uid           string `json:"uid"`
	Created_at    string `json:"created_at"`
	Updated_at    string `json:"updated_at"`
	Deleted_at    string `json:"deleted_at"`
	Update_user   string `json:"update_user"`
	Update_reason string `json:"update_reason"`
	Merchant_name string `json:"merchant_name"`
	Merchant_uid  string `json:"merchant_uid"`
	Cny           string `json:"cny"`
	Usd           string `json:"usd"`
	Thb           string `json:"thb"`
	Vnd           string `json:"vnd"`
	Php           string `json:"php"`
	Sgd           string `json:"sgd"`
	Idr           string `json:"idr"`
	Bdt           string `json:"bdt"`
}

type BalanceMessage struct {
	List []AccountMessage `json:"list"`
	Sign string           `json:"sign"`
}

type AccountMessage struct {
	AccountNo     string `json:"accountNo"`
	Currency      string `json:"currency"`
	Balance       string `json:"balance"`
	FrozenBalance string `json:"frozenBalance"`
}

func InitConfig(appid, hamcRSA, key, serverPubRSA, clientPriRSA, notifyUrl, withdrawNotifyUrl, payUrl, payoutUrl string) {
	Config = &Pay99Config{
		Merchant:           appid,
		HamcRSA:            hamcRSA,
		Key:                key,
		RSAPubRSA:          serverPubRSA,
		RSAPriRSA:          clientPriRSA,
		PayOrderUrl:        fmt.Sprintf("%s%s", payUrl, "/octopus/api/order/new"),
		WithDrawUrl:        fmt.Sprintf("%s%s", payoutUrl, "/octopus/api/order/new"),
		PayNotifyUrl:       fmt.Sprintf("%s%s", notifyUrl, "/99pay"),
		WithDrawNotifyUrl:  fmt.Sprintf("%s%s", withdrawNotifyUrl, "/99pay"),
		InspectOrderUrl:    fmt.Sprintf("%s%s", payUrl, "/payin/orderQuery"),
		InspectWithdrawUrl: fmt.Sprintf("%s%s", payoutUrl, "/payout/singleQuery"),
		QueryUrl:           fmt.Sprintf("%s%s", payoutUrl, "/octopus/api/ledger/get_merchant_ledger"),
	}
	service.Regist(service.MJL_99, Config)
	InitWallet()
}

func InitWallet() {
	WalletMap = map[string]string{
		"bkash": "bkash",
		// "upay":  "upay",
		"nagad": "nagad",
	}
}

func checkWallet(wallet string) bool {
	noSpaces := strings.ReplaceAll(wallet, " ", "")
	str := strings.ToLower(noSpaces)
	return WalletMap[str] != ""
}
