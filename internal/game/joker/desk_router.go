package joker

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
	case *pb.GameRechargeAmount:
		a.GameRechargeAmount(ctx)
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
	case *pb.ChatTextReq:
		a.ChatTextReq(ctx)
	// case *pb.ChatVoiceReq:
	// 	a.ChatVoiceReq(ctx)
	case *pb.JOKERCoinEnterRoomReq:
		a.JOKERCoinEnterRoomReq(ctx)
	case *pb.JOKERFreeEnterRoomReq:
		a.JOKERFreeEnterRoomReq(ctx)
	case *pb.JOKERFreeDealerReq:
		a.JOKERFreeDealerReq(ctx)
	case *pb.JOKERFreeDealerListReq:
		a.JOKERFreeDealerListReq(ctx)
	case *pb.JOKERSitReq:
		a.JOKERSitReq(ctx)
	case *pb.JOKERFreeBetReq:
		a.JOKERFreeBetReq(ctx)
	case *pb.JOKERFreeTrendReq:
		a.JOKERFreeTrendReq(ctx)
	case *pb.JOKERFreeWinersReq:
		a.JOKERFreeWinersReq(ctx)
	case *pb.JOKERFreeRolesReq:
		a.JOKERFreeRolesReq(ctx)
	//case *pb.JOKERRoomListReq:
	//	arg := msg.(*pb.JOKERRoomListReq)
	//	glog.Debugf("JOKERRoomListReq %#v", arg)
	//	//TODO
	//case *pb.JOKERCreateRoomReq:
	//	arg := msg.(*pb.JOKERCreateRoomReq)
	//	glog.Debugf("JOKERCreateRoomReq %#v", arg)
	//	//TODO
	case *pb.JOKEREnterRoomReq:
		a.JOKEREnterRoomReq(ctx)
	case *pb.JOKERLeaveReq:
		a.JOKERLeaveReq(ctx)
	// case *pb.JOKERReadyReq:
	// 	arg := msg.(*pb.JOKERReadyReq)
	// 	glog.Debugf("JOKERReadyReq %#v", arg)
	// 	userid := a.getRouter(ctx)
	// 	var ready bool = arg.GetReady()
	// 	rsp := a.readying(userid, ready)
	// 	if rsp.Error == pb.OK {
	// 		return
	// 	}
	// 	ctx.Respond(rsp)
	case *pb.JOKERReady2Req:
		a.JOKERReady2Req(ctx)
	case *pb.JOKERGameRecordReq:
		a.JOKERGameRecordReq(ctx)
	case *pb.JOKERLaunchVoteReq:
		a.JOKERLaunchVoteReq(ctx)
	case *pb.JOKERVoteReq:
		a.JOKERVoteReq(ctx)
	case *pb.JOKERCoinSeeReq:
		a.JOKERCoinSeeReq(ctx)
	case *pb.JOKERCoinCallReq:
		a.JOKERCoinCallReq(ctx)
	case *pb.JOKERCoinRaiseReq:
		a.JOKERCoinRaiseReq(ctx)
	case *pb.JOKERCoinFoldReq:
		a.JOKERCoinFoldReq(ctx)
	case *pb.JOKERCoinBiReq:
		a.JOKERCoinBiReq(ctx)
	case *pb.JOKERCoinReplyBiReq:
		a.JOKERCoinReplyBiReq(ctx)
	case *pb.JOKERCoinChangeRoomReq:
		a.JOKERCoinChangeRoomReq(ctx)
	case *pb.BankGive:
		a.BankGive(ctx)
	case *pb.PointControl:
		a.PointControl(ctx)
	case *pb.UserGameState:
		a.UserGameState(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}
