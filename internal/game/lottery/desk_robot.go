package lottery

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/algo"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
)

func (t *Desk) robotHandler(userid string) {
	role := t.roles[userid]
	// 机器人逻辑
	if role == nil || !role.GetRobot() {
		return
	}
	// 配置
	robot := t.Game.LOTTERY.Robot
	for {
		if robot.BetWeight == nil {
			continue
		}
		ntf := &pb.CPRobotStrategyNtf{
			InitScore:  robot.InitScore,
			BetWeight:  robot.BetWeight,
			TimeArea:   robot.TimeArea,
			Min:        robot.Min,
			BetArea:    robot.BetArea,
			BetPro:     robot.BetPro,
			LeaveLimit: robot.LeaveLimit,
			BetTimes:   robot.BetTimes,
			Chips:      t.Game.LOTTERY.ChipLimit,
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
	room := t.Game
	// if t.Dtype == int32(pb.DESK_TYPE_NORMAL_B) {
	// 	room = t.DeskData.Game
	// }
	if room.Id == "" {
		glog.Errorf("no find newbiew room config")
		return
	}
	for _, robot := range t.NewbiewRobot {
		if utils.RandWan(room.LOTTERY.Robot.BetPro[0]) {
			continue
		}
		chip := t.getChip(robot.Userid)
		if robot.Coin < chip || chip <= 0 {
			continue
		}
		t.CPRobotBets[robot.Userid] -= chip

		t.newBiewFreeBet(robot.Userid, t.getCPSeat(robot.Userid, room), chip)
	}
}

func (t *Desk) getChip(userid string) int64 {
	bet := t.CPRobotBets[userid]
	if _, ok := t.CPRobotBets[userid]; !ok {
		room := t.DeskData.Game.LOTTERY
		betInterval := room.Robot.BetArea
		bet = int64((utils.RandInt32N((betInterval[1]+betInterval[0])/room.Robot.Min) + betInterval[0]/room.Robot.Min) * room.Robot.Min)
		t.CPRobotBets[userid] = bet
	}

	weights := []int{24, 35, 21, 7, 4, 3, 2, 2, 1, 1}
	choices := make([]utils.Choice, 0)
	chips := t.DeskData.Game.LOTTERY.ChipLimit
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

func (t *Desk) getCPSeat(userid string, room data.Game) uint32 {
	if seat, ok := t.CPRobotSeat[userid]; ok {
		return seat
	}

	var choices []utils.Choice
	weights := room.LOTTERY.Robot.BetWeight // 下注权重
	if len(weights) < 6 {
		return algo.GaoPai
	}
	for i, v := range weights {
		choices = append(choices, utils.Choice{Weight: int(v), Item: uint32(i + 1)})
	}
	choice, err := utils.WeightedChoice(choices)
	if err != nil {
		t.CPRobotSeat[userid] = algo.GaoPai
		return algo.GaoPai
	}
	t.CPRobotSeat[userid] = choice.Item.(uint32)
	return choice.Item.(uint32)
}
