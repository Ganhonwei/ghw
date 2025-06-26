package flypay

import (
	"fmt"
	"goserver/internal/global/pay/service"
)

type FlypayConfig struct {
	MachId            string // 商户号
	Md5Key            string // MD5秘钥(代收)
	ServerPublicKey   string // 第三方平台公钥
	PrivateKey        string // 私钥
	PublicKey         string // 公钥
	PayOrderUrl       string // 代收地址
	WithDrawUrl       string // 代付地址
	PayNotifyUrl      string // 支付通知地址
	WithDrawNotifyUrl string // 提现通知地址
	InspectOrderUrl   string // 核单地址
}

// 下单响应
type OrderRespond struct {
	Code         int    `json:"code"`         // 接口状态：1获取成功 0 失败
	Msg          string `json:"msg"`          // 返回错误信息
	Mer_no       string `json:"mer_no"`       // 商户号
	Mer_order_no string `json:"mer_order_no"` // 商户订单号
	Pay_type     string `json:"pay_type"`     // UPI
	Pay_url      string `json:"pay_url"`      // 支付跳转链接
	Order_number string `json:"order_number"` // 平台订单号
	Order_amount int    `json:"order_amount"` // 订单金额 整数
	Pay_amount   string `json:"pay_amount"`   // 实际支付金额（提单100 付200；实际支付金额=200）
	Sign         string `json:"sign"`         // 签名
}

// 提现响应
type WithdrawRespond struct {
	Code         int    `json:"code"`         // 接口状态：1:获取成功 0:失败,提交失败 ; 76:繁忙；75:重复订单号；
	Msg          string `json:"msg"`          // 返回错误信息
	Mer_no       string `json:"mer_no"`       // 商户号
	Mer_order_no string `json:"mer_order_no"` // 商户订单号
	Order_number string `json:"order_number"` // 平台订单号
	Order_amount string `json:"order_amount"` // 订单金额
	Pay_type     string `json:"pay_type"`     // 支付类型
	Cyy_no       string `json:"cyy_no"`       // 币种： 印度=INR
	Acc_no       string `json:"acc_no"`       // 银行卡号
	Acc_name     string `json:"acc_name"`     // 姓名
	Province     string `json:"province"`     // IFSC
	Sign         string `json:"sign"`         // 签名
}

// 充值完成响应
type RechargeCallback struct {
	Code         int    `json:"code"`         // 接口状态：1获取成功 0 失败
	Msg          string `json:"msg"`          // 返回错误信息
	Order_status string `json:"order_status"` // 订单状态:-1=订单关闭,0=待支付,1=支付处理中,2=支付超时,3=支付失败,4=支付成功 ,5=待回调(等待最终确认订单是否成功)
	Mer_no       string `json:"mer_no"`       // 商户号
	Mer_order_no string `json:"mer_order_no"` // 商户订单号
	Order_amount string `json:"order_amount"` // 订单金额
	Pay_amount   string `json:"pay_amount"`   // 实际支付金额
	Order_no     string `json:"order_no"`     // 平台订单号
	Order_time   int64  `json:"order_time"`   // 订单时间
	Sign         string `json:"sign"`         // 签名
}

// 代付回调
type WithdrawCallback struct {
	Code         int    `json:"code"`         // 接口状态：1获取成功 0 失败
	Msg          string `json:"msg"`          // 返回错误信息
	Order_status string `json:"order_status"` // 订单状态:-1=订单异常,0=待处理,1=转账处理中,2=转账拒绝,3=转账失败,4=转账成功,5=转账撤销
	Mer_no       string `json:"mer_no"`       // 商户号
	Mer_order_no string `json:"mer_order_no"` // 商户订单号
	Order_amount string `json:"order_amount"` // 订单金额
	Order_no     string `json:"order_no"`     // 平台订单号
	Order_time   int64  `json:"order_time"`   // 订单时间
	Sign         string `json:"sign"`         // 签名
}

// 核单响应
type InspectResponse struct {
	Code        int    `json:"code"`
	Msg         string `json:"msg"`
	MerNo       string `json:"mer_no"`
	MerOrderNo  string `json:"mer_order_no"`
	OrderNumber string `json:"order_number"`
	PayAmount   string `json:"pay_amount"`
	OrderStatus string `json:"order_status"`
	Sign        string `json:"sign"`
}

func InitConfig(url, appid, md5key, spublickey, privatekey, publickey, notifyUrl, withdrawNotifyUrl string) {
	Config = &FlypayConfig{
		MachId:            appid,
		Md5Key:            md5key,
		ServerPublicKey:   spublickey,
		PrivateKey:        privatekey,
		PublicKey:         publickey,
		PayOrderUrl:       fmt.Sprintf("%s%s", url, "/poi/pay/index/PayOrderCreate"),
		WithDrawUrl:       fmt.Sprintf("%s%s", url, "/poi/dai/index/DaiOrderCreate"),
		PayNotifyUrl:      fmt.Sprintf("%s%s", notifyUrl, "/flypay"),
		WithDrawNotifyUrl: fmt.Sprintf("%s%s", withdrawNotifyUrl, "/flypay"),
		InspectOrderUrl:   fmt.Sprintf("%s%s", url, "/poi/pay/Orderquery"),
	}
	// 注册
	service.Regist(service.FLYPAY, Config)
}
