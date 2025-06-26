package oepay

import (
	"crypto"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"goserver/pkg/glog"
	"testing"

	"gopkg.in/ini.v1"
)

func initConf() {
	cfg, err := ini.Load("../../bin/app.conf")
	if err != nil {
		panic(err)
	}
	channelId := cfg.Section("").Key("oepay.channelId").String()
	keyServerPub := cfg.Section("").Key("oepay.keyServerPub").String()
	keyClientPriv := cfg.Section("").Key("oepay.keyClientPriv").String()
	keyClientPub := cfg.Section("").Key("oepay.keyClientPub").String()
	payUrl := cfg.Section("").Key("oepay.payUrl").String()
	successUrl := cfg.Section("").Key("oepay.successUrl").String()
	failUrl := cfg.Section("").Key("oepay.failUrl").String()

	InitConfig(channelId, keyServerPub, keyClientPriv, keyClientPub, payUrl, successUrl, failUrl)

	// fmt.Println(Config.keyServerPub)
	// fmt.Println(Config.keyClientPriv)
	// fmt.Println(Config.keyClientPub)
}

func getParams() map[string]string {
	params := make(map[string]string)
	params["orderNo"] = "oe10003"
	params["amount"] = "10.00"
	params["firstname"] = "Sophia Wilson"
	params["mobile"] = "7001643406"
	params["email"] = "7001643406@gmail.com"
	params["surl"] = "https://google.com"
	params["furl"] = "https://bing.com"
	params["remark"] = "oe"
	return params
}

func TestOepay(t *testing.T) {
	initConf()
	// params := getParams()
	params := map[string]string{
		"orderNo": "243190146696871939",
	}
	body, _ := json.Marshal(params)
	fmt.Println(string(body))

	sign := PaySign(params, Config.keyClientPriv)
	fmt.Println(sign)
}

func TestOepayCallbackSign(t *testing.T) {
	initConf()
	outSign := "UtIxRSHqADjcB5bgwjQAs06XdPcnRHpfyQHximoHqakn5TO6zQ9sm9nL8oE2Qe2JnzXVTZf8vg5RVIHZtOy/hwi9IR0/Lh5ts1/ZZbk8nBrAkmbK3xxbG5KHVdSsEnz5KGVSMjotyYZdxlFaLXiXw9mCcZWlqqmGNa7vg+GvI1Y="
	content := "amount=100.00&fee=7.00&orderNo=243190146696871939&platformOrderNo=P40409142701PBKJ8SA5&status=1"
	hash := crypto.SHA512
	shaNew := hash.New()
	shaNew.Write([]byte(content))
	hashed := shaNew.Sum(nil)

	// outSign, err := url.QueryUnescape(outSign)
	// if err != nil {
	// 	t.Error(err)
	// 	return
	// }

	outSignBytes, err := base64.StdEncoding.DecodeString(outSign)
	if err != nil {
		glog.Error("verify sign fail, decode sign fail")
		t.Error(err)
		return
	}
	err = rsa.VerifyPKCS1v15(Config.keyServerPub, hash, hashed[:], outSignBytes)
	if err != nil {
		glog.Error("rsa fail", err)
		t.Error(err)
	}
}

func TestOepayCallbackSign2(t *testing.T) {
	initConf()
	// content := "amount=100.00&fee=7.00&orderNo=243190146696871939&platformOrderNo=P40409142701PBKJ8SA5&status=1"
	params := getParams()
	// body, _ := json.Marshal(params)
	// fmt.Println(string(body))

	outSign := PaySign(params, Config.keyClientPriv)
	fmt.Println(outSign)
	content := signContent(params)

	hash := crypto.SHA512
	shaNew := hash.New()
	shaNew.Write([]byte(content))
	hashed := shaNew.Sum(nil)

	outSignBytes, err := base64.StdEncoding.DecodeString(outSign)
	if err != nil {
		glog.Error("verify sign fail, decode sign fail")
		t.Error(err)
		return
	}
	err = rsa.VerifyPKCS1v15(Config.keyClientPub, hash, hashed[:], outSignBytes)
	if err != nil {
		glog.Error("rsa fail", err)
		t.Error(err)
	}
}
