package data

import (
	"math"
	"sync"

	"github.com/AsynkronIT/protoactor-go/actor"
	"gopkg.in/mgo.v2/bson"
)

// Role 角色在线临时数据
type Role struct {
	Gate string     //玩家所在节点
	Pid  *actor.PID //进程ID
}

// Gate 角色在线临时数据
type LoginRole struct {
	Token string     //当前登录token
	Pid   *actor.PID //进程ID
}

// 货币
type Currency struct {
	Diamond int64  // 彩金
	Coin    int64  // 奖励金
	Out     int64  // 可提现金
	Tag     uint32 // 自定义标记: 0,1,2,3,4...
}

// ab分类
type RoleAB struct {
	Seed  int
	Mutex sync.Mutex
}

type GiftPop struct {
	Id          string
	TodayPop    int32 // 今日弹出次数
	LastPopTime int64 // 上次弹出时间
	EntryTime   int64 // 生效时间
}

// crash类游戏策略
type CrashUserStrategy struct {
	FYZS CrashFYZS `json:"fyzs" bson:"fyzs"` // 扶摇直上
	YHWM CrashYHWM `json:"yhwm" bson:"yhwm"` // 欲薅无门
	QSHS CrashQSHS `json:"qshs" bson:"qshs"` // 起死回生
	JCFK CrashJCFK `json:"jcfk" bson:"jcfk"` // 奖池风控
	MXJL CrashMXJL `json:"mxjl" bson:"mxjl"` // 冒险奖励
	RKYH CrashRKYH `json:"rkyh" bson:"rkyh"` // 人狂有祸
	GCYX CrashGCYX `json:"gcyx" bson:"gcyx"` // 高潮涌现
}

type CrashFYZS struct {
	PlayTimes      int   `json:"playTimes" bson:"play_times"`            // 扶摇直上:玩游戏次数(退出游戏才算一次)
	TiggerTimes    int   `json:"tiggerTimes" bson:"tigger_times"`        // 扶摇直上:触发次数
	EvoTimes       int   `json:"evoTimes" bson:"evo_times"`              // 扶摇直上:当次玩的局数(与playtimes关联)
	AllTiggerTimes int   `json:"allTiggerTimes" bson:"all_tigger_times"` // 扶摇直上:累计触发次数
	AllEvoTimes    int   `json:"allEvoTimes" bson:"all_evo_times"`       // 扶摇直上:累计玩游戏次数
	WinScore       int64 `json:"winScore" bson:"win_score"`              // 扶摇直上:赢分
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

type CrashGCYX struct {
	EvoTimes     int     `json:"evoTimes" bson:"evo_times"`    // 高潮涌现:当次玩的局数
	IsStateGC    bool    `json:"isStateGC" bson:"is_state_gc"` // 高潮涌现:是否高潮状态
	IsStateXZ    bool    `json:"isStateXZ" bson:"is_state_xz"` // 高潮涌现:是否贤者状态
	XZTimes      int     `json:"xzTimes" bson:"xz_times"`      // 高潮涌现:贤者状态次数
	WinScore     int64   `json:"winScore" bson:"win_score"`    // 高潮涌现:赢分
	M            int     `json:"m" bson:"m"`                   // 高潮涌现:高潮次数m
	R            int     `json:"r" bson:"r"`                   // 高潮涌现:高潮局数r
	DMRecord     []int64 `json:"dmRecord" bson:"dm_record"`    // 高潮涌现:打码记录
	DailyGCTimes int     `json:"gcTimes" bson:"gc_times"`      // 高潮涌现:每日高潮次数
	AvgBet       int64   `json:"avgBet" bson:"avg_bet"`        // 高潮涌现:均码量
}

// lhd游戏策略
type LHDUserStrategy struct {
	RoundBet []int64 `json:"roundBet" bson:"round_bet"` // 每轮下注
	XXSC     LHDXXSC `json:"xxsc" bson:"xxsc"`          // 心想事成
	QSBN     LHDQSBN `json:"qsbn" bson:"qsbn"`          // 求死不能
	LKYH     LHDLKYH `json:"lkyh" bson:"lkyh"`          // 龙狂有祸
	GCYX     LHDGCYX `json:"gcyx" bson:"gcyx"`          // 高潮涌现
}
type LHDXXSC struct {
	DisturbCoolDown int   `json:"disturbCoolDown" bson:"disturb_cool_down"` // 心想事成:干扰cd
	WinLength       int   `json:"winLength" bson:"win_length"`              // 心想事成:连赢局数
	Evo             bool  `json:"evo" bson:"evo"`                           // 心想事成:当次生效过策略没有
	TriggerTimes    int   `json:"triggerTimes" bson:"trigger_times"`        // 心想事成:有效触发次数u
	MaxTriggerTimes int   `json:"maxTriggerTimes" bson:"max_trigger_times"` // 心想事成:触发次数u总
	WinScore        int64 `json:"winScore" bson:"win_score"`                // 心想事成:净赢分
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

type LHDGCYX struct {
	Highing      bool    `json:"highing" bson:"highing"`            // 是否处于高潮状态
	TriggerTimes int     `json:"triggerTimes" bson:"trigger_times"` // 高潮涌现:每日有效触发次数T
	Hp           int     `json:"hp" bson:"hp"`                      // 当前高潮率
	C            int     `json:"c" bson:"c"`                        // 贤者局冷却
	M            int     `json:"m" bson:"m"`                        // 高潮次数 人为制造高潮的次数
	BetAvg       float64 `json:"bet_avg" bson:"bet_avg"`            // 进入高潮时前f局均打码量

	R    int   `json:"r" bson:"r"`       // 高潮局数 一次高潮可能对应多局
	Win  int64 `json:"win" bson:"win"`   // 高潮局中赢钱金额
	Lose int64 `json:"lose" bson:"lose"` // 高潮局中输钱金额
}

// ab游戏策略
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

// 新手配置
type Beginner struct {
	Id               int32     `json:"id" bson:"_id"`
	RegistMode1      []int32   `json:"registMode1" bson:"regist_mode1"`
	RegistMode2      []int32   `json:"registMode2" bson:"regist_mode2"`
	NoviceGame       [][]int32 `json:"noviceGame" bson:"novice_game"`
	GiveBouns        int32     `json:"giveBouns" bson:"give_bouns"`
	Recharge         int32     `json:"recharge" bson:"recharge"`
	PlayerGames      int32     `json:"playerGames" bson:"player_games"`
	OnlineTime       int32     `json:"onlineTime" bson:"online_time"`
	OutCashInterval  []int32   `json:"outCashInterval" bson:"out_cash_interval"`
	Winning          []int32   `json:"winning" bson:"winning"`
	OutCashLimited   int32     `json:"outCashLimited" bson:"out_cash_limited"`
	FrothMaxRecharge int32     `json:"frothMaxRecharge" bson:"froth_max_recharge"`
	FrothGiftRate    int       `json:"frothGiftRate" bson:"froth_gift_rate"`
	FrothWonRate     []int     `json:"frothWonRate" bson:"froth_won_rate"`
	FrothBetInterval []int     `json:"frothCrashBoom" bson:"froth_crash_boom"`
	FactorSeed       float64   `json:"factorSeed" bson:"factor_seed"`
	NewbiewToCivil   []int32   `json:"newbiewToCivil" bson:"newbiew_to_civil"`
}

func (b *Beginner) Save() {
	Upsert(Beginners, bson.M{"_id": b.Id}, b)
}

func GetBeginnerList() []Beginner {
	var list []Beginner
	ListByQ(Beginners, bson.M{}, &list)
	return list
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

func (b *ChargeClassify) Save() {
	Upsert(ChargeClassifys, bson.M{"_id": 1}, b)
}

func GetChargeClassify() ChargeClassify {
	bean := ChargeClassify{}
	GetByQ(ChargeClassifys, bson.M{"_id": 1}, &bean)
	return bean
}

func (c Currency) GetSum() int64 {
	return c.Diamond + c.Coin
}

func (c Currency) Scale(s float64) Currency {
	c.Diamond = int64(math.Floor(float64(c.Diamond) * s))
	c.Coin = int64(math.Ceil(float64(c.Coin) * s))
	return c
}

func (c *Currency) Merge(s Currency) {
	c.Diamond += s.Diamond
	c.Coin += s.Coin
}

func (c *Currency) Sub(s Currency) {
	c.Diamond -= s.Diamond
	c.Coin -= s.Coin
}

func (c *Currency) SetTag(tag uint32) {
	tag = 1 << tag
	c.Tag |= tag
}

func (c *Currency) HasTag(tag uint32) bool {
	tag = 1 << tag
	return c.Tag&tag == tag
}

type TpChargeAmount struct {
	ChargeAmount  uint32 // 充值金额
	GiveAmount    uint32 // 赠送金额
	Target        int32  // 目标值
	ChargeLoss    int64  // 平台亏损值修正
	ViceAccountUp int32  // 小号升档
}

// 转盘活动
type ActivityTurn struct {
	TurnTime          int64   `json:"turnTime" bson:"turn_time"`                    // 本轮游戏开始时间
	TurnEtime         int64   `json:"turnEtime" bson:"turn_etime"`                  // 本轮结束游戏时间
	GiveSelected      bool    `json:"giveSelected" bson:"give_selected"`            // 本轮是否已领取4选1礼包
	GiveIndex         int32   `json:"giveIndex" bson:"give_index"`                  // 本轮打开的4选1礼包索引
	GiveAmount        int64   `json:"giveAmount" bson:"give_amount"`                // 礼包序幕金
	GiveAmount3       []int64 `json:"giveAmount3" bson:"give_amount3"`              // 其他3个展示礼包序幕金
	DrawedTimes       int32   `json:"drawedTimes" bson:"drawed_times"`              // 已抽奖次数
	DrawTimesFree     int32   `json:"drawTimesFree" bson:"draw_times_free"`         // 剩余免费抽奖次数
	DrawTimesInvite   int32   `json:"drawTimesInvite" bson:"draw_times_invite"`     // 剩余邀请抽奖次数
	DrawTimesLucky    int32   `json:"drawTimesLucky" bson:"draw_times_lucky"`       // 有效邀请中满额概率抽奖次数
	NextFreeDrawTime  int64   `json:"nextFreeDrawTime" bson:"next_free_draw_time"`  // 下次获得免费抽奖时间(秒)
	Score             int64   `json:"score" bson:"score"`                           // 已获得金额
	ScoreTarget       int64   `json:"scoreTarget" bson:"score_target"`              // 目标金额
	ScoreType         int32   `json:"scoreType" bson:"score_type"`                  // 奖金类型:（1.bonus,2.cash,3.withdrawalble）
	TackedPrize       bool    `json:"tackedPrize" bson:"tacked_prize"`              // 是否已领取奖励
	DrawedTimesFree   int32   `json:"drawedTimesFree" bson:"drawed_times_free"`     // 已使用免费抽奖次数
	DrawedTimesInvite int32   `json:"drawedTimesInvite" bson:"drawed_times_invite"` // 已使用邀请抽奖次数
}

type PKBankInfo struct {
	BankName    string `json:"bankName" bson:"bank_name"`
	BankAccount string `json:"bankAccount" bson:"bank_account"`
	BankHolder  string `json:"bankHolder" bson:"bank_holder"`
}

type PKWithdrawInfo struct {
	DefaultType   int32            `json:"defaultType" bson:"default_type"`
	WithdrawTimes map[string]int64 `json:"withdrawBanks" bson:"withdraw_banks"`
}
