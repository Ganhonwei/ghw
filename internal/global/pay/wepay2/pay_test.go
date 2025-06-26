package wepay2

import (
	"goserver/internal/global/pay/entity"
	"testing"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"gopkg.in/ini.v1"
)

func Test(t *testing.T) {
	cfg, err := ini.Load("../../../../configs/app.conf")
	if err != nil {
		panic(err)
	}

	payurl := cfg.Section("").Key("wepay2.payurl").String()
	withdrawurl := cfg.Section("").Key("wepay2.withdrawurl").String()
	merchant := cfg.Section("").Key("wepay2.merchant").String()
	appid := cfg.Section("").Key("wepay2.appid").String()
	serverPubRSA := cfg.Section("").Key("wepay2.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("wepay2.clientPriRSA").String()
	pay_md5key := cfg.Section("").Key("wepay2.pay.md5key").String()
	withdraw_md5key := cfg.Section("").Key("wepay2.withdraw.md5key").String()
	payNotifyUrl := cfg.Section("").Key("paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()
	pay_type := cfg.Section("").Key("wepay2.pay_type").String()

	InitConfig(payurl, withdrawurl, merchant, appid, serverPubRSA, clientPriRSA, pay_md5key, withdraw_md5key, payNotifyUrl, withdrawNotifyUrl, pay_type)

	id, _ := gonanoid.New()

	id1, _ := gonanoid.New(16)

	_ = id1

	// order := &entity.PayOrder{}
	// order.Amount = 20000
	// order.OrderID = id
	// order.RealName = "ZhangSan"
	// order.Email = "test@mail.com"
	// order.Mobile = "9852146882"
	// // order.RegistIp = "127.0.0.1"
	// order.Userid = "1234567890"

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
	order.Userid = "1234567890"
	resp, _ := Config.WithdrawSubmit(order)

	ok, _ := Config.WithdrawSubmitResponse(resp, order)
	_ = ok

	// b := Config.BalanceQuery()
	// _ = b
}
