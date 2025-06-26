package data

import (
	"time"

	"goserver/pkg/data/mq"

	"github.com/globalsign/mgo/bson"
)

const (
	AutoExamineWay = 1 // 自动审核
	HandExamineWay = 2 // 手动审核
	// 提现
	Withdrawing             = 1  //待审核
	WaitDelivery            = 8  //待派单
	OrderSuccess            = 6  //派单成功
	OrderFail               = 7  //未知原因派单失败
	DeliveryFailTransfer    = 9  //派单失败需转单
	DeliveryFailRefund      = 10 //派单失败需退款
	WithdrawUrgentTransfer  = 11 //提现超时急需转单
	WithdrawTimeoutTransfer = 15 //提现超时需转单
	WithdrawFailTransfer    = 13 //提现失败需转单
	WithdrawFailRefund      = 14 //提现失败需退款
	WithdrawFail            = 5  //未知原因提现失败
	WithdrawSuccess         = 2  //提现成功
	WithdrawBack            = 3  //回退
	WithdrawFreeze          = 4  //冻结

	// WithdrawInspect = 8 //核单中
	// WithdrawBackBankErr = 9  //银行信息错误，回退
	// WithdrawNoMoneyErr  = 10 //渠道没钱，回退
	// WithdrawChannelErr  = 11 //通道错误，回退
	WithdrawRefunded = 12 //玩家主动申请退款

	//充值
	PaySuccess = 2 //充值成功
	PayFail    = 3 //充值失败

	// 转单类型
	NoTransfer   = 1 // 非转单(原始渠道)
	AutoTransfer = 2 // 自动转单
	HandTransfer = 3 // 手动转单

	// 转单原因
	NormalTransfer        = 0 // 正常
	TimeoutTransfer       = 1 // 超时
	TimeoutUrgentTransfer = 2 // 超时急需
	FailedTransfer        = 3 // 失败

	TransferCashKey = "withdraw_transfer_cash"
)

// 配置
type Withdraw struct {
	Id            int32  `bson:"_id" json:"id"`                        //提现Id
	Name          string `bson:"name" json:"name"`                     //商品名称
	Limit         int32  `bson:"limit" json:"limit"`                   //提现次数限制
	RechargeLimit int64  `bson:"recharge_limit" json:"recharge_limit"` //充值金额限制
	FlowingLimit  uint32 `bson:"flowing_limit" json:"flowing_limit"`   //流水限制
	CostDiamond   uint32 `bson:"cost_diamond" json:"cost_diamond"`     //消耗彩金
	GetCash       uint32 `bson:"get_cash" json:"get_cash"`             //提现金额(分)
	Commission    uint32 `bson:"commission" json:"commission"`         //手续费(分)
	Index         int32  `bson:"index" json:"index"`                   //展示权重
}

// 提现订单
type WithdrawRecord struct {
	OrderID        string                 `bson:"_id" json:"orderid"`                     // 订单ID
	Userid         string                 `bson:"userid" json:"user_id" `                 // 用户id
	NickName       string                 `json:"nickName" bson:"nick_name"`              // 昵称
	RealName       string                 `json:"realName" bson:"real_name"`              // 真实名称
	Email          string                 `json:"email" bson:"email"`                     // 邮箱
	Mobile         string                 `json:"mobile" bson:"mobile"`                   // 手机号
	Amount         uint32                 `json:"amount" bson:"amount"`                   // 实际提现金额
	Score          uint32                 `json:"score" bson:"score"`                     // 移除彩金数
	Commission     uint32                 `json:"commission" bson:"commission"`           // 手续费
	PayWay         int32                  `json:"payWay" bson:"pay_way"`                  // 提现方式0bank,1usdt
	Blank          string                 `json:"blank" bson:"blank"`                     // 用户银行
	BlankNumber    string                 `json:"blankNumber" bson:"blank_number"`        // 银行卡号
	IFSC           string                 `json:"ifsc" bson:"ifsc"`                       // ifsc
	OutChannel     string                 `json:"outChannel" bson:"out_channel"`          // 代付渠道
	OrderStatus    int32                  `json:"orderStatus" bson:"order_status" `       // 订单状态
	Ctime          time.Time              `bson:"ctime" json:"ctime"`                     // 订单时间
	OrderAddress   string                 `json:"orderAddress" bson:"order_address"`      // 下单地址IP
	WithdrawTime   time.Time              `json:"payTime" bson:"pay_time"`                // 代付完成时间
	OutTradeNo     string                 `json:"outTradeNo" bson:"out_trade_no"`         // 第三方订单号
	PackageId      string                 `json:"packageId" bson:"package_id"`            // 包名
	ShopId         int32                  `json:"shopId" bson:"shop_id"`                  // 商品ID
	ShopName       string                 `json:"ShopName" bson:"shop_name"`              // 商品名称
	ExamineWay     int32                  `json:"examineWay" bson:"examine_way"`          // 审核方式
	ExamineAccount string                 `json:"ExamineAccount" bson:"examine_account"`  // 审核人
	RefuseReason   string                 `json:"refuseReason" bson:"refuse_reason"`      // 支付方返回错误信息
	OutTradeStatus string                 `json:"outTradeStatus" bson:"out_trade_status"` // 支付方返回订单状态
	Utime          time.Time              `bson:"utime" json:"utime"`                     // 订单更新时间
	ErrorMsg       string                 `json:"errorMsg" bson:"error_msg"`              // 自动审核错误信息
	ETime          time.Time              `json:"etime" bson:"e_time"`                    // 操作时间
	UserType       int                    `json:"userType" bson:"user_type"`              // 玩家类型 0:A类  1:B类
	Diamond        int64                  `json:"diamond" bson:"diamond"`                 // 玩家提单时携带金额
	RepeatTimes    int                    `json:"repeatTimes" bson:"repeat_times"`        // 重复提现次数
	TransferDetail []*TransferOrderDetail `json:"transferDetail" bson:"transfer_detail"`  // 转单详情

	OvertimePayTacks     int64 `json:"overtimePayTacks" bson:"overtime_pay_tacks"`          // 提现超时赔付已领取金额
	OvertimePayTackTimes int32 `json:"overtimePayTackTimes" bson:"overtime_pay_tack_times"` // 提现超时赔付已领取次数
	KeepWaitTime         int64 `json:"keepWaitTime" bson:"keep_wait_time"`                  // 超过可退款时间后,玩家点击KeepWait时间戳(秒)
	RefundCompensate     int64 `json:"refundCompensate" bson:"refund_compensate"`           // 提现退款补偿
	RefundCompensateType int32 `json:"refundCompensateType" bson:"refund_compensate_type"`  // 提现退款补偿类型:1代表bonus,2代表cash,3代表withdrawable
}

type TransferOrderDetail struct {
	ChannelId uint32 `json:"channelId" bson:"channel_id"` // 渠道id
	TType     int    `json:"tType" bson:"t_type"`         // 转单类型
	Status    int    `json:"status" bson:"status"`        // 状态
	UserName  string `json:"userName" bson:"user_name"`   // 用户名
	Ctime     int64  `json:"ctime" bson:"ctime"`          // 创建时间
	Utime     int64  `json:"utime" bson:"utime"`          // 更新时间
	OrderID   string `bson:"-" json:"orderid"`            // 订单ID, 冗余字段同步ck时填充
	Userid    string `bson:"-" json:"user_id" `           // 用户id, 冗余字段同步ck时填充
	Reason    int    `json:"reason" bson:"reason"`        // 转单原因 0:正常 1:超时 2.超时急需 3.失败
}

// 提现日志
type WithdrawLog struct {
	Id     string `json:"id" bson:"id"`          // 订单id
	Ctime  int64  `json:"ctime" bson:"ctime"`    // 时间
	Amount uint32 `json:"amount" bson:"amount"`  // 金额
	Status int    `json:"status" bson:"status"`  // 状态
	PayWay int32  `json:"payWay" bson:"pay_way"` // 支付方式0bank,1ustd
}

// 审核操作
type WithdrawOpreate struct {
	OrderID  string `json:"orderid"`  // 订单ID
	Op       int    `json:"opreate"`  // 操作
	UserName string `json:"username"` // 操作人
}

func GetWithdrawList() []Withdraw {
	var list []Withdraw
	ListByQ(SetWithdraws, nil, &list)
	return list
}

// Save 写入数据库
func (w *Withdraw) Save() bool {
	//t.Ctime = bson.Now()
	return Upsert(SetWithdraws, bson.M{"_id": w.Id}, w)
}

// del 写入数据库
func DelWithdraw(k any) bool {
	//t.Ctime = bson.Now()
	return Delete(SetWithdraws, bson.M{"_id": k})
}

// 交易结果记录
func (t *WithdrawRecord) Get() {
	Get(WithdrawRecords, t.OrderID, t)
}

func (t *WithdrawRecord) Has() bool {
	return Has(WithdrawRecords, bson.M{"_id": t.OrderID})
}

func (t *WithdrawRecord) Update() bool {
	// 发布到kafka
	defer DataProducers.PublishDatas(mq.TopicSyncWithdrawRecord, []any{t})

	t.Utime = bson.Now()
	return Update(WithdrawRecords, bson.M{"_id": t.OrderID}, t)
}

func (t *WithdrawRecord) Save() bool {
	// 发布到kafka
	defer DataProducers.PublishDatas(mq.TopicSyncWithdrawRecord, []any{t})

	t.Ctime = bson.Now()
	return Insert(WithdrawRecords, t)
}

func (t *WithdrawRecord) Upsert() bool {
	// 发布到kafka
	defer DataProducers.PublishDatas(mq.TopicSyncWithdrawRecord, []any{t})

	t.Utime = bson.Now()
	return Upsert(WithdrawRecords, bson.M{"_id": t.OrderID}, t)
}

func (t *WithdrawRecord) AddRepeatTimes() {
	// 发布到kafka
	defer func() {
		if len(t.TransferDetail) > 0 {
			// 发布到kafka
			t.RepeatTimes += 1
			DataProducers.PublishDatas(mq.TopicSyncWithdrawRecord, []any{t})
		}
	}()

	Update(WithdrawRecords, bson.M{"_id": t.OrderID}, bson.M{"$inc": bson.M{"repeat_times": 1}, "$set": bson.M{"transfer_detail": t.TransferDetail}})
}

func (t *WithdrawRecord) UpdateTransferLog() {
	// 发布到kafka
	defer func() {
		if len(t.TransferDetail) > 0 {
			trans := t.TransferDetail[len(t.TransferDetail)-1]
			trans.OrderID = t.OrderID
			trans.Userid = t.Userid
			DataProducers.PublishDatas(mq.TopicSyncWithdrawRecordTransferOrderDetail, []any{trans})
		}
	}()

	Update(WithdrawRecords, bson.M{"_id": t.OrderID}, bson.M{"$set": bson.M{"transfer_detail": t.TransferDetail}})
}

func WithdrawRecordListByQ(q bson.M) []*WithdrawRecord {
	var list []*WithdrawRecord
	ListByQ(WithdrawRecords, q, &list)
	return list
}

func (t *WithdrawRecord) UpdateOrderStatus(status int32) {
	defer DataProducers.PublishDatas(mq.TopicSyncWithdrawRecord, []any{t})

	t.Utime = bson.Now()
	t.OrderStatus = status
	Update(WithdrawRecords, bson.M{"_id": t.OrderID}, bson.M{"$set": bson.M{"order_status": status, "utime": t.Utime}})
}

func (t *WithdrawRecord) GetTransferReason() int {
	switch t.OrderStatus {
	case DeliveryFailTransfer, WithdrawFailTransfer:
		return FailedTransfer
	case WithdrawTimeoutTransfer:
		return TimeoutTransfer
	case WithdrawUrgentTransfer:
		return TimeoutUrgentTransfer
	}
	return NormalTransfer
}
