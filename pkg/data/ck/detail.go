package ck

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/algo"
	"goserver/pkg/game/config"
	"math"
	"strings"
	"time"

	"github.com/bytedance/sonic"
)

const (
	WinTypeWin      int8 = 1
	WinTypeLose     int8 = 2
	WinTypeTie      int8 = 3
	WinTypeObserver int8 = 4
)

// 人机id18位长度+
func IsRobot(userid string) bool {
	return len(userid) >= 16
}

type Detail struct {
	Ver                                int64     `gorm:"column:ver" json:"ver"`                                                                             // 插入时间戳
	WaterId                            string    `gorm:"column:id" json:"_id"`                                                                              //局号
	UserId                             string    `gorm:"column:userid" json:"userid"`                                                                       //参与玩家ID, 拆players
	Nickname                           string    `gorm:"column:nickname" json:"nickname"`                                                                   //用户名称
	Photo                              string    `gorm:"column:photo" json:"photo"`                                                                         //用户头像
	VipLv                              int32     `gorm:"column:vip_lv" json:"vipLv"`                                                                        //用户vip等级
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
	PlayerStageId                      int       `gorm:"column:player_stage_id" json:"player_stage_id"`                                                     //玩家阶段ID
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
	ChargeInGame                       [][]int32 `gorm:"column:charge_in_game;type:Array(Array(Int32))" json:"charge_in_game"`                              //局内充值触发 [seatid,金额,拉单(1是/0否),实付(1是/0否)]
	RealGame                           bool      `gorm:"column:real_game" json:"real_game"`                                                                 //是否真人对局
	NonDirty                           []uint32  `gorm:"column:non_dirty;type:Array(UInt32)" json:"non_dirty"`                                              //没有污染的人机座位
	ShowCards                          string    `gorm:"column:show_cards" json:"show_cards"`                                                               //亮牌
	IsMustLose                         bool      `gorm:"column:is_must_lose" json:"is_must_lose"`                                                           //是否是必输局
	MustLoseScore                      int       `gorm:"column:must_lose_score" json:"must_lose_score"`                                                     //必输分数线
	CoreRobotSeat                      uint32    `gorm:"column:core_robot_seat" json:"core_robot_seat"`                                                     //核心人机位置
	IsUpCardPool                       bool      `gorm:"column:is_up_card_pool" json:"is_up_card_pool"`                                                     //是否升档牌库
	CardPoolId                         int       `gorm:"column:card_pool_id" json:"card_pool_id"`                                                           //牌库编号
	DealType                           int       `gorm:"column:deal_type" json:"deal_type"`                                                                 //发牌方式
	IsChangeTable                      bool      `gorm:"column:is_change_table" json:"is_change_table"`                                                     // 是否是换桌
	IsTriggerWinScoreLimit             bool      `gorm:"column:is_trigger_win_score_limit" json:"is_trigger_win_score_limit"`                               //是否触发赢分限制
	IsTriggerMustLose                  bool      `gorm:"column:is_trigger_must_lose" json:"is_trigger_must_lose"`                                           //是否触发必输控制
	IsTriggerMustLoseDrawCard          bool      `gorm:"column:is_trigger_must_lose_draw_card" json:"is_trigger_must_lose_draw_card"`                       //是否触发必输控制摸牌限制
	IsTriggerMustLoseLastCard          bool      `gorm:"column:is_trigger_must_lose_last_card" json:"is_trigger_must_lose_last_card"`                       //是否触发必输控制上家出牌限制
	IsTriggerMustLoseCoreRobotGoodCard bool      `gorm:"column:is_trigger_must_lose_core_robot_good_card" json:"is_trigger_must_lose_core_robot_good_card"` //是否触发必输控制核心人机摸好牌
	TPModel                            int32     `gorm:"column:tp_model" json:"tp_model"`                                                                   // tp模式 (0新手，1免费，2正常，3剧情，4控制策略)
	IsTPModel3StoryPlus                bool      `gorm:"column:is_tp_model3story_plus" json:"is_tp_model3story_plus"`                                       // 是否tp剧情局plus
	RMGameOverReason                   int       `gorm:"column:rm_game_over_reason" json:"rm_game_over_reason"`                                             //rm游戏结束原因 (1.天胡,2.自摸胡牌,3.吃牌胡牌,4.对手弃牌)
	// 同步时计算字段 打码量 输赢分 ====================================================================================================
	// Winner         string `gorm:"column:winner" json:"winner"`                     // 赢家userid
	// WinnerRobot    bool   `gorm:"column:winner_robot" json:"winner_robot"`         // 赢家是人机
	WinType           int8  `gorm:"column:win_type" json:"win_type"`                       // 1.赢 2.输 3.平 4.观察局
	Robot             bool  `gorm:"column:robot" json:"robot"`                             // 是否人机
	BetAmount         int64 `gorm:"column:bet_amount" json:"bet_amount"`                   // 本局打码量/总投注
	SettleScore       int64 `gorm:"column:settle_score" json:"settle_score"`               // 对局记录分
	Score             int64 `gorm:"column:score" json:"score"`                             // 对局结算分(减下注减去税)
	ScoreUntax        int64 `gorm:"column:score_untax" json:"score_untax"`                 // 对局结算分(减下注未减税)
	BeforeScore       int64 `json:"before_score" gorm:"column:before_score"`               // 账变前分数
	AfterScore        int64 `json:"after_score" gorm:"column:after_score"`                 // 账变后分数
	CashMingTax       int64 `json:"cash_ming_tax" gorm:"column:cash_ming_tax"`             // 彩金明税
	BonusMingTax      int64 `json:"bonus_ming_tax" gorm:"column:bonus_ming_tax"`           // 奖励金明税
	CashAnTax         int64 `json:"cash_an_tax" gorm:"column:cash_an_tax"`                 // 彩金暗税
	BonusAnTax        int64 `json:"bonus_an_tax" gorm:"column:bonus_an_tax"`               // 奖励金暗税
	PlayerNum         int32 `json:"player_num" gorm:"column:player_num"`                   // 对局玩家数(人机+真人)
	RobotNum          int32 `json:"robot_num" gorm:"column:robot_num"`                     // 对局人机数
	PlayerUserNum     int32 `json:"player_user_num" gorm:"column:player_user_num"`         // 对局真人数
	ChargeTimes       int16 `gorm:"column:charge_times" json:"charge_times"`               // 充值触发次数
	ChargeLaunchTimes int16 `gorm:"column:charge_launch_times" json:"charge_launch_times"` // 局内充值拉单次数
	ChargePayTimes    int16 `gorm:"column:charge_pay_times" json:"charge_pay_times"`       // 局内充值实付次数
	ChargeAmount      int32 `gorm:"column:charge_amount" json:"charge_amount"`             // 局内充值实付金额

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
	// TP计算属性
	TpHandType        uint32 `json:"-" gorm:"column:tp_hand_type"`         // tp手牌牌型
	TpHandTypeUpDown  uint32 `json:"-" gorm:"column:tp_hand_type_up_down"` // tp手牌牌型(10上下)
	TpTurns           int32  `json:"-" gorm:"column:tp_turns"`             // tp轮数
	TpPack            bool   `json:"-" gorm:"column:tp_pack"`              // tp是否弃牌
	TpPackTurn        int32  `json:"-" gorm:"column:tp_pack_turn"`         // tp弃牌轮数
	TpSee             bool   `json:"-" gorm:"column:tp_ses"`               // tp是否看牌
	TpSeeTurn         int32  `json:"-" gorm:"column:tp_see_turn"`          // tp看牌轮数
	TpRaiseTimes      int32  `json:"-" gorm:"column:tp_raise_times"`       // tp加注次数
	TpRaiseBlindTimes int32  `json:"-" gorm:"column:tp_raise_blind_times"` // tp盲加注次数
	TpCall            bool   `json:"-" gorm:"column:tp_call"`              // tp有跟注
	TpCallTimes       int32  `json:"-" gorm:"column:tp_call_times"`        // tp跟注次数
	TpCallBlindTimes  int32  `json:"-" gorm:"column:tp_call_blind_times"`  // tp盲跟注次数
	TpFollowTimes     int32  `json:"-" gorm:"column:tp_follow_times"`      // tp下注次数
	TpRoomUp          bool   `json:"-" gorm:"column:tp_room_up"`           // tp可选两档房间时选高的一档
	TpRoomDown        bool   `json:"-" gorm:"column:tp_room_down"`         // tp可选两档房间时选低的一档
	TpBigWin          bool   `json:"-" gorm:"column:tp_big_win"`           // tp大赢局
	TpBigLose         bool   `json:"-" gorm:"column:tp_big_lose"`          // tp大输局
	TpOpponent        bool   `json:"-" gorm:"column:tp_opponent"`          // tp冤家局(玩家和人机都有同花及以上牌)
	TpHurt            bool   `json:"-" gorm:"column:tp_hurt"`              // tp玩家偷鸡局(玩家不最大赢了,玩家最大输了)
	// TpStrategy        *TpStrategy `bson:"tp_strategy" json:"tp_strategy"` // tp 对局策略
	TpPrxdActive         bool  `json:"-" gorm:"column:tp_prxd_active"`           // 怦然心动生效
	TpPrxdHighCardRounds int32 `json:"-" gorm:"column:tp_prxd_high_card_rounds"` // 怦然心动生效时连续拿高牌局数
	TpLjsbActive         bool  `json:"-" gorm:"column:tp_ljsb_active"`           // 乐极生悲判定生效
	TpLjsbJLType         int8  `json:"-" gorm:"column:tp_ljsb_jl_type"`          // 生效时极乐状态: 1.超率极乐 2.超利极乐 3.利率极乐
	TpLjsbPy             bool  `json:"-" gorm:"column:tp_ljsb_py"`               // 被冤局
	TpLjsbPs             bool  `json:"-" gorm:"column:tp_ljsb_ps"`               // 冤杀局
	TpLjsbPd             bool  `json:"-" gorm:"column:tp_ljsb_pd"`               // 冤大局
	TpLjsbPt             bool  `json:"-" gorm:"column:tp_ljsb_pt"`               // 冤逃局
	TpGcyxActive         bool  `json:"-" gorm:"column:tp_gcyx_active"`           // 高潮涌现判定生效
	TpGcyxRobotNum       int32 `json:"-" gorm:"column:tp_gcyx_robot_num"`        // 高潮人机数
	TpGcyxRp             bool  `json:"-" gorm:"column:tp_gcyx_rp"`               // 压制概率生效
	TpGcyxBp             bool  `json:"-" gorm:"column:tp_gcyx_bp"`               // 恩赐概率生效

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
	UpPointValue int32            `json:"-" gorm:"column:up_point_value"`                      //点数
	UpBets       int64            `json:"-" gorm:"column:up_bets"`                             //总下注
	UpPlayerWin  int64            `json:"-" gorm:"column:up_player_win"`                       //玩家赢分
	UpStrategyId int              `json:"-" gorm:"column:up_strategy_id"`                      //策略id
	UpIsSuppress bool             `json:"-" gorm:"column:up_is_suppress"`                      //是否压制
	UpCrit7up    bool             `json:"-" gorm:"column:up_crit_7up"`                         //是否暴击7up对局,区分老数据
	UpCrited     int16            `json:"-" gorm:"column:up_crited"`                           //触发暴击位置数
	UpCritOdds   map[uint32]int32 `json:"-" gorm:"column:up_crit_odds;type:Map(UInt32,Int32)"` //触发暴击位置和倍数
	// UpUserDetail 用户详细展开
	UpUserid         string           `json:"-" gorm:"column:up_userid"`                           //用户id
	UpDragon         int64            `json:"-" gorm:"column:up_dragon"`                           //龙下注
	UpTiger          int64            `json:"-" gorm:"column:up_tiger"`                            //虎下注
	UpTie            int64            `json:"-" gorm:"column:up_tie"`                              //和下注
	UpObserve        bool             `json:"-" gorm:"column:up_observe"`                          //观察局
	UpResult         string           `json:"-" gorm:"column:up_result"`                           //输赢平
	UpWin            int64            `json:"-" gorm:"column:up_win"`                              //输赢
	UpSeatBets       map[uint32]int64 `json:"-" gorm:"column:up_seat_bets;type:Map(UInt32,Int64)"` //_位置下注: 0小,1大,2,3,4,...,12
	UpCrit           int32            `json:"-" gorm:"column:up_crit"`                             //_暴击位押中次数
	UpBeforeScore    int64            `json:"-" gorm:"column:up_before_score"`                     //账变前分数
	UpAfterScore     int64            `json:"-" gorm:"column:up_after_score"`                      //账变后分数
	UpBeforeCash     int64            `json:"-" gorm:"column:up_before_cash"`                      //账变前彩金
	UpAfterCash      int64            `json:"-" gorm:"column:up_after_cash"`                       //账变后彩金
	UpBeforeBonus    int64            `json:"-" gorm:"column:up_before_bonus"`                     //账变前奖励金
	UpAfterBonus     int64            `json:"-" gorm:"column:up_after_bonus"`                      //账变后奖励金
	UpCashMingTax    int64            `json:"-" gorm:"column:up_cash_ming_tax"`                    //彩金明税
	UpBonusMingTax   int64            `json:"-" gorm:"column:up_bonus_ming_tax"`                   //奖励金明税
	UpCashAnTax      int64            `json:"-" gorm:"column:up_cash_an_tax"`                      //彩金暗税
	UpBonusAnTax     int64            `json:"-" gorm:"column:up_bonus_an_tax"`                     //奖励金暗税
	UpBeforeBackRate int              `json:"-" gorm:"column:up_before_back_rate"`                 //账变前返奖率
	UpAfterBackRate  int              `json:"-" gorm:"column:up_after_back_rate"`                  //账变后返奖率

	// Crash详情展开 ====================================================================================================
	CrashWinResult    string  `json:"-" gorm:"column:crash_win_result"`                      //开奖结果
	CrashWinResult2   int32   `json:"-" gorm:"column:crash_win_result2"`                     //开奖结果*100
	CrashBets         int64   `json:"-" gorm:"column:crash_bets"`                            //总下注
	CrashPlayerLose   int64   `json:"-" gorm:"column:crash_player_lose"`                     //玩家赢分
	CrashStrategyType []int32 `json:"-" gorm:"column:crash_strategy_type;type:Array(Int32)"` //策略类型
	CrashRkyhRfge     bool    `json:"-" gorm:"column:crash_rkyh_rfge"`                       //人狂有祸r肥割
	// CrashUserDetail 用户详细展开
	CrashUserid       string `json:"-" gorm:"column:crash_userid"`         //用户id
	CrashObserve      bool   `json:"-" gorm:"column:crash_observe"`        //是否观察局
	CrashBet          int64  `json:"-" gorm:"column:crash_bet"`            //下注
	CrashBet0         int64  `json:"-" gorm:"column:crash_bet0"`           //位置0下注
	CrashBet1         int64  `json:"-" gorm:"column:crash_bet1"`           //位置1下注
	CrashMultiple     string `json:"-" gorm:"column:crash_multiple"`       //位置0逃脱倍数
	CrashMultiple2    int32  `json:"-" gorm:"column:crash_multiple2"`      //位置0逃脱倍数*100
	CrashMultiple1    string `json:"-" gorm:"column:crash_multiple1"`      //位置1逃脱倍数
	CrashMultiple1_2  int32  `json:"-" gorm:"column:crash_multiple1_2"`    //位置1逃脱倍数*100
	CrashResult       string `json:"-" gorm:"column:crash_result"`         //输赢平
	CrashWin          int64  `json:"-" gorm:"column:crash_win"`            //输赢
	CrashWin0         int64  `json:"-" gorm:"column:crash_win0"`           //位置0输赢
	CrashWin1         int64  `json:"-" gorm:"column:crash_win1"`           //位置1输赢
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
	CrashRobot        bool   `json:"-" gorm:"column:crash_robot"`          //人机
	CrashNextBet0     bool   `json:"-" gorm:"column:crash_next_bet0"`      //位置0是否上一轮下注
	CrashNextBet1     bool   `json:"-" gorm:"column:crash_next_bet1"`      //位置1是否上一轮下注
	CrashAutoCrash0   bool   `json:"-" gorm:"column:crash_auto_crash0"`    //位置0是否自动逃离
	CrashAutoCrash1   bool   `json:"-" gorm:"column:crash_auto_crash1"`    //位置1是否自动逃离
	CrashAutoCrashM0  int32  `json:"-" gorm:"column:crash_auto_crash_m0"`  //位置0自动逃离倍数
	CrashAutoCrashM1  int32  `json:"-" gorm:"column:crash_auto_crash_m1"`  //位置1自动逃离倍数

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

	// MinesDetail 地雷详细展开
	MinesUserId       string  `json:"-" gorm:"column:mines_user_id"`                       //用户ID
	MinesScore        int64   `json:"-" gorm:"column:mines_score"`                         //结算
	MinesBeforeScore  int64   `json:"-" gorm:"column:mines_before_score"`                  //账变前分数
	MinesAfterScore   int64   `json:"-" gorm:"column:mines_after_score"`                   //账变后分数
	MinesBeforeCash   int64   `json:"-" gorm:"column:mines_before_cash"`                   //账变前彩金
	MinesAfterCash    int64   `json:"-" gorm:"column:mines_after_cash"`                    //账变后彩金
	MinesBeforeBonus  int64   `json:"-" gorm:"column:mines_before_bonus"`                  //账变前奖励金
	MinesAfterBonus   int64   `json:"-" gorm:"column:mines_after_bonus"`                   //账变后奖励金
	MinesMines        int32   `json:"-" gorm:"column:mines_mines"`                         // 埋雷数
	MinesMinesPits    []int32 `json:"-" gorm:"column:mines_mines_pits;type:Array(UInt32)"` // 坑位*25: 0未知,1已踩无雷,2已踩有雷,3未踩无雷,4未踩有雷
	MinesStep         int32   `json:"-" gorm:"column:mines_step"`                          // 选了几个格子了
	MinesStepPits     []int32 `json:"-" gorm:"column:mines_step_pits;type:Array(UInt32)"`  // 选的格子
	MinesMultiple     float64 `json:"-" gorm:"column:mines_multiple"`                      // 返奖倍数
	MinesAutoMines    bool    `json:"-" gorm:"column:mines_auto_mines"`                    // 自动对局
	MinesBets         int64   `json:"-" gorm:"column:mines_bets"`                          // 总投注
	MinesWinType      int32   `json:"-" gorm:"column:mines_win_type"`                      // 结果:1.赢,2.输,3.强制结算退还下注额
	MinesForce        bool    `json:"-" gorm:"column:mines_force"`                         // 强制结算
	MinesForceCashOut bool    `json:"-" gorm:"column:mines_force_cash_out"`                // 玩家一步没走强制结束

	//fortune_gems2详情展开
	FortuneGems2UserId           string          `json:"-" gorm:"column:fortune_gems2_user_id"`                                 //用户ID
	FortuneGems2BeforeScore      int64           `json:"-" gorm:"column:fortune_gems2_before_score"`                            //账变前分数
	FortuneGems2AfterScore       int64           `json:"-" gorm:"column:fortune_gems2_after_score"`                             //账变后分数
	FortuneGems2BeforeCash       int64           `json:"-" gorm:"column:fortune_gems2_before_cash"`                             //账变前彩金
	FortuneGems2AfterCash        int64           `json:"-" gorm:"column:fortune_gems2_after_cash"`                              //账变后彩金
	FortuneGems2Bet              int64           `json:"-" gorm:"column:fortune_gems2_bet"`                                     // 总投注
	FortuneGems2ExtraBet         bool            `json:"-" gorm:"column:fortune_gems2_extra_bet"`                               // 额外投注
	FortuneGems2Score            int64           `json:"-" gorm:"column:fortune_gems2_score"`                                   // 实际扣分
	FortuneGems2Reels            map[int32]int32 `json:"-" gorm:"column:fortune_gems2_reels;type:Map(Int32,Int32)"`             // 转轴结果
	FortuneGems2SpecificSymbol   int32           `json:"-" gorm:"column:fortune_gems2_specific_symbol"`                         // 特定符号
	FortuneGems2WinLines         []int32         `json:"-" gorm:"column:fortune_gems2_win_lines;type:Array(Int32)"`             // 连线
	FortuneGems2WinMultiplier    int32           `json:"-" gorm:"column:fortune_gems2_win_multiplier"`                          // 连线倍数
	FortuneGems2NormalWin        int64           `json:"-" gorm:"column:fortune_gems2_normal_win"`                              // 连线赢得
	FortuneGems2WheelMultiplier  int32           `json:"-" gorm:"column:fortune_gems2_wheel_multiplier"`                        // 幸运转盘倍数
	FortuneGems2ExtraMultipliers map[int32]int32 `json:"-" gorm:"column:fortune_gems2_extra_multipliers;type:Map(Int32,Int32)"` // 幸运转盘额外倍数
	FortuneGems2WheelWin         int64           `json:"-" gorm:"column:fortune_gems2_wheel_win"`                               // 幸运转盘赢得
	FortuneGems2TotalWin         int64           `json:"-" gorm:"column:fortune_gems2_total_win"`                               // 本局总赢得

	//fortune_gems详情展开
	FortuneGemsUserId         string          `json:"-" gorm:"column:fortune_gems_user_id"`                     //用户ID
	FortuneGemsBeforeScore    int64           `json:"-" gorm:"column:fortune_gems_before_score"`                //账变前分数
	FortuneGemsAfterScore     int64           `json:"-" gorm:"column:fortune_gems_after_score"`                 //账变后分数
	FortuneGemsBeforeCash     int64           `json:"-" gorm:"column:fortune_gems_before_cash"`                 //账变前彩金
	FortuneGemsAfterCash      int64           `json:"-" gorm:"column:fortune_gems_after_cash"`                  //账变后彩金
	FortuneGemsBet            int64           `json:"-" gorm:"column:fortune_gems_bet"`                         // 总投注
	FortuneGemsExtraBet       bool            `json:"-" gorm:"column:fortune_gems_extra_bet"`                   // 额外投注
	FortuneGemsScore          int64           `json:"-" gorm:"column:fortune_gems_score"`                       // 实际扣分
	FortuneGemsReels          map[int32]int32 `json:"-" gorm:"column:fortune_gems_reels;type:Map(Int32,Int32)"` // 转轴结果
	FortuneGemsSpecificSymbol int32           `json:"-" gorm:"column:fortune_gems_specific_symbol"`             // 特定符号
	FortuneGemsWinLines       []int32         `json:"-" gorm:"column:fortune_gems_win_lines;type:Array(Int32)"` // 连线
	FortuneGemsWinMultiplier  int32           `json:"-" gorm:"column:fortune_gems_win_multiplier"`              // 连线倍数
	FortuneGemsNormalWin      int64           `json:"-" gorm:"column:fortune_gems_normal_win"`                  // 连线赢得
	FortuneGemsTotalWin       int64           `json:"-" gorm:"column:fortune_gems_total_win"`                   // 本局总赢得
}

func (*Detail) TableName() string {
	return "col_detail"
}

func (*Detail) New() CkEntity {
	return new(Detail)
}

func (c *Detail) ParseData(d interface{}) (rs []interface{}, err error) {
	// check type
	from, ok := d.(*data.Detail)
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
	c.Begin = time.Unix(c.BeginTime, 0)
	c.End = time.Unix(c.EndTime, 0)

	// 根据gtype转换
	var details []*Detail
	switch from.Gtype {
	case int32(pb.HUA), int32(pb.HUA2):
		details, err = c.parseTp(from, c)
	case int32(pb.JOKER):
		details, err = c.parseJoker(from, c)
	case int32(pb.AK47):
		details, err = c.parseAk47(from, c)
	case int32(pb.RUMMY), int32(pb.RUMMY2):
		details, err = c.parseRummy(from, c)
	case int32(pb.LHD):
		details, err = c.parseLhd(from, c)
	case int32(pb.SEVEN):
		details, err = c.parse7up(from, c)
	case int32(pb.CRASH), int32(pb.PLANE):
		details, err = c.parseCrash(from, c)
	case int32(pb.ABAR):
		details, err = c.parseAndarBahar(from, c)
	case int32(pb.LOTTERY):
		details, err = c.parseCp(from, c)
	case int32(pb.REDBLACK):
		details, err = c.parseRedBlack(from, c)
	case int32(pb.MINES):
		details, err = c.parseMines(from, c)
	case int32(pb.FORTUNE_GEMS2):
		details, err = c.parseFortuneGems2(from, c)
	case int32(pb.FORTUNE_GEMS):
		details, err = c.parseFortuneGems(from, c)
	}
	if err != nil {
		return
	}

	for _, det := range details {
		rs = append(rs, det)
	}
	return
}

func (c *Detail) parseTp(src *data.Detail, dst *Detail) (details []*Detail, err error) {
	bytes, err := sonic.Marshal(dst) // 基本属性copy
	var playerNum, robotNum, playerUserNum int32
	for _, user := range src.TPDetail {
		playerNum++
		if IsRobot(user.UserId) {
			robotNum++
		} else {
			playerUserNum++
		}
	}

	for _, user := range src.TPDetail {
		dtl := new(Detail)
		if err = sonic.Unmarshal(bytes, dtl); err != nil {
			return
		}
		dtl.UserId = user.UserId
		dtl.TpSeatId = user.SeatId
		dtl.TpUserId = user.UserId
		dtl.TpCards = user.Cards
		dtl.TpScore = user.Score
		dtl.TpBeforeScore = user.BeforeScore
		dtl.TpAfterScore = user.AfterScore
		dtl.TpBeforeCash = user.BeforeCash
		dtl.TpAfterCash = user.AfterCash
		dtl.TpBeforeBonus = user.BeforeBonus
		dtl.TpAfterBonus = user.AfterBonus
		dtl.TpBet = user.Bet
		dtl.TpBottom = user.Bottom
		dtl.TpCashStock = user.CashStock
		dtl.TpBonusStock = user.BonusStock
		dtl.TpCashMingTax = user.CashMingTax
		dtl.TpBonusMingTax = user.BonusMingTax
		dtl.TpCashAnTax = user.CashAnTax
		dtl.TpBonusAnTax = user.BonusAnTax
		dtl.TpNewBieProbeId = user.NewBieProbeId
		dtl.TpNewBieProbeRound = user.NewBieProbeRound
		for _, turn := range user.TPDetailTurn {
			dtl.TpTurn = append(dtl.TpTurn, int32(turn.Turn))
			dtl.TpOperation = append(dtl.TpOperation, turn.Operation)
		}
		if src.TpStrategy != nil {
			dtl.TpPrxdActive = src.TpStrategy.PRXD.Active
			dtl.TpPrxdHighCardRounds = src.TpStrategy.PRXD.HighCardRounds

			dtl.TpLjsbActive = src.TpStrategy.LJSB.Active
			dtl.TpLjsbJLType = src.TpStrategy.LJSB.JLType
			dtl.TpLjsbPy = src.TpStrategy.LJSB.Py
			dtl.TpLjsbPs = src.TpStrategy.LJSB.Ps
			dtl.TpLjsbPd = src.TpStrategy.LJSB.Pd
			dtl.TpLjsbPt = src.TpStrategy.LJSB.Pt

			dtl.TpGcyxActive = src.TpStrategy.GCYX.Active
			dtl.TpGcyxRobotNum = src.TpStrategy.GCYX.RobotNum
			dtl.TpGcyxRp = src.TpStrategy.GCYX.Rp
			dtl.TpGcyxBp = src.TpStrategy.GCYX.Bp
		}

		// 计算属性
		if user.Score > 0 {
			dtl.WinType = WinTypeWin
		} else {
			dtl.WinType = WinTypeLose
		}
		dtl.Robot = IsRobot(dtl.UserId)
		dtl.BetAmount = user.Bet
		dtl.SettleScore = user.Score
		if user.Score > 0 {
			dtl.Score = user.Score - user.Bet
			dtl.ScoreUntax = user.Score - user.Bet + user.CashMingTax
		} else {
			dtl.Score = user.Score
			dtl.ScoreUntax = user.Score
		}
		dtl.BeforeScore = user.BeforeScore
		dtl.AfterScore = user.AfterScore
		dtl.CashMingTax = user.CashMingTax
		dtl.BonusMingTax = user.BonusMingTax
		dtl.CashAnTax = user.CashAnTax
		dtl.BonusAnTax = user.BonusAnTax
		dtl.PlayerNum = playerNum
		dtl.RobotNum = robotNum
		dtl.PlayerUserNum = playerUserNum

		// 用户局内充值
		for _, charge := range src.ChargeInGame {
			if charge[0] == int32(user.SeatId) {
				dtl.ChargeTimes++
				if charge[2] == 1 {
					dtl.ChargeLaunchTimes++
				}
				if charge[3] == 1 {
					dtl.ChargePayTimes++
					dtl.ChargeAmount += charge[1]
				}
			}
		}

		// tp统计属性
		err = c.parseTpStats(src, user, dtl)
		if err != nil {
			return
		}

		details = append(details, dtl)
	}
	return
}

// tp 统计属性
func (c *Detail) parseTpStats(src *data.Detail, user *data.TPDetail, dtl *Detail) (err error) {
	dtl.TpHandType = algo.HuaType(user.Cards)
	dtl.TpHandTypeUpDown = uint32(algo.HuaTypeUpOrDown(user.Cards))
	dtl.TpTurns = int32(len(dtl.TpTurn))
	// config.GetGame(user.RoomId)
	for i, turn := range dtl.TpTurn {
		opt := dtl.TpOperation[i]
		// 弃牌
		if strings.Contains(opt, "pack") {
			if !dtl.TpPack {
				dtl.TpPack = true
				dtl.TpPackTurn = turn
			}
		} else if strings.Contains(opt, "see") {
			if !dtl.TpSee {
				dtl.TpSee = true
				dtl.TpSeeTurn = turn
			}
		} else if strings.Contains(opt, "blindx2") {
			dtl.TpRaiseBlindTimes++
			dtl.TpFollowTimes++
		} else if strings.Contains(opt, "chaalx2") {
			dtl.TpRaiseTimes++
			dtl.TpFollowTimes++
		} else if strings.Contains(opt, "blind") {
			dtl.TpCallBlindTimes++
			dtl.TpFollowTimes++
		} else if strings.Contains(opt, "chaal") {
			dtl.TpCallTimes++
			dtl.TpFollowTimes++
			dtl.TpCall = true
		}
	}

	// 10倍底注大输赢局
	var bigwin = user.Bottom > 0 && math.Abs(float64((user.Score-user.Bet)/user.Bottom)) > 10
	if bigwin {
		if user.Score > 0 {
			dtl.TpBigWin = true
		} else {
			dtl.TpBigLose = true
		}
	}

	// 玩家主动升降档, tp可选两档房间时选高/低的一档
	if config.GameMap != nil {
		score := user.BeforeScore
		var minRoom, maxRoom string
		var minBottom, maxBottom int
		for _, game := range config.GetGames() {
			if (game.Gtype == int32(pb.HUA) || game.Gtype == int32(pb.HUA2)) && game.RoomType == int(src.Rtype) && game.Status == 1 {
				if (game.Min_Access == -1 || score >= int64(game.Min_Access)) &&
					(game.Max_Access == -1 || score <= int64(game.Max_Access)) {
					if minBottom == 0 || minBottom > game.TP.Bottom {
						minBottom = game.TP.Bottom
						minRoom = game.Id
					}
					if maxBottom == 0 || maxBottom < game.TP.Bottom {
						maxBottom = game.TP.Bottom
						maxRoom = game.Id
					}
				}
			}
		}
		if minRoom != "" && maxRoom != "" && minRoom != maxRoom {
			if src.RoomId == maxRoom {
				dtl.TpRoomUp = true
			}
			if src.RoomId == minRoom {
				dtl.TpRoomDown = true
			}
		}
	}

	// 冤人局 是否偷鸡
	var rHasColor, pHasColor bool
	dtlMax, dtlWin := true, user.Score > 0
	for _, udtl := range src.TPDetail {
		if algo.HuaType(udtl.Cards) >= algo.TongHua {
			if IsRobot(udtl.UserId) {
				rHasColor = true
			} else {
				pHasColor = true
			}
		}
		if udtl.UserId != dtl.UserId && algo.HuaCompare(udtl.Cards, user.Cards) {
			dtlMax = false
		}
	}
	dtl.TpOpponent = rHasColor && pHasColor
	dtl.TpHurt = (dtlMax && !dtlWin) || (!dtlMax && dtlWin)
	return
}

func (c *Detail) parseJoker(src *data.Detail, dst *Detail) (details []*Detail, err error) {
	bytes, err := sonic.Marshal(dst) // 基本属性copy
	var playerNum, robotNum, playerUserNum int32
	for _, user := range src.JOKERDetail {
		playerNum++
		if IsRobot(user.UserId) {
			robotNum++
		} else {
			playerUserNum++
		}
	}
	for _, user := range src.JOKERDetail {
		dtl := new(Detail)
		if err = sonic.Unmarshal(bytes, dtl); err != nil {
			return
		}
		dtl.UserId = user.UserId
		dtl.JokerSeatId = user.SeatId
		dtl.JokerUserId = user.UserId
		dtl.JokerCards = user.Cards
		dtl.JokerScore = user.Score
		dtl.JokerBeforeScore = user.BeforeScore
		dtl.JokerAfterScore = user.AfterScore
		dtl.JokerBeforeCash = user.BeforeCash
		dtl.JokerAfterCash = user.AfterCash
		dtl.JokerBeforeBonus = user.BeforeBonus
		dtl.JokerAfterBonus = user.AfterBonus
		dtl.JokerBet = user.Bet
		dtl.JokerBottom = user.Bottom
		dtl.JokerCashStock = user.CashStock
		dtl.JokerBonusStock = user.BonusStock
		dtl.JokerCashMingTax = user.CashMingTax
		dtl.JokerBonusMingTax = user.BonusMingTax
		dtl.JokerCashAnTax = user.CashAnTax
		dtl.JokerBonusAnTax = user.BonusAnTax
		for _, turn := range user.JOKERDetailTurn {
			dtl.JokerTurn = append(dtl.JokerTurn, int32(turn.Turn))
			dtl.JokerOperation = append(dtl.JokerOperation, turn.Operation)
		}

		// 计算属性
		if user.Score > 0 {
			dtl.WinType = WinTypeWin
		} else {
			dtl.WinType = WinTypeLose
		}
		dtl.Robot = IsRobot(dtl.UserId)
		dtl.BetAmount = user.Bet
		dtl.SettleScore = user.Score
		if user.Score > 0 {
			dtl.Score = user.Score - user.Bet
			dtl.ScoreUntax = user.Score - user.Bet + user.CashMingTax
		} else {
			dtl.Score = user.Score
			dtl.ScoreUntax = user.Score
		}
		dtl.BeforeScore = user.BeforeScore
		dtl.AfterScore = user.AfterScore
		dtl.CashMingTax = user.CashMingTax
		dtl.BonusMingTax = user.BonusMingTax
		dtl.CashAnTax = user.CashAnTax
		dtl.BonusAnTax = user.BonusAnTax
		dtl.PlayerNum = playerNum
		dtl.RobotNum = robotNum
		dtl.PlayerUserNum = playerUserNum
		details = append(details, dtl)
	}
	return
}

func (c *Detail) parseAk47(src *data.Detail, dst *Detail) (details []*Detail, err error) {
	bytes, err := sonic.Marshal(dst) // 基本属性copy
	var playerNum, robotNum, playerUserNum int32
	for _, user := range src.AK47Detail {
		playerNum++
		if IsRobot(user.UserId) {
			robotNum++
		} else {
			playerUserNum++
		}
	}
	for _, user := range src.AK47Detail {
		dtl := new(Detail)
		if err = sonic.Unmarshal(bytes, dtl); err != nil {
			return
		}
		dtl.UserId = user.UserId
		dtl.Ak47SeatId = user.SeatId
		dtl.Ak47UserId = user.UserId
		dtl.Ak47Cards = user.Cards
		dtl.Ak47Score = user.Score
		dtl.Ak47BeforeScore = user.BeforeScore
		dtl.Ak47AfterScore = user.AfterScore
		dtl.Ak47BeforeCash = user.BeforeCash
		dtl.Ak47AfterCash = user.AfterCash
		dtl.Ak47BeforeBonus = user.BeforeBonus
		dtl.Ak47AfterBonus = user.AfterBonus
		dtl.Ak47Bet = user.Bet
		dtl.Ak47Bottom = user.Bottom
		dtl.Ak47CashStock = user.CashStock
		dtl.Ak47BonusStock = user.BonusStock
		dtl.Ak47CashMingTax = user.CashMingTax
		dtl.Ak47BonusMingTax = user.BonusMingTax
		dtl.Ak47CashAnTax = user.CashAnTax
		dtl.Ak47BonusAnTax = user.BonusAnTax
		for _, turn := range user.AK47DetailTurn {
			dtl.Ak47Turn = append(dtl.Ak47Turn, int32(turn.Turn))
			dtl.Ak47Operation = append(dtl.Ak47Operation, turn.Operation)
		}

		// 计算属性
		if user.Score > 0 {
			dtl.WinType = WinTypeWin
		} else {
			dtl.WinType = WinTypeLose
		}
		dtl.Robot = IsRobot(dtl.UserId)
		dtl.BetAmount = user.Bet
		dtl.SettleScore = user.Score
		if user.Score > 0 {
			dtl.Score = user.Score - user.Bet
			dtl.ScoreUntax = user.Score - user.Bet + user.CashMingTax
		} else {
			dtl.Score = user.Score
			dtl.ScoreUntax = user.Score
		}
		dtl.BeforeScore = user.BeforeScore
		dtl.AfterScore = user.AfterScore
		dtl.CashMingTax = user.CashMingTax
		dtl.BonusMingTax = user.BonusMingTax
		dtl.CashAnTax = user.CashAnTax
		dtl.BonusAnTax = user.BonusAnTax
		dtl.PlayerNum = playerNum
		dtl.RobotNum = robotNum
		dtl.PlayerUserNum = playerUserNum
		details = append(details, dtl)
	}
	return
}

func (c *Detail) parseRummy(src *data.Detail, dst *Detail) (details []*Detail, err error) {
	bytes, err := sonic.Marshal(dst) // 基本属性copy
	var playerNum, robotNum, playerUserNum int32
	for _, user := range src.TPDetail {
		playerNum++
		if IsRobot(user.UserId) {
			robotNum++
		} else {
			playerUserNum++
		}
	}
	for _, user := range src.RMDetail {
		dtl := new(Detail)
		if err = sonic.Unmarshal(bytes, dtl); err != nil {
			return
		}
		dtl.RmcOk = src.RMControl != nil
		if dtl.RmcOk {
			dtl.RmcCtype = src.RMControl.Ctype
			dtl.RmcRoiId = src.RMControl.RoiId
			dtl.RmcDropId = src.RMControl.DropId
			dtl.RmcControlEffect = src.RMControl.ControlEffect
			dtl.RmcHierarchy = src.RMControl.Hierarchy
			dtl.RmcDrawNumPlayer = src.RMControl.DrawNumPlayer
			dtl.RmcDrawNumRobot = src.RMControl.DrawNumRobot
			dtl.RmcNotDrawRobot1st = src.RMControl.NotDrawRobot1st
			dtl.RmcNotDrawPlayer1st = src.RMControl.NotDrawPlayer1st
			dtl.RmcControlRobotDraw = src.RMControl.ControlRobotDraw
			dtl.RmcControlRobotDrawRound = src.RMControl.ControlRobotDrawRound
		}

		dtl.UserId = user.UserId
		dtl.RmSeatId = user.SeatId
		dtl.RmUserId = user.UserId
		dtl.RmResult = user.Result
		dtl.RmScore = user.Score
		dtl.RmBeforeScore = user.BeforeScore
		dtl.RmAfterScore = user.AfterScore
		dtl.RmBeforeCash = user.BeforeCash
		dtl.RmAfterCash = user.AfterCash
		dtl.RmBeforeBonus = user.BeforeBonus
		dtl.RmAfterBonus = user.AfterBonus
		dtl.RmFinalCards = user.FinalCards
		dtl.RmInitCards = user.InitCards
		dtl.RmWildCard = user.WildCard
		dtl.RmMoCards = user.MoCards
		dtl.RmChuCards = user.ChuCards
		dtl.RmCashStock = user.CashStock
		dtl.RmBonusStock = user.BonusStock
		dtl.RmCashMingTax = user.CashMingTax
		dtl.RmBonusMingTax = user.BonusMingTax
		dtl.RmCashAnTax = user.CashAnTax
		dtl.RmBonusAnTax = user.BonusAnTax

		// 计算属性
		switch user.Result {
		case "赢":
			dtl.WinType = WinTypeWin
		case "输":
			dtl.WinType = WinTypeLose
		case "平":
			dtl.WinType = WinTypeTie
		default:
			if user.Score > 0 {
				dtl.WinType = WinTypeWin
			} else if user.Score < 0 {
				dtl.WinType = WinTypeLose
			} else {
				dtl.WinType = WinTypeTie
			}
		}

		dtl.Robot = IsRobot(dtl.UserId)
		dtl.BetAmount = 0
		dtl.SettleScore = user.Score
		dtl.Score = user.Score
		dtl.ScoreUntax = user.Score
		if user.Score > 0 {
			dtl.ScoreUntax = user.Score + user.CashMingTax
		}
		dtl.BeforeScore = user.BeforeScore
		dtl.AfterScore = user.AfterScore
		dtl.CashMingTax = user.CashMingTax
		dtl.BonusMingTax = user.BonusMingTax
		dtl.CashAnTax = user.CashAnTax
		dtl.BonusAnTax = user.BonusAnTax
		dtl.PlayerNum = playerNum
		dtl.RobotNum = robotNum
		dtl.PlayerUserNum = playerUserNum
		if config.GameMap != nil {
			if g, ok := config.GameMap.Load(dst.RoomId); ok {
				if game, ok := g.(data.Game); ok {
					bottom := game.GetBottom()
					dtl.BetAmount = bottom * 80
				}
			}
		}

		details = append(details, dtl)
	}
	return
}

func (c *Detail) parseLhd(src *data.Detail, dst *Detail) (details []*Detail, err error) {
	if src.LHDetail == nil {
		details = append(details, dst)
		return
	}

	bytes, err := sonic.Marshal(dst) // 基本属性copy
	for _, user := range src.LHDetail.UserDetail {
		dtl := new(Detail)
		if err = sonic.Unmarshal(bytes, dtl); err != nil {
			return
		}
		dtl.LhdWinner = src.LHDetail.Winner
		dtl.LhdDragonValue = src.LHDetail.DragonValue
		dtl.LhdTigerValue = src.LHDetail.TigerValue
		dtl.LhdBets = src.LHDetail.Bets
		dtl.LhdPlayerWin = src.LHDetail.PlayerWin
		dtl.LhdStrategyId = src.LHDetail.StrategyId
		dtl.LhdIsSuppress = src.LHDetail.IsSuppress

		dtl.UserId = user.Userid
		dtl.LhdUserid = user.Userid
		dtl.LhdDragon = user.Dragon
		dtl.LhdTiger = user.Tiger
		dtl.LhdTie = user.Tie
		dtl.LhdObserve = user.Observe
		dtl.LhdResult = user.Result
		dtl.LhdWin = user.Win
		dtl.LhdBeforeScore = user.BeforeScore
		dtl.LhdAfterScore = user.AfterScore
		dtl.LhdBeforeCash = user.BeforeCash
		dtl.LhdAfterCash = user.AfterCash
		dtl.LhdBeforeBonus = user.BeforeBonus
		dtl.LhdAfterBonus = user.AfterBonus
		dtl.LhdCashMingTax = user.CashMingTax
		dtl.LhdBonusMingTax = user.BonusMingTax
		dtl.LhdCashAnTax = user.CashAnTax
		dtl.LhdBonusAnTax = user.BonusAnTax
		dtl.LhdBeforeBackRate = user.BeforeBackRate
		dtl.LhdAfterBackRate = user.AfterBackRate

		// 计算属性

		// 计算属性
		if user.Observe {
			dtl.WinType = WinTypeObserver
		} else {
			switch user.Result {
			case "赢":
				dtl.WinType = WinTypeWin
			case "输":
				dtl.WinType = WinTypeLose
			case "平":
				dtl.WinType = WinTypeTie
			default:
				if user.Win > 0 {
					dtl.WinType = WinTypeWin
				} else if user.Win < 0 {
					dtl.WinType = WinTypeLose
				} else {
					dtl.WinType = WinTypeTie
				}
			}
		}

		dtl.Robot = IsRobot(dtl.UserId)
		dtl.BetAmount = user.Dragon + user.Tiger + user.Tie
		dtl.SettleScore = user.Win
		dtl.Score = user.Win - user.CashMingTax
		dtl.ScoreUntax = user.Win
		dtl.BeforeScore = user.BeforeScore
		dtl.AfterScore = user.AfterScore
		dtl.CashMingTax = user.CashMingTax
		dtl.BonusMingTax = user.BonusMingTax
		dtl.CashAnTax = user.CashAnTax
		dtl.BonusAnTax = user.BonusAnTax
		dtl.PlayerNum = int32(len(src.LHDetail.UserDetail))
		dtl.PlayerUserNum = dtl.PlayerNum
		details = append(details, dtl)
	}
	return
}

func (c *Detail) parse7up(src *data.Detail, dst *Detail) (details []*Detail, err error) {
	if src.UPDetail == nil {
		details = append(details, dst)
		return
	}

	bytes, err := sonic.Marshal(dst) // 基本属性copy
	for _, user := range src.UPDetail.UserDetail {
		dtl := new(Detail)
		if err = sonic.Unmarshal(bytes, dtl); err != nil {
			return
		}
		dtl.UpWinner = src.UPDetail.Winner
		dtl.UpPointValue = src.UPDetail.PointValue
		dtl.UpBets = src.UPDetail.Bets
		dtl.UpPlayerWin = src.UPDetail.PlayerWin
		dtl.UpStrategyId = src.UPDetail.StrategyId
		dtl.UpIsSuppress = src.UPDetail.IsSuppress
		dtl.UpCrit7up = src.UPDetail.Crit7up
		dtl.UpCrited = src.UPDetail.Crited
		dtl.UpCritOdds = src.UPDetail.CritOdds

		dtl.UserId = user.Userid
		dtl.UpUserid = user.Userid
		dtl.UpDragon = user.Dragon
		dtl.UpTiger = user.Tiger
		dtl.UpTie = user.Tie
		dtl.UpObserve = user.Observe
		dtl.UpResult = user.Result
		dtl.UpWin = user.Win
		dtl.UpSeatBets = user.SeatBets
		dtl.UpCrit = user.Crit
		dtl.UpBeforeScore = user.BeforeScore
		dtl.UpAfterScore = user.AfterScore
		dtl.UpBeforeCash = user.BeforeCash
		dtl.UpAfterCash = user.AfterCash
		dtl.UpBeforeBonus = user.BeforeBonus
		dtl.UpAfterBonus = user.AfterBonus
		dtl.UpCashMingTax = user.CashMingTax
		dtl.UpBonusMingTax = user.BonusMingTax
		dtl.UpCashAnTax = user.CashAnTax
		dtl.UpBonusAnTax = user.BonusAnTax
		dtl.UpBeforeBackRate = user.BeforeBackRate
		dtl.UpAfterBackRate = user.AfterBackRate

		// 计算属性
		if user.Observe {
			dtl.WinType = WinTypeObserver
		} else {
			switch user.Result {
			case "赢":
				dtl.WinType = WinTypeWin
			case "输":
				dtl.WinType = WinTypeLose
			case "平":
				dtl.WinType = WinTypeTie
			default:
				if user.Win > 0 {
					dtl.WinType = WinTypeWin
				} else if user.Win < 0 {
					dtl.WinType = WinTypeLose
				} else {
					dtl.WinType = WinTypeTie
				}
			}
		}

		dtl.Robot = IsRobot(dtl.UserId)
		if src.UPDetail.Crit7up {
			for _, bets := range user.SeatBets {
				dtl.BetAmount += bets
			}
		} else {
			dtl.BetAmount = user.Dragon + user.Tiger + user.Tie
		}
		dtl.SettleScore = user.Win
		dtl.Score = user.Win - user.CashMingTax
		dtl.ScoreUntax = user.Win
		dtl.BeforeScore = user.BeforeScore
		dtl.AfterScore = user.AfterScore
		dtl.CashMingTax = user.CashMingTax
		dtl.BonusMingTax = user.BonusMingTax
		dtl.CashAnTax = user.CashAnTax
		dtl.BonusAnTax = user.BonusAnTax
		dtl.PlayerNum = int32(len(src.UPDetail.UserDetail))
		dtl.PlayerUserNum = dtl.PlayerNum
		details = append(details, dtl)
	}
	return
}

func (c *Detail) parseCrash(src *data.Detail, dst *Detail) (details []*Detail, err error) {
	if src.CRASHDetail == nil {
		details = append(details, dst)
		return
	}
	bytes, err := sonic.Marshal(dst) // 基本属性copy
	for _, user := range src.CRASHDetail.UserDetail {
		dtl := new(Detail)
		if err = sonic.Unmarshal(bytes, dtl); err != nil {
			return
		}
		dtl.CrashWinResult = src.CRASHDetail.Result
		var mulpitle float64
		if _, err = fmt.Sscanf(src.CRASHDetail.Result, "x%f", &mulpitle); err == nil {
			dtl.CrashWinResult2 = int32(mulpitle * 100)
		}
		dtl.CrashBets = src.CRASHDetail.Bets
		dtl.CrashPlayerLose = src.CRASHDetail.PlayerLose
		for _, st := range src.CRASHDetail.StrategyType {
			dtl.CrashStrategyType = append(dtl.CrashStrategyType, int32(st))
		}
		dtl.CrashRkyhRfge = src.CRASHDetail.RkyhRfge

		dtl.UserId = user.Userid
		dtl.Nickname = user.Nickname
		dtl.Photo = user.Photo
		dtl.VipLv = user.VipLv
		dtl.CrashUserid = user.Userid
		dtl.CrashObserve = user.Observe
		dtl.CrashBet = user.Bet
		dtl.CrashBet0 = user.Bet0
		dtl.CrashBet1 = user.Bet1
		dtl.CrashMultiple = user.Multiple
		dtl.CrashMultiple1 = user.Multiple1
		if user.Multiple != "" {
			var mulpitle float64
			if _, err = fmt.Sscanf(user.Multiple, "%f", &mulpitle); err == nil {
				dtl.CrashMultiple2 = int32(mulpitle * 100)
			}
		}
		if user.Multiple1 != "" {
			var mulpitle float64
			if _, err = fmt.Sscanf(user.Multiple1, "%f", &mulpitle); err == nil {
				dtl.CrashMultiple1_2 = int32(mulpitle * 100)
			}
		}
		dtl.CrashResult = user.Result
		dtl.CrashWin = user.Win
		dtl.CrashWin0 = user.Win0
		dtl.CrashWin1 = user.Win1
		dtl.CrashBeforeScore = user.BeforeScore
		dtl.CrashAfterScore = user.AfterScore
		dtl.CrashBeforeCash = user.BeforeCash
		dtl.CrashAfterCash = user.AfterCash
		dtl.CrashBeforeBonus = user.BeforeBonus
		dtl.CrashAfterBonus = user.AfterBonus
		dtl.CrashCashMingTax = user.CashMingTax
		dtl.CrashBonusMingTax = user.BonusMingTax
		dtl.CrashCashAnTax = user.CashAnTax
		dtl.CrashBonusAnTax = user.BonusAnTax
		dtl.CrashRobot = user.Robot
		dtl.CrashNextBet0 = user.NextBet0
		dtl.CrashNextBet1 = user.NextBet1
		dtl.CrashAutoCrash0 = user.AutoCrash0
		dtl.CrashAutoCrash1 = user.AutoCrash1
		dtl.CrashAutoCrashM0 = user.AutoCrashM0
		dtl.CrashAutoCrashM1 = user.AutoCrashM1

		// 计算属性
		if user.Observe {
			dtl.WinType = WinTypeObserver
		} else {
			switch user.Result {
			case "赢":
				dtl.WinType = WinTypeWin
			case "输":
				dtl.WinType = WinTypeLose
			case "平":
				dtl.WinType = WinTypeTie
			default:
				if user.Win > 0 {
					dtl.WinType = WinTypeWin
				} else if user.Win < 0 {
					dtl.WinType = WinTypeLose
				} else {
					dtl.WinType = WinTypeTie
				}
			}
		}

		dtl.Robot = user.Robot || IsRobot(dtl.UserId)
		dtl.BetAmount = user.Bet
		dtl.SettleScore = user.Win
		dtl.Score = user.Win
		dtl.ScoreUntax = user.Win + user.CashMingTax
		dtl.BeforeScore = user.BeforeScore
		dtl.AfterScore = user.AfterScore
		dtl.CashMingTax = user.CashMingTax
		dtl.BonusMingTax = user.BonusMingTax
		dtl.CashAnTax = user.CashAnTax
		dtl.BonusAnTax = user.BonusAnTax
		dtl.PlayerNum = int32(len(src.CRASHDetail.UserDetail))
		dtl.PlayerUserNum = dtl.PlayerNum
		details = append(details, dtl)
	}
	return
}

func (c *Detail) parseAndarBahar(src *data.Detail, dst *Detail) (details []*Detail, err error) {
	if src.ABDetail == nil {
		details = append(details, dst)
		return
	}

	bytes, err := sonic.Marshal(dst) // 基本属性copy
	for _, user := range src.ABDetail.UserDetail {
		dtl := new(Detail)
		if err = sonic.Unmarshal(bytes, dtl); err != nil {
			return
		}
		dtl.AbJoker = src.ABDetail.Joker
		dtl.AbJackpotValue = src.ABDetail.JackpotValue
		dtl.AbACards = src.ABDetail.ACards
		dtl.AbBCards = src.ABDetail.BCards
		dtl.AbWinner = src.ABDetail.Winner
		dtl.AbSideWinner = src.ABDetail.SideWinner
		dtl.AbBets = src.ABDetail.Bets
		dtl.AbPlayerWin = src.ABDetail.PlayerWin
		dtl.AbStrategyId = src.ABDetail.StrategyId
		dtl.AbIsRandom = src.ABDetail.IsRandom
		dtl.AbFinalFactor = src.ABDetail.FinalFactor
		dtl.AbCanWinScore = src.ABDetail.CanWinScore

		dtl.UserId = user.Userid
		dtl.AbUserid = user.Userid
		dtl.AbSeatBets = user.SeatBets
		dtl.AbObserve = user.Observe
		dtl.AbResult = user.Result
		dtl.AbWin = user.Win
		dtl.AbBeforeScore = user.BeforeScore
		dtl.AbAfterScore = user.AfterScore
		dtl.AbBeforeCash = user.BeforeCash
		dtl.AbAfterCash = user.AfterCash
		dtl.AbBeforeBonus = user.BeforeBonus
		dtl.AbAfterBonus = user.AfterBonus
		dtl.AbCashMingTax = user.CashMingTax
		dtl.AbBonusMingTax = user.BonusMingTax
		dtl.AbCashAnTax = user.CashAnTax
		dtl.AbBonusAnTax = user.BonusAnTax

		// 计算属性
		if user.Observe {
			dtl.WinType = WinTypeObserver
		} else {
			switch user.Result {
			case "赢":
				dtl.WinType = WinTypeWin
			case "输":
				dtl.WinType = WinTypeLose
			case "平":
				dtl.WinType = WinTypeTie
			default:
				if user.Win > 0 {
					dtl.WinType = WinTypeWin
				} else if user.Win < 0 {
					dtl.WinType = WinTypeLose
				} else {
					dtl.WinType = WinTypeTie
				}
			}
		}

		dtl.Robot = IsRobot(dtl.UserId)
		for _, bet := range user.SeatBets {
			dtl.BetAmount += bet
		}
		dtl.SettleScore = user.Win
		dtl.Score = user.Win - user.CashMingTax
		dtl.ScoreUntax = user.Win
		dtl.BeforeScore = user.BeforeScore
		dtl.AfterScore = user.AfterScore
		dtl.CashMingTax = user.CashMingTax
		dtl.BonusMingTax = user.BonusMingTax
		dtl.CashAnTax = user.CashAnTax
		dtl.BonusAnTax = user.BonusAnTax
		dtl.PlayerNum = int32(len(src.ABDetail.UserDetail))
		dtl.PlayerUserNum = dtl.PlayerNum
		details = append(details, dtl)
	}
	return
}

func (c *Detail) parseCp(src *data.Detail, dst *Detail) (details []*Detail, err error) {
	if src.CPDetail == nil {
		details = append(details, dst)
		return
	}

	bytes, err := sonic.Marshal(dst) // 基本属性copy
	for _, user := range src.CPDetail.UserDetail {
		dtl := new(Detail)
		if err = sonic.Unmarshal(bytes, dtl); err != nil {
			return
		}
		dtl.CpJackpotOutput = src.CPDetail.JackpotOutput
		dtl.CpCardType = src.CPDetail.CardType
		dtl.CpCards = src.CPDetail.Cards
		dtl.CpWinner = src.CPDetail.Winner
		dtl.CpBets = src.CPDetail.Bets
		dtl.CpPlayerWin = src.CPDetail.PlayerWin
		dtl.CpStrategyId = src.CPDetail.StrategyId
		dtl.CpIsRandom = src.CPDetail.IsRandom
		dtl.CpFinalFactor = src.CPDetail.FinalFactor
		dtl.CpCanWinScore = src.CPDetail.CanWinScore

		dtl.UserId = user.Userid
		dtl.CpObserve = user.Observe
		dtl.CpUserid = user.Userid
		dtl.CpSeatBets = user.SeatBets
		dtl.CpResult = user.Result
		dtl.CpWin = user.Win
		dtl.CpBeforeScore = user.BeforeScore
		dtl.CpAfterScore = user.AfterScore
		dtl.CpBeforeCash = user.BeforeCash
		dtl.CpAfterCash = user.AfterCash
		dtl.CpBeforeBonus = user.BeforeBonus
		dtl.CpAfterBonus = user.AfterBonus
		dtl.CpCashMingTax = user.CashMingTax
		dtl.CpBonusMingTax = user.BonusMingTax
		dtl.CpCashAnTax = user.CashAnTax
		dtl.CpBonusAnTax = user.BonusAnTax

		// 计算属性

		if user.Observe {
			dtl.WinType = WinTypeObserver
		} else {
			switch user.Result {
			case "赢":
				dtl.WinType = WinTypeWin
			case "输":
				dtl.WinType = WinTypeLose
			case "平":
				dtl.WinType = WinTypeTie
			default:
				if user.Win > 0 {
					dtl.WinType = WinTypeWin
				} else if user.Win < 0 {
					dtl.WinType = WinTypeLose
				} else {
					dtl.WinType = WinTypeTie
				}
			}
		}

		dtl.Robot = IsRobot(dtl.UserId)
		for _, bet := range user.SeatBets {
			dtl.BetAmount += bet
		}
		dtl.SettleScore = user.Win
		dtl.Score = user.Win - user.CashMingTax
		dtl.ScoreUntax = user.Win
		dtl.BeforeScore = user.BeforeScore
		dtl.AfterScore = user.AfterScore
		dtl.CashMingTax = user.CashMingTax
		dtl.BonusMingTax = user.BonusMingTax
		dtl.CashAnTax = user.CashAnTax
		dtl.BonusAnTax = user.BonusAnTax
		dtl.PlayerNum = int32(len(src.CPDetail.UserDetail))
		dtl.PlayerUserNum = dtl.PlayerNum
		details = append(details, dtl)
	}
	return
}

func (c *Detail) parseRedBlack(src *data.Detail, dst *Detail) (details []*Detail, err error) {
	if src.RBDetail == nil {
		details = append(details, dst)
		return
	}

	bytes, err := sonic.Marshal(dst) // 基本属性copy
	for _, user := range src.RBDetail.UserDetail {
		dtl := new(Detail)
		if err = sonic.Unmarshal(bytes, dtl); err != nil {
			return
		}
		dtl.RBCardType = src.RBDetail.CardType
		dtl.RBCards = src.RBDetail.Cards
		dtl.RBWinner = src.RBDetail.Winner
		dtl.RBBets = src.RBDetail.Bets
		dtl.RBPlayerWin = src.RBDetail.PlayerWin
		dtl.RBStrategyId = src.RBDetail.StrategyId
		dtl.RBCanWinScore = src.RBDetail.CanWinScore

		dtl.UserId = user.Userid
		dtl.RBUserid = user.Userid
		dtl.RBSeatBets = user.SeatBets
		dtl.RBResult = user.Result
		dtl.RBWin = user.Win
		dtl.RBObserve = user.Observe
		dtl.RBBeforeScore = user.BeforeScore
		dtl.RBAfterScore = user.AfterScore
		dtl.RBBeforeCash = user.BeforeCash
		dtl.RBAfterCash = user.AfterCash
		dtl.RBBeforeBonus = user.BeforeBonus
		dtl.RBAfterBonus = user.AfterBonus
		dtl.RBCashMingTax = user.CashMingTax
		dtl.RBBonusMingTax = user.BonusMingTax
		dtl.RBCashAnTax = user.CashAnTax
		dtl.RBBonusAnTax = user.BonusAnTax

		// 计算属性
		if user.Observe {
			dtl.WinType = WinTypeObserver
		} else {
			switch user.Result {
			case "赢":
				dtl.WinType = WinTypeWin
			case "输":
				dtl.WinType = WinTypeLose
			case "平":
				dtl.WinType = WinTypeTie
			default:
				if user.Win > 0 {
					dtl.WinType = WinTypeWin
				} else if user.Win < 0 {
					dtl.WinType = WinTypeLose
				} else {
					dtl.WinType = WinTypeTie
				}
			}
		}

		dtl.Robot = IsRobot(dtl.UserId)
		for _, bet := range user.SeatBets {
			dtl.BetAmount += bet
		}
		dtl.SettleScore = user.Win
		dtl.Score = user.Win - user.CashMingTax
		dtl.ScoreUntax = user.Win
		dtl.BeforeScore = user.BeforeScore
		dtl.AfterScore = user.AfterScore
		dtl.CashMingTax = user.CashMingTax
		dtl.BonusMingTax = user.BonusMingTax
		dtl.CashAnTax = user.CashAnTax
		dtl.BonusAnTax = user.BonusAnTax
		dtl.PlayerNum = int32(len(src.RBDetail.UserDetail))
		dtl.PlayerUserNum = dtl.PlayerNum
		details = append(details, dtl)
	}
	return
}

func (c *Detail) parseMines(src *data.Detail, dst *Detail) (details []*Detail, err error) {
	if src.MinesDetail == nil {
		details = append(details, dst)
		return
	}

	dst.UserId = src.MinesDetail.UserId
	dst.MinesUserId = src.MinesDetail.UserId
	dst.MinesScore = src.MinesDetail.Score
	dst.MinesBeforeScore = src.MinesDetail.BeforeScore
	dst.MinesAfterScore = src.MinesDetail.AfterScore
	dst.MinesBeforeCash = src.MinesDetail.BeforeCash
	dst.MinesAfterCash = src.MinesDetail.AfterCash
	dst.MinesBeforeBonus = src.MinesDetail.BeforeBonus
	dst.MinesAfterBonus = src.MinesDetail.AfterBonus
	dst.MinesMines = src.MinesDetail.Mines
	dst.MinesMinesPits = src.MinesDetail.MinesPits
	dst.MinesStep = src.MinesDetail.Step
	dst.MinesStepPits = src.MinesDetail.StepPits
	dst.MinesMultiple = src.MinesDetail.Multiple
	dst.MinesAutoMines = src.MinesDetail.AutoMines
	dst.MinesBets = src.MinesDetail.Bets
	dst.MinesWinType = src.MinesDetail.WinType
	dst.MinesForce = src.MinesDetail.Force
	dst.MinesForceCashOut = src.MinesDetail.ForceCashOut

	// 计算属性
	user := src.MinesDetail
	dst.WinType = int8(user.WinType)
	dst.Robot = false
	dst.BetAmount = user.Bets

	dst.SettleScore = user.Score
	dst.Score = user.Score
	dst.ScoreUntax = user.Score
	if user.Score > 0 {
		dst.Score = user.Score - user.Bets
		dst.ScoreUntax = user.Score - user.Bets
	}
	dst.BeforeScore = user.BeforeScore
	dst.AfterScore = user.AfterScore
	dst.CashMingTax = 0
	dst.BonusMingTax = 0
	dst.CashAnTax = 0
	dst.BonusAnTax = 0
	dst.PlayerNum = 1
	dst.PlayerUserNum = 1

	details = append(details, dst)
	return
}

func (c *Detail) parseFortuneGems2(src *data.Detail, dst *Detail) (details []*Detail, err error) {
	if src.FortuneGems2Detail == nil {
		details = append(details, dst)
		return
	}

	dst.UserId = src.FortuneGems2Detail.UserId
	dst.FortuneGems2UserId = src.FortuneGems2Detail.UserId
	dst.FortuneGems2BeforeScore = src.FortuneGems2Detail.BeforeScore
	dst.FortuneGems2AfterScore = src.FortuneGems2Detail.AfterScore
	dst.FortuneGems2BeforeCash = src.FortuneGems2Detail.BeforeCash
	dst.FortuneGems2AfterCash = src.FortuneGems2Detail.AfterCash
	dst.FortuneGems2Bet = src.FortuneGems2Detail.Bet
	dst.FortuneGems2ExtraBet = src.FortuneGems2Detail.ExtraBet
	dst.FortuneGems2Score = src.FortuneGems2Detail.Score
	dst.FortuneGems2Reels = make(map[int32]int32)
	for k, v := range src.FortuneGems2Detail.Reels {
		dst.FortuneGems2Reels[int32(k)] = int32(v)
	}
	dst.FortuneGems2SpecificSymbol = int32(src.FortuneGems2Detail.SpecificSymbol)
	dst.FortuneGems2WinLines = src.FortuneGems2Detail.WinLines
	dst.FortuneGems2WinMultiplier = src.FortuneGems2Detail.WinMultiplier
	dst.FortuneGems2NormalWin = src.FortuneGems2Detail.NormalWin
	dst.FortuneGems2WheelMultiplier = src.FortuneGems2Detail.WheelMultiplier
	dst.FortuneGems2ExtraMultipliers = make(map[int32]int32)
	for k, v := range src.FortuneGems2Detail.ExtraMultipliers {
		dst.FortuneGems2ExtraMultipliers[int32(k)] = int32(v)
	}
	dst.FortuneGems2WheelWin = src.FortuneGems2Detail.WheelWin
	dst.FortuneGems2TotalWin = src.FortuneGems2Detail.TotalWin

	//计算属性
	user := src.FortuneGems2Detail
	delta := user.TotalWin - user.Score
	if delta > 0 {
		dst.WinType = WinTypeWin
	} else if delta < 0 {
		dst.WinType = WinTypeLose
	} else {
		dst.WinType = WinTypeTie
	}
	dst.Robot = false
	dst.BetAmount = user.Score
	dst.SettleScore = delta
	dst.Score = delta
	dst.ScoreUntax = delta
	dst.BeforeScore = user.BeforeScore
	dst.AfterScore = user.AfterScore
	dst.CashMingTax = 0
	dst.BonusMingTax = 0
	dst.CashAnTax = 0
	dst.BonusAnTax = 0
	dst.PlayerNum = 1
	dst.PlayerUserNum = 1

	details = append(details, dst)
	return
}

func (c *Detail) parseFortuneGems(src *data.Detail, dst *Detail) (details []*Detail, err error) {
	if src.FortuneGemsDetail == nil {
		details = append(details, dst)
		return
	}

	dst.UserId = src.FortuneGemsDetail.UserId
	dst.FortuneGemsUserId = src.FortuneGemsDetail.UserId
	dst.FortuneGemsBeforeScore = src.FortuneGemsDetail.BeforeScore
	dst.FortuneGemsAfterScore = src.FortuneGemsDetail.AfterScore
	dst.FortuneGemsBeforeCash = src.FortuneGemsDetail.BeforeCash
	dst.FortuneGemsAfterCash = src.FortuneGemsDetail.AfterCash
	dst.FortuneGemsBet = src.FortuneGemsDetail.Bet
	dst.FortuneGemsExtraBet = src.FortuneGemsDetail.ExtraBet
	dst.FortuneGemsScore = src.FortuneGemsDetail.Score
	dst.FortuneGemsReels = make(map[int32]int32)
	for k, v := range src.FortuneGemsDetail.Reels {
		dst.FortuneGemsReels[int32(k)] = int32(v)
	}
	dst.FortuneGemsSpecificSymbol = int32(src.FortuneGemsDetail.SpecificSymbol)
	dst.FortuneGemsWinLines = src.FortuneGemsDetail.WinLines
	dst.FortuneGemsWinMultiplier = src.FortuneGemsDetail.WinMultiplier
	dst.FortuneGemsNormalWin = src.FortuneGemsDetail.NormalWin
	dst.FortuneGemsTotalWin = src.FortuneGemsDetail.TotalWin

	//计算属性
	user := src.FortuneGemsDetail
	delta := user.TotalWin - user.Score
	if delta > 0 {
		dst.WinType = WinTypeWin
	} else if delta < 0 {
		dst.WinType = WinTypeLose
	} else {
		dst.WinType = WinTypeTie
	}
	dst.Robot = false
	dst.BetAmount = user.Score
	dst.SettleScore = delta
	dst.Score = delta
	dst.ScoreUntax = delta
	dst.BeforeScore = user.BeforeScore
	dst.AfterScore = user.AfterScore
	dst.CashMingTax = 0
	dst.BonusMingTax = 0
	dst.CashAnTax = 0
	dst.BonusAnTax = 0
	dst.PlayerNum = 1
	dst.PlayerUserNum = 1

	details = append(details, dst)
	return
}
