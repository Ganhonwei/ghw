package rm2

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
	case *pb.UserGameState:
		a.UserGameState(ctx)
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
	case *pb.RMCoinEnterRoomReq:
		a.RMCoinEnterRoomReq(ctx)
	case *pb.RMLeaveReq:
		a.RMLeaveReq(ctx)
	case *pb.RMReady2Req:
		a.RMReady2Req(ctx)
	case *pb.RMDrawCardReq:
		a.RMDrawCardReq(ctx)
	case *pb.RMDiscardReq:
		a.RMDiscardReq(ctx)
	case *pb.RMFinishReq:
		a.RMFinishReq(ctx)
	case *pb.RMDeclareReq:
		a.RMDeclareReq(ctx)
	case *pb.RMSortReq:
		a.RMSortReq(ctx)
	case *pb.RMDropReq:
		a.RMDropReq(ctx)
	case *pb.RMDropScoreReq:
		a.RMDropScoreReq(ctx)
	// case *pb.RMFreeEnterRoomReq:
	// 	a.RMFreeEnterRoomReq(ctx)
	// case *pb.RMFreeDealerReq:
	// 	a.RMFreeDealerReq(ctx)
	// case *pb.RMFreeDealerListReq:
	// 	a.RMFreeDealerListReq(ctx)
	// case *pb.RMSitReq:
	// 	a.RMSitReq(ctx)
	// case *pb.RMFreeBetReq:
	// 	a.RMFreeBetReq(ctx)
	// case *pb.RMFreeTrendReq:
	// 	a.RMFreeTrendReq(ctx)
	// case *pb.RMFreeWinersReq:
	// 	a.RMFreeWinersReq(ctx)
	// case *pb.RMFreeRolesReq:
	// 	a.RMFreeRolesReq(ctx)
	//case *pb.RMRoomListReq:
	//	arg := msg.(*pb.RMRoomListReq)
	//	glog.Debugf("RMRoomListReq %#v", arg)
	//	//TODO
	//case *pb.RMCreateRoomReq:
	//	arg := msg.(*pb.RMCreateRoomReq)
	//	glog.Debugf("RMCreateRoomReq %#v", arg)
	//	//TODO
	case *pb.RMEnterRoomReq:
		a.RMEnterRoomReq(ctx)
	case *pb.RMKickoutRoomReq:
		a.RMKickoutRoomReq(ctx)
	case *pb.RMChangeSeatReq:
		a.RMChangeSeatReq(ctx)
	case *pb.RMGameStartReq:
		a.RMGameStartReq(ctx)
	case *pb.PrivLaunchAgainReq:
		a.PrivLaunchAgainReq(ctx)
	case *pb.PrivAgainReq:
		a.PrivAgainReq(ctx)
	// case *pb.RMReadyReq:
	// 	arg := msg.(*pb.RMReadyReq)
	// 	glog.Debugf("RMReadyReq %#v", arg)
	// 	userid := a.getRouter(ctx)
	// 	var ready bool = arg.GetReady()
	// 	rsp := a.readying(userid, ready)
	// 	if rsp.Error == pb.OK {
	// 		return
	// 	}
	// 	ctx.Respond(rsp)
	// case *pb.RMGameRecordReq:
	// 	a.RMGameRecordReq(ctx)
	case *pb.RMLaunchVoteReq:
		a.RMLaunchVoteReq(ctx)
	case *pb.RMVoteReq:
		a.RMVoteReq(ctx)
	case *pb.RMAutoSortReq:
		a.RMAutoSortReq(ctx)
	case *pb.RMQiCardsReq:
		a.RMQiCardsReq(ctx)
	case *pb.RMRoiRecordsync:
		a.RMRoiRecordsync(ctx)
	// case *pb.RMCoinSeeReq:
	// 	a.RMCoinSeeReq(ctx)
	// case *pb.RMCoinCallReq:
	// 	a.RMCoinCallReq(ctx)
	// case *pb.RMCoinRaiseReq:
	// 	a.RMCoinRaiseReq(ctx)
	// case *pb.RMCoinFoldReq:
	// 	a.RMCoinFoldReq(ctx)
	// case *pb.RMCoinBiReq:
	// 	a.RMCoinBiReq(ctx)
	// case *pb.RMCoinReplyBiReq:
	// 	a.RMCoinReplyBiReq(ctx)
	// case *pb.RMCoinChangeRoomReq:
	// 	a.RMCoinChangeRoomReq(ctx)
	// case *pb.BankGive:
	// 	a.BankGive(ctx)
	// case *pb.PointControl:
	// 	a.PointControl(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}
