package config

import "goserver/pkg/data"

//

// 提现排行榜配置
var RankWithdraws []*data.RankWithdraw

// InitPvpRoom 启动初始化
func InitRankWithdraw() {
	RankWithdraws = data.GetRankWithdraw()
}

// ----------------------对战房配置----------------------------------
func SetRankWithdraw(rankWithdraw []*data.RankWithdraw) {
	RankWithdraws = rankWithdraw
}

func GetRankWithdraw() []*data.RankWithdraw {
	return RankWithdraws
}
