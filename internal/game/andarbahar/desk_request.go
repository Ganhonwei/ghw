package andarbahar

import (
	"goserver/gen/pb"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/utils"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// 文本消息
func (a *Desk) ChatTextReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ChatTextReq)
	glog.Debugf("ChatTextReq %#v", arg)
	a.chatText(arg, ctx)
}

// 进入百人场
func (a *Desk) ABFreeEnterRoomReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ABFreeEnterRoomReq)
	glog.Debugf("ABFreeEnterRoomReq %#v", arg)
	userid := a.getRouter(ctx)
	msg1 := a.freeEnterMsg(userid)
	ctx.Respond(msg1)
	a.freeCameinMsg(userid)
	a.robotHandler(userid)

	if _, ok := a.UserFactorMap[userid]; !ok {
		if role, ok := a.roles[userid]; ok {
			a.UserFactorMap[userid] = handler.GetFactor(role.User)
		}
	}
}

// 进入百人场
func (a *Desk) ABEnterSuccessReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ABEnterSuccessReq)
	glog.Debugf("ABEnterSuccessReq %#v", arg)
	msg := new(pb.ABEnterSuccessRsp)
	msg.Roominfo = handler.PackABFreeRoom(a.DeskData)
	a.freeRoomDataMsg(msg.Roominfo)
	//坐下玩家信息
	msg.Userinfo = a.freeSeatBetsMsg()
	//排行榜前6玩家
	msg.Rank = a.freeRankMsg(20)
	ctx.Respond(msg)
}

// 下注
func (a *Desk) ABFreeBetReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ABFreeBetReq)
	// glog.Debugf("ABFreeBetReq %#v", arg)
	userid := a.getRouter(ctx)
	seatBet := arg.GetSeat() //位置
	val := arg.GetValue()    //值
	errcode := a.freeBet(userid, seatBet, int64(val))
	if errcode == pb.OK {
		return
	}
	//响应
	rsp := new(pb.ABFreeBetRsp)
	rsp.Error = errcode
	ctx.Respond(rsp)
}

// 离开
func (a *Desk) ABLeaveReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ABLeaveReq)
	glog.Debugf("ABLeaveReq %#v", arg)
	userid := a.getRouter(ctx)

	// 私人房再来一句投票时退出算放弃
	if a.isPrivRoom() {
		a.chargeInGameCancel(userid)

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

// 银行赠送
func (a *Desk) BankGive(ctx actor.Context) {
	arg := ctx.Message().(*pb.BankGive)
	glog.Debugf("BankGive %#v", arg)
	if v, ok := a.roles[arg.GetUserid()]; ok && v != nil {
		v.User.AddCoin(arg.GetCoin())
	}
}

// 桌子状态发生变化
func (a *Desk) ChangeFreeDeskStatus(ctx actor.Context) {
	arg := ctx.Message().(*pb.ChangeFreeDeskStatus)
	glog.Debugf("ChangeFreeDeskStatus %#v", arg)
	switch arg.Status {
	case int32(pb.STATE_READY):
		// 初始化牌局
		a.freeInit()
		//机器人操作
		a.kickRobot()
		//玩家状态检测
		a.stateCheck()
		//踢人
		a.limitOver()
		//更新排行榜
		a.updateRank("")
	case int32(pb.STATE_BET):
		// 每个人的不下注轮数+1
		a.dealing()

		// 计算玩家系数
		for _, role := range a.roles {
			a.UserFactorMap[role.Userid] = handler.GetFactor(role.User)
			a.ABObserver = append(a.ABObserver, role.Userid)
		}
	case int32(pb.STATE_LEAD):
		// 开牌
		a.lead()
		// 取消暂停
		a.tickStop = false
		a.state = int32(pb.STATE_LEAD)
	case int32(pb.STATE_OVER):
		// 结算
		a.state = int32(pb.STATE_OVER)
		a.freeGameOver()
		a.tickStop = false
	}
	// 广播
	glog.Debugf("rid:%s, ABPushStateNtf %d", a.Rid, arg.Status)
	ntf := new(pb.ABPushStateNtf)
	ntf.State = pb.DeskState(arg.Status)
	ntf.During = arg.During
	a.broadcast(ntf)
}

// 打包桌子信息 todo 庄家
func (a *Desk) ABEnterRoomReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ABEnterRoomReq)
	glog.Debugf("ABEnterRoomReq %#v", arg)
	userid := a.getRouter(ctx)
	msg1 := a.privEnterMsg()
	ctx.Respond(msg1)
	a.privCameinMsg(userid)
}

// 对战房手动开始游戏
func (a *Desk) ABGameStartReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ABGameStartReq)
	glog.Debugf("ABGameStartReq %#v", arg)
	rsp := new(pb.ABGameStartRsp)

	userid := a.getRouter(ctx)
	if a.Cid != userid {
		rsp.Error = pb.NotRoomCreator
		ctx.Respond(rsp)
		return
	}
	// 人数>2
	if 2 > len(a.seats) {
		rsp.Error = pb.DeskUndernumber
		ctx.Respond(rsp)
		return
	}
	// 对局结束解散倒计时中
	if a.dismissOnRoundOver || a.settleExitTime != 0 {
		rsp.Error = pb.RoomMaintenance
		ctx.Respond(rsp)
		return
	}
	// 庄家
	if _, ok := a.seats[a.DealerSeat]; !ok {
		rsp.Error = pb.NotDealerRoom
		ctx.Respond(rsp)
		return
	}

	//a.InitDesk()
	err := a.privGameStart()
	if err != pb.OK {
		rsp.Error = err
	}
	ctx.Respond(rsp)
}

// 踢出玩家
func (a *Desk) ABKickoutRoomReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ABKickoutRoomReq)
	glog.Debugf("ABKickoutRoomReq %#v", arg)
	rsp := new(pb.JHKickoutRoomRsp)
	userId := a.getRouter(ctx)
	// 游戏中不能踢人
	if a.IsGaming() {
		rsp.Error = pb.GameStarted
		ctx.Respond(rsp)
		return
	}
	// 非房主
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
	// 通知网关玩家被踢出了
	msg2 := &pb.KickoutedRoom{
		Gameid: a.Game.Id,
		Roomid: a.DeskData.Rid,
		Gtype:  a.Gtype,
		Rtype:  a.Rtype,
	}
	a.send2userid(arg.Userid, msg2)

	// 清除玩家
	rsp.KickoutSeat = a.getSeat(arg.Userid)
	//玩家离开牌桌
	a.notifyGateUserLeft(arg.Userid, pb.OK, int32(pb.OK))
	//清除数据
	a.userLeaveDesk(arg.Userid)
	ctx.Respond(rsp)
}

// 玩家换座
func (a *Desk) ABChangeSeatReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ABChangeSeatReq)
	glog.Debugf("ABChangeSeatReq %#v", arg)
	rsp := new(pb.ABChangeSeatRsp)
	userId := a.getRouter(ctx)
	seat := a.getSeat(userId)
	if a.IsGaming() {
		rsp.Error = pb.GameStarted
		ctx.Respond(rsp)
		return
	}
	var isDealer, toDealer bool
	isDealer = a.Dealer == userId
	toDealer = a.DealerSeat == arg.ToSeat
	// 换自己
	if arg.ToSeat == a.getSeat(userId) {
		rsp.Error = pb.ChangeSeatSelf
		ctx.Respond(rsp)
		return
	}
	// 闲家不可互换
	if !isDealer && !toDealer {
		// 位置上没人可以换
		if _, ok := a.seats[arg.ToSeat]; ok {
			rsp.Error = pb.SeatTaken
			ctx.Respond(rsp)
			return
		}
	}
	// 1分钟内申请一次
	chaning, ok := a.changingSeat[userId]
	if ok {
		now := utils.Timestamp()
		if now-chaning.Time < 60 {
			rsp.Error = pb.NotYourTurn
			ctx.Respond(rsp)
			return
		}
		delete(a.changingSeat, userId)
	}
	// 有未处理的换座请求
	for _, c := range a.changingSeat {
		if c.Accept == 0 && c.ToUserid == userId {
			rsp.Error = pb.NotYourTurn
			ctx.Respond(rsp)
			return
		}
	}

	// 换到庄家位置，判断金额
	if toDealer {
		score, access := a.getRole(userId).GetScore(), int64(a.DeskData.Game.AB.DealerAccess)
		// if a.getRole(userId).GetScore() < int64(a.DeskData.Game.AB.DealerAccess) {
		if score < access {
			rsp.Error = pb.BeDealerNotEnough
			ctx.Respond(rsp)
			return
		}
	}

	// 位置上没人直接换
	if _, ok := a.seats[arg.ToSeat]; !ok {
		glog.Infof("dealer %s change seat: %d to %d", rsp.Userid, rsp.Seat, rsp.ToSeat)
		rsp.Userid = userId
		rsp.Seat = a.getSeat(userId)
		rsp.ToSeat = arg.ToSeat
		rsp.DealerSeat = a.DeskGame.DealerSeat
		//ctx.Respond(rsp)

		a.seats[rsp.ToSeat] = a.seats[rsp.Seat]
		delete(a.seats, rsp.Seat)
		a.getRole(rsp.Userid).Seat = rsp.ToSeat
		a.Dealer = "" // 庄家置空
		if dealer, ok := a.seats[a.DealerSeat]; ok {
			a.Dealer = dealer.Userid // 更新庄家id
		}
		rsp.Userinfo = a.privSeatBetsMsg()
		rsp.LeftSeat = a.privLeftSeats() // 剩余座位号
		a.broadcast(rsp)
		return
	}

	// 发出通知
	rsp.Error = pb.WaitAccept
	ctx.Respond(rsp)

	// 记录换座，待接受
	a.changingSeat[userId] = &ChangingSeat{
		Time:     utils.Timestamp(),
		Userid:   userId,
		ToUserid: a.getUserid(arg.ToSeat),
	}
	ntf := &pb.ABChangeSeatAcceptNtf{
		Userid: userId,
		Seat:   seat,
		ToSeat: arg.ToSeat,
	}
	a.send2seat(arg.ToSeat, ntf)
}

// 是否同意换座位
func (a *Desk) ABChangeSeatAcceptReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ABChangeSeatAcceptReq)
	glog.Debugf("ABChangeSeatAcceptReq %#v", arg)
	rsp := new(pb.ABChangeSeatAcceptRsp)
	rsp.Accept = arg.Accept
	userId := a.getRouter(ctx)
	if a.IsGaming() {
		rsp.Error = pb.GameStarted
		ctx.Respond(rsp)
		return
	}
	var changed *ChangingSeat
	for _, changing := range a.changingSeat {
		if changing.ToUserid == userId {
			changing.Accept = arg.Accept
			changed = changing
			if arg.Accept != 1 { // 拒绝
				ctx.Respond(rsp)
				// 通知被拒绝
				a.send2userid(changing.Userid, rsp)
				return
			}
			// 同意换座
			seat := a.getSeat(changing.Userid)
			toSeat := a.getSeat(changing.ToUserid)
			temp := a.seats[toSeat]
			a.seats[toSeat] = a.seats[seat]
			a.seats[seat] = temp
			a.getRole(changing.Userid).Seat = toSeat
			a.getRole(changing.ToUserid).Seat = seat
			if dealer, ok := a.seats[a.DealerSeat]; ok {
				a.Dealer = dealer.Userid // 更新庄家id
			}
			ctx.Respond(rsp)
			// 通知被同意换座
			a.send2userid(changing.Userid, rsp)
			// 广播换了位置
			msg := &pb.ABChangeSeatRsp{
				Userid:     changing.Userid,
				Seat:       seat,
				ToSeat:     toSeat,
				DealerSeat: a.DeskGame.DealerSeat,
			}
			msg.Userinfo = a.privSeatBetsMsg()
			msg.LeftSeat = a.privLeftSeats() // 剩余座位号
			a.broadcast(msg)
			break
		}
	}

	// 同意交换了位置，清除其他换座请求
	if changed != nil && changed.Accept == 1 {
		for _, c := range a.changingSeat {
			if c.Accept == 0 && c.ToUserid == changed.ToUserid {
				c.Accept = 2
				// 通知玩家被拒绝了
				msg := new(pb.ABChangeSeatAcceptRsp)
				msg.Accept = c.Accept
				a.send2userid(c.Userid, msg)
			}
		}
		return
	}

	rsp.Error = pb.NotYourTurn
	ctx.Respond(rsp)
}

// 是否同意换座位
func (a *Desk) ABLaunchVoteReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ABLaunchVoteReq)
	glog.Debugf("ABLaunchVoteReq %#v", arg)
	userid := a.getRouter(ctx)
	rsp := new(pb.ABLaunchVoteRsp)
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
	if a.Cid != userid && !a.IsGaming() && a.DeskGame.Round == 0 {
		rsp.Error = pb.LeaveEarly
		ctx.Respond(rsp)

		// a.notifyNodeLeaveEarly(a.Rid, userid, pb.LeaveEarly)
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

// 是否同意换座位
func (a *Desk) ABVoteReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ABVoteReq)
	glog.Debugf("ABVoteReq %#v", arg)
	userid := a.getRouter(ctx)
	var vote uint32 = arg.GetVote()
	rsp := a.privVote(userid, vote)
	if rsp.Error == pb.OK {
		return
	}
	ctx.Respond(rsp)
}

// 私人房下注
func (a *Desk) ABBetReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ABBetReq)
	glog.Debugf("ABBetReq %#v", arg)
	userid := a.getRouter(ctx)
	rsp := a.privBet(userid, arg.Seat, int64(arg.Chip))
	if rsp.Error == pb.OK || rsp.Error == pb.NotEnoughCoin {
		return
	}
	ctx.Respond(rsp)
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

	// 再次开始游戏
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
