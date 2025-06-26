package plane

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/config"
	"goserver/pkg/game/handler"
	"goserver/pkg/game/login"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"strconv"
)

// '踢出不足坐下玩家或超额玩家
func (t *Desk) limitOver() {
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE0): //自由
	case int32(pb.ROOM_TYPE1): //私人
		if !t.DeskData.Pub {
			//return
		}
	case int32(pb.ROOM_TYPE2): //百人
	}
	for k, v := range t.roles {
		ntf := &pb.PLANELeaveNtf{Userid: k}

		var err pb.ErrCode
		// if !t.checkGameTime(v) {
		// 	err = pb.KickGameTime
		// }
		if !t.checkCloseServer() {
			err = pb.KickCloseServer
		} else if !t.check5RoundNoBet(v) {
			err = pb.NoBet5Round
		}
		if err == pb.OK {
			continue
		}

		errcode := t.leave(k, int32(err))
		if errcode != pb.OK {
			continue
		}

		ntf.Error = err
		t.send2userid(k, ntf)

		//离开状态消息
		t.userLeaveDesk(k)
	}
}

// 剔除不满足条件的机器人
func (t *Desk) kickRobot() {
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE0), //自由
		int32(pb.ROOM_TYPE1), //私人
		int32(pb.ROOM_TYPE2): //百人
		robot := table.GetTables().PlaneRobotTable.Get() // 人机策略
		for k, v := range t.roles {
			if !v.GetRobot() {
				continue
			}
			score := v.GetScore() // 分数
			if score >= int64(robot.LeaveLimit[0]) && score <= int64(robot.LeaveLimit[1]) {
				// 分数区间正常
				continue
			}
			errcode := t.leave(k, 0)
			if errcode != pb.OK {
				continue
			}
			//离开状态消息
			t.userLeaveDesk(k)
		}
		var kicks []*data.User
		for _, u := range t.NewbiewRobot {
			if u.Coin >= int64(robot.LeaveLimit[0]) && u.Coin <= int64(robot.LeaveLimit[1]) {
				// 分数区间正常
				continue
			}
			glog.Infof("newbiew robot %s score no enough, kickout room", u.Userid)
			kicks = append(kicks, u)
		}

		for _, u := range kicks {
			prevUserid := u.Userid
			// 还头像
			t.returnHead(u)
			// 换个名字和分数
			u.Userid = fmt.Sprintf("%d", handler.GenerateOrderId(1))
			u.Nickname = login.RandName()
			u.Photo = strconv.Itoa(utils.RandIntN(30) + 1)
			u.Coin = int64(utils.RandMN(int(robot.InitScore[0]), int(robot.InitScore[1])))
			handler.RobotVip(u)
			t.getHead(u)

			delete(t.NewbiewRobot, prevUserid)
			t.NewbiewRobot[u.Userid] = u
			//离开消息
			t.broadcast(&pb.PLANELeaveNtf{Userid: prevUserid})
			// 进入消息
			t.freeCameinMsg(u.Userid)
		}
	}
}

// 踢除离线玩家
func (t *Desk) kickOffline() {
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE0), //自由
		int32(pb.ROOM_TYPE1), //私人
		int32(pb.ROOM_TYPE2): //百人
		for k, v := range t.roles {
			if !v.Offline {
				continue
			}
			errcode := t.leave(k, 0)
			if errcode != pb.OK {
				continue
			}
			//离开状态消息
			t.userLeaveDesk(k)
		}
	}
}

// 剔除5局不下注的玩家
func (t *Desk) kickNoBet() {
	for k, v := range t.roles {
		if v.GetRobot() {
			continue
		}
		if v.NoBetTimes < 5 {
			continue
		}
		errcode := t.leave(k, int32(pb.NoBet5Round))
		if errcode != pb.OK {
			continue
		}
		// 通知自己回到大厅
		ntf := &pb.PLANELeaveNtf{
			Userid: k,
			Error:  pb.NoBet5Round,
		}
		t.send2userid(k, ntf)
		//离开状态消息
		t.userLeaveDesk(k)
		t.offlineMsg(k)
	}
}

// 状态检测
func (t *Desk) stateCheck() {
	for k, role := range t.roles {
		if role.Robot {
			continue
		}

		kick := false
		var err pb.ErrCode
		switch t.Dtype {
		case int32(pb.DESK_TYPE_NEWBIEW):
			// 新手、平民房间
			kick = role.State != data.NoveiceState && role.State != data.ExceptionState && role.State != data.FrothState
			err = pb.NewbieKick
		case int32(pb.DESK_TYPE_NORMAL):
			// 正常房间
			kick = role.State != data.NormalState || role.PCSwitch
			// kick = role.PCSwitch
		case int32(pb.DESK_TYPE_NORMAL_B):
			// 正常房间
			kick = role.State != data.NormalState || role.PCSwitch || role.RegistArea != 1
		}
		if kick {
			errcode := t.leave(k, int32(err))
			if errcode != pb.OK {
				continue
			}
			// 通知自己回到大厅
			ntf := &pb.PLANELeaveNtf{
				Userid: k,
			}
			t.send2userid(k, ntf)
			//离开状态消息
			t.userLeaveDesk(k)
			t.offlineMsg(k)
		}
	}
}

// pub房间人数为0时解散
func (t *Desk) _checkPubOver() {
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE1): //私人
		if !t.DeskData.Pub {
			//return
		}
	default:
		//return
	}
	if len(t.roles) != 0 {
		return
	}
	g := config.GetGame(t.DeskData.Unique)
	if g.Id == t.DeskData.Unique {
		return //配置房间不关闭
	}
	//停止服务
	msg1 := new(pb.ServeStop)
	t.selfPid.Tell(msg1)
}

func (t *Desk) _randomMulpitle() int32 {
	choices := make([]utils.Choice, 0)
	fly := t.Game.PLANE.FlyMultiple
	if len(fly) <= 0 {
		return utils.RandInt32N(4901) + 100
	}
	for i, v := range fly[1] {
		choices = append(choices, utils.Choice{Weight: int(v), Item: i})
	}
	c, _ := utils.WeightedChoice(choices)
	index := c.Item.(int)
	max := fly[0][index]
	var min int32
	if index == 0 {
		min = 100
	} else {
		min = fly[0][index-1]
	}
	return utils.RandInt32N(max-min) + min
}

func (t *Desk) check105(role *data.DeskRole) bool {
	// if role.Kick105Flag {
	// 	return true
	// }
	// if role.OutDiamond >= 105*100 && role.Money == 0 {
	// 	return false
	// } else {
	// 	return true
	// }
	return true
}

func (t *Desk) check200(role *data.DeskRole) bool {
	// if role.Kick200Flag {
	// 	return true
	// }
	// if role.OutDiamond >= 200*100 && role.Money == 0 {
	// 	return false
	// } else {
	// 	return true
	// }
	return true
}

func (t *Desk) _checkGameTime(role *data.DeskRole) bool {
	if role.State == 3 {
		return true
	}
	bean := table.GetTables().NewbieTable.Get()
	if bean == nil {
		return true
	}

	if role.TotalGameTime >= uint64(bean.Transfer[0])*60 && role.Money == 0 {
		return false
	} else {
		return true
	}
}

func (t *Desk) checkCloseServer() bool {
	if t.closeServer {
		return false
	} else {
		return true
	}

}

func (t *Desk) check5RoundNoBet(role *data.DeskRole) bool {
	if role.GetRobot() {
		return true
	}
	if role.NoBetTimes >= 100 {
		return false
	}
	return true
}

func (t *Desk) _Boom() bool {
	boom := handler.CrashIsBoom(t.DeskFree.CRASHMulpitle, t.Game.PLANE.BackRate, 0, r) // 默认直接爆炸,在多个策略的情况下，按照顺序执行策略，前一个不爆的情况下执行再后续策略
	for _, s := range t.CRASHStrategys {
		switch s.Id {
		case FYZS:
			// 扶摇直上
			t.CRASHTriggerStrategys[s.Id] = s
			return t.fyzsStrategy()
		case YHWM:
			// 欲薅无门
			t.CRASHTriggerStrategys[s.Id] = s
			return t.yhwmStrategy()
		case QSHS:
			t.CRASHTriggerStrategys[s.Id] = s
			return t.qshsStrategy()
		case XJQB:
			// 是否触发
			if tigger, b := t.xjqbStrategy(); tigger {
				t.CRASHTriggerStrategys[s.Id] = s
				return b
			}
		case JCFK:
			// 是否触发
			if tigger, b := t.jcfkStrategy(); tigger {
				t.CRASHTriggerStrategys[s.Id] = s
				return b
			}
		case MXJL:
			// 冒险奖励
			t.CRASHTriggerStrategys[s.Id] = s
			return t.mxjlStrategy()
		case RKYH:
			// 是否触发
			if tigger, b := t.rkyhStrategy(); tigger {
				t.CRASHTriggerStrategys[s.Id] = s
				return b
			}
		case GCYX:
			// 高潮涌现
			t.CRASHTriggerStrategys[s.Id] = s
			return t.gcyxStrategy()
		}
	}

	return boom
}

//.

// 抽水
// func (t *Desk) drawcoin(userid string, val int64) int64 {
// 	if val <= 0 {
// 		return val
// 	}
// 	var num int64 = handler.DrawCoin(t.DeskData.Rtype, t.DeskData.Mode, val)
// 	switch t.DeskData.Rtype {
// 	case int32(pb.ROOM_TYPE2): //百人
// 		//反佣和收益消息,抽成日志记录 val - num
// 		// msg2 := handler.AgentProfitNumMsg(userid, t.DeskData.Gtype, num)
// 		// t.send3userid(userid, msg2)
// 	}
// 	return val - num
// }

//.

// vim: set foldmethod=marker foldmarker=//',//.:
