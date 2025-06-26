package algo

import (
	"sort"
	"strconv"
	"strings"
)

// 手牌分组
type HandCardDivide struct {
	BaseGroups [][]uint32
	LeftCards  []uint32
}

func (d HandCardDivide) String() (ret string) {
	ret += "group->{"
	for _, g := range d.BaseGroups {
		ret += "["
		for _, card := range g {
			if card == RMJOKERS[0] {
				ret += "Jk1,"
			} else if card == RMJOKERS[1] {
				ret += "Jk2,"
			} else {
				ret += GetSuitString(card) + GetRankString(card) + ","
			}
		}
		ret += "]"
	}
	ret += "}"

	ret += " left->{"
	for _, card := range d.LeftCards {
		if card == RMJOKERS[0] {
			ret += "Jk1"
		} else if card == RMJOKERS[1] {
			ret += "Jk2"
		} else {
			ret += GetSuitString(card) + GetRankString(card) + ","
		}
	}
	ret += "}"
	return
	// return fmt.Sprintf("%#v, %v", d.BaseGroups, d.LeftCards)
}

func newHandCardDivide(BaseGroups [][]uint32, LeftCards []uint32) *HandCardDivide {
	instance := new(HandCardDivide)
	instance.BaseGroups = BaseGroups
	instance.LeftCards = LeftCards
	return instance
}

// 相同基本组去重
func distinctAllBaseGroup(baseGroups [][]uint32) (distinct [][]uint32) {
	keys := make(map[string]bool, 16)
	for _, g := range baseGroups {
		key := sliceKey(g)
		// 手牌中有多组相同的纯同花, 纯豹子
		if keys[key] && !(IsPureTonghuashun(g) || IsPureBaozi(g)) {
			continue
		}
		distinct = append(distinct, g)
		keys[key] = true
	}
	return
}

func sliceKey[T int | uint32](slice []T) (ret string) {
	cc := make([]T, len(slice))
	copy(cc, slice)
	sort.Slice(cc, func(i, j int) bool { return cc[i] < cc[j] })

	for _, c := range cc {
		ret += "," + strconv.Itoa(int(c))
	}
	return
}

// 手牌分组
func FindAllHandCardDivide(cards []uint32, wildCard uint32, baseGroupCount int) []*HandCardDivide {
	tempCards := make([]uint32, len(cards))
	copy(tempCards, cards)
	// 1 找到所有的基本组(3张的顺子或豹子)
	allBaseGroup := FindAllBaseGroup(tempCards, wildCard)
	// 基本组相同牌组去重一下
	allBaseGroup = distinctAllBaseGroup(allBaseGroup)

	// fmt.Println("FindAllHandCardDivide", "allBaseGroup=", allBaseGroup)
	if len(allBaseGroup) <= 0 {
		handCardDivides := make([]*HandCardDivide, 0)
		// 直接按花色分组
		allGroups := make([][]uint32, 0)
		cardsSlitByColor := make(map[uint32][]uint32, 0)
		for _, v := range cards {
			cardColor := Suit(v)
			if cardsSlitByColor[cardColor] != nil {
				cardsSlitByColor[cardColor] = append(cardsSlitByColor[cardColor], v)
			} else {
				cardsSlitByColor[cardColor] = []uint32{v}
			}
		}
		for _, v := range cardsSlitByColor {
			Sort(v) // 每组排序
			allGroups = append(allGroups, v)
		}
		handCardDivides = append(handCardDivides, newHandCardDivide(allGroups, []uint32{}))
		return handCardDivides
	}
	// 2 生成基本组的所有可能的排列组合，多基本组的列表 [][]int
	allCombinations := make([][]int, 0)
	for groupCount := 1; groupCount <= baseGroupCount; groupCount++ {
		tempCombinations := GenerateCombination(len(allBaseGroup), groupCount) // {{0,1,2,3},{0,1,2,4},{1,2,3,4}}
		allCombinations = append(allCombinations, tempCombinations...)
	}
	// fmt.Println("FindAllHandCardDivide", "allCombinations=", allCombinations)
	// 3 找出所有有效的排列组合的基本组与剩余牌
	handCardDivides := make([]*HandCardDivide, 0)
	laiziDivides := make([]*HandCardDivide, 0)
	for _, comnimation := range allCombinations {
		groupCards := make([]uint32, 0)
		for _, v := range comnimation {
			groupCards = append(groupCards, allBaseGroup[v]...)
		}

		if !ContainsArray(cards, groupCards) {
			continue
		}

		// fmt.Println("FindAllHandCardDivide", "groupCards=", groupCards)
		leftCards := KickOutCards(cards, groupCards)
		allGroups := make([][]uint32, 0)
		for _, v := range comnimation {
			allGroups = append(allGroups, allBaseGroup[v])
		}

		// 剩余癞子牌2张以上
		var laizis []uint32 //癞子
		var normals []uint32
		for _, v := range leftCards {
			if IsLaizi(v, wildCard) {
				laizis = append(laizis, v)
			} else {
				normals = append(normals, v)
			}
		}
		// 癞子牌2张以上,普通散牌太多影响性能
		if len(laizis) >= 2 && len(normals) > 0 && len(normals) <= 3 {
			extNormals := make(map[uint32]bool, len(normals))
			// 降序大小王排前面
			sort.Slice(laizis, func(i, j int) bool { return laizis[i] > laizis[j] })
			for i, card := range normals {
				extGroup := []uint32{card, laizis[0], laizis[1]}
				if extNormals[card] {
					continue
				} else {
					extNormals[card] = true
				}
				lefts := make([]uint32, 0, len(leftCards)-3)
				lefts = append(lefts, normals[:i]...)
				lefts = append(lefts, normals[i+1:]...)
				lefts = append(lefts, laizis[2:]...)
				extGroups := make([][]uint32, 0, len(allGroups)+1)
				extGroups = append(extGroups, allGroups...)
				extGroups = append(extGroups, extGroup)
				laiziDivides = append(laiziDivides, newHandCardDivide(extGroups, lefts))
			}
		}

		handCardDivides = append(handCardDivides, newHandCardDivide(allGroups, leftCards))
	}
	handCardDivides = append(handCardDivides, laiziDivides...)
	return handCardDivides
}

// @Desc 判断cards中是否拥有card,拥有返回true
func Contains(cards []uint32, card uint32) bool {
	for _, tempCard := range cards {
		if card == tempCard {
			return true
		}
	}
	return false
}

func ContainsArray(cards, cards2 []uint32) bool {
	cardsMap := make(map[uint32]int)
	for _, card := range cards {
		cardsMap[card]++
	}
	for _, card := range cards2 {
		cardsMap[card]--
		if cardsMap[card] < 0 {
			return false
		}
	}
	return true
}

// 返回src中踢出kickOut后的结果
func KickOutCards(src []uint32, kickOut []uint32) []uint32 {
	srcMap := make(map[uint32]int)
	for _, v := range src {
		srcMap[v]++
	}
	for _, card := range kickOut {
		if _, ok := srcMap[card]; ok {
			srcMap[card] -= 1
		} else {
			srcMap[card] = 0
		}
	}
	result := make([]uint32, 0)
	for cardValue, cardCount := range srcMap {
		for k := 0; k < cardCount; k++ {
			result = append(result, cardValue)
		}
	}
	return result
}

// 生成基本组的所有的排列组合
func GenerateCombination(baseGroupCount int, combinationSize int) [][]int {
	if combinationSize == 5 {
		result := make([][]int, 0)
		for i := 0; i < baseGroupCount-combinationSize+1; i++ {
			for j := i + 1; j < baseGroupCount-combinationSize+2; j++ {
				for k := j + 1; k < baseGroupCount-combinationSize+3; k++ {
					for m := k + 1; m < baseGroupCount-combinationSize+4; m++ {
						for n := m + 1; n < baseGroupCount-combinationSize+5; n++ {
							combination := []int{i, j, k, m, n}
							result = append(result, combination)
						}
					}
				}
			}
		}
		return result
	}
	if combinationSize == 4 {
		result := make([][]int, 0)
		for i := 0; i < baseGroupCount-combinationSize+1; i++ {
			for j := i + 1; j < baseGroupCount-combinationSize+2; j++ {
				for k := j + 1; k < baseGroupCount-combinationSize+3; k++ {
					for m := k + 1; m < baseGroupCount-combinationSize+4; m++ {
						combination := []int{i, j, k, m}
						result = append(result, combination)
					}
				}
			}
		}
		return result
	}
	if combinationSize == 3 {
		result := make([][]int, 0)
		for i := 0; i < baseGroupCount-combinationSize+1; i++ {
			for j := i + 1; j < baseGroupCount-combinationSize+2; j++ {
				for k := j + 1; k < baseGroupCount-combinationSize+3; k++ {
					combination := []int{i, j, k}
					result = append(result, combination)
				}
			}
		}
		return result
	}
	if combinationSize == 2 {
		result := make([][]int, 0)
		for i := 0; i < baseGroupCount-combinationSize+1; i++ {
			for j := i + 1; j < baseGroupCount-combinationSize+2; j++ {
				combination := []int{i, j}
				result = append(result, combination)
			}
		}
		return result
	}
	if combinationSize == 1 {
		result := make([][]int, 0)
		for i := 0; i < baseGroupCount-combinationSize+1; i++ {
			combination := []int{i}
			result = append(result, combination)
		}
		return result
	}
	return [][]int{}
}

// 找出所有的基本组(3张的顺子或豹子)
func FindAllBaseGroup(cards []uint32, wildCard uint32) [][]uint32 {
	result := make([][]uint32, 0)

	// 在牌中找出所有纯基本顺
	allBaseSequence := FindAllBasePureSequence(cards)
	// fmt.Println("FindAllBaseGroup", "allBaseSequence=", allBaseSequence)
	result = append(result, allBaseSequence...)
	// 在普通牌中找出所有纯基本三条
	allBaseSet := FindAllBaseSet(cards)
	// fmt.Println("FindAllBaseSet", "allBaseSet=", allBaseSet)
	result = append(result, allBaseSet...)
	// 在牌中找出所有鬼基本顺
	allImpureSequence := FindAllBaseImpureSequence(cards, wildCard)
	// fmt.Println("FindAllBaseImpureSequence", "allImpureSequence=", allImpureSequence)
	result = append(result, allImpureSequence...)
	// 在牌中找出所有鬼基本三条
	allImpureSet := FindAllBaseImpureSet(cards, wildCard)
	// fmt.Println("FindAllBaseImpureSet", "allImpureSet=", allImpureSet)
	result = append(result, allImpureSet...)
	return result
}

// 找出所有的基本顺(3张) 1 去重
func FindAllBasePureSequence(cards []uint32) [][]uint32 {
	result := make([][]uint32, 0)
	// 按牌色分组
	cardsSlitByColor := make(map[uint32][]uint32, 0)
	for _, v := range cards {
		cardColor := Suit(v)
		if cardsSlitByColor[cardColor] != nil {
			cardsSlitByColor[cardColor] = append(cardsSlitByColor[cardColor], v)
		} else {
			cardsSlitByColor[cardColor] = []uint32{v}
		}
	}
	for _, v := range cardsSlitByColor {
		tempCardsSlitByCount := SplitColorCards(v)
		for _, v := range tempCardsSlitByCount {
			if len(v) < 3 {
				continue
			}
			// 每组排序
			Sort(v)

			// 每组遍历
			for i := 0; i < len(v)-2; i++ {
				group := []uint32{v[i], v[i+1], v[i+2]}
				if IsPureTonghuashun(group) {
					result = append(result, group)
				}
			}
			// QKA
			group := []uint32{v[len(v)-1], v[len(v)-2], v[0]}
			if IsPureTonghuashun(group) {
				result = append(result, group)
			}
		}
	}
	return result
}

// 将同花色牌分组按数量
func SplitColorCards(cardsSlitByColor []uint32) [][]uint32 {
	// 按牌点分组
	cardsSlitByValue := make(map[uint32]int, 0)
	for _, cardValue := range cardsSlitByColor {
		cardsSlitByValue[cardValue] += 1
	}
	result := make([][]uint32, 2)
	result[0] = make([]uint32, 0)
	result[1] = make([]uint32, 0)
	for cardValue, cardCount := range cardsSlitByValue {
		result[0] = append(result[0], cardValue)
		if cardCount >= 2 {
			result[1] = append(result[1], cardValue)
		}
	}
	return result
}

// 找出所有的基本豹(3张)
func FindAllBaseSet(cards []uint32) [][]uint32 {
	cardsSlitByValue := GetCardsForSet(cards)
	result := make([][]uint32, 0)
	// 每组遍历
	for _, v := range cardsSlitByValue {
		if len(v) == 3 {
			result = append(result, []uint32{v[0], v[1], v[2]})
		} else if len(v) == 4 {
			result = append(result, []uint32{v[0], v[1], v[2]})
			result = append(result, []uint32{v[0], v[1], v[3]})
			result = append(result, []uint32{v[0], v[2], v[3]})
			result = append(result, []uint32{v[1], v[2], v[3]})
		}
	}
	return result
}

func GetCardsForSet(cards []uint32) map[uint32][]uint32 {
	added := make(map[uint32]int, 0) //去重用
	// 按牌点分组
	cardsSlitByValue := make(map[uint32][]uint32, 0)
	for _, v := range cards {
		if IsCardJoker(v) {
			continue
		}
		cardValue := Rank(v)
		if cardsSlitByValue[cardValue] != nil {
			if _, ok := added[v]; !ok {
				added[v] = 1
				cardsSlitByValue[cardValue] = append(cardsSlitByValue[cardValue], v)
			}
		} else {
			cardsSlitByValue[cardValue] = []uint32{v}
			added[v] = 1
		}
	}
	// 每组排序
	for _, v := range cardsSlitByValue {
		Sort(v)
	}
	return cardsSlitByValue
}

// 找出所有的非纯基本顺(3张) 1 去重
func FindAllBaseImpureSequence(cards []uint32, wildCard uint32) [][]uint32 {
	result := make([][]uint32, 0)
	// 找出所有的癞子
	normalCards, wildCards := CheckOutAllWild(cards, wildCard)
	// 找出所有的缺一门顺(2张)
	allLackSequence := FindAllLackSequence(normalCards)
	for _, lackSequence := range allLackSequence {
		for _, v := range wildCards {
			impureSequence := make([]uint32, 0)
			impureSequence = append(impureSequence, lackSequence...)
			impureSequence = append(impureSequence, v)
			result = append(result, impureSequence)
		}
	}
	return result
}

// 找出所有的癞子牌
func CheckOutAllWild(cards []uint32, wildCard uint32) ([]uint32, []uint32) {
	normalCards := make([]uint32, 0)
	wildCards := make([]uint32, 0)
	for _, v := range cards {
		if IsLaizi(v, wildCard) {
			wildCards = append(wildCards, v)
		} else {
			normalCards = append(normalCards, v)
		}
	}
	return normalCards, wildCards
}

// 在散牌牌集中分析出所有的缺一门组合(2张的顺子或豹子)
func FindAllLackGroup(cards []uint32) [][]uint32 {
	result := make([][]uint32, 0)
	AllLackSequence := FindAllLackSequence(cards)
	result = append(result, AllLackSequence...)
	AllLackSet := FindAllLackSet(cards)
	result = append(result, AllLackSet...)
	return result
}

// 找出所有的缺一门顺(2张) 1 去重
func FindAllLackSequence(cards []uint32) [][]uint32 {
	result := make([][]uint32, 0)
	// 按牌色分组
	cardsSlitByColor := make(map[uint32][]uint32, 0)
	for _, v := range cards {
		cardColor := Suit(v)
		cardsSlitByColor[cardColor] = append(cardsSlitByColor[cardColor], v)
	}
	for _, v := range cardsSlitByColor {
		tempCardsSlitByCount := SplitColorCards(v)
		for _, v := range tempCardsSlitByCount {
			if len(v) < 2 {
				continue
			}
			// 每组排序
			Sort(v)

			// 每组遍历
			for i := 0; i < len(v)-1; i++ {
				//去除花色是否相连
				group := []uint32{v[i], v[i+1]}
				ranks := GetRanks(group)
				if ranks[0] == ranks[1]-1 || ranks[0] == ranks[1]-2 {
					result = append(result, group)
				}
			}

			group := []uint32{v[0], v[len(v)-1]}
			ranks := GetRanks(group)
			if (ranks[1] == Queen || ranks[1] == King) && ranks[0] == Ace {
				result = append(result, group)
			}
		}
	}
	return result
}

// 找出所有的非纯基本豹(3张)
func FindAllBaseImpureSet(cards []uint32, wildCard uint32) [][]uint32 {
	result := make([][]uint32, 0)
	// 找出所有的癞子
	normalCards, wildCards := CheckOutAllWild(cards, wildCard)
	// 找出所有的缺一门豹(2张)
	allLackSet := FindAllLackSet(normalCards)
	for _, lackSet := range allLackSet {
		for _, v := range wildCards {
			impureSet := make([]uint32, 0)
			impureSet = append(impureSet, lackSet...)
			impureSet = append(impureSet, v)
			result = append(result, impureSet)
		}
	}
	return result
}

// 找出所有的缺一门豹(2张)
func FindAllLackSet(cards []uint32) [][]uint32 {
	cardsSlitByValue := GetCardsForSet(cards)
	result := make([][]uint32, 0)
	// 每组遍历
	for _, v := range cardsSlitByValue {
		if len(v) == 2 {
			result = append(result, []uint32{v[0], v[1]})
		} else if len(v) == 3 {
			result = append(result, []uint32{v[0], v[1]})
			result = append(result, []uint32{v[0], v[2]})
			result = append(result, []uint32{v[1], v[2]})
		} else if len(v) == 4 {
			result = append(result, []uint32{v[0], v[1]})
			result = append(result, []uint32{v[0], v[2]})
			result = append(result, []uint32{v[0], v[3]})
			result = append(result, []uint32{v[1], v[2]})
			result = append(result, []uint32{v[1], v[3]})
			result = append(result, []uint32{v[2], v[3]})
		}
	}
	return result
}

// 获取 rm 牌分值
// @Param hasOneLife 是否已经拥有第一生命
func GetCardPoint(card uint32, wildCard uint32, hasOneLife bool) uint32 {
	// joker
	if card == RMJOKERS[0] || card == RMJOKERS[1] {
		return 0
	}
	r1, r2 := Rank(card), Rank(wildCard)
	if r1 == r2 && hasOneLife { //百搭牌在拥有1st Life后，变为0分，其余时候分数与点数相同
		return 0
	}
	if r1 > 10 || r1 == Ace { // J、Q、K、A均为10分
		return 10
	}
	return r1
}

/*
@Desc 获取Cards中所有牌在结算时应该计算的总点数
@Param cards 牌
@Param wildCard 当前牌局的随机百搭牌
@Param hasOneLife 是否已经拥有第一生命
*/
func GetCardsPoint(cards []uint32, wildCard uint32, hasOneLife bool) (cardPoint int64) {
	cardPoint = 0
	for _, tempCard := range cards {
		cardPoint = cardPoint + int64(GetCardPoint(tempCard, wildCard, hasOneLife))
	}
	return cardPoint
}

func IsCardWild(card uint32, wildCard uint32) bool {
	r1, r2 := Rank(card), Rank(wildCard)
	return r1 == r2
}
func IsCardJoker(card uint32) bool {
	return card == RMJOKERS[0] || card == RMJOKERS[1]
}

// 生成 baseGroups 唯一key
func getBaseGroupsKey(baseGroups [][]uint32) string {
	groupKeys := make([]string, len(baseGroups))
	for i, g := range baseGroups {
		key := getGroupKey(g)
		groupKeys[i] = key
	}
	sort.Strings(groupKeys)
	return strings.Join(groupKeys, "-")
}

func getGroupKey(group []uint32) string {
	cc := make([]uint32, len(group))
	copy(cc, group)
	sort.Slice(cc, func(i, j int) bool { return cc[i] < cc[j] })
	var key string
	for i, c := range cc {
		if i == 0 {
			key += strconv.Itoa(int(c))
		} else {
			key += "," + strconv.Itoa(int(c))
		}
	}
	return key
}

func HandleGroupAppend(cardArray []uint32, wildCard uint32, handCardDivides []*HandCardDivide) []*HandCardDivide {
	validCardGroupList := make([]*HandCardDivide, 0)
	baseGroupsKeys := make(map[string]bool, 16)
	for ii, handCardDivide := range handCardDivides {
		_ = ii
		if len(handCardDivide.LeftCards) > 5 {
			// 癞子牌张数
			// var laizis int //癞子
			// for _, v := range handCardDivide.LeftCards {
			// 	if IsLaizi(v, wildCard) {
			// 		laizis++
			// 	}
			// }
			// if len(handCardDivide.LeftCards)-laizis > 5 { // 剩余牌太多了，生成全排列的性能很差，直接按照花色分组好了
			// fmt.Println("剩余牌太多了，生成全排列的性能很差，直接按照花色分组好了", handCardDivide.LeftCards)
			cardGroupList := &HandCardDivide{BaseGroups: handCardDivide.BaseGroups}
			// 按牌色分组
			cardsSlitByColor := make(map[uint32][]uint32, 0)
			for _, v := range handCardDivide.LeftCards {
				cardColor := Suit(v)
				cardsSlitByColor[cardColor] = append(cardsSlitByColor[cardColor], v)
			}
			for _, v := range cardsSlitByColor {
				if len(v) > 0 {
					cardGroupList.BaseGroups = append(cardGroupList.BaseGroups, v)
				}
			}
			validCardGroupList = append(validCardGroupList, cardGroupList)
			continue
			// }
		}

		baseGroupPermutationNums := make([]int, 0)
		for k := 0; k < len(handCardDivide.BaseGroups); k++ {
			baseGroupPermutationNums = append(baseGroupPermutationNums, k)
		}
		// fmt.Println("GroupCardMode1 GroupHandCards", "groupId=", groupId, "baseGroupPermutationNums=", baseGroupPermutationNums)
		baseGroupPermutations := GenIntSlicePermutationOptimize(baseGroupPermutationNums) //加入顺序要考虑baseGroup的全排列
		// fmt.Println("GroupCardMode1 GroupHandCards", "groupId=", groupId, "baseGroupPermutations=", baseGroupPermutations)
		leftCardPermutationNums := make([]int, 0)
		for k := 0; k < len(handCardDivide.LeftCards); k++ {
			leftCardPermutationNums = append(leftCardPermutationNums, k)
		}
		// fmt.Println("GroupCardMode1 GroupHandCards", "groupId=", groupId, "leftCardPermutationNums=", leftCardPermutationNums)
		leftCardPermutations := GenIntSlicePermutationOptimize(leftCardPermutationNums) //加入顺序要考虑leftCard的全排列
		// fmt.Println("GroupCardMode1 GroupHandCards", "groupId=", groupId, "leftCardPermutations=", leftCardPermutations)
		// 找出所有有效排列组合中能胡牌的排列组合，即剩余牌能否放入组合牌中
		for _, baseGroupPermutation := range baseGroupPermutations {
			for _, leftCardPermutation := range leftCardPermutations {
				baseGroups := make([][]uint32, 0)
				for _, v := range baseGroupPermutation {
					baseGroup := handCardDivide.BaseGroups[v]
					newBaseGroup := make([]uint32, len(baseGroup))
					copy(newBaseGroup, baseGroup)
					baseGroups = append(baseGroups, newBaseGroup)
				}
				for _, v := range leftCardPermutation {
					for i := range baseGroups {
						if ValidAppend(baseGroups[i], handCardDivide.LeftCards[v], wildCard) {
							baseGroups[i] = append(baseGroups[i], handCardDivide.LeftCards[v])
							break
						}
					}
				}

				// var findJoker = func(cards []uint32) bool {
				// 	cardsMap := make(map[uint32]int, 3)
				// 	for _, card := range cards {
				// 		cardsMap[card]++
				// 	}
				// 	return cardsMap[RMJOKERS[0]] == 2 && cardsMap[38] == 1
				// }
				// _ = findJoker

				// baseGroups 去重
				key := getBaseGroupsKey(baseGroups)
				if baseGroupsKeys[key] {
					continue
				}
				baseGroupsKeys[key] = true

				// cardGroupList := make([]*card.CardGroup, 0)
				cardGroupList := &HandCardDivide{}
				groupCards := make([]uint32, 0)
				for _, baseGroup := range baseGroups {
					groupCards = append(groupCards, baseGroup...)
					cardGroupList.BaseGroups = append(cardGroupList.BaseGroups, baseGroup)
				}
				leftCards := KickOutCards(cardArray, groupCards) //baseGroupPermutation 配 leftCardPermutation 后剩余的散牌
				if len(leftCards) > 0 {

					// 剩余牌为多张癞子单张散牌情况
					// if len(handCardDivide.LeftCards) == 3 {
					// 	fmt.Println(CardsString(handCardDivide.LeftCards))
					// 	if len(handCardDivide.LeftCards) >= 3 && ValidRmGroup(handCardDivide.LeftCards, wildCard) {
					// 		// handCardDivide.LeftCards
					// 		// allGroups = append(allGroups, leftCards)
					// 		// leftCards = []uint32{}
					// 		cardGroupList := &HandCardDivide{BaseGroups: handCardDivide.BaseGroups}
					// 		cardGroupList.BaseGroups = append(cardGroupList.BaseGroups, handCardDivide.LeftCards)
					// 		validCardGroupList = append(validCardGroupList, cardGroupList)
					// 		continue
					// 	}
					// }

					cardGroupList.BaseGroups = append(cardGroupList.BaseGroups, leftCards)
				}
				validCardGroupList = append(validCardGroupList, cardGroupList)
			}
		}
	}
	// fmt.Println("GroupCardMode1 GroupHandCards", "groupId=", groupId, "End")
	return validCardGroupList
}

// 生成全排列
func GenIntSlicePermutationOptimize(nums []int) [][]int {
	var result [][]int
	visited := make([]bool, len(nums))
	backtrack(nums, []int{}, &result, visited)
	return result
}

func backtrack(nums []int, current []int, result *[][]int, visited []bool) {
	if len(current) == len(nums) {
		*result = append(*result, current)
		return
	}

	for i := 0; i < len(nums); i++ {
		if visited[i] {
			continue
		}

		visited[i] = true
		current = append(current, nums[i])
		backtrack(nums, current, result, visited)
		current = current[:len(current)-1]
		visited[i] = false
	}
}

func ValidRmGroup(cards []uint32, wildCard uint32) bool {
	if len(cards) < 3 {
		return false
	}
	if IsPureTonghuashun(cards) {
		return true
	}
	if IsTonghuashun(cards, wildCard) {
		return true
	}
	if IsBaozi(cards, wildCard) {
		return true
	}
	return false
}

// 能否进行有效追加
func ValidAppend(baseGroup []uint32, appendCard, wildCard uint32) bool {
	newGroupCards := make([]uint32, len(baseGroup))
	copy(newGroupCards, baseGroup)
	newGroupCards = append(newGroupCards, appendCard)
	return ValidRmGroup(newGroupCards, wildCard)
}

// 将牌组列表分解为有效组合与散牌牌集
func splitCookiesCards(cardGroupList [][]uint32, wildCard uint32) ([][]uint32, []uint32) {
	allBaseGroups := make([][]uint32, 0) // 所有的有效组合
	allCookiesCards := make([]uint32, 0) // 散牌牌集
	for _, group := range cardGroupList {
		if ValidRmGroup(group, wildCard) {
			allBaseGroups = append(allBaseGroups, group)
		} else {
			allCookiesCards = append(allCookiesCards, group...)
		}
	}
	return allBaseGroups, allCookiesCards
}

// TidyDivideLeftCard 将HandCardDivide 的无效牌组移动到 leftCards 中
func TidyDivideLeftCard(divide *HandCardDivide, wildCard uint32) {
	groups := make([][]uint32, 0)
	for _, group := range divide.BaseGroups {
		if !ValidRmGroup(group, wildCard) {
			divide.LeftCards = append(divide.LeftCards, group...)
		} else {
			groups = append(groups, group)
		}
	}
	divide.BaseGroups = groups
}

// 是否有纯顺子
func hasPureSequence(cardGroupList [][]uint32) bool {
	for _, cardGroup := range cardGroupList {
		if IsPureTonghuashun(cardGroup) {
			return true
		}
	}
	return false
}

// 返回相同点数的4种花色牌
func cardPointTo4Color(card uint32) []uint32 {
	rank := Rank(card)
	return []uint32{rank | Spade, rank | Heart, rank | Club, rank | Diamond}
}

// 返回相同点数的其他三种花色牌
func cardPointTo3Color(card uint32) (ret []uint32) {
	rank, suit := Rank(card), Suit(card)
	colors := []uint32{Spade, Heart, Club, Diamond}
	for _, color := range colors {
		if color == suit {
			continue
		}
		ret = append(ret, rank|color)
	}
	return
}
