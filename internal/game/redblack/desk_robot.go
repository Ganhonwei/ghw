package redblack

import (
	"goserver/gen/pb"
	"goserver/pkg/table"
	"goserver/pkg/utils"
)

func (t *Desk) robotHandler(userid string) {
	role := t.roles[userid]
	// 机器人逻辑
	if role == nil || !role.GetRobot() {
		return
	}
	// 配置
	// robot := table.GetTables().RbRoomRobotTable.Get()
	// for {
	// 	if robot.BetWeight == nil {
	// 		continue
	// 	}
	// 	ntf := &pb.RBRobotStrategyNtf{
	// 		InitScore: robot.InitScore,
	// 		BetWeight: robot.BetWeight,
	// 		// TimeArea:   robot.TimeArea,
	// 		Min:        robot.BaseBet,
	// 		BetArea:    robot.BetInterval,
	// 		BetPro:     robot.BetProbility,
	// 		LeaveLimit: robot.RobotLeave,
	// 		BetTimes:   robot.RobotBetLeave,
	// 		Chips:      table.GetTables().RbRoomBaseTable.Get().ChipLimit,
	// 	}
	// 	t.send2userid(userid, ntf)
	// 	break
	// }

}

// 新手人机下注
func (t *Desk) robotBet() {
	if t.state != int32(pb.STATE_BET) {
		return
	}
	// room := table.GetTables().RbRoomRobotTable.Get()
	// betPro := room.BetProbility // 下注概率
	for _, robot := range t.NewbiewRobot {
		if utils.RandWan(9000) {
			continue
		}
		chip := t.getChip(robot.Userid)
		if robot.Coin < chip || chip <= 0 {
			continue
		}
		t.RBRobotBets[robot.Userid] -= chip

		// 下注
		t.newBiewFreeBet(robot.Userid, t.getRBSeat(robot.Userid), chip)
	}
}

func (t *Desk) getChip(userid string) int64 {
	bet := t.RBRobotBets[userid]
	if _, ok := t.RBRobotBets[userid]; !ok {
		room := table.GetTables().RbRoomRobotTable.Get()
		betInterval := room.BetInterval
		bet = int64((utils.RandInt32N((betInterval[1]+betInterval[0])/room.BaseBet) + betInterval[0]/room.BaseBet) * room.BaseBet)
		t.RBRobotBets[userid] = bet
	}

	weights := []int{10, 40, 40, 20, 15, 8, 5, 2, 1}
	choices := make([]utils.Choice, 0)
	chips := table.GetTables().RbRoomBaseTable.Get().ChipLimit
	for i, v := range chips {
		if bet >= int64(v) {
			choices = append(choices, utils.Choice{Weight: weights[i], Item: v})
		}
	}

	choice, err := utils.WeightedChoice(choices)
	if err != nil {
		return 0
	}
	return int64(choice.Item.(int32))
}

// 根据人机策略选择下注
func (t *Desk) getRBSeat(userid string) uint32 {
	if seat, ok := t.RBRobotSeat[userid]; ok {
		return seat
	}

	var choices []utils.Choice
	weights := table.GetTables().RbRoomRobotTable.Get().BetWeight                    // 下注权重
	choices = append(choices, utils.Choice{Weight: int(weights[0]), Item: RED})   // 红
	choices = append(choices, utils.Choice{Weight: int(weights[1]), Item: BLACK}) // 黑
	choices = append(choices, utils.Choice{Weight: int(weights[2]), Item: LUCK})  // 幸运一击
	choice, err := utils.WeightedChoice(choices)
	if err != nil {
		t.RBRobotSeat[userid] = RED
		return RED
	}
	t.RBRobotSeat[userid] = choice.Item.(uint32)
	return choice.Item.(uint32)
}
