package smsxxy

import (
	"encoding/json"
	"fmt"
	"net/url"
)

func newParam(appkey, appcode, appsecret, phone, msg string) string {
	return fmt.Sprintf("?appkey=%s&appsecret=%s&appcode=%s&phone=%s&msg=%s", appkey, appsecret, appcode, phone, url.QueryEscape(msg))
	// params := make(map[string]string)
	// params["appkey"] = appkey
	// params["appcode"] = appcode
	// params["appsecret"] = appsecret
	// params["msg"] = msg
	// params["phone"] = phone
	// bytesData, err := json.Marshal(params)
	// return bytesData, err
}

/*
	{

"status" : "00000", //状态码，0表示成功
"desc" : 说明
"uid" : 标识
}
*/
type SmsReponse struct {
	Code   string    `json:"code"`
	Desc   string    `json:"desc"`
	Uid    string    `json:"uid"`
	Result []Message `json:"result"`
}

type Message struct {
	Status string `json:"status"`
	Desc   string `json:"desc"`
	Phone  string `json:"phone"`
}

type SmsCallback struct {
	AppKey     string `json:"appkey"`
	Phone      string `json:"phone"`
	Status     string `json:"status"`
	Desc       string `json:"desc"`
	Uid        string `json:"uid"`
	ReportTime string `json:"report_time"`
}

// parse the reponse message
func parse(resp []byte) (*SmsReponse, error) {
	result := new(SmsReponse)
	err := json.Unmarshal(resp, result)
	if err != nil {
		return result, err
	}
	return result, nil
}

// parse the callback message
func ParseCallback(resp []byte) ([]SmsCallback, error) {
	result := make([]SmsCallback, 0)
	err := json.Unmarshal(resp, &result)
	if err != nil {
		return result, err
	}
	return result, nil
}
