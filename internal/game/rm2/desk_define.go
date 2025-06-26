package rm2

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
	"sync"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// Desk 房间牌桌数据结构
type Desk struct {
	//name
	Name string
	//数据中心
	dbmsPid *actor.PID
	//房间服务
	roomPid *actor.PID
	//角色服务
	rolePid *actor.PID
	//日志服务
	// loggerPid *actor.PID
	//当前服务
	selfPid *actor.PID

	//房间类型基础数据
	*data.DeskData

	//房间玩家 key=userid
	roles map[string]*data.DeskRole
	//位置数据 key=seat (seat:1+)
	seats map[uint32]*data.DeskSeat
	//消息路由 playerPid-userid
	router map[string]string

	//牌桌当局数据
	*data.DeskGame
	//私人当局数据
	*data.DeskPriv
	//百人当局数据
	*data.DeskFree

	//牌桌当局数据
	*data.DeskAct

	//牌桌详情
	detail *data.Detail

	//房间状态
	state int32
	// 局内充值前状态
	chargePrevState int32
	//本轮是否局内充值
	isCharge bool
	// 充值中玩家座位号
	chargingSeats []uint32
	//计时
	timer int
	//关闭通道
	stopCh chan struct{}
	//关闭时间
	closeTime int
	robotTime int
	//暂停原因
	reason int
	//暂停时间
	pauseTime int
	//暂停自定义参数
	pauseArg interface{}
	//本轮机器人数
	robotNum int
	// 召唤中机器人数
	robotCalling int
	//本轮牌型ID
	cardTypeId int
	//本轮是否加牌
	addCard bool
	//分数
	score map[uint32]int64
	//结算
	over map[uint32]*pb.RMCoinGameOverInfo
	//停服标记
	closeServer bool
	// 该回合结束后关闭牌桌
	dismissOnRoundOver bool
	// 关闭牌桌延迟时间
	dismissDelaySeconds int64
	// 结算页面延迟退出时间30s
	settleExitTime int64
	//是否是必输局
	isMustLose bool
	//必输分数线
	mustLoseScore int
	//必输局核心人机摸好牌概率
	mustLoseCoreRobotGoodCardProb int
	//赢分限制
	winScoreLimit int
	//核心人机位置
	coreRobotSeat uint32
	//是否是升档牌库
	isUpCardPool bool
	//牌库编号
	cardPoolId int
	//发牌方式(1-随机分配 2-配置分配)
	dealType int
	//是否触发赢分限制
	isTriggerWinScoreLimit bool
	//是否触发必输控制
	isTriggerMustLose bool
	//是否触发必输控制摸牌限制
	isTriggerMustLoseDrawCard bool
	//是否触发必输控制上家出牌限制
	isTriggerMustLoseLastCard bool
	//是否触发必输控制核心人机摸好牌
	isTriggerMustLoseCoreRobotGoodCard bool
	// 控制局配置
	// control *tb.RmStockControlRecord
	control *RmControl
	// rm 控制
	*data.RMControl
}

var (
	tableMutex  sync.RWMutex
	rmRoiConfig *RMRoiConfig
)

const (
	//底池上限
	PotLimit = 1000
	//ReadyTime 准备超时时间
	ReadyTime = 3
	//播放发牌动画超时时间
	DealingTime = 5
	//BetTime 下注超时时间
	BetTime = 45
	//RestTime 休息超时时间
	RestTime = 10
	//SysCarry 系统上庄限额
	SysCarry int64 = 50000000
	//DealerTimes 做庄次数限制
	DealerTimes uint32 = 8
	//Debug开关
	IsDebug = true
	//等待太久
	WaitTooLongTime = 10
	//局内充值时间
	ChargeInGameTime = 5 * 60
	// 私人局结算页自动退出时间
	PrivSettleExitDelay int64 = 30
)

const (
	ActionNone    = iota
	ActionChu     // rm出牌
	ActionHu      // rm胡牌
	ActionZhaHu   // rm诈胡
	ActionQi      // rm弃牌
	ActionDeclare // rm结束组牌declare
)

const (
	TouchNone    = iota
	TouchCards   // 摸牌堆
	TouchQiCards // 摸弃牌堆
)

type RMRoiConfig struct {
	Drop []RMRoiDropStrategy `json:"drop"` // 弃牌配置
	Roi  []RMRoiStrategy     `json:"roi"`  // roi配置
}

type RMRoiStrategy struct {
	Id               string `json:"id"`               // 策略id
	Desc             string `json:"desc"`             // 描述
	RcCountRange     string `json:"rc_count_range"`   // 充值次数范围
	AmountRange      string `json:"amount_range"`     // 钱包携带金币数量
	Limit            int32  `json:"limit"`            // 对应ID策略的总生效次数
	DayLimit         int32  `json:"day_limit"`        // 每个自然日，对应ID策略生效次数
	WinRateRange     string `json:"win_rate_range"`   // 玩家收益率所在区间
	WinRate          int32  `json:"win_rate"`         // 期望玩家赢于玩家输的概率，负数为期望玩家输(万分比),负数判断EffctRate
	IsDie            string `json:"is_die"`           // 玩家的筹码是否小于房间底分的60倍
	NotDrawRobot1st  bool   `json:"notDrawRobot1st"`  // 是否保留机器人第一顺子
	NotDrawPlayer1st bool   `json:"notDrawPlayer1st"` // 是否保留玩家第一顺子
	DrawNumPlayer    string `json:"drawNumPlayer"`    // 玩家抽牌数量
	DrawNumRobot     string `json:"drawNumRobot"`     // 机器人抽牌数量
	Hand             string `json:"hand"`             // 对应rummy抽牌放的5，9，20
	CtrlMin          int32  `json:"ctrlMin"`          // 控制下限
	CtrlMax          int32  `json:"ctrlMax"`          // 控制上限
	EffctRate        int32  `json:"effctRate"`        // 人机干预概率(万分比)
}

// 人机弃牌策略
type RMRoiDropStrategy struct {
	Id              string `json:"id"`                // 弃牌策略id
	RcCountRange    string `json:"rc_count_range"`    // 充值次数范围
	PDropRate       string `json:"p_drop_rate"`       // 玩家在没有1Life时候首回合的弃牌率所在区间
	RDropRate       int    `json:"r_drop_rate"`       // 机器人首回合弃牌的概率，除100等于概率
	RScore          string `json:"r_score"`           // 玩家当前牌面分所处区间
	R1life          bool   `json:"r_1life"`           // 机器人是否有第一生命
	P1life          bool   `json:"p_1life"`           // 玩家是否有第一生命
	SubInningsRange string `json:"sub_innings_range"` // 子游戏对局数
	WinRateRange    string `json:"win_rate_range"`    // 玩家收益率所在区间
}

// rm控制参数
type RmControl struct {
	Ctype            int32 // 0.随机,1.库存控制,2.ROI策略
	Hierarchy        int32 // 进入控制档位
	WinRate          int32 // 期望玩家赢于玩家输的概率，负数为期望玩家输,该值除一百为概率，例30，为30%系统赢。
	NotDrawRobot1st  bool
	NotDrawPlayer1st bool
	DrawNumPlayer    []int32
	DrawNumRobot     []int32
	Hand             []int32

	CtrlMin   int32
	CtrlMax   int32
	EffctRate int32

	controlEffect    bool // 系统赢概率是否生效，是否控制发牌牌型
	controlRobotDraw bool // 系统进入干预流程，如果玩家打出的牌不是人机需要的牌，直接摸好牌
}
