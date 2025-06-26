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
	IsIpRepeat    bool `bson:"-"` // IP是否重复
	IsAppIdRepeat bool `bson:"-"` // 设备号是否重复
	AndroidScore  int  `bson:"-"` // 安卓评分
	IsBlogger     bool `bson:"-"` // 是否是博主账号 0:否; 1:是
}
type ShareData struct {
	BetAmount int64  `json:"betAmount" bson:"bet_amount"` //总打码量
	UserId    string `json:"userId" bson:"user_id"`
	Name      string `json:"name" bson:"name"`            //名称
	LastLogin int64  `json:"lastLogin" bson:"last_login"` //上次登录
	Recharge  int64  `json:"recharge" bson:"recharge"`    //充值金额
	Receive   bool   `json:"receive" bson:"receive"`      //是否领取奖励
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
	WaterId            string         `bson:"_id" json:"_id"`               //局号
	BeginTime          int64          `bson:"begin_time" json:"begin_time"` //开始时间
	STime              time.Time      `bson:"-"`
	EndTime            int64          `bson:"end_time" json:"end_time"` //结束时间
	ETime              time.Time      `bson:"-"`
	Gtype              int32          `bson:"gtype" json:"gtype"` //所在游戏 1：TP;2: LHD; 3：SEVEN; 4:RUMMY; 5:AK47; 6:JOKER; 7:CRASH
	GtypeName          string         `bson:"-"`
	RoomId             string         `bson:"room_id" json:"room_id"`                   //房间ID
	DeskId             string         `bson:"desk_id" json:"desk_id"`                   //桌子ID
	Players            string         `bson:"players" json:"players"`                   //参与玩家ID
	CardTypeId         int            `bson:"card_type_id" json:"card_type_id"`         //牌型ID
	ChangeCardType     int            `bson:"change_card_type" json:"change_card_type"` // 换牌局类型 0：没有 1：开局换牌 2：局中换牌
	ChangeCardTypeName string         `bson:"-"`
	WinScore           int            `bson:"win_score" json:"win_score"`         //当前赢分
	PlayerFactor       int            `bson:"player_factor" json:"player_factor"` //玩家系数
	ControlType        int            `bson:"control_type" json:"control_type"`   //控制方式 1:当前赢分;2：房间系数
	ControlTypeName    string         `bson:"-"`
	WinScore1          int            `bson:"win_score_1" json:"win_score_1"`    //当局赢分
	WinScore2          int            `bson:"win_score_2" json:"win_score_2"`    //当前赢分
	ChargeMoney        int            `bson:"charge_money" json:"charge_money"`  //充值金额
	FWinScore          float64        `bson:"-"`                                 //当前赢分(开局)
	FWinScore1         float64        `bson:"-"`                                 //当局赢分(局中)
	FWinScore2         float64        `bson:"-"`                                 //当前赢分(局中)
	FChargeMoney       float64        `bson:"-"`                                 //充值金额
	IsStrategy         bool           `bson:"is_strategy" json:"is_strategy"`    //是否是策略局
	StrategyType       int            `json:"strategyType" bson:"strategy_type"` // 策略局类型
	TPDetail           []TPDetail     //TP详情
	JOKERDetail        []*JOKERDetail //JOKER详情
	AK47Detail         []*AK47Detail  //AK47详情
	RMDetail           []*RMDetail    //RM详情
	LHDetail           *LHDetail      //LH详情
	UPDetail           *UPDetail      //7up详情
	CRASHDetail        *CRASHDetail   //crash
}

// TP游戏详情
type TPDetail struct {
	SeatId        uint32         `json:"seat_id" bson:"seat_id"`               //座位ID
	UserId        string         `json:"user_id" bson:"user_id"`               //用户ID
	Cards         []uint32       `json:"cards" bson:"cards"`                   //牌型
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
	Winner      int            `json:"winner" bson:"winner"`            //赢家
	DragonValue string         `json:"dragonValue" bson:"dragon_value"` //龙牌值
	TigerValue  string         `json:"tigerValue" bson:"tiger_value"`   //虎牌值
	Bets        int64          `json:"bets" bson:"bets"`                //总下注
	FBets       float64        `bson:"-"`
	PlayerWin   int64          `json:"playerWin" bson:"player_win"` //玩家赢分
	FPlayerWin  float64        `bson:"-"`
	UserDetail  []LHUserDetail `json:"userDetail" bson:"user_detail"` //用户详细
}

type UPDetail struct {
	Winner     int            `json:"winner" bson:"winner"`          //赢家
	PointValue int32          `json:"pointValue" bson:"point_value"` //点数
	Bets       int64          `json:"bets" bson:"bets"`              //总下注
	FBets      float64        `bson:"-"`
	PlayerWin  int64          `json:"playerWin" bson:"player_win"` //玩家赢分
	FPlayerWin float64        `bson:"-"`
	UserDetail []LHUserDetail `json:"userDetail" bson:"user_detail"` //用户详细
}

// 玩家下注详情
type LHUserDetail struct {
	Userid        string  `json:"userid" bson:"userid"`                 //用户id
	Dragon        int64   `json:"dragon" bson:"dragon"`                 //龙下注
	Tiger         int64   `json:"tiger" bson:"tiger"`                   //虎下注
	Tie           int64   `json:"tie" bson:"tie"`                       //和下注
	Result        string  `json:"result" bson:"result"`                 //输赢平
	Win           int64   `json:"win" bson:"win"`                       //输赢
	BeforeScore   int64   `json:"beforeScore" bson:"before_score"`      //账变前分数
	AfterScore    int64   `json:"after_score" bson:"after_score"`       //账变后分数
	BeforeCash    int64   `json:"before_cash" bson:"before_cash"`       //账变前彩金
	AfterCash     int64   `json:"after_cash" bson:"after_cash"`         //账变后彩金
	BeforeBonus   int64   `json:"before_bonus" bson:"before_bonus"`     //账变前奖励金
	AfterBonus    int64   `json:"after_bonus" bson:"after_bonus"`       //账变后奖励金
	CashStock     int64   `json:"cash_stock" bson:"cash_stock"`         //彩金库存
	BonusStock    int64   `json:"bonus_stock" bson:"bonus_stock"`       //奖励金库存
	CashMingTax   int64   `json:"cash_ming_tax" bson:"cash_ming_tax"`   //彩金明税
	BonusMingTax  int64   `json:"bonus_ming_tax" bson:"bonus_ming_tax"` //奖励金明税
	CashAnTax     int64   `json:"cash_an_tax" bson:"cash_an_tax"`       //彩金暗税
	BonusAnTax    int64   `json:"bonus_an_tax" bson:"bonus_an_tax"`     //奖励金暗税
	FDragon       float64 `bson:"-"`                                    //龙下注
	FTiger        float64 `bson:"-"`                                    //虎下注
	FTie          float64 `bson:"-"`                                    //和下注
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

type CRASHDetail struct {
	Result     string            `json:"result" bson:"result"` //开奖结果
	Bets       int64             `json:"bets" bson:"bets"`     //总下注
	FBets      float64           `bson:"-"`
	PlayerLose int64             `json:"playerLose" bson:"player_lose"` //玩家赢分
	FPlayerWin float64           `bson:"-"`
	UserDetail []CRASHUserDetail `json:"userDetail" bson:"user_detail"` //用户详细
}

// crash用户详情
type CRASHUserDetail struct {
	Userid        string  `json:"userid" bson:"userid"`                 //用户id
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
