package transafepay

import (
	"fmt"
	"goserver/internal/global/pay/service"
	"strings"
)

var (
	WalletMap map[string]string
)

type TransafePayConfig struct {
	Merchant           string // 商户号
	AppId              string // 商户appid
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

type TransafeOrderRespond struct {
	Status int             `json:"status"` // 状态码
	Msg    string          `json:"msg"`    // 返回的请求信息
	Data   TransafePayData `json:"data"`
}

// 代收回调
type TransafeRechargeCallback struct {
	Status          int    `json:"status"`      //状态 200成功
	MemberId        string `json:"memberId"`    //商户id
	OrderId         string `json:"platOrderNo"` //平台订单号
	MerchantOrderNo string `json:"mchOrderNo"`  //商户订单号
	Amount          int    `json:"amount"`      //实际支付金额
	OrderStatus     string `json:"orderStatus"` //订单状态
	Fee             int    `json:"fee"`         //手续费
	Msg             string `json:"msg"`         //信息
	TimeEnd         string `json:"timeEnd"`     //完结时间 秒级时间戳
	Sign            string `json:"sign"`        //签名
}

type TransafePayData struct {
	MchOrderNo  string `json:"mchOrderNo"`  // 商户订单号
	PlatOrderNo string `json:"platOrderNo"` // 平台订单号
	ApplyTime   int    `json:"applyTime"`   // 申请时间
	Amount      int    `json:"amount"`      // 实际支付金额
	Reference   string `json:"reference"`
	Link        string `json:"link"`       // 支付链接
	ExpireTime  int    `json:"expireTime"` // 过期时间
	OrderStatus string `json:"orderStatus"`
	TrxId       string `json:"trxId"`   // UTR
	Fee         int    `json:"fee"`     // 手续费用
	TimeEnd     int    `json:"timeEnd"` // 完结时间
}

type TransafeBalanceRespond struct {
	Status int            `json:"status"`
	Msg    string         `json:"msg"`
	Data   BalanceMessage `json:"data"`
}

type BalanceMessage struct {
	Balance int `json:"balance"`
	Frozen  int `json:"frozen"`
}

func InitConfig(merno, appid, serverPubRSA, clientPriRSA, notifyUrl, withdrawNotifyUrl, payUrl string) {
	Config = &TransafePayConfig{
		Merchant:           merno,
		AppId:              appid,
		RSAPubRSA:          serverPubRSA,
		RSAPriRSA:          clientPriRSA,
		PayOrderUrl:        fmt.Sprintf("%s%s", payUrl, "/payment/bengal/order/create"),
		WithDrawUrl:        fmt.Sprintf("%s%s", payUrl, "/payout/bengal/order/create"),
		PayNotifyUrl:       fmt.Sprintf("%s%s", notifyUrl, "/transafepay"),
		WithDrawNotifyUrl:  fmt.Sprintf("%s%s", withdrawNotifyUrl, "/transafepay"),
		InspectOrderUrl:    fmt.Sprintf("%s%s", payUrl, "/payout/bengal/order/query"),
		InspectWithdrawUrl: fmt.Sprintf("%s%s", payUrl, "/payout/bengal/order/query"),
		QueryUrl:           fmt.Sprintf("%s%s", payUrl, "/payout/balance"),
	}
	service.Regist(service.MJL_Transafe, Config)
	InitWallet()
}

func InitWallet() {
	WalletMap = map[string]string{
		"UPAY":  "UPAY",
		"NAGAD": "NAGAD",
		"BKASH": "BKASH",
	}
}

func checkWallet(wallet string) bool {
	noSpaces := strings.ReplaceAll(wallet, " ", "")
	str := strings.ToUpper(noSpaces)
	return WalletMap[str] != ""
}

func getWallet(wallet string) string {
	noSpaces := strings.ReplaceAll(wallet, " ", "")
	str := strings.ToUpper(noSpaces)
	return WalletMap[str]
}
