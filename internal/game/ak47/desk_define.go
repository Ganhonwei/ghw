package ak47

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
	//分数
	score map[uint32]int64
	//结算
	over map[uint32]*pb.AK47CoinOver
	//本轮换牌类型
	changeCardType int
	//控制方式
	controlType int
	//当前赢分
	winScore int
	//玩家系数
	playerFactor int
	//当局用户赢分
	winScore1 int
	//当前赢分
	winScore2 int
	//充值金额
	chargeMoney int
	//本轮是否是策略局
	isStrategy bool
	//本轮是否局内充值
	isCharge bool
	//局内充值条件123
	condition int
	//策略局类型
	strategyType int
	//停服标记
	closeServer bool
	//比牌记录
	cmpRecord [][2]uint32
	//局内充值金额
	chargeAmount uint32
	//策略300主机A的牌比玩家大
	strategy300Bigger bool
	// 机器人假装充值时间20-40s随机
	robotChargingTime int
	// 局内充值中玩家座位号
	chargingSeat uint32
	//下注信息
	betInfo map[int64]int32
}

const (
	//底池上限
	PotLimit = 1000
	//ReadyTime 准备超时时间
	ReadyTime = 3
	//播放发牌动画超时时间
	DealingTime = 4
	//BetTime 下注超时时间
	BetTime = 12
	//RestTime 休息超时时间
	RestTime = 10
	//SysCarry 系统上庄限额
	SysCarry int64 = 50000000
	//DealerTimes 做庄次数限制
	DealerTimes uint32 = 8
	//局内充值时间
	ChargeInGameTime = 5 * 60
	//等待太久
	WaitTooLongTime = 10
)
