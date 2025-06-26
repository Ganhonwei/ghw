package fortune_gems2

import (
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
	// roles map[string]*data.DeskRole
	role *data.DeskRole

	//消息路由 playerPid-userid
	router map[string]string

	GameId string //对局号
	//牌桌详情
	detail *data.Detail

	closeServer bool

	timer       int //状态计时
	timerDuring int //状态持续时间
	//关闭通道
	stopCh chan struct{}
	// //关闭时间
	// closeTime int
	// robotTime int
	// //暂停原因
	// reason int
	// //暂停时间
	// pauseTime int
	// //暂停自定义参数
	// pauseArg any

	state            int32   // 房间状态
	betNum           int64   // 下注额
	mines            int32   // 埋雷数
	minesPits        []int32 // 坑位*25 0安全,1有雷
	minesPitsUser    []int32 // 用户坑位*25: 0未知,1已踩无雷,2已踩有雷,3未踩无雷,4未踩有雷
	step             int32   // 选了几个格子了
	stepPits         []int32 // 选的格子
	autoMines        bool    // 自动对局
	autoRound        int32   // 自动下注剩余局数
	autoBets         int64   // 自动下注额
	autoPits         []int32 // 自动下注选择坑位
	minesStepBoomPit int32   // 最后踩雷坑位, -1没踩雷
	forceCashOut     bool    // 下注后一步都没有走, 强制结束
	settleScore      int64   // 结算分数
	settleMultiple   string  // 结算倍数
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
	//等待太久时间
	WaitTooLongTime = 10
	// 私人局结算页自动退出时间
	PrivSettleExitDelay int64 = 30
)
