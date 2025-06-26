package lhd

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/config"
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
	robot := t.Game.LHD.Robot
	for {
		if robot.BetWeight == nil {
			continue
		}
		ntf := &pb.LHRobotStrategyNtf{
			InitScore:  robot.InitScore,
			BetWeight:  robot.BetWeight,
			TimeArea:   robot.TimeArea,
			Min:        robot.Min,
			BetArea:    robot.BetArea,
			BetPro:     robot.BetPro,
			LeaveLimit: robot.LeaveLimit,
			BetTimes:   robot.BetTimes,
			Chips:      t.Game.LHD.ChipLimit,
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
	room := config.GetGame("30021")
	if t.Dtype == int32(pb.DESK_TYPE_NORMAL_B) {
		room = t.DeskData.Game
	}
	if room.Id == "" {
		glog.Errorf("no find newbiew room config")
		return
	}
	for _, robot := range t.NewbiewRobot {
		if utils.RandWan(room.LHD.Robot.BetPro[0]) {
			continue
		}
		chip := t.getChip(robot.Userid)
		if robot.Coin < chip || chip <= 0 {
			continue
		}
		t.LHRobotBets[robot.Userid] -= chip

		// 下注
		t.newBiewFreeBet(robot.Userid, t.getLHDSeat(robot.Userid, room), chip)
	}
}

// 根据人机策略选择下注
func (t *Desk) getLHDSeat(userid string, room data.Game) uint32 {
	if seat, ok := t.LHRobotSeat[userid]; ok {
		return seat
	}

	var choices []utils.Choice
	weights := room.LHD.Robot.BetWeight                                            // 下注权重
	choices = append(choices, utils.Choice{Weight: int(weights[0]), Item: Dragon}) // 龙
	choices = append(choices, utils.Choice{Weight: int(weights[1]), Item: Tiger})  // 虎
	choices = append(choices, utils.Choice{Weight: int(weights[2]), Item: Tie})    // 和
	choice, err := utils.WeightedChoice(choices)
	if err != nil {
		t.LHRobotSeat[userid] = Dragon
		return Dragon
	}
	t.LHRobotSeat[userid] = choice.Item.(uint32)
	return choice.Item.(uint32)
}

func (t *Desk) getChip(userid string) int64 {
	bet := t.LHRobotBets[userid]
	if _, ok := t.LHRobotBets[userid]; !ok {
		room := t.DeskData.Game.LHD
		betInterval := room.Robot.BetArea
		bet = int64((utils.RandInt32N((betInterval[1]+betInterval[0])/room.Robot.Min) + betInterval[0]/room.Robot.Min) * room.Robot.Min)
		t.LHRobotBets[userid] = bet
	}

	weights := []int{24, 35, 21, 7, 4, 3, 2, 2, 1, 1}
	choices := make([]utils.Choice, 0)
	chips := t.DeskData.Game.LHD.ChipLimit
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
