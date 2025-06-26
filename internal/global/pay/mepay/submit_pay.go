package mepay

import (
	"fmt"
	"goserver/internal/global/pay/common"
	"goserver/internal/global/pay/entity"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"time"

	jsoniter "github.com/json-iterator/go"
)

// Submit 下单
func (t *PayConfig) Submit(order *entity.PayOrder) ([]byte, error) {
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

// 组装参数
func (c *PayConfig) compose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["mchNo"] = c.Merchant
	dictionary["appId"] = c.AppId
	dictionary["mchOrderNo"] = order.OrderID
	dictionary["mchUserId"] = order.Userid
	dictionary["amount"] = common.FenToFenStr(order.Amount)
	dictionary["countryId"] = "10022"
	dictionary["payMethod"] = "indiaupi"
	dictionary["currency"] = "INR"
	dictionary["subject"] = "bigwin"
	dictionary["body"] = "bigwin"
	dictionary["returnUrl"] = "https://www.google.com"
	dictionary["notifyUrl"] = Config.PayNotifyUrl
	dictionary["reqTime"] = fmt.Sprintf("%d", time.Now().UnixMilli())
	dictionary["version"] = "1.0"
	dictionary["signType"] = "MD5"

	strSign := PaySign(dictionary, c.MD5Key)
	dictionary["sign"] = data.ToUpper(strSign)
	return dictionary
}

func (c *PayConfig) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	result := new(PayOrderRespond)
	// param := make(map[string]string)
	jsoniter.Unmarshal(body, result)
	// param["code"] = fmt.Sprintf("%d", result.Code)
	// param["message"] = result.Message
	// param["order_no"] = result.Data.Order
	// param["merchant_order_no"] = result.Data.MerchantOrderNo
	// param["merchant_code"] = result.Data.MerchantCode
	// param["amount"] = result.Data.Amount
	// param["payUrl"] = result.Data.PayLink
	// param["extendInfo"] = result.ExtendInfo
	if result.Code != 0 || result.Data.OrderState != 1 { // 没成功TODO
		return "", fmt.Errorf("code:%d,message:%s, OrderNo:%s", result.Code, result.Msg, result.Data.MchOrderNo)
	}
	// if !c.VerifySing(result.Data.Sign, SignContent(param), crypto.SHA256) {
	// 	return "", errors.New("verify sign fail")
	// }
	return result.Data.PayData, nil
	// return "", nil
}
