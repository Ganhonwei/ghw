package andarbahar

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
		//关闭服务
		glog.Infof("ServeStop: %#v", msg)
		a.handlerStop(ctx)
		//响应登录
		//rsp := new(pb.ServeStoped)
		//ctx.Respond(rsp)
	case *pb.ServeStart:
		a.start(ctx)
		//响应
		//rsp := new(pb.ServeStarted)
		//ctx.Respond(rsp)
	case *pb.Tick:
		a.ding(ctx)
	case *pb.Tick2:
		a.newbiewDing(ctx)
	case *pb.CloseServer:
		a.CloseServer(ctx)
	default:
		//glog.Errorf("unknown message %v", msg)
		a.handlerLogic(msg, ctx)
	}
}

func (a *Desk) handlerLogic(msg interface{}, ctx actor.Context) {
	switch msg.(type) {
	case *pb.CloseDesk:
		a.CloseDesk(ctx)
	case *pb.LeaveDesk:
		a.LeaveDesk(ctx)
	case *pb.SyncConfig:
		a.SyncConfig(ctx)
	case *pb.PrintDesk:
		//打印牌局状态信息,test
		a.printOver()
	case *pb.EnterDesk:
		a.EnterDesk(ctx)
	case *pb.OfflineDesk:
		a.OfflineDesk(ctx)
	case *pb.ChangeCurrency:
		a.ChangeCurrency(ctx)
	case *pb.DeskStatus:
		a.DeskStatus(ctx)
	case *pb.UserGameState:
		a.UserGameState(ctx)
	case *pb.PointControl:
		a.PointControl(ctx)
	case *pb.GameRechargeAmount:
		a.GameRechargeAmount(ctx)
	case *pb.TriggerFreeWelfare:
		a.TriggerFreeWelfare(ctx)
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
	case *pb.ABFreeEnterRoomReq:
		a.ABFreeEnterRoomReq(ctx)
	case *pb.ABEnterSuccessReq:
		a.ABEnterSuccessReq(ctx)
	case *pb.ABFreeBetReq:
		a.ABFreeBetReq(ctx)
	case *pb.ABLeaveReq:
		a.ABLeaveReq(ctx)
	case *pb.ABEnterRoomReq:
		a.ABEnterRoomReq(ctx)
	case *pb.ABGameStartReq:
		a.ABGameStartReq(ctx)
	case *pb.ChangeFreeDeskStatus:
		a.ChangeFreeDeskStatus(ctx)
	case *pb.ABKickoutRoomReq:
		a.ABKickoutRoomReq(ctx)
	case *pb.ABChangeSeatReq:
		a.ABChangeSeatReq(ctx)
	case *pb.ABChangeSeatAcceptReq:
		a.ABChangeSeatAcceptReq(ctx)
	case *pb.ABLaunchVoteReq:
		a.ABLaunchVoteReq(ctx)
	case *pb.ABVoteReq:
		a.ABVoteReq(ctx)
	case *pb.ABBetReq:
		a.ABBetReq(ctx)
	case *pb.PrivLaunchAgainReq:
		a.PrivLaunchAgainReq(ctx)
	case *pb.PrivAgainReq:
		a.PrivAgainReq(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}
