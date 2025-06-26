package unipay

import (
	"bytes"
	"crypto/tls"
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

var Config *UniPayConfig

// Submit 下单
func (t *UniPayConfig) Submit(order *entity.PayOrder) ([]byte, error) {
	body := t.compose(order) //打包post数据
	order.RequestParam = string(body)
	resp, err := t.doHttpForm(t.PayOrderUrl, body)
	if err != nil {
		return nil, err
	}
	glog.Infof("Submit order success, id:%s", order.OrderID)
	return resp, nil
}

// 提现下单
func (t *UniPayConfig) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	if !checkWallet(order.Bank) {
		return nil, fmt.Errorf("unsupport wallet")
	}
	body := t.WithdrawCompose(order) //打包post数据
	// strReq, _ := jsoniter.Marshal(body)
	order.RequestParam = string(body)
	resp, err := t.doHttpForm(t.WithDrawUrl, body)
	if err != nil {
		return nil, err
	}
	glog.Infof("WithdrawSubmit order success, id:%s, resp:%s", order.OrderID, string(resp))
	return resp, nil
}

// 组装参数
func (c *UniPayConfig) compose(order *entity.PayOrder) []byte {
	type Data struct {
		VendorTranID string `json:"VendorTranID"`
		Amount       string `json:"Amount"`
		Player       string `json:"Player"`
		NotifyURL    string `json:"NotifyURL"`
	}
	type Param struct {
		Sign string `json:"sign"`
		Data Data   `json:"data"`
	}

	dictionary := make(map[string]any)

	dictionary["Amount"] = fmt.Sprintf("%d", order.Amount/100)
	dictionary["Player"] = order.Userid
	dictionary["NotifyURL"] = c.PayNotifyUrl
	dictionary["VendorTranID"] = order.OrderID

	// p, _ := jsoniter.Marshal(dictionary)
	// 签名
	strSign := service.PaySignMd5(dictionary, c.Appid)
	param := Param{
		Sign: strSign,
		Data: Data{
			VendorTranID: order.OrderID,
			Amount:       fmt.Sprintf("%d", order.Amount/100),
			Player:       order.Userid,
			NotifyURL:    c.PayNotifyUrl,
		},
	}

	reslut, _ := jsoniter.Marshal(param)
	return reslut
}

// 提现组装参数
func (c *UniPayConfig) WithdrawCompose(order *entity.WithdrawOrder) []byte {
	type Data struct {
		VendorTranID    string `json:"VendorTranID"`
		Amount          string `json:"Amount"`
		Player          string `json:"Player"`
		NotifyURL       string `json:"NotifyURL"`
		PlayerWalletNum string `json:"PlayerWalletNum"`
		WalletNameID    int    `json:"WalletNameID"`
	}
	type Param struct {
		Sign string `json:"sign"`
		Data Data   `json:"data"`
	}

	dictionary := make(map[string]any)
	dictionary["Amount"] = fmt.Sprintf("%d", order.Amount/100)
	dictionary["Player"] = order.Userid
	dictionary["NotifyURL"] = c.WithDrawNotifyUrl
	dictionary["VendorTranID"] = order.OrderID
	dictionary["PlayerWalletNum"] = order.BankNumber
	dictionary["WalletNameID"] = getWalletId(order.Bank)

	// p, _ := jsoniter.Marshal(dictionary)
	// 签名
	strSign := service.PaySignMd5(dictionary, c.Appid)
	param := Param{
		Sign: strSign,
		Data: Data{
			VendorTranID:    order.OrderID,
			Amount:          fmt.Sprintf("%d", order.Amount/100),
			Player:          order.Userid,
			NotifyURL:       c.WithDrawNotifyUrl,
			PlayerWalletNum: order.BankNumber,
			WalletNameID:    getWalletId(order.Bank),
		},
	}

	reslut, _ := jsoniter.Marshal(param)
	return reslut
}

// application/json PSOT请求
func (c *UniPayConfig) doHttpForm(targetUrl string, body []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", targetUrl, bytes.NewReader(body))
	if err != nil {
		glog.Errorf("err:%s", err)
		return []byte(""), err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("NAME", c.Name)
	req.Header.Add("APPID", c.Appid)

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

// GET请求
func (c *UniPayConfig) doHttpGet(targetUrl string) ([]byte, error) {
	req, err := http.NewRequest("GET", targetUrl, nil)
	if err != nil {
		glog.Errorf("err:%s", err)
		return []byte(""), err
	}

	req.Header.Add("NAME", c.Name)
	req.Header.Add("APPID", c.Appid)

	client := &http.Client{}

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

func (c *UniPayConfig) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	res := make(map[string]any)
	err := jsoniter.Unmarshal(body, &res)
	if err != nil {
		return "", err
	}

	var status float64 = 0
	if s, ok := res["status"]; ok {
		status = s.(float64)
	}

	if status != 1 || res["url"] == nil || res["url"] == "" { // 没成功TODO
		return "", fmt.Errorf("code:%f,message:%s, OrderNo:%s", status, res["message"], order.OrderID)
	}
	return res["url"].(string), nil
}

func (c *UniPayConfig) ParsePayResult(body []byte) (map[string]string, error) {
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

func (c *UniPayConfig) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	res := make(map[string]any)
	err := jsoniter.Unmarshal(body, &res)
	if err != nil {
		order.OrderStatus = data.DeliveryFailTransfer
		return false, "Unmarshal fail"
	}

	var status float64 = 0
	if s, ok := res["status"]; ok {
		status = s.(float64)
	}

	if status != 1 { // 没成功TODO
		order.OrderStatus = data.OrderFail
		return false, fmt.Sprintf("code:%f,message:%s, OrderNo:%s", status, res["message"], order.OrderID)
	}

	order.OrderStatus = data.OrderSuccess
	return true, ""
}

// 充值查单接口
func (t *UniPayConfig) InspectSubmit(order *entity.PayOrder) ([]byte, error) {
	url := fmt.Sprintf("%s?orderId=%s&orderType=ds", t.InspectOrderUrl, order.OrderID)
	resp, err := t.doHttpGet(url)
	if err != nil {
		return nil, err
	}
	glog.Infof("InspectSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

// 充值查单响应
func (c *UniPayConfig) InspectResponse(body []byte, order *entity.PayOrder) error {
	result := new(UniPayInspectRespond)
	jsoniter.Unmarshal(body, result)
	if result.Status != 1 { // 没成功TODO
		return fmt.Errorf("code:%d", result.Status)
	}
	if result.OrderStatus != 2 {
		return fmt.Errorf("status:%d, message:%s", result.OrderStatus, result.Message)
	}
	return nil
}

// 提现查单接口
func (t *UniPayConfig) InspectWithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	url := fmt.Sprintf("%s?orderId=%s&orderType=df", t.InspectWithdrawUrl, order.OrderID)
	resp, err := t.doHttpGet(url)
	if err != nil {
		return nil, err
	}
	glog.Infof("InspectWithdrawSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

// 提现查单响应
func (c *UniPayConfig) InspectWdResponse(body []byte, order *entity.WithdrawOrder) int {
	result := new(UniPayInspectRespond)
	jsoniter.Unmarshal(body, result)
	if result.Status != 1 { // 没成功TODO
		return 3
	}
	if result.OrderStatus != 2 {
		return 3
	}
	return 2
}

// 余额查询
func (c *UniPayConfig) BalanceQuery() int64 {
	resp, err := c.doHttpGet(c.QueryUrl)
	if err != nil {
		glog.Error("[twpay] BalanceQuery order fail, err,", err)
		return 0
	}

	reslut := make(map[string]any)
	err1 := jsoniter.Unmarshal(resp, &reslut)
	if err1 != nil {
		glog.Error("[twpay] BalanceQuery order fail, err,", err1)
		return 0
	}
	if b, ok := reslut["status"]; !ok || b != 1 {
		glog.Error("[twpay] BalanceQuery order fail")
		return 0
	}

	var balance float64 = 0
	if b, ok := reslut["balance"]; ok {
		balance = b.(float64)
	}
	glog.Infof("[twpay] BalanceQuery order success")
	return int64(balance * 100)
}
