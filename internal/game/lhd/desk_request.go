package lhd

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
func (a *Desk) LHFreeEnterRoomReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.LHFreeEnterRoomReq)
	glog.Debugf("LHFreeEnterRoomReq %#v", arg)
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

// 进入百人场 废弃
func (a *Desk) LHEnterSuccessReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.LHEnterSuccessReq)
	glog.Debugf("LHEnterSuccessReq %#v", arg)
	msg := new(pb.LHEnterSuccessRsp)
	msg.Roominfo = handler.PackLHFreeRoom(a.DeskData)
	a.freeRoomDataMsg(msg.Roominfo)
	//坐下玩家信息
	msg.Userinfo = a.freeSeatBetsMsg()
	//排行榜前6玩家
	msg.Rank = a.freeRankMsg(20)
	ctx.Respond(msg)
}

// 庄家
// func (a *Desk) LHFreeDealerReq(ctx actor.Context) {
// 	arg := ctx.Message().(*pb.LHFreeDealerReq)
// 	glog.Debugf("LHFreeDealerReq %#v", arg)
// 	userid := a.getRouter(ctx)
// 	var state int32 = arg.GetState()
// 	var num uint32 = arg.GetCoin()
// 	errcode := a.beDealer(userid, state, num)
// 	if errcode == pb.OK {
// 		return
// 	}
// 	//响应
// 	rsp := new(pb.LHFreeDealerRsp)
// 	rsp.Error = errcode
// 	ctx.Respond(rsp)
// }

// 庄家列表
// func (a *Desk) LHFreeDealerListReq(ctx actor.Context) {
// 	arg := ctx.Message().(*pb.LHFreeDealerListReq)
// 	glog.Debugf("LHFreeDealerListReq %#v", arg)
// 	rsp := a.dealerListMsg()
// 	ctx.Respond(rsp)
// }

// 坐下
// func (a *Desk) LHSitReq(ctx actor.Context) {
// 	arg := ctx.Message().(*pb.LHSitReq)
// 	glog.Debugf("LHSitReq %#v", arg)
// 	userid := a.getRouter(ctx)
// 	rsp := a.freeSit(userid, arg)
// 	if rsp.Error == pb.OK {
// 		a.broadcast(rsp)
// 		return
// 	}
// 	ctx.Respond(rsp)
// }

// 下注
func (a *Desk) LHFreeBetReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.LHFreeBetReq)
	// glog.Debugf("LHFreeBetReq %#v", arg)
	userid := a.getRouter(ctx)
	seatBet := arg.GetSeat() //位置
	val := arg.GetValue()    //值
	errcode := a.freeBet(userid, seatBet, int64(val))
	if errcode == pb.OK {
		return
	}
	//响应
	rsp := new(pb.LHFreeBetRsp)
	rsp.Error = errcode
	ctx.Respond(rsp)
}

// 离开
func (a *Desk) LHLeaveReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.LHLeaveReq)
	glog.Debugf("LHLeaveReq %#v", arg)
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
		//玩家状态检测
		a.stateCheck()
		//踢出超时或余额不足的人
		a.limitOver()
		//更新排行榜
		a.updateRank("")
	case int32(pb.STATE_DEALING):
		// 每个人的不下注轮数+1
		a.dealing()
	case int32(pb.STATE_BET):
		a.state = int32(pb.STATE_BET) // 下注状态
		// 计算玩家系数
		for _, role := range a.roles {
			a.UserFactorMap[role.Userid] = handler.GetFactor(role.User)
			a.LHObserver = append(a.LHObserver, role.Userid)
		}
	case int32(pb.STATE_LEAD):
		// 开牌
		a.lead()
		a.state = int32(pb.STATE_LEAD)
		a.tickStop = false
	case int32(pb.STATE_OVER):
		// 结算
		a.state = int32(pb.STATE_OVER)
		a.freeGameOver()
		a.tickStop = false
	}
	// 广播
	glog.Debugf("rid:%s, LHPushStateNtf %d", a.Rid, arg.Status)
	ntf := new(pb.LHPushStateNtf)
	ntf.State = pb.DeskState(arg.Status)
	ntf.During = arg.During
	a.broadcast(ntf)
}
