package mepay

import (
	"encoding/json"
	"goserver/internal/global/pay/common"
	"goserver/internal/global/pay/service"
	"goserver/pkg/data"
	"strings"
)

func (t *PayConfig) ParseWithdrawParamInfo(paramMap map[string]string) (*service.ParamInfo, error) {
	paramInfo := new(service.ParamInfo)
	paramInfo.OrderId = paramMap["mchOrderNo"]
	paramInfo.Status = paramMap["state"]
	paramInfo.Message = paramMap["errMsg"]
	paramInfo.ErrCode = paramMap["errCode"]
	paramInfo.TradeNo = paramMap["transferId"]
	paramInfo.Amount = common.FenToFenInt(paramMap["amount"])

	if paramInfo.Status != "2" {
		if strings.Contains(paramInfo.Message, "IMPS") ||
			strings.Contains(paramInfo.Message, "frozen") {
			paramInfo.StatusCode = data.WithdrawFailRefund
		} else if strings.Contains(paramInfo.Message, "balance") ||
			strings.Contains(paramInfo.Message, "NPCI") {
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
