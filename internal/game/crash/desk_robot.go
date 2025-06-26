package crash

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/table"
	"goserver/pkg/utils"
)

func (t *Desk) _robotHandler(userid string) {
	role := t.roles[userid]
	// 机器人逻辑
	if role == nil || !role.GetRobot() {
		return
	}
	// 配置
	// robot := table.GetTables().CrashRobotTable.Get()
	robot := t.Game.CRASH.Robot
	for {
		if robot.RobotLeave == nil {
			continue
		}
		ntf := &pb.CRASHRobotStrategyNtf{
			InitScore:  robot.InitScore,
			TimeArea:   robot.TimeArea,
			Min:        robot.Min,
			BetArea:    robot.BetArea,
			BetPro:     robot.BetPro,
			LeaveLimit: robot.LeaveLimit,
			BetTimes:   robot.BetTimes,
			Chips:      t.Game.CRASH.ChipLimit,
			RobotCrash: robot.RobotLeave,
		}
		t.send2userid(userid, ntf)
		break
	}

}

// 新手人机下注
func (t *Desk) robotBet() {
	if t.state != int32(pb.STATE_BET) {
		return
	}
	robotTable := table.GetTables().CrashRobotTable.Get()

	betPro0 := robotTable.BetPro1 // 下注概率
	betPro1 := robotTable.BetPro2 // 下注概率

	weights0 := robotTable.RobotChipPro1
	weights1 := robotTable.RobotChipPro2
	chips0 := table.GetTables().CrashTable.Get(1).Chips[0]
	chips1 := table.GetTables().CrashTable.Get(2).Chips[0]

	for _, robot := range t.NewbiewRobot {
		// 位置1
		if utils.RandInt32N(betPro0[0]+betPro0[1]) >= betPro0[0] {
			index, _ := utils.ChoiceInt32Index(weights0)
			chip := int64(chips0.Value[index])
			if robot.Coin >= chip {
				t.newBiewFreeBet(robot.Userid, int64(chip), 0)
			}
		}

		// 位置2
		if utils.RandInt32N(betPro1[0]+betPro1[1]) >= betPro1[0] {
			index, _ := utils.ChoiceInt32Index(weights1)
			chip := int64(chips1.Value[index])
			if robot.Coin >= chip {
				t.newBiewFreeBet(robot.Userid, int64(chip), 1)
			}
		}
	}
}

// Deprecated
func (t *Desk) _robotSettlement(id string, c data.Currency) bool {
	if user, ok := t.NewbiewRobot[id]; ok {
		if t.CRASHMulpitle == 5000 {
			m := float64(utils.RandInt32N(t.Game.CRASH.Jackpot[1]-t.Game.CRASH.Jackpot[0]) + t.Game.CRASH.Jackpot[0])
			user.Coin += int64(float64(c.Diamond) * m / 100)
			// 玩家输赢记录
			// t.CRASHRecord(k, back.GetSum())
			// // 记录输赢分
			t.CRASHScoreMap[id] = c.Scale(float64(m) / 100)
		} else {
			t.CRASHScoreMap[id] = c.Scale(-1)
		}
		return true
	}
	return false
}

func (t *Desk) robotBack(id string, multiple, pos int32) {
	if user, ok := t.NewbiewRobot[id]; ok {
		crashBets := t.PosCrashBetsMap(pos)
		if c, ok := crashBets[id]; ok {
			// 玩家输赢分
			win := c.Scale(float64(multiple-1) / 100)
			// 返奖上限
			winLimit := table.GetTables().CrashTable.Get(pos + 1).WinLimit[0]
			if win.GetSum() > winLimit {
				win = data.Currency{Coin: winLimit}
			}

			user.Coin += win.GetSum()
			win.Sub(c)

			// 逃脱倍数
			crashBack := t.PosCrashBackMap(pos)
			crashBack[id] = multiple
			// 输赢分
			crashScoreMap := t.PosCrashScoreMap(pos)
			crashScoreMap[id] = win

			// glog.Debugf("robot back %s", user.Userid)
			// 输赢分
			// if user.FreeWinMap == nil {
			// 	user.FreeWinMap = make(map[int32][]data.FreeWin)
			// }
			// if wins, ok := user.FreeWinMap[int32(pb.CRASH)]; ok {
			// 	wins = append(wins, data.FreeWin{Wtype: 1, Score: win.GetSum()})
			// 	size := len(wins)
			// 	if size > 20 {
			// 		wins = wins[size-20:]
			// 	}
			// 	user.FreeWinMap[int32(pb.CRASH)] = wins
			// }
			// 通知其他玩家
			ntf := new(pb.CRASHBackNtf)
			ntf.Name = user.Nickname
			ntf.Multiple = multiple
			ntf.Pos = pos
			ntf.Userid = user.Userid
			ntf.Photo = user.Photo
			ntf.VipLv = int32(user.Vip.Lv)
			ntf.Bets = c.GetSum()
			ntf.Score = win.GetSum() + c.GetSum()
			t.broadcast4(ntf)
		}
	}
}
