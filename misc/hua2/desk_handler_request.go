package main

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

func (a *Desk) JHCoinEnterRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHCoinEnterRoomReq)
	glog.Debugf("JHCoinEnterRoomReq %#v", arg)
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

func (a *Desk) JHFreeEnterRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHFreeEnterRoomReq)
	glog.Debugf("JHFreeEnterRoomReq %#v", arg)
	userid := a.getRouter(ctx)
	msg1 := a.freeEnterMsg(userid)
	ctx.Respond(msg1)
	a.freeCameinMsg(userid)
}

func (a *Desk) JHFreeDealerReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHFreeDealerReq)
	glog.Debugf("JHFreeDealerReq %#v", arg)
	userid := a.getRouter(ctx)
	var state int32 = arg.GetState()
	var num uint32 = arg.GetCoin()
	errcode := a.beDealer(userid, state, num)
	if errcode == pb.OK {
		return
	}
	//响应
	rsp := new(pb.JHFreeDealerRsp)
	rsp.Error = errcode
	ctx.Respond(rsp)
}

func (a *Desk) JHFreeDealerListReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHFreeDealerListReq)
	glog.Debugf("JHFreeDealerListReq %#v", arg)
	rsp := a.dealerListMsg()
	ctx.Respond(rsp)
}

func (a *Desk) JHSitReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHSitReq)
	glog.Debugf("JHSitReq %#v", arg)
	userid := a.getRouter(ctx)
	rsp := a.freeSit(userid, arg)
	if rsp.Error == pb.OK {
		a.broadcast(rsp)
		return
	}
	ctx.Respond(rsp)
}

func (a *Desk) JHFreeBetReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHFreeBetReq)
	glog.Debugf("JHFreeBetReq %#v", arg)
	userid := a.getRouter(ctx)
	var seatBet uint32 = arg.GetSeat()
	var val uint32 = arg.GetValue()
	errcode := a.freeBet(userid, seatBet, int64(val))
	if errcode == pb.OK {
		return
	}
	//响应
	rsp := new(pb.JHFreeBetRsp)
	rsp.Error = errcode
	ctx.Respond(rsp)
}

func (a *Desk) JHFreeTrendReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHFreeTrendReq)
	glog.Debugf("JHFreeTrendReq %#v", arg)
	rsp := a.freeTrends()
	ctx.Respond(rsp)
}

func (a *Desk) JHFreeWinersReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHFreeWinersReq)
	glog.Debugf("JHFreeWinersReq %#v", arg)
	rsp := a.freeWiners()
	ctx.Respond(rsp)
}

func (a *Desk) JHFreeRolesReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHFreeRolesReq)
	glog.Debugf("JHFreeRolesReq %#v", arg)
	rsp := a.freeRoles()
	ctx.Respond(rsp)
}

func (a *Desk) JHEnterRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHEnterRoomReq)
	glog.Debugf("JHEnterRoomReq %#v", arg)
	userid := a.getRouter(ctx)
	msg1 := a.privEnterMsg(userid)
	ctx.Respond(msg1)
	a.coinCameinMsg(userid)

	//房间状态检查
	if a.state == int32(pb.STATE_FREE) && a.roleNum() >= 2 {
		a.state = int32(pb.STATE_READY)
		a.pushState() //广播状态变更
	}
}

func (a *Desk) JHLeaveReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHLeaveReq)
	glog.Debugf("JHLeaveReq %#v", arg)
	userid := a.getRouter(ctx)

	if a.Rtype == int32(pb.ROOM_TYPE1) {
		a.chargeInGameCancel(userid)

		// 私人房再来一句投票时退出算放弃
		if a.DeskPriv.AgainSeat != 0 && a.DeskPriv.Again == 0 {
			a.privAgain(userid, 2)
		}
		// 房主提前退出返还房费
		if userid == a.DeskData.Cid && a.isPrivFunRoom() {
			a.backCost()
		}
	}

	a.nnLeave(userid, ctx)

	// 离开中人机标记
	if a.robotLeaving != nil && a.robotLeaving[userid] {
		delete(a.robotLeaving, userid)
	}
}

func (a *Desk) JHReady2Req(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHReady2Req)
	glog.Debugf("JHReady2Req %#v", arg)
	userid := a.getRouter(ctx)
	a.readying2(userid)
	// if rsp.Error == pb.OK {
	// 	return
	// }
}

// 私人房手动开始游戏
func (a *Desk) JHGameStartReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHGameStartReq)
	glog.Debugf("JHGameStartReq %#v", arg)
	res := new(pb.JHGameStartRsp)

	userid := a.getRouter(ctx)
	if a.Cid != userid {
		res.Error = pb.NotRoomCreator
		ctx.Respond(res)
		return
	}
	// 准备人数>2
	if 2 > len(a.seats) {
		res.Error = pb.NotReady
		ctx.Respond(res)
		return
	}
	// 对局结束解散倒计时中
	if a.dismissOnRoundOver || a.settleExitTime != 0 {
		res.Error = pb.RoomMaintenance
		ctx.Respond(res)
		return
	}
	a.gameStart()
	ctx.Respond(res)
}

func (a *Desk) JHGameRecordReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHGameRecordReq)
	glog.Debugf("JHGameRecordReq %#v", arg)
	//TODO
}
func (a *Desk) JHLaunchVoteReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHLaunchVoteReq)
	glog.Debugf("JHLaunchVoteReq %#v", arg)
	userid := a.getRouter(ctx)
	rsp := new(pb.JHLaunchVoteRsp)

	// 房间只有一个人解散房间
	if len(a.seats) == 1 {
		rsp.Error = pb.LeaveEarly
		ctx.Respond(rsp)
		msg1 := new(pb.ServeStop)
		a.selfPid.Tell(msg1)
		return
	}
	// 已投票再来一句
	if a.DeskPriv.AgainSeat != 0 || a.DeskPriv.Again != 0 {
		rsp.Error = pb.RoomMaintenance
		return
	}
	// 游戏未开始,发起者非房主,直接退出
	if a.Cid != userid && a.state == int32(pb.STATE_READY) && a.DeskGame.Round == 0 {
		rsp.Error = pb.LeaveEarly
		ctx.Respond(rsp)

		a.notifyNodeLeaveEarly(a.Rid, userid, pb.LeaveEarly)
		a.notifyGateUserLeft(userid, pb.LeaveEarly, int32(pb.LeaveEarly))
		//清除数据
		a.userLeaveDesk(userid)
		return
	}
	// 发起投票
	rsp = a.launchVote(userid, 1)
	if rsp.Error == pb.OK {
		return
	}
	ctx.Respond(rsp)
}

func (a *Desk) JHVoteReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHVoteReq)
	glog.Debugf("JHVoteReq %#v", arg)
	userid := a.getRouter(ctx)
	var vote uint32 = arg.GetVote()
	rsp := a.privVote(userid, vote)
	if rsp.Error == pb.OK {
		return
	}
	ctx.Respond(rsp)
}

func (a *Desk) JHCoinSeeReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHCoinSeeReq)
	glog.Debugf("JHCoinSeeReq %#v", arg)
	userid := a.getRouter(ctx)
	a.coinSee(userid)
}

func (a *Desk) JHCoinCallReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHCoinCallReq)
	glog.Debugf("JHCoinCallReq %#v", arg)
	userid := a.getRouter(ctx)
	a.coinCall(userid)
}

func (a *Desk) JHCoinRaiseReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHCoinRaiseReq)
	glog.Debugf("JHCoinRaiseReq %#v", arg)
	userid := a.getRouter(ctx)
	a.coinRaise(userid)
}

func (a *Desk) JHCoinFoldReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHCoinFoldReq)
	glog.Debugf("JHCoinFoldReq %#v", arg)
	userid := a.getRouter(ctx)
	a.coinFold(userid)
}

func (a *Desk) JHCoinBiReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHCoinBiReq)
	glog.Debugf("JHCoinBiReq %#v", arg)
	userid := a.getRouter(ctx)
	a.coinBi(userid)
}

func (a *Desk) JHCoinReplyBiReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHCoinReplyBiReq)
	glog.Debugf("JHCoinReplyBiReq %#v", arg)
	userid := a.getRouter(ctx)
	a.coinReplyBi(userid, arg.Agree)
}

func (a *Desk) JHCoinChangeRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHCoinChangeRoomReq)
	glog.Debugf("JHCoinChangeRoomReq %#v", arg)
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
	}
}

// 重置tp触发次数
func (a *Desk) TPTriggeTimes(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.TPTriggeTimes)
	glog.Debugf("TPTriggeTimes %#v", arg)
	userid := a.getRouter(ctx)
	role := a.getRole(userid)
	if role == nil {
		return
	}
	if arg.Isreset {
		role.ResetTPTiggerTimes()
	}
}

// 对战房房主踢人
func (t *Desk) JHKickoutRoomReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.JHKickoutRoomReq)
	rsp := new(pb.JHKickoutRoomRsp)
	// 游戏中不能踢人
	if t.IsGaming() {
		rsp.Error = pb.GameStarted
		ctx.Respond(rsp)
		return
	}
	// 非房主
	userId := t.getRouter(ctx)
	if t.Cid != userId {
		rsp.Error = pb.NotRoomCreator
		ctx.Respond(rsp)
		return
	}
	if _, ok := t.roles[arg.Userid]; !ok {
		rsp.Error = pb.NotInPrivateRoom
		ctx.Respond(rsp)
		return
	}

	// 通知玩家被踢出了
	msg2 := &pb.KickoutedRoom{
		Gameid: t.Game.Id,
		Roomid: t.DeskData.Rid,
		Gtype:  t.Gtype,
		Rtype:  t.Rtype,
	}
	t.send2userid(arg.Userid, msg2)

	// 清除玩家
	rsp.KickoutSeat = t.getSeatid(arg.Userid)
	//玩家离开牌桌
	t.notifyNodeLeaveEarly(t.Rid, arg.Userid, pb.OK)
	t.notifyGateUserLeft(arg.Userid, pb.OK, int32(pb.OK))
	//清除数据
	t.userLeaveDesk(arg.Userid)
	ctx.Respond(rsp)
}

// 对战房换座位
func (t *Desk) JHChangeSeatReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.JHChangeSeatReq)
	rsp := new(pb.JHChangeSeatRsp)
	if t.IsGaming() {
		rsp.Error = pb.GameStarted
		ctx.Respond(rsp)
		return
	}
	if arg.ToSeat > t.Count {
		rsp.Error = pb.SeatNotExists
		ctx.Respond(rsp)
		return
	}
	if _, ok := t.seats[arg.ToSeat]; ok {
		rsp.Error = pb.SeatTaken
		ctx.Respond(rsp)
		return
	}
	rsp.Userid = t.getRouter(ctx)
	rsp.Seat = t.getSeatid(rsp.Userid)
	rsp.ToSeat = arg.ToSeat

	t.seats[rsp.ToSeat] = t.seats[rsp.Seat]
	delete(t.seats, rsp.Seat)
	t.getRole(rsp.Userid).Seat = rsp.ToSeat

	glog.Infof("user %s change seat: %d to %d", rsp.Userid, rsp.Seat, rsp.ToSeat)
	// t.broadcast(rsp)

	// 通知所有玩家换位置
	for userid, role := range t.roles {
		rsp.Userinfo = t.coinSeatBetsMsg(userid)
		role.Pid.Tell(rsp)
	}
}

// 私人房牌局结束再来一局发起投票
func (t *Desk) PrivLaunchAgainReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.PrivLaunchAgainReq)
	glog.Debugf("PrivLaunchAgainReq %#v", arg)
	rsp := new(pb.PrivLaunchAgainRsp)

	// 房间只有一个人解散房间
	if len(t.seats) == 1 {
		rsp.Error = pb.LeaveEarly
		ctx.Respond(rsp)
		msg1 := new(pb.ServeStop)
		t.selfPid.Tell(msg1)
		return
	}

	// 即将解散，或投票解散中
	if t.dismissOnRoundOver || t.DeskPriv.VoteSeat != 0 {
		rsp.Error = pb.RoomMaintenance
		ctx.Respond(rsp)
		return
	}

	// 娱乐分房主余额不足
	if t.isPrivFunRoom() && t.DeskData.Cost > 0 {
		masterRole, ok := t.roles[t.DeskData.Cid]
		if !ok {
			rsp.Error = pb.NotRoomCreator
			ctx.Respond(rsp)
			return
		}
		if masterRole.GetScore() < int64(t.DeskData.Cost) {
			rsp.Error = pb.NotEnoughDiamond
			ctx.Respond(rsp)
			return
		}
	}

	// 再次开始游戏 todo投票
	// t.InitDesk()
	// t.DeskData.Ctime = uint32(utils.Timestamp())
	// t.DeskData.Expire = utils.Timestamp() + int64(config.GetPvpRoom().GameStartWait)
	// t.AgainRound++
	// t.gameStart()
	// rsp.Userid = t.getRouter(ctx)
	// rsp.Seat = t.getSeatid(rsp.Userid)
	// ctx.Respond(rsp)

	userid := t.getRouter(ctx)
	rsp = t.launchAgain(userid)
	if rsp.Error == pb.OK {
		return
	}
	ctx.Respond(rsp)
}

// 私人房牌局结束再来一局投票
func (t *Desk) PrivAgainReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.PrivAgainReq)
	glog.Debugf("PrivAgainReq %#v", arg)
	userid := t.getRouter(ctx)
	rsp := t.privAgain(userid, uint32(arg.Again))
	if rsp.Error == pb.OK {
		return
	}
	ctx.Respond(rsp)
}

// 同步修改用户小号信息
func (t *Desk) ChangeViceAccount(ctx actor.Context) {
	arg := ctx.Message().(*pb.ChangeViceAccount)
	for _, role := range t.roles {
		if role.Userid == arg.Userid {
			role.ViceAccount = arg.ViceAccount
		}
	}
}
