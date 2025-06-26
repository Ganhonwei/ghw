package robot

import (
	"goserver/gen/pb"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/myactor"
	"goserver/pkg/utils"
	"sort"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (a *RobotActor) start(ctx actor.Context) {
	glog.Infof("robotgate start: %v", ctx.Self().String())
	myactor.InitLogger(cfg)
	//dbms
	bind := cfg.Section("dbms").Key("bind").Value()
	kind := cfg.Section("dbms").Key("kind").Value()
	// logger := cfg.Section("dbms").Key("logger").Value()
	room := cfg.Section("dbms").Key("room").Value()
	a.dbmsPid = actor.NewPID(bind, kind)
	// a.loggerPid = actor.NewPID(bind, logger)
	a.roomPid = actor.NewPID(bind, room)
	glog.Infof("a.dbmsPid: %s", a.dbmsPid.String())
	// glog.Infof("a.loggerPid: %s", a.loggerPid.String())
	//连接
	connect := &pb.Connect{
		Name: a.Name,
	}
	a.dbmsPid.Request(connect, ctx.Self())
	go a.ticker()
	go a.ticker100Mill()
}

func (a *RobotActor) handlerStop(ctx actor.Context) {
	glog.Debugf("handlerStop: %s", a.Name)
	//关闭
	a.closeTick()
	for _, r := range a.rules {
		r.Pid.Tell(new(pb.ServeClose))
	}
	for _, p := range a.idleRoles {
		p.Tell(new(pb.ServeClose))
	}
	//延迟
	<-time.After(10 * time.Second)
	//断开处理
	msg := &pb.Disconnect{
		Name: a.Name,
	}
	if a.dbmsPid != nil {
		a.dbmsPid.Tell(msg)
	}
	//延迟
	<-time.After(3 * time.Second)
}

// 关闭时钟
func (a *RobotActor) closeTick() {
	select {
	case <-a.stopCh:
		return
	default:
		//停止发送消息
		close(a.stopCh)
	}
}

func (a *RobotActor) ticker100Mill() {
	tick := time.Tick(time.Millisecond * 100)
	msg := new(pb.Tick2)
	for {
		select {
		case <-a.stopCh:
			glog.Info("robot node ticker100Mill closed")
			return
		default:
		}
		select {
		case <-a.stopCh:
			glog.Info("robot node ticker100Mill closed")
			return
		case <-tick:
			nodePid.Tell(msg)
		}
	}
}

func (a *RobotActor) ticker() {
	tick := time.Tick(time.Second * 30)
	msg := new(pb.Tick)
	for {
		select {
		case <-a.stopCh:
			glog.Info("robot node ticker closed")
			return
		default:
		}
		select {
		case <-a.stopCh:
			glog.Info("robot node ticker closed")
			return
		case <-tick:
			nodePid.Tell(msg)
		}
	}
}

// 30秒同步一次
func (rs *RobotActor) ding(ctx actor.Context) {
	glog.Infof("robot num:%d", len(rs.rules))
	glog.Infof("idle robot num:%d", len(rs.idleRoles))
	rs.timer++
	switch rs.timer {
	case 20:
		if node == "1" {
			cleanHead("man", time.Now().Unix())
			cleanHead("woman", time.Now().Unix())
			glog.Infof("clean head")
		}
		rs.timer = 0
	}
}

// 100ms同步一次
func (rs *RobotActor) ding2(ctx actor.Context) {
	if len(rs.CallQeuee) <= 0 {
		return
	}
	sort.Slice(rs.CallQeuee, func(i, j int) bool {
		return rs.CallQeuee[i].CallTime < rs.CallQeuee[j].CallTime
	})

	removeIndex := -1
	now := utils.BsonNow().UnixMilli()
	for i, c := range rs.CallQeuee {
		if now < c.CallTime {
			break
		}
		arg := &pb.RobotMsg{
			Rtype:  c.Msg.Rtype,
			Ltype:  c.Msg.Ltype,
			Gameid: c.Msg.Gameid,
			Roomid: c.Msg.Roomid,
			EnvBet: c.Msg.EnvBet,
			Min:    c.Msg.Min,
			Max:    c.Msg.Max,
			Gtype:  c.Msg.Gtype,
			Emoji:  c.Msg.Emoji,
		}
		rs.callRobot(arg)
		removeIndex = i
	}

	if removeIndex >= 0 {
		rs.CallQeuee = append(make([]CallRobot, 0), rs.CallQeuee[removeIndex+1:]...)
	}
}

func (rs *RobotActor) SyncConfig(ctx actor.Context) {
	msg := ctx.Message()
	//同步配置
	arg := msg.(*pb.SyncConfig)
	handler.SyncConfig(arg, rs.Name)
}
