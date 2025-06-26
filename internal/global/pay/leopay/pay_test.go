package leopay

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

	payurl := cfg.Section("").Key("leopay.payurl").String()
	withdrawurl := cfg.Section("").Key("leopay.withdrawurl").String()
	merchant := cfg.Section("").Key("leopay.merchant").String()
	appid := cfg.Section("").Key("leopay.appid").String()
	serverPubRSA := cfg.Section("").Key("leopay.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("leopay.clientPriRSA").String()
	md5key := cfg.Section("").Key("leopay.md5key").String()
	payNotifyUrl := cfg.Section("").Key("paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	InitConfig(payurl, withdrawurl, merchant, appid, serverPubRSA, clientPriRSA, md5key, payNotifyUrl, withdrawNotifyUrl)

	id, _ := gonanoid.New()

	id1, _ := gonanoid.New(16)

	_ = id1

	// order := &entity.PayOrder{}
	// order.Amount = 10000
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
	order.Amount = 20000
	order.OrderID = id
	order.IFSC = "TEXT0000430"
	order.BankNumber = "919011111111"
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
