package smscl

import (
	"testing"
)

func TestSend(t *testing.T) {
	code := "1111"
	// content := fmt.Sprintf("[CWin]The verification code is:%s.Please do not share it with others.Team QUICKCELLAR", code)
	phone := "918077425674"
	apiAcc := "I0108126"
	pwd := "8DQM062Q71082d"
	url := "http://intapi.253.com/send/json"
	err := SendSmsCL(url, apiAcc, pwd, phone, code, "")
	if err != nil {
		t.Logf("err %v", err)
		t.Logf("send sms failed phone %s, code %s", phone, code)
	} else {
		t.Logf("send sms successful phone %s, code %s", phone, code)
	}
}
