package universalpay

import (
	"crypto"
	"fmt"
	"goserver/internal/global/pay/entity"
	"testing"

	"gopkg.in/ini.v1"
)

func universalpayInit() {
	cfg, err := ini.Load("../../bin/config/app.conf")
	if err != nil {
		panic(err)
	}
	appid := cfg.Section("").Key("universalpay.appid").String()
	prirsaClient := cfg.Section("").Key("universalpay.prirsaClient").String()
	pubrsaClient := cfg.Section("").Key("universalpay.pubrsaClient").String()
	pubrsaServer := cfg.Section("").Key("universalpay.pubrsaServer").String()
	payNotifyUrl := cfg.Section("").Key("universalpay.payNotifyUrl").String()
	withdrawNotifyUrl := cfg.Section("").Key("universalpay.withdrawNotifyUrl").String()
	InitConfig(appid, prirsaClient, pubrsaClient, pubrsaServer, payNotifyUrl, withdrawNotifyUrl)
}

func TestSubmit(t *testing.T) {
	universalpayInit()

	order := &entity.PayOrder{
		OrderID:  "10007",
		Amount:   20000,
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
	universalpayInit()

	order := &entity.PayOrder{
		OrderID: "10007",
	}
	resp, err := Config.InspectSubmit(order)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(resp))
	err = Config.InspectResponse(resp, order)
	if err != nil {
		t.Error(err)
	}
}

func TestWithdrawSubmit(t *testing.T) {
	universalpayInit()

	order := &entity.WithdrawOrder{
		OrderID:    "10011",
		Amount:     10000,
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

	ok, msg := Config.WithdrawSubmitResponse(rsp, order)
	fmt.Println(ok, msg)
}

func TestInspectWithdrawSubmit(t *testing.T) {
	universalpayInit()

	order := &entity.WithdrawOrder{
		OrderID: "10010",
	}
	resp, err := Config.InspectWithdrawSubmit(order)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(resp))
	status := Config.InspectWdResponse(resp, order)
	fmt.Println("status:", status)
}

func TestBalanceQuery(t *testing.T) {
	universalpayInit()

	fmt.Println(Config.BalanceQuery())
}

func TestSign(t *testing.T) {
	universalpayInit()

	body := "{\"amount\":\"200.00\",\"orderNo\":\"10007\",\"platformOrderNo\":\"PP20241224153424VBDQU5\",\"fee\":\"11.00\",\"linkUrl\":\"pay_order\",\"udf5\":null,\"udf3\":null,\"udf4\":null,\"udf1\":null,\"udf2\":null,\"status\":\"0\"}"
	sign := "ycPKNOI/CEnJK0dIfiaR4j6dQJjjabVYwrEOEKRUpi0gBEW/Qie8aFL0PlYYY0bkuRkiJA+b6l9RvFSidj7FBNPkxDFJ6WSmCQ2Nlc19qMDz5YIqZEA5irX7J/g5s411YHQ6rP3+KWdCXpHxqvDKVn5o2bPJfRgwZ4tNp7JV+1w="
	params, err := Config.ParsePayResult([]byte(body))
	if err != nil {
		t.Error(err)
	}
	c := SignContent(params)
	fmt.Println("sign content1: ", c)
	ok := Config.VerifySing(sign, c, crypto.SHA512)
	fmt.Println(ok)
}
