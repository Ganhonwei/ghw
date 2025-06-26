package upay

import (
	"encoding/json"
	"goserver/internal/global/pay/common"
	"goserver/internal/global/pay/service"
	"goserver/pkg/data"
	"strings"
)

func (t *PayConfig) ParseWithdrawParamInfo(paramMap map[string]string) (*service.ParamInfo, error) {
	paramInfo := new(service.ParamInfo)
	paramInfo.OrderId = paramMap["outTradeNo"]
	paramInfo.Status = paramMap["payStatus"]
	paramInfo.Message = paramMap["failMessage"]
	paramInfo.ErrCode = paramMap["errCode"]
	paramInfo.TradeNo = paramMap["platTradeNo"]
	paramInfo.Amount = common.YuanToFen(paramMap["amount"])

	if paramInfo.Status != "Success" {
		if strings.Contains(paramInfo.Message, "IFSC") ||
			strings.Contains(paramInfo.Message, "account") {
			paramInfo.StatusCode = data.WithdrawFailRefund
		} else if strings.Contains(paramInfo.Message, "NPCI") {
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
