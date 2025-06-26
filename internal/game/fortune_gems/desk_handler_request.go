package fortune_gems

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// mines 进入房间
func (a *Desk) FortuneGemsEnterRoomReq(ctx actor.Context) {
	// arg := ctx.Message().(*pb.FortuneGemsEnterRoomReq)
	userid := a.getRouter(ctx)
	rsp := a.fortuneGemsEnterMsg(userid)
	ctx.Respond(rsp)
}

// mines 下注
func (a *Desk) FortuneGemsBetReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.FortuneGemsBetReq)
	userid := a.getRouter(ctx)
	rsp := a.fortuneGemsBet(userid, arg)
	ctx.Respond(rsp)
}

// mines 踩雷
// func (a *Desk) FortuneGemsPitStepReq(ctx actor.Context) {
// 	arg := ctx.Message().(*pb.FortuneGemsPitStepReq)
// 	rsp := new(pb.FortuneGemsPitStepRsp)

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
// func (a *Desk) FortuneGemsCashOutReq(ctx actor.Context) {
// 	rsp := new(pb.FortuneGemsCashOutRsp)

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
// func (a *Desk) FortuneGemsAutoBetCancelReq(ctx actor.Context) {
// 	// arg := ctx.Message().(*pb.FortuneGemsAutoBetCancelReq)
// 	rsp := new(pb.FortuneGemsAutoBetCancelRsp)

// 	a.autoRound = 0
// 	ctx.Respond(rsp)
// }

// mines 离开
func (a *Desk) FortuneGemsLeaveReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.FortuneGemsLeaveReq)
	glog.Debugf("FortuneGemsLeaveReq %#v", arg)
	a.nnLeave(ctx)
}
