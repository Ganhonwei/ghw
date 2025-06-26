package fortune_gems2

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// mines 进入房间
func (a *Desk) FortuneGems2EnterRoomReq(ctx actor.Context) {
	// arg := ctx.Message().(*pb.FortuneGems2EnterRoomReq)
	userid := a.getRouter(ctx)
	rsp := a.fortuneGems2EnterMsg(userid)
	ctx.Respond(rsp)
}

// mines 下注
func (a *Desk) FortuneGems2BetReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.FortuneGems2BetReq)
	userid := a.getRouter(ctx)
	rsp := a.fortuneGems2Bet(userid, arg)
	ctx.Respond(rsp)
}

// mines 踩雷
// func (a *Desk) FortuneGems2PitStepReq(ctx actor.Context) {
// 	arg := ctx.Message().(*pb.FortuneGems2PitStepReq)
// 	rsp := new(pb.FortuneGems2PitStepRsp)

// 	if a.state != int32(pb.STATE_LEAD) {
// 		rsp.Error = pb.OperateError
// 		ctx.Respond(rsp)
// 		return
// 	}
// 	// 位置有误, 或已踩过
// 	if arg.Pit > 24 || (arg.Pit >= 0 && a.minesPitsUser[arg.Pit] != 0) {
// 		rsp.Error = pb.ErrorOperateValue
// 		ctx.Respond(rsp)
// 		return
// 	}
// 	pit := arg.Pit
// 	if pit < 0 {
// 		// 随机选个没踩过的坑位
// 		pit = a.minesRandomPitStep()
// 	}

// 	boom := a.minesPitStep(pit, false)
// 	rsp.Pit = pit
// 	rsp.PitV = int32(utils.CaseElse(boom, 2, 1))

// 	// 当前返奖倍数,下一步返奖倍数
// 	var curStepMultiple, nextStepMultiple string
// 	curStepMultiple = getStepMultiple(a.mines, a.step)
// 	if !boom {
// 		nextStepMultiple = getStepMultiple(a.mines, a.step+1)
// 		// 剩余安全步数
// 		rsp.SafePits = a.minesSafePits()
// 	}
// 	if nextStepMultiple != "" {
// 		rsp.Multiple = nextStepMultiple
// 	} else {
// 		rsp.Multiple = curStepMultiple
// 	}
// 	rsp.Score = int64(float64(a.betNum) * str2float(curStepMultiple))
// 	ctx.Respond(rsp)

// 	// 踩到雷了 或坑已踩完
// 	if boom || (!boom && a.minesSafePits() == 0) {
// 		a.gameOver(false, false)
// 	}
// }

// mines 手动对局提现请求
// func (a *Desk) FortuneGems2CashOutReq(ctx actor.Context) {
// 	rsp := new(pb.FortuneGems2CashOutRsp)

// 	if a.state != int32(pb.STATE_LEAD) || a.autoMines {
// 		rsp.Error = pb.OperateError
// 		ctx.Respond(rsp)
// 		return
// 	}
// 	// 强制结束算玩家输
// 	if a.step <= 0 {
// 		a.forceCashOut = true
// 	}
// 	a.gameOver(false, false)

// 	rsp.Multiple = a.settleMultiple
// 	rsp.Score = a.settleScore
// 	ctx.Respond(rsp)
// }

// mines 取消自动下注
// func (a *Desk) FortuneGems2AutoBetCancelReq(ctx actor.Context) {
// 	// arg := ctx.Message().(*pb.FortuneGems2AutoBetCancelReq)
// 	rsp := new(pb.FortuneGems2AutoBetCancelRsp)

// 	a.autoRound = 0
// 	ctx.Respond(rsp)
// }

// mines 离开
func (a *Desk) FortuneGems2LeaveReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.FortuneGems2LeaveReq)
	glog.Debugf("FortuneGems2LeaveReq %#v", arg)
	a.nnLeave(ctx)
}
