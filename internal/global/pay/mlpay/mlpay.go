package mlpay

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
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var Config *MlPayConfig

// Submit 充值下单
func (t *MlPayConfig) Submit(order *entity.PayOrder) ([]byte, error) {
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
func (t *MlPayConfig) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
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

// 充值查单接口
func (t *MlPayConfig) InspectSubmit(order *entity.PayOrder) ([]byte, error) {
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
func (t *MlPayConfig) InspectWithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	body := t.InspectWdCompose(order) //打包post数据
	strReq := ToQueryString(body)
	resp, err := doHttpForm(t.InspectWithdrawUrl, strReq)
	if err != nil {
		return nil, err
	}
	glog.Infof("InspectWithdrawSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

// 充值组装参数
func (this *MlPayConfig) compose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["partnerId"] = this.Appid
	dictionary["applicationId"] = this.ApplicationId
	dictionary["payWay"] = "2"
	dictionary["partnerOrderNo"] = fmt.Sprint(order.OrderID)
	dictionary["amount"] = strconv.Itoa(int(order.Amount))
	dictionary["currency"] = "INR"
	dictionary["name"] = order.NickName
	dictionary["gameId"] = order.Userid
	dictionary["clientIp"] = "127.0.0.1" //order.OrderAddress // 获取IP地址
	dictionary["notifyUrl"] = this.PayNotifyUrl
	dictionary["subject"] = "Recharge"
	dictionary["body"] = order.Userid + " Game Recharge"
	dictionary["callbackUrl"] = Config.PayNotifyUrl // 充值后返回地址

	dicExtra := make(map[string]string)
	// 向 map 中添加键值对
	dicExtra["userName"] = order.RealName
	dicExtra["userEmail"] = order.Email
	dicExtra["userPhone"] = order.Mobile

	jsonData, err := json.Marshal(dicExtra)
	if err != nil {
		fmt.Println("JSON encoding failed:", err)
	}
	strExtra := string(jsonData)
	dictionary["extra"] = strExtra
	dictionary["version"] = "1.0"

	// 签名
	strSign := PaySign(dictionary, this.Md5Key)
	dictionary["sign"] = data.ToUpper(strSign)

	// URLENCODE编码,计算签名使用编码前数值
	dictionary["notifyUrl"] = data.ToUrlEncode(this.PayNotifyUrl)
	dictionary["subject"] = data.ToUrlEncode("Recharge")
	dictionary["body"] = data.ToUrlEncode(order.Userid + " Game Recharge")
	dictionary["extra"] = data.ToUrlEncode(strExtra)
	return dictionary
}

// 提现组装参数
func (t *MlPayConfig) WithdrawCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["partnerId"] = t.Appid
	dictionary["partnerWithdrawNo"] = fmt.Sprint(order.OrderID)
	dictionary["amount"] = strconv.Itoa(int(order.Amount))
	dictionary["currency"] = "INR"
	dictionary["gameId"] = order.Userid
	dictionary["notifyUrl"] = t.WithDrawNotifyUrl
	dictionary["receiptMode"] = "1" // 0:UPI 1:IMPS
	dictionary["accountNumber"] = order.BankNumber
	dictionary["accountName"] = order.RealName
	dictionary["accountPhone"] = order.Mobile
	dictionary["accountEmail"] = order.Email
	dictionary["accountExtra1"] = order.IFSC
	dictionary["accountExtra2"] = order.IFSC[:4]
	dictionary["version"] = "1.0"

	// 签名
	strSign := PaySign(dictionary, t.WithDrawMDd5Key)
	dictionary["sign"] = data.ToUpper(strSign)

	// URLENCODE编码,计算签名使用编码前数值
	dictionary["notifyUrl"] = data.ToUrlEncode(t.WithDrawNotifyUrl)
	dictionary["accountNumber"] = data.ToUrlEncode(order.BankNumber)
	dictionary["accountName"] = data.ToUrlEncode(order.RealName)
	dictionary["accountPhone"] = data.ToUrlEncode(order.Mobile)
	dictionary["accountEmail"] = data.ToUrlEncode(order.Email)
	return dictionary
}

// 充值查单
func (t *MlPayConfig) InspectCompose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["partnerId"] = t.Appid
	dictionary["applicationId"] = t.ApplicationId
	dictionary["partnerOrderNo"] = fmt.Sprint(order.OrderID)
	dictionary["version"] = "1.0"

	// 签名
	strSign := PaySign(dictionary, t.Md5Key)
	dictionary["sign"] = data.ToUpper(strSign)

	return dictionary
}

// 提现查单
func (t *MlPayConfig) InspectWdCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["partnerId"] = t.Appid
	dictionary["partnerWithdrawNo"] = fmt.Sprint(order.OrderID)
	dictionary["version"] = "1.0"

	// 签名
	strSign := PaySign(dictionary, t.WithDrawMDd5Key)
	dictionary["sign"] = data.ToUpper(strSign)

	return dictionary
}

// 将结构体转成map
func structToMapString(input interface{}) map[string]string {
	result := make(map[string]string)

	v := reflect.ValueOf(input)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fieldName := t.Field(i).Name
		fieldValue := fmt.Sprintf("%v", v.Field(i).Interface())
		result[fieldName] = fieldValue
	}

	return result
}

func ToQueryString(param map[string]string) string {
	var args string
	for k, v := range param {
		if args == "" {
			args += fmt.Sprintf("%s=%s", k, v)
		} else {
			args += fmt.Sprintf("&%s=%s", k, v)
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
			result.WriteString(fmt.Sprintf("%s=%s&", v, param[v]))
		}
	}

	if md5key != "" {
		result.WriteString(fmt.Sprintf("key=%s", md5key))
	}
	return utils.Md5(result.String())
}

// 充值响应
func (c *MlPayConfig) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	result := new(MlOrderRespond)
	jsoniter.Unmarshal(body, result)
	if result.Code != "0000" { // 没成功TODO
		return "", fmt.Errorf("code:%s, message:%s", result.Code, result.Message)
	}
	return result.Data, nil
	// return "", nil
}

// 提现响应
func (c *MlPayConfig) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	result := new(MlOrderRespond)
	jsoniter.Unmarshal(body, result)
	if result.Code == "0000" { // 没成功TODO
		// 发条消息提单成功
		// glog.Infof("[withdraw fail] submit success. order %s, user:%s", order.OrderID, order.Userid)
		order.OrderStatus = data.OrderSuccess
		return true, result.Code
	}
	glog.Infof("[ml withdraw] submit fail. code:%s, msg:%s", result.Code, result.Message)
	order.OrderStatus = data.OrderFail
	return false, result.Code

}

// 充值查单响应
func (c *MlPayConfig) InspectResponse(body []byte, order *entity.PayOrder) error {
	result := new(MlInspectRespond)
	jsoniter.Unmarshal(body, result)
	if result.Code != "0000" { // 没成功TODO
		return fmt.Errorf("code:%s, message:%s", result.Code, result.Message)
	}
	if result.Data.Status != 2 {
		return fmt.Errorf("status:%s, message:%s", result.Code, result.Message)
	}
	order.OutTradeNo = result.Data.OrderNo
	return nil
}

// 提现查单响应
func (c *MlPayConfig) InspectWdResponse(body []byte, order *entity.WithdrawOrder) int {
	result := new(MlInspectWdRespond)
	jsoniter.Unmarshal(body, result)
	if result.Code != "0000" { // 没成功TODO
		return 3
	}
	return int(result.Data.Status)
}

// 余额查询
func (c *MlPayConfig) BalanceQuery() int64 {
	glog.Errorf("unsupport")
	return 0
}

// 获取提现查单失败状态
func (c *MlPayConfig) GetWithdrawStatus(order *entity.WithdrawOrder) int {
	param := make(map[string]any)
	if order.RefuseReason != "" {
		err := json.Unmarshal([]byte(order.RefuseReason), &param)
		if err != nil {
			glog.Errorf("json unmarshal `refuseReason` fail ,order %s", order.OrderID)
			return 4
		}
		if msg, ok := param["message"]; ok {
			if m, ok := msg.(string); ok {
				if strings.Contains(m, "IFSC") || strings.Contains(m, "accountNo") ||
					strings.Contains(m, "Invalid") {
					return 5
				}
				if strings.Contains(m, "Insufficient balance") {
					return 7
				}
				if strings.Contains(m, "channel") {
					return 6
				}
				return 4
			}
		}
	}

	if order.RequestMsg != "" {
		err := json.Unmarshal([]byte(order.RequestMsg), &param)
		if err != nil {
			glog.Errorf("json unmarshal `RequestMsg` fail ,order %s", order.OrderID)
			return 4
		}
		if msg, ok := param["message"]; ok {
			if m, ok := msg.(string); ok {
				if strings.Contains(m, "IFSC") || strings.Contains(m, "accountNo") ||
					strings.Contains(m, "Invalid") {
					return 5
				}
				if strings.Contains(m, "Insufficient balance") {
					return 7
				}
				if strings.Contains(m, "channel") {
					return 6
				}
				return 4
			}
		}
	}
	return 4
}
