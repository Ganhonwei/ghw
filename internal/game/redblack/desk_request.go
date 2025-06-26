package redblack

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

// 进入百人场
func (a *Desk) RBFreeEnterRoomReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.RBFreeEnterRoomReq)
	glog.Debugf("RBFreeEnterRoomReq %#v", arg)
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

// 下注
func (a *Desk) RBFreeBetReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.RBFreeBetReq)
	// glog.Debugf("RBFreeBetReq %#v", arg)
	userid := a.getRouter(ctx)
	seatBet := arg.GetSeat() //位置
	val := arg.GetValue()    //值
	errcode := a.freeBet(userid, seatBet, int64(val))
	if errcode == pb.OK {
		return
	}
	//响应
	rsp := new(pb.RBFreeBetRsp)
	rsp.Error = errcode
	ctx.Respond(rsp)
}

// 离开
func (a *Desk) RBLeaveReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.RBLeaveReq)
	glog.Debugf("RBLeaveReq %#v", arg)
	userid := a.getRouter(ctx)
	a.nnLeave(userid, ctx)
}

// 银行赠送
func (a *Desk) BankGive(ctx actor.Context) {
	arg := ctx.Message().(*pb.BankGive)
	glog.Debugf("BankGive %#v", arg)
	if v, ok := a.roles[arg.GetUserid()]; ok && v != nil {
		v.User.AddDiamond(arg.GetCoin())
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
	case int32(pb.STATE_DEALING):
		// 每个人的不下注轮数+1
		a.dealing()
	case int32(pb.STATE_BET):
		a.state = int32(pb.STATE_BET) // 下注状态
		// 计算玩家系数
		for _, role := range a.roles {
			a.UserFactorMap[role.Userid] = handler.GetFactor(role.User)
			a.RBObserver = append(a.RBObserver, role.Userid)
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
	glog.Debugf("rid:%s, RBPushStateNtf %d", a.Rid, arg.Status)
	ntf := new(pb.RBPushStateNtf)
	ntf.State = pb.DeskState(arg.Status)
	ntf.During = arg.During
	a.broadcast(ntf)
}
