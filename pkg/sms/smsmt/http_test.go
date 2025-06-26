package smsmt

import (
	"fmt"
	"testing"
)

func TestSend(t *testing.T) {
	code := "1111"
	content := fmt.Sprintf("%s is your OTP to register with BSG JEWELLERS", code)
	phone := "918077425674"
	appKey := "acQ3fF"
	appsecret := "PaaQr1"
	appcode := "1000"
	url := "http://43.133.60.95:9090/sms/batch/v2"
	err := SendSmsMT(url, appKey, appsecret, appcode, phone, content)
	if err != nil {
		t.Logf("err %v", err)
		t.Logf("send sms failed phone %s, code %s", phone, code)
	} else {
		t.Logf("send sms successful phone %s, code %s", phone, code)
	}
}
