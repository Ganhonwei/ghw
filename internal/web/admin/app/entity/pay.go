package entity

import "time"

// 充值订单
type PayOrder struct {
	OrderID        string    `bson:"_id" json:"orderid"`             // 订单ID
	Userid         string    `bson:"userid" json:"user_id" `         // 用户id
	NickName       string    `json:"nickName" bson:"nick_name"`      // 昵称
	RealName       string    `json:"realName" bson:"real_name"`      // 真实名称
	Email          string    `json:"email" bson:"email"`             // 邮箱
	Mobile         string    `json:"mobile" bson:"mobile"`           // 手机号
	Amount         uint32    `json:"amount" bson:"amount"`           // 金额
	RealAmount     uint32    `json:"real_amount" bson:"real_amount"` // 实际支付金额
	FAmount        float64   `bson:"-"`
	FRealAmount    string    `bson:"-"`
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
	ReportId       string    `json:"reportId" bson:"report_id"`              // 埋点id
	Gtype          int32     `json:"gtype" bson:"gtype"`                     // 游戏类型(大厅充值就是上一局游戏)
	UserType       int       `json:"userType" bson:"user_type"`              // 玩家类型 0:A类  1:B类
	IsFirst        bool      `json:"-"`                                      // 是否首充

	SuccessTimes int    `json:"successTimes" bson:"-"` // 成功次数
	AvgPay       string `json:"avgPay" bson:"-"`       // 均单价
	USDTMsg      string `json:"usdtMsg" bson:"-"`      // usdt订单信息
}

// 充值订单
type PayRecord struct {
	OrderID        string           `bson:"_id" json:"orderid"`                     // 订单ID
	Userid         string           `bson:"userid" json:"user_id" `                 // 用户id
	NickName       string           `json:"nickName" bson:"nick_name"`              // 昵称
	RealName       string           `json:"realName" bson:"real_name"`              // 真实名称
	Email          string           `json:"email" bson:"email"`                     // 邮箱
	Mobile         string           `json:"mobile" bson:"mobile"`                   // 手机号
	Amount         uint32           `json:"amount" bson:"amount"`                   // 金额
	RealAmount     uint32           `json:"real_amount" bson:"real_amount"`         // 实际支付金额
	Score          uint32           `json:"score" bson:"score"`                     // 金币数
	OtherPresent   string           `json:"otherPresent" bson:"other_present"`      // 额外赠送
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

	// ... UID 昵称
	FPackageAlias string `json:"fPackageAlias" bson:"-"` // 渠道别名
	FPackageClass string `json:"fPackageClass" bson:"-"` // 渠道类
	FChannel      string `json:"fChannel" bson:"-"`      // 支付通道
	FChannelType  string `json:"fChannelType" bson:"-"`  // 通道类型
	// ... 订单号 第三方订单号
	FCtime   string `json:"fCtime" bson:"-"`   // 订单生成时间
	FPayTime string `json:"fPayTime" bson:"-"` // 完成时间
	// ...商品名称
	FShopType     string   `json:"fShopType" bson:"-"`     // 商品类型
	FFirstPay     string   `json:"fFirstPay" bson:"-"`     // 是否首充
	FCash         string   `json:"fCash" bson:"-"`         // 到账cash
	FGiveCash     string   `json:"fGiveCash" bson:"-"`     // 赠送cash金额
	FAmount       string   `json:"fAmount" bson:"-"`       // 充值金额
	FRealAmount   string   `json:"fRealAmount" bson:"-"`   // 实际支付金额
	FOtherPresent string   `json:"fOtherPresent" bson:"-"` // 赠送金额
	FOrderStatus  string   `json:"fOrderStatus" bson:"-"`  // 订单状态
	FUtrs         []string `json:"fUtrs" bson:"-"`         // UTR
	ErrorMsgs     []string `json:"errorMsgs" bson:"-"`     // 标记
	FRepair       string   `json:"fRepair" bson:"-"`       // 手动补单信息
}

type TradeRecordUTR struct {
	UTR  string `json:"utr" bson:"utr"`
	Time int64  `json:"time" bson:"time"`
}

type TotalPayOrder struct {
	SuccessfulOrder   int64   `json:"successfulOrder" bson:"successful_order"`
	SuccessfulAmount  int64   `json:"successfulAmount" bson:"successful_amount"`
	TotalOrder        int64   `json:"totalOrder" bson:"total_order"`
	TotalAmount       int64   `json:"totalAmount" bson:"total_amount"`
	FSuccessfulAmount float64 `json:"fSuccessfulAmount" bson:"f_successful_amount"`
	FTotalAmount      float64 `json:"fTotalAmount" bson:"f_total_amount"`
}

// 提现订单
type WithdrawOrder struct {
	OrderID        string                 `bson:"_id" json:"orderid"`        // 订单ID
	Userid         string                 `bson:"userid" json:"user_id" `    // 用户id
	NickName       string                 `json:"nickName" bson:"nick_name"` // 昵称
	RealName       string                 `json:"realName" bson:"real_name"` // 真实名称
	Email          string                 `json:"email" bson:"email"`        // 邮箱
	Mobile         string                 `json:"mobile" bson:"mobile"`      // 手机号
	Total          uint32                 `json:"total" bson:"total"`        //提现金额
	FTotal         float64                `bson:"-"`
	Amount         uint32                 `json:"amount" bson:"amount"` // 实际提现金额
	FAmount        float64                `bson:"-"`
	Score          uint32                 `json:"score" bson:"score"`           // 移除彩金数
	Commission     uint32                 `json:"commission" bson:"commission"` // 手续费
	FCommission    float64                `bson:"-"`
	Blank          string                 `json:"blank" bson:"blank"`                     // 用户银行
	BlankNumber    string                 `json:"blankNumber" bson:"blank_number"`        // 银行卡号
	IFSC           string                 `json:"ifsc" bson:"ifsc"`                       // ifsc
	OutChannel     string                 `json:"outChannel" bson:"out_channel"`          // 代付渠道
	OutChannelStr  string                 `json:"outChannelStr" bson:"-"`                 //代付渠道名称
	OrderStatus    int                    `json:"orderStatus" bson:"order_status" `       // 订单状态
	Ctime          time.Time              `bson:"ctime" json:"ctime"`                     // 订单时间
	OrderAddress   string                 `json:"orderAddress" bson:"order_address"`      // 下单地址IP
	WithdrawTime   time.Time              `json:"payTime" bson:"pay_time"`                // 代付完成时间
	OutTradeNo     string                 `json:"outTradeNo" bson:"out_trade_no"`         // 第三方订单号
	PackageId      string                 `json:"packageId" bson:"package_id"`            // 包名
	ShopId         int32                  `json:"shopId" bson:"shop_id"`                  // 商品ID
	ShopName       string                 `json:"ShopName" bson:"shop_name"`              // 商品名称
	ExamineWay     int                    `json:"examineWay" bson:"examine_way"`          // 审核方式
	ExamineAccount string                 `json:"ExamineAccount" bson:"examine_account"`  // 审核人
	RefuseReason   string                 `json:"refuseReason" bson:"refuse_reason"`      // 支付方返回错误信息
	OutTradeStatus string                 `json:"outTradeStatus" bson:"out_trade_status"` // 支付方返回订单状态
	IsTag          bool                   `json:"isTag" bson:"is_tag"`                    // 订单是否标记
	Utime          time.Time              `bson:"utime" json:"utime"`                     // 订单更新时间
	ErrorMsg       string                 `json:"errorMsg" bson:"error_msg"`              // 自动审核错误信息
	ErrorMsgs      []string               `bson:"-"`                                      // 自动审核错误信息
	ETime          time.Time              `json:"etime" bson:"e_time"`                    // 操作时间
	IsBlankNumber  bool                   `bson:"-"`
	AndroidScore   int                    `bson:"-"`                                     // 安卓评分
	FMoney         float64                `bson:"-"`                                     // 充值总金额
	FCashOut       float64                `bson:"-"`                                     // 提现总金额
	UserType       int                    `json:"userType" bson:"user_type"`             // 玩家类型 0:A类  1:B类
	Diamond        int64                  `json:"diamond" bson:"diamond"`                // 玩家提单时携带金额
	RepeatTimes    int                    `json:"repeatTimes" bson:"repeat_times"`       // 重复提现次数 打款1次记一次
	TransferDetail []*TransferOrderDetail `json:"transferDetail" bson:"transfer_detail"` // 转单详情
	DelayTime      float64                `bson:"-"`                                     // 滞单时长
	TimeOfArrival  float64                `bson:"-"`                                     // 到账时效
	AutioTime      float64                `bson:"-"`                                     // 审核时长
	IsDelay        bool                   `bson:"-"`                                     // 是否超过滞单时长预警
	TransferCount  int                    `bson:"-"`
	USDTMsg        string                 `bson:"-"`
	Regtime        time.Time              `bson:"-"` // 注册时间

	Game1TimesStats      string `bson:"-"` // 玩局数最多的游戏/返奖率
	Game1BetStats        string `bson:"-"` // 打码最多的游戏/返奖率
	RebateRate           string `bson:"-"` // 玩家总返奖率 （玩家携带金额+玩家已提现到账金额+玩家提单待审核金额）/玩家总充值成功金额
	RechargeSuccessRate  string `bson:"-"` // 玩家充值成功率
	WithdrawSuccessRate  string `bson:"-"` // 玩家提现成功率
	RechargeAmount       string `bson:"-"` // 总充值金额
	RechargeAmount0      int64  `bson:"-"` // 总充值金额
	WithdrawAmountAll    string `bson:"-"` // 总提单金额 (包括被退回的、冻结的这些所有的)
	WithdrawAmount       string `bson:"-"` // 已提现到账金额
	WithdrawAmount0      int64  `bson:"-"` // 已提现到账金额
	WithdrawWaitAmount   string `bson:"-"` // 提单待审核金额
	WithdrawWaitAmount0  int64  `bson:"-"` // 提单待审核金额
	WithdrawOtherAmount  string `bson:"-"` // 其他提现金额
	WithdrawOtherAmount0 int64  `bson:"-"` // 其他提现金额
	AliasId              string `bson:"-"` // 渠道别名

	FProfit          string  `bson:"-"`           // 玩家当前净盈利
	BackAmount5      string  `bson:"-"`           // 提现失败被自动退回金额
	BackAmount7      string  `bson:"-"`           // 提单失败被自动退回金额
	FreezeAmount     string  `bson:"-"`           // 提现被冻结金额
	ProcessingAmount string  `bson:"-"`           // 审核通过未到账金额
	FAfterDiamond    float64 `bson:"-"`           // 玩家提单后携带金额
	FDiamond         string  `bson:"-"`           // 玩家当前携带金额
	OrderStatusName  string  `json:"-" bson:"-" ` // 订单状态名称
}

// 提现操作记录
type WithdrawOptRecord struct {
	Id      string `bson:"_id" json:"id"`           // id
	OrderId string `json:"orderId" bson:"order_id"` // 订单ID
	OptId   int64  `json:"optId" bson:"opt_id"`     // 操作类型ID 1:自动审核;2:自动退回;3:手动通过;4:标记；5:退回;6:冻结;7:转单;8:自动审核限制;9封号
	OptName string `json:"optName" bson:"opt_name"` // 操作人
	Ctime   int64  `json:"ctime" bson:"ctime"`      // 操作时间
}

type TransferOrderDetail struct {
	ChannelId   uint32 `json:"channelId" bson:"channel_id"` // 渠道id
	TType       int    `json:"tType" bson:"t_type"`         // 转单类型 1.非转单；2.自动状态；3.手动转单
	Status      int    `json:"status" bson:"status"`        // 状态
	UserName    string `json:"userName" bson:"user_name"`   // 用户名
	CTime       int64  `json:"cTime" bson:"ctime"`          // 创建时间
	UTime       int64  `json:"uTime" bson:"utime"`          // 修改时间
	ChannelName string `bson:"-"`                           // 支付渠道名称
	TTypeName   string `bson:"-"`
	ReasonName  string `bson:"-"`
	StatusName  string `bson:"-"`
	Reason      int    `json:"reason" bson:"reason"` // 转单原因 0:正常 1:超时 2.超时急需 3.失败
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
	DelayTime     int64     `json:"delay_time" bson:"delay_time"` // 滞留时长告警值(分)
	Index         int32     `bson:"index" json:"index"  `         //展示权重
	Ctime         time.Time `json:"ctime" bson:"ctime"`           // 创建时间
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

type WithdrawSetting struct {
	Id               string  `json:"id" bson:"_id"`                                // 设置id,用户类型 新手平民普小中大超大*7
	JQWithdrawMax    int64   `json:"jq_withdraw_max" bson:"jq_withdraw_max"`       // 机审单价上限
	JQWinMax         int64   `json:"jq_win_max" bson:"jq_win_max"`                 // 机审赢钱上限
	JQWinRateMax     float64 `json:"jq_win_rate_max" bson:"jq_win_rate_max"`       // 机审返奖率上限
	IsBankRepeatable bool    `json:"is_bank_repeatable" bson:"is_bank_repeatable"` // 银行卡号是否可重复
}

type PhoneUpload struct {
	Phone string `xls:"支付号码"` //支付手机号
}

// 用户充值、提现金额汇总
type UserCashRecod struct {
	Userid         string    `json:"id" bson:"_id"`                          // 用户id
	PayAmount      int64     `json:"payAmount" bson:"pay_amount"`            // 充值金额
	PayCount       int64     `json:"payCount" bson:"pay_count"`              // 充值次数
	WithdrawAmount int64     `json:"withdrawAmount" bson:"withdraw_amount"`  // 提现金额
	WithdrawCount  int64     `json:"withdrawCount" bson:"withdraw_count"`    // 提现次数
	GainAmout      int64     `json:"gainAmout" bson:"gain_amout"`            // 盈利金额
	FirstPayAmount int64     `json:"firstPayAmount" bson:"first_pay_amount"` // 首次充值金额
	FirstPayDate   time.Time `json:"firstPayDate" bson:"first_pay_date"`     // 首次充值金额时间
	CardNumber     []string  `json:"cardNumber" bson:"card_number"`          // 提现使用过的银行卡
}

// 提现汇总统计
type WithdrawStatsData struct {
	Date             int64   `bson:"-"` // 日期
	SDate            string  `bson:"-"` // 日期
	TotalNumber      int64   `bson:"-"` // 提现总单数
	TotalAmount      float64 `bson:"-"` // 提现总额
	PassNumber       int64   `bson:"-"` // 总通过单数
	PassAmount       float64 `bson:"-"` // 通过总额
	RgPassNumber     int64   `bson:"-"` // 人工通过单数
	JqPassNumber     int64   `bson:"-"` // 自动通过单数
	RgAuditNUmber    int64   `bson:"-"` // 人工审核单数
	JqAuditNUmber    int64   `bson:"-"` // 机器审核单数
	RgAuditRatio     float64 `bson:"-"` // 人审占比
	JqAuditRatio     float64 `bson:"-"` // 机审占比
	RefuseNumber     int64   `bson:"-"` // 被拒单数
	RefuseAmount     float64 `bson:"-"` // 被拒总额
	RefuseAvg        float64 `bson:"-"` // 被拒单均价
	SuccessfulNumber int64   `bson:"-"` // 付成单数
	SuccessfulAmount float64 `bson:"-"` // 付成总额
	SuccessfulAvg    float64 `bson:"-"` // 付成单均价
	PassRatio        float64 `bson:"-"` // 通过率
	SuccessfulRatio  float64 `bson:"-"` // 通过订单的付成率
	TotalAuditTime   float64 `bson:"-"` // 总审核时长
	DelayTimeTime    float64 `bson:"-"` // 总滞单时长
	ArrivalTimeTime  float64 `bson:"-"` // 总到账时长
	AuditTimeAvg     float64 `bson:"-"` // 平均审核时长
	DelayTimeAvg     float64 `bson:"-"` // 平均滞单时长
	ArrivalTimeAvg   float64 `bson:"-"` // 平均到账时长
	ZdNumber         int64   `bson:"-"` // 转单数
	ZdSGNumber       int64   `bson:"-"` // 转单付成单数
	ZdRatio          float64 `bson:"-"` // 转单付成率
	CfNumber         int64   `bson:"-"` // 重复付款单数
	CfCount          int64   `bson:"-"` // 重复付款总次数
	CfAmount         int64   `bson:"-"` // 重复付款金额
	FCfAmount        float64 `bson:"-"` // 浮点数 重复付款金额
	CfFrequency      float64 `bson:"-"` // 重复付款频次
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

// 支付渠道统计 收付汇总
type PayChannelCollect struct {
	Date                   string // 日期
	ChannelClass           string // 渠道类
	Channel1               string // 渠道别名
	PayChannel             string // 支付渠道
	PayAmount              string // 代收金额（玩家实付）
	PayAmountTax           string // 代收税费
	WithdrawAmount         string // 代付金额（渠道实付）
	WithdrawAmountTaxRate  string // 代付税费（按率）
	WithdrawAmountTaxTimes string // 代付税费（按单）
	Earning                string // 理论实收入
	Expense                string // 理论实付出
	Profit                 string // 理论净收益
	EarningActual          string // 渠道实收入（已扣税）
	ExpenseActual          string // 渠道实付出（已加税）
	ProfitActual           string // 渠道收-支
	ProfitDiff             string // 理论与实际的差额
	PayChannelBalance      string // 渠道实时余额（接口值）

	PayAmount0              int64
	PayAmountTax0           float64
	WithdrawAmount0         int64
	WithdrawAmountTaxRate0  float64
	WithdrawAmountTaxTimes0 int64
	PayChannelBalance0      int64

	PackageMap map[string]bool
}

// 支付渠道日志
type PayChannelLog struct {
	Id              string    `bson:"_id"`                           // 主键id
	Date            int64     `bson:"date"`                          // 日期
	SDate           time.Time `bson:"-"`                             // 统计日期
	OrderId         string    `json:"orderId" bson:"order_id"`       // 订单ID
	PayChannel      uint32    `json:"payChannel" bson:"pay_channel"` // 支付渠道
	PayChannelName  string    `bson:"-"`
	PType           int32     `json:"pType" bson:"p_type"`                   // 支付类型 1：代收;2:代付
	UserId          string    `json:"userId" bson:"user_id"`                 // 玩家ID
	Amount          int64     `json:"amount" bson:"amount"`                  // 金额
	FAmount         float64   `bson:"-"`                                     // 金额
	Rate            float64   `json:"rate" bson:"rate"`                      // 税率
	ActualAmount    int64     `json:"actualAmount" bson:"actual_amount"`     // 实际金额
	FActualAmount   float64   `bson:"-"`                                     // 实际金额
	HandlingCharge  int64     `json:"handlingCharge" bson:"handling_charge"` // 税费
	FHandlingCharge float64   `bson:"-"`                                     // 税费
	Balance         int64     `json:"balance" bson:"balance"`                // 实时余额
	FBalance        float64   `bson:"-"`                                     // 实时余额
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

// usdt下单响应
type UsdtPayOrderRespond struct {
	Code      int32               `json:"status_code"`
	Msg       string              `json:"message"`
	RequestId string              `json:"request_id"`
	Data      UsdtPayOrderRspData `json:"data"`
}
type UsdtPayOrderRspData struct {
	TradeId        string  `json:"trade_id"`        // 交易号
	OrderId        string  `json:"order_id"`        // 请求支付订单号
	Amount         float64 `json:"amount"`          // 请求支付金额 CNY,保留2位小数
	ActualAmount   string  `json:"actual_amount"`   // 实际需要支付的金额 USDT,保留四位小数
	Token          string  `json:"token"`           // 钱包地址
	ExpirationTime int64   `json:"expiration_time"` // 过期时间 时间戳秒
	PaymentUrl     string  `json:"payment_url"`     // 收银台地址
}

// usdt提现下单响应
type UsdtPayWithdrawRespond struct {
	Code      int32                  `json:"status_code"`
	Msg       string                 `json:"message"`
	RequestId string                 `json:"request_id"`
	Data      UsdtPayWithdrawRspData `json:"data"`
}
type UsdtPayWithdrawRspData struct {
	TradeId        string  `json:"trade_id"`        // 交易号
	OrderId        string  `json:"order_id"`        // 请求支付订单号
	Amount         float64 `json:"amount"`          // 请求支付金额 CNY,保留2位小数
	ActualAmount   string  `json:"actual_amount"`   // 实际需要支付的金额 USDT,保留四位小数
	ExpirationTime int64   `json:"expiration_time"` // 过期时间 时间戳秒
}

// usdt支付回调
type UsdtPayRechargeCallback struct {
	TradeId            string  `json:"trade_id"`             // 交易号
	OrderId            string  `json:"order_id"`             // 请求支付订单号
	Amount             float64 `json:"amount"`               // 请求支付金额 CNY,保留2位小数
	ActualAmount       string  `json:"actual_amount"`        // 实际需要支付的金额 USDT,保留四位小数
	Token              string  `json:"token"`                // 钱包地址
	BlockTransactionId string  `json:"block_transaction_id"` // 区块交易号
	Signature          string  `json:"signature"`            // 签名
	Status             string  `json:"status"`               // 订单状态 1：等待支付，2：支付成功，3：已过期；提现:3成功
}

// 支付渠道统计备注记录
type PayChannelStatRecords struct {
	Id             string    `json:"id" bson:"_id"`                 //
	Date           int64     `bson:"date"`                          // 统计日期时间戳
	SDate          time.Time `bson:"-"`                             // 统计日期
	PayChannel     uint32    `json:"payChannel" bson:"pay_channel"` // 支付渠道
	PayChannelName string    `bson:"-"`
	Remark         string    `json:"remark" bson:"remark"`    // 备注
	OptName        string    `json:"optName" bson:"opt_name"` // 操作人
	Ctime          int64     `json:"ctime" bson:"ctime"`      // 操作时间
	CDate          time.Time `bson:"-"`                       // 操作时间
}

// 回收周期消耗记录
type RecoverycycleDateConsume struct {
	Id       string    `bson:"_id"`
	Date     string    `bson:"date"`
	Consume  float64   `bson:"consume"`
	Ptype    int       `bson:"ptype"`    // 0全部,1渠道类,2别名,3渠道,4混合
	Packages string    `bson:"packages"` // 选择的相关渠道
	Ctime    time.Time `bson:"ctime"`
	OptName  string    `bson:"opt_name"` // 操作人
}

// 支付渠道统计tg通知配置
type PayChannelStatNoticeConfig struct {
	Id string `bson:"_id" json:"id"` // 游戏ID
	// 当天统计通知时间
	DayTicker   int64 `bson:"day_ticker" json:"day_ticker"`       // 当天统计触发一次时间(分钟) 0不触发
	DayNextTime int64 `bson:"day_next_time" json:"day_next_time"` // 当天统计下次触发时间戳(秒)
	// 时段统计时间
	TimeTicker   int64 `bson:"time_ticker" json:"time_ticker"`       // 时段统计单次时段(分钟) 0不触发
	TimeNextTime int64 `bson:"time_next_time" json:"time_next_time"` // 时段统计触发时间(下次触发时间)
	// 提现标记订单时间
	WithdrawTicker   int64 `bson:"withdraw_ticker" json:"withdraw_ticker"`       // 提现标记统计单次时段(分钟) 0不触发
	WithdrawNextTime int64 `bson:"withdraw_next_time" json:"withdraw_next_time"` // 提现标记统计触发时间(下次触发时间)
	// 支付订单UTR补分
	PayUtrTicker   int64 `bson:"pay_utr_ticker" json:"payUtrTicker"`      // 支付订单UTR补分通知单次时段(分钟) 0不触发
	PayUtrNextTime int64 `bson:"pay_utr_next_time" json:"payUtrNextTime"` // 支付订单UTR补分通知(下次触发时间)
}

type PayChannelStat struct {
	Channel              string
	ChannelName          string
	TotalTimes           int64
	SuccessTimes         int64
	ErrorTimes           int64
	SuccessRate          string
	PayTotalTimes        int64
	PaySuccessTimes      int64
	PayErrorTimes        int64
	PaySuccessRate       string
	WithdrawTotalTimes   int64
	WithdrawSuccessTimes int64
	WithdrawErrorTimes   int64
	WithdrawSuccessRate  string
	// Time                 string // 时间段 00:00:00~10:11:12
}
