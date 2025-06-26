package aipay

import (
	"fmt"
	"goserver/internal/global/pay/common"
	"testing"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"gopkg.in/ini.v1"
)

func Test(t *testing.T) {
	cfg, err := ini.Load("../../../../configs/app.conf")
	if err != nil {
		panic(err)
	}

	payurl := cfg.Section("").Key("aipay.payurl").String()
	withdrawurl := cfg.Section("").Key("aipay.withdrawurl").String()
	merchant := cfg.Section("").Key("aipay.merchant").String()
	appid := cfg.Section("").Key("aipay.appid").String()
	serverPubRSA := cfg.Section("").Key("aipay.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("aipay.clientPriRSA").String()
	md5key := cfg.Section("").Key("aipay.md5key").String()
	payNotifyUrl := cfg.Section("").Key("paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	InitConfig(payurl, withdrawurl, merchant, appid, serverPubRSA, clientPriRSA, md5key, payNotifyUrl, withdrawNotifyUrl)

	id, _ := gonanoid.New()

	id1, _ := gonanoid.New(16)

	_ = id
	_ = id1

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

	// order := &entity.WithdrawOrder{}
	// order.Amount = 10000
	// order.OrderID = id
	// order.IFSC = "UTIB0000430"
	// order.BankNumber = "919010054285006"
	// order.RealName = "test"
	// order.Mobile = "9852146882"
	// order.Email = "test@mail.com"
	// resp, _ := Config.WithdrawSubmit(order)

	// ok, _ := Config.WithdrawSubmitResponse(resp, order)
	// _ = ok

	b := Config.BalanceQuery()
	_ = b

	urls := []string{
		"https://www.google.com",
		"invalid-url",
		"http://localhost:8080",
		"https://nonexistent.domain.com",
		"https://ght.magrummy.in/pay/index?order_number=Y711J2025031811561271699068",
	}

	for _, u := range urls {
		if common.IsValidURL(u) {
			fmt.Printf("%s 是有效的链接\n", u)
		} else {
			fmt.Printf("%s 是无效的链接\n", u)
		}
	}
}
