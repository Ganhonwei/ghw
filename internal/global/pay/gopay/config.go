package gopay

import (
	"fmt"
	"goserver/internal/global/pay/service"
	"strings"
)

var (
	WalletMap map[string]string
)

type GoPayConfig struct {
	Account            string
	Key                string
	PayNotifyUrl       string
	WithdrawNotifyUrl  string
	PayUrl             string
	WithdrawUrl        string
	InspectOrderUrl    string
	InspectWithdrawUrl string
	QueryUrl           string
}

type GoPayOrderRespond struct {
	Code int32          `json:"code"`
	Msg  string         `json:"msg"`
	Data GoPayOrderData `json:"data"`
}

type GoPayOrderData struct {
	TransactionId string `json:"transactionId"`
	Url           string `json:"url"`
}

type GoPayWithdrawOrderRespond struct {
	Code int32                  `json:"code"`
	Msg  string                 `json:"msg"`
	Data GoPayWithdrawOrderData `json:"data"`
}

type GoPayWithdrawOrderData struct {
	Account       string `json:"account"`
	OrderId       string `json:"orderId"`
	TransactionId string `json:"transactionId"`
	Status        int32  `json:"status"`
	ErrorMsg      string `json:"errorMsg"`
	Amount        string `json:"amount"`
	Fee           string `json:"fee"`
	Sign          string `json:"sign"`
}

type InspectRespond struct {
	Code int32       `json:"code"`
	Msg  string      `json:"msg"`
	Data InspectData `json:"data"`
}

type InspectData struct {
	OrderId         string `json:"orderId"`
	TransactionId   string `json:"transactionId"`
	Status          int32  `json:"status"`
	Amount          string `json:"amount"`
	ActualAmount    string `json:"actualAmount"`
	TransactionTime string `json:"transactionTime"`
	Fee             string `json:"fee"`
}

type InspectWithdrawRespond struct {
	Code int32               `json:"code"`
	Msg  string              `json:"msg"`
	Data InspectWithdrawData `json:"data"`
}

type InspectWithdrawData struct {
	OrderId       string `json:"orderId"`
	TransactionId string `json:"transactionId"`
	Status        int32  `json:"status"`
	ErrorMsg      string `json:"errorMsg"`
	Amount        string `json:"amount"`
	Fee           string `json:"fee"`
}

type BalanceRespond struct {
	Code int32       `json:"code"`
	Msg  string      `json:"msg"`
	Data BalanceData `json:"data"`
}

type BalanceData struct {
	Account string `json:"account"`
	Balance string `json:"balance"`
}

func InitConfig(account, key, payNotifyUrl, withdrawNotifyUrl, apiUrl string) {
	Config = &GoPayConfig{
		Account:            account,
		Key:                key,
		PayNotifyUrl:       payNotifyUrl,
		WithdrawNotifyUrl:  withdrawNotifyUrl,
		PayUrl:             fmt.Sprintf("%s%s", apiUrl, "/api/repay/link"),
		WithdrawUrl:        fmt.Sprintf("%s%s", apiUrl, "/api/withdraw"),
		InspectOrderUrl:    fmt.Sprintf("%s%s", apiUrl, "/api/repay/result"),
		InspectWithdrawUrl: fmt.Sprintf("%s%s", apiUrl, "/api/withdraw/result"),
		QueryUrl:           fmt.Sprintf("%s%s", apiUrl, "/api/merchant/balance"),
	}
	service.Regist(service.MJL_GOPAY, Config)
	InitWallet()
}

func InitWallet() {
	WalletMap = map[string]string{
		"nagad": "Nagad",
		"bkash": "bKash",
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
