package fortune_gems2

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// Handler 消息处理
func (a *DeskActor) Handler(msg interface{}, ctx actor.Context) {
	switch msg.(type) {
	case *pb.ServeStop:
		a.ServeStop(ctx)
	case *pb.ServeStart:
		a.ServeStart(ctx)
	case *pb.Tick:
		a.Tick(ctx)
	case *pb.Ping:
		ctx.Respond(&pb.Pong{Msg: "ok"})
	default:
		//glog.Errorf("unknown message %v", msg)
		a.handlerMsg(ctx)
	}
}

func (a *DeskActor) handlerMsg(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.Connected:
		a.Connected(ctx)
	case *pb.Disconnected:
		a.Disconnected(ctx)
	case *pb.CloseDesk:
		a.CloseDesk(ctx)
	case *pb.LeaveDesk:
		a.LeaveDesk(ctx)
	case *pb.JoinDesk:
		a.JoinDesk(ctx)
	case *pb.EnterDesk:
		a.EnterDesk(ctx)
	// case *pb.CreateDesk:
	// 	a.CreateDesk(ctx)
	case *pb.SyncConfig:
		a.SyncConfig(ctx)
	case *pb.GetRoomList:
		a.GetRoomList(ctx)
	case *pb.ChangeDesk:
		a.ChangeDesk(ctx)
	case *pb.EarlyLeave:
		a.EarlyLeave(ctx)
	case *pb.ClearLeave:
		a.ClearLeave(ctx)
	case *pb.GameStart:
		a.GameStart(ctx)
	case *pb.CloseServer:
		a.CloseServer(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}
