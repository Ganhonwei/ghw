package dbms

import (
	"strconv"
	"time"

	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (a *RoomActor) Connected(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.Connected)
	glog.Infof("Connected %s", arg.Name)
}

func (a *RoomActor) Disconnected(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.Disconnected)
	glog.Infof("Disconnected %s", arg.Name)
}

func (a *RoomActor) ServeStop(ctx actor.Context) {
	//关闭服务
	a.handlerStop(ctx)
	//响应登录
	rsp := new(pb.ServeStoped)
	ctx.Respond(rsp)
}

func (a *RoomActor) ServeStart(ctx actor.Context) {
	a.start(ctx)

	//响应
	//rsp := new(pb.ServeStarted)
	//ctx.Respond(rsp)
}

// 启动服务
func (a *RoomActor) start(ctx actor.Context) {
	glog.Infof("room start: %v", ctx.Self().String())
	//读取库存数据
	stocks := data.GetStockList()
	for _, v := range stocks {
		a.stocks[v.Id] = v
	}
	// 游戏库存
	stocks = data.GetGameStockList()
	for _, v := range stocks {
		id, _ := strconv.Atoi(v.Id)
		a.gameStocks[int32(id)] = v
	}
	//读取新手库存数据
	newbieStocks := data.GetNewbieStockList()
	for _, v := range newbieStocks {
		a.newbieStocks[v.Id] = v
	}
	//读取房间资源
	flows := data.GetCashFlowList()
	for _, v := range flows {
		a.gameScores[v.Id] = v
	}
	//初始化剧情模式局内充值平台亏损
	tpStocks := data.GetTPStoryChargeStockList()
	for _, stock := range tpStocks {
		a.tpStoryChargeStock[stock.Id] = stock
	}
	//启动
	go a.ticker(ctx)
	a.refreshRoomNum()
}

// 时钟
func (a *RoomActor) ticker(ctx actor.Context) {
	tick := time.Tick(5 * time.Second)
	msg := new(pb.Tick)
	for {
		select {
		case <-a.stopCh:
			glog.Info("room ticker closed")
			return
		default: //防止阻塞
		}
		select {
		case <-a.stopCh:
			glog.Info("room ticker closed")
			return
		case <-tick:
			ctx.Self().Tell(msg)
		}
	}
}

// 钟声
func (a *RoomActor) Tick(ctx actor.Context) {
	//glog.Debugf("ding: %v", ctx.Self().String())
	//每5tick一次
	go a.SaveStock() //5s保存一次

	switch a.timer {
	case 60: //5m
		go a.SaveHistory()
		a.refreshRoomNum()
		a.timer = 0
	// case 6: //30s
	// a.SaveStock()
	// a.timer = 0
	// ctx.Self().Tell(&pb.GetRoomFactor{})
	// ctx.Self().Tell(&pb.ChangeStock{})
	// ctx.Self().Tell(&pb.ModifyStock{})
	default:
		a.timer++
	}

	// 保存提现排行榜
	// a.RankWithdrawTime()
}

// 关闭时钟
func (a *RoomActor) closeTick() {
	select {
	case <-a.stopCh:
		return
	default:
		//停止发送消息
		close(a.stopCh)
	}
}

func (a *RoomActor) handlerStop(ctx actor.Context) {
	glog.Debugf("handlerStop: %s", a.Name)
	//关闭
	a.closeTick()
	//回存数据
	if a.uniqueid != nil {
		a.uniqueid.Save()
	}
	for k := range a.rooms {
		glog.Debugf("Stop room: %s", k)
		//TODO
	}
}
