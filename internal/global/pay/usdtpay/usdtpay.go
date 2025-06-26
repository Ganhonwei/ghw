package usdtpay

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"goserver/internal/global/pay/entity"
	"goserver/internal/global/pay/service"
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

var Config *UsdtPayConfig

// 充值订单
func (t *UsdtPayConfig) Submit(order *entity.PayOrder) ([]byte, error) {
	body := t.compose(order) //打包post数据
	strpar, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strpar)
	resp, err := doHttpPostJson(t.PayOrderUrl, strpar)
	if err != nil {
		return nil, err
	}
	glog.Infof("usdtpay Submit order success, id:%s", order.OrderID)
	return resp, nil
}

// 充值响应
func (t *UsdtPayConfig) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	result := new(UsdtPayOrderRespond)
	err := json.Unmarshal(body, result)
	if err != nil {
		return "", err
	}
	// 没成功
	if result.Code != 200 || result.Data.PaymentUrl == "" {
		return "", fmt.Errorf("code:%d, message:%s, payment=%s", result.Code, result.Msg, result.Data.PaymentUrl)
	}
	order.OutTradeNo = result.Data.TradeId
	return result.Data.PaymentUrl, nil
}

// 提现订单
func (t *UsdtPayConfig) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	body := t.withdrawCompose(order) //打包post数据
	strpar, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strpar)
	resp, err := doHttpPostJson(t.WithDrawUrl, strpar)
	if err != nil {
		return nil, err
	}
	glog.Infof("usdt WithdrawSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

func (t *UsdtPayConfig) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	result := new(UsdtPayWithdrawRespond)
	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		return false, "parse json fail"
	}
	if result.Code == 200 { // 成功
		// 发条消息提单成功
		glog.Infof("[usdt withdraw] submit success. order:%s, user:%s, traceId:%s", order.OrderID, order.Userid, result.Data.TradeId)
		return true, fmt.Sprint(result.Code)
	}
	// 没成功
	glog.Infof("[usdt withdraw] submit fail. code:%s, msg:%s", result.Code, result.Msg)
	return false, fmt.Sprintf("code:%d, message:%s", result.Code, result.Msg)
}

// 待收订单查询
func (t *UsdtPayConfig) InspectSubmit(order *entity.PayOrder) ([]byte, error) {

	return nil, nil
}

// 充值查单响应
func (t *UsdtPayConfig) InspectResponse(body []byte, order *entity.PayOrder) error {
	return errors.New("not support usdt inspect")
}

// 提现查单接口
func (t *UsdtPayConfig) InspectWithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	return nil, nil
}

func (t *UsdtPayConfig) InspectWdResponse(body []byte, order *entity.WithdrawOrder) int {
	// 2成功3没成功
	return 3
}

// 充值组装参数
func (t *UsdtPayConfig) compose(order *entity.PayOrder) map[string]any {
	dictionary := make(map[string]any)

	dictionary["order_id"] = order.OrderID
	dictionary["notify_url"] = t.PayNotifyUrl
	dictionary["redirect_url"] = t.RedirectUrl
	amountF := fmt.Sprintf("%.2f", float64(order.Amount)/100)
	amount, _ := strconv.ParseFloat(amountF, 64)
	dictionary["amount"] = amount

	// 签名
	strSign := PaySign(dictionary, t.Md5Key)
	dictionary["signature"] = strSign

	return dictionary
}

// 余额查询
func (t *UsdtPayConfig) BalanceQuery() int64 {
	return 0
}

// 提现组装参数
func (t *UsdtPayConfig) withdrawCompose(order *entity.WithdrawOrder) map[string]any {
	dictionary := make(map[string]any)

	dictionary["address"] = order.BankNumber // "TWm1vshXNSJEWYW4bQjeuubQEo9aafpxaH"
	dictionary["order_id"] = order.OrderID
	dictionary["notify_url"] = t.WithDrawNotifyUrl
	amountF := fmt.Sprintf("%.2f", float64(order.Amount)/100)
	amount, _ := strconv.ParseFloat(amountF, 64)
	dictionary["amount"] = amount

	// 签名
	strSign := PaySign(dictionary, t.Md5Key)
	dictionary["signature"] = strSign
	return dictionary
}

func PaySign(param map[string]any, md5key string) string {
	keys := make([]string, 0, len(param))
	for k := range param {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var result strings.Builder
	for i, v := range keys {
		if val, ok := param[v]; ok && val != "" {
			result.WriteString(fmt.Sprintf("%s=%v", v, val))
			if i < len(keys)-1 {
				result.WriteString("&")
			}
		}
	}

	if md5key != "" {
		result.WriteString(md5key)
	} else {
		glog.Errorf("usdtpay md5 key is empty... order: %s", param["orderNo"])
	}

	a := result.String()
	_ = a

	md5 := utils.Md5(result.String())
	return strings.ToLower(md5)
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

func (t *UsdtPayConfig) ParsePayResult(body []byte) (map[string]any, error) {
	tradeResult := UsdtPayRechargeCallback{}
	err := jsoniter.Unmarshal(body, &tradeResult)
	if err != nil {
		return nil, err
	}

	result := make(map[string]any)
	result["trade_id"] = tradeResult.TradeId
	result["order_id"] = tradeResult.OrderId
	result["amount"] = tradeResult.Amount
	result["actual_amount"] = tradeResult.ActualAmount
	result["token"] = tradeResult.Token
	result["block_transaction_id"] = tradeResult.BlockTransactionId
	result["signature"] = tradeResult.Signature
	result["status"] = fmt.Sprint(tradeResult.Status)
	return result, nil
}
