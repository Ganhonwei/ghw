package config

import (
	"goserver/internal/global/pay/bjstpay/ai2pay"
	"goserver/internal/global/pay/bjstpay/bjsteocpay"
	"goserver/internal/global/pay/bjstpay/pakpay"

	"gopkg.in/ini.v1"
)

func BjstAi2PayInit(cfg *ini.File) {
	payurl := cfg.Section("").Key("bjst_ai2.payurl").String()
	withdrawurl := cfg.Section("").Key("bjst_ai2.withdrawurl").String()
	merchant := cfg.Section("").Key("bjst_ai2.merchant").String()
	appid := cfg.Section("").Key("bjst_ai2.appid").String()
	serverPubRSA := cfg.Section("").Key("bjst_ai2.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("bjst_ai2.clientPriRSA").String()
	md5key := cfg.Section("").Key("bjst_ai2.md5key").String()
	payNotifyUrl := cfg.Section("").Key("paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	ai2pay.InitConfig(payurl, withdrawurl, merchant, appid, serverPubRSA, clientPriRSA, md5key, payNotifyUrl, withdrawNotifyUrl)

}

func BjstPakPayInit(cfg *ini.File) {
	payurl := cfg.Section("").Key("bjst_pakpay.payurl").String()
	withdrawurl := cfg.Section("").Key("bjst_pakpay.withdrawurl").String()
	merchant := cfg.Section("").Key("bjst_pakpay.merchant").String()
	appid := cfg.Section("").Key("bjst_pakpay.appid").String()
	serverPubRSA := cfg.Section("").Key("bjst_pakpay.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("bjst_pakpay.clientPriRSA").String()
	md5key := cfg.Section("").Key("bjst_pakpay.md5key").String()
	payNotifyUrl := cfg.Section("").Key("paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	pakpay.InitConfig(payurl, withdrawurl, merchant, appid, serverPubRSA, clientPriRSA, md5key, payNotifyUrl, withdrawNotifyUrl)
}

func BjstEocPayInit(cfg *ini.File) {
	payurl := cfg.Section("").Key("bjst_eocpay.payurl").String()
	withdrawurl := cfg.Section("").Key("bjst_eocpay.withdrawurl").String()
	merchant := cfg.Section("").Key("bjst_eocpay.merchant").String()
	appid := cfg.Section("").Key("bjst_eocpay.appid").String()
	serverPubRSA := cfg.Section("").Key("bjst_eocpay.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("bjst_eocpay.clientPriRSA").String()
	md5key := cfg.Section("").Key("bjst_eocpay.md5key").String()
	payNotifyUrl := cfg.Section("").Key("paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	bjsteocpay.InitConfig(payurl, withdrawurl, merchant, appid, serverPubRSA, clientPriRSA, md5key, payNotifyUrl, withdrawNotifyUrl)
}
