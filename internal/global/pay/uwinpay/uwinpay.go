package uwinpay

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"encoding/base64"
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

var Config *UWinPayConfig

// Submit 下单
func (t *UWinPayConfig) Submit(order *entity.PayOrder) ([]byte, error) {
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
func (t *UWinPayConfig) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
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
func (c *UWinPayConfig) compose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["merchant_code"] = c.Merchant
	dictionary["merchant_order_no"] = order.OrderID
	dictionary["account_type"] = "IFSC"
	dictionary["name"] = order.RealName
	// dictionary["name"] = "sss"
	dictionary["mobile"] = order.Mobile
	dictionary["email"] = order.Email
	dictionary["amount"] = fmt.Sprintf("%.2f", float32(order.Amount)/100)
	dictionary["notify_url"] = Config.PayNotifyUrl // 充值后返回地址
	dictionary["order_time"] = fmt.Sprintf("%d", utils.BsonNow().UnixMilli())
	dictionary["customer_ip"] = order.RegistIp
	dictionary["page_url"] = "ss"
	dictionary["description"] = fmt.Sprintf("rehcarge%.2f", float32(order.Amount)/100)
	// dictionary["extendInfo"] = ""

	// 签名
	strSign := rsaSign(SignContent(dictionary), c.ClientPriRSA, crypto.SHA256)
	// dictionary["sign"] = pay.ToUpper(strSign)
	dictionary["sign"] = strSign
	return dictionary
}

// 提现组装参数
func (c *UWinPayConfig) WithdrawCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["merchant_code"] = c.Merchant
	dictionary["merchant_order_no"] = order.OrderID
	dictionary["account_type"] = "IFSC"
	dictionary["account_code"] = order.IFSC
	dictionary["account_number"] = order.BankNumber
	// dictionary["currency"] = "INR"
	dictionary["amount"] = fmt.Sprintf("%.2f", float32(order.Amount)/100)
	dictionary["name"] = order.RealName
	dictionary["mobile"] = order.Mobile
	dictionary["email"] = order.Email
	dictionary["notify_url"] = Config.WithDrawNotifyUrl // 提现后回调地址
	dictionary["order_time"] = fmt.Sprintf("%d", utils.BsonNow().UnixMilli())
	dictionary["description"] = fmt.Sprintf("withdraw %.2f", float32(order.Amount)/100)

	// 签名
	strSign := rsaSign(SignContent(dictionary), c.ClientPriRSA, crypto.SHA256)
	// dictionary["sign"] = pay.ToUpper(strSign)
	dictionary["sign"] = strSign
	return dictionary
}

func rsaSign(content, key string, hash crypto.Hash) string {
	shaNew := hash.New()
	shaNew.Write([]byte(content))
	hashed := shaNew.Sum(nil)
	prikey, err := service.ParsePrivateKey(Config.ClientPriRSA)
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
func (c *UWinPayConfig) VerifySing(sign, content string, hash crypto.Hash) bool {
	shaNew := hash.New()
	shaNew.Write([]byte(content))
	hashed := shaNew.Sum(nil)
	pubkey, err := service.ParsePublickKey(c.ServerPubRSA)
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

// =========================实现接口方法=========================

func (c *UWinPayConfig) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	result := new(UWinOrderRespond)
	param := make(map[string]string)
	jsoniter.Unmarshal(body, result)
	param["code"] = fmt.Sprintf("%d", result.Code)
	param["message"] = result.Message
	param["order_no"] = result.Data.Order
	param["merchant_order_no"] = result.Data.MerchantOrderNo
	param["merchant_code"] = result.Data.MerchantCode
	param["amount"] = result.Data.Amount
	param["payUrl"] = result.Data.PayLink
	// param["extendInfo"] = result.ExtendInfo
	if result.Code != 200 { // 没成功TODO
		return "", fmt.Errorf("code:%d,message:%s, OrderNo:%s", result.Code, result.Message, result.Data.MerchantOrderNo)
	}
	// if !c.VerifySing(result.Data.Sign, SignContent(param), crypto.SHA256) {
	// 	return "", errors.New("verify sign fail")
	// }
	return result.Data.PayLink, nil
	// return "", nil
}

func (c *UWinPayConfig) ParsePayResult(body []byte) (map[string]string, error) {
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

func (c *UWinPayConfig) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	result := new(UWinOrderRespond)
	param := make(map[string]string)
	jsoniter.Unmarshal(body, result)
	param["code"] = fmt.Sprintf("%d", result.Code)
	param["message"] = result.Message
	param["order_no"] = result.Data.Order
	param["merchant_order_no"] = result.Data.MerchantOrderNo
	param["merchant_code"] = result.Data.MerchantCode
	param["amount"] = result.Data.Amount
	param["payUrl"] = result.Data.PayLink
	// param["extendInfo"] = result.ExtendInfo
	if result.Code != 200 { // 没成功TODO
		order.OrderStatus = data.OrderFail
		return false, fmt.Sprintf("%d", result.Code)
	}
	// if !c.VerifySing(result.Data.Sign, SignContent(param), crypto.SHA256) {
	// 	return false, fmt.Sprintf("%d", result.Code)
	// }
	order.OrderStatus = data.OrderSuccess
	return true, fmt.Sprintf("%d", result.Code)
}

// 充值查单接口
func (t *UWinPayConfig) InspectSubmit(order *entity.PayOrder) ([]byte, error) {
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
func (c *UWinPayConfig) InspectResponse(body []byte, order *entity.PayOrder) error {
	result := new(UWinPayInspectRespond)
	jsoniter.Unmarshal(body, result)
	if result.Code != 200 { // 没成功TODO
		return fmt.Errorf("code:%d, message:%s", result.Code, result.Message)
	}
	if result.Data.Status != 2 {
		return fmt.Errorf("status:%d, message:%s", result.Data.Status, result.Message)
	}
	return nil
}

// 充值查单
func (c *UWinPayConfig) InspectCompose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["merchant_code"] = c.Merchant
	dictionary["merchant_order_no"] = fmt.Sprint(order.OrderID)

	// 签名
	strSign := rsaSign(SignContent(dictionary), c.ClientPriRSA, crypto.SHA256)
	// dictionary["sign"] = pay.ToUpper(strSign)
	dictionary["sign"] = strSign

	return dictionary
}

// 提现查单接口
func (t *UWinPayConfig) InspectWithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
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
func (c *UWinPayConfig) InspectWdResponse(body []byte, order *entity.WithdrawOrder) int {
	result := new(UWinPayInspectRespond)
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
func (c *UWinPayConfig) InspectWdCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["merchant_code"] = c.Merchant
	dictionary["merchant_order_no"] = fmt.Sprint(order.OrderID)

	// 签名
	strSign := rsaSign(SignContent(dictionary), c.ClientPriRSA, crypto.SHA256)
	dictionary["sign"] = strSign

	return dictionary
}

// 余额查询
func (c *UWinPayConfig) BalanceQuery() int64 {
	param := c.QueryCompose()
	strReq, _ := jsoniter.Marshal(param)
	resp, err := doHttpForm(c.QueryUrl, strReq)
	if err != nil {
		glog.Info("[uwinpay] BalanceQuery order fail, err,", err)
		return 0
	}

	rsp := new(UWinBalanceRespond)
	err = jsoniter.Unmarshal(resp, &rsp)
	if err != nil {
		glog.Info("[uwinpay] BalanceQuery order fail, err,", err)
		return 0
	}

	if rsp.Code != 200 {
		glog.Info("[uwinpay] BalanceQuery order fail, code:", rsp.Code)
		return 0
	}
	glog.Infof("[uwinpay] BalanceQuery order success")

	balance, _ := strconv.ParseFloat(rsp.Data.AvailableBalance, 64)
	return int64(balance * 100)
}

func (c *UWinPayConfig) QueryCompose() map[string]string {
	dictionary := map[string]string{
		"merchant_code": c.Merchant,
	}

	strSign := rsaSign(SignContent(dictionary), c.ClientPriRSA, crypto.SHA256)
	// dictionary["sign"] = pay.ToUpper(strSign)
	dictionary["sign"] = strSign

	return dictionary
}
