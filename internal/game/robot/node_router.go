package robot

import (
	"goserver/gen/pb"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// Handler 消息处理
func (a *RobotActor) Handler(msg interface{}, ctx actor.Context) {
	switch msg.(type) {
	case *pb.ServeStop:
		//关闭服务
		a.handlerStop(ctx)
		//响应登录
		rsp := new(pb.ServeStoped)
		ctx.Respond(rsp)
	case *pb.ServeStart:
		a.start(ctx)
		//响应
		//rsp := new(pb.ServeStarted)
		//ctx.Respond(rsp)
	case *pb.Tick:
		a.ding(ctx)
	case *pb.Tick2:
		a.ding2(ctx)
	case *pb.Ping:
		ctx.Respond(new(pb.Pong))
	default:
		//glog.Errorf("unknown message %v", msg)
		a.handlerMsg(msg, ctx)
	}
}

func (a *RobotActor) handlerMsg(msg interface{}, ctx actor.Context) {
	switch msg.(type) {
	case *pb.SyncConfig:
		a.SyncConfig(ctx)
	case *pb.RobotMsg:
		a.RobotMsg(ctx)
	case *pb.RobotLeave:
		a.RobotLeave(ctx)
	case *pb.RobotBaseGet:
		a.RobotBaseGet(ctx)
	}
}
