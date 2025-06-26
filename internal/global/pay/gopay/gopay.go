package gopay

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

var Config *GoPayConfig

func (c *GoPayConfig) Submit(order *entity.PayOrder) ([]byte, error) {
	body := c.compose(order) //打包post数据
	strpar, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strpar)
	resp, err := doHttpPostJson(c.PayUrl, strpar, [2]string{"account", c.Account})
	if err != nil {
		return nil, err
	}
	glog.Infof("gopay Submit order success, id:%s", order.OrderID)
	return resp, nil
}

func (c *GoPayConfig) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	result := new(GoPayOrderRespond)
	err := json.Unmarshal(body, result)
	if err != nil {
		return "", err
	}
	// 没成功
	if result.Code != 0 || result.Data.Url == "" {
		return "", fmt.Errorf("code:%d, message:%s", result.Code, result.Msg)
	}
	order.OutTradeNo = result.Data.TransactionId
	return result.Data.Url, nil
}

func (c *GoPayConfig) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	if !checkWallet(order.Bank) {
		return nil, errors.New("wallet format error")
	}

	body := c.withdrawCompose(order) //打包post数据
	strReq, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strReq)
	resp, err := doHttpPostJson(c.WithdrawUrl, strReq, [2]string{"account", c.Account})
	if err != nil {
		return nil, err
	}
	glog.Infof("gopay WithdrawSubmit order success, id:%s, resp:%s", order.OrderID, string(resp))
	return resp, nil
}

func (c *GoPayConfig) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	result := new(GoPayWithdrawOrderRespond)
	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		return false, "gopay parse json fail"
	}

	if result.Code == 0 && (result.Data.Status == 0 || result.Data.Status == 1) { // 成功
		// 发条消息提单成功
		glog.Infof("[gopay withdraw] submit success. order:%s, user:%s", order.OrderID, order.Userid)
		order.OrderStatus = data.OrderSuccess
		return true, result.Msg + "," + result.Data.ErrorMsg
	}
	// 没成功
	glog.Infof("[gopay withdraw] submit fail. code:%d, msg:%s", result.Code, result.Msg)
	order.OrderStatus = data.OrderFail
	return false, fmt.Sprintf("code:%d, message:%s %s", result.Code, result.Msg, result.Data.ErrorMsg)
}

// 待收订单查询
func (c *GoPayConfig) InspectSubmit(order *entity.PayOrder) ([]byte, error) {
	dictionary := make(map[string]any)
	dictionary["id"] = order.OrderID
	dictionary["merchantRef"] = order.OrderID
	dictionary["timestamp"] = time.Now().UnixMilli()
	// 签名
	strSign := PaySign(dictionary, c.Key)
	dictionary["sign"] = strSign
	strpar, err := jsoniter.Marshal(dictionary)
	if err != nil {
		return nil, err
	}
	resp, err := doHttpPostJson(c.InspectOrderUrl, strpar, [2]string{"account", c.Account})
	if err != nil {
		return nil, err
	}
	glog.Infof("InspectSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

func (c *GoPayConfig) InspectResponse(body []byte, order *entity.PayOrder) error {
	result := new(InspectRespond)
	if err := jsoniter.Unmarshal(body, result); err != nil {
		return err
	}
	if result.Code != 0 { // 没成功
		return fmt.Errorf("code:%d, message:%s", result.Code, result.Msg)
	}
	if result.Data.Status != 1 {
		return fmt.Errorf("status:%d, message:%s", result.Data.Status, result.Msg)
	}
	return nil
}

func (c *GoPayConfig) InspectWithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	dictionary := make(map[string]any)
	dictionary["id"] = order.OrderID
	dictionary["merchantRef"] = order.OrderID
	dictionary["timestamp"] = time.Now().UnixMilli()
	// 签名
	strSign := PaySign(dictionary, c.Key)
	dictionary["sign"] = strSign

	strpar, _ := jsoniter.Marshal(dictionary)
	resp, err := doHttpPostJson(c.InspectWithdrawUrl, strpar, [2]string{"account", c.Account})
	if err != nil {
		return nil, err
	}
	glog.Infof("InspectWithdrawSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

func (c *GoPayConfig) InspectWdResponse(body []byte, order *entity.WithdrawOrder) int {
	result := new(InspectWithdrawRespond)
	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		glog.Error("InspectWdResponse json unmarshal error:", err, string(body))
		return 3
	}
	if result.Code != 0 { // 没成功
		return 3
	}
	if result.Data.Status != 1 {
		return 3
	}
	return 2
}

func (c *GoPayConfig) BalanceQuery() int64 {
	dictionary := make(map[string]any)
	dictionary["timestamp"] = time.Now().UnixMilli()
	// 签名
	strSign := PaySign(dictionary, c.Key)
	dictionary["sign"] = strSign

	strpar, _ := jsoniter.Marshal(dictionary)
	resp, err := doHttpPostJson(c.QueryUrl, strpar, [2]string{"account", c.Account})
	if err != nil {
		glog.Info("[gopay] BalanceQuery order fail, err,", err)
		return 0
	}

	rsp := new(BalanceRespond)
	err = jsoniter.Unmarshal(resp, &rsp)
	if err != nil {
		glog.Info("[gopay] BalanceQuery order fail, err,", err)
		return 0
	}

	if rsp.Code != 0 {
		glog.Info("[gopay] BalanceQuery order fail, code:", rsp.Code)
		return 0
	}
	glog.Infof("[gopay] BalanceQuery order success, %#v", rsp.Data.Balance)

	balance, _ := strconv.ParseFloat(rsp.Data.Balance, 64)
	return int64(balance * 100)
}

// application/json PSOT请求
func doHttpPostJson(targetUrl string, body []byte, headers ...[2]string) ([]byte, error) {
	req, err := http.NewRequest("POST", targetUrl, bytes.NewReader(body))
	if err != nil {
		glog.Errorf("err:%s", err)
		return []byte(""), err
	}
	req.Header.Add("Content-type", "application/json")
	for _, header := range headers {
		req.Header.Add(header[0], header[1])
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
func (c *GoPayConfig) compose(order *entity.PayOrder) map[string]any {
	dictionary := make(map[string]any)

	wallets := []string{"Nagad", "bKash"}
	wallet, _ := utils.ChoiceString(wallets)
	if wallet == "" {
		wallet = "Nagad"
	}
	phone := service.GenerateMJLPhone()
	_ = phone
	dictionary["orderId"] = order.OrderID
	dictionary["amount"] = fmt.Sprintf("%.2f", float64(order.Amount)/100)
	dictionary["customerName"] = order.RealName
	dictionary["customerMobile"] = phone
	dictionary["walletType"] = wallet
	dictionary["notifyUrl"] = c.PayNotifyUrl
	dictionary["timestamp"] = time.Now().UnixMilli()

	// 签名
	strSign := PaySign(dictionary, c.Key)
	dictionary["sign"] = strSign

	return dictionary
}

// 提现组装参数
func (c *GoPayConfig) withdrawCompose(order *entity.WithdrawOrder) map[string]any {
	dictionary := make(map[string]any)

	dictionary["orderId"] = order.OrderID
	dictionary["amount"] = fmt.Sprintf("%.2f", float64(order.Amount)/100)
	dictionary["walletType"] = getWallet(order.Bank)
	dictionary["receiveUsername"] = order.RealName
	dictionary["receiveWalletNo"] = order.BankNumber
	dictionary["notifyUrl"] = c.WithdrawNotifyUrl
	dictionary["timestamp"] = time.Now().UnixMilli()

	// 签名
	strSign := PaySign(dictionary, c.Key)
	dictionary["sign"] = strSign

	return dictionary
}

func PaySign(param map[string]any, md5key string) string {
	keys := make([]string, 0, len(param))
	for k := range param {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var result strings.Builder
	for _, v := range keys {
		if val, ok := param[v]; ok {
			if str, ok := val.(string); ok && str == "" {
				continue
			}
			result.WriteString(fmt.Sprintf("%s=%v&", v, val))
		}
	}

	if md5key != "" {
		result.WriteString(fmt.Sprintf("key=%s", md5key))
	} else {
		glog.Errorf("gopay md5 key is empty... order: %s", param["orderId"])
	}

	md5 := utils.Md5(result.String())
	return strings.ToLower(md5)
}

func (c *GoPayConfig) ParsePayResult(body []byte) (map[string]any, error) {
	reslut := make(map[string]any)
	err := jsoniter.Unmarshal(body, &reslut)
	if err != nil {
		return nil, err
	}
	return reslut, nil
}
