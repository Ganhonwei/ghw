package pay66

import (
	"encoding/json"
	"goserver/internal/global/pay/common"
	"goserver/internal/global/pay/service"
	"goserver/pkg/data"
)

func (t *PayConfig) ParseWithdrawParamInfo(paramMap map[string]string) (*service.ParamInfo, error) {
	paramInfo := new(service.ParamInfo)
	paramInfo.OrderId = paramMap["orderNo"]
	paramInfo.Status = paramMap["payResult"]
	paramInfo.Message = ""
	paramInfo.TradeNo = paramMap["ptOrderNo"]
	paramInfo.Amount = common.YuanToFen(paramMap["payAmount"])

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
