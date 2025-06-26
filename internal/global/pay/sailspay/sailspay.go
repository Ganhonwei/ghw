package sailspay

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"goserver/internal/global/pay/entity"
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

var Config *SailspayConfig

// Submit 下单
func (t *SailspayConfig) Submit(order *entity.PayOrder) ([]byte, error) {
	body := t.compose(order) //打包post数据
	// strReq := pay.ToQueryString(body)

	strReq, _ := jsoniter.Marshal(body)
	// strRes := strReq
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
func (t *SailspayConfig) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	body := t.WithdrawCompose(order) //打包post数据
	// strReq := pay.ToQueryString(body)
	strReq, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strReq)
	resp, err := doHttpForm(t.WithDrawUrl, strReq)
	//fmt.Printf("resp %s, err %v", string(resp), err)
	if err != nil {
		return nil, err
	}
	glog.Infof("WithdrawSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

// 组装参数
func (c *SailspayConfig) compose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)
	timestamp := fmt.Sprint(order.Ctime)
	dictionary["amount"] = strconv.Itoa(int(order.Amount))
	dictionary["merchantId"] = c.MachId
	dictionary["orderId"] = fmt.Sprint(order.OrderID)
	dictionary["timestamp"] = timestamp
	dictionary["notifyUrl"] = Config.PayNotifyUrl // 充值后返回地址

	// 签名
	strSign := PaySign(order.OrderID, timestamp, c.Md5Key)
	// dictionary["sign"] = pay.ToUpper(strSign)
	dictionary["sign"] = strSign
	return dictionary
}

// 提现组装参数
func (t *SailspayConfig) WithdrawCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)
	timestamp := fmt.Sprint(order.Ctime)
	dictionary["amount"] = strconv.Itoa(int(order.Amount))
	dictionary["merchantId"] = t.MachId
	dictionary["orderId"] = order.OrderID
	dictionary["timestamp"] = timestamp
	dictionary["notifyUrl"] = t.WithDrawNotifyUrl
	dictionary["outType"] = "IMPS"
	dictionary["accountNumber"] = order.BankNumber
	dictionary["ifsc"] = order.IFSC
	dictionary["accountHolder"] = order.RealName
	// 签名
	strSign := PaySign(order.OrderID, timestamp, t.Md5Key)

	// dictionary["sign"] = pay.ToUpper(strSign)
	dictionary["sign"] = strSign
	return dictionary
}

// func PaySign(param map[string]string, md5key string) string {
// 	keys := make([]string, 0, len(param))
// 	for k := range param {
// 		keys = append(keys, k)
// 	}
// 	sort.Strings(keys)

// 	var result strings.Builder
// 	for _, v := range keys {
// 		if val, ok := param[v]; ok && val != "" {
// 			result.WriteString(fmt.Sprintf("%s=%s&", v, param[v]))
// 		}
// 	}

// 	if md5key != "" {
// 		result.WriteString(fmt.Sprintf("key=%s", md5key))
// 	}
// 	return utils.Md5(result.String())
// }

func PaySign(orderId, timestamp, md5key string) string {
	result := orderId + timestamp + md5key
	return utils.Md5(result)
}

// application/x-www-form-urlencoded PSOT请求
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

func (c *SailspayConfig) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	result := new(OrderRespond)
	jsoniter.Unmarshal(body, result)
	if result.Code != 200 { // 没成功TODO
		return "", fmt.Errorf("code:%s, OrderNo:%s", result.Code, result.PayOrderId)
	}
	return result.PaymentUrl, nil
}

func (c *SailspayConfig) ParsePayResult(body []byte) (map[string]string, error) {
	tradeResult := RechargeCallback{}
	err := jsoniter.Unmarshal(body, &tradeResult)
	if err != nil {
		return nil, err
	}
	reslut := make(map[string]string)
	reslut["merchantId"] = fmt.Sprint(tradeResult.MerchantId)
	reslut["amount"] = fmt.Sprint(tradeResult.Amount)
	reslut["orderId"] = tradeResult.OrderId
	reslut["payOrderId"] = tradeResult.PayOrderId
	reslut["payTime"] = fmt.Sprint(tradeResult.Timestamp)
	reslut["status"] = fmt.Sprint(tradeResult.Status)
	reslut["sign"] = tradeResult.Sign
	return reslut, nil
}

func (c *SailspayConfig) ParseWithdrawResult(body []byte) (map[string]string, error) {
	tradeResult := WithdrawCallback{}
	err := jsoniter.Unmarshal(body, &tradeResult)
	if err != nil {
		return nil, err
	}
	reslut := make(map[string]string)
	reslut["merchantId"] = fmt.Sprint(tradeResult.MerchantId)
	reslut["amount"] = fmt.Sprint(tradeResult.Amount)
	reslut["orderId"] = tradeResult.OrderId
	reslut["payOrderId"] = tradeResult.PayOrderId
	reslut["payTime"] = fmt.Sprint(tradeResult.Timestamp)
	reslut["status"] = fmt.Sprint(tradeResult.Status)
	reslut["sign"] = tradeResult.Sign
	return reslut, nil
}

func (c *SailspayConfig) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	result := new(WithdrawRespond)
	jsoniter.Unmarshal(body, result)
	if result.Code != 200 { // 没成功TODO
		order.OrderStatus = data.OrderFail
		glog.Errorf("[sailspay] withdraw fail submit, code:%s, msg:%s", result.Code, result.Msg)
		return false, fmt.Sprint(result.Msg)
	}
	order.OrderStatus = data.OrderSuccess
	return true, fmt.Sprint(result.Code)
}

// 充值查单接口
func (t *SailspayConfig) InspectSubmit(order *entity.PayOrder) ([]byte, error) {
	// body := t.InspectCompose(order) //打包post数据
	// strReq, _ := jsoniter.Marshal(body)
	// resp, err := doHttpForm(t.InspectOrderUrl, strReq)
	// if err != nil {
	// 	return nil, err
	// }
	// glog.Infof("InspectSubmit order success, id:%s", order.OrderID)
	return nil, fmt.Errorf("fail")
}

// 充值查单响应
func (c *SailspayConfig) InspectResponse(body []byte, order *entity.PayOrder) error {
	// result := new(XFInspectRespond)
	// jsoniter.Unmarshal(body, result)
	// if result.Status != "2" { // 没成功TODO
	// 	return fmt.Errorf("code:%s, message:%s", result.Status, result.Remark)
	// }
	// return nil
	return fmt.Errorf("fail")
}

// 充值查单
// func (c *SailspayConfig) InspectCompose(order *entity.PayOrder) map[string]string {
// 	dictionary := make(map[string]string)

// 	dictionary["machId"] = c.AppId
// 	dictionary["appId"] = c.MachId
// 	dictionary["merchantOrderNo"] = fmt.Sprint(order.OrderID)

// 	// 签名
// 	strSign := PaySign(dictionary, c.Md5Key)
// 	// dictionary["sign"] = pay.ToUpper(strSign)
// 	dictionary["sign"] = strSign

// 	return dictionary
// }

// 提现查单接口
func (t *SailspayConfig) InspectWithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	// body := t.InspectWdCompose(order) //打包post数据
	// strReq, _ := jsoniter.Marshal(body)
	// resp, err := doHttpForm(t.InspectWithdrawUrl, strReq)
	// if err != nil {
	// 	return nil, err
	// }
	// glog.Infof("InspectWithdrawSubmit order success, id:%s", order.OrderID)
	// return resp, nil
	return nil, fmt.Errorf("fail")
}

// 提现查单响应
func (c *SailspayConfig) InspectWdResponse(body []byte, order *entity.WithdrawOrder) int {
	// result := new(XFInspectRespond)
	// jsoniter.Unmarshal(body, result)
	// if result.Status == "2" { // 成功
	// 	return 2
	// }
	return 3
}

// 提现查单
// func (c *SailspayConfig) InspectWdCompose(order *entity.WithdrawOrder) map[string]string {
// 	dictionary := make(map[string]string)

// 	dictionary["machId"] = c.AppId
// 	dictionary["appId"] = c.MachId
// 	dictionary["merchantOrderNo"] = fmt.Sprint(order.OrderID)

// 	// 签名
// 	strSign := PaySign(dictionary, c.WithDrawMDd5Key)
// 	// dictionary["sign"] = pay.ToUpper(strSign)
// 	dictionary["sign"] = strSign

// 	return dictionary
// }

// 余额查询
func (t *SailspayConfig) BalanceQuery() int64 {
	// return nil, fmt.Errorf("unsupport")
	glog.Info("unsupport")
	return 0
}
