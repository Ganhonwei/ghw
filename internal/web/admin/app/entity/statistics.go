package entity

import (
	"time"
)

// 数据汇总
type DataSummary struct {
	Id                  string    `bson:"_id"`                                              // 主键id
	Date                int64     `bson:"date"`                                             // 统计日期时间戳
	SDate               time.Time `bson:"-"`                                                // 统计日期
	NewRegister         int64     `json:"newRegister" bson:"new_register"`                  // 新注册数
	NewEquipment        int64     `json:"newEquipment" bson:"new_equipment"`                // 新设备数
	LoginNum            int64     `json:"loginNum" bson:"login_num"`                        // 登录人数
	NextDayRetention    float64   `json:"nextDayRetention" bson:"next_day_retention"`       // 次日留存
	PayRetention        float64   `json:"payRetention" bson:"pay_retention"`                // 次日付费留存
	NewRechargeNum      int64     `json:"newRechargeNum" bson:"new_recharge_num"`           // 新用户充值人数
	NewRechargeAmount   float64   `json:"newRechargeAmount" bson:"new_recharge_amount"`     // 新用户充值金额                                           // 新用户充值金额
	NewRechargeArpu     float64   `json:"newRechargeArpu" bson:"new_recharge_arpu"`         // 新用户充值ARPU=新用户充值金额÷新注册数
	NewRechargeArppu    float64   `json:"newRechargeArppu" bson:"new_recharge_arppu"`       // 新用户充值ARRPU=新用户充值金额÷新用户充值人数
	NewUserPaymentRate  float64   `json:"newUserPaymentRate" bson:"new_user_payment_rate"`  // 新用户付费率=新用户充值人数÷新注册数
	TotalRecharge       int64     `json:"totalRecharge" bson:"total_recharge"`              // 总充值人数
	TotalAmount         float64   `json:"totalAmount" bson:"total_amount"`                  // 总充值金额
	TotalRechargeArpu   float64   `json:"totalRechargeArpu" bson:"total_recharge_arpu"`     // 总充值ARRPU=总充值金额÷总充值人数
	TotalRechargeArppu  float64   `json:"totalRechargeArppu" bson:"total_recharge_arppu"`   // 总充值ARRPU=总充值金额÷总充值人数
	TotalPaymentRate    float64   `json:"totalPaymentRate" bson:"total_payment_rate"`       // 总付费率=总充值人数÷登录数
	TotalWithdrawNum    int64     `json:"totalWithdrawNum" bson:"total_withdraw_num"`       // 提现总人数
	TotalWithdrawAmount float64   `json:"totalWithdrawAmount" bson:"total_withdraw_amount"` // 提现总金额
	HandlingCharge      float64   `json:"handlingCharge" bson:"handling_charge"`            // 手续费
	CostRatio           float64   `json:"costRatio" bson:"cost_ratio"`                      // 成本比=提现总金额÷总充值金额
}

// 用户留存
type UserRetainedOld struct {
	Id           string    `bson:"_id"`                              // 主键id
	Date         int64     `bson:"date"`                             // 统计日期时间戳
	SDate        time.Time `bson:"-"`                                // 统计日期
	Channel      string    `json:"channel" bson:"channel"`           // 渠道
	Channel1     string    `json:"channel1" bson:"channel1"`         // 渠道别名
	ChannelClass string    `json:"channelClass" bson:"channelClass"` // 渠道别名
	SType        int       `bson:"s_type"`                           // 数据类型 0：全部;1：按渠道
	NewNumber    int64     `json:"newNumber" bson:"new_number"`      // 新增人数
	Login1       int64     `json:"login1" bson:"login1"`             // 1日留存=昨日新注册数今日登录÷昨日新注册数
	Register1    int64     `json:"register1" bson:"register1"`       // 1日留存=昨日新注册数今日登录÷昨日新注册数
	Login2       int64     `json:"login2" bson:"login2"`
	Register2    int64     `json:"register2" bson:"register2"`
	Login3       int64     `json:"login3" bson:"login3"`
	Register3    int64     `json:"register3" bson:"register3"`
	Login4       int64     `json:"login4" bson:"login4"`
	Register4    int64     `json:"register4" bson:"register4"`
	Login5       int64     `json:"login5" bson:"login5"`
	Register5    int64     `json:"register5" bson:"register5"`
	Login6       int64     `json:"login6" bson:"login6"`
	Register6    int64     `json:"register6" bson:"register6"`
	Login7       int64     `json:"login7" bson:"login7"`
	Register7    int64     `json:"register7" bson:"register7"`
	Login15      int64     `json:"login15" bson:"login15"`
	Register15   int64     `json:"register15" bson:"register15"`
	Login30      int64     `json:"login30" bson:"login30"`
	Register30   int64     `json:"register30" bson:"register30"`
	Login40      int64     `json:"login40" bson:"login40"`
	Register40   int64     `json:"register40" bson:"register40"`
	Login50      int64     `json:"login50" bson:"login50"`
	Register50   int64     `json:"register50" bson:"register50"`
	Login60      int64     `json:"login60" bson:"login60"`
	Register60   int64     `json:"register60" bson:"register60"`
	Day1         float64   `bson:"-"`
	Day2         float64   `bson:"-"`
	Day3         float64   `bson:"-"`
	Day4         float64   `bson:"-"`
	Day5         float64   `bson:"-"`
	Day6         float64   `bson:"-"`
	Day7         float64   `bson:"-"`
	Day15        float64   `bson:"-"`
	Day30        float64   `bson:"-"`
	Day40        float64   `bson:"-"`
	Day50        float64   `bson:"-"`
	Day60        float64   `bson:"-"`
}

// 用户留存
type UserRetained struct {
	SDate        string
	Channel1     string // 渠道别名
	ChannelClass string // 渠道类
	RegistCount  int64  // 新增人数
	PayCount     int64  // 累计付费人数
	PayRate      string // 累计付费率
	RepayCount   int64  // 累计复充人数
	RepayRate    string // 累计复充/付费
	BePayCount   int64  // 当日成为付费用户的人数
	BeRepayCount int64  // 当日成为复充用户的人数
	BeRCount     int64  // 累计R人数
	BeRRate      string // 累计R占比

	Day1Rate  string
	Day2Rate  string
	Day3Rate  string
	Day4Rate  string
	Day5Rate  string
	Day6Rate  string
	Day7Rate  string
	Day15Rate string
	Day30Rate string
	Day40Rate string
	Day50Rate string
	Day60Rate string
	Day90Rate string

	// Day1  int64
	// Day2  int64
	// Day3  int64
	// Day4  int64
	// Day5  int64
	// Day6  int64
	// Day7  int64
	// Day15 int64
	// Day30 int64
	// Day40 int64
	// Day50 int64
	// Day60 int64
}

// 付费留存
type PayUserRetained struct {
	Id          string    `bson:"_id"`                              // 主键id
	Date        int64     `bson:"date"`                             // 统计日期时间戳
	SDate       time.Time `bson:"-"`                                // 统计日期
	Channel     string    `json:"channel" bson:"channel"`           // 渠道
	Channel1    string    `json:"channel1" bson:"channel1"`         // 渠道别名
	PayUserType int       `json:"payUserType" bson:"pay_user_type"` // 充值用户类型 0:首充;2:复充
	NewNumber   int64     `json:"newNumber" bson:"new_number"`      // 充值人数
	Pay1        int64     `json:"pay1" bson:"pay1"`                 // 1日留存=昨日新注册数且充值今日登录÷昨日新注册数充值
	Register1   int64     `json:"register1" bson:"register1"`       // 1日留存=昨日新注册数且充值今日登录÷昨日新注册数充值
	Pay2        int64     `json:"pay2" bson:"pay2"`
	Register2   int64     `json:"register2" bson:"register2"`
	Pay3        int64     `json:"pay3" bson:"pay3"`
	Register3   int64     `json:"register3" bson:"register3"`
	Pay4        int64     `json:"pay4" bson:"pay4"`
	Register4   int64     `json:"register4" bson:"register4"`
	Pay5        int64     `json:"pay5" bson:"pay5"`
	Register5   int64     `json:"register5" bson:"register5"`
	Pay6        int64     `json:"pay6" bson:"pay6"`
	Register6   int64     `json:"register6" bson:"register6"`
	Pay7        int64     `json:"pay7" bson:"pay7"`
	Register7   int64     `json:"register7" bson:"register7"`
	Pay15       int64     `json:"pay15" bson:"pay15"`
	Register15  int64     `json:"register15" bson:"register15"`
	Pay30       int64     `json:"pay30" bson:"pay30"`
	Register30  int64     `json:"register30" bson:"register30"`
	Pay60       int64     `json:"pay60" bson:"pay60"`
	Register60  int64     `json:"register60" bson:"register60"`
	Day1        float64   `bson:"-"`
	Day2        float64   `bson:"-"`
	Day3        float64   `bson:"-"`
	Day4        float64   `bson:"-"`
	Day5        float64   `bson:"-"`
	Day6        float64   `bson:"-"`
	Day7        float64   `bson:"-"`
	Day15       float64   `bson:"-"`
	Day30       float64   `bson:"-"`
	Day60       float64   `bson:"-"`
}

// 充提排名
type PayRanking struct {
	Ranking        int64     `bson:"-"` // 排名
	UserId         string    `bson:"-"` // 用户ID
	RegistArea     int       `bson:"-"` // ab测试 0:A 1:B
	PayAmount      float64   `bson:"-"` // 充值金额
	WithdrawAmount float64   `bson:"-"` // 提现金额
	ProfitAmount   float64   `bson:"-"` // 盈利
	LoginTime      time.Time `bson:"-"` // 最后登录时间
}

// 渠道数据
type ChannelData struct {
	Id                   string    `bson:"_id"`                                                // 主键id
	Date                 int64     `bson:"date"`                                               // 统计日期时间戳
	SDate                time.Time `bson:"-"`                                                  // 统计日期
	Channel              string    `json:"channel" bson:"channel"`                             // 渠道
	Channel1             string    `json:"channel1" bson:"channel1"`                           // 渠道别名
	NewRegister          int64     `json:"newRegister" bson:"new_register"`                    // 新注册数
	NewEquipment         int64     `json:"newEquipment" bson:"new_equipment"`                  // 新设备数
	LoginNum             int64     `json:"loginNum" bson:"login_num"`                          // 登录人数
	NextDayRetention     float64   `json:"nextDayRetention" bson:"next_day_retention"`         // 次日留存
	NewRechargeNum       int64     `json:"newRechargeNum" bson:"new_recharge_num"`             // 新用户充值人数
	NewRechargeAmount    float64   `json:"newRechargeAmount" bson:"new_recharge_amount"`       // 新用户充值金额                                           // 新用户充值金额
	NewRechargeArpu      float64   `json:"newRechargeArpu" bson:"new_recharge_arpu"`           // 新用户充值ARPU=新用户充值金额÷新注册数
	NewRechargeArppu     float64   `json:"newRechargeArppu" bson:"new_recharge_arppu"`         // 新用户充值ARRPU=新用户充值金额÷新用户充值人数
	NewUserPaymentRate   float64   `json:"newUserPaymentRate" bson:"new_user_payment_rate"`    // 新用户付费率=新用户充值人数÷新注册数
	TotalRecharge        int64     `json:"totalRecharge" bson:"total_recharge"`                // 总充值人数
	TotalAmount          float64   `json:"totalAmount" bson:"total_amount"`                    // 总充值金额
	TotalRechargeArpu    float64   `json:"totalRechargeArpu" bson:"total_recharge_arpu"`       // 总充值ARRPU=总充值金额÷总充值人数
	TotalRechargeArppu   float64   `json:"totalRechargeArppu" bson:"total_recharge_arppu"`     // 总充值ARRPU=总充值金额÷总充值人数
	TotalPaymentRate     float64   `json:"totalPaymentRate" bson:"total_payment_rate"`         // 总付费率=总充值人数÷登录数
	PayRequest           int64     `json:"payRequest" bson:"pay_request"`                      // 充值请求人数
	PayRequestOrder      int64     `json:"payRequestOrder" bson:"pay_request_order"`           // 充值请求订单数
	PaySuccessOrder      int64     `json:"paySuccessOrder" bson:"pay_success_order"`           // 充值成功订单数
	PaySuccessRate       float64   `json:"paySuccessRate" bson:"pay_success_rate"`             // 充值成功率
	WithdrawRequest      int64     `json:"withdrawRequest" bson:"withdraw_request"`            // 提现请求人数
	WithdrawRequestOrder int64     `json:"withdrawRequestOrder" bson:"withdraw_request_order"` // 提现请求订单数
	WithdrawSuccessOrder int64     `json:"withdrawSuccessOrder" bson:"withdraw_success_order"` // 提现成功订单数
	WithdrawSuccessRate  float64   `json:"withdrawSuccessRate" bson:"withdraw_success_rate"`   // 提现成功率
	TotalWithdrawNum     int64     `json:"totalWithdrawNum" bson:"total_withdraw_num"`         // 提现总人数
	TotalWithdrawAmount  float64   `json:"totalWithdrawAmount" bson:"total_withdraw_amount"`   // 提现总金额
	HandlingCharge       float64   `json:"handlingCharge" bson:"handling_charge"`              // 手续费
	CostRatio            float64   `json:"costRatio" bson:"cost_ratio"`                        // 成本比=提现总金额÷总充值金额
}

// 数据汇总
type DataStatistics struct {
	Id                   string    `bson:"_id"`                                                // 主键id
	Date                 int64     `bson:"date"`                                               // 统计日期时间戳
	SDate                time.Time `bson:"-"`                                                  // 统计日期
	FDate                string    `bson:"-"`                                                  // 统计日期
	DateStr              string    `bson:"-"`                                                  // 统计日期
	Channel              string    `json:"channel" bson:"channel"`                             // 渠道
	Channel1             string    `json:"channel1" bson:"channel1"`                           // 渠道别名
	DataTypes            int       `json:"dataTypes" bson:"data_types"`                        // 数据类型 0:A类；1：B类
	NewRegister          int64     `json:"newRegister" bson:"new_register"`                    // 新注册数
	NewEquipment         int64     `json:"newEquipment" bson:"new_equipment"`                  // 新设备数
	LoginNum             int64     `json:"loginNum" bson:"login_num"`                          // 登录人数
	OldLoginNum          int64     `json:"oldLoginNum" bson:"old_login_num"`                   // 老用户当天登录人数
	NextNewRegister      int64     `json:"nextNewRegister" bson:"next_new_register"`           // 昨日新注册数
	NextLogin            int64     `json:"nextLogin" bson:"next_login"`                        // 昨日新注册今日登录数
	NextDayRetention     float64   `bson:"-"`                                                  // 次日留存
	JRLoginNextPay       int64     `bson:"jr_login_next_pay"`                                  // 今日登录昨日充值玩家
	ZRPayCount           int64     `bson:"zr_pay_count"`                                       // 昨日充值的玩家人数
	PayRetention         float64   `bson:"-"`                                                  // 次日付费留存
	OldRechargeNum       int64     `json:"oldRechargeNum" bson:"old_recharge_num"`             // 老用户充值人数
	OldRechargeAmount    int64     `json:"oldRechargeAmount" bson:"old_recharge_amount"`       // 老用户充值金额
	OldRechargeArpu      float64   `bson:"-"`                                                  // 老用户充值ARPU 老用户充值金额 / 老用户当天登录人数
	OldRechargeArppu     float64   `bson:"-"`                                                  // 老用户充值ARRPU 老用户充值金额/ 老用户人数
	OldUserPaymentRate   float64   `bson:"-"`                                                  // 老用户付费率 老用户付费人数/老用户当天登录人数
	NewRechargeNum       int64     `json:"newRechargeNum" bson:"new_recharge_num"`             // 新用户充值人数
	NewRechargeAmount    int64     `json:"newRechargeAmount" bson:"new_recharge_amount"`       // 新用户充值金额
	NewRechargeArpu      float64   `bson:"-"`                                                  // 新用户充值ARPU=新用户充值金额÷新注册数
	NewRechargeArppu     float64   `bson:"-"`                                                  // 新用户充值ARRPU=新用户充值金额÷新用户充值人数
	NewUserPaymentRate   float64   `bson:"-"`                                                  // 新用户付费率=新用户充值人数÷新注册数
	TotalRecharge        int64     `json:"totalRecharge" bson:"total_recharge"`                // 总充值人数
	TotalAmount          int64     `json:"totalAmount" bson:"total_amount"`                    // 总充值金额
	TotalRechargeArpu    float64   `bson:"-"`                                                  // 总充值ARPU=总充值金额÷登录人数
	TotalRechargeArppu   float64   `bson:"-"`                                                  // 总充值ARRPU=总充值金额÷总充值人数
	TotalPaymentRate     float64   `bson:"-"`                                                  // 总付费率=总充值人数÷登录数
	PayRequest           int64     `json:"payRequest" bson:"pay_request"`                      // 充值请求人数
	PayRequestOrder      int64     `json:"payRequestOrder" bson:"pay_request_order"`           // 充值请求订单数
	PaySuccessOrder      int64     `json:"paySuccessOrder" bson:"pay_success_order"`           // 充值成功订单数
	PaySuccessRate       float64   `bson:"-"`                                                  // 充值成功率
	WithdrawRequest      int64     `json:"withdrawRequest" bson:"withdraw_request"`            // 提现请求人数
	WithdrawRequestOrder int64     `json:"withdrawRequestOrder" bson:"withdraw_request_order"` // 提现请求订单数
	WithdrawSuccessOrder int64     `json:"withdrawSuccessOrder" bson:"withdraw_success_order"` // 提现成功订单数
	WithdrawSuccessRate  float64   `bson:"-"`                                                  // 提现成功率
	TotalWithdrawNum     int64     `json:"totalWithdrawNum" bson:"total_withdraw_num"`         // 提现总人数
	TotalWithdrawAmount  int64     `json:"totalWithdrawAmount" bson:"total_withdraw_amount"`   // 提现总金额
	HandlingCharge       int64     `json:"handlingCharge" bson:"handling_charge"`              // 手续费
	CostRatio            float64   `bson:"-"`                                                  // 成本比=提现总金额÷总充值金额
	FNewRechargeAmount   float64   `bson:"-"`                                                  // 新用户充值金额
	FOldRechargeAmount   float64   `bson:"-"`                                                  // 新用户充值金额
	FTotalAmount         float64   `bson:"-"`                                                  // 总充值金额
	FTotalWithdrawAmount float64   `bson:"-"`                                                  // 提现总金额
	FHandlingCharge      float64   `bson:"-"`                                                  // 手续费
	OldRechargeNum2      int64     `json:"oldRechargeNum2" bson:"old_recharge_num2"`           // 老用户充值2次及以上人数
	NewRechargeNum2      int64     `json:"newRechargeNum2" bson:"new_recharge_num2"`           // 新用户充值2次及以上人数
	TotalRechargeNum2    int64     `json:"totalRechargeNum2" bson:"total_recharge_num2"`       // 总充值2次及以上人数
	OldPaySuccessOrder   int64     `json:"oldPaySuccessOrder" bson:"old_pay_success_order"`    // 老用户充值成功订单数
	NewPaySuccessOrder   int64     `json:"newPaySuccessOrder" bson:"new_pay_success_order"`    // 新用户充值成功订单数
	OldRepayRate         float64   `bson:"-"`                                                  // 老用户复购率
	NewRepayRate         float64   `bson:"-"`                                                  // 新用户复购率
	TotalRepayRate       float64   `bson:"-"`                                                  // 总复购率
	OldPayRate           float64   `bson:"-"`                                                  // 老用户人均付费次数
	NewPayRate           float64   `bson:"-"`                                                  // 新用户人均付费次数
	TotalPayRate         float64   `bson:"-"`                                                  // 总人均付费次数
	ChannelZCRate        float64   `bson:"-"`                                                  // 渠道总充额占比 = 该渠道当日总充值 / 当日所有渠道总充值
	ChannelNewPayRate    float64   `bson:"-"`                                                  // 渠道新充额占比 = 该渠道当日新顾客总充值 / 当日所有渠道新顾客总充值
	ChannelOldPayRate    float64   `bson:"-"`                                                  // 渠道老充额占比 = 该渠道当日老顾客总充值 / 当日所有渠道老用户总充值
	ChannelRHRate        float64   `bson:"-"`                                                  // 渠道日活占比 = 该渠道日活人数 / 所有渠道日活总人数(日活人数是登录人数)
	//新增字段
	PayLoginNumber      int64   `json:"payLoginNumber" bson:"pay_login_number"` // 付费用户登录人数
	TotalProfit         int64   `bson:"-"`                                      // 总利润=总充值金额-总提现金额-手续费
	DynamicProfitAvg    float64 `bson:"-"`                                      // 活跃人均利润=总利润÷登录人数
	PayDynamicProfitAvg float64 `bson:"-"`                                      // 活付人均利润=总利润÷付费用户登录人数
	PayMinutes30        int64   `json:"payMinutes30" bson:"pay_minutes30"`      // 注册30分钟内付费人数
	PayMinutes60        int64   `json:"payMinutes60" bson:"pay_minutes60"`      // 注册60分钟内付费人数
	PayMinutes120       int64   `json:"payMinutes120" bson:"pay_minutes120"`    // 注册120分钟内付费人数
	PayMinutesRate30    float64 `bson:"-"`                                      // 30分钟付费率=30分钟内付费人数 /当天注册人数
	PayMinutesRate60    float64 `bson:"-"`                                      // 60分钟付费率=60分钟内付费人数 /当天注册人数
	PayMinutesRate120   float64 `bson:"-"`                                      // 120分钟付费率=120分钟内付费人数 /当天注册人数
	FTotalProfit        float64 `bson:"-"`
	PayLoginNumberRate  float64 `bson:"-"`                                            // 显示百分比：付费用户登录人数/登录人数
	PayRequestRate      float64 `bson:"-"`                                            // 显示百分比：充值请求人数÷登录人数
	PaySuccessOrderRate float64 `bson:"-"`                                            // 百分比显示：充值成功单数÷充值请求单数
	TrustUserPay        int64   `json:"trustUserPay" bson:"trust_user_pay"`           // 信任用户充值
	TrustUserWithdraw   int64   `json:"trustUserWithdraw" bson:"trust_user_withdraw"` // 信任用户提现
	FTrustUserPay       float64 `bson:"-"`
	FTrustUserWithdraw  float64 `bson:"-"`
	TrustRate           float64 `bson:"-"`
}

type DateTotalInfo struct {
	Date        int64 `bson:"-"`
	NewAmount   int64 `bson:"-"`
	OldAmount   int64 `bson:"-"`
	TotalAmount int64 `bson:"-"`
	LoginNum    int64 `bson:"-"`
}

// 用户资源
type UserResource struct {
	Id         string    `bson:"_id"`         // 主键id
	Date       int64     `bson:"date"`        // 统计日期时间戳
	SDate      time.Time `bson:"-"`           // 统计日期
	UType      int       `bson:"u_type"`      // 用户类型 0：全部用户;1：活跃用户
	UserNumber int64     `bson:"user_number"` // 用户数
	Diamond    float64   `bson:"diamond"`     // 钻石(彩金cash)
	Coin       float64   `bson:"coin"`        //金币(奖励金bonus)
}

// 资源流动 -> 全局
type GlobalResource struct {
	Id          string    `bson:"_id"`                             // 主键id
	Date        int64     `bson:"date"`                            // 统计日期时间戳
	SDate       time.Time `bson:"-"`                               // 统计日期
	LoginNum    int64     `json:"loginNum" bson:"login_num"`       // 登录人数
	NewRegister int64     `json:"newRegister" bson:"new_register"` // 新增人数
	// TotalRecharge  int64     `json:"totalRecharge" bson:"total_recharge"`   // 总充值人数
	// TotalAmount    float64   `json:"totalAmount" bson:"total_amount"`       // 总充值金额
	// WithdrawNum    int64     `json:"withdrawNum" bson:"withdraw_num"`       // 总提现人数
	// WithdrawAmount float64   `json:"withdrawAmount" bson:"withdraw_amount"` // 提现总金额
	PutDiamond    float64 `json:"putDiamond" bson:"put_diamond"`       // 游戏产出彩金
	PutCoin       float64 `json:"putCoin" bson:"put_coin"`             // 游戏产出奖励金
	ExpendDiamond float64 `json:"expendDiamond" bson:"expend_diamond"` // 游戏消耗彩金
	ExpendCoin    float64 `json:"expendCoin" bson:"expend_coin"`       // 游戏消耗奖励金
	GiftDiamond   float64 `json:"giftDiamond" bson:"gift_diamond"`     // 新用户赠送彩金
	GiftCoin      float64 `json:"giftCoin" bson:"gift_coin"`           // 新用户赠送奖励金
	ExpendBonus   float64 `json:"expendBonus" bson:"expend_bonus"`     // 消耗BONUS
	PutBonus      float64 `json:"putBonus" bson:"put_bonus"`           // 产出BOUNS
	// PayPutDiamond float64 `json:"payPutDiamond" bson:"pay_put_diamond"` // 充值投放彩金
	// PayPutCoin    float64 `json:"payPutCoin" bson:"pay_put_coin"`       // 充值投放奖励金
}

// 资源流动 -> 游戏
type GameResource struct {
	Id                 string    `bson:"_id"`                                            // 主键id
	Date               int64     `bson:"date"`                                           // 统计日期时间戳
	SDate              time.Time `bson:"-"`                                              // 统计日期
	TpNumber           int64     `json:"tpNumber" bson:"tp_number"`                      // TP参与人数
	TpDiamond          float64   `json:"tpDiamond" bson:"tp_diamond"`                    // TP产出彩金
	TpCoin             float64   `json:"tpCoin" bson:"tp_coin"`                          // TP产出奖励金
	TpExpendDiamond    float64   `json:"tpExpendDiamond" bson:"tp_expend_diamond"`       // TP消耗彩金
	TpExpendCoin       float64   `json:"tpExpendCoin" bson:"tp_expend_coin"`             // TP消耗奖励金
	AkNumber           int64     `json:"akNumber" bson:"ak_number"`                      // AK47参与人数
	AkDiamond          float64   `json:"akDiamond" bson:"ak_diamond"`                    // AK47产出彩金
	AkCoin             float64   `json:"akCoin" bson:"ak_coin"`                          // AK47产出奖励金
	AkExpendDiamond    float64   `json:"akExpendDiamond" bson:"ak_expend_diamond"`       // AK47消耗彩金
	AkExpendCoin       float64   `json:"akExpendCoin" bson:"ak_expend_coin"`             // AK47消耗奖励金
	JokerNumber        int64     `json:"jokerNumber" bson:"joker_number"`                // Joker参与人数
	JokerDiamond       float64   `json:"jokerDiamond" bson:"joker_diamond"`              // Joker产出彩金
	JokerCoin          float64   `json:"jokerCoin" bson:"joker_coin"`                    // Joker产出奖励金
	JokerExpendDiamond float64   `json:"jokerExpendDiamond" bson:"joker_expend_diamond"` // Joker消耗彩金
	JokerExpendCoin    float64   `json:"jokerExpendCoin" bson:"joker_expend_coin"`       // Joker消耗奖励金
	RmNumber           int64     `json:"rmNumber" bson:"rm_number"`                      // Rummy参与人数
	RmDiamond          float64   `json:"rmDiamond" bson:"rm_diamond"`                    // Rummy产出彩金
	RmCoin             float64   `json:"rmCoin" bson:"rm_coin"`                          // Rummy产出奖励金
	RmExpendDiamond    float64   `json:"rmExpendDiamond" bson:"rm_expend_diamond"`       // Rummy消耗彩金
	RmExpendCoin       float64   `json:"rmExpendCoin" bson:"rm_expend_coin"`             // Rummy消耗奖励金
	LhdNumber          int64     `json:"lhdNumber" bson:"lhd_number"`                    // 龙虎斗参与人数
	LhdDiamond         float64   `json:"lhdDiamond" bson:"lhd_diamond"`                  // 龙虎斗产出彩金
	LhdCoin            float64   `json:"lhdCoin" bson:"lhd_coin"`                        // 龙虎斗产出奖励金
	LhdExpendDiamond   float64   `json:"lhdExpendDiamond" bson:"lhd_expend_diamond"`     // 龙虎斗消耗彩金
	LhdExpendCoin      float64   `json:"lhdExpendCoin" bson:"lhd_expend_coin"`           // 龙虎斗消耗奖励金
	UpNumber           int64     `json:"upNumber" bson:"up_number"`                      // 7updown参与人数
	UpDiamond          float64   `json:"upDiamond" bson:"up_diamond"`                    // 7updown产出彩金
	UpCoin             float64   `json:"upCoin" bson:"up_coin"`                          // 7updown产出奖励金
	UpExpendDiamond    float64   `json:"upExpendDiamond" bson:"up_expend_diamond"`       // 7updown消耗彩金
	UpExpendCoin       float64   `json:"upExpendCoin" bson:"up_expend_coin"`             // 7updown消耗奖励金
	CRASHNumber        int64     `json:"crashNumber" bson:"crash_number"`                // CRASH参与人数
	CRASHDiamond       float64   `json:"crashDiamond" bson:"crash_diamond"`              // CRASH产出彩金
	CRASHCoin          float64   `json:"crashCoin" bson:"crash_coin"`                    // CRASH产出奖励金
	CRASHExpendDiamond float64   `json:"crashExpendDiamond" bson:"crash_expend_diamond"` // CRASH消耗彩金
	CRASHExpendCoin    float64   `json:"crashExpendCoin" bson:"crash_expend_coin"`       // CRASH消耗奖励金
}

// 实时数据
type RealTimeData struct {
	Id             string    `bson:"_id"`             // 主键id
	Date           time.Time `bson:"date"`            // 统计日期
	SDate          string    `bson:"-"`               // 统计日期
	LoginNumber    int64     `bson:"login_number"`    // 登录人数
	OnlineNumber   int64     `bson:"online_number"`   // 在线人数
	PayNumber      int64     `bson:"pay_number"`      // 充值人数
	PayAmount      float64   `bson:"pay_amount"`      // 充值金额
	WithdrawNumber int64     `bson:"withdraw_number"` // 提现人数
	WithdrawAmount float64   `bson:"withdraw_amount"` // 提现金额
}

// 渠道成功率
type ChannelSuccessRate struct {
	Id                     string    `bson:"_id"`                                                    // 订单id
	Date                   int64     `bson:"date"`                                                   // 统计日期时间戳
	SDate                  time.Time `bson:"-"`                                                      // 统计日期
	PackageId              string    `json:"packageId" bson:"package_id"`                            // 渠道
	PackageName            string    `json:"packageName" bson:"package_name"`                        // 渠道别名
	Channel                int64     `json:"channel" bson:"channel"`                                 // 支付渠道
	ChannelName            string    `bson:"-"`                                                      // 支付渠道名称
	PayRequest             int64     `json:"payRequest" bson:"pay_request"`                          // 充值请求人数
	LoginNumber            int64     `json:"loginNumber" bson:"login_number"`                        // 今日登录人数
	PullRate               float64   `bson:"-"`                                                      // 拉单用户比例
	PayRequestOrder        int64     `json:"payRequestOrder" bson:"pay_request_order"`               // 充值请求订单数
	PaySuccessOrder        int64     `json:"paySuccessOrder" bson:"pay_success_order"`               // 充值成功订单数
	PaySuccessRate         float64   `json:"paySuccessRate" bson:"pay_success_rate"`                 // 充值成功率
	PaySuccessMoney        float64   `json:"paySuccessMoney" bson:"pay_success_money"`               // 充值成功金额
	PayFailMoney           float64   `json:"payFailMoney" bson:"pay_fail_money"`                     // 充值失败金额
	WithdrawRequest        int64     `json:"withdrawRequest" bson:"withdraw_request"`                // 提现请求人数
	WithdrawRequestOrder   int64     `json:"withdrawRequestOrder" bson:"withdraw_request_order"`     // 提现请求订单数
	WithdrawSuccessOrder   int64     `json:"withdrawSuccessOrder" bson:"withdraw_success_order"`     // 提现成功订单数
	WithdrawSuccessRate    float64   `json:"withdrawSuccessRate" bson:"withdraw_success_rate"`       // 提现成功率
	WithdrawSuccessMoney   float64   `json:"withdrawSuccessMoney" bson:"withdraw_success_money"`     // 提现成功金额
	WithdrawFailMoney      float64   `json:"withdrawFailMoney" bson:"withdraw_fail_money"`           // 提现失败金额
	ThirdpartyOrder        int64     `json:"thirdpartyOrder" bson:"thirdparty_order"`                // 第三方提交订单数
	ThirdpartySuccessOrder int64     `json:"thirdpartySuccessOrder" bson:"thirdparty_success_order"` // 第三方提交请求成功订单数
	ThirdpartyOrderRate    float64   `json:"thirdpartyOrderRate" bson:"thirdparty_order_rate"`       // 第三方提交订单成功率
}

type PaySuccessRate struct {
	Date             int64
	XdPayRate        float64
	MlPayRate        float64
	Pay1916Rate      float64
	KingPayRate      float64
	XfPayRate        float64
	SailsPayRate     float64
	FlyPayRate       float64
	UwinPayRate      float64
	LetsPayRate      float64
	RamaPayRate      float64
	IcePayRate       float64
	WePayRate        float64
	OePayRate        float64
	Pay9sRate        float64
	BlizzardpyRate   float64
	MetagopayRate    float64
	UniversalpayRate float64
	USDTRate         float64
}

// 埋点数据
type PointData struct {
	Id                   string         `bson:"_id"`                                      // 订单id
	Date                 int64          `bson:"date"`                                     // 统计日期时间戳
	SDate                time.Time      `bson:"-"`                                        // 统计日期
	Channel              string         `json:"channel" bson:"channel"`                   // 渠道
	Channel1             string         `json:"channel1" bson:"channel1"`                 // 渠道别名
	TotalRegister        int64          `json:"totalRegister" bson:"total_register"`      // 总注册
	TotalTourist         int64          `json:"totalTourist" bson:"total_tourist"`        // 游客注册
	TotalTouristRatio    float64        `bson:"-"`                                        // 游客注册占比
	TotalMobile          int64          `json:"totalMobile" bson:"total_mobile"`          // 手机注册
	TotalMobileRatio     float64        `bson:"-"`                                        // 手机注册占比
	GuidanceBinding      int64          `json:"guidanceBinding" bson:"guidance_binding"`  // 引导绑定手机人数
	GuidanceGetGold      int64          `json:"guidanceGetGold" bson:"guidance_get_gold"` // 引导领取金币人数
	GuidanceGetGoldRatio float64        `bson:"-"`                                        // 引导领取金币人数占比
	TPNumber1            int64          `json:"tpNumber1" bson:"tp_number1"`              // 玩第一局TP人数
	TPNumber1Ratio       float64        `bson:"-"`                                        // 玩第一局TP人数占比
	TPNumber2            int64          `json:"tpNumber2" bson:"tp_number2"`              // 玩第二局TP人数
	TPNumber2Ratio       float64        `bson:"-"`                                        // 玩第二局TP人数占比
	TPExitManually       int64          `json:"tpExitManually" bson:"tp_exit_manually"`   // 手动退出TP人数
	TPExitManuallyRatio  float64        `bson:"-"`                                        // 手动退出TP人数占比
	Gold100              int64          `json:"gold100" bson:"gold100"`                   // 金币≥10500   ＜20000的人数
	Gold100Ratio         float64        `bson:"-"`                                        // 金币≥10500   ＜20000的人数占比
	Gold200              int64          `json:"gold200" bson:"gold200"`                   // 金币>20000以上的人数
	Gold200Ratio         float64        `bson:"-"`                                        // 金币>20000以上的人数占比
	FirstGames           *PlayersNumber `json:"firstGames" bson:"first_games"`            // 第一局玩的游戏局数统计
	SecondGames          *PlayersNumber `json:"secondGames" bson:"second_games"`          // 第一局玩的游戏局数统计
	// ExitRate
}

type PlayersNumber struct {
	TpNumber    int64 `json:"tpNumber" bson:"tp_number"`
	RmNumber    int64 `json:"rmNumber" bson:"rm_number"`
	AkNumber    int64 `json:"akNumber" bson:"ak_number"`
	JokerNumber int64 `json:"jokerNumber" bson:"joker_number"`
	LHDNumber   int64 `json:"lhdNumber" bson:"lhd_number"`
	UPNumber    int64 `json:"upNumber" bson:"up_number"`
	CrashNumber int64 `json:"crashNumber" bson:"crash_number"`
}

type LogEventTrack struct {
	Id     string `json:"_id" bson:"_id"`
	Typ    int32  `json:"typ" bson:"typ"`
	Userid string `json:"userid" bson:"userid"`
	Gtype  int32  `json:"gtype" bson:"gtype"`
	Ctime  int64  `json:"ctime" bson:"ctime"`
}

// 局数分析
type GameNumberAnalysis struct {
	Id             string    `bson:"_id"`                                 // 订单id
	Date           int64     `bson:"date"`                                // 统计日期时间戳
	SDate          time.Time `bson:"-"`                                   // 统计日期
	Channel        string    `json:"channel" bson:"channel"`              // 渠道
	Channel1       string    `json:"channel1" bson:"channel1"`            // 渠道别名
	TotalRegister  int64     `json:"totalRegister" bson:"total_register"` // 总注册
	GNumber        int64     `json:"gNumber" bson:"g_number"`             // 0局
	GNumberRatio   float64   `json:"gNumberRatio" bson:"-"`               // 0局数占比
	GNumber1       int64     `json:"gNumber1" bson:"g_number1"`           // 1局
	GNumberRatio1  float64   `json:"gNumberRatio1" bson:"-"`              // 1局数占比
	GNumber2       int64     `json:"gNumber2" bson:"g_number2"`           // 2局
	GNumberRatio2  float64   `json:"gNumberRatio2" bson:"-"`              // 2局数占比
	GNumber3       int64     `json:"gNumber3" bson:"g_number3"`           // 3局
	GNumberRatio3  float64   `json:"gNumberRatio3" bson:"-"`              // 3局数占比
	GNumber4       int64     `json:"gNumber4" bson:"g_number4"`           // 4局
	GNumberRatio4  float64   `json:"gNumberRatio4" bson:"-"`              // 4局数占比
	GNumber5       int64     `json:"gNumber5" bson:"g_number5"`           // 5局
	GNumberRatio5  float64   `json:"gNumberRatio5" bson:"-"`              // 5局数占比
	GNumber6       int64     `json:"gNumber6" bson:"g_number6"`           // 6-10局
	GNumberRatio6  float64   `json:"gNumberRatio6" bson:"-"`              // 6-10局数占比
	GNumber11      int64     `json:"gNumber11" bson:"g_number11"`         // 11-20局
	GNumberRatio11 float64   `json:"gNumberRatio11" bson:"-"`             // 11-20局数占比
	GNumber21      int64     `json:"gNumber21" bson:"g_number21"`         // 21-30局
	GNumberRatio21 float64   `json:"gNumberRatio21" bson:"-"`             // 21-30局数占比
	GNumber31      int64     `json:"gNumber31" bson:"g_number31"`         // 31局以上
	GNumberRatio31 float64   `json:"gNumberRatio31" bson:"-"`             // 31局以上局数占比
}

// Bug统计
type BugStatistics struct {
	Id       string    `bson:"_id"`                       // id
	Date     int64     `bson:"date"`                      // 统计日期时间戳
	SDate    time.Time `bson:"-"`                         // 统计日期
	BugType1 int64     `json:"bugType1" bson:"bug_type1"` // Bug类型1领取次数
	BugType2 int64     `json:"bugType2" bson:"bug_type2"` // Bug类型2领取次数
	BugType3 int64     `json:"bugType3" bson:"bug_type3"` // Bug类型3领取次数
	BugType4 int64     `json:"bugType4" bson:"bug_type4"` // Bug类型4领取次数
	BugType5 int64     `json:"bugType5" bson:"bug_type5"` // Bug类型5领取次数
}

// // Bug提交日志
// type BugCommitLogs struct {
// 	Id        string `bson:"_id"`                         // id
// 	Date      int64  `bson:"date"`                        // 统计日期时间戳
// 	UserId    string `bson:"userid"`                      // 用户ID
// 	BugType   int64  `json:"bugType" bson:"bug_type"`     // bug类型
// 	DaySubmit int64  `json:"daySubmit" bson:"day_submit"` // 当日提交次数
// }

// bug日志
type BugFeedbackLog struct {
	Id        string    `json:"id" bson:"_id"`         // id
	UserId    string    `json:"userId" bson:"user_id"` // 用户id
	BugId     []int32   `json:"bugId" bson:"bug_id"`   // bugid
	Ctime     int64     `json:"ctime" bson:"ctime"`    // 提交时间
	SDate     time.Time `bson:"-"`                     // 统计日期
	DaySubmit int64     `bson:"-"`                     // 当日提交次数
}

// 时间分析
type PlaytimeAnalysis struct {
	Id              string    `bson:"_id"`                                 // id
	Date            int64     `bson:"date"`                                // 统计日期时间戳
	SDate           time.Time `bson:"-"`                                   // 统计日期
	Channel         string    `json:"channel" bson:"channel"`              // 渠道
	Channel1        string    `json:"channel1" bson:"channel1"`            // 渠道别名
	TotalRegister   int64     `json:"totalRegister" bson:"total_register"` // 总注册
	Playtime1       int64     `json:"playtime1" bson:"playtime1"`          // 游戏时间≤1分钟
	PlaytimeRatio1  float64   `json:"playtimeRatio1" bson:"-"`
	Playtime2       int64     `json:"playtime2" bson:"playtime2"` // 游戏时间在2-5分钟以内
	PlaytimeRatio2  float64   `json:"playtimeRatio2" bson:"-"`
	Playtime6       int64     `json:"playtime6" bson:"playtime6"` // 游戏时间在6-10分钟以内
	PlaytimeRatio6  float64   `json:"playtimeRatio6" bson:"-"`
	Playtime11      int64     `json:"playtime11" bson:"playtime11"` // 游戏时间在11-20分钟以内
	PlaytimeRatio11 float64   `json:"playtimeRatio11" bson:"-"`
	Playtime21      int64     `json:"playtime21" bson:"playtime21"` // 游戏时间在21-30分钟以内
	PlaytimeRatio21 float64   `json:"playtimeRatio21" bson:"-"`
	Playtime31      int64     `json:"playtime31" bson:"playtime31"` // 游戏时间在≥30分钟以上
	PlaytimeRatio31 float64   `json:"playtimeRatio31" bson:"-"`
}

type LogGameTime struct {
	Id     string `json:"_id" bson:"_id"`
	Userid string `json:"userid" bson:"userid"`
	Gtype  int32  `json:"gtype" bson:"gtype"`
	Time   int64  `json:"time" bson:"time"`
	Ctime  int64  `json:"ctime" bson:"ctime"`
}

type StrategyData struct {
	Id            string `bson:"_id"` // id
	Number        int    `json:"number" bson:"number"`
	GoldNumber    int    `json:"goldNumber" bson:"gold_number"`
	Number100     int    `json:"number100" bson:"number100"`
	GoldNumber100 int    `json:"goldNumber100" bson:"gold_number100"`
	Number200     int    `json:"number200" bson:"number200"`
	GoldNumber200 int    `json:"goldNumber200" bson:"gold_number200"`
}

/*
房间数据
*/
type RoomData struct {
	Id                  string    `bson:"_id"`                                         // id
	Date                int64     `bson:"date"`                                        // 统计日期时间戳
	SDate               time.Time `bson:"-"`                                           // 统计日期
	Channel             string    `json:"channel" bson:"channel"`                      // 渠道
	Channel1            string    `json:"channel1" bson:"channel1"`                    // 渠道别名
	PlayerTypes         int64     `json:"playerTypes" bson:"player_types"`             // 玩家类型 1.新玩家 2.老玩家
	NumberTypes         int       `json:"numberTypes" bson:"number_types"`             // 局数类型
	PlayerTotal         int64     `json:"playerTotal" bson:"player_total"`             // 玩家总数
	TPGameNumber        int64     `json:"tpGameNumber" bson:"tp_game_number"`          // tp游戏局数
	TPGameRealNumber    int64     `json:"tpGameRealNumber" bson:"tp_game_real_number"` // TP游戏真人局数
	TPBattleGame        int64     `json:"tpBattleGame" bson:"tp_battle_game"`          // TP对战房局数
	TPNumber            int64     `json:"tpNumber" bson:"tp_number"`                   // tp玩家人数
	TPGameTime          int64     `json:"tpGameTime" bson:"tp_game_time"`              // tp游戏时长
	RummyGameNumber     int64     `json:"rummyGameNumber" bson:"rummy_game_number"`
	RMGameRealNumber    int64     `json:"rmGameRealNumber" bson:"rm_game_real_number"` // RM游戏真人局数
	RMBattleGame        int64     `json:"rmBattleGame" bson:"rm_battle_game"`          // RM对战房局数
	RummyNumber         int64     `json:"rummyNumber" bson:"rummy_number"`
	RummyGameTime       int64     `json:"rummyGameTime" bson:"rummy_game_time"`
	LHDGameNumber       int64     `json:"lhdGameNumber" bson:"lhd_game_number"`
	LHDNumber           int64     `json:"lhdNumber" bson:"lhd_number"`
	LHDGameTime         int64     `json:"lhdGameTime" bson:"lhd_game_time"`
	UPGameNumber        int64     `json:"upGameNumber" bson:"up_game_number"`
	UPNumber            int64     `json:"upNumber" bson:"up_number"`
	UPGameTime          int64     `json:"upGameTime" bson:"up_game_time"`
	AKGameNumber        int64     `json:"akGameNumber" bson:"ak_game_number"`
	AKGameRealNumber    int64     `json:"akGameRealNumber" bson:"ak_game_real_number"` // AK游戏真人局数
	AKNumber            int64     `json:"akNumber" bson:"ak_number"`
	AKGameTime          int64     `json:"akGameTime" bson:"ak_game_time"`
	JokerGameNumber     int64     `json:"jokerGameNumber" bson:"joker_game_number"`
	JokerGameRealNumber int64     `json:"jokerGameRealNumber" bson:"joker_game_real_number"` // Joker游戏真人局数
	JokerNumber         int64     `json:"jokerNumber" bson:"joker_number"`
	JokerGameTime       int64     `json:"jokerGameTime" bson:"joker_game_time"`
	CrashGameNumber     int64     `json:"crashGameNumber" bson:"crash_game_number"`
	CrashNumber         int64     `json:"crashNumber" bson:"crash_number"`
	CrashGameTime       int64     `json:"crashGameTime" bson:"crash_game_time"`
	ABGameNumber        int64     `json:"abGameNumber" bson:"ab_game_number"`
	ABBattleGame        int64     `json:"abBattleGame" bson:"ab_battle_game"` // AB对战房局数
	ABNumber            int64     `json:"abNumber" bson:"ab_number"`
	ABGameTime          int64     `json:"abGameTime" bson:"ab_game_time"`
	CPGameNumber        int64     `json:"cpGameNumber" bson:"cp_game_number"`
	CPNumber            int64     `json:"cpNumber" bson:"cp_number"`
	CPGameTime          int64     `json:"cpGameTime" bson:"cp_game_time"`
	FJGameNumber        int64     `json:"fjGameNumber" bson:"fj_game_number"`
	FJNumber            int64     `json:"fjNumber" bson:"fj_number"`
	FJGameTime          int64     `json:"fjGameTime" bson:"fj_game_time"`
	RMTwoGameNumber     int64     `json:"rmTwoGameNumber" bson:"rm_two_game_number"`
	RMTwoNumber         int64     `json:"rmTwoNumber" bson:"rm_two_number"`
	RMTwoGameTime       int64     `json:"rmTwoGameTime" bson:"rm_two_game_time"`
	SlotsGameNumber     int64     `json:"slotsGameNumber" bson:"slots_game_number"`
	SlotsNumber         int64     `json:"slotsNumber" bson:"slots_number"`
	SlotsGameTime       int64     `json:"slotsGameTime" bson:"slots_game_time"`
	ZRSXGameNumber      int64     `json:"zrsxGameNumber" bson:"zrsx_game_number"`
	ZRSXNumber          int64     `json:"zrsxNumber" bson:"zrsx_number"`
}

// 房间数据统计列表展示数据
type RoomDataList struct {
	Id                  int64     `bson:"_id"`                                         // id
	Date                int64     `bson:"date"`                                        // 统计日期时间戳
	SDate               time.Time `bson:"-"`                                           // 统计日期
	Channel             string    `json:"channel" bson:"channel"`                      // 渠道
	Channel1            string    `json:"channel1" bson:"channel1"`                    // 渠道别名
	PlayerTypes         int64     `json:"playerTypes" bson:"player_types"`             // 玩家类型 1.新玩家 2.老玩家
	NumberTypes         int       `json:"numberTypes" bson:"number_types"`             // 局数类型
	PlayerTotal         int64     `json:"playerTotal" bson:"player_total"`             // 玩家总数
	TPGameNumber        int64     `json:"tpGameNumber" bson:"tp_game_number"`          // tp游戏局数
	TPGameRealNumber    int64     `json:"tpGameRealNumber" bson:"tp_game_real_number"` // TP游戏真人局数
	TPBattleGame        int64     `json:"tpBattleGame" bson:"tp_battle_game"`          // TP对战房局数
	TPNumber            int64     `json:"tpNumber" bson:"tp_number"`                   // tp玩家人数
	TPGameTime          int64     `json:"tpGameTime" bson:"tp_game_time"`              // tp游戏时长
	TPPerCapita         float64   `json:"tpPerCapita" bson:"-"`                        // tp人均时长
	TPTime              float64   `json:"tpTime" bson:"-"`                             // tp单局时长
	RummyGameNumber     int64     `json:"rummyGameNumber" bson:"rummy_game_number"`
	RMGameRealNumber    int64     `json:"rmGameRealNumber" bson:"rm_game_real_number"` // RM游戏真人局数
	RMBattleGame        int64     `json:"rmBattleGame" bson:"rm_battle_game"`          // RM对战房局数
	RummyNumber         int64     `json:"rummyNumber" bson:"rummy_number"`
	RummyGameTime       int64     `json:"rummyGameTime" bson:"rummy_game_time"`
	RummyPerCapita      float64   `json:"rummyPerCapita" bson:"-"`
	RummyTime           float64   `json:"rummyTime" bson:"-"`
	LHDGameNumber       int64     `json:"lhdGameNumber" bson:"lhd_game_number"`
	LHDNumber           int64     `json:"lhdNumber" bson:"lhd_number"`
	LHDGameTime         int64     `json:"lhdGameTime" bson:"lhd_game_time"`
	LHDPerCapita        float64   `json:"lhdPerCapita" bson:"-"`
	LHDTime             float64   `json:"lhdTime" bson:"-"`
	UPGameNumber        int64     `json:"upGameNumber" bson:"up_game_number"`
	UPNumber            int64     `json:"upNumber" bson:"up_number"`
	UPGameTime          int64     `json:"upGameTime" bson:"up_game_time"`
	UPPerCapita         float64   `json:"upPerCapita" bson:"-"`
	UPTime              float64   `json:"upTime" bson:"-"`
	AKGameNumber        int64     `json:"akGameNumber" bson:"ak_game_number"`
	AKGameRealNumber    int64     `json:"akGameRealNumber" bson:"ak_game_real_number"` // AK游戏真人局数
	AKNumber            int64     `json:"akNumber" bson:"ak_number"`
	AKGameTime          int64     `json:"akGameTime" bson:"ak_game_time"`
	AKPerCapita         float64   `json:"akPerCapita" bson:"-"`
	AKTime              float64   `json:"akTime" bson:"-"`
	JokerGameNumber     int64     `json:"jokerGameNumber" bson:"joker_game_number"`
	JokerGameRealNumber int64     `json:"jokerGameRealNumber" bson:"joker_game_real_number"` // Joker游戏真人局数
	JokerNumber         int64     `json:"jokerNumber" bson:"joker_number"`
	JokerGameTime       int64     `json:"jokerGameTime" bson:"joker_game_time"`
	JokerPerCapita      float64   `json:"jokerPerCapita" bson:"-"`
	JokerTime           float64   `json:"jokerTime" bson:"-"`
	CrashGameNumber     int64     `json:"crashGameNumber" bson:"crash_game_number"`
	CrashNumber         int64     `json:"crashNumber" bson:"crash_number"`
	CrashGameTime       int64     `json:"crashGameTime" bson:"crash_game_time"`
	CrashPerCapita      float64   `json:"crashPerCapita" bson:"-"`
	CrashTime           float64   `json:"crashTime" bson:"-"`
	ABGameNumber        int64     `json:"abGameNumber" bson:"ab_game_number"`
	ABBattleGame        int64     `json:"abBattleGame" bson:"ab_battle_game"` // AB对战房局数
	ABNumber            int64     `json:"abNumber" bson:"ab_number"`
	ABGameTime          int64     `json:"abGameTime" bson:"ab_game_time"`
	ABPerCapita         float64   `bson:"-"`
	ABTime              float64   `bson:"-"`
	CPGameNumber        int64     `json:"cpGameNumber" bson:"cp_game_number"`
	CPNumber            int64     `json:"cpNumber" bson:"cp_number"`
	CPGameTime          int64     `json:"cpGameTime" bson:"cp_game_time"`
	CPPerCapita         float64   `bson:"-"`
	CPTime              float64   `bson:"-"`
	FJGameNumber        int64     `json:"fjGameNumber" bson:"fj_game_number"`
	FJNumber            int64     `json:"fjNumber" bson:"fj_number"`
	FJGameTime          int64     `json:"fjGameTime" bson:"fj_game_time"`
	FJPerCapita         float64   `bson:"-"`
	FJTime              float64   `bson:"-"`
	RBGameNumber        int64     `json:"rbGameNumber" bson:"rb_game_number"`
	RBNumber            int64     `json:"rbNumber" bson:"rb_number"`
	RBGameTime          int64     `json:"rbGameTime" bson:"rb_game_time"`
	RBPerCapita         float64   `bson:"-"`
	RBTime              float64   `bson:"-"`
	RMTwoGameNumber     int64     `json:"rmTwoGameNumber" bson:"rm_two_game_number"`
	RMTwoNumber         int64     `json:"rmTwoNumber" bson:"rm_two_number"`
	RMTwoGameTime       int64     `json:"rmTwoGameTime" bson:"rm_two_game_time"`
	RMTwoPerCapita      float64   `bson:"-"`
	RMTwoTime           float64   `bson:"-"`
	TP2GameNumber       int64     `json:"tp2GameNumber" bson:"tp2_game_number"`
	TP2Number           int64     `json:"tp2Number" bson:"tp2_number"`
	TP2GameTime         int64     `json:"tp2GameTime" bson:"tp2_game_time"`
	TP2PerCapita        float64   `bson:"-"`
	TP2Time             float64   `bson:"-"`
	SlotsGameNumber     int64     `json:"slotsGameNumber" bson:"slots_game_number"`
	SlotsNumber         int64     `json:"slotsNumber" bson:"slots_number"`
	SlotsGameTime       int64     `json:"slotsGameTime" bson:"slots_game_time"`
	SlotsPerCapita      float64   `bson:"-"`
	SlotsTime           float64   `bson:"-"`
	ZRSXGameNumber      int64     `json:"zrsxGameNumber" bson:"zrsx_game_number"`
	ZRSXNumber          int64     `json:"zrsxNumber" bson:"zrsx_number"`
	ZRSXPerCapita       float64   `bson:"-"`
	ZRSXTime            float64   `bson:"-"`
}

// 商品购买
type GoodsBuyData struct {
	Id                   string    `bson:"_id"`                             // id
	Date                 int64     `bson:"date"`                            // 统计日期时间戳
	SDate                time.Time `bson:"-"`                               // 统计日期
	PlayerTypes          int64     `json:"playerTypes" bson:"player_types"` // 玩家类型 1.新玩家 2.老玩家
	ShopType             int32     `json:"shopType" bson:"shop_type"`       // 商品类型
	ShopName             string    `json:"ShopName" bson:"shop_name"`       // 商品名称
	Amount               uint32    `json:"amount" bson:"amount"`            // 金额
	FAmount              float64   `bson:"-"`
	TotalPullOrder       int64     `json:"totalPullOrder" bson:"total_pull_order"`  // 总拉单
	SuccessfulOrder      int64     `json:"successfulOrder" bson:"successful_order"` // 成功订单
	SuccessfulOrderRatio float64   `bson:"-"`                                       // 成功率
	PullNumber           int64     `json:"pullNumber" bson:"pull_number"`           // 拉起人数
	BuyNumber            int64     `json:"buyNumber" bson:"buy_number"`             // 成功购买人数
	BuyNumberRatio       float64   `bson:"-"`                                       // 购买率
	ZBRatio              float64   `bson:"-"`                                       // 占比率
}

// AD上报统计
type AdReportData struct {
	Id                string    `bson:"_id"`                      // 主键id
	Date              int64     `bson:"date"`                     // 统计日期时间戳
	SDate             time.Time `bson:"-"`                        // 统计日期
	Channel           string    `json:"channel" bson:"channel"`   // 渠道
	Channel1          string    `json:"channel1" bson:"channel1"` // 渠道别名
	RegisterCount     int64     `bson:"register_count"`           // 注册条数
	RegisterCountDay  int64     `bson:"register_count_day"`       // 当天上报注册条数（etime和ctime为同一天）
	RegisterCountDay1 int64     `bson:"register_count_day1"`      // 跨天上报注册条数（etime和ctime为不同一天）
	FailRegisterCount int64     `bson:"fail_register_count"`      // 失败注册条数
	LoginCount        int64     `bson:"login_count"`              // 登录条数
	LoginCountDay     int64     `bson:"login_count_day"`          // 当天上报登录条数（etime和ctime为同一天）
	LoginCountDay1    int64     `bson:"login_count_day1"`         // 当天上报登录条数（etime和ctime为同一天）
	FailLoginCount    int64     `bson:"fail_login_count"`         // 失败登录条数
	DepositCount      int64     `bson:"deposit_count"`            // 成功充值条数
	DepositCountDay   int64     `bson:"deposit_count_day"`        // 当天上报充值条数（etime和ctime为同一天）
	DepositCountDay1  int64     `bson:"deposit_count_day1"`       // 跨天上报充值条数（etime和ctime为不同一天）
	FailDepositCount  int64     `bson:"fail_deposit_count"`       // 失败充值条数
}

// 分享数据统计
type ShareDataStatistics struct {
	Id             string    `bson:"_id"`                      // 主键id
	Date           int64     `bson:"date"`                     // 统计日期时间戳
	SDate          time.Time `bson:"-"`                        // 统计日期
	Channel        string    `json:"channel" bson:"channel"`   // 渠道
	Channel1       string    `json:"channel1" bson:"channel1"` // 渠道别名
	RegisterNumber int64     `bson:"register_number"`          // 邀请注册人数
	PayNumber      int64     `bson:"pay_number"`               // 被邀请充值人数
	PayMoney       int64     `bson:"pay_money"`                // 被邀请人充值金额
	FPayMoney      float64   `bson:"-"`                        // 被邀请人充值金额
	ShareArpu      float64   `bson:"-"`
	ShareArppu     float64   `bson:"-"`
}

// 输赢分
type CashFlow struct {
	Id            string  `bson:"_id" json:"id"`               // id
	Gtype         int32   `json:"gtype" bson:"gtype"`          // 游戏类型
	WinScore      int64   `json:"winScore" bson:"win_score"`   // 赢分
	LoseScore     int64   `json:"loseScore" bson:"lose_score"` // 输分
	Date          string  `json:"date" bson:"date"`            // 日期
	Ctime         int64   `json:"ctime" bson:"ctime"`          // 时间
	FWinScore     float64 `bson:"-"`                           // 赢分
	FLoseScore    float64 `bson:"-"`                           // 输分
	ProfitAndLoss float64 `bson:"-"`                           // 平台盈亏
	IsWin         bool    `bson:"-"`                           // 是否正数
}

// 短信统计
type SmsData struct {
	Id           string    `bson:"_id"`            // 主键id
	Date         int64     `bson:"date"`           // 统计日期时间戳
	SDate        time.Time `bson:"-"`              // 统计日期
	TotalSendSMS int64     `bson:"total_send_sms"` // 总发送
	SendCount    int64     `bson:"send_count"`     // 发送成功
	SendSMSRatio float64   `bson:"-"`              // 发送成功率
	UseSMS       int64     `bson:"use_sms"`        // 使用成功
	UseSMSRatio  float64   `bson:"-"`              // 使用成功率
}

// 充值来源
type PaySourceData struct {
	Id           string    `bson:"_id"`                       // 主键id
	Date         int64     `bson:"date"`                      // 统计日期时间戳
	SDate        time.Time `bson:"-"`                         // 统计日期
	TPAmount     int64     `json:"tpAmount" bson:"tp_amount"` // TP充值金额
	TPNumber     int64     `json:"tpNumber" bson:"tp_number"` // TP充值笔数
	FTPAmount    float64   `bson:"-"`                         // TP充值金额
	TPAvg        float64   `bson:"-"`                         // TP单笔均价
	LHDAmount    int64     `json:"lhdAmount" bson:"lhd_amount"`
	LHDNumber    int64     `json:"lhdNumber" bson:"lhd_number"`
	FLHDAmount   float64   `bson:"-"`
	LHDAvg       float64   `bson:"-"`
	RMAmount     int64     `json:"rmAmount" bson:"rm_amount"`
	RMNumber     int64     `json:"rmNumber" bson:"rm_number"`
	FRMAmount    float64   `bson:"-"`
	RMAvg        float64   `bson:"-"`
	UPAmount     int64     `json:"upAmount" bson:"up_amount"`
	UPNumber     int64     `json:"upNumber" bson:"up_number"`
	FUPAmount    float64   `bson:"-"`
	UPAvg        float64   `bson:"-"`
	AKAmount     int64     `json:"akAmount" bson:"ak_amount"`
	AKNumber     int64     `json:"akNumber" bson:"ak_number"`
	FAKAmount    float64   `bson:"-"`
	AKAvg        float64   `bson:"-"`
	JOKERAmount  int64     `json:"jokerAmount" bson:"joker_amount"`
	JOKERNumber  int64     `json:"jokerNumber" bson:"joker_number"`
	FJOKERAmount float64   `bson:"-"`
	JOKERAvg     float64   `bson:"-"`
	CRASHAmount  int64     `json:"crashAmount" bson:"crash_amount"`
	CRASHNumber  int64     `json:"crashNumber" bson:"crash_number"`
	FCRASHAmount float64   `bson:"-"`
	CRASHAvg     float64   `bson:"-"`
	ABAmount     int64     `json:"abAmount" bson:"ab_amount"`
	ABNumber     int64     `json:"abNumber" bson:"ab_number"`
	FABAmount    float64   `bson:"-"`
	ABAvg        float64   `bson:"-"`
	CPAmount     int64     `json:"cpAmount" bson:"cp_amount"`
	CPNumber     int64     `json:"cpNumber" bson:"cp_number"`
	FCPAmount    float64   `bson:"-"`
	CPAvg        float64   `bson:"-"`
	FJAmount     int64     `json:"fjAmount" bson:"fj_amount"`
	FJNumber     int64     `json:"fjNumber" bson:"fj_number"`
	FFJAmount    float64   `bson:"-"`
	FJAvg        float64   `bson:"-"`
	RMTwoAmount  int64     `json:"rmTwoAmount" bson:"rm_two_amount"`
	RMTwoNumber  int64     `json:"rmTwoNumber" bson:"rm_two_number"`
	FRMTwoAmount float64   `bson:"-"`
	RMTwoAvg     float64   `bson:"-"`
}

// 对战房埋点记录
type BattleRoomData struct {
	Id                 string    `bson:"_id"`                                            // 主键id
	Date               int64     `bson:"date"`                                           // 统计日期时间戳
	SDate              time.Time `bson:"-"`                                              // 统计日期
	ClicksCount        int64     `json:"clicksCount" bson:"clicks_count"`                // 点击次数
	TotalNumber        int64     `json:"totalNumber" bson:"total_number"`                // 点击人数
	CreateRoomCount    int64     `json:"createRoomCount" bson:"create_room_count"`       // 创建房间次数
	CreateRoomNumber   int64     `json:"createRoomNumber" bson:"create_room_number"`     // 创建房间人数
	SuccessEnterCount  int64     `json:"successEnterCount" bson:"success_enter_count"`   // 成功进入房间次数
	SuccessEnterNumber int64     `json:"successEnterNumber" bson:"success_enter_number"` // 成功进入房间人数
	TPEnterCount       int64     `json:"tpEnterCount" bson:"tp_enter_count"`
	TPEnterNumber      int64     `json:"tpEnterNumber" bson:"tp_enter_number"`
	RMEnterCount       int64     `json:"rmEnterCount" bson:"rm_enter_count"`
	RMEnterNumber      int64     `json:"rmEnterNumber" bson:"rm_enter_number"`
	ABEnterCount       int64     `json:"abEnterCount" bson:"ab_enter_count"`
	ABEnterNumber      int64     `json:"abEnterNumber" bson:"ab_enter_number"`
	JoinRoomCount      int64     `json:"joinRoomCount" bson:"join_room_count"`
	JoinRoomNumber     int64     `json:"joinRoomNumber" bson:"join_room_number"`
	SuccessJoinCount   int64     `json:"successJoinCount" bson:"success_join_count"`
	SuccessJoinNumber  int64     `json:"successJoinNumber" bson:"success_join_number"`
	TPJoinCount        int64     `json:"tpJoinCount" bson:"tp_join_count"`
	TPJoinNumber       int64     `json:"tpJoinNumber" bson:"tp_join_number"`
	RMJoinCount        int64     `json:"rmJoinCount" bson:"rm_join_count"`
	RMJoinNumber       int64     `json:"rmJoinNumber" bson:"rm_join_number"`
	ABJoinCount        int64     `json:"abJoinCount" bson:"ab_join_count"`
	ABJoinNumber       int64     `json:"abJoinNumber" bson:"ab_join_number"`
}

// 大R分析展示结构体
type UserAnalysisInfo struct {
	UserId          string    `json:"userId" bson:"-"`          // 用户id
	Channel         string    `json:"channel" bson:"-"`         // 渠道
	Channel1        string    `json:"channel1" bson:"-"`        // 渠道别名
	RegistTime      time.Time `json:"registTime" bson:"-"`      // 注册日期
	LastLoginTime   time.Time `json:"lastLoginTime" bson:"-"`   // 最后登录日期
	ActiveTime      int64     `json:"activeTime" bson:"-"`      // 存活时长（最后登陆日-注册日）
	ActiveDay       int64     `json:"activeDay" bson:"-"`       // 活跃天数（有登陆游戏的天数）
	AssociatedUsers []string  `json:"associatedUsers" bson:"-"` // 关联账号（设备、银行卡有关联的都算上）
	AssociatedCount int64     `json:"associatedCount" bson:"-"` // 关联账号
	FirstPayDate    time.Time `json:"firstPayDate" bson:"-"`    // 首充日期
	FirstPayAmount  float64   `json:"firstPayAmount" bson:"-"`  // 首充金额
	PayAmount       float64   `json:"payAmount" bson:"-"`       // 总充值金额
	PayCount        int64     `json:"payCount" bson:"-"`        // 总充值次数
	PayAvg          float64   `json:"payAvg" bson:"-"`          // 充值单均价
	WithdrawAmount  float64   `json:"withdrawAmount" bson:"-"`  // 总提现金额
	WithdrawCount   int64     `json:"withdrawCount" bson:"-"`   // 总提现次数
	WithdrawAvg     float64   `json:"withdrawAvg" bson:"-"`     // 提现单均价
	GainAmout       float64   `json:"gainAmout" bson:"-"`       // 盈利金额
	GameCount       int64     `json:"gameCount" bson:"-"`       // 玩的最多的游戏
	GameName        string    `json:"gameName" bson:"-"`        // 玩的最多的游戏名称
	TwoGameCount    int64     `json:"twoGameCount" bson:"-"`    // 玩的第二多的游戏
	TwoGameName     string    `json:"twoGameName" bson:"-"`     // 玩的第二多的游戏名称
}

// PlayerAnalysisInfo 玩家分析
type PlayerAnalysisInfo struct {
	UserId        string // 用户ID
	Money         int64  // 充值金额
	UserType      string // 用户类型（新手、平民、小R、中R、大R、超大R）
	ChannelClass  string // 渠道类
	Channel       string // 渠道
	Channel1      string // 渠道别名
	RegistTime    string // 注册日期
	LastLoginTime string // 最后登录日期
	LoseDays      string // 流失天数
	ActiveTime    string // 存活时长（最后登陆日-注册日）
	ActiveDay     int64  // 活跃天数
	FActiveDay    string // 活跃天数（有登陆游戏的天数）
	RewardRate    string // 总返奖率
	Profit        string // 盈利金额
	RefUsers      string // 关联账号

	FirstPayDate        string // 首充日期
	FirstPayAmount      string // 首充金额
	PayAmount           string // 总充金额
	PayCount            int64  // 总充单数
	PayAvg              string // 充值单均价
	PaySuccessRate      string // 充值成功率
	FirstWithdrawDate   string // 首提日期
	FirstWithdrawAmount string // 首提金额
	WithdrawAmount      string // 总提金额
	WithdrawCount       int64  // 总提单数
	WithdrawAvg         string // 提现单均价
	WithdrawSuccessRate string // 提现成功率
	FirstPayEfficience  string // 首充效率
	PayWithdrawDiffDay  string // 首提距首充间隔
	PayBeforeCarry      string // 每次充值前携带金额均值
	PayWithdrawTimes    string // 充值次数/提现次数

	MaxPayDate                      string // 单笔最大充值金额/日期
	MaxWithdrawDate                 string // 单笔最大提现金额/日期
	MaxPayDateSuccessRate           string // 单日累计最大充值金额/日期/成功率
	MaxWithdrawDateSuccessRate      string // 单日累计最大提现金额/日期/成功率
	MaxPayTimesDateSuccessRate      string // 单日最高充值次数/日期/成功率
	MaxWithdrawTimesDateSuccessRate string // 单日最高提现次数/日期/成功率

	Game1TimesStats   string // 玩最多的游戏/局数/局均码量/返奖率
	Game1BetStats     string // 打码最多的游戏/局数/局均码量/返奖率
	Game2TimesStats   string // 玩第二多的游戏/局数/局均码量/返奖率
	Game2BetStats     string // 打码第二多的游戏/局数/局均码量/返奖率
	Game3TimesStats   string // 玩第三多的游戏/局数/局均码量/返奖率
	Game3BetStats     string // 打码第三多的游戏/局数/局均码量/返奖率
	Game1PlayWinsAvg  string // 玩的多的游戏赢局局均赢钱金额
	Game1PlayLosesAvg string // 玩的多的游戏输局局均输钱金额
	Game1BetWinsAvg   string // 打码多的游戏赢局局均赢钱金额
	Game1BetLosesAvg  string // 打码多的游戏输局局均输钱金额
	MaxWinStats       string // 最大单局赢钱金额/所在游戏/日期
	MaxLosesStats     string // 最大单局输钱金额/所在游戏/日期
	ActiveDayAvg      string // 日均在线时长

	AD__ADID string   // 设备号
	IPs      []string // 关联ip
	Banks    []string // 关联银行卡号
}

type UserLevelInfo struct {
	Date         string // 日期
	ChannelClass string // 渠道类
	Channel1     string // 渠道别名
	RegistCount  int64  // 注册人数
	ActiveCount  int64  // 日活人数
	Regist0Count int64  // 注册：新手数
	Regist1Count int64  // 注册：平民数
	Regist2Count int64  // 注册：普r数
	Regist3Count int64  // 注册：小r数
	Regist4Count int64  // 注册：中r数
	Regist5Count int64  // 注册：大r数
	Regist6Count int64  // 注册：超大r数
	Regist0Rate  string // 注册：新手占比
	Regist1Rate  string // 注册：平民占比
	Regist2Rate  string // 注册：普r占比
	Regist3Rate  string // 注册：小r占比
	Regist4Rate  string // 注册：中r占比
	Regist5Rate  string // 注册：大r占比
	Regist6Rate  string // 注册：超大r占比
	Active0Count int64  // 日活：新手数
	Active1Count int64  // 日活：平民数
	Active2Count int64  // 日活：普r数
	Active3Count int64  // 日活：小r数
	Active4Count int64  // 日活：中r数
	Active5Count int64  // 日活：大r数
	Active6Count int64  // 日活：超大r数
	Active0Rate  string // 日活：新手占比
	Active1Rate  string // 日活：平民占比
	Active2Rate  string // 日活：普r占比
	Active3Rate  string // 日活：小r占比
	Active4Rate  string // 日活：中r占比
	Active5Rate  string // 日活：大r占比
	Active6Rate  string // 日活：超大r占比

	ChannelClassMap map[string]bool // 渠道类列表
	ChannelMap      map[string]bool // 渠道列表
	Channels        []string        // 渠道列表
}

// 用户转化效率表
type UserTransEffect struct {
	SDate        string // 日期
	ChannelClass string // 渠道类
	Channel1     string // 渠道别名
	RegistCount  int64  // 注册人数
	ActiveCount  int64  // 日活人数
	C0_1         int64  // 1日零充数
	C0_3         int64  // 3日零充数
	C0_7         int64  // 7日零充数
	C0_10        int64  // 10日零充数
	C0_15        int64  // 15日零充数
	C0_20        int64  // 20日零充数
	C0_25        int64  // 25日零充数
	C0_30        int64  // 30日零充数
	C2_1         int64  // 1日普充数
	C2_3         int64  // 3日普充数
	C2_7         int64  // 7日普充数
	C2_10        int64  // 10日普充数
	C2_15        int64  // 15日普充数
	C2_20        int64  // 20日普充数
	C2_25        int64  // 25日普充数
	C2_30        int64  // 30日普充数
	C3_1         int64  // 1日小r数
	C3_3         int64  // 3日小r数
	C3_7         int64  // 7日小r数
	C3_10        int64  // 10日小r数
	C3_15        int64  // 15日小r数
	C3_20        int64  // 20日小r数
	C3_25        int64  // 25日小r数
	C3_30        int64  // 30日小r数
	C4_1         int64  // 1日中r数
	C4_3         int64  // 3日中r数
	C4_7         int64  // 7日中r数
	C4_10        int64  // 10日中r数
	C4_15        int64  // 15日中r数
	C4_20        int64  // 20日中r数
	C4_25        int64  // 25日中r数
	C4_30        int64  // 30日中r数
	C5_1         int64  // 1日大r数
	C5_3         int64  // 3日大r数
	C5_7         int64  // 7日大r数
	C5_10        int64  // 10日大r数
	C5_15        int64  // 15日大r数
	C5_20        int64  // 20日大r数
	C5_25        int64  // 25日大r数
	C5_30        int64  // 30日大r数
	C6_1         int64  // 1日超大r数
	C6_3         int64  // 3日超大r数
	C6_7         int64  // 7日超大r数
	C6_10        int64  // 10日超大r数
	C6_15        int64  // 15日超大r数
	C6_20        int64  // 20日超大r数
	C6_25        int64  // 25日超大r数
	C6_30        int64  // 30日超大r数
	C0Rate_1     string // 1日零充占比
	C0Rate_3     string // 3日零充占比
	C0Rate_7     string // 7日零充占比
	C0Rate_10    string // 10日零充占比
	C0Rate_15    string // 15日零充占比
	C0Rate_20    string // 20日零充占比
	C0Rate_25    string // 25日零充占比
	C0Rate_30    string // 30日零充占比
	C2Rate_1     string // 1日普充占比
	C2Rate_3     string // 3日普充占比
	C2Rate_7     string // 7日普充占比
	C2Rate_10    string // 10日普充占比
	C2Rate_15    string // 15日普充占比
	C2Rate_20    string // 20日普充占比
	C2Rate_25    string // 25日普充占比
	C2Rate_30    string // 30日普充占比
	C3Rate_1     string // 1日小r占比
	C3Rate_3     string // 3日小r占比
	C3Rate_7     string // 7日小r占比
	C3Rate_10    string // 10日小r占比
	C3Rate_15    string // 15日小r占比
	C3Rate_20    string // 20日小r占比
	C3Rate_25    string // 25日小r占比
	C3Rate_30    string // 30日小r占比
	C4Rate_1     string // 1日中r占比
	C4Rate_3     string // 3日中r占比
	C4Rate_7     string // 7日中r占比
	C4Rate_10    string // 10日中r占比
	C4Rate_15    string // 15日中r占比
	C4Rate_20    string // 20日中r占比
	C4Rate_25    string // 25日中r占比
	C4Rate_30    string // 30日中r占比
	C5Rate_1     string // 1日大r占比
	C5Rate_3     string // 3日大r占比
	C5Rate_7     string // 7日大r占比
	C5Rate_10    string // 10日大r占比
	C5Rate_15    string // 15日大r占比
	C5Rate_20    string // 20日大r占比
	C5Rate_25    string // 25日大r占比
	C5Rate_30    string // 30日大r占比
	C6Rate_1     string // 1日超大r占比
	C6Rate_3     string // 3日超大r占比
	C6Rate_7     string // 7日超大r占比
	C6Rate_10    string // 10日超大r占比
	C6Rate_15    string // 15日超大r占比
	C6Rate_20    string // 20日超大r占比
	C6Rate_25    string // 25日超大r占比
	C6Rate_30    string // 30日超大r占比

	// Channels          []string
	Channels1Map      map[string]bool
	ChannelClassesMap map[string]bool
}

type UserLevelFunnelStats struct {
	Channel      string
	Channel1     string
	ChannelClass string

	Stats []*UserLevelFunnelStat
}

// 进量分布图
type UserRegChannelStat struct {
	SDate    string             `json:"sdate"` // 日期
	RegUsers int64              `json:"regUsers"`
	Channels []*UserRegsChannel `json:"channels"`
}

type UserRegsChannel struct {
	Channel      string `json:"channel"`
	Channel1     string `json:"channel1"`
	ChannelClass string `json:"channelClass"`
	RegUsers     int64  `json:"regUsers"`
}

type UserLevelFunnelStat struct {
	SDate        string // 日期
	Channel      string
	Channel1     string
	ChannelClass string
	C0Rate       string
	C2Rate       string
	C3Rate       string
	C4Rate       string
	C5Rate       string
	C6Rate       string

	C0 int64
	C2 int64
	C3 int64
	C4 int64
	C5 int64
	C6 int64
}

// 玩家充值统计
type PlayerRechargeInfo struct {
	Userid       string                   `bson:"-"` // 用户ID
	TotalCount   int64                    `bson:"-"` // 总充值次数
	TotalAmount  int64                    `bson:"-"` // 总充值金额
	SucceedCount int64                    `bson:"-"` // 成功单数
	TotalAvg     float64                  `bson:"-"` // 总充值均价
	Details      *[]PlayerRechargeDetails `bson:"-"`
}

type PlayerRechargeDetails struct {
	Amount   int64   `bson:"-"` // 金额
	FAmount  float64 `bson:"-"` // 金额
	ShopType int64   `bson:"-"` // 商品类型
	ShopName string  `bson:"-"` // 商品名称
	Gtype    int64   `bson:"-"` // 打码量最多的游戏id
	GameName string  `bson:"-"` // 打码量最多的游戏名称
}

type LogGameFlowWater struct {
	Userid string                        `json:"userid" bson:"_id"`
	Detail map[int][]*FlowWaterDetailLog `json:"detail" bson:"detail"`
	Ctime  int64                         `json:"ctime" bson:"ctime"`
}

type FlowWaterDetailLog struct {
	Value   int64 `json:"value" bson:"value"`
	Gtype   int32 `json:"gtype" bson:"gtype"`
	Outside int32 `json:"outside" bson:"outside"` // 1：slots 2：live
}

// 新手库存
type NewbieStock struct {
	Id          int32   `bson:"_id" json:"id"` //gtype
	GName       string  `bson:"-"`
	CashStock   int64   `bson:"cash_stock" json:"cash_stock"`       //彩金库存
	CashMingTax int64   `bson:"cash_ming_tax" json:"cash_ming_tax"` //彩金明税
	CashAnTax   int64   `bson:"cash_an_tax" json:"cash_an_tax"`     //彩金暗税
	PartIn      int64   `json:"partIn" bson:"part_in"`              //参与人次
	PlayTimes   int64   `bson:"play_times" json:"play_times"`       //总局数
	GameTime    int64   `bson:"game_time" json:"game_time"`         //总时长
	FCashStock  float64 `bson:"-"`
	RoundsAvg   float64 `bson:"-"`
	GTimeAvg    float64 `bson:"-"`
}

// 通知服务器
type ModifyNewbieStock struct {
	Gtype int32 `json:"gtype" bson:"gtype"` // 新手库存
	Stock int64 `json:"stock" bson:"stock"` // 值
}

// 首充分析
type FirstChargeAnalysis struct {
	Id          string       `bson:"_id"`                              // 主键id
	Date        int64        `bson:"date"`                             // 统计日期时间戳
	SDate       time.Time    `bson:"-"`                                // 统计日期
	Channel     string       `json:"channel" bson:"channel"`           // 渠道
	Channel1    string       `json:"channel1" bson:"channel1"`         // 渠道别名
	RegNumber   int64        `json:"regNumber" bson:"reg_number"`      // 注册人数
	FirstNumber int64        `json:"firstNumber" bson:"first_number"`  // 首充人数
	TPCharge    []GameCharge `json:"tpCharge" bson:"tp_charge"`        // TP
	RMCharge    []GameCharge `json:"rmCharge" bson:"rm_charge"`        // Rummy
	LHDCharge   []GameCharge `json:"lhdCharge" bson:"lhd_charge"`      // LHD
	UPCharge    []GameCharge `json:"upCharge" bson:"up_charge"`        // 7UP
	AK47Charge  []GameCharge `json:"ak47Charge" bson:"ak47_charge"`    // AK47
	JokerCharge []GameCharge `json:"jokerCharge" bson:"joker_charge"`  // Joker
	CrashCharge []GameCharge `json:"crashCharge" bson:"crash_charge"`  // Crash
	CPCharge    []GameCharge `json:"cpCharge" bson:"cp_charge"`        // 彩票
	ABCharge    []GameCharge `json:"abCharge" bson:"ab_charge"`        // AB
	FJCharge    []GameCharge `json:"fjCharge" bson:"fj_charge"`        // 飞机
	RMTwoCharge []GameCharge `json:"rmTwoCharge" bson:"rm_two_charge"` // RM双人
	QTNumber    int64        `json:"qtNumber" bson:"qt_number"`        // 其他
	QTRate      float64      `bson:"-"`                                // 其他
}

type GameCharge struct {
	GType  int     `json:"gType" bson:"gtype"`   // 游戏类型
	Amount int64   `json:"amount" bson:"amount"` // 充值金额
	Number int64   `json:"number" bson:"number"` // 充值人数
	Rate   float64 `bson:"-"`
}

type LtvStat struct {
	SDate         string        `json:"sdate"`         // 时间
	RegUsers      int64         `json:"regUsers"`      // 新注册数
	RegDevices    int64         `json:"regDevices"`    // 新设备数
	PayUsers      int64         `json:"payUsers"`      // 付费用户数
	Pay2UsersRate string        `json:"pay2UsersRate"` // 复购率
	PayOrdersAvg  string        `json:"payOrdersAvg"`  // 人均付费次数
	Pay2Users     int64         `json:"pay2Users"`     // 复购人数
	PayOrders     int64         `json:"payOrders"`     // 总充值单数
	Ltvs          []*LtvStatDay `json:"ltvs"`          // ltv天统计
}

type LtvStatDay struct {
	Future              bool   `json:"future"` // 是否未来时间
	Day                 int64  `json:"day"`
	PayUsers            int64  `json:"payUsers"`
	PayAvg              string `json:"payAvg"`              // 充值平均值
	WithdrawAvg         string `json:"withdrawAvg"`         // 提现平均值
	ProfitAvg           string `json:"profitAvg"`           // 盈利
	ProfitAvgLtv1Rate   string `json:"profitAvgLtv1Rate"`   // 与首日盈利比
	NVProfitAvg         string `json:"nvProfitAvg"`         // ltv净值
	NVProfitAvgLtv1Rate string `json:"nvProfitAvgLtv1Rate"` // 与首日净值比
	Pays                string `json:"pays"`                // 总充值
	Withdraws           string `json:"withdraws"`           // 总提现
	Profit              string `json:"profit"`              // 总盈利
	ProfitLtv1Rate      string `json:"profitLtv1Rate"`      // 与首日盈利比
	NVProfit            string `json:"nvProfit"`            // 总净值
	NVProfitLtv1Rate    string `json:"nvProfitLtv1Rate"`    // 与首日净值比

	Pays0         int64           `json:"pays0"`         // 总充值
	Withdraws0    int64           `json:"withdraws0"`    // 总提现
	PaysTax0      int64           `json:"paysTax0"`      // 总充值代收手续费
	WithdrawsTax0 int64           `json:"withdrawsTax0"` // 总提现代付手续费
	ProfitAvg0    float64         `json:"profitAvg0"`
	NVProfitAvg0  float64         `json:"nvProfitAvg0"`
	Profit0       int64           `json:"profit0"`
	NVProfit0     int64           `json:"nvProfit0"`
	PayUsersMap   map[string]bool `json:"-"`

	BetsAvg               string `json:"betsAvg"`               // 打码平均值
	RebateAvg             string `json:"rebateAvg"`             // 返奖平均值
	GameIncomeAvg         string `json:"gameIncomeAvg"`         // 游戏收入
	GameIncomeAvgLtv1Rate string `json:"gameIncomeAvgLtv1Rate"` // 与首日游戏收入比
	Bets                  string `json:"bets"`                  // 总打码
	Rebates               string `json:"rebates"`               // 总返奖
	GameIncome            string `json:"gameIncome"`            // 总游戏收入
	GameIncomeLtv1Rate    string `json:"gameIncomeLtv1Rate"`    // 与首日游戏收入比
	Bets0                 int64  `json:"bets0"`                 // 总打码
	Rebate0               int64  `json:"rebate0"`               // 总返奖
	GiveCash0             int64  `json:"giveCash0"`             // 发放cash额

	GameIncomeSubGiveAvg         string `json:"gameIncomeSubGiveAvg"`         // 游戏收入减去发放
	GameIncomeSubGiveAvgLtv1Rate string `json:"gameIncomeSubGiveAvgLtv1Rate"` // 与首日游戏收入比
	GameIncomeSubGive            string `json:"gameIncomeSubGive"`            // 总游戏收入减去发放
	GameIncomeSubGiveLtv1Rate    string `json:"gameIncomeSubGiveLtv1Rate"`    // 与首日游戏收入比
}

type RecoveryCycle struct {
	SDate          string // 日期
	Consume        string // 消耗
	NewRegs        int64  // 新增用户数
	FirstCharges   int64  // 首充人数
	Charges        int64  // 累计付费人数
	PayRates       string // 累计付费率
	EachRegCost    string // 单个注成本
	EachChargeCost string // 单个首充成本
	EachPayCost    string // 单个付费成本
	Profit         string // 累计利润
	ROI1           string // 首日ROI
	ROI3           string // 3日ROI
	ROI7           string // 7日ROI
	ROI10          string // 10日ROI
	ROI15          string // 15日ROI
	ROI20          string // 20日ROI
	ROI30          string // 30日ROI
	ROI60          string // 60日ROI
	ROI45          string // 45日ROI
	ROI1Profit     string // 首日ROI利润
	ROI3Profit     string // 3日ROI利润
	ROI7Profit     string // 7日ROI利润
	ROI10Profit    string // 10日ROI利润
	ROI15Profit    string // 15日ROI利润
	ROI20Profit    string // 20日ROI利润
	ROI30Profit    string // 30日ROI利润
	ROI45Profit    string // 45日ROI利润
	ROI60Profit    string // 60日ROI利润
}

type UserGeneralStatsFinance struct {
	No                    string // 序号
	Userid                string // 玩家id
	Nickname              string // 玩家昵称
	ChannelClass          string // 渠道类
	ChannelAlias          string // 渠道别名
	UserType              string // 玩家类型
	RegistArea            string // 账号类型
	Ctime                 string // 注册时间
	LoginTime             string // 最后登陆时间
	LiveDays              string // 存活天数
	LoseDays              string // 流失天数
	Diamond               string // 携带金额
	PayAmount             string // 总充值金额
	WithdrawAmount        string // 总提走金额
	WithdrawAmountWait    string // 待审核金额
	WithdrawAmountProcess string // 审核通过未到账金额
	WithdrawAmountFreeze  string // 被冻结金额
	RebateRate            string // 账面返奖率
	Profit                string // 账面净盈利
	ProfitCash            string // 实得净盈利
	ProfitGames           string // 自研游戏净盈利
	ProfitExternalSlots   string // 外接slots净盈利
	ProfitExternalZrsx    string // 外接真人净盈利
	VipBankUnlockBonus    string // VIPbank未解锁bonus
	VipBankClaimedBonus   string // 打码解锁并领走的金额
	VipBankUnclaimedBonus string // 打码已经解锁未领走的金额
	OtherCash             string // 其他领取的cash
	TryBonus              string // 试玩金转化的bonus占比
	FaultCash             string // 误差额
	VipCashPayRate        string // VIP中产出cash占总充值比例

	TpScoreStats          string // tp净盈利/返奖率
	RummyScoreStats       string // rummy双人净盈利/返奖率
	CrashScoreStats       string // crash净盈利/返奖率
	AviatorScoreStats     string // aviator净盈利/返奖率
	LhdScoreStats         string // lhd净盈利/返奖率
	UpScoreStats          string // 7up净盈利/返奖率
	KingvsqueenScoreStats string // kingvsqueen净盈利/返奖率
	AbScoreStats          string // ab净盈利/返奖率
	L3ScoreStats          string // l3净盈利/返奖率
	JokerScoreStats       string // joker净盈利/返奖率
	Ak47ScoreStats        string // ak47净盈利/返奖率

	PayAmount0           int64 // 总充值金额
	Profit0              int64 // 账面净盈利
	ProfitGames0         int64 // 自研游戏净盈利
	ProfitExternalSlots0 int64 // 外接slots净盈利
	ProfitExternalZrsx0  int64 // 外接真人净盈利
	VipBankClaimedBonus0 int64 // 打码解锁并领走的金额
	OtherCash0           int64 // 其他领取的cash
	TryBonus0            int64 // 试玩金转化成bonus的金额

	// 类型游戏输赢
	GtypeWins map[int64][]int64
}

type UserGeneralStatsGames struct {
	No           string // 序号
	Userid       string // 玩家id
	Nickname     string // 玩家昵称
	ChannelClass string // 渠道类
	ChannelAlias string // 渠道别名
	UserType     string // 玩家类型
	RegistArea   string // 账号类型
	Ctime        string // 注册时间
	LoginTime    string // 最后登陆时间
	LiveDays     string // 存活天数
	LoseDays     string // 流失天数

	AllRounds       int64  // 总游戏局数
	Bets            string // 总打码量
	BetsAvg         string // 局均打码
	Cash            string // 净赢
	RoundsGame      int64  // 自研局数合计
	RoundsSlots     int64  // 外接slots局数合计
	RoundsZrsx      int64  // 外接真人局数合计
	RoundsRateGame  string // 自研局数占比
	RoundsRateSlots string // 外接slots局数占比
	RoundsRateZrsx  string // 外接真人局数占比
	BetsGames       string // 自研打码量合计
	BetsSlots       string // 外接slots打码量合计
	BetsZrsx        string // 外接真人打码量合计
	BetsRateGames   string // 自研打码量占比
	BetsRateSlots   string // 外接slots打码量占比
	BetsRateZrsx    string // 外接真人打码量占比

	TpRoundStats          string // tp局数/占比
	RummyRoundStats       string // rummy局数/占比
	CrashRoundStats       string // crash局数/占比
	AviatorRoundStats     string // aviator局数/占比
	LhdRoundStats         string // lhd局数/占比
	UpRoundStats          string // 7up局数/占比
	KingvsqueenRoundStats string // kingvsqueen局数/占比
	AbRoundStats          string // ab局数/占比
	L3RoundStats          string // l3局数/占比
	JokerRoundStats       string // joker局数/占比
	Ak47RoundStats        string // ak47局数/占比

	TpBetsStats          string // tp打码量/占比
	RummyBetsStats       string // rummy打码量/占比
	CrashBetsStats       string // crash打码量/占比
	AviatorBetsStats     string // aviator打码量/占比
	LhdBetsStats         string // lhd打码量/占比
	UpBetsStats          string // 7up打码量/占比
	KingvsqueenBetsStats string // kingvsqueen打码量/占比
	AbBetsStats          string // ab打码量/占比
	L3BetsStats          string // l3打码量/占比
	JokerBetsStats       string // joker打码量/占比
	Ak47BetsStats        string // ak47打码量/占比

	Bets0      int64 // 总打码量
	Cash0      int64 // 净赢
	BetsGames0 int64 // 自研打码量合计
	BetsSlots0 int64 // 外接slots打码量合计
	BetsZrsx0  int64 // 外接真人打码量合计

	// 类型游戏局数和打码
	GtypeStats map[int64][]int64
}

type UserGeneralStatsVBBank struct {
	No           string // 序号
	Userid       string // 玩家id
	Nickname     string // 玩家昵称
	ChannelClass string // 渠道类
	ChannelAlias string // 渠道别名
	UserType     string // 玩家类型
	RegistArea   string // 账号类型
	Ctime        string // 注册时间
	LoginTime    string // 最后登陆时间
	LiveDays     string // 存活天数
	LoseDays     string // 流失天数

	VipLevelMax    string // VIP最高等级
	VipLevel       string // VIP当前等级
	PayAmount      string // 总充值金额
	Bets           string // 总打码量
	BetsPayRate    string // 打码量与充值金额比
	BetsBonusRate  string // 打码解锁比例
	BonusBetsRate  string // bonus与打码量比
	BonusPayRate   string // bonus与总充值金额比
	VipCashPayRate string // VIP中产出cash占总充值比例
	VipBonus       string // VIPbank累计bonus总额
	VipBankBonus   string // VIPbank未解锁bonus
	CashoutBonus   string // 打码解锁并领走的金额
	UnlockBonus    string // 打码已经解锁未领走的金额
	FaultBank      string // bank总流水误差额

	TryBonus        string // 试玩金转化成bonus的金额
	VipUpgradeBonus string // VIP升级奖金
	VipWeekBonus    string // VIP周奖金
	ChargeBonus     string // 充值时赠送的bonus
	OnlineBonus     string // 在线时长抽奖获得的bonus
	TaskBonus       string // 完成任务获得的bonus
	ShareBetsBonus  string // 分享下家的打码返bonus
	ShareBonus      string // 分享下家首充返bonus
	InterestBonus   string // 利息
	FaultBonus      string // bonus来源误差额

	TryBonusRate        string // 试玩金转化的bonus占比
	VipUpgradeBonusRate string // VIP升级奖金的bonus占比
	VipWeekBonusRate    string // VIP周奖金的bonus占比
	ChargeBonusRate     string // 充值时赠送的的bonus占比
	OnlineBonusRate     string // 在线时长抽奖获得的bonus占比
	TaskBonusRate       string // 完成任务获得的bonus占比
	ShareBetsBonusRate  string // 分享下家的打码返bonus占比
	ShareBonusRate      string // 分享下家首充返bonus占比

	Bets0         int64 // 总打码量
	PayAmount0    int64 // 总充值金额
	VipBankBonus0 int64 // VIPbank未解锁bonus
	CashoutBonus0 int64 // 打码解锁并领走的金额
	UnlockBonus0  int64 // 打码已经解锁未领走的金额

	AllBonus         int64 // VIPbank累计bonus总额
	TryBonus0        int64 // 试玩金转化成bonus的金额
	VipUpgradeBonus0 int64 // VIP升级奖金
	VipWeekBonus0    int64 // VIP周奖金
	ChargeBonus0     int64 // 充值时赠送的bonus
	OnlineBonus0     int64 // 在线时长抽奖获得的bonus
	TaskBonus0       int64 // 完成任务获得的bonus
	ShareBetsBonus0  int64 // 分享下家的打码返bonus
	ShareBonus0      int64 // 分享下家首充返bonus
	InterestBonus0   int64 // 利息
}

// 实时数据今日统计
type LiveDataTodayStats struct {
	Users         int64   `json:"users"`
	Pays          float64 `json:"pays"`
	Withdraws     float64 `json:"withdraws"`
	Bets          float64 `json:"bets"`
	OnlineUsers   int64   `json:"onlineUsers"`
	PayUsers      int64   `json:"payUsers"`
	WithdrawUsers int64   `json:"withdrawUsers"`
	RabateRate    string  `json:"rabateRate"`
}

// 实时数据走势曲线
type LiveDataTrendStats struct {
	Trend   int32     `json:"trend"` // 指标类型
	Name    string    `json:"name"`  // 指标名
	DayTime uint32    `json:"dayTime"`
	Datas   []float64 `json:"datas"` // 数值*1440
}

// 玩家打码分层
type PlayerBetsLevel struct {
	SDate              string // 日期
	Bets               string // 总打码量
	Top05Users         int64  // 打码前0.5%玩家人数
	Top05UsersBets     string // 打码前0.5%玩家码量总和
	Top05UsersBetsAvg  string // 打码前0.5%玩家人均码量
	Top05UsersBetsRate string // 打码前0.5%玩家码量占总码量比
	Top1Users          int64  // 打码前1%玩家人数
	Top1UsersBets      string // 打码前1%玩家码量总和
	Top1UsersBetsAvg   string // 打码前1%玩家人均码量
	Top1UsersBetsRate  string // 打码前1%玩家码量占总码量比
	Top3Users          int64  // 打码前3%玩家人数
	Top3UsersBets      string // 打码前3%玩家码量总和
	Top3UsersBetsAvg   string // 打码前3%玩家人均码量
	Top3UsersBetsRate  string // 打码前3%玩家码量占总码量比
	Top5Users          int64  // 打码前5%玩家人数
	Top5UsersBets      string // 打码前5%玩家码量总和
	Top5UsersBetsAvg   string // 打码前5%玩家人均码量
	Top5UsersBetsRate  string // 打码前5%玩家码量占总码量比
	Top10Users         int64  // 打码前10%玩家人数
	Top10UsersBets     string // 打码前10%玩家码量总和
	Top10UsersBetsAvg  string // 打码前10%玩家人均码量
	Top10UsersBetsRate string // 打码前10%玩家码量占总码量比

	Bets0           int64 // 总打码量
	Top05UsersBets0 int64 // 打码前0.5%玩家码量总和
	Top1UsersBets0  int64 // 打码前1%玩家码量总和
	Top3UsersBets0  int64 // 打码前3%玩家码量总和
	Top5UsersBets0  int64 // 打码前5%玩家码量总和
	Top10UsersBets0 int64 // 打码前10%玩家码量总和
}

type PlayerBetsUser struct {
	SDate            string // 日期
	Rank             string // 名次
	Userid           string // UID
	Ctime            string // 注册时间
	LastLoginTime    string // 最后登录时间
	LiveDays         string // 存活天数
	LoseDays         string // 流失天数
	VipLv            string // VIP等级
	UserType         string // 玩家类型
	Pays             string // 玩家总充值
	Withdraws        string // 玩家总提现
	PaysSubWithdraws string // 玩家总充-提
	Bets             string // 玩家总打码量
	RabateRate       string // 玩家总返奖率
	Wins             string // 玩家总输赢
	Top1Game         string // 打码第1多的游戏
	Top1GameBets     string // 打码第1多的游戏的码量
	Top1GameRounds   string // 打码第1多的游戏的局数
	Top1GameBetsAvg  string // 打码第1多的游戏的局均码
	Top2Game         string // 打码第2多的游戏
	Top2GameBets     string // 打码第2多的游戏的码量
	Top2GameRounds   string // 打码第2多的游戏的局数
	Top2GameBetsAvg  string // 打码第2多的游戏的局均码
	Top3Game         string // 打码第3多的游戏
	Top3GameBets     string // 打码第3多的游戏的码量
	Top3GameRounds   string // 打码第3多的游戏的局数
	Top3GameBetsAvg  string // 打码第3多的游戏的局均码

	Bets0   int64 // 玩家总打码量
	Rabate0 int64 // 玩家总返奖
}

// 子游戏打码统计
type PlayerBetsGtypeDate struct {
	SDate          string // 日期
	GtypeCategory1 string // 游戏大类
	GtypeCategory2 string // 游戏小类
	Factory        string // 厂商名称
	GameId         string // 游戏ID
	GameName       string // 游戏名称
	BetUsers       int64  // 投注人数
	Bets           string // 投注金额
	BetsAvg        string // 人均投注
	Rebates        string // 厂商回退金额
	BetsRollback   string // 厂商回滚金额
	Wins           string // 玩家总输赢
	RTP            string // RTP
}
