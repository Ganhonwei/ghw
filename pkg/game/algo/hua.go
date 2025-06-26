package algo

import (
	"fmt"
	"goserver/pkg/utils"
	"math/rand"
	"sort"
)

/*
牌型
豹子：三张同样大小的牌。
同花顺：花色相同的三张连牌。
顺子：三张花色不全相同的连牌。
同花：三张花色相同的牌。
对子：三张牌中有两张同样大小的牌。
高牌：除以上牌型的牌。

牌型的比较
1. 豹子>同花顺>顺子>同花>对子>高牌。
2．牌点中，2为最小，A为最大。从大到小依次为：A、K、Q、J、10、9、8、7、6、5、4、3、2
*/

// 牌型定义
const (
	Null        uint32 = iota
	GaoPai             //高牌
	DuiZi              //对子
	TongHua            //同花
	ShunZi             //顺子
	TongHuaShun        //同花顺
	BaoZi              //豹子
	BaoZi1             //AAA
	BaoZi2             //KKK-888
	BaoZi3             //777-222
)

const (
	BaoziUp         = 10 //豹子10以上
	BaoziDown       = 11 //豹子10以下
	TonghuashunUp   = 20 //同花顺10以上
	TonghuashunDown = 21 //同花顺10以下
	ShunziUp        = 30 //顺子10以上
	ShunziDown      = 31 //顺子10以下
	TonghuaUp       = 40 //同花10以上
	TonghuaDown     = 41 //同花10以下
	DuiziUp         = 50 //对子10以上
	DuiziDown       = 51 //对子10以下
	GaopaiUp        = 60 //高牌10以上
	GaopaiDown      = 61 //高牌10以下
)

const (
	Null1 uint32 = iota
	Duizi28
	Duizi9A
)

// func Cardstringify(c []uint32) string {
// 	var ret string
// 	t := HuaType(c)
// 	switch t {
// 	case BaoZi:
// 		ret += "豹子,"
// 	case TongHuaShun:
// 		ret += "同花顺,"
// 	case ShunZi:
// 		ret += "顺子,"
// 	case TongHua:
// 		ret += "同花,"
// 	case DuiZi:
// 		ret += "对子,"
// 	case GaoPai:
// 		ret += "高牌,"
// 	}
// 	hands := toHands(c)
// 	for _, v := range hands {
// 		switch v.Suit {
// 		case Spade:
// 			ret += "黑桃"
// 		case Heart:
// 			ret += "红桃"
// 		case Club:
// 			ret += "梅花"
// 		case Diamond:
// 			ret += "方块"
// 		}
// 		switch v.Rank {
// 		case Ace:
// 			ret += "A"
// 		case Deuce:
// 			ret += "2"
// 		case Trey:
// 			ret += "3"
// 		case Four:
// 			ret += "4"
// 		case Five:
// 			ret += "5"
// 		case Six:
// 			ret += "6"
// 		case Seven:
// 			ret += "7"
// 		case Eight:
// 			ret += "8"
// 		case Nine:
// 			ret += "9"
// 		case Ten:
// 			ret += "10"
// 		case Jack:
// 			ret += "J"
// 		case Queen:
// 			ret += "Q"
// 		case King:
// 			ret += "K"
// 		}
// 	}
// 	return ret
// }

// 计算胜率
func CaclWinRate(cards []uint32) float64 {
	allCards := append([]uint32{}, NiuCARDS...) //所有的牌
	leftCards := []uint32{}                     //剩余牌

	for _, v := range allCards {
		if v != cards[0] && v != cards[1] && v != cards[2] {
			leftCards = append(leftCards, v)
		}
	}

	win := 0
	total := 0
	l := len(leftCards)
	for i := 0; i < l; i++ {
		for j := i + 1; j < l; j++ {
			for k := j + 1; k < l; k++ {
				c := []uint32{leftCards[i], leftCards[j], leftCards[k]}
				if HuaCompare(cards, c) {
					win++
				}
				total++
			}
		}
	}
	return float64(win) / float64(total)
}

// 从剩余牌中获取高牌
func GetGaoPaiCard(current_cards []uint32) (cards []uint32, remain_cards []uint32) {
	gao := [][]uint32{}
	l := len(current_cards)
	for i := 0; i < l; i++ {
		for j := i + 1; j < l; j++ {
			for k := j + 1; k < l; k++ {
				c := []uint32{current_cards[i], current_cards[j], current_cards[k]}
				if HuaType(c) == GaoPai {
					gao = append(gao, c)
				}
			}
		}
	}
	cards = gao[utils.RandIntN(len(gao))]

	for _, v := range current_cards {
		if v != cards[0] && v != cards[1] && v != cards[2] {
			remain_cards = append(remain_cards, v)
		}
	}
	return
}

// 获取一副更大的牌
func GetBiggerCard(cards []uint32, current_cards []uint32) (bigger []uint32, remain_cards []uint32) {
	big := [][]uint32{}

	l := len(current_cards)
	for i := 0; i < l; i++ {
		for j := i + 1; j < l; j++ {
			for k := j + 1; k < l; k++ {
				c := []uint32{current_cards[i], current_cards[j], current_cards[k]}
				if HuaCompare(c, cards) {
					typ := HuaType(c)
					if typ == GaoPai || typ == DuiZi { //只能是对子和单牌
						big = append(big, c)
					}
				}
			}
		}
	}

	if len(big) == 0 {
		remain_cards = current_cards
		return
	}

	bigger = big[utils.RandIntN(len(big))]

	for _, v := range current_cards {
		if v != bigger[0] && v != bigger[1] && v != bigger[2] {
			remain_cards = append(remain_cards, v)
		}
	}

	return
}

// 获取一副更小的牌
func GetSmallerCard(cards []uint32, current_cards []uint32) (smaller []uint32, remain_cards []uint32) {
	small := [][]uint32{}

	l := len(current_cards)
	for i := 0; i < l; i++ {
		for j := i + 1; j < l; j++ {
			for k := j + 1; k < l; k++ {
				c := []uint32{current_cards[i], current_cards[j], current_cards[k]}
				if !HuaCompare(c, cards) {
					small = append(small, c)
				}
			}
		}
	}

	if len(small) == 0 {
		remain_cards = current_cards
		return
	}

	smaller = small[utils.RandIntN(len(small))]

	for _, v := range current_cards {
		if v != smaller[0] && v != smaller[1] && v != smaller[2] {
			remain_cards = append(remain_cards, v)
		}
	}

	return
}

// 从剩余牌中随机一副比最大牌还大的牌
func GetMaxCard(max []uint32, current_cards []uint32) (bigger []uint32, remain_cards []uint32) {
	big := [][]uint32{}

	l := len(current_cards)
	for i := 0; i < l; i++ {
		for j := i + 1; j < l; j++ {
			for k := j + 1; k < l; k++ {
				c := []uint32{current_cards[i], current_cards[j], current_cards[k]}
				if HuaCompare(c, max) {
					big = append(big, c)
				}
			}
		}
	}

	if len(big) == 0 {
		remain_cards = current_cards
		return
	}

	bigger = big[utils.RandIntN(len(big))]

	for _, v := range current_cards {
		if v != bigger[0] && v != bigger[1] && v != bigger[2] {
			remain_cards = append(remain_cards, v)
		}
	}
	return
}

// 从剩余牌中随机一副比最大牌还大一档的牌, 没有大一档的换成同档中更大的牌
func GetCardHuaTypeBigger1(max []uint32, current_cards []uint32) (bigger []uint32, remain_cards []uint32) {
	levelBiggers := map[int][][]uint32{}
	maxHuaType := HuaTypeUpOrDown(max)

	l := len(current_cards)
	for i := 0; i < l; i++ {
		for j := i + 1; j < l; j++ {
			for k := j + 1; k < l; k++ {
				c := []uint32{current_cards[i], current_cards[j], current_cards[k]}
				if HuaCompare(c, max) {
					// 大几档
					bigger := maxHuaType - HuaTypeUpOrDown(c)
					level := bigger/9 + bigger%9
					levelBiggers[level] = append(levelBiggers[level], c)
				}
			}
		}
	}

	if len(levelBiggers) == 0 {
		remain_cards = current_cards
		return
	}

	for i := 1; i <= 11; i++ {
		if biggers, ok := levelBiggers[i]; ok {
			bigger = biggers[utils.RandIntN(len(biggers))]
			break
		}
	}
	if bigger == nil {
		if biggers, ok := levelBiggers[0]; ok {
			bigger = biggers[utils.RandIntN(len(biggers))]
		}
	}
	if bigger == nil {
		remain_cards = current_cards
		return
	}

	for _, v := range current_cards {
		if v != bigger[0] && v != bigger[1] && v != bigger[2] {
			remain_cards = append(remain_cards, v)
		}
	}
	return
}

// 从剩余牌中取n副大于等于指定牌型的牌
func GetCardsHuaTypeBugger(huaType int, count int, current_cards []uint32) (cardss [][]uint32, remain_cards []uint32) {
	remain_cards = make([]uint32, len(current_cards))
	copy(remain_cards, current_cards)

cardsLoop:
	for i := 0; i < count; i++ {
		l := len(remain_cards)
		for i := 0; i < l; i++ {
			for j := i + 1; j < l; j++ {
				for k := j + 1; k < l; k++ {
					c := []uint32{remain_cards[i], remain_cards[j], remain_cards[k]}
					t := HuaTypeUpOrDown(c)
					if t <= huaType {
						cardss = append(cardss, c)
						remain_cards, _ = RemoveCards(remain_cards, c)                            // 从剩余牌组移除
						sort.Slice(remain_cards, func(i, j int) bool { return utils.RandBool() }) // 洗牌
						continue cardsLoop
					}
				}
			}
		}
	}
	return
}

// 从剩余牌中随机一副比当前牌还小N档的牌, 没有小N档的换成大一档中更小的牌
func GetCardsHuaTypeLessN(cards []uint32, current_cards []uint32, level int) (toCards []uint32, remain_cards []uint32) {
	remain_cards = append([]uint32{}, current_cards...)
	remain_cards = append(remain_cards, cards...)

	huaType := HuaTypeUpOrDown(cards)
	l := len(remain_cards)
	lessHuaTypeCards := map[int][][]uint32{}
	for i := 0; i < l; i++ {
		for j := i + 1; j < l; j++ {
			for k := j + 1; k < l; k++ {
				c := []uint32{remain_cards[i], remain_cards[j], remain_cards[k]}
				if HuaCompare(cards, c) {
					// 小几档
					less := HuaTypeUpOrDown(c) - huaType
					lessLevel := less/9 + less%9
					if lessLevel == level {
						toCards = c
						remain_cards, _ = RemoveCards(remain_cards, c) // 从剩余牌组移除
						// 洗牌
						sort.Slice(remain_cards, func(i, j int) bool { return utils.RandBool() })
						return
					} else { //  if lessLevel < level
						lessHuaTypeCards[lessLevel] = append(lessHuaTypeCards[lessLevel], c)
					}
				}
			}
		}
	}

	for i := level - 1; i >= 0; i-- {
		if lesss, ok := lessHuaTypeCards[i]; ok && len(lesss) > 0 {
			c := lesss[utils.RandIntN(len(lesss))]
			toCards = c
			remain_cards, _ = RemoveCards(remain_cards, c) // 从剩余牌组移除
			sort.Slice(remain_cards, func(i, j int) bool { return utils.RandBool() })
			return
		}
	}

	toCards = cards
	remain_cards, _ = RemoveCards(remain_cards, cards) // 从剩余牌组移除
	sort.Slice(remain_cards, func(i, j int) bool { return utils.RandBool() })
	return
}

// 从剩余牌中随机一副比当前牌还大N档的牌, 没有大N档的换成低一档中更大的牌
func GetCardssHuaTypeBiggerN(cardss [][]uint32, current_cards []uint32, level int) (toCardss [][]uint32, remain_cards []uint32) {
	var deskCards = append([]uint32{}, current_cards...)
	for _, cards := range cardss {
		deskCards = append(deskCards, cards...)
	}

	// 洗牌
	d := deskCards
	for i := len(d) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		d[i], d[j] = d[j], d[i]
	}

	// huaTypeCards := map[int][][]uint32{}
	// huaTypes := []int{BaoziUp, BaoziDown, TonghuashunUp, TonghuashunDown, ShunziUp, ShunziDown, TonghuaUp, TonghuaDown, DuiziUp, DuiziDown, GaopaiUp, GaopaiDown}
	var usedCards = make(map[uint32]bool)

cardsLoop:
	for _, cards := range cardss {
		huaType := HuaTypeUpOrDown(cards)

		l := len(deskCards)
		biggerHuaTypeCards := map[int][][]uint32{}
		for i := 0; i < l; i++ {
			for j := i + 1; j < l; j++ {
				for k := j + 1; k < l; k++ {
					c := []uint32{deskCards[i], deskCards[j], deskCards[k]}
					if usedCards[c[0]] || usedCards[c[1]] || usedCards[c[2]] {
						continue
					}
					if HuaCompare(c, cards) {
						// 大几档
						bigger := huaType - HuaTypeUpOrDown(c)
						biggerLevel := bigger/9 + bigger%9
						if biggerLevel == level {
							toCardss = append(toCardss, c)
							usedCards[c[0]] = true
							usedCards[c[1]] = true
							usedCards[c[2]] = true
							continue cardsLoop
						} else if biggerLevel < level {
							biggerHuaTypeCards[biggerLevel] = append(biggerHuaTypeCards[biggerLevel], c)
						}
					}
				}
			}
		}
		for i := level - 1; i >= 0; i-- {
			if biggers, ok := biggerHuaTypeCards[i]; ok && len(biggers) > 0 {
				// c := biggers[utils.RandIntN(len(biggers))]
				for _, c := range biggers {
					if usedCards[c[0]] || usedCards[c[1]] || usedCards[c[2]] {
						continue
					}
					fmt.Printf("use less level cards: %v\n", CardsString(c))
					toCardss = append(toCardss, c)
					usedCards[c[0]] = true
					usedCards[c[1]] = true
					usedCards[c[2]] = true
					continue cardsLoop
				}
			}
		}

		c := cards
		toCardss = append(toCardss, c)
		usedCards[c[0]] = true
		usedCards[c[1]] = true
		usedCards[c[2]] = true
	}

	var useCards []uint32
	for _, cards := range toCardss {
		useCards = append(useCards, cards...)
	}
	remain_cards, _ = RemoveCards(deskCards, useCards)
	return
}

func RandomIndex(length int) (i, j, k int) {
	if length < 3 { //长度必须大于3
		return 0, 0, 0
	}
	var arr []uint32
	for i := 0; i < length; i++ {
		arr = append(arr, uint32(i))
	}
	Shuffle(arr)

	i = int(arr[0])
	j = int(arr[1])
	k = int(arr[2])
	return
}

func HuaTypeUpOrDown(cs []uint32) int {
	h := toHands(cs)
	i := toType(h)

	switch i {
	case BaoZi:
		return BaoziUpOrDown(h)
	case TongHuaShun:
		return TonghuashunUpOrDown(h)
	case ShunZi:
		return ShunziUpOrDown(h)
	case TongHua:
		return TonghuaUpOrDown(h)
	case DuiZi:
		return DuiziUpOrDown(h)
	case GaoPai:
		return GaoPaiUpOrDown(h)
	default:
		return 0
	}
}

func BaoziUpOrDown(h []hands) int {
	if h[0].Rank == Ace {
		return BaoziUp
	}
	if h[0].Rank >= Ten {
		return BaoziUp
	} else {
		return BaoziDown
	}
}

func TonghuashunUpOrDown(h []hands) int {
	if isAQK(h) {
		return TonghuashunUp
	}
	if is123(h) {
		return TonghuashunUp
	}
	if h[0].Rank >= Ten {
		return TonghuashunUp
	} else {
		return TonghuashunDown
	}
}

func ShunziUpOrDown(h []hands) int {
	if isAQK(h) {
		return ShunziUp
	}
	if is123(h) {
		return ShunziUp
	}
	if h[0].Rank >= Ten {
		return ShunziUp
	} else {
		return ShunziDown
	}
}
func TonghuaUpOrDown(h []hands) int {
	if h[0].Rank == Ace {
		return TonghuaUp
	}
	if h[0].Rank >= Ten {
		return TonghuaUp
	} else {
		return TonghuaDown
	}
}

func DuiziUpOrDown(h []hands) int {
	a1, _, _ := huaDuiZiPoint(h)
	if a1 == Ace {
		return DuiziUp
	}
	if a1 >= Ten {
		return DuiziUp
	} else {
		return DuiziDown
	}
}
func Duizi28Or9A(c []uint32) uint32 {
	if HuaType(c) != DuiZi {
		return Null1
	}

	h := toHands(c)
	a1, _, _ := huaDuiZiPoint(h)
	if a1 == Ace {
		return Duizi9A
	}
	if a1 >= Nine {
		return Duizi9A
	} else {
		return Duizi28
	}

}
func GaoPaiUpOrDown(h []hands) int {
	if h[0].Rank == Ace {
		return GaopaiUp
	}
	if h[0].Rank >= Ten {
		return GaopaiUp
	} else {
		return GaopaiDown
	}
}

// 是否K以上的高牌
func GaoPaiAK(cs []uint32) bool {
	h := toHands(cs)
	i := toType(h)
	return i == GaoPai && (h[0].Rank == Ace || h[0].Rank == King)
}

func GetCard4(cardtype uint32, current_cards []uint32) (cards []uint32, remain_cards []uint32) {
	l := len(current_cards)

	BaoZiUp_List := [][]uint32{}
	BaoZiDown_List := [][]uint32{}
	TongHuaShunUp_List := [][]uint32{}
	TongHuaShunDown_List := [][]uint32{}
	ShunZiUp_List := [][]uint32{}
	ShunZiDown_List := [][]uint32{}
	TongHuaUp_List := [][]uint32{}
	TongHuaDown_List := [][]uint32{}
	DuiZiUp_List := [][]uint32{}
	DuiZiDown_List := [][]uint32{}
	GaoPaiUp_List := [][]uint32{}
	GaoPaiDown_List := [][]uint32{}

	for i := 0; i < l; i++ {
		for j := i + 1; j < l; j++ {
			for k := j + 1; k < l; k++ {
				c := []uint32{current_cards[i], current_cards[j], current_cards[k]}
				switch HuaTypeUpOrDown(c) {
				case BaoziUp:
					BaoZiUp_List = append(BaoZiUp_List, c)
				case BaoziDown:
					BaoZiDown_List = append(BaoZiDown_List, c)
				case TonghuashunUp:
					TongHuaShunUp_List = append(TongHuaShunUp_List, c)
				case TonghuashunDown:
					TongHuaShunDown_List = append(TongHuaShunDown_List, c)
				case ShunziUp:
					ShunZiUp_List = append(ShunZiUp_List, c)
				case ShunziDown:
					ShunZiDown_List = append(ShunZiDown_List, c)
				case TonghuaUp:
					TongHuaUp_List = append(TongHuaUp_List, c)
				case TonghuaDown:
					TongHuaDown_List = append(TongHuaDown_List, c)
				case DuiziUp:
					DuiZiUp_List = append(DuiZiUp_List, c)
				case DuiziDown:
					DuiZiDown_List = append(DuiZiDown_List, c)
				case GaopaiUp:
					GaoPaiUp_List = append(GaoPaiUp_List, c)
				case GaopaiDown:
					GaoPaiDown_List = append(GaoPaiDown_List, c)
				}
			}
		}
	}

	switch cardtype {
	case BaoziUp:
		if len(BaoZiUp_List) != 0 {
			cards = BaoZiUp_List[utils.RandIntN(len(BaoZiUp_List))]
		} else {
			i, j, k := RandomIndex(len(current_cards))
			cards = []uint32{current_cards[i], current_cards[j], current_cards[k]}
		}
	case BaoziDown:
		if len(BaoZiDown_List) != 0 {
			cards = BaoZiDown_List[utils.RandIntN(len(BaoZiDown_List))]
		} else {
			i, j, k := RandomIndex(len(current_cards))
			cards = []uint32{current_cards[i], current_cards[j], current_cards[k]}
		}
	case TonghuashunUp:
		if len(TongHuaShunUp_List) != 0 {
			cards = TongHuaShunUp_List[utils.RandIntN(len(TongHuaShunUp_List))]
		} else {
			i, j, k := RandomIndex(len(current_cards))
			cards = []uint32{current_cards[i], current_cards[j], current_cards[k]}
		}
	case TonghuashunDown:
		if len(TongHuaShunDown_List) != 0 {
			cards = TongHuaShunDown_List[utils.RandIntN(len(TongHuaShunDown_List))]
		} else {
			i, j, k := RandomIndex(len(current_cards))
			cards = []uint32{current_cards[i], current_cards[j], current_cards[k]}
		}
	case ShunziUp:
		if len(ShunZiUp_List) != 0 {
			cards = ShunZiUp_List[utils.RandIntN(len(ShunZiUp_List))]
		} else {
			i, j, k := RandomIndex(len(current_cards))
			cards = []uint32{current_cards[i], current_cards[j], current_cards[k]}
		}
	case ShunziDown:
		if len(ShunZiDown_List) != 0 {
			cards = ShunZiDown_List[utils.RandIntN(len(ShunZiDown_List))]
		} else {
			i, j, k := RandomIndex(len(current_cards))
			cards = []uint32{current_cards[i], current_cards[j], current_cards[k]}
		}
	case TonghuaUp:
		if len(TongHuaUp_List) != 0 {
			cards = TongHuaUp_List[utils.RandIntN(len(TongHuaUp_List))]
		} else {
			i, j, k := RandomIndex(len(current_cards))
			cards = []uint32{current_cards[i], current_cards[j], current_cards[k]}
		}
	case TonghuaDown:
		if len(TongHuaDown_List) != 0 {
			cards = TongHuaDown_List[utils.RandIntN(len(TongHuaDown_List))]
		} else {
			i, j, k := RandomIndex(len(current_cards))
			cards = []uint32{current_cards[i], current_cards[j], current_cards[k]}
		}
	case DuiziUp:
		if len(DuiZiUp_List) != 0 {
			cards = DuiZiUp_List[utils.RandIntN(len(DuiZiUp_List))]
		} else {
			i, j, k := RandomIndex(len(current_cards))
			cards = []uint32{current_cards[i], current_cards[j], current_cards[k]}
		}
	case DuiziDown:
		if len(DuiZiDown_List) != 0 {
			cards = DuiZiDown_List[utils.RandIntN(len(DuiZiDown_List))]
		} else {
			i, j, k := RandomIndex(len(current_cards))
			cards = []uint32{current_cards[i], current_cards[j], current_cards[k]}
		}
	case GaopaiUp:
		if len(GaoPaiUp_List) != 0 {
			cards = GaoPaiUp_List[utils.RandIntN(len(GaoPaiUp_List))]
		} else {
			i, j, k := RandomIndex(len(current_cards))
			cards = []uint32{current_cards[i], current_cards[j], current_cards[k]}
		}
	case GaopaiDown:
		if len(GaoPaiDown_List) != 0 {
			cards = GaoPaiDown_List[utils.RandIntN(len(GaoPaiDown_List))]
		} else {
			i, j, k := RandomIndex(len(current_cards))
			cards = []uint32{current_cards[i], current_cards[j], current_cards[k]}
		}
	}

	for _, v := range current_cards {
		if v != cards[0] && v != cards[1] && v != cards[2] {
			remain_cards = append(remain_cards, v)
		}
	}

	return
}

func GetCard(cardtype uint32, current_cards []uint32) (cards []uint32, remain_cards []uint32) {
	l := len(current_cards)

	BaoZi_List := [][]uint32{}
	BaoZi1_List := [][]uint32{}
	BaoZi2_List := [][]uint32{}
	BaoZi3_List := [][]uint32{}
	TongHuaShun_List := [][]uint32{}
	ShunZi_List := [][]uint32{}
	TongHua_List := [][]uint32{}
	DuiZi_List := [][]uint32{}
	GaoPai_List := [][]uint32{}

	for i := 0; i < l; i++ {
		for j := i + 1; j < l; j++ {
			for k := j + 1; k < l; k++ {
				c := []uint32{current_cards[i], current_cards[j], current_cards[k]}
				switch HuaType(c) {
				case BaoZi:
					rank := Rank(current_cards[i])
					BaoZi_List = append(BaoZi_List, c)
					if rank == Ace {
						BaoZi1_List = append(BaoZi1_List, c)
					} else if rank >= Eight && rank <= King {
						BaoZi2_List = append(BaoZi2_List, c)
					} else if rank >= Deuce && rank <= Seven {
						BaoZi3_List = append(BaoZi3_List, c)
					}
				case TongHuaShun:
					TongHuaShun_List = append(TongHuaShun_List, c)
				case ShunZi:
					ShunZi_List = append(ShunZi_List, c)
				case TongHua:
					TongHua_List = append(TongHua_List, c)
				case DuiZi:
					DuiZi_List = append(DuiZi_List, c)
				case GaoPai:
					GaoPai_List = append(GaoPai_List, c)
				}
			}
		}
	}

	switch cardtype {
	case BaoZi:
		if len(BaoZi_List) != 0 {
			cards = BaoZi_List[utils.RandIntN(len(BaoZi_List))]
		} else {
			i, j, k := RandomIndex(len(current_cards))
			cards = []uint32{current_cards[i], current_cards[j], current_cards[k]}
		}
	case BaoZi1:
		if len(BaoZi1_List) != 0 {
			cards = BaoZi1_List[utils.RandIntN(len(BaoZi1_List))]
		} else {
			i, j, k := RandomIndex(len(current_cards))
			cards = []uint32{current_cards[i], current_cards[j], current_cards[k]}
		}
	case BaoZi2:
		if len(BaoZi2_List) != 0 {
			cards = BaoZi2_List[utils.RandIntN(len(BaoZi2_List))]
		} else {
			i, j, k := RandomIndex(len(current_cards))
			cards = []uint32{current_cards[i], current_cards[j], current_cards[k]}
		}
	case BaoZi3:
		if len(BaoZi3_List) != 0 {
			cards = BaoZi3_List[utils.RandIntN(len(BaoZi3_List))]
		} else {
			i, j, k := RandomIndex(len(current_cards))
			cards = []uint32{current_cards[i], current_cards[j], current_cards[k]}
		}
	case TongHuaShun:
		if len(TongHuaShun_List) != 0 {
			cards = TongHuaShun_List[utils.RandIntN(len(TongHuaShun_List))]
		} else {
			i, j, k := RandomIndex(len(current_cards))
			cards = []uint32{current_cards[i], current_cards[j], current_cards[k]}
		}
	case ShunZi:
		if len(ShunZi_List) != 0 {
			cards = ShunZi_List[utils.RandIntN(len(ShunZi_List))]
		} else {
			i, j, k := RandomIndex(len(current_cards))
			cards = []uint32{current_cards[i], current_cards[j], current_cards[k]}
		}
	case TongHua:
		if len(TongHua_List) != 0 {
			cards = TongHua_List[utils.RandIntN(len(TongHua_List))]
		} else {
			i, j, k := RandomIndex(len(current_cards))
			cards = []uint32{current_cards[i], current_cards[j], current_cards[k]}
		}
	case DuiZi:
		if len(DuiZi_List) != 0 {
			cards = DuiZi_List[utils.RandIntN(len(DuiZi_List))]
		} else {
			i, j, k := RandomIndex(len(current_cards))
			cards = []uint32{current_cards[i], current_cards[j], current_cards[k]}
		}
	case GaoPai:
		if len(GaoPai_List) != 0 {
			cards = GaoPai_List[utils.RandIntN(len(GaoPai_List))]
		} else {
			i, j, k := RandomIndex(len(current_cards))
			cards = []uint32{current_cards[i], current_cards[j], current_cards[k]}
		}
	}

	for _, v := range current_cards {
		if v != cards[0] && v != cards[1] && v != cards[2] {
			remain_cards = append(remain_cards, v)
		}
	}

	return
}

// 获取单牌最大牌值
func GetGaoPaiRank(cs []uint32) uint32 {
	hs := toHands(cs)
	if hs == nil {
		return 0
	}
	return hs[0].Rank
}

// 牌数据转换为手牌
func toHands(cs []uint32) (hs []hands) {
	if len(cs) != 3 {
		return nil
	}

	hs = make([]hands, len(cs))
	for i, v := range cs {
		hs[i].Suit = Suit(v)
		hs[i].Rank = Rank(v)
	}
	descSort2Hands(hs)
	return
}

// 排序规则(降序排序)
func descSort2Hands(cards []hands) {
	sort.Slice(cards, func(i, j int) bool {
		//先比牌值
		if cards[i].Rank != cards[j].Rank {
			//A最大
			if cards[i].Rank == Ace {
				return true
			}
			if cards[j].Rank == Ace {
				return false
			}
			//正常牌值比较
			return cards[i].Rank > cards[j].Rank
		} else { //牌值相同比花色
			return cards[i].Suit > cards[j].Suit
		}
	})
}

// 通过牌数据获取牌型
func HuaType(cs []uint32) (i uint32) {
	hs := toHands(cs)
	i = toType(hs)
	return
}

// 是否是豹子
func isBaoZi(hs []hands) bool {
	if hs[0].Rank == hs[1].Rank && hs[1].Rank == hs[2].Rank {
		return true
	}
	return false
}

// 是否是同花顺
func isTongHuaShun(hs []hands) bool {
	if isTongHua(hs) && isShunZi(hs) {
		return true
	}
	return false
}

// 是否是AQK
func isAQK(hs []hands) bool {
	if hs[0].Rank == Ace && hs[1].Rank == King && hs[2].Rank == Queen {
		return true
	}
	return false
}

// 是否是A23
func is123(hs []hands) bool {
	if hs[0].Rank == Ace && hs[1].Rank == Trey && hs[2].Rank == Deuce {
		return true
	}
	return false
}

// 是否是顺子
func isShunZi(hs []hands) bool {
	//AKQ
	if isAQK(hs) {
		return true
	}
	//A32（A23）
	if is123(hs) {
		return true
	}
	//常规顺子
	if hs[0].Rank == hs[1].Rank+1 && hs[1].Rank == hs[2].Rank+1 {
		return true
	}
	return false
}

// 是否是同花
func isTongHua(hs []hands) bool {
	if hs[0].Suit == hs[1].Suit && hs[1].Suit == hs[2].Suit {
		return true
	}
	return false
}

// 是否是对子
func isDuiZi(hs []hands) bool {
	if hs[0].Rank == hs[1].Rank || hs[1].Rank == hs[2].Rank {
		return true
	}
	return false
}

// 通过手牌获取牌型
func toType(hs []hands) uint32 {
	if len(hs) != 3 {
		return Null
	}
	//豹子
	if isBaoZi(hs) {
		return BaoZi
	}
	//同花顺
	if isTongHuaShun(hs) {
		return TongHuaShun
	}
	//顺子
	if isShunZi(hs) {
		return ShunZi
	}
	//同花
	if isTongHua(hs) {
		return TongHua
	}
	//对子
	if isDuiZi(hs) {
		return DuiZi
	}
	return GaoPai

}

// 获取牌点数
func huaPoint(hs []hands) uint32 {
	return hs[0].Rank + hs[1].Rank + hs[2].Rank
}

// 获取对子点数、单牌点数、最大花色
func huaDuiZiPoint(hs []hands) (uint32, uint32, uint32) {
	if hs[0].Rank == hs[1].Rank {
		return hs[1].Rank, hs[2].Rank, hs[0].Suit
	}
	if hs[1].Rank == hs[2].Rank {
		return hs[1].Rank, hs[0].Rank, hs[1].Suit
	}
	return 0, 0, 0
}

// 比较豹子
func cmpBaoZi(a, b []hands) bool {
	if a[0].Rank == Ace {
		return true
	}
	if b[0].Rank == Ace {
		return false
	}
	return a[0].Rank > b[0].Rank
}

// 比较同花顺
func cmpTongHuaShun(a, b []hands) bool {
	if isAQK(a) && !isAQK(b) {
		return true
	}
	if !isAQK(a) && isAQK(b) {
		return false
	}
	if is123(a) && !is123(b) {
		return true
	}
	if !is123(a) && is123(b) {
		return false
	}

	p1 := huaPoint(a)
	p2 := huaPoint(b)
	if p1 == p2 {
		return a[0].Suit > b[0].Suit
	} else {
		return p1 > p2
	}
}

// 比较顺子
func cmpShunZi(a, b []hands) bool {
	return cmpTongHuaShun(a, b)
}

// 比较同花
func cmpTongHua(a, b []hands) bool {
	if a[0].Rank == Ace && b[0].Rank != Ace {
		return true
	}
	if a[0].Rank != Ace && b[0].Rank == Ace {
		return false
	}
	for i := 0; i < 3; i++ {
		if a[i].Rank > b[i].Rank {
			return true
		}
		if a[i].Rank < b[i].Rank {
			return false
		}
	}
	return a[0].Suit > b[0].Suit
}

// 比较对子
func cmpDuiZi(a, b []hands) bool {
	a1, a2, a3 := huaDuiZiPoint(a)
	b1, b2, b3 := huaDuiZiPoint(b)

	//比较对子
	if a1 == Ace && b1 != Ace {
		return true
	}
	if a1 != Ace && b1 == Ace {
		return false
	}
	if a1 > b1 {
		return true
	}
	if a1 < b1 {
		return false
	}

	//比较单牌
	if a2 == Ace && b2 != Ace {
		return true
	}
	if a2 != Ace && b2 == Ace {
		return false
	}
	if a2 > b2 {
		return true
	}
	if a2 < b2 {
		return false
	}

	return a3 > b3
}

// 比较高牌
func cmpGaoPai(a, b []hands) bool {
	return cmpTongHua(a, b)
}

// HuaCompare 比较 a > b
func HuaCompare(a, b []uint32) bool {
	//转换手牌
	a1 := toHands(a)
	b1 := toHands(b)
	if a1 == nil || b1 == nil {
		return false
	}

	//获取牌型
	a2 := toType(a1)
	b2 := toType(b1)
	if a2 == Null || b2 == Null {
		return false
	}

	//比较牌型
	if a2 != b2 {
		return a2 > b2
	}
	switch a2 {
	case BaoZi:
		return cmpBaoZi(a1, b1)
	case TongHuaShun:
		return cmpTongHuaShun(a1, b1)
	case ShunZi:
		return cmpShunZi(a1, b1)
	case TongHua:
		return cmpTongHua(a1, b1)
	case DuiZi:
		return cmpDuiZi(a1, b1)
	case GaoPai:
		return cmpGaoPai(a1, b1)
	default:
		return false
	}
	//对子
	// if a2 == DuiZi && b2 == DuiZi {
	// 	a3, av3 := pairVal(a1)
	// 	b3, bv3 := pairVal(b1)
	// 	if a3 != b3 {
	// 		//A最大
	// 		if a3 == Ace {
	// 			return true
	// 		}
	// 		if b3 == Ace {
	// 			return false
	// 		}
	// 		return a3 > b3
	// 	}
	// 	//A最大
	// 	if av3 == Ace {
	// 		return true
	// 	}
	// 	if bv3 == Ace {
	// 		return false
	// 	}
	// 	return av3 > bv3
	// }

	//牌值比较
	// for i, v := range a1 {
	// 	if v.Rank == b1[i].Rank {
	// 		continue
	// 	}
	// 	//A最大
	// 	if v.Rank == Ace {
	// 		return true
	// 	}
	// 	if b1[i].Rank == Ace {
	// 		return false
	// 	}
	// 	if v.Rank < b1[i].Rank {
	// 		return false
	// 	} else if v.Rank > b1[i].Rank {
	// 		return true
	// 	}
	// }
	//先开牌者输
	// return true
}

// HuaMultiple 积分倍数
// func HuaMultiple(n uint32) uint32 {
// 	switch n {
// 	case DuiZi:
// 	case ShunZi:
// 		return 2
// 	case TongHua:
// 		return 3
// 	case TongHuaShun:
// 		return 5
// 	case BaoZi:
// 		return 10
// 	}
// 	return 1
// }

// func pairVal(hs []hands) (uint32, uint32) {
// 	if len(hs) != 3 {
// 		return 0, 0
// 	}
// 	if hs[0].Rank == hs[1].Rank {
// 		return hs[0].Rank, hs[2].Rank
// 	}
// 	if hs[1].Rank == hs[2].Rank {
// 		return hs[1].Rank, hs[0].Rank
// 	}
// 	if hs[0].Rank == hs[2].Rank {
// 		return hs[0].Rank, hs[1].Rank
// 	}
// 	return 0, 0
// }
