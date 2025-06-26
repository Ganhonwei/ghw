package main

import (
	"goserver/gen/pb"
	"goserver/pkg/data"

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
	loggerPid *actor.PID
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
	// 离开中机器人数
	robotLeaving map[string]bool
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
}

const (
	//底池上限
	PotLimit = 1000
	//ReadyTime 准备超时时间
	ReadyTime = 3
	//播放发牌动画超时时间
	DealingTime = 5
	//BetTime 下注超时时间
	BetTime = 30
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
