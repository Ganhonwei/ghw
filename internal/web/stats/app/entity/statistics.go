package entity

import "time"

// 数据汇总
type DataStatistics struct {
	Id                   string    `bson:"_id"`                                                // 主键id
	Date                 int64     `bson:"date"`                                               // 统计日期时间戳
	SDate                time.Time `bson:"-"`                                                  // 统计日期
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
	TotalProfit         int64   `json:"totalProfit" bson:"total_profit"`        // 总利润=总充值金额-总提现金额-手续费
	DynamicProfitAvg    float64 `bson:"-"`                                      // 活跃人均利润=总利润÷登录人数
	PayDynamicProfitAvg float64 `bson:"-"`                                      // 活付人均利润=总利润÷付费用户登录人数
	PayMinutes30        int64   `json:"payMinutes30" bson:"pay_minutes30"`      // 注册30分钟内付费人数
	PayMinutes60        int64   `json:"payMinutes60" bson:"pay_minutes60"`      // 注册60分钟内付费人数
	PayMinutes120       int64   `json:"payMinutes120" bson:"pay_minutes120"`    // 注册120分钟内付费人数
	PayMinutesRate30    float64 `bson:"-"`                                      // 30分钟付费率=30分钟内付费人数 /当天注册人数
	PayMinutesRate60    float64 `bson:"-"`                                      // 60分钟付费率=60分钟内付费人数 /当天注册人数
	PayMinutesRate120   float64 `bson:"-"`                                      // 120分钟付费率=120分钟内付费人数 /当天注册人数
	FTotalProfit        float64 `bson:"-"`
}

// 数据汇总
type DataStatistics4Web struct {
	Id                   string    `bson:"_id"`                                                // 主键id
	Date                 int64     `bson:"date"`                                               // 统计日期时间戳
	SDate                time.Time `bson:"-"`                                                  // 统计日期
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
	//新增字段
	PayLoginNumber      int64   `json:"payLoginNumber" bson:"pay_login_number"` // 付费用户登录人数
	TotalProfit         int64   `json:"totalProfit" bson:"total_profit"`        // 总利润=总充值金额-总提现金额-手续费
	DynamicProfitAvg    float64 `bson:"-"`                                      // 活跃人均利润=总利润÷登录人数
	PayDynamicProfitAvg float64 `bson:"-"`                                      // 活付人均利润=总利润÷付费用户登录人数
	PayMinutes30        int64   `json:"payMinutes30" bson:"pay_minutes30"`      // 注册30分钟内付费人数
	PayMinutes60        int64   `json:"payMinutes60" bson:"pay_minutes60"`      // 注册60分钟内付费人数
	PayMinutes120       int64   `json:"payMinutes120" bson:"pay_minutes120"`    // 注册120分钟内付费人数
	PayMinutesRate30    float64 `bson:"-"`                                      // 30分钟付费率=30分钟内付费人数 /当天注册人数
	PayMinutesRate60    float64 `bson:"-"`                                      // 60分钟付费率=60分钟内付费人数 /当天注册人数
	PayMinutesRate120   float64 `bson:"-"`                                      // 120分钟付费率=120分钟内付费人数 /当天注册人数
	FTotalProfit        float64 `bson:"-"`
	TrustUserPay        int64   `json:"trustUserPay" bson:"trust_user_pay"`           // 信任用户充值
	TrustUserWithdraw   int64   `json:"trustUserWithdraw" bson:"trust_user_withdraw"` // 信任用户提现
	FTrustUserPay       float64 `bson:"-"`
	FTrustUserWithdraw  float64 `bson:"-"`
	TrustRate           float64 `bson:"-"`
}

// 分享数据统计
type ShareDataStatistics struct {
	Id             string    `bson:"_id"`                      // 主键id
	Date           int64     `bson:"date"`                     // 统计日期时间戳
	SDate          time.Time `bson:"-"`                        // 统计日期
	Channel        string    `json:"channel" bson:"channel"`   // 渠道
	Channel1       string    `json:"channel1" bson:"channel1"` // 渠道别名
	BloggerId      string    `bson:"blogger_id"`               // 博主用户ID
	RegisterNumber int64     `bson:"register_number"`          // 邀请注册人数
	PayNumber      int64     `bson:"pay_number"`               // 被邀请充值人数
	PayMoney       int64     `bson:"pay_money"`                // 被邀请人充值金额
	FPayMoney      float64   `bson:"-"`                        // 被邀请人充值金额
	ShareArpu      float64   `bson:"-"`
	ShareArppu     float64   `bson:"-"`
}

// 用户留存
type UserRetained struct {
	Id         string    `bson:"_id"`                          // 主键id
	Date       int64     `bson:"date"`                         // 统计日期时间戳
	SDate      time.Time `bson:"-"`                            // 统计日期
	Channel    string    `json:"channel" bson:"channel"`       // 渠道
	Channel1   string    `json:"channel1" bson:"channel1"`     // 渠道别名
	BloggerId  string    `bson:"blogger_id"`                   // 博主用户ID
	NewNumber  int64     `json:"newNumber" bson:"new_number"`  // 新增人数
	Login1     int64     `json:"login1" bson:"login1"`         // 昨日注册且充值今日登录
	Register1  int64     `json:"register1" bson:"register1"`   // 昨日注册
	Login2     int64     `json:"login2" bson:"login2"`         // 前两日注册且充值今日登录
	Register2  int64     `json:"register2" bson:"register2"`   // 前两日注册
	Login3     int64     `json:"login3" bson:"login3"`         // 前三日注册且充值今日登录
	Register3  int64     `json:"register3" bson:"register3"`   // 前三日注册
	Login4     int64     `json:"login4" bson:"login4"`         // 前四日注册且充值今日登录
	Register4  int64     `json:"register4" bson:"register4"`   // 前四日注册
	Login5     int64     `json:"login5" bson:"login5"`         // 前五日注册且充值今日登录
	Register5  int64     `json:"register5" bson:"register5"`   // 前五日注册
	Login6     int64     `json:"login6" bson:"login6"`         // 前六日注册且充值今日登录
	Register6  int64     `json:"register6" bson:"register6"`   // 前六日注册
	Login7     int64     `json:"login7" bson:"login7"`         // 前七日注册且充值今日登录
	Register7  int64     `json:"register7" bson:"register7"`   // 前七日注册
	Login15    int64     `json:"login15" bson:"login15"`       // 前十五日注册且充值今日登录
	Register15 int64     `json:"register15" bson:"register15"` // 前十五日注册
	Login30    int64     `json:"login30" bson:"login30"`       // 前三十日注册且充值今日登录
	Register30 int64     `json:"register30" bson:"register30"` // 前三十日注册
	Login60    int64     `json:"login60" bson:"login60"`       // 前六十日注册且充值今日登录
	Register60 int64     `json:"register60" bson:"register60"` // 前六十日注册
	Day1       float64   `bson:"-"`                            // 1日留存
	Day2       float64   `bson:"-"`                            // 2日留存
	Day3       float64   `bson:"-"`                            // 3日留存
	Day4       float64   `bson:"-"`                            // 4日留存
	Day5       float64   `bson:"-"`                            // 5日留存
	Day6       float64   `bson:"-"`                            // 6日留存
	Day7       float64   `bson:"-"`                            // 7日留存
	Day15      float64   `bson:"-"`                            // 15日留存
	Day30      float64   `bson:"-"`                            // 30日留存
	Day60      float64   `bson:"-"`                            // 60日留存
}

// 付费留存
type PayUserRetained struct {
	Id         string    `bson:"_id"`                          // 主键id
	Date       int64     `bson:"date"`                         // 统计日期时间戳
	SDate      time.Time `bson:"-"`                            // 统计日期
	Channel    string    `json:"channel" bson:"channel"`       // 渠道
	Channel1   string    `json:"channel1" bson:"channel1"`     // 渠道别名
	BloggerId  string    `bson:"blogger_id"`                   // 博主用户ID
	NewNumber  int64     `json:"newNumber" bson:"new_number"`  // 充值人数
	Login1     int64     `json:"login1" bson:"login1"`         // 昨日注册且充值今日登录
	Register1  int64     `json:"register1" bson:"register1"`   // 昨日注册
	Login2     int64     `json:"login2" bson:"login2"`         // 前两日注册且充值今日登录
	Register2  int64     `json:"register2" bson:"register2"`   // 前两日注册
	Login3     int64     `json:"login3" bson:"login3"`         // 前三日注册且充值今日登录
	Register3  int64     `json:"register3" bson:"register3"`   // 前三日注册
	Login4     int64     `json:"login4" bson:"login4"`         // 前四日注册且充值今日登录
	Register4  int64     `json:"register4" bson:"register4"`   // 前四日注册
	Login5     int64     `json:"login5" bson:"login5"`         // 前五日注册且充值今日登录
	Register5  int64     `json:"register5" bson:"register5"`   // 前五日注册
	Login6     int64     `json:"login6" bson:"login6"`         // 前六日注册且充值今日登录
	Register6  int64     `json:"register6" bson:"register6"`   // 前六日注册
	Login7     int64     `json:"login7" bson:"login7"`         // 前七日注册且充值今日登录
	Register7  int64     `json:"register7" bson:"register7"`   // 前七日注册
	Login15    int64     `json:"login15" bson:"login15"`       // 前十五日注册且充值今日登录
	Register15 int64     `json:"register15" bson:"register15"` // 前十五日注册
	Login30    int64     `json:"login30" bson:"login30"`       // 前三十日注册且充值今日登录
	Register30 int64     `json:"register30" bson:"register30"` // 前三十日注册
	Login60    int64     `json:"login60" bson:"login60"`       // 前六十日注册且充值今日登录
	Register60 int64     `json:"register60" bson:"register60"` // 前六十日注册
	Day1       float64   `bson:"-"`                            // 1日留存
	Day2       float64   `bson:"-"`                            // 2日留存
	Day3       float64   `bson:"-"`                            // 3日留存
	Day4       float64   `bson:"-"`                            // 4日留存
	Day5       float64   `bson:"-"`                            // 5日留存
	Day6       float64   `bson:"-"`                            // 6日留存
	Day7       float64   `bson:"-"`                            // 7日留存
	Day15      float64   `bson:"-"`                            // 15日留存
	Day30      float64   `bson:"-"`                            // 30日留存
	Day60      float64   `bson:"-"`                            // 60日留存
}

// 埋点数据
type PointData struct {
	Id                   string         `bson:"_id"`                                      // 订单id
	Date                 int64          `bson:"date"`                                     // 统计日期时间戳
	SDate                time.Time      `bson:"-"`                                        // 统计日期
	Channel              string         `json:"channel" bson:"channel"`                   // 渠道
	Channel1             string         `json:"channel1" bson:"channel1"`                 // 渠道别名
	BloggerId            string         `bson:"blogger_id"`                               // 博主用户ID
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
	BloggerId      string    `bson:"blogger_id"`                          // 博主用户ID
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

// 时间分析
type PlaytimeAnalysis struct {
	Id              string    `bson:"_id"`                                 // id
	Date            int64     `bson:"date"`                                // 统计日期时间戳
	SDate           time.Time `bson:"-"`                                   // 统计日期
	Channel         string    `json:"channel" bson:"channel"`              // 渠道
	Channel1        string    `json:"channel1" bson:"channel1"`            // 渠道别名
	BloggerId       string    `bson:"blogger_id"`                          // 博主用户ID
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

// 用户留存
type UserRetained4Web struct {
	Id        string    `bson:"_id"`                         // 主键id
	Date      int64     `bson:"date"`                        // 统计日期时间戳
	SDate     time.Time `bson:"-"`                           // 统计日期
	Channel   string    `json:"channel" bson:"channel"`      // 渠道
	Channel1  string    `json:"channel1" bson:"channel1"`    // 渠道别名
	SType     int       `bson:"s_type"`                      // 数据类型 0：全部;1：按渠道
	NewNumber int64     `json:"newNumber" bson:"new_number"` // 新增人数
	Day1      float64   `json:"day1" bson:"day1"`            // 1日留存
	Day2      float64   `json:"day2" bson:"day2"`            // 2日留存
	Day3      float64   `json:"day3" bson:"day3"`            // 3日留存
	Day4      float64   `json:"day4" bson:"day4"`            // 4日留存
	Day5      float64   `json:"day5" bson:"day5"`            // 5日留存
	Day6      float64   `json:"day6" bson:"day6"`            // 6日留存
	Day7      float64   `json:"day7" bson:"day7"`            // 7日留存
	Day15     float64   `json:"day15" bson:"day15"`          // 15日留存
	Day30     float64   `json:"day30" bson:"day30"`          // 30日留存
	Day60     float64   `json:"day60" bson:"day60"`          // 60日留存
}

// 用户留存新结构体
type UserRetainedData struct {
	Id        string    `bson:"_id"`                         // 主键id
	Date      int64     `bson:"date"`                        // 统计日期时间戳
	SDate     time.Time `bson:"-"`                           // 统计日期
	Channel   string    `json:"channel" bson:"channel"`      // 渠道
	Channel1  string    `json:"channel1" bson:"channel1"`    // 渠道别名
	NewNumber int64     `json:"newNumber" bson:"new_number"` // 新增人数
	Login1    int64     `json:"login1" bson:"login1"`        // 1日留存=昨日新注册数今日登录÷昨日新注册数
	Login2    int64     `json:"login2" bson:"login2"`
	Login3    int64     `json:"login3" bson:"login3"`

	Login4 int64 `json:"login4" bson:"login4"`

	Login5 int64 `json:"login5" bson:"login5"`

	Login6 int64 `json:"login6" bson:"login6"`

	Login7 int64 `json:"login7" bson:"login7"`

	Login15 int64 `json:"login15" bson:"login15"`

	Login30 int64 `json:"login30" bson:"login30"`

	Login60 int64 `json:"login60" bson:"login60"`
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
	RBGameNumber        int64     `json:"rbGameNumber" bson:"rb_game_number"`
	RBNumber            int64     `json:"rbNumber" bson:"rb_number"`
	RBGameTime          int64     `json:"rbGameTime" bson:"rb_game_time"`
	RMTwoGameNumber     int64     `json:"rmTwoGameNumber" bson:"rm_two_game_number"`
	RMTwoNumber         int64     `json:"rmTwoNumber" bson:"rm_two_number"`
	RMTwoGameTime       int64     `json:"rmTwoGameTime" bson:"rm_two_game_time"`
	TP2GameNumber       int64     `json:"tp2GameNumber" bson:"tp2_game_number"`
	TP2Number           int64     `json:"tp2Number" bson:"tp2_number"`
	TP2GameTime         int64     `json:"tp2GameTime" bson:"tp2_game_time"`
	SlotsGameNumber     int64     `json:"slotsGameNumber" bson:"slots_game_number"`
	SlotsNumber         int64     `json:"slotsNumber" bson:"slots_number"`
	SlotsGameTime       int64     `json:"slotsGameTime" bson:"slots_game_time"`
	ZRSXGameNumber      int64     `json:"zrsxGameNumber" bson:"zrsx_game_number"`
	ZRSXNumber          int64     `json:"zrsxNumber" bson:"zrsx_number"`
	Userids             []string  `json:"userids" bson:"userids"` // 查询的用户id
}

// 用户游戏局数统计
type UserGameData struct {
	Id         string `bson:"_id"`                                                     // id
	UserId     string `gorm:"column:userid" bson:"userid" json:"userid"`               // userid
	Gtype      int64  `gorm:"column:gtype" bson:"gtype"`                               // 游戏类型
	Number     int64  `gorm:"column:number" bson:"number"`                             // 游戏局数
	GameTime   int64  `gorm:"column:game_time" json:"gameTime" bson:"game_time"`       // 游戏时长
	Bets       int64  `gorm:"column:bets" json:"bets" bson:"bets"`                     // 总下注
	WinBets    int64  `gorm:"column:win_bets" json:"winBets" bson:"win_bets"`          // 赢局结算
	LoseBets   int64  `gorm:"column:lose_bets" json:"loseBets" bson:"lose_bets"`       // 输局结算
	Profit     int64  `json:"profit" bson:"profit"`                                    // 盈亏
	WinNumber  int64  `gorm:"column:win_number" json:"winNumber" bson:"win_number"`    // 赢局局数
	LoseNumber int64  `gorm:"column:lose_number" json:"loseNumber" bson:"lose_number"` // 输局局数
	Ctime      int64  `json:"ctime" bson:"ctime"`                                      // 创建时间
}

type GameStrategy struct {
	Id              string `bson:"_id"`                                     // id
	UserId          string `bson:"userid"`                                  // userid
	Gtype           int64  `bson:"gtype"`                                   // 游戏类型
	StrategyId      string `json:"strategyId" bson:"strategy_id"`           // 策略模型ID
	EnterNumber     int64  `json:"enterNumber" bson:"enter_number"`         // 进入次数
	EffectiveNumber int64  `json:"effectiveNumber" bson:"effective_number"` // 生效次数
	EffectiveIncome int64  `json:"effectiveIncome" bson:"effective_income"` // 生效收益
	Ctime           int64  `json:"ctime" bson:"ctime"`                      // 创建时间
}

// 拼多多分享活动
type PddStat struct {
	Id                     string `bson:"_id"`                       // id
	Date                   int64  `bson:"date"`                      // 统计日期时间戳
	LoginedUsers           int32  `bson:"logined_users"`             // 渠道总日活
	PddUsers               int32  `bson:"pdd_users"`                 // 参与用户
	BeInvoteUsers          int32  `bson:"be_invote_users"`           // 今日参与活动中，分享来的用户数，有效分享数
	BeInvoteLoginedUsers   int32  `bson:"be_invote_logined_users"`   // 被分享的用户登录人数，日活人数
	DrawUsers              int32  `bson:"draw_users"`                // 今日参与活动中，分享来的用户数
	BeInvoteUserDraws      int32  `bson:"be_invote_user_draws"`      // 分享用户参与抽奖总次数
	UserDraws              int32  `bson:"user_draws"`                // 所有渠道参与抽奖总次数
	AwardUsers             int32  `bson:"award_users"`               // 领取到最终奖励的人数
	AwardJackpot           int64  `bson:"award_jackpot"`             // 累计发出去的金额
	BeInvoteRegUsers       int32  `bson:"be_invote_reg_users"`       // 所有渠道分享来注册的用户数
	BeInvoteDevices        int32  `bson:"be_invote_devices"`         // 所有渠道分享来注册的设备数
	Share2Users            int32  `bson:"share2users"`               // 分享用户分享来人数
	Share2DrawUsers        int32  `bson:"share2draw_users"`          // 分享用户有效分享数
	Share2Devices          int32  `bson:"share2devices"`             // 分享用户分享来设备数
	PlayedReg24            int32  `bson:"played_reg24"`              // 注册24小时玩游戏人数
	RegUserChages          int32  `bson:"reg_user_chages"`           // 新顾客充值人数
	RegUserChargeAmount    int64  `bson:"reg_user_charge_amount"`    // 新顾客充值金额
	RegUserWithdrawAmount  int64  `bson:"reg_user_withdraw_amount"`  // 新顾客提现金额
	ChargeAmount           int64  `bson:"charge_amount"`             // 总充值金额
	WithdrawAmount         int64  `bson:"withdraw_amount"`           // 总提现金额
	BeInvoteChargeAmount   int64  `bson:"be_invote_charge_amount"`   // 分享用户总充值金额
	BeInvoteWithdrawAmount int64  `bson:"be_invote_withdraw_amount"` // 分线用户总提现金额
}

type CrashPlayerStat struct {
	Id                   string    `bson:"_id"` // date-userid
	Date                 int64     `bson:"date"`
	DateStr              string    `bson:"date_str"`
	Userid               string    `bson:"userid"`
	NewReg               bool      `bson:"new_reg"`                 // 是新顾客
	AllRounds            int32     `bson:"all_rounds"`              // 总局数
	GameTimes            int64     `bson:"game_times"`              // 游戏总时长(秒)
	Bets                 int64     `bson:"bets"`                    // 总打码量
	BetRounds            int32     `bson:"bet_rounds"`              // 下注局数
	WinRounds            int32     `bson:"win_rounds"`              // 胜利局数
	LoseRounds           int32     `bson:"lose_rounds"`             // 失败局数
	TieRounds            int32     `bson:"tie_rounds"`              // 和局数
	WinBets              int64     `bson:"win_bets"`                // 胜局打码量
	LoseBets             int64     `bson:"lose_bets"`               // 败局打码量
	TieBets              int64     `bson:"tie_bets"`                // 和局打码量
	Wins                 int64     `bson:"wins"`                    // 胜局赢钱金额
	Loses                int64     `bson:"loses"`                   // 败局输钱金额
	Cash                 int64     `bson:"cash"`                    // 净赢
	MulpitleSum          float64   `bson:"mulpitle_sum"`            // 所有局总爆炸倍数和
	Mulpitles            []float64 `bson:"mulpitles"`               // 局数爆炸倍数
	WinEscapeMulpitleSum float64   `bson:"win_escape_mulpitle_sum"` // 赢局逃脱倍数和
	WinEscapeMulpitles   []float64 `bson:"win_escape_mulpitles"`    // 赢局逃脱倍数
	WinMulpitleSum       float64   `bson:"win_mulpitle_sum"`        // 赢局爆炸倍数和
	WinMulpitles         []float64 `bson:"win_mulpitles"`           // 赢局数爆炸倍数
	LoseMulpitleSum      float64   `bson:"lose_mulpitle_sum"`       // 输局爆炸倍数和
	LoseMulpitles        []float64 `bson:"lose_mulpitles"`          // 输局爆炸倍数,算中位数
	RoundBetAvg          int64     `bson:"round_bet_avg"`           // 局均码
	AD_BundleId          string    `bson:"ad__bundle_id"`           // 渠道
	Channel1             string    `bson:"channel1"`                // 渠道别名
	RegistArea           int       `bson:"regist_area"`             // 账号类型 ab测试 0:A 1:B 2:C
	Money                uint32    `bson:"money"`                   // 充值总金额(分)
	ObserveRounds        int32     `bson:"observe_rounds"`          // 观察局数
}

// Crash扶摇直上
type CrashFYZSStat struct {
	Id                       string  `bson:"_id"` // date-userid
	Date                     int64   `bson:"date"`
	DateStr                  string  `bson:"date_str"`
	Userid                   string  `bson:"userid"`
	IsFit                    bool    `json:"isFit" bson:"is_fit"`                                         // 是否是适配策略的人
	IsTrigger                bool    `json:"isTrigger" bson:"is_trigger"`                                 // 是否是触发策略的人
	TiggerTimes              int     `json:"tiggerTimes" bson:"tigger_times"`                             // u 触发策略有效次数
	AllTiggerTimes           int     `json:"allTiggerTimes" bson:"all_tigger_times"`                      // u总 累计触发次数
	AllEvoTimes              int     `json:"allEvoTimes" bson:"all_evo_times"`                            // r总 累计玩游戏次数
	WinRounds                int32   `bson:"win_rounds"`                                                  // 触发策略时逃跑的局数
	LoseRounds               int     `json:"loseRounds" bson:"lose_rounds"`                               // 触发策略时未能逃跑的局数
	MaximumEscapeMultiple    float64 `json:"maximumEscapeMultiple" bson:"maximum_escape_multiple"`        // 策略中最高逃跑倍数
	MinimumEscapeMultiple    float64 `json:"minimumEscapeMultiple" bson:"minimum_escape_multiple"`        // 策略中最低逃跑倍数
	SumEscapeMulpitle        float64 `json:"sumEscapeMultiple" bson:"sum_escape_multiple"`                // 策略中逃跑倍数和
	AverageEscapeMultiple    float64 `json:"averageEscapeMultiple" bson:"average_escape_multiple"`        // 策略中平均逃跑倍数
	WinMulpitleSum           float64 `bson:"win_mulpitle_sum"`                                            // 策略中逃跑倍数和
	Bets                     int64   `bson:"bets"`                                                        // 策略中总打码量
	BetRounds                int32   `bson:"bet_rounds"`                                                  // 下注局数
	MaximumEscapeMultipleBet int64   `json:"maximumEscapeMultipleBet" bson:"maximum_escape_multiple_bet"` // 策略中最高逃跑倍数时打码量
	MinimumEscapeMultipleBet int64   `json:"minimumEscapeMultipleBet" bson:"minimum_escape_multiple_bet"` // 策略中最低逃跑倍数时打码量
	LoseMulpitleSum          float64 `json:"loseMulpitleSum" bson:"lose_mulpitle_sum"`                    // 触发策略未能逃跑的爆炸倍数和
	WinBets                  int64   `bson:"win_bets"`                                                    // 触发策略时逃跑的总打码量
	LoseBets                 int64   `json:"loseBets" bson:"lose_bets"`                                   // 触发策略时未能逃跑的总打码量
	Wins                     int64   `bson:"wins"`                                                        // 胜局赢钱金额
	Loses                    int64   `bson:"loses"`                                                       // 败局输钱金额
	WinEscapeMultiple        float64 `json:"winEscapeMultiple" bson:"win_escape_multiple"`                // 胜局平均逃脱倍数
	Cash                     int64   `bson:"cash"`                                                        // 净赢
	AD_BundleId              string  `bson:"ad__bundle_id"`                                               // 渠道
	Channel1                 string  `bson:"channel1"`                                                    // 渠道别名
	RegistArea               int     `bson:"regist_area"`                                                 // 账号类型 ab测试 0:A 1:B 2:C
	AllRounds                int32   `bson:"all_rounds"`                                                  // 总局数
	RoundBetAvg              int64   `bson:"round_bet_avg"`                                               // 局均码
}

// 欲薅无门
type CrashYHWMStat struct {
	Id                string  `bson:"_id"` // date-userid
	Date              int64   `bson:"date"`
	DateStr           string  `bson:"date_str"`
	Userid            string  `bson:"userid"`
	IsFit             bool    `json:"isFit" bson:"is_fit"`                      // 是否是适配策略的人
	IsTrigger         bool    `json:"isTrigger" bson:"is_trigger"`              // 是否是触发策略的人
	TiggerTimes       int     `json:"tiggerTimes" bson:"tigger_times"`          // 触发策略有效次数
	AllTiggerTimes    int     `json:"allTiggerTimes" bson:"all_tigger_times"`   // u总 累计触发次数
	AllEvoTimes       int     `json:"allEvoTimes" bson:"all_evo_times"`         // 触发策略局数
	BurstNumber       int     `json:"burstNumber" bson:"burst_number"`          // 瞬爆发生局数
	BurstAmount       int64   `json:"burstAmount" bson:"burst_amount"`          // 瞬爆收割金额 策略触发后，因为瞬爆导致玩家输钱的总金额
	TotalReapAmount   int64   `json:"totalReapAmount" bson:"total_reap_amount"` // 策略中总割金额 策略开始触发到策略结束进入冷却期间，玩家输钱的总金额
	WinRounds         int     `bson:"win_rounds"`                               // 策略局中玩家胜局数
	LoseRounds        int     `json:"loseRounds" bson:"lose_rounds"`            // 策略局中玩家败局数
	Bets              int64   `bson:"bets"`                                     // 总打码量
	BetRounds         int32   `bson:"bet_rounds"`                               // 下注局数
	WinBets           int64   `bson:"win_bets"`                                 // 赢局打码量
	LoseBets          int64   `json:"loseBets" bson:"lose_bets"`                // 输局打码量
	Wins              int64   `bson:"wins"`                                     // 胜局赢钱金额
	Loses             int64   `bson:"loses"`                                    // 败局输钱金额
	Cash              int64   `bson:"cash"`                                     // 净赢
	WinEscapeMultiple float64 `json:"-" bson:"-"`                               // 胜局平均逃脱倍数
	WinMulpitleSum    float64 `bson:"win_mulpitle_sum"`                         // 胜局逃跑倍数和
	AD_BundleId       string  `bson:"ad__bundle_id"`                            // 渠道
	Channel1          string  `bson:"channel1"`                                 // 渠道别名
	RegistArea        int     `bson:"regist_area"`                              // 账号类型 ab测试 0:A 1:B 2:C
	AllRounds         int32   `bson:"all_rounds"`                               // 总局数
	RoundBetAvg       int64   `bson:"round_bet_avg"`                            // 局均码
}

// 起死回生
type CrashQSHSStat struct {
	Id                    string  `bson:"_id"` // date-userid
	Date                  int64   `bson:"date"`
	DateStr               string  `bson:"date_str"`
	Userid                string  `bson:"userid"`
	AllinNumber           int64   `json:"allinNumber" bson:"allin_number"`                       // allin局数
	AllinBets             int64   `json:"allinBets" bson:"allin_bets"`                           // allin局打码量
	AllEvoTimes           int     `json:"allEvoTimes" bson:"all_evo_times"`                      // 触发策略局数
	TiggerBets            int64   `json:"tiggerBets" bson:"tigger_bets"`                         // 触发策略局的打码量
	WinRounds             int32   `bson:"win_rounds"`                                            // 触发策略时逃跑的局数
	LoseRounds            int     `json:"loseRounds" bson:"lose_rounds"`                         // 触发策略时未能逃跑的局数
	TiggerWinMulpitleSum  float64 `json:"tiggerWinMulpitleSum" bson:"tigger_win_mulpitle_sum"`   // 策略中逃跑倍数和
	MaximumEscapeMultiple float64 `json:"maximumEscapeMultiple" bson:"maximum_escape_multiple"`  // 策略中最高逃跑倍数
	LoseMaxEscapeMultiple float64 `json:"loseMaxEscapeMultiple" bson:"lose_max_escape_multiple"` // 触发后未能逃跑局的最高爆炸倍数
	LoseEscapeMultiple    float64 `json:"loseEscapeMultiple" bson:"lose_escape_multiple"`        // 触发后未能逃跑局的爆炸倍和
	Bets                  int64   `bson:"bets"`                                                  // 总打码量
	AllRounds             int32   `bson:"all_rounds"`                                            // 总局数
	AllWinBets            int64   `json:"allWinBets" bson:"all_win_bets"`                        // 总赢
	AllLoseBets           int64   `json:"allLoseBets" bson:"all_lose_bets"`                      // 总输
	// BetRounds             int32   `bson:"bet_rounds"`                                            // 下注局数
	WinBets        int64  `bson:"win_bets"`                  // 赢局打码量
	LoseBets       int64  `json:"loseBets" bson:"lose_bets"` // 输局打码量
	Wins           int64  `bson:"wins"`                      // 胜局赢钱金额
	Loses          int64  `bson:"loses"`                     // 败局输钱金额
	Cash           int64  `bson:"cash"`                      // 净赢
	WithdrawAmount int64  `json:"withdrawAmount" bson:"withdraw_amount"`
	PayAmount      int64  `json:"payAmount" bson:"pay_amount"`
	CarryAmount    int64  `json:"carryAmount" bson:"carry_amount"`
	AD_BundleId    string `bson:"ad__bundle_id"` // 渠道
	Channel1       string `bson:"channel1"`      // 渠道别名
	RegistArea     int    `bson:"regist_area"`   // 账号类型 ab测试 0:A 1:B 2:C
	RoundBetAvg    int64  `bson:"round_bet_avg"` // 局均码
}

// 奖池风控
type CrashJCFKStat struct {
	Id                    string  `bson:"_id"` // date-userid
	Date                  int64   `bson:"date"`
	DateStr               string  `bson:"date_str"`
	Userid                string  `bson:"userid"`
	TriggerTimes          int     `json:"triggerTimes" bson:"trigger_times"`                    // jackpot次数
	TriggerAmount         int64   `json:"triggerAmount" bson:"trigger_amount"`                  // 获得jackpot金额
	AllEvoTimes           int64   `json:"allEvoTimes" bson:"all_evo_times"`                     // 触发策略局数
	TiggerBets            int64   `json:"tiggerBets" bson:"tigger_bets"`                        // 触发策略局的打码量
	MaximumEscapeMultiple float64 `json:"maximumEscapeMultiple" bson:"maximum_escape_multiple"` // 触发后最高爆炸倍数
	TiggerMulpitleSum     float64 `json:"tiggerMulpitleSum" bson:"tigger_mulpitle_sum"`         // 触发后爆炸倍数和
	WinRounds             int32   `bson:"win_rounds"`                                           // 触发策略时逃跑的局数
	WinMulpitleSum        float64 `json:"winMulpitleSum" bson:"win_mulpitle_sum"`               // 触发后逃跑倍数和
	LoseRounds            int     `json:"loseRounds" bson:"lose_rounds"`                        // 触发策略时未能逃跑的局数
	LoseMulpitleSum       float64 `json:"loseMulpitleSum" bson:"lose_mulpitle_sum"`             // 触发后未能逃跑局的爆炸倍数和
	TriggerRounds         int     `json:"triggerRounds" bson:"trigger_rounds"`                  // 触发后获得jackpot局数
	WinBets               int64   `bson:"win_bets"`                                             // 赢局打码量
	LoseBets              int64   `json:"loseBets" bson:"lose_bets"`                            // 输局打码量
	Wins                  int64   `bson:"wins"`                                                 // 胜局赢钱金额
	Loses                 int64   `bson:"loses"`                                                // 败局输钱金额
	Cash                  int64   `bson:"cash"`                                                 // 净赢
	Bets                  int64   `bson:"bets"`                                                 // 总打码量
	AllRounds             int32   `bson:"all_rounds"`                                           // 总局数
	AllWinBets            int64   `json:"allWinBets" bson:"all_win_bets"`                       // 总赢
	AllLoseBets           int64   `json:"allLoseBets" bson:"all_lose_bets"`                     // 总输
	WithdrawAmount        int64   `json:"withdrawAmount" bson:"withdraw_amount"`
	PayAmount             int64   `json:"payAmount" bson:"pay_amount"`
	CarryAmount           int64   `json:"carryAmount" bson:"carry_amount"`
	AD_BundleId           string  `bson:"ad__bundle_id"` // 渠道
	Channel1              string  `bson:"channel1"`      // 渠道别名
	RegistArea            int     `bson:"regist_area"`   // 账号类型 ab测试 0:A 1:B 2:C
	RoundBetAvg           int64   `bson:"round_bet_avg"` // 局均码
}

// 冒险奖励
type CrashMXJLStat struct {
	Id                    string  `bson:"_id"` // date-userid
	Date                  int64   `bson:"date"`
	DateStr               string  `bson:"date_str"`
	Userid                string  `bson:"userid"`
	Bets                  int64   `bson:"bets"`                                                 // 总打码量
	AllRounds             int32   `bson:"all_rounds"`                                           // 总局数
	AllWinRounds          int32   `bson:"all_win_rounds"`                                       // 总赢局数
	AllWinBets            int64   `json:"allWinBets" bson:"all_win_bets"`                       // 总赢
	AllLoseBets           int64   `json:"allLoseBets" bson:"all_lose_bets"`                     // 总输
	AllEvoTimes           int64   `json:"allEvoTimes" bson:"all_evo_times"`                     // 触发策略局数
	TiggerBets            int64   `json:"tiggerBets" bson:"tigger_bets"`                        // 触发策略局的打码量
	MaximumEscapeMultiple float64 `json:"maximumEscapeMultiple" bson:"maximum_escape_multiple"` // 触发后最高爆炸倍数
	TiggerMulpitleSum     float64 `json:"tiggerMulpitleSum" bson:"tigger_mulpitle_sum"`         // 触发后爆炸倍数和
	WinRounds             int32   `bson:"win_rounds"`                                           // 触发策略时逃跑的局数
	WinMaxMulpitle        float64 `json:"winMaxMulpitle" bson:"win_max_mulpitle"`               // 触发后最高逃跑倍数
	WinMulpitleSum        float64 `json:"winMulpitleSum" bson:"win_mulpitle_sum"`               // 触发后逃跑倍数和
	LoseRounds            int     `json:"loseRounds" bson:"lose_rounds"`                        // 触发策略时未能逃跑的局数
	LoseMulpitleSum       float64 `json:"loseMulpitleSum" bson:"lose_mulpitle_sum"`             // 触发后未能逃跑局的爆炸倍数和
	WinBets               int64   `bson:"win_bets"`                                             // 策略赢局打码量
	LoseBets              int64   `json:"loseBets" bson:"lose_bets"`                            // 策略输局打码量
	Wins                  int64   `bson:"wins"`                                                 // 胜局赢钱金额
	Loses                 int64   `bson:"loses"`                                                // 败局输钱金额
	Cash                  int64   `bson:"cash"`                                                 // 策略净赢
	WithdrawAmount        int64   `json:"withdrawAmount" bson:"withdraw_amount"`
	PayAmount             int64   `json:"payAmount" bson:"pay_amount"`
	CarryAmount           int64   `json:"carryAmount" bson:"carry_amount"`
	AD_BundleId           string  `bson:"ad__bundle_id"` // 渠道
	Channel1              string  `bson:"channel1"`      // 渠道别名
	RegistArea            int     `bson:"regist_area"`   // 账号类型 ab测试 0:A 1:B 2:C
	RoundBetAvg           int64   `bson:"round_bet_avg"` // 局均码
}

// 人狂有祸
type CrashRKYSStat struct {
	Id                    string  `bson:"_id"` // date-userid
	Date                  int64   `bson:"date"`
	DateStr               string  `bson:"date_str"`
	Userid                string  `bson:"userid"`
	Bets                  int64   `bson:"bets"`                                                 // 总打码量
	AllRounds             int32   `bson:"all_rounds"`                                           // 总局数
	AllWinBets            int64   `json:"allWinBets" bson:"all_win_bets"`                       // 总赢
	AllLoseBets           int64   `json:"allLoseBets" bson:"all_lose_bets"`                     // 总输
	AllEvoTimes           int64   `json:"allEvoTimes" bson:"all_evo_times"`                     // 触发策略局数
	TiggerBets            int64   `json:"tiggerBets" bson:"tigger_bets"`                        // 触发策略局的打码量
	MaximumEscapeMultiple float64 `json:"maximumEscapeMultiple" bson:"maximum_escape_multiple"` // 触发后最高爆炸倍数
	TiggerMulpitleSum     float64 `json:"tiggerMulpitleSum" bson:"tigger_mulpitle_sum"`         // 触发后爆炸倍数和
	WinRounds             int32   `bson:"win_rounds"`                                           // 触发策略时逃跑的局数
	WinMaxMulpitle        float64 `json:"winMaxMulpitle" bson:"win_max_mulpitle"`               // 触发后最高逃跑倍数
	WinMulpitleSum        float64 `json:"winMulpitleSum" bson:"win_mulpitle_sum"`               // 触发后逃跑倍数和
	LoseRounds            int     `json:"loseRounds" bson:"lose_rounds"`                        // 触发策略时未能逃跑的局数
	LoseMulpitleSum       float64 `json:"loseMulpitleSum" bson:"lose_mulpitle_sum"`             // 触发后未能逃跑局的爆炸倍数和
	RFRounds              int64   `json:"rfRounds" bson:"rf_rounds"`                            // 肥码收割局数
	RFBets                int64   `json:"rfBets" bson:"rf_bets"`                                // 肥码收割局总码量
	RFLoseRounds          int64   `json:"rfLoseRounds" bson:"rf_lose_rounds"`                   // 肥码局输
	RFReapBets            int64   `json:"rfReapBets" bson:"rf_reap_bets"`                       // 肥码收割金额
	WinBets               int64   `bson:"win_bets"`                                             // 策略赢局打码量
	LoseBets              int64   `json:"loseBets" bson:"lose_bets"`                            // 策略输局打码量
	Wins                  int64   `bson:"wins"`                                                 // 胜局赢钱金额
	Loses                 int64   `bson:"loses"`                                                // 败局输钱金额
	Cash                  int64   `bson:"cash"`                                                 // 策略净赢
	WithdrawAmount        int64   `json:"withdrawAmount" bson:"withdraw_amount"`
	PayAmount             int64   `json:"payAmount" bson:"pay_amount"`
	CarryAmount           int64   `json:"carryAmount" bson:"carry_amount"`
	AD_BundleId           string  `bson:"ad__bundle_id"` // 渠道
	Channel1              string  `bson:"channel1"`      // 渠道别名
	RegistArea            int     `bson:"regist_area"`   // 账号类型 ab测试 0:A 1:B 2:C
	RoundBetAvg           int64   `bson:"round_bet_avg"` // 局均码
}

type LHDPlayerStat struct {
	Id               string `bson:"_id"` // date-userid
	Date             int64  `bson:"date"`
	DateStr          string `bson:"date_str"`
	Userid           string `bson:"userid"`
	NewReg           bool   `bson:"new_reg"`            // 是新顾客
	AllRounds        int32  `bson:"all_rounds"`         // 总局数
	GameTimes        int64  `bson:"game_times"`         // 游戏总时长(秒)
	Bets             int64  `bson:"bets"`               // 总打码量
	BetRounds        int32  `bson:"bet_rounds"`         // 下注局数
	WinRounds        int32  `bson:"win_rounds"`         // 胜利局数
	LoseRounds       int32  `bson:"lose_rounds"`        // 失败局数
	TieRounds        int32  `bson:"tie_rounds"`         // 和局数
	WinBets          int64  `bson:"win_bets"`           // 胜局打码量
	LoseBets         int64  `bson:"lose_bets"`          // 败局打码量
	TieBets          int64  `bson:"tie_bets"`           // 和局打码量
	Wins             int64  `bson:"wins"`               // 胜局赢钱金额
	Loses            int64  `bson:"loses"`              // 败局输钱金额
	Cash             int64  `bson:"cash"`               // 净赢
	Dragons          int64  `bson:"dragons"`            // 下注局开龙数
	Tigers           int64  `bson:"tigers"`             // 下注局开虎数
	Ties             int64  `bson:"ties"`               // 下注局开和数
	PlayerDragons    int64  `bson:"player_dragons"`     // 玩家下龙局数
	PlayerTigers     int64  `bson:"player_tigers"`      // 玩家下虎局数
	PlayerTies       int64  `bson:"player_ties"`        // 玩家下和局数
	PlayerWinDragons int64  `bson:"player_win_dragons"` // 玩家下龙赢的局数
	PlayerWinTigers  int64  `bson:"player_win_tigers"`  // 玩家下虎赢的局数
	PlayerWinTies    int64  `bson:"player_win_ties"`    // 玩家下和赢的局数
	PlayerMultis     int64  `bson:"player_multis"`      // 多门局数 玩家下注2门及以上的局数
	RoundBetAvg      int64  `bson:"round_bet_avg"`      // 局均码
	AD_BundleId      string `bson:"ad__bundle_id"`      // 渠道
	Channel1         string `bson:"channel1"`           // 渠道别名
	RegistArea       int    `bson:"regist_area"`        // 账号类型 ab测试 0:A 1:B 2:C
	Money            uint32 `bson:"money"`              // 充值总金额(分)
	ObserveRounds    int32  `bson:"observe_rounds"`     // 观察局数
}

// lhd 心想事成统计
type LHDXxscStat struct {
	Id                    string `bson:"_id"` // date-userid
	Date                  int64  `bson:"date"`
	DateStr               string `bson:"date_str"`
	Userid                string `bson:"userid"`
	StrategyPlayers       int32  `bson:"strategy_players"`        // 触发策略人数
	StrategyTimes         int32  `bson:"-"`                       // 触发策略总次数: u总 LHDXXSC.MaxTriggerTimes
	StrategyValidTimes    int32  `bson:"-"`                       // 触发策略有效次数: u LHDXXSC.TriggerTimes
	StrategyRounds        int32  `bson:"strategy_rounds"`         // 触发策略局数 r总
	StrategyDisturbRounds int32  `bson:"strategy_disturb_rounds"` // 扰动局数
	StrategyMultiRounds   int32  `bson:"strategy_multi_rounds"`   // 策略中多门局数
	StrategyBets          int64  `bson:"strategy_bets"`           // 策略中总打码量
	StrategyWinRounds     int64  `bson:"strategy_win_rounds"`     // 策略中总赢局数
	StrategyWins          int64  `bson:"strategy_wins"`           // 策略中总赢
	StrategyLoses         int64  `bson:"strategy_loses"`          // 策略中总输
	StrategyCash          int64  `bson:"-"`                       // 策略中净赢: 赢-输
	AD_BundleId           string `bson:"ad__bundle_id"`           // 渠道
	Channel1              string `bson:"channel1"`                // 渠道别名
	RegistArea            int    `bson:"regist_area"`             // 账号类型 ab测试 0:A 1:B 2:C
	Money                 uint32 `bson:"money"`                   // 充值总金额(分)
	// StrategyBetsAvg       int64  // 策略中局均打码量
}

// lhd 求死不能统计
type LHDQsbnStat struct {
	Id                  string `bson:"_id"` // date-userid
	Date                int64  `bson:"date"`
	DateStr             string `bson:"date_str"`
	Userid              string `bson:"userid"`
	StrategyPlayers     int32  `bson:"strategy_players"`      // 触发策略人数
	StrategyTimes       int32  `bson:"strategy_times"`        // 触发策略总次数
	AllInRounds         int32  `bson:"all_in_rounds"`         // allin局数
	StrategyMultiRounds int32  `bson:"strategy_multi_rounds"` // 策略中多门局数
	StrategyBets        int64  `bson:"strategy_bets"`         // 策略中总打码量
	StrategyWins        int64  `bson:"strategy_wins"`         // 策略中总赢
	StrategyLoses       int64  `bson:"strategy_loses"`        // 策略中总输
	StrategyCash        int64  `bson:"-"`                     // 策略中净赢: 赢-输
	BeforeBackRate      int    `bson:"before_back_rate"`      // 账变前返奖率
	AfterBackRate       int    `bson:"after_back_rate"`       // 账变后返奖率
	BeforeBackRateSum   int    `bson:"-"`                     // 账变前返奖率
	AfterBackRateSum    int    `bson:"-"`                     // 账变后返奖率
	BeforeBackRateCount int    `bson:"-"`                     // 账变前返奖率
	AfterBackRateCount  int    `bson:"-"`                     // 账变后返奖率
	AD_BundleId         string `bson:"ad__bundle_id"`         // 渠道
	Channel1            string `bson:"channel1"`              // 渠道别名
	RegistArea          int    `bson:"regist_area"`           // 账号类型 ab测试 0:A 1:B 2:C
	Money               uint32 `bson:"money"`                 // 充值总金额(分)
}

// 龙狂有祸
type LHDLkyhStat struct {
	Id                  string `bson:"_id"` // date-userid
	Date                int64  `bson:"date"`
	DateStr             string `bson:"date_str"`
	Userid              string `bson:"userid"`
	BSTimes             int    `json:"bstimes" bson:"bstimes"`                         // 倍杀状态次数
	BSDays              int    `json:"bsDays" bson:"bs_days"`                          // 倍杀天数
	TriggerTimes        int64  `json:"triggerTimes" bson:"trigger_times"`              // 每日倍杀次数 t日
	TZ                  int64  `json:"tz" bson:"tz"`                                   // 总倍杀次数 T总
	NZ                  int    `json:"nz" bson:"nz"`                                   // 总倍杀局数
	BSBets              int64  `json:"bsBets" bson:"bs_bets"`                          // 倍杀局打码量
	YZNumber            int64  `json:"yzNumber" bson:"yz_number"`                      // 压制次数
	YZRounds            int64  `json:"yzRounds" bson:"yz_rounds"`                      // 压制局数
	YZBets              int64  `json:"yzBets" bson:"yz_bets"`                          // 压制打码量
	RZ                  int    `json:"rz" bson:"rz"`                                   // r策
	StrategyMultiRounds int32  `bson:"strategy_multi_rounds"`                          // 策略中多门局数
	Bets                int64  `json:"bets" bson:"bets"`                               // 总打码量
	Rounds              int64  `json:"rounds" bson:"rounds"`                           // 总局数
	WinRounds           int64  `json:"winRounds" bson:"win_rounds"`                    // 赢局
	WinBets             int64  `json:"winBets" bson:"win_bets"`                        // 总赢
	LoseRounds          int64  `json:"loseRounds" bson:"lose_rounds"`                  // 输局
	LoseBets            int64  `json:"loseBets" bson:"lose_bets"`                      // 总输
	StrategyTimes       int32  `bson:"strategy_times"`                                 // 触发策略总次数
	StrategyBets        int64  `bson:"strategy_bets"`                                  // 策略中总打码量
	StrategyWins        int64  `bson:"strategy_wins"`                                  // 策略中总赢
	StrategyWinRounds   int64  `json:"strategyWinRounds" bson:"strategy_win_rounds"`   // 策略局赢局
	StrategyLoses       int64  `bson:"strategy_loses"`                                 // 策略中总输
	StrategyLoseRounds  int64  `json:"strategyLoseRounds" bson:"strategy_lose_rounds"` // 策略局输局
	StrategyCash        int64  `bson:"-"`                                              // 策略中净赢: 赢-输
	AD_BundleId         string `bson:"ad__bundle_id"`                                  // 渠道
	Channel1            string `bson:"channel1"`                                       // 渠道别名
	RegistArea          int    `bson:"regist_area"`                                    // 账号类型 ab测试 0:A 1:B 2:C
	Money               uint32 `bson:"money"`                                          // 充值总金额(分)
	RoundBetAvg         int64  `bson:"round_bet_avg"`                                  // 局均码
}

type RankingList struct {
	Id             string `json:"id" bson:"_id"`
	Date           int64  `bson:"date"`
	No             int    `json:"no" bson:"no"`                  // 排名
	RType          int    `json:"rType" bson:"r_type"`           // 排行榜类型 0:付费；1：提现;
	Userid         string `json:"userid" bson:"userid"`          // 用户id
	RegistArea     int    `json:"registArea" bson:"regist_area"` // ab测试 0:A 1:B
	PayAmount      int64  `json:"payAmount" bson:"pay_amount"`
	WithdrawAmount int64  `json:"withdrawAmount" bson:"withdraw_amount"`
	CarryAmount    int64  `json:"carryAmount" bson:"carry_amount"`
	ProfitAmount   int64  `json:"profitAmount" bson:"profit_amount"`
	ETime          int64  `json:"eTime" bson:"e_time"` // 刷新时间
}

// TP游戏统计
type TPPlayerStat struct {
	Id            string `bson:"_id"` // date-userid
	Date          int64  `bson:"date"`
	DateStr       string `bson:"date_str"`
	Userid        string `bson:"userid"`
	NewReg        bool   `bson:"new_reg"`        // 是新顾客
	AllRounds     int32  `bson:"all_rounds"`     // 总局数
	GameTimes     int64  `bson:"game_times"`     // 游戏总时长(秒)
	Bets          int64  `bson:"bets"`           // 总打码量
	BetRounds     int32  `bson:"bet_rounds"`     // 下注局数
	WinRounds     int32  `bson:"win_rounds"`     // 胜利局数
	LoseRounds    int32  `bson:"lose_rounds"`    // 失败局数
	TieRounds     int32  `bson:"tie_rounds"`     // 和局数
	WinBets       int64  `bson:"win_bets"`       // 胜局打码量
	LoseBets      int64  `bson:"lose_bets"`      // 败局打码量
	TieBets       int64  `bson:"tie_bets"`       // 和局打码量
	Wins          int64  `bson:"wins"`           // 胜局赢钱金额
	Loses         int64  `bson:"loses"`          // 败局输钱金额
	Cash          int64  `bson:"cash"`           // 净赢
	RoundBetAvg   int64  `bson:"round_bet_avg"`  // 局均码
	AD_BundleId   string `bson:"ad__bundle_id"`  // 渠道
	Channel1      string `bson:"channel1"`       // 渠道别名
	RegistArea    int    `bson:"regist_area"`    // 账号类型 ab测试 0:A 1:B 2:C
	Money         uint32 `bson:"money"`          // 充值总金额(分)
	ObserveRounds int32  `bson:"observe_rounds"` // 观察局数
}

// TP房间统计
type TPRoomStat struct {
	Id                     string `bson:"_id"` // date-userid
	Date                   int64  `bson:"date"`
	DateStr                string `bson:"date_str"`
	Userid                 string `bson:"userid"`
	Roomid                 string `json:"roomid" bson:"roomid"`                                     // 房间ID
	AllRounds              int64  `json:"allRounds" bson:"all_rounds"`                              // 局数
	Bets                   int64  `bson:"bets"`                                                     // 总下注
	WinRounds              int32  `bson:"win_rounds"`                                               // 胜利局数
	LoseRounds             int32  `bson:"lose_rounds"`                                              // 失败局数
	WinBets                int64  `bson:"win_bets"`                                                 // 胜局打码量
	LoseBets               int64  `bson:"lose_bets"`                                                // 败局打码量
	GameTime               int64  `json:"gameTime" bson:"game_time"`                                // 游戏时长
	UpNumber               int64  `json:"upNumber" bson:"up_number"`                                // 主动升到此场次数
	DownNumber             int64  `json:"downNumber" bson:"down_number"`                            // 主动降到此场次数
	TwoRounds              int64  `json:"twoRounds" bson:"two_rounds"`                              // 两人局
	WinTwoRounds           int64  `json:"winTwoRounds" bson:"win_two_rounds"`                       // 两人赢局
	TwoBets                int64  `json:"twoBets" bson:"two_bets"`                                  // 两人局总下注
	WinTwoBets             int64  `json:"winTwoBets" bson:"win_two_bets"`                           // 两人局总返奖
	ThreeRounds            int64  `json:"threeRounds" bson:"three_rounds"`                          // 三人局
	WinThreeRounds         int64  `json:"winThreeRounds" bson:"win_three_rounds"`                   // 三人赢局
	ThreeBets              int64  `json:"threeBets" bson:"three_bets"`                              // 三人局总下注
	WinThreeBets           int64  `json:"winThreeBets" bson:"win_three_bets"`                       // 三人局总返奖
	FourRounds             int64  `json:"fourRounds" bson:"four_rounds"`                            // 四人局
	WinFourRounds          int64  `json:"winFourRounds" bson:"win_four_rounds"`                     // 四人赢局
	FourBets               int64  `json:"fourBets" bson:"four_bets"`                                // 四人局总下注
	WinFourBets            int64  `json:"winFourBets" bson:"win_four_bets"`                         // 四人局总返奖
	FiveRounds             int64  `json:"fiveRounds" bson:"five_rounds"`                            // 五人局
	WinFiveRounds          int64  `json:"winFiveRounds" bson:"win_five_rounds"`                     // 五人赢局
	FiveBets               int64  `json:"fiveBets" bson:"five_bets"`                                // 五人局总下注
	WinFiveBets            int64  `json:"winFiveBets" bson:"win_five_bets"`                         // 五人局总返奖
	ChipPool               int64  `json:"chipPool" bson:"chip_pool"`                                // 筹码池上限
	BZUpNumber             int64  `json:"bzUpNumber" bson:"bz_up_number"`                           // 大豹子局数
	BZUpRounds             int64  `json:"bzUpRounds" bson:"bz_up_rounds"`                           // 大豹子轮次
	BZUpBets               int64  `json:"bzUpBets" bson:"bz_up_bets"`                               // 大豹子下注
	BZUpWinRounds          int64  `json:"bzUpWinRounds" bson:"bz_up_win_rounds"`                    // 大豹子赢局
	BZUpLoseRounds         int64  `json:"bzUpLoseRounds" bson:"bz_up_lose_rounds"`                  // 大豹子输局
	BZUpTwoDiscardNum      int64  `json:"bzUpTwoDiscardNum" bson:"bz_up_two_discard_num"`           // 大豹子两人局弃牌次数
	BZUpTwoFollowNum       int64  `json:"bzUpTwoFollowNum" bson:"bz_up_two_follow_num"`             // 大豹子两人局跟注次数
	BZUpThreeDiscardNum    int64  `json:"bzUpThreeDiscardNum" bson:"bz_up_three_discard_num"`       // 大豹子三人局弃牌次数
	BZUpThreeFollowNum     int64  `json:"bzUpThreeFollowNum" bson:"bz_up_three_follow_num"`         // 大豹子三人局跟注次数
	BZUpFourDiscardNum     int64  `json:"bzUpFourDiscardNum" bson:"bz_up_four_discard_num"`         // 大豹子四人局弃牌次数
	BZUpFourFollowNum      int64  `json:"bzUpFourFollowNum" bson:"bz_up_four_follow_num"`           // 大豹子四人局跟注次数
	BZUpFiveDiscardNum     int64  `json:"bzUpFiveDiscardNum" bson:"bz_up_five_discard_num"`         // 大豹子五人局弃牌次数
	BZUpFiveFollowNum      int64  `json:"bzUpFiveFollowNum" bson:"bz_up_five_follow_num"`           // 大豹子五人局跟注次数
	BZDownNumber           int64  `json:"bzDownNumber" bson:"bz_down_number"`                       // 小豹子局数
	BZDownRounds           int64  `json:"bzDownRounds" bson:"bz_down_rounds"`                       // 小豹子轮次
	BZDownBets             int64  `json:"bzDownBets" bson:"bz_down_bets"`                           // 小豹子下注
	BZDownWinRounds        int64  `json:"bzDownWinRounds" bson:"bz_down_win_rounds"`                // 小豹子赢局
	BZDownLoseRounds       int64  `json:"bzDownLoseRounds" bson:"bz_down_lose_rounds"`              // 小豹子输局
	BZDownTwoDiscardNum    int64  `json:"bzDownTwoDiscardNum" bson:"bz_down_two_discard_num"`       // 小豹子两人局弃牌次数
	BZDownTwoFollowNum     int64  `json:"bzDownTwoFollowNum" bson:"bz_down_two_follow_num"`         // 小豹子两人局跟注次数
	BZDownThreeDiscardNum  int64  `json:"bzDownThreeDiscardNum" bson:"bz_down_three_discard_num"`   // 小豹子三人局弃牌次数
	BZDownThreeFollowNum   int64  `json:"bzDownThreeFollowNum" bson:"bz_down_three_follow_num"`     // 小豹子三人局跟注次数
	BZDownFourDiscardNum   int64  `json:"bzDownFourDiscardNum" bson:"bz_down_four_discard_num"`     // 小豹子四人局弃牌次数
	BZDownFourFollowNum    int64  `json:"bzDownFourFollowNum" bson:"bz_down_four_follow_num"`       // 小豹子四人局跟注次数
	BZDownFiveDiscardNum   int64  `json:"bzDownFiveDiscardNum" bson:"bz_down_five_discard_num"`     // 小豹子五人局弃牌次数
	BZDownFiveFollowNum    int64  `json:"bzDownFiveFollowNum" bson:"bz_down_five_follow_num"`       // 小豹子五人局跟注次数
	THSUpNumber            int64  `json:"thsUpNumber" bson:"ths_up_number"`                         // 大同花顺局数
	THSUpRounds            int64  `json:"thsUpRounds" bson:"ths_up_rounds"`                         // 大同花顺轮次
	THSUpBets              int64  `json:"thsUpBets" bson:"ths_up_bets"`                             // 大同花下注
	THSUpWinRounds         int64  `json:"thsUpWinRounds" bson:"ths_up_win_rounds"`                  // 大同花赢局
	THSUpLoseRounds        int64  `json:"thsUpLoseRounds" bson:"ths_up_lose_rounds"`                // 大同花输局
	THSUpTwoDiscardNum     int64  `json:"thsUpTwoDiscardNum" bson:"ths_up_two_discard_num"`         // 大同花两人局弃牌次数
	THSUpTwoFollowNum      int64  `json:"thsUpTwoFollowNum" bson:"ths_up_two_follow_num"`           // 大同花两人局跟注次数
	THSUpThreeDiscardNum   int64  `json:"thsUpThreeDiscardNum" bson:"ths_up_three_discard_num"`     // 大同花三人局弃牌次数
	THSUpThreeFollowNum    int64  `json:"thsUpThreeFollowNum" bson:"ths_up_three_follow_num"`       // 大同花三人局跟注次数
	THSUpFourDiscardNum    int64  `json:"thsUpFourDiscardNum" bson:"ths_up_four_discard_num"`       // 大同花四人局弃牌次数
	THSUpFourFollowNum     int64  `json:"thsUpFourFollowNum" bson:"ths_up_four_follow_num"`         // 大同花四人局跟注次数
	THSUpFiveDiscardNum    int64  `json:"thsUpFiveDiscardNum" bson:"ths_up_five_discard_num"`       // 大同花五人局弃牌次数
	THSUpFiveFollowNum     int64  `json:"thsUpFiveFollowNum" bson:"ths_up_five_follow_num"`         // 大同花五人局跟注次数
	THSDownNumber          int64  `json:"thsDownNumber" bson:"ths_down_number"`                     // 小同花顺局数
	THSDownRounds          int64  `json:"thsDownRounds" bson:"ths_down_rounds"`                     // 小同花顺轮次
	THSDownBets            int64  `json:"thsDownBets" bson:"ths_down_bets"`                         // 小同花下注
	THSDownWinRounds       int64  `json:"thsDownWinRounds" bson:"ths_down_win_rounds"`              // 小同花赢局
	THSDownLoseRounds      int64  `json:"thsDownLoseRounds" bson:"ths_down_lose_rounds"`            // 小同花输局
	THSDownTwoDiscardNum   int64  `json:"thsDownTwoDiscardNum" bson:"ths_down_two_discard_num"`     // 小同花两人局弃牌次数
	THSDownTwoFollowNum    int64  `json:"thsDownTwoFollowNum" bson:"ths_down_two_follow_num"`       // 小同花两人局跟注次数
	THSDownThreeDiscardNum int64  `json:"thsDownThreeDiscardNum" bson:"ths_down_three_discard_num"` // 小同花三人局弃牌次数
	THSDownThreeFollowNum  int64  `json:"thsDownThreeFollowNum" bson:"ths_down_three_follow_num"`   // 小同花三人局跟注次数
	THSDownFourDiscardNum  int64  `json:"thsDownFourDiscardNum" bson:"ths_down_four_discard_num"`   // 小同花四人局弃牌次数
	THSDownFourFollowNum   int64  `json:"thsDownFourFollowNum" bson:"ths_down_four_follow_num"`     // 小同花四人局跟注次数
	THSDownFiveDiscardNum  int64  `json:"thsDownFiveDiscardNum" bson:"ths_down_five_discard_num"`   // 小同花五人局弃牌次数
	THSDownFiveFollowNum   int64  `json:"thsDownFiveFollowNum" bson:"ths_down_five_follow_num"`     // 小同花五人局跟注次数
	DSZUpNumber            int64  `json:"dszUpNumber" bson:"dsz_up_number"`                         // 大顺子局数
	DSZUpRounds            int64  `json:"dszUpRounds" bson:"dsz_up_rounds"`                         // 大顺子轮次
	DSZUpBets              int64  `json:"dszUpBets" bson:"dsz_up_bets"`                             // 大顺子下注
	DSZUpWinRounds         int64  `json:"dszUpWinRounds" bson:"dsz_up_win_rounds"`                  // 大顺子赢局
	DSZUpLoseRounds        int64  `json:"dszUpLoseRounds" bson:"dsz_up_lose_rounds"`                // 大顺子输局
	DSZUpTwoDiscardNum     int64  `json:"dszUpTwoDiscardNum" bson:"dsz_up_two_discard_num"`         // 大顺子两人局弃牌次数
	DSZUpTwoFollowNum      int64  `json:"dszUpTwoFollowNum" bson:"dsz_up_two_follow_num"`           // 大顺子两人局跟注次数
	DSZUpThreeDiscardNum   int64  `json:"dszUpThreeDiscardNum" bson:"dsz_up_three_discard_num"`     // 大顺子三人局弃牌次数
	DSZUpThreeFollowNum    int64  `json:"dszUpThreeFollowNum" bson:"dsz_up_three_follow_num"`       // 大顺子三人局跟注次数
	DSZUpFourDiscardNum    int64  `json:"dszUpFourDiscardNum" bson:"dsz_up_four_discard_num"`       // 大顺子四人局弃牌次数
	DSZUpFourFollowNum     int64  `json:"dszUpFourFollowNum" bson:"dsz_up_four_follow_num"`         // 大顺子四人局跟注次数
	DSZUpFiveDiscardNum    int64  `json:"dszUpFiveDiscardNum" bson:"dsz_up_five_discard_num"`       // 大顺子五人局弃牌次数
	DSZUpFiveFollowNum     int64  `json:"dszUpFiveFollowNum" bson:"dsz_up_five_follow_num"`         // 大顺子五人局跟注次数
	DSZDownNumber          int64  `json:"dszDownNumber" bson:"dsz_down_number"`                     // 小顺子局数
	DSZDownRounds          int64  `json:"dszDownRounds" bson:"dsz_down_rounds"`                     // 小顺子轮次
	DSZDownBets            int64  `json:"dszDownBets" bson:"dsz_down_bets"`                         // 小顺子下注
	DSZDownWinRounds       int64  `json:"dszDownWinRounds" bson:"dsz_down_win_rounds"`              // 小顺子赢局
	DSZDownLoseRounds      int64  `json:"dszDownLoseRounds" bson:"dsz_down_lose_rounds"`            // 小顺子输局
	DSZDownTwoDiscardNum   int64  `json:"dszDownTwoDiscardNum" bson:"dsz_down_two_discard_num"`     // 小顺子两人局弃牌次数
	DSZDownTwoFollowNum    int64  `json:"dszDownTwoFollowNum" bson:"dsz_down_two_follow_num"`       // 小顺子两人局跟注次数
	DSZDownThreeDiscardNum int64  `json:"dszDownThreeDiscardNum" bson:"dsz_down_three_discard_num"` // 小顺子三人局弃牌次数
	DSZDownThreeFollowNum  int64  `json:"dszDownThreeFollowNum" bson:"dsz_down_three_follow_num"`   // 小顺子三人局跟注次数
	DSZDownFourDiscardNum  int64  `json:"dszDownFourDiscardNum" bson:"dsz_down_four_discard_num"`   // 小顺子四人局弃牌次数
	DSZDownFourFollowNum   int64  `json:"dszDownFourFollowNum" bson:"dsz_down_four_follow_num"`     // 小顺子四人局跟注次数
	DSZDownFiveDiscardNum  int64  `json:"dszDownFiveDiscardNum" bson:"dsz_down_five_discard_num"`   // 小顺子五人局弃牌次数
	DSZDownFiveFollowNum   int64  `json:"dszDownFiveFollowNum" bson:"dsz_down_five_follow_num"`     // 小顺子五人局跟注次数
	DTHUpNumber            int64  `json:"dthUpNumber" bson:"dth_up_number"`                         // 大同花局数
	DTHUpRounds            int64  `json:"dthUpRounds" bson:"dth_up_rounds"`                         // 大同花轮次
	DTHUpBets              int64  `json:"dthUpBets" bson:"dth_up_bets"`                             // 大同花下注
	DTHUpWinRounds         int64  `json:"dthUpWinRounds" bson:"dth_up_win_rounds"`                  // 大同花赢局
	DTHUpLoseRounds        int64  `json:"dthUpLoseRounds" bson:"dth_up_lose_rounds"`                // 大同花输局
	DTHUpTwoDiscardNum     int64  `json:"dthUpTwoDiscardNum" bson:"dth_up_two_discard_num"`         // 大同花两人局弃牌次数
	DTHUpTwoFollowNum      int64  `json:"dthUpTwoFollowNum" bson:"dth_up_two_follow_num"`           // 大同花两人局跟注次数
	DTHUpThreeDiscardNum   int64  `json:"dthUpThreeDiscardNum" bson:"dth_up_three_discard_num"`     // 大同花三人局弃牌次数
	DTHUpThreeFollowNum    int64  `json:"dthUpThreeFollowNum" bson:"dth_up_three_follow_num"`       // 大同花三人局跟注次数
	DTHUpFourDiscardNum    int64  `json:"dthUpFourDiscardNum" bson:"dth_up_four_discard_num"`       // 大同花四人局弃牌次数
	DTHUpFourFollowNum     int64  `json:"dthUpFourFollowNum" bson:"dth_up_four_follow_num"`         // 大同花四人局跟注次数
	DTHUpFiveDiscardNum    int64  `json:"dthUpFiveDiscardNum" bson:"dth_up_five_discard_num"`       // 大同花五人局弃牌次数
	DTHUpFiveFollowNum     int64  `json:"dthUpFiveFollowNum" bson:"dth_up_five_follow_num"`         // 大同花五人局跟注次数
	DTHDownNumber          int64  `json:"dthDownNumber" bson:"dth_down_number"`                     // 小同花局数
	DTHDownRounds          int64  `json:"dthDownRounds" bson:"dth_down_rounds"`                     // 小同花轮次
	DTHDownBets            int64  `json:"dthDownBets" bson:"dth_down_bets"`                         // 小同花下注
	DTHDownWinRounds       int64  `json:"dthDownWinRounds" bson:"dth_down_win_rounds"`              // 小同花赢局
	DTHDownLoseRounds      int64  `json:"dthDownLoseRounds" bson:"dth_down_lose_rounds"`            // 小同花输局
	DTHDownTwoDiscardNum   int64  `json:"dthDownTwoDiscardNum" bson:"dth_down_two_discard_num"`     // 小同花两人局弃牌次数
	DTHDownTwoFollowNum    int64  `json:"dthDownTwoFollowNum" bson:"dth_down_two_follow_num"`       // 小同花两人局跟注次数
	DTHDownThreeDiscardNum int64  `json:"dthDownThreeDiscardNum" bson:"dth_down_three_discard_num"` // 小同花三人局弃牌次数
	DTHDownThreeFollowNum  int64  `json:"dthDownThreeFollowNum" bson:"dth_down_three_follow_num"`   // 小同花三人局跟注次数
	DTHDownFourDiscardNum  int64  `json:"dthDownFourDiscardNum" bson:"dth_down_four_discard_num"`   // 小同花四人局弃牌次数
	DTHDownFourFollowNum   int64  `json:"dthDownFourFollowNum" bson:"dth_down_four_follow_num"`     // 小同花四人局跟注次数
	DTHDownFiveDiscardNum  int64  `json:"dthDownFiveDiscardNum" bson:"dth_down_five_discard_num"`   // 小同花五人局弃牌次数
	DTHDownFiveFollowNum   int64  `json:"dthDownFiveFollowNum" bson:"dth_down_five_follow_num"`     // 小同花五人局跟注次数
	DDZUpNumber            int64  `json:"ddzUpNumber" bson:"ddz_up_number"`                         // 大对子局数
	DDZUpRounds            int64  `json:"ddzUpRounds" bson:"ddz_up_rounds"`                         // 大对子轮次
	DDZUpBets              int64  `json:"ddzUpBets" bson:"ddz_up_bets"`                             // 大对子下注
	DDZUpWinRounds         int64  `json:"ddzUpWinRounds" bson:"ddz_up_win_rounds"`                  // 大对子赢局
	DDZUpLoseRounds        int64  `json:"ddzUpLoseRounds" bson:"ddz_up_lose_rounds"`                // 大对子输局
	DDZUpTwoDiscardNum     int64  `json:"ddzUpTwoDiscardNum" bson:"ddz_up_two_discard_num"`         // 大对子两人局弃牌次数
	DDZUpTwoFollowNum      int64  `json:"ddzUpTwoFollowNum" bson:"ddz_up_two_follow_num"`           // 大对子两人局跟注次数
	DDZUpThreeDiscardNum   int64  `json:"ddzUpThreeDiscardNum" bson:"ddz_up_three_discard_num"`     // 大对子三人局弃牌次数
	DDZUpThreeFollowNum    int64  `json:"ddzUpThreeFollowNum" bson:"ddz_up_three_follow_num"`       // 大对子三人局跟注次数
	DDZUpFourDiscardNum    int64  `json:"ddzUpFourDiscardNum" bson:"ddz_up_four_discard_num"`       // 大对子四人局弃牌次数
	DDZUpFourFollowNum     int64  `json:"ddzUpFourFollowNum" bson:"ddz_up_four_follow_num"`         // 大对子四人局跟注次数
	DDZUpFiveDiscardNum    int64  `json:"ddzUpFiveDiscardNum" bson:"ddz_up_five_discard_num"`       // 大对子五人局弃牌次数
	DDZUpFiveFollowNum     int64  `json:"ddzUpFiveFollowNum" bson:"ddz_up_five_follow_num"`         // 大对子五人局跟注次数
	DDZDownNumber          int64  `json:"ddzDownNumber" bson:"ddz_down_number"`                     // 小对子局数
	DDZDownRounds          int64  `json:"ddzDownRounds" bson:"ddz_down_rounds"`                     // 小对子轮次
	DDZDownBets            int64  `json:"ddzDownBets" bson:"ddz_down_bets"`                         // 小对子下注
	DDZDownWinRounds       int64  `json:"ddzDownWinRounds" bson:"ddz_down_win_rounds"`              // 小对子赢局
	DDZDownLoseRounds      int64  `json:"ddzDownLoseRounds" bson:"ddz_down_lose_rounds"`            // 小对子输局
	DDZDownTwoDiscardNum   int64  `json:"ddzDownTwoDiscardNum" bson:"ddz_down_two_discard_num"`     // 小对子两人局弃牌次数
	DDZDownTwoFollowNum    int64  `json:"ddzDownTwoFollowNum" bson:"ddz_down_two_follow_num"`       // 小对子两人局跟注次数
	DDZDownThreeDiscardNum int64  `json:"ddzDownThreeDiscardNum" bson:"ddz_down_three_discard_num"` // 小对子三人局弃牌次数
	DDZDownThreeFollowNum  int64  `json:"ddzDownThreeFollowNum" bson:"ddz_down_three_follow_num"`   // 小对子三人局跟注次数
	DDZDownFourDiscardNum  int64  `json:"ddzDownFourDiscardNum" bson:"ddz_down_four_discard_num"`   // 小对子四人局弃牌次数
	DDZDownFourFollowNum   int64  `json:"ddzDownFourFollowNum" bson:"ddz_down_four_follow_num"`     // 小对子四人局跟注次数
	DDZDownFiveDiscardNum  int64  `json:"ddzDownFiveDiscardNum" bson:"ddz_down_five_discard_num"`   // 小对子五人局弃牌次数
	DDZDownFiveFollowNum   int64  `json:"ddzDownFiveFollowNum" bson:"ddz_down_five_follow_num"`     // 小对子五人局跟注次数
	DGPUpNumber            int64  `json:"dgpUpNumber" bson:"dgp_up_number"`                         // 大高牌局数
	DGPUpRounds            int64  `json:"dgpUpRounds" bson:"dgp_up_rounds"`                         // 大高牌轮次
	DGPUpBets              int64  `json:"dgpUpBets" bson:"dgp_up_bets"`                             // 大高牌下注
	DGPUpWinRounds         int64  `json:"dgpUpWinRounds" bson:"dgp_up_win_rounds"`                  // 大高牌赢局
	DGPUpLoseRounds        int64  `json:"dgpUpLoseRounds" bson:"dgp_up_lose_rounds"`                // 大高牌输局
	DGPUpTwoDiscardNum     int64  `json:"dgpUpTwoDiscardNum" bson:"dgp_up_two_discard_num"`         // 大高牌两人局弃牌次数
	DGPUpTwoFollowNum      int64  `json:"dgpUpTwoFollowNum" bson:"dgp_up_two_follow_num"`           // 大高牌两人局跟注次数
	DGPUpThreeDiscardNum   int64  `json:"dgpUpThreeDiscardNum" bson:"dgp_up_three_discard_num"`     // 大高牌三人局弃牌次数
	DGPUpThreeFollowNum    int64  `json:"dgpUpThreeFollowNum" bson:"dgp_up_three_follow_num"`       // 大高牌三人局跟注次数
	DGPUpFourDiscardNum    int64  `json:"dgpUpFourDiscardNum" bson:"dgp_up_four_discard_num"`       // 大高牌四人局弃牌次数
	DGPUpFourFollowNum     int64  `json:"dgpUpFourFollowNum" bson:"dgp_up_four_follow_num"`         // 大高牌四人局跟注次数
	DGPUpFiveDiscardNum    int64  `json:"dgpUpFiveDiscardNum" bson:"dgp_up_five_discard_num"`       // 大高牌五人局弃牌次数
	DGPUpFiveFollowNum     int64  `json:"dgpUpFiveFollowNum" bson:"dgp_up_five_follow_num"`         // 大高牌五人局跟注次数
	DGPDownNumber          int64  `json:"dgpDownNumber" bson:"dgp_down_number"`                     // 小高牌局数
	DGPDownRounds          int64  `json:"dgpDownRounds" bson:"dgp_down_rounds"`                     // 小高牌轮次
	DGPDownBets            int64  `json:"dgpDownBets" bson:"dgp_down_bets"`                         // 小高牌下注
	DGPDownWinRounds       int64  `json:"dgpDownWinRounds" bson:"dgp_down_win_rounds"`              // 小高牌赢局
	DGPDownLoseRounds      int64  `json:"dgpDownLoseRounds" bson:"dgp_down_lose_rounds"`            // 小高牌输局
	DGPDownTwoDiscardNum   int64  `json:"dgpDownTwoDiscardNum" bson:"dgp_down_two_discard_num"`     // 小高牌两人局弃牌次数
	DGPDownTwoFollowNum    int64  `json:"dgpDownTwoFollowNum" bson:"dgp_down_two_follow_num"`       // 小高牌两人局跟注次数
	DGPDownThreeDiscardNum int64  `json:"dgpDownThreeDiscardNum" bson:"dgp_down_three_discard_num"` // 小高牌三人局弃牌次数
	DGPDownThreeFollowNum  int64  `json:"dgpDownThreeFollowNum" bson:"dgp_down_three_follow_num"`   // 小高牌三人局跟注次数
	DGPDownFourDiscardNum  int64  `json:"dgpDownFourDiscardNum" bson:"dgp_down_four_discard_num"`   // 小高牌四人局弃牌次数
	DGPDownFourFollowNum   int64  `json:"dgpDownFourFollowNum" bson:"dgp_down_four_follow_num"`     // 小高牌四人局跟注次数
	DGPDownFiveDiscardNum  int64  `json:"dgpDownFiveDiscardNum" bson:"dgp_down_five_discard_num"`   // 小高牌五人局弃牌次数
	DGPDownFiveFollowNum   int64  `json:"dgpDownFiveFollowNum" bson:"dgp_down_five_follow_num"`     // 小高牌五人局跟注次数
	AD_BundleId            string `bson:"ad__bundle_id"`                                            // 渠道
	Channel1               string `bson:"channel1"`                                                 // 渠道别名
	RegistArea             int    `bson:"regist_area"`                                              // 账号类型 ab测试 0:A 1:B 2:C
	Money                  uint32 `bson:"money"`                                                    // 充值总金额(分)
}

/*
	7Updown
*/

type UpdownPlayerStat struct {
	Id               string `bson:"_id"` // date-userid
	Date             int64  `bson:"date"`
	DateStr          string `bson:"date_str"`
	Userid           string `bson:"userid"`
	NewReg           bool   `bson:"new_reg"`            // 是新顾客
	AllRounds        int32  `bson:"all_rounds"`         // 总局数
	GameTimes        int64  `bson:"game_times"`         // 游戏总时长(秒)
	Bets             int64  `bson:"bets"`               // 总打码量
	BetRounds        int32  `bson:"bet_rounds"`         // 下注局数
	WinRounds        int32  `bson:"win_rounds"`         // 胜利局数
	LoseRounds       int32  `bson:"lose_rounds"`        // 失败局数
	TieRounds        int32  `bson:"tie_rounds"`         // 和局数
	WinBets          int64  `bson:"win_bets"`           // 胜局打码量
	LoseBets         int64  `bson:"lose_bets"`          // 败局打码量
	TieBets          int64  `bson:"tie_bets"`           // 和局打码量
	Wins             int64  `bson:"wins"`               // 胜局赢钱金额
	Loses            int64  `bson:"loses"`              // 败局输钱金额
	Cash             int64  `bson:"cash"`               // 净赢
	Dragons          int64  `bson:"dragons"`            // 下注局开龙数
	Tigers           int64  `bson:"tigers"`             // 下注局开虎数
	Ties             int64  `bson:"ties"`               // 下注局开和数
	PlayerDragons    int64  `bson:"player_dragons"`     // 玩家下龙局数
	PlayerTigers     int64  `bson:"player_tigers"`      // 玩家下虎局数
	PlayerTies       int64  `bson:"player_ties"`        // 玩家下和局数
	PlayerWinDragons int64  `bson:"player_win_dragons"` // 玩家下龙赢的局数
	PlayerWinTigers  int64  `bson:"player_win_tigers"`  // 玩家下虎赢的局数
	PlayerWinTies    int64  `bson:"player_win_ties"`    // 玩家下和赢的局数
	PlayerMultis     int64  `bson:"player_multis"`      // 多门局数 玩家下注2门及以上的局数
	RoundBetAvg      int64  `bson:"round_bet_avg"`      // 局均码
	AD_BundleId      string `bson:"ad__bundle_id"`      // 渠道
	Channel1         string `bson:"channel1"`           // 渠道别名
	RegistArea       int    `bson:"regist_area"`        // 账号类型 ab测试 0:A 1:B 2:C
	Money            uint32 `bson:"money"`              // 充值总金额(分)
	ObserveRounds    int32  `bson:"observe_rounds"`     // 观察局数
}

// 7Updown 心想事成统计
type UpdownXxscStat struct {
	Id                    string `bson:"_id"` // date-userid
	Date                  int64  `bson:"date"`
	DateStr               string `bson:"date_str"`
	Userid                string `bson:"userid"`
	StrategyPlayers       int32  `bson:"strategy_players"`        // 触发策略人数
	StrategyTimes         int32  `bson:"-"`                       // 触发策略总次数: u总 LHDXXSC.MaxTriggerTimes
	StrategyValidTimes    int32  `bson:"-"`                       // 触发策略有效次数: u LHDXXSC.TriggerTimes
	StrategyRounds        int32  `bson:"strategy_rounds"`         // 触发策略局数 r总
	StrategyDisturbRounds int32  `bson:"strategy_disturb_rounds"` // 扰动局数
	StrategyMultiRounds   int32  `bson:"strategy_multi_rounds"`   // 策略中多门局数
	StrategyBets          int64  `bson:"strategy_bets"`           // 策略中总打码量
	StrategyWinRounds     int64  `bson:"strategy_win_rounds"`     // 策略中总赢局数
	StrategyWins          int64  `bson:"strategy_wins"`           // 策略中总赢
	StrategyLoses         int64  `bson:"strategy_loses"`          // 策略中总输
	StrategyCash          int64  `bson:"-"`                       // 策略中净赢: 赢-输
	AD_BundleId           string `bson:"ad__bundle_id"`           // 渠道
	Channel1              string `bson:"channel1"`                // 渠道别名
	RegistArea            int    `bson:"regist_area"`             // 账号类型 ab测试 0:A 1:B 2:C
	Money                 uint32 `bson:"money"`                   // 充值总金额(分)
	// StrategyBetsAvg       int64  // 策略中局均打码量
}

// 7Updown 求死不能统计
type UpdownQsbnStat struct {
	Id                  string `bson:"_id"` // date-userid
	Date                int64  `bson:"date"`
	DateStr             string `bson:"date_str"`
	Userid              string `bson:"userid"`
	StrategyPlayers     int32  `bson:"strategy_players"`      // 触发策略人数
	StrategyTimes       int32  `bson:"strategy_times"`        // 触发策略总次数
	AllInRounds         int32  `bson:"all_in_rounds"`         // allin局数
	StrategyMultiRounds int32  `bson:"strategy_multi_rounds"` // 策略中多门局数
	StrategyBets        int64  `bson:"strategy_bets"`         // 策略中总打码量
	StrategyWins        int64  `bson:"strategy_wins"`         // 策略中总赢
	StrategyLoses       int64  `bson:"strategy_loses"`        // 策略中总输
	StrategyCash        int64  `bson:"-"`                     // 策略中净赢: 赢-输
	BeforeBackRate      int    `bson:"before_back_rate"`      // 账变前返奖率
	AfterBackRate       int    `bson:"after_back_rate"`       // 账变后返奖率
	BeforeBackRateSum   int    `bson:"-"`                     // 账变前返奖率
	AfterBackRateSum    int    `bson:"-"`                     // 账变后返奖率
	BeforeBackRateCount int    `bson:"-"`                     // 账变前返奖率
	AfterBackRateCount  int    `bson:"-"`                     // 账变后返奖率
	AD_BundleId         string `bson:"ad__bundle_id"`         // 渠道
	Channel1            string `bson:"channel1"`              // 渠道别名
	RegistArea          int    `bson:"regist_area"`           // 账号类型 ab测试 0:A 1:B 2:C
	Money               uint32 `bson:"money"`                 // 充值总金额(分)
}

// 7updown 龙狂有祸
type UpdownLkyhStat struct {
	Id                  string `bson:"_id"` // date-userid
	Date                int64  `bson:"date"`
	DateStr             string `bson:"date_str"`
	Userid              string `bson:"userid"`
	BSTimes             int    `json:"bstimes" bson:"bstimes"`                         // 倍杀状态次数
	BSDays              int    `json:"bsDays" bson:"bs_days"`                          // 倍杀天数
	TriggerTimes        int64  `json:"triggerTimes" bson:"trigger_times"`              // 每日倍杀次数 t日
	TZ                  int64  `json:"tz" bson:"tz"`                                   // 总倍杀次数 T总
	NZ                  int    `json:"nz" bson:"nz"`                                   // 总倍杀局数
	BSBets              int64  `json:"bsBets" bson:"bs_bets"`                          // 倍杀局打码量
	YZNumber            int64  `json:"yzNumber" bson:"yz_number"`                      // 压制次数
	YZRounds            int64  `json:"yzRounds" bson:"yz_rounds"`                      // 压制局数
	YZBets              int64  `json:"yzBets" bson:"yz_bets"`                          // 压制打码量
	RZ                  int    `json:"rz" bson:"rz"`                                   // r策
	StrategyMultiRounds int32  `bson:"strategy_multi_rounds"`                          // 策略中多门局数
	Bets                int64  `json:"bets" bson:"bets"`                               // 总打码量
	Rounds              int64  `json:"rounds" bson:"rounds"`                           // 总局数
	WinRounds           int64  `json:"winRounds" bson:"win_rounds"`                    // 赢局
	WinBets             int64  `json:"winBets" bson:"win_bets"`                        // 总赢
	LoseRounds          int64  `json:"loseRounds" bson:"lose_rounds"`                  // 输局
	LoseBets            int64  `json:"loseBets" bson:"lose_bets"`                      // 总输
	StrategyTimes       int32  `bson:"strategy_times"`                                 // 触发策略总次数
	StrategyBets        int64  `bson:"strategy_bets"`                                  // 策略中总打码量
	StrategyWins        int64  `bson:"strategy_wins"`                                  // 策略中总赢
	StrategyWinRounds   int64  `json:"strategyWinRounds" bson:"strategy_win_rounds"`   // 策略局赢局
	StrategyLoses       int64  `bson:"strategy_loses"`                                 // 策略中总输
	StrategyLoseRounds  int64  `json:"strategyLoseRounds" bson:"strategy_lose_rounds"` // 策略局输局
	StrategyCash        int64  `bson:"-"`                                              // 策略中净赢: 赢-输
	AD_BundleId         string `bson:"ad__bundle_id"`                                  // 渠道
	Channel1            string `bson:"channel1"`                                       // 渠道别名
	RegistArea          int    `bson:"regist_area"`                                    // 账号类型 ab测试 0:A 1:B 2:C
	Money               uint32 `bson:"money"`                                          // 充值总金额(分)
	RoundBetAvg         int64  `bson:"round_bet_avg"`                                  // 局均码
}

/*
AB 游戏统计
*/
type ABPlayerStat struct {
	Id              string `bson:"_id"` // date-userid
	Date            int64  `bson:"date"`
	DateStr         string `bson:"date_str"`
	Userid          string `bson:"userid"`
	NewReg          bool   `bson:"new_reg"`                                  // 是新顾客
	AllRounds       int32  `bson:"all_rounds"`                               // 总局数
	GameTimes       int64  `bson:"game_times"`                               // 游戏总时长(秒)
	Bets            int64  `bson:"bets"`                                     // 总打码量
	BetRounds       int32  `bson:"bet_rounds"`                               // 下注局数
	WinRounds       int32  `bson:"win_rounds"`                               // 胜利局数
	LoseRounds      int32  `bson:"lose_rounds"`                              // 失败局数
	TieRounds       int32  `bson:"tie_rounds"`                               // 和局数
	WinBets         int64  `bson:"win_bets"`                                 // 胜局打码量
	LoseBets        int64  `bson:"lose_bets"`                                // 败局打码量
	TieBets         int64  `bson:"tie_bets"`                                 // 和局打码量
	Wins            int64  `bson:"wins"`                                     // 胜局赢钱金额
	Loses           int64  `bson:"loses"`                                    // 败局输钱金额
	Cash            int64  `bson:"cash"`                                     // 净赢
	SideWinner1     int64  `json:"sideWinner1" bson:"side_winner1"`          // 开奖位置ANDAR局数
	SideWinner2     int64  `json:"sideWinner2" bson:"side_winner2"`          // BAHAR
	SideWinner3     int64  `json:"sideWinner3" bson:"side_winner3"`          // 1-5
	SideWinner4     int64  `json:"sideWinner4" bson:"side_winner4"`          // 6-10
	SideWinner5     int64  `json:"sideWinner5" bson:"side_winner5"`          // 11-15
	SideWinner6     int64  `json:"sideWinner6" bson:"side_winner6"`          // 16-25
	SideWinner7     int64  `json:"sideWinner7" bson:"side_winner7"`          // 26-30
	SideWinner8     int64  `json:"sideWinner8" bson:"side_winner8"`          // 31-35
	SideWinner9     int64  `json:"sideWinner9" bson:"side_winner9"`          // 36-40
	SideWinner10    int64  `json:"sideWinner10" bson:"side_winner10"`        // 41以上
	SeatBets1       int64  `json:"seatBets1" bson:"seat_bets1"`              // 玩家下注ANDAR
	SeatBets2       int64  `json:"seatBets2" bson:"seat_bets2"`              // BAHAR
	SeatBets3       int64  `json:"seatBets3" bson:"seat_bets3"`              // 1-5
	SeatBets4       int64  `json:"seatBets4" bson:"seat_bets4"`              // 6-10
	SeatBets5       int64  `json:"seatBets5" bson:"seat_bets5"`              // 11-15
	SeatBets6       int64  `json:"seatBets6" bson:"seat_bets6"`              // 16-25
	SeatBets7       int64  `json:"seatBets7" bson:"seat_bets7"`              // 26-30
	SeatBets8       int64  `json:"seatBets8" bson:"seat_bets8"`              // 31-35
	SeatBets9       int64  `json:"seatBets9" bson:"seat_bets9"`              // 36-40
	SeatBets10      int64  `json:"seatBets10" bson:"seat_bets10"`            // 41以上
	PlayerWinSeat1  int64  `json:"playerWinSeat1" bson:"player_win_seat1"`   // 玩家下注位置ANDAR赢的局数
	PlayerWinSeat2  int64  `json:"playerWinSeat2" bson:"player_win_seat2"`   // BAHAR
	PlayerWinSeat3  int64  `json:"playerWinSeat3" bson:"player_win_seat3"`   // 1-5
	PlayerWinSeat4  int64  `json:"playerWinSeat4" bson:"player_win_seat4"`   // 6-10
	PlayerWinSeat5  int64  `json:"playerWinSeat5" bson:"player_win_seat5"`   // 11-15
	PlayerWinSeat6  int64  `json:"playerWinSeat6" bson:"player_win_seat6"`   // 16-25
	PlayerWinSeat7  int64  `json:"playerWinSeat7" bson:"player_win_seat7"`   // 26-30
	PlayerWinSeat8  int64  `json:"playerWinSeat8" bson:"player_win_seat8"`   // 31-35
	PlayerWinSeat9  int64  `json:"playerWinSeat9" bson:"player_win_seat9"`   // 36-40
	PlayerWinSeat10 int64  `json:"playerWinSeat10" bson:"player_win_seat10"` // 41以上
	PlayerMultis    int64  `bson:"player_multis"`                            // 多门局数 玩家下注2门及以上的局数
	RoundBetAvg     int64  `bson:"round_bet_avg"`                            // 局均码
	AD_BundleId     string `bson:"ad__bundle_id"`                            // 渠道
	Channel1        string `bson:"channel1"`                                 // 渠道别名
	RegistArea      int    `bson:"regist_area"`                              // 账号类型 ab测试 0:A 1:B 2:C
	Money           uint32 `bson:"money"`                                    // 充值总金额(分)
	ObserveRounds   int32  `bson:"observe_rounds"`                           // 观察局数
}

// AB 安能求死
type ABAnqsStat struct {
	Id                  string `bson:"_id"` // date-userid
	Date                int64  `bson:"date"`
	DateStr             string `bson:"date_str"`
	Userid              string `bson:"userid"`
	StrategyPlayers     int32  `bson:"strategy_players"`                             // 触发策略人数
	StrategyTimes       int32  `bson:"strategy_times"`                               // 触发策略总次数
	AllInRounds         int32  `bson:"all_in_rounds"`                                // allin局数
	StrategyMultiRounds int32  `bson:"strategy_multi_rounds"`                        // 策略中多门局数
	StrategyBets        int64  `bson:"strategy_bets"`                                // 策略中总打码量
	StrategyWinRounds   int64  `json:"strategyWinRounds" bson:"strategy_win_rounds"` // 策略局中赢的局数
	StrategyWins        int64  `bson:"strategy_wins"`                                // 策略中总赢
	StrategyLoseRounds  int64  `json:"strategyLoseRounds" bson:"strategy_lose_rounds"`
	StrategyLoses       int64  `bson:"strategy_loses"`                        // 策略中总输
	StrategyCash        int64  `bson:"-"`                                     // 策略中净赢: 赢-输
	Rounds              int64  `json:"rounds" bson:"rounds"`                  // AB总局数
	WinRounds           int64  `json:"winRounds" bson:"win_rounds"`           // AB总赢局数
	StrategyRounds      int64  `json:"strategyRounds" bson:"strategy_rounds"` // 触发策略总局数
	Wins                int64  `json:"wins" bson:"wins"`                      // 总赢
	LoseRounds          int64  `json:"loseRounds" bson:"lose_rounds"`
	Loses               int64  `json:"loses" bson:"loses"` // 总输
	AD_BundleId         string `bson:"ad__bundle_id"`      // 渠道
	Channel1            string `bson:"channel1"`           // 渠道别名
	RegistArea          int    `bson:"regist_area"`        // 账号类型 ab测试 0:A 1:B 2:C
	Money               uint32 `bson:"money"`              // 充值总金额(分)

}

// AB玩家下注情况
type ABAnqsUserBet struct {
	Userid string `bson:"-"`
	Score  int64  `bson:"-"`
}

// AB 安然躺赢
type ABArtyStat struct {
	Id                  string `bson:"_id"` // date-userid
	Date                int64  `bson:"date"`
	DateStr             string `bson:"date_str"`
	Userid              string `bson:"userid"`
	StrategyPlayers     int32  `bson:"strategy_players"`                             // 触发策略人数
	StrategyTimes       int32  `bson:"strategy_times"`                               // 触发策略总次数
	StrategyRounds      int64  `json:"strategyRounds" bson:"strategy_rounds"`        // 触发策略总局数
	TiggerTimes         int    `json:"tiggerTimes" bson:"tigger_times"`              // u 触发策略有效次数
	AllEvoTimes         int    `json:"allEvoTimes" bson:"all_evo_times"`             // 触发策略局数 r总
	StrategyMultiRounds int32  `bson:"strategy_multi_rounds"`                        // 策略中多门局数
	StrategyBets        int64  `bson:"strategy_bets"`                                // 策略中总打码量
	StrategyWins        int64  `bson:"strategy_wins"`                                // 策略中总赢
	StrategyWinRounds   int64  `json:"strategyWinRounds" bson:"strategy_win_rounds"` // 策略局中赢的局数
	StrategyLoses       int64  `bson:"strategy_loses"`                               // 策略中总输
	StrategyLoseRounds  int64  `json:"strategyLoseRounds" bson:"strategy_lose_rounds"`
	StrategyCash        int64  `bson:"-"`             // 策略中净赢: 赢-输
	AD_BundleId         string `bson:"ad__bundle_id"` // 渠道
	Channel1            string `bson:"channel1"`      // 渠道别名
	RegistArea          int    `bson:"regist_area"`   // 账号类型 ab测试 0:A 1:B 2:C
	Money               uint32 `bson:"money"`         // 充值总金额(分)
}

/*
CP 游戏统计
*/
type CPPlayerStat struct {
	Id             string `bson:"_id"` // date-userid
	Date           int64  `bson:"date"`
	DateStr        string `bson:"date_str"`
	Userid         string `bson:"userid"`
	NewReg         bool   `bson:"new_reg"`                                // 是新顾客
	AllRounds      int32  `bson:"all_rounds"`                             // 总局数
	GameTimes      int64  `bson:"game_times"`                             // 游戏总时长(秒)
	Bets           int64  `bson:"bets"`                                   // 总打码量
	BetRounds      int32  `bson:"bet_rounds"`                             // 下注局数
	WinRounds      int32  `bson:"win_rounds"`                             // 胜利局数
	LoseRounds     int32  `bson:"lose_rounds"`                            // 失败局数
	TieRounds      int32  `bson:"tie_rounds"`                             // 和局数
	WinBets        int64  `bson:"win_bets"`                               // 胜局打码量
	LoseBets       int64  `bson:"lose_bets"`                              // 败局打码量
	TieBets        int64  `bson:"tie_bets"`                               // 和局打码量
	Wins           int64  `bson:"wins"`                                   // 胜局赢钱金额
	Loses          int64  `bson:"loses"`                                  // 败局输钱金额
	Cash           int64  `bson:"cash"`                                   // 净赢
	SideWinner1    int64  `json:"sideWinner1" bson:"side_winner1"`        // 开奖位置High局数
	SideWinner2    int64  `json:"sideWinner2" bson:"side_winner2"`        // Pair
	SideWinner3    int64  `json:"sideWinner3" bson:"side_winner3"`        // Color
	SideWinner4    int64  `json:"sideWinner4" bson:"side_winner4"`        // seq
	SideWinner5    int64  `json:"sideWinner5" bson:"side_winner5"`        // pureseq
	SideWinner6    int64  `json:"sideWinner6" bson:"side_winner6"`        // set
	SeatBets1      int64  `json:"seatBets1" bson:"seat_bets1"`            // 玩家下注High
	SeatBets2      int64  `json:"seatBets2" bson:"seat_bets2"`            // Pair
	SeatBets3      int64  `json:"seatBets3" bson:"seat_bets3"`            // Color
	SeatBets4      int64  `json:"seatBets4" bson:"seat_bets4"`            // seq
	SeatBets5      int64  `json:"seatBets5" bson:"seat_bets5"`            // pureseq
	SeatBets6      int64  `json:"seatBets6" bson:"seat_bets6"`            // set
	PlayerWinSeat1 int64  `json:"playerWinSeat1" bson:"player_win_seat1"` // 玩家下注位置High赢的局数
	PlayerWinSeat2 int64  `json:"playerWinSeat2" bson:"player_win_seat2"` // Pair
	PlayerWinSeat3 int64  `json:"playerWinSeat3" bson:"player_win_seat3"` // Color
	PlayerWinSeat4 int64  `json:"playerWinSeat4" bson:"player_win_seat4"` // seq
	PlayerWinSeat5 int64  `json:"playerWinSeat5" bson:"player_win_seat5"` // pureseq
	PlayerWinSeat6 int64  `json:"playerWinSeat6" bson:"player_win_seat6"` // set
	PlayerMultis   int64  `bson:"player_multis"`                          // 多门局数 玩家下注2门及以上的局数
	RoundBetAvg    int64  `bson:"round_bet_avg"`                          // 局均码
	AD_BundleId    string `bson:"ad__bundle_id"`                          // 渠道
	Channel1       string `bson:"channel1"`                               // 渠道别名
	RegistArea     int    `bson:"regist_area"`                            // 账号类型 ab测试 0:A 1:B 2:C
	Money          uint32 `bson:"money"`                                  // 充值总金额(分)
	ObserveRounds  int32  `bson:"observe_rounds"`                         // 观察局数
}

// CP 来玩就赢
type CPLwjyStat struct {
	Id                  string `bson:"_id"` // date-userid
	Date                int64  `bson:"date"`
	DateStr             string `bson:"date_str"`
	Userid              string `bson:"userid"`
	StrategyPlayers     int32  `bson:"strategy_players"`                             // 触发策略人数
	StrategyTimes       int32  `bson:"strategy_times"`                               // 触发策略总次数
	TiggerTimes         int    `json:"tiggerTimes" bson:"tigger_times"`              // u 触发策略有效次数
	AllEvoTimes         int    `json:"allEvoTimes" bson:"all_evo_times"`             // 触发策略局数 r总
	StrategyMultiRounds int32  `bson:"strategy_multi_rounds"`                        // 策略中多门局数
	StrategyBets        int64  `bson:"strategy_bets"`                                // 策略中总打码量
	StrategyWinRounds   int64  `json:"strategyWinRounds" bson:"strategy_win_rounds"` // 策略局中赢的局数
	StrategyWins        int64  `bson:"strategy_wins"`                                // 策略中总赢
	StrategyLoseRounds  int64  `json:"strategyLoseRounds" bson:"strategy_lose_rounds"`
	StrategyLoses       int64  `bson:"strategy_loses"` // 策略中总输
	StrategyCash        int64  `bson:"-"`              // 策略中净赢: 赢-输
	AD_BundleId         string `bson:"ad__bundle_id"`  // 渠道
	Channel1            string `bson:"channel1"`       // 渠道别名
	RegistArea          int    `bson:"regist_area"`    // 账号类型 ab测试 0:A 1:B 2:C
	Money               uint32 `bson:"money"`          // 充值总金额(分)

}

// CP 来易去难
type CPLyqnStat struct {
	Id                  string `bson:"_id"` // date-userid
	Date                int64  `bson:"date"`
	DateStr             string `bson:"date_str"`
	Userid              string `bson:"userid"`
	StrategyPlayers     int32  `bson:"strategy_players"`                             // 触发策略人数
	StrategyRounds      int64  `json:"strategyRounds" bson:"strategy_rounds"`        // 触发策略总局数
	AllInRounds         int32  `bson:"all_in_rounds"`                                // allin局数
	StrategyMultiRounds int32  `bson:"strategy_multi_rounds"`                        // 策略中多门局数
	StrategyBets        int64  `bson:"strategy_bets"`                                // 策略中总打码量
	StrategyWins        int64  `bson:"strategy_wins"`                                // 策略中总赢
	StrategyWinRounds   int64  `json:"strategyWinRounds" bson:"strategy_win_rounds"` // 策略局中赢的局数
	StrategyLoses       int64  `bson:"strategy_loses"`                               // 策略中总输
	StrategyLoseRounds  int64  `json:"strategyLoseRounds" bson:"strategy_lose_rounds"`
	StrategyCash        int64  `bson:"-"`             // 策略中净赢: 赢-输
	BetRounds           int32  `bson:"bet_rounds"`    // 下注局数
	WinRounds           int32  `bson:"win_rounds"`    // 胜利局数
	LoseRounds          int32  `bson:"lose_rounds"`   // 失败局数
	Wins                int64  `bson:"wins"`          // 胜局赢钱金额
	Loses               int64  `bson:"loses"`         // 败局输钱金额
	AD_BundleId         string `bson:"ad__bundle_id"` // 渠道
	Channel1            string `bson:"channel1"`      // 渠道别名
	RegistArea          int    `bson:"regist_area"`   // 账号类型 ab测试 0:A 1:B 2:C
	Money               uint32 `bson:"money"`         // 充值总金额(分)
}

/*
RB
*/
type RBPlayerStat struct {
	Id             string `bson:"_id"` // date-userid
	Date           int64  `bson:"date"`
	DateStr        string `bson:"date_str"`
	Userid         string `bson:"userid"`
	NewReg         bool   `bson:"new_reg"`                                // 是新顾客
	AllRounds      int32  `bson:"all_rounds"`                             // 总局数
	GameTimes      int64  `bson:"game_times"`                             // 游戏总时长(秒)
	Bets           int64  `bson:"bets"`                                   // 总打码量
	BetRounds      int32  `bson:"bet_rounds"`                             // 下注局数
	WinRounds      int32  `bson:"win_rounds"`                             // 胜利局数
	LoseRounds     int32  `bson:"lose_rounds"`                            // 失败局数
	TieRounds      int32  `bson:"tie_rounds"`                             // 和局数
	WinBets        int64  `bson:"win_bets"`                               // 胜局打码量
	LoseBets       int64  `bson:"lose_bets"`                              // 败局打码量
	TieBets        int64  `bson:"tie_bets"`                               // 和局打码量
	Wins           int64  `bson:"wins"`                                   // 胜局赢钱金额
	Loses          int64  `bson:"loses"`                                  // 败局输钱金额
	Cash           int64  `bson:"cash"`                                   // 净赢
	SideWinner1    int64  `json:"sideWinner1" bson:"side_winner1"`        // 下注局开blue率
	SideWinner2    int64  `json:"sideWinner2" bson:"side_winner2"`        // 下注局开red率
	SideWinner3    int64  `json:"sideWinner3" bson:"side_winner3"`        // 下注局开Luckyshot率
	SideWinner4    int64  `json:"sideWinner4" bson:"side_winner4"`        // 下注局开9-Apair率
	SideWinner5    int64  `json:"sideWinner5" bson:"side_winner5"`        // 下注局开color率
	SideWinner6    int64  `json:"sideWinner6" bson:"side_winner6"`        // 下注局开seq率
	SideWinner7    int64  `json:"sideWinner7" bson:"side_winner7"`        // 下注局开pureseq率
	SideWinner8    int64  `json:"sideWinner8" bson:"side_winner8"`        // 下注局开set率
	SeatBets1      int64  `json:"seatBets1" bson:"seat_bets1"`            // 玩家押blue率
	SeatBets2      int64  `json:"seatBets2" bson:"seat_bets2"`            // 玩家押red率
	SeatBets3      int64  `json:"seatBets3" bson:"seat_bets3"`            // 玩家押Luckyshot率
	PlayerWinSeat1 int64  `json:"playerWinSeat1" bson:"player_win_seat1"` // 玩家blue中率
	PlayerWinSeat2 int64  `json:"playerWinSeat2" bson:"player_win_seat2"` // 玩家red中率
	PlayerWinSeat3 int64  `json:"playerWinSeat3" bson:"player_win_seat3"` // 玩家luckyshot中率
	PlayerWinSeat4 int64  `json:"playerWinSeat4" bson:"player_win_seat4"` // 玩家9-Apair中率
	PlayerWinSeat5 int64  `json:"playerWinSeat5" bson:"player_win_seat5"` // 玩家color中率
	PlayerWinSeat6 int64  `json:"playerWinSeat6" bson:"player_win_seat6"` // 玩家seq中率
	PlayerWinSeat7 int64  `json:"playerWinSeat7" bson:"player_win_seat7"` // 玩家pureseq中率
	PlayerWinSeat8 int64  `json:"playerWinSeat8" bson:"player_win_seat8"` // 玩家set中率
	PlayerMultis   int64  `bson:"player_multis"`                          // 多门局数 玩家下注2门及以上的局数
	RoundBetAvg    int64  `bson:"round_bet_avg"`                          // 局均码
	AD_BundleId    string `bson:"ad__bundle_id"`                          // 渠道
	Channel1       string `bson:"channel1"`                               // 渠道别名
	RegistArea     int    `bson:"regist_area"`                            // 账号类型 ab测试 0:A 1:B 2:C
	Money          uint32 `bson:"money"`                                  // 充值总金额(分)
	ObserveRounds  int32  `bson:"observe_rounds"`                         // 观察局数

}

// RB 红运当头
type RBHydtStat struct {
	Id                  string `bson:"_id"` // date-userid
	Date                int64  `bson:"date"`
	DateStr             string `bson:"date_str"`
	Userid              string `bson:"userid"`
	StrategyPlayers     int32  `bson:"strategy_players"`                             // 触发策略人数
	StrategyTimes       int32  `bson:"strategy_times"`                               // 触发策略总次数
	TiggerTimes         int    `json:"tiggerTimes" bson:"tigger_times"`              // u 触发策略有效次数
	AllEvoTimes         int    `json:"allEvoTimes" bson:"all_evo_times"`             // 触发策略局数 r总
	StrategyMultiRounds int32  `bson:"strategy_multi_rounds"`                        // 策略中多门局数
	StrategyBets        int64  `bson:"strategy_bets"`                                // 策略中总打码量
	StrategyWinRounds   int64  `json:"strategyWinRounds" bson:"strategy_win_rounds"` // 策略局中赢的局数
	StrategyWins        int64  `bson:"strategy_wins"`                                // 策略中总赢
	StrategyLoseRounds  int64  `json:"strategyLoseRounds" bson:"strategy_lose_rounds"`
	StrategyLoses       int64  `bson:"strategy_loses"` // 策略中总输
	StrategyCash        int64  `bson:"-"`              // 策略中净赢: 赢-输
	AD_BundleId         string `bson:"ad__bundle_id"`  // 渠道
	Channel1            string `bson:"channel1"`       // 渠道别名
	RegistArea          int    `bson:"regist_area"`    // 账号类型 ab测试 0:A 1:B 2:C
	Money               uint32 `bson:"money"`          // 充值总金额(分)

}

// RB 绝处逢生
type RBJcfsStat struct {
	Id                  string `bson:"_id"` // date-userid
	Date                int64  `bson:"date"`
	DateStr             string `bson:"date_str"`
	Userid              string `bson:"userid"`
	StrategyPlayers     int32  `bson:"strategy_players"`                             // 触发策略人数
	StrategyRounds      int64  `json:"strategyRounds" bson:"strategy_rounds"`        // 触发策略总局数
	AllInRounds         int32  `bson:"all_in_rounds"`                                // allin局数
	StrategyMultiRounds int32  `bson:"strategy_multi_rounds"`                        // 策略中多门局数
	StrategyBets        int64  `bson:"strategy_bets"`                                // 策略中总打码量
	StrategyWins        int64  `bson:"strategy_wins"`                                // 策略中总赢
	StrategyWinRounds   int64  `json:"strategyWinRounds" bson:"strategy_win_rounds"` // 策略局中赢的局数
	StrategyLoses       int64  `bson:"strategy_loses"`                               // 策略中总输
	StrategyLoseRounds  int64  `json:"strategyLoseRounds" bson:"strategy_lose_rounds"`
	StrategyCash        int64  `bson:"-"`             // 策略中净赢: 赢-输
	BetRounds           int32  `bson:"bet_rounds"`    // 下注局数
	WinRounds           int32  `bson:"win_rounds"`    // 胜利局数
	LoseRounds          int32  `bson:"lose_rounds"`   // 失败局数
	Wins                int64  `bson:"wins"`          // 胜局赢钱金额
	Loses               int64  `bson:"loses"`         // 败局输钱金额
	AD_BundleId         string `bson:"ad__bundle_id"` // 渠道
	Channel1            string `bson:"channel1"`      // 渠道别名
	RegistArea          int    `bson:"regist_area"`   // 账号类型 ab测试 0:A 1:B 2:C
	Money               uint32 `bson:"money"`         // 充值总金额(分)
}

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
}

type GameCharge struct {
	GType  int   // 游戏类型
	Amount int64 // 充值金额
	Number int64 // 充值人数
}
