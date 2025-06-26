package ak47

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
	case *pb.AK47CoinEnterRoomReq:
		a.AK47CoinEnterRoomReq(ctx)
	case *pb.AK47FreeEnterRoomReq:
		a.AK47FreeEnterRoomReq(ctx)
	case *pb.AK47FreeDealerReq:
		a.AK47FreeDealerReq(ctx)
	case *pb.AK47FreeDealerListReq:
		a.AK47FreeDealerListReq(ctx)
	case *pb.AK47SitReq:
		a.AK47SitReq(ctx)
	case *pb.AK47FreeBetReq:
		a.AK47FreeBetReq(ctx)
	case *pb.AK47FreeTrendReq:
		a.AK47FreeTrendReq(ctx)
	case *pb.AK47FreeWinersReq:
		a.AK47FreeWinersReq(ctx)
	case *pb.AK47FreeRolesReq:
		a.AK47FreeRolesReq(ctx)
	//case *pb.AK47RoomListReq:
	//	arg := msg.(*pb.AK47RoomListReq)
	//	glog.Debugf("AK47RoomListReq %#v", arg)
	//	//TODO
	//case *pb.AK47CreateRoomReq:
	//	arg := msg.(*pb.AK47CreateRoomReq)
	//	glog.Debugf("AK47CreateRoomReq %#v", arg)
	//	//TODO
	case *pb.AK47EnterRoomReq:
		a.AK47EnterRoomReq(ctx)
	case *pb.AK47LeaveReq:
		a.AK47LeaveReq(ctx)
	// case *pb.AK47ReadyReq:
	// 	arg := msg.(*pb.AK47ReadyReq)
	// 	glog.Debugf("AK47ReadyReq %#v", arg)
	// 	userid := a.getRouter(ctx)
	// 	var ready bool = arg.GetReady()
	// 	rsp := a.readying(userid, ready)
	// 	if rsp.Error == pb.OK {
	// 		return
	// 	}
	// 	ctx.Respond(rsp)
	case *pb.AK47Ready2Req:
		a.AK47Ready2Req(ctx)
	case *pb.AK47GameRecordReq:
		a.AK47GameRecordReq(ctx)
	case *pb.AK47LaunchVoteReq:
		a.AK47LaunchVoteReq(ctx)
	case *pb.AK47VoteReq:
		a.AK47VoteReq(ctx)
	case *pb.AK47CoinSeeReq:
		a.AK47CoinSeeReq(ctx)
	case *pb.AK47CoinCallReq:
		a.AK47CoinCallReq(ctx)
	case *pb.AK47CoinRaiseReq:
		a.AK47CoinRaiseReq(ctx)
	case *pb.AK47CoinFoldReq:
		a.AK47CoinFoldReq(ctx)
	case *pb.AK47CoinBiReq:
		a.AK47CoinBiReq(ctx)
	case *pb.AK47CoinReplyBiReq:
		a.AK47CoinReplyBiReq(ctx)
	case *pb.AK47CoinChangeRoomReq:
		a.AK47CoinChangeRoomReq(ctx)
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
