package fortunegems

import (
	"goserver/pkg/table"
	"math/rand"
)

func GetElementsByConfig(column int32, isExtraBet bool) (result [3]Symbol) {
	reelRecord := table.GetTables().FortuneGemsReelTable.Get(column)

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

func IsGM(userid string) bool {
	gmTable := table.GetTables().FortuneGems2GMTable.Get() //直接用宝石2的配置表
	for _, v := range gmTable.GMUserID {
		if v == userid {
			return true
		}
	}
	return false
}
