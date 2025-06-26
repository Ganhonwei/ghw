package data

// 飞机游戏策略
type AviatorUserStrategy struct {
	FYZS AviatorFYZS `json:"fyzs" bson:"fyzs"` // 扶摇直上
	YHWM AviatorYHWM `json:"yhwm" bson:"yhwm"` // 欲薅无门
	QSHS AviatorQSHS `json:"qshs" bson:"qshs"` // 起死回生
	JCFK AviatorJCFK `json:"jcfk" bson:"jcfk"` // 奖池风控
	MXJL AviatorMXJL `json:"mxjl" bson:"mxjl"` // 冒险奖励
	RKYH AviatorRKYH `json:"rkyh" bson:"rkyh"` // 人狂有祸
	GCYX AviatorGCYX `json:"gcyx" bson:"gcyx"` // 高潮涌现
}

type AviatorFYZS struct {
	TiggerTimes    int   `json:"tiggerTimes" bson:"tigger_times"`        // 扶摇直上:触发次数
	EvoTimes       int   `json:"evoTimes" bson:"evo_times"`              // 扶摇直上:当次玩的局数(与playtimes关联)
	AllTiggerTimes int   `json:"allTiggerTimes" bson:"all_tigger_times"` // 扶摇直上:累计触发次数
	AllEvoTimes    int   `json:"allEvoTimes" bson:"all_evo_times"`       // 扶摇直上:累计玩游戏次数
	WinScore       int64 `json:"winScore" bson:"win_score"`              // 扶摇直上:赢分
}

type AviatorYHWM struct {
	MonitorRounds    int     `json:"monitorRounds" bson:"monitor_rounds"`    // 欲薅无门:监控胜局(连胜不能断)
	AviatorMulpitles []int32 `json:"crashMulpitles" bson:"crash_mulpitles"`  // 欲薅无门:逃跑倍数
	TiggerTimes      int     `json:"tiggerTimes" bson:"tigger_times"`        // 欲薅无门:触发次数
	WinScore         int64   `json:"winScore" bson:"win_score"`              // 欲薅无门:赢分
	RecyleScore      int64   `json:"recyleScore" bson:"recyle_score"`        // 欲薅无门:回收分数
	AllRecyleScore   int64   `json:"allRecyleScore" bson:"all_recyle_score"` // 欲薅无门:累计回收分数
	Tigger           bool    `json:"tigger" bson:"tigger"`                   // 欲薅无门:是否生效中
}

type AviatorQSHS struct {
	WinRounds    []int32 `json:"winRounds" bson:"win_rounds"`       // 赢局逃跑倍数
	TriggerTimes int     `json:"triggerTimes" bson:"trigger_times"` // 起死回生：生效次数
}

type AviatorJCFK struct {
	TriggerTimes int   `json:"triggerTimes" bson:"trigger_times"` // 奖池风控:触发jackpot次数
	JackpotVal   int64 `json:"jackpotVal" bson:"jackpot_val"`     // 奖池风控:获得的jackpot值
}

type AviatorMXJL struct {
	TriggerTimes int     `json:"triggerTimes" bson:"trigger_times"` // 冒险奖励:触发次数
	Bets         []int64 `json:"bets" bson:"bets"`                  // 冒险奖励:下注
}

type AviatorRKYH struct {
	TriggerTimes int     `json:"triggerTimes" bson:"trigger_times"` // 触发rkyh次数
	WinMulpitles []int32 `json:"winMulpitles" bson:"win_mulpitles"` // 前n赢局的局逃跑倍数
}

type AviatorGCYX struct {
	EvoTimes     int     `json:"evoTimes" bson:"evo_times"`    // 高潮涌现:当次玩的局数
	IsStateGC    bool    `json:"isStateGC" bson:"is_state_gc"` // 高潮涌现:是否高潮状态
	IsStateXZ    bool    `json:"isStateXZ" bson:"is_state_xz"` // 高潮涌现:是否贤者状态
	XZTimes      int     `json:"xzTimes" bson:"xz_times"`      // 高潮涌现:贤者状态次数
	WinScore     int64   `json:"winScore" bson:"win_score"`    // 高潮涌现:赢分
	M            int     `json:"m" bson:"m"`                   // 高潮涌现:高潮次数m
	R            int     `json:"r" bson:"r"`                   // 高潮涌现:高潮局数r
	DMRecord     []int64 `json:"dmRecord" bson:"dm_record"`    // 高潮涌现:打码记录
	DailyGCTimes int     `json:"gcTimes" bson:"gc_times"`      // 高潮涌现:每日高潮次数
	AvgBet       int64   `json:"avgBet" bson:"avg_bet"`        // 高潮涌现:平均打码量
}
