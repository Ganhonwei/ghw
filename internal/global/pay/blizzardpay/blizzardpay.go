package blizzardpay

import (
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

var Config *BlizzardPayConfig

func ToQueryString(param map[string]string) string {
	var args string
	for k, v := range param {
		vStr := fmt.Sprintf("%v", v)

		if args == "" {
			args += fmt.Sprintf("%s=%s", k, vStr)
		} else {
			args += fmt.Sprintf("&%s=%s", k, vStr)
		}
	}

	return args
}

// Submit 下单
func (t *BlizzardPayConfig) Submit(order *entity.PayOrder) ([]byte, error) {
	body := t.compose(order) //打包post数据
	// sig := ComputeHMAC(body, Config.HamcRSA)
	strReq := ToQueryString(body)
	strpar, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strpar)
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
func (t *BlizzardPayConfig) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	// if !checkWallet(order.Bank) {
	// 	return nil, errors.New("wallet format error")
	// }

	body := t.WithdrawCompose(order) //打包post数据
	// sig := ComputeHMAC(body, Config.HamcRSA)
	strReq := ToQueryString(body)
	strpar, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strpar)
	resp, err := doHttpForm(t.WithDrawUrl, strReq)
	//fmt.Printf("resp %s, err %v", string(resp), err)
	if err != nil {
		return nil, err
	}
	glog.Infof("WithdrawSubmit order success, id:%s, resp:%s", order.OrderID, string(resp))
	return resp, nil
}

func PaySign(param map[string]string, md5key string) string {
	keys := make([]string, 0, len(param))
	for k := range param {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var result strings.Builder
	for _, v := range keys {
		if val, ok := param[v]; ok && val != "" {
			vStr := fmt.Sprintf("%v", val)
			result.WriteString(fmt.Sprintf("%s=%s&", v, vStr))
		}
	}

	if md5key != "" {
		result.WriteString(fmt.Sprintf("key=%s", md5key))
	}

	return utils.Md5(result.String())
}

// 组装参数
func (c *BlizzardPayConfig) compose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)

	// wallets := []string{"baksh", "ngand"}
	// wallet, _ := utils.ChoiceString(wallets)
	// if wallet == "" {
	// 	wallet = "baksh"
	// }

	dictionary["appId"] = c.Merchant
	dictionary["outTradeNo"] = order.OrderID
	dictionary["channelId"] = "108"
	dictionary["amount"] = fmt.Sprintf("%.2f", float32(order.Amount)/100)
	dictionary["callbackUrl"] = Config.PayNotifyUrl
	dictionary["successUrl"] = "https://google.com"
	dictionary["clientUserIp"] = order.RegistIp
	dictionary["clientUserId"] = order.Userid

	strSign := PaySign(dictionary, c.Key)
	dictionary["sign"] = strSign
	// phone := service.GenerateMJLPhone()
	// dictionary["transaction_type"] = "income"
	// dictionary["amount"] = fmt.Sprintf("%.2f", float32(order.Amount)/100)
	// dictionary["merchant_order_id"] = order.OrderID
	// dictionary["merchant_uid"] = c.Merchant
	// dictionary["currency"] = "bdt"
	// dictionary["notify_url"] = Config.PayNotifyUrl
	// dictionary["response_type"] = "json"

	// dictionary["bankCode"] = wallet
	// dictionary["merNo"] = c.Merchant
	// dictionary["merOrderNo"] = order.OrderID
	// dictionary["name"] = order.RealName
	// dictionary["phone"] = phone
	// dictionary["email"] = fmt.Sprintf("%s@gmail.com", phone)
	// dictionary["orderAmount"] = fmt.Sprintf("%.2f", float32(order.Amount)/100)
	// dictionary["notifyUrl"] = Config.PayNotifyUrl // 充值后返回地址
	// dictionary["currency"] = "BDT"
	// dictionary["busiCode"] = "124001"
	// dictionary["pageUrl"] = "https://www.google.com/"
	// dictionary["timestamp"] = utils.String(utils.BsonNow().UnixMilli())

	// 签名
	// strSign := HmacRSASign(SignContent(dictionary), c.HamcRSA, crypto.SHA256)
	// dictionary["sign"] = strSign
	return dictionary
}

// 提现组装参数
func (c *BlizzardPayConfig) WithdrawCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["appId"] = c.Merchant
	dictionary["outOrderNo"] = order.OrderID
	dictionary["amount"] = fmt.Sprintf("%.2f", float32(order.Amount)/100)
	dictionary["bankName"] = order.IFSC
	dictionary["bankUserName"] = order.RealName
	dictionary["bankCard"] = order.BankNumber
	dictionary["currency"] = "INR"
	dictionary["callbackUrl"] = Config.WithDrawNotifyUrl

	strSign := PaySign(dictionary, c.Key)
	dictionary["sign"] = strSign
	// dictionary["transaction_type"] = "expense"
	// dictionary["amount"] = fmt.Sprintf("%.2f", float32(order.Amount)/100)
	// dictionary["merchant_order_id"] = order.OrderID
	// dictionary["merchant_uid"] = c.Merchant
	// dictionary["currency"] = "BDT"
	// dictionary["notify_url"] = Config.WithDrawNotifyUrl
	// dictionary["response_type"] = "json"
	// dictionary["to"] = order.BankNumber
	// dictionary["to_name"] = order.RealName
	// dictionary["to_bank_name"] = WalletMap[order.Bank]

	// phone := service.GenerateMJLPhone()
	// dictionary["accName"] = order.RealName
	// dictionary["accNo"] = order.BankNumber
	// dictionary["bankCode"] = WalletMap[order.Bank]
	// dictionary["busiCode"] = "223001"
	// dictionary["currency"] = "BDT"
	// dictionary["email"] = fmt.Sprintf("%s@gmail.com", phone)
	// dictionary["merNo"] = c.Merchant
	// dictionary["merOrderNo"] = order.OrderID
	// dictionary["notifyUrl"] = Config.WithDrawNotifyUrl // 提现后回调地址
	// dictionary["orderAmount"] = fmt.Sprintf("%.2f", float32(order.Amount)/100)
	// dictionary["phone"] = phone
	// dictionary["timestamp"] = fmt.Sprintf("%d", utils.BsonNow().UnixMilli())

	// // 签名
	// strSign := HmacRSASign(SignContent(dictionary), c.HamcRSA, crypto.SHA256)
	// // 再加密
	// dictionary["sign"] = rsaSign(strSign, c.RSAPriRSA, crypto.SHA256)
	return dictionary
}

func ComputeHMAC(params map[string]string, secret string) string {

	var keys []string
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Construct the string
	var kvs []string
	for _, key := range keys {
		kvs = append(kvs, key+"="+params[key])
	}
	dataToSign := strings.Join(kvs, "&")

	secretBytes := []byte(secret)
	h := hmac.New(sha256.New, secretBytes)
	h.Write([]byte(dataToSign))
	return hex.EncodeToString(h.Sum(nil))
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

// application/x-www-form-urlencoded PSOT请求
func doHttpForm(targetUrl string, body string) ([]byte, error) {
	req, err := http.NewRequest("POST", targetUrl, strings.NewReader(body))
	if err != nil {
		glog.Errorf("err:%s", err)
		return []byte(""), err
	}
	req.Header.Add("Content-type", "application/x-www-form-urlencoded")

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
func (c *BlizzardPayConfig) VerifySing(sign, content string, hash crypto.Hash) bool {
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
func (c *BlizzardPayConfig) VerifyHmacSing(sign, content string, hash crypto.Hash) bool {
	outSign := HmacRSASign(content, c.HamcRSA, hash)
	return outSign == sign
}

// =========================实现接口方法=========================

func (c *BlizzardPayConfig) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	result := new(BlizzardPayOrderRespond)
	// param := make(map[string]string)
	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		return "", err
	}
	// param["merNo"] = result.Data.MerNo
	// param["merOrderNo"] = result.Data.MerOrderNo
	// param["orderNo"] = result.Data.OrderNo
	// param["subCode"] = utils.String(result.Data.SubCode)
	// param["subMsg"] = result.Data.SubMsg
	// param["status"] = utils.String(result.Data.Status)
	// param["orderData"] = result.Data.OrderData
	// param["common"] = result.Data.Common

	if result.Code != 200 { // 没成功TODO
		return "", fmt.Errorf("result:%d,reason:%s, OrderNo:%s", result.Code, result.Message, order.OrderID)
	}
	// if !c.VerifyHmacSing(result.Data.Sign, SignContent(param), crypto.SHA256) {
	// 	return "", errors.New("verify sign fail")
	// }

	if result.Data.PayUrl == "" {
		return "", errors.New("pay link is null")
	}

	order.OutTradeNo = result.Data.OrderNo
	return result.Data.PayUrl, nil
	// return "", nil
}

func (c *BlizzardPayConfig) ParsePayResult(body []byte) (map[string]string, error) {
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

func (c *BlizzardPayConfig) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	result := new(BlizzardPayOrderRespond)
	// param := make(map[string]string)
	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		return false, err.Error()
	}

	if result.Code != 200 {
		order.OrderStatus = data.OrderFail
		return false, fmt.Sprintf("result:%d,reason:%s, OrderNo:%s", result.Code, result.Message, order.OrderID)
	}
	// param["merNo"] = result.Data.MerNo
	// param["merOrderNo"] = result.Data.MerOrderNo
	// param["orderNo"] = result.Data.OrderNo
	// param["subCode"] = utils.String(result.Data.SubCode)
	// param["subMsg"] = result.Data.SubMsg
	// param["status"] = utils.String(result.Data.Status)
	// if result.Code != 200 { // 没成功TODO
	// 	return false, fmt.Sprintf("%d", result.Code)
	// }
	// if !c.VerifyHmacSing(result.Data.Sign, SignContent(param), crypto.SHA256) {
	// 	return false, fmt.Sprintf("%d", result.Code)
	// }

	order.OutTradeNo = result.Data.OrderNo
	order.OrderStatus = data.OrderSuccess
	return true, fmt.Sprintf("%d", result.Code)
}

// 充值查单接口
func (t *BlizzardPayConfig) InspectSubmit(order *entity.PayOrder) ([]byte, error) {
	return nil, fmt.Errorf("not implement")
	// body := t.InspectCompose(order) //打包post数据
	// strReq, _ := jsoniter.Marshal(body)
	// resp, err := doHttpForm(t.InspectOrderUrl, strReq, "", "")
	// if err != nil {
	// 	return nil, err
	// }
	// glog.Infof("InspectSubmit order success, id:%s", order.OrderID)
	// return resp, nil
}

// 充值查单响应
func (c *BlizzardPayConfig) InspectResponse(body []byte, order *entity.PayOrder) error {
	// result := new(BlizzardPayOrderRespond)
	// param := make(map[string]string)
	// err := jsoniter.Unmarshal(body, result)
	// if err != nil {
	// return err
	// }

	// param["merNo"] = result.Data.MerNo
	// param["orderNo"] = result.Data.OrderNo
	// param["subCode"] = utils.String(result.Data.SubCode)
	// param["subMsg"] = result.Data.SubMsg
	// param["status"] = utils.String(result.Data.Status)
	// param["orderAmount"] = result.Data.OrderAmount
	// param["payAmount"] = utils.String(result.Data.PayAmount)
	// param["busiCode"] = result.Data.BusiCode
	// param["payTime"] = result.Data.PayTime

	// if result.Code != 200 { // 没成功TODO
	// 	return fmt.Errorf("code:%d, message:%s", result.Code, result.Message)
	// }

	// if !c.VerifyHmacSing(result.Data.Sign, SignContent(param), crypto.SHA256) {
	// 	return fmt.Errorf("verify sign fail")
	// }

	// if result.Data.Status != 2 {
	// 	return fmt.Errorf("status:%d, message:%s", result.Data.Status, result.Message)
	// }
	return fmt.Errorf("not implement")
}

// 充值查单
func (c *BlizzardPayConfig) InspectCompose(order *entity.PayOrder) map[string]string {
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
func (t *BlizzardPayConfig) InspectWithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	return nil, fmt.Errorf("not implement")
	// body := t.InspectWdCompose(order) //打包post数据
	// strReq, _ := jsoniter.Marshal(body)
	// resp, err := doHttpForm(t.InspectWithdrawUrl, strReq, "", "")
	// if err != nil {
	// 	return nil, err
	// }
	// glog.Infof("InspectWithdrawSubmit order success, id:%s", order.OrderID)
	// return resp, nil
}

// 提现查单响应
func (c *BlizzardPayConfig) InspectWdResponse(body []byte, order *entity.WithdrawOrder) int {
	// result := new(BlizzardPayOrderRespond)
	// jsoniter.Unmarshal(body, result)
	// if result.Code != 200 { // 不成功
	// 	return 3
	// }
	// if result.Data.Status != 2 {
	// 	return 3
	// }
	return 3
}

// 提现查单
func (c *BlizzardPayConfig) InspectWdCompose(order *entity.WithdrawOrder) map[string]string {
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
func (c *BlizzardPayConfig) BalanceQuery() int64 {
	return 0
	// param := c.QueryCompose()
	// sig := ComputeHMAC(param, Config.HamcRSA)
	// strReq, _ := jsoniter.Marshal(param)
	// resp, err := doHttpForm(c.QueryUrl, strReq, Config.Key, sig)
	// if err != nil {
	// 	glog.Info("blizzardpay BalanceQuery order fail, err,", err)
	// 	return 0
	// }

	// rsp := new(_99BalanceRespond)
	// err = jsoniter.Unmarshal(resp, &rsp)
	// if err != nil {
	// 	glog.Info("blizzardpay BalanceQuery order fail, err,", err)
	// 	return 0
	// }

	// if rsp.Result != "success" {
	// 	glog.Info("blizzardpay BalanceQuery order fail")
	// 	return 0
	// }
	// glog.Infof("blizzardpay BalanceQuery order success")

	// var balance float64
	// // for _, l := range rsp.Data.List {
	// b, _ := strconv.ParseFloat(rsp.Data.Bdt, 64)
	// balance += b
	// // }
	// return int64(balance * 100)
}

func (c *BlizzardPayConfig) QueryCompose() map[string]string {
	dictionary := map[string]string{
		"merchant_uid": c.Merchant,
		// "requestNo": utils.String(utils.BsonNow().UnixMilli()),
		// "timestamp": utils.String(utils.BsonNow().UnixMilli()),
	}

	// strSign := HmacRSASign(SignContent(dictionary), c.HamcRSA, crypto.SHA256)
	// dictionary["sign"] = pay.ToUpper(strSign)
	// dictionary["sign"] = strSign

	return dictionary
}
