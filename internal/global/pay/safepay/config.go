package safepay

import (
	"goserver/internal/global/pay/service"
	"strings"
)

var (
	WalletMap map[string]string
)

type SafePayConfig struct {
	Merno             string // 商户号
	Key               string // 秘钥
	PayReturnUrl      string
	WithdrawReturnUrl string
	ApiUrl            string
	// PageUrl string // 待收跳转地址(非必填)
}

// 下单响应
type SafePayOrderRespond struct {
	MerNo       string `json:"mer_no"`
	OrderNo     string `json:"order_no"`
	OrderAmount string `json:"order_amount"`
	Status      string `json:"status"`
	StatusMes   string `json:"status_mes"`
	OrderData   string `json:"order_data"`
}

// 代付响应
type SafePayWithdrawOrderRespond struct {
	Status    string `json:"status"`
	StatusMes string `json:"status_mes"`
	SysNo     string `json:"sys_no"`
}

type InspectRespond struct {
	MerNo              string `json:"mer_no"`
	OrderNo            string `json:"order_no"`
	OrderAmount        string `json:"order_amount"`
	CheckStatus        string `json:"checkstatus"`
	ResultStatus       string `json:"resultstatus"`
	OrderRealityAmount string `json:"order_realityamount"`
	StatusMes          string `json:"status_mes"`
}

// 代收待付订单查询
type InspectWithdrawRespond struct {
	MerNo              string `json:"mer_no"`
	OrderNo            string `json:"order_no"`
	OrderAmount        string `json:"order_amount"`
	CheckStatus        string `json:"checkstatus"`
	ResultStatus       string `json:"resultstatus"`
	OrderRealityAmount string `json:"order_realityamount"`
	StatusMes          string `json:"status_mes"`
	Currency           string `json:"currency"`
	SysNo              string `json:"sys_no"`
}

// 查询账户余额
type SafePayBalanceRespond struct {
	MerNo       string            `json:"mer_no"`
	OrderNo     string            `json:"order_no"`
	OrderAmount string            `json:"order_amount"`
	CheckStatus string            `json:"checkstatus"`
	Currency    map[string]string `json:"currency"`
}

type SafePayRechargeCallback struct {
	MerNo              string `json:"mer_no"`
	OrderNo            string `json:"order_no"`
	PayTypeCode        string `json:"paytypecode"`
	OrderAmount        string `json:"order_amount"`
	OrderRealityAmount string `json:"order_realityamount"`
	Status             string `json:"status"`
	Sign               string `json:"sign"`
	SysNo              string `json:"sys_no"`
}

func InitConfig(merno, key, payReturnUrl, withdrawReturnUrl, apiUrl string) {
	Config = &SafePayConfig{
		Merno:             merno,
		Key:               key,
		PayReturnUrl:      payReturnUrl,
		WithdrawReturnUrl: withdrawReturnUrl,
		ApiUrl:            apiUrl,
	}
	service.Regist(service.MJL_SAFEPAY, Config)
	InitWallet()
}

func InitWallet() {
	WalletMap = map[string]string{
		"upay":   "25000f001",
		"nagad":  "25000f002",
		"bkash":  "25000f003",
		"rockey": "25000f004",
	}
}

func checkWallet(wallet string) bool {
	// noSpaces := strings.ReplaceAll(wallet, " ", "")
	// str := strings.ToLower(noSpaces)
	// return WalletMap[str] != ""
	return getWallet(wallet) != ""
}

func getWallet(wallet string) string {
	noSpaces := strings.ReplaceAll(wallet, " ", "")
	str := strings.ToLower(noSpaces)
	return WalletMap[str]
}
