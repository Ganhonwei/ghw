package fortunegems

import (
	"math"
)

// RTP统计结构体
// 用于存储和计算RTP相关数据
// 可根据实际需求扩展字段

type RTPStats struct {
	TotalRounds int64   // 总游戏次数
	WinRounds   int64   // 中奖次数
	TotalBet    int64   // 总押分
	TotalReward int64   // 总奖赏
	StartAmount int64   // 初始金额
	EndAmount   int64   // 剩余金额
	Rewards     []int64 // 每局奖励（用于方差/标准差计算）
}

// 中奖率
func (r *RTPStats) WinRate() float64 {
	if r.TotalRounds == 0 {
		return 0
	}
	return float64(r.WinRounds) / float64(r.TotalRounds)
}

// 回报率（RTP）
func (r *RTPStats) RTP() float64 {
	if r.TotalBet == 0 {
		return 0
	}
	return float64(r.TotalReward) / float64(r.TotalBet)
}

// 统计方差
func (r *RTPStats) Variance() float64 {
	n := float64(len(r.Rewards))
	if n == 0 {
		return 0
	}
	mean := r.MeanReward()
	var sum float64
	for _, v := range r.Rewards {
		delta := float64(v) - mean
		sum += delta * delta
	}
	return sum / n
}

// 统计标准差
func (r *RTPStats) StdDev() float64 {
	return math.Sqrt(r.Variance())
}

// 平均奖励
func (r *RTPStats) MeanReward() float64 {
	n := float64(len(r.Rewards))
	if n == 0 {
		return 0
	}
	var sum float64
	for _, v := range r.Rewards {
		sum += float64(v)
	}
	return sum / n
}

// 统计函数示例
// 可在测试或主流程中调用
func CalcRTPStats(totalRounds, winRounds, totalBet, totalReward, startAmount, endAmount int64, rewards []int64) *RTPStats {
	return &RTPStats{
		TotalRounds: totalRounds,
		WinRounds:   winRounds,
		TotalBet:    totalBet,
		TotalReward: totalReward,
		StartAmount: startAmount,
		EndAmount:   endAmount,
		Rewards:     rewards,
	}
}
