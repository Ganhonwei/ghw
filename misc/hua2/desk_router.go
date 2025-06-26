package main

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
	case *pb.JHCoinEnterRoomReq:
		a.JHCoinEnterRoomReq(ctx)
	case *pb.JHFreeEnterRoomReq:
		a.JHFreeEnterRoomReq(ctx)
	case *pb.JHFreeDealerReq:
		a.JHFreeDealerReq(ctx)
	case *pb.JHFreeDealerListReq:
		a.JHFreeDealerListReq(ctx)
	case *pb.JHSitReq:
		a.JHSitReq(ctx)
	case *pb.JHFreeBetReq:
		a.JHFreeBetReq(ctx)
	case *pb.JHFreeTrendReq:
		a.JHFreeTrendReq(ctx)
	case *pb.JHFreeWinersReq:
		a.JHFreeWinersReq(ctx)
	case *pb.JHFreeRolesReq:
		a.JHFreeRolesReq(ctx)
	//case *pb.JHRoomListReq:
	//	arg := msg.(*pb.JHRoomListReq)
	//	glog.Debugf("JHRoomListReq %#v", arg)
	//	//TODO
	//case *pb.JHCreateRoomReq:
	//	arg := msg.(*pb.JHCreateRoomReq)
	//	glog.Debugf("JHCreateRoomReq %#v", arg)
	//	//TODO
	case *pb.JHEnterRoomReq:
		a.JHEnterRoomReq(ctx)
	case *pb.JHLeaveReq:
		a.JHLeaveReq(ctx)
	case *pb.JHGameStartReq:
		a.JHGameStartReq(ctx)
	// case *pb.JHReadyReq:
	// 	arg := msg.(*pb.JHReadyReq)
	// 	glog.Debugf("JHReadyReq %#v", arg)
	// 	userid := a.getRouter(ctx)
	// 	var ready bool = arg.GetReady()
	// 	rsp := a.readying(userid, ready)
	// 	if rsp.Error == pb.OK {
	// 		return
	// 	}
	// 	ctx.Respond(rsp)
	case *pb.JHReady2Req:
		a.JHReady2Req(ctx)
	case *pb.JHGameRecordReq:
		a.JHGameRecordReq(ctx)
	case *pb.JHLaunchVoteReq:
		a.JHLaunchVoteReq(ctx)
	case *pb.JHVoteReq:
		a.JHVoteReq(ctx)
	case *pb.JHCoinSeeReq:
		a.JHCoinSeeReq(ctx)
	case *pb.JHCoinCallReq:
		a.JHCoinCallReq(ctx)
	case *pb.JHCoinRaiseReq:
		a.JHCoinRaiseReq(ctx)
	case *pb.JHCoinFoldReq:
		a.JHCoinFoldReq(ctx)
	case *pb.JHCoinBiReq:
		a.JHCoinBiReq(ctx)
	case *pb.JHCoinReplyBiReq:
		a.JHCoinReplyBiReq(ctx)
	case *pb.JHCoinChangeRoomReq:
		a.JHCoinChangeRoomReq(ctx)
	case *pb.BankGive:
		a.BankGive(ctx)
	case *pb.PointControl:
		a.PointControl(ctx)
	case *pb.UserGameState:
		a.UserGameState(ctx)
	case *pb.TPTriggeTimes:
		a.TPTriggeTimes(ctx)
	case *pb.JHKickoutRoomReq:
		a.JHKickoutRoomReq(ctx)
	case *pb.JHChangeSeatReq:
		a.JHChangeSeatReq(ctx)
	case *pb.PrivLaunchAgainReq:
		a.PrivLaunchAgainReq(ctx)
	case *pb.PrivAgainReq:
		a.PrivAgainReq(ctx)
	case *pb.ChangeViceAccount:
		a.ChangeViceAccount(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}
