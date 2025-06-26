package ck

import (
	"fmt"
	"goserver/pkg/data"
	"time"

	"github.com/bytedance/sonic"
)

// 交易订单
type TradeRecord struct {
	Ver            int64     `gorm:"column:ver" json:"ver"`                            // 插入时间戳
	OrderID        string    `gorm:"column:order_id;primaryKey" json:"orderid"`        // 订单ID
	Userid         string    `gorm:"column:userid" json:"user_id"`                     // 用户id
	NickName       string    `json:"nickName" gorm:"column:nick_name"`                 // 昵称
	RealName       string    `json:"realName" gorm:"column:real_name"`                 // 真实名称
	Email          string    `json:"email" gorm:"column:email"`                        // 邮箱
	Mobile         string    `json:"mobile" gorm:"column:mobile"`                      // 手机号
	Amount         uint32    `json:"amount" gorm:"column:amount"`                      // 金额
	RealAmount     uint32    `json:"real_amount" gorm:"column:real_amount"`            // 实际支付金额
	Score          uint32    `json:"score" gorm:"column:score"`                        // 金币数
	OtherPresent   string    `json:"otherPresent" gorm:"column:other_present"`         // 额外赠送
	GiveCash       int64     `json:"giveCash" gorm:"column:give_cash"`                 // 赠送现金
	GiveWithdrawal int64     `json:"giveWithdrawal" gorm:"column:give_withdrawal"`     // 赠送提现金
	OrderStatus    int32     `json:"orderStatus" gorm:"column:order_status" `          // 订单状态
	Ctime          time.Time `gorm:"column:ctime" json:"ctime"`                        // 订单时间
	Utime          time.Time `gorm:"column:utime" json:"utime"`                        // 订单更新时间
	OrderAddress   string    `json:"orderAddress" gorm:"column:order_address"`         // 下单地址IP
	PayTime        time.Time `json:"payTime" gorm:"column:pay_time"`                   // 付款时间
	PayAddress     string    `json:"payAddress" gorm:"column:pay_address"`             // 付款地址IP
	OutTradeNo     string    `json:"outTradeNo" gorm:"column:out_trade_no"`            // 第三方订单号
	PackageId      string    `json:"packageId" gorm:"column:package_id"`               // 包名
	ChannelId      uint32    `json:"channelId" gorm:"column:channel_id"`               // 支付渠道
	ShopId         string    `json:"shopId" gorm:"column:shop_id"`                     // 商店ID
	ShopType       int32     `json:"shopType" gorm:"column:shop_type"`                 // 商品类型
	ShopName       string    `json:"ShopName" gorm:"column:shop_name"`                 // 商品名称
	RefuseReason   string    `json:"refuseReason" gorm:"column:refuse_reason"`         // 支付方返回错误信息
	OutTradeStatus string    `json:"outTradeStatus" gorm:"column:out_trade_status"`    // 支付方返回订单状态
	ReportId       string    `json:"reportId" gorm:"column:report_id"`                 // 埋点id
	Repair         bool      `json:"repair" gorm:"column:repair"`                      // 是否补单
	Gtype          int32     `json:"gtype" gorm:"column:gtype"`                        // 游戏类型(大厅充值就是上一局游戏)
	UserType       int       `json:"userType" gorm:"column:user_type"`                 // 玩家类型 0:A类  1:B类
	PayWay         int32     `json:"payWay" gorm:"column:pay_way"`                     // 支付方式
	PayOption      int32     `json:"payOption" gorm:"column:pay_option"`               // 选择的支付方式: 1.DirectLaunchTheApp 2.QRcode 3.FillInUTR
	PayApp         int32     `json:"payApp" gorm:"column:pay_app"`                     // 选择的支付应用: 1.Paytm 2.PhonePe 3.Mobikwik 4.BHIM 5.GooglePay 6.Other UPI
	InGtype        int32     `json:"inGtype" gorm:"column:in_gtype"`                   // 充值时所在游戏类型
	UserUtrs       []string  `json:"-" gorm:"column:user_utrs;type:Array(String)"`     // 用户回填utr
	UserUtrTimes   []int64   `json:"-" gorm:"column:user_utr_times;type:Array(Int64)"` // 用户回填utr时间
	LastUtr        string    `json:"lastUtr" gorm:"column:last_utr"`                   // 用户最后回填utr
	LastUtrTime    int64     `json:"lastUtrTime" gorm:"column:last_utr_time"`          // 用户最后回填utr时间戳秒
	FirstPay       bool      `json:"firstPay" gorm:"column:first_pay"`                 // 是否首次支付
	// ErrorMsg       string    `json:"errorMsg" gorm:"column:error_msg"`                 // 标记信息
	// RepairAdmin    string    `json:"repairAdmin" gorm:"column:repair_admin"`           // 手动补单用户
}

func (*TradeRecord) TableName() string {
	return "col_trade_record"
}

func (*TradeRecord) New() CkEntity {
	return new(TradeRecord)
}

func (c *TradeRecord) ParseData(d interface{}) (rs []interface{}, err error) {
	// check type
	from, ok := d.(*data.TradeRecord)
	if !ok {
		return nil, fmt.Errorf("parse type %T to %T error", d, c)
	}

	bytes, err := sonic.Marshal(from)
	if err != nil {
		return
	}
	if err = sonic.Unmarshal(bytes, c); err != nil {
		return
	}
	c.Ver = time.Now().UnixMilli()
	for _, utr := range from.UserUtrs {
		c.UserUtrs = append(c.UserUtrs, utr.UTR)
		c.UserUtrTimes = append(c.UserUtrTimes, utr.Time)
	}
	rs = append(rs, c)
	return
}
