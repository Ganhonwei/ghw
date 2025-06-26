package login

import (
	"fmt"
	"time"


	"goserver/pkg/sms/smskmi"
	"goserver/pkg/sms/smsmt"
	"goserver/pkg/sms/smsxxy"
	"goserver/pkg/sms/smscl"
	"goserver/pkg/glog"
)

// SendSms 发送验证码
// func SendSms(phone, code, username, password string) bool {
// 	return sendSms2(phone, code, username, password, 0)
// }

// func sendSms2(phone, code, username, password string, n int) bool {
// 	content := fmt.Sprintf("【欢乐】你的验证码%s，请勿泄露。", code)
// 	err := smsbao.SendSmsbao(phone, content, username, password)
// 	if err == nil {
// 		glog.Debugf("send sms successfully phone %s, code %s", phone, code)
// 		return true
// 	}
// 	glog.Errorf("send sms failed phone %s, code %s, err %v", phone, code, err)
// 	if n >= 3 { //失败尝试次数3次
// 		return false
// 	}
// 	<-time.After(3 * time.Second)
// 	return sendSms2(phone, code, username, password, n+1)
// }

// SendSms253 发送验证码
// func SendSms253(targetURL, account, password, phone, code string) bool {
// 	return sendSms3(targetURL, account, password, phone, code, 0)
// }

// func sendSms3(targetURL, account, password, phone, code string, n int) bool {
// 	msg := fmt.Sprintf("【欢斗娱乐】你的验证码%s", code) //平台报备签名才能成功,否则不添加签名
// 	err := sms253.SendSms(targetURL, account, password, phone, msg)
// 	if err == nil {
// 		glog.Debugf("send sms successfully phone %s, code %s", phone, code)
// 		return true
// 	}
// 	glog.Errorf("send sms failed phone %s, code %s, err %v", phone, code, err)
// 	if n >= 3 { //失败尝试次数3次
// 		return false
// 	}
// 	<-time.After(3 * time.Second)
// 	return sendSms3(targetURL, account, password, phone, code, n+1)
// }

// buka短信
// func SendSmsBK(url, appid, phone, code, username, password string) bool {
// 	return sendSmsBK(url, appid, phone, code, username, password, 0)
// }

// func sendSmsBK(url, appid, phone, code, username, password string, n int) bool {
// 	// content := fmt.Sprintf("【欢乐】你的验证码%s，请勿泄露。", code)
// 	// err := smsbao.SendSmsbao(phone, content, username, password)
// 	err := smsbk.SendSmsbk(url, appid, phone, code, username, password, "")
// 	if err == nil {
// 		glog.Debugf("send sms successfully phone %s, code %s", phone, code)
// 		return true
// 	}
// 	glog.Errorf("send sms failed phone %s, code %s, err %v", phone, code, err)
// 	if n >= 3 { //失败尝试次数3次
// 		return false
// 	}
// 	<-time.After(3 * time.Second)
// 	return sendSmsBK(url, appid, phone, code, username, password, n+1)
// }

// kmi短信
func SendSmsKMI(url, accKey, secKey, sender, temp, phone, code string) bool {
	return sendSmsKMI(url, accKey, secKey, sender, temp, phone, code, 0)
}

func sendSmsKMI(url, accKey, secKey, sender, temp, phone, code string, n int) bool {
	content := fmt.Sprintf(temp, code)
	// err := smsbao.SendSmsbao(phone, content, username, password)
	err := smskmi.SendSmsKMI(url, accKey, secKey, phone, content, sender)
	if err == nil {
		glog.Debugf("send sms successfully phone %s, code %s", phone, code)
		return true
	}
	glog.Errorf("send sms failed phone %s, code %s, err %v", phone, code, err)
	if n >= 1 { //失败尝试次数1次
		return false
	}
	<-time.After(3 * time.Second)
	return sendSmsKMI(url, accKey, secKey, sender, temp, phone, code, n+1)
}

// xxy短信
func SendSmsXXY(url, appKey, appSecret, appCode, temp, phone, code string) bool {
	return sendSmsXXY(url, appKey, appSecret, appCode, temp, phone, code, 0)
}

func sendSmsXXY(url, appKey, appSecret, appCode, temp, phone, code string, n int) bool {
	content := fmt.Sprintf(temp, code)
	// err := smsbao.SendSmsbao(phone, content, username, password)
	err := smsxxy.SendSmsXXY(url, appKey, appSecret, appCode, phone, content)
	if err == nil {
		glog.Debugf("send sms successfully phone %s, code %s", phone, code)
		return true
	}
	glog.Errorf("send sms failed phone %s, code %s, err %v", phone, code, err)
	if n >= 1 { //失败尝试次数1次
		return false
	}
	<-time.After(2 * time.Second)
	return sendSmsXXY(url, appKey, appSecret, appCode, temp, phone, code, n+1)
}

// mt短信
func SendSmsMT(url, appKey, appSecret, appCode, temp, phone, code string) bool {
	return sendSmsMT(url, appKey, appSecret, appCode, temp, phone, code, 0)
}

func sendSmsMT(url, appKey, appSecret, appCode, temp, phone, code string, n int) bool {
	content := fmt.Sprintf(temp, code)
	// err := smsbao.SendSmsbao(phone, content, username, password)
	err := smsmt.SendSmsMT(url, appKey, appSecret, appCode, phone, content)
	if err == nil {
		glog.Debugf("send sms successfully phone %s, code %s", phone, code)
		return true
	}
	glog.Errorf("send sms failed phone %s, code %s, err %v", phone, code, err)
	if n >= 1 { //失败尝试次数1次
		return false
	}
	<-time.After(2 * time.Second)
	return sendSmsMT(url, appKey, appSecret, appCode, temp, phone, code, n+1)
}

// 创蓝短信
func SendSmsCL(url, apiAcc, pwd, phone, code string) bool {
	return sendSmsCL(url, apiAcc, pwd, phone, code, 0)
}

func sendSmsCL(url, apiAcc, pwd, phone, code string, n int) bool {
	// content := fmt.Sprintf("[CasinoWin]The verification code is:%s.Please do not share it with others.Team QUICKCELLAR", code)
	// err := smsbao.SendSmsbao(phone, content, username, password)
	err := smscl.SendSmsCL(url, apiAcc, pwd, phone, code, "")
	if err == nil {
		glog.Debugf("send sms successfully phone %s, code %s", phone, code)
		return true
	}
	glog.Errorf("send sms failed phone %s, code %s, err %v", phone, code, err)
	if n >= 1 { //失败尝试次数1次
		return false
	}
	<-time.After(2 * time.Second)
	return sendSmsCL(url, apiAcc, pwd, phone, code, n+1)
}
