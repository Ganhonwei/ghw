package smscl

import (
	"encoding/json"
	"goserver/pkg/utils"
	"strings"

	"github.com/gogo/protobuf/sortkeys"
)

func newParam(appid, pwd, phone, msg, sender string) map[string]string {
	params := make(map[string]string)
	params["account"] = appid
	params["mobile"] = phone
	params["password"] = pwd
	// params["content"] = url.QueryEscape(msg)
	params["msg"] = msg
	if sender != "" {
		params["senderId"] = sender
	}
	return params
}

/*
	{

"status" : "0", //状态码，0表示成功
"Message" : 状态码说明
"Data" : 消息体
}
*/
type SmsReponse struct {
	Code    string    `json:"code"`
	Message string    `json:"message"`
	Data    []Message `json:"data"`
}

type Message struct {
	MessageId string `json:"messageId"`
}

type SmsCallback struct {
	Receiver    string `json:"receiver"`
	Pswd        string `json:"pswd"`
	Msgid       string `json:"msgid"`
	BatchSeq    string `json:"batchSeq"`
	ReportTime  string `json:"reportTime"`
	NotifyTime  string `json:"notifyTime"`
	RequestTime string `json:"requestTime"`
	Mobile      string `json:"mobile"`
	Status      string `json:"status"`
	SmsNum      string `json:"smsNum"`
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

// 内容签名
func signContent(param map[string]string, passward string) string {
	content := ""
	keys := make([]string, 0)
	for k := range param {
		keys = append(keys, k)
	}
	sortkeys.Strings(keys)
	for _, v := range keys {
		val := param[v]
		if val != "" {
			content += v + val
		}
	}
	content += passward
	content = utils.Md5(content)
	return strings.ToLower(content)
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
