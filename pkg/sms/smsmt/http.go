package smsmt

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// 发送
func SendSmsMT(url, appKey, appSecret, appCode, phone, content string) error {
	param := newParam(appKey, appCode, appSecret, phone, content)

	resp, err := doHTTPGet(url, param)
	if err != nil {
		return err
	}

	reslut, resErr := parse(resp)
	if resErr != nil {
		return resErr
	}
	//fmt.Printf("resp %s\n", string(resp))
	if reslut.Code == "00000" {
		return nil
	}
	return fmt.Errorf("error code %s", string(resp))
}

func doHTTPGet(targetUrl string, body string) ([]byte, error) {
	resp, err := http.Get(targetUrl + body)
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
