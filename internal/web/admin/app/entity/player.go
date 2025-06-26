package entity

import "time"

// 玩家数据
type PlayerUser struct {
	Userid        string `bson:"_id" json:"id"`                        // 用户id
	UserType      string `bson:"-"`                                    // 用户类型名称
	Nickname      string `bson:"nickname" json:"nickname"`             // 用户昵称
	RealName      string `bson:"real_name" json:"real_name"`           // 真实姓名
	Photo         string `bson:"photo" json:"photo"`                   // 头像
	PhotoUrl      string `bson:"-"`                                    // 头像地址
	Sex           uint32 `bson:"sex" json:"sex"`                       // 用户性别,男1 女2 非男非女3
	Phone         string `bson:"phone" json:"phone"`                   // 绑定的手机号码
	Phone2        string `bson:"phone2" json:"phone2"`                 // 提现绑定的手机号码
	Tourist       string `bson:"tourist" json:"tourist"`               // 游客
	RegistArea    int    `json:"registArea" bson:"regist_area"`        // ab测试 0:A 1:B
	Auth          string `bson:"auth" json:"auth"`                     // 密码验证码
	Password      string `bson:"password" json:"password"`             // MD5密码
	RegistIP      string `bson:"regist_ip" json:"regist_ip"`           // 注册账户时的IP地址
	LoginIP       string `bson:"login_ip" json:"login_ip"`             // 登录账户时的IP地址
	Diamond       int64  `bson:"diamond" json:"diamond"`               // 钻石(彩金cash)
	ShadowDiamond int64  `bson:"shadow_diamond" json:"shadow_diamond"` // 隐藏钻石(用户看不见的)
	Coin          int64  `bson:"coin" json:"coin"`                     // 金币(奖励金bonus)
	GiveDiamond   int64  `json:"giveDiamond" bson:"give_diamond"`      // 赠送彩金
	OutDiamond    int64  `json:"outDiamond" bson:"out_diamond"`        // 可提现彩金
	CashWater     uint64 `bson:"cash_water" json:"cash_water"`         // 彩金流水(游戏产生的)
	Status        int    `bson:"status" json:"status"`                 // 正常1  锁定2  黑名单3 白名单4
	Robot         bool   `bson:"robot" json:"robot"`                   // 是否是机器人
	State         int    `json:"state" bson:"state"`                   // 玩家状态 1.新手 2.正常 3.平民
	CustomTypes   string `json:"customtypes" bson:"custom_types"`      // 自定义用户类型
	LoginTimes    int32  `bson:"login_times" json:"login_times"`       // 登录天数
	OnlineStatus  bool   `bson:"online_status" json:"online_status"`   // 是否在线
	// 总资产、显示金额
	Assets         float64 `bson:"-"` // 总资产
	ShowAmount     float64 `bson:"-"` // 显示金额
	FDiamond       float64 `bson:"-"` // 钻石(彩金cash)
	FShadowDiamond float64 `bson:"-"` // 隐藏钻石(用户看不见的)
	FCoin          float64 `bson:"-"` // 金币(奖励金bonus)
	FMoney         float64 `bson:"-"` // 充值总金额(分)
	FCashOut       float64 `bson:"-"` // 提现总金额(分)
	FWin           float64 `bson:"-"` // 赢
	FGiveDiamond   float64 `bson:"-"` // 赠送彩金
	FOutDiamond    float64 `bson:"-"` // 可提现彩金
	// af参数
	AppId       string `bson:"app_id" json:"app_id"`
	AFId        string `bson:"af_id" json:"af_id"`
	OS          string `bson:"os" json:"os"`
	BundleId    string `bson:"bundle_id" json:"bundle_id"` // 包名
	RefGameId   string `bson:"ref_game_id" json:"ref_game_id"`
	RefPkgName  string `bson:"ref_pkg_name" json:"ref_pkg_name"`
	MediaSource string `bson:"media_source" json:"media_source"` // 渠道
	AFKey       string `bson:"af_key" json:"af_key"`
	DeviceId    string `bson:"device_id" json:"device_id"`
	// ad参数
	AD_BundleId            string `json:"ad__bundle_id" bson:"ad__bundle_id"` // 包名
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
	AD_ADID                string `json:"ad__adid" bson:"ad__adid"` // 设备码
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
	AD_User_Agent          string `json:"ad__user__agent" bson:"ad__user__agent"`
	AD_Tracker_Channel     string `json:"ad__tracker__channel" bson:"ad__tracker__channel"` // 推广渠道
	// 活动
	FirstRecharge string               `bson:"first_recharge" json:"first_recharge"`         // 是否首冲
	Receive       bool                 `bson:"first_receive" json:"first_receive"`           // 今日是否领取首冲奖励
	ReceiveDay    int32                `bson:"first_receive_day" json:"first_receive_day"`   // 已领取多少天的首冲奖励
	NoviceGift    string               `bson:"novice_gift" json:"novice_gift"`               // 是否购买新手礼包
	WeekCardMap   map[string]*WeekCard `bson:"weekcard,omitempty" json:"weekcard,omitempty"` // 金银铜卡
	SignDay       uint64               `bson:"sign_day" json:"sign_day"`                     // 已领取的签到奖励
	OnlineReward  []int32              `bson:"online_reward" json:"online_reward"`           // 已领取的在线奖励
	//时间
	Ctime      time.Time `bson:"ctime" json:"ctime"`             // 注册时间
	LoginTime  time.Time `bson:"login_time" json:"login_time"`   // 最后登录时间
	LogoutTime time.Time `bson:"logout_time" json:"logout_time"` // 离线时间
	ResetTime  time.Time `bson:"reset_time" json:"reset_time"`   // 上次跨天时间
	OnlineTime uint64    `bson:"online_time" json:"online_time"` // 当日在线时间(/s)
	StatusTime time.Time `bson:"status_time" json:"status_time"` // 状态时间
	//战绩
	Win  uint32 `bson:"win" json:"win"`   // 赢
	Lost uint32 `bson:"lost" json:"lost"` // 输
	Ping uint32 `bson:"ping" json:"ping"` // 平
	//充值提现
	Money          uint32                  `bson:"money" json:"money"`               // 充值总金额(分)
	CashOut        int32                   `bson:"cash_out" json:"cash_out"`         // 提现总金额(分)
	WithdrawLogMap map[string]*WithdrawLog `bson:"withdraw_log" json:"withdraw_log"` // 提现记录
	//最高
	TopDiamonds int64 `bson:"top_diamonds" json:"top_diamonds"` // 最高拥有钻石总金额
	TopCoins    int64 `bson:"top_coins" json:"top_coins"`       // 最高拥有金币总金额
	// TopChips    int64 `bson:"top_chips" json:"top_chips"`       // 最高拥有筹码总金额
	// TopCards    int64 `bson:"top_cards" json:"top_cards"`       // 最高拥有房卡总数
	//单局
	TopWinDiamond int64 `bson:"top_win_diamond" json:"top_win_diamond"` // 单局赢最高钻石金额
	TopWinCoin    int64 `bson:"top_win_coin" json:"top_win_coin"`       // 单局赢最高金币金额
	// TopWinChip    int64 `bson:"top_win_chip" json:"top_win_chip"`       // 单局赢最高筹码金额
	// 提现次数
	WithDrawCountMap map[string]int32 `bson:"withdraw_count" json:"withdraw_count"` // 提现次数
	//任务
	PhProgress int32                `bson:"ph_progress" json:"ph_progress"`       // 牌型当前进度
	PhRewardId int32                `bson:"ph_reward_id" json:"ph_reward_id"`     // 牌型当前id
	Task       map[string]*TaskInfo `bson:"task,omitempty" json:"task,omitempty"` // 已经完成或者还在继续的任务
	// 聊天
	FeedBackTimes int32     `bson:"feedback_times" json:"feedback_times"` // 客服回复前发送了多少条
	FeedBackLog   []ChatLog `bson:"feedback_log" json:"feedback_log"`     // 收到的回复
	// 分享
	ShareBelow    map[string]ShareData      `json:"shareBelow" bson:"share_below"`       // 分享的下级
	ShareBetLog   map[string]ShareAmountLog `json:"shareBetLog" bson:"share_bet_log"`    // 分享打码量记录
	ShareSuperior string                    `json:"shareSuperior" bson:"share_superior"` // 上级
	ShareWithdraw int64                     `json:"shareWithdraw" bson:"share_withdraw"` // 可提金额
	ShareTotal    int64                     `json:"shareTotal" bson:"share_total"`       // 总的可提现金
	//弹窗
	// BindWindow     bool `json:"bindWindow" bson:"bind_window"`         //游客绑定手机号弹窗
	WithdrawWindow bool `json:"withdrawWindow" bson:"withdraw_window"` //第一次满100提现弹窗
	//银行
	Bank         string `bson:"bank" json:"bank"`                   // 个人银行
	BankAccounts string `bson:"bank_accounts" json:"bank_accounts"` // 银行卡号
	IFSC         string `bson:"ifsc" json:"ifsc"`                   // IFSC
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
	PCSwitch        bool    `bson:"point_control_switch"`                //点控开关
	PCFactor        int32   `bson:"point_control_factor"`                //系数设置
	PCScore         int64   `bson:"point_control_score"`                 //分数设置
	PCScoreComplete int64   `bson:"point_control_score_complete"`        //已经达到的分数
	FluctuateLine   float64 `json:"fluctuateLine" bson:"fluctuate_line"` // 起伏线（非实时）
	PCSchedule      float64 `bson:"-"`                                   // 完成进度
	// 游戏
	LHDStrategy   LHDUserStrategy   `json:"LHDStrategy" bson:"lhd_strategy"`     //LHD策略
	SevenStrategy LHDUserStrategy   `json:"sevenStrategy" bson:"seven_strategy"` //7up策略
	CrashStrategy CrashUserStrategy `json:"crashStrategy" bson:"crash_strategy"` //crash类策略
	// 检测多账号
	IsIpRepeat    bool  `bson:"-"` // IP是否重复
	IsAppIdRepeat bool  `bson:"-"` // 设备号是否重复
	AndroidScore  int   `bson:"-"` // 安卓评分
	IsBlogger     bool  `bson:"-"` // 是否是博主账号 0:否; 1:是
	GameNumber    int64 `bson:"-"` // 总游戏局数

}

type PlayerUserCK struct {
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
	ShareBelow             int64  `json:"shareBelow" gorm:"column:share_below"`                            // 分享人数
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
	ShareWithdraw int64  `json:"shareWithdraw" gorm:"column:share_withdraw"` // 可提金额
	ShareTotal    int64  `json:"shareTotal" gorm:"column:share_total"`       // 总的可提现金
	//引导弹窗
	WithdrawWindow  bool `json:"withdrawWindow" gorm:"column:withdraw_window"`   //第一次满100提现弹窗
	WithdrawWindow2 bool `json:"withdrawWindow2" gorm:"column:withdraw_window2"` //第一次满200提现弹窗
	//vip
	VipLv            int     `json:"vip_lv" gorm:"column:vip_lv"`                                      //当前等级
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

	// RM相关
	RmRoiLimits         map[string]int32 `json:"rm_roi_limits" gorm:"column:rm_roi_limits;type:Map(String,Int32)"`         // rm roi控制策略生效次数
	RmRoiDayLimits      map[string]int32 `json:"rm_roi_day_limits" gorm:"column:rm_roi_day_limits;type:Map(String,Int32)"` // rm 每个自然日，对应ID策略生效次数
	RmWithout1stDrop    int              `json:"rm_without_1st_drop" gorm:"column:rm_without_1st_drop"`                    // rm 玩家在没有1Life时候首回合弃牌次数
	RmWithout1stNotDrop int              `json:"rm_without_1st_not_drop" gorm:"column:rm_without_1st_not_drop"`            // rm 玩家在没有1Life时候首回合未弃牌次数
	UserType            string           `bson:"-"`                                                                        // 用户类型名称
	PhotoUrl            string           `bson:"-"`                                                                        // 头像地址
	// 总资产、显示金额
	Assets         float64 `bson:"-"` // 总资产
	ShowAmount     float64 `bson:"-"` // 显示金额
	FDiamond       float64 `bson:"-"` // 钻石(彩金cash)
	FShadowDiamond float64 `bson:"-"` // 隐藏钻石(用户看不见的)
	FCoin          float64 `bson:"-"` // 金币(奖励金bonus)
	FMoney         float64 `bson:"-"` // 充值总金额(分)
	FCashOut       float64 `bson:"-"` // 提现总金额(分)
	FWin           float64 `bson:"-"` // 赢
	FGiveDiamond   float64 `bson:"-"` // 赠送彩金
	FOutDiamond    float64 `bson:"-"` // 可提现彩金
	// 检测多账号
	IsIpRepeat    bool    `bson:"-"` // IP是否重复
	IsAppIdRepeat bool    `bson:"-"` // 设备号是否重复
	AndroidScore  int     `bson:"-"` // 安卓评分
	IsBlogger     bool    `bson:"-"` // 是否是博主账号 0:否; 1:是
	GameNumber    int64   `bson:"-"` // 总游戏局数
	PCSchedule    float64 `bson:"-"` // 完成进度
}

// crash类游戏策略
type CrashUserStrategy struct {
	FYZS CrashFYZS `json:"fyzs" bson:"fyzs"` // 扶摇直上
}

type CrashFYZS struct {
	PlayTimes      int `json:"playTimes" bson:"play_times"`            // 扶摇直上:玩游戏次数(退出游戏才算一次)
	TiggerTimes    int `json:"tiggerTimes" bson:"tigger_times"`        // 扶摇直上:触发次数
	EvoTimes       int `json:"evoTimes" bson:"evo_times"`              // 扶摇直上:当次玩的局数(与playtimes关联)
	AllTiggerTimes int `json:"allTiggerTimes" bson:"all_tigger_times"` // 扶摇直上:累计触发次数
	AllEvoTimes    int `json:"allEvoTimes" bson:"all_evo_times"`       // 扶摇直上:累计玩游戏次数
}

// 玩家游戏局数
type UserGames struct {
	Id     int32  `bson:"-"` // 游戏类型ID
	Name   string `bson:"-"` // 游戏类型名称
	Number int32  `bson:"-"` // 游戏局数
}

// 玩家游戏局数页面显示内容
type UserGameInfo struct {
	Id               string `bson:"-"` // 游戏id
	TPNumber         int32  `bson:"-"` // 游戏局数
	LHDNumber        int32  `bson:"-"` // 游戏局数
	UPNumber         int32  `bson:"-"` // 游戏局数
	RMNumber         int32  `bson:"-"` // 游戏局数
	AKNumber         int32  `bson:"-"` // 游戏局数
	JOKERNumber      int32  `bson:"-"` // 游戏局数
	CRASHNumber      int32  `bson:"-"` // 游戏局数
	ABNumber         int32  `bson:"-"` // 游戏局数
	CPNumber         int32  `bson:"-"` // 游戏局数
	FJNumber         int32  `bson:"-"` // 游戏局数
	RBNumber         int32  `bson:"-"` // 游戏局数
	RMTwoNumber      int32  `bson:"-"` // 游戏局数
	TP2Number        int32  `bson:"-"` // 游戏局数
	G_F_Number       int32  `bson:"-"` // 游戏局数
	L_N_Number       int32  `bson:"-"` // 游戏局数
	G_G_Number       int32  `bson:"-"` // 游戏局数
	F_O_Number       int32  `bson:"-"` // 游戏局数
	S_W_Number       int32  `bson:"-"` // 游戏局数
	T_O_A_Number     int32  `bson:"-"` // 游戏局数
	W_B_S_Number     int32  `bson:"-"` // 游戏局数
	R_O_A_Number     int32  `bson:"-"` // 游戏局数
	F_T_Number       int32  `bson:"-"` // 游戏局数
	D_O_S_M_Number   int32  `bson:"-"` // 游戏局数
	L_O_P_Number     int32  `bson:"-"` // 游戏局数
	C_W_Number       int32  `bson:"-"` // 游戏局数
	S_S_Number       int32  `bson:"-"` // 游戏局数
	A_R_Number       int32  `bson:"-"` // 游戏局数
	J_K_Number       int32  `bson:"-"` // 游戏局数
	W_B_Number       int32  `bson:"-"` // 游戏局数
	SUP_S_Number     int32  `bson:"-"` // 游戏局数
	C_N_Number       int32  `bson:"-"` // 游戏局数
	Gal_G_Number     int32  `bson:"-"` // 游戏局数
	W_O_T_Q_Number   int32  `bson:"-"` // 游戏局数
	H_T_O_D_C_Number int32  `bson:"-"` // 游戏局数
	D_H_Number       int32  `bson:"-"` // 游戏局数
	L_RICH_Number    int32  `bson:"-"` // 游戏局数
	E_B_O_M_Number   int32  `bson:"-"` // 游戏局数
	D_O_M_Number     int32  `bson:"-"` // 游戏局数
	Q_O_B_Number     int32  `bson:"-"` // 游戏局数
	C_B_Number       int32  `bson:"-"` // 游戏局数
	Auto_R_Number    int32  `bson:"-"` // 游戏局数
	L_B_Number       int32  `bson:"-"` // 游戏局数
	L_R_Number       int32  `bson:"-"` // 游戏局数
	S_S_B_Number     int32  `bson:"-"` // 游戏局数
	D_T_Number       int32  `bson:"-"` // 游戏局数
	D_C_Number       int32  `bson:"-"` // 游戏局数
	FAN_T_Number     int32  `bson:"-"` // 游戏局数
	G_W_B_Number     int32  `bson:"-"` // 游戏局数
	B_B_Number       int32  `bson:"-"` // 游戏局数
}

// 修改玩家数据
type ModifyUserData struct {
	UserId      string `json:"userId"`     //用户id
	Photo       string `json:"photo"`      //头像
	NickName    string `json:"nickname"`   //名称
	Phone       string `json:"phone"`      //电话
	BankName    string `json:"bankName"`   //银行名称
	BankNumber  string `json:"bankNumber"` //银行卡号
	Ifsc        string `json:"ifsc"`       //ifsc
	BundleId    string `json:"bundleId"`
	OsVersion   string `json:"osVersion"`
	AfId        string `json:"afId"`
	MediaSource string `json:"mediaSource"`
	AfKey       string `json:"afKey"`
	AppId       string `json:"app_id"`
	RefGameId   string `json:"ref_game_id"`
	RefPkgName  string `json:"ref_pkg_name"`
}

// 修改点控
type PointControl struct {
	UserId string `json:"userid"` //用户id
	Switch bool   `json:"switch"` //开关
	Factor int32  `json:"factor"` //系数
	Score  int64  `json:"score"`  //分数
}

// IP限制
type IPwhite struct {
	IP       string    `bson:"_id" json:"id"`            // IP
	Operator string    `json:"operator" bson:"operator"` // 操作人
	Ctime    time.Time `bson:"ctime" json:"ctime"`       // 添加时间
}

// 银行卡黑名单
type CardBlacklist struct {
	Card     string    `bson:"_id" json:"id"`            // 银行卡号
	Operator string    `json:"operator" bson:"operator"` // 添加人
	Ctime    time.Time `bson:"ctime" json:"ctime"`       // 添加时间
}

// 提现用户黑名单
type WithdrawUserBlacklist struct {
	Userid   string    `bson:"_id" json:"id"`            // 用户id
	Operator string    `json:"operator" bson:"operator"` // 添加人
	Ctime    time.Time `bson:"ctime" json:"ctime"`       // 添加时间
}

// 设备码黑名单
type EquipmentBlacklist struct {
	Code     string    `bson:"_id" json:"id"`            // 设备码
	Operator string    `json:"operator" bson:"operator"` // 添加人
	Ctime    time.Time `bson:"ctime" json:"ctime"`       // 添加时间
}

const (
	TradeSuccess = 0 //交易成功
	TradeFail    = 1 //交易失败
	Tradeing     = 2 //交易中(下单状态)
	TradeGoods   = 3 //发货失败
)

var TradeResult = map[int]string{
	TradeSuccess: "成功",
	TradeFail:    "交易失败",
	//Tradeing:     "交易中",
	TradeGoods: "发货失败",
}

// TaskInfo 玩家的任务信息
type TaskInfo struct {
	Taskid   int32     `bson:"taskid" json:"taskid"`       //unique
	TaskType int32     `bson:"task_type" json:"task_type"` //任务类型
	Prize    int32     `bson:"prize" json:"prize"`         //任务状态 1:未完成 2:可领奖 3:已领奖
	Num      uint32    `bson:"num" json:"num"`             //完成数值
	Utime    time.Time `bson:"utime" json:"utime"`         //更新时间
}

// 交易记录
type TradeRecord struct {
	Id        string    `bson:"_id"`       //商户订单号(游戏内自定义订单号)
	Transid   string    `bson:"transid"`   //交易流水号(计费支付平台的交易流水号,微信订单号)
	Userid    string    `bson:"userid"`    //用户在商户应用的唯一标识(userid)
	Itemid    string    `bson:"itemid"`    //购买商品ID
	Amount    string    `bson:"amount"`    //购买商品数量
	Diamond   uint32    `bson:"diamond"`   //购买钻石数量
	Money     uint32    `bson:"money"`     //交易总金额(单位为分)
	Transtime string    `bson:"transtime"` //交易完成时间 yyyy-mm-dd hh24:mi:ss
	Result    int       `bson:"result"`    //交易结果(0–交易成功,1–交易失败,2-交易中,3-发货中)
	Waresid   uint32    `bson:"waresid"`   //商品编码(平台为应用内需计费商品分配的编码)
	Currency  string    `bson:"currency"`  //货币类型(RMB,CNY)
	Transtype int       `bson:"transtype"` //交易类型(0–支付交易)
	Feetype   int       `bson:"feetype"`   //计费方式(表示商品采用的计费方式)
	Paytype   uint32    `bson:"paytype"`   //支付方式(表示用户采用的支付方式,403-微信支付)
	Clientip  string    `bson:"clientip"`  //客户端ip
	Agent     string    `bson:"agent"`     //绑定的父级代理商游戏ID
	Atype     uint32    `bson:"atype"`     //代理包类型
	First     int       `bson:"first"`     //首次充值
	Utime     time.Time `bson:"utime"`     //本条记录更新unix时间戳
	DayStamp  time.Time `bson:"day_stamp"` //Time Today
	Ctime     time.Time `bson:"ctime"`     //本条记录生成unix时间戳
}

const (
	NOTICE_TYPE1 = 1 //活动公告
	NOTICE_TYPE2 = 2 //广播消息
)

const (
	NOTICE_ACT_TYPE0 = 0 //无操作消息
	NOTICE_ACT_TYPE1 = 1 //支付消息
	NOTICE_ACT_TYPE2 = 2 //活动消息
)

// 账号自增id
type ShopIDGen struct {
	Id         string `bson:"_id"`
	LastUserId string `bson:"last_user_id"`
}

// 商城
// type Shop struct {
// 	Id           string    `bson:"_id"`    //购买ID
// 	Status       int       `bson:"status"` //物品状态,1=热卖
// 	Propid       int       `bson:"propid"` //兑换的物品,1=钻石
// 	Payway       int       `bson:"payway"` //支付方式,1=RMB
// 	Number       uint32    `bson:"number"` //兑换的数量
// 	FNumber      float64   `bson:"-"`
// 	Give         uint32    `bson:"give"` //赠送数量
// 	FGive        float64   `bson:"-"`
// 	GiveType     int       `json:"giveType" bson:"give_type"`         //赠送方式 1:直接赠送  2:打码量
// 	FlowMultiple int       `json:"flowMultiple" bson:"flow_multiple"` //打码量倍数
// 	Price        uint32    `bson:"price"`                             //支付价格(单位元)
// 	FPrice       float64   `bson:"-"`
// 	Name         string    `bson:"name"`  //物品名字
// 	Info         string    `bson:"info"`  //物品信息
// 	Del          int       `bson:"del"`   //是否移除
// 	Ctime        time.Time `bson:"ctime"` //创建时间
// }

type Shop struct {
	Id     string `bson:"_id" json:"id"`        //购买ID
	Weight int    `bson:"weight" json:"weight"` //权重
	// Status       int       `bson:"status" json:"status"`              //物品状态,1=热卖
	// Propid       int       `bson:"propid" json:"propid"`              //兑换的物品,1=钻石,2=金币
	// Payway       int       `bson:"payway" json:"payway"`              //支付方式,1=卢比,,2=钻石
	Number       int64   `bson:"number" json:"number"` //兑换的数量
	FNumber      float64 `bson:"-"`
	Give         []int64 `bson:"give" json:"give"` //赠送数量
	FGive        float64 `bson:"-"`
	GiveType     []int   `json:"giveType" bson:"give_type"`         //赠送方式 1:直接赠送  2:打码量
	FlowMultiple []int   `json:"flowMultiple" bson:"flow_multiple"` //打码量倍数
	Price        uint32  `bson:"price" json:"price"`                //支付价格(单位卢比)
	FPrice       float64 `bson:"-"`
	Name         string  `bson:"name" json:"name"` //物品名字
	Show         []int   `bson:"show" json:"show"` //是否显示
	VB           int64   `bson:"vb" json:"vb"`     //vb释放
	FVB          float64 `bson:"-"`
	// Info         string    `bson:"info" json:"info"`                  //物品信息
	// Del          int       `bson:"del" json:"del"`                    //是否移除
	// Etime        time.Time `bson:"etime" json:"etime"`                //过期时间
	// Ctime        time.Time `bson:"ctime" json:"ctime"`                //创建时间
}

const (
	EnvType1  = 1  //"注册赠送钻石",
	EnvType2  = 2  //"注册赠送金币",
	EnvType3  = 3  //"注册赠送筹码",
	EnvType4  = 4  //"注册赠送房卡",
	EnvType5  = 5  //"绑定赠送",
	EnvType6  = 6  //"首充送n倍",
	EnvType7  = 7  //"首充送金币",
	EnvType8  = 8  //"救济金次数",
	EnvType9  = 9  //"转盘抽奖次数",
	EnvType10 = 10 //"破产金额",
	EnvType11 = 11 //"救济金额",
	EnvType12 = 12 //"虚假人数",
	EnvType13 = 13 //"机器人分配1",
	EnvType14 = 14 //"机器人分配2",
	EnvType15 = 15 //"机器人下注AI",
)

var EnvTypeValue = map[int]string{
	EnvType1: "注册赠送钻石",
	EnvType2: "注册赠送金币",
	EnvType3: "注册赠送筹码",
	EnvType4: "注册赠送房卡",
	//EnvType5:  "绑定赠送",
	//EnvType6:  "首充送n倍",
	//EnvType7:  "首充送金币",
	//EnvType8:  "救济金次数",
	//EnvType9:  "转盘抽奖次数",
	//EnvType10: "破产金额",
	//EnvType11: "救济金额",
	// EnvType12: "虚假人数",
	// EnvType13: "机器人分配1",
	// EnvType14: "机器人分配2",
	// EnvType15: "机器人下注AI",
}

var EnvTypeKey = map[int]string{
	EnvType1:  "regist_diamond",
	EnvType2:  "regist_coin",
	EnvType3:  "regist_chip",
	EnvType4:  "regist_card",
	EnvType5:  "build",
	EnvType6:  "first_pay_multi",
	EnvType7:  "first_pay_coin",
	EnvType8:  "relieve",
	EnvType9:  "prizedraw",
	EnvType10: "bankrupt_coin",
	EnvType11: "relieve_coin",
	EnvType12: "robot_num",
	EnvType13: "robot_allot1",
	EnvType14: "robot_allot2",
	EnvType15: "robot_bet",
}

var EnvKeyType = map[string]int{
	"regist_diamond":  EnvType1,
	"regist_coin":     EnvType2,
	"regist_chip":     EnvType3,
	"regist_card":     EnvType4,
	"build":           EnvType5,
	"first_pay_multi": EnvType6,
	"first_pay_coin":  EnvType7,
	"relieve":         EnvType8,
	"prizedraw":       EnvType9,
	"bankrupt_coin":   EnvType10,
	"relieve_coin":    EnvType11,
	"robot_num":       EnvType12,
	"robot_allot1":    EnvType13,
	"robot_allot2":    EnvType14,
	"robot_bet":       EnvType15,
}

// vip
type Vip struct {
	Id     string    `bson:"_id"`    //ID
	Level  int       `bson:"level"`  //等级
	Number uint32    `bson:"number"` //等级充值金额数量限制(分)
	Pay    uint32    `bson:"pay"`    //充值赠送百分比5=赠送充值的5%
	Prize  uint32    `bson:"prize"`  //赠送抽奖次数
	Kick   int       `bson:"kick"`   //经典场可踢人次数
	Ctime  time.Time `bson:"ctime"`  //创建时间
}

type Env struct {
	Key   string `bson:"_id" json:"key"`     //key
	Value int32  `bson:"value" json:"value"` //value
	Name  string `bson:"-"`                  //key
}

// 在线玩家
type OnlineUser struct {
	Id         string `json:"id"`          //用户id
	Nanme      string `json:"name"`        //呢称
	Channel    string `json:"channel"`     //渠道
	GameId     string `json:"game_id"`     //游戏id
	GameName   string `json:"-"`           //游戏名称
	RoomId     string `json:"room_id"`     //房间
	Asset      int64  `json:"asset"`       //总资产
	ShowAsset  int64  `json:"show_asset"`  //显示资产
	Cash       int64  `json:"cash"`        //彩金
	Bouns      int64  `json:"bouns"`       //游戏金
	OtherAsset int64  `json:"other_asset"` //其他资产
	Recharge   uint32 `json:"recharge"`    //已充值
	CashOut    uint32 `json:"cash_out"`    //已提现
	WinScore   int64  `json:"win_score"`   //当前赢分
}

// 流水日志
type LogWater struct {
	Id             string    `bson:"_id"`             //id
	Userid         string    `bson:"userid"`          //账户ID
	Name           string    `bson:"name"`            //名称
	MediaSource    string    `bson:"media_source"`    //渠道
	WaterDesc      string    `bson:"water_desc"`      //流水产生描述
	LType          int       `bson:"ltype"`           //type
	AddDiamond     int64     `bson:"add_diamond"`     //增加彩金
	AddCoin        int64     `bson:"add_coin"`        //增加奖励金
	AddOtherAsset  int64     `bson:"add_other_asset"` //增加其他资产
	ChangeAsset    int64     `bson:"change_asset"`    //变化总资产
	OldDiamond     int64     `bson:"old_diamond"`     //变化前彩金
	NowDiamond     int64     `bson:"now_diamond"`     //当前彩金
	OldCoin        int64     `bson:"old_coin"`        //变化前奖励金
	NowCoin        int64     `bson:"now_coin"`        //当前奖励金
	OldOtherAsset  int64     `bson:"old_other_asset"` //变化前其他资产
	NowOtherAsset  int64     `bson:"now_other_asset"` //当前其他资产
	FAddDiamond    float64   `bson:"-"`               //增加彩金
	FAddCoin       float64   `bson:"-"`               //增加奖励金
	FAddOtherAsset float64   `bson:"-"`               //增加其他资产
	FChangeAsset   float64   `bson:"-"`               //变化总资产
	OldShowAmount  float64   `bson:"-"`               // 显示金额
	NowShowAmount  float64   `bson:"-"`               // 显示金额
	FOldDiamond    float64   `bson:"-"`               //变化前彩金
	FNowDiamond    float64   `bson:"-"`               //当前彩金
	FOldCoin       float64   `bson:"-"`               //变化前奖励金
	FNowCoin       float64   `bson:"-"`               //当前奖励金
	FOldOtherAsset float64   `bson:"-"`               //变化前其他资产
	FNowOtherAsset float64   `bson:"-"`               //当前其他资产
	WaterId        string    `bson:"water_id"`        //流水号
	Ctime          time.Time `bson:"ctime"`           //流水产生时间
	Control        bool      `bson:"control"`         //点控局
	IsChangeAsset  bool      `bson:"-"`               //变化总资产是否负值
}
type TotalDetail struct {
	Number     int64   `bson:"-"` // 总局数
	WinNumber  int64   `bson:"-"` // 赢局
	LoseNumber int64   `bson:"-"` // 输局
	TieNumber  int64   `bson:"-"` // 平局
	Revenue    float64 `bson:"-"` // 玩家综合收益
	Win        int64   `bson:"-"` // 赢局
	WinAvg     float64 `bson:"-"` // 赢局结算均值
	WinMedian  float64 `bson:"-"` // 赢局结算中位数 -> 中间的的数，如果是双数集合 取中间两位的平均数
	WinMode    string  `bson:"-"` // 赢局结算众数 -> 出现次数最高的几个数
	Lose       int64   `bson:"-"` // 输局
	LoseAvg    float64 `bson:"-"` // 输局结算均值
	LoseMedian float64 `bson:"-"` // 输局结算中位数
	LoseMode   string  `bson:"-"` // 输局结算众数
}

type DetailList struct {
	WaterId             string    `gorm:"column:id"`
	Players             string    `gorm:"column:players"`    //参与玩家ID
	UserId              string    `gorm:"column:userid"`     //参与玩家ID
	BeginTime           int64     `gorm:"column:begin_time"` //开始时间
	STime               time.Time `gorm:"-"`
	EndTime             int64     `gorm:"column:end_time"` //结束时间
	ETime               time.Time `gorm:"-"`
	Gtype               int32     `gorm:"column:gtype"` //所在游戏 1：TP;2: LHD; 3：SEVEN; 4:RUMMY; 5:AK47; 6:JOKER; 7:CRASH
	GtypeName           string    `gorm:"-"`
	BetAmount           int64     `gorm:"column:bet_amount"`                                           // 打码量
	SettleScore         int64     `gorm:"column:settle_score"`                                         // 结算分
	BeforeScore         int64     `gorm:"column:before_score"`                                         // 账变前分数
	AfterScore          int64     `gorm:"column:after_score"`                                          // 账变后分数
	CashMingTax         int64     `gorm:"column:cash_ming_tax"`                                        // 彩金明税
	BonusMingTax        int64     `gorm:"column:bonus_ming_tax"`                                       // 奖励金明税
	CashAnTax           int64     `gorm:"column:cash_an_tax"`                                          // 彩金暗税
	BonusAnTax          int64     `gorm:"column:bonus_an_tax"`                                         // 奖励金暗税
	WinType             int8      `gorm:"column:win_type"`                                             // 1.赢 2.输 3.平 4.观察局
	RoomId              string    `gorm:"column:room_id"`                                              //房间ID
	RMGameOverReason    int       `gorm:"column:rm_game_over_reason"`                                  //rm游戏结束原因 (1.天胡,2.自摸胡牌,3.吃牌胡牌,4.对手弃牌)
	PlayerStageId       int       `gorm:"column:player_stage_id" json:"player_stage_id"`               //玩家阶段ID
	TPModel             int32     `gorm:"column:tp_model" json:"tp_model"`                             // tp模式 (0新手，1免费，2正常，3剧情，4控制策略)
	IsTPModel3StoryPlus bool      `gorm:"column:is_tp_model3story_plus" json:"is_tp_model3story_plus"` // 是否tp剧情局plus
	// TpStrategy        *TpStrategy `bson:"tp_strategy" json:"tp_strategy"` // tp 对局策略
	TpPrxdActive         bool   `json:"-" gorm:"column:tp_prxd_active"`           // 怦然心动生效
	TpPrxdHighCardRounds int32  `json:"-" gorm:"column:tp_prxd_high_card_rounds"` // 怦然心动生效时连续拿高牌局数
	TpLjsbActive         bool   `json:"-" gorm:"column:tp_ljsb_active"`           // 乐极生悲判定生效
	TpLjsbJLType         int8   `json:"-" gorm:"column:tp_ljsb_jl_type"`          // 生效时极乐状态: 1.超率极乐 2.超利极乐 3.利率极乐
	TpLjsbPy             bool   `json:"-" gorm:"column:tp_ljsb_py"`               // 被冤局
	TpLjsbPs             bool   `json:"-" gorm:"column:tp_ljsb_ps"`               // 冤杀局
	TpLjsbPd             bool   `json:"-" gorm:"column:tp_ljsb_pd"`               // 冤大局
	TpLjsbPt             bool   `json:"-" gorm:"column:tp_ljsb_pt"`               // 冤逃局
	TpGcyxActive         bool   `json:"-" gorm:"column:tp_gcyx_active"`           // 高潮涌现判定生效
	TpGcyxRobotNum       int32  `json:"-" gorm:"column:tp_gcyx_robot_num"`        // 高潮人机数
	TpGcyxRp             bool   `json:"-" gorm:"column:tp_gcyx_rp"`               // 压制概率生效
	TpGcyxBp             bool   `json:"-" gorm:"column:tp_gcyx_bp"`               // 恩赐概率生效
	FPlayerIds           string `bson:"-"`                                        //玩家ID
	FBeforeScore         string `bson:"-"`                                        //交易前携带
	FBet                 string `bson:"-"`                                        //下注
	FWin                 string `bson:"-"`                                        //中奖
	FScore               string `bson:"-"`                                        //结算(-tax)
	IsScore              bool   `bson:"-"`                                        //是否为正数
	FAfterScore          string `bson:"-"`                                        //交易后携带
	FControl             string `bson:"-"`                                        //控制模型

	IsCharge                 bool    `gorm:"column:is_charge" json:"is_charge"`                     //是否充值
	IsStrategy               bool    `gorm:"column:is_strategy" json:"is_strategy"`                 //是否是策略局
	PlayerFactor             int     `gorm:"column:player_factor" json:"player_factor"`             //玩家系数(开局)
	ChangeCardType           int     `gorm:"column:change_card_type" json:"change_card_type"`       //换牌局类型(开局换，局中换)
	ChangeCardTypeName       string  `gorm:"-"`                                                     //换牌局类型 0：没有 1：开局换牌 2：局中换牌
	ControlType              int     `gorm:"column:control_type" json:"control_type"`               //控制方式(开局)
	LhdStrategyId            int     `json:"-" gorm:"column:lhd_strategy_id"`                       //策略id
	UpStrategyId             int     `json:"-" gorm:"column:up_strategy_id"`                        //策略id
	AbStrategyId             int     `json:"-" gorm:"column:ab_strategy_id"`                        //策略id
	CpStrategyId             int     `json:"-" gorm:"column:cp_strategy_id"`                        //策略id
	RBStrategyId             int     `json:"-" gorm:"column:rb_strategy_id"`                        //策略id
	CrashStrategyType        []int32 `json:"-" gorm:"column:crash_strategy_type;type:Array(Int32)"` //策略类型
	RmcOk                    bool    `gorm:"column:rmc_ok" json:"rmc_ok"`                           // 有rm控制详情
	RmcCtype                 int32   `json:"-" gorm:"column:rmc_ctype"`                             // 0.随机,1.库存控制,2.ROI策略
	RmcRoiId                 string  `json:"-" gorm:"column:rmc_roi_id"`                            // ctype=2 roi策略id
	RmcDropId                string  `json:"-" gorm:"column:rmc_drop_id"`                           // 弃牌策略id
	RmcControlEffect         bool    `json:"-" gorm:"column:rmc_control_effect"`                    // rm进入控制模式 控制概率生效
	RmcHierarchy             int32   `json:"-" gorm:"column:rmc_hierarchy"`                         // 进入控制档位
	RmcDrawNumPlayer         int     `json:"-" gorm:"column:rmc_draw_num_player"`                   // 玩家抽牌张数
	RmcDrawNumRobot          int     `json:"-" gorm:"column:rmc_draw_num_robot"`                    // 人机抽牌张数
	RmcNotDrawRobot1st       bool    `json:"-" gorm:"column:rmc_not_draw_robot1st"`                 // 人机保顺金
	RmcNotDrawPlayer1st      bool    `json:"-" gorm:"column:rmc_not_draw_player1st"`                // 玩家保顺金
	RmcControlRobotDraw      bool    `json:"-" gorm:"column:rmc_control_robot_draw"`                // 人机进入干预
	RmcControlRobotDrawRound int32   `json:"-" gorm:"column:rmc_control_robot_draw_round"`          // 人机进入干预回合

	// mines地雷
	MinesMines     int32   `json:"-" gorm:"column:mines_mines"`      // 埋雷数
	MinesStep      int32   `json:"-" gorm:"column:mines_step"`       // 选了几个格子了
	MinesMultiple  float64 `json:"-" gorm:"column:mines_multiple"`   // 返奖倍数
	MinesAutoMines bool    `json:"-" gorm:"column:mines_auto_mines"` // 自动对局
	MinesForce     bool    `json:"-" gorm:"column:mines_force"`      // 强制结算
}

// 对局详情
type Detail struct {
	WaterId                            string         `bson:"_id" json:"_id"`               //局号
	BeginTime                          int64          `bson:"begin_time" json:"begin_time"` //开始时间
	STime                              time.Time      `bson:"-"`
	EndTime                            int64          `bson:"end_time" json:"end_time"` //结束时间
	ETime                              time.Time      `bson:"-"`
	Gtype                              int32          `bson:"gtype" json:"gtype"` //所在游戏 1：TP;2: LHD; 3：SEVEN; 4:RUMMY; 5:AK47; 6:JOKER; 7:CRASH
	GtypeName                          string         `bson:"-"`
	RoomId                             string         `bson:"room_id" json:"room_id"`                                                                     //房间ID
	DeskId                             string         `bson:"desk_id" json:"desk_id"`                                                                     //桌子ID
	Players                            string         `bson:"players" json:"players"`                                                                     //参与玩家ID
	CardTypeId                         int            `bson:"card_type_id" json:"card_type_id"`                                                           //牌型ID
	PlayerStageId                      int            `bson:"player_stage_id" json:"player_stage_id"`                                                     //玩家阶段ID
	ChangeCardType                     int            `bson:"change_card_type" json:"change_card_type"`                                                   // 换牌局类型 0：没有 1：开局换牌 2：局中换牌
	WinScore                           int            `bson:"win_score" json:"win_score"`                                                                 //当前赢分
	PlayerFactor                       int            `bson:"player_factor" json:"player_factor"`                                                         //玩家系数
	ControlType                        int            `bson:"control_type" json:"control_type"`                                                           //控制方式 1:当前赢分;2：房间系数
	WinScore1                          int            `bson:"win_score_1" json:"win_score_1"`                                                             //当局赢分
	WinScore2                          int            `bson:"win_score_2" json:"win_score_2"`                                                             //当前赢分
	ChargeMoney                        int            `bson:"charge_money" json:"charge_money"`                                                           //充值金额
	IsStrategy                         bool           `bson:"is_strategy" json:"is_strategy"`                                                             //是否是策略局
	StrategyType                       int            `json:"strategyType" bson:"strategy_type"`                                                          // 策略局类型
	IsCharge                           bool           `bson:"is_charge" json:"is_charge"`                                                                 // 是否充值
	RealGame                           bool           `bson:"real_game" json:"real_game"`                                                                 //是否真人对局
	NonDirty                           []uint32       `bson:"non_dirty" json:"non_dirty"`                                                                 //没有污染的人机座位
	ShowCards                          string         `bson:"show_cards" json:"show_cards"`                                                               //亮牌
	Rtype                              int32          `bson:"rtype" json:"rtype"`                                                                         // 房间类型 0:自由；1：私人；2：百人
	Gmode                              int32          `bson:"gmode" json:"gmode"`                                                                         // 私人房类型：0：真金 1：娱乐模式
	IsMustLose                         bool           `bson:"is_must_lose" json:"is_must_lose"`                                                           //是否是必输局
	MustLoseScore                      int            `bson:"must_lose_score" json:"must_lose_score"`                                                     //必输分数线
	CoreRobotSeat                      uint32         `bson:"core_robot_seat" json:"core_robot_seat"`                                                     //核心人机位置
	IsUpCardPool                       bool           `bson:"is_up_card_pool" json:"is_up_card_pool"`                                                     //是否升档牌库
	CardPoolId                         int            `bson:"card_pool_id" json:"card_pool_id"`                                                           //牌库编号
	DealType                           int            `bson:"deal_type" json:"deal_type"`                                                                 //发牌方式
	IsTriggerWinScoreLimit             bool           `bson:"is_trigger_win_score_limit" json:"is_trigger_win_score_limit"`                               //是否触发赢分限制
	IsTriggerMustLose                  bool           `bson:"is_trigger_must_lose" json:"is_trigger_must_lose"`                                           //是否触发必输控制
	IsTriggerMustLoseDrawCard          bool           `bson:"is_trigger_must_lose_draw_card" json:"is_trigger_must_lose_draw_card"`                       //是否触发必输控制摸牌限制
	IsTriggerMustLoseLastCard          bool           `bson:"is_trigger_must_lose_last_card" json:"is_trigger_must_lose_last_card"`                       //是否触发必输控制上家出牌限制
	IsTriggerMustLoseCoreRobotGoodCard bool           `bson:"is_trigger_must_lose_core_robot_good_card" json:"is_trigger_must_lose_core_robot_good_card"` //是否触发必输控制核心人机摸好牌
	TPModel                            int32          `bson:"tp_model" json:"tp_model"`                                                                   // tp模式 (0新手，1免费，2正常，3剧情，4控制策略)
	IsTPModel3StoryPlus                bool           `bson:"is_tp_model3story_plus" json:"is_tp_model3story_plus"`                                       // 是否tp剧情局plus
	RMControl                          *RMControl     `bson:"rm_control" json:"rm_control"`                                                               //rm 控制详情
	RMGameOverReason                   int            `bson:"rm_game_over_reason" json:"rm_game_over_reason"`                                             //rm游戏结束原因 (1.天胡,2.自摸胡牌,3.吃牌胡牌,4.对手弃牌)
	FPlayerFactor                      float64        `bson:"-" json:"-"`                                                                                 //玩家系数
	ChangeCardTypeName                 string         `bson:"-"`                                                                                          //换牌局类型 0：没有 1：开局换牌 2：局中换牌
	ControlTypeName                    string         `bson:"-"`                                                                                          //控制方式 1:当前赢分;2：房间系数
	FWinScore                          float64        `bson:"-"`                                                                                          //当前赢分(开局)
	FWinScore1                         float64        `bson:"-"`                                                                                          //当局赢分(局中)
	FWinScore2                         float64        `bson:"-"`                                                                                          //当前赢分(局中)
	FChargeMoney                       float64        `bson:"-"`                                                                                          //充值金额
	FPlayerWin                         float64        `bson:"-"`                                                                                          //玩家赢分
	NewBieProbeId                      int32          `bson:"-"`                                                                                          // 新手试探期/稳定局策略id
	NewBieProbeRound                   int32          `bson:"-"`                                                                                          // 玩家新手试探期/稳定局策略生效局数
	FPlayerIds                         string         `bson:"-"`                                                                                          //玩家ID
	FBeforeScore                       string         `bson:"-"`                                                                                          //交易前携带
	FBet                               string         `bson:"-"`                                                                                          //下注
	FWin                               string         `bson:"-"`                                                                                          //中奖
	FScore                             string         `bson:"-"`                                                                                          //结算(-tax)
	IsScore                            bool           `bson:"-"`                                                                                          //是否为正数
	FAfterScore                        string         `bson:"-"`                                                                                          //交易后携带
	FControl                           string         `bson:"-"`                                                                                          //控制模型
	TPDetail                           []TPDetail     //TP详情
	JOKERDetail                        []*JOKERDetail //JOKER详情
	AK47Detail                         []*AK47Detail  //AK47详情
	RMDetail                           []*RMDetail    //RM详情
	LHDetail                           *LHDetail      //LH详情
	UPDetail                           *UPDetail      //7up详情
	CRASHDetail                        *CRASHDetail   //crash
	ABDetail                           *ABDetail      //ab详情
	*CPDetail                                         //彩票
	*RBDetail                                         //红黑
	MinesDetail                        *MinesDetail   // 地雷
}

type UserDetail struct {
	WaterId    string    `gorm:"column:id" bson:"id" json:"id"` //局号
	BeginTime  int64     `bson:"begin_time" json:"begin_time"`  //开始时间
	STime      time.Time `bson:"-"`
	EndTime    int64     `bson:"end_time" json:"end_time"` //结束时间
	ETime      time.Time `bson:"-"`
	Gtype      int32     `bson:"gtype" json:"gtype"` //所在游戏 1：TP;2: LHD; 3：SEVEN; 4:RUMMY; 5:AK47; 6:JOKER; 7:CRASH
	GtypeName  string    `bson:"-"`
	ScoreUntax int64     `json:"scoreUntax" bson:"score_untax"`
	AfterScore int64     `gorm:"column:after_score"` // 账变后分数
}

// TP游戏详情
type TPDetail struct {
	SeatId           uint32         `json:"seat_id" bson:"seat_id"`                         //座位ID
	UserId           string         `json:"user_id" bson:"user_id"`                         //用户ID
	Cards            []uint32       `json:"cards" bson:"cards"`                             //牌型
	Score            int64          `json:"score" bson:"score"`                             //结算
	BeforeScore      int64          `json:"before_score" bson:"before_score"`               //账变前分数
	AfterScore       int64          `json:"after_score" bson:"after_score"`                 //账变后分数
	BeforeCash       int64          `json:"before_cash" bson:"before_cash"`                 //账变前彩金
	AfterCash        int64          `json:"after_cash" bson:"after_cash"`                   //账变后彩金
	BeforeBonus      int64          `json:"before_bonus" bson:"before_bonus"`               //账变前奖励金
	AfterBonus       int64          `json:"after_bonus" bson:"after_bonus"`                 //账变后奖励金
	Bet              int64          `json:"bet" bson:"bet"`                                 //总投注
	Bottom           int64          `json:"bottom" bson:"bottom"`                           //底注
	CashStock        int64          `json:"cash_stock" bson:"cash_stock"`                   //彩金库存
	BonusStock       int64          `json:"bonus_stock" bson:"bonus_stock"`                 //奖励金库存
	CashMingTax      int64          `json:"cash_ming_tax" bson:"cash_ming_tax"`             //彩金明税
	BonusMingTax     int64          `json:"bonus_ming_tax" bson:"bonus_ming_tax"`           //奖励金明税
	CashAnTax        int64          `json:"cash_an_tax" bson:"cash_an_tax"`                 //彩金暗税
	BonusAnTax       int64          `json:"bonus_an_tax" bson:"bonus_an_tax"`               //奖励金暗税
	NewBieProbeId    int32          `json:"new_bie_probe_id" bson:"new_bie_probe_id"`       // 新手试探期/稳定局策略id
	NewBieProbeRound int32          `json:"new_bie_probe_round" bson:"new_bie_probe_round"` // 玩家新手试探期/稳定局策略生效局数
	FScore           float64        `bson:"-"`                                              //结算
	FBeforeScore     float64        `bson:"-"`                                              //账变前分数
	FAfterScore      float64        `bson:"-"`                                              //账变后分数
	FBeforeCash      float64        `bson:"-"`                                              //账变前彩金
	FAfterCash       float64        `bson:"-"`                                              //账变后彩金
	FBeforeBonus     float64        `bson:"-"`                                              //账变前奖励金
	FAfterBonus      float64        `bson:"-"`                                              //账变后奖励金
	FBet             float64        `bson:"-"`                                              //总投注
	FBottom          float64        `bson:"-"`                                              //底注
	FCashStock       float64        `bson:"-"`                                              //彩金库存
	FBonusStock      float64        `bson:"-"`                                              //奖励金库存
	FCashMingTax     float64        `bson:"-"`                                              //彩金明税
	FBonusMingTax    float64        `bson:"-"`                                              //奖励金明税
	FCashAnTax       float64        `bson:"-"`                                              //彩金暗税
	FBonusAnTax      float64        `bson:"-"`                                              //奖励金暗税
	TPDetailTurn     []TPDetailTurn `json:"tp_detail_turn" bson:"tp_detail_turn"`           //轮次详情
}

// 操作轮次详情
type TPDetailTurn struct {
	Turn      int    `json:"turn" bson:"turn"`           //轮次ID
	Operation string `json:"operation" bson:"operation"` //操作
}

// Joker游戏详情
type JOKERDetail struct {
	SeatId          uint32             `json:"seat_id" bson:"seat_id"`               //座位ID
	UserId          string             `json:"user_id" bson:"user_id"`               //用户ID
	Cards           []uint32           `json:"cards" bson:"cards"`                   //牌型
	Score           int64              `json:"score" bson:"score"`                   //结算
	BeforeScore     int64              `json:"before_score" bson:"before_score"`     //账变前分数
	AfterScore      int64              `json:"after_score" bson:"after_score"`       //账变后分数
	BeforeCash      int64              `json:"before_cash" bson:"before_cash"`       //账变前彩金
	AfterCash       int64              `json:"after_cash" bson:"after_cash"`         //账变后彩金
	BeforeBonus     int64              `json:"before_bonus" bson:"before_bonus"`     //账变前奖励金
	AfterBonus      int64              `json:"after_bonus" bson:"after_bonus"`       //账变后奖励金
	Bet             int64              `json:"bet" bson:"bet"`                       //总投注
	Bottom          int64              `json:"bottom" bson:"bottom"`                 //底注
	CashStock       int64              `json:"cash_stock" bson:"cash_stock"`         //彩金库存
	BonusStock      int64              `json:"bonus_stock" bson:"bonus_stock"`       //奖励金库存
	CashMingTax     int64              `json:"cash_ming_tax" bson:"cash_ming_tax"`   //彩金明税
	BonusMingTax    int64              `json:"bonus_ming_tax" bson:"bonus_ming_tax"` //奖励金明税
	CashAnTax       int64              `json:"cash_an_tax" bson:"cash_an_tax"`       //彩金暗税
	BonusAnTax      int64              `json:"bonus_an_tax" bson:"bonus_an_tax"`     //奖励金暗税
	FScore          float64            `bson:"-"`                                    //结算
	FBeforeScore    float64            `bson:"-"`                                    //账变前分数
	FAfterScore     float64            `bson:"-"`                                    //账变后分数
	FBeforeCash     float64            `bson:"-"`                                    //账变前彩金
	FAfterCash      float64            `bson:"-"`                                    //账变后彩金
	FBeforeBonus    float64            `bson:"-"`                                    //账变前奖励金
	FAfterBonus     float64            `bson:"-"`                                    //账变后奖励金
	FBet            float64            `bson:"-"`                                    //总投注
	FBottom         float64            `bson:"-"`                                    //底注
	FCashStock      float64            `bson:"-"`                                    //彩金库存
	FBonusStock     float64            `bson:"-"`                                    //奖励金库存
	FCashMingTax    float64            `bson:"-"`                                    //彩金明税
	FBonusMingTax   float64            `bson:"-"`                                    //奖励金明税
	FCashAnTax      float64            `bson:"-"`                                    //彩金暗税
	FBonusAnTax     float64            `bson:"-"`                                    //奖励金暗税
	JOKERDetailTurn []*JOKERDetailTurn `json:"tp_detail_turn" bson:"tp_detail_turn"` //轮次详情
	// LHDetail     LHDetail        `json:"lhUserDetail" bson:"lh_user_detail"`   //
}

type JOKERDetailTurn struct {
	Turn      int    `json:"turn" bson:"turn"`           //轮次ID
	Operation string `json:"operation" bson:"operation"` //操作
}

// AK47
type AK47Detail struct {
	SeatId         uint32            `json:"seat_id" bson:"seat_id"`               //座位ID
	UserId         string            `json:"user_id" bson:"user_id"`               //用户ID
	Cards          []uint32          `json:"cards" bson:"cards"`                   //牌型
	ChangeCards    []uint32          `json:"change_cards" bson:"change_cards"`     //变牌
	Score          int64             `json:"score" bson:"score"`                   //结算
	BeforeScore    int64             `json:"before_score" bson:"before_score"`     //账变前分数
	AfterScore     int64             `json:"after_score" bson:"after_score"`       //账变后分数
	BeforeCash     int64             `json:"before_cash" bson:"before_cash"`       //账变前彩金
	AfterCash      int64             `json:"after_cash" bson:"after_cash"`         //账变后彩金
	BeforeBonus    int64             `json:"before_bonus" bson:"before_bonus"`     //账变前奖励金
	AfterBonus     int64             `json:"after_bonus" bson:"after_bonus"`       //账变后奖励金
	Bet            int64             `json:"bet" bson:"bet"`                       //总投注
	Bottom         int64             `json:"bottom" bson:"bottom"`                 //底注
	CashStock      int64             `json:"cash_stock" bson:"cash_stock"`         //彩金库存
	BonusStock     int64             `json:"bonus_stock" bson:"bonus_stock"`       //奖励金库存
	CashMingTax    int64             `json:"cash_ming_tax" bson:"cash_ming_tax"`   //彩金明税
	BonusMingTax   int64             `json:"bonus_ming_tax" bson:"bonus_ming_tax"` //奖励金明税
	CashAnTax      int64             `json:"cash_an_tax" bson:"cash_an_tax"`       //彩金暗税
	BonusAnTax     int64             `json:"bonus_an_tax" bson:"bonus_an_tax"`     //奖励金暗税
	FScore         float64           `bson:"-"`                                    //结算
	FBeforeScore   float64           `bson:"-"`                                    //账变前分数
	FAfterScore    float64           `bson:"-"`                                    //账变后分数
	FBeforeCash    float64           `bson:"-"`                                    //账变前彩金
	FAfterCash     float64           `bson:"-"`                                    //账变后彩金
	FBeforeBonus   float64           `bson:"-"`                                    //账变前奖励金
	FAfterBonus    float64           `bson:"-"`                                    //账变后奖励金
	FBet           float64           `bson:"-"`                                    //总投注
	FBottom        float64           `bson:"-"`                                    //底注
	FCashStock     float64           `bson:"-"`                                    //彩金库存
	FBonusStock    float64           `bson:"-"`                                    //奖励金库存
	FCashMingTax   float64           `bson:"-"`                                    //彩金明税
	FBonusMingTax  float64           `bson:"-"`                                    //奖励金明税
	FCashAnTax     float64           `bson:"-"`                                    //彩金暗税
	FBonusAnTax    float64           `bson:"-"`                                    //奖励金暗税
	AK47DetailTurn []*AK47DetailTurn `json:"tp_detail_turn" bson:"tp_detail_turn"` //轮次详情
	// LHDetail     LHDetail        `json:"lhUserDetail" bson:"lh_user_detail"`   //
}
type AK47DetailTurn struct {
	Turn      int    `json:"turn" bson:"turn"`           //轮次ID
	Operation string `json:"operation" bson:"operation"` //操作
}

type RMControl struct {
	Ctype                 int32  `json:"ctype" bson:"ctype"`                                       // 0.随机,1.库存控制,2.ROI策略
	RoiId                 string `json:"roi_id" bson:"roi_id"`                                     // ctype=2 roi策略id
	DropId                string `json:"drop_id" bson:"drop_id"`                                   // 弃牌策略id
	ControlEffect         bool   `json:"control_effect" bson:"control_effect"`                     // rm进入控制模式 控制概率生效
	Hierarchy             int32  `json:"hierarchy" bson:"hierarchy"`                               // 进入控制档位
	DrawNumPlayer         int    `json:"draw_num_player" bson:"draw_num_player"`                   // 玩家抽牌张数
	DrawNumRobot          int    `json:"draw_num_robot" bson:"draw_num_robot"`                     // 人机抽牌张数
	NotDrawRobot1st       bool   `json:"not_draw_robot1st" bson:"not_draw_robot1st"`               // 人机保顺金
	NotDrawPlayer1st      bool   `json:"not_draw_player1st" bson:"not_draw_player1st"`             // 玩家保顺金
	ControlRobotDraw      bool   `json:"control_robot_draw" bson:"control_robot_draw"`             // 人机进入干预
	ControlRobotDrawRound int32  `json:"control_robot_draw_round" bson:"control_robot_draw_round"` // 人机进入干预回合
}

// Rummy
type RMDetail struct {
	SeatId        uint32     `json:"seat_id" bson:"seat_id"`               //座位ID
	UserId        string     `json:"user_id" bson:"user_id"`               //用户ID
	Result        string     `json:"result" bson:"result"`                 //结果
	Score         int64      `json:"score" bson:"score"`                   //结算
	BeforeScore   int64      `json:"before_score" bson:"before_score"`     //账变前分数
	AfterScore    int64      `json:"after_score" bson:"after_score"`       //账变后分数
	BeforeCash    int64      `json:"before_cash" bson:"before_cash"`       //账变前彩金
	AfterCash     int64      `json:"after_cash" bson:"after_cash"`         //账变后彩金
	BeforeBonus   int64      `json:"before_bonus" bson:"before_bonus"`     //账变前奖励金
	AfterBonus    int64      `json:"after_bonus" bson:"after_bonus"`       //账变后奖励金
	FinalCards    [][]uint32 `json:"final_cards" bson:"final_cards"`       //最终牌型
	InitCards     []uint32   `json:"init_cards" bson:"init_cards"`         //初始牌型
	WildCard      uint32     `json:"wild_card" bson:"wild_card"`           //万能牌
	MoCards       []uint32   `json:"mo_cards" bson:"mo_cards"`             //每轮摸到的牌
	ChuCards      []uint32   `json:"chu_cards" bson:"chu_cards"`           //每轮出的牌
	CashStock     int64      `json:"cash_stock" bson:"cash_stock"`         //彩金库存
	BonusStock    int64      `json:"bonus_stock" bson:"bonus_stock"`       //奖励金库存
	CashMingTax   int64      `json:"cash_ming_tax" bson:"cash_ming_tax"`   //彩金明税
	BonusMingTax  int64      `json:"bonus_ming_tax" bson:"bonus_ming_tax"` //奖励金明税
	CashAnTax     int64      `json:"cash_an_tax" bson:"cash_an_tax"`       //彩金暗税
	BonusAnTax    int64      `json:"bonus_an_tax" bson:"bonus_an_tax"`     //奖励金暗税
	FScore        float64    `bson:"-"`                                    //结算
	FBeforeScore  float64    `bson:"-"`                                    //账变前分数
	FAfterScore   float64    `bson:"-"`                                    //账变后分数
	FBeforeCash   float64    `bson:"-"`                                    //账变前彩金
	FAfterCash    float64    `bson:"-"`                                    //账变后彩金
	FBeforeBonus  float64    `bson:"-"`                                    //账变前奖励金
	FAfterBonus   float64    `bson:"-"`                                    //账变后奖励金
	FCashStock    float64    `bson:"-"`                                    //彩金库存
	FBonusStock   float64    `bson:"-"`                                    //奖励金库存
	FCashMingTax  float64    `bson:"-"`                                    //彩金明税
	FBonusMingTax float64    `bson:"-"`                                    //奖励金明税
	FCashAnTax    float64    `bson:"-"`                                    //彩金暗税
	FBonusAnTax   float64    `bson:"-"`                                    //奖励金暗税
}

// 龙虎斗游戏详情
type LHDetail struct {
	Winner      int     `json:"winner" bson:"winner"`            //赢家
	DragonValue string  `json:"dragonValue" bson:"dragon_value"` //龙牌值
	TigerValue  string  `json:"tigerValue" bson:"tiger_value"`   //虎牌值
	Bets        int64   `json:"bets" bson:"bets"`                //总下注
	FBets       float64 `bson:"-"`
	PlayerWin   int64   `json:"playerWin" bson:"player_win"` //玩家赢分

	StrategyTigger bool `json:"strategyTigger" bson:"strategy_tigger"` //策略生效

	StrategyId int            `json:"strategyId" bson:"strategy_id"` //策略id
	IsSuppress bool           `json:"isSuppress" bson:"is_suppress"` //是否压制
	FPlayerWin float64        `bson:"-"`
	UserDetail []LHUserDetail `json:"userDetail" bson:"user_detail"` //用户详细
}

type UPDetail struct {
	Winner     int              `json:"winner" bson:"winner"`          //赢家
	PointValue int32            `json:"pointValue" bson:"point_value"` //点数
	Bets       int64            `json:"bets" bson:"bets"`              //总下注
	PlayerWin  int64            `json:"playerWin" bson:"player_win"`   //玩家赢分
	StrategyId int              `json:"strategyId" bson:"strategy_id"` //策略id
	Crit7up    bool             `json:"crit7up" bson:"crit_7up"`       //是否暴击7up对局,区分老数据
	Crited     int16            `json:"crited" bson:"crited"`          //触发暴击位置数
	CritOdds   map[uint32]int32 `json:"critOdds" bson:"crit_odds"`     //触发暴击位置和倍数
	UserDetail []LHUserDetail   `json:"userDetail" bson:"user_detail"` //用户详细

	FBets      float64      `bson:"-"`
	FPlayerWin float64      `bson:"-"`
	FCritOdds  []FLHCritOdd `bson:"-"`
}

// 玩家下注详情
type LHUserDetail struct {
	Userid        string           `json:"userid" bson:"userid"`                 //用户id
	Dragon        int64            `json:"dragon" bson:"dragon"`                 //龙下注
	Tiger         int64            `json:"tiger" bson:"tiger"`                   //虎下注
	Tie           int64            `json:"tie" bson:"tie"`                       //和下注
	SeatBets      map[uint32]int64 `json:"seatBets" bson:"seat_bets"`            //_位置下注: 0小,1大,2,3,4,...,12
	Crit          int32            `json:"crit" bson:"crit"`                     //_暴击位押中次数
	Result        string           `json:"result" bson:"result"`                 //输赢平
	Win           int64            `json:"win" bson:"win"`                       //输赢
	BeforeScore   int64            `json:"beforeScore" bson:"before_score"`      //账变前分数
	AfterScore    int64            `json:"after_score" bson:"after_score"`       //账变后分数
	BeforeCash    int64            `json:"before_cash" bson:"before_cash"`       //账变前彩金
	AfterCash     int64            `json:"after_cash" bson:"after_cash"`         //账变后彩金
	BeforeBonus   int64            `json:"before_bonus" bson:"before_bonus"`     //账变前奖励金
	AfterBonus    int64            `json:"after_bonus" bson:"after_bonus"`       //账变后奖励金
	CashStock     int64            `json:"cash_stock" bson:"cash_stock"`         //彩金库存
	BonusStock    int64            `json:"bonus_stock" bson:"bonus_stock"`       //奖励金库存
	CashMingTax   int64            `json:"cash_ming_tax" bson:"cash_ming_tax"`   //彩金明税
	BonusMingTax  int64            `json:"bonus_ming_tax" bson:"bonus_ming_tax"` //奖励金明税
	CashAnTax     int64            `json:"cash_an_tax" bson:"cash_an_tax"`       //彩金暗税
	BonusAnTax    int64            `json:"bonus_an_tax" bson:"bonus_an_tax"`     //奖励金暗税
	FDragon       float64          `bson:"-"`                                    //龙下注
	FTiger        float64          `bson:"-"`                                    //虎下注
	FTie          float64          `bson:"-"`                                    //和下注
	FWin          float64          `bson:"-"`                                    //输赢
	FBeforeScore  float64          `bson:"-"`                                    //账变前分数
	FAfterScore   float64          `bson:"-"`                                    //账变后分数
	FBeforeCash   float64          `bson:"-"`                                    //账变前彩金
	FAfterCash    float64          `bson:"-"`                                    //账变后彩金
	FBeforeBonus  float64          `bson:"-"`                                    //账变前奖励金
	FAfterBonus   float64          `bson:"-"`                                    //账变后奖励金
	FCashStock    float64          `bson:"-"`                                    //彩金库存
	FBonusStock   float64          `bson:"-"`                                    //奖励金库存
	FCashMingTax  float64          `bson:"-"`                                    //彩金明税
	FBonusMingTax float64          `bson:"-"`                                    //奖励金明税
	FCashAnTax    float64          `bson:"-"`                                    //彩金暗税
	FBonusAnTax   float64          `bson:"-"`                                    //奖励金暗税
	Crit7up       bool             `bson:"-"`                                    // 是否暴击对局
	FSeatBets     []FLHSeatBets    `bson:"-"`                                    // _位置下注
}

type FLHCritOdd struct {
	Seat     uint32
	FSeat    string
	Multiple string
}

// UP用户下注位置详情
type FLHSeatBets struct {
	Seat     uint32
	FSeat    string
	Bets     string
	Hit      bool // 中奖
	Crit     bool // 暴击并中奖
	Multiple string
}

type CRASHDetail struct {
	Result       string            `json:"result" bson:"result"` //开奖结果
	Bets         int64             `json:"bets" bson:"bets"`     //总下注
	FBets        float64           `bson:"-"`
	PlayerLose   int64             `json:"playerLose" bson:"player_lose"` //玩家赢分
	FPlayerWin   float64           `bson:"-"`
	StrategyType []int             `json:"strategyType" bson:"strategy_type"` //策略类型 0：常规；1：扶摇直上
	UserDetail   []CRASHUserDetail `json:"userDetail" bson:"user_detail"`     //用户详细
	IsStrategy   bool              `bson:"-"`                                 // 判断是否有策略
}

// crash用户详情
type CRASHUserDetail struct {
	Userid       string `json:"userid" bson:"userid"`                 //用户id
	Nickname     string `json:"nickname" bson:"nickname"`             //用户名称
	Photo        string `json:"photo" bson:"photo"`                   //用户头像
	VipLv        int32  `json:"vipLv" bson:"vip_lv"`                  //用户vip等级
	Observe      bool   `json:"observe" bson:"observe"`               //是否观察局
	Bet          int64  `json:"bet" bson:"bet"`                       //下注
	Bet0         int64  `json:"bet0" bson:"bet0"`                     //位置0下注
	Bet1         int64  `json:"bet1" bson:"bet1"`                     //位置1下注
	Multiple     string `json:"multiple" bson:"multiple"`             //位置0逃脱倍数
	Multiple1    string `json:"multiple1" bson:"multiple1"`           //位置1逃脱倍数
	Result       string `json:"result" bson:"result"`                 //输赢平
	Win          int64  `json:"win" bson:"win"`                       //输赢
	Win0         int64  `json:"win0" bson:"win0"`                     //位置0输赢
	Win1         int64  `json:"win1" bson:"win1"`                     //位置1输赢
	BeforeScore  int64  `json:"beforeScore" bson:"before_score"`      //账变前分数
	AfterScore   int64  `json:"after_score" bson:"after_score"`       //账变后分数
	BeforeCash   int64  `json:"before_cash" bson:"before_cash"`       //账变前彩金
	AfterCash    int64  `json:"after_cash" bson:"after_cash"`         //账变后彩金
	BeforeBonus  int64  `json:"before_bonus" bson:"before_bonus"`     //账变前奖励金
	AfterBonus   int64  `json:"after_bonus" bson:"after_bonus"`       //账变后奖励金
	CashMingTax  int64  `json:"cash_ming_tax" bson:"cash_ming_tax"`   //彩金明税
	BonusMingTax int64  `json:"bonus_ming_tax" bson:"bonus_ming_tax"` //奖励金明税
	CashAnTax    int64  `json:"cash_an_tax" bson:"cash_an_tax"`       //彩金暗税
	BonusAnTax   int64  `json:"bonus_an_tax" bson:"bonus_an_tax"`     //奖励金暗税
	Robot        bool   `json:"robot" bson:"robot"`                   //人机
	NextBet0     bool   `json:"nextBet0" bson:"next_bet0"`            //位置0是否上一轮下注
	NextBet1     bool   `json:"nextBet1" bson:"next_bet1"`            //位置1是否上一轮下注
	AutoCrash0   bool   `json:"autoCrash0" bson:"auto_crash0"`        //位置0是否自动逃离
	AutoCrash1   bool   `json:"autoCrash1" bson:"auto_crash1"`        //位置1是否自动逃离
	AutoCrashM0  int32  `json:"autoCrashM0" bson:"auto_crash_m0"`     //位置0自动逃离倍数
	AutoCrashM1  int32  `json:"autoCrashM1" bson:"auto_crash_m1"`     //位置1自动逃离倍数

	FBet          float64 `bson:"-"` //下注
	FBet0         float64 `bson:"-"` //位置0下注
	FBet1         float64 `bson:"-"` //位置1下注
	FWin          float64 `bson:"-"` //输赢
	FWin0         float64 `bson:"-"` //位置0输赢
	FWin1         float64 `bson:"-"` //位置1输赢
	FBeforeScore  float64 `bson:"-"` //账变前分数
	FAfterScore   float64 `bson:"-"` //账变后分数
	FBeforeCash   float64 `bson:"-"` //账变前彩金
	FAfterCash    float64 `bson:"-"` //账变后彩金
	FBeforeBonus  float64 `bson:"-"` //账变前奖励金
	FAfterBonus   float64 `bson:"-"` //账变后奖励金
	FCashStock    float64 `bson:"-"` //彩金库存
	FBonusStock   float64 `bson:"-"` //奖励金库存
	FCashMingTax  float64 `bson:"-"` //彩金明税
	FBonusMingTax float64 `bson:"-"` //奖励金明税
	FCashAnTax    float64 `bson:"-"` //彩金暗税
	FBonusAnTax   float64 `bson:"-"` //奖励金暗税
}

type ABDetail struct {
	Joker         uint32          `json:"joker" bson:"joker"`                //Key牌
	JackpotValue  uint32          `json:"jackpotValue" bson:"jackpot_value"` //中奖牌
	ACards        []uint32        `json:"aCards" bson:"a_cards"`             //ANDAR
	BCards        []uint32        `json:"bCards" bson:"b_cards"`             //BAHAR
	Winner        uint32          `json:"winner" bson:"winner"`              //赢家 1:ANDAR 2:BAHAR
	SideWinner    uint32          `json:"sideWinner" bson:"side_winner"`     //赢家2 位置3-10
	Bets          int64           `json:"bets" bson:"bets"`                  //总下注
	PlayerWin     int64           `json:"playerWin" bson:"player_win"`       //玩家赢分
	StrategyId    int             `json:"strategyId" bson:"strategy_id"`     //策略id
	IsRandom      bool            `json:"isRandom" bson:"is_random"`         //是否随机开的
	FinalFactor   float64         `json:"finalFactor" bson:"final_factor"`   //最终系数
	CanWinScore   int64           `json:"canWinScore" bson:"can_win_score"`  //可赢分
	WinnerStr     string          `bson:"-"`                                 //赢家 1:ANDAR 2:BAHAR
	SideWinnerStr string          `bson:"-"`                                 //赢家2 位置3-10
	FBets         float64         `bson:"-"`
	FPlayerWin    float64         `bson:"-"`
	FCanWinScore  float64         `bson:"-"`
	UserDetail    []*ABUserDetail `json:"userDetail" bson:"user_detail"` //用户详细
}

// 玩家下注详情
type ABUserDetail struct {
	Userid        string           `json:"userid" bson:"userid"`                 //用户id
	SeatBets      map[string]int64 `json:"seatBets" bson:"seat_bets"`            //位置下注
	Result        string           `json:"result" bson:"result"`                 //输赢平
	Win           int64            `json:"win" bson:"win"`                       //输赢
	BeforeScore   int64            `json:"beforeScore" bson:"before_score"`      //账变前分数
	AfterScore    int64            `json:"after_score" bson:"after_score"`       //账变后分数
	BeforeCash    int64            `json:"before_cash" bson:"before_cash"`       //账变前彩金
	AfterCash     int64            `json:"after_cash" bson:"after_cash"`         //账变后彩金
	BeforeBonus   int64            `json:"before_bonus" bson:"before_bonus"`     //账变前奖励金
	AfterBonus    int64            `json:"after_bonus" bson:"after_bonus"`       //账变后奖励金
	CashMingTax   int64            `json:"cash_ming_tax" bson:"cash_ming_tax"`   //彩金明税
	BonusMingTax  int64            `json:"bonus_ming_tax" bson:"bonus_ming_tax"` //奖励金明税
	CashAnTax     int64            `json:"cash_an_tax" bson:"cash_an_tax"`       //彩金暗税
	BonusAnTax    int64            `json:"bonus_an_tax" bson:"bonus_an_tax"`     //奖励金暗税
	FBet          float64          `bson:"-"`                                    //下注
	FWin          float64          `bson:"-"`                                    //输赢
	FBeforeScore  float64          `bson:"-"`                                    //账变前分数
	FAfterScore   float64          `bson:"-"`                                    //账变后分数
	FBeforeCash   float64          `bson:"-"`                                    //账变前彩金
	FAfterCash    float64          `bson:"-"`                                    //账变后彩金
	FBeforeBonus  float64          `bson:"-"`                                    //账变前奖励金
	FAfterBonus   float64          `bson:"-"`                                    //账变后奖励金
	FCashStock    float64          `bson:"-"`                                    //彩金库存
	FBonusStock   float64          `bson:"-"`                                    //奖励金库存
	FCashMingTax  float64          `bson:"-"`                                    //彩金明税
	FBonusMingTax float64          `bson:"-"`                                    //奖励金明税
	FCashAnTax    float64          `bson:"-"`                                    //彩金暗税
	FBonusAnTax   float64          `bson:"-"`                                    //奖励金暗税
	ABInfo        *ABSeatBets      `bson:"-"`                                    // 位置下注
}
type ABSeatBets struct {
	Seat1  float64 `bson:"-"`
	Seat2  float64 `bson:"-"`
	Seat3  float64 `bson:"-"`
	Seat4  float64 `bson:"-"`
	Seat5  float64 `bson:"-"`
	Seat6  float64 `bson:"-"`
	Seat7  float64 `bson:"-"`
	Seat8  float64 `bson:"-"`
	Seat9  float64 `bson:"-"`
	Seat10 float64 `bson:"-"`
}

type CPDetail struct {
	JackpotOutput uint32         `json:"jackpotOutput" bson:"jackpot_output"` //奖池产出
	CardType      uint32         `json:"cardType" bson:"card_type"`           //牌型
	Cards         []uint32       `json:"aCards" bson:"a_cards"`               //牌值
	Winner        uint32         `json:"winner" bson:"winner"`                //赢家
	Bets          int64          `json:"bets" bson:"bets"`                    //总下注
	PlayerWin     int64          `json:"playerWin" bson:"player_win"`         //玩家赢分
	StrategyId    int            `json:"strategyId" bson:"strategy_id"`       //策略id
	IsRandom      bool           `json:"isRandom" bson:"is_random"`           //是否随机开的
	FinalFactor   float64        `json:"finalFactor" bson:"final_factor"`     //最终系数
	CanWinScore   int64          `json:"canWinScore" bson:"can_win_score"`    //可赢分
	CardTypeStr   string         `bson:"-"`                                   //牌型位置
	FBets         float64        `bson:"-"`
	FPlayerWin    float64        `bson:"-"`
	FCanWinScore  float64        `bson:"-"`
	UserDetail    []CPUserDetail `json:"userDetail" bson:"user_detail"` //用户详细
}

// 彩票玩家下注详情
type CPUserDetail struct {
	Userid        string           `json:"userid" bson:"userid"`                 //用户id
	SeatBets      map[string]int64 `json:"seatBets" bson:"seat_bets"`            //位置下注
	Result        string           `json:"result" bson:"result"`                 //输赢平
	Win           int64            `json:"win" bson:"win"`                       //输赢
	BeforeScore   int64            `json:"beforeScore" bson:"before_score"`      //账变前分数
	AfterScore    int64            `json:"after_score" bson:"after_score"`       //账变后分数
	BeforeCash    int64            `json:"before_cash" bson:"before_cash"`       //账变前彩金
	AfterCash     int64            `json:"after_cash" bson:"after_cash"`         //账变后彩金
	BeforeBonus   int64            `json:"before_bonus" bson:"before_bonus"`     //账变前奖励金
	AfterBonus    int64            `json:"after_bonus" bson:"after_bonus"`       //账变后奖励金
	CashMingTax   int64            `json:"cash_ming_tax" bson:"cash_ming_tax"`   //彩金明税
	BonusMingTax  int64            `json:"bonus_ming_tax" bson:"bonus_ming_tax"` //奖励金明税
	CashAnTax     int64            `json:"cash_an_tax" bson:"cash_an_tax"`       //彩金暗税
	BonusAnTax    int64            `json:"bonus_an_tax" bson:"bonus_an_tax"`     //奖励金暗税
	FBet          float64          `bson:"-"`                                    //下注
	FWin          float64          `bson:"-"`                                    //输赢
	FBeforeScore  float64          `bson:"-"`                                    //账变前分数
	FAfterScore   float64          `bson:"-"`                                    //账变后分数
	FBeforeCash   float64          `bson:"-"`                                    //账变前彩金
	FAfterCash    float64          `bson:"-"`                                    //账变后彩金
	FBeforeBonus  float64          `bson:"-"`                                    //账变前奖励金
	FAfterBonus   float64          `bson:"-"`                                    //账变后奖励金
	FCashStock    float64          `bson:"-"`                                    //彩金库存
	FBonusStock   float64          `bson:"-"`                                    //奖励金库存
	FCashMingTax  float64          `bson:"-"`                                    //彩金明税
	FBonusMingTax float64          `bson:"-"`                                    //奖励金明税
	FCashAnTax    float64          `bson:"-"`                                    //彩金暗税
	FBonusAnTax   float64          `bson:"-"`                                    //奖励金暗税
	CPInfo        *CPSeatBets      `bson:"-"`                                    // 位置下注
}

type CPSeatBets struct {
	Seat1  float64 `bson:"-"`
	Seat2  float64 `bson:"-"`
	Seat3  float64 `bson:"-"`
	Seat4  float64 `bson:"-"`
	Seat5  float64 `bson:"-"`
	Seat6  float64 `bson:"-"`
	Seat7  float64 `bson:"-"`
	Seat8  float64 `bson:"-"`
	Seat9  float64 `bson:"-"`
	Seat10 float64 `bson:"-"`
}

type RBDetail struct {
	CardType     []uint32       `json:"cardType" bson:"card_type"`        //牌型
	Cards        [][]uint32     `json:"aCards" bson:"a_cards"`            //牌值
	Winner       []uint32       `json:"winner" bson:"winner"`             //赢家
	Bets         int64          `json:"bets" bson:"bets"`                 //总下注
	PlayerWin    int64          `json:"playerWin" bson:"player_win"`      //玩家赢分
	StrategyId   int            `json:"strategyId" bson:"strategy_id"`    //策略id
	CanWinScore  int64          `json:"canWinScore" bson:"can_win_score"` //可赢分
	UserDetail   []RBUserDetail `json:"userDetail" bson:"user_detail"`    //用户详细
	CardType1    string         `bson:"-"`                                //牌型位置
	CardType2    string         `bson:"-"`                                //牌型位置
	WinnerStr    string         `bson:"-"`
	FBets        float64        `bson:"-"`
	FPlayerWin   float64        `bson:"-"`
	FCanWinScore float64        `bson:"-"`
}

// 红黑
type RBUserDetail struct {
	Userid        string           `json:"userid" bson:"userid"`                 //用户id
	SeatBets      map[string]int64 `json:"seatBets" bson:"seat_bets"`            //位置下注
	Result        string           `json:"result" bson:"result"`                 //输赢平
	Win           int64            `json:"win" bson:"win"`                       //输赢
	Observe       bool             `json:"observe" bson:"observe"`               //观察局
	BeforeScore   int64            `json:"beforeScore" bson:"before_score"`      //账变前分数
	AfterScore    int64            `json:"after_score" bson:"after_score"`       //账变后分数
	BeforeCash    int64            `json:"before_cash" bson:"before_cash"`       //账变前彩金
	AfterCash     int64            `json:"after_cash" bson:"after_cash"`         //账变后彩金
	BeforeBonus   int64            `json:"before_bonus" bson:"before_bonus"`     //账变前奖励金
	AfterBonus    int64            `json:"after_bonus" bson:"after_bonus"`       //账变后奖励金
	CashMingTax   int64            `json:"cash_ming_tax" bson:"cash_ming_tax"`   //彩金明税
	BonusMingTax  int64            `json:"bonus_ming_tax" bson:"bonus_ming_tax"` //奖励金明税
	CashAnTax     int64            `json:"cash_an_tax" bson:"cash_an_tax"`       //彩金暗税
	BonusAnTax    int64            `json:"bonus_an_tax" bson:"bonus_an_tax"`     //奖励金暗税
	FBet          float64          `bson:"-"`                                    //下注
	FWin          float64          `bson:"-"`                                    //输赢
	FBeforeScore  float64          `bson:"-"`                                    //账变前分数
	FAfterScore   float64          `bson:"-"`                                    //账变后分数
	FBeforeCash   float64          `bson:"-"`                                    //账变前彩金
	FAfterCash    float64          `bson:"-"`                                    //账变后彩金
	FBeforeBonus  float64          `bson:"-"`                                    //账变前奖励金
	FAfterBonus   float64          `bson:"-"`                                    //账变后奖励金
	FCashStock    float64          `bson:"-"`                                    //彩金库存
	FBonusStock   float64          `bson:"-"`                                    //奖励金库存
	FCashMingTax  float64          `bson:"-"`                                    //彩金明税
	FBonusMingTax float64          `bson:"-"`                                    //奖励金明税
	FCashAnTax    float64          `bson:"-"`                                    //彩金暗税
	FBonusAnTax   float64          `bson:"-"`                                    //奖励金暗税
	ABInfo        *ABSeatBets      `bson:"-"`                                    // 位置下注
}

// 地雷
type MinesDetail struct {
	UserId       string  `json:"user_id" bson:"user_id"`             //用户ID
	Score        int64   `json:"score" bson:"score"`                 //结算
	BeforeScore  int64   `json:"before_score" bson:"before_score"`   //账变前分数
	AfterScore   int64   `json:"after_score" bson:"after_score"`     //账变后分数
	BeforeCash   int64   `json:"before_cash" bson:"before_cash"`     //账变前彩金
	AfterCash    int64   `json:"after_cash" bson:"after_cash"`       //账变后彩金
	BeforeBonus  int64   `json:"before_bonus" bson:"before_bonus"`   //账变前奖励金
	AfterBonus   int64   `json:"after_bonus" bson:"after_bonus"`     //账变后奖励金
	Mines        int32   `json:"mines" bson:"mines"`                 // 埋雷数
	MinesPits    []int32 `json:"minesPits" bson:"mines_pits"`        // 坑位*25:0未踩未知,1已踩无雷,2已踩有雷,3未踩无雷,4未踩有雷
	Step         int32   `json:"step" bson:"step"`                   // 选了几个格子了
	StepPits     []int32 `json:"stepPits" bson:"step_pits"`          // 选的格子
	Multiple     float64 `json:"multiple" bson:"multiple"`           // 返奖倍数
	AutoMines    bool    `json:"autoMines" bson:"auto_mines"`        // 自动对局
	Bets         int64   `json:"bets" bson:"bets"`                   // 总投注
	WinType      int32   `json:"winType" bson:"win_type"`            // 结果:1.赢,2.输,3.强制结算退还下注额
	Force        bool    `json:"force" bson:"force"`                 // 强制结算
	ForceCashOut bool    `json:"forceCashOut" bson:"force_cash_out"` // 玩家一步没走强制结束

	FBets        float64     `bson:"-"` //总投注
	FScore       float64     `bson:"-"` //结算
	FBeforeScore float64     `bson:"-"` //账变前分数
	FAfterScore  float64     `bson:"-"` //账变后分数
	FBeforeCash  float64     `bson:"-"` //账变前彩金
	FAfterCash   float64     `bson:"-"` //账变后彩金
	FMinesPits   []FMinesPit `bson:"-"` //坑位*25
}

type FMinesPit struct {
	Pit     int32 `bson:"-"` // :0未踩未知,1已踩无雷,2已踩有雷,3未踩无雷,4未踩有雷
	Step    int32 `bson:"-"` // 第几步
	StepPit bool  `bson:"-"` // 已踩
}

// 外接下注详情
type NsqBet struct {
	NsqId           string `bson:"_id" json:"_id"`                           //id
	UserId          string `json:"userId" bson:"user_id"`                    // 用户id
	GameId          int32  `json:"gameId" bson:"game_id"`                    // 游戏房间id
	RoundId         string `json:"roundId" bson:"round_id"`                  // 句号
	MerchantOrderNo string `json:"merchantOrderNo" bson:"merchant_order_no"` // 商户订单号
	OrderNo         string `json:"orderNo" bson:"order_no"`                  // 订单号
	Amount          int64  `json:"amount" bson:"amount"`                     // 下注金额
	Ctime           int64  `json:"ctime" bson:"ctime"`                       // 时间
	Cancel          bool   `json:"cancel" bson:"cancel"`                     //
}

// 外接返奖
type NsqReward struct {
	NsqId           string    `bson:"_id" json:"_id"`                           //id
	UserId          string    `json:"userId" bson:"user_id"`                    // 用户id
	GameId          int32     `json:"gameId" bson:"game_id"`                    // 游戏房间id
	RoundId         string    `json:"roundId" bson:"round_id"`                  // 句号
	MerchantOrderNo string    `json:"merchantOrderNo" bson:"merchant_order_no"` // 商户订单号
	OrderNo         string    `json:"orderNo" bson:"order_no"`                  // 订单号
	Amount          int64     `json:"amount" bson:"amount"`                     // 返奖金额
	Ctime           int64     `json:"ctime" bson:"ctime"`                       // 时间
	Cancel          bool      `json:"cancel" bson:"cancel"`                     //
	FAmount         float64   `json:"-" bson:"-"`
	STime           time.Time `json:"-" bson:"-"`
}

// 外接游戏对局详情
type ExternalDetail struct {
	UserId          string      `json:"userId" bson:"user_id"`                    // 用户id
	GameId          int32       `json:"gameId" bson:"game_id"`                    // 游戏房间id
	RoundId         string      `json:"roundId" bson:"round_id"`                  // 局号
	MerchantOrderNo string      `json:"merchantOrderNo" bson:"merchant_order_no"` // 商户订单号
	OrderNo         string      `json:"orderNo" bson:"order_no"`                  // 订单号
	Amount          int64       `json:"amount" bson:"amount"`                     // 下注金额
	Ctime           int64       `json:"ctime" bson:"ctime"`                       // 时间
	GameName        string      `json:"gameName" bson:"-"`
	FAmount         float64     `json:"-" bson:"-"`
	RewardDetail    []NsqReward //返奖详情
}

type MergeDetail struct {
	Id    string `json:"id" bson:"_id"`
	Ctime int64  `json:"ctime" bson:"ctime"`
}

type ExternalReward struct {
	Id     string `json:"id" bson:"_id"`
	UserId string `json:"userId" bson:"user_id"`
	Amount int64  `json:"amount" bson:"amount"`
}

// 赠送可提现金
type GiveWithdrawCash struct {
	Userid string `json:"userid" bson:"userid"` // 玩家id
	Cash   int64  `json:"cash" bson:"cash"`     // 添加的可提现金
}

/*安卓使用结构体*/
// 评分结构体
type AndroidScore struct {
	Bz_Id     int                `json:"bz_id"`
	Code      int                `json:"code"`
	Data      []AndroidScoreInfo `json:"data"`
	Message   string             `json:"message"`
	Time_Cost int                `json:"time_cost"`
}

type AndroidScoreInfo struct {
	Ad_Id         string   `json:"ad_id"`
	Score         int      `json:"score"`
	Score_Reasion []string `json:"score_reasion"`
	Business_Id   int64    `json:"business_id"`
	Status        int      `json:"status"`
}

type AndroidCheckstand struct {
	Bz_Id     int       `json:"bz_id"`
	Code      int       `json:"code"`
	Data      []StatDay `json:"data"`
	Message   string    `json:"message"`
	Time_Cost int       `json:"time_cost"`
}

type StatDay struct {
	Day                      string  `json:"day" bson:"day"`
	PayRequest               int64   `json:"payRequest" bson:"pay_request"` // 充值请求人数
	PayRequestRatio          float64 `bson:"-"`
	DeskUserCount            int     `json:"desk_user_count" bson:"desk_user_count"` // 进入收银台人数
	DeskUserCountRatio       float64 `bson:"-"`
	DeskTimeCount            int     `json:"desk_time_count" bson:"desk_time_count"` // 进入收银台次数
	DeskTimeCountRatio       float64 `bson:"-"`
	PayTotalCount            int     `json:"pay_total_count" bson:"pay_total_count"` // 总次数
	PayTotalCountRatio       float64 `bson:"-"`
	PaySuccessTimeCount      int     `json:"pay_success_time_count" bson:"pay_success_time_count"` // 跳转成功次数
	PaySuccessTimeCountRatio float64 `bson:"-"`
	PaySuccessUserCount      int     `json:"pay_success_user_count" bson:"pay_success_user_count"` // 跳转成功人数
	PaySuccessUserCountRatio float64 `bson:"-"`
}

// lhd游戏策略
type LHDUserStrategy struct {
	XXSC LHDXXSC `json:"xxsc" bson:"xxsc"` // 心想事成
	QSBN LHDQSBN `json:"qsbn" bson:"qsbn"` // 求死不能
}
type LHDXXSC struct {
	DisturbCoolDown int  `json:"disturbCoolDown" bson:"disturb_cool_down"` // 心想事成:干扰cd
	WinLength       int  `json:"winLength" bson:"win_length"`              // 心想事成:连赢局数
	Evo             bool `json:"evo" bson:"evo"`                           // 心想事成:当次生效过策略没有
	TriggerTimes    int  `json:"triggerTimes" bson:"trigger_times"`        // 心想事成:有效触发次数u
	MaxTriggerTimes int  `json:"maxTriggerTimes" bson:"max_trigger_times"` // 心想事成:触发次数u总
}

type LHDQSBN struct {
	TriggerTimes int `json:"triggerTimes" bson:"trigger_times"` // 求死不能:有效触发次数T
}
