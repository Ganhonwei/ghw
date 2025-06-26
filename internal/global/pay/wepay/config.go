package wepay

import (
	"fmt"
	"goserver/internal/global/pay/service"
)

type WePayConfig struct {
	Appid              string // 商户号
	PayPassageId       string // 代收通道id
	WithdrawPassageId  string // 代付通道id
	Md5Key             string // 支付MD5秘钥
	PayOrderUrl        string // 代收地址
	WithDrawUrl        string // 代付地址
	PayNotifyUrl       string // 支付通知地址
	WithDrawNotifyUrl  string // 提现通知地址
	InspectOrderUrl    string // 代收订单查询
	InspectWithdrawUrl string // 代付订单查询
	QueryUrl           string // 余额查询
}

// 下单响应
type WePayOrderRespond struct {
	Code    int32             `json:"code"`
	Msg     string            `json:"msg"`
	Success bool              `json:"success"`
	Desc    string            `json:"desc"`
	Data    WePayOrderRspData `json:"data"`
}
type WePayOrderRspData struct {
	PayUrl  string `json:"payUrl"`  // 支付地址
	OrderNo string `json:"orderNo"` // 商户单号
	TradeNo string `json:"tradeNo"` // 系统单号
}

// 提现下单响应
type WePayWithdrawRespond struct {
	Code    int32                `json:"code"`
	Msg     string               `json:"msg"`
	Success bool                 `json:"success"`
	Desc    string               `json:"desc"`
	Data    WePayWithdrawRspData `json:"data"`
}
type WePayWithdrawRspData struct {
	Id      string `json:"id"`      // 系统订单号
	OrderNo string `json:"orderNo"` // 商户订单号
}

// 回调通知
type WePayRechargeCallback struct {
	TradeNo     string  `json:"tradeNo"`     // 系统订单号
	OrderNo     string  `json:"orderNo"`     // 商户订单号
	OrderAmount float64 `json:"orderAmount"` // 订单金额
	Amount      float64 `json:"amount"`      // 实际支付金额
	PayStatus   int32   `json:"payStatus"`   // 支付状态,0-订单生成,1-支付成功,2-支付失败	integer(
	PayTime     int64   `json:"payTime"`     // 支付时间
	Charge      float64 `json:"charge"`      // 手续费
	OtherData   string  `json:"otherData"`   // 扩展字段 支付中心回调时会原样返回，不传不返回
	Reverse     bool    `json:"reverse"`     // 是否反转订单
	Remark      string  `json:"remark"`      // 付款失败备注
	Sign        string  `json:"sign"`        // 6a6c386a90ce5e37e8ab22a8e1ef9e5e 签名值，详见签名算法
}

// 代收待付订单查询
type InspectRespond struct {
	Code    int32       `json:"code"`
	Msg     string      `json:"msg"`
	Success bool        `json:"success"`
	Desc    string      `json:"desc"`
	Data    InspectData `json:"data"`
}

type InspectData struct {
	TradeNo     string  `json:"tradeNo"`     // 系统订单号
	OrderNo     string  `json:"orderNo"`     // 商户订单号
	OrderAmount float64 `json:"orderAmount"` // 订单金额
	Amount      float64 `json:"amount"`      // 实际支付金额
	PayStatus   int32   `json:"payStatus"`   // 支付状态,0-订单生成,1-支付成功,2-支付失败
	PayTime     string  `json:"payTime"`     // 支付时间	string(date-time) "2024-04-08 17:15:22"
	Charge      float64 `json:"charge"`      // 手续费
	Utr         string  `json:"utr"`         // 12位交易码
}

// 查余额
type BalanceRespond struct {
	Code    int32       `json:"code"`
	Msg     string      `json:"msg"`
	Success bool        `json:"success"`
	Desc    string      `json:"desc"`
	Data    BalanceData `json:"data"`
}

type BalanceData struct {
	MchId         string  `json:"mchId"`         // 商户号
	BalanceAll    float64 `json:"balanceAll"`    // 总余额
	BalanceUsable float64 `json:"balanceUsable"` // 可用余额
	BalanceIce    float64 `json:"balanceIce"`    // 冻结余额
}

func InitConfig(url, checkUrl, appid, payPassageId, withdrawPassageId, md5key, notifyUrl, withdrawNotifyUrl string) {
	Config = &WePayConfig{
		Appid:              appid,
		PayPassageId:       payPassageId,
		WithdrawPassageId:  withdrawPassageId,
		Md5Key:             md5key,
		PayOrderUrl:        fmt.Sprintf("%s%s", url, "/collect/create"),
		WithDrawUrl:        fmt.Sprintf("%s%s", url, "/pay/create"),
		PayNotifyUrl:       fmt.Sprintf("%s%s", notifyUrl, "/wepay"),
		WithDrawNotifyUrl:  fmt.Sprintf("%s%s", withdrawNotifyUrl, "/wepay"),
		InspectOrderUrl:    fmt.Sprintf("%s%s", checkUrl, "/collect/query"),
		InspectWithdrawUrl: fmt.Sprintf("%s%s", checkUrl, "/pay/query"),
		QueryUrl:           fmt.Sprintf("%s%s", checkUrl, "/order/balance"),
	}
	service.Regist(service.WEPAY, Config)
}
