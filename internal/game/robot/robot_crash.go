package robot

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (a *RoleActor) matchCRASHDesk(arg *pb.RobotMsg) *pb.MatchDesk {
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
func (a *RoleActor) CRASHLeaveNtf(ctx actor.Context) {
	ntf := ctx.Message().(*pb.CRASHLeaveNtf)
	glog.Debugf("CRASHLeaveNtf %v", ntf)
	if ntf.Userid == a.Userid {
		// 关闭节点
		stop := new(pb.ServeStop)
		a.pid.Tell(stop)
	}
}

// 离开消息
func (a *RoleActor) CRASHLeaveRsp(ctx actor.Context) {
	ntf := ctx.Message().(*pb.CRASHLeaveRsp)
	glog.Debugf("CRASHLeaveRsp %v", ntf)
	if ntf.Userid == a.Userid {
		// 关闭节点
		stop := new(pb.ServeStop)
		a.pid.Tell(stop)
	}
}

func (a *RoleActor) CRASHEnterRoomRsp(ctx actor.Context) {
	res := ctx.Message().(*pb.CRASHEnterRoomRsp)
	glog.Debugf("CRASHEnterRoomRsp %v", res)
	if res.Error != pb.OK {
		glog.Infof("CRASHEnterRoomRsp enter err:%s", res.Error.String())
		stop := new(pb.ServeStop)
		a.pid.Tell(stop)
	}
}

// 人机策略
func (r *RoleActor) recvCRASHRobotStrategy(ctx actor.Context) {
	s2c := ctx.Message().(*pb.CRASHRobotStrategyNtf)
	r.crash.cashstrategy = s2c
	r.crash.betCount = 0
	// 下注轮数
	if len(s2c.BetTimes) <= 0 {
		r.crash.maxBetCount = 50
		return
	}
	r.crash.maxBetCount = int32(utils.RandMN(int(s2c.BetTimes[0]), int(s2c.BetTimes[1])))
}

// 爆炸通知
func (r *RoleActor) CRASHBoomNtf(ctx actor.Context) {
	// msg := ctx.Message().(*pb.CRASHBoomNtf)
	r.crash.boom = true
}

// 状态变化
func (r *RoleActor) recvCRASHState(ctx actor.Context) {
	msg := ctx.Message().(*pb.CRASHPushStateNtf)
	r.crash.state = int32(msg.State)
	switch msg.State {
	case pb.STATE_READY:
		if r.crash.betCount >= r.crash.maxBetCount {
			// 离开
			msg := new(pb.CRASHLeaveReq)
			r.Sender(msg)
		}
		r.crash.back = false
		r.crash.boom = false
	case pb.STATE_DEALING:
		r.NowChip = 0
	case pb.STATE_BET:
		r.timer = 10
		r.crash.betCount++
	}
}

// 下注
func (c *RoleActor) crashFreeBet() {
	// glog.Debugf("lhdFreeBet %s", c.roomid)
	if c.crash.state != int32(pb.STATE_BET) {
		// 不在下注状态
		return
	}
	if c.NowChip > 0 {
		// 没下完继续下
		// 拆分筹码,然后下注
		c.crashBet(true)
		return
	}
	if c.crash.timer > c.crash.waitTime {
		c.crash.timer = 0
		c.crash.waitTime = utils.RandInt32N(20) + 10
	} else {
		c.crash.timer++
		return
	}
	// 下注金额
	c.NowChip = c.getCRASHChip()
	if c.NowChip <= 0 {
		// utils.SleepRand(3)
		// c.crashFreeBet()
		return
	}
	if uint32(c.Diamond+c.Coin) < c.NowChip {
		// 筹码不够
		return
	}
	// glog.Infof("user %s crash bet seat:%d score:%d", c.Userid, c.NowSeat, c.NowChip)
	// 拆分筹码,然后下注
	// c.crashBet(false)
}

// 拆分筹码下注
func (c *RoleActor) crashBet(old bool) {
	if c.crash.cashstrategy == nil {
		return
	}
	chips := c.crash.cashstrategy.Chips
	if len(chips) <= 0 {
		return
	}
	if utils.LocalTime().UnixMilli() < c.nextTime {
		return
	}
	if c.NowChip < chips[0] {
		c.NowChip = chips[0]
	}
	for i := len(chips) - 1; i >= 0; i-- {
		if c.NowChip >= chips[i] {
			c.NowChip -= chips[i]
			// 下注
			c2s := new(pb.CRASHBetReq)
			c2s.Value = chips[i]
			if c.gamePid != nil {
				c.gamePid.Request(c2s, c.pid)
				c.nextTime = utils.LocalTime().UnixMilli() + time.Millisecond.Milliseconds()*500
			}
			// if !old {
			// 	// 间隔
			// 	c.SendDefer(c2s)
			// } else {
			// 	// 0.5ms
			// 	c.SendDefer2(c2s, 500)
			// }
			break
		}
	}

	// 延迟时间下注
	// utils.Sleep(utils.RandIntN(2) + 3) //随机
	// c.lhdFreeBet()
}

// 根据人机策略选择筹码
func (c *RoleActor) getCRASHChip() uint32 {
	// 根据概率判断下不下注TODO
	pro := c.crash.cashstrategy.BetPro
	var choices []utils.Choice
	choices = append(choices, utils.Choice{Weight: int(pro[0]), Item: 0}) // 不下注
	choices = append(choices, utils.Choice{Weight: int(pro[1]), Item: 1}) // 下注
	choice, err := utils.WeightedChoice(choices)
	if err == nil && choice.Item == 0 {
		return 0
	}
	// 选择筹码
	betArea := c.crash.cashstrategy.BetArea
	base := c.crash.cashstrategy.Min
	bet := utils.RandInt32N(betArea[1]/base-betArea[0]/base) + betArea[0]/base
	return uint32(bet * base)
}

func (a *RoleActor) crashCheckStatus() {
	switch a.crash.state {
	case int32(pb.STATE_LEAD):
		// 检测是否撤离
		a.crashBack()
	case int32(pb.STATE_BET):
		// 下注
		a.crashFreeBet()
	}
}

func (a *RoleActor) crashBack() {
	if a.crash.boom {
		// 已经炸了
		return
	}
	if a.crash.cashstrategy == nil {
		return
	}
	crash := a.crash.cashstrategy.RobotCrash
	if !a.crash.back {
		// 计算是否逃离
		if utils.RandWan(crash[1]) {
			msg := new(pb.CRASHBackReq)
			a.Sender(msg)
			a.crash.back = true
		}
	}
}
