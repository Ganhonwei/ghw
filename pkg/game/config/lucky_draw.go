package config

import "goserver/pkg/data"

// 小米手机活动
var LuckyDrawRewards []*data.LuckyDrawReward
var LuckyDrawRobots []*data.LuckyDrawRobot

func InitLuckyDraw() {
	LuckyDrawRewards = data.GetLuckyDrawRewardList()
	LuckyDrawRobots = data.GetLuckyDrawRobotList()
}

func SetLuckyDrawRewards(luckyDrawRewards []*data.LuckyDrawReward) {
	LuckyDrawRewards = luckyDrawRewards
}
func SetLuckyDrawRobots(luckyDrawRobots []*data.LuckyDrawRobot) {
	LuckyDrawRobots = luckyDrawRobots
}

func GetLuckyDrawRewards() []*data.LuckyDrawReward {
	return LuckyDrawRewards
}

func GetLuckyDrawRobots() []*data.LuckyDrawRobot {
	return LuckyDrawRobots
}
