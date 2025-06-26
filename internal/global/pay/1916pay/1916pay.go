package pay1916

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
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var Config *Pay1916Config

// Submit 充值下单
func (t *Pay1916Config) Submit(order *entity.PayOrder) ([]byte, error) {
	body := t.compose(order) //打包post数据
	strReq := ToQueryString(body)
	//fmt.Printf("body %s, order %#v\n", body, order)
	strpar, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strpar)
	resp, err := doHttpForm(t.PayOrderUrl, strReq)
	//fmt.Printf("resp %s, err %v", string(resp), err)
	if err != nil {
		return nil, err
	}
	glog.Infof("Submit order success, id:%s", order.OrderID)
	return resp, nil
}

// 提现下单
func (t *Pay1916Config) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
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
func (t *Pay1916Config) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	result := new(Pay1916OrderRespond)
	if err := jsoniter.Unmarshal(body, result); err != nil {
		glog.Error(err)
		return "", err
	}
	if result.ErrorCode != 0 || result.Data.PayUrl == "" { // 没成功TODO
		return "", fmt.Errorf("code:%d, message:%s", result.ErrorCode, result.ErrorMsg)
	}
	return result.Data.PayUrl, nil
}

// 提现响应
func (t *Pay1916Config) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	result := new(Pay1916OrderRespond)
	if err := jsoniter.Unmarshal(body, result); err != nil {
		glog.Error(err)
		return false, err.Error()
	}
	if result.ErrorCode == 0 { // 没成功TODO
		// 发条消息提单成功
		// glog.Infof("[withdraw fail] submit success. order %s, user:%s", order.OrderID, order.Userid)
		order.OrderStatus = data.OrderSuccess
		return true, strconv.Itoa(result.ErrorCode)
	}
	glog.Infof("[1916 withdraw] submit fail. code:%d, msg:%s", result.ErrorCode, result.ErrorMsg)
	order.OrderStatus = data.OrderFail
	return false, strconv.Itoa(result.ErrorCode)
}

func (t *Pay1916Config) ParsePayResult(body []byte) (map[string]interface{}, error) {
	tradeResult := Pay1916RechargeCallback{}
	err := jsoniter.Unmarshal(body, &tradeResult)
	if err != nil {
		return nil, err
	}
	reslut := make(map[string]interface{})
	reslut["merchantNo"] = tradeResult.MerchantNo
	reslut["orderNo"] = tradeResult.OrderNo
	reslut["pOrderNo"] = tradeResult.POrderNo
	reslut["amount"] = tradeResult.Amount
	reslut["timestamp"] = tradeResult.Timestamp
	reslut["state"] = tradeResult.State
	reslut["utr"] = tradeResult.Utr
	reslut["sign"] = tradeResult.Sign

	return reslut, nil
}

// 充值组装参数
func (t *Pay1916Config) compose(order *entity.PayOrder) map[string]interface{} {
	dictionary := make(map[string]interface{})

	dictionary["merchantNo"] = t.Appid
	dictionary["orderNo"] = order.OrderID
	dictionary["channelType"] = "1"
	dictionary["amount"] = int(order.Amount / 100)
	dictionary["returnUrl"] = "https://www.google.ocm"
	dictionary["notifyUrl"] = t.PayNotifyUrl
	dictionary["timestamp"] = strconv.Itoa(int(time.Now().Unix()))

	// dictionary["partnerId"] = t.Appid
	// dictionary["applicationId"] = t.ApplicationId
	// dictionary["payWay"] = "2"
	// dictionary["partnerOrderNo"] = fmt.Sprint(order.OrderID)
	// dictionary["amount"] = strconv.Itoa(int(order.Amount))
	// dictionary["currency"] = "INR"
	// dictionary["name"] = order.NickName
	// dictionary["gameId"] = order.Userid
	// dictionary["clientIp"] = "127.0.0.1" //order.OrderAddress // 获取IP地址
	// dictionary["notifyUrl"] = t.PayNotifyUrl
	// dictionary["subject"] = "Recharge"
	// dictionary["body"] = order.Userid + " Game Recharge"
	// dictionary["callbackUrl"] = Config.PayNotifyUrl // 充值后返回地址

	dicExtra := make(map[string]string)
	// 向 map 中添加键值对
	dicExtra["username"] = order.RealName
	dicExtra["email"] = order.Email
	dicExtra["phone"] = order.Mobile

	jsonData, err := json.Marshal(dicExtra)
	if err != nil {
		fmt.Println("JSON encoding failed:", err)
	}
	strExtra := string(jsonData)
	dictionary["extra"] = strExtra

	// 签名
	strSign := PaySign(dictionary, t.Md5Key)
	dictionary["sign"] = strSign

	// URLENCODE编码,计算签名使用编码前数值
	// dictionary["notifyUrl"] = pay.ToUrlEncode(t.PayNotifyUrl)
	// dictionary["subject"] = pay.ToUrlEncode("Recharge")
	// dictionary["body"] = pay.ToUrlEncode(order.Userid + " Game Recharge")
	// dictionary["extra"] = pay.ToUrlEncode(strExtra)
	return dictionary
}

// 提现组装参数
func (t *Pay1916Config) WithdrawCompose(order *entity.WithdrawOrder) map[string]interface{} {
	dictionary := make(map[string]interface{})

	dictionary["merchantNo"] = t.Appid
	dictionary["orderNo"] = order.OrderID
	dictionary["channelType"] = "102"
	dictionary["amount"] = int(order.Amount / 100)
	dictionary["notifyUrl"] = t.WithDrawNotifyUrl
	dictionary["timestamp"] = strconv.Itoa(int(time.Now().Unix()))

	dicExtra := make(map[string]string)
	// 向 map 中添加键值对
	dicExtra["username"] = order.RealName
	dicExtra["email"] = order.Email
	dicExtra["phone"] = order.Mobile
	dicExtra["bankname"] = order.Bank
	dicExtra["accountNo"] = order.BankNumber
	dicExtra["ifsc"] = order.IFSC

	jsonData, err := json.Marshal(dicExtra)
	if err != nil {
		fmt.Println("JSON encoding failed:", err)
	}
	strExtra := string(jsonData)
	dictionary["extra"] = strExtra

	// dictionary["partnerId"] = t.Appid
	// dictionary["partnerWithdrawNo"] = fmt.Sprint(order.OrderID)
	// dictionary["amount"] = strconv.Itoa(int(order.Amount))
	// dictionary["currency"] = "INR"
	// dictionary["gameId"] = order.Userid
	// dictionary["notifyUrl"] = t.WithDrawNotifyUrl
	// dictionary["receiptMode"] = "1" // 0:UPI 1:IMPS
	// dictionary["accountNumber"] = order.BankNumber
	// dictionary["accountName"] = order.RealName
	// dictionary["accountPhone"] = order.Mobile
	// dictionary["accountEmail"] = order.Email
	// dictionary["accountExtra1"] = order.IFSC
	// dictionary["accountExtra2"] = order.IFSC[:4]
	// dictionary["version"] = "1.0"

	// 签名
	strSign := PaySign(dictionary, t.WithDrawMDd5Key)
	dictionary["sign"] = strSign

	// // URLENCODE编码,计算签名使用编码前数值
	// dictionary["notifyUrl"] = pay.ToUrlEncode(t.WithDrawNotifyUrl)
	// dictionary["accountNumber"] = pay.ToUrlEncode(order.BankNumber)
	// dictionary["accountName"] = pay.ToUrlEncode(order.RealName)
	// dictionary["accountPhone"] = pay.ToUrlEncode(order.Mobile)
	// dictionary["accountEmail"] = pay.ToUrlEncode(order.Email)
	return dictionary
}

// 充值查单接口
func (t *Pay1916Config) InspectSubmit(order *entity.PayOrder) ([]byte, error) {
	body := t.InspectCompose(order) //打包post数据
	strReq := ToQueryString(body)
	resp, err := doHttpForm(t.InspectOrderUrl, strReq)
	if err != nil {
		return nil, err
	}
	glog.Infof("InspectSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

// 提现查单接口
func (t *Pay1916Config) InspectWithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	body := t.InspectWdCompose(order) //打包post数据
	strReq := ToQueryString(body)
	resp, err := doHttpForm(t.InspectWithdrawUrl, strReq)
	if err != nil {
		return nil, err
	}
	glog.Infof("InspectWithdrawSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

// 充值查单
func (t *Pay1916Config) InspectCompose(order *entity.PayOrder) map[string]any {
	dictionary := make(map[string]any)

	dictionary["merchantNo"] = t.Appid
	dictionary["orderNo"] = fmt.Sprint(order.OrderID)

	// 签名
	strSign := PaySign(dictionary, t.Md5Key)
	dictionary["sign"] = data.ToUpper(strSign)

	return dictionary
}

// 提现查单
func (t *Pay1916Config) InspectWdCompose(order *entity.WithdrawOrder) map[string]any {
	dictionary := make(map[string]any)

	dictionary["merchantNo"] = t.Appid
	dictionary["orderNo"] = fmt.Sprint(order.OrderID)

	// 签名
	strSign := PaySign(dictionary, t.WithDrawMDd5Key)
	dictionary["sign"] = data.ToUpper(strSign)

	return dictionary
}

// 充值查单响应
func (c *Pay1916Config) InspectResponse(body []byte, order *entity.PayOrder) error {
	result := new(Pay1916InspectRespond)
	if err := jsoniter.Unmarshal(body, result); err != nil {
		return err
	}
	if result.ErrorCode != 0 || result.ErrorMsg != "SUCCESS" { // 没成功TODO
		return fmt.Errorf("code:%d, message:%s", result.ErrorCode, result.ErrorMsg)
	}
	if result.Data.State != 1 {
		return fmt.Errorf("state:%d, message:%s", result.Data.State, result.Data.Msg)
	}
	order.OutTradeNo = result.Data.POrderNo
	return nil
}

// 提现查单响应
func (c *Pay1916Config) InspectWdResponse(body []byte, order *entity.WithdrawOrder) int {
	result := new(Pay1916InspectWdRespond)
	if err := jsoniter.Unmarshal(body, result); err != nil {
		return 3
	}
	if result.ErrorCode != 0 || result.ErrorMsg != "SUCCESS" { // 没成功TODO
		return 3
	}
	if result.Data.State == 1 {
		return 2
	}
	return 3
}

// 余额查询
func (c *Pay1916Config) BalanceQuery() int64 {
	param := map[string]any{
		"merchantNo": c.Appid,
		"area":       "印度",
	}
	strReq := ToQueryString(param)
	resp, err := doHttpForm(c.QueryUrl, strReq)
	if err != nil {
		glog.Info("[1916] BalanceQuery order fail,err:", err)
		return 0
	}

	rsp := new(Pay1916BalanceRespond)
	err = json.Unmarshal(resp, &rsp)
	if err != nil {
		glog.Info("[1916] BalanceQuery order fail,err:", err)
		return 0
	}

	if rsp.ErrorCode != 0 && rsp.ErrorMsg != "SUCCESS" {
		return 0
	}

	balance, _ := strconv.ParseFloat(rsp.Data.Balance, 64)
	glog.Infof("[1916] BalanceQuery order success")
	return int64(balance * 100)
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

func PaySign(param map[string]interface{}, md5key string) string {
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
	return utils.Md5(result.String())
}
