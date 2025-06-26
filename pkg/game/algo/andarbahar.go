package algo

import "goserver/pkg/utils"

// 发牌次数
func ABDispaterTimes(winner, sideWinner uint32) (times int) {
	min, max := 0, 0
	switch sideWinner {
	case 3:
		min, max = 1, 5
	case 4:
		min, max = 6, 10
	case 5:
		min, max = 11, 15
	case 6:
		min, max = 16, 25
	case 7:
		min, max = 26, 30
	case 8:
		min, max = 31, 35
	case 9:
		min, max = 36, 40
	default:
		min, max = 41, 49
	}

	times = utils.RandIntN(max-min+1) + min
	if winner == 1 && times%2 == 0 {
		// 开andar,times为奇数
		if times+1 > max {
			times--
		} else {
			times++
		}
	} else if winner == 2 && times%2 != 0 {
		// 开badar,times为偶数
		if times+1 > max {
			times--
		} else {
			times++
		}
	}
	return
}

func ABTimesConvertWinner(times int) (uint32, uint32) {
	var winner, sideWinner uint32
	if times%2 == 0 {
		winner = 2
	} else {
		winner = 1
	}

	if times >= 1 && times <= 5 {
		sideWinner = 3
	} else if times >= 6 && times <= 10 {
		sideWinner = 4
	} else if times >= 11 && times <= 15 {
		sideWinner = 5
	} else if times >= 16 && times <= 25 {
		sideWinner = 6
	} else if times >= 26 && times <= 30 {
		sideWinner = 7
	} else if times >= 31 && times <= 35 {
		sideWinner = 8
	} else if times >= 36 && times <= 40 {
		sideWinner = 9
	} else {
		sideWinner = 10
	}

	return winner, sideWinner
}

// 比较牌值是否相同
func ABKeyPoker(joker, value uint32) bool {
	return Rank(joker) == Rank(value)
}

// 生成Key牌
func ABGenKey(joker uint32) uint32 {
	suit := Suit(joker) // joker花色
	rank := Rank(joker) // joker牌值
	for i := 1; i <= int(utils.RandInt32N(3)+1); i++ {
		suit -= 1 << 4
		if suit == 0 {
			suit = Spade
		}
	}
	return rank | suit
}
