package algo

import "goserver/pkg/utils"

// 牌型定义
const (
	LuckDuiZi       = 3 //对子
	LuckTongHua     = 4 //同花
	LuckShunZi      = 5 //顺子
	LuckTongHuaShun = 6 //同花顺
	LuckBaoZi       = 7 //豹子
)

// 通过牌数据获取牌型
func RedBlackType(cs []uint32) (i uint32) {
	hs := toHands(cs)
	i = toRedBlackType(hs)
	return
}

func toRedBlackType(hs []hands) uint32 {
	if len(hs) != 3 {
		return Null
	}
	//豹子
	if isBaoZi(hs) {
		return LuckBaoZi
	}
	//同花顺
	if isTongHuaShun(hs) {
		return LuckTongHuaShun
	}
	//顺子
	if isShunZi(hs) {
		return LuckShunZi
	}
	//同花
	if isTongHua(hs) {
		return LuckTongHua
	}
	//大对子
	if isLuckDuiZi(hs) {
		return LuckDuiZi
	}
	//小对子
	if isDuiZi(hs) {
		return DuiZi
	}
	return GaoPai
}

// 是否是对子
func isLuckDuiZi(hs []hands) bool {
	if (hs[0].Rank == hs[1].Rank && (hs[0].Rank >= Nine || hs[0].Rank == Ace)) || (hs[1].Rank == hs[2].Rank && (hs[1].Rank >= Nine || hs[1].Rank == Ace)) {

		return true
	}
	return false
}

func RBGetCard(cardtype uint32, current_cards []uint32) (cards []uint32, remain_cards []uint32) {
	l := len(current_cards)

	BaoZi_List := [][]uint32{}
	TongHuaShun_List := [][]uint32{}
	ShunZi_List := [][]uint32{}
	TongHua_List := [][]uint32{}
	LuckDuiZi_List := [][]uint32{}
	DuiZi_List := [][]uint32{}
	GaoPai_List := [][]uint32{}

	for i := 0; i < l; i++ {
		for j := i + 1; j < l; j++ {
			for k := j + 1; k < l; k++ {
				c := []uint32{current_cards[i], current_cards[j], current_cards[k]}
				switch RedBlackType(c) {
				case LuckBaoZi:
					BaoZi_List = append(BaoZi_List, c)
				case LuckTongHuaShun:
					TongHuaShun_List = append(TongHuaShun_List, c)
				case LuckShunZi:
					ShunZi_List = append(ShunZi_List, c)
				case LuckTongHua:
					TongHua_List = append(TongHua_List, c)
				case LuckDuiZi:
					LuckDuiZi_List = append(LuckDuiZi_List, c)
				case DuiZi:
					DuiZi_List = append(DuiZi_List, c)
				case GaoPai:
					GaoPai_List = append(GaoPai_List, c)
				}
			}
		}
	}

	switch cardtype {
	case LuckBaoZi:
		if len(BaoZi_List) != 0 {
			cards = BaoZi_List[utils.RandIntN(len(BaoZi_List))]
		} else {
			i, j, k := RandomIndex(len(current_cards))
			cards = []uint32{current_cards[i], current_cards[j], current_cards[k]}
		}
	case LuckTongHuaShun:
		if len(TongHuaShun_List) != 0 {
			cards = TongHuaShun_List[utils.RandIntN(len(TongHuaShun_List))]
		} else {
			i, j, k := RandomIndex(len(current_cards))
			cards = []uint32{current_cards[i], current_cards[j], current_cards[k]}
		}
	case LuckShunZi:
		if len(ShunZi_List) != 0 {
			cards = ShunZi_List[utils.RandIntN(len(ShunZi_List))]
		} else {
			i, j, k := RandomIndex(len(current_cards))
			cards = []uint32{current_cards[i], current_cards[j], current_cards[k]}
		}
	case LuckTongHua:
		if len(TongHua_List) != 0 {
			cards = TongHua_List[utils.RandIntN(len(TongHua_List))]
		} else {
			i, j, k := RandomIndex(len(current_cards))
			cards = []uint32{current_cards[i], current_cards[j], current_cards[k]}
		}
	case LuckDuiZi:
		if len(LuckDuiZi_List) != 0 {
			cards = LuckDuiZi_List[utils.RandIntN(len(LuckDuiZi_List))]
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
