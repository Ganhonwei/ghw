package bjsteocpay

import (
	"fmt"
	"goserver/internal/global/pay/common"
	"goserver/internal/global/pay/entity"
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

	dictionary["mch_num"] = c.Merchant
	dictionary["currency"] = "PKR"
	dictionary["amount"] = common.FenToYuan(order.Amount)
	dictionary["mch_order_no"] = order.OrderID
	dictionary["notify_url"] = Config.PayNotifyUrl
	dictionary["timestamp"] = fmt.Sprint("", time.Now().UnixMilli())

	strSign := PaySign(dictionary, c.MD5Key)
	dictionary["sign"] = strSign
	return dictionary
}

func (c *PayConfig) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	result := new(PayOrderRespond)
	jsoniter.Unmarshal(body, result)
	if !result.Success || result.Data.PaymentUrl == "" { // 没成功TODO
		return "", fmt.Errorf("code:%s,message:%s, OrderNo:%s", result.ErrCode, result.ErrMsg, result.Data.MchOrderNo)
	}
	order.OutTradeNo = result.Data.PlatformOrderNo
	return result.Data.PaymentUrl, nil
}
