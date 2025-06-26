package shpay

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
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var Config *ShPayConfig

// Submit 下单
func (t *ShPayConfig) Submit(order *entity.PayOrder) ([]byte, error) {
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
func (t *ShPayConfig) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
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
func (c *ShPayConfig) compose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)

	// phone := service.GenerateMJLPhone()
	dictionary["mchtId"] = c.Merchant
	dictionary["appId"] = c.AppId
	dictionary["countryCode"] = "BD"
	dictionary["requestTime"] = utils.String(utils.BsonNow().Unix())
	dictionary["notifyUrl"] = Config.PayNotifyUrl // 充值后返回地址
	dictionary["signType"] = "MD5"
	dictionary["outTradeNo"] = order.OrderID
	dictionary["transAmt"] = fmt.Sprintf("%d", order.Amount/100)
	dictionary["subject"] = "PAY " + fmt.Sprintf("%d", order.Amount/100)

	// 签名
	strSign := service.PaySignMd5StringNoKey(dictionary, Config.Md5Key)
	dictionary["sign"] = strings.ToUpper(strSign)
	return dictionary
}

// 提现组装参数
func (c *ShPayConfig) WithdrawCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["mchtId"] = c.Merchant
	dictionary["appId"] = c.AppId
	dictionary["countryCode"] = "BD"
	dictionary["requestTime"] = utils.Time2Str(utils.BsonNow())
	dictionary["notifyUrl"] = Config.WithDrawNotifyUrl // 充值后返回地址
	dictionary["signType"] = "MD5"
	dictionary["outTradeNo"] = order.OrderID
	dictionary["transAmt"] = fmt.Sprintf("%d", order.Amount/100)
	dictionary["subject"] = "withdraw " + fmt.Sprintf("%d", order.Amount/100)
	dictionary["walletNo"] = order.BankNumber
	dictionary["walletType"] = getWallet(order.Bank)

	// 签名
	strSign := service.PaySignMd5StringNoKey(dictionary, c.Md5Key)
	// 再加密
	dictionary["sign"] = strings.ToUpper(strSign)
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

func doHttpGet(targetUrl string) ([]byte, error) {
	req, err := http.NewRequest("GET", targetUrl, nil)
	if err != nil {
		glog.Errorf("err:%s", err)
		return []byte(""), err
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

func (c *ShPayConfig) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	result := new(BaseRespone)
	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		return "", err
	}

	if !result.Success {
		return "", fmt.Errorf("pay submit fail, %s, code: %d", order.OrderID, result.Code)
	}

	param := result.Data
	if param == nil {
		return "", errors.New("pay link is null")
	}

	if url, ok := param["link"]; ok && url != nil {
		return url.(string), nil
	}
	return "", errors.New("pay link is null")
}

func (c *ShPayConfig) ParsePayResult(body []byte) (map[string]string, error) {
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

func (c *ShPayConfig) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	result := new(BaseRespone)

	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		order.OrderStatus = data.DeliveryFailTransfer
		return false, fmt.Sprintf("withdraw Unmarshal fail, %s", order.OrderID)
	}

	if !result.Success {
		order.OrderStatus = data.DeliveryFailTransfer
		return false, fmt.Sprintf("withdraw code fail, %s", order.OrderID)
	}

	param := result.Data
	if param == nil {
		order.OrderStatus = data.OrderFail
		return false, fmt.Sprintf("withdraw submit fail, %s", order.OrderID)
	}

	if no, ok := param["transNo"]; ok && no != nil {
		order.OrderStatus = data.OrderSuccess
		order.OutTradeNo = no.(string)
		return true, "200"
	}
	order.OrderStatus = data.OrderFail
	return false, fmt.Sprintf("sys_no is nil,%s", order.OrderID)
}

// 充值查单接口
func (t *ShPayConfig) InspectSubmit(order *entity.PayOrder) ([]byte, error) {
	body := t.InspectCompose(order)
	resp, err := doHttpGet(t.InspectOrderUrl + "?" + body)
	if err != nil {
		return nil, err
	}
	glog.Infof("InspectSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

// 充值查单响应
func (c *ShPayConfig) InspectResponse(body []byte, order *entity.PayOrder) error {
	result := new(BaseRespone)

	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		return err
	}

	if !result.Success {
		return errors.New("response fail")
	}

	param := result.Data

	if param["result_status"] != "success" {
		return fmt.Errorf("status:%s", param["result_status"])
	}
	return nil
}

// 充值查单
func (c *ShPayConfig) InspectCompose(order *entity.PayOrder) string {
	var result strings.Builder
	dictionary := make(map[string]string)

	dictionary["mchtId"] = c.Merchant
	result.WriteString(fmt.Sprintf("mchtId=%s&", c.Merchant))
	dictionary["appId"] = c.AppId
	result.WriteString(fmt.Sprintf("appId=%s&", c.AppId))
	dictionary["requestTime"] = utils.Time2Str(utils.BsonNow())
	result.WriteString(fmt.Sprintf("requestTime=%s&", utils.Time2Str(utils.BsonNow())))
	dictionary["signType"] = "MD5"
	result.WriteString("signType=MD5&")
	dictionary["outTradeNo"] = order.OrderID
	result.WriteString(fmt.Sprintf("outTradeNo=%s&", order.OrderID))

	// 签名
	strSign := service.PaySignMd5StringNoKey(dictionary, c.Md5Key)
	dictionary["sign"] = strings.ToUpper(strSign)
	result.WriteString(fmt.Sprintf("sign=%s", strings.ToUpper(strSign)))

	return result.String()
}

// 提现查单接口
func (t *ShPayConfig) InspectWithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	// body := t.InspectWdCompose(order) //打包post数据
	// strReq, _ := jsoniter.Marshal(body)
	// resp, err := doHttpForm(t.InspectWithdrawUrl, strReq)
	// if err != nil {
	// 	return nil, err
	// }
	// glog.Infof("InspectWithdrawSubmit order success, id:%s", order.OrderID)
	return []byte{}, nil
}

// 提现查单响应
func (c *ShPayConfig) InspectWdResponse(body []byte, order *entity.WithdrawOrder) int {
	// result := new(BaseRespone)

	// err := jsoniter.Unmarshal(body, result)
	// if err != nil {
	// 	return 3
	// }

	// if !result.Success {
	// 	return 3
	// }

	// param := result.Data

	// if param["result_status"] != "success" {
	// 	return 3
	// }
	return 3
}

// 提现查单
func (c *ShPayConfig) InspectWdCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["mchtId"] = c.Merchant
	dictionary["appId"] = c.AppId
	dictionary["requestTime"] = utils.Time2Str(utils.BsonNow())
	dictionary["signType"] = "MD5"
	dictionary["outTradeNo"] = order.OrderID

	// 签名
	strSign := service.PaySignMd5StringNoKey(dictionary, c.Md5Key)
	dictionary["sign"] = strings.ToUpper(strSign)

	return dictionary
}

// 余额查询
func (c *ShPayConfig) BalanceQuery() int64 {
	param := c.QueryCompose()
	strReq, _ := jsoniter.Marshal(param)
	resp, err := doHttpForm(c.QueryUrl, strReq)
	if err != nil {
		glog.Info("[shpay] BalanceQuery order fail, err,", err)
		return 0
	}

	rsp := new(BaseRespone)
	err = jsoniter.Unmarshal(resp, &rsp)
	if err != nil {
		glog.Info("[shpay] BalanceQuery order fail, err,", err)
		return 0
	}

	if !rsp.Success {
		glog.Info("[shpay] BalanceQuery order fail, message:", rsp.Message)
		return 0
	}
	glog.Infof("[shpay] BalanceQuery order success")

	var balance float64
	if b, ok := rsp.Data["availableAmt"]; ok {
		balance += b.(float64)
	}
	return int64(balance * 100)
}

func (c *ShPayConfig) QueryCompose() map[string]string {
	dictionary := make(map[string]string)

	dictionary["mchtId"] = c.Merchant
	dictionary["appId"] = c.AppId
	dictionary["requestTime"] = utils.Time2Str(utils.BsonNow())
	dictionary["signType"] = "MD5"

	strSign := service.PaySignMd5StringNoKey(dictionary, c.Md5Key)
	// dictionary["sign"] = pay.ToUpper(strSign)
	dictionary["sign"] = strings.ToUpper(strSign)

	return dictionary
}
