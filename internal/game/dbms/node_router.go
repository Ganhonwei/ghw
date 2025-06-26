package dbms

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/gogo/protobuf/proto"
)

// Handler 消息处理
func (a *DBMSActor) Handler(msg interface{}, ctx actor.Context) {
	switch msg.(type) {
	case *pb.Connected:
		a.Connected(ctx)
	case *pb.Disconnected:
		a.Disconnected(ctx)
	case *pb.Connect:
		a.Connect(ctx)
	case *pb.Disconnect:
		a.Disconnect(ctx)
	case *pb.ServeStop:
		a.ServeStop(ctx)
	case *pb.ServeStart:
		a.ServeStart(ctx)
	case *pb.Tick:
		a.Tick(ctx)
	case *pb.Ping:
		a.Ping(ctx)
	case proto.Message:
		a.handlerUser(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// 玩家请求处理
func (a *DBMSActor) handlerUser(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.GetConfig:
		a.GetConfig(ctx)
	case *pb.SyncConfig:
		a.SyncConfig(ctx)
	case *pb.WebRequest:
		a.WebRequest(ctx)
	case *pb.MatchDesk:
		a.MatchDesk(ctx)
	case *pb.CreateDesk:
		a.CreateDesk(ctx)
	case *pb.GetRoomList:
		a.GetRoomList(ctx)
	// case *pb.GetRoomRecord:
	// 	a.GetRoomRecord(ctx)
	case *pb.RobotMsg:
		a.RobotMsg(ctx)
	case *pb.GamePeopleNtf:
		a.GamePeopleNtf(ctx)
	case *pb.SendSmsCode:
		a.SendSmsCode(ctx)
	case *pb.RobotBaseGet:
		a.RobotBaseGet(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}
