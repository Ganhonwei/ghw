package metagopay

import (
	"goserver/internal/global/pay/service"
)

func InitConfig(OrgId, MchId, AccountId, Md5Key, PayNotifyUrl, WithdrawNotifyUrl string) {
	Config = &MetagopayConfig{
		OrgId:             OrgId,
		MchId:             MchId,
		AccountId:         AccountId,
		Md5Key:            Md5Key,
		PayNotifyUrl:      PayNotifyUrl,
		WithdrawNotifyUrl: WithdrawNotifyUrl,
	}
	service.Regist(service.METAGOPAY, Config)
}

type MetagopayConfig struct {
	OrgId             string // 机构号
	MchId             string // 商户号
	AccountId         string // 商户子账号
	Md5Key            string // 支付MD5秘钥
	PayNotifyUrl      string
	WithdrawNotifyUrl string
}

type PayOrderRespond struct {
	OrgNo       string `json:"orgNo"`
	OrdStatus   string `json:"ordStatus"` // 00:未交易 01:成功 02:失败 03:被拒绝 04:处理中 05:已取消 06:未支付 07:已退款 08:退款中
	Code        string `json:"code"`      // 000000 表示请求该接口正常
	Msg         string `json:"msg"`
	CustId      string `json:"custId"`
	CustOrderNo string `json:"custOrderNo"`
	PrdOrdNo    string `json:"prdOrdNo"`
	ContentType string `json:"contentType"`
	BusContent  string `json:"busContent"`
	OrdDesc     string `json:"ordDesc"`
	Sign        string `json:"sign"`
}

// 核单响应
type PayInspectRespond struct {
	OrgNo              string `json:"orgNo"`
	OrdStatus          string `json:"ordStatus"` // 00:未交易 01:成功 02:失败 03:被拒绝 04:处理中 05:已取消 06:未支付 07:已退款 08:退款中
	Code               string `json:"code"`      // 000000 表示请求该接口正常
	Msg                string `json:"msg"`
	CustId             string `json:"custId"`
	CustOrderNo        string `json:"custOrderNo"`
	PrdOrdNo           string `json:"prdOrdNo"`
	ChannelReferenceNo string `json:"channelReferenceNo"`
	OrdAmt             string `json:"ordAmt"`
	PayAmt             string `json:"payAmt"`
	OrdTime            string `json:"ordTime"`
	OrdDesc            string `json:"ordDesc"`
	Utr                string `json:"utr"`
	Sign               string `json:"sign"`
}

// 下单响应
type WithdrawOrderRespond struct {
	OrgNo     string `json:"orgNo"`
	OrdStatus string `json:"ordStatus"` // 00:未处理 01:待结算 06:清算中 07:清算完成 08:清算失败 09:清算撤销
	Code      string `json:"code"`      // 000000 表示请求该接口正常
	Msg       string `json:"msg"`
	CustId    string `json:"custId"`
	CustOrdNo string `json:"custOrdNo"`
	CasOrdNo  string `json:"casOrdNo"`
	CasAmt    string `json:"casAmt"`
	CasTime   string `json:"casTime"`
	Sign      string `json:"sign"`
}

// 提现查单响应
type WithdrawInspectRespond struct {
	OrgNo      string `json:"orgNo"`
	OrdStatus  string `json:"ordStatus"` // 00:未处理 01:待结算 06:清算中 07:清算完成 08:清算失败 09:清算撤销
	Code       string `json:"code"`      // 000000 表示请求该接口正常
	Msg        string `json:"msg"`
	CustId     string `json:"custId"`
	CustOrdNo  string `json:"custOrdNo"`
	CasOrdNo   string `json:"casOrdNo"`
	CasAmt     string `json:"casAmt"`
	CasDesc    string `json:"casDesc"`
	CasType    string `json:"casType"`
	CasDate    string `json:"casDate"`
	Fee        string `json:"fee"`
	ServiceFee string `json:"serviceFee"`
	NetrecAmt  string `json:"netrecAmt"`
	OrdDesc    string `json:"ordDesc"`
	Utr        string `json:"utr"`
	Sign       string `json:"sign"`
}

type BalanceRespond struct {
	OrgNo    string `json:"orgNo"`
	Code     string `json:"code"` // 000000 表示请求该接口正常
	Msg      string `json:"msg"`
	CustId   string `json:"custId"`
	AcType   string `json:"acType"`
	AcBal    string `json:"acBal"`
	AcT0     string `json:"acT0"`
	AcT1     string `json:"acT1"`
	AcT0Froz string `json:"acT0Froz"`
	Sign     string `json:"sign"`
}

type PayRechargeCallback struct {
	Version     string `json:"version"`
	OrgNo       string `json:"orgNo"`
	CustId      string `json:"custId"`
	CustOrderNo string `json:"custOrderNo"`
	PrdOrdNo    string `json:"prdOrdNo"`
	OrdAmt      string `json:"ordAmt"`
	OrdTime     string `json:"ordTime"`
	PayAmt      string `json:"payAmt"`
	Utr         string `json:"utr"`
	OrdStatus   string `json:"ordStatus"` // 00:未交易 01:成功 02:失败 03:被拒绝 04:处理中 05:已取消 06:未支付 07:已退款 08:退款中
	Sign        string `json:"sign"`
}

type PayWithdrawCallback struct {
	OrgNo       string `json:"orgNo"`
	CustId      string `json:"custId"`
	CustOrderNo string `json:"custOrderNo"`
	PrdOrdNo    string `json:"prdOrdNo"`
	PayAmt      string `json:"payAmt"`
	OrdStatus   string `json:"ordStatus"` // 00:未处理 01:待结算 06:清算中 07:清算完成 08:清算失败 09:清算撤销
	CasDesc     string `json:"casDesc"`
	Utr         string `json:"utr"`
	Sign        string `json:"sign"`
}
