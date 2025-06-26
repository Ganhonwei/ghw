package pay1916

import (
	"fmt"
	"goserver/internal/global/pay/service"
)

type Pay1916Config struct {
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
	QueryUrl           string // 余额查询地址
}

// 下单响应
type Pay1916OrderRespond struct {
	ErrorCode int    `json:"errorCode"`
	ErrorMsg  string `json:"errorMsg"`
	Data      Pay1916OrderRespondData
}

type Pay1916OrderRespondData struct {
	MerchantNo string `json:"merchantNo"`
	OrderNo    string `json:"orderNo"`
	POrderNo   string `json:"pOrderNo"`
	PayUrl     string `json:"payUrl"`
}

// 充值完成响应
type Pay1916RechargeCallback struct {
	MerchantNo string `json:"merchantNo"`
	OrderNo    string `json:"orderNo"`
	POrderNo   string `json:"pOrderNo"`
	Amount     int    `json:"amount"`
	Timestamp  int64  `json:"timestamp"`
	State      int    `json:"state"`
	Utr        string `json:"utr"`
	Sign       string `json:"sign"`
}

// 核单响应
type Pay1916InspectRespond struct {
	ErrorCode int            `json:"errorCode"`
	ErrorMsg  string         `json:"errorMsg"`
	Data      Pay1916Message `json:"data"`
}

// 代付核单响应
type Pay1916InspectWdRespond struct {
	ErrorCode int            `json:"errorCode"`
	ErrorMsg  string         `json:"errorMsg"`
	Data      Pay1916Message `json:"data"`
}

type Pay1916Message struct {
	MerchantNo string `json:"merchantNo"`
	OrderNo    string `json:"orderNo"`
	POrderNo   string `json:"pOrderNo"`
	Amount     int    `json:"amount"`
	State      int    `json:"state"`
	Msg        string `json:"msg"`
}

type Pay1916BalanceRespond struct {
	ErrorCode int                 `json:"errorCode"`
	ErrorMsg  string              `json:"errorMsg"`
	Data      Pay1916QueryMessage `json:"data"`
}

type Pay1916QueryMessage struct {
	MerchantNo string `json:"merchantNo"`
	Balance    string `json:"balance"`
}

func InitConfig(url, appid, applicationId, md5key, withDrawMd5, notifyUrl, withdrawNotifyUrl string) {
	Config = &Pay1916Config{
		Appid:              appid,
		ApplicationId:      applicationId,
		Md5Key:             md5key,
		WithDrawMDd5Key:    withDrawMd5,
		PayOrderUrl:        fmt.Sprintf("%s%s", url, "/order/pay"),
		WithDrawUrl:        fmt.Sprintf("%s%s", url, "/order/draw"),
		PayNotifyUrl:       fmt.Sprintf("%s%s", notifyUrl, "/1916"),
		WithDrawNotifyUrl:  fmt.Sprintf("%s%s", withdrawNotifyUrl, "/1916"),
		InspectOrderUrl:    fmt.Sprintf("%s%s", url, "/order/pno"),
		InspectWithdrawUrl: fmt.Sprintf("%s%s", url, "/order/dno"),
		QueryUrl:           fmt.Sprintf("%s%s", url, "/mer/balance"),
	}
	service.Regist(service.PAY1916, Config)
}
