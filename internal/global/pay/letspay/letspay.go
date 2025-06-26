package letspay

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"goserver/internal/global/pay/entity"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"io"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var Config *LetsPayConfig

// Submit 充值下单
func (t *LetsPayConfig) Submit(order *entity.PayOrder) ([]byte, error) {
	body := t.compose(order) //打包post数据
	strReq := ToQueryString(body)
	strpar, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strpar)
	resp, err := doHttpForm(t.PayOrderUrl, strReq)
	if err != nil {
		return nil, err
	}
	glog.Infof("Submit order success, id:%s", order.OrderID)
	return resp, nil
}

// 提现下单
func (t *LetsPayConfig) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	body := t.WithdrawCompose(order) //打包post数据
	strReq := ToQueryString(body)
	strpar, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strpar)
	resp, err := doHttpForm(t.WithDrawUrl, strReq)
	//fmt.Printf("resp %s, err %v", string(resp), err)
	if err != nil {
		return nil, err
	}
	glog.Infof("WithdrawSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

// 充值响应
func (t *LetsPayConfig) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	result := new(LetsPayOrderRespond)
	err := json.Unmarshal(body, result)
	if err != nil {
		return "", err
	}
	if result.RetCode != "SUCCESS" || result.PayUrl == "" { // 没成功TODO
		return "", fmt.Errorf("code:%s, message:%s", result.Code, result.RetMsg)
	}
	order.OutTradeNo = result.PlatOrder
	return result.PayUrl, nil
}

// 提现响应
func (t *LetsPayConfig) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	result := new(LetsPayWithdrawRespond)
	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		return false, "parse json fail"
	}
	if result.RetCode == "SUCCESS" { // 没成功TODO
		// 发条消息提单成功
		// glog.Infof("[withdraw fail] submit success. order %s, user:%s", order.OrderID, order.Userid)
		order.OrderStatus = data.OrderSuccess
		return true, result.RetCode
	}
	glog.Infof("[lets withdraw] submit fail. code:%s, msg:%s", result.RetCode, result.RetMsg)
	order.OrderStatus = data.OrderFail
	return false, result.RetCode

}

func (t *LetsPayConfig) ParsePayResult(body []byte) (map[string]string, error) {
	reslut := make(map[string]string)

	str := string(body)
	params := strings.Split(str, "&")
	for _, v := range params {
		s := strings.Split(v, "=")
		if len(s) < 2 {
			continue
		}
		reslut[s[0]] = s[1]
	}

	return reslut, nil
}

// 充值组装参数
func (t *LetsPayConfig) compose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["mchId"] = t.Appid
	dictionary["orderNo"] = order.OrderID
	dictionary["amount"] = fmt.Sprintf("%.2f", float64(order.Amount)/100)
	dictionary["product"] = "indiaupi"
	dictionary["bankcode"] = "all"
	dictionary["goods"] = fmt.Sprintf("email:%s/name:%s/phone:%s/ip:%s", order.Email, order.RealName, order.Mobile, order.RegistIp)
	dictionary["returnUrl"] = "https://www.google.ocm"
	dictionary["notifyUrl"] = t.PayNotifyUrl

	// 签名
	strSign := PaySign(dictionary, t.Md5Key)
	dictionary["sign"] = strSign

	return dictionary
}

// 提现组装参数
func (t *LetsPayConfig) WithdrawCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["mchId"] = t.Appid
	dictionary["mchTransNo"] = order.OrderID
	dictionary["type"] = "api"
	dictionary["amount"] = fmt.Sprintf("%.2f", float64(order.Amount)/100)
	dictionary["notifyUrl"] = t.WithDrawNotifyUrl
	dictionary["accountName"] = order.RealName
	dictionary["accountNo"] = order.BankNumber
	dictionary["bankCode"] = order.IFSC
	dictionary["remarkInfo"] = fmt.Sprintf("email:%s/phone:%s/mode:bank", order.Email, order.Mobile)

	// 签名
	strSign := PaySign(dictionary, t.Md5Key)
	dictionary["sign"] = strSign
	return dictionary
}

// 充值查单接口
func (t *LetsPayConfig) InspectSubmit(order *entity.PayOrder) ([]byte, error) {
	body := t.InspectCompose(order) //打包post数据
	strReq := ToQueryString(body)
	resp, err := doHttpForm(t.InspectOrderUrl, strReq)
	if err != nil {
		return nil, err
	}
	glog.Infof("InspectSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

// 充值查单响应
func (c *LetsPayConfig) InspectResponse(body []byte, order *entity.PayOrder) error {
	result := new(LetsPayInspectRespond)
	jsoniter.Unmarshal(body, result)
	if result.Status != 2 { // 没成功TODO
		return fmt.Errorf("code:%d, message:%s", result.Status, result.RetCode)
	}
	return nil
}

// 充值查单
func (t *LetsPayConfig) InspectCompose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["mchId"] = t.Appid
	dictionary["orderNo"] = fmt.Sprint(order.OrderID)

	// 签名
	strSign := PaySign(dictionary, t.Md5Key)
	dictionary["sign"] = strSign

	return dictionary
}

// 提现查单接口
func (t *LetsPayConfig) InspectWithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	body := t.InspectWdCompose(order) //打包post数据
	strReq := ToQueryString(body)
	resp, err := doHttpForm(t.InspectWithdrawUrl, strReq)
	if err != nil {
		return nil, err
	}
	glog.Infof("InspectWithdrawSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

// 提现查单响应
func (c *LetsPayConfig) InspectWdResponse(body []byte, order *entity.WithdrawOrder) int {
	result := new(LetsPayInspectWdRespond)
	jsoniter.Unmarshal(body, result)
	if result.Status == 2 { // 没成功TODO
		return 2
	}
	return 3
}

// 提现查单
func (t *LetsPayConfig) InspectWdCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["mchId"] = t.Appid
	dictionary["mchTransNo"] = fmt.Sprint(order.OrderID)

	// 签名
	strSign := PaySign(dictionary, t.Md5Key)
	dictionary["sign"] = strSign

	return dictionary
}

// 余额查询
func (t *LetsPayConfig) BalanceQuery() int64 {
	param := t.QueryCompose()
	strReq := ToQueryString(param)
	resp, err := doHttpForm(t.QueryUrl, strReq)
	if err != nil {
		glog.Infof("[letspay] BalanceQuery order fail, err:", err)
		return 0
	}

	rsp := make(map[string]any)
	err = jsoniter.Unmarshal(resp, &rsp)
	if err != nil {
		glog.Infof("[letspay] BalanceQuery order fail, err:", err)
		return 0
	}

	if ret, ok := rsp["retCode"]; ok {
		if ret.(string) == "SUCCESS" {
			glog.Infof("[letspay] BalanceQuery order success")
			b := rsp["balance"].(string)
			return int64(utils.Float64(b) * 100)
		}
	}

	return 0
}

func (t *LetsPayConfig) QueryCompose() map[string]string {
	dictionary := map[string]string{
		"mchId": t.Appid,
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
func doHttpForm(targetUrl string, body string) ([]byte, error) {
	req, err := http.NewRequest("POST", targetUrl, strings.NewReader(body))
	if err != nil {
		glog.Errorf("err:%s", err)
		return []byte(""), err
	}
	req.Header.Add("Content-type", "application/x-www-form-urlencoded;charset=UTF-8")

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

func PaySign(param map[string]string, md5key string) string {
	keys := make([]string, 0, len(param))
	for k := range param {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var result strings.Builder
	for _, v := range keys {
		if val, ok := param[v]; ok && val != "" {
			vStr := fmt.Sprintf("%v", val)
			result.WriteString(fmt.Sprintf("%s=%s&", v, vStr))
		}
	}

	if md5key != "" {
		result.WriteString(fmt.Sprintf("key=%s", md5key))
	}

	md5 := utils.Md5(result.String())
	return strings.ToUpper(md5)
}
