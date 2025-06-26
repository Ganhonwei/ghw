package main

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"math"
	"sort"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/globalsign/mgo/bson"
)

func StatPlayersM4() {
	m1 := bson.M{
		"robot":            false,
		"simulation_robot": false,
	}
	start, end := utils.Str2Time("2024-04-01 00:00:00", location), utils.Str2Time("2024-05-01 00:00:00", location)
	m1["ctime"] = bson.M{"$gte": start, "$lt": end}

	m2 := []bson.M{
		{"$match": m1},
		{"$project": bson.M{"_id": "$_id"}},
	}
	var r2 []bson.M
	err := data.PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		glog.Errorf("error: ", err)
		return
	}
	userids := utils.SliceMapping(r2, func(r bson.M) string {
		userid := r["_id"].(string)
		return userid
	})

	// 查对局记录
	m3 := []bson.M{
		{"$match": bson.M{"userid": bson.M{"$in": userids}, "number": bson.M{"$gt": 0}}},
		{"$group": bson.M{"_id": "$userid"}},
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": 1}}},
	}
	var r3 []bson.M
	err = data.UserGameDatas.Pipe(m3).All(&r3)
	if err != nil {
		glog.Error("err 2: ", err)
		return
	}
	var total int
	if len(r3) > 0 {
		total = r3[0]["total"].(int)
	}

	// 导出...
	f := excelize.NewFile()
	sheet := "Sheet1"
	// 设置表头

	f.SetCellValue(sheet, "A1", "4月份注册数")
	f.SetCellValue(sheet, "B1", "4月份有对局用户")
	f.SetCellValue(sheet, "A2", len(userids))
	f.SetCellValue(sheet, "B2", total)

	// 保存文件
	if err := f.SaveAs(getExportFileName("4月份注册玩家统计")); err != nil {
		glog.Error("error3: ", err)
	}
}

// StatPlayerRounds 普通用户，付费用户，小R用户，中R用户，大R用户的游戏局数
func StatPlayerRounds() {
	m1 := bson.M{
		"robot":            false,
		"simulation_robot": false,
		"money":            bson.M{"$gte": 10000, "$lte": 99900},
	}
	types := []string{
		"普通玩家", "付费用户", "小R用户", "中R用户", "大R用户",
	}
	typesMap := map[string]bson.M{
		"普通玩家": {"$eq": 0},
		"付费用户": {"$gte": 10000, "$lte": 99999},
		"小R用户": {"$gte": 100000, "$lte": 499999},
		"中R用户": {"$gte": 500000, "$lte": 999999},
		"大R用户": {"$gte": 1000000, "$lte": 9999999},
	}
	s := fmt.Sprintf("%s 00:00:00", "2024-01-01")
	s1 := utils.Str2Time(s, location)
	statStartTime := utils.Time2Stamp(s1)

	typesRound := make(map[string]int)
	externalRound := make(map[string]int)
	for _, t := range types {
		money := typesMap[t]
		m1["money"] = money

		// 该类型下用户id列表
		players := []bson.M{}
		err := data.PlayerUsers.Pipe([]bson.M{
			{"$match": m1},
			{"$project": bson.M{"userid": "$_id"}},
		}).All(&players)
		if err != nil {
			glog.Errorf("%s error: %v", t, err)
			continue
		}
		playerIds := utils.SliceMapping(players, func(p bson.M) string {
			return p["userid"].(string)
		})

		// 查对局数据
		m2 := []bson.M{
			{
				"$match": bson.M{"begin_time": bson.M{"$gt": statStartTime}},
			},
			{
				"$project": bson.M{
					"userids": bson.M{"$split": []string{"$players", ","}},
				},
			},
			{
				"$match": bson.M{"userids": bson.M{"$in": playerIds}},
			},
			{
				"$group": bson.M{"_id": nil, "rounds": bson.M{"$sum": 1}},
			},
		}
		rounds := []bson.M{}
		err = data.Details.Pipe(m2).All(&rounds)
		if err != nil {
			glog.Error("error1: ", err)
			continue
		}
		var round int
		if len(rounds) > 0 {
			round = rounds[0]["rounds"].(int)
		}
		typesRound[t] = round

		// 外接游戏局数
		m3 := []bson.M{
			{
				"$match": bson.M{
					"user_id": bson.M{"$in": playerIds},
					"ctime":   bson.M{"$gt": statStartTime},
				},
			},
			{
				"$group": bson.M{
					"_id":  "$round_id",
					"bets": bson.M{"$sum": 1},
				},
			},
			{
				"$group": bson.M{
					"_id":    nil,
					"rounds": bson.M{"$sum": 1},
				},
			},
		}

		rounds = []bson.M{}
		err = data.NsqLogExternalBets.Pipe(m3).All(&rounds)
		if err != nil {
			glog.Error("error2: ", err)
			continue
		}
		if len(rounds) > 0 {
			round = rounds[0]["rounds"].(int)
		}
		externalRound[t] = round
	}

	// 导出...
	f := excelize.NewFile()
	sheet := "Sheet1"
	// 设置表头
	f.SetCellValue(sheet, "A1", "玩家类型")
	f.SetCellValue(sheet, "B1", "局数")
	f.SetCellValue(sheet, "C1", "外接游戏局数")
	line := 1
	for _, t := range types {
		line++
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", line), t)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", line), typesRound[t])
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", line), externalRound[t])
	}
	// 保存文件
	if err := f.SaveAs(getExportFileName("StatPlayerRounds")); err != nil {
		glog.Error("error3: ", err)
	}
}

func RoundStats() {
	roundRanges := [][2]int{
		[2]int{5, 10}, [2]int{11, 20}, [2]int{21, 30}, [2]int{31, 40}, [2]int{41, 50}, [2]int{51, 60}, [2]int{61, 70}, [2]int{71, 80}, [2]int{81, 90}, [2]int{91, 100}, [2]int{101, 110}, [2]int{111, 120}, [2]int{121, 130}, [2]int{131, 140}, [2]int{141, 150}, [2]int{151, 160}, [2]int{161, 170}, [2]int{171, 180}, [2]int{181, 190}, [2]int{191, 200}, [2]int{201, 300}, [2]int{301, 400}, [2]int{401, 500}, [2]int{501, 600}, [2]int{601, 700}, [2]int{701, 800}, [2]int{801, 900}, [2]int{901, 1000}, [2]int{1001, 1500}, [2]int{1501, 2000}, [2]int{2001, 2500}, [2]int{2501, 3000},
	}
	// fmt.Println(roundRanges)
	typeRangePlayers := make(map[string]map[string]int) // 用户类型,范围,人数

	m1 := bson.M{
		"robot":            false,
		"simulation_robot": false,
		// "money":            bson.M{"$gte": 10000, "$lte": 99900},
	}
	types := []string{
		"普通玩家", "付费用户", "小R用户", "中R用户", "大R用户",
	}
	typesMap := map[string]bson.M{
		"普通玩家": {"$eq": 0},
		"付费用户": {"$gte": 10000, "$lte": 99999},
		"小R用户": {"$gte": 100000, "$lte": 499999},
		"中R用户": {"$gte": 500000, "$lte": 999999},
		"大R用户": {"$gte": 1000000, "$lte": 9999999},
	}

	s := fmt.Sprintf("%s 00:00:00", "2024-01-01")
	s1 := utils.Str2Time(s, location)
	statStartTime := utils.Time2Stamp(s1)

	for _, t := range types {
		typeRangePlayers[t] = make(map[string]int)

		money := typesMap[t]
		m1["money"] = money

		// 该类型下用户id列表
		players := []bson.M{}
		err := data.PlayerUsers.Pipe([]bson.M{
			{"$match": m1},
			{"$project": bson.M{"userid": "$_id"}},
		}).All(&players)
		if err != nil {
			glog.Errorf("%s error: %v", t, err)
			continue
		}
		var playerIds []string
		var playerIdMap = make(map[string]bool, len(playerIds))
		for _, p := range players {
			playerId := p["userid"].(string)
			playerIds = append(playerIds, playerId)
			playerIdMap[playerId] = true
		}
		// playerIds := utils.SliceMapping(players, func(p bson.M) string {
		// 	return p["userid"].(string)
		// })

		// 查对局数据
		m2 := []bson.M{
			{
				"$project": bson.M{
					"userids": bson.M{"$split": []string{"$players", ","}},
				},
			},
			{
				"$match": bson.M{"userids": bson.M{"$in": playerIds}},
			},
			// {
			// 	"$group": bson.M{"_id": nil, "rounds": bson.M{"$sum": 1}},
			// },
		}
		details := []bson.M{}
		err = data.Details.Pipe(m2).All(&details)
		if err != nil {
			glog.Error("error1: ", err)
			continue
		}
		playerRounds := make(map[string]int)
		for _, detail := range details {
			for _, userid := range detail["userids"].([]interface{}) {
				if !playerIdMap[userid.(string)] {
					continue
				}
				// if len(userid.(string)) >= 16 { // 人机id18位长度+
				// 	continue
				// }
				playerRounds[userid.(string)]++
			}
		}

		for _, r := range roundRanges {
			key := fmt.Sprintf("%d-%d", r[0], r[1])
			for _, round := range playerRounds {
				if round >= r[0] && round <= r[1] {
					typeRangePlayers[t][key]++
				}
			}
		}

		// 外接记录
		m3 := []bson.M{
			{
				"$match": bson.M{
					"user_id": bson.M{"$in": playerIds},
					"ctime":   bson.M{"$gt": statStartTime},
				},
			},
			{
				"$group": bson.M{
					"_id":  bson.M{"round_id": "$round_id", "user_id": "$user_id"},
					"bets": bson.M{"$sum": 1},
				},
			},
			{
				"$group": bson.M{
					"_id":    "$_id.user_id",
					"rounds": bson.M{"$sum": 1},
				},
			},
		}

		userRounds := []bson.M{}
		err = data.NsqLogExternalBets.Pipe(m3).All(&userRounds)
		if err != nil {
			glog.Error("error2: ", err)
			continue
		}

		for _, r := range roundRanges {
			key := fmt.Sprintf("%d-%d", r[0], r[1])
			for _, ur := range userRounds {
				// userid := ur["_id"].(string)
				round := ur["rounds"].(int)

				if round >= r[0] && round <= r[1] {
					typeRangePlayers[t][key]++
				}
			}
		}
	}

	// 导出...
	f := excelize.NewFile()
	sheet := "Sheet1"
	// 设置表头

	f.SetCellValue(sheet, "A1", "玩家类型")
	for i, r := range roundRanges {
		key := fmt.Sprintf("%d-%d", r[0], r[1])

		var column string
		for cnum := i + 1; cnum >= 0; cnum -= 26 {
			if cnum >= 26 { // Z
				column += fmt.Sprintf("%c", 0+65)
			} else {
				column += fmt.Sprintf("%c", cnum+65)
			}
		}
		// fmt.Println(column)

		f.SetCellValue(sheet, column+"1", key)

		line := 1
		for _, t := range types {
			line++
			f.SetCellValue("Sheet1", fmt.Sprintf("A%d", line), t)
			f.SetCellValue("Sheet1", fmt.Sprintf("%s%d", column, line), typeRangePlayers[t][key])
		}
	}

	// 保存文件
	if err := f.SaveAs(getExportFileName("RoundStats")); err != nil {
		glog.Error("error3: ", err)
	}
}

// 普通用户201-300局，平均胜率是多少，提现金额是都少，玩的最多的游戏
func RoundStats201() {
	// roundRanges := [][2]int{
	// 	[2]int{5, 10}, [2]int{11, 20}, [2]int{21, 30}, [2]int{31, 40}, [2]int{41, 50}, [2]int{51, 60}, [2]int{61, 70}, [2]int{71, 80}, [2]int{81, 90}, [2]int{91, 100}, [2]int{101, 110}, [2]int{111, 120}, [2]int{121, 130}, [2]int{131, 140}, [2]int{141, 150}, [2]int{151, 160}, [2]int{161, 170}, [2]int{171, 180}, [2]int{181, 190}, [2]int{191, 200}, [2]int{201, 300}, [2]int{301, 400}, [2]int{401, 500}, [2]int{501, 600}, [2]int{601, 700}, [2]int{701, 800}, [2]int{801, 900}, [2]int{901, 1000}, [2]int{1001, 1500}, [2]int{1501, 2000}, [2]int{2001, 2500}, [2]int{2501, 3000},
	// }
	// // fmt.Println(roundRanges)
	// typeRangePlayers := make(map[string]map[string]int) // 用户类型,范围,人数

	m1 := bson.M{
		"robot":            false,
		"simulation_robot": false,
		"money":            bson.M{"$eq": 0},
	}
	// types := []string{
	// 	"普通玩家", "付费用户", "小R用户", "中R用户", "大R用户",
	// }
	// typesMap := map[string]bson.M{
	// 	"普通玩家": {"$eq": 0},
	// 	"付费用户": {"$gte": 10000, "$lte": 99999},
	// 	"小R用户": {"$gte": 100000, "$lte": 499999},
	// 	"中R用户": {"$gte": 500000, "$lte": 999999},
	// 	"大R用户": {"$gte": 1000000, "$lte": 9999999},
	// }

	s := fmt.Sprintf("%s 00:00:00", "2024-01-01")
	s1 := utils.Str2Time(s, location)
	statStartTime := utils.Time2Stamp(s1)

	// playersRounds := make(map[string]int)

	// 该类型下用户id列表
	players := []bson.M{}
	err := data.PlayerUsers.Pipe([]bson.M{
		{"$match": m1},
		{"$project": bson.M{"userid": "$_id"}},
	}).All(&players)
	if err != nil {
		glog.Errorf("统计错误 error: %v", err)
		return
	}
	var playerIds []string
	var playerIdMap = make(map[string]bool, len(playerIds))
	for _, p := range players {
		playerId := p["userid"].(string)
		playerIds = append(playerIds, playerId)
		playerIdMap[playerId] = true
	}

	// 查对局数据
	m2 := []bson.M{
		{
			"$project": bson.M{
				"userids": bson.M{"$split": []string{"$players", ","}},
			},
		},
		{
			"$match": bson.M{"userids": bson.M{"$in": playerIds}},
		},
		// {
		// 	"$group": bson.M{"_id": nil, "rounds": bson.M{"$sum": 1}},
		// },
	}
	details := []bson.M{}
	err = data.Details.Pipe(m2).All(&details)
	if err != nil {
		glog.Error("error1: ", err)
		return
	}
	playerRounds := make(map[string]int)
	for _, detail := range details {
		for _, userid := range detail["userids"].([]interface{}) {
			if !playerIdMap[userid.(string)] {
				continue
			}
			playerRounds[userid.(string)]++
		}
	}

	// 外接记录
	m3 := []bson.M{
		{
			"$match": bson.M{
				"user_id": bson.M{"$in": playerIds},
				"ctime":   bson.M{"$gt": statStartTime},
			},
		},
		{
			"$group": bson.M{
				"_id":  bson.M{"round_id": "$round_id", "user_id": "$user_id"},
				"bets": bson.M{"$sum": 1},
			},
		},
		{
			"$group": bson.M{
				"_id":    "$_id.user_id",
				"rounds": bson.M{"$sum": 1},
			},
		},
	}

	userRounds := []bson.M{}
	err = data.NsqLogExternalBets.Pipe(m3).All(&userRounds)
	if err != nil {
		glog.Error("error2: ", err)
		return
	}

	for _, ur := range userRounds {
		userid := ur["_id"].(string)
		round := ur["rounds"].(int)
		if !playerIdMap[userid] {
			continue
		}
		playerRounds[userid] += round
	}

	var hitUserids []string
	for userid, round := range playerRounds {
		if round >= 201 && round <= 300 {
			hitUserids = append(hitUserids, userid)
		}
	}

	if len(hitUserids) == 0 {
		glog.Warning("not found 201-300 round plain users")
		return
	}
	// 统计提现金额
	m11 := []bson.M{
		{
			"$match": bson.M{
				"robot": false, "simulation_robot": false,
				"_id": bson.M{"$in": hitUserids},
			},
		},
		{
			"$group": bson.M{
				"_id":      nil,
				"cash_out": bson.M{"$sum": "$cash_out"},
			},
		},
	}
	var r11 []bson.M
	err = data.PlayerUsers.Pipe(m11).All(&r11)
	if err != nil {
		glog.Error("m11 error", err)
		return
	}
	// var hitWithdraw int64
	// if len(r11) > 0 {
	// 	hitWithdraw = int64(r11[0]["cash_out"].(int))
	// }

	// 统计游戏类型
	m12 := []bson.M{
		{
			"$project": bson.M{
				"userids":     bson.M{"$split": []string{"$players", ","}},
				"gtype":       "$gtype",
				"desk_id":     "$desk_id",
				"tpdetail":    "$tpdetail",
				"jokerdetail": "$jokerdetail",
				"ak47detail":  "$ak47detail",
				"rmdetail":    "$rmdetail",
				"lhdetail":    "$lhdetail",
				"updetail":    "$updetail",
				"crashdetail": "$crashdetail",
			},
		},
		{
			"$match": bson.M{"userids": bson.M{"$in": hitUserids}},
		},
	}
	var r12 []bson.M
	err = data.Details.Pipe(m12).All(&r12)
	if err != nil {
		glog.Error(err)
		return
	}

	// for _, r := range r12 {
	// 	gtype := r["gtype"].(int)

	// }

	// 导出...
	// f := excelize.NewFile()
	// sheet := "Sheet1"
	// // 设置表头

	// f.SetCellValue(sheet, "A1", "玩家类型")
	// for i, r := range roundRanges {
	// 	key := fmt.Sprintf("%d-%d", r[0], r[1])

	// 	var column string
	// 	for cnum := i + 1; cnum >= 0; cnum -= 26 {
	// 		if cnum >= 26 { // Z
	// 			column += fmt.Sprintf("%c", 0+65)
	// 		} else {
	// 			column += fmt.Sprintf("%c", cnum+65)
	// 		}
	// 	}
	// 	// fmt.Println(column)

	// 	f.SetCellValue(sheet, column+"1", key)

	// 	line := 1
	// 	for _, t := range types {
	// 		line++
	// 		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", line), t)
	// 		f.SetCellValue("Sheet1", fmt.Sprintf("%s%d", column, line), typeRangePlayers[t][key])
	// 	}
	// }

	// // 保存文件
	// if err := f.SaveAs(getExportFileName("RoundStats")); err != nil {
	// 	glog.Error("error3: ", err)
	// }
}

// 拉一个数据7月8日-10日，新注册用户，局数低于5局的ID和局数
func StatLess5Rounds() {
	m1Match := bson.M{
		"robot":            false,
		"simulation_robot": false,
	}
	start, end := utils.Str2Time("2024-07-08 00:00:00", location), utils.Str2Time("2024-07-11 00:00:00", location)
	m1Match["ctime"] = bson.M{"$gte": start, "$lt": end}

	m2 := []bson.M{
		{"$match": m1Match},
		{"$project": bson.M{
			"_id":       "$_id",
			"ctime":     "$ctime",
			"win_rates": "$win_rates",
		}},
	}

	var r1 []*data.User
	err = data.PlayerUsers.Pipe(m2).All(&r1)
	if err != nil {
		glog.Error("users error: ", err)
	}
	var datas []map[string]any
	for _, user := range r1 {
		var totalRounds int64
		for _, rate := range user.WinRates {
			totalRounds += (rate.WinRound + rate.LoseRound)
		}
		if totalRounds < 5 {
			datas = append(datas, map[string]any{"userid": user.Userid, "round": totalRounds})
		}
	}

	f := excelize.NewFile()
	sheet := "Sheet1"
	// 设置表头
	f.SetCellValue(sheet, "A1", "玩家id")
	f.SetCellValue(sheet, "B1", "局数")
	line := 1
	for _, data := range datas {
		line++
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", line), data["userid"])
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", line), data["round"])
	}

	// 保存文件
	if err := f.SaveAs(getExportFileName("StatLess5Rounds")); err != nil {
		glog.Error("error3: ", err)
	}
}

type RMRound struct {
	WaterId       string
	UserId        string
	Score         int64 // 输赢点数
	Hands         int   // 结束轮数
	DrawNumRobot  int
	DrawNumPlayer int
}

func IsPlayer(userid string) bool {
	return len(userid) < 16
}

// 玩家ID	输赢点数	结束轮数
// 玩家操作一次视为一轮（非文档内的手的概念）
func StatsRummyScore() {
	glog.Infof("查询rummy游戏对局 start: ....")

	roomBets := map[string]int64{
		"2801": 50,
		"2802": 100,
		"2803": 200,
		"2804": 500,
		"2805": 1000,
		"2806": 2000,
		"2807": 5000,
		"2808": 10000,
		"2809": 50000,
		"2810": 100000,
		"2101": 10,
		"2102": 100,
		"2103": 500,
		"2104": 1000,
		"2105": 2000,
	}
	now := time.Now()
	startTime := time.Date(2024, 7, 13, 0, 0, 0, 0, now.Location())

	// {"$match": bson.M{"gtype": 4, "begin_time": bson.M{"$gte": startTime.Unix(), "$lte": now.Unix()}}},

	var rmRounds []RMRound = make([]RMRound, 0, 128)
	var r1 []*data.Detail
	for page, size := 1, 500; page == 1 || len(r1) >= 100; page++ {
		r1 = make([]*data.Detail, 0)

		skipNum := (page - 1) * size
		if skipNum < 0 {
			skipNum = 0
		}
		m1 := []bson.M{
			{"$match": bson.M{"gtype": 12, "begin_time": bson.M{"$gte": startTime.Unix(), "$lte": now.Unix()}}}, // "rm_control.ctype": 0,
			{"$sort": bson.M{"begin_time": 1}},
			{"$skip": skipNum},
			{"$limit": size},
		}
		err := data.Details.Pipe(m1).All(&r1)
		if err != nil {
			glog.Errorf("stats detail error: ", err)
			break
		}

		glog.Infof("rummy details: %d, %d", len(r1), len(rmRounds))
		for _, detail := range r1 {
			bet, ok := roomBets[detail.RoomId]
			if !ok || bet == 0 {
				glog.Errorf("room not found: %s", detail.RoomId)
				return
			}
			// if detail.RMControl == nil || detail.RMControl.Ctype != 0 {
			// 	continue
			// }
			for _, item := range detail.RMDetail {
				if !IsPlayer(item.UserId) {
					continue
				}
				var diamond int64
				switch item.Result {
				case "赢":
					diamond = item.Score + item.CashMingTax
				case "输":
					diamond = item.Score
				}

				// 本局输赢分数
				score := int64(math.Ceil(float64(diamond) / float64(bet)))
				r := RMRound{
					WaterId: detail.WaterId,
					UserId:  item.UserId,
					Score:   score,
					Hands:   len(item.MoCards) + len(item.ChuCards),
				}
				if detail.RMControl != nil {
					r.DrawNumRobot = detail.RMControl.DrawNumRobot
					r.DrawNumPlayer = detail.RMControl.DrawNumPlayer
				}
				rmRounds = append(rmRounds, r)
			}
		}
		// if len(r1) > 0 && r1[0] != nil {
		// 	glog.Infof("%d, %d, %d", r1[0].Bets, r1[0].Gtype, len(r1[0].RMDetail))
		// }

	}

	// 导出...
	f := excelize.NewFile()
	sheet := "Sheet1"
	// 设置表头
	f.SetCellValue(sheet, "A1", "局号")
	f.SetCellValue(sheet, "B1", "玩家ID")
	f.SetCellValue(sheet, "C1", "输赢点数")
	f.SetCellValue(sheet, "D1", "结束轮数")
	f.SetCellValue(sheet, "E1", "玩家抽牌数")
	f.SetCellValue(sheet, "F1", "人机抽牌数")
	line := 1
	for _, r := range rmRounds {
		line++
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", line), r.WaterId)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", line), r.UserId)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", line), r.Score)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", line), r.Hands)
		f.SetCellValue("Sheet1", fmt.Sprintf("E%d", line), r.DrawNumPlayer)
		f.SetCellValue("Sheet1", fmt.Sprintf("F%d", line), r.DrawNumRobot)
	}
	// 保存文件
	if err := f.SaveAs(getExportFileName("StatsRummyScore")); err != nil {
		glog.Error("error3: ", err)
	}

	glog.Infof("查询rummy游戏对局 finish: %d, %d-%d", len(r1), startTime.Unix(), now.Unix())
}

// 拉一个数据7月8日-10日，新注册用户，局数低于5局的ID和局数
// 7月1号到7月14日注册的用户，前10局的游戏分布，输赢
func Stat7_1to7_14Round10() {
	start, end := utils.Str2Time("2024-07-01 00:00:00", location), utils.Str2Time("2024-07-15 00:00:00", location)
	_ = end

	m1Match := bson.M{
		"robot":            false,
		"simulation_robot": false,
	}
	m1Match["ctime"] = bson.M{"$gte": start, "$lt": end}

	m2 := []bson.M{
		{"$match": m1Match},
		{"$project": bson.M{
			"_id":       "$_id",
			"ctime":     "$ctime",
			"win_rates": "$win_rates",
		}},
	}

	var r1 []*data.User
	err = data.PlayerUsers.Pipe(m2).All(&r1)
	if err != nil {
		glog.Error("users error: ", err)
		return
	}

	// r1 := []*data.User{&data.User{Userid: "6614250"}}

	// var datas []map[string]any
	var datas = make(map[string]map[int32][2]int, len(r1)) // <userid, <gtype, [win, lose]>>
	var total, total2 = len(r1), 0
	var pct = 0
	startSec := time.Now().Unix()
	for _, user := range r1 {
		// 查用户前10局游戏流水
		m := []bson.M{
			{"$match": bson.M{
				"userid":   user.Userid,
				"ltype":    bson.M{"$in": []int32{110, 111, 75, 67, 68, 69, 76, 70, 71, 72, 116, 117, 114, 115, 112, 113, 79, 73, 87, 88, 89, 90, 91, 92, 100, 101, 106, 107, 108, 109, 127, 128, 93, 94, 95, 105}},
				"ctime":    bson.M{"$gt": start}, // , "$lt": end
				"water_id": bson.M{"$exists": true, "$ne": ""},
			}},
			{"$sort": bson.M{"ctime": 1}},
			{"$skip": 0},
			{"$limit": 200},
		}
		var waters []data.LogWater
		err = data.LogWaters.Pipe(m).All(&waters)
		if err != nil {
			glog.Error("find log water error: %s %v", user.Userid, err)
			continue
		}
		// var curWaterId string
		// var curWaters []data.LogWater
		var round int

		// todo sort check
		sort.Slice(waters, func(i, j int) bool {
			return waters[i].Ctime.Before(waters[j].Ctime)
		})
		var watersGroup = make(map[string][]data.LogWater)
		for _, water := range waters {
			watersGroup[water.WaterId] = append(watersGroup[water.WaterId], water)
		}
		for _, water := range waters {
			curWaters, ok := watersGroup[water.WaterId]
			if !ok {
				continue
			}
			delete(watersGroup, water.WaterId)

			// next round
			// if curWaterId == water.WaterId {
			// 	curWaters = append(curWaters, water)
			// 	continue
			// }

			// 计算这一局
			if len(curWaters) > 0 {
				// // todo sort check
				// sort.Slice(curWaters, func(i, j int) bool {
				// 	return curWaters[i].Ctime.Before(curWaters[j].Ctime)
				// })

				ltypes := utils.SliceMapping(curWaters, func(w data.LogWater) int32 { return w.LType })
				lastLtype := ltypes[len(ltypes)-1]
				for i := len(ltypes) - 1; i >= 0; i-- {
					if !utils.SliceIn(ltypes[i], 69, 72, 89, 92, 109) { // 扣税
						lastLtype = ltypes[i]
						break
					}
				}
				var gtype int32
				var win bool
				if utils.SliceIn(lastLtype, []int32{110, 111}...) { // tp
					gtype = int32(pb.HUA)
					win = lastLtype == 111
				} else if utils.SliceIn(lastLtype, []int32{75, 67, 68, 69}...) { // lhd
					gtype = int32(pb.LHD)
					win = lastLtype == 67
				} else if utils.SliceIn(lastLtype, []int32{76, 70, 71, 72}...) { // 7up
					gtype = int32(pb.SEVEN)
					win = lastLtype == 70
				} else if utils.SliceIn(lastLtype, []int32{116, 117}...) { // rummy
					gtype = int32(pb.RUMMY)
					win = lastLtype == 117
				} else if utils.SliceIn(lastLtype, []int32{114, 115}...) { // ak47
					gtype = int32(pb.AK47)
					win = lastLtype == 115
				} else if utils.SliceIn(lastLtype, []int32{112, 113}...) { // joker
					gtype = int32(pb.JOKER)
					win = lastLtype == 113
				} else if utils.SliceIn(lastLtype, []int32{79, 73}...) { // crash
					gtype = int32(pb.CRASH)
					win = lastLtype == 73
				} else if utils.SliceIn(lastLtype, []int32{87, 88, 89}...) { // ab
					gtype = int32(pb.ABAR)
					win = lastLtype == 88
				} else if utils.SliceIn(lastLtype, []int32{90, 91, 92}...) { // 彩票
					gtype = int32(pb.LOTTERY)
					win = lastLtype == 91
				} else if utils.SliceIn(lastLtype, []int32{100, 101}...) { // plane
					gtype = int32(pb.PLANE)
					win = lastLtype == 101
				} else if utils.SliceIn(lastLtype, []int32{127, 128}...) { // rm2
					gtype = int32(pb.RUMMY2)
					win = lastLtype == 128
				} else if utils.SliceIn(lastLtype, []int32{93, 94, 95, 105}...) { // 外接
					gtype = int32(pb.GAME)
					last := curWaters[len(curWaters)-1]
					win = lastLtype == 95 && last.AddDiamond > 0
				}
				gtypeWins, ok := datas[user.Userid]
				if !ok {
					gtypeWins = make(map[int32][2]int)
					datas[user.Userid] = gtypeWins
				}
				wins, ok := gtypeWins[gtype]
				if !ok {
					wins = [2]int{0, 0}
					gtypeWins[gtype] = wins
				}
				if win {
					wins[0]++ // 赢
				} else {
					wins[1]++ // 输
				}
				gtypeWins[gtype] = wins

				// 只要10局
				round++
				if round >= 10 {
					break
				}
			}

			// reset
			// curWaterId = water.WaterId
			// curWaters = append([]data.LogWater{}, water)
		}

		total2++
		if cur := float64(total2) / float64(total) * 100; int(cur/2) > pct {
			pct++

			useSec := time.Now().Unix() - startSec
			useMin := useSec / 60
			glog.Infof("stats round %.2f%%, %d/%d use %dm%ds", cur, total2, total, useMin, useSec-useMin*60)
		}

		// var win, lose int
		// for _, rate := range user.WinRates {
		// 	win += int(rate.WinRound)
		// 	lose += int(rate.LoseRound)
		// 	// totalRounds += (rate.WinRound + rate.LoseRound)
		// }
		// if totalRounds < 5 {
		// 	datas = append(datas, map[string]any{"userid": user.Userid, "round": totalRounds})
		// }
	}

	f := excelize.NewFile()
	sheet := "Sheet1"
	// 设置表头
	f.SetCellValue(sheet, "A1", "玩家id")
	f.SetCellValue(sheet, "B1", "赢")
	f.SetCellValue(sheet, "C1", "输")
	f.SetCellValue(sheet, "D1", "tp")
	f.SetCellValue(sheet, "E1", "龙虎斗")
	f.SetCellValue(sheet, "F1", "7updown")
	f.SetCellValue(sheet, "G1", "rummy")
	f.SetCellValue(sheet, "H1", "ak47")
	f.SetCellValue(sheet, "I1", "joker")
	f.SetCellValue(sheet, "J1", "crash")
	f.SetCellValue(sheet, "K1", "andarbahar")
	f.SetCellValue(sheet, "L1", "彩票")
	f.SetCellValue(sheet, "M1", "飞机")
	f.SetCellValue(sheet, "N1", "双人rummy")
	f.SetCellValue(sheet, "O1", "外接")
	var gtypeColumn = map[int32]string{
		0:  "O%d",
		1:  "D%d",
		2:  "E%d",
		3:  "F%d",
		4:  "G%d",
		5:  "H%d",
		6:  "I%d",
		7:  "J%d",
		8:  "K%d",
		9:  "L%d",
		10: "M%d",
		12: "N%d",
	}
	line := 1
	for userid, gtypeWins := range datas {
		line++
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", line), userid)
		var winR, loseR int
		for gtype, wins := range gtypeWins {
			winR += wins[0]
			loseR += wins[1]
			f.SetCellValue("Sheet1", fmt.Sprintf(gtypeColumn[gtype], line), wins[0]+wins[1])
		}
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", line), winR)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", line), loseR)
	}

	// 保存文件
	if err := f.SaveAs(getExportFileName("Stat7_1to7_14Round10")); err != nil {
		glog.Error("error3: ", err)
	}
}

// 拉一个数据7月8日-10日，新注册用户，局数低于5局的ID和局数
// 7月1号到7月14日注册的用户，前10局的游戏分布，输赢
func Stat7_1to7_14Round10_2() {
	m1Match := bson.M{
		"robot":            false,
		"simulation_robot": false,
	}
	start, end := utils.Str2Time("2024-07-01 00:00:00", location), utils.Str2Time("2024-07-15 00:00:00", location)
	m1Match["ctime"] = bson.M{"$gte": start, "$lt": end}

	m2 := []bson.M{
		{"$match": m1Match},
		{"$project": bson.M{
			"_id":   "$_id",
			"ctime": "$ctime",
			// "win_rates": "$win_rates",
		}},
	}

	var r1 []*data.User
	err = data.PlayerUsers.Pipe(m2).All(&r1)
	if err != nil {
		glog.Error("users error: ", err)
		return
	}
	// var datas []map[string]any
	// var datas = make(map[string]map[int32][2]int, len(r1)) // <userid, <gtype, [win, lose]>>
	var total, total2 = len(r1), 0
	var pct = 0
	startSec := time.Now().Unix()

	var userDatas = make(map[string][]GameRound)
	for _, user := range r1 {
		// 查10局对局详情
		m := []bson.M{
			{"$match": bson.M{
				"players": bson.M{
					"$regex":   fmt.Sprintf("\\b%s\\b", user.Userid),
					"$options": "i",
				},
			}},
			{"$sort": bson.M{"begin_time": 1}},
			{"$skip": 0},
			{"$limit": 10},
		}
		var details []data.Detail
		err = data.Details.Pipe(m).All(&details)
		// fmt.Printf("use %d ms\n", time.Now().UnixMilli()-startMs)
		if err != nil {
			glog.Error("details error: ", err)
			return
		}
		for _, detail := range details {
			g := GameRound{Ctime: detail.BeginTime, Gtype: detail.Gtype}
			switch detail.Gtype {
			case int32(pb.HUA):
				for _, d := range detail.TPDetail {
					if d.UserId == user.Userid && d.Score > 0 {
						g.Win = true
						break
					}
				}
			case int32(pb.LHD):
				for _, d := range detail.LHDetail.UserDetail {
					if d.Userid == user.Userid && d.Result == "赢" {
						g.Win = true
						break
					}
				}
			case int32(pb.SEVEN):
				for _, d := range detail.UPDetail.UserDetail {
					if d.Userid == user.Userid && d.Result == "赢" {
						g.Win = true
						break
					}
				}
			case int32(pb.RUMMY):
				for _, d := range detail.RMDetail {
					if d.UserId == user.Userid && d.Score > 0 {
						g.Win = true
						break
					}
				}
			case int32(pb.AK47):
				for _, d := range detail.AK47Detail {
					if d.UserId == user.Userid && d.Score > 0 {
						g.Win = true
						break
					}
				}
			case int32(pb.JOKER):
				for _, d := range detail.JOKERDetail {
					if d.UserId == user.Userid && d.Score > 0 {
						g.Win = true
						break
					}
				}
			case int32(pb.CRASH):
				for _, d := range detail.CRASHDetail.UserDetail {
					if d.Userid == user.Userid && d.Result == "赢" {
						g.Win = true
						break
					}
				}
			case int32(pb.ABAR):
				for _, d := range detail.ABDetail.UserDetail {
					if d.Userid == user.Userid && d.Result == "赢" {
						g.Win = true
						break
					}
				}
			case int32(pb.LOTTERY):
				for _, d := range detail.CPDetail.UserDetail {
					if d.Userid == user.Userid && d.Result == "赢" {
						g.Win = true
						break
					}
				}
			case int32(pb.PLANE):
				for _, d := range detail.CRASHDetail.UserDetail {
					if d.Userid == user.Userid && d.Result == "赢" {
						g.Win = true
						break
					}
				}
			case int32(pb.RUMMY2):
				for _, d := range detail.RMDetail {
					if d.UserId == user.Userid && d.Score > 0 {
						g.Win = true
						break
					}
				}
			}
			userDatas[user.Userid] = append(userDatas[user.Userid], g)
		}

		// 查10局外接
		m2 := []bson.M{
			{
				"$match": bson.M{
					// "ctime":   bson.M{"$gte": startTime, "$lt": endTime},
					"user_id": user.Userid,
				},
			},
			{
				"$group": bson.M{
					"_id":   bson.M{"user_id": "$user_id", "round_id": "$round_id"},
					"ctime": bson.M{"$min": "$ctime"},
				},
			},
			{"$sort": bson.M{"ctime": 1}},
			{"$skip": 0},
			{"$limit": 10},
		}
		r2 := []bson.M{}
		err = data.NsqLogExternalBets.Pipe(m2).All(&r2)
		if err != nil {
			glog.Error("extBets error: ", err)
			return
		}
		// 查输赢
		var extRoundWins = make(map[string]bool)
		if len(r2) > 0 {
			var rids []string
			for _, bet := range r2 {
				rids = append(rids, bet["_id"].(bson.M)["round_id"].(string))
			}
			m3 := bson.M{"user_id": user.Userid, "round_id": bson.M{"$in": rids}}
			var r3 []data.NsqLogExternalReward
			err = data.NsqLogExternalRewards.Find(m3).All(&r3)
			if err != nil {
				return
			}
			for _, r := range r3 {
				extRoundWins[r.RoundId] = r.RewardAmount > 0
			}

			for _, r := range r2 {
				roundId := r["_id"].(bson.M)["round_id"].(string)
				ctime := r["ctime"].(int64)
				g := GameRound{
					Ctime: ctime,
					Win:   extRoundWins[roundId],
					Gtype: 0,
				}
				userDatas[user.Userid] = append(userDatas[user.Userid], g)
			}
		}

		total2++
		if cur := float64(total2) / float64(total) * 100; int(cur/2) > pct {
			pct++

			useSec := time.Now().Unix() - startSec
			useMin := useSec / 60
			glog.Infof("stats round %.2f%%, %d/%d use %dm%ds", cur, total2, total, useMin, useSec-useMin*60)
		}
	}
	f := excelize.NewFile()
	sheet := "Sheet1"
	// 设置表头
	f.SetCellValue(sheet, "A1", "玩家id")
	f.SetCellValue(sheet, "B1", "赢")
	f.SetCellValue(sheet, "C1", "输")
	f.SetCellValue(sheet, "D1", "tp")
	f.SetCellValue(sheet, "E1", "龙虎斗")
	f.SetCellValue(sheet, "F1", "7updown")
	f.SetCellValue(sheet, "G1", "rummy")
	f.SetCellValue(sheet, "H1", "ak47")
	f.SetCellValue(sheet, "I1", "joker")
	f.SetCellValue(sheet, "J1", "crash")
	f.SetCellValue(sheet, "K1", "andarbahar")
	f.SetCellValue(sheet, "L1", "彩票")
	f.SetCellValue(sheet, "M1", "飞机")
	f.SetCellValue(sheet, "N1", "双人rummy")
	f.SetCellValue(sheet, "O1", "外接")
	var gtypeColumn = map[int32]string{
		0:  "O%d",
		1:  "D%d",
		2:  "E%d",
		3:  "F%d",
		4:  "G%d",
		5:  "H%d",
		6:  "I%d",
		7:  "J%d",
		8:  "K%d",
		9:  "L%d",
		10: "M%d",
		12: "N%d",
	}
	line := 1

	for userid, games := range userDatas {
		line++
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", line), userid)

		sort.Slice(games, func(i, j int) bool {
			return games[i].Ctime < games[j].Ctime
		})

		var gtypeWins = make(map[int32][2]int)
		for i, g := range games {
			if i >= 10 {
				break
			}
			wins, ok := gtypeWins[g.Gtype]
			if !ok {
				wins = [2]int{}
				// gtypeWins[g.Gtype] = wins
			}
			if g.Win {
				wins[0]++
			} else {
				wins[1]++
			}
			gtypeWins[g.Gtype] = wins
		}

		var winR, loseR int
		for gtype, wins := range gtypeWins {
			winR += wins[0]
			loseR += wins[1]
			f.SetCellValue("Sheet1", fmt.Sprintf(gtypeColumn[gtype], line), wins[0]+wins[1])
		}
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", line), winR)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", line), loseR)
	}

	// 保存文件
	if err := f.SaveAs(getExportFileName("Stat7_1to7_14Round10")); err != nil {
		glog.Error("error3: ", err)
	}
}

type GameRound struct {
	Ctime int64
	Win   bool
	Gtype int32
}
