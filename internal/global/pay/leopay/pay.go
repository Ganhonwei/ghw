package leopay

import (
	"errors"
	"fmt"
	"goserver/internal/global/pay/entity"
	"goserver/internal/global/pay/service"
	"strconv"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var Config *PayConfig

func (t *PayConfig) GetChannelId() uint32 {
	return service.LEOPAY
}

func (t *PayConfig) GetChannelName() string {
	return "leopay"
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
	// body := t.InspectCompose(order) //打包post数据
	// strReq, _ := jsoniter.Marshal(body)
	// resp, err := doHttpForm(t.InspectOrderUrl, strReq)
	// if err != nil {
	// 	return nil, err
	// }
	// glog.Infof("InspectSubmit order success, id:%s", order.OrderID)
	// return resp, nil
	return nil, errors.New("not implement")
}

// 充值查单响应
func (c *PayConfig) InspectResponse(body []byte, order *entity.PayOrder) error {
	result := new(PayInspectRespond)
	jsoniter.Unmarshal(body, result)
	if result.Code != 1000 { // 没成功TODO
		return fmt.Errorf("code:%d, message:%s", result.Code, result.Message)
	}
	if result.Data.OrderStatus != "success" {
		return fmt.Errorf("status:%s, message:%s", result.Data.OrderStatus, result.Message)
	}
	return nil
}

// 充值查单
func (c *PayConfig) InspectCompose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["pay_memberid"] = c.Merchant
	dictionary["pay_orderid"] = order.OrderID
	dictionary["pay_time"] = strconv.FormatInt(time.Now().Unix(), 10)

	// 签名
	strSign := PaySign(dictionary, c.MD5Key)
	dictionary["sign"] = strSign

	return dictionary
}

// 提现查单接口
func (t *PayConfig) InspectWithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	// body := t.InspectWdCompose(order) //打包post数据
	// strReq, _ := jsoniter.Marshal(body)
	// resp, err := doHttpForm(t.InspectWithdrawUrl, strReq)
	// if err != nil {
	// 	return nil, err
	// }
	// glog.Infof("InspectWithdrawSubmit order success, id:%s", order.OrderID)
	// return resp, nil
	return nil, errors.New("not implement")
}

// 提现查单响应
func (c *PayConfig) InspectWdResponse(body []byte, order *entity.WithdrawOrder) int {
	result := new(PayInspectRespond)
	jsoniter.Unmarshal(body, result)
	if result.Code != 1000 { // 不成功
		return 3
	}
	if result.Data.OrderStatus != "success" {
		return 3
	}
	return 2
}

// 提现查单
func (c *PayConfig) InspectWdCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["pay_memberid"] = c.Merchant
	dictionary["pay_orderid"] = order.OrderID
	dictionary["pay_time"] = strconv.FormatInt(time.Now().Unix(), 10)

	// 签名
	strSign := PaySign(dictionary, c.MD5Key)
	dictionary["sign"] = strSign

	return dictionary
}

// 余额查询
func (c *PayConfig) BalanceQuery() int64 {
	// param := c.QueryCompose()
	// strReq, _ := jsoniter.Marshal(param)
	// resp, err := doHttpForm(c.QueryUrl, strReq)
	// if err != nil {
	// 	glog.Info("[leopay] BalanceQuery order fail, err,", err)
	// 	return 0
	// }

	// rsp := new(PayBalanceRespond)
	// err = jsoniter.Unmarshal(resp, &rsp)
	// if err != nil {
	// 	glog.Info("[leopay] BalanceQuery order fail, err,", err)
	// 	return 0
	// }

	// if rsp.Code != 1000 {
	// 	glog.Info("[leopay] BalanceQuery order fail, code:", rsp.Code)
	// 	return 0
	// }
	// glog.Infof("[leopay] BalanceQuery order success")

	// balance, _ := strconv.ParseFloat(rsp.Data.Balance, 64)
	// return int64(balance * 100)
	return 0
}

func (c *PayConfig) QueryCompose() map[string]string {
	dictionary := map[string]string{
		"pay_memberid": c.Merchant,
		"currency":     "PKR",
		"pay_time":     strconv.FormatInt(time.Now().Unix(), 10),
	}

	strSign := PaySign(dictionary, c.MD5Key)
	dictionary["sign"] = strSign

	return dictionary
}
