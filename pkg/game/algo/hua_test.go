package algo

import (
	"fmt"
	"goserver/pkg/utils"
	"math/rand"
	"sort"
	"testing"
	"time"
)

// 测试
func TestHua(t *testing.T) {
	// cs := []uint32{0x33, 0x23, 0x17}
	// t.Log(toHands(cs))
	// cs = []uint32{0x1d, 0x4b, 0x41}
	// t.Log(toHands(cs))
	// cs = []uint32{0x4d, 0x4c, 0x41}
	// t.Log(Hua(cs))
	// cs = []uint32{0x1d, 0x2d, 0x3d}
	// t.Log(Hua(cs))
	// //
	// cs1 := []uint32{0x19, 0x25, 0x35}
	// cs2 := []uint32{0x12, 0x15, 0x45}
	// t.Log(HuaCompare(cs1, cs2))
	// cs1 = []uint32{0x19, 0x25, 0x35}
	// cs2 = []uint32{0x12, 0x11, 0x41}
	// t.Log(HuaCompare(cs1, cs2))

	a := []uint32{0x24, 0x34, 0x22}
	b := []uint32{0x44, 0x14, 0x32}

	result := HuaCompare(a, b)
	t.Log(result)
	// hs1 := toHands(cs1)
	// t.Log(hs1)
	// t.Log(isShun(hs1))

}

func TestHuaType(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	cards1 := []uint32{NewCard(1, 5), NewCard(3, 2), NewCard(1, 2)}
	cards2 := []uint32{NewCard(1, 1), NewCard(2, 1), NewCard(3, 1)}
	cards3 := []uint32{NewCard(1, 3), NewCard(2, 11), NewCard(3, 11)}
	fmt.Println(CardsString(cards1))
	fmt.Println(CardsString(cards2))
	fmt.Println(CardsString(cards3))
	for i := 0; i < 101; i++ {
		d := make([]uint32, NumCard)
		copy(d, NiuCARDS)
		d, _ = removeCards(d, cards1)
		d, _ = removeCards(d, cards2)
		d, _ = removeCards(d, cards3)

		toCardss, remain := GetCardssHuaTypeBiggerN([][]uint32{cards1, cards2, cards3}, d, 2)
		if len(remain) != 43 {
			t.Errorf("remain cards: %d\n", len(remain))
			return
		}
		// toCardss = toCardss[1:]
		for _, cards := range toCardss {
			fmt.Println(CardsString(cards))
			if hasRepeat(cards) {
				t.Errorf("has repeat %v", cards)
				return
			}
		}
	}

	// t1 := HuaTypeUpOrDown(cards1)
	// t2 := HuaTypeUpOrDown(cards2)

	// // bigger := HuaTypeUpOrDown(c) - huaType
	// bigger := t2 - t1
	// biggerLevel := bigger/9 + bigger%9
	// fmt.Printf("%d, %d\n", t1, t2)
	// fmt.Println(biggerLevel)
}

func hasRepeat(cards []uint32) bool {
	return cards[0] == cards[1] || cards[0] == cards[2] || cards[1] == cards[2]
}

func TestGetCardsHuaTypeLessN(t *testing.T) {
	// cards1 := []uint32{NewCard(1, 5), NewCard(3, 2), NewCard(1, 2)}
	cards1 := []uint32{NewCard(1, 8), NewCard(2, 9), NewCard(3, 10)}
	// cards3 := []uint32{NewCard(1, 2), NewCard(2, 10), NewCard(3, 10)}
	fmt.Println(CardsString(cards1))

	rand.Seed(time.Now().UnixNano())
	d := make([]uint32, NumCard)
	copy(d, NiuCARDS)
	//洗牌
	for i := len(d) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		d[i], d[j] = d[j], d[i]
	}
	d, _ = removeCards(d, cards1)
	sort.Slice(d, func(i, j int) bool { return utils.RandBool() }) // 洗牌

	cards, remain := GetCardsHuaTypeLessN(cards1, d, 2)
	fmt.Println(len(remain))
	fmt.Println(CardsString(cards))
	fmt.Println(5)
}

func TestGetCardHuaTypeBigger1(t *testing.T) {
	// cards1 := []uint32{NewCard(1, 5), NewCard(3, 2), NewCard(1, 2)}
	cards1 := []uint32{NewCard(1, 8), NewCard(2, 9), NewCard(3, 10)}
	// cards3 := []uint32{NewCard(1, 2), NewCard(2, 10), NewCard(3, 10)}
	fmt.Println(CardsString(cards1))

	rand.Seed(time.Now().UnixNano())
	d := make([]uint32, NumCard)
	copy(d, NiuCARDS)
	//洗牌
	for i := len(d) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		d[i], d[j] = d[j], d[i]
	}
	d, _ = removeCards(d, cards1)
	sort.Slice(d, func(i, j int) bool { return utils.RandBool() }) // 洗牌

	cards, remain := GetCardHuaTypeBigger1(cards1, d)
	fmt.Println(len(remain))
	fmt.Println(CardsString(cards))
	fmt.Println(7)
}

func TestGetCardsHuaTypeBugger(t *testing.T) {
	d := make([]uint32, NumCard)
	copy(d, NiuCARDS)
	//洗牌
	for i := len(d) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		d[i], d[j] = d[j], d[i]
	}

	cards, remain := GetCardsHuaTypeBugger(TonghuaDown, 10, d)
	fmt.Println(len(remain))
	for _, c := range cards {
		fmt.Println(CardsString(c))
	}
}
