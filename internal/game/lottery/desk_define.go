package lottery

import (
	"goserver/pkg/data"
	"goserver/pkg/game/algo"

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
	// FreeDealTime 开牌时间
	FreeLeadTime = 1
	//FreeSettlementTime 结算持续时间
	FreeSettlementTime = 3

	SF = 1 //劫富
	JP = 2 //济贫

	LWJY = 1 // 来玩就赢
	LYQN = 2 // 来易去难
)

var SeatProMap map[uint32]int32 = map[uint32]int32{1: 6429, 2: 1667, 3: 452, 4: 301, 5: 20, 6: 23}

func (t *Desk) initOdds() {
	t.DeskFree.Multiple[algo.BaoZi] = 0
	t.DeskFree.Multiple[algo.TongHuaShun] = 40000
	t.DeskFree.Multiple[algo.ShunZi] = 2000
	t.DeskFree.Multiple[algo.TongHua] = 1000
	t.DeskFree.Multiple[algo.DuiZi] = 500
	t.DeskFree.Multiple[algo.GaoPai] = 120
}
