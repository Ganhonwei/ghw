package bjsteocpay

import (
	"encoding/json"
	"goserver/internal/global/pay/common"
	"goserver/internal/global/pay/service"
	"goserver/pkg/data"
)

func (t *PayConfig) ParseIncomeParamInfo(paramMap map[string]string) (*service.ParamInfo, error) {
	paramInfo := new(service.ParamInfo)
	paramInfo.OrderId = paramMap["mch_order_no"]
	paramInfo.Status = paramMap["status"]
	paramInfo.Message = paramMap["error_message"]
	paramInfo.TradeNo = paramMap["platform_order_no"]
	paramInfo.Amount = common.YuanToFen(paramMap["payment_amount"])

	if paramInfo.Status != "PAID" {
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
