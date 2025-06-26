package xfpay

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"goserver/internal/global/pay/entity"
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

var Config *XFPayConfig

// Submit 下单
func (t *XFPayConfig) Submit(order *entity.PayOrder) ([]byte, error) {
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
func (t *XFPayConfig) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
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
func (c *XFPayConfig) compose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["machId"] = c.MachId
	dictionary["merchantOrderNo"] = fmt.Sprint(order.OrderID)
	dictionary["amount"] = strconv.Itoa(int(order.Amount / 100))
	dictionary["name"] = order.NickName
	dictionary["email"] = order.Email
	dictionary["phone"] = order.Mobile
	dictionary["remark"] = "Recharge"
	dictionary["returnUrl"] = Config.PayNotifyUrl // 充值后返回地址
	// dictionary["successUrl"] = ""

	// 签名
	strSign := PaySign(dictionary, c.Md5Key)
	// dictionary["sign"] = pay.ToUpper(strSign)
	dictionary["sign"] = strSign
	return dictionary
}

// 提现组装参数
func (t *XFPayConfig) WithdrawCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["machId"] = t.MachId
	dictionary["name"] = order.RealName
	dictionary["amount"] = strconv.Itoa(int(order.Amount / 100))
	dictionary["account"] = order.BankNumber
	dictionary["returnUrl"] = t.WithDrawNotifyUrl
	dictionary["idCode"] = order.IFSC
	dictionary["bankName"] = order.Bank
	dictionary["phone"] = order.Mobile
	dictionary["merchantOrderNo"] = order.OrderID

	// 签名
	strSign := PaySign(dictionary, t.WithDrawMDd5Key)
	// dictionary["sign"] = pay.ToUpper(strSign)
	dictionary["sign"] = strSign
	return dictionary
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
			result.WriteString(fmt.Sprintf("%s=%s&", v, param[v]))
		}
	}

	if md5key != "" {
		result.WriteString(fmt.Sprintf("key=%s", md5key))
	}
	return utils.Md5(result.String())
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

func (c *XFPayConfig) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	result := new(XFOrderRespond)
	jsoniter.Unmarshal(body, result)
	if result.Code != "200" { // 没成功TODO
		return "", fmt.Errorf("code:%s, OrderNo:%s", result.Code, result.OrderNo)
	}
	order.OutTradeNo = result.OrderNo
	return result.PaymentUrl, nil
}

func (c *XFPayConfig) ParsePayResult(body []byte) (map[string]string, error) {
	tradeResult := XFRechargeCallback{}
	err := jsoniter.Unmarshal(body, &tradeResult)
	if err != nil {
		return nil, err
	}
	reslut := make(map[string]string)
	reslut["merchantOrderNo"] = tradeResult.MerchantOrderNo
	reslut["payMoney"] = tradeResult.PayMoney
	reslut["payTime"] = tradeResult.PayTime
	reslut["status"] = tradeResult.Status
	reslut["sign"] = tradeResult.Sign
	return reslut, nil
}

func (c *XFPayConfig) ParseWithdrawResult(body []byte) (map[string]string, error) {
	tradeResult := XFWithdrawCallback{}
	err := jsoniter.Unmarshal(body, &tradeResult)
	if err != nil {
		return nil, err
	}
	reslut := make(map[string]string)
	reslut["merchantOrderNo"] = tradeResult.MerchantOrderNo
	reslut["withdrawalMoney"] = tradeResult.WithdrawalMoney
	reslut["withdrawalTime"] = tradeResult.WithdrawalTime
	reslut["status"] = tradeResult.Status
	reslut["sign"] = tradeResult.Sign
	return reslut, nil
}

func (c *XFPayConfig) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	result := new(XFWithdrawRespond)
	jsoniter.Unmarshal(body, result)
	if result.Code != "200" || result.Status != "SUCCESS" { // 没成功TODO
		glog.Errorf("[xfpay] withdraw fail submit, code:%s, status:%s", result.Code, result.Status)
		order.OrderStatus = data.OrderFail
		return false, result.Code
	}
	order.OrderStatus = data.OrderSuccess
	return true, result.Code
}

// 充值查单接口
func (t *XFPayConfig) InspectSubmit(order *entity.PayOrder) ([]byte, error) {
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
func (c *XFPayConfig) InspectResponse(body []byte, order *entity.PayOrder) error {
	result := new(XFInspectRespond)
	jsoniter.Unmarshal(body, result)
	if result.Status != "2" { // 没成功TODO
		return fmt.Errorf("code:%s, message:%s", result.Status, result.Remark)
	}
	return nil
}

// 充值查单
func (c *XFPayConfig) InspectCompose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["machId"] = c.AppId
	dictionary["appId"] = c.MachId
	dictionary["merchantOrderNo"] = fmt.Sprint(order.OrderID)

	// 签名
	strSign := PaySign(dictionary, c.Md5Key)
	// dictionary["sign"] = pay.ToUpper(strSign)
	dictionary["sign"] = strSign

	return dictionary
}

// 提现查单接口
func (t *XFPayConfig) InspectWithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
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
func (c *XFPayConfig) InspectWdResponse(body []byte, order *entity.WithdrawOrder) int {
	result := new(XFInspectRespond)
	jsoniter.Unmarshal(body, result)
	if result.Status == "2" { // 成功
		return 2
	}
	return 3
}

// 提现查单
func (c *XFPayConfig) InspectWdCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["machId"] = c.AppId
	dictionary["appId"] = c.MachId
	dictionary["merchantOrderNo"] = fmt.Sprint(order.OrderID)

	// 签名
	strSign := PaySign(dictionary, c.WithDrawMDd5Key)
	// dictionary["sign"] = pay.ToUpper(strSign)
	dictionary["sign"] = strSign

	return dictionary
}

// 余额查询
func (c *XFPayConfig) BalanceQuery() int64 {

	return 0
	// return resp, nil
}
