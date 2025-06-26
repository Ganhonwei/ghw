package robot

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (a *RoleActor) matchcpDesk(arg *pb.RobotMsg) *pb.MatchDesk {
	// 获取游戏服节点或者房间进程
	msg := new(pb.MatchDesk)
	// TODO 优化查找规则
	msg.Name = getGameName(arg.Gtype)
	msg.Gtype = arg.Gtype
	msg.Rtype = arg.Rtype
	msg.Roomid = arg.Roomid
	return msg
}

// 离开消息
func (a *RoleActor) LotteryLeaveNtf(ctx actor.Context) {
	ntf := ctx.Message().(*pb.LotteryLeaveNtf)
	// 关闭节点
	glog.Debugf("LotteryLeaveNtf %v", ntf)
	stop := new(pb.ServeStop)
	a.pid.Tell(stop)
}

// 离开消息
func (a *RoleActor) LotteryLeaveRsp(ctx actor.Context) {
	ntf := ctx.Message().(*pb.LotteryLeaveRsp)
	// 关闭节点
	glog.Debugf("LotteryLeaveRsp %v", ntf)
	stop := new(pb.ServeStop)
	a.pid.Tell(stop)
}

func (a *RoleActor) LotteryEnterRsp(ctx actor.Context) {
	res := ctx.Message().(*pb.LotteryEnterRsp)
	glog.Debugf("LotteryEnterRsp %v", res)
	if res.Error != pb.OK {
		glog.Infof("LotteryEnterRsp enter err:%s", res.Error.String())
		stop := new(pb.ServeStop)
		a.pid.Tell(stop)
	}
}

func (a *RoleActor) cpCheckStatus() {
	switch a.cp.state {
	case int32(pb.STATE_READY):
		a.cp.timer = 15
	case int32(pb.STATE_BET):
		// 下注
		a.cpFreeBet()
	}
}

// 人机策略
func (r *RoleActor) recvCPRobotStrategy(ctx actor.Context) {
	s2c := ctx.Message().(*pb.CPRobotStrategyNtf)
	r.cp.cpstrategy = s2c
	r.cp.betCount = 0
	// 下注轮数
	if len(s2c.BetTimes) <= 0 {
		r.cp.maxBetCount = 50
		return
	}
	r.cp.maxBetCount = int32(utils.RandMN(int(s2c.BetTimes[0]), int(s2c.BetTimes[1])))
}

// 状态变化
func (r *RoleActor) recvCPState(ctx actor.Context) {
	msg := ctx.Message().(*pb.LotteryPushStateNtf)
	r.cp.state = int32(msg.State)
	switch msg.State {
	case pb.STATE_READY:
		if r.cp.betCount >= r.cp.maxBetCount {
			// 离开
			msg := new(pb.LotteryLeaveReq)
			r.Sender(msg)
		}
	case pb.STATE_DEALING:
		r.NowChip = 0
	case pb.STATE_BET:
		// 下注
		r.cp.betCount++
	}
}

// 下注
func (r *RoleActor) cpFreeBet() {
	if r.cp.cpstrategy == nil {
		return
	}
	chips := r.cp.cpstrategy.Chips
	seat := r.NowSeat
	if r.NowChip > 0 {
		// 没下完继续下
		// 拆分筹码,然后下注
		if utils.LocalTime().UnixMilli() < r.nextTime {
			return
		}
		if r.NowChip < chips[0] {
			r.NowChip = chips[0]
		}
		for i := len(chips) - 1; i >= 0; i-- {
			if r.NowChip >= chips[i] {
				r.NowChip -= chips[i]
				// 下注
				c2s := new(pb.LotteryBetReq)
				c2s.Seat = seat
				c2s.Value = chips[i]
				if r.gamePid != nil {
					r.Sender(c2s)
					r.nextTime = utils.LocalTime().UnixMilli() + time.Millisecond.Milliseconds()*500
				}
				break
			}
		}
		return
	}
	if r.cp.timer >= r.cp.waitTime {
		r.cp.timer = 0
		r.cp.waitTime = utils.RandInt32N(11) + 20
	} else {
		r.cp.timer++
		return
	}
	// 随机下
	r.NowSeat = r.getCPSeat()
	// 下注金额
	r.NowChip = r.getCPChip()
	if uint32(r.Coin) < r.NowChip {
		// 筹码不够
		return
	}
	// glog.Infof("user %s cp bet seat:%d score:%d ", r.Userid, r.NowSeat, r.NowChip)

}

// 根据人机策略选择下注
func (c *RoleActor) getCPSeat() uint32 {
	var choices []utils.Choice
	if c.cp.cpstrategy == nil || len(c.cp.cpstrategy.BetWeight) == 0 {
		return 1
	}
	weights := c.cp.cpstrategy.BetWeight
	for i, v := range weights {
		choices = append(choices, utils.Choice{Weight: int(v), Item: uint32(i + 1)})
	}
	choice, err := utils.WeightedChoice(choices)
	if err != nil {
		return 1
	}
	return choice.Item.(uint32)
}

// 根据人机策略选择筹码
func (c *RoleActor) getCPChip() uint32 {
	// 根据概率判断下不下注TODO
	pro := []int32{25, 75}
	if c.cp.cpstrategy != nil {
		pro = c.cp.cpstrategy.BetPro
	}
	var choices []utils.Choice
	choices = append(choices, utils.Choice{Weight: int(pro[0]), Item: 0}) // 不下注
	choices = append(choices, utils.Choice{Weight: int(pro[1]), Item: 1}) // 下注
	choice, err := utils.WeightedChoice(choices)
	if err == nil && choice.Item == 0 {
		return 0
	}
	// 选择筹码
	betArea := c.cp.cpstrategy.BetArea
	base := c.cp.cpstrategy.Min
	bet := utils.RandInt32N(betArea[1]/base-betArea[0]/base) + betArea[0]/base
	return uint32(bet * base)
}
