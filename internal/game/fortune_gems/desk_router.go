package fortune_gems

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/gogo/protobuf/proto"
)

// Handler 消息处理
func (a *Desk) Handler(msg interface{}, ctx actor.Context) {
	switch msg.(type) {
	case *pb.ServeStop:
		a.ServeStop(ctx)
	case *pb.ServeStart:
		a.ServeStart(ctx)
	case *pb.Tick:
		a.Tick(ctx)
	case *pb.CloseServer:
		a.CloseServer(ctx)
	default:
		//glog.Errorf("unknown message %v", msg)
		a.handlerLogic(ctx)
	}
}

func (a *Desk) handlerLogic(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.CloseDesk:
		a.CloseDesk(ctx)
	case *pb.LeaveDesk:
		a.LeaveDesk(ctx)
	case *pb.SyncConfig:
		a.SyncConfig(ctx)
	case *pb.PrintDesk:
		a.PrintDesk(ctx)
	case *pb.EnterDesk:
		a.EnterDesk(ctx)
	case *pb.OfflineDesk:
		a.OfflineDesk(ctx)
	case *pb.ChangeCurrency:
		a.ChangeCurrency(ctx)
	case *pb.AddBlackList:
		a.AddBlackList(ctx)
	case proto.Message:
		//请求消息
		a.handlerRequest(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// 玩家请求
func (a *Desk) handlerRequest(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.FortuneGemsEnterRoomReq:
		a.FortuneGemsEnterRoomReq(ctx)
	case *pb.FortuneGemsBetReq:
		a.FortuneGemsBetReq(ctx)
	case *pb.FortuneGemsLeaveReq:
		a.FortuneGemsLeaveReq(ctx)
	// case *pb.FortuneGemsPitStepReq:
	// a.FortuneGemsPitStepReq(ctx)
	// case *pb.FortuneGemsCashOutReq:
	// a.FortuneGemsCashOutReq(ctx)
	// case *pb.FortuneGemsAutoBetCancelReq:
	// a.FortuneGemsAutoBetCancelReq(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}
