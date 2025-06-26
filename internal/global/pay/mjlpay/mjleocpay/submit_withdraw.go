package mjleocpay

import (
	"fmt"
	"goserver/internal/global/pay/common"
	"goserver/internal/global/pay/entity"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

// 提现下单
func (t *PayConfig) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	param := make(map[string]any)
	body := t.WithdrawCompose(order) //打包post数据
	for k, v := range body {
		param[k] = v
	}
	tempMap := make(map[string]any)
	err := jsoniter.Unmarshal([]byte(body["extra"]), &tempMap)
	if err != nil {
		return nil, err
	}
	param["extra"] = tempMap
	// strReq := pay.ToQueryString(body)
	strReq, _ := jsoniter.Marshal(param)
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

	dictionary["mch_num"] = c.Merchant
	dictionary["currency"] = "BDT"
	dictionary["amount"] = common.FenToYuan(order.Amount)
	dictionary["mch_order_no"] = order.OrderID
	dictionary["notify_url"] = Config.WithDrawNotifyUrl
	dictionary["timestamp"] = fmt.Sprint("", time.Now().UnixMilli())
	phone := order.BankNumber
	extra := map[string]string{
		"channel_type": strings.ToUpper(order.Bank),
	}
	if len(phone) == 10 {
		phone = "0" + phone
	} else if len(phone) == 9 {
		phone = "01" + phone
	}
	extra["mobile"] = phone
	dictionary["extra"], _ = jsoniter.MarshalToString(extra)

	strSign := PaySign(dictionary, c.MD5Key)
	dictionary["sign"] = strSign

	return dictionary
}

func (c *PayConfig) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	result := new(PayWithdrawRespond)
	jsoniter.Unmarshal(body, result)

	if !result.Success || result.Data.Status != "DEDUCTION" { // 没成功TODO
		order.OrderStatus = data.OrderFail
		return false, result.ErrMsg
	}

	order.OrderStatus = data.OrderSuccess
	return true, result.ErrMsg
}
