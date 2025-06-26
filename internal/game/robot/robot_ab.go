package robot

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (a *RoleActor) matchabDesk(arg *pb.RobotMsg) *pb.MatchDesk {
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
func (a *RoleActor) ABLeaveNtf(ctx actor.Context) {
	ntf := ctx.Message().(*pb.ABLeaveNtf)
	// glog.Debugf("ABLeaveNtf %v", ntf)
	if ntf.Userid == a.Userid {
		// 关闭节点
		glog.Debugf("ABLeaveNtf %v", ntf)
		stop := new(pb.ServeStop)
		a.pid.Tell(stop)
	}
}

// 离开消息
func (a *RoleActor) ABLeaveRsp(ctx actor.Context) {
	ntf := ctx.Message().(*pb.ABLeaveRsp)
	if ntf.Userid == a.Userid {
		// 关闭节点
		glog.Debugf("ABLeaveRsp %v", ntf)
		stop := new(pb.ServeStop)
		a.pid.Tell(stop)
	}
}

func (a *RoleActor) ABFreeEnterRoomRsp(ctx actor.Context) {
	res := ctx.Message().(*pb.ABFreeEnterRoomRsp)
	glog.Debugf("ABFreeEnterRoomRsp %v", res)
	if res.Error != pb.OK {
		glog.Infof("ABFreeEnterRoomRsp enter err:%s", res.Error.String())
		stop := new(pb.ServeStop)
		a.pid.Tell(stop)
	}
}

func (a *RoleActor) abCheckStatus() {
	switch a.ab.state {
	case int32(pb.STATE_READY):
		a.ab.timer = 15
	case int32(pb.STATE_BET):
		// 下注
		a.abFreeBet()
	}
}

// 人机策略
func (r *RoleActor) recvABRobotStrategy(ctx actor.Context) {
	s2c := ctx.Message().(*pb.ABRobotStrategyNtf)
	r.ab.abstrategy = s2c
	r.ab.betCount = 0
	// 下注轮数
	if len(s2c.BetTimes) <= 0 {
		r.ab.maxBetCount = 50
		return
	}
	r.ab.maxBetCount = int32(utils.RandMN(int(s2c.BetTimes[0]), int(s2c.BetTimes[1])))
}

// 状态变化
func (r *RoleActor) recvABState(ctx actor.Context) {
	msg := ctx.Message().(*pb.ABPushStateNtf)
	r.ab.state = int32(msg.State)
	switch msg.State {
	case pb.STATE_READY:
		if r.ab.betCount >= r.ab.maxBetCount {
			// 离开
			msg := new(pb.ABLeaveReq)
			r.Sender(msg)
			// r.gamePid.Request(msg, r.pid)
		}
	case pb.STATE_DEALING:
		r.NowChip = 0
	case pb.STATE_BET:
		// 下注
		r.ab.betCount++
	}
}

// 下注
func (r *RoleActor) abFreeBet() {
	if r.ab.abstrategy == nil {
		return
	}
	if r.ab.nowChip == nil {
		r.ab.nowChip = make(map[uint32]uint32)
	}

	chips := r.ab.abstrategy.Chips
	for seat := range r.ab.nowChip {
		if r.ab.nowChip[seat] > 0 {
			// 没下完继续下
			// 拆分筹码,然后下注
			if utils.LocalTime().UnixMilli() < r.nextTime {
				return
			}
			if r.ab.nowChip[seat] < chips[0] {
				r.ab.nowChip[seat] = chips[0]
			}
			for i := len(chips) - 1; i >= 0; i-- {
				if r.ab.nowChip[seat] >= chips[i] {
					r.ab.nowChip[seat] -= chips[i]
					// 下注
					c2s := new(pb.ABFreeBetReq)
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
	}

	if r.ab.timer >= r.ab.waitTime {
		r.ab.timer = 0
		r.ab.waitTime = utils.RandInt32N(16) + 25
	} else {
		r.ab.timer++
		return
	}
	// 随机下注金额
	r.getABSeat()
	// glog.Infof("user %s ab bet seat:%d score:%d ", r.Userid, r.NowSeat, r.NowChip)

}

// 根据人机策略选择下注
func (c *RoleActor) getABSeat() {
	if c.ab.abstrategy == nil || len(c.ab.abstrategy.BetWeight) == 0 {
		return
	}
	c.ab.nowChip = make(map[uint32]uint32)

	weights := c.ab.abstrategy.BetWeight
	weights1 := c.ab.abstrategy.BetWeight2

	seat, _ := utils.ChoiceInt32Index(weights)
	sideSeat, _ := utils.ChoiceInt32Index(weights1)
	if seat > 0 {
		chip := c.getABChip()

		if chip > 0 && chip <= uint32(c.Coin) {
			c.ab.nowChip[uint32(seat)] = chip
		}
	}

	// 3-10位置
	if sideSeat > 0 {
		chip := c.getABChip()

		if chip > 0 && chip <= uint32(c.Coin) {
			c.ab.nowChip[uint32(sideSeat+2)] = chip
		}
	}
}

// 根据人机策略选择筹码
func (c *RoleActor) getABChip() uint32 {
	// 根据概率判断下不下注TODO
	pro := []int32{25, 75}
	if c.ab.abstrategy != nil {
		pro = c.ab.abstrategy.BetPro
	}
	var choices []utils.Choice
	choices = append(choices, utils.Choice{Weight: int(pro[0]), Item: 0}) // 不下注
	choices = append(choices, utils.Choice{Weight: int(pro[1]), Item: 1}) // 下注
	choice, err := utils.WeightedChoice(choices)
	if err == nil && choice.Item == 0 {
		return 0
	}
	// 选择筹码
	betArea := c.ab.abstrategy.BetArea
	base := c.ab.abstrategy.Min
	bet := utils.RandInt32N(betArea[1]/base-betArea[0]/base) + betArea[0]/base
	return uint32(bet * base)
}
