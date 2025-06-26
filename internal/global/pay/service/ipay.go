package service

import (
	"fmt"
	"goserver/internal/global/pay/entity"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"sort"
	"strings"
)

var ChannelRegister = make(map[uint32]IPay)

type IPay interface {
	Submit(order *entity.PayOrder) ([]byte, error)
	SubmitResponse(body []byte, order *entity.PayOrder) (string, error)
	WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error)
	WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string)
	InspectSubmit(order *entity.PayOrder) ([]byte, error)
	InspectResponse(body []byte, order *entity.PayOrder) error
	InspectWithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error)
	InspectWdResponse(body []byte, order *entity.WithdrawOrder) int
	BalanceQuery() int64
}

// 每个支付对象初始化之后调用此方法
// 注册到map中
func Regist(channel uint32, p IPay) {
	ChannelRegister[channel] = p
}

// 获取通道入口
func GetChannelMent(channel uint32) IPay {
	if c, ok := ChannelRegister[channel]; ok {
		return c
	}
	return nil
}

func PaySignMd5(param map[string]interface{}, md5key string) string {
	keys := make([]string, 0, len(param))
	for k := range param {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var result strings.Builder
	for _, v := range keys {
		if val, ok := param[v]; ok && val != "" {
			vStr := fmt.Sprintf("%v", val)
			result.WriteString(fmt.Sprintf("%s=%s&", v, vStr))
		}
	}

	if md5key != "" {
		result.WriteString(fmt.Sprintf("key=%s", md5key))
	}
	return utils.Md5(result.String())
}

func PaySignMd5String(param map[string]string, md5key string) string {
	keys := make([]string, 0, len(param))
	for k := range param {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var result strings.Builder
	for _, v := range keys {
		if val, ok := param[v]; ok && val != "" {
			vStr := fmt.Sprintf("%v", val)
			result.WriteString(fmt.Sprintf("%s=%s&", v, vStr))
		}
	}

	if md5key != "" {
		result.WriteString(fmt.Sprintf("key=%s", md5key))
	}
	return utils.Md5(result.String())
}

func PaySignMd5StringNoKey(param map[string]string, md5key string) string {
	keys := make([]string, 0, len(param))
	for k := range param {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var result strings.Builder
	for i, v := range keys {
		if val, ok := param[v]; ok && val != "" {
			vStr := fmt.Sprintf("%v", val)
			if i == len(keys)-1 {
				result.WriteString(fmt.Sprintf("%s=%s%s", v, vStr, md5key))
			} else {
				result.WriteString(fmt.Sprintf("%s=%s&", v, vStr))
			}
		}
	}

	glog.Debugf("PaySignMd5StringNoKey:%s", result.String())
	return utils.Md5(result.String())
}
