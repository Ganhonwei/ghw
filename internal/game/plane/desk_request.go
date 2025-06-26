package plane

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
func (a *Desk) PLANEEnterRoomReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.PLANEEnterRoomReq)
	glog.Debugf("PLANEEnterRoomReq %#v", arg)
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

// 撤离
func (a *Desk) PLANEBackReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.PLANEBackReq)
	glog.Debugf("PLANEBackReq %#v", arg)
	userid := a.getRouter(ctx)
	if arg.Userid != "" {
		userid = arg.Userid
	}
	a.back(userid, false, -1, arg.Pos)
}

// 下注
func (a *Desk) PLANEBetReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.PLANEBetReq)
	// glog.Debugf("PLANEBetReq %#v", arg)
	userid := a.getRouter(ctx)
	val := arg.GetValue() //值
	rsp := new(pb.PLANEBetRsp)

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

// 取消下注
func (a *Desk) PLANECancelBetReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.PLANECancelBetReq)
	rsp := new(pb.PLANECancelBetRsp)

	if !a.checkBetPos(arg.Pos) {
		rsp.Error = pb.Failed
		ctx.Respond(rsp)
		return
	}

	rsp = a.freeCancelBet(arg.Userid, arg.NextBet, arg.Pos)
	ctx.Respond(rsp)
}

// 离开
func (a *Desk) PLANELeaveReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.PLANELeaveReq)
	glog.Debugf("PLANELeaveReq %#v", arg)
	userid := a.getRouter(ctx)
	a.nnLeave(userid, ctx)
}

// 设置自动逃离
func (a *Desk) PLANESetAutoLeave(ctx actor.Context) {
	arg := ctx.Message().(*pb.PLANECrashMultipleReq)
	glog.Debugf("PLANECrashMultipleReq %#v", arg)
	rsp := new(pb.PLANECrashMultipleRsp)
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
	autoLimit := table.GetTables().PlaneTable.Get(arg.Pos + 1).AutoBackMulpitleLimit[role.RegistArea]
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

// 上一句记录
func (a *Desk) PLANELastRoundRecordReq(ctx actor.Context) {
	// arg := ctx.Message().(*pb.PLANELastRoundRecordReq)
	rsp := new(pb.PLANELastRoundRecordRsp)
	rsp.LastRecord = a.CRASHDeskFree.LastRecord
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
	ntf := new(pb.PLANEPushStateNtf)
	switch arg.Status {
	case int32(pb.STATE_READY):
		// 初始化牌局
		a.freeInit()
		// 状态检测
		// a.stateCheck()
		//机器人操作
		a.kickRobot()
		// 踢人
		a.limitOver()
		//更新排行榜
		// a.updateRank("")
		// 每个人的不下注轮数+1
		a.dealing()
	case int32(pb.STATE_BET):
		a.state = int32(pb.STATE_BET) // 下注状态
		// //奖池增加
		// addScore := utils.RandInt64N(10000) + 5000
		// a.CrashJackpot += addScore
		// ntf := &pb.PLANEAddJackpotNtf{AddScore: int32(addScore)}
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
	glog.Debugf("PLANEPushStateNtf %d", arg.Status)
	ntf.State = pb.DeskState(arg.Status)
	ntf.During = arg.During
	a.broadcast(ntf)
}
