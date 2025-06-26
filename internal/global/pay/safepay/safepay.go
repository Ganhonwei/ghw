package safepay

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
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
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var Config *SafePayConfig

func (c *SafePayConfig) Submit(order *entity.PayOrder) ([]byte, error) {
	body := c.compose(order) //打包post数据
	strpar, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strpar)
	resp, err := doHttpPostJson(c.ApiUrl, strpar)
	if err != nil {
		return nil, err
	}
	glog.Infof("safepay Submit order success, id:%s", order.OrderID)
	return resp, nil
}

func (c *SafePayConfig) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	result := new(SafePayOrderRespond)
	err := json.Unmarshal(body, result)
	if err != nil {
		return "", err
	}
	// 没成功
	if result.Status != "success" || result.OrderData == "" {
		return "", fmt.Errorf("code:%s, message:%s", result.Status, result.StatusMes)
	}
	order.OutTradeNo = result.OrderNo
	return result.OrderData, nil
}

func (c *SafePayConfig) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	if !checkWallet(order.Bank) {
		return nil, errors.New("wallet format error")
	}

	body := c.withdrawCompose(order) //打包post数据
	strReq, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strReq)
	resp, err := doHttpPostJson(c.ApiUrl, strReq)
	if err != nil {
		return nil, err
	}
	glog.Infof("safepay WithdrawSubmit order success, id:%s, resp:%s", order.OrderID, string(resp))
	return resp, nil
}

func (c *SafePayConfig) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	result := new(SafePayWithdrawOrderRespond)
	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		return false, "safepay parse json fail"
	}
	if result.Status == "success" { // 成功
		// 发条消息提单成功
		order.OrderStatus = data.OrderSuccess
		glog.Infof("[safepay withdraw] submit success. order:%s, user:%s, sys_no:%s", order.OrderID, order.Userid, result.SysNo)
		return true, result.Status
	}
	// 没成功
	glog.Infof("[safepay withdraw] submit fail. code:%s, msg:%s, sys_no=%s", result.Status, result.StatusMes, result.SysNo)
	order.OrderStatus = data.OrderFail
	return false, fmt.Sprintf("code:%s, message:%s", result.Status, result.StatusMes)
}

// 待收订单查询
func (c *SafePayConfig) InspectSubmit(order *entity.PayOrder) ([]byte, error) {
	dictionary := make(map[string]string)
	dictionary["mer_no"] = c.Merno
	dictionary["order_no"] = order.OrderID
	dictionary["method"] = "trade.check"
	// 签名
	strSign := PaySign(dictionary, c.Key)
	dictionary["sign"] = strSign
	strpar, err := jsoniter.Marshal(dictionary)
	if err != nil {
		return nil, err
	}
	resp, err := doHttpPostJson(c.ApiUrl, strpar)
	if err != nil {
		return nil, err
	}
	glog.Infof("InspectSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

func (c *SafePayConfig) InspectResponse(body []byte, order *entity.PayOrder) error {
	result := new(InspectRespond)
	if err := jsoniter.Unmarshal(body, result); err != nil {
		return err
	}
	if result.CheckStatus != "success" { // 没成功
		return fmt.Errorf("code:%s, message:%s", result.CheckStatus, result.StatusMes)
	}
	if result.ResultStatus != "success" {
		return fmt.Errorf("state:%s, message:%s", result.ResultStatus, result.StatusMes)
	}
	return nil
}

func (c *SafePayConfig) InspectWithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	dictionary := make(map[string]string)
	dictionary["mer_no"] = c.Merno
	dictionary["order_no"] = order.OrderID
	dictionary["method"] = "fund.apply.check"
	// 签名
	strSign := PaySign(dictionary, c.Key)
	dictionary["sign"] = strSign

	strpar, _ := jsoniter.Marshal(dictionary)
	resp, err := doHttpPostJson(c.ApiUrl, strpar)
	if err != nil {
		return nil, err
	}
	glog.Infof("InspectWithdrawSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

func (c *SafePayConfig) InspectWdResponse(body []byte, order *entity.WithdrawOrder) int {
	result := new(InspectWithdrawRespond)
	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		glog.Error("InspectWdResponse json unmarshal error:", err, string(body))
		return 3
	}
	if result.CheckStatus != "success" { // 没成功
		return 3
	}
	if result.ResultStatus != "success" {
		return 3
	}
	return 2
}

func (c *SafePayConfig) BalanceQuery() int64 {
	dictionary := make(map[string]string)
	dictionary["mer_no"] = c.Merno
	dictionary["method"] = "fund.query"
	// 签名
	strSign := PaySign(dictionary, c.Key)
	dictionary["sign"] = strSign

	strpar, _ := jsoniter.Marshal(dictionary)
	resp, err := doHttpPostJson(c.ApiUrl, strpar)
	if err != nil {
		glog.Info("[safepay] BalanceQuery order fail, err,", err)
		return 0
	}

	rsp := new(SafePayBalanceRespond)
	err = jsoniter.Unmarshal(resp, &rsp)
	if err != nil {
		glog.Info("[safepay] BalanceQuery order fail, err,", err)
		return 0
	}

	if rsp.CheckStatus != "success" {
		glog.Info("[safepay] BalanceQuery order fail, code:", rsp.CheckStatus)
		return 0
	}
	glog.Infof("[safepay] BalanceQuery order success, %#v", rsp.Currency)

	var balance float64
	for _, l := range rsp.Currency {
		b, _ := strconv.ParseFloat(l, 64)
		balance += b
	}
	return int64(balance * 100)
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
func (c *SafePayConfig) compose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)

	wallets := []string{"25001", "25003", "25004", "25006"} // 孟加拉钱包 孟加拉UPAY 孟加拉NAGAD 孟加拉BKASH("25005") 孟加拉ROCKET
	wallet, _ := utils.ChoiceString(wallets)
	if wallet == "" {
		wallet = "25001"
	}

	phone := service.GenerateMJLPhone()
	dictionary["mer_no"] = c.Merno
	dictionary["order_no"] = order.OrderID
	dictionary["order_amount"] = fmt.Sprintf("%.2f", float64(order.Amount)/100)
	dictionary["payname"] = order.RealName
	dictionary["payemail"] = fmt.Sprintf("%s@gmail.com", phone)
	dictionary["payphone"] = phone
	dictionary["currency"] = "BDT"
	dictionary["paytypecode"] = wallet
	dictionary["method"] = "trade.create"
	dictionary["returnurl"] = c.PayReturnUrl

	// 签名
	strSign := PaySign(dictionary, c.Key)
	dictionary["sign"] = strSign

	return dictionary
}

// 提现组装参数
func (c *SafePayConfig) withdrawCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["mer_no"] = c.Merno
	dictionary["order_no"] = order.OrderID
	dictionary["method"] = "fund.apply"
	dictionary["order_amount"] = fmt.Sprintf("%.2f", float64(order.Amount)/100)
	dictionary["currency"] = "BDT"
	dictionary["acc_code"] = getWallet(order.Bank)
	dictionary["acc_name"] = order.RealName
	dictionary["acc_no"] = order.BankNumber
	dictionary["returnurl"] = c.WithdrawReturnUrl

	// 签名
	strSign := PaySign(dictionary, c.Key)
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
	for i, v := range keys {
		if val, ok := param[v]; ok && val != "" {
			if i == 0 {
				result.WriteString(fmt.Sprintf("%s=%s", v, val))
			} else {
				result.WriteString(fmt.Sprintf("&%s=%s", v, val))
			}
		}
	}

	if md5key != "" {
		result.WriteString(md5key)
	} else {
		glog.Errorf("safepay md5 key is empty... order: %s", param["orderNo"])
	}

	a := result.String()
	_ = a

	md5 := utils.Md5(result.String())
	return strings.ToLower(md5)
}

func (c *SafePayConfig) ParsePayResult(body []byte) (map[string]string, error) {
	// SafePayRechargeCallback
	reslut := make(map[string]string)
	err := jsoniter.Unmarshal(body, &reslut)
	if err != nil {
		return nil, err
	}
	return reslut, nil
}
