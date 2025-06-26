package yunpay

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

	dictionary["mchNo"] = c.Merchant
	dictionary["appId"] = c.AppId
	dictionary["mchOrderNo"] = order.OrderID
	dictionary["amount"] = common.FenToYuan(order.Amount)
	dictionary["currency"] = "INR"
	dictionary["accountNo"] = order.BankNumber
	dictionary["accountName"] = order.RealName
	dictionary["ifscCode"] = order.IFSC
	dictionary["bankName"] = order.Bank
	dictionary["phone"] = order.Mobile
	dictionary["email"] = order.Email
	dictionary["address"] = "address"
	dictionary["transferDesc"] = "transferDesc"
	dictionary["notifyUrl"] = Config.WithDrawNotifyUrl

	strSign := PaySign(dictionary, c.MD5Key)
	dictionary["sign"] = data.ToUpper(strSign)

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
	// param["yunpay"] = result.Data.PayLink
	// param["extendInfo"] = result.ExtendInfo

	if result.Code != 2000 || result.Data.State != 1 { // 没成功TODO
		if result.Code == 6003 || result.Code == 6004 || result.Code == 7000 {
			order.OrderStatus = data.DeliveryFailTransfer
		} else if result.Code == 2011 {
			order.OrderStatus = data.DeliveryFailRefund
		} else {
			order.OrderStatus = data.OrderFail
		}
		return false, result.Msg
	}

	// if !c.VerifySing(result.Data.Sign, SignContent(param), crypto.SHA256) {
	// 	return false, fmt.Sprintf("%d", result.Code)
	// }
	order.OrderStatus = data.OrderSuccess
	return true, result.Msg
}
