package data

import (
	"goserver/gen/pb"
	"time"

	"github.com/globalsign/mgo/bson"
)

//TODO 添加房间id,人数同步后台显示

// 输赢记录
type FreeWin struct {
	Wtype int32 `bson:"wtype" json:"wtype"` // 类型
	Score int64 `bson:"score" json:"score"` // 输赢分
}

type WinLoseRound struct {
	WinRound  int64 `bson:"win_round" json:"win_round"`  // 赢的回合
	LoseRound int64 `json:"loseRound" bson:"lose_round"` // 输的回合
}

type GameGuid struct {
	Gtype  int32   `json:"gtype" bson:"gtype"`    //游戏类型
	Ftype  int32   `json:"ftype" bson:"ftype"`    //功能类型
	GuidId []int32 `json:"guidId" bson:"guid_id"` //引导id
}

func (g *GameGuid) BuildGuid() *pb.GameGuid {
	bean := new(pb.GameGuid)
	bean.Gtype = g.Gtype
	bean.Ftype = g.Ftype
	for _, v := range g.GuidId {
		bean.GuidId = append(bean.GuidId, pb.GuidType(v))
	}
	return bean
}

// 房间人数配置
type RoomPeople struct {
	ID              int32   `json:"id" bson:"id"`                            // id
	Gtype           int32   `json:"gtype" bson:"gtype"`                      // 游戏类型
	RoomId          string  `json:"roomId" bson:"room_id"`                   // 房间id
	PeopleInterval1 []int32 `json:"peopleInterval1" bson:"people_interval1"` // 闲时上下限
	PeopleInterval2 []int32 `json:"peopleInterval2" bson:"people_interval2"` // 正常上下限
	PeopleInterval3 []int32 `json:"peopleInterval3" bson:"people_interval3"` // 忙时上下限
}

type Game struct {
	Unique   string `json:"unique"`
	DeskType int32  `json:"desk_type"`
	// Game 游戏金币房间配置
	Id           string `bson:"_id" json:"_id"`                   //房间ID
	Name         string `bson:"name" json:"name"`                 //房间名称
	Gtype        int32  `bson:"gtype" json:"gtype"`               //游戏类型 1：TP; 2：LHD; 3：SEVEN; 4:RUMMY; 5:AK47; 6:JOKER; 7:CRASH
	Status       int    `bson:"status" json:"status"`             //房间开关 0: 关; 1: 开;
	Ai_Status    int    `bson:"ai_status" json:"ai_status"`       //房间AI开关 0: 关; 1: 开;
	Algo_Switch  int    `bson:"algo_switch" json:"algo_switch"`   //算法（1=old,2=new）
	Kick_Score   int    `json:"kick_score" bson:"kick_score"`     //踢人分数
	Min_Access   int    `json:"min_access" bson:"min_access"`     //最低准入
	Max_Access   int    `json:"max_access" bson:"max_access"`     //最高准入
	Stock_Expect int64  `json:"stock_expect" bson:"stock_expect"` //库存期望
	Stock_Alarm  int    `json:"stock_alarm" bson:"stock_alarm"`   //库存报警
	SortId       int    `json:"sortId" bson:"sort_id"`            //房间分组内排序标识
	Count        uint32 `bson:"count" json:"count"`               //房间人数上限
	//ID bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	// Id     string    `bson:"_id" json:"id"`        //unique ID
	// Gtype  int32     `bson:"gtype" json:"gtype"`   //游戏类型1 niu,2 san,3 jiu
	Rtype int32 `bson:"rtype" json:"rtype"` //房间类型0免佣,1抽佣
	Dtype int32 `bson:"dtype" json:"dtype"` //桌子类型
	Ltype int32 `bson:"ltype" json:"ltype"` //彩票类型1bjpk10,1mlaft
	// Name   string    `bson:"name" json:"name"`     //房间名称
	// Status uint32    `bson:"status" json:"status"` //房间状态1打开,2关闭,3隐藏
	// Count  uint32    `bson:"count" json:"count"`   //房间限制人数
	Ante  uint32    `bson:"ante" json:"ante"`   //房间底分
	Cost  uint32    `bson:"cost" json:"cost"`   //房间抽佣百分比
	Vip   uint32    `bson:"vip" json:"vip"`     //房间vip限制
	Chip  uint32    `bson:"chip" json:"chip"`   //房间进入筹码限制
	Deal  bool      `bson:"deal" json:"deal"`   //房间是否可以上庄
	Carry uint32    `bson:"carry" json:"carry"` //房间上庄最小携带筹码限制
	Down  uint32    `bson:"down" json:"down"`   //房间下庄最小携带筹码限制
	Top   uint32    `bson:"top" json:"Top"`     //房间下庄最大携带筹码限制
	Sit   uint32    `bson:"sit" json:"sit"`     //房间内坐下限制
	Del   int       `bson:"del" json:"del"`     //是否移除
	Node  string    `bson:"node" json:"node"`   //所在节点(game.huiyin1|game.huiyin2)
	Ctime time.Time `bson:"ctime" json:"ctime"` //创建时间
	//Num   uint32    `bson:"num" json:"num"`      //启动房间数量
	Minimum       int64     `bson:"minimum" json:"minimum"`                 //房间最低限制
	Maximum       int64     `bson:"maximum" json:"maximum"`                 //房间最高限制
	Pub           bool      `bson:"pub" json:"pub"`                         //公开展示
	Mode          uint32    `bson:"mode" json:"mode"`                       //模式，0普通，1疯狂
	Multiple      uint32    `bson:"multiple" json:"multiple"`               //倍数，0低，1中，2高
	BreakingPrice int       `bson:"break_price" json:"break_price"`         //破产礼包价格
	BreakingGive  int       `bson:"break_give" json:"break_give"`           //破产礼包赠送
	BreakingTimes int       `bson:"break_times" json:"break_times"`         //破产礼包每日次数
	RoomType      int       `bson:"room_type" json:"room_type"`             // 房间类型(0匹配,1对战房,2对战房娱乐模式)
	MinFirstEntry int       `bson:"min_first_entry" json:"min_first_entry"` // 初始携带要求
	GoldMingTax   int       `bson:"gold_ming_tax" json:"gold_ming_tax"`     // 对战真金模式明税（万分比）
	TP            TPGame    //tp
	JOKER         JOKERGame //joker
	AK47          AK47Game  //ak47
	LHD           LHDGame   //lhd
	RM            RMGame    //rm
	UP            LHDGame   //7up
	CRASH         CRASHGame //crash
	AB            ABGame    //andarbahar
	LOTTERY       CPGame    //彩票
	PLANE         PLANEGame //飞机
}

type GameList []Game

func (gl GameList) Len() int {
	return len(gl)
}

func (gl GameList) Less(i, j int) bool {
	return gl[i].Id < gl[j].Id
}

func (gl GameList) Swap(i, j int) {
	gl[i], gl[j] = gl[j], gl[i]
}

func (g Game) GetBottom() int64 {
	switch g.Gtype {
	case int32(pb.HUA), int32(pb.HUA2):
		return int64(g.TP.Bottom)
	case int32(pb.JOKER):
		return int64(g.JOKER.Bottom)
	case int32(pb.AK47):
		return int64(g.AK47.Bottom)
	case int32(pb.RUMMY), int32(pb.RUMMY2):
		return int64(g.RM.Bottom)
	default:
		return 0
	}
}

func (r *RoomPeople) Save() {
	Upsert(RoomPeoples, bson.M{"_id": r.ID}, r)
}

type TPGame struct {
	Bottom       int       `json:"bottom" bson:"bottom"`             //底注
	Rounds       int       `json:"rounds" bson:"rounds"`             //轮次上限
	Than_Rounds  int       `json:"than_rounds" bson:"than_rounds"`   //比牌轮次
	Pool_Limit   int       `json:"pool_limit" bson:"pool_limit"`     //筹码池上限
	Otime        int       `json:"otime" bson:"otime"`               //操作时间
	Tcountdown   int       `json:"tcountdown" bson:"tcountdown"`     //比牌倒计时
	Scountdown   int       `json:"scountdown" bson:"scountdown"`     //结算倒数时间
	Match_Time   []int     `json:"match_time" bson:"match_time"`     //匹配时间区间
	Single_Robot []int     `json:"single_robot" bson:"single_robot"` //单局人机数区间
	Robot_Join   []int     `json:"robot_join" bson:"robot_join"`     //人机加入概率
	Robot_Leave  int       `json:"robot_leave" bson:"robot_leave"`   //人机离开概率(万分比)
	Prevent_Time int       `json:"prevent_time" bson:"prevent_time"` //防作弊匹配时间检测
	Prevent_Num  int       `json:"prevent_num" bson:"prevent_num"`   //防作弊匹配次数检测
	Prevent_Thaw int       `json:"prevent_thaw" bson:"prevent_thaw"` //防作弊匹配解冻时间
	Msg_Score    int       `json:"msg_score" bson:"msg_score"`       //跑马灯显示分
	Ctime        time.Time `bson:"ctime"`                            //修改时间
	// NewbieType         int                        `json:"newbie_type" bson:"newbie_type"`               //新手状态使用牌型
	// AbnormalType       int                        `json:"abnormal_type" bson:"abnormal_type"`           //异常状态使用牌型
	FinalFactorRange   []int                      `json:"final_factor_range" bson:"final_factor_range"` //正常状态和点控最终系数区间
	CardTypeRange      []int                      `json:"card_type_range" bson:"card_type_range"`       //正常状态和点控系数对应使用牌型
	MingTax            int32                      `json:"ming_tax" bson:"ming_tax"`                     //明税
	AnTax              int32                      `json:"an_tax" bson:"an_tax"`                         //暗税
	RoomRecharge       []int32                    `json:"room_recharge" bson:"room_recharge"`           //局内充值
	RoomRechargeGive   []int32                    `json:"room_recharge_give" bson:"room_recharge_give"` //局内充值赠送
	AnteFactor         int                        `json:"ante_factor" bson:"ante_factor"`               //底注系数
	ShowSwitch         int                        `json:"show_switch" bson:"show_switch"`               //结算摊牌开关
	Strategy100Switch  int                        `json:"strategy100_switch" bson:"strategy100_switch"` //策略100开关
	Strategy200Switch  int                        `json:"strategy200_switch" bson:"strategy200_switch"` //策略200开关
	Strategy300Switch  int                        `json:"strategy300_switch" bson:"strategy300_switch"` //策略300开关
	Strategy100Round   int                        `json:"strategy100_round" bson:"strategy100_round"`   //策略100累计局数
	Strategy200Round   int                        `json:"strategy200_round" bson:"strategy200_round"`   //策略200累计局数
	Strategy300Round   int                        `json:"strategy300_round" bson:"strategy300_round"`   //策略300累计局数
	RoomFactorSwitch   int                        `json:"room_factor_switch" bson:"room_factor_switch"` //房间系数开关
	ACardType          int                        `json:"a_card_type" bson:"a_card_type"`               //A类指定牌型
	CardType           []TPGameCardType           //发牌牌型配置
	RobotStrategyGroup []TPGameRobotStrategyGroup //人机策略组配置
	RobotStrategy      []TPGameRobotStrategy      //人机策略配置
	NewbieMode         TPNewbieMode               //B类新手模式
	ANewbieMode        TPNewbieMode               //A类新手模式
	ControlType        []TPControlType            //控制方式
}

func (tp TPGame) FindCardType(id int) TPGameCardType {
	for _, v := range tp.CardType {
		if v.Id == id {
			return v
		}
	}
	return TPGameCardType{}
}

func (tp TPGame) FindRobotStrategyGroup(id int) TPGameRobotStrategyGroup {
	for _, v := range tp.RobotStrategyGroup {
		if v.Id == id {
			return v
		}
	}
	return TPGameRobotStrategyGroup{}
}

func (tp TPGame) FindRobotStrategy(id int) TPGameRobotStrategy {
	for _, v := range tp.RobotStrategy {
		if int(v.Id) == id {
			return v
		}
	}
	return TPGameRobotStrategy{}
}

type TPGameCardType struct {
	Id                 int   `json:"id" bson:"id"`                                     //牌型ID
	PlayerWeight       []int `json:"player_weight" bson:"player_weight"`               //玩家权重
	RobotWeight        []int `json:"robot_weight" bson:"robot_weight"`                 //人机权重
	WinRateCheck       int   `json:"win_rate_check" bson:"win_rate_check"`             //胜率检测
	PlayerWinRate      int   `json:"player_win_rate" bson:"player_win_rate"`           //玩家获胜概率
	RobotStrategyGroup int   `json:"robot_strategy_group" bson:"robot_strategy_group"` //人机策略组
}

type TPGameRobotStrategyGroup struct {
	Id                 int `json:"id" bson:"id"` //策略组ID
	BaoZiBigger        int `json:"bao_zi_bigger" bson:"bao_zi_bigger"`
	TongHuaShunBigger  int `json:"tong_hua_shun_bigger" bson:"tong_hua_shun_bigger"`
	ShunZiBigger       int `json:"shun_zi_bigger" bson:"shun_zi_bigger"`
	TongHuaBigger      int `json:"tong_hua_bigger" bson:"tong_hua_bigger"`
	DuiZiBigger        int `json:"dui_zi_bigger" bson:"dui_zi_bigger"`
	GaoPaiBigger       int `json:"gao_pai_bigger" bson:"gao_pai_bigger"`
	BaoZiSmaller       int `json:"bao_zi_smaller" bson:"bao_zi_smaller"`
	TongHuaShunSmaller int `json:"tong_hua_shun_smaller" bson:"tong_hua_shun_smaller"`
	ShunZiSmaller      int `json:"shun_zi_smaller" bson:"shun_zi_smaller"`
	TongHuaSmaller     int `json:"tong_hua_smaller" bson:"tong_hua_smaller"`
	DuiZiSmaller       int `json:"dui_zi_smaller" bson:"dui_zi_smaller"`
	GaoPaiSmaller      int `json:"gao_pai_smaller" bson:"gao_pai_smaller"`
}

type TPGameRobotStrategy struct {
	Id           int32     `json:"id" bson:"id"`                       //人机策略ID
	ActionWeight [][]int32 `json:"action_weight" bson:"action_weight"` //操作权重
	ActionTime   []int32   `json:"action_time" bson:"action_time"`     //操作时间
	SeeWeight    []int32   `json:"see_weight" bson:"see_weight"`       //看牌权重
	SeeTime      [][]int32 `json:"see_time" bson:"see_time"`           //看牌时机
	AgreeBi      int32     `json:"agree_bi" bson:"agree_bi"`           //被比牌同意概率
}

type TPNewbieMode struct {
	CanWithdrawRange              []int `json:"can_withdraw_range" bson:"can_withdraw_range"`
	WinRate                       []int `json:"win_rate" bson:"win_rate"`
	CanWithdrawLimit              int   `json:"can_withdraw_limit" bson:"can_withdraw_limit"`
	WinWeight                     []int `json:"win_weight" bson:"win_weight"`
	WinCardType                   []int `json:"win_card_type" bson:"win_card_type"`
	LoseWeight                    []int `json:"lose_weight" bson:"lose_weight"`
	LoseCardType                  []int `json:"lose_card_type" bson:"lose_card_type"`
	CivilianCanWithdrawRange      []int `json:"civilian_can_withdraw_range" bson:"civilian_can_withdraw_range"`
	CivilianWinRate               []int `json:"civilian_win_rate" bson:"civilian_win_rate"`
	CivilianCanWithdrawLimit      int   `json:"civilian_can_withdraw_limit" bson:"civilian_can_withdraw_limit"`
	CivilianWinWeight             []int `json:"civilian_win_weight" bson:"civilian_win_weight"`
	CivilianWinCardType           []int `json:"civilian_win_card_type" bson:"civilian_win_card_type"`
	CivilianLoseWeight            []int `json:"civilian_lose_weight" bson:"civilian_lose_weight"`
	CivilianLoseCardType          []int `json:"civilian_lose_card_type" bson:"civilian_lose_card_type"`
	CivilianTriggerStrategyProb   int   `json:"civilian_trigger_strategy_prob" bson:"civilian_trigger_strategy_prob"`
	SpecialRound                  []int `json:"special_round" bson:"special_round"`
	SpecialRoundCardType          []int `json:"special_round_card_type" bson:"special_round_card_type"`
	NewbieToCivilianWithdrawLimit int   `json:"newbie_to_civilian_withdraw_limit" bson:"newbie_to_civilian_withdraw_limit"`
	FoamCheckTotalDepositLimit    int   `json:"foam_check_total_deposit_limit" bson:"foam_check_total_deposit_limit"`
	FoamGiftRate                  int   `json:"foam_gift_rate" bson:"foam_gift_rate"`
	FoamCardType                  int   `json:"foam_card_type" bson:"foam_card_type"`
	FoamTriggerStrategyProb       int   `json:"foam_trigger_strategy_prob" bson:"foam_trigger_strategy_prob"`
}

type TPControlType struct {
	ChargeRange        []int `json:"charge_range" bson:"charge_range"`
	ControlType        int   `json:"control_type" bson:"control_type"`
	WinScoreConfig     []int `json:"win_score_config" bson:"win_score_config"`
	WinScoreWeight     []int `json:"win_score_weight" bson:"win_score_weight"`
	PlayerFactorRange  []int `json:"player_factor_range" bson:"player_factor_range"`
	PlayerFactorWeight []int `json:"player_factor_weight" bson:"player_factor_weight"`
	WinScoreLimit      int   `json:"win_score_limit" bson:"win_score_limit"`
	Switch             int   `json:"switch" bson:"switch"`
}

type JOKERGame struct {
	Bottom             int                           `json:"bottom" bson:"bottom"`                         //底注
	Rounds             int                           `json:"rounds" bson:"rounds"`                         //轮次上限
	Than_Rounds        int                           `json:"than_rounds" bson:"than_rounds"`               //比牌轮次
	Pool_Limit         int                           `json:"pool_limit" bson:"pool_limit"`                 //筹码池上限
	Otime              int                           `json:"otime" bson:"otime"`                           //操作时间
	Tcountdown         int                           `json:"tcountdown" bson:"tcountdown"`                 //比牌倒计时
	Scountdown         int                           `json:"scountdown" bson:"scountdown"`                 //结算倒数时间
	Match_Time         []int                         `json:"match_time" bson:"match_time"`                 //匹配时间区间
	Single_Robot       []int                         `json:"single_robot" bson:"single_robot"`             //单局人机数区间
	Robot_Join         []int                         `json:"robot_join" bson:"robot_join"`                 //人机加入概率
	Robot_Leave        int                           `json:"robot_leave" bson:"robot_leave"`               //人机离开概率(万分比)
	Prevent_Time       int                           `json:"prevent_time" bson:"prevent_time"`             //防作弊匹配时间检测
	Prevent_Num        int                           `json:"prevent_num" bson:"prevent_num"`               //防作弊匹配次数检测
	Prevent_Thaw       int                           `json:"prevent_thaw" bson:"prevent_thaw"`             //防作弊匹配解冻时间
	Msg_Score          int                           `json:"msg_score" bson:"msg_score"`                   //跑马灯显示分
	Ctime              time.Time                     `bson:"ctime"`                                        //修改时间
	NewbieType         int                           `json:"newbie_type" bson:"newbie_type"`               //新手状态使用牌型
	AbnormalType       int                           `json:"abnormal_type" bson:"abnormal_type"`           //异常状态使用牌型
	FinalFactorRange   []int                         `json:"final_factor_range" bson:"final_factor_range"` //正常状态和点控最终系数区间
	CardTypeRange      []int                         `json:"card_type_range" bson:"card_type_range"`       //正常状态和点控系数对应使用牌型
	MingTax            int32                         `json:"ming_tax" bson:"ming_tax"`                     //明税
	AnTax              int32                         `json:"an_tax" bson:"an_tax"`                         //暗税
	RoomRecharge       []int32                       `json:"room_recharge" bson:"room_recharge"`           //局内充值
	RoomRechargeGive   []int32                       `json:"room_recharge_give" bson:"room_recharge_give"` //局内充值赠送
	AnteFactor         int                           `json:"ante_factor" bson:"ante_factor"`               //底注系数
	ShowSwitch         int                           `json:"show_switch" bson:"show_switch"`               //结算摊牌开关
	Strategy100Switch  int                           `json:"strategy100_switch" bson:"strategy100_switch"` //策略100开关
	Strategy200Switch  int                           `json:"strategy200_switch" bson:"strategy200_switch"` //策略200开关
	Strategy300Switch  int                           `json:"strategy300_switch" bson:"strategy300_switch"` //策略300开关
	Strategy100Round   int                           `json:"strategy100_round" bson:"strategy100_round"`   //策略100累计局数
	Strategy200Round   int                           `json:"strategy200_round" bson:"strategy200_round"`   //策略200累计局数
	Strategy300Round   int                           `json:"strategy300_round" bson:"strategy300_round"`   //策略300累计局数
	RoomFactorSwitch   int                           `json:"room_factor_switch" bson:"room_factor_switch"` //房间系数开关
	ACardType          int                           `json:"a_card_type" bson:"a_card_type"`               //A类指定牌型
	CardType           []JOKERGameCardType           //发牌牌型配置
	RobotStrategyGroup []JOKERGameRobotStrategyGroup //人机策略组配置
	RobotStrategy      []JOKERGameRobotStrategy      //人机策略配置
	NewbieMode         JOKERNewbieMode               //新手模式
	ANewbieMode        JOKERNewbieMode               //新手模式(A类)
	ControlType        []JOKERControlType            //控制方式
}

func (joker JOKERGame) FindCardType(id int) JOKERGameCardType {
	for _, v := range joker.CardType {
		if v.Id == id {
			return v
		}
	}
	return JOKERGameCardType{}
}

func (joker JOKERGame) FindRobotStrategyGroup(id int) JOKERGameRobotStrategyGroup {
	for _, v := range joker.RobotStrategyGroup {
		if v.Id == id {
			return v
		}
	}
	return JOKERGameRobotStrategyGroup{}
}

func (joker JOKERGame) FindRobotStrategy(id int) JOKERGameRobotStrategy {
	for _, v := range joker.RobotStrategy {
		if int(v.Id) == id {
			return v
		}
	}
	return JOKERGameRobotStrategy{}
}

type JOKERGameCardType struct {
	Id                 int   `json:"id" bson:"id"`                                     //牌型ID
	PlayerWeight       []int `json:"player_weight" bson:"player_weight"`               //玩家权重
	RobotWeight        []int `json:"robot_weight" bson:"robot_weight"`                 //人机权重
	WinRateCheck       int   `json:"win_rate_check" bson:"win_rate_check"`             //胜率检测
	PlayerWinRate      int   `json:"player_win_rate" bson:"player_win_rate"`           //玩家获胜概率
	RobotStrategyGroup int   `json:"robot_strategy_group" bson:"robot_strategy_group"` //人机策略组
}

type JOKERGameRobotStrategyGroup struct {
	Id                 int `json:"id" bson:"id"` //策略组ID
	BaoZiBigger        int `json:"bao_zi_bigger" bson:"bao_zi_bigger"`
	TongHuaShunBigger  int `json:"tong_hua_shun_bigger" bson:"tong_hua_shun_bigger"`
	ShunZiBigger       int `json:"shun_zi_bigger" bson:"shun_zi_bigger"`
	TongHuaBigger      int `json:"tong_hua_bigger" bson:"tong_hua_bigger"`
	DuiZiBigger        int `json:"dui_zi_bigger" bson:"dui_zi_bigger"`
	BaoZiSmaller       int `json:"bao_zi_smaller" bson:"bao_zi_smaller"`
	TongHuaShunSmaller int `json:"tong_hua_shun_smaller" bson:"tong_hua_shun_smaller"`
	ShunZiSmaller      int `json:"shun_zi_smaller" bson:"shun_zi_smaller"`
	TongHuaSmaller     int `json:"tong_hua_smaller" bson:"tong_hua_smaller"`
	DuiZiSmaller       int `json:"dui_zi_smaller" bson:"dui_zi_smaller"`
}

type JOKERGameRobotStrategy struct {
	Id           int32     `json:"id" bson:"id"`                       //人机策略ID
	ActionWeight [][]int32 `json:"action_weight" bson:"action_weight"` //操作权重
	ActionTime   []int32   `json:"action_time" bson:"action_time"`     //操作时间
	SeeWeight    []int32   `json:"see_weight" bson:"see_weight"`       //看牌权重
	SeeTime      [][]int32 `json:"see_time" bson:"see_time"`           //看牌时机
	AgreeBi      int32     `json:"agree_bi" bson:"agree_bi"`           //被比牌同意概率
}

type JOKERNewbieMode struct {
	CanWithdrawRange              []int `json:"can_withdraw_range" bson:"can_withdraw_range"`
	WinRate                       []int `json:"win_rate" bson:"win_rate"`
	CanWithdrawLimit              int   `json:"can_withdraw_limit" bson:"can_withdraw_limit"`
	WinWeight                     []int `json:"win_weight" bson:"win_weight"`
	WinCardType                   []int `json:"win_card_type" bson:"win_card_type"`
	LoseWeight                    []int `json:"lose_weight" bson:"lose_weight"`
	LoseCardType                  []int `json:"lose_card_type" bson:"lose_card_type"`
	CivilianCanWithdrawRange      []int `json:"civilian_can_withdraw_range" bson:"civilian_can_withdraw_range"`
	CivilianWinRate               []int `json:"civilian_win_rate" bson:"civilian_win_rate"`
	CivilianCanWithdrawLimit      int   `json:"civilian_can_withdraw_limit" bson:"civilian_can_withdraw_limit"`
	CivilianWinWeight             []int `json:"civilian_win_weight" bson:"civilian_win_weight"`
	CivilianWinCardType           []int `json:"civilian_win_card_type" bson:"civilian_win_card_type"`
	CivilianLoseWeight            []int `json:"civilian_lose_weight" bson:"civilian_lose_weight"`
	CivilianLoseCardType          []int `json:"civilian_lose_card_type" bson:"civilian_lose_card_type"`
	CivilianTriggerStrategyProb   int   `json:"civilian_trigger_strategy_prob" bson:"civilian_trigger_strategy_prob"`
	SpecialRound                  []int `json:"special_round" bson:"special_round"`
	SpecialRoundCardType          []int `json:"special_round_card_type" bson:"special_round_card_type"`
	NewbieToCivilianWithdrawLimit int   `json:"newbie_to_civilian_withdraw_limit" bson:"newbie_to_civilian_withdraw_limit"`
	FoamCheckTotalDepositLimit    int   `json:"foam_check_total_deposit_limit" bson:"foam_check_total_deposit_limit"`
	FoamGiftRate                  int   `json:"foam_gift_rate" bson:"foam_gift_rate"`
	FoamCardType                  int   `json:"foam_card_type" bson:"foam_card_type"`
	FoamTriggerStrategyProb       int   `json:"foam_trigger_strategy_prob" bson:"foam_trigger_strategy_prob"`
}

type JOKERControlType struct {
	ChargeRange        []int `json:"charge_range" bson:"charge_range"`
	ControlType        int   `json:"control_type" bson:"control_type"`
	WinScoreConfig     []int `json:"win_score_config" bson:"win_score_config"`
	WinScoreWeight     []int `json:"win_score_weight" bson:"win_score_weight"`
	PlayerFactorRange  []int `json:"player_factor_range" bson:"player_factor_range"`
	PlayerFactorWeight []int `json:"player_factor_weight" bson:"player_factor_weight"`
	WinScoreLimit      int   `json:"win_score_limit" bson:"win_score_limit"`
}

type AK47Game struct {
	Bottom             int                          `json:"bottom" bson:"bottom"`                         //底注
	Rounds             int                          `json:"rounds" bson:"rounds"`                         //轮次上限
	Than_Rounds        int                          `json:"than_rounds" bson:"than_rounds"`               //比牌轮次
	Pool_Limit         int                          `json:"pool_limit" bson:"pool_limit"`                 //筹码池上限
	Otime              int                          `json:"otime" bson:"otime"`                           //操作时间
	Tcountdown         int                          `json:"tcountdown" bson:"tcountdown"`                 //比牌倒计时
	Scountdown         int                          `json:"scountdown" bson:"scountdown"`                 //结算倒数时间
	Match_Time         []int                        `json:"match_time" bson:"match_time"`                 //匹配时间区间
	Single_Robot       []int                        `json:"single_robot" bson:"single_robot"`             //单局人机数区间
	Robot_Join         []int                        `json:"robot_join" bson:"robot_join"`                 //人机加入概率
	Robot_Leave        int                          `json:"robot_leave" bson:"robot_leave"`               //人机离开概率(万分比)
	Prevent_Time       int                          `json:"prevent_time" bson:"prevent_time"`             //防作弊匹配时间检测
	Prevent_Num        int                          `json:"prevent_num" bson:"prevent_num"`               //防作弊匹配次数检测
	Prevent_Thaw       int                          `json:"prevent_thaw" bson:"prevent_thaw"`             //防作弊匹配解冻时间
	Msg_Score          int                          `json:"msg_score" bson:"msg_score"`                   //跑马灯显示分
	Ctime              time.Time                    `bson:"ctime"`                                        //修改时间
	NewbieType         int                          `json:"newbie_type" bson:"newbie_type"`               //新手状态使用牌型
	AbnormalType       int                          `json:"abnormal_type" bson:"abnormal_type"`           //异常状态使用牌型
	FinalFactorRange   []int                        `json:"final_factor_range" bson:"final_factor_range"` //正常状态和点控最终系数区间
	CardTypeRange      []int                        `json:"card_type_range" bson:"card_type_range"`       //正常状态和点控系数对应使用牌型
	MingTax            int32                        `json:"ming_tax" bson:"ming_tax"`                     //明税
	AnTax              int32                        `json:"an_tax" bson:"an_tax"`                         //暗税
	RoomRecharge       []int32                      `json:"room_recharge" bson:"room_recharge"`           //局内充值
	RoomRechargeGive   []int32                      `json:"room_recharge_give" bson:"room_recharge_give"` //局内充值赠送
	AnteFactor         int                          `json:"ante_factor" bson:"ante_factor"`               //底注系数
	ShowSwitch         int                          `json:"show_switch" bson:"show_switch"`               //结算摊牌开关
	Strategy100Switch  int                          `json:"strategy100_switch" bson:"strategy100_switch"` //策略100开关
	Strategy200Switch  int                          `json:"strategy200_switch" bson:"strategy200_switch"` //策略200开关
	Strategy300Switch  int                          `json:"strategy300_switch" bson:"strategy300_switch"` //策略300开关
	Strategy100Round   int                          `json:"strategy100_round" bson:"strategy100_round"`   //策略100累计局数
	Strategy200Round   int                          `json:"strategy200_round" bson:"strategy200_round"`   //策略200累计局数
	Strategy300Round   int                          `json:"strategy300_round" bson:"strategy300_round"`   //策略300累计局数
	RoomFactorSwitch   int                          `json:"room_factor_switch" bson:"room_factor_switch"` //房间系数开关
	ACardType          int                          `json:"a_card_type" bson:"a_card_type"`               //A类指定牌型
	CardType           []AK47GameCardType           //发牌牌型配置
	RobotStrategyGroup []AK47GameRobotStrategyGroup //人机策略组配置
	RobotStrategy      []AK47GameRobotStrategy      //人机策略配置
	NewbieMode         AK47NewbieMode               //新手模式
	ANewbieMode        AK47NewbieMode               //新手模式A类
	ControlType        []AK47ControlType            //控制方式
}

func (ak47 AK47Game) FindCardType(id int) AK47GameCardType {
	for _, v := range ak47.CardType {
		if v.Id == id {
			return v
		}
	}
	return AK47GameCardType{}
}

func (ak47 AK47Game) FindRobotStrategyGroup(id int) AK47GameRobotStrategyGroup {
	for _, v := range ak47.RobotStrategyGroup {
		if v.Id == id {
			return v
		}
	}
	return AK47GameRobotStrategyGroup{}
}

func (ak47 AK47Game) FindRobotStrategy(id int) AK47GameRobotStrategy {
	for _, v := range ak47.RobotStrategy {
		if int(v.Id) == id {
			return v
		}
	}
	return AK47GameRobotStrategy{}
}

type AK47GameCardType struct {
	Id                    int   `json:"id" bson:"id"` //牌型ID
	PlayerWildCardWeight  []int `json:"player_wild_card_weight" bson:"player_wild_card_weight"`
	Player0WildCardWeight []int `json:"player0_wild_card_weight" bson:"player0_wild_card_weight"`
	Player1WildCardWeight []int `json:"player1_wild_card_weight" bson:"player1_wild_card_weight"`
	RobotWildCardWeight   []int `json:"robot_wild_card_weight" bson:"robot_wild_card_weight"`
	Robot0WildCardWeight  []int `json:"robot0_wild_card_weight" bson:"robot0_wild_card_weight"`
	Robot1WildCardWeight  []int `json:"robot1_wild_card_weight" bson:"robot1_wild_card_weight"`
	WinRateCheck          int   `json:"win_rate_check" bson:"win_rate_check"`             //胜率检测
	PlayerWinRate         int   `json:"player_win_rate" bson:"player_win_rate"`           //玩家获胜概率
	RobotStrategyGroup    int   `json:"robot_strategy_group" bson:"robot_strategy_group"` //人机策略组
}

type AK47GameRobotStrategyGroup struct {
	Id                 int `json:"id" bson:"id"` //策略组ID
	BaoZiBigger        int `json:"bao_zi_bigger" bson:"bao_zi_bigger"`
	TongHuaShunBigger  int `json:"tong_hua_shun_bigger" bson:"tong_hua_shun_bigger"`
	ShunZiBigger       int `json:"shun_zi_bigger" bson:"shun_zi_bigger"`
	TongHuaBigger      int `json:"tong_hua_bigger" bson:"tong_hua_bigger"`
	DuiZiBigger        int `json:"dui_zi_bigger" bson:"dui_zi_bigger"`
	GaoPaiBigger       int `json:"gao_pai_bigger" bson:"gao_pai_bigger"`
	BaoZiSmaller       int `json:"bao_zi_smaller" bson:"bao_zi_smaller"`
	TongHuaShunSmaller int `json:"tong_hua_shun_smaller" bson:"tong_hua_shun_smaller"`
	ShunZiSmaller      int `json:"shun_zi_smaller" bson:"shun_zi_smaller"`
	TongHuaSmaller     int `json:"tong_hua_smaller" bson:"tong_hua_smaller"`
	DuiZiSmaller       int `json:"dui_zi_smaller" bson:"dui_zi_smaller"`
	GaoPaiSmaller      int `json:"gao_pai_smaller" bson:"gao_pai_smaller"`
}

type AK47GameRobotStrategy struct {
	Id           int32     `json:"id" bson:"id"`                       //人机策略ID
	ActionWeight [][]int32 `json:"action_weight" bson:"action_weight"` //操作权重
	ActionTime   []int32   `json:"action_time" bson:"action_time"`     //操作时间
	SeeWeight    []int32   `json:"see_weight" bson:"see_weight"`       //看牌权重
	SeeTime      [][]int32 `json:"see_time" bson:"see_time"`           //看牌时机
	AgreeBi      int32     `json:"agree_bi" bson:"agree_bi"`           //被比牌同意概率
}

type AK47NewbieMode struct {
	CanWithdrawRange              []int `json:"can_withdraw_range" bson:"can_withdraw_range"`
	WinRate                       []int `json:"win_rate" bson:"win_rate"`
	CanWithdrawLimit              int   `json:"can_withdraw_limit" bson:"can_withdraw_limit"`
	WinWeight                     []int `json:"win_weight" bson:"win_weight"`
	WinCardType                   []int `json:"win_card_type" bson:"win_card_type"`
	LoseWeight                    []int `json:"lose_weight" bson:"lose_weight"`
	LoseCardType                  []int `json:"lose_card_type" bson:"lose_card_type"`
	CivilianCanWithdrawRange      []int `json:"civilian_can_withdraw_range" bson:"civilian_can_withdraw_range"`
	CivilianWinRate               []int `json:"civilian_win_rate" bson:"civilian_win_rate"`
	CivilianCanWithdrawLimit      int   `json:"civilian_can_withdraw_limit" bson:"civilian_can_withdraw_limit"`
	CivilianWinWeight             []int `json:"civilian_win_weight" bson:"civilian_win_weight"`
	CivilianWinCardType           []int `json:"civilian_win_card_type" bson:"civilian_win_card_type"`
	CivilianLoseWeight            []int `json:"civilian_lose_weight" bson:"civilian_lose_weight"`
	CivilianLoseCardType          []int `json:"civilian_lose_card_type" bson:"civilian_lose_card_type"`
	CivilianTriggerStrategyProb   int   `json:"civilian_trigger_strategy_prob" bson:"civilian_trigger_strategy_prob"`
	SpecialRound                  []int `json:"special_round" bson:"special_round"`
	SpecialRoundCardType          []int `json:"special_round_card_type" bson:"special_round_card_type"`
	NewbieToCivilianWithdrawLimit int   `json:"newbie_to_civilian_withdraw_limit" bson:"newbie_to_civilian_withdraw_limit"`
	FoamCheckTotalDepositLimit    int   `json:"foam_check_total_deposit_limit" bson:"foam_check_total_deposit_limit"`
	FoamGiftRate                  int   `json:"foam_gift_rate" bson:"foam_gift_rate"`
	FoamCardType                  int   `json:"foam_card_type" bson:"foam_card_type"`
	FoamTriggerStrategyProb       int   `json:"foam_trigger_strategy_prob" bson:"foam_trigger_strategy_prob"`
}

type AK47ControlType struct {
	ChargeRange        []int `json:"charge_range" bson:"charge_range"`
	ControlType        int   `json:"control_type" bson:"control_type"`
	WinScoreConfig     []int `json:"win_score_config" bson:"win_score_config"`
	WinScoreWeight     []int `json:"win_score_weight" bson:"win_score_weight"`
	PlayerFactorRange  []int `json:"player_factor_range" bson:"player_factor_range"`
	PlayerFactorWeight []int `json:"player_factor_weight" bson:"player_factor_weight"`
	WinScoreLimit      int   `json:"win_score_limit" bson:"win_score_limit"`
}

type RMGame struct {
	Bottom           int               `json:"bottom" bson:"bottom"`                         //底注
	Otime            int               `json:"otime" bson:"otime"`                           //操作时间
	Scountdown       int               `json:"scountdown" bson:"scountdown"`                 //结算倒数时间
	Match_Time       []int             `json:"match_time" bson:"match_time"`                 //匹配时间区间
	Single_Robot     []int             `json:"single_robot" bson:"single_robot"`             //单局人机数区间
	Robot_Join       []int             `json:"robot_join" bson:"robot_join"`                 //人机加入概率
	Robot_Leave      []int             `json:"robot_leave" bson:"robot_leave"`               //人机离开概率(万分比)
	Prevent_Time     int               `json:"prevent_time" bson:"prevent_time"`             //防作弊匹配时间检测
	Prevent_Num      int               `json:"prevent_num" bson:"prevent_num"`               //防作弊匹配次数检测
	Prevent_Thaw     int               `json:"prevent_thaw" bson:"prevent_thaw"`             //防作弊匹配解冻时间
	Msg_Score        int               `json:"msg_score" bson:"msg_score"`                   //跑马灯显示分
	NewbieType       int               `json:"newbie_type" bson:"newbie_type"`               //新手状态使用摸牌ID
	AbnormalType     int               `json:"abnormal_type" bson:"abnormal_type"`           //异常状态使用摸牌ID
	FinalFactorRange []int             `json:"final_factor_range" bson:"final_factor_range"` //正常状态和点控最终系数区间
	CardTypeRange    []int             `json:"card_type_range" bson:"card_type_range"`       //正常状态和点控系数对应使用牌型
	MingTax          int32             `json:"ming_tax" bson:"ming_tax"`                     //明税
	AnTax            int32             `json:"an_tax" bson:"an_tax"`                         //暗税
	RoomFactorSwitch int               `json:"room_factor_switch" bson:"room_factor_switch"` //房间系数开关
	Mode             int               `json:"mode" bson:"mode"`                             //模式
	ACardType        int               `json:"a_card_type" bson:"a_card_type"`               //A类指定牌型
	CardType         []RMGameCardType  //发牌牌型配置
	CardType2        []RMGameCardType2 //牌型配置
	NewbieMode       RMNewbieMode      // B类新手模式
	ANewbieMode      RMNewbieMode      // A类新手模式
	RoomRecharge     []int32           `json:"room_recharge" bson:"room_recharge"`           //局内充值
	RoomRechargeGive []int32           `json:"room_recharge_give" bson:"room_recharge_give"` //局内充值赠送
}

func (rm RMGame) FindCardType(id int) RMGameCardType {
	for _, v := range rm.CardType {
		if v.Id == id {
			return v
		}
	}
	return RMGameCardType{}
}

func (rm RMGame) FindCardType2(id int) RMGameCardType2 {
	for _, v := range rm.CardType2 {
		if v.Id == id {
			return v
		}
	}
	return RMGameCardType2{}
}

type RMGameCardType struct {
	Id                            int   `json:"id" bson:"id"`                                                   //牌型ID
	PlayerWeight                  []int `json:"player_weight" bson:"player_weight"`                             //玩家初始牌型权重
	PlayerWeight2CardTypeId       []int `json:"player_weight2_card_type_id" bson:"player_weight2_card_type_id"` //玩家初始牌型权重对应初始牌型ID
	PlayerDrawCardWeight          []int `json:"player_draw_card_weight" bson:"player_draw_card_weight"`         //玩家摸牌权重[随机牌，边张牌，JOKER牌]
	RobotWeigtht                  []int `json:"robot_weigtht" bson:"robot_weigtht"`                             //人机初始牌型权重
	RobotWeight2CardTypeId        []int `json:"robot_weight2_card_type_id" bson:"robot_weight2_card_type_id"`   //人机初始牌型权重对应初始牌型ID
	RobotDrawCardWeight           []int `json:"robot_draw_card_weight" bson:"robot_draw_card_weight"`           //人机摸牌权重[随机牌，边张牌，JOKER牌]
	RobotActionTime               []int `json:"robot_action_time" bson:"robot_action_time"`                     //人机操作思考时间毫秒（最小值，最大值）
	CardPoolIdWeight              []int `json:"card_pool_id_weight" bson:"card_pool_id_weight"`
	CardPoolId                    []int `json:"card_pool_id" bson:"card_pool_id"`
	UpCardPool                    int   `json:"up_card_pool" bson:"up_card_pool"`
	MustLoseProb                  int   `json:"must_lose_prob" bson:"must_lose_prob"`
	MustLoseScore                 []int `json:"must_lose_score" bson:"must_lose_score"`
	MustLoseCoreRobotGoodCardProb int   `json:"must_lose_core_robot_good_card_prob" bson:"must_lose_core_robot_good_card_prob"`
	WinScoreLimit                 int   `json:"win_score_limit" bson:"win_score_limit"`
}

type RMGameCardType2 struct {
	Id             int      `json:"id" bson:"id"`                             //初始牌型ID
	CardTypeConfig []string `json:"card_type_config" bson:"card_type_config"` //牌型配置
}

type RMNewbieMode struct {
	Id                       string `xls:"房间ID"`
	CanWithdrawRange         []int  `json:"can_withdraw_range" bson:"can_withdraw_range"`
	WinRate                  []int  `json:"win_rate" bson:"win_rate"`
	CanWithdrawLimit         int    `json:"can_withdraw_limit" bson:"can_withdraw_limit"`
	WinWeight                []int  `json:"win_weight" bson:"win_weight"`
	WinCardType              []int  `json:"win_card_type" bson:"win_card_type"`
	LoseWeight               []int  `json:"lose_weight" bson:"lose_weight"`
	LoseCardType             []int  `json:"lose_card_type" bson:"lose_card_type"`
	CivilianCanWithdrawRange []int  `json:"civilian_can_withdraw_range" bson:"civilian_can_withdraw_range"`
	CivilianWinRate          []int  `json:"civilian_win_rate" bson:"civilian_win_rate"`
	CivilianCanWithdrawLimit int    `json:"civilian_can_withdraw_limit" bson:"civilian_can_withdraw_limit"`
	CivilianWinWeight        []int  `json:"civilian_win_weight" bson:"civilian_win_weight"`
	CivilianWinCardType      []int  `json:"civilian_win_card_type" bson:"civilian_win_card_type"`
	CivilianLoseWeight       []int  `json:"civilian_lose_weight" bson:"civilian_lose_weight"`
	CivilianLoseCardType     []int  `json:"civilian_lose_card_type" bson:"civilian_lose_card_type"`
	SpecialRound             []int  `json:"special_round" bson:"special_round"`
	SpecialRoundCardType     []int  `json:"special_round_card_type" bson:"special_round_card_type"`
	FoamCardType             int    `json:"foam_card_type" bson:"foam_card_type"`
}

type LHDGame struct {
	ChipLimit        []uint32             `bson:"lh_chip_limit" json:"lh_chip_limit"`         //筹码额度
	BetLimit         []int32              `bson:"lh_bet_limit" json:"lh_bet_limit"`           //下注上限(龙，虎，和)
	BetInterval      [][]int32            `bson:"lh_bet_interval" json:"lh_bet_interval"`     //总下注区间(-1无限制)
	BetRoom          [][]int32            `bson:"lh_bet_room" json:"lh_bet_room"`             //总下注对应房间
	NoviceRoom       []int32              `bson:"lh_novice_room" json:"lh_novice_room"`       //新手房间
	ExceptionRoom    []int32              `bson:"lh_exception_room" json:"lh_exception_room"` //异常房间
	ControlRoom      []int32              `bson:"lh_control_room" json:"lh_control_room"`     //点控房间
	FinalCoefficient []int32              `bson:"final_coefficient" json:"final_coefficient"` //最终系数区间
	Winning          []int32              `bson:"winning" json:"winning"`                     //系数对应胜率
	LoseLimited      int32                `bson:"lose_limited" json:"lose_limited"`           //单局输分上限
	Cheat            int32                `bson:"cheat" json:"cheat"`                         //防刷水局数
	CheatTie         int32                `bson:"cheat_tie" json:"cheat_tie"`                 //刷水开和概率(万分比)
	LHBetTime        int32                `bson:"lh_bet_time" json:"lh_bet_time"`             //下注时间
	LHMTax           int32                `bson:"lh_m_tax" json:"lh_m_tax"`                   //明税(万分比)
	LHATax           int32                `bson:"lh_a_tax" json:"lh_a_tax"`                   //暗税(万分比)
	Msg_Score        int64                `json:"msg_score" bson:"msg_score"`                 //跑马灯显示分
	NewbiewMode      FreeNewbiewMode      `json:"newbiew_mode" bson:"newbiew_mode"`           //B新手模式
	ANewbiewMode     FreeNewbiewMode      `json:"aNewbiewMode" bson:"a_newbiew_mode"`         //A新手模式
	Robot            LHDGameRobotStrategy `json:"robot" bson:"robot"`                         //人机策略
}

type LHDGameRobotStrategy struct {
	Num        int32   `json:"num" bson:"num"`                 //人数
	InitScore  []int32 `json:"init_score" bson:"init_score"`   //初始分数
	BetWeight  []int32 `json:"bet_weight" bson:"bet_weight"`   //下注权重
	TimeArea   []int32 `json:"time_area" bson:"time_area"`     //时间区间
	Min        int32   `json:"min" bson:"min"`                 //最小下注基数
	BetArea    []int32 `json:"bet_area" bson:"bet_area"`       //下注区间
	BetPro     []int32 `json:"bet_Pro" bson:"bet_Pro"`         //下注概率
	LeaveLimit []int32 `json:"leave_limit" bson:"leave_limit"` //携带退出上下限
	BetTimes   []int32 `json:"bet_times" bson:"bet_times"`     //下注次数退出上下限
}

type LHDXXSCStrategy struct {
	*BaseStrategy         // 基础信息
	M             []int64 `json:"m" bson:"m"` //援助返奖点（m）
	// A             []int   `json:"a" bson:"a"`   //怀疑警戒（a）
	// D             []int64 `json:"d" bson:"d"`   //allin特征（d）
	// B             []int   `json:"b" bson:"b"`   //扰动冷却（b）
	// S             []int32 `json:"s" bson:"s"`   //赢倍上限（s）
	U  []int   `json:"u" bson:"u"`   //有效触发次数上限（u）
	UZ []int   `json:"uz" bson:"uz"` //总触发次数上限（u总）
	C  []int64 `json:"c" bson:"c"`   //冷却线（c）
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

type LHDGCYXStrategy struct {
	*BaseStrategy           // 基础信息
	X             int64     //未付费玩家默充金额（x)
	HP            []int     // 基础高潮率（hp）
	HP1           []int     // 递增高潮率（hp1）
	FP            []float64 // 禁止高潮返奖率（fp）
	H             []int     // 不爆倍数上限（h）
	P0            []float64 // 潮后返奖率（p0）
	G             []int     // 高潮盈利倍率上限（g）
	F             []int     // 前戏局局数（f）
	S             []int     // 单日高潮次数上限（s）
	C             []int     // 贤者局数（c）
}

type FreeNewbiewMode struct {
	OutCashInterval []int32 // 新手可提彩金区间
	Winning         []int32 // 对应胜率(万分比）
	OutCashLimited  int32   // 新手可提现彩金上限
}

type CRASHGame struct {
	ChipLimit        []uint32               `bson:"lh_chip_limit" json:"lh_chip_limit"`         //筹码额度
	ChipSeat         int                    `bson:"lh_chip_seat" json:"lh_chip_seat"`           //筹码位置
	RobotChipPro     []int32                `json:"robotChipPro" bson:"robot_chip_pro"`         //人机选筹码概率
	BetLimit         int32                  `bson:"lh_bet_limit" json:"lh_bet_limit"`           //下注上限
	BetInterval      [][]int32              `bson:"lh_bet_interval" json:"lh_bet_interval"`     //总下注区间(-1无限制)
	BetRoom          [][]int32              `bson:"lh_bet_room" json:"lh_bet_room"`             //总下注对应房间
	NoviceRoom       []int32                `bson:"lh_novice_room" json:"lh_novice_room"`       //新手房间
	ExceptionRoom    []int32                `bson:"lh_exception_room" json:"lh_exception_room"` //异常房间
	ControlRoom      []int32                `bson:"lh_control_room" json:"lh_control_room"`     //点控房间
	FinalCoefficient []int32                `bson:"final_coefficient" json:"final_coefficient"` //最终系数区间
	CoefficientFix   []int32                `bson:"winning" json:"winning"`                     //系数修正
	LoseLimited      int32                  `bson:"lose_limited" json:"lose_limited"`           //单局输分上限
	BetTime          int32                  `json:"betTime" bson:"bet_time"`                    //下注时间
	MTax             int32                  `json:"mTax" bson:"m_tax"`                          //明税(万分比)
	ATax             int32                  `json:"aTax" bson:"a_tax"`                          //暗税(万分比)
	BoomPro1         int32                  `json:"boomPro1" bson:"boom_pro1"`                  //区间1爆炸概率(万分比)
	BoomPro2         int32                  `json:"boomPro2" bson:"boom_pro2"`                  //区间2爆炸概率(万分比)
	BoomPro3         int32                  `json:"boomPro3" bson:"boom_pro3"`                  //区间3爆炸概率(万分比)
	BoomPro4         int32                  `json:"boomPro4" bson:"boom_pro4"`                  //区间4爆炸概率(万分比)
	InstantBangPro   int32                  `json:"instant_bang_pro" bson:"instant_bang_pro"`   //秒爆
	Msg_Score        int64                  `json:"msg_score" bson:"msg_score"`                 //跑马灯显示分
	Jackpot          []int32                `json:"jackpot" bson:"jackpot"`                     //jackpot奖励倍数
	LoseFlow         []int32                `json:"loseFlow" bson:"lose_flow"`                  //可赔付分数浮动区间
	FlyMultiple      [][]int32              `json:"flyMultiple" bson:"fly_multiple"`            //无人下注飞行倍数区间
	MaxWithdrawable  []int32                `json:"maxWithdrawable" bson:"max_withdrawable"`    //新手可提彩金上限
	BackRate         float64                `json:"backRate" bson:"back_rate"`                  // 返奖率
	Robot            CRASHGameRobotStrategy `json:"robot" bson:"robot"`                         //人机策略
}

type CRASHGameRobotStrategy struct {
	Num        int32   `json:"num" bson:"num"`                 //人数
	InitScore  []int32 `json:"init_score" bson:"init_score"`   //初始分数
	TimeArea   []int32 `json:"time_area" bson:"time_area"`     //时间区间
	Min        int32   `json:"min" bson:"min"`                 //最小下注基数
	BetArea    []int32 `json:"bet_area" bson:"bet_area"`       //下注区间
	BetPro     []int32 `json:"bet_Pro" bson:"bet_Pro"`         //下注概率
	LeaveLimit []int32 `json:"leave_limit" bson:"leave_limit"` //携带退出上下限
	BetTimes   []int32 `json:"bet_times" bson:"bet_times"`     //下注次数退出上下限
	RobotLeave []int32 `json:"robotLeave" bson:"robot_leave"`  //人机逃离概率
}

type BaseStrategy struct {
	Id         int   `json:"id" bson:"_id"`                // id
	O          int   `json:"o" bson:"o"`                   // 开关
	Weight     int   `json:"weight" bson:"weight"`         // 优先级
	Mutex      []int `json:"mutex" bson:"mutex"`           // 互斥策略
	UserType   []int `json:"userType" bson:"userType"`     // 适用用户类型
	ChargeType []int `json:"chargeType" bson:"chargeType"` // 充值类型
}

type CRASHFYZSStrategy struct {
	*BaseStrategy `json:"crashbasestrategy" bson:"crashbasestrategy"` // 基础信息
	// O          int     `json:"o" bson:"o"`                   // 开关
	// UserType   []int   `json:"userType" bson:"userType"`     // 适用用户类型
	// ChargeType []int   `json:"chargeType" bson:"chargeType"` // 充值类型
	// T  int     `json:"t" bson:"t"`   // 玩次数上限
	// R  int     `json:"r" bson:"r"`   // 局数上限
	// RZ int     `json:"rz" bson:"rz"` // 总生效局数上限
	X  int64     `json:"x" bson:"x"`   // 玩家携带金额上限
	Q  float64   `json:"q" bson:"q"`   // 单点爆率修正值
	C  []int64   `json:"c" bson:"c"`   // 冷却线
	U  []int     `json:"u" bson:"u"`   // 触发次数上限
	UZ []int     `json:"uz" bson:"uz"` // 总触发次数上限
	H  []float64 `json:"h" bson:"h"`   // 不爆倍数上限
	M  []int64   `json:"m" bson:"m"`   // 携带金额上限
	RR []float64 `json:"rr" bson:"rr"` // 风控返奖率
	PR []int64   `json:"pr" bson:"pr"` // 风控赢钱上限
}

type CRASHYHWMStrategy struct {
	*BaseStrategy `json:"crashbasestrategy" bson:"crashbasestrategy"` // 基础信息
	R             []int                                               `json:"r" bson:"r"` // 监控局数（r）
	M             []float64                                           `json:"m" bson:"m"` // 逃跑中位数警戒线（m）
	A             []float64                                           `json:"a" bson:"a"` // 逃跑平均倍数警戒线（a）
	P             []int32                                             `json:"p" bson:"p"` // 瞬爆概率（p）
	X             []float64                                           `json:"x" bson:"x"` // 收割倍率（x）
	T             []int                                               `json:"t" bson:"t"` // 生效次数上限（t）
}

type CRASHXJQBStrategy struct {
	*BaseStrategy `json:"crashbasestrategy" bson:"crashbasestrategy"` // 基础信息
	P0            float64                                             `json:"p0" bson:"p0"` // 录单返奖率（p1）
	P1            []float64                                           `json:"p1" bson:"p1"` // 观察返奖率（p1）
	P2            []float64                                           `json:"p2" bson:"p2"` // 逃后返奖率（p2）
}

// 起死回生
type CRASHQSHSStrategy struct {
	*BaseStrategy `json:"crashbasestrategy" bson:"crashbasestrategy"` // 基础信息
	M             []float64                                           `json:"m" bson:"m"`   // 援助返奖点m
	D             []int64                                             `json:"d" bson:"d"`   // allin特征d
	N             []float64                                           `json:"n" bson:"n"`   // 期望返奖点n
	A             []float64                                           `json:"a" bson:"a"`   // 期高返奖点o
	Q             float64                                             `json:"q" bson:"q"`   // 单点爆率修正值
	S             []float64                                           `json:"s" bson:"s"`   // 不爆倍数上限（s）
	R             []int                                               `json:"r" bson:"r"`   // 赢局次数(r)
	RR            []float64                                           `json:"rr" bson:"rr"` // 风控返奖率（rr）
	PF            []int64                                             `json:"pf" bson:"pf"` // 风控赢钱上限（pf）
	T             []int                                               `json:"t" bson:"t"`   // 固定触发次数上限（t）
	P             []int32                                             `json:"p" bson:"p"`   // 随机拯救率
	X             int64                                               `json:"x" bson:"x"`
}

// 奖池风控
type CRASHJCFKStrategy struct {
	*BaseStrategy `json:"crashbasestrategy" bson:"crashbasestrategy"` // 基础信息
	M             []float64                                           `json:"m" bson:"m"` // 赢家上限（m）
	N             int64                                               `json:"n" bson:"n"` // 未付费玩家默充金额（n）
	B             []float64                                           `json:"b" bson:"b"` // 警戒倍数（b）
	Q             float64                                             `json:"q" bson:"q"` // 单点爆率修正值（q）
	T             []int                                               `json:"t" bson:"t"` // 自由空间（t）
}

// 冒险奖励
type CRASHMXJLStrategy struct {
	*BaseStrategy `json:"crashbasestrategy" bson:"crashbasestrategy"` // 基础信息
	F             []float64                                           `json:"f" bson:"f"`   // 奖励点（f)
	R             []int                                               `json:"r" bson:"r"`   // 码局监控（r）
	C             []int64                                             `json:"c" bson:"c"`   // 码量监控（c）
	W             []int                                               `json:"w" bson:"w"`   // 博局监控（w）
	M             []float64                                           `json:"m" bson:"m"`   // 博倍监控（m）
	N             []int64                                             `json:"n" bson:"n"`   // 当局码量（n）
	X             int64                                               `json:"x" bson:"x"`   // 未付费玩家默充金额（x）
	S             []float64                                           `json:"s" bson:"s"`   // 不爆倍数上限（s）
	Q             float64                                             `json:"q" bson:"q"`   // 单点爆率修正值（q）
	RR            []float64                                           `json:"rr" bson:"rr"` // 风控返奖率（rr）
	PF            []int64                                             `json:"pf" bson:"pf"` // 风控赢钱上限（pf）
	T             []int                                               `json:"t" bson:"t"`   // 固定触发次数上限（t）
	P             []int32                                             `json:"p" bson:"p"`   // 随机拯救概率（p）
}

// 人狂有祸
type CRASHRKYHStrategy struct {
	*BaseStrategy `json:"crashbasestrategy" bson:"crashbasestrategy"` // 基础信息
	PU            []float64                                           `json:"pu" bson:"pu"` // 天罚线（pu）
	P7            []float64                                           `json:"p7" bson:"p7"` // 天罚返奖率（p7）
	P1            []int32                                             `json:"p1" bson:"p1"` // 瞬爆概率（p1）
	W             []int64                                             `json:"w" bson:"w"`   // 赢局监控（w）
	C             []int64                                             `json:"c" bson:"c"`   // 肥码上限（c）
	P2            []int32                                             `json:"p2" bson:"p2"` // 天罚概率（p2）
	N             []float64                                           `json:"n" bson:"n"`   // 罚线调控（n）
	X             int64                                               `json:"x" bson:"x"`   // 未付费玩家默充金额（x）
}

// 高潮涌现
type CRASHGCYXStrategy struct {
	*BaseStrategy `json:"crashbasestrategy" bson:"crashbasestrategy"` // 基础信息
	X             int64                                               `json:"x" bson:"x"`     // 未付费玩家默充金额（x）
	HP            []int32                                             `json:"hp" bson:"hp"`   // 基础高潮率（hp）
	HP1           []int32                                             `json:"hp1" bson:"hp1"` // 递增高潮率（hp1）
	FP            []float64                                           `json:"fp" bson:"fp"`   // 禁止高潮返奖率（fp）
	H             []int32                                             `json:"h" bson:"h"`     // 不爆倍数上限（h）
	P0            []float64                                           `json:"p0" bson:"p0"`   // 潮后返奖率（p0）
	G             []int32                                             `json:"g" bson:"g"`     // 高潮盈利倍率上限（g）
	F             []int32                                             `json:"f" bson:"f"`     // 前戏局局数（f）
	S             []int32                                             `json:"s" bson:"s"`     // 单日高潮次数上限（s）
	C             []int32                                             `json:"c" bson:"c"`     // 贤者局数（c）
}

type ABGame struct {
	ChipLimit        []uint32            `bson:"ab_chip_limit" json:"ab_chip_limit"`           //筹码额度
	BetLimit         []int32             `bson:"ab_bet_limit" json:"ab_bet_limit"`             //下注上限(位置1-10)
	BetInterval      [][]int32           `bson:"ab_bet_interval" json:"ab_bet_interval"`       //总下注区间(-1无限制)
	BetRoom          [][]int32           `bson:"ab_bet_room" json:"ab_bet_room"`               //总下注对应房间
	NoviceRoom       []int32             `bson:"ab_novice_room" json:"ab_novice_room"`         //新手房间
	ExceptionRoom    []int32             `bson:"ab_exception_room" json:"ab_exception_room"`   //异常房间
	ControlRoom      []int32             `bson:"ab_control_room" json:"ab_control_room"`       //点控房间
	FinalCoefficient []int32             `bson:"final_coefficient" json:"final_coefficient"`   //最终系数区间
	CardTypeID       []int32             `json:"cardTypeId" bson:"card_type_id"`               //系数对应牌型ID
	LoseLimited      int32               `bson:"lose_limited" json:"lose_limited"`             //单局输分上限
	BetTime          int32               `bson:"ab_bet_time" json:"ab_bet_time"`               //下注时间
	MTax             int32               `bson:"ab_m_tax" json:"ab_m_tax"`                     //明税(万分比)
	ATax             int32               `bson:"ab_a_tax" json:"ab_a_tax"`                     //暗税(万分比)
	Msg_Score        int64               `json:"msg_score" bson:"msg_score"`                   //跑马灯显示分
	NoPeopleBet      int                 `json:"noPeopleBet" bson:"no_people_bet"`             //无人下注牌型
	CardType         []ABGameCardType    `json:"cardType" bson:"card_type"`                    //牌型
	Robot            ABGameRobotStrategy `json:"robot" bson:"robot"`                           //人机策略
	NewbiewMode      ABNewbieMode        `json:"newbiewMode" bson:"newbiew_mode"`              //新手策略
	ANewbiewMode     ABNewbieMode        `json:"aNewbiewMode" bson:"a_newbiew_mode"`           //A类新手策略
	LeadWeights      []ABLeadWeight      `json:"leadWeights" bson:"lead_weights"`              // 结果权重配置
	DealerAccess     int                 `json:"dealer_access" bson:"dealer_access"`           // 庄家携带需求
	MinFirstEntry    int                 `json:"min_first_entry" bson:"min_first_entry"`       // 初始携带要求
	Min_Access       int                 `json:"min_access" bson:"min_access"`                 // 每局最低准入
	Count            int                 `json:"count" bson:"count"`                           // 人数上限
	Bottom           uint32              `json:"bottom" bson:"bottom"`                         // 底注
	OperateTime      int                 `json:"operate_time" bson:"operate_time"`             // 操作时间
	RoomRecharge     []int32             `json:"room_recharge" bson:"room_recharge"`           //局内充值
	RoomRechargeGive []int32             `json:"room_recharge_give" bson:"room_recharge_give"` //局内充值赠送
	RTPFixRate       []float64           `json:"rtp_fix_rate" bson:"rtp_fix_rate"`             //RTP修正系数
}

type ABGameRobotStrategy struct {
	Num        int32   `json:"num" bson:"num"`                 //人数
	InitScore  []int32 `json:"init_score" bson:"init_score"`   //初始分数
	BetWeight1 []int32 `json:"betWeight1" bson:"bet_weight1"`  //下注权重1
	BetWeight2 []int32 `json:"betWeight2" bson:"bet_weight2"`  //下注权重2
	TimeArea   []int32 `json:"time_area" bson:"time_area"`     //时间区间
	Min        int32   `json:"min" bson:"min"`                 //最小下注基数
	BetPro     []int32 `json:"betPro" bson:"bet_pro"`          //下注概率
	BetArea    []int32 `json:"bet_area" bson:"bet_area"`       //下注区间
	LeaveLimit []int32 `json:"leave_limit" bson:"leave_limit"` //携带退出上下限
	BetTimes   []int32 `json:"bet_times" bson:"bet_times"`     //下注次数退出上下限
}

type ABNewbieMode struct {
	CanWithdrawRange         []int `json:"can_withdraw_range" bson:"can_withdraw_range"`
	WinRate                  []int `json:"win_rate" bson:"win_rate"`
	CanWithdrawLimit         int   `json:"can_withdraw_limit" bson:"can_withdraw_limit"`
	WinWeight                []int `json:"win_weight" bson:"win_weight"`
	WinCardType              []int `json:"win_card_type" bson:"win_card_type"`
	LoseWeight               []int `json:"lose_weight" bson:"lose_weight"`
	LoseCardType             []int `json:"lose_card_type" bson:"lose_card_type"`
	CivilianCanWithdrawRange []int `json:"civilian_can_withdraw_range" bson:"civilian_can_withdraw_range"`
	CivilianWinRate          []int `json:"civilian_win_rate" bson:"civilian_win_rate"`
	CivilianCanWithdrawLimit int   `json:"civilian_can_withdraw_limit" bson:"civilian_can_withdraw_limit"`
	CivilianWinWeight        []int `json:"civilian_win_weight" bson:"civilian_win_weight"`
	CivilianWinCardType      []int `json:"civilian_win_card_type" bson:"civilian_win_card_type"`
	CivilianLoseWeight       []int `json:"civilian_lose_weight" bson:"civilian_lose_weight"`
	CivilianLoseCardType     []int `json:"civilian_lose_card_type" bson:"civilian_lose_card_type"`
	FrothCardType            int   `json:"frothCardType" bson:"froth_card_type"`
}

type ABSFJPStrategy struct {
	SFChargeFloor int64 `json:"sfChargeFloor" bson:"sf_charge_floor"` // 杀富充值下限
	SFFactorFloor int32 `json:"sfFactorFloor" bson:"sf_factor_floor"` // 杀富个人系数下限
	SFStockUpper  int64 `json:"sfStockUpper" bson:"sf_stock_upper"`   // 杀富库存上限
	SFTriggerPro  int32 `json:"sfTriggerPro" bson:"sf_trigger_pro"`   // 杀富触发概率
	JPChargeUpper int64 `json:"jpChargeUpper" bson:"jp_charge_upper"` // 济贫充值上限
	JPFactorUpper int32 `json:"jpFactorUpper" bson:"jp_factor_upper"` // 济贫个人系数上限
	JPStockFloor  int64 `json:"jpStockFloor" bson:"jp_stock_floor"`   // 济贫库存下限
	JPTriggerPro  int32 `json:"jpTriggerPro" bson:"jp_trigger_pro"`   // 济贫触发概率
}

type ABGameCardType struct {
	Id     int `json:"id" bson:"id"`          //牌型ID
	WinPro int `json:"winPro" bson:"win_pro"` // 平台胜率
	// DrawPro []int `json:"drawPro" bson:"draw_pro"` //开奖概率
}

type ABLeadWeight struct {
	Group           string `json:"group" bson:"group"`                       //组id
	PlatWinWeights  []int  `json:"platWinWeights" bson:"plat_win_weights"`   // 平台赢权重
	PlatLoseWeights []int  `json:"platLoseWeights" bson:"plat_lose_weights"` // 玩家赢权重
	NatureWeights   []int  `json:"natureWeights" bson:"nature_weights"`      // 自然概率加权权重
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

// 彩票
type CPGame struct {
	ChipLimit        []uint32            `bson:"ab_chip_limit" json:"ab_chip_limit"`         //筹码额度
	BetLimit         []int32             `bson:"ab_bet_limit" json:"ab_bet_limit"`           //下注上限(位置1-6)
	BetInterval      [][]int32           `bson:"ab_bet_interval" json:"ab_bet_interval"`     //总下注区间(-1无限制)
	BetRoom          [][]int32           `bson:"ab_bet_room" json:"ab_bet_room"`             //总下注对应房间
	NoviceRoom       []int32             `bson:"ab_novice_room" json:"ab_novice_room"`       //新手房间
	ExceptionRoom    []int32             `bson:"ab_exception_room" json:"ab_exception_room"` //异常房间
	ControlRoom      []int32             `bson:"ab_control_room" json:"ab_control_room"`     //点控房间
	FinalCoefficient []int32             `bson:"final_coefficient" json:"final_coefficient"` //最终系数区间
	WinningPro       []int32             `json:"winningPro" bson:"winning_pro"`              //系数对应平台胜率
	LoseLimited      int32               `bson:"lose_limited" json:"lose_limited"`           //单局输分上限
	BetTime          int32               `bson:"ab_bet_time" json:"ab_bet_time"`             //下注时间
	MTax             int32               `bson:"ab_m_tax" json:"ab_m_tax"`                   //明税(万分比)
	ATax             int32               `bson:"ab_a_tax" json:"ab_a_tax"`                   //暗税(万分比)
	Msg_Score        int64               `json:"msg_score" bson:"msg_score"`                 //跑马灯显示分
	UnRandomRate     int32               `json:"unRandomRate" bson:"un_random_rate"`         //非随机开牌概率
	NewbieMaxWin     int64               `json:"newbie_max_win" bson:"newbie_max_win"`       //新手赢分上限
	NewbieMinWin     int64               `json:"newbieMinWin" bson:"newbie_min_win"`         //新手赢分下限
	NewbieGiveRate   int32               `json:"newbieGiveRate" bson:"newbie_give_rate"`     //新手赢分下限送分概率
	WinnabilityRate  int32               `json:"winnabilityRate" bson:"winnability_rate"`    //可赢系数
	JackpotExceed    int64               `json:"jackpotExceed" bson:"jackpot_exceed"`        //jackpot超分
	DrawPro          []int               `json:"cardType" bson:"card_type"`                  //开奖权重
	NewbiewMode      FreeNewbiewMode     `json:"newbiew_mode" bson:"newbiew_mode"`           //B新手模式
	ANewbiewMode     FreeNewbiewMode     `json:"aNewbiewMode" bson:"a_newbiew_mode"`         //A新手模式
	Robot            CPGameRobotStrategy `json:"robot" bson:"robot"`                         //人机策略
	SFJP             CPSFJPStrategy      `json:"xfjp" bson:"xfjp"`                           //杀富济贫
	RTPFixRate       []float64           `json:"rtp_fix_rate" bson:"rtp_fix_rate"`           //概率调整系数
}

type CPGameRobotStrategy struct {
	Num        int32   `json:"num" bson:"num"`                 //人数
	InitScore  []int32 `json:"init_score" bson:"init_score"`   //初始分数
	BetWeight  []int32 `json:"bet_weight" bson:"bet_weight"`   //下注权重
	TimeArea   []int32 `json:"time_area" bson:"time_area"`     //时间区间
	Min        int32   `json:"min" bson:"min"`                 //最小下注基数
	BetArea    []int32 `json:"bet_area" bson:"bet_area"`       //下注区间
	BetPro     []int32 `json:"betPro" bson:"bet_pro"`          //下注概率
	LeaveLimit []int32 `json:"leave_limit" bson:"leave_limit"` //携带退出上下限
	BetTimes   []int32 `json:"bet_times" bson:"bet_times"`     //下注次数退出上下限
}

type CPSFJPStrategy struct {
	SFChargeFloor int64 `json:"xfChargeFloor" bson:"xf_charge_floor"` // 杀富充值下限
	SFFactorFloor int32 `json:"xfFactorFloor" bson:"xf_factor_floor"` // 杀富个人系数下限
	SFStockUpper  int64 `json:"xfStockUpper" bson:"xf_stock_upper"`   // 杀富库存上限
	SFTriggerPro  int32 `json:"xfTriggerPro" bson:"xf_trigger_pro"`   // 杀富触发概率
	JPChargeUpper int64 `json:"jpChargeUpper" bson:"jp_charge_upper"` // 济贫充值上限
	JPFactorUpper int32 `json:"jpFactorUpper" bson:"jp_factor_upper"` // 济贫个人系数上限
	JPStockFloor  int64 `json:"jpStockFloor" bson:"jp_stock_floor"`   // 济贫库存下限
	JPTriggerPro  int32 `json:"jpTriggerPro" bson:"jp_trigger_pro"`   // 济贫触发概率
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

type PLANEGame struct {
	ChipLimit        []uint32               `bson:"lh_chip_limit" json:"lh_chip_limit"`         //筹码额度
	RobotChipPro     []int32                `bson:"robot_chip_pro" json:"robot_chip_pro"`       //人机选筹码
	BetLimit         int32                  `bson:"lh_bet_limit" json:"lh_bet_limit"`           //下注上限
	BetInterval      [][]int32              `bson:"lh_bet_interval" json:"lh_bet_interval"`     //总下注区间(-1无限制)
	BetRoom          [][]int32              `bson:"lh_bet_room" json:"lh_bet_room"`             //总下注对应房间
	NoviceRoom       []int32                `bson:"lh_novice_room" json:"lh_novice_room"`       //新手房间
	ExceptionRoom    []int32                `bson:"lh_exception_room" json:"lh_exception_room"` //异常房间
	ControlRoom      []int32                `bson:"lh_control_room" json:"lh_control_room"`     //点控房间
	FinalCoefficient []int32                `bson:"final_coefficient" json:"final_coefficient"` //最终系数区间
	CoefficientFix   []int32                `bson:"winning" json:"winning"`                     //系数修正
	LoseLimited      int32                  `bson:"lose_limited" json:"lose_limited"`           //单局输分上限
	BetTime          int32                  `json:"betTime" bson:"bet_time"`                    //下注时间
	MTax             int32                  `json:"mTax" bson:"m_tax"`                          //明税(万分比)
	ATax             int32                  `json:"aTax" bson:"a_tax"`                          //暗税(万分比)
	BoomPro1         int32                  `json:"boomPro1" bson:"boom_pro1"`                  //区间1爆炸概率(万分比)
	BoomPro2         int32                  `json:"boomPro2" bson:"boom_pro2"`                  //区间2爆炸概率(万分比)
	BoomPro3         int32                  `json:"boomPro3" bson:"boom_pro3"`                  //区间3爆炸概率(万分比)
	BoomPro4         int32                  `json:"boomPro4" bson:"boom_pro4"`                  //区间4爆炸概率(万分比)
	InstantBangPro   int32                  `json:"instant_bang_pro" bson:"instant_bang_pro"`   //秒爆
	Msg_Score        int64                  `json:"msg_score" bson:"msg_score"`                 //跑马灯显示分
	Jackpot          []int32                `json:"jackpot" bson:"jackpot"`                     //jackpot奖励倍数
	LoseFlow         []int32                `json:"loseFlow" bson:"lose_flow"`                  //可赔付分数浮动区间
	FlyMultiple      [][]int32              `json:"flyMultiple" bson:"fly_multiple"`            //无人下注飞行倍数区间
	MaxWithdrawable  []int32                `json:"maxWithdrawable" bson:"max_withdrawable"`    //新手可提现彩金上限
	BackRate         float64                `json:"backRate" bson:"back_rate"`                  // 返奖率
	Robot            PLANEGameRobotStrategy `json:"robot" bson:"robot"`                         //人机策略
}

type PLANEGameRobotStrategy struct {
	Num        int32   `json:"num" bson:"num"`                 //人数
	InitScore  []int32 `json:"init_score" bson:"init_score"`   //初始分数
	TimeArea   []int32 `json:"time_area" bson:"time_area"`     //时间区间
	Min        int32   `json:"min" bson:"min"`                 //最小下注基数
	BetArea    []int32 `json:"bet_area" bson:"bet_area"`       //下注区间
	BetPro     []int32 `json:"bet_Pro" bson:"bet_Pro"`         //下注概率
	LeaveLimit []int32 `json:"leave_limit" bson:"leave_limit"` //携带退出上下限
	BetTimes   []int32 `json:"bet_times" bson:"bet_times"`     //下注次数退出上下限
	RobotLeave []int32 `json:"robotLeave" bson:"robot_leave"`  //人机逃离概率
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

// 高潮涌现
type PLANEGCYXStrategy struct {
	*BaseStrategy `json:"crashbasestrategy" bson:"crashbasestrategy"` // 基础信息
	X             int64                                               `json:"x" bson:"x"`     // 未付费玩家默充金额（x）
	HP            []int32                                             `json:"hp" bson:"hp"`   // 基础高潮率（hp）
	HP1           []int32                                             `json:"hp1" bson:"hp1"` // 递增高潮率（hp1）
	FP            []float64                                           `json:"fp" bson:"fp"`   // 禁止高潮返奖率（fp）
	H             []int32                                             `json:"h" bson:"h"`     // 不爆倍数上限（h）
	P0            []float64                                           `json:"p0" bson:"p0"`   // 潮后返奖率（p0）
	G             []int32                                             `json:"g" bson:"g"`     // 高潮盈利倍率上限（g）
	F             []int32                                             `json:"f" bson:"f"`     // 前戏局局数（f）
	S             []int32                                             `json:"s" bson:"s"`     // 单日高潮次数上限（s）
	C             []int32                                             `json:"c" bson:"c"`     // 贤者局数（c）
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

func GetGameList() []Game {
	var list []Game
	ListByQ(Games, bson.M{}, &list)
	return list
}

// 获取库存期望
func (game Game) GetStockExpect() int64 {
	return game.Stock_Expect
}

// Save 写入数据库
func (t *Game) Save() bool {
	//t.Ctime = bson.Now()
	return Upsert(Games, bson.M{"_id": t.Id}, t)
}

func (t *Game) Init() bool {
	return Insert(Games, t)
}

func GetPeopleList() []RoomPeople {
	var list []RoomPeople
	ListByQ(RoomPeoples, bson.M{}, &list)
	return list
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
	GCYX CRASHGCYXStrategy `json:"gcyx" bson:"gcyx"` // 高潮涌现
}

func GetCrashStrategy() CrashStrategy {
	strategy := CrashStrategy{}
	GetByQ(CrashStrategies, bson.M{"_id": 1}, &strategy)
	return strategy
}

func (c *CrashStrategy) Save() {
	Upsert(CrashStrategies, bson.M{"_id": 1}, &c)
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
	GCYX PLANEGCYXStrategy `json:"gcyx" bson:"gcyx"` // 高潮涌现
}

func GetAviatorStrategy() AviatorStrategy {
	strategy := AviatorStrategy{}
	GetByQ(PlaneStrategies, bson.M{"_id": 1}, &strategy)
	return strategy
}

func (c *AviatorStrategy) Save() {
	Upsert(PlaneStrategies, bson.M{"_id": 1}, &c)
}

type LhdStrategy struct {
	Id   int             `json:"id" bson:"_id"`
	XXSC LHDXXSCStrategy `json:"xxsc" bson:"xxsc"` // 心想事成
	QSBN LHDQSBNStrategy `json:"qsbn" bson:"qsbn"` // 求死不能
	LKYH LHDLKYHStrategy `json:"lkyh" bson:"lkyh"` // 龙狂有祸
	GCYX LHDGCYXStrategy `json:"gcyx" bson:"gcyx"` // 高潮涌现
}

func GetLHDStrategy() LhdStrategy {
	strategy := LhdStrategy{}
	GetByQ(LHDStrategies, bson.M{"_id": 1}, &strategy)
	return strategy
}

func (c *LhdStrategy) Save() {
	Upsert(LHDStrategies, bson.M{"_id": 1}, &c)
}

type SevenStrategy struct {
	Id   int             `json:"id" bson:"_id"`
	XXSC LHDXXSCStrategy `json:"xxsc" bson:"xxsc"` // 心想事成
	QSBN LHDQSBNStrategy `json:"qsbn" bson:"qsbn"` // 求死不能
	QGY  LHDLKYHStrategy `json:"lkyh" bson:"lkyh"` // 龙狂有祸
	GCYX LHDGCYXStrategy `json:"gcyx" bson:"gcyx"` // 高潮涌现
}

func GetSevenStrategy() SevenStrategy {
	strategy := SevenStrategy{}
	GetByQ(SevenStrategies, bson.M{"_id": 1}, &strategy)
	return strategy
}

func (c *SevenStrategy) Save() {
	Upsert(SevenStrategies, bson.M{"_id": 1}, &c)
}

type ABStrategy struct {
	Id   int            `json:"id" bson:"_id"`
	ANQS ABANQSStrategy `json:"anqs" bson:"anqs"` // 安能求死
	ARTY ABARTYStrategy `json:"arty" bson:"arty"` // 安然躺赢
}

func GetABStrategy() ABStrategy {
	strategy := ABStrategy{}
	GetByQ(ABStrategies, bson.M{"_id": 1}, &strategy)
	return strategy
}

func (c *ABStrategy) Save() {
	Upsert(ABStrategies, bson.M{"_id": 1}, &c)
}

type CPStrategy struct {
	Id   int            `json:"id" bson:"_id"`
	LWJY CPLWJYStrategy `json:"lwjy" bson:"lwjy"` // 来玩就赢
	LYQN CPLYQNStrategy `json:"lyqn" bson:"lyqn"` // 来易去难
}

func GetCPStrategy() CPStrategy {
	strategy := CPStrategy{}
	GetByQ(CPStrategies, bson.M{"_id": 1}, &strategy)
	return strategy
}

func (c *CPStrategy) Save() {
	Upsert(CPStrategies, bson.M{"_id": 1}, &c)
}

type RBStrategy struct {
	Id   int            `json:"id" bson:"_id"`
	HYDT RBHYDTStrategy `json:"hydt" bson:"hydt"` // 红运当头
	JCFS RBJCFSStrategy `json:"jcfs" bson:"jcfs"` // 决出逢生
}

func GetRBStrategy() RBStrategy {
	strategy := RBStrategy{}
	GetByQ(RBStrategies, bson.M{"_id": 1}, &strategy)
	return strategy
}

func (c *RBStrategy) Save() {
	Upsert(RBStrategies, bson.M{"_id": 1}, &c)
}
