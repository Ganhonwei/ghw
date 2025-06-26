package algo

import (
	"fmt"
	"goserver/pkg/utils"
	"sort"
	"strconv"
)

var RMCARDS = []uint32{
	0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, //方块
	0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, //梅花
	0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, //红桃
	0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, //黑桃
}

var RMJOKERS = []uint32{0x51, 0x52} //小王大王

// 比较两张牌
func RMCompare(a, b uint32) bool {
	rank1 := Rank(a)
	suit1 := Suit(a)

	rank2 := Rank(b)
	suit2 := Suit(b)

	if rank1 != rank2 { //点数不同
		if rank1 == Ace {
			return true
		}
		if rank2 == Ace {
			return false
		}

		return rank1 > rank2 //比点数
	} else { //点数相同
		return suit1 > suit2 //比花色
	}
}

// 排序手牌
func Sort(cards []uint32) {
	sort.Slice(cards, func(i, j int) bool {
		return cards[i] < cards[j]
	})
}

// 获取牌型
func getCardsType(cards []uint32, laizi uint32) int {
	if IsPureTonghuashun(cards) {
		return 0
	}
	if IsTonghuashun(cards, laizi) {
		return 1
	}
	if IsBaozi(cards, laizi) {
		return 2
	}
	return 3
}

// 排序摆牌
func SortCards2(cards [][]uint32, laizi uint32) {
	sort.Slice(cards, func(i, j int) bool {
		carda := cards[i]
		cardb := cards[j]
		typea := getCardsType(carda, laizi)
		typeb := getCardsType(cardb, laizi)
		return typea < typeb
	})
}

// 比较两幅牌是否完全相同
func IsSame(a, b []uint32) bool {

	newa := append([]uint32{}, a...)
	newb := append([]uint32{}, b...)

	Sort(newa)
	Sort(newb)

	if len(newa) != len(newb) {
		return false
	}

	for k, v := range newa {
		if v != newb[k] {
			return false
		}
	}
	return true
}

func _IsLian(cards []uint32) bool {
	for i := 0; i < len(cards)-1; i++ {
		current := cards[i]
		next := cards[i+1]
		if current+1 != next {
			return false
		}
	}
	return true
}

// 判断牌是否相连
func IsLian(cards []uint32) bool {

	new := append([]uint32{}, cards...)
	Sort(new)

	ret := _IsLian(new)
	if new[0] == 1 && !ret {
		new[0] = 0x0e
		Sort(new)
		ret = _IsLian(new)
	}
	return ret
}

func _NeedLaizi(cards []uint32) int {
	count := 0

	for i := 0; i < len(cards)-1; i++ {
		current := cards[i]
		next := cards[i+1]
		if current+1 != next {
			count += int(next - current - 1)
		}
	}
	return count
}

// 最少需要多少癞子
func NeedLaizi(cards []uint32) int {
	new := append([]uint32{}, cards...)
	Sort(new)

	count := _NeedLaizi(new)

	if new[0] == 1 {
		new[0] = 0x0e
		Sort(new)
		count2 := _NeedLaizi(new)
		if count2 < count {
			count = count2
		}
	}
	return count
}

// 是否非不同花色
func IsNotSameSuit(cards []uint32) bool {
	new := append([]uint32{}, cards...)
	Sort(new)

	for i := 0; i < len(new)-1; i++ {
		current := new[i]
		next := new[i+1]
		if Suit(current) == Suit(next) {
			return false
		}
	}
	return true
}

// 是否为相同牌值
func IsSameRank(cards []uint32) bool {
	for i := 0; i < len(cards)-1; i++ {
		current := cards[i]
		next := cards[i+1]
		if Rank(current) != Rank(next) {
			return false
		}
	}
	return true
}

// 是否同花色
func IsSameSuit(cards []uint32) bool {
	for i := 0; i < len(cards)-1; i++ {
		current := cards[i]
		next := cards[i+1]
		if Suit(current) != Suit(next) {
			return false
		}
	}
	return true
}

// 获取牌值
func GetRanks(cards []uint32) (new []uint32) {
	for _, v := range cards {
		new = append(new, Rank(v))
	}
	return
}

// 是否是癞子
func IsLaizi(card, laizi uint32) bool {
	if card == 0x51 || card == 0x52 {
		return true
	}
	return Rank(card) == Rank(laizi)
}

// 是否是纯同花顺(不能带癞子)
func IsPureTonghuashun(cards []uint32) bool {
	if len(cards) < 3 || len(cards) > 13 { //3-13张
		return false
	}
	if !IsSameSuit(cards) { //判断是否同花色
		return false
	}
	ranks := GetRanks(cards) //去除花色

	ret := IsLian(ranks)
	return ret
}

// 是否是同花顺(可以带癞子)
func IsTonghuashun(cards []uint32, laizi uint32) bool {
	if len(cards) < 3 || len(cards) > 13 {
		return false
	}

	var laizis []uint32    //癞子
	var feilaizis []uint32 //非癞子

	for _, v := range cards {
		if IsLaizi(v, laizi) {
			laizis = append(laizis, v)
		} else {
			feilaizis = append(feilaizis, v)
		}
	}

	if len(feilaizis) == 0 || len(feilaizis) == 1 {
		return true
	}
	if !IsSameSuit(feilaizis) {
		return false
	}
	ranks := GetRanks(feilaizis)

	lian := IsLian(ranks)
	if lian {
		return true
	}
	need := NeedLaizi(ranks)
	if len(laizis) >= need {
		return true
	} else {
		return false
	}
}

// 是否纯豹子
func IsPureBaozi(cards []uint32) bool {
	if len(cards) != 3 && len(cards) != 4 {
		return false
	}
	if !IsSameRank(cards) {
		return false
	}
	if !IsNotSameSuit(cards) {
		return false
	}
	return true
}

// 是否是豹子
func IsBaozi(cards []uint32, laizi uint32) bool {
	if len(cards) != 3 && len(cards) != 4 {
		return false
	}

	var laizis []uint32    //癞子
	var feilaizis []uint32 //非癞子

	for _, v := range cards {
		if IsLaizi(v, laizi) {
			laizis = append(laizis, v)
		} else {
			feilaizis = append(feilaizis, v)
		}
	}

	if len(feilaizis) == 0 || len(feilaizis) == 1 {
		return true
	}
	if !IsSameRank(feilaizis) {
		return false
	}
	if !IsNotSameSuit(feilaizis) {
		return false
	}
	return true
}

func getTotalScore(cards []uint32, laizi uint32) uint32 {
	var total uint32 = 0
	for _, v := range cards {
		if v == 0x51 || v == 0x52 || Rank(v) == Rank(laizi) { //癞子不算分
			continue
		}
		rank := Rank(v)
		switch rank {
		case Ace, Jack, Queen, King: //A,J,Q,K 10分
			total += 10
		default:
			total += rank
		}
	}

	if total > 80 {
		total = 80
	}

	return total
}

func CalcScore(cards [][]uint32, wildCard uint32) int64 {

	SortCards2(cards, wildCard)

	if len(cards) == 1 && len(cards[0]) == 13 {
		if !IsPureTonghuashun(cards[0]) {
			return 80
		} else {
			return 0
		}
	}

	found1 := false
	found2 := false
	for _, v := range cards {
		if len(v) == 0 {
			continue
		}

		if !found1 {
			if IsPureTonghuashun(v) {
				found1 = true
				continue
			}
		}

		if found1 && !found2 {
			if IsPureTonghuashun(v) || IsTonghuashun(v, wildCard) {
				found2 = true
				continue
			}
		}

	}

	if !found1 {
		return 80
	}

	left_card := []uint32{}

	for _, v := range cards {
		if len(v) == 0 {
			continue
		}

		if found1 && !found2 {
			if !IsPureTonghuashun(v) {
				left_card = append(left_card, v...)
			}
		} else if found1 && found2 {
			if !IsPureTonghuashun(v) && !IsTonghuashun(v, wildCard) && !IsBaozi(v, wildCard) {
				left_card = append(left_card, v...)
			}
		}
	}

	return int64(getTotalScore(left_card, wildCard))

	// if len(cards) < 2 { //最少2组
	// 	return 80
	// }
	// if !algo.IsPureTonghuashun(cards[0]) {
	// 	return 80
	// }

	// if !algo.IsPureTonghuashun(cards[1]) && !algo.IsTonghuashun(cards[1], t.WildCard) {
	// 	left_cards := []uint32{}
	// 	left_cards = append(left_cards, cards[1]...)
	// 	for k, v := range cards {
	// 		if k == 0 || k == 1 {
	// 			continue
	// 		}
	// 		left_cards = append(left_cards, v...)
	// 	}
	// 	return int64(t.getTotalScore(left_cards, t.WildCard))
	// }

	// left_cards := []uint32{}
	// for k, v := range cards {
	// 	if k == 0 || k == 1 {
	// 		continue
	// 	}
	// 	if algo.IsPureTonghuashun(v) || algo.IsTonghuashun(v, t.WildCard) || algo.IsBaozi(v, t.WildCard) {
	// 		continue
	// 	}
	// 	left_cards = append(left_cards, v...)
	// }
	// return int64(t.getTotalScore(left_cards, t.WildCard))
}

// 是否胡牌
func IsFinish(cards [][]uint32, laizi uint32) bool {

	//判断是否是一条龙(纯顺子)
	if len(cards) == 1 && len(cards[0]) == 13 {
		if !IsPureTonghuashun(cards[0]) {
			return false
		} else {
			return true
		}
	}

	//判断普通情况
	//必要条件：1.一个纯顺子2.一个顺子（纯顺子或者假顺子）3.其他牌都成组（顺子或者豹子）
	found1 := false //找纯顺子
	found2 := false //找假顺子
	found3 := false //找杂牌

	for _, v := range cards {
		if len(v) == 0 { //跳过空数组
			continue
		}

		if IsPureTonghuashun(v) {
			if !found1 { //先找纯顺子
				found1 = true
			} else {
				found2 = true
			}
			continue
		}
		if IsTonghuashun(v, laizi) {
			found2 = true
			continue
		}
		if !IsBaozi(v, laizi) {
			found3 = true
			break
		}
		// if !found1 {
		// 	if IsPureTonghuashun(v) {
		// 		found1 = true
		// 		continue
		// 	}
		// }

		// //再找顺子
		// if found1 && !found2 {
		// 	if IsPureTonghuashun(v) || IsTonghuashun(v, laizi) {
		// 		found2 = true
		// 		continue
		// 	}
		// }

		//既不是纯顺子，也不是假顺子，也不是豹子，杂牌
		// if !IsPureTonghuashun(v) && !IsTonghuashun(v, laizi) && !IsBaozi(v, laizi) {
		// 	found3 = true
		// 	continue
		// }
	}

	if found3 { //有杂牌，没胡
		return false
	}

	if found1 && found2 {
		return true
	} else {
		return false
	}

	// if len(cards) < 3 {
	// 	return false
	// }
	// first := cards[0]
	// second := cards[1]

	// if !IsPureTonghuashun(first) {
	// 	return false
	// }

	// if !IsPureTonghuashun(second) && !IsTonghuashun(second, laizi) {
	// 	return false
	// }

	// cards = cards[2:]
	// for _, v := range cards {
	// 	if len(v) == 0 {
	// 		continue
	// 	}
	// 	if !IsPureTonghuashun(v) && !IsTonghuashun(v, laizi) && !IsBaozi(v, laizi) {
	// 		return false
	// 	}
	// }
	// return true

}

// 获取牌对应的key值
func getKey(card uint32) string {
	if card == 0x51 {
		return "JOKER1" //小王
	} else if card == 0x52 {
		return "JOKER2" //大王
	}

	rank := Rank(card)
	suit := Suit(card)

	var hua string
	switch suit {
	case Spade:
		hua = "A"
	case Heart:
		hua = "B"
	case Club:
		hua = "C"
	case Diamond:
		hua = "D"
	}

	return fmt.Sprintf("%s%02d", hua, rank)
}

func getValue(key string) uint32 {
	if key == "JOKER1" {
		return 0x51
	} else if key == "JOKER2" {
		return 0x52
	}

	part1 := key[:1]
	part2 := key[1:]

	var suit uint32
	var rank uint32
	switch part1 {
	case "A":
		suit = Spade
	case "B":
		suit = Heart
	case "C":
		suit = Club
	case "D":
		suit = Diamond
	}

	fmt.Sscanf(part2, "%02d", &rank)

	return suit + rank
}

// 将card从map转换到slice
func convertSliceCard(cards map[string]int) []uint32 {
	var ret []uint32

	for k, v := range cards {
		for i := 0; i < v; i++ {
			ret = append(ret, getValue(k))
		}
	}
	Sort(ret)
	return ret
}

// 将card从slice转换倒map
func convertMapCard(cards []uint32) map[string]int {
	var ret map[string]int = make(map[string]int)
	for _, v := range cards {
		key := getKey(v)
		if _, ok := ret[key]; !ok {
			ret[key] = 1
		} else {
			ret[key] += 1
		}
	}
	return ret
}

func RMGetCard(card_type_config []string, current_cards []uint32, laizi uint32) (cards []uint32, remain_cards []uint32) {
	var ret [13]uint32

	mapCards := convertMapCard(current_cards)
	for k, v := range card_type_config {
		if k >= 13 {
			break
		}

		if v == "N00" {
			ret[k] = Random(mapCards)

		} else if v == "N01" {
			ret[k] = RandomExceptJoker(mapCards, laizi)

		} else if v == "N02" {
			ret[k] = RandomJoker(mapCards, laizi)

		} else if v == "A00" || v == "B00" || v == "C00" || v == "D00" {
			ret[k] = RandomABCD(mapCards, v[:1])

		} else if v[:1] == "A" || v[:1] == "B" || v[:1] == "C" || v[:1] == "D" {
			ret[k] = DealABCD(mapCards, v)

		} else if v[:1] == "E" {
			var idx int
			fmt.Sscanf(v[1:], "%02d", &idx)
			if idx >= 13 {
				ret[k] = Random(mapCards)

			} else {
				zhupai := ret[idx]
				ret[k] = DealTonghuaXiaoKao(mapCards, zhupai)

			}
		} else if v[:1] == "F" {
			var idx int
			fmt.Sscanf(v[1:], "%02d", &idx)
			if idx >= 13 {
				ret[k] = Random(mapCards)

			} else {
				zhupai := ret[idx]
				ret[k] = DealTonghuaDaKao(mapCards, zhupai)

			}
		} else if v[:1] == "G" {
			var idx int
			fmt.Sscanf(v[1:], "%02d", &idx)
			if idx >= 13 {
				ret[k] = Random(mapCards)

			} else {
				zhupai := ret[idx]
				ret[k] = DealTonghuaFeiKao(mapCards, zhupai)

			}
		} else if v[:1] == "H" {
			var idx int
			fmt.Sscanf(v[1:], "%02d", &idx)
			if idx >= 13 {
				ret[k] = Random(mapCards)

			} else {
				zhupai := ret[idx]
				ret[k] = DealFeiTonghua(mapCards, zhupai)

			}
		} else {
			ret[k] = Random(mapCards)

		}
	}

	remain_cards = convertSliceCard(mapCards)
	for _, v := range ret {
		cards = append(cards, v)
	}
	// map_cards := conver
	return
}

// 全随机牌-N00
func Random(cards map[string]int) uint32 {
	var keys []string
	for k, v := range cards {
		if v > 0 {
			keys = append(keys, k)
		}
	}
	key := keys[utils.RandIntN(len(keys))]
	cards[key]--
	return getValue(key)
}

// 随机非joker牌-N01
func RandomExceptJoker(cards map[string]int, laizi uint32) uint32 {
	var keys []string
	for k, v := range cards {
		if k == "JOKER1" || k == "JOKER2" || Rank(getValue(k)) == Rank(laizi) {
			continue
		}
		if v > 0 {
			keys = append(keys, k)
		}
	}
	key := keys[utils.RandIntN(len(keys))]
	cards[key]--
	return getValue(key)
}

// 随机非边张非joker牌
func RandomExceptBianZhangCardAndJoker(cards map[string]int, keys []string, laizi uint32) uint32 {
	skipKeys := make(map[string]struct{})
	for _, v := range keys {
		if _, ok := skipKeys[v]; !ok {
			skipKeys[v] = struct{}{}
		}
	}

	var randoms []string
	for k, v := range cards {
		if k == "JOKER1" || k == "JOKER2" || Rank(getValue(k)) == Rank(laizi) {
			continue
		}
		if _, ok := skipKeys[k]; ok {
			continue
		}

		if v > 0 {
			randoms = append(randoms, k)
		}
	}
	random := randoms[utils.RandIntN(len(randoms))]
	cards[random]--
	return getValue(random)
}

// 随机joker牌-N02
func RandomJoker(cards map[string]int, laizi uint32) uint32 {
	var keys []string
	for k, v := range cards {
		if k == "JOKER1" || k == "JOKER2" || Rank(getValue(k)) == Rank(laizi) {
			if v > 0 {
				keys = append(keys, k)
			}
		}
	}
	if len(keys) == 0 {
		return Random(cards)
	}
	key := keys[utils.RandIntN(len(keys))]
	cards[key]--
	return getValue(key)
}

// 随机发非同花ABCD
func RandomABCDExcept(cards map[string]int, part1 string) uint32 {
	var keys []string
	for k, v := range cards {
		if k == "JOKER1" || k == "JOKER2" {
			continue
		}
		if k[:1] == part1 {
			continue
		}
		if v > 0 {
			keys = append(keys, k)
		}
	}
	key := keys[utils.RandIntN(len(keys))]
	cards[key]--
	return getValue(key)
}

// 随机发同花ABCD-A00,B00,C00,D00
func RandomABCD(cards map[string]int, part1 string) uint32 {
	var keys []string
	for k, v := range cards {
		if k == "JOKER1" || k == "JOKER2" {
			continue
		}
		if k[:1] != part1 {
			continue
		}
		if v > 0 {
			keys = append(keys, k)
		}
	}
	key := keys[utils.RandIntN(len(keys))]
	cards[key]--
	return getValue(key)
}

// 发指定同花ABCD-A01-13,B01-13,C01-13,D01-13
func DealABCD(cards map[string]int, key string) uint32 {
	if v, ok := cards[key]; ok && v > 0 {
		cards[key]--
		return getValue(key)
	} else {
		return Random(cards)
	}
}

// 同花小点靠张
func DealTonghuaXiaoKao(cards map[string]int, zhupai uint32) uint32 {
	if zhupai == 0x51 || zhupai == 0x52 || zhupai == 0 {
		return Random(cards)
	}
	key := getKey(zhupai)
	rank := Rank(zhupai)
	rank -= 1
	if rank == 0 { //主牌是A，发K
		rank = 13
	}
	newkey := key[:1] + fmt.Sprintf("%02d", rank)
	return DealABCD(cards, newkey)
}

// 同花大点靠张
func DealTonghuaDaKao(cards map[string]int, zhupai uint32) uint32 {
	if zhupai == 0x51 || zhupai == 0x52 || zhupai == 0 {
		return Random(cards)
	}
	key := getKey(zhupai)
	rank := Rank(zhupai)
	rank += 1
	if rank == 14 { //主牌是K,发A
		rank = 1
	}
	newkey := key[:1] + fmt.Sprintf("%02d", rank)
	return DealABCD(cards, newkey)
}

// 同花非靠
func DealTonghuaFeiKao(cards map[string]int, zhupai uint32) uint32 {
	if zhupai == 0x51 || zhupai == 0x52 || zhupai == 0 {
		return Random(cards)
	}
	key := getKey(zhupai)
	return RandomABCD(cards, key[:1])
}

// 非同花
func DealFeiTonghua(cards map[string]int, zhupai uint32) uint32 {
	if zhupai == 0x51 || zhupai == 0x52 || zhupai == 0 {
		return Random(cards)
	}
	key := getKey(zhupai)
	return RandomABCDExcept(cards, key[:1])
}

// 发指定牌
func DealAssignCard(cards []uint32, assignCard uint32) (remain_cards []uint32, ok bool) {
	remain_cards = cards

	for i, c := range cards {
		ok = c == assignCard
		if ok {
			remain_cards = append(cards[:i], cards[i+1:]...)
			return
		}
	}
	ok = false
	return
}

// 发指定牌之外的牌
func DealExceptCards(cards []uint32, exceptsCards []uint32) (card uint32, remain_cards []uint32, ok bool) {
	excepts := make(map[uint32]bool, len(exceptsCards))
	for _, card := range exceptsCards {
		excepts[card] = true
	}
	for i, c := range cards {
		if !excepts[c] {
			ok = true
			card = c
			remain_cards = append(cards[:i], cards[i+1:]...)
			return
		}
	}
	remain_cards = cards
	return
}

// 发牌堆下一张牌
func DealNextCard(cards []uint32) (card uint32, remain_cards []uint32) {
	card = cards[0]
	remain_cards = cards[1:]
	return
}

// 发随机牌
func DealRandomCard(cards []uint32) (card uint32, remain_cards []uint32) {
	mapCards := convertMapCard(cards)
	card = Random(mapCards)
	remain_cards = convertSliceCard(mapCards)
	return
}

// 发随机joker
func DealRandomJoker(cards []uint32, laizi uint32) (card uint32, remain_cards []uint32) {
	mapCards := convertMapCard(cards)
	card = RandomJoker(mapCards, laizi)
	remain_cards = convertSliceCard(mapCards)
	return
}

// 发随机非joker牌
func DealRandomExceptJoker(cards []uint32, laizi uint32) (card uint32, remain_cards []uint32) {
	mapCards := convertMapCard(cards)
	card = RandomExceptJoker(mapCards, laizi)
	remain_cards = convertSliceCard(mapCards)
	return
}

// 发随机非joker牌边涨牌(包含豹子牌)
func DealRandomExceptBianZhangCardAndJoker(cards []uint32, hand_cards []uint32, laizi uint32) (card uint32, remain_cards []uint32) {
	mapCards := convertMapCard(cards)
	keys1 := calcBianZhangCard(hand_cards)
	keys2 := calcBaoziCard(hand_cards)
	keys3 := calcBianZhangCardWithLaizi(hand_cards, laizi)
	keys4 := calcBaoziCardWithLaizi(hand_cards, laizi)
	mergeKeys := append(append(append(keys1, keys2...), keys3...), keys4...)
	card = RandomExceptBianZhangCardAndJoker(mapCards, mergeKeys, laizi)
	remain_cards = convertSliceCard(mapCards)
	return
}

func calcBaoziCard(hand_cards []uint32) []string {
	isNeed := func(hand_cards []uint32, card uint32) bool {
		new := append([]uint32{}, hand_cards...)
		new = append(new, card)

		mapHandCards := convertMapCard(new)

		rank := Rank(card)

		count := 0
		//检查黑桃
		keya := fmt.Sprintf("%s%02d", "A", rank)
		if v, ok := mapHandCards[keya]; ok && v > 0 {
			count++
		}

		//检查红桃
		keyb := fmt.Sprintf("%s%02d", "B", rank)
		if v, ok := mapHandCards[keyb]; ok && v > 0 {
			count++
		}

		//检查梅花
		keyc := fmt.Sprintf("%s%02d", "C", rank)
		if v, ok := mapHandCards[keyc]; ok && v > 0 {
			count++
		}

		//检查方块
		keyd := fmt.Sprintf("%s%02d", "D", rank)
		if v, ok := mapHandCards[keyd]; ok && v > 0 {
			count++
		}

		//如果总数大于2则是豹子
		return count > 2
	}

	var keys []string
	for i := Ace; i <= King; i++ {
		for j := Spade; j <= Diamond; j++ {
			need := isNeed(hand_cards, j+i)
			if need {
				key := getKey(j + i)
				keys = append(keys, key)
			}
		}
	}
	return keys
}

func calcBaoziCardWithLaizi(hand_cards []uint32, laizi uint32) []string {
	isNeed := func(hand_cards []uint32, card uint32) bool {
		new := append([]uint32{}, hand_cards...)
		new = append(new, card)

		mapHandCards := convertMapCard(new)

		rank := Rank(card)

		count := 0
		//检查黑桃
		keya := fmt.Sprintf("%s%02d", "A", rank)
		if v, ok := mapHandCards[keya]; ok && v > 0 {
			count++
		}

		//检查红桃
		keyb := fmt.Sprintf("%s%02d", "B", rank)
		if v, ok := mapHandCards[keyb]; ok && v > 0 {
			count++
		}

		//检查梅花
		keyc := fmt.Sprintf("%s%02d", "C", rank)
		if v, ok := mapHandCards[keyc]; ok && v > 0 {
			count++
		}

		//检查方块
		keyd := fmt.Sprintf("%s%02d", "D", rank)
		if v, ok := mapHandCards[keyd]; ok && v > 0 {
			count++
		}

		//如果总数大于1则是豹子
		return count > 1
	}

	var keys []string

	if !isLaiziExist(hand_cards, laizi) {
		return keys
	}

	for i := Ace; i <= King; i++ {
		for j := Spade; j <= Diamond; j++ {
			need := isNeed(hand_cards, j+i)
			if need {
				key := getKey(j + i)
				keys = append(keys, key)
			}
		}
	}
	return keys
}

func calcBianZhangCard(hand_cards []uint32) []string {
	var keys []string

	//计算每张牌的边张
	for _, v := range hand_cards {
		if v == 0x51 || v == 0x52 || v == 0 {
			continue
		}
		key := getKey(v)
		//大点靠张
		rank_xiaodian := Rank(v)
		rank_xiaodian -= 1
		if rank_xiaodian == 0 {
			rank_xiaodian = 13
		}
		xiaodian_newkey := key[:1] + fmt.Sprintf("%02d", rank_xiaodian)
		keys = append(keys, xiaodian_newkey)

		//小点靠张
		rank_dadian := Rank(v)
		rank_dadian += 1
		if rank_dadian == 14 {
			rank_dadian = 1
		}
		dadian_newkey := key[:1] + fmt.Sprintf("%02d", rank_dadian)
		keys = append(keys, dadian_newkey)
	}

	var removed_keys []string
	mapHandCards := convertMapCard(hand_cards)

	//去除已经有得边张牌
	for _, key := range keys {
		if v, ok := mapHandCards[key]; ok && v > 0 {
			continue
		}
		removed_keys = append(removed_keys, key)
	}

	return removed_keys
}

func calcBianZhangCardWithLaizi(hand_cards []uint32, laizi uint32) []string {
	var keys []string

	if !isLaiziExist(hand_cards, laizi) {
		return keys
	}

	//计算每张牌的边张
	for _, v := range hand_cards {
		if v == 0x51 || v == 0x52 || v == 0 {
			continue
		}

		key := getKey(v)
		//大点靠涨
		rank_xiaodian := int32(Rank(v))
		rank_xiaodian -= 2
		if rank_xiaodian == 0 { //2的情况下不靠
			continue
		}
		if rank_xiaodian == -1 { //A的情况下靠张Q
			rank_xiaodian = 12
		}
		xiaodian_newkey := key[:1] + fmt.Sprintf("%02d", rank_xiaodian)
		keys = append(keys, xiaodian_newkey)

		//小点靠张
		rank_dadian := int32(Rank(v))
		rank_dadian += 2
		if rank_dadian == 14 { //q的情况下靠A
			rank_dadian = 1
		}
		if rank_dadian == 15 { //k的情况下不靠
			continue
		}
		dadian_newkey := key[:1] + fmt.Sprintf("%02d", rank_dadian)
		keys = append(keys, dadian_newkey)
	}

	var removed_keys []string
	mapHandCards := convertMapCard(hand_cards)

	//去除已经有得边张牌
	for _, key := range keys {
		if v, ok := mapHandCards[key]; ok && v > 0 {
			continue
		}
		removed_keys = append(removed_keys, key)
	}

	return removed_keys

}

func isLaiziExist(hand_cards []uint32, laizi uint32) bool {
	//遍历手牌
	for _, v := range hand_cards {
		if IsLaizi(v, laizi) {
			return true
		}
	}
	return false
}

func IsNeedCard(hand_cards []uint32, card uint32, laizi uint32) bool {
	need1 := IsNeedBianZhangCard(hand_cards, card)
	need2 := IsNeedBaoZiCard(hand_cards, card)
	need3 := IsNeedBianZhangCardWithLaizi(hand_cards, card, laizi)
	need4 := IsNeedBaoZiCardWithLaizi(hand_cards, card, laizi)
	return need1 || need2 || need3 || need4
}

func IsNeedBianZhangCardWithLaizi(hand_cards []uint32, card uint32, laizi uint32) bool {
	keys := calcBianZhangCardWithLaizi(hand_cards, laizi)
	key := getKey(card)

	for _, v := range keys {
		if v == key {
			return true
		}
	}
	return false
}

// 是否是需要的边张
func IsNeedBianZhangCard(hand_cards []uint32, card uint32) bool {
	keys := calcBianZhangCard(hand_cards)
	key := getKey(card)

	for _, v := range keys {
		if v == key {
			return true
		}
	}
	return false
}

func IsNeedBaoZiCardWithLaizi(hand_cards []uint32, card uint32, laizi uint32) bool {
	keys := calcBaoziCardWithLaizi(hand_cards, laizi)
	key := getKey(card)

	for _, v := range keys {
		if v == key {
			return true
		}
	}
	return false
}

// 是否是需要的豹子牌
func IsNeedBaoZiCard(hand_cards []uint32, card uint32) bool {
	keys := calcBaoziCard(hand_cards)
	key := getKey(card)

	for _, v := range keys {
		if v == key {
			return true
		}
	}
	return false
}

// 发随机边张牌
func DealRandomBianZhangCard(cards []uint32, hand_cards []uint32) (card uint32, remain_cards []uint32) {
	mapCards := convertMapCard(cards)

	removed_keys := calcBianZhangCard(hand_cards)

	if len(removed_keys) == 0 {
		card = Random(mapCards)
	} else {
		key := removed_keys[utils.RandIntN(len(removed_keys))]
		card = DealABCD(mapCards, key)
	}

	remain_cards = convertSliceCard(mapCards)
	return
}

// 移除牌
func removeCards(cards []uint32, remove_cards []uint32) (left_cards []uint32, success bool) {
	mapCards := convertMapCard(cards)
	mapRemoveCards := convertMapCard(remove_cards)

	success = true
	for k, needNum := range mapRemoveCards {
		if haveNum, ok := mapCards[k]; ok {
			if haveNum >= needNum {
				mapCards[k] -= needNum
			} else {
				success = false
			}
		} else {
			success = false
		}
	}
	left_cards = convertSliceCard(mapCards)
	return
}

func RemoveCard(cards []uint32, card uint32) (left_cards []uint32) {
	left_cards, _ = removeCards(cards, []uint32{card})
	return
}

func RemoveCards(cards []uint32, remove_cards []uint32) (left_cards []uint32, success bool) {
	return removeCards(cards, remove_cards)
}

// 移除大小王（不包括癞子）
func removeJoker(cards []uint32) (ret_cards []uint32) {
	mapCards := convertMapCard(cards)
	if v, ok := mapCards["JOKER1"]; ok && v > 0 {
		mapCards["JOKER1"] = 0
	}
	if v, ok := mapCards["JOKER2"]; ok && v > 0 {
		mapCards["JOKER2"] = 0
	}
	ret_cards = convertSliceCard(mapCards)
	return
}

// 移除万能牌
func removeWildCard(cards []uint32, laizi uint32) (ret_cards []uint32) {
	mapCards := convertMapCard(cards)

	rank := Rank(laizi)
	keya := fmt.Sprintf("%s%02d", "A", rank)
	keyb := fmt.Sprintf("%s%02d", "B", rank)
	keyc := fmt.Sprintf("%s%02d", "C", rank)
	keyd := fmt.Sprintf("%s%02d", "D", rank)
	if v, ok := mapCards[keya]; ok && v > 0 {
		mapCards[keya] = 0
	}
	if v, ok := mapCards[keyb]; ok && v > 0 {
		mapCards[keyb] = 0
	}
	if v, ok := mapCards[keyc]; ok && v > 0 {
		mapCards[keyc] = 0
	}
	if v, ok := mapCards[keyd]; ok && v > 0 {
		mapCards[keyd] = 0
	}
	ret_cards = convertSliceCard(mapCards)
	return
}

// 按花色分组
func splitABCD(cards []uint32) (a, b, c, d []uint32) {
	for _, v := range cards {
		switch Suit(v) {
		case Spade:
			a = append(a, v)
		case Heart:
			b = append(b, v)
		case Club:
			c = append(c, v)
		case Diamond:
			d = append(d, v)
		}
	}
	return
}

func _findPureTonghuashun(cards []uint32) (group []uint32) {
	if len(cards) < 3 {
		return
	}
	for i := 0; i < len(cards); i++ {
		for j := i + 1; j < len(cards); j++ {
			for k := j + 1; k < len(cards); k++ {
				a := cards[i]
				b := cards[j]
				c := cards[k]
				// glog.Debugf("%d,%d,%d", a, b, c)
				ranks := GetRanks([]uint32{a, b, c})
				if IsLian(ranks) {
					group = append(group, a, b, c)
					return
				}
			}
		}
	}
	// over := false
	// for i := 0; i <= len(cards)-2 && !over; i++ {
	// 	a_idx := i
	// 	b_idx := i + 1
	// 	c_idx := i + 2
	// 	if c_idx == len(cards) {
	// 		c_idx = 0
	// 		over = true
	// 	}

	// 	a := cards[a_idx]
	// 	b := cards[b_idx]
	// 	c := cards[c_idx]

	// 	if a+1 == b && b+1 == c {
	// 		group = append(group, a, b, c)
	// 		break
	// 	} else if Rank(a) == Queen && Rank(b) == King && Rank(c) == Ace {
	// 		group = append(group, c, a, b)
	// 		break
	// 	}
	// }
	return
}

// 找纯同花顺3张（不带癞子，万能牌当普通牌）
func findPureTonghuashun(cards []uint32, laizi uint32) (group []uint32, left_cards []uint32) {
	if len(cards) < 3 {
		left_cards = cards
		return
	}

	//先移除癞子找（大小王和万能牌）
	removeLaizi := removeWildCard(removeJoker(cards), laizi)
	a, b, c, d := splitABCD(removeLaizi)
	abcd := [][]uint32{a, b, c, d}

	for _, v := range abcd {
		group = _findPureTonghuashun(v)
		if len(group) > 0 {
			var ok bool
			left_cards, ok = removeCards(cards, group)
			if ok {
				return
			}
		}
	}

	//只移除大小王找（万能牌当普通牌）
	removeJokerCards := removeJoker(cards) //移除大小王
	a1, b1, c1, d1 := splitABCD(removeJokerCards)
	abcd1 := [][]uint32{a1, b1, c1, d1}

	for _, v := range abcd1 {
		group = _findPureTonghuashun(v)
		if len(group) > 0 {
			var ok bool
			left_cards, ok = removeCards(cards, group)
			if ok {
				return
			}
		}
	}

	left_cards = cards
	return
}

func _findTonghuashun(cards []uint32) (group []uint32) {
	if len(cards) < 2 {
		return
	}
	over := false
	for i := 0; i <= len(cards)-1 && !over; i++ {
		a_idx := i
		b_idx := i + 1
		if b_idx == len(cards) {
			b_idx = 0
			over = true
		}
		a := cards[a_idx]
		b := cards[b_idx]

		if a+1 == b { //连着的
			group = append(group, a, b)
			break
		} else if Rank(a) == King && Rank(b) == Ace {
			group = append(group, b, a)
			break
		} else if a+2 == b { //隔一张牌
			group = append(group, a, b)
			break
		} else if Rank(a) == Queen && Rank(b) == Ace {
			group = append(group, b, a)
			break
		}
	}
	return
}

// 找同花顺3张(只带1个癞子)
func findTonghuashun(cards []uint32, laizi uint32) (group []uint32, left_cards []uint32) {
	if len(cards) < 3 {
		left_cards = cards
		return
	}

	var laizis []uint32   //癞子(包括大小王和万能牌)
	var feilaizi []uint32 //非癞子牌

	for _, v := range cards {
		if IsLaizi(v, laizi) {
			laizis = append(laizis, v)
		} else {
			feilaizi = append(feilaizi, v)
		}
	}

	if len(laizis) > 0 {
		a, b, c, d := splitABCD(feilaizi)
		abcd := [][]uint32{a, b, c, d}

		for _, v := range abcd {
			group = _findTonghuashun(v)
			if len(group) > 0 {
				laizi := laizis[utils.RandIntN(len(laizis))]
				group = append(group, laizi)
				var ok bool
				left_cards, ok = removeCards(cards, group)
				if ok {
					return
				}
			}
		}
	}

	left_cards = cards
	return
}

// 找豹子3张(不带癞子或带一个癞子)
func findBaozi(cards []uint32, laizi uint32) (group []uint32, left_cards []uint32) {
	if len(cards) < 3 {
		left_cards = cards
		return
	}

	var laizis []uint32   //癞子(包括大小王和万能牌)
	var feilaizi []uint32 //非癞子牌

	for _, v := range cards {
		if IsLaizi(v, laizi) {
			laizis = append(laizis, v)
		} else {
			feilaizi = append(feilaizi, v)
		}
	}

	mapCard := convertMapCard(feilaizi)

	var temp []uint32
	for i := 1; i <= 13; i++ {
		count := 0
		akey := fmt.Sprintf("A%02d", i)
		bkey := fmt.Sprintf("B%02d", i)
		ckey := fmt.Sprintf("C%02d", i)
		dkey := fmt.Sprintf("D%02d", i)

		if v, ok := mapCard[akey]; ok && v > 0 {
			count++
			temp = append(temp, getValue(akey))
		}
		if v, ok := mapCard[bkey]; ok && v > 0 {
			count++
			temp = append(temp, getValue(bkey))
		}
		if v, ok := mapCard[ckey]; ok && v > 0 {
			count++
			temp = append(temp, getValue(ckey))
		}
		if v, ok := mapCard[dkey]; ok && v > 0 {
			count++
			temp = append(temp, getValue(dkey))
		}
		if count < 2 {
			temp = []uint32{}
		} else {
			break
		}
	}

	if len(temp) >= 3 { //3张不用带癞子
		group = append(group, temp[:3]...)
		var ok bool
		left_cards, ok = removeCards(cards, group)
		if ok {
			return
		}
	} else if len(temp) == 2 && len(laizis) > 0 { //2张带一个癞子
		group = append(group, temp...)
		laizi := laizis[utils.RandIntN(len(laizis))]
		group = append(group, laizi)
		var ok bool
		left_cards, ok = removeCards(cards, group)
		if ok {
			return
		}
	}

	left_cards = cards
	return
}

func _sortCards(cards *[]uint32, laizi uint32, result *[][]uint32) {
	if len(*cards) < 3 {
		return
	}

	find := false
	var group []uint32
	//先找纯同花顺
	if !find {
		group, *cards = findPureTonghuashun(*cards, laizi)
		if len(group) > 0 {
			*result = append(*result, group)
			find = true
		}
	}

	//再找同花顺
	if !find {
		group, *cards = findTonghuashun(*cards, laizi)
		if len(group) > 0 {
			*result = append(*result, group)
			find = true
		}
	}

	// 最后找豹子
	if !find {
		group, *cards = findBaozi(*cards, laizi)
		if len(group) > 0 {
			*result = append(*result, group)
			find = true
		}
	}

	if !find {
		// if len(*cards) > 0 {
		// *result = append(*result, group)
		// _ = result
		// }
		return
	} else {
		// 递归调用
		_sortCards(cards, laizi, result)
	}
}

// 插牌
func AddCard(group []uint32, card uint32, laizi uint32) bool {
	if IsPureTonghuashun(group) {
		group = append(group, card)
		if IsPureTonghuashun(group) {
			return true
		} else {
			return false
		}
	} else if IsTonghuashun(group, laizi) {
		group = append(group, card)
		if IsTonghuashun(group, laizi) {
			return true
		} else {
			return false
		}
	} else if IsBaozi(group, laizi) && len(group) == 3 {
		group = append(group, card)
		if IsBaozi(group, laizi) {
			return true
		} else {
			return false
		}
	}
	return false
}

// 摆牌
func SortCards(cards []uint32, laizi uint32) (groups [][]uint32, left_cards []uint32) {
	_sortCards(&cards, laizi, &groups)

	var temp []uint32
	if len(groups) > 0 && len(cards) > 0 {
		for _, card := range cards {
			for k, group := range groups {
				if AddCard(group, card, laizi) {
					groups[k] = append(groups[k], card)
					temp = append(temp, card)
					break
				}
			}
		}
	}

	left_cards, _ = removeCards(cards, temp)

	// result = append(result, cards)

	return
}

func SortCards3(cards [13]uint32, laizi uint32) [][]uint32 {
	_cards := cards[:]
	g, l := SortCards(_cards, laizi)
	if len(l) != 0 {
		g = append(g, l)
	}
	return g
}

func SortCards4(cards []uint32, laizi uint32) [][]uint32 {
	g, l := SortCards(cards, laizi)
	if len(l) != 0 {
		g = append(g, l)
	}
	return g
}

// 判断是否需要弃牌
func NeedCard(cards []uint32, qi_card uint32, laizi uint32) bool {
	groups, left_cards := SortCards(cards, laizi)

	if len(groups) > 0 {
		for _, v := range groups {
			if AddCard(v, qi_card, laizi) {
				return false
			}
		}
	}

	left_cards = append(left_cards, qi_card)
	if group, _ := SortCards(left_cards, laizi); len(group) > 0 {
		return true
	}

	return false
}

func find4(cards []uint32) (za []uint32, left_cards []uint32) {
	a, b, c, d := splitABCD(cards)
	abcd := [][]uint32{a, b, c, d}

	for _, v := range abcd {
		if len(v) < 2 {
			continue
		}
		over := false
		for i := 0; i <= len(v)-1 && !over; i++ {
			a_idx := i
			b_idx := i + 1
			if b_idx == len(v) {
				b_idx = 0
				over = true
			}
			a := v[a_idx]
			b := v[b_idx]

			if a+1 == b {
				za = append(za, a, b)
				left_cards, _ = removeCards(cards, za)
				return
			} else if Rank(a) == King && Rank(b) == Ace {
				za = append(za, b, a)
				left_cards, _ = removeCards(cards, za)
				return
			}
		}
	}
	left_cards = cards
	return
}

func find3(cards []uint32) (za []uint32, left_cards []uint32) {
	a, b, c, d := splitABCD(cards)
	abcd := [][]uint32{a, b, c, d}

	for _, v := range abcd {
		if len(v) < 2 {
			continue
		}
		over := false
		for i := 0; i <= len(v)-1 && !over; i++ {
			a_idx := i
			b_idx := i + 1
			if b_idx == len(v) {
				b_idx = 0
				over = true
			}
			a := v[a_idx]
			b := v[b_idx]

			if a+2 == b {
				za = append(za, a, b)
				left_cards, _ = removeCards(cards, za)
				return
			} else if Rank(a) == Queen && Rank(b) == Ace {
				za = append(za, b, a)
				left_cards, _ = removeCards(cards, za)
				return
			}
		}
	}
	left_cards = cards
	return
}

func find2(cards []uint32) (za []uint32, left_cards []uint32) {
	mapCard := convertMapCard(cards)

	var temp []uint32
	for i := 1; i <= 13; i++ {
		count := 0
		akey := fmt.Sprintf("A%02d", i)
		bkey := fmt.Sprintf("B%02d", i)
		ckey := fmt.Sprintf("C%02d", i)
		dkey := fmt.Sprintf("D%02d", i)

		if v, ok := mapCard[akey]; ok && v > 0 {
			count++
			temp = append(temp, getValue(akey))
		}
		if v, ok := mapCard[bkey]; ok && v > 0 {
			count++
			temp = append(temp, getValue(bkey))
		}
		if v, ok := mapCard[ckey]; ok && v > 0 {
			count++
			temp = append(temp, getValue(ckey))
		}
		if v, ok := mapCard[dkey]; ok && v > 0 {
			count++
			temp = append(temp, getValue(dkey))
		}
		if count < 2 {
			temp = []uint32{}
		} else {
			break
		}
	}

	if len(temp) >= 2 {
		za = append(za, temp[:2]...)
		left_cards, _ = removeCards(cards, za)
	} else {
		left_cards = cards
	}
	return
}

// 杂牌拆牌 1.纯单牌 2.拆对子 3.拆差中张的顺子 4.拆差边张的顺子
func splitza1234(cards []uint32, laizi uint32) (za4 [][]uint32, za3 [][]uint32, za2 [][]uint32, za1 []uint32) {
	//过滤癞子
	var feilaizis []uint32
	for _, v := range cards {
		if !IsLaizi(v, laizi) {
			feilaizis = append(feilaizis, v)
		}
	}

	cards = feilaizis

	no4 := false
	for !no4 {
		var za []uint32
		za, cards = find4(cards)
		if len(za) > 0 {
			za4 = append(za4, za)
		} else {
			no4 = true
		}
	}

	no3 := false
	for !no3 {
		var za []uint32
		za, cards = find3(cards)
		if len(za) > 0 {
			za3 = append(za3, za)
		} else {
			no3 = true
		}
	}

	no2 := false
	for !no2 {
		var za []uint32
		za, cards = find2(cards)
		if len(za) > 0 {
			za2 = append(za2, za)
		} else {
			no2 = true
		}
	}

	za1 = cards
	return
}

// 拆牌打
func splitCardToDiscard(cards []uint32, laizi uint32) (discard_card uint32) {
	z4, z3, z2, z1 := splitza1234(cards, laizi)
	if len(z1) > 0 {
		discard_card = z1[utils.RandIntN(len(z1))]
	} else if len(z2) > 0 {
		randz2 := z2[utils.RandIntN(len(z2))]
		discard_card = randz2[utils.RandIntN(len(randz2))]
	} else if len(z3) > 0 {
		randz3 := z3[utils.RandIntN(len(z3))]
		discard_card = randz3[utils.RandIntN(len(randz3))]
	} else if len(z4) > 0 {
		randz4 := z4[utils.RandIntN(len(z4))]
		discard_card = randz4[utils.RandIntN(len(randz4))]
	}
	return
}

// 获取一张牌打出
func GetOneCardToDiscard(cards []uint32, laizi uint32) (finish bool, discard_card uint32) {
	groups, left_cards := SortCards(cards, laizi)

	length := len(left_cards)
	if length == 0 {
		if IsFinish(groups, laizi) {
			finish = true
		}
		for i := len(groups) - 1; i >= 0; i-- {
			if len(groups[i]) > 3 {
				discard_card = groups[i][len(groups[i])-1]
				break
			}
		}
	} else if length == 1 {
		if IsFinish(groups, laizi) {
			finish = true
		}
		discard_card = left_cards[0]
	} else {
		discard_card = splitCardToDiscard(left_cards, laizi)
	}

	if !finish && IsLaizi(discard_card, laizi) {
		discard_card = splitCardToDiscard(cards, laizi)
	}

	if discard_card == 0 {
		discard_card = cards[utils.RandIntN(len(cards))]
	}

	return
}

// 获取rm癞子数
func GetRMLaiziNum(cards []uint32, laizi uint32) int {
	count := 0
	for _, v := range cards {
		if IsLaizi(v, laizi) {
			count++
		}
	}
	return count
}

func CardsString(cards []uint32) (ret string) {
	ret += "["
	for _, card := range cards {
		if card == RMJOKERS[0] {
			ret += "JOKER1" + "(" + strconv.Itoa(int(card)) + "),"
		} else if card == RMJOKERS[1] {
			ret += "JOKER2," + strconv.Itoa(int(card))
		} else {
			ret += GetSuitString(card) + GetRankString(card) + "(" + strconv.Itoa(int(card)) + "),"
		}
	}
	ret += "]"
	return
}

func CardGroupsString(groups [][]uint32) (ret string) {
	ret += "{"
	for _, g := range groups {
		ret += "["
		for _, card := range g {
			if card == RMJOKERS[0] {
				ret += "JOKER1,"
			} else if card == RMJOKERS[1] {
				ret += "JOKER2,"
			} else {
				ret += GetSuitString(card) + GetRankString(card) + ","
			}
		}
		ret += "]"
	}
	ret += "}"
	return
}
