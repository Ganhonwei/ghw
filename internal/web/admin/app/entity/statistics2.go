package entity

import "time"

// 子项目收益
type SubprojectIncome struct {
	Id            string    `json:"id" bson:"_id"`
	Gtype         int32     `json:"gtype" bson:"gtype"`
	GameName      string    `bson:"-"`
	RoomId        string    `json:"roomId" bson:"room_id"`
	RoomName      string    `bson:"-"`
	GameNumber    int64     `json:"gameNumber" bson:"game_number"`
	WinNumber     int64     `json:"winNumber" bson:"win_number"`
	LoseNumber    int64     `json:"loseNumber" bson:"lose_number"`
	PlayerBet     int64     `json:"playerBet" bson:"player_bet"`
	WinBet        int64     `json:"winBet" bson:"win_bet"`
	LoseBet       int64     `json:"loseBet" bson:"lose_bet"`
	TotalTax      int64     `json:"totalTax" bson:"total_tax"`
	BetAvg        float64   `bson:"-"`
	BetMultiple   float64   `json:"betMultiple" bson:"bet_multiple"`
	WinRate       float64   `json:"winRate" bson:"win_rate"`
	SystemWin     int64     `json:"systemWin" bson:"system_win"`
	TotalRevenue  int64     `json:"totalRevenue" bson:"total_revenue"`
	LastReset     int64     `json:"lastReset" bson:"last_reset"`
	RefreshTime   int64     `json:"refreshTime" bson:"refresh_time"`
	BrNumber      int64     `json:"brNumber" bson:"br_number"` // 百人游戏局数
	BrBets        int64     `json:"brBets" bson:"br_bets"`     // 百人下注
	FBetMultiple  float64   `bson:"-"`
	FSystemWin    float64   `bson:"-"`
	FTotalTax     float64   `bson:"-"`
	FTotalRevenue float64   `bson:"-"`
	SLastReset    time.Time `bson:"-"`
	SRefreshTime  time.Time `bson:"-"`
}

type SubprojectIncomeStats struct {
	UserId      string  `bson:"-"` // 用户id
	GameNumber  int64   `bson:"-"` // 游戏局数
	WinNumber   int64   `bson:"-"` // 赢局
	Bets        int64   `bson:"-"` // 总下注
	WinBets     int64   `bson:"-"` // 总赢
	BrNumber    int64   `bson:"-"` // 百人游戏局数
	BrBets      int64   `bson:"-"` // 百人下注
	TotalTax    int64   `bson:"-"` // 总税收
	BetAvg      int64   `bson:"-"` // 平均注
	BetMultiple float64 `bson:"-"` // 平均倍
}

// 游戏状态
type GameStateData struct {
	GameId    string `json:"game_id"` // 游戏id
	GameName  string `json:"-"`       // 游戏名称
	Number    int    `json:"-"`       // 在线人数
	PayNumber int    `json:"-"`       // 付费玩家在线人数
	SortId    int    `json:"-"`       // 排序ID
}

// 排行榜
type RankingList struct {
	Id              int     `json:"id" bson:"id"`                  // 排名
	Userid          string  `json:"userid" bson:"userid"`          // 用户id
	RegistArea      int     `json:"registArea" bson:"regist_area"` // ab测试 0:A 1:B
	PayAmount       int64   `json:"payAmount" bson:"pay_amount"`
	WithdrawAmount  int64   `json:"withdrawAmount" bson:"withdraw_amount"`
	CarryAmount     int64   `json:"carryAmount" bson:"carry_amount"`
	ProfitAmount    int64   `json:"profitAmount" bson:"profit_amount"`
	FPayAmount      float64 `bson:"-"`
	FWithdrawAmount float64 `bson:"-"`
	FCarryAmount    float64 `bson:"-"`
	FProfitAmount   float64 `bson:"-"`
}

// 排行榜往期数据
type RankingListData struct {
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

// 玩家数据

type PlayerData struct {
	Players        []PlayerInfo    `json:"players" bson:"players"`     // 玩家基础信息
	Lifecycle      []LifecycleInfo `json:"lifecycle" bson:"lifecycle"` // 生命周期
	Games          []UserGameData  `json:"games" bson:"games"`         // 对局统计
	LifecycleCount int             `json:"lifecycleCount" bson:"lifecycle_count"`
	Controls       []GameStrategy  `json:"controls" bson:"controls"` // 控制模型
	ControlCount   int             `json:"controlCount" bson:"control_count"`
}

// 玩家基础信息
type PlayerInfo struct {
	Userid        string `bson:"_id" json:"userid"`                    // 用户id
	UserType      string `bson:"-"`                                    // 用户类型名称
	Nickname      string `bson:"nickname" json:"nickname"`             // 用户昵称
	RealName      string `bson:"real_name" json:"real_name"`           // 真实姓名
	Phone         string `bson:"phone" json:"phone"`                   // 绑定的手机号码
	RegistArea    int    `json:"registArea" bson:"regist_area"`        // ab测试 0:A 1:B
	Diamond       int64  `bson:"diamond" json:"diamond"`               // 钻石(彩金cash)
	ShadowDiamond int64  `bson:"shadow_diamond" json:"shadow_diamond"` // 隐藏钻石(用户看不见的)
	Coin          int64  `bson:"coin" json:"coin"`                     // 金币(奖励金bonus)
	GiveDiamond   int64  `json:"giveDiamond" bson:"give_diamond"`      // 赠送彩金
	OutDiamond    int64  `json:"outDiamond" bson:"out_diamond"`        // 可提现彩金
	CashWater     uint64 `bson:"cash_water" json:"cash_water"`         // 彩金流水(游戏产生的)
	Status        int    `bson:"status" json:"status"`                 // 正常1  锁定2  黑名单3 白名单4
	State         int    `json:"state" bson:"state"`                   // 玩家状态 1.新手 2.正常 3.平民
	CustomTypes   string `json:"customtypes" bson:"custom_types"`      // 自定义用户类型
	// 总资产、显示金额
	Assets         float64                 `bson:"-"`                                // 总资产
	ShowAmount     float64                 `bson:"-"`                                // 显示金额
	FDiamond       float64                 `bson:"-"`                                // 钻石(彩金cash)
	FShadowDiamond float64                 `bson:"-"`                                // 隐藏钻石(用户看不见的)
	FCoin          float64                 `bson:"-"`                                // 金币(奖励金bonus)
	FMoney         float64                 `bson:"-"`                                // 充值总金额(分)
	FCashOut       float64                 `bson:"-"`                                // 提现总金额(分)
	FWin           float64                 `bson:"-"`                                // 赢
	FGiveDiamond   float64                 `bson:"-"`                                // 赠送彩金
	FOutDiamond    float64                 `bson:"-"`                                // 可提现彩金
	Money          uint32                  `bson:"money" json:"money"`               // 充值总金额(分)
	CashOut        int32                   `bson:"cash_out" json:"cash_out"`         // 提现总金额(分)
	WithdrawLogMap map[string]*WithdrawLog `bson:"withdraw_log" json:"withdraw_log"` // 提现记录
	// ad参数
	AD_BundleId string `json:"ad__bundle_id" bson:"ad__bundle_id"` // 包名
	AD_ADID     string `json:"ad__adid" bson:"ad__adid"`           // 设备码
	//时间
	Ctime          time.Time `bson:"ctime" json:"ctime"`           // 注册时间
	LoginTime      time.Time `bson:"login_time" json:"login_time"` // 最后登录时间
	BankCardNumber int       `bson:"-"`                            // 同一银行卡数
	DeviceNumber   int       `bson:"-"`                            //  同一设备数
}

// 生命周期
type LifecycleInfo struct {
	Id          string    `bson:"-"` //id
	LType       int       `bson:"-"` //type
	LTypeName   string    `bson:"-"` //type
	EventType   int       `bson:"-"` // 事件类型 1：注册；2：充值；3：提现
	ProductName string    `bson:"-"` //
	AddDiamond  int64     `bson:"-"` //增加彩金
	NowDiamond  int64     `bson:"-"` //当前彩金
	FAddDiamond float64   `bson:"-"` //增加彩金
	FNowDiamond float64   `bson:"-"` //当前彩金
	Ctime       time.Time `bson:"-"` //流水产生时间
	Stime       int64     `bson:"-"` //流水产生时间
	Ytime       time.Time `bson:"-"` // 原本的流水产生时间
}

// 用户游戏局数统计
type UserGameData struct {
	Id         int64   `bson:"_id"`                           //id
	Userid     string  `bson:"userid"`                        // 用户id
	Gtype      int64   `bson:"gtype"`                         // 游戏类型
	Number     int64   `bson:"number"`                        // 游戏局数
	GameTime   int64   `json:"gameTime" bson:"game_time"`     // 游戏时长
	Bets       int64   `json:"bets" bson:"bets"`              // 总下注
	WinBets    int64   `json:"winBets" bson:"win_bets"`       // 赢局结算
	LoseBets   int64   `json:"loseBets" bson:"lose_bets"`     // 输局结算
	WinNumber  int64   `json:"winNumber" bson:"win_number"`   // 赢局局数
	LoseNumber int64   `json:"loseNumber" bson:"lose_number"` // 输局局数
	GtypeName  string  `bson:"-"`                             // 游戏名称
	Profit     float64 `bson:"-"`                             // 盈亏
	FBets      float64 `bson:"-"`                             // 总下注
	FWinBets   float64 `bson:"-"`                             // 赢局结算
	FLoseBets  float64 `bson:"-"`                             // 输局结算
	WinRate    float64 `bson:"-"`                             // 胜率
	IsProfit   bool    `bson:"-"`                             // 是否为正数
}

// 游戏策略模型记录
type GameStrategy struct {
	Id               string    `bson:"_id"`                                     // id
	UserId           string    `bson:"userid"`                                  // userid
	Gtype            int64     `bson:"gtype"`                                   // 游戏类型
	GtypeName        string    `bson:"-"`                                       // 游戏名称
	StrategyId       string    `json:"strategyId" bson:"strategy_id"`           // 策略模型ID
	EnterNumber      int64     `json:"enterNumber" bson:"enter_number"`         // 进入次数
	EffectiveNumber  int64     `json:"effectiveNumber" bson:"effective_number"` // 生效次数
	EffectiveIncome  int64     `json:"effectiveIncome" bson:"effective_income"` // 生效收益
	FEffectiveIncome float64   `bson:"-"`
	Ctime            int64     `json:"ctime" bson:"ctime"` // 创建时间
	SDate            time.Time `bson:"-"`
}

type ControlModel struct {
	ControId         string    `json:"controId" bson:"contro_id"` // 模型ID
	Date             int64     `bson:"date"`                      // 日期
	SDate            time.Time `bson:"-"`
	GType            int64     `json:"gType" bson:"gtype"`                      // 游戏类型
	GTypeName        string    `bson:"-"`                                       // 游戏名称
	EnterNumber      int64     `json:"enterNumber" bson:"enter_number"`         // 进入次数
	EffectiveNumber  int64     `json:"effectiveNumber" bson:"effective_number"` // 生效次数
	EffectiveIncome  int64     `json:"effectiveIncome" bson:"effective_income"` // 生效收益
	FEffectiveIncome float64   `bson:"-"`
}

type TPStoryChargeStock struct {
	Id        string `bson:"_id" json:"id"`      //房间ID
	Stock     int64  `bson:"stock" json:"stock"` //库存值
	StockMin  int64  `bson:"-" json:"stock"`     //最低使用值
	FStock    string `bson:"-" json:"fStock"`    //库存值
	FStockMin string `bson:"-" json:"fStockMin"` //最低使用值
	Name      string `bson:"-" json:"name"`      //房间名称
}

type TPStoryChargeStockHistory struct {
	Id         string `bson:"_id" json:"id"`              // id
	Userid     string `bson:"userid" json:"userid"`       // 用户id
	GameId     string `bson:"gameid" json:"gameid"`       // 游戏id
	Change     int64  `bson:"change" json:"change"`       // 变化
	Before     int64  `bson:"before" json:"before"`       // 前值
	After      int64  `bson:"after" json:"after"`         // 账变后
	DetailId   string `bson:"detailid" json:"detailid"`   // 轮次
	Timestamp  int64  `bson:"timestamp" json:"timestamp"` // 时间戳
	FChange    string `bson:"-" json:"fchange"`           // 变化
	FBefore    string `bson:"-" json:"fbefore"`           // 前值
	FAfter     string `bson:"-" json:"fafter"`            // 账变后
	FTimestamp string `bson:"-" json:"ftimestamp"`        // 时间戳
}

// 日数据概览
type DayDataList struct {
	Date      uint32 // 日期 20250419
	PackageId string // 渠道
	RegUsers0 int64  // 新注册用户数

	SDate                  string // 日期
	ClassId                string // 渠道类
	AliasId                string // 渠道别名
	LoginUsers             int64  // 日活
	RegUsers               string // 新注册用户数
	AdRegUsers             int64  // 新增设备数
	EffectRegUserRate      string // 有效新增率
	OldLoginUsers          string // 老用户日活
	PayedLoginUsers        int64  // 付费用户日活
	PayAmounts             string // 总充值金额
	WithdrawAmounts        string // 总提现金额
	PaySubWithdraw         string // 充-提
	PayTaxs                string // 代收手续费(分)
	WithdrawTaxs           string // 代付手续费(分)
	ChannelSurplus         string // 通道盈余
	PayWithdrawSurplusRate string // 充提盈余率
	ChannelSurplusRate     string // 通道盈余率

	PayUsers                int64  // 总充值人数
	NewPayUsers             string // 新用户充值人数
	OldPayUsers             string // 老用户充值人数
	FirstPayUsers           int64  // 首充人数
	WithdrawUsers           int64  // 总提现人数
	NewWithdrawUsers        int64  // 新用户提现人数
	OldWithdrawUsers        int64  // 老用户提现人数
	FirstPayWithdrawUsers   int64  // 首充提现人数
	PayRate                 string // 总付费率
	NewPayRate              string // 新用户付费率
	OldPayRate              string // 老用户付费率
	FirstPayRate            string // 首充付费率
	FirstPayRate2           string // 首充付费率(竞)
	FirstPayNextDatePayRate string // 首充次日复充率
	WithdrawRate            string // 总提现率
	NewWithdrawRate         string // 新用户提现率
	OldWithdrawRate         string // 老用户提现率
	PayWithdrawRate         string // 总充值者中提现者占比
	NewPayWithdrawRate      string // 新用户充值者中提现者占比
	OldPayWithdrawRate      string // 老用户充值者中提现者占比
	FirstPayWithdrawRate    string // 首充者中提现者占比

	NewPayAmounts      string // 新用户充值总金额
	OldPayAmounts      string // 老用户充值总金额
	NewWithdrawAmounts string // 新用户提现总金额
	OldWithdrawAmounts string // 老用户提现总金额
	NewPaySubWithdraw  string // 新用户充-提
	OldPaySubWithdraw  string // 老用户充-提
	NewSurplusRate     string // 新用户充提盈余率
	OldSurplusRate     string // 老用户充提盈余率
	ARPU               string // 总充值ARPU 总充值金额 / 登录人数
	NewARPU            string // 新用户ARPU 新用户充值金额 / 新注册数
	OldARPU            string // 老用户ARPU 老用户充值金额 / 老用户当天登录人数
	ARPPU              string // 总充值ARPPU 总充值金额÷总充值人数
	NewARPPU           string // 新用户ARPPU 新用户充值金额÷新用户充值人数
	OldARPPU           string // 老用户ARPPU 老用户充值金额÷老用户充值人数

	Pay2TimesUsers        string // 总当日复购人数+复购率
	NewPay2TimesUsers     string // 新用户当日复购人数+复购率
	OldPay2TimesUsers     string // 老用户当日复购人数+复购率
	FirstPayDatePay2Users string // 首充用户当日复购人数
	PayOrderAvg           string // 总人均付费次数
	NewPayOrderAvg        string // 新用户人均付费次数
	OldPayOrderAvg        string // 老用户人均付费次数

	UnpayLoginUsers          int64            // # 未付费用户日活
	WithdrawAmountsUntax     int64            // # 总提现金额(除扣税)
	ChannelPays              map[string]int64 // # 通道支付金额
	ChannelWithdraws         map[string]int64 // # 通道提现金额
	ChannelWithdrawOrders    map[string]int64 // # 通道提现次数
	PayWithdrawUsers         int64            // # 总充值人数中当日提现的人数
	NewPayWithdrawUsers      int64            // # 新用户充值人数中当日提现的人数
	OldPayWithdrawUsers      int64            // # 老用户充值人数中当日提现的人数
	PayOrders                int64            // # 总付费次数
	NewPayOrders             int64            // # 新用户付费次数
	OldPayOrders             int64            // # 老用户付费次数
	FirstPayDayWithdrawUsers int64            // # 首充当日提现人数
	FirstPayNextDatePayUsers int64            // # 首充次日复充人数

	PayAmounts0      int64 // *总充值金额
	WithdrawAmounts0 int64 // *总提现金额
	PaySubWithdraw0  int64 // *充-提
	PayTaxs0         int64 // *代收手续费(分)
	WithdrawTaxs0    int64 // *代付手续费(分)

	OldLoginUsers0      int64 // *老用户日活
	NewPayUsers0        int64 // *新用户充值人数
	OldPayUsers0        int64 // *老用户充值人数
	NewPayAmounts0      int64 // *新用户充值总金额
	OldPayAmounts0      int64 // *老用户充值总金额
	NewWithdrawAmounts0 int64 // *新用户提现总金额
	OldWithdrawAmounts0 int64 // *老用户提现总金额
	NewPaySubWithdraw0  int64 // *新用户充-提
	OldPaySubWithdraw0  int64 // *老用户充-提

	Pay2TimesUsers0        int64 // *总当日复购人数+复购率
	NewPay2TimesUsers0     int64 // *新用户当日复购人数+复购率
	OldPay2TimesUsers0     int64 // *老用户当日复购人数+复购率
	FirstPayDatePay2Users0 int64 // *首充用户当日复购人数
}

// 经济日报-钱包余额
type FinanceWalletStat struct {
	Id                  string    `json:"id" bson:"_id"`
	Date                uint32    `json:"date" bson:"date"`
	BundleId            string    `json:"bundleId" bson:"bundle_id"`                        // 用户渠道
	RegistArea          int32     `json:"registArea" bson:"regist_area"`                    // 0.A类 1.B类 2.C类
	Ctime               time.Time `json:"ctime" bson:"ctime"`                               // 更新时间
	Cash                int64     `json:"cash" bson:"cash"`                                 // 钱包总余额
	Withdrawable        int64     `json:"withdrawable" bson:"withdrawable"`                 // 可提现余额
	NonWithdrawable     int64     `json:"nonWithdrawable" bson:"non_withdrawable"`          // 不可提现余额
	LiveCash            int64     `json:"liveCash" bson:"live_cash"`                        // 日活钱包余额
	LiveWithdrawable    int64     `json:"liveWithdrawable" bson:"live_withdrawable"`        // 日活可提现余额
	LiveNonWithdrawable int64     `json:"liveNonWithdrawable" bson:"live_non_withdrawable"` // 日活不可提现余额
	Bonus               int64     `json:"bonus" bson:"bonus"`
	LiveBonus           int64     `json:"liveBonus" bson:"live_bonus"`
}

// 经济日报
type FinanceStat struct {
	SDate                string // 日期
	LoginUsers           int64  // 日活
	Cash                 string // 钱包总余额
	Withdrawable         string // 可提现余额
	NonWithdrawable      string // 不可提现余额
	WithdrawableRate     string // 可提现占总余额比
	LiveCash             string // 日活钱包余额
	LiveWithdrawable     string // 日活可提现余额
	LiveNonWithdrawable  string // 日活不可提现余额
	LiveWithdrawableRate string // 日活可提现占总余额比

	GameIncome            string // 游戏收入
	BonusGift             string // bonus发放总额
	BonusIncomeRate       string // bonus发放占收入比
	CashFlow              string // cash流入总额
	CashGift              string // cash发放总额
	CashGiftIncomeRate    string // cash发放占收入比
	CashVb                string // bonus转cash总额
	CashVbIncomeRate      string // bonus转cash占收入比
	PaySubWithdraws       string // 充-提收入
	GameIncomeSubCashFlow string // 游戏收入-cash流入总额
	SurplusMiss           string // 盈余误差

	BonusVip                 string // VIP发放总额
	BonusVipRate             string // VIP占比
	BonusVipUpgrade          string // VIP升级奖励
	BonusVipUpgradeRate      string // 升级奖励占比
	BonusVipWeek             string // VIP周奖励
	BonusVipWeekRate         string // 周奖励占比
	BonusBetRank             string // 排行榜发放总额
	BonusBetRankRate         string // 排行榜占比
	BonusBetRankDaily        string // 日榜
	BonusBetRankDailyRate    string // 日榜占比
	BonusBetRankWeekly       string // 周榜
	BonusBetRankWeeklyRate   string // 周榜占比
	BonusBetRankMonthly      string // 月榜
	BonusBetRankMonthlyRate  string // 月榜占比
	BonusShareAgent          string // 代理发放总额
	BonusShareAgentRate      string // 代理占比
	BonusShareAgentBets      string // 下注返佣总额
	BonusShareAgentBetsRate  string // 下注返佣占比
	BonusShareAgentHeads     string // 人头奖总额
	BonusShareAgentHeadsRate string // 人头奖占比
	BonusShareAgentTasks     string // 里程碑奖总额
	BonusShareAgentTasksRate string // 里程碑奖占比
	BonusTurn                string // 转盘发放总额
	BonusTurnRate            string // 转盘占比
	BonusSubsidy             string // 波动返水总额
	BonusSubsidyRate         string // 返水占比
	BonusGivePack            string // 礼包码发放总额
	BonusGivePackRate        string // 礼包码占比
	BonusWeekCard            string // 周卡发放总额
	BonusWeekCardRate        string // 周卡占比
	BonusPay1Give            string // 首充发放总额
	BonusPay1GiveRate        string // 首充占比
	BonusPay2Give            string // 二充发放总额
	BonusPay2GiveRate        string // 二充占比
	BonusPay3Give            string // 三充发放总额
	BonusPay3GiveRate        string // 三充占比
	BonusPayGive             string // 普充发放总额
	BonusPayGiveRate         string // 普充占比

	CashVip                 string // VIP发放总额
	CashVipRate             string // VIP占比
	CashVipUpgrade          string // VIP升级奖励
	CashVipUpgradeRate      string // 升级奖励占比
	CashVipWeek             string // VIP周奖励
	CashVipWeekRate         string // 周奖励占比
	CashBetRank             string // 排行榜发放总额
	CashBetRankRate         string // 排行榜占比
	CashBetRankDaily        string // 日榜
	CashBetRankDailyRate    string // 日榜占比
	CashBetRankWeekly       string // 周榜
	CashBetRankWeeklyRate   string // 周榜占比
	CashBetRankMonthly      string // 月榜
	CashBetRankMonthlyRate  string // 月榜占比
	CashShareAgent          string // 代理发放总额
	CashShareAgentRate      string // 代理占比
	CashShareAgentBets      string // 下注返佣总额
	CashShareAgentBetsRate  string // 下注返佣占比
	CashShareAgentHeads     string // 人头奖总额
	CashShareAgentHeadsRate string // 人头奖占比
	CashShareAgentTasks     string // 里程碑奖总额
	CashShareAgentTasksRate string // 里程碑奖占比
	CashTurn                string // 转盘发放总额
	CashTurnRate            string // 转盘占比
	CashSubsidy             string // 波动返水总额
	CashSubsidyRate         string // 返水占比
	CashGivePack            string // 礼包码发放总额
	CashGivePackRate        string // 礼包码占比
	CashWeekCard            string // 周卡发放总额
	CashWeekCardRate        string // 周卡占比
	CashPay1Give            string // 首充发放总额
	CashPay1GiveRate        string // 首充占比
	CashPay2Give            string // 二充发放总额
	CashPay2GiveRate        string // 二充占比
	CashPay3Give            string // 三充发放总额
	CashPay3GiveRate        string // 三充占比
	CashPayGive             string // 普充发放总额
	CashPayGiveRate         string // 普充占比

	// ---------------------------------------------------------------------

	Cash0                int64 // # 钱包总余额
	Withdrawable0        int64 // # 可提现余额
	NonWithdrawable0     int64 // # 不可提现余额
	LiveCash0            int64 // # 日活钱包余额
	LiveWithdrawable0    int64 // # 日活可提现余额
	LiveNonWithdrawable0 int64 // # 日活不可提现余额

	GameIncome0      int64 // # 游戏收入
	BonusGift0       int64 // # bonus发放总额
	CashFlow0        int64 // # cash流入总额
	CashGift0        int64 // # cash发放总额
	CashVb0          int64 // # bonus转cash总额
	PaySubWithdraws0 int64 // # 充-提收入

	BonusVip0             int64 // # VIP发放总额
	BonusVipUpgrade0      int64 // # VIP升级奖励
	BonusVipWeek0         int64 // # VIP周奖励
	BonusBetRank0         int64 // # 排行榜发放总额
	BonusBetRankDaily0    int64 // # 日榜
	BonusBetRankWeekly0   int64 // # 周榜
	BonusBetRankMonthly0  int64 // # 月榜
	BonusShareAgent0      int64 // # 代理发放总额
	BonusShareAgentBets0  int64 // # 下注返佣总额
	BonusShareAgentHeads0 int64 // # 人头奖总额
	BonusShareAgentTasks0 int64 // # 里程碑奖总额
	BonusTurn0            int64 // # 转盘发放总额
	BonusSubsidy0         int64 // # 波动返水总额
	BonusGivePack0        int64 // # 礼包码发放总额
	BonusWeekCard0        int64 // # 周卡发放总额
	BonusPay1Give0        int64 // # 首充发放总额
	BonusPay2Give0        int64 // # 二充发放总额
	BonusPay3Give0        int64 // # 三充发放总额
	BonusPayGive0         int64 // # 普充发放总额

	CashVip0             int64 // # VIP发放总额
	CashVipUpgrade0      int64 // # VIP升级奖励
	CashVipWeek0         int64 // # VIP周奖励
	CashBetRank0         int64 // # 排行榜发放总额
	CashBetRankDaily0    int64 // # 日榜
	CashBetRankWeekly0   int64 // # 周榜
	CashBetRankMonthly0  int64 // # 月榜
	CashShareAgent0      int64 // # 代理发放总额
	CashShareAgentBets0  int64 // # 下注返佣总额
	CashShareAgentHeads0 int64 // # 人头奖总额
	CashShareAgentTasks0 int64 // # 里程碑奖总额
	CashTurn0            int64 // # 转盘发放总额
	CashSubsidy0         int64 // # 波动返水总额
	CashGivePack0        int64 // # 礼包码发放总额
	CashWeekCard0        int64 // # 周卡发放总额
	CashPay1Give0        int64 // # 首充发放总额
	CashPay2Give0        int64 // # 二充发放总额
	CashPay3Give0        int64 // # 三充发放总额
	CashPayGive0         int64 // # 普充发放总额
}

// 游戏行为日报
type GameBetStat struct {
	SDate          string // 日期
	LoginUsers     int64  // 日活
	NewLoginUsers  int64  // 新用户日活
	OldLoginUsers  int64  // 老用户日活
	PayLoginUsers  int64  // 付费用户日活
	BetUsers       int64  // 总投注人数
	NewBetUsers    int64  // 新用户投注人数
	OldBetUsers    int64  // 老用户投注人数
	BetPayRate     string // 总充投比
	NewBetPayRate  string // 新用户总充投比
	OldBetPayRate  string // 老用户总充投比
	Bets           string // 总投注金额
	NewBets        string // 新用户投注金额
	OldBets        string // 老用户投注金额
	BetsAvg        string // 总人均日投注额
	NewBetsAvg     string // 新用户人均日投注额
	OldBetsAvg     string // 老用户人均日投注额
	BetsRate       string // 总日活投注率
	PayBetsRate    string // 总付费投注率
	NewBetsRate    string // 新用户日活投注率
	NewPayBetsRate string // 新用户付费投注率
	OldBetsRate    string // 老用户日活投注率
	OldPayBetsRate string // 老用户付费投注率
	RebateRate     string // 总返奖率
	NewRebateRate  string // 新用户返奖率
	OldRebateRate  string // 老用户返奖率
	Income         string // 总游戏收入
	NewIncome      string // 新用户游戏收入
	OldIncome      string // 老用户游戏收入
	KillRate       string // 总杀率
	NewKillRate    string // 新用户杀率
	OldKillRate    string // 老用户杀率

	Bets0      int64 // # 总投注金额
	NewBets0   int64 // # 新用户投注金额
	OldBets0   int64 // # 老用户投注金额
	Rebate0    int64 // # 返奖
	NewRebate0 int64 // # 新用户返奖
	OldRebate0 int64 // # 老用户返奖
	Pays0      int64 // # 总充值金额
	NewPays0   int64 // # 新用户总充值金额
	OldPays0   int64 // # 老用户总充值金额

	Income0    int64 // # 总游戏收入
	NewIncome0 int64 // # 新用户游戏收入
	OldIncome0 int64 // # 老用户游戏收入

	NewPayLoginUsers0 int64 // 新用户付费用户日活
	OldPayLoginUsers0 int64 // 老用户付费用户日活
	PayBetUsers       int64 // 付费打码用户
	NewPayBetUsers    int64 // 新用户付费打码用户
	OldPayBetUsers    int64 // 老用户付费打码用户
}
