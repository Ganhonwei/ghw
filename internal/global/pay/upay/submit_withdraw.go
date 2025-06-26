package upay

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
func (c *PayConfig) WithdrawCompose(order *entity.WithdrawOrder) map[string]any {
	dictionary := make(map[string]any)

	dictionary["version"] = "V1"
	dictionary["appId"] = c.AppId
	dictionary["outTradeNo"] = order.OrderID
	dictionary["integrate"] = "PAYOUT"
	dictionary["currency"] = "INR"
	dictionary["methodType"] = "IMPS"
	dictionary["amount"] = common.FenToYuan(order.Amount)
	dictionary["accountNo"] = order.BankNumber
	dictionary["bankIfsc"] = order.IFSC
	dictionary["payeeInfo"] = map[string]string{
		"payeeName":   order.RealName,
		"payeeMobile": "91" + order.Mobile,
		"payeeEmail":  order.Email,
	}
	dictionary["userId"] = order.Userid
	dictionary["notifyUrl"] = Config.WithDrawNotifyUrl

	strSign := PaySign(dictionary, c.MD5Key)
	dictionary["sign"] = data.ToUpper(strSign)

	return dictionary
}

func (c *PayConfig) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	result := new(PayWithdrawRespond)
	jsoniter.Unmarshal(body, result)

	if result.Status != 200 || !result.Rel { // 没成功
		if strings.Contains(result.Message, "balance") {
			order.OrderStatus = data.DeliveryFailTransfer
		} else if result.Status == 400 ||
			strings.Contains(result.Message, "IFSC") ||
			strings.Contains(result.Message, "IMPS") {
			order.OrderStatus = data.DeliveryFailRefund
		} else {
			order.OrderStatus = data.OrderFail
		}
		return false, result.Message
	}
	order.OrderStatus = data.OrderSuccess
	return true, result.Message
}
