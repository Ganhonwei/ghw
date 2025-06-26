package wepay

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

var Config *WePayConfig

// 充值订单
func (t *WePayConfig) Submit(order *entity.PayOrder) ([]byte, error) {
	body := t.compose(order) //打包post数据
	strpar, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strpar)
	resp, err := doHttpPostJson(t.PayOrderUrl, strpar)
	if err != nil {
		return nil, err
	}
	glog.Infof("wepay Submit order success, id:%s", order.OrderID)
	return resp, nil
}

// 充值响应
func (t *WePayConfig) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	result := new(WePayOrderRespond)
	err := json.Unmarshal(body, result)
	if err != nil {
		return "", err
	}
	// 没成功
	if !result.Success || result.Data.PayUrl == "" {
		return "", fmt.Errorf("code:%d, message:%s", result.Code, result.Msg)
	}
	order.OutTradeNo = result.Data.TradeNo
	return result.Data.PayUrl, nil
}

// 提现订单
func (t *WePayConfig) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	body := t.withdrawCompose(order) //打包post数据
	strpar, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strpar)
	resp, err := doHttpPostJson(t.WithDrawUrl, strpar)
	if err != nil {
		return nil, err
	}
	glog.Infof("wepay WithdrawSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

func (t *WePayConfig) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	result := new(WePayWithdrawRespond)
	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		return false, "parse json fail"
	}
	if result.Success { // 成功
		// 发条消息提单成功
		glog.Infof("[wepay withdraw] submit success. order:%s, user:%s, traceId:%s", order.OrderID, order.Userid, result.Data.Id)
		order.OrderStatus = data.OrderSuccess
		return true, fmt.Sprint(result.Code)
	}
	// 没成功
	glog.Infof("[wepay withdraw] submit fail. code:%s, msg:%s", result.Code, result.Msg)
	order.OrderStatus = data.OrderFail
	return false, fmt.Sprintf("code:%d, message:%s", result.Code, result.Msg)
}

// 待收订单查询
func (t *WePayConfig) InspectSubmit(order *entity.PayOrder) ([]byte, error) {
	dictionary := make(map[string]string)
	dictionary["mchId"] = t.Appid
	dictionary["orderNo"] = order.OrderID
	// 签名
	strSign := PaySign(dictionary, t.Md5Key)
	dictionary["sign"] = strSign
	strpar, err := jsoniter.Marshal(dictionary)
	if err != nil {
		return nil, err
	}
	resp, err := doHttpPostJson(t.InspectOrderUrl, strpar)
	if err != nil {
		return nil, err
	}
	glog.Infof("InspectSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

// 充值查单响应
func (t *WePayConfig) InspectResponse(body []byte, order *entity.PayOrder) error {
	result := new(InspectRespond)
	if err := jsoniter.Unmarshal(body, result); err != nil {
		return err
	}
	if !result.Success { // 没成功
		return fmt.Errorf("code:%d, message:%s", result.Code, result.Msg)
	}

	if result.Data.PayStatus != 1 {
		return fmt.Errorf("state:%d, message:%s", result.Data.PayStatus, result.Msg)
	}
	return nil
}

// 提现查单接口
func (t *WePayConfig) InspectWithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	dictionary := make(map[string]string)
	dictionary["mchId"] = t.Appid
	dictionary["orderNo"] = order.OrderID
	// 签名
	strSign := PaySign(dictionary, t.Md5Key)
	dictionary["sign"] = strSign

	strpar, _ := jsoniter.Marshal(dictionary)
	resp, err := doHttpPostJson(t.InspectOrderUrl, strpar)
	if err != nil {
		return nil, err
	}
	glog.Infof("InspectWithdrawSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

func (t *WePayConfig) InspectWdResponse(body []byte, order *entity.WithdrawOrder) int {
	result := new(InspectRespond)
	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		glog.Error("InspectWdResponse json unmarshal error:", err, string(body))
		return 3
	}
	if !result.Success { // 没成功
		return 3
	}
	if result.Data.PayStatus == 1 {
		return 2
	}
	return 3
}

// 余额查询
func (t *WePayConfig) BalanceQuery() int64 {
	param := t.QueryCompose()
	strpar, _ := jsoniter.Marshal(param)
	resp, err := doHttpPostJson(t.QueryUrl, strpar)
	if err != nil {
		glog.Info("[wepay] BalanceQuery order fail,err:", err)
		return 0
	}

	rsp := new(BalanceRespond)
	err = jsoniter.Unmarshal(resp, &rsp)
	if err != nil {
		glog.Info("[wepay] BalanceQuery order fail,err:", err)
		return 0
	}

	if rsp.Code != 200 {
		glog.Info("[wepay] BalanceQuery order fail,code:", rsp.Code)
		return 0
	}

	glog.Infof("[wepay] BalanceQuery order success")
	return int64(rsp.Data.BalanceUsable * 100)
}

func (t *WePayConfig) QueryCompose() map[string]string {
	dictionary := map[string]string{
		"mchId": t.Appid,
	}

	strSign := PaySign(dictionary, t.Md5Key)
	dictionary["sign"] = strSign

	return dictionary
}

// application/json PSOT请求
func doHttpPostJson(targetUrl string, body []byte) ([]byte, error) {
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

// 充值组装参数
func (t *WePayConfig) compose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["mchId"] = t.Appid
	dictionary["passageId"] = t.PayPassageId
	dictionary["orderNo"] = order.OrderID
	dictionary["notifyUrl"] = t.PayNotifyUrl
	dictionary["amount"] = fmt.Sprintf("%.2f", float64(order.Amount)/100)

	// 签名
	strSign := PaySign(dictionary, t.Md5Key)
	dictionary["sign"] = strSign

	return dictionary
}

// 提现组装参数
func (t *WePayConfig) withdrawCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["mchId"] = t.Appid
	dictionary["passageId"] = t.WithdrawPassageId
	dictionary["orderNo"] = order.OrderID
	dictionary["account"] = order.BankNumber
	dictionary["userName"] = order.RealName
	dictionary["amount"] = fmt.Sprintf("%.2f", float64(order.Amount)/100)
	dictionary["notifyUrl"] = t.WithDrawNotifyUrl
	if order.IFSC != "" {
		dictionary["ifsc"] = order.IFSC
	}

	// 签名
	strSign := PaySign(dictionary, t.Md5Key)
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
			result.WriteString(fmt.Sprintf("%s=%s&", v, val))
		}
	}

	if md5key != "" {
		result.WriteString(fmt.Sprintf("key=%s", md5key))
	} else {
		glog.Errorf("wepay md5 key is empty... order: %s", param["orderNo"])
	}

	a := result.String()
	_ = a

	md5 := utils.Md5(result.String())
	return strings.ToLower(md5)
}

func (t *WePayConfig) ParsePayResult(body []byte) (map[string]string, error) {
	tradeResult := WePayRechargeCallback{}
	err := jsoniter.Unmarshal(body, &tradeResult)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	result["tradeNo"] = tradeResult.TradeNo
	result["orderNo"] = tradeResult.OrderNo
	result["orderAmount"] = fmt.Sprint(tradeResult.OrderAmount)
	result["amount"] = fmt.Sprint(tradeResult.Amount)
	result["payStatus"] = fmt.Sprint(tradeResult.PayStatus)
	result["charge"] = fmt.Sprint(tradeResult.Charge)
	result["otherData"] = tradeResult.OtherData
	result["reverse"] = fmt.Sprint(tradeResult.Reverse)
	result["remark"] = tradeResult.Remark
	result["sign"] = tradeResult.Sign
	if tradeResult.PayTime != 0 {
		result["payTime"] = fmt.Sprint(tradeResult.PayTime)
	}
	return result, nil
}
