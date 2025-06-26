package dragonpay

import (
	"fmt"
	"goserver/internal/global/pay/common"
	"goserver/internal/global/pay/entity"
	"testing"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"gopkg.in/ini.v1"
)

func TestDragonPay(t *testing.T) {
	a := ""
	b := common.YuanToFen(a)
	fmt.Println(b)

	cfg, err := ini.Load("../../../../configs/app.conf")
	if err != nil {
		panic(err)
	}

	payurl := cfg.Section("").Key("dragonpay.payurl").String()
	withdrawurl := cfg.Section("").Key("dragonpay.withdrawurl").String()
	appid := cfg.Section("").Key("dragonpay.merchant").String()
	serverPubRSA := cfg.Section("").Key("dragonpay.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("dragonpay.clientPriRSA").String()
	md5key := cfg.Section("").Key("dragonpay.md5key").String()
	notifyUrl := cfg.Section("").Key("dragonpay.paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("dragonpay.withDrawnotifyurl").String()

	InitConfig(payurl, withdrawurl, appid, serverPubRSA, clientPriRSA, md5key, notifyUrl, withdrawNotifyUrl)

	id, _ := gonanoid.New()

	id1, _ := gonanoid.New(16)

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
