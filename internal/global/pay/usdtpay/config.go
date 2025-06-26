package usdtpay

import "goserver/internal/global/pay/service"

const (
	OrderStatusExpired = 3
	OrderStatusSuccess = 2
	OrderStatusWaiting = 1
)

type UsdtPayConfig struct {
	Md5Key            string // 支付MD5秘钥
	PayOrderUrl       string // 代收地址
	WithDrawUrl       string // 代付地址
	PayNotifyUrl      string // 支付通知地址
	WithDrawNotifyUrl string // 提现通知地址
	RedirectUrl       string // 充值重定向地址
}

// 下单响应
type UsdtPayOrderRespond struct {
	Code      int32               `json:"status_code"`
	Msg       string              `json:"message"`
	RequestId string              `json:"request_id"`
	Data      UsdtPayOrderRspData `json:"data"`
}
type UsdtPayOrderRspData struct {
	TradeId        string  `json:"trade_id"`        // 交易号
	OrderId        string  `json:"order_id"`        // 请求支付订单号
	Amount         float64 `json:"amount"`          // 请求支付金额 CNY,保留2位小数
	ActualAmount   string  `json:"actual_amount"`   // 实际需要支付的金额 USDT,保留四位小数
	Token          string  `json:"token"`           // 钱包地址
	ExpirationTime int64   `json:"expiration_time"` // 过期时间 时间戳秒
	PaymentUrl     string  `json:"payment_url"`     // 收银台地址
}

// 提现下单响应
type UsdtPayWithdrawRespond struct {
	Code      int32                  `json:"status_code"`
	Msg       string                 `json:"message"`
	RequestId string                 `json:"request_id"`
	Data      UsdtPayWithdrawRspData `json:"data"`
}
type UsdtPayWithdrawRspData struct {
	TradeId        string  `json:"trade_id"`        // 交易号
	OrderId        string  `json:"order_id"`        // 请求支付订单号
	Amount         float64 `json:"amount"`          // 请求支付金额 CNY,保留2位小数
	ActualAmount   string  `json:"actual_amount"`   // 实际需要支付的金额 USDT,保留四位小数
	ExpirationTime int64   `json:"expiration_time"` // 过期时间 时间戳秒
}

type UsdtPayRechargeCallback struct {
	TradeId            string  `json:"trade_id"`             // 交易号
	OrderId            string  `json:"order_id"`             // 请求支付订单号
	Amount             float64 `json:"amount"`               // 请求支付金额 CNY,保留2位小数
	ActualAmount       string  `json:"actual_amount"`        // 实际需要支付的金额 USDT,保留四位小数
	Token              string  `json:"token"`                // 钱包地址
	BlockTransactionId string  `json:"block_transaction_id"` // 区块交易号
	Signature          string  `json:"signature"`            // 签名
	Status             int     `json:"status"`               // 订单状态 1：等待支付，2：支付成功，3：已过期；提现:3成功
}

func InitConfig(md5key, payOrderUrl, withDarwUrl, payNotifyUrl, withDrawNotifyUrl, redirectUrl string) {
	Config = &UsdtPayConfig{
		Md5Key:            md5key,
		PayOrderUrl:       payOrderUrl,
		WithDrawUrl:       withDarwUrl,
		PayNotifyUrl:      payNotifyUrl,
		WithDrawNotifyUrl: withDrawNotifyUrl,
		RedirectUrl:       redirectUrl,
	}
	service.Regist(service.USDT, Config)
}
