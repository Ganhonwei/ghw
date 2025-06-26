package icepay

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
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
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var Config *IcePayConfig

// Submit 充值下单
func (t *IcePayConfig) Submit(order *entity.PayOrder) ([]byte, error) {
	body := t.compose(order) //打包post数据
	// strReq := ToQueryString(body)
	strpar, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strpar)
	resp, err := doHttpForm(t.PayOrderUrl, strpar)
	if err != nil {
		return nil, err
	}
	glog.Infof("Submit order success, id:%s", order.OrderID)
	return resp, nil
}

// 提现下单
func (t *IcePayConfig) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	body := t.WithdrawCompose(order) //打包post数据
	// strReq := ToQueryString(body)
	strpar, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strpar)
	resp, err := doHttpForm(t.WithDrawUrl, strpar)
	//fmt.Printf("resp %s, err %v", string(resp), err)
	if err != nil {
		return nil, err
	}
	glog.Infof("WithdrawSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

// 充值响应
func (t *IcePayConfig) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	result := new(IcePayOrderRespond)
	err := json.Unmarshal(body, result)
	if err != nil {
		return "", err
	}
	if result.Code != 100 || result.PaymentUrl == "" { // 没成功TODO
		return "", fmt.Errorf("code:%d, message:%s", result.Code, result.Msg)
	}
	order.OutTradeNo = result.PayOrderId
	return result.PaymentUrl, nil
}

// 提现响应
func (t *IcePayConfig) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	result := new(IcePayWithdrawRespond)
	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		return false, "parse json fail"
	}
	if result.Code == 100 { // 没成功TODO
		// 发条消息提单成功
		// glog.Infof("[withdraw fail] submit success. order %s, user:%s", order.OrderID, order.Userid)
		order.OrderStatus = data.OrderSuccess
		return true, fmt.Sprint(result.Code)
	}
	glog.Infof("[icepay withdraw] submit fail. code:%s, msg:%s", result.Code, result.Msg)
	order.OrderStatus = data.OrderFail
	return false, fmt.Sprint(result.Code)
}

func (t *IcePayConfig) ParsePayResult(body []byte) (map[string]string, error) {
	tradeResult := IcePayRechargeCallback{}
	err := jsoniter.Unmarshal(body, &tradeResult)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	result["amount"] = fmt.Sprint(tradeResult.Amount)
	result["merchantId"] = fmt.Sprint(tradeResult.MerchantId)
	result["orderId"] = tradeResult.OrderId
	result["payOrderId"] = tradeResult.PayOrderId
	result["status"] = fmt.Sprint(tradeResult.Status)
	result["timestamp"] = fmt.Sprint(tradeResult.Timestamp)
	result["sign"] = tradeResult.Sign

	return result, nil
}

// 充值组装参数
func (t *IcePayConfig) compose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["merchantId"] = t.Appid
	dictionary["orderId"] = order.OrderID
	dictionary["amount"] = fmt.Sprintf("%d", order.Amount)
	dictionary["timestamp"] = fmt.Sprintf("%d", time.Now().UnixMilli())

	// dictionary["amount"] = fmt.Sprintf("%.2f", float64(order.Amount)/100)
	// dictionary["product"] = "indiaupi"
	// dictionary["bankcode"] = "all"
	// dictionary["goods"] = fmt.Sprintf("email:%s/name:%s/phone:%s/ip:%s", order.Email, order.RealName, order.Mobile, order.RegistIp)
	// dictionary["returnUrl"] = "https://www.google.ocm"
	// dictionary["notifyUrl"] = t.PayNotifyUrl

	// 签名
	strSign := PaySign(dictionary, t.Md5Key)
	dictionary["sign"] = strSign

	dictionary["notifyUrl"] = t.PayNotifyUrl

	return dictionary
}

// 提现组装参数
func (t *IcePayConfig) WithdrawCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["amount"] = fmt.Sprintf("%d", order.Amount)
	dictionary["merchantId"] = t.Appid
	dictionary["orderId"] = order.OrderID
	dictionary["timestamp"] = fmt.Sprintf("%d", time.Now().UnixMilli())
	// dictionary["type"] = "api"

	// dictionary["notifyUrl"] = t.WithDrawNotifyUrl
	// dictionary["accountName"] = order.RealName
	// dictionary["accountNo"] = order.BankNumber
	// dictionary["bankCode"] = order.IFSC
	// dictionary["remarkInfo"] = fmt.Sprintf("email:%s/phone:%s/mode:bank", order.Email, order.Mobile)

	// 签名
	strSign := PaySign(dictionary, t.Md5Key)
	dictionary["sign"] = strSign

	dictionary["notifyUrl"] = t.WithDrawNotifyUrl
	dictionary["outType"] = "IMPS"
	dictionary["accountNumber"] = order.BankNumber
	dictionary["ifsc"] = order.IFSC
	dictionary["accountHolder"] = order.RealName

	return dictionary
}

// 充值查单接口
func (t *IcePayConfig) InspectSubmit(order *entity.PayOrder) ([]byte, error) {
	body := t.InspectCompose(order) //打包post数据
	// strReq := ToQueryString(body)
	strpar, _ := jsoniter.Marshal(body)
	resp, err := doHttpForm(t.InspectOrderUrl, strpar)
	if err != nil {
		return nil, err
	}
	glog.Infof("InspectSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

// 提现查单接口
func (t *IcePayConfig) InspectWithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	body := t.InspectWdCompose(order) //打包post数据
	// strReq := ToQueryString(body)
	strpar, _ := jsoniter.Marshal(body)
	resp, err := doHttpForm(t.InspectWithdrawUrl, strpar)
	if err != nil {
		return nil, err
	}
	glog.Infof("InspectWithdrawSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

// 充值查单
func (t *IcePayConfig) InspectCompose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["merchantId"] = t.Appid
	dictionary["orderId"] = order.OrderID
	dictionary["timestamp"] = fmt.Sprint(utils.LocalTime().UnixMilli())

	// 签名
	strSign := PaySign(dictionary, t.Md5Key)
	dictionary["sign"] = strSign

	return dictionary
}

// 提现查单
func (t *IcePayConfig) InspectWdCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["merchantId"] = t.Appid
	dictionary["orderId"] = order.OrderID
	dictionary["timestamp"] = fmt.Sprint(utils.LocalTime().UnixMilli())

	// 签名
	strSign := PaySign(dictionary, t.Md5Key)
	dictionary["sign"] = strSign

	return dictionary
}

// 充值查单响应
func (c *IcePayConfig) InspectResponse(body []byte, order *entity.PayOrder) error {
	result := new(InspectRespond)
	jsoniter.Unmarshal(body, result)
	if result.Code != 100 { // 没成功TODO
		return fmt.Errorf("code:%d, message:%s", result.Code, result.Msg)
	}
	if !c.InspectVerfiySign(&result.Data) {
		return fmt.Errorf("verify inspect fail, order:%s", order.OrderID)
	}

	if result.Data.Status != 1 {
		return fmt.Errorf("state:%d, message:%s", result.Data.Status, result.Msg)
	}
	return nil
}

// 提现查单响应
func (c *IcePayConfig) InspectWdResponse(body []byte, order *entity.WithdrawOrder) int {
	result := new(InspectRespond)
	jsoniter.Unmarshal(body, result)
	if result.Code != 100 { // 没成功TODO
		return 3
	}

	if !c.InspectVerfiySign(&result.Data) {
		glog.Errorf("verify inspect fail, order:%s", order.OrderID)
		return 3
	}

	if result.Data.Status == 1 {
		return 2
	}
	return 3
}

func (c *IcePayConfig) InspectVerfiySign(data *InspectData) bool {
	if data == nil {
		glog.Errorf("verify inspect sign fail")
		return false
	}
	outSign := data.Sign

	param := make(map[string]string)

	param["merchantId"] = data.MerchantId
	param["amount"] = utils.String(data.Amount)
	param["orderId"] = data.OrderId
	param["payOrderId"] = data.PayOrderId
	param["status"] = utils.String(data.Status)
	param["statusDesc"] = data.StatusDesc
	param["timestamp"] = utils.String(data.Timestamp)
	sign := PaySign(param, c.Md5Key)

	return outSign == sign
}

// 余额查询
func (t *IcePayConfig) BalanceQuery() int64 {
	param := t.QueryCompose()
	strpar, _ := jsoniter.Marshal(param)
	resp, err := doHttpForm(t.QueryUrl, strpar)
	if err != nil {
		glog.Info("[icepay] BalanceQuery order fail,err:", err)
		return 0
	}

	rsp := make(map[string]any)
	err = jsoniter.Unmarshal(resp, &rsp)
	if err != nil {
		glog.Info("[icepay] BalanceQuery order fail,err:", err)
		return 0
	}

	if b, ok := rsp["balance"]; ok {
		glog.Infof("[icepay] BalanceQuery order success")
		return b.(int64)
	}
	return 0
}

// 提现查单
func (t *IcePayConfig) QueryCompose() map[string]string {
	dictionary := map[string]string{
		"merchantId": t.Appid,
		"timestamp":  utils.String(utils.BsonNow().UnixMilli()),
	}

	sign := PaySign(dictionary, t.Md5Key)
	dictionary["sign"] = sign

	return dictionary
}

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

func PaySign(param map[string]string, md5key string) string {
	keys := make([]string, 0, len(param))
	for k := range param {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var result strings.Builder
	for _, v := range keys {
		if val, ok := param[v]; ok && val != "" {
			// vStr := fmt.Sprintf("%v", val)
			result.WriteString(val)
		}
	}

	if md5key != "" {
		result.WriteString(md5key)
	}

	a := result.String()
	_ = a

	md5 := utils.Md5(result.String())
	return md5
	// return strings.ToUpper(md5)
}
