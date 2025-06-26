package pay99

import (
	"testing"

	"gopkg.in/ini.v1"
)

func TestPay99(t *testing.T) {
	cfg, err := ini.Load("app.conf")
	if err != nil {
		panic(err)
	}

	merNo := cfg.Section("").Key("mjl99.merno").String()
	hmacKey := cfg.Section("").Key("mjl99.hmackey").String()
	key := cfg.Section("").Key("mjl99.key").String()
	pubRSA := cfg.Section("").Key("mjl99.pubrsa").String()
	priRSA := cfg.Section("").Key("mjl99.prirsa").String()
	payUrl := cfg.Section("").Key("mjl99.payUrl").String()
	withdrawUrl := cfg.Section("").Key("mjl99.withdrawUrl").String()
	noticUrl := cfg.Section("").Key("paynotifyurl").String()
	withDrawnotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	InitConfig(merNo, hmacKey, key, pubRSA, priRSA, noticUrl, withDrawnotifyUrl, payUrl, withdrawUrl)

	// u4 := uuid.New()

	// order := &entity.PayOrder{}
	// order.Amount = 10000
	// order.OrderID = u4.String()

	// resp, _ := Config.Submit(order)

	// link, _ := Config.SubmitResponse(resp, order)

	// _ = link

	// order := &entity.WithdrawOrder{}
	// order.Amount = 20000
	// order.OrderID = u4.String()
	// order.Bank = "nagad"
	// order.BankNumber = "01622222222"
	// order.RealName = "test"

	// resp, _ := Config.WithdrawSubmit(order)

	// ok, _ := Config.WithdrawSubmitResponse(resp, order)
	// _ = ok

	b := Config.BalanceQuery()
	_ = b
}
