package kkpluspay

import (
	"fmt"
	"goserver/internal/global/pay/service"
)

type KKPlusPayConfig struct {
	Merchant           string // 商户号
	ServerPubRSA       string // 服务端公钥
	ClientPriRSA       string // 客户端私钥
	MD5Key             string // md5key
	PayOrderUrl        string // 代收地址
	WithDrawUrl        string // 代付地址
	PayNotifyUrl       string // 支付通知地址
	WithDrawNotifyUrl  string // 提现通知地址
	InspectOrderUrl    string // 查单接口
	InspectWithdrawUrl string // 代付查询
	QueryUrl           string // 余额查询
}

type KKPlusOrderRespond struct {
	MerNo       string `json:"mer_no"`
	MerOrderNo  string `json:"mer_order_no"`
	OrderAmount string `json:"order_amount"`
	BusiCode    string `json:"busi_code"`
	NotifyUrl   string `json:"notifyUrl"`
	PageUrl     string `json:"pageUrl"`
	OrderNo     string `json:"order_no"`
	OrderTime   string `json:"order_time"`
	Status      string `json:"status"`
	OrderData   string `json:"order_data"`
	Pname       string `json:"pname"`
	Pemail      string `json:"pemail"`
	Phone       string `json:"phone"`
	CcyNo       string `json:"ccy_no"`
	Sign        string `json:"sign"`
	ErrCode     string `json:"err_code"`
	ErrMsg      string `json:"err_msg"`
}

type KKPlusWithdrawRespond struct {
	Status      string `json:"status"`
	ErrCode     string `json:"err_code"`
	ErrMsg      string `json:"err_msg"`
	MerNo       string `json:"mer_no"`
	MerOrderNo  string `json:"mer_order_no"`
	OrderNo     string `json:"order_no"`
	AccountNo   string `json:"account_no"`
	AccNo       string `json:"acc_no"`
	AccName     string `json:"acc_name"`
	CcyNo       string `json:"ccy_no"`
	OrderAmount string `json:"order_amount"`
	Summary     string `json:"summary"`
}

// 代收回调
type KKPlusRechargeCallback struct {
	MerchantCode    string `json:"merchant_code"`     //商户号
	OrderId         string `json:"order_no"`          //平台订单号
	MerchantOrderNo string `json:"merchant_order_no"` //商户订单号
	Amount          string `json:"amount"`            //金额
	Status          int32  `json:"status"`            //订单状态
	ErrorMessage    string `json:"error_message"`     //错误信息
	Timestamp       int64  `json:"timestamp"`         //时间
	Sign            string `json:"sign"`              //签名
}

// 核单响应
type KKPlusPayInspectRespond struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    KKPlusPayMessage `json:"data"`
}

type KKPlusPayMessage struct {
	OrderNo         string `json:"order_no"`
	MerchantOrderNo string `json:"merchant_order_no"`
	MerchantCode    string `json:"merchant_code"`
	Amount          string `json:"amount"`
	Status          int    `json:"status"`
	AccountCode     string `json:"account_code"`
	PayerName       string `json:"payer_name"`
	Timestamp       int    `json:"timestamp"`
	Sign            int    `json:"sign"`
}

type KKPlusBalanceRespond struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    BalanceMessage `json:"data"`
}

type BalanceMessage struct {
	MerchantCode      string `json:"merchant_code"`
	AvailableBalance  string `json:"available_balance"`
	UnsettledAmount   string `json:"unsettled_amount"`
	WithdrawingAmount string `json:"withdrawing_amount"`
	FrozenAmount      string `json:"frozen_amount"`
	Sign              string `json:"sign"`
}

func InitConfig(payurl, withdrawurl, appid, serverPubRSA, clientPriRSA, md5key, notifyUrl, withdrawNotifyUrl string) {
	Config = &KKPlusPayConfig{
		Merchant:          appid,
		ServerPubRSA:      serverPubRSA,
		ClientPriRSA:      clientPriRSA,
		MD5Key:            md5key,
		PayOrderUrl:       payurl,
		WithDrawUrl:       withdrawurl,
		PayNotifyUrl:      fmt.Sprintf("%s%s", notifyUrl, "/kkplus"),
		WithDrawNotifyUrl: fmt.Sprintf("%s%s", withdrawNotifyUrl, "/kkplus"),
		// InspectOrderUrl:    fmt.Sprintf("%s%s", url, "/pay/query"),
		// InspectWithdrawUrl: fmt.Sprintf("%s%s", url, "/transfer/query"),
		// QueryUrl:           fmt.Sprintf("%s%s", url, "/balance"),
	}
	service.Regist(service.KKPLUSPAY, Config)
}
