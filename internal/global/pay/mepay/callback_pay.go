package mepay

import (
	"encoding/json"
	"goserver/internal/global/pay/common"
	"goserver/internal/global/pay/service"
	"goserver/pkg/data"
)

func (t *PayConfig) ParseIncomeParamInfo(paramMap map[string]string) (*service.ParamInfo, error) {
	paramInfo := new(service.ParamInfo)
	paramInfo.OrderId = paramMap["mchOrderNo"]
	paramInfo.Status = paramMap["state"]
	paramInfo.Message = paramMap["errMsg"]
	paramInfo.TradeNo = paramMap["payOrderId"]
	paramInfo.Amount = common.FenToFenInt(paramMap["amount"])

	if paramInfo.Status != "2" {
		paramInfo.StatusCode = data.PayFail
	} else {
		paramInfo.StatusCode = data.PaySuccess
	}

	jsonBytes, err := json.Marshal(paramMap)
	if err != nil {
		return nil, err
	}
	paramInfo.JsonStr = string(jsonBytes)
	return paramInfo, nil
}
