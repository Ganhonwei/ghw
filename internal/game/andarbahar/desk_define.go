package andarbahar

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
	// 局内充值前状态
	chargePrevState int32
	//本轮是否局内充值
	isCharge bool
	// 充值中玩家座位号
	chargingSeats []uint32
	//位置充值次数
	ActRechargeTimes map[uint32]int
	//下注时充值成功的玩家座位号
	betingChargeSeatid uint32
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
	// 私人房申请换座
	changingSeat map[string]*ChangingSeat
	// 该回合结束后关闭牌桌
	dismissOnRoundOver bool
	// 关闭牌桌时间
	dismissDelayTime int64
	//暂停原因
	reason int
	//暂停时间
	pauseTime int
	//暂停自定义参数
	pauseArg interface{}
	// 结算页面延迟退出时间30s
	settleExitTime int64
	//
	privRoomRounds int
}

// 开奖结果
type DrawResult struct {
	Winner     uint32 // andar或bahar
	SideWinner uint32 // 边注位置
	Score      int64  // 输赢分
}

const (
	//FreeBetTime 下注超时时间
	FreeBetTime = 15
	// FreeDealTime 发牌时间
	ReadyTime = 6
	//FreeSettlementTime 结算持续时间
	FreeSettlementTime = 6
	//RestTime 休息超时时间
	RestTime = 10
	//SysCarry 系统上庄限额
	SysCarry int64 = 50000000
	//DealerTimes 做庄次数限制
	DealerTimes uint32 = 10
	//局内充值时间
	ChargeInGameTime = 5 * 60
	// 私人局结算页自动退出时间
	PrivSettleExitDelay int64 = 30

	Andar  = 1
	Bahar  = 2
	Seat3  = 3
	Seat4  = 4
	Seat5  = 5
	Seat6  = 6
	Seat7  = 7
	Seat8  = 8
	Seat9  = 9
	Seat10 = 10
)

// 策略
var PokerPro = make(map[int]float64)

const (
	ANQS = 1
	ARTY = 2
)

// 初始化赔率
func (a *Desk) initOdds() {
	a.ABDeskFree.OddsMap = map[uint32]int{Andar: 190,
		Bahar: 200, Seat3: 350, Seat4: 450, Seat5: 550,
		Seat6: 450, Seat7: 1500, Seat8: 2500, Seat9: 5000,
		Seat10: 12000}
}

func (a *Desk) initProbility() {
	lastPro := 0.0
	for i := 1; i <= 51; i++ {
		if i > 49 {
			break
		}
		pro := 3 / float64(52-i) * (1 - lastPro)
		lastPro += pro
		PokerPro[i] = pro
	}
}

// 私人房申请换座待同意
type ChangingSeat struct {
	Time     int64  // 申请时间,10秒后超时拒绝,1分钟内不能再申请
	Userid   string // 申请userid
	ToUserid string // 待同意userid
	Accept   uint32 // 1同意2拒绝
}
