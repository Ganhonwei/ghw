package oepay

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"goserver/internal/global/pay/entity"
	"goserver/internal/global/pay/service"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var Config *OePayConfig

// 充值订单
func (c *OePayConfig) Submit(order *entity.PayOrder) ([]byte, error) {
	body, sign := c.compose(order) //打包post数据
	strpar, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strpar)
	// glog.Debugf("sign body: %s", order.RequestParam)
	// glog.Debugf("sign: %s", sign)

	headers := map[string]string{
		"X-SERVICE-CODE": c.ChannelId,
		"X-SIGN":         sign,
	}
	resp, err := doHttpPostJson(c.PayOrderUrl, strpar, headers)
	if err != nil {
		return nil, err
	}
	glog.Infof("oepay Submit order success, id:%s", order.OrderID)
	return resp, nil
}

// 充值响应
func (c *OePayConfig) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	result := new(OePayOrderRespond)
	err := json.Unmarshal(body, result)
	if err != nil {
		return "", err
	}
	// 没成功
	if result.Code != "0000" || result.Data.LinkUrl == "" {
		return "", fmt.Errorf("code:%s, message:%s", result.Code, result.Msg)
	}
	order.OutTradeNo = result.Data.PlatformOrderNo
	return result.Data.LinkUrl, nil
}

func (c *OePayConfig) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	body, sign := c.withdrawCompose(order) //打包post数据
	strpar, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strpar)
	headers := map[string]string{
		"X-SERVICE-CODE": c.ChannelId,
		"X-SIGN":         sign,
	}
	resp, err := doHttpPostJson(c.WithDrawUrl, strpar, headers)
	if err != nil {
		return nil, err
	}
	glog.Infof("oepay WithdrawSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

func (c *OePayConfig) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	result := new(OePayWithdrawRespond)
	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		order.OrderStatus = data.DeliveryFailTransfer
		return false, "parse json fail:" + err.Error()
	}
	if result.Code == "0000" { // 成功
		// 发条消息提单成功
		glog.Infof("[oepay withdraw] submit success. order:%s, user:%s, traceId:%s", order.OrderID, order.Userid, result.Data.PlatformOrderNo)
		order.OrderStatus = data.OrderSuccess
		return true, fmt.Sprint(result.Code)
	}
	// 没成功
	glog.Infof("[oepay withdraw] submit fail. code:%s, msg:%s", result.Code, result.Msg)
	order.OrderStatus = data.OrderFail
	return false, fmt.Sprintf("code:%s, message:%s", result.Code, result.Msg)
}

// 待收订单查询
func (c *OePayConfig) InspectSubmit(order *entity.PayOrder) ([]byte, error) {
	params := make(map[string]string)
	params["orderNo"] = order.OrderID
	// 签名
	strSign := PaySign(params, c.keyClientPriv)
	headers := map[string]string{
		"X-SERVICE-CODE": c.ChannelId,
		"X-SIGN":         strSign,
	}
	strpar, err := jsoniter.Marshal(params)
	if err != nil {
		return nil, err
	}
	resp, err := doHttpPostJson(c.InspectOrderUrl, strpar, headers)
	if err != nil {
		return nil, err
	}
	glog.Infof("InspectSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

func (c *OePayConfig) InspectResponse(body []byte, order *entity.PayOrder) error {
	result := new(InspectRespond)
	if err := jsoniter.Unmarshal(body, result); err != nil {
		return err
	}
	if result.Code != "0000" { // 没成功
		return fmt.Errorf("code:%s, message:%s", result.Code, result.Msg)
	}

	if result.Data.Status != 1 {
		return fmt.Errorf("state:%d, message:%s", result.Data.Status, result.Msg)
	}
	return nil
}

// 提现查单接口
func (c *OePayConfig) InspectWithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	params := make(map[string]string)
	params["orderNo"] = order.OrderID
	// 签名
	strSign := PaySign(params, c.keyClientPriv)
	headers := map[string]string{
		"X-SERVICE-CODE": c.ChannelId,
		"X-SIGN":         strSign,
	}
	strpar, err := jsoniter.Marshal(params)
	if err != nil {
		return nil, err
	}
	resp, err := doHttpPostJson(c.InspectWithdrawUrl, strpar, headers)
	if err != nil {
		return nil, err
	}
	glog.Infof("InspectWithdrawSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

func (c *OePayConfig) InspectWdResponse(body []byte, order *entity.WithdrawOrder) int {
	result := new(InspectRespond)
	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		glog.Error("InspectWdResponse json unmarshal error:", err, string(body))
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

// application/json PSOT请求
func doHttpPostJson(targetUrl string, body []byte, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest("POST", targetUrl, bytes.NewReader(body))
	if err != nil {
		glog.Errorf("err:%s", err)
		return []byte(""), err
	}
	req.Header.Add("Content-type", "application/json")
	if len(headers) > 0 {
		for name, val := range headers {
			req.Header.Add(name, val)
		}
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   0,
			KeepAlive: 0,
		}).Dial,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		TLSHandshakeTimeout: 10 * time.Second,
	}

	if service.Proxy != "" {
		transport.Proxy = func(r *http.Request) (*url.URL, error) {
			return url.Parse(service.Proxy)
		}
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

// 充值组装参数
func (c *OePayConfig) compose(order *entity.PayOrder) (params map[string]string, sign string) {
	params = make(map[string]string)
	params["orderNo"] = order.OrderID
	params["amount"] = fmt.Sprintf("%.2f", float64(order.Amount)/100)
	params["firstname"] = order.RealName
	params["mobile"] = order.Mobile
	params["email"] = order.Email
	params["surl"] = c.SuccessUrl
	params["furl"] = c.FailUrl
	params["remark"] = fmt.Sprintf("oe,%s,%s", params["amount"], order.MerOrderID)

	// 签名
	sign = PaySign(params, c.keyClientPriv)
	return
}

// 提现组装参数
func (c *OePayConfig) withdrawCompose(order *entity.WithdrawOrder) (params map[string]string, sign string) {
	params = make(map[string]string)
	params["orderNo"] = order.OrderID
	params["amount"] = fmt.Sprintf("%.2f", float64(order.Amount)/100)
	params["paymentType"] = "0" // 0银行卡,1UPI
	params["upi"] = ""
	params["beneficiaryName"] = order.RealName
	params["beneficiaryMobile"] = order.Mobile
	params["bankAccount"] = order.BankNumber
	params["ifsc"] = order.IFSC
	params["beneficiaryEmail"] = order.Email
	params["remark"] = fmt.Sprintf("oew,%s,%s,", params["amount"], order.MerOrderID)

	// 签名
	strSign := PaySign(params, c.keyClientPriv)
	sign = strSign
	return
}

// 余额查询
func (c *OePayConfig) BalanceQuery() int64 {
	// strReq := ToQueryString(param)
	resp, err := doHttpGET(c.QueryUrl + "?serviceCode=" + c.ChannelId)
	if err != nil {
		glog.Infof("[oepay] BalanceQuery order FAIL,err:", err)
		return 0
	}

	rsp := new(BalanceRespond)
	err = jsoniter.Unmarshal(resp, &rsp)
	if err != nil {
		glog.Infof("[oepay] BalanceQuery order FAIL,err:", err)
		return 0
	}

	if rsp.Code != "0000" {
		glog.Infof("[oepay] BalanceQuery order Code fail,code:", rsp.Code)
		return 0
	}
	glog.Infof("[oepay] BalanceQuery order success")
	return int64((rsp.Data.DeductAmount + rsp.Data.PaymentAmount - rsp.Data.FrozenAmount) * 100)
}

// 参数转换为签名对象, 去空值, 排序key
func signContent(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var result strings.Builder
	for i, v := range keys {
		if val, ok := params[v]; ok && val != "" {
			urlencode := url.QueryEscape(val)
			urlencode = strings.ReplaceAll(urlencode, "+", "%20") // 空格被编码为+，而+本身被编码为%2B, 这里需要把空格编码为%20
			tmp := "&%s=%s"
			if i == 0 {
				tmp = "%s=%s"
			}
			result.WriteString(fmt.Sprintf(tmp, v, urlencode))
		}
	}
	return result.String()
}

func PaySign(params map[string]string, keyClientPriv *rsa.PrivateKey) string {
	content := signContent(params)
	// glog.Debugf("sign content: %s", content)

	hash := crypto.SHA512
	shaNew := hash.New()
	shaNew.Write([]byte(content))
	hashed := shaNew.Sum(nil)
	sign, err := rsa.SignPKCS1v15(rand.Reader, keyClientPriv, hash, hashed)
	if err != nil {
		glog.Error("oepay rsa sign fail, err:", err)
		return ""
	}
	s := base64.StdEncoding.EncodeToString(sign)
	return s
}

// 验签
func (c *OePayConfig) VerifySing(outSign string, params map[string]string) bool {
	content := signContent(params)

	hash := crypto.SHA512
	shaNew := hash.New()
	shaNew.Write([]byte(content))
	hashed := shaNew.Sum(nil)

	outSignBytes, err := base64.StdEncoding.DecodeString(outSign)
	if err != nil {
		glog.Error("verify sign fail, decode sign fail")
		return false
	}
	err = rsa.VerifyPKCS1v15(c.keyServerPub, hash, hashed[:], outSignBytes)
	if err != nil {
		glog.Error("rsa fail", err)
		return false
	}
	return true
}

func (t *OePayConfig) ParsePayResult(body []byte) (map[string]string, error) {
	glog.Debugf("oepay notice body: %s", string(body))
	tradeResult := OePayRechargeCallback{}
	err := jsoniter.Unmarshal(body, &tradeResult)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	result["platformOrderNo"] = tradeResult.PlatformOrderNo
	result["orderNo"] = tradeResult.OrderNo
	result["status"] = tradeResult.Status // fmt.Sprint(tradeResult.Status)
	result["amount"] = tradeResult.Amount // fmt.Sprintf("%.2f", tradeResult.Amount)
	result["fee"] = tradeResult.Fee       // fmt.Sprintf("%.2f", tradeResult.Fee)
	result["linkUrl"] = tradeResult.LinkUrl
	result["utrNo"] = tradeResult.UtrNo
	result["errorMsg"] = tradeResult.ErrorMsg
	result["udf1"] = tradeResult.Udf1
	result["udf2"] = tradeResult.Udf2
	result["udf3"] = tradeResult.Udf3
	result["udf4"] = tradeResult.Udf4
	result["udf5"] = tradeResult.Udf5
	return result, nil
}

// application/json PSOT请求
func doHttpGET(targetUrl string) ([]byte, error) {
	req, err := http.NewRequest("GET", targetUrl, nil)
	if err != nil {
		glog.Errorf("err:%s", err)
		return []byte(""), err
	}
	// req.Header.Add("Content-type", "application/json")
	// if len(headers) > 0 {
	// 	for name, val := range headers {
	// 		req.Header.Add(name, val)
	// 	}
	// }

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   0,
			KeepAlive: 0,
		}).Dial,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		TLSHandshakeTimeout: 10 * time.Second,
	}

	// if service.Proxy != "" {
	// 	transport.Proxy = func(r *http.Request) (*url.URL, error) {
	// 		return url.Parse(service.Proxy)
	// 	}
	// }

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
