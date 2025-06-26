package smscl

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"goserver/pkg/utils"
	"io/ioutil"
	"net/http"
)

// 发送
func SendSmsCL(url, apiAcc, password, phone, content, sender string) error {
	param := newParam(apiAcc, password, phone, content, "")

	sign := signContent(param, password) // 签名
	timeStamp := utils.BsonNow().UnixMilli()

	body, err := json.Marshal(param)
	if err != nil {
		return err
	}

	resp, err := doHTTPPost(url, sign, timeStamp, body)
	if err != nil {
		return err
	}

	reslut, resErr := parse(resp)
	if resErr != nil {
		return resErr
	}
	//fmt.Printf("resp %s\n", string(resp))
	if reslut.Code == "0" {
		return nil
	}
	return fmt.Errorf("error code %s", string(resp))
}

func doHTTPPost(targetUrl, sign string, timeStamp int64, body []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", targetUrl, bytes.NewBuffer(body))
	if err != nil {
		return []byte(""), err
	}
	req.Header.Add("Content-type", "application/json;charset=UTF-8")
	// req.Header.Add("sign", sign)
	// req.Header.Add("nonce", fmt.Sprintf("%d", timeStamp))

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		return []byte(""), err
	}

	defer resp.Body.Close()
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte(""), err
	}

	return respData, nil
}
