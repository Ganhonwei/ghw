package transafepay

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

var Config *TransafePayConfig

const (
	pem_begin = "-----BEGIN PRIVATE KEY-----\n"
	pem_end   = "\n-----END PRIVATE KEY-----"
	pub_begin = "-----BEGIN PUBLIC KEY-----\n"
	pub_end   = "\n-----END PUBLIC KEY-----"
)

// Submit 下单
func (t *TransafePayConfig) Submit(order *entity.PayOrder) ([]byte, error) {
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
func (t *TransafePayConfig) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
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
func (c *TransafePayConfig) compose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)

	phone := service.GenerateMJLPhone()
	dictionary["txChannel"] = "TX_BD_001"
	dictionary["mchId"] = c.Merchant
	dictionary["appId"] = c.AppId
	dictionary["mchOrderNo"] = order.OrderID
	dictionary["bankCode"] = "NAGAD"
	dictionary["amount"] = fmt.Sprintf("%.0f", float32(order.Amount)/100)
	dictionary["name"] = order.RealName
	dictionary["phone"] = phone
	dictionary["email"] = fmt.Sprintf("%s@gmail.com", phone)
	dictionary["notifyUrl"] = Config.PayNotifyUrl // 充值后返回地址
	dictionary["productInfo"] = fmt.Sprintf("%s-Rechange", order.OrderID)
	// dictionary["busiCode"] = "124001"
	// dictionary["pageUrl"] = "https://www.google.com/"
	dictionary["timestamp"] = utils.String(utils.BsonNow().Unix())

	// 签名
	strSign := rsaSign(SignContent(dictionary), c.RSAPriRSA, crypto.SHA1)
	dictionary["sign"] = strSign
	return dictionary
}

// 提现组装参数
func (c *TransafePayConfig) WithdrawCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)

	phone := service.GenerateMJLPhone()
	dictionary["name"] = order.RealName
	dictionary["account"] = order.BankNumber
	dictionary["bankCode"] = getWallet(order.Bank)
	// dictionary["busiCode"] = "223001"
	// dictionary["currency"] = "BDT"
	dictionary["email"] = fmt.Sprintf("%s@gmail.com", phone)
	dictionary["mchId"] = c.Merchant
	dictionary["appId"] = c.AppId
	dictionary["mchOrderNo"] = order.OrderID
	dictionary["notifyUrl"] = Config.WithDrawNotifyUrl // 提现后回调地址
	dictionary["amount"] = fmt.Sprintf("%.2f", float32(order.Amount)/100)
	dictionary["phone"] = phone
	dictionary["txChannel"] = "TX_BD_001"
	dictionary["timestamp"] = fmt.Sprintf("%d", utils.BsonNow().Unix())

	// 签名
	strSign := rsaSign(SignContent(dictionary), c.RSAPriRSA, crypto.SHA1)
	dictionary["sign"] = strSign //rsaSign(strSign, c.RSAPriRSA, crypto.SHA256)
	return dictionary
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
func (c *TransafePayConfig) VerifySing(sign, content string, hash crypto.Hash) bool {
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

// =========================实现接口方法=========================

func (c *TransafePayConfig) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	result := new(TransafeOrderRespond)
	param := make(map[string]string)
	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		return "", err
	}
	param["mchOrderNo"] = result.Data.MchOrderNo
	param["platOrderNo"] = result.Data.PlatOrderNo
	param["applyTime"] = utils.String(result.Data.ApplyTime)
	param["amount"] = utils.String(result.Data.Amount)
	param["reference"] = result.Data.Reference
	param["link"] = result.Data.Link
	param["expireTime"] = utils.String(result.Data.ExpireTime)
	if result.Status != 200 { // 没成功TODO
		return "", fmt.Errorf("code:%d,message:%s, OrderNo:%s", result.Status, result.Msg, order.OrderID)
	}
	// if !c.VerifyHmacSing(result.Data.Sign, SignContent(param), crypto.SHA256) {
	// 	return "", errors.New("verify sign fail")
	// }

	if result.Data.Link == "" {
		return "", errors.New("pay link is null")
	}

	order.OutTradeNo = result.Data.MchOrderNo
	return result.Data.Link, nil
	// return "", nil
}

func (c *TransafePayConfig) ParsePayResult(body []byte) (map[string]string, error) {
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

func (c *TransafePayConfig) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	result := new(TransafeOrderRespond)
	param := make(map[string]string)
	jsoniter.Unmarshal(body, result)
	param["platOrderNo"] = result.Data.PlatOrderNo
	param["mchOrderNo"] = result.Data.MchOrderNo
	param["applyTime"] = utils.String(result.Data.ApplyTime)
	param["amount"] = utils.String(result.Data.Amount)
	param["orderStatus"] = result.Data.OrderStatus
	if result.Status != 200 { // 没成功TODO
		order.OrderStatus = data.OrderFail
		return false, fmt.Sprintf("%d", result.Status)
	}
	// if !c.VerifyHmacSing(result.Data.Sign, SignContent(param), crypto.SHA256) {
	// 	return false, fmt.Sprintf("%d", result.Status)
	// }

	order.OrderStatus = data.OrderSuccess
	order.OutTradeNo = result.Data.PlatOrderNo
	return true, fmt.Sprintf("%d", result.Status)
}

// 充值查单接口
func (t *TransafePayConfig) InspectSubmit(order *entity.PayOrder) ([]byte, error) {
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
func (c *TransafePayConfig) InspectResponse(body []byte, order *entity.PayOrder) error {
	result := new(TransafeOrderRespond)
	param := make(map[string]string)
	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		return err
	}
	param["mchOrderNo"] = result.Data.MchOrderNo
	param["platOrderNo"] = result.Data.PlatOrderNo
	param["orderStatus"] = result.Data.OrderStatus
	param["trxId"] = utils.String(result.Data.TrxId)
	param["amount"] = utils.String(result.Data.Amount)
	param["fee"] = utils.String(result.Data.Fee)
	param["timeEnd"] = utils.String(result.Data.TimeEnd)
	if result.Status != 200 { // 没成功TODO
		return fmt.Errorf("code:%d, message:%s", result.Status, result.Msg)
	}

	// if !c.VerifyHmacSing(result.Data.Sign, SignContent(param), crypto.SHA256) {
	// 	return fmt.Errorf("verify sign fail")
	// }

	if result.Data.OrderStatus != "SUCCESS" {
		return fmt.Errorf("status:%s, message:%s", result.Data.OrderStatus, result.Msg)
	}
	return nil
}

// 充值查单
func (c *TransafePayConfig) InspectCompose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["mchId"] = c.Merchant
	dictionary["mchOrderNo"] = order.OrderID
	dictionary["orderNo"] = order.OutTradeNo
	dictionary["timestamp"] = utils.String(utils.BsonNow().Unix())

	// 签名
	strSign := rsaSign(SignContent(dictionary), c.RSAPriRSA, crypto.SHA1)
	dictionary["sign"] = strSign

	return dictionary
}

// 提现查单接口
func (t *TransafePayConfig) InspectWithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
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
func (c *TransafePayConfig) InspectWdResponse(body []byte, order *entity.WithdrawOrder) int {
	result := new(TransafeOrderRespond)
	jsoniter.Unmarshal(body, result)
	if result.Status != 200 { // 不成功
		return 3
	}
	if result.Data.OrderStatus != "SUCCESS" {
		return 3
	}
	return 2
}

// 提现查单
func (c *TransafePayConfig) InspectWdCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["mchId"] = c.Merchant
	dictionary["mchOrderNo"] = order.OrderID
	dictionary["timestamp"] = utils.String(utils.BsonNow().Unix())

	// 签名
	strSign := rsaSign(SignContent(dictionary), c.RSAPriRSA, crypto.SHA1)
	dictionary["sign"] = strSign

	return dictionary
}

// 余额查询
func (c *TransafePayConfig) BalanceQuery() int64 {
	param := c.QueryCompose()
	strReq, _ := jsoniter.Marshal(param)
	resp, err := doHttpForm(c.QueryUrl, strReq)
	if err != nil {
		glog.Info("[TransafePay] BalanceQuery order fail, err,", err)
		return 0
	}

	rsp := new(TransafeBalanceRespond)
	err = jsoniter.Unmarshal(resp, &rsp)
	if err != nil {
		glog.Info("[TransafePay] BalanceQuery order fail, err,", err)
		return 0
	}

	if rsp.Status != 200 {
		glog.Info("[TransafePay] BalanceQuery order fail, code:", rsp.Status)
		return 0
	}
	glog.Infof("[TransafePay] BalanceQuery order success")

	var balance float64
	if rsp.Data.Balance != 0 {
		b, _ := strconv.ParseFloat(utils.String(rsp.Data.Balance), 64)
		balance += b
	}
	if rsp.Data.Frozen != 0 {
		b, _ := strconv.ParseFloat(utils.String(rsp.Data.Frozen), 64)
		balance += b
	}
	return int64(balance * 100)
}

func (c *TransafePayConfig) QueryCompose() map[string]string {
	dictionary := make(map[string]string)

	dictionary["mchId"] = c.Merchant
	dictionary["timestamp"] = utils.String(utils.BsonNow().Unix())

	strSign := rsaSign(SignContent(dictionary), c.RSAPriRSA, crypto.SHA1)
	dictionary["sign"] = strSign

	return dictionary
}
