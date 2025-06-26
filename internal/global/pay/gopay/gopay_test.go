package gopay

import (
	"encoding/json"
	"fmt"
	"goserver/internal/global/pay/entity"
	"testing"

	"gopkg.in/ini.v1"
)

func MjlGoPAYInit(cfg *ini.File) {
	account := cfg.Section("").Key("gopay.account").String()
	key := cfg.Section("").Key("gopay.key").String()
	payNotifyUrl := cfg.Section("").Key("gopay.payNotifyUrl").String()
	withdrawNotifyUrl := cfg.Section("").Key("gopay.withdrawNotifyUrl").String()
	apiUrl := cfg.Section("").Key("gopay.apiUrl").String()

	InitConfig(account, key, payNotifyUrl, withdrawNotifyUrl, apiUrl)
}

func TestSubmit(t *testing.T) {
	cfg, err := ini.Load("../../bin/app.conf")
	if err != nil {
		panic(err)
	}
	MjlGoPAYInit(cfg)

	payParams := Config.compose(&entity.PayOrder{
		OrderID:  "10001",
		Amount:   20000,
		RealName: "joker",
	})
	payBytes, _ := json.Marshal(payParams)
	fmt.Printf("payParams4: %v\n", string(payBytes))
}

func TestSubmitWithdraw(t *testing.T) {
	cfg, err := ini.Load("../../bin/app.conf")
	if err != nil {
		panic(err)
	}
	MjlGoPAYInit(cfg)

	payParams := Config.withdrawCompose(&entity.WithdrawOrder{
		OrderID:    "10002",
		Amount:     20000,
		Bank:       "nagad",
		RealName:   "LILAILAI",
		BankNumber: "01789855400",
	})
	payBytes, _ := json.Marshal(payParams)
	fmt.Printf("payParams: %v\n", string(payBytes))
}

func TestBalanceQuery(t *testing.T) {
	cfg, err := ini.Load("../../bin/app.conf")
	if err != nil {
		panic(err)
	}
	MjlGoPAYInit(cfg)

	balance := Config.BalanceQuery()
	fmt.Printf("balance: %d", balance)
}
