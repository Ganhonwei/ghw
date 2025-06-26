package yunpay

import (
	"crypto"
	"fmt"
	"goserver/internal/global/pay/entity"
	"goserver/internal/global/pay/service"
	"goserver/pkg/glog"
	"strconv"

	jsoniter "github.com/json-iterator/go"
)

var Config *PayConfig

func (t *PayConfig) GetChannelId() uint32 {
	return service.YUNPAY
}

func (t *PayConfig) GetChannelName() string {
	return "yunpay"
}

// 验签
// func (c *PayConfig) VerifySing(sign, content string, hash crypto.Hash) bool {
// 	shaNew := hash.New()
// 	shaNew.Write([]byte(content))
// 	hashed := shaNew.Sum(nil)
// 	pubkey, err := service.ParsePublickKey(c.ServerPubRSA)
// 	if err != nil {
// 		glog.Error("get pubkey fail")
// 		return false
// 	}
// 	s, err2 := base64.StdEncoding.DecodeString(sign)
// 	if err2 != nil {
// 		glog.Error("verify sign fail, decode sign fail")
// 		return false
// 	}
// 	err1 := rsa.VerifyPKCS1v15(pubkey, hash, hashed[:], s)
// 	if err1 != nil {
// 		glog.Error("rsa fail")
// 		return false
// 	}
// 	return true
// }

// =========================实现接口方法=========================

// 充值查单接口
func (t *PayConfig) InspectSubmit(order *entity.PayOrder) ([]byte, error) {
	body := t.InspectCompose(order) //打包post数据
	strReq, _ := jsoniter.Marshal(body)
	resp, err := doHttpForm(t.InspectOrderUrl, strReq)
	if err != nil {
		return nil, err
	}
	glog.Infof("InspectSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

// 充值查单响应
func (c *PayConfig) InspectResponse(body []byte, order *entity.PayOrder) error {
	result := new(PayInspectRespond)
	jsoniter.Unmarshal(body, result)
	if result.Code != 200 { // 没成功TODO
		return fmt.Errorf("code:%d, message:%s", result.Code, result.Message)
	}
	if result.Data.Status != 2 {
		return fmt.Errorf("status:%d, message:%s", result.Data.Status, result.Message)
	}
	return nil
}

// 充值查单
func (c *PayConfig) InspectCompose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["merchant_code"] = c.Merchant
	dictionary["merchant_order_no"] = fmt.Sprint(order.OrderID)

	// 签名
	strSign := rsaSign(SignContent(dictionary), c.ClientPriRSA, crypto.SHA256)
	// dictionary["sign"] = pay.ToUpper(strSign)
	dictionary["sign"] = strSign

	return dictionary
}

// 提现查单接口
func (t *PayConfig) InspectWithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	body := t.InspectWdCompose(order) //打包post数据
	strReq, _ := jsoniter.Marshal(body)
	resp, err := doHttpForm(t.InspectWithdrawUrl, strReq)
	if err != nil {
		return nil, err
	}
	glog.Infof("InspectWithdrawSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

// 提现查单响应
func (c *PayConfig) InspectWdResponse(body []byte, order *entity.WithdrawOrder) int {
	result := new(PayInspectRespond)
	jsoniter.Unmarshal(body, result)
	if result.Code != 200 { // 不成功
		return 3
	}
	if result.Data.Status != 2 {
		return 3
	}
	return 2
}

// 提现查单
func (c *PayConfig) InspectWdCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["merchant_code"] = c.Merchant
	dictionary["merchant_order_no"] = fmt.Sprint(order.OrderID)

	// 签名
	strSign := rsaSign(SignContent(dictionary), c.ClientPriRSA, crypto.SHA256)
	dictionary["sign"] = strSign

	return dictionary
}

// 余额查询
func (c *PayConfig) BalanceQuery() int64 {
	param := c.QueryCompose()
	strReq, _ := jsoniter.Marshal(param)
	resp, err := doHttpForm(c.QueryUrl, strReq)
	if err != nil {
		glog.Info("[uwinpay] BalanceQuery order fail, err,", err)
		return 0
	}

	rsp := new(PayBalanceRespond)
	err = jsoniter.Unmarshal(resp, &rsp)
	if err != nil {
		glog.Info("[uwinpay] BalanceQuery order fail, err,", err)
		return 0
	}

	if rsp.Code != 200 {
		glog.Info("[uwinpay] BalanceQuery order fail, code:", rsp.Code)
		return 0
	}
	glog.Infof("[uwinpay] BalanceQuery order success")

	balance, _ := strconv.ParseFloat(rsp.Data.AvailableBalance, 64)
	return int64(balance * 100)
}

func (c *PayConfig) QueryCompose() map[string]string {
	dictionary := map[string]string{
		"merchant_code": c.Merchant,
	}

	strSign := rsaSign(SignContent(dictionary), c.ClientPriRSA, crypto.SHA256)
	// dictionary["sign"] = pay.ToUpper(strSign)
	dictionary["sign"] = strSign

	return dictionary
}
