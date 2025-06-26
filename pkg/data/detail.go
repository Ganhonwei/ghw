package data

import "goserver/gen/pb"

type Detail struct {
	WaterId                            string              `bson:"_id" json:"_id"`                                                                             //局号
	BeginTime                          int64               `bson:"begin_time" json:"begin_time"`                                                               //开始时间
	EndTime                            int64               `bson:"end_time" json:"end_time"`                                                                   //结束时间
	Gtype                              int32               `bson:"gtype" json:"gtype"`                                                                         //所在游戏
	Rtype                              int32               `bson:"rtype" json:"rtype"`                                                                         //房间类型: 0自由,1私人,2百人
	Gmode                              int32               `bson:"gmode" json:"gmode"`                                                                         //私人房模式: 0真金,1娱乐模式
	RoomId                             string              `bson:"room_id" json:"room_id"`                                                                     //房间ID
	DeskId                             string              `bson:"desk_id" json:"desk_id"`                                                                     //桌子ID
	Players                            string              `bson:"players" json:"players"`                                                                     //参与玩家ID
	CardTypeId                         int                 `bson:"card_type_id" json:"card_type_id"`                                                           //牌型ID
	PlayerStageId                      int                 `bson:"player_stage_id" json:"player_stage_id"`                                                     //玩家阶段ID
	ChangeCardType                     int                 `bson:"change_card_type" json:"change_card_type"`                                                   //换牌局类型(开局换，局中换)
	ControlType                        int                 `bson:"control_type" json:"control_type"`                                                           //控制方式(开局)
	WinScore                           int                 `bson:"win_score" json:"win_score"`                                                                 //当前赢分(开局)
	PlayerFactor                       int                 `bson:"player_factor" json:"player_factor"`                                                         //玩家系数(开局)
	WinScore1                          int                 `bson:"win_score_1" json:"win_score_1"`                                                             //当局赢分(局中)
	WinScore2                          int                 `bson:"win_score_2" json:"win_score_2"`                                                             //当前赢分(局中)
	ChargeMoney                        int                 `bson:"charge_money" json:"charge_money"`                                                           //充值金额(开局，局中)
	IsStrategy                         bool                `bson:"is_strategy" json:"is_strategy"`                                                             //是否是策略局
	StrategyType                       int                 `bson:"strategy_type" json:"strategy_type"`                                                         //策略局类型
	IsCharge                           bool                `bson:"is_charge" json:"is_charge"`                                                                 //是否充值
	ChargeInGame                       [][4]int32          `bson:"charge_in_game" json:"charge_in_game"`                                                       // 局内充值触发 [seatid,金额,拉单(0/1),实付(0/1)]
	RealGame                           bool                `bson:"real_game" json:"real_game"`                                                                 //是否真人对局
	NonDirty                           []uint32            `bson:"non_dirty" json:"non_dirty"`                                                                 //没有污染的人机座位
	ShowCards                          string              `bson:"show_cards" json:"show_cards"`                                                               //亮牌
	IsMustLose                         bool                `bson:"is_must_lose" json:"is_must_lose"`                                                           //是否是必输局
	MustLoseScore                      int                 `bson:"must_lose_score" json:"must_lose_score"`                                                     //必输分数线
	CoreRobotSeat                      uint32              `bson:"core_robot_seat" json:"core_robot_seat"`                                                     //核心人机位置
	IsUpCardPool                       bool                `bson:"is_up_card_pool" json:"is_up_card_pool"`                                                     //是否升档牌库
	CardPoolId                         int                 `bson:"card_pool_id" json:"card_pool_id"`                                                           //牌库编号
	DealType                           int                 `bson:"deal_type" json:"deal_type"`                                                                 //发牌方式
	IsChangeTable                      bool                `bson:"is_change_table" json:"is_change_table"`                                                     // 是否是换桌
	IsTriggerWinScoreLimit             bool                `bson:"is_trigger_win_score_limit" json:"is_trigger_win_score_limit"`                               //是否触发赢分限制
	IsTriggerMustLose                  bool                `bson:"is_trigger_must_lose" json:"is_trigger_must_lose"`                                           //是否触发必输控制
	IsTriggerMustLoseDrawCard          bool                `bson:"is_trigger_must_lose_draw_card" json:"is_trigger_must_lose_draw_card"`                       //是否触发必输控制摸牌限制
	IsTriggerMustLoseLastCard          bool                `bson:"is_trigger_must_lose_last_card" json:"is_trigger_must_lose_last_card"`                       //是否触发必输控制上家出牌限制
	IsTriggerMustLoseCoreRobotGoodCard bool                `bson:"is_trigger_must_lose_core_robot_good_card" json:"is_trigger_must_lose_core_robot_good_card"` //是否触发必输控制核心人机摸好牌
	TPModel                            int32               `bson:"tp_model" json:"tp_model"`                                                                   // tp模式 (0新手，1免费，2正常，3剧情，4控制策略)
	IsTPModel3StoryPlus                bool                `bson:"is_tp_model3story_plus" json:"is_tp_model3story_plus"`                                       // 是否tp剧情局plus
	RMControl                          *RMControl          `bson:"rm_control" json:"rm_control"`                                                               //rm 控制详情
	RMGameOverReason                   int                 `bson:"rm_game_over_reason" json:"rm_game_over_reason"`                                             //rm游戏结束原因 (1.天胡,2.自摸胡牌,3.吃牌胡牌,4.对手弃牌)
	TpStrategy                         *TpStrategy         `bson:"tp_strategy" json:"tp_strategy"`                                                             // tp 对局策略
	TPDetail                           []*TPDetail         //TP详情
	JOKERDetail                        []*JOKERDetail      //JOKER详情
	AK47Detail                         []*AK47Detail       //AK47详情
	RMDetail                           []*RMDetail         //RM详情
	LHDetail                           *LHDetail           //LH详情
	UPDetail                           *UPDetail           //7up详情
	CRASHDetail                        *CRASHDetail        //crash
	ABDetail                           *ABDetail           //ab详情
	*CPDetail                                              //彩票
	RBDetail                           *RBDetail           //红黑
	MinesDetail                        *MinesDetail        //地雷
	FortuneGems2Detail                 *FortuneGems2Detail //宝石2详情
	FortuneGemsDetail                  *FortuneGemsDetail  //宝石详情
}

func (d *Detail) FindTPDetail(id uint32) *TPDetail {
	for _, v := range d.TPDetail {
		if v.SeatId == id {
			return v
		}
	}
	return nil
}

func (d *Detail) FindJOKERDetail(id uint32) *JOKERDetail {
	for _, v := range d.JOKERDetail {
		if v.SeatId == id {
			return v
		}
	}
	return nil
}

func (d *Detail) FindAK47Detail(id uint32) *AK47Detail {
	for _, v := range d.AK47Detail {
		if v.SeatId == id {
			return v
		}
	}
	return nil
}

func (d *Detail) FindRMDetail(id uint32) *RMDetail {
	for _, v := range d.RMDetail {
		if v.SeatId == id {
			return v
		}
	}
	return nil
}

type TPDetail struct {
	SeatId       uint32          `json:"seat_id" bson:"seat_id"`               //座位ID
	UserId       string          `json:"user_id" bson:"user_id"`               //用户ID
	Cards        []uint32        `json:"cards" bson:"cards"`                   //牌型
	Score        int64           `json:"score" bson:"score"`                   //结算
	BeforeScore  int64           `json:"before_score" bson:"before_score"`     //账变前分数
	AfterScore   int64           `json:"after_score" bson:"after_score"`       //账变后分数
	BeforeCash   int64           `json:"before_cash" bson:"before_cash"`       //账变前彩金
	AfterCash    int64           `json:"after_cash" bson:"after_cash"`         //账变后彩金
	BeforeBonus  int64           `json:"before_bonus" bson:"before_bonus"`     //账变前奖励金
	AfterBonus   int64           `json:"after_bonus" bson:"after_bonus"`       //账变后奖励金
	Bet          int64           `json:"bet" bson:"bet"`                       //总投注
	Bottom       int64           `json:"bottom" bson:"bottom"`                 //底注
	CashStock    int64           `json:"cash_stock" bson:"cash_stock"`         //彩金库存
	BonusStock   int64           `json:"bonus_stock" bson:"bonus_stock"`       //奖励金库存
	CashMingTax  int64           `json:"cash_ming_tax" bson:"cash_ming_tax"`   //彩金明税
	BonusMingTax int64           `json:"bonus_ming_tax" bson:"bonus_ming_tax"` //奖励金明税
	CashAnTax    int64           `json:"cash_an_tax" bson:"cash_an_tax"`       //彩金暗税
	BonusAnTax   int64           `json:"bonus_an_tax" bson:"bonus_an_tax"`     //奖励金暗税
	TPDetailTurn []*TPDetailTurn `json:"tp_detail_turn" bson:"tp_detail_turn"` //轮次详情
	// LHDetail     LHDetail        `json:"lhUserDetail" bson:"lh_user_detail"`   //
	NewBieProbeId    int32 `json:"new_bie_probe_id" bson:"new_bie_probe_id"`       // 新手试探期/稳定局策略id
	NewBieProbeRound int32 `json:"new_bie_probe_round" bson:"new_bie_probe_round"` // 玩家新手试探期/稳定局策略生效局数

}

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
	JOKERDetailTurn []*JOKERDetailTurn `json:"tp_detail_turn" bson:"tp_detail_turn"` //轮次详情
	// LHDetail     LHDetail        `json:"lhUserDetail" bson:"lh_user_detail"`   //
}

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
	AK47DetailTurn []*AK47DetailTurn `json:"tp_detail_turn" bson:"tp_detail_turn"` //轮次详情
	// LHDetail     LHDetail        `json:"lhUserDetail" bson:"lh_user_detail"`   //
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

type RMDetail struct {
	SeatId      uint32     `json:"seat_id" bson:"seat_id"`           //座位ID
	UserId      string     `json:"user_id" bson:"user_id"`           //用户ID
	Result      string     `json:"result" bson:"result"`             //结果
	Score       int64      `json:"score" bson:"score"`               //结算
	BeforeScore int64      `json:"before_score" bson:"before_score"` //账变前分数
	AfterScore  int64      `json:"after_score" bson:"after_score"`   //账变后分数
	BeforeCash  int64      `json:"before_cash" bson:"before_cash"`   //账变前彩金
	AfterCash   int64      `json:"after_cash" bson:"after_cash"`     //账变后彩金
	BeforeBonus int64      `json:"before_bonus" bson:"before_bonus"` //账变前奖励金
	AfterBonus  int64      `json:"after_bonus" bson:"after_bonus"`   //账变后奖励金
	FinalCards  [][]uint32 `json:"final_cards" bson:"final_cards"`   //最终牌型
	InitCards   []uint32   `json:"init_cards" bson:"init_cards"`     //初始牌型
	WildCard    uint32     `json:"wild_card" bson:"wild_card"`       //万能牌
	MoCards     []uint32   `json:"mo_cards" bson:"mo_cards"`         //每轮摸到的牌
	ChuCards    []uint32   `json:"chu_cards" bson:"chu_cards"`       //每轮出的牌

	CashStock    int64 `json:"cash_stock" bson:"cash_stock"`         //彩金库存
	BonusStock   int64 `json:"bonus_stock" bson:"bonus_stock"`       //奖励金库存
	CashMingTax  int64 `json:"cash_ming_tax" bson:"cash_ming_tax"`   //彩金明税
	BonusMingTax int64 `json:"bonus_ming_tax" bson:"bonus_ming_tax"` //奖励金明税
	CashAnTax    int64 `json:"cash_an_tax" bson:"cash_an_tax"`       //彩金暗税
	BonusAnTax   int64 `json:"bonus_an_tax" bson:"bonus_an_tax"`     //奖励金暗税
}

type LHDetail struct {
	Winner      uint32         `json:"winner" bson:"winner"`            //赢家
	DragonValue string         `json:"dragonValue" bson:"dragon_value"` //龙牌值
	TigerValue  string         `json:"tigerValue" bson:"tiger_value"`   //虎牌值
	Bets        int64          `json:"bets" bson:"bets"`                //总下注
	PlayerWin   int64          `json:"playerWin" bson:"player_win"`     //玩家赢分
	StrategyId  int            `json:"strategyId" bson:"strategy_id"`   //策略id
	IsSuppress  bool           `json:"isSuppress" bson:"is_suppress"`   //是否压制
	UserDetail  []LHUserDetail `json:"userDetail" bson:"user_detail"`   //用户详细
}

type UPDetail struct {
	Winner     uint32           `json:"winner" bson:"winner"`          //赢家
	PointValue int32            `json:"pointValue" bson:"point_value"` //点数
	Bets       int64            `json:"bets" bson:"bets"`              //总下注
	PlayerWin  int64            `json:"playerWin" bson:"player_win"`   //玩家赢分
	StrategyId int              `json:"strategyId" bson:"strategy_id"` //策略id
	IsSuppress bool             `json:"isSuppress" bson:"is_suppress"` //是否压制
	Crit7up    bool             `json:"crit7up" bson:"crit_7up"`       //是否暴击7up对局,区分老数据
	Crited     int16            `json:"crited" bson:"crited"`          //触发暴击位置数
	CritOdds   map[uint32]int32 `json:"critOdds" bson:"crit_odds"`     //触发暴击位置和倍数
	UserDetail []LHUserDetail   `json:"userDetail" bson:"user_detail"` //用户详细
}

type CRASHDetail struct {
	Result       string            `json:"result" bson:"result"`              //开奖结果
	Bets         int64             `json:"bets" bson:"bets"`                  //总下注
	PlayerLose   int64             `json:"playerLose" bson:"player_lose"`     //玩家赢分
	StrategyType []int             `json:"strategyType" bson:"strategy_type"` //策略类型
	RkyhRfge     bool              `json:"RkyhRfge" bson:"rkyh_rfge"`         //人狂有祸r肥割
	UserDetail   []CRASHUserDetail `json:"userDetail" bson:"user_detail"`     //用户详细
	// TakeOffTime  int64             `json:"takeOffTime" bson:"take_off_time"`  //飞行时间微秒
}

type ABDetail struct {
	Joker        uint32          `json:"joker" bson:"joker"`                //Key牌
	JackpotValue uint32          `json:"jackpotValue" bson:"jackpot_value"` //中奖牌
	ACards       []uint32        `json:"aCards" bson:"a_cards"`             //ANDAR
	BCards       []uint32        `json:"bCards" bson:"b_cards"`             //BAHAR
	Winner       uint32          `json:"winner" bson:"winner"`              //赢家 1:ANDAR 2:BAHAR
	SideWinner   uint32          `json:"sideWinner" bson:"side_winner"`     //赢家2 位置3-10
	Bets         int64           `json:"bets" bson:"bets"`                  //总下注
	PlayerWin    int64           `json:"playerWin" bson:"player_win"`       //玩家赢分
	StrategyId   int             `json:"strategyId" bson:"strategy_id"`     //策略id
	IsRandom     bool            `json:"isRandom" bson:"is_random"`         //是否随机开的
	FinalFactor  float64         `json:"finalFactor" bson:"final_factor"`   //最终系数
	CanWinScore  int64           `json:"canWinScore" bson:"can_win_score"`  //可赢分
	UserDetail   []*ABUserDetail `json:"userDetail" bson:"user_detail"`     //用户详细
}

type CPDetail struct {
	JackpotOutput int64          `json:"jackpotOutput" bson:"jackpot_output"` //奖池产出
	CardType      uint32         `json:"cardType" bson:"card_type"`           //牌型
	Cards         []uint32       `json:"aCards" bson:"a_cards"`               //牌值
	Winner        uint32         `json:"winner" bson:"winner"`                //赢家
	Bets          int64          `json:"bets" bson:"bets"`                    //总下注
	PlayerWin     int64          `json:"playerWin" bson:"player_win"`         //玩家赢分
	StrategyId    int            `json:"strategyId" bson:"strategy_id"`       //策略id
	IsRandom      bool           `json:"isRandom" bson:"is_random"`           //是否随机开的
	FinalFactor   float64        `json:"finalFactor" bson:"final_factor"`     //最终系数
	CanWinScore   int64          `json:"canWinScore" bson:"can_win_score"`    //可赢分
	UserDetail    []CPUserDetail `json:"userDetail" bson:"user_detail"`       //用户详细
}

type RBDetail struct {
	CardType    []uint32       `json:"cardType" bson:"card_type"`        //牌型
	Cards       [][]uint32     `json:"aCards" bson:"a_cards"`            //牌值
	Winner      []uint32       `json:"winner" bson:"winner"`             //赢家
	Bets        int64          `json:"bets" bson:"bets"`                 //总下注
	PlayerWin   int64          `json:"playerWin" bson:"player_win"`      //玩家赢分
	StrategyId  int            `json:"strategyId" bson:"strategy_id"`    //策略id
	CanWinScore int64          `json:"canWinScore" bson:"can_win_score"` //可赢分
	UserDetail  []RBUserDetail `json:"userDetail" bson:"user_detail"`    //用户详细
}

// 玩家下注详情
type LHUserDetail struct {
	Userid         string           `json:"userid" bson:"userid"`                     //用户id
	Dragon         int64            `json:"dragon" bson:"dragon"`                     //龙下注
	Tiger          int64            `json:"tiger" bson:"tiger"`                       //虎下注
	Tie            int64            `json:"tie" bson:"tie"`                           //和下注
	SeatBets       map[uint32]int64 `json:"seatBets" bson:"seat_bets"`                //_位置下注: 0小,1大,2,3,4,...,12
	Observe        bool             `json:"observe" bson:"observe"`                   //观察局
	Result         string           `json:"result" bson:"result"`                     //输赢平
	Win            int64            `json:"win" bson:"win"`                           //输赢
	Crit           int32            `json:"crit" bson:"crit"`                         //_暴击位押中次数
	BeforeScore    int64            `json:"beforeScore" bson:"before_score"`          //账变前分数
	AfterScore     int64            `json:"after_score" bson:"after_score"`           //账变后分数
	BeforeCash     int64            `json:"before_cash" bson:"before_cash"`           //账变前彩金
	AfterCash      int64            `json:"after_cash" bson:"after_cash"`             //账变后彩金
	BeforeBonus    int64            `json:"before_bonus" bson:"before_bonus"`         //账变前奖励金
	AfterBonus     int64            `json:"after_bonus" bson:"after_bonus"`           //账变后奖励金
	CashMingTax    int64            `json:"cash_ming_tax" bson:"cash_ming_tax"`       //彩金明税
	BonusMingTax   int64            `json:"bonus_ming_tax" bson:"bonus_ming_tax"`     //奖励金明税
	CashAnTax      int64            `json:"cash_an_tax" bson:"cash_an_tax"`           //彩金暗税
	BonusAnTax     int64            `json:"bonus_an_tax" bson:"bonus_an_tax"`         //奖励金暗税
	BeforeBackRate int              `json:"before_back_rate" bson:"before_back_rate"` //账变前返奖率
	AfterBackRate  int              `json:"after_back_rate" bson:"after_back_rate"`   //账变后返奖率
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
}

// 玩家下注详情
type ABUserDetail struct {
	Userid       string           `json:"userid" bson:"userid"`                 //用户id
	SeatBets     map[string]int64 `json:"seatBets" bson:"seat_bets"`            //位置下注
	Observe      bool             `json:"observe" bson:"observe"`               //是否观察局
	Result       string           `json:"result" bson:"result"`                 //输赢平
	Win          int64            `json:"win" bson:"win"`                       //输赢
	BeforeScore  int64            `json:"beforeScore" bson:"before_score"`      //账变前分数
	AfterScore   int64            `json:"after_score" bson:"after_score"`       //账变后分数
	BeforeCash   int64            `json:"before_cash" bson:"before_cash"`       //账变前彩金
	AfterCash    int64            `json:"after_cash" bson:"after_cash"`         //账变后彩金
	BeforeBonus  int64            `json:"before_bonus" bson:"before_bonus"`     //账变前奖励金
	AfterBonus   int64            `json:"after_bonus" bson:"after_bonus"`       //账变后奖励金
	CashMingTax  int64            `json:"cash_ming_tax" bson:"cash_ming_tax"`   //彩金明税
	BonusMingTax int64            `json:"bonus_ming_tax" bson:"bonus_ming_tax"` //奖励金明税
	CashAnTax    int64            `json:"cash_an_tax" bson:"cash_an_tax"`       //彩金暗税
	BonusAnTax   int64            `json:"bonus_an_tax" bson:"bonus_an_tax"`     //奖励金暗税
}

// 彩票玩家下注详情
type CPUserDetail struct {
	Userid       string           `json:"userid" bson:"userid"`                 //用户id
	SeatBets     map[string]int64 `json:"seatBets" bson:"seat_bets"`            //位置下注
	Result       string           `json:"result" bson:"result"`                 //输赢平
	Win          int64            `json:"win" bson:"win"`                       //输赢
	Observe      bool             `json:"observe" bson:"observe"`               //观察局
	BeforeScore  int64            `json:"beforeScore" bson:"before_score"`      //账变前分数
	AfterScore   int64            `json:"after_score" bson:"after_score"`       //账变后分数
	BeforeCash   int64            `json:"before_cash" bson:"before_cash"`       //账变前彩金
	AfterCash    int64            `json:"after_cash" bson:"after_cash"`         //账变后彩金
	BeforeBonus  int64            `json:"before_bonus" bson:"before_bonus"`     //账变前奖励金
	AfterBonus   int64            `json:"after_bonus" bson:"after_bonus"`       //账变后奖励金
	CashMingTax  int64            `json:"cash_ming_tax" bson:"cash_ming_tax"`   //彩金明税
	BonusMingTax int64            `json:"bonus_ming_tax" bson:"bonus_ming_tax"` //奖励金明税
	CashAnTax    int64            `json:"cash_an_tax" bson:"cash_an_tax"`       //彩金暗税
	BonusAnTax   int64            `json:"bonus_an_tax" bson:"bonus_an_tax"`     //奖励金暗税
}

// 红黑
type RBUserDetail struct {
	Userid       string           `json:"userid" bson:"userid"`                 //用户id
	SeatBets     map[string]int64 `json:"seatBets" bson:"seat_bets"`            //位置下注
	Result       string           `json:"result" bson:"result"`                 //输赢平
	Win          int64            `json:"win" bson:"win"`                       //输赢
	Observe      bool             `json:"observe" bson:"observe"`               //观察局
	BeforeScore  int64            `json:"beforeScore" bson:"before_score"`      //账变前分数
	AfterScore   int64            `json:"after_score" bson:"after_score"`       //账变后分数
	BeforeCash   int64            `json:"before_cash" bson:"before_cash"`       //账变前彩金
	AfterCash    int64            `json:"after_cash" bson:"after_cash"`         //账变后彩金
	BeforeBonus  int64            `json:"before_bonus" bson:"before_bonus"`     //账变前奖励金
	AfterBonus   int64            `json:"after_bonus" bson:"after_bonus"`       //账变后奖励金
	CashMingTax  int64            `json:"cash_ming_tax" bson:"cash_ming_tax"`   //彩金明税
	BonusMingTax int64            `json:"bonus_ming_tax" bson:"bonus_ming_tax"` //奖励金明税
	CashAnTax    int64            `json:"cash_an_tax" bson:"cash_an_tax"`       //彩金暗税
	BonusAnTax   int64            `json:"bonus_an_tax" bson:"bonus_an_tax"`     //奖励金暗税
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
	MinesPits    []int32 `json:"minesPits" bson:"mines_pits"`        // 坑位*25: 0未知,1已踩无雷,2已踩有雷,3未踩无雷,4未踩有雷
	Step         int32   `json:"step" bson:"step"`                   // 选了几个格子了
	StepPits     []int32 `json:"stepPits" bson:"step_pits"`          // 选的格子
	Multiple     float64 `json:"multiple" bson:"multiple"`           // 返奖倍数
	AutoMines    bool    `json:"autoMines" bson:"auto_mines"`        // 自动对局
	Bets         int64   `json:"bets" bson:"bets"`                   // 总投注
	WinType      int32   `json:"winType" bson:"win_type"`            // 结果:1.赢,2.输,3.强制结算退还下注额
	Force        bool    `json:"force" bson:"force"`                 // 强制结算
	ForceCashOut bool    `json:"forceCashOut" bson:"force_cash_out"` // 玩家一步没走强制结束
}

type FortuneGems2Detail struct {
	UserId string `json:"user_id" bson:"user_id"` //用户ID
	// Score       int64  `json:"score" bson:"score"`               //结算
	BeforeScore int64 `json:"before_score" bson:"before_score"` //账变前分数
	AfterScore  int64 `json:"after_score" bson:"after_score"`   //账变后分数
	BeforeCash  int64 `json:"before_cash" bson:"before_cash"`   //账变前彩金
	AfterCash   int64 `json:"after_cash" bson:"after_cash"`     //账变后彩金

	Bet              int64                           `json:"bet" bson:"bet"`                             // 总投注
	ExtraBet         bool                            `json:"extra_bet" bson:"extra_bet"`                 // 额外投注
	Score            int64                           `json:"score" bson:"score"`                         // 实际扣分
	Reels            map[int32]pb.FortuneGems2Symbol `json:"reels" bson:"reels"`                         // 转轴结果
	SpecificSymbol   pb.FortuneGems2Symbol           `json:"specific_symbol" bson:"specific_symbol"`     // 特定符号
	WinLines         []int32                         `json:"win_lines" bson:"win_lines"`                 // 连线
	WinMultiplier    int32                           `json:"win_multiplier" bson:"win_multiplier"`       // 连线倍数
	NormalWin        int64                           `json:"normal_win" bson:"normal_win"`               // 连线赢得
	WheelMultiplier  int32                           `json:"wheel_multiplier" bson:"wheel_multiplier"`   // 幸运转盘倍数
	ExtraMultipliers map[int32]int32                 `json:"extra_multipliers" bson:"extra_multipliers"` // 幸运转盘额外倍数
	WheelWin         int64                           `json:"wheel_win" bson:"wheel_win"`                 // 幸运转盘赢得
	TotalWin         int64                           `json:"total_win" bson:"total_win"`                 // 本局总赢得
}

type FortuneGemsDetail struct {
	UserId string `json:"user_id" bson:"user_id"` //用户ID
	// Score       int64  `json:"score" bson:"score"`               //结算
	BeforeScore int64 `json:"before_score" bson:"before_score"` //账变前分数
	AfterScore  int64 `json:"after_score" bson:"after_score"`   //账变后分数
	BeforeCash  int64 `json:"before_cash" bson:"before_cash"`   //账变前彩金
	AfterCash   int64 `json:"after_cash" bson:"after_cash"`     //账变后彩金

	Bet            int64                          `json:"bet" bson:"bet"`                         // 总投注
	ExtraBet       bool                           `json:"extra_bet" bson:"extra_bet"`             // 额外投注
	Score          int64                          `json:"score" bson:"score"`                     // 实际扣分
	Reels          map[int32]pb.FortuneGemsSymbol `json:"reels" bson:"reels"`                     // 转轴结果
	SpecificSymbol pb.FortuneGemsSymbol           `json:"specific_symbol" bson:"specific_symbol"` // 特定符号
	WinLines       []int32                        `json:"win_lines" bson:"win_lines"`             // 连线
	WinMultiplier  int32                          `json:"win_multiplier" bson:"win_multiplier"`   // 连线倍数
	NormalWin      int64                          `json:"normal_win" bson:"normal_win"`           // 连线赢得
	TotalWin       int64                          `json:"total_win" bson:"total_win"`             // 本局总赢得
}

func (tpd *TPDetail) FindTPDetailTurn(turn int) *TPDetailTurn {
	if turn >= 0 && turn < len(tpd.TPDetailTurn) {
		return tpd.TPDetailTurn[turn]
	} else {
		tpdt := &TPDetailTurn{Turn: turn}
		tpd.TPDetailTurn = append(tpd.TPDetailTurn, tpdt)
		return tpdt
	}
}

func (jokerd *JOKERDetail) FindJOKERDetailTurn(turn int) *JOKERDetailTurn {
	if turn >= 0 && turn < len(jokerd.JOKERDetailTurn) {
		return jokerd.JOKERDetailTurn[turn]
	} else {
		tpdt := &JOKERDetailTurn{Turn: turn}
		jokerd.JOKERDetailTurn = append(jokerd.JOKERDetailTurn, tpdt)
		return tpdt
	}
}

func (ak47d *AK47Detail) FindAK47DetailTurn(turn int) *AK47DetailTurn {
	if turn >= 0 && turn < len(ak47d.AK47DetailTurn) {
		return ak47d.AK47DetailTurn[turn]
	} else {
		tpdt := &AK47DetailTurn{Turn: turn}
		ak47d.AK47DetailTurn = append(ak47d.AK47DetailTurn, tpdt)
		return tpdt
	}
}

type TPDetailTurn struct {
	Turn      int    `json:"turn" bson:"turn"`           //轮次ID
	Operation string `json:"operation" bson:"operation"` //操作
}

type JOKERDetailTurn struct {
	Turn      int    `json:"turn" bson:"turn"`           //轮次ID
	Operation string `json:"operation" bson:"operation"` //操作
}

type AK47DetailTurn struct {
	Turn      int    `json:"turn" bson:"turn"`           //轮次ID
	Operation string `json:"operation" bson:"operation"` //操作
}

// Save 写入数据库
func (d *Detail) Save() bool {
	//t.Ctime = bson.Now()
	return Insert(Details, d)
}

type TpStrategy struct {
	PRXD TpStrategyPRXD `json:"prxd" bson:"prxd"`
	LJSB TpStrategyLJSB `json:"ljsb" bson:"ljsb"`
	GCYX TpStrategyGCYX `json:"gcyx" bson:"gcyx"`
}

type TpStrategyPRXD struct {
	Active         bool  `json:"active" bson:"active"`                     // 怦然心动生效
	HighCardRounds int32 `json:"high_card_rounds" bson:"high_card_rounds"` // 生效时拿高牌局数
}

type TpStrategyLJSB struct {
	Active bool `json:"active" bson:"active"`  // 乐极生悲判定生效
	JLType int8 `json:"jlType" bson:"jl_type"` // 生效时极乐状态: 1.超率极乐 2.超利极乐 3.利率极乐
	Py     bool `json:"py" bson:"py"`          // 被冤局
	Ps     bool `json:"ps" bson:"ps"`          // 冤杀局
	Pd     bool `json:"pd" bson:"pd"`          // 冤大局
	Pt     bool `json:"pt" bson:"pt"`          // 冤逃局
}

type TpStrategyGCYX struct {
	Active   bool  `json:"active" bson:"active"`       // 高潮涌现生效
	RobotNum int32 `json:"robot_num" bson:"robot_num"` // 高潮人机数
	Rp       bool  `json:"rp" bson:"rp"`               // 压制概率生效
	Bp       bool  `json:"bp" bson:"bp"`               // 恩赐概率生效
}
