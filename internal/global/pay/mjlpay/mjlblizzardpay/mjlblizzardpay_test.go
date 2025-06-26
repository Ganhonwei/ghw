package mjlblizzardpay

import (
	"goserver/internal/global/pay/entity"
	"testing"

	"github.com/google/uuid"
	"gopkg.in/ini.v1"
)

func TestPay99(t *testing.T) {
	cfg, err := ini.Load("../../../../../configs/app.conf")
	if err != nil {
		panic(err)
	}

	merNo := cfg.Section("").Key("mjlblizzardpay.merno").String()
	hmacKey := cfg.Section("").Key("mjlblizzardpay.hmackey").String()
	key := cfg.Section("").Key("mjlblizzardpay.key").String()
	pubRSA := cfg.Section("").Key("mjlblizzardpay.pubrsa").String()
	priRSA := cfg.Section("").Key("mjlblizzardpay.prirsa").String()
	payUrl := cfg.Section("").Key("mjlblizzardpay.payUrl").String()
	withdrawUrl := cfg.Section("").Key("mjlblizzardpay.withdrawUrl").String()
	noticUrl := cfg.Section("").Key("paynotifyurl").String()
	withDrawnotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	InitConfig(merNo, hmacKey, key, pubRSA, priRSA, noticUrl, withDrawnotifyUrl, payUrl, withdrawUrl)

	u4 := uuid.New()

	order := &entity.PayOrder{}
	order.Amount = 100000
	order.OrderID = u4.String()
	order.RegistIp = "127.0.0.1"
	order.Userid = "1234567890"

	resp, _ := Config.Submit(order)

	link, _ := Config.SubmitResponse(resp, order)

	_ = link

	// order := &entity.WithdrawOrder{}
	// order.Amount = 1000
	// order.OrderID = u4.String()
	// // order.IFSC = "UTIB0000430"
	// order.Bank = "bkash"
	// order.BankNumber = "919010054285006"
	// order.RealName = "test"

	// resp, _ := Config.WithdrawSubmit(order)

	// ok, _ := Config.WithdrawSubmitResponse(resp, order)
	// _ = ok

	// b := Config.BalanceQuery()
	// _ = b
}
