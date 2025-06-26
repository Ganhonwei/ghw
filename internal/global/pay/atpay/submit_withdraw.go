package atpay

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

	dictionary["mid"] = c.Merchant
	dictionary["tradeNo"] = order.OrderID
	dictionary["amount"] = common.FenToYuan(order.Amount)
	dictionary["accountType"] = "0"
	dictionary["certNo"] = "certNo"
	dictionary["cardName"] = order.RealName
	dictionary["cardNo"] = order.BankNumber
	dictionary["bankName"] = "bankName"
	dictionary["notifyUrl"] = Config.WithDrawNotifyUrl
	dictionary["currency"] = "INR"

	strSign := PaySign(dictionary, c.MD5Key)
	dictionary["sign"] = strSign

	dictionary["ifscCode"] = order.IFSC

	return dictionary
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

	if result.Code != "0" || result.Data.Status != 0 { // 没成功TODO
		order.OrderStatus = data.OrderFail
		return false, result.Msg
	}

	// if !c.VerifySing(result.Data.Sign, SignContent(param), crypto.SHA256) {
	// 	return false, fmt.Sprintf("%d", result.Code)
	// }
	order.OrderStatus = data.OrderSuccess
	return true, result.Msg
}
