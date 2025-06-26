package bjsteocpay

import (
	"encoding/json"
	"goserver/internal/global/pay/common"
	"goserver/internal/global/pay/service"
	"goserver/pkg/data"
	"strings"
)

func (t *PayConfig) ParseWithdrawParamInfo(paramMap map[string]string) (*service.ParamInfo, error) {
	paramInfo := new(service.ParamInfo)
	paramInfo.OrderId = paramMap["mch_order_no"]
	paramInfo.Status = paramMap["status"]
	paramInfo.Message = paramMap["error_message"]
	paramInfo.ErrCode = paramMap["error_code"]
	paramInfo.TradeNo = paramMap["platform_order_no"]
	paramInfo.Amount = common.YuanToFen(paramMap["order_amount"])

	if paramInfo.Status != "SUCCESS" {
		if strings.Contains(paramInfo.Message, "IFSC") ||
			strings.Contains(paramInfo.Message, "account") {
			paramInfo.StatusCode = data.WithdrawFailRefund
		} else if strings.Contains(paramInfo.Message, "NPCI") ||
			strings.Contains(paramInfo.Message, "balance") {
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
