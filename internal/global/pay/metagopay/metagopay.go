package metagopay

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

var Config *MetagopayConfig

// 充值订单
func (t *MetagopayConfig) Submit(order *entity.PayOrder) ([]byte, error) {
	body := t.compose(order) //打包post数据
	strReq := ToQueryString(body)
	// fmt.Printf("body %s, order %#v\n", body, order)
	strpar, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strpar)
	resp, err := doHttpForm("https://api.metagopayments.com/cashier/pay.ac", strReq)
	if err != nil {
		return nil, err
	}
	glog.Infof("Submit order success, id:%s", order.OrderID)
	// fmt.Println("resp:", string(resp))
	return resp, nil
}

func (t *MetagopayConfig) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	result := new(PayOrderRespond)
	jsoniter.Unmarshal(body, result)
	if result.Code != "000000" || result.OrdStatus == "02" || result.BusContent == "" {
		return "", fmt.Errorf("code:%s, ordStatus=%s, message:%s", result.Code, result.OrdStatus, result.Msg)
	}
	return result.BusContent, nil
}

func (t *MetagopayConfig) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	body := t.WithdrawCompose(order) //打包post数据
	strReq := ToQueryString(body)
	strpar, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strpar)
	resp, err := doHttpForm("https://paypout.metagopayments.com/cashier/TX0001.ac", strReq)
	//fmt.Printf("resp %s, err %v", string(resp), err)
	if err != nil {
		return nil, err
	}
	glog.Infof("WithdrawSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

func (t *MetagopayConfig) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	result := new(WithdrawOrderRespond)
	jsoniter.Unmarshal(body, result)
	if result.Code == "000000" && result.OrdStatus != "08" {
		// 发条消息提单成功
		order.OrderStatus = data.OrderSuccess
		// glog.Infof("[withdraw fail] submit success. order %s, user:%s", order.OrderID, order.Userid)
		return true, result.Code
	}
	glog.Infof("[metagopay withdraw] submit fail. code:%d, ordStatus=%s, msg:%s", result.Code, result.OrdStatus, result.Msg)
	order.OrderStatus = data.OrderFail
	return false, result.Code
}

func (t *MetagopayConfig) InspectSubmit(order *entity.PayOrder) ([]byte, error) {
	body := t.InspectCompose(order) //打包post数据
	strReq := ToQueryString(body)
	resp, err := doHttpForm("https://queryapi.metagopayments.com/cashier/query.ac", strReq)
	if err != nil {
		return nil, err
	}
	glog.Infof("InspectSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

func (t *MetagopayConfig) InspectResponse(body []byte, order *entity.PayOrder) error {
	result := new(PayInspectRespond)
	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		return err
	}
	if result.Code != "000000" { // 没成功
		return fmt.Errorf("code:%s, message:%s", result.Code, result.Msg)
	}
	if result.OrdStatus != "01" {
		return fmt.Errorf("state:%s, message:%s", result.OrdStatus, result.Msg)
	}
	order.OutTradeNo = result.PrdOrdNo
	return nil
}

func (t *MetagopayConfig) InspectWithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	body := t.InspectWdCompose(order) //打包post数据
	strReq := ToQueryString(body)
	resp, err := doHttpForm("https://queryapi.metagopayments.com/cashier/TX0002.ac", strReq)
	if err != nil {
		return nil, err
	}
	glog.Infof("InspectWithdrawSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

func (t *MetagopayConfig) InspectWdResponse(body []byte, order *entity.WithdrawOrder) int {
	result := new(WithdrawInspectRespond)
	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		return 3
	}
	if result.Code != "000000" { // 没成功
		return 3
	}
	if result.OrdStatus == "07" {
		return 2
	}
	return 3
}

func (t *MetagopayConfig) BalanceQuery() int64 {
	param := map[string]any{
		"version": "2.1",
		"orgNo":   t.OrgId,
		"custId":  t.MchId,
		"account": t.AccountId,
	}
	sign := PaySign(param, t.Md5Key)
	param["sign"] = sign
	strReq := ToQueryString(param)
	resp, err := doHttpForm("https://queryapi.metagopayments.com/cashier/balance.ac", strReq)
	if err != nil {
		glog.Info("[metagopay] BalanceQuery order fail,err:", err)
		return 0
	}
	// fmt.Println(string(resp))

	rsp := new(BalanceRespond)
	err = json.Unmarshal(resp, &rsp)
	if err != nil {
		glog.Info("[metagopay] BalanceQuery order fail,err:", err)
		return 0
	}

	if rsp.Code != "000000" {
		return 0
	}
	balance, _ := strconv.ParseFloat(rsp.AcBal, 64)
	glog.Infof("[metagopay] BalanceQuery success")
	return int64(balance)
}

func (t *MetagopayConfig) ParsePayResult(body []byte) (map[string]interface{}, error) {
	// bodyStr := string(body)
	// fmt.Println("bodyStr:", bodyStr)
	tradeResult := PayRechargeCallback{}
	err := jsoniter.Unmarshal(body, &tradeResult)
	if err != nil {
		return nil, err
	}
	reslut := make(map[string]interface{})
	reslut["version"] = tradeResult.Version
	reslut["orgNo"] = tradeResult.OrgNo
	reslut["custId"] = tradeResult.CustId
	reslut["custOrderNo"] = tradeResult.CustOrderNo
	reslut["prdOrdNo"] = tradeResult.PrdOrdNo
	reslut["ordAmt"] = tradeResult.OrdAmt
	reslut["ordTime"] = tradeResult.OrdTime
	reslut["payAmt"] = tradeResult.PayAmt
	reslut["utr"] = tradeResult.Utr
	reslut["ordStatus"] = tradeResult.OrdStatus
	reslut["sign"] = tradeResult.Sign
	return reslut, nil
}

func (t *MetagopayConfig) ParseWithdrawResult(body []byte) (map[string]interface{}, error) {
	tradeResult := PayWithdrawCallback{}
	err := jsoniter.Unmarshal(body, &tradeResult)
	if err != nil {
		return nil, err
	}
	reslut := make(map[string]interface{})
	reslut["orgNo"] = tradeResult.OrgNo
	reslut["custId"] = tradeResult.CustId
	reslut["custOrderNo"] = tradeResult.CustOrderNo
	reslut["prdOrdNo"] = tradeResult.PrdOrdNo
	reslut["payAmt"] = tradeResult.PayAmt
	reslut["ordStatus"] = tradeResult.OrdStatus
	reslut["casDesc"] = tradeResult.CasDesc
	reslut["utr"] = tradeResult.Utr
	reslut["sign"] = tradeResult.Sign
	return reslut, nil
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
			result.WriteString(fmt.Sprintf("%s=%v&", v, val))
		}
	}

	if md5key != "" {
		result.WriteString(fmt.Sprintf("key=%s", md5key))
	} else {
		glog.Errorf("metagopay md5 key is empty... order: %v", param["custOrderNo"])
	}

	signParam := result.String()
	// fmt.Println("signParam:", signParam)

	md5 := utils.Md5(signParam)
	return strings.ToUpper(md5)
}

func ToQueryString(param map[string]interface{}) string {
	var args string
	for k, v := range param {
		vStr := fmt.Sprintf("%v", v)
		vStr = url.QueryEscape(vStr)

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

// 充值组装参数
func (t *MetagopayConfig) compose(order *entity.PayOrder) map[string]interface{} {
	dictionary := make(map[string]interface{})

	dictionary["version"] = "2.1"
	dictionary["orgNo"] = t.OrgId
	dictionary["custId"] = t.MchId
	dictionary["custOrderNo"] = order.OrderID
	dictionary["tranType"] = "0412" // 交易类型: UPI-支付
	dictionary["clearType"] = "01"
	dictionary["payAmt"] = order.Amount
	dictionary["backUrl"] = t.PayNotifyUrl
	dictionary["frontUrl"] = "https://google.com"
	dictionary["goodsName"] = fmt.Sprintf("Cash %.2f", float64(order.Amount)/100)
	dictionary["orderDesc"] = dictionary["goodsName"]
	dictionary["buyIp"] = order.RegistIp
	dictionary["userName"] = order.RealName
	dictionary["userEmail"] = order.Email
	dictionary["userPhone"] = order.Mobile
	dictionary["countryCode"] = "IN"
	dictionary["currency"] = "INR"

	// 签名
	strSign := PaySign(dictionary, t.Md5Key)
	// fmt.Println("sign:", strSign)
	dictionary["sign"] = strSign

	return dictionary
}

// 充值查单
func (t *MetagopayConfig) InspectCompose(order *entity.PayOrder) map[string]any {
	dictionary := make(map[string]any)

	dictionary["version"] = "2.1"
	dictionary["orgNo"] = t.OrgId
	dictionary["custId"] = t.MchId
	dictionary["custOrderNo"] = order.OrderID

	// 签名
	strSign := PaySign(dictionary, t.Md5Key)
	dictionary["sign"] = data.ToUpper(strSign)

	return dictionary
}

// 提现组装参数
func (t *MetagopayConfig) WithdrawCompose(order *entity.WithdrawOrder) map[string]interface{} {
	dictionary := make(map[string]interface{})

	dictionary["version"] = "2.1"
	dictionary["orgNo"] = t.OrgId
	dictionary["custId"] = t.MchId
	dictionary["custOrdNo"] = order.OrderID
	dictionary["casType"] = "00"
	dictionary["country"] = "IN"
	dictionary["currency"] = "INR"
	dictionary["casAmt"] = order.Amount
	dictionary["deductWay"] = "02"
	dictionary["callBackUrl"] = t.WithdrawNotifyUrl
	dictionary["account"] = t.AccountId // 代付子账户名称
	dictionary["payoutType"] = "Card"
	dictionary["accountName"] = order.RealName
	dictionary["payeeBankCode"] = order.Bank
	dictionary["cardType"] = "IMPS" // todo
	dictionary["cnapsCode"] = order.IFSC
	dictionary["cardNo"] = order.BankNumber
	dictionary["phone"] = order.Mobile
	dictionary["email"] = order.Email

	// 签名
	strSign := PaySign(dictionary, t.Md5Key)
	dictionary["sign"] = strSign

	return dictionary
}

// 提现查单
func (t *MetagopayConfig) InspectWdCompose(order *entity.WithdrawOrder) map[string]any {
	dictionary := make(map[string]any)

	dictionary["version"] = "2.1"
	dictionary["orgNo"] = t.OrgId
	dictionary["custId"] = t.MchId
	dictionary["custOrdNo"] = order.OrderID

	// 签名
	strSign := PaySign(dictionary, t.Md5Key)
	dictionary["sign"] = data.ToUpper(strSign)

	return dictionary
}
