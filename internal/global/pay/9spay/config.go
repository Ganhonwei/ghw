package pay9s

import (
	"fmt"
	"goserver/internal/global/pay/service"
)

type Pay9sConfig struct {
	Appid              string // 商户号
	ChargeId           string // 代收通道
	WithdrawId         string // 代付通道
	Md5Key             string // 支付MD5秘钥
	PayOrderUrl        string // 代收地址
	WithDrawUrl        string // 代付地址
	PayNotifyUrl       string // 支付通知地址
	WithDrawNotifyUrl  string // 提现通知地址
	InspectOrderUrl    string // 查单接口
	InspectWithdrawUrl string // 代付查询
	QueryUrl           string // 余额查询
}

// 下单响应
// type Pay9sOrderRespond struct {
// 	Code            int    `json:"code"`
// 	Msg             string `json:"msg"`
// 	MerchantId      string `json:"merchantId"`
// 	MerchantOrderNo string `json:"merchantOrderNo"`
// 	ShouldAmount    string `json:"shouldAmount"`
// 	RealAmount      string `json:"realAmount"`
// 	Amount          string `json:"amount"`
// 	OrderDate       string `json:"orderDate"`
// 	OrderNo         string `json:"orderNo"`
// 	Status          string `json:"status"`
// 	PayUrl          string `json:"payUrl"`
// 	PayUrl2         string `json:"payUrl2"`
// }

// type Pay1916OrderRespondData struct {
// 	MerchantNo string `json:"merchantNo"`
// 	OrderNo    string `json:"orderNo"`
// 	POrderNo   string `json:"pOrderNo"`
// 	PayUrl     string `json:"payUrl"`
// }

// // 充值完成响应
// type Pay1916RechargeCallback struct {
// 	MerchantNo string `json:"merchantNo"`
// 	OrderNo    string `json:"orderNo"`
// 	POrderNo   string `json:"pOrderNo"`
// 	Amount     int    `json:"amount"`
// 	Timestamp  int64  `json:"timestamp"`
// 	State      int    `json:"state"`
// 	Utr        string `json:"utr"`
// 	Sign       string `json:"sign"`
// }

// // 核单响应
// type Pay1916InspectRespond struct {
// 	ErrorCode int            `json:"errorCode"`
// 	ErrorMsg  string         `json:"errorMsg"`
// 	Data      Pay1916Message `json:"data"`
// }

// // 代付核单响应
// type Pay1916InspectWdRespond struct {
// 	ErrorCode int            `json:"errorCode"`
// 	ErrorMsg  string         `json:"errorMsg"`
// 	Data      Pay1916Message `json:"data"`
// }

// type Pay1916Message struct {
// 	MerchantNo string `json:"merchantNo"`
// 	OrderNo    string `json:"orderNo"`
// 	POrderNo   string `json:"pOrderNo"`
// 	Amount     int    `json:"amount"`
// 	State      int    `json:"state"`
// 	Msg        string `json:"msg"`
// }

func InitConfig(url, appid, chargeId, withdrawId, md5key, notifyUrl, withdrawNotifyUrl string) {
	Config = &Pay9sConfig{
		Appid:              appid,
		Md5Key:             md5key,
		ChargeId:           chargeId,
		WithdrawId:         withdrawId,
		PayOrderUrl:        fmt.Sprintf("%s%s", url, "/api/payIn"),
		WithDrawUrl:        fmt.Sprintf("%s%s", url, "/api/payOut"),
		PayNotifyUrl:       fmt.Sprintf("%s%s", notifyUrl, "/9spay"),
		WithDrawNotifyUrl:  fmt.Sprintf("%s%s", withdrawNotifyUrl, "/9spay"),
		InspectOrderUrl:    fmt.Sprintf("%s%s", url, "/api/payIn/query"),
		InspectWithdrawUrl: fmt.Sprintf("%s%s", url, "/api/payOut/query"),
		QueryUrl:           fmt.Sprintf("%s%s", url, "/api/queryBalance"),
	}
	service.Regist(service.PAY9S, Config)
}
