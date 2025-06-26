package sms

import ini "gopkg.in/ini.v1"

var (
	// kmi
	SmskmiAccKey string
	SmskmiSecKey string
	SmskmiURL    string
	SmskmiSender string
	SmskmiTemp   string

	// 孟加拉kmi
	SmsMJLkmiAccKey string
	SmsMJLkmiSecKey string
	SmsMJLkmiURL    string
	SmsMJLkmiSender string
	SmsMJLkmiTemp   string

	// 星信云
	SmsxxyAppKey    string
	SmsxxyAppSecret string
	SmsxxyAppCode   string
	SmsxxyURL       string
	SmsxxySender    string
	SmsxxyTemp      string

	// motion
	SmsmtAppKey    string
	SmsmtAppSecret string
	SmsmtAppCode   string
	SmsmtURL       string
	SmsmtSender    string
	SmsmtTemp      string

	// 创蓝
	SmsclApiKey   string
	SmsclPassword string
	SmsclUrl      string
)

func SmsInit() {
	cfg, err := ini.Load("config/sms.ini")
	if err != nil {
		panic("init sms error")
	}
	SmskmiAccKey = cfg.Section("smskmi").Key("accessKey").Value()
	SmskmiSecKey = cfg.Section("smskmi").Key("secretKey").Value()
	SmskmiURL = cfg.Section("smskmi").Key("url").Value()
	SmskmiSender = cfg.Section("smskmi").Key("sender").Value()
	SmskmiTemp = cfg.Section("smskmi").Key("template").Value()

	SmsxxyAppKey = cfg.Section("smsxxy").Key("appkey").Value()
	SmsxxyAppSecret = cfg.Section("smsxxy").Key("appsecret").Value()
	SmsxxyAppCode = cfg.Section("smsxxy").Key("appcode").Value()
	SmsxxyURL = cfg.Section("smsxxy").Key("url").Value()
	SmsxxySender = cfg.Section("smsxxy").Key("sender").Value()
	SmsxxyTemp = cfg.Section("smsxxy").Key("template").Value()

	SmsmtAppKey = cfg.Section("smsmt").Key("appkey").Value()
	SmsmtAppSecret = cfg.Section("smsmt").Key("appsecret").Value()
	SmsmtAppCode = cfg.Section("smsmt").Key("appcode").Value()
	SmsmtURL = cfg.Section("smsmt").Key("url").Value()
	SmsmtSender = cfg.Section("smsmt").Key("sender").Value()
	SmsmtTemp = cfg.Section("smsmt").Key("template").Value()

	SmsclApiKey = cfg.Section("smscl").Key("apiaccount").Value()
	SmsclPassword = cfg.Section("smscl").Key("password").Value()
	SmsclUrl = cfg.Section("smscl").Key("url").Value()

	// 孟加拉
	SmsMJLkmiAccKey = cfg.Section("smsmjlkmi").Key("accessKey").Value()
	SmsMJLkmiSecKey = cfg.Section("smsmjlkmi").Key("secretKey").Value()
	SmsMJLkmiURL = cfg.Section("smsmjlkmi").Key("url").Value()
	SmsMJLkmiSender = cfg.Section("smsmjlkmi").Key("sender").Value()
	SmsMJLkmiTemp = cfg.Section("smsmjlkmi").Key("template").Value()
}
