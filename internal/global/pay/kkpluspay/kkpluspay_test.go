package kkpluspay

import (
	"goserver/internal/global/pay/entity"
	"testing"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"gopkg.in/ini.v1"
)

func TestKKPlusPay(t *testing.T) {
	cfg, err := ini.Load("../../../../configs/app.conf")
	if err != nil {
		panic(err)
	}

	payurl := cfg.Section("").Key("kkpluspay.payurl").String()
	withdrawurl := cfg.Section("").Key("kkpluspay.withdrawurl").String()
	appid := cfg.Section("").Key("kkpluspay.merchant").String()
	serverPubRSA := cfg.Section("").Key("kkpluspay.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("kkpluspay.clientPriRSA").String()
	md5key := cfg.Section("").Key("kkpluspay.md5key").String()
	notifyUrl := cfg.Section("").Key("kkpluspay.paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("kkpluspay.withDrawnotifyurl").String()

	InitConfig(payurl, withdrawurl, appid, serverPubRSA, clientPriRSA, md5key, notifyUrl, withdrawNotifyUrl)

	id, _ := gonanoid.New()

	// order := &entity.PayOrder{}
	// order.Amount = 100000
	// order.OrderID = id
	// order.RealName = "ZhangSan"
	// order.Email = "test@mail.com"
	// order.Mobile = "9852146882"
	// // order.RegistIp = "127.0.0.1"
	// // order.Userid = "1234567890"

	// resp, _ := Config.Submit(order)

	// link, _ := Config.SubmitResponse(resp, order)

	// _ = link

	order := &entity.WithdrawOrder{}
	order.Amount = 10000
	order.OrderID = id
	order.IFSC = "UTIB0000430"
	order.BankNumber = "919010054285006"
	order.RealName = "test"
	order.Mobile = "9852146882"
	order.Email = "test@mail.com"
	resp, _ := Config.WithdrawSubmit(order)

	ok, _ := Config.WithdrawSubmitResponse(resp, order)
	_ = ok

	// b := Config.BalanceQuery()
	// _ = b
}
