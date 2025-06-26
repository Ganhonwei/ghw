package entity

import "time"

// 玩家数据
type PlayerUser struct {
	Userid        string `bson:"_id" json:"userid"`                    // 用户id
	UserType      string `bson:"-"`                                    // 用户类型名称
	Nickname      string `bson:"nickname" json:"nickname"`             // 用户昵称
	RealName      string `bson:"real_name" json:"real_name"`           // 真实姓名
	Photo         string `bson:"photo" json:"photo"`                   // 头像
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
	Money   uint32 `bson:"money" json:"money"`       // 充值总金额(分)
	CashOut int32  `bson:"cash_out" json:"cash_out"` // 提现总金额(分)
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
	FeedBackTimes int32 `bson:"feedback_times" json:"feedback_times"` // 客服回复前发送了多少条
	// FeedBackLog   []ChatLog `bson:"feedback_log" json:"feedback_log"`     // 收到的回复
	// // 分享
	ShareBelow map[string]ShareData `json:"shareBelow" bson:"share_below"` // 分享的下级
	// ShareBetLog   map[string]ShareAmountLog `json:"shareBetLog" bson:"share_bet_log"`    // 分享打码量记录
	ShareSuperior string `json:"shareSuperior" bson:"share_superior"` // 上级
	ShareWithdraw int64  `json:"shareWithdraw" bson:"share_withdraw"` // 可提金额
	ShareTotal    int64  `json:"shareTotal" bson:"share_total"`       // 总的可提现金
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
	PCSwitch        bool    `bson:"point_control_switch"`         //点控开关
	PCFactor        int32   `bson:"point_control_factor"`         //系数设置
	PCScore         int64   `bson:"point_control_score"`          //分数设置
	PCScoreComplete int64   `bson:"point_control_score_complete"` //已经达到的分数
	PCSchedule      float64 `bson:"-"`                            // 完成进度
	// 检测多账号
	IsIpRepeat    bool              `bson:"-"`                                   // IP是否重复
	IsAppIdRepeat bool              `bson:"-"`                                   // 设备号是否重复
	AndroidScore  int               `bson:"-"`                                   // 安卓评分
	IsBlogger     bool              `bson:"-"`                                   // 是否是博主账号 0:否; 1:是
	CrashStrategy CrashUserStrategy `json:"crashStrategy" bson:"crash_strategy"` //crash类策略
	LHDStrategy   LHDUserStrategy   `json:"LHDStrategy" bson:"lhd_strategy"`     //LHD策略
	SevenStrategy LHDUserStrategy   `json:"sevenStrategy" bson:"seven_strategy"` //7up策略
	ABStrategy    ABUserStrategy    `json:"ABStrategy" bson:"ab_strategy"`       //AB策略
	CPStrategy    CPUserStrategy    `json:"CPStrategy"  bson:"cp_strategy"`      //彩票策略
	RBStrategy    RBUserStrategy    `json:"RBStrategy" bson:"rb_strategy"`       //黑红策略
}

// CK 玩家数据
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
	Diamond                int64  `gorm:"column:diamond" json:"diamond"`                               // 钻石(彩金cash)
	GiveDiamond            int64  `json:"giveDiamond" gorm:"column:give_diamond"`                      // 赠送彩金
	VBBank                 int64  `gorm:"column:vbbank" json:"vbbank"`                                 // VB银行
	OutDiamond             int64  `json:"outDiamond" gorm:"column:out_diamond"`                        // 可提现彩金
	ShadowDiamond          int64  `gorm:"column:shadow_diamond" json:"shadow_diamond"`                 // 隐藏钻石(用户看不见的)
	Coin                   int64  `gorm:"column:coin" json:"coin"`                                     // 金币(奖励金bonus)
	Money                  uint32 `gorm:"column:money" json:"money"`                                   // 充值总金额(分)
	CashOut                int32  `gorm:"column:cash_out" json:"cash_out"`                             // 提现总金额(分)
	Profit                 int64  `gorm:"column:profit" json:"profit"`                                 // 利润 (提现+携带)-充值
	CashWater              uint64 `gorm:"column:cash_water" json:"cash_water"`                         // 彩金流水(游戏产生的)
	PrivDiamond            int64  `gorm:"column:priv_diamond" json:"priv_diamond"`                     // 对战房赢的彩金记录
	Status                 int    `gorm:"column:status" json:"status"`                                 // 正常1  锁定2  黑名单3 白名单4
	Robot                  bool   `gorm:"column:robot" json:"robot"`                                   // 是否是机器人
	SimRobot               bool   `gorm:"column:simulation_robot" json:"simulation_robot"`             //模拟机器人
	LoginTimes             int32  `gorm:"column:login_times" json:"login_times"`                       // 登录天数
	OnlineStatus           bool   `gorm:"column:online_status" json:"online_status"`                   // 是否在线
	State                  int    `json:"state" gorm:"column:state"`                                   //玩家状态 1.新手 2.正常 3.平民 4.泡沫
	RegistReward           bool   `json:"registReward" gorm:"column:regist_reward"`                    //是否领取注册奖励
	RegistMode             int32  `json:"registMode" gorm:"column:regist_mode"`                        //注册模式
	Platform               string `json:"platform" gorm:"column:platform"`                             //平台 g:谷歌 o:落地页 pc:PC端
	RegistArea             int    `json:"registArea" gorm:"column:regist_area"`                        //ab测试 0:A 1:B 2:C
	VBOutDiamond           int64  `gorm:"column:vb_out_diamond" json:"vbOutDiamond"`                   // VB可提现分数输分补偿累计
	VBOutDiamondDayTimes   int32  `gorm:"column:vb_out_diamond_day_times" json:"vbOutDiamondDayTimes"` // VB可提现金今日补偿次数
	VBOutDiamondRewardTime int64  `bson:"vb_out_diamond_reward_time" json:"vbOutDiamondRewardTime"`    // 最后领取vb补偿时间戳秒
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
	CommonChannel   uint32 `json:"commonChannel" gorm:"column:common_channel"`                   // 常用充值渠道
	ToggleChannel   int    `json:"toggleChannel" gorm:"column:toggle_channel"`                   // 切换支付渠道
	WithdrawLock    int32  `json:"withdrawLock" gorm:"column:withdraw_lock"`                     // 解锁提现需要充值的金额
	RechargeTarge   []int  `json:"rechargeTarge" gorm:"column:recharge_targe;type:Array(Int32)"` // 充值目标类型
	FirstChargeVal  int32  `json:"firstChargeVal" gorm:"column:first_charge_val"`                // 首次充值金额
	FirstChargeTime int64  `json:"firstChargeTime" gorm:"column:first_charge_time"`              // 首次充值时间
	LastChargeTime  int64  `json:"lastChargeTime" gorm:"column:last_charge_time"`                // 最近充值时间
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
	// CrashStrategy  CrashUserStrategy `json:"crashStrategy" gorm:"column:crash_strategy"`    //crash类策略
	CrashFYZSPlayTimes      int     `json:"-" bson:"play_times"`                                          // 扶摇直上:玩游戏次数(退出游戏才算一次)
	CrashFYZSTiggerTimes    int     `json:"-" bson:"tigger_times"`                                        // 扶摇直上:触发次数
	CrashFYZSEvoTimes       int     `json:"-" bson:"evo_times"`                                           // 扶摇直上:当次玩的局数(与playtimes关联)
	CrashFYZSAllTiggerTimes int     `json:"-" bson:"all_tigger_times"`                                    // 扶摇直上:累计触发次数
	CrashFYZSAllEvoTimes    int     `json:"-" bson:"all_evo_times"`                                       // 扶摇直上:累计玩游戏次数
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

	TpUserTodayChargeNum           int32           `json:"tpUserTodayChargeNum" gorm:"column:tp_user_today_charge_num"`                                          //今日充值次数
	TpUserTodayChargeAmount        int64           `json:"tpUserTodayChargeAmount" gorm:"column:tp_user_today_charge_amount"`                                    //今日充值金额
	TpUserTodayWithdrawNum         int32           `json:"tpUserTodayWithdrawNum" gorm:"column:tp_user_today_withdraw_num"`                                      //今日提现次数
	TpUserTodayWithdrawAmount      int64           `json:"tpUserTodayWithdrawAmount" gorm:"column:tp_user_today_withdraw_amount"`                                //今日提现金额
	TpUserTotalWinOrLoseAmount     int64           `json:"tpUserTotalWinOrLoseAmount" gorm:"column:tp_user_total_win_or_lose_amount"`                            //总输赢额
	TpUserChargeInGameNumOfTrigger int32           `json:"tpUserChargeInGameNumOfTrigger" gorm:"column:tp_user_charge_in_game_num_of_trigger"`                   //局内充值触发数
	TpUserChargeInGameNumOfSuccess int32           `json:"tpUserChargeInGameNumOfSuccess" gorm:"column:tp_user_charge_in_game_num_of_success"`                   //局内充值成功数
	TpUserFollowRateTrigger        map[int32]int32 `json:"tpUserFollowRateTrigger" gorm:"column:tp_user_follow_rate_trigger;type:Map(Int32,Int32)"`              //跟牌率触发数
	TpUserFollowRateSuccess        map[int32]int32 `json:"tpUserFollowRateSuccess" gorm:"column:tp_user_follow_rate_success;type:Map(Int32,Int32)"`              //跟牌率成功数
	TpUserStoryCD                  map[int32]int32 `json:"tpUserStoryCd" gorm:"column:tp_user_story_cd;type:Map(Int32,Int32)"`                                   //剧情模式cd
	TpUserControlStrategyHistory   []int32         `json:"tpUserControlStrategyHistory" gorm:"column:tp_user_control_strategy_history;type:Array(Int32)"`        //控制策略历史
	TpUserGameTime                 int64           `json:"tpUserGameTime" gorm:"column:tp_user_game_time"`                                                       //上次充值后游戏时长
	TpUserTodayGameRound           int32           `json:"tpUserTodayGameRound" gorm:"column:tp_user_today_game_round"`                                          //今日游戏局数
	TpUserTodayControlStrategyNum  map[int32]int32 `json:"tpUserTodayControlStrategyNum" gorm:"column:tp_user_today_control_strategy_num;type:Map(Int32,Int32)"` //今日控制策略触发次数

	// RM相关
	RmRoiLimits         map[string]int32 `json:"rm_roi_limits" gorm:"column:rm_roi_limits;type:Map(String,Int32)"`         // rm roi控制策略生效次数
	RmRoiDayLimits      map[string]int32 `json:"rm_roi_day_limits" gorm:"column:rm_roi_day_limits;type:Map(String,Int32)"` // rm 每个自然日，对应ID策略生效次数
	RmWithout1stDrop    int              `json:"rm_without_1st_drop" gorm:"column:rm_without_1st_drop"`                    // rm 玩家在没有1Life时候首回合弃牌次数
	RmWithout1stNotDrop int              `json:"rm_without_1st_not_drop" gorm:"column:rm_without_1st_not_drop"`            // rm 玩家在没有1Life时候首回合未弃牌次数
}

// crash类游戏策略
type CrashUserStrategy struct {
	FYZS CrashFYZS `json:"fyzs" bson:"fyzs"` // 扶摇直上
	YHWM CrashYHWM `json:"yhwm" bson:"yhwm"` // 欲薅无门
	QSHS CrashQSHS `json:"qshs" bson:"qshs"` // 起死回生
	JCFK CrashJCFK `json:"jcfk" bson:"jcfk"` // 奖池风控
	MXJL CrashMXJL `json:"mxjl" bson:"mxjl"` // 冒险奖励
	RKYH CrashRKYH `json:"rkyh" bson:"rkyh"` // 人狂有祸
}

type CrashFYZS struct {
	PlayTimes      int `json:"playTimes" bson:"play_times"`            // 扶摇直上:玩游戏次数(退出游戏才算一次)
	TiggerTimes    int `json:"tiggerTimes" bson:"tigger_times"`        // 扶摇直上:触发次数
	EvoTimes       int `json:"evoTimes" bson:"evo_times"`              // 扶摇直上:当次玩的局数(与playtimes关联)
	AllTiggerTimes int `json:"allTiggerTimes" bson:"all_tigger_times"` // 扶摇直上:累计触发次数
	AllEvoTimes    int `json:"allEvoTimes" bson:"all_evo_times"`       // 扶摇直上:累计玩游戏次数
}

type CrashYHWM struct {
	MonitorRounds  int     `json:"monitorRounds" bson:"monitor_rounds"`    // 欲薅无门:监控胜局(连胜不能断)
	CrashMulpitles []int32 `json:"crashMulpitles" bson:"crash_mulpitles"`  // 欲薅无门:逃跑倍数
	TiggerTimes    int     `json:"tiggerTimes" bson:"tigger_times"`        // 欲薅无门:触发次数
	WinScore       int64   `json:"winScore" bson:"win_score"`              // 欲薅无门:赢分
	RecyleScore    int64   `json:"recyleScore" bson:"recyle_score"`        // 欲薅无门:回收分数
	AllRecyleScore int64   `json:"allRecyleScore" bson:"all_recyle_score"` // 欲薅无门:累计回收分数
	Tigger         bool    `json:"tigger" bson:"tigger"`                   // 欲薅无门:是否生效中
}

type CrashQSHS struct {
	WinRounds    []int32 `json:"winRounds" bson:"win_rounds"`       // 赢局逃跑倍数
	TriggerTimes int     `json:"triggerTimes" bson:"trigger_times"` // 起死回生：生效次数
	AllInRounds  int     `json:"allInRounds" bson:"all_in_rounds"`  // 起死回生：allin局数
}

type CrashJCFK struct {
	TriggerTimes int   `json:"triggerTimes" bson:"trigger_times"` // 奖池风控:触发jackpot次数
	JackpotVal   int64 `json:"jackpotVal" bson:"jackpot_val"`     // 奖池风控:获得的jackpot值
}

type CrashMXJL struct {
	TriggerTimes int     `json:"triggerTimes" bson:"trigger_times"` // 冒险奖励:触发次数
	Bets         []int64 `json:"bets" bson:"bets"`                  // 冒险奖励:下注
}

type CrashRKYH struct {
	TriggerTimes int     `json:"triggerTimes" bson:"trigger_times"` // 触发rkyh次数
	WinMulpitles []int32 `json:"winMulpitles" bson:"win_mulpitles"` // 前n赢局的局逃跑倍数
}

// lhd游戏策略
type LHDUserStrategy struct {
	RoundBet []int64 `json:"roundBet" bson:"round_bet"` // 每轮下注
	XXSC     LHDXXSC `json:"xxsc" bson:"xxsc"`          // 心想事成
	QSBN     LHDQSBN `json:"qsbn" bson:"qsbn"`          // 求死不能
	LKYH     LHDLKYH `json:"lkyh" bson:"lkyh"`          // 龙狂有祸
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

type LHDLKYH struct {
	TriggerTimes   int   `json:"triggerTimes" bson:"trigger_times"`    // 龙狂有祸:每日有效触发次数T
	HistoryT       []int `json:"historyT" bson:"history_t"`            // 龙狂有祸:历史每日有效触发次数T
	TZ             int   `json:"tz" bson:"tz"`                         // 龙狂有祸:总有效触发次数TZ
	RZ             int   `json:"rz" bson:"rz"`                         // 龙狂有祸:满足策略条件次数
	BSCoolDown     int64 `json:"bsCoolDown" bson:"bs_cool_down"`       // 龙狂有祸:倍杀cd
	BSRounds       int   `json:"bsRounds" bson:"bs_rounds"`            // 龙狂有祸:倍杀局数
	Suppress       int   `json:"suppress" bson:"suppress"`             // 龙狂有祸:压制次数
	NZ             int   `json:"nz" bson:"nz"`                         // 龙狂有祸:倍杀触发局数
	BSTimes        int   `json:"bstimes" bson:"bstimes"`               // 龙狂有祸:倍杀触发次数
	DoubleBetTimes int   `json:"doubleBetTimes" bson:"doubleBetTimes"` // 龙狂有祸:倍投次数
}

type BaseStrategy struct {
	Id         int   `json:"id" bson:"_id"`                // id
	O          int   `json:"o" bson:"o"`                   // 开关
	Weight     int   `json:"weight" bson:"weight"`         // 优先级
	Mutex      []int `json:"mutex" bson:"mutex"`           // 互斥策略
	UserType   []int `json:"userType" bson:"userType"`     // 适用用户类型
	ChargeType []int `json:"chargeType" bson:"chargeType"` // 充值类型
}
type CrashStrategy struct {
	Id   int               `json:"id" bson:"_id"`
	FYZS CRASHFYZSStrategy `json:"fyzs" bson:"fyzs"` // 扶摇直上
	YHWM CRASHYHWMStrategy `json:"yhwm" bson:"yhwm"` // 欲薅无门
	XJQB CRASHXJQBStrategy `json:"xjqb" bson:"xjqb"` // 虚假情报
	QSHS CRASHQSHSStrategy `json:"qshs" bson:"qshs"` // 起死回生
	JCFK CRASHJCFKStrategy `json:"jcfk" bson:"jcfk"` // 奖池风控
	MXJL CRASHMXJLStrategy `json:"mxjl" bson:"mxjl"` // 冒险奖励
	RKYH CRASHRKYHStrategy `json:"rkyh" bson:"rkyh"` // 人狂有祸
}
type CRASHFYZSStrategy struct {
	O          int     `json:"o" bson:"o"`                   // 开关
	UserType   []int   `json:"userType" bson:"userType"`     // 适用用户类型
	ChargeType []int   `json:"chargeType" bson:"chargeType"` // 充值类型
	T          int     `json:"t" bson:"t"`                   // 玩次数上限
	R          int     `json:"r" bson:"r"`                   // 局数上限
	RZ         int     `json:"rz" bson:"rz"`                 // 总生效局数上限
	Q          float64 `json:"q" bson:"q"`                   // 单点爆率修正值
	U          int     `json:"u" bson:"u"`                   // 触发次数上限
	UZ         int     `json:"uz" bson:"uz"`                 // 总触发次数上限
	H          float64 `json:"h" bson:"h"`                   // 不爆倍数上限
	M          int64   `json:"m" bson:"m"`                   // 携带金额上限
}
type CRASHBaseStrategy struct {
	Id         int   `json:"id" bson:"_id"`                // id
	O          int   `json:"o" bson:"o"`                   // 开关
	Weight     int   `json:"weight" bson:"weight"`         // 优先级
	Mutex      []int `json:"mutex" bson:"mutex"`           // 互斥策略
	UserType   []int `json:"userType" bson:"userType"`     // 适用用户类型
	ChargeType []int `json:"chargeType" bson:"chargeType"` // 充值类型
}

type CRASHYHWMStrategy struct {
	*CRASHBaseStrategy // 基础信息
	// O          int     json:"o" bson:"o"                   // 开关
	// UserType   []int   json:"userType" bson:"userType"     // 适用用户类型
	// ChargeType []int   json:"chargeType" bson:"chargeType" // 充值类型
	R int     `json:"r" bson:"r"` // 监控局数（r）
	M float64 `json:"m" bson:"m"` // 逃跑中位数警戒线（m）
	A float64 `json:"a" bson:"a"` // 逃跑平均倍数警戒线（a）
	P int32   `json:"p" bson:"p"` // 瞬爆概率（p）
	X float64 `json:"x" bson:"x"` // 收割倍率（x）
	T int     `json:"t" bson:"t"` // 生效次数上限（t）
}

type CRASHXJQBStrategy struct {
	*BaseStrategy `json:"crashbasestrategy" bson:"crashbasestrategy"` // 基础信息
	P1            float64                                             `json:"p1" bson:"p1"` // 观察返奖率（p1）
	P2            float64                                             `json:"p2" bson:"p2"` // 逃后返奖率（p2）
	B1            float64                                             `json:"b1" bson:"b1"` // 进前录单最高倍数（b1）
	B2            float64                                             `json:"b2" bson:"b2"` // 进前录单第二高倍数（b2）
	B3            float64                                             `json:"b3" bson:"b3"` // 进前录单最低倍数（b3）
}

// 起死回生
type CRASHQSHSStrategy struct {
	*BaseStrategy `json:"crashbasestrategy" bson:"crashbasestrategy"` // 基础信息
	M             float64                                             `json:"m" bson:"m"`   // 援助返奖点m
	D             int64                                               `json:"d" bson:"d"`   // allin特征d
	N             float64                                             `json:"n" bson:"n"`   // 期望返奖点n
	A             float64                                             `json:"a" bson:"a"`   // 期高返奖点o
	Q             float64                                             `json:"q" bson:"q"`   // 单点爆率修正值
	S             float64                                             `json:"s" bson:"s"`   // 不爆倍数上限（s）
	R             int                                                 `json:"r" bson:"r"`   // 赢局次数(r)
	RR            float64                                             `json:"rr" bson:"rr"` // 风控返奖率（rr）
	PF            int64                                               `json:"pf" bson:"pf"` // 风控赢钱上限（pf）
	T             int                                                 `json:"t" bson:"t"`   // 固定触发次数上限（t）
	P             int32                                               `json:"p" bson:"p"`   // 随机拯救率
}

// 奖池风控
type CRASHJCFKStrategy struct {
	*BaseStrategy `json:"crashbasestrategy" bson:"crashbasestrategy"` // 基础信息
	M             float64                                             `json:"m" bson:"m"` // 赢家上限（m）
	N             int64                                               `json:"n" bson:"n"` // 未付费玩家默充金额（n）
	B             float64                                             `json:"b" bson:"b"` // 警戒倍数（b）
	Q             float64                                             `json:"q" bson:"q"` // 单点爆率修正值（q）
	T             int                                                 `json:"t" bson:"t"` // 自由空间（t）
}

// 冒险奖励
type CRASHMXJLStrategy struct {
	*BaseStrategy `json:"crashbasestrategy" bson:"crashbasestrategy"` // 基础信息
	F             float64                                             `json:"f" bson:"f"`   // 奖励点（f)
	R             int                                                 `json:"r" bson:"r"`   // 码局监控（r）
	C             int64                                               `json:"c" bson:"c"`   // 码量监控（c）
	W             int                                                 `json:"w" bson:"w"`   // 博局监控（w）
	M             float64                                             `json:"m" bson:"m"`   // 博倍监控（m）
	N             int64                                               `json:"n" bson:"n"`   // 当局码量（n）
	X             int64                                               `json:"x" bson:"x"`   // 未付费玩家默充金额（x）
	S             float64                                             `json:"s" bson:"s"`   // 不爆倍数上限（s）
	Q             float64                                             `json:"q" bson:"q"`   // 单点爆率修正值（q）
	RR            float64                                             `json:"rr" bson:"rr"` // 风控返奖率（rr）
	PF            int64                                               `json:"pf" bson:"pf"` // 风控赢钱上限（pf）
	T             int                                                 `json:"t" bson:"t"`   // 固定触发次数上限（t）
	P             int32                                               `json:"p" bson:"p"`   // 随机拯救概率（p）
}

// 人狂有祸
type CRASHRKYHStrategy struct {
	*BaseStrategy `json:"crashbasestrategy" bson:"crashbasestrategy"` // 基础信息
	PU            float64                                             `json:"pu" bson:"pu"` // 天罚线（pu）
	P7            float64                                             `json:"p7" bson:"p7"` // 天罚返奖率（p7）
	P1            int32                                               `json:"p1" bson:"p1"` // 瞬爆概率（p1）
	W             int64                                               `json:"w" bson:"w"`   // 赢局监控（w）
	C             int64                                               `json:"c" bson:"c"`   // 肥码上限（c）
	P2            int32                                               `json:"p2" bson:"p2"` // 天罚概率（p2）
	N             float64                                             `json:"n" bson:"n"`   // 罚线调控（n）
	X             int64                                               `json:"x" bson:"x"`   // 未付费玩家默充金额（x）
}

// 充值类型表
type ChargeClassify struct {
	Id     int32   `json:"id" bson:"_id"`
	None   []int64 `json:"none" bson:"none"`
	Normal []int64 `json:"normal" bson:"normal"`
	XR     []int64 `json:"xr" bson:"xr"`
	ZR     []int64 `json:"zr" bson:"zr"`
	DR     []int64 `json:"dr" bson:"dr"`
	CDR    []int64 `json:"cdr" bson:"cdr"`
}

type AviatorStrategy struct {
	Id   int               `json:"id" bson:"_id"`
	FYZS PLANEFYZSStrategy `json:"fyzs" bson:"fyzs"` // 扶摇直上
	YHWM PLANEYHWMStrategy `json:"yhwm" bson:"yhwm"` // 欲薅无门
	XJQB PLANEXJQBStrategy `json:"xjqb" bson:"xjqb"` // 虚假情报
	QSHS PLANEQSHSStrategy `json:"qshs" bson:"qshs"` // 起死回生
	JCFK PLANEJCFKStrategy `json:"jcfk" bson:"jcfk"` // 奖池风控
	MXJL PLANEMXJLStrategy `json:"mxjl" bson:"mxjl"` // 冒险奖励
	RKYH PLANERKYHStrategy `json:"rkyh" bson:"rkyh"` // 人狂有祸
}

type PLANEFYZSStrategy struct {
	*BaseStrategy           // 基础信息
	X             int64     `json:"x" bson:"x"`   // 玩家携带金额上限
	Q             float64   `json:"q" bson:"q"`   // 单点爆率修正值
	C             []int64   `json:"c" bson:"c"`   // 冷却线
	U             []int     `json:"u" bson:"u"`   // 触发次数上限
	UZ            []int     `json:"uz" bson:"uz"` // 总触发次数上限
	H             []float64 `json:"h" bson:"h"`   // 不爆倍数上限
	M             []int64   `json:"m" bson:"m"`   // 携带金额上限
	RR            []float64 `json:"rr" bson:"rr"` // 风控返奖率
	PR            []int64   `json:"pr" bson:"pr"` // 风控赢钱上限
}

type PLANEYHWMStrategy struct {
	*BaseStrategy           // 基础信息
	R             []int     `json:"r" bson:"r"` // 监控局数（r）
	M             []float64 `json:"m" bson:"m"` // 逃跑中位数警戒线（m）
	A             []float64 `json:"a" bson:"a"` // 逃跑平均倍数警戒线（a）
	P             []int32   `json:"p" bson:"p"` // 瞬爆概率（p）
	X             []float64 `json:"x" bson:"x"` // 收割倍率（x）
	T             []int     `json:"t" bson:"t"` // 生效次数上限（t）
}

type PLANEXJQBStrategy struct {
	*BaseStrategy           // 基础信息
	P0            float64   `json:"p0" bson:"p0"` // 录单返奖率（p1）
	P1            []float64 `json:"p1" bson:"p1"` // 观察返奖率（p1）
	P2            []float64 `json:"p2" bson:"p2"` // 逃后返奖率（p2）
}

// 起死回生
type PLANEQSHSStrategy struct {
	*BaseStrategy           // 基础信息
	M             []float64 `json:"m" bson:"m"`   // 援助返奖点m
	D             []int64   `json:"d" bson:"d"`   // allin特征d
	N             []float64 `json:"n" bson:"n"`   // 期望返奖点n
	A             []float64 `json:"a" bson:"a"`   // 期高返奖点o
	Q             float64   `json:"q" bson:"q"`   // 单点爆率修正值
	S             []float64 `json:"s" bson:"s"`   // 不爆倍数上限（s）
	R             []int     `json:"r" bson:"r"`   // 赢局次数(r)
	RR            []float64 `json:"rr" bson:"rr"` // 风控返奖率（rr）
	PF            []int64   `json:"pf" bson:"pf"` // 风控赢钱上限（pf）
	T             []int     `json:"t" bson:"t"`   // 固定触发次数上限（t）
	P             []int32   `json:"p" bson:"p"`   // 随机拯救率
	X             int64     `json:"x" bson:"x"`
}

// 奖池风控
type PLANEJCFKStrategy struct {
	*BaseStrategy           // 基础信息
	M             []float64 `json:"m" bson:"m"` // 赢家上限（m）
	N             int64     `json:"n" bson:"n"` // 未付费玩家默充金额（n）
	B             []float64 `json:"b" bson:"b"` // 警戒倍数（b）
	Q             float64   `json:"q" bson:"q"` // 单点爆率修正值（q）
	T             []int     `json:"t" bson:"t"` // 自由空间（t）
}

// 冒险奖励
type PLANEMXJLStrategy struct {
	*BaseStrategy           // 基础信息
	F             []float64 `json:"f" bson:"f"`   // 奖励点（f)
	R             []int     `json:"r" bson:"r"`   // 码局监控（r）
	C             []int64   `json:"c" bson:"c"`   // 码量监控（c）
	W             []int     `json:"w" bson:"w"`   // 博局监控（w）
	M             []float64 `json:"m" bson:"m"`   // 博倍监控（m）
	N             []int64   `json:"n" bson:"n"`   // 当局码量（n）
	X             int64     `json:"x" bson:"x"`   // 未付费玩家默充金额（x）
	S             []float64 `json:"s" bson:"s"`   // 不爆倍数上限（s）
	Q             float64   `json:"q" bson:"q"`   // 单点爆率修正值（q）
	RR            []float64 `json:"rr" bson:"rr"` // 风控返奖率（rr）
	PF            []int64   `json:"pf" bson:"pf"` // 风控赢钱上限（pf）
	T             []int     `json:"t" bson:"t"`   // 固定触发次数上限（t）
	P             []int32   `json:"p" bson:"p"`   // 随机拯救概率（p）
}

// 人狂有祸
type PLANERKYHStrategy struct {
	*BaseStrategy           // 基础信息
	PU            []float64 `json:"pu" bson:"pu"` // 天罚线（pu）
	P7            []float64 `json:"p7" bson:"p7"` // 天罚返奖率（p7）
	P1            []int32   `json:"p1" bson:"p1"` // 瞬爆概率（p1）
	W             []int64   `json:"w" bson:"w"`   // 赢局监控（w）
	C             []int64   `json:"c" bson:"c"`   // 肥码上限（c）
	P2            []int32   `json:"p2" bson:"p2"` // 天罚概率（p2）
	N             []float64 `json:"n" bson:"n"`   // 罚线调控（n）
	X             int64     `json:"x" bson:"x"`   // 未付费玩家默充金额（x）
}

type LhdStrategy struct {
	Id   int             `json:"id" bson:"_id"`
	XXSC LHDXXSCStrategy `json:"xxsc" bson:"xxsc"` // 心想事成
	QSBN LHDQSBNStrategy `json:"qsbn" bson:"qsbn"` // 求死不能
	LKYH LHDLKYHStrategy `json:"lkyh" bson:"lkyh"` // 龙狂有祸
}
type LHDXXSCStrategy struct {
	*BaseStrategy         // 基础信息
	M             []int64 `json:"m" bson:"m"`   //援助返奖点（m）
	A             []int   `json:"a" bson:"a"`   //怀疑警戒（a）
	D             []int64 `json:"d" bson:"d"`   //allin特征（d）
	B             []int   `json:"b" bson:"b"`   //扰动冷却（b）
	S             []int32 `json:"s" bson:"s"`   //赢倍上限（s）
	U             []int   `json:"u" bson:"u"`   //有效触发次数上限（u）
	UZ            []int   `json:"uz" bson:"uz"` //总触发次数上限（u总）
}

type LHDQSBNStrategy struct {
	*BaseStrategy           // 基础信息
	M             []float64 `json:"m" bson:"m"`   //援助返奖点（m）
	RR            []float64 `json:"rr" bson:"rr"` //风控返奖率（rr）
	D             []int64   `json:"d" bson:"d"`   //allin特征（d）
	PR            []int64   `json:"pr" bson:"pr"` //风控赢钱上限（pr）
	T             []int     `json:"t" bson:"t"`   //固定触发次数上限（t）
	P             []int32   `json:"p" bson:"p"`   //随机拯救概率（p）
}

type LHDLKYHStrategy struct {
	*BaseStrategy           // 基础信息
	X             int64     //未付费玩家默充金额（x)
	PU            []float64 //天罚线（pu）
	PP            []int64   //天罚利（pp）
	P3            []int32   //压制概率（p3）
	P4            []int32   //倍杀概率（p4）
	N             []int     //倍杀轮回（n）
	C             []int64   //倍杀冷却（c
	T             []int     //单日倍杀次数上限（t日）
	TZ            []int     //总倍杀次数上限（tz总
}

type SevenStrategy struct {
	Id   int             `json:"id" bson:"_id"`
	XXSC LHDXXSCStrategy `json:"xxsc" bson:"xxsc"` // 心想事成
	QSBN LHDQSBNStrategy `json:"qsbn" bson:"qsbn"` // 求死不能
	LKYH LHDLKYHStrategy `json:"lkyh" bson:"lkyh"` // 龙狂有祸
}

type ShareData struct {
	BetAmount int64  `json:"betAmount" bson:"bet_amount"` //总打码量
	UserId    string `json:"userId" bson:"user_id"`
	Name      string `json:"name" bson:"name"`            //名称
	LastLogin int64  `json:"lastLogin" bson:"last_login"` //上次登录
	Recharge  int64  `json:"recharge" bson:"recharge"`    //充值金额
	Receive   bool   `json:"receive" bson:"receive"`      //是否领取奖励
}

// ab游戏策略
type ABStrategy struct {
	Id   int            `json:"id" bson:"_id"`
	ANQS ABANQSStrategy `json:"anqs" bson:"anqs"` // 安能求死
	ARTY ABARTYStrategy `json:"arty" bson:"arty"` // 安然躺赢
}

type ABANQSStrategy struct {
	*BaseStrategy           // 基础信息
	X             int64     `json:"x" bson:"x"`   // 未付费玩家默充金额（x）
	D             []int64   `json:"d" bson:"d"`   // allin特征（d）
	M             []float64 `json:"m" bson:"m"`   // 援助返奖点（m）
	RR            []float64 `json:"rr" bson:"rr"` // 风控返奖率（rr）
	PR            []int64   `json:"pr" bson:"pr"` // 风控赢钱上限（pr）
	T             []int     `json:"t" bson:"t"`   // 固定触发次数上限（t）
	PS            []int32   `json:"ps" bson:"ps"` // 随机拯救概率（ps）
}

type ABARTYStrategy struct {
	*BaseStrategy           // 基础信息
	X             int64     `json:"x" bson:"x"`   //未付费玩家默充金额（x）
	M             []int64   `json:"m" bson:"m"`   //援助返奖点（m）
	RR            []float64 `json:"rr" bson:"rr"` //风控返奖率（rr）
	PR            []int64   `json:"pr" bson:"pr"` //风控赢钱上限（pr）
	C             []int64   `json:"c" bson:"c"`   //冷却线（c）
	U             []int     `json:"u" bson:"u"`   //有效触发次数上限（u）
	UZ            []int     `json:"uz" bson:"uz"` //总次数上限（u总）
}

type ABUserStrategy struct {
	RoundBet []int64 `json:"roundBet" bson:"round_bet"` // 每轮下注
	ANQS     ABANQS  `json:"anqs" bson:"anqs"`          // 安能求死
	ARTY     ABARTY  `json:"arty" bson:"arty"`          // 安然躺赢
}

type ABANQS struct {
	TriggerTimes int `json:"triggerTimes" bson:"trigger_times"` // 安能求死:有效触发次数T
}

type ABARTY struct {
	U        int   `json:"u" bson:"u"`                // 安然躺赢:有效触发次数T
	UZ       int   `json:"uz" bson:"uz"`              // 安然躺赢:触发次数T
	EvoTimes bool  `json:"evoTimes" bson:"evoTimes"`  // 安然躺赢:当次触发过策略没有
	WinScore int64 `json:"winScore" bson:"win_score"` // 安然躺赢:净赢分
}

type CPStrategy struct {
	Id   int            `json:"id" bson:"_id"`
	LWJY CPLWJYStrategy `json:"lwjy" bson:"lwjy"` // 来玩就赢
	LYQN CPLYQNStrategy `json:"lyqn" bson:"lyqn"` // 来易去难
}

type CPLWJYStrategy struct {
	*BaseStrategy           // 基础信息
	X             int64     `json:"x" bson:"x"`   //未付费玩家默充金额（x）
	M             []int64   `json:"m" bson:"m"`   //援助返奖点（m）
	RR            []float64 `json:"rr" bson:"rr"` //风控返奖率（rr）
	PR            []int64   `json:"pr" bson:"pr"` //风控赢钱上限（pr）
	C             []int64   `json:"c" bson:"c"`   //冷却线（c）
	U             []int     `json:"u" bson:"u"`   //有效触发次数上限（u）
	UZ            []int     `json:"uz" bson:"uz"` //总次数上限（u总）
}

type CPLYQNStrategy struct {
	*BaseStrategy           // 基础信息
	X             int64     `json:"x" bson:"x"`   // 未付费玩家默充金额（x）
	D             []int64   `json:"d" bson:"d"`   // allin特征（d）
	M             []float64 `json:"m" bson:"m"`   // 援助返奖点（m）
	RR            []float64 `json:"rr" bson:"rr"` // 风控返奖率（rr）
	PR            []int64   `json:"pr" bson:"pr"` // 风控赢钱上限（pr）
	T             []int     `json:"t" bson:"t"`   // 固定触发次数上限（t）
	PS            []int32   `json:"ps" bson:"ps"` // 随机拯救概率（ps）
}

// cp游戏策略
type CPUserStrategy struct {
	RoundBet []int64 `json:"roundBet" bson:"round_bet"` // 每轮下注
	LWJY     CPLWJY  `json:"lwjy" bson:"lwjy"`          // 来玩就赢
	LYQN     CPLYQN  `json:"lyqn" bson:"lyqn"`          // 来易去难
}

type CPLYQN struct {
	TriggerTimes int `json:"triggerTimes" bson:"trigger_times"` // 来易去难:有效触发次数T
}

type CPLWJY struct {
	U        int   `json:"u" bson:"u"`                // 来玩就赢:有效触发次数T
	UZ       int   `json:"uz" bson:"uz"`              // 来玩就赢:触发次数T
	EvoTimes bool  `json:"evoTimes" bson:"evoTimes"`  // 来玩就赢:当次触发过策略没有
	WinScore int64 `json:"winScore" bson:"win_score"` // 来玩就赢:净赢分
}

/*
RB 策略
*/
type RBStrategy struct {
	Id   int            `json:"id" bson:"_id"`
	HYDT RBHYDTStrategy `json:"hydt" bson:"hydt"` // 红运当头
	JCFS RBJCFSStrategy `json:"jcfs" bson:"jcfs"` // 决出逢生
}

type RBHYDTStrategy struct {
	*BaseStrategy           // 基础信息
	X             int64     `json:"x" bson:"x"`   //未付费玩家默充金额（x）
	M             []int64   `json:"m" bson:"m"`   //援助返奖点（m）
	RR            []float64 `json:"rr" bson:"rr"` //风控返奖率（rr）
	PR            []int64   `json:"pr" bson:"pr"` //风控赢钱上限（pr）
	C             []int64   `json:"c" bson:"c"`   //冷却线（c）
	U             []int     `json:"u" bson:"u"`   //有效触发次数上限（u）
	UZ            []int     `json:"uz" bson:"uz"` //总次数上限（u总）
}

type RBJCFSStrategy struct {
	*BaseStrategy           // 基础信息
	X             int64     `json:"x" bson:"x"`   // 未付费玩家默充金额（x）
	D             []int64   `json:"d" bson:"d"`   // allin特征（d）
	M             []float64 `json:"m" bson:"m"`   // 援助返奖点（m）
	RR            []float64 `json:"rr" bson:"rr"` // 风控返奖率（rr）
	PR            []int64   `json:"pr" bson:"pr"` // 风控赢钱上限（pr）
	T             []int     `json:"t" bson:"t"`   // 固定触发次数上限（t）
	PS            []int32   `json:"ps" bson:"ps"` // 随机拯救概率（ps）
}

// rb游戏策略
type RBUserStrategy struct {
	RoundBet []int64 `json:"roundBet" bson:"round_bet"` // 每轮下注
	HYDT     RBHYDT  `json:"hydt" bson:"hydt"`          // 红运当头
	JCFS     RBJCFS  `json:"jcfs" bson:"jcfs"`          // 绝处逢生
}

type RBJCFS struct {
	TriggerTimes int `json:"triggerTimes" bson:"trigger_times"` // 绝处逢生:有效触发次数T
}

type RBHYDT struct {
	U        int   `json:"u" bson:"u"`                // 红运当头:有效触发次数T
	UZ       int   `json:"uz" bson:"uz"`              // 红运当头:触发次数T
	EvoTimes bool  `json:"evoTimes" bson:"evoTimes"`  // 红运当头:当次触发过策略没有
	WinScore int64 `json:"winScore" bson:"win_score"` // 红运当头:净赢分
}

// 玩家游戏局数
type UserGames struct {
	Id     int32  `bson:"-"` // 游戏类型ID
	Name   string `bson:"-"` // 游戏类型名称
	Number int32  `bson:"-"` // 游戏局数
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
type Shop struct {
	Id      string    `bson:"_id"`    //购买ID
	Status  int       `bson:"status"` //物品状态,1=热卖
	Propid  int       `bson:"propid"` //兑换的物品,1=钻石
	Payway  int       `bson:"payway"` //支付方式,1=RMB
	Number  uint32    `bson:"number"` //兑换的数量
	FNumber float64   `bson:"-"`
	Give    uint32    `bson:"give"` //赠送数量
	FGive   float64   `bson:"-"`
	Price   uint32    `bson:"price"` //支付价格(单位元)
	FPrice  float64   `bson:"-"`
	Name    string    `bson:"name"`  //物品名字
	Info    string    `bson:"info"`  //物品信息
	Del     int       `bson:"del"`   //是否移除
	Ctime   time.Time `bson:"ctime"` //创建时间
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
	FPlayerFactor                      float64        `bson:"-" json:"-"`                                                                                 //玩家系数
	ChangeCardTypeName                 string         `bson:"-"`                                                                                          //换牌局类型 0：没有 1：开局换牌 2：局中换牌
	ControlTypeName                    string         `bson:"-"`                                                                                          //控制方式 1:当前赢分;2：房间系数
	FWinScore                          float64        `bson:"-"`                                                                                          //当前赢分(开局)
	FWinScore1                         float64        `bson:"-"`                                                                                          //当局赢分(局中)
	FWinScore2                         float64        `bson:"-"`                                                                                          //当前赢分(局中)
	FChargeMoney                       float64        `bson:"-"`                                                                                          //充值金额
	FPlayerWin                         float64        `bson:"-"`                                                                                          //玩家赢分
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
}

type DetailCK struct {
	Ver                                int64     `gorm:"column:ver" json:"ver"`                                                                             // 插入时间戳
	WaterId                            string    `gorm:"column:id" json:"_id"`                                                                              //局号
	UserId                             string    `gorm:"column:userid" json:"userid"`                                                                       //参与玩家ID, 拆players
	BeginTime                          int64     `gorm:"column:begin_time" json:"begin_time"`                                                               //开始时间
	EndTime                            int64     `gorm:"column:end_time" json:"end_time"`                                                                   //结束时间
	Begin                              time.Time `gorm:"column:begin" json:"begin"`                                                                         //开始时间
	End                                time.Time `gorm:"column:end" json:"end"`                                                                             //结束时间
	Gtype                              int32     `gorm:"column:gtype" json:"gtype"`                                                                         //所在游戏
	Rtype                              int32     `gorm:"column:rtype" json:"rtype"`                                                                         //房间类型: 0自由,1私人,2百人
	Gmode                              int32     `gorm:"column:gmode" json:"gmode"`                                                                         //私人房模式: 0真金,1娱乐模式
	RoomId                             string    `gorm:"column:room_id" json:"room_id"`                                                                     //房间ID
	DeskId                             string    `gorm:"column:desk_id" json:"desk_id"`                                                                     //桌子ID
	Players                            string    `gorm:"column:players" json:"players"`                                                                     //参与玩家ID
	CardTypeId                         int       `gorm:"column:card_type_id" json:"card_type_id"`                                                           //牌型ID
	ChangeCardType                     int       `gorm:"column:change_card_type" json:"change_card_type"`                                                   //换牌局类型(开局换，局中换)
	ControlType                        int       `gorm:"column:control_type" json:"control_type"`                                                           //控制方式(开局)
	WinScore                           int       `gorm:"column:win_score" json:"win_score"`                                                                 //当前赢分(开局)
	PlayerFactor                       int       `gorm:"column:player_factor" json:"player_factor"`                                                         //玩家系数(开局)
	WinScore1                          int       `gorm:"column:win_score_1" json:"win_score_1"`                                                             //当局赢分(局中)
	WinScore2                          int       `gorm:"column:win_score_2" json:"win_score_2"`                                                             //当前赢分(局中)
	ChargeMoney                        int       `gorm:"column:charge_money" json:"charge_money"`                                                           //充值金额(开局，局中)
	IsStrategy                         bool      `gorm:"column:is_strategy" json:"is_strategy"`                                                             //是否是策略局
	StrategyType                       int       `gorm:"column:strategy_type" json:"strategy_type"`                                                         //策略局类型
	IsCharge                           bool      `gorm:"column:is_charge" json:"is_charge"`                                                                 //是否充值
	RealGame                           bool      `gorm:"column:real_game" json:"real_game"`                                                                 //是否真人对局
	NonDirty                           []uint32  `gorm:"column:non_dirty;type:Array(UInt32)" json:"non_dirty"`                                              //没有污染的人机座位
	ShowCards                          string    `gorm:"column:show_cards" json:"show_cards"`                                                               //亮牌
	IsMustLose                         bool      `gorm:"column:is_must_lose" json:"is_must_lose"`                                                           //是否是必输局
	MustLoseScore                      int       `gorm:"column:must_lose_score" json:"must_lose_score"`                                                     //必输分数线
	CoreRobotSeat                      uint32    `gorm:"column:core_robot_seat" json:"core_robot_seat"`                                                     //核心人机位置
	IsUpCardPool                       bool      `gorm:"column:is_up_card_pool" json:"is_up_card_pool"`                                                     //是否升档牌库
	CardPoolId                         int       `gorm:"column:card_pool_id" json:"card_pool_id"`                                                           //牌库编号
	DealType                           int       `gorm:"column:deal_type" json:"deal_type"`                                                                 //发牌方式
	IsTriggerWinScoreLimit             bool      `gorm:"column:is_trigger_win_score_limit" json:"is_trigger_win_score_limit"`                               //是否触发赢分限制
	IsTriggerMustLose                  bool      `gorm:"column:is_trigger_must_lose" json:"is_trigger_must_lose"`                                           //是否触发必输控制
	IsTriggerMustLoseDrawCard          bool      `gorm:"column:is_trigger_must_lose_draw_card" json:"is_trigger_must_lose_draw_card"`                       //是否触发必输控制摸牌限制
	IsTriggerMustLoseLastCard          bool      `gorm:"column:is_trigger_must_lose_last_card" json:"is_trigger_must_lose_last_card"`                       //是否触发必输控制上家出牌限制
	IsTriggerMustLoseCoreRobotGoodCard bool      `gorm:"column:is_trigger_must_lose_core_robot_good_card" json:"is_trigger_must_lose_core_robot_good_card"` //是否触发必输控制核心人机摸好牌
	RMGameOverReason                   int       `gorm:"column:rm_game_over_reason" json:"rm_game_over_reason"`                                             //rm游戏结束原因 (1.天胡,2.自摸胡牌,3.吃牌胡牌,4.对手弃牌)
	// 同步时计算字段 打码量 输赢分 ====================================================================================================
	// Winner         string `gorm:"column:winner" json:"winner"`                     // 赢家userid
	// WinnerRobot    bool   `gorm:"column:winner_robot" json:"winner_robot"`         // 赢家是人机
	WinType      int8  `gorm:"column:win_type" json:"win_type"`             // 1.赢 2.输 3.平 4.观察局
	Robot        bool  `gorm:"column:robot" json:"robot"`                   // 是否人机
	BetAmount    int64 `gorm:"column:bet_amount" json:"bet_amount"`         // 本局打码量/总投注
	SettleScore  int64 `gorm:"column:settle_score" json:"settle_score"`     // 结算分
	Score        int64 `gorm:"column:score" json:"score"`                   // 对局结算分(减下注减去税)
	ScoreUntax   int64 `gorm:"column:score_untax" json:"score_untax"`       // 对局结算分(减下注未减税)
	BeforeScore  int64 `json:"before_score" gorm:"column:before_score"`     // 账变前分数
	AfterScore   int64 `json:"after_score" gorm:"column:after_score"`       // 账变后分数
	CashMingTax  int64 `json:"cash_ming_tax" gorm:"column:cash_ming_tax"`   // 彩金明税
	BonusMingTax int64 `json:"bonus_ming_tax" gorm:"column:bonus_ming_tax"` // 奖励金明税
	CashAnTax    int64 `json:"cash_an_tax" gorm:"column:cash_an_tax"`       // 彩金暗税
	BonusAnTax   int64 `json:"bonus_an_tax" gorm:"column:bonus_an_tax"`     // 奖励金暗税

	// TP详情展开 ====================================================================================================
	// TPDetail []*TPDetail //TP详情
	TpSeatId           uint32   `json:"-" gorm:"column:tp_seat_id"`                      //座位ID
	TpUserId           string   `json:"-" gorm:"column:tp_user_id"`                      //用户ID
	TpCards            []uint32 `json:"-" gorm:"column:tp_cards;type:Array(UInt32)"`     //牌型
	TpScore            int64    `json:"-" gorm:"column:tp_score"`                        //结算
	TpBeforeScore      int64    `json:"-" gorm:"column:tp_before_score"`                 //账变前分数
	TpAfterScore       int64    `json:"-" gorm:"column:tp_after_score"`                  //账变后分数
	TpBeforeCash       int64    `json:"-" gorm:"column:tp_before_cash"`                  //账变前彩金
	TpAfterCash        int64    `json:"-" gorm:"column:tp_after_cash"`                   //账变后彩金
	TpBeforeBonus      int64    `json:"-" gorm:"column:tp_before_bonus"`                 //账变前奖励金
	TpAfterBonus       int64    `json:"-" gorm:"column:tp_after_bonus"`                  //账变后奖励金
	TpBet              int64    `json:"-" gorm:"column:tp_bet"`                          //总投注
	TpBottom           int64    `json:"-" gorm:"column:tp_bottom"`                       //底注
	TpCashStock        int64    `json:"-" gorm:"column:tp_cash_stock"`                   //彩金库存
	TpBonusStock       int64    `json:"-" gorm:"column:tp_bonus_stock"`                  //奖励金库存
	TpCashMingTax      int64    `json:"-" gorm:"column:tp_cash_ming_tax"`                //彩金明税
	TpBonusMingTax     int64    `json:"-" gorm:"column:tp_bonus_ming_tax"`               //奖励金明税
	TpCashAnTax        int64    `json:"-" gorm:"column:tp_cash_an_tax"`                  //彩金暗税
	TpBonusAnTax       int64    `json:"-" gorm:"column:tp_bonus_an_tax"`                 //奖励金暗税
	TpTurn             []int32  `json:"-" gorm:"column:tp_turn;type:Array(Int32)"`       //轮次详情展开,轮次ID
	TpOperation        []string `json:"-" gorm:"column:tp_operation;type:Array(String)"` //轮次详情展开,操作
	TpNewBieProbeId    int32    `json:"-" gorm:"column:tp_new_bie_probe_id"`             // 新手试探期/稳定局策略id
	TpNewBieProbeRound int32    `json:"-" gorm:"column:tp_new_bie_probe_round"`          // 玩家新手试探期/稳定局策略生效局数

	// JOKER详情展开 ====================================================================================================
	// JOKERDetail []*JOKERDetail //JOKER详情
	JokerSeatId       uint32   `json:"-" gorm:"column:joker_seat_id"`                      //座位ID
	JokerUserId       string   `json:"-" gorm:"column:joker_user_id"`                      //用户ID
	JokerCards        []uint32 `json:"-" gorm:"column:joker_cards;type:Array(UInt32)"`     //牌型
	JokerScore        int64    `json:"-" gorm:"column:joker_score"`                        //结算
	JokerBeforeScore  int64    `json:"-" gorm:"column:joker_before_score"`                 //账变前分数
	JokerAfterScore   int64    `json:"-" gorm:"column:joker_after_score"`                  //账变后分数
	JokerBeforeCash   int64    `json:"-" gorm:"column:joker_before_cash"`                  //账变前彩金
	JokerAfterCash    int64    `json:"-" gorm:"column:joker_after_cash"`                   //账变后彩金
	JokerBeforeBonus  int64    `json:"-" gorm:"column:joker_before_bonus"`                 //账变前奖励金
	JokerAfterBonus   int64    `json:"-" gorm:"column:joker_after_bonus"`                  //账变后奖励金
	JokerBet          int64    `json:"-" gorm:"column:joker_bet"`                          //总投注
	JokerBottom       int64    `json:"-" gorm:"column:joker_bottom"`                       //底注
	JokerCashStock    int64    `json:"-" gorm:"column:joker_cash_stock"`                   //彩金库存
	JokerBonusStock   int64    `json:"-" gorm:"column:joker_bonus_stock"`                  //奖励金库存
	JokerCashMingTax  int64    `json:"-" gorm:"column:joker_cash_ming_tax"`                //彩金明税
	JokerBonusMingTax int64    `json:"-" gorm:"column:joker_bonus_ming_tax"`               //奖励金明税
	JokerCashAnTax    int64    `json:"-" gorm:"column:joker_cash_an_tax"`                  //彩金暗税
	JokerBonusAnTax   int64    `json:"-" gorm:"column:joker_bonus_an_tax"`                 //奖励金暗税
	JokerTurn         []int32  `json:"-" gorm:"column:joker_turn;type:Array(Int32)"`       //轮次详情展开,轮次ID
	JokerOperation    []string `json:"-" gorm:"column:joker_operation;type:Array(String)"` //轮次详情展开,操作

	// Ak47详情展开 ====================================================================================================
	// AK47Detail []*AK47Detail //AK47详情
	Ak47SeatId       uint32   `json:"-" gorm:"column:ak47_seat_id"`                         //座位ID
	Ak47UserId       string   `json:"-" gorm:"column:ak47_user_id"`                         //用户ID
	Ak47Cards        []uint32 `json:"-" gorm:"column:ak47_cards;type:Array(UInt32)"`        //牌型
	Ak47ChangeCards  []uint32 `json:"-" gorm:"column:ak47_change_cards;type:Array(UInt32)"` //变牌
	Ak47Score        int64    `json:"-" gorm:"column:ak47_score"`                           //结算
	Ak47BeforeScore  int64    `json:"-" gorm:"column:ak47_before_score"`                    //账变前分数
	Ak47AfterScore   int64    `json:"-" gorm:"column:ak47_after_score"`                     //账变后分数
	Ak47BeforeCash   int64    `json:"-" gorm:"column:ak47_before_cash"`                     //账变前彩金
	Ak47AfterCash    int64    `json:"-" gorm:"column:ak47_after_cash"`                      //账变后彩金
	Ak47BeforeBonus  int64    `json:"-" gorm:"column:ak47_before_bonus"`                    //账变前奖励金
	Ak47AfterBonus   int64    `json:"-" gorm:"column:ak47_after_bonus"`                     //账变后奖励金
	Ak47Bet          int64    `json:"-" gorm:"column:ak47_bet"`                             //总投注
	Ak47Bottom       int64    `json:"-" gorm:"column:ak47_bottom"`                          //底注
	Ak47CashStock    int64    `json:"-" gorm:"column:ak47_cash_stock"`                      //彩金库存
	Ak47BonusStock   int64    `json:"-" gorm:"column:ak47_bonus_stock"`                     //奖励金库存
	Ak47CashMingTax  int64    `json:"-" gorm:"column:ak47_cash_ming_tax"`                   //彩金明税
	Ak47BonusMingTax int64    `json:"-" gorm:"column:ak47_bonus_ming_tax"`                  //奖励金明税
	Ak47CashAnTax    int64    `json:"-" gorm:"column:ak47_cash_an_tax"`                     //彩金暗税
	Ak47BonusAnTax   int64    `json:"-" gorm:"column:ak47_bonus_an_tax"`                    //奖励金暗税
	Ak47Turn         []int32  `json:"-" gorm:"column:ak47_turn;type:Array(Int32)"`          //轮次详情展开,轮次ID
	Ak47Operation    []string `json:"-" gorm:"column:ak47_operation;type:Array(String)"`    //轮次详情展开,操作

	// RM详情展开 ====================================================================================================
	// RMDetail []*RMDetail //RM详情
	RmSeatId       uint32     `json:"-" gorm:"column:rm_seat_id"`                               //座位ID
	RmUserId       string     `json:"-" gorm:"column:rm_user_id"`                               //用户ID
	RmResult       string     `json:"-" gorm:"column:rm_result"`                                //结果
	RmScore        int64      `json:"-" gorm:"column:rm_score"`                                 //结算
	RmBeforeScore  int64      `json:"-" gorm:"column:rm_before_score"`                          //账变前分数
	RmAfterScore   int64      `json:"-" gorm:"column:rm_after_score"`                           //账变后分数
	RmBeforeCash   int64      `json:"-" gorm:"column:rm_before_cash"`                           //账变前彩金
	RmAfterCash    int64      `json:"-" gorm:"column:rm_after_cash"`                            //账变后彩金
	RmBeforeBonus  int64      `json:"-" gorm:"column:rm_before_bonus"`                          //账变前奖励金
	RmAfterBonus   int64      `json:"-" gorm:"column:rm_after_bonus"`                           //账变后奖励金
	RmFinalCards   [][]uint32 `json:"-" gorm:"column:rm_final_cards;type:Array(Array(UInt32))"` //最终牌型
	RmInitCards    []uint32   `json:"-" gorm:"column:rm_init_cards;type:Array(UInt32)"`         //初始牌型
	RmWildCard     uint32     `json:"-" gorm:"column:rm_wild_card"`                             //万能牌
	RmMoCards      []uint32   `json:"-" gorm:"column:rm_mo_cards;type:Array(UInt32)"`           //每轮摸到的牌
	RmChuCards     []uint32   `json:"-" gorm:"column:rm_chu_cards;type:Array(UInt32)"`          //每轮出的牌
	RmCashStock    int64      `json:"-" gorm:"column:rm_cash_stock"`                            //彩金库存
	RmBonusStock   int64      `json:"-" gorm:"column:rm_bonus_stock"`                           //奖励金库存
	RmCashMingTax  int64      `json:"-" gorm:"column:rm_cash_ming_tax"`                         //彩金明税
	RmBonusMingTax int64      `json:"-" gorm:"column:rm_bonus_ming_tax"`                        //奖励金明税
	RmCashAnTax    int64      `json:"-" gorm:"column:rm_cash_an_tax"`                           //彩金暗税
	RmBonusAnTax   int64      `json:"-" gorm:"column:rm_bonus_an_tax"`                          //奖励金暗税

	// RMControl 展开 ====================================================================================================
	// RMControl                *RMControl `gorm:"column:rm_control" json:"rm_control"`                      //rm 控制详情
	RmcOk                    bool   `gorm:"column:rmc_ok" json:"rmc_ok"`                  // 有rm控制详情
	RmcCtype                 int32  `json:"-" gorm:"column:rmc_ctype"`                    // 0.随机,1.库存控制,2.ROI策略
	RmcRoiId                 string `json:"-" gorm:"column:rmc_roi_id"`                   // ctype=2 roi策略id
	RmcDropId                string `json:"-" gorm:"column:rmc_drop_id"`                  // 弃牌策略id
	RmcControlEffect         bool   `json:"-" gorm:"column:rmc_control_effect"`           // rm进入控制模式 控制概率生效
	RmcHierarchy             int32  `json:"-" gorm:"column:rmc_hierarchy"`                // 进入控制档位
	RmcDrawNumPlayer         int    `json:"-" gorm:"column:rmc_draw_num_player"`          // 玩家抽牌张数
	RmcDrawNumRobot          int    `json:"-" gorm:"column:rmc_draw_num_robot"`           // 人机抽牌张数
	RmcNotDrawRobot1st       bool   `json:"-" gorm:"column:rmc_not_draw_robot1st"`        // 人机保顺金
	RmcNotDrawPlayer1st      bool   `json:"-" gorm:"column:rmc_not_draw_player1st"`       // 玩家保顺金
	RmcControlRobotDraw      bool   `json:"-" gorm:"column:rmc_control_robot_draw"`       // 人机进入干预
	RmcControlRobotDrawRound int32  `json:"-" gorm:"column:rmc_control_robot_draw_round"` // 人机进入干预回合

	// LH详情展开 ====================================================================================================
	// LHDetail *LHDetail //LH详情
	LhdWinner      uint32 `json:"-" gorm:"column:lhd_winner"`       //赢家
	LhdDragonValue string `json:"-" gorm:"column:lhd_dragon_value"` //龙牌值
	LhdTigerValue  string `json:"-" gorm:"column:lhd_tiger_value"`  //虎牌值
	LhdBets        int64  `json:"-" gorm:"column:lhd_bets"`         //总下注
	LhdPlayerWin   int64  `json:"-" gorm:"column:lhd_player_win"`   //玩家赢分
	LhdStrategyId  int    `json:"-" gorm:"column:lhd_strategy_id"`  //策略id
	LhdIsSuppress  bool   `json:"-" gorm:"column:lhd_is_suppress"`  //是否压制
	// LHUserDetail 用户详细展开
	LhdUserid         string `json:"-" gorm:"column:lhd_userid"`           //用户id
	LhdDragon         int64  `json:"-" gorm:"column:lhd_dragon"`           //龙下注
	LhdTiger          int64  `json:"-" gorm:"column:lhd_tiger"`            //虎下注
	LhdTie            int64  `json:"-" gorm:"column:lhd_tie"`              //和下注
	LhdObserve        bool   `json:"-" gorm:"column:lhd_observe"`          //观察局
	LhdResult         string `json:"-" gorm:"column:lhd_result"`           //输赢平
	LhdWin            int64  `json:"-" gorm:"column:lhd_win"`              //输赢
	LhdBeforeScore    int64  `json:"-" gorm:"column:lhd_before_score"`     //账变前分数
	LhdAfterScore     int64  `json:"-" gorm:"column:lhd_after_score"`      //账变后分数
	LhdBeforeCash     int64  `json:"-" gorm:"column:lhd_before_cash"`      //账变前彩金
	LhdAfterCash      int64  `json:"-" gorm:"column:lhd_after_cash"`       //账变后彩金
	LhdBeforeBonus    int64  `json:"-" gorm:"column:lhd_before_bonus"`     //账变前奖励金
	LhdAfterBonus     int64  `json:"-" gorm:"column:lhd_after_bonus"`      //账变后奖励金
	LhdCashMingTax    int64  `json:"-" gorm:"column:lhd_cash_ming_tax"`    //彩金明税
	LhdBonusMingTax   int64  `json:"-" gorm:"column:lhd_bonus_ming_tax"`   //奖励金明税
	LhdCashAnTax      int64  `json:"-" gorm:"column:lhd_cash_an_tax"`      //彩金暗税
	LhdBonusAnTax     int64  `json:"-" gorm:"column:lhd_bonus_an_tax"`     //奖励金暗税
	LhdBeforeBackRate int    `json:"-" gorm:"column:lhd_before_back_rate"` //账变前返奖率
	LhdAfterBackRate  int    `json:"-" gorm:"column:lhd_after_back_rate"`  //账变后返奖率

	// 7UP详情展开 ====================================================================================================
	// 7UP *7UP //7UP详情
	UpWinner     uint32           `json:"-" gorm:"column:up_winner"`                           //赢家
	UpSeatBets   map[uint32]int64 `json:"-" gorm:"column:up_seat_bets;type:Map(UInt32,Int64)"` //_位置下注: 0小,1大,2,3,4,...,12
	UpPointValue int32            `json:"-" gorm:"column:up_point_value"`                      //点数
	UpBets       int64            `json:"-" gorm:"column:up_bets"`                             //总下注
	UpPlayerWin  int64            `json:"-" gorm:"column:up_player_win"`                       //玩家赢分
	UpStrategyId int              `json:"-" gorm:"column:up_strategy_id"`                      //策略id
	UpIsSuppress bool             `json:"-" gorm:"column:up_is_suppress"`                      //是否压制
	// UpUserDetail 用户详细展开
	UpUserid         string `json:"-" gorm:"column:up_userid"`           //用户id
	UpDragon         int64  `json:"-" gorm:"column:up_dragon"`           //龙下注
	UpTiger          int64  `json:"-" gorm:"column:up_tiger"`            //虎下注
	UpTie            int64  `json:"-" gorm:"column:up_tie"`              //和下注
	UpObserve        bool   `json:"-" gorm:"column:up_observe"`          //观察局
	UpResult         string `json:"-" gorm:"column:up_result"`           //输赢平
	UpWin            int64  `json:"-" gorm:"column:up_win"`              //输赢
	UpBeforeScore    int64  `json:"-" gorm:"column:up_before_score"`     //账变前分数
	UpAfterScore     int64  `json:"-" gorm:"column:up_after_score"`      //账变后分数
	UpBeforeCash     int64  `json:"-" gorm:"column:up_before_cash"`      //账变前彩金
	UpAfterCash      int64  `json:"-" gorm:"column:up_after_cash"`       //账变后彩金
	UpBeforeBonus    int64  `json:"-" gorm:"column:up_before_bonus"`     //账变前奖励金
	UpAfterBonus     int64  `json:"-" gorm:"column:up_after_bonus"`      //账变后奖励金
	UpCashMingTax    int64  `json:"-" gorm:"column:up_cash_ming_tax"`    //彩金明税
	UpBonusMingTax   int64  `json:"-" gorm:"column:up_bonus_ming_tax"`   //奖励金明税
	UpCashAnTax      int64  `json:"-" gorm:"column:up_cash_an_tax"`      //彩金暗税
	UpBonusAnTax     int64  `json:"-" gorm:"column:up_bonus_an_tax"`     //奖励金暗税
	UpBeforeBackRate int    `json:"-" gorm:"column:up_before_back_rate"` //账变前返奖率
	UpAfterBackRate  int    `json:"-" gorm:"column:up_after_back_rate"`  //账变后返奖率

	// Crash详情展开 ====================================================================================================
	CrashWinResult    string  `json:"-" gorm:"column:crash_win_result"`                      //开奖结果
	CrashBets         int64   `json:"-" gorm:"column:crash_bets"`                            //总下注
	CrashPlayerLose   int64   `json:"-" gorm:"column:crash_player_lose"`                     //玩家赢分
	CrashStrategyType []int32 `json:"-" gorm:"column:crash_strategy_type;type:Array(Int32)"` //策略类型
	CrashRkyhRfge     bool    `json:"-" gorm:"column:crash_rkyh_rfge"`                       //人狂有祸r肥割
	// CrashUserDetail 用户详细展开
	CrashUserid       string `json:"-" gorm:"column:crash_userid"`         //用户id
	CrashObserve      bool   `json:"-" gorm:"column:crash_observe"`        //是否观察局
	CrashBet          int64  `json:"-" gorm:"column:crash_bet"`            //下注
	CrashMultiple     string `json:"-" gorm:"column:crash_multiple"`       //逃脱倍数
	CrashResult       string `json:"-" gorm:"column:crash_result"`         //输赢平
	CrashWin          int64  `json:"-" gorm:"column:crash_win"`            //输赢
	CrashBeforeScore  int64  `json:"-" gorm:"column:crash_before_score"`   //账变前分数
	CrashAfterScore   int64  `json:"-" gorm:"column:crash_after_score"`    //账变后分数
	CrashBeforeCash   int64  `json:"-" gorm:"column:crash_before_cash"`    //账变前彩金
	CrashAfterCash    int64  `json:"-" gorm:"column:crash_after_cash"`     //账变后彩金
	CrashBeforeBonus  int64  `json:"-" gorm:"column:crash_before_bonus"`   //账变前奖励金
	CrashAfterBonus   int64  `json:"-" gorm:"column:crash_after_bonus"`    //账变后奖励金
	CrashCashMingTax  int64  `json:"-" gorm:"column:crash_cash_ming_tax"`  //彩金明税
	CrashBonusMingTax int64  `json:"-" gorm:"column:crash_bonus_ming_tax"` //奖励金明税
	CrashCashAnTax    int64  `json:"-" gorm:"column:crash_cash_an_tax"`    //彩金暗税
	CrashBonusAnTax   int64  `json:"-" gorm:"column:crash_bonus_an_tax"`   //奖励金暗税

	// AB详情展开 ====================================================================================================
	AbJoker        uint32   `json:"-" gorm:"column:ab_joker"`                      //Key牌
	AbJackpotValue uint32   `json:"-" gorm:"column:ab_jackpot_value"`              //中奖牌
	AbACards       []uint32 `json:"-" gorm:"column:ab_a_cards;type:Array(UInt32)"` //ANDAR
	AbBCards       []uint32 `json:"-" gorm:"column:ab_b_cards;type:Array(UInt32)"` //BAHAR
	AbWinner       uint32   `json:"-" gorm:"column:ab_winner"`                     //赢家 1:ANDAR 2:BAHAR
	AbSideWinner   uint32   `json:"-" gorm:"column:ab_side_winner"`                //赢家2 位置3-10
	AbBets         int64    `json:"-" gorm:"column:ab_bets"`                       //总下注
	AbPlayerWin    int64    `json:"-" gorm:"column:ab_player_win"`                 //玩家赢分
	AbStrategyId   int      `json:"-" gorm:"column:ab_strategy_id"`                //策略id
	AbIsRandom     bool     `json:"-" gorm:"column:ab_is_random"`                  //是否随机开的
	AbFinalFactor  float64  `json:"-" gorm:"column:ab_final_factor"`               //最终系数
	AbCanWinScore  int64    `json:"-" gorm:"column:ab_can_win_score"`              //可赢分
	// AbUserDetail 用户详细展开
	AbUserid       string           `json:"-" gorm:"column:ab_userid"`                           //用户id
	AbSeatBets     map[string]int64 `json:"-" gorm:"column:ab_seat_bets;type:Map(String,Int64)"` //位置下注
	AbObserve      bool             `json:"-" gorm:"column:ab_observe"`                          //是否观察局
	AbResult       string           `json:"-" gorm:"column:ab_result"`                           //输赢平
	AbWin          int64            `json:"-" gorm:"column:ab_win"`                              //输赢
	AbBeforeScore  int64            `json:"-" gorm:"column:ab_before_score"`                     //账变前分数
	AbAfterScore   int64            `json:"-" gorm:"column:ab_after_score"`                      //账变后分数
	AbBeforeCash   int64            `json:"-" gorm:"column:ab_before_cash"`                      //账变前彩金
	AbAfterCash    int64            `json:"-" gorm:"column:ab_after_cash"`                       //账变后彩金
	AbBeforeBonus  int64            `json:"-" gorm:"column:ab_before_bonus"`                     //账变前奖励金
	AbAfterBonus   int64            `json:"-" gorm:"column:ab_after_bonus"`                      //账变后奖励金
	AbCashMingTax  int64            `json:"-" gorm:"column:ab_cash_ming_tax"`                    //彩金明税
	AbBonusMingTax int64            `json:"-" gorm:"column:ab_bonus_ming_tax"`                   //奖励金明税
	AbCashAnTax    int64            `json:"-" gorm:"column:ab_cash_an_tax"`                      //彩金暗税
	AbBonusAnTax   int64            `json:"-" gorm:"column:ab_bonus_an_tax"`                     //奖励金暗税

	// CP详情展开 ====================================================================================================
	CpJackpotOutput int64    `json:"-" gorm:"column:cp_jackpot_output"`             //奖池产出
	CpCardType      uint32   `json:"-" gorm:"column:cp_card_type"`                  //牌型
	CpCards         []uint32 `json:"-" gorm:"column:cp_a_cards;type:Array(UInt32)"` //牌值
	CpWinner        uint32   `json:"-" gorm:"column:cp_winner"`                     //赢家
	CpBets          int64    `json:"-" gorm:"column:cp_bets"`                       //总下注
	CpPlayerWin     int64    `json:"-" gorm:"column:cp_player_win"`                 //玩家赢分
	CpStrategyId    int      `json:"-" gorm:"column:cp_strategy_id"`                //策略id
	CpIsRandom      bool     `json:"-" gorm:"column:cp_is_random"`                  //是否随机开的
	CpFinalFactor   float64  `json:"-" gorm:"column:cp_final_factor"`               //最终系数
	CpCanWinScore   int64    `json:"-" gorm:"column:cp_can_win_score"`              //可赢分
	// CpUserDetail 用户详细展开
	CpUserid       string           `json:"-" gorm:"column:cp_userid"`                           //用户id
	CpSeatBets     map[string]int64 `json:"-" gorm:"column:cp_seat_bets;type:Map(String,Int64)"` //位置下注
	CpResult       string           `json:"-" gorm:"column:cp_result"`                           //输赢平
	CpWin          int64            `json:"-" gorm:"column:cp_win"`                              //输赢
	CpObserve      bool             `json:"-" gorm:"column:cp_observe"`                          //观察局
	CpBeforeScore  int64            `json:"-" gorm:"column:cp_before_score"`                     //账变前分数
	CpAfterScore   int64            `json:"-" gorm:"column:cp_after_score"`                      //账变后分数
	CpBeforeCash   int64            `json:"-" gorm:"column:cp_before_cash"`                      //账变前彩金
	CpAfterCash    int64            `json:"-" gorm:"column:cp_after_cash"`                       //账变后彩金
	CpBeforeBonus  int64            `json:"-" gorm:"column:cp_before_bonus"`                     //账变前奖励金
	CpAfterBonus   int64            `json:"-" gorm:"column:cp_after_bonus"`                      //账变后奖励金
	CpCashMingTax  int64            `json:"-" gorm:"column:cp_cash_ming_tax"`                    //彩金明税
	CpBonusMingTax int64            `json:"-" gorm:"column:cp_bonus_ming_tax"`                   //奖励金明税
	CpCashAnTax    int64            `json:"-" gorm:"column:cp_cash_an_tax"`                      //彩金暗税
	CpBonusAnTax   int64            `json:"-" gorm:"column:cp_bonus_an_tax"`                     //奖励金暗税
	// 红黑详情展开 ====================================================================================================
	RBCardType    []uint32   `json:"-" gorm:"column:rb_card_type;type:Array(UInt32)"`      //牌型
	RBCards       [][]uint32 `json:"-" gorm:"column:rb_a_cards;type:Array(Array(UInt32))"` //牌值
	RBWinner      []uint32   `json:"-" gorm:"column:rb_winner;type:Array(UInt32)"`         //赢家
	RBBets        int64      `json:"-" gorm:"column:rb_bets"`                              //总下注
	RBPlayerWin   int64      `json:"-" gorm:"column:rb_player_win"`                        //玩家赢分
	RBStrategyId  int        `json:"-" gorm:"column:rb_strategy_id"`                       //策略id
	RBCanWinScore int64      `json:"-" gorm:"column:rb_can_win_score"`                     //可赢分
	// RBUserDetail 用户详细展开
	RBUserid       string           `json:"-" gorm:"column:rb_userid"`                           //用户id
	RBSeatBets     map[string]int64 `json:"-" gorm:"column:rb_seat_bets;type:Map(String,Int64)"` //位置下注
	RBResult       string           `json:"-" gorm:"column:rb_result"`                           //输赢平
	RBWin          int64            `json:"-" gorm:"column:rb_win"`                              //输赢
	RBObserve      bool             `json:"-" gorm:"column:rb_observe"`                          //观察局
	RBBeforeScore  int64            `json:"-" gorm:"column:rb_before_score"`                     //账变前分数
	RBAfterScore   int64            `json:"-" gorm:"column:rb_after_score"`                      //账变后分数
	RBBeforeCash   int64            `json:"-" gorm:"column:rb_before_cash"`                      //账变前彩金
	RBAfterCash    int64            `json:"-" gorm:"column:rb_after_cash"`                       //账变后彩金
	RBBeforeBonus  int64            `json:"-" gorm:"column:rb_before_bonus"`                     //账变前奖励金
	RBAfterBonus   int64            `json:"-" gorm:"column:rb_after_bonus"`                      //账变后奖励金
	RBCashMingTax  int64            `json:"-" gorm:"column:rb_cash_ming_tax"`                    //彩金明税
	RBBonusMingTax int64            `json:"-" gorm:"column:rb_bonus_ming_tax"`                   //奖励金明税
	RBCashAnTax    int64            `json:"-" gorm:"column:rb_cash_an_tax"`                      //彩金暗税
	RBBonusAnTax   int64            `json:"-" gorm:"column:rb_bonus_an_tax"`                     //奖励金暗税
}

// TP游戏详情
type TPDetail struct {
	WaterId       string         `bson:"-"`                                    //局号
	SeatId        uint32         `json:"seat_id" bson:"seat_id"`               //座位ID
	UserId        string         `json:"user_id" bson:"user_id"`               //用户ID
	Cards         []uint32       `json:"cards" bson:"cards"`                   //牌型
	CardStr       string         `bson:"-"`                                    //牌型名称
	Score         int64          `json:"score" bson:"score"`                   //结算
	BeforeScore   int64          `json:"before_score" bson:"before_score"`     //账变前分数
	AfterScore    int64          `json:"after_score" bson:"after_score"`       //账变后分数
	BeforeCash    int64          `json:"before_cash" bson:"before_cash"`       //账变前彩金
	AfterCash     int64          `json:"after_cash" bson:"after_cash"`         //账变后彩金
	BeforeBonus   int64          `json:"before_bonus" bson:"before_bonus"`     //账变前奖励金
	AfterBonus    int64          `json:"after_bonus" bson:"after_bonus"`       //账变后奖励金
	Bet           int64          `json:"bet" bson:"bet"`                       //总投注
	Bottom        int64          `json:"bottom" bson:"bottom"`                 //底注
	CashStock     int64          `json:"cash_stock" bson:"cash_stock"`         //彩金库存
	BonusStock    int64          `json:"bonus_stock" bson:"bonus_stock"`       //奖励金库存
	CashMingTax   int64          `json:"cash_ming_tax" bson:"cash_ming_tax"`   //彩金明税
	BonusMingTax  int64          `json:"bonus_ming_tax" bson:"bonus_ming_tax"` //奖励金明税
	CashAnTax     int64          `json:"cash_an_tax" bson:"cash_an_tax"`       //彩金暗税
	BonusAnTax    int64          `json:"bonus_an_tax" bson:"bonus_an_tax"`     //奖励金暗税
	FScore        float64        `bson:"-"`                                    //结算
	FBeforeScore  float64        `bson:"-"`                                    //账变前分数
	FAfterScore   float64        `bson:"-"`                                    //账变后分数
	FBeforeCash   float64        `bson:"-"`                                    //账变前彩金
	FAfterCash    float64        `bson:"-"`                                    //账变后彩金
	FBeforeBonus  float64        `bson:"-"`                                    //账变前奖励金
	FAfterBonus   float64        `bson:"-"`                                    //账变后奖励金
	FBet          float64        `bson:"-"`                                    //总投注
	FBottom       float64        `bson:"-"`                                    //底注
	FCashStock    float64        `bson:"-"`                                    //彩金库存
	FBonusStock   float64        `bson:"-"`                                    //奖励金库存
	FCashMingTax  float64        `bson:"-"`                                    //彩金明税
	FBonusMingTax float64        `bson:"-"`                                    //奖励金明税
	FCashAnTax    float64        `bson:"-"`                                    //彩金暗税
	FBonusAnTax   float64        `bson:"-"`                                    //奖励金暗税
	TPDetailTurn  []TPDetailTurn `json:"tp_detail_turn" bson:"tp_detail_turn"` //轮次详情
	TurnStr1      string         `bson:"-"`                                    // 轮次1详情
	TurnStr2      string         `bson:"-"`                                    // 轮次2详情
	TurnStr3      string         `bson:"-"`                                    // 轮次3详情
	TurnStr4      string         `bson:"-"`                                    // 轮次4详情
}

// 操作轮次详情
type TPDetailTurn struct {
	Turn      int    `json:"turn" bson:"turn"`           //轮次ID
	Operation string `json:"operation" bson:"operation"` //操作
}

// Joker游戏详情
type JOKERDetail struct {
	WaterId         string             `bson:"-"`                                    //局号
	SeatId          uint32             `json:"seat_id" bson:"seat_id"`               //座位ID
	UserId          string             `json:"user_id" bson:"user_id"`               //用户ID
	Cards           []uint32           `json:"cards" bson:"cards"`                   //牌型
	CardStr         string             `bson:"-"`                                    //牌型名称
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
	TurnStr1 string `bson:"-"` // 轮次1详情
	TurnStr2 string `bson:"-"` // 轮次2详情
	TurnStr3 string `bson:"-"` // 轮次3详情
	TurnStr4 string `bson:"-"` // 轮次4详情
}

type JOKERDetailTurn struct {
	Turn      int    `json:"turn" bson:"turn"`           //轮次ID
	Operation string `json:"operation" bson:"operation"` //操作
}

// AK47
type AK47Detail struct {
	WaterId        string            `bson:"-"`                                    //局号
	SeatId         uint32            `json:"seat_id" bson:"seat_id"`               //座位ID
	UserId         string            `json:"user_id" bson:"user_id"`               //用户ID
	Cards          []uint32          `json:"cards" bson:"cards"`                   //牌型
	ChangeCards    []uint32          `json:"change_cards" bson:"change_cards"`     //变牌
	CardStr        string            `bson:"-"`                                    //牌型名称
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
	TurnStr1       string            `bson:"-"`                                    // 轮次1详情
	TurnStr2       string            `bson:"-"`                                    // 轮次2详情
	TurnStr3       string            `bson:"-"`                                    // 轮次3详情
	TurnStr4       string            `bson:"-"`                                    // 轮次4详情
	TurnStr5       string            `bson:"-"`                                    // 轮次5详情
	// LHDetail     LHDetail        `json:"lhUserDetail" bson:"lh_user_detail"`   //
}
type AK47DetailTurn struct {
	Turn      int    `json:"turn" bson:"turn"`           //轮次ID
	Operation string `json:"operation" bson:"operation"` //操作
}

// Rummy
type RMDetail struct {
	WaterId       string     `bson:"-"`                                    //局号
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
	FinalCardsStr string     `bson:"-"`
	InitCardsStr  string     `bson:"-"`
	WildCardStr   string     `bson:"-"`
}

// 龙虎斗游戏详情
type LHDetail struct {
	Winner      int     `json:"winner" bson:"winner"`            //赢家
	DragonValue string  `json:"dragonValue" bson:"dragon_value"` //龙牌值
	TigerValue  string  `json:"tigerValue" bson:"tiger_value"`   //虎牌值
	Bets        int64   `json:"bets" bson:"bets"`                //总下注
	FBets       float64 `bson:"-"`
	PlayerWin   int64   `json:"playerWin" bson:"player_win"` //玩家赢分
	FPlayerWin  float64 `bson:"-"`
	StrategyId  int     `json:"strategyId" bson:"strategy_id"` //策略id
	IsSuppress  bool    `json:"isSuppress" bson:"is_suppress"` //是否压制
	// StrategyTigger bool           `json:"strategyTigger" bson:"strategy_tigger"` //策略生效
	UserDetail   []LHUserDetail `json:"userDetail" bson:"user_detail"` //用户详细
	WaterId      string         `bson:"-"`                             //局号
	Userid       string         `bson:"-"`                             //用户id
	FDragon      float64        `bson:"-"`                             //龙下注
	FTiger       float64        `bson:"-"`                             //虎下注
	FTie         float64        `bson:"-"`                             //和下注
	FWin         float64        `bson:"-"`                             //输赢
	Result       string         `bson:"-"`                             //输赢平
	FBeforeScore float64        `bson:"-"`                             //账变前分数
	FAfterScore  float64        `bson:"-"`                             //账变后分数

}

type UPDetail struct {
	Winner        int            `json:"winner" bson:"winner"`          //赢家
	PointValue    int32          `json:"pointValue" bson:"point_value"` //点数
	Bets          int64          `json:"bets" bson:"bets"`              //总下注
	FBets         float64        `bson:"-"`
	PlayerWin     int64          `json:"playerWin" bson:"player_win"` //玩家赢分
	FPlayerWin    float64        `bson:"-"`
	StrategyId    int            `json:"strategyId" bson:"strategy_id"` //策略id
	IsSuppress    bool           `json:"isSuppress" bson:"is_suppress"` //是否压制
	UserDetail    []LHUserDetail `json:"userDetail" bson:"user_detail"` //用户详细
	WaterId       string         `bson:"-"`                             //局号
	Userid        string         `bson:"-"`                             //用户id
	FDragon       float64        `bson:"-"`                             //龙下注
	FTiger        float64        `bson:"-"`                             //虎下注
	FTie          float64        `bson:"-"`                             //和下注
	FWin          float64        `bson:"-"`                             //输赢
	Result        string         `bson:"-"`                             //输赢平
	Win           int64          `bson:"-"`                             //输赢
	FBeforeScore  float64        `bson:"-"`                             //账变前分数
	FAfterScore   float64        `bson:"-"`                             //账变后分数
	FPlayerFactor float64        `bson:"-"`                             //玩家系数
}

// 玩家下注详情
type LHUserDetail struct {
	Userid         string  `json:"userid" bson:"userid"`                     //用户id
	Dragon         int64   `json:"dragon" bson:"dragon"`                     //龙下注
	Tiger          int64   `json:"tiger" bson:"tiger"`                       //虎下注
	Tie            int64   `json:"tie" bson:"tie"`                           //和下注
	Result         string  `json:"result" bson:"result"`                     //输赢平
	Win            int64   `json:"win" bson:"win"`                           //输赢
	BeforeScore    int64   `json:"beforeScore" bson:"before_score"`          //账变前分数
	AfterScore     int64   `json:"after_score" bson:"after_score"`           //账变后分数
	BeforeCash     int64   `json:"before_cash" bson:"before_cash"`           //账变前彩金
	AfterCash      int64   `json:"after_cash" bson:"after_cash"`             //账变后彩金
	BeforeBonus    int64   `json:"before_bonus" bson:"before_bonus"`         //账变前奖励金
	AfterBonus     int64   `json:"after_bonus" bson:"after_bonus"`           //账变后奖励金
	CashStock      int64   `json:"cash_stock" bson:"cash_stock"`             //彩金库存
	BonusStock     int64   `json:"bonus_stock" bson:"bonus_stock"`           //奖励金库存
	CashMingTax    int64   `json:"cash_ming_tax" bson:"cash_ming_tax"`       //彩金明税
	BonusMingTax   int64   `json:"bonus_ming_tax" bson:"bonus_ming_tax"`     //奖励金明税
	CashAnTax      int64   `json:"cash_an_tax" bson:"cash_an_tax"`           //彩金暗税
	BonusAnTax     int64   `json:"bonus_an_tax" bson:"bonus_an_tax"`         //奖励金暗税
	Observe        bool    `json:"observe" bson:"observe"`                   //是否观察局
	FDragon        float64 `bson:"-"`                                        //龙下注
	FTiger         float64 `bson:"-"`                                        //虎下注
	FTie           float64 `bson:"-"`                                        //和下注
	FWin           float64 `bson:"-"`                                        //输赢
	FBeforeScore   float64 `bson:"-"`                                        //账变前分数
	FAfterScore    float64 `bson:"-"`                                        //账变后分数
	FBeforeCash    float64 `bson:"-"`                                        //账变前彩金
	FAfterCash     float64 `bson:"-"`                                        //账变后彩金
	FBeforeBonus   float64 `bson:"-"`                                        //账变前奖励金
	FAfterBonus    float64 `bson:"-"`                                        //账变后奖励金
	FCashStock     float64 `bson:"-"`                                        //彩金库存
	FBonusStock    float64 `bson:"-"`                                        //奖励金库存
	FCashMingTax   float64 `bson:"-"`                                        //彩金明税
	FBonusMingTax  float64 `bson:"-"`                                        //奖励金明税
	FCashAnTax     float64 `bson:"-"`                                        //彩金暗税
	FBonusAnTax    float64 `bson:"-"`                                        //奖励金暗税
	BeforeBackRate int     `json:"before_back_rate" bson:"before_back_rate"` //账变前返奖率
	AfterBackRate  int     `json:"after_back_rate" bson:"after_back_rate"`   //账变后返奖率
}

type CRASHDetail struct {
	Result        string            `json:"result" bson:"result"` //开奖结果
	Bets          int64             `json:"bets" bson:"bets"`     //总下注
	FBets         float64           `bson:"-"`
	PlayerLose    int64             `json:"playerLose" bson:"player_lose"` //玩家赢分
	FPlayerWin    float64           `bson:"-"`
	StrategyType  []int             `json:"strategyType" bson:"strategy_type"` //策略类型 0：常规；1：扶摇直上;2:欲薅无门
	RkyhRfge      bool              `json:"RkyhRfge" bson:"rkyh_rfge"`         //人狂有祸r肥割
	UserDetail    []CRASHUserDetail `json:"userDetail" bson:"user_detail"`     //用户详细
	WaterId       string            `bson:"-"`                                 //局号
	Userid        string            `bson:"-"`                                 //用户id
	Multiple      string            `bson:"-"`                                 //逃脱倍数
	DResult       string            `bson:"-"`                                 //输赢平
	FBet          float64           `bson:"-"`                                 //下注
	FWin          float64           `bson:"-"`                                 //输赢
	FBeforeScore  float64           `bson:"-"`                                 //账变前分数
	FAfterScore   float64           `bson:"-"`                                 //账变后分数
	FPlayerFactor float64           `bson:"-"`                                 //玩家系数
}

// crash用户详情
type CRASHUserDetail struct {
	Userid        string  `json:"userid" bson:"userid"`                 //用户id
	Observe       bool    `json:"observe" bson:"observe"`               //是否观察局
	Bet           int64   `json:"bet" bson:"bet"`                       // 下注
	Multiple      string  `json:"multiple" bson:"multiple"`             //逃脱倍数
	Result        string  `json:"result" bson:"result"`                 //输赢平
	Win           int64   `json:"win" bson:"win"`                       //输赢
	BeforeScore   int64   `json:"beforeScore" bson:"before_score"`      //账变前分数
	AfterScore    int64   `json:"after_score" bson:"after_score"`       //账变后分数
	BeforeCash    int64   `json:"before_cash" bson:"before_cash"`       //账变前彩金
	AfterCash     int64   `json:"after_cash" bson:"after_cash"`         //账变后彩金
	BeforeBonus   int64   `json:"before_bonus" bson:"before_bonus"`     //账变前奖励金
	AfterBonus    int64   `json:"after_bonus" bson:"after_bonus"`       //账变后奖励金
	CashMingTax   int64   `json:"cash_ming_tax" bson:"cash_ming_tax"`   //彩金明税
	BonusMingTax  int64   `json:"bonus_ming_tax" bson:"bonus_ming_tax"` //奖励金明税
	CashAnTax     int64   `json:"cash_an_tax" bson:"cash_an_tax"`       //彩金暗税
	BonusAnTax    int64   `json:"bonus_an_tax" bson:"bonus_an_tax"`     //奖励金暗税
	FBet          float64 `bson:"-"`                                    //下注
	FWin          float64 `bson:"-"`                                    //输赢
	FBeforeScore  float64 `bson:"-"`                                    //账变前分数
	FAfterScore   float64 `bson:"-"`                                    //账变后分数
	FBeforeCash   float64 `bson:"-"`                                    //账变前彩金
	FAfterCash    float64 `bson:"-"`                                    //账变后彩金
	FBeforeBonus  float64 `bson:"-"`                                    //账变前奖励金
	FAfterBonus   float64 `bson:"-"`                                    //账变后奖励金
	FCashStock    float64 `bson:"-"`                                    //彩金库存
	FBonusStock   float64 `bson:"-"`                                    //奖励金库存
	FCashMingTax  float64 `bson:"-"`                                    //彩金明税
	FBonusMingTax float64 `bson:"-"`                                    //奖励金明税
	FCashAnTax    float64 `bson:"-"`                                    //彩金暗税
	FBonusAnTax   float64 `bson:"-"`                                    //奖励金暗税
}

type ABDetail struct {
	Joker           uint32          `json:"joker" bson:"joker"`                //Key牌
	JackpotValue    uint32          `json:"jackpotValue" bson:"jackpot_value"` //中奖牌
	ACards          []uint32        `json:"aCards" bson:"a_cards"`             //ANDAR
	BCards          []uint32        `json:"bCards" bson:"b_cards"`             //BAHAR
	Winner          uint32          `json:"winner" bson:"winner"`              //赢家 1:ANDAR 2:BAHAR
	SideWinner      uint32          `json:"sideWinner" bson:"side_winner"`     //赢家2 位置3-10
	Bets            int64           `json:"bets" bson:"bets"`                  //总下注
	PlayerWin       int64           `json:"playerWin" bson:"player_win"`       //玩家赢分
	StrategyId      int             `json:"strategyId" bson:"strategy_id"`     //策略id
	IsRandom        bool            `json:"isRandom" bson:"is_random"`         //是否随机开的
	FinalFactor     float64         `json:"finalFactor" bson:"final_factor"`   //最终系数
	CanWinScore     int64           `json:"canWinScore" bson:"can_win_score"`  //可赢分
	WinnerStr       string          `bson:"-"`                                 //赢家 1:ANDAR 2:BAHAR
	SideWinnerStr   string          `bson:"-"`                                 //赢家2 位置3-10
	FBets           float64         `bson:"-"`
	FPlayerWin      float64         `bson:"-"`
	UserDetail      []*ABUserDetail `json:"userDetail" bson:"user_detail"` //用户详细
	JokerStr        string          `bson:"-"`
	JackpotValueStr string          `bson:"-"`
	ANDAR           string          `bson:"-"`
	BAHAR           string          `bson:"-"`
	WaterId         string          `bson:"-"` // 局号
	Userid          string          `bson:"-"` // 用户id
	SeatBets1       float64         `bson:"-"` // ANDAR
	SeatBets2       float64         `bson:"-"` // BAHAR
	SeatBets3       float64         `bson:"-"` // 1-5
	SeatBets4       float64         `bson:"-"` // 6-10
	SeatBets5       float64         `bson:"-"` // 11-15
	SeatBets6       float64         `bson:"-"` // 16-25
	SeatBets7       float64         `bson:"-"` // 26-30
	SeatBets8       float64         `bson:"-"` // 31-35
	SeatBets9       float64         `bson:"-"` // 36-40
	SeatBets10      float64         `bson:"-"` // 41以上
	Result          string          `bson:"-"` // 输赢平
	FWin            float64         `bson:"-"` // 输赢
	FBeforeScore    float64         `bson:"-"` // 账变前分数
	FAfterScore     float64         `bson:"-"` // 账变后分数
}

// 玩家下注详情
type ABUserDetail struct {
	Userid        string           `json:"userid" bson:"userid"`                 //用户id
	SeatBets      map[string]int64 `json:"seatBets" bson:"seat_bets"`            //位置下注
	Observe       bool             `json:"observe" bson:"observe"`               //是否观察局
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
	CardStr       string         `bson:"-"`                                   //牌值
	Winner        uint32         `json:"winner" bson:"winner"`                //赢家
	Bets          int64          `json:"bets" bson:"bets"`                    //总下注
	PlayerWin     int64          `json:"playerWin" bson:"player_win"`         //玩家赢分
	CardTypeStr   string         `bson:"-"`                                   //牌型位置
	StrategyId    int            `json:"strategyId" bson:"strategy_id"`       //策略id
	FBets         float64        `bson:"-"`
	FPlayerWin    float64        `bson:"-"`
	UserDetail    []CPUserDetail `json:"userDetail" bson:"user_detail"` // 用户详细
	WaterId       string         `bson:"-"`                             // 局号
	Userid        string         `bson:"-"`                             // 用户id
	SeatBets1     float64        `bson:"-"`                             // 高牌
	SeatBets2     float64        `bson:"-"`                             // 对子
	SeatBets3     float64        `bson:"-"`                             // 同花
	SeatBets4     float64        `bson:"-"`                             // 顺子
	SeatBets5     float64        `bson:"-"`                             // 同花顺
	SeatBets6     float64        `bson:"-"`                             // 豹子
	Result        string         `bson:"-"`                             // 输赢平
	FBet          float64        `bson:"-"`                             //下注
	FWin          float64        `bson:"-"`                             // 输赢
	FBeforeScore  float64        `bson:"-"`                             // 账变前分数
	FAfterScore   float64        `bson:"-"`                             // 账变后分数
}

// 彩票玩家下注详情
type CPUserDetail struct {
	Userid        string           `json:"userid" bson:"userid"`                 //用户id
	Observe       bool             `json:"observe" bson:"observe"`               //是否观察局
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

type UserGameInning struct {
	Userid    string // 用户ID
	Waterid   string // 局号
	Gtype     int32  // 游戏id
	GtypeName string // 游戏名称
	Result    int    // 输赢 0:输；1：赢
}

type UserGameInningNew struct {
	Userid         string    `bson:"_id" json:"_id"`                 // 用户ID
	RegistArea     int       `bson:"regist_area" json:"regist_area"` // 用户类型
	RegistAreaName string    `json:"-" bson:"-"`
	Diamond        int64     `bson:"diamond" json:"diamond"`         // 彩金
	Ctime          time.Time `bson:"ctime" json:"ctime"`             // 注册时间
	GameNumber     int       `bson:"game_number" json:"game_number"` // 游戏局数
}
