package wypay

import (
	"encoding/json"
	"goserver/internal/global/pay/common"
	"goserver/internal/global/pay/service"
	"goserver/pkg/data"
)

func (t *PayConfig) ParseIncomeParamInfo(paramMap map[string]string) (*service.ParamInfo, error) {
	paramInfo := new(service.ParamInfo)
	paramInfo.OrderId = paramMap["order_out_no"]
	paramInfo.Status = paramMap["order_status"]
	paramInfo.Message = paramMap["message"]
	paramInfo.TradeNo = paramMap["order_no"]
	paramInfo.Amount = common.YuanToFen(paramMap["order_amount"])

	if paramInfo.Status != "20" {
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
