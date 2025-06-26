package tansafe

import (
	"fmt"
)

type TansafePayConfig struct {
	MachId            string // 商户号
	ChannelId         string // 交易通道
	appId             string // 商户aapid
	ServerPublicKey   string // 第三方平台公钥
	PrivateKey        string // 私钥
	PublicKey         string // 公钥
	PayOrderUrl       string // 代收地址
	WithDrawUrl       string // 代付地址
	PayNotifyUrl      string // 支付通知地址
	WithDrawNotifyUrl string // 提现通知地址
}

// 下单响应
type OrderRespond struct {
	Status int        `json:"status"`
	Msg    string     `json:"msg"`
	Data   *OrderData `json:"data"`
}
type OrderData struct {
	MchOrderNo  string `json:"mchOrderNo"`  // 商户订单号
	PlatOrderNo string `json:"platOrderNo"` // 平台订单号
	ApplyTime   int    `json:"applyTime"`   // 申请时间
	Amount      int    `json:"amount"`      // 实际支付金额
	Reference   string `json:"reference"`   // 参考码
	Link        string `json:"link"`        // 支付链接
	ExpireTime  int    `json:"expireTime"`  // 过期时间
}

// 下单回调
type OrderCallback struct {
	Status      int    `json:"status"`      // 状态
	MemberId    string `json:"memberId"`    // 商户Id
	MchOrderNo  string `json:"mchOrderNo"`  // 商户订单号
	PlatOrderNo string `json:"platOrderNo"` // 平台订单号
	OrderStatus string `json:"orderStatus"` // 订单状态
	Amount      int    `json:"amount"`      // 到账金额
	Fee         int    `json:"fee"`         // 手续费
	TimeEnd     int64  `json:"timeEnd"`     // 完结时间戳
	Msg         string `json:"msg"`         // 信息
	Sign        string `json:"sign"`        // 签名
}

// 提现响应
type WithdrawRespond struct {
	Status int           `json:"status"`
	Msg    string        `json:"msg"`
	Data   *WithdrawData `json:"data"`
}
type WithdrawData struct {
	MchOrderNo  string `json:"mchOrderNo"`  // 商户订单号
	PlatOrderNo string `json:"platOrderNo"` // 平台订单号
	ApplyTime   int    `json:"applyTime"`   // 申请时间
	Amount      int    `json:"amount"`      // 到账金额
	OrderStatus string `json:"orderStatus"` // 订单状态
}

// 提现回调
type WithdrawCallback struct {
	Status      int    `json:"status"`      // 状态
	MemberId    string `json:"memberId"`    // 商户Id
	MchOrderNo  string `json:"mchOrderNo"`  // 商户订单号
	PlatOrderNo string `json:"platOrderNo"` // 平台订单号
	OrderStatus string `json:"orderStatus"` // 订单状态
	Amount      int    `json:"amount"`      // 到账金额
	Fee         int    `json:"fee"`         // 手续费
	TimeEnd     int64  `json:"timeEnd"`     // 完结时间戳
	Msg         string `json:"msg"`         // 信息
	Sign        string `json:"sign"`        // 签名
}

func InitConfig(url, machid, channelid, appid, serverpublickey, privatekey, publickey, notifyUrl, withdrawNotifyUrl string) {
	Config = &TansafePayConfig{
		MachId:            machid,
		ChannelId:         channelid,
		appId:             appid,
		ServerPublicKey:   serverpublickey,
		PrivateKey:        privatekey,
		PublicKey:         publickey,
		PayOrderUrl:       fmt.Sprintf("%s%s", url, "/payment/india/order/create"),
		WithDrawUrl:       fmt.Sprintf("%s%s", url, "/payout/india/order/create"),
		PayNotifyUrl:      fmt.Sprintf("%s%s", notifyUrl, "/tansafe"),
		WithDrawNotifyUrl: fmt.Sprintf("%s%s", withdrawNotifyUrl, "/tansafe"),
	}
}
