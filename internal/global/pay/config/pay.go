package config

import (
	pay1916 "goserver/internal/global/pay/1916pay"
	pay66 "goserver/internal/global/pay/66pay"
	pay9s "goserver/internal/global/pay/9spay"
	"goserver/internal/global/pay/ai2pay"
	"goserver/internal/global/pay/aipay"
	"goserver/internal/global/pay/atpay"
	"goserver/internal/global/pay/blizzardpay"
	"goserver/internal/global/pay/cowpay"
	"goserver/internal/global/pay/csmpay"
	"goserver/internal/global/pay/cxpay"
	"goserver/internal/global/pay/ddpay"
	"goserver/internal/global/pay/dragonpay"
	"goserver/internal/global/pay/eocpay"
	"goserver/internal/global/pay/flypay"
	"goserver/internal/global/pay/gamepay"
	"goserver/internal/global/pay/gentlepay"
	"goserver/internal/global/pay/icepay"
	"goserver/internal/global/pay/kingpay"
	"goserver/internal/global/pay/kkpluspay"
	"goserver/internal/global/pay/leopay"
	"goserver/internal/global/pay/letspay"
	"goserver/internal/global/pay/letspayfast"
	"goserver/internal/global/pay/mepay"
	"goserver/internal/global/pay/metagopay"
	"goserver/internal/global/pay/mlpay"
	"goserver/internal/global/pay/netpay"
	"goserver/internal/global/pay/oepay"
	"goserver/internal/global/pay/paywook"
	"goserver/internal/global/pay/ramapay"
	"goserver/internal/global/pay/sailspay"
	"goserver/internal/global/pay/tansafe"
	"goserver/internal/global/pay/universalpay"
	"goserver/internal/global/pay/upay"
	"goserver/internal/global/pay/usdtpay"
	"goserver/internal/global/pay/uwinpay"
	"goserver/internal/global/pay/wepay"
	"goserver/internal/global/pay/wepay2"
	"goserver/internal/global/pay/wypay"
	"goserver/internal/global/pay/xdpay"
	"goserver/internal/global/pay/xfpay"
	"goserver/internal/global/pay/yunpay"

	"gopkg.in/ini.v1"
)

// 充值配置
func PayConfigInit(cfg *ini.File) {
	MlpayInit(cfg)
	XFPayInit(cfg)
	KingPayInit(cfg)
	SailsPayInit(cfg)
	FlyPayInit(cfg)
	XdPayInit(cfg)
	UWinPayInit(cfg)
	KKPlusPayInit(cfg)
	DragonPayInit(cfg)
	AiPayInit(cfg)
	GamePayInit(cfg)
	GentlePayInit(cfg)
	Pay66Init(cfg)
	Ai2PayInit(cfg)
	CowPayInit(cfg)
	LetspayfastInit(cfg)
	PaywookInit(cfg)
	MepayInit(cfg)
	DdPayInit(cfg)
	AtPayInit(cfg)
	NetPayInit(cfg)
	WypayInit(cfg)
	YunPayInit(cfg)
	CxPayInit(cfg)
	CsmpayInit(cfg)
	WePay2Init(cfg)

	TansafeInit(cfg)
	Pay1916Init(cfg)
	BlizzardPayInit(cfg)
	LetsPayInit(cfg)
	RamaPayInit(cfg)
	IcePayInit(cfg)
	WePayInit(cfg)
	OePayConfig(cfg)
	Pay9sInit(cfg)
	MetagopayPayInit(cfg)
	UniversalpayInit(cfg)
	LeoPayInit(cfg)
	UPayInit(cfg)
	EocPayInit(cfg)

	// ustd
	UstdPayInit(cfg)

	// 孟加拉
	MjlUniPAYInit(cfg)
	MjlJyPAYInit(cfg)
	MjlPAY99Init(cfg)
	MjlTransafePAYInit(cfg)
	MjlMMPAYInit(cfg)
	MjlSafePayInit(cfg)
	MjlGoPAYInit(cfg)
	MjlSHPAYInit(cfg)
	MjlBlizzardPayInit(cfg)
	MjlAi2PayInit(cfg)
	MjlPay66Init(cfg)
	MjlEocPayInit(cfg)

	//巴基斯坦
	BjstAi2PayInit(cfg)
	BjstPakPayInit(cfg)
	BjstEocPayInit(cfg)

}

// MLPAY支付配置加载
func MlpayInit(cfg *ini.File) {
	url := cfg.Section("").Key("mlpay.payorderurl").String()
	appid := cfg.Section("").Key("mlpay.appid").String()
	applicationId := cfg.Section("").Key("mlpay.applicationId").String()
	md5 := cfg.Section("").Key("mlpay.md5").String()
	withdrawMd5 := cfg.Section("").Key("mlpay.withDrawMd5").String()
	noticUrl := cfg.Section("").Key("mlpay.paynotifyurl").String()
	withDrawnotifyUrl := cfg.Section("").Key("mlpay.withDrawnotifyurl").String()

	mlpay.InitConfig(url, appid, applicationId, md5, withdrawMd5, noticUrl, withDrawnotifyUrl)
}

func Pay1916Init(cfg *ini.File) {
	url := cfg.Section("").Key("1916pay.payorderurl").String()
	appid := cfg.Section("").Key("1916pay.appid").String()
	applicationId := cfg.Section("").Key("1916pay.applicationId").String()
	md5 := cfg.Section("").Key("1916pay.md5").String()
	withdrawMd5 := cfg.Section("").Key("1916pay.withDrawMd5").String()
	noticUrl := cfg.Section("").Key("1916pay.paynotifyurl").String()
	withDrawnotifyUrl := cfg.Section("").Key("1916pay.withDrawnotifyurl").String()

	pay1916.InitConfig(url, appid, applicationId, md5, withdrawMd5, noticUrl, withDrawnotifyUrl)
}

func BlizzardPayInit(cfg *ini.File) {
	merNo := cfg.Section("").Key("blizzardpay.merno").String()
	hmacKey := cfg.Section("").Key("blizzardpay.hmackey").String()
	key := cfg.Section("").Key("blizzardpay.key").String()
	pubRSA := cfg.Section("").Key("blizzardpay.pubrsa").String()
	priRSA := cfg.Section("").Key("blizzardpay.prirsa").String()
	payUrl := cfg.Section("").Key("blizzardpay.payUrl").String()
	withdrawUrl := cfg.Section("").Key("blizzardpay.withdrawUrl").String()
	noticUrl := cfg.Section("").Key("paynotifyurl").String()
	withDrawnotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	blizzardpay.InitConfig(merNo, hmacKey, key, pubRSA, priRSA, noticUrl, withDrawnotifyUrl, payUrl, withdrawUrl)
}

func LetsPayInit(cfg *ini.File) {
	url := cfg.Section("").Key("letspay.payorderurl").String()
	checkUrl := cfg.Section("").Key("letspay.checkurl").String()
	appid := cfg.Section("").Key("letspay.appid").String()
	md5 := cfg.Section("").Key("letspay.md5").String()
	noticUrl := cfg.Section("").Key("letspay.paynotifyurl").String()
	withDrawnotifyUrl := cfg.Section("").Key("letspay.withDrawnotifyurl").String()

	letspay.InitConfig(url, checkUrl, appid, md5, noticUrl, withDrawnotifyUrl)
}

// XFPAY支付配置加载
func XFPayInit(cfg *ini.File) {
	url := cfg.Section("").Key("xfpay.payorderurl").String()
	machId := cfg.Section("").Key("xfpay.machid").String()
	appid := cfg.Section("").Key("xfpay.appid").String()
	md5 := cfg.Section("").Key("xfpay.md5").String()
	withdrawMd5 := cfg.Section("").Key("xfpay.withDrawMd5").String()
	noticUrl := cfg.Section("").Key("xfpay.paynotifyurl").String()
	withDrawnotifyUrl := cfg.Section("").Key("xfpay.withDrawnotifyurl").String()
	xfpay.InitConfig(url, appid, machId, md5, withdrawMd5, noticUrl, withDrawnotifyUrl)
}

// kingpay支付配置加载
func KingPayInit(cfg *ini.File) {
	url := cfg.Section("").Key("kingpay.payorderurl").String()
	appid := cfg.Section("").Key("kingpay.appid").String()
	applicationId := cfg.Section("").Key("kingpay.applicationId").String()
	serverPubRSA := cfg.Section("").Key("kingpay.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("kingpay.clientPriRSA").String()
	noticUrl := cfg.Section("").Key("kingpay.paynotifyurl").String()
	withDrawnotifyUrl := cfg.Section("").Key("kingpay.withDrawnotifyurl").String()
	kingpay.InitConfig(url, appid, applicationId, serverPubRSA, clientPriRSA, noticUrl, withDrawnotifyUrl)
}

// sailspay支付配置加载
func SailsPayInit(cfg *ini.File) {
	url := cfg.Section("").Key("sailspay.payorderurl").String()
	md5 := cfg.Section("").Key("sailspay.md5").String()
	appid := cfg.Section("").Key("sailspay.machid").String()
	noticUrl := cfg.Section("").Key("sailspay.paynotifyurl").String()
	withDrawnotifyUrl := cfg.Section("").Key("sailspay.withDrawnotifyurl").String()
	sailspay.InitConfig(url, appid, md5, noticUrl, withDrawnotifyUrl)
}

// flypay支付配置加载
func FlyPayInit(cfg *ini.File) {
	url := cfg.Section("").Key("flypay.payorderurl").String()
	md5 := cfg.Section("").Key("flypay.md5").String()
	spublickey := cfg.Section("").Key("flypay.serverpublickey").String()
	privatekey := cfg.Section("").Key("flypay.privatekey").String()
	publickey := cfg.Section("").Key("flypay.publickey").String()
	appid := cfg.Section("").Key("flypay.machid").String()
	noticUrl := cfg.Section("").Key("flypay.paynotifyurl").String()
	withDrawnotifyUrl := cfg.Section("").Key("flypay.withDrawnotifyurl").String()
	flypay.InitConfig(url, appid, md5, spublickey, privatekey, publickey, noticUrl, withDrawnotifyUrl)
}

// xdpay支付配置加载
func XdPayInit(cfg *ini.File) {
	url := cfg.Section("").Key("xdpay.payorderurl").String()
	paycode := cfg.Section("").Key("xdpay.paycode").String()
	withdrawcode := cfg.Section("").Key("xdpay.withdrawcode").String()
	md5 := cfg.Section("").Key("xdpay.md5").String()
	appid := cfg.Section("").Key("xdpay.machid").String()
	noticUrl := cfg.Section("").Key("xdpay.paynotifyurl").String()
	withDrawnotifyUrl := cfg.Section("").Key("xdpay.withDrawnotifyurl").String()
	xdpay.InitConfig(url, appid, paycode, withdrawcode, md5, noticUrl, withDrawnotifyUrl)
}

// uwinpay支付配置加载
func UWinPayInit(cfg *ini.File) {
	url := cfg.Section("").Key("uwinpay.orderurl").String()
	appid := cfg.Section("").Key("uwinpay.merchant").String()
	serverPubRSA := cfg.Section("").Key("uwinpay.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("uwinpay.clientPriRSA").String()
	noticUrl := cfg.Section("").Key("uwinpay.paynotifyurl").String()
	withDrawnotifyUrl := cfg.Section("").Key("uwinpay.withDrawnotifyurl").String()
	uwinpay.InitConfig(url, appid, serverPubRSA, clientPriRSA, noticUrl, withDrawnotifyUrl)
}

// kkpluspay支付配置加载
func KKPlusPayInit(cfg *ini.File) {
	payurl := cfg.Section("").Key("kkpluspay.payurl").String()
	withdrawurl := cfg.Section("").Key("kkpluspay.withdrawurl").String()
	appid := cfg.Section("").Key("kkpluspay.merchant").String()
	serverPubRSA := cfg.Section("").Key("kkpluspay.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("kkpluspay.clientPriRSA").String()
	md5key := cfg.Section("").Key("kkpluspay.md5key").String()
	notifyUrl := cfg.Section("").Key("kkpluspay.paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("kkpluspay.withDrawnotifyurl").String()
	kkpluspay.InitConfig(payurl, withdrawurl, appid, serverPubRSA, clientPriRSA, md5key, notifyUrl, withdrawNotifyUrl)
}

func DragonPayInit(cfg *ini.File) {
	payurl := cfg.Section("").Key("dragonpay.payurl").String()
	withdrawurl := cfg.Section("").Key("dragonpay.withdrawurl").String()
	appid := cfg.Section("").Key("dragonpay.merchant").String()
	serverPubRSA := cfg.Section("").Key("dragonpay.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("kkpluspay.clientPriRSA").String()
	md5key := cfg.Section("").Key("dragonpay.md5key").String()
	notifyUrl := cfg.Section("").Key("dragonpay.paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("dragonpay.withDrawnotifyurl").String()
	dragonpay.InitConfig(payurl, withdrawurl, appid, serverPubRSA, clientPriRSA, md5key, notifyUrl, withdrawNotifyUrl)
}

func AiPayInit(cfg *ini.File) {
	payurl := cfg.Section("").Key("aipay.payurl").String()
	withdrawurl := cfg.Section("").Key("aipay.withdrawurl").String()
	merchant := cfg.Section("").Key("aipay.merchant").String()
	appid := cfg.Section("").Key("aipay.appid").String()
	serverPubRSA := cfg.Section("").Key("aipay.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("aipay.clientPriRSA").String()
	md5key := cfg.Section("").Key("aipay.md5key").String()
	payNotifyUrl := cfg.Section("").Key("paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	aipay.InitConfig(payurl, withdrawurl, merchant, appid, serverPubRSA, clientPriRSA, md5key, payNotifyUrl, withdrawNotifyUrl)
}

func GamePayInit(cfg *ini.File) {
	payurl := cfg.Section("").Key("gamepay.payurl").String()
	withdrawurl := cfg.Section("").Key("gamepay.withdrawurl").String()
	merchant := cfg.Section("").Key("gamepay.merchant").String()
	appid := cfg.Section("").Key("gamepay.appid").String()
	serverPubRSA := cfg.Section("").Key("gamepay.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("gamepay.clientPriRSA").String()
	md5key := cfg.Section("").Key("gamepay.md5key").String()
	payNotifyUrl := cfg.Section("").Key("paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	gamepay.InitConfig(payurl, withdrawurl, merchant, appid, serverPubRSA, clientPriRSA, md5key, payNotifyUrl, withdrawNotifyUrl)
}

func GentlePayInit(cfg *ini.File) {
	payurl := cfg.Section("").Key("gentlepay.payurl").String()
	withdrawurl := cfg.Section("").Key("gentlepay.withdrawurl").String()
	merchant := cfg.Section("").Key("gentlepay.merchant").String()
	appid := cfg.Section("").Key("gentlepay.appid").String()
	serverPubRSA := cfg.Section("").Key("gentlepay.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("gentlepay.clientPriRSA").String()
	md5key := cfg.Section("").Key("gentlepay.md5key").String()
	payNotifyUrl := cfg.Section("").Key("paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	gentlepay.InitConfig(payurl, withdrawurl, merchant, appid, serverPubRSA, clientPriRSA, md5key, payNotifyUrl, withdrawNotifyUrl)
}

func Pay66Init(cfg *ini.File) {
	payurl := cfg.Section("").Key("pay66.payurl").String()
	withdrawurl := cfg.Section("").Key("pay66.withdrawurl").String()
	merchant := cfg.Section("").Key("pay66.merchant").String()
	appid := cfg.Section("").Key("pay66.appid").String()
	serverPubRSA := cfg.Section("").Key("pay66.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("pay66.clientPriRSA").String()
	md5key := cfg.Section("").Key("pay66.md5key").String()
	payNotifyUrl := cfg.Section("").Key("paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	pay66.InitConfig(payurl, withdrawurl, merchant, appid, serverPubRSA, clientPriRSA, md5key, payNotifyUrl, withdrawNotifyUrl)

}

func Ai2PayInit(cfg *ini.File) {
	payurl := cfg.Section("").Key("ai2pay.payurl").String()
	withdrawurl := cfg.Section("").Key("ai2pay.withdrawurl").String()
	merchant := cfg.Section("").Key("ai2pay.merchant").String()
	appid := cfg.Section("").Key("ai2pay.appid").String()
	serverPubRSA := cfg.Section("").Key("ai2pay.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("ai2pay.clientPriRSA").String()
	md5key := cfg.Section("").Key("ai2pay.md5key").String()
	payNotifyUrl := cfg.Section("").Key("paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	ai2pay.InitConfig(payurl, withdrawurl, merchant, appid, serverPubRSA, clientPriRSA, md5key, payNotifyUrl, withdrawNotifyUrl)

}

func CowPayInit(cfg *ini.File) {
	payurl := cfg.Section("").Key("cowpay.payurl").String()
	withdrawurl := cfg.Section("").Key("cowpay.withdrawurl").String()
	merchant := cfg.Section("").Key("cowpay.merchant").String()
	appid := cfg.Section("").Key("cowpay.appid").String()
	serverPubRSA := cfg.Section("").Key("cowpay.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("cowpay.clientPriRSA").String()
	md5key := cfg.Section("").Key("cowpay.md5key").String()
	payNotifyUrl := cfg.Section("").Key("paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	cowpay.InitConfig(payurl, withdrawurl, merchant, appid, serverPubRSA, clientPriRSA, md5key, payNotifyUrl, withdrawNotifyUrl)
}

func LetspayfastInit(cfg *ini.File) {
	payurl := cfg.Section("").Key("letspayfast.payurl").String()
	withdrawurl := cfg.Section("").Key("letspayfast.withdrawurl").String()
	merchant := cfg.Section("").Key("letspayfast.merchant").String()
	appid := cfg.Section("").Key("letspayfast.appid").String()
	serverPubRSA := cfg.Section("").Key("letspayfast.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("letspayfast.clientPriRSA").String()
	md5key := cfg.Section("").Key("letspayfast.md5key").String()
	payNotifyUrl := cfg.Section("").Key("paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	letspayfast.InitConfig(payurl, withdrawurl, merchant, appid, serverPubRSA, clientPriRSA, md5key, payNotifyUrl, withdrawNotifyUrl)
}

func PaywookInit(cfg *ini.File) {
	payurl := cfg.Section("").Key("paywook.payurl").String()
	withdrawurl := cfg.Section("").Key("paywook.withdrawurl").String()
	merchant := cfg.Section("").Key("paywook.merchant").String()
	appid := cfg.Section("").Key("paywook.appid").String()
	serverPubRSA := cfg.Section("").Key("paywook.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("paywook.clientPriRSA").String()
	md5key := cfg.Section("").Key("paywook.md5key").String()
	payNotifyUrl := cfg.Section("").Key("paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	paywook.InitConfig(payurl, withdrawurl, merchant, appid, serverPubRSA, clientPriRSA, md5key, payNotifyUrl, withdrawNotifyUrl)
}

func MepayInit(cfg *ini.File) {
	payurl := cfg.Section("").Key("mepay.payurl").String()
	withdrawurl := cfg.Section("").Key("mepay.withdrawurl").String()
	merchant := cfg.Section("").Key("mepay.merchant").String()
	appid := cfg.Section("").Key("mepay.appid").String()
	serverPubRSA := cfg.Section("").Key("mepay.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("mepay.clientPriRSA").String()
	md5key := cfg.Section("").Key("mepay.md5key").String()
	payNotifyUrl := cfg.Section("").Key("paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	mepay.InitConfig(payurl, withdrawurl, merchant, appid, serverPubRSA, clientPriRSA, md5key, payNotifyUrl, withdrawNotifyUrl)

}

func DdPayInit(cfg *ini.File) {
	payurl := cfg.Section("").Key("ddpay.payurl").String()
	withdrawurl := cfg.Section("").Key("ddpay.withdrawurl").String()
	merchant := cfg.Section("").Key("ddpay.merchant").String()
	appid := cfg.Section("").Key("ddpay.appid").String()
	serverPubRSA := cfg.Section("").Key("ddpay.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("ddpay.clientPriRSA").String()
	md5key := cfg.Section("").Key("ddpay.md5key").String()
	payNotifyUrl := cfg.Section("").Key("paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	ddpay.InitConfig(payurl, withdrawurl, merchant, appid, serverPubRSA, clientPriRSA, md5key, payNotifyUrl, withdrawNotifyUrl)
}

func AtPayInit(cfg *ini.File) {
	payurl := cfg.Section("").Key("atpay.payurl").String()
	withdrawurl := cfg.Section("").Key("atpay.withdrawurl").String()
	merchant := cfg.Section("").Key("atpay.merchant").String()
	appid := cfg.Section("").Key("atpay.appid").String()
	serverPubRSA := cfg.Section("").Key("atpay.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("atpay.clientPriRSA").String()
	md5key := cfg.Section("").Key("atpay.md5key").String()
	payNotifyUrl := cfg.Section("").Key("paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	atpay.InitConfig(payurl, withdrawurl, merchant, appid, serverPubRSA, clientPriRSA, md5key, payNotifyUrl, withdrawNotifyUrl)
}

func NetPayInit(cfg *ini.File) {

	payurl := cfg.Section("").Key("netpay.payurl").String()
	withdrawurl := cfg.Section("").Key("netpay.withdrawurl").String()
	merchant := cfg.Section("").Key("netpay.merchant").String()
	appid := cfg.Section("").Key("netpay.appid").String()
	serverPubRSA := cfg.Section("").Key("netpay.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("netpay.clientPriRSA").String()
	md5keyin := cfg.Section("").Key("netpay.md5keyin").String()
	md5keyout := cfg.Section("").Key("netpay.md5keyout").String()
	payNotifyUrl := cfg.Section("").Key("paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	netpay.InitConfig(payurl, withdrawurl, merchant, appid, serverPubRSA, clientPriRSA, md5keyin, md5keyout, payNotifyUrl, withdrawNotifyUrl)
}

func WypayInit(cfg *ini.File) {
	payurl := cfg.Section("").Key("wypay.payurl").String()
	withdrawurl := cfg.Section("").Key("wypay.withdrawurl").String()
	merchant := cfg.Section("").Key("wypay.merchant").String()
	appid := cfg.Section("").Key("wypay.appid").String()
	serverPubRSA := cfg.Section("").Key("wypay.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("wypay.clientPriRSA").String()
	md5key := cfg.Section("").Key("wypay.md5key").String()
	payNotifyUrl := cfg.Section("").Key("paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	wypay.InitConfig(payurl, withdrawurl, merchant, appid, serverPubRSA, clientPriRSA, md5key, payNotifyUrl, withdrawNotifyUrl)
}

func YunPayInit(cfg *ini.File) {
	payurl := cfg.Section("").Key("yunpay.payurl").String()
	withdrawurl := cfg.Section("").Key("yunpay.withdrawurl").String()
	merchant := cfg.Section("").Key("yunpay.merchant").String()
	appid := cfg.Section("").Key("yunpay.appid").String()
	serverPubRSA := cfg.Section("").Key("yunpay.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("yunpay.clientPriRSA").String()
	md5key := cfg.Section("").Key("yunpay.md5key").String()
	payNotifyUrl := cfg.Section("").Key("paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	yunpay.InitConfig(payurl, withdrawurl, merchant, appid, serverPubRSA, clientPriRSA, md5key, payNotifyUrl, withdrawNotifyUrl)
}

func CxPayInit(cfg *ini.File) {
	payurl := cfg.Section("").Key("cxpay.payurl").String()
	withdrawurl := cfg.Section("").Key("cxpay.withdrawurl").String()
	merchant := cfg.Section("").Key("cxpay.merchant").String()
	appid := cfg.Section("").Key("cxpay.appid").String()
	serverPubRSA := cfg.Section("").Key("cxpay.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("cxpay.clientPriRSA").String()
	md5key := cfg.Section("").Key("cxpay.md5key").String()
	payNotifyUrl := cfg.Section("").Key("paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	cxpay.InitConfig(payurl, withdrawurl, merchant, appid, serverPubRSA, clientPriRSA, md5key, payNotifyUrl, withdrawNotifyUrl)
}

func CsmpayInit(cfg *ini.File) {
	payurl := cfg.Section("").Key("csmpay.payurl").String()
	withdrawurl := cfg.Section("").Key("csmpay.withdrawurl").String()
	merchant := cfg.Section("").Key("csmpay.merchant").String()
	appid := cfg.Section("").Key("csmpay.appid").String()
	serverPubRSA := cfg.Section("").Key("csmpay.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("csmpay.clientPriRSA").String()
	md5key := cfg.Section("").Key("csmpay.md5key").String()
	payNotifyUrl := cfg.Section("").Key("paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	csmpay.InitConfig(payurl, withdrawurl, merchant, appid, serverPubRSA, clientPriRSA, md5key, payNotifyUrl, withdrawNotifyUrl)
}

func WePay2Init(cfg *ini.File) {
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

	wepay2.InitConfig(payurl, withdrawurl, merchant, appid, serverPubRSA, clientPriRSA, pay_md5key, withdraw_md5key, payNotifyUrl, withdrawNotifyUrl, pay_type)
}

// tansafe支付配置加载
func TansafeInit(cfg *ini.File) {
	url := cfg.Section("").Key("tansafe.payorderurl").String()
	machid := cfg.Section("").Key("tansafe.machid").String()
	channelid := cfg.Section("").Key("tansafe.channelid").String()
	appid := cfg.Section("").Key("tansafe.machid").String()
	serverpublickey := cfg.Section("").Key("tansafe.serverpublickey").String()
	privatekey := cfg.Section("").Key("tansafe.privatekey").String()
	publickey := cfg.Section("").Key("tansafe.publickey").String()
	noticUrl := cfg.Section("").Key("tansafe.paynotifyurl").String()
	withDrawnotifyUrl := cfg.Section("").Key("tansafe.withDrawnotifyurl").String()
	tansafe.InitConfig(url, machid, channelid, appid, serverpublickey, privatekey, publickey, noticUrl, withDrawnotifyUrl)
}

// ramapay支付配置加载
func RamaPayInit(cfg *ini.File) {
	url := cfg.Section("").Key("ramapay.orderurl").String()
	appid := cfg.Section("").Key("ramapay.merchant").String()
	serverPubRSA := cfg.Section("").Key("ramapay.serverPubRSA").String()
	clientPriRSA := cfg.Section("").Key("ramapay.clientPriRSA").String()
	noticUrl := cfg.Section("").Key("ramapay.paynotifyurl").String()
	withDrawnotifyUrl := cfg.Section("").Key("ramapay.withDrawnotifyurl").String()
	ramapay.InitConfig(url, appid, serverPubRSA, clientPriRSA, noticUrl, withDrawnotifyUrl)
}

func IcePayInit(cfg *ini.File) {
	url := cfg.Section("").Key("icepay.payorderurl").String()
	checkUrl := cfg.Section("").Key("icepay.payorderurl").String()
	appid := cfg.Section("").Key("icepay.appid").String()
	md5 := cfg.Section("").Key("icepay.md5").String()
	noticUrl := cfg.Section("").Key("icepay.paynotifyurl").String()
	withDrawnotifyUrl := cfg.Section("").Key("icepay.withDrawnotifyurl").String()

	icepay.InitConfig(url, checkUrl, appid, md5, noticUrl, withDrawnotifyUrl)
}

func WePayInit(cfg *ini.File) {
	url := cfg.Section("").Key("wepay.payorderurl").String()
	checkUrl := cfg.Section("").Key("wepay.checkurl").String()
	appid := cfg.Section("").Key("wepay.appid").String()
	payPassageId := cfg.Section("").Key("wepay.payPassageId").String()
	withdrawPassageId := cfg.Section("").Key("wepay.withdrawPassageId").String()
	md5 := cfg.Section("").Key("wepay.md5").String()
	noticUrl := cfg.Section("").Key("wepay.paynotifyurl").String()
	withDrawnotifyUrl := cfg.Section("").Key("wepay.withDrawnotifyurl").String()

	wepay.InitConfig(url, checkUrl, appid, payPassageId, withdrawPassageId, md5, noticUrl, withDrawnotifyUrl)
}

func OePayConfig(cfg *ini.File) {
	channelId := cfg.Section("").Key("oepay.channelId").String()
	keyServerPub := cfg.Section("").Key("oepay.keyServerPub").String()
	keyClientPriv := cfg.Section("").Key("oepay.keyClientPriv").String()
	keyClientPub := cfg.Section("").Key("oepay.keyClientPub").String()
	payUrl := cfg.Section("").Key("oepay.payUrl").String()
	successUrl := cfg.Section("").Key("oepay.successUrl").String()
	failUrl := cfg.Section("").Key("oepay.failUrl").String()

	oepay.InitConfig(channelId, keyServerPub, keyClientPriv, keyClientPub, payUrl, successUrl, failUrl)
}

func Pay9sInit(cfg *ini.File) {
	appid := cfg.Section("").Key("9spay.appId").String()
	url := cfg.Section("").Key("9spay.payUrl").String()
	chargeId := cfg.Section("").Key("9spay.chargeId").String()
	md5 := cfg.Section("").Key("9spay.md5").String()
	withdrawId := cfg.Section("").Key("9spay.withdrawId").String()
	noticUrl := cfg.Section("").Key("paynotifyurl").String()
	withDrawnotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	pay9s.InitConfig(url, appid, chargeId, withdrawId, md5, noticUrl, withDrawnotifyUrl)
}

func MetagopayPayInit(cfg *ini.File) {
	orgId := cfg.Section("").Key("metagopay.orgId").String()
	mchId := cfg.Section("").Key("metagopay.mchId").String()
	accountId := cfg.Section("").Key("metagopay.accountId").String()
	md5Key := cfg.Section("").Key("metagopay.md5Key").String()
	payNotifyUrl := cfg.Section("").Key("metagopay.payNotifyUrl").String()
	withdrawNotifyUrl := cfg.Section("").Key("metagopay.withdrawNotifyUrl").String()
	metagopay.InitConfig(orgId, mchId, accountId, md5Key, payNotifyUrl, withdrawNotifyUrl)
}

func UniversalpayInit(cfg *ini.File) {
	appid := cfg.Section("").Key("universalpay.appid").String()
	prirsaClient := cfg.Section("").Key("universalpay.prirsaClient").String()
	pubrsaClient := cfg.Section("").Key("universalpay.pubrsaClient").String()
	pubrsaServer := cfg.Section("").Key("universalpay.pubrsaServer").String()
	payNotifyUrl := cfg.Section("").Key("universalpay.payNotifyUrl").String()
	withdrawNotifyUrl := cfg.Section("").Key("universalpay.withdrawNotifyUrl").String()
	universalpay.InitConfig(appid, prirsaClient, pubrsaClient, pubrsaServer, payNotifyUrl, withdrawNotifyUrl)
}

func UstdPayInit(cfg *ini.File) {
	md5key := cfg.Section("").Key("ustdpay.md5key").String()
	payOrderUrl := cfg.Section("").Key("ustdpay.payOrderUrl").String()
	withDarwUrl := cfg.Section("").Key("ustdpay.withDarwUrl").String()
	payNotifyUrl := cfg.Section("").Key("ustdpay.payNotifyUrl").String()
	withDrawNotifyUrl := cfg.Section("").Key("ustdpay.withDrawNotifyUrl").String()
	redirectUrl := cfg.Section("").Key("ustdpay.redirectUrl").String()
	usdtpay.InitConfig(md5key, payOrderUrl, withDarwUrl, payNotifyUrl, withDrawNotifyUrl, redirectUrl)
}

func LeoPayInit(cfg *ini.File) {
	payurl := cfg.Section("").Key("leopay.payurl").String()
	withdrawurl := cfg.Section("").Key("leopay.withdrawurl").String()
	merchant := cfg.Section("").Key("leopay.merchant").String()
	md5key := cfg.Section("").Key("leopay.md5key").String()
	payNotifyUrl := cfg.Section("").Key("paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	leopay.InitConfig(payurl, withdrawurl, merchant, "", "", "", md5key, payNotifyUrl, withdrawNotifyUrl)
}

func UPayInit(cfg *ini.File) {
	payurl := cfg.Section("").Key("upay.payurl").String()
	withdrawurl := cfg.Section("").Key("upay.withdrawurl").String()
	merchant := cfg.Section("").Key("upay.merchantCode").String()
	appid := cfg.Section("").Key("upay.appid").String()
	md5key := cfg.Section("").Key("upay.md5key").String()
	payNotifyUrl := cfg.Section("").Key("paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	upay.InitConfig(payurl, withdrawurl, merchant, appid, "", "", md5key, payNotifyUrl, withdrawNotifyUrl)
}

func EocPayInit(cfg *ini.File) {
	payurl := cfg.Section("").Key("eocpay.payurl").String()
	withdrawurl := cfg.Section("").Key("eocpay.withdrawurl").String()
	merchant := cfg.Section("").Key("eocpay.merchant").String()
	appid := cfg.Section("").Key("eocpay.appid").String()
	md5key := cfg.Section("").Key("eocpay.md5key").String()
	payNotifyUrl := cfg.Section("").Key("paynotifyurl").String()
	withdrawNotifyUrl := cfg.Section("").Key("withDrawnotifyurl").String()

	eocpay.InitConfig(payurl, withdrawurl, merchant, appid, "", "", md5key, payNotifyUrl, withdrawNotifyUrl)
}
