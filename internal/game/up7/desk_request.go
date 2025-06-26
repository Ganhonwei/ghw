package up7

import (
	"goserver/gen/pb"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// 文本消息
func (a *Desk) ChatTextReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ChatTextReq)
	glog.Debugf("ChatTextReq %#v", arg)
	a.chatText(arg, ctx)
}

// 语音消息
// func (a *Desk) ChatVoiceReq(ctx actor.Context) {
// 	arg := ctx.Message().(*pb.ChatVoiceReq)
// 	glog.Debugf("ChatVoiceReq %#v", arg)
// 	a.chatVoiceReq(arg, ctx)
// }

// 进入百人场
func (a *Desk) UPFreeEnterRoomReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.UPFreeEnterRoomReq)
	glog.Debugf("UPFreeEnterRoomReq %#v", arg)
	userid := a.getRouter(ctx)
	msg1 := a.freeEnterMsg(userid)
	ctx.Respond(msg1)
	a.freeCameinMsg(userid)
	a._robotHandler(userid)

	if _, ok := a.UserFactorMap[userid]; !ok {
		if role, ok := a.roles[userid]; ok {
			a.UserFactorMap[userid] = handler.GetFactor(role.User)
		}
	}
}

// 进入百人场
func (a *Desk) UPEnterSuccessReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.UPEnterSuccessReq)
	glog.Debugf("UPEnterSuccessReq %#v", arg)
	userid := a.getRouter(ctx)
	msg := new(pb.UPEnterSuccessRsp)
	msg.Roominfo = handler.PackUPFreeRoom(a.DeskData)
	a.freeRoomDataMsg(userid, msg.Roominfo)
	//坐下玩家信息
	msg.Userinfo = a.freeSeatBetsMsg()
	//排行榜前20玩家
	msg.Rank = a.freeRankMsg(20)
	ctx.Respond(msg)
	glog.Debugf("history %v,state:%d", msg.Roominfo.Point, a.state)
}

// 玩家倍投下注(当前下注每一门注码都加倍)
func (a *Desk) UPDoubleBetReq(ctx actor.Context) {
	_ = ctx.Message().(*pb.UPDoubleBetReq)
	userid := a.getRouter(ctx)
	rsp := a.freeDoubleBet(userid)
	ctx.Respond(rsp)

	if rsp.Error == pb.OK {
		msg := &pb.UPDoubleBetNtf{
			Userid:    rsp.Userid,
			Value:     rsp.Value,
			Bets:      rsp.Bets,
			SeatValue: rsp.SeatValue,
			SeatBets:  rsp.SeatBets,
			SeatPools: rsp.SeatPools,
		}
		a.broadcast2(userid, msg)
	}
}

// 撤销上一步下注
func (a *Desk) UPUndoBetReq(ctx actor.Context) {
	_ = ctx.Message().(*pb.UPUndoBetReq)
	userid := a.getRouter(ctx)
	rsp := a.freeUndoBet(userid)
	ctx.Respond(rsp)

	if rsp.Error == pb.OK {
		msg := &pb.UPUndoBetNtf{
			Userid:   userid,
			UndoBets: rsp.UndoBets,
			BetSeat:  rsp.BetSeat,
			SeatBets: rsp.SeatBets,
			SeatPool: rsp.SeatPool,
			Bets:     rsp.Bets,
			Players:  rsp.Players,
		}
		a.broadcast2(userid, msg)
	}
}

// 下注
func (a *Desk) UPFreeBetReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.UPFreeBetReq)
	// glog.Debugf("UPFreeBetReq %#v", arg)
	userid := a.getRouter(ctx)
	seatBet := arg.GetSeat() //位置
	val := arg.GetValue()    //值
	errcode := a.freeBet(userid, seatBet, int64(val))
	if errcode == pb.OK {
		return
	}
	//响应
	rsp := new(pb.UPFreeBetRsp)
	rsp.Error = errcode
	ctx.Respond(rsp)
}

// 离开
func (a *Desk) UPLeaveReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.UPLeaveReq)
	glog.Debugf("UPLeaveReq %#v", arg)
	userid := a.getRouter(ctx)
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
		//更新排行榜
		a.updateRank("")
	case int32(pb.STATE_DEALING):
		// 每个人的不下注轮数+1
		a.dealing()
	case int32(pb.STATE_BET):
		// 计算玩家系数
		for _, role := range a.roles {
			// a.UserFactorMap[role.Userid] = handler.GetFactor(role.User)
			a.LHObserver = append(a.LHObserver, role.Userid)
		}
	case int32(pb.STATE_LEAD):
		// 开牌
		a.lead()
		a.state = int32(pb.STATE_LEAD)
		// a.tickStop = false
	case int32(pb.STATE_OVER):
		// 结算
		a.state = int32(pb.STATE_OVER)
		a.freeGameOver()
		//玩家状态检测
		// a.stateCheck()
		// a.tickStop = false
	}
	// 广播
	glog.Debugf("UPPushStateNtf %d", arg.Status)
	ntf := new(pb.UPPushStateNtf)
	ntf.State = pb.DeskState(arg.Status)
	ntf.During = arg.During
	a.broadcast(ntf)
}
