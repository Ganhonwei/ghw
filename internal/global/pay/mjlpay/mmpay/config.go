package mmpay

import (
	"fmt"
	"goserver/internal/global/pay/service"
	"strings"
)

var (
	WalletMap map[string]string
)

type MMPayConfig struct {
	Merchant           string // 商户号
	Md5Key             string // MD5
	PayOrderUrl        string // 代收地址
	WithDrawUrl        string // 代付地址
	PayNotifyUrl       string // 支付通知地址
	WithDrawNotifyUrl  string // 提现通知地址
	InspectOrderUrl    string // 查单接口
	InspectWithdrawUrl string // 代付查询
	QueryUrl           string // 余额查询
}

type BaseRespone struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    map[string]any
}

type MMBalanceRespond struct {
	Data BalanceMessage `json:"data"`
	// MerNo string           `json:"mer_no"`
	// Data  []BalanceMessage `json:"funds"`
}

type BalanceMessage struct {
	MerNo string         `json:"mer_no"`
	Data  []BalanceFunds `json:"funds"`
	// Currency       string `json:"currency"`
	// Balance        string `json:"balance"`
	// AvailableFunds string `json:"available_funds"`
}

type BalanceFunds struct {
	Currency       string `json:"currency"`
	Balance        string `json:"balance"`
	AvailableFunds string `json:"available_funds"`
}

func InitConfig(appid, md5, notifyUrl, withdrawNotifyUrl, baseUrl string) {
	Config = &MMPayConfig{
		Merchant:           appid,
		Md5Key:             md5,
		PayOrderUrl:        fmt.Sprintf("%s%s", baseUrl, "/open/api/order/in"),
		WithDrawUrl:        fmt.Sprintf("%s%s", baseUrl, "/open/api/order/out"),
		PayNotifyUrl:       fmt.Sprintf("%s%s", notifyUrl, "/mmpay"),
		WithDrawNotifyUrl:  fmt.Sprintf("%s%s", withdrawNotifyUrl, "/mmpay"),
		InspectOrderUrl:    fmt.Sprintf("%s%s", baseUrl, "/open/api/order/query/in"),
		InspectWithdrawUrl: fmt.Sprintf("%s%s", baseUrl, "/open/api/order/query/out"),
		QueryUrl:           fmt.Sprintf("%s%s", baseUrl, "/open/api/merchant/balance"),
	}
	service.Regist(service.MJL_MM, Config)
	InitWallet()
}

func InitWallet() {
	WalletMap = map[string]string{
		"bkash": "003",
		"nagad": "002",
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
