package entity

import "time"

// 数据汇总
type DataStatistics struct {
	Id                   string    `bson:"_id"`                                                // 主键id
	Date                 int64     `bson:"date"`                                               // 统计日期时间戳
	SDate                time.Time `bson:"-"`                                                  // 统计日期
	Channel              string    `json:"channel" bson:"channel"`                             // 渠道
	Channel1             string    `json:"channel1" bson:"channel1"`                           // 渠道别名
	BloggerId            string    `bson:"blogger_id"`                                         // 博主用户ID
	NewRegister          int64     `json:"newRegister" bson:"new_register"`                    // 新注册数
	NewEquipment         int64     `json:"newEquipment" bson:"new_equipment"`                  // 新设备数
	LoginNum             int64     `json:"loginNum" bson:"login_num"`                          // 登录人数
	NextNewRegister      int64     `json:"nextNewRegister" bson:"next_new_register"`           // 昨日新注册数
	NextLogin            int64     `json:"nextLogin" bson:"next_login"`                        // 昨日新注册今日登录数
	NextDayRetention     float64   `bson:"-"`                                                  // 次日留存
	JRLoginNextPay       int64     `bson:"jr_login_next_pay"`                                  // 今日登录昨日充值玩家
	ZRPayCount           int64     `bson:"zr_pay_count"`                                       // 昨日充值的玩家人数
	PayRetention         float64   `bson:"-"`                                                  // 次日付费留存
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
	FTotalAmount         float64   `bson:"-"`                                                  // 总充值金额
	FTotalWithdrawAmount float64   `bson:"-"`                                                  // 提现总金额
	FHandlingCharge      float64   `bson:"-"`                                                  // 手续费
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
