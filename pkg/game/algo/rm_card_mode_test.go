package algo

import (
	"fmt"
	"goserver/pkg/utils"
	"strconv"
	"testing"
)

func TestFindAllLackSequence(t *testing.T) {
	// validCardGroupList := FindAllHandCardDivide(handCards, wildCard, 4)
	testCount := 1
	for i := 0; i < testCount; i++ {
		// normalCards := []uint32{
		// 	RMCARDS[0], RMCARDS[8], RMCARDS[12], RMCARDS[10],
		// 	0x2c, 0x2d, 0x41, 0x42,
		// 	0x33, 0x34, 0x35, 0x35, 0x39, 0x3b, //0x3a,
		// }
		normalCards := []uint32{RMCARDS[49], RMCARDS[51]}

		fmt.Println("开始分析缺一门顺子组合：", normalCards)
		allLackSequence := FindAllLackSequence(normalCards)
		for _, lackSequence := range allLackSequence {
			fmt.Println("缺一门顺子组合：", lackSequence)
		}
		fmt.Println("结束分析缺一门顺子组合：", normalCards)
	}
}

func TestFindAllHandCardDivide(t *testing.T) {
	handCards := []uint32{
		RMCARDS[24], RMCARDS[25], RMCARDS[26],
		RMCARDS[19], RMCARDS[23], RMCARDS[31],
		RMCARDS[22], RMCARDS[26], RMCARDS[34],
		RMCARDS[0], RMCARDS[2], RMCARDS[28],
		RMCARDS[28], RMCARDS[2], RMCARDS[27],

		// poker.NewCard(3, 14), poker.NewCard(3, 2), poker.NewCard(3, 3),
		// poker.NewCard(2, 8), poker.NewCard(3, 8), poker.NewCard(4, 8),
		// poker.NewCard(2, 11), poker.NewCard(3, 11), poker.NewCard(4, 11),
		// poker.NewCard(1, 14), poker.NewCard(1, 3), poker.NewCard(3, 5),
		// poker.NewCard(4, 5), poker.NewCard(2, 2),
	}
	wildCard := RMCARDS[9]

	fmt.Println("开始手牌分组", handCards, wildCard)
	handCardDivides := FindAllHandCardDivide(handCards, wildCard, 4)
	for _, handCardDivide := range handCardDivides {
		fmt.Println("handCardDivide:", "BaseGroups=", handCardDivide.BaseGroups, "LeftCards=", handCardDivide.LeftCards)
	}
	fmt.Println("结束手牌分组")

	fmt.Println("2顺子分组", handCards, wildCard)
	shunDivides := findAllSecondLifeValidCardGroupList(handCardDivides, wildCard)
	for _, divide := range shunDivides {
		fmt.Println("handCardDivide:", "BaseGroups=", divide.BaseGroups, "LeftCards=", divide.LeftCards)
	}
	fmt.Println("2顺子分组", handCards, wildCard)
}

func TestValidRmGroup(t *testing.T) {
	fmt.Println(ValidRmGroup([]uint32{70, 71, 26}, uint32(74)))
	fmt.Println(ValidRmGroup([]uint32{65, 76, 77}, uint32(74)))
	fmt.Println(ValidRmGroup([]uint32{23, 22, 21}, uint32(74)))
	fmt.Println(ValidRmGroup([]uint32{37, 35, 36}, uint32(74)))

	fmt.Println(CardsString([]uint32{70, 71, 26}))
	fmt.Println(CardsString([]uint32{65, 76, 77}))
	fmt.Println(CardsString([]uint32{23, 22, 21}))
	fmt.Println(CardsString([]uint32{37, 35, 36}))
	fmt.Println(CardsString([]uint32{74}))

	cards := []uint32{
		70, 71, 26,
		65, 76, 77,
		23, 22, 21,
		37, 35, 36, 55,
	}
	LetsOutCard(cards, []uint32{}, 74)
}

func TestLetsOutCard(t *testing.T) {
	fmt.Println("1")
	// 可胡牌
	// handCards := []uint32{
	// 	NewCard(2, 11), NewCard(2, 10), NewCard(2, 9),
	// 	NewCard(2, 8), NewCard(2, 7), NewCard(2, 6),
	// 	NewCard(4, 12), NewCard(3, 12), NewCard(2, 12),
	// 	NewCard(1, 12), RMJOKERS[0], NewCard(3, 2),
	// 	NewCard(2, 2), NewCard(1, 4), // NewCard(2, 9),
	// }
	// 有1纯顺
	// handCards := []uint32{
	// 	NewCard(2, 11), NewCard(2, 10), NewCard(2, 9),
	// 	NewCard(3, 9), NewCard(3, 7), NewCard(3, 5),
	// 	NewCard(4, 13), NewCard(4, 11), NewCard(4, 9),
	// 	NewCard(1, 12), RMJOKERS[0], NewCard(3, 2),
	// 	NewCard(2, 2), NewCard(1, 4), // NewCard(2, 9),
	// }
	// 无纯顺
	// handCards := []uint32{
	// 	NewCard(2, 12), NewCard(2, 10), NewCard(2, 9),
	// 	NewCard(3, 9), NewCard(3, 7), NewCard(3, 5),
	// 	NewCard(4, 13), NewCard(4, 11), NewCard(4, 9),
	// 	NewCard(1, 12), RMJOKERS[0], NewCard(3, 2),
	// 	NewCard(2, 2), NewCard(1, 4), // NewCard(2, 9),
	// }
	// wildCard := NewCard(4, 14)

	// handCards := []uint32{17, 28, 29, 19, 20, 21, 65, 66, 71, 68, 51, 52, 82, 24}
	// handCards := []uint32{33, 34, 35, 19, 17, 18, 49, 61, 42, 39, 54, 76, 71, 29}
	// wildCard := uint32(42)

	// 异常牌型4
	// handCards := []uint32{
	// 	NewCard(1, 14), NewCard(1, 3), NewCard(1, 4),
	// 	NewCard(1, 13), NewCard(2, 7), NewCard(2, 8),
	// 	NewCard(3, 7), NewCard(3, 8), NewCard(4, 3),
	// 	NewCard(4, 4), NewCard(4, 5), NewCard(4, 6),
	// 	NewCard(4, 7), NewCard(3, 3),
	// }
	// handCards = []uint32{17, 19, 71, 68, 55, 20, 70, 39, 67, 69, 40, 56, 29, NewCard(3, 3)}
	// wildCard := NewCard(3, 8)

	// 异常牌型8
	// handCards := []uint32{
	// 	NewCard(1, 9), NewCard(1, 10), NewCard(1, 11),
	// 	NewCard(1, 7), NewCard(4, 9), NewCard(4, 10),
	// 	NewCard(4, 11), NewCard(4, 12), NewCard(3, 6),
	// 	NewCard(3, 7), RMJOKERS[1], NewCard(2, 4),
	// 	NewCard(1, 4), NewCard(3, 8),
	// }
	// wildCard := NewCard(4, 11)

	// 异常牌型5
	// handCards := []uint32{
	// 	NewCard(1, 11), NewCard(1, 12), NewCard(4, 7),
	// 	NewCard(2, 10), NewCard(3, 10), RMJOKERS[1],
	// 	NewCard(2, 8), NewCard(2, 4), NewCard(2, 2),
	// 	NewCard(1, 4), NewCard(1, 8), NewCard(1, 8),
	// 	NewCard(1, 9), NewCard(3, 8),
	// }
	// wildCard := NewCard(4, 11)

	// 异常牌型0603-1
	// handCards := []uint32{
	// 	NewCard(3, 2), NewCard(3, 3), NewCard(3, 4),
	// 	NewCard(3, 9), NewCard(3, 10), NewCard(3, 11),
	// 	NewCard(4, 7), NewCard(4, 9), NewCard(3, 3),
	// 	NewCard(2, 8), NewCard(2, 9), NewCard(3, 3),
	// 	NewCard(2, 2), NewCard(4, 11),
	// }
	// wildCard := NewCard(2, 3)

	// 异常牌型0603-2
	// handCards := []uint32{
	// 	NewCard(1, 3), NewCard(1, 4), NewCard(1, 5),
	// 	NewCard(4, 6), NewCard(4, 7), NewCard(2, 8),
	// 	NewCard(4, 9), NewCard(4, 10), NewCard(2, 13),
	// 	NewCard(3, 13), NewCard(4, 13), NewCard(2, 3),
	// 	NewCard(2, 4), NewCard(3, 5),
	// }
	// wildCard := NewCard(4, 8)

	// 异常牌型0603-3
	// handCards := []uint32{
	// 	NewCard(2, 1), NewCard(3, 1), NewCard(4, 1),
	// 	NewCard(2, 1), NewCard(2, 13), NewCard(1, 5),
	// 	NewCard(1, 1), NewCard(1, 2), NewCard(1, 3),
	// 	NewCard(2, 10), NewCard(2, 2), NewCard(3, 9),
	// 	NewCard(4, 6), RMJOKERS[1],
	// }
	// wildCard := NewCard(2, 5)
	// fmt.Println("==============================")

	// 异常牌型0603-4
	// handCards := []uint32{
	// 	NewCard(4, 3), NewCard(4, 4), NewCard(4, 5),
	// 	NewCard(4, 9), NewCard(4, 10), RMJOKERS[0],
	// 	RMJOKERS[1], NewCard(4, 13), NewCard(1, 10),
	// 	NewCard(2, 10), NewCard(3, 10), RMJOKERS[1],
	// 	NewCard(3, 7), NewCard(3, 6),
	// }
	// wildCard := NewCard(4, 11)

	// 异常牌型0603-4 ========================= ok
	// handCards := []uint32{
	// 	NewCard(4, 6), NewCard(4, 7), NewCard(4, 8),
	// 	NewCard(4, 9), NewCard(4, 10), NewCard(2, 7),
	// 	NewCard(2, 8), NewCard(2, 11), NewCard(2, 14),
	// 	NewCard(2, 13), RMJOKERS[0], NewCard(2, 4),
	// 	NewCard(4, 4), NewCard(2, 11),
	// }
	// wildCard := NewCard(1, 11)

	// 正常牌型0603-4
	// handCards := []uint32{
	// 	NewCard(4, 6), NewCard(4, 7), NewCard(4, 8),
	// 	NewCard(4, 9), NewCard(4, 10), NewCard(4, 11),
	// 	NewCard(2, 8), NewCard(2, 9), NewCard(2, 10),
	// 	NewCard(3, 5), RMJOKERS[0], NewCard(3, 7),
	// 	NewCard(3, 8), NewCard(3, 9),
	// }
	// wildCard := NewCard(3, 5)

	// 异常牌型0603-5
	// handCards := []uint32{
	// 	NewCard(2, 7), NewCard(4, 7), NewCard(3, 7),
	// 	NewCard(3, 11), NewCard(3, 13), NewCard(4, 2),
	// 	NewCard(2, 11), NewCard(2, 12), NewCard(2, 2),
	// 	NewCard(2, 14), NewCard(3, 5), NewCard(3, 4),
	// 	NewCard(3, 3), NewCard(3, 3),
	// }
	// wildCard := NewCard(3, 2)

	// 异常牌型0604-1
	// handCards := []uint32{
	// 	NewCard(1, 13), NewCard(2, 8), NewCard(2, 11),
	// 	NewCard(4, 12), NewCard(2, 7), NewCard(2, 9),
	// 	NewCard(2, 10), NewCard(4, 13), NewCard(3, 3),
	// 	RMJOKERS[1], NewCard(1, 4), NewCard(1, 12),
	// 	NewCard(4, 11), NewCard(3, 4),
	// }
	// wildCard := uint32(67)

	// 异常牌型0604-1
	// handCards := []uint32{
	// 	NewCard(1, 1), NewCard(1, 2), NewCard(1, 3),
	// 	NewCard(4, 8), NewCard(3, 11), NewCard(3, 12),
	// 	NewCard(3, 13), NewCard(1, 11), NewCard(2, 11),
	// 	NewCard(4, 11), NewCard(2, 14), NewCard(3, 14),
	// 	NewCard(4, 14), // NewCard(3, 4),
	// }
	// wildCard := NewCard(2, 8)

	// 异常牌型0604-2
	// handCards := []uint32{
	// 	NewCard(2, 13), NewCard(2, 12), NewCard(2, 14),
	// 	NewCard(1, 5), NewCard(2, 5), NewCard(3, 5),
	// 	NewCard(4, 5), NewCard(1, 9), NewCard(1, 10),
	// 	NewCard(3, 8), NewCard(1, 12), NewCard(1, 13),
	// 	NewCard(3, 4), NewCard(1, 14),
	// }
	// wildCard := NewCard(1, 8)

	// 异常牌型0604-3
	// handCards := []uint32{
	// 	NewCard(4, 14), NewCard(4, 3), NewCard(4, 4),
	// 	NewCard(4, 2), NewCard(4, 3), NewCard(4, 4),
	// 	NewCard(4, 12), NewCard(1, 11), NewCard(2, 11),
	// 	NewCard(4, 11), RMJOKERS[0], RMJOKERS[0],
	// 	NewCard(2, 6), // NewCard(1, 14),
	// }
	// wildCard := NewCard(4, 12)

	// 异常牌型0605-1
	// handCards := []uint32{
	// 	NewCard(2, 14), NewCard(3, 14),
	// 	NewCard(1, 14), NewCard(1, 2), NewCard(1, 3),
	// 	NewCard(3, 13), NewCard(4, 11), NewCard(2, 11),
	// 	NewCard(1, 10), NewCard(4, 10), NewCard(1, 11),
	// 	NewCard(1, 11), NewCard(4, 10),
	// }
	// wildCard := NewCard(4, 11)

	// GetTouchType timeout: 225123284832485573, 11s,
	// cards=[方块3,方块9,方块9,方块J,方块J,方块K,梅花7,梅花8,红桃2,红桃3,黑桃9,黑桃J,JOKER1,], qiCards=[黑桃6,梅花7,黑桃8,红桃K,红桃8,梅花4,方块7,], wildCard=57
	// 异常牌型0605-2
	// handCards := []uint32{
	// 	NewCard(1, 3), NewCard(1, 9), RMJOKERS[0],
	// 	NewCard(1, 9), NewCard(1, 11), NewCard(1, 11),
	// 	NewCard(1, 13), NewCard(2, 7), NewCard(2, 8),
	// 	NewCard(3, 2), NewCard(3, 2), NewCard(3, 3),
	// 	NewCard(4, 11), // NewCard(4, 11),
	// }
	// wildCard := uint32(57)

	// 异常牌型0605-3
	// handCards := []uint32{
	// 	NewCard(2, 8), NewCard(2, 9), NewCard(2, 10),
	// 	NewCard(3, 14), NewCard(3, 12), NewCard(2, 10),
	// 	NewCard(2, 9), NewCard(3, 10), NewCard(1, 10),
	// 	NewCard(4, 12), NewCard(4, 13), NewCard(4, 13),
	// 	NewCard(2, 14), NewCard(1, 13),
	// }
	// wildCard := NewCard(1, 10)

	// 异常牌型0605-4
	// handCards := []uint32{
	// 	NewCard(3, 6), NewCard(4, 13), NewCard(4, 13),
	// 	NewCard(3, 4), NewCard(1, 3), NewCard(4, 14),
	// 	NewCard(4, 12), NewCard(3, 14), NewCard(3, 3),
	// 	NewCard(3, 5), NewCard(2, 3), RMJOKERS[0],
	// 	NewCard(4, 4), NewCard(4, 11),
	// }
	// wildCard := NewCard(4, 3)

	// 异常牌型0607-1
	// handCards := []uint32{
	// 	NewCard(3, 9), NewCard(3, 10), NewCard(3, 11),
	// 	NewCard(1, 2), NewCard(4, 2), RMJOKERS[1],
	// 	NewCard(2, 14), NewCard(2, 2), NewCard(2, 12),
	// 	NewCard(1, 10), NewCard(3, 10), NewCard(4, 10),
	// 	NewCard(2, 9), NewCard(3, 2),
	// }
	// wildCard := NewCard(1, 12)

	// 异常牌型0607-2
	// handCards := []uint32{
	// 	NewCard(1, 14), NewCard(1, 13), NewCard(1, 2),
	// 	NewCard(1, 11), NewCard(2, 11), NewCard(2, 11),
	// 	RMJOKERS[0], NewCard(4, 13), NewCard(2, 13),
	// 	NewCard(2, 2), NewCard(3, 14), NewCard(2, 14),
	// 	RMJOKERS[1], NewCard(4, 14),
	// }
	// wildCard := NewCard(2, 2)
	// qiCards := []uint32{
	// 	76, 56, 54, 52, 26, 77, 82,
	// 	43, 26, 38, 17, 27, 41, 29,
	// }
	// 异常牌型0613-1
	// handCards := []uint32{
	// 	NewCard(4, 3), NewCard(4, 4), NewCard(4, 5),
	// 	NewCard(2, 9), NewCard(2, 10), NewCard(4, 8),
	// 	NewCard(4, 3), NewCard(4, 4), NewCard(4, 8),
	// 	NewCard(4, 5), NewCard(4, 7), NewCard(4, 9),
	// 	NewCard(1, 5),
	// }
	// wildCard := NewCard(1, 8)
	// 异常牌型0613-1
	// handCards := []uint32{
	// 	NewCard(3, 10), NewCard(3, 11), NewCard(3, 12),
	// 	NewCard(1, 11), NewCard(1, 12), NewCard(4, 4),
	// 	NewCard(4, 11), NewCard(2, 4), NewCard(2, 10),
	// 	NewCard(1, 11), NewCard(3, 14), NewCard(3, 2),
	// 	NewCard(3, 4),
	// }
	// wildCard := NewCard(2, 4)

	// handCards := []uint32{
	// 	22, 33, 33, 58, 45, 34, 45, 23, 57, 59, 25, 41, 38,
	// }
	// handCards := []uint32{
	// 	69, 27, 75, 75, 26, 28, 74, 70, 74, 76, 66, RMJOKERS[1], 29,
	// }
	// wildCard := NewCard(2, 4)
	// handCards := []uint32{
	// 	75, 36, 75, 57, 35, 37, 74, 76, 56, 58, 81, RMJOKERS[1], 44,
	// }
	// handCards := []uint32{ 53, 22, 70, 66, 21, 23, 69, 54, 65, 67, 44, 44, 17}

	handCards := [][]uint32{
		{NewCard(1, 5), NewCard(1, 6), NewCard(1, 7)},
		{NewCard(4, 8), NewCard(4, 9), NewCard(2, 2)},
		{NewCard(1, 11), NewCard(3, 11), NewCard(4, 11)},
		{NewCard(2, 5), NewCard(1, 2), NewCard(1, 5), NewCard(3, 2)},
		// {NewCard(3, 2)},
	}
	// handCards := []uint32{
	// 	17, 44, 44, 53, 65, 66, 67, 69, 70, 24, 25, 49, 67,
	// }
	wildCard := NewCard(2, 2)
	fmt.Println("===============================")
	fmt.Println(handCards)

	ret := IsFinish(handCards, wildCard)
	fmt.Println("finish:", ret)

	// sec := time.Now().Unix()
	// area := GetTouchType(handCards, []uint32{NewCard(1, 7)}, wildCard)
	// delay := time.Now().Unix() - sec
	// fmt.Printf("touch: %d %d\n", delay, area)

	// fmt.Printf("wildCard=%s\n", GetSuitString(wildCard)+GetRankString(wildCard))
	// fmt.Printf("cards=%s\n", CardsString(handCards))

	// groups := GroupTheCards(handCards, wildCard)
	// dGroups := &HandCardDivide{BaseGroups: groups}
	// fmt.Println("GroupTheCards: ", dGroups)

	// handCards = append(handCards, NewCard(2, 14))
	// outCard, h, finish := LetsOutCard(handCards, []uint32{NewCard(2, 2)}, wildCard)
	// fmt.Println(h)
	// fmt.Println(CardsString([]uint32{outCard}), finish)
	// if finish {
	// 	ret := IsFinish(h.BaseGroups, wildCard)
	// 	fmt.Println("finish:", ret)
	// 	for _, g := range h.BaseGroups {
	// 		fmt.Println("shun: ", IsTonghuashun(g, wildCard), CardsString(g))
	// 	}
	// }

	// hands, lefts := SortCards(handCards, wildCard)
	// divide := &HandCardDivide{BaseGroups: hands, LeftCards: lefts}
	// fmt.Printf("sort=%s\n", divide)
	// finish, discard := GetOneCardToDiscard(handCards, wildCard)
	// fmt.Printf("discard=%v, %s\n", finish, GetSuitString(discard)+GetRankString(discard))
}

func NewCard(suit, rank uint32) uint32 {
	switch suit {
	case 1:
		suit = Diamond
	case 2:
		suit = Club
	case 3:
		suit = Heart
	case 4:
		suit = Spade
	}
	switch rank {
	case 14:
		rank = Ace
	default:
		rank = rank
	}
	return suit | rank
}

func TestLetsOutCardRand(t *testing.T) {
	handCards := []uint32{33, 34, 35, 19, 17, 18, 49, 61, 42, 39, 54, 76, 71}
	wildCard := uint32(42)

	area := GetTouchType(handCards, []uint32{29}, wildCard)
	fmt.Println("area: ", area)
}

func TestFindLackGroupNeedCards(t *testing.T) {
	lackGroups := [][]uint32{
		{34, 35},
		{33, 35},
		{RMCARDS[0], RMCARDS[13]},
	}

	for _, group := range lackGroups {
		needs := FindLackGroupNeedCards(group)
		fmt.Printf("%s needs %s\n", CardsString(group), CardsString(needs))
	}

}

func Test2(t *testing.T) {
	cards := []uint32{RMJOKERS[0], RMJOKERS[0], NewCard(2, 6)}
	fmt.Println(CardsString(cards))
	fmt.Println(ValidRmGroup(cards, NewCard(4, 12)))
}

// 从手牌中抽牌替换
func deal3Draw(hand, cards []uint32, drawNum int) {
	for i := 0; i < drawNum; i++ {
		m, n := utils.RandIntN(len(hand)), utils.RandIntN(len(cards))
		hand[m], cards[n] = cards[n], hand[m]
	}
}

func TestNeedCardsNext(t *testing.T) {
	fmt.Println("2")
	// todo 发天胡牌, ... next, ...next, ...group
	// todo 发散牌

	var d []uint32
	d = append(d, RMCARDS...)
	d = append(d, RMCARDS...)

	//测试暂时去掉洗牌
	for i := len(d) - 1; i > 0; i-- {
		j := utils.RandIntN(i + 1)
		d[i], d[j] = d[j], d[i]
	}

	wildCard := d[0] //万能牌
	_ = wildCard
	qiCards := []uint32{d[1]} //弃牌堆
	_ = qiCards
	cards := d[2:]

	cards = append(cards, RMJOKERS...)
	cards = append(cards, RMJOKERS...)
	for i := len(cards) - 1; i > 0; i-- {
		j := utils.RandIntN(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}

	// 随机各发13张牌
	// var cardsCount = make(map[uint32]int, 16)
	// var hand = 13
	// for i := 0; i < 6; i++ {
	// 	myHand := make([]uint32, 0, hand)
	// 	myHand = append(myHand, cards[:hand]...)
	// 	cards = cards[hand:]
	// 	fmt.Println(algo.CardsString(myHand))
	// 	checkRepeat(myHand, cardsCount)
	// }
	fmt.Println("wildCard: ", CardsString([]uint32{wildCard}))

	huCt := []string{"N01", "N01", "N01", "N01", "E01", "F01", "E02", "F0E", "E03", "F03", "N02", "N02", "G04"}
	var hand []uint32
	hand, cards = RMGetCard(huCt, cards, wildCard)
	deal3Draw(hand, cards, 4)

	fmt.Println("------------------------------")
	group := GroupTheCards(hand, wildCard)
	fmt.Println(&HandCardDivide{BaseGroups: group})
	cardList, finish := NeedNextCardList(hand, nil, cards, wildCard)
	fmt.Println(CardsString(cardList))
	if finish {
		fmt.Println("finish1")
	}

	for i := 0; i < 10; i++ {
		// fmt.Println("need: ", CardsString([]uint32{nextCard}))

		// cards = dealCard(cards, nextCard)
		// hand = append(hand, nextCard)
		// outCard, divide, finish := LetsOutCard(hand, qiCards, wildCard)
		// fmt.Println("outCard:", CardsString([]uint32{outCard}))
		// fmt.Println("divide:", divide)
		// if finish {
		// 	fmt.Println("finish2")
		// 	break
		// }
		// qiCards = append(qiCards, outCard)
		// hand = make([]uint32, 0, 13)
		// for _, group := range divide.BaseGroups {
		// 	hand = append(hand, group...)
		// }
		// hand = append(hand, divide.LeftCards...)
	}
}

func TestAB(t *testing.T) {
	group := []uint32{18, 66, 50}
	ranks := GetRanks(group)
	fmt.Println(ranks)
	fmt.Println(IsBaozi(group, 12))

	handCards := []uint32{
		NewCard(3, 9), NewCard(3, 10), NewCard(3, 11),
		NewCard(1, 2), NewCard(4, 2), RMJOKERS[1],
		NewCard(2, 14), NewCard(2, 2), NewCard(2, 12),
		NewCard(1, 10), NewCard(3, 10), NewCard(4, 10),
		NewCard(2, 9), NewCard(3, 2),
	}
	fmt.Println("cards: ", handCards)
	fmt.Println("cards: ", CardsString(handCards))
	sets := FindAllBaseSet(handCards)
	fmt.Println("sets: ", sets)
}

func TestNeedCard(t *testing.T) {
	for i := 0; i < 10; i++ {
		fmt.Println("============================" + strconv.Itoa(i))
		testNeedCard(t)
	}
}

func testNeedCard(t *testing.T) {
	fmt.Println("4")
	// todo 发天胡牌, ... next, ...next, ...group
	// todo 发散牌

	var d []uint32
	d = append(d, RMCARDS...)
	d = append(d, RMCARDS...)

	//测试暂时去掉洗牌
	for i := len(d) - 1; i > 0; i-- {
		j := utils.RandIntN(i + 1)
		d[i], d[j] = d[j], d[i]
	}

	wildCard := d[0] //万能牌
	_ = wildCard
	qiCards := []uint32{d[1]} //弃牌堆
	_ = qiCards
	cards := d[2:]

	cards = append(cards, RMJOKERS...)
	cards = append(cards, RMJOKERS...)
	for i := len(cards) - 1; i > 0; i-- {
		j := utils.RandIntN(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}

	// 随机各发13张牌
	// var cardsCount = make(map[uint32]int, 16)
	// var hand = 13
	// for i := 0; i < 6; i++ {
	// 	myHand := make([]uint32, 0, hand)
	// 	myHand = append(myHand, cards[:hand]...)
	// 	cards = cards[hand:]
	// 	fmt.Println(algo.CardsString(myHand))
	// 	checkRepeat(myHand, cardsCount)
	// }
	fmt.Println("wildCard: ", CardsString([]uint32{wildCard}))

	huCt := []string{"N01", "N01", "N01", "N01", "E01", "F01", "E02", "F0E", "E03", "F03", "N02", "N02", "G04"}
	var hand []uint32
	hand, cards = RMGetCard(huCt, cards, wildCard)
	deal3Draw(hand, cards, 4)

	for i := 0; i < 10; i++ {
		fmt.Println("------------------------------" + strconv.Itoa(i))
		group := GroupTheCards(hand, wildCard)
		fmt.Println(&HandCardDivide{BaseGroups: group})
		nextCard, finish := NeedNextCard(hand, nil, cards, wildCard)
		if finish {
			fmt.Println("finish1")
			break
		}
		fmt.Println("need: ", CardsString([]uint32{nextCard}))

		cards = dealCard(cards, nextCard)
		hand = append(hand, nextCard)
		outCard, divide, finish := LetsOutCard(hand, qiCards, wildCard)
		fmt.Println("outCard:", CardsString([]uint32{outCard}))
		fmt.Println("divide:", divide)
		if finish {
			fmt.Println("finish2")
			break
		}
		qiCards = append(qiCards, outCard)
		hand = make([]uint32, 0, 13)
		for _, group := range divide.BaseGroups {
			hand = append(hand, group...)
		}
		hand = append(hand, divide.LeftCards...)
	}
}

func dealCard(cards []uint32, card uint32) []uint32 {
	for i, c := range cards {
		if c == card {
			return append(cards[:i], cards[i+1:]...)
		}
	}
	fmt.Println("not found card: ", card)
	return cards
}

func TestFinish(t *testing.T) {
	finish := [][]uint32{
		{29,
			28,
			17,
			50},
		{
			41,
			42,
			43,
		},
		{
			41,
			57,
			18,
		},
		{
			26,
			58,
			74,
		},
	}
	a := IsFinish(finish, NewCard(1, 2))
	fmt.Println("r", a)
}
