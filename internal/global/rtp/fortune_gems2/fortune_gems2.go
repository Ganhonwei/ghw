package fortunegems2

import (
	"math"
	"math/rand"
)

type Symbol int

// 转轴列符号
const (
	J           Symbol = iota + 1 //J
	Q                             //Q
	K                             //K
	A                             //A
	GREEN                         //绿宝石
	BLUE                          //蓝宝石
	RED                           //红宝石
	WILD                          //百搭
	ONE                           //1倍
	TWO                           //2倍
	THREE                         //3倍
	FIVE                          //5倍
	TEN                           //10倍
	FIFTEEN                       //15倍
	GREEN_WHEEL                   //绿色WHEEL
	RED_WHEEL                     //红色WHEEL
)

// 普通转轴列符号
var NormalSymbols = []Symbol{J, Q, K, A, GREEN, BLUE, RED, WILD}

// 生成普通转轴列
func GenNormalReelColumn() [3]Symbol {
	var col [3]Symbol
	for row := 0; row < 3; row++ {
		col[row] = NormalSymbols[rand.Intn(len(NormalSymbols))]
	}
	return col
}

// 特殊转轴列符号
var SpecialSymbols = []Symbol{ONE, TWO, THREE, FIVE, TEN, FIFTEEN, GREEN_WHEEL}
var SpecialSymbolsExtra = []Symbol{TWO, THREE, FIVE, TEN, FIFTEEN, RED_WHEEL}

// 生成特殊转轴列
func GenSpecialReelColumn(extraBet bool) [3]Symbol {
	var genSymbols []Symbol
	if extraBet {
		genSymbols = SpecialSymbolsExtra
	} else {
		genSymbols = SpecialSymbols
	}

	var col [3]Symbol
	for row := 0; row < 3; row++ {
		col[row] = genSymbols[rand.Intn(len(genSymbols))]
	}
	return col
}

func IsWheel(symbol Symbol) bool {
	return symbol == GREEN_WHEEL || symbol == RED_WHEEL
}

func IsValidSymbol(symbol Symbol) bool {
	return symbol >= J && symbol <= RED_WHEEL
}

func IsInputReelsEmpty(inputReels map[int32]Symbol) bool {
	return len(inputReels) == 0
}

func IsInputReelsValid(inputReels map[int32]Symbol, extraBet bool) bool {
	if len(inputReels) != 12 {
		return false
	}
	for i := int32(1); i <= 12; i++ {
		symbol, ok := inputReels[i]
		if !ok {
			return false
		}
		if i >= 1 && i <= 9 {
			// 1~9格子只能是J~WILD
			if symbol < J || symbol > WILD {
				return false
			}
		} else if i >= 10 && i <= 12 {
			// 10~12格子只能是ONE~RED_WHEEL
			if symbol < ONE || symbol > RED_WHEEL {
				return false
			}
			// 如果开启了额外投注，不能有GREEN_WHEEL和ONE
			if extraBet && (symbol == GREEN_WHEEL || symbol == ONE) {
				return false
			}
			// 如果没有开启额外投注，不能有RED_WHEEL
			if !extraBet && symbol == RED_WHEEL {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

func SymbolToMultiplier(s Symbol) int32 {
	switch s {
	case ONE:
		return 1
	case TWO:
		return 2
	case THREE:
		return 3
	case FIVE:
		return 5
	case TEN:
		return 10
	case FIFTEEN:
		return 15
	default:
		return 1
	}
}

func GetSpecificSymbol(reels map[int32]Symbol) Symbol {
	return reels[11]
}

// 图案及赔率
// var Symbols = []string{"WILD", "RED", "BLUE", "GREEN", "A", "K", "Q", "J"}
// var SpecialSymbols = []string{"1", "2", "3", "5", "10", "15", "WHEEL"}

var Payouts = map[Symbol]int{
	J:     2,
	Q:     5,
	K:     8,
	A:     10,
	GREEN: 12,
	BLUE:  15,
	RED:   20,
	WILD:  25,
}

// 幸运转盘倍数池
var WheelMultipliers = []int32{1, 3, 5, 8, 10, 15, 20, 30, 50, 100, 200, 1000}

// 额外押注模式中，额外倍数
var WheelExtraMultipliers = []int32{1, 2, 3, 5, 10, 15}

// 赔付线（每条线在每个转轴上的行号）
var Paylines = [][]int{
	{2, 5, 8}, // 中间行
	{1, 4, 7}, // 第一行
	{3, 6, 9}, // 第三行
	{1, 5, 9}, // 斜线1
	{3, 5, 7}, // 斜线2
}

// 随机生成转轴
func SpinReels(extraBet bool) map[int32]Symbol {
	grid := make(map[int32]Symbol)
	id := 1
	// 生成普通轴（3列，每列3行，格子ID 1~9）
	for col := 0; col < 3; col++ {
		// colSymbols := GenNormalReelColumn()
		colSymbols := GetElementsByConfig(int32(col+1), extraBet)
		for row := 0; row < 3; row++ {
			grid[int32(id)] = colSymbols[row]
			id++
		}
	}
	// 生成特殊轴（1列3行，格子ID 10~12）
	// specialCol := GenSpecialReelColumn(extraBet)
	specialCol := GetElementsByConfig(4, extraBet)
	for row := 0; row < 3; row++ {
		grid[int32(id)] = specialCol[row]
		id++
	}
	return grid
}

// 判断每条赔付线的中奖情况
func CheckPaylines(reels map[int32]Symbol, bet int64) (int64, []int32) {
	totalWin := int64(0)
	winLines := []int32{}
	for i, line := range Paylines {
		symbols := []Symbol{
			reels[int32(line[0])],
			reels[int32(line[1])],
			reels[int32(line[2])],
		}
		winSymbol := GetWinSymbol(symbols)
		if winSymbol != 0 {
			lineWin := int64(Payouts[winSymbol]) * bet / 5
			totalWin += lineWin
			winLines = append(winLines, int32(i+1)) // 线号从1开始
		}
	}
	return totalWin, winLines
}

// 判断一条线上是否中奖，返回中奖符号
func GetWinSymbol(symbols []Symbol) Symbol {
	// 从后向前循环，这样WILD会被优先匹配
	for i := len(NormalSymbols) - 1; i >= 0; i-- {
		symbol := NormalSymbols[i]
		match := true
		for _, s := range symbols {
			if s != symbol && s != WILD {
				match = false
				break
			}
		}
		if match {
			return symbol
		}
	}

	return 0
}

func GetNormalWin(bet int64, reels map[int32]Symbol, specificSymbol Symbol) (int64, []int32, int32) {
	baseWin, winLines := CheckPaylines(reels, bet)

	if IsWheel(specificSymbol) { //1倍
		return baseWin, winLines, 1
	} else {
		multiplier := SymbolToMultiplier(specificSymbol)
		return baseWin * int64(multiplier), winLines, multiplier
	}
}

func PlayWheel(bet int64, extraBet bool) (wheelMultiplier int32, extraMultipliers map[int32]int32, wheelWin int64) {
	if extraBet {
		extraMultipliers = make(map[int32]int32)
		for _, multiplier := range WheelMultipliers {
			// extraMultiplier := WheelExtraMultipliers[rand.Intn(len(WheelExtraMultipliers))]
			extraMultiplier := GetExtraMultiplierByConfig()
			extraMultipliers[multiplier] = extraMultiplier
		}
		// wheelMultiplier = WheelMultipliers[rand.Intn(len(WheelMultipliers))]
		wheelMultiplier = GetMultiplierByConfig(extraBet)
		extraMultiplier := extraMultipliers[wheelMultiplier]
		wheelWin = int64(wheelMultiplier) * int64(extraMultiplier) * bet
		return
	} else {
		// wheelMultiplier = WheelMultipliers[rand.Intn(len(WheelMultipliers))]
		wheelMultiplier = GetMultiplierByConfig(extraBet)
		wheelWin = int64(wheelMultiplier) * bet
		return
	}
}

func PlaySlot(
	bet int64,
	extraBet bool,
	inputReels map[int32]Symbol,
	userid string,
) (
	score int64,
	reels map[int32]Symbol,
	specificSymbol Symbol,
	winLines []int32,
	winMultiplier int32,
	normalWin int64,
	wheelMultiplier int32,
	extraMultipliers map[int32]int32,
	wheelWin int64,
	valid bool,
) {
	score = int64(0)
	if extraBet {
		score = bet + int64(math.Round(float64(bet)*0.5))
	} else {
		score = bet
	}

	_ = score

	//判断是否需要控制
	control := false
	if !IsInputReelsEmpty(inputReels) {
		control = true
	}

	//判断数据是否有效
	valid = true
	if control {
		if !IsInputReelsValid(inputReels, extraBet) {
			valid = false
		}
	}

	if control && valid { //需要控制且有效
		reels = inputReels
	} else {
		reels = SpinReels(extraBet)
	}

	specificSymbol = GetSpecificSymbol(reels)

	normalWin, winLines, winMultiplier = GetNormalWin(bet, reels, specificSymbol)

	if IsWheel(specificSymbol) {
		wheelMultiplier, extraMultipliers, wheelWin = PlayWheel(bet, extraBet)
	}

	return

}
