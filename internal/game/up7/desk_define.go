package up7

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

	//房间状态
	state int32
	//定时器状态
	tickStop bool
	//计时
	timer int
	//关闭通道
	stopCh chan struct{}
	//关闭时间
	closeTime int
	robotTime int
	//停服标记
	closeServer bool
}

const (
	// FreeReadyTime 准备时间
	FreeReadyTime = 1
	//FreeBetTime 下注超时时间
	// FreeBetTime = 15
	// FreeDealTime 发牌时间
	FreeDealTime = 2
	// FreeDealTime 开牌时间
	FreeLeadTime = 6
	//FreeSettlementTime 结算持续时间
	FreeSettlementTime = 5
	FreeCritTime       = 2 // 暴击展示等待时间

	Tie  uint32 = 0
	DOWN uint32 = 1
	UP   uint32 = 2
)

const (
	// 心想事成
	XXSC = 1
	// 求死不能
	QSBN = 2
	// 7管严
	QGY = 3
	// 高潮涌现
	GCYX = 4
)
