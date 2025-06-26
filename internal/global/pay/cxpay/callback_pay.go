package cxpay

import (
	"encoding/json"
	"goserver/internal/global/pay/common"
	"goserver/internal/global/pay/service"
	"goserver/pkg/data"
)

func (t *PayConfig) ParseIncomeParamInfo(paramMap map[string]string) (*service.ParamInfo, error) {
	paramInfo := new(service.ParamInfo)
	paramInfo.OrderId = paramMap["orderId"]
	paramInfo.Status = paramMap["status"]
	paramInfo.Message = paramMap["remark"]
	paramInfo.TradeNo = paramMap["platOrderId"]
	paramInfo.Amount = common.YuanToFen(paramMap["amount"])

	if paramInfo.Status != "1" {
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
