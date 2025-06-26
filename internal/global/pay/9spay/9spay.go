package pay9s

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
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var Config *Pay9sConfig

// Submit 充值下单
func (t *Pay9sConfig) Submit(order *entity.PayOrder) ([]byte, error) {
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

// 充值响应
func (t *Pay9sConfig) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	result := make(map[string]any)
	jsoniter.Unmarshal(body, &result)
	outSign := ""
	if s, ok := result["sign"]; ok && s != nil {
		outSign = s.(string)
		delete(result, "sign")
	}
	// 验签
	sign := PaySign(result, t.Md5Key)
	if outSign != sign {
		return "", fmt.Errorf("[9spay] 下单验签失败,userid:%s result:%v", order.Userid, result)
	}

	code := result["code"]
	if code != "0000" {
		return "", fmt.Errorf("[9spay] code:%d, message:%s", code, result["msg"])
	}

	payUrl := result["payUrl"]
	if payUrl == nil || payUrl == "" {
		payUrl = result["payUrl2"]
	}

	if payUrl == nil || payUrl == "" {
		return "", fmt.Errorf("[9spay] no found paylink %s, code:%d, message:%s", order.Userid, code, result["msg"])
	}

	order.OutTradeNo = result["orderNo"].(string)

	return utils.String(payUrl), nil
}

// 提现下单
func (t *Pay9sConfig) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	body := t.WithdrawCompose(order) //打包post数据
	strReq := ToQueryString(body)
	strpar, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strpar)
	resp, err := doHttpForm(t.WithDrawUrl, strReq)
	//fmt.Printf("resp %s, err %v", string(resp), err)
	if err != nil {
		return nil, err
	}
	glog.Infof("[9spay] WithdrawSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

// 提现响应
func (t *Pay9sConfig) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	result := make(map[string]any)
	jsoniter.Unmarshal(body, &result)
	outSign := ""
	if s, ok := result["sign"]; ok && s != nil {
		outSign = s.(string)
		delete(result, "sign")
	}

	// 验签
	var msg string
	if m, ok := result["msg"]; ok && m != nil {
		msg = m.(string)
	}
	sign := PaySign(result, t.Md5Key)
	if outSign != sign {
		order.OrderStatus = data.OrderFail
		glog.Infof("[9spay withdraw] WithdrawSubmitResponse sign fail.")
		return false, msg
	}

	code := result["code"]
	if code != nil && code != "0000" {
		order.OrderStatus = data.OrderFail
		return false, msg
	}

	status := result["status"]
	if status != nil && status != "PROCESSED" {
		order.OrderStatus = data.OrderFail
		return false, msg
	}

	order.OrderStatus = data.OrderSuccess
	glog.Infof("[9spay withdraw] submit success. code:%s, msg:%s", code, msg)
	return true, utils.String(code)
}

// 充值查单接口
func (t *Pay9sConfig) InspectSubmit(order *entity.PayOrder) ([]byte, error) {
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
func (t *Pay9sConfig) InspectResponse(body []byte, order *entity.PayOrder) error {
	result := make(map[string]any)
	jsoniter.Unmarshal(body, result)
	outSign := ""
	if s, ok := result["sign"]; ok && s != nil {
		outSign = s.(string)
		delete(result, "sign")
	}
	// 验签
	sign := PaySign(result, t.Md5Key)
	if outSign != sign {
		return fmt.Errorf("[9spay] InspectResponse verify sign fail, %s", order.OrderID)
	}

	code := result["code"]

	if code == nil || code != "0000" {
		return fmt.Errorf("[9spay] InspectResponse code:%d, message:%s", code, result["msg"])
	}

	if result["status"] != "PAID" {
		return fmt.Errorf("state:%d, message:%s", result["status"], result["msg"])
	}
	// order.OutTradeNo = result.Data.POrderNo
	return nil
}

// 提现查单接口
func (t *Pay9sConfig) InspectWithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
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
func (t *Pay9sConfig) InspectWdResponse(body []byte, order *entity.WithdrawOrder) int {
	result := make(map[string]any)
	jsoniter.Unmarshal(body, result)
	outSign := ""
	if s, ok := result["sign"]; ok && s != nil {
		outSign = s.(string)
		delete(result, "sign")
	}
	// 验签
	sign := PaySign(result, t.Md5Key)
	if outSign != sign {
		return data.WithdrawFail
	}

	code := result["code"]

	if code == nil || code != "0000" {
		return data.WithdrawFail
	}

	status := result["status"]
	if status == nil {
		return data.WithdrawFail
	}

	if status == "PROCESSED" || status == "PAYING" || status == "UN_PAYD" {
		return data.Withdrawing
	}

	if status == "PAID" {
		return data.WithdrawSuccess
	}

	return data.WithdrawFail
}

// 余额查询
func (t *Pay9sConfig) BalanceQuery() int64 {
	param := map[string]any{
		"merchantId": t.Appid,
	}
	sign := PaySign(param, t.Md5Key)
	param["sign"] = sign
	strReq := ToQueryString(param)
	resp, err := doHttpForm(t.QueryUrl, strReq)
	if err != nil {
		glog.Info("[9spay] BalanceQuery order fail,err:", err)
		return 0
	}

	rsp := make(map[string]any)
	err = json.Unmarshal(resp, &rsp)
	if err != nil {
		glog.Info("[9spay] BalanceQuery order fail,err:", err)
		return 0
	}

	balance := rsp["totalBalance"]
	if balance == nil {
		return 0
	}

	b, _ := balance.(float64)
	glog.Infof("[9spay] BalanceQuery order success")
	return int64(b * 100)
}

// ======================================== 以上为接口的实现 ========================================

func (t *Pay9sConfig) ParsePayResult(body []byte) (map[string]interface{}, error) {
	reslut := make(map[string]interface{})
	str := string(body)
	str, _ = url.QueryUnescape(str)
	strs := strings.Split(str, "&")
	for _, v := range strs {
		vals := strings.Split(v, "=")
		reslut[vals[0]] = vals[1]
	}

	return reslut, nil
}

// 充值组装参数
func (t *Pay9sConfig) compose(order *entity.PayOrder) map[string]interface{} {
	dictionary := make(map[string]interface{})

	dictionary["amount"] = int(order.Amount / 100)
	dictionary["channelNo"] = t.ChargeId
	dictionary["merchantOrderNo"] = order.OrderID
	dictionary["userId"] = order.Userid
	dictionary["userBankAccount"] = "123456789"
	dictionary["userName"] = order.RealName
	dictionary["userPhone"] = order.Mobile
	dictionary["userEmail"] = order.Email
	dictionary["notifyUrl"] = t.PayNotifyUrl
	dictionary["createOrderTime"] = strconv.Itoa(int(time.Now().Unix()))
	dictionary["merchantId"] = t.Appid

	// 签名
	strSign := PaySign(dictionary, t.Md5Key)
	dictionary["sign"] = strSign

	return dictionary
}

// 提现组装参数
func (t *Pay9sConfig) WithdrawCompose(order *entity.WithdrawOrder) map[string]interface{} {
	dictionary := make(map[string]interface{})

	dictionary["amount"] = int(order.Amount / 100)
	dictionary["channelNo"] = t.WithdrawId
	dictionary["merchantId"] = t.Appid
	dictionary["merchantOrderNo"] = order.OrderID
	dictionary["payType"] = "IMPS"
	dictionary["userBankAccount"] = order.BankNumber
	dictionary["ifscCode"] = order.IFSC
	dictionary["userName"] = order.RealName
	dictionary["userPhone"] = order.Mobile
	dictionary["userEmail"] = order.Email
	dictionary["notifyUrl"] = t.WithDrawNotifyUrl
	dictionary["createOrderTime"] = strconv.Itoa(int(time.Now().Unix()))

	// 签名
	strSign := PaySign(dictionary, t.Md5Key)
	dictionary["sign"] = strSign

	return dictionary
}

// 提现查单
func (t *Pay9sConfig) InspectWdCompose(order *entity.WithdrawOrder) map[string]any {
	dictionary := make(map[string]any)

	dictionary["merchantId"] = t.Appid
	dictionary["orderNo"] = fmt.Sprint(order.OrderID)

	// 签名
	strSign := PaySign(dictionary, t.Md5Key)
	dictionary["sign"] = data.ToUpper(strSign)

	return dictionary
}

// 充值查单
func (t *Pay9sConfig) InspectCompose(order *entity.PayOrder) map[string]any {
	dictionary := make(map[string]any)

	dictionary["merchantId"] = t.Appid
	dictionary["orderNo"] = fmt.Sprint(order.OrderID)

	// 签名
	strSign := PaySign(dictionary, t.Md5Key)
	dictionary["sign"] = data.ToUpper(strSign)

	return dictionary
}

func ToQueryString(param map[string]interface{}) string {
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

func PaySign(param map[string]interface{}, md5key string) string {
	keys := make([]string, 0, len(param))
	for k := range param {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var result strings.Builder
	for _, v := range keys {
		if val, ok := param[v]; ok && val != nil && val != "" {
			// if s, ok := val.(string); ok {
			// 	val, _ = url.QueryUnescape(s)
			// }
			vStr := fmt.Sprintf("%v", val)
			result.WriteString(fmt.Sprintf("%s=%s&", v, vStr))
		}
	}

	if md5key != "" {
		result.WriteString(fmt.Sprintf("key=%s", md5key))
	}
	return utils.Md5(result.String())
}

// application/x-www-form-urlencoded PSOT请求
func doHttpForm(targetUrl string, body string) ([]byte, error) {
	req, err := http.NewRequest("POST", targetUrl, strings.NewReader(body))
	if err != nil {
		glog.Errorf("err:%s", err)
		return []byte(""), err
	}
	req.Header.Add("Content-type", "application/x-www-form-urlencoded")

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
