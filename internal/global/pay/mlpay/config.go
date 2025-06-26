package mlpay

import (
	"fmt"
	"goserver/internal/global/pay/service"
)

type MlPayConfig struct {
	Appid              string // 商户号
	ApplicationId      string // 应用Id
	Md5Key             string // 支付MD5秘钥
	WithDrawMDd5Key    string // 提现MD5秘钥
	PayOrderUrl        string // 代收地址
	WithDrawUrl        string // 代付地址
	PayNotifyUrl       string // 支付通知地址
	WithDrawNotifyUrl  string // 提现通知地址
	InspectOrderUrl    string // 查单接口
	InspectWithdrawUrl string // 代付查询
}

// 下单响应
type MlOrderRespond struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

// 充值响应数据
type MlPayResult struct {
	Status         int32  `json:"status"`         // 状态 0:失败 1:成功
	ApplicationId  int64  `json:"applicationId"`  // 商户id
	PayWay         int32  `json:"payWay"`         // 支付方式
	PartnerOrderNo string `json:"partnerOrderNo"` // 订单号
	OrderNo        string `json:"orderNo"`        // 平台生成的订单号
	ChannelOrderNo string `json:"channelOrderNo"` // 上游订单号
	Amount         int32  `json:"amount"`         // 金额(分)
	Sign           string `json:"sign"`           // 签名
}

// 核单响应
type MlInspectRespond struct {
	Code    string  `json:"code"`
	Message string  `json:"message"`
	Data    Message `json:"data"`
}

type Message struct {
	PayPartnerId     int64  `json:"payPartnerId"`
	PayApplicationId int64  `json:"payApplicationId"`
	OrderNo          string `json:"orderNo"`
	PartnerOrderNo   string `json:"partnerOrderNo"`
	Amount           string `json:"amount"`
	Status           int32  `json:"status"`
	SuccessTime      string `json:"successTime"`
	CreateTime       string `json:"createTime"`
}

// 代付核单响应
type MlInspectWdRespond struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Data    WihtdrawMessage `json:"data"`
}

type WihtdrawMessage struct {
	PayPartnerId      int64  `json:"payPartnerId"`
	WithdrawNo        string `json:"withdrawNo"`
	PartnerWithdrawNo string `json:"partnerWithdrawNo"`
	Amount            string `json:"amount"`
	Status            int32  `json:"status"`
	SuccessTime       string `json:"successTime"`
	CreateTime        string `json:"createTime"`
	ErrorMsg          string `json:"errorMsg"`
}

func InitConfig(url, appid, applicationId, md5key, withDrawMd5, notifyUrl, withdrawNotifyUrl string) {
	Config = &MlPayConfig{
		Appid:              appid,
		ApplicationId:      applicationId,
		Md5Key:             md5key,
		WithDrawMDd5Key:    withDrawMd5,
		PayOrderUrl:        fmt.Sprintf("%s%s", url, "/pay/order"),
		WithDrawUrl:        fmt.Sprintf("%s%s", url, "/pay/withdraw"),
		PayNotifyUrl:       fmt.Sprintf("%s%s", notifyUrl, "/ml"),
		WithDrawNotifyUrl:  fmt.Sprintf("%s%s", withdrawNotifyUrl, "/ml"),
		InspectOrderUrl:    fmt.Sprintf("%s%s", url, "/pay/queryOrder"),
		InspectWithdrawUrl: fmt.Sprintf("%s%s", url, "/pay/queryWithdraw"),
	}
	service.Regist(3011, Config)
}
