package algo

import (
	"math/rand"
	"sort"
	"time"
)

const (
	HgihCard uint32 = iota + 0x00
	Niu1
	Niu2
	Niu3
	Niu4
	Niu5
	Niu6
	Niu7
	Niu8
	Niu9
	NiuNiu
	Straight
	FullHouse
	Flush
	FiveFlower
	Bomb
	StraightFlush
	FiveTiny
)

// rank 牌值
const (
	Ace   uint32 = iota + 0x01 //A
	Deuce                      //2
	Trey                       //3
	Four                       //4
	Five                       //5
	Six                        //6
	Seven                      //7
	Eight                      //8
	Nine                       //9
	Ten                        //10
	Jack                       //J
	Queen                      //Q
	King                       //K

	RankMask uint32 = 0x0F
)

// suit 花色
const (
	Spade   uint32 = 0x40 //黑桃  64
	Heart   uint32 = 0x30 //红桃  48
	Club    uint32 = 0x20 //梅花  32
	Diamond uint32 = 0x10 //方块  16

	SuitMask uint32 = 0xF0
)

const (
	NumCard = 52
)

func Rank(card uint32) uint32 {
	return card & RankMask
}

func Suit(card uint32) uint32 {
	return card & SuitMask
}

var NiuCARDS = []uint32{
	0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, //方块
	0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, //梅花
	0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, //红桃
	0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, //黑桃
}
//  A     2     3     4     5     6     7     8     9     10    J     Q     K
//  0     1     2     3     4     5     6     7     8     9     10    11    12    //方
//  13    14    15    16    17    18    19    20    21    22    23    24    25    //梅
//  26    27    28    29    30    31    32    33    34    35    36    37    38    //红
//  39    40    41    42    43    44    45    46    47    48    49    50    51    //黑

var CARDSIndex = map[uint32]int{
	0x11: 0, 0x12: 1, 0x13: 2, 0x14: 3, 0x15: 4, 0x16: 5, 0x17: 6, 0x18: 7, 0x19: 8, 0x1a: 9, 0x1b: 10, 0x1c: 11, 0x1d: 12,
	0x21: 13, 0x22: 14, 0x23: 15, 0x24: 16, 0x25: 17, 0x26: 18, 0x27: 19, 0x28: 20, 0x29: 21, 0x2a: 22, 0x2b: 23, 0x2c: 24, 0x2d: 25,
	0x31: 26, 0x32: 27, 0x33: 28, 0x34: 29, 0x35: 30, 0x36: 31, 0x37: 32, 0x38: 33, 0x39: 34, 0x3a: 35, 0x3b: 36, 0x3c: 37, 0x3d: 38,
	0x41: 39, 0x42: 40, 0x43: 41, 0x44: 42, 0x45: 43, 0x46: 44, 0x47: 45, 0x48: 46, 0x49: 47, 0x4a: 48, 0x4b: 49, 0x4c: 50, 0x4d: 51,
}

func IndexsToCards(indexs []int) []uint32 {
	c1 := NiuCARDS[indexs[0]]
	c2 := NiuCARDS[indexs[1]]
	c3 := NiuCARDS[indexs[2]]
	return []uint32{c1, c2, c3}
}

func CardsToIndexs(cards []uint32) (indexs []int) {
	return []int{
		CARDSIndex[cards[0]],
		CARDSIndex[cards[1]],
		CARDSIndex[cards[2]],
	}
}

var NIUS [][]int = [][]int{{0, 1, 2}, {0, 1, 3}, {0, 2, 3}, {1, 2, 3}, {0, 1, 4}, {0, 2, 4}, {1, 2, 4}, {0, 3, 4}, {1, 3, 4}, {2, 3, 4}}
var NIUL [][]int = [][]int{{3, 4}, {2, 4}, {1, 4}, {0, 4}, {2, 3}, {1, 3}, {0, 3}, {1, 2}, {0, 2}, {0, 1}}

// Algo []uint32{1, 5, 8, 9, K}
func Algo(mode uint32, cs []uint32) uint32 {
	if len(cs) != 5 {
		return 0
	}
	descSort(cs)
	var niu uint32 = HgihCard
	if mode != 0 {
		niu = Algo1(cs)
		if niu != 0 {
			return niu
		}
	}
	for k, v := range NIUS {
		if ((Trunc(cs[v[0]]) + Trunc(cs[v[1]]) + Trunc(cs[v[2]])) % 10) != 0 {
			continue
		}
		switch (Trunc(cs[NIUL[k][0]]) + Trunc(cs[NIUL[k][1]])) % 10 {
		case 0:
			return NiuNiu
		case 1:
			niu = max(niu, Niu1)
		case 2:
			niu = max(niu, Niu2)
		case 3:
			niu = max(niu, Niu3)
		case 4:
			niu = max(niu, Niu4)
		case 5:
			niu = max(niu, Niu5)
		case 6:
			niu = max(niu, Niu6)
		case 7:
			niu = max(niu, Niu7)
		case 8:
			niu = max(niu, Niu8)
		case 9:
			niu = max(niu, Niu9)
		}
	}
	return niu
}

// Algo1 原有特殊玩法
func Algo1(cs []uint32) uint32 {
	bomb_n := make(map[uint32]int)
	var tiny_n int
	var tiny_v uint32
	var flower int
	var ten int
	for _, v := range cs {
		bomb_n[Rank(v)] += 1
		switch Rank(v) {
		case Jack, Queen, King:
			flower++
		case Ten:
			ten++
		case Ace, Deuce, Trey, Four, Five, Six:
			tiny_n++
			tiny_v += Rank(v)
		}
	}
	if tiny_n == 5 && tiny_v <= Ten {
		return FiveTiny
	}
	niu := Algo2(cs)
	if niu == StraightFlush {
		return niu
	}
	for _, v := range bomb_n {
		if v == 4 {
			return Bomb
		}
	}
	if flower == 5 {
		return FiveFlower
	}
	return niu
}

// Algo2 新加特殊玩法
func Algo2(cs []uint32) uint32 {
	var straight bool
	var flush bool
	cards := make([]hands, len(cs))
	for k, v := range cs {
		cards[k].Suit = Suit(v)
		cards[k].Rank = Rank(v)
	}
	ascSortHands(cards)
	if cards[0].Suit == cards[1].Suit &&
		cards[1].Suit == cards[2].Suit &&
		cards[2].Suit == cards[3].Suit &&
		cards[3].Suit == cards[4].Suit {
		flush = true
	}
	if (cards[0].Rank+1) == cards[1].Rank &&
		(cards[1].Rank+1) == cards[2].Rank &&
		(cards[2].Rank+1) == cards[3].Rank &&
		(cards[3].Rank+1) == cards[4].Rank {
		straight = true
	}
	if straight && flush {
		return StraightFlush
	}
	if flush {
		return Flush
	}
	if (cards[0].Rank == cards[1].Rank &&
		cards[1].Rank == cards[2].Rank &&
		cards[3].Rank == cards[4].Rank) ||
		(cards[1].Rank == cards[2].Rank &&
			cards[2].Rank == cards[3].Rank &&
			cards[0].Rank == cards[4].Rank) ||
		(cards[2].Rank == cards[3].Rank &&
			cards[3].Rank == cards[4].Rank &&
			cards[0].Rank == cards[1].Rank) {
		return FullHouse
	}
	if straight {
		return Straight
	}
	return 0
}

// Trunc 取整
func Trunc(n uint32) uint32 {
	if Rank(n) > Ten {
		return Ten
	}
	return Rank(n)
}

// 取大值
func max(n, m uint32) uint32 {
	if n > m {
		return n
	}
	return m
}

// 降序排序
func descSort(cards []uint32) {
	sort.Slice(cards, func(i, j int) bool {
		return cards[i] >= cards[j]
	})
}

// Equal 比较 a == b
func Equal(a, b []uint32) bool {
	if len(a) != len(b) {
		return false
	}

	if (a == nil) != (b == nil) {
		return false
	}

	descSort(a)
	descSort(b)
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

// 比较 a >= b
func Compare2(a, b []uint32) bool {
	if len(a) != len(b) {
		return false
	}

	if (a == nil) != (b == nil) {
		return false
	}

	descSort(a)
	descSort(b)
	for i, v := range a {
		if v < b[i] {
			return false
		}
	}

	return true
}

type hands struct {
	Suit uint32 //花色
	Rank uint32 //牌值
}

// 降序排序
func descSortHands(cards []hands) {
	sort.Slice(cards, func(i, j int) bool {
		if cards[i].Rank == cards[j].Rank {
			return cards[i].Suit > cards[j].Suit
		}
		return cards[i].Rank > cards[j].Rank
	})
}

// 升序排序
func ascSortHands(cards []hands) {
	sort.Slice(cards, func(i, j int) bool {
		if cards[i].Rank == cards[j].Rank {
			return cards[i].Suit < cards[j].Suit
		}
		return cards[i].Rank < cards[j].Rank
	})
}

// Compare 比较 a >= b (先比较牌值,牌值相同再比较花)
// 同等牛的比其中牌值最大的一个,如果最大的一个牌值一样,则比花色(相同牌花色永远不同)
func Compare(a, b []uint32) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	if len(a) != len(b) {
		return false
	}

	if (a == nil) != (b == nil) {
		return false
	}

	a1 := make([]hands, len(a))
	b1 := make([]hands, len(b))
	for i, v := range a {
		a1[i].Suit = Suit(v)
		a1[i].Rank = Rank(v)
	}
	for i, v := range b {
		b1[i].Suit = Suit(v)
		b1[i].Rank = Rank(v)
	}

	descSortHands(a1)
	descSortHands(b1)
	//牌值比较
	if a1[0].Rank == b1[0].Rank {
		return a1[0].Suit > b1[0].Suit
	}
	return a1[0].Rank > b1[0].Rank
}

// Compare3 比较 a >= b (先比较牌值,牌值相同再比较花)
func Compare3(a, b []uint32) bool {
	if len(a) != len(b) {
		return false
	}

	if (a == nil) != (b == nil) {
		return false
	}

	a1 := make([]hands, len(a))
	b1 := make([]hands, len(b))
	for i, v := range a {
		a1[i].Suit = Suit(v)
		a1[i].Rank = Rank(v)
	}
	for i, v := range b {
		b1[i].Suit = Suit(v)
		b1[i].Rank = Rank(v)
	}

	descSortHands(a1)
	descSortHands(b1)
	//牌值比较
	for i, v := range a1 {
		if v.Rank < b1[i].Rank {
			return false
		} else if v.Rank > b1[i].Rank {
			return true
		}
	}

	//牌值相同,比较花色
	for i, v := range a1 {
		if v.Suit < b1[i].Suit {
			return false
		} else if v.Suit > b1[i].Suit {
			return true
		}
	}

	return true
}

// Multiple 积分倍数
func Multiple(mode, n uint32) uint32 {
	if mode != 0 {
		return Multiple1(n)
	}
	switch n {
	case HgihCard, Niu1, Niu2, Niu3, Niu4, Niu5, Niu6, Niu7:
		return 1
	case Niu8:
		return 2
	case Niu9:
		return 3
	case NiuNiu:
		return 4
	}
	return 1
}

// Multiple1 积分倍数
func Multiple1(n uint32) uint32 {
	switch n {
	case HgihCard, Niu1:
		return 1
	case Niu2:
		return 2
	case Niu3:
		return 3
	case Niu4:
		return 4
	case Niu5:
		return 5
	case Niu6:
		return 6
	case Niu7:
		return 7
	case Niu8:
		return 8
	case Niu9:
		return 9
	case NiuNiu:
		return 10
	case Straight:
		return 11
	case FullHouse:
		return 12
	case Flush:
		return 13
	case FiveFlower:
		return 14
	case Bomb:
		return 15
	case StraightFlush:
		return 16
	case FiveTiny:
		return 17
	}
	return 1
}

// AlgoVerify []uint32{1, 5, 8, 9, K}
func AlgoVerify(cs []uint32, val uint32) bool {
	if len(cs) != 5 {
		return false
	}
	descSort(cs)
	bomb_n := make(map[uint32]int)
	var tiny_n int
	var tiny_v uint32
	var flower int
	var ten int
	for _, v := range cs {
		bomb_n[Rank(v)] += 1
		switch Rank(v) {
		case Jack, Queen, King:
			flower++
		case Ten:
			ten++
		case Ace, Deuce, Trey, Four, Five, Six:
			tiny_n++
			tiny_v += Rank(v)
		}
	}
	if tiny_n == 5 && tiny_v <= Ten && val == FiveTiny {
		return true
	}
	for _, v := range bomb_n {
		if v == 4 && val == Bomb {
			return true
		}
	}
	if flower == 5 && val == FiveFlower {
		return true
	}
	for k, v := range NIUS {
		if ((Trunc(cs[v[0]]) + Trunc(cs[v[1]]) + Trunc(cs[v[2]])) % 10) != 0 {
			continue
		}
		switch (Trunc(cs[NIUL[k][0]]) + Trunc(cs[NIUL[k][1]])) % 10 {
		case 0:
			if val == NiuNiu {
				return true
			}
		case 1:
			if val == Niu1 {
				return true
			}
		case 2:
			if val == Niu2 {
				return true
			}
		case 3:
			if val == Niu3 {
				return true
			}
		case 4:
			if val == Niu4 {
				return true
			}
		case 5:
			if val == Niu5 {
				return true
			}
		case 6:
			if val == Niu6 {
				return true
			}
		case 7:
			if val == Niu7 {
				return true
			}
		case 8:
			if val == Niu8 {
				return true
			}
		case 9:
			if val == Niu9 {
				return true
			}
		}
	}
	return false
}

// Remove 移除一个牌
func Remove(c uint32, cs []uint32) []uint32 {
	for i, v := range cs {
		if c == v {
			cs = append(cs[:i], cs[i+1:]...)
			break
		}
	}
	return cs
}

// SameCard 移除一个牌
func SameCard(cs, hs []uint32) bool {
	for _, c := range cs {
		for _, h := range hs {
			if c == h {
				return true
			}
		}
	}
	return false
}

// VerifyCard 验证手牌(设置时存在0)
func VerifyCard(cs []uint32) bool {
	for _, c := range cs {
		if c == 0 {
			continue
		}
		switch Suit(c) {
		case Club, Heart, Spade, Diamond:
		default:
			return false
		}
		switch Rank(c) {
		case Ace, Deuce, Trey, Four, Five, Six:
		case Seven, Eight, Nine, Ten, Jack, Queen, King:
		default:
			return false
		}
	}
	return true
}

func GetSuitString(c uint32) string {
	var ret string
	suit := Suit(c)
	switch suit {
	case Spade:
		ret += "黑桃"
	case Heart:
		ret += "红桃"
	case Club:
		ret += "梅花"
	case Diamond:
		ret += "方块"
	}
	return ret
}

func GetRankString(c uint32) string {
	var ret string
	rank := Rank(c)
	switch rank {
	case Ace:
		ret += "A"
	case Deuce:
		ret += "2"
	case Trey:
		ret += "3"
	case Four:
		ret += "4"
	case Five:
		ret += "5"
	case Six:
		ret += "6"
	case Seven:
		ret += "7"
	case Eight:
		ret += "8"
	case Nine:
		ret += "9"
	case Ten:
		ret += "10"
	case Jack:
		ret += "J"
	case Queen:
		ret += "Q"
	case King:
		ret += "K"
	}
	return ret
}

// '洗牌
func ShuffleCards() []uint32 {
	rand.Seed(time.Now().UnixNano())
	d := make([]uint32, NumCard, NumCard)
	copy(d, NiuCARDS)
	//测试暂时去掉洗牌
	for i := len(d) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		d[i], d[j] = d[j], d[i]
	}
	return d
}
