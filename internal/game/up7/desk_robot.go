package up7

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
	robot := t.Game.UP.Robot
	for {
		if robot.BetWeight == nil {
			continue
		}
		ntf := &pb.UPRobotStrategyNtf{
			InitScore:  robot.InitScore,
			BetWeight:  robot.BetWeight,
			TimeArea:   robot.TimeArea,
			Min:        robot.Min,
			BetArea:    robot.BetArea,
			BetPro:     robot.BetPro,
			LeaveLimit: robot.LeaveLimit,
			BetTimes:   robot.BetTimes,
			Chips:      t.Game.UP.ChipLimit,
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

	// room := config.GetGame("40021")
	// if t.Dtype == int32(pb.DESK_TYPE_NORMAL_B) {
	// 	room = t.DeskData.Game
	// }

	up7 := table.GetTables().Up7Table.Get(1)
	up7Robot := table.GetTables().Up7RobotTable.Get()
	betPro := up7Robot.BetPro         // 下注概率
	betChipPro := up7Robot.BetChipPro // 下注筹码概率
	betSeatPro := up7Robot.BetSeatPro // 下注位置概率

	for _, robot := range t.NewbiewRobot {
		if utils.RandInt32N(betPro[0]+betPro[1]) < betPro[0] {
			continue
		}

		// 下注筹码
		chipIndex, _ := utils.ChoiceInt32Index(betChipPro)
		chip := int64(up7.Chips[0].Value[chipIndex])
		if robot.Coin < chip || chip <= 0 {
			continue
		}
		// t.LHRobotBets[robot.Userid] -= chip

		// 下注位置
		seatIndex, _ := utils.ChoiceInt32Index(betSeatPro)
		seat := uint32(seatIndex)

		// 下注
		t.newBiewFreeBet(robot.Userid, seat, chip)
	}
}

// 根据人机策略选择下注
func (t *Desk) _getUPSeat(userid string, room data.Game) uint32 {
	if seat, ok := t.LHRobotSeat[userid]; ok {
		return seat
	}

	var choices []utils.Choice
	weights := room.UP.Robot.BetWeight                                           // 下注权重
	choices = append(choices, utils.Choice{Weight: int(weights[0]), Item: DOWN}) // XIAO
	choices = append(choices, utils.Choice{Weight: int(weights[1]), Item: UP})   // DA
	choices = append(choices, utils.Choice{Weight: int(weights[2]), Item: Tie})  // 和
	choice, err := utils.WeightedChoice(choices)
	if err != nil {
		t.LHRobotSeat[userid] = DOWN
		return DOWN
	}
	t.LHRobotSeat[userid] = choice.Item.(uint32)
	return choice.Item.(uint32)
}

func (t *Desk) _getChip(userid string) int64 {
	bet := t.LHRobotBets[userid]
	if _, ok := t.LHRobotBets[userid]; !ok {
		room := t.DeskData.Game.UP
		betInterval := room.Robot.BetArea
		bet = int64((utils.RandInt32N((betInterval[1]+betInterval[0])/room.Robot.Min) + betInterval[0]/room.Robot.Min) * room.Robot.Min)
		t.LHRobotBets[userid] = bet
	}

	weights := []int{24, 35, 21, 7, 4, 3, 2, 2, 1, 1}
	choices := make([]utils.Choice, 0)
	chips := t.DeskData.Game.UP.ChipLimit
	for i, v := range chips {
		if bet >= int64(v) {
			choices = append(choices, utils.Choice{Weight: weights[i], Item: v})
		}
	}

	choice, err := utils.WeightedChoice(choices)
	if err != nil {
		return 0
	}
	return int64(choice.Item.(uint32))
}
