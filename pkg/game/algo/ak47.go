package algo

import (
	"goserver/pkg/utils"
)

// 非癞子
var Feilaizi = []uint32{
	0x12, 0x13, 0x15, 0x16, 0x18, 0x19, 0x1a, 0x1b, 0x1c, //方块
	0x22, 0x23, 0x25, 0x26, 0x28, 0x29, 0x2a, 0x2b, 0x2c, //梅花
	0x32, 0x33, 0x35, 0x36, 0x38, 0x39, 0x3a, 0x3b, 0x3c, //红桃
	0x42, 0x43, 0x45, 0x46, 0x48, 0x49, 0x4a, 0x4b, 0x4c, //黑桃
}

// 癞子 A,k,4,7,
var Laizi = []uint32{
	0x11, 0x1d, 0x14, 0x17,
	0x21, 0x2d, 0x24, 0x27,
	0x31, 0x3d, 0x34, 0x37,
	0x41, 0x4d, 0x44, 0x47,
}

// 是否是癞子
func isLaizi(card uint32) bool {
	rank := Rank(card)
	if rank == Ace || rank == King || rank == Four || rank == Seven {
		return true
	} else {
		return false
	}
}

func AK47IsLaizi(card uint32) bool {
	return isLaizi(card)
}

// 洗牌
func Shuffle(cards []uint32) {
	for i := len(cards) - 1; i > 0; i-- {
		j := utils.RandIntN(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}
}

func GetCard0(cardtype uint32, current_cards []uint32) (cards, change_cards, remain_cards []uint32) {
	l := len(current_cards)

	BaoZi_List := [][]uint32{}
	BaoZi1_List := [][]uint32{} //aaa
	BaoZi2_List := [][]uint32{} //kkk-888
	BaoZi3_List := [][]uint32{} //777-222

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
	case GaoPai:
		cards = GaoPai_List[utils.RandIntN(len(GaoPai_List))]
	}

	change_cards = append([]uint32{}, cards...) //没有癞子，原牌和变后的牌是一样的

	for _, v := range current_cards {
		if v != cards[0] && v != cards[1] && v != cards[2] {
			remain_cards = append(remain_cards, v)
		}
	}

	return
}

func GetCard1(wildcard, cardtype uint32, current_cards []uint32) (cards, change_cards, remain_cards []uint32) {
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

	cards = append(cards, wildcard) //癞子放前面

	var feilaizi []uint32 //两张非赖子牌
	switch cardtype {
	case BaoZi1:
		feilaizi = BaoZi1_List[utils.RandIntN(len(BaoZi1_List))]
	case BaoZi2:
		feilaizi = BaoZi2_List[utils.RandIntN(len(BaoZi2_List))]
	case BaoZi3:
		feilaizi = BaoZi3_List[utils.RandIntN(len(BaoZi3_List))]
	case BaoZi:
		feilaizi = BaoZi_List[utils.RandIntN(len(BaoZi_List))]
	case TongHuaShun:
		feilaizi = TongHuaShun_List[utils.RandIntN(len(TongHuaShun_List))]
	case ShunZi:
		feilaizi = ShunZi_List[utils.RandIntN(len(ShunZi_List))]
	case TongHua:
		feilaizi = TongHua_List[utils.RandIntN(len(TongHua_List))]
	case DuiZi:
		feilaizi = DuiZi_List[utils.RandIntN(len(DuiZi_List))]
	}

	for _, v := range current_cards {
		if v != feilaizi[0] && v != feilaizi[1] {
			remain_cards = append(remain_cards, v)
		}
	}

	cards = append(cards, feilaizi...) //再添加两张普通牌

	var change_card uint32 //变牌
	switch cardtype {
	case BaoZi, BaoZi1, BaoZi2, BaoZi3:
		change_card = ChangeA(feilaizi)
	case TongHuaShun:
		change_card = ChangeB(feilaizi)
	case ShunZi:
		change_card = ChangeC(feilaizi)
	case TongHua:
		change_card = ChangeD(feilaizi)
	case DuiZi:
		change_card = ChangeE(feilaizi)
	}

	change_cards = append(change_cards, change_card)
	change_cards = append(change_cards, feilaizi...)

	return
}

func GetSuit(suit uint32) (suita, suitb uint32) {
	if suit == Spade {
		return Heart, Club
	} else if suit == Heart {
		return Spade, Club
	} else if suit == Club {
		return Spade, Heart
	} else if suit == Diamond {
		return Spade, Heart
	}
	return 0, 0
}

func GetCard2(wildcard []uint32, cardtype uint32, current_cards []uint32) (cards, change_cards, remain_cards []uint32) {
	cards = append(cards, wildcard...)                        //两张癞子牌
	card := current_cards[utils.RandIntN(len(current_cards))] //随机一张牌
	cards = append(cards, card)                               //一张普通牌

	for _, v := range current_cards {
		if v != card {
			remain_cards = append(remain_cards, v)
		}
	}

	rank := Rank(card)
	suita, suitb := GetSuit(Suit(card))

	carda := suita + rank
	cardb := suitb + rank

	change_cards = append(change_cards, carda, cardb)
	change_cards = append(change_cards, card)
	return
}

func GetCard3(wildcard []uint32, cardtype uint32, current_cards []uint32) (cards, change_cards, remain_cards []uint32) {
	cards = append(cards, wildcard...) //3张癞子

	remain_cards = current_cards

	carda := Spade + Ace
	cardb := Heart + Ace
	cardc := Club + Ace
	change_cards = append(change_cards, carda, cardb, cardc) //3张A
	return
}

func AK47GetCard(wild_cards []uint32, cardtype uint32, current_cards []uint32) (cards, change_cards, remain_cards []uint32) {
	if len(wild_cards) == 0 {
		return GetCard0(cardtype, current_cards)
	} else if len(wild_cards) == 1 {
		return GetCard1(wild_cards[0], cardtype, current_cards)
	} else if len(wild_cards) == 2 {
		return GetCard2(wild_cards, cardtype, current_cards)
	} else if len(wild_cards) == 3 {
		return GetCard3(wild_cards, cardtype, current_cards)
	}
	return
}

func GetLaiziNum(h []uint32) int {
	count := 0
	for _, v := range h {
		if isLaizi(v) {
			count++
		}
	}
	return count
}

func ak47Compare(a, b []uint32) bool {
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

// 判断两幅牌是否完全一样
func IsSuitAndRankSame(a, b []hands) bool {
	if a[0].Rank == b[0].Rank &&
		a[1].Rank == b[1].Rank &&
		a[2].Rank == b[2].Rank &&
		a[0].Suit == b[0].Suit &&
		a[1].Suit == b[1].Suit &&
		a[2].Suit == b[2].Suit {
		return true
	} else {
		return false
	}
}

func haveA(cards []uint32) bool {
	for _, v := range cards {
		if isLaizi(v) { //跳过癞子
			continue
		}
		if Suit(v) == Spade {
			return true
		}
	}
	return false
}

func haveB(cards []uint32) bool {
	for _, v := range cards {
		if isLaizi(v) { //跳过癞子
			continue
		}
		if Suit(v) == Heart {
			return true
		}
	}
	return false
}

func haveC(cards []uint32) bool {
	for _, v := range cards {
		if isLaizi(v) { //跳过癞子
			continue
		}
		if Suit(v) == Club {
			return true
		}
	}
	return false
}

func haveD(cards []uint32) bool {
	for _, v := range cards {
		if isLaizi(v) { //跳过癞子
			continue
		}
		if Suit(v) == Diamond {
			return true
		}
	}
	return false
}

// 排除癞子后比牌
func ak47Compare2(a, b []uint32) bool {
	if haveA(a) && !haveA(b) { //a有黑桃b没有
		return true
	}
	if haveA(b) && !haveA(a) { //b有黑桃a没有
		return false
	}

	if haveB(a) && !haveB(b) { //a有红桃b没有
		return true
	}
	if haveB(b) && !haveB(a) { //b有红桃a没有
		return false
	}

	if haveC(a) && !haveC(b) { //a有梅花桃b没有
		return true
	}
	if haveC(b) && !haveC(a) { //b有梅花a没有
		return false
	}

	return false
}

func AK47Compare(a, b []uint32, ca, cb []uint32) bool {
	// 计算手牌癞子数
	num1 := GetLaiziNum(a)
	num2 := GetLaiziNum(b)

	if num1 == 3 && num2 == 3 { //两家都是3个癞子
		return ak47Compare(a, b) //比原牌
	}

	//转换手牌
	a1 := toHands(a)
	b1 := toHands(b)
	// 转换变牌
	ca1 := toHands(ca)
	cb1 := toHands(cb)

	if a1 == nil || b1 == nil || ca1 == nil || cb1 == nil {
		return false
	}

	if IsSuitAndRankSame(ca1, cb1) { //判断变牌是否完全相同
		return ak47Compare2(a, b) //比原牌
	}

	return ak47Compare(ca, cb)
}

func splitCards(cards []uint32) (wild_cards []uint32, feilaizi_cards []uint32) {
	for _, v := range cards {
		if isLaizi(v) {
			wild_cards = append(wild_cards, v)
		} else {
			feilaizi_cards = append(feilaizi_cards, v)
		}
	}
	return
}

func getChangeCard(cards []uint32) (new_cards, ccards []uint32) {
	wild_cards, feilaizi_cards := splitCards(cards)
	if len(wild_cards) == 0 { //没有癞子,变牌等于原牌
		new_cards = cards
		ccards = cards
	} else if len(wild_cards) == 1 { //一张癞子，按joker变牌
		new_cards = append(new_cards, wild_cards...)     //癞子在前
		new_cards = append(new_cards, feilaizi_cards...) //普通牌在后
		if IsA(feilaizi_cards) {
			ccards = append(ccards, ChangeA(feilaizi_cards)) //变牌在前
			ccards = append(ccards, feilaizi_cards...)       //普通牌在后
		} else if IsB(feilaizi_cards) {
			ccards = append(ccards, ChangeB(feilaizi_cards))
			ccards = append(ccards, feilaizi_cards...)
		} else if IsC(feilaizi_cards) {
			ccards = append(ccards, ChangeC(feilaizi_cards))
			ccards = append(ccards, feilaizi_cards...)
		} else if IsD(feilaizi_cards) {
			ccards = append(ccards, ChangeD(feilaizi_cards))
			ccards = append(ccards, feilaizi_cards...)
		} else if IsE(feilaizi_cards) {
			ccards = append(ccards, ChangeE(feilaizi_cards))
			ccards = append(ccards, feilaizi_cards...)
		}
	} else if len(wild_cards) == 2 { //两张癞子，豹子
		new_cards = append(new_cards, wild_cards...)     //癞子在前
		new_cards = append(new_cards, feilaizi_cards...) //普通牌在后

		card := feilaizi_cards[0]
		rank := Rank(card)
		suita, suitb := GetSuit(Suit(card))

		carda := suita + rank
		cardb := suitb + rank

		ccards = append(ccards, carda, cardb)
		ccards = append(ccards, card)

	} else if len(wild_cards) == 3 { //三张癞子，AAA
		new_cards = append(new_cards, wild_cards...)     //癞子在前
		new_cards = append(new_cards, feilaizi_cards...) //普通牌在后

		carda := Spade + Ace
		cardb := Heart + Ace
		cardc := Club + Ace

		ccards = append(ccards, carda, cardb, cardc) //3张A
	}
	return
}

func GetAK47MaxCard(max, cmax, wild_cards, current_cards []uint32) (bigger, cbigger, remain_wild_cards, remain_current_cards []uint32) {
	big := [][][]uint32{}

	all_cards := append([]uint32{}, wild_cards...)
	all_cards = append(all_cards, current_cards...)

	l := len(all_cards)
	for i := 0; i < l; i++ {
		for j := i + 1; j < l; j++ {
			for k := j + 1; k < l; k++ {
				c := []uint32{all_cards[i], all_cards[j], all_cards[k]}
				nc, cc := getChangeCard(c)
				if AK47Compare(nc, max, cc, cmax) {
					big = append(big, [][]uint32{nc, cc})
				}
			}
		}
	}

	if len(big) == 0 {
		remain_wild_cards = wild_cards
		remain_current_cards = current_cards
		return
	}

	random := big[utils.RandIntN(len(big))]
	bigger = random[0]  //原牌
	cbigger = random[1] //变牌

	for _, v := range wild_cards {
		if v != bigger[0] && v != bigger[1] && v != bigger[2] {
			remain_wild_cards = append(remain_wild_cards, v)
		}
	}

	for _, v := range current_cards {
		if v != bigger[0] && v != bigger[1] && v != bigger[2] {
			remain_current_cards = append(remain_current_cards, v)
		}
	}

	return
}
