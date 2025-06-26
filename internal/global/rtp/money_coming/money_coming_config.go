package moneycoming

import (
	"goserver/pkg/table"
	"goserver/pkg/utils"
)

func GetElementsByConfig(column int32, level int32) (result [3]Symbol) {
	reelRecord := table.GetTables().MoneyComingReelTable.Get(column)

	var rangeStart, rangeEnd int

	for _, v := range reelRecord.ReelCountRange {
		if v.Count == level {
			rangeStart = int(v.Start)
			rangeEnd = int(v.End)
			break
		}
	}

	choices := []utils.Choice{}
	for i := rangeStart - 1; i <= rangeEnd-1; i++ {
		choices = append(choices, utils.Choice{Weight: int(reelRecord.ElementWeightList[i]), Item: i + 1})
	}
	choice, _ := utils.WeightedChoice(choices)
	weightPos := choice.Item.(int)

	for i := 0; i < 3; i++ {
		pos := (weightPos-1+i)%(rangeEnd-rangeStart+1) + rangeStart
		result[i] = Symbol(reelRecord.ElementList[int32(pos-1)])
	}

	return
}
