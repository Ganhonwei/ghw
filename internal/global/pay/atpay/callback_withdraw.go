package atpay

import (
	"encoding/json"
	"goserver/internal/global/pay/common"
	"goserver/internal/global/pay/service"
	"goserver/pkg/data"
)

func (t *PayConfig) ParseWithdrawParamInfo(paramMap map[string]string) (*service.ParamInfo, error) {
	paramInfo := new(service.ParamInfo)
	paramInfo.OrderId = paramMap["tradeNo"]
	paramInfo.Status = paramMap["status"]
	paramInfo.Message = ""
	paramInfo.TradeNo = paramMap["channelOrderNo"]
	paramInfo.Amount = common.YuanToFen(paramMap["amount"])

	if paramInfo.Status != "1" {
		paramInfo.StatusCode = data.WithdrawFail
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
