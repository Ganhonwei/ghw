package algo

import (
	"fmt"
)

/*玩家摸牌,1摸牌或者2吃牌*/
func GetTouchType(mCards []uint32, FaceUpCard []uint32, WildCard uint32) int32 {
	latestCard := FaceUpCard[len(FaceUpCard)-1]
	// 如果翻拍是癞子牌，则只能摸牌不能吃牌
	if IsCardWild(latestCard, WildCard) || IsCardJoker(latestCard) {
		// loggerx.RummyLog("翻拍是癞子牌，不能吃牌只能摸牌")
		fmt.Println("翻拍是癞子牌，不能吃牌只能摸牌")
		return 1
	}
	handCards := make([]uint32, 0)
	handCards = append(handCards, mCards...)
	// 1 判断吃牌是否对我有效
	handCardsAfterEat := make([]uint32, 0)
	handCardsAfterEat = append(handCardsAfterEat, mCards...)
	handCardsAfterEat = append(handCardsAfterEat, latestCard)
	// 1.1 吃牌是否胡牌
	// groupId := time.Now().Unix()

	handCardDivides := FindAllHandCardDivide(handCardsAfterEat, WildCard, 4)
	validCardGroupList := HandleGroupAppend(handCardsAfterEat, WildCard, handCardDivides)
	// validCardGroupList := NewGroupCardMode1().GroupHandCards(handCardsAfterEat, WildCard, groupId)

	if len(validCardGroupList) <= 0 {
		panic(fmt.Errorf("invalid validCardGroupList,cardArray=%v,wildCard=%v,count=%v", handCardsAfterEat, WildCard, len(validCardGroupList)))
	}
	// 找出所有有纯顺子并且有第二个顺子的
	lifeValidCardGroupList := make([]*HandCardDivide, 0)
	for _, v := range validCardGroupList {
		hasPureSequence := false
		sequenceCount := 0
		for _, group := range v.BaseGroups {
			if IsPureTonghuashun(group) {
				sequenceCount = sequenceCount + 1
				hasPureSequence = true
			} else if IsTonghuashun(group, WildCard) {
				sequenceCount = sequenceCount + 1
			}
		}
		if hasPureSequence && sequenceCount >= 2 {
			lifeValidCardGroupList = append(lifeValidCardGroupList, v)
		}
	}
	if len(lifeValidCardGroupList) > 0 {
		for _, cardGroupList := range lifeValidCardGroupList {
			allCookiesCards := make([]uint32, 0)
			for _, group := range cardGroupList.BaseGroups {
				if !ValidRmGroup(group, WildCard) {
					allCookiesCards = append(allCookiesCards, group...)
				}
			}
			if len(allCookiesCards) == 1 { // 可以胡牌了
				fmt.Println("2可以胡牌了")
				return 2
			}
			if len(allCookiesCards) == 0 { // 没有散牌,可以胡牌了
				fmt.Println("2没有散牌,可以胡牌了")
				return 2
			}
		}
	}
	fmt.Println("吃牌不能胡牌")
	// 1.2 手中未成组合的牌在吃牌后是否可以形成组合
	cardGroupList := GroupTheCards(handCardsAfterEat, WildCard)
	allCookiesCardsAfterEat := make([]uint32, 0) // 所有不成组合的牌集
	for _, cardGroup := range cardGroupList {
		if !ValidRmGroup(cardGroup, WildCard) {
			allCookiesCardsAfterEat = append(allCookiesCardsAfterEat, cardGroup...)
		}
	}
	fmt.Println("吃牌不能胡牌时,吃牌后有效牌组:", cardGroupList, "吃牌后散牌牌集:", allCookiesCardsAfterEat)
	if !Contains(allCookiesCardsAfterEat, latestCard) { // 散牌牌集不包含可以吃的牌
		fmt.Println("吃牌可以与散牌牌集构成有效组合或者吃牌可以对有效牌组进行扩展")
		return filterOutCardWithEat(mCards, FaceUpCard, WildCard, latestCard)
	}
	fmt.Println("吃牌不能形成有效组合")
	///////////////////////////////////////////////////////
	cardGroupListBeforeEat := GroupTheCards(handCards, WildCard)
	// 检查手中是否有纯顺子
	hasPureSequence := hasPureSequence(cardGroupListBeforeEat)
	if !hasPureSequence { // 没有纯顺子
		fmt.Println("手牌中没有纯顺子组合")
		if !Contains(handCards, latestCard) {
			for _, v := range handCards {
				temp := int(v) - int(latestCard)
				fmt.Println("吃牌", latestCard, "手牌", v, "temp", temp)
				if temp == 1 || temp == -1 || temp == 2 || temp == -2 { // 可以组成成牌率更高的缺一门纯顺子组合
					fmt.Println("吃牌可以组成缺一门的纯顺子组合")
					return filterOutCardWithEat(mCards, FaceUpCard, WildCard, latestCard)
				}
			}
		}
	}
	// 手中有纯顺子
	allCookiesCardsBeforeEat := make([]uint32, 0) // 所有不成组合的牌集
	for _, cardGroup := range cardGroupListBeforeEat {
		if ValidRmGroup(cardGroup, WildCard) {
			allCookiesCardsBeforeEat = append(allCookiesCardsBeforeEat, cardGroup...)
		}
	}
	// 使用散牌牌集，分析出所有的缺一门组合
	lackGroups := FindAllLackGroup(allCookiesCardsBeforeEat)
	maxRate := int64(0) //最大成牌概率
	for _, lackGroup := range lackGroups {
		completeRate := GetCompleteRateForLack(handCards, FaceUpCard, lackGroup)
		if completeRate > maxRate {
			maxRate = completeRate
		}
	}
	// 吃牌后是否可以形成成牌概率更高的缺一门组合
	lackGroupsAfterEat := FindAllLackGroup(allCookiesCardsAfterEat)
	maxRateAfterEat := int64(0) //最大成牌概率
	for _, lackGroup := range lackGroupsAfterEat {
		completeRate := GetCompleteRateForLack(handCards, FaceUpCard, lackGroup)
		if completeRate > maxRateAfterEat {
			maxRateAfterEat = completeRate
		}
	}
	if maxRateAfterEat > maxRate {
		fmt.Println("吃牌后可以形成成牌概率更高的缺一门组合")
		return filterOutCardWithEat(mCards, FaceUpCard, WildCard, latestCard)
	}
	fmt.Println("吃牌后不能形成成牌概率更高的缺一门组合")
	return 1
}

// 过滤出牌是否是吃牌
func filterOutCardWithEat(mCards []uint32, FaceUpCard []uint32, WildCard, eatCard uint32) int32 {
	handCardsAfterEat := make([]uint32, 0)
	handCardsAfterEat = append(handCardsAfterEat, mCards...)
	handCardsAfterEat = append(handCardsAfterEat, eatCard)
	outCard, _, _ := LetsOutCard(handCardsAfterEat, FaceUpCard, WildCard)
	if eatCard == outCard {
		return 1
	}
	return 2
}
