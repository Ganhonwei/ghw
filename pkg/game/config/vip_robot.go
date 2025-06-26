package config

import (
	"goserver/pkg/data"
	"goserver/pkg/utils"
)

// 提现排行榜配置
var VipRobots []*data.VipRobot

// InitPvpRoom 启动初始化
func InitVipRobots() {
	VipRobots = data.GetVipRobots()
}

// ----------------------对战房配置----------------------------------
func SetVipRobot(vipRobot []*data.VipRobot) {
	VipRobots = vipRobot
}

func GetVipRobot() []*data.VipRobot {
	return VipRobots
}

func GetVipRobotChoices() (choices []utils.Choice) {
	if len(VipRobots) == 0 {
		return
	}
	for i, lv := range VipRobots[0].VipChoices {
		choices = append(choices, utils.Choice{Weight: VipRobots[0].VipWeights[i], Item: lv})
	}
	return
}
