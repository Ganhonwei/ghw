package upay

import (
	"fmt"
	"goserver/internal/global/pay/entity"
	"goserver/pkg/data"
	"goserver/pkg/glog"

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
func (c *PayConfig) compose(order *entity.PayOrder) map[string]any {
	dictionary := make(map[string]any)

	dictionary["version"] = "V1"
	dictionary["appId"] = c.AppId
	dictionary["outTradeNo"] = order.OrderID
	dictionary["integrate"] = "CASHIER"
	dictionary["currency"] = "INR"
	dictionary["amount"] = order.Amount / 100
	dictionary["paymentInfo"] = map[string]string{
		"methodType": "UPI",
	}
	dictionary["goodsInfo"] = map[string]string{
		"goodsId":   "1001",
		"goodsName": fmt.Sprintf("recharge %d", order.Amount/100),
	}
	dictionary["userInfo"] = map[string]string{
		"id":     order.Userid,
		"name":   order.RealName,
		"mobile": "91" + order.Mobile,
		"email":  order.Email,
	}
	dictionary["notifyUrl"] = Config.PayNotifyUrl
	dictionary["frontCallbackUrl"] = "https://www.google.com"

	strSign := PaySign(dictionary, c.MD5Key)
	dictionary["sign"] = data.ToUpper(strSign)
	return dictionary
}

func (c *PayConfig) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	result := new(PayOrderRespond)
	jsoniter.Unmarshal(body, result)
	if result.Status != 200 || !result.Rel { // 没成功TODO
		return "", fmt.Errorf("code:%d,message:%s, OrderNo:%s", result.Status, result.Message, result.Data.OutTradeNo)
	}
	order.OutTradeNo = result.Data.PlatTradeNo
	return result.Data.CashierUrl, nil
}
