package algo

import (
	"fmt"
	"math/rand"
)

/*机器人玩家对手牌进行分组*/
func GroupTheCards(cardArray []uint32, wildCard uint32) [][]uint32 {
	// groupId := time.Now().Unix()
	handCardDivides := FindAllHandCardDivide(cardArray, wildCard, 4)
	validCardGroupList := HandleGroupAppend(cardArray, wildCard, handCardDivides)

	resultCardGroupList := make([][]uint32, 0)
	if len(validCardGroupList) > 0 {
		// 找出所有有纯顺子的
		hasPureCardGroupList := make([]*HandCardDivide, 0)
		for _, v := range validCardGroupList {
			for _, group := range v.BaseGroups {
				if IsPureTonghuashun(group) {
					hasPureCardGroupList = append(hasPureCardGroupList, v)
				}
			}
		}
		if len(hasPureCardGroupList) > 0 {
			// 然后在找一个点数最小的
			minPoint := int64(100000)
			for _, v := range hasPureCardGroupList {
				cookiesCards := make([]uint32, 0)
				for _, group := range v.BaseGroups {
					if !ValidRmGroup(group, wildCard) {
						cookiesCards = append(cookiesCards, group...)
					}
				}
				pointCount := GetCardsPoint(cookiesCards, wildCard, true)
				if pointCount == 0 { // 避免选中JOKER单牌未组合的情况
					pointCount = int64(len(cookiesCards))
				}
				if pointCount < minPoint {
					minPoint = pointCount
					resultCardGroupList = v.BaseGroups
				}
			}
		} else {
			// 然后在找一个点数最小的
			minPoint := int64(100000)
			for _, v := range validCardGroupList {
				cookiesCards := make([]uint32, 0)
				for _, group := range v.BaseGroups {
					if !ValidRmGroup(group, wildCard) {
						cookiesCards = append(cookiesCards, group...)
					}
				}
				pointCount := GetCardsPoint(cookiesCards, wildCard, true)
				if pointCount == 0 { // 避免选中JOKER单牌未组合的情况
					pointCount = int64(len(cookiesCards))
				}
				if pointCount < minPoint {
					minPoint = pointCount
					resultCardGroupList = v.BaseGroups
				}
			}
		}
		if resultCardGroupList == nil {
			resultCardGroupList = validCardGroupList[rand.Intn(len(validCardGroupList))].BaseGroups
		}
		return resultCardGroupList
	} else {
		fmt.Println("没有找到有效的分组***************************")
		resultCardGroupList = append(resultCardGroupList, cardArray)
	}
	// fmt.Println(" GroupHandCards", "groupId=", groupId, "一种可能的分组 Start")
	// card.PrintServerCardGroupList(resultCardGroupList) // 一种可能的分组
	// fmt.Println(" GroupHandCards", "groupId=", groupId, "一种可能的分组 End")
	return resultCardGroupList
}
