// 充值交易记录
package data

import (
	"encoding/json"
	"net/url"
	"time"
	"unicode"

	"goserver/pkg/data/mq"
	"goserver/pkg/utils"

	"github.com/globalsign/mgo/bson"
)

type PayOrder struct {
	OrderID        string    `bson:"_id" json:"orderid"`                     // 订单ID
	Userid         string    `bson:"userid" json:"user_id" `                 // 用户id
	NickName       string    `json:"nickName" bson:"nick_name"`              // 昵称
	RealName       string    `json:"realName" bson:"real_name"`              // 真实名称
	Email          string    `json:"email" bson:"email"`                     // 邮箱
	Mobile         string    `json:"mobile" bson:"mobile"`                   // 手机号
	RegistIp       string    `json:"registIp"`                               // 注册ip
	Amount         uint32    `json:"amount" bson:"amount"`                   // 金额
	Score          uint32    `json:"score" bson:"score"`                     // 彩金数
	OtherPresent   string    `json:"otherPresent" bson:"other_present"`      // 额外赠送bonus
	GiveCash       int64     `json:"giveCash" bson:"give_cash"`              // 赠送现金
	GiveWithdrawal int64     `json:"giveWithdrawal" bson:"give_withdrawal"`  // 赠送提现金
	OrderStatus    int32     `json:"orderStatus" bson:"order_status" `       // 订单状态
	Ctime          time.Time `bson:"ctime" json:"ctime"`                     // 订单时间
	OrderAddress   string    `json:"orderAddress" bson:"order_address"`      // 下单地址IP
	PayTime        time.Time `json:"payTime" bson:"pay_time"`                // 付款时间
	PayAddress     string    `json:"payAddress" bson:"pay_address"`          // 付款地址IP
	OutTradeNo     string    `json:"outTradeNo" bson:"out_trade_no"`         // 第三方订单号
	PackageId      string    `json:"packageId" bson:"package_id"`            // 包名
	ChannelId      uint32    `json:"channelId" bson:"channel_id"`            // 支付渠道
	ShopId         string    `json:"shopId" bson:"shop_id"`                  // 商店ID
	ShopType       int32     `json:"shopType" bson:"shop_type"`              // 商品类型
	ShopName       string    `json:"ShopName" bson:"shop_name"`              // 商品名称
	RefuseReason   string    `json:"refuseReason" bson:"refuse_reason"`      // 支付方返回错误信息
	OutTradeStatus string    `json:"outTradeStatus" bson:"out_trade_status"` // 支付方返回订单状态
	ReportId       string    `json:"reportId" bson:"report_id"`              // 埋点id
	Gtype          int32     `json:"gtype" bson:"gtype"`                     // 游戏类型(大厅充值就是上一局游戏)
	UserType       int       `json:"userType" bson:"user_type"`              // 玩家类型 0:A类  1:B类
	PayWay         int32     `json:"payWay" bson:"pay_way"`                  // 提现方式0bank,1usdt
	PayOption      int32     `json:"payOption" bson:"pay_option"`            // 选择的支付方式: 1.DirectLaunchTheApp 2.QRcode 3.FillInUTR
	PayApp         int32     `json:"payApp" bson:"pay_app"`                  // 选择的支付应用: 1.Paytm 2.PhonePe 3.Mobikwik 4.BHIM 5.GooglePay 6.Other UPI
	InGtype        int32     `json:"inGtype" bson:"in_gtype"`                // 充值时所在游戏类型
}

type WithdrawOrder struct {
	OrderID        string    `bson:"_id" json:"orderid"`                     // 订单ID
	Userid         string    `bson:"userid" json:"user_id" `                 // 用户id
	NickName       string    `json:"nickName" bson:"nick_name"`              // 昵称
	RealName       string    `json:"realName" bson:"real_name"`              // 真实名称
	Email          string    `json:"email" bson:"email"`                     // 邮箱
	Mobile         string    `json:"mobile" bson:"mobile"`                   // 手机号
	Amount         uint32    `json:"amount" bson:"amount"`                   // 实际提现金额
	Score          uint32    `json:"score" bson:"score"`                     // 移除金币数
	Commission     uint32    `json:"commission" bson:"commission"`           // 手续费
	PayWay         int32     `json:"payWay" bson:"pay_way"`                  // 提现方式0bank,1usdt
	Bank           string    `json:"blank" bson:"blank"`                     // 用户银行
	BankNumber     string    `json:"blankNumber" bson:"blank_number"`        // 银行卡号
	IFSC           string    `json:"ifsc" bson:"ifsc"`                       // ifsc
	OutChannel     string    `json:"outChannel" bson:"out_channel"`          // 代付渠道
	OrderStatus    int32     `json:"orderStatus" bson:"order_status" `       // 订单状态
	Ctime          time.Time `bson:"ctime" json:"ctime"`                     // 订单时间
	OrderAddress   string    `json:"orderAddress" bson:"order_address"`      // 下单地址IP
	WithdrawTime   time.Time `json:"payTime" bson:"pay_time"`                // 代付完成时间
	OutTradeNo     string    `json:"outTradeNo" bson:"out_trade_no"`         // 第三方订单号
	PackageId      string    `json:"packageId" bson:"package_id"`            // 包名
	ShopId         int32     `json:"shopId" bson:"shop_id"`                  // 商品ID
	ShopName       string    `json:"ShopName" bson:"shop_name"`              // 商品名称
	ExamineWay     int32     `json:"examineWay" bson:"examine_way"`          // 审核方式
	ExamineAccount string    `json:"ExamineAccount" bson:"examine_account"`  // 审核人
	RefuseReason   string    `json:"refuseReason" bson:"refuse_reason"`      // 支付方返回错误信息
	OutTradeStatus string    `json:"outTradeStatus" bson:"out_trade_status"` // 支付方返回订单状态
	ErrorMsg       string    `json:"errorMsg" bson:"error_msg"`              // 自动审核错误信息
	UserType       int       `json:"userType" bson:"user_type"`              // 玩家类型 0:A类  1:B类
	Diamond        int64     `json:"diamond" bson:"diamond"`                 // 玩家提单时携带金额
}

func CovertWithdrawOrder(body []byte) *WithdrawOrder {
	order := &WithdrawOrder{}
	json.Unmarshal(body, order)
	return order
}

// ToUpper golang make the caracter in a string uppercase
func ToUpper(str string) string {
	var s string
	for _, v := range str {
		s += string(unicode.ToUpper(v))
	}
	return s
}

func ToLower(str string) string {
	var s string
	for _, v := range str {
		s += string(unicode.ToLower(v))
	}
	return s
}

// URLENCODE编码转换
func ToUrlEncode(str string) string {
	encodedURL := url.QueryEscape(str)
	return encodedURL
}

// URLDECODE解码转换
func ToUrlDecode(str string) string {
	decodedURL, err := url.QueryUnescape(str)
	if err != nil {
		return str // 如果解码失败，返回原字符串
	}
	return decodedURL
}

const (
	Tradeing       = 1 //交易中(下单状态)
	TradeFail      = 2 //交易失败
	TradeGoods     = 3 //发货失败
	TradeSuccess   = 4 //交易成功
	TradeOrderFail = 5 //下单失败
	// 充值类型
	SHOP            = 1
	RECHARGE_GIFT   = 2  //入门礼包
	METAL_CARD      = 3  //金银铜卡
	FIRST_RECHARGE  = 4  //首充
	GAME_RECHARGE   = 5  //局内充值
	LIMITED_GIFT    = 6  //破冰高送
	BREAKING_GITF   = 7  //破产礼包
	SIGN_RECHARGE   = 8  //签到充值
	CUSTOM_RECHARGE = 9  //自定义充值
	Withdraw_GIFT   = 10 //破冰解提
	COUPON          = 11 //优惠券
	WEB_TEST_SHOP   = 99 //后台测试单
)

// 支付渠道
type PayChannel struct {
	Id      string `bson:"_id" json:"id"`          // 渠道id
	Name    string `bson:"name" json:"name"`       // 渠道名称
	Pattern int    `json:"pattern" bson:"pattern"` // 模式:0：随机模式；1：权重模式
	Status  int    `bson:"status" json:"status"`   // 开关 0: 关; 1: 开;
	Wstatus int    `json:"wstatus" bson:"wstatus"` // 提现开关 0: 关; 1: 开;
}

// 交易记录
type TradeRecord struct {
	OrderID        string           `bson:"_id" json:"orderid"`                     // 订单ID
	Userid         string           `bson:"userid" json:"user_id" `                 // 用户id
	NickName       string           `json:"nickName" bson:"nick_name"`              // 昵称
	RealName       string           `json:"realName" bson:"real_name"`              // 真实名称
	Email          string           `json:"email" bson:"email"`                     // 邮箱
	Mobile         string           `json:"mobile" bson:"mobile"`                   // 手机号
	Amount         uint32           `json:"amount" bson:"amount"`                   // 金额
	RealAmount     uint32           `json:"real_amount" bson:"real_amount"`         // 实际支付金额
	Score          uint32           `json:"score" bson:"score"`                     // 金币数
	OtherPresent   string           `json:"otherPresent" bson:"other_present"`      // 额外赠送bonus
	GiveCash       int64            `json:"giveCash" bson:"give_cash"`              // 赠送现金
	GiveWithdrawal int64            `json:"giveWithdrawal" bson:"give_withdrawal"`  // 赠送提现金
	OrderStatus    int32            `json:"orderStatus" bson:"order_status" `       // 订单状态
	Ctime          time.Time        `bson:"ctime" json:"ctime"`                     // 订单时间
	Utime          time.Time        `bson:"utime" json:"utime"`                     // 订单更新时间
	OrderAddress   string           `json:"orderAddress" bson:"order_address"`      // 下单地址IP
	PayTime        time.Time        `json:"payTime" bson:"pay_time"`                // 付款时间
	PayAddress     string           `json:"payAddress" bson:"pay_address"`          // 付款地址IP
	OutTradeNo     string           `json:"outTradeNo" bson:"out_trade_no"`         // 第三方订单号
	PackageId      string           `json:"packageId" bson:"package_id"`            // 包名
	ChannelId      uint32           `json:"channelId" bson:"channel_id"`            // 支付渠道
	ShopId         string           `json:"shopId" bson:"shop_id"`                  // 商店ID
	ShopType       int32            `json:"shopType" bson:"shop_type"`              // 商品类型
	ShopName       string           `json:"ShopName" bson:"shop_name"`              // 商品名称
	RefuseReason   string           `json:"refuseReason" bson:"refuse_reason"`      // 支付方返回错误信息
	OutTradeStatus string           `json:"outTradeStatus" bson:"out_trade_status"` // 支付方返回订单状态
	ReportId       string           `json:"reportId" bson:"report_id"`              // 埋点id
	Repair         bool             `json:"repair" bson:"repair"`                   // 是否补单
	Gtype          int32            `json:"gtype" bson:"gtype"`                     // 游戏类型(大厅充值就是上一局游戏)
	UserType       int              `json:"userType" bson:"user_type"`              // 玩家类型 0:A类  1:B类
	PayWay         int32            `json:"payWay" bson:"pay_way"`                  // 支付方式
	PayOption      int32            `json:"payOption" bson:"pay_option"`            // 选择的支付方式: 1.DirectLaunchTheApp 2.QRcode 3.FillInUTR
	PayApp         int32            `json:"payApp" bson:"pay_app"`                  // 选择的支付应用: 1.Paytm 2.PhonePe 3.Mobikwik 4.BHIM 5.GooglePay 6.Other UPI
	InGtype        int32            `json:"inGtype" bson:"in_gtype"`                // 充值时所在游戏类型
	UserUtrs       []TradeRecordUTR `json:"userUtrs" bson:"user_utrs"`              // 用户回填utr
	LastUtr        string           `json:"lastUtr" bson:"last_utr"`                // 用户最后回填utr
	LastUtrTime    int64            `json:"lastUtrTime" bson:"last_utr_time"`       // 用户最后回填utr时间戳秒
	FirstPay       bool             `json:"firstPay" bson:"first_pay"`              // 是否首次支付
	ErrorMsg       string           `json:"errorMsg" bson:"error_msg"`              // 标记信息
	RepairAdmin    string           `json:"repairAdmin" bson:"repair_admin"`        // 手动补单用户
}

type TradeRecordUTR struct {
	UTR  string `json:"utr" bson:"utr"`
	Time int64  `json:"time" bson:"time"`
}

func GetPayChannelList() []PayChannel {
	var list []PayChannel
	ListByQ(PayChannels, nil, &list)
	return list
}

// 生成订单id,(时间截+角色id)
func GenCporderid(userid string) string {
	return utils.Base62encode(uint64(utils.TimestampNano())) + userid
}

func GenOrderid() string {
	return bson.NewObjectId().Hex()
}

// 交易结果记录
func (this *TradeRecord) Get() {
	Get(TradeRecords, this.OrderID, this)
}

func (this *TradeRecord) Has() bool {
	return Has(TradeRecords, bson.M{"_id": this.OrderID})
}

func (this *TradeRecord) Update() bool {
	defer DataProducers.PublishDatas(mq.TopicSyncTradeRecord, []any{this})

	this.Utime = bson.Now()
	return Update(TradeRecords, bson.M{"_id": this.OrderID}, this)
}

func (this *TradeRecord) Save() bool {
	defer DataProducers.PublishDatas(mq.TopicSyncTradeRecord, []any{this})

	this.Ctime = bson.Now()
	return Insert(TradeRecords, this)
}

func (this *TradeRecord) Upsert() bool {
	defer DataProducers.PublishDatas(mq.TopicSyncTradeRecord, []any{this})

	this.Utime = bson.Now()
	return Upsert(TradeRecords, bson.M{"_id": this.OrderID}, this)
}

/*
func (this *TradeRecord) Delete() bool {
	return Delete(TradeRecords, bson.M{"_id": this.Id})
}
*/

// 获取某玩家的所有离线订单,用于上线补单
func GetTradeOff(userid string) []*TradeRecord {
	var list []*TradeRecord
	ListByQ(TradeRecords, bson.M{"userid": userid, "result": TradeFail}, &list)
	return list
}
