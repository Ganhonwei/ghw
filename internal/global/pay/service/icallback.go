package service

import "github.com/valyala/fasthttp"

var Channels = make(map[string]IPayCallback)

type ParamInfo struct {
	OrderId    string // 订单号
	Status     string // 订单状态
	ErrCode    string // 错误码
	Message    string // 错误信息
	TradeNo    string // 交易号
	Amount     int    // 订单金额
	StatusCode int    // 处理后的状态码
	JsonStr    string // JSON字符串
}

// 支付回调接口
type IPayCallback interface {
	GetChannelId() uint32
	GetChannelName() string
	CheckMethod(ctx *fasthttp.RequestCtx) bool
	ParseParamMap(ctx *fasthttp.RequestCtx) (map[string]string, error)
	VerifySign(paramMap map[string]string) (bool, error)
	ParseIncomeParamInfo(paramMap map[string]string) (*ParamInfo, error)
	ParseWithdrawParamInfo(paramMap map[string]string) (*ParamInfo, error)
	GetSuccessMsg() string
	// UtrOrderEnable() bool
}

// 注册渠道
func RegistCallback(p IPayCallback) {
	Channels[p.GetChannelName()] = p
}

// 获取渠道
func GetCallback(channel string) IPayCallback {
	return Channels[channel]
}
