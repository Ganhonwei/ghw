package lottery

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// Handler 消息处理
func (a *DeskActor) Handler(msg interface{}, ctx actor.Context) {
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
	case *pb.Ping:
		ctx.Respond(&pb.Pong{Msg: "ok"})
	default:
		//glog.Errorf("unknown message %v", msg)
		a.handlerMsg(msg, ctx)
	}
}

func (a *DeskActor) handlerMsg(msg interface{}, ctx actor.Context) {
	switch msg.(type) {
	case *pb.Connected:
		//连接成功
		arg := msg.(*pb.Connected)
		glog.Infof("Connected %s", arg.Name)
	case *pb.Disconnected:
		//成功断开
		arg := msg.(*pb.Disconnected)
		glog.Infof("Disconnected %s", arg.Name)
	case *pb.CloseDesk:
		arg := msg.(*pb.CloseDesk)
		glog.Debugf("CloseDesk %#v", arg)
		//移除
		delete(a.desks, arg.Roomid)
		delete(a.rules, arg.Unique)
	case *pb.LeaveDesk:
		arg := msg.(*pb.LeaveDesk)
		glog.Debugf("LeaveDesk %#v", arg)
		if v, ok := a.desks[arg.Roomid]; ok &&
			v.Number > 0 {
			v.Number--
			delete(a.userDesks, arg.Userid)
			glog.Debugf("LeaveDesk, user:%s", arg, arg.Userid)
		}
	case *pb.JoinDesk:
		arg := msg.(*pb.JoinDesk)
		glog.Debugf("JoinDesk %#v", arg)
		//房间数据变更
		if v, ok := a.desks[arg.Roomid]; ok {
			v.Number++
		}
		a.userDesks[arg.Userid] = arg.Roomid
	case *pb.EnterDesk:
		arg := msg.(*pb.EnterDesk)
		glog.Debugf("EnterDesk %#v", arg.Sender)
		a.enterDesk(arg, ctx)
	case *pb.SyncConfig:
		//同步配置
		arg := msg.(*pb.SyncConfig)
		// glog.Debugf("SyncConfig %#v", arg)
		a.syncDesk(arg, ctx)
	case *pb.CloseServer:
		a.CloseServer(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}
