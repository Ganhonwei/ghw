package dbms

import (
	"strings"
	"time"

	"goserver/gen/pb"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/myactor"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (a *DBMSActor) Connected(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.Connected)
	glog.Infof("Connected %s", arg.Name)
}

func (a *DBMSActor) Disconnected(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.Disconnected)
	glog.Infof("Disconnected %s", arg.Name)
}

func (a *DBMSActor) Connect(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.Connect)
	//服务注册
	a.serve[arg.Name] = ctx.Sender()
	//响应
	connected := &pb.Connected{
		Name: a.Name,
	}
	ctx.Respond(connected)
	glog.Infof("Connect %s", arg.Name)
	//同步配置到gate,game
	a.syncConfig(arg.Name)
}

func (a *DBMSActor) Disconnect(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.Disconnect)
	//服务注销
	delete(a.serve, arg.Name)
	//响应
	//disconnected := &pb.Disconnected{
	//	Name: a.Name,
	//}
	//ctx.Respond(disconnected)
	glog.Infof("Disconnect %s", arg.Name)
}

func (a *DBMSActor) ServeStop(ctx actor.Context) {
	//关闭服务
	a.handlerStop(ctx)
	//响应登录
	rsp := new(pb.ServeStoped)
	ctx.Respond(rsp)
}

func (a *DBMSActor) ServeStart(ctx actor.Context) {
	a.start(ctx)
	//响应
	//rsp := new(pb.ServeStarted)
	//ctx.Respond(rsp)
}

// 启动服务
func (a *DBMSActor) start(ctx actor.Context) {
	glog.Infof("dbms start: %v", ctx.Self().String())
	myactor.InitLogger(cfg)
	// handler.SetDailySign()
	// handler.SetOnlineReward()
	handler.SetPokerHands()
	// handler.SetDailySign()
	handler.SetTaskList()
	//TODO 设置测试数据,正式后台配置
	/*
		handler.SetActivityList()
		handler.SetTaskList5()
		handler.SetTaskList4()
		handler.SetTaskList3()
		handler.SetEBGCoinGame()
		handler.SetShopList()
		handler.SetNiuCoinGame()
		handler.SetLuckyList()
		//handler.SetGameList()
		handler.SetTaskList()
		handler.SetLoginPrizeList()
		head := cfg.Section("domain").Key("headimag").Value()
		passwd := cfg.Section("robot").Key("passwd").Value()
		phone := cfg.Section("robot").Key("phone").Value()
		rs := data.RegistRobots4(head, passwd, phone)
		for _, v := range rs {
			rolePid.Tell(v)
		}
		list, err := data.GetAgentDayProfit(&pb.AgentDayProfitReq{Selfid:"105757",Page:1})
		glog.Debugf("list %#v, err %v", list, err)*/
	//启动
	go a.ticker(ctx)
	//统计服务
	a.statPid = NewStat()
	statStart(a.statPid)
}

// 时钟
func (a *DBMSActor) ticker(ctx actor.Context) {
	tick := time.Tick(30 * time.Second)
	msg := new(pb.Tick)
	for {
		select {
		case <-a.stopCh:
			glog.Info("dbms ticker closed")
			return
		default: //防止阻塞
		}
		select {
		case <-a.stopCh:
			glog.Info("dbms ticker closed")
			return
		case <-tick:
			ctx.Self().Tell(msg)
		}
	}
}

// 钟声
func (a *DBMSActor) Tick(ctx actor.Context) {
	//glog.Debugf("ding: %v", ctx.Self().String())
	//TODO
}

// 关闭时钟
func (a *DBMSActor) closeTick() {
	select {
	case <-a.stopCh:
		return
	default:
		//停止发送消息
		close(a.stopCh)
	}
}

func (a *DBMSActor) handlerStop(ctx actor.Context) {
	glog.Debugf("handlerStop: %s", a.Name)
	//关闭
	a.closeTick()
	//回存数据
	for k := range a.serve {
		glog.Debugf("Stop gate: %s", k)
	}
	//关闭统计服务
	statStop(a.statPid)
}

// 同步配置
func (a *DBMSActor) syncConfig(key string) {
	if _, ok := a.serve[key]; !ok {
		glog.Errorf("gate not exists: %s", key)
		return
	}
	pid := a.serve[key]
	a.syncConfig2(pid)
}

// 同步配置
func (a *DBMSActor) syncConfig2(pid *actor.PID) {
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_ENV))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_NOTICE))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_SHOP))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_GAMES))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_TASK))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_LOGIN))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_LUCKY))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_ACT))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_SIGN))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_ONLINE))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_WITHDRAW))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_POKERHANDS))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_SWITCH))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_PAY_CHANNEL))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_SHARE))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_SHARE_ADDR))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_BEGINNER))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_IPWHITE))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_WEEKCARD))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_EMOJI))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_RECHARGELIMIT))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_CHANNEL))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_SEVERWHITE))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_CUSTOMERADDR))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_SHARE_WAY))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_LIMITEDGIFT))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_PVP_ROOM))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_VIP))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_RANK_WITHDRAW))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_VIP_ROBOT))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_LUCKY_DRAW_REWORD))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_LUCKY_DRAW_ROBOT))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_DEVICELIMIT))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_SCRATCHTICK))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_PLAYSHARE))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_CRASHSTRATEGY))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_LHDSTRATEGY))
	pid.Tell(handler.GetSyncConfig(pb.CHARGE_CLASSIFY))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_UPSTRATEGY))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_ABSTRATEGY))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_CPSTRATEGY))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_RBSTRATEGY))
	pid.Tell(handler.GetSyncConfig(pb.CONFIG_AVSTRATEGY))
}

func (a *DBMSActor) Ping(ctx actor.Context) {
	res := new(pb.Pong)
	for name, ser := range a.serve {
		if strings.Contains(name, "login") {
			continue
		}
		_, err := ctx.RequestFuture(ser, ctx.Message(), 3*time.Second).Result()
		if err != nil {
			res.Msg = "failed"
			ctx.Respond(res)
			glog.Errorf("Ping err: %s", name)
			return
		}
	}
	res.Msg = "ok"
	ctx.Respond(res)
}
