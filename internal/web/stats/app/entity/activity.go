package entity

import "time"

// 基础付费活动
type BasicPayActivity struct {
	Id          string    `bson:"_id"`                             // 主键id
	Date        int64     `bson:"date"`                            // 统计日期时间戳
	SDate       time.Time `bson:"-"`                               // 统计日期
	ShopNumber  int64     `json:"shopNumber" bson:"shop_number"`   // 商场充值参与人数
	ShopDiamond float64   `json:"shopDiamond" bson:"shop_diamond"` // 商场充值彩金
	ShopCoin    float64   `json:"shopCoin" bson:"shop_coin"`       // 商场充值奖励金
	RmNumber    int64     `json:"rmNumber" bson:"rm_number"`       // 入门礼包参与人数
	RmDiamond   float64   `json:"rmDiamond" bson:"rm_diamond"`     // 入门礼包彩金
	RmCoin      float64   `json:"rmCoin" bson:"rm_coin"`           // 入门礼包奖励金
	JyNumber    int64     `json:"jyNumber" bson:"jy_number"`       // 金银铜卡参与人数
	JyDiamond   float64   `json:"jyDiamond" bson:"jy_diamond"`     // 金银铜卡彩金
	JyCoin      float64   `json:"jyCoin" bson:"jy_coin"`           // 金银铜卡奖励金
	ScNumber    int64     `json:"scNumber" bson:"sc_number"`       // 首充礼包参与人数
	ScDiamond   float64   `json:"scDiamond" bson:"sc_diamond"`     // 首充礼包彩金
	ScCoin      float64   `json:"scCoin" bson:"sc_coin"`           // 首充礼包奖励金
}

// 基础免费活动
type BasicFreeActivity struct {
	Id                string    `bson:"_id"`                                           // 主键id
	Date              int64     `bson:"date"`                                          // 统计日期时间戳
	SDate             time.Time `bson:"-"`                                             // 统计日期
	SignInNumber      int64     `json:"signInNumber" bson:"sign_in_number"`            // 每日签到参与人数
	SignInDiamond     float64   `json:"signInDiamond" bson:"sign_in_diamond"`          // 每日签到赠送彩金
	SignInCoin        float64   `json:"signInCoin" bson:"sign_in_coin"`                // 每日签到赠送奖励金
	OnlineNumber      int64     `json:"onlineNumber" bson:"online_number"`             // 在线奖励参与人数
	OnlineDiamond     float64   `json:"onlineDiamond" bson:"online_diamond"`           // 在线奖励赠送彩金
	OnlineCoin        float64   `json:"onlineCoin" bson:"online_coin"`                 // 在线奖励赠送奖励金
	TaskNumber        int64     `json:"taskNumber" bson:"task_number"`                 // 每日任务领取人数
	TaskDiamond       float64   `json:"taskDiamond" bson:"task_diamond"`               // 每日任务产出彩金
	TaskCoin          float64   `json:"taskCoin" bson:"task_coin"`                     // 每日任务产出奖励金
	TrailOrSetNumber  int64     `json:"trailOrSetNumber" bson:"trail_or_set_number"`   // 豹子牌型活动领取人数
	TrailOrSetDiamond float64   `json:"trailOrSetDiamond" bson:"trail_or_set_diamond"` // 豹子牌型活动产出彩金
	TrailOrSetCoin    float64   `json:"trailOrSetCoin" bson:"trail_or_set_coin"`       // 豹子牌型活动产出奖励金
}

// 分享活动数据
type ShareActivityData struct {
	Id                string    `json:"_id" bson:"_id"`
	RegisterNumber    int64     `json:"registerNumber" bson:"register_number"`
	WaterOutPut       float64   `json:"waterOutPut" bson:"water_out_put"`
	FWaterOutPut      float64   `bson:"-"`
	FirstOrderOutPut  float64   `json:"firstOrderOutPut" bson:"first_order_out_put"`
	FFirstOrderOutPut float64   `bson:"-"`
	GrossOutput       float64   `json:"grossOutput" bson:"gross_output"`
	FGrossOutput      float64   `bson:"-"`
	Ctime             int64     `json:"ctime" bson:"ctime"`
	Rtime             time.Time `bson:"-"`
}

// 埋点数据
type PointData4Web struct {
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

// 局数分析
type GameNumberAnalysis4Web struct {
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

// 时间分析
type PlaytimeAnalysis4Web struct {
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

// 付费留存
type PayUserRetained4Web struct {
	Id          string    `bson:"_id"`                              // 主键id
	Date        int64     `bson:"date"`                             // 统计日期时间戳
	SDate       time.Time `bson:"-"`                                // 统计日期
	Channel     string    `json:"channel" bson:"channel"`           // 渠道
	Channel1    string    `json:"channel1" bson:"channel1"`         // 渠道别名
	SType       int       `bson:"s_type"`                           // 数据类型 0：全部;1：按渠道
	PayUserType int       `json:"payUserType" bson:"pay_user_type"` // 充值用户类型 0:首充;2:复充
	NewNumber   int64     `json:"newNumber" bson:"new_number"`      // 充值人数
	Day1        float64   `json:"day1" bson:"day1"`                 // 1日留存
	Day2        float64   `json:"day2" bson:"day2"`                 // 2日留存
	Day3        float64   `json:"day3" bson:"day3"`                 // 3日留存
	Day4        float64   `json:"day4" bson:"day4"`                 // 4日留存
	Day5        float64   `json:"day5" bson:"day5"`                 // 5日留存
	Day6        float64   `json:"day6" bson:"day6"`                 // 6日留存
	Day7        float64   `json:"day7" bson:"day7"`                 // 7日留存
	Day15       float64   `json:"day15" bson:"day15"`               // 15日留存
	Day30       float64   `json:"day30" bson:"day30"`               // 30日留存
	Day60       float64   `json:"day60" bson:"day60"`               // 60日留存
}

type PayUserRetainedData struct {
	Id          string    `bson:"_id"`                              // 主键id
	Date        int64     `bson:"date"`                             // 统计日期时间戳
	SDate       time.Time `bson:"-"`                                // 统计日期
	Channel     string    `json:"channel" bson:"channel"`           // 渠道
	Channel1    string    `json:"channel1" bson:"channel1"`         // 渠道别名
	PayUserType int       `json:"payUserType" bson:"pay_user_type"` // 充值用户类型 0:首充;2:复充
	NewNumber   int64     `json:"newNumber" bson:"new_number"`      // 充值人数
	Pay1        int64     `json:"pay1" bson:"pay1"`                 // 1日留存=昨日新注册数且充值今日登录÷昨日新注册数充值
	// Register1   int64     `json:"register1" bson:"register1"`       // 1日留存=昨日新注册数且充值今日登录÷昨日新注册数充值
	Pay2 int64 `json:"pay2" bson:"pay2"`
	// Register2   int64     `json:"register2" bson:"register2"`
	Pay3 int64 `json:"pay3" bson:"pay3"`
	// Register3   int64     `json:"register3" bson:"register3"`
	Pay4 int64 `json:"pay4" bson:"pay4"`
	// Register4   int64     `json:"register4" bson:"register4"`
	Pay5 int64 `json:"pay5" bson:"pay5"`
	// Register5   int64     `json:"register5" bson:"register5"`
	Pay6 int64 `json:"pay6" bson:"pay6"`
	// Register6   int64     `json:"register6" bson:"register6"`
	Pay7 int64 `json:"pay7" bson:"pay7"`
	// Register7   int64     `json:"register7" bson:"register7"`
	Pay15 int64 `json:"pay15" bson:"pay15"`
	// Register15  int64     `json:"register15" bson:"register15"`
	Pay30 int64 `json:"pay30" bson:"pay30"`
	// Register30  int64     `json:"register30" bson:"register30"`
	Pay60 int64 `json:"pay60" bson:"pay60"`
	// Register60  int64     `json:"register60" bson:"register60"`
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

// 彩票活动
type LotteryActivity struct {
	Id          string `bson:"_id" json:"id"`                   // unique ID
	Date        int64  `bson:"date"`                            // 统计日期时间戳
	ClickNumber int64  `json:"clickNumber" bson:"click_number"` // 点击人数
	GJNumber    int64  `json:"gjNumber" bson:"gj_number"`       // 刮奖人数
	HJNumber    int64  `json:"hjNumber" bson:"hj_number"`       // 获奖人数
	FirstPrize  int64  `json:"firstPrize" bson:"first_prize"`   // 一等奖
	SecondPrize int64  `json:"secondPrize" bson:"second_prize"` // 二等奖
	ThirdPrize  int64  `json:"thirdPrize" bson:"third_prize"`   // 三等奖
	FourthPrize int64  `json:"fourthPrize" bson:"fourth_prize"` // 四等奖
	FifthPrize  int64  `json:"fifthPrize" bson:"fifth_prize"`   // 五等奖
	ExpendBouns int64  `json:"expendBouns" bson:"expend_bouns"` // BONUS消耗
}

type LogScratchTicket struct {
	Id        string `json:"id" bson:"_id"` // id
	Userid    string `json:"userid" bson:"userid"`
	Win       bool   `json:"win" bson:"win"`              //是否中奖
	Jackpot1  int64  `json:"jackpot1" bson:"jackpot1"`    //一等奖
	Jackpot2  int64  `json:"jackpot2" bson:"jackpot2"`    //二等奖
	Jackpot3  int64  `json:"jackpot3" bson:"jackpot3"`    //三等奖
	Jackpot4  int64  `json:"jackpot4" bson:"jackpot4"`    //四等奖
	Jackpot5  int64  `json:"jackpot5" bson:"jackpot5"`    //五等奖
	CostBonus int64  `json:"costBonus" bson:"cost_bonus"` //消耗bonus
	Ctime     int64  `json:"ctime" bson:"ctime"`
}
