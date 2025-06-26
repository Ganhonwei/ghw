package ck

import (
	"fmt"
	"goserver/pkg/data"
	"time"

	"github.com/bytedance/sonic"
)

// 提现订单
type WithdrawRecord struct {
	Ver                  int64     `gorm:"column:ver" json:"ver"`                                      // 插入时间戳
	OrderID              string    `gorm:"column:order_id" json:"orderid"`                             // 订单ID
	Userid               string    `gorm:"column:userid" json:"user_id" `                              // 用户id
	NickName             string    `json:"nickName" gorm:"column:nick_name"`                           // 昵称
	RealName             string    `json:"realName" gorm:"column:real_name"`                           // 真实名称
	Email                string    `json:"email" gorm:"column:email"`                                  // 邮箱
	Mobile               string    `json:"mobile" gorm:"column:mobile"`                                // 手机号
	Amount               uint32    `json:"amount" gorm:"column:amount"`                                // 实际提现金额
	Score                uint32    `json:"score" gorm:"column:score"`                                  // 移除彩金数
	Commission           uint32    `json:"commission" gorm:"column:commission"`                        // 手续费
	PayWay               int32     `json:"payWay" gorm:"column:pay_way"`                               // 提现方式0bank,1usdt
	Blank                string    `json:"blank" gorm:"column:blank"`                                  // 用户银行
	BlankNumber          string    `json:"blankNumber" gorm:"column:blank_number"`                     // 银行卡号
	IFSC                 string    `json:"ifsc" gorm:"column:ifsc"`                                    // ifsc
	OutChannel           string    `json:"outChannel" gorm:"column:out_channel"`                       // 代付渠道
	OrderStatus          int32     `json:"orderStatus" gorm:"column:order_status" `                    // 订单状态
	Ctime                time.Time `gorm:"column:ctime" json:"ctime"`                                  // 订单时间
	OrderAddress         string    `json:"orderAddress" gorm:"column:order_address"`                   // 下单地址IP
	WithdrawTime         time.Time `json:"payTime" gorm:"column:pay_time"`                             // 代付完成时间
	OutTradeNo           string    `json:"outTradeNo" gorm:"column:out_trade_no"`                      // 第三方订单号
	PackageId            string    `json:"packageId" gorm:"column:package_id"`                         // 包名
	ShopId               int32     `json:"shopId" gorm:"column:shop_id"`                               // 商品ID
	ShopName             string    `json:"ShopName" gorm:"column:shop_name"`                           // 商品名称
	ExamineWay           int32     `json:"examineWay" gorm:"column:examine_way"`                       // 审核方式
	ExamineAccount       string    `json:"ExamineAccount" gorm:"column:examine_account"`               // 审核人
	RefuseReason         string    `json:"refuseReason" gorm:"column:refuse_reason"`                   // 支付方返回错误信息
	OutTradeStatus       string    `json:"outTradeStatus" gorm:"column:out_trade_status"`              // 支付方返回订单状态
	Utime                time.Time `gorm:"column:utime" json:"utime"`                                  // 订单更新时间
	ErrorMsg             string    `json:"errorMsg" gorm:"column:error_msg"`                           // 自动审核错误信息
	ETime                time.Time `json:"etime" gorm:"column:e_time"`                                 // 操作时间
	UserType             int       `json:"userType" gorm:"column:user_type"`                           // 玩家类型 0:A类  1:B类
	Diamond              int64     `json:"diamond" gorm:"diamond"`                                     // 玩家提单时携带金额
	RepeatTimes          int       `json:"repeatTimes" gorm:"column:repeat_times"`                     // 重复提现次数
	OvertimePayTacks     int64     `json:"overtimePayTacks" gorm:"column:overtime_pay_tacks"`          // 提现超时赔付已领取金额
	OvertimePayTackTimes int32     `json:"overtimePayTackTimes" gorm:"column:overtime_pay_tack_times"` // 提现超时赔付已领取次数
	KeepWaitTime         int64     `json:"keepWaitTime" gorm:"column:keep_wait_time"`                  // 超过可退款时间后,玩家点击KeepWait时间戳(秒)
	RefundCompensate     int64     `json:"refundCompensate" gorm:"column:refund_compensate"`           // 提现退款补偿
	RefundCompensateType int32     `json:"refundCompensateType" gorm:"column:refund_compensate_type"`  // 提现退款补偿类型:1代表bonus,2代表cash,3代表withdrawable
	// TransferDetail []*TransferOrderDetail `json:"transferDetail" gorm:"column:transfer_detail"`  // 转单详情
}

func (*WithdrawRecord) TableName() string {
	return "col_withdraw_record"
}

func (*WithdrawRecord) New() CkEntity {
	return new(WithdrawRecord)
}

func (c *WithdrawRecord) ParseData(d interface{}) (rs []interface{}, err error) {
	// check type
	from, ok := d.(*data.WithdrawRecord)
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
	rs = append(rs, c)

	// 转单详情子表
	for _, trans := range from.TransferDetail {
		trans.OrderID = from.OrderID
		trans.Userid = from.Userid
		item := &WithdrawRecordTransferOrderDetail{}
		r, e := item.ParseData(trans)
		if e != nil {
			err = e
			return
		}
		rs = append(rs, r...)
	}
	return
}

// 提现订单转单详情
type WithdrawRecordTransferOrderDetail struct {
	Ver       int64  `gorm:"column:ver" json:"ver"`              // 插入时间戳
	OrderID   string `gorm:"column:order_id" json:"orderid"`     // 订单ID
	Userid    string `gorm:"column:userid" json:"user_id" `      // 用户id
	ChannelId uint32 `json:"channelId" gorm:"column:channel_id"` // 渠道id
	TType     int    `json:"tType" gorm:"column:t_type"`         // 转单类型
	Status    int    `json:"status" gorm:"column:status"`        // 状态
	UserName  string `json:"userName" gorm:"column:user_name"`   // 用户名
	Ctime     int64  `json:"ctime" gorm:"column:ctime"`          // 创建时间
	Utime     int64  `json:"utime" gorm:"column:utime"`          // 更新时间
	Reason    int    `json:"reason" gorm:"reason"`               // 转单原因 0:正常 1:超时 2.超时急需 3.失败
}

func (*WithdrawRecordTransferOrderDetail) TableName() string {
	return "col_withdraw_record_transfer_order_detail"
}

func (*WithdrawRecordTransferOrderDetail) New() CkEntity {
	return new(WithdrawRecordTransferOrderDetail)
}

func (c *WithdrawRecordTransferOrderDetail) ParseData(d interface{}) (rs []interface{}, err error) {
	// check type
	from, ok := d.(*data.TransferOrderDetail)
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
	rs = append(rs, c)
	return
}
