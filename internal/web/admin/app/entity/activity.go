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

// 分享活动配置
type Share struct {
	Id             int32   `json:"id" bson:"_id"`            // id
	Recharge       int32   `json:"recharge" bson:"recharge"` // 充值领奖要求
	FRecharge      float64 `bson:"-"`
	FirstRecharge  int32   `json:"firstRecharge" bson:"first_recharge"` // 下线首单充值奖励
	FFirstRecharge float64 `bson:"-"`
	MaxPeople      int32   `json:"maxPeople" bson:"max_people"`     // 下线人数上限
	Url            string  `json:"url" bson:"url"`                  // 分享地址
	FirstReward    int32   `json:"firstReward" bson:"first_reward"` // 第一级奖励
	FFirstReward   float64 `bson:"-"`
	SecondReward   int32   `json:"secondReward" bson:"second_reward"` // 第二季奖励
	FSecondReward  float64 `bson:"-"`
}

type ShareRanking struct {
	Userid      string               `bson:"_id" json:"userid"`                  // 用户id
	Nickname    string               `bson:"nickname" json:"nickname"`           // 用户昵称
	AD_BundleId string               `json:"ad__bundle_id" bson:"ad__bundle_id"` // 包名
	Ctime       time.Time            `bson:"ctime" json:"ctime"`                 // 注册时间
	LoginTime   time.Time            `bson:"login_time" json:"login_time"`       // 最后登录时间
	ShareBelow  map[string]ShareData `json:"shareBelow" bson:"share_below"`      // 分享的下级
	ShareCount  int                  `bson:"-"`
}

type ShareData struct {
	BetAmount int64  `json:"betAmount" bson:"bet_amount"` //总打码量
	UserId    string `json:"userId" bson:"user_id"`
	Name      string `json:"name" bson:"name"`            //名称
	LastLogin int64  `json:"lastLogin" bson:"last_login"` //上次登录
	Recharge  int64  `json:"recharge" bson:"recharge"`    //充值金额
	Receive   bool   `json:"receive" bson:"receive"`      //是否领取奖励
}

// 打码量记录
type ShareAmountLog struct {
	Id        string        `json:"id" bson:"_id"`
	Ctime     int64         `json:"ctime" bson:"ctime"`          // 时间
	BetAmount int64         `json:"betAmount" bson:"bet_amount"` // 总打码量
	Detail    []ShareDetail `json:"detail" bson:"detail"`        // 详情
}

type ShareDetail struct {
	BetAmount int64  `json:"betAmount" bson:"bet_amount"` //总打码量
	UserId    string `json:"userId" bson:"user_id"`
	Name      string `json:"name" bson:"name"`     //名称
	Income    int64  `json:"income" bson:"income"` //今日进账
}

// 下级玩家打码量
type ShareBetAmount struct {
	Userid   string `json:"userid" bson:"_id"`        // 玩家id
	Superior string `json:"superior" bson:"superior"` // 上级玩家id
	Score    int64  `json:"score" bson:"score"`       // 打码量
	Ctime    int64  `json:"ctime" bson:"ctime"`       // 时间
}

// 分享提取流水
type LogShareWithDraw struct {
	Id        string    `json:"_id" bson:"_id"`
	Userid    string    `json:"userid" bson:"userid"`
	BundleId  string    `json:"bundleId" bson:"bundle_id"`
	Gtype     int32     `json:"gtype" bson:"gtype"` //流水类型 1:流水奖励；2：首单奖励
	GtypeName string    `bson:"-"`
	Cash      int64     `json:"cash" bson:"cash"`
	FCash     float64   `bson:"-"`
	Coin      int64     `json:"coin" bson:"coin"`
	FCoin     float64   `bson:"-"`
	Ctime     int64     `json:"ctime" bson:"ctime"`
	Rtime     time.Time `bson:"-"`
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

// 小米活动中奖记录
type LuckyDrawGive struct {
	Id       string    `bson:"_id" json:"id"`              // unique ID
	Round    int64     `bson:"round" json:"round"`         // 中奖轮数，第几期
	RewordId int32     `bson:"reward_id" json:"reward_id"` // 几等奖
	Reword   int64     `bson:"reword" json:"reword"`       // 中奖彩金
	Userid   string    `bson:"userid" json:"userid"`       // 中奖人id
	Number   string    `bson:"number" json:"number"`       // 中奖号码
	Robot    bool      `bson:"robot" json:"robot"`         // 是否机器人
	Taked    bool      `bson:"taked" json:"taked"`         // 已领取
	Ctime    time.Time `bson:"ctime" json:"ctime"`         // 创建时间
	FReword  float64   `bson:"-"`                          // 中奖彩金
}

// 彩票活动
type LotteryActivity struct {
	Id           string    `bson:"_id" json:"id"`                   // unique ID
	Date         int64     `bson:"date"`                            // 统计日期时间戳
	SDate        time.Time `bson:"-"`                               // 统计日期
	ClickNumber  int64     `json:"clickNumber" bson:"click_number"` // 点击人数
	GJNumber     int64     `json:"gjNumber" bson:"gj_number"`       // 刮奖人数
	HJNumber     int64     `json:"hjNumber" bson:"hj_number"`       // 获奖人数
	FirstPrize   int64     `json:"firstPrize" bson:"first_prize"`   // 一等奖
	SecondPrize  int64     `json:"secondPrize" bson:"second_prize"` // 二等奖
	ThirdPrize   int64     `json:"thirdPrize" bson:"third_prize"`   // 三等奖
	FourthPrize  int64     `json:"fourthPrize" bson:"fourth_prize"` // 四等奖
	FifthPrize   int64     `json:"fifthPrize" bson:"fifth_prize"`   // 五等奖
	ExpendBouns  int64     `json:"expendBouns" bson:"expend_bouns"` // BONUS消耗
	FFirstPrize  float64   `bson:"-"`                               // 一等奖
	FSecondPrize float64   `bson:"-"`                               // 二等奖
	FThirdPrize  float64   `bson:"-"`                               // 三等奖
	FFourthPrize float64   `bson:"-"`                               // 四等奖
	FFifthPrize  float64   `bson:"-"`                               // 五等奖
	FExpendBouns float64   `bson:"-"`                               // BONUS消耗
}

// 拼多多分享活动
type PddStat struct {
	Id                     string `bson:"_id"`                       // id
	Date                   int64  `bson:"date"`                      // 统计日期时间戳
	LoginedUsers           int32  `bson:"logined_users"`             // 渠道总日活
	PddUsers               int32  `bson:"pdd_users"`                 // 参与用户
	BeInvoteUsers          int32  `bson:"be_invote_users"`           // 今日参与活动中，分享来的用户数，有效分享数
	UserDraws              int32  `bson:"user_draws"`                // 所有渠道参与抽奖总次数
	BeInvoteLoginedUsers   int32  `bson:"be_invote_logined_users"`   // 被分享的用户登录人数，日活人数
	DrawUsers              int32  `bson:"draw_users"`                // 今日参与活动中，分享来的用户数
	BeInvoteUserDraws      int32  `bson:"be_invote_user_draws"`      // 分享用户参与抽奖总次数
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

	SDate         string `bson:"-"`
	PddRate       string `bson:"-"` // 总参与率
	BeInvoteRate  string `bson:"-"` // 分享用户参与率
	ShareDrawRate string `bson:"-"` // 总人均抽奖次数
	ShareDrawAvg  string `bson:"-"` // 分享用户人均抽奖次数
	SAwardJackpot string `bson:"-"` // 累计发出去的金额
	YxShareRate   string `bson:"-"` // 总有效分享率
	YxShare2Rate  string `bson:"-"` // 分享用户有效分享数
	ZCrp          string `bson:"-"` // 总CRP
	YXCrp         string `bson:"-"` // 有效CRP
	SCCpp         string `bson:"-"` // 首充CPP
	Profit        string `bson:"-"` // 活动利润
	Roi           string `bson:"-"` // D0净ROI
	LjRoi         string `bson:"-"` // 累计净ROI
	DoRoi         string `bson:"-"` // D0毛ROI
	LjMRoi        string `bson:"-"` // 累计毛ROI
	LoginedRate   string `bson:"-"` // 日活占比
	PddChargeRate string `bson:"-"` // 活动总充额占比
}

// 打码排行榜活动
type BetRankStat struct {
	SDate        string // 发放日期
	Week         string // 星期
	RankType     string // 榜单类型
	PrizeUsers   int64  // 获奖人数
	Jackpot      string // 奖池金额
	PrizeSum     string // 分到金额
	PrizeAvg     string // 人均奖金
	PrizeType    string // 奖金类型
	UserMaxRank  string // 真人最高名次
	MaxUserid    string // 最高名次真人的id
	MaxUserPrize string // 最高名次真人分到金额
}

// 打码排行榜明细
type BetRankUserStat struct {
	RankDate        string // 榜单日期 2024/12/23-2024/12/29
	SDate           string // 发放日期
	Userid          string // 玩家ID
	RankType        string // 榜单类型
	Etime           string // 结算时间 时分秒
	Stime           string // 发放时间 时分秒
	ReceiveTime     string // 领取时间
	ReceiveHourGap  string // 发领间隔(小时)
	Pays            string // 总充值
	Withdraws       string // 总提现
	Contribute      string // 玩家总贡献
	StimePays       string // 当日充值
	StimeWithdraws  string // 当日提现
	StimeContribute string // 当日总贡献
	RankBets        string // 榜单期打码
	StimeBets       string // 发放日打码
	Rank            string // 名次
	Prize           string // 分到金额
	PrizeType       string // 奖金类型
	Phone           string // 玩家手机号

	SDate2 [2]int64 // 发放日时间段
}

// 转盘活动领奖记录
type ActivityTurnPrizeLog struct {
	Id                string `bson:"_id" json:"id"`                                // 抽奖记录id
	TurnStime         int64  `bson:"turn_stime" json:"turnStime"`                  // 本轮转盘开始时间
	TurnEtime         int64  `bson:"turn_etime" json:"turnEtime"`                  // 本轮转盘结束时间
	Userid            string `bson:"userid" json:"userid"`                         // 用户id
	ChannelId         string `bson:"channel_id" json:"channelId"`                  // 用户渠道
	RegistArea        int8   `bson:"regist_area" json:"registArea"`                // 用户ab测试 0:A 1:B 2:C
	Rtime             int64  `bson:"rtime" json:"rtime"`                           // 用户注册时间(时间戳秒)
	DrawedTimesFree   int32  `bson:"drawed_times_free" json:"drawedTimesFree"`     // 已使用免费抽奖次数
	DrawedTimesInvite int32  `bson:"drawed_times_invite" json:"drawedTimesInvite"` // 已使用邀请抽奖次数
	Gives             int64  `bson:"gives" json:"gives"`                           // 赠送序幕金
	Prize             int64  `bson:"prize" json:"prize"`                           // 奖金金额
	PrizeType         int32  `bson:"prize_type" json:"prizeType"`                  // 奖金类型:（1.bonus,2.cash,3.withdrawalble）
	Ctime             int64  `bson:"ctime" json:"ctime"`                           // 申请领奖时间(时间戳秒)
	State             int32  `bson:"state" json:"state"`                           // 审核状态: 0审核中,1通过,2拒绝
	Stime             int64  `bson:"stime" json:"stime"`                           // 审核时间(时间戳秒)
	Stype             int32  `bson:"stype" json:"stype"`                           // 审核类型:1机审,2人工审核
	Reason            string `bson:"reason" json:"reason"`                         // 机审结果原因
	AuditUser         string `bson:"audit_user" json:"auditUser"`                  // 审核人
	Remark            string `bson:"remark" json:"remark"`                         // 后台备注
}

// ActivityTurnStatDate 转盘活动汇总
type ActivityTurnStatDate struct {
	SDate              string // 日期
	LoginedUsers       int64  // 总日活人数
	NewLoginedUsers    int64  // 新用户日活
	OldLoginedUsers    int64  // 老用户日活
	TurnUsers          int64  // 总参与人数
	NewTurnUsers       int64  // 新用户参与人数
	OldTurnUsers       int64  // 老用户参与人数
	FirstTurnUsers     int64  // 首次参与总人数
	FirstNewTurnUsers  int64  // 首次参与新用户人数
	FirstOldTurnUsers  int64  // 首次参与老用户人数
	PrizeUsers         int64  // 中奖人数
	TackPrizeUsers     int64  // 当日申请领奖人数
	TackPrizes         string // 当日申请领奖金额
	TackPassPrizeUsers int64  // 当日申请且通过的人数
	TackPassPrizes     string // 当日申请且通过的金额
	PassPrizeUsers     int64  // 当日审核通过总人数
	PassPrizes         string // 当日审核通过总金额
	//
	LaunchInviters        int64 // 当日发起邀请总人数
	LaunchNewInviters     int64 // 发起邀请的新用户人数
	LaunchOldInviters     int64 // 发起邀请的老用户人数
	InvitedInviters       int64 // 当日邀请就成功的邀请者人数
	InvitedNewInviters    int64 // 当日新用户邀请且成功的人数
	InvitedOldInviters    int64 // 当日老用户邀请且成功的人数
	NowInvitedInviters    int64 // 当日邀请截至目前成功的邀请者人数
	NowInvitedNewInviters int64 // 当日新用户邀请截至目前成功的邀请者人数
	NowInvitedOldInviters int64 // 当日老用户邀请截至目前成功的邀请者人数
	ShareUsers            int64 // 裂变总人数
	NewInviterShareUsers  int64 // 新用户裂变人数
	OldInviterShareUsers  int64 // 老用户裂变人数
	//
	LoginTurnRate        string // 日活参与率
	NewUserTurnRate      string // 新用户参与率
	OldUserTurnRate      string // 老用户参与率
	LoginFirstTurnRate   string // 日活首次参与率
	NewUserFirstTurnRate string // 新用户首次参与率
	OldUserFirstTurnRate string // 老用户首次参与率
	PrizeTurnRate        string // 参与中奖率
	TackPrizeRate        string // 中奖后申请率
	TackPrizePassRate    string // 当日申请且当日通过率
	//
	TurnUserInviteRate       string // 参与者总邀请率
	TurnNewUserInviteRate    string // 参与者新用户邀请率
	TurnOldUserInviteRate    string // 参与者老用户邀请率
	InviterInvitedRate       string // 当日邀请者当日总成功率
	NewInviterInvitedRate    string // 当日新用户邀请且当日成功的成功率
	OldInviterInvitedRate    string // 当日老用户邀请且当日成功的成功率
	InviterNowInvitedRate    string // 当日邀请者目前总成功率
	NewInviterNowInvitedRate string // 当日新用户邀请截至目前成功的成功率
	OldInviterNowInvitedRate string // 当日老用户邀请截至目前成功的成功率
	ShareUsersRate           string // 总参与裂变率
	NewInviterShareUsersRate string // 新用户参与裂变率
	OldInviterShareUsersRate string // 老用户参与裂变率
	ShareInviterRate         string // 总邀请裂变率
	NewShareInviterRate      string // 新用户邀请裂变率
	OldShareInviterRate      string // 老用户邀请裂变率

	TackPrizes0     int64 // 当日申请领奖金额
	TackPassPrizes0 int64 // 当日申请且通过的金额
	PassPrizes0     int64 // 当日审核通过总金额
}

// ActivityTurnStatUser 转盘活动明细
type ActivityTurnStatUser struct {
	Userid                string // UID
	ChannelClass          string // 渠道类
	ChannelId             string // 渠道
	RegistArea            string // 账号类型
	Ctime                 string // 注册时间
	LoginTime             string // 最后登录时间
	LiveDays              string // 存活天数
	LoseDays              string // 流失天数
	VipLv                 string // VIP等级
	FUserType             string // 玩家类型
	Pays                  string // 玩家总充值
	Withdraws             string // 玩家总提现
	Turns                 int64  // 总经历轮数
	DrawTurns             int64  // 总参与轮数
	InviteUsers           int64  // 总邀请来人数
	InviteUsersPayed      int64  // 总邀请来付费人数
	InviteUsersPays       string // 总邀请来付费金额
	InviteUsersWithdraws  string // 总邀请来提现金额
	InviteUsersContribute string // 总邀请来人的贡献
	//
	TackPrizeTurns            int64  // 总申请领奖轮数
	TackPrizePassTurns        int64  // 申请已通过的轮数
	TackPrizePassScore        string // 已领到总金额
	TackPrizePassBonus        string // 已领到总bonus
	TackPrizePassCash         string // 已领到总cash
	TackPrizePassWithdrawable string // 已领到总withdrawable
	ScoreRoi                  string // 总奖金roi
	BonusRoi                  string // bonus roi
	CashRoi                   string // cash roi
	WithdrawableRoi           string // withdrawable roi
	//
	DrawTurnInviteUsersAvg           string // 参与轮均邀请来人数
	DrawTurnInviteUsersPayedAvg      string // 参与轮均邀请来付费人数
	TackPrizeTurnInviteUsersAvg      string // 申请轮均邀请来人数
	TackPrizeTurnInviteUsersPayedAvg string // 申请轮均邀请来付费人数
	TurnRate                         string // 参与率
	TackPrizeRate                    string // 参与中奖率
	TackPrizePassRate                string // 中奖通过率
	InviteUsersPayedRate             string // 总邀请来付费率

	InviteUsersContribute0 int64 // 总邀请来人的贡献
	BonusContribute        int64 // bonus 轮邀请来的人贡献
	CashContribute         int64 // cash 轮邀请来的人贡献
	WithdrawableContribute int64 // withdrawable 轮邀请来的人贡献
}

// ActivityTurnStatUser 转盘活动审核
type ActivityTurnStatAudit struct {
	Id               string   // 申请id
	Userid           string   // UID
	ChannelClass     string   // 渠道类
	ChannelId        string   // 渠道
	RegistArea       string   // 账号类型
	Ctime            string   // 注册时间
	LoginTime        string   // 最后登录时间
	LiveDays         string   // 存活天数
	LoseDays         string   // 流失天数
	VipLv            string   // VIP等级
	UserType         string   // 玩家类型
	UserBets         string   // 玩家总打码
	Pays             string   // 玩家总充值
	Withdraws        string   // 玩家总提现
	TurnInvites      int64    // 本轮邀请来人数
	TurnUserids      []string // 本轮邀请来人id
	TurnInvitesPayed int64    // 本轮邀请来付费人数
	TurnInvitesPays  string   // 本轮邀请来付费金额
	Prize            string   // 奖金金额
	PrizeType        string   // 奖金类型
	TurnNo           string   // 当前是总的第几轮
	TackPrizeNo      string   // 当前是第几次申请
	Status           string   // 审核状态
	Remark           string   // 备注
	SubmitTime       string   // 申请时间
	AuditTime        string   // 审定时间
	AuditUser        string   // 审核人

	State     int64 // 审核状态0待审核1通过2拒绝
	Remarks   []string
	TurnStime int64 // 本轮开始时间
}

// 礼包码
type GiftPackCode struct {
	Code           string   `bson:"_id" json:"id"`                         // id,码号
	Status         int32    `bson:"status" json:"status"`                  // 0.关 1.开
	GiftType       int32    `bson:"gift_type" json:"giftType"`             // 1.随机码 2.等额码
	ScoreType      int32    `bson:"score_type" json:"scoreType"`           // 奖金类型: 1.bonus, 2.cash, 3.withdrawable
	Score          int64    `bson:"score" json:"score"`                    // 总金额
	Pieces         int32    `bson:"pieces" json:"pieces"`                  // 份数
	ScoreReceived  int64    `bson:"score_received" json:"scoreReceived"`   // 已领取份数
	PiecesReceived int32    `bson:"pieces_received" json:"piecesReceived"` // 已领取金额
	FinishTime     int64    `bson:"finish_time" json:"finishTime"`         // 领完时间
	PieceMin       int64    `bson:"piece_min" json:"pieceMin"`             // 单份最小金额
	PieceMax       int64    `bson:"piece_max" json:"pieceMax"`             // 单份最大金额
	TimeOpen       int64    `bson:"time_open" json:"timeOpen"`             // 有效期开始秒
	TimeClose      int64    `bson:"time_close" json:"timeClose"`           // 有效期结束秒
	Utypes         []int32  `bson:"utypes" json:"utypes"`                  // 标签满足一个即可: 0:A 1:B 2:C, 10.新手 11.平民 12.普R 13.小R 14.中R 15.大R 16.超大R
	Ctime          int64    `bson:"ctime" json:"ctime"`                    // 创建时间秒
	Cuser          string   `bson:"cuser" json:"cuser"`                    // 创建人
	Remark         string   `bson:"remark" json:"remark"`                  // 礼包码标签
	Userids        []string `bson:"-" json:"userids"`                      // 已领取用户列表

	FCtime         string `bson:"-" json:"-"`
	FStatus        string `bson:"-" json:"-"`
	FCode          string `bson:"-" json:"-"`
	FGiftType      string `bson:"-" json:"-"`
	FScoreType     string `bson:"-" json:"-"`
	FScore         string `bson:"-" json:"-"`
	FScoreReceived string `bson:"-" json:"-"`
	FFinishTime    string `bson:"-" json:"-"`
	FPieceMin      string `bson:"-" json:"-"`
	FPieceMax      string `bson:"-" json:"-"`
	FTimeOpen      string `bson:"-" json:"-"`
	FTimeClose     string `bson:"-" json:"-"`
	FUtypes        string `bson:"-" json:"-"`
	FCuser         string `bson:"-" json:"-"`
	FStatus2       string `bson:"-" json:"-"`
	FUtypes2       string `bson:"-" json:"-"`
}

// 波动返水各类水池
type VolatilitySubsidyPool struct {
	Id         string `bson:"_id" json:"id"`                 // id: date-ptype-registArea
	Date       string `bson:"date" json:"date"`              // 日期
	Ptype      int32  `bson:"ptype" json:"ptype"`            // 1.保底水池,2.静态水池,3.动态水池,4.已补贴金额
	RegistArea int32  `bson:"regist_area" json:"registArea"` // abc类
	Amount     int64  `bson:"amount" json:"amount"`          // 水池金额
	Utime      int64  `bson:"utime" json:"utime"`            // 更新时间
}

// VolatilitySubsidyDate 波动返水汇总
type VolatilitySubsidyDate struct {
	SDate                     string   // 日期
	LoginedUsers              int64    // 日活用户数
	FirstPayUsers             int64    // 首充用户数
	SubsidyUsers              int64    // 领到补贴总人数
	FirstSubsidyUsers         int64    // 首次领到补贴总人数
	FirstSubsidyRate          string   // 首领占比
	SubsidyTimes              int64    // 领到补贴总次数
	SubsidyAvg                string   // 人均领到次数
	FirstPaySubsidyUsers      int64    // 当日首充用户领到人数
	FirstPaySubsidyRate       string   // 当日首充用户被补贴率
	FirstPaySubsidyUsersRate  string   // 当日首充用户领到人数占总领人数比
	SubsidyPool               string   // 水池总金额
	SubsidyAmount             string   // 补贴总金额
	SubsidyConsumRate         string   // 水池消耗率
	SubsidyAmountAvg          string   // 人均补贴金额
	FirstSubsidyRetentionDay1 string   // 首次补贴用户次日留存率
	FirstSubsidyRetentionDay3 string   // 首次补贴用户3日留存率
	FirstSubsidyRetentionDay5 string   // 首次补贴用户5日留存率
	FirstSubsidyRetentionDay7 string   // 首次补贴用户7日留存率
	FirstPayRetentionDay1     string   // 首充用户次日留存率
	FirstPayRetentionDay3     string   // 首充用户3日留存率
	FirstPayRetentionDay5     string   // 首充用户5日留存率
	FirstPayRetentionDay7     string   // 首充用户7日留存率
	FirstSubsidyPayAmountDay7 string   // 首次补贴用户7日人均付费总额
	FirstPayAmountDay7Avg     string   // 首充7日人均付费总额
	SubsidyTop3               []string // 当日补贴金额top123
	TopSubsidy3               []int64  // 当日补贴金额top123

	SubsidyPools          map[int32]int64 // 各类水池金额
	SubsidyPool0          int64           // 水池总金额
	SubsidyAmount0        int64           // 补贴总金额
	FirstSubsidyLoginDay1 int64           // 当日领到补贴的用户，次日登录的人数
	FirstSubsidyLoginDay3 int64           // 当日领到补贴的用户，第3日登录的人数
	FirstSubsidyLoginDay5 int64           // 当日领到补贴的用户，第5日登录的人数
	FirstSubsidyLoginDay7 int64           // 当日领到补贴的用户，第7日登录的人数
	FirstPayLoginDay1     int64           // 首充用户次日登录人数
	FirstPayLoginDay3     int64           // 首充用户3日登录人数
	FirstPayLoginDay5     int64           // 首充用户5日登录人数
	FirstPayLoginDay7     int64           // 首充用户7日登录人数
	PayAmount1            int64
	PayUsers1             int64
	PayAmount2            int64
	PayUsers2             int64
}

// VolatilitySubsidyDate 波动返水明细
type VolatilitySubsidyUser struct {
	Userid               string // UID
	VipLv                string // VIP等级
	ChargeType           string // 用户类型
	RegistDate           string // 注册日期
	LiveDays             string // 存活天数
	LoseDays             string // 流失天数
	FirstPayTime         string // 首充时间
	FirstSubsidyTime     string // 首次补贴时间
	FirstPay2SubsidyDays string // 首次补贴距首充天数
	FirstPayAmount       string // 首充金额
	Pays                 string // 总充值金额
	Withdraws            string // 总提现金额
	PaySubWithdraws      string // 充-提
	Subsidys             string // 总补贴金额
	SubsidyTimes         int64  // 总补贴次数
}

// ShareAgentCost 代理后台-当日产出/实付总成本
type ShareAgentCost struct {
	Date           uint32
	DayProductCost int64 // 当日产出总成本
	DayPayOutCost  int64 // 当日实付总成本

	DayProductCostHeadChild int64 // 当日受邀者总成本
}

// ShareAgentDate 代理后台-裂变体系日报
type ShareAgentDate struct {
	SDate          string // 日期
	DayProductCost string // 当日产出总成本
	DayPayOutCost  string // 当日实付总成本
	DayROI         string // 当日ROI (充减提/当日实付总成本)

	ShareLoginUsers    int64  // 体系日活人数
	ShareApkLoginUsers int64  // 分享包日活人数
	ShareApkLoginRate  string // 体系日活的分享包占比
	NewShareLoginUsers int64  // 新用户日活
	OldShareLoginUsers int64  // 老用户日活
	NewInviterUsers    int64  // 新增代理头目人数

	SharePayAmounts      string // 体系充值总额
	ShareWithdrawAmounts string // 体系提现金额
	SharePaySubWithdraw  string // 充减提
	ShareProfitRate      string // 体系盈利率
	SharePayUsers        int64  // 体系充值人数
	ShareWithdrawUsers   int64  // 体系提现人数
	SharePayRate         string // 体系付费率
	ShareWithdrawRate    string // 体系提现人数占比
	ShareARPU            string // 体系ARPU 总充值金额 / 登录人数
	ShareARPPU           string // 体系ARPPU 总充值金额÷总充值人数

	ShareFirstPayUsers           int64  // 体系首充人数
	ShareFirstPayAmounts         string // 首充人充值总额
	ShareFirstPayWithdrawUsers   int64  // 首充人的提现人数
	ShareFirstPayWithdrawAmounts string // 首充人提现总额
	ShareFirstPaySubWithdraw     string // 首充人的充减提
	ShareFirstPayProfit          string // 首充人盈余率
	ShareFirstPayUnpayRate       string // 未充人首充率
	ShareFirstPayWithdrawRate    string // 首充人提现占比

	NewSharePayUsers        int64  // 体系新用户充值人数
	NewSharePayAmounts      string // 新用户充值金额
	NewShareWithdrawUsers   int64  // 新用户提现人数
	NewShareWithdrawAmounts string // 新用户提现总额
	NewSharePaySubWithdraw  string // 新用户的充减提
	NewShareProfitRate      string // 新用户盈余率
	NewSharePayRate         string // 新用户付费率
	NewShareWithdrawRate    string // 新用户提现率
	NewShareARPU            string // 新用户ARPU
	NewShareARPPU           string // 新用户ARPPU

	OldSharePayUsers        int64  // 体系老用户充值人数
	OldSharePayAmounts      string // 老用户充值金额
	OldShareWithdrawUsers   int64  // 老用户提现人数
	OldShareWithdrawAmounts string // 老用户提现总额
	OldSharePaySubWithdraw  string // 老用户的充减提
	OldShareProfitRate      string // 老用户盈余率
	OldSharePayRate         string // 老用户付费率
	OldShareWithdrawRate    string // 老用户提现率
	OldShareARPU            string // 老用户ARPU
	OldShareARPPU           string // 老用户ARPPU

	ShareLoginPayUsers            int64 // 体系日活付费人数
	DayProductCost0               int64 // 当日产出总成本 (当日产生打码返佣额+当日产生转盘奖金额+当日产生单个人头奖金额+当日产生累计人头奖金额)
	DayPayOutCost0                int64 // 当日实付总成本 (当日领走打码返佣额+当日领走转盘奖金额+当日领走单个人头奖金额+当日领走累计人头奖金额)
	SharePayAmounts0              int64 // 体系充值总额
	ShareWithdrawAmounts0         int64 // 体系提现金额
	ShareFirstPayAmounts0         int64 // 首充人充值总额
	ShareFirstPayWithdrawAmounts0 int64 // 首充人提现总额
	ShareFirstPaySubWithdraw0     int64 // 首充人的充减提
	NewSharePayAmounts0           int64 // 新用户充值金额
	NewShareWithdrawAmounts0      int64 // 新用户提现总额
	OldSharePayAmounts0           int64 // 老用户充值金额
	OldShareWithdrawAmounts0      int64 // 老用户提现总额
}

// ShareAgentFinanceDate 代理后台-代理活动经济日览
type ShareAgentFinanceDate struct {
	SDate               string // 日期
	DayProductCost      string // 当日产出总成本
	DayPayOutCost       string // 当日实付总成本
	ShareBetsReward     string // 当日产生打码返佣额
	ShareBetsRewardTake string // 当日领走打码返佣额
	ShareHeadReward     string // 当日产生单个人头奖金额
	ShareHeadRewardTake string // 当日领走单个人头奖金额
	ShareTaskReward     string // 当日产生累计人头奖金额
	ShareTaskRewardTake string // 当日领走累计人头奖金额

	ShareUsers       int64 // 当日裂变总人数
	ShareTeamV1Users int64 // 属于1级团队的人数
	ShareTeamV2Users int64 // 属于2级团队的人数
	ShareTeamV3Users int64 // 属于3级团队的人数
	ShareTeamV4Users int64 // 属于4级团队的人数

	ShareInvalidUsers       int64 // 当日裂变算人头奖人数
	ShareTeamV1InvalidUsers int64 // 属于1级团队的人数
	ShareTeamV2InvalidUsers int64 // 属于2级团队的人数
	ShareTeamV3InvalidUsers int64 // 属于3级团队的人数
	ShareTeamV4InvalidUsers int64 // 属于4级团队的人数

	ShareHeadReward2      string // 单个人头奖励总金额
	ShareHeadRewardSuper  string // 当日邀请者单个人头奖励总金额
	ShareHeadRewardChild  string // 当日受邀者单个人头奖励总金额
	ShareTeamV1HeadReward string // 属于1级团队的金额 这里不用去分邀请者还是受邀者，只看邀请者属于哪个团队等级就行。因为邀请者属于某等级，其邀请来的人肯定也是这个等级的。
	ShareTeamV2HeadReward string // 属于2级团队的金额
	ShareTeamV3HeadReward string // 属于3级团队的金额
	ShareTeamV4HeadReward string // 属于4级团队的金额

	ShareTaskReward2      string // 累计人头奖励总金额
	ShareTeamV1TaskReward string // 属于1级团队的金额
	ShareTeamV2TaskReward string // 属于2级团队的金额
	ShareTeamV3TaskReward string // 属于3级团队的金额
	ShareTeamV4TaskReward string // 属于4级团队的金额

	ShareBetsReward2      string // 打码返佣总额
	ShareTeamV1BetsReward string // 属于1级团队的金额
	ShareTeamV2BetsReward string // 属于2级团队的金额
	ShareTeamV3BetsReward string // 属于3级团队的金额
	ShareTeamV4BetsReward string // 属于4级团队的金额

	DayProductCost0      int64 // 当日产出总成本
	DayPayOutCost0       int64 // 当日实付总成本
	ShareBetsReward0     int64 // 当日产生打码返佣额
	ShareBetsRewardTake0 int64 // 当日领走打码返佣额
	ShareHeadReward0     int64 // 当日产生单个人头奖金额
	ShareHeadRewardTake0 int64 // 当日领走单个人头奖金额
	ShareTaskReward0     int64 // 当日产生累计人头奖金额
	ShareTaskRewardTake0 int64 // 当日领走累计人头奖金额

	ShareHeadReward20      int64 // 单个人头奖励总金额
	ShareHeadRewardSuper0  int64 // 当日邀请者单个人头奖励总金额
	ShareHeadRewardChild0  int64 // 当日受邀者单个人头奖励总金额
	ShareTeamV1HeadReward0 int64 // 属于1级团队的金额 这里不用去分邀请者还是受邀者，只看邀请者属于哪个团队等级就行。因为邀请者属于某等级，其邀请来的人肯定也是这个等级的。
	ShareTeamV2HeadReward0 int64 // 属于2级团队的金额
	ShareTeamV3HeadReward0 int64 // 属于3级团队的金额
	ShareTeamV4HeadReward0 int64 // 属于4级团队的金额

	ShareTaskReward20      int64 // 累计人头奖励总金额
	ShareTeamV1TaskReward0 int64 // 属于1级团队的金额
	ShareTeamV2TaskReward0 int64 // 属于2级团队的金额
	ShareTeamV3TaskReward0 int64 // 属于3级团队的金额
	ShareTeamV4TaskReward0 int64 // 属于4级团队的金额

	ShareBetsReward20      int64 // 打码返佣总额
	ShareTeamV1BetsReward0 int64 // 属于1级团队的金额
	ShareTeamV2BetsReward0 int64 // 属于2级团队的金额
	ShareTeamV3BetsReward0 int64 // 属于3级团队的金额
	ShareTeamV4BetsReward0 int64 // 属于4级团队的金额
}

// ShareAgentSuperStat 代理后台-代理头目明细
type ShareAgentSuperStat struct {
	SuperId       string // 头目UID
	SuperCtime    string // 注册日期
	BeSuperTime   string // 成为头目的日期
	GrowDays      string // 发育效率
	LastLoginDays int64  // 团队死亡天数
	SuperROI      string // 团头总ROI
	TeamROI       string // 团队ROI

	ProdAmountsBet   string // 产生打码返佣总额
	TurnPrize        string // 产生转盘奖金总额
	ProdAmountsHead  string // 产生单个人头奖励总额（给头目）
	ProdAmountsChild string // 产生单个人头奖励总额（给团队）
	ProdAmountsTask  string // 产生累计人头奖励总额
	TakeAmountsBet   string // 领走打码返佣总额
	TakeTurnPrize    string // 领走转盘奖金总额
	TakeAmountsHead  string // 领走单个人头奖励总额（给头目）
	TakeAmountsChild string // 领走单个人头奖励总额（给团队）
	TakeAmountsTask  string // 领走累计人头奖励总额

	SuperPays        string // 头目总充值
	SuperCashOut     string // 头目总提现
	SuperProfit      string // 头目充-提
	SuperProfitRate  string // 头目盈余率
	SuperBets        string // 头目总打码 x
	SuperBetsPayRate string // 头目充投比 x

	TeamLv           int64  // 团队等级
	GroupUsers       int64  // 团队人数
	GroupPayUsers    int64  // 团队付费人数
	GroupPayRate     string // 团队付费率
	GroupPays        string // 团队总充值
	GroupCashOut     string // 团队总提现
	GroupProfit      string // 团队充-提
	GroupProfitRate  string // 团队盈余率
	GroupBets        string // 团队总打码 x
	GroupBetsPayRate string // 团队充投比 x
	GroupArpu        string // 团队arpu
	GroupArppu       string // 团队arppu
}

// ShareAgentSuperStat 代理后台-代理头目奖金日排名
type ShareAgentSuperBonusStats struct {
	SDate            string // 日期
	SuperId          string // 头目UID
	SuperROI         string // 团头总ROI
	TeamROI          string // 团队ROI
	ProdAmounts      string // 当日总产生奖金
	TakeAmounts      string // 当日总领走奖金
	ProdAmountsBet   string // 产生打码返佣总额
	TurnPrize        string // 产生转盘奖金总额
	ProdAmountsHead  string // 产生单个人头奖励总额（给头目）
	ProdAmountsChild string // 产生单个人头奖励总额（给团队）
	ProdAmountsTask  string // 产生累计人头奖励总额
	TakeAmountsBet   string // 领走打码返佣总额
	TakeTurnPrize    string // 领走转盘奖金总额
	TakeAmountsHead  string // 领走单个人头奖励总额（给头目）
	TakeAmountsChild string // 领走单个人头奖励总额（给团队）
	TakeAmountsTask  string // 领走累计人头奖励总额
}

// 代理后台-代理团队每日数量变化
type ShareAgentGroupVary struct {
	SDate             string // 日期
	Lv1Groups         int64  // 1级团队总个数
	Lv1DeadGroups     int64  // 1级死团队个数
	Lv1LiveGroups     int64  // 1级活团队个数
	Lv1GroupUsers     int64  // 1级团队总人数
	Lv1GroupAvgUsers  string // 1级团队队均人数
	Lv1DeadGroupUsers int64  // 1级死团队人数
	Lv1LiveGroupUsers int64  // 1级活团队人数
	Lv1GroupDeadRate  string // 1级团队死亡率
	Lv1GroupLossRate  string // 1级团队人数流失率

	Lv2Groups         int64  // 2级团队总个数
	Lv2DeadGroups     int64  // 2级死团队个数
	Lv2LiveGroups     int64  // 2级活团队个数
	Lv2GroupUsers     int64  // 2级团队总人数
	Lv2GroupAvgUsers  string // 2级团队队均人数
	Lv2DeadGroupUsers int64  // 2级死团队人数
	Lv2LiveGroupUsers int64  // 2级活团队人数
	Lv2GroupDeadRate  string // 2级团队死亡率
	Lv2GroupLossRate  string // 2级团队人数流失率

	Lv3Groups         int64  // 3级团队总个数
	Lv3DeadGroups     int64  // 3级死团队个数
	Lv3LiveGroups     int64  // 3级活团队个数
	Lv3GroupUsers     int64  // 3级团队总人数
	Lv3GroupAvgUsers  string // 3级团队队均人数
	Lv3DeadGroupUsers int64  // 3级死团队人数
	Lv3LiveGroupUsers int64  // 3级活团队人数
	Lv3GroupDeadRate  string // 3级团队死亡率
	Lv3GroupLossRate  string // 3级团队人数流失率

	Lv4Groups         int64  // 4级团队总个数
	Lv4DeadGroups     int64  // 4级死团队个数
	Lv4LiveGroups     int64  // 4级活团队个数
	Lv4GroupUsers     int64  // 4级团队总人数
	Lv4GroupAvgUsers  string // 4级团队队均人数
	Lv4DeadGroupUsers int64  // 4级死团队人数
	Lv4LiveGroupUsers int64  // 4级活团队人数
	Lv4GroupDeadRate  string // 4级团队死亡率
	Lv4GroupLossRate  string // 4级团队人数流失率
}

// UserCouponDate 满减券汇总
type UserCouponDate struct {
	SDate                  string `json:"sdate"`                  // 日期
	GiveCoupons            int64  `json:"giveCoupons"`            // 当日发放总张数
	MinRecharge            string `json:"minRecharge"`            // 当日发放总最低要求金额
	DerateAmounts          string `json:"derateAmounts"`          // 当日发放总减免金额
	UseCouponOrders        int64  `json:"useCouponOrders"`        // 当日用券支付总张数 x
	UseCouponOrderAmounts  string `json:"useCouponOrderAmounts"`  // 当日用券支付总金额 x
	UseCouponDerateAmounts string `json:"useCouponDerateAmounts"` // 当日用券减免总金额 x
	UseRate                string `json:"useRate"`                // 总发券使用率
	CurrentUseRate         string `json:"currentUseRate"`         // 当日总发当日即用率

	AutoGiveCoupons            int64  `json:"autoGiveCoupons"`            // 当日自动发放总张数
	AutoMinRecharge            string `json:"autoMinRecharge"`            // 当日自动发放总最低要求金额
	AutoDerateAmounts          string `json:"autoDerateAmounts"`          // 当日自动发放总减免金额
	AutoUseCouponOrders        int64  `json:"autoUseCouponOrders"`        // 当日用自动券支付总张数
	AutoUseCouponOrderAmounts  string `json:"autoUseCouponOrderAmounts"`  // 当日用自发券支付总金额
	AutoUseCouponDerateAmounts string `json:"autoUseCouponDerateAmounts"` // 当日用自发券减免总金额
	AutoUseRate                string `json:"autoUseRate"`                // 自发券使用率
	AutoCurrentUseRate         string `json:"autoCurrentUseRate"`         // 当日自发当日即用率

	HandGiveCoupons            int64  `json:"handGiveCoupons"`            // 当日手动发放总张数
	HandMinRecharge            string `json:"handMinRecharge"`            // 当日手动发放总最低要求金额
	HandDerateAmounts          string `json:"handDerateAmounts"`          // 当日手动发放总减免金额
	HandUseCouponOrders        int64  `json:"handUseCouponOrders"`        // 当日用手发券支付总张数
	HandUseCouponOrderAmounts  string `json:"handUseCouponOrderAmounts"`  // 当日用手发券支付总金额
	HandUseCouponDerateAmounts string `json:"handUseCouponDerateAmounts"` // 当日用手发券减免总金额
	HandUseRate                string `json:"handUseRate"`                // 手发券使用率
	HandCurrentUseRate         string `json:"handCurrentUseRate"`         // 当日手发当日即用率

	CurrentUseCoupons           int64 `json:"-"` // 当日发当日用张数
	AutoCurrentUseCoupons       int64 `json:"-"` // 当日发当日用张数
	HandCurrentUseCoupons       int64 `json:"-"` // 当日发当日用张数
	MinRecharge0                int64 `json:"-"` // 当日发放总最低要求金额
	DerateAmounts0              int64 `json:"-"` // 当日发放总减免金额
	UseCouponOrderAmounts0      int64 `json:"-"` // 当日用券支付总金额 x
	UseCouponDerateAmounts0     int64 `json:"-"` // 当日用券减免总金额 x
	AutoMinRecharge0            int64 `json:"-"` // 当日自动发放总最低要求金额
	AutoDerateAmounts0          int64 `json:"-"` // 当日自动发放总减免金额
	AutoUseCouponOrderAmounts0  int64 `json:"-"` // 当日用自发券支付总金额
	AutoUseCouponDerateAmounts0 int64 `json:"-"` // 当日用自发券减免总金额
	HandMinRecharge0            int64 `json:"-"` // 当日手动发放总最低要求金额
	HandDerateAmounts0          int64 `json:"-"` // 当日手动发放总减免金额
	HandUseCouponOrderAmounts0  int64 `json:"-"` // 当日用手发券支付总金额
	HandUseCouponDerateAmounts0 int64 `json:"-"` // 当日用手发券减免总金额
}
