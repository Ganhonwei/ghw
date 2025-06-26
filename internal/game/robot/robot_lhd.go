package robot

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (a *RoleActor) matchlhdDesk(arg *pb.RobotMsg) *pb.MatchDesk {
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
func (a *RoleActor) LHLeaveNtf(ctx actor.Context) {
	ntf := ctx.Message().(*pb.LHLeaveNtf)
	// glog.Debugf("LHLeaveNtf %v", ntf)
	if ntf.Userid == a.Userid {
		// 关闭节点
		glog.Debugf("LHLeaveNtf %v", ntf)
		stop := new(pb.ServeStop)
		a.pid.Tell(stop)
	}
}

// 离开消息
func (a *RoleActor) LHLeaveRsp(ctx actor.Context) {
	ntf := ctx.Message().(*pb.LHLeaveRsp)
	if ntf.Userid == a.Userid {
		// 关闭节点
		glog.Debugf("LHLeaveRsp %v", ntf)
		stop := new(pb.ServeStop)
		a.pid.Tell(stop)
	}
}

func (a *RoleActor) LHFreeEnterRoomRsp(ctx actor.Context) {
	res := ctx.Message().(*pb.LHFreeEnterRoomRsp)
	glog.Debugf("LHFreeEnterRoomRsp %v", res)
	if res.Error != pb.OK {
		glog.Infof("LHFreeEnterRoomRsp enter err:%s", res.Error.String())
		stop := new(pb.ServeStop)
		a.pid.Tell(stop)
	}
}

func (a *RoleActor) lhdCheckStatus() {
	switch a.lh.state {
	case int32(pb.STATE_READY):
		a.lh.timer = 15
	case int32(pb.STATE_BET):
		// 下注
		a.lhFreeBet()
	}
}

// 人机策略
func (r *RoleActor) recvLHRobotStrategy(ctx actor.Context) {
	s2c := ctx.Message().(*pb.LHRobotStrategyNtf)
	r.lh.lhstrategy = s2c
	r.lh.betCount = 0
	// 下注轮数
	if len(s2c.BetTimes) <= 0 {
		r.lh.maxBetCount = 50
		return
	}
	r.lh.maxBetCount = int32(utils.RandMN(int(s2c.BetTimes[0]), int(s2c.BetTimes[1])))
}

// 状态变化
func (r *RoleActor) recvLHState(ctx actor.Context) {
	msg := ctx.Message().(*pb.LHPushStateNtf)
	r.lh.state = int32(msg.State)
	switch msg.State {
	case pb.STATE_READY:
		if r.lh.betCount >= r.lh.maxBetCount {
			// 离开
			msg := new(pb.LHLeaveReq)
			r.Sender(msg)
			// r.gamePid.Request(msg, r.pid)
		}
	case pb.STATE_DEALING:
		r.NowChip = 0
	case pb.STATE_BET:
		// 下注
		r.lh.betCount++
	}
}

// 下注
func (r *RoleActor) lhFreeBet() {
	if r.lh.lhstrategy == nil {
		return
	}
	chips := r.lh.lhstrategy.Chips
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
				c2s := new(pb.LHFreeBetReq)
				c2s.Seat = seat
				c2s.Value = chips[i]
				if r.gamePid != nil {
					r.Sender(c2s)
					// r.gamePid.Request(c2s, r.pid)
					r.nextTime = utils.LocalTime().UnixMilli() + time.Millisecond.Milliseconds()*500
				}
				break
			}
		}
		return
	}
	if r.lh.timer >= r.lh.waitTime {
		r.lh.timer = 0
		r.lh.waitTime = utils.RandInt32N(11) + 20
	} else {
		r.lh.timer++
		return
	}
	// 随机下龙虎和
	r.NowSeat = r.getLHSeat()
	// 下注金额
	r.NowChip = r.getLHChip()
	if uint32(r.Coin) < r.NowChip {
		// 筹码不够
		return
	}
	// glog.Infof("user %s lh bet seat:%d score:%d ", r.Userid, r.NowSeat, r.NowChip)

}

// 根据人机策略选择下注
func (c *RoleActor) getLHSeat() uint32 {
	var choices []utils.Choice
	if c.lh.lhstrategy == nil || len(c.lh.lhstrategy.BetWeight) == 0 {
		return 1
	}
	weights := c.lh.lhstrategy.BetWeight
	choices = append(choices, utils.Choice{Weight: int(weights[0]), Item: 1}) // 龙
	choices = append(choices, utils.Choice{Weight: int(weights[1]), Item: 2}) // 虎
	choices = append(choices, utils.Choice{Weight: int(weights[2]), Item: 0}) // 和
	choice, err := utils.WeightedChoice(choices)
	if err != nil {
		return 1
	}
	return uint32(choice.Item.(int))
}

// 根据人机策略选择筹码
func (c *RoleActor) getLHChip() uint32 {
	// 根据概率判断下不下注TODO
	pro := []int32{25, 75}
	if c.lh.lhstrategy != nil {
		pro = c.lh.lhstrategy.BetPro
	}
	var choices []utils.Choice
	choices = append(choices, utils.Choice{Weight: int(pro[0]), Item: 0}) // 不下注
	choices = append(choices, utils.Choice{Weight: int(pro[1]), Item: 1}) // 下注
	choice, err := utils.WeightedChoice(choices)
	if err == nil && choice.Item == 0 {
		return 0
	}
	// 选择筹码
	betArea := c.lh.lhstrategy.BetArea
	base := c.lh.lhstrategy.Min
	bet := utils.RandInt32N(betArea[1]/base-betArea[0]/base) + betArea[0]/base
	return uint32(bet * base)
}
