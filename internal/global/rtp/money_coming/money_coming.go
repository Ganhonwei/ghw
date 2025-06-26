package moneycoming

import "strconv"

type Symbol int

// 转轴列符号
const (
	Empty    Symbol = iota //空元素
	TWOX                   //2倍
	FIVEX                  //5倍
	TENX                   //10倍
	RESPIN                 //重转
	SCATTER1               //幸运转盘1
	SCATTER2               //幸运转盘2
	TEN                    //10
	ZEROZERO               //00
	FIVE                   //5
	ONE                    //1
	ZERO                   //0
	SCATTER3               //幸运转盘3
	SCATTER4               //幸运转盘4
)

func IsRespin(symbol Symbol) bool {
	return symbol == RESPIN
}

func IsWheel(symbol Symbol) bool {
	return symbol == SCATTER1 || symbol == SCATTER2 || symbol == SCATTER3 || symbol == SCATTER4
}

func IsInputReelsEmpty(inputReels map[int32]Symbol) bool {
	return len(inputReels) == 0
}

func IsNormalSymbol(symbol Symbol) bool {
	switch symbol {
	case TEN, ZEROZERO, FIVE, ONE, ZERO:
		return true
	default:
		return false
	}
}

func IsSpecialSymbol(symbol Symbol) bool {
	switch symbol {
	case TWOX, FIVEX, TENX, RESPIN, SCATTER1, SCATTER2, SCATTER3, SCATTER4:
		return true
	default:
		return false
	}
}

func IsInputReelsValid(inputReels map[int32]Symbol) bool {
	if len(inputReels) != 12 {
		return false
	}
	for i := int32(1); i <= 12; i++ {
		symbol, ok := inputReels[i]
		if !ok {
			return false
		}
		if symbol == Empty {
			continue
		}
		if i >= 1 && i <= 9 {
			if !IsNormalSymbol(symbol) {
				return false
			}
		} else if i >= 10 && i <= 12 {
			if !IsSpecialSymbol(symbol) {
				return false
			}
		}
	}
	return true
}

func SymbolToMultiplier(s Symbol) int32 {
	switch s {
	case TWOX:
		return 2
	case FIVEX:
		return 5
	case TENX:
		return 10
	default:
		return 1
	}
}

func GetSpecificSymbol(reels map[int32]Symbol) Symbol {
	return reels[11]
}

// 幸运转盘倍数池
var WheelMultipliers = []int32{1, 3, 5, 8, 10, 15, 20, 30, 50, 100, 200, 1000}

// 额外押注模式中，额外倍数
var WheelExtraMultipliers = []int32{1, 2, 3, 5, 10, 15}

// 赔付线（每条线在每个转轴上的行号）
var Paylines = [][]int{
	{2, 5, 8}, // 中间行
}

// 随机生成转轴
func SpinReels(level int32) map[int32]Symbol {
	grid := make(map[int32]Symbol)
	id := 1
	// 生成普通轴（3列，每列3行，格子ID 1~9）
	for col := 0; col < 3; col++ {
		colSymbols := GetElementsByConfig(int32(col+1), level)
		for row := 0; row < 3; row++ {
			grid[int32(id)] = colSymbols[row]
			id++
		}
	}
	// 生成特殊轴（1列3行，格子ID 10~12）
	specialCol := GetElementsByConfig(4, level)
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
		winNumber := GetWinSymbol(symbols)
		if winNumber > 0 {
			totalWin += winNumber
			winLines = append(winLines, int32(i+1)) // 线号从1开始
		}
	}
	return totalWin, winLines
}

func GetWinSymbol(symbols []Symbol) int64 {
	// 定义符号到字符串的映射
	symbolToStr := map[Symbol]string{
		Empty:    "",
		TEN:      "10",
		ZEROZERO: "00",
		FIVE:     "5",
		ONE:      "1",
		ZERO:     "0",
	}

	numStr := ""
	for _, s := range symbols {
		str, ok := symbolToStr[s]
		if !ok {
			// 如果遇到未定义的符号，直接返回0，不中奖
			return 0
		}
		numStr += str
	}

	// 转成整数
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0
	}
	return int64(num)
}

func GetNormalWin(bet int64, reels map[int32]Symbol, specificSymbol Symbol) (int64, []int32, int32) {
	baseWin, winLines := CheckPaylines(reels, bet)

	multiplier := SymbolToMultiplier(specificSymbol)
	return baseWin * int64(multiplier), winLines, multiplier
}

func GetLevel(bet int64) int32 {
	switch bet {
	case 100:
		return 1
	case 500:
		return 2
	case 1000:
		return 3
	case 5000:
		return 4
	case 10000:
		return 5
	default:
		return 1
	}
}

// func PlayWheel(bet int64, extraBet bool) (wheelMultiplier int32, extraMultipliers map[int32]int32, wheelWin int64) {
// 	if extraBet {
// 		extraMultipliers = make(map[int32]int32)
// 		for _, multiplier := range WheelMultipliers {
// 			// extraMultiplier := WheelExtraMultipliers[rand.Intn(len(WheelExtraMultipliers))]
// 			extraMultiplier := GetExtraMultiplierByConfig()
// 			extraMultipliers[multiplier] = extraMultiplier
// 		}
// 		// wheelMultiplier = WheelMultipliers[rand.Intn(len(WheelMultipliers))]
// 		wheelMultiplier = GetMultiplierByConfig(extraBet)
// 		extraMultiplier := extraMultipliers[wheelMultiplier]
// 		wheelWin = int64(wheelMultiplier) * int64(extraMultiplier) * bet
// 		return
// 	} else {
// 		// wheelMultiplier = WheelMultipliers[rand.Intn(len(WheelMultipliers))]
// 		wheelMultiplier = GetMultiplierByConfig(extraBet)
// 		wheelWin = int64(wheelMultiplier) * bet
// 		return
// 	}
// }

func PlaySlot(
	bet int64,
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
	score = bet

	level := GetLevel(bet)

	//判断是否需要控制
	control := false
	if !IsInputReelsEmpty(inputReels) {
		control = true
	}

	//判断数据是否有效
	valid = true
	if control {
		if !IsInputReelsValid(inputReels) {
			valid = false
		}
	}

	if control && valid { //需要控制且有效
		reels = inputReels
	} else {
		reels = SpinReels(level)
	}

	specificSymbol = GetSpecificSymbol(reels)

	normalWin, winLines, winMultiplier = GetNormalWin(bet, reels, specificSymbol)

	// if IsWheel(specificSymbol) {
	// 	wheelMultiplier, extraMultipliers, wheelWin = PlayWheel(bet, extraBet)
	// }

	return

}
