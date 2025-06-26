package andarbahar

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
	robot := t.Game.AB.Robot
	for {
		if robot.BetWeight1 == nil {
			continue
		}
		ntf := &pb.ABRobotStrategyNtf{
			InitScore:  robot.InitScore,
			BetWeight:  robot.BetWeight1,
			BetWeight2: robot.BetWeight2,
			TimeArea:   robot.TimeArea,
			Min:        robot.Min,
			BetPro:     robot.BetPro,
			BetArea:    robot.BetArea,
			LeaveLimit: robot.LeaveLimit,
			BetTimes:   robot.BetTimes,
			Chips:      t.Game.AB.ChipLimit,
		}
		t.send2userid(userid, ntf)
		break
	}

}

// 新手人机下注
func (t *Desk) robotBet() {
	// if utils.RandWan(5000) {
	// 	return
	// }
	if t.state != int32(pb.STATE_BET) {
		return
	}
	room := config.GetGame("60021")
	if t.Dtype == int32(pb.DESK_TYPE_NORMAL_B) {
		room = t.DeskData.Game
	}
	if room.Id == "" {
		glog.Errorf("no find newbiew room config")
		return
	}

	for _, robot := range t.NewbiewRobot {
		seat, _ := utils.ChoiceInt32Index(room.AB.Robot.BetWeight1) // andar或bahar 0不下注
		if seat > 0 {
			t.robotBetBySeat(robot, uint32(seat))
		}
		sideSeat, _ := utils.ChoiceInt32Index(room.AB.Robot.BetWeight2) // 3-10位置 0不下注
		if sideSeat > 0 {
			t.robotBetBySeat(robot, uint32(sideSeat+2))
		}
	}
}

// 根据位置下注
func (t *Desk) robotBetBySeat(robot *data.User, seat uint32) {
	betPro := t.Game.AB.Robot.BetPro // 下注概率
	if utils.RandInt32N(betPro[0]+betPro[1]) < betPro[0] {
		// 不下注
		return
	}

	chips := t.Game.AB.ChipLimit
	choices := make([]utils.Choice, 0)
	weights := []int{24, 35, 21, 7, 4, 3, 2, 2, 1, 1}

	for i, v := range chips {
		choices = append(choices, utils.Choice{Weight: weights[i], Item: v})
	}
	// 选择筹码
	choice, _ := utils.WeightedChoice(choices)
	chip := int64(choice.Item.(uint32))
	if robot.Diamond < chip {
		return
	}

	// now := utils.BsonNow().UnixMilli()
	// msg := data.NewbiewFreeBetQeuee{
	// 	Robotid: robot.Userid,
	// 	BetTime: now + utils.RandInt64N(900) + 100,
	// 	Roomid:  roomid,
	// 	Chip:    chip,
	// 	Seat:    seat,
	// }
	// t.FreeDeskNewbiew.BetQeuee = append(t.FreeDeskNewbiew.BetQeuee, msg)
	t.newBiewFreeBet(robot.Userid, seat, chip)
}
