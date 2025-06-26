package unipay

import (
	"fmt"
	"goserver/internal/global/pay/service"
	"strings"
)

var (
	WalletMap map[string]int
)

type UniPayConfig struct {
	Appid              string // 商户号
	Name               string // 商户名称
	PayOrderUrl        string // 代收地址
	WithDrawUrl        string // 代付地址
	PayNotifyUrl       string // 支付通知地址
	WithDrawNotifyUrl  string // 提现通知地址
	InspectOrderUrl    string // 查单接口
	InspectWithdrawUrl string // 代付查询
	QueryUrl           string // 余额查询地址
}

// 核单响应
type UniPayInspectRespond struct {
	Status      int    `json:"status"`
	OrderStatus int    `json:"order_status"`
	OrderId     string `json:"order_id"`
	Message     string `json:"message"`
	Merchant    string `json:"merchant"`
	Amount      int64  `json:"amount"`
	TotalAmount int64  `json:"total_amount"`
}

func InitConfig(appid, name, url, notifyUrl, withdrawNotifyUrl string) {
	Config = &UniPayConfig{
		Appid:              appid,
		Name:               name,
		PayOrderUrl:        fmt.Sprintf("%s%s", url, "/api/client/mjl/ds"),
		WithDrawUrl:        fmt.Sprintf("%s%s", url, "/api/client/mjl/df"),
		PayNotifyUrl:       fmt.Sprintf("%s%s", notifyUrl, "/universe"),
		WithDrawNotifyUrl:  fmt.Sprintf("%s%s", withdrawNotifyUrl, "/universe"),
		InspectOrderUrl:    fmt.Sprintf("%s%s", url, "/api/order/status"),
		InspectWithdrawUrl: fmt.Sprintf("%s%s", url, "/api/order/status"),
		QueryUrl:           fmt.Sprintf("%s%s", url, "/api/client/balance"),
	}
	service.Regist(service.MJL_UNI, Config)
	InitWallet()
}

func InitWallet() {
	WalletMap = map[string]int{
		"bkash":          1,
		"rocket":         2,
		"nagad":          3,
		"brac":           4,
		"surecash":       5,
		"tap":            6,
		"upay":           7,
		"okwallet":       8,
		"easypaisa":      9,
		"jazzcash":       10,
		"upaisa":         11,
		"nayapay":        12,
		"paymax":         13,
		"dbbl":           14,
		"ibbl":           15,
		"cbl":            16,
		"bankasia":       17,
		"easternbank":    18,
		"ificbank":       19,
		"primebank":      20,
		"abbank":         21,
		"jazzcashtillid": 22,
		"sebl":           23,
	}
}

func checkWallet(wallet string) bool {
	noSpaces := strings.ReplaceAll(wallet, " ", "")
	str := strings.ToLower(noSpaces)
	return WalletMap[str] > 0
}

func getWalletId(wallet string) int {
	noSpaces := strings.ReplaceAll(wallet, " ", "")
	str := strings.ToLower(noSpaces)
	return WalletMap[str]
}
