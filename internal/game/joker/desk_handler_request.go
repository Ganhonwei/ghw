package joker

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

func (a *Desk) JOKERCoinEnterRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERCoinEnterRoomReq)
	glog.Debugf("JOKERCoinEnterRoomReq %#v", arg)
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

func (a *Desk) JOKERFreeEnterRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERFreeEnterRoomReq)
	glog.Debugf("JOKERFreeEnterRoomReq %#v", arg)
	userid := a.getRouter(ctx)
	msg1 := a.freeEnterMsg(userid)
	ctx.Respond(msg1)
	a.freeCameinMsg(userid)
}

func (a *Desk) JOKERFreeDealerReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERFreeDealerReq)
	glog.Debugf("JOKERFreeDealerReq %#v", arg)
	userid := a.getRouter(ctx)
	var state int32 = arg.GetState()
	var num uint32 = arg.GetCoin()
	errcode := a.beDealer(userid, state, num)
	if errcode == pb.OK {
		return
	}
	//响应
	rsp := new(pb.JOKERFreeDealerRsp)
	rsp.Error = errcode
	ctx.Respond(rsp)
}

func (a *Desk) JOKERFreeDealerListReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERFreeDealerListReq)
	glog.Debugf("JOKERFreeDealerListReq %#v", arg)
	rsp := a.dealerListMsg()
	ctx.Respond(rsp)
}

func (a *Desk) JOKERSitReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERSitReq)
	glog.Debugf("JOKERSitReq %#v", arg)
	userid := a.getRouter(ctx)
	rsp := a.freeSit(userid, arg)
	if rsp.Error == pb.OK {
		a.broadcast(rsp)
		return
	}
	ctx.Respond(rsp)
}

func (a *Desk) JOKERFreeBetReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERFreeBetReq)
	glog.Debugf("JOKERFreeBetReq %#v", arg)
	userid := a.getRouter(ctx)
	var seatBet uint32 = arg.GetSeat()
	var val uint32 = arg.GetValue()
	errcode := a.freeBet(userid, seatBet, int64(val))
	if errcode == pb.OK {
		return
	}
	//响应
	rsp := new(pb.JOKERFreeBetRsp)
	rsp.Error = errcode
	ctx.Respond(rsp)
}

func (a *Desk) JOKERFreeTrendReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERFreeTrendReq)
	glog.Debugf("JOKERFreeTrendReq %#v", arg)
	rsp := a.freeTrends()
	ctx.Respond(rsp)
}

func (a *Desk) JOKERFreeWinersReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERFreeWinersReq)
	glog.Debugf("JOKERFreeWinersReq %#v", arg)
	rsp := a.freeWiners()
	ctx.Respond(rsp)
}

func (a *Desk) JOKERFreeRolesReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERFreeRolesReq)
	glog.Debugf("JOKERFreeRolesReq %#v", arg)
	rsp := a.freeRoles()
	ctx.Respond(rsp)
}

func (a *Desk) JOKEREnterRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKEREnterRoomReq)
	glog.Debugf("JOKEREnterRoomReq %#v", arg)
	userid := a.getRouter(ctx)
	msg1 := a.privEnterMsg(userid)
	ctx.Respond(msg1)
	a.coinCameinMsg(userid)
}

func (a *Desk) JOKERLeaveReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERLeaveReq)
	glog.Debugf("JOKERLeaveReq %#v", arg)
	userid := a.getRouter(ctx)
	a.nnLeave(userid, ctx)

	// 离开中人机标记
	if a.robotLeaving != nil && a.robotLeaving[userid] {
		delete(a.robotLeaving, userid)
	}
}

func (a *Desk) JOKERReady2Req(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERReady2Req)
	glog.Debugf("JOKERReady2Req %#v", arg)
	userid := a.getRouter(ctx)
	a.readying2(userid)
	// if rsp.Error == pb.OK {
	// 	return
	// }
}

func (a *Desk) JOKERGameRecordReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERGameRecordReq)
	glog.Debugf("JOKERGameRecordReq %#v", arg)
	//TODO
}
func (a *Desk) JOKERLaunchVoteReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERLaunchVoteReq)
	glog.Debugf("JOKERLaunchVoteReq %#v", arg)
	userid := a.getRouter(ctx)
	rsp := a.launchVote(userid, 1)
	if rsp.Error == pb.OK {
		return
	}
	ctx.Respond(rsp)
}

func (a *Desk) JOKERVoteReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERVoteReq)
	glog.Debugf("JOKERVoteReq %#v", arg)
	userid := a.getRouter(ctx)
	var vote uint32 = arg.GetVote()
	rsp := a.privVote(userid, vote)
	if rsp.Error == pb.OK {
		return
	}
	ctx.Respond(rsp)
}

func (a *Desk) JOKERCoinSeeReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERCoinSeeReq)
	glog.Debugf("JOKERCoinSeeReq %#v", arg)
	userid := a.getRouter(ctx)
	a.coinSee(userid)
}

func (a *Desk) JOKERCoinCallReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERCoinCallReq)
	glog.Debugf("JOKERCoinCallReq %#v", arg)
	userid := a.getRouter(ctx)
	a.coinCall(userid)
}

func (a *Desk) JOKERCoinRaiseReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERCoinRaiseReq)
	glog.Debugf("JOKERCoinRaiseReq %#v", arg)
	userid := a.getRouter(ctx)
	a.coinRaise(userid)
}

func (a *Desk) JOKERCoinFoldReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERCoinFoldReq)
	glog.Debugf("JOKERCoinFoldReq %#v", arg)
	userid := a.getRouter(ctx)
	a.coinFold(userid)
}

func (a *Desk) JOKERCoinBiReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERCoinBiReq)
	glog.Debugf("JOKERCoinBiReq %#v", arg)
	userid := a.getRouter(ctx)
	a.coinBi(userid)
}

func (a *Desk) JOKERCoinReplyBiReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERCoinReplyBiReq)
	glog.Debugf("JOKERCoinReplyBiReq %#v", arg)
	userid := a.getRouter(ctx)
	a.coinReplyBi(userid, arg.Agree)
}

func (a *Desk) JOKERCoinChangeRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERCoinChangeRoomReq)
	glog.Debugf("JOKERCoinChangeRoomReq %#v", arg)
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
