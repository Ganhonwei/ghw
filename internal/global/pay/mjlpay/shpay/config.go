package shpay

import (
	"fmt"
	"goserver/internal/global/pay/service"
	"strings"
)

var (
	WalletMap map[string]string
)

type ShPayConfig struct {
	Merchant           string // 商户号
	AppId              string // appid
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
	Success bool           `json:"success"`
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    map[string]any `json:"result"`
}

func InitConfig(merchant, appid, md5, notifyUrl, withdrawNotifyUrl, baseUrl string) {
	Config = &ShPayConfig{
		Merchant:           merchant,
		AppId:              appid,
		Md5Key:             md5,
		PayOrderUrl:        fmt.Sprintf("%s%s", baseUrl, "/v1/trans/payIn"),
		WithDrawUrl:        fmt.Sprintf("%s%s", baseUrl, "/v1/trans/payOut"),
		PayNotifyUrl:       fmt.Sprintf("%s%s", notifyUrl, "/shpay"),
		WithDrawNotifyUrl:  fmt.Sprintf("%s%s", withdrawNotifyUrl, "/shpay"),
		InspectOrderUrl:    fmt.Sprintf("%s%s", baseUrl, "/v1/trans/payQuery"),
		InspectWithdrawUrl: fmt.Sprintf("%s%s", baseUrl, "/v1/trans/payQuery"),
		QueryUrl:           fmt.Sprintf("%s%s", baseUrl, "/v1/trans/appAvailableAmt"),
	}
	service.Regist(service.MJL_SHPAY, Config)
	InitWallet()
}
func InitWallet() {
	WalletMap = map[string]string{
		"bkash": "bKash",
		"nagad": "Nagad",
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
