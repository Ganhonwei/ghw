package entity

// CrashPlayerStat crash统计
type CrashPlayerStat struct {
	Id                   string      `bson:"_id"` // date-userid
	Date                 int64       `bson:"date"`
	DateStr              string      `bson:"date_str"`
	Userid               string      `bson:"userid"`
	NewReg               bool        `bson:"new_reg"`                 // 是新顾客
	AllRounds            int32       `bson:"all_rounds"`              // 总局数
	GameTimes            int64       `bson:"game_times"`              // 游戏总时长(秒)
	Bets                 int64       `bson:"bets"`                    // 总打码量
	BetRounds            int32       `bson:"bet_rounds"`              // 下注局数
	WinRounds            int32       `bson:"win_rounds"`              // 胜利局数
	LoseRounds           int32       `bson:"lose_rounds"`             // 失败局数
	TieRounds            int32       `bson:"tie_rounds"`              // 和局数
	WinBets              int64       `bson:"win_bets"`                // 胜局打码量
	LoseBets             int64       `bson:"lose_bets"`               // 败局打码量
	TieBets              int64       `bson:"tie_bets"`                // 和局打码量
	Wins                 int64       `bson:"wins"`                    // 胜局赢钱金额
	Loses                int64       `bson:"loses"`                   // 败局输钱金额
	Cash                 int64       `bson:"cash"`                    // 净赢
	MulpitleSum          float64     `bson:"mulpitle_sum"`            // 所有局总爆炸倍数和
	Mulpitles            [][]float64 `bson:"mulpitles"`               // 局数爆炸倍数
	WinEscapeMulpitleSum float64     `bson:"win_escape_mulpitle_sum"` // 赢局逃脱倍数和
	WinEscapeMulpitles   [][]float64 `bson:"win_escape_mulpitles"`    // 赢局逃脱倍数
	WinMulpitleSum       float64     `bson:"win_mulpitle_sum"`        // 赢局爆炸倍数和
	WinMulpitles         [][]float64 `bson:"win_mulpitles"`           // 赢局数爆炸倍数
	LoseMulpitleSum      float64     `bson:"lose_mulpitle_sum"`       // 输局爆炸倍数和
	LoseMulpitles        [][]float64 `bson:"lose_mulpitles"`          // 输局爆炸倍数,算中位数
	RoundBetAvg          int64       `bson:"round_bet_avg"`           // 局均码
	AD_BundleId          string      `bson:"ad__bundle_id"`           // 渠道
	Channel1             string      `bson:"channel1"`                // 渠道别名
	RegistArea           int         `bson:"regist_area"`             // 账号类型 ab测试 0:A 1:B 2:C
	Money                uint32      `bson:"money"`                   // 充值总金额(分)
	ObserveRounds        int32       `bson:"observe_rounds"`          // 观察局数
	Players              int32       `bson:"players"`                 // 总玩crash人数
	NewPlayers           int32       `bson:"new_players"`             // 新顾客玩crash人数
	OldPlayers           int32       `bson:"old_players"`             // 老顾客玩crash人数
	AllRoundsList        []int32     `bson:"all_rounds_list"`         // 玩家局数list
	Userids              []string    `bson:"userids"`                 // 汇总时的userid列表

	No                       string `bson:"-"` // 序号
	FNickname                string `bson:"-"` // 总打码量
	FCtime                   string `bson:"-"` // 注册时间
	FLoginTime               string `bson:"-"` // 最后登录时间
	FLiveDays                int    `bson:"-"` // 存活天数
	FLoseDays                int    `bson:"-"` // 流失天数
	FRegistArea              string `bson:"-"` // 账号类型
	FUserType                string `bson:"-"` // 玩家类型
	FGameTimes               string `bson:"-"` // 游戏总时长(分)
	FBets                    string `bson:"-"` // 总打码量
	FRoundBetsAvg            string `bson:"-"` // 局均打码量
	FWinRate                 string `bson:"-"` // 胜率
	FRebateRate              string `bson:"-"` // 返奖率
	FWinBets                 string `bson:"-"` // 胜局打码量
	FLoseBets                string `bson:"-"` // 败局打码量
	FWinBetsAvg              string `bson:"-"` // 胜局均码量
	FLoseBetsAvg             string `bson:"-"` // 败局均码量
	FWins                    string `bson:"-"` // 胜局赢钱金额
	FLoses                   string `bson:"-"` // 败局输钱金额
	FCash                    string `bson:"-"` // 净赢
	FWinsAvg                 string `bson:"-"` // 胜局均盈利
	FLosesAvg                string `bson:"-"` // 败局均输额
	FCashAvg                 string `bson:"-"` // 局均净赢
	FMulpitleAvg             string `bson:"-"` // 局均实际爆炸倍数
	FMulpitleMedian          string `bson:"-"` // 每局实际爆炸倍数中位数
	FWinEscapeMulpitleMax    string `bson:"-"` // 赢局最高逃脱倍数
	FWinEscapeMulpitleAvg    string `bson:"-"` // 赢局局均逃脱倍数
	FWinEscapeMulpitleMedian string `bson:"-"` // 赢局逃脱倍数中位数
	FWinMulpitleAvg          string `bson:"-"` // 赢局局均实际爆炸倍数
	FWinMulpitleMedian       string `bson:"-"` // 赢局实际爆炸倍数中位数
	FMulpitleDiff            string `bson:"-"` // 倍差空间
	FLoseMulpitleAvg         string `bson:"-"` // 输局局均实际爆炸倍数
	FLoseMulpitleMedian      string `bson:"-"` // 输局实际爆炸倍数中位数
	FWinEscapeRate           string `bson:"-"` // 赢逃输爆比
	FWinLoseRate             string `bson:"-"` // 赢输爆倍比
	ObserveRoundsAvg         string `bson:"-"` // 局均观察次数
	ObserveRoundsRate        string `bson:"-"` // 观察局数占比
	FMoney                   string `bson:"-"` // 总充值
	FCashOut                 string `bson:"-"` // 总提现
	FWinMoney                string `bson:"-"` // 总赢

	SMoney   int   `bson:"-"` // 总汇充值
	SCashOut int   `bson:"-"` // 总汇提现
	SDiamond int64 `bson:"-"` // 总汇携带分

	FLoginedUsers       int    `bson:"-"` // 日活
	FLoginedNewUsers    int    `bson:"-"` // 新用户日活
	FLoginedOldUsers    int    `bson:"-"` // 老用户日活
	FPlayerRate         string `bson:"-"` // 玩crash人数占日活比
	FNewPlayerRate      string `bson:"-"` // 新顾客玩crash占新顾客日活比
	FOldPlayerRate      string `bson:"-"` // 老顾客玩crash占老顾客日活比
	FPlayerRoundsAvg    string `bson:"-"` // 人均局数
	FPlayerRoundsMedian int32  `bson:"-"` // 局数中位数
	FGameTimesAvg       string `bson:"-"` // 游戏局均时长（min)
	FGameTimesPlayerAvg string `bson:"-"` // 游戏人均时长（min)
	FPlayerBetsAvg      string `bson:"-"` // 人均打码
}

// Crash扶摇直上
type CrashFYZSStat struct {
	Id                       string  `bson:"_id"` // date-userid
	Date                     int64   `bson:"date"`
	DateStr                  string  `bson:"date_str"`
	Userid                   string  `bson:"userid"`
	IsFit                    bool    `json:"isFit" bson:"is_fit"`                                         // 是否是适配策略的人
	IsTrigger                bool    `json:"isTrigger" bson:"is_trigger"`                                 // 是否是触发策略的人
	TiggerTimes              int     `json:"tiggerTimes" bson:"tigger_times"`                             // u 触发策略有效次数
	AllTiggerTimes           int     `json:"allTiggerTimes" bson:"all_tigger_times"`                      // u总 累计触发次数
	AllEvoTimes              int     `json:"allEvoTimes" bson:"all_evo_times"`                            // r总 累计玩游戏次数
	WinRounds                int32   `bson:"win_rounds"`                                                  // 触发策略时逃跑的局数
	LoseRounds               int     `json:"loseRounds" bson:"lose_rounds"`                               // 触发策略时未能逃跑的局数
	MaximumEscapeMultiple    float64 `json:"maximumEscapeMultiple" bson:"maximum_escape_multiple"`        // 策略中最高逃跑倍数
	MinimumEscapeMultiple    float64 `json:"minimumEscapeMultiple" bson:"minimum_escape_multiple"`        // 策略中最低逃跑倍数
	SumEscapeMulpitle        float64 `json:"sumEscapeMultiple" bson:"sum_escape_multiple"`                // 策略中逃跑倍数和
	AverageEscapeMultiple    float64 `json:"averageEscapeMultiple" bson:"average_escape_multiple"`        // 策略中平均逃跑倍数
	WinMulpitleSum           float64 `bson:"win_mulpitle_sum"`                                            // 策略中逃跑倍数和
	Bets                     int64   `bson:"bets"`                                                        // 策略中总打码量
	BetRounds                int32   `bson:"bet_rounds"`                                                  // 下注局数
	MaximumEscapeMultipleBet int64   `json:"maximumEscapeMultipleBet" bson:"maximum_escape_multiple_bet"` // 策略中最高逃跑倍数时打码量
	MinimumEscapeMultipleBet int64   `json:"minimumEscapeMultipleBet" bson:"minimum_escape_multiple_bet"` // 策略中最低逃跑倍数时打码量
	LoseMulpitleSum          float64 `json:"loseMulpitleSum" bson:"lose_mulpitle_sum"`                    // 触发策略未能逃跑的爆炸倍数和
	WinBets                  int64   `bson:"win_bets"`                                                    // 触发策略时逃跑的总打码量
	LoseBets                 int64   `json:"loseBets" bson:"lose_bets"`                                   // 触发策略时未能逃跑的总打码量
	Wins                     int64   `bson:"wins"`                                                        // 胜局赢钱金额
	Loses                    int64   `bson:"loses"`                                                       // 败局输钱金额
	WinEscapeMultiple        float64 `json:"winEscapeMultiple" bson:"win_escape_multiple"`                // 胜局平均逃脱倍数
	Cash                     int64   `bson:"cash"`                                                        // 净赢
	AD_BundleId              string  `bson:"ad__bundle_id"`                                               // 渠道
	Channel1                 string  `bson:"channel1"`                                                    // 渠道别名
	RegistArea               int     `bson:"regist_area"`                                                 // 账号类型 ab测试 0:A 1:B 2:C
	AllRounds                int32   `bson:"all_rounds"`                                                  // 总局数
	RoundBetAvg              int64   `bson:"round_bet_avg"`                                               // 局均码

	No          string `bson:"-"` // 序号
	FNickname   string `bson:"-"` // 玩家昵称
	FCtime      string `bson:"-"` // 注册时间
	FLoginTime  string `bson:"-"` // 最后登录时间
	FLiveDays   int    `bson:"-"` // 存活天数
	FLoseDays   int    `bson:"-"` // 流失天数
	FRegistArea string `bson:"-"` // 账号类型
	FUserType   string `bson:"-"` // 玩家类型

	FitNumber                 int     `bson:"fit_count"`     // 适配策略的人数
	TriggerNumber             int     `bson:"trigger_count"` // 触发策略的人
	FitTriggerRate            float64 `bson:"-"`             // 适配人群触发率 触发策略人数/玩crsah的人里玩适配策略的人数
	TriggerRate               float64 `bson:"-"`             // 策略有效次数占总触发次数比 触发策略有效次数/触发策略总次数
	PerCapitaRate             float64 `bson:"-"`             // 策略目的达成率（人均有效次数） 触发策略有效次数/触发策略人数
	ExpiredRate               float64 `bson:"-"`             // 失效局占比 触发策略时未能逃跑的局数/触发策略局数
	TriggerRoundsAvg          float64 `bson:"-"`             // 每次触发平均所用局数 触发策略局数/触发策略总次数
	ValidRoundsAvg            float64 `bson:"-"`             // 每次有效触发平均所用局数 触发策略局数/触发策略有效次数
	FBets                     float64 `bson:"-"`             // 总打码量
	FRoundBetsAvg             float64 `bson:"-"`             // 局均打码量
	FMaximumEscapeMultipleBet float64 `bson:"-"`             // 策略中最高逃跑倍数时打码量
	FMinimumEscapeMultipleBet float64 `bson:"-"`             // 策略中最低逃跑倍数时打码量
	FLoseMultipleAvg          float64 `bson:"-"`             // 触发策略时未能逃跑的平均爆炸倍数
	FLoseBets                 float64 `bson:"-"`             // 触发策略时未能逃跑的总打码量
	FLoseBetsAvg              float64 `bson:"-"`             // 触发策略时未能逃跑的局均打码量
	FWins                     float64 `bson:"-"`             // 策略中总赢
	FLoses                    float64 `bson:"-"`             // 策略中总输
	FCash                     float64 `bson:"-"`             // 策略中净赢
	AvgEscapeMultiple         float64 `bson:"-"`             // 策略中平均逃跑倍数
}

// Crash欲薅无门
type CrashYHWMStat struct {
	Id                string  `bson:"_id"` // date-userid
	Date              int64   `bson:"date"`
	DateStr           string  `bson:"date_str"`
	Userid            string  `bson:"userid"`
	IsFit             bool    `json:"isFit" bson:"is_fit"`                          // 是否是适配策略的人
	IsTrigger         bool    `json:"isTrigger" bson:"is_trigger"`                  // 是否是触发策略的人
	TiggerTimes       int     `json:"tiggerTimes" bson:"tigger_times"`              // 触发策略有效次数
	AllTiggerTimes    int     `json:"allTiggerTimes" bson:"all_tigger_times"`       // u总 累计触发次数
	AllEvoTimes       int     `json:"allEvoTimes" bson:"all_evo_times"`             // 触发策略局数
	BurstNumber       int     `json:"burstNumber" bson:"burst_number"`              // 瞬爆发生局数
	BurstAmount       int64   `json:"burstAmount" bson:"burst_amount"`              // 瞬爆收割金额 策略触发后，因为瞬爆导致玩家输钱的总金额
	TotalReapAmount   int64   `json:"totalReapAmount" bson:"total_reap_amount"`     // 策略中总割金额 策略开始触发到策略结束进入冷却期间，玩家输钱的总金额
	WinRounds         int     `bson:"win_rounds"`                                   // 策略局中玩家胜局数
	LoseRounds        int     `json:"loseRounds" bson:"lose_rounds"`                // 策略局中玩家败局数
	Bets              int64   `bson:"bets"`                                         // 总打码量
	BetRounds         int32   `bson:"bet_rounds"`                                   // 下注局数
	WinBets           int64   `bson:"win_bets"`                                     // 赢局打码量
	LoseBets          int64   `json:"loseBets" bson:"lose_bets"`                    // 输局打码量
	Wins              int64   `bson:"wins"`                                         // 胜局赢钱金额
	Loses             int64   `bson:"loses"`                                        // 败局输钱金额
	Cash              int64   `bson:"cash"`                                         // 净赢
	WinMulpitleSum    float64 `bson:"win_mulpitle_sum"`                             // 胜局逃跑倍数和
	WinEscapeMultiple float64 `json:"winEscapeMultiple" bson:"win_escape_multiple"` // 胜局平均逃脱倍数
	RoundBetAvg       int64   `bson:"round_bet_avg"`                                // 局均码
	No                string  `bson:"-"`                                            // 序号
	FNickname         string  `bson:"-"`                                            // 玩家昵称
	FCtime            string  `bson:"-"`                                            // 注册时间
	FLoginTime        string  `bson:"-"`                                            // 最后登录时间
	FLiveDays         int     `bson:"-"`                                            // 存活天数
	FLoseDays         int     `bson:"-"`                                            // 流失天数
	FRegistArea       string  `bson:"-"`                                            // 账号类型
	FUserType         string  `bson:"-"`                                            // 玩家类型

	AllNumber          int64   `bson:"all_number"`    // 总玩人数
	FitNumber          int     `bson:"fit_count"`     // 适配策略的人数
	FitRate            float64 `bson:"-"`             // 适配人群占比 玩crash的人里适配策略的人数/筛选条件下总玩crash人数
	TriggerNumber      int     `bson:"trigger_count"` // 触发策略的人
	FitTriggerRate     float64 `bson:"-"`             // 适配人群触发率 触发策略人数/玩crsah的人里玩适配策略的人数
	TriggerRate        float64 `bson:"-"`             // 策略有效次数占总触发次数比 触发策略有效次数/触发策略总次数
	PerCapitaRate      float64 `bson:"-"`             // 策略目的达成率（人均有效次数） 触发策略有效次数/触发策略人数
	BurstRate          float64 `bson:"-"`             // 瞬爆发生率 瞬爆发生局数/触发策略局数
	ValidRoundsAvg     float64 `bson:"-"`             // 每次有效触发平均所用局数 触发策略局数/触发策略有效次数
	ValidBurstAvg      float64 `bson:"-"`             // 每次有效触发平均所用瞬爆数 瞬爆发生局数/触发策略有效次数
	FBurstAmount       float64 `bson:"-"`             // 瞬爆收割金额
	FTotalReapAmount   float64 `bson:"-"`             // 策略中总割金额
	BurstReapRats      float64 `bson:"-"`             // 瞬爆收割贡献率 瞬爆收割金额/策略中收割金额
	BurstReapAmountAvg float64 `bson:"-"`             // 瞬爆人均收割金额 瞬爆收割金额/策略触发人数
	PerCapitaHarvest   float64 `bson:"-"`             // 策略中人均收割金额
	PlayerWinRatio     float64 `bson:"-"`             // 策略局中玩家胜率 策略局中胜局数/触发策略局数
	FRoundBetsAvg      float64 `bson:"-"`             // 策略局中玩家局均打码
	FWinRoundBetsAvg   float64 `bson:"-"`             // 策略局中玩家胜局均打码
	FLoseRoundBetsAvg  float64 `bson:"-"`             // 策略局中玩家败局均打码
	FBets              float64 `bson:"-"`             // 总打码量
	FWins              float64 `bson:"-"`             // 策略中总赢
	FLoses             float64 `bson:"-"`             // 策略中总输
	FCash              float64 `bson:"-"`             // 策略中净赢
}

// Crash起死回生
type CrashQSHSStat struct {
	Id                    string   `bson:"_id"` // date-userid
	Date                  int64    `bson:"date"`
	DateStr               string   `bson:"date_str"`
	Userid                string   `bson:"userid"`
	AllinNumber           int64    `json:"allinNumber" bson:"allin_number"`                       // allin局数
	AllinBets             int64    `json:"allinBets" bson:"allin_bets"`                           // allin局打码量
	AllEvoTimes           int      `json:"allEvoTimes" bson:"all_evo_times"`                      // 触发策略局数
	TiggerBets            int64    `json:"tiggerBets" bson:"tigger_bets"`                         // 触发策略局的打码量
	WinRounds             int32    `bson:"win_rounds"`                                            // 触发策略时逃跑的局数
	LoseRounds            int      `json:"loseRounds" bson:"lose_rounds"`                         // 触发策略时未能逃跑的局数
	TiggerWinMulpitleSum  float64  `json:"tiggerWinMulpitleSum" bson:"tigger_win_mulpitle_sum"`   // 策略中逃跑倍数和
	MaximumEscapeMultiple float64  `json:"maximumEscapeMultiple" bson:"maximum_escape_multiple"`  // 策略中最高逃跑倍数
	LoseMaxEscapeMultiple float64  `json:"loseMaxEscapeMultiple" bson:"lose_max_escape_multiple"` // 触发后未能逃跑局的最高爆炸倍数
	LoseEscapeMultiple    float64  `json:"loseEscapeMultiple" bson:"lose_escape_multiple"`        // 触发后未能逃跑局的爆炸倍和
	Bets                  int64    `bson:"bets"`                                                  // 总打码量
	AllRounds             int32    `bson:"all_rounds"`                                            // 总局数
	AllWinBets            int64    `json:"allWinBets" bson:"all_win_bets"`                        // 总赢
	AllLoseBets           int64    `json:"allLoseBets" bson:"all_lose_bets"`                      // 总输
	WinBets               int64    `bson:"win_bets"`                                              // 赢局打码量
	LoseBets              int64    `json:"loseBets" bson:"lose_bets"`                             // 输局打码量
	Wins                  int64    `bson:"wins"`                                                  // 胜局赢钱金额
	Loses                 int64    `bson:"loses"`                                                 // 败局输钱金额
	Cash                  int64    `bson:"cash"`                                                  // 净赢
	WithdrawAmount        int64    `json:"withdrawAmount" bson:"withdraw_amount"`
	PayAmount             int64    `json:"payAmount" bson:"pay_amount"`
	CarryAmount           int64    `json:"carryAmount" bson:"carry_amount"`
	AD_BundleId           string   `bson:"ad__bundle_id"` // 渠道
	Channel1              string   `bson:"channel1"`      // 渠道别名
	RegistArea            int      `bson:"regist_area"`   // 账号类型 ab测试 0:A 1:B 2:C
	RoundBetAvg           int64    `bson:"round_bet_avg"` // 局均码
	Userids               []string `bson:"userids"`       // 汇总时的userid列表

	No          string `bson:"-"` // 序号
	FNickname   string `bson:"-"` // 玩家昵称
	FCtime      string `bson:"-"` // 注册时间
	FLoginTime  string `bson:"-"` // 最后登录时间
	FLiveDays   int    `bson:"-"` // 存活天数
	FLoseDays   int    `bson:"-"` // 流失天数
	FRegistArea string `bson:"-"` // 账号类型
	FUserType   string `bson:"-"` // 玩家类型

	AllinBetAvg           float64 `bson:"-"` // allin局平均打码量
	TiggerBetAvg          float64 `bson:"-"` // 触发策略局的平均打码量
	TiggerRate            float64 `bson:"-"` // 触发率=触发策略局数/allin局数
	ExpiredRate           float64 `bson:"-"` // 策略失效率=触发后未能逃跑的局数/触发策略局数
	TiggerMulpitle        float64 `bson:"-"` // 触发后的平均逃跑倍数
	LoseEscapeMultipleAvg float64 `bson:"-"` // 触发后未能逃跑局的平均爆炸倍数

	FWinBets         float64 `bson:"-"` // 赢局打码量
	FLoseBets        float64 `bson:"-"` // 输局打码量
	FWins            float64 `bson:"-"` // 策略局总赢
	FLoses           float64 `bson:"-"` // 策略局总输
	FCash            float64 `bson:"-"` // 策略局净赢
	FBets            float64 `bson:"-"` // crash总码量
	FRoundBetAvg     float64 `bson:"-"` // crash局均码
	TotalRewardsRate string  `bson:"-"` // 玩家总返奖率=（提现金额+携带金额）/充值金额
	UserRewardRate   string  `bson:"-"` // 玩家crash返奖率 = ｜crash中总赢｜/｜crash中总输｜
}

// 奖池风控
type CrashJCFKStat struct {
	Id                    string   `bson:"_id"` // date-userid
	Date                  int64    `bson:"date"`
	DateStr               string   `bson:"date_str"`
	Userid                string   `bson:"userid"`
	TriggerTimes          int      `json:"triggerTimes" bson:"trigger_times"`                    // jackpot次数
	TriggerAmount         int64    `json:"triggerAmount" bson:"trigger_amount"`                  // 获得jackpot金额
	AllEvoTimes           int64    `json:"allEvoTimes" bson:"all_evo_times"`                     // 触发策略局数
	TiggerBets            int64    `json:"tiggerBets" bson:"tigger_bets"`                        // 触发策略局的打码量
	MaximumEscapeMultiple float64  `json:"maximumEscapeMultiple" bson:"maximum_escape_multiple"` // 触发后最高爆炸倍数
	TiggerMulpitleSum     float64  `json:"tiggerMulpitleSum" bson:"tigger_mulpitle_sum"`         // 触发后爆炸倍数和
	WinRounds             int32    `bson:"win_rounds"`                                           // 触发策略时逃跑的局数
	WinMulpitleSum        float64  `json:"winMulpitleSum" bson:"win_mulpitle_sum"`               // 触发后逃跑倍数和
	LoseRounds            int      `json:"loseRounds" bson:"lose_rounds"`                        // 触发策略时未能逃跑的局数
	LoseMulpitleSum       float64  `json:"loseMulpitleSum" bson:"lose_mulpitle_sum"`             // 触发后未能逃跑局的爆炸倍数和
	TriggerRounds         int      `json:"triggerRounds" bson:"trigger_rounds"`                  // 触发后获得jackpot局数
	WinBets               int64    `bson:"win_bets"`                                             // 赢局打码量
	LoseBets              int64    `json:"loseBets" bson:"lose_bets"`                            // 输局打码量
	Wins                  int64    `bson:"wins"`                                                 // 胜局赢钱金额
	Loses                 int64    `bson:"loses"`                                                // 败局输钱金额
	Cash                  int64    `bson:"cash"`                                                 // 净赢
	Bets                  int64    `bson:"bets"`                                                 // 总打码量
	AllRounds             int32    `bson:"all_rounds"`                                           // 总局数
	AllWinBets            int64    `json:"allWinBets" bson:"all_win_bets"`                       // 总赢
	AllLoseBets           int64    `json:"allLoseBets" bson:"all_lose_bets"`                     // 总输
	WithdrawAmount        int64    `json:"withdrawAmount" bson:"withdraw_amount"`
	PayAmount             int64    `json:"payAmount" bson:"pay_amount"`
	CarryAmount           int64    `json:"carryAmount" bson:"carry_amount"`
	AD_BundleId           string   `bson:"ad__bundle_id"` // 渠道
	Channel1              string   `bson:"channel1"`      // 渠道别名
	RegistArea            int      `bson:"regist_area"`   // 账号类型 ab测试 0:A 1:B 2:C
	RoundBetAvg           int64    `bson:"round_bet_avg"` // 局均码
	Userids               []string `bson:"userids"`       // 汇总时的userid列表

	No                   string  `bson:"-"` // 序号
	FNickname            string  `bson:"-"` // 玩家昵称
	FCtime               string  `bson:"-"` // 注册时间
	FLoginTime           string  `bson:"-"` // 最后登录时间
	FLiveDays            int     `bson:"-"` // 存活天数
	FLoseDays            int     `bson:"-"` // 流失天数
	FRegistArea          string  `bson:"-"` // 账号类型
	FUserType            string  `bson:"-"` // 玩家类型
	FTriggerAmount       float64 `bson:"-"` // 获得jackpot金额
	TiggerBetAvg         float64 `bson:"-"` // 触发策略局的平均打码量
	TiggerMulpitleSumAvg float64 `bson:"-"` // 触发后平均爆炸倍数
	WinMulpitleSumAvg    float64 `bson:"-"` // 触发后平均逃跑倍数
	LoseMulpitleSumAvg   float64 `bson:"-"` // 触发后未能逃跑局的平均爆炸倍数
	ExpiredRate          float64 `bson:"-"` // 策略失效率=触发后获得jackpot局数/触发策略局数
	FWinBets             float64 `bson:"-"` //
	FLoseBets            float64 `bson:"-"` //
	FWins                float64 `bson:"-"` // 策略局总赢
	FLoses               float64 `bson:"-"` // 策略局总输
	FCash                float64 `bson:"-"` // 策略局净赢
	FBets                float64 `bson:"-"` // crash总码量
	FRoundBetAvg         float64 `bson:"-"` // crash局均码
	TotalRewardsRate     string  `bson:"-"` // 玩家总返奖率=（提现金额+携带金额）/充值金额
	UserRewardRate       string  `bson:"-"` // 玩家crash返奖率 = ｜crash中总赢｜/｜crash中总输｜
}

// 冒险奖励
type CrashMXJLStat struct {
	Id                    string   `bson:"_id"` // date-userid
	Date                  int64    `bson:"date"`
	DateStr               string   `bson:"date_str"`
	Userid                string   `bson:"userid"`
	Bets                  int64    `bson:"bets"`                                                 // 总打码量
	AllRounds             int32    `bson:"all_rounds"`                                           // 总局数
	AllWinRounds          int32    `bson:"all_win_rounds"`                                       // 总赢局数
	AllWinBets            int64    `json:"allWinBets" bson:"all_win_bets"`                       // 总赢
	AllLoseBets           int64    `json:"allLoseBets" bson:"all_lose_bets"`                     // 总输
	AllEvoTimes           int64    `json:"allEvoTimes" bson:"all_evo_times"`                     // 触发策略局数
	TiggerBets            int64    `json:"tiggerBets" bson:"tigger_bets"`                        // 触发策略局的打码量
	MaximumEscapeMultiple float64  `json:"maximumEscapeMultiple" bson:"maximum_escape_multiple"` // 触发后最高爆炸倍数
	TiggerMulpitleSum     float64  `json:"tiggerMulpitleSum" bson:"tigger_mulpitle_sum"`         // 触发后爆炸倍数和
	WinRounds             int32    `bson:"win_rounds"`                                           // 触发策略时逃跑的局数
	WinMaxMulpitle        float64  `json:"winMaxMulpitle" bson:"win_max_mulpitle"`               // 触发后最高逃跑倍数
	WinMulpitleSum        float64  `json:"winMulpitleSum" bson:"win_mulpitle_sum"`               // 触发后逃跑倍数和
	LoseRounds            int      `json:"loseRounds" bson:"lose_rounds"`                        // 触发策略时未能逃跑的局数
	LoseMulpitleSum       float64  `json:"loseMulpitleSum" bson:"lose_mulpitle_sum"`             // 触发后未能逃跑局的爆炸倍数和
	WinBets               int64    `bson:"win_bets"`                                             // 策略赢局打码量
	LoseBets              int64    `json:"loseBets" bson:"lose_bets"`                            // 策略输局打码量
	Wins                  int64    `bson:"wins"`                                                 // 胜局赢钱金额
	Loses                 int64    `bson:"loses"`                                                // 败局输钱金额
	Cash                  int64    `bson:"cash"`                                                 // 策略净赢
	WithdrawAmount        int64    `json:"withdrawAmount" bson:"withdraw_amount"`
	PayAmount             int64    `json:"payAmount" bson:"pay_amount"`
	CarryAmount           int64    `json:"carryAmount" bson:"carry_amount"`
	AD_BundleId           string   `bson:"ad__bundle_id"` // 渠道
	Channel1              string   `bson:"channel1"`      // 渠道别名
	RegistArea            int      `bson:"regist_area"`   // 账号类型 ab测试 0:A 1:B 2:C
	RoundBetAvg           int64    `bson:"round_bet_avg"` // 局均码
	Userids               []string `bson:"userids"`       // 汇总时的userid列表

	No                   string  `bson:"-"` // 序号
	FNickname            string  `bson:"-"` // 玩家昵称
	FCtime               string  `bson:"-"` // 注册时间
	FLoginTime           string  `bson:"-"` // 最后登录时间
	FLiveDays            int     `bson:"-"` // 存活天数
	FLoseDays            int     `bson:"-"` // 流失天数
	FRegistArea          string  `bson:"-"` // 账号类型
	FUserType            string  `bson:"-"` // 玩家类型
	FBets                float64 `bson:"-"` // 总打码
	FRoundBetAvg         float64 `bson:"-"` // 局均打码
	FTiggerBets          float64 `bson:"-"` // 策略局打码
	TiggerBetAvg         float64 `bson:"-"` // 触发策略局的平均打码量
	TiggerMulpitleSumAvg float64 `bson:"-"` // 触发后平均爆炸倍数
	WinMulpitleSumAvg    float64 `bson:"-"` // 触发后平均逃跑倍数
	LoseMulpitleSumAvg   float64 `bson:"-"` // 触发后未能逃跑局的平均爆炸倍数
	SuccessRate          string  `bson:"-"` // 成功逃跑率=触发后成功逃跑的局数/触发策略局数
	FWinBets             float64 `bson:"-"` // 策略赢局打码量
	FLoseBets            float64 `bson:"-"` // 策略输局打码量
	FWins                float64 `bson:"-"` // 策略局总赢
	FLoses               float64 `bson:"-"` // 策略局总输
	FCash                float64 `bson:"-"` // 策略局净赢
	WinRate              string  `bson:"-"` // 总胜率 = 所有赢钱局数/总下注局数
	StrategyWinRate      string  `bson:"-"` // 策略局中胜率 = 策略局中赢钱局数/策略局中总局数
	StrategyRewardsRate  string  `bson:"-"` // 策略局中返奖率 = 策略中总赢/策略中总输
	TotalRewardsRate     string  `bson:"-"` // crash总返奖率 = 总赢/总输
	PlayerRewardRate     string  `bson:"-"` // 玩家大盘返奖率 = （提现金额+携带金额）/充值金额
}

// 人狂有祸
type CrashRKYSStat struct {
	Id                    string   `bson:"_id"` // date-userid
	Date                  int64    `bson:"date"`
	DateStr               string   `bson:"date_str"`
	Userid                string   `bson:"userid"`
	Bets                  int64    `bson:"bets"`                                                 // 总打码量
	AllRounds             int32    `bson:"all_rounds"`                                           // 总局数
	AllWinBets            int64    `json:"allWinBets" bson:"all_win_bets"`                       // 总赢
	AllLoseBets           int64    `json:"allLoseBets" bson:"all_lose_bets"`                     // 总输
	AllEvoTimes           int64    `json:"allEvoTimes" bson:"all_evo_times"`                     // 触发策略局数
	TiggerBets            int64    `json:"tiggerBets" bson:"tigger_bets"`                        // 触发策略局的打码量
	MaximumEscapeMultiple float64  `json:"maximumEscapeMultiple" bson:"maximum_escape_multiple"` // 触发后最高爆炸倍数
	TiggerMulpitleSum     float64  `json:"tiggerMulpitleSum" bson:"tigger_mulpitle_sum"`         // 触发后爆炸倍数和
	WinRounds             int32    `bson:"win_rounds"`                                           // 触发策略时逃跑的局数
	WinMaxMulpitle        float64  `json:"winMaxMulpitle" bson:"win_max_mulpitle"`               // 触发后最高逃跑倍数
	WinMulpitleSum        float64  `json:"winMulpitleSum" bson:"win_mulpitle_sum"`               // 触发后逃跑倍数和
	LoseRounds            int      `json:"loseRounds" bson:"lose_rounds"`                        // 触发策略时未能逃跑的局数
	LoseMulpitleSum       float64  `json:"loseMulpitleSum" bson:"lose_mulpitle_sum"`             // 触发后未能逃跑局的爆炸倍数和
	RFRounds              int64    `json:"rfRounds" bson:"rf_rounds"`                            // 肥码收割局数
	RFBets                int64    `json:"rfBets" bson:"rf_bets"`                                // 肥码收割局总码量
	RFLoseRounds          int64    `json:"rfLoseRounds" bson:"rf_lose_rounds"`                   // 肥码局输
	RFReapBets            int64    `json:"rfReapBets" bson:"rf_reap_bets"`                       // 肥码收割金额
	WinBets               int64    `bson:"win_bets"`                                             // 策略赢局打码量
	LoseBets              int64    `json:"loseBets" bson:"lose_bets"`                            // 策略输局打码量
	Wins                  int64    `bson:"wins"`                                                 // 胜局赢钱金额
	Loses                 int64    `bson:"loses"`                                                // 败局输钱金额
	Cash                  int64    `bson:"cash"`                                                 // 策略净赢
	WithdrawAmount        int64    `json:"withdrawAmount" bson:"withdraw_amount"`
	PayAmount             int64    `json:"payAmount" bson:"pay_amount"`
	CarryAmount           int64    `json:"carryAmount" bson:"carry_amount"`
	AD_BundleId           string   `bson:"ad__bundle_id"` // 渠道
	Channel1              string   `bson:"channel1"`      // 渠道别名
	RegistArea            int      `bson:"regist_area"`   // 账号类型 ab测试 0:A 1:B 2:C
	RoundBetAvg           int64    `bson:"round_bet_avg"` // 局均码
	Userids               []string `bson:"userids"`       // 汇总时的userid列表
	No                    string   `bson:"-"`             // 序号
	FNickname             string   `bson:"-"`             // 玩家昵称
	FCtime                string   `bson:"-"`             // 注册时间
	FLoginTime            string   `bson:"-"`             // 最后登录时间
	FLiveDays             int      `bson:"-"`             // 存活天数
	FLoseDays             int      `bson:"-"`             // 流失天数
	FRegistArea           string   `bson:"-"`             // 账号类型
	FUserType             string   `bson:"-"`             // 玩家类型
	FBets                 float64  `bson:"-"`             // 总打码
	FRoundBetAvg          float64  `bson:"-"`             // 局均打码
	FTiggerBets           float64  `bson:"-"`             // 策略局打码
	TiggerBetAvg          float64  `bson:"-"`             // 触发策略局的平均打码量
	TiggerMulpitleSumAvg  float64  `bson:"-"`             // 触发后平均爆炸倍数
	WinMulpitleSumAvg     float64  `bson:"-"`             // 触发后平均逃跑倍数
	LoseMulpitleSumAvg    float64  `bson:"-"`             // 触发后未能逃跑局的平均爆炸倍数
	SuccessRate           float64  `bson:"-"`             // 成功逃跑率=触发后成功逃跑的局数/触发策略局数
	FRFBets               float64  `bson:"-"`             // 肥码收割局总码量
	FRFReapBetAvg         float64  `bson:"-"`             // 肥码收割局均码量
	FRFReapBets           float64  `bson:"-"`             // 肥码收割金额
	RFRate                float64  `bson:"-"`             // 肥码局占比 = r肥/r策
	RFReapRate            float64  `bson:"-"`             // 肥码收割成功率 = r肥割/r肥
	FWinBets              float64  `bson:"-"`             //
	FLoseBets             float64  `bson:"-"`             //
	FWins                 float64  `bson:"-"`             // 策略局总赢
	FLoses                float64  `bson:"-"`             // 策略局总输
	FCash                 float64  `bson:"-"`             // 策略局净赢
	WinRate               float64  `bson:"-"`             // 总胜率 = 所有赢钱局数/总下注局数
	StrategyWinRate       float64  `bson:"-"`             // 策略局中胜率 = 策略局中赢钱局数/策略局中总局数
	StrategyRewardsRate   float64  `bson:"-"`             // 策略局中返奖率 = 策略中总赢/策略中总输
	TotalRewardsRate      string   `bson:"-"`             // crash总返奖率 = 总赢/总输
	PlayerRewardRate      string   `bson:"-"`             // 玩家大盘返奖率 = （提现金额+携带金额）/充值金额
}

// tp 统计
type TPPlayerStat struct {
	Id                   string      `bson:"_id"` // date-userid
	Date                 int64       `bson:"date"`
	DateStr              string      `bson:"date_str"`
	Userid               string      `bson:"userid"`
	NewReg               bool        `bson:"new_reg"`                 // 是新顾客
	AllRounds            int32       `bson:"all_rounds"`              // 总局数
	GameTimes            int64       `bson:"game_times"`              // 游戏总时长(秒)
	Bets                 int64       `bson:"bets"`                    // 总打码量
	BetRounds            int32       `bson:"bet_rounds"`              // 下注局数
	WinRounds            int32       `bson:"win_rounds"`              // 胜利局数
	LoseRounds           int32       `bson:"lose_rounds"`             // 失败局数
	TieRounds            int32       `bson:"tie_rounds"`              // 和局数
	WinBets              int64       `bson:"win_bets"`                // 胜局打码量
	LoseBets             int64       `bson:"lose_bets"`               // 败局打码量
	TieBets              int64       `bson:"tie_bets"`                // 和局打码量
	Wins                 int64       `bson:"wins"`                    // 胜局赢钱金额
	Loses                int64       `bson:"loses"`                   // 败局输钱金额
	Cash                 int64       `bson:"cash"`                    // 净赢
	MulpitleSum          float64     `bson:"mulpitle_sum"`            // 所有局总爆炸倍数和
	Mulpitles            [][]float64 `bson:"mulpitles"`               // 局数爆炸倍数
	WinEscapeMulpitleSum float64     `bson:"win_escape_mulpitle_sum"` // 赢局逃脱倍数和
	WinEscapeMulpitles   [][]float64 `bson:"win_escape_mulpitles"`    // 赢局逃脱倍数
	WinMulpitleSum       float64     `bson:"win_mulpitle_sum"`        // 赢局爆炸倍数和
	WinMulpitles         [][]float64 `bson:"win_mulpitles"`           // 赢局数爆炸倍数
	LoseMulpitleSum      float64     `bson:"lose_mulpitle_sum"`       // 输局爆炸倍数和
	LoseMulpitles        [][]float64 `bson:"lose_mulpitles"`          // 输局爆炸倍数,算中位数
	RoundBetAvg          int64       `bson:"round_bet_avg"`           // 局均码
	AD_BundleId          string      `bson:"ad__bundle_id"`           // 渠道
	Channel1             string      `bson:"channel1"`                // 渠道别名
	RegistArea           int         `bson:"regist_area"`             // 账号类型 ab测试 0:A 1:B 2:C
	Money                uint32      `bson:"money"`                   // 充值总金额(分)
	ObserveRounds        int32       `bson:"observe_rounds"`          // 观察局数
	Players              int32       `bson:"players"`                 // 总玩crash人数
	NewPlayers           int32       `bson:"new_players"`             // 新顾客玩crash人数
	OldPlayers           int32       `bson:"old_players"`             // 老顾客玩crash人数
	AllRoundsList        []int32     `bson:"all_rounds_list"`         // 玩家局数list
	Userids              []string    `bson:"userids"`                 // 汇总时的userid列表

	No                       string `bson:"-"` // 序号
	FNickname                string `bson:"-"` // 总打码量
	FCtime                   string `bson:"-"` // 注册时间
	FLoginTime               string `bson:"-"` // 最后登录时间
	FLiveDays                int    `bson:"-"` // 存活天数
	FLoseDays                int    `bson:"-"` // 流失天数
	FRegistArea              string `bson:"-"` // 账号类型
	FUserType                string `bson:"-"` // 玩家类型
	FGameTimes               string `bson:"-"` // 游戏总时长(分)
	FBets                    string `bson:"-"` // 总打码量
	FRoundBetsAvg            string `bson:"-"` // 局均打码量
	FWinRate                 string `bson:"-"` // 胜率
	FRebateRate              string `bson:"-"` // 返奖率
	FWinBets                 string `bson:"-"` // 胜局打码量
	FLoseBets                string `bson:"-"` // 败局打码量
	FWinBetsAvg              string `bson:"-"` // 胜局均码量
	FLoseBetsAvg             string `bson:"-"` // 败局均码量
	FWins                    string `bson:"-"` // 胜局赢钱金额
	FLoses                   string `bson:"-"` // 败局输钱金额
	FCash                    string `bson:"-"` // 净赢
	FWinsAvg                 string `bson:"-"` // 胜局均盈利
	FLosesAvg                string `bson:"-"` // 败局均输额
	FCashAvg                 string `bson:"-"` // 局均净赢
	FMulpitleAvg             string `bson:"-"` // 局均实际爆炸倍数
	FMulpitleMedian          string `bson:"-"` // 每局实际爆炸倍数中位数
	FWinEscapeMulpitleMax    string `bson:"-"` // 赢局最高逃脱倍数
	FWinEscapeMulpitleAvg    string `bson:"-"` // 赢局局均逃脱倍数
	FWinEscapeMulpitleMedian string `bson:"-"` // 赢局逃脱倍数中位数
	FWinMulpitleAvg          string `bson:"-"` // 赢局局均实际爆炸倍数
	FWinMulpitleMedian       string `bson:"-"` // 赢局实际爆炸倍数中位数
	FMulpitleDiff            string `bson:"-"` // 倍差空间
	FLoseMulpitleAvg         string `bson:"-"` // 输局局均实际爆炸倍数
	FLoseMulpitleMedian      string `bson:"-"` // 输局实际爆炸倍数中位数
	FWinEscapeRate           string `bson:"-"` // 赢逃输爆比
	FWinLoseRate             string `bson:"-"` // 赢输爆倍比
	ObserveRoundsAvg         string `bson:"-"` // 局均观察次数
	ObserveRoundsRate        string `bson:"-"` // 观察局数占比
	FMoney                   string `bson:"-"` // 总充值
	FCashOut                 string `bson:"-"` // 总提现
	FWinMoney                string `bson:"-"` // 总赢

	SMoney   int   `bson:"-"` // 总汇充值
	SCashOut int   `bson:"-"` // 总汇提现
	SDiamond int64 `bson:"-"` // 总汇携带分

	FLoginedUsers       int    `bson:"-"` // 日活
	FLoginedNewUsers    int    `bson:"-"` // 新用户日活
	FLoginedOldUsers    int    `bson:"-"` // 老用户日活
	FPlayerRate         string `bson:"-"` // 玩crash人数占日活比
	FNewPlayerRate      string `bson:"-"` // 新顾客玩crash占新顾客日活比
	FOldPlayerRate      string `bson:"-"` // 老顾客玩crash占老顾客日活比
	FPlayerRoundsAvg    string `bson:"-"` // 人均局数
	FPlayerRoundsMedian int32  `bson:"-"` // 局数中位数
	FGameTimesAvg       string `bson:"-"` // 游戏局均时长（min)
	FGameTimesPlayerAvg string `bson:"-"` // 游戏人均时长（min)
	FPlayerBetsAvg      string `bson:"-"` // 人均打码
}

type TPPlayerRoomShow struct {
	Userid        string  `bson:"userid"`
	AllRounds     int64   `json:"allRounds" bson:"all_rounds"` // 局数
	Bets          int64   `bson:"bets"`                        // 总下注
	FBets         float64 `bson:"-"`
	RoundBetAvg   float64 `json:"roundBetAvg" bson:"round_bet_avg"` // 场均下注
	TotalProfit   float64 `json:"totalProfit" bson:"total_profit"`  // 总盈亏
	GameTime      int64
	WinBets       int64         `bson:"-"`
	RoomInfo      []*TPRoomStat // 房间数据
	TurnInfo      []*TPGameTurn // 轮次数据
	WinInfo       []*TPGameWin  // 大赢局数据
	LoseInfo      []*TPGameLose // 大输局数据
	TwoRateInfo   []*TPGameRate
	ThreeRateInfo []*TPGameRate
	FourRateInfo  []*TPGameRate
	FiveRateInfo  []*TPGameRate
}

type TPProfitShow struct {
	Dates    []interface{}
	TP       []interface{}
	Pay      []interface{}
	Withdraw []interface{}
	InPay    []interface{}
	UpRoom   []interface{}
	DownRoom []interface{}
}

// TP玩家房间数据
type TPRoomStat struct {
	Roomid         string  `json:"roomid" bson:"roomid"`                   // 房间ID
	AllRounds      int64   `json:"allRounds" bson:"all_rounds"`            // 局数
	Bets           int64   `bson:"bets"`                                   // 总下注
	RoundBetAvg    float64 `json:"roundBetAvg" bson:"round_bet_avg"`       // 场均下注
	TotalProfit    float64 `json:"totalProfit" bson:"total_profit"`        // 总盈亏
	UpNumber       int64   `json:"upNumber" bson:"up_number"`              // 主动升到此场次数
	DownNumber     int64   `json:"downNumber" bson:"down_number"`          // 主动降到此场次数
	TwoWinRate     float64 `json:"twoWinRate" bson:"two_win_rate"`         // 2人局胜率
	ThreeWinRate   float64 `json:"threeWinRate" bson:"three_win_rate"`     // 3人局胜率
	FourWinRate    float64 `json:"fourWinRate" bson:"four_win_rate"`       // 4人局胜率
	FiveWinRate    float64 `json:"fiveWinRate" bson:"five_win_rate"`       // 5人局胜率
	TwoBonusRate   float64 `json:"twoBonusRate" bson:"two_bonus_rate"`     // 2人局返奖率
	ThreeBonusRate float64 `json:"threeBonusRate" bson:"three_bonus_rate"` // 3人局返奖率
	FourBonusRate  float64 `json:"fourBonusRate" bson:"four_bonus_rate"`   // 4人局返奖率
	FiveBonusRate  float64 `json:"fiveBonusRate" bson:"five_bonus_rate"`   // 5人局返奖率
	ChipPool       float64 `json:"chipPool" bson:"chip_pool"`              // 筹码池上限
	BZUpBetAvg     float64 `json:"bzUpBetAvg" bson:"bz_up_bet_avg"`        // 大豹子下注平均值
	BZDownBetAvg   float64 `json:"bzDownBetAvg" bson:"bz_down_bet_avg"`    // 小豹子下注平均值
	THSUpBetAvg    float64 `json:"thsUpBetAvg" bson:"ths_up_bet_avg"`      // 大同花顺下注平均值
	THSDownBetAvg  float64 `json:"thsDownBetAvg" bson:"ths_down_bet_avg"`  // 小同花顺下注平均值
	DSZUpBetAvg    float64 `json:"dszUpBetAvg" bson:"dsz_up_bet_avg"`      // 大顺子下注平均值
	DSZDownBetAvg  float64 `json:"dszDownBetAvg" bson:"dsz_down_bet_avg"`  // 小顺子下注平均值
	DTHUpBetAvg    float64 `json:"dthUpBetAvg" bson:"dth_up_bet_avg"`      // 大同花下注平均值
	DTHDownBetAvg  float64 `json:"dthDownBetAvg" bson:"dth_down_bet_avg"`  // 小同花下注平均值
	DDZUpBetAvg    float64 `json:"ddzUpBetAvg" bson:"ddz_up_bet_avg"`      // 大对子下注平均值
	DDZDownBetAvg  float64 `json:"ddzDownBetAvg" bson:"ddz_down_bet_avg"`  // 小对子下注平均值
	DGPUpBetAvg    float64 `json:"dgpUpBetAvg" bson:"dgp_up_bet_avg"`      // 大高牌下注平均值
	DGPDownBetAvg  float64 `json:"dgpDownBetAvg" bson:"dgp_down_bet_avg"`  // 小高牌下注平均值
	FBets          float64 `bson:"-"`
}

// TP期望轮次平均值
type TPGameTurn struct {
	BZUpNumber      int64   `json:"bzUpNumber" bson:"bz_up_number"`       // 大豹子局数
	BZUpRounds      int64   `json:"bzUpRounds" bson:"bz_up_rounds"`       // 大豹子轮次
	BZUpRoundAvg    float64 `json:"bzUpRounds" bson:"bz_up_rounds"`       // 大豹子轮次
	BZDownNumber    int64   `json:"bzDownNumber" bson:"bz_down_number"`   // 小豹子局数
	BZDownRounds    int64   `json:"bzDownRounds" bson:"bz_down_rounds"`   // 小豹子轮次
	BZDownRoundAvg  float64 `json:"bzDownRounds" bson:"bz_down_rounds"`   // 小豹子轮次
	THSUpNumber     int64   `json:"thsUpNumber" bson:"ths_up_number"`     // 大同花顺局数
	THSUpRounds     int64   `json:"thsUpRounds" bson:"ths_up_rounds"`     // 大同花顺轮次
	THSUpRoundAvg   float64 `json:"thsUpRounds" bson:"ths_up_rounds"`     // 大同花顺轮次
	THSDownRoundAvg float64 `json:"thsDownRounds" bson:"ths_down_rounds"` // 小同花顺轮次
	THSDownNumber   int64   `json:"thsDownNumber" bson:"ths_down_number"` // 小同花顺局数
	THSDownRounds   int64   `json:"thsDownRounds" bson:"ths_down_rounds"` // 小同花顺轮次
	DSZUpNumber     int64   `json:"dszUpNumber" bson:"dsz_up_number"`     // 大顺子局数
	DSZUpRounds     int64   `json:"dszUpRounds" bson:"dsz_up_rounds"`     // 大顺子轮次
	DSZUpRoundAvg   float64 `json:"dszUpRounds" bson:"dsz_up_rounds"`     // 大顺子轮次
	DSZDownNumber   int64   `json:"dszDownNumber" bson:"dsz_down_number"` // 小顺子局数
	DSZDownRounds   int64   `json:"dszDownRounds" bson:"dsz_down_rounds"` // 小顺子轮次
	DSZDownRoundAvg float64 `json:"dszDownRounds" bson:"dsz_down_rounds"` // 小顺子轮次
	DTHUpNumber     int64   `json:"dthUpNumber" bson:"dth_up_number"`     // 大同花局数
	DTHUpRounds     int64   `json:"dthUpRounds" bson:"dth_up_rounds"`     // 大同花轮次
	DTHUpRoundAvg   float64 `json:"dthUpRounds" bson:"dth_up_rounds"`     // 大同花轮次
	DTHDownNumber   int64   `json:"dthDownNumber" bson:"dth_down_number"` // 小同花局数
	DTHDownRounds   int64   `json:"dthDownRounds" bson:"dth_down_rounds"` // 小同花轮次
	DTHDownRoundAvg float64 `json:"dthDownRounds" bson:"dth_down_rounds"` // 小同花轮次
	DDZUpNumber     int64   `json:"ddzUpNumber" bson:"ddz_up_number"`     // 大对子局数
	DDZUpRounds     int64   `json:"ddzUpRounds" bson:"ddz_up_rounds"`     // 大对子轮次
	DDZUpRoundAvg   float64 `json:"ddzUpRounds" bson:"ddz_up_rounds"`     // 大对子轮次
	DDZDownNumber   int64   `json:"ddzDownNumber" bson:"ddz_down_number"` // 小对子局数
	DDZDownRounds   int64   `json:"ddzDownRounds" bson:"ddz_down_rounds"` // 小对子轮次
	DDZDownRoundAvg float64 `json:"ddzDownRounds" bson:"ddz_down_rounds"` // 小对子轮次
	DGPUpNumber     int64   `json:"dgpUpNumber" bson:"dgp_up_number"`     // 大高牌局数
	DGPUpRounds     int64   `json:"dgpUpRounds" bson:"dgp_up_rounds"`     // 大高牌轮次
	DGPUpRoundAvg   float64 `json:"dgpUpRounds" bson:"dgp_up_rounds"`     // 大高牌轮次
	DGPDownNumber   int64   `json:"dgpDownNumber" bson:"dgp_down_number"` // 小高牌局数
	DGPDownRounds   int64   `json:"dgpDownRounds" bson:"dgp_down_rounds"` // 小高牌轮次
	DGPDownRoundAvg float64 `json:"dgpDownRounds" bson:"dgp_down_rounds"` // 小高牌轮次
}

// TP 大赢局
type TPGameWin struct {
	BZUpWinRounds    int64 `json:"bzUpWinRounds" bson:"bz_up_win_rounds"`       // 大豹子赢局
	BZDownWinRounds  int64 `json:"bzDownWinRounds" bson:"bz_down_win_rounds"`   // 小豹子赢局
	THSUpWinRounds   int64 `json:"thsUpWinRounds" bson:"ths_up_win_rounds"`     // 大同花赢局
	THSDownWinRounds int64 `json:"thsDownWinRounds" bson:"ths_down_win_rounds"` // 小同花赢局
	DSZUpWinRounds   int64 `json:"dszUpWinRounds" bson:"dsz_up_win_rounds"`     // 大顺子赢局
	DSZDownWinRounds int64 `json:"dszDownWinRounds" bson:"dsz_down_win_rounds"` // 小顺子赢局
	DTHUpWinRounds   int64 `json:"dthUpWinRounds" bson:"dth_up_win_rounds"`     // 大同花赢局
	DTHDownWinRounds int64 `json:"dthDownWinRounds" bson:"dth_down_win_rounds"` // 小同花赢局
	DDZUpWinRounds   int64 `json:"ddzUpWinRounds" bson:"ddz_up_win_rounds"`     // 大对子赢局
	DDZDownWinRounds int64 `json:"ddzDownWinRounds" bson:"ddz_down_win_rounds"` // 小对子赢局
	DGPUpWinRounds   int64 `json:"dgpUpWinRounds" bson:"dgp_up_win_rounds"`     // 大高牌赢局
	DGPDownWinRounds int64 `json:"dgpDownWinRounds" bson:"dgp_down_win_rounds"` // 小高牌赢局
}

// TP 大输局
type TPGameLose struct {
	BZUpLoseRounds    int64 `json:"bzUpLoseRounds" bson:"bz_up_lose_rounds"`       // 大豹子输局
	BZDownLoseRounds  int64 `json:"bzDownLoseRounds" bson:"bz_down_lose_rounds"`   // 小豹子输局
	THSUpLoseRounds   int64 `json:"thsUpLoseRounds" bson:"ths_up_lose_rounds"`     // 大同花输局
	THSDownLoseRounds int64 `json:"thsDownLoseRounds" bson:"ths_down_lose_rounds"` // 小同花输局
	DSZUpLoseRounds   int64 `json:"dszUpLoseRounds" bson:"dsz_up_lose_rounds"`     // 大顺子输局
	DSZDownLoseRounds int64 `json:"dszDownLoseRounds" bson:"dsz_down_lose_rounds"` // 小顺子输局
	DTHUpLoseRounds   int64 `json:"dthUpLoseRounds" bson:"dth_up_lose_rounds"`     // 大同花输局
	DTHDownLoseRounds int64 `json:"dthDownLoseRounds" bson:"dth_down_lose_rounds"` // 小同花输局
	DDZUpLoseRounds   int64 `json:"ddzUpLoseRounds" bson:"ddz_up_lose_rounds"`     // 大对子输局
	DDZDownLoseRounds int64 `json:"ddzDownLoseRounds" bson:"ddz_down_lose_rounds"` // 小对子输局
	DGPUpLoseRounds   int64 `json:"dgpUpLoseRounds" bson:"dgp_up_lose_rounds"`     // 大高牌输局
	DGPDownLoseRounds int64 `json:"dgpDownLoseRounds" bson:"dgp_down_lose_rounds"` // 小高牌输局
}

// TP局 弃牌率、跟注率
type TPGameRate struct {
	BZUpDiscardRate    float64 `json:"bzUpDiscardRate" bson:"bz_up_discard_rate"`       // 大豹子弃牌
	BZUpFollowRate     float64 `json:"bzUpFollowRate" bson:"bz_up_follow_rate"`         // 大豹子跟注
	BZDownDiscardRate  float64 `json:"bzDownDiscardRate" bson:"bz_down_discard_rate"`   // 小豹子弃牌
	BZDownFollowRate   float64 `json:"bzDownFollowRate" bson:"bz_down_follow_rate"`     // 小豹子跟注
	THSUpDiscardRate   float64 `json:"thsUpDiscardRate" bson:"ths_up_discard_rate"`     // 大同花弃牌
	THSUpFollowRate    float64 `json:"thsUpFollowRate" bson:"ths_up_follow_rate"`       // 大同花跟注
	THSDownDiscardRate float64 `json:"thsDownDiscardRate" bson:"ths_down_discard_rate"` // 小同花弃牌
	THSDownFollowRate  float64 `json:"thsDownFollowRate" bson:"ths_down_follow_rate"`   // 小同花跟注
	DSZUpDiscardRate   float64 `json:"dszUpDiscardRate" bson:"dsz_up_discard_rate"`     // 大顺子弃牌
	DSZUpFollowRate    float64 `json:"dszUpFollowRate" bson:"dsz_up_follow_rate"`       // 大顺子跟注
	DSZDownDiscardRate float64 `json:"dszDownDiscardRate" bson:"dsz_down_discard_rate"` // 小顺子弃牌
	DSZDownFollowRate  float64 `json:"dszDownFollowRate" bson:"dsz_down_follow_rate"`   // 小顺子跟注
	DTHUpDiscardRate   float64 `json:"dthUpDiscardRate" bson:"dth_up_discard_rate"`     // 大同花弃牌
	DTHUpFollowRate    float64 `json:"dthUpFollowRate" bson:"dth_up_follow_rate"`       // 大同花跟注
	DTHDownDiscardRate float64 `json:"dthDownDiscardRate" bson:"dth_down_discard_rate"` // 小同花弃牌
	DTHDownFollowRate  float64 `json:"dthDownFollowRate" bson:"dth_down_follow_rate"`   // 小同花跟注
	DDZUpDiscardRate   float64 `json:"ddzUpDiscardRate" bson:"ddz_up_discard_rate"`     // 大对子弃牌
	DDZUpFollowRate    float64 `json:"ddzUpFollowRate" bson:"ddz_up_follow_rate"`       // 大对子跟注
	DDZDownDiscardRate float64 `json:"ddzDownDiscardRate" bson:"ddz_down_discard_rate"` // 小对子弃牌
	DDZDownFollowRate  float64 `json:"ddzDownFollowRate" bson:"ddz_down_follow_rate"`   // 小对子跟注
	DGPUpDiscardRate   float64 `json:"dgpUpDiscardRate" bson:"dgp_up_discard_rate"`     // 大高牌弃牌
	DGPUpFollowRate    float64 `json:"dgpUpFollowRate" bson:"dgp_up_follow_rate"`       // 大高牌跟注
	DGPDownDiscardRate float64 `json:"dgpDownDiscardRate" bson:"dgp_down_discard_rate"` // 小高牌弃牌
	DGPDownFollowRate  float64 `json:"dgpDownFollowRate" bson:"dgp_down_follow_rate"`   // 小高牌跟注
}

// TP 房间统计
type TPPlayerRoomStat struct {
	Id                     string `bson:"_id"` // date-userid
	Date                   int64  `bson:"date"`
	DateStr                string `bson:"date_str"`
	Userid                 string `bson:"userid"`
	Roomid                 string `json:"roomid" bson:"roomid"`                                     // 房间ID
	AllRounds              int64  `json:"allRounds" bson:"all_rounds"`                              // 局数
	Bets                   int64  `bson:"bets"`                                                     // 总下注
	WinRounds              int32  `bson:"win_rounds"`                                               // 胜利局数
	LoseRounds             int32  `bson:"lose_rounds"`                                              // 失败局数
	WinBets                int64  `bson:"win_bets"`                                                 // 胜局打码量
	LoseBets               int64  `bson:"lose_bets"`                                                // 败局打码量
	GameTime               int64  `json:"gameTime" bson:"game_time"`                                // 游戏时长
	UpNumber               int64  `json:"upNumber" bson:"up_number"`                                // 主动升到此场次数
	DownNumber             int64  `json:"downNumber" bson:"down_number"`                            // 主动降到此场次数
	TwoRounds              int64  `json:"twoRounds" bson:"two_rounds"`                              // 两人局
	WinTwoRounds           int64  `json:"winTwoRounds" bson:"win_two_rounds"`                       // 两人赢局
	TwoBets                int64  `json:"twoBets" bson:"two_bets"`                                  // 两人局总下注
	WinTwoBets             int64  `json:"winTwoBets" bson:"win_two_bets"`                           // 两人局总返奖
	ThreeRounds            int64  `json:"threeRounds" bson:"three_rounds"`                          // 三人局
	WinThreeRounds         int64  `json:"winThreeRounds" bson:"win_three_rounds"`                   // 三人赢局
	ThreeBets              int64  `json:"threeBets" bson:"three_bets"`                              // 三人局总下注
	WinThreeBets           int64  `json:"winThreeBets" bson:"win_three_bets"`                       // 三人局总返奖
	FourRounds             int64  `json:"fourRounds" bson:"four_rounds"`                            // 四人局
	WinFourRounds          int64  `json:"winFourRounds" bson:"win_four_rounds"`                     // 四人赢局
	FourBets               int64  `json:"fourBets" bson:"four_bets"`                                // 四人局总下注
	WinFourBets            int64  `json:"winFourBets" bson:"win_four_bets"`                         // 四人局总返奖
	FiveRounds             int64  `json:"fiveRounds" bson:"five_rounds"`                            // 五人局
	WinFiveRounds          int64  `json:"winFiveRounds" bson:"win_five_rounds"`                     // 五人赢局
	FiveBets               int64  `json:"fiveBets" bson:"five_bets"`                                // 五人局总下注
	WinFiveBets            int64  `json:"winFiveBets" bson:"win_five_bets"`                         // 五人局总返奖
	ChipPool               int64  `json:"chipPool" bson:"chip_pool"`                                // 筹码池上限
	BZUpNumber             int64  `json:"bzUpNumber" bson:"bz_up_number"`                           // 大豹子局数
	BZUpRounds             int64  `json:"bzUpRounds" bson:"bz_up_rounds"`                           // 大豹子轮次
	BZUpBets               int64  `json:"bzUpBets" bson:"bz_up_bets"`                               // 大豹子下注
	BZUpWinRounds          int64  `json:"bzUpWinRounds" bson:"bz_up_win_rounds"`                    // 大豹子赢局
	BZUpLoseRounds         int64  `json:"bzUpLoseRounds" bson:"bz_up_lose_rounds"`                  // 大豹子输局
	BZUpTwoDiscardNum      int64  `json:"bzUpTwoDiscardNum" bson:"bz_up_two_discard_num"`           // 大豹子两人局弃牌次数
	BZUpTwoFollowNum       int64  `json:"bzUpTwoFollowNum" bson:"bz_up_two_follow_num"`             // 大豹子两人局跟注次数
	BZUpThreeDiscardNum    int64  `json:"bzUpThreeDiscardNum" bson:"bz_up_three_discard_num"`       // 大豹子三人局弃牌次数
	BZUpThreeFollowNum     int64  `json:"bzUpThreeFollowNum" bson:"bz_up_three_follow_num"`         // 大豹子三人局跟注次数
	BZUpFourDiscardNum     int64  `json:"bzUpFourDiscardNum" bson:"bz_up_four_discard_num"`         // 大豹子四人局弃牌次数
	BZUpFourFollowNum      int64  `json:"bzUpFourFollowNum" bson:"bz_up_four_follow_num"`           // 大豹子四人局跟注次数
	BZUpFiveDiscardNum     int64  `json:"bzUpFiveDiscardNum" bson:"bz_up_five_discard_num"`         // 大豹子五人局弃牌次数
	BZUpFiveFollowNum      int64  `json:"bzUpFiveFollowNum" bson:"bz_up_five_follow_num"`           // 大豹子五人局跟注次数
	BZDownNumber           int64  `json:"bzDownNumber" bson:"bz_down_number"`                       // 小豹子局数
	BZDownRounds           int64  `json:"bzDownRounds" bson:"bz_down_rounds"`                       // 小豹子轮次
	BZDownBets             int64  `json:"bzDownBets" bson:"bz_down_bets"`                           // 小豹子下注
	BZDownWinRounds        int64  `json:"bzDownWinRounds" bson:"bz_down_win_rounds"`                // 小豹子赢局
	BZDownLoseRounds       int64  `json:"bzDownLoseRounds" bson:"bz_down_lose_rounds"`              // 小豹子输局
	BZDownTwoDiscardNum    int64  `json:"bzDownTwoDiscardNum" bson:"bz_down_two_discard_num"`       // 小豹子两人局弃牌次数
	BZDownTwoFollowNum     int64  `json:"bzDownTwoFollowNum" bson:"bz_down_two_follow_num"`         // 小豹子两人局跟注次数
	BZDownThreeDiscardNum  int64  `json:"bzDownThreeDiscardNum" bson:"bz_down_three_discard_num"`   // 小豹子三人局弃牌次数
	BZDownThreeFollowNum   int64  `json:"bzDownThreeFollowNum" bson:"bz_down_three_follow_num"`     // 小豹子三人局跟注次数
	BZDownFourDiscardNum   int64  `json:"bzDownFourDiscardNum" bson:"bz_down_four_discard_num"`     // 小豹子四人局弃牌次数
	BZDownFourFollowNum    int64  `json:"bzDownFourFollowNum" bson:"bz_down_four_follow_num"`       // 小豹子四人局跟注次数
	BZDownFiveDiscardNum   int64  `json:"bzDownFiveDiscardNum" bson:"bz_down_five_discard_num"`     // 小豹子五人局弃牌次数
	BZDownFiveFollowNum    int64  `json:"bzDownFiveFollowNum" bson:"bz_down_five_follow_num"`       // 小豹子五人局跟注次数
	THSUpNumber            int64  `json:"thsUpNumber" bson:"ths_up_number"`                         // 大同花顺局数
	THSUpRounds            int64  `json:"thsUpRounds" bson:"ths_up_rounds"`                         // 大同花顺轮次
	THSUpBets              int64  `json:"thsUpBets" bson:"ths_up_bets"`                             // 大同花下注
	THSUpWinRounds         int64  `json:"thsUpWinRounds" bson:"ths_up_win_rounds"`                  // 大同花赢局
	THSUpLoseRounds        int64  `json:"thsUpLoseRounds" bson:"ths_up_lose_rounds"`                // 大同花输局
	THSUpTwoDiscardNum     int64  `json:"thsUpTwoDiscardNum" bson:"ths_up_two_discard_num"`         // 大同花两人局弃牌次数
	THSUpTwoFollowNum      int64  `json:"thsUpTwoFollowNum" bson:"ths_up_two_follow_num"`           // 大同花两人局跟注次数
	THSUpThreeDiscardNum   int64  `json:"thsUpThreeDiscardNum" bson:"ths_up_three_discard_num"`     // 大同花三人局弃牌次数
	THSUpThreeFollowNum    int64  `json:"thsUpThreeFollowNum" bson:"ths_up_three_follow_num"`       // 大同花三人局跟注次数
	THSUpFourDiscardNum    int64  `json:"thsUpFourDiscardNum" bson:"ths_up_four_discard_num"`       // 大同花四人局弃牌次数
	THSUpFourFollowNum     int64  `json:"thsUpFourFollowNum" bson:"ths_up_four_follow_num"`         // 大同花四人局跟注次数
	THSUpFiveDiscardNum    int64  `json:"thsUpFiveDiscardNum" bson:"ths_up_five_discard_num"`       // 大同花五人局弃牌次数
	THSUpFiveFollowNum     int64  `json:"thsUpFiveFollowNum" bson:"ths_up_five_follow_num"`         // 大同花五人局跟注次数
	THSDownNumber          int64  `json:"thsDownNumber" bson:"ths_down_number"`                     // 小同花顺局数
	THSDownRounds          int64  `json:"thsDownRounds" bson:"ths_down_rounds"`                     // 小同花顺轮次
	THSDownBets            int64  `json:"thsDownBets" bson:"ths_down_bets"`                         // 小同花下注
	THSDownWinRounds       int64  `json:"thsDownWinRounds" bson:"ths_down_win_rounds"`              // 小同花赢局
	THSDownLoseRounds      int64  `json:"thsDownLoseRounds" bson:"ths_down_lose_rounds"`            // 小同花输局
	THSDownTwoDiscardNum   int64  `json:"thsDownTwoDiscardNum" bson:"ths_down_two_discard_num"`     // 小同花两人局弃牌次数
	THSDownTwoFollowNum    int64  `json:"thsDownTwoFollowNum" bson:"ths_down_two_follow_num"`       // 小同花两人局跟注次数
	THSDownThreeDiscardNum int64  `json:"thsDownThreeDiscardNum" bson:"ths_down_three_discard_num"` // 小同花三人局弃牌次数
	THSDownThreeFollowNum  int64  `json:"thsDownThreeFollowNum" bson:"ths_down_three_follow_num"`   // 小同花三人局跟注次数
	THSDownFourDiscardNum  int64  `json:"thsDownFourDiscardNum" bson:"ths_down_four_discard_num"`   // 小同花四人局弃牌次数
	THSDownFourFollowNum   int64  `json:"thsDownFourFollowNum" bson:"ths_down_four_follow_num"`     // 小同花四人局跟注次数
	THSDownFiveDiscardNum  int64  `json:"thsDownFiveDiscardNum" bson:"ths_down_five_discard_num"`   // 小同花五人局弃牌次数
	THSDownFiveFollowNum   int64  `json:"thsDownFiveFollowNum" bson:"ths_down_five_follow_num"`     // 小同花五人局跟注次数
	DSZUpNumber            int64  `json:"dszUpNumber" bson:"dsz_up_number"`                         // 大顺子局数
	DSZUpRounds            int64  `json:"dszUpRounds" bson:"dsz_up_rounds"`                         // 大顺子轮次
	DSZUpBets              int64  `json:"dszUpBets" bson:"dsz_up_bets"`                             // 大顺子下注
	DSZUpWinRounds         int64  `json:"dszUpWinRounds" bson:"dsz_up_win_rounds"`                  // 大顺子赢局
	DSZUpLoseRounds        int64  `json:"dszUpLoseRounds" bson:"dsz_up_lose_rounds"`                // 大顺子输局
	DSZUpTwoDiscardNum     int64  `json:"dszUpTwoDiscardNum" bson:"dsz_up_two_discard_num"`         // 大顺子两人局弃牌次数
	DSZUpTwoFollowNum      int64  `json:"dszUpTwoFollowNum" bson:"dsz_up_two_follow_num"`           // 大顺子两人局跟注次数
	DSZUpThreeDiscardNum   int64  `json:"dszUpThreeDiscardNum" bson:"dsz_up_three_discard_num"`     // 大顺子三人局弃牌次数
	DSZUpThreeFollowNum    int64  `json:"dszUpThreeFollowNum" bson:"dsz_up_three_follow_num"`       // 大顺子三人局跟注次数
	DSZUpFourDiscardNum    int64  `json:"dszUpFourDiscardNum" bson:"dsz_up_four_discard_num"`       // 大顺子四人局弃牌次数
	DSZUpFourFollowNum     int64  `json:"dszUpFourFollowNum" bson:"dsz_up_four_follow_num"`         // 大顺子四人局跟注次数
	DSZUpFiveDiscardNum    int64  `json:"dszUpFiveDiscardNum" bson:"dsz_up_five_discard_num"`       // 大顺子五人局弃牌次数
	DSZUpFiveFollowNum     int64  `json:"dszUpFiveFollowNum" bson:"dsz_up_five_follow_num"`         // 大顺子五人局跟注次数
	DSZDownNumber          int64  `json:"dszDownNumber" bson:"dsz_down_number"`                     // 小顺子局数
	DSZDownRounds          int64  `json:"dszDownRounds" bson:"dsz_down_rounds"`                     // 小顺子轮次
	DSZDownBets            int64  `json:"dszDownBets" bson:"dsz_down_bets"`                         // 小顺子下注
	DSZDownWinRounds       int64  `json:"dszDownWinRounds" bson:"dsz_down_win_rounds"`              // 小顺子赢局
	DSZDownLoseRounds      int64  `json:"dszDownLoseRounds" bson:"dsz_down_lose_rounds"`            // 小顺子输局
	DSZDownTwoDiscardNum   int64  `json:"dszDownTwoDiscardNum" bson:"dsz_down_two_discard_num"`     // 小顺子两人局弃牌次数
	DSZDownTwoFollowNum    int64  `json:"dszDownTwoFollowNum" bson:"dsz_down_two_follow_num"`       // 小顺子两人局跟注次数
	DSZDownThreeDiscardNum int64  `json:"dszDownThreeDiscardNum" bson:"dsz_down_three_discard_num"` // 小顺子三人局弃牌次数
	DSZDownThreeFollowNum  int64  `json:"dszDownThreeFollowNum" bson:"dsz_down_three_follow_num"`   // 小顺子三人局跟注次数
	DSZDownFourDiscardNum  int64  `json:"dszDownFourDiscardNum" bson:"dsz_down_four_discard_num"`   // 小顺子四人局弃牌次数
	DSZDownFourFollowNum   int64  `json:"dszDownFourFollowNum" bson:"dsz_down_four_follow_num"`     // 小顺子四人局跟注次数
	DSZDownFiveDiscardNum  int64  `json:"dszDownFiveDiscardNum" bson:"dsz_down_five_discard_num"`   // 小顺子五人局弃牌次数
	DSZDownFiveFollowNum   int64  `json:"dszDownFiveFollowNum" bson:"dsz_down_five_follow_num"`     // 小顺子五人局跟注次数
	DTHUpNumber            int64  `json:"dthUpNumber" bson:"dth_up_number"`                         // 大同花局数
	DTHUpRounds            int64  `json:"dthUpRounds" bson:"dth_up_rounds"`                         // 大同花轮次
	DTHUpBets              int64  `json:"dthUpBets" bson:"dth_up_bets"`                             // 大同花下注
	DTHUpWinRounds         int64  `json:"dthUpWinRounds" bson:"dth_up_win_rounds"`                  // 大同花赢局
	DTHUpLoseRounds        int64  `json:"dthUpLoseRounds" bson:"dth_up_lose_rounds"`                // 大同花输局
	DTHUpTwoDiscardNum     int64  `json:"dthUpTwoDiscardNum" bson:"dth_up_two_discard_num"`         // 大同花两人局弃牌次数
	DTHUpTwoFollowNum      int64  `json:"dthUpTwoFollowNum" bson:"dth_up_two_follow_num"`           // 大同花两人局跟注次数
	DTHUpThreeDiscardNum   int64  `json:"dthUpThreeDiscardNum" bson:"dth_up_three_discard_num"`     // 大同花三人局弃牌次数
	DTHUpThreeFollowNum    int64  `json:"dthUpThreeFollowNum" bson:"dth_up_three_follow_num"`       // 大同花三人局跟注次数
	DTHUpFourDiscardNum    int64  `json:"dthUpFourDiscardNum" bson:"dth_up_four_discard_num"`       // 大同花四人局弃牌次数
	DTHUpFourFollowNum     int64  `json:"dthUpFourFollowNum" bson:"dth_up_four_follow_num"`         // 大同花四人局跟注次数
	DTHUpFiveDiscardNum    int64  `json:"dthUpFiveDiscardNum" bson:"dth_up_five_discard_num"`       // 大同花五人局弃牌次数
	DTHUpFiveFollowNum     int64  `json:"dthUpFiveFollowNum" bson:"dth_up_five_follow_num"`         // 大同花五人局跟注次数
	DTHDownNumber          int64  `json:"dthDownNumber" bson:"dth_down_number"`                     // 小同花局数
	DTHDownRounds          int64  `json:"dthDownRounds" bson:"dth_down_rounds"`                     // 小同花轮次
	DTHDownBets            int64  `json:"dthDownBets" bson:"dth_down_bets"`                         // 小同花下注
	DTHDownWinRounds       int64  `json:"dthDownWinRounds" bson:"dth_down_win_rounds"`              // 小同花赢局
	DTHDownLoseRounds      int64  `json:"dthDownLoseRounds" bson:"dth_down_lose_rounds"`            // 小同花输局
	DTHDownTwoDiscardNum   int64  `json:"dthDownTwoDiscardNum" bson:"dth_down_two_discard_num"`     // 小同花两人局弃牌次数
	DTHDownTwoFollowNum    int64  `json:"dthDownTwoFollowNum" bson:"dth_down_two_follow_num"`       // 小同花两人局跟注次数
	DTHDownThreeDiscardNum int64  `json:"dthDownThreeDiscardNum" bson:"dth_down_three_discard_num"` // 小同花三人局弃牌次数
	DTHDownThreeFollowNum  int64  `json:"dthDownThreeFollowNum" bson:"dth_down_three_follow_num"`   // 小同花三人局跟注次数
	DTHDownFourDiscardNum  int64  `json:"dthDownFourDiscardNum" bson:"dth_down_four_discard_num"`   // 小同花四人局弃牌次数
	DTHDownFourFollowNum   int64  `json:"dthDownFourFollowNum" bson:"dth_down_four_follow_num"`     // 小同花四人局跟注次数
	DTHDownFiveDiscardNum  int64  `json:"dthDownFiveDiscardNum" bson:"dth_down_five_discard_num"`   // 小同花五人局弃牌次数
	DTHDownFiveFollowNum   int64  `json:"dthDownFiveFollowNum" bson:"dth_down_five_follow_num"`     // 小同花五人局跟注次数
	DDZUpNumber            int64  `json:"ddzUpNumber" bson:"ddz_up_number"`                         // 大对子局数
	DDZUpRounds            int64  `json:"ddzUpRounds" bson:"ddz_up_rounds"`                         // 大对子轮次
	DDZUpBets              int64  `json:"ddzUpBets" bson:"ddz_up_bets"`                             // 大对子下注
	DDZUpWinRounds         int64  `json:"ddzUpWinRounds" bson:"ddz_up_win_rounds"`                  // 大对子赢局
	DDZUpLoseRounds        int64  `json:"ddzUpLoseRounds" bson:"ddz_up_lose_rounds"`                // 大对子输局
	DDZUpTwoDiscardNum     int64  `json:"ddzUpTwoDiscardNum" bson:"ddz_up_two_discard_num"`         // 大对子两人局弃牌次数
	DDZUpTwoFollowNum      int64  `json:"ddzUpTwoFollowNum" bson:"ddz_up_two_follow_num"`           // 大对子两人局跟注次数
	DDZUpThreeDiscardNum   int64  `json:"ddzUpThreeDiscardNum" bson:"ddz_up_three_discard_num"`     // 大对子三人局弃牌次数
	DDZUpThreeFollowNum    int64  `json:"ddzUpThreeFollowNum" bson:"ddz_up_three_follow_num"`       // 大对子三人局跟注次数
	DDZUpFourDiscardNum    int64  `json:"ddzUpFourDiscardNum" bson:"ddz_up_four_discard_num"`       // 大对子四人局弃牌次数
	DDZUpFourFollowNum     int64  `json:"ddzUpFourFollowNum" bson:"ddz_up_four_follow_num"`         // 大对子四人局跟注次数
	DDZUpFiveDiscardNum    int64  `json:"ddzUpFiveDiscardNum" bson:"ddz_up_five_discard_num"`       // 大对子五人局弃牌次数
	DDZUpFiveFollowNum     int64  `json:"ddzUpFiveFollowNum" bson:"ddz_up_five_follow_num"`         // 大对子五人局跟注次数
	DDZDownNumber          int64  `json:"ddzDownNumber" bson:"ddz_down_number"`                     // 小对子局数
	DDZDownRounds          int64  `json:"ddzDownRounds" bson:"ddz_down_rounds"`                     // 小对子轮次
	DDZDownBets            int64  `json:"ddzDownBets" bson:"ddz_down_bets"`                         // 小对子下注
	DDZDownWinRounds       int64  `json:"ddzDownWinRounds" bson:"ddz_down_win_rounds"`              // 小对子赢局
	DDZDownLoseRounds      int64  `json:"ddzDownLoseRounds" bson:"ddz_down_lose_rounds"`            // 小对子输局
	DDZDownTwoDiscardNum   int64  `json:"ddzDownTwoDiscardNum" bson:"ddz_down_two_discard_num"`     // 小对子两人局弃牌次数
	DDZDownTwoFollowNum    int64  `json:"ddzDownTwoFollowNum" bson:"ddz_down_two_follow_num"`       // 小对子两人局跟注次数
	DDZDownThreeDiscardNum int64  `json:"ddzDownThreeDiscardNum" bson:"ddz_down_three_discard_num"` // 小对子三人局弃牌次数
	DDZDownThreeFollowNum  int64  `json:"ddzDownThreeFollowNum" bson:"ddz_down_three_follow_num"`   // 小对子三人局跟注次数
	DDZDownFourDiscardNum  int64  `json:"ddzDownFourDiscardNum" bson:"ddz_down_four_discard_num"`   // 小对子四人局弃牌次数
	DDZDownFourFollowNum   int64  `json:"ddzDownFourFollowNum" bson:"ddz_down_four_follow_num"`     // 小对子四人局跟注次数
	DDZDownFiveDiscardNum  int64  `json:"ddzDownFiveDiscardNum" bson:"ddz_down_five_discard_num"`   // 小对子五人局弃牌次数
	DDZDownFiveFollowNum   int64  `json:"ddzDownFiveFollowNum" bson:"ddz_down_five_follow_num"`     // 小对子五人局跟注次数
	DGPUpNumber            int64  `json:"dgpUpNumber" bson:"dgp_up_number"`                         // 大高牌局数
	DGPUpRounds            int64  `json:"dgpUpRounds" bson:"dgp_up_rounds"`                         // 大高牌轮次
	DGPUpBets              int64  `json:"dgpUpBets" bson:"dgp_up_bets"`                             // 大高牌下注
	DGPUpWinRounds         int64  `json:"dgpUpWinRounds" bson:"dgp_up_win_rounds"`                  // 大高牌赢局
	DGPUpLoseRounds        int64  `json:"dgpUpLoseRounds" bson:"dgp_up_lose_rounds"`                // 大高牌输局
	DGPUpTwoDiscardNum     int64  `json:"dgpUpTwoDiscardNum" bson:"dgp_up_two_discard_num"`         // 大高牌两人局弃牌次数
	DGPUpTwoFollowNum      int64  `json:"dgpUpTwoFollowNum" bson:"dgp_up_two_follow_num"`           // 大高牌两人局跟注次数
	DGPUpThreeDiscardNum   int64  `json:"dgpUpThreeDiscardNum" bson:"dgp_up_three_discard_num"`     // 大高牌三人局弃牌次数
	DGPUpThreeFollowNum    int64  `json:"dgpUpThreeFollowNum" bson:"dgp_up_three_follow_num"`       // 大高牌三人局跟注次数
	DGPUpFourDiscardNum    int64  `json:"dgpUpFourDiscardNum" bson:"dgp_up_four_discard_num"`       // 大高牌四人局弃牌次数
	DGPUpFourFollowNum     int64  `json:"dgpUpFourFollowNum" bson:"dgp_up_four_follow_num"`         // 大高牌四人局跟注次数
	DGPUpFiveDiscardNum    int64  `json:"dgpUpFiveDiscardNum" bson:"dgp_up_five_discard_num"`       // 大高牌五人局弃牌次数
	DGPUpFiveFollowNum     int64  `json:"dgpUpFiveFollowNum" bson:"dgp_up_five_follow_num"`         // 大高牌五人局跟注次数
	DGPDownNumber          int64  `json:"dgpDownNumber" bson:"dgp_down_number"`                     // 小高牌局数
	DGPDownRounds          int64  `json:"dgpDownRounds" bson:"dgp_down_rounds"`                     // 小高牌轮次
	DGPDownBets            int64  `json:"dgpDownBets" bson:"dgp_down_bets"`                         // 小高牌下注
	DGPDownWinRounds       int64  `json:"dgpDownWinRounds" bson:"dgp_down_win_rounds"`              // 小高牌赢局
	DGPDownLoseRounds      int64  `json:"dgpDownLoseRounds" bson:"dgp_down_lose_rounds"`            // 小高牌输局
	DGPDownTwoDiscardNum   int64  `json:"dgpDownTwoDiscardNum" bson:"dgp_down_two_discard_num"`     // 小高牌两人局弃牌次数
	DGPDownTwoFollowNum    int64  `json:"dgpDownTwoFollowNum" bson:"dgp_down_two_follow_num"`       // 小高牌两人局跟注次数
	DGPDownThreeDiscardNum int64  `json:"dgpDownThreeDiscardNum" bson:"dgp_down_three_discard_num"` // 小高牌三人局弃牌次数
	DGPDownThreeFollowNum  int64  `json:"dgpDownThreeFollowNum" bson:"dgp_down_three_follow_num"`   // 小高牌三人局跟注次数
	DGPDownFourDiscardNum  int64  `json:"dgpDownFourDiscardNum" bson:"dgp_down_four_discard_num"`   // 小高牌四人局弃牌次数
	DGPDownFourFollowNum   int64  `json:"dgpDownFourFollowNum" bson:"dgp_down_four_follow_num"`     // 小高牌四人局跟注次数
	DGPDownFiveDiscardNum  int64  `json:"dgpDownFiveDiscardNum" bson:"dgp_down_five_discard_num"`   // 小高牌五人局弃牌次数
	DGPDownFiveFollowNum   int64  `json:"dgpDownFiveFollowNum" bson:"dgp_down_five_follow_num"`     // 小高牌五人局跟注次数
	AD_BundleId            string `bson:"ad__bundle_id"`                                            // 渠道
	Channel1               string `bson:"channel1"`                                                 // 渠道别名
	RegistArea             int    `bson:"regist_area"`                                              // 账号类型 ab测试 0:A 1:B 2:C
	Money                  uint32 `bson:"money"`                                                    // 充值总金额(分)
}

type LHDPlayerStat struct {
	Id               string   `bson:"_id"` // date-userid
	Date             int64    `bson:"date"`
	DateStr          string   `bson:"date_str"`
	Userid           string   `bson:"userid"`
	NewReg           bool     `bson:"new_reg"`            // 是新顾客
	AllRounds        int32    `bson:"all_rounds"`         // 总局数
	GameTimes        int64    `bson:"game_times"`         // 游戏总时长(秒)
	Bets             int64    `bson:"bets"`               // 总打码量
	BetRounds        int32    `bson:"bet_rounds"`         // 下注局数
	WinRounds        int32    `bson:"win_rounds"`         // 胜利局数
	LoseRounds       int32    `bson:"lose_rounds"`        // 失败局数
	TieRounds        int32    `bson:"tie_rounds"`         // 和局数
	WinBets          int64    `bson:"win_bets"`           // 胜局打码量
	LoseBets         int64    `bson:"lose_bets"`          // 败局打码量
	TieBets          int64    `bson:"tie_bets"`           // 和局打码量
	Wins             int64    `bson:"wins"`               // 胜局赢钱金额
	Loses            int64    `bson:"loses"`              // 败局输钱金额
	Cash             int64    `bson:"cash"`               // 净赢
	Dragons          int64    `bson:"dragons"`            // 下注局开龙数
	Tigers           int64    `bson:"tigers"`             // 下注局开虎数
	Ties             int64    `bson:"ties"`               // 下注局开和数
	PlayerDragons    int64    `bson:"player_dragons"`     // 玩家下龙局数
	PlayerTigers     int64    `bson:"player_tigers"`      // 玩家下虎局数
	PlayerTies       int64    `bson:"player_ties"`        // 玩家下和局数
	PlayerWinDragons int64    `bson:"player_win_dragons"` // 玩家下龙赢的局数
	PlayerWinTigers  int64    `bson:"player_win_tigers"`  // 玩家下虎赢的局数
	PlayerWinTies    int64    `bson:"player_win_ties"`    // 玩家下和赢的局数
	PlayerMultis     int64    `bson:"player_multis"`      // 多门局数 玩家下注2门及以上的局数
	RoundBetAvg      int64    `bson:"round_bet_avg"`      // 局均码
	AD_BundleId      string   `bson:"ad__bundle_id"`      // 渠道
	Channel1         string   `bson:"channel1"`           // 渠道别名
	RegistArea       int      `bson:"regist_area"`        // 账号类型 ab测试 0:A 1:B 2:C
	Money            uint32   `bson:"money"`              // 充值总金额(分)
	ObserveRounds    int32    `bson:"observe_rounds"`     // 观察局数
	Players          int32    `bson:"players"`            // 总玩crash人数
	NewPlayers       int32    `bson:"new_players"`        // 新顾客玩crash人数
	OldPlayers       int32    `bson:"old_players"`        // 老顾客玩crash人数
	AllRoundsList    []int32  `bson:"all_rounds_list"`    // 玩家局数list
	Userids          []string `bson:"userids"`            // 汇总时的userid列表

	No          string `bson:"-"` // 序号
	FNickname   string `bson:"-"` // 玩家昵称
	FCtime      string `bson:"-"` // 注册时间
	FLoginTime  string `bson:"-"` // 最后登录时间
	FLiveDays   int    `bson:"-"` // 存活天数
	FLoseDays   int    `bson:"-"` // 流失天数
	FRegistArea string `bson:"-"` // 账号类型
	FUserType   string `bson:"-"` // 玩家类型

	FMoney    string `bson:"-"` // 总充值
	FCashOut  string `bson:"-"` // 总提现
	FWinMoney string `bson:"-"` // 总赢

	SMoney   int   `bson:"-"` // 总汇充值
	SCashOut int   `bson:"-"` // 总汇提现
	SDiamond int64 `bson:"-"` // 总汇携带分

	FBets             string `bson:"-"` // 总打码量
	FRoundBetsAvg     string `bson:"-"` // 局均打码量
	FWinRate          string `bson:"-"` // 胜率
	FRebateRate       string `bson:"-"` // 返奖率
	FWinBets          string `bson:"-"` // 胜局打码量
	FLoseBets         string `bson:"-"` // 败局打码量
	FTieBets          string `bson:"-"` // 平局打码量
	FWinBetsAvg       string `bson:"-"` // 胜局均码量
	FLoseBetsAvg      string `bson:"-"` // 败局均码量
	FTieBetsAvg       string `bson:"-"` // 平局均码量
	FWins             string `bson:"-"` // 胜局赢钱金额
	FLoses            string `bson:"-"` // 败局输钱金额
	FCash             string `bson:"-"` // 净赢
	FWinsAvg          string `bson:"-"` // 胜局均盈利
	FLosesAvg         string `bson:"-"` // 败局均输额
	FCashAvg          string `bson:"-"` // 局均净赢
	ObserveRoundsAvg  string `bson:"-"` // 局均观察次数
	ObserveRoundsRate string `bson:"-"` // 观察局数占比

	PlayerDragonsRate    string `bson:"-"` // 玩家押龙率
	PlayerTigersRate     string `bson:"-"` // 玩家押虎率
	PlayerTiesRate       string `bson:"-"` // 玩家押和率
	PlayerDragonsWinRate string `bson:"-"` // 玩家押龙率
	PlayerTigersWinRate  string `bson:"-"` // 玩家押虎率
	PlayerTiesWinRate    string `bson:"-"` // 玩家押和率
	PlayerMultisRate     string `bson:"-"` // 多门率

	FLoginedUsers       int    `bson:"-"` // 日活
	FLoginedNewUsers    int    `bson:"-"` // 新用户日活
	FLoginedOldUsers    int    `bson:"-"` // 老用户日活
	FPlayerRate         string `bson:"-"` // 玩crash人数占日活比
	FNewPlayerRate      string `bson:"-"` // 新顾客玩crash占新顾客日活比
	FOldPlayerRate      string `bson:"-"` // 老顾客玩crash占老顾客日活比
	FPlayerRoundsAvg    string `bson:"-"` // 人均局数
	FPlayerRoundsMedian int32  `bson:"-"` // 局数中位数
	FGameTimesAvg       string `bson:"-"` // 游戏局均时长（min)
	FGameTimesPlayerAvg string `bson:"-"` // 游戏人均时长（min)
	FPlayerBetsAvg      string `bson:"-"` // 人均打码
}

// lhd 心想事成统计
type LHDXxscStat struct {
	Id                    string   `bson:"_id"` // date-userid
	Date                  int64    `bson:"date"`
	DateStr               string   `bson:"date_str"`
	Userid                string   `bson:"userid"`
	StrategyPlayers       int32    `bson:"strategy_players"`        // 触发策略人数
	StrategyTimes         int32    `bson:"-"`                       // 触发策略总次数: u总 LHDXXSC.MaxTriggerTimes
	StrategyValidTimes    int32    `bson:"-"`                       // 触发策略有效次数: u LHDXXSC.TriggerTimes
	StrategyRounds        int32    `bson:"strategy_rounds"`         // 触发策略局数 r总
	StrategyDisturbRounds int32    `bson:"strategy_disturb_rounds"` // 扰动局数
	StrategyMultiRounds   int32    `bson:"strategy_multi_rounds"`   // 策略中多门局数
	StrategyBets          int64    `bson:"strategy_bets"`           // 策略中总打码量
	StrategyWinRounds     int64    `bson:"strategy_win_rounds"`     // 策略中总赢局数
	StrategyWins          int64    `bson:"strategy_wins"`           // 策略中总赢
	StrategyLoses         int64    `bson:"strategy_loses"`          // 策略中总输
	StrategyCash          int64    `bson:"-"`                       // 策略中净赢: 赢-输
	Userids               []string `bson:"userids"`                 // 汇总时的userid列表

	No          string `bson:"-"` // 序号
	FNickname   string `bson:"-"` // 总打码量
	FCtime      string `bson:"-"` // 注册时间
	FLoginTime  string `bson:"-"` // 最后登录时间
	FLiveDays   int    `bson:"-"` // 存活天数
	FLoseDays   int    `bson:"-"` // 流失天数
	FRegistArea string `bson:"-"` // 账号类型
	FUserType   string `bson:"-"` // 玩家类型

	StrategyValidTimesRate  string `bson:"-"` // 触发策略有效次数占触发策略总次数比
	StrategyValidTimesAvg   string `bson:"-"` // 人均有效次数
	StrategyRoundAvg        string `bson:"-"` // 每次触发平均所用局数
	StrategyValidRoundAvg   string `bson:"-"` // 每次有效触发平均所用局数
	StrategyDisturbRate     string `bson:"-"` // 扰动局数占比
	StrategyMultiRoundsRate string `bson:"-"` // 多门局数占比
	FStrategyBets           string `bson:"-"` // 策略中总打码量
	FStrategyBetRoundsAvg   string `bson:"-"` // 策略中局均打码量
	FStrategyWins           string `bson:"-"` // 策略中总赢
	FStrategyLoses          string `bson:"-"` // 策略中总输
	FStrategyCash           string `bson:"-"` // 策略中净赢: 赢-输
	FStrategyWinRate        string `bson:"-"` // 策略局胜率
	FStrategyRebateRate     string `bson:"-"` // 策略局返奖率
}

// lhd 求死不能统计
type LHDQsbnStat struct {
	Id                  string   `bson:"_id"` // date-userid
	Date                int64    `bson:"date"`
	DateStr             string   `bson:"date_str"`
	Userid              string   `bson:"userid"`
	StrategyPlayers     int32    `bson:"strategy_players"`      // 触发策略人数
	StrategyTimes       int32    `bson:"strategy_times"`        // 触发策略总次数
	AllInRounds         int32    `bson:"all_in_rounds"`         // allin局数
	StrategyMultiRounds int32    `bson:"strategy_multi_rounds"` // 策略中多门局数
	StrategyBets        int64    `bson:"strategy_bets"`         // 策略中总打码量
	StrategyWins        int64    `bson:"strategy_wins"`         // 策略中总赢
	StrategyLoses       int64    `bson:"strategy_loses"`        // 策略中总输
	StrategyCash        int64    `bson:"-"`                     // 策略中净赢: 赢-输
	BeforeBackRate      int      `bson:"before_back_rate"`      // 账变前返奖率
	AfterBackRate       int      `bson:"after_back_rate"`       // 账变后返奖率
	Userids             []string `bson:"userids"`               // 汇总时的userid列表

	No          string `bson:"-"` // 序号
	FNickname   string `bson:"-"` // 总打码量
	FCtime      string `bson:"-"` // 注册时间
	FLoginTime  string `bson:"-"` // 最后登录时间
	FLiveDays   int    `bson:"-"` // 存活天数
	FLoseDays   int    `bson:"-"` // 流失天数
	FRegistArea string `bson:"-"` // 账号类型
	FUserType   string `bson:"-"` // 玩家类型

	StrategyPlayersAvg string `bson:"-"` // 人均触发次数
	StrategyRate       string `bson:"-"` // 触发率
	StrategyMultiRate  string `bson:"-"` // 多门数占比
	FStrategyBets      string `bson:"-"` // 策略中总打码量
	FStrategyBetsAvg   string `bson:"-"` // 策略中局均打码量
	FStrategyWins      string `bson:"-"` // 策略中总赢
	FStrategyLoses     string `bson:"-"` // 策略中总输
	FStrategyCash      string `bson:"-"` // 策略中净赢
	FBeforeBackRate    string `bson:"-"` // 触发策略前返奖率
	FAfterBackRate     string `bson:"-"` // 策略局后返奖率
	FBackRate          string `bson:"-"` // 玩家当前返奖率
}

// lhd 龙狂有祸统计
type LHDLkyhStat struct {
	Id                  string   `bson:"_id"` // date-userid
	Date                int64    `bson:"date"`
	DateStr             string   `bson:"date_str"`
	Userid              string   `bson:"userid"`
	BSTimes             int      `json:"bstimes" bson:"bstimes"`                         // 倍杀状态次数
	BSDays              int      `json:"bsDays" bson:"bs_days"`                          // 倍杀天数
	TriggerTimes        int64    `json:"triggerTimes" bson:"trigger_times"`              // 每日倍杀次数 t日
	TZ                  int64    `json:"tz" bson:"tz"`                                   // 总倍杀次数 T总
	NZ                  int      `json:"nz" bson:"nz"`                                   // 总倍杀局数
	BSBets              int64    `json:"bsBets" bson:"bs_bets"`                          // 倍杀局打码量
	YZNumber            int64    `json:"yzNumber" bson:"yz_number"`                      // 压制次数
	YZRounds            int64    `json:"yzRounds" bson:"yz_rounds"`                      // 压制局数
	YZBets              int64    `json:"yzBets" bson:"yz_bets"`                          // 压制打码量
	RZ                  int      `json:"rz" bson:"rz"`                                   // r策
	StrategyMultiRounds int32    `bson:"strategy_multi_rounds"`                          // 策略中多门局数
	Bets                int64    `json:"bets" bson:"bets"`                               // 总打码量
	Rounds              int64    `json:"rounds" bson:"rounds"`                           // 总局数
	WinRounds           int64    `json:"winRounds" bson:"win_rounds"`                    // 赢局
	WinBets             int64    `json:"winBets" bson:"win_bets"`                        // 总赢
	LoseRounds          int64    `json:"loseRounds" bson:"lose_rounds"`                  // 输局
	LoseBets            int64    `json:"loseBets" bson:"lose_bets"`                      // 总输
	StrategyTimes       int32    `bson:"strategy_times"`                                 // 触发策略总次数
	StrategyBets        int64    `bson:"strategy_bets"`                                  // 策略中总打码量
	StrategyWins        int64    `bson:"strategy_wins"`                                  // 策略中总赢
	StrategyWinRounds   int64    `json:"strategyWinRounds" bson:"strategy_win_rounds"`   // 策略局赢局
	StrategyLoses       int64    `bson:"strategy_loses"`                                 // 策略中总输
	StrategyLoseRounds  int64    `json:"strategyLoseRounds" bson:"strategy_lose_rounds"` // 策略局输局
	StrategyCash        int64    `bson:"-"`                                              // 策略中净赢: 赢-输
	AD_BundleId         string   `bson:"ad__bundle_id"`                                  // 渠道
	Channel1            string   `bson:"channel1"`                                       // 渠道别名
	RegistArea          int      `bson:"regist_area"`                                    // 账号类型 ab测试 0:A 1:B 2:C
	Money               uint32   `bson:"money"`                                          // 充值总金额(分)
	Userids             []string `bson:"userids"`                                        // 汇总时的userid列表

	No          string `bson:"-"` // 序号
	FNickname   string `bson:"-"` // 总打码量
	FCtime      string `bson:"-"` // 注册时间
	FLoginTime  string `bson:"-"` // 最后登录时间
	FLiveDays   int    `bson:"-"` // 存活天数
	FLoseDays   int    `bson:"-"` // 流失天数
	FRegistArea string `bson:"-"` // 账号类型
	FUserType   string `bson:"-"` // 玩家类型

	BSNumberAvg        string `bson:"-"` // 日均倍杀次数
	BSRate             string `bson:"-"` // 倍杀触发率
	YZRate             string `bson:"-"` // 压制率
	StrategyMultiRate  string `bson:"-"` // 多门数占比
	BSBetAvg           string `bson:"-"` // 倍杀局均码量
	YZBetAvg           string `bson:"-"` // 压制局均码量
	FBets              string `bson:"-"` // 总打码
	RoundBetAvg        string `bson:"-"` // 局均码
	FStrategyBets      string `bson:"-"` // 策略中总打码量
	FStrategyBetsAvg   string `bson:"-"` // 策略中局均打码量
	FStrategyWins      string `bson:"-"` // 策略中总赢
	FStrategyLoses     string `bson:"-"` // 策略中总输
	FStrategyCash      string `bson:"-"` // 策略中净赢
	TotalWinRate       string `bson:"-"` // 总胜率
	StrategyWinRate    string `bson:"-"` // 策略局中胜率
	TotalRewardRate    string `bson:"-"` // 龙虎总返奖率
	StrategyRewardRate string `bson:"-"` // 策略局中返奖率
	PlayerRewardRate   string `bson:"-"` // 玩家大盘返奖率
}

// lhd 高潮涌现统计
type LHDGcyxStat struct {
	No                 string // 序号
	UserId             string // 玩家id
	Nickname           string // 玩家昵称
	UserType           string // 玩家类型
	RegistArea         string // 账号类型
	Ctime              string // 注册时间
	LoginTime          string // 最后登录时间
	LiveDays           int    // 存活天数
	LoseDays           int    // 流失天数
	AllRounds          int64  // 总局数
	Bets               int64  // -总打码量
	FBets              string // 总打码
	BetsAvg            string // 局均打码量
	StrategyM          int64  // 高潮次数
	StrategyRoundsAvg  string // 平均单次高潮持续局数
	StrategyRounds     int64  // 高潮局数
	StrategyBets       int64  // 高潮局总打码
	FStrategyBets      string // 高潮局总打码
	StrategyBetsAvg    string // 高潮局局均打码量
	StrategyRate       string // 高潮频率
	StrategyWinRounds  int64  // - 高潮局赢局数
	StrategyWinRate    string // 高潮局中胜率
	StrategyRewardRate string // 高潮局返奖率
	Wins               int64  // - 龙虎赢金额
	Loses              int64  // - 龙虎输金额
	LHDRebateRate      string // 龙虎总返奖率
	WinRounds          int64  // - 玩龙虎赢局数
	LHDWinRate         string // 龙虎总胜率
	StrategyWinBetAvg  string // 高潮局中赢局均码量
	StrategyLoseBetAvg string // 高潮局中输局均码量
	StrategyWins       int64  // -高潮局中赢局赢钱金额
	StrategyLoses      int64  // -高潮局中输局输钱金额
	FStrategyWins      string // 高潮局中赢局赢钱金额
	FStrategyLoses     string // 高潮局中输局输钱金额
	StrategyCash       string // 高潮局中净赢
}

/*
	7updown
*/

type UpDownPlayerStat struct {
	Id               string   `bson:"_id"` // date-userid
	Date             int64    `bson:"date"`
	DateStr          string   `bson:"date_str"`
	Userid           string   `bson:"userid"`
	NewReg           bool     `bson:"new_reg"`            // 是新顾客
	AllRounds        int32    `bson:"all_rounds"`         // 总局数
	GameTimes        int64    `bson:"game_times"`         // 游戏总时长(秒)
	Bets             int64    `bson:"bets"`               // 总打码量
	BetRounds        int32    `bson:"bet_rounds"`         // 下注局数
	WinRounds        int32    `bson:"win_rounds"`         // 胜利局数
	LoseRounds       int32    `bson:"lose_rounds"`        // 失败局数
	TieRounds        int32    `bson:"tie_rounds"`         // 和局数
	WinBets          int64    `bson:"win_bets"`           // 胜局打码量
	LoseBets         int64    `bson:"lose_bets"`          // 败局打码量
	TieBets          int64    `bson:"tie_bets"`           // 和局打码量
	Wins             int64    `bson:"wins"`               // 胜局赢钱金额
	Loses            int64    `bson:"loses"`              // 败局输钱金额
	Cash             int64    `bson:"cash"`               // 净赢
	Dragons          int64    `bson:"dragons"`            // 下注局开龙数
	Tigers           int64    `bson:"tigers"`             // 下注局开虎数
	Ties             int64    `bson:"ties"`               // 下注局开和数
	PlayerDragons    int64    `bson:"player_dragons"`     // 玩家下龙局数
	PlayerTigers     int64    `bson:"player_tigers"`      // 玩家下虎局数
	PlayerTies       int64    `bson:"player_ties"`        // 玩家下和局数
	PlayerWinDragons int64    `bson:"player_win_dragons"` // 玩家下龙赢的局数
	PlayerWinTigers  int64    `bson:"player_win_tigers"`  // 玩家下虎赢的局数
	PlayerWinTies    int64    `bson:"player_win_ties"`    // 玩家下和赢的局数
	PlayerMultis     int64    `bson:"player_multis"`      // 多门局数 玩家下注2门及以上的局数
	RoundBetAvg      int64    `bson:"round_bet_avg"`      // 局均码
	AD_BundleId      string   `bson:"ad__bundle_id"`      // 渠道
	Channel1         string   `bson:"channel1"`           // 渠道别名
	RegistArea       int      `bson:"regist_area"`        // 账号类型 ab测试 0:A 1:B 2:C
	Money            uint32   `bson:"money"`              // 充值总金额(分)
	ObserveRounds    int32    `bson:"observe_rounds"`     // 观察局数
	Players          int32    `bson:"players"`            // 总玩crash人数
	NewPlayers       int32    `bson:"new_players"`        // 新顾客玩crash人数
	OldPlayers       int32    `bson:"old_players"`        // 老顾客玩crash人数
	AllRoundsList    []int32  `bson:"all_rounds_list"`    // 玩家局数list
	Userids          []string `bson:"userids"`            // 汇总时的userid列表

	No          string `bson:"-"` // 序号
	FNickname   string `bson:"-"` // 玩家昵称
	FCtime      string `bson:"-"` // 注册时间
	FLoginTime  string `bson:"-"` // 最后登录时间
	FLiveDays   int    `bson:"-"` // 存活天数
	FLoseDays   int    `bson:"-"` // 流失天数
	FRegistArea string `bson:"-"` // 账号类型
	FUserType   string `bson:"-"` // 玩家类型

	FMoney    string `bson:"-"` // 总充值
	FCashOut  string `bson:"-"` // 总提现
	FWinMoney string `bson:"-"` // 总赢

	SMoney   int   `bson:"-"` // 总汇充值
	SCashOut int   `bson:"-"` // 总汇提现
	SDiamond int64 `bson:"-"` // 总汇携带分

	FBets             string `bson:"-"` // 总打码量
	FRoundBetsAvg     string `bson:"-"` // 局均打码量
	FWinRate          string `bson:"-"` // 胜率
	FRebateRate       string `bson:"-"` // 返奖率
	FWinBets          string `bson:"-"` // 胜局打码量
	FLoseBets         string `bson:"-"` // 败局打码量
	FTieBets          string `bson:"-"` // 平局打码量
	FWinBetsAvg       string `bson:"-"` // 胜局均码量
	FLoseBetsAvg      string `bson:"-"` // 败局均码量
	FTieBetsAvg       string `bson:"-"` // 平局均码量
	FWins             string `bson:"-"` // 胜局赢钱金额
	FLoses            string `bson:"-"` // 败局输钱金额
	FCash             string `bson:"-"` // 净赢
	FWinsAvg          string `bson:"-"` // 胜局均盈利
	FLosesAvg         string `bson:"-"` // 败局均输额
	FCashAvg          string `bson:"-"` // 局均净赢
	ObserveRoundsAvg  string `bson:"-"` // 局均观察次数
	ObserveRoundsRate string `bson:"-"` // 观察局数占比

	PlayerDragonsRate    string `bson:"-"` // 玩家押龙率
	PlayerTigersRate     string `bson:"-"` // 玩家押虎率
	PlayerTiesRate       string `bson:"-"` // 玩家押和率
	PlayerDragonsWinRate string `bson:"-"` // 玩家押龙率
	PlayerTigersWinRate  string `bson:"-"` // 玩家押虎率
	PlayerTiesWinRate    string `bson:"-"` // 玩家押和率
	PlayerMultisRate     string `bson:"-"` // 多门率

	FLoginedUsers       int    `bson:"-"` // 日活
	FLoginedNewUsers    int    `bson:"-"` // 新用户日活
	FLoginedOldUsers    int    `bson:"-"` // 老用户日活
	FPlayerRate         string `bson:"-"` // 玩crash人数占日活比
	FNewPlayerRate      string `bson:"-"` // 新顾客玩crash占新顾客日活比
	FOldPlayerRate      string `bson:"-"` // 老顾客玩crash占老顾客日活比
	FPlayerRoundsAvg    string `bson:"-"` // 人均局数
	FPlayerRoundsMedian int32  `bson:"-"` // 局数中位数
	FGameTimesAvg       string `bson:"-"` // 游戏局均时长（min)
	FGameTimesPlayerAvg string `bson:"-"` // 游戏人均时长（min)
	FPlayerBetsAvg      string `bson:"-"` // 人均打码
}

// 7updown 心想事成统计
type UpDownXxscStat struct {
	Id                    string   `bson:"_id"` // date-userid
	Date                  int64    `bson:"date"`
	DateStr               string   `bson:"date_str"`
	Userid                string   `bson:"userid"`
	StrategyPlayers       int32    `bson:"strategy_players"`        // 触发策略人数
	StrategyTimes         int32    `bson:"-"`                       // 触发策略总次数: u总 LHDXXSC.MaxTriggerTimes
	StrategyValidTimes    int32    `bson:"-"`                       // 触发策略有效次数: u LHDXXSC.TriggerTimes
	StrategyRounds        int32    `bson:"strategy_rounds"`         // 触发策略局数 r总
	StrategyDisturbRounds int32    `bson:"strategy_disturb_rounds"` // 扰动局数
	StrategyMultiRounds   int32    `bson:"strategy_multi_rounds"`   // 策略中多门局数
	StrategyBets          int64    `bson:"strategy_bets"`           // 策略中总打码量
	StrategyWinRounds     int64    `bson:"strategy_win_rounds"`     // 策略中总赢局数
	StrategyWins          int64    `bson:"strategy_wins"`           // 策略中总赢
	StrategyLoses         int64    `bson:"strategy_loses"`          // 策略中总输
	StrategyCash          int64    `bson:"-"`                       // 策略中净赢: 赢-输
	Userids               []string `bson:"userids"`                 // 汇总时的userid列表

	No          string `bson:"-"` // 序号
	FNickname   string `bson:"-"` // 总打码量
	FCtime      string `bson:"-"` // 注册时间
	FLoginTime  string `bson:"-"` // 最后登录时间
	FLiveDays   int    `bson:"-"` // 存活天数
	FLoseDays   int    `bson:"-"` // 流失天数
	FRegistArea string `bson:"-"` // 账号类型
	FUserType   string `bson:"-"` // 玩家类型

	StrategyValidTimesRate  string `bson:"-"` // 触发策略有效次数占触发策略总次数比
	StrategyValidTimesAvg   string `bson:"-"` // 人均有效次数
	StrategyRoundAvg        string `bson:"-"` // 每次触发平均所用局数
	StrategyValidRoundAvg   string `bson:"-"` // 每次有效触发平均所用局数
	StrategyDisturbRate     string `bson:"-"` // 扰动局数占比
	StrategyMultiRoundsRate string `bson:"-"` // 多门局数占比
	FStrategyBets           string `bson:"-"` // 策略中总打码量
	FStrategyBetRoundsAvg   string `bson:"-"` // 策略中局均打码量
	FStrategyWins           string `bson:"-"` // 策略中总赢
	FStrategyLoses          string `bson:"-"` // 策略中总输
	FStrategyCash           string `bson:"-"` // 策略中净赢: 赢-输
	FStrategyWinRate        string `bson:"-"` // 策略局胜率
	FStrategyRebateRate     string `bson:"-"` // 策略局返奖率
}

// 7UpDown 求死不能统计
type UpDownQsbnStat struct {
	Id                  string   `bson:"_id"` // date-userid
	Date                int64    `bson:"date"`
	DateStr             string   `bson:"date_str"`
	Userid              string   `bson:"userid"`
	StrategyPlayers     int32    `bson:"strategy_players"`      // 触发策略人数
	StrategyTimes       int32    `bson:"strategy_times"`        // 触发策略总次数
	AllInRounds         int32    `bson:"all_in_rounds"`         // allin局数
	StrategyMultiRounds int32    `bson:"strategy_multi_rounds"` // 策略中多门局数
	StrategyBets        int64    `bson:"strategy_bets"`         // 策略中总打码量
	StrategyWins        int64    `bson:"strategy_wins"`         // 策略中总赢
	StrategyLoses       int64    `bson:"strategy_loses"`        // 策略中总输
	StrategyCash        int64    `bson:"-"`                     // 策略中净赢: 赢-输
	BeforeBackRate      int      `bson:"before_back_rate"`      // 账变前返奖率
	AfterBackRate       int      `bson:"after_back_rate"`       // 账变后返奖率
	Userids             []string `bson:"userids"`               // 汇总时的userid列表

	No          string `bson:"-"` // 序号
	FNickname   string `bson:"-"` // 总打码量
	FCtime      string `bson:"-"` // 注册时间
	FLoginTime  string `bson:"-"` // 最后登录时间
	FLiveDays   int    `bson:"-"` // 存活天数
	FLoseDays   int    `bson:"-"` // 流失天数
	FRegistArea string `bson:"-"` // 账号类型
	FUserType   string `bson:"-"` // 玩家类型

	StrategyPlayersAvg string `bson:"-"` // 人均触发次数
	StrategyRate       string `bson:"-"` // 触发率
	StrategyMultiRate  string `bson:"-"` // 多门数占比
	FStrategyBets      string `bson:"-"` // 策略中总打码量
	FStrategyBetsAvg   string `bson:"-"` // 策略中局均打码量
	FStrategyWins      string `bson:"-"` // 策略中总赢
	FStrategyLoses     string `bson:"-"` // 策略中总输
	FStrategyCash      string `bson:"-"` // 策略中净赢
	FBeforeBackRate    string `bson:"-"` // 触发策略前返奖率
	FAfterBackRate     string `bson:"-"` // 策略局后返奖率
	FBackRate          string `bson:"-"` // 玩家当前返奖率
}

// 7updown 龙狂有祸
type UpdownLkyhStat struct {
	Id                  string   `bson:"_id"` // date-userid
	Date                int64    `bson:"date"`
	DateStr             string   `bson:"date_str"`
	Userid              string   `bson:"userid"`
	BSTimes             int      `json:"bstimes" bson:"bstimes"`                         // 倍杀状态次数
	BSDays              int      `json:"bsDays" bson:"bs_days"`                          // 倍杀天数
	TriggerTimes        int64    `json:"triggerTimes" bson:"trigger_times"`              // 每日倍杀次数 t日
	TZ                  int64    `json:"tz" bson:"tz"`                                   // 总倍杀次数 T总
	NZ                  int      `json:"nz" bson:"nz"`                                   // 总倍杀局数
	BSBets              int64    `json:"bsBets" bson:"bs_bets"`                          // 倍杀局打码量
	YZNumber            int64    `json:"yzNumber" bson:"yz_number"`                      // 压制次数
	YZRounds            int64    `json:"yzRounds" bson:"yz_rounds"`                      // 压制局数
	YZBets              int64    `json:"yzBets" bson:"yz_bets"`                          // 压制打码量
	RZ                  int      `json:"rz" bson:"rz"`                                   // r策
	StrategyMultiRounds int32    `bson:"strategy_multi_rounds"`                          // 策略中多门局数
	Bets                int64    `json:"bets" bson:"bets"`                               // 总打码量
	Rounds              int64    `json:"rounds" bson:"rounds"`                           // 总局数
	WinRounds           int64    `json:"winRounds" bson:"win_rounds"`                    // 赢局
	WinBets             int64    `json:"winBets" bson:"win_bets"`                        // 总赢
	LoseRounds          int64    `json:"loseRounds" bson:"lose_rounds"`                  // 输局
	LoseBets            int64    `json:"loseBets" bson:"lose_bets"`                      // 总输
	StrategyTimes       int32    `bson:"strategy_times"`                                 // 触发策略总次数
	StrategyBets        int64    `bson:"strategy_bets"`                                  // 策略中总打码量
	StrategyWins        int64    `bson:"strategy_wins"`                                  // 策略中总赢
	StrategyWinRounds   int64    `json:"strategyWinRounds" bson:"strategy_win_rounds"`   // 策略局赢局
	StrategyLoses       int64    `bson:"strategy_loses"`                                 // 策略中总输
	StrategyLoseRounds  int64    `json:"strategyLoseRounds" bson:"strategy_lose_rounds"` // 策略局输局
	StrategyCash        int64    `bson:"-"`                                              // 策略中净赢: 赢-输
	AD_BundleId         string   `bson:"ad__bundle_id"`                                  // 渠道
	Channel1            string   `bson:"channel1"`                                       // 渠道别名
	RegistArea          int      `bson:"regist_area"`                                    // 账号类型 ab测试 0:A 1:B 2:C
	Money               uint32   `bson:"money"`                                          // 充值总金额(分)
	Userids             []string `bson:"userids"`                                        // 汇总时的userid列表

	No          string `bson:"-"` // 序号
	FNickname   string `bson:"-"` // 总打码量
	FCtime      string `bson:"-"` // 注册时间
	FLoginTime  string `bson:"-"` // 最后登录时间
	FLiveDays   int    `bson:"-"` // 存活天数
	FLoseDays   int    `bson:"-"` // 流失天数
	FRegistArea string `bson:"-"` // 账号类型
	FUserType   string `bson:"-"` // 玩家类型

	BSNumberAvg        string `bson:"-"` // 日均倍杀次数
	BSRate             string `bson:"-"` // 倍杀触发率
	YZRate             string `bson:"-"` // 压制率
	StrategyMultiRate  string `bson:"-"` // 多门数占比
	BSBetAvg           string `bson:"-"` // 倍杀局均码量
	YZBetAvg           string `bson:"-"` // 压制局均码量
	FBets              string `bson:"-"` // 总打码
	RoundBetAvg        string `bson:"-"` // 局均码
	FStrategyBets      string `bson:"-"` // 策略中总打码量
	FStrategyBetsAvg   string `bson:"-"` // 策略中局均打码量
	FStrategyWins      string `bson:"-"` // 策略中总赢
	FStrategyLoses     string `bson:"-"` // 策略中总输
	FStrategyCash      string `bson:"-"` // 策略中净赢
	TotalWinRate       string `bson:"-"` // 总胜率
	StrategyWinRate    string `bson:"-"` // 策略局中胜率
	TotalRewardRate    string `bson:"-"` // 龙虎总返奖率
	StrategyRewardRate string `bson:"-"` // 策略局中返奖率
	PlayerRewardRate   string `bson:"-"` // 玩家大盘返奖率
}

/*
AB
*/
type ABPlayerStat struct {
	Id              string `bson:"_id"` // date-userid
	Date            int64  `bson:"date"`
	DateStr         string `bson:"date_str"`
	Userid          string `bson:"userid"`
	NewReg          bool   `bson:"new_reg"`                                  // 是新顾客
	AllRounds       int32  `bson:"all_rounds"`                               // 总局数
	GameTimes       int64  `bson:"game_times"`                               // 游戏总时长(秒)
	Bets            int64  `bson:"bets"`                                     // 总打码量
	BetRounds       int32  `bson:"bet_rounds"`                               // 下注局数
	WinRounds       int32  `bson:"win_rounds"`                               // 胜利局数
	LoseRounds      int32  `bson:"lose_rounds"`                              // 失败局数
	TieRounds       int32  `bson:"tie_rounds"`                               // 和局数
	WinBets         int64  `bson:"win_bets"`                                 // 胜局打码量
	LoseBets        int64  `bson:"lose_bets"`                                // 败局打码量
	TieBets         int64  `bson:"tie_bets"`                                 // 和局打码量
	Wins            int64  `bson:"wins"`                                     // 胜局赢钱金额
	Loses           int64  `bson:"loses"`                                    // 败局输钱金额
	Cash            int64  `bson:"cash"`                                     // 净赢
	SideWinner1     int64  `json:"sideWinner1" bson:"side_winner1"`          // 开奖位置ANDAR局数
	SideWinner2     int64  `json:"sideWinner2" bson:"side_winner2"`          // BAHAR
	SideWinner3     int64  `json:"sideWinner3" bson:"side_winner3"`          // 1-5
	SideWinner4     int64  `json:"sideWinner4" bson:"side_winner4"`          // 6-10
	SideWinner5     int64  `json:"sideWinner5" bson:"side_winner5"`          // 11-15
	SideWinner6     int64  `json:"sideWinner6" bson:"side_winner6"`          // 16-25
	SideWinner7     int64  `json:"sideWinner7" bson:"side_winner7"`          // 26-30
	SideWinner8     int64  `json:"sideWinner8" bson:"side_winner8"`          // 31-35
	SideWinner9     int64  `json:"sideWinner9" bson:"side_winner9"`          // 36-40
	SideWinner10    int64  `json:"sideWinner10" bson:"side_winner10"`        // 41以上
	SeatBets1       int64  `json:"seatBets1" bson:"seat_bets1"`              // 玩家下注ANDAR
	SeatBets2       int64  `json:"seatBets2" bson:"seat_bets2"`              // BAHAR
	SeatBets3       int64  `json:"seatBets3" bson:"seat_bets3"`              // 1-5
	SeatBets4       int64  `json:"seatBets4" bson:"seat_bets4"`              // 6-10
	SeatBets5       int64  `json:"seatBets5" bson:"seat_bets5"`              // 11-15
	SeatBets6       int64  `json:"seatBets6" bson:"seat_bets6"`              // 16-25
	SeatBets7       int64  `json:"seatBets7" bson:"seat_bets7"`              // 26-30
	SeatBets8       int64  `json:"seatBets8" bson:"seat_bets8"`              // 31-35
	SeatBets9       int64  `json:"seatBets9" bson:"seat_bets9"`              // 36-40
	SeatBets10      int64  `json:"seatBets10" bson:"seat_bets10"`            // 41以上
	PlayerWinSeat1  int64  `json:"playerWinSeat1" bson:"player_win_seat1"`   // 玩家下注位置ANDAR赢的局数
	PlayerWinSeat2  int64  `json:"playerWinSeat2" bson:"player_win_seat2"`   // BAHAR
	PlayerWinSeat3  int64  `json:"playerWinSeat3" bson:"player_win_seat3"`   // 1-5
	PlayerWinSeat4  int64  `json:"playerWinSeat4" bson:"player_win_seat4"`   // 6-10
	PlayerWinSeat5  int64  `json:"playerWinSeat5" bson:"player_win_seat5"`   // 11-15
	PlayerWinSeat6  int64  `json:"playerWinSeat6" bson:"player_win_seat6"`   // 16-25
	PlayerWinSeat7  int64  `json:"playerWinSeat7" bson:"player_win_seat7"`   // 26-30
	PlayerWinSeat8  int64  `json:"playerWinSeat8" bson:"player_win_seat8"`   // 31-35
	PlayerWinSeat9  int64  `json:"playerWinSeat9" bson:"player_win_seat9"`   // 36-40
	PlayerWinSeat10 int64  `json:"playerWinSeat10" bson:"player_win_seat10"` // 41以上
	PlayerMultis    int64  `bson:"player_multis"`                            // 多门局数 玩家下注2门及以上的局数
	RoundBetAvg     int64  `bson:"round_bet_avg"`                            // 局均码
	AD_BundleId     string `bson:"ad__bundle_id"`                            // 渠道
	Channel1        string `bson:"channel1"`                                 // 渠道别名
	RegistArea      int    `bson:"regist_area"`                              // 账号类型 ab测试 0:A 1:B 2:C
	Money           uint32 `bson:"money"`                                    // 充值总金额(分)
	ObserveRounds   int32  `bson:"observe_rounds"`                           // 观察局数

	Players       int32    `bson:"players"`         // 总玩crash人数
	NewPlayers    int32    `bson:"new_players"`     // 新顾客玩crash人数
	OldPlayers    int32    `bson:"old_players"`     // 老顾客玩crash人数
	AllRoundsList []int32  `bson:"all_rounds_list"` // 玩家局数list
	Userids       []string `bson:"userids"`         // 汇总时的userid列表

	No          string `bson:"-"` // 序号
	FNickname   string `bson:"-"` // 玩家昵称
	FCtime      string `bson:"-"` // 注册时间
	FLoginTime  string `bson:"-"` // 最后登录时间
	FLiveDays   int    `bson:"-"` // 存活天数
	FLoseDays   int    `bson:"-"` // 流失天数
	FRegistArea string `bson:"-"` // 账号类型
	FUserType   string `bson:"-"` // 玩家类型

	FMoney    string `bson:"-"` // 总充值
	FCashOut  string `bson:"-"` // 总提现
	FWinMoney string `bson:"-"` // 总赢

	SMoney   int   `bson:"-"` // 总汇充值
	SCashOut int   `bson:"-"` // 总汇提现
	SDiamond int64 `bson:"-"` // 总汇携带分

	FBets             string `bson:"-"` // 总打码量
	FRoundBetsAvg     string `bson:"-"` // 局均打码量
	FWinRate          string `bson:"-"` // 胜率
	FRebateRate       string `bson:"-"` // 返奖率
	FWinBets          string `bson:"-"` // 胜局打码量
	FLoseBets         string `bson:"-"` // 败局打码量
	FTieBets          string `bson:"-"` // 平局打码量
	FWinBetsAvg       string `bson:"-"` // 胜局均码量
	FLoseBetsAvg      string `bson:"-"` // 败局均码量
	FTieBetsAvg       string `bson:"-"` // 平局均码量
	FWins             string `bson:"-"` // 胜局赢钱金额
	FLoses            string `bson:"-"` // 败局输钱金额
	FCash             string `bson:"-"` // 净赢
	FWinsAvg          string `bson:"-"` // 胜局均盈利
	FLosesAvg         string `bson:"-"` // 败局均输额
	FCashAvg          string `bson:"-"` // 局均净赢
	ObserveRoundsAvg  string `bson:"-"` // 局均观察次数
	ObserveRoundsRate string `bson:"-"` // 观察局数占比

	SideWinnerRate1     string `bson:"-"` // 开奖位置ANDAR局数
	SideWinnerRate2     string `bson:"-"` // BAHAR
	SideWinnerRate3     string `bson:"-"` // 1-5
	SideWinnerRate4     string `bson:"-"` // 6-10
	SideWinnerRate5     string `bson:"-"` // 11-15
	SideWinnerRate6     string `bson:"-"` // 16-25
	SideWinnerRate7     string `bson:"-"` // 26-30
	SideWinnerRate8     string `bson:"-"` // 31-35
	SideWinnerRate9     string `bson:"-"` // 36-40
	SideWinnerRate10    string `bson:"-"` // 41以上
	SeatBetsRate1       string `bson:"-"` // 玩家下注ANDAR
	SeatBetsRate2       string `bson:"-"` // BAHAR
	SeatBetsRate3       string `bson:"-"` // 1-5
	SeatBetsRate4       string `bson:"-"` // 6-10
	SeatBetsRate5       string `bson:"-"` // 11-15
	SeatBetsRate6       string `bson:"-"` // 16-25
	SeatBetsRate7       string `bson:"-"` // 26-30
	SeatBetsRate8       string `bson:"-"` // 31-35
	SeatBetsRate9       string `bson:"-"` // 36-40
	SeatBetsRate10      string `bson:"-"` // 41以上
	PlayerWinSeatRate1  string `bson:"-"` // 玩家下注位置ANDAR赢的局数
	PlayerWinSeatRate2  string `bson:"-"` // BAHAR
	PlayerWinSeatRate3  string `bson:"-"` // 1-5
	PlayerWinSeatRate4  string `bson:"-"` // 6-10
	PlayerWinSeatRate5  string `bson:"-"` // 11-15
	PlayerWinSeatRate6  string `bson:"-"` // 16-25
	PlayerWinSeatRate7  string `bson:"-"` // 26-30
	PlayerWinSeatRate8  string `bson:"-"` // 31-35
	PlayerWinSeatRate9  string `bson:"-"` // 36-40
	PlayerWinSeatRate10 string `bson:"-"` // 41以上
	PlayerMultisRate    string `bson:"-"` // 多门率

	FLoginedUsers       int    `bson:"-"` // 日活
	FLoginedNewUsers    int    `bson:"-"` // 新用户日活
	FLoginedOldUsers    int    `bson:"-"` // 老用户日活
	FPlayerRate         string `bson:"-"` // 玩crash人数占日活比
	FNewPlayerRate      string `bson:"-"` // 新顾客玩crash占新顾客日活比
	FOldPlayerRate      string `bson:"-"` // 老顾客玩crash占老顾客日活比
	FPlayerRoundsAvg    string `bson:"-"` // 人均局数
	FPlayerRoundsMedian int32  `bson:"-"` // 局数中位数
	FGameTimesAvg       string `bson:"-"` // 游戏局均时长（min)
	FGameTimesPlayerAvg string `bson:"-"` // 游戏人均时长（min)
	FPlayerBetsAvg      string `bson:"-"` // 人均打码
}

// AB 安能求死
type ABAnqsStat struct {
	Id                  string   `bson:"_id"` // date-userid
	Date                int64    `bson:"date"`
	DateStr             string   `bson:"date_str"`
	Userid              string   `bson:"userid"`
	StrategyPlayers     int32    `bson:"strategy_players"`                             // 触发策略人数
	StrategyTimes       int32    `bson:"strategy_times"`                               // 触发策略总次数
	AllInRounds         int32    `bson:"all_in_rounds"`                                // allin局数
	StrategyMultiRounds int32    `bson:"strategy_multi_rounds"`                        // 策略中多门局数
	StrategyBets        int64    `bson:"strategy_bets"`                                // 策略中总打码量
	StrategyWinRounds   int64    `json:"strategyWinRounds" bson:"strategy_win_rounds"` // 策略局中赢的局数
	StrategyWins        int64    `bson:"strategy_wins"`                                // 策略中总赢
	StrategyLoseRounds  int64    `json:"strategyLoseRounds" bson:"strategy_lose_rounds"`
	StrategyLoses       int64    `bson:"strategy_loses"`                        // 策略中总输
	StrategyCash        int64    `bson:"-"`                                     // 策略中净赢: 赢-输
	Rounds              int64    `json:"rounds" bson:"rounds"`                  // AB总局数
	WinRounds           int64    `json:"winRounds" bson:"win_rounds"`           // AB总赢局数
	StrategyRounds      int64    `json:"strategyRounds" bson:"strategy_rounds"` // 触发策略总局数
	Wins                int64    `json:"wins" bson:"wins"`                      // 总赢
	Loses               int64    `json:"loses" bson:"loses"`                    // 总输
	AD_BundleId         string   `bson:"ad__bundle_id"`                         // 渠道
	Channel1            string   `bson:"channel1"`                              // 渠道别名
	RegistArea          int      `bson:"regist_area"`                           // 账号类型 ab测试 0:A 1:B 2:C
	Money               uint32   `bson:"money"`                                 // 充值总金额(分)
	Userids             []string `bson:"userids"`                               // 汇总时的userid列表

	No          string `bson:"-"` // 序号
	FNickname   string `bson:"-"` // 玩家昵称
	FCtime      string `bson:"-"` // 注册时间
	FLoginTime  string `bson:"-"` // 最后登录时间
	FLiveDays   int    `bson:"-"` // 存活天数
	FLoseDays   int    `bson:"-"` // 流失天数
	FRegistArea string `bson:"-"` // 账号类型
	FUserType   string `bson:"-"` // 玩家类型

	TriggerRoundsAvg      string `bson:"-"` //人均触发局数=触发策略总局数/触发策略人数
	TiggerRate            string `bson:"-"` // 触发率=触发策略总局数/allin局数
	ExpiredRate           string `bson:"-"` // 失效率=输的局数/触发策略总局数
	PlayerMultisRate      string `bson:"-"` // 多门率
	FStrategyBets         string `bson:"-"` // 策略中总打码量
	FStrategyBetRoundsAvg string `bson:"-"` // 策略中局均打码量
	FStrategyWins         string `bson:"-"` // 策略中总赢
	FStrategyLoses        string `bson:"-"` // 策略中总输
	FStrategyCash         string `bson:"-"` // 策略中净赢
	TotalWinRate          string `bson:"-"` // AB总胜率
	StrategyWinRate       string `bson:"-"` // 策略局中胜率
	TotalRewardRate       string `bson:"-"` // AB总返奖率
	StrategyRewardRate    string `bson:"-"` // 策略局中返奖率
	PlayerRewardRate      string `bson:"-"` // 玩家大盘返奖率

}

// AB 安然躺赢
type ABArtyStat struct {
	Id                  string   `bson:"_id"` // date-userid
	Date                int64    `bson:"date"`
	DateStr             string   `bson:"date_str"`
	Userid              string   `bson:"userid"`
	StrategyPlayers     int32    `bson:"strategy_players"`                             // 触发策略人数
	StrategyTimes       int32    `bson:"strategy_times"`                               // 触发策略总次数
	StrategyRounds      int64    `json:"strategyRounds" bson:"strategy_rounds"`        // 触发策略总局数
	TiggerTimes         int      `json:"tiggerTimes" bson:"tigger_times"`              // u 触发策略有效次数
	AllEvoTimes         int      `json:"allEvoTimes" bson:"all_evo_times"`             // 触发策略局数 r总
	StrategyMultiRounds int32    `bson:"strategy_multi_rounds"`                        // 策略中多门局数
	StrategyBets        int64    `bson:"strategy_bets"`                                // 策略中总打码量
	StrategyWins        int64    `bson:"strategy_wins"`                                // 策略中总赢
	StrategyWinRounds   int64    `json:"strategyWinRounds" bson:"strategy_win_rounds"` // 策略局中赢的局数
	StrategyLoses       int64    `bson:"strategy_loses"`                               // 策略中总输
	StrategyLoseRounds  int64    `json:"strategyLoseRounds" bson:"strategy_lose_rounds"`
	StrategyCash        int64    `bson:"-"`             // 策略中净赢: 赢-输
	AD_BundleId         string   `bson:"ad__bundle_id"` // 渠道
	Channel1            string   `bson:"channel1"`      // 渠道别名
	RegistArea          int      `bson:"regist_area"`   // 账号类型 ab测试 0:A 1:B 2:C
	Money               uint32   `bson:"money"`         // 充值总金额(分)
	Userids             []string `bson:"userids"`       // 汇总时的userid列表

	No          string `bson:"-"` // 序号
	FNickname   string `bson:"-"` // 玩家昵称
	FCtime      string `bson:"-"` // 注册时间
	FLoginTime  string `bson:"-"` // 最后登录时间
	FLiveDays   int    `bson:"-"` // 存活天数
	FLoseDays   int    `bson:"-"` // 流失天数
	FRegistArea string `bson:"-"` // 账号类型
	FUserType   string `bson:"-"` // 玩家类型

	TriggerRate           string `bson:"-"` // 策略有效次数占总触发次数比=触发策略有效次数/触发策略总次数
	PerCapitaAvg          string `bson:"-"` // 人均有效次数=触发策略有效次数/触发策略人数
	TiggerStrategyAvg     string `bson:"-"` // 每次触发平均所用局数=触发策略局数/触发策略总次数
	YxTiggerStrategyAvg   string `bson:"-"` // 每次有效触发平均所用局数 = 触发策略局数/触发策略有效次数
	PlayerMultisRate      string `bson:"-"` // 多门率
	FStrategyBets         string `bson:"-"` // 策略中总打码量
	FStrategyBetRoundsAvg string `bson:"-"` // 策略中局均打码量
	FStrategyWins         string `bson:"-"` // 策略中总赢
	FStrategyLoses        string `bson:"-"` // 策略中总输
	FStrategyCash         string `bson:"-"` // 策略中净赢
	FStrategyWinRate      string `bson:"-"` // 策略局胜率=策略局中玩家赢钱的局/触发策略局数
	FStrategyRebateRate   string `bson:"-"` // 策略局返奖率=策略局总赢/策略中总输
}

/*
CP
*/
type CPPlayerStat struct {
	Id             string `bson:"_id"` // date-userid
	Date           int64  `bson:"date"`
	DateStr        string `bson:"date_str"`
	Userid         string `bson:"userid"`
	NewReg         bool   `bson:"new_reg"`                                // 是新顾客
	AllRounds      int32  `bson:"all_rounds"`                             // 总局数
	GameTimes      int64  `bson:"game_times"`                             // 游戏总时长(秒)
	Bets           int64  `bson:"bets"`                                   // 总打码量
	BetRounds      int32  `bson:"bet_rounds"`                             // 下注局数
	WinRounds      int32  `bson:"win_rounds"`                             // 胜利局数
	LoseRounds     int32  `bson:"lose_rounds"`                            // 失败局数
	TieRounds      int32  `bson:"tie_rounds"`                             // 和局数
	WinBets        int64  `bson:"win_bets"`                               // 胜局打码量
	LoseBets       int64  `bson:"lose_bets"`                              // 败局打码量
	TieBets        int64  `bson:"tie_bets"`                               // 和局打码量
	Wins           int64  `bson:"wins"`                                   // 胜局赢钱金额
	Loses          int64  `bson:"loses"`                                  // 败局输钱金额
	Cash           int64  `bson:"cash"`                                   // 净赢
	SideWinner1    int64  `json:"sideWinner1" bson:"side_winner1"`        // 开奖位置High局数
	SideWinner2    int64  `json:"sideWinner2" bson:"side_winner2"`        // Pair
	SideWinner3    int64  `json:"sideWinner3" bson:"side_winner3"`        // Color
	SideWinner4    int64  `json:"sideWinner4" bson:"side_winner4"`        // seq
	SideWinner5    int64  `json:"sideWinner5" bson:"side_winner5"`        // pureseq
	SideWinner6    int64  `json:"sideWinner6" bson:"side_winner6"`        // set
	SeatBets1      int64  `json:"seatBets1" bson:"seat_bets1"`            // 玩家下注High
	SeatBets2      int64  `json:"seatBets2" bson:"seat_bets2"`            // Pair
	SeatBets3      int64  `json:"seatBets3" bson:"seat_bets3"`            // Color
	SeatBets4      int64  `json:"seatBets4" bson:"seat_bets4"`            // seq
	SeatBets5      int64  `json:"seatBets5" bson:"seat_bets5"`            // pureseq
	SeatBets6      int64  `json:"seatBets6" bson:"seat_bets6"`            // set
	PlayerWinSeat1 int64  `json:"playerWinSeat1" bson:"player_win_seat1"` // 玩家下注位置High赢的局数
	PlayerWinSeat2 int64  `json:"playerWinSeat2" bson:"player_win_seat2"` // Pair
	PlayerWinSeat3 int64  `json:"playerWinSeat3" bson:"player_win_seat3"` // Color
	PlayerWinSeat4 int64  `json:"playerWinSeat4" bson:"player_win_seat4"` // seq
	PlayerWinSeat5 int64  `json:"playerWinSeat5" bson:"player_win_seat5"` // pureseq
	PlayerWinSeat6 int64  `json:"playerWinSeat6" bson:"player_win_seat6"` // set
	PlayerMultis   int64  `bson:"player_multis"`                          // 多门局数 玩家下注2门及以上的局数
	RoundBetAvg    int64  `bson:"round_bet_avg"`                          // 局均码
	AD_BundleId    string `bson:"ad__bundle_id"`                          // 渠道
	Channel1       string `bson:"channel1"`                               // 渠道别名
	RegistArea     int    `bson:"regist_area"`                            // 账号类型 ab测试 0:A 1:B 2:C
	Money          uint32 `bson:"money"`                                  // 充值总金额(分)
	ObserveRounds  int32  `bson:"observe_rounds"`                         // 观察局数

	Players       int32    `bson:"players"`         // 总玩cp人数
	NewPlayers    int32    `bson:"new_players"`     // 新顾客玩cp人数
	OldPlayers    int32    `bson:"old_players"`     // 老顾客玩cp人数
	AllRoundsList []int32  `bson:"all_rounds_list"` // 玩家局数list
	Userids       []string `bson:"userids"`         // 汇总时的userid列表

	No          string `bson:"-"` // 序号
	FNickname   string `bson:"-"` // 玩家昵称
	FCtime      string `bson:"-"` // 注册时间
	FLoginTime  string `bson:"-"` // 最后登录时间
	FLiveDays   int    `bson:"-"` // 存活天数
	FLoseDays   int    `bson:"-"` // 流失天数
	FRegistArea string `bson:"-"` // 账号类型
	FUserType   string `bson:"-"` // 玩家类型

	FMoney    string `bson:"-"` // 总充值
	FCashOut  string `bson:"-"` // 总提现
	FWinMoney string `bson:"-"` // 总赢

	SMoney   int   `bson:"-"` // 总汇充值
	SCashOut int   `bson:"-"` // 总汇提现
	SDiamond int64 `bson:"-"` // 总汇携带分

	FBets             string `bson:"-"` // 总打码量
	FRoundBetsAvg     string `bson:"-"` // 局均打码量
	FWinRate          string `bson:"-"` // 胜率
	FRebateRate       string `bson:"-"` // 返奖率
	FWinBets          string `bson:"-"` // 胜局打码量
	FLoseBets         string `bson:"-"` // 败局打码量
	FTieBets          string `bson:"-"` // 平局打码量
	FWinBetsAvg       string `bson:"-"` // 胜局均码量
	FLoseBetsAvg      string `bson:"-"` // 败局均码量
	FTieBetsAvg       string `bson:"-"` // 平局均码量
	FWins             string `bson:"-"` // 胜局赢钱金额
	FLoses            string `bson:"-"` // 败局输钱金额
	FCash             string `bson:"-"` // 净赢
	FWinsAvg          string `bson:"-"` // 胜局均盈利
	FLosesAvg         string `bson:"-"` // 败局均输额
	FCashAvg          string `bson:"-"` // 局均净赢
	ObserveRoundsAvg  string `bson:"-"` // 局均观察次数
	ObserveRoundsRate string `bson:"-"` // 观察局数占比

	SideWinnerRate1    string `bson:"-"` // 开奖位置High局数
	SideWinnerRate2    string `bson:"-"` // Pair
	SideWinnerRate3    string `bson:"-"` // Color
	SideWinnerRate4    string `bson:"-"` // seq
	SideWinnerRate5    string `bson:"-"` // pureseq
	SideWinnerRate6    string `bson:"-"` // set
	SeatBetsRate1      string `bson:"-"` // 玩家下注High
	SeatBetsRate2      string `bson:"-"` // Pair
	SeatBetsRate3      string `bson:"-"` // Color
	SeatBetsRate4      string `bson:"-"` // seq
	SeatBetsRate5      string `bson:"-"` // pureseq
	SeatBetsRate6      string `bson:"-"` // set
	PlayerWinSeatRate1 string `bson:"-"` // 玩家下注位置High赢的局数
	PlayerWinSeatRate2 string `bson:"-"` // Pair
	PlayerWinSeatRate3 string `bson:"-"` // Color
	PlayerWinSeatRate4 string `bson:"-"` // seq
	PlayerWinSeatRate5 string `bson:"-"` // pureseq
	PlayerWinSeatRate6 string `bson:"-"` // set
	PlayerMultisRate   string `bson:"-"` // 多门率

	FLoginedUsers       int    `bson:"-"` // 日活
	FLoginedNewUsers    int    `bson:"-"` // 新用户日活
	FLoginedOldUsers    int    `bson:"-"` // 老用户日活
	FPlayerRate         string `bson:"-"` // 玩cp人数占日活比
	FNewPlayerRate      string `bson:"-"` // 新顾客玩cp占新顾客日活比
	FOldPlayerRate      string `bson:"-"` // 老顾客玩cp占老顾客日活比
	FPlayerRoundsAvg    string `bson:"-"` // 人均局数
	FPlayerRoundsMedian int32  `bson:"-"` // 局数中位数
	FGameTimesAvg       string `bson:"-"` // 游戏局均时长（min)
	FGameTimesPlayerAvg string `bson:"-"` // 游戏人均时长（min)
	FPlayerBetsAvg      string `bson:"-"` // 人均打码
}

// CP 来玩就赢
type CPLwjyStat struct {
	Id                  string   `bson:"_id"` // date-userid
	Date                int64    `bson:"date"`
	DateStr             string   `bson:"date_str"`
	Userid              string   `bson:"userid"`
	StrategyPlayers     int32    `bson:"strategy_players"`                             // 触发策略人数
	StrategyTimes       int32    `bson:"strategy_times"`                               // 触发策略总次数
	TiggerTimes         int      `json:"tiggerTimes" bson:"tigger_times"`              // u 触发策略有效次数
	AllEvoTimes         int      `json:"allEvoTimes" bson:"all_evo_times"`             // 触发策略局数 r总
	StrategyMultiRounds int32    `bson:"strategy_multi_rounds"`                        // 策略中多门局数
	StrategyBets        int64    `bson:"strategy_bets"`                                // 策略中总打码量
	StrategyWinRounds   int64    `json:"strategyWinRounds" bson:"strategy_win_rounds"` // 策略局中赢的局数
	StrategyWins        int64    `bson:"strategy_wins"`                                // 策略中总赢
	StrategyLoseRounds  int64    `json:"strategyLoseRounds" bson:"strategy_lose_rounds"`
	StrategyLoses       int64    `bson:"strategy_loses"` // 策略中总输
	StrategyCash        int64    `bson:"-"`              // 策略中净赢: 赢-输
	AD_BundleId         string   `bson:"ad__bundle_id"`  // 渠道
	Channel1            string   `bson:"channel1"`       // 渠道别名
	RegistArea          int      `bson:"regist_area"`    // 账号类型 ab测试 0:A 1:B 2:C
	Money               uint32   `bson:"money"`          // 充值总金额(分)
	Userids             []string `bson:"userids"`        // 汇总时的userid列表

	No          string `bson:"-"` // 序号
	FNickname   string `bson:"-"` // 玩家昵称
	FCtime      string `bson:"-"` // 注册时间
	FLoginTime  string `bson:"-"` // 最后登录时间
	FLiveDays   int    `bson:"-"` // 存活天数
	FLoseDays   int    `bson:"-"` // 流失天数
	FRegistArea string `bson:"-"` // 账号类型
	FUserType   string `bson:"-"` // 玩家类型

	TriggerRate           string `bson:"-"` // 策略有效次数占总触发次数比=触发策略有效次数/触发策略总次数
	PerCapitaAvg          string `bson:"-"` // 人均有效次数=触发策略有效次数/触发策略人数
	TiggerStrategyAvg     string `bson:"-"` // 每次触发平均所用局数=触发策略局数/触发策略总次数
	YxTiggerStrategyAvg   string `bson:"-"` // 每次有效触发平均所用局数 = 触发策略局数/触发策略有效次数
	PlayerMultisRate      string `bson:"-"` // 多门率
	FStrategyBets         string `bson:"-"` // 策略中总打码量
	FStrategyBetRoundsAvg string `bson:"-"` // 策略中局均打码量
	FStrategyWins         string `bson:"-"` // 策略中总赢
	FStrategyLoses        string `bson:"-"` // 策略中总输
	FStrategyCash         string `bson:"-"` // 策略中净赢
	FStrategyWinRate      string `bson:"-"` // 策略局胜率=策略局中玩家赢钱的局/触发策略局数
	FStrategyRebateRate   string `bson:"-"` // 策略局返奖率=策略局总赢/策略中总输

}

// CP 来易去难
type CPLyqnStat struct {
	Id                  string   `bson:"_id"` // date-userid
	Date                int64    `bson:"date"`
	DateStr             string   `bson:"date_str"`
	Userid              string   `bson:"userid"`
	StrategyPlayers     int32    `bson:"strategy_players"`                             // 触发策略人数
	StrategyRounds      int64    `json:"strategyRounds" bson:"strategy_rounds"`        // 触发策略总局数
	AllInRounds         int32    `bson:"all_in_rounds"`                                // allin局数
	StrategyMultiRounds int32    `bson:"strategy_multi_rounds"`                        // 策略中多门局数
	StrategyBets        int64    `bson:"strategy_bets"`                                // 策略中总打码量
	StrategyWins        int64    `bson:"strategy_wins"`                                // 策略中总赢
	StrategyWinRounds   int64    `json:"strategyWinRounds" bson:"strategy_win_rounds"` // 策略局中赢的局数
	StrategyLoses       int64    `bson:"strategy_loses"`                               // 策略中总输
	StrategyLoseRounds  int64    `json:"strategyLoseRounds" bson:"strategy_lose_rounds"`
	StrategyCash        int64    `bson:"-"`             // 策略中净赢: 赢-输
	BetRounds           int32    `bson:"bet_rounds"`    // 下注局数
	WinRounds           int32    `bson:"win_rounds"`    // 胜利局数
	LoseRounds          int32    `bson:"lose_rounds"`   // 失败局数
	Wins                int64    `bson:"wins"`          // 胜局赢钱金额
	Loses               int64    `bson:"loses"`         // 败局输钱金额
	AD_BundleId         string   `bson:"ad__bundle_id"` // 渠道
	Channel1            string   `bson:"channel1"`      // 渠道别名
	RegistArea          int      `bson:"regist_area"`   // 账号类型 ab测试 0:A 1:B 2:C
	Money               uint32   `bson:"money"`         // 充值总金额(分)
	Userids             []string `bson:"userids"`       // 汇总时的userid列表

	No          string `bson:"-"` // 序号
	FNickname   string `bson:"-"` // 玩家昵称
	FCtime      string `bson:"-"` // 注册时间
	FLoginTime  string `bson:"-"` // 最后登录时间
	FLiveDays   int    `bson:"-"` // 存活天数
	FLoseDays   int    `bson:"-"` // 流失天数
	FRegistArea string `bson:"-"` // 账号类型
	FUserType   string `bson:"-"` // 玩家类型

	TriggerRoundsAvg      string `bson:"-"` //人均触发局数=触发策略总局数/触发策略人数
	TiggerRate            string `bson:"-"` // 触发率=触发策略总局数/allin局数
	ExpiredRate           string `bson:"-"` // 失效率=输的局数/触发策略总局数
	PlayerMultisRate      string `bson:"-"` // 多门率
	FStrategyBets         string `bson:"-"` // 策略中总打码量
	FStrategyBetRoundsAvg string `bson:"-"` // 策略中局均打码量
	FStrategyWins         string `bson:"-"` // 策略中总赢
	FStrategyLoses        string `bson:"-"` // 策略中总输
	FStrategyCash         string `bson:"-"` // 策略中净赢
	TotalWinRate          string `bson:"-"` // AB总胜率
	StrategyWinRate       string `bson:"-"` // 策略局中胜率
	TotalRewardRate       string `bson:"-"` // AB总返奖率
	StrategyRewardRate    string `bson:"-"` // 策略局中返奖率
	PlayerRewardRate      string `bson:"-"` // 玩家大盘返奖率
}

/*
RB
*/
type RBPlayerStat struct {
	Id             string `bson:"_id"` // date-userid
	Date           int64  `bson:"date"`
	DateStr        string `bson:"date_str"`
	Userid         string `bson:"userid"`
	NewReg         bool   `bson:"new_reg"`                                // 是新顾客
	AllRounds      int32  `bson:"all_rounds"`                             // 总局数
	GameTimes      int64  `bson:"game_times"`                             // 游戏总时长(秒)
	Bets           int64  `bson:"bets"`                                   // 总打码量
	BetRounds      int32  `bson:"bet_rounds"`                             // 下注局数
	WinRounds      int32  `bson:"win_rounds"`                             // 胜利局数
	LoseRounds     int32  `bson:"lose_rounds"`                            // 失败局数
	TieRounds      int32  `bson:"tie_rounds"`                             // 和局数
	WinBets        int64  `bson:"win_bets"`                               // 胜局打码量
	LoseBets       int64  `bson:"lose_bets"`                              // 败局打码量
	TieBets        int64  `bson:"tie_bets"`                               // 和局打码量
	Wins           int64  `bson:"wins"`                                   // 胜局赢钱金额
	Loses          int64  `bson:"loses"`                                  // 败局输钱金额
	Cash           int64  `bson:"cash"`                                   // 净赢
	SideWinner1    int64  `json:"sideWinner1" bson:"side_winner1"`        // 下注局开blue率
	SideWinner2    int64  `json:"sideWinner2" bson:"side_winner2"`        // 下注局开red率
	SideWinner3    int64  `json:"sideWinner3" bson:"side_winner3"`        // 下注局开Luckyshot率
	SideWinner4    int64  `json:"sideWinner4" bson:"side_winner4"`        // 下注局开9-Apair率
	SideWinner5    int64  `json:"sideWinner5" bson:"side_winner5"`        // 下注局开color率
	SideWinner6    int64  `json:"sideWinner6" bson:"side_winner6"`        // 下注局开seq率
	SideWinner7    int64  `json:"sideWinner7" bson:"side_winner7"`        // 下注局开pureseq率
	SideWinner8    int64  `json:"sideWinner8" bson:"side_winner8"`        // 下注局开set率
	SeatBets1      int64  `json:"seatBets1" bson:"seat_bets1"`            // 玩家押blue率
	SeatBets2      int64  `json:"seatBets2" bson:"seat_bets2"`            // 玩家押red率
	SeatBets3      int64  `json:"seatBets3" bson:"seat_bets3"`            // 玩家押Luckyshot率
	PlayerWinSeat1 int64  `json:"playerWinSeat1" bson:"player_win_seat1"` // 玩家blue中率
	PlayerWinSeat2 int64  `json:"playerWinSeat2" bson:"player_win_seat2"` // 玩家red中率
	PlayerWinSeat3 int64  `json:"playerWinSeat3" bson:"player_win_seat3"` // 玩家luckyshot中率
	PlayerWinSeat4 int64  `json:"playerWinSeat4" bson:"player_win_seat4"` // 玩家9-Apair中率
	PlayerWinSeat5 int64  `json:"playerWinSeat5" bson:"player_win_seat5"` // 玩家color中率
	PlayerWinSeat6 int64  `json:"playerWinSeat6" bson:"player_win_seat6"` // 玩家seq中率
	PlayerWinSeat7 int64  `json:"playerWinSeat7" bson:"player_win_seat7"` // 玩家pureseq中率
	PlayerWinSeat8 int64  `json:"playerWinSeat8" bson:"player_win_seat8"` // 玩家set中率
	PlayerMultis   int64  `bson:"player_multis"`                          // 多门局数 玩家下注2门及以上的局数
	RoundBetAvg    int64  `bson:"round_bet_avg"`                          // 局均码
	AD_BundleId    string `bson:"ad__bundle_id"`                          // 渠道
	Channel1       string `bson:"channel1"`                               // 渠道别名
	RegistArea     int    `bson:"regist_area"`                            // 账号类型 ab测试 0:A 1:B 2:C
	Money          uint32 `bson:"money"`                                  // 充值总金额(分)
	ObserveRounds  int32  `bson:"observe_rounds"`                         // 观察局数

	Players       int32    `bson:"players"`         // 总玩cp人数
	NewPlayers    int32    `bson:"new_players"`     // 新顾客玩cp人数
	OldPlayers    int32    `bson:"old_players"`     // 老顾客玩cp人数
	AllRoundsList []int32  `bson:"all_rounds_list"` // 玩家局数list
	Userids       []string `bson:"userids"`         // 汇总时的userid列表

	No          string `bson:"-"` // 序号
	FNickname   string `bson:"-"` // 玩家昵称
	FCtime      string `bson:"-"` // 注册时间
	FLoginTime  string `bson:"-"` // 最后登录时间
	FLiveDays   int    `bson:"-"` // 存活天数
	FLoseDays   int    `bson:"-"` // 流失天数
	FRegistArea string `bson:"-"` // 账号类型
	FUserType   string `bson:"-"` // 玩家类型

	FMoney    string `bson:"-"` // 总充值
	FCashOut  string `bson:"-"` // 总提现
	FWinMoney string `bson:"-"` // 总赢

	SMoney   int   `bson:"-"` // 总汇充值
	SCashOut int   `bson:"-"` // 总汇提现
	SDiamond int64 `bson:"-"` // 总汇携带分

	FBets               string `bson:"-"` // 总打码量
	FRoundBetsAvg       string `bson:"-"` // 局均打码量
	FWinRate            string `bson:"-"` // 胜率
	FRebateRate         string `bson:"-"` // 返奖率
	FWinBets            string `bson:"-"` // 胜局打码量
	FLoseBets           string `bson:"-"` // 败局打码量
	FTieBets            string `bson:"-"` // 平局打码量
	FWinBetsAvg         string `bson:"-"` // 胜局均码量
	FLoseBetsAvg        string `bson:"-"` // 败局均码量
	FTieBetsAvg         string `bson:"-"` // 平局均码量
	FWins               string `bson:"-"` // 胜局赢钱金额
	FLoses              string `bson:"-"` // 败局输钱金额
	FCash               string `bson:"-"` // 净赢
	FWinsAvg            string `bson:"-"` // 胜局均盈利
	FLosesAvg           string `bson:"-"` // 败局均输额
	FCashAvg            string `bson:"-"` // 局均净赢
	ObserveRoundsAvg    string `bson:"-"` // 局均观察次数
	ObserveRoundsRate   string `bson:"-"` // 观察局数占比
	SideWinnerRate1     string `bson:"-"` // 下注局开blue率
	SideWinnerRate2     string `bson:"-"` // 下注局开red率
	SideWinnerRate3     string `bson:"-"` // 下注局开Luckyshot率
	SideWinnerRate4     string `bson:"-"` // 下注局开9-Apair率
	SideWinnerRate5     string `bson:"-"` // 下注局开color率
	SideWinnerRate6     string `bson:"-"` // 下注局开seq率
	SideWinnerRate7     string `bson:"-"` // 下注局开pureseq率
	SideWinnerRate8     string `bson:"-"` // 下注局开set率
	SeatBetsRate1       string `bson:"-"` // 玩家押blue率
	SeatBetsRate2       string `bson:"-"` // 玩家押red率
	SeatBetsRate3       string `bson:"-"` // 玩家押Luckyshot率
	PlayerWinSeatRate1  string `bson:"-"` // 玩家blue中率
	PlayerWinSeatRate2  string `bson:"-"` // 玩家red中率
	PlayerWinSeatRate3  string `bson:"-"` // 玩家luckyshot中率
	PlayerWinSeatRate4  string `bson:"-"` // 玩家9-Apair中率
	PlayerWinSeatRate5  string `bson:"-"` // 玩家color中率
	PlayerWinSeatRate6  string `bson:"-"` // 玩家seq中率
	PlayerWinSeatRate7  string `bson:"-"` // 玩家pureseq中率
	PlayerWinSeatRate8  string `bson:"-"` // 玩家set中率
	PlayerMultisRate    string `bson:"-"` // 多门率
	FLoginedUsers       int    `bson:"-"` // 日活
	FLoginedNewUsers    int    `bson:"-"` // 新用户日活
	FLoginedOldUsers    int    `bson:"-"` // 老用户日活
	FPlayerRate         string `bson:"-"` // 玩cp人数占日活比
	FNewPlayerRate      string `bson:"-"` // 新顾客玩cp占新顾客日活比
	FOldPlayerRate      string `bson:"-"` // 老顾客玩cp占老顾客日活比
	FPlayerRoundsAvg    string `bson:"-"` // 人均局数
	FPlayerRoundsMedian int32  `bson:"-"` // 局数中位数
	FGameTimesAvg       string `bson:"-"` // 游戏局均时长（min)
	FGameTimesPlayerAvg string `bson:"-"` // 游戏人均时长（min)
	FPlayerBetsAvg      string `bson:"-"` // 人均打码
}

// RB 红运当头
type RBHydtStat struct {
	Id                  string   `bson:"_id"` // date-userid
	Date                int64    `bson:"date"`
	DateStr             string   `bson:"date_str"`
	Userid              string   `bson:"userid"`
	StrategyPlayers     int32    `bson:"strategy_players"`                             // 触发策略人数
	StrategyTimes       int32    `bson:"strategy_times"`                               // 触发策略总次数
	TiggerTimes         int      `json:"tiggerTimes" bson:"tigger_times"`              // u 触发策略有效次数
	AllEvoTimes         int      `json:"allEvoTimes" bson:"all_evo_times"`             // 触发策略局数 r总
	StrategyMultiRounds int32    `bson:"strategy_multi_rounds"`                        // 策略中多门局数
	StrategyBets        int64    `bson:"strategy_bets"`                                // 策略中总打码量
	StrategyWinRounds   int64    `json:"strategyWinRounds" bson:"strategy_win_rounds"` // 策略局中赢的局数
	StrategyWins        int64    `bson:"strategy_wins"`                                // 策略中总赢
	StrategyLoseRounds  int64    `json:"strategyLoseRounds" bson:"strategy_lose_rounds"`
	StrategyLoses       int64    `bson:"strategy_loses"` // 策略中总输
	StrategyCash        int64    `bson:"-"`              // 策略中净赢: 赢-输
	AD_BundleId         string   `bson:"ad__bundle_id"`  // 渠道
	Channel1            string   `bson:"channel1"`       // 渠道别名
	RegistArea          int      `bson:"regist_area"`    // 账号类型 ab测试 0:A 1:B 2:C
	Money               uint32   `bson:"money"`          // 充值总金额(分)
	Userids             []string `bson:"userids"`        // 汇总时的userid列表

	No          string `bson:"-"` // 序号
	FNickname   string `bson:"-"` // 玩家昵称
	FCtime      string `bson:"-"` // 注册时间
	FLoginTime  string `bson:"-"` // 最后登录时间
	FLiveDays   int    `bson:"-"` // 存活天数
	FLoseDays   int    `bson:"-"` // 流失天数
	FRegistArea string `bson:"-"` // 账号类型
	FUserType   string `bson:"-"` // 玩家类型

	TriggerRate           string `bson:"-"` // 策略有效次数占总触发次数比=触发策略有效次数/触发策略总次数
	PerCapitaAvg          string `bson:"-"` // 人均有效次数=触发策略有效次数/触发策略人数
	TiggerStrategyAvg     string `bson:"-"` // 每次触发平均所用局数=触发策略局数/触发策略总次数
	YxTiggerStrategyAvg   string `bson:"-"` // 每次有效触发平均所用局数 = 触发策略局数/触发策略有效次数
	PlayerMultisRate      string `bson:"-"` // 多门率
	FStrategyBets         string `bson:"-"` // 策略中总打码量
	FStrategyBetRoundsAvg string `bson:"-"` // 策略中局均打码量
	FStrategyWins         string `bson:"-"` // 策略中总赢
	FStrategyLoses        string `bson:"-"` // 策略中总输
	FStrategyCash         string `bson:"-"` // 策略中净赢
	FStrategyWinRate      string `bson:"-"` // 策略局胜率=策略局中玩家赢钱的局/触发策略局数
	FStrategyRebateRate   string `bson:"-"` // 策略局返奖率=策略局总赢/策略中总输

}

// RB 绝处逢生
type RBJcfsStat struct {
	Id                  string `bson:"_id"` // date-userid
	Date                int64  `bson:"date"`
	DateStr             string `bson:"date_str"`
	Userid              string `bson:"userid"`
	StrategyPlayers     int32  `bson:"strategy_players"`                             // 触发策略人数
	StrategyRounds      int64  `json:"strategyRounds" bson:"strategy_rounds"`        // 触发策略总局数
	AllInRounds         int32  `bson:"all_in_rounds"`                                // allin局数
	StrategyMultiRounds int32  `bson:"strategy_multi_rounds"`                        // 策略中多门局数
	StrategyBets        int64  `bson:"strategy_bets"`                                // 策略中总打码量
	StrategyWins        int64  `bson:"strategy_wins"`                                // 策略中总赢
	StrategyWinRounds   int64  `json:"strategyWinRounds" bson:"strategy_win_rounds"` // 策略局中赢的局数
	StrategyLoses       int64  `bson:"strategy_loses"`                               // 策略中总输
	StrategyLoseRounds  int64  `json:"strategyLoseRounds" bson:"strategy_lose_rounds"`
	StrategyCash        int64  `bson:"-"`             // 策略中净赢: 赢-输
	BetRounds           int32  `bson:"bet_rounds"`    // 下注局数
	WinRounds           int32  `bson:"win_rounds"`    // 胜利局数
	LoseRounds          int32  `bson:"lose_rounds"`   // 失败局数
	Wins                int64  `bson:"wins"`          // 胜局赢钱金额
	Loses               int64  `bson:"loses"`         // 败局输钱金额
	AD_BundleId         string `bson:"ad__bundle_id"` // 渠道
	Channel1            string `bson:"channel1"`      // 渠道别名
	RegistArea          int    `bson:"regist_area"`   // 账号类型 ab测试 0:A 1:B 2:C
	Money               uint32 `bson:"money"`         // 充值总金额(分)

	No          string   `bson:"-"`       // 序号
	FNickname   string   `bson:"-"`       // 玩家昵称
	FCtime      string   `bson:"-"`       // 注册时间
	FLoginTime  string   `bson:"-"`       // 最后登录时间
	FLiveDays   int      `bson:"-"`       // 存活天数
	FLoseDays   int      `bson:"-"`       // 流失天数
	FRegistArea string   `bson:"-"`       // 账号类型
	FUserType   string   `bson:"-"`       // 玩家类型
	Userids     []string `bson:"userids"` // 汇总时的userid列表

	TriggerRoundsAvg      string `bson:"-"` //人均触发局数=触发策略总局数/触发策略人数
	TiggerRate            string `bson:"-"` // 触发率=触发策略总局数/allin局数
	ExpiredRate           string `bson:"-"` // 失效率=输的局数/触发策略总局数
	PlayerMultisRate      string `bson:"-"` // 多门率
	FStrategyBets         string `bson:"-"` // 策略中总打码量
	FStrategyBetRoundsAvg string `bson:"-"` // 策略中局均打码量
	FStrategyWins         string `bson:"-"` // 策略中总赢
	FStrategyLoses        string `bson:"-"` // 策略中总输
	FStrategyCash         string `bson:"-"` // 策略中净赢
	TotalWinRate          string `bson:"-"` // RB总胜率
	StrategyWinRate       string `bson:"-"` // 策略局中胜率
	StrategyRewardRate    string `bson:"-"` // 策略局中返奖率
	TotalRewardRate       string `bson:"-"` // RB总返奖率
	PlayerRewardRate      string `bson:"-"` // 玩家大盘返奖率
}

type TpNewStatPlayer struct {
	No                       string  // 序号
	UserId                   string  // 玩家id
	Nickname                 string  // 玩家昵称
	Ctime                    string  // 注册时间
	LoginTime                string  // 最后登录时间
	LiveDays                 int     // 存活天数
	LoseDays                 int     // 流失天数
	UserType                 string  // 玩家类型
	RegistArea               string  // 账号类型
	AllRounds                int32   // 总局数
	HZFrequency              string  // 换桌频率
	ChangeTableCount         int64   // 主动点换桌次数
	Bets                     int64   // 总打码量
	BetsAvg                  string  // 局均打码量
	WinRounds                int32   // 赢局数
	LoseRounds               int32   // 输局数
	WinRate                  string  // 玩家总胜率
	Rounds2                  int64   // 玩家2人局
	Rounds3                  int64   // 玩家3人局
	Rounds4                  int64   // 玩家4人局
	Rounds5                  int64   // 玩家5人局
	WinRounds2               int64   // 玩家2人局赢局
	WinRounds3               int64   // 玩家3人局赢局
	WinRounds4               int64   // 玩家4人局赢局
	WinRounds5               int64   // 玩家5人局赢局
	WinRounds2Rate           string  // 玩家2人局胜率
	WinRounds3Rate           string  // 玩家3人局胜率
	WinRounds4Rate           string  // 玩家4人局胜率
	WinRounds5Rate           string  // 玩家5人局胜率
	RewardRate               string  // 总返奖率
	TpHeartbeatHr            float64 // 玩家心跳
	TPNum                    int64   // tp人数
	PlayerHeartRate          string  // 玩家心率
	ChargeTimes              int16   // 充值触发次数
	ChargeLaunchTimes        int16   // 局内充值拉单次数
	ChargePayTimes           int16   // 局内充值实付次数
	WinBets                  int64   // 赢局打码量
	LoseBets                 int64   // 输局打码量
	WinBetsAvg               int64   // 赢局均码量
	LoseBetsAvg              int64   // 输局均码量
	Wins                     int64   // 赢局赢钱金额
	Loses                    int64   // 输局输钱金额
	Cash                     int64   // 净赢
	FWinsAvg                 string  // 胜局均盈利
	FLosesAvg                string  // 败局均输额
	FCashAvg                 string  // 局均净赢
	HuaTypeTimes             [13]int // 拿到牌型次数 1-12
	HuaTypeWinTimes          [13]int // 拿到牌型赢次数 1-12
	Hua1Rate                 string  // 玩家拿大豹子率
	Hua2Rate                 string  // 玩家拿小豹子率
	Hua3Rate                 string  // 玩家拿大顺金率
	Hua4Rate                 string  // 玩家拿小顺金率
	Hua5Rate                 string  // 玩家拿大顺子率
	Hua6Rate                 string  // 玩家拿小顺子率
	Hua7Rate                 string  // 玩家拿大同花率
	Hua8Rate                 string  // 玩家拿小同花率
	Hua9Rate                 string  // 玩家拿大对子率
	Hua10Rate                string  // 玩家拿小对子率
	Hua11Rate                string  // 玩家拿大高牌率
	Hua12Rate                string  // 玩家拿小高牌率
	Hua1WinRate              string  // 玩家拿大豹子胜率
	Hua2WinRate              string  // 玩家拿小豹子胜率
	Hua3WinRate              string  // 玩家拿大顺金胜率
	Hua4WinRate              string  // 玩家拿小顺金胜率
	Hua5WinRate              string  // 玩家拿大顺子胜率
	Hua6WinRate              string  // 玩家拿小顺子胜率
	Hua7WinRate              string  // 玩家拿大同花胜率
	Hua8WinRate              string  // 玩家拿小同花胜率
	Hua9WinRate              string  // 玩家拿大对子胜率
	Hua10WinRate             string  // 玩家拿小对子胜率
	Hua11WinRate             string  // 玩家拿大高牌胜率
	Hua12WinRate             string  // 玩家拿小高牌胜率
	OpponentWinRound         int64   // 冤人赢局数
	OpponentLoseRound        int64   // 冤人输局数
	OpponentWins             float64 // 冤人赢局赢钱金额
	OpponentLoses            float64 // 冤人输局输钱金额
	OpponentCash             float64 // 冤人局净赢
	HurtWinRound             int64   // 偷鸡赢局数
	HurtLoseRound            int64   // 偷鸡输局数
	HurtWins                 float64 // 偷鸡赢钱额
	HurtLoses                float64 // 被偷鸡输钱额
	HurtCash                 float64 // 偷与被偷净赢
	MaxWinHuaType            string  // 玩家总赢得最多的牌型
	MaxLoseHuaType           string  // 玩家总输得最多的牌型
	MaxWinHuaTypeRound       int64   // 玩家总赢得最多的牌型拿到的局数
	MaxLoseHuaTypeRound      int64   // 玩家总输得最多的牌型拿到的局数
	MaxWinHuaTypeAmount      int64   // 玩家总赢的最多的牌型总赢金额
	MaxLoseHuaTypeAmount     int64   // 玩家总输得最多的牌型总输金额
	MaxWinHuaTypeAmountAvg   int64   // 玩家总赢最多的牌型局均赢钱
	MaxLoseHuaTypeAmountAvg  int64   // 玩家总输最多牌型局均输钱
	OnceMaxWinHuaType        string  // 玩家单局赢的最多的牌型
	OnceMaxLoseHuaType       string  // 玩家单局输的最多的牌型
	OnceMaxWinHuaTypeAmount  int64   // 玩家单局赢最多牌型的赢钱金额
	OnceMaxLoseHuaTypeAmount int64   // 玩家单局输最多牌型的输钱金额

	FBets                     string // 总打码量
	FWinBets                  string // 赢局打码量
	FLoseBets                 string // 输局打码量
	FWinBetsAvg               string // 赢局均码量
	FLoseBetsAvg              string // 输局均码量
	FWins                     string // 赢局赢钱金额
	FLoses                    string // 输局输钱金额
	FCash                     string // 净赢
	FOpponentWins             string // 冤人赢局赢钱金额
	FOpponentLoses            string // 冤人输局输钱金额
	FOpponentCash             string // 冤人局净赢
	FHurtWins                 string // 偷鸡赢钱额
	FHurtLoses                string // 被偷鸡输钱额
	FHurtCash                 string // 偷与被偷净赢
	FMaxWinHuaTypeAmount      string // 玩家总赢的最多的牌型总赢金额
	FMaxLoseHuaTypeAmount     string // 玩家总输得最多的牌型总输金额
	FMaxWinHuaTypeAmountAvg   string // 玩家总赢最多的牌型局均赢钱
	FMaxLoseHuaTypeAmountAvg  string // 玩家总输最多牌型局均输钱
	FOnceMaxWinHuaTypeAmount  string // 玩家单局赢最多牌型的赢钱金额
	FOnceMaxLoseHuaTypeAmount string // 玩家单局输最多牌型的输钱金额
}

type TpNewStatDate struct {
	DateStr            string  // 日期
	LoginedUsers       int32   // 日活数量
	NewLoginedUsers    int32   // 新顾客日活数量
	OldLoginedUsers    int32   // 老顾客日活数量
	Players            int32   // 总玩tp人数
	NewPlayers         int32   // 新顾客玩tp人数
	OldPlayers         int32   // 老顾客玩tp人数
	PlayerRate         string  // 玩tp人数占日活比
	NewPlayerRate      string  // 新顾客玩tp占新顾客日活比
	OldPlayerRate      string  // 老顾客玩tp占老顾客日活比
	AllRounds          int32   // 总局数
	HZFrequency        string  // 换桌频率
	ChangeTableCount   int64   // 主动点换桌次数
	PlayerRoundsAvg    string  // 人均局数
	PlayerRoundsMedian string  // 局数中位数
	PlayerBets         int64   // 总打码
	PlayerBetsAvg      string  // 人均打码
	RoundsBetsAvg      string  // 局均打码量
	WinRounds          int32   // 赢局数
	LoseRounds         int32   // 输局数
	WinRate            string  // 玩家胜率
	Rounds2            int64   // 玩家2人局
	Rounds3            int64   // 玩家3人局
	Rounds4            int64   // 玩家4人局
	Rounds5            int64   // 玩家5人局
	WinRounds2         int64   // 玩家2人局赢局
	WinRounds3         int64   // 玩家3人局赢局
	WinRounds4         int64   // 玩家4人局赢局
	WinRounds5         int64   // 玩家5人局赢局
	WinRounds2Rate     string  // 玩家2人局胜率
	WinRounds3Rate     string  // 玩家3人局胜率
	WinRounds4Rate     string  // 玩家4人局胜率
	WinRounds5Rate     string  // 玩家5人局胜率
	RewardRate         string  // 总返奖率
	TpHeartbeatHr      float64 // 玩家心跳
	TPNum              int64   // tp人数
	PlayerHeartRate    string  // 玩家心率
	ChargeTimes        int16   // 充值触发次数
	ChargeLaunchTimes  int16   // 局内充值拉单次数
	ChargePayTimes     int16   // 局内充值实付次数
	WinBets            int64   // 赢局打码量
	LoseBets           int64   // 输局打码量
	WinBetsAvg         int64   // 赢局均码量
	LoseBetsAvg        int64   // 输局均码量
	Wins               int64   // 赢局赢钱金额
	Loses              int64   // 输局输钱金额
	Cash               int64   // 净赢
	FWinsAvg           string  // 胜局均盈利
	FLosesAvg          string  // 败局均输额
	FCashAvg           string  // 局均净赢
	HuaTypeTimes       [13]int // 拿到牌型次数 1-12
	HuaTypeWinTimes    [13]int // 拿到牌型赢次数 1-12
	Hua1Rate           string  // 玩家拿大豹子率
	Hua2Rate           string  // 玩家拿小豹子率
	Hua3Rate           string  // 玩家拿大顺金率
	Hua4Rate           string  // 玩家拿小顺金率
	Hua5Rate           string  // 玩家拿大顺子率
	Hua6Rate           string  // 玩家拿小顺子率
	Hua7Rate           string  // 玩家拿大同花率
	Hua8Rate           string  // 玩家拿小同花率
	Hua9Rate           string  // 玩家拿大对子率
	Hua10Rate          string  // 玩家拿小对子率
	Hua11Rate          string  // 玩家拿大高牌率
	Hua12Rate          string  // 玩家拿小高牌率
	Hua1WinRate        string  // 玩家拿大豹子胜率
	Hua2WinRate        string  // 玩家拿小豹子胜率
	Hua3WinRate        string  // 玩家拿大顺金胜率
	Hua4WinRate        string  // 玩家拿小顺金胜率
	Hua5WinRate        string  // 玩家拿大顺子胜率
	Hua6WinRate        string  // 玩家拿小顺子胜率
	Hua7WinRate        string  // 玩家拿大同花胜率
	Hua8WinRate        string  // 玩家拿小同花胜率
	Hua9WinRate        string  // 玩家拿大对子胜率
	Hua10WinRate       string  // 玩家拿小对子胜率
	Hua11WinRate       string  // 玩家拿大高牌胜率
	Hua12WinRate       string  // 玩家拿小高牌胜率
	OpponentWinRound   int64   // 冤人赢局数
	OpponentLoseRound  int64   // 冤人输局数
	OpponentWins       float64 // 冤人赢局赢钱金额
	OpponentLoses      float64 // 冤人输局输钱金额
	OpponentCash       float64 // 冤人局净赢
	HurtWinRound       int64   // 偷鸡赢局数
	HurtLoseRound      int64   // 偷鸡输局数
	HurtWins           float64 // 偷鸡赢钱额
	HurtLoses          float64 // 被偷鸡输钱额
	HurtCash           float64 // 偷与被偷净赢

	FBets          string // 总打码量
	FWinBets       string // 赢局打码量
	FLoseBets      string // 输局打码量
	FWinBetsAvg    string // 赢局均码量
	FLoseBetsAvg   string // 输局均码量
	FWins          string // 赢局赢钱金额
	FLoses         string // 输局输钱金额
	FCash          string // 净赢
	FOpponentWins  string // 冤人赢局赢钱金额
	FOpponentLoses string // 冤人输局输钱金额
	FOpponentCash  string // 冤人局净赢
	FHurtWins      string // 偷鸡赢钱额
	FHurtLoses     string // 被偷鸡输钱额
	FHurtCash      string // 偷与被偷净赢
}

// 乐极生悲
type TpNewLJSBStat struct {
	No                  string  // 序号
	UserId              string  // 玩家id
	Nickname            string  // 玩家昵称
	Ctime               string  // 注册时间
	LoginTime           string  // 最后登录时间
	LiveDays            int     // 存活天数
	LoseDays            int     // 流失天数
	UserType            string  // 玩家类型
	RegistArea          string  // 账号类型
	AllRounds           int32   // 总局数
	Bets                int64   // 总打码量
	BetsAvg             string  // 局均打码量
	StrategyRounds      int64   // 策略局数
	StrategyBets        int64   // 策略局总打码
	StrategyBetsAvg     string  // 策略局局均打码量
	StrategyWinRate     string  // 策略中胜率
	StrategyRewardRate  string  // 策略中返奖率
	StrategyWinRounds   int64   // 策略赢局数
	StrategyWinBets     int64   // 策略局赢局总打码
	StrategyLoseBets    int64   // 策略局输局总打码
	WinRate             string  // 玩家总胜率
	RewardRate          string  // 总返奖率
	WinRounds           int32   // 赢局数
	LoseRounds          int32   // 输局数
	TpHeartbeatHr       float64 // 玩家心跳
	TPNum               int64   // tp人数
	PlayerHeartRate     string  // 玩家心率
	JlStatus            string  // 极乐状态
	ByRounds            int64   // 被冤局数
	YsRounds            int64   // 冤杀局数
	YdRounds            int64   // 冤大局数
	YtRounds            int64   // 冤逃局数
	OpponentRound       int64   // 冤人总局输
	ByRate              string  // 被冤率
	YsRate              string  // 冤杀率
	YdRate              string  // 冤大率
	YtRate              string  // 冤逃率
	OpponentWinRound    int64   // 冤人赢局数
	OpponentLoseRound   int64   // 冤人输局数
	OpponentWins        float64 // 冤人赢局赢钱金额
	OpponentLoses       float64 // 冤人输局输钱金额
	OpponentCash        float64 // 冤人局净赢
	OpponentWinRate     string  // 策略中冤家局胜率
	WinBets             int64   // 赢局打码量
	LoseBets            int64   // 输局打码量
	WinBetsAvg          int64   // 赢局均码量
	LoseBetsAvg         int64   // 输局均码量
	Wins                int64   // 赢局赢钱金额
	Loses               int64   // 输局输钱金额
	Cash                int64   // 净赢
	HuaTypeTimes        [13]int // 拿到牌型次数 1-12
	HuaTypeWinTimes     [13]int // 拿到牌型赢次数 1-12
	Hua1Rate            string  // 玩家拿大豹子率
	Hua2Rate            string  // 玩家拿小豹子率
	Hua3Rate            string  // 玩家拿大顺金率
	Hua4Rate            string  // 玩家拿小顺金率
	Hua5Rate            string  // 玩家拿大顺子率
	Hua6Rate            string  // 玩家拿小顺子率
	Hua7Rate            string  // 玩家拿大同花率
	Hua8Rate            string  // 玩家拿小同花率
	Hua9Rate            string  // 玩家拿大对子率
	Hua10Rate           string  // 玩家拿小对子率
	Hua11Rate           string  // 玩家拿大高牌率
	Hua12Rate           string  // 玩家拿小高牌率
	Hua1WinRate         string  // 玩家拿大豹子胜率
	Hua2WinRate         string  // 玩家拿小豹子胜率
	Hua3WinRate         string  // 玩家拿大顺金胜率
	Hua4WinRate         string  // 玩家拿小顺金胜率
	Hua5WinRate         string  // 玩家拿大顺子胜率
	Hua6WinRate         string  // 玩家拿小顺子胜率
	Hua7WinRate         string  // 玩家拿大同花胜率
	Hua8WinRate         string  // 玩家拿小同花胜率
	Hua9WinRate         string  // 玩家拿大对子胜率
	Hua10WinRate        string  // 玩家拿小对子胜率
	Hua11WinRate        string  // 玩家拿大高牌胜率
	Hua12WinRate        string  // 玩家拿小高牌胜率
	FBets               string  // 总打码量
	FWinBets            string  // 赢局打码量
	FLoseBets           string  // 输局打码量
	FWinBetsAvg         string  // 赢局均码量
	FLoseBetsAvg        string  // 输局均码量
	FWins               string  // 赢局赢钱金额
	FLoses              string  // 输局输钱金额
	FCash               string  // 净赢
	FOpponentWins       string  // 冤人赢局赢钱金额
	FOpponentLoses      string  // 冤人输局输钱金额
	FOpponentCash       string  // 冤人局净赢
	FStrategyBets       string  // 策略局总打码
	FStrategyWinBets    string  // 策略局赢局总打码
	FStrategyLoseBets   string  // 策略局输局总打码
	FStrategyWinBetAvg  string  // 策略中赢局均码量
	FStrategyLoseBetAvg string  // 策略中输局均码量
	FStrategyWins       string  // 策略中赢局赢钱金额
	FStrategyLoses      string  // 策略中输局输钱金额
	FStrategyCash       string  // 策略中净赢
}

// 高潮涌现
type TpNewGCYXStat struct {
	No                    string  // 序号
	UserId                string  // 玩家id
	Nickname              string  // 玩家昵称
	Ctime                 string  // 注册时间
	LoginTime             string  // 最后登录时间
	UserType              string  // 玩家类型
	RegistArea            string  // 账号类型
	LiveDays              int     // 存活天数
	LoseDays              int     // 流失天数
	AllRounds             int64   // 总局数
	Bets                  int64   // 总打码量
	FBets                 string  // 总打码量
	BetsAvg               string  // 局均打码量
	StrategyRounds        int64   // 高潮局数
	StrategyBets          int64   // 高潮局总打码
	FStrategyBets         string  // 高潮局总打码
	StrategyBetsAvg       string  // 高潮局局均打码量
	StrategyRate          string  // 高潮频率
	StrategyWinRounds     int64   // - 高潮局赢局数
	StrategyWinRate       string  // 高潮局中胜率
	StrategyRewardRate    string  // 高潮局返奖率
	Wins                  int64   // - 胜局总赢
	Loses                 int64   // - 败局总输
	TPNum                 int64   // summary tp人数
	TpRebateRate          string  // tp总返奖率
	WinRounds             int64   // - 赢局数
	TpWinRate             string  // tp总胜率
	TpHeartbeatHr         float64 // 玩家心跳
	PlayerHeartRate       string  // 玩家心率
	HighRpRounds          int64   // - 高潮局压制次数
	HighRpRate            string  // 高潮局压制次数占比
	HighBpRounds          int64   // - 高潮局恩赐次数
	HighBpRate            string  // 高潮局恩赐次数占比
	StrategyWinBetAvg     string  // 高潮局中赢局均码量
	StrategyLoseBetAvg    string  // 高潮局中输局均码量
	StrategyWins          int64   // - 高潮局中赢局赢钱金额
	StrategyLoses         int64   // - 高潮局中输局输钱金额
	FStrategyWins         string  // 高潮局中赢局赢钱金额
	FStrategyLoses        string  // 高潮局中输局输钱金额
	StrategyCash          string  // 高潮局中净赢
	HighChargeTimes       int16   // 高潮局中局内充触发次数
	HighChargeLaunchTimes int16   // 高潮局中局内充拉单数
	HighChargePayTimes    int16   // 高潮局中局内实际付款次数

	HuaTypeTimes    [13]int // 拿到牌型次数 1-12
	HuaTypeWinTimes [13]int // 拿到牌型赢次数 1-12
	Hua1Rate        string  // 高潮局中玩家拿大豹子率
	Hua2Rate        string  // 高潮局中玩家拿小豹子率
	Hua3Rate        string  // 高潮局中玩家拿大顺金率
	Hua4Rate        string  // 高潮局中玩家拿小顺金率
	Hua5Rate        string  // 高潮局中玩家拿大顺子率
	Hua6Rate        string  // 高潮局中玩家拿小顺子率
	Hua7Rate        string  // 高潮局中玩家拿大同花率
	Hua8Rate        string  // 高潮局中玩家拿小同花率
	Hua9Rate        string  // 高潮局中玩家拿大对子率
	Hua10Rate       string  // 高潮局中玩家拿小对子率
	Hua11Rate       string  // 高潮局中玩家拿大高牌率
	Hua12Rate       string  // 高潮局中玩家拿小高牌率

	Hua1WinRate  string // 高潮局中玩家大豹子胜率
	Hua2WinRate  string // 高潮局中玩家小豹子胜率
	Hua3WinRate  string // 高潮局中玩家大顺金胜率
	Hua4WinRate  string // 高潮局中玩家小顺金胜率
	Hua5WinRate  string // 高潮局中玩家大顺子胜率
	Hua6WinRate  string // 高潮局中玩家小顺子胜率
	Hua7WinRate  string // 高潮局中玩家大同花胜率
	Hua8WinRate  string // 高潮局中玩家小同花胜率
	Hua9WinRate  string // 高潮局中玩家大对子胜率
	Hua10WinRate string // 高潮局中玩家小对子胜率
	Hua11WinRate string // 高潮局中玩家大高牌胜率
	Hua12WinRate string // 高潮局中玩家小高牌胜率
}

// mines 汇总统计
type MinesStatDate struct {
	DateStr            string // 日期
	LoginedUsers       int32  // 日活数量
	NewLoginedUsers    int32  // 新顾客日活数量
	OldLoginedUsers    int32  // 老顾客日活数量
	Players            int32  // 总玩mines人数
	NewPlayers         int32  // 新顾客玩mines人数
	OldPlayers         int32  // 老顾客玩mines人数
	PlayerRate         string // 玩mines人数占日活比
	NewPlayerRate      string // 新顾客玩mines占新顾客日活比
	OldPlayerRate      string // 老顾客玩mines占老顾客日活比
	AllRounds          int32  // 总局数
	PlayerRoundsAvg    string // 人均局数
	PlayerRoundsMedian string // 局数中位数
	GameTimes          string // 游戏总时长（min）
	GameTimesAvg       string // 游戏局均时长（min)
	GameTimesPlayerAvg string // 游戏人均时长（min)
	PlayerBets         int64  // 总打码
	PlayerBetsAvg      string // 人均打码
	RoundsBetsAvg      string // 局均打码量
	WinRounds          int32  // 胜利局数
	LoseRounds         int32  // 失败局数
	WinRate            string // 胜率
	RewardRate         string // 返奖率
	WinBets            int64  // 胜局打码量
	LoseBets           int64  // 败局打码量
	WinBetsAvg         int64  // 胜局均码量
	LoseBetsAvg        int64  // 败局均码量
	Wins               int64  // 胜局赢钱金额
	Loses              int64  // 败局输钱金额
	Cash               int64  // 净赢
	FWinsAvg           string // 胜局均盈利
	FLosesAvg          string // 败局均输额
	FCashAvg           string // 局均净赢

	WinMulpitleMax     string // 赢局最高返奖倍数
	WinMulpitleAvg     string // 赢局局均返奖倍数
	WinMulpitleMedian  string // 赢局返奖倍数中位数
	LoseMulpitleMax    string // 输局最高止步倍数
	LoseMulpitleAvg    string // 输局局均止步倍数
	LoseMulpitleMedian string // 输局止步倍数中位数

	ChargeTimes   int64  // 局内充成单数
	ChargeAmounts string // 局内充成金额
	ChargePayAvg  string // 局内充成单均价
	ChargeRate    string // 局内充单数占总充单数比

	FBets        string // 总打码量
	FWinBets     string // 胜局打码量
	FLoseBets    string // 败局打码量
	FWinBetsAvg  string // 胜局均码量
	FLoseBetsAvg string // 败局均码量
	FWins        string // 胜局赢钱金额
	FLoses       string // 败局输钱金额
	FCash        string // 净赢

	GameSeconds int64 // 游戏总时长（seconds）
}

type MinesStatPlayer struct {
	No          string // 序号
	UserId      string // 玩家id
	Nickname    string // 玩家昵称
	VipLv       int32  // vip等级
	UserType    string // 玩家类型
	RegistArea  string // 账号类型
	Ctime       string // 注册时间
	LoginTime   string // 最后登录时间
	LiveDays    int    // 存活天数
	LoseDays    int    // 流失天数
	BetsAll     int64  // 所有游戏总打码量
	Bets        int64  // mines打码量
	BetsRate    string // mines打码量占比
	AllRounds   int32  // 下注局数
	BetsAvg     string // 局均打码量
	WinRounds   int32  // 胜利局数
	LoseRounds  int32  // 失败局数
	WinRate     string // 胜率
	RewardRate  string // 返奖率
	WinBets     int64  // 胜局打码量
	LoseBets    int64  // 败局打码量
	WinBetsAvg  int64  // 胜局均码量
	LoseBetsAvg int64  // 败局均码量
	Wins        int64  // 胜局赢钱金额
	Loses       int64  // 败局输钱金额
	Cash        int64  // 净赢
	FWinsAvg    string // 胜局均盈利
	FLosesAvg   string // 败局均输额
	FCashAvg    string // 局均净赢

	WinMulpitleMax     string // 赢局最高返奖倍数
	WinMulpitleAvg     string // 赢局局均返奖倍数
	WinMulpitleMedian  string // 赢局返奖倍数中位数
	LoseMulpitleMax    string // 输局最高止步倍数
	LoseMulpitleAvg    string // 输局局均止步倍数
	LoseMulpitleMedian string // 输局止步倍数中位数

	ChargeTimes   string // 局内充成单数
	ChargeAmounts string // 局内充成金额
	ChargePayAvg  string // 局内充成单均价
	ChargeRate    string // 局内充单数占总充单数比

	FMoney   string // 总充值
	FCashOut string // 总提现
	FProfit  string // 总赢

	FBetsAll     string // 所有游戏总打码量
	FBets        string // 总打码量
	FWinBets     string // 赢局打码量
	FLoseBets    string // 输局打码量
	FWinBetsAvg  string // 赢局均码量
	FLoseBetsAvg string // 输局均码量
	FWins        string // 赢局赢钱金额
	FLoses       string // 输局输钱金额
	FCash        string // 净赢
}
