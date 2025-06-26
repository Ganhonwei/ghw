package robot

import (
	"goserver/gen/pb"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (a *RoleActor) Handler(msg interface{}, ctx actor.Context) {
	switch msg.(type) {
	case *pb.ServeStop:
		//关闭服务
		a.handlerStop(ctx)
	case *pb.ServeStart:
		a.ServeStart(ctx)
		//响应
		//rsp := new(pb.ServeStarted)
		//ctx.Respond(rsp)
	case *pb.ServeClose:
		a.ServeClose(ctx)
	case *pb.Tick:
		a.ding(ctx)
	default:
		//glog.Errorf("unknown message %v", msg)
		a.handlerDesk(msg, ctx)
	}
}

func (a *RoleActor) handlerDesk(msg interface{}, ctx actor.Context) {
	switch msg.(type) {
	case *pb.RobotMsg:
		a.RobotMsg(ctx)
	case *pb.EnteredDesk:
		a.EnteredDesk(ctx)
	case *pb.MatchedDesk:
		a.MatchedDesk(ctx)
	case *pb.PushCurrencyNtf:
		a.PushCurrencyNtf(ctx)
	case *pb.ChargeInGameNtf:
		a.ChargeInGameNtf(ctx)
	case *pb.ChargeInGameFinishNtf:
		a.ChargeInGameFinishNtf(ctx)
	default:
		a.handlerLeave(msg, ctx)
	}
}

// 离开消息
func (a *RoleActor) handlerLeave(msg interface{}, ctx actor.Context) {
	switch msg.(type) {
	case *pb.LHLeaveNtf:
		a.LHLeaveNtf(ctx)
	case *pb.LHLeaveRsp:
		a.LHLeaveRsp(ctx)
	case *pb.UPLeaveNtf:
		a.UPLeaveNtf(ctx)
	case *pb.UPLeaveRsp:
		a.UPLeaveRsp(ctx)
	case *pb.CRASHLeaveNtf:
		a.CRASHLeaveNtf(ctx)
	case *pb.CRASHLeaveRsp:
		a.CRASHLeaveRsp(ctx)
	case *pb.ABLeaveNtf:
		a.ABLeaveNtf(ctx)
	case *pb.ABLeaveRsp:
		a.ABLeaveRsp(ctx)
	case *pb.JHLeaveNtf:
		a.JHLeaveNtf(ctx)
	case *pb.JHLeaveRsp:
		a.JHLeaveRsp(ctx)
	case *pb.AK47LeaveNtf:
		a.AK47LeaveNtf(ctx)
	case *pb.AK47LeaveRsp:
		a.AK47LeaveRsp(ctx)
	case *pb.JOKERLeaveNtf:
		a.JOKERLeaveNtf(ctx)
	case *pb.JOKERLeaveRsp:
		a.JOKERLeaveRsp(ctx)
	case *pb.LotteryLeaveNtf:
		a.LotteryLeaveNtf(ctx)
	case *pb.LotteryLeaveRsp:
		a.LotteryLeaveRsp(ctx)
	case *pb.PLANELeaveNtf:
		a.PLANELeaveNtf(ctx)
	case *pb.PLANELeaveRsp:
		a.PLANELeaveRsp(ctx)
	case *pb.RobotLeaveNtf:
		a.RobotLeaveNtf(ctx)
	default:
		a.handlerHUA(msg, ctx)
	}
}

func (a *RoleActor) handlerHUA(msg interface{}, ctx actor.Context) {
	switch msg.(type) {
	case *pb.JHCoinEnterRoomRsp:
		a.recvJHCoinEnterRoomRsp(ctx)
	case *pb.JHPushStateNtf:
		a.recvJHPushStateNtf(ctx)
	case *pb.JHPushActStateNtf:
		a.recvJHPushActStateNtf(ctx)
	case *pb.JHCoinReplyBiNtf:
		a.recvJHCoinReplyBiNtf(ctx)
	case *pb.JHCoinRobotStrategyNtf:
		a.JHCoinRobotStrategyNtf(ctx)
	case *pb.JHCoinRobotStrategyInfoNtf:
		a.JHCoinRobotStrategyInfoNtf(ctx)
	case *pb.JHCoinGameoverNtf:
		a.JHCoinGameoverNtf(ctx)
	case *pb.JHCoinSeeNtf:
		a.JHCoinSeeNtf(ctx)
	case *pb.JHRobotSeeInOtherNtf:
		a.JHRobotSeeInOtherNtf(ctx)
	case *pb.JHCoinCallNtf:
		a.JHCoinCallNtf(ctx)
	case *pb.JHCoinRaiseNtf:
		a.JHCoinRaiseNtf(ctx)
	case *pb.JHCoinBiNtf:
		a.recvJHCoinBiNtf(ctx)
	case *pb.JHCoinWaitTooLongNtf:
		a.recvJHCoinWaitTooLongNtf(ctx)
	case *pb.JHCoinSeeRsp:
		a.JHCoinSeeRsp(ctx)
	default:
		a.handlerAK47(msg, ctx)
	}
}

func (a *RoleActor) handlerAK47(msg interface{}, ctx actor.Context) {
	switch msg.(type) {
	case *pb.AK47CoinEnterRoomRsp:
		a.recvAK47CoinEnterRoomRsp(ctx)
	case *pb.AK47PushStateNtf:
		a.recvAK47PushStateNtf(ctx)
	case *pb.AK47PushActStateNtf:
		a.recvAK47PushActStateNtf(ctx)
	case *pb.AK47CoinRobotStrategyNtf:
		a.AK47CoinRobotStrategyNtf(ctx)
	case *pb.AK47CoinGameoverNtf:
		a.AK47CoinGameoverNtf(ctx)
	case *pb.AK47CoinSeeNtf:
		a.AK47CoinSeeNtf(ctx)
	case *pb.AK47CoinRaiseNtf:
		a.AK47CoinRaiseNtf(ctx)
	case *pb.AK47CoinBiNtf:
		a.recvAK47CoinBiNtf(ctx)
	case *pb.AK47CoinWaitTooLongNtf:
		a.AK47CoinWaitTooLongNtf(ctx)
	case *pb.AK47CoinSeeRsp:
		a.AK47CoinSeeRsp(ctx)
	default:
		a.handlerJoker(msg, ctx)
	}
}

// joker
func (a *RoleActor) handlerJoker(msg interface{}, ctx actor.Context) {
	switch msg.(type) {
	case *pb.JOKERCoinEnterRoomRsp:
		a.recvJOKERCoinEnterRoomRsp(ctx)
	case *pb.JOKERPushStateNtf:
		a.recvJOKERPushStateNtf(ctx)
	case *pb.JOKERPushActStateNtf:
		a.recvJOKERPushActStateNtf(ctx)
	case *pb.JOKERCoinRobotStrategyNtf:
		a.JOKERCoinRobotStrategyNtf(ctx)
	case *pb.JOKERCoinGameoverNtf:
		a.JOKERCoinGameoverNtf(ctx)
	case *pb.JOKERCoinSeeNtf:
		a.JOKERCoinSeeNtf(ctx)
	case *pb.JOKERCoinRaiseNtf:
		a.JOKERCoinRaiseNtf(ctx)
	case *pb.JOKERCoinBiNtf:
		a.recvJOKERCoinBiNtf(ctx)
	case *pb.JOKERCoinWaitTooLongNtf:
		a.JOKERCoinWaitTooLongNtf(ctx)
	case *pb.JOKERCoinSeeRsp:
		a.JOKERCoinSeeRsp(ctx)
	default:
		a.handlerRummy(msg, ctx)
	}
}

// rummy
func (a *RoleActor) handlerRummy(msg interface{}, ctx actor.Context) {
	switch msg.(type) {
	//rm
	case *pb.RMCoinEnterRoomRsp:
		a.recvRMCoinEnterRoomRsp(ctx)
	case *pb.RMPushStateNtf:
		a.recvRMPushStateNtf(ctx)
	case *pb.RMPushActStateNtf:
		a.recvRMPushActStateNtf(ctx)
	case *pb.RMPushDealerNtf:
		a.recvRMPushDealerNtf(ctx)
	case *pb.RMDrawCardNtf:
		a.recvRMDrawCardNtf(ctx)
	case *pb.RMPushCardsNtf:
		a.recvRMPushCardsNtf(ctx)
	case *pb.RMDiscardNtf:
		a.recvRMDiscardNtf(ctx)
	case *pb.RMDrawCardRsp:
		a.recvRMDrawCardRsp(ctx)
	case *pb.RMDiscardRsp:
		a.recvRMDiscardRsp(ctx)
	case *pb.RMFinishNtf:
		a.recvRMFinishNtf(ctx)
	case *pb.RMCoinGameoverNtf:
		a.recvRMCoinGameoverNtf(ctx)
	case *pb.RMLeaveRsp:
		a.recvRMLeaveRsp(ctx)
	case *pb.RMCoinWaitTooLongNtf:
		a.RMCoinWaitTooLongNtf(ctx)
	default:
		a.handlerLHD(msg, ctx)
	}
}

// lhd
func (a *RoleActor) handlerLHD(msg interface{}, ctx actor.Context) {
	switch msg.(type) {
	case *pb.LHFreeEnterRoomRsp:
		a.LHFreeEnterRoomRsp(ctx)
	case *pb.LHRobotStrategyNtf:
		a.recvLHRobotStrategy(ctx)
	case *pb.LHPushStateNtf:
		a.recvLHState(ctx)
	default:
		a.handlerUP(msg, ctx)
	}
}

// 7up
func (a *RoleActor) handlerUP(msg interface{}, ctx actor.Context) {
	switch msg.(type) {
	case *pb.UPFreeEnterRoomRsp:
		a.UPFreeEnterRoomRsp(ctx)
	case *pb.UPRobotStrategyNtf:
		a.recvUPRobotStrategy(ctx)
	case *pb.UPPushStateNtf:
		a.recvUPState(ctx)
	default:
		a.handlerCrash(msg, ctx)
	}
}

// crash
func (a *RoleActor) handlerCrash(msg interface{}, ctx actor.Context) {
	switch msg.(type) {
	case *pb.CRASHEnterRoomRsp:
		a.CRASHEnterRoomRsp(ctx)
	case *pb.CRASHRobotStrategyNtf:
		a.recvCRASHRobotStrategy(ctx)
	case *pb.CRASHPushStateNtf:
		a.recvCRASHState(ctx)
	case *pb.CRASHBoomNtf:
		a.CRASHBoomNtf(ctx)
	default:
		a.handlerPlane(msg, ctx)
	}
}

// plane
func (a *RoleActor) handlerPlane(msg interface{}, ctx actor.Context) {
	switch msg.(type) {
	case *pb.PLANEEnterRoomRsp:
		a.PLANEEnterRoomRsp(ctx)
	case *pb.PLANERobotStrategyNtf:
		a.recvPLANERobotStrategy(ctx)
	case *pb.PLANEPushStateNtf:
		a.recvPLANEState(ctx)
	case *pb.PLANEBoomNtf:
		a.PLANEBoomNtf(ctx)
	default:
		a.handlerAB(msg, ctx)
	}
}

// andarbahar
func (a *RoleActor) handlerAB(msg interface{}, ctx actor.Context) {
	switch msg.(type) {
	case *pb.ABFreeEnterRoomRsp:
		a.ABFreeEnterRoomRsp(ctx)
	case *pb.ABRobotStrategyNtf:
		a.recvABRobotStrategy(ctx)
	case *pb.ABPushStateNtf:
		a.recvABState(ctx)
	default:
		a.handlerLottery(msg, ctx)
	}
}

// 彩票
func (a *RoleActor) handlerLottery(msg interface{}, ctx actor.Context) {
	switch msg.(type) {
	case *pb.LotteryEnterRsp:
		a.LotteryEnterRsp(ctx)
	case *pb.CPRobotStrategyNtf:
		a.recvCPRobotStrategy(ctx)
	case *pb.LotteryPushStateNtf:
		a.recvCPState(ctx)
	default:
		// a.handlerCrash(msg, ctx)
	}
}
