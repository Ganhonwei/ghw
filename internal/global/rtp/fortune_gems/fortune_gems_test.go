package fortunegems

import (
	"fmt"
	"goserver/pkg/table"
	"os"
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
)

func SymbolToString(s Symbol) string {
	switch s {
	case J:
		return "J"
	case Q:
		return "Q"
	case K:
		return "K"
	case A:
		return "A"
	case GREEN:
		return "绿宝石"
	case BLUE:
		return "蓝宝石"
	case RED:
		return "红宝石"
	case WILD:
		return "百搭"
	case ONE:
		return "1x"
	case TWO:
		return "2x"
	case THREE:
		return "3x"
	case FIVE:
		return "5x"
	case TEN:
		return "10x"
	case FIFTEEN:
		return "15x"
	default:
		return "?"
	}
}

func pad(s string, width int) string {
	w := runewidth.StringWidth(s)
	if w < width {
		return s + strings.Repeat(" ", width-w)
	}
	return s
}

func PrintReelsWithMultiplier(reels map[int32]Symbol) {
	fmt.Println("转轴结果: ")
	fmt.Println("+--------+--------+--------+--------+")
	fmt.Printf("|%s|%s|%s|%s|\n",
		pad(" 轴1", 8), pad(" 轴2", 8), pad(" 轴3", 8), pad(" 倍数轴", 8))
	fmt.Println("+--------+--------+--------+--------+")
	for row := 0; row < 3; row++ {
		fmt.Printf("|%s|%s|%s|%s|\n",
			pad(SymbolToString(reels[int32(row+1)]), 8),
			pad(SymbolToString(reels[int32(row+4)]), 8),
			pad(SymbolToString(reels[int32(row+7)]), 8),
			pad(SymbolToString(reels[int32(row+10)]), 8),
		)
	}
	fmt.Println("+--------+--------+--------+--------+")
}

func TestFortuneGems2(t *testing.T) {

	err := os.Chdir("/root/work/goserver/run")
	if err != nil {
		t.Fatal(err)
	}

	//加载配置表
	err = table.LoadTables()
	if err != nil {
		panic(err)
	}

	bet := int64(100) // 假设押注100分（1元）
	extraBet := true
	score, reels, specificSymbol, winLines, winMultiplier, normalWin, _ := PlaySlot(bet, extraBet, map[int32]Symbol{}, "")
	PrintReelsWithMultiplier(reels)
	fmt.Printf("押注：%d\n", bet)
	fmt.Printf("扣分：%d\n", score)
	fmt.Printf("连线：%v\n", winLines)
	fmt.Printf("特定符号：%s\n", SymbolToString(specificSymbol))
	fmt.Printf("连线倍数：%d\n", winMultiplier)
	fmt.Printf("连线赢得：%d\n", normalWin)
}

func TestGetElementsByConfig(t *testing.T) {
	err := os.Chdir("/root/work/goserver/run")
	if err != nil {
		t.Fatal(err)
	}

	//加载配置表
	err = table.LoadTables()
	if err != nil {
		panic(err)
	}

	elements := GetElementsByConfig(1, true)
	fmt.Println(elements)
}

func TestRTPStatsByPlaySlot(t *testing.T) {
	err := os.Chdir("/root/work/goserver/run")
	if err != nil {
		t.Fatal(err)
	}

	// 加载配置表
	err = table.LoadTables()
	if err != nil {
		panic(err)
	}

	const totalRounds = 1000000
	const bet = int64(100)
	const extraBet = true

	var winRounds int64
	var totalBet int64
	var totalReward int64
	var rewards []int64

	var maxMultiplier float64
	for i := 0; i < totalRounds; i++ {
		score, _, _, _, _, normalWin, _ := PlaySlot(bet, extraBet, map[int32]Symbol{}, "")
		totalBet += score
		reward := normalWin
		totalReward += reward
		rewards = append(rewards, reward)
		if reward > 0 {
			winRounds++
			multiplier := float64(reward) / float64(score)
			if multiplier > maxMultiplier {
				maxMultiplier = multiplier
			}
		}
	}

	stats := CalcRTPStats(
		int64(totalRounds),
		winRounds,
		totalBet,
		totalReward,
		0, // 初始金额（可根据实际需求填写）
		0, // 剩余金额（可根据实际需求填写）
		rewards,
	)

	fmt.Printf("是否额外投注: %t\n", extraBet)
	fmt.Printf("总局数: %d\n", stats.TotalRounds)
	fmt.Printf("中奖次数: %d\n", stats.WinRounds)
	fmt.Printf("中奖率: %f\n", stats.WinRate())
	fmt.Printf("总押分: %d\n", stats.TotalBet)
	fmt.Printf("总奖赏: %d\n", stats.TotalReward)
	fmt.Printf("RTP: %f\n", stats.RTP())
	fmt.Printf("方差: %f\n", stats.Variance())
	fmt.Printf("标准差: %f\n", stats.StdDev())
	fmt.Printf("最大倍数: %f\n", maxMultiplier)
}

func TestPlaySlotWithValidInputReels(t *testing.T) {
	err := os.Chdir("/root/work/goserver/run")
	if err != nil {
		t.Fatal(err)
	}

	// 加载配置表
	err = table.LoadTables()
	if err != nil {
		panic(err)
	}

	inputReels := map[int32]Symbol{
		1:  J,
		2:  Q,
		3:  K,
		4:  A,
		5:  GREEN,
		6:  BLUE,
		7:  RED,
		8:  WILD,
		9:  J,
		10: ONE,  // 合法
		11: TWO,  // 合法
		12: FIVE, // 合法
	}

	// inputReels := map[int32]Symbol{
	// 	1:  J,
	// 	2:  Q,
	// 	3:  K,
	// 	4:  A,
	// 	5:  GREEN,
	// 	6:  BLUE,
	// 	7:  RED,
	// 	8:  WILD,
	// 	9:  J,
	// 	10: TWO,   // 合法
	// 	11: THREE, // 合法
	// 	12: FIVE,  // 合法
	// }

	bet := int64(100)
	extraBet := true
	userid := "6680306"

	score, reels, specificSymbol, winLines, winMultiplier, normalWin, _ := PlaySlot(bet, extraBet, inputReels, userid)

	PrintReelsWithMultiplier(reels)
	fmt.Printf("押注：%d\n", bet)
	fmt.Printf("扣分：%d\n", score)
	fmt.Printf("连线：%v\n", winLines)
	fmt.Printf("特定符号：%s\n", SymbolToString(specificSymbol))
	fmt.Printf("连线倍数：%d\n", winMultiplier)
	fmt.Printf("连线赢得：%d\n", normalWin)
}
