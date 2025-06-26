package lottery

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/handler"
	"goserver/pkg/game/login"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"strconv"
)

// '踢出不足坐下玩家或超额玩家
func (t *Desk) limitOver() {
	for k, v := range t.roles {
		ntf := new(pb.LotteryLeaveNtf)

		var err pb.ErrCode
		if !t.checkGameTime(v) {
			err = pb.KickGameTime
		} else if !t.checkCloseServer() {
			err = pb.KickCloseServer
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
	robot := t.Game.LOTTERY.Robot // 人机策略 TODO
	for k, v := range t.roles {
		if !v.GetRobot() {
			continue
		}
		score := v.GetScore() // 分数
		if score >= int64(robot.LeaveLimit[0]) && score <= int64(robot.LeaveLimit[1]) {
			// 分数区间正常
			continue
		}
		glog.Infof("robot %s score no enough, kickout room", k)
		errcode := t.leave(k, 0)
		if errcode != pb.OK {
			continue
		}
		// 通知自己回到大厅
		ntf := new(pb.LotteryLeaveNtf)
		t.send2userid(k, ntf)
		//离开状态消息
		t.userLeaveDesk(k)
	}
	for _, u := range t.NewbiewRobot {
		if u.Coin >= int64(robot.LeaveLimit[0]) && u.Coin <= int64(robot.LeaveLimit[1]) {
			// 分数区间正常
			continue
		}
		glog.Infof("newbiew robot %s score no enough, kickout room", u.Userid)
		// 换个名字和分数
		u.Nickname = login.RandName()
		u.Photo = strconv.Itoa(utils.RandIntN(30) + 1)
		u.Coin = int64(utils.RandMN(int(robot.InitScore[0]), int(robot.InitScore[1])))
		handler.RobotVip(u)
	}
}

// 踢除离线玩家
func (t *Desk) kickOffline() {
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE2): //百人
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

// 剔除点控房间多余玩家
func (t *Desk) kickPCMoreUser() {
	if t.Dtype != int32(pb.DESK_TYPE_POINTCONTROL) {
		return
	}
	r, _ := t.roleCountNum()
	if r <= 1 {
		return
	}
	for k, v := range t.roles {
		if v.GetRobot() {
			continue
		}
		errcode := t.leave(k, 0)
		if errcode != pb.OK {
			continue
		}
		// 通知自己回到大厅
		ntf := new(pb.LotteryLeaveNtf)
		t.send2userid(k, ntf)
		//离开状态消息
		t.userLeaveDesk(k)
		r--
		if r <= 1 {
			break
		}
	}
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
			// 新手房间
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
			glog.Infof("user state change,dtype:%d,state:%d", t.Dtype, role.State)
			t.leave(k, int32(err))
			// 通知自己回到大厅
			ntf := &pb.LotteryLeaveNtf{
				Error: err,
			}
			t.send2userid(k, ntf)
			//离开状态消息
			t.userLeaveDesk(k)
		}
	}
}

func (t *Desk) checkGameTime(role *data.DeskRole) bool {
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
