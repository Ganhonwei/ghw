package crash

import (
	"goserver/gen/pb"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"time"

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
func (a *Desk) CRASHEnterRoomReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.CRASHEnterRoomReq)
	glog.Debugf("CRASHEnterRoomReq %#v", arg)
	userid := a.getRouter(ctx)
	msg1 := a.freeEnterMsg(userid)
	ctx.Respond(msg1)
	a.freeCameinMsg(userid)
	// a.robotHandler(userid)

	if _, ok := a.UserFactorMap[userid]; !ok {
		if role, ok := a.roles[userid]; ok {
			a.UserFactorMap[userid] = handler.GetFactor(role.User)
		}
	}
}

// 进入百人场成功
func (a *Desk) CRASHEnterSuccessReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.CRASHEnterSuccessReq)
	glog.Debugf("CRASHEnterSuccessReq %#v", arg)
	msg := new(pb.CRASHEnterSuccessRsp)
	userid := a.getRouter(ctx)
	// msg.Roominfo = handler.PackCrashFreeRoom(a.DeskData)
	msg.Roominfo = a.packCrashFreeRoom(userid)
	a.freeRoomDataMsg(msg.Roominfo, userid)
	//坐下玩家信息
	msg.Userinfo = a.freeSeatBetsMsg()
	//排行榜前20玩家
	// msg.Rank = a.freeRankMsg(5)
	// 自动逃离
	if v, ok := a.CRASHDeskFree.CrashAutoLeave[userid]; ok {
		msg.AutoCrash = true
		msg.CrashMultiple = v
	}
	if v, ok := a.CRASHDeskFree.CrashAutoLeave1[userid]; ok {
		msg.AutoCrash1 = true
		msg.CrashMultiple1 = v
	}
	ctx.Respond(msg)
	glog.Debugf("history %v,state:%d", msg.Roominfo.History, a.state)
}

// 撤离
func (a *Desk) CRASHBackReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.CRASHBackReq)
	glog.Debugf("CRASHBackReq %#v", arg)
	userid := a.getRouter(ctx)
	if arg.Userid != "" {
		userid = arg.Userid
	}

	a.back(userid, false, -1, arg.Pos)
}

// 下注
func (a *Desk) CRASHBetReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.CRASHBetReq)
	// glog.Debugf("CRASHBetReq %#v", arg)
	userid := a.getRouter(ctx)
	val := arg.GetValue() //值
	rsp := new(pb.CRASHBetRsp)

	if !a.checkBetPos(arg.Pos) {
		rsp.Error = pb.Failed
		ctx.Respond(rsp)
		return
	}

	errcode := a.freeBet(userid, int64(val), arg.Pos)
	if errcode != pb.OK {
		rsp.Error = errcode
		ctx.Respond(rsp)
		return
	}
}

// 下注
func (a *Desk) CRASHCancelBetReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.CRASHCancelBetReq)
	rsp := new(pb.CRASHCancelBetRsp)

	if !a.checkBetPos(arg.Pos) {
		rsp.Error = pb.Failed
		ctx.Respond(rsp)
		return
	}

	rsp = a.freeCancelBet(arg.Userid, arg.NextBet, arg.Pos)
	ctx.Respond(rsp)
}

// 离开
func (a *Desk) CRASHLeaveReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.CRASHLeaveReq)
	glog.Debugf("CRASHLeaveReq %#v", arg)
	userid := a.getRouter(ctx)
	a.nnLeave(userid, ctx)
}

// 设置自动逃离
func (a *Desk) CRASHSetAutoLeave(ctx actor.Context) {
	arg := ctx.Message().(*pb.CRASHCrashMultipleReq)
	glog.Debugf("CRASHCrashMultipleReq %#v", arg)
	rsp := new(pb.CRASHCrashMultipleRsp)

	userid := a.getRouter(ctx)
	role, ok := a.roles[userid]
	if !ok {
		rsp.Error = pb.Failed
		ctx.Respond(rsp)
		return
	}

	if !a.checkBetPos(arg.Pos) {
		rsp.Error = pb.Failed
		ctx.Respond(rsp)
		return
	}

	// 设置自动逃离倍数限制
	autoLimit := table.GetTables().CrashTable.Get(arg.Pos + 1).AutoBackMulpitleLimit[role.RegistArea]
	m1, m2 := int32(autoLimit.Value[0]*100), int32(autoLimit.Value[1]*100)
	if (m1 > 0 && arg.Multiple < m1) || (m2 > 0 && arg.Multiple > m2) {
		rsp.Error = pb.Failed
		ctx.Respond(rsp)
		return
	}

	a.setAutoLeave(userid, arg.AutoCrash, arg.Multiple, arg.Pos)

	rsp.Multiple = arg.Multiple
	rsp.AutoCrash = arg.AutoCrash
	rsp.Pos = arg.Pos
	ctx.Respond(rsp)
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
	ntf := new(pb.CRASHPushStateNtf)
	switch arg.Status {
	case int32(pb.STATE_READY):
		// 初始化牌局
		a.freeInit()
		// 检查状态
		// a.stateCheck()
		//更新排行榜
		// a.updateRank("")
		// 每个人的不下注轮数+1
		a.dealing()
		//踢出超时或余额不足的人
		a.limitOver()
	case int32(pb.STATE_BET):
		//奖池增加
		a.state = int32(pb.STATE_BET) // 下注状态
		// addScore := utils.RandInt64N(10000) + 5000
		// a.CrashJackpot += addScore
		// ntf := &pb.CRASHAddJackpotNtf{AddScore: int32(addScore)}
		// a.broadcast4(ntf)

		// 将下一轮下注提出来下注
		a.nextBetBetting()

		// 计算玩家系数
		for _, role := range a.roles {
			// a.UserFactorMap[role.Userid] = handler.GetFactor(role.User)
			a.CRASHObserver = append(a.CRASHObserver, role.Userid)
		}
	case int32(pb.STATE_LEAD):
		// 起飞
		a.takeoff()
		ntf.TakeoffMs = a.CRASHTakeoffTime.UnixMilli()
		ntf.CurMs = time.Now().UnixMilli()
	case int32(pb.STATE_OVER):
		a.state = int32(pb.STATE_OVER)
		a.freeGameOver()
	}
	// 广播
	glog.Debugf("CRASHPushStateNtf %d", arg.Status)
	ntf.State = pb.DeskState(arg.Status)
	ntf.During = arg.During
	a.broadcast(ntf)
}
