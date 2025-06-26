package dbms

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/gogo/protobuf/proto"
)

// Handler 房间管理消息处理
func (a *RoomActor) Handler(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.Connected:
		a.Connected(ctx)
	case *pb.Disconnected:
		a.Disconnected(ctx)
	case *pb.ServeStop:
		a.ServeStop(ctx)
	case *pb.ServeStart:
		a.ServeStart(ctx)
	case *pb.Tick:
		a.Tick(ctx)
	case proto.Message:
		a.handlerDesk(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// 请求处理
func (a *RoomActor) handlerDesk(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	//case *pb.GenDesk:
	//	arg := msg.(*pb.GenDesk)
	//	glog.Debugf("GenDesk: %v", arg)
	//	a.genDesk(arg, ctx)
	case *pb.AddDesk:
		a.AddDesk(ctx)
	case *pb.JoinDesk:
		a.JoinDesk(ctx)
	case *pb.LeaveDesk:
		a.LeaveDesk(ctx)
	case *pb.Logout:
		a.Logout(ctx)
	case *pb.CloseDesk:
		a.CloseDesk(ctx)
	case *pb.MatchDesk:
		a.MatchDesk(ctx)
	case *pb.SyncGamePeople:
		a.SyncGamePeople(ctx)
	case *pb.GetGameListReq:
		a.GetGameListReq(ctx)
	case *pb.ShareBetAmount:
		a.ShareBetAmount(ctx)
	case *pb.PrivEnterRoomReq:
		a.PrivEnterRoomReq(ctx)
	// case *pb.RankWithdrawReq:
	// 	a.RankWithdrawReq(ctx)
	case proto.Message:
		a.handlerStock(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

func (a *RoomActor) handlerStock(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.GetRoomFactor:
		a.GetRoomFactor(ctx)
	case *pb.ChangeStock:
		a.ChangeStock(ctx)
	case *pb.ModifyStock:
		a.ModifyStock(ctx)
	case *pb.WebRequest:
		a.WebRequest(ctx)
	case *pb.NewbieStock:
		a.NewbieStock(ctx)
	case *pb.GetGameStock:
		a.GetGameStock(ctx)
	case *pb.GetTpStoryChargeLoss:
		a.GetTpStoryChargeLoss(ctx)
	case *pb.ChangeTpStoryChargeLoss:
		a.ChangeTpStoryChargeLoss(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}
