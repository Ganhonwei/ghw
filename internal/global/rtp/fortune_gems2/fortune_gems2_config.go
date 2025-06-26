package fortunegems2

import (
	"goserver/gen/tb"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"math/rand"
)

func GetElementsByConfig(column int32, isExtraBet bool) (result [3]Symbol) {
	reelRecord := table.GetTables().FortuneGems2ReelTable.Get(column)

	var rangeStart, rangeEnd int
	var index int32

	if isExtraBet {
		index = 2
	} else {
		index = 1
	}

	for _, v := range reelRecord.ReelCountRange {
		if v.Count == index {
			rangeStart = int(v.Start)
			rangeEnd = int(v.End)
			break
		}
	}

	randPos := rand.Intn(rangeEnd-rangeStart+1) + rangeStart

	for i := 0; i < 3; i++ {
		pos := (randPos-1+i)%(rangeEnd-rangeStart+1) + rangeStart
		result[i] = Symbol(reelRecord.ElementList[int32(pos-1)])
	}

	return
}

func GetMultiplierByConfig(isExtraBet bool) int32 {
	var probWeight []*tb.ProbWeight
	if isExtraBet {
		probWeight = table.GetTables().FortuneGems2WheelTable.Get().ExtraBetBaseMultiplier
	} else {
		probWeight = table.GetTables().FortuneGems2WheelTable.Get().BaseMultiplier
	}

	var choices []utils.Choice
	for _, v := range probWeight {
		choices = append(choices, utils.Choice{Weight: int(v.Weight), Item: v.Multiple})
	}

	choice, _ := utils.WeightedChoice(choices)

	return choice.Item.(int32)
}

func GetExtraMultiplierByConfig() int32 {
	probWeight := table.GetTables().FortuneGems2WheelTable.Get().ExtraBetAdditionalMultiplier

	var choices []utils.Choice
	for _, v := range probWeight {
		choices = append(choices, utils.Choice{Weight: int(v.Weight), Item: v.Multiple})
	}

	choice, _ := utils.WeightedChoice(choices)

	return choice.Item.(int32)
}

func IsGM(userid string) bool {
	gmTable := table.GetTables().FortuneGems2GMTable.Get()
	for _, v := range gmTable.GMUserID {
		if v == userid {
			return true
		}
	}
	return false
}
