package ak47

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (a *Desk) ChatTextReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ChatTextReq)
	glog.Debugf("ChatTextReq %#v", arg)
	a.chatText(arg, ctx)
}

// func (a *Desk) ChatVoiceReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.ChatVoiceReq)
// 	glog.Debugf("ChatVoiceReq %#v", arg)
// 	a.chatVoice(arg, ctx)
// }

func (a *Desk) AK47CoinEnterRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47CoinEnterRoomReq)
	glog.Debugf("AK47CoinEnterRoomReq %#v", arg)
	userid := a.getRouter(ctx)
	msg1 := a.coinEnterMsg(userid)
	ctx.Respond(msg1)
	a.coinCameinMsg(userid)

	// a.callRobot()

	//房间状态检查
	if a.state == int32(pb.STATE_FREE) && a.roleNum() >= 2 {
		a.state = int32(pb.STATE_READY)
		a.pushState() //广播状态变更
	}
}

func (a *Desk) AK47FreeEnterRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47FreeEnterRoomReq)
	glog.Debugf("AK47FreeEnterRoomReq %#v", arg)
	userid := a.getRouter(ctx)
	msg1 := a.freeEnterMsg(userid)
	ctx.Respond(msg1)
	a.freeCameinMsg(userid)
}

func (a *Desk) AK47FreeDealerReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47FreeDealerReq)
	glog.Debugf("AK47FreeDealerReq %#v", arg)
	userid := a.getRouter(ctx)
	var state int32 = arg.GetState()
	var num uint32 = arg.GetCoin()
	errcode := a.beDealer(userid, state, num)
	if errcode == pb.OK {
		return
	}
	//响应
	rsp := new(pb.AK47FreeDealerRsp)
	rsp.Error = errcode
	ctx.Respond(rsp)
}

func (a *Desk) AK47FreeDealerListReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47FreeDealerListReq)
	glog.Debugf("AK47FreeDealerListReq %#v", arg)
	rsp := a.dealerListMsg()
	ctx.Respond(rsp)
}

func (a *Desk) AK47SitReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47SitReq)
	glog.Debugf("AK47SitReq %#v", arg)
	userid := a.getRouter(ctx)
	rsp := a.freeSit(userid, arg)
	if rsp.Error == pb.OK {
		a.broadcast(rsp)
		return
	}
	ctx.Respond(rsp)
}

func (a *Desk) AK47FreeBetReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47FreeBetReq)
	glog.Debugf("AK47FreeBetReq %#v", arg)
	userid := a.getRouter(ctx)
	var seatBet uint32 = arg.GetSeat()
	var val uint32 = arg.GetValue()
	errcode := a.freeBet(userid, seatBet, int64(val))
	if errcode == pb.OK {
		return
	}
	//响应
	rsp := new(pb.AK47FreeBetRsp)
	rsp.Error = errcode
	ctx.Respond(rsp)
}

func (a *Desk) AK47FreeTrendReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47FreeTrendReq)
	glog.Debugf("AK47FreeTrendReq %#v", arg)
	rsp := a.freeTrends()
	ctx.Respond(rsp)
}

func (a *Desk) AK47FreeWinersReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47FreeWinersReq)
	glog.Debugf("AK47FreeWinersReq %#v", arg)
	rsp := a.freeWiners()
	ctx.Respond(rsp)
}

func (a *Desk) AK47FreeRolesReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47FreeRolesReq)
	glog.Debugf("AK47FreeRolesReq %#v", arg)
	rsp := a.freeRoles()
	ctx.Respond(rsp)
}

func (a *Desk) AK47EnterRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47EnterRoomReq)
	glog.Debugf("AK47EnterRoomReq %#v", arg)
	userid := a.getRouter(ctx)
	msg1 := a.privEnterMsg(userid)
	ctx.Respond(msg1)
	a.coinCameinMsg(userid)
}

func (a *Desk) AK47LeaveReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47LeaveReq)
	glog.Debugf("AK47LeaveReq %#v", arg)
	userid := a.getRouter(ctx)
	a.nnLeave(userid, ctx)
	// 离开中人机标记
	if a.robotLeaving != nil && a.robotLeaving[userid] {
		delete(a.robotLeaving, userid)
	}
}

func (a *Desk) AK47Ready2Req(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47Ready2Req)
	glog.Debugf("AK47Ready2Req %#v", arg)
	userid := a.getRouter(ctx)
	a.readying2(userid)
	// if rsp.Error == pb.OK {
	// 	return
	// }
}

func (a *Desk) AK47GameRecordReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47GameRecordReq)
	glog.Debugf("AK47GameRecordReq %#v", arg)
	//TODO
}
func (a *Desk) AK47LaunchVoteReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47LaunchVoteReq)
	glog.Debugf("AK47LaunchVoteReq %#v", arg)
	userid := a.getRouter(ctx)
	rsp := a.launchVote(userid, 1)
	if rsp.Error == pb.OK {
		return
	}
	ctx.Respond(rsp)
}

func (a *Desk) AK47VoteReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47VoteReq)
	glog.Debugf("AK47VoteReq %#v", arg)
	userid := a.getRouter(ctx)
	var vote uint32 = arg.GetVote()
	rsp := a.privVote(userid, vote)
	if rsp.Error == pb.OK {
		return
	}
	ctx.Respond(rsp)
}

func (a *Desk) AK47CoinSeeReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47CoinSeeReq)
	glog.Debugf("AK47CoinSeeReq %#v", arg)
	userid := a.getRouter(ctx)
	a.coinSee(userid)
}

func (a *Desk) AK47CoinCallReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47CoinCallReq)
	glog.Debugf("AK47CoinCallReq %#v", arg)
	userid := a.getRouter(ctx)
	a.coinCall(userid)
}

func (a *Desk) AK47CoinRaiseReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47CoinRaiseReq)
	glog.Debugf("AK47CoinRaiseReq %#v", arg)
	userid := a.getRouter(ctx)
	a.coinRaise(userid)
}

func (a *Desk) AK47CoinFoldReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47CoinFoldReq)
	glog.Debugf("AK47CoinFoldReq %#v", arg)
	userid := a.getRouter(ctx)
	a.coinFold(userid)
}

func (a *Desk) AK47CoinBiReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47CoinBiReq)
	glog.Debugf("AK47CoinBiReq %#v", arg)
	userid := a.getRouter(ctx)
	a.coinBi(userid)
}

func (a *Desk) AK47CoinReplyBiReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47CoinReplyBiReq)
	glog.Debugf("AK47CoinReplyBiReq %#v", arg)
	userid := a.getRouter(ctx)
	a.coinReplyBi(userid, arg.Agree)
}

func (a *Desk) AK47CoinChangeRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47CoinChangeRoomReq)
	glog.Debugf("AK47CoinChangeRoomReq %#v", arg)
	a.changeDesk(ctx)
}

func (a *Desk) BankGive(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.BankGive)
	glog.Debugf("BankGive %#v", arg)
	if v, ok := a.roles[arg.GetUserid()]; ok && v != nil {
		v.User.AddCoin(arg.GetCoin())
	}
}

func (a *Desk) PointControl(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.PointControl)
	glog.Debugf("PointControl %#v", arg)

	if v, ok := a.roles[arg.UserId]; ok && v != nil {
		data := &data.PointControl{
			UserId: arg.UserId,
			Switch: arg.Switch,
			Factor: arg.Factor,
			Score:  arg.Score,
		}
		handler.PointControl(v.User, data)
	}
}

func (a *Desk) UserGameState(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.UserGameState)
	glog.Debugf("UserGameState %#v", arg)
	role := a.getRole(arg.Userid)
	if role != nil {
		role.State = int(arg.State)
		// a.limitOver()
	}
}
