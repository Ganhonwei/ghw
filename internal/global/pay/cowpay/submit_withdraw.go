package cowpay

import (
	"goserver/internal/global/pay/common"
	"goserver/internal/global/pay/entity"
	"goserver/pkg/data"
	"goserver/pkg/glog"

	jsoniter "github.com/json-iterator/go"
)

// 提现下单
func (t *PayConfig) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	body := t.WithdrawCompose(order) //打包post数据
	// strReq := pay.ToQueryString(body)
	strReq, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strReq)
	resp, err := doHttpForm(t.WithDrawUrl, strReq)
	//fmt.Printf("resp %s, err %v", string(resp), err)
	if err != nil {
		return nil, err
	}
	glog.Infof("WithdrawSubmit order success, id:%s, resp:%s", order.OrderID, string(resp))
	return resp, nil
}

// 提现组装参数
func (c *PayConfig) WithdrawCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)

	dictionary["merchant_code"] = c.Merchant
	dictionary["order_no"] = order.OrderID
	dictionary["order_amount"] = common.FenToYuan(order.Amount)
	dictionary["pay_type"] = "india-bank-repay"
	dictionary["bank_name"] = order.Bank
	dictionary["bank_card"] = order.BankNumber
	dictionary["bank_branch"] = order.IFSC
	dictionary["user_name"] = order.RealName
	dictionary["notify_url"] = Config.WithDrawNotifyUrl

	strSign := PaySign(dictionary, c.MD5Key)

	transdata, _ := jsoniter.Marshal(dictionary)

	dictionary2 := make(map[string]string)
	dictionary2["transdata"] = data.ToUrlEncode(string(transdata))

	dictionary2["sign"] = data.ToUrlEncode(data.ToUpper(strSign))
	dictionary2["signtype"] = "MD5"

	return dictionary2
}

func (c *PayConfig) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	result := new(PayWithdrawRespond)
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

	if result.Status != true { // 没成功TODO
		order.OrderStatus = data.OrderFail
		return false, result.Msg
	}

	// if !c.VerifySing(result.Data.Sign, SignContent(param), crypto.SHA256) {
	// 	return false, fmt.Sprintf("%d", result.Code)
	// }
	order.OrderStatus = data.OrderSuccess
	return true, result.Msg
}
