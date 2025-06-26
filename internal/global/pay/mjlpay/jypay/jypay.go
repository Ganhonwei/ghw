package jypay

import (
	"bytes"
	"crypto"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"goserver/internal/global/pay/entity"
	"goserver/internal/global/pay/service"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"io"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var Config *JYPayConfig

// Submit 下单
func (t *JYPayConfig) Submit(order *entity.PayOrder) ([]byte, error) {
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
func (t *JYPayConfig) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	if !checkWallet(order.Bank) {
		return nil, errors.New("wallet format error")
	}

	body := t.WithdrawCompose(order) //打包post数据
	// strReq := pay.ToQueryString(body)
	strReq, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strReq)
	resp, err := doHttpForm(t.WithDrawUrl, strReq)
	//fmt.Printf("resp %s, err %v", string(resp), err)
	if err != nil {
		return nil, err
	}
	glog.Infof("WithdrawSubmit order success, id:%s, resp:%s", order.OrderID, string(resp))
	return resp, nil
}

// 组装参数
func (c *JYPayConfig) compose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)

	// wallets := []string{"baksh", "nagad"}
	// wallet, _ := utils.ChoiceString(wallets)
	// if wallet == "" {
	// 	wallet = "baksh"
	// }
	wallet := func(app int32) string {
		switch app {
		case 101:
			return "ngand"
		case 102:
			return "baksh"
		}
		return "baksh"
	}

	phone := service.GenerateMJLPhone()
	dictionary["bankCode"] = wallet(order.PayApp)
	dictionary["merNo"] = c.Merchant
	dictionary["merOrderNo"] = order.OrderID
	dictionary["name"] = order.RealName
	dictionary["phone"] = phone
	dictionary["email"] = fmt.Sprintf("%s@gmail.com", phone)
	dictionary["orderAmount"] = fmt.Sprintf("%.2f", float32(order.Amount)/100)
	dictionary["notifyUrl"] = Config.PayNotifyUrl // 充值后返回地址
	dictionary["currency"] = "BDT"
	dictionary["busiCode"] = "124001"
	dictionary["pageUrl"] = "https://www.google.com/"
	dictionary["timestamp"] = utils.String(utils.BsonNow().UnixMilli())

	// 签名
	strSign := HmacRSASign(SignContent(dictionary), c.HamcRSA, crypto.SHA256)
	dictionary["sign"] = strSign
	return dictionary
}

// 提现组装参数
func (c *JYPayConfig) WithdrawCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)

	phone := service.GenerateMJLPhone()
	dictionary["accName"] = order.RealName
	dictionary["accNo"] = order.BankNumber
	dictionary["bankCode"] = getWallet(order.Bank)
	dictionary["busiCode"] = "223001"
	dictionary["currency"] = "BDT"
	dictionary["email"] = fmt.Sprintf("%s@gmail.com", phone)
	dictionary["merNo"] = c.Merchant
	dictionary["merOrderNo"] = order.OrderID
	dictionary["notifyUrl"] = Config.WithDrawNotifyUrl // 提现后回调地址
	dictionary["orderAmount"] = fmt.Sprintf("%.2f", float32(order.Amount)/100)
	dictionary["phone"] = phone
	dictionary["timestamp"] = fmt.Sprintf("%d", utils.BsonNow().UnixMilli())

	// 签名
	strSign := HmacRSASign(SignContent(dictionary), c.HamcRSA, crypto.SHA256)
	// 再加密
	dictionary["sign"] = rsaSign(strSign, c.RSAPriRSA, crypto.SHA256)
	return dictionary
}

func rsaSign(content, key string, hash crypto.Hash) string {
	// shaNew := hash.New()
	// shaNew.Write([]byte(content))
	// hashed := shaNew.Sum(nil)
	prikey, err := service.ParsePrivateKey(key)
	if err != nil {
		glog.Errorf("get prikey fail,err:%s", err)
		return ""
	}
	sign, err1 := rsa.SignPKCS1v15(nil, prikey, crypto.Hash(0), []byte(content))
	if err1 != nil {
		glog.Error("rsa fail,err:%s", err1)
		return ""
	}
	s := base64.StdEncoding.EncodeToString(sign)
	return s
}

func HmacRSASign(content, key string, hash crypto.Hash) string {
	keyBytes := []byte(key)
	shaNew := hmac.New(sha256.New, keyBytes)
	shaNew.Write([]byte(content))
	hashed := shaNew.Sum(nil)
	return hex.EncodeToString(hashed)
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
				// val = pay.ToUrlEncode(val)
				result.WriteString(fmt.Sprintf("%s=%s&", v, val))
			}
		}
	}
	s := result.String()
	if len(s) > 0 {
		s = s[:len(s)-1]
	}
	return s
}

// application/json PSOT请求
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

// 验签
func (c *JYPayConfig) VerifySing(sign, content string, hash crypto.Hash) bool {
	shaNew := hash.New()
	shaNew.Write([]byte(content))
	hashed := shaNew.Sum(nil)
	pubkey, err := service.ParsePublickKey(c.RSAPubRSA)
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

// 验签
func (c *JYPayConfig) VerifyHmacSing(sign, content string, hash crypto.Hash) bool {
	outSign := HmacRSASign(content, c.HamcRSA, hash)
	return outSign == sign
}

// =========================实现接口方法=========================

func (c *JYPayConfig) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	result := new(JyOrderRespond)
	param := make(map[string]string)
	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		return "", err
	}
	param["merNo"] = result.Data.MerNo
	param["merOrderNo"] = result.Data.MerOrderNo
	param["orderNo"] = result.Data.OrderNo
	param["subCode"] = utils.String(result.Data.SubCode)
	param["subMsg"] = result.Data.SubMsg
	param["status"] = utils.String(result.Data.Status)
	param["orderData"] = result.Data.OrderData
	param["common"] = result.Data.Common
	if result.Code != 200 { // 没成功TODO
		return "", fmt.Errorf("code:%d,message:%s, OrderNo:%s", result.Code, result.Message, order.OrderID)
	}
	if !c.VerifyHmacSing(result.Data.Sign, SignContent(param), crypto.SHA256) {
		return "", errors.New("verify sign fail")
	}

	if result.Data.OrderData == "" {
		return "", errors.New("pay link is null")
	}

	order.OutTradeNo = result.Data.OrderNo
	return result.Data.OrderData, nil
	// return "", nil
}

func (c *JYPayConfig) ParsePayResult(body []byte) (map[string]string, error) {
	rsp := make(map[string]any)
	reslut := make(map[string]string)
	err := jsoniter.Unmarshal(body, &rsp)
	if err != nil {
		return nil, err
	}
	for k, v := range rsp {
		if f, ok := v.(float64); ok {
			reslut[k] = strconv.FormatFloat(f, 'f', -1, 64)
			continue
		}
		reslut[k] = utils.String(v)
	}
	return reslut, nil
}

func (c *JYPayConfig) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	result := new(JyOrderRespond)
	param := make(map[string]string)
	jsoniter.Unmarshal(body, result)
	param["merNo"] = result.Data.MerNo
	param["merOrderNo"] = result.Data.MerOrderNo
	param["orderNo"] = result.Data.OrderNo
	param["subCode"] = utils.String(result.Data.SubCode)
	param["subMsg"] = result.Data.SubMsg
	param["status"] = utils.String(result.Data.Status)
	if result.Code != 200 { // 没成功TODO
		return false, fmt.Sprintf("%d", result.Code)
	}
	if !c.VerifyHmacSing(result.Data.Sign, SignContent(param), crypto.SHA256) {
		order.OrderStatus = data.OrderFail
		return false, fmt.Sprintf("%d", result.Code)
	}
	order.OrderStatus = data.OrderSuccess
	order.OutTradeNo = result.Data.OrderNo
	return true, fmt.Sprintf("%d", result.Code)
}

// 充值查单接口
func (t *JYPayConfig) InspectSubmit(order *entity.PayOrder) ([]byte, error) {
	body := t.InspectCompose(order) //打包post数据
	strReq, _ := jsoniter.Marshal(body)
	resp, err := doHttpForm(t.InspectOrderUrl, strReq)
	if err != nil {
		return nil, err
	}
	glog.Infof("InspectSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

// 充值查单响应
func (c *JYPayConfig) InspectResponse(body []byte, order *entity.PayOrder) error {
	result := new(JyOrderRespond)
	param := make(map[string]string)
	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		return err
	}

	param["merNo"] = result.Data.MerNo
	param["orderNo"] = result.Data.OrderNo
	param["subCode"] = utils.String(result.Data.SubCode)
	param["subMsg"] = result.Data.SubMsg
	param["status"] = utils.String(result.Data.Status)
	param["orderAmount"] = result.Data.OrderAmount
	param["payAmount"] = utils.String(result.Data.PayAmount)
	param["busiCode"] = result.Data.BusiCode
	param["payTime"] = result.Data.PayTime

	if result.Code != 200 { // 没成功TODO
		return fmt.Errorf("code:%d, message:%s", result.Code, result.Message)
	}

	if !c.VerifyHmacSing(result.Data.Sign, SignContent(param), crypto.SHA256) {
		return fmt.Errorf("verify sign fail")
	}

	if result.Data.Status != 2 {
		return fmt.Errorf("status:%d, message:%s", result.Data.Status, result.Message)
	}
	return nil
}

// 充值查单
func (c *JYPayConfig) InspectCompose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["merNo"] = c.Merchant
	dictionary["merOrderNo"] = order.OrderID
	dictionary["orderNo"] = order.OutTradeNo
	dictionary["requestNo"] = utils.String(utils.BsonNow().UnixMilli())
	dictionary["timestamp"] = utils.String(utils.BsonNow().UnixMilli())

	// 签名
	strSign := HmacRSASign(SignContent(dictionary), c.HamcRSA, crypto.SHA256)
	dictionary["sign"] = strSign

	return dictionary
}

// 提现查单接口
func (t *JYPayConfig) InspectWithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	body := t.InspectWdCompose(order) //打包post数据
	strReq, _ := jsoniter.Marshal(body)
	resp, err := doHttpForm(t.InspectWithdrawUrl, strReq)
	if err != nil {
		return nil, err
	}
	glog.Infof("InspectWithdrawSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

// 提现查单响应
func (c *JYPayConfig) InspectWdResponse(body []byte, order *entity.WithdrawOrder) int {
	result := new(JyOrderRespond)
	jsoniter.Unmarshal(body, result)
	if result.Code != 200 { // 不成功
		return 3
	}
	if result.Data.Status != 2 {
		return 3
	}
	return 2
}

// 提现查单
func (c *JYPayConfig) InspectWdCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["merNo"] = c.Merchant
	dictionary["merOrderNo"] = order.OrderID
	dictionary["requestNo"] = utils.String(utils.BsonNow().UnixMilli())
	dictionary["orderNo"] = order.OutTradeNo
	dictionary["timestamp"] = utils.String(utils.BsonNow().UnixMilli())

	// 签名
	strSign := HmacRSASign(SignContent(dictionary), c.HamcRSA, crypto.SHA256)
	dictionary["sign"] = strSign

	return dictionary
}

// 余额查询
func (c *JYPayConfig) BalanceQuery() int64 {
	param := c.QueryCompose()
	strReq, _ := jsoniter.Marshal(param)
	resp, err := doHttpForm(c.QueryUrl, strReq)
	if err != nil {
		glog.Info("[jypay] BalanceQuery order fail, err,", err)
		return 0
	}

	rsp := new(JyBalanceRespond)
	err = jsoniter.Unmarshal(resp, &rsp)
	if err != nil {
		glog.Info("[jypay] BalanceQuery order fail, err,", err)
		return 0
	}

	if rsp.Code != 200 {
		glog.Info("[jypay] BalanceQuery order fail, code:", rsp.Code)
		return 0
	}
	glog.Infof("[jypay] BalanceQuery order success")

	var balance float64
	for _, l := range rsp.Data.List {
		// b, _ := strconv.ParseFloat(l.Balance, 64)
		balance += l.Balance
	}
	return int64(balance * 100)
}

func (c *JYPayConfig) QueryCompose() map[string]string {
	dictionary := map[string]string{
		"merNo":     c.Merchant,
		"requestNo": utils.String(utils.BsonNow().UnixMilli()),
		"timestamp": utils.String(utils.BsonNow().UnixMilli()),
	}

	strSign := HmacRSASign(SignContent(dictionary), c.HamcRSA, crypto.SHA256)
	// dictionary["sign"] = pay.ToUpper(strSign)
	dictionary["sign"] = strSign

	return dictionary
}
