package leopay

import (
	"goserver/internal/global/pay/common"
	"goserver/internal/global/pay/entity"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"strings"

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

	dictionary["merchantCode"] = c.Merchant
	dictionary["orderNo"] = order.OrderID
	dictionary["type"] = "1"
	dictionary["accNo"] = order.BankNumber
	dictionary["accName"] = order.RealName
	dictionary["amount"] = common.FenToYuan(order.Amount)
	dictionary["bankName"] = "MTN"
	dictionary["ccy_no"] = "INR"
	dictionary["notifyUrl"] = Config.WithDrawNotifyUrl
	dictionary["summary"] = order.IFSC
	dictionary["iphone"] = order.Mobile
	dictionary["bankBranch"] = "ZM0001"

	strSign := PaySign(dictionary, c.MD5Key)
	dictionary["signature"] = strSign

	return dictionary
}

func (c *PayConfig) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	result := new(PayWithdrawRespond)
	jsoniter.Unmarshal(body, result)

	if result.Status != "SUCCESS" { // 没成功TODO
		if strings.Contains(result.ErrMsg, "余额不足") {
			order.OrderStatus = data.DeliveryFailTransfer
		} else {
			order.OrderStatus = data.OrderFail
		}
		return false, result.ErrMsg
	}
	order.OrderStatus = data.OrderSuccess
	return true, result.ErrMsg
}
