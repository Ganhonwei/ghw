package mjl66pay

import (
	"fmt"
	"goserver/internal/global/pay/common"
	"goserver/internal/global/pay/entity"
	"goserver/pkg/glog"
	"net/url"
	"time"

	jsoniter "github.com/json-iterator/go"
)

// Submit 下单
func (t *PayConfig) Submit(order *entity.PayOrder) ([]byte, error) {
	body := t.compose(order) //打包post数据
	// strReq := pay.ToQueryString(body)
	strReq, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strReq)

	values := url.Values{}
	for key, value := range body {
		values.Add(key, fmt.Sprint(value))
	}
	strReq1 := values.Encode()
	//fmt.Printf("body %s, order %#v\n", body, order)
	resp, err := doHttpForm(t.PayOrderUrl, []byte(strReq1))
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

	dictionary["mer_no"] = c.Merchant
	dictionary["order_no"] = order.OrderID
	dictionary["pay_code"] = "2601"
	dictionary["currency"] = "BDT"
	dictionary["order_amount"] = common.FenToYuan(order.Amount)
	dictionary["order_date"] = time.Now().Format("2006-01-02 15:04:05")
	dictionary["notifyUrl"] = Config.PayNotifyUrl

	strSign := PaySign(dictionary, c.MD5Key)
	dictionary["sign"] = strSign
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
	if result.Code != "SUCCESS" { // 没成功TODO
		return "", fmt.Errorf("code:%s,message:%s, OrderNo:%s", result.Code, result.Msg, result.OrderId)
	}
	// if !c.VerifySing(result.Data.Sign, SignContent(param), crypto.SHA256) {
	// 	return "", errors.New("verify sign fail")
	// }
	return result.PayUrl, nil
	// return "", nil
}
