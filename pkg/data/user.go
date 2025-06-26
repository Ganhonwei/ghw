package data

import (
	"fmt"
	"strconv"
	"time"

	"goserver/pkg/data/mq"
	"goserver/pkg/utils"

	"github.com/globalsign/mgo/bson"
)

const (
	NoveiceState   = 1
	NormalState    = 2
	ExceptionState = 3
	FrothState     = 4

	// 用户充值类型
	// NoneCharge   = 0 // 零充
	NewBie       = 0 // 新手
	Civilian     = 1 // 平民
	NormalCharge = 2 // 普R
	XR           = 3 // 小R
	ZR           = 4 // 中R
	DR           = 5 // 大R
	CDR          = 6 // 超大R
)

type User struct {
	Userid   string `bson:"_id" json:"userid"`          // 用户id
	Token    string `bson:"token" json:"token"`         // token
	Nickname string `bson:"nickname" json:"nickname"`   // 用户昵称
	RealName string `bson:"real_name" json:"real_name"` // 真实姓名
	Photo    string `bson:"photo" json:"photo"`         // 头像
	// Wxuid         string `bson:"wxuid" json:"wxuid"`                   // 微信uid
	// OpenID        string `bson:"openid" json:"openid"`                 //微信openid nolint
	// UnionID       string `bson:"unionid" json:"unionid"`               //微信unionid nolint
	Sex           uint32 `bson:"sex" json:"sex"`                       // 用户性别,男1 女2 非男非女3
	Phone         string `bson:"phone" json:"phone"`                   // 绑定的手机号码
	Phone2        string `bson:"phone2" json:"phone2"`                 // 提现绑定的手机号码
	Tourist       string `bson:"tourist" json:"tourist"`               // 游客
	Auth          string `bson:"auth" json:"auth"`                     // 密码验证码
	Password      string `bson:"password" json:"password"`             // MD5密码
	RegistIP      string `bson:"regist_ip" json:"regist_ip"`           // 注册账户时的IP地址
	LoginIP       string `bson:"login_ip" json:"login_ip"`             // 登录账户时的IP地址
	Diamond       int64  `bson:"diamond" json:"diamond"`               // 钻石(彩金cash)
	GiveDiamond   int64  `json:"giveDiamond" bson:"give_diamond"`      // 赠送彩金
	VBBank        int64  `bson:"vbbank" json:"vbbank"`                 // VB银行
	UnlockBonus   int64  `bson:"unlock_bonus" json:"unlock_bonus"`     // VB银行
	CashOutBonus  int64  `bson:"cashout_bonus" json:"cashout_bonus"`   // 已提现vb
	OutDiamond    int64  `json:"outDiamond" bson:"out_diamond"`        // 可提现彩金
	ShadowDiamond int64  `bson:"shadow_diamond" json:"shadow_diamond"` // 隐藏钻石(用户看不见的)
	Coin          int64  `bson:"coin" json:"coin"`                     // 金币(奖励金bonus)
	CashWater     uint64 `bson:"cash_water" json:"cash_water"`         // 彩金流水(游戏产生的)
	TransferCash  int64  `bson:"transfer_cash" json:"transfer_cash"`   // 转单补贴金
	Fraction      int64  `bson:"-" json:"-"`                           // 对战房娱乐模式分数
	PrivDiamond   int64  `bson:"priv_diamond" json:"priv_diamond"`     // 对战房赢的彩金记录
	// Chip          int64  `bson:"chip" json:"chip"`                     // 筹码
	// Card          int64  `bson:"card" json:"card"`                     // 房卡
	// Vip           uint32 `bson:"vip" json:"vip"`                       // vip
	Status                 int    `bson:"status" json:"status"`                                     // 正常1  锁定2  黑名单3 白名单4
	Robot                  bool   `bson:"robot" json:"robot"`                                       // 是否是机器人
	SimRobot               bool   `bson:"simulation_robot" json:"simulation_robot"`                 //模拟机器人
	LoginTimes             int32  `bson:"login_times" json:"login_times"`                           // 登录天数
	OnlineStatus           bool   `bson:"online_status" json:"online_status"`                       // 是否在线
	State                  int    `json:"state" bson:"state"`                                       //玩家状态 1.新手 2.正常 3.平民 4.泡沫
	RegistReward           bool   `json:"registReward" bson:"regist_reward"`                        //是否领取注册奖励
	RegistMode             int32  `json:"registMode" bson:"regist_mode"`                            //注册模式
	Platform               string `json:"platform" bson:"platform"`                                 //平台 g:谷歌 o:落地页 pc:PC端
	RegistArea             int    `json:"registArea" bson:"regist_area"`                            //ab测试 0:A 1:B 2:C
	VBOutDiamond           int64  `bson:"vb_out_diamond" json:"vbOutDiamond"`                       // VB可提现分数输分补偿累计
	VBOutDiamondDayTimes   int32  `bson:"vb_out_diamond_day_times" json:"vbOutDiamondDayTimes"`     // VB可提现金今日补偿次数
	VBOutDiamondRewardTime int64  `bson:"vb_out_diamond_reward_time" json:"vbOutDiamondRewardTime"` // 最后领取vb补偿时间戳秒
	// af参数
	AppId       string `bson:"app_id" json:"app_id"`
	AFId        string `bson:"af_id" json:"af_id"`
	OS          string `bson:"os" json:"os"`
	BundleId    string `bson:"bundle_id" json:"bundle_id"`
	RefGameId   string `bson:"ref_game_id" json:"ref_game_id"`
	RefPkgName  string `bson:"ref_pkg_name" json:"ref_pkg_name"`
	MediaSource string `bson:"media_source" json:"media_source"`
	AFKey       string `bson:"af_key" json:"af_key"`
	DeviceId    string `bson:"device_id" json:"device_id"`
	Channel     string `bson:"channel" json:"channel"`
	// ad参数
	AD_BundleId            string `json:"ad__bundle_id" bson:"ad__bundle_id"`
	AD_OsVersion           string `json:"ad__os_version" bson:"ad__os_version"`
	AD_DeviceId            string `json:"ad__device_id" bson:"ad__device_id"`
	AD_AppId               string `json:"ad__app_id" bson:"ad__app_id"`
	AD_RefGameId           string `json:"ad__ref_game_id" bson:"ad__ref_game_id"`
	AD_RefPkgName          string `json:"ad__ref_pkg_name" bson:"ad__ref_pkg_name"`
	AD_Channel             string `json:"ad__channel" bson:"ad__channel"`
	AD_Tracker_Token       string `json:"ad__tracker__token" bson:"ad__tracker__token"`
	AD_Tracker_Name        string `json:"ad__tracker__name" bson:"ad__tracker__name"`
	AD_Network             string `json:"ad__network" bson:"ad__network"`
	AD_Campaign            string `json:"ad__campaign" bson:"ad__campaign"`
	AD_ADGroup             string `json:"ad__ad_group" bson:"ad__ad_group"`
	AD_Creative            string `json:"ad__creative" bson:"ad__creative"`
	AD_Click_Lable         string `json:"ad__click__lable" bson:"ad__click__lable"`
	AD_ADID                string `json:"ad__adid" bson:"ad__adid"`
	AD_Cost_Type           string `json:"ad__cost__type" bson:"ad__cost__type"`
	AD_Cost_Amount         string `json:"ad__cost__amount" bson:"ad__cost__amount"`
	AD_Cost_Currency       string `json:"ad__cost__currency" bson:"ad__cost__currency"`
	AD_FB_Install_Referrer string `json:"ad__fb__install__referrer" bson:"ad__fb__install__referrer"`
	AD_Key                 string `json:"ad__key" bson:"ad__key"`
	AD_S2S_Code            string `json:"ad_s2s__code" bson:"ad_s2s__code"`
	AD_Event_Code1         string `json:"ad__event__code1" bson:"ad__event__code1"`
	AD_Event_Code2         string `json:"ad__event__code2" bson:"ad__event__code2"`
	AD_Event_Code3         string `json:"ad__event__code3" bson:"ad__event__code3"`
	AD_Event_Code4         string `json:"ad__event__code4" bson:"ad__event__code4"`
	AD_Event_Code5         string `json:"ad_Event_Code5" bson:"ad__event__code5"`
	AD_User_Agent          string `json:"ad__user__agent" bson:"ad__user__agent"`
	AD_Fake_User_Agent     string `json:"ad_Fake_User_Agent" bson:"ad__fake__user__agent"`
	AD_Tracker_Channel     string `json:"ad__tracker__channel" bson:"ad__tracker__channel"`
	//FB参数
	FB_Fbc string `json:"fb_Fbc" bson:"fb__fbc"`
	FB_Fbp string `json:"fb_Fbp" bson:"fb__fbp"`
	// 活动
	FirstRecharge       string               `bson:"first_recharge" json:"first_recharge"`              // 是否首冲
	Receive             bool                 `bson:"first_receive" json:"first_receive"`                // 今日是否领取首冲奖励
	ReceiveDay          int32                `bson:"first_receive_day" json:"first_receive_day"`        // 已领取多少天的首冲奖励
	NoviceGift          string               `bson:"novice_gift" json:"novice_gift"`                    // 是否购买新手礼包
	WeekCardMap         map[string]*WeekCard `bson:"weekcard,omitempty" json:"weekcard,omitempty"`      // 金银铜卡
	SignDay             uint64               `bson:"sign_day" json:"sign_day"`                          // 已领取的签到奖励
	OnlineReward        []int32              `bson:"online_reward" json:"online_reward"`                // 已领取的在线奖励
	WeeklyCardMap       map[int32]*WeekCard  `json:"weeklyCardMap" bson:"weekly_card_map"`              // 周卡
	BugFeedback         int                  `json:"bugFeedback" bson:"bug_feedback"`                   // bug反馈
	OverlimitedGift     []int32              `json:"overlimitedGift" bson:"overlimited_gift"`           // 过期或者买过的限时礼包
	LimitedGiftId       int32                `json:"limitedGiftId" bson:"limited_gift_id"`              // 限时礼包id
	LimitedGiftOverTime int64                `json:"limitedGiftOverTime" bson:"limited_gift_over_time"` // 过期时间
	CdkCodeMap          map[string]string    `json:"cdkCode" bson:"cdk_code"`                           // 领过的cdk码
	BreakingGift        []*BreakingGift      `json:"breakingGift" bson:"breaking_gift"`                 // 破产礼包
	BreakGtype          int32                `json:"breakGtype" bson:"-"`                               // 当前破产礼包gtype
	BreakGameId         string               `json:"breakGameId" bson:"-"`                              // 当前破产礼包gameId
	ActivatedPot        []ShopPot            `json:"activatedPot" bson:"activated_pot"`                 // 激活中的商城罐子
	PotFlow             int64                `json:"potFlow" bson:"pot_flow"`                           // 罐子打码量
	ScratchTicket       ScratchTickets       `json:"scratchTicket" bson:"scratch_ticket"`               // 大富翁刮刮乐
	PlayShareData       PlayShareAct         `json:"playShareData" bson:"play_share_data"`              // 玩游戏分享数据
	CouponIssueTime     map[string]int64     `json:"coupon_issue_time" bson:"coupon_issue_time"`        // 优惠券发放时间
	//时间
	Ctime           time.Time `bson:"ctime" json:"ctime"`                         // 注册时间
	LoginTime       time.Time `bson:"login_time" json:"login_time"`               // 最后登录时间
	LogoutTime      time.Time `bson:"logout_time" json:"logout_time"`             // 离线时间
	ResetTime       time.Time `bson:"reset_time" json:"reset_time"`               // 上次跨天时间
	LastResetTime   int64     `json:"lastResetTime" bson:"last_reset_time"`       // 上次跨天时间
	OnlineTime      uint64    `bson:"online_time" json:"online_time"`             // 当日在线时间(/s)
	TotalOnlineTime uint64    `json:"total_online_time" bson:"total_online_time"` // 总在线时间(/s)
	TotalGameTime   uint64    `json:"total_game_time" bson:"total_game_time"`     //总游戏时长
	StatusTime      time.Time `bson:"status_time" json:"status_time"`             // 状态时间
	FirstLoginToday bool      `json:"firstLoginToday" bson:"-"`                   // 今日首次登录
	BeKickoutTime   int64     `json:"beKickoutTime" bson:"-"`                     // 对战房最后被踢出时间
	//战绩
	Win        uint32                 `bson:"win" json:"win"`                 // 赢
	Lost       uint32                 `bson:"lost" json:"lost"`               // 输
	Ping       uint32                 `bson:"ping" json:"ping"`               // 平
	FreeWinMap map[int32][]FreeWin    `json:"freeWinMap" bson:"free_win_map"` // 百人战绩(20场)
	Round      uint32                 `json:"round" bson:"round"`             // 对局局数
	Hedge      int                    `json:"hedge" bson:"hedge"`             // 冤家牌触发局数
	LastGType  int32                  `json:"lastgType" bson:"lastg_type"`    // 最后一局游戏
	WinRates   map[int32]WinLoseRound `json:"winRates" bson:"win_rates"`      // 胜率 游戏->局数
	//充值提现
	Money           uint32                  `bson:"money" json:"money"`                       // 充值总金额(分)
	CashOut         int32                   `bson:"cash_out" json:"cash_out"`                 // 提现总金额(分)
	WithdrawLogMap  map[string]*WithdrawLog `bson:"withdraw_log" json:"withdraw_log"`         // 提现记录
	CommonChannel   uint32                  `json:"commonChannel" bson:"common_channel"`      // 常用充值渠道
	ToggleChannel   int                     `json:"toggleChannel" bson:"toggle_channel"`      // 切换支付渠道
	WithdrawLock    int32                   `json:"withdrawLock" bson:"withdraw_lock"`        // 解锁提现需要充值的金额
	RechargeTarge   []int                   `json:"rechargeTarge" bson:"recharge_targe"`      // 充值目标类型
	FirstChargeVal  int32                   `json:"firstChargeVal" bson:"first_charge_val"`   // 首次充值金额
	FirstChargeTime int64                   `json:"firstChargeTime" bson:"first_charge_time"` // 首次充值时间
	LastChargeTime  int64                   `json:"lastChargeTime" bson:"last_charge_time"`   // 最近充值时间
	ChargeTimes     int32                   `json:"chargeTimes" bson:"charge_times"`          // 充值次数
	HistoryVersion  int32                   `json:"historyVersion" bson:"history_version"`    // 用户初始化历史数据版本
	PkWithdrawInfo  PKWithdrawInfo          `json:"pkWithdrawInfo" bson:"pk_Withdraw_Info"`   // 巴基斯坦提现结果
	PayAvg          int64                   `json:"-" bson:"-"`                               // 登录时缓存均单价
	//最高
	TopDiamonds     int64 `bson:"top_diamonds" json:"top_diamonds"`         // 最高拥有钻石总金额
	TopGiveDiamonds int64 `json:"topGiveDiamonds" bson:"top_give_diamonds"` // 最高拥有赠送钻石总金额
	TopCoins        int64 `bson:"top_coins" json:"top_coins"`               // 最高拥有金币总金额
	// TopChips    int64 `bson:"top_chips" json:"top_chips"`       // 最高拥有筹码总金额
	// TopCards    int64 `bson:"top_cards" json:"top_cards"`       // 最高拥有房卡总数
	//单局
	TopWinDiamond int64 `bson:"top_win_diamond" json:"top_win_diamond"` // 单局赢最高钻石金额
	TopWinCoin    int64 `bson:"top_win_coin" json:"top_win_coin"`       // 单局赢最高金币金额
	// TopWinChip    int64 `bson:"top_win_chip" json:"top_win_chip"`       // 单局赢最高筹码金额
	// 提现次数
	WithDrawCountMap map[int32]int32 `bson:"withdraw_count" json:"withdraw_count"` // 提现次数
	//任务
	PhProgress int32                 `bson:"ph_progress" json:"ph_progress"`             // 牌型当前进度
	PhRewardId int32                 `bson:"ph_reward_id" json:"ph_reward_id"`           // 牌型当前id
	Task       map[int32]*TaskInfo   `bson:"task,omitempty" json:"task,omitempty"`       // 已经完成或者还在继续的任务
	VBTask     map[int32]*VBTaskInfo `bson:"vb_task,omitempty" json:"vb_task,omitempty"` // vip bank 已经完成或者还在继续的任务
	// 邮件
	FeedBackLogMap map[string]ChatLog `bson:"feedback_log_map" json:"feedback_log_map"` // 收到的回复
	// 聊天
	FeedBackTimes int32            `bson:"feedback_times" json:"feedback_times"` // 客服回复前发送了多少条
	OfflineChats  []OfflineChatLog `json:"offlineChats" bson:"offline_chats"`    // 离线消息
	// 分享
	ShareBelow    map[string]ShareData      `json:"shareBelow" bson:"share_below"`       // 分享的下级
	ShareBetLog   map[string]ShareAmountLog `json:"shareBetLog" bson:"share_bet_log"`    // 分享打码量记录
	ShareSuperior string                    `json:"shareSuperior" bson:"share_superior"` // 上级
	ShareSource   int32                     `json:"shareSource" bson:"share_source"`     // 上级分享来源: 1.转盘,2.代理
	ShareWithdraw int64                     `json:"shareWithdraw" bson:"share_withdraw"` // 可提金额
	ShareTotal    int64                     `json:"shareTotal" bson:"share_total"`       // 总的可提现金
	//引导弹窗
	GameGuildSlice  []*GameGuid `json:"gameGuildSlic" bson:"game_guild_slic"`    //游戏完成的引导切片
	WithdrawWindow  bool        `json:"withdrawWindow" bson:"withdraw_window"`   //第一次满100提现弹窗
	WithdrawWindow2 bool        `json:"withdrawWindow2" bson:"withdraw_window2"` //第一次满200提现弹窗
	//vip
	Vip UserVip `json:"vip" bson:"vip"` // vip
	//弹窗礼包
	GiftPopMap      map[string]GiftPop `json:"gift_pop" bson:"gift_pop"`                   // 弹窗礼包
	SkipGiftWelfare bool               `json:"skip_gift_welfare" bson:"skip_gift_welfare"` // 今日跳过充值礼包
	SkipTurntable   bool               `json:"skip_turntable" bson:"skip_turntable"`       // 今日跳过转盘
	SkipRank        bool               `json:"skip_rank" bson:"skip_rank"`                 // 今日跳过排行榜
	SkipGiftCode    bool               `json:"skip_gift_code" bson:"skip_gift_code"`       // 今日跳过礼包码
	//代理
	// Agent            string            `bson:"agent" json:"agent"`                                 // 绑定的代理ID
	// Atime            time.Time         `bson:"atime" json:"atime"`                                 // 绑定代理时间
	// AgentJoinTime    time.Time         `bson:"agent_join_time" json:"agent_join_time"`             // 申请成为代理时间
	// AgentState       uint32            `bson:"agent_state" json:"agent_state"`                     // 是否是代理状态1通过
	// AgentLevel       uint32            `bson:"agent_level" json:"agent_level"`                     // 代理等级,1，2，3，4
	// Build            uint32            `bson:"build" json:"build"`                                 // 下属绑定数量
	// AgentChild       uint32            `bson:"agent_child" json:"agent_child"`                     // 下属代理数量
	// BuildVaild       uint32            `bson:"build_vaild" json:"build_vaild"`                     // 下属有效绑定数量
	// AgentName        string            `bson:"agent_name" json:"agent_name"`                       // 代理名字
	// Weixin           string            `bson:"weixin" json:"weixin"`                               // 微信
	// ProfitRateSum    uint32            `bson:"profit_rate_sum" json:"profit_rate_sum"`             // 分佣比例总数
	// ProfitRate       map[string]uint32 `bson:"profit_rate,omitempty" json:"profit_rate,omitempty"` // 多线分佣比例
	// Profit           int64             `bson:"profit" json:"profit"`                               // 收益
	// ProfitMonth      int64             `bson:"profit_month" json:"profit_month"`                   // 月收益
	// ProfitLastMonth  int64             `bson:"profit_last_month" json:"profit_last_month"`         // 上月收益
	// Month            int               `bson:"month" json:"month"`                                 // 当前月
	// WeekProfit       int64             `bson:"week_profit" json:"week_profit"`                     // 本周收益
	// WeekPlayerProfit int64             `bson:"week_player_profit" json:"week_player_profit"`       // 本周玩家收益
	// WeekStart        time.Time         `bson:"week_start" json:"week_start"`                       // 每周日重置
	// WeekEnd          time.Time         `bson:"week_end" json:"week_end"`                           // 每周日重置
	// HistoryProfit    int64             `bson:"history_profit" json:"history_profit"`               // 历史收益
	// SubPlayerProfit  int64             `bson:"sub_player_profit" json:"sub_player_profit"`         // 下属玩家业绩收益
	// SubAgentProfit   int64             `bson:"sub_agent_profit" json:"sub_agent_profit"`           // 下属代理业绩收益
	// AgentNote        string            `bson:"agent_note" json:"agent_note"`                       // 代理备注
	// BringProfit      int64             `bson:"bring_profit" json:"bring_profit"`                   // 贡献收益
	// ProfitFirst      int64             `bson:"profit_first" json:"profit_first"`                   // 一级收益
	// ProfitSecond     int64             `bson:"profit_second" json:"profit_second"`                 // 二级收益
	//银行
	BackType          int32        `bson:"bank_type" json:"bank_type"`                     // 记录上次提现类型0bank,1ustd
	Bank              string       `bson:"bank" json:"bank"`                               // 个人银行 实际为bank name
	BankAccounts      string       `bson:"bank_accounts" json:"bank_accounts"`             // 银行卡号 实际为account number
	IFSC              string       `bson:"ifsc" json:"ifsc"`                               // IFSC
	BankAccountHolder string       `bson:"bank_account_holder" json:"bank_account_holder"` // 银行账户持有人?
	Email             string       `bson:"email" json:"email"`                             // 邮箱
	USDT              string       `bson:"usdt" json:"usdt"`                               // usdt 地址
	VBankLog          []VipBankLog `bson:"vbank_log" json:"vbank_log"`                     // VIP银行记录
	PKBank            []PKBankInfo `bson:"pk_bank" json:"pk_bank"`                         // pkbankinfo
	//登录
	// LoginTimes uint32 `bson:"login_times" json:"login_times"` //连续登录次数
	// LoginPrize uint32 `bson:"login_prize" json:"login_prize"` //连续登录奖励
	// LoginLoop  uint32 `bson:"login_loop" json:"login_loop"`   //连续登录循环
	Sign string `bson:"sign" json:"sign"` //个性签名
	//位置
	Lat     string `bson:"lat" json:"lat"`         //Latitude
	Lng     string `bson:"lng" json:"lng"`         //Longitude
	Address string `bson:"address" json:"address"` //Address
	//点控
	PCSwitch        bool  `bson:"point_control_switch"`         //点控开关
	PCFactor        int32 `bson:"point_control_factor"`         //系数设置
	PCScore         int64 `bson:"point_control_score"`          //分数设置(彩金)
	PCScoreComplete int64 `bson:"point_control_score_complete"` //已经达到的分数
	// 游戏
	CrashAutoLeave bool                `json:"crashAutoLeave" bson:"crash_auto_leave"` //crash自动撤离
	CrashMultiple  int32               `json:"crashMultiple" bson:"crash_multiple"`    //crash自动撤离倍数
	PlaneAutoLeave bool                `json:"planeAutoLeave" bson:"plane_auto_leave"` //plane自动撤离
	PlaneMultiple  int32               `json:"planeMultiple" bson:"plane_multiple"`    //plane自动撤离倍数
	CrashStrategy  CrashUserStrategy   `json:"crashStrategy" bson:"crash_strategy"`    //crash类策略
	AvStrategy     AviatorUserStrategy `json:"avStrategy" bson:"av_strategy"`          //Aviator类策略
	LHDStrategy    LHDUserStrategy     `json:"LHDStrategy" bson:"lhd_strategy"`        //LHD策略
	SevenStrategy  LHDUserStrategy     `json:"sevenStrategy" bson:"seven_strategy"`    //7up策略
	ABStrategy     ABUserStrategy      `json:"ABStrategy" bson:"ab_strategy"`          //AB策略
	CPStrategy     CPUserStrategy      `json:"CPStrategy" bson:"cp_strategy"`          //彩票策略
	RBStrategy     RBUserStrategy      `json:"RBStrategy" bson:"rb_strategy"`          //黑红策略
	RoundGames     map[int32]int32     `json:"round_games" bson:"round_games"`         //游戏对应局数
	// TP类策略相关
	TPTotalRound  int32 `json:"tp_total_round" bson:"tp_total_round"`   //累计局数
	TPTiggerTimes int32 `json:"tp_tigger_times" bson:"tp_tigger_times"` //触发次数
	//埋点
	NewbieGuid       bool `json:"newbie_guid" bson:"newbie_guid"`             //新手引导标记
	NewbieCondition1 bool `json:"newbie_condition1" bson:"newbie_condition1"` //新手状态1
	NewbieCondition2 bool `json:"newbie_condition2" bson:"newbie_condition2"` //新手状态2

	Kick105Flag      bool  `json:"kick105_flag" bson:"kick105_flag"`             //105踢出标记
	Kick200Flag      bool  `json:"kick200_flag" bson:"kick200_flag"`             //200踢出标记
	KickWithdrawFlag int   `json:"kick_withdraw_flag" bson:"kick_withdraw_flag"` //新手提现踢出标记
	WithdrawPOPTime  int64 `json:"withdraw_pop_time" bson:"withdraw_pop_time"`   //上次提现弹窗时间

	CustomTypes string `json:"customtypes" bson:"-"` // 自定义标记

	Strategy100Flag bool `json:"trigger_strategy100" bson:"trigger_strategy100"` //策略100标记
	Strategy200Flag bool `json:"trigger_strategy200" bson:"trigger_strategy200"` //策略200标记
	// B类玩家
	BetRecord       []int64 `json:"betRecord" bson:"bet_record"`         // 投注记录(200局)
	FluctuateLine   float64 `json:"fluctuateLine" bson:"fluctuate_line"` // 起伏线（非实时,后台展示用的）
	Withdrawable200 bool    `json:"withdraw200" bson:"withdraw200"`      // 200可提现标记

	// SurplusGame            int     `json:"surplusGame" bson:"surplus_game"`                        // 转平民余局数
	// SurplusGameTime        int64   `json:"surplusGameTime" bson:"surplus_game_time"`               // 转平民剩余时间
	Novice200StrategyCount int             `json:"novice200StrategyCount" bson:"novice200_strategy_count"` // B类新手200策略触发次数
	FreeWelfare            int             `json:"freeWelfare" bson:"free_welfare"`                        // 百人场免费福利
	BGiveCash              int64           `json:"bGiveCash" bson:"b_give_cash"`                           // B类新手充值获得的彩金
	ViceAccount            bool            `json:"viceAccount" bson:"vice_account"`                        // 小号标记 该账号是否有重复的adid或银行卡号
	TpNewbieProbeId        int32           `json:"tp_newbie_probe_id" bson:"tp_newbie_probe_id"`           // tp新手试探期id
	TpNewbieStableId       int32           `json:"tp_newbie_stable_id" bson:"tp_newbie_stable_id"`         // tp新手稳定局id
	TpNewbieStableRounds   map[int32]int32 `json:"tp_newbie_stable_rounds" bson:"tp_newbie_stable_rounds"` // tp新手稳定局策略局数

	//tp相关状态
	TpUserModel                 int32   `json:"tpUserModel" bson:"tp_user_model"`                                   //玩家所处模式 0 新手 1免费 2正常
	TpUserStage                 int32   `json:"tpUserStage" bson:"tp_user_stage"`                                   //玩家所处阶段
	TpUserStageHistory          []int32 `json:"tpUserStageHistory" bson:"tp_user_stage_history"`                    //玩家所处阶段历史
	TpUserRound                 int32   `json:"tpUserRound" bson:"tp_user_round"`                                   //局数
	TpUserWinOrLoseAmount       int64   `json:"tpUserWinOrLoseAmount" bson:"tp_user_win_or_lose_amount"`            //输赢额
	TpUserChangeCorrectionValue int64   `json:"tpUserChangeCorrectionValue" bson:"tp_user_change_correction_value"` //变化修正值 正数为多赢，负数为多输

	TpUserTodayChargeNum           int32              `json:"tpUserTodayChargeNum" bson:"tp_user_today_charge_num"`                        //今日充值次数
	TpUserTodayChargeAmount        int64              `json:"tpUserTodayChargeAmount" bson:"tp_user_today_charge_amount"`                  //今日充值金额
	TpUserTodayWithdrawNum         int32              `json:"tpUserTodayWithdrawNum" bson:"tp_user_today_withdraw_num"`                    //今日提现次数
	TpUserTodayWithdrawAmount      int64              `json:"tpUserTodayWithdrawAmount" bson:"tp_user_today_withdraw_amount"`              //今日提现金额
	TpUserTotalWinOrLoseAmount     int64              `json:"tpUserTotalWinOrLoseAmount" bson:"tp_user_total_win_or_lose_amount"`          //总输赢额
	TpUserChargeInGameNumOfTrigger int32              `json:"tpUserChargeInGameNumOfTrigger" bson:"tp_user_charge_in_game_num_of_trigger"` //局内充值触发数
	TpUserChargeInGameNumOfSuccess int32              `json:"tpUserChargeInGameNumOfSuccess" bson:"tp_user_charge_in_game_num_of_success"` //局内充值成功数
	TpUserFollowRateTrigger        map[int32]int32    `json:"tpUserFollowRateTrigger" bson:"tp_user_follow_rate_trigger"`                  //跟牌率触发数
	TpUserFollowRateSuccess        map[int32]int32    `json:"tpUserFollowRateSuccess" bson:"tp_user_follow_rate_success"`                  //跟牌率成功数
	TpUserStoryCD                  map[int32]int32    `json:"tpUserStoryCd" bson:"tp_user_story_cd"`                                       //剧情模式cd
	TpUserControlStrategyHistory   []int32            `json:"tpUserControlStrategyHistory" bson:"tp_user_control_strategy_history"`        //控制策略历史
	TpUserGameTime                 int64              `json:"tpUserGameTime" bson:"tp_user_game_time"`                                     //上次充值后游戏时长
	TpUserTodayGameRound           int32              `json:"tpUserTodayGameRound" bson:"tp_user_today_game_round"`                        //今日游戏局数
	TpUserTodayControlStrategyNum  map[int32]int32    `json:"tpUserTodayControlStrategyNum" bson:"tp_user_today_control_strategy_num"`     //今日控制策略触发次数
	TpUserChargeInGameStory        int32              `json:"tpUserChargeInGameStory" bson:"tp_user_charge_in_game_story"`                 //剧情局局内充值成功数
	TpUserStoryPlus                bool               `json:"tpUserStoryPlus" bson:"tp_user_story_plus"`                                   //tp用户剧情局plus
	TpHurtWinRound                 int32              `json:"tpHurtWinRound" bson:"tp_hurt_win_round"`                                     // tp 偷鸡赢钱局数
	TpBeHurtLoseRound              int32              `json:"tpBeHurtLoseRound" bson:"tp_be_hurt_lose_round"`                              // tp 被偷鸡输钱局数
	TpHurtWinScore                 int64              `json:"tpHurtWinScore" bson:"tp_hurt_win_score"`                                     // tp 偷鸡赢钱金额
	TpBeHurtLoseScore              int64              `json:"tpBeHurtLoseScore" bson:"tp_be_hurt_lose_score"`                              // tp 被偷鸡输钱金额
	TpUserVHRecord                 map[int32][6]int64 `json:"tpUserVHRecord" bson:"tp_user_vh_record"`                                     // tp玩家vh记录 <int32(牌型6bit, ante(25bit), [牌型,底注,牌型次数,下注总金额,赢次数,输次数]>
	TpHighCardRounds               int32              `json:"tpHighCardRounds" bson:"tp_high_card_rounds"`                                 // tp 连续拿高牌次数
	TpHeartbeatRound               int32              `json:"tpHeartbeatRound" bson:"tp_heartbeat_round"`                                  // tp怦然心动记录开始总玩局数
	TpHeartbeatResetTimes          int32              `json:"tpHeartbeatResetTimes" bson:"tp_heartbeat_reset_times"`                       // tp心跳重置次数
	TpUserLjsbPyCD                 int32              `json:"tpUserLjsbPyCD" bson:"tp_user_ljsb_py_cd"`                                    // tp乐极生悲策略被冤局cd
	TpUserGcyxCD                   int32              `json:"tpUserGcyxCD" bson:"tp_user_gcyx_cd"`                                         // tp高潮涌现策略cd
	TpUserGcyxHp                   int32              `json:"tpUserGcyxHp" bson:"tp_user_gcyx_hp"`                                         // tp高潮涌现策略当前概率
	TpUserGcyxWin                  int64              `json:"tpUserGcyxWin" bson:"tp_user_gcyx_win"`                                       // tp高潮涌现策略玩家赢分
	TpUserGcyxLose                 int64              `json:"tpUserGcyxLose" bson:"tp_user_gcyx_lose"`                                     // tp高潮涌现策略玩家输分

	// RM相关
	RmRoiLimits         map[string]int32 `json:"rm_roi_limits" bson:"rm_roi_limits"`                     // rm roi控制策略生效次数
	RmRoiDayLimits      map[string]int32 `json:"rm_roi_day_limits" bson:"rm_roi_day_limits"`             // rm 每个自然日，对应ID策略生效次数
	RmWithout1stDrop    int              `json:"rm_without_1st_drop" bson:"rm_without_1st_drop"`         // rm 玩家在没有1Life时候首回合弃牌次数
	RmWithout1stNotDrop int              `json:"rm_without_1st_not_drop" bson:"rm_without_1st_not_drop"` // rm 玩家在没有1Life时候首回合未弃牌次数
	// 回合数 -> WinRates
	Favorite []string `json:"favorite" bson:"favorite"` // 收藏游戏

	ActivityTurn               *ActivityTurn `json:"activityTurn" bson:"activity_turn"`                               // 转盘活动本轮活动数据
	ActivityTurnPrizeTimes     int32         `json:"activityTurnPrizeTimes" bson:"activity_turn_prize_times"`         // 转盘活动申请领奖次数
	ShareAgent                 ShareAgent    `json:"shareAgent" bson:"share_agent"`                                   // 代理活动
	VolatilitySubsidyTimes     int32         `json:"volatilitySubsidyTimes" bson:"volatility_subsidy_times"`          // 波动返水已补贴次数
	VolatilitySubsidyAmounts   int64         `json:"volatilitySubsidyAmounts" bson:"volatility_subsidy_amounts"`      // 波动返水已补贴额
	VolatilitySubsidyFirstTime int64         `json:"volatilitySubsidyFirstTime" bson:"volatility_subsidy_first_time"` // 波动返水首次补贴时间(毫秒)

	LastPayOrderBrushLimitTime int64 `json:"-" bson:"-"` // 上一次支付订单刷子弹窗时间戳秒

	//累充转盘
	CumRechargeWheelAmount int64 `json:"crw_recharge_amount" bson:"crw_recharge_amount"` //当日累计充值金额

}

// 数据库操作

func (u *User) Save() bool {
	defer DataProducers.PublishDatas(mq.TopicSyncUser, []any{u})

	return Upsert(PlayerUsers, bson.M{"_id": u.Userid}, u)
}

func (u *User) UpdateCurrency() bool {
	// 发布到kafka
	defer DataProducers.PublishDatas(mq.TopicSyncUserFinance, []any{&UserFinance{
		Userid:        u.Userid,
		Diamond:       u.Diamond,
		GiveDiamond:   u.GiveDiamond,
		VBOutDiamond:  u.VBOutDiamond,
		OutDiamond:    u.OutDiamond,
		ShadowDiamond: u.ShadowDiamond,
		Coin:          u.Coin,
		Money:         u.Money,
		CashOut:       u.CashOut,
		Ctime:         u.Ctime,
	}})

	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"diamond": u.Diamond, "coin": u.Coin, "vbbank": u.VBBank, "out_diamond": u.OutDiamond, "priv_diamond": u.PrivDiamond}})
}

func (u *User) UpdateGiveAndOut() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"give_diamond": u.GiveDiamond, "out_diamond": u.OutDiamond}})
}

func (u *User) UpdatePointControl() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"point_control_switch": u.PCSwitch, "point_control_factor": u.PCFactor, "point_control_score": u.PCScore, "point_control_score_complete": u.PCScoreComplete}})
}

func (u *User) UpdateBank() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"bank": u.Bank, "bank_accounts": u.BankAccounts, "ifsc": u.IFSC}})
}

func (u *User) UpdateStatus() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"status": u.Status}})
}

func (u *User) UpdateState() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"state": u.State}})
}

func (u *User) UpdateTask() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"task": u.Task}})
}

func (u *User) UpdateTransferCash() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"transfer_cash": u.TransferCash}})
}

// func (u *User) UpdateBankInfo() bool {
// 	return Update(PlayerUsers, bson.M{"_id": u.Userid},
// 		bson.M{"$set": bson.M{"bank": u.Bank,
// 			"bank_phone": u.BankPhone, "bank_password": u.BankPassword}})
// }

/*

func (u *User) UpdateLucky() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"lucky": u.Lucky}})
}

func (u *User) UpdateLogin() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"login_times": u.LoginTimes,
			"login_prize": u.LoginPrize, "login_time": u.LoginTime,
			"login_loop": u.LoginLoop, "login_ip": u.LoginIP}})
} */

func (u *User) UpdateSign() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"sign": u.Sign}})
}

func (u *User) UpdatePhone() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"phone": u.Phone, "auth": u.Auth, "password": u.Password}})
}

func (u *User) UpdatePhone2() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"phone2": u.Phone2, "real_name": u.RealName}})
}

func (u *User) UpdatePhoto() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"photo": u.Photo}})
}

func (u *User) UpdateOnlineReward() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"online_reward": u.OnlineReward, "online_time": u.OnlineTime}})
}

func (u *User) UpdateDailySign() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"sign_day": u.SignDay}})
}

func (u *User) UpdateNovice() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"novice_gift": u.NoviceGift}})
}

func (u *User) UpdateWeekCard() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"weekly_card_map": u.WeeklyCardMap}})
}

func (u *User) UpdateFirstRecharge() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"first_recharge": u.FirstRecharge, "first_receive": u.Receive, "first_receive_day": u.ReceiveDay}})
}

func (u *User) UpdateWithdrawLog() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"withdraw_log": u.WithdrawLogMap, "cash_out": u.CashOut, "withdraw_count": u.WithDrawCountMap, "vip": u.Vip}})
}

func (u *User) UpdatePKWithdrawInfo() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"pk_withdraw_info": u.PkWithdrawInfo}})
}

func (u *User) UpdateMoney() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"money": u.Money}})
}

func (u *User) UpdateChargeTimes() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"charge_times": u.ChargeTimes}})
}

func (u *User) UpdateFeedBack() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"feedback_times": u.FeedBackTimes, "feedback_log_map": u.FeedBackLogMap}})
}

func (u *User) UpdateOfflineLog() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"feedback_times": u.FeedBackTimes, "offline_chats": u.OfflineChats}})
}

func (u *User) UpdateShare() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"share_below": u.ShareBelow, "share_superior": u.ShareSuperior, "share_source": u.ShareSource, "share_bet_log": u.ShareBetLog, "share_total": u.ShareTotal, "share_withdraw": u.ShareWithdraw}})
}

func (u *User) UpdateRegistReward() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"regist_reward": u.RegistReward}})
}

func (u *User) UpdateCommonChannel() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"common_channel": u.CommonChannel, "toggle_channel": u.ToggleChannel}})
}

func (u *User) UpdateLimitedGift() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"overlimited_gift": u.OverlimitedGift, "limited_gift_id": u.LimitedGiftId, "limited_gift_over_time": u.LimitedGiftOverTime}})
}

func (u *User) UpdateBreakingGift() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"breaking_gift": u.BreakingGift}})
}

func (u *User) UpdateWithdrawLock() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"withdraw_lock": u.WithdrawLock, "first_charge_val": u.FirstChargeVal}})
}

func (u *User) UpdateRechargeTarge() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"recharge_targe": u.RechargeTarge}})
}

func (u *User) UpdateActivedPot() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"activated_pot": u.ActivatedPot}})
}

func (u *User) UpdateVIP() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"vip": u.Vip}})
}

func (u *User) UpdateVBankLog() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"vbank_log": u.VBankLog}})
}

func (u *User) UpdateBGiveCash() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"b_give_cash": u.BGiveCash}})
}

func (u *User) UpdatePlayShare() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"play_share_data": u.PlayShareData}})
}

func (u *User) UpdateUnlockBonus() bool {
	return Update(PlayerUsers, bson.M{"_id": u.Userid},
		bson.M{"$set": bson.M{"unlock_bonus": u.UnlockBonus}})
}

func (u *User) Get() {
	Get(PlayerUsers, u.Userid, u)
}

func (u *User) GetById(userid string) {
	GetByQ(PlayerUsers, bson.M{"_id": userid}, u)
}

func (u *User) GetByPhone() {
	GetByQ(PlayerUsers, bson.M{"phone": u.Phone}, u)
}

func (u *User) GetByTourist() {
	GetByQ(PlayerUsers, bson.M{"tourist": u.Tourist}, u)
}

// 密码验证
func (u *User) VerifyPwd(pwd string) bool {
	return utils.Md5(pwd+u.Auth) == u.Password
}

// 非数据库操作

// 获取彩金
func (u *User) GetDiamond() int64 {
	return u.Diamond
}

func (u *User) AddDiamond(num int64) {
	u.Diamond += num
	if u.Diamond < 0 {
		u.Diamond = 0
	}
	if u.Diamond > u.TopDiamonds {
		u.TopDiamonds = u.Diamond
	}
	if num > u.TopWinDiamond {
		u.TopWinDiamond = num
	}
}

// 增加可提现彩金
func (u *User) AddOutDiamond(num int64) {
	u.OutDiamond += num
	if u.OutDiamond < 0 {
		u.OutDiamond = 0
	}
	// u.OutDiamond = int64(math.Min(float64(u.OutDiamond), float64(u.Diamond)))
}

// 增加赠送彩金
func (u *User) AddGiveDiamond(num int64) {
	u.GiveDiamond += num
	if u.GiveDiamond < 0 {
		u.GiveDiamond = 0
	}
	if u.GiveDiamond > u.TopGiveDiamonds {
		u.TopGiveDiamonds = u.GiveDiamond
	}
}

// VB银行增加彩金
func (u *User) AddVBDiamond(num int64) {
	u.VBBank += num
	if u.VBBank < 0 {
		u.VBBank = 0
	}
	// if u.VBBank > u.TopGiveDiamonds {
	// 	u.TopGiveDiamonds = u.GiveDiamond
	// }
}

// 获取奖励金
func (u *User) GetCoin() int64 {
	return u.Coin
}

func (u *User) AddCoin(num int64) {
	u.Coin += num
	if u.Coin < 0 {
		u.Coin = 0
	}
	if u.Coin > u.TopCoins {
		u.TopCoins = u.Coin
	}
	if num > u.TopWinCoin {
		u.TopWinCoin = num
	}
}

// 获取分数：彩金+奖励金
func (u *User) GetScore() int64 {
	return u.GetCoin() + u.GetDiamond()
}

// 获取比例：彩金/(彩金+奖励金)
func (u *User) GetRatio() float64 {
	if u.GetScore() == 0 { //分母为0的情况下，彩金比例为0
		return 0
	}
	return float64(u.GetDiamond()) / float64(u.GetScore())
}

func (u *User) AddMoney(num uint32) {
	u.Money += num
}

func (u *User) AddChargeTimes() {
	u.ChargeTimes++
}

// 获取充值总额
func (u *User) GetMoney() uint32 {
	return u.Money
}

// 获取提现总额
func (u *User) GetCashOut() int32 {
	return u.CashOut
}

func (u *User) GetUserid() string {
	if u == nil {
		return ""
	}
	return u.Userid
}

func (u *User) GetToken() string {
	if u == nil {
		return ""
	}
	return u.Token
}

func (u *User) GetNickname() string {
	return u.Nickname
}

func (u *User) SetNickname(name string) {
	u.Nickname = name
}

func (u *User) SetEmail(email string) {
	u.Email = email
}

func (u *User) SetBankName(name string) {
	u.Bank = name
}

func (u *User) SetBankAccountHolder(name string) {
	u.BankAccountHolder = name
}

func (u *User) SetAccountNumber(name string) {
	u.BankAccounts = name
}

func (u *User) SetIFSC(name string) {
	u.IFSC = name
}

func (u *User) GetSex() uint32 {
	return u.Sex
}

func (u *User) GetPhoto() string {
	return u.Photo
}

func (u *User) SetPhoto(photo string) {
	u.Photo = photo
}

func (u *User) GetPhone() string {
	return u.Phone
}

func (u *User) GetTourist() string {
	return u.Tourist
}

func (u *User) GetRobot() bool {
	return u.Robot
}

func (u *User) IsTourist() bool {
	if u.GetPhone() != "" {
		return false
	}
	if u.GetTourist() != "" {
		return true
	}
	return false
}

func (u *User) GetTopDiamonds() int64 {
	return u.TopDiamonds
}

func (u *User) GetTopCoins() int64 {
	return u.TopCoins
}

func (u *User) GetTopWinDiamond() int64 {
	return u.TopWinDiamond
}

func (u *User) GetTopWinCoin() int64 {
	return u.TopWinCoin
}

func (u *User) GetRegistTime() time.Time {
	return u.Ctime
}

func (u *User) GetLoginTime() time.Time {
	return u.LoginTime
}

func (u *User) SetRecord(value int32) {
	if value > 0 {
		u.Win++
	} else if value < 0 {
		u.Lost++
	} else {
		u.Ping++
	}
}

func (u *User) AddCurrency(diamond, coin, card, chip, give, out int64) {
	u.AddDiamond(diamond)
	// 赠送金
	u.AddVBDiamond(give)
	// 提现
	u.AddOutDiamond(out)
	u.AddCoin(coin)
}

// func (u *User) GetBank() int64 {
// 	return u.Bank
// }

// func (u *User) AddBank(num int64) {
// 	u.Bank += num
// 	if u.Bank < 0 {
// 		u.Bank = 0
// 	}
// }

func (u *User) SetFraction(num int64) {
	u.Fraction = num
}

func (u *User) AddFraction(num int64) {
	u.Fraction += num
}

func (u *User) GetFraction() int64 {
	return u.Fraction
}

func (u *User) AddPrivDiamond(num int64) {
	u.PrivDiamond += num
}

func (u *User) GetPrivDiamond() int64 {
	return u.PrivDiamond
}

func (u *User) SetSign(content string) {
	u.Sign = content
}

func (u *User) GetSign() string {
	return u.Sign
}

// func (u *User) SetDeviceId(deviceId string) {
// 	u.DeviceId = deviceId
// }

func (u *User) SetBundleId(bundleId string) {
	u.BundleId = bundleId
}

func (u *User) SetOS(os string) {
	u.OS = os
}

func (u *User) SeAfId(afId string) {
	u.AFId = afId
}

func (u *User) SetMediaSource(mediaSource string) {
	u.MediaSource = mediaSource
}

func (u *User) SetAFKey(afKey string) {
	u.AFKey = afKey
}

// 获取总资产
func (u *User) GetAsset() int64 {
	return u.Diamond + u.ShadowDiamond + u.Coin
}

// 玩家系数
// func (u *User) GetFactor() float64 {
// 	money := u.GetMoney()
// 	if money == 0 {
// 		return 1
// 	}
// 	diamond := int64(math.Max(float64(u.GetDiamond()-u.PrivDiamond), 0))
// 	//(总彩金+总提现)/总充值/[0.99^(充值金额/10000)]
// 	m := math.Pow(0.993, float64(u.GetMoney())/10000)
// 	ret, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(diamond+int64(u.GetCashOut()))/float64(u.GetMoney())/m), 64)
// 	return math.Min(2, ret)
// }

// 起伏线
// func (u *User) FluctuateLine() float64 {
// 	betAvg := u.BetAvg()
// 	// 充值均价
// 	avgCharge := u.ChargeAvg()
// 	// 提现均价
// 	avgWithdraw := u.WithdrawAvg()
// 	line := betAvg*2 + float64(u.FirstChargeVal) + avgCharge - 40000 + (avgWithdraw-20000)*0.5

// 	// 如果起付线算出负值，那么起伏线=0
// 	return math.Max(0, line)
// }

func (u *User) BetAvg() float64 {
	var bets int64 = 0
	var round50Bets int64 = 0
	for i, v := range u.BetRecord {
		bets += v
		if i < 50 {
			round50Bets += v
		}
	}

	size := len(u.BetRecord)
	if size > 0 && size <= 50 {
		//平均下注值=玩家X局总下注÷总局数
		ret, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(bets)/float64(size)), 64)
		return ret
	}
	if size > 50 {
		//平均下注值=（玩家近50局总下注÷50）*0.5+[玩家50局外总下注÷（总局数-50局）]*0.5
		value := float64(round50Bets)/50*0.5 + float64(bets-round50Bets)/float64(size-50)*0.5
		ret, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
		return ret
	}

	return 0
}

func (u *User) ChargeAvg() float64 {
	var charges float64 = 0
	if len(u.RechargeTarge) <= 0 {
		return charges
	}
	avgCharge := float64(u.Money) / float64(len(u.RechargeTarge))
	charges, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", avgCharge), 64)
	return charges
}

func (u *User) WithdrawAvg() float64 {
	var avg float64 = 0
	if u.CashOut <= 0 {
		return avg
	}

	successCount := 0
	for _, w := range u.WithdrawLogMap {
		if w.Status == WithdrawSuccess {
			successCount++
		}
	}

	if successCount <= 0 {
		return avg
	}

	avg = float64(u.CashOut) / float64(successCount)
	avg, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", avg), 64)
	return avg
}

// 当前赢分
func (u *User) GetWinScore() int {
	return (int(u.CashOut) + int(u.OutDiamond)) - int(u.Money)
}

// 电话是否存在
func (u *User) ExistsPhone(phone string) bool {
	count := Count(PlayerUsers, bson.M{"phone": phone})
	return count > 0
}

// 获取注册天数
func (u *User) GetRegistDays() int {
	now := time.Now()
	duration := now.Sub(u.Ctime)
	days := int(duration.Hours() / 24)
	return days
}

// 增加tp类累计局数
func (u *User) IncreTPTotalRound() {
	u.TPTotalRound++
}

// 重置tp类累计局数
func (u *User) ResetTPTotalRound() {
	u.TPTotalRound = 0
}

// 增加tp类触发次数
func (u *User) IncreTPTiggerTimes() {
	u.TPTiggerTimes++
}

// 重置tp类触发次数
func (u *User) ResetTPTiggerTimes() {
	u.TPTiggerTimes = 0
}

func (u *User) SetStrategy100() {
	u.Strategy100Flag = true
}

func (u *User) SetStrategy200() {
	u.Strategy200Flag = true
	if u.RegistArea == 1 {
		u.Novice200StrategyCount++
	}
}

func (u *User) IsA() bool {
	return u.RegistArea == 0 && !u.Robot
}

func (u *User) IsB() bool {
	return u.RegistArea == 1 && !u.Robot
}

// tp用户阶段变化
func (u *User) TpUserStageChange(stage int32) bool {
	if u.TpUserStage != stage {
		u.TpUserStage = stage
		u.TpUserStageHistory = append(u.TpUserStageHistory, stage)
		return true
	} else {
		return false
	}
}

// 设备码是否存在
func ExistsAdid(ad__adid string, excludeUserids ...string) bool {
	m := bson.M{"ad__adid": ad__adid}
	if len(excludeUserids) > 0 {
		m["_id"] = bson.M{"$nin": excludeUserids}
	}
	count := Count(PlayerUsers, m)
	return count > 0
}

// 用户账户金额相关 clickhouse同步用
type UserFinance struct {
	Ver           int64     `bson:"ver" json:"ver"`                       // 插入时间戳
	Userid        string    `bson:"userid;primaryKey" json:"userid"`      // 用户id
	Diamond       int64     `bson:"diamond" json:"diamond"`               // 钻石(彩金cash)
	GiveDiamond   int64     `json:"giveDiamond" bson:"give_diamond"`      // 赠送彩金
	VBOutDiamond  int64     `bson:"vb_out_diamond" json:"vbOutDiamond"`   // VB可提现分数输分补偿累计
	OutDiamond    int64     `json:"outDiamond" bson:"out_diamond"`        // 可提现彩金
	ShadowDiamond int64     `bson:"shadow_diamond" json:"shadow_diamond"` // 隐藏钻石(用户看不见的)
	Coin          int64     `bson:"coin" json:"coin"`                     // 金币(奖励金bonus)
	Money         uint32    `bson:"money" json:"money"`                   // 充值总金额(分)
	CashOut       int32     `bson:"cash_out" json:"cash_out"`             // 提现总金额(分)
	Ctime         time.Time `bson:"ctime" json:"ctime"`                   // 注册时间
}
