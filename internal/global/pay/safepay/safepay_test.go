package safepay

import (
	"encoding/json"
	"fmt"
	"goserver/internal/global/pay/entity"
	"testing"

	"gopkg.in/ini.v1"
)

func SafePayInit(cfg *ini.File) {
	merno := cfg.Section("").Key("safepay.merno").String()
	key := cfg.Section("").Key("safepay.key").String()
	payReturnUrl := cfg.Section("").Key("safepay.payReturnUrl").String()
	withdrawReturnUrl := cfg.Section("").Key("safepay.withdrawReturnUrl").String()
	apiUrl := cfg.Section("").Key("safepay.apiUrl").String()

	InitConfig(merno, key, payReturnUrl, withdrawReturnUrl, apiUrl)
}

func TestSubmit(t *testing.T) {
	cfg, err := ini.Load("../../bin/app.conf")
	if err != nil {
		panic(err)
	}
	SafePayInit(cfg)

	payParams := Config.compose(&entity.PayOrder{
		OrderID:  "10003",
		Amount:   20000,
		RealName: "joker",
	})
	payBytes, _ := json.Marshal(payParams)
	fmt.Printf("payParams: %v\n", string(payBytes))
}

func TestSubmitWithdraw(t *testing.T) {
	cfg, err := ini.Load("../../bin/app.conf")
	if err != nil {
		panic(err)
	}
	SafePayInit(cfg)

	payParams := Config.withdrawCompose(&entity.WithdrawOrder{
		OrderID:    "10005",
		Amount:     20000,
		Bank:       "upay",
		RealName:   "LILAILAI",
		BankNumber: "557898554",
	})
	payBytes, _ := json.Marshal(payParams)
	fmt.Printf("payParams: %v\n", string(payBytes))
}

func TestBalanceQuery(t *testing.T) {
	cfg, err := ini.Load("../../bin/app.conf")
	if err != nil {
		panic(err)
	}
	SafePayInit(cfg)

	balance := Config.BalanceQuery()
	fmt.Printf("balance: %d", balance)
}
