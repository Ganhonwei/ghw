package wepay2

import (
	"fmt"
	"goserver/internal/global/pay/common"
	"goserver/internal/global/pay/entity"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"net/url"
	"time"

	jsoniter "github.com/json-iterator/go"
)

// 提现下单
func (t *PayConfig) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	body := t.WithdrawCompose(order) //打包post数据
	// strReq := pay.ToQueryString(body)
	strReq, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strReq)

	values := url.Values{}
	for key, value := range body {
		values.Add(key, fmt.Sprint(value))
	}
	strReq1 := values.Encode()

	resp, err := doHttpForm(t.WithDrawUrl, []byte(strReq1))
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

	loc, _ := time.LoadLocation("Asia/Kolkata")
	now := time.Now().In(loc)
	formatted := now.Format("2006-01-02 15:04:05")

	dictionary["mch_id"] = c.Merchant
	dictionary["mch_transferId"] = order.OrderID
	dictionary["transfer_amount"] = common.FenToYuan(order.Amount)
	dictionary["apply_date"] = formatted
	dictionary["bank_code"] = "IDPT0001"
	dictionary["receive_name"] = order.RealName
	dictionary["receive_account"] = order.BankNumber
	dictionary["remark"] = order.IFSC
	dictionary["back_url"] = Config.WithDrawNotifyUrl

	strSign := PaySign(dictionary, c.WithdrawMD5Key)
	dictionary["sign"] = strSign
	dictionary["sign_type"] = "MD5"

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

	if result.RespCode != "SUCCESS" || result.TradeResult != "0" { // 没成功TODO
		order.OrderStatus = data.OrderFail
		return false, result.ErrorMsg
	}

	// if !c.VerifySing(result.Data.Sign, SignContent(param), crypto.SHA256) {
	// 	return false, fmt.Sprintf("%d", result.Code)
	// }
	order.OrderStatus = data.OrderSuccess
	return true, result.ErrorMsg
}
