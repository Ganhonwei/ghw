package algo

import (
	"goserver/pkg/utils"
	"sort"
)

func jokerSort(cards []uint32) { //从小到大
	sort.Slice(cards, func(i, j int) bool {
		if Rank(cards[i]) != Rank(cards[j]) {
			return Rank(cards[i]) < Rank(cards[j])
		} else {
			return Suit(cards[i]) < Suit(cards[j])
		}
	})
}

// 对子(组豹子)
func IsA(cards []uint32) bool {
	return Rank(cards[0]) == Rank(cards[1])
}

// 缺边张同花点数或缺中张同花点数(组同花顺)
func IsB(cards []uint32) bool {
	new := append([]uint32{}, cards...)
	jokerSort(new)

	a := new[0]
	b := new[1]

	if Suit(a) != Suit(b) {
		return false
	}

	if Rank(a) == Ace && Rank(b) == King {
		return true
	} else if Rank(a) == Ace && Rank(b) == Queen {
		return true
	} else if Rank(a)+1 == Rank(b) {
		return true
	} else if Rank(a)+2 == Rank(b) {
		return true
	}
	return false
}

// 缺边张非同花点数或缺中张非同花点数(组顺子)
func IsC(cards []uint32) bool {
	new := append([]uint32{}, cards...)
	jokerSort(new)
	a := new[0]
	b := new[1]

	if Suit(a) == Suit(b) {
		return false
	}

	if Rank(a) == Ace && Rank(b) == King {
		return true
	} else if Rank(a) == Ace && Rank(b) == Queen {
		return true
	} else if Rank(a)+1 == Rank(b) {
		return true
	} else if Rank(a)+2 == Rank(b) {
		return true
	}
	return false
}

// 非靠张和隔张的同花(组同花)
func IsD(cards []uint32) bool {
	new := append([]uint32{}, cards...)
	jokerSort(new)
	a := new[0]
	b := new[1]

	if Suit(a) != Suit(b) {
		return false
	}

	if Rank(a) == Ace && Rank(b) == King {
		return false
	} else if Rank(a) == Ace && Rank(b) == Queen {
		return false
	} else if Rank(a)+1 == Rank(b) {
		return false
	} else if Rank(a)+2 == Rank(b) {
		return false
	}
	return true
}

// 非靠张和隔张两张非同花（对子）
func IsE(cards []uint32) bool {
	new := append([]uint32{}, cards...)
	jokerSort(new)
	a := new[0]
	b := new[1]

	if Suit(a) == Suit(b) {
		return false
	}

	if Rank(a) == Ace && Rank(b) == King {
		return false
	} else if Rank(a) == Ace && Rank(b) == Queen {
		return false
	} else if Rank(a)+1 == Rank(b) {
		return false
	} else if Rank(a)+2 == Rank(b) {
		return false
	}
	return true
}

func ExistRank(cards []uint32, rank uint32) bool {
	for _, v := range cards {
		if Rank(v) == rank {
			return true
		}
	}
	return false
}

func ExistSuit(cards []uint32, suit uint32) bool {
	for _, v := range cards {
		if Suit(v) == suit {
			return true
		}
	}
	return false
}

func ChangeA(cards []uint32) uint32 {
	rank := Rank(cards[0])
	var suit uint32

	if !ExistSuit(cards, Spade) {
		suit = Spade
	} else if !ExistSuit(cards, Heart) {
		suit = Heart
	} else if !ExistSuit(cards, Club) {
		suit = Club
	} else if !ExistSuit(cards, Diamond) {
		suit = Diamond
	}
	return suit + rank
}

func ChangeB(cards []uint32) uint32 {
	new := append([]uint32{}, cards...)
	jokerSort(new)

	a := new[0]
	b := new[1]

	suit := Suit(a)
	var rank uint32

	if Rank(a) == Ace && Rank(b) == King {
		rank = Queen
	} else if Rank(a) == Ace && Rank(b) == Queen {
		rank = King
	} else if Rank(a) == Deuce && Rank(b) == Trey {
		rank = Ace
	} else if Rank(a)+1 == Rank(b) {
		rank = Rank(b) + 1
		if rank == 0x0e {
			rank = Ace
		}
	} else if Rank(a)+2 == Rank(b) {
		rank = Rank(a) + 1
	}
	return suit + rank
}

func ChangeC(cards []uint32) uint32 {
	new := append([]uint32{}, cards...)
	jokerSort(new)

	a := new[0]
	b := new[1]

	suit := Spade
	var rank uint32

	if Rank(a) == Ace && Rank(b) == King {
		rank = Queen
	} else if Rank(a) == Ace && Rank(b) == Queen {
		rank = King
	} else if Rank(a) == Deuce && Rank(b) == Trey {
		rank = Ace
	} else if Rank(a)+1 == Rank(b) {
		rank = Rank(b) + 1
		if rank == 0x0e {
			rank = Ace
		}
	} else if Rank(a)+2 == Rank(b) {
		rank = Rank(a) + 1
	}
	return suit + rank
}

func ChangeD(cards []uint32) uint32 {
	suit := Suit(cards[0])
	var rank uint32
	if !ExistRank(cards, Ace) {
		rank = Ace
	} else {
		for i := King; i > Ace; i-- {
			if !ExistRank(cards, i) {
				rank = i
				break
			}
		}
	}
	return suit + rank
}

func ChangeE(cards []uint32) uint32 {
	new := append([]uint32{}, cards...)
	jokerSort(new)

	var big uint32

	a := new[0]
	b := new[1]

	if Rank(a) == Ace {
		big = a
	} else {
		big = b
	}

	rank := Rank(big)
	var suit uint32

	if !ExistSuit([]uint32{big}, Spade) {
		suit = Spade
	} else if !ExistSuit([]uint32{big}, Heart) {
		suit = Heart
	} else if !ExistSuit([]uint32{big}, Club) {
		suit = Club
	} else if !ExistSuit([]uint32{big}, Diamond) {
		suit = Diamond
	}
	return suit + rank
}

func GetJokerDuiZiCard(current_cards []uint32) (cards []uint32, remain_cards []uint32) {
	dui := [][]uint32{}

	l := len(current_cards)
	for i := 0; i < l; i++ {
		for j := i + 1; j < l; j++ {
			c := []uint32{current_cards[i], current_cards[j]}
			if IsE(c) {
				c = append(c, ChangeE(c))
				dui = append(dui, c)
			}
		}
	}

	cards = dui[utils.RandIntN(len(dui))]

	for _, v := range current_cards {
		if v != cards[0] && v != cards[1] {
			remain_cards = append(remain_cards, v)
		}
	}
	return
}

func GetJokerMaxCard(max []uint32, current_cards []uint32) (bigger []uint32, remain_cards []uint32) {
	big := [][]uint32{}

	l := len(current_cards)
	for i := 0; i < l; i++ {
		for j := i + 1; j < l; j++ {
			c := []uint32{current_cards[i], current_cards[j]}
			if IsA(c) {
				c = append(c, ChangeA(c))
				if JokerCompare(c, max) {
					big = append(big, c)
				}
			} else if IsB(c) {
				c = append(c, ChangeB(c))
				if JokerCompare(c, max) {
					big = append(big, c)
				}
			} else if IsC(c) {
				c = append(c, ChangeC(c))
				if JokerCompare(c, max) {
					big = append(big, c)
				}
			} else if IsD(c) {
				c = append(c, ChangeD(c))
				if JokerCompare(c, max) {
					big = append(big, c)
				}
			} else if IsE(c) {
				c = append(c, ChangeE(c))
				if JokerCompare(c, max) {
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
		if v != bigger[0] && v != bigger[1] { //只有前两张牌是真牌，第三张牌是变出来的
			remain_cards = append(remain_cards, v)
		}
	}

	return
}

func GetJokerCard(cardtype uint32, current_cards []uint32) (cards []uint32, remain_cards []uint32) {
	l := len(current_cards)

	BaoZi_List := [][]uint32{}
	BaoZi1_List := [][]uint32{} //aaa
	BaoZi2_List := [][]uint32{} //kkk-888
	BaoZi3_List := [][]uint32{} //777-222
	TongHuaShun_List := [][]uint32{}
	ShunZi_List := [][]uint32{}
	TongHua_List := [][]uint32{}
	DuiZi_List := [][]uint32{}
	// GaoPai_List := [][]uint32{}

	for i := 0; i < l; i++ {
		for j := i + 1; j < l; j++ {
			c := []uint32{current_cards[i], current_cards[j]}
			if IsA(c) {
				BaoZi_List = append(BaoZi_List, c)
				//aaa kkk-888 777-222拆分
				rank := Rank(c[0])
				if rank == Ace {
					BaoZi1_List = append(BaoZi1_List, c)
				} else if rank >= Eight && rank <= King {
					BaoZi2_List = append(BaoZi2_List, c)
				} else if rank >= Deuce && rank <= Seven {
					BaoZi3_List = append(BaoZi3_List, c)
				}
			} else if IsB(c) {
				TongHuaShun_List = append(TongHuaShun_List, c)
			} else if IsC(c) {
				ShunZi_List = append(ShunZi_List, c)
			} else if IsD(c) {
				TongHua_List = append(TongHua_List, c)
			} else if IsE(c) {
				DuiZi_List = append(DuiZi_List, c)
			}
		}
	}

	switch cardtype {
	case BaoZi1:
		cards = BaoZi1_List[utils.RandIntN(len(BaoZi1_List))]
	case BaoZi2:
		cards = BaoZi2_List[utils.RandIntN(len(BaoZi2_List))]
	case BaoZi3:
		cards = BaoZi3_List[utils.RandIntN(len(BaoZi3_List))]
	case BaoZi:
		cards = BaoZi_List[utils.RandIntN(len(BaoZi_List))]
	case TongHuaShun:
		cards = TongHuaShun_List[utils.RandIntN(len(TongHuaShun_List))]
	case ShunZi:
		cards = ShunZi_List[utils.RandIntN(len(ShunZi_List))]
	case TongHua:
		cards = TongHua_List[utils.RandIntN(len(TongHua_List))]
	case DuiZi:
		cards = DuiZi_List[utils.RandIntN(len(DuiZi_List))]
	}

	for _, v := range current_cards {
		if v != cards[0] && v != cards[1] {
			remain_cards = append(remain_cards, v)
		}
	}

	switch cardtype {
	case BaoZi, BaoZi1, BaoZi2, BaoZi3:
		card := ChangeA(cards)
		cards = append(cards, card)
	case TongHuaShun:
		card := ChangeB(cards)
		cards = append(cards, card)
	case ShunZi:
		card := ChangeC(cards)
		cards = append(cards, card)
	case TongHua:
		card := ChangeD(cards)
		cards = append(cards, card)
	case DuiZi:
		card := ChangeE(cards)
		cards = append(cards, card)
	}

	return
}

func JokerCompare(a, b []uint32) bool {
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
		return jokerCmpBaoZi(a1, b1)
	case TongHuaShun:
		return jokerCmpTongHuaShun(a1, b1)
	case ShunZi:
		return jokerCmpShunZi(a1, b1)
	case TongHua:
		return jokerCmpTongHua(a1, b1)
	case DuiZi:
		return jokerCmpDuiZi(a1, b1)
	case GaoPai:
		return jokerCmpGaoPai(a1, b1)
	default:
		return false
	}
}

// 比较豹子
func jokerCmpBaoZi(a, b []hands) bool {
	if a[0].Rank == b[0].Rank { //点数相同
		for i := 0; i < 3; i++ {
			if a[i].Suit == b[i].Suit {
				continue
			}
			return a[i].Suit > b[i].Suit
		}
		return a[2].Suit > b[2].Suit
	} else { //点数不同
		if a[0].Rank == Ace {
			return true
		}
		if b[0].Rank == Ace {
			return false
		}
		return a[0].Rank > b[0].Rank
	}
}

// 比较同花顺
func jokerCmpTongHuaShun(a, b []hands) bool {
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
		for i := 0; i < 3; i++ {
			if a[i].Suit == b[i].Suit {
				continue
			}
			return a[i].Suit > b[i].Suit
		}
		return a[2].Suit > b[2].Suit
	} else {
		return p1 > p2
	}
}

// 比较顺子
func jokerCmpShunZi(a, b []hands) bool {
	return jokerCmpTongHuaShun(a, b)
}

// 比较同花
func jokerCmpTongHua(a, b []hands) bool {
	//先比点数
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
	//再比花色
	for i := 0; i < 3; i++ {
		if a[i].Suit == b[i].Suit {
			continue
		}
		return a[i].Suit > b[i].Suit
	}
	return a[2].Suit > b[2].Suit
}

// 获取对子点数、单牌点数
func jokerDuiZiPoint(hs []hands) (uint32, uint32) {
	if hs[0].Rank == hs[1].Rank {
		return hs[1].Rank, hs[2].Rank
	}
	if hs[1].Rank == hs[2].Rank {
		return hs[1].Rank, hs[0].Rank
	}
	return 0, 0
}

// 比较对子
func jokerCmpDuiZi(a, b []hands) bool {
	a1, a2 := jokerDuiZiPoint(a)
	b1, b2 := jokerDuiZiPoint(b)

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

	for i := 0; i < 3; i++ {
		if a[i].Suit == b[i].Suit {
			continue
		}
		return a[i].Suit > b[i].Suit
	}
	return a[2].Suit > b[2].Suit
}

// 比较高牌
func jokerCmpGaoPai(a, b []hands) bool {
	return jokerCmpTongHua(a, b)
}
