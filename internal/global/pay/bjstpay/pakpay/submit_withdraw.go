package pakpay

import (
	"errors"
	"goserver/internal/global/pay/common"
	"goserver/internal/global/pay/entity"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

// 提现下单
func (t *PayConfig) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	_, ok := BankCode[order.Bank]
	if !ok {
		glog.Errorf("WithdrawSubmit order bankcode err, id:%s, bank:%s", order.OrderID, order.Bank)
		return nil, errors.New("bank not found")
	}

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
	model, ok := ChannelCode[order.Bank]
	if !ok {
		model = "PKR_BankAccount"
	}
	bankCode := BankCode[order.Bank]
	phone, _ := strings.CutPrefix(order.Mobile, "+")

	dictionary["pay_memberid"] = c.Merchant
	dictionary["pay_orderid"] = order.OrderID
	dictionary["model"] = model
	dictionary["pay_amount"] = common.FenToYuan(order.Amount)
	dictionary["pay_notifyurl"] = Config.WithDrawNotifyUrl
	dictionary["currency"] = "PKR"
	dictionary["pay_name"] = strings.TrimSpace(order.RealName)
	dictionary["pay_mobile"] = phone
	dictionary["account_number"] = order.BankNumber
	dictionary["bank_code"] = bankCode

	strSign := PaySign(dictionary, c.MD5Key)
	dictionary["pay_sign"] = strings.ToUpper(strSign)

	return dictionary
}

func (c *PayConfig) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	result := new(PayWithdrawRespond)
	jsoniter.Unmarshal(body, result)

	if result.Code != 1000 || result.Data.OrderStatus == "failed" { // 没成功TODO
		order.OrderStatus = data.OrderFail
		return false, result.Msg
	}

	order.OrderStatus = data.OrderSuccess
	return true, result.Msg
}
