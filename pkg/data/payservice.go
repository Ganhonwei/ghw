// 支付服务结构体
package data

// 支付请求结构体
type PayRequest struct {
	OrderID    string `json:"orderid" bson:"orderid"`        // 订单ID
	BusinessID string `json:"businessId" bson:"business_id"` // 业务id
	Userid     string `bson:"userid" json:"userid" `         // 用户id
	Amount     uint32 `json:"amount" bson:"amount"`          // 金额
	RegistIp   string `json:"registIp" bson:"regist_ip"`     // 注册ip
	Channel    uint32 `json:"channel" bson:"channel"`        // 指定渠道
	PayWay     int32  `json:"payWay" bson:"pay_way"`         // 支付方式0bank,1usdt
	PayOption  int32  `json:"payOption" bson:"pay_option"`   // 选择的支付方式: 1.DirectLaunchTheApp 2.QRcode 3.FillInUTR
	PayApp     int32  `json:"payApp" bson:"pay_app"`         // 选择的支付应用: 1.Paytm 2.PhonePe 3.Mobikwik 4.BHIM 5.GooglePay 6.Other UPI
}

type WebPayRequest struct {
	OrderID    string `json:"orderid" bson:"orderid"`        // 订单ID
	BusinessID string `json:"businessId" bson:"business_id"` // 业务id
	Userid     string `bson:"userid" json:"userid" `         // 用户id
	Amount     uint32 `json:"amount" bson:"amount"`          // 金额
	RegistIp   string `json:"registIp" bson:"regist_ip"`     // 注册ip
	ChannelId  uint32 `json:"ChannelId" bson:"channel_id"`   // 支付渠道
}

// 支付请求响应结构体
type PayResponse struct {
	Code       int    `json:"code"`        // 请求结果码，成功为200,
	OrderID    string `json:"orderid"`     // 订单ID
	MerOrderID string `json:"mer_orderid"` // 商户订单ID
	ChannelID  uint32 `json:"channelId"`   // 支付渠道ID
	PaymentUrl string `json:"paymentUrl"`  // 支付地址
	Msg        string `json:"msg"`         // 支付信息
}

// 通知服务器支付状态结构体
type PayNotify struct {
	OrderID    string `json:"orderid"`                       // 订单ID(第三方支付订单ID)
	BusinessID string `json:"businessId" bson:"business_id"` // 业务id
	MerOrderID string `json:"mer_orderid"`                   // 商户订单ID
	// 提现:1.提单成功 2:提现成功 3:提单失败 4:提现失败 5:资料不对 6:通道不稳定 7:钱不够
	// 充值:1:待支付;2:支付成功;3:支付失败
	Status    int    `json:"status"`    // 状态
	Amount    int    `json:"amount"`    // 金额(分)
	Timestamp int64  `json:"timestamp"` // 时间戳，精确到毫秒
	Repair    bool   `json:"repair"`    // 是否补单
	ChannelId uint32 `json:"channelId"` // 渠道
	Reason    int    `json:"reason"`    // 转单原因
}

// 提现请求结构体
type WithdrawRequest struct {
	OrderID    string `json:"orderid" bson:"orderid"`          // 订单ID
	BusinessID string `json:"businessId" bson:"business_id"`   // 业务id
	Userid     string `bson:"userid" json:"userid" `           // 用户id
	RealName   string `json:"realName" bson:"realname"`        // 真实名称
	Bank       string `json:"blank" bson:"blank"`              // 用户银行
	BankNumber string `json:"blankNumber" bson:"blank_number"` // 银行卡号
	IFSC       string `json:"ifsc" bson:"ifsc"`                // ifsc
	Amount     uint32 `json:"amount" bson:"amount"`            // 金额
	Transfer   bool   `json:"transfer" bson:"transfer"`        // 是否转单
	PayWay     int32  `json:"payWay" bson:"pay_way"`           // 提现方式0bank,1usdt
	Country    string `json:"country" bson:"country"`          // 国家
}

// 提现请求响应结构体
type WithdrawResponse struct {
	Code       int    `json:"code"`        // 请求结果码，成功为200, 1：不是post请求；2：解析请求参数失败；3：支付服务创建订单失败；100：其他错误；200：成功
	OrderID    string `json:"orderid"`     // 订单ID
	MerOrderID string `json:"mer_orderid"` // 商户订单ID
	ChannelID  uint32 `json:"channelId"`   // 支付渠道ID
	Msg        string `json:"msg"`         // 支付信息
	Status     int32  `json:"status"`      // 状态
}

// 重复提现
type RepeatWithdrawReq struct {
	OrderID   string `json:"orderid"`   // 订单ID
	ChannelID uint32 `json:"channelId"` // 支付渠道ID
	Status    int    `json:"status"`    // 状态
	Msg       string `json:"msg"`       // 支付信息
}

type PayChannelShopRequest struct {
}

type PayChannelShopResponse struct {
	Code     int              `json:"code"` // 请求结果码, 成功为200
	Msg      string           `json:"msg"`  // 请求结果消息
	Channels []PayChannelShop `json:"channels"`
}

// 支付渠道信息
type PayChannelShop struct {
	Id              string
	PayOptions      []int32  `json:"payOptions"`      // 支持的支付方式: 1.DirectLaunchTheApp 2.QRcode (3.FillInUTR)
	PayApps         []int32  `json:"payApps"`         // 支持的支付app: 1.Paytm 2.PhonePe 3.Mobikwik 4.BHIM 5.GooglePay 6.Other UPI
	WithdrawBanks   []string `json:"withdrawBanks"`   // 支持的提现银行
	WithdrawWallets []string `json:"withdrawWallets"` // 支持的提现钱包
	UtrRequired     bool     `json:"utrRequired"`     // 是否需要UTR FillInUTR
	Payable         bool     `json:"payable"`         // 支付可用
	PayMin          int64    `json:"payMin"`          // 充值金额最小
	PayMax          int64    `json:"payMax"`          // 充值金额最大
	Withdrawable    bool     `json:"withdrawable"`    // 提现可用
	WithdrawMin     int64    `json:"withdrawMin"`     // 提现金额最小
	WithdrawMax     int64    `json:"withdrawMax"`     // 提现金额最大
}
