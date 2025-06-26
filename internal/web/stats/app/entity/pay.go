package entity

import "time"

// 充值订单
type PayOrder struct {
	OrderID        string    `bson:"_id" json:"orderid"`        // 订单ID
	Userid         string    `bson:"userid" json:"user_id" `    // 用户id
	NickName       string    `json:"nickName" bson:"nick_name"` // 昵称
	RealName       string    `json:"realName" bson:"real_name"` // 真实名称
	Email          string    `json:"email" bson:"email"`        // 邮箱
	Mobile         string    `json:"mobile" bson:"mobile"`      // 手机号
	Amount         uint32    `json:"amount" bson:"amount"`      // 金额
	FAmount        float64   `bson:"-"`
	Score          uint32    `json:"score" bson:"score"`                     // 金币数
	OtherPresent   uint32    `json:"otherPresent" bson:"other_present"`      // 额外赠送
	OrderStatus    int32     `json:"orderStatus" bson:"order_status" `       // 订单状态
	Ctime          time.Time `bson:"ctime" json:"ctime"`                     // 订单时间
	Utime          time.Time `bson:"utime" json:"utime"`                     // 订单更新时间
	OrderAddress   string    `json:"orderAddress" bson:"order_address"`      // 下单地址IP
	PayTime        time.Time `json:"payTime" bson:"pay_time"`                // 付款时间
	PayAddress     string    `json:"payAddress" bson:"pay_address"`          // 付款地址IP
	OutTradeNo     string    `json:"outTradeNo" bson:"out_trade_no"`         // 第三方订单号
	PackageId      string    `json:"packageId" bson:"package_id"`            // 包名
	ChannelId      uint32    `json:"channelId" bson:"channel_id"`            // 支付渠道
	ChannelName    string    `json:"-"`                                      // 支付渠道名称
	ShopId         string    `json:"shopId" bson:"shop_id"`                  // 商店ID
	ShopType       int32     `json:"shopType" bson:"shop_type"`              // 商品类型
	ShopName       string    `json:"ShopName" bson:"shop_name"`              // 商品名称
	RefuseReason   string    `json:"refuseReason" bson:"refuse_reason"`      // 支付方返回错误信息
	OutTradeStatus string    `json:"outTradeStatus" bson:"out_trade_status"` // 支付方返回订单状态
}

// 提现订单
type WithdrawOrder struct {
	OrderID        string    `bson:"_id" json:"orderid"`        // 订单ID
	Userid         string    `bson:"userid" json:"user_id" `    // 用户id
	NickName       string    `json:"nickName" bson:"nick_name"` // 昵称
	RealName       string    `json:"realName" bson:"real_name"` // 真实名称
	Email          string    `json:"email" bson:"email"`        // 邮箱
	Mobile         string    `json:"mobile" bson:"mobile"`      // 手机号
	Total          uint32    `json:"total" bson:"total"`        //提现金额
	FTotal         float64   `bson:"-"`
	Amount         uint32    `json:"amount" bson:"amount"` // 实际提现金额
	FAmount        float64   `bson:"-"`
	Score          uint32    `json:"score" bson:"score"`           // 移除彩金数
	Commission     uint32    `json:"commission" bson:"commission"` // 手续费
	FCommission    float64   `bson:"-"`
	Blank          string    `json:"blank" bson:"blank"`                     // 用户银行
	BlankNumber    string    `json:"blankNumber" bson:"blank_number"`        // 银行卡号
	IFSC           string    `json:"ifsc" bson:"ifsc"`                       // ifsc
	OutChannel     string    `json:"outChannel" bson:"out_channel"`          // 代付渠道
	OutChannelStr  string    `json:"outChannelStr" bson:"-"`                 //代付渠道名称
	OrderStatus    int       `json:"orderStatus" bson:"order_status" `       // 订单状态
	Ctime          time.Time `bson:"ctime" json:"ctime"`                     // 订单时间
	OrderAddress   string    `json:"orderAddress" bson:"order_address"`      // 下单地址IP
	WithdrawTime   time.Time `json:"payTime" bson:"pay_time"`                // 代付完成时间
	OutTradeNo     string    `json:"outTradeNo" bson:"out_trade_no"`         // 第三方订单号
	PackageId      string    `json:"packageId" bson:"package_id"`            // 包名
	ShopId         int32     `json:"shopId" bson:"shop_id"`                  // 商品ID
	ShopName       string    `json:"ShopName" bson:"shop_name"`              // 商品名称
	ExamineWay     int       `json:"examineWay" bson:"examine_way"`          // 审核方式
	ExamineAccount string    `json:"ExamineAccount" bson:"examine_account"`  // 审核人
	RefuseReason   string    `json:"refuseReason" bson:"refuse_reason"`      // 支付方返回错误信息
	OutTradeStatus string    `json:"outTradeStatus" bson:"out_trade_status"` // 支付方返回订单状态
	IsTag          bool      `json:"isTag" bson:"is_tag"`                    // 订单是否标记
	Utime          time.Time `bson:"utime" json:"utime"`                     // 订单更新时间
	ErrorMsg       string    `json:"errorMsg" bson:"error_msg"`              // 自动审核错误信息
	ETime          time.Time `json:"etime" bson:"e_time"`                    // 操作时间
	IsBlankNumber  bool      `bson:"-"`
	AndroidScore   int       `bson:"-"` // 安卓评分
	FMoney         float64   `bson:"-"` // 充值总金额
	FCashOut       float64   `bson:"-"` // 提现总金额
}

// 提现玩家用户ID、银行卡号
type WithdrawAccount struct {
	OrderID     string `bson:"-" ` // 订单ID
	Userid      string `bson:"-"`  // 用户id
	BlankNumber string `bson:"-"`  // 银行卡号
}

// 后台主动赠送金币日志
type GoldGiftLog struct {
	Id           string    `json:"id" bson:"_id"`             // id
	Userid       string    `json:"userId" bson:"user_id"`     // 用户ID
	Nickname     string    `json:"nickname" bson:"nickname"`  // 用户昵称
	BundleId     string    `json:"bundleId" bson:"bundle_id"` // 包名
	GiftType     int       `json:"giftType" bson:"gift_type"` // 赠送类型 1:充值彩金;2:修改彩金;3:充值奖励金;4:修改奖励金;5:增加可提现金额
	Score        int64     `json:"score" bson:"score"`        // 赠送分数
	FScore       float64   `bson:"-"`
	CurScore     int64     `json:"curScore" bson:"cur_score"` // 赠送之前分数
	FCurScore    float64   `bson:"-"`
	ChangeScore  int64     `json:"changeScore" bson:"change_score"` // 赠送之后分数
	FChangeScore float64   `bson:"-"`
	Remark       string    `json:"remark" bson:"remark"`           // 备注
	Operator     string    `json:"operator" bson:"operator"`       // 操作人
	RemarkPerson string    `json:"remarkPerson" bson:"rem_person"` // 备注人
	Ctime        time.Time `json:"ctime" bson:"ctime"`             // 创建时间
}

type SetWithdraw struct {
	Id            int32     `bson:"_id" json:"id"`                        //提现Id
	Name          string    `bson:"name" json:"name"`                     //商品名称
	Limit         int32     `bson:"limit" json:"limit" `                  //提现次数限制
	RechargeLimit int64     `bson:"recharge_limit" json:"recharge_limit"` //充值金额限制
	FlowingLimit  uint32    `bson:"flowing_limit" json:"flowing_limit"`   //流水限制
	CostDiamond   uint32    `bson:"cost_diamond" json:"cost_diamond"`     //消耗彩金
	FCostDiamond  float64   `bson:"-" `
	GetCash       uint32    `bson:"get_cash" json:"get_cash"  ` //提现金额(分)
	FGetCash      float64   `bson:"-" `
	Commission    uint32    `bson:"commission" json:"commission" ` //手续费(分)
	FCommission   float64   `bson:"-" `
	Index         int32     `bson:"index" json:"index"  ` //展示权重
	Ctime         time.Time `json:"ctime" bson:"ctime"`   // 创建时间
}

// 提现日志
type WithdrawLog struct {
	Id     string `json:"id" bson:"id"`         // 订单id
	Ctime  int64  `json:"ctime" bson:"ctime"`   // 时间
	Amount uint32 `json:"amount" bson:"amount"` // 金额
	Status int    `json:"status" bson:"status"` // 状态
}

// 支付黑名单
type PayBlackList struct {
	Phone    string    `bson:"_id" json:"id"`            //支付手机号
	Operator string    `json:"operator" bson:"operator"` // 操作人
	Ctime    time.Time `bson:"ctime" json:"ctime"`       // 添加时间
}

type PhoneUpload struct {
	Phone string `xls:"支付号码"` //支付手机号
}

// 支付渠道统计
type PayChannelStatsData struct {
	Id                   string    `bson:"_id"`                                              // 主键id
	Date                 int64     `bson:"date"`                                             // 统计日期时间戳
	SDate                time.Time `bson:"-"`                                                // 统计日期
	PayChannel           uint32    `json:"payChannel" bson:"pay_channel"`                    // 支付渠道
	TotalPay             int64     `json:"totalPay" bson:"total_pay"`                        // 总代收流水
	PayRate              float64   `json:"payRate" bson:"pay_rate"`                          // 代收税率
	TotalPayAmount       int64     `json:"totalPayAmount" bson:"total_pay_amount"`           // 实际代收流水
	TotalWithdraw        int64     `json:"totalWithdraw" bson:"total_withdraw"`              // 总代付流水
	WithdrawRate         float64   `json:"withdrawRate" bson:"withdraw_rate"`                // 代付税率
	TotalWithdrawAmount  int64     `json:"totalWithdrawAmount" bson:"total_withdraw_amount"` // 实际代付流水
	PayHandlingFee       int64     `json:"payHandlingFee" bson:"pay_handling_fee"`           // 代收税费
	WithdrawHandlingFee  int64     `json:"withdrawHandlingFee" bson:"withdraw_handling_fee"` // 代付税费
	TotalProfit          float64   `bson:"-"`                                                // 总利润
	ActualProfit         float64   `bson:"-"`                                                // 实际利润
	YesterdayBalance     int64     `json:"yesterdayBalance" bson:"yesterday_balance"`        // 昨日0点余额
	TodayBalance         int64     `json:"todayBalance" bson:"today_balance"`                // 当日0点余额
	Balance              int64     `json:"balance" bson:"balance"`                           // 我统计余额
	CurrentBalance       int64     `json:"currentBalance" bson:"current_balance"`            // 当前余额
	PayChannelBalance    int64     `json:"payChannelBalance" bson:"pay_channel_balance"`     // 渠道记录余额
	AmountDifference     float64   `bson:"-"`                                                // 差额=我方统计余额-渠道统计余额
	PayChannelName       string    `bson:"-"`                                                // 支取渠道名称
	FTotalPay            float64   `bson:"-"`                                                // 总代收流水
	FTotalPayAmount      float64   `bson:"-"`                                                // 实际代收流水
	FTotalWithdraw       float64   `bson:"-"`                                                // 总代付流水
	FTotalWithdrawAmount float64   `bson:"-"`
	FPayHandlingFee      float64   `bson:"-"`
	FWithdrawHandlingFee float64   `bson:"-"`
	FYesterdayBalance    float64   `bson:"-"`
	FTodayBalance        float64   `bson:"-"`
	FCurrentBalance      float64   `bson:"-"`
	FPayChannelBalance   float64   `bson:"-"`
	FAmountDifference    float64   `bson:"-"`
	IsTotal              bool      `bson:"-"`
	IsActual             bool      `bson:"-"`
}

// 支付渠道统计
type ThirdPartyBalance struct {
	Id         string `bson:"_id"`                           // 主键id
	PayChannel uint32 `json:"payChannel" bson:"pay_channel"` // 支付渠道
	Balance    int64  `json:"balance" bson:"balance"`        // 渠道余额
	Ctime      int64  `json:"ctime" bson:"ctime"`            // 获取时间
	OwnBalance int64  `json:"ownBalance" bson:"own_balance"` // 平台自己的余额
	Difference int64  `json:"difference" bson:"difference"`  // 差额
}

type WebResponse struct {
	Code   int    `json:"code"` // 错误码    200:success
	ErrMsg string `json:"msg"`  // 错误信息
	Body   []byte `json:"body"` // 响应体
}

type WebQueryBalence struct {
	ChannelId uint32 `json:"channel_id"` // 渠道id
	Balence   int64  `json:"balence"`    // 可用余额
}
