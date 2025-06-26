package gate

import (
	"context"
	"strconv"
	"time"

	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/data/mq"
	"goserver/pkg/glog"
	"goserver/pkg/myactor"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (a *GateActor) Connected(ctx actor.Context) {
	//连接成功
	arg := ctx.Message().(*pb.Connected)
	glog.Infof("Connected %s", arg.Name)
}

func (a *GateActor) Disconnected(ctx actor.Context) {
	//成功断开
	arg := ctx.Message().(*pb.Disconnected)
	glog.Infof("Disconnected %s", arg.Name)
}

// 启动服务
func (a *GateActor) ServeStart(ctx actor.Context) {
	glog.Infof("gate start: %v", ctx.Self().String())
	myactor.InitLogger(cfg)
	//dbms
	bind := cfg.Section("dbms").Key("bind").Value()
	kind := cfg.Section("dbms").Key("kind").Value()
	room := cfg.Section("dbms").Key("room").Value()
	role := cfg.Section("dbms").Key("role").Value()
	// logger := cfg.Section("dbms").Key("logger").Value()
	// report := cfg.Section("dbms").Key("report").Value()
	a.dbmsPid = actor.NewPID(bind, kind)
	a.roomPid = actor.NewPID(bind, room)
	a.rolePid = actor.NewPID(bind, role)
	// a.loggerPid = actor.NewPID(bind, logger)
	// a.reportPid = actor.NewPID(bind, report)
	connect := &pb.Connect{
		Name: a.Name,
	}
	a.dbmsPid.Request(connect, ctx.Self())
	glog.Infof("a.dbmsPid: %s", a.dbmsPid.String())
	glog.Infof("a.roomPid: %s", a.roomPid.String())
	glog.Infof("a.rolePid: %s", a.rolePid.String())
	// glog.Infof("a.loggerPid: %s", a.loggerPid.String())
	// glog.Infof("a.reportPid: %s", a.reportPid.String())
	//启动
	go a.ticker(ctx)
	// 响应
	rsp := new(pb.ServeStarted)
	ctx.Respond(rsp)

	// 大富翁的奖牌
	a.loadScratchPoker()

	// 消息队列订阅
	a.natsConsumerInit(ctx)
}

func (a *GateActor) loadScratchPoker() {
	a.resetScratch() // 重置检测

	value, err := client.LRange(context.Background(), data.SCRATCHTICKET_POKER, 0, -1).Result()
	if err != nil {
		panic("loadScratchPoker:" + err.Error())
	}
	version, err := client.Get(context.Background(), data.SCRATCHTICKET_VERSION).Result()
	if err != nil {
		panic("loadScratchPoker:" + err.Error())
	}

	scratchTick.Poker = make([]uint32, 0)
	for i, v := range value {
		if i > 4 {
			break
		}
		poker, _ := strconv.Atoi(v)
		scratchTick.Poker = append(scratchTick.Poker, uint32(poker))
	}
	v, _ := strconv.Atoi(version)
	scratchTick.Version = int32(v)
}

// 订阅nats主题
func (a *GateActor) natsConsumerInit(gateCtx actor.Context) {
	consumerName := "gate"

	// 打码排行榜更新订阅
	mq.NatsCreateConsumer(consumerName, mq.StreamActivity, mq.TopicActivityBetRankUpdate, func() *pb.PublishActivityBetRankUpdate { return new(pb.PublishActivityBetRankUpdate) },
		func(arg *pb.PublishActivityBetRankUpdate) (err error) {
			gateCtx.Self().Tell(arg)
			return
		})

	// 订阅支付渠道更新
	mq.NatsCreateConsumer(consumerName, mq.StreamGame, mq.TopicPaychannelUpdate, func() *pb.PublishPayChannelUpdate { return new(pb.PublishPayChannelUpdate) },
		func(arg *pb.PublishPayChannelUpdate) (err error) {
			glog.Infof("---------------TopicPaychannelUpdate: %#v", arg)
			gateCtx.Self().Tell(arg)
			return
		})

	// 订阅客服回复消息
	mq.NatsCreateConsumer(consumerName, mq.StreamGame, mq.TopicGameCustomerReply, func() *pb.ConsumerReplyMessage { return new(pb.ConsumerReplyMessage) },
		func(msg *pb.ConsumerReplyMessage) (err error) {
			gateCtx.Self().Tell(msg)
			return
		})
	// 订阅客服结束对话
	mq.NatsCreateConsumer(consumerName, mq.StreamGame, mq.TopicGameCustomerSessionOver, func() *pb.PublishCustomerSessionOver { return new(pb.PublishCustomerSessionOver) },
		func(msg *pb.PublishCustomerSessionOver) (err error) {
			gateCtx.Self().Tell(msg)
			return
		})
	// 订阅客服撤回消息
	mq.NatsCreateConsumer(consumerName, mq.StreamGame, mq.TopicGameCustomerRevoke, func() *pb.PublishCustomerMessageRevoke { return new(pb.PublishCustomerMessageRevoke) },
		func(msg *pb.PublishCustomerMessageRevoke) (err error) {
			gateCtx.Self().Tell(msg)
			return
		})
}

// 时钟
func (a *GateActor) ticker(ctx actor.Context) {
	tick := time.Tick(30 * time.Second)
	msg := new(pb.Tick)
	for {
		select {
		case <-a.stopCh:
			glog.Info("gate ticker closed")
			return
		default: //防止阻塞
		}
		select {
		case <-a.stopCh:
			glog.Info("gate ticker closed")
			return
		case <-tick:
			ctx.Self().Tell(msg)
		}
	}
}

// 钟声
func (a *GateActor) Tick(ctx actor.Context) {
	//glog.Debugf("ding: %v", ctx.Self().String())

	//下线离线玩家
	a.offlineStop(ctx)
	//大富翁
	a.resetScratch()
}

// 关闭时钟
func (a *GateActor) closeTick() {
	select {
	case <-a.stopCh:
		return
	default:
		//停止发送消息
		close(a.stopCh)
	}
}

// 关闭服务
func (a *GateActor) ServeStop(ctx actor.Context) {
	glog.Debugf("handlerStop: %s", a.Name)
	//关闭
	a.closeTick()
	//关闭消息
	msg1 := &pb.OfflineStop{Ltype: pb.LOGOUT_TYPE2}
	for k, v := range a.offline {
		glog.Debugf("Stop offline role: %s", k)
		v.Pid.Tell(msg1)
	}
	//关闭消息
	msg2 := new(pb.ServeClose)
	for k, v := range a.online {
		glog.Debugf("Stop role: %s", k)
		v.Pid.Tell(msg1)
		v.Pid.Tell(msg2)
	}
	//延迟
	<-time.After(3 * time.Second)
	//断开处理
	msg := &pb.Disconnect{
		Name: a.Name,
	}
	if a.dbmsPid != nil {
		a.dbmsPid.Request(msg, ctx.Self())
	}
	//延迟
	<-time.After(2 * time.Second)
	//响应登录
	rsp := new(pb.ServeStoped)
	ctx.Respond(rsp)
}

// 停服踢人
func (a *GateActor) CloseServer(ctx actor.Context) {
	arg := ctx.Message().(*pb.CloseServer)
	for _, p := range a.online {
		p.Pid.Tell(arg)
	}
}
