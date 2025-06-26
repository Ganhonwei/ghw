package crash

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
	//计时
	timer      int
	tickStop   bool
	robotTimer int
	//关闭通道
	stopCh chan struct{}
	//关闭时间
	closeTime int
	robotTime int
	//停服标记
	closeServer bool
}

const (
	//FreeBetTime 下注超时时间
	// FreeBetTime = 15
	// FreeDealTime 准备时间
	// FreeDealTime = 60
	// FreeDealTime 开牌时间
	// FreeLeadTime = 5
	//FreeSettlementTime 结算持续时间
	FreeSettlementTime = 3

	// 策略类型
	NORMAL = 0 // 常规
	FYZS   = 1 // 扶摇直上
	YHWM   = 2 // 欲薅无门
	XJQB   = 3 // 虚假情报
	QSHS   = 4 // 起死回生
	JCFK   = 5 // 奖池风控
	MXJL   = 6 // 冒险奖励
	RKYH   = 7 // 人狂有祸
	GCYX   = 8 // 高潮涌现

	TagAutoCrash uint32 = 1 // 自动逃离标记
	TagNextBet   uint32 = 2 // 下一轮下注标记
)
