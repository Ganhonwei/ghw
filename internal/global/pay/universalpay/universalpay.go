package universalpay

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"goserver/internal/global/pay/entity"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"io"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var Config = &UniversalpayConfig{}

const (
	pem_begin = "-----BEGIN PRIVATE KEY-----\n"
	pem_end   = "\n-----END PRIVATE KEY-----"
	pub_begin = "-----BEGIN PUBLIC KEY-----\n"
	pub_end   = "\n-----END PUBLIC KEY-----"
)

func (p *UniversalpayConfig) Submit(order *entity.PayOrder) ([]byte, error) {
	body := p.compose(order) //打包post数据
	strReq, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strReq)
	//fmt.Printf("body %s, order %#v\n", body, order)
	// 签名
	sign := rsaSign(SignContent(body), p.PrirsaClient, crypto.SHA512)
	resp, err := doHttpJSON("https://gamerplayers.com/gold-pay/portal/payment", strReq, "POST", sign)
	//fmt.Printf("resp %s, err %v", string(resp), err)
	if err != nil {
		return nil, err
	}
	glog.Infof("Submit order success, id:%s", order.OrderID)
	return resp, nil
}

func (p *UniversalpayConfig) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	type R struct {
		Code string          `json:"code"`
		Msg  string          `json:"msg"`
		Data PayOrderRespond `json:"data"`
	}

	result := new(R)
	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		return "", err
	}

	if result.Code != "0000" {
		return "", fmt.Errorf("[universalpay]submit pay order error: %s", result.Msg)
	}
	return result.Data.LinkUrl, nil
}

func (p *UniversalpayConfig) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	body := p.WithdrawCompose(order) //打包post数据
	strReq, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strReq)
	//fmt.Printf("body %s, order %#v\n", body, order)
	// 签名
	sign := rsaSign(SignContent(body), p.PrirsaClient, crypto.SHA512)

	order.RequestParam = string(strReq)
	resp, err := doHttpJSON("https://gamerplayers.com/gold-pay/portal/payout", strReq, "POST", sign)
	//fmt.Printf("resp %s, err %v", string(resp), err)
	if err != nil {
		return nil, err
	}
	glog.Infof("WithdrawSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

func (p *UniversalpayConfig) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	type R struct {
		Code string               `json:"code"`
		Msg  string               `json:"msg"`
		Data WithdrawOrderRespond `json:"data"`
	}
	result := new(R)
	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		order.OrderStatus = data.DeliveryFailTransfer
		return false, err.Error()
	}

	if result.Code != "0000" || result.Data.Status != 1 {
		order.OrderStatus = data.OrderFail
		return false, fmt.Sprintf("code:%s,status:%d,msg:%s", result.Code, result.Data.Status, result.Msg)
	}
	order.OrderStatus = data.OrderSuccess
	return true, result.Code
}

func (p *UniversalpayConfig) InspectSubmit(order *entity.PayOrder) ([]byte, error) {
	body := map[string]string{
		"orderNo": order.OrderID,
	}
	strReq, _ := jsoniter.Marshal(body)
	sign := rsaSign(SignContent(body), p.PrirsaClient, crypto.SHA512)
	resp, err := doHttpJSON("https://gamerplayers.com/gold-pay/portal/queryPayOrderStatus", strReq, "POST", sign)
	if err != nil {
		return nil, err
	}
	glog.Infof("InspectSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

func (p *UniversalpayConfig) InspectResponse(body []byte, order *entity.PayOrder) error {
	type R struct {
		Code string            `json:"code"`
		Msg  string            `json:"msg"`
		Data PayInspectRespond `json:"data"`
	}

	result := new(R)
	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		return err
	}

	if result.Code != "0000" { // 没成功
		return fmt.Errorf("code:%s, message:%s", result.Code, result.Msg)
	}
	if result.Data.Status != 1 {
		return fmt.Errorf("state:%d, message:%s", result.Data.Status, result.Msg)
	}
	order.OutTradeNo = result.Data.PlatformOrderNo
	return nil
}

func (p *UniversalpayConfig) InspectWithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	body := map[string]string{
		"orderNo": order.OrderID,
	}
	strReq, _ := jsoniter.Marshal(body)
	sign := rsaSign(SignContent(body), p.PrirsaClient, crypto.SHA512)
	resp, err := doHttpJSON("https://gamerplayers.com/gold-pay/portal/queryTransferOrderStatus", strReq, "POST", sign)
	if err != nil {
		return nil, err
	}
	glog.Infof("InspectWithdrawSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

func (p *UniversalpayConfig) InspectWdResponse(body []byte, order *entity.WithdrawOrder) int {
	type R struct {
		Code string                 `json:"code"`
		Msg  string                 `json:"msg"`
		Data WithdrawInspectRespond `json:"data"`
	}

	result := new(R)
	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		return 3
	}
	if result.Code != "0000" { // 没成功
		return 3
	}
	if result.Data.Status == 1 {
		return 2
	}
	return 3
}

func (p *UniversalpayConfig) BalanceQuery() int64 {
	resp, err := doHttpJSON(fmt.Sprintf("https://gamerplayers.com/gold-pay/portal/queryBalance?serviceCode=%s", p.Appid), []byte{}, "GET", "")
	if err != nil {
		glog.Info("[universalpay] BalanceQuery order fail,err:", err)
		return 0
	}
	fmt.Println(string(resp))

	type R struct {
		Code string         `json:"code"`
		Msg  string         `json:"msg"`
		Data BalanceRespond `json:"data"`
	}
	rsp := new(R)
	err = json.Unmarshal(resp, &rsp)
	if err != nil {
		glog.Info("[universalpay] BalanceQuery order fail,err:", err)
		return 0
	}

	if rsp.Code != "0000" {
		return 0
	}
	glog.Infof("[universalpay] BalanceQuery success")
	return int64(rsp.Data.TotalAmount)
}

// application/json PSOT请求
func doHttpJSON(targetUrl string, body []byte, method, sign string) ([]byte, error) {
	req, err := http.NewRequest(method, targetUrl, bytes.NewReader(body))
	if err != nil {
		glog.Errorf("err:%s", err)
		return []byte(""), err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-SIGN", sign)
	req.Header.Add("X-SERVICE-CODE", Config.Appid)

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

// 组装参数
func (c *UniversalpayConfig) compose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["orderNo"] = order.OrderID
	dictionary["amount"] = fmt.Sprintf("%.2f", float32(order.Amount)/100)
	dictionary["firstname"] = order.RealName
	dictionary["mobile"] = order.Mobile
	dictionary["email"] = order.Email
	dictionary["surl"] = "https://google.com"
	dictionary["furl"] = "https://bing.com"
	dictionary["remark"] = fmt.Sprintf("Cash %.2f", float64(order.Amount)/100)
	dictionary["callbackUrl"] = c.PayNotifyUrl

	return dictionary
}

// 提现组装参数
func (t *UniversalpayConfig) WithdrawCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)
	dictionary["orderNo"] = order.OrderID
	dictionary["amount"] = fmt.Sprintf("%.2f", float32(order.Amount)/100)
	dictionary["bankAccount"] = order.BankNumber
	dictionary["ifsc"] = order.IFSC
	dictionary["beneficiaryName"] = order.RealName
	dictionary["beneficiaryMobile"] = order.Mobile
	dictionary["beneficiaryEmail"] = order.Email
	dictionary["paymentType"] = "0"
	dictionary["remark"] = fmt.Sprintf("withdraw %v", dictionary["amount"])
	dictionary["callbackUrl"] = t.WithdrawNotifyUrl

	return dictionary
}

func SignContent(param map[string]string) string {
	keys := make([]string, 0, len(param))
	for k := range param {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var fields []string
	for _, v := range keys {
		if val, ok := param[v]; ok {
			if val != "" {
				// val = pay.ToUrlEncode(val)
				fields = append(fields, fmt.Sprintf("%s=%s", v, val))
			}
		}
	}
	s := strings.Join(fields, "&")
	fmt.Println("signParam:", s)
	return s
}

func rsaSign(content, key string, hash crypto.Hash) string {
	shaNew := hash.New()
	shaNew.Write([]byte(content))
	hashed := shaNew.Sum(nil)
	prikey, err := parsePrivateKey(key)
	if err != nil {
		glog.Errorf("get prikey fail,err:%s", err)
		return ""
	}
	sign, err1 := rsa.SignPKCS1v15(rand.Reader, prikey, hash, hashed)
	if err1 != nil {
		glog.Error("rsa fail,err:%s", err1)
		return ""
	}
	s := base64.StdEncoding.EncodeToString(sign)
	return s
}

// 生成私钥对象
func parsePrivateKey(key string) (*rsa.PrivateKey, error) {
	if !strings.HasPrefix(key, pem_begin) {
		key = pem_begin + key
	}
	if !strings.HasSuffix(key, pem_end) {
		key = key + pem_end
	}
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return nil, errors.New("key is error")
	}
	// 解析DER编码的私钥，生成私钥对象
	prikey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return prikey.(*rsa.PrivateKey), nil
}

// 生成公钥对象
func parsePublickKey(key string) (*rsa.PublicKey, error) {
	if !strings.HasPrefix(key, pub_begin) {
		key = pub_begin + key
	}
	if !strings.HasSuffix(key, pub_end) {
		key = key + pub_end
	}
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return nil, errors.New("key is error")
	}
	// 解析DER编码的公钥，生成公钥对象
	pubkey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return pubkey.(*rsa.PublicKey), nil
}

// 验签
func (c *UniversalpayConfig) VerifySing(sign, content string, hash crypto.Hash) bool {
	shaNew := hash.New()
	shaNew.Write([]byte(content))
	hashed := shaNew.Sum(nil)
	pubkey, err := parsePublickKey(c.PubrsaServer)
	if err != nil {
		glog.Error("get pubkey fail")
		return false
	}
	s, err2 := base64.StdEncoding.DecodeString(sign)
	if err2 != nil {
		glog.Error("verify sign fail, decode sign fail")
		return false
	}
	err1 := rsa.VerifyPKCS1v15(pubkey, hash, hashed[:], s)
	if err1 != nil {
		glog.Error("rsa fail")
		return false
	}
	return true
}

func (c *UniversalpayConfig) ParsePayResult(body []byte) (map[string]string, error) {
	tradeResult := PayRechargeCallback{}
	err := jsoniter.Unmarshal(body, &tradeResult)
	if err != nil {
		return nil, err
	}
	reslut := make(map[string]string)
	reslut["platformOrderNo"] = tradeResult.PlatformOrderNo
	reslut["orderNo"] = tradeResult.OrderNo
	reslut["status"] = tradeResult.Status
	reslut["amount"] = tradeResult.Amount
	reslut["fee"] = tradeResult.Fee
	reslut["linkUrl"] = tradeResult.LinkUrl
	reslut["utrNo"] = tradeResult.UtrNo
	return reslut, nil
}

func (c *UniversalpayConfig) ParseWithdrawResult(body []byte) (map[string]string, error) {
	tradeResult := WithdrawCallback{}
	err := jsoniter.Unmarshal(body, &tradeResult)
	if err != nil {
		return nil, err
	}
	reslut := make(map[string]string)
	reslut["platformOrderNo"] = tradeResult.PlatformOrderNo
	reslut["orderNo"] = tradeResult.OrderNo
	reslut["status"] = tradeResult.Status
	reslut["amount"] = tradeResult.Amount
	reslut["fee"] = tradeResult.Fee
	reslut["utrNo"] = tradeResult.UtrNo
	reslut["errorMsg"] = tradeResult.ErrorMsg
	return reslut, nil
}
