package rm2

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/algo"
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

func (a *Desk) RMCoinEnterRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMCoinEnterRoomReq)
	glog.Debugf("RMCoinEnterRoomReq %#v", arg)
	userid := a.getRouter(ctx)
	msg1 := a.coinEnterMsg(userid)
	ctx.Respond(msg1)
	a.coinCameinMsg(userid)

	//房间状态检查
	// if a.state == int32(pb.STATE_FREE) && a.roleNum() >= 2 {
	// 	a.state = int32(pb.STATE_READY)
	// 	a.pushState() //广播状态变更
	// }
}

func (a *Desk) RMFreeEnterRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMFreeEnterRoomReq)
	glog.Debugf("RMFreeEnterRoomReq %#v", arg)
	userid := a.getRouter(ctx)
	msg1 := a.freeEnterMsg(userid)
	ctx.Respond(msg1)
	a.freeCameinMsg(userid)
}

func (a *Desk) RMFreeDealerReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMFreeDealerReq)
	glog.Debugf("RMFreeDealerReq %#v", arg)
	userid := a.getRouter(ctx)
	var state int32 = arg.GetState()
	var num uint32 = arg.GetCoin()
	errcode := a.beDealer(userid, state, num)
	if errcode == pb.OK {
		return
	}
	//响应
	rsp := new(pb.RMFreeDealerRsp)
	rsp.Error = errcode
	ctx.Respond(rsp)
}

func (a *Desk) RMFreeDealerListReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMFreeDealerListReq)
	glog.Debugf("RMFreeDealerListReq %#v", arg)
	rsp := a.dealerListMsg()
	ctx.Respond(rsp)
}

func (a *Desk) RMSitReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMSitReq)
	glog.Debugf("RMSitReq %#v", arg)
	userid := a.getRouter(ctx)
	rsp := a.freeSit(userid, arg)
	if rsp.Error == pb.OK {
		a.broadcast(rsp)
		return
	}
	ctx.Respond(rsp)
}

func (a *Desk) RMFreeBetReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMFreeBetReq)
	glog.Debugf("RMFreeBetReq %#v", arg)
	userid := a.getRouter(ctx)
	var seatBet uint32 = arg.GetSeat()
	var val uint32 = arg.GetValue()
	errcode := a.freeBet(userid, seatBet, int64(val))
	if errcode == pb.OK {
		return
	}
	//响应
	rsp := new(pb.RMFreeBetRsp)
	rsp.Error = errcode
	ctx.Respond(rsp)
}

func (a *Desk) RMFreeTrendReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMFreeTrendReq)
	glog.Debugf("RMFreeTrendReq %#v", arg)
	rsp := a.freeTrends()
	ctx.Respond(rsp)
}

func (a *Desk) RMFreeWinersReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMFreeWinersReq)
	glog.Debugf("RMFreeWinersReq %#v", arg)
	rsp := a.freeWiners()
	ctx.Respond(rsp)
}

func (a *Desk) RMFreeRolesReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMFreeRolesReq)
	glog.Debugf("RMFreeRolesReq %#v", arg)
	rsp := a.freeRoles()
	ctx.Respond(rsp)
}

func (a *Desk) RMEnterRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMEnterRoomReq)
	glog.Debugf("RMEnterRoomReq %#v", arg)
	userid := a.getRouter(ctx)
	msg1 := a.privEnterMsg(userid)
	ctx.Respond(msg1)
	a.coinCameinMsg(userid)
}

func (a *Desk) RMLeaveReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMLeaveReq)
	glog.Debugf("RMLeaveReq %#v", arg)
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
}

func (a *Desk) RMReady2Req(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMReady2Req)
	glog.Debugf("RMReady2Req %#v", arg)
	userid := a.getRouter(ctx)
	a.readying2(userid)
	// if rsp.Error == pb.OK {
	// 	return
	// }
}

// 私人房手动开始游戏
func (a *Desk) RMGameStartReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMGameStartReq)
	glog.Debugf("RMGameStartReq %#v", arg)
	res := new(pb.RMGameStartRsp)

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

func (a *Desk) RMDrawCardReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMDrawCardReq)
	glog.Debugf("RMDrawCardReq %#v", arg)
	userid := a.getRouter(ctx)
	a.drawCard(userid, arg.Area, false)
}

func (a *Desk) RMDiscardReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMDiscardReq)
	glog.Debugf("RMDiscardReq %#v", arg)
	userid := a.getRouter(ctx)
	a.discard(userid, arg.Card, false)
}

func (a *Desk) RMFinishReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMFinishReq)
	glog.Debugf("RMFinishReq %#v", arg)
	userid := a.getRouter(ctx)
	a.finish(userid, arg)
}

func (a *Desk) RMDeclareReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMDeclareReq)
	glog.Debugf("RMDeclareReq %#v", arg)
	userid := a.getRouter(ctx)
	a.declare(userid, arg)
}

func (a *Desk) RMSortReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMSortReq)
	glog.Debugf("RMSortReq %#v", arg)
	userid := a.getRouter(ctx)
	a.sort(userid, arg)
}

func (a *Desk) RMDropReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMDropReq)
	glog.Debugf("RMDropReq %#v", arg)
	userid := a.getRouter(ctx)
	a.drop(userid, arg)
}

func (a *Desk) RMDropScoreReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMDropScoreReq)
	glog.Debugf("RMDropScoreReq %#v", arg)
	userid := a.getRouter(ctx)
	a.dropScore(userid, arg)
}

func (a *Desk) RMGameRecordReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMGameRecordReq)
	glog.Debugf("RMGameRecordReq %#v", arg)
	//TODO
}
func (a *Desk) RMLaunchVoteReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMLaunchVoteReq)
	glog.Debugf("RMLaunchVoteReq %#v", arg)
	userid := a.getRouter(ctx)
	rsp := new(pb.RMLaunchVoteRsp)

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

	rsp = a.launchVote(userid, 1)
	if rsp.Error == pb.OK {
		return
	}
	ctx.Respond(rsp)
}

func (a *Desk) RMVoteReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMVoteReq)
	glog.Debugf("RMVoteReq %#v", arg)
	userid := a.getRouter(ctx)
	var vote uint32 = arg.GetVote()
	rsp := a.privVote(userid, vote)
	if rsp.Error == pb.OK {
		return
	}
	ctx.Respond(rsp)
}

// 玩家摆牌
func (a *Desk) RMAutoSortReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMAutoSortReq)
	glog.Debugf("RMAutoSortReq %#v", arg)
	rsp := new(pb.RMAutoSortRsp)
	userid := a.getRouter(ctx)
	seatid := a.getSeatid(userid)
	seat := a.getSeat(seatid)
	if seat == nil {
		rsp.Error = pb.NotInRoom
		ctx.Respond(rsp)
		return
	}
	groups := algo.GroupTheCards(seat.Cards, a.WildCard) // 一般延迟2秒左右
	// groups, lefts := algo.SortCards(seat.Cards, a.WildCard)
	// groups = append(groups, lefts)
	seat.SortCards = groups
	for _, group := range groups {
		rsp.Infos = append(rsp.Infos, &pb.RMSortInfo{Cards: group})
	}
	ctx.Respond(rsp)
}

// 查询弃牌堆
func (a *Desk) RMQiCardsReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMQiCardsReq)
	glog.Debugf("RMQiCardsReq %#v", arg)
	rsp := new(pb.RMQiCardsRsp)
	rsp.QiCards = a.QiCards
	ctx.Respond(rsp)
}

func (a *Desk) RMCoinSeeReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMCoinSeeReq)
	glog.Debugf("RMCoinSeeReq %#v", arg)
	userid := a.getRouter(ctx)
	a.coinSee(userid)
}

func (a *Desk) RMCoinCallReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMCoinCallReq)
	glog.Debugf("RMCoinCallReq %#v", arg)
	userid := a.getRouter(ctx)
	a.coinCall(userid)
}

func (a *Desk) RMCoinRaiseReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMCoinRaiseReq)
	glog.Debugf("RMCoinRaiseReq %#v", arg)
	userid := a.getRouter(ctx)
	a.coinRaise(userid)
}

func (a *Desk) RMCoinFoldReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMCoinFoldReq)
	glog.Debugf("RMCoinFoldReq %#v", arg)
	userid := a.getRouter(ctx)
	a.coinFold(userid)
}

func (a *Desk) RMCoinBiReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMCoinBiReq)
	glog.Debugf("RMCoinBiReq %#v", arg)
	userid := a.getRouter(ctx)
	a.coinBi(userid)
}

func (a *Desk) RMCoinReplyBiReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMCoinReplyBiReq)
	glog.Debugf("RMCoinReplyBiReq %#v", arg)
	userid := a.getRouter(ctx)
	a.coinReplyBi(userid, arg.Agree)
}

func (a *Desk) RMCoinChangeRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMCoinChangeRoomReq)
	glog.Debugf("RMCoinChangeRoomReq %#v", arg)
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

// 对战房房主t人
func (a *Desk) RMKickoutRoomReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.RMKickoutRoomReq)
	rsp := new(pb.RMKickoutRoomRsp)
	ctx.Sender()
	// 游戏中不能踢人
	if a.IsGaming() {
		rsp.Error = pb.GameStarted
		ctx.Respond(rsp)
		return
	}
	// 非房主
	userId := a.getRouter(ctx)
	if a.Cid != userId {
		rsp.Error = pb.NotRoomCreator
		ctx.Respond(rsp)
		return
	}
	if _, ok := a.roles[arg.Userid]; !ok {
		rsp.Error = pb.NotInPrivateRoom
		ctx.Respond(rsp)
		return
	}

	// 通知玩家被踢出了
	msg2 := &pb.KickoutedRoom{
		Gameid: a.Game.Id,
		Roomid: a.DeskData.Rid,
		Gtype:  a.Gtype,
		Rtype:  a.Rtype,
	}
	a.send2userid(arg.Userid, msg2)

	rsp.KickoutSeat = a.getSeatid(arg.Userid)
	//玩家离开牌桌
	a.notifyNodeLeaveEarly(a.Rid, arg.Userid, pb.OK)
	a.notifyGateUserLeft(arg.Userid, pb.OK, int32(pb.OK))
	//清除数据
	a.userLeaveDesk(arg.Userid)

	ctx.Respond(rsp)
}

// 对战房换座位
func (t *Desk) RMChangeSeatReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.RMChangeSeatReq)
	rsp := new(pb.RMChangeSeatRsp)
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

// 从 gate 同步重置天roi次数重置
func (t *Desk) RMRoiRecordsync(ctx actor.Context) {
	arg := ctx.Message().(*pb.RMRoiRecordsync)
	player := t.getPlayer(arg.Userid)
	if player != nil && arg.ResetDayRoi {
		player.RmRoiDayLimits = make(map[string]int32)
	}
}
