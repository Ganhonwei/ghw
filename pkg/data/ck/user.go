package ck

import (
	"fmt"
	"goserver/pkg/data"
	"time"

	"github.com/bytedance/sonic"
)

// 用户表
type User struct {
	Ver      int64  `gorm:"column:ver" json:"ver"`                  // 插入时间戳
	Userid   string `gorm:"column:userid;primaryKey" json:"userid"` // 用户id
	Token    string `gorm:"column:token" json:"token"`              // token
	Nickname string `gorm:"column:nickname" json:"nickname"`        // 用户昵称
	RealName string `gorm:"column:real_name" json:"real_name"`      // 真实姓名
	Photo    string `gorm:"column:photo" json:"photo"`              // 头像
	Sex      uint32 `gorm:"column:sex" json:"sex"`                  // 用户性别,男1 女2 非男非女3
	Phone    string `gorm:"column:phone" json:"phone"`              // 绑定的手机号码
	Phone2   string `gorm:"column:phone2" json:"phone2"`            // 提现绑定的手机号码
	Tourist  string `gorm:"column:tourist" json:"tourist"`          // 游客
	Auth     string `gorm:"column:auth" json:"auth"`                // 密码验证码
	Password string `gorm:"column:password" json:"password"`        // MD5密码
	RegistIP string `gorm:"column:regist_ip" json:"regist_ip"`      // 注册账户时的IP地址
	LoginIP  string `gorm:"column:login_ip" json:"login_ip"`        // 登录账户时的IP地址
	// 实时账户查询 UserFinance ======
	Diamond                int64  `gorm:"column:diamond" json:"diamond"`                                   // 钻石(彩金cash)
	GiveDiamond            int64  `json:"giveDiamond" gorm:"column:give_diamond"`                          // 赠送彩金
	VBBank                 int64  `gorm:"column:vbbank" json:"vbbank"`                                     // VB银行
	UnlockBonus            int64  `gorm:"column:unlock_bonus" json:"unlock_bonus"`                         // VB银行
	CashOutBonus           int64  `gorm:"column:cashout_bonus" json:"cashout_bonus"`                       // 已提现vb
	OutDiamond             int64  `json:"outDiamond" gorm:"column:out_diamond"`                            // 可提现彩金
	ShadowDiamond          int64  `gorm:"column:shadow_diamond" json:"shadow_diamond"`                     // 隐藏钻石(用户看不见的)
	Coin                   int64  `gorm:"column:coin" json:"coin"`                                         // 金币(奖励金bonus)
	Money                  uint32 `gorm:"column:money" json:"money"`                                       // 充值总金额(分)
	CashOut                int32  `gorm:"column:cash_out" json:"cash_out"`                                 // 提现总金额(分)
	Profit                 int64  `gorm:"column:profit" json:"profit"`                                     // 利润 (提现+携带)-充值
	CashWater              uint64 `gorm:"column:cash_water" json:"cash_water"`                             // 彩金流水(游戏产生的)
	PrivDiamond            int64  `gorm:"column:priv_diamond" json:"priv_diamond"`                         // 对战房赢的彩金记录
	Status                 int    `gorm:"column:status" json:"status"`                                     // 正常1  锁定2  黑名单3 白名单4
	Robot                  bool   `gorm:"column:robot" json:"robot"`                                       // 是否是机器人
	SimRobot               bool   `gorm:"column:simulation_robot" json:"simulation_robot"`                 //模拟机器人
	LoginTimes             int32  `gorm:"column:login_times" json:"login_times"`                           // 登录天数
	OnlineStatus           bool   `gorm:"column:online_status" json:"online_status"`                       // 是否在线
	State                  int    `json:"state" gorm:"column:state"`                                       //玩家状态 1.新手 2.正常 3.平民 4.泡沫
	RegistReward           bool   `json:"registReward" gorm:"column:regist_reward"`                        //是否领取注册奖励
	RegistMode             int32  `json:"registMode" gorm:"column:regist_mode"`                            //注册模式
	Platform               string `json:"platform" gorm:"column:platform"`                                 //平台 g:谷歌 o:落地页 pc:PC端
	RegistArea             int    `json:"registArea" gorm:"column:regist_area"`                            //ab测试 0:A 1:B 2:C
	VBOutDiamond           int64  `gorm:"column:vb_out_diamond" json:"vbOutDiamond"`                       // VB可提现分数输分补偿累计
	VBOutDiamondDayTimes   int32  `gorm:"column:vb_out_diamond_day_times" json:"vbOutDiamondDayTimes"`     // VB可提现金今日补偿次数
	VBOutDiamondRewardTime int64  `gorm:"column:vb_out_diamond_reward_time" json:"vbOutDiamondRewardTime"` // 最后领取vb补偿时间戳秒
	// af参数
	AppId       string `gorm:"column:app_id" json:"app_id"`
	AFId        string `gorm:"column:af_id" json:"af_id"`
	OS          string `gorm:"column:os" json:"os"`
	BundleId    string `gorm:"column:bundle_id" json:"bundle_id"`
	RefGameId   string `gorm:"column:ref_game_id" json:"ref_game_id"`
	RefPkgName  string `gorm:"column:ref_pkg_name" json:"ref_pkg_name"`
	MediaSource string `gorm:"column:media_source" json:"media_source"`
	AFKey       string `gorm:"column:af_key" json:"af_key"`
	DeviceId    string `gorm:"column:device_id" json:"device_id"`
	Channel     string `gorm:"column:channel" json:"channel"`
	// ad参数
	AD_BundleId            string `json:"ad__bundle_id" gorm:"column:ad__bundle_id"`
	AD_OsVersion           string `json:"ad__os_version" gorm:"column:ad__os_version"`
	AD_DeviceId            string `json:"ad__device_id" gorm:"column:ad__device_id"`
	AD_AppId               string `json:"ad__app_id" gorm:"column:ad__app_id"`
	AD_RefGameId           string `json:"ad__ref_game_id" gorm:"column:ad__ref_game_id"`
	AD_RefPkgName          string `json:"ad__ref_pkg_name" gorm:"column:ad__ref_pkg_name"`
	AD_Channel             string `json:"ad__channel" gorm:"column:ad__channel"`
	AD_Tracker_Token       string `json:"ad__tracker__token" gorm:"column:ad__tracker__token"`
	AD_Tracker_Name        string `json:"ad__tracker__name" gorm:"column:ad__tracker__name"`
	AD_Network             string `json:"ad__network" gorm:"column:ad__network"`
	AD_Campaign            string `json:"ad__campaign" gorm:"column:ad__campaign"`
	AD_ADGroup             string `json:"ad__ad_group" gorm:"column:ad__ad_group"`
	AD_Creative            string `json:"ad__creative" gorm:"column:ad__creative"`
	AD_Click_Lable         string `json:"ad__click__lable" gorm:"column:ad__click__lable"`
	AD_ADID                string `json:"ad__adid" gorm:"column:ad__adid"`
	AD_Cost_Type           string `json:"ad__cost__type" gorm:"column:ad__cost__type"`
	AD_Cost_Amount         string `json:"ad__cost__amount" gorm:"column:ad__cost__amount"`
	AD_Cost_Currency       string `json:"ad__cost__currency" gorm:"column:ad__cost__currency"`
	AD_FB_Install_Referrer string `json:"ad__fb__install__referrer" gorm:"column:ad__fb__install__referrer"`
	AD_Key                 string `json:"ad__key" gorm:"column:ad__key"`
	AD_S2S_Code            string `json:"ad_s2s__code" gorm:"column:ad_s2s__code"`
	AD_Event_Code1         string `json:"ad__event__code1" gorm:"column:ad__event__code1"`
	AD_Event_Code2         string `json:"ad__event__code2" gorm:"column:ad__event__code2"`
	AD_Event_Code3         string `json:"ad__event__code3" gorm:"column:ad__event__code3"`
	AD_Event_Code4         string `json:"ad__event__code4" gorm:"column:ad__event__code4"`
	AD_Event_Code5         string `json:"ad_Event_Code5" gorm:"column:ad__event__code5"`
	AD_User_Agent          string `json:"ad__user__agent" gorm:"column:ad__user__agent"`
	AD_Fake_User_Agent     string `json:"ad_Fake_User_Agent" gorm:"column:ad__fake__user__agent"`
	AD_Tracker_Channel     string `json:"ad__tracker__channel" gorm:"column:ad__tracker__channel"`
	//FB参数
	FB_Fbc string `json:"fb_Fbc" gorm:"column:fb__fbc"`
	FB_Fbp string `json:"fb_Fbp" gorm:"column:fb__fbp"`
	//时间
	Ctime           time.Time `gorm:"column:ctime" json:"ctime"`                         // 注册时间
	LoginTime       time.Time `gorm:"column:login_time" json:"login_time"`               // 最后登录时间
	LogoutTime      time.Time `gorm:"column:logout_time" json:"logout_time"`             // 离线时间
	ResetTime       time.Time `gorm:"column:reset_time" json:"reset_time"`               // 上次跨天时间
	LastResetTime   int64     `json:"lastResetTime" gorm:"column:last_reset_time"`       // 上次跨天时间
	OnlineTime      uint64    `gorm:"column:online_time" json:"online_time"`             // 当日在线时间(/s)
	TotalOnlineTime uint64    `json:"total_online_time" gorm:"column:total_online_time"` // 总在线时间(/s)
	TotalGameTime   uint64    `json:"total_game_time" gorm:"column:total_game_time"`     //总游戏时长
	StatusTime      time.Time `gorm:"column:status_time" json:"status_time"`             // 状态时间
	//战绩
	Win       uint32 `gorm:"column:win" json:"win"`              // 赢
	Lost      uint32 `gorm:"column:lost" json:"lost"`            // 输
	Ping      uint32 `gorm:"column:ping" json:"ping"`            // 平
	Round     uint32 `json:"round" gorm:"column:round"`          // 对局局数
	Hedge     int    `json:"hedge" gorm:"column:hedge"`          // 冤家牌触发局数
	LastGType int32  `json:"lastgType" gorm:"column:lastg_type"` // 最后一局游戏
	//充值提现
	CommonChannel   uint32  `json:"commonChannel" gorm:"column:common_channel"`                   // 常用充值渠道
	ToggleChannel   int     `json:"toggleChannel" gorm:"column:toggle_channel"`                   // 切换支付渠道
	WithdrawLock    int32   `json:"withdrawLock" gorm:"column:withdraw_lock"`                     // 解锁提现需要充值的金额
	RechargeTarge   []int32 `json:"rechargeTarge" gorm:"column:recharge_targe;type:Array(Int32)"` // 充值目标类型
	FirstChargeVal  int32   `json:"firstChargeVal" gorm:"column:first_charge_val"`                // 首次充值金额
	FirstChargeTime int64   `json:"firstChargeTime" gorm:"column:first_charge_time"`              // 首次充值时间
	LastChargeTime  int64   `json:"lastChargeTime" gorm:"column:last_charge_time"`                // 最近充值时间
	ChargeTimes     int32   `json:"chargeTimes" gorm:"column:charge_times"`                       // 充值次数
	HistoryVersion  int32   `json:"historyVersion" gorm:"column:history_version"`                 // 用户初始化历史数据版本
	//最高
	TopDiamonds     int64 `gorm:"column:top_diamonds" json:"top_diamonds"`         // 最高拥有钻石总金额
	TopGiveDiamonds int64 `json:"topGiveDiamonds" gorm:"column:top_give_diamonds"` // 最高拥有赠送钻石总金额
	TopCoins        int64 `gorm:"column:top_coins" json:"top_coins"`               // 最高拥有金币总金额
	//单局
	TopWinDiamond int64 `gorm:"column:top_win_diamond" json:"top_win_diamond"` // 单局赢最高钻石金额
	TopWinCoin    int64 `gorm:"column:top_win_coin" json:"top_win_coin"`       // 单局赢最高金币金额
	// 提现次数
	WithDrawCountMap map[int32]int32 `gorm:"column:withdraw_count;type:Map(Int32,Int32)" json:"withdraw_count"` // 提现次数
	// 分享
	ShareSuperior string `json:"shareSuperior" gorm:"column:share_superior"` // 上级
	ShareSource   int32  `json:"shareSource" gorm:"column:share_source"`     // 上级分享来源: 1.转盘,2.代理
	ShareWithdraw int64  `json:"shareWithdraw" gorm:"column:share_withdraw"` // 可提金额
	ShareTotal    int64  `json:"shareTotal" gorm:"column:share_total"`       // 总的可提现金
	//引导弹窗
	WithdrawWindow  bool `json:"withdrawWindow" gorm:"column:withdraw_window"`   //第一次满100提现弹窗
	WithdrawWindow2 bool `json:"withdrawWindow2" gorm:"column:withdraw_window2"` //第一次满200提现弹窗
	//vip
	VipLv            int     `json:"vip_lv" gorm:"column:vip_lv"`                                      //当前等级
	VipMaxLv         int     `json:"vip_max_lv" gorm:"column:vip_max_lv"`                              //历史最高等级
	VipDaily         bool    `json:"vip_daily" gorm:"column:vip_daily"`                                //每日奖励
	VipWeekly        bool    `json:"vip_weekly" gorm:"column:vip_weekly"`                              //每周奖励
	VipWeeklyTime    int64   `json:"vip_weeklyTime" gorm:"column:vip_weekly_time"`                     //下次领取周领奖时间
	VipWeekday       int     `json:"vip_weekday" gorm:"column:vip_weekday"`                            //周几
	VipLevelReward   []int32 `json:"vip_levelReward" gorm:"column:vip_level_reward;type:Array(Int32)"` //升级奖励
	VipExp           int64   `json:"vip_exp" gorm:"column:vip_exp"`                                    //当前经验值
	VipWithdrawCount int     `json:"vip_withdrawCount" gorm:"column:vip_withdraw_count"`               //每日提现次数
	VipVersion       int     `json:"vip_version" gorm:"column:vip_version"`                            //版本
	//银行
	BackType     int32  `gorm:"column:bank_type" json:"bank_type"`         // 记录上次提现类型0bank,1ustd
	Bank         string `gorm:"column:bank" json:"bank"`                   // 个人银行
	BankAccounts string `gorm:"column:bank_accounts" json:"bank_accounts"` // 银行卡号
	IFSC         string `gorm:"column:ifsc" json:"ifsc"`                   // IFSC
	USDT         string `gorm:"column:usdt" json:"usdt"`                   // usdt 地址
	Sign         string `gorm:"column:sign" json:"sign"`                   //个性签名
	//位置
	Lat     string `gorm:"column:lat" json:"lat"`         //Latitude
	Lng     string `gorm:"column:lng" json:"lng"`         //Longitude
	Address string `gorm:"column:address" json:"address"` //Address
	//点控
	PCSwitch        bool  `gorm:"column:point_control_switch"`         //点控开关
	PCFactor        int32 `gorm:"column:point_control_factor"`         //系数设置
	PCScore         int64 `gorm:"column:point_control_score"`          //分数设置(彩金)
	PCScoreComplete int64 `gorm:"column:point_control_score_complete"` //已经达到的分数
	// 游戏
	CrashMultiple  int32           `json:"crashMultiple" gorm:"column:crash_multiple"`                  //crash自动撤离倍数
	PlaneAutoLeave bool            `json:"planeAutoLeave" gorm:"column:plane_auto_leave"`               //plane自动撤离
	PlaneMultiple  int32           `json:"planeMultiple" gorm:"column:plane_multiple"`                  //plane自动撤离倍数
	RoundGames     map[int32]int32 `json:"round_games" gorm:"column:round_games;type:Map(Int32,Int32)"` //游戏对应局数
	// AvStrategy     AviatorUserStrategy `json:"avStrategy" bson:"av_strategy"`                               //Aviator类策略
	AvFYZSTiggerTimes      int     `json:"-" gorm:"column:av_fyzs_tigger_times"`                      // 扶摇直上:触发次数
	AvFYZSEvoTimes         int     `json:"-" gorm:"column:av_fyzs_evo_times"`                         // 扶摇直上:当次玩的局数(与playtimes关联)
	AvFYZSAllTiggerTimes   int     `json:"-" gorm:"column:av_fyzs_all_tigger_times"`                  // 扶摇直上:累计触发次数
	AvFYZSAllEvoTimes      int     `json:"-" gorm:"column:av_fyzs_all_evo_times"`                     // 扶摇直上:累计玩游戏次数
	AvFYZSWinScore         int64   `json:"-" gorm:"column:av_fyzs_win_score"`                         // 扶摇直上:赢分
	AvYHWMMonitorRounds    int     `json:"-" gorm:"column:av_yhwm_monitor_rounds"`                    // 欲薅无门:监控胜局(连胜不能断)
	AvYHWMAviatorMulpitles []int32 `json:"-" gorm:"column:av_yhwm_crash_mulpitles;type:Array(Int32)"` // 欲薅无门:逃跑倍数
	AvYHWMTiggerTimes      int     `json:"-" gorm:"column:av_yhwm_tigger_times"`                      // 欲薅无门:触发次数
	AvYHWMWinScore         int64   `json:"-" gorm:"column:av_yhwm_win_score"`                         // 欲薅无门:赢分
	AvYHWMRecyleScore      int64   `json:"-" gorm:"column:av_yhwm_recyle_score"`                      // 欲薅无门:回收分数
	AvYHWMAllRecyleScore   int64   `json:"-" gorm:"column:av_yhwm_all_recyle_score"`                  // 欲薅无门:累计回收分数
	AvYHWMTigger           bool    `json:"-" gorm:"column:av_yhwm_tigger"`                            // 欲薅无门:是否生效中
	AvQSHSWinRounds        []int32 `json:"-" gorm:"column:av_qshs_win_rounds;type:Array(Int32)"`      // 赢局逃跑倍数
	AvQSHSTriggerTimes     int     `json:"-" gorm:"column:av_qshs_trigger_times"`                     // 起死回生：生效次数
	AvJCFKTriggerTimes     int     `json:"-" gorm:"column:av_jcfk_trigger_times"`                     // 奖池风控:触发jackpot次数
	AvJCFKJackpotVal       int64   `json:"-" gorm:"column:av_jcfk_jackpot_val"`                       // 奖池风控:获得的jackpot值
	AvMXJLTriggerTimes     int     `json:"-" gorm:"column:av_mxjl_trigger_times"`                     // 冒险奖励:触发次数
	AvMXJLBets             []int64 `json:"-" gorm:"column:av_mxjl_bets;type:Array(Int64)"`            // 冒险奖励:下注
	AvRKYHTriggerTimes     int     `json:"-" gorm:"column:av_rkyh_trigger_times"`                     // 触发rkyh次数
	AvRKYHWinMulpitles     []int32 `json:"-" gorm:"column:av_rkyh_win_mulpitles;type:Array(Int32)"`   // 前n赢局的局逃跑倍数
	AvGcyxEvoTimes         int     `json:"-" gorm:"column:av_gcyx_evo_times"`                         // 高潮涌现:当次玩的局数
	AvGcyxIsStateGC        bool    `json:"-" gorm:"column:av_gcyx_is_state_gc"`                       // 高潮涌现:是否高潮状态
	AvGcyxIsStateXZ        bool    `json:"-" gorm:"column:av_gcyx_is_state_xz"`                       // 高潮涌现:是否贤者状态
	AvGcyxXZTimes          int     `json:"-" gorm:"column:av_gcyx_xz_times"`                          // 高潮涌现:贤者状态次数
	AvGcyxWinScore         int64   `json:"-" gorm:"column:av_gcyx_win_score"`                         // 高潮涌现:赢分
	AvGcyxM                int     `json:"-" gorm:"column:av_gcyx_m"`                                 // 高潮涌现:高潮次数m
	AvGcyxR                int     `json:"-" gorm:"column:av_gcyx_r"`                                 // 高潮涌现:高潮局数r
	AvGcyxDMRecord         []int64 `json:"-" gorm:"column:av_gcyx_dm_record;type:Array(Int64)"`       // 高潮涌现:打码记录
	AvGcyxDailyGCTimes     int     `json:"-" gorm:"column:av_gcyx_gc_times"`                          // 高潮涌现:每日高潮次数
	AvGcyxAvgBet           int64   `json:"-" gorm:"column:av_gcyx_avg_bet"`                           // 高潮涌现:均码量
	// CrashStrategy  CrashUserStrategy `json:"crashStrategy" gorm:"column:crash_strategy"`    //crash类策略
	CrashFYZSPlayTimes      int     `json:"-" gorm:"column:crash_fyzs_play_times"`                        // 扶摇直上:玩游戏次数(退出游戏才算一次)
	CrashFYZSTiggerTimes    int     `json:"-" gorm:"column:crash_fyzs_tigger_times"`                      // 扶摇直上:触发次数
	CrashFYZSEvoTimes       int     `json:"-" gorm:"column:crash_fyzs_evo_times"`                         // 扶摇直上:当次玩的局数(与playtimes关联)
	CrashFYZSAllTiggerTimes int     `json:"-" gorm:"column:crash_fyzs_all_tigger_times"`                  // 扶摇直上:累计触发次数
	CrashFYZSAllEvoTimes    int     `json:"-" gorm:"column:crash_fyzs_all_evo_times"`                     // 扶摇直上:累计玩游戏次数
	CrashYHWMMonitorRounds  int     `json:"-" gorm:"column:crash_yhwm_monitor_rounds"`                    // 欲薅无门:监控胜局(连胜不能断)
	CrashYHWMCrashMulpitles []int32 `json:"-" gorm:"column:crash_yhwm_crash_mulpitles;type:Array(Int32)"` // 欲薅无门:逃跑倍数
	CrashYHWMTiggerTimes    int     `json:"-" gorm:"column:crash_yhwm_tigger_times"`                      // 欲薅无门:触发次数
	CrashYHWMWinScore       int64   `json:"-" gorm:"column:crash_yhwm_win_score"`                         // 欲薅无门:赢分
	CrashYHWMRecyleScore    int64   `json:"-" gorm:"column:crash_yhwm_recyle_score"`                      // 欲薅无门:回收分数
	CrashYHWMAllRecyleScore int64   `json:"-" gorm:"column:crash_yhwm_all_recyle_score"`                  // 欲薅无门:累计回收分数
	CrashYHWMTigger         bool    `json:"-" gorm:"column:crash_yhwm_tigger"`                            // 欲薅无门:是否生效中
	CrashQSHSWinRounds      []int32 `json:"-" gorm:"column:crash_qshs_win_rounds;type:Array(Int32)"`      // 赢局逃跑倍数
	CrashQSHSTriggerTimes   int     `json:"-" gorm:"column:crash_qshs_trigger_times"`                     // 起死回生：生效次数
	CrashJCFKTriggerTimes   int     `json:"-" gorm:"column:crash_jcfk_trigger_times"`                     // 奖池风控:触发jackpot次数
	CrashJCFKJackpotVal     int64   `json:"-" gorm:"column:crash_jcfk_jackpot_val"`                       // 奖池风控:获得的jackpot值
	CrashMXJLTriggerTimes   int     `json:"-" gorm:"column:crash_mxjl_trigger_times"`                     // 冒险奖励:触发次数
	CrashMXJLBets           []int64 `json:"-" gorm:"column:crash_mxjl_bets;type:Array(Int64)"`            // 冒险奖励:下注
	CrashRKYHTriggerTimes   int     `json:"-" gorm:"column:crash_rkyh_trigger_times"`                     // 触发rkyh次数
	CrashRKYHWinMulpitles   []int32 `json:"-" gorm:"column:crash_rkyh_win_mulpitles;type:Array(Int32)"`   // 前n赢局的局逃跑倍数
	CrashEvoTimes           int     `json:"-" gorm:"column:crash_gcyx_evo_times"`                         // 高潮涌现:当次玩的局数
	CrashIsStateGC          bool    `json:"-" gorm:"column:crash_gcyx_is_state_gc"`                       // 高潮涌现:是否高潮状态
	CrashIsStateXZ          bool    `json:"-" gorm:"column:crash_gcyx_is_state_xz"`                       // 高潮涌现:是否贤者状态
	CrashXZTimes            int     `json:"-" gorm:"column:crash_gcyx_xz_times"`                          // 高潮涌现:贤者状态次数
	CrashWinScore           int64   `json:"-" gorm:"column:crash_gcyx_win_score"`                         // 高潮涌现:赢分
	CrashM                  int     `json:"-" gorm:"column:crash_gcyx_m"`                                 // 高潮涌现:高潮次数m
	CrashR                  int     `json:"-" gorm:"column:crash_gcyx_r"`                                 // 高潮涌现:高潮局数r
	CrashDMRecord           []int64 `json:"-" gorm:"column:crash_gcyx_dm_record;type:Array(Int64)"`       // 高潮涌现:打码记录
	CrashDailyGCTimes       int     `json:"-" gorm:"column:crash_gcyx_gc_times"`                          // 高潮涌现:每日高潮次数
	CrashGcyxAvgBet         int64   `json:"-" gorm:"column:crash_gcyx_avg_bet"`                           // 高潮涌现:均码量
	// LHDStrategy    LHDUserStrategy   `json:"LHDStrategy" gorm:"column:lhd_strategy"`        //LHD策略
	LHDRoundBet            []int64 `json:"-" gorm:"column:lhd_round_bet;type:Array(Int64)"`      // 每轮下注
	LHDXXSCDisturbCoolDown int     `json:"-" gorm:"column:lhd_xxsc_disturb_cool_down"`           // 心想事成:干扰cd
	LHDXXSCWinLength       int     `json:"-" gorm:"column:lhd_xxsc_win_length"`                  // 心想事成:连赢局数
	LHDXXSCEvo             bool    `json:"-" gorm:"column:lhd_xxsc_evo"`                         // 心想事成:当次生效过策略没有
	LHDXXSCTriggerTimes    int     `json:"-" gorm:"column:lhd_xxsc_trigger_times"`               // 心想事成:有效触发次数u
	LHDXXSCMaxTriggerTimes int     `json:"-" gorm:"column:lhd_xxsc_max_trigger_times"`           // 心想事成:触发次数u总
	LHDQSBNTriggerTimes    int     `json:"-" gorm:"column:lhd_qsbn_trigger_times"`               // 求死不能:有效触发次数T
	LHDLKYHTriggerTimes    int     `json:"-" gorm:"column:lhd_lkyh_trigger_times"`               // 龙狂有祸:每日有效触发次数T
	LHDLKYHHistoryT        []int32 `json:"-" gorm:"column:lhd_lkyh_history_t;type:Array(Int32)"` // 龙狂有祸:历史每日有效触发次数T
	LHDLKYHTZ              int     `json:"-" gorm:"column:lhd_lkyh_tz"`                          // 龙狂有祸:总有效触发次数TZ
	LHDLKYHRZ              int     `json:"-" gorm:"column:lhd_lkyh_rz"`                          // 龙狂有祸:满足策略条件次数
	LHDLKYHBSCoolDown      int64   `json:"-" gorm:"column:lhd_lkyh_bs_cool_down"`                // 龙狂有祸:倍杀cd
	LHDLKYHBSRounds        int     `json:"-" gorm:"column:lhd_lkyh_bs_rounds"`                   // 龙狂有祸:倍杀局数
	LHDLKYHSuppress        int     `json:"-" gorm:"column:lhd_lkyh_suppress"`                    // 龙狂有祸:压制次数
	LHDLKYHNZ              int     `json:"-" gorm:"column:lhd_lkyh_nz"`                          // 龙狂有祸:倍杀触发局数
	LHDLKYHBSTimes         int     `json:"-" gorm:"column:lhd_lkyh_bstimes"`                     // 龙狂有祸:倍杀触发次数
	LHDLKYHDoubleBetTimes  int     `json:"-" gorm:"column:lhd_lkyh_doubleBetTimes"`              // 龙狂有祸:倍投次数
	LHDGcyxHighing         bool    `json:"-" gorm:"column:lhd_gcyx_highing"`                     // 高潮涌现:是否处于高潮状态
	LHDGcyxTriggerTimes    int     `json:"-" gorm:"column:lhd_gcyx_trigger_times"`               // 高潮涌现:高潮涌现:每日有效触发次数T
	LHDGcyxHp              int     `json:"-" gorm:"column:lhd_gcyx_hp"`                          // 高潮涌现:当前高潮率
	LHDGcyxC               int     `json:"-" gorm:"column:lhd_gcyx_c"`                           // 高潮涌现:贤者局冷却
	LHDGcyxM               int     `json:"-" gorm:"column:lhd_gcyx_m"`                           // 高潮涌现:高潮次数 人为制造高潮的次数
	LHDGcyxR               int     `json:"-" gorm:"column:lhd_gcyx_r"`                           // 高潮涌现:高潮局数 一次高潮可能对应多局
	LHDGcyxWin             int64   `json:"-" gorm:"column:lhd_gcyx_win"`                         // 高潮涌现:高潮局中赢钱金额
	LHDGcyxLose            int64   `json:"-" gorm:"column:lhd_gcyx_lose"`                        // 高潮涌现:高潮局中输钱金额
	LHDGcyxBetAvg          float64 `json:"-" gorm:"column:lhd_gcyx_bet_avg"`                     // 进入高潮时前f局均打码量
	// SevenStrategy  LHDUserStrategy   `json:"sevenStrategy" gorm:"column:seven_strategy"`    //7up策略
	SevenRoundBet            []int64 `json:"-" gorm:"column:seven_round_bet;type:Array(Int64)"`      // 每轮下注
	SevenXXSCDisturbCoolDown int     `json:"-" gorm:"column:seven_xxsc_disturb_cool_down"`           // 心想事成:干扰cd
	SevenXXSCWinLength       int     `json:"-" gorm:"column:seven_xxsc_win_length"`                  // 心想事成:连赢局数
	SevenXXSCEvo             bool    `json:"-" gorm:"column:seven_xxsc_evo"`                         // 心想事成:当次生效过策略没有
	SevenXXSCTriggerTimes    int     `json:"-" gorm:"column:seven_xxsc_trigger_times"`               // 心想事成:有效触发次数u
	SevenXXSCMaxTriggerTimes int     `json:"-" gorm:"column:seven_xxsc_max_trigger_times"`           // 心想事成:触发次数u总
	SevenQSBNTriggerTimes    int     `json:"-" gorm:"column:seven_qsbn_trigger_times"`               // 求死不能:有效触发次数T
	SevenLKYHTriggerTimes    int     `json:"-" gorm:"column:seven_lkyh_trigger_times"`               // 龙狂有祸:每日有效触发次数T
	SevenLKYHHistoryT        []int32 `json:"-" gorm:"column:seven_lkyh_history_t;type:Array(Int32)"` // 龙狂有祸:历史每日有效触发次数T
	SevenLKYHTZ              int     `json:"-" gorm:"column:seven_lkyh_tz"`                          // 龙狂有祸:总有效触发次数TZ
	SevenLKYHRZ              int     `json:"-" gorm:"column:seven_lkyh_rz"`                          // 龙狂有祸:满足策略条件次数
	SevenLKYHBSCoolDown      int64   `json:"-" gorm:"column:seven_lkyh_bs_cool_down"`                // 龙狂有祸:倍杀cd
	SevenLKYHBSRounds        int     `json:"-" gorm:"column:seven_lkyh_bs_rounds"`                   // 龙狂有祸:倍杀局数
	SevenLKYHSuppress        int     `json:"-" gorm:"column:seven_lkyh_suppress"`                    // 龙狂有祸:压制次数
	SevenLKYHNZ              int     `json:"-" gorm:"column:seven_lkyh_nz"`                          // 龙狂有祸:倍杀触发局数
	SevenLKYHBSTimes         int     `json:"-" gorm:"column:seven_lkyh_bstimes"`                     // 龙狂有祸:倍杀触发次数
	SevenLKYHDoubleBetTimes  int     `json:"-" gorm:"column:seven_lkyh_doubleBetTimes"`              // 龙狂有祸:倍投次数
	SevenGcyxHighing         bool    `json:"-" gorm:"column:seven_gcyx_highing"`                     // 高潮涌现:是否处于高潮状态
	SevenGcyxTriggerTimes    int     `json:"-" gorm:"column:seven_gcyx_trigger_times"`               // 高潮涌现:高潮涌现:每日有效触发次数T
	SevenGcyxHp              int     `json:"-" gorm:"column:seven_gcyx_hp"`                          // 高潮涌现:当前高潮率
	SevenGcyxC               int     `json:"-" gorm:"column:seven_gcyx_c"`                           // 高潮涌现:贤者局冷却
	SevenGcyxM               int     `json:"-" gorm:"column:seven_gcyx_m"`                           // 高潮涌现:高潮次数 人为制造高潮的次数
	SevenGcyxR               int     `json:"-" gorm:"column:seven_gcyx_r"`                           // 高潮涌现:高潮局数 一次高潮可能对应多局
	SevenGcyxWin             int64   `json:"-" gorm:"column:seven_gcyx_win"`                         // 高潮涌现:高潮局中赢钱金额
	SevenGcyxLose            int64   `json:"-" gorm:"column:seven_gcyx_lose"`                        // 高潮涌现:高潮局中输钱金额
	SevenGcyxBetAvg          float64 `json:"-" gorm:"column:seven_gcyx_bet_avg"`                     // 进入高潮时前f局均打码量
	// ABStrategy     ABUserStrategy    `json:"ABStrategy" gorm:"column:ab_strategy"`          //AB策略
	ABRoundBet         []int64 `json:"-" gorm:"column:ab_round_bet;type:Array(Int64)"` // 每轮下注
	ABANQSTriggerTimes int     `json:"-" gorm:"column:ab_anqs_trigger_times"`          // 安能求死:有效触发次数T
	ABARTYU            int     `json:"-" gorm:"column:ab_arty_u"`                      // 安然躺赢:有效触发次数T
	ABARTYUZ           int     `json:"-" gorm:"column:ab_arty_uz"`                     // 安然躺赢:触发次数T
	ABARTYEvoTimes     bool    `json:"-" gorm:"column:ab_arty_evoTimes"`               // 安然躺赢:当次触发过策略没有
	ABARTYWinScore     int64   `json:"-" gorm:"column:ab_arty_win_score"`              // 安然躺赢:净赢分
	// CPStrategy     CPUserStrategy    `json:"CPStrategy" gorm:"column:cp_strategy"`          //彩票策略
	CPRoundBet         []int64 `json:"roundBet" gorm:"column:cp_round_bet;type:Array(Int64)"` // 每轮下注
	CPLYQNTriggerTimes int     `json:"-" gorm:"column:trigger_times"`                         // 来易去难:有效触发次数T
	CPLWJYU            int     `json:"-" gorm:"column:cp_lwjy_u"`                             // 来玩就赢:有效触发次数T
	CPLWJYUZ           int     `json:"-" gorm:"column:cp_lwjy_uz"`                            // 来玩就赢:触发次数T
	CPLWJYEvoTimes     bool    `json:"-" gorm:"column:cp_lwjy_evoTimes"`                      // 来玩就赢:当次触发过策略没有
	CPLWJYWinScore     int64   `json:"-" gorm:"column:cp_lwjy_win_score"`                     // 来玩就赢:净赢分
	// RBUserStrategy     RBUserStrategy `json:"CPStrategy" gorm:"column:cp_strategy"`         //红黑策略
	RBRoundBet         []int64 `json:"-" gorm:"column:round_bet;type:Array(Int64)"` // 每轮下注
	RBJCFSTriggerTimes int     `json:"-" gorm:"column:rb_jcfs_trigger_times"`       // 绝处逢生:有效触发次数T
	RBHYDTU            int     `json:"-" gorm:"column:rb_hydt_u"`                   // 红运当头:有效触发次数T
	RBHYDTUZ           int     `json:"-" gorm:"column:rb_hydt_uz"`                  // 红运当头:触发次数T
	RBHYDTEvoTimes     bool    `json:"-" gorm:"column:rb_hydt_evoTimes"`            // 红运当头:当次触发过策略没有
	RBHYDTWinScore     int64   `json:"-" gorm:"column:rb_hydt_win_score"`           // 红运当头:净赢分
	// TP类策略相关
	TPTotalRound  int32 `json:"tp_total_round" gorm:"column:tp_total_round"`   //累计局数
	TPTiggerTimes int32 `json:"tp_tigger_times" gorm:"column:tp_tigger_times"` //触发次数
	//埋点
	NewbieGuid       bool `json:"newbie_guid" gorm:"column:newbie_guid"`             //新手引导标记
	NewbieCondition1 bool `json:"newbie_condition1" gorm:"column:newbie_condition1"` //新手状态1
	NewbieCondition2 bool `json:"newbie_condition2" gorm:"column:newbie_condition2"` //新手状态2

	Kick105Flag      bool  `json:"kick105_flag" gorm:"column:kick105_flag"`             //105踢出标记
	Kick200Flag      bool  `json:"kick200_flag" gorm:"column:kick200_flag"`             //200踢出标记
	KickWithdrawFlag int   `json:"kick_withdraw_flag" gorm:"column:kick_withdraw_flag"` //新手提现踢出标记
	WithdrawPOPTime  int64 `json:"withdraw_pop_time" gorm:"column:withdraw_pop_time"`   //上次提现弹窗时间

	CustomTypes string `json:"customtypes" gorm:"column:-"` // 自定义标记

	Strategy100Flag bool `json:"trigger_strategy100" gorm:"column:trigger_strategy100"` //策略100标记
	Strategy200Flag bool `json:"trigger_strategy200" gorm:"column:trigger_strategy200"` //策略200标记
	// B类玩家
	BetRecord       []int64 `json:"betRecord" gorm:"column:bet_record;type:Array(Int64)"` // 投注记录(200局)
	FluctuateLine   float64 `json:"fluctuateLine" gorm:"column:fluctuate_line"`           // 起伏线（非实时,后台展示用的）
	Withdrawable200 bool    `json:"withdraw200" gorm:"column:withdraw200"`                // 200可提现标记

	// SurplusGame            int     `json:"surplusGame" gorm:"column:surplus_game"`                        // 转平民余局数
	// SurplusGameTime        int64   `json:"surplusGameTime" gorm:"column:surplus_game_time"`               // 转平民剩余时间
	Novice200StrategyCount int             `json:"novice200StrategyCount" gorm:"column:novice200_strategy_count"`                       // B类新手200策略触发次数
	FreeWelfare            int             `json:"freeWelfare" gorm:"column:free_welfare"`                                              // 百人场免费福利
	BGiveCash              int64           `json:"bGiveCash" gorm:"column:b_give_cash"`                                                 // B类新手充值获得的彩金
	ViceAccount            bool            `json:"viceAccount" gorm:"column:vice_account"`                                              // 小号标记 该账号是否有重复的adid或银行卡号
	TpNewbieProbeId        int32           `json:"tp_newbie_probe_id" gorm:"column:tp_newbie_probe_id"`                                 // tp新手试探期id
	TpNewbieStableId       int32           `json:"tp_newbie_stable_id" gorm:"column:tp_newbie_stable_id"`                               // tp新手稳定局id
	TpNewbieStableRounds   map[int32]int32 `json:"tp_newbie_stable_rounds" gorm:"column:tp_newbie_stable_rounds;type:Map(Int32,Int32)"` // tp新手稳定局策略局数

	//tp相关状态
	TpUserModel                 int32   `json:"tpUserModel" gorm:"column:tp_user_model"`                                   //玩家所处模式 0 新手 1免费 2正常
	TpUserStage                 int32   `json:"tpUserStage" gorm:"column:tp_user_stage"`                                   //玩家所处阶段
	TpUserStageHistory          []int32 `json:"tpUserStageHistory" gorm:"column:tp_user_stage_history;type:Array(Int32)"`  //玩家所处阶段历史
	TpUserRound                 int32   `json:"tpUserRound" gorm:"column:tp_user_round"`                                   //局数
	TpUserWinOrLoseAmount       int64   `json:"tpUserWinOrLoseAmount" gorm:"column:tp_user_win_or_lose_amount"`            //输赢额
	TpUserChangeCorrectionValue int64   `json:"tpUserChangeCorrectionValue" gorm:"column:tp_user_change_correction_value"` //变化修正值 正数为多赢，负数为多输

	TpUserTodayChargeNum           int32             `json:"tpUserTodayChargeNum" gorm:"column:tp_user_today_charge_num"`                                          //今日充值次数
	TpUserTodayChargeAmount        int64             `json:"tpUserTodayChargeAmount" gorm:"column:tp_user_today_charge_amount"`                                    //今日充值金额
	TpUserTodayWithdrawNum         int32             `json:"tpUserTodayWithdrawNum" gorm:"column:tp_user_today_withdraw_num"`                                      //今日提现次数
	TpUserTodayWithdrawAmount      int64             `json:"tpUserTodayWithdrawAmount" gorm:"column:tp_user_today_withdraw_amount"`                                //今日提现金额
	TpUserTotalWinOrLoseAmount     int64             `json:"tpUserTotalWinOrLoseAmount" gorm:"column:tp_user_total_win_or_lose_amount"`                            //总输赢额
	TpUserChargeInGameNumOfTrigger int32             `json:"tpUserChargeInGameNumOfTrigger" gorm:"column:tp_user_charge_in_game_num_of_trigger"`                   //局内充值触发数
	TpUserChargeInGameNumOfSuccess int32             `json:"tpUserChargeInGameNumOfSuccess" gorm:"column:tp_user_charge_in_game_num_of_success"`                   //局内充值成功数
	TpUserFollowRateTrigger        map[int32]int32   `json:"tpUserFollowRateTrigger" gorm:"column:tp_user_follow_rate_trigger;type:Map(Int32,Int32)"`              //跟牌率触发数
	TpUserFollowRateSuccess        map[int32]int32   `json:"tpUserFollowRateSuccess" gorm:"column:tp_user_follow_rate_success;type:Map(Int32,Int32)"`              //跟牌率成功数
	TpUserStoryCD                  map[int32]int32   `json:"tpUserStoryCd" gorm:"column:tp_user_story_cd;type:Map(Int32,Int32)"`                                   //剧情模式cd
	TpUserControlStrategyHistory   []int32           `json:"tpUserControlStrategyHistory" gorm:"column:tp_user_control_strategy_history;type:Array(Int32)"`        //控制策略历史
	TpUserGameTime                 int64             `json:"tpUserGameTime" gorm:"column:tp_user_game_time"`                                                       //上次充值后游戏时长
	TpUserTodayGameRound           int32             `json:"tpUserTodayGameRound" gorm:"column:tp_user_today_game_round"`                                          //今日游戏局数
	TpUserTodayControlStrategyNum  map[int32]int32   `json:"tpUserTodayControlStrategyNum" gorm:"column:tp_user_today_control_strategy_num;type:Map(Int32,Int32)"` //今日控制策略触发次数
	TpUserChargeInGameStory        int32             `json:"tpUserChargeInGameStory" gorm:"column:tp_user_charge_in_game_story"`                                   //剧情局局内充值成功数
	TpUserStoryPlus                bool              `json:"tpUserStoryPlus" gorm:"column:tp_user_story_plus"`                                                     //tp用户剧情局plus
	TpHurtWinRound                 int32             `json:"tpHurtWinRound" gorm:"column:tp_hurt_win_round"`                                                       // tp 偷鸡赢钱局数
	TpBeHurtLoseRound              int32             `json:"tpBeHurtLoseRound" gorm:"column:tp_be_hurt_lose_round"`                                                // tp 被偷鸡输钱局数
	TpHurtWinScore                 int64             `json:"tpHurtWinScore" gorm:"column:tp_hurt_win_score"`                                                       // tp 偷鸡赢钱金额
	TpBeHurtLoseScore              int64             `json:"tpBeHurtLoseScore" gorm:"column:tp_be_hurt_lose_score"`                                                // tp 被偷鸡输钱金额
	TpUserVHRecord                 map[int32][]int64 `json:"-" gorm:"column:tp_user_vh_record;type:Map(Int32,Array(Int64))"`                                       // tp玩家vh记录 <int32(牌型6bit, ante(25bit), [牌型,底注,牌型次数,下注总金额,赢次数,输次数]>
	TpHighCardRounds               int32             `json:"tpHighCardRounds" gorm:"column:tp_high_card_rounds"`                                                   // tp 连续拿高牌次数
	TpHeartbeatRound               int32             `json:"tpHeartbeatRound" gorm:"column:tp_heartbeat_round"`                                                    // tp怦然心动记录开始总玩局数
	TpHeartbeatResetTimes          int32             `json:"tpHeartbeatResetTimes" gorm:"column:tp_heartbeat_reset_times"`                                         // tp心跳重置次数
	TpHeartbeatHR                  float64           `json:"tpHeartbeatHR" gorm:"column:tp_heartbeat_hr"`                                                          // tp怦然心动心率
	TpUserLjsbPyCD                 int32             `json:"tpUserLjsbPyCD" gorm:"column:tp_user_ljsb_py_cd"`                                                      // tp乐极生悲策略被冤局cd
	TpUserGcyxCD                   int32             `json:"tpUserGcyxCD" gorm:"column:tp_user_gcyx_cd"`                                                           // tp高潮涌现策略cd
	TpUserGcyxHp                   int32             `json:"tpUserGcyxHp" gorm:"column:tp_user_gcyx_hp"`                                                           // tp高潮涌现策略当前概率
	TpUserGcyxWin                  int64             `json:"tpUserGcyxWin" gorm:"column:tp_user_gcyx_win"`                                                         // tp高潮涌现策略玩家赢分
	TpUserGcyxLose                 int64             `json:"tpUserGcyxLose" gorm:"column:tp_user_gcyx_lose"`                                                       // tp高潮涌现策略玩家输分

	// RM相关
	RmRoiLimits         map[string]int32 `json:"rm_roi_limits" gorm:"column:rm_roi_limits;type:Map(String,Int32)"`         // rm roi控制策略生效次数
	RmRoiDayLimits      map[string]int32 `json:"rm_roi_day_limits" gorm:"column:rm_roi_day_limits;type:Map(String,Int32)"` // rm 每个自然日，对应ID策略生效次数
	RmWithout1stDrop    int              `json:"rm_without_1st_drop" gorm:"column:rm_without_1st_drop"`                    // rm 玩家在没有1Life时候首回合弃牌次数
	RmWithout1stNotDrop int              `json:"rm_without_1st_not_drop" gorm:"column:rm_without_1st_not_drop"`            // rm 玩家在没有1Life时候首回合未弃牌次数

	ActivityTurnPrizeTimes int32 `json:"activityTurnPrizeTimes" gorm:"column:activity_turn_prize_times"` // 转盘活动申请领奖次数
	// 代理
	ShareAgentTeamLv                int32    `json:"teamLv" gorm:"column:team_lv"`                                                // 团队等级
	ShareAgentTeamAgents            int32    `json:"teamAgents" gorm:"column:team_agents"`                                        // 当前团队总人数
	ShareAgentTeamBets              int64    `json:"teamBets" gorm:"column:team_bets"`                                            // 当前团队打码量
	ShareAgentShareUsers            int32    `json:"shareUsers" gorm:"column:share_users"`                                        // 当前累计有效人头数
	ShareAgentShareUsersPrizes      []int32  `json:"shareUsersPrizes" gorm:"column:share_users_prizes;type:Array(Int32)"`         // 当前已达成累计有效人头数奖励
	ShareAgentShareSuper            bool     `json:"shareSuper" gorm:"column:share_super"`                                        // 是否已计算为有效人头
	ShareAgentEarningsBetTotal      int64    `json:"earningsBetTotal" gorm:"column:earnings_bet_total"`                           // 已领取的打码返佣金额
	ShareAgentEarningsBetUnclaimed  int64    `json:"earningsBetUnclaimed" gorm:"column:earnings_bet_unclaimed"`                   // 还没领取的打码返佣金额
	ShareAgentEarningsBetPendding   int64    `json:"earningsBetPendding" gorm:"column:earnings_bet_pendding"`                     // 处理中的打码返佣金额(0点后加到unclaimed)(单位:厘)
	ShareAgentEarningsUserTotal     int64    `json:"earningsUserTotal" gorm:"column:earnings_user_total"`                         // 已领取的人头奖励和累计任务返佣金额
	ShareAgentEarningsUserUnclaimed int64    `json:"earningsUserUnclaimed" gorm:"column:earnings_user_unclaimed"`                 // 还没领取的人头奖励和累计任务返佣金额
	ShareAgentEarningsUserPendding  int64    `json:"earningsUserPendding" gorm:"column:earnings_user_pendding"`                   // 处理中的人头奖励和累计任务返佣金额(0点后加到unclaimed)(单位:分)
	ShareAgentTodayEarningsUser     int32    `json:"todayEarningsUser" gorm:"column:today_earnings_user"`                         // 今日领取单个人头奖数 -> 单个人头奖的单日上限（人）
	ShareAgentTodayTakeItypeDates   []string `json:"todayTakeItypeDates" gorm:"column:today_take_itype_dates;type:Array(String)"` // 今日手动领取奖励类型列表
	VolatilitySubsidyTimes          int32    `json:"volatilitySubsidyTimes" gorm:"column:volatility_subsidy_times"`               // 波动返水已补贴次数
	VolatilitySubsidyAmounts        int64    `json:"volatilitySubsidyAmounts" gorm:"column:volatility_subsidy_amounts"`           // 波动返水已补贴额
	VolatilitySubsidyFirstTime      int64    `json:"volatilitySubsidyFirstTime" gorm:"column:volatility_subsidy_first_time"`      // 波动返水首次补贴时间(毫秒)
}

func (*User) TableName() string {
	return "col_user"
}

func (*User) New() CkEntity {
	return new(User)
}

func (c *User) ParseData(d interface{}) (rs []interface{}, err error) {
	// check type
	from, ok := d.(*data.User)
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
	//user vip
	c.VipLv = from.Vip.Lv
	c.VipMaxLv = from.Vip.MaxLv
	c.VipDaily = from.Vip.Daily
	c.VipWeekly = from.Vip.Weekly
	c.VipWeeklyTime = from.Vip.WeeklyTime
	c.VipWeekday = from.Vip.Weekday
	c.VipLevelReward = from.Vip.LevelReward
	c.VipExp = from.Vip.Exp
	c.VipWithdrawCount = from.Vip.WithdrawCount
	c.VipVersion = from.Vip.Version
	c.parseGameData(from, c)
	rs = append(rs, c)

	// 用户账户金额信息
	uf := new(UserFinance)
	if err = sonic.Unmarshal(bytes, uf); err != nil {
		return
	}
	uf.Ver = time.Now().UnixMilli()
	// 利润
	uf.Profit = (int64(uf.CashOut) + uf.Diamond + uf.ShadowDiamond + uf.Coin) - int64(uf.Money)
	rs = append(rs, uf)
	return
}

func (c *User) parseGameData(src *data.User, dst *User) {
	dst.CrashMultiple = src.CrashMultiple
	dst.PlaneAutoLeave = src.PlaneAutoLeave
	dst.PlaneMultiple = src.PlaneMultiple
	dst.RoundGames = src.RoundGames
	// AvStrategy     AviatorUserStrategy `json:"avStrategy" bson:"av_strategy"`   //Aviator类策略
	dst.AvFYZSTiggerTimes = src.AvStrategy.FYZS.TiggerTimes
	dst.AvFYZSEvoTimes = src.AvStrategy.FYZS.EvoTimes
	dst.AvFYZSAllTiggerTimes = src.AvStrategy.FYZS.AllTiggerTimes
	dst.AvFYZSAllEvoTimes = src.AvStrategy.FYZS.AllEvoTimes
	dst.AvFYZSWinScore = src.AvStrategy.FYZS.WinScore
	dst.AvYHWMMonitorRounds = src.AvStrategy.YHWM.MonitorRounds
	dst.AvYHWMAviatorMulpitles = src.AvStrategy.YHWM.AviatorMulpitles
	dst.AvYHWMTiggerTimes = src.AvStrategy.YHWM.TiggerTimes
	dst.AvYHWMWinScore = src.AvStrategy.YHWM.WinScore
	dst.AvYHWMRecyleScore = src.AvStrategy.YHWM.RecyleScore
	dst.AvYHWMAllRecyleScore = src.AvStrategy.YHWM.AllRecyleScore
	dst.AvYHWMTigger = src.AvStrategy.YHWM.Tigger
	dst.AvQSHSWinRounds = src.AvStrategy.QSHS.WinRounds
	dst.AvQSHSTriggerTimes = src.AvStrategy.QSHS.TriggerTimes
	dst.AvJCFKTriggerTimes = src.AvStrategy.JCFK.TriggerTimes
	dst.AvJCFKJackpotVal = src.AvStrategy.JCFK.JackpotVal
	dst.AvMXJLTriggerTimes = src.AvStrategy.MXJL.TriggerTimes
	dst.AvMXJLBets = src.AvStrategy.MXJL.Bets
	dst.AvRKYHTriggerTimes = src.AvStrategy.RKYH.TriggerTimes
	dst.AvRKYHWinMulpitles = src.AvStrategy.RKYH.WinMulpitles
	dst.AvGcyxEvoTimes = src.AvStrategy.GCYX.EvoTimes
	dst.AvGcyxIsStateGC = src.AvStrategy.GCYX.IsStateGC
	dst.AvGcyxIsStateXZ = src.AvStrategy.GCYX.IsStateXZ
	dst.AvGcyxXZTimes = src.AvStrategy.GCYX.XZTimes
	dst.AvGcyxWinScore = src.AvStrategy.GCYX.WinScore
	dst.AvGcyxM = src.AvStrategy.GCYX.M
	dst.AvGcyxR = src.AvStrategy.GCYX.R
	dst.AvGcyxDMRecord = src.AvStrategy.GCYX.DMRecord
	dst.AvGcyxDailyGCTimes = src.AvStrategy.GCYX.DailyGCTimes
	dst.AvGcyxAvgBet = src.AvStrategy.GCYX.AvgBet
	// CrashStrategy  CrashUserStrategy `json:"crashStrategy" gorm:"column:crash_strategy"`    //crash类策略
	dst.CrashFYZSPlayTimes = src.CrashStrategy.FYZS.PlayTimes
	dst.CrashFYZSTiggerTimes = src.CrashStrategy.FYZS.TiggerTimes
	dst.CrashFYZSEvoTimes = src.CrashStrategy.FYZS.EvoTimes
	dst.CrashFYZSAllTiggerTimes = src.CrashStrategy.FYZS.AllTiggerTimes
	dst.CrashFYZSAllEvoTimes = src.CrashStrategy.FYZS.AllEvoTimes
	dst.CrashYHWMMonitorRounds = src.CrashStrategy.YHWM.MonitorRounds
	dst.CrashYHWMCrashMulpitles = src.CrashStrategy.YHWM.CrashMulpitles
	dst.CrashYHWMTiggerTimes = src.CrashStrategy.YHWM.TiggerTimes
	dst.CrashYHWMWinScore = src.CrashStrategy.YHWM.WinScore
	dst.CrashYHWMRecyleScore = src.CrashStrategy.YHWM.RecyleScore
	dst.CrashYHWMAllRecyleScore = src.CrashStrategy.YHWM.AllRecyleScore
	dst.CrashYHWMTigger = src.CrashStrategy.YHWM.Tigger
	dst.CrashQSHSWinRounds = src.CrashStrategy.QSHS.WinRounds
	dst.CrashQSHSTriggerTimes = src.CrashStrategy.QSHS.TriggerTimes
	dst.CrashJCFKTriggerTimes = src.CrashStrategy.JCFK.TriggerTimes
	dst.CrashJCFKJackpotVal = src.CrashStrategy.JCFK.JackpotVal
	dst.CrashMXJLTriggerTimes = src.CrashStrategy.MXJL.TriggerTimes
	dst.CrashMXJLBets = src.CrashStrategy.MXJL.Bets
	dst.CrashRKYHTriggerTimes = src.CrashStrategy.RKYH.TriggerTimes
	dst.CrashRKYHWinMulpitles = src.CrashStrategy.RKYH.WinMulpitles
	dst.CrashEvoTimes = src.CrashStrategy.GCYX.EvoTimes
	dst.CrashIsStateGC = src.CrashStrategy.GCYX.IsStateGC
	dst.CrashIsStateXZ = src.CrashStrategy.GCYX.IsStateXZ
	dst.CrashXZTimes = src.CrashStrategy.GCYX.XZTimes
	dst.CrashWinScore = src.CrashStrategy.GCYX.WinScore
	dst.CrashM = src.CrashStrategy.GCYX.M
	dst.CrashR = src.CrashStrategy.GCYX.R
	dst.CrashDMRecord = src.CrashStrategy.GCYX.DMRecord
	dst.CrashDailyGCTimes = src.CrashStrategy.GCYX.DailyGCTimes
	dst.CrashGcyxAvgBet = src.CrashStrategy.GCYX.AvgBet
	// LHDStrategy    LHDUserStrategy   `json:"LHDStrategy" gorm:"column:lhd_strategy"`        //LHD策略
	dst.LHDRoundBet = src.LHDStrategy.RoundBet
	dst.LHDXXSCDisturbCoolDown = src.LHDStrategy.XXSC.DisturbCoolDown
	dst.LHDXXSCWinLength = src.LHDStrategy.XXSC.WinLength
	dst.LHDXXSCEvo = src.LHDStrategy.XXSC.Evo
	dst.LHDXXSCTriggerTimes = src.LHDStrategy.XXSC.TriggerTimes
	dst.LHDXXSCMaxTriggerTimes = src.LHDStrategy.XXSC.MaxTriggerTimes
	dst.LHDQSBNTriggerTimes = src.LHDStrategy.QSBN.TriggerTimes
	dst.LHDLKYHTriggerTimes = src.LHDStrategy.LKYH.TriggerTimes
	for _, item := range src.LHDStrategy.LKYH.HistoryT {
		dst.LHDLKYHHistoryT = append(dst.LHDLKYHHistoryT, int32(item))
	}
	dst.LHDLKYHTZ = src.LHDStrategy.LKYH.TZ
	dst.LHDLKYHRZ = src.LHDStrategy.LKYH.RZ
	dst.LHDLKYHBSCoolDown = src.LHDStrategy.LKYH.BSCoolDown
	dst.LHDLKYHBSRounds = src.LHDStrategy.LKYH.BSRounds
	dst.LHDLKYHSuppress = src.LHDStrategy.LKYH.Suppress
	dst.LHDLKYHNZ = src.LHDStrategy.LKYH.NZ
	dst.LHDLKYHBSTimes = src.LHDStrategy.LKYH.BSTimes
	dst.LHDLKYHDoubleBetTimes = src.LHDStrategy.LKYH.DoubleBetTimes
	dst.LHDGcyxHighing = src.LHDStrategy.GCYX.Highing
	dst.LHDGcyxTriggerTimes = src.LHDStrategy.GCYX.TriggerTimes
	dst.LHDGcyxHp = src.LHDStrategy.GCYX.Hp
	dst.LHDGcyxC = src.LHDStrategy.GCYX.C
	dst.LHDGcyxM = src.LHDStrategy.GCYX.M
	dst.LHDGcyxR = src.LHDStrategy.GCYX.R
	dst.LHDGcyxWin = src.LHDStrategy.GCYX.Win
	dst.LHDGcyxLose = src.LHDStrategy.GCYX.Lose
	dst.LHDGcyxBetAvg = src.LHDStrategy.GCYX.BetAvg
	// SevenStrategy  LHDUserStrategy   `json:"sevenStrategy" gorm:"column:seven_strategy"`    //7up策略
	dst.SevenRoundBet = src.SevenStrategy.RoundBet
	dst.SevenXXSCDisturbCoolDown = src.SevenStrategy.XXSC.DisturbCoolDown
	dst.SevenXXSCWinLength = src.SevenStrategy.XXSC.WinLength
	dst.SevenXXSCEvo = src.SevenStrategy.XXSC.Evo
	dst.SevenXXSCTriggerTimes = src.SevenStrategy.XXSC.TriggerTimes
	dst.SevenXXSCMaxTriggerTimes = src.SevenStrategy.XXSC.MaxTriggerTimes
	dst.SevenQSBNTriggerTimes = src.SevenStrategy.QSBN.TriggerTimes
	dst.SevenLKYHTriggerTimes = src.SevenStrategy.LKYH.TriggerTimes
	for _, item := range src.SevenStrategy.LKYH.HistoryT {
		dst.SevenLKYHHistoryT = append(dst.SevenLKYHHistoryT, int32(item))
	}
	dst.SevenLKYHTZ = src.SevenStrategy.LKYH.TZ
	dst.SevenLKYHRZ = src.SevenStrategy.LKYH.RZ
	dst.SevenLKYHBSCoolDown = src.SevenStrategy.LKYH.BSCoolDown
	dst.SevenLKYHBSRounds = src.SevenStrategy.LKYH.BSRounds
	dst.SevenLKYHSuppress = src.SevenStrategy.LKYH.Suppress
	dst.SevenLKYHNZ = src.SevenStrategy.LKYH.NZ
	dst.SevenLKYHBSTimes = src.SevenStrategy.LKYH.BSTimes
	dst.SevenLKYHDoubleBetTimes = src.SevenStrategy.LKYH.DoubleBetTimes
	dst.SevenGcyxHighing = src.SevenStrategy.GCYX.Highing
	dst.SevenGcyxTriggerTimes = src.SevenStrategy.GCYX.TriggerTimes
	dst.SevenGcyxHp = src.SevenStrategy.GCYX.Hp
	dst.SevenGcyxC = src.SevenStrategy.GCYX.C
	dst.SevenGcyxM = src.SevenStrategy.GCYX.M
	dst.SevenGcyxR = src.SevenStrategy.GCYX.R
	dst.SevenGcyxWin = src.SevenStrategy.GCYX.Win
	dst.SevenGcyxLose = src.SevenStrategy.GCYX.Lose
	dst.SevenGcyxBetAvg = src.SevenStrategy.GCYX.BetAvg
	// ABStrategy     ABUserStrategy    `json:"ABStrategy" gorm:"column:ab_strategy"`          //AB策略
	dst.ABRoundBet = src.ABStrategy.RoundBet
	dst.ABANQSTriggerTimes = src.ABStrategy.ANQS.TriggerTimes
	dst.ABARTYU = src.ABStrategy.ARTY.U
	dst.ABARTYUZ = src.ABStrategy.ARTY.UZ
	dst.ABARTYEvoTimes = src.ABStrategy.ARTY.EvoTimes
	dst.ABARTYWinScore = src.ABStrategy.ARTY.WinScore
	// CPStrategy     CPUserStrategy    `json:"CPStrategy" gorm:"column:cp_strategy"`          //彩票策略
	dst.CPRoundBet = src.CPStrategy.RoundBet
	dst.CPLYQNTriggerTimes = src.CPStrategy.LYQN.TriggerTimes
	dst.CPLWJYU = src.CPStrategy.LWJY.U
	dst.CPLWJYUZ = src.CPStrategy.LWJY.UZ
	dst.CPLWJYEvoTimes = src.CPStrategy.LWJY.EvoTimes
	dst.CPLWJYWinScore = src.CPStrategy.LWJY.WinScore
	// RBUserStrategy     RBUserStrategy `json:"CPStrategy" gorm:"column:cp_strategy"`         //红黑策略
	dst.RBRoundBet = src.RBStrategy.RoundBet
	dst.RBJCFSTriggerTimes = src.RBStrategy.JCFS.TriggerTimes
	dst.RBHYDTU = src.RBStrategy.HYDT.U
	dst.RBHYDTUZ = src.RBStrategy.HYDT.UZ
	dst.RBHYDTEvoTimes = src.RBStrategy.HYDT.EvoTimes
	dst.RBHYDTWinScore = src.RBStrategy.HYDT.WinScore

	// TP vh记录
	dst.TpUserVHRecord = make(map[int32][]int64, len(src.TpUserVHRecord))
	for key, record := range src.TpUserVHRecord {
		dst.TpUserVHRecord[key] = record[:]
	}
	// TP 怦然心动心率hr
	if src.TpHeartbeatResetTimes == 0 {
		dst.TpHeartbeatHR = 0
	} else {
		dst.TpHeartbeatHR = float64(src.TpHeartbeatRound) / float64(src.TpHeartbeatResetTimes)
	}

	// 代理
	dst.ShareAgentTeamLv = src.ShareAgent.TeamLv
	dst.ShareAgentTeamAgents = src.ShareAgent.TeamAgents
	dst.ShareAgentTeamBets = src.ShareAgent.TeamBets
	dst.ShareAgentShareUsers = src.ShareAgent.ShareUsers
	dst.ShareAgentShareUsersPrizes = src.ShareAgent.ShareUsersPrizes
	dst.ShareAgentShareSuper = src.ShareAgent.ShareSuper
	dst.ShareAgentEarningsBetTotal = src.ShareAgent.EarningsBetTotal
	dst.ShareAgentEarningsBetUnclaimed = src.ShareAgent.EarningsBetUnclaimed
	dst.ShareAgentEarningsBetPendding = src.ShareAgent.EarningsBetPendding
	dst.ShareAgentEarningsUserTotal = src.ShareAgent.EarningsUserTotal
	dst.ShareAgentEarningsUserUnclaimed = src.ShareAgent.EarningsUserUnclaimed
	dst.ShareAgentEarningsUserPendding = src.ShareAgent.EarningsUserPendding
	dst.ShareAgentTodayEarningsUser = src.ShareAgent.TodayEarningsUser
	dst.ShareAgentTodayTakeItypeDates = src.ShareAgent.TodayTakeItypeDates
}

// 用户账户金额相关
type UserFinance struct {
	Ver           int64     `gorm:"column:ver" json:"ver"`                       // 插入时间戳
	Userid        string    `gorm:"column:userid;primaryKey" json:"userid"`      // 用户id
	Diamond       int64     `gorm:"column:diamond" json:"diamond"`               // 钻石(彩金cash)
	GiveDiamond   int64     `json:"giveDiamond" gorm:"column:give_diamond"`      // 赠送彩金
	VBBank        int64     `gorm:"column:vbbank" json:"vbbank"`                 // VB银行
	OutDiamond    int64     `json:"outDiamond" gorm:"column:out_diamond"`        // 可提现彩金
	ShadowDiamond int64     `gorm:"column:shadow_diamond" json:"shadow_diamond"` // 隐藏钻石(用户看不见的)
	Coin          int64     `gorm:"column:coin" json:"coin"`                     // 金币(奖励金bonus)
	Money         uint32    `gorm:"column:money" json:"money"`                   // 充值总金额(分)
	CashOut       int32     `gorm:"column:cash_out" json:"cash_out"`             // 提现总金额(分)
	Ctime         time.Time `gorm:"column:ctime" json:"ctime"`                   // 注册时间
	Profit        int64     `gorm:"column:profit" json:"profit"`                 // 利润 (提现+携带)-充值
}

func (*UserFinance) TableName() string {
	return "col_user_finance"
}

func (*UserFinance) New() CkEntity {
	return new(UserFinance)
}

func (c *UserFinance) ParseData(d interface{}) (rs []interface{}, err error) {
	// check type
	from, ok := d.(*data.UserFinance)
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
	// 利润
	c.Profit = (int64(c.CashOut) + c.Diamond + c.ShadowDiamond + c.Coin) - int64(c.Money)
	rs = append(rs, c)
	return
}
