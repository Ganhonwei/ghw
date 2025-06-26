package algo

import (
	"fmt"
	"goserver/pkg/utils"
	"sort"
)

/*玩家弃牌，机器人在每回合开始摸牌前，根据手牌的好还，根据以下概率选择是否放弃本局游戏*/
func ShouldDrop(mCards []uint32, wildCard uint32, divide *HandCardDivide) bool {
	// cardGroupList := SimpleGroupTheCards(gameData.mCards, gameData.WildCard)
	// cookiesCount := gameData.GetCookiesCount(cardGroupList) //玩家手牌中未成功组合牌的数量
	cookiesCount := len(divide.LeftCards)
	if cookiesCount >= 13 {
		return utils.RandMN(1, 100) >= 15
	}
	if cookiesCount >= 7 {
		return utils.RandMN(1, 100) >= 15
	}
	if cookiesCount >= 4 {
		return utils.RandMN(1, 100) >= 10
	}
	if cookiesCount >= 1 {
		return utils.RandMN(1, 100) >= 5
	}
	return false
}

// LetsOutCard rummy 机器人出牌
// FaceUpCard 弃牌堆
func LetsOutCard(mCards []uint32, FaceUpCard []uint32, wildCard uint32) (uint32, *HandCardDivide, bool) {
	handCards := make([]uint32, 0)
	handCards = append(handCards, mCards...)
	if len(handCards) != 14 {
		panic(fmt.Errorf("invalid cards, count=%v", len(handCards)))
	}
	// validCardGroupList := converHandCardDivideToCardGroupList(FindAllHandCardDivide(handCards, wildCard, 4), wildCard)
	validCardGroupList := FindAllHandCardDivide(handCards, wildCard, 4)
	_ = validCardGroupList

	// 找出有纯顺子并且有第二个顺子的
	secondLifeValidCardGroupList := findAllSecondLifeValidCardGroupList(validCardGroupList, wildCard)
	_ = secondLifeValidCardGroupList
	if len(secondLifeValidCardGroupList) > 0 {
		return handleSecondLifeValid(handCards, validCardGroupList, FaceUpCard, wildCard)
	}
	// 找出有纯顺子并且没有第二个顺子的
	firstLifeValidCardGroupList := findAllFirstLifeValidCardGroupList(validCardGroupList, wildCard)
	if len(firstLifeValidCardGroupList) > 0 {
		return handleFirstLifeValid(handCards, validCardGroupList, FaceUpCard, wildCard)
	}

	return handleNoneLifeValid(handCards, validCardGroupList, FaceUpCard, wildCard)
}

// 将手牌分组转换为牌组列表
// func converHandCardDivideToCardGroupList(handCardDivides []*HandCardDivide, wildCard uint32) [][]*card.CardGroup {
// 	validCardGroupList := make([][]*card.CardGroup, 0)
// 	for _, handCardDivide := range handCardDivides {
// 		cardGroupList := make([]*card.CardGroup, 0)
// 		for _, baseGroup := range handCardDivide.BaseGroups {
// 			cardGroupList = append(cardGroupList, card.GroupCards(baseGroup, wildCard))
// 		}
// 		cardGroupList = append(cardGroupList, card.NewCardGroup(card.CardGroupType_Cookies, handCardDivide.LeftCards))
// 		validCardGroupList = append(validCardGroupList, cardGroupList)
// 	}
// 	return validCardGroupList
// }

// 找出所有有纯顺子并且有第二个顺子的
func findAllSecondLifeValidCardGroupList(validCardGroupList []*HandCardDivide, wildCard uint32) []*HandCardDivide {
	secondLifeValidCardGroupList := make([]*HandCardDivide, 0)
	for _, v := range validCardGroupList {
		hasPureSequence := false
		sequenceCount := 0
		for _, group := range v.BaseGroups {
			if IsPureTonghuashun(group) {
				sequenceCount = sequenceCount + 1
				hasPureSequence = true
			} else if IsTonghuashun(group, wildCard) {
				sequenceCount = sequenceCount + 1
			}
		}
		if hasPureSequence && sequenceCount >= 2 {
			secondLifeValidCardGroupList = append(secondLifeValidCardGroupList, v)
		}
	}
	return secondLifeValidCardGroupList
}

// 出牌：有至少两条顺子并且其中至少有一条纯顺子牌组
func handleSecondLifeValid(handCards []uint32, validCardGroupList []*HandCardDivide, FaceUpCard []uint32, WildCard uint32) (
	uint32, *HandCardDivide, bool,
) {
	validCardGroupList = HandleGroupAppend(handCards, WildCard, validCardGroupList)
	secondLifeValidCardGroupList := findAllSecondLifeValidCardGroupList(validCardGroupList, WildCard)

	// 有两条生命是否可以胡牌
	for _, divide := range secondLifeValidCardGroupList {
		// allBaseGroups, allCookiesCards := splitCookiesCards(divide.BaseGroups, WildCard)
		TidyDivideLeftCard(divide, WildCard)
		// fmt.Println("--有效牌组", divide.BaseGroups, "以外的剩余散牌牌集", divide.LeftCards)
		// if len(divide.LeftCards) == 0 { // 可以胡牌了
		// 	return 0, divide, true
		// }
		if len(divide.LeftCards) == 1 { // 可以胡牌了
			return divide.LeftCards[0], &HandCardDivide{BaseGroups: divide.BaseGroups}, true
		}
		if len(divide.LeftCards) == 0 { // 没有散牌,可以胡牌了
			found := false
			var outCard uint32
			cardGroups := &HandCardDivide{}
			for _, cardGroup := range divide.BaseGroups {
				if !found {
					if len(cardGroup) >= 4 {
						// 从牌组中抽一张打出, todo 牌组全癞子
						Sort(cardGroup)
						for _, v := range cardGroup {
							if IsCardWild(v, WildCard) || IsCardJoker(v) {
								continue
							}
							outCard = v
							found = true
							newCards := KickOutCards(cardGroup, []uint32{v})
							cardGroups.BaseGroups = append(cardGroups.BaseGroups, newCards)
							break
						}
					} else {
						cardGroups.BaseGroups = append(cardGroups.BaseGroups, cardGroup)
					}
				} else {
					cardGroups.BaseGroups = append(cardGroups.BaseGroups, cardGroup)
				}
			}
			return outCard, cardGroups, true
		}
	}

	// 有两条生命但是不能胡牌,散牌数 >= 2
	var optCardGroups []*ValidCardGroup = make([]*ValidCardGroup, 0) // 散牌牌集的所有可选分组方案
	for _, divide := range secondLifeValidCardGroupList {
		// allBaseGroups, allCookiesCards := splitCookiesCards(cardGroupList)
		// allBaseGroups, allCookiesCards := splitCookiesCards(divide.BaseGroups, WildCard)
		cookiesCardsList := getAllCookiesCards(divide.LeftCards) // 所有散牌牌集可能的分组
		// fmt.Println("--有效牌组", allBaseGroups, "以外的剩余散牌牌集", allCookiesCards)
		// 计算所有cookiesCardsList的成牌概率并取最大值作为首选
		var selectedCookiesCards *CookiesCards
		var maxRate int64 = -1
		for _, cookiesCards := range cookiesCardsList {
			rate := int64(0)
			for i := 0; i < len(divide.BaseGroups); i++ {
				rate += 106 * 106
			}
			for _, lackGroup := range cookiesCards.lackGroups {
				rate += GetCompleteRateForLack(handCards, FaceUpCard, lackGroup) * 106
			}
			for _, singleCard := range cookiesCards.leftCards {
				rate += GetCompleteRateForCard(handCards, FaceUpCard, singleCard, WildCard)
			}
			// fmt.Println("----有效牌组", allBaseGroups, "以外的剩余散牌牌集,可选的分组", "缺一门牌组", cookiesCards.lackGroups, "散牌牌集", cookiesCards.leftCards, "成牌概率", rate)
			if rate > maxRate {
				maxRate = rate
				selectedCookiesCards = cookiesCards
			}
		}
		validCardGroup := newValidCardGroup(divide.BaseGroups, selectedCookiesCards.lackGroups, selectedCookiesCards.leftCards, maxRate)
		optCardGroups = append(optCardGroups, validCardGroup)
		// fmt.Println("--有效牌组", validCardGroup.baseGroups, "以外的剩余散牌牌集,选择的分组", "缺一门牌组", validCardGroup.lackGroups, "散牌牌集", validCardGroup.leftCards, "成牌概率", validCardGroup.rate)
	}

	//  在所有optCardGroups中找出一种成牌概率最高的
	selectedCardGroup := selectValidCardGroup(optCardGroups)
	fmt.Println("选择的分组", "有效牌组", selectedCardGroup.baseGroups, "缺一门牌组", selectedCardGroup.lackGroups, "散牌牌集", selectedCardGroup.leftCards, "成牌概率", selectedCardGroup.rate)
	// 如果除了缺一门组合以外还有散牌牌集,那么打出散牌牌集中成牌概率最低的牌中点数最大的一张牌
	if len(selectedCardGroup.leftCards) > 0 {
		minRate := int64(10000000000)
		maxPoint := uint32(0)
		var outCard uint32 // 成牌概率最低的一个组合
		for _, v := range selectedCardGroup.leftCards {
			rate := GetCompleteRateForCard(handCards, FaceUpCard, v, WildCard)
			// 牌点
			var point uint32 = GetCardPoint(v, WildCard, true)
			// point := card.GetCardPoint(v, WildCard, true)
			if rate < minRate {
				minRate = rate
				outCard = v
				maxPoint = point
			} else if rate == minRate {
				if point > maxPoint {
					minRate = rate
					outCard = v
					maxPoint = point
				}
			}
		}
		cardDivide := &HandCardDivide{BaseGroups: selectedCardGroup.baseGroups}
		cardDivide.BaseGroups = append(cardDivide.BaseGroups, selectedCardGroup.lackGroups...)
		// for _, lackGroup := range selectedCardGroup.lackGroups {
		// 	cardDivide.BaseGroups = append(cardDivide.BaseGroups, lackGroup)
		// }
		newCards := KickOutCards(selectedCardGroup.leftCards, []uint32{outCard})
		if len(newCards) > 0 {
			cardDivide.LeftCards = newCards
		}
		return outCard, cardDivide, false
		// // cardGroups := make([]*card.CardGroup, 0)
		// cardGroups := make([][]uint32, 0)
		// cardGroups = append(cardGroups, selectedCardGroup.baseGroups...)
		// for _, lackGroup := range selectedCardGroup.lackGroups {
		// 	cardGroups = append(cardGroups, append(lackGroup, WildCard))
		// }
		// newCards := KickOutCards(selectedCardGroup.leftCards, []uint32{outCard})
		// // newCards := card.KickTheCardArray(selectedCardGroup.leftCards, outCard)
		// if len(newCards) > 0 {
		// 	cardGroups = append(cardGroups, append(newCards, WildCard))
		// }
		// return outCard, cardGroups, false
	}

	// 如果除了缺一门组合以外没有散牌牌集,找出成牌概率最低的缺一门，打出其中点数最大的牌
	minRate := int64(10000000000)
	var lackGroupIndex int // 成牌概率最低的一个组合
	for k, lackGroup := range selectedCardGroup.lackGroups {
		rate := GetCompleteRateForLack(handCards, FaceUpCard, lackGroup)
		if rate < minRate {
			minRate = rate
			lackGroupIndex = k
		}
	}

	cardDivide := &HandCardDivide{BaseGroups: selectedCardGroup.baseGroups}
	var outCard uint32 // 非癞子的牌面分值最高的一张牌
	for k, lackGroup := range selectedCardGroup.lackGroups {
		if k != lackGroupIndex {
			cardDivide.BaseGroups = append(cardDivide.BaseGroups, lackGroup)
			// cardGroups = append(cardGroups, card.GroupCards(lackGroup, WildCard))
		} else {
			maxScore := uint32(0)
			for _, v := range lackGroup {
				cardScore := GetCardPoint(v, WildCard, true)
				if cardScore > maxScore {
					maxScore = cardScore
					outCard = v
				}
			}
			newCards := KickOutCards(lackGroup, []uint32{outCard})
			// newCards := card.KickTheCardArray(lackGroup, []poker.Card{outCard})
			if len(newCards) > 0 {
				cardDivide.BaseGroups = append(cardDivide.BaseGroups, newCards)
				// cardGroups = append(cardGroups, card.GroupCards(newCards, WildCard))
			}
		}
	}
	fmt.Println("有两条生命时，打出非癞子的牌面分值最高的一张牌")
	return outCard, cardDivide, false // 打出非癞子的牌面分值最高的一张牌

	// cardGroups := make([]*card.CardGroup, 0)
	// cardGroups = append(cardGroups, selectedCardGroup.baseGroups...)
	// var outCard poker.Card // 非癞子的牌面分值最高的一张牌
	// for k, lackGroup := range selectedCardGroup.lackGroups {
	// 	if k != lackGroupIndex {
	// 		cardGroups = append(cardGroups, card.GroupCards(lackGroup, WildCard))
	// 	} else {
	// 		maxScore := int64(-1)
	// 		for _, v := range lackGroup {
	// 			cardScore := card.GetCardPoint(v, WildCard, true)
	// 			if cardScore > maxScore {
	// 				maxScore = cardScore
	// 				outCard = v
	// 			}
	// 		}
	// 		newCards := card.KickTheCardArray(lackGroup, []poker.Card{outCard})
	// 		if len(newCards) > 0 {
	// 			cardGroups = append(cardGroups, card.GroupCards(newCards, WildCard))
	// 		}
	// 	}
	// }
	// fmt.Println("有两条生命时，打出非癞子的牌面分值最高的一张牌")
	// return outCard, cardGroups, false // 打出非癞子的牌面分值最高的一张牌
}

// 散牌牌集allCookiesCards的所有可能分组
func getAllCookiesCards(allCookiesCards []uint32) []*CookiesCards {
	// fmt.Println("剩余散牌牌集", allCookiesCards)
	cookiesCardsList := make([]*CookiesCards, 0)
	if len(allCookiesCards) <= 0 {
		return cookiesCardsList
	}
	allLackGroups := FindAllLackGroup(allCookiesCards) // 分析出所有缺一门组合
	if len(allLackGroups) <= 0 {                       // 无法找到缺一门组合
		// fmt.Println("剩余散牌牌集中无法找到缺一门组合")
		lackGroups := make([][]uint32, 0)
		leftCards := make([]uint32, 0)
		leftCards = append(leftCards, allCookiesCards...)
		cookiesCards := newCookiesCards(lackGroups, leftCards)
		cookiesCardsList = append(cookiesCardsList, cookiesCards)
	} else {
		// fmt.Println("剩余散牌牌集的所有缺一门组合", allLackGroups)
		lackGroupPermutations := make([][]int, 0)
		for groupCount := 1; groupCount <= len(allLackGroups); groupCount++ {
			tempCombinations := GenerateCombination(len(allLackGroups), groupCount) // {{0,1,2,3},{0,1,2,4},{1,2,3,4}}
			lackGroupPermutations = append(lackGroupPermutations, tempCombinations...)
		}
		for _, comnimation := range lackGroupPermutations {
			lackGroups := make([][]uint32, 0)
			lackCards := make([]uint32, 0)
			for _, v := range comnimation {
				lackCards = append(lackCards, allLackGroups[v]...)
				lackGroups = append(lackGroups, allLackGroups[v])
			}
			if !ContainsArray(allCookiesCards, lackCards) { // 缺一门排列不合法
				continue
			}
			leftCards := KickOutCards(allCookiesCards, lackCards) // 除了缺一门组合以外的牌集
			// fmt.Println("有效缺一门组合排列", comnimation, "缺一门", lackGroups, "散牌牌集", leftCards)
			cookiesCardsList = append(cookiesCardsList, newCookiesCards(lackGroups, leftCards))
		}
	}
	return cookiesCardsList
}

// 获取缺一门组合成有效牌组的成牌概率
func GetCompleteRateForLack(handCards []uint32, seenCards []uint32, lackGroup []uint32) int64 {
	// 缺一门组合牌值排序
	sort.Slice(lackGroup, func(x, y int) bool {
		return lackGroup[x] < lackGroup[y]
	})
	// 缺一门组合的类型
	if lackGroup[1]-lackGroup[0] == 1 { // 纯顺子
		// 缺一门组合缺的牌
		if Rank(lackGroup[1]) == Ace {
			return GetCardCountUnUsed(handCards, seenCards, lackGroup[0]-1)
		} else if Rank(lackGroup[0]) == Deuce {
			return GetCardCountUnUsed(handCards, seenCards, lackGroup[1]+1)
		} else {
			return GetCardCountUnUsed(handCards, seenCards, lackGroup[0]-1) + GetCardCountUnUsed(handCards, seenCards, lackGroup[1]+1)
		}
	} else if lackGroup[1]-lackGroup[0] == 2 { // 纯顺子
		// 缺一门组合缺的牌
		return GetCardCountUnUsed(handCards, seenCards, lackGroup[1]-1)
	} else if Rank(lackGroup[1]) == Rank(lackGroup[0]) { // 三条
		lackColor := make([]uint32, 0)
		colors := []uint32{Spade, Heart, Club, Diamond}
		for _, color := range colors {
			if color == Suit(lackGroup[1]) || color == Suit(lackGroup[0]) {
				continue
			}
			lackColor = append(lackColor, color)
		}
		totalCount := int64(0)
		for _, color := range lackColor {
			totalCount += GetCardCountUnUsed(handCards, seenCards, color|Rank(lackGroup[0]))
		}
		return totalCount
	}
	return 0
}

// 获取cardValue未被使用的数量
func GetCardCountUnUsed(handCards []uint32, seenCards []uint32, cardValue uint32) int64 {
	totalRate := int64(2)
	for _, v := range handCards {
		if v == cardValue {
			totalRate -= 1
		}
	}
	for _, v := range seenCards {
		// diffValue := int64(poker.Card(v) - cardValue)
		if v == cardValue {
			totalRate -= 1
		}
	}
	return totalRate
}

// 获取单牌组合成有效牌组的成牌概率
func GetCompleteRateForCard(handCards []uint32, seenCards []uint32, cardValue, wildCard uint32) int64 {
	return GetCompleteRateToLackForCard(handCards, seenCards, cardValue, wildCard) * GetCompleteRateToLackForCard(handCards, seenCards, cardValue, wildCard)
}

// 获取单牌组合成缺一门组合的成牌概率
func GetCompleteRateToLackForCard(handCards []uint32, seenCards []uint32, cardValue, wildCard uint32) int64 {
	totalRate := int64(4 + 6)
	if IsLaizi(cardValue, wildCard) { // 癞子牌最大概率
		return totalRate + 1
	}

	cardRank := Rank(cardValue)
	// 手牌已存在
	for _, v := range handCards {
		diffValue := int64(v - cardValue)
		if diffValue == 1 || diffValue == 2 || diffValue == -1 || diffValue == -2 { // 顺子
			totalRate -= 1
		} else if diffValue != 0 && Rank(v) == cardRank { // 豹子
			totalRate -= 1
		}
		// else if diffValue == 32 || diffValue == 64 || diffValue == 96 { // 豹子
		// 	totalRate -= 1
		// } else if diffValue == -32 || diffValue == -64 || diffValue == -96 { // 豹子
		// 	totalRate -= 1
		// }
	}
	// 弃牌堆已存在
	for _, v := range seenCards {
		diffValue := int64(v - cardValue)
		if diffValue == 1 || diffValue == 2 || diffValue == -1 || diffValue == -2 {
			totalRate -= 1
		} else if diffValue != 0 && Rank(v) == cardRank { // 豹子
			totalRate -= 1
		}
		// else if diffValue == 32 || diffValue == 64 || diffValue == 96 {
		// 	totalRate -= 1
		// } else if diffValue == -32 || diffValue == -64 || diffValue == -96 {
		// 	totalRate -= 1
		// }
	}
	return totalRate
}

// 在所有optCardGroups中找出一种纯顺子成牌概率最高的
func selectValidCardGroup(optCardGroups []*ValidCardGroup) *ValidCardGroup {
	var selectCardGroup *ValidCardGroup
	var maxRate int64 = -1
	for _, validCardGroup := range optCardGroups {
		// fmt.Println("可选分组", "有效牌组", validCardGroup.baseGroups, "缺一门牌组", validCardGroup.lackGroups, "散牌牌集", validCardGroup.leftCards, "成牌概率", validCardGroup.rate)
		if validCardGroup.rate > maxRate {
			maxRate = validCardGroup.rate
			selectCardGroup = validCardGroup
		}
	}
	return selectCardGroup
}

// 找出所有有纯顺子并且没有第二个顺子的
func findAllFirstLifeValidCardGroupList(validCardGroupList []*HandCardDivide, wildCard uint32) []*HandCardDivide {
	firstLifeValidCardGroupList := make([]*HandCardDivide, 0)
	for _, v := range validCardGroupList {
		hasPureSequence := false
		sequenceCount := 0
		for _, group := range v.BaseGroups {
			if IsPureTonghuashun(group) {
				sequenceCount = sequenceCount + 1
				hasPureSequence = true
			} else if IsTonghuashun(group, wildCard) {
				sequenceCount = sequenceCount + 1
			}
			// if group.GroupType == card.CardGroupType_PureSequence {
			// 	sequenceCount = sequenceCount + 1
			// 	hasPureSequence = true
			// } else if group.GroupType == card.CardGroupType_ImpureSequence {
			// 	sequenceCount = sequenceCount + 1
			// }
		}
		if hasPureSequence && sequenceCount < 2 {
			firstLifeValidCardGroupList = append(firstLifeValidCardGroupList, v)
		}
	}
	return firstLifeValidCardGroupList
}

// 有且仅有一条纯顺子
func handleFirstLifeValid(handCards []uint32, validCardGroupList []*HandCardDivide, FaceUpCard []uint32, WildCard uint32) (uint32, *HandCardDivide, bool) {
	fmt.Println("只有第一生命-----------------------------------------------------------------------")
	// validCardGroupList := NewGroupCardMode1().GroupHandCards(handCards, WildCard, time.Now().Unix())
	// firstLifeValidCardGroupList := findAllFirstLifeValidCardGroupList(validCardGroupList)
	// firstLifeValidCardGroupList = HandleGroupAppend(handCards, WildCard, firstLifeValidCardGroupList)
	validCardGroupList = HandleGroupAppend(handCards, WildCard, validCardGroupList)
	firstLifeValidCardGroupList := findAllFirstLifeValidCardGroupList(validCardGroupList, WildCard)

	// 查找set数量
	zeroSetValidCardGroupList := make([]*HandCardDivide, 0)
	oneSetValidCardGroupList := make([]*HandCardDivide, 0)
	twoSetValidCardGroupList := make([]*HandCardDivide, 0)
	threeSetValidCardGroupList := make([]*HandCardDivide, 0)
	for _, cardGroupList := range firstLifeValidCardGroupList {
		TidyDivideLeftCard(cardGroupList, WildCard)
		setCount := 0
		for _, v := range cardGroupList.BaseGroups {
			if IsBaozi(v, WildCard) {
				setCount += 1
			}
			// if v.GroupType == card.CardGroupType_ThreeSet {
			// 	setCount += 1
			// }
		}
		if setCount == 0 {
			zeroSetValidCardGroupList = append(zeroSetValidCardGroupList, cardGroupList)
		} else if setCount == 1 {
			oneSetValidCardGroupList = append(oneSetValidCardGroupList, cardGroupList)
		} else if setCount == 2 {
			twoSetValidCardGroupList = append(twoSetValidCardGroupList, cardGroupList)
		} else if setCount == 3 {
			threeSetValidCardGroupList = append(threeSetValidCardGroupList, cardGroupList)
		}
	}
	if len(twoSetValidCardGroupList) > 0 { // 从散牌中出牌
		return outCardFromCookies(handCards, FaceUpCard, WildCard, twoSetValidCardGroupList)
	}
	if len(oneSetValidCardGroupList) > 0 { // 从散牌中出牌
		return outCardFromCookies(handCards, FaceUpCard, WildCard, oneSetValidCardGroupList)
	}
	if len(threeSetValidCardGroupList) > 0 { // 从set中拆牌出牌
		return outCardFromCookies(handCards, FaceUpCard, WildCard, threeSetValidCardGroupList) //outCardFromSet
	}
	if len(zeroSetValidCardGroupList) > 0 { // 从散牌中出牌
		return outCardFromCookies(handCards, FaceUpCard, WildCard, zeroSetValidCardGroupList)
	}
	return outCardFromCookies(handCards, FaceUpCard, WildCard, firstLifeValidCardGroupList)
}

// 有且仅有一条纯顺子时从散牌中出牌
func outCardFromCookies(handCards []uint32, FaceUpCard []uint32, WildCard uint32, lifeValidCardGroupList []*HandCardDivide) (uint32, *HandCardDivide, bool) {

	// for _, cardGroupList := range lifeValidCardGroupList {
	// 	// 1 计算剩余散牌组成的顺子缺一门数量
	// 	allBaseGroups, allCookiesCards := splitCookiesCards(cardGroupList)
	// 	// 2 如果散牌中的顺子缺一门数量 <= 0
	// 	// 		2.1 拆三条
	// 	// 3 如果散牌中的顺子缺一门数量 > 0
	// }

	////////////////////////////////
	// 只有一张散牌
	for _, cardGroupList := range lifeValidCardGroupList {
		// allBaseGroups, allCookiesCards := splitCookiesCards(cardGroupList.BaseGroups, WildCard)
		// if len(allCookiesCards) == 1 {
		// 	return allCookiesCards[0], &HandCardDivide{BaseGroups: allBaseGroups}, false
		// }
		// tidyDivideLeftCard(cardGroupList, WildCard)
		if len(cardGroupList.LeftCards) == 1 {
			return cardGroupList.LeftCards[0], &HandCardDivide{BaseGroups: cardGroupList.BaseGroups}, false
		}
	}

	// 没有散牌
	for _, cardGroupList := range lifeValidCardGroupList {
		// allBaseGroups, allCookiesCards := splitCookiesCards(cardGroupList)
		if len(cardGroupList.LeftCards) == 0 {
			found := false
			var outCard uint32
			cardGroups := &HandCardDivide{}
			// cardGroups := make([]*card.CardGroup, 0)
			for _, cardGroup := range cardGroupList.BaseGroups {
				if !found {
					if IsPureTonghuashun(cardGroup) && len(cardGroup) >= 4 {
						for _, v := range cardGroup {
							if IsCardWild(v, WildCard) || IsCardJoker(v) {
								continue
							}
							outCard = v
							found = true
							newCards := KickOutCards(cardGroup, []uint32{v})
							cardGroups.BaseGroups = append(cardGroups.BaseGroups, newCards)
							break
						}
					} else {
						cardGroups.BaseGroups = append(cardGroups.BaseGroups, cardGroup)
					}
				} else {
					cardGroups.BaseGroups = append(cardGroups.BaseGroups, cardGroup)
				}
			}
			if found {
				return outCard, cardGroups, false
			}
			found = false
			cardGroups = &HandCardDivide{}
			// cardGroups = make([]*card.CardGroup, 0)
			for _, cardGroup := range cardGroupList.BaseGroups {
				if !found {
					if IsPureTonghuashun(cardGroup) {
						for _, v := range cardGroup {
							if IsCardWild(v, WildCard) || IsCardJoker(v) {
								continue
							}
							outCard = v
							found = true
							newCards := KickOutCards(cardGroup, []uint32{v})
							cardGroups.BaseGroups = append(cardGroups.BaseGroups, newCards)
							break
						}
					} else {
						cardGroups.BaseGroups = append(cardGroups.BaseGroups, cardGroup)
					}
				} else {
					cardGroups.BaseGroups = append(cardGroups.BaseGroups, cardGroup)
				}
			}
			return outCard, cardGroups, false
		}
	}
	////////////////////////////////
	// 只有第一生命，散牌数 >= 2
	var optCardGroups []*ValidCardGroup = make([]*ValidCardGroup, 0) // 散牌牌集的所有可选分组方案
	for _, cardGroupList := range lifeValidCardGroupList {
		// allBaseGroups, allCookiesCards := splitCookiesCards(cardGroupList)
		cookiesCardsList := getAllCookiesCards(cardGroupList.LeftCards)
		// fmt.Println("--有效牌组", allBaseGroups, "以外的剩余散牌牌集", allCookiesCards)
		// 计算所有cookiesCardsList的成牌概率并取最大值作为首选
		var validCardGroup *ValidCardGroup
		var maxRate int64 = -1
		for _, cookiesCards := range cookiesCardsList {
			rate := int64(0)
			for i := 0; i < len(cardGroupList.BaseGroups); i++ {
				rate += 106 * 106
			}
			for _, lackGroup := range cookiesCards.lackGroups {
				rate += GetCompleteRateToSequenceForLack(handCards, FaceUpCard, lackGroup, WildCard) * 106
			}
			for _, singleCard := range cookiesCards.leftCards {
				rate += GetCompleteRateToSequenceForCard(handCards, FaceUpCard, singleCard, WildCard)
			}
			if rate > maxRate {
				maxRate = rate
				validCardGroup = newValidCardGroup(cardGroupList.BaseGroups, cookiesCards.lackGroups, cookiesCards.leftCards, rate)
			}
		}
		optCardGroups = append(optCardGroups, validCardGroup)
	}
	//  在所有optCardGroups中找出一种成牌概率最高的
	selectedCardGroup := selectValidCardGroup(optCardGroups)
	fmt.Println("选择的分组", "有效牌组", selectedCardGroup.baseGroups, "缺一门牌组", selectedCardGroup.lackGroups, "散牌牌集", selectedCardGroup.leftCards, "成牌概率", selectedCardGroup.rate)
	minRate := int64(10000000000)

	// 如果除了缺一门组合以外还有散牌牌集,那么打出散牌牌集中成牌概率最低的牌中点数最大的一张牌
	if len(selectedCardGroup.leftCards) > 0 {
		minRate := int64(10000000000)
		maxPoint := uint32(0)
		var outCard uint32 // 成牌概率最低的一个组合
		for _, v := range selectedCardGroup.leftCards {
			rate := GetCompleteRateToSequenceForCard(handCards, FaceUpCard, v, WildCard)
			point := GetCardPoint(v, WildCard, true)
			if rate < minRate {
				minRate = rate
				outCard = v
				maxPoint = point
			} else if rate == minRate {
				if point > maxPoint {
					minRate = rate
					outCard = v
					maxPoint = point
				}
			}
		}
		// cardGroups := make([]*card.CardGroup, 0)
		cardGroups := &HandCardDivide{BaseGroups: selectedCardGroup.baseGroups}
		cardGroups.BaseGroups = append(cardGroups.BaseGroups, selectedCardGroup.lackGroups...)
		// for _, lackGroup := range selectedCardGroup.lackGroups {
		// 	cardGroups.BaseGroups = append(cardGroups.BaseGroups, lackGroup)
		// }
		newCards := KickOutCards(selectedCardGroup.leftCards, []uint32{outCard})
		// newCards := card.KickTheCardArray(selectedCardGroup.leftCards, []poker.Card{outCard})
		if len(newCards) > 0 {
			cardGroups.BaseGroups = append(cardGroups.BaseGroups, newCards)
		}
		return outCard, cardGroups, false
	}
	// 如果除了缺一门组合以外没有散牌牌集,找出成牌概率最低的缺一门，打出其中点数最大的牌
	var lackGroupIndex int // 成牌概率最低的一个组合
	for k, v := range selectedCardGroup.lackGroups {
		rate := GetCompleteRateToSequenceForLack(handCards, FaceUpCard, v, WildCard)
		if rate < minRate {
			minRate = rate
			lackGroupIndex = k
		}
	}

	cardGroups := &HandCardDivide{BaseGroups: selectedCardGroup.baseGroups}
	var outCard uint32 // 非癞子的牌面分值最高的一张牌
	for k, lackGroup := range selectedCardGroup.lackGroups {
		if k != lackGroupIndex {
			cardGroups.BaseGroups = append(cardGroups.BaseGroups, lackGroup)
		} else {
			maxScore := uint32(0)
			for _, v := range lackGroup {
				// cardScore := card.GetCardPoint(v, WildCard, true)
				cardScore := GetCardPoint(v, WildCard, true)
				if cardScore > maxScore {
					maxScore = cardScore
					outCard = v
				}
			}
			newCards := KickOutCards(lackGroup, []uint32{outCard})
			// newCards := card.KickTheCardArray(lackGroup, []poker.Card{outCard})
			if len(newCards) > 0 {
				cardGroups.BaseGroups = append(cardGroups.BaseGroups, newCards)
			}
		}
	}
	return outCard, cardGroups, false // 打出非癞子的牌面分值最高的一张牌
}

// 获取缺一门组合成顺子的成牌概率
func GetCompleteRateToSequenceForLack(handCards []uint32, seenCards []uint32, lackGroup []uint32, wilCard uint32) int64 {
	wildCount := 0
	if IsCardWild(lackGroup[0], wilCard) || IsCardJoker(lackGroup[0]) {
		wildCount += 1
	}
	if IsCardWild(lackGroup[1], wilCard) || IsCardJoker(lackGroup[1]) {
		wildCount += 1
	}
	if wildCount == 2 {
		return 106
	} else if wildCount == 1 {
		cardValue := lackGroup[0]
		if IsCardWild(cardValue, wilCard) || IsCardJoker(cardValue) {
			cardValue = lackGroup[1]
		}
		rank, suit := Rank(cardValue), Suit(cardValue)
		if rank == Ace {
			rate := int64(0)
			// 23QK
			passableCards := []uint32{suit | Deuce, suit | Trey, suit | Queen, suit | King, wilCard}
			for _, card := range passableCards {
				rate += GetCardCountUnUsed(handCards, seenCards, card)
			}

			return rate
		} else if rank == Deuce {
			rate := int64(0)
			// 134
			passableCards := []uint32{suit | Ace, suit | Trey, suit | Four, wilCard}
			for _, card := range passableCards {
				rate += GetCardCountUnUsed(handCards, seenCards, card)
			}
			return rate
		} else if rank == Trey {
			rate := int64(0)
			// A245
			passableCards := []uint32{suit | Ace, suit | Deuce, suit | Four, suit | Five, wilCard}
			for _, card := range passableCards {
				rate += GetCardCountUnUsed(handCards, seenCards, card)
			}
			return rate
		} else if rank == Queen {
			rate := int64(0)
			// 10 JKA
			passableCards := []uint32{suit | Ten, suit | Jack, suit | Queen, suit | Ace, wilCard}
			for _, card := range passableCards {
				rate += GetCardCountUnUsed(handCards, seenCards, card)
			}
			return rate
		} else if rank == King {
			rate := int64(0)
			// AJQ
			passableCards := []uint32{suit | Ace, suit | Jack, suit | Queen, wilCard}
			for _, card := range passableCards {
				rate += GetCardCountUnUsed(handCards, seenCards, card)
			}
			return rate
		} else {
			rate := int64(0)
			card2 := cardValue - 2
			card3 := cardValue - 1
			cardQ := cardValue + 1
			cardK := cardValue + 2
			rate += GetCardCountUnUsed(handCards, seenCards, card2)
			rate += GetCardCountUnUsed(handCards, seenCards, card3)
			rate += GetCardCountUnUsed(handCards, seenCards, cardQ)
			rate += GetCardCountUnUsed(handCards, seenCards, cardK)
			rate += GetCardCountUnUsed(handCards, seenCards, wilCard)
			return rate
		}
	} else {
		// 缺一门组合牌值排序
		sort.Slice(lackGroup, func(x, y int) bool {
			return lackGroup[x] < lackGroup[y]
		})
		// 缺一门组合的类型
		if lackGroup[1]-lackGroup[0] == 1 {
			if Rank(lackGroup[1]) == King {
				rate := int64(0)
				rate += GetCardCountUnUsed(handCards, seenCards, lackGroup[0]-1)
				rate += GetCardCountUnUsed(handCards, seenCards, wilCard)
				return rate
			} else if Rank(lackGroup[0]) == Ace {
				rate := int64(0)
				rate += GetCardCountUnUsed(handCards, seenCards, lackGroup[1]+1)
				rate += GetCardCountUnUsed(handCards, seenCards, wilCard)
				return rate
			} else {
				rate := int64(0)
				rate += GetCardCountUnUsed(handCards, seenCards, lackGroup[0]-1)
				rate += GetCardCountUnUsed(handCards, seenCards, lackGroup[1]+1)
				rate += GetCardCountUnUsed(handCards, seenCards, wilCard)
				return rate
			}
		} else if lackGroup[1]-lackGroup[0] == 2 {
			rate := int64(0)
			rate += GetCardCountUnUsed(handCards, seenCards, lackGroup[1]-1)
			rate += GetCardCountUnUsed(handCards, seenCards, wilCard)
			return rate
		} else if Rank(lackGroup[1]) == King && Rank(lackGroup[0]) == Ace {
			rate := int64(0)
			rate += GetCardCountUnUsed(handCards, seenCards, lackGroup[1]-1)
			rate += GetCardCountUnUsed(handCards, seenCards, wilCard)
			return rate
		}
	}
	return 0
}

// 获取单牌组合成顺子的成牌概率
func GetCompleteRateToSequenceForCard(handCards []uint32, seenCards []uint32, cardValue uint32, wilCard uint32) int64 {
	rank, suit := Rank(cardValue), Suit(cardValue)
	if rank == Deuce {
		rate := int64(0)
		cardA := suit | Ace
		card3 := suit | Trey
		card4 := suit | Four
		rate += GetCardCountUnUsed(handCards, seenCards, cardA) * GetCardCountUnUsed(handCards, seenCards, card3)
		rate += GetCardCountUnUsed(handCards, seenCards, card3) * GetCardCountUnUsed(handCards, seenCards, card4)

		rate += GetCardCountUnUsed(handCards, seenCards, wilCard) * GetCardCountUnUsed(handCards, seenCards, card3)
		rate += GetCardCountUnUsed(handCards, seenCards, cardA) * GetCardCountUnUsed(handCards, seenCards, wilCard)

		rate += GetCardCountUnUsed(handCards, seenCards, wilCard) * GetCardCountUnUsed(handCards, seenCards, card4)
		rate += GetCardCountUnUsed(handCards, seenCards, card3) * GetCardCountUnUsed(handCards, seenCards, wilCard)
		return rate
	} else if rank == Queen {
		rate := int64(0)
		cardA := suit | Ace
		cardK := suit | King
		cardJ := suit | Jack
		card10 := suit | Ten

		rate += GetCardCountUnUsed(handCards, seenCards, cardA) * GetCardCountUnUsed(handCards, seenCards, cardK)
		rate += GetCardCountUnUsed(handCards, seenCards, cardK) * GetCardCountUnUsed(handCards, seenCards, cardJ)
		rate += GetCardCountUnUsed(handCards, seenCards, cardJ) * GetCardCountUnUsed(handCards, seenCards, card10)

		rate += GetCardCountUnUsed(handCards, seenCards, wilCard) * GetCardCountUnUsed(handCards, seenCards, cardK)
		rate += GetCardCountUnUsed(handCards, seenCards, cardA) * GetCardCountUnUsed(handCards, seenCards, wilCard)

		rate += GetCardCountUnUsed(handCards, seenCards, wilCard) * GetCardCountUnUsed(handCards, seenCards, cardJ)
		rate += GetCardCountUnUsed(handCards, seenCards, cardK) * GetCardCountUnUsed(handCards, seenCards, wilCard)

		rate += GetCardCountUnUsed(handCards, seenCards, wilCard) * GetCardCountUnUsed(handCards, seenCards, card10)
		rate += GetCardCountUnUsed(handCards, seenCards, cardJ) * GetCardCountUnUsed(handCards, seenCards, wilCard)
		return rate
	} else if rank == Ace {
		rate := int64(0)
		card2 := suit | Deuce
		card3 := suit | Trey
		cardQ := suit | Queen
		cardK := suit | King
		rate += GetCardCountUnUsed(handCards, seenCards, card2) * GetCardCountUnUsed(handCards, seenCards, card3)
		rate += GetCardCountUnUsed(handCards, seenCards, cardQ) * GetCardCountUnUsed(handCards, seenCards, cardK)

		rate += GetCardCountUnUsed(handCards, seenCards, wilCard) * GetCardCountUnUsed(handCards, seenCards, card3)
		rate += GetCardCountUnUsed(handCards, seenCards, card2) * GetCardCountUnUsed(handCards, seenCards, wilCard)

		rate += GetCardCountUnUsed(handCards, seenCards, wilCard) * GetCardCountUnUsed(handCards, seenCards, cardK)
		rate += GetCardCountUnUsed(handCards, seenCards, cardQ) * GetCardCountUnUsed(handCards, seenCards, wilCard)
		return rate
	} else if rank == King {
		rate := int64(0)
		cardA := suit | Ace
		cardJ := suit | Jack
		cardQ := suit | Queen
		rate += GetCardCountUnUsed(handCards, seenCards, cardJ) * GetCardCountUnUsed(handCards, seenCards, cardQ)
		rate += GetCardCountUnUsed(handCards, seenCards, cardQ) * GetCardCountUnUsed(handCards, seenCards, cardA)

		rate += GetCardCountUnUsed(handCards, seenCards, wilCard) * GetCardCountUnUsed(handCards, seenCards, cardQ)
		rate += GetCardCountUnUsed(handCards, seenCards, cardJ) * GetCardCountUnUsed(handCards, seenCards, wilCard)

		rate += GetCardCountUnUsed(handCards, seenCards, wilCard) * GetCardCountUnUsed(handCards, seenCards, cardA)
		rate += GetCardCountUnUsed(handCards, seenCards, cardQ) * GetCardCountUnUsed(handCards, seenCards, wilCard)
		return rate
	}
	rate := int64(0)
	card2 := suit | (rank - 2)
	card3 := suit | (rank - 1)
	cardQ := suit | (rank + 1)
	cardK := suit | (rank + 2)
	rate += GetCardCountUnUsed(handCards, seenCards, card2) * GetCardCountUnUsed(handCards, seenCards, card3)
	rate += GetCardCountUnUsed(handCards, seenCards, card3) * GetCardCountUnUsed(handCards, seenCards, cardQ)
	rate += GetCardCountUnUsed(handCards, seenCards, cardQ) * GetCardCountUnUsed(handCards, seenCards, cardK)

	rate += GetCardCountUnUsed(handCards, seenCards, wilCard) * GetCardCountUnUsed(handCards, seenCards, card3)
	rate += GetCardCountUnUsed(handCards, seenCards, card2) * GetCardCountUnUsed(handCards, seenCards, wilCard)

	rate += GetCardCountUnUsed(handCards, seenCards, wilCard) * GetCardCountUnUsed(handCards, seenCards, cardQ)
	rate += GetCardCountUnUsed(handCards, seenCards, card3) * GetCardCountUnUsed(handCards, seenCards, wilCard)

	rate += GetCardCountUnUsed(handCards, seenCards, wilCard) * GetCardCountUnUsed(handCards, seenCards, cardK)
	rate += GetCardCountUnUsed(handCards, seenCards, cardQ) * GetCardCountUnUsed(handCards, seenCards, wilCard)
	return rate
}

// 没有纯顺子
func handleNoneLifeValid(handCards []uint32, validCardGroupList []*HandCardDivide, FaceUpCard []uint32, WildCard uint32) (uint32, *HandCardDivide, bool) {
	fmt.Println("没有第一生命-----------------------------------------------------------------------")
	// validCardGroupList := NewGroupCardMode1().GroupHandCards(handCards, WildCard, time.Now().Unix())
	validCardGroupList = HandleGroupAppend(handCards, WildCard, validCardGroupList)

	var optCardGroups []*ValidCardGroup = make([]*ValidCardGroup, 0) // 散牌牌集的所有可选分组方案
	for _, cardGroupList := range validCardGroupList {
		TidyDivideLeftCard(cardGroupList, WildCard)
		// allBaseGroups, allCookiesCards := splitCookiesCards(cardGroupList)
		if len(cardGroupList.LeftCards) > 0 {
			cookiesCardsList := getAllCookiesCards(cardGroupList.LeftCards)
			// fmt.Println("--有效牌组", allBaseGroups, "以外的剩余散牌牌集", allCookiesCards)
			//  计算所有cookiesCardsList的纯顺子成牌概率并取最大值作为首选
			var selectedCookiesCards *CookiesCards
			var maxRate int64 = -1
			for _, cookiesCards := range cookiesCardsList {
				rate := int64(0)
				for i := 0; i < len(cardGroupList.BaseGroups); i++ {
					rate += 106 * 106
				}
				for _, lackGroup := range cookiesCards.lackGroups {
					rate += GetCompleteRateToPureSequenceForLack(handCards, FaceUpCard, lackGroup) * 106
				}
				for _, singleCard := range cookiesCards.leftCards {
					rate += GetCompleteRateToPureSequenceForCard(handCards, FaceUpCard, singleCard)
				}
				// fmt.Println("----有效牌组", allBaseGroups, "以外的剩余散牌牌集,可选的分组", "缺一门牌组", cookiesCards.lackGroups, "散牌牌集", cookiesCards.leftCards, "成牌概率", rate)
				if rate > maxRate {
					maxRate = rate
					selectedCookiesCards = cookiesCards
				}
			}
			validCardGroup := newValidCardGroup(cardGroupList.BaseGroups, selectedCookiesCards.lackGroups, selectedCookiesCards.leftCards, maxRate)
			optCardGroups = append(optCardGroups, validCardGroup)
			// fmt.Println("--有效牌组", validCardGroup.baseGroups, "以外的剩余散牌牌集,选择的分组", "缺一门牌组", validCardGroup.lackGroups, "散牌牌集", validCardGroup.leftCards, "成牌概率", validCardGroup.rate)
		} else {
			rate := int64(0)
			for i := 0; i < len(cardGroupList.BaseGroups); i++ {
				rate += 106 * 106
			}
			validCardGroup := newValidCardGroup(cardGroupList.BaseGroups, [][]uint32{}, []uint32{}, rate)
			optCardGroups = append(optCardGroups, validCardGroup)
			// fmt.Println("--有效牌组", validCardGroup.baseGroups, "以外的剩余散牌牌集,选择的分组", "缺一门牌组", validCardGroup.lackGroups, "散牌牌集", validCardGroup.leftCards, "成牌概率", validCardGroup.rate)
		}
	}

	// 在所有optCardGroups中找出一种纯顺子成牌概率最高的
	selectedCardGroup := selectValidCardGroup(optCardGroups)
	// fmt.Println("选择的分组", "有效牌组", selectedCardGroup.baseGroups, "缺一门牌组", selectedCardGroup.lackGroups, "散牌牌集", selectedCardGroup.leftCards, "成牌概率", selectedCardGroup.rate)
	// 如果除了缺一门组合以外还有散牌牌集,那么打出散牌牌集中成牌概率最低的牌中点数最大的一张牌
	if len(selectedCardGroup.leftCards) > 0 {
		fmt.Println("selectCardGroup", "除了缺一门组合以外还有散牌牌集")
		minRate := int64(10000000000)
		maxPoint := uint32(0)
		var outCard uint32 // 成牌概率最低的一个组合
		fmt.Println("散牌:", CardsString(selectedCardGroup.leftCards))
		for _, v := range selectedCardGroup.leftCards {
			rate := GetCompleteRateToPureSequenceForCard(handCards, FaceUpCard, v)
			// point := card.GetCardPoint(v, WildCard, true)
			point := GetCardPoint(v, WildCard, true)
			fmt.Println("散牌", v, "纯顺子成牌概率", rate)
			if rate < minRate {
				minRate = rate
				outCard = v
				maxPoint = point
			} else if rate == minRate {
				if point > maxPoint {
					minRate = rate
					outCard = v
					maxPoint = point
				}
			}
		}
		fmt.Println("没有第一生命时，打出散牌牌集中成牌概率最低的一张牌")
		cardGroups := &HandCardDivide{BaseGroups: selectedCardGroup.baseGroups}
		cardGroups.BaseGroups = append(cardGroups.BaseGroups, selectedCardGroup.lackGroups...)
		// for _, lackGroup := range selectedCardGroup.lackGroups {
		// 	cardGroups.BaseGroups = append(cardGroups.BaseGroups, lackGroup)
		// }
		newCards := KickOutCards(selectedCardGroup.leftCards, []uint32{outCard})
		// newCards := card.KickTheCardArray(selectedCardGroup.leftCards, []poker.Card{outCard})
		if len(newCards) > 0 {
			cardGroups.BaseGroups = append(cardGroups.BaseGroups, newCards)
		}
		return outCard, cardGroups, false
	}
	// 如果除了缺一门组合以外没有散牌牌集,找出成牌概率最低的缺一门，打出其中点数最大的牌
	if len(selectedCardGroup.lackGroups) > 0 {
		minRate := int64(10000000000)
		var lackGroupIndex int // 成牌概率最低的一个组合
		for k, v := range selectedCardGroup.lackGroups {
			rate := GetCompleteRateToPureSequenceForLack(handCards, FaceUpCard, v)
			if rate < minRate {
				minRate = rate
				lackGroupIndex = k
			}
		}
		cardGroups := &HandCardDivide{BaseGroups: selectedCardGroup.baseGroups}
		// cardGroups := make([]*card.CardGroup, 0)
		// cardGroups = append(cardGroups, selectedCardGroup.baseGroups...)
		var outCard uint32 // 非癞子的牌面分值最高的一张牌
		for k, lackGroup := range selectedCardGroup.lackGroups {
			if k != lackGroupIndex {
				cardGroups.BaseGroups = append(cardGroups.BaseGroups, lackGroup)
			} else {
				maxScore := uint32(0)
				for _, v := range lackGroup {
					cardScore := GetCardPoint(v, WildCard, true)
					if cardScore > maxScore {
						maxScore = cardScore
						outCard = v
					}
				}
				newCards := KickOutCards(lackGroup, []uint32{outCard})
				if len(newCards) > 0 {
					cardGroups.BaseGroups = append(cardGroups.BaseGroups, newCards)
				}
			}
		}
		fmt.Println("没有第一生命时，打出缺一门中非癞子的牌面分值最高的一张牌")
		return outCard, cardGroups, false // 打出非癞子的牌面分值最高的一张牌
	}
	// 如果除了缺一门组合以外没有缺一门,找出4顺，打出其中最靠边的
	found := false
	var outCard uint32
	cardGroups := &HandCardDivide{}
	// cardGroups := make([]*card.CardGroup, 0)
	for _, cardGroup := range selectedCardGroup.baseGroups {
		if !found {
			if len(cardGroup) >= 4 {
				for _, v := range cardGroup {
					if IsCardWild(v, WildCard) || IsCardJoker(v) {
						continue
					}
					outCard = v
					found = true
					newCards := KickOutCards(cardGroup, []uint32{v})
					cardGroups.BaseGroups = append(cardGroups.BaseGroups, newCards)
					break
				}
			} else {
				cardGroups.BaseGroups = append(cardGroups.BaseGroups, cardGroup)
			}
		} else {
			cardGroups.BaseGroups = append(cardGroups.BaseGroups, cardGroup)
		}
	}
	fmt.Println("没有第一生命时，打出非纯顺子中非癞子的牌面分值最高的一张牌")
	return outCard, cardGroups, false
}

// 获取缺一门组合成纯顺子的成牌概率
func GetCompleteRateToPureSequenceForLack(handCards []uint32, seenCards []uint32, lackGroup []uint32) int64 {
	// 缺一门组合牌值排序
	sort.Slice(lackGroup, func(x, y int) bool {
		return lackGroup[x] < lackGroup[y]
	})
	// 缺一门组合的类型
	if lackGroup[1]-lackGroup[0] == 1 { // 纯顺子
		if Rank(lackGroup[1]) == King {
			return GetCardCountUnUsed(handCards, seenCards, lackGroup[0]-1)
		} else if Rank(lackGroup[0]) == Ace {
			return GetCardCountUnUsed(handCards, seenCards, lackGroup[1]+1)
		} else {
			return GetCardCountUnUsed(handCards, seenCards, lackGroup[0]-1) + GetCardCountUnUsed(handCards, seenCards, lackGroup[1]+1)
		}
	} else if lackGroup[1]-lackGroup[0] == 2 { // 纯顺子
		return GetCardCountUnUsed(handCards, seenCards, lackGroup[1]-1)
	} else if Rank(lackGroup[1]) == King && Rank(lackGroup[0]) == Ace { // 纯顺子
		return GetCardCountUnUsed(handCards, seenCards, lackGroup[1]-1)
	}
	return 0
}

// 获取单牌组合成纯顺子的成牌概率
func GetCompleteRateToPureSequenceForCard(handCards []uint32, seenCards []uint32, cardValue uint32) int64 {
	rank, suit := Rank(cardValue), Suit(cardValue)
	if rank == Deuce {
		rate := int64(0)
		cardA := suit | Ace
		card3 := suit | Trey
		card4 := suit | Four
		rate += GetCardCountUnUsed(handCards, seenCards, cardA) * GetCardCountUnUsed(handCards, seenCards, card3)
		rate += GetCardCountUnUsed(handCards, seenCards, card3) * GetCardCountUnUsed(handCards, seenCards, card4)
		return rate
	} else if rank == Queen {
		rate := int64(0)
		cardA := suit | Ace
		carK := suit | King
		cardJ := suit | Jack
		card10 := suit | Ten
		rate += GetCardCountUnUsed(handCards, seenCards, cardA) * GetCardCountUnUsed(handCards, seenCards, carK)
		rate += GetCardCountUnUsed(handCards, seenCards, carK) * GetCardCountUnUsed(handCards, seenCards, cardJ)
		rate += GetCardCountUnUsed(handCards, seenCards, cardJ) * GetCardCountUnUsed(handCards, seenCards, card10)
		return rate
	} else if rank == Ace {
		rate := int64(0)
		card2 := suit | Deuce
		card3 := suit | Trey
		cardQ := suit | Queen
		cardK := suit | King
		rate += GetCardCountUnUsed(handCards, seenCards, card2) * GetCardCountUnUsed(handCards, seenCards, card3)
		rate += GetCardCountUnUsed(handCards, seenCards, cardQ) * GetCardCountUnUsed(handCards, seenCards, cardK)
		return rate
	} else if rank == King {
		rate := int64(0)
		cardA := suit | Ace
		cardJ := suit | Jack
		cardQ := suit | Queen
		rate += GetCardCountUnUsed(handCards, seenCards, cardJ) * GetCardCountUnUsed(handCards, seenCards, cardQ)
		rate += GetCardCountUnUsed(handCards, seenCards, cardQ) * GetCardCountUnUsed(handCards, seenCards, cardA)
		return rate
	}
	rate := int64(0)
	card2 := suit | (rank - 2)
	card3 := suit | (rank - 1)
	cardQ := suit | (rank + 1)
	cardK := suit | (rank + 2)
	rate += GetCardCountUnUsed(handCards, seenCards, card2) * GetCardCountUnUsed(handCards, seenCards, card3)
	rate += GetCardCountUnUsed(handCards, seenCards, card3) * GetCardCountUnUsed(handCards, seenCards, cardQ)
	rate += GetCardCountUnUsed(handCards, seenCards, cardQ) * GetCardCountUnUsed(handCards, seenCards, cardK)
	return rate
}

type CookiesCards struct {
	lackGroups [][]uint32
	leftCards  []uint32
}

func newCookiesCards(lackGroups [][]uint32, leftCards []uint32) *CookiesCards {
	instance := new(CookiesCards)
	instance.lackGroups = lackGroups
	instance.leftCards = leftCards
	return instance
}

type ValidCardGroup struct {
	baseGroups [][]uint32
	lackGroups [][]uint32
	leftCards  []uint32
	rate       int64
}

func newValidCardGroup(baseGroups [][]uint32, lackGroups [][]uint32, leftCards []uint32, rate int64) *ValidCardGroup {
	instance := new(ValidCardGroup)
	instance.baseGroups = baseGroups
	instance.lackGroups = lackGroups
	instance.leftCards = leftCards
	instance.rate = rate
	return instance
}
