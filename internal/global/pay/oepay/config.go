package oepay

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"goserver/internal/global/pay/service"
	"goserver/pkg/glog"
	"strings"
)

const (
	pem_begin = "-----BEGIN PRIVATE KEY-----\n"
	pem_end   = "\n-----END PRIVATE KEY-----"
	pub_begin = "-----BEGIN PUBLIC KEY-----\n"
	pub_end   = "\n-----END PUBLIC KEY-----"
)

type OePayConfig struct {
	ChannelId          string // 通道id
	KeyServerPub       string // 服务器公钥
	KeyClientPriv      string // 客户端公钥
	KeyClientPub       string // 客户端私钥
	PayUrl             string // API域名
	SuccessUrl         string // 支付成功跳转地址
	FailUrl            string // 支付失败跳转地址
	PayOrderUrl        string // 代收地址
	WithDrawUrl        string // 代付地址
	InspectOrderUrl    string // 代收订单查询
	InspectWithdrawUrl string // 代付订单查询
	QueryUrl           string // 余额查询

	keyServerPub  *rsa.PublicKey
	keyClientPriv *rsa.PrivateKey
	keyClientPub  *rsa.PublicKey
}

func (c *OePayConfig) initKeys() (err error) {
	c.keyServerPub, err = parsePublickKey(c.KeyServerPub)
	if err != nil {
		return
	}
	c.keyClientPriv, err = parsePrivateKey(c.KeyClientPriv)
	if err != nil {
		return
	}
	c.keyClientPub, err = parsePublickKey(c.KeyClientPub)
	if err != nil {
		return
	}
	return
}

// 下单响应
type OePayOrderRespond struct {
	Code string            `json:"code"`
	Msg  string            `json:"msg"`
	Data OePayOrderRspData `json:"data"`
}
type OePayOrderRspData struct {
	LinkUrl         string `json:"linkUrl"`         // 收银台地址
	Status          int32  `json:"status"`          // 代收状态 0支付失败 1支付成功 2待支付 3支付中 4已退款 5待创建
	OrderNo         string `json:"orderNo"`         // 商户订单号
	PlatformOrderNo string `json:"platformOrderNo"` // 平台订单编号
}

// 提现下单响应
type OePayWithdrawRespond struct {
	Code string               `json:"code"`
	Msg  string               `json:"msg"`
	Data OePayWithdrawRspData `json:"data"`
}
type OePayWithdrawRspData struct {
	Status          int32  `json:"status"`          // 代付状态 0失败 1成功 2待支付 3支付中 4已支付等待查询 5资金反转
	PlatformOrderNo string `json:"platformOrderNo"` // 平台订单编号
	OrderNo         string `json:"orderNo"`         // 商户订单号
}

type InspectRespond struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Data InspectData `json:"data"`
}

type InspectData struct {
	PlatformOrderNo string  `json:"platformOrderNo"` // 平台订单编号
	OrderNo         string  `json:"orderNo"`         // 商户订单号
	Status          int32   `json:"status"`          // 代收状态/代付状态
	Amount          float64 `json:"amount"`          // 支付金额
	Fee             float64 `json:"fee"`             // 手续费
	LinkUrl         string  `json:"linkUrl"`
	Udf1            string  `json:"udf1"` // 交易码
	Udf2            string  `json:"udf2"` // 交易码
	Udf3            string  `json:"udf3"` // 交易码
	Udf4            string  `json:"udf4"` // 交易码
	Udf5            string  `json:"udf5"` // 交易码
}

type BalanceRespond struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Data BalanceData `json:"data"`
}

type BalanceData struct {
	DeductAmount  float64 `json:"deductAmount"`  // 支付账户
	PaymentAmount float64 `json:"paymentAmount"` // 代付账户
	FrozenAmount  float64 `json:"frozenAmount"`  // 冻结账户
}

// 回调消息
type OePayRechargeCallback struct {
	PlatformOrderNo string `json:"platformOrderNo"` // 平台订单编号
	OrderNo         string `json:"orderNo"`         // 商户订单号
	Status          string `json:"status"`          // 代收状态/代付状态
	Amount          string `json:"amount"`          // 支付金额
	Fee             string `json:"fee"`             // 手续费
	LinkUrl         string `json:"linkUrl"`         // pay only 支付链接地址
	UtrNo           string `json:"utrNo"`           // withdraw only 代付凭证号
	ErrorMsg        string `json:"errorMsg"`        // withdraw only 失败原因
	Udf1            string `json:"udf1"`            // udf1 自定义字段1
	Udf2            string `json:"udf2"`            // udf2
	Udf3            string `json:"udf3"`            // udf3
	Udf4            string `json:"udf4"`            // udf4
	Udf5            string `json:"udf5"`            // udf5
}

func InitConfig(channelId, keyServerPub, keyClientPriv, keyClientPub, payUrl, successUrl, failUrl string) {
	Config = &OePayConfig{
		ChannelId:          channelId,
		KeyServerPub:       keyServerPub,
		KeyClientPriv:      keyClientPriv,
		KeyClientPub:       keyClientPub,
		PayUrl:             payUrl,
		SuccessUrl:         successUrl,
		FailUrl:            failUrl,
		PayOrderUrl:        fmt.Sprintf("%s%s", payUrl, "/gold-pay/portal/createH5PayLink"),
		WithDrawUrl:        fmt.Sprintf("%s%s", payUrl, "/gold-pay/portal/transfer"),
		InspectOrderUrl:    fmt.Sprintf("%s%s", payUrl, "/gold-pay/portal/queryPayOrderStatus"),
		InspectWithdrawUrl: fmt.Sprintf("%s%s", payUrl, "/gold-pay/portal/queryTransferOrderStatus"),
		QueryUrl:           fmt.Sprintf("%s%s", payUrl, "/gold-pay/portal/queryBalance"),
	}
	if err := Config.initKeys(); err != nil {
		glog.Error("oepay rsa key init error: ", err)
	}
	service.Regist(service.OEPAY, Config)
}

// 生成私钥对象
func parsePrivateKey(key string) (*rsa.PrivateKey, error) {
	if !strings.HasPrefix(key, pem_begin) {
		key = pem_begin + key
	}
	if !strings.HasSuffix(key, pem_end) {
		key = key + pem_end
	}
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return nil, errors.New("key is error")
	}
	// 解析DER编码的私钥，生成私钥对象
	prikey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return prikey.(*rsa.PrivateKey), nil
}

// 生成公钥对象
func parsePublickKey(key string) (*rsa.PublicKey, error) {
	if !strings.HasPrefix(key, pub_begin) {
		key = pub_begin + key
	}
	if !strings.HasSuffix(key, pub_end) {
		key = key + pub_end
	}
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return nil, errors.New("key is error")
	}
	// 解析DER编码的公钥，生成公钥对象
	pubkey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return pubkey.(*rsa.PublicKey), nil
}
