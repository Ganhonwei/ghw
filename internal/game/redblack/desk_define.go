package redblack

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

type DrawResult struct {
	Winner  uint32 // 获胜位置
	WinType uint32 // 获胜牌型
	Score   int64  // 输赢分
}

const (
	// FreeDealTime 发牌时间
	FreeDealTime = 4
	// FreeDealTime 开牌时间
	FreeLeadTime = 5
	//FreeSettlementTime 结算持续时间
	FreeSettlementTime = 6

	LUCK     uint32 = 0
	RED      uint32 = 1
	BLACK    uint32 = 2
	LUCkPAIR uint32 = 3
	COLOR    uint32 = 4
	SEQ      uint32 = 5
	PURESEQ  uint32 = 6
	SET      uint32 = 7

	// 策略
	HYDT = 1
	JCFS = 2
)

func (t *Desk) initOdds() {
	t.DeskFree.Multiple[RED] = 2
	t.DeskFree.Multiple[BLACK] = 2
	t.DeskFree.Multiple[LUCkPAIR] = 2
	t.DeskFree.Multiple[COLOR] = 3
	t.DeskFree.Multiple[SEQ] = 4
	t.DeskFree.Multiple[PURESEQ] = 10
	t.DeskFree.Multiple[SET] = 20
}
