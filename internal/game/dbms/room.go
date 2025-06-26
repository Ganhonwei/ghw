package dbms

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"runtime/debug"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/gogo/protobuf/proto"
)

var (
	roomPid *actor.PID
)

// RoomActor 房间列表服务
type RoomActor struct {
	Name string
	//房间列表
	rooms map[string]*actor.PID
	//玩家所在桌子userid-roomid
	router map[string]string
	//邀请码code-roomid
	codes map[string]*data.PrivRoom
	//房间人数roomid-numbers
	count map[string]uint32
	//匹配规则 rules
	//配置映射unique-roomid
	rules map[string]string
	//房间库存gameid-stock
	stocks map[string]*data.Stock
	//游戏库存gtype-stock
	gameStocks map[int32]*data.Stock
	//新手库存gtype-stock
	newbieStocks map[int32]*data.NewbieStock
	//游戏输赢分 gtype-score
	gameScores map[string]*data.CashFlow
	//房间人数
	onlineNum map[int32]map[string]int32
	//唯一id生成
	uniqueid *data.IDGen
	//关闭通道
	stopCh chan struct{}
	//更新状态
	status bool
	//计时
	timer int
	//提现排行榜当天结束时间 秒
	rankWithdrawTodayEndTime int64
	//提现排行榜本周结束时间 秒
	rankWithdrawWeekEndTime int64
	//提现排行榜机器人触发时间
	rankWithdrawRobotTimes map[string]int64
	//tp剧情局局内充值库存
	tpStoryChargeStock map[string]*data.TPStoryChargeStock
}

// Receive is sent messages to be processed from the mailbox associated with the instance of the actor
func (a *RoomActor) Receive(ctx actor.Context) {
	defer func() {
		if r := recover(); r != nil {
			glog.Error("Receive handler recover error:", r)
			debug.PrintStack()
		}
	}()
	switch msg := ctx.Message().(type) {
	case *pb.Request:
		ctx.Respond(&pb.Response{})
	case *actor.Started:
		glog.Notice("Starting, initialize actor here")
	case *actor.Stopping:
		glog.Notice("Stopping, actor is about to shut down")
	case *actor.Stopped:
		glog.Notice("Stopped, actor and its children are stopped")
	case *actor.Restarting:
		glog.Notice("Restarting, actor is about to restart")
	case *actor.ReceiveTimeout:
		glog.Infof("ReceiveTimeout: %v", ctx.Self().String())
	case proto.Message:
		a.Handler(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

func newRoomActor() actor.Actor {
	a := new(RoomActor)
	a.Name = cfg.Section("room").Name()
	glog.Debugf("a.Name %s", a.Name)
	a.rooms = make(map[string]*actor.PID)
	a.stocks = make(map[string]*data.Stock)
	a.gameStocks = make(map[int32]*data.Stock)
	a.newbieStocks = make(map[int32]*data.NewbieStock)
	a.gameScores = make(map[string]*data.CashFlow)
	a.tpStoryChargeStock = make(map[string]*data.TPStoryChargeStock)
	//唯一id初始化
	a.uniqueid = data.InitIDGen(data.ROOMID_KEY)
	glog.Debugf("uniqueid %#v", a.uniqueid)
	a.router = make(map[string]string)
	a.codes = make(map[string]*data.PrivRoom)
	a.rules = make(map[string]string)
	a.count = make(map[string]uint32)
	a.rankWithdrawRobotTimes = make(map[string]int64)
	a.stopCh = make(chan struct{})
	return a
}
