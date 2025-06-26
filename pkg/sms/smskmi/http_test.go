package smskmi

import (
	"fmt"
	"testing"
)

func TestSend(t *testing.T) {
	code := "1111"
	content := fmt.Sprintf("[CasinoWin]The verification code is:%s.Please do not share it with others.Team QUICKCELLAR", code)
	phone := "918279428074"
	ACCKey := "e9fce16d048b4225bdd676b649309f59"
	SECKey := "e391cecb906941da983ab2595be59394"
	url := "http://api.kmicloud.com/sms/send/v1/otp"
	err := SendSmsKMI(url, ACCKey, SECKey, phone, content, "QUICCE")
	if err != nil {
		t.Logf("err %v", err)
		t.Logf("send sms failed phone %s, code %s", phone, code)
	} else {
		t.Logf("send sms successful phone %s, code %s", phone, code)
	}
}
