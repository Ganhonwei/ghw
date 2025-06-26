package mmpay

import (
	"bytes"
	"crypto/tls"
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
	"strconv"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var Config *MMPayConfig

// Submit 下单
func (t *MMPayConfig) Submit(order *entity.PayOrder) ([]byte, error) {
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
func (t *MMPayConfig) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
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
func (c *MMPayConfig) compose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)

	phone := service.GenerateMJLPhone()
	dictionary["mer_no"] = c.Merchant
	dictionary["order_no"] = order.OrderID
	dictionary["amount"] = fmt.Sprintf("%d", order.Amount/100)
	dictionary["name"] = order.RealName
	dictionary["phone"] = phone
	dictionary["email"] = fmt.Sprintf("%s@gmail.com", phone)
	dictionary["currency"] = "BDT"
	dictionary["pay_code"] = "101400"
	dictionary["notify_url"] = Config.PayNotifyUrl // 充值后返回地址

	// 签名
	strSign := service.PaySignMd5StringNoKey(dictionary, Config.Md5Key)
	dictionary["sign"] = strSign
	return dictionary
}

// 提现组装参数
func (c *MMPayConfig) WithdrawCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)

	phone := service.GenerateMJLPhone()
	dictionary["mer_no"] = c.Merchant
	dictionary["order_no"] = order.OrderID
	dictionary["amount"] = fmt.Sprintf("%.2f", float32(order.Amount)/100)
	dictionary["currency"] = "BDT"
	dictionary["bank_code"] = getWallet(order.Bank)
	dictionary["name"] = order.RealName
	dictionary["email"] = fmt.Sprintf("%s@gmail.com", phone)
	dictionary["account"] = order.BankNumber
	dictionary["notify_url"] = Config.WithDrawNotifyUrl // 提现后回调地址

	// 签名
	strSign := service.PaySignMd5StringNoKey(dictionary, c.Md5Key)
	// 再加密
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

// =========================实现接口方法=========================

func (c *MMPayConfig) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	result := new(BaseRespone)
	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		return "", err
	}

	if result.Code != 200 || !result.Success {
		return "", fmt.Errorf("pay submit fail, %s, code: %d", order.OrderID, result.Code)
	}

	param := result.Data
	if param == nil {
		return "", errors.New("pay link is null")
	}

	if url, ok := param["order_data"]; ok && url != nil {
		return url.(string), nil
	}
	return "", errors.New("pay link is null")
}

func (c *MMPayConfig) ParsePayResult(body []byte) (map[string]string, error) {
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

func (c *MMPayConfig) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	result := new(BaseRespone)

	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		order.OrderStatus = data.DeliveryFailTransfer
		return false, fmt.Sprintf("withdraw Unmarshal fail, %s", order.OrderID)
	}

	if result.Code != 200 || !result.Success {
		order.OrderStatus = data.DeliveryFailTransfer
		return false, fmt.Sprintf("withdraw code fail, %s", order.OrderID)
	}

	param := result.Data
	if param == nil {
		order.OrderStatus = data.OrderFail
		return false, fmt.Sprintf("withdraw submit fail, %s", order.OrderID)
	}

	if no, ok := param["sys_no"]; ok && no != nil {
		order.OutTradeNo = no.(string)
		order.OrderStatus = data.OrderSuccess
		return true, "200"
	}
	order.OrderStatus = data.OrderFail
	return false, fmt.Sprintf("sys_no is nil,%s", order.OrderID)
}

// 充值查单接口
func (t *MMPayConfig) InspectSubmit(order *entity.PayOrder) ([]byte, error) {
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
func (c *MMPayConfig) InspectResponse(body []byte, order *entity.PayOrder) error {
	param := make(map[string]string)
	err := jsoniter.Unmarshal(body, param)
	if err != nil {
		return err
	}

	if param["result_status"] != "success" {
		return fmt.Errorf("status:%s", param["result_status"])
	}
	return nil
}

// 充值查单
func (c *MMPayConfig) InspectCompose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["mer_no"] = c.Merchant
	dictionary["order_no"] = order.OrderID

	// 签名
	strSign := service.PaySignMd5StringNoKey(dictionary, c.Md5Key)
	dictionary["sign"] = strSign

	return dictionary
}

// 提现查单接口
func (t *MMPayConfig) InspectWithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
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
func (c *MMPayConfig) InspectWdResponse(body []byte, order *entity.WithdrawOrder) int {
	param := make(map[string]string)
	err := jsoniter.Unmarshal(body, param)
	if err != nil {
		return 3
	}
	if param["result_status"] == "success" {
		return 2
	}
	return 3
}

// 提现查单
func (c *MMPayConfig) InspectWdCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["mer_no"] = c.Merchant
	dictionary["order_no"] = order.OrderID

	// 签名
	strSign := service.PaySignMd5StringNoKey(dictionary, c.Md5Key)
	dictionary["sign"] = strSign

	return dictionary
}

// 余额查询
func (c *MMPayConfig) BalanceQuery() int64 {
	param := c.QueryCompose()
	strReq, _ := jsoniter.Marshal(param)
	resp, err := doHttpForm(c.QueryUrl, strReq)
	if err != nil {
		glog.Info("[mmpay] BalanceQuery order fail, err,", err)
		return 0
	}

	rsp := new(MMBalanceRespond)
	err = jsoniter.Unmarshal(resp, &rsp)
	if err != nil {
		glog.Info("[mmpay] BalanceQuery order fail, err,", err)
		return 0
	}

	reslut := rsp.Data

	if reslut.MerNo != c.Merchant {
		glog.Info("[mmpay] BalanceQuery order fail, MerNo:", reslut.MerNo)
		return 0
	}
	glog.Infof("[mmpay] BalanceQuery order success")

	var balance float64
	for _, l := range reslut.Data {
		if l.Currency == "BDT" {
			b, _ := strconv.ParseFloat(l.Balance, 64)
			balance += b
		}
	}
	return int64(balance * 100)
}

func (c *MMPayConfig) QueryCompose() map[string]string {
	dictionary := map[string]string{
		"mer_no": c.Merchant,
	}

	strSign := service.PaySignMd5StringNoKey(dictionary, c.Md5Key)
	// dictionary["sign"] = pay.ToUpper(strSign)
	dictionary["sign"] = strSign

	return dictionary
}
