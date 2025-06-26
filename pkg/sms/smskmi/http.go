package smskmi

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
)

// 发送
func SendSmsKMI(url, accKey, secKey, phone, content, sender string) error {
	param, paramErr := newParam(accKey, secKey, phone, content, sender)
	if paramErr != nil {
		return paramErr
	}

	// timeStamp := strconv.FormatInt(time.Now().Unix(), 10)
	// // 生成签名
	// sign := utils.Md5(fmt.Sprintf("%s%s%s", username, password, timeStamp))
	resp, err := doHTTPPost(url, param)
	if err != nil {
		return err
	}

	reslut, resErr := parse(resp)
	if resErr != nil {
		return resErr
	}
	//fmt.Printf("resp %s\n", string(resp))
	if reslut.Code == 200 {
		return nil
	}
	return fmt.Errorf("error code %s", string(resp))
}

func doHTTPPost(targetUrl string, body []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", targetUrl, bytes.NewBuffer(body))
	if err != nil {
		return []byte(""), err
	}
	req.Header.Add("Content-type", "application/json;charset=UTF-8")
	// req.Header.Add("Sign", sign)
	// req.Header.Add("Timestamp", timeStamp)
	// req.Header.Add("Api-Key", apikey)

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
