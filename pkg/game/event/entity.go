package event

// 牌型任务
type PokerHandsEvent struct {
	GameType int    `json:"gameType"` // 游戏类型
	PX       uint32 `json:"px"`       // 牌型
	Win      bool   `json:"win"`      // 是否胜利
}

// 充值任务
type RechargeTaskEvent struct {
	Amount uint32 `json:"Amount"` //充值金额
}

// 游戏记录
type GameRecordEvent struct {
	Ttype int32  `json:"ttype"` //任务类型
	Gtype uint32 `json:"gtype"` //游戏类型
	Win   bool   `json:"win"`   //输赢
}

// 充值完成
type RechargeMailEvent struct {
	Amount uint32 `json:"Amount"` //充值金额
}

// 提现完成
type WithdrawMailEvent struct {
	Status int `json:"status"` // 状态
}

// 后台充值
type GiveCash struct {
	Amount uint32 `json:"Amount"` //充值金额
}

// 限时礼包事件
type LimitedGiftEvent struct {
}

// 状态检测事件
type CheckStateEvent struct {
}

// 充值罐子
type RechargeJarEvent struct {
	Give     int64  `json:"give"`     // 赠送金额
	ShopType int    `json:"shopType"` // 充值类型
	ShopId   string `json:"shopId"`   // 充值id
}

// 罐子打码量
type ShopPotFlowEvent struct {
	Score int64 `json:"score"` //输赢分
}

// vip充值
type VIPEvent struct {
	Amount int64 `json:"amount"` // 充值金额
}

// 玩游戏分享
type PlayShareEvent struct {
	Ptype    int32  `json:"ptype"`    // 分享类型 1：玩游戏  2:分享
	DeviceId string `json:"deviceId"` // 设备id
}

// 赠送的罐子
type NormalJarEvent struct {
	Give     int64 `json:"give"`     // 赠送金额
	Multiple int   `json:"multiple"` // 打码倍数
}

// 主动退出
type InitiativeLeaveEvent struct {
	Gtype int `json:"gtype"`
}

type CrashFYZSEvent struct {
	Score int64 `json:"score"` // 输赢分
}

type CrashGCYXScoreEvent struct {
	Score int64 `json:"score"` // 输赢分
}

type CrashGCYXBetEvent struct {
	Bet int64 `json:"bet"` // 下注
}

type CrashGCXYGCOverEvent struct {
}

type CrashGCXYXZOverEvent struct {
	IsOver bool `json:"isOver"` // 是否结束
}

type CrashRKYHTiggerEvent struct {
	Tigger   bool  `json:"tigger"` // 是否触发rkyh
	WinMulti int32 `json:"multi"`  // 赢倍数
	W        int32 `json:"w"`      // 赢局监控
}

type CrashSettlementEvent struct {
	Score     int64 `json:"score"`    // 输赢分
	Multiple  int32 `json:"multiple"` // 位置0打码倍数
	Multiple1 int32 `json:"multiple"` // 位置1打码倍数
	Bet       int64 `json:"bet"`      // 下注
}

type CrashQSHSEvent struct {
}

type CrashMXJLEvent struct {
}

type AviatorFYZSEvent struct {
	Score int64 `json:"score"` // 输赢分
}

type AviatorGCYXScoreEvent struct {
	Score int64 `json:"score"` // 输赢分
}

type AviatorGCYXBetEvent struct {
	Bet int64 `json:"bet"` // 下注
}

type AviatorGCXYGCOverEvent struct {
}

type AviatorGCXYXZOverEvent struct {
	IsOver bool `json:"isOver"` // 是否结束
}

type AviatorFYZSTiggerEvent struct {
	Ttype int `json:"ttype"` // 触发类型 1:钱到了, 2:局数到了
}

type AviatorRKYHTiggerEvent struct {
	Tigger   bool  `json:"tigger"` // 是否触发rkyh
	WinMulti int32 `json:"multi"`  // 赢倍数
	W        int32 `json:"w"`      // 赢局监控
}

type AviatorSettlementEvent struct {
	Score    int64 `json:"score"`    // 输赢分
	Multiple int32 `json:"multiple"` // 打码倍数
	Bet      int64 `json:"bet"`      // 下注
}

type AviatorQSHSEvent struct {
}

type AviatorMXJLEvent struct {
}

type AviatorJackpotEvent struct {
	Jackpot int64 `json:"jackpot"` // 获取的奖池
}

type LHDStrategySettlementEvent struct {
	Id       int   `json:"id" bson:"id"`              // 策略id
	WinScore int64 `json:"winScore" bson:"win_score"` // 赢分
}

type LHDXXSCSettlementEvent struct {
	Disturb bool // 是否扰动
}

type SevenStrategySettlementEvent struct {
	Id       int   `json:"id" bson:"id"`              // 策略id
	WinScore int64 `json:"winScore" bson:"win_score"` // 赢分
}

type SevenXXSCSettlementEvent struct {
	Disturb bool // 是否扰动
}

// vip bank 游戏任务
type VBGameTaskEvent struct {
	GameType     int32   `json:"gameType"`      // 游戏类型
	Win          bool    `json:"win"`           // 是否胜利
	HuaType      uint32  `json:"hua_type"`      // tp,ak47,joker牌型
	CrashEscape  float64 `json:"crash_escape"`  // crash,aviator 位置0逃脱倍数
	CrashEscape1 float64 `json:"crash_escape1"` // crash,aviator 位置1逃脱倍数
	ActTimes     int32   `json:"act_times"`     // 操作次数,回合数
}

type CrashJackpotEvent struct {
	Jackpot int64 `json:"jackpot"` // 获取的奖池
}

type CrashGCYXTriggerEvent struct {
	AvgBet int64 `json:"avgBet"` // 均码量
}

type AviatorGCYXTriggerEvent struct {
	AvgBet int64 `json:"avgBet"` // 均码量
}

type LHDLKYHBSEvent struct {
}

type LHDLKYHTriggerEvent struct {
}

type LHDStrategyResetEvent struct {
	Strategys []int `json:"strategys"`
}

type ABStrategySettlementEvent struct {
	Id       int   `json:"id" bson:"id"`              // 策略id
	WinScore int64 `json:"winScore" bson:"win_score"` // 赢分
}

type CPStrategySettlementEvent struct {
	Id       int   `json:"id" bson:"id"`              // 策略id
	WinScore int64 `json:"winScore" bson:"win_score"` // 赢分
}

type SevenQGYTriggerEvent struct{}

type RBStrategySettlementEvent struct {
	Id       int   `json:"id" bson:"id"`              // 策略id
	WinScore int64 `json:"winScore" bson:"win_score"` // 赢分
}

type BreakingGiftEvent struct {
}
