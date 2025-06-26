package algo

import (
	"fmt"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"goserver/pkg/zlog"
	"sort"
)

// 从手牌中抽牌替换
func DrawNeedCards(robot bool, hand, cards []uint32, wildCard uint32, drawNum int, notDraw1st bool, actTimeCards [3]int32) (
	newHand, newCards []uint32,
) {
	zlog.Infof("robot=%v 换牌: %d, %s, %s\n", robot, drawNum, CardsString([]uint32{wildCard}), CardsString(hand))
	// if robot {
	// 	for i := 0; i < drawNum; i++ {
	// 		m, n := utils.RandIntN(len(hand)), utils.RandIntN(len(cards))
	// 		fmt.Println("人机换牌", CardsString([]uint32{hand[m]}), "->", CardsString([]uint32{cards[n]}), ", ")
	// 		hand[m], cards[n] = cards[n], hand[m]
	// 	}
	// 	newHand = hand
	// 	newCards = cards
	// 	return
	// }

	// 玩家需要的牌, 放到牌组后面
	// 同花色，靠牌，joker牌
	groups := GroupTheCards(hand, wildCard)
	var baoPureShunKey string
	if notDraw1st {
		for _, g := range groups {
			if IsPureTonghuashun(g) {
				baoPureShunKey = getGroupKey(g)
				break
			}
		}
	}

	var draws, needCards []uint32
	drawsMap := make(map[uint32]bool, drawNum) // 避免抽中同一张牌,移除牌时手牌数量不对
	for i, loop := 0, 0; i < drawNum && loop < 100; i++ {
		loop++
		group := groups[utils.RandIntN(len(groups))]
		if notDraw1st {
			if IsPureTonghuashun(group) && baoPureShunKey == getGroupKey(group) {
				i-- // 保同花顺换一组抽
				continue
			}
		}
		card := group[utils.RandIntN(len(group))]
		if drawsMap[card] {
			i-- // 抽过这张了重新抽
			continue
		}
		draws = append(draws, card)
		drawsMap[card] = true
		zlog.Infof("抽牌: %s", CardsString([]uint32{card}))
		if IsLaizi(card, wildCard) {
			continue
		} else {
			needCards = append(needCards, card)
		}
		if IsTonghuashun(group, wildCard) {
			Sort(group)
			// 抽的左边第一张非癞子，补右边
			for _, c := range group {
				if IsLaizi(c, wildCard) {
					continue
				}
				if card != c {
					break
				}
				// 补右边
				for i := len(group) - 1; i >= 0; i-- {
					if IsLaizi(group[i], wildCard) {
						continue
					}
					rank, suit := Rank(group[i]), Suit(group[i])
					if rank == King {
						needCards = append(needCards, Ace|suit)
					} else {
						needCards = append(needCards, (rank+1)|suit)
					}
					break
				}
				break
			}
			for i := len(group) - 1; i >= 0; i-- {
				if IsLaizi(group[i], wildCard) {
					continue
				}
				if card != group[i] {
					break
				}
				// 补左边
				for _, c := range group {
					if IsLaizi(c, wildCard) {
						continue
					}
					rank, suit := Rank(c), Suit(c)
					// 抽的右边第一张非癞子，补左边
					// if Rank(card) == King && rank != Ace
					if rank != Ace {
						needCards = append(needCards, (rank-1)|suit)
					}
				}
			}

		} else if IsBaozi(group, wildCard) {
			// 豹子...
			var suits = make(map[uint32]bool, 4)
			rank := Rank(card)
			for _, c := range group {
				if IsLaizi(c, wildCard) {
					continue
				}
				s := Suit(c)
				suits[s] = true
			}
			for _, suit := range []uint32{Spade, Heart, Club, Diamond} {
				if !suits[suit] {
					needCards = append(needCards, rank|suit)
				}
			}
		}
	}

	zlog.Infof("需要的牌: %s", CardsString(needCards))
	// 控制牌堆 needCards
	needCards = append(needCards, cardPointTo4Color(wildCard)...) // 癞子牌
	needCards = append(needCards, RMJOKERS[0], RMJOKERS[1])

	needCardsMap := make(map[uint32]bool, len(needCards))
	for _, c := range needCards {
		needCardsMap[c] = true
	}

	// 抽出的牌放到牌堆
	cards = append(cards, draws...)
	cardsMap := make(map[uint32]int, len(cards))
	for _, c := range cards {
		cardsMap[c]++
	}

	// 补齐人机不需要的手牌, 整理牌堆
	if robot {
		var newDraw []uint32
		for range draws {
			var found bool
			for card, count := range cardsMap {
				if count > 0 {
					if _, ok := needCardsMap[card]; !ok {
						newDraw = append(newDraw, card)
						cardsMap[card]--
						found = true
						break
					}
				}
			}
			if !found {
				for card, count := range cardsMap {
					if count > 0 {
						cardsMap[card]--
						newDraw = append(newDraw, card)
					}
				}
			}
		}
		newHand, _ = RemoveCards(hand, draws)
		newHand = append(newHand, newDraw...)

		for card, count := range cardsMap {
			for i := 0; i < count; i++ {
				newCards = append(newCards, card)
			}
		}
		sort.Slice(newCards, func(_, _ int) bool {
			return utils.RandWan(5000)
		})
		return
	}

	// 将需要的牌按策略散到牌堆
	cardsPlans := make([][]uint32, 3)
	ranges := [][]int{{0, 5}, {5, 10}, {10, 20}}
	for act, r := range ranges {
		for i := r[0]; i < r[1]; i++ { // actTimeCards[0]
			// 补需要的牌
			if actTimeCards[act] > 0 {
				var needCard uint32
				for c, ok := range needCardsMap {
					if ok && cardsMap[c] > 0 {
						needCard = c
						cardsMap[c]--
						needCardsMap[c] = false
						actTimeCards[act]--
						break
					}
				}
				if needCard != 0 {
					cardsPlans[act] = append(cardsPlans[act], needCard)
					continue
				}
			}
			// 补不需要的牌
			var neCard uint32
			for c, count := range cardsMap {
				if count > 0 {
					if _, ok := needCardsMap[c]; !ok {
						neCard = c
						cardsMap[c]--
						break
					}
				}
			}
			if neCard != 0 {
				cardsPlans[act] = append(cardsPlans[act], neCard)
				continue
			}
			// 随便补张
			for c, count := range cardsMap {
				if count > 0 {
					cardsMap[c]--
					cardsPlans[act] = append(cardsPlans[act], c)
				}
			}
		}
		// 打乱一下
		sort.Slice(cardsPlans[act], func(_, _ int) bool {
			return utils.RandWan(5000)
		})
	}

	// 补齐玩家手牌
	var newDraw []uint32
	for range draws {
		var found bool
		for card, count := range cardsMap {
			if count > 0 {
				if _, ok := needCardsMap[card]; !ok {
					newDraw = append(newDraw, card)
					cardsMap[card]--
					found = true
					break
				}
			}
		}
		if !found {
			for card, count := range cardsMap {
				if count > 0 {
					cardsMap[card]--
					newDraw = append(newDraw, card)
				}
			}
		}
	}
	newHand, _ = RemoveCards(hand, draws)
	newHand = append(newHand, newDraw...)

	// 重新整理牌堆
	newCards = make([]uint32, 0, len(cards))
	zlog.Infof("cardPlans: %s", CardGroupsString(cardsPlans))
	for _, cards := range cardsPlans {
		newCards = append(newCards, cards...)
	}
	leftCards := make([]uint32, 0)
	for card, count := range cardsMap {
		for i := 0; i < count; i++ {
			leftCards = append(leftCards, card)
		}
	}
	sort.Slice(leftCards, func(_, _ int) bool {
		return utils.RandWan(5000)
	})
	newCards = append(newCards, leftCards...)
	return
}

// 分析玩家下一步需要的牌列表
func NeedNextCardList(mCards []uint32, faceUpCard []uint32, untouchedCards []uint32, wildCard uint32) (
	needCardsNext []uint32, finish bool,
) {
	untouchedCardsMap := convertMapCard(untouchedCards)
	handCards := make([]uint32, len(mCards))
	copy(handCards, mCards)

	groups := GroupTheCards(handCards, wildCard)
	// 有无1顺,2顺
	var pureShuns [][]uint32
	var shuns [][]uint32
	var baozis [][]uint32
	var allCookieCards []uint32
	for _, group := range groups {
		if len(group) < 3 {
			allCookieCards = append(allCookieCards, group...)
			continue
		}
		if IsPureTonghuashun(group) {
			pureShuns = append(pureShuns, group)
			continue
		}
		if IsTonghuashun(group, wildCard) {
			shuns = append(shuns, group)
			continue
		}
		if IsBaozi(group, wildCard) {
			baozis = append(baozis, group)
			continue
		}
		allCookieCards = append(allCookieCards, group...)
	}
	// 已胡牌
	if len(pureShuns) > 0 && len(pureShuns)+len(shuns) >= 2 && len(allCookieCards) == 0 {
		finish = true
		return
	}

	// 缺一门牌组
	var lackGroups [][]uint32
	// var selectedCardGroup *ValidCardGroup
	if len(allCookieCards) > 1 {
		var optCardGroups []*ValidCardGroup = make([]*ValidCardGroup, 0) // 散牌牌集的所有可选分组方案
		cookiesCardsList := getAllCookiesCards(allCookieCards)           // 所有散牌牌集可能的分组
		// fmt.Println("--有效牌组", allBaseGroups, "以外的剩余散牌牌集", allCookiesCards)
		// 计算所有cookiesCardsList的成牌概率并取最大值作为首选
		var selectedCookiesCards *CookiesCards
		var maxRate int64 = -1
		for _, cookiesCards := range cookiesCardsList {
			rate := int64(0)
			for _, lackGroup := range cookiesCards.lackGroups {
				rate += GetCompleteRateForLack(handCards, faceUpCard, lackGroup) * 106
			}
			for _, singleCard := range cookiesCards.leftCards {
				rate += GetCompleteRateForCard(handCards, faceUpCard, singleCard, wildCard)
			}
			// fmt.Println("----有效牌组", allBaseGroups, "以外的剩余散牌牌集,可选的分组", "缺一门牌组", cookiesCards.lackGroups, "散牌牌集", cookiesCards.leftCards, "成牌概率", rate)
			if rate > maxRate {
				maxRate = rate
				selectedCookiesCards = cookiesCards
			}
		}

		validCardGroup := newValidCardGroup([][]uint32{}, selectedCookiesCards.lackGroups, selectedCookiesCards.leftCards, maxRate)
		optCardGroups = append(optCardGroups, validCardGroup)
		// fmt.Println("--有效牌组", validCardGroup.baseGroups, "以外的剩余散牌牌集,选择的分组", "缺一门牌组", validCardGroup.lackGroups, "散牌牌集", validCardGroup.leftCards, "成牌概率", validCardGroup.rate)
		//  在所有optCardGroups中找出一种成牌概率最高的
		selectedCardGroup := selectValidCardGroup(optCardGroups)

		lackGroups = selectedCardGroup.lackGroups
		allCookieCards = selectedCardGroup.leftCards
	}

	// 凑1st 需牌
	if len(pureShuns) == 0 {
		var baoziLackGroups [][]uint32
		// 从缺一门中找
	lackLoop:
		for _, lackGroup := range lackGroups {
			for _, card := range lackGroup { // 癞子跳过
				if IsLaizi(card, wildCard) {
					continue lackLoop
				}
			}

			// 顺子 lackGroup
			if IsTonghuashun(append(lackGroup, wildCard), wildCard) {
				needCards := FindLackGroupNeedCards(lackGroup)
				if hasCards := hasNeedCards(needCards, untouchedCardsMap, wildCard); len(hasCards) > 0 {
					needCardsNext = append(needCardsNext, hasCards...)
				}
			} else {
				baoziLackGroups = append(baoziLackGroups, lackGroup)
			}
		}

		// 散牌凑两张
		for _, card := range allCookieCards {
			if IsLaizi(card, wildCard) {
				continue
			}
			needCards := FindPureShunNeedCards(card)
			if hasCards := hasNeedCards(needCards, untouchedCardsMap, wildCard); len(hasCards) > 0 {
				needCardsNext = append(needCardsNext, hasCards...)
			}
		}

		// 拆豹子缺牌
		for _, lackGroup := range baoziLackGroups {
			for _, card := range lackGroup {
				FindPureShunNeedCards(card)
				needCards := FindPureShunNeedCards(card)
				if hasCards := hasNeedCards(needCards, untouchedCardsMap, wildCard); len(hasCards) > 0 {
					needCardsNext = append(needCardsNext, hasCards...)
				}
			}
		}

		// 拆2st顺子
		if len(shuns) > 0 {
			for _, shun := range shuns {
				var normals []uint32
				for _, card := range shun {
					if !IsLaizi(card, wildCard) {
						normals = append(normals, card)
					}
				}
				if len(normals) == 2 { // 2st 缺1张就是同花顺
					needCards := FindLackGroupNeedCards(normals)
					if hasCards := hasNeedCards(needCards, untouchedCardsMap, wildCard); len(hasCards) > 0 {
						needCardsNext = append(needCardsNext, hasCards...)
					}
				}
			}
		}

		// 拆豹子
		if len(baozis) > 0 {
			for _, baozi := range baozis {
				for _, card := range baozi {
					if !IsLaizi(card, wildCard) {
						needCards := FindPureShunNeedCards(card)
						if hasCards := hasNeedCards(needCards, untouchedCardsMap, wildCard); len(hasCards) > 0 {
							needCardsNext = append(needCardsNext, hasCards...)
						}
					}
				}
			}
		}
	}

	// 凑2st
	if len(shuns) == 0 {
		// 从缺一门中找
		for _, lackGroup := range lackGroups {
			// 顺子 lackGroup
			if IsTonghuashun(append(lackGroup, wildCard), wildCard) {
				needCards := FindLackGroupNeedCards(lackGroup)
				// needCards = append(needCards, cardPointTo4Color(wildCard)...) // 癞子牌
				// needCards = append(needCards, RMJOKERS[0], RMJOKERS[1])
				if hasCards := hasNeedCards(needCards, untouchedCardsMap, wildCard); len(hasCards) > 0 {
					needCardsNext = append(needCardsNext, hasCards...)
				}
			}
		}

		// 散牌凑两张
		for _, card := range allCookieCards {
			needCards := FindPureShunNeedCards(card)
			// needCards = append(needCards, cardPointTo4Color(wildCard)...) // 癞子牌
			// needCards = append(needCards, RMJOKERS[0], RMJOKERS[1])
			if hasCards := hasNeedCards(needCards, untouchedCardsMap, wildCard); len(hasCards) > 0 {
				needCardsNext = append(needCardsNext, hasCards...)
			}
		}
	}

	// 癞顺子不能让补齐
	for _, shun := range shuns {
		var normals []uint32
		for _, card := range shun {
			if !IsLaizi(card, wildCard) {
				normals = append(normals, card)
			}
		}
		length := len(normals)
		if length < 2 {
			continue
		}
		Sort(normals)
		var needCards []uint32
		suit := Suit(normals[0])
		// 补中间
		for i, card := range normals {
			if i == length-1 {
				break
			}
			r1, r2 := Rank(card), Rank(normals[i+1])
			if r1 != r2-1 {
				for j := 0; j < int(r2-r1)-1; j++ {
					needCards = append(needCards, (r1+1)|suit)
				}
			}
		}
		r0, r9 := Rank(normals[0]), Rank(normals[len(normals)-1])
		if r0 != Ace {
			// 补左边
			needCards = append(needCards, (r0-1)|suit)
		} else {
			if r9 == Queen { // 缺K
				needCards = append(needCards, King|suit)
			}
		}
		if r9 != King {
			// 补右边
			needCards = append(needCards, (r9+1)|suit)
		}

		if hasCards := hasNeedCards(needCards, untouchedCardsMap, wildCard); len(hasCards) > 0 {
			needCardsNext = append(needCardsNext, hasCards...)
		}
	}

	// 凑缺一门
	for _, lackGroup := range lackGroups {
		// 顺子 lackGroup
		if ValidRmGroup(append(lackGroup, wildCard), wildCard) {
			needCards := FindLackGroupNeedCards(lackGroup)
			// needCards = append(needCards, cardPointTo4Color(wildCard)...) // 癞子牌
			// needCards = append(needCards, RMJOKERS[0], RMJOKERS[1])
			if hasCards := hasNeedCards(needCards, untouchedCardsMap, wildCard); len(hasCards) > 0 {
				needCardsNext = append(needCardsNext, hasCards...)
			}
		}
	}

	// 还剩2张以下散牌, 补已成组的靠张
	if len(allCookieCards) <= 2 {
		groups := make([][]uint32, 0, len(pureShuns)+len(shuns)+len(baozis))
		groups = append(groups, pureShuns...)
		groups = append(groups, shuns...)
		groups = append(groups, baozis...)
		for _, group := range groups {
			Sort(group)
			var needCards []uint32
			if IsTonghuashun(group, wildCard) {
				// 补顺子左边
				for _, card := range group {
					if IsLaizi(card, wildCard) {
						continue
					}
					rank, suit := Rank(card), Suit(card)
					if rank == Ace { // 最左边是A
						break
					}
					needCards = append(needCards, (rank-1)|suit)
					break
				}
				// 补顺子右边
				if !(Rank(group[0]) == Ace && Rank(group[len(group)-1]) == King) {
					for i := len(group) - 1; i >= 0; i-- {
						card := group[i]
						if IsLaizi(card, wildCard) {
							continue
						}
						rank, suit := Rank(card), Suit(card)
						if rank == King { // 最右边是K
							needCards = append(needCards, (Ace+1)|suit)
							break
						}
						needCards = append(needCards, (rank+1)|suit)
						break
					}
				}

			} else if IsBaozi(group, wildCard) {
				var suits = make(map[uint32]bool, 3)
				var lRank uint32
				for _, card := range group {
					suits[Suit(card)] = true
					lRank = Rank(card)
				}
				if len(suits) < 4 {
					Suits := []uint32{Spade, Heart, Club, Diamond}
					for _, s := range Suits {
						if !suits[s] {
							needCards = append(needCards, s|lRank)
						}
					}
				}
			}

			if hasCards := hasNeedCards(needCards, untouchedCardsMap, wildCard); len(hasCards) > 0 {
				needCardsNext = append(needCardsNext, hasCards...)
			}
		}
	}

	// todo 找缺N门(最小), 2张散牌总补牌数<4
	// 散牌凑两张
	for _, card := range allCookieCards {
		needCards := FindPureShunNeedCards(card)                  // 顺子
		needCards = append(needCards, cardPointTo3Color(card)...) // 豹子
		// needCards = append(needCards, cardPointTo4Color(wildCard)...) // 癞子牌
		// needCards = append(needCards, RMJOKERS[0], RMJOKERS[1])
		if hasCards := hasNeedCards(needCards, untouchedCardsMap, wildCard); len(hasCards) > 0 {
			needCardsNext = append(needCardsNext, hasCards...)
		}
	}

	needCardsNext = append(needCardsNext, cardPointTo4Color(wildCard)...) // 癞子牌
	needCardsNext = append(needCardsNext, RMJOKERS[0], RMJOKERS[1])

	// 去重一下
	cc := make(map[uint32]bool, len(needCardsNext)/2)
	var distinct []uint32
	for _, card := range needCardsNext {
		if !cc[card] {
			cc[card] = true
			distinct = append(distinct, card)
		}
	}
	needCardsNext = distinct
	return
}

// 分析机器人下一张需要的牌
func NeedNextCard(mCards []uint32, faceUpCard []uint32, untouchedCards []uint32, wildCard uint32) (
	touchCard uint32, finish bool,
) {
	var ok bool
	untouchedCardsMap := convertMapCard(untouchedCards)
	handCards := make([]uint32, len(mCards))
	copy(handCards, mCards)

	groups := GroupTheCards(handCards, wildCard)
	// 有无1顺,2顺
	var pureShuns [][]uint32
	var shuns [][]uint32
	var baozis [][]uint32
	var allCookieCards []uint32
	for _, group := range groups {
		if len(group) < 3 {
			allCookieCards = append(allCookieCards, group...)
			continue
		}
		if IsPureTonghuashun(group) {
			pureShuns = append(pureShuns, group)
			continue
		}
		if IsTonghuashun(group, wildCard) {
			shuns = append(shuns, group)
			continue
		}
		if IsBaozi(group, wildCard) {
			baozis = append(baozis, group)
			continue
		}
		allCookieCards = append(allCookieCards, group...)
	}
	// 已胡牌
	if len(pureShuns) > 0 && len(pureShuns)+len(shuns) >= 2 && len(allCookieCards) == 0 {
		finish = true
		return
	}

	// 缺一门牌组
	var lackGroups [][]uint32
	// var selectedCardGroup *ValidCardGroup
	if len(allCookieCards) > 1 {
		var optCardGroups []*ValidCardGroup = make([]*ValidCardGroup, 0) // 散牌牌集的所有可选分组方案
		cookiesCardsList := getAllCookiesCards(allCookieCards)           // 所有散牌牌集可能的分组
		// fmt.Println("--有效牌组", allBaseGroups, "以外的剩余散牌牌集", allCookiesCards)
		// 计算所有cookiesCardsList的成牌概率并取最大值作为首选
		var selectedCookiesCards *CookiesCards
		var maxRate int64 = -1
		for _, cookiesCards := range cookiesCardsList {
			rate := int64(0)
			for _, lackGroup := range cookiesCards.lackGroups {
				rate += GetCompleteRateForLack(handCards, faceUpCard, lackGroup) * 106
			}
			for _, singleCard := range cookiesCards.leftCards {
				rate += GetCompleteRateForCard(handCards, faceUpCard, singleCard, wildCard)
			}
			// fmt.Println("----有效牌组", allBaseGroups, "以外的剩余散牌牌集,可选的分组", "缺一门牌组", cookiesCards.lackGroups, "散牌牌集", cookiesCards.leftCards, "成牌概率", rate)
			if rate > maxRate {
				maxRate = rate
				selectedCookiesCards = cookiesCards
			}
		}

		validCardGroup := newValidCardGroup([][]uint32{}, selectedCookiesCards.lackGroups, selectedCookiesCards.leftCards, maxRate)
		optCardGroups = append(optCardGroups, validCardGroup)
		// fmt.Println("--有效牌组", validCardGroup.baseGroups, "以外的剩余散牌牌集,选择的分组", "缺一门牌组", validCardGroup.lackGroups, "散牌牌集", validCardGroup.leftCards, "成牌概率", validCardGroup.rate)
		//  在所有optCardGroups中找出一种成牌概率最高的
		selectedCardGroup := selectValidCardGroup(optCardGroups)

		lackGroups = selectedCardGroup.lackGroups
		allCookieCards = selectedCardGroup.leftCards
	}

	// 凑1st 需牌
	if len(pureShuns) == 0 {
		var baoziLackGroups [][]uint32
		// 从缺一门中找
	lackLoop:
		for _, lackGroup := range lackGroups {
			for _, card := range lackGroup { // 癞子跳过
				if IsLaizi(card, wildCard) {
					continue lackLoop
				}
			}

			// 顺子 lackGroup
			if IsTonghuashun(append(lackGroup, wildCard), wildCard) {
				needCards := FindLackGroupNeedCards(lackGroup)
				if touchCard, ok = choiceNeedCards(needCards, untouchedCardsMap, wildCard); ok {
					return
				}
			} else {
				baoziLackGroups = append(baoziLackGroups, lackGroup)
			}
		}

		// 散牌凑两张
		for _, card := range allCookieCards {
			if IsLaizi(card, wildCard) {
				continue
			}
			needCards := FindPureShunNeedCards(card)
			if touchCard, ok = choiceNeedCards(needCards, untouchedCardsMap, wildCard); ok {
				return
			}
		}

		// 拆豹子缺牌
		for _, lackGroup := range baoziLackGroups {
			for _, card := range lackGroup {
				FindPureShunNeedCards(card)
				needCards := FindPureShunNeedCards(card)
				if touchCard, ok = choiceNeedCards(needCards, untouchedCardsMap, wildCard); ok {
					return
				}
			}
		}

		// 拆2st顺子
		if len(shuns) > 0 {
			for _, shun := range shuns {
				var normals []uint32
				for _, card := range shun {
					if !IsLaizi(card, wildCard) {
						normals = append(normals, card)
					}
				}
				if len(normals) == 2 { // 2st 缺1张就是同花顺
					needCards := FindLackGroupNeedCards(normals)
					if touchCard, ok = choiceNeedCards(needCards, untouchedCardsMap, wildCard); ok {
						return
					}
				}
			}
		}

		// 拆豹子
		if len(baozis) > 0 {
			for _, baozi := range baozis {
				for _, card := range baozi {
					if !IsLaizi(card, wildCard) {
						needCards := FindPureShunNeedCards(card)
						if touchCard, ok = choiceNeedCards(needCards, untouchedCardsMap, wildCard); ok {
							return
						}
					}
				}
			}
		}

	}

	// 凑2st
	if len(shuns) == 0 {
		// 从缺一门中找
		for _, lackGroup := range lackGroups {
			// 顺子 lackGroup
			if IsTonghuashun(append(lackGroup, wildCard), wildCard) {
				needCards := FindLackGroupNeedCards(lackGroup)
				needCards = append(needCards, cardPointTo4Color(wildCard)...) // 癞子牌
				needCards = append(needCards, RMJOKERS[0], RMJOKERS[1])

				if touchCard, ok = choiceNeedCards(needCards, untouchedCardsMap, wildCard); ok {
					return
				}
			}
		}

		// 散牌凑两张
		for _, card := range allCookieCards {
			needCards := FindPureShunNeedCards(card)
			needCards = append(needCards, cardPointTo4Color(wildCard)...) // 癞子牌
			needCards = append(needCards, RMJOKERS[0], RMJOKERS[1])

			if touchCard, ok = choiceNeedCards(needCards, untouchedCardsMap, wildCard); ok {
				return
			}
		}
	}

	// 凑缺一门
	for _, lackGroup := range lackGroups {
		// 顺子 lackGroup
		if ValidRmGroup(append(lackGroup, wildCard), wildCard) {
			needCards := FindLackGroupNeedCards(lackGroup)
			needCards = append(needCards, cardPointTo4Color(wildCard)...) // 癞子牌
			needCards = append(needCards, RMJOKERS[0], RMJOKERS[1])

			if touchCard, ok = choiceNeedCards(needCards, untouchedCardsMap, wildCard); ok {
				return
			}
		}
	}

	// 还剩2张以下散牌, 补已成组的靠张
	if len(allCookieCards) <= 2 {
		groups := make([][]uint32, 0, len(pureShuns)+len(shuns)+len(baozis))
		groups = append(groups, pureShuns...)
		groups = append(groups, shuns...)
		groups = append(groups, baozis...)
		sort.Slice(groups, func(_, _ int) bool {
			return utils.RandIntN(100) < 50
		})
		for _, group := range groups {
			Sort(group)
			var needCards []uint32
			if IsTonghuashun(group, wildCard) {
				// 补顺子左边
				for _, card := range group {
					if IsLaizi(card, wildCard) {
						continue
					}
					rank, suit := Rank(card), Suit(card)
					if rank == Ace { // 最左边是A
						break
					}
					needCards = append(needCards, (rank-1)|suit)
					break
				}
				// 补顺子右边
				if !(Rank(group[0]) == Ace && Rank(group[len(group)-1]) == King) {
					for i := len(group) - 1; i >= 0; i-- {
						card := group[i]
						if IsLaizi(card, wildCard) {
							continue
						}
						rank, suit := Rank(card), Suit(card)
						if rank == King { // 最右边是K
							needCards = append(needCards, (Ace+1)|suit)
							break
						}
						needCards = append(needCards, (rank+1)|suit)
						break
					}
				}

			} else if IsBaozi(group, wildCard) {
				var suits = make(map[uint32]bool, 3)
				var lRank uint32
				for _, card := range group {
					suits[Suit(card)] = true
					lRank = Rank(card)
				}
				if len(suits) < 4 {
					Suits := []uint32{Spade, Heart, Club, Diamond}
					for _, s := range Suits {
						if !suits[s] {
							needCards = append(needCards, s|lRank)
						}
					}
				}
			}

			if touchCard, ok = choiceNeedCards(needCards, untouchedCardsMap, wildCard); ok {
				return
			}
		}
	}

	// todo 找缺N门(最小), 2张散牌总补牌数<4
	// 散牌凑两张
	for _, card := range allCookieCards {
		needCards := FindPureShunNeedCards(card)                      // 顺子
		needCards = append(needCards, cardPointTo3Color(card)...)     // 豹子
		needCards = append(needCards, cardPointTo4Color(wildCard)...) // 癞子牌
		needCards = append(needCards, RMJOKERS[0], RMJOKERS[1])
		if touchCard, ok = choiceNeedCards(needCards, untouchedCardsMap, wildCard); ok {
			return
		}
	}

	// 随机摸一张
	return untouchedCards[utils.RandIntN(len(untouchedCards))], false
}

func hasNeedCards(needCards []uint32, untouchedCardsMap map[string]int, wildCard uint32) (hasCards []uint32) {
	if len(needCards) == 0 {
		return
	}
	for _, card := range needCards {
		key := getKey(card)
		if untouchedCardsMap[key] > 0 {
			hasCards = append(hasCards, card)
		}
	}
	return
}

// choiceNeedCards 从需要的牌中选择一张
func choiceNeedCards(needCards []uint32, untouchedCardsMap map[string]int, wildCard uint32) (touchCard uint32, ok bool) {
	if len(needCards) == 0 {
		return
	}
	ok = true
	hasNeedCards := hasNeedCards(needCards, untouchedCardsMap, wildCard)
	// var hasNeedCards []uint32
	// for _, card := range needCards {
	// 	key := getKey(card)
	// 	if untouchedCardsMap[key] > 0 {
	// 		// untouchedCardsMap[key]--
	// 		hasNeedCards = append(hasNeedCards, card)
	// 	}
	// }
	if len(hasNeedCards) == 0 { // 死牌
		ok = false
		return
	}

	choices := utils.SliceMapping(hasNeedCards, func(card uint32) utils.Choice {
		var weight = 100
		if IsLaizi(card, wildCard) {
			weight = 10
		}
		return utils.Choice{Item: card, Weight: weight}
	})
	c, err := utils.WeightedChoice(choices)
	if err != nil {
		glog.Error(err)
		touchCard = hasNeedCards[utils.RandIntN(len(hasNeedCards))]
		return
	}
	touchCard = c.Item.(uint32)
	return
}

func FindLackGroupNeedCards(lackGroup []uint32) (needCards []uint32) {
	if len(lackGroup) != 2 {
		panic(fmt.Errorf("invalid lack cards, count=%v", len(lackGroup)))
	}
	Sort(lackGroup)

	var ranks []uint32
	var suits = make(map[uint32]bool, 2)
	var lSuit uint32
	for _, card := range lackGroup {
		ranks = append(ranks, Rank(card))
		lSuit = Suit(card)
		suits[lSuit] = true
	}
	// 豹子缺一门
	if ranks[0] == ranks[1] {
		if len(suits) != 2 { // 同花豹.
			return
		}
		Suits := []uint32{Spade, Heart, Club, Diamond}
		for _, s := range Suits {
			if !suits[s] {
				needCards = append(needCards, s|ranks[0])
			}
		}
		return
	}

	// 不是同花顺
	if len(suits) > 1 {
		return
	}
	// 顺子缺一门
	if ranks[1]-ranks[0] == 2 { // 缺中间
		needCards = append(needCards, (ranks[1]-1)|lSuit)
		return
	}
	if ranks[0] == Ace && ranks[1] == Queen { // QA
		needCards = append(needCards, King|lSuit)
		return
	}
	if ranks[0] == Ace && ranks[1] == King { // KA
		needCards = append(needCards, Queen|lSuit)
		return
	}

	// 缺2边
	if ranks[1]-ranks[0] == 1 {
		if ranks[0] == Ace && ranks[1] == Deuce { // A2
			needCards = append(needCards, Trey|lSuit)
			return
		}
		if ranks[0] == Queen && ranks[1] == King { // QK
			needCards = append(needCards, Jack|lSuit, Ace|lSuit)
			return
		}
		needCards = append(needCards, (ranks[0]-1)|lSuit, (ranks[1]+1)|lSuit)
	}
	return
}

func FindPureShunNeedCards(card uint32) (needCards []uint32) {
	rank, suit := Rank(card), Suit(card)
	if rank == Ace {
		return []uint32{Deuce | suit, Trey | suit}
	}
	if rank == King {
		return []uint32{Queen | suit, King | suit}
	}
	return []uint32{(rank - 1) | suit, (rank + 1) | suit}
}
