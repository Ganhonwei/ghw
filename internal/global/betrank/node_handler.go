package betrank

import (
	"goserver/gen/pb"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (a *BetRankActor) Handler(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.ServeStart:
		a.ServeStart(ctx)
	case *pb.ServeStop:
		a.ServeStop(ctx)
	case *pb.Connected:
		a.Connected(ctx)
	case *pb.Disconnected:
		a.Disconnected(ctx)
	case *pb.SyncConfig:
		a.SyncConfig(ctx)
	case *pb.Tick:
		a.ding(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// 启动服务
func (a *BetRankActor) ServeStart(ctx actor.Context) {
	glog.Infof("betrank start: %v", ctx.Self().String())

	//dbms
	// bind := cfg.Section("dbms").Key("bind").Value()
	// kind := cfg.Section("dbms").Key("kind").Value()
	// room := cfg.Section("dbms").Key("room").Value()
	// role := cfg.Section("dbms").Key("role").Value()
	// logger := cfg.Section("dbms").Key("logger").Value()
	// a.dbmsPid = actor.NewPID(bind, kind)
	// a.roomPid = actor.NewPID(bind, room)
	// a.rolePid = actor.NewPID(bind, role)
	// a.loggerPid = actor.NewPID(bind, logger)
	// glog.Infof("a.dbmsPid: %s", a.dbmsPid.String())
	// glog.Infof("a.roomPid: %s", a.roomPid.String())
	// glog.Infof("a.rolePid: %s", a.rolePid.String())
	// glog.Infof("a.loggerPid: %s", a.loggerPid.String())
	// todo 连接
	// connect := &pb.Connect{Name: a.Name}
	// a.dbmsPid.Request(connect, ctx.Self())

	//启动
	go a.ticker(ctx)
}

// 停止服务
func (a *BetRankActor) ServeStop(ctx actor.Context) {
	glog.Debugf("handlerStop: %s", a.Name)
	//关闭
	a.closeTick()
	//响应
	rsp := new(pb.ServeStoped)
	ctx.Respond(rsp)
}

func (a *BetRankActor) Connected(ctx actor.Context) {
	msg := ctx.Message()
	//连接成功
	arg := msg.(*pb.Connected)
	glog.Infof("Connected %s", arg.Name)
}

func (a *BetRankActor) Disconnected(ctx actor.Context) {
	msg := ctx.Message()
	//成功断开
	arg := msg.(*pb.Disconnected)
	glog.Infof("Disconnected %s", arg.Name)
}

// 关闭时钟
func (a *BetRankActor) closeTick() {
	select {
	case <-a.stopCh:
		return
	default:
		//停止发送消息
		close(a.stopCh)
	}
}

func (a *BetRankActor) ticker(ctx actor.Context) {
	tick := time.Tick(1 * time.Second)
	msg := new(pb.Tick)
	for {
		select {
		case <-a.stopCh:
			glog.Info("ticker closed")
			return
		case <-tick:
			ctx.Self().Tell(msg)
		}
	}
}

// 滴答
func (a *BetRankActor) ding(_ actor.Context) {
	// 排行榜派奖检查
	handleBetRankRewordTick(time.Now().In(location))
	// 排行榜更新通知
	handleUpdateBetRankNtfsTick(time.Now().Unix())

	switch a.timer {
	case 9:
		// 人机滴答
		a.betrankRobotTick10s(time.Now().In(location))
		a.timer = 0
	default:
		a.timer++
	}
}

func (a *BetRankActor) SyncConfig(ctx actor.Context) {
	arg := ctx.Message().(*pb.SyncConfig)
	if arg.Type != pb.CONFIG_RELOAD {
		return
	}
	//更新配置
	err := handler.SyncConfig(arg, a.Name)
	if err != nil {
		glog.Error(err)
	}
}
