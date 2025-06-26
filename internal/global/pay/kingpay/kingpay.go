package kingpay

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"goserver/internal/global/pay/entity"

	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"io"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var Config *KingPayConfig

const (
	pem_begin = "-----BEGIN PRIVATE KEY-----\n"
	pem_end   = "\n-----END PRIVATE KEY-----"
	pub_begin = "-----BEGIN PUBLIC KEY-----\n"
	pub_end   = "\n-----END PUBLIC KEY-----"
)

// Submit 下单
func (t *KingPayConfig) Submit(order *entity.PayOrder) ([]byte, error) {
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
func (t *KingPayConfig) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
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
func (c *KingPayConfig) compose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["merchantId"] = c.ApplicationId
	dictionary["appId"] = c.Appid
	dictionary["timestamp"] = utils.String(utils.BsonNow().Unix())
	dictionary["outOrderId"] = order.OrderID
	dictionary["payType"] = "301"
	dictionary["product"] = "1"
	dictionary["describe"] = "buy"
	dictionary["amount"] = fmt.Sprintf("%.2f", float32(order.Amount)/100)
	dictionary["payerEmail"] = order.Email
	dictionary["payerPhone"] = order.Mobile
	dictionary["returnUrl"] = ""
	dictionary["notifyUrl"] = Config.PayNotifyUrl // 充值后返回地址
	// dictionary["extendInfo"] = ""

	// 签名
	strSign := rsaSign(SignContent(dictionary), c.ClientPriRSA, crypto.SHA256)
	// dictionary["sign"] = pay.ToUpper(strSign)
	dictionary["sign"] = strSign
	return dictionary
}

// 提现组装参数
func (c *KingPayConfig) WithdrawCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["merchantId"] = c.ApplicationId
	dictionary["appId"] = c.Appid
	dictionary["timestamp"] = utils.String(utils.BsonNow().Unix())
	// dictionary["currency"] = "INR"
	dictionary["outOrderId"] = order.OrderID
	dictionary["withdrawAmount"] = fmt.Sprintf("%.2f", float32(order.Amount)/100)
	dictionary["accountName"] = order.RealName
	dictionary["cardNumber"] = order.BankNumber
	dictionary["bankCode"] = order.IFSC
	dictionary["bankName"] = order.Bank
	dictionary["bankSubbranch"] = ""
	dictionary["payeePhone"] = order.Mobile
	dictionary["notifyUrl"] = Config.WithDrawNotifyUrl // 提现后回调地址

	// 签名
	strSign := rsaSign(SignContent(dictionary), c.ClientPriRSA, crypto.SHA256)
	// dictionary["sign"] = pay.ToUpper(strSign)
	dictionary["sign"] = strSign
	return dictionary
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

func rsaSign(content, key string, hash crypto.Hash) string {
	shaNew := hash.New()
	shaNew.Write([]byte(content))
	hashed := shaNew.Sum(nil)
	prikey, err := parsePrivateKey(Config.ClientPriRSA)
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
	// b, _ := base64.StdEncoding.DecodeString(s)
	// shaNew = hash.New()
	// shaNew.Write([]byte(content))
	// hashed = shaNew.Sum(nil)
	// pubkey, _ := parsePublickKey("MFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAMc8H3hgSEiSCoSbWOYsUzd34cqdFzE3HJza7lr3Zofo2IrdLsN4r0lpMsLTCgHF6NxC3akUzSvkWMnRtsaU+Y0CAwEAAQ==")
	// err2 := rsa.VerifyPKCS1v15(pubkey, hash, hashed[:], b)
	// if err2 != nil {
	// 	return s
	// }
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
func (c *KingPayConfig) VerifySing(sign, content string, hash crypto.Hash) bool {
	shaNew := hash.New()
	shaNew.Write([]byte(content))
	hashed := shaNew.Sum(nil)
	pubkey, err := parsePublickKey(c.ServerPubRSA)
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

func (c *KingPayConfig) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	result := new(KingOrderRespond)
	param := make(map[string]string)
	jsoniter.Unmarshal(body, result)
	param["code"] = result.Code
	param["message"] = result.Message
	param["timestamp"] = fmt.Sprintf("%d", result.Timestamp)
	if result.Timestamp == 0 {
		param["timestamp"] = ""
	}
	// param["sign"] = result.Sign
	param["orderId"] = result.OrderId
	param["outOrderId"] = result.OutOrderId
	if result.Status == 0 {
		param["status"] = ""
	} else {
		param["status"] = fmt.Sprintf("%d", result.Status)
	}
	param["statusDesc"] = result.StatusDesc
	param["payUrl"] = result.PayUrl
	// param["extendInfo"] = result.ExtendInfo
	if !c.VerifySing(result.Sign, SignContent(param), crypto.SHA256) {
		return "", errors.New("verify sign fail")
	}
	if result.Code != "0" { // 没成功TODO
		return "", fmt.Errorf("code:%s,message:%s, OrderNo:%s", result.Code, result.Message, result.OutOrderId)
	}
	return result.PayUrl, nil
}

func (c *KingPayConfig) ParsePayResult(body []byte) (map[string]string, error) {
	tradeResult := KingRechargeCallback{}
	err := jsoniter.Unmarshal(body, &tradeResult)
	if err != nil {
		return nil, err
	}
	reslut := make(map[string]string)
	reslut["timestamp"] = fmt.Sprintf("%d", tradeResult.Timestamp)
	reslut["orderId"] = tradeResult.OrderId
	reslut["outOrderId"] = tradeResult.OutOrderId
	reslut["status"] = fmt.Sprintf("%d", tradeResult.Status)
	reslut["statusDesc"] = tradeResult.StatusDesc
	reslut["amount"] = tradeResult.Amount
	reslut["tradeFee"] = tradeResult.TradeFee
	if tradeResult.PayTime != 0 {
		reslut["payTime"] = fmt.Sprintf("%d", tradeResult.PayTime)
	} else {
		reslut["payTime"] = ""
	}
	reslut["sign"] = tradeResult.Sign
	return reslut, nil
}

func (c *KingPayConfig) ParseWithdrawResult(body []byte) (map[string]string, error) {
	tradeResult := KingWithdrawCallback{}
	err := jsoniter.Unmarshal(body, &tradeResult)
	if err != nil {
		return nil, err
	}
	reslut := make(map[string]string)
	if tradeResult.Timestamp == 0 {
		reslut["timestamp"] = ""
	} else {
		reslut["timestamp"] = fmt.Sprintf("%d", tradeResult.Timestamp)
	}
	reslut["orderId"] = tradeResult.OrderId
	reslut["outOrderId"] = tradeResult.OutOrderId
	reslut["status"] = fmt.Sprintf("%d", tradeResult.Status)
	reslut["statusDesc"] = tradeResult.StatusDesc
	reslut["withdrawAmount"] = tradeResult.WithdrawAmount
	if tradeResult.TransferTime != 0 {
		reslut["transferTime"] = fmt.Sprintf("%d", tradeResult.TransferTime)
	} else {
		reslut["transferTime"] = ""
	}
	reslut["sign"] = tradeResult.Sign
	return reslut, nil
}

func (c *KingPayConfig) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	result := new(KingOrderRespond)
	param := make(map[string]string)
	jsoniter.Unmarshal(body, result)
	param["code"] = result.Code
	param["message"] = result.Message
	if result.Timestamp != 0 {
		param["timestamp"] = fmt.Sprintf("%d", result.Timestamp)
	} else {
		param["timestamp"] = ""
	}
	param["orderId"] = result.OrderId
	param["outOrderId"] = result.OutOrderId
	if result.Status == 0 {
		param["status"] = ""
	} else {
		param["status"] = fmt.Sprintf("%d", result.Status)
	}
	param["statusDesc"] = result.StatusDesc
	if !c.VerifySing(result.Sign, SignContent(param), crypto.SHA256) {
		glog.Errorf("[kingpay withdraw] verify sign fail, %v", param)
		return false, result.Code
	}
	if result.Code != "0" { // 没成功TODO
		order.OrderStatus = data.OrderFail
		return false, result.Code
	}
	order.OrderStatus = data.OrderSuccess
	return true, result.Code
}

// InspectSubmit 核查
func (t *KingPayConfig) InspectSubmit(order *entity.PayOrder) ([]byte, error) {
	// body := t.inspectCompose(order) //打包post数据
	// // strReq := pay.ToQueryString(body)
	// strReq, _ := jsoniter.Marshal(body)
	// order.RequestParam = string(strReq)
	// //fmt.Printf("body %s, order %#v\n", body, order)
	// resp, err := doHttpForm(t.InspectOrderUrl, strReq)
	// //fmt.Printf("resp %s, err %v", string(resp), err)
	// if err != nil {
	// 	return nil, err
	// }
	// glog.Infof("Inspect order success, id:%s", order.OrderID)
	// return resp, nil
	return nil, fmt.Errorf("fail")
}

// 核单响应
func (c *KingPayConfig) InspectResponse(body []byte, order *entity.PayOrder) error {
	return nil
}

// 核单响应
func (c *KingPayConfig) InspectWdResponse(body []byte, order *entity.WithdrawOrder) int {
	return 3
}

// 提现查单接口
func (t *KingPayConfig) InspectWithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	return nil, fmt.Errorf("fail")
}

// 余额查询
func (t *KingPayConfig) BalanceQuery() int64 {
	// param := t.QueryCompose()
	// strpar, _ := jsoniter.Marshal(param)
	// resp, err := doHttpForm(t.QueryUrl, strpar)
	// if err != nil {
	// 	return nil, err
	// }
	glog.Infof("[kingpay] BalanceQuery unsupport")
	// return resp, nil
	return 0
}
