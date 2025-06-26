package robot

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (a *RoleActor) matchUPDesk(arg *pb.RobotMsg) *pb.MatchDesk {
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
func (a *RoleActor) UPLeaveNtf(ctx actor.Context) {
	ntf := ctx.Message().(*pb.UPLeaveNtf)
	if ntf.Userid == a.Userid {
		// 关闭节点
		glog.Infof("UPLeaveNtf %v", ntf)
		stop := new(pb.ServeStop)
		a.pid.Tell(stop)
	}
}

// 离开消息
func (a *RoleActor) UPLeaveRsp(ctx actor.Context) {
	ntf := ctx.Message().(*pb.UPLeaveRsp)
	if ntf.Userid == a.Userid {
		// 关闭节点
		glog.Infof("UPLeaveRsp %v", ntf)
		stop := new(pb.ServeStop)
		a.pid.Tell(stop)
	}
}

func (a *RoleActor) UPFreeEnterRoomRsp(ctx actor.Context) {
	res := ctx.Message().(*pb.UPFreeEnterRoomRsp)
	glog.Debugf("UPFreeEnterRoomRsp %v", res)
	if res.Error != pb.OK {
		glog.Infof("UPFreeEnterRoomRsp enter err:%s", res.Error.String())
		stop := new(pb.ServeStop)
		a.pid.Tell(stop)
	}
}

func (a *RoleActor) upCheckStatus() {
	switch a.up.state {
	case int32(pb.STATE_READY):
		a.up.timer = 15
	case int32(pb.STATE_BET):
		// 下注
		a.upFreeBet()
	}
}

// 人机策略
func (r *RoleActor) recvUPRobotStrategy(ctx actor.Context) {
	s2c := ctx.Message().(*pb.UPRobotStrategyNtf)
	r.up.lhstrategy = s2c
	r.up.betCount = 0
	// 下注轮数
	if len(s2c.BetTimes) <= 0 {
		r.up.maxBetCount = 50
		return
	}
	r.up.maxBetCount = int32(utils.RandMN(int(s2c.BetTimes[0]), int(s2c.BetTimes[1])))
}

// 状态变化
func (r *RoleActor) recvUPState(ctx actor.Context) {
	msg := ctx.Message().(*pb.UPPushStateNtf)
	r.up.state = int32(msg.State)
	switch msg.State {
	case pb.STATE_READY:
		if r.up.betCount >= r.up.maxBetCount {
			// 离开
			msg := new(pb.UPLeaveReq)
			r.Sender(msg)
			// r.gamePid.Request(msg, r.pid)
		}
	case pb.STATE_DEALING:
		r.NowChip = 0
	case pb.STATE_BET:
		// 下注
		r.up.betCount++
	}
}

// 下注
func (r *RoleActor) upFreeBet() {
	if r.up.lhstrategy == nil {
		return
	}
	chips := r.up.lhstrategy.Chips
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
				c2s := new(pb.UPFreeBetReq)
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
	if r.up.timer >= r.up.waitTime {
		r.up.timer = 0
		r.up.waitTime = utils.RandInt32N(11) + 20
	} else {
		r.up.timer++
		return
	}
	// 随机下龙虎和
	r.NowSeat = r.getUPSeat()
	// 下注金额
	r.NowChip = r.getUPChip()
	if uint32(r.Coin) < r.NowChip {
		// 筹码不够
		return
	}
	// glog.Infof("user %s up bet seat:%d score:%d ", r.Userid, r.NowSeat, r.NowChip)

}

// 根据人机策略选择下注
func (c *RoleActor) getUPSeat() uint32 {
	var choices []utils.Choice
	if c.up.lhstrategy == nil || len(c.up.lhstrategy.BetWeight) == 0 {
		return 1
	}
	weights := c.up.lhstrategy.BetWeight
	choices = append(choices, utils.Choice{Weight: int(weights[0]), Item: 1}) // 小
	choices = append(choices, utils.Choice{Weight: int(weights[1]), Item: 2}) // 大
	choices = append(choices, utils.Choice{Weight: int(weights[2]), Item: 0}) // 和
	choice, err := utils.WeightedChoice(choices)
	if err != nil {
		return 1
	}
	return uint32(choice.Item.(int))
}

// 根据人机策略选择筹码
func (c *RoleActor) getUPChip() uint32 {
	// 根据概率判断下不下注TODO
	pro := []int32{25, 75}
	if c.up.lhstrategy != nil {
		pro = c.up.lhstrategy.BetPro
	}
	var choices []utils.Choice
	choices = append(choices, utils.Choice{Weight: int(pro[0]), Item: 0}) // 不下注
	choices = append(choices, utils.Choice{Weight: int(pro[1]), Item: 1}) // 下注
	choice, err := utils.WeightedChoice(choices)
	if err == nil && choice.Item == 0 {
		return 0
	}
	// 选择筹码
	betArea := c.up.lhstrategy.BetArea
	base := c.up.lhstrategy.Min
	bet := utils.RandInt32N(betArea[1]/base-betArea[0]/base) + betArea[0]/base
	return uint32(bet * base)
}
