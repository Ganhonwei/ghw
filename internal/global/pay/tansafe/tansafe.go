package tansafe

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"goserver/internal/global/pay/entity"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"io"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var Config *TansafePayConfig

// Submit 下单
func (t *TansafePayConfig) Submit(order *entity.PayOrder) ([]byte, error) {
	body := t.compose(order) //打包post数据
	// strReq := pay.ToQueryString(body)
	strReq, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strReq)
	//fmt.Printf("body %s, order %#v\n", body, order)
	resp, err := doHttpForm(t.PayOrderUrl, strReq)
	//fmt.Printf("resp %s, err %v", string(resp), err)
	if err != nil {
		return nil, err
	}
	glog.Infof("Submit order success, id:%s", order.OrderID)
	return resp, nil
}

// 提现下单
func (t *TansafePayConfig) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	body := t.WithdrawCompose(order) //打包post数据
	// strReq := pay.ToQueryString(body)
	strReq, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strReq)
	resp, err := doHttpForm(t.WithDrawUrl, strReq)
	//fmt.Printf("resp %s, err %v", string(resp), err)
	if err != nil {
		return nil, err
	}
	glog.Infof("WithdrawSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

// 组装参数
func (c *TansafePayConfig) compose(order *entity.PayOrder) map[string]string {
	seconds := order.Ctime / 1000 // 毫秒转换为秒
	tm := time.Unix(seconds, 0)
	dictionary := make(map[string]string)
	dictionary["mchId"] = c.MachId
	dictionary["txChannel"] = c.ChannelId
	dictionary["appId"] = c.appId
	dictionary["timestamp"] = fmt.Sprint(tm)
	dictionary["mchOrderNo"] = order.OrderID
	dictionary["bankCode"] = "UPI"
	dictionary["amount"] = strconv.Itoa(int(order.Amount) / 100)
	dictionary["name"] = order.RealName
	dictionary["phone"] = order.Mobile
	dictionary["email"] = order.Email
	dictionary["productInfo"] = "TP-Rechange"
	dictionary["notifyUrl"] = c.PayNotifyUrl
	// dictionary["returnUrl"] =
	// 签名
	signByte, _ := PaySign(dictionary, c.PrivateKey, crypto.SHA1)
	strSign := base64.StdEncoding.EncodeToString(signByte)
	// dictionary["sign"] = pay.ToUpper(strSign)
	dictionary["sign"] = strSign
	return dictionary
}

// 提现组装参数
func (t *TansafePayConfig) WithdrawCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)
	seconds := order.Ctime / 1000 // 毫秒转换为秒
	tm := time.Unix(seconds, 0)
	dictionary["mchId"] = t.MachId
	dictionary["txChannel"] = t.ChannelId
	dictionary["appId"] = t.appId
	dictionary["timestamp"] = fmt.Sprint(tm)
	dictionary["mchOrderNo"] = order.OrderID
	dictionary["name"] = order.RealName
	dictionary["phone"] = order.Mobile
	dictionary["email"] = order.Email
	dictionary["bankCode"] = "BANK_IN"
	dictionary["account"] = order.BankNumber
	dictionary["amount"] = strconv.Itoa(int(order.Amount) / 100)
	dictionary["notifyUrl"] = t.PayNotifyUrl
	dictionary["ifsc"] = order.IFSC

	// 签名
	signByte, _ := PaySign(dictionary, t.PrivateKey, crypto.SHA1)
	strSign := base64.StdEncoding.EncodeToString(signByte)
	// dictionary["sign"] = pay.ToUpper(strSign)
	dictionary["sign"] = strSign
	return dictionary
}

func PaySign(param map[string]string, privateKeyStr string, sHash crypto.Hash) ([]byte, error) {
	keys := make([]string, 0, len(param))
	for k := range param {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var result strings.Builder
	for _, v := range keys {
		if val, ok := param[v]; ok && val != "" {
			result.WriteString(fmt.Sprintf("%s=%s&", v, param[v]))
		}
	}
	data := []byte(result.String())

	hash := sHash.New()
	hash.Write(data)

	//将私钥解码
	block, _ := pem.Decode([]byte(privateKeyStr))
	//解析私钥
	privateKey, _ := x509.ParsePKCS8PrivateKey(block.Bytes)
	//转换格式  类型断言
	rsaPrivateKey := privateKey.(*rsa.PrivateKey)

	sign, err := rsa.SignPKCS1v15(rand.Reader, rsaPrivateKey, sHash, hash.Sum(nil))
	if err != nil {
		return nil, err
	}
	return sign, nil
}

// // Sign 签名
// func PaySign(data []byte, privateKeyStr string, sHash crypto.Hash) ([]byte, error) {
// 	hash := sHash.New()
// 	hash.Write(data)

// 	//将私钥解码
// 	block, _ := pem.Decode([]byte(privateKeyStr))
// 	//解析私钥
// 	privateKey, _ := x509.ParsePKCS8PrivateKey(block.Bytes)
// 	//转换格式  类型断言
// 	rsaPrivateKey := privateKey.(*rsa.PrivateKey)

// 	sign, err := rsa.SignPKCS1v15(rand.Reader, rsaPrivateKey, sHash, hash.Sum(nil))
// 	if err != nil {
// 		return nil, err
// 	}
// 	return sign, nil
// }

// Verify 验签
// func Verify(data []byte, sign []byte, publicKeyStr string, sHash crypto.Hash) bool {
// 	h := sHash.New()
// 	h.Write(data)

//		//将公钥解码 解析 转换格式
//		block, _ := pem.Decode([]byte(publicKeyStr))
//		publicKey, _ := x509.ParsePKIXPublicKey(block.Bytes)
//		rsaPublicKey := publicKey.(*rsa.PublicKey)
//		return rsa.VerifyPKCS1v15(rsaPublicKey, sHash, h.Sum(nil), sign) == nil
//	}
func (c *TansafePayConfig) Verify(result string, sign []byte, sHash crypto.Hash) bool {
	data := []byte(result)
	h := sHash.New()
	h.Write(data)

	//将公钥解码 解析 转换格式
	block, _ := pem.Decode([]byte(c.ServerPublicKey))
	publicKey, _ := x509.ParsePKIXPublicKey(block.Bytes)
	rsaPublicKey := publicKey.(*rsa.PublicKey)
	return rsa.VerifyPKCS1v15(rsaPublicKey, sHash, h.Sum(nil), sign) == nil
}

func SignContent(param map[string]string) string {
	keys := make([]string, 0, len(param))
	for k := range param {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var result strings.Builder
	for _, v := range keys {
		if val, ok := param[v]; ok {
			if val != "" {
				val = data.ToUrlEncode(val)
				result.WriteString(fmt.Sprintf("%s=%s&", v, val))
			} else {
				result.WriteString(fmt.Sprintf("%s&", v))
			}
		}
	}
	s := result.String()
	if len(s) > 0 {
		s = s[:len(s)-1]
	}
	// s = url.QueryEscape(s)
	return s
}

// application/x-www-form-urlencoded PSOT请求
func doHttpForm(targetUrl string, body []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", targetUrl, bytes.NewReader(body))
	if err != nil {
		glog.Errorf("err:%s", err)
		return []byte(""), err
	}
	req.Header.Add("Content-type", "application/json")

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   0,
			KeepAlive: 0,
		}).Dial,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		TLSHandshakeTimeout: 10 * time.Second,
	}

	client := &http.Client{Transport: transport}

	resp, err := client.Do(req)
	if err != nil {
		glog.Errorf("err:%s", err)
		return []byte(""), err
	}

	defer resp.Body.Close()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("err:%s", err)
		return []byte(""), err
	}

	return respData, nil
}

func (c *TansafePayConfig) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	result := new(OrderRespond)
	jsoniter.Unmarshal(body, result)
	if result.Status != 200 { // 没成功TODO
		return "", fmt.Errorf("code:%d, msg:%s", result.Status, result.Msg)
	}
	return result.Data.Link, nil
}

func (c *TansafePayConfig) ParsePayResult(body []byte) (map[string]string, error) {
	tradeResult := OrderCallback{}
	err := jsoniter.Unmarshal(body, &tradeResult)
	if err != nil {
		return nil, err
	}
	reslut := make(map[string]string)
	reslut["status"] = fmt.Sprint(tradeResult.Status)
	reslut["memberId"] = tradeResult.MemberId
	reslut["mchOrderNo"] = tradeResult.MchOrderNo
	reslut["platOrderNo"] = tradeResult.PlatOrderNo
	reslut["orderStatus"] = tradeResult.OrderStatus
	reslut["amount"] = fmt.Sprint(tradeResult.Amount)
	reslut["fee"] = fmt.Sprint(tradeResult.Fee)
	reslut["msg"] = tradeResult.Msg
	reslut["timeEnd"] = fmt.Sprint(tradeResult.TimeEnd)
	reslut["sign"] = tradeResult.Sign
	return reslut, nil
}

func (c *TansafePayConfig) ParseWithdrawResult(body []byte) (map[string]string, error) {
	tradeResult := WithdrawCallback{}
	err := jsoniter.Unmarshal(body, &tradeResult)
	if err != nil {
		return nil, err
	}
	reslut := make(map[string]string)
	reslut["status"] = fmt.Sprint(tradeResult.Status)
	reslut["memberId"] = tradeResult.MemberId
	reslut["mchOrderNo"] = tradeResult.MchOrderNo
	reslut["platOrderNo"] = tradeResult.PlatOrderNo
	reslut["orderStatus"] = tradeResult.OrderStatus
	reslut["amount"] = fmt.Sprint(tradeResult.Amount)
	reslut["fee"] = fmt.Sprint(tradeResult.Fee)
	reslut["msg"] = tradeResult.Msg
	reslut["timeEnd"] = fmt.Sprint(tradeResult.TimeEnd)
	reslut["sign"] = tradeResult.Sign
	return reslut, nil
}

func (c *TansafePayConfig) WithdrawSubmitResponse(body []byte) (bool, string) {
	result := new(WithdrawRespond)
	jsoniter.Unmarshal(body, result)
	if result.Status != 200 { // 没成功TODO
		glog.Errorf("[xfpay] withdraw fail submit, code:%d, msg:%s", result.Status, result.Msg)
		return false, fmt.Sprint(result.Msg)
	}
	return true, fmt.Sprint(result.Msg)
}
