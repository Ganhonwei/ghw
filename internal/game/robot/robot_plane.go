package robot

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (a *RoleActor) matchPLANEDesk(arg *pb.RobotMsg) *pb.MatchDesk {
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
func (a *RoleActor) PLANELeaveNtf(ctx actor.Context) {
	ntf := ctx.Message().(*pb.PLANELeaveNtf)
	glog.Debugf("PLANELeaveNtf %v", ntf)
	if ntf.Userid == a.Userid {
		// 关闭节点
		stop := new(pb.ServeStop)
		a.pid.Tell(stop)
	}
}

// 离开消息
func (a *RoleActor) PLANELeaveRsp(ctx actor.Context) {
	ntf := ctx.Message().(*pb.PLANELeaveRsp)
	glog.Debugf("PLANELeaveRsp %v", ntf)
	if ntf.Userid == a.Userid {
		// 关闭节点
		stop := new(pb.ServeStop)
		a.pid.Tell(stop)
	}
}

// RobotLeaveNtf 通知人机主动离开房间
func (a *RoleActor) RobotLeaveNtf(ctx actor.Context) {
	ntf := ctx.Message().(*pb.RobotLeaveNtf)
	glog.Debugf("RobotLeaveNtf %v", ntf)
	if ntf.RobotId == a.Userid {
		switch a.gtype {
		case int32(pb.HUA):
			if ntf.LeaveMs > 0 {
				c2s := new(pb.JHLeaveReq)
				a.SendDefer2(c2s, int64(ntf.LeaveMs))
			} else {
				a.sendJHLeaveReq()
			}
		case int32(pb.HUA2):
			a.sendJHLeaveReq()
		case int32(pb.RUMMY), int32(pb.RUMMY2):
			a.sendRMLeaveReq()
		case int32(pb.AK47):
			a.sendAK47LeaveReq()
		case int32(pb.JOKER):
			a.sendJOKERLeaveReq()
		}
	}
}

func (a *RoleActor) PLANEEnterRoomRsp(ctx actor.Context) {
	res := ctx.Message().(*pb.PLANEEnterRoomRsp)
	glog.Debugf("PLANEEnterRoomRsp ")
	if res.Error != pb.OK {
		glog.Infof("PLANEEnterRoomRsp enter err:%s", res.Error.String())
		stop := new(pb.ServeStop)
		a.pid.Tell(stop)
	}
}

// 人机策略
func (r *RoleActor) recvPLANERobotStrategy(ctx actor.Context) {
	s2c := ctx.Message().(*pb.PLANERobotStrategyNtf)
	r.plane.cashstrategy = s2c
	r.plane.betCount = 0
	// 下注轮数
	if len(s2c.BetTimes) <= 0 {
		r.plane.maxBetCount = 50
		return
	}
	r.plane.maxBetCount = int32(utils.RandMN(int(s2c.BetTimes[0]), int(s2c.BetTimes[1])))
}

// 爆炸通知
func (r *RoleActor) PLANEBoomNtf(ctx actor.Context) {
	// msg := ctx.Message().(*pb.PLANEBoomNtf)
	r.plane.boom = true
}

// 状态变化
func (r *RoleActor) recvPLANEState(ctx actor.Context) {
	msg := ctx.Message().(*pb.PLANEPushStateNtf)
	r.plane.state = int32(msg.State)
	switch msg.State {
	case pb.STATE_READY:
		if r.plane.betCount >= r.plane.maxBetCount {
			// 离开
			msg := new(pb.PLANELeaveReq)
			r.Sender(msg)
		}
		r.plane.back = false
		r.plane.boom = false
	case pb.STATE_DEALING:
		r.NowChip = 0
	case pb.STATE_BET:
		r.timer = 10
		r.plane.betCount++
	}
}

// 下注
func (c *RoleActor) planeFreeBet() {
	// glog.Debugf("lhdFreeBet %s", c.roomid)
	if c.plane.state != int32(pb.STATE_BET) {
		// 不在下注状态
		return
	}
	if c.NowChip > 0 {
		// 没下完继续下
		// 拆分筹码,然后下注
		c.planeBet(true)
		return
	}
	if c.plane.timer > c.plane.waitTime {
		c.plane.timer = 0
		c.plane.waitTime = utils.RandInt32N(20) + 10
	} else {
		c.plane.timer++
		return
	}
	// 下注金额
	c.NowChip = c.getPLANEChip()
	if c.NowChip <= 0 {
		// utils.SleepRand(3)
		// c.planeFreeBet()
		return
	}
	if uint32(c.Diamond+c.Coin) < c.NowChip {
		// 筹码不够
		return
	}
	// glog.Infof("user %s plane bet seat:%d score:%d", c.Userid, c.NowSeat, c.NowChip)
	// 拆分筹码,然后下注
	// c.planeBet(false)
}

// 拆分筹码下注
func (c *RoleActor) planeBet(old bool) {
	if c.plane.cashstrategy == nil {
		return
	}
	chips := c.plane.cashstrategy.Chips
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
			c2s := new(pb.PLANEBetReq)
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
func (c *RoleActor) getPLANEChip() uint32 {
	// 根据概率判断下不下注TODO
	pro := c.plane.cashstrategy.BetPro
	var choices []utils.Choice
	choices = append(choices, utils.Choice{Weight: int(pro[0]), Item: 0}) // 不下注
	choices = append(choices, utils.Choice{Weight: int(pro[1]), Item: 1}) // 下注
	choice, err := utils.WeightedChoice(choices)
	if err == nil && choice.Item == 0 {
		return 0
	}
	// 选择筹码
	betArea := c.plane.cashstrategy.BetArea
	base := c.plane.cashstrategy.Min
	bet := utils.RandInt32N(betArea[1]/base-betArea[0]/base) + betArea[0]/base
	return uint32(bet * base)
}

func (a *RoleActor) planeCheckStatus() {
	switch a.plane.state {
	case int32(pb.STATE_LEAD):
		// 检测是否撤离
		a.planeBack()
	case int32(pb.STATE_BET):
		// 下注
		a.planeFreeBet()
	}
}

func (a *RoleActor) planeBack() {
	if a.plane.boom {
		// 已经炸了
		return
	}
	if a.plane.cashstrategy == nil {
		return
	}
	plane := a.plane.cashstrategy.RobotCrash
	if !a.plane.back {
		// 计算是否逃离
		if utils.RandWan(plane[1]) {
			msg := new(pb.PLANEBackReq)
			a.Sender(msg)
			a.plane.back = true
		}
	}
}
