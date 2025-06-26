package gentlepay

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"goserver/internal/global/pay/common"
	"goserver/internal/global/pay/entity"
	"goserver/internal/global/pay/service"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
)

const (
	pem_begin = "-----BEGIN PRIVATE KEY-----\n"
	pem_end   = "\n-----END PRIVATE KEY-----"
)

var Config *PayConfig

func (t *PayConfig) GetChannelId() uint32 {
	return service.GENTLEPAY
}

func (t *PayConfig) GetChannelName() string {
	return "gentlepay"
}

func (t *PayConfig) CheckMethod(ctx *fasthttp.RequestCtx) bool {
	return ctx.IsPost()
}

func (t *PayConfig) ParseParamMap(ctx *fasthttp.RequestCtx) (map[string]string, error) {
	body := ctx.PostBody()

	paramMap, err := t.ParsePayResult(body)
	if err != nil {
		return nil, err
	}

	return paramMap, nil
}

func (t *PayConfig) VerifySign(paramMap map[string]string) (bool, error) {
	outSign := paramMap["sign"]
	delete(paramMap, "sign")

	sign := PaySign(paramMap, t.MD5Key)

	return outSign == sign, nil
}

func (t *PayConfig) ParseIncomeParamInfo(paramMap map[string]string) (*service.ParamInfo, error) {
	paramInfo := new(service.ParamInfo)
	paramInfo.OrderId = paramMap["merchantOrderId"]
	paramInfo.Status = paramMap["state"]
	paramInfo.Message = ""
	paramInfo.TradeNo = paramMap["orderId"]
	paramInfo.Amount = common.YuanToFen(paramMap["amount"])

	if paramInfo.Status != "1" {
		paramInfo.StatusCode = data.PayFail
	} else {
		paramInfo.StatusCode = data.PaySuccess
	}

	jsonBytes, err := json.Marshal(paramMap)
	if err != nil {
		return nil, err
	}
	paramInfo.JsonStr = string(jsonBytes)
	return paramInfo, nil
}

func (t *PayConfig) ParseWithdrawParamInfo(paramMap map[string]string) (*service.ParamInfo, error) {
	paramInfo := new(service.ParamInfo)
	paramInfo.OrderId = paramMap["merchantOrderId"]
	paramInfo.Status = paramMap["state"]
	paramInfo.Message = paramMap["failReason"]
	paramInfo.TradeNo = paramMap["orderId"]
	paramInfo.Amount = common.YuanToFen(paramMap["amount"])

	if paramInfo.Status != "1" {
		paramInfo.StatusCode = data.WithdrawFail
	} else {
		paramInfo.StatusCode = data.WithdrawSuccess
	}

	jsonBytes, err := json.Marshal(paramMap)
	if err != nil {
		return nil, err
	}
	paramInfo.JsonStr = string(jsonBytes)
	return paramInfo, nil
}

func (t *PayConfig) GetSuccessMsg() string {
	return "success"
}

// Submit 下单
func (t *PayConfig) Submit(order *entity.PayOrder) ([]byte, error) {
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
func (t *PayConfig) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
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
func (c *PayConfig) compose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["merchantId"] = c.Merchant
	dictionary["amount"] = common.FenToYuan(order.Amount)
	dictionary["merchantOrderId"] = order.OrderID
	dictionary["notifyUrl"] = Config.PayNotifyUrl
	dictionary["userId"] = order.Userid
	dictionary["route"] = "1"

	strSign := PaySign(dictionary, c.MD5Key)
	dictionary["sign"] = strSign
	return dictionary
}

// 提现组装参数
func (c *PayConfig) WithdrawCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["merchantId"] = c.Merchant
	dictionary["amount"] = common.FenToYuan(order.Amount)
	dictionary["merchantOrderId"] = order.OrderID
	dictionary["notifyUrl"] = Config.WithDrawNotifyUrl
	dictionary["userId"] = order.Userid
	dictionary["payType"] = "2"
	dictionary["bankCard"] = order.BankNumber
	dictionary["bankCode"] = order.IFSC
	dictionary["route"] = "1"

	strSign := PaySign(dictionary, c.MD5Key)
	dictionary["sign"] = strSign

	return dictionary
}

func getPrivateKey(key string) string {
	if !strings.HasPrefix(key, pem_begin) {
		key = pem_begin + key
	}
	if !strings.HasSuffix(key, pem_end) {
		key = key + pem_end
	}
	return key
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
func (c *PayConfig) VerifySing(sign, content string, hash crypto.Hash) bool {
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

// 签名
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
		result.WriteString(fmt.Sprintf("secretKey=%s", md5key))
	}

	// return utils.Md5(result.String())
	signStr := result.String()
	signBytes := []byte(signStr)

	h := sha256.New()
	h.Write(signBytes)
	hashBytes := h.Sum(nil)

	return hex.EncodeToString(hashBytes)
}

// =========================实现接口方法=========================

func (c *PayConfig) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	result := new(PayOrderRespond)
	// param := make(map[string]string)
	jsoniter.Unmarshal(body, result)
	// param["code"] = fmt.Sprintf("%d", result.Code)
	// param["message"] = result.Message
	// param["order_no"] = result.Data.Order
	// param["merchant_order_no"] = result.Data.MerchantOrderNo
	// param["merchant_code"] = result.Data.MerchantCode
	// param["amount"] = result.Data.Amount
	// param["payUrl"] = result.Data.PayLink
	// param["extendInfo"] = result.ExtendInfo
	if result.Code != "1000" || result.Data.State != "4" { // 没成功TODO
		return "", fmt.Errorf("code:%s,message:%s, OrderNo:%s", result.Code, result.Msg, result.Data.MerchantOrderId)
	}
	// if !c.VerifySing(result.Data.Sign, SignContent(param), crypto.SHA256) {
	// 	return "", errors.New("verify sign fail")
	// }
	return result.Data.PayUrl, nil
	// return "", nil
}

func (c *PayConfig) ParsePayResult(body []byte) (map[string]string, error) {
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

func (c *PayConfig) ParseFormToMap(ctx *fasthttp.RequestCtx) map[string]string {
	// 创建 map 存储表单数据
	formData := make(map[string]string)

	// 遍历所有表单参数
	ctx.PostArgs().VisitAll(func(key, value []byte) {
		formData[string(key)] = string(value)
	})

	return formData
}

func (c *PayConfig) ValuesToMap(values url.Values) map[string]string {
	result := make(map[string]string)

	for key := range values {
		result[key] = values.Get(key)
	}

	return result
}

func (c *PayConfig) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	result := new(PayWithdrawRespond)
	// param := make(map[string]string)
	jsoniter.Unmarshal(body, result)
	// param["code"] = fmt.Sprintf("%d", result.Code)
	// param["message"] = result.Message
	// param["order_no"] = result.Data.Order
	// param["merchant_order_no"] = result.Data.MerchantOrderNo
	// param["merchant_code"] = result.Data.MerchantCode
	// param["amount"] = result.Data.Amount
	// param["payUrl"] = result.Data.PayLink
	// param["extendInfo"] = result.ExtendInfo

	if result.Code != "1000" || result.Data.State != "5" { // 没成功TODO
		order.OrderStatus = data.OrderFail
		return false, result.Msg
	}
	order.OrderStatus = data.OrderSuccess
	// if !c.VerifySing(result.Data.Sign, SignContent(param), crypto.SHA256) {
	// 	return false, fmt.Sprintf("%d", result.Code)
	// }
	return true, result.Msg
}

// 充值查单接口
func (t *PayConfig) InspectSubmit(order *entity.PayOrder) ([]byte, error) {
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
func (c *PayConfig) InspectResponse(body []byte, order *entity.PayOrder) error {
	result := new(PayInspectRespond)
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
func (c *PayConfig) InspectCompose(order *entity.PayOrder) map[string]string {
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
func (t *PayConfig) InspectWithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
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
func (c *PayConfig) InspectWdResponse(body []byte, order *entity.WithdrawOrder) int {
	result := new(PayInspectRespond)
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
func (c *PayConfig) InspectWdCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["merchant_code"] = c.Merchant
	dictionary["merchant_order_no"] = fmt.Sprint(order.OrderID)

	// 签名
	strSign := rsaSign(SignContent(dictionary), c.ClientPriRSA, crypto.SHA256)
	dictionary["sign"] = strSign

	return dictionary
}

// 余额查询
func (c *PayConfig) BalanceQuery() int64 {
	param := c.QueryCompose()
	strReq, _ := jsoniter.Marshal(param)
	resp, err := doHttpForm(c.QueryUrl, strReq)
	if err != nil {
		glog.Info("[uwinpay] BalanceQuery order fail, err,", err)
		return 0
	}

	rsp := new(PayBalanceRespond)
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

func (c *PayConfig) QueryCompose() map[string]string {
	dictionary := map[string]string{
		"merchant_code": c.Merchant,
	}

	strSign := rsaSign(SignContent(dictionary), c.ClientPriRSA, crypto.SHA256)
	// dictionary["sign"] = pay.ToUpper(strSign)
	dictionary["sign"] = strSign

	return dictionary
}
