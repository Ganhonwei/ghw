package entity

import (
	"encoding/json"
	"fmt"
)

type IPay interface {
	Submit(order *PayOrder) ([]byte, error)
	SubmitResponse(body []byte) (string, error)
	WithdrawSubmit(order *WithdrawOrder) ([]byte, error)
	WithdrawSubmitResponse(body []byte) (bool, string)
}

// 支付渠道
type PayChannel struct {
	Id           string  `bson:"_id" json:"id"`                     // 渠道id
	Name         string  `bson:"name" json:"name"`                  // 渠道名称
	Pattern      int     `json:"pattern" bson:"pattern"`            // 模式:0：随机模式；1：权重模式
	Status       int     `bson:"status" json:"status"`              // 开关 0: 关; 1: 开;
	Wstatus      int     `json:"wstatus" bson:"wstatus"`            // 提现开关 0: 关; 1: 开;
	PayRate      float64 `json:"payRate" bson:"pay_rate"`           // 代收税率
	WithdrawRate float64 `json:"withdrawRate" bson:"withdraw_rate"` // 代付税率
	WithdrawFee  int64   `json:"withdrawFee" bson:"withdraw_fee"`   // 代付固定手续费
	Balance      int     `json:"balance" bson:"balance"`            // 账户余额

	PayOptions      []int32  `bson:"pay_options" json:"payOptions"`           // 支持的支付方式: 1.DirectLaunchTheApp 2.QRcode (3.FillInUTR)
	PayApps         []int32  `bson:"pay_apps" json:"payApps"`                 // 支持的支付app: 1.Paytm 2.PhonePe 3.Mobikwik 4.BHIM 5.GooglePay 6.Other UPI
	WithdrawBanks   []string `bson:"withdraw_banks" json:"withdrawBanks"`     // 支持的提现银行
	WithdrawWallets []string `bson:"withdraw_wallets" json:"withdrawWallets"` // 支持的提现钱包
	UtrRequired     bool     `bson:"utr_required" json:"utrRequired"`         // 是否需要UTR FillInUTR
	PayWeight       int32    `bson:"pay_weight" json:"payWeight"`             // 代收权重
	PayMin          int64    `bson:"pay_min" json:"payMin"`                   // 充值金额最小
	PayMax          int64    `bson:"pay_max" json:"payMax"`                   // 充值金额最大
	WithdrawWeight  int32    `bson:"withdraw_weight" json:"withdrawWeight"`   // 代付权重
	WithdrawMin     int64    `bson:"withdraw_min" json:"withdrawMin"`         // 提现金额最小
	WithdrawMax     int64    `bson:"withdraw_max" json:"withdrawMax"`         // 提现金额最大
}

// 代收订单
type PayOrder struct {
	OrderID        string `bson:"_id" json:"orderid"`                     // 订单ID
	BusinessID     string `json:"businessId" bson:"business_id"`          // 业务id
	MerOrderID     string `bson:"mer_order_id" json:"mer_orderid"`        // 商户订单ID
	Userid         string `bson:"userid" json:"userid" `                  // 用户id
	NickName       string `json:"nickName" bson:"nick_name"`              // 名字
	RealName       string `json:"realName" bson:"real_name"`              // 真实名字
	Email          string `json:"email" bson:"email"`                     // 邮箱地址
	Mobile         string `json:"mobile" bson:"mobile"`                   // 电话
	RegistIp       string `json:"registIp"`                               // 注册ip
	Amount         uint32 `json:"amount" bson:"amount"`                   // 金额
	OrderStatus    int32  `json:"orderStatus" bson:"order_status" `       // 订单状态 1:待支付;2:支付成功;3:支付失败
	Ctime          int64  `bson:"ctime" json:"ctime"`                     // 下单时间
	PayTime        int64  `json:"payTime" bson:"pay_time"`                // 支付完成时间
	OutTradeNo     string `json:"outTradeNo" bson:"out_trade_no"`         // 第三方订单号
	ChannelId      uint32 `json:"channelId" bson:"channel_id"`            // 支付渠道
	OutTradeStatus string `json:"outTradeStatus" bson:"out_trade_status"` // 支付方返回订单状态
	RefuseReason   string `json:"refuseReason" bson:"refuse_reason"`      // 支付方返回错误信息
	PlaceAnOrder   string `json:"placeAnOrder" bson:"place_an_order"`     // 下单参数JSON
	RequestParam   string `json:"requestParam" bson:"request_param"`      // 请求支付参数JSON
	RequestMsg     string `json:"requestMsg" bson:"request_msg"`          // 创建请求响应信息
	CallBackMsg    string `json:"callBackMsg" bson:"call_back_msg"`       // 代收下单回调信息
	PayWay         int32  `json:"payWay" bson:"pay_way"`                  // 支付方式0bank,1usdt
	PayOption      int32  `json:"payOption" bson:"pay_option"`            // 选择的支付方式: 1.DirectLaunchTheApp 2.QRcode 3.FillInUTR
	PayApp         int32  `json:"payApp" bson:"pay_app"`                  // 选择的支付应用: 印度：1.Paytm 2.PhonePe 3.Mobikwik 4.BHIM 5.GooglePay 6.Other UPI 孟加拉：101.Nagad 102.bkash 巴基斯坦：1001.EASYPAISA 1002.JAZZCASH
}

// 代付订单
type WithdrawOrder struct {
	OrderID        string   `bson:"_id" json:"orderid"`                     // 订单ID
	BusinessID     string   `json:"businessId" bson:"business_id"`          // 业务id
	MerOrderID     string   `bson:"mer_order_id" json:"mer_orderid"`        // 商户订单ID
	Userid         string   `bson:"userid" json:"userid" `                  // 用户id
	NickName       string   `json:"nickName" bson:"nick_name"`              // 名字
	Email          string   `json:"email" bson:"email"`                     // 邮箱地址
	Mobile         string   `json:"mobile" bson:"mobile"`                   // 电话
	RealName       string   `json:"realName" bson:"realName"`               // 真实名称
	Bank           string   `json:"blank" bson:"blank"`                     // 用户银行
	BankNumber     string   `json:"blankNumber" bson:"blank_number"`        // 银行卡号
	IFSC           string   `json:"ifsc" bson:"ifsc"`                       // ifsc
	Amount         uint32   `json:"amount" bson:"amount"`                   // 金额
	OrderStatus    int32    `json:"orderStatus" bson:"order_status" `       // 订单状态 1:待支付;2:支付成功;3:支付失败
	Ctime          int64    `bson:"ctime" json:"ctime"`                     // 下单时间
	PayTime        int64    `json:"payTime" bson:"pay_time"`                // 支付完成时间
	OutTradeNo     string   `json:"outTradeNo" bson:"out_trade_no"`         // 第三方订单号
	ChannelId      uint32   `json:"channelId" bson:"channel_id"`            // 支付渠道
	OutTradeStatus string   `json:"outTradeStatus" bson:"out_trade_status"` // 支付方返回订单状态
	RefuseReason   string   `json:"refuseReason" bson:"refuse_reason"`      // 支付方返回错误信息
	PlaceAnOrder   string   `json:"placeAnOrder" bson:"place_an_order"`     // 下单参数JSON
	RequestParam   string   `json:"requestParam" bson:"request_param"`      // 请求参数JSON
	RequestMsg     string   `json:"requestMsg" bson:"request_msg"`          // 创建请求响应信息
	CallBackMsg    string   `json:"callBackMsg" bson:"call_back_msg"`       // 代付下单回调信息
	InspectTimes   int      `json:"inspectTimes" bson:"inspect_times"`      // 审核次数
	UsedChannelId  []string `json:"usedChannelId" bson:"used_channel_id"`   // 使用过的通道
	PayWay         int32    `json:"payWay" bson:"pay_way"`                  // 提现方式0bank,1usdt
	Country        string   `json:"country" bson:"country"`                 // 国家
}

type PayPointRecord struct {
	ID         string `bson:"_id" json:"id"`                 // 记录ID
	OrderID    string `bson:"order_id" json:"orderid"`       // 订单ID
	BusinessID string `json:"businessId" bson:"business_id"` // 业务id
	Types      int    `json:"types" bson:"types"`            // 类型 1：代收；2：代付
	Module     string `json:"module" bson:"module"`          // 模块
	Msg        string `json:"msg" bson:"msg"`                // 错误消息
	Body       string `json:"body" bson:"body"`              // 数据源
	Ctime      int64  `bson:"ctime" json:"ctime"`            // 埋点时间
}

const (
	OrderAmount   = 1 // 订单金额错误
	SignError     = 2 // 下单签名错误
	OrderError    = 3 // 下单错误
	ChannelError  = 4 // 通道错误
	VerifySign    = 5 // 回调验签错误
	CallBackError = 6 // 回调错误
	CallBackOrder = 7 // 回调订单处理异常
	NotifyServer  = 8 // 通知服务器异常
)

func CovertWithdrawOrder(body []byte) *WithdrawOrder {
	order := &WithdrawOrder{}
	json.Unmarshal(body, order)
	return order
}

func ToQueryString(param map[string]string) string {
	var args string
	for k, v := range param {
		if args == "" {
			args += fmt.Sprintf("%s=%s", k, v)
		} else {
			args += fmt.Sprintf("&%s=%s", k, v)
		}
	}

	return args
}
