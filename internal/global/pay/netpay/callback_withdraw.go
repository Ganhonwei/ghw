package netpay

import (
	"encoding/json"
	"goserver/internal/global/pay/common"
	"goserver/internal/global/pay/service"
	"goserver/pkg/data"
	"strings"
)

func (t *PayConfig) ParseWithdrawParamInfo(paramMap map[string]string) (*service.ParamInfo, error) {
	paramInfo := new(service.ParamInfo)
	paramInfo.OrderId = paramMap["merOrder"]
	paramInfo.Status = paramMap["status"]
	paramInfo.Message = paramMap["message"]
	paramInfo.TradeNo = paramMap["orderNo"]
	paramInfo.Amount = common.YuanToFen(paramMap["amount"])

	if paramInfo.Status != "2" {
		if strings.Contains(paramInfo.Message, "IFSC") ||
			strings.Contains(paramInfo.Message, "account") ||
			strings.Contains(paramInfo.Message, "IMPS") {
			paramInfo.StatusCode = data.WithdrawFailRefund
		} else if strings.Contains(paramInfo.Message, "Transaction failed") {
			paramInfo.StatusCode = data.WithdrawFailTransfer
		} else {
			paramInfo.StatusCode = data.WithdrawFail
		}
	} else {
		paramInfo.StatusCode = data.WithdrawSuccess
	}

	jsonBytes, err := json.Marshal(paramMap)
	if err != nil {
		return nil, err
	}
	paramInfo.JsonStr = string(jsonBytes)
	return paramInfo, nil
}
