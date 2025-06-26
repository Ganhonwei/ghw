package metagopay

import (
	"fmt"
	"goserver/internal/global/pay/entity"
	"testing"
	"time"

	"gopkg.in/ini.v1"
)

func metagopayInit() {
	cfg, err := ini.Load("../../bin/config/app.conf")
	if err != nil {
		panic(err)
	}

	orgId := cfg.Section("").Key("metagopay.orgId").String()
	mchId := cfg.Section("").Key("metagopay.mchId").String()
	accountId := cfg.Section("").Key("metagopay.accountId").String()
	md5Key := cfg.Section("").Key("metagopay.md5Key").String()
	payNotifyUrl := cfg.Section("").Key("metagopay.payNotifyUrl").String()
	withdrawNotifyUrl := cfg.Section("").Key("metagopay.withdrawNotifyUrl").String()
	InitConfig(orgId, mchId, accountId, md5Key, payNotifyUrl, withdrawNotifyUrl)
}

func TestSubmit(t *testing.T) {
	metagopayInit()

	order := &entity.PayOrder{
		OrderID:  "10005",
		Amount:   15000,
		RegistIp: "157.35.66.37",
		RealName: "joker",
		Email:    "2295758@gmail.com",
		Mobile:   "7112223333",
	}
	rsp, err := Config.Submit(order)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(rsp))
}

func TestInspectSubmit(t *testing.T) {
	metagopayInit()

	order := &entity.PayOrder{
		OrderID: "10005",
	}
	rsp, err := Config.InspectSubmit(order)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(rsp))
	err = Config.InspectResponse(rsp, order)
	if err != nil {
		t.Error(err)
	}
}

func TestWithdrawSubmit(t *testing.T) {
	metagopayInit()

	order := &entity.WithdrawOrder{
		OrderID:    "10010",
		Amount:     1000,
		RealName:   "test",
		Bank:       "test",
		IFSC:       "UTIB0000430",
		BankNumber: "919010054285006",
		Mobile:     "7112223333",
		Email:      "1163129@gmail.com",
	}
	rsp, err := Config.WithdrawSubmit(order)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(rsp))
}

func TestXxx(t *testing.T) {
	layout := "20060102150405" // 时间字符串的格式
	ti := "20241224105719"
	tt, err := time.Parse(layout, ti)
	fmt.Println(err, tt)
}
func TestBalanceQuery(t *testing.T) {
	metagopayInit()
	fmt.Println(Config.BalanceQuery())
}
