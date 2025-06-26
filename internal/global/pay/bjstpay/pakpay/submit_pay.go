package pakpay

import (
	"fmt"
	"goserver/internal/global/pay/common"
	"goserver/internal/global/pay/entity"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/utils"

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
	getwallet := func() string {
		if order.PayApp == 1001 {
			return "PKR_Easypaisa"
		} else if order.PayApp == 1002 {
			return "PKR_Jazzcash"
		} else {
			return "PKR_Easypaisa"
		}
	}

	dictionary := make(map[string]string)

	dictionary["pay_memberid"] = c.Merchant
	dictionary["pay_orderid"] = order.OrderID
	dictionary["pay_type"] = getwallet()
	dictionary["pay_amount"] = common.FenToYuan(order.Amount)
	dictionary["pay_applytime"] = fmt.Sprintf("%d", utils.BsonNow().Unix())
	dictionary["pay_notifyurl"] = Config.PayNotifyUrl
	dictionary["pay_returnurl"] = "https://www.google.com"
	dictionary["pay_name"] = order.RealName
	dictionary["pay_mobile"] = order.Mobile

	strSign := PaySign(dictionary, c.MD5Key)
	dictionary["pay_sign"] = data.ToUpper(strSign)
	return dictionary
}

func (c *PayConfig) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	result := new(PayOrderRespond)
	// param := make(map[string]string)
	err := jsoniter.Unmarshal(body, result)
	if err != nil {
		return "", err
	}
	// param["code"] = fmt.Sprintf("%d", result.Code)
	// param["message"] = result.Message
	// param["order_no"] = result.Data.Order
	// param["merchant_order_no"] = result.Data.MerchantOrderNo
	// param["merchant_code"] = result.Data.MerchantCode
	// param["amount"] = result.Data.Amount
	// param["payUrl"] = result.Data.PayLink
	// param["extendInfo"] = result.ExtendInfo
	if result.Code != 1000 { // 没成功TODO
		return "", fmt.Errorf("code:%d,message:%s, OrderNo:%s", result.Code, result.Msg, order.OrderID)
	}
	// if !c.VerifySing(result.Data.Sign, SignContent(param), crypto.SHA256) {
	// 	return "", errors.New("verify sign fail")
	// }
	return result.Data.PayUrl, nil
	// return "", nil
}
