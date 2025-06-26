package up7

import (
	"goserver/pkg/utils"
	"math"
	"math/rand"
)

type BetArea struct {
	ID          int
	ProbConfig  []ProbItem
	Description string
}

type ProbItem struct {
	Multiple int
	Prob     float64
}

// 示例：根据你的表格，填充13个区域的赔率配置
var betAreas = []BetArea{
	{1, []ProbItem{{1, 0.827}, {2, 0.137}, {3, 0.007}, {5, 0.018}, {10, 0.008}, {12, 0.002}, {15, 0.001}}, "7以下down"},
	{2, []ProbItem{{4, 0.86699}, {8, 0.108}, {12, 0.01}, {20, 0.01}, {40, 0.005}, {80, 0.00001}}, "7点"},
	{3, []ProbItem{{1, 0.827}, {2, 0.137}, {3, 0.007}, {5, 0.018}, {10, 0.008}, {12, 0.002}, {15, 0.001}}, "7以上up"},
	{4, []ProbItem{{26, 0.8364}, {52, 0.1125}, {78, 0.008}, {130, 0.0429}, {520, 0.0002}}, "2点"},
	{5, []ProbItem{{12, 0.7809}, {24, 0.1761}, {36, 0.004}, {60, 0.032}, {120, 0.007}, {144, 0}, {180, 0}, {240, 0}, {300, 0}}, "3点"},
	{6, []ProbItem{{8, 0.797}, {16, 0.156}, {24, 0.008}, {40, 0.038}, {80, 0.001}}, "4点"},
	{7, []ProbItem{{6, 0.83}, {12, 0.121}, {18, 0.03}, {30, 0.012}, {60, 0.007}}, "5点"},
	{8, []ProbItem{{5, 0.862}, {10, 0.12}, {15, 0.004}, {25, 0.01}, {50, 0.004}}, "6点"},
	{9, []ProbItem{{5, 0.862}, {10, 0.12}, {15, 0.004}, {25, 0.01}, {50, 0.004}}, "8点"},
	{10, []ProbItem{{6, 0.83}, {12, 0.121}, {18, 0.03}, {30, 0.012}, {60, 0.007}}, "9点"},
	{11, []ProbItem{{8, 0.797}, {16, 0.156}, {24, 0.008}, {40, 0.038}, {80, 0.001}}, "10点"},
	{12, []ProbItem{{12, 0.7809}, {24, 0.1761}, {36, 0.004}, {60, 0.032}, {120, 0.007}, {144, 0}, {180, 0}, {240, 0}, {300, 0}}, "11点"},
	{13, []ProbItem{{26, 0.8364}, {52, 0.1125}, {78, 0.008}, {130, 0.0429}, {520, 0.0002}}, "12点"},
}

const (
	SimulateTimes = 100000 // 每个区域模拟次数
)

// 下注金额区间
const (
	minBet = 100
	maxBet = 1000
)

// 玩家初始携带金额
const (
	initMoney = 100000000
)

// 掷骰子1
func rollDice1() int {
	dice1 := utils.RandInt32N(6) + 1
	dice2 := utils.RandInt32N(6) + 1
	return int(dice1 + dice2)
}

// 掷骰子2
func rollDice2() int {
	return int(utils.RandInt32N(11) + 2)
}

// 获取掷骰子结果
func getResult() int {
	return rollDice2()
}

// 判断是否赢
func isWin(areaID, result int) bool {
	switch areaID {
	case 1:
		return result >= 2 && result <= 6
	case 2:
		return result == 7
	case 3:
		return result >= 8 && result <= 12
	case 4:
		return result == 2
	case 5:
		return result == 3
	case 6:
		return result == 4
	case 7:
		return result == 5
	case 8:
		return result == 6
	case 9:
		return result == 8
	case 10:
		return result == 9
	case 11:
		return result == 10
	case 12:
		return result == 11
	case 13:
		return result == 12
	default:
		return false
	}
}

type SimResult struct {
	BetAreaID     int    //下注区域ID
	Description   string //描述
	TotalCount    int    //局数
	TotalWinCount int    //赢局数
	FinalMoney    int64  //结束分数
	StopReason    string //停止原因
	TotalBet      int64  //所有押分
	TotalRetBet   int64  //所有返还押分
	TotalWin      int64  //所有赢分
	TotalLose     int64  //所有输分

	WinRate float64 //中奖率
	RTP     float64 //RTP

	TotalWinLose int64 //总输赢
	TotalBaoji   int   //总暴击
}

func simulateArea(area BetArea) SimResult {
	totalBet := int64(0)
	totalRetBet := int64(0)
	totalWin := int64(0)
	totalLose := int64(0)
	totalWinCount := 0
	totalCount := 0
	currentMoney := int64(initMoney)
	stopReason := ""
	totalBaoji := 0

	for i := 0; i < SimulateTimes; i++ {
		// 检查当前金额是否足够下注
		if currentMoney < minBet {
			stopReason = "余额不足"
			break
		}

		// 生成下注金额，但不超过当前余额
		maxPossibleBet := min(int64(maxBet), currentMoney)
		bet := minBet + rand.Int63n(maxPossibleBet-minBet+1)

		totalCount++    //累加局数
		totalBet += bet //累加所有押分
		result := getResult()
		if isWin(area.ID, result) { //赢
			totalWinCount++    //累加所有赢局数
			totalRetBet += bet //累加所有返还押分
			base := area.ProbConfig[0].Multiple
			var choices []utils.Choice
			for _, prob := range area.ProbConfig {
				choices = append(choices, utils.Choice{Weight: int(math.Floor(prob.Prob * 100000)), Item: prob.Multiple})
			}
			c, _ := utils.WeightedChoice(choices)
			multiple := c.Item.(int)
			winAmount := bet * int64(multiple)
			totalWin += winAmount //累加所有赢分
			currentMoney += winAmount
			if multiple != base {
				totalBaoji++
			}
		} else { //输
			totalLose += bet //累加所有输分
			currentMoney -= bet
		}
	}

	if stopReason == "" && totalCount == SimulateTimes {
		stopReason = "达到最大游戏次数"
	}

	winRate := float64(totalWinCount) / float64(totalCount)
	totalWinLose := totalWin - totalBet + totalRetBet
	rtp := float64(totalWinLose+totalBet) / float64(totalBet)

	return SimResult{
		BetAreaID:     area.ID,
		Description:   area.Description,
		TotalCount:    totalCount,
		TotalWinCount: totalWinCount,
		FinalMoney:    currentMoney,
		StopReason:    stopReason,
		TotalBet:      totalBet,
		TotalRetBet:   totalRetBet,
		TotalWin:      totalWin,
		TotalLose:     totalLose,

		WinRate:      winRate,
		RTP:          rtp,
		TotalWinLose: totalWinLose,
		TotalBaoji:   totalBaoji,
	}
}
