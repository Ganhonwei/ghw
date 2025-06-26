package config

import (
	"goserver/internal/global/pay/gopay"
	"goserver/internal/global/pay/mjlpay/ai2pay"
	"goserver/internal/global/pay/mjlpay/jypay"
	"goserver/internal/global/pay/mjlpay/mjl66pay"
	"goserver/internal/global/pay/mjlpay/mjlblizzardpay"
	"goserver/internal/global/pay/mjlpay/mjleocpay"
	"goserver/internal/global/pay/mjlpay/mmpay"
	"goserver/internal/global/pay/mjlpay/pay99"
	"goserver/internal/global/pay/mjlpay/shpay"
	"goserver/internal/global/pay/mjlpay/transafepay"
	"goserver/internal/global/pay/mjlpay/unipay"
	"goserver/internal/global/pay/safepay"

	"gopkg.in/ini.v1"
)

func MjlUniPAYInit(cfg *ini.File) {
	appid := cfg.Section("").Key("mjluni.appId").String()
	name := cfg.Section("").Key("mjluni.name").String()
	url := cfg.Section("").Key("mjluni.payUrl").String()
	noticUrl := cfg.Section("").Key("paynotifyurl").String()
	withDrawnotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	unipay.InitConfig(appid, name, url, noticUrl, withDrawnotifyUrl)
}

func MjlJyPAYInit(cfg *ini.File) {
	merNo := cfg.Section("").Key("mjljy.merno").String()
	hmacKey := cfg.Section("").Key("mjljy.hmackey").String()
	pubRSA := cfg.Section("").Key("mjljy.pubrsa").String()
	priRSA := cfg.Section("").Key("mjljy.prirsa").String()
	payUrl := cfg.Section("").Key("mjljy.payUrl").String()
	withdrawUrl := cfg.Section("").Key("mjljy.withdrawUrl").String()
	noticUrl := cfg.Section("").Key("paynotifyurl").String()
	withDrawnotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	jypay.InitConfig(merNo, hmacKey, pubRSA, priRSA, noticUrl, withDrawnotifyUrl, payUrl, withdrawUrl)
}

func MjlPAY99Init(cfg *ini.File) {
	merNo := cfg.Section("").Key("mjl99.merno").String()
	hmacKey := cfg.Section("").Key("mjl99.hmackey").String()
	key := cfg.Section("").Key("mjl99.key").String()
	pubRSA := cfg.Section("").Key("mjl99.pubrsa").String()
	priRSA := cfg.Section("").Key("mjl99.prirsa").String()
	payUrl := cfg.Section("").Key("mjl99.payUrl").String()
	withdrawUrl := cfg.Section("").Key("mjl99.withdrawUrl").String()
	noticUrl := cfg.Section("").Key("paynotifyurl").String()
	withDrawnotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	pay99.InitConfig(merNo, hmacKey, key, pubRSA, priRSA, noticUrl, withDrawnotifyUrl, payUrl, withdrawUrl)
}

func MjlTransafePAYInit(cfg *ini.File) {
	merNo := cfg.Section("").Key("mjltransafe.merno").String()
	appId := cfg.Section("").Key("mjltransafe.appId").String()
	pubRSA := cfg.Section("").Key("mjltransafe.pubrsa").String()
	priRSA := cfg.Section("").Key("mjltransafe.prirsa").String()
	payUrl := cfg.Section("").Key("mjltransafe.redirectUrl").String()
	noticUrl := cfg.Section("").Key("paynotifyurl").String()
	withDrawnotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	transafepay.InitConfig(merNo, appId, pubRSA, priRSA, noticUrl, withDrawnotifyUrl, payUrl)
}

func MjlMMPAYInit(cfg *ini.File) {
	merNo := cfg.Section("").Key("mjlmm.merno").String()
	md5 := cfg.Section("").Key("mjlmm.md5key").String()
	baseUrl := cfg.Section("").Key("mjlmm.baseUrl").String()
	noticUrl := cfg.Section("").Key("paynotifyurl").String()
	withDrawnotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	mmpay.InitConfig(merNo, md5, noticUrl, withDrawnotifyUrl, baseUrl)
}

func MjlSafePayInit(cfg *ini.File) {
	merno := cfg.Section("").Key("safepay.merno").String()
	key := cfg.Section("").Key("safepay.key").String()
	payReturnUrl := cfg.Section("").Key("safepay.payReturnUrl").String()
	withdrawReturnUrl := cfg.Section("").Key("safepay.withdrawReturnUrl").String()
	apiUrl := cfg.Section("").Key("safepay.apiUrl").String()

	safepay.InitConfig(merno, key, payReturnUrl, withdrawReturnUrl, apiUrl)
}

func MjlGoPAYInit(cfg *ini.File) {
	account := cfg.Section("").Key("gopay.account").String()
	key := cfg.Section("").Key("gopay.key").String()
	payNotifyUrl := cfg.Section("").Key("gopay.payNotifyUrl").String()
	withdrawNotifyUrl := cfg.Section("").Key("gopay.withdrawNotifyUrl").String()
	apiUrl := cfg.Section("").Key("gopay.apiUrl").String()

	gopay.InitConfig(account, key, payNotifyUrl, withdrawNotifyUrl, apiUrl)
}

func MjlSHPAYInit(cfg *ini.File) {
	merNo := cfg.Section("").Key("mjlsh.merno").String()
	appId := cfg.Section("").Key("mjlsh.appid").String()
	md5 := cfg.Section("").Key("mjlsh.key").String()
	baseUrl := cfg.Section("").Key("mjlsh.apiUrl").String()
	noticUrl := cfg.Section("").Key("paynotifyurl").String()
	withDrawnotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	shpay.InitConfig(merNo, appId, md5, noticUrl, withDrawnotifyUrl, baseUrl)
}

func MjlBlizzardPayInit(cfg *ini.File) {
	merNo := cfg.Section("").Key("mjlblizzardpay.merno").String()
	hmacKey := cfg.Section("").Key("mjlblizzardpay.hmackey").String()
	key := cfg.Section("").Key("mjlblizzardpay.key").String()
	pubRSA := cfg.Section("").Key("mjlblizzardpay.pubrsa").String()
	priRSA := cfg.Section("").Key("mjlblizzardpay.prirsa").String()
	payUrl := cfg.Section("").Key("mjlblizzardpay.payUrl").String()
	withdrawUrl := cfg.Section("").Key("mjlblizzardpay.withdrawUrl").String()
	noticUrl := cfg.Section("").Key("paynotifyurl").String()
	withDrawnotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	mjlblizzardpay.InitConfig(merNo, hmacKey, key, pubRSA, priRSA, noticUrl, withDrawnotifyUrl, payUrl, withdrawUrl)
}

func MjlAi2PayInit(cfg *ini.File) {
	payurl := cfg.Section("").Key("mjl_ai2.payurl").String()
	withdrawurl := cfg.Section("").Key("mjl_ai2.withdrawurl").String()
	merchant := cfg.Section("").Key("mjl_ai2.merchant").String()
	appid := cfg.Section("").Key("mjl_ai2.appid").String()
	serverPubRSA := cfg.Section("").Key("mjl_ai2.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("mjl_ai2.clientPriRSA").String()
	md5key := cfg.Section("").Key("mjl_ai2.md5key").String()
	payNotifyUrl := cfg.Section("").Key("paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	ai2pay.InitConfig(payurl, withdrawurl, merchant, appid, serverPubRSA, clientPriRSA, md5key, payNotifyUrl, withdrawNotifyUrl)

}

func MjlPay66Init(cfg *ini.File) {
	payurl := cfg.Section("").Key("mjl_66pay.payurl").String()
	withdrawurl := cfg.Section("").Key("mjl_66pay.withdrawurl").String()
	merchant := cfg.Section("").Key("mjl_66pay.merchant").String()
	appid := cfg.Section("").Key("mjl_66pay.appid").String()
	serverPubRSA := cfg.Section("").Key("mjl_66pay.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("mjl_66pay.clientPriRSA").String()
	md5key := cfg.Section("").Key("mjl_66pay.md5key").String()
	payNotifyUrl := cfg.Section("").Key("paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	mjl66pay.InitConfig(payurl, withdrawurl, merchant, appid, serverPubRSA, clientPriRSA, md5key, payNotifyUrl, withdrawNotifyUrl)

}

func MjlEocPayInit(cfg *ini.File) {
	payurl := cfg.Section("").Key("mjl_eocpay.payurl").String()
	withdrawurl := cfg.Section("").Key("mjl_eocpay.withdrawurl").String()
	merchant := cfg.Section("").Key("mjl_eocpay.merchant").String()
	appid := cfg.Section("").Key("mjl_eocpay.appid").String()
	serverPubRSA := cfg.Section("").Key("mjl_eocpay.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("mjl_eocpay.clientPriRSA").String()
	md5key := cfg.Section("").Key("mjl_eocpay.md5key").String()
	payNotifyUrl := cfg.Section("").Key("paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	mjleocpay.InitConfig(payurl, withdrawurl, merchant, appid, serverPubRSA, clientPriRSA, md5key, payNotifyUrl, withdrawNotifyUrl)

}
