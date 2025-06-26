package wepay2

import (
	"encoding/json"
	"goserver/internal/global/pay/common"
	"goserver/internal/global/pay/service"
	"goserver/pkg/data"
)

func (t *PayConfig) ParseWithdrawParamInfo(paramMap map[string]string) (*service.ParamInfo, error) {
	paramInfo := new(service.ParamInfo)
	paramInfo.OrderId = paramMap["merTransferId"]
	paramInfo.Status = paramMap["tradeResult"]
	paramInfo.Message = ""
	paramInfo.TradeNo = paramMap["tradeNo"]
	paramInfo.Amount = common.YuanToFen(paramMap["transferAmount"])

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
