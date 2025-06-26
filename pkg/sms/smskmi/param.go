package smskmi

import (
	"encoding/json"
	"net/url"
	"strings"
)

func urlencoder(str string) string {
	return url.QueryEscape(str)
}

func newParam(accKey, secKey, phone, msg, sender string) ([]byte, error) {
	params := make(map[string]string)
	params["accessKey"] = accKey
	params["secretKey"] = secKey
	params["message"] = msg
	params["to"] = phone
	if sender != "" {
		params["from"] = sender
	}
	bytesData, err := json.Marshal(params)
	return bytesData, err
}

/*
	{

"status" : "0", //状态码，0表示成功
"reason" : 失败原因说明
"success" : 提交成功的号码个数
"fail" : 提交失败的号码个数

	"array" : [{
		"msgId" : "17041010383624511", //消息Id
		"number" : "20170410103836" //提交号码
	}]提交成功的json集合

}
*/
type SmsReponse struct {
	Success   bool    `json:"success"`
	Message   string  `json:"message"`
	Code      int     `json:"code"`
	Timestamp int64   `json:"timestamp"`
	Result    Message `json:"result"`
}

type Message struct {
	To         string `json:"to"`
	SmsId      string `json:"smsId"`
	SendResult string `json:"sendResult"`
}

type SmsCallback struct {
	SmsId      string `json:"smsId"`
	SendResult string `json:"sendResult"`
	SmsCount   string `json:"smsCount"`
	Dr         string `json:"dr"`
	To         string `json:"to"`
	SendTime   string `json:"sendTime"`
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

// parse the reponse message
func ParseCallback(resp []byte) (SmsCallback, error) {
	result := SmsCallback{}

	param := make(map[string]any)
	bodys := strings.Split(string(resp), "&")

	for _, v := range bodys {
		val := strings.Split(v, "=")
		param[val[0]] = val[1]
	}

	b, err1 := json.Marshal(param)
	if err1 != nil {
		return SmsCallback{}, err1
	}

	err := json.Unmarshal(b, &result)
	if err != nil {
		return result, err
	}
	return result, nil
}
