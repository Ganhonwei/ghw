package xfpay

import (
	"fmt"
	"goserver/internal/global/pay/service"
)

type XFPayConfig struct {
	MachId             string // appid
	AppId              string // 商户号
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
type XFOrderRespond struct {
	Code       string `json:"code"`
	OrderNo    string `json:"orderNo"`
	PaymentUrl string `json:"paymentUrl"`
}

// 提现响应
type XFWithdrawRespond struct {
	Code   string `json:"code"`
	Status string `json:"status"`
}

// 充值完成响应
type XFRechargeCallback struct {
	MerchantOrderNo string `json:"merchantOrderNo"`
	PayMoney        string `json:"payMoney"`
	PayTime         string `json:"payTime"`
	Status          string `json:"status"`
	Sign            string `json:"sign"`
}

// 代付回调
type XFWithdrawCallback struct {
	MerchantOrderNo string `json:"merchantOrderNo"`
	WithdrawalMoney string `json:"withdrawalMoney"`
	WithdrawalTime  string `json:"withdrawalTime"`
	Status          string `json:"status"`
	Sign            string `json:"sign"`
	// Msg             string `json:"msg"`
}

type XFInspectRespond struct {
	OrderNo         string `json:"orderNo"`
	MerchantOrderNo string `json:"merchantOrderNo"`
	Status          string `json:"status"`
	OrderMoney      string `json:"orderMoney"`
	Remark          string `json:"remark"`
	CreateTime      string `json:"createTime"`
	PayTime         string `json:"payTime"`
}

func InitConfig(url, appid, machid, md5key, withDrawMd5, notifyUrl, withdrawNotifyUrl string) {
	Config = &XFPayConfig{
		MachId:             machid,
		AppId:              appid,
		Md5Key:             md5key,
		WithDrawMDd5Key:    withDrawMd5,
		PayOrderUrl:        fmt.Sprintf("%s%s", url, "/pay/payOrder/yd/pay"),
		WithDrawUrl:        fmt.Sprintf("%s%s", url, "/pay/payOrder/yd/withdrawal"),
		PayNotifyUrl:       fmt.Sprintf("%s%s", notifyUrl, "/xf"),
		WithDrawNotifyUrl:  fmt.Sprintf("%s%s", withdrawNotifyUrl, "/xf"),
		InspectOrderUrl:    fmt.Sprintf("%s%s", url, "/pay/query"),
		InspectWithdrawUrl: fmt.Sprintf("%s%s", url, "/transfer/query"),
	}
	service.Regist(service.XFPAY, Config)
}
