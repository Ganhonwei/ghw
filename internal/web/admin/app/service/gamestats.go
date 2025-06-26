package service

import (
	"encoding/json"
	"fmt"
	"goserver/internal/web/admin/app/entity"
	"goserver/internal/web/agent/app/service"
	"goserver/pkg/data/ck"
	"goserver/pkg/utils"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type gameStatsService struct {
}

// GetCrashStatPlayers crash玩家明细
func (s *gameStatsService) GetCrashStatPlayers(page, pageSize int, m, m2, m3 bson.M, startDate, endDate string) (
	stats []*entity.CrashPlayerStat, count int, err error,
) {
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)

	// 用户列表筛选
	var filterPlayerIds []string
	if len(m3) > 0 || len(m2) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				"_id":        "$userid",
				"bets":       bson.M{"$sum": "$bets"},
				"bet_rounds": bson.M{"$sum": "$bet_rounds"},
				"cash":       bson.M{"$sum": "$cash"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				"cash": "$cash",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = CrashPlayerStats.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetCrashStatPlayers error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			userid := r["_id"].(string)
			userids = append(userids, userid)
		}
		if len(userids) == 0 { // 未匹配到
			return
		}
		now := utils.BsonNow()
		m3_2 := []bson.M{
			{"$match": bson.M{"_id": bson.M{"$in": userids}}},
			{"$project": bson.M{
				"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
				"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
			}},
			{"$match": m3},
		}
		var r3_2 []bson.M
		err = PlayerUsers.Pipe(m3_2).All(&r3_2)
		if err != nil {
			beego.Error("GetCrashStatPlayers error32:", err)
		} else if len(r3_2) == 0 { // 未匹配到
			return
		}
		for _, r := range r3_2 {
			userid := r["_id"].(string)
			filterPlayerIds = append(filterPlayerIds, userid)
		}
	}
	if len(filterPlayerIds) > 0 {
		m["userid"] = bson.M{"$in": filterPlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":                     "$userid",
			"date":                    bson.M{"$max": "$date"},
			"all_rounds":              bson.M{"$sum": "$all_rounds"},
			"game_times":              bson.M{"$sum": "$game_times"},
			"bets":                    bson.M{"$sum": "$bets"},
			"bet_rounds":              bson.M{"$sum": "$bet_rounds"},
			"observe_rounds":          bson.M{"$sum": "$observe_rounds"},
			"win_rounds":              bson.M{"$sum": "$win_rounds"},
			"lose_rounds":             bson.M{"$sum": "$lose_rounds"},
			"tie_rounds":              bson.M{"$sum": "$tie_rounds"},
			"win_bets":                bson.M{"$sum": "$win_bets"},
			"lose_bets":               bson.M{"$sum": "$lose_bets"},
			"tie_bets":                bson.M{"$sum": "$tie_bets"},
			"wins":                    bson.M{"$sum": "$wins"},
			"loses":                   bson.M{"$sum": "$loses"},
			"cash":                    bson.M{"$sum": "$cash"},
			"mulpitle_sum":            bson.M{"$sum": "$mulpitle_sum"},
			"win_escape_mulpitle_sum": bson.M{"$sum": "$win_escape_mulpitle_sum"},
			"win_mulpitle_sum":        bson.M{"$sum": "$win_mulpitle_sum"},
			"lose_mulpitle_sum":       bson.M{"$sum": "$lose_mulpitle_sum"},
			"mulpitles":               bson.M{"$push": "$mulpitles"},
			"win_escape_mulpitles":    bson.M{"$push": "$win_escape_mulpitles"},
			"win_mulpitles":           bson.M{"$push": "$win_mulpitles"},
			"lose_mulpitles":          bson.M{"$push": "$lose_mulpitles"},
		}},
		{"$project": bson.M{
			"_id": "$_id", "all_rounds": "$all_rounds", "game_times": "$game_times", "bets": "$bets", "bet_rounds": "$bet_rounds", "observe_rounds": "$observe_rounds", "win_rounds": "$win_rounds", "lose_rounds": "$lose_rounds", "tie_rounds": "$tie_rounds", "win_bets": "$win_bets", "lose_bets": "$lose_bets", "tie_bets": "$tie_bets", "wins": "$wins", "loses": "$loses", "cash": "$cash", "mulpitle_sum": "$mulpitle_sum", "win_escape_mulpitle_sum": "$win_escape_mulpitle_sum", "win_mulpitle_sum": "$win_mulpitle_sum", "lose_mulpitle_sum": "$lose_mulpitle_sum", "mulpitles": "$mulpitles", "win_escape_mulpitles": "$win_escape_mulpitles", "win_mulpitles": "$win_mulpitles", "lose_mulpitles": "$lose_mulpitles",
			// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}}, // "can't $divide by zero"
			"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
		}},
		{"$match": m2},
		{"$count": "count"},
	}
	// 查总条数
	var rCount []bson.M
	err = CrashPlayerStats.Pipe(m1).All(&rCount)
	if err != nil {
		return
	} else if len(rCount) > 0 {
		count = rCount[0]["count"].(int)
	}

	// 查列表数据
	m1 = m1[:len(m1)-1]
	m1 = append(m1, []bson.M{
		{"$sort": bson.M{"date": -1}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}...)
	err = CrashPlayerStats.Pipe(m1).All(&stats)
	if err != nil {
		return
	}
	crashStatPlayersMapping(stats)

	// 查总汇数据
	mark := "--"
	summary := &entity.CrashPlayerStat{}
	m1 = m1[:len(m1)-3]
	m1[1]["$group"].(bson.M)["_id"] = nil
	delete(m1[1]["$group"].(bson.M), "date")
	m1[1]["$group"].(bson.M)["userids"] = bson.M{"$addToSet": "$userid"}
	m1[2]["$project"].(bson.M)["userids"] = "$userids"
	var summarys []*entity.CrashPlayerStat
	err = CrashPlayerStats.Pipe(m1).All(&summarys)
	if err != nil {
		return
	} else if len(summarys) > 0 {
		summary = summarys[0]
		summary.No = fmt.Sprintf("%s-%s总汇", startDate, endDate)
		summary.Id = mark
		summary.FNickname = mark
		summary.FUserType = mark
		summary.FRegistArea = mark
		summary.FCtime = mark
		summary.FLoginTime = mark
		// summary.FLiveDays = mark
		// summary.FLoseDays = mark
		crashStatPlayersSummaryMapping(summary)
	}
	if len(summarys) > 0 {
		stats = append([]*entity.CrashPlayerStat{summary}, stats...)
	}
	return
}

func crashStatPlayersMapping(stats []*entity.CrashPlayerStat) {
	var playerIds []string
	// var playerStats = make(map[string]*entity.CrashPlayerStat, len(stats))
	for _, stat := range stats {
		playerIds = append(playerIds, stat.Id)
		// playerStats[stat.Id] = stat
	}
	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": playerIds}}},
		{"$project": bson.M{
			"_id":         "$_id",
			"nickname":    "$nickname",
			"money":       "$money",
			"cash_out":    "$cash_out",
			"diamond":     "$diamond",
			"ctime":       "$ctime",
			"login_time":  "$login_time",
			"regist_area": "$regist_area", //ab测试 0:A 1:B 2:C
			"state":       "$state",
		}},
	}
	var r2 []bson.M
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("crash users error:", err)
		return
	}
	var r2Map = make(map[string]bson.M, len(r2))
	for _, r := range r2 {
		userid := r["_id"].(string)
		r2Map[userid] = r
	}
	now := utils.BsonNow()
	for i, stat := range stats {
		r, ok := r2Map[stat.Id]
		if !ok {
			continue
		}
		// 玩家昵称	玩家类型 账号类型 注册时间 最后登陆时间 存活天数 流失天数
		// userid := r["_id"].(string)
		nickname := r["nickname"].(string)
		money := r["money"].(int)
		cash_out := r["cash_out"].(int)
		diamond := r["diamond"].(int64)
		ctime := r["ctime"].(time.Time)
		login_time := r["login_time"].(time.Time)
		regist_area := r["regist_area"].(int)
		state := r["state"].(int)

		// 将日期转成印度时间
		login_time, _ = ConvertToIndiaTime(login_time.Unix())
		ctime, _ = ConvertToIndiaTime(ctime.Unix())

		stat.No = strconv.Itoa(i + 1)
		stat.FNickname = nickname
		stat.Money = uint32(money)
		stat.FCtime = ctime.Format(utils.FORMAT)
		stat.FLoginTime = login_time.Format(utils.FORMAT)
		stat.FLiveDays = int(math.Ceil(login_time.Sub(ctime).Hours() / 24))
		stat.FLoseDays = int(math.Floor(now.Sub(login_time).Hours() / 24))

		switch regist_area {
		case 0:
			stat.FRegistArea = "A类"
		case 1:
			stat.FRegistArea = "B类"
		case 2:
			stat.FRegistArea = "C类"
		}

		switch state {
		case 1:
			stat.FUserType = "新手"
		case 2:
			if money == 0 {
				stat.FUserType = "零充"
			} else if money >= 20000 && money <= 99999 {
				stat.FUserType = "普充"
			} else if money >= 100000 && money <= 499999 {
				stat.FUserType = "小R"
			} else if money >= 500000 && money <= 999999 {
				stat.FUserType = "中R"
			} else if money >= 1000000 && money <= 9999999 {
				stat.FUserType = "大R"
			} else if money >= 10000000 {
				stat.FUserType = "超大R"
			}
		case 3:
			stat.FUserType = "平民"
		case 4:
			stat.FUserType = "泡沫"
		}

		stat.FGameTimes = fmt.Sprintf("%d分%d秒", stat.GameTimes/60, stat.GameTimes&60)
		stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
		if stat.BetRounds > 0 {
			// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.BetRounds)/100.0)
			stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
		}
		stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
		stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
		if stat.BetRounds > 0 {
			stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
		}
		if stat.Bets > 0 {
			// stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
			stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
		}
		if stat.WinRounds > 0 {
			stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
		}
		stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
		stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
		stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)
		if stat.WinRounds > 0 {
			stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FMulpitleAvg = fmt.Sprintf("%.2f", stat.MulpitleSum/float64(stat.BetRounds))
		}
		if len(stat.Mulpitles) > 0 {
			var mulpitles []float64
			for _, m := range stat.Mulpitles {
				mulpitles = append(mulpitles, m...)
			}
			sort.Slice(mulpitles, func(i, j int) bool {
				return mulpitles[i] < mulpitles[j]
			})
			length := len(mulpitles)
			if length > 0 {
				if length%2 == 1 {
					stat.FMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
				} else {
					// 中间两位算平均
					stat.FMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
				}
			}
		}
		// 赢局局均逃脱倍数
		var winEscapeMulpitleAvg float64
		if stat.WinRounds > 0 {
			winEscapeMulpitleAvg = stat.WinEscapeMulpitleSum / float64(stat.WinRounds)
			stat.FWinEscapeMulpitleAvg = fmt.Sprintf("%.2f", winEscapeMulpitleAvg)
		}
		if len(stat.WinEscapeMulpitles) > 0 {
			var mulpitles []float64
			for _, m := range stat.WinEscapeMulpitles {
				mulpitles = append(mulpitles, m...)
			}
			sort.Slice(mulpitles, func(i, j int) bool {
				return mulpitles[i] < mulpitles[j]
			})
			length := len(mulpitles)
			if length > 0 {
				// 赢局最高逃脱倍数
				stat.FWinEscapeMulpitleMax = fmt.Sprintf("%.2f", mulpitles[len(mulpitles)-1])
				if length%2 == 1 {
					stat.FWinEscapeMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
				} else {
					// 中间两位算平均
					stat.FWinEscapeMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
				}
			}
		}

		// 赢局局均实际爆炸倍数
		var winMulpitleAvg float64
		if stat.WinRounds > 0 {
			winMulpitleAvg = stat.WinMulpitleSum / float64(stat.WinRounds)
			stat.FWinMulpitleAvg = fmt.Sprintf("%.2f", winMulpitleAvg)
		}
		if len(stat.WinMulpitles) > 0 {
			var mulpitles []float64
			for _, m := range stat.WinMulpitles {
				mulpitles = append(mulpitles, m...)
			}
			sort.Slice(mulpitles, func(i, j int) bool {
				return mulpitles[i] < mulpitles[j]
			})
			length := len(mulpitles)
			if length > 0 {
				if length%2 == 1 {
					stat.FWinMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
				} else {
					// 中间两位算平均
					stat.FWinMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
				}
			}
		}
		// 倍差空间
		stat.FMulpitleDiff = fmt.Sprintf("%.2f", winMulpitleAvg-winEscapeMulpitleAvg)

		// 输局局均实际爆炸倍数
		var loseMulpitleAvg float64
		if stat.LoseRounds > 0 {
			loseMulpitleAvg = stat.LoseMulpitleSum / float64(stat.LoseRounds)
			stat.FLoseMulpitleAvg = fmt.Sprintf("%.2f", loseMulpitleAvg)
		}
		if len(stat.LoseMulpitles) > 0 {
			var mulpitles []float64
			for _, m := range stat.LoseMulpitles {
				mulpitles = append(mulpitles, m...)
			}
			sort.Slice(mulpitles, func(i, j int) bool {
				return mulpitles[i] < mulpitles[j]
			})
			length := len(mulpitles)
			if length > 0 {
				if length%2 == 1 {
					stat.FLoseMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
				} else {
					// 中间两位算平均
					stat.FLoseMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
				}
			}
		}
		if loseMulpitleAvg > 0 {
			stat.FWinEscapeRate = fmt.Sprintf("%.2f", winEscapeMulpitleAvg/loseMulpitleAvg)
			stat.FWinLoseRate = fmt.Sprintf("%.2f", winMulpitleAvg/loseMulpitleAvg)
		}
		if stat.BetRounds > 0 {
			stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
		}
		if stat.BetRounds > 0 {
			stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
		}
		stat.FMoney = fmt.Sprintf("%.2f", float64(money)/100)
		stat.FCashOut = fmt.Sprintf("%.2f", float64(cash_out)/100)
		stat.FWinMoney = fmt.Sprintf("%.2f", float64(int(diamond)+cash_out-money)/100)
	}
}

func crashStatPlayersSummaryMapping(stat *entity.CrashPlayerStat) {

	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": stat.Userids}}},
		{"$group": bson.M{
			"_id":      nil,
			"money":    bson.M{"$sum": "$money"},
			"cash_out": bson.M{"$sum": "$cash_out"},
			"diamond":  bson.M{"$sum": "$diamond"},
		}},
	}
	var r2 []bson.M
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("crash users error:", err)
		return
	} else if len(r2) > 0 {
		money := r2[0]["money"].(int)
		cash_out := r2[0]["cash_out"].(int)
		diamond := r2[0]["diamond"].(int64)
		stat.SMoney = money
		stat.SCashOut = cash_out
		stat.SDiamond = diamond
	}

	stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
	if stat.BetRounds > 0 {
		stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.BetRounds)/100.0)
		// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
	}

	stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
	stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
	if stat.BetRounds > 0 {
		stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
	}
	if stat.Bets > 0 {
		stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
	}
	if stat.WinRounds > 0 {
		stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
	}
	stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
	stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
	stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)
	if stat.WinRounds > 0 {
		stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FMulpitleAvg = fmt.Sprintf("%.2f", stat.MulpitleSum/float64(stat.BetRounds))
	}
	if len(stat.Mulpitles) > 0 {
		var mulpitles []float64
		for _, m := range stat.Mulpitles {
			mulpitles = append(mulpitles, m...)
		}
		sort.Slice(mulpitles, func(i, j int) bool {
			return mulpitles[i] < mulpitles[j]
		})
		length := len(mulpitles)
		if length > 0 {
			if length%2 == 1 {
				stat.FMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
			} else {
				// 中间两位算平均
				stat.FMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
			}
		}
	}
	// 赢局局均逃脱倍数
	var winEscapeMulpitleAvg float64
	if stat.WinRounds > 0 {
		winEscapeMulpitleAvg = stat.WinEscapeMulpitleSum / float64(stat.WinRounds)
		stat.FWinEscapeMulpitleAvg = fmt.Sprintf("%.2f", winEscapeMulpitleAvg)
	}
	if len(stat.WinEscapeMulpitles) > 0 {
		var mulpitles []float64
		for _, m := range stat.WinEscapeMulpitles {
			mulpitles = append(mulpitles, m...)
		}
		sort.Slice(mulpitles, func(i, j int) bool {
			return mulpitles[i] < mulpitles[j]
		})
		length := len(mulpitles)
		if length > 0 {
			// 赢局最高逃脱倍数
			stat.FWinEscapeMulpitleMax = fmt.Sprintf("%.2f", mulpitles[len(mulpitles)-1])
			if length%2 == 1 {
				stat.FWinEscapeMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
			} else {
				// 中间两位算平均
				stat.FWinEscapeMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
			}
		}
	}

	// 赢局局均实际爆炸倍数
	var winMulpitleAvg float64
	if stat.WinRounds > 0 {
		winMulpitleAvg = stat.WinMulpitleSum / float64(stat.WinRounds)
		stat.FWinMulpitleAvg = fmt.Sprintf("%.2f", winMulpitleAvg)
	}
	if len(stat.WinMulpitles) > 0 {
		var mulpitles []float64
		for _, m := range stat.WinMulpitles {
			mulpitles = append(mulpitles, m...)
		}
		sort.Slice(mulpitles, func(i, j int) bool {
			return mulpitles[i] < mulpitles[j]
		})
		length := len(mulpitles)
		if length > 0 {
			if length%2 == 1 {
				stat.FWinMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
			} else {
				// 中间两位算平均
				stat.FWinMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
			}
		}
	}
	// 倍差空间
	stat.FMulpitleDiff = fmt.Sprintf("%.2f", winMulpitleAvg-winEscapeMulpitleAvg)

	// 输局局均实际爆炸倍数
	var loseMulpitleAvg float64
	if stat.LoseRounds > 0 {
		loseMulpitleAvg = stat.LoseMulpitleSum / float64(stat.LoseRounds)
		stat.FLoseMulpitleAvg = fmt.Sprintf("%.2f", loseMulpitleAvg)
	}
	if len(stat.LoseMulpitles) > 0 {
		var mulpitles []float64
		for _, m := range stat.LoseMulpitles {
			mulpitles = append(mulpitles, m...)
		}
		sort.Slice(mulpitles, func(i, j int) bool {
			return mulpitles[i] < mulpitles[j]
		})
		length := len(mulpitles)
		if length > 0 {
			if length%2 == 1 {
				stat.FLoseMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
			} else {
				// 中间两位算平均
				stat.FLoseMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
			}
		}
	}
	if loseMulpitleAvg > 0 {
		stat.FWinEscapeRate = fmt.Sprintf("%.2f", winEscapeMulpitleAvg/loseMulpitleAvg)
		stat.FWinLoseRate = fmt.Sprintf("%.2f", winMulpitleAvg/loseMulpitleAvg)
	}
	if stat.BetRounds > 0 {
		stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
	}
	if stat.BetRounds > 0 {
		stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
	}
	stat.FMoney = fmt.Sprintf("%.2f", float64(stat.SMoney)/100)
	stat.FCashOut = fmt.Sprintf("%.2f", float64(stat.SCashOut)/100)
	stat.FWinMoney = fmt.Sprintf("%.2f", float64(int(stat.SDiamond)+stat.SCashOut-stat.SMoney)/100)
}

// GetCrashStatDates crash汇总明细
func (s *gameStatsService) GetCrashStatDates(m, m2, m3 bson.M, startDate, endDate string) (stats []*entity.CrashPlayerStat, err error) {
	// 用户列表筛选
	var filterDatePlayerIds []string
	if len(m2) > 0 || len(m3) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				// "_id":        bson.M{"userid":"$userid", "date", "$date"},
				"_id":        "$_id",
				"bets":       bson.M{"$sum": "$bets"},
				"bet_rounds": bson.M{"$sum": "$bet_rounds"},
				"cash":       bson.M{"$sum": "$cash"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				"cash": "$cash",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = CrashPlayerStats.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetCrashStatPlayers error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			id := r["_id"].(string)
			filterDatePlayerIds = append(filterDatePlayerIds, id)
			ids := strings.Split(id, "-") // id=date-userid
			if len(ids) >= 2 {
				userids = append(userids, ids[1])
			}
		}
		if len(filterDatePlayerIds) == 0 { // 未匹配到
			return
		}
		if len(m3) > 0 { // 流失天数，局均打码量
			if len(userids) == 0 { // 未匹配到
				return
			}
			now := utils.BsonNow()
			m3_2 := []bson.M{
				{"$match": bson.M{"_id": bson.M{"$in": userids}}},
				{"$project": bson.M{
					"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
					"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
				}},
				{"$match": m3},
			}
			var r3_2 []bson.M
			err = PlayerUsers.Pipe(m3_2).All(&r3_2)
			if err != nil {
				beego.Error("GetCrashStatPlayers error32:", err)
			} else if len(r3_2) == 0 { // 未匹配到
				return
			}
			var filterId2 []string
			for _, id := range filterDatePlayerIds {
				for _, r := range r3_2 {
					userid := r["_id"].(string)
					if strings.HasSuffix(id, "-"+userid) {
						filterId2 = append(filterId2, id)
						break
					}
				}
			}
			if len(filterId2) == 0 { // 未匹配到
				return
			}
			filterDatePlayerIds = filterId2
		}
	}
	if len(filterDatePlayerIds) > 0 {
		m["_id"] = bson.M{"$in": filterDatePlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":                     "$date_str",
			"date":                    bson.M{"$min": "$date"},
			"players":                 bson.M{"$sum": 1},
			"new_players":             bson.M{"$sum": bson.M{"$cond": bson.M{"if": "$new_reg", "then": 1, "else": 0}}},
			"old_players":             bson.M{"$sum": bson.M{"$cond": bson.M{"if": "$new_reg", "then": 0, "else": 1}}},
			"all_rounds_list":         bson.M{"$push": "$bet_rounds"},
			"all_rounds":              bson.M{"$sum": "$all_rounds"},
			"game_times":              bson.M{"$sum": "$game_times"},
			"bets":                    bson.M{"$sum": "$bets"},
			"bet_rounds":              bson.M{"$sum": "$bet_rounds"},
			"observe_rounds":          bson.M{"$sum": "$observe_rounds"},
			"win_rounds":              bson.M{"$sum": "$win_rounds"},
			"lose_rounds":             bson.M{"$sum": "$lose_rounds"},
			"tie_rounds":              bson.M{"$sum": "$tie_rounds"},
			"win_bets":                bson.M{"$sum": "$win_bets"},
			"lose_bets":               bson.M{"$sum": "$lose_bets"},
			"tie_bets":                bson.M{"$sum": "$tie_bets"},
			"wins":                    bson.M{"$sum": "$wins"},
			"loses":                   bson.M{"$sum": "$loses"},
			"cash":                    bson.M{"$sum": "$cash"},
			"mulpitle_sum":            bson.M{"$sum": "$mulpitle_sum"},
			"win_escape_mulpitle_sum": bson.M{"$sum": "$win_escape_mulpitle_sum"},
			"win_mulpitle_sum":        bson.M{"$sum": "$win_mulpitle_sum"},
			"lose_mulpitle_sum":       bson.M{"$sum": "$lose_mulpitle_sum"},
			"mulpitles":               bson.M{"$push": "$mulpitles"},
			"win_escape_mulpitles":    bson.M{"$push": "$win_escape_mulpitles"},
			"win_mulpitles":           bson.M{"$push": "$win_mulpitles"},
			"lose_mulpitles":          bson.M{"$push": "$lose_mulpitles"},
			"round_bet_avg":           bson.M{"$sum": "$round_bet_avg"},
		}},
		{"$sort": bson.M{"date": -1}},
	}
	err = CrashPlayerStats.Pipe(m1).All(&stats)
	if err != nil {
		return
	}

	// 查询日活数据
	var dateLoginedUsers = make(map[int64]int)
	m9 := []bson.M{
		{"$match": bson.M{"date": m["date"]}},
		{"$project": bson.M{"date": "$date", "logined_users": "$logined_users"}},
	}
	var r9 []bson.M
	err = PddStats.Pipe(m9).All(&r9)
	if err != nil {
		beego.Error("GetCrashStatDates logined_users error: ", err)
	} else {
		for _, r := range r9 {
			date := r["date"].(int64)
			logined_users := r["logined_users"].(int)
			dateLoginedUsers[date] = logined_users
		}
	}

	mark := "--"
	summary := &entity.CrashPlayerStat{
		DateStr: fmt.Sprintf("%s-%s总汇", startDate, endDate),
		Id:      mark,
	}
	crashStatDatesMapping(summary, stats, dateLoginedUsers)
	crashStatDatesSummaryMapping(summary)
	stats = append([]*entity.CrashPlayerStat{summary}, stats...)
	return
}

func crashStatDatesMapping(summary *entity.CrashPlayerStat, stats []*entity.CrashPlayerStat, dateLoginedUsers map[int64]int) {
	for _, stat := range stats {
		dateStr := stat.Id
		date, _ := utils.Unix(dateStr)
		stat.Date = date
		stat.DateStr = dateStr
		loginedUsers := dateLoginedUsers[stat.Date] // 日活

		stat.DateStr = utils.Stamp2Time(stat.Date).Format("2006-01-02")
		stat.FLoginedUsers = loginedUsers

		summary.FLoginedUsers += stat.FLoginedUsers
		summary.FLiveDays += stat.FLiveDays
		summary.FLoseDays += stat.FLoseDays
		summary.AllRounds += stat.AllRounds
		summary.GameTimes += stat.GameTimes
		summary.Bets += stat.Bets
		summary.BetRounds += stat.BetRounds
		summary.WinRounds += stat.WinRounds
		summary.LoseRounds += stat.LoseRounds
		summary.TieRounds += stat.TieRounds
		summary.WinBets += stat.WinBets
		summary.LoseBets += stat.LoseBets
		summary.TieBets += stat.TieBets
		summary.Wins += stat.Wins
		summary.Loses += stat.Loses
		summary.Cash += stat.Cash
		summary.MulpitleSum += stat.MulpitleSum
		summary.WinEscapeMulpitleSum += stat.WinEscapeMulpitleSum
		summary.WinMulpitleSum += stat.WinMulpitleSum
		summary.LoseMulpitleSum += stat.LoseMulpitleSum
		summary.Mulpitles = append(summary.Mulpitles, stat.Mulpitles...)
		summary.WinEscapeMulpitles = append(summary.WinEscapeMulpitles, stat.WinEscapeMulpitles...)
		summary.WinMulpitles = append(summary.WinMulpitles, stat.WinMulpitles...)
		summary.LoseMulpitles = append(summary.LoseMulpitles, stat.LoseMulpitles...)
		summary.ObserveRounds += stat.ObserveRounds
		summary.Players += stat.Players
		summary.NewPlayers += stat.NewPlayers
		summary.OldPlayers += stat.OldPlayers
		summary.AllRoundsList = append(summary.AllRoundsList, stat.AllRoundsList...)

		// 玩 crash 人数占日活比
		if stat.FLoginedUsers > 0 {
			stat.FPlayerRate = fmt.Sprintf("%.2f", float64(stat.Players)/float64(stat.FLoginedUsers))
		}

		// 查询当天注册的新用户
		startTime := utils.Str2Time(fmt.Sprintf("%s 00:00:00", dateStr), Location())
		endTime := utils.Str2Time(fmt.Sprintf("%s 23:59:59", dateStr), Location())
		m7 := []bson.M{
			{"$match": bson.M{
				"ctime":            bson.M{"$gte": startTime, "$lte": endTime},
				"robot":            false,
				"simulation_robot": false,
			}},
			{"$group": bson.M{
				"_id":   nil,
				"total": bson.M{"$sum": 1},
			}},
		}
		var r7 []bson.M
		err := PlayerUsers.Pipe(m7).All(&r7)
		if err != nil {
			beego.Error("查询当天注册的新用户 error: ", err)
		} else if len(r7) > 0 {
			stat.FLoginedNewUsers = r7[0]["total"].(int)
		}
		stat.FLoginedOldUsers = stat.FLoginedUsers - stat.FLoginedNewUsers
		if stat.FLoginedOldUsers < 0 {
			stat.FLoginedOldUsers = 0
		}
		summary.FLoginedNewUsers += stat.FLoginedNewUsers
		summary.FLoginedOldUsers += stat.FLoginedOldUsers
		// FNewPlayerRate, FOldPlayerRate 占新老玩家日活
		if stat.FLoginedNewUsers > 0 {
			stat.FNewPlayerRate = fmt.Sprintf("%.2f", float64(stat.NewPlayers)/float64(stat.FLoginedNewUsers))
		}
		if stat.FLoginedOldUsers > 0 {
			stat.FOldPlayerRate = fmt.Sprintf("%.2f", float64(stat.OldPlayers)/float64(stat.FLoginedOldUsers))
		}

		if stat.Players > 0 {
			stat.FPlayerRoundsAvg = fmt.Sprintf("%.2f", float64(stat.BetRounds)/float64(stat.Players))
		}
		// 局数中位数
		if len(stat.AllRoundsList) > 0 {
			var rounds = stat.AllRoundsList
			sort.Slice(rounds, func(i, j int) bool {
				return rounds[i] < rounds[j]
			})
			length := len(rounds)
			if length > 0 {
				if length%2 == 1 {
					stat.FPlayerRoundsMedian = rounds[length/2]
				} else {
					// 中间两位算平均
					stat.FPlayerRoundsMedian = (rounds[length/2] + rounds[length/2-1]) / 2
				}
			}
		}
		stat.FGameTimes = fmt.Sprintf("%.2f", float64(stat.GameTimes)/60)
		if stat.BetRounds > 0 {
			stat.FGameTimesAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.BetRounds)/60)
		}
		if stat.Players > 0 {
			stat.FGameTimesPlayerAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.Players)/60)
		}
		stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
		if stat.Players > 0 {
			stat.FPlayerBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.Players)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.BetRounds)/100.0)
			// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
		}
		if stat.Bets > 0 {
			stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
		}
		stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
		stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
		if stat.WinRounds > 0 {
			stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
		}
		stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
		stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
		stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)
		if stat.WinRounds > 0 {
			stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FMulpitleAvg = fmt.Sprintf("%.2f", stat.MulpitleSum/float64(stat.BetRounds))
		}
		if len(stat.Mulpitles) > 0 {
			var mulpitles []float64
			for _, m := range stat.Mulpitles {
				mulpitles = append(mulpitles, m...)
			}
			sort.Slice(mulpitles, func(i, j int) bool {
				return mulpitles[i] < mulpitles[j]
			})
			length := len(mulpitles)
			if length > 0 {
				if length%2 == 1 {
					stat.FMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
				} else {
					// 中间两位算平均
					stat.FMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
				}
			}
		}

		// 赢局局均逃脱倍数
		var winEscapeMulpitleAvg float64
		if stat.WinRounds > 0 {
			winEscapeMulpitleAvg = stat.WinEscapeMulpitleSum / float64(stat.WinRounds)
			stat.FWinEscapeMulpitleAvg = fmt.Sprintf("%.2f", winEscapeMulpitleAvg)
		}
		if len(stat.WinEscapeMulpitles) > 0 {
			var mulpitles []float64
			for _, m := range stat.WinEscapeMulpitles {
				mulpitles = append(mulpitles, m...)
			}
			sort.Slice(mulpitles, func(i, j int) bool {
				return mulpitles[i] < mulpitles[j]
			})
			length := len(mulpitles)
			if length > 0 {
				// 赢局最高逃脱倍数
				stat.FWinEscapeMulpitleMax = fmt.Sprintf("%.2f", mulpitles[len(mulpitles)-1])
				if length%2 == 1 {
					stat.FWinEscapeMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
				} else {
					// 中间两位算平均
					stat.FWinEscapeMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
				}
			}
		}
		// 赢局局均实际爆炸倍数
		var winMulpitleAvg float64
		if stat.WinRounds > 0 {
			winMulpitleAvg = stat.WinMulpitleSum / float64(stat.WinRounds)
			stat.FWinMulpitleAvg = fmt.Sprintf("%.2f", winMulpitleAvg)
		}
		if len(stat.WinMulpitles) > 0 {
			var mulpitles []float64
			for _, m := range stat.WinMulpitles {
				mulpitles = append(mulpitles, m...)
			}
			sort.Slice(mulpitles, func(i, j int) bool {
				return mulpitles[i] < mulpitles[j]
			})
			length := len(mulpitles)
			if length > 0 {
				if length%2 == 1 {
					stat.FWinMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
				} else {
					// 中间两位算平均
					stat.FWinMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
				}
			}
		}
		// 倍差空间
		stat.FMulpitleDiff = fmt.Sprintf("%.2f", winMulpitleAvg-winEscapeMulpitleAvg)

		// 输局局均实际爆炸倍数
		var loseMulpitleAvg float64
		if stat.LoseRounds > 0 {
			loseMulpitleAvg = stat.LoseMulpitleSum / float64(stat.LoseRounds)
			stat.FLoseMulpitleAvg = fmt.Sprintf("%.2f", loseMulpitleAvg)
		}
		if len(stat.LoseMulpitles) > 0 {
			var mulpitles []float64
			for _, m := range stat.LoseMulpitles {
				mulpitles = append(mulpitles, m...)
			}
			sort.Slice(mulpitles, func(i, j int) bool {
				return mulpitles[i] < mulpitles[j]
			})
			length := len(mulpitles)
			if length > 0 {
				if length%2 == 1 {
					stat.FLoseMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
				} else {
					// 中间两位算平均
					stat.FLoseMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
				}
			}
		}
		if loseMulpitleAvg > 0 {
			stat.FWinEscapeRate = fmt.Sprintf("%.2f", winEscapeMulpitleAvg/loseMulpitleAvg)
			stat.FWinLoseRate = fmt.Sprintf("%.2f", winMulpitleAvg/loseMulpitleAvg)
		}
		if stat.BetRounds > 0 {
			stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
		}
		if stat.BetRounds > 0 {
			stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
		}
	}
}

func crashStatDatesSummaryMapping(stat *entity.CrashPlayerStat) {
	// 玩 crash 人数占日活比
	if stat.FLoginedUsers > 0 {
		stat.FPlayerRate = fmt.Sprintf("%.2f", float64(stat.Players)/float64(stat.FLoginedUsers))
	}

	// FNewPlayerRate, FOldPlayerRate 占新老玩家日活
	if stat.FLoginedNewUsers > 0 {
		stat.FNewPlayerRate = fmt.Sprintf("%.2f", float64(stat.NewPlayers)/float64(stat.FLoginedNewUsers))
	}
	if stat.FLoginedOldUsers > 0 {
		stat.FOldPlayerRate = fmt.Sprintf("%.2f", float64(stat.OldPlayers)/float64(stat.FLoginedOldUsers))
	}

	if stat.Players > 0 {
		stat.FPlayerRoundsAvg = fmt.Sprintf("%.2f", float64(stat.BetRounds)/float64(stat.Players))
	}
	// 局数中位数
	if len(stat.AllRoundsList) > 0 {
		var rounds = stat.AllRoundsList
		sort.Slice(rounds, func(i, j int) bool {
			return rounds[i] < rounds[j]
		})
		length := len(rounds)
		if length > 0 {
			if length%2 == 1 {
				stat.FPlayerRoundsMedian = rounds[length/2]
			} else {
				// 中间两位算平均
				stat.FPlayerRoundsMedian = (rounds[length/2] + rounds[length/2-1]) / 2
			}
		}
	}
	stat.FGameTimes = fmt.Sprintf("%.2f", float64(stat.GameTimes)/60)
	if stat.BetRounds > 0 {
		stat.FGameTimesAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.BetRounds)/60)
	}
	if stat.Players > 0 {
		stat.FGameTimesPlayerAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.Players)/60)
	}
	stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
	if stat.Players > 0 {
		stat.FPlayerBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.Players)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.BetRounds)/100.0)
		// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
	}
	if stat.Bets > 0 {
		stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
	}
	stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
	stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
	if stat.WinRounds > 0 {
		stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
	}
	stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
	stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
	stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)
	if stat.WinRounds > 0 {
		stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FMulpitleAvg = fmt.Sprintf("%.2f", stat.MulpitleSum/float64(stat.BetRounds))
	}
	if len(stat.Mulpitles) > 0 {
		var mulpitles []float64
		for _, m := range stat.Mulpitles {
			mulpitles = append(mulpitles, m...)
		}
		sort.Slice(mulpitles, func(i, j int) bool {
			return mulpitles[i] < mulpitles[j]
		})
		length := len(mulpitles)
		if length > 0 {
			if length%2 == 1 {
				stat.FMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
			} else {
				// 中间两位算平均
				stat.FMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
			}
		}
	}
	// 赢局局均逃脱倍数
	var winEscapeMulpitleAvg float64
	if stat.WinRounds > 0 {
		winEscapeMulpitleAvg = stat.WinEscapeMulpitleSum / float64(stat.WinRounds)
		stat.FWinEscapeMulpitleAvg = fmt.Sprintf("%.2f", winEscapeMulpitleAvg)
	}
	if len(stat.WinEscapeMulpitles) > 0 {
		var mulpitles []float64
		for _, m := range stat.WinEscapeMulpitles {
			mulpitles = append(mulpitles, m...)
		}
		sort.Slice(mulpitles, func(i, j int) bool {
			return mulpitles[i] < mulpitles[j]
		})
		length := len(mulpitles)
		if length > 0 {
			// 赢局最高逃脱倍数
			stat.FWinEscapeMulpitleMax = fmt.Sprintf("%.2f", mulpitles[len(mulpitles)-1])
			if length%2 == 1 {
				stat.FWinEscapeMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
			} else {
				// 中间两位算平均
				stat.FWinEscapeMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
			}
		}
	}
	// 赢局局均实际爆炸倍数
	var winMulpitleAvg float64
	if stat.WinRounds > 0 {
		winMulpitleAvg = stat.WinMulpitleSum / float64(stat.WinRounds)
		stat.FWinMulpitleAvg = fmt.Sprintf("%.2f", winMulpitleAvg)
	}
	if len(stat.WinMulpitles) > 0 {
		var mulpitles []float64
		for _, m := range stat.WinMulpitles {
			mulpitles = append(mulpitles, m...)
		}
		sort.Slice(mulpitles, func(i, j int) bool {
			return mulpitles[i] < mulpitles[j]
		})
		length := len(mulpitles)
		if length > 0 {
			if length%2 == 1 {
				stat.FWinMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
			} else {
				// 中间两位算平均
				stat.FWinMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
			}
		}
	}
	// 倍差空间
	stat.FMulpitleDiff = fmt.Sprintf("%.2f", winMulpitleAvg-winEscapeMulpitleAvg)

	// 输局局均实际爆炸倍数
	var loseMulpitleAvg float64
	if stat.LoseRounds > 0 {
		loseMulpitleAvg = stat.LoseMulpitleSum / float64(stat.LoseRounds)
		stat.FLoseMulpitleAvg = fmt.Sprintf("%.2f", loseMulpitleAvg)
	}
	if len(stat.LoseMulpitles) > 0 {
		var mulpitles []float64
		for _, m := range stat.LoseMulpitles {
			mulpitles = append(mulpitles, m...)
		}
		sort.Slice(mulpitles, func(i, j int) bool {
			return mulpitles[i] < mulpitles[j]
		})
		length := len(mulpitles)
		if length > 0 {
			if length%2 == 1 {
				stat.FLoseMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
			} else {
				// 中间两位算平均
				stat.FLoseMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
			}
		}
	}
	if loseMulpitleAvg > 0 {
		stat.FWinEscapeRate = fmt.Sprintf("%.2f", winEscapeMulpitleAvg/loseMulpitleAvg)
		stat.FWinLoseRate = fmt.Sprintf("%.2f", winMulpitleAvg/loseMulpitleAvg)
	}
	if stat.BetRounds > 0 {
		stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
	}
	if stat.BetRounds > 0 {
		stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
	}
}

// GetPlaneStatPlayers 飞机玩家明细
func (s *gameStatsService) GetPlaneStatPlayers(page, pageSize int, m, m2, m3 bson.M, startDate, endDate string) (
	stats []*entity.CrashPlayerStat, count int, err error,
) {
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)

	// 用户列表筛选
	var filterPlayerIds []string
	if len(m3) > 0 || len(m2) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				"_id":        "$userid",
				"bets":       bson.M{"$sum": "$bets"},
				"bet_rounds": bson.M{"$sum": "$bet_rounds"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = PlanePlayerStats.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetCrashStatPlayers error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			userid := r["_id"].(string)
			userids = append(userids, userid)
		}
		if len(userids) == 0 { // 未匹配到
			return
		}
		now := utils.BsonNow()
		m3_2 := []bson.M{
			{"$match": bson.M{"_id": bson.M{"$in": userids}}},
			{"$project": bson.M{
				"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
				"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
			}},
			{"$match": m3},
		}
		var r3_2 []bson.M
		err = PlayerUsers.Pipe(m3_2).All(&r3_2)
		if err != nil {
			beego.Error("GetCrashStatPlayers error32:", err)
		} else if len(r3_2) == 0 { // 未匹配到
			return
		}
		for _, r := range r3_2 {
			userid := r["_id"].(string)
			filterPlayerIds = append(filterPlayerIds, userid)
		}
	}
	if len(filterPlayerIds) > 0 {
		m["userid"] = bson.M{"$in": filterPlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":                     "$userid",
			"date":                    bson.M{"$max": "$date"},
			"all_rounds":              bson.M{"$sum": "$all_rounds"},
			"game_times":              bson.M{"$sum": "$game_times"},
			"bets":                    bson.M{"$sum": "$bets"},
			"bet_rounds":              bson.M{"$sum": "$bet_rounds"},
			"observe_rounds":          bson.M{"$sum": "$observe_rounds"},
			"win_rounds":              bson.M{"$sum": "$win_rounds"},
			"lose_rounds":             bson.M{"$sum": "$lose_rounds"},
			"tie_rounds":              bson.M{"$sum": "$tie_rounds"},
			"win_bets":                bson.M{"$sum": "$win_bets"},
			"lose_bets":               bson.M{"$sum": "$lose_bets"},
			"tie_bets":                bson.M{"$sum": "$tie_bets"},
			"wins":                    bson.M{"$sum": "$wins"},
			"loses":                   bson.M{"$sum": "$loses"},
			"cash":                    bson.M{"$sum": "$cash"},
			"mulpitle_sum":            bson.M{"$sum": "$mulpitle_sum"},
			"win_escape_mulpitle_sum": bson.M{"$sum": "$win_escape_mulpitle_sum"},
			"win_mulpitle_sum":        bson.M{"$sum": "$win_mulpitle_sum"},
			"lose_mulpitle_sum":       bson.M{"$sum": "$lose_mulpitle_sum"},
			"mulpitles":               bson.M{"$push": "$mulpitles"},
			"win_escape_mulpitles":    bson.M{"$push": "$win_escape_mulpitles"},
			"win_mulpitles":           bson.M{"$push": "$win_mulpitles"},
			"lose_mulpitles":          bson.M{"$push": "$lose_mulpitles"},
		}},
		{"$project": bson.M{
			"_id": "$_id", "all_rounds": "$all_rounds", "game_times": "$game_times", "bets": "$bets", "bet_rounds": "$bet_rounds", "observe_rounds": "$observe_rounds", "win_rounds": "$win_rounds", "lose_rounds": "$lose_rounds", "tie_rounds": "$tie_rounds", "win_bets": "$win_bets", "lose_bets": "$lose_bets", "tie_bets": "$tie_bets", "wins": "$wins", "loses": "$loses", "cash": "$cash", "mulpitle_sum": "$mulpitle_sum", "win_escape_mulpitle_sum": "$win_escape_mulpitle_sum", "win_mulpitle_sum": "$win_mulpitle_sum", "lose_mulpitle_sum": "$lose_mulpitle_sum", "mulpitles": "$mulpitles", "win_escape_mulpitles": "$win_escape_mulpitles", "win_mulpitles": "$win_mulpitles", "lose_mulpitles": "$lose_mulpitles",
			// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}}, // "can't $divide by zero"
			"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
		}},
		{"$match": m2},
		{"$count": "count"},
	}
	// 查总条数
	var rCount []bson.M
	err = PlanePlayerStats.Pipe(m1).All(&rCount)
	if err != nil {
		return
	} else if len(rCount) > 0 {
		count = rCount[0]["count"].(int)
	}

	// 查列表数据
	m1 = m1[:len(m1)-1]
	m1 = append(m1, []bson.M{
		{"$sort": bson.M{"date": -1}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}...)
	err = PlanePlayerStats.Pipe(m1).All(&stats)
	if err != nil {
		return
	}
	planeStatPlayersMapping(stats)

	// 查总汇数据
	mark := "--"
	summary := &entity.CrashPlayerStat{}
	m1 = m1[:len(m1)-3]
	m1[1]["$group"].(bson.M)["_id"] = nil
	delete(m1[1]["$group"].(bson.M), "date")
	m1[1]["$group"].(bson.M)["userids"] = bson.M{"$addToSet": "$userid"}
	m1[2]["$project"].(bson.M)["userids"] = "$userids"
	var summarys []*entity.CrashPlayerStat
	err = PlanePlayerStats.Pipe(m1).All(&summarys)
	if err != nil {
		return
	} else if len(summarys) > 0 {
		summary = summarys[0]
		summary.No = fmt.Sprintf("%s-%s总汇", startDate, endDate)
		summary.Id = mark
		summary.FNickname = mark
		summary.FUserType = mark
		summary.FRegistArea = mark
		summary.FCtime = mark
		summary.FLoginTime = mark
		// summary.FLiveDays = mark
		// summary.FLoseDays = mark
		planeStatPlayersSummaryMapping(summary)
	}
	stats = append([]*entity.CrashPlayerStat{summary}, stats...)
	return
}

func planeStatPlayersMapping(stats []*entity.CrashPlayerStat) {
	var playerIds []string
	// var playerStats = make(map[string]*entity.CrashPlayerStat, len(stats))
	for _, stat := range stats {
		playerIds = append(playerIds, stat.Id)
		// playerStats[stat.Id] = stat
	}
	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": playerIds}}},
		{"$project": bson.M{
			"_id":         "$_id",
			"nickname":    "$nickname",
			"money":       "$money",
			"cash_out":    "$cash_out",
			"diamond":     "$diamond",
			"ctime":       "$ctime",
			"login_time":  "$login_time",
			"regist_area": "$regist_area", //ab测试 0:A 1:B 2:C
			"state":       "$state",
		}},
	}
	var r2 []bson.M
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("crash users error:", err)
		return
	}
	var r2Map = make(map[string]bson.M, len(r2))
	for _, r := range r2 {
		userid := r["_id"].(string)
		r2Map[userid] = r
	}
	now := utils.BsonNow()
	for i, stat := range stats {
		r, ok := r2Map[stat.Id]
		if !ok {
			continue
		}
		// 玩家昵称	玩家类型 账号类型 注册时间 最后登陆时间 存活天数 流失天数
		// userid := r["_id"].(string)
		nickname := r["nickname"].(string)
		money := r["money"].(int)
		cash_out := r["cash_out"].(int)
		diamond := r["diamond"].(int64)
		ctime := r["ctime"].(time.Time)
		login_time := r["login_time"].(time.Time)
		regist_area := r["regist_area"].(int)
		state := r["state"].(int)

		stat.No = strconv.Itoa(i + 1)
		stat.FNickname = nickname
		stat.Money = uint32(money)
		stat.FCtime = ctime.Format(utils.FORMAT)
		stat.FLoginTime = login_time.Format(utils.FORMAT)
		stat.FLiveDays = int(math.Ceil(login_time.Sub(ctime).Hours() / 24))
		stat.FLoseDays = int(math.Floor(now.Sub(login_time).Hours() / 24))

		// summary.FLiveDays += stat.FLiveDays
		// summary.FLoseDays += stat.FLoseDays
		// summary.AllRounds += stat.AllRounds
		// summary.GameTimes += stat.GameTimes
		// summary.Bets += stat.Bets
		// summary.BetRounds += stat.BetRounds
		// summary.WinRounds += stat.WinRounds
		// summary.LoseRounds += stat.LoseRounds
		// summary.TieRounds += stat.TieRounds
		// summary.WinBets += stat.WinBets
		// summary.LoseBets += stat.LoseBets
		// summary.TieBets += stat.TieBets
		// summary.Wins += stat.Wins
		// summary.Loses += stat.Loses
		// summary.Cash += stat.Cash
		// summary.MulpitleSum += stat.MulpitleSum
		// summary.WinEscapeMulpitleSum += stat.WinEscapeMulpitleSum
		// summary.WinMulpitleSum += stat.WinMulpitleSum
		// summary.LoseMulpitleSum += stat.LoseMulpitleSum
		// summary.Mulpitles = append(summary.Mulpitles, stat.Mulpitles...)
		// summary.WinEscapeMulpitles = append(summary.WinEscapeMulpitles, stat.WinEscapeMulpitles...)
		// summary.WinMulpitles = append(summary.WinMulpitles, stat.WinMulpitles...)
		// summary.LoseMulpitles = append(summary.LoseMulpitles, stat.LoseMulpitles...)
		// summary.ObserveRounds += stat.ObserveRounds
		// summary.SMoney += money
		// summary.SCashOut += cash_out
		// summary.SDiamond += diamond

		switch regist_area {
		case 0:
			stat.FRegistArea = "A类"
		case 1:
			stat.FRegistArea = "B类"
		case 2:
			stat.FRegistArea = "C类"
		}

		switch state {
		case 1:
			stat.FUserType = "新手"
		case 2:
			if money == 0 {
				stat.FUserType = "零充"
			} else if money >= 20000 && money <= 99999 {
				stat.FUserType = "普充"
			} else if money >= 100000 && money <= 499999 {
				stat.FUserType = "小R"
			} else if money >= 500000 && money <= 999999 {
				stat.FUserType = "中R"
			} else if money >= 1000000 && money <= 9999999 {
				stat.FUserType = "大R"
			} else if money >= 10000000 {
				stat.FUserType = "超大R"
			}
		case 3:
			stat.FUserType = "平民"
		case 4:
			stat.FUserType = "泡沫"
		}

		stat.FGameTimes = fmt.Sprintf("%d分%d秒", stat.GameTimes/60, stat.GameTimes&60)
		stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
		if stat.BetRounds > 0 {
			stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.BetRounds)/100.0)
			// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
		}
		stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
		stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
		if stat.BetRounds > 0 {
			stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
		}
		if stat.Bets > 0 {
			stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
		}
		if stat.WinRounds > 0 {
			stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
		}
		stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
		stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
		stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)
		if stat.WinRounds > 0 {
			stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FMulpitleAvg = fmt.Sprintf("%.2f", stat.MulpitleSum/float64(stat.BetRounds))
		}
		if len(stat.Mulpitles) > 0 {
			var mulpitles []float64
			for _, m := range stat.Mulpitles {
				mulpitles = append(mulpitles, m...)
			}
			sort.Slice(mulpitles, func(i, j int) bool {
				return mulpitles[i] < mulpitles[j]
			})
			length := len(mulpitles)
			if length > 0 {
				if length%2 == 1 {
					stat.FMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
				} else {
					// 中间两位算平均
					stat.FMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
				}
			}
		}
		// 赢局局均逃脱倍数
		var winEscapeMulpitleAvg float64
		if stat.WinRounds > 0 {
			winEscapeMulpitleAvg = stat.WinEscapeMulpitleSum / float64(stat.WinRounds)
			stat.FWinEscapeMulpitleAvg = fmt.Sprintf("%.2f", winEscapeMulpitleAvg)
		}
		if len(stat.WinEscapeMulpitles) > 0 {
			var mulpitles []float64
			for _, m := range stat.WinEscapeMulpitles {
				mulpitles = append(mulpitles, m...)
			}
			sort.Slice(mulpitles, func(i, j int) bool {
				return mulpitles[i] < mulpitles[j]
			})
			length := len(mulpitles)
			if length > 0 {
				// 赢局最高逃脱倍数
				stat.FWinEscapeMulpitleMax = fmt.Sprintf("%.2f", mulpitles[len(mulpitles)-1])
				if length%2 == 1 {
					stat.FWinEscapeMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
				} else {
					// 中间两位算平均
					stat.FWinEscapeMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
				}
			}
		}
		// 赢局局均实际爆炸倍数
		var winMulpitleAvg float64
		if stat.WinRounds > 0 {
			winMulpitleAvg = stat.WinMulpitleSum / float64(stat.WinRounds)
			stat.FWinMulpitleAvg = fmt.Sprintf("%.2f", winMulpitleAvg)
		}
		if len(stat.WinMulpitles) > 0 {
			var mulpitles []float64
			for _, m := range stat.WinMulpitles {
				mulpitles = append(mulpitles, m...)
			}
			sort.Slice(mulpitles, func(i, j int) bool {
				return mulpitles[i] < mulpitles[j]
			})
			length := len(mulpitles)
			if length > 0 {
				if length%2 == 1 {
					stat.FWinMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
				} else {
					// 中间两位算平均
					stat.FWinMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
				}
			}
		}
		// 倍差空间
		stat.FMulpitleDiff = fmt.Sprintf("%.2f", winMulpitleAvg-winEscapeMulpitleAvg)

		// 输局局均实际爆炸倍数
		var loseMulpitleAvg float64
		if stat.LoseRounds > 0 {
			loseMulpitleAvg = stat.LoseMulpitleSum / float64(stat.LoseRounds)
			stat.FLoseMulpitleAvg = fmt.Sprintf("%.2f", loseMulpitleAvg)
		}
		if len(stat.LoseMulpitles) > 0 {
			var mulpitles []float64
			for _, m := range stat.LoseMulpitles {
				mulpitles = append(mulpitles, m...)
			}
			sort.Slice(mulpitles, func(i, j int) bool {
				return mulpitles[i] < mulpitles[j]
			})
			length := len(mulpitles)
			if length > 0 {
				if length%2 == 1 {
					stat.FLoseMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
				} else {
					// 中间两位算平均
					stat.FLoseMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
				}
			}
		}
		if loseMulpitleAvg > 0 {
			stat.FWinEscapeRate = fmt.Sprintf("%.2f", winEscapeMulpitleAvg/loseMulpitleAvg)
			stat.FWinLoseRate = fmt.Sprintf("%.2f", winMulpitleAvg/loseMulpitleAvg)
		}
		if stat.BetRounds > 0 {
			stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
		}
		if stat.BetRounds > 0 {
			stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
		}
		stat.FMoney = fmt.Sprintf("%.2f", float64(money)/100)
		stat.FCashOut = fmt.Sprintf("%.2f", float64(cash_out)/100)
		stat.FWinMoney = fmt.Sprintf("%.2f", float64(int(diamond)+cash_out-money)/100)
	}
}

func planeStatPlayersSummaryMapping(stat *entity.CrashPlayerStat) {

	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": stat.Userids}}},
		{"$group": bson.M{
			"_id":      nil,
			"money":    bson.M{"$sum": "$money"},
			"cash_out": bson.M{"$sum": "$cash_out"},
			"diamond":  bson.M{"$sum": "$diamond"},
		}},
	}
	var r2 []bson.M
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("crash users error:", err)
		return
	} else if len(r2) > 0 {
		money := r2[0]["money"].(int)
		cash_out := r2[0]["cash_out"].(int)
		diamond := r2[0]["diamond"].(int64)
		stat.SMoney = money
		stat.SCashOut = cash_out
		stat.SDiamond = diamond
	}

	stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
	if stat.BetRounds > 0 {
		stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.BetRounds)/100.0)
		// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
	}

	stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
	stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
	if stat.BetRounds > 0 {
		stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
	}
	if stat.Bets > 0 {
		stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
	}
	if stat.WinRounds > 0 {
		stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
	}
	stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
	stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
	stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)
	if stat.WinRounds > 0 {
		stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FMulpitleAvg = fmt.Sprintf("%.2f", stat.MulpitleSum/float64(stat.BetRounds))
	}
	if len(stat.Mulpitles) > 0 {
		var mulpitles []float64
		for _, m := range stat.Mulpitles {
			mulpitles = append(mulpitles, m...)
		}
		sort.Slice(mulpitles, func(i, j int) bool {
			return mulpitles[i] < mulpitles[j]
		})
		length := len(mulpitles)
		if length > 0 {
			if length%2 == 1 {
				stat.FMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
			} else {
				// 中间两位算平均
				stat.FMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
			}
		}
	}
	// 赢局局均逃脱倍数
	var winEscapeMulpitleAvg float64
	if stat.WinRounds > 0 {
		winEscapeMulpitleAvg = stat.WinEscapeMulpitleSum / float64(stat.WinRounds)
		stat.FWinEscapeMulpitleAvg = fmt.Sprintf("%.2f", winEscapeMulpitleAvg)
	}
	if len(stat.WinEscapeMulpitles) > 0 {
		var mulpitles []float64
		for _, m := range stat.WinEscapeMulpitles {
			mulpitles = append(mulpitles, m...)
		}
		sort.Slice(mulpitles, func(i, j int) bool {
			return mulpitles[i] < mulpitles[j]
		})
		length := len(mulpitles)
		if length > 0 {
			// 赢局最高逃脱倍数
			stat.FWinEscapeMulpitleMax = fmt.Sprintf("%.2f", mulpitles[len(mulpitles)-1])
			if length%2 == 1 {
				stat.FWinEscapeMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
			} else {
				// 中间两位算平均
				stat.FWinEscapeMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
			}
		}
	}
	// 赢局局均实际爆炸倍数
	var winMulpitleAvg float64
	if stat.WinRounds > 0 {
		winMulpitleAvg = stat.WinMulpitleSum / float64(stat.WinRounds)
		stat.FWinMulpitleAvg = fmt.Sprintf("%.2f", winMulpitleAvg)
	}
	if len(stat.WinMulpitles) > 0 {
		var mulpitles []float64
		for _, m := range stat.WinMulpitles {
			mulpitles = append(mulpitles, m...)
		}
		sort.Slice(mulpitles, func(i, j int) bool {
			return mulpitles[i] < mulpitles[j]
		})
		length := len(mulpitles)
		if length > 0 {
			if length%2 == 1 {
				stat.FWinMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
			} else {
				// 中间两位算平均
				stat.FWinMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
			}
		}
	}
	// 倍差空间
	stat.FMulpitleDiff = fmt.Sprintf("%.2f", winMulpitleAvg-winEscapeMulpitleAvg)

	// 输局局均实际爆炸倍数
	var loseMulpitleAvg float64
	if stat.LoseRounds > 0 {
		loseMulpitleAvg = stat.LoseMulpitleSum / float64(stat.LoseRounds)
		stat.FLoseMulpitleAvg = fmt.Sprintf("%.2f", loseMulpitleAvg)
	}
	if len(stat.LoseMulpitles) > 0 {
		var mulpitles []float64
		for _, m := range stat.LoseMulpitles {
			mulpitles = append(mulpitles, m...)
		}
		sort.Slice(mulpitles, func(i, j int) bool {
			return mulpitles[i] < mulpitles[j]
		})
		length := len(mulpitles)
		if length > 0 {
			if length%2 == 1 {
				stat.FLoseMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
			} else {
				// 中间两位算平均
				stat.FLoseMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
			}
		}
	}
	if loseMulpitleAvg > 0 {
		stat.FWinEscapeRate = fmt.Sprintf("%.2f", winEscapeMulpitleAvg/loseMulpitleAvg)
		stat.FWinLoseRate = fmt.Sprintf("%.2f", winMulpitleAvg/loseMulpitleAvg)
	}
	if stat.BetRounds > 0 {
		stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
	}
	if stat.BetRounds > 0 {
		stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
	}
	stat.FMoney = fmt.Sprintf("%.2f", float64(stat.SMoney)/100)
	stat.FCashOut = fmt.Sprintf("%.2f", float64(stat.SCashOut)/100)
	stat.FWinMoney = fmt.Sprintf("%.2f", float64(int(stat.SDiamond)+stat.SCashOut-stat.SMoney)/100)
}

// GetPlaneStatDates plane汇总明细
func (s *gameStatsService) GetPlaneStatDates(m, m2, m3 bson.M, startDate, endDate string) (stats []*entity.CrashPlayerStat, err error) {
	// 用户列表筛选
	var filterDatePlayerIds []string
	if len(m2) > 0 || len(m3) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				// "_id":        bson.M{"userid":"$userid", "date", "$date"},
				"_id":        "$_id",
				"bets":       bson.M{"$sum": "$bets"},
				"bet_rounds": bson.M{"$sum": "$bet_rounds"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = PlanePlayerStats.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetCrashStatPlayers error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			id := r["_id"].(string)
			filterDatePlayerIds = append(filterDatePlayerIds, id)
			ids := strings.Split(id, "-") // id=date-userid
			if len(ids) >= 2 {
				userids = append(userids, ids[1])
			}
		}
		if len(filterDatePlayerIds) == 0 { // 未匹配到
			return
		}
		if len(m3) > 0 { // 流失天数，局均打码量
			if len(userids) == 0 { // 未匹配到
				return
			}
			now := utils.BsonNow()
			m3_2 := []bson.M{
				{"$match": bson.M{"_id": bson.M{"$in": userids}}},
				{"$project": bson.M{
					"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
					"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
				}},
				{"$match": m3},
			}
			var r3_2 []bson.M
			err = PlayerUsers.Pipe(m3_2).All(&r3_2)
			if err != nil {
				beego.Error("GetCrashStatPlayers error32:", err)
			} else if len(r3_2) == 0 { // 未匹配到
				return
			}
			var filterId2 []string
			for _, id := range filterDatePlayerIds {
				for _, r := range r3_2 {
					userid := r["_id"].(string)
					if strings.HasSuffix(id, "-"+userid) {
						filterId2 = append(filterId2, id)
						break
					}
				}
			}
			if len(filterId2) == 0 { // 未匹配到
				return
			}
			filterDatePlayerIds = filterId2
		}
	}
	if len(filterDatePlayerIds) > 0 {
		m["_id"] = bson.M{"$in": filterDatePlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":                     "$date_str",
			"date":                    bson.M{"$min": "$date"},
			"players":                 bson.M{"$sum": 1},
			"new_players":             bson.M{"$sum": bson.M{"$cond": bson.M{"if": "$new_reg", "then": 1, "else": 0}}},
			"old_players":             bson.M{"$sum": bson.M{"$cond": bson.M{"if": "$new_reg", "then": 0, "else": 1}}},
			"all_rounds_list":         bson.M{"$push": "$bet_rounds"},
			"all_rounds":              bson.M{"$sum": "$all_rounds"},
			"game_times":              bson.M{"$sum": "$game_times"},
			"bets":                    bson.M{"$sum": "$bets"},
			"bet_rounds":              bson.M{"$sum": "$bet_rounds"},
			"observe_rounds":          bson.M{"$sum": "$observe_rounds"},
			"win_rounds":              bson.M{"$sum": "$win_rounds"},
			"lose_rounds":             bson.M{"$sum": "$lose_rounds"},
			"tie_rounds":              bson.M{"$sum": "$tie_rounds"},
			"win_bets":                bson.M{"$sum": "$win_bets"},
			"lose_bets":               bson.M{"$sum": "$lose_bets"},
			"tie_bets":                bson.M{"$sum": "$tie_bets"},
			"wins":                    bson.M{"$sum": "$wins"},
			"loses":                   bson.M{"$sum": "$loses"},
			"cash":                    bson.M{"$sum": "$cash"},
			"mulpitle_sum":            bson.M{"$sum": "$mulpitle_sum"},
			"win_escape_mulpitle_sum": bson.M{"$sum": "$win_escape_mulpitle_sum"},
			"win_mulpitle_sum":        bson.M{"$sum": "$win_mulpitle_sum"},
			"lose_mulpitle_sum":       bson.M{"$sum": "$lose_mulpitle_sum"},
			"mulpitles":               bson.M{"$push": "$mulpitles"},
			"win_escape_mulpitles":    bson.M{"$push": "$win_escape_mulpitles"},
			"win_mulpitles":           bson.M{"$push": "$win_mulpitles"},
			"lose_mulpitles":          bson.M{"$push": "$lose_mulpitles"},
			"round_bet_avg":           bson.M{"$sum": "$round_bet_avg"},
		}},
		{"$sort": bson.M{"date": -1}},
	}
	err = PlanePlayerStats.Pipe(m1).All(&stats)
	if err != nil {
		return
	}

	// 查询日活数据
	var dateLoginedUsers = make(map[int64]int)
	m9 := []bson.M{
		{"$match": bson.M{"date": m["date"]}},
		{"$project": bson.M{"date": "$date", "logined_users": "$logined_users"}},
	}
	var r9 []bson.M
	err = PddStats.Pipe(m9).All(&r9)
	if err != nil {
		beego.Error("GetCrashStatDates logined_users error: ", err)
	} else {
		for _, r := range r9 {
			date := r["date"].(int64)
			logined_users := r["logined_users"].(int)
			dateLoginedUsers[date] = logined_users
		}
	}

	mark := "--"
	summary := &entity.CrashPlayerStat{
		DateStr: fmt.Sprintf("%s-%s总汇", startDate, endDate),
		Id:      mark,
	}
	planeStatDatesMapping(summary, stats, dateLoginedUsers)
	planeStatDatesSummaryMapping(summary)
	stats = append([]*entity.CrashPlayerStat{summary}, stats...)
	return
}

func planeStatDatesMapping(summary *entity.CrashPlayerStat, stats []*entity.CrashPlayerStat, dateLoginedUsers map[int64]int) {
	for _, stat := range stats {
		dateStr := stat.Id
		date, _ := utils.Unix(dateStr)
		stat.Date = date
		stat.DateStr = dateStr
		loginedUsers := dateLoginedUsers[stat.Date] // 日活

		stat.DateStr = utils.Stamp2Time(stat.Date).Format("2006-01-02")
		stat.FLoginedUsers = loginedUsers

		summary.FLoginedUsers += stat.FLoginedUsers
		summary.FLiveDays += stat.FLiveDays
		summary.FLoseDays += stat.FLoseDays
		summary.AllRounds += stat.AllRounds
		summary.GameTimes += stat.GameTimes
		summary.Bets += stat.Bets
		summary.BetRounds += stat.BetRounds
		summary.WinRounds += stat.WinRounds
		summary.LoseRounds += stat.LoseRounds
		summary.TieRounds += stat.TieRounds
		summary.WinBets += stat.WinBets
		summary.LoseBets += stat.LoseBets
		summary.TieBets += stat.TieBets
		summary.Wins += stat.Wins
		summary.Loses += stat.Loses
		summary.Cash += stat.Cash
		summary.MulpitleSum += stat.MulpitleSum
		summary.WinEscapeMulpitleSum += stat.WinEscapeMulpitleSum
		summary.WinMulpitleSum += stat.WinMulpitleSum
		summary.LoseMulpitleSum += stat.LoseMulpitleSum
		summary.Mulpitles = append(summary.Mulpitles, stat.Mulpitles...)
		summary.WinEscapeMulpitles = append(summary.WinEscapeMulpitles, stat.WinEscapeMulpitles...)
		summary.WinMulpitles = append(summary.WinMulpitles, stat.WinMulpitles...)
		summary.LoseMulpitles = append(summary.LoseMulpitles, stat.LoseMulpitles...)
		summary.ObserveRounds += stat.ObserveRounds
		summary.Players += stat.Players
		summary.NewPlayers += stat.NewPlayers
		summary.OldPlayers += stat.OldPlayers
		summary.AllRoundsList = append(summary.AllRoundsList, stat.AllRoundsList...)

		// 玩 crash 人数占日活比
		if stat.FLoginedUsers > 0 {
			stat.FPlayerRate = fmt.Sprintf("%.2f", float64(stat.Players)/float64(stat.FLoginedUsers))
		}

		// 查询当天注册的新用户
		startTime := utils.Str2Time(fmt.Sprintf("%s 00:00:00", dateStr), Location())
		endTime := utils.Str2Time(fmt.Sprintf("%s 23:59:59", dateStr), Location())
		m7 := []bson.M{
			{"$match": bson.M{
				"ctime":            bson.M{"$gte": startTime, "$lte": endTime},
				"robot":            false,
				"simulation_robot": false,
			}},
			{"$group": bson.M{
				"_id":   nil,
				"total": bson.M{"$sum": 1},
			}},
		}
		var r7 []bson.M
		err := PlayerUsers.Pipe(m7).All(&r7)
		if err != nil {
			beego.Error("查询当天注册的新用户 error: ", err)
		} else if len(r7) > 0 {
			stat.FLoginedNewUsers = r7[0]["total"].(int)
		}
		stat.FLoginedOldUsers = stat.FLoginedUsers - stat.FLoginedNewUsers
		if stat.FLoginedOldUsers < 0 {
			stat.FLoginedOldUsers = 0
		}
		summary.FLoginedNewUsers += stat.FLoginedNewUsers
		summary.FLoginedOldUsers += stat.FLoginedOldUsers
		// FNewPlayerRate, FOldPlayerRate 占新老玩家日活
		if stat.FLoginedNewUsers > 0 {
			stat.FNewPlayerRate = fmt.Sprintf("%.2f", float64(stat.NewPlayers)/float64(stat.FLoginedNewUsers))
		}
		if stat.FLoginedOldUsers > 0 {
			stat.FOldPlayerRate = fmt.Sprintf("%.2f", float64(stat.OldPlayers)/float64(stat.FLoginedOldUsers))
		}

		if stat.Players > 0 {
			stat.FPlayerRoundsAvg = fmt.Sprintf("%.2f", float64(stat.BetRounds)/float64(stat.Players))
		}
		// 局数中位数
		if len(stat.AllRoundsList) > 0 {
			var rounds = stat.AllRoundsList
			sort.Slice(rounds, func(i, j int) bool {
				return rounds[i] < rounds[j]
			})
			length := len(rounds)
			if length > 0 {
				if length%2 == 1 {
					stat.FPlayerRoundsMedian = rounds[length/2]
				} else {
					// 中间两位算平均
					stat.FPlayerRoundsMedian = (rounds[length/2] + rounds[length/2-1]) / 2
				}
			}
		}
		stat.FGameTimes = fmt.Sprintf("%.2f", float64(stat.GameTimes)/60)
		if stat.BetRounds > 0 {
			stat.FGameTimesAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.BetRounds)/60)
		}
		if stat.Players > 0 {
			stat.FGameTimesPlayerAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.Players)/60)
		}
		stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
		if stat.Players > 0 {
			stat.FPlayerBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.Players)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.BetRounds)/100.0)
			// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
		}
		if stat.Bets > 0 {
			stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
		}
		stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
		stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
		if stat.WinRounds > 0 {
			stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
		}
		stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
		stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
		stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)
		if stat.WinRounds > 0 {
			stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FMulpitleAvg = fmt.Sprintf("%.2f", stat.MulpitleSum/float64(stat.BetRounds))
		}
		if len(stat.Mulpitles) > 0 {
			var mulpitles []float64
			for _, m := range stat.Mulpitles {
				mulpitles = append(mulpitles, m...)
			}
			sort.Slice(mulpitles, func(i, j int) bool {
				return mulpitles[i] < mulpitles[j]
			})
			length := len(mulpitles)
			if length > 0 {
				if length%2 == 1 {
					stat.FMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
				} else {
					// 中间两位算平均
					stat.FMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
				}
			}
		}
		// 赢局局均逃脱倍数
		var winEscapeMulpitleAvg float64
		if stat.WinRounds > 0 {
			winEscapeMulpitleAvg = stat.WinEscapeMulpitleSum / float64(stat.WinRounds)
			stat.FWinEscapeMulpitleAvg = fmt.Sprintf("%.2f", winEscapeMulpitleAvg)
		}
		if len(stat.WinEscapeMulpitles) > 0 {
			var mulpitles []float64
			for _, m := range stat.WinEscapeMulpitles {
				mulpitles = append(mulpitles, m...)
			}
			sort.Slice(mulpitles, func(i, j int) bool {
				return mulpitles[i] < mulpitles[j]
			})
			length := len(mulpitles)
			if length > 0 {
				// 赢局最高逃脱倍数
				stat.FWinEscapeMulpitleMax = fmt.Sprintf("%.2f", mulpitles[len(mulpitles)-1])
				if length%2 == 1 {
					stat.FWinEscapeMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
				} else {
					// 中间两位算平均
					stat.FWinEscapeMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
				}
			}
		}
		// 赢局局均实际爆炸倍数
		var winMulpitleAvg float64
		if stat.WinRounds > 0 {
			winMulpitleAvg = stat.WinMulpitleSum / float64(stat.WinRounds)
			stat.FWinMulpitleAvg = fmt.Sprintf("%.2f", winMulpitleAvg)
		}
		if len(stat.WinMulpitles) > 0 {
			var mulpitles []float64
			for _, m := range stat.WinMulpitles {
				mulpitles = append(mulpitles, m...)
			}
			sort.Slice(mulpitles, func(i, j int) bool {
				return mulpitles[i] < mulpitles[j]
			})
			length := len(mulpitles)
			if length > 0 {
				if length%2 == 1 {
					stat.FWinMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
				} else {
					// 中间两位算平均
					stat.FWinMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
				}
			}
		}
		// 倍差空间
		stat.FMulpitleDiff = fmt.Sprintf("%.2f", winMulpitleAvg-winEscapeMulpitleAvg)

		// 输局局均实际爆炸倍数
		var loseMulpitleAvg float64
		if stat.LoseRounds > 0 {
			loseMulpitleAvg = stat.LoseMulpitleSum / float64(stat.LoseRounds)
			stat.FLoseMulpitleAvg = fmt.Sprintf("%.2f", loseMulpitleAvg)
		}
		if len(stat.LoseMulpitles) > 0 {
			var mulpitles []float64
			for _, m := range stat.LoseMulpitles {
				mulpitles = append(mulpitles, m...)
			}
			sort.Slice(mulpitles, func(i, j int) bool {
				return mulpitles[i] < mulpitles[j]
			})
			length := len(mulpitles)
			if length > 0 {
				if length%2 == 1 {
					stat.FLoseMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
				} else {
					// 中间两位算平均
					stat.FLoseMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
				}
			}
		}
		if loseMulpitleAvg > 0 {
			stat.FWinEscapeRate = fmt.Sprintf("%.2f", winEscapeMulpitleAvg/loseMulpitleAvg)
			stat.FWinLoseRate = fmt.Sprintf("%.2f", winMulpitleAvg/loseMulpitleAvg)
		}
		if stat.BetRounds > 0 {
			stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
		}
		if stat.BetRounds > 0 {
			stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
		}
	}
}

func planeStatDatesSummaryMapping(stat *entity.CrashPlayerStat) {
	// 玩 crash 人数占日活比
	if stat.FLoginedUsers > 0 {
		stat.FPlayerRate = fmt.Sprintf("%.2f", float64(stat.Players)/float64(stat.FLoginedUsers))
	}

	// FNewPlayerRate, FOldPlayerRate 占新老玩家日活
	if stat.FLoginedNewUsers > 0 {
		stat.FNewPlayerRate = fmt.Sprintf("%.2f", float64(stat.NewPlayers)/float64(stat.FLoginedNewUsers))
	}
	if stat.FLoginedOldUsers > 0 {
		stat.FOldPlayerRate = fmt.Sprintf("%.2f", float64(stat.OldPlayers)/float64(stat.FLoginedOldUsers))
	}

	if stat.Players > 0 {
		stat.FPlayerRoundsAvg = fmt.Sprintf("%.2f", float64(stat.BetRounds)/float64(stat.Players))
	}
	// 局数中位数
	if len(stat.AllRoundsList) > 0 {
		var rounds = stat.AllRoundsList
		sort.Slice(rounds, func(i, j int) bool {
			return rounds[i] < rounds[j]
		})
		length := len(rounds)
		if length > 0 {
			if length%2 == 1 {
				stat.FPlayerRoundsMedian = rounds[length/2]
			} else {
				// 中间两位算平均
				stat.FPlayerRoundsMedian = (rounds[length/2] + rounds[length/2-1]) / 2
			}
		}
	}
	stat.FGameTimes = fmt.Sprintf("%.2f", float64(stat.GameTimes)/60)
	if stat.BetRounds > 0 {
		stat.FGameTimesAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.BetRounds)/60)
	}
	if stat.Players > 0 {
		stat.FGameTimesPlayerAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.Players)/60)
	}
	stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
	if stat.Players > 0 {
		stat.FPlayerBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.Players)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.BetRounds)/100.0)
		// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
	}
	if stat.Bets > 0 {
		stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
	}
	stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
	stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
	if stat.WinRounds > 0 {
		stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
	}
	stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
	stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
	stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)
	if stat.WinRounds > 0 {
		stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FMulpitleAvg = fmt.Sprintf("%.2f", stat.MulpitleSum/float64(stat.BetRounds))
	}
	if len(stat.Mulpitles) > 0 {
		var mulpitles []float64
		for _, m := range stat.Mulpitles {
			mulpitles = append(mulpitles, m...)
		}
		sort.Slice(mulpitles, func(i, j int) bool {
			return mulpitles[i] < mulpitles[j]
		})
		length := len(mulpitles)
		if length > 0 {
			if length%2 == 1 {
				stat.FMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
			} else {
				// 中间两位算平均
				stat.FMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
			}
		}
	}
	// 赢局局均逃脱倍数
	var winEscapeMulpitleAvg float64
	if stat.WinRounds > 0 {
		winEscapeMulpitleAvg = stat.WinEscapeMulpitleSum / float64(stat.WinRounds)
		stat.FWinEscapeMulpitleAvg = fmt.Sprintf("%.2f", winEscapeMulpitleAvg)
	}
	if len(stat.WinEscapeMulpitles) > 0 {
		var mulpitles []float64
		for _, m := range stat.WinEscapeMulpitles {
			mulpitles = append(mulpitles, m...)
		}
		sort.Slice(mulpitles, func(i, j int) bool {
			return mulpitles[i] < mulpitles[j]
		})
		length := len(mulpitles)
		if length > 0 {
			// 赢局最高逃脱倍数
			stat.FWinEscapeMulpitleMax = fmt.Sprintf("%.2f", mulpitles[len(mulpitles)-1])
			if length%2 == 1 {
				stat.FWinEscapeMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
			} else {
				// 中间两位算平均
				stat.FWinEscapeMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
			}
		}
	}
	// 赢局局均实际爆炸倍数
	var winMulpitleAvg float64
	if stat.WinRounds > 0 {
		winMulpitleAvg = stat.WinMulpitleSum / float64(stat.WinRounds)
		stat.FWinMulpitleAvg = fmt.Sprintf("%.2f", winMulpitleAvg)
	}
	if len(stat.WinMulpitles) > 0 {
		var mulpitles []float64
		for _, m := range stat.WinMulpitles {
			mulpitles = append(mulpitles, m...)
		}
		sort.Slice(mulpitles, func(i, j int) bool {
			return mulpitles[i] < mulpitles[j]
		})
		length := len(mulpitles)
		if length > 0 {
			if length%2 == 1 {
				stat.FWinMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
			} else {
				// 中间两位算平均
				stat.FWinMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
			}
		}
	}
	// 倍差空间
	stat.FMulpitleDiff = fmt.Sprintf("%.2f", winMulpitleAvg-winEscapeMulpitleAvg)

	// 输局局均实际爆炸倍数
	var loseMulpitleAvg float64
	if stat.LoseRounds > 0 {
		loseMulpitleAvg = stat.LoseMulpitleSum / float64(stat.LoseRounds)
		stat.FLoseMulpitleAvg = fmt.Sprintf("%.2f", loseMulpitleAvg)
	}
	if len(stat.LoseMulpitles) > 0 {
		var mulpitles []float64
		for _, m := range stat.LoseMulpitles {
			mulpitles = append(mulpitles, m...)
		}
		sort.Slice(mulpitles, func(i, j int) bool {
			return mulpitles[i] < mulpitles[j]
		})
		length := len(mulpitles)
		if length > 0 {
			if length%2 == 1 {
				stat.FLoseMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
			} else {
				// 中间两位算平均
				stat.FLoseMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
			}
		}
	}
	if loseMulpitleAvg > 0 {
		stat.FWinEscapeRate = fmt.Sprintf("%.2f", winEscapeMulpitleAvg/loseMulpitleAvg)
		stat.FWinLoseRate = fmt.Sprintf("%.2f", winMulpitleAvg/loseMulpitleAvg)
	}
	if stat.BetRounds > 0 {
		stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
	}
	if stat.BetRounds > 0 {
		stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
	}
}

// Crash扶摇直上
func (s *gameStatsService) GetCrashFYZSStat(col *mgo.Collection, page, pageSize int, m, m2, m3 bson.M, startDate, endDate string) (
	stats []*entity.CrashFYZSStat, count int, err error,
) {
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)

	// 用户列表筛选
	var filterPlayerIds []string
	if len(m3) > 0 || len(m2) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				"_id":        "$userid",
				"bets":       bson.M{"$sum": "$bets"},
				"bet_rounds": bson.M{"$sum": "$bet_rounds"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = col.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetCrashFYZSStat error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			userid := r["_id"].(string)
			userids = append(userids, userid)
		}
		if len(userids) == 0 { // 未匹配到
			return
		}
		now := utils.BsonNow()
		m3_2 := []bson.M{
			{"$match": bson.M{"_id": bson.M{"$in": userids}}},
			{"$project": bson.M{
				"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
				"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
			}},
			{"$match": m3},
		}
		var r3_2 []bson.M
		err = PlayerUsers.Pipe(m3_2).All(&r3_2)
		if err != nil {
			beego.Error("GetCrashFYZSStat error32:", err)
		} else if len(r3_2) == 0 { // 未匹配到
			return
		}
		for _, r := range r3_2 {
			userid := r["_id"].(string)
			filterPlayerIds = append(filterPlayerIds, userid)
		}
	}
	if len(filterPlayerIds) > 0 {
		m["userid"] = bson.M{"$in": filterPlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":  "$userid",
			"date": bson.M{"$max": "$date"},
			"fit_count": bson.M{"$sum": bson.M{"$cond": []interface{}{
				bson.M{"$eq": []interface{}{"$is_fit", true}}, 1, 0,
			}}},
			"trigger_count": bson.M{"$sum": bson.M{"$cond": []interface{}{
				bson.M{"$eq": []interface{}{"$is_trigger", true}}, 1, 0,
			}}},
			"all_rounds":                  bson.M{"$sum": "$all_rounds"},
			"tigger_times":                bson.M{"$sum": "$tigger_times"},
			"all_tigger_times":            bson.M{"$sum": "$all_tigger_times"},
			"all_evo_times":               bson.M{"$sum": "$all_evo_times"},
			"win_rounds":                  bson.M{"$sum": "$win_rounds"},
			"lose_rounds":                 bson.M{"$sum": "$lose_rounds"},
			"maximum_escape_multiple":     bson.M{"$max": "$maximum_escape_multiple"},
			"minimum_escape_multiple":     bson.M{"$min": "$minimum_escape_multiple"},
			"win_mulpitle_sum":            bson.M{"$sum": "$win_mulpitle_sum"},
			"bets":                        bson.M{"$sum": "$bets"},
			"bet_rounds":                  bson.M{"$sum": "$bet_rounds"},
			"maximum_escape_multiple_bet": bson.M{"$max": "$maximum_escape_multiple_bet"},
			"minimum_escape_multiple_bet": bson.M{"$min": "$minimum_escape_multiple_bet"},
			"sum_escape_multiple":         bson.M{"$sum": "$sum_escape_multiple"},
			"lose_mulpitle_sum":           bson.M{"$sum": "$lose_mulpitle_sum"},
			"win_bets":                    bson.M{"$sum": "$win_bets"},
			"lose_bets":                   bson.M{"$sum": "$lose_bets"},
			"win_escape_mulpitles":        bson.M{"$push": "$win_escape_mulpitles"},
			"wins":                        bson.M{"$sum": "$wins"},
			"loses":                       bson.M{"$sum": "$loses"},
			"cash":                        bson.M{"$sum": "$cash"},
		}},
		{"$project": bson.M{
			"_id":                         "$_id",
			"fit_count":                   "$fit_count",
			"trigger_count":               "$trigger_count",
			"all_rounds":                  "$all_rounds",
			"tigger_times":                "$tigger_times",
			"all_tigger_times":            "$all_tigger_times",
			"all_evo_times":               "$all_evo_times",
			"win_rounds":                  "$win_rounds",
			"lose_rounds":                 "$lose_rounds",
			"maximum_escape_multiple":     "$maximum_escape_multiple",
			"minimum_escape_multiple":     "$minimum_escape_multiple",
			"win_mulpitle_sum":            "$win_mulpitle_sum",
			"bets":                        "$bets",
			"bet_rounds":                  "$bet_rounds",
			"maximum_escape_multiple_bet": "$maximum_escape_multiple_bet",
			"minimum_escape_multiple_bet": "$minimum_escape_multiple_bet",
			"sum_escape_multiple":         "$sum_escape_multiple",
			"lose_mulpitle_sum":           "$lose_mulpitle_sum",
			"win_bets":                    "$win_bets",
			"lose_bets":                   "$lose_bets",
			"wins":                        "$wins",
			"loses":                       "$loses",
			"cash":                        "$cash",
			"win_escape_mulpitles":        "$win_escape_mulpitles",
			"win_escape_multiple":         bson.M{"$cond": []any{bson.M{"$eq": []any{"$win_mulpitle_sum", 0}}, 0, bson.M{"$divide": []string{"$win_mulpitle_sum", "$win_rounds"}}}},
			"round_bet_avg":               bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
		}},
		{"$match": m2},
		{"$count": "count"},
	}
	// 查总条数
	var rCount []bson.M
	err = col.Pipe(m1).All(&rCount)
	if err != nil {
		return
	} else if len(rCount) > 0 {
		count = rCount[0]["count"].(int)
	}

	// 查列表数据
	m1 = m1[:len(m1)-1]
	m1 = append(m1, []bson.M{
		{"$sort": bson.M{"date": -1}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}...)
	err = col.Pipe(m1).All(&stats)
	if err != nil {
		return
	}
	CrashFyzsStatMapping(stats)

	// 查总汇数据
	mark := "--"
	summary := &entity.CrashFYZSStat{}
	m1 = m1[:len(m1)-3]
	m1[1]["$group"].(bson.M)["_id"] = nil
	delete(m1[1]["$group"].(bson.M), "date")
	m1[1]["$group"].(bson.M)["userids"] = bson.M{"$addToSet": "$userid"}
	m1[2]["$project"].(bson.M)["userids"] = "$userids"
	var summarys []*entity.CrashFYZSStat
	err = col.Pipe(m1).All(&summarys)
	if err != nil {
		return
	} else if len(summarys) > 0 {
		summary = summarys[0]
		summary.No = fmt.Sprintf("%s-%s总汇", startDate, endDate)
		summary.Id = mark
		summary.FNickname = mark
		summary.FUserType = mark
		summary.FRegistArea = mark
		summary.FCtime = mark
		summary.FLoginTime = mark
		// summary.FLiveDays = mark
		// summary.FLoseDays = mark
		CrashFyzsStatSummaryMapping(summary)
	}
	if len(summarys) > 0 {
		stats = append([]*entity.CrashFYZSStat{summary}, stats...)
	}

	return
}

func CrashFyzsStatMapping(stats []*entity.CrashFYZSStat) {
	var playerIds []string
	// var playerStats = make(map[string]*entity.CrashPlayerStat, len(stats))
	for _, stat := range stats {
		playerIds = append(playerIds, stat.Id)
		// playerStats[stat.Id] = stat
	}
	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": playerIds}}},
		{"$project": bson.M{
			"_id":         "$_id",
			"nickname":    "$nickname",
			"money":       "$money",
			"cash_out":    "$cash_out",
			"diamond":     "$diamond",
			"ctime":       "$ctime",
			"login_time":  "$login_time",
			"regist_area": "$regist_area", //ab测试 0:A 1:B 2:C
			"state":       "$state",
		}},
	}
	var r2 []bson.M
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("crash users error:", err)
		return
	}
	var r2Map = make(map[string]bson.M, len(r2))
	for _, r := range r2 {
		userid := r["_id"].(string)
		r2Map[userid] = r
	}
	now := utils.BsonNow()
	for i, stat := range stats {
		r, ok := r2Map[stat.Id]
		if !ok {
			continue
		}
		// 玩家昵称	玩家类型 账号类型 注册时间 最后登陆时间 存活天数 流失天数
		// userid := r["_id"].(string)
		nickname := r["nickname"].(string)
		money := r["money"].(int)
		// cash_out := r["cash_out"].(int)
		// diamond := r["diamond"].(int64)
		ctime := r["ctime"].(time.Time)
		login_time := r["login_time"].(time.Time)
		regist_area := r["regist_area"].(int)
		state := r["state"].(int)

		// 将日期转成印度时间
		login_time, _ = ConvertToIndiaTime(login_time.Unix())
		ctime, _ = ConvertToIndiaTime(ctime.Unix())

		stat.No = strconv.Itoa(i + 1)
		stat.FNickname = nickname
		//stat.Money = uint32(money)
		stat.FCtime = ctime.Format(utils.FORMAT)
		stat.FLoginTime = login_time.Format(utils.FORMAT)
		stat.FLiveDays = int(math.Ceil(login_time.Sub(ctime).Hours() / 24))
		stat.FLoseDays = int(math.Floor(now.Sub(login_time).Hours() / 24))

		switch regist_area {
		case 0:
			stat.FRegistArea = "A类"
		case 1:
			stat.FRegistArea = "B类"
		case 2:
			stat.FRegistArea = "C类"
		}

		switch state {
		case 1:
			stat.FUserType = "新手"
		case 2:
			if money == 0 {
				stat.FUserType = "零充"
			} else if money >= 20000 && money <= 99999 {
				stat.FUserType = "普充"
			} else if money >= 100000 && money <= 499999 {
				stat.FUserType = "小R"
			} else if money >= 500000 && money <= 999999 {
				stat.FUserType = "中R"
			} else if money >= 1000000 && money <= 9999999 {
				stat.FUserType = "大R"
			} else if money >= 10000000 {
				stat.FUserType = "超大R"
			}
		case 3:
			stat.FUserType = "平民"
		case 4:
			stat.FUserType = "泡沫"
		}

		// stat.FGameTimes = fmt.Sprintf("%d分%d秒", stat.GameTimes/60, stat.GameTimes&60)
		stat.FBets = Chip2Float(stat.Bets)
		if stat.BetRounds > 0 {
			// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.BetRounds)/100.0)
			stat.FRoundBetsAvg = Chip2Float(stat.RoundBetAvg)
		}
		if stat.TriggerNumber > 0 {
			stat.FitTriggerRate = ComputeFloat(int64(stat.TriggerNumber), int64(stat.FitNumber)) * 100.0
		}
		if stat.TiggerTimes > 0 {
			stat.TriggerRate = ComputeFloat(int64(stat.TiggerTimes), int64(stat.AllTiggerTimes)) * 100.0
			stat.PerCapitaRate = ComputeFloat(int64(stat.TiggerTimes), int64(stat.TriggerNumber)) * 100.0
		}
		if stat.AllRounds > 0 {
			stat.AvgEscapeMultiple = stat.WinMulpitleSum / float64(stat.AllRounds)
		}
		if stat.LoseRounds > 0 {
			stat.ExpiredRate = ComputeFloat(int64(stat.LoseRounds), int64(stat.AllEvoTimes)) * 100.0
		}
		if stat.AllEvoTimes > 0 {
			stat.TriggerRoundsAvg = ComputeFloat(int64(stat.AllEvoTimes), int64(stat.AllTiggerTimes))
			stat.ValidRoundsAvg = ComputeFloat(int64(stat.AllEvoTimes), int64(stat.TiggerTimes))
		}
		stat.FMaximumEscapeMultipleBet = Chip2Float(stat.MaximumEscapeMultipleBet)
		stat.FMinimumEscapeMultipleBet = Chip2Float(stat.MinimumEscapeMultipleBet)
		if stat.LoseMulpitleSum > 0 {
			stat.FLoseMultipleAvg = stat.LoseMulpitleSum / float64(stat.LoseRounds)
		}

		stat.FLoseBets = Chip2Float(stat.LoseBets)
		if stat.LoseRounds > 0 {
			stat.FLoseBetsAvg = float64(stat.LoseBets) / float64(stat.LoseRounds) / 100.0
		}

		stat.FWins = Chip2Float(stat.Wins)
		stat.FLoses = Chip2Float(stat.Loses)
		stat.FCash = Chip2Float(stat.Cash)
	}
}

func CrashFyzsStatSummaryMapping(stat *entity.CrashFYZSStat) {
	stat.FBets = Chip2Float(stat.Bets)
	if stat.BetRounds > 0 {
		// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.BetRounds)/100.0)
		stat.FRoundBetsAvg = Chip2Float(stat.RoundBetAvg)
	}
	if stat.TriggerNumber > 0 {
		stat.FitTriggerRate = ComputeFloat(int64(stat.TriggerNumber), int64(stat.FitNumber)) * 100.0
	}
	if stat.TiggerTimes > 0 {
		stat.TriggerRate = ComputeFloat(int64(stat.TiggerTimes), int64(stat.AllTiggerTimes)) * 100.0
		stat.PerCapitaRate = ComputeFloat(int64(stat.TiggerTimes), int64(stat.TriggerNumber)) * 100.0
	}
	if stat.AllRounds > 0 {
		stat.AvgEscapeMultiple = stat.WinMulpitleSum / float64(stat.AllRounds)
	}
	if stat.LoseRounds > 0 {
		stat.ExpiredRate = ComputeFloat(int64(stat.LoseRounds), int64(stat.AllEvoTimes)) * 100.0
	}
	if stat.AllEvoTimes > 0 {
		stat.TriggerRoundsAvg = ComputeFloat(int64(stat.AllEvoTimes), int64(stat.AllTiggerTimes))
		stat.ValidRoundsAvg = ComputeFloat(int64(stat.AllEvoTimes), int64(stat.TiggerTimes))
	}
	stat.FMaximumEscapeMultipleBet = Chip2Float(stat.MaximumEscapeMultipleBet)
	stat.FMinimumEscapeMultipleBet = Chip2Float(stat.MinimumEscapeMultipleBet)

	if stat.LoseMulpitleSum > 0 {
		stat.FLoseMultipleAvg = stat.LoseMulpitleSum / float64(stat.LoseRounds)
	}
	stat.FLoseBets = Chip2Float(stat.LoseBets)
	if stat.LoseRounds > 0 {
		stat.FLoseBetsAvg = float64(stat.LoseBets) / float64(stat.LoseRounds) / 100.0
	}

	stat.FWins = Chip2Float(stat.Wins)
	stat.FLoses = Chip2Float(stat.Loses)
	stat.FCash = Chip2Float(stat.Cash)

}

/*
Crash欲薅无门
*/
func (s *gameStatsService) GetCrashYHWMStat(col *mgo.Collection, page, pageSize int, m, m2, m3 bson.M, startDate, endDate string) (
	stats []*entity.CrashYHWMStat, count int, err error,
) {
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)

	// 用户列表筛选
	var filterPlayerIds []string
	if len(m3) > 0 || len(m2) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				"_id":        "$userid",
				"bets":       bson.M{"$sum": "$bets"},
				"bet_rounds": bson.M{"$sum": "$bet_rounds"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = col.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetCrashYHWMStat error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			userid := r["_id"].(string)
			userids = append(userids, userid)
		}
		if len(userids) == 0 { // 未匹配到
			return
		}
		now := utils.BsonNow()
		m3_2 := []bson.M{
			{"$match": bson.M{"_id": bson.M{"$in": userids}}},
			{"$project": bson.M{
				"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
				"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
			}},
			{"$match": m3},
		}
		var r3_2 []bson.M
		err = PlayerUsers.Pipe(m3_2).All(&r3_2)
		if err != nil {
			beego.Error("GetCrashYHWMStat error32:", err)
		} else if len(r3_2) == 0 { // 未匹配到
			return
		}
		for _, r := range r3_2 {
			userid := r["_id"].(string)
			filterPlayerIds = append(filterPlayerIds, userid)
		}
	}
	if len(filterPlayerIds) > 0 {
		m["userid"] = bson.M{"$in": filterPlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":        "$userid",
			"date":       bson.M{"$max": "$date"},
			"all_number": bson.M{"$sum": 1},
			"fit_count": bson.M{"$sum": bson.M{"$cond": []interface{}{
				bson.M{"$eq": []interface{}{"$is_fit", true}}, 1, 0,
			}}},
			"trigger_count": bson.M{"$sum": bson.M{"$cond": []interface{}{
				bson.M{"$eq": []interface{}{"$is_trigger", true}}, 1, 0,
			}}},
			"tigger_times":      bson.M{"$sum": "$tigger_times"},
			"all_tigger_times":  bson.M{"$sum": "$all_tigger_times"},
			"all_evo_times":     bson.M{"$sum": "$all_evo_times"},
			"burst_number":      bson.M{"$sum": "$burst_number"},
			"burst_amount":      bson.M{"$sum": "$burst_amount"},
			"total_reap_amount": bson.M{"$sum": "$total_reap_amount"},
			"win_rounds":        bson.M{"$sum": "$win_rounds"},
			"lose_rounds":       bson.M{"$sum": "$lose_rounds"},
			"bets":              bson.M{"$sum": "$bets"},
			"bet_rounds":        bson.M{"$sum": "$bet_rounds"},
			"win_bets":          bson.M{"$sum": "$win_bets"},
			"lose_bets":         bson.M{"$sum": "$lose_bets"},
			"wins":              bson.M{"$sum": "$wins"},
			"loses":             bson.M{"$sum": "$loses"},
			"cash":              bson.M{"$sum": "$cash"},
			"win_mulpitle_sum":  bson.M{"$sum": "$win_mulpitle_sum"},
		}},
		{"$project": bson.M{
			"_id":                 "$_id",
			"all_number":          "$all_number",
			"fit_count":           "$fit_count",
			"trigger_count":       "$trigger_count",
			"tigger_times":        "$tigger_times",
			"all_tigger_times":    "$all_tigger_times",
			"all_evo_times":       "$all_evo_times",
			"burst_number":        "$burst_number",
			"burst_amount":        "$burst_amount",
			"total_reap_amount":   "$total_reap_amount",
			"win_rounds":          "$win_rounds",
			"lose_rounds":         "$lose_rounds",
			"bets":                "$bets",
			"bet_rounds":          "$bet_rounds",
			"win_bets":            "$win_bets",
			"lose_bets":           "$lose_bets",
			"wins":                "$wins",
			"loses":               "$loses",
			"cash":                "$cash",
			"win_mulpitle_sum":    "$win_mulpitle_sum",
			"win_escape_multiple": bson.M{"$cond": []any{bson.M{"$eq": []any{"$win_mulpitle_sum", 0}}, 0, bson.M{"$divide": []string{"$win_mulpitle_sum", "$win_rounds"}}}},
			"round_bet_avg":       bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
		}},
		{"$match": m2},
		{"$count": "count"},
	}
	// 查总条数
	var rCount []bson.M
	err = col.Pipe(m1).All(&rCount)
	if err != nil {
		return
	} else if len(rCount) > 0 {
		count = rCount[0]["count"].(int)
	}

	// 查列表数据
	m1 = m1[:len(m1)-1]
	m1 = append(m1, []bson.M{
		{"$sort": bson.M{"date": -1}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}...)
	err = col.Pipe(m1).All(&stats)
	if err != nil {
		return
	}
	CrashYhwmStatMapping(stats)

	// 查总汇数据
	mark := "--"
	summary := &entity.CrashYHWMStat{}
	m1 = m1[:len(m1)-3]
	m1[1]["$group"].(bson.M)["_id"] = nil
	delete(m1[1]["$group"].(bson.M), "date")
	m1[1]["$group"].(bson.M)["userids"] = bson.M{"$addToSet": "$userid"}
	m1[2]["$project"].(bson.M)["userids"] = "$userids"
	var summarys []*entity.CrashYHWMStat
	err = col.Pipe(m1).All(&summarys)
	if err != nil {
		return
	} else if len(summarys) > 0 {
		summary = summarys[0]
		summary.No = fmt.Sprintf("%s-%s总汇", startDate, endDate)
		summary.Id = mark
		summary.FNickname = mark
		summary.FUserType = mark
		summary.FRegistArea = mark
		summary.FCtime = mark
		summary.FLoginTime = mark
		// summary.FLiveDays = mark
		// summary.FLoseDays = mark
		CrashYhwmStatSummaryMapping(summary)
	}
	if len(summarys) > 0 {
		stats = append([]*entity.CrashYHWMStat{summary}, stats...)
	}

	return
}

func CrashYhwmStatMapping(stats []*entity.CrashYHWMStat) {
	var playerIds []string
	// var playerStats = make(map[string]*entity.CrashPlayerStat, len(stats))
	for _, stat := range stats {
		playerIds = append(playerIds, stat.Id)
		// playerStats[stat.Id] = stat
	}
	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": playerIds}}},
		{"$project": bson.M{
			"_id":         "$_id",
			"nickname":    "$nickname",
			"money":       "$money",
			"cash_out":    "$cash_out",
			"diamond":     "$diamond",
			"ctime":       "$ctime",
			"login_time":  "$login_time",
			"regist_area": "$regist_area", //ab测试 0:A 1:B 2:C
			"state":       "$state",
		}},
	}
	var r2 []bson.M
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("crash users error:", err)
		return
	}
	var r2Map = make(map[string]bson.M, len(r2))
	for _, r := range r2 {
		userid := r["_id"].(string)
		r2Map[userid] = r
	}
	now := utils.BsonNow()
	for i, stat := range stats {
		r, ok := r2Map[stat.Id]
		if !ok {
			continue
		}
		// 玩家昵称	玩家类型 账号类型 注册时间 最后登陆时间 存活天数 流失天数
		// userid := r["_id"].(string)
		nickname := r["nickname"].(string)
		money := r["money"].(int)
		// cash_out := r["cash_out"].(int)
		// diamond := r["diamond"].(int64)
		ctime := r["ctime"].(time.Time)
		login_time := r["login_time"].(time.Time)
		regist_area := r["regist_area"].(int)
		state := r["state"].(int)

		// 将日期转成印度时间
		login_time, _ = ConvertToIndiaTime(login_time.Unix())
		ctime, _ = ConvertToIndiaTime(ctime.Unix())

		stat.No = strconv.Itoa(i + 1)
		stat.FNickname = nickname
		//stat.Money = uint32(money)
		stat.FCtime = ctime.Format(utils.FORMAT)
		stat.FLoginTime = login_time.Format(utils.FORMAT)
		stat.FLiveDays = int(math.Ceil(login_time.Sub(ctime).Hours() / 24))
		stat.FLoseDays = int(math.Floor(now.Sub(login_time).Hours() / 24))

		switch regist_area {
		case 0:
			stat.FRegistArea = "A类"
		case 1:
			stat.FRegistArea = "B类"
		case 2:
			stat.FRegistArea = "C类"
		}

		switch state {
		case 1:
			stat.FUserType = "新手"
		case 2:
			if money == 0 {
				stat.FUserType = "零充"
			} else if money >= 20000 && money <= 99999 {
				stat.FUserType = "普充"
			} else if money >= 100000 && money <= 499999 {
				stat.FUserType = "小R"
			} else if money >= 500000 && money <= 999999 {
				stat.FUserType = "中R"
			} else if money >= 1000000 && money <= 9999999 {
				stat.FUserType = "大R"
			} else if money >= 10000000 {
				stat.FUserType = "超大R"
			}
		case 3:
			stat.FUserType = "平民"
		case 4:
			stat.FUserType = "泡沫"
		}

		stat.FBets = Chip2Float(stat.Bets)
		stat.FitRate = ComputeFloat(int64(stat.FitNumber), stat.AllNumber) * 100.0
		stat.FitTriggerRate = ComputeFloat(int64(stat.TriggerNumber), int64(stat.FitNumber)) * 100.0
		stat.TriggerRate = ComputeFloat(int64(stat.TiggerTimes), int64(stat.AllTiggerTimes)) * 100.0
		stat.PerCapitaRate = ComputeFloat(int64(stat.TiggerTimes), int64(stat.TriggerNumber)) * 100.0
		stat.BurstRate = ComputeFloat(int64(stat.BurstNumber), int64(stat.AllEvoTimes)) * 100.0
		stat.ValidRoundsAvg = ComputeFloat(int64(stat.AllEvoTimes), int64(stat.TiggerTimes))
		stat.ValidBurstAvg = ComputeFloat(int64(stat.BurstNumber), int64(stat.TiggerTimes))
		stat.FBurstAmount = ComputeFloat(stat.BurstAmount, 100)
		stat.FTotalReapAmount = ComputeFloat(stat.TotalReapAmount, 100)
		stat.BurstReapRats = ComputeFloat(stat.BurstAmount, stat.TotalReapAmount) * 100.0
		stat.BurstReapAmountAvg = ComputeFloat(stat.BurstAmount, int64(stat.TriggerNumber)) / 100.0
		stat.PerCapitaHarvest = ComputeFloat(stat.TotalReapAmount, int64(stat.TriggerNumber)) / 100.0
		stat.PlayerWinRatio = ComputeFloat(int64(stat.WinRounds), int64(stat.AllEvoTimes)) * 100.0
		stat.FRoundBetsAvg = ComputeFloat(int64(stat.Bets), int64(stat.BetRounds)) / 100.0
		stat.FWinRoundBetsAvg = ComputeFloat(int64(stat.WinBets), int64(stat.WinRounds)) / 100.0
		stat.FLoseRoundBetsAvg = ComputeFloat(int64(stat.LoseBets), int64(stat.LoseRounds)) / 100.0
		stat.FWins = Chip2Float(stat.Wins)
		stat.FLoses = Chip2Float(stat.Loses)
		stat.FCash = Chip2Float(stat.Cash)
	}
}

func CrashYhwmStatSummaryMapping(stat *entity.CrashYHWMStat) {
	stat.FBets = Chip2Float(stat.Bets)
	stat.FitRate = ComputeFloat(int64(stat.FitNumber), stat.AllNumber) * 100.0
	stat.FitTriggerRate = ComputeFloat(int64(stat.TriggerNumber), int64(stat.FitNumber)) * 100.0
	stat.TriggerRate = ComputeFloat(int64(stat.TiggerTimes), int64(stat.AllTiggerTimes)) * 100.0
	stat.PerCapitaRate = ComputeFloat(int64(stat.TiggerTimes), int64(stat.TriggerNumber)) * 100.0
	stat.BurstRate = ComputeFloat(int64(stat.BurstNumber), int64(stat.AllEvoTimes)) * 100.0
	stat.ValidRoundsAvg = ComputeFloat(int64(stat.AllEvoTimes), int64(stat.TiggerTimes))
	stat.ValidBurstAvg = ComputeFloat(int64(stat.BurstNumber), int64(stat.TiggerTimes))
	stat.FBurstAmount = ComputeFloat(stat.BurstAmount, 100)
	stat.FTotalReapAmount = ComputeFloat(stat.TotalReapAmount, 100)
	stat.BurstReapRats = ComputeFloat(stat.BurstAmount, stat.TotalReapAmount) * 100.0
	stat.BurstReapAmountAvg = ComputeFloat(stat.BurstAmount, int64(stat.TriggerNumber)) / 100.0
	stat.PerCapitaHarvest = ComputeFloat(stat.TotalReapAmount, int64(stat.TriggerNumber)) / 100.0
	stat.PlayerWinRatio = ComputeFloat(int64(stat.WinRounds), int64(stat.AllEvoTimes)) * 100.0
	stat.FRoundBetsAvg = ComputeFloat(int64(stat.Bets), int64(stat.BetRounds)) / 100.0
	stat.FWinRoundBetsAvg = ComputeFloat(int64(stat.WinBets), int64(stat.WinRounds)) / 100.0
	stat.FLoseRoundBetsAvg = ComputeFloat(int64(stat.LoseBets), int64(stat.LoseRounds)) / 100.0
	stat.FWins = Chip2Float(stat.Wins)
	stat.FLoses = Chip2Float(stat.Loses)
	stat.FCash = Chip2Float(stat.Cash)
}

/*
Crash起死回生
*/
func (s *gameStatsService) GetCrashQSHSStat(col *mgo.Collection, page, pageSize int, m, m2, m3 bson.M, startDate, endDate string) (
	stats []*entity.CrashQSHSStat, count int, err error,
) {
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)

	// 用户列表筛选
	var filterPlayerIds []string
	if len(m3) > 0 || len(m2) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				"_id":        "$userid",
				"bets":       bson.M{"$sum": "$bets"},
				"bet_rounds": bson.M{"$sum": "$bet_rounds"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = col.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetCrashQSHSStat error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			userid := r["_id"].(string)
			userids = append(userids, userid)
		}
		if len(userids) == 0 { // 未匹配到
			return
		}
		now := utils.BsonNow()
		m3_2 := []bson.M{
			{"$match": bson.M{"_id": bson.M{"$in": userids}}},
			{"$project": bson.M{
				"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
				"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
			}},
			{"$match": m3},
		}
		var r3_2 []bson.M
		err = PlayerUsers.Pipe(m3_2).All(&r3_2)
		if err != nil {
			beego.Error("GetCrashQSHSStat error32:", err)
		} else if len(r3_2) == 0 { // 未匹配到
			return
		}
		for _, r := range r3_2 {
			userid := r["_id"].(string)
			filterPlayerIds = append(filterPlayerIds, userid)
		}
	}
	if len(filterPlayerIds) > 0 {
		m["userid"] = bson.M{"$in": filterPlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":  "$userid",
			"date": bson.M{"$max": "$date"},
			// "all_number": bson.M{"$sum": 1},
			// "fit_count": bson.M{"$sum": bson.M{"$cond": []interface{}{
			// 	bson.M{"$eq": []interface{}{"$is_fit", true}}, 1, 0,
			// }}},
			// "trigger_count": bson.M{"$sum": bson.M{"$cond": []interface{}{
			// 	bson.M{"$eq": []interface{}{"$is_trigger", true}}, 1, 0,
			// }}},
			"allin_number":             bson.M{"$sum": "$allin_number"},
			"allin_bets":               bson.M{"$sum": "$allin_bets"},
			"all_evo_times":            bson.M{"$sum": "$all_evo_times"},
			"tigger_bets":              bson.M{"$sum": "$tigger_bets"},
			"win_rounds":               bson.M{"$sum": "$win_rounds"},
			"lose_rounds":              bson.M{"$sum": "$lose_rounds"},
			"tigger_win_mulpitle_sum":  bson.M{"$sum": "$tigger_win_mulpitle_sum"},
			"maximum_escape_multiple":  bson.M{"$max": "$maximum_escape_multiple"},
			"lose_max_escape_multiple": bson.M{"$max": "$lose_max_escape_multiple"},
			"lose_escape_multiple":     bson.M{"$sum": "$lose_escape_multiple"},
			"bets":                     bson.M{"$sum": "$bets"},
			// "bet_rounds":               bson.M{"$sum": "$bet_rounds"},
			"all_rounds":      bson.M{"$sum": "$all_rounds"},
			"all_win_bets":    bson.M{"$sum": "$all_win_bets"},
			"all_lose_bets":   bson.M{"$sum": "$all_lose_bets"},
			"win_bets":        bson.M{"$sum": "$win_bets"},
			"lose_bets":       bson.M{"$sum": "$lose_bets"},
			"wins":            bson.M{"$sum": "$wins"},
			"loses":           bson.M{"$sum": "$loses"},
			"cash":            bson.M{"$sum": "$cash"},
			"withdraw_amount": bson.M{"$last": "$withdraw_amount"},
			"pay_amount":      bson.M{"$last": "$pay_amount"},
			"carry_amount":    bson.M{"$last": "$carry_amount"},
		}},
		{"$sort": bson.M{"date": 1}},
		{"$project": bson.M{
			"_id":                      "$_id",
			"allin_number":             "$allin_number",
			"allin_bets":               "$allin_bets",
			"all_evo_times":            "$all_evo_times",
			"tigger_bets":              "$tigger_bets",
			"win_rounds":               "$win_rounds",
			"lose_rounds":              "$lose_rounds",
			"maximum_escape_multiple":  "$maximum_escape_multiple",
			"tigger_win_mulpitle_sum":  "$tigger_win_mulpitle_sum",
			"lose_max_escape_multiple": "$lose_max_escape_multiple",
			"lose_escape_multiple":     "$lose_escape_multiple",
			"bets":                     "$bets",
			"all_rounds":               "$all_rounds",
			"all_win_bets":             "$all_win_bets",
			"all_lose_bets":            "$all_lose_bets",
			"win_bets":                 "$win_bets",
			"lose_bets":                "$lose_bets",
			"wins":                     "$wins",
			"loses":                    "$loses",
			"cash":                     "$cash",
			"withdraw_amount":          "$withdraw_amount",
			"pay_amount":               "$pay_amount",
			"carry_amount":             "$carry_amount",
			// "win_escape_multiple": bson.M{"$cond": []any{bson.M{"$eq": []any{"$win_mulpitle_sum", 0}}, 0, bson.M{"$divide": []string{"$win_mulpitle_sum", "$win_rounds"}}}},
			"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$all_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$all_rounds"}}}},
		}},
		{"$match": m2},
		{"$count": "count"},
	}
	// 查总条数
	var rCount []bson.M
	err = col.Pipe(m1).All(&rCount)
	if err != nil {
		return
	} else if len(rCount) > 0 {
		count = rCount[0]["count"].(int)
	}

	// 查列表数据
	m1 = m1[:len(m1)-1]
	m1 = append(m1, []bson.M{
		{"$sort": bson.M{"date": -1}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}...)
	err = col.Pipe(m1).All(&stats)
	if err != nil {
		return
	}
	new_col := CrashPlayerStats
	if strings.Contains(col.Name, "_plane_") {
		new_col = PlanePlayerStats
	}
	CrashQshsStatMapping(stats, new_col, m)

	// 查总汇数据
	mark := "--"
	summary := &entity.CrashQSHSStat{}
	m1 = m1[:len(m1)-3]
	m1[1]["$group"].(bson.M)["_id"] = nil
	delete(m1[1]["$group"].(bson.M), "date")
	m1[1]["$group"].(bson.M)["userids"] = bson.M{"$addToSet": "$userid"}
	m1[3]["$project"].(bson.M)["userids"] = "$userids"
	var summarys []*entity.CrashQSHSStat
	err = col.Pipe(m1).All(&summarys)
	if err != nil {
		return
	} else if len(summarys) > 0 {
		summary = summarys[0]
		summary.No = fmt.Sprintf("%s-%s总汇", startDate, endDate)
		summary.Id = mark
		summary.FNickname = mark
		summary.FUserType = mark
		summary.FRegistArea = mark
		summary.FCtime = mark
		summary.FLoginTime = mark
		// summary.FLiveDays = mark
		// summary.FLoseDays = mark
		CrashQshsStatSummaryMapping(summary, new_col, m)
	}
	if len(summarys) > 0 {
		stats = append([]*entity.CrashQSHSStat{summary}, stats...)
	}

	return
}

func CrashQshsStatMapping(stats []*entity.CrashQSHSStat, col *mgo.Collection, m bson.M) {
	var playerIds []string
	// var playerStats = make(map[string]*entity.CrashPlayerStat, len(stats))
	for _, stat := range stats {
		playerIds = append(playerIds, stat.Id)
		// playerStats[stat.Id] = stat
	}
	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": playerIds}}},
		{"$project": bson.M{
			"_id":         "$_id",
			"nickname":    "$nickname",
			"money":       "$money",
			"cash_out":    "$cash_out",
			"diamond":     "$diamond",
			"ctime":       "$ctime",
			"login_time":  "$login_time",
			"regist_area": "$regist_area", //ab测试 0:A 1:B 2:C
			"state":       "$state",
		}},
	}
	var r2 []bson.M
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("crash users error:", err)
		return
	}
	var r2Map = make(map[string]bson.M, len(r2))
	for _, r := range r2 {
		userid := r["_id"].(string)
		r2Map[userid] = r
	}
	// 查询总数据
	var crashlist []entity.CrashPlayerStat
	n := bson.M{}
	// 手动复制 m 到 n
	for k, v := range m {
		n[k] = v
	}
	n["userid"] = bson.M{"$in": playerIds}
	m3 := []bson.M{
		{"$match": n},
		{"$group": bson.M{
			"_id":        "$userid",
			"bet_rounds": bson.M{"$sum": "$bet_rounds"},
			"bets":       bson.M{"$sum": "$bets"},
			"win_bets":   bson.M{"$sum": "$win_bets"},
			"lose_bets":  bson.M{"$sum": "$lose_bets"},
			"wins":       bson.M{"$sum": "$wins"},
			"loses":      bson.M{"$sum": "$loses"},
		}},
	}
	col.Pipe(m3).All(&crashlist)

	now := utils.BsonNow()
	for i, stat := range stats {
		r, ok := r2Map[stat.Id]
		if !ok {
			continue
		}
		// 玩家昵称	玩家类型 账号类型 注册时间 最后登陆时间 存活天数 流失天数
		// userid := r["_id"].(string)
		nickname := r["nickname"].(string)
		money := r["money"].(int)
		// cash_out := r["cash_out"].(int)
		// diamond := r["diamond"].(int64)
		ctime := r["ctime"].(time.Time)
		login_time := r["login_time"].(time.Time)
		regist_area := r["regist_area"].(int)
		state := r["state"].(int)

		// 将日期转成印度时间
		login_time, _ = ConvertToIndiaTime(login_time.Unix())
		ctime, _ = ConvertToIndiaTime(ctime.Unix())

		stat.No = strconv.Itoa(i + 1)
		stat.FNickname = nickname
		//stat.Money = uint32(money)
		stat.FCtime = ctime.Format(utils.FORMAT)
		stat.FLoginTime = login_time.Format(utils.FORMAT)
		stat.FLiveDays = int(math.Ceil(login_time.Sub(ctime).Hours() / 24))
		stat.FLoseDays = int(math.Floor(now.Sub(login_time).Hours() / 24))

		switch regist_area {
		case 0:
			stat.FRegistArea = "A类"
		case 1:
			stat.FRegistArea = "B类"
		case 2:
			stat.FRegistArea = "C类"
		}

		switch state {
		case 1:
			stat.FUserType = "新手"
		case 2:
			if money == 0 {
				stat.FUserType = "零充"
			} else if money >= 20000 && money <= 99999 {
				stat.FUserType = "普充"
			} else if money >= 100000 && money <= 499999 {
				stat.FUserType = "小R"
			} else if money >= 500000 && money <= 999999 {
				stat.FUserType = "中R"
			} else if money >= 1000000 && money <= 9999999 {
				stat.FUserType = "大R"
			} else if money >= 10000000 {
				stat.FUserType = "超大R"
			}
		case 3:
			stat.FUserType = "平民"
		case 4:
			stat.FUserType = "泡沫"
		}

		if stat.AllinNumber != 0 {
			stat.AllinBetAvg = (ComputeFloat(int64(stat.AllinBets), 100) / float64(stat.AllinNumber))
		}
		if stat.AllEvoTimes != 0 {
			stat.TiggerBetAvg = (ComputeFloat(int64(stat.TiggerBets), 100) / float64(stat.AllEvoTimes))
		}
		stat.TiggerRate = ComputeFloat(int64(stat.AllEvoTimes), int64(stat.AllinNumber)) * 100.0
		stat.ExpiredRate = ComputeFloat(int64(stat.LoseRounds), int64(stat.AllEvoTimes)) * 100.0
		if stat.WinRounds != 0 {
			stat.TiggerMulpitle = (stat.TiggerWinMulpitleSum / float64(stat.WinRounds))
		}
		if stat.LoseRounds != 0 {
			stat.LoseEscapeMultipleAvg = (stat.LoseEscapeMultiple / float64(stat.LoseRounds))
		}

		stat.FWinBets = ComputeFloat(int64(stat.WinBets), 100)
		stat.FLoseBets = ComputeFloat(int64(stat.LoseBets), 100)
		stat.FWins = ComputeFloat(int64(stat.Wins), 100)
		stat.FLoses = ComputeFloat(int64(stat.Loses), 100)
		stat.FCash = ComputeFloat(int64(stat.Cash), 100)
		if len(crashlist) > 0 {
			for _, c := range crashlist {
				if c.Id == stat.Id {
					stat.Bets = c.Bets
					stat.AllRounds = c.BetRounds
					stat.AllWinBets = c.Wins
					stat.AllLoseBets = c.Loses
					break
				}
			}
		}
		stat.FBets = ComputeFloat(int64(stat.Bets), 100)
		if stat.AllRounds != 0 {
			stat.FRoundBetAvg = ComputeFloat(int64(stat.Bets), 100) / float64(stat.AllRounds)
		}
		if stat.PayAmount != 0 {
			// stat.TotalRewardsRate = (ComputeFloat(stat.WithdrawAmount-stat.CarryAmount, stat.PayAmount) / 100.0) * 100.0
			stat.TotalRewardsRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.WithdrawAmount+stat.CarryAmount, stat.PayAmount))*100.0)
		}
		if -stat.AllLoseBets != 0 {
			stat.UserRewardRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.AllWinBets, 100)/ComputeFloat(-stat.AllLoseBets, 100)*100.0)
		}
	}
}

func CrashQshsStatSummaryMapping(stat *entity.CrashQSHSStat, col *mgo.Collection, m bson.M) {
	// 查询总数据
	carshinfo := new(entity.CrashPlayerStat)
	n := bson.M{}
	// 手动复制 m 到 n
	for k, v := range m {
		n[k] = v
	}
	n["userid"] = bson.M{"$in": stat.Userids}
	m3 := []bson.M{
		{"$match": n},
		{"$group": bson.M{
			"_id":        nil,
			"bet_rounds": bson.M{"$sum": "$bet_rounds"},
			"bets":       bson.M{"$sum": "$bets"},
			"win_bets":   bson.M{"$sum": "$win_bets"},
			"lose_bets":  bson.M{"$sum": "$lose_bets"},
			"wins":       bson.M{"$sum": "$wins"},
			"loses":      bson.M{"$sum": "$loses"},
		}},
	}
	col.Pipe(m3).One(&carshinfo)
	// stat.FBets = Chip2Float(stat.Bets)
	if stat.AllinNumber != 0 {
		stat.AllinBetAvg = (ComputeFloat(int64(stat.AllinBets), 100) / float64(stat.AllinNumber))
	}
	if stat.AllEvoTimes != 0 {
		stat.TiggerBetAvg = (ComputeFloat(int64(stat.TiggerBets), 100) / float64(stat.AllEvoTimes))
	}
	stat.TiggerRate = ComputeFloat(int64(stat.AllEvoTimes), int64(stat.AllinNumber)) * 100.0
	stat.ExpiredRate = ComputeFloat(int64(stat.LoseRounds), int64(stat.AllEvoTimes)) * 100.0
	if stat.WinRounds != 0 {
		stat.TiggerMulpitle = (stat.TiggerWinMulpitleSum / float64(stat.WinRounds))
	}
	if stat.LoseRounds != 0 {
		stat.LoseEscapeMultipleAvg = (stat.LoseEscapeMultiple / float64(stat.LoseRounds)) * 100.0
	}

	stat.FWinBets = ComputeFloat(int64(stat.WinBets), 100)
	stat.FLoseBets = ComputeFloat(int64(stat.LoseBets), 100)
	stat.FWins = ComputeFloat(int64(stat.Wins), 100)
	stat.FLoses = ComputeFloat(int64(stat.Loses), 100)
	stat.FCash = ComputeFloat(int64(stat.Cash), 100)
	// 查询总局数
	stat.Bets = carshinfo.Bets
	stat.AllRounds = carshinfo.BetRounds
	stat.AllWinBets = carshinfo.Wins
	stat.AllLoseBets = carshinfo.Loses

	stat.FBets = ComputeFloat(int64(stat.Bets), 100)
	if stat.AllRounds != 0 {
		stat.FRoundBetAvg = ComputeFloat(int64(stat.Bets), 100) / float64(stat.AllRounds)
	}
	if stat.PayAmount != 0 {
		// stat.TotalRewardsRate = (ComputeFloat(stat.WithdrawAmount-stat.CarryAmount, stat.PayAmount) / 100.0) * 100.0
		stat.TotalRewardsRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.WithdrawAmount+stat.CarryAmount, stat.PayAmount))*100.0)
	}
	if -stat.AllLoseBets != 0 {
		stat.UserRewardRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.AllWinBets, 100)/ComputeFloat(-stat.AllLoseBets, 100)*100.0)
	}
}

/*
Crash奖池风控
*/
func (s *gameStatsService) GetCrashJCFKStat(col *mgo.Collection, page, pageSize int, m, m2, m3 bson.M, startDate, endDate string) (
	stats []*entity.CrashJCFKStat, count int, err error,
) {
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)

	// 用户列表筛选
	var filterPlayerIds []string
	if len(m3) > 0 || len(m2) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				"_id":        "$userid",
				"bets":       bson.M{"$sum": "$bets"},
				"bet_rounds": bson.M{"$sum": "$bet_rounds"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = col.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetCrashJCFKStat error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			userid := r["_id"].(string)
			userids = append(userids, userid)
		}
		if len(userids) == 0 { // 未匹配到
			return
		}
		now := utils.BsonNow()
		m3_2 := []bson.M{
			{"$match": bson.M{"_id": bson.M{"$in": userids}}},
			{"$project": bson.M{
				"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
				"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
			}},
			{"$match": m3},
		}
		var r3_2 []bson.M
		err = PlayerUsers.Pipe(m3_2).All(&r3_2)
		if err != nil {
			beego.Error("GetCrashJCFKStat error32:", err)
		} else if len(r3_2) == 0 { // 未匹配到
			return
		}
		for _, r := range r3_2 {
			userid := r["_id"].(string)
			filterPlayerIds = append(filterPlayerIds, userid)
		}
	}
	if len(filterPlayerIds) > 0 {
		m["userid"] = bson.M{"$in": filterPlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":  "$userid",
			"date": bson.M{"$max": "$date"},
			// "all_number": bson.M{"$sum": 1},
			// "fit_count": bson.M{"$sum": bson.M{"$cond": []interface{}{
			// 	bson.M{"$eq": []interface{}{"$is_fit", true}}, 1, 0,
			// }}},
			// "trigger_count": bson.M{"$sum": bson.M{"$cond": []interface{}{
			// 	bson.M{"$eq": []interface{}{"$is_trigger", true}}, 1, 0,
			// }}},
			"trigger_times":           bson.M{"$sum": "$trigger_times"},
			"trigger_amount":          bson.M{"$sum": "$trigger_amount"},
			"all_evo_times":           bson.M{"$sum": "$all_evo_times"},
			"tigger_bets":             bson.M{"$sum": "$tigger_bets"},
			"maximum_escape_multiple": bson.M{"$max": "$maximum_escape_multiple"},
			"tigger_mulpitle_sum":     bson.M{"$sum": "$tigger_mulpitle_sum"},
			"win_rounds":              bson.M{"$sum": "$win_rounds"},
			"win_mulpitle_sum":        bson.M{"$sum": "$win_mulpitle_sum"},
			"lose_rounds":             bson.M{"$sum": "$lose_rounds"},
			"lose_mulpitle_sum":       bson.M{"$sum": "$lose_mulpitle_sum"},
			"trigger_rounds":          bson.M{"$sum": "$trigger_rounds"},
			"win_bets":                bson.M{"$sum": "$win_bets"},
			"lose_bets":               bson.M{"$sum": "$lose_bets"},
			"wins":                    bson.M{"$sum": "$wins"},
			"loses":                   bson.M{"$sum": "$loses"},
			"cash":                    bson.M{"$sum": "$cash"},
			"bets":                    bson.M{"$sum": "$bets"},
			"all_rounds":              bson.M{"$sum": "$all_rounds"},
			"all_win_bets":            bson.M{"$sum": "$all_win_bets"},
			"all_lose_bets":           bson.M{"$sum": "$all_lose_bets"},
			"withdraw_amount":         bson.M{"$last": "$withdraw_amount"},
			"pay_amount":              bson.M{"$last": "$pay_amount"},
			"carry_amount":            bson.M{"$last": "$carry_amount"},
		}},
		{"$sort": bson.M{"date": 1}},
		{"$project": bson.M{
			"_id":                     "$_id",
			"trigger_times":           "$trigger_times",
			"trigger_amount":          "$trigger_amount",
			"all_evo_times":           "$all_evo_times",
			"tigger_bets":             "$tigger_bets",
			"maximum_escape_multiple": "$maximum_escape_multiple",
			"tigger_mulpitle_sum":     "$tigger_mulpitle_sum",
			"win_rounds":              "$win_rounds",
			"win_mulpitle_sum":        "$win_mulpitle_sum",
			"lose_rounds":             "$lose_rounds",
			"lose_mulpitle_sum":       "$lose_mulpitle_sum",
			"trigger_rounds":          "$trigger_rounds",
			"win_bets":                "$win_bets",
			"lose_bets":               "$lose_bets",
			"wins":                    "$wins",
			"loses":                   "$loses",
			"cash":                    "$cash",
			"bets":                    "$bets",
			"all_rounds":              "$all_rounds",
			"all_win_bets":            "$all_win_bets",
			"all_lose_bets":           "$all_lose_bets",
			"withdraw_amount":         "$withdraw_amount",
			"pay_amount":              "$pay_amount",
			"carry_amount":            "$carry_amount",
			// "win_escape_multiple": bson.M{"$cond": []any{bson.M{"$eq": []any{"$win_mulpitle_sum", 0}}, 0, bson.M{"$divide": []string{"$win_mulpitle_sum", "$win_rounds"}}}},
			"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$all_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$all_rounds"}}}},
		}},
		{"$match": m2},
		{"$count": "count"},
	}
	// 查总条数
	var rCount []bson.M
	err = col.Pipe(m1).All(&rCount)
	if err != nil {
		return
	} else if len(rCount) > 0 {
		count = rCount[0]["count"].(int)
	}

	// 查列表数据
	m1 = m1[:len(m1)-1]
	m1 = append(m1, []bson.M{
		{"$sort": bson.M{"date": -1}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}...)
	err = col.Pipe(m1).All(&stats)
	if err != nil {
		return
	}
	new_col := CrashPlayerStats
	if strings.Contains(col.Name, "_plane_") {
		new_col = PlanePlayerStats
	}
	CrashJcfkStatMapping(stats, new_col, m)

	// 查总汇数据
	mark := "--"
	summary := &entity.CrashJCFKStat{}
	m1 = m1[:len(m1)-3]
	m1[1]["$group"].(bson.M)["_id"] = nil
	delete(m1[1]["$group"].(bson.M), "date")
	m1[1]["$group"].(bson.M)["userids"] = bson.M{"$addToSet": "$userid"}
	m1[3]["$project"].(bson.M)["userids"] = "$userids"
	var summarys []*entity.CrashJCFKStat
	err = col.Pipe(m1).All(&summarys)
	if err != nil {
		return
	} else if len(summarys) > 0 {
		summary = summarys[0]
		summary.No = fmt.Sprintf("%s-%s总汇", startDate, endDate)
		summary.Id = mark
		summary.FNickname = mark
		summary.FUserType = mark
		summary.FRegistArea = mark
		summary.FCtime = mark
		summary.FLoginTime = mark
		// summary.FLiveDays = mark
		// summary.FLoseDays = mark
		CrashJcfkStatSummaryMapping(summary, new_col, m)
	}
	if len(summarys) > 0 {
		stats = append([]*entity.CrashJCFKStat{summary}, stats...)
	}

	return
}

func CrashJcfkStatMapping(stats []*entity.CrashJCFKStat, col *mgo.Collection, m bson.M) {
	var playerIds []string
	// var playerStats = make(map[string]*entity.CrashPlayerStat, len(stats))
	for _, stat := range stats {
		playerIds = append(playerIds, stat.Id)
		// playerStats[stat.Id] = stat
	}
	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": playerIds}}},
		{"$project": bson.M{
			"_id":         "$_id",
			"nickname":    "$nickname",
			"money":       "$money",
			"cash_out":    "$cash_out",
			"diamond":     "$diamond",
			"ctime":       "$ctime",
			"login_time":  "$login_time",
			"regist_area": "$regist_area", //ab测试 0:A 1:B 2:C
			"state":       "$state",
		}},
	}
	var r2 []bson.M
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("crash users error:", err)
		return
	}
	var r2Map = make(map[string]bson.M, len(r2))
	for _, r := range r2 {
		userid := r["_id"].(string)
		r2Map[userid] = r
	}
	// 查询总数据
	var crashlist []entity.CrashPlayerStat
	n := bson.M{}
	// 手动复制 m 到 n
	for k, v := range m {
		n[k] = v
	}
	n["userid"] = bson.M{"$in": playerIds}
	m3 := []bson.M{
		{"$match": n},
		{"$group": bson.M{
			"_id":        "$userid",
			"bet_rounds": bson.M{"$sum": "$bet_rounds"},
			"bets":       bson.M{"$sum": "$bets"},
			"win_bets":   bson.M{"$sum": "$win_bets"},
			"lose_bets":  bson.M{"$sum": "$lose_bets"},
			"wins":       bson.M{"$sum": "$wins"},
			"loses":      bson.M{"$sum": "$loses"},
		}},
	}
	col.Pipe(m3).All(&crashlist)
	now := utils.BsonNow()
	for i, stat := range stats {
		r, ok := r2Map[stat.Id]
		if !ok {
			continue
		}
		// 玩家昵称	玩家类型 账号类型 注册时间 最后登陆时间 存活天数 流失天数
		// userid := r["_id"].(string)
		nickname := r["nickname"].(string)
		money := r["money"].(int)
		// cash_out := r["cash_out"].(int)
		// diamond := r["diamond"].(int64)
		ctime := r["ctime"].(time.Time)
		login_time := r["login_time"].(time.Time)
		regist_area := r["regist_area"].(int)
		state := r["state"].(int)

		// 将日期转成印度时间
		login_time, _ = ConvertToIndiaTime(login_time.Unix())
		ctime, _ = ConvertToIndiaTime(ctime.Unix())

		stat.No = strconv.Itoa(i + 1)
		stat.FNickname = nickname
		//stat.Money = uint32(money)
		stat.FCtime = ctime.Format(utils.FORMAT)
		stat.FLoginTime = login_time.Format(utils.FORMAT)
		stat.FLiveDays = int(math.Ceil(login_time.Sub(ctime).Hours() / 24))
		stat.FLoseDays = int(math.Floor(now.Sub(login_time).Hours() / 24))

		switch regist_area {
		case 0:
			stat.FRegistArea = "A类"
		case 1:
			stat.FRegistArea = "B类"
		case 2:
			stat.FRegistArea = "C类"
		}

		switch state {
		case 1:
			stat.FUserType = "新手"
		case 2:
			if money == 0 {
				stat.FUserType = "零充"
			} else if money >= 20000 && money <= 99999 {
				stat.FUserType = "普充"
			} else if money >= 100000 && money <= 499999 {
				stat.FUserType = "小R"
			} else if money >= 500000 && money <= 999999 {
				stat.FUserType = "中R"
			} else if money >= 1000000 && money <= 9999999 {
				stat.FUserType = "大R"
			} else if money >= 10000000 {
				stat.FUserType = "超大R"
			}
		case 3:
			stat.FUserType = "平民"
		case 4:
			stat.FUserType = "泡沫"
		}

		stat.FTriggerAmount = ComputeFloat(stat.TriggerAmount, 100)
		if stat.AllEvoTimes != 0 {
			stat.TiggerMulpitleSumAvg = stat.TiggerMulpitleSum / float64(stat.AllEvoTimes)
			stat.TiggerBetAvg = (ComputeFloat(int64(stat.TiggerBets), 100) / float64(stat.AllEvoTimes))
			stat.ExpiredRate = ComputeFloat(int64(stat.TriggerRounds), int64(stat.AllEvoTimes)) * 100.0
		}
		stat.FBets = ComputeFloat(stat.Bets, 100)
		if stat.WinRounds != 0 {
			stat.WinMulpitleSumAvg = (ComputeFloat(int64(stat.WinMulpitleSum), 100) / float64(stat.WinRounds)) * 100.0
		}
		if stat.LoseRounds != 0 {
			stat.LoseMulpitleSumAvg = (ComputeFloat(int64(stat.LoseMulpitleSum), 100) / float64(stat.LoseRounds)) * 100.0
		}

		stat.FWinBets = ComputeFloat(int64(stat.WinBets), 100)
		stat.FLoseBets = ComputeFloat(int64(stat.LoseBets), 100)
		stat.FWins = ComputeFloat(int64(stat.Wins), 100)
		stat.FLoses = ComputeFloat(int64(stat.Loses), 100)
		stat.FCash = ComputeFloat(int64(stat.Cash), 100)
		if len(crashlist) > 0 {
			for _, c := range crashlist {
				if c.Id == stat.Id {
					stat.Bets = c.Bets
					stat.AllRounds = c.BetRounds
					stat.AllWinBets = c.Wins
					stat.AllLoseBets = c.Loses
					break
				}
			}
		}
		stat.FBets = ComputeFloat(int64(stat.Bets), 100)
		if stat.AllRounds != 0 {
			stat.FRoundBetAvg = ComputeFloat(int64(stat.Bets), 100) / float64(stat.AllRounds)
		}
		if stat.PayAmount != 0 {
			// stat.TotalRewardsRate = (ComputeFloat(stat.WithdrawAmount-stat.CarryAmount, stat.PayAmount) / 100.0) * 100.0
			stat.TotalRewardsRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.WithdrawAmount+stat.CarryAmount, stat.PayAmount))*100.0)
		}
		if -stat.AllLoseBets != 0 {
			stat.UserRewardRate = fmt.Sprintf("%.2f%%", ComputeFloat(int64(stat.AllWinBets), 100)/ComputeFloat(int64(int64(-stat.AllLoseBets)), 100)*100.0)
		}
	}
}

func CrashJcfkStatSummaryMapping(stat *entity.CrashJCFKStat, col *mgo.Collection, m bson.M) {
	// 查询总数据
	carshinfo := new(entity.CrashPlayerStat)
	n := bson.M{}
	// 手动复制 m 到 n
	for k, v := range m {
		n[k] = v
	}
	n["userid"] = bson.M{"$in": stat.Userids}
	m3 := []bson.M{
		{"$match": n},
		{"$group": bson.M{
			"_id":        nil,
			"bet_rounds": bson.M{"$sum": "$bet_rounds"},
			"bets":       bson.M{"$sum": "$bets"},
			"win_bets":   bson.M{"$sum": "$win_bets"},
			"lose_bets":  bson.M{"$sum": "$lose_bets"},
			"wins":       bson.M{"$sum": "$wins"},
			"loses":      bson.M{"$sum": "$loses"},
		}},
	}
	col.Pipe(m3).One(&carshinfo)
	stat.FTriggerAmount = ComputeFloat(stat.TriggerAmount, 100)
	if stat.AllEvoTimes != 0 {
		stat.TiggerMulpitleSumAvg = stat.TiggerMulpitleSum / float64(stat.AllEvoTimes)
		stat.TiggerBetAvg = (ComputeFloat(int64(stat.TiggerBets), 100) / float64(stat.AllEvoTimes))
		stat.ExpiredRate = ComputeFloat(int64(stat.TriggerRounds), int64(stat.AllEvoTimes)) * 100.0
	}
	if stat.WinRounds != 0 {
		stat.WinMulpitleSumAvg = (ComputeFloat(int64(stat.WinMulpitleSum), 100) / float64(stat.WinRounds)) * 100.0
	}
	if stat.LoseRounds != 0 {
		stat.LoseMulpitleSumAvg = (ComputeFloat(int64(stat.LoseMulpitleSum), 100) / float64(stat.LoseRounds)) * 100.0
	}

	stat.FWinBets = ComputeFloat(int64(stat.WinBets), 100)
	stat.FLoseBets = ComputeFloat(int64(stat.LoseBets), 100)
	stat.FWins = ComputeFloat(int64(stat.Wins), 100)
	stat.FLoses = ComputeFloat(int64(stat.Loses), 100)
	stat.FCash = ComputeFloat(int64(stat.Cash), 100)
	// 查询总局数
	stat.Bets = carshinfo.Bets
	stat.AllRounds = carshinfo.BetRounds
	stat.AllWinBets = carshinfo.Wins
	stat.AllLoseBets = carshinfo.Loses

	stat.FBets = ComputeFloat(int64(stat.Bets), 100)
	if stat.AllRounds != 0 {
		stat.FRoundBetAvg = ComputeFloat(int64(stat.Bets), 100) / float64(stat.AllRounds)
	}
	if stat.PayAmount != 0 {
		// stat.TotalRewardsRate = (ComputeFloat(stat.WithdrawAmount-stat.CarryAmount, stat.PayAmount) / 100.0) * 100.0
		stat.TotalRewardsRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.WithdrawAmount+stat.CarryAmount, stat.PayAmount))*100.0)
	}
	// if -stat.AllLoseBets != 0 {
	// 	stat.UserRewardRate = ComputeFloat(int64(stat.AllWinBets), 100) / ComputeFloat(int64(int64(-stat.AllLoseBets)), 100) * 100.0
	// }
	if -stat.AllLoseBets != 0 {
		stat.UserRewardRate = fmt.Sprintf("%.2f%%", ComputeFloat(int64(stat.AllWinBets), 100)/ComputeFloat(int64(int64(-stat.AllLoseBets)), 100)*100.0)
	}
}

/*
Crash冒险奖励
*/
func (s *gameStatsService) GetCrashMXJLStat(col *mgo.Collection, page, pageSize int, m, m2, m3 bson.M, startDate, endDate string) (
	stats []*entity.CrashMXJLStat, count int, err error,
) {
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)

	// 用户列表筛选
	var filterPlayerIds []string
	if len(m3) > 0 || len(m2) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				"_id":        "$userid",
				"bets":       bson.M{"$sum": "$bets"},
				"bet_rounds": bson.M{"$sum": "$bet_rounds"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = col.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetCrashJCFKStat error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			userid := r["_id"].(string)
			userids = append(userids, userid)
		}
		if len(userids) == 0 { // 未匹配到
			return
		}
		now := utils.BsonNow()
		m3_2 := []bson.M{
			{"$match": bson.M{"_id": bson.M{"$in": userids}}},
			{"$project": bson.M{
				"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
				"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
			}},
			{"$match": m3},
		}
		var r3_2 []bson.M
		err = PlayerUsers.Pipe(m3_2).All(&r3_2)
		if err != nil {
			beego.Error("GetCrashJCFKStat error32:", err)
		} else if len(r3_2) == 0 { // 未匹配到
			return
		}
		for _, r := range r3_2 {
			userid := r["_id"].(string)
			filterPlayerIds = append(filterPlayerIds, userid)
		}
	}
	if len(filterPlayerIds) > 0 {
		m["userid"] = bson.M{"$in": filterPlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":                     "$userid",
			"date":                    bson.M{"$max": "$date"},
			"bets":                    bson.M{"$sum": "$bets"},
			"all_rounds":              bson.M{"$sum": "$all_rounds"},
			"all_win_rounds":          bson.M{"$sum": "$all_win_rounds"},
			"all_win_bets":            bson.M{"$sum": "$all_win_bets"},
			"all_lose_bets":           bson.M{"$sum": "$all_lose_bets"},
			"all_evo_times":           bson.M{"$sum": "$all_evo_times"},
			"tigger_bets":             bson.M{"$sum": "$tigger_bets"},
			"maximum_escape_multiple": bson.M{"$max": "$maximum_escape_multiple"},
			"tigger_mulpitle_sum":     bson.M{"$sum": "$tigger_mulpitle_sum"},
			"win_rounds":              bson.M{"$sum": "$win_rounds"},
			"win_max_mulpitle":        bson.M{"$max": "$win_max_mulpitle"},
			"win_mulpitle_sum":        bson.M{"$sum": "$win_mulpitle_sum"},
			"lose_rounds":             bson.M{"$sum": "$lose_rounds"},
			"lose_mulpitle_sum":       bson.M{"$sum": "$lose_mulpitle_sum"},
			"trigger_rounds":          bson.M{"$sum": "$trigger_rounds"},
			"win_bets":                bson.M{"$sum": "$win_bets"},
			"lose_bets":               bson.M{"$sum": "$lose_bets"},
			"wins":                    bson.M{"$sum": "$wins"},
			"loses":                   bson.M{"$sum": "$loses"},
			"cash":                    bson.M{"$sum": "$cash"},
			"withdraw_amount":         bson.M{"$last": "$withdraw_amount"},
			"pay_amount":              bson.M{"$last": "$pay_amount"},
			"carry_amount":            bson.M{"$last": "$carry_amount"},
		}},
		{"$sort": bson.M{"date": 1}},
		{"$project": bson.M{
			"_id":                     "$_id",
			"bets":                    "$bets",
			"all_rounds":              "$all_rounds",
			"all_win_rounds":          "$all_win_rounds",
			"all_win_bets":            "$all_win_bets",
			"all_lose_bets":           "$all_lose_bets",
			"all_evo_times":           "$all_evo_times",
			"tigger_bets":             "$tigger_bets",
			"maximum_escape_multiple": "$maximum_escape_multiple",
			"tigger_mulpitle_sum":     "$tigger_mulpitle_sum",
			"win_rounds":              "$win_rounds",
			"win_max_mulpitle":        "$win_max_mulpitle",
			"win_mulpitle_sum":        "$win_mulpitle_sum",
			"lose_rounds":             "$lose_rounds",
			"lose_mulpitle_sum":       "$lose_mulpitle_sum",
			"trigger_rounds":          "$trigger_rounds",
			"win_bets":                "$win_bets",
			"lose_bets":               "$lose_bets",
			"wins":                    "$wins",
			"loses":                   "$loses",
			"cash":                    "$cash",
			"withdraw_amount":         "$withdraw_amount",
			"pay_amount":              "$pay_amount",
			"carry_amount":            "$carry_amount",
			// "win_escape_multiple": bson.M{"$cond": []any{bson.M{"$eq": []any{"$win_mulpitle_sum", 0}}, 0, bson.M{"$divide": []string{"$win_mulpitle_sum", "$win_rounds"}}}},
			"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$all_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$all_rounds"}}}},
		}},
		{"$match": m2},
		{"$count": "count"},
	}
	// 查总条数
	var rCount []bson.M
	err = col.Pipe(m1).All(&rCount)
	if err != nil {
		return
	} else if len(rCount) > 0 {
		count = rCount[0]["count"].(int)
	}

	// 查列表数据
	m1 = m1[:len(m1)-1]
	m1 = append(m1, []bson.M{
		{"$sort": bson.M{"date": -1}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}...)
	err = col.Pipe(m1).All(&stats)
	if err != nil {
		return
	}
	new_col := CrashPlayerStats
	if strings.Contains(col.Name, "_plane_") {
		new_col = PlanePlayerStats
	}
	CrashMxjlStatMapping(stats, new_col, m)

	// 查总汇数据
	mark := "--"
	summary := &entity.CrashMXJLStat{}
	m1 = m1[:len(m1)-3]
	m1[1]["$group"].(bson.M)["_id"] = nil
	delete(m1[1]["$group"].(bson.M), "date")
	m1[1]["$group"].(bson.M)["userids"] = bson.M{"$addToSet": "$userid"}
	m1[3]["$project"].(bson.M)["userids"] = "$userids"
	var summarys []*entity.CrashMXJLStat
	err = col.Pipe(m1).All(&summarys)
	if err != nil {
		return
	} else if len(summarys) > 0 {
		summary = summarys[0]
		summary.No = fmt.Sprintf("%s-%s总汇", startDate, endDate)
		summary.Id = mark
		summary.FNickname = mark
		summary.FUserType = mark
		summary.FRegistArea = mark
		summary.FCtime = mark
		summary.FLoginTime = mark
		summary.PlayerRewardRate = mark
		// summary.FLiveDays = mark
		// summary.FLoseDays = mark
		CrashMxjlStatSummaryMapping(summary, new_col, m)
	}
	if len(summarys) > 0 {
		stats = append([]*entity.CrashMXJLStat{summary}, stats...)
	}

	return
}

func CrashMxjlStatMapping(stats []*entity.CrashMXJLStat, col *mgo.Collection, m bson.M) {
	var playerIds []string
	// var playerStats = make(map[string]*entity.CrashPlayerStat, len(stats))
	for _, stat := range stats {
		playerIds = append(playerIds, stat.Id)
		// playerStats[stat.Id] = stat
	}
	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": playerIds}}},
		{"$project": bson.M{
			"_id":         "$_id",
			"nickname":    "$nickname",
			"money":       "$money",
			"cash_out":    "$cash_out",
			"diamond":     "$diamond",
			"ctime":       "$ctime",
			"login_time":  "$login_time",
			"regist_area": "$regist_area", //ab测试 0:A 1:B 2:C
			"state":       "$state",
		}},
	}
	var r2 []bson.M
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("crash users error:", err)
		return
	}
	var r2Map = make(map[string]bson.M, len(r2))
	for _, r := range r2 {
		userid := r["_id"].(string)
		r2Map[userid] = r
	}
	// 查询总数据
	var crashlist []entity.CrashPlayerStat
	n := bson.M{}
	// 手动复制 m 到 n
	for k, v := range m {
		n[k] = v
	}
	n["userid"] = bson.M{"$in": playerIds}
	m3 := []bson.M{
		{"$match": n},
		{"$group": bson.M{
			"_id":        "$userid",
			"bet_rounds": bson.M{"$sum": "$bet_rounds"},
			"win_rounds": bson.M{"$sum": "$win_rounds"},
			"bets":       bson.M{"$sum": "$bets"},
			"win_bets":   bson.M{"$sum": "$win_bets"},
			"lose_bets":  bson.M{"$sum": "$lose_bets"},
			"wins":       bson.M{"$sum": "$wins"},
			"loses":      bson.M{"$sum": "$loses"},
		}},
	}
	col.Pipe(m3).All(&crashlist)
	now := utils.BsonNow()
	for i, stat := range stats {
		r, ok := r2Map[stat.Id]
		if !ok {
			continue
		}
		// 玩家昵称	玩家类型 账号类型 注册时间 最后登陆时间 存活天数 流失天数
		// userid := r["_id"].(string)
		nickname := r["nickname"].(string)
		money := r["money"].(int)
		// cash_out := r["cash_out"].(int)
		// diamond := r["diamond"].(int64)
		ctime := r["ctime"].(time.Time)
		login_time := r["login_time"].(time.Time)
		regist_area := r["regist_area"].(int)
		state := r["state"].(int)

		// 将日期转成印度时间
		login_time, _ = ConvertToIndiaTime(login_time.Unix())
		ctime, _ = ConvertToIndiaTime(ctime.Unix())

		stat.No = strconv.Itoa(i + 1)
		stat.FNickname = nickname
		//stat.Money = uint32(money)
		stat.FCtime = ctime.Format(utils.FORMAT)
		stat.FLoginTime = login_time.Format(utils.FORMAT)
		stat.FLiveDays = int(math.Ceil(login_time.Sub(ctime).Hours() / 24))
		stat.FLoseDays = int(math.Floor(now.Sub(login_time).Hours() / 24))

		switch regist_area {
		case 0:
			stat.FRegistArea = "A类"
		case 1:
			stat.FRegistArea = "B类"
		case 2:
			stat.FRegistArea = "C类"
		}

		switch state {
		case 1:
			stat.FUserType = "新手"
		case 2:
			if money == 0 {
				stat.FUserType = "零充"
			} else if money >= 20000 && money <= 99999 {
				stat.FUserType = "普充"
			} else if money >= 100000 && money <= 499999 {
				stat.FUserType = "小R"
			} else if money >= 500000 && money <= 999999 {
				stat.FUserType = "中R"
			} else if money >= 1000000 && money <= 9999999 {
				stat.FUserType = "大R"
			} else if money >= 10000000 {
				stat.FUserType = "超大R"
			}
		case 3:
			stat.FUserType = "平民"
		case 4:
			stat.FUserType = "泡沫"
		}
		stat.FTiggerBets = ComputeFloat(int64(stat.TiggerBets), 100)
		if stat.AllEvoTimes != 0 {
			stat.TiggerBetAvg = ComputeFloat(int64(stat.TiggerBets), 100) / float64(stat.AllEvoTimes)
			stat.TiggerMulpitleSumAvg = stat.TiggerMulpitleSum / float64(stat.AllEvoTimes)
			// stat.TiggerBetAvg = (ComputeFloat(int64(stat.TiggerBets), 100) / float64(stat.AllEvoTimes)) * 100.0
			stat.SuccessRate = fmt.Sprintf("%.2f%%", ComputeFloat(int64(stat.WinRounds), int64(stat.AllEvoTimes))*100.0)
			stat.StrategyWinRate = fmt.Sprintf("%.2f%%", ComputeFloat(int64(stat.WinRounds), int64(stat.AllEvoTimes))*100.0)
		}
		if stat.WinRounds != 0 {
			stat.WinMulpitleSumAvg = (ComputeFloat(int64(stat.WinMulpitleSum), 100) / float64(stat.WinRounds)) * 100.0
		}
		if stat.LoseRounds != 0 {
			stat.LoseMulpitleSumAvg = (ComputeFloat(int64(stat.LoseMulpitleSum), 100) / float64(stat.LoseRounds)) * 100.0
		}
		stat.FWinBets = ComputeFloat(int64(stat.WinBets), 100)
		stat.FLoseBets = ComputeFloat(int64(stat.LoseBets), 100)
		stat.FWins = ComputeFloat(int64(stat.Wins), 100)
		stat.FLoses = ComputeFloat(int64(stat.Loses), 100)
		stat.FCash = ComputeFloat(int64(stat.Cash), 100)
		if len(crashlist) > 0 {
			for _, c := range crashlist {
				if c.Id == stat.Id {
					stat.Bets = c.Bets
					stat.AllRounds = c.BetRounds
					stat.AllWinRounds = c.WinRounds
					stat.AllWinBets = c.Wins
					stat.AllLoseBets = c.Loses
					break
				}
			}
		}
		stat.FBets = ComputeFloat(int64(stat.Bets), 100)
		if stat.AllRounds != 0 {
			stat.FRoundBetAvg = ComputeFloat(int64(stat.Bets), 100) / float64(stat.AllRounds)
		}
		if stat.PayAmount != 0 {
			stat.PlayerRewardRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.WithdrawAmount+stat.CarryAmount, stat.PayAmount))*100.0)
		}
		if stat.AllRounds != 0 {
			stat.WinRate = fmt.Sprintf("%.2f%%", ComputeFloat(int64(stat.AllWinRounds), int64(stat.AllRounds))*100.0)
		}
		if stat.Loses != 0 {
			stat.StrategyRewardsRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.Wins, -stat.Loses))
		}
		if -stat.AllLoseBets != 0 {
			stat.TotalRewardsRate = fmt.Sprintf("%.2f%%", ComputeFloat(int64(stat.AllWinBets), 100)/ComputeFloat(int64(int64(-stat.AllLoseBets)), 100)*100.0)
		}
	}
}

func CrashMxjlStatSummaryMapping(stat *entity.CrashMXJLStat, col *mgo.Collection, m bson.M) {
	// 查询总数据
	carshinfo := new(entity.CrashPlayerStat)
	n := bson.M{}
	// 手动复制 m 到 n
	for k, v := range m {
		n[k] = v
	}
	n["userid"] = bson.M{"$in": stat.Userids}
	m3 := []bson.M{
		{"$match": n},
		{"$group": bson.M{
			"_id":        nil,
			"bet_rounds": bson.M{"$sum": "$bet_rounds"},
			"win_rounds": bson.M{"$sum": "$win_rounds"},
			"bets":       bson.M{"$sum": "$bets"},
			"win_bets":   bson.M{"$sum": "$win_bets"},
			"lose_bets":  bson.M{"$sum": "$lose_bets"},
			"wins":       bson.M{"$sum": "$wins"},
			"loses":      bson.M{"$sum": "$loses"},
		}},
	}
	col.Pipe(m3).One(&carshinfo)

	stat.FTiggerBets = ComputeFloat(int64(stat.TiggerBets), 100)
	if stat.AllEvoTimes != 0 {
		stat.TiggerBetAvg = ComputeFloat(int64(stat.TiggerBets), 100) / float64(stat.AllEvoTimes)
		stat.TiggerMulpitleSumAvg = stat.TiggerMulpitleSum / float64(stat.AllEvoTimes)
		stat.TiggerBetAvg = (ComputeFloat(int64(stat.TiggerBets), 100) / float64(stat.AllEvoTimes))
		stat.SuccessRate = fmt.Sprintf("%.2f%%", ComputeFloat(int64(stat.WinRounds), int64(stat.AllEvoTimes))*100.0)
		stat.StrategyWinRate = fmt.Sprintf("%.2f%%", ComputeFloat(int64(stat.WinRounds), int64(stat.AllEvoTimes))*100.0)
	}
	stat.FBets = ComputeFloat(stat.Bets, 100)
	if stat.WinRounds != 0 {
		stat.WinMulpitleSumAvg = (ComputeFloat(int64(stat.WinMulpitleSum), 100) / float64(stat.WinRounds)) * 100.0
	}
	if stat.LoseRounds != 0 {
		stat.LoseMulpitleSumAvg = (ComputeFloat(int64(stat.LoseMulpitleSum), 100) / float64(stat.LoseRounds)) * 100.0
	}
	stat.FWinBets = ComputeFloat(int64(stat.WinBets), 100)
	stat.FLoseBets = ComputeFloat(int64(stat.LoseBets), 100)
	stat.FWins = ComputeFloat(int64(stat.Wins), 100)
	stat.FLoses = ComputeFloat(int64(stat.Loses), 100)
	stat.FCash = ComputeFloat(int64(stat.Cash), 100)
	// 查询总局数
	stat.Bets = carshinfo.Bets
	stat.AllRounds = carshinfo.BetRounds
	stat.AllWinRounds = carshinfo.WinRounds
	stat.AllWinBets = carshinfo.Wins
	stat.AllLoseBets = carshinfo.Loses
	stat.FBets = ComputeFloat(int64(stat.Bets), 100)
	if stat.AllRounds != 0 {
		stat.FRoundBetAvg = ComputeFloat(int64(stat.Bets), 100) / float64(stat.AllRounds)
	}
	// if stat.PayAmount != 0 {
	// 	stat.PlayerRewardRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.WithdrawAmount+stat.CarryAmount, stat.PayAmount))*100.0)
	// }
	if stat.AllRounds != 0 {
		stat.WinRate = fmt.Sprintf("%.2f%%", ComputeFloat(int64(stat.AllWinRounds), int64(stat.AllRounds))*100.0)
	}
	if stat.Loses != 0 {
		stat.StrategyRewardsRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.Wins, -stat.Loses))
	}
	if stat.AllLoseBets != 0 {
		stat.TotalRewardsRate = fmt.Sprintf("%.2f%%", ComputeFloat(int64(stat.AllWinBets), 100)/ComputeFloat(int64(int64(-stat.AllLoseBets)), 100)*100.0)
	}
}

/*
Crash人狂有祸
*/
func (s *gameStatsService) GetCrashRKYSStat(col *mgo.Collection, page, pageSize int, m, m2, m3 bson.M, startDate, endDate string) (
	stats []*entity.CrashRKYSStat, count int, err error,
) {
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)

	// 用户列表筛选
	var filterPlayerIds []string
	if len(m3) > 0 || len(m2) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				"_id":        "$userid",
				"bets":       bson.M{"$sum": "$bets"},
				"all_rounds": bson.M{"$sum": "$all_rounds"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$all_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$all_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = col.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetCrashRKYSStat error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			userid := r["_id"].(string)
			userids = append(userids, userid)
		}
		if len(userids) == 0 { // 未匹配到
			return
		}
		now := utils.BsonNow()
		m3_2 := []bson.M{
			{"$match": bson.M{"_id": bson.M{"$in": userids}}},
			{"$project": bson.M{
				"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
				"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
			}},
			{"$match": m3},
		}
		var r3_2 []bson.M
		err = PlayerUsers.Pipe(m3_2).All(&r3_2)
		if err != nil {
			beego.Error("GetCrashRKYSStat error32:", err)
		} else if len(r3_2) == 0 { // 未匹配到
			return
		}
		for _, r := range r3_2 {
			userid := r["_id"].(string)
			filterPlayerIds = append(filterPlayerIds, userid)
		}
	}
	if len(filterPlayerIds) > 0 {
		m["userid"] = bson.M{"$in": filterPlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":                     "$userid",
			"date":                    bson.M{"$max": "$date"},
			"bets":                    bson.M{"$sum": "$bets"},
			"all_rounds":              bson.M{"$sum": "$all_rounds"},
			"all_win_bets":            bson.M{"$sum": "$all_win_bets"},
			"all_lose_bets":           bson.M{"$sum": "$all_lose_bets"},
			"all_evo_times":           bson.M{"$sum": "$all_evo_times"},
			"tigger_bets":             bson.M{"$sum": "$tigger_bets"},
			"maximum_escape_multiple": bson.M{"$max": "$maximum_escape_multiple"},
			"tigger_mulpitle_sum":     bson.M{"$sum": "$tigger_mulpitle_sum"},
			"win_rounds":              bson.M{"$sum": "$win_rounds"},
			"win_max_mulpitle":        bson.M{"$max": "$win_max_mulpitle"},
			"win_mulpitle_sum":        bson.M{"$sum": "$win_mulpitle_sum"},
			"lose_rounds":             bson.M{"$sum": "$lose_rounds"},
			"lose_mulpitle_sum":       bson.M{"$sum": "$lose_mulpitle_sum"},
			"rf_rounds":               bson.M{"$sum": "$rf_rounds"},
			"rf_bets":                 bson.M{"$sum": "$rf_bets"},
			"rf_lose_rounds":          bson.M{"$sum": "$rf_lose_rounds"},
			"rf_reap_bets":            bson.M{"$sum": "$rf_reap_bets"},
			"win_bets":                bson.M{"$sum": "$win_bets"},
			"lose_bets":               bson.M{"$sum": "$lose_bets"},
			"wins":                    bson.M{"$sum": "$wins"},
			"loses":                   bson.M{"$sum": "$loses"},
			"cash":                    bson.M{"$sum": "$cash"},
			"withdraw_amount":         bson.M{"$last": "$withdraw_amount"},
			"pay_amount":              bson.M{"$last": "$pay_amount"},
			"carry_amount":            bson.M{"$last": "$carry_amount"},
		}},
		{"$sort": bson.M{"date": 1}},
		{"$project": bson.M{
			"_id":                     "$_id",
			"bets":                    "$bets",
			"all_rounds":              "$all_rounds",
			"all_win_bets":            "$all_win_bets",
			"all_lose_bets":           "$all_lose_bets",
			"all_evo_times":           "$all_evo_times",
			"tigger_bets":             "$tigger_bets",
			"maximum_escape_multiple": "$maximum_escape_multiple",
			"tigger_mulpitle_sum":     "$tigger_mulpitle_sum",
			"win_rounds":              "$win_rounds",
			"win_max_mulpitle":        "$win_max_mulpitle",
			"win_mulpitle_sum":        "$win_mulpitle_sum",
			"lose_rounds":             "$lose_rounds",
			"lose_mulpitle_sum":       "$lose_mulpitle_sum",
			"rf_rounds":               "$rf_rounds",
			"rf_bets":                 "$rf_bets",
			"rf_lose_rounds":          "$rf_lose_rounds",
			"rf_reap_bets":            "$rf_reap_bets",
			"win_bets":                "$win_bets",
			"lose_bets":               "$lose_bets",
			"wins":                    "$wins",
			"loses":                   "$loses",
			"cash":                    "$cash",
			"withdraw_amount":         "$withdraw_amount",
			"pay_amount":              "$pay_amount",
			"carry_amount":            "$carry_amount",
			// "win_escape_multiple": bson.M{"$cond": []any{bson.M{"$eq": []any{"$win_mulpitle_sum", 0}}, 0, bson.M{"$divide": []string{"$win_mulpitle_sum", "$win_rounds"}}}},
			"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$all_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$all_rounds"}}}},
		}},
		{"$match": m2},
		{"$count": "count"},
	}
	// 查总条数
	var rCount []bson.M
	err = col.Pipe(m1).All(&rCount)
	if err != nil {
		return
	} else if len(rCount) > 0 {
		count = rCount[0]["count"].(int)
	}

	// 查列表数据
	m1 = m1[:len(m1)-1]
	m1 = append(m1, []bson.M{
		{"$sort": bson.M{"date": -1}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}...)
	err = col.Pipe(m1).All(&stats)
	if err != nil {
		return
	}
	new_col := CrashPlayerStats
	if strings.Contains(col.Name, "_plane_") {
		new_col = PlanePlayerStats
	}
	CrashRkysStatMapping(stats, new_col, m)

	// 查总汇数据
	mark := "--"
	summary := &entity.CrashRKYSStat{}
	m1 = m1[:len(m1)-3]
	m1[1]["$group"].(bson.M)["_id"] = nil
	delete(m1[1]["$group"].(bson.M), "date")
	m1[1]["$group"].(bson.M)["userids"] = bson.M{"$addToSet": "$userid"}
	m1[3]["$project"].(bson.M)["userids"] = "$userids"
	var summarys []*entity.CrashRKYSStat
	err = col.Pipe(m1).All(&summarys)
	if err != nil {
		return
	} else if len(summarys) > 0 {
		summary = summarys[0]
		summary.No = fmt.Sprintf("%s-%s总汇", startDate, endDate)
		summary.Id = mark
		summary.FNickname = mark
		summary.FUserType = mark
		summary.FRegistArea = mark
		summary.FCtime = mark
		summary.FLoginTime = mark
		summary.PlayerRewardRate = mark
		// summary.FLiveDays = mark
		// summary.FLoseDays = mark
		CrashRkysStatSummaryMapping(summary, new_col, m)
	}
	if len(summarys) > 0 {
		stats = append([]*entity.CrashRKYSStat{summary}, stats...)
	}

	return
}

func CrashRkysStatMapping(stats []*entity.CrashRKYSStat, col *mgo.Collection, m bson.M) {
	var playerIds []string
	// var playerStats = make(map[string]*entity.CrashPlayerStat, len(stats))
	for _, stat := range stats {
		playerIds = append(playerIds, stat.Id)
		// playerStats[stat.Id] = stat
	}
	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": playerIds}}},
		{"$project": bson.M{
			"_id":         "$_id",
			"nickname":    "$nickname",
			"money":       "$money",
			"cash_out":    "$cash_out",
			"diamond":     "$diamond",
			"ctime":       "$ctime",
			"login_time":  "$login_time",
			"regist_area": "$regist_area", //ab测试 0:A 1:B 2:C
			"state":       "$state",
		}},
	}
	var r2 []bson.M
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("crash users error:", err)
		return
	}
	var r2Map = make(map[string]bson.M, len(r2))
	for _, r := range r2 {
		userid := r["_id"].(string)
		r2Map[userid] = r
	}

	// 查询总数据
	var crashlist []entity.CrashPlayerStat
	n := bson.M{}
	// 手动复制 m 到 n
	for k, v := range m {
		n[k] = v
	}
	n["userid"] = bson.M{"$in": playerIds}
	m3 := []bson.M{
		{"$match": n},
		{"$group": bson.M{
			"_id":        "$userid",
			"bet_rounds": bson.M{"$sum": "$bet_rounds"},
			"win_rounds": bson.M{"$sum": "$win_rounds"},
			"bets":       bson.M{"$sum": "$bets"},
			"win_bets":   bson.M{"$sum": "$win_bets"},
			"lose_bets":  bson.M{"$sum": "$lose_bets"},
			"wins":       bson.M{"$sum": "$wins"},
			"loses":      bson.M{"$sum": "$loses"},
		}},
	}
	col.Pipe(m3).All(&crashlist)
	now := utils.BsonNow()
	for i, stat := range stats {
		r, ok := r2Map[stat.Id]
		if !ok {
			continue
		}
		// 玩家昵称	玩家类型 账号类型 注册时间 最后登陆时间 存活天数 流失天数
		// userid := r["_id"].(string)
		nickname := r["nickname"].(string)
		money := r["money"].(int)
		// cash_out := r["cash_out"].(int)
		// diamond := r["diamond"].(int64)
		ctime := r["ctime"].(time.Time)
		login_time := r["login_time"].(time.Time)
		regist_area := r["regist_area"].(int)
		state := r["state"].(int)

		// 将日期转成印度时间
		login_time, _ = ConvertToIndiaTime(login_time.Unix())
		ctime, _ = ConvertToIndiaTime(ctime.Unix())

		stat.No = strconv.Itoa(i + 1)
		stat.FNickname = nickname
		//stat.Money = uint32(money)
		stat.FCtime = ctime.Format(utils.FORMAT)
		stat.FLoginTime = login_time.Format(utils.FORMAT)
		stat.FLiveDays = int(math.Ceil(login_time.Sub(ctime).Hours() / 24))
		stat.FLoseDays = int(math.Floor(now.Sub(login_time).Hours() / 24))

		switch regist_area {
		case 0:
			stat.FRegistArea = "A类"
		case 1:
			stat.FRegistArea = "B类"
		case 2:
			stat.FRegistArea = "C类"
		}

		switch state {
		case 1:
			stat.FUserType = "新手"
		case 2:
			if money == 0 {
				stat.FUserType = "零充"
			} else if money >= 20000 && money <= 99999 {
				stat.FUserType = "普充"
			} else if money >= 100000 && money <= 499999 {
				stat.FUserType = "小R"
			} else if money >= 500000 && money <= 999999 {
				stat.FUserType = "中R"
			} else if money >= 1000000 && money <= 9999999 {
				stat.FUserType = "大R"
			} else if money >= 10000000 {
				stat.FUserType = "超大R"
			}
		case 3:
			stat.FUserType = "平民"
		case 4:
			stat.FUserType = "泡沫"
		}
		stat.FTiggerBets = ComputeFloat(int64(stat.TiggerBets), 100)
		if stat.AllEvoTimes != 0 {
			stat.TiggerBetAvg = ComputeFloat(int64(stat.TiggerBets), 100) / float64(stat.AllEvoTimes)
			stat.TiggerMulpitleSumAvg = stat.TiggerMulpitleSum / float64(stat.AllEvoTimes)
			// stat.TiggerBetAvg = (ComputeFloat(int64(stat.TiggerBets), 100) / float64(stat.AllEvoTimes)) * 100.0
			stat.SuccessRate = ComputeFloat(int64(stat.WinRounds), int64(stat.AllEvoTimes)) * 100.0
			stat.StrategyWinRate = ComputeFloat(int64(stat.WinRounds), int64(stat.AllEvoTimes)) * 100.0
			stat.RFRate = ComputeFloat(int64(stat.RFLoseRounds), int64(stat.AllEvoTimes)) * 100.0
			stat.RFReapRate = ComputeFloat(stat.RFRounds, stat.RFLoseRounds) * 100.0
		}
		if stat.WinRounds != 0 {
			stat.WinMulpitleSumAvg = (ComputeFloat(int64(stat.WinMulpitleSum), 100) / float64(stat.WinRounds)) * 100.0
		}
		if stat.LoseRounds != 0 {
			stat.LoseMulpitleSumAvg = (ComputeFloat(int64(stat.LoseMulpitleSum), 100) / float64(stat.LoseRounds)) * 100.0
		}
		stat.FRFBets = ComputeFloat(stat.RFBets, 100)
		stat.FRFReapBets = ComputeFloat(stat.RFReapBets, 100)

		stat.FWinBets = ComputeFloat(int64(stat.WinBets), 100)
		stat.FLoseBets = ComputeFloat(int64(stat.LoseBets), 100)
		stat.FWins = ComputeFloat(int64(stat.Wins), 100)
		stat.FLoses = ComputeFloat(int64(stat.Loses), 100)
		stat.FCash = ComputeFloat(int64(stat.Cash), 100)
		if len(crashlist) > 0 {
			for _, c := range crashlist {
				if c.Id == stat.Id {
					stat.Bets = c.Bets
					stat.AllRounds = c.BetRounds
					stat.WinRounds = c.WinRounds
					stat.AllWinBets = c.Wins
					stat.AllLoseBets = c.Loses
					break
				}
			}
		}
		stat.FBets = ComputeFloat(int64(stat.Bets), 100)
		if stat.AllRounds != 0 {
			stat.FRoundBetAvg = ComputeFloat(int64(stat.Bets), 100) / float64(stat.AllRounds)
		}
		if stat.RFRounds != 0 {
			stat.FRFReapBetAvg = ComputeFloat(int64(stat.RFBets), 100) / float64(stat.RFRounds)
		}
		if stat.PayAmount != 0 {
			// stat.PlayerRewardRate = (ComputeFloat(stat.WithdrawAmount-stat.CarryAmount, stat.PayAmount) / 100.0) * 100.0
			stat.PlayerRewardRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.WithdrawAmount+stat.CarryAmount, stat.PayAmount))*100.0)
		}
		if stat.AllRounds != 0 {
			stat.WinRate = ComputeFloat(int64(stat.WinRounds), int64(stat.AllRounds)) * 100.0
		}
		if stat.Loses != 0 {
			stat.StrategyRewardsRate = ComputeFloat(stat.Wins, -stat.Loses)
		}
		if stat.AllLoseBets != 0 {
			// stat.TotalRewardsRate = ComputeFloat(int64(stat.AllWinBets), 100) / ComputeFloat(int64(int64(stat.AllLoseBets)), 100) * 100.0
			stat.TotalRewardsRate = fmt.Sprintf("%.2f%%", ComputeFloat(int64(stat.AllWinBets), 100)/ComputeFloat(int64(int64(-stat.AllLoseBets)), 100)*100.0)
		}
	}
}

func CrashRkysStatSummaryMapping(stat *entity.CrashRKYSStat, col *mgo.Collection, m bson.M) {
	carshinfo := new(entity.CrashPlayerStat)
	n := bson.M{}
	// 手动复制 m 到 n
	for k, v := range m {
		n[k] = v
	}
	n["userid"] = bson.M{"$in": stat.Userids}
	m3 := []bson.M{
		{"$match": n},
		{"$group": bson.M{
			"_id":        nil,
			"bet_rounds": bson.M{"$sum": "$bet_rounds"},
			"win_rounds": bson.M{"$sum": "$win_rounds"},
			"bets":       bson.M{"$sum": "$bets"},
			"win_bets":   bson.M{"$sum": "$win_bets"},
			"lose_bets":  bson.M{"$sum": "$lose_bets"},
			"wins":       bson.M{"$sum": "$wins"},
			"loses":      bson.M{"$sum": "$loses"},
		}},
	}
	col.Pipe(m3).One(&carshinfo)
	stat.FTiggerBets = ComputeFloat(int64(stat.TiggerBets), 100)
	if stat.AllEvoTimes != 0 {
		stat.TiggerBetAvg = ComputeFloat(int64(stat.TiggerBets), 100) / float64(stat.AllEvoTimes)
		stat.TiggerMulpitleSumAvg = stat.TiggerMulpitleSum / float64(stat.AllEvoTimes)
		stat.TiggerBetAvg = (ComputeFloat(int64(stat.TiggerBets), 100) / float64(stat.AllEvoTimes))
		stat.SuccessRate = ComputeFloat(int64(stat.WinRounds), int64(stat.AllEvoTimes)) * 100.0
		stat.StrategyWinRate = ComputeFloat(int64(stat.WinRounds), int64(stat.AllEvoTimes)) * 100.0
		stat.RFRate = ComputeFloat(int64(stat.RFLoseRounds), int64(stat.AllEvoTimes)) * 100.0
		stat.RFReapRate = ComputeFloat(stat.RFRounds, stat.RFLoseRounds) * 100.0
	}
	stat.FBets = ComputeFloat(stat.Bets, 100)
	if stat.WinRounds != 0 {
		stat.WinMulpitleSumAvg = (ComputeFloat(int64(stat.WinMulpitleSum), 100) / float64(stat.WinRounds)) * 100.0
	}
	if stat.LoseRounds != 0 {
		stat.LoseMulpitleSumAvg = (ComputeFloat(int64(stat.LoseMulpitleSum), 100) / float64(stat.LoseRounds)) * 100.0
	}

	stat.FRFBets = ComputeFloat(stat.RFBets, 100)
	stat.FRFReapBets = ComputeFloat(stat.RFReapBets, 100)

	stat.FWinBets = ComputeFloat(int64(stat.WinBets), 100)
	stat.FLoseBets = ComputeFloat(int64(stat.LoseBets), 100)
	stat.FWins = ComputeFloat(int64(stat.Wins), 100)
	stat.FLoses = ComputeFloat(int64(stat.Loses), 100)
	stat.FCash = ComputeFloat(int64(stat.Cash), 100)
	// 查询总局数
	stat.Bets = carshinfo.Bets
	stat.AllRounds = carshinfo.BetRounds
	stat.WinRounds = carshinfo.WinRounds
	stat.AllWinBets = carshinfo.Wins
	stat.AllLoseBets = carshinfo.Loses
	stat.FBets = ComputeFloat(int64(stat.Bets), 100)
	if stat.AllRounds != 0 {
		stat.FRoundBetAvg = ComputeFloat(int64(stat.Bets), 100) / float64(stat.AllRounds)
	}
	if stat.RFRounds != 0 {
		stat.FRFReapBetAvg = ComputeFloat(int64(stat.RFBets), 100) / float64(stat.RFRounds)
	}
	if stat.PayAmount != 0 {
		// stat.PlayerRewardRate = (ComputeFloat(stat.WithdrawAmount-stat.CarryAmount, stat.PayAmount) / 100.0) * 100.0
		// stat.PlayerRewardRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.WithdrawAmount+stat.CarryAmount, stat.PayAmount))*100.0)
	}
	if stat.AllRounds != 0 {
		stat.WinRate = ComputeFloat(int64(stat.WinRounds), int64(stat.AllRounds)) * 100.0
	}
	if stat.Loses != 0 {
		stat.StrategyRewardsRate = ComputeFloat(stat.Wins, -stat.Loses)
	}
	if stat.AllLoseBets != 0 {
		// stat.TotalRewardsRate = ComputeFloat(int64(stat.AllWinBets), 100) / ComputeFloat(int64(int64(stat.AllLoseBets)), 100) * 100.0
		stat.TotalRewardsRate = fmt.Sprintf("%.2f%%", ComputeFloat(int64(stat.AllWinBets), 100)/ComputeFloat(int64(int64(-stat.AllLoseBets)), 100)*100.0)
	}
}

/*
TP 游戏统计
*/
// GetTPStatPlayers tp玩家明细
func (s *gameStatsService) GetTPStatPlayers(page, pageSize int, m, m2, m3 bson.M, startDate, endDate string) (
	stats []*entity.TPPlayerStat, count int, err error,
) {
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)

	// 用户列表筛选
	var filterPlayerIds []string
	if len(m3) > 0 || len(m2) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				"_id":        "$userid",
				"bets":       bson.M{"$sum": "$bets"},
				"bet_rounds": bson.M{"$sum": "$bet_rounds"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = TPPlayerStats.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetCrashStatPlayers error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			userid := r["_id"].(string)
			userids = append(userids, userid)
		}
		if len(userids) == 0 { // 未匹配到
			return
		}
		now := utils.BsonNow()
		m3_2 := []bson.M{
			{"$match": bson.M{"_id": bson.M{"$in": userids}}},
			{"$project": bson.M{
				"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
				"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
			}},
			{"$match": m3},
		}
		var r3_2 []bson.M
		err = PlayerUsers.Pipe(m3_2).All(&r3_2)
		if err != nil {
			beego.Error("GetTPStatPlayers error32:", err)
		} else if len(r3_2) == 0 { // 未匹配到
			return
		}
		for _, r := range r3_2 {
			userid := r["_id"].(string)
			filterPlayerIds = append(filterPlayerIds, userid)
		}
	}
	if len(filterPlayerIds) > 0 {
		m["userid"] = bson.M{"$in": filterPlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":                     "$userid",
			"date":                    bson.M{"$max": "$date"},
			"all_rounds":              bson.M{"$sum": "$all_rounds"},
			"game_times":              bson.M{"$sum": "$game_times"},
			"bets":                    bson.M{"$sum": "$bets"},
			"bet_rounds":              bson.M{"$sum": "$bet_rounds"},
			"observe_rounds":          bson.M{"$sum": "$observe_rounds"},
			"win_rounds":              bson.M{"$sum": "$win_rounds"},
			"lose_rounds":             bson.M{"$sum": "$lose_rounds"},
			"tie_rounds":              bson.M{"$sum": "$tie_rounds"},
			"win_bets":                bson.M{"$sum": "$win_bets"},
			"lose_bets":               bson.M{"$sum": "$lose_bets"},
			"tie_bets":                bson.M{"$sum": "$tie_bets"},
			"wins":                    bson.M{"$sum": "$wins"},
			"loses":                   bson.M{"$sum": "$loses"},
			"cash":                    bson.M{"$sum": "$cash"},
			"mulpitle_sum":            bson.M{"$sum": "$mulpitle_sum"},
			"win_escape_mulpitle_sum": bson.M{"$sum": "$win_escape_mulpitle_sum"},
			"win_mulpitle_sum":        bson.M{"$sum": "$win_mulpitle_sum"},
			"lose_mulpitle_sum":       bson.M{"$sum": "$lose_mulpitle_sum"},
			"mulpitles":               bson.M{"$push": "$mulpitles"},
			"win_escape_mulpitles":    bson.M{"$push": "$win_escape_mulpitles"},
			"win_mulpitles":           bson.M{"$push": "$win_mulpitles"},
			"lose_mulpitles":          bson.M{"$push": "$lose_mulpitles"},
		}},
		{"$project": bson.M{
			"_id": "$_id", "all_rounds": "$all_rounds", "game_times": "$game_times", "bets": "$bets", "bet_rounds": "$bet_rounds", "observe_rounds": "$observe_rounds", "win_rounds": "$win_rounds", "lose_rounds": "$lose_rounds", "tie_rounds": "$tie_rounds", "win_bets": "$win_bets", "lose_bets": "$lose_bets", "tie_bets": "$tie_bets", "wins": "$wins", "loses": "$loses", "cash": "$cash", "mulpitle_sum": "$mulpitle_sum", "win_escape_mulpitle_sum": "$win_escape_mulpitle_sum", "win_mulpitle_sum": "$win_mulpitle_sum", "lose_mulpitle_sum": "$lose_mulpitle_sum", "mulpitles": "$mulpitles", "win_escape_mulpitles": "$win_escape_mulpitles", "win_mulpitles": "$win_mulpitles", "lose_mulpitles": "$lose_mulpitles",
			// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}}, // "can't $divide by zero"
			"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
		}},
		{"$match": m2},
		{"$count": "count"},
	}
	// 查总条数
	var rCount []bson.M
	err = TPPlayerStats.Pipe(m1).All(&rCount)
	if err != nil {
		return
	} else if len(rCount) > 0 {
		count = rCount[0]["count"].(int)
	}

	// 查列表数据
	m1 = m1[:len(m1)-1]
	m1 = append(m1, []bson.M{
		{"$sort": bson.M{"date": -1}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}...)
	err = TPPlayerStats.Pipe(m1).All(&stats)
	if err != nil {
		return
	}
	tpStatPlayersMapping(stats)

	// 查总汇数据
	mark := "--"
	summary := &entity.TPPlayerStat{}
	m1 = m1[:len(m1)-3]
	m1[1]["$group"].(bson.M)["_id"] = nil
	delete(m1[1]["$group"].(bson.M), "date")
	m1[1]["$group"].(bson.M)["userids"] = bson.M{"$addToSet": "$userid"}
	m1[2]["$project"].(bson.M)["userids"] = "$userids"
	var summarys []*entity.TPPlayerStat
	err = TPPlayerStats.Pipe(m1).All(&summarys)
	if err != nil {
		return
	} else if len(summarys) > 0 {
		summary = summarys[0]
		summary.No = fmt.Sprintf("%s-%s总汇", startDate, endDate)
		summary.Id = mark
		summary.FNickname = mark
		summary.FUserType = mark
		summary.FRegistArea = mark
		summary.FCtime = mark
		summary.FLoginTime = mark
		// summary.FLiveDays = mark
		// summary.FLoseDays = mark
		tpStatPlayersSummaryMapping(summary)
	}
	if len(summarys) > 0 {
		stats = append([]*entity.TPPlayerStat{summary}, stats...)
	}
	return
}

func tpStatPlayersMapping(stats []*entity.TPPlayerStat) {
	var playerIds []string
	// var playerStats = make(map[string]*entity.CrashPlayerStat, len(stats))
	for _, stat := range stats {
		playerIds = append(playerIds, stat.Id)
		// playerStats[stat.Id] = stat
	}
	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": playerIds}}},
		{"$project": bson.M{
			"_id":         "$_id",
			"nickname":    "$nickname",
			"money":       "$money",
			"cash_out":    "$cash_out",
			"diamond":     "$diamond",
			"ctime":       "$ctime",
			"login_time":  "$login_time",
			"regist_area": "$regist_area", //ab测试 0:A 1:B 2:C
			"state":       "$state",
		}},
	}
	var r2 []bson.M
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("crash users error:", err)
		return
	}
	var r2Map = make(map[string]bson.M, len(r2))
	for _, r := range r2 {
		userid := r["_id"].(string)
		r2Map[userid] = r
	}
	now := utils.BsonNow()
	for i, stat := range stats {
		r, ok := r2Map[stat.Id]
		if !ok {
			continue
		}
		// 玩家昵称	玩家类型 账号类型 注册时间 最后登陆时间 存活天数 流失天数
		// userid := r["_id"].(string)
		nickname := r["nickname"].(string)
		money := r["money"].(int)
		cash_out := r["cash_out"].(int)
		diamond := r["diamond"].(int64)
		ctime := r["ctime"].(time.Time)
		login_time := r["login_time"].(time.Time)
		regist_area := r["regist_area"].(int)
		state := r["state"].(int)

		// 将日期转成印度时间
		login_time, _ = ConvertToIndiaTime(login_time.Unix())
		ctime, _ = ConvertToIndiaTime(ctime.Unix())

		stat.No = strconv.Itoa(i + 1)
		stat.FNickname = nickname
		stat.Money = uint32(money)
		stat.FCtime = ctime.Format(utils.FORMAT)
		stat.FLoginTime = login_time.Format(utils.FORMAT)
		stat.FLiveDays = int(math.Ceil(login_time.Sub(ctime).Hours() / 24))
		stat.FLoseDays = int(math.Floor(now.Sub(login_time).Hours() / 24))

		// summary.FLiveDays += stat.FLiveDays
		// summary.FLoseDays += stat.FLoseDays
		// summary.AllRounds += stat.AllRounds
		// summary.GameTimes += stat.GameTimes
		// summary.Bets += stat.Bets
		// summary.BetRounds += stat.BetRounds
		// summary.WinRounds += stat.WinRounds
		// summary.LoseRounds += stat.LoseRounds
		// summary.TieRounds += stat.TieRounds
		// summary.WinBets += stat.WinBets
		// summary.LoseBets += stat.LoseBets
		// summary.TieBets += stat.TieBets
		// summary.Wins += stat.Wins
		// summary.Loses += stat.Loses
		// summary.Cash += stat.Cash
		// summary.MulpitleSum += stat.MulpitleSum
		// summary.WinEscapeMulpitleSum += stat.WinEscapeMulpitleSum
		// summary.WinMulpitleSum += stat.WinMulpitleSum
		// summary.LoseMulpitleSum += stat.LoseMulpitleSum
		// summary.Mulpitles = append(summary.Mulpitles, stat.Mulpitles...)
		// summary.WinEscapeMulpitles = append(summary.WinEscapeMulpitles, stat.WinEscapeMulpitles...)
		// summary.WinMulpitles = append(summary.WinMulpitles, stat.WinMulpitles...)
		// summary.LoseMulpitles = append(summary.LoseMulpitles, stat.LoseMulpitles...)
		// summary.ObserveRounds += stat.ObserveRounds
		// summary.SMoney += money
		// summary.SCashOut += cash_out
		// summary.SDiamond += diamond

		switch regist_area {
		case 0:
			stat.FRegistArea = "A类"
		case 1:
			stat.FRegistArea = "B类"
		case 2:
			stat.FRegistArea = "C类"
		}

		switch state {
		case 1:
			stat.FUserType = "新手"
		case 2:
			if money == 0 {
				stat.FUserType = "零充"
			} else if money >= 20000 && money <= 99999 {
				stat.FUserType = "普充"
			} else if money >= 100000 && money <= 499999 {
				stat.FUserType = "小R"
			} else if money >= 500000 && money <= 999999 {
				stat.FUserType = "中R"
			} else if money >= 1000000 && money <= 9999999 {
				stat.FUserType = "大R"
			} else if money >= 10000000 {
				stat.FUserType = "超大R"
			}
		case 3:
			stat.FUserType = "平民"
		case 4:
			stat.FUserType = "泡沫"
		}

		stat.FGameTimes = fmt.Sprintf("%d分%d秒", stat.GameTimes/60, stat.GameTimes&60)
		stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
		if stat.BetRounds > 0 {
			// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.BetRounds)/100.0)
			stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
		}
		stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
		stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
		if stat.BetRounds > 0 {
			stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
		}
		if stat.Bets > 0 {
			stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
		}
		if stat.WinRounds > 0 {
			stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
		}
		stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
		stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
		stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)
		if stat.WinRounds > 0 {
			stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FMulpitleAvg = fmt.Sprintf("%.2f", stat.MulpitleSum/float64(stat.BetRounds))
		}
		if len(stat.Mulpitles) > 0 {
			var mulpitles []float64
			for _, m := range stat.Mulpitles {
				mulpitles = append(mulpitles, m...)
			}
			sort.Slice(mulpitles, func(i, j int) bool {
				return mulpitles[i] < mulpitles[j]
			})
			length := len(mulpitles)
			if length > 0 {
				if length%2 == 1 {
					stat.FMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
				} else {
					// 中间两位算平均
					stat.FMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
				}
			}
		}
		// 赢局局均逃脱倍数
		var winEscapeMulpitleAvg float64
		if stat.WinRounds > 0 {
			winEscapeMulpitleAvg = stat.WinEscapeMulpitleSum / float64(stat.WinRounds)
			stat.FWinEscapeMulpitleAvg = fmt.Sprintf("%.2f", winEscapeMulpitleAvg)
		}
		if len(stat.WinEscapeMulpitles) > 0 {
			var mulpitles []float64
			for _, m := range stat.WinEscapeMulpitles {
				mulpitles = append(mulpitles, m...)
			}
			sort.Slice(mulpitles, func(i, j int) bool {
				return mulpitles[i] < mulpitles[j]
			})
			length := len(mulpitles)
			if length > 0 {
				// 赢局最高逃脱倍数
				stat.FWinEscapeMulpitleMax = fmt.Sprintf("%.2f", mulpitles[len(mulpitles)-1])
				if length%2 == 1 {
					stat.FWinEscapeMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
				} else {
					// 中间两位算平均
					stat.FWinEscapeMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
				}
			}
		}

		// 赢局局均实际爆炸倍数
		var winMulpitleAvg float64
		if stat.WinRounds > 0 {
			winMulpitleAvg = stat.WinMulpitleSum / float64(stat.WinRounds)
			stat.FWinMulpitleAvg = fmt.Sprintf("%.2f", winMulpitleAvg)
		}
		if len(stat.WinMulpitles) > 0 {
			var mulpitles []float64
			for _, m := range stat.WinMulpitles {
				mulpitles = append(mulpitles, m...)
			}
			sort.Slice(mulpitles, func(i, j int) bool {
				return mulpitles[i] < mulpitles[j]
			})
			length := len(mulpitles)
			if length > 0 {
				if length%2 == 1 {
					stat.FWinMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
				} else {
					// 中间两位算平均
					stat.FWinMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
				}
			}
		}
		// 倍差空间
		stat.FMulpitleDiff = fmt.Sprintf("%.2f", winMulpitleAvg-winEscapeMulpitleAvg)

		// 输局局均实际爆炸倍数
		var loseMulpitleAvg float64
		if stat.LoseRounds > 0 {
			loseMulpitleAvg = stat.LoseMulpitleSum / float64(stat.LoseRounds)
			stat.FLoseMulpitleAvg = fmt.Sprintf("%.2f", loseMulpitleAvg)
		}
		if len(stat.LoseMulpitles) > 0 {
			var mulpitles []float64
			for _, m := range stat.LoseMulpitles {
				mulpitles = append(mulpitles, m...)
			}
			sort.Slice(mulpitles, func(i, j int) bool {
				return mulpitles[i] < mulpitles[j]
			})
			length := len(mulpitles)
			if length > 0 {
				if length%2 == 1 {
					stat.FLoseMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
				} else {
					// 中间两位算平均
					stat.FLoseMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
				}
			}
		}
		if loseMulpitleAvg > 0 {
			stat.FWinEscapeRate = fmt.Sprintf("%.2f", winEscapeMulpitleAvg/loseMulpitleAvg)
			stat.FWinLoseRate = fmt.Sprintf("%.2f", winMulpitleAvg/loseMulpitleAvg)
		}
		if stat.BetRounds > 0 {
			stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
		}
		if stat.BetRounds > 0 {
			stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
		}
		stat.FMoney = fmt.Sprintf("%.2f", float64(money)/100)
		stat.FCashOut = fmt.Sprintf("%.2f", float64(cash_out)/100)
		stat.FWinMoney = fmt.Sprintf("%.2f", float64(int(diamond)+cash_out-money)/100)
	}
}

func tpStatPlayersSummaryMapping(stat *entity.TPPlayerStat) {

	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": stat.Userids}}},
		{"$group": bson.M{
			"_id":      nil,
			"money":    bson.M{"$sum": "$money"},
			"cash_out": bson.M{"$sum": "$cash_out"},
			"diamond":  bson.M{"$sum": "$diamond"},
		}},
	}
	var r2 []bson.M
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("crash users error:", err)
		return
	} else if len(r2) > 0 {
		money := r2[0]["money"].(int)
		cash_out := r2[0]["cash_out"].(int)
		diamond := r2[0]["diamond"].(int64)
		stat.SMoney = money
		stat.SCashOut = cash_out
		stat.SDiamond = diamond
	}

	stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
	if stat.BetRounds > 0 {
		stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.BetRounds)/100.0)
		// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
	}

	stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
	stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
	if stat.BetRounds > 0 {
		stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
	}
	if stat.Bets > 0 {
		stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
	}
	if stat.WinRounds > 0 {
		stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
	}
	stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
	stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
	stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)
	if stat.WinRounds > 0 {
		stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FMulpitleAvg = fmt.Sprintf("%.2f", stat.MulpitleSum/float64(stat.BetRounds))
	}
	if len(stat.Mulpitles) > 0 {
		var mulpitles []float64
		for _, m := range stat.Mulpitles {
			mulpitles = append(mulpitles, m...)
		}
		sort.Slice(mulpitles, func(i, j int) bool {
			return mulpitles[i] < mulpitles[j]
		})
		length := len(mulpitles)
		if length > 0 {
			if length%2 == 1 {
				stat.FMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
			} else {
				// 中间两位算平均
				stat.FMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
			}
		}
	}
	// 赢局局均逃脱倍数
	var winEscapeMulpitleAvg float64
	if stat.WinRounds > 0 {
		winEscapeMulpitleAvg = stat.WinEscapeMulpitleSum / float64(stat.WinRounds)
		stat.FWinEscapeMulpitleAvg = fmt.Sprintf("%.2f", winEscapeMulpitleAvg)
	}
	if len(stat.WinEscapeMulpitles) > 0 {
		var mulpitles []float64
		for _, m := range stat.WinEscapeMulpitles {
			mulpitles = append(mulpitles, m...)
		}
		sort.Slice(mulpitles, func(i, j int) bool {
			return mulpitles[i] < mulpitles[j]
		})
		length := len(mulpitles)
		if length > 0 {
			// 赢局最高逃脱倍数
			stat.FWinEscapeMulpitleMax = fmt.Sprintf("%.2f", mulpitles[len(mulpitles)-1])
			if length%2 == 1 {
				stat.FWinEscapeMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
			} else {
				// 中间两位算平均
				stat.FWinEscapeMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
			}
		}
	}

	// 赢局局均实际爆炸倍数
	var winMulpitleAvg float64
	if stat.WinRounds > 0 {
		winMulpitleAvg = stat.WinMulpitleSum / float64(stat.WinRounds)
		stat.FWinMulpitleAvg = fmt.Sprintf("%.2f", winMulpitleAvg)
	}
	if len(stat.WinMulpitles) > 0 {
		var mulpitles []float64
		for _, m := range stat.WinMulpitles {
			mulpitles = append(mulpitles, m...)
		}
		sort.Slice(mulpitles, func(i, j int) bool {
			return mulpitles[i] < mulpitles[j]
		})
		length := len(mulpitles)
		if length > 0 {
			if length%2 == 1 {
				stat.FWinMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
			} else {
				// 中间两位算平均
				stat.FWinMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
			}
		}
	}
	// 倍差空间
	stat.FMulpitleDiff = fmt.Sprintf("%.2f", winMulpitleAvg-winEscapeMulpitleAvg)

	// 输局局均实际爆炸倍数
	var loseMulpitleAvg float64
	if stat.LoseRounds > 0 {
		loseMulpitleAvg = stat.LoseMulpitleSum / float64(stat.LoseRounds)
		stat.FLoseMulpitleAvg = fmt.Sprintf("%.2f", loseMulpitleAvg)
	}
	if len(stat.LoseMulpitles) > 0 {
		var mulpitles []float64
		for _, m := range stat.LoseMulpitles {
			mulpitles = append(mulpitles, m...)
		}
		sort.Slice(mulpitles, func(i, j int) bool {
			return mulpitles[i] < mulpitles[j]
		})
		length := len(mulpitles)
		if length > 0 {
			if length%2 == 1 {
				stat.FLoseMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
			} else {
				// 中间两位算平均
				stat.FLoseMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
			}
		}
	}
	if loseMulpitleAvg > 0 {
		stat.FWinEscapeRate = fmt.Sprintf("%.2f", winEscapeMulpitleAvg/loseMulpitleAvg)
		stat.FWinLoseRate = fmt.Sprintf("%.2f", winMulpitleAvg/loseMulpitleAvg)
	}
	if stat.BetRounds > 0 {
		stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
	}
	if stat.BetRounds > 0 {
		stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
	}
	stat.FMoney = fmt.Sprintf("%.2f", float64(stat.SMoney)/100)
	stat.FCashOut = fmt.Sprintf("%.2f", float64(stat.SCashOut)/100)
	stat.FWinMoney = fmt.Sprintf("%.2f", float64(int(stat.SDiamond)+stat.SCashOut-stat.SMoney)/100)
}

// GetTPStatDates tp汇总明细
func (s *gameStatsService) GetTPStatDates(m, m2, m3 bson.M, startDate, endDate string) (stats []*entity.TPPlayerStat, err error) {
	// 用户列表筛选
	var filterDatePlayerIds []string
	if len(m2) > 0 || len(m3) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				// "_id":        bson.M{"userid":"$userid", "date", "$date"},
				"_id":        "$_id",
				"bets":       bson.M{"$sum": "$bets"},
				"bet_rounds": bson.M{"$sum": "$bet_rounds"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = TPPlayerStats.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetCrashStatPlayers error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			id := r["_id"].(string)
			filterDatePlayerIds = append(filterDatePlayerIds, id)
			ids := strings.Split(id, "-") // id=date-userid
			if len(ids) >= 2 {
				userids = append(userids, ids[1])
			}
		}
		if len(filterDatePlayerIds) == 0 { // 未匹配到
			return
		}
		if len(m3) > 0 { // 流失天数，局均打码量
			if len(userids) == 0 { // 未匹配到
				return
			}
			now := utils.BsonNow()
			m3_2 := []bson.M{
				{"$match": bson.M{"_id": bson.M{"$in": userids}}},
				{"$project": bson.M{
					"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
					"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
				}},
				{"$match": m3},
			}
			var r3_2 []bson.M
			err = PlayerUsers.Pipe(m3_2).All(&r3_2)
			if err != nil {
				beego.Error("GetCrashStatPlayers error32:", err)
			} else if len(r3_2) == 0 { // 未匹配到
				return
			}
			var filterId2 []string
			for _, id := range filterDatePlayerIds {
				for _, r := range r3_2 {
					userid := r["_id"].(string)
					if strings.HasSuffix(id, "-"+userid) {
						filterId2 = append(filterId2, id)
						break
					}
				}
			}
			if len(filterId2) == 0 { // 未匹配到
				return
			}
			filterDatePlayerIds = filterId2
		}
	}
	if len(filterDatePlayerIds) > 0 {
		m["_id"] = bson.M{"$in": filterDatePlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":                     "$date_str",
			"date":                    bson.M{"$min": "$date"},
			"players":                 bson.M{"$sum": 1},
			"new_players":             bson.M{"$sum": bson.M{"$cond": bson.M{"if": "$new_reg", "then": 1, "else": 0}}},
			"old_players":             bson.M{"$sum": bson.M{"$cond": bson.M{"if": "$new_reg", "then": 0, "else": 1}}},
			"all_rounds_list":         bson.M{"$push": "$bet_rounds"},
			"all_rounds":              bson.M{"$sum": "$all_rounds"},
			"game_times":              bson.M{"$sum": "$game_times"},
			"bets":                    bson.M{"$sum": "$bets"},
			"bet_rounds":              bson.M{"$sum": "$bet_rounds"},
			"observe_rounds":          bson.M{"$sum": "$observe_rounds"},
			"win_rounds":              bson.M{"$sum": "$win_rounds"},
			"lose_rounds":             bson.M{"$sum": "$lose_rounds"},
			"tie_rounds":              bson.M{"$sum": "$tie_rounds"},
			"win_bets":                bson.M{"$sum": "$win_bets"},
			"lose_bets":               bson.M{"$sum": "$lose_bets"},
			"tie_bets":                bson.M{"$sum": "$tie_bets"},
			"wins":                    bson.M{"$sum": "$wins"},
			"loses":                   bson.M{"$sum": "$loses"},
			"cash":                    bson.M{"$sum": "$cash"},
			"mulpitle_sum":            bson.M{"$sum": "$mulpitle_sum"},
			"win_escape_mulpitle_sum": bson.M{"$sum": "$win_escape_mulpitle_sum"},
			"win_mulpitle_sum":        bson.M{"$sum": "$win_mulpitle_sum"},
			"lose_mulpitle_sum":       bson.M{"$sum": "$lose_mulpitle_sum"},
			"mulpitles":               bson.M{"$push": "$mulpitles"},
			"win_escape_mulpitles":    bson.M{"$push": "$win_escape_mulpitles"},
			"win_mulpitles":           bson.M{"$push": "$win_mulpitles"},
			"lose_mulpitles":          bson.M{"$push": "$lose_mulpitles"},
			"round_bet_avg":           bson.M{"$sum": "$round_bet_avg"},
		}},
		{"$sort": bson.M{"date": -1}},
	}
	err = TPPlayerStats.Pipe(m1).All(&stats)
	if err != nil {
		return
	}

	// 查询日活数据
	var dateLoginedUsers = make(map[int64]int)
	m9 := []bson.M{
		{"$match": bson.M{"date": m["date"]}},
		{"$project": bson.M{"date": "$date", "logined_users": "$logined_users"}},
	}
	var r9 []bson.M
	err = PddStats.Pipe(m9).All(&r9)
	if err != nil {
		beego.Error("GetCrashStatDates logined_users error: ", err)
	} else {
		for _, r := range r9 {
			date := r["date"].(int64)
			logined_users := r["logined_users"].(int)
			dateLoginedUsers[date] = logined_users
		}
	}

	mark := "--"
	summary := &entity.TPPlayerStat{
		DateStr: fmt.Sprintf("%s-%s总汇", startDate, endDate),
		Id:      mark,
	}
	tpStatDatesMapping(summary, stats, dateLoginedUsers)
	tpStatDatesSummaryMapping(summary)
	stats = append([]*entity.TPPlayerStat{summary}, stats...)
	return
}

func tpStatDatesMapping(summary *entity.TPPlayerStat, stats []*entity.TPPlayerStat, dateLoginedUsers map[int64]int) {
	for _, stat := range stats {
		dateStr := stat.Id
		date, _ := utils.Unix(dateStr)
		stat.Date = date
		stat.DateStr = dateStr
		loginedUsers := dateLoginedUsers[stat.Date] // 日活

		stat.DateStr = utils.Stamp2Time(stat.Date).Format("2006-01-02")
		stat.FLoginedUsers = loginedUsers

		summary.FLoginedUsers += stat.FLoginedUsers
		summary.FLiveDays += stat.FLiveDays
		summary.FLoseDays += stat.FLoseDays
		summary.AllRounds += stat.AllRounds
		summary.GameTimes += stat.GameTimes
		summary.Bets += stat.Bets
		summary.BetRounds += stat.BetRounds
		summary.WinRounds += stat.WinRounds
		summary.LoseRounds += stat.LoseRounds
		summary.TieRounds += stat.TieRounds
		summary.WinBets += stat.WinBets
		summary.LoseBets += stat.LoseBets
		summary.TieBets += stat.TieBets
		summary.Wins += stat.Wins
		summary.Loses += stat.Loses
		summary.Cash += stat.Cash
		summary.MulpitleSum += stat.MulpitleSum
		summary.WinEscapeMulpitleSum += stat.WinEscapeMulpitleSum
		summary.WinMulpitleSum += stat.WinMulpitleSum
		summary.LoseMulpitleSum += stat.LoseMulpitleSum
		summary.Mulpitles = append(summary.Mulpitles, stat.Mulpitles...)
		summary.WinEscapeMulpitles = append(summary.WinEscapeMulpitles, stat.WinEscapeMulpitles...)
		summary.WinMulpitles = append(summary.WinMulpitles, stat.WinMulpitles...)
		summary.LoseMulpitles = append(summary.LoseMulpitles, stat.LoseMulpitles...)
		summary.ObserveRounds += stat.ObserveRounds
		summary.Players += stat.Players
		summary.NewPlayers += stat.NewPlayers
		summary.OldPlayers += stat.OldPlayers
		summary.AllRoundsList = append(summary.AllRoundsList, stat.AllRoundsList...)

		// 玩 crash 人数占日活比
		if stat.FLoginedUsers > 0 {
			stat.FPlayerRate = fmt.Sprintf("%.2f", float64(stat.Players)/float64(stat.FLoginedUsers))
		}

		// 查询当天注册的新用户
		startTime := utils.Str2Time(fmt.Sprintf("%s 00:00:00", dateStr), Location())
		endTime := utils.Str2Time(fmt.Sprintf("%s 23:59:59", dateStr), Location())
		m7 := []bson.M{
			{"$match": bson.M{
				"ctime":            bson.M{"$gte": startTime, "$lte": endTime},
				"robot":            false,
				"simulation_robot": false,
			}},
			{"$group": bson.M{
				"_id":   nil,
				"total": bson.M{"$sum": 1},
			}},
		}
		var r7 []bson.M
		err := PlayerUsers.Pipe(m7).All(&r7)
		if err != nil {
			beego.Error("查询当天注册的新用户 error: ", err)
		} else if len(r7) > 0 {
			stat.FLoginedNewUsers = r7[0]["total"].(int)
		}
		stat.FLoginedOldUsers = stat.FLoginedUsers - stat.FLoginedNewUsers
		if stat.FLoginedOldUsers < 0 {
			stat.FLoginedOldUsers = 0
		}
		summary.FLoginedNewUsers += stat.FLoginedNewUsers
		summary.FLoginedOldUsers += stat.FLoginedOldUsers
		// FNewPlayerRate, FOldPlayerRate 占新老玩家日活
		if stat.FLoginedNewUsers > 0 {
			stat.FNewPlayerRate = fmt.Sprintf("%.2f", float64(stat.NewPlayers)/float64(stat.FLoginedNewUsers))
		}
		if stat.FLoginedOldUsers > 0 {
			stat.FOldPlayerRate = fmt.Sprintf("%.2f", float64(stat.OldPlayers)/float64(stat.FLoginedOldUsers))
		}

		if stat.Players > 0 {
			stat.FPlayerRoundsAvg = fmt.Sprintf("%.2f", float64(stat.BetRounds)/float64(stat.Players))
		}
		// 局数中位数
		if len(stat.AllRoundsList) > 0 {
			var rounds = stat.AllRoundsList
			sort.Slice(rounds, func(i, j int) bool {
				return rounds[i] < rounds[j]
			})
			length := len(rounds)
			if length > 0 {
				if length%2 == 1 {
					stat.FPlayerRoundsMedian = rounds[length/2]
				} else {
					// 中间两位算平均
					stat.FPlayerRoundsMedian = (rounds[length/2] + rounds[length/2-1]) / 2
				}
			}
		}
		stat.FGameTimes = fmt.Sprintf("%.2f", float64(stat.GameTimes)/60)
		if stat.BetRounds > 0 {
			stat.FGameTimesAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.BetRounds)/60)
		}
		if stat.Players > 0 {
			stat.FGameTimesPlayerAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.Players)/60)
		}
		stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
		if stat.Players > 0 {
			stat.FPlayerBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.Players)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.BetRounds)/100.0)
			// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
		}
		if stat.Bets > 0 {
			stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
		}
		stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
		stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
		if stat.WinRounds > 0 {
			stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
		}
		stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
		stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
		stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)
		if stat.WinRounds > 0 {
			stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FMulpitleAvg = fmt.Sprintf("%.2f", stat.MulpitleSum/float64(stat.BetRounds))
		}
		if len(stat.Mulpitles) > 0 {
			var mulpitles []float64
			for _, m := range stat.Mulpitles {
				mulpitles = append(mulpitles, m...)
			}
			sort.Slice(mulpitles, func(i, j int) bool {
				return mulpitles[i] < mulpitles[j]
			})
			length := len(mulpitles)
			if length > 0 {
				// 赢局最高逃脱倍数
				stat.FWinEscapeMulpitleMax = fmt.Sprintf("%.2f", mulpitles[len(mulpitles)-1])
				if length%2 == 1 {
					stat.FMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
				} else {
					// 中间两位算平均
					stat.FMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
				}
			}
		}

		// 赢局局均逃脱倍数
		var winEscapeMulpitleAvg float64
		if stat.WinRounds > 0 {
			winEscapeMulpitleAvg = stat.WinEscapeMulpitleSum / float64(stat.WinRounds)
			stat.FWinEscapeMulpitleAvg = fmt.Sprintf("%.2f", winEscapeMulpitleAvg)
		}
		if len(stat.WinEscapeMulpitles) > 0 {
			var mulpitles []float64
			for _, m := range stat.WinEscapeMulpitles {
				mulpitles = append(mulpitles, m...)
			}
			sort.Slice(mulpitles, func(i, j int) bool {
				return mulpitles[i] < mulpitles[j]
			})
			length := len(mulpitles)
			if length > 0 {
				if length%2 == 1 {
					stat.FWinEscapeMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
				} else {
					// 中间两位算平均
					stat.FWinEscapeMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
				}
			}
		}
		// 赢局局均实际爆炸倍数
		var winMulpitleAvg float64
		if stat.WinRounds > 0 {
			winMulpitleAvg = stat.WinMulpitleSum / float64(stat.WinRounds)
			stat.FWinMulpitleAvg = fmt.Sprintf("%.2f", winMulpitleAvg)
		}
		if len(stat.WinMulpitles) > 0 {
			var mulpitles []float64
			for _, m := range stat.WinMulpitles {
				mulpitles = append(mulpitles, m...)
			}
			sort.Slice(mulpitles, func(i, j int) bool {
				return mulpitles[i] < mulpitles[j]
			})
			length := len(mulpitles)
			if length > 0 {
				if length%2 == 1 {
					stat.FWinMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
				} else {
					// 中间两位算平均
					stat.FWinMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
				}
			}
		}
		// 倍差空间
		stat.FMulpitleDiff = fmt.Sprintf("%.2f", winMulpitleAvg-winEscapeMulpitleAvg)

		// 输局局均实际爆炸倍数
		var loseMulpitleAvg float64
		if stat.LoseRounds > 0 {
			loseMulpitleAvg = stat.LoseMulpitleSum / float64(stat.LoseRounds)
			stat.FLoseMulpitleAvg = fmt.Sprintf("%.2f", loseMulpitleAvg)
		}
		if len(stat.LoseMulpitles) > 0 {
			var mulpitles []float64
			for _, m := range stat.LoseMulpitles {
				mulpitles = append(mulpitles, m...)
			}
			sort.Slice(mulpitles, func(i, j int) bool {
				return mulpitles[i] < mulpitles[j]
			})
			length := len(mulpitles)
			if length > 0 {
				if length%2 == 1 {
					stat.FLoseMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
				} else {
					// 中间两位算平均
					stat.FLoseMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
				}
			}
		}
		if loseMulpitleAvg > 0 {
			stat.FWinEscapeRate = fmt.Sprintf("%.2f", winEscapeMulpitleAvg/loseMulpitleAvg)
			stat.FWinLoseRate = fmt.Sprintf("%.2f", winMulpitleAvg/loseMulpitleAvg)
		}
		if stat.BetRounds > 0 {
			stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
		}
		if stat.BetRounds > 0 {
			stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
		}
	}
}

func tpStatDatesSummaryMapping(stat *entity.TPPlayerStat) {
	// 玩 crash 人数占日活比
	if stat.FLoginedUsers > 0 {
		stat.FPlayerRate = fmt.Sprintf("%.2f", float64(stat.Players)/float64(stat.FLoginedUsers))
	}

	// FNewPlayerRate, FOldPlayerRate 占新老玩家日活
	if stat.FLoginedNewUsers > 0 {
		stat.FNewPlayerRate = fmt.Sprintf("%.2f", float64(stat.NewPlayers)/float64(stat.FLoginedNewUsers))
	}
	if stat.FLoginedOldUsers > 0 {
		stat.FOldPlayerRate = fmt.Sprintf("%.2f", float64(stat.OldPlayers)/float64(stat.FLoginedOldUsers))
	}

	if stat.Players > 0 {
		stat.FPlayerRoundsAvg = fmt.Sprintf("%.2f", float64(stat.BetRounds)/float64(stat.Players))
	}
	// 局数中位数
	if len(stat.AllRoundsList) > 0 {
		var rounds = stat.AllRoundsList
		sort.Slice(rounds, func(i, j int) bool {
			return rounds[i] < rounds[j]
		})
		length := len(rounds)
		if length > 0 {
			if length%2 == 1 {
				stat.FPlayerRoundsMedian = rounds[length/2]
			} else {
				// 中间两位算平均
				stat.FPlayerRoundsMedian = (rounds[length/2] + rounds[length/2-1]) / 2
			}
		}
	}
	stat.FGameTimes = fmt.Sprintf("%.2f", float64(stat.GameTimes)/60)
	if stat.BetRounds > 0 {
		stat.FGameTimesAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.BetRounds)/60)
	}
	if stat.Players > 0 {
		stat.FGameTimesPlayerAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.Players)/60)
	}
	stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
	if stat.Players > 0 {
		stat.FPlayerBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.Players)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.BetRounds)/100.0)
		// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
	}
	if stat.Bets > 0 {
		stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
	}
	stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
	stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
	if stat.WinRounds > 0 {
		stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
	}
	stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
	stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
	stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)
	if stat.WinRounds > 0 {
		stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FMulpitleAvg = fmt.Sprintf("%.2f", stat.MulpitleSum/float64(stat.BetRounds))
	}
	if len(stat.Mulpitles) > 0 {
		var mulpitles []float64
		for _, m := range stat.Mulpitles {
			mulpitles = append(mulpitles, m...)
		}
		sort.Slice(mulpitles, func(i, j int) bool {
			return mulpitles[i] < mulpitles[j]
		})
		length := len(mulpitles)
		if length > 0 {
			// 赢局最高逃脱倍数
			stat.FWinEscapeMulpitleMax = fmt.Sprintf("%.2f", mulpitles[len(mulpitles)-1])
			if length%2 == 1 {
				stat.FMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
			} else {
				// 中间两位算平均
				stat.FMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
			}
		}
	}
	// 赢局局均逃脱倍数
	var winEscapeMulpitleAvg float64
	if stat.WinRounds > 0 {
		winEscapeMulpitleAvg = stat.WinEscapeMulpitleSum / float64(stat.WinRounds)
		stat.FWinEscapeMulpitleAvg = fmt.Sprintf("%.2f", winEscapeMulpitleAvg)
	}
	if len(stat.WinEscapeMulpitles) > 0 {
		var mulpitles []float64
		for _, m := range stat.WinEscapeMulpitles {
			mulpitles = append(mulpitles, m...)
		}
		sort.Slice(mulpitles, func(i, j int) bool {
			return mulpitles[i] < mulpitles[j]
		})
		length := len(mulpitles)
		if length > 0 {
			if length%2 == 1 {
				stat.FWinEscapeMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
			} else {
				// 中间两位算平均
				stat.FWinEscapeMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
			}
		}
	}
	// 赢局局均实际爆炸倍数
	var winMulpitleAvg float64
	if stat.WinRounds > 0 {
		winMulpitleAvg = stat.WinMulpitleSum / float64(stat.WinRounds)
		stat.FWinMulpitleAvg = fmt.Sprintf("%.2f", winMulpitleAvg)
	}
	if len(stat.WinMulpitles) > 0 {
		var mulpitles []float64
		for _, m := range stat.WinMulpitles {
			mulpitles = append(mulpitles, m...)
		}
		sort.Slice(mulpitles, func(i, j int) bool {
			return mulpitles[i] < mulpitles[j]
		})
		length := len(mulpitles)
		if length > 0 {
			if length%2 == 1 {
				stat.FWinMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
			} else {
				// 中间两位算平均
				stat.FWinMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
			}
		}
	}
	// 倍差空间
	stat.FMulpitleDiff = fmt.Sprintf("%.2f", winMulpitleAvg-winEscapeMulpitleAvg)

	// 输局局均实际爆炸倍数
	var loseMulpitleAvg float64
	if stat.LoseRounds > 0 {
		loseMulpitleAvg = stat.LoseMulpitleSum / float64(stat.LoseRounds)
		stat.FLoseMulpitleAvg = fmt.Sprintf("%.2f", loseMulpitleAvg)
	}
	if len(stat.LoseMulpitles) > 0 {
		var mulpitles []float64
		for _, m := range stat.LoseMulpitles {
			mulpitles = append(mulpitles, m...)
		}
		sort.Slice(mulpitles, func(i, j int) bool {
			return mulpitles[i] < mulpitles[j]
		})
		length := len(mulpitles)
		if length > 0 {
			if length%2 == 1 {
				stat.FLoseMulpitleMedian = fmt.Sprintf("%.2f", mulpitles[length/2])
			} else {
				// 中间两位算平均
				stat.FLoseMulpitleMedian = fmt.Sprintf("%.2f", (mulpitles[length/2]+mulpitles[length/2-1])/2)
			}
		}
	}
	if loseMulpitleAvg > 0 {
		stat.FWinEscapeRate = fmt.Sprintf("%.2f", winEscapeMulpitleAvg/loseMulpitleAvg)
		stat.FWinLoseRate = fmt.Sprintf("%.2f", winMulpitleAvg/loseMulpitleAvg)
	}
	if stat.BetRounds > 0 {
		stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
	}
	if stat.BetRounds > 0 {
		stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
	}
}

func (s *gameStatsService) GetTpPlayerRoomStat(m, m2, m3 bson.M, startDate, endDate string) (stats *entity.TPPlayerRoomShow, err error) {
	// 用户列表筛选
	// var filterPlayerIds []string
	stats = new(entity.TPPlayerRoomShow)
	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":                        "$roomid",
			"all_rounds":                 bson.M{"$sum": "$all_rounds"},
			"bets":                       bson.M{"$sum": "$bets"},
			"win_rounds":                 bson.M{"$sum": "$win_rounds"},
			"lose_rounds":                bson.M{"$sum": "$lose_rounds"},
			"win_bets":                   bson.M{"$sum": "$win_bets"},
			"lose_bets":                  bson.M{"$sum": "$lose_bets"},
			"up_number":                  bson.M{"$sum": "$up_number"},
			"down_number":                bson.M{"$sum": "$down_number"},
			"two_rounds":                 bson.M{"$sum": "$two_rounds"},
			"win_two_rounds":             bson.M{"$sum": "$win_two_rounds"},
			"two_bets":                   bson.M{"$sum": "$two_bets"},
			"win_two_bets":               bson.M{"$sum": "$win_two_bets"},
			"three_rounds":               bson.M{"$sum": "$three_rounds"},
			"win_three_rounds":           bson.M{"$sum": "$win_three_rounds"},
			"three_bets":                 bson.M{"$sum": "$three_bets"},
			"win_three_bets":             bson.M{"$sum": "$win_three_bets"},
			"four_rounds":                bson.M{"$sum": "$four_rounds"},
			"win_four_rounds":            bson.M{"$sum": "$win_four_rounds"},
			"four_bets":                  bson.M{"$sum": "$four_bets"},
			"win_four_bets":              bson.M{"$sum": "$win_four_bets"},
			"five_rounds":                bson.M{"$sum": "$five_rounds"},
			"win_five_rounds":            bson.M{"$sum": "$win_five_rounds"},
			"five_bets":                  bson.M{"$sum": "$five_bets"},
			"win_five_bets":              bson.M{"$sum": "$win_five_bets"},
			"chip_pool":                  bson.M{"$first": "$chip_pool"},
			"bz_up_number":               bson.M{"$sum": "$bz_up_number"},
			"bz_up_rounds":               bson.M{"$sum": "$bz_up_rounds"},
			"bz_up_bets":                 bson.M{"$sum": "$bz_up_bets"},
			"bz_up_win_rounds":           bson.M{"$sum": "$bz_up_win_rounds"},
			"bz_up_lose_rounds":          bson.M{"$sum": "$bz_up_lose_rounds"},
			"bz_up_two_discard_num":      bson.M{"$sum": "$bz_up_two_discard_num"},
			"bz_up_two_follow_num":       bson.M{"$sum": "$bz_up_two_follow_num"},
			"bz_up_three_discard_num":    bson.M{"$sum": "$bz_up_three_discard_num"},
			"bz_up_three_follow_num":     bson.M{"$sum": "$bz_up_three_follow_num"},
			"bz_up_four_discard_num":     bson.M{"$sum": "$bz_up_four_discard_num"},
			"bz_up_four_follow_num":      bson.M{"$sum": "$bz_up_four_follow_num"},
			"bz_up_five_discard_num":     bson.M{"$sum": "$bz_up_five_discard_num"},
			"bz_up_five_follow_num":      bson.M{"$sum": "$bz_up_five_follow_num"},
			"bz_down_number":             bson.M{"$sum": "$bz_down_number"},
			"bz_down_rounds":             bson.M{"$sum": "$bz_down_rounds"},
			"bz_down_bets":               bson.M{"$sum": "$bz_down_bets"},
			"bz_down_win_rounds":         bson.M{"$sum": "$bz_down_win_rounds"},
			"bz_down_lose_rounds":        bson.M{"$sum": "$bz_down_lose_rounds"},
			"bz_down_two_discard_num":    bson.M{"$sum": "$bz_down_two_discard_num"},
			"bz_down_two_follow_num":     bson.M{"$sum": "$bz_down_two_follow_num"},
			"bz_down_three_discard_num":  bson.M{"$sum": "$bz_down_three_discard_num"},
			"bz_down_three_follow_num":   bson.M{"$sum": "$bz_down_three_follow_num"},
			"bz_down_four_discard_num":   bson.M{"$sum": "$bz_down_four_discard_num"},
			"bz_down_four_follow_num":    bson.M{"$sum": "$bz_down_four_follow_num"},
			"bz_down_five_discard_num":   bson.M{"$sum": "$bz_down_five_discard_num"},
			"bz_down_five_follow_num":    bson.M{"$sum": "$bz_down_five_follow_num"},
			"ths_up_number":              bson.M{"$sum": "$ths_up_number"},
			"ths_up_rounds":              bson.M{"$sum": "$ths_up_rounds"},
			"ths_up_bets":                bson.M{"$sum": "$ths_up_bets"},
			"ths_up_win_rounds":          bson.M{"$sum": "$ths_up_win_rounds"},
			"ths_up_lose_rounds":         bson.M{"$sum": "$ths_up_lose_rounds"},
			"ths_up_two_discard_num":     bson.M{"$sum": "$ths_up_two_discard_num"},
			"ths_up_two_follow_num":      bson.M{"$sum": "$ths_up_two_follow_num"},
			"ths_up_three_discard_num":   bson.M{"$sum": "$ths_up_three_discard_num"},
			"ths_up_three_follow_num":    bson.M{"$sum": "$ths_up_three_follow_num"},
			"ths_up_four_discard_num":    bson.M{"$sum": "$ths_up_four_discard_num"},
			"ths_up_four_follow_num":     bson.M{"$sum": "$ths_up_four_follow_num"},
			"ths_up_five_discard_num":    bson.M{"$sum": "$ths_up_five_discard_num"},
			"ths_up_five_follow_num":     bson.M{"$sum": "$ths_up_five_follow_num"},
			"ths_down_number":            bson.M{"$sum": "$ths_down_number"},
			"ths_down_rounds":            bson.M{"$sum": "$ths_down_rounds"},
			"ths_down_bets":              bson.M{"$sum": "$ths_down_bets"},
			"ths_down_win_rounds":        bson.M{"$sum": "$ths_down_win_rounds"},
			"ths_down_lose_rounds":       bson.M{"$sum": "$ths_down_lose_rounds"},
			"ths_down_two_discard_num":   bson.M{"$sum": "$ths_down_two_discard_num"},
			"ths_down_two_follow_num":    bson.M{"$sum": "$ths_down_two_follow_num"},
			"ths_down_three_discard_num": bson.M{"$sum": "$ths_down_three_discard_num"},
			"ths_down_three_follow_num":  bson.M{"$sum": "$ths_down_three_follow_num"},
			"ths_down_four_discard_num":  bson.M{"$sum": "$ths_down_four_discard_num"},
			"ths_down_four_follow_num":   bson.M{"$sum": "$ths_down_four_follow_num"},
			"ths_down_five_discard_num":  bson.M{"$sum": "$ths_down_five_discard_num"},
			"ths_down_five_follow_num":   bson.M{"$sum": "$ths_down_five_follow_num"},
			"dsz_up_number":              bson.M{"$sum": "$dsz_up_number"},
			"dsz_up_rounds":              bson.M{"$sum": "$dsz_up_rounds"},
			"dsz_up_bets":                bson.M{"$sum": "$dsz_up_bets"},
			"dsz_up_win_rounds":          bson.M{"$sum": "$dsz_up_win_rounds"},
			"dsz_up_lose_rounds":         bson.M{"$sum": "$dsz_up_lose_rounds"},
			"dsz_up_two_discard_num":     bson.M{"$sum": "$dsz_up_two_discard_num"},
			"dsz_up_two_follow_num":      bson.M{"$sum": "$dsz_up_two_follow_num"},
			"dsz_up_three_discard_num":   bson.M{"$sum": "$dsz_up_three_discard_num"},
			"dsz_up_three_follow_num":    bson.M{"$sum": "$dsz_up_three_follow_num"},
			"dsz_up_four_discard_num":    bson.M{"$sum": "$dsz_up_four_discard_num"},
			"dsz_up_four_follow_num":     bson.M{"$sum": "$dsz_up_four_follow_num"},
			"dsz_up_five_discard_num":    bson.M{"$sum": "$dsz_up_five_discard_num"},
			"dsz_up_five_follow_num":     bson.M{"$sum": "$dsz_up_five_follow_num"},
			"dsz_down_number":            bson.M{"$sum": "$dsz_down_number"},
			"dsz_down_rounds":            bson.M{"$sum": "$dsz_down_rounds"},
			"dsz_down_bets":              bson.M{"$sum": "$dsz_down_bets"},
			"dsz_down_win_rounds":        bson.M{"$sum": "$dsz_down_win_rounds"},
			"dsz_down_lose_rounds":       bson.M{"$sum": "$dsz_down_lose_rounds"},
			"dsz_down_two_discard_num":   bson.M{"$sum": "$dsz_down_two_discard_num"},
			"dsz_down_two_follow_num":    bson.M{"$sum": "$dsz_down_two_follow_num"},
			"dsz_down_three_discard_num": bson.M{"$sum": "$dsz_down_three_discard_num"},
			"dsz_down_three_follow_num":  bson.M{"$sum": "$dsz_down_three_follow_num"},
			"dsz_down_four_discard_num":  bson.M{"$sum": "$dsz_down_four_discard_num"},
			"dsz_down_four_follow_num":   bson.M{"$sum": "$dsz_down_four_follow_num"},
			"dsz_down_five_discard_num":  bson.M{"$sum": "$dsz_down_five_discard_num"},
			"dsz_down_five_follow_num":   bson.M{"$sum": "$dsz_down_five_follow_num"},
			"dth_up_number":              bson.M{"$sum": "$dth_up_number"},
			"dth_up_rounds":              bson.M{"$sum": "$dth_up_rounds"},
			"dth_up_bets":                bson.M{"$sum": "$dth_up_bets"},
			"dth_up_win_rounds":          bson.M{"$sum": "$dth_up_win_rounds"},
			"dth_up_lose_rounds":         bson.M{"$sum": "$dth_up_lose_rounds"},
			"dth_up_two_discard_num":     bson.M{"$sum": "$dth_up_two_discard_num"},
			"dth_up_two_follow_num":      bson.M{"$sum": "$dth_up_two_follow_num"},
			"dth_up_three_discard_num":   bson.M{"$sum": "$dth_up_three_discard_num"},
			"dth_up_three_follow_num":    bson.M{"$sum": "$dth_up_three_follow_num"},
			"dth_up_four_discard_num":    bson.M{"$sum": "$dth_up_four_discard_num"},
			"dth_up_four_follow_num":     bson.M{"$sum": "$dth_up_four_follow_num"},
			"dth_up_five_discard_num":    bson.M{"$sum": "$dth_up_five_discard_num"},
			"dth_up_five_follow_num":     bson.M{"$sum": "$dth_up_five_follow_num"},
			"dth_down_number":            bson.M{"$sum": "$dth_down_number"},
			"dth_down_rounds":            bson.M{"$sum": "$dth_down_rounds"},
			"dth_down_bets":              bson.M{"$sum": "$dth_down_bets"},
			"dth_down_win_rounds":        bson.M{"$sum": "$dth_down_win_rounds"},
			"dth_down_lose_rounds":       bson.M{"$sum": "$dth_down_lose_rounds"},
			"dth_down_two_discard_num":   bson.M{"$sum": "$dth_down_two_discard_num"},
			"dth_down_two_follow_num":    bson.M{"$sum": "$dth_down_two_follow_num"},
			"dth_down_three_discard_num": bson.M{"$sum": "$dth_down_three_discard_num"},
			"dth_down_three_follow_num":  bson.M{"$sum": "$dth_down_three_follow_num"},
			"dth_down_four_discard_num":  bson.M{"$sum": "$dth_down_four_discard_num"},
			"dth_down_four_follow_num":   bson.M{"$sum": "$dth_down_four_follow_num"},
			"dth_down_five_discard_num":  bson.M{"$sum": "$dth_down_five_discard_num"},
			"dth_down_five_follow_num":   bson.M{"$sum": "$dth_down_five_follow_num"},
			"ddz_up_number":              bson.M{"$sum": "$ddz_up_number"},
			"ddz_up_rounds":              bson.M{"$sum": "$ddz_up_rounds"},
			"ddz_up_bets":                bson.M{"$sum": "$ddz_up_bets"},
			"ddz_up_win_rounds":          bson.M{"$sum": "$ddz_up_win_rounds"},
			"ddz_up_lose_rounds":         bson.M{"$sum": "$ddz_up_lose_rounds"},
			"ddz_up_two_discard_num":     bson.M{"$sum": "$ddz_up_two_discard_num"},
			"ddz_up_two_follow_num":      bson.M{"$sum": "$ddz_up_two_follow_num"},
			"ddz_up_three_discard_num":   bson.M{"$sum": "$ddz_up_three_discard_num"},
			"ddz_up_three_follow_num":    bson.M{"$sum": "$ddz_up_three_follow_num"},
			"ddz_up_four_discard_num":    bson.M{"$sum": "$ddz_up_four_discard_num"},
			"ddz_up_four_follow_num":     bson.M{"$sum": "$ddz_up_four_follow_num"},
			"ddz_up_five_discard_num":    bson.M{"$sum": "$ddz_up_five_discard_num"},
			"ddz_up_five_follow_num":     bson.M{"$sum": "$ddz_up_five_follow_num"},
			"ddz_down_number":            bson.M{"$sum": "$ddz_down_number"},
			"ddz_down_rounds":            bson.M{"$sum": "$ddz_down_rounds"},
			"ddz_down_bets":              bson.M{"$sum": "$ddz_down_bets"},
			"ddz_down_win_rounds":        bson.M{"$sum": "$ddz_down_win_rounds"},
			"ddz_down_lose_rounds":       bson.M{"$sum": "$ddz_down_lose_rounds"},
			"ddz_down_two_discard_num":   bson.M{"$sum": "$ddz_down_two_discard_num"},
			"ddz_down_two_follow_num":    bson.M{"$sum": "$ddz_down_two_follow_num"},
			"ddz_down_three_discard_num": bson.M{"$sum": "$ddz_down_three_discard_num"},
			"ddz_down_three_follow_num":  bson.M{"$sum": "$ddz_down_three_follow_num"},
			"ddz_down_four_discard_num":  bson.M{"$sum": "$ddz_down_four_discard_num"},
			"ddz_down_four_follow_num":   bson.M{"$sum": "$ddz_down_four_follow_num"},
			"ddz_down_five_discard_num":  bson.M{"$sum": "$ddz_down_five_discard_num"},
			"ddz_down_five_follow_num":   bson.M{"$sum": "$ddz_down_five_follow_num"},
			"dgp_up_number":              bson.M{"$sum": "$dgp_up_number"},
			"dgp_up_rounds":              bson.M{"$sum": "$dgp_up_rounds"},
			"dgp_up_bets":                bson.M{"$sum": "$dgp_up_bets"},
			"dgp_up_win_rounds":          bson.M{"$sum": "$dgp_up_win_rounds"},
			"dgp_up_lose_rounds":         bson.M{"$sum": "$dgp_up_lose_rounds"},
			"dgp_up_two_discard_num":     bson.M{"$sum": "$dgp_up_two_discard_num"},
			"dgp_up_two_follow_num":      bson.M{"$sum": "$dgp_up_two_follow_num"},
			"dgp_up_three_discard_num":   bson.M{"$sum": "$dgp_up_three_discard_num"},
			"dgp_up_three_follow_num":    bson.M{"$sum": "$dgp_up_three_follow_num"},
			"dgp_up_four_discard_num":    bson.M{"$sum": "$dgp_up_four_discard_num"},
			"dgp_up_four_follow_num":     bson.M{"$sum": "$dgp_up_four_follow_num"},
			"dgp_up_five_discard_num":    bson.M{"$sum": "$dgp_up_five_discard_num"},
			"dgp_up_five_follow_num":     bson.M{"$sum": "$dgp_up_five_follow_num"},
			"dgp_down_number":            bson.M{"$sum": "$dgp_down_number"},
			"dgp_down_rounds":            bson.M{"$sum": "$dgp_down_rounds"},
			"dgp_down_bets":              bson.M{"$sum": "$dgp_down_bets"},
			"dgp_down_win_rounds":        bson.M{"$sum": "$dgp_down_win_rounds"},
			"dgp_down_lose_rounds":       bson.M{"$sum": "$dgp_down_lose_rounds"},
			"dgp_down_two_discard_num":   bson.M{"$sum": "$dgp_down_two_discard_num"},
			"dgp_down_two_follow_num":    bson.M{"$sum": "$dgp_down_two_follow_num"},
			"dgp_down_three_discard_num": bson.M{"$sum": "$dgp_down_three_discard_num"},
			"dgp_down_three_follow_num":  bson.M{"$sum": "$dgp_down_three_follow_num"},
			"dgp_down_four_discard_num":  bson.M{"$sum": "$dgp_down_four_discard_num"},
			"dgp_down_four_follow_num":   bson.M{"$sum": "$dgp_down_four_follow_num"},
			"dgp_down_five_discard_num":  bson.M{"$sum": "$dgp_down_five_discard_num"},
			"dgp_down_five_follow_num":   bson.M{"$sum": "$dgp_down_five_follow_num"},
		}},
	}
	// 查总条数
	var res []entity.TPPlayerRoomStat
	err = TPRoomPlayerStats.Pipe(m1).All(&res)
	if err != nil {
		return
	}
	roomlist := make([]*entity.TPRoomStat, 0)
	turnlist := new(entity.TPGameTurn)
	winlist := new(entity.TPGameWin)
	loselist := new(entity.TPGameLose)
	for _, item := range res {
		new_info := new(entity.TPRoomStat)
		new_info.Roomid = item.Id
		new_info.AllRounds = item.AllRounds
		new_info.Bets = item.Bets
		new_info.FBets = Chip2Float(item.Bets)
		new_info.RoundBetAvg = ComputeFloat(item.Bets, item.AllRounds) / 100.0
		new_info.TotalProfit = Chip2Float(item.Bets - item.WinBets)
		new_info.UpNumber = item.UpNumber
		new_info.DownNumber = item.DownNumber
		new_info.TwoWinRate = ComputeFloat(item.WinTwoRounds, item.TwoRounds) * 100.0
		new_info.ThreeWinRate = ComputeFloat(item.WinThreeRounds, item.ThreeRounds) * 100.0
		new_info.FourWinRate = ComputeFloat(item.WinFourRounds, item.FourRounds) * 100.0
		new_info.FiveWinRate = ComputeFloat(item.WinFiveRounds, item.FiveRounds) * 100.0
		new_info.TwoBonusRate = ComputeFloat(item.WinTwoBets, item.TwoBets) * 100
		new_info.ThreeBonusRate = ComputeFloat(item.WinThreeBets, item.ThreeBets) * 100
		new_info.FourBonusRate = ComputeFloat(item.WinFourBets, item.FourBets) * 100
		new_info.FiveBonusRate = ComputeFloat(item.WinFiveBets, item.FiveBets) * 100
		new_info.ChipPool = Chip2Float(item.ChipPool)
		new_info.BZUpBetAvg = ComputeFloat(item.BZUpBets, item.BZUpNumber) / 100.0
		new_info.BZDownBetAvg = ComputeFloat(item.BZDownBets, item.BZDownNumber) / 100.0
		new_info.THSUpBetAvg = ComputeFloat(item.THSUpBets, item.THSUpNumber) / 100.0
		new_info.THSDownBetAvg = ComputeFloat(item.THSDownBets, item.THSDownNumber) / 100.0
		new_info.DSZUpBetAvg = ComputeFloat(item.DSZUpBets, item.DSZUpNumber) / 100.0
		new_info.DSZDownBetAvg = ComputeFloat(item.DSZDownBets, item.DSZDownNumber) / 100.0
		new_info.DTHUpBetAvg = ComputeFloat(item.DTHUpBets, item.DTHUpNumber) / 100.0
		new_info.DDZUpBetAvg = ComputeFloat(item.DDZUpBets, item.DDZUpNumber) / 100.0
		new_info.DDZDownBetAvg = ComputeFloat(item.DDZDownBets, item.DDZDownNumber) / 100.0
		new_info.DGPUpBetAvg = ComputeFloat(item.DGPUpBets, item.DGPUpNumber) / 100.0
		if item.DGPDownNumber != 0 {
			new_info.DGPDownBetAvg = Chip2Float(item.DGPDownBets) / float64(item.DGPDownNumber) // ComputeFloat(item.DGPDownBets, item.DGPDownNumber) / 100.0
		}
		roomlist = append(roomlist, new_info)
		// 期望轮次
		turnlist.BZUpNumber += item.BZUpNumber
		turnlist.BZUpRounds += item.BZUpRounds
		turnlist.BZDownNumber += item.BZDownNumber
		turnlist.BZDownRounds += item.BZDownRounds
		turnlist.THSUpNumber += item.THSUpNumber
		turnlist.THSUpRounds += item.THSUpRounds
		turnlist.THSDownNumber += item.THSDownNumber
		turnlist.THSDownRounds += item.THSDownRounds
		turnlist.DSZUpNumber += item.DSZUpNumber
		turnlist.DSZUpRounds += item.DSZUpRounds
		turnlist.DSZDownNumber += item.DSZDownNumber
		turnlist.DSZDownRounds += item.DSZDownRounds
		turnlist.DTHUpNumber += item.DTHUpNumber
		turnlist.DTHUpRounds += item.DTHUpRounds
		turnlist.DTHDownNumber += item.DTHDownNumber
		turnlist.DTHDownRounds += item.DTHDownRounds
		turnlist.DDZUpNumber += item.DDZUpNumber
		turnlist.DDZUpRounds += item.DDZUpRounds
		turnlist.DDZDownNumber += item.DDZDownNumber
		turnlist.DDZDownRounds += item.DDZDownRounds
		turnlist.DGPUpNumber += item.DGPUpNumber
		turnlist.DGPUpRounds += item.DGPUpRounds
		turnlist.DGPDownNumber += item.DGPDownNumber
		turnlist.DGPDownRounds += item.DGPDownRounds

		// 大赢局
		winlist.BZUpWinRounds += item.BZUpWinRounds
		winlist.BZDownWinRounds += item.BZDownWinRounds
		winlist.THSUpWinRounds += item.THSUpWinRounds
		winlist.THSDownWinRounds += item.THSDownWinRounds
		winlist.DSZUpWinRounds += item.DSZUpWinRounds
		winlist.DSZDownWinRounds += item.DSZDownWinRounds
		winlist.DTHUpWinRounds += item.DTHUpWinRounds
		winlist.DTHDownWinRounds += item.DTHDownWinRounds
		winlist.DDZUpWinRounds += item.DDZUpWinRounds
		winlist.DDZDownWinRounds += item.DDZDownWinRounds
		winlist.DGPUpWinRounds += item.DGPUpWinRounds
		winlist.DGPDownWinRounds += item.DGPDownWinRounds

		// 大输家
		loselist.BZUpLoseRounds += item.BZUpLoseRounds
		loselist.BZDownLoseRounds += item.BZUpLoseRounds
		loselist.THSUpLoseRounds += item.THSUpLoseRounds
		loselist.THSDownLoseRounds += item.THSDownLoseRounds
		loselist.DSZUpLoseRounds += item.DSZUpLoseRounds
		loselist.DSZDownLoseRounds += item.DSZDownLoseRounds
		loselist.DTHUpLoseRounds += item.DTHUpLoseRounds
		loselist.DTHDownLoseRounds += item.DTHDownLoseRounds
		loselist.DDZUpLoseRounds += item.DDZUpLoseRounds
		loselist.DDZDownLoseRounds += item.DDZDownLoseRounds
		loselist.DGPUpLoseRounds += item.DGPUpLoseRounds
		loselist.DGPDownLoseRounds += item.DGPDownLoseRounds

	}
	if turnlist.BZUpRounds != 0 {
		turnlist.BZUpRoundAvg = ComputeFloat(turnlist.BZUpRounds, turnlist.BZUpNumber)
	}
	if turnlist.BZDownRounds != 0 {
		turnlist.BZDownRoundAvg = ComputeFloat(turnlist.BZDownRounds, turnlist.BZDownNumber)
	}

	if turnlist.THSUpRounds != 0 {
		turnlist.THSUpRoundAvg = ComputeFloat(turnlist.THSUpRounds, turnlist.THSUpNumber)
	}
	if turnlist.THSDownRounds != 0 {
		turnlist.THSDownRoundAvg = ComputeFloat(turnlist.THSDownRounds, turnlist.THSDownNumber)
	}

	if turnlist.DSZUpRounds != 0 {
		turnlist.DSZUpRoundAvg = ComputeFloat(turnlist.DSZUpRounds, turnlist.DSZUpNumber)
	}
	if turnlist.DSZDownRounds != 0 {
		turnlist.DSZDownRoundAvg = ComputeFloat(turnlist.DSZDownRounds, turnlist.DSZDownNumber)
	}

	if turnlist.DTHUpRounds != 0 {
		turnlist.DTHUpRoundAvg = ComputeFloat(turnlist.DTHUpRounds, turnlist.DTHUpNumber)
	}
	if turnlist.DTHDownRounds != 0 {
		turnlist.DTHDownRoundAvg = ComputeFloat(turnlist.DTHDownRounds, turnlist.DTHDownNumber)
	}

	if turnlist.DDZUpRounds != 0 {
		turnlist.DDZUpRoundAvg = ComputeFloat(turnlist.DDZUpRounds, turnlist.DDZUpNumber)
	}
	if turnlist.DDZDownRounds != 0 {
		turnlist.DDZDownRoundAvg = ComputeFloat(turnlist.DDZDownRounds, turnlist.DDZDownNumber)
	}

	if turnlist.DGPUpRounds != 0 {
		turnlist.DGPUpRoundAvg = ComputeFloat(turnlist.DGPUpRounds, turnlist.DGPUpNumber)
	}
	if turnlist.DGPDownRounds != 0 {
		turnlist.DGPDownRoundAvg = ComputeFloat(turnlist.DDZDownRounds, turnlist.DGPDownNumber)
	}
	stats.RoomInfo = roomlist
	stats.TurnInfo = append(stats.TurnInfo, turnlist)
	stats.WinInfo = append(stats.WinInfo, winlist)
	stats.LoseInfo = append(stats.LoseInfo, loselist)
	return
}

func (s *gameStatsService) GetTpPlayerStat(userid string) (stats *entity.TPPlayerRoomShow, line *entity.TPProfitShow, err error) {
	// 用户列表筛选
	// var filterPlayerIds []string
	stats = new(entity.TPPlayerRoomShow)
	line = new(entity.TPProfitShow)
	m := bson.M{}
	m["userid"] = userid
	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":                        "$userid",
			"all_rounds":                 bson.M{"$sum": "$all_rounds"},
			"bets":                       bson.M{"$sum": "$bets"},
			"win_rounds":                 bson.M{"$sum": "$win_rounds"},
			"lose_rounds":                bson.M{"$sum": "$lose_rounds"},
			"win_bets":                   bson.M{"$sum": "$win_bets"},
			"lose_bets":                  bson.M{"$sum": "$lose_bets"},
			"up_number":                  bson.M{"$sum": "$up_number"},
			"down_number":                bson.M{"$sum": "$down_number"},
			"two_rounds":                 bson.M{"$sum": "$two_rounds"},
			"win_two_rounds":             bson.M{"$sum": "$win_two_rounds"},
			"two_bets":                   bson.M{"$sum": "$two_bets"},
			"win_two_bets":               bson.M{"$sum": "$win_two_bets"},
			"three_rounds":               bson.M{"$sum": "$three_rounds"},
			"win_three_rounds":           bson.M{"$sum": "$win_three_rounds"},
			"three_bets":                 bson.M{"$sum": "$three_bets"},
			"win_three_bets":             bson.M{"$sum": "$win_three_bets"},
			"four_rounds":                bson.M{"$sum": "$four_rounds"},
			"win_four_rounds":            bson.M{"$sum": "$win_four_rounds"},
			"four_bets":                  bson.M{"$sum": "$four_bets"},
			"win_four_bets":              bson.M{"$sum": "$win_four_bets"},
			"five_rounds":                bson.M{"$sum": "$five_rounds"},
			"win_five_rounds":            bson.M{"$sum": "$win_five_rounds"},
			"five_bets":                  bson.M{"$sum": "$five_bets"},
			"win_five_bets":              bson.M{"$sum": "$win_five_bets"},
			"chip_pool":                  bson.M{"$first": "$chip_pool"},
			"bz_up_number":               bson.M{"$sum": "$bz_up_number"},
			"bz_up_rounds":               bson.M{"$sum": "$bz_up_rounds"},
			"bz_up_bets":                 bson.M{"$sum": "$bz_up_bets"},
			"bz_up_win_rounds":           bson.M{"$sum": "$bz_up_win_rounds"},
			"bz_up_lose_rounds":          bson.M{"$sum": "$bz_up_lose_rounds"},
			"bz_up_two_discard_num":      bson.M{"$sum": "$bz_up_two_discard_num"},
			"bz_up_two_follow_num":       bson.M{"$sum": "$bz_up_two_follow_num"},
			"bz_up_three_discard_num":    bson.M{"$sum": "$bz_up_three_discard_num"},
			"bz_up_three_follow_num":     bson.M{"$sum": "$bz_up_three_follow_num"},
			"bz_up_four_discard_num":     bson.M{"$sum": "$bz_up_four_discard_num"},
			"bz_up_four_follow_num":      bson.M{"$sum": "$bz_up_four_follow_num"},
			"bz_up_five_discard_num":     bson.M{"$sum": "$bz_up_five_discard_num"},
			"bz_up_five_follow_num":      bson.M{"$sum": "$bz_up_five_follow_num"},
			"bz_down_number":             bson.M{"$sum": "$bz_down_number"},
			"bz_down_rounds":             bson.M{"$sum": "$bz_down_rounds"},
			"bz_down_bets":               bson.M{"$sum": "$bz_down_bets"},
			"bz_down_win_rounds":         bson.M{"$sum": "$bz_down_win_rounds"},
			"bz_down_lose_rounds":        bson.M{"$sum": "$bz_down_lose_rounds"},
			"bz_down_two_discard_num":    bson.M{"$sum": "$bz_down_two_discard_num"},
			"bz_down_two_follow_num":     bson.M{"$sum": "$bz_down_two_follow_num"},
			"bz_down_three_discard_num":  bson.M{"$sum": "$bz_down_three_discard_num"},
			"bz_down_three_follow_num":   bson.M{"$sum": "$bz_down_three_follow_num"},
			"bz_down_four_discard_num":   bson.M{"$sum": "$bz_down_four_discard_num"},
			"bz_down_four_follow_num":    bson.M{"$sum": "$bz_down_four_follow_num"},
			"bz_down_five_discard_num":   bson.M{"$sum": "$bz_down_five_discard_num"},
			"bz_down_five_follow_num":    bson.M{"$sum": "$bz_down_five_follow_num"},
			"ths_up_number":              bson.M{"$sum": "$ths_up_number"},
			"ths_up_rounds":              bson.M{"$sum": "$ths_up_rounds"},
			"ths_up_bets":                bson.M{"$sum": "$ths_up_bets"},
			"ths_up_win_rounds":          bson.M{"$sum": "$ths_up_win_rounds"},
			"ths_up_lose_rounds":         bson.M{"$sum": "$ths_up_lose_rounds"},
			"ths_up_two_discard_num":     bson.M{"$sum": "$ths_up_two_discard_num"},
			"ths_up_two_follow_num":      bson.M{"$sum": "$ths_up_two_follow_num"},
			"ths_up_three_discard_num":   bson.M{"$sum": "$ths_up_three_discard_num"},
			"ths_up_three_follow_num":    bson.M{"$sum": "$ths_up_three_follow_num"},
			"ths_up_four_discard_num":    bson.M{"$sum": "$ths_up_four_discard_num"},
			"ths_up_four_follow_num":     bson.M{"$sum": "$ths_up_four_follow_num"},
			"ths_up_five_discard_num":    bson.M{"$sum": "$ths_up_five_discard_num"},
			"ths_up_five_follow_num":     bson.M{"$sum": "$ths_up_five_follow_num"},
			"ths_down_number":            bson.M{"$sum": "$ths_down_number"},
			"ths_down_rounds":            bson.M{"$sum": "$ths_down_rounds"},
			"ths_down_bets":              bson.M{"$sum": "$ths_down_bets"},
			"ths_down_win_rounds":        bson.M{"$sum": "$ths_down_win_rounds"},
			"ths_down_lose_rounds":       bson.M{"$sum": "$ths_down_lose_rounds"},
			"ths_down_two_discard_num":   bson.M{"$sum": "$ths_down_two_discard_num"},
			"ths_down_two_follow_num":    bson.M{"$sum": "$ths_down_two_follow_num"},
			"ths_down_three_discard_num": bson.M{"$sum": "$ths_down_three_discard_num"},
			"ths_down_three_follow_num":  bson.M{"$sum": "$ths_down_three_follow_num"},
			"ths_down_four_discard_num":  bson.M{"$sum": "$ths_down_four_discard_num"},
			"ths_down_four_follow_num":   bson.M{"$sum": "$ths_down_four_follow_num"},
			"ths_down_five_discard_num":  bson.M{"$sum": "$ths_down_five_discard_num"},
			"ths_down_five_follow_num":   bson.M{"$sum": "$ths_down_five_follow_num"},
			"dsz_up_number":              bson.M{"$sum": "$dsz_up_number"},
			"dsz_up_rounds":              bson.M{"$sum": "$dsz_up_rounds"},
			"dsz_up_bets":                bson.M{"$sum": "$dsz_up_bets"},
			"dsz_up_win_rounds":          bson.M{"$sum": "$dsz_up_win_rounds"},
			"dsz_up_lose_rounds":         bson.M{"$sum": "$dsz_up_lose_rounds"},
			"dsz_up_two_discard_num":     bson.M{"$sum": "$dsz_up_two_discard_num"},
			"dsz_up_two_follow_num":      bson.M{"$sum": "$dsz_up_two_follow_num"},
			"dsz_up_three_discard_num":   bson.M{"$sum": "$dsz_up_three_discard_num"},
			"dsz_up_three_follow_num":    bson.M{"$sum": "$dsz_up_three_follow_num"},
			"dsz_up_four_discard_num":    bson.M{"$sum": "$dsz_up_four_discard_num"},
			"dsz_up_four_follow_num":     bson.M{"$sum": "$dsz_up_four_follow_num"},
			"dsz_up_five_discard_num":    bson.M{"$sum": "$dsz_up_five_discard_num"},
			"dsz_up_five_follow_num":     bson.M{"$sum": "$dsz_up_five_follow_num"},
			"dsz_down_number":            bson.M{"$sum": "$dsz_down_number"},
			"dsz_down_rounds":            bson.M{"$sum": "$dsz_down_rounds"},
			"dsz_down_bets":              bson.M{"$sum": "$dsz_down_bets"},
			"dsz_down_win_rounds":        bson.M{"$sum": "$dsz_down_win_rounds"},
			"dsz_down_lose_rounds":       bson.M{"$sum": "$dsz_down_lose_rounds"},
			"dsz_down_two_discard_num":   bson.M{"$sum": "$dsz_down_two_discard_num"},
			"dsz_down_two_follow_num":    bson.M{"$sum": "$dsz_down_two_follow_num"},
			"dsz_down_three_discard_num": bson.M{"$sum": "$dsz_down_three_discard_num"},
			"dsz_down_three_follow_num":  bson.M{"$sum": "$dsz_down_three_follow_num"},
			"dsz_down_four_discard_num":  bson.M{"$sum": "$dsz_down_four_discard_num"},
			"dsz_down_four_follow_num":   bson.M{"$sum": "$dsz_down_four_follow_num"},
			"dsz_down_five_discard_num":  bson.M{"$sum": "$dsz_down_five_discard_num"},
			"dsz_down_five_follow_num":   bson.M{"$sum": "$dsz_down_five_follow_num"},
			"dth_up_number":              bson.M{"$sum": "$dth_up_number"},
			"dth_up_rounds":              bson.M{"$sum": "$dth_up_rounds"},
			"dth_up_bets":                bson.M{"$sum": "$dth_up_bets"},
			"dth_up_win_rounds":          bson.M{"$sum": "$dth_up_win_rounds"},
			"dth_up_lose_rounds":         bson.M{"$sum": "$dth_up_lose_rounds"},
			"dth_up_two_discard_num":     bson.M{"$sum": "$dth_up_two_discard_num"},
			"dth_up_two_follow_num":      bson.M{"$sum": "$dth_up_two_follow_num"},
			"dth_up_three_discard_num":   bson.M{"$sum": "$dth_up_three_discard_num"},
			"dth_up_three_follow_num":    bson.M{"$sum": "$dth_up_three_follow_num"},
			"dth_up_four_discard_num":    bson.M{"$sum": "$dth_up_four_discard_num"},
			"dth_up_four_follow_num":     bson.M{"$sum": "$dth_up_four_follow_num"},
			"dth_up_five_discard_num":    bson.M{"$sum": "$dth_up_five_discard_num"},
			"dth_up_five_follow_num":     bson.M{"$sum": "$dth_up_five_follow_num"},
			"dth_down_number":            bson.M{"$sum": "$dth_down_number"},
			"dth_down_rounds":            bson.M{"$sum": "$dth_down_rounds"},
			"dth_down_bets":              bson.M{"$sum": "$dth_down_bets"},
			"dth_down_win_rounds":        bson.M{"$sum": "$dth_down_win_rounds"},
			"dth_down_lose_rounds":       bson.M{"$sum": "$dth_down_lose_rounds"},
			"dth_down_two_discard_num":   bson.M{"$sum": "$dth_down_two_discard_num"},
			"dth_down_two_follow_num":    bson.M{"$sum": "$dth_down_two_follow_num"},
			"dth_down_three_discard_num": bson.M{"$sum": "$dth_down_three_discard_num"},
			"dth_down_three_follow_num":  bson.M{"$sum": "$dth_down_three_follow_num"},
			"dth_down_four_discard_num":  bson.M{"$sum": "$dth_down_four_discard_num"},
			"dth_down_four_follow_num":   bson.M{"$sum": "$dth_down_four_follow_num"},
			"dth_down_five_discard_num":  bson.M{"$sum": "$dth_down_five_discard_num"},
			"dth_down_five_follow_num":   bson.M{"$sum": "$dth_down_five_follow_num"},
			"ddz_up_number":              bson.M{"$sum": "$ddz_up_number"},
			"ddz_up_rounds":              bson.M{"$sum": "$ddz_up_rounds"},
			"ddz_up_bets":                bson.M{"$sum": "$ddz_up_bets"},
			"ddz_up_win_rounds":          bson.M{"$sum": "$ddz_up_win_rounds"},
			"ddz_up_lose_rounds":         bson.M{"$sum": "$ddz_up_lose_rounds"},
			"ddz_up_two_discard_num":     bson.M{"$sum": "$ddz_up_two_discard_num"},
			"ddz_up_two_follow_num":      bson.M{"$sum": "$ddz_up_two_follow_num"},
			"ddz_up_three_discard_num":   bson.M{"$sum": "$ddz_up_three_discard_num"},
			"ddz_up_three_follow_num":    bson.M{"$sum": "$ddz_up_three_follow_num"},
			"ddz_up_four_discard_num":    bson.M{"$sum": "$ddz_up_four_discard_num"},
			"ddz_up_four_follow_num":     bson.M{"$sum": "$ddz_up_four_follow_num"},
			"ddz_up_five_discard_num":    bson.M{"$sum": "$ddz_up_five_discard_num"},
			"ddz_up_five_follow_num":     bson.M{"$sum": "$ddz_up_five_follow_num"},
			"ddz_down_number":            bson.M{"$sum": "$ddz_down_number"},
			"ddz_down_rounds":            bson.M{"$sum": "$ddz_down_rounds"},
			"ddz_down_bets":              bson.M{"$sum": "$ddz_down_bets"},
			"ddz_down_win_rounds":        bson.M{"$sum": "$ddz_down_win_rounds"},
			"ddz_down_lose_rounds":       bson.M{"$sum": "$ddz_down_lose_rounds"},
			"ddz_down_two_discard_num":   bson.M{"$sum": "$ddz_down_two_discard_num"},
			"ddz_down_two_follow_num":    bson.M{"$sum": "$ddz_down_two_follow_num"},
			"ddz_down_three_discard_num": bson.M{"$sum": "$ddz_down_three_discard_num"},
			"ddz_down_three_follow_num":  bson.M{"$sum": "$ddz_down_three_follow_num"},
			"ddz_down_four_discard_num":  bson.M{"$sum": "$ddz_down_four_discard_num"},
			"ddz_down_four_follow_num":   bson.M{"$sum": "$ddz_down_four_follow_num"},
			"ddz_down_five_discard_num":  bson.M{"$sum": "$ddz_down_five_discard_num"},
			"ddz_down_five_follow_num":   bson.M{"$sum": "$ddz_down_five_follow_num"},
			"dgp_up_number":              bson.M{"$sum": "$dgp_up_number"},
			"dgp_up_rounds":              bson.M{"$sum": "$dgp_up_rounds"},
			"dgp_up_bets":                bson.M{"$sum": "$dgp_up_bets"},
			"dgp_up_win_rounds":          bson.M{"$sum": "$dgp_up_win_rounds"},
			"dgp_up_lose_rounds":         bson.M{"$sum": "$dgp_up_lose_rounds"},
			"dgp_up_two_discard_num":     bson.M{"$sum": "$dgp_up_two_discard_num"},
			"dgp_up_two_follow_num":      bson.M{"$sum": "$dgp_up_two_follow_num"},
			"dgp_up_three_discard_num":   bson.M{"$sum": "$dgp_up_three_discard_num"},
			"dgp_up_three_follow_num":    bson.M{"$sum": "$dgp_up_three_follow_num"},
			"dgp_up_four_discard_num":    bson.M{"$sum": "$dgp_up_four_discard_num"},
			"dgp_up_four_follow_num":     bson.M{"$sum": "$dgp_up_four_follow_num"},
			"dgp_up_five_discard_num":    bson.M{"$sum": "$dgp_up_five_discard_num"},
			"dgp_up_five_follow_num":     bson.M{"$sum": "$dgp_up_five_follow_num"},
			"dgp_down_number":            bson.M{"$sum": "$dgp_down_number"},
			"dgp_down_rounds":            bson.M{"$sum": "$dgp_down_rounds"},
			"dgp_down_bets":              bson.M{"$sum": "$dgp_down_bets"},
			"dgp_down_win_rounds":        bson.M{"$sum": "$dgp_down_win_rounds"},
			"dgp_down_lose_rounds":       bson.M{"$sum": "$dgp_down_lose_rounds"},
			"dgp_down_two_discard_num":   bson.M{"$sum": "$dgp_down_two_discard_num"},
			"dgp_down_two_follow_num":    bson.M{"$sum": "$dgp_down_two_follow_num"},
			"dgp_down_three_discard_num": bson.M{"$sum": "$dgp_down_three_discard_num"},
			"dgp_down_three_follow_num":  bson.M{"$sum": "$dgp_down_three_follow_num"},
			"dgp_down_four_discard_num":  bson.M{"$sum": "$dgp_down_four_discard_num"},
			"dgp_down_four_follow_num":   bson.M{"$sum": "$dgp_down_four_follow_num"},
			"dgp_down_five_discard_num":  bson.M{"$sum": "$dgp_down_five_discard_num"},
			"dgp_down_five_follow_num":   bson.M{"$sum": "$dgp_down_five_follow_num"},
		}},
	}
	// 查总条数
	var res []entity.TPPlayerRoomStat
	err = TPRoomPlayerStats.Pipe(m1).All(&res)
	if err != nil {
		return
	}
	roomlist := make([]*entity.TPRoomStat, 0)
	turnlist := new(entity.TPGameTurn)
	winlist := new(entity.TPGameWin)
	loselist := new(entity.TPGameLose)
	twolist := new(entity.TPGameRate)
	threelist := new(entity.TPGameRate)
	fourlist := new(entity.TPGameRate)
	fivelist := new(entity.TPGameRate)
	for _, item := range res {
		new_info := new(entity.TPRoomStat)
		new_info.Roomid = item.Id
		new_info.AllRounds = item.AllRounds
		new_info.Bets = item.Bets
		new_info.FBets = ComputeFloat(item.Bets, 100)
		new_info.RoundBetAvg = ComputeFloat(item.Bets, item.AllRounds) / 100.0
		new_info.TotalProfit = ComputeFloat(item.Bets-item.WinBets, 100)
		new_info.UpNumber = item.UpNumber
		new_info.DownNumber = item.DownNumber
		new_info.TwoWinRate = ComputeFloat(item.WinTwoRounds, item.TwoRounds) * 100.0
		new_info.ThreeWinRate = ComputeFloat(item.WinThreeRounds, item.ThreeRounds) * 100.0
		new_info.FourWinRate = ComputeFloat(item.WinFourRounds, item.FourRounds) * 100.0
		new_info.FiveWinRate = ComputeFloat(item.WinFiveRounds, item.FiveRounds) * 100.0
		new_info.TwoBonusRate = ComputeFloat(item.WinTwoBets, item.TwoBets) / 100.0
		new_info.ThreeBonusRate = ComputeFloat(item.WinThreeBets, item.ThreeBets) / 100.0
		new_info.FourBonusRate = ComputeFloat(item.WinFourBets, item.FourBets) / 100.0
		new_info.FiveBonusRate = ComputeFloat(item.WinFiveBets, item.FiveBets) / 100.0
		new_info.ChipPool = ComputeFloat(item.ChipPool, 100)
		new_info.BZUpBetAvg = ComputeFloat(item.BZUpBets, item.BZUpNumber) / 100.0
		new_info.BZDownBetAvg = ComputeFloat(item.BZDownBets, item.BZDownNumber) / 100.0
		new_info.THSUpBetAvg = ComputeFloat(item.THSUpBets, item.THSUpNumber) / 100.0
		new_info.THSDownBetAvg = ComputeFloat(item.THSDownBets, item.THSDownNumber) / 100.0
		new_info.DSZUpBetAvg = ComputeFloat(item.DSZUpBets, item.DSZUpNumber) / 100.0
		new_info.DSZDownBetAvg = ComputeFloat(item.DSZDownBets, item.DSZDownNumber) / 100.0
		new_info.DTHUpBetAvg = ComputeFloat(item.DTHUpBets, item.DTHUpNumber) / 100.0
		new_info.DDZUpBetAvg = ComputeFloat(item.DDZUpBets, item.DDZUpNumber) / 100.0
		new_info.DDZDownBetAvg = ComputeFloat(item.DDZDownBets, item.DDZDownNumber) / 100.0
		new_info.DGPUpBetAvg = ComputeFloat(item.DGPUpBets, item.DGPUpNumber) / 100.0
		if item.DGPDownBets != 0 {
			new_info.DGPDownBetAvg = ComputeFloat(item.DGPDownBets, 100) / float64(item.DGPDownNumber) // ComputeFloat(item.DGPDownBets, item.DGPDownNumber) / 100.0
		}
		roomlist = append(roomlist, new_info)
		// 期望轮次
		turnlist.BZUpNumber += item.BZUpNumber
		turnlist.BZUpRounds += item.BZUpRounds
		turnlist.BZDownNumber += item.BZDownNumber
		turnlist.BZDownRounds += item.BZDownRounds
		turnlist.THSUpNumber += item.THSUpNumber
		turnlist.THSUpRounds += item.THSUpRounds
		turnlist.THSDownNumber += item.THSDownNumber
		turnlist.THSDownRounds += item.THSDownRounds
		turnlist.DSZUpNumber += item.DSZUpNumber
		turnlist.DSZUpRounds += item.DSZUpRounds
		turnlist.DSZDownNumber += item.DSZDownNumber
		turnlist.DSZDownRounds += item.DSZDownRounds
		turnlist.DTHUpNumber += item.DTHUpNumber
		turnlist.DTHUpRounds += item.DTHUpRounds
		turnlist.DTHDownNumber += item.DTHDownNumber
		turnlist.DTHDownRounds += item.DTHDownRounds
		turnlist.DDZUpNumber += item.DDZUpNumber
		turnlist.DDZUpRounds += item.DDZUpRounds
		turnlist.DDZDownNumber += item.DDZDownNumber
		turnlist.DDZDownRounds += item.DDZDownRounds
		turnlist.DGPUpNumber += item.DGPUpNumber
		turnlist.DGPUpRounds += item.DGPUpRounds
		turnlist.DGPDownNumber += item.DGPDownNumber
		turnlist.DGPDownRounds += item.DGPDownRounds

		// 大赢局
		winlist.BZUpWinRounds += item.BZUpWinRounds
		winlist.BZDownWinRounds += item.BZDownWinRounds
		winlist.THSUpWinRounds += item.THSUpWinRounds
		winlist.THSDownWinRounds += item.THSDownWinRounds
		winlist.DSZUpWinRounds += item.DSZUpWinRounds
		winlist.DSZDownWinRounds += item.DSZDownWinRounds
		winlist.DTHUpWinRounds += item.DTHUpWinRounds
		winlist.DTHDownWinRounds += item.DTHDownWinRounds
		winlist.DDZUpWinRounds += item.DDZUpWinRounds
		winlist.DDZDownWinRounds += item.DDZDownWinRounds
		winlist.DGPUpWinRounds += item.DGPUpWinRounds
		winlist.DGPDownWinRounds += item.DGPDownWinRounds

		// 大输家
		loselist.BZUpLoseRounds += item.BZUpLoseRounds
		loselist.BZDownLoseRounds += item.BZUpLoseRounds
		loselist.THSUpLoseRounds += item.THSUpLoseRounds
		loselist.THSDownLoseRounds += item.THSDownLoseRounds
		loselist.DSZUpLoseRounds += item.DSZUpLoseRounds
		loselist.DSZDownLoseRounds += item.DSZDownLoseRounds
		loselist.DTHUpLoseRounds += item.DTHUpLoseRounds
		loselist.DTHDownLoseRounds += item.DTHDownLoseRounds
		loselist.DDZUpLoseRounds += item.DDZUpLoseRounds
		loselist.DDZDownLoseRounds += item.DDZDownLoseRounds
		loselist.DGPUpLoseRounds += item.DGPUpLoseRounds
		loselist.DGPDownLoseRounds += item.DGPDownLoseRounds
		// 总信息
		stats.AllRounds += item.AllRounds
		stats.Bets += item.Bets
		stats.WinBets += item.WinBets
		stats.Userid = item.Id

		// 两人局弃牌率
		twolist.BZUpDiscardRate = ComputeFloat(item.BZUpTwoDiscardNum, item.BZUpNumber) * 100.0
		twolist.BZDownDiscardRate = ComputeFloat(item.BZDownTwoDiscardNum, item.BZDownNumber) * 100.0
		twolist.THSUpDiscardRate = ComputeFloat(item.THSUpTwoDiscardNum, item.THSUpNumber) * 100.0
		twolist.THSDownDiscardRate = ComputeFloat(item.THSDownTwoDiscardNum, item.THSDownNumber) * 100.0
		twolist.DSZUpDiscardRate = ComputeFloat(item.DSZUpTwoDiscardNum, item.DSZUpNumber) * 100.0
		twolist.DSZDownDiscardRate = ComputeFloat(item.DSZDownTwoDiscardNum, item.DSZDownNumber) * 100.0
		twolist.DTHUpDiscardRate = ComputeFloat(item.DTHUpTwoDiscardNum, item.DTHUpNumber) * 100.0
		twolist.DTHDownDiscardRate = ComputeFloat(item.DTHDownTwoDiscardNum, item.DTHDownNumber) * 100.0
		twolist.DDZUpDiscardRate = ComputeFloat(item.DDZUpTwoDiscardNum, item.DDZUpNumber) * 100.0
		twolist.DDZDownDiscardRate = ComputeFloat(item.DDZDownTwoDiscardNum, item.DDZDownNumber) * 100.0
		twolist.DGPUpDiscardRate = ComputeFloat(item.DGPUpTwoDiscardNum, item.DGPUpNumber) * 100.0
		twolist.DGPDownDiscardRate = ComputeFloat(item.DGPDownTwoDiscardNum, item.DGPDownNumber) * 100.0
		// 两人局跟注率
		twolist.BZUpFollowRate = ComputeFloat(item.BZUpTwoFollowNum, item.BZUpNumber) * 100.0
		twolist.BZDownFollowRate = ComputeFloat(item.BZDownTwoFollowNum, item.BZDownNumber) * 100.0
		twolist.THSUpFollowRate = ComputeFloat(item.THSUpTwoFollowNum, item.THSUpNumber) * 100.0
		twolist.THSDownFollowRate = ComputeFloat(item.THSDownTwoFollowNum, item.THSDownNumber) * 100.0
		twolist.DSZUpFollowRate = ComputeFloat(item.DSZUpTwoFollowNum, item.DSZUpNumber) * 100.0
		twolist.DSZDownFollowRate = ComputeFloat(item.DSZDownTwoFollowNum, item.DSZDownNumber) * 100.0
		twolist.DTHUpFollowRate = ComputeFloat(item.DTHUpTwoFollowNum, item.DTHUpNumber) * 100.0
		twolist.DTHDownFollowRate = ComputeFloat(item.DTHDownTwoFollowNum, item.DTHDownNumber) * 100.0
		twolist.DDZUpFollowRate = ComputeFloat(item.DDZUpTwoFollowNum, item.DDZUpNumber) * 100.0
		twolist.DDZDownFollowRate = ComputeFloat(item.DDZDownTwoFollowNum, item.DDZDownNumber) * 100.0
		twolist.DGPUpFollowRate = ComputeFloat(item.DGPUpTwoFollowNum, item.DGPUpNumber) * 100.0
		twolist.DGPDownFollowRate = ComputeFloat(item.DGPDownTwoFollowNum, item.DGPDownNumber) * 100.0

		// 三人局弃牌率
		threelist.BZUpDiscardRate = ComputeFloat(item.BZUpThreeDiscardNum, item.BZUpNumber) * 100.0
		threelist.BZDownDiscardRate = ComputeFloat(item.BZDownThreeDiscardNum, item.BZDownNumber) * 100.0
		threelist.THSUpDiscardRate = ComputeFloat(item.THSUpThreeDiscardNum, item.THSUpNumber) * 100.0
		threelist.THSDownDiscardRate = ComputeFloat(item.THSDownThreeDiscardNum, item.THSDownNumber) * 100.0
		threelist.DSZUpDiscardRate = ComputeFloat(item.DSZUpThreeDiscardNum, item.DSZUpNumber) * 100.0
		threelist.DSZDownDiscardRate = ComputeFloat(item.DSZDownThreeDiscardNum, item.DSZDownNumber) * 100.0
		threelist.DTHUpDiscardRate = ComputeFloat(item.DTHUpThreeDiscardNum, item.DTHUpNumber) * 100.0
		threelist.DTHDownDiscardRate = ComputeFloat(item.DTHDownThreeDiscardNum, item.DTHDownNumber) * 100.0
		threelist.DDZUpDiscardRate = ComputeFloat(item.DDZUpThreeDiscardNum, item.DDZUpNumber) * 100.0
		threelist.DDZDownDiscardRate = ComputeFloat(item.DDZDownThreeDiscardNum, item.DDZDownNumber) * 100.0
		threelist.DGPUpDiscardRate = ComputeFloat(item.DGPUpThreeDiscardNum, item.DGPUpNumber) * 100.0
		threelist.DGPDownDiscardRate = ComputeFloat(item.DGPDownThreeDiscardNum, item.DGPDownNumber) * 100.0
		// 三人局跟注率
		threelist.BZUpFollowRate = ComputeFloat(item.BZUpThreeFollowNum, item.BZUpNumber) * 100.0
		threelist.BZDownFollowRate = ComputeFloat(item.BZDownThreeFollowNum, item.BZDownNumber) * 100.0
		threelist.THSUpFollowRate = ComputeFloat(item.THSUpThreeFollowNum, item.THSUpNumber) * 100.0
		threelist.THSDownFollowRate = ComputeFloat(item.THSDownThreeFollowNum, item.THSDownNumber) * 100.0
		threelist.DSZUpFollowRate = ComputeFloat(item.DSZUpThreeFollowNum, item.DSZUpNumber) * 100.0
		threelist.DSZDownFollowRate = ComputeFloat(item.DSZDownThreeFollowNum, item.DSZDownNumber) * 100.0
		threelist.DTHUpFollowRate = ComputeFloat(item.DTHUpThreeFollowNum, item.DTHUpNumber) * 100.0
		threelist.DTHDownFollowRate = ComputeFloat(item.DTHDownThreeFollowNum, item.DTHDownNumber) * 100.0
		threelist.DDZUpFollowRate = ComputeFloat(item.DDZUpThreeFollowNum, item.DDZUpNumber) * 100.0
		threelist.DDZDownFollowRate = ComputeFloat(item.DDZDownThreeFollowNum, item.DDZDownNumber) * 100.0
		threelist.DGPUpFollowRate = ComputeFloat(item.DGPUpThreeFollowNum, item.DGPUpNumber) * 100.0
		threelist.DGPDownFollowRate = ComputeFloat(item.DGPDownThreeFollowNum, item.DGPDownNumber) * 100.0

		// 四人局弃牌率
		fourlist.BZUpDiscardRate = ComputeFloat(item.BZUpFourDiscardNum, item.BZUpNumber) * 100.0
		fourlist.BZDownDiscardRate = ComputeFloat(item.BZDownFourDiscardNum, item.BZDownNumber) * 100.0
		fourlist.THSUpDiscardRate = ComputeFloat(item.THSUpFourDiscardNum, item.THSUpNumber) * 100.0
		fourlist.THSDownDiscardRate = ComputeFloat(item.THSDownFourDiscardNum, item.THSDownNumber) * 100.0
		fourlist.DSZUpDiscardRate = ComputeFloat(item.DSZUpFourDiscardNum, item.DSZUpNumber) * 100.0
		fourlist.DSZDownDiscardRate = ComputeFloat(item.DSZDownFourDiscardNum, item.DSZDownNumber) * 100.0
		fourlist.DTHUpDiscardRate = ComputeFloat(item.DTHUpFourDiscardNum, item.DTHUpNumber) * 100.0
		fourlist.DTHDownDiscardRate = ComputeFloat(item.DTHDownFourDiscardNum, item.DTHDownNumber) * 100.0
		fourlist.DDZUpDiscardRate = ComputeFloat(item.DDZUpFourDiscardNum, item.DDZUpNumber) * 100.0
		fourlist.DDZDownDiscardRate = ComputeFloat(item.DDZDownFourDiscardNum, item.DDZDownNumber) * 100.0
		fourlist.DGPUpDiscardRate = ComputeFloat(item.DGPUpFourDiscardNum, item.DGPUpNumber) * 100.0
		fourlist.DGPDownDiscardRate = ComputeFloat(item.DGPDownFourDiscardNum, item.DGPDownNumber) * 100.0
		// 四人局跟注率
		fourlist.BZUpFollowRate = ComputeFloat(item.BZUpFourFollowNum, item.BZUpNumber) * 100.0
		fourlist.BZDownFollowRate = ComputeFloat(item.BZDownFourFollowNum, item.BZDownNumber) * 100.0
		fourlist.THSUpFollowRate = ComputeFloat(item.THSUpFourFollowNum, item.THSUpNumber) * 100.0
		fourlist.THSDownFollowRate = ComputeFloat(item.THSDownFourFollowNum, item.THSDownNumber) * 100.0
		fourlist.DSZUpFollowRate = ComputeFloat(item.DSZUpFourFollowNum, item.DSZUpNumber) * 100.0
		fourlist.DSZDownFollowRate = ComputeFloat(item.DSZDownFourFollowNum, item.DSZDownNumber) * 100.0
		fourlist.DTHUpFollowRate = ComputeFloat(item.DTHUpFourFollowNum, item.DTHUpNumber) * 100.0
		fourlist.DTHDownFollowRate = ComputeFloat(item.DTHDownFourFollowNum, item.DTHDownNumber) * 100.0
		fourlist.DDZUpFollowRate = ComputeFloat(item.DDZUpFourFollowNum, item.DDZUpNumber) * 100.0
		fourlist.DDZDownFollowRate = ComputeFloat(item.DDZDownFourFollowNum, item.DDZDownNumber) * 100.0
		fourlist.DGPUpFollowRate = ComputeFloat(item.DGPUpFourFollowNum, item.DGPUpNumber) * 100.0
		fourlist.DGPDownFollowRate = ComputeFloat(item.DGPDownFourFollowNum, item.DGPDownNumber) * 100.0

		// 五人局弃牌率
		fivelist.BZUpDiscardRate = ComputeFloat(item.BZUpFiveDiscardNum, item.BZUpNumber) * 100.0
		fivelist.BZDownDiscardRate = ComputeFloat(item.BZDownFiveDiscardNum, item.BZDownNumber) * 100.0
		fivelist.THSUpDiscardRate = ComputeFloat(item.THSUpFiveDiscardNum, item.THSUpNumber) * 100.0
		fivelist.THSDownDiscardRate = ComputeFloat(item.THSDownFiveDiscardNum, item.THSDownNumber) * 100.0
		fivelist.DSZUpDiscardRate = ComputeFloat(item.DSZUpFiveDiscardNum, item.DSZUpNumber) * 100.0
		fivelist.DSZDownDiscardRate = ComputeFloat(item.DSZDownFiveDiscardNum, item.DSZDownNumber) * 100.0
		fivelist.DTHUpDiscardRate = ComputeFloat(item.DTHUpFiveDiscardNum, item.DTHUpNumber) * 100.0
		fivelist.DTHDownDiscardRate = ComputeFloat(item.DTHDownFiveDiscardNum, item.DTHDownNumber) * 100.0
		fivelist.DDZUpDiscardRate = ComputeFloat(item.DDZUpFiveDiscardNum, item.DDZUpNumber) * 100.0
		fivelist.DDZDownDiscardRate = ComputeFloat(item.DDZDownFiveDiscardNum, item.DDZDownNumber) * 100.0
		fivelist.DGPUpDiscardRate = ComputeFloat(item.DGPUpFiveDiscardNum, item.DGPUpNumber) * 100.0
		fivelist.DGPDownDiscardRate = ComputeFloat(item.DGPDownFiveDiscardNum, item.DGPDownNumber) * 100.0
		// 五人局跟注率
		fivelist.BZUpFollowRate = ComputeFloat(item.BZUpFiveFollowNum, item.BZUpNumber) * 100.0
		fivelist.BZDownFollowRate = ComputeFloat(item.BZDownFiveFollowNum, item.BZDownNumber) * 100.0
		fivelist.THSUpFollowRate = ComputeFloat(item.THSUpFiveFollowNum, item.THSUpNumber) * 100.0
		fivelist.THSDownFollowRate = ComputeFloat(item.THSDownFiveFollowNum, item.THSDownNumber) * 100.0
		fivelist.DSZUpFollowRate = ComputeFloat(item.DSZUpFiveFollowNum, item.DSZUpNumber) * 100.0
		fivelist.DSZDownFollowRate = ComputeFloat(item.DSZDownFiveFollowNum, item.DSZDownNumber) * 100.0
		fivelist.DTHUpFollowRate = ComputeFloat(item.DTHUpFiveFollowNum, item.DTHUpNumber) * 100.0
		fivelist.DTHDownFollowRate = ComputeFloat(item.DTHDownFiveFollowNum, item.DTHDownNumber) * 100.0
		fivelist.DDZUpFollowRate = ComputeFloat(item.DDZUpFiveFollowNum, item.DDZUpNumber) * 100.0
		fivelist.DDZDownFollowRate = ComputeFloat(item.DDZDownFiveFollowNum, item.DDZDownNumber) * 100.0
		fivelist.DGPUpFollowRate = ComputeFloat(item.DGPUpFiveFollowNum, item.DGPUpNumber) * 100.0
		fivelist.DGPDownFollowRate = ComputeFloat(item.DGPDownFiveFollowNum, item.DGPDownNumber) * 100.0
	}
	if stats.Bets > 0 {
		stats.FBets = ComputeFloat(stats.Bets, 100)
		if stats.Bets != 0 {
			stats.RoundBetAvg = ComputeFloat(stats.Bets, 100) / float64(stats.AllRounds)
			stats.TotalProfit = ComputeFloat(stats.Bets-stats.WinBets, 100)
		}
	}
	if turnlist.BZUpRounds != 0 {
		turnlist.BZUpRoundAvg = ComputeFloat(turnlist.BZUpRounds, turnlist.BZUpNumber)
	}
	if turnlist.BZDownRounds != 0 {
		turnlist.BZDownRoundAvg = ComputeFloat(turnlist.BZDownRounds, turnlist.BZDownNumber)
	}

	if turnlist.THSUpRounds != 0 {
		turnlist.THSUpRoundAvg = ComputeFloat(turnlist.THSUpRounds, turnlist.THSUpNumber)
	}
	if turnlist.THSDownRounds != 0 {
		turnlist.THSDownRoundAvg = ComputeFloat(turnlist.THSDownRounds, turnlist.THSDownNumber)
	}

	if turnlist.DSZUpRounds != 0 {
		turnlist.DSZUpRoundAvg = ComputeFloat(turnlist.DSZUpRounds, turnlist.DSZUpNumber)
	}
	if turnlist.DSZDownRounds != 0 {
		turnlist.DSZDownRoundAvg = ComputeFloat(turnlist.DSZDownRounds, turnlist.DSZDownNumber)
	}

	if turnlist.DTHUpRounds != 0 {
		turnlist.DTHUpRoundAvg = ComputeFloat(turnlist.DTHUpRounds, turnlist.DTHUpNumber)
	}
	if turnlist.DTHDownRounds != 0 {
		turnlist.DTHDownRoundAvg = ComputeFloat(turnlist.DTHDownRounds, turnlist.DTHDownNumber)
	}

	if turnlist.DDZUpRounds != 0 {
		turnlist.DDZUpRoundAvg = ComputeFloat(turnlist.DDZUpRounds, turnlist.DDZUpNumber)
	}
	if turnlist.DDZDownRounds != 0 {
		turnlist.DDZDownRoundAvg = ComputeFloat(turnlist.DDZDownRounds, turnlist.DDZDownNumber)
	}

	if turnlist.DGPUpRounds != 0 {
		turnlist.DGPUpRoundAvg = ComputeFloat(turnlist.DGPUpRounds, turnlist.DGPUpNumber)
	}
	if turnlist.DGPDownRounds != 0 {
		turnlist.DGPDownRoundAvg = ComputeFloat(turnlist.DDZDownRounds, turnlist.DGPDownNumber)
	}
	stats.RoomInfo = roomlist
	stats.TurnInfo = append(stats.TurnInfo, turnlist)
	stats.WinInfo = append(stats.WinInfo, winlist)
	stats.LoseInfo = append(stats.LoseInfo, loselist)
	stats.TwoRateInfo = append(stats.TwoRateInfo, twolist)
	stats.ThreeRateInfo = append(stats.ThreeRateInfo, threelist)
	stats.FourRateInfo = append(stats.FourRateInfo, fourlist)
	stats.FiveRateInfo = append(stats.FiveRateInfo, fivelist)

	// TP盈亏曲线
	list := make([]entity.Detail, 0)
	m2 := bson.M{}
	m2["gtype"] = 1
	if userid != "" {
		m2["players"] = bson.M{
			"$regex":   fmt.Sprintf("\\b%s\\b", userid),
			"$options": "i",
		}
	}
	Details.Find(m2).All(&list)
	dates := make([]interface{}, 0) // 日期
	data := make([]interface{}, 0)  // TP
	data1 := make([]interface{}, 0) // pay
	data2 := make([]interface{}, 0) // withdraw
	data3 := make([]interface{}, 0) // uproom
	data4 := make([]interface{}, 0) // downroom
	data5 := make([]interface{}, 0) // InPay
	// 获取房间最低准入
	g_m := bson.M{}
	g_m["gtype"] = 1
	var gamelist []entity.Game
	Games.Find(g_m).All(&gamelist)
	if len(list) > 0 {
		sort.Slice(list, func(i, j int) bool {
			return list[i].BeginTime < list[j].BeginTime
		})
		for _, item := range list {
			c, _ := service.ConvertToIndiaTime(item.BeginTime)
			strTime := c.Format("2006-01-02 15:04:05")
			dates = append(dates, strTime)
			switch item.Gtype {
			case 1:
				// TP
				for _, v := range item.TPDetail {
					if v.UserId == userid {
						res := service.Chip2Float(v.Score)
						data = append(data, res)

						minRoomId := ""
						maxRoomId := ""
						minAccess := 0
						maxAccess := 0
						totalNum := 0
						carrycoin := int(v.BeforeScore)
						// 根据房间信息获取筹码池、次数
						for _, room := range gamelist {
							if carrycoin >= room.Min_Access {
								if minRoomId == "" {
									minRoomId = room.Id
									minAccess = room.Min_Access
								}
								if maxRoomId == "" {
									maxRoomId = room.Id
									maxAccess = room.Min_Access
								}
								totalNum++
								if minAccess > room.Min_Access {
									minAccess = room.Min_Access
									minRoomId = room.Id
								}
								if maxAccess < room.Min_Access {
									maxAccess = room.Min_Access
									maxRoomId = room.Id
								}
							}
						}
						if totalNum >= 2 {
							if maxRoomId == item.RoomId {
								data3 = append(data3, 1)
							}
							if minRoomId == item.RoomId {
								data4 = append(data4, 1)
							}
						}
					}
				}
			}
		}
	}
	// 支付
	paylist := make([]entity.PayOrder, 0)
	pay_m := bson.M{}
	pay_m["userid"] = userid
	pay_m["order_status"] = 4
	Pays.Pipe(pay_m).All(&paylist)
	for _, p := range paylist {
		res := service.Chip2Float(int64(p.Amount))
		data1 = append(data1, res)
	}
	// 提现
	withlist := make([]entity.WithdrawOrder, 0)
	with_m := bson.M{}
	with_m["userid"] = userid
	with_m["order_status"] = 2
	Pays.Pipe(with_m).All(&withlist)
	for _, p := range withlist {
		res := service.Chip2Float(int64(p.Amount))
		data2 = append(data2, res)
	}
	// 局内充值 shop_type = 5
	inPaylist := make([]entity.PayOrder, 0)
	pay_m1 := bson.M{}
	pay_m1["userid"] = userid
	pay_m1["order_status"] = 4
	pay_m1["shop_type"] = 5
	Pays.Pipe(pay_m1).All(&inPaylist)
	for _, p := range inPaylist {
		res := service.Chip2Float(int64(p.Amount))
		data5 = append(data5, res)
	}
	line.Dates = dates
	line.TP = data
	line.Pay = data1
	line.Withdraw = data2
	line.UpRoom = data3
	line.DownRoom = data4
	line.InPay = data5
	return
}

// GetLHDStatPlayers lhd玩家明细
func (s *gameStatsService) GetLHDStatPlayers(page, pageSize int, m, m2, m3 bson.M, startDate, endDate string) (
	stats []*entity.LHDPlayerStat, count int, err error,
) {
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)

	// 用户列表筛选
	var filterPlayerIds []string
	var userids []string
	if len(m2) > 0 || len(m3) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				"_id":        "$userid",
				"bets":       bson.M{"$sum": "$bets"},
				"bet_rounds": bson.M{"$sum": "$bet_rounds"},
				"cash":       bson.M{"$sum": "$cash"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				"cash": "$cash",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = LHDPlayerStats.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetLHDStatPlayers error31:", err)
		}
		for _, r := range r3_1 {
			userid := r["_id"].(string)
			userids = append(userids, userid)
		}
		if len(userids) == 0 { // 未匹配到
			return
		}
		now := utils.BsonNow()
		m3_2 := []bson.M{
			{"$match": bson.M{"_id": bson.M{"$in": userids}}},
			{"$project": bson.M{
				"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
				"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
			}},
			{"$match": m3},
		}
		var r3_2 []bson.M
		err = PlayerUsers.Pipe(m3_2).All(&r3_2)
		if err != nil {
			beego.Error("GetLHDStatPlayers error32:", err)
		} else if len(r3_2) == 0 { // 未匹配到
			return
		}
		for _, r := range r3_2 {
			userid := r["_id"].(string)
			filterPlayerIds = append(filterPlayerIds, userid)
		}
	}
	if len(filterPlayerIds) > 0 {
		m["userid"] = bson.M{"$in": filterPlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":                "$userid",
			"date":               bson.M{"$max": "$date"},
			"all_rounds":         bson.M{"$sum": "$all_rounds"},
			"game_times":         bson.M{"$sum": "$game_times"},
			"bets":               bson.M{"$sum": "$bets"},
			"bet_rounds":         bson.M{"$sum": "$bet_rounds"},
			"win_rounds":         bson.M{"$sum": "$win_rounds"},
			"lose_rounds":        bson.M{"$sum": "$lose_rounds"},
			"tie_rounds":         bson.M{"$sum": "$tie_rounds"},
			"win_bets":           bson.M{"$sum": "$win_bets"},
			"lose_bets":          bson.M{"$sum": "$lose_bets"},
			"tie_bets":           bson.M{"$sum": "$tie_bets"},
			"wins":               bson.M{"$sum": "$wins"},
			"loses":              bson.M{"$sum": "$loses"},
			"cash":               bson.M{"$sum": "$cash"},
			"dragons":            bson.M{"$sum": "$dragons"},
			"tigers":             bson.M{"$sum": "$tigers"},
			"ties":               bson.M{"$sum": "$ties"},
			"player_dragons":     bson.M{"$sum": "$player_dragons"},
			"player_tigers":      bson.M{"$sum": "$player_tigers"},
			"player_ties":        bson.M{"$sum": "$player_ties"},
			"player_win_dragons": bson.M{"$sum": "$player_win_dragons"},
			"player_win_tigers":  bson.M{"$sum": "$player_win_tigers"},
			"player_win_ties":    bson.M{"$sum": "$player_win_ties"},
			"player_multis":      bson.M{"$sum": "$player_multis"},
			"round_bet_avg":      bson.M{"$sum": "$round_bet_avg"},
			"money":              bson.M{"$sum": "$money"},
			"observe_rounds":     bson.M{"$sum": "$observe_rounds"},
		}},
		{"$project": bson.M{
			"_id": "$_id", "all_rounds": "$all_rounds", "game_times": "$game_times", "bets": "$bets", "bet_rounds": "$bet_rounds", "win_rounds": "$win_rounds", "lose_rounds": "$lose_rounds", "tie_rounds": "$tie_rounds", "win_bets": "$win_bets", "lose_bets": "$lose_bets", "tie_bets": "$tie_bets", "wins": "$wins", "loses": "$loses", "cash": "$cash", "dragons": "$dragons", "tigers": "$tigers", "ties": "$ties", "player_dragons": "$player_dragons", "player_tigers": "$player_tigers", "player_ties": "$player_ties", "player_win_dragons": "$player_win_dragons", "player_win_tigers": "$player_win_tigers", "player_win_ties": "$player_win_ties", "player_multis": "$player_multis", "money": "$money", "observe_rounds": "$observe_rounds",
			// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}}, // "can't $divide by zero"
			"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
		}},
		{"$match": m2},
		{"$count": "count"},
	}
	// 查总条数
	var rCount []bson.M
	err = LHDPlayerStats.Pipe(m1).All(&rCount)
	if err != nil {
		return
	} else if len(rCount) > 0 {
		count = rCount[0]["count"].(int)
	}

	// 查列表数据
	m1 = m1[:len(m1)-1]
	m1 = append(m1, []bson.M{
		{"$sort": bson.M{"date": -1}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}...)
	err = LHDPlayerStats.Pipe(m1).All(&stats)
	if err != nil {
		return
	}
	lhdStatPlayersMapping(stats)

	// 查总汇数据
	mark := "--"
	summary := &entity.LHDPlayerStat{}
	m1 = m1[:len(m1)-3]
	m1[1]["$group"].(bson.M)["_id"] = nil
	delete(m1[1]["$group"].(bson.M), "date")
	m1[1]["$group"].(bson.M)["userids"] = bson.M{"$addToSet": "$userid"}
	m1[2]["$project"].(bson.M)["userids"] = "$userids"
	var summarys []*entity.LHDPlayerStat
	err = LHDPlayerStats.Pipe(m1).All(&summarys)
	if err != nil {
		return
	} else if len(summarys) > 0 {
		summary = summarys[0]
		summary.No = fmt.Sprintf("%s-%s总汇", startDate, endDate)
		summary.Id = mark
		summary.FNickname = mark
		summary.FUserType = mark
		summary.FRegistArea = mark
		summary.FCtime = mark
		summary.FLoginTime = mark
		// summary.FLiveDays = mark
		// summary.FLoseDays = mark
		lhdStatPlayersSummaryMapping(summary)
	}
	if len(summarys) > 0 {
		stats = append([]*entity.LHDPlayerStat{summary}, stats...)
	}
	return
}

func lhdStatPlayersMapping(stats []*entity.LHDPlayerStat) {
	var playerIds []string
	// var playerStats = make(map[string]*entity.CrashPlayerStat, len(stats))
	for _, stat := range stats {
		playerIds = append(playerIds, stat.Id)
		// playerStats[stat.Id] = stat
	}
	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": playerIds}}},
		{"$project": bson.M{
			"_id":         "$_id",
			"nickname":    "$nickname",
			"money":       "$money",
			"cash_out":    "$cash_out",
			"diamond":     "$diamond",
			"ctime":       "$ctime",
			"login_time":  "$login_time",
			"regist_area": "$regist_area", //ab测试 0:A 1:B 2:C
			"state":       "$state",
		}},
	}
	var r2 []bson.M
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("lhd users error:", err)
		return
	}
	var r2Map = make(map[string]bson.M, len(r2))
	for _, r := range r2 {
		userid := r["_id"].(string)
		r2Map[userid] = r
	}
	now := utils.BsonNow()
	for i, stat := range stats {
		r, ok := r2Map[stat.Id]
		if !ok {
			continue
		}
		// 玩家昵称	玩家类型 账号类型 注册时间 最后登陆时间 存活天数 流失天数
		// userid := r["_id"].(string)
		nickname := r["nickname"].(string)
		money := r["money"].(int)
		cash_out := r["cash_out"].(int)
		diamond := r["diamond"].(int64)
		ctime := r["ctime"].(time.Time)
		login_time := r["login_time"].(time.Time)
		regist_area := r["regist_area"].(int)
		state := r["state"].(int)

		// 将日期转成印度时间
		login_time, _ = ConvertToIndiaTime(login_time.Unix())
		ctime, _ = ConvertToIndiaTime(ctime.Unix())

		stat.No = strconv.Itoa(i + 1)
		stat.FNickname = nickname
		stat.Money = uint32(money)
		stat.FCtime = ctime.Format(utils.FORMAT)
		stat.FLoginTime = login_time.Format(utils.FORMAT)
		stat.FLiveDays = int(math.Ceil(login_time.Sub(ctime).Hours() / 24))
		stat.FLoseDays = int(math.Floor(now.Sub(login_time).Hours() / 24))

		// summary.FLiveDays += stat.FLiveDays
		// summary.FLoseDays += stat.FLoseDays
		// summary.AllRounds += stat.AllRounds
		// summary.GameTimes += stat.GameTimes
		// summary.Bets += stat.Bets
		// summary.BetRounds += stat.BetRounds
		// summary.WinRounds += stat.WinRounds
		// summary.LoseRounds += stat.LoseRounds
		// summary.TieRounds += stat.TieRounds
		// summary.WinBets += stat.WinBets
		// summary.LoseBets += stat.LoseBets
		// summary.TieBets += stat.TieBets
		// summary.Wins += stat.Wins
		// summary.Loses += stat.Loses
		// summary.Cash += stat.Cash
		// summary.MulpitleSum += stat.MulpitleSum
		// summary.WinEscapeMulpitleSum += stat.WinEscapeMulpitleSum
		// summary.WinMulpitleSum += stat.WinMulpitleSum
		// summary.LoseMulpitleSum += stat.LoseMulpitleSum
		// summary.Mulpitles = append(summary.Mulpitles, stat.Mulpitles...)
		// summary.WinEscapeMulpitles = append(summary.WinEscapeMulpitles, stat.WinEscapeMulpitles...)
		// summary.WinMulpitles = append(summary.WinMulpitles, stat.WinMulpitles...)
		// summary.LoseMulpitles = append(summary.LoseMulpitles, stat.LoseMulpitles...)
		// summary.ObserveRounds += stat.ObserveRounds
		// summary.SMoney += money
		// summary.SCashOut += cash_out
		// summary.SDiamond += diamond

		switch regist_area {
		case 0:
			stat.FRegistArea = "A类"
		case 1:
			stat.FRegistArea = "B类"
		case 2:
			stat.FRegistArea = "C类"
		}

		switch state {
		case 1:
			stat.FUserType = "新手"
		case 2:
			if money == 0 {
				stat.FUserType = "零充"
			} else if money >= 20000 && money <= 99999 {
				stat.FUserType = "普充"
			} else if money >= 100000 && money <= 499999 {
				stat.FUserType = "小R"
			} else if money >= 500000 && money <= 999999 {
				stat.FUserType = "中R"
			} else if money >= 1000000 && money <= 9999999 {
				stat.FUserType = "大R"
			} else if money >= 10000000 {
				stat.FUserType = "超大R"
			}
		case 3:
			stat.FUserType = "平民"
		case 4:
			stat.FUserType = "泡沫"
		}

		stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
		if stat.BetRounds > 0 {
			// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.BetRounds)/100.0)
			stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
		}
		stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
		stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
		stat.FTieBets = fmt.Sprintf("%.2f", float64(stat.TieBets)/100.0)
		if stat.BetRounds > 0 {
			stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
		}
		if stat.Bets > 0 {
			// stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
			stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
		}
		if stat.WinRounds > 0 {
			stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
		}
		if stat.TieRounds > 0 {
			stat.FTieBetsAvg = fmt.Sprintf("%.2f", float64(stat.TieBets)/float64(stat.TieRounds)/100.0)
		}
		stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
		stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
		stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)

		if stat.WinRounds > 0 {
			stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
		}
		// 押龙/虎/和率
		if stat.BetRounds > 0 {
			stat.PlayerDragonsRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerDragons)/float64(stat.BetRounds)*100)
			stat.PlayerTigersRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerTigers)/float64(stat.BetRounds)*100)
			stat.PlayerTiesRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerTies)/float64(stat.BetRounds)*100)
		}
		if stat.Dragons > 0 {
			stat.PlayerDragonsWinRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinDragons)/float64(stat.Dragons)*100)
		}
		if stat.Tigers > 0 {
			stat.PlayerTigersWinRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinTigers)/float64(stat.Tigers)*100)
		}
		if stat.Ties > 0 {
			stat.PlayerTiesWinRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinTies)/float64(stat.Ties)*100)
		}
		// 多门率
		if stat.BetRounds > 0 {
			stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerMultis)/float64(stat.BetRounds)*100)
		}

		// 观察局
		if stat.BetRounds > 0 {
			stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
		}
		if stat.BetRounds > 0 {
			stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
		}
		stat.FMoney = fmt.Sprintf("%.2f", float64(money)/100)
		stat.FCashOut = fmt.Sprintf("%.2f", float64(cash_out)/100)
		stat.FWinMoney = fmt.Sprintf("%.2f", float64(int(diamond)+cash_out-money)/100)
	}
}

func lhdStatPlayersSummaryMapping(stat *entity.LHDPlayerStat) {
	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": stat.Userids}}},
		{"$group": bson.M{
			"_id":      nil,
			"money":    bson.M{"$sum": "$money"},
			"cash_out": bson.M{"$sum": "$cash_out"},
			"diamond":  bson.M{"$sum": "$diamond"},
		}},
	}
	var r2 []bson.M
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("lhd users error:", err)
		return
	} else if len(r2) > 0 {
		money := r2[0]["money"].(int)
		cash_out := r2[0]["cash_out"].(int)
		diamond := r2[0]["diamond"].(int64)
		stat.SMoney = money
		stat.SCashOut = cash_out
		stat.SDiamond = diamond
	}

	stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
	if stat.BetRounds > 0 {
		stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.BetRounds)/100.0)
		// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
	}

	stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
	stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
	stat.FTieBets = fmt.Sprintf("%.2f", float64(stat.TieBets)/100.0)
	if stat.BetRounds > 0 {
		stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
	}
	if stat.Bets > 0 {
		stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
	}
	if stat.WinRounds > 0 {
		stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
	}
	if stat.TieRounds > 0 {
		stat.FTieBetsAvg = fmt.Sprintf("%.2f", float64(stat.TieBets)/float64(stat.TieRounds)/100.0)
	}
	stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
	stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
	stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)
	if stat.WinRounds > 0 {
		stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
	}
	// 押龙/虎/和率
	if stat.BetRounds > 0 {
		stat.PlayerDragonsRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerDragons)/float64(stat.BetRounds)*100)
		stat.PlayerTigersRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerTigers)/float64(stat.BetRounds)*100)
		stat.PlayerTiesRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerTies)/float64(stat.BetRounds)*100)
	}
	if stat.Dragons > 0 {
		stat.PlayerDragonsWinRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinDragons)/float64(stat.Dragons)*100)
	}
	if stat.Tigers > 0 {
		stat.PlayerTigersWinRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinTigers)/float64(stat.Tigers)*100)
	}
	if stat.Ties > 0 {
		stat.PlayerTiesWinRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinTies)/float64(stat.Ties)*100)
	}
	// 多门率
	if stat.BetRounds > 0 {
		stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerMultis)/float64(stat.BetRounds)*100)
	}

	// 观察局
	if stat.BetRounds > 0 {
		stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
	}
	if stat.BetRounds > 0 {
		stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
	}
	stat.FMoney = fmt.Sprintf("%.2f", float64(stat.SMoney)/100)
	stat.FCashOut = fmt.Sprintf("%.2f", float64(stat.SCashOut)/100)
	stat.FWinMoney = fmt.Sprintf("%.2f", float64(int(stat.SDiamond)+stat.SCashOut-stat.SMoney)/100)
}

// GetLHDStatDates lhd汇总明细
func (s *gameStatsService) GetLHDStatDates(m, m2, m3 bson.M, startDate, endDate string) (stats []*entity.LHDPlayerStat, err error) {
	// 用户列表筛选
	var filterDatePlayerIds []string
	if len(m2) > 0 || len(m3) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				// "_id":        bson.M{"userid":"$userid", "date", "$date"},
				"_id":        "$_id",
				"bets":       bson.M{"$sum": "$bets"},
				"bet_rounds": bson.M{"$sum": "$bet_rounds"},
				"cash":       bson.M{"$sum": "$cash"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				"cash": "$cash",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = LHDPlayerStats.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetLHDStatDates error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			id := r["_id"].(string)
			filterDatePlayerIds = append(filterDatePlayerIds, id)
			ids := strings.Split(id, "-") // id=date-userid
			if len(ids) >= 2 {
				userids = append(userids, ids[1])
			}
		}
		if len(filterDatePlayerIds) == 0 { // 未匹配到
			return
		}
		if len(m3) > 0 { // 流失天数，局均打码量
			if len(userids) == 0 { // 未匹配到
				return
			}
			now := utils.BsonNow()
			m3_2 := []bson.M{
				{"$match": bson.M{"_id": bson.M{"$in": userids}}},
				{"$project": bson.M{
					"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
					"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
				}},
				{"$match": m3},
			}
			var r3_2 []bson.M
			err = PlayerUsers.Pipe(m3_2).All(&r3_2)
			if err != nil {
				beego.Error("GetCrashStatPlayers error32:", err)
			} else if len(r3_2) == 0 { // 未匹配到
				return
			}
			var filterId2 []string
			for _, id := range filterDatePlayerIds {
				for _, r := range r3_2 {
					userid := r["_id"].(string)
					if strings.HasSuffix(id, "-"+userid) {
						filterId2 = append(filterId2, id)
						break
					}
				}
			}
			if len(filterId2) == 0 { // 未匹配到
				return
			}
			filterDatePlayerIds = filterId2
		}
	}
	if len(filterDatePlayerIds) > 0 {
		m["_id"] = bson.M{"$in": filterDatePlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":                "$date_str",
			"date":               bson.M{"$min": "$date"},
			"players":            bson.M{"$sum": 1},
			"new_players":        bson.M{"$sum": bson.M{"$cond": bson.M{"if": "$new_reg", "then": 1, "else": 0}}},
			"old_players":        bson.M{"$sum": bson.M{"$cond": bson.M{"if": "$new_reg", "then": 0, "else": 1}}},
			"all_rounds":         bson.M{"$sum": "$all_rounds"},
			"game_times":         bson.M{"$sum": "$game_times"},
			"bets":               bson.M{"$sum": "$bets"},
			"bet_rounds":         bson.M{"$sum": "$bet_rounds"},
			"win_rounds":         bson.M{"$sum": "$win_rounds"},
			"lose_rounds":        bson.M{"$sum": "$lose_rounds"},
			"tie_rounds":         bson.M{"$sum": "$tie_rounds"},
			"win_bets":           bson.M{"$sum": "$win_bets"},
			"lose_bets":          bson.M{"$sum": "$lose_bets"},
			"tie_bets":           bson.M{"$sum": "$tie_bets"},
			"wins":               bson.M{"$sum": "$wins"},
			"loses":              bson.M{"$sum": "$loses"},
			"cash":               bson.M{"$sum": "$cash"},
			"dragons":            bson.M{"$sum": "$dragons"},
			"tigers":             bson.M{"$sum": "$tigers"},
			"ties":               bson.M{"$sum": "$ties"},
			"player_dragons":     bson.M{"$sum": "$player_dragons"},
			"player_tigers":      bson.M{"$sum": "$player_tigers"},
			"player_ties":        bson.M{"$sum": "$player_ties"},
			"player_win_dragons": bson.M{"$sum": "$player_win_dragons"},
			"player_win_tigers":  bson.M{"$sum": "$player_win_tigers"},
			"player_win_ties":    bson.M{"$sum": "$player_win_ties"},
			"player_multis":      bson.M{"$sum": "$player_multis"},
			"round_bet_avg":      bson.M{"$sum": "$round_bet_avg"},
			"money":              bson.M{"$sum": "$money"},
			"observe_rounds":     bson.M{"$sum": "$observe_rounds"},
			"all_rounds_list":    bson.M{"$push": "$bet_rounds"},
		}},
		{"$sort": bson.M{"date": -1}},
	}
	err = LHDPlayerStats.Pipe(m1).All(&stats)
	if err != nil {
		return
	}

	// 查询日活数据
	var dateLoginedUsers = make(map[int64]int)
	m9 := []bson.M{
		{"$match": bson.M{"date": m["date"]}},
		{"$project": bson.M{"date": "$date", "logined_users": "$logined_users"}},
	}
	var r9 []bson.M
	err = PddStats.Pipe(m9).All(&r9)
	if err != nil {
		beego.Error("GetCrashStatDates logined_users error: ", err)
	} else {
		for _, r := range r9 {
			date := r["date"].(int64)
			logined_users := r["logined_users"].(int)
			dateLoginedUsers[date] = logined_users
		}
	}

	mark := "--"
	summary := &entity.LHDPlayerStat{
		DateStr: fmt.Sprintf("%s-%s总汇", startDate, endDate),
		Id:      mark,
	}
	lhdStatDatesMapping(summary, stats, dateLoginedUsers)
	lhdStatDatesSummaryMapping(summary)
	stats = append([]*entity.LHDPlayerStat{summary}, stats...)
	return
}

func lhdStatDatesMapping(summary *entity.LHDPlayerStat, stats []*entity.LHDPlayerStat, dateLoginedUsers map[int64]int) {
	for _, stat := range stats {
		dateStr := stat.Id
		date, _ := utils.Unix(dateStr)
		stat.Date = date
		stat.DateStr = dateStr
		loginedUsers := dateLoginedUsers[stat.Date] // 日活

		stat.DateStr = utils.Stamp2Time(stat.Date).Format("2006-01-02")
		stat.FLoginedUsers = loginedUsers

		summary.FLoginedUsers += stat.FLoginedUsers
		summary.FLiveDays += stat.FLiveDays
		summary.FLoseDays += stat.FLoseDays
		summary.AllRounds += stat.AllRounds
		summary.GameTimes += stat.GameTimes
		summary.Bets += stat.Bets
		summary.BetRounds += stat.BetRounds
		summary.WinRounds += stat.WinRounds
		summary.LoseRounds += stat.LoseRounds
		summary.TieRounds += stat.TieRounds
		summary.WinBets += stat.WinBets
		summary.LoseBets += stat.LoseBets
		summary.TieBets += stat.TieBets
		summary.Wins += stat.Wins
		summary.Loses += stat.Loses
		summary.Cash += stat.Cash
		summary.Dragons += stat.Dragons
		summary.Tigers += stat.Tigers
		summary.Ties += stat.Ties
		summary.PlayerDragons += stat.PlayerDragons
		summary.PlayerTigers += stat.PlayerTigers
		summary.PlayerTies += stat.PlayerTies
		summary.PlayerWinDragons += stat.PlayerWinDragons
		summary.PlayerWinTigers += stat.PlayerWinTigers
		summary.PlayerWinTies += stat.PlayerWinTies
		summary.PlayerMultis += stat.PlayerMultis
		// summary.RoundBetAvg += stat.RoundBetAvg

		summary.ObserveRounds += stat.ObserveRounds
		summary.Players += stat.Players
		summary.NewPlayers += stat.NewPlayers
		summary.OldPlayers += stat.OldPlayers
		summary.AllRoundsList = append(summary.AllRoundsList, stat.AllRoundsList...)

		// 玩 crash 人数占日活比
		if stat.FLoginedUsers > 0 {
			stat.FPlayerRate = fmt.Sprintf("%.2f", float64(stat.Players)/float64(stat.FLoginedUsers))
		}

		// 查询当天注册的新用户
		startTime := utils.Str2Time(fmt.Sprintf("%s 00:00:00", dateStr), Location())
		endTime := utils.Str2Time(fmt.Sprintf("%s 23:59:59", dateStr), Location())
		m7 := []bson.M{
			{"$match": bson.M{
				"ctime":            bson.M{"$gte": startTime, "$lte": endTime},
				"robot":            false,
				"simulation_robot": false,
			}},
			{"$group": bson.M{
				"_id":   nil,
				"total": bson.M{"$sum": 1},
			}},
		}
		var r7 []bson.M
		err := PlayerUsers.Pipe(m7).All(&r7)
		if err != nil {
			beego.Error("查询当天注册的新用户 error: ", err)
		} else if len(r7) > 0 {
			stat.FLoginedNewUsers = r7[0]["total"].(int)
		}
		stat.FLoginedOldUsers = stat.FLoginedUsers - stat.FLoginedNewUsers
		if stat.FLoginedOldUsers < 0 {
			stat.FLoginedOldUsers = 0
		}
		summary.FLoginedNewUsers += stat.FLoginedNewUsers
		summary.FLoginedOldUsers += stat.FLoginedOldUsers
		// FNewPlayerRate, FOldPlayerRate 占新老玩家日活
		if stat.FLoginedNewUsers > 0 {
			stat.FNewPlayerRate = fmt.Sprintf("%.2f", float64(stat.NewPlayers)/float64(stat.FLoginedNewUsers))
		}
		if stat.FLoginedOldUsers > 0 {
			stat.FOldPlayerRate = fmt.Sprintf("%.2f", float64(stat.OldPlayers)/float64(stat.FLoginedOldUsers))
		}

		if stat.Players > 0 {
			stat.FPlayerRoundsAvg = fmt.Sprintf("%.2f", float64(stat.BetRounds)/float64(stat.Players))
		}
		// 局数中位数
		if len(stat.AllRoundsList) > 0 {
			var rounds = stat.AllRoundsList
			sort.Slice(rounds, func(i, j int) bool {
				return rounds[i] < rounds[j]
			})
			length := len(rounds)
			if length > 0 {
				if length%2 == 1 {
					stat.FPlayerRoundsMedian = rounds[length/2]
				} else {
					// 中间两位算平均
					stat.FPlayerRoundsMedian = (rounds[length/2] + rounds[length/2-1]) / 2
				}
			}
		}
		// stat.FGameTimes = fmt.Sprintf("%.2f", float64(stat.GameTimes)/60)
		if stat.BetRounds > 0 {
			stat.FGameTimesAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.BetRounds)/60)
		}
		if stat.Players > 0 {
			stat.FGameTimesPlayerAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.Players)/60)
		}

		stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
		if stat.Players > 0 {
			stat.FPlayerBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.Players)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.BetRounds)/100.0)
			// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
		}
		stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
		stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
		stat.FTieBets = fmt.Sprintf("%.2f", float64(stat.TieBets)/100.0)
		if stat.BetRounds > 0 {
			stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
		}
		if stat.Bets > 0 {
			// stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
			stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
		}
		if stat.WinRounds > 0 {
			stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
		}
		if stat.TieRounds > 0 {
			stat.FTieBetsAvg = fmt.Sprintf("%.2f", float64(stat.TieBets)/float64(stat.TieRounds)/100.0)
		}
		stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
		stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
		stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)

		if stat.WinRounds > 0 {
			stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
		}
		// 押龙/虎/和率
		if stat.BetRounds > 0 {
			stat.PlayerDragonsRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerDragons)/float64(stat.BetRounds)*100)
			stat.PlayerTigersRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerTigers)/float64(stat.BetRounds)*100)
			stat.PlayerTiesRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerTies)/float64(stat.BetRounds)*100)
		}
		if stat.Dragons > 0 {
			stat.PlayerDragonsWinRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinDragons)/float64(stat.Dragons)*100)
		}
		if stat.Tigers > 0 {
			stat.PlayerTigersWinRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinTigers)/float64(stat.Tigers)*100)
		}
		if stat.Ties > 0 {
			stat.PlayerTiesWinRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinTies)/float64(stat.Ties)*100)
		}
		// 多门率
		if stat.BetRounds > 0 {
			stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerMultis)/float64(stat.BetRounds)*100)
		}

		// 观察局
		if stat.BetRounds > 0 {
			stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
		}
		if stat.BetRounds > 0 {
			stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
		}

	}
}

func lhdStatDatesSummaryMapping(stat *entity.LHDPlayerStat) {
	// 玩 lhd 人数占日活比
	if stat.FLoginedUsers > 0 {
		stat.FPlayerRate = fmt.Sprintf("%.2f", float64(stat.Players)/float64(stat.FLoginedUsers))
	}

	// FNewPlayerRate, FOldPlayerRate 占新老玩家日活
	if stat.FLoginedNewUsers > 0 {
		stat.FNewPlayerRate = fmt.Sprintf("%.2f", float64(stat.NewPlayers)/float64(stat.FLoginedNewUsers))
	}
	if stat.FLoginedOldUsers > 0 {
		stat.FOldPlayerRate = fmt.Sprintf("%.2f", float64(stat.OldPlayers)/float64(stat.FLoginedOldUsers))
	}

	if stat.Players > 0 {
		stat.FPlayerRoundsAvg = fmt.Sprintf("%.2f", float64(stat.BetRounds)/float64(stat.Players))
	}
	// 局数中位数
	if len(stat.AllRoundsList) > 0 {
		var rounds = stat.AllRoundsList
		sort.Slice(rounds, func(i, j int) bool {
			return rounds[i] < rounds[j]
		})
		length := len(rounds)
		if length > 0 {
			if length%2 == 1 {
				stat.FPlayerRoundsMedian = rounds[length/2]
			} else {
				// 中间两位算平均
				stat.FPlayerRoundsMedian = (rounds[length/2] + rounds[length/2-1]) / 2
			}
		}
	}
	// stat.FGameTimes = fmt.Sprintf("%.2f", float64(stat.GameTimes)/60)
	if stat.BetRounds > 0 {
		stat.FGameTimesAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.BetRounds)/60)
	}
	if stat.Players > 0 {
		stat.FGameTimesPlayerAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.Players)/60)
	}

	if stat.BetRounds > 0 {
		stat.FGameTimesAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.BetRounds)/60)
	}
	if stat.Players > 0 {
		stat.FGameTimesPlayerAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.Players)/60)
	}

	stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
	if stat.Players > 0 {
		stat.FPlayerBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.Players)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.BetRounds)/100.0)
		// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
	}
	stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
	stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
	stat.FTieBets = fmt.Sprintf("%.2f", float64(stat.TieBets)/100.0)

	if stat.BetRounds > 0 {
		stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
	}
	if stat.Bets > 0 {
		// stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
		stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
	}
	if stat.WinRounds > 0 {
		stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
	}
	stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
	stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
	stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)

	if stat.WinRounds > 0 {
		stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
	}
	if stat.TieRounds > 0 {
		stat.FTieBetsAvg = fmt.Sprintf("%.2f", float64(stat.TieBets)/float64(stat.TieRounds)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
	}
	// 押龙/虎/和率
	if stat.BetRounds > 0 {
		stat.PlayerDragonsRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerDragons)/float64(stat.BetRounds)*100)
		stat.PlayerTigersRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerTigers)/float64(stat.BetRounds)*100)
		stat.PlayerTiesRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerTies)/float64(stat.BetRounds)*100)
	}
	if stat.Dragons > 0 {
		stat.PlayerDragonsWinRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinDragons)/float64(stat.Dragons)*100)
	}
	if stat.Tigers > 0 {
		stat.PlayerTigersWinRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinTigers)/float64(stat.Tigers)*100)
	}
	if stat.Ties > 0 {
		stat.PlayerTiesWinRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinTies)/float64(stat.Ties)*100)
	}
	// 多门率
	if stat.BetRounds > 0 {
		stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerMultis)/float64(stat.BetRounds)*100)
	}

	// 观察局
	if stat.BetRounds > 0 {
		stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
	}
	if stat.BetRounds > 0 {
		stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
	}
}

// GetLHDStatPlayers lhd新想事成
func (s *gameStatsService) GetLHDXxscStat(page, pageSize int, m, m2, m3 bson.M, startDate, endDate string) (
	stats []*entity.LHDXxscStat, count int, err error,
) {
	// LHDXxscStat
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)

	// 用户列表筛选
	var filterPlayerIds []string
	if len(m2) > 0 || len(m3) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				"_id":        "$userid",
				"bets":       bson.M{"$sum": "$bets"},
				"bet_rounds": bson.M{"$sum": "$bet_rounds"},
				"cash":       bson.M{"$sum": "$cash"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				"cash": "$cash",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = LHDPlayerStats.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("LHDPlayerStats error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			userid := r["_id"].(string)
			userids = append(userids, userid)
		}
		if len(userids) == 0 { // 未匹配到
			return
		}
		now := utils.BsonNow()
		m3_2 := []bson.M{
			{"$match": bson.M{"_id": bson.M{"$in": userids}}},
			{"$project": bson.M{
				"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
				"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
			}},
			{"$match": m3},
		}
		var r3_2 []bson.M
		err = PlayerUsers.Pipe(m3_2).All(&r3_2)
		if err != nil {
			beego.Error("lhd PlayerUsers error32:", err)
		} else if len(r3_2) == 0 { // 未匹配到
			return
		}
		for _, r := range r3_2 {
			userid := r["_id"].(string)
			filterPlayerIds = append(filterPlayerIds, userid)
		}
	}

	if len(filterPlayerIds) > 0 {
		m["userid"] = bson.M{"$in": filterPlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":  "$userid",
			"date": bson.M{"$max": "$date"},
			// "date_str":                bson.M{"$sum": "$date_str"},
			"strategy_players":        bson.M{"$max": "$strategy_players"},
			"strategy_rounds":         bson.M{"$sum": "$strategy_rounds"},
			"strategy_disturb_rounds": bson.M{"$sum": "$strategy_disturb_rounds"},
			"strategy_multi_rounds":   bson.M{"$sum": "$strategy_multi_rounds"},
			"strategy_bets":           bson.M{"$sum": "$strategy_bets"},
			"strategy_win_rounds":     bson.M{"$sum": "$strategy_win_rounds"},
			"strategy_wins":           bson.M{"$sum": "$strategy_wins"},
			"strategy_loses":          bson.M{"$sum": "$strategy_loses"},
		}},
		{"$project": bson.M{
			"_id": "$_id", "date": "$date", "userid": "$userid", "strategy_players": "$strategy_players", "strategy_rounds": "$strategy_rounds", "strategy_disturb_rounds": "$strategy_disturb_rounds", "strategy_multi_rounds": "$strategy_multi_rounds", "strategy_bets": "$strategy_bets", "strategy_win_rounds": "$strategy_win_rounds", "strategy_wins": "$strategy_wins", "strategy_loses": "$strategy_loses",
		}},
		{"$match": m2},
		{"$count": "count"},
	}
	// 查总条数
	var rCount []bson.M
	err = LHDXxscStats.Pipe(m1).All(&rCount)
	if err != nil {
		return
	} else if len(rCount) > 0 {
		count = rCount[0]["count"].(int)
	}

	// 查列表数据
	m1 = m1[:len(m1)-1]
	m1 = append(m1, []bson.M{
		{"$sort": bson.M{"date": -1}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}...)
	err = LHDXxscStats.Pipe(m1).All(&stats)
	if err != nil {
		return
	}
	lhdXxscStatMapping(stats)

	// 查总汇数据
	mark := "--"
	summary := &entity.LHDXxscStat{}
	m1 = m1[:len(m1)-3]
	m1[1]["$group"].(bson.M)["_id"] = nil
	delete(m1[1]["$group"].(bson.M), "date")
	m1[1]["$group"].(bson.M)["userids"] = bson.M{"$addToSet": "$userid"}
	m1[1]["$group"].(bson.M)["strategy_players"] = bson.M{"$sum": "$strategy_players"} // 触发策略人数总汇
	m1[2]["$project"].(bson.M)["userids"] = "$userids"
	str, err := json.Marshal(m1)
	fmt.Println(err, string(str))
	var summarys []*entity.LHDXxscStat
	err = LHDXxscStats.Pipe(m1).All(&summarys)
	if err != nil {
		beego.Error("lhd summary error:", err)
		return
	} else if len(summarys) > 0 {
		summary = summarys[0]
		summary.No = fmt.Sprintf("%s-%s总汇", startDate, endDate)
		summary.Id = mark
		summary.FNickname = mark
		summary.FUserType = mark
		summary.FRegistArea = mark
		summary.FCtime = mark
		summary.FLoginTime = mark
		lhdXxscStatSummaryMapping(summary)
	}
	if len(summarys) > 0 {
		stats = append([]*entity.LHDXxscStat{summary}, stats...)
	}
	return
}

func lhdXxscStatMapping(stats []*entity.LHDXxscStat) {
	var playerIds []string
	// var playerStats = make(map[string]*entity.CrashPlayerStat, len(stats))
	for _, stat := range stats {
		playerIds = append(playerIds, stat.Id)
		// playerStats[stat.Id] = stat
	}
	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": playerIds}}},
		{"$project": bson.M{
			"_id":          "$_id",
			"nickname":     "$nickname",
			"money":        "$money",
			"cash_out":     "$cash_out",
			"diamond":      "$diamond",
			"ctime":        "$ctime",
			"login_time":   "$login_time",
			"regist_area":  "$regist_area", //ab测试 0:A 1:B 2:C
			"state":        "$state",
			"lhd_strategy": "$lhd_strategy",
		}},
	}

	var r2 []*entity.PlayerUser
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("lhd users error:", err)
		return
	}
	var r2Map = make(map[string]*entity.PlayerUser, len(r2))
	for _, r := range r2 {
		userid := r.Userid
		r2Map[userid] = r
	}
	now := utils.BsonNow()
	for i, stat := range stats {
		r, ok := r2Map[stat.Id]
		if !ok {
			continue
		}
		// 玩家昵称	玩家类型 账号类型 注册时间 最后登陆时间 存活天数 流失天数
		nickname := r.Nickname
		money := r.Money
		ctime := r.Ctime
		login_time := r.LoginTime
		regist_area := r.RegistArea
		state := r.State

		// 将日期转成印度时间
		login_time, _ = ConvertToIndiaTime(login_time.Unix())
		ctime, _ = ConvertToIndiaTime(ctime.Unix())

		stat.No = strconv.Itoa(i + 1)
		stat.FNickname = nickname
		stat.FCtime = ctime.Format(utils.FORMAT)
		stat.FLoginTime = login_time.Format(utils.FORMAT)
		stat.FLiveDays = int(math.Ceil(login_time.Sub(ctime).Hours() / 24))
		stat.FLoseDays = int(math.Floor(now.Sub(login_time).Hours() / 24))

		switch regist_area {
		case 0:
			stat.FRegistArea = "A类"
		case 1:
			stat.FRegistArea = "B类"
		case 2:
			stat.FRegistArea = "C类"
		}

		switch state {
		case 1:
			stat.FUserType = "新手"
		case 2:
			if money == 0 {
				stat.FUserType = "零充"
			} else if money >= 20000 && money <= 99999 {
				stat.FUserType = "普充"
			} else if money >= 100000 && money <= 499999 {
				stat.FUserType = "小R"
			} else if money >= 500000 && money <= 999999 {
				stat.FUserType = "中R"
			} else if money >= 1000000 && money <= 9999999 {
				stat.FUserType = "大R"
			} else if money >= 10000000 {
				stat.FUserType = "超大R"
			}
		case 3:
			stat.FUserType = "平民"
		case 4:
			stat.FUserType = "泡沫"
		}

		// stat.StrategyTimes = int32(r.LHDXXSCMaxTriggerTimes)
		// stat.StrategyValidTimes = int32(r.LHDXXSCTriggerTimes)
		stat.StrategyTimes = int32(r.LHDStrategy.XXSC.MaxTriggerTimes)
		stat.StrategyValidTimes = int32(r.LHDStrategy.XXSC.TriggerTimes)
		stat.StrategyCash = stat.StrategyWins + stat.StrategyLoses // 赢-输
		if stat.StrategyTimes > 0 {
			stat.StrategyValidTimesRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyValidTimes)/float64(stat.StrategyTimes)*100.0)
		}
		if stat.StrategyPlayers > 0 {
			stat.StrategyValidTimesAvg = fmt.Sprintf("%.2f", float64(stat.StrategyValidTimes)/float64(stat.StrategyPlayers))
		}
		if stat.StrategyTimes > 0 {
			stat.StrategyRoundAvg = fmt.Sprintf("%.2f", float64(stat.StrategyRounds)/float64(stat.StrategyTimes))
		}
		if stat.StrategyValidTimes > 0 {
			stat.StrategyValidRoundAvg = fmt.Sprintf("%.2f", float64(stat.StrategyRounds)/float64(stat.StrategyValidTimes))
		}
		if stat.StrategyPlayers > 0 {
			stat.StrategyDisturbRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyDisturbRounds)/float64(stat.StrategyRounds)*100.0)
		}
		stat.StrategyMultiRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyMultiRounds)/float64(stat.StrategyRounds)*100.0)
		stat.FStrategyBets = fmt.Sprintf("%.2f", float64(stat.StrategyBets)/100.0)
		if stat.StrategyRounds > 0 {
			stat.FStrategyBetRoundsAvg = fmt.Sprintf("%.2f", float64(stat.StrategyBets)/float64(stat.StrategyRounds)/100.0)
		}

		stat.FStrategyWins = fmt.Sprintf("%.2f", float64(stat.StrategyWins)/100.0)
		stat.FStrategyLoses = fmt.Sprintf("%.2f", float64(stat.StrategyLoses)/100.0)
		stat.FStrategyCash = fmt.Sprintf("%.2f", float64(stat.StrategyWins+stat.StrategyLoses)/100.0)
		if stat.StrategyRounds > 0 {
			stat.FStrategyWinRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyWinRounds)/float64(stat.StrategyRounds)*100.0)
		}
		if -stat.StrategyLoses != 0 {
			stat.FStrategyRebateRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyWins)/float64(-stat.StrategyLoses)*100.0)
		} else {
			stat.FStrategyRebateRate = "未输过"
		}
	}
}

func lhdXxscStatSummaryMapping(stat *entity.LHDXxscStat) {
	if len(stat.Userids) == 0 {
		return
	}
	// 查用户策略触发次数
	m1 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": stat.Userids}}},
		{"$group": bson.M{
			"_id":               nil,
			"trigger_times":     bson.M{"$sum": "$lhd_strategy.xxsc.trigger_times"},
			"max_trigger_times": bson.M{"$sum": "$lhd_strategy.xxsc.max_trigger_times"},
		}},
	}
	var r1 bson.M
	err := PlayerUsers.Pipe(m1).One(&r1)
	if err != nil {
		beego.Error("lhdXxscStatSummaryMapping err:", err)
		return
	}
	stat.StrategyTimes = int32(r1["max_trigger_times"].(int))
	stat.StrategyValidTimes = int32(r1["trigger_times"].(int))
	stat.StrategyPlayers = int32(len(stat.Userids))
	// for _, s := range stats {
	// 	stat.StrategyTimes += s.StrategyTimes
	// 	stat.StrategyValidTimes += s.StrategyValidTimes
	// }

	// stat.StrategyTimes = int32(r.LHDStrategy.XXSC.MaxTriggerTimes)
	// stat.StrategyValidTimes = int32(r.LHDStrategy.XXSC.TriggerTimes)
	stat.StrategyCash = stat.StrategyWins + stat.StrategyLoses // 赢-输
	if stat.StrategyTimes > 0 {
		stat.StrategyValidTimesRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyValidTimes)/float64(stat.StrategyTimes)*100.0)
	}
	if stat.StrategyPlayers > 0 {
		stat.StrategyValidTimesAvg = fmt.Sprintf("%.2f", float64(stat.StrategyValidTimes)/float64(stat.StrategyPlayers))
	}
	if stat.StrategyTimes > 0 {
		stat.StrategyRoundAvg = fmt.Sprintf("%.2f", float64(stat.StrategyRounds)/float64(stat.StrategyTimes))
	}
	if stat.StrategyValidTimes > 0 {
		stat.StrategyValidRoundAvg = fmt.Sprintf("%.2f", float64(stat.StrategyRounds)/float64(stat.StrategyValidTimes))
	}
	if stat.StrategyPlayers > 0 {
		stat.StrategyDisturbRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyDisturbRounds)/float64(stat.StrategyRounds)*100.0)
	}
	stat.StrategyMultiRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyMultiRounds)/float64(stat.StrategyRounds)*100.0)
	stat.FStrategyBets = fmt.Sprintf("%.2f", float64(stat.StrategyBets)/100.0)
	if stat.StrategyRounds > 0 {
		stat.FStrategyBetRoundsAvg = fmt.Sprintf("%.2f", float64(stat.StrategyBets)/float64(stat.StrategyRounds)/100.0)
	}

	stat.FStrategyWins = fmt.Sprintf("%.2f", float64(stat.StrategyWins)/100.0)
	stat.FStrategyLoses = fmt.Sprintf("%.2f", float64(stat.StrategyLoses)/100.0)
	stat.FStrategyCash = fmt.Sprintf("%.2f", float64(stat.StrategyWins+stat.StrategyLoses)/100.0)
	if stat.StrategyRounds > 0 {
		stat.FStrategyWinRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyWinRounds)/float64(stat.StrategyRounds)*100.0)
	}
	if -stat.StrategyLoses != 0 {
		stat.FStrategyRebateRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyWins)/float64(-stat.StrategyLoses)*100.0)
	} else {
		stat.FStrategyRebateRate = "未输过"
	}
}

// GetLHDQsbnStat lhd求死不能
func (s *gameStatsService) GetLHDQsbnStat(page, pageSize int, m, m2, m3 bson.M, startDate, endDate string) (
	stats []*entity.LHDQsbnStat, count int, err error,
) {
	// LHDXxscStat
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)

	// 用户列表筛选
	var filterPlayerIds []string
	if len(m2) > 0 || len(m3) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				"_id":        "$userid",
				"bets":       bson.M{"$sum": "$bets"},
				"bet_rounds": bson.M{"$sum": "$bet_rounds"},
				"cash":       bson.M{"$sum": "$cash"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				"cash": "$cash",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = LHDPlayerStats.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("LHDPlayerStats error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			userid := r["_id"].(string)
			userids = append(userids, userid)
		}
		if len(userids) == 0 { // 未匹配到
			return
		}
		now := utils.BsonNow()
		m3_2 := []bson.M{
			{"$match": bson.M{"_id": bson.M{"$in": userids}}},
			{"$project": bson.M{
				"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
				"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
			}},
			{"$match": m3},
		}
		var r3_2 []bson.M
		err = PlayerUsers.Pipe(m3_2).All(&r3_2)
		if err != nil {
			beego.Error("lhd PlayerUsers error32:", err)
		} else if len(r3_2) == 0 { // 未匹配到
			return
		}
		for _, r := range r3_2 {
			userid := r["_id"].(string)
			filterPlayerIds = append(filterPlayerIds, userid)
		}
	}

	if len(filterPlayerIds) > 0 {
		m["userid"] = bson.M{"$in": filterPlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":  "$userid",
			"date": bson.M{"$max": "$date"},
			// "date_str":                bson.M{"$sum": "$date_str"},
			"strategy_players":      bson.M{"$max": "$strategy_players"},
			"strategy_times":        bson.M{"$sum": "$strategy_times"},
			"all_in_rounds":         bson.M{"$sum": "$all_in_rounds"},
			"strategy_multi_rounds": bson.M{"$sum": "$strategy_multi_rounds"},
			"strategy_bets":         bson.M{"$sum": "$strategy_bets"},
			"strategy_wins":         bson.M{"$sum": "$strategy_wins"},
			"strategy_loses":        bson.M{"$sum": "$strategy_loses"},
			"before_back_rate":      bson.M{"$avg": "$before_back_rate"},
			"after_back_rate":       bson.M{"$avg": "$after_back_rate"},
		}},
		{"$project": bson.M{
			"_id":                   "$_id",
			"date":                  "$date",
			"strategy_players":      "$strategy_players",
			"strategy_times":        "$strategy_times",
			"all_in_rounds":         "$all_in_rounds",
			"strategy_multi_rounds": "$strategy_multi_rounds",
			"strategy_bets":         "$strategy_bets",
			"strategy_wins":         "$strategy_wins",
			"strategy_loses":        "$strategy_loses",
			"before_back_rate":      "$before_back_rate",
			"after_back_rate":       "$after_back_rate",
		}},
		{"$match": m2},
		{"$count": "count"},
	}
	// 查总条数
	var rCount []bson.M
	err = LHDQsbnStats.Pipe(m1).All(&rCount)
	if err != nil {
		return
	} else if len(rCount) > 0 {
		count = rCount[0]["count"].(int)
	}

	// 查列表数据
	m1 = m1[:len(m1)-1]
	m1 = append(m1, []bson.M{
		{"$sort": bson.M{"date": -1}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}...)
	err = LHDQsbnStats.Pipe(m1).All(&stats)
	if err != nil {
		return
	}
	lhdQsbnStatMapping(stats)

	// 查总汇数据
	mark := "--"
	summary := &entity.LHDQsbnStat{}
	m1 = m1[:len(m1)-3]
	m1[1]["$group"].(bson.M)["_id"] = nil
	delete(m1[1]["$group"].(bson.M), "date")
	m1[1]["$group"].(bson.M)["userids"] = bson.M{"$addToSet": "$userid"}
	m1[1]["$group"].(bson.M)["strategy_players"] = bson.M{"$sum": "$strategy_players"} // 触发策略人数总汇
	m1[2]["$project"].(bson.M)["userids"] = "$userids"
	str, err := json.Marshal(m1)
	fmt.Println(err, string(str))
	var summarys []*entity.LHDQsbnStat
	err = LHDQsbnStats.Pipe(m1).All(&summarys)
	if err != nil {
		beego.Error("lhd summary error:", err)
		return
	} else if len(summarys) > 0 {
		summary = summarys[0]
		summary.No = fmt.Sprintf("%s-%s总汇", startDate, endDate)
		summary.Id = mark
		summary.FNickname = mark
		summary.FUserType = mark
		summary.FRegistArea = mark
		summary.FCtime = mark
		summary.FLoginTime = mark
		summary.FBackRate = mark
		lhdQsbnStatSummaryMapping(summary)
	}
	if len(summarys) > 0 {
		stats = append([]*entity.LHDQsbnStat{summary}, stats...)
	}
	return
}

func lhdQsbnStatMapping(stats []*entity.LHDQsbnStat) {
	var playerIds []string
	// var playerStats = make(map[string]*entity.CrashPlayerStat, len(stats))
	for _, stat := range stats {
		playerIds = append(playerIds, stat.Id)
		// playerStats[stat.Id] = stat
	}
	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": playerIds}}},
		{"$project": bson.M{
			"_id":          "$_id",
			"nickname":     "$nickname",
			"money":        "$money",
			"cash_out":     "$cash_out",
			"diamond":      "$diamond",
			"ctime":        "$ctime",
			"login_time":   "$login_time",
			"regist_area":  "$regist_area", //ab测试 0:A 1:B 2:C
			"state":        "$state",
			"lhd_strategy": "$lhd_strategy",
		}},
	}

	var r2 []*entity.PlayerUser
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("lhd users error:", err)
		return
	}
	var r2Map = make(map[string]*entity.PlayerUser, len(r2))
	for _, r := range r2 {
		userid := r.Userid
		r2Map[userid] = r
	}
	now := utils.BsonNow()
	for i, stat := range stats {
		r, ok := r2Map[stat.Id]
		if !ok {
			continue
		}
		// 玩家昵称	玩家类型 账号类型 注册时间 最后登陆时间 存活天数 流失天数
		nickname := r.Nickname
		money := r.Money
		ctime := r.Ctime
		login_time := r.LoginTime
		regist_area := r.RegistArea
		state := r.State

		// 将日期转成印度时间
		login_time, _ = ConvertToIndiaTime(login_time.Unix())
		ctime, _ = ConvertToIndiaTime(ctime.Unix())

		stat.No = strconv.Itoa(i + 1)
		stat.FNickname = nickname
		stat.FCtime = ctime.Format(utils.FORMAT)
		stat.FLoginTime = login_time.Format(utils.FORMAT)
		stat.FLiveDays = int(math.Ceil(login_time.Sub(ctime).Hours() / 24))
		stat.FLoseDays = int(math.Floor(now.Sub(login_time).Hours() / 24))

		switch regist_area {
		case 0:
			stat.FRegistArea = "A类"
		case 1:
			stat.FRegistArea = "B类"
		case 2:
			stat.FRegistArea = "C类"
		}

		switch state {
		case 1:
			stat.FUserType = "新手"
		case 2:
			if money == 0 {
				stat.FUserType = "零充"
			} else if money >= 20000 && money <= 99999 {
				stat.FUserType = "普充"
			} else if money >= 100000 && money <= 499999 {
				stat.FUserType = "小R"
			} else if money >= 500000 && money <= 999999 {
				stat.FUserType = "中R"
			} else if money >= 1000000 && money <= 9999999 {
				stat.FUserType = "大R"
			} else if money >= 10000000 {
				stat.FUserType = "超大R"
			}
		case 3:
			stat.FUserType = "平民"
		case 4:
			stat.FUserType = "泡沫"
		}
		if stat.StrategyPlayers > 0 {
			stat.StrategyPlayersAvg = fmt.Sprintf("%.2f", float64(stat.StrategyTimes)/float64(stat.StrategyPlayers))
		}
		if stat.AllInRounds > 0 {
			stat.StrategyRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyTimes)/float64(stat.AllInRounds)*100.0)
		}
		if stat.StrategyTimes > 0 {
			stat.StrategyMultiRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyMultiRounds)/float64(stat.StrategyTimes)*100.0)
		}
		stat.FStrategyBets = fmt.Sprintf("%.2f", float64(stat.StrategyBets)/100.0)
		if stat.StrategyTimes > 0 {
			stat.FStrategyBetsAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.StrategyBets, 100)/float64(stat.StrategyTimes))
		}
		stat.FStrategyWins = fmt.Sprintf("%.2f", float64(stat.StrategyWins)/100.0)
		stat.FStrategyLoses = fmt.Sprintf("%.2f", float64(stat.StrategyLoses)/100.0)
		stat.FStrategyCash = fmt.Sprintf("%.2f", float64(stat.StrategyWins+stat.StrategyLoses)/100.0)
		stat.FBeforeBackRate = fmt.Sprintf("%d", stat.BeforeBackRate)
		stat.FAfterBackRate = fmt.Sprintf("%d", stat.AfterBackRate)
		if r.Money > 0 {
			stat.FBackRate = fmt.Sprintf("%d", int((int64(r.CashOut)+r.Diamond)*10000/int64(r.Money)))
		}
	}
}

func lhdQsbnStatSummaryMapping(stat *entity.LHDQsbnStat) {
	stat.StrategyPlayers = int32(len(stat.Userids))
	if stat.StrategyPlayers > 0 {
		stat.StrategyPlayersAvg = fmt.Sprintf("%.2f", float64(stat.StrategyTimes)/float64(stat.StrategyPlayers))
	}
	if stat.AllInRounds > 0 {
		stat.StrategyRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyTimes)/float64(stat.AllInRounds)*100.0)
	}
	if stat.StrategyTimes > 0 {
		stat.StrategyMultiRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyMultiRounds)/float64(stat.StrategyTimes)*100.0)
	}
	stat.FStrategyBets = fmt.Sprintf("%.2f", float64(stat.StrategyBets)/100.0)
	if stat.StrategyTimes > 0 {
		stat.FStrategyBetsAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.StrategyBets, 100)/float64(stat.StrategyTimes))
	}
	stat.FStrategyWins = fmt.Sprintf("%.2f", float64(stat.StrategyWins)/100.0)
	stat.FStrategyLoses = fmt.Sprintf("%.2f", float64(stat.StrategyLoses)/100.0)
	stat.FStrategyCash = fmt.Sprintf("%.2f", float64(stat.StrategyWins+stat.StrategyLoses)/100.0)
	stat.FBeforeBackRate = fmt.Sprintf("%d", stat.BeforeBackRate)
	stat.FAfterBackRate = fmt.Sprintf("%d", stat.AfterBackRate)
}

/*
	lhd 龙狂有祸
*/
// GetLHDLkysStat lhd龙狂有祸
func (s *gameStatsService) GetLHDLkyhStat(page, pageSize int, m, m2, m3 bson.M, startDate, endDate string) (
	stats []*entity.LHDLkyhStat, count int, err error,
) {
	// LHDXxscStat
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)

	// 用户列表筛选
	var filterPlayerIds []string
	if len(m2) > 0 || len(m3) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				"_id":    "$userid",
				"bets":   bson.M{"$sum": "$bets"},
				"rounds": bson.M{"$sum": "$rounds"},
				"cash":   bson.M{"$sum": "$cash"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				"cash": "$cash",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = LHDLkyhStats.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetLHDLkyhStat error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			userid := r["_id"].(string)
			userids = append(userids, userid)
		}
		if len(userids) == 0 { // 未匹配到
			return
		}
		now := utils.BsonNow()
		m3_2 := []bson.M{
			{"$match": bson.M{"_id": bson.M{"$in": userids}}},
			{"$project": bson.M{
				"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
				"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
			}},
			{"$match": m3},
		}
		var r3_2 []bson.M
		err = PlayerUsers.Pipe(m3_2).All(&r3_2)
		if err != nil {
			beego.Error("lhd LkyhStat error32:", err)
		} else if len(r3_2) == 0 { // 未匹配到
			return
		}
		for _, r := range r3_2 {
			userid := r["_id"].(string)
			filterPlayerIds = append(filterPlayerIds, userid)
		}
	}
	if len(filterPlayerIds) > 0 {
		m["userid"] = bson.M{"$in": filterPlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":  "$userid",
			"date": bson.M{"$max": "$date"},
			// "date_str":                bson.M{"$sum": "$date_str"},
			"bstimes":               bson.M{"$max": "$bstimes"},
			"bs_days":               bson.M{"$sum": "$bs_days"},
			"trigger_times":         bson.M{"$sum": "$trigger_times"},
			"tz":                    bson.M{"$max": "$tz"},
			"nz":                    bson.M{"$max": "$nz"},
			"bs_bets":               bson.M{"$sum": "$bs_bets"},
			"yz_number":             bson.M{"$sum": "$yz_number"},
			"yz_rounds":             bson.M{"$sum": "$yz_rounds"},
			"yz_bets":               bson.M{"$sum": "$yz_bets"},
			"rz":                    bson.M{"$sum": "$rz"},
			"strategy_multi_rounds": bson.M{"$sum": "$strategy_multi_rounds"},
			"bets":                  bson.M{"$sum": "$bets"},
			"rounds":                bson.M{"$sum": "$rounds"},
			"win_bets":              bson.M{"$sum": "$win_bets"},
			"lose_bets":             bson.M{"$sum": "$lose_bets"},
			"strategy_bets":         bson.M{"$sum": "$strategy_bets"},
			"strategy_wins":         bson.M{"$sum": "$strategy_wins"},
			"strategy_loses":        bson.M{"$sum": "$strategy_loses"},
			"strategy_win_rounds":   bson.M{"$sum": "$strategy_win_rounds"},
			"strategy_lose_rounds":  bson.M{"$sum": "$strategy_lose_rounds"},
			"strategy_times":        bson.M{"$sum": "$strategy_times"},
			"win_rounds":            bson.M{"$sum": "$win_rounds"},
			"lose_rounds":           bson.M{"$sum": "$lose_rounds"},
		}},
		{"$project": bson.M{
			"_id":                   "$_id",
			"date":                  "$date",
			"bstimes":               "$bstimes",
			"bs_days":               "$bs_days",
			"trigger_times":         "$trigger_times",
			"tz":                    "$tz",
			"nz":                    "$nz",
			"bs_bets":               "$bs_bets",
			"yz_number":             "$yz_number",
			"yz_rounds":             "$yz_rounds",
			"yz_bets":               "$yz_bets",
			"rz":                    "$rz",
			"strategy_multi_rounds": "$strategy_multi_rounds",
			"bets":                  "$bets",
			"rounds":                "$rounds",
			"win_bets":              "$win_bets",
			"lose_bets":             "$lose_bets",
			"strategy_bets":         "$strategy_bets",
			"strategy_wins":         "$strategy_wins",
			"strategy_loses":        "$strategy_loses",
			"strategy_win_rounds":   "$strategy_win_rounds",
			"strategy_lose_rounds":  "$strategy_lose_rounds",
			"strategy_times":        "$strategy_times",
			"win_rounds":            "$win_rounds",
			"lose_rounds":           "$lose_rounds",
		}},
		{"$match": m2},
		{"$count": "count"},
	}
	// 查总条数
	var rCount []bson.M
	err = LHDLkyhStats.Pipe(m1).All(&rCount)
	if err != nil {
		return
	} else if len(rCount) > 0 {
		count = rCount[0]["count"].(int)
	}

	// 查列表数据
	m1 = m1[:len(m1)-1]
	m1 = append(m1, []bson.M{
		{"$sort": bson.M{"date": -1}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}...)
	err = LHDLkyhStats.Pipe(m1).All(&stats)
	if err != nil {
		return
	}
	lhdLkyhStatMapping(stats, m)

	// 查总汇数据
	mark := "--"
	summary := &entity.LHDLkyhStat{}
	m1 = m1[:len(m1)-3]
	m1[1]["$group"].(bson.M)["_id"] = nil
	delete(m1[1]["$group"].(bson.M), "date")
	m1[1]["$group"].(bson.M)["userids"] = bson.M{"$addToSet": "$userid"}
	m1[1]["$group"].(bson.M)["bstimes"] = bson.M{"$sum": "$bstimes"} // 触发策略人数总汇
	m1[1]["$group"].(bson.M)["tz"] = bson.M{"$sum": "$tz"}           // 触发策略人数总汇
	m1[1]["$group"].(bson.M)["nz"] = bson.M{"$sum": "$nz"}           // 触发策略人数总汇

	m1[2]["$project"].(bson.M)["userids"] = "$userids"
	str, err := json.Marshal(m1)
	fmt.Println(err, string(str))
	var summarys []*entity.LHDLkyhStat
	err = LHDLkyhStats.Pipe(m1).All(&summarys)
	if err != nil {
		beego.Error("lhd summary error:", err)
		return
	} else if len(summarys) > 0 {
		summary = summarys[0]
		summary.No = fmt.Sprintf("%s-%s总汇", startDate, endDate)
		summary.Id = mark
		summary.FNickname = mark
		summary.FUserType = mark
		summary.FRegistArea = mark
		summary.FCtime = mark
		summary.FLoginTime = mark
		// summary.FBackRate = mark
		lhdLkyhStatSummaryMapping(summary, m)
	}
	if len(summarys) > 0 {
		stats = append([]*entity.LHDLkyhStat{summary}, stats...)
	}
	return
}

func lhdLkyhStatMapping(stats []*entity.LHDLkyhStat, m bson.M) {
	var playerIds []string
	// var playerStats = make(map[string]*entity.CrashPlayerStat, len(stats))
	for _, stat := range stats {
		playerIds = append(playerIds, stat.Id)
		// playerStats[stat.Id] = stat
	}
	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": playerIds}}},
		{"$project": bson.M{
			"_id":          "$_id",
			"nickname":     "$nickname",
			"money":        "$money",
			"cash_out":     "$cash_out",
			"diamond":      "$diamond",
			"ctime":        "$ctime",
			"login_time":   "$login_time",
			"regist_area":  "$regist_area", //ab测试 0:A 1:B 2:C
			"state":        "$state",
			"lhd_strategy": "$lhd_strategy",
		}},
	}

	var r2 []*entity.PlayerUser
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("lhd users error:", err)
		return
	}
	var r2Map = make(map[string]*entity.PlayerUser, len(r2))
	for _, r := range r2 {
		userid := r.Userid
		r2Map[userid] = r
	}
	// 查询总数据
	var lhdlist []entity.LHDPlayerStat
	n := bson.M{}
	// 手动复制 m 到 n
	for k, v := range m {
		n[k] = v
	}
	n["userid"] = bson.M{"$in": playerIds}
	m3 := []bson.M{
		{"$match": n},
		{"$group": bson.M{
			"_id":        "$userid",
			"bet_rounds": bson.M{"$sum": "$bet_rounds"},
			"win_rounds": bson.M{"$sum": "$win_rounds"},
			"bets":       bson.M{"$sum": "$bets"},
			"win_bets":   bson.M{"$sum": "$win_bets"},
			"lose_bets":  bson.M{"$sum": "$lose_bets"},
			"wins":       bson.M{"$sum": "$wins"},
			"loses":      bson.M{"$sum": "$loses"},
		}},
	}
	LHDPlayerStats.Pipe(m3).All(&lhdlist)

	now := utils.BsonNow()
	for i, stat := range stats {
		r, ok := r2Map[stat.Id]
		if !ok {
			continue
		}
		// 玩家昵称	玩家类型 账号类型 注册时间 最后登陆时间 存活天数 流失天数
		nickname := r.Nickname
		money := r.Money
		ctime := r.Ctime
		login_time := r.LoginTime
		regist_area := r.RegistArea
		state := r.State

		// 将日期转成印度时间
		login_time, _ = ConvertToIndiaTime(login_time.Unix())
		ctime, _ = ConvertToIndiaTime(ctime.Unix())

		stat.No = strconv.Itoa(i + 1)
		stat.FNickname = nickname
		stat.FCtime = ctime.Format(utils.FORMAT)
		stat.FLoginTime = login_time.Format(utils.FORMAT)
		stat.FLiveDays = int(math.Ceil(login_time.Sub(ctime).Hours() / 24))
		stat.FLoseDays = int(math.Floor(now.Sub(login_time).Hours() / 24))

		switch regist_area {
		case 0:
			stat.FRegistArea = "A类"
		case 1:
			stat.FRegistArea = "B类"
		case 2:
			stat.FRegistArea = "C类"
		}

		switch state {
		case 1:
			stat.FUserType = "新手"
		case 2:
			if money == 0 {
				stat.FUserType = "零充"
			} else if money >= 20000 && money <= 99999 {
				stat.FUserType = "普充"
			} else if money >= 100000 && money <= 499999 {
				stat.FUserType = "小R"
			} else if money >= 500000 && money <= 999999 {
				stat.FUserType = "中R"
			} else if money >= 1000000 && money <= 9999999 {
				stat.FUserType = "大R"
			} else if money >= 10000000 {
				stat.FUserType = "超大R"
			}
		case 3:
			stat.FUserType = "平民"
		case 4:
			stat.FUserType = "泡沫"
		}
		if stat.BSDays != 0 {
			stat.BSNumberAvg = fmt.Sprintf("%.2f", float64(stat.BSTimes)/float64(stat.BSDays))
		}
		if stat.BSTimes != 0 {
			stat.BSRate = fmt.Sprintf("%.2f%%", float64(stat.TZ)/float64(stat.BSTimes)*100.0)
			stat.BSBetAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.BSBets, 100)/float64(stat.BSTimes))
		}
		if stat.RZ != 0 {
			stat.YZRate = fmt.Sprintf("%.2f%%", float64(stat.YZNumber)/float64(stat.RZ)*100.0)
		}
		if stat.YZRounds != 0 {
			stat.YZBetAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.YZBets, 100)/float64(stat.YZRounds))
		}
		if stat.StrategyTimes > 0 {
			stat.StrategyMultiRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyMultiRounds)/float64(stat.StrategyTimes)*100.0)
			stat.StrategyWinRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyWinRounds)/float64(stat.StrategyTimes)*100.0)
		}
		stat.FStrategyBets = fmt.Sprintf("%.2f", float64(stat.StrategyBets)/100.0)
		if stat.StrategyTimes > 0 {
			stat.FStrategyBetsAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.StrategyBets, 100)/float64(stat.StrategyTimes))
		}
		if -stat.StrategyLoses != 0 {
			stat.StrategyRewardRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.StrategyWins, 100)/ComputeFloat(-stat.StrategyLoses, 100))*100.0)
		}
		if len(lhdlist) > 0 {
			for _, c := range lhdlist {
				if c.Id == stat.Id {
					stat.Bets = c.Bets
					stat.Rounds = int64(c.BetRounds)
					stat.WinRounds = int64(c.WinRounds)
					stat.WinBets = c.Wins
					stat.LoseBets = c.Loses
					break
				}
			}
		}
		if stat.Rounds != 0 {
			stat.RoundBetAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.Bets, 100)/float64(stat.Rounds))
			stat.TotalWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.Rounds)*100.0)
			// fmt.Printf("===========lhd winRound=%d, round=%d\n", stat.WinRounds, stat.Rounds)
		}
		stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
		stat.FStrategyWins = fmt.Sprintf("%.2f", float64(stat.StrategyWins)/100.0)
		stat.FStrategyLoses = fmt.Sprintf("%.2f", float64(stat.StrategyLoses)/100.0)
		stat.FStrategyCash = fmt.Sprintf("%.2f", float64(stat.StrategyWins+stat.StrategyLoses)/100.0)
		if -stat.LoseBets != 0 {
			stat.TotalRewardRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.WinBets, 100)/ComputeFloat(-stat.LoseBets, 100))*100.0)
		}
		if r.Money > 0 {
			// stat.PlayerRewardRate = fmt.Sprintf("%d", int((int64(r.CashOut)+r.Diamond)*10000/int64(r.Money)))
			stat.PlayerRewardRate = fmt.Sprintf("%.2f%%", ComputeFloat(int64(r.CashOut+int32(r.Diamond)), 100)/ComputeFloat(int64(r.Money), 100)*100.0)
		} else {
			stat.PlayerRewardRate = "未充值"
		}
	}
}

func lhdLkyhStatSummaryMapping(stat *entity.LHDLkyhStat, m bson.M) {
	// 查询总数据
	lhdinfo := new(entity.LHDPlayerStat)
	n := bson.M{}
	// 手动复制 m 到 n
	for k, v := range m {
		n[k] = v
	}
	n["userid"] = bson.M{"$in": stat.Userids}
	m3 := []bson.M{
		{"$match": n},
		{"$group": bson.M{
			"_id":        nil,
			"bet_rounds": bson.M{"$sum": "$bet_rounds"},
			"win_rounds": bson.M{"$sum": "$win_rounds"},
			"bets":       bson.M{"$sum": "$bets"},
			"win_bets":   bson.M{"$sum": "$win_bets"},
			"lose_bets":  bson.M{"$sum": "$lose_bets"},
			"wins":       bson.M{"$sum": "$wins"},
			"loses":      bson.M{"$sum": "$loses"},
		}},
	}
	LHDPlayerStats.Pipe(m3).One(&lhdinfo)
	if stat.BSDays != 0 {
		stat.BSNumberAvg = fmt.Sprintf("%.2f", float64(stat.BSTimes)/float64(stat.BSDays))
	}
	if stat.BSTimes != 0 {
		stat.BSRate = fmt.Sprintf("%.2f%%", float64(stat.TZ)/float64(stat.BSTimes)*100.0)
		stat.BSBetAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.BSBets, 100)/float64(stat.BSTimes))
	}
	if stat.RZ != 0 {
		stat.YZRate = fmt.Sprintf("%.2f%%", float64(stat.YZNumber)/float64(stat.RZ)*100.0)
	}
	if stat.YZRounds != 0 {
		stat.YZBetAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.YZBets, 100)/float64(stat.YZRounds))
	}
	if stat.StrategyTimes > 0 {
		stat.StrategyMultiRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyMultiRounds)/float64(stat.StrategyTimes)*100.0)
		stat.StrategyWinRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyWinRounds)/float64(stat.StrategyTimes)*100.0)
	}
	stat.FStrategyBets = fmt.Sprintf("%.2f", float64(stat.StrategyBets)/100.0)
	if stat.StrategyTimes > 0 {
		stat.FStrategyBetsAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.StrategyBets, 100)/float64(stat.StrategyTimes))
	}
	if -stat.StrategyLoses != 0 {
		stat.StrategyRewardRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.StrategyWins, 100)/ComputeFloat(-stat.StrategyLoses, 100))*100.0)
	}
	// 查询总局数
	stat.Bets = lhdinfo.Bets
	stat.Rounds = int64(lhdinfo.BetRounds)
	stat.WinRounds = int64(lhdinfo.WinRounds)
	stat.WinBets = lhdinfo.Wins
	stat.LoseBets = lhdinfo.Loses

	if stat.Rounds != 0 {
		stat.RoundBetAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.Bets, 100)/float64(stat.Rounds))
		stat.TotalWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.Rounds)*100.0)
	}
	stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
	stat.FStrategyWins = fmt.Sprintf("%.2f", float64(stat.StrategyWins)/100.0)
	stat.FStrategyLoses = fmt.Sprintf("%.2f", float64(stat.StrategyLoses)/100.0)
	stat.FStrategyCash = fmt.Sprintf("%.2f", float64(stat.StrategyWins+stat.StrategyLoses)/100.0)
	if -stat.LoseBets != 0 {
		stat.TotalRewardRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.WinBets, 100)/ComputeFloat(-stat.LoseBets, 100))*100.0)
	}
	// if r.Money > 0 {
	// 	stat.PlayerRewardRate = fmt.Sprintf("%d", int((int64(r.CashOut)+r.Diamond)*10000/int64(r.Money)))
	// }
}

/*
	7updown统计
*/

// Get7updownStatPlayers 7updown玩家明细
func (s *gameStatsService) GetUpdownStatPlayers(page, pageSize int, m, m2, m3 bson.M, startDate, endDate string) (
	stats []*entity.UpDownPlayerStat, count int, err error,
) {
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)

	// 用户列表筛选
	var filterPlayerIds []string
	if len(m2) > 0 || len(m3) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				"_id":        "$userid",
				"bets":       bson.M{"$sum": "$bets"},
				"bet_rounds": bson.M{"$sum": "$bet_rounds"},
				"cash":       bson.M{"$sum": "$cash"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				"cash": "$cash",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = UpdownPlayerStats.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetLHDStatPlayers error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			userid := r["_id"].(string)
			userids = append(userids, userid)
		}
		if len(userids) == 0 { // 未匹配到
			return
		}
		now := utils.BsonNow()
		m3_2 := []bson.M{
			{"$match": bson.M{"_id": bson.M{"$in": userids}}},
			{"$project": bson.M{
				"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
				"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
			}},
			{"$match": m3},
		}
		var r3_2 []bson.M
		err = PlayerUsers.Pipe(m3_2).All(&r3_2)
		if err != nil {
			beego.Error("GetLHDStatPlayers error32:", err)
		} else if len(r3_2) == 0 { // 未匹配到
			return
		}
		for _, r := range r3_2 {
			userid := r["_id"].(string)
			filterPlayerIds = append(filterPlayerIds, userid)
		}
	}
	if len(filterPlayerIds) > 0 {
		m["userid"] = bson.M{"$in": filterPlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":                "$userid",
			"date":               bson.M{"$max": "$date"},
			"all_rounds":         bson.M{"$sum": "$all_rounds"},
			"game_times":         bson.M{"$sum": "$game_times"},
			"bets":               bson.M{"$sum": "$bets"},
			"bet_rounds":         bson.M{"$sum": "$bet_rounds"},
			"win_rounds":         bson.M{"$sum": "$win_rounds"},
			"lose_rounds":        bson.M{"$sum": "$lose_rounds"},
			"tie_rounds":         bson.M{"$sum": "$tie_rounds"},
			"win_bets":           bson.M{"$sum": "$win_bets"},
			"lose_bets":          bson.M{"$sum": "$lose_bets"},
			"tie_bets":           bson.M{"$sum": "$tie_bets"},
			"wins":               bson.M{"$sum": "$wins"},
			"loses":              bson.M{"$sum": "$loses"},
			"cash":               bson.M{"$sum": "$cash"},
			"dragons":            bson.M{"$sum": "$dragons"},
			"tigers":             bson.M{"$sum": "$tigers"},
			"ties":               bson.M{"$sum": "$ties"},
			"player_dragons":     bson.M{"$sum": "$player_dragons"},
			"player_tigers":      bson.M{"$sum": "$player_tigers"},
			"player_ties":        bson.M{"$sum": "$player_ties"},
			"player_win_dragons": bson.M{"$sum": "$player_win_dragons"},
			"player_win_tigers":  bson.M{"$sum": "$player_win_tigers"},
			"player_win_ties":    bson.M{"$sum": "$player_win_ties"},
			"player_multis":      bson.M{"$sum": "$player_multis"},
			"round_bet_avg":      bson.M{"$sum": "$round_bet_avg"},
			"money":              bson.M{"$sum": "$money"},
			"observe_rounds":     bson.M{"$sum": "$observe_rounds"},
		}},
		{"$project": bson.M{
			"_id": "$_id", "all_rounds": "$all_rounds", "game_times": "$game_times", "bets": "$bets", "bet_rounds": "$bet_rounds", "win_rounds": "$win_rounds", "lose_rounds": "$lose_rounds", "tie_rounds": "$tie_rounds", "win_bets": "$win_bets", "lose_bets": "$lose_bets", "tie_bets": "$tie_bets", "wins": "$wins", "loses": "$loses", "cash": "$cash", "dragons": "$dragons", "tigers": "$tigers", "ties": "$ties", "player_dragons": "$player_dragons", "player_tigers": "$player_tigers", "player_ties": "$player_ties", "player_win_dragons": "$player_win_dragons", "player_win_tigers": "$player_win_tigers", "player_win_ties": "$player_win_ties", "player_multis": "$player_multis", "money": "$money", "observe_rounds": "$observe_rounds",
			// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}}, // "can't $divide by zero"
			"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
		}},
		{"$match": m2},
		{"$count": "count"},
	}
	// 查总条数
	var rCount []bson.M
	err = UpdownPlayerStats.Pipe(m1).All(&rCount)
	if err != nil {
		return
	} else if len(rCount) > 0 {
		count = rCount[0]["count"].(int)
	}

	// 查列表数据
	m1 = m1[:len(m1)-1]
	m1 = append(m1, []bson.M{
		{"$sort": bson.M{"date": -1}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}...)
	err = UpdownPlayerStats.Pipe(m1).All(&stats)
	if err != nil {
		return
	}
	UpdownStatPlayersMapping(stats)

	// 查总汇数据
	mark := "--"
	summary := &entity.UpDownPlayerStat{}
	m1 = m1[:len(m1)-3]
	m1[1]["$group"].(bson.M)["_id"] = nil
	delete(m1[1]["$group"].(bson.M), "date")
	m1[1]["$group"].(bson.M)["userids"] = bson.M{"$addToSet": "$userid"}
	m1[2]["$project"].(bson.M)["userids"] = "$userids"
	var summarys []*entity.UpDownPlayerStat
	err = UpdownPlayerStats.Pipe(m1).All(&summarys)
	if err != nil {
		return
	} else if len(summarys) > 0 {
		summary = summarys[0]
		summary.No = fmt.Sprintf("%s-%s总汇", startDate, endDate)
		summary.Id = mark
		summary.FNickname = mark
		summary.FUserType = mark
		summary.FRegistArea = mark
		summary.FCtime = mark
		summary.FLoginTime = mark
		// summary.FLiveDays = mark
		// summary.FLoseDays = mark
		UpdownStatPlayersSummaryMapping(summary)
	}
	if len(summarys) > 0 {
		stats = append([]*entity.UpDownPlayerStat{summary}, stats...)
	}
	return
}

func UpdownStatPlayersMapping(stats []*entity.UpDownPlayerStat) {
	var playerIds []string
	// var playerStats = make(map[string]*entity.CrashPlayerStat, len(stats))
	for _, stat := range stats {
		playerIds = append(playerIds, stat.Id)
		// playerStats[stat.Id] = stat
	}
	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": playerIds}}},
		{"$project": bson.M{
			"_id":         "$_id",
			"nickname":    "$nickname",
			"money":       "$money",
			"cash_out":    "$cash_out",
			"diamond":     "$diamond",
			"ctime":       "$ctime",
			"login_time":  "$login_time",
			"regist_area": "$regist_area", //ab测试 0:A 1:B 2:C
			"state":       "$state",
		}},
	}
	var r2 []bson.M
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("lhd users error:", err)
		return
	}
	var r2Map = make(map[string]bson.M, len(r2))
	for _, r := range r2 {
		userid := r["_id"].(string)
		r2Map[userid] = r
	}
	now := utils.BsonNow()
	for i, stat := range stats {
		r, ok := r2Map[stat.Id]
		if !ok {
			continue
		}
		// 玩家昵称	玩家类型 账号类型 注册时间 最后登陆时间 存活天数 流失天数
		// userid := r["_id"].(string)
		nickname := r["nickname"].(string)
		money := r["money"].(int)
		cash_out := r["cash_out"].(int)
		diamond := r["diamond"].(int64)
		ctime := r["ctime"].(time.Time)
		login_time := r["login_time"].(time.Time)
		regist_area := r["regist_area"].(int)
		state := r["state"].(int)

		// 将日期转成印度时间
		login_time, _ = ConvertToIndiaTime(login_time.Unix())
		ctime, _ = ConvertToIndiaTime(ctime.Unix())

		stat.No = strconv.Itoa(i + 1)
		stat.FNickname = nickname
		stat.Money = uint32(money)
		stat.FCtime = ctime.Format(utils.FORMAT)
		stat.FLoginTime = login_time.Format(utils.FORMAT)
		stat.FLiveDays = int(math.Ceil(login_time.Sub(ctime).Hours() / 24))
		stat.FLoseDays = int(math.Floor(now.Sub(login_time).Hours() / 24))

		// summary.FLiveDays += stat.FLiveDays
		// summary.FLoseDays += stat.FLoseDays
		// summary.AllRounds += stat.AllRounds
		// summary.GameTimes += stat.GameTimes
		// summary.Bets += stat.Bets
		// summary.BetRounds += stat.BetRounds
		// summary.WinRounds += stat.WinRounds
		// summary.LoseRounds += stat.LoseRounds
		// summary.TieRounds += stat.TieRounds
		// summary.WinBets += stat.WinBets
		// summary.LoseBets += stat.LoseBets
		// summary.TieBets += stat.TieBets
		// summary.Wins += stat.Wins
		// summary.Loses += stat.Loses
		// summary.Cash += stat.Cash
		// summary.MulpitleSum += stat.MulpitleSum
		// summary.WinEscapeMulpitleSum += stat.WinEscapeMulpitleSum
		// summary.WinMulpitleSum += stat.WinMulpitleSum
		// summary.LoseMulpitleSum += stat.LoseMulpitleSum
		// summary.Mulpitles = append(summary.Mulpitles, stat.Mulpitles...)
		// summary.WinEscapeMulpitles = append(summary.WinEscapeMulpitles, stat.WinEscapeMulpitles...)
		// summary.WinMulpitles = append(summary.WinMulpitles, stat.WinMulpitles...)
		// summary.LoseMulpitles = append(summary.LoseMulpitles, stat.LoseMulpitles...)
		// summary.ObserveRounds += stat.ObserveRounds
		// summary.SMoney += money
		// summary.SCashOut += cash_out
		// summary.SDiamond += diamond

		switch regist_area {
		case 0:
			stat.FRegistArea = "A类"
		case 1:
			stat.FRegistArea = "B类"
		case 2:
			stat.FRegistArea = "C类"
		}

		switch state {
		case 1:
			stat.FUserType = "新手"
		case 2:
			if money == 0 {
				stat.FUserType = "零充"
			} else if money >= 20000 && money <= 99999 {
				stat.FUserType = "普充"
			} else if money >= 100000 && money <= 499999 {
				stat.FUserType = "小R"
			} else if money >= 500000 && money <= 999999 {
				stat.FUserType = "中R"
			} else if money >= 1000000 && money <= 9999999 {
				stat.FUserType = "大R"
			} else if money >= 10000000 {
				stat.FUserType = "超大R"
			}
		case 3:
			stat.FUserType = "平民"
		case 4:
			stat.FUserType = "泡沫"
		}

		stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
		if stat.BetRounds > 0 {
			// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.BetRounds)/100.0)
			stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
		}
		stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
		stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
		stat.FTieBets = fmt.Sprintf("%.2f", float64(stat.TieBets)/100.0)
		if stat.BetRounds > 0 {
			stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
		}
		if stat.Bets > 0 {
			// stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
			stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
		}
		if stat.WinRounds > 0 {
			stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
		}
		if stat.TieRounds > 0 {
			stat.FTieBetsAvg = fmt.Sprintf("%.2f", float64(stat.TieBets)/float64(stat.TieRounds)/100.0)
		}
		stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
		stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
		stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)

		if stat.WinRounds > 0 {
			stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
		}
		// 押龙/虎/和率
		if stat.BetRounds > 0 {
			stat.PlayerDragonsRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerDragons)/float64(stat.BetRounds)*100)
			stat.PlayerTigersRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerTigers)/float64(stat.BetRounds)*100)
			stat.PlayerTiesRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerTies)/float64(stat.BetRounds)*100)
		}
		if stat.Dragons > 0 {
			stat.PlayerDragonsWinRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinDragons)/float64(stat.Dragons)*100)
		}
		if stat.Tigers > 0 {
			stat.PlayerTigersWinRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinTigers)/float64(stat.Tigers)*100)
		}
		if stat.Ties > 0 {
			stat.PlayerTiesWinRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinTies)/float64(stat.Ties)*100)
		}
		// 多门率
		if stat.BetRounds > 0 {
			stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerMultis)/float64(stat.BetRounds)*100)
		}

		// 观察局
		if stat.BetRounds > 0 {
			stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
		}
		if stat.BetRounds > 0 {
			stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
		}
		stat.FMoney = fmt.Sprintf("%.2f", float64(money)/100)
		stat.FCashOut = fmt.Sprintf("%.2f", float64(cash_out)/100)
		stat.FWinMoney = fmt.Sprintf("%.2f", float64(int(diamond)+cash_out-money)/100)
	}
}

func UpdownStatPlayersSummaryMapping(stat *entity.UpDownPlayerStat) {
	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": stat.Userids}}},
		{"$group": bson.M{
			"_id":      nil,
			"money":    bson.M{"$sum": "$money"},
			"cash_out": bson.M{"$sum": "$cash_out"},
			"diamond":  bson.M{"$sum": "$diamond"},
		}},
	}
	var r2 []bson.M
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("lhd users error:", err)
		return
	} else if len(r2) > 0 {
		money := r2[0]["money"].(int)
		cash_out := r2[0]["cash_out"].(int)
		diamond := r2[0]["diamond"].(int64)
		stat.SMoney = money
		stat.SCashOut = cash_out
		stat.SDiamond = diamond
	}

	stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
	if stat.BetRounds > 0 {
		stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.BetRounds)/100.0)
		// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
	}

	stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
	stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
	stat.FTieBets = fmt.Sprintf("%.2f", float64(stat.TieBets)/100.0)
	if stat.BetRounds > 0 {
		stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
	}
	if stat.Bets > 0 {
		// stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
		stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
	}
	if stat.WinRounds > 0 {
		stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
	}
	if stat.TieRounds > 0 {
		stat.FTieBetsAvg = fmt.Sprintf("%.2f", float64(stat.TieBets)/float64(stat.TieRounds)/100.0)
	}
	stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
	stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
	stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)
	if stat.WinRounds > 0 {
		stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
	}
	// 押龙/虎/和率
	if stat.BetRounds > 0 {
		stat.PlayerDragonsRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerDragons)/float64(stat.BetRounds)*100)
		stat.PlayerTigersRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerTigers)/float64(stat.BetRounds)*100)
		stat.PlayerTiesRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerTies)/float64(stat.BetRounds)*100)
	}
	if stat.Dragons > 0 {
		stat.PlayerDragonsWinRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinDragons)/float64(stat.Dragons)*100)
	}
	if stat.Tigers > 0 {
		stat.PlayerTigersWinRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinTigers)/float64(stat.Tigers)*100)
	}
	if stat.Ties > 0 {
		stat.PlayerTiesWinRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinTies)/float64(stat.Ties)*100)
	}
	// 多门率
	if stat.BetRounds > 0 {
		stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerMultis)/float64(stat.BetRounds)*100)
	}

	// 观察局
	if stat.BetRounds > 0 {
		stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
	}
	if stat.BetRounds > 0 {
		stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
	}
	stat.FMoney = fmt.Sprintf("%.2f", float64(stat.SMoney)/100)
	stat.FCashOut = fmt.Sprintf("%.2f", float64(stat.SCashOut)/100)
	stat.FWinMoney = fmt.Sprintf("%.2f", float64(int(stat.SDiamond)+stat.SCashOut-stat.SMoney)/100)
}

// GetUpDownStatDates 7updown汇总明细
func (s *gameStatsService) GetUpDownStatDates(m, m2, m3 bson.M, startDate, endDate string) (stats []*entity.UpDownPlayerStat, err error) {
	// 用户列表筛选
	var filterDatePlayerIds []string
	if len(m2) > 0 || len(m3) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				// "_id":        bson.M{"userid":"$userid", "date", "$date"},
				"_id":        "$_id",
				"bets":       bson.M{"$sum": "$bets"},
				"bet_rounds": bson.M{"$sum": "$bet_rounds"},
				"cash":       bson.M{"$sum": "$cash"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				"cash": "$cash",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = UpdownPlayerStats.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetLHDStatDates error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			id := r["_id"].(string)
			filterDatePlayerIds = append(filterDatePlayerIds, id)
			ids := strings.Split(id, "-") // id=date-userid
			if len(ids) >= 2 {
				userids = append(userids, ids[1])
			}
		}
		if len(filterDatePlayerIds) == 0 { // 未匹配到
			return
		}
		if len(m3) > 0 { // 流失天数，局均打码量
			if len(userids) == 0 { // 未匹配到
				return
			}
			now := utils.BsonNow()
			m3_2 := []bson.M{
				{"$match": bson.M{"_id": bson.M{"$in": userids}}},
				{"$project": bson.M{
					"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
					"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
				}},
				{"$match": m3},
			}
			var r3_2 []bson.M
			err = PlayerUsers.Pipe(m3_2).All(&r3_2)
			if err != nil {
				beego.Error("GetCrashStatPlayers error32:", err)
			} else if len(r3_2) == 0 { // 未匹配到
				return
			}
			var filterId2 []string
			for _, id := range filterDatePlayerIds {
				for _, r := range r3_2 {
					userid := r["_id"].(string)
					if strings.HasSuffix(id, "-"+userid) {
						filterId2 = append(filterId2, id)
						break
					}
				}
			}
			if len(filterId2) == 0 { // 未匹配到
				return
			}
			filterDatePlayerIds = filterId2
		}
	}
	if len(filterDatePlayerIds) > 0 {
		m["_id"] = bson.M{"$in": filterDatePlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":                "$date_str",
			"date":               bson.M{"$min": "$date"},
			"players":            bson.M{"$sum": 1},
			"new_players":        bson.M{"$sum": bson.M{"$cond": bson.M{"if": "$new_reg", "then": 1, "else": 0}}},
			"old_players":        bson.M{"$sum": bson.M{"$cond": bson.M{"if": "$new_reg", "then": 0, "else": 1}}},
			"all_rounds":         bson.M{"$sum": "$all_rounds"},
			"game_times":         bson.M{"$sum": "$game_times"},
			"bets":               bson.M{"$sum": "$bets"},
			"bet_rounds":         bson.M{"$sum": "$bet_rounds"},
			"win_rounds":         bson.M{"$sum": "$win_rounds"},
			"lose_rounds":        bson.M{"$sum": "$lose_rounds"},
			"tie_rounds":         bson.M{"$sum": "$tie_rounds"},
			"win_bets":           bson.M{"$sum": "$win_bets"},
			"lose_bets":          bson.M{"$sum": "$lose_bets"},
			"tie_bets":           bson.M{"$sum": "$tie_bets"},
			"wins":               bson.M{"$sum": "$wins"},
			"loses":              bson.M{"$sum": "$loses"},
			"cash":               bson.M{"$sum": "$cash"},
			"dragons":            bson.M{"$sum": "$dragons"},
			"tigers":             bson.M{"$sum": "$tigers"},
			"ties":               bson.M{"$sum": "$ties"},
			"player_dragons":     bson.M{"$sum": "$player_dragons"},
			"player_tigers":      bson.M{"$sum": "$player_tigers"},
			"player_ties":        bson.M{"$sum": "$player_ties"},
			"player_win_dragons": bson.M{"$sum": "$player_win_dragons"},
			"player_win_tigers":  bson.M{"$sum": "$player_win_tigers"},
			"player_win_ties":    bson.M{"$sum": "$player_win_ties"},
			"player_multis":      bson.M{"$sum": "$player_multis"},
			"round_bet_avg":      bson.M{"$sum": "$round_bet_avg"},
			"money":              bson.M{"$sum": "$money"},
			"observe_rounds":     bson.M{"$sum": "$observe_rounds"},
			"all_rounds_list":    bson.M{"$push": "$bet_rounds"},
		}},
		{"$sort": bson.M{"date": -1}},
	}
	err = UpdownPlayerStats.Pipe(m1).All(&stats)
	if err != nil {
		return
	}

	// 查询日活数据
	var dateLoginedUsers = make(map[int64]int)
	m9 := []bson.M{
		{"$match": bson.M{"date": m["date"]}},
		{"$project": bson.M{"date": "$date", "logined_users": "$logined_users"}},
	}
	var r9 []bson.M
	err = PddStats.Pipe(m9).All(&r9)
	if err != nil {
		beego.Error("GetCrashStatDates logined_users error: ", err)
	} else {
		for _, r := range r9 {
			date := r["date"].(int64)
			logined_users := r["logined_users"].(int)
			dateLoginedUsers[date] = logined_users
		}
	}

	mark := "--"
	summary := &entity.UpDownPlayerStat{
		DateStr: fmt.Sprintf("%s-%s总汇", startDate, endDate),
		Id:      mark,
	}
	UpdownStatDatesMapping(summary, stats, dateLoginedUsers)
	UpdownStatDatesSummaryMapping(summary)
	stats = append([]*entity.UpDownPlayerStat{summary}, stats...)
	return
}

func UpdownStatDatesMapping(summary *entity.UpDownPlayerStat, stats []*entity.UpDownPlayerStat, dateLoginedUsers map[int64]int) {
	for _, stat := range stats {
		dateStr := stat.Id
		date, _ := utils.Unix(dateStr)
		stat.Date = date
		stat.DateStr = dateStr
		loginedUsers := dateLoginedUsers[stat.Date] // 日活

		stat.DateStr = utils.Stamp2Time(stat.Date).Format("2006-01-02")
		stat.FLoginedUsers = loginedUsers

		summary.FLoginedUsers += stat.FLoginedUsers
		summary.FLiveDays += stat.FLiveDays
		summary.FLoseDays += stat.FLoseDays
		summary.AllRounds += stat.AllRounds
		summary.GameTimes += stat.GameTimes
		summary.Bets += stat.Bets
		summary.BetRounds += stat.BetRounds
		summary.WinRounds += stat.WinRounds
		summary.LoseRounds += stat.LoseRounds
		summary.TieRounds += stat.TieRounds
		summary.WinBets += stat.WinBets
		summary.LoseBets += stat.LoseBets
		summary.TieBets += stat.TieBets
		summary.Wins += stat.Wins
		summary.Loses += stat.Loses
		summary.Cash += stat.Cash
		summary.Dragons += stat.Dragons
		summary.Tigers += stat.Tigers
		summary.Ties += stat.Ties
		summary.PlayerDragons += stat.PlayerDragons
		summary.PlayerTigers += stat.PlayerTigers
		summary.PlayerTies += stat.PlayerTies
		summary.PlayerWinDragons += stat.PlayerWinDragons
		summary.PlayerWinTigers += stat.PlayerWinTigers
		summary.PlayerWinTies += stat.PlayerWinTies
		summary.PlayerMultis += stat.PlayerMultis
		// summary.RoundBetAvg += stat.RoundBetAvg

		summary.ObserveRounds += stat.ObserveRounds
		summary.Players += stat.Players
		summary.NewPlayers += stat.NewPlayers
		summary.OldPlayers += stat.OldPlayers
		summary.AllRoundsList = append(summary.AllRoundsList, stat.AllRoundsList...)

		// 玩 crash 人数占日活比
		if stat.FLoginedUsers > 0 {
			stat.FPlayerRate = fmt.Sprintf("%.2f", float64(stat.Players)/float64(stat.FLoginedUsers))
		}

		// 查询当天注册的新用户
		startTime := utils.Str2Time(fmt.Sprintf("%s 00:00:00", dateStr), Location())
		endTime := utils.Str2Time(fmt.Sprintf("%s 23:59:59", dateStr), Location())
		m7 := []bson.M{
			{"$match": bson.M{
				"ctime":            bson.M{"$gte": startTime, "$lte": endTime},
				"robot":            false,
				"simulation_robot": false,
			}},
			{"$group": bson.M{
				"_id":   nil,
				"total": bson.M{"$sum": 1},
			}},
		}
		var r7 []bson.M
		err := PlayerUsers.Pipe(m7).All(&r7)
		if err != nil {
			beego.Error("查询当天注册的新用户 error: ", err)
		} else if len(r7) > 0 {
			stat.FLoginedNewUsers = r7[0]["total"].(int)
		}
		stat.FLoginedOldUsers = stat.FLoginedUsers - stat.FLoginedNewUsers
		if stat.FLoginedOldUsers < 0 {
			stat.FLoginedOldUsers = 0
		}
		summary.FLoginedNewUsers += stat.FLoginedNewUsers
		summary.FLoginedOldUsers += stat.FLoginedOldUsers
		// FNewPlayerRate, FOldPlayerRate 占新老玩家日活
		if stat.FLoginedNewUsers > 0 {
			stat.FNewPlayerRate = fmt.Sprintf("%.2f", float64(stat.NewPlayers)/float64(stat.FLoginedNewUsers))
		}
		if stat.FLoginedOldUsers > 0 {
			stat.FOldPlayerRate = fmt.Sprintf("%.2f", float64(stat.OldPlayers)/float64(stat.FLoginedOldUsers))
		}

		if stat.Players > 0 {
			stat.FPlayerRoundsAvg = fmt.Sprintf("%.2f", float64(stat.BetRounds)/float64(stat.Players))
		}
		// 局数中位数
		if len(stat.AllRoundsList) > 0 {
			var rounds = stat.AllRoundsList
			sort.Slice(rounds, func(i, j int) bool {
				return rounds[i] < rounds[j]
			})
			length := len(rounds)
			if length > 0 {
				if length%2 == 1 {
					stat.FPlayerRoundsMedian = rounds[length/2]
				} else {
					// 中间两位算平均
					stat.FPlayerRoundsMedian = (rounds[length/2] + rounds[length/2-1]) / 2
				}
			}
		}
		// stat.FGameTimes = fmt.Sprintf("%.2f", float64(stat.GameTimes)/60)
		if stat.BetRounds > 0 {
			stat.FGameTimesAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.BetRounds)/60)
		}
		if stat.Players > 0 {
			stat.FGameTimesPlayerAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.Players)/60)
		}

		stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
		if stat.Players > 0 {
			stat.FPlayerBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.Players)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.BetRounds)/100.0)
			// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
		}
		stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
		stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
		stat.FTieBets = fmt.Sprintf("%.2f", float64(stat.TieBets)/100.0)
		if stat.BetRounds > 0 {
			stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
		}
		if stat.Bets > 0 {
			stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
		}
		if stat.WinRounds > 0 {
			stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
		}
		if stat.TieRounds > 0 {
			stat.FTieBetsAvg = fmt.Sprintf("%.2f", float64(stat.TieBets)/float64(stat.TieRounds)/100.0)
		}
		stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
		stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
		stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)

		if stat.WinRounds > 0 {
			stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
		}
		// 押龙/虎/和率
		if stat.BetRounds > 0 {
			stat.PlayerDragonsRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerDragons)/float64(stat.BetRounds)*100)
			stat.PlayerTigersRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerTigers)/float64(stat.BetRounds)*100)
			stat.PlayerTiesRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerTies)/float64(stat.BetRounds)*100)
		}
		if stat.Dragons > 0 {
			stat.PlayerDragonsWinRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinDragons)/float64(stat.Dragons)*100)
		}
		if stat.Tigers > 0 {
			stat.PlayerTigersWinRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinTigers)/float64(stat.Tigers)*100)
		}
		if stat.Ties > 0 {
			stat.PlayerTiesWinRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinTies)/float64(stat.Ties)*100)
		}
		// 多门率
		if stat.BetRounds > 0 {
			stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerMultis)/float64(stat.BetRounds)*100)
		}

		// 观察局
		if stat.BetRounds > 0 {
			stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
		}
		if stat.BetRounds > 0 {
			stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
		}

	}
}

func UpdownStatDatesSummaryMapping(stat *entity.UpDownPlayerStat) {
	// 玩 lhd 人数占日活比
	if stat.FLoginedUsers > 0 {
		stat.FPlayerRate = fmt.Sprintf("%.2f", float64(stat.Players)/float64(stat.FLoginedUsers))
	}

	// FNewPlayerRate, FOldPlayerRate 占新老玩家日活
	if stat.FLoginedNewUsers > 0 {
		stat.FNewPlayerRate = fmt.Sprintf("%.2f", float64(stat.NewPlayers)/float64(stat.FLoginedNewUsers))
	}
	if stat.FLoginedOldUsers > 0 {
		stat.FOldPlayerRate = fmt.Sprintf("%.2f", float64(stat.OldPlayers)/float64(stat.FLoginedOldUsers))
	}

	if stat.Players > 0 {
		stat.FPlayerRoundsAvg = fmt.Sprintf("%.2f", float64(stat.BetRounds)/float64(stat.Players))
	}
	// 局数中位数
	if len(stat.AllRoundsList) > 0 {
		var rounds = stat.AllRoundsList
		sort.Slice(rounds, func(i, j int) bool {
			return rounds[i] < rounds[j]
		})
		length := len(rounds)
		if length > 0 {
			if length%2 == 1 {
				stat.FPlayerRoundsMedian = rounds[length/2]
			} else {
				// 中间两位算平均
				stat.FPlayerRoundsMedian = (rounds[length/2] + rounds[length/2-1]) / 2
			}
		}
	}
	// stat.FGameTimes = fmt.Sprintf("%.2f", float64(stat.GameTimes)/60)
	if stat.BetRounds > 0 {
		stat.FGameTimesAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.BetRounds)/60)
	}
	if stat.Players > 0 {
		stat.FGameTimesPlayerAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.Players)/60)
	}

	if stat.BetRounds > 0 {
		stat.FGameTimesAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.BetRounds)/60)
	}
	if stat.Players > 0 {
		stat.FGameTimesPlayerAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.Players)/60)
	}

	stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
	if stat.Players > 0 {
		stat.FPlayerBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.Players)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.BetRounds)/100.0)
		// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
	}
	stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
	stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
	stat.FTieBets = fmt.Sprintf("%.2f", float64(stat.TieBets)/100.0)

	if stat.BetRounds > 0 {
		stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
	}
	if stat.Bets > 0 {
		stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
	}
	if stat.WinRounds > 0 {
		stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
	}
	stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
	stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
	stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)

	if stat.WinRounds > 0 {
		stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
	}
	if stat.TieRounds > 0 {
		stat.FTieBetsAvg = fmt.Sprintf("%.2f", float64(stat.TieBets)/float64(stat.TieRounds)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
	}
	// 押龙/虎/和率
	if stat.BetRounds > 0 {
		stat.PlayerDragonsRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerDragons)/float64(stat.BetRounds)*100)
		stat.PlayerTigersRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerTigers)/float64(stat.BetRounds)*100)
		stat.PlayerTiesRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerTies)/float64(stat.BetRounds)*100)
	}
	if stat.Dragons > 0 {
		stat.PlayerDragonsWinRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinDragons)/float64(stat.Dragons)*100)
	}
	if stat.Tigers > 0 {
		stat.PlayerTigersWinRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinTigers)/float64(stat.Tigers)*100)
	}
	if stat.Ties > 0 {
		stat.PlayerTiesWinRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinTies)/float64(stat.Ties)*100)
	}
	// 多门率
	if stat.BetRounds > 0 {
		stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerMultis)/float64(stat.BetRounds)*100)
	}

	// 观察局
	if stat.BetRounds > 0 {
		stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
	}
	if stat.BetRounds > 0 {
		stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
	}
}

// GetUpdownXxscStat 7updown新想事成
func (s *gameStatsService) GetUpdownXxscStat(page, pageSize int, m, m2, m3 bson.M, startDate, endDate string) (
	stats []*entity.UpDownXxscStat, count int, err error,
) {
	// UpDownXxscStat
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)

	// 用户列表筛选
	var filterPlayerIds []string
	if len(m2) > 0 || len(m3) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				"_id":        "$userid",
				"bets":       bson.M{"$sum": "$bets"},
				"bet_rounds": bson.M{"$sum": "$bet_rounds"},
				"cash":       bson.M{"$sum": "$cash"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				"cash": "$cash",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = UpdownPlayerStats.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("LHDPlayerStats error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			userid := r["_id"].(string)
			userids = append(userids, userid)
		}
		if len(userids) == 0 { // 未匹配到
			return
		}
		now := utils.BsonNow()
		m3_2 := []bson.M{
			{"$match": bson.M{"_id": bson.M{"$in": userids}}},
			{"$project": bson.M{
				"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
				"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
			}},
			{"$match": m3},
		}
		var r3_2 []bson.M
		err = PlayerUsers.Pipe(m3_2).All(&r3_2)
		if err != nil {
			beego.Error("lhd PlayerUsers error32:", err)
		} else if len(r3_2) == 0 { // 未匹配到
			return
		}
		for _, r := range r3_2 {
			userid := r["_id"].(string)
			filterPlayerIds = append(filterPlayerIds, userid)
		}
	}

	if len(filterPlayerIds) > 0 {
		m["userid"] = bson.M{"$in": filterPlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":  "$userid",
			"date": bson.M{"$max": "$date"},
			// "date_str":                bson.M{"$sum": "$date_str"},
			"strategy_players":        bson.M{"$max": "$strategy_players"},
			"strategy_rounds":         bson.M{"$sum": "$strategy_rounds"},
			"strategy_disturb_rounds": bson.M{"$sum": "$strategy_disturb_rounds"},
			"strategy_multi_rounds":   bson.M{"$sum": "$strategy_multi_rounds"},
			"strategy_bets":           bson.M{"$sum": "$strategy_bets"},
			"strategy_win_rounds":     bson.M{"$sum": "$strategy_win_rounds"},
			"strategy_wins":           bson.M{"$sum": "$strategy_wins"},
			"strategy_loses":          bson.M{"$sum": "$strategy_loses"},
		}},
		{"$project": bson.M{
			"_id": "$_id", "date": "$date", "userid": "$userid", "strategy_players": "$strategy_players", "strategy_rounds": "$strategy_rounds", "strategy_disturb_rounds": "$strategy_disturb_rounds", "strategy_multi_rounds": "$strategy_multi_rounds", "strategy_bets": "$strategy_bets", "strategy_win_rounds": "$strategy_win_rounds", "strategy_wins": "$strategy_wins", "strategy_loses": "$strategy_loses",
		}},
		{"$match": m2},
		{"$count": "count"},
	}
	// 查总条数
	var rCount []bson.M
	err = UpdownXxscStats.Pipe(m1).All(&rCount)
	if err != nil {
		return
	} else if len(rCount) > 0 {
		count = rCount[0]["count"].(int)
	}

	// 查列表数据
	m1 = m1[:len(m1)-1]
	m1 = append(m1, []bson.M{
		{"$sort": bson.M{"date": -1}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}...)
	err = UpdownXxscStats.Pipe(m1).All(&stats)
	if err != nil {
		return
	}
	UpdownXxscStatMapping(stats)

	// 查总汇数据
	mark := "--"
	summary := &entity.UpDownXxscStat{}
	m1 = m1[:len(m1)-3]
	m1[1]["$group"].(bson.M)["_id"] = nil
	delete(m1[1]["$group"].(bson.M), "date")
	m1[1]["$group"].(bson.M)["userids"] = bson.M{"$addToSet": "$userid"}
	m1[1]["$group"].(bson.M)["strategy_players"] = bson.M{"$sum": "$strategy_players"} // 触发策略人数总汇
	m1[2]["$project"].(bson.M)["userids"] = "$userids"
	str, err := json.Marshal(m1)
	fmt.Println(err, string(str))
	var summarys []*entity.UpDownXxscStat
	err = UpdownXxscStats.Pipe(m1).All(&summarys)
	if err != nil {
		beego.Error("lhd summary error:", err)
		return
	} else if len(summarys) > 0 {
		summary = summarys[0]
		summary.No = fmt.Sprintf("%s-%s总汇", startDate, endDate)
		summary.Id = mark
		summary.FNickname = mark
		summary.FUserType = mark
		summary.FRegistArea = mark
		summary.FCtime = mark
		summary.FLoginTime = mark
		UpdownXxscStatSummaryMapping(summary)
	}
	if len(summarys) > 0 {
		stats = append([]*entity.UpDownXxscStat{summary}, stats...)
	}
	return
}

func UpdownXxscStatMapping(stats []*entity.UpDownXxscStat) {
	var playerIds []string
	// var playerStats = make(map[string]*entity.CrashPlayerStat, len(stats))
	for _, stat := range stats {
		playerIds = append(playerIds, stat.Id)
		// playerStats[stat.Id] = stat
	}
	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": playerIds}}},
		{"$project": bson.M{
			"_id":            "$_id",
			"nickname":       "$nickname",
			"money":          "$money",
			"cash_out":       "$cash_out",
			"diamond":        "$diamond",
			"ctime":          "$ctime",
			"login_time":     "$login_time",
			"regist_area":    "$regist_area", //ab测试 0:A 1:B 2:C
			"state":          "$state",
			"seven_strategy": "$seven_strategy",
		}},
	}

	var r2 []*entity.PlayerUser
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("lhd users error:", err)
		return
	}
	var r2Map = make(map[string]*entity.PlayerUser, len(r2))
	for _, r := range r2 {
		userid := r.Userid
		r2Map[userid] = r
	}
	now := utils.BsonNow()
	for i, stat := range stats {
		r, ok := r2Map[stat.Id]
		if !ok {
			continue
		}
		// 玩家昵称	玩家类型 账号类型 注册时间 最后登陆时间 存活天数 流失天数
		nickname := r.Nickname
		money := r.Money
		ctime := r.Ctime
		login_time := r.LoginTime
		regist_area := r.RegistArea
		state := r.State

		// 将日期转成印度时间
		login_time, _ = ConvertToIndiaTime(login_time.Unix())
		ctime, _ = ConvertToIndiaTime(ctime.Unix())

		stat.No = strconv.Itoa(i + 1)
		stat.FNickname = nickname
		stat.FCtime = ctime.Format(utils.FORMAT)
		stat.FLoginTime = login_time.Format(utils.FORMAT)
		stat.FLiveDays = int(math.Ceil(login_time.Sub(ctime).Hours() / 24))
		stat.FLoseDays = int(math.Floor(now.Sub(login_time).Hours() / 24))

		switch regist_area {
		case 0:
			stat.FRegistArea = "A类"
		case 1:
			stat.FRegistArea = "B类"
		case 2:
			stat.FRegistArea = "C类"
		}

		switch state {
		case 1:
			stat.FUserType = "新手"
		case 2:
			if money == 0 {
				stat.FUserType = "零充"
			} else if money >= 20000 && money <= 99999 {
				stat.FUserType = "普充"
			} else if money >= 100000 && money <= 499999 {
				stat.FUserType = "小R"
			} else if money >= 500000 && money <= 999999 {
				stat.FUserType = "中R"
			} else if money >= 1000000 && money <= 9999999 {
				stat.FUserType = "大R"
			} else if money >= 10000000 {
				stat.FUserType = "超大R"
			}
		case 3:
			stat.FUserType = "平民"
		case 4:
			stat.FUserType = "泡沫"
		}

		// stat.StrategyTimes = int32(r.SevenXXSCMaxTriggerTimes)
		// stat.StrategyValidTimes = int32(r.SevenXXSCTriggerTimes)
		stat.StrategyTimes = int32(r.SevenStrategy.XXSC.MaxTriggerTimes)
		stat.StrategyValidTimes = int32(r.SevenStrategy.XXSC.TriggerTimes)
		stat.StrategyCash = stat.StrategyWins + stat.StrategyLoses // 赢-输
		if stat.StrategyTimes > 0 {
			stat.StrategyValidTimesRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyValidTimes)/float64(stat.StrategyTimes)*100.0)
		}
		if stat.StrategyPlayers > 0 {
			stat.StrategyValidTimesAvg = fmt.Sprintf("%.2f", float64(stat.StrategyValidTimes)/float64(stat.StrategyPlayers))
		}
		if stat.StrategyTimes > 0 {
			stat.StrategyRoundAvg = fmt.Sprintf("%.2f", float64(stat.StrategyRounds)/float64(stat.StrategyTimes))
		}
		if stat.StrategyValidTimes > 0 {
			stat.StrategyValidRoundAvg = fmt.Sprintf("%.2f", float64(stat.StrategyRounds)/float64(stat.StrategyValidTimes))
		}
		if stat.StrategyRounds > 0 {
			stat.StrategyDisturbRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyDisturbRounds)/float64(stat.StrategyRounds)*100.0)
		}
		stat.StrategyMultiRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyMultiRounds)/float64(stat.StrategyRounds)*100.0)
		stat.FStrategyBets = fmt.Sprintf("%.2f", float64(stat.StrategyBets)/100.0)
		if stat.StrategyRounds > 0 {
			stat.FStrategyBetRoundsAvg = fmt.Sprintf("%.2f", float64(stat.StrategyBets)/float64(stat.StrategyRounds)/100.0)
		}

		stat.FStrategyWins = fmt.Sprintf("%.2f", float64(stat.StrategyWins)/100.0)
		stat.FStrategyLoses = fmt.Sprintf("%.2f", float64(stat.StrategyLoses)/100.0)
		stat.FStrategyCash = fmt.Sprintf("%.2f", float64(stat.StrategyWins+stat.StrategyLoses)/100.0)
		if stat.StrategyRounds > 0 {
			stat.FStrategyWinRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyWinRounds)/float64(stat.StrategyRounds)*100.0)
		}
		if -stat.StrategyLoses != 0 {
			stat.FStrategyRebateRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyWins)/float64(-stat.StrategyLoses)*100.0)
		} else {
			stat.FStrategyRebateRate = "未输过"
		}
	}
}

func UpdownXxscStatSummaryMapping(stat *entity.UpDownXxscStat) {
	if len(stat.Userids) == 0 {
		return
	}
	// 查用户策略触发次数
	m1 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": stat.Userids}}},
		{"$group": bson.M{
			"_id":               nil,
			"trigger_times":     bson.M{"$sum": "$seven_strategy.xxsc.trigger_times"},
			"max_trigger_times": bson.M{"$sum": "$seven_strategy.xxsc.max_trigger_times"},
		}},
	}
	var r1 bson.M
	err := PlayerUsers.Pipe(m1).One(&r1)
	if err != nil {
		beego.Error("UpdownXxscStatSummaryMapping err:", err)
		return
	}
	stat.StrategyTimes = int32(r1["max_trigger_times"].(int))
	stat.StrategyValidTimes = int32(r1["trigger_times"].(int))
	stat.StrategyPlayers = int32(len(stat.Userids))
	// for _, s := range stats {
	// 	stat.StrategyTimes += s.StrategyTimes
	// 	stat.StrategyValidTimes += s.StrategyValidTimes
	// }

	// stat.StrategyTimes = int32(r.LHDStrategy.XXSC.MaxTriggerTimes)
	// stat.StrategyValidTimes = int32(r.LHDStrategy.XXSC.TriggerTimes)
	stat.StrategyCash = stat.StrategyWins + stat.StrategyLoses // 赢-输
	if stat.StrategyTimes > 0 {
		stat.StrategyValidTimesRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyValidTimes)/float64(stat.StrategyTimes)*100.0)
	}
	if stat.StrategyPlayers > 0 {
		stat.StrategyValidTimesAvg = fmt.Sprintf("%.2f", float64(stat.StrategyValidTimes)/float64(stat.StrategyPlayers))
	}
	if stat.StrategyTimes > 0 {
		stat.StrategyRoundAvg = fmt.Sprintf("%.2f", float64(stat.StrategyRounds)/float64(stat.StrategyTimes))
	}
	if stat.StrategyValidTimes > 0 {
		stat.StrategyValidRoundAvg = fmt.Sprintf("%.2f", float64(stat.StrategyRounds)/float64(stat.StrategyValidTimes))
	}
	if stat.StrategyRounds > 0 {
		stat.StrategyDisturbRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyDisturbRounds)/float64(stat.StrategyRounds)*100.0)
	}
	stat.StrategyMultiRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyMultiRounds)/float64(stat.StrategyRounds)*100.0)
	stat.FStrategyBets = fmt.Sprintf("%.2f", float64(stat.StrategyBets)/100.0)
	if stat.StrategyRounds > 0 {
		stat.FStrategyBetRoundsAvg = fmt.Sprintf("%.2f", float64(stat.StrategyBets)/float64(stat.StrategyRounds)/100.0)
	}

	stat.FStrategyWins = fmt.Sprintf("%.2f", float64(stat.StrategyWins)/100.0)
	stat.FStrategyLoses = fmt.Sprintf("%.2f", float64(stat.StrategyLoses)/100.0)
	stat.FStrategyCash = fmt.Sprintf("%.2f", float64(stat.StrategyWins+stat.StrategyLoses)/100.0)
	if stat.StrategyRounds > 0 {
		stat.FStrategyWinRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyWinRounds)/float64(stat.StrategyRounds)*100.0)
	}
	if -stat.StrategyLoses != 0 {
		stat.FStrategyRebateRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyWins)/float64(-stat.StrategyLoses)*100.0)
	} else {
		stat.FStrategyRebateRate = "未输过"
	}
}

// GetUpdownQsbnStat 7updown求死不能
func (s *gameStatsService) GetUpdownQsbnStat(page, pageSize int, m, m2, m3 bson.M, startDate, endDate string) (
	stats []*entity.UpDownQsbnStat, count int, err error,
) {
	// GetUpdownQsbnStat
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)

	// 用户列表筛选
	var filterPlayerIds []string
	if len(m2) > 0 || len(m3) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				"_id":        "$userid",
				"bets":       bson.M{"$sum": "$bets"},
				"bet_rounds": bson.M{"$sum": "$bet_rounds"},
				"cash":       bson.M{"$sum": "$cash"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				"cash": "$cash",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = LHDPlayerStats.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("LHDPlayerStats error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			userid := r["_id"].(string)
			userids = append(userids, userid)
		}
		if len(userids) == 0 { // 未匹配到
			return
		}
		now := utils.BsonNow()
		m3_2 := []bson.M{
			{"$match": bson.M{"_id": bson.M{"$in": userids}}},
			{"$project": bson.M{
				"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
				"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
			}},
			{"$match": m3},
		}
		var r3_2 []bson.M
		err = PlayerUsers.Pipe(m3_2).All(&r3_2)
		if err != nil {
			beego.Error("lhd PlayerUsers error32:", err)
		} else if len(r3_2) == 0 { // 未匹配到
			return
		}
		for _, r := range r3_2 {
			userid := r["_id"].(string)
			filterPlayerIds = append(filterPlayerIds, userid)
		}
	}
	if len(filterPlayerIds) > 0 {
		m["userid"] = bson.M{"$in": filterPlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":  "$userid",
			"date": bson.M{"$max": "$date"},
			// "date_str":                bson.M{"$sum": "$date_str"},
			"strategy_players":      bson.M{"$max": "$strategy_players"},
			"strategy_times":        bson.M{"$sum": "$strategy_times"},
			"all_in_rounds":         bson.M{"$sum": "$all_in_rounds"},
			"strategy_multi_rounds": bson.M{"$sum": "$strategy_multi_rounds"},
			"strategy_bets":         bson.M{"$sum": "$strategy_bets"},
			"strategy_wins":         bson.M{"$sum": "$strategy_wins"},
			"strategy_loses":        bson.M{"$sum": "$strategy_loses"},
			"before_back_rate":      bson.M{"$avg": "$before_back_rate"},
			"after_back_rate":       bson.M{"$avg": "$after_back_rate"},
		}},
		{"$project": bson.M{
			"_id":                   "$_id",
			"date":                  "$date",
			"strategy_players":      "$strategy_players",
			"strategy_times":        "$strategy_times",
			"all_in_rounds":         "$all_in_rounds",
			"strategy_multi_rounds": "$strategy_multi_rounds",
			"strategy_bets":         "$strategy_bets",
			"strategy_wins":         "$strategy_wins",
			"strategy_loses":        "$strategy_loses",
			"before_back_rate":      "$before_back_rate",
			"after_back_rate":       "$after_back_rate",
		}},
		{"$match": m2},
		{"$count": "count"},
	}
	// 查总条数
	var rCount []bson.M
	err = UpdownQsbnStats.Pipe(m1).All(&rCount)
	if err != nil {
		return
	} else if len(rCount) > 0 {
		count = rCount[0]["count"].(int)
	}

	// 查列表数据
	m1 = m1[:len(m1)-1]
	m1 = append(m1, []bson.M{
		{"$sort": bson.M{"date": -1}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}...)
	err = UpdownQsbnStats.Pipe(m1).All(&stats)
	if err != nil {
		return
	}
	UpdownQsbnStatMapping(stats)

	// 查总汇数据
	mark := "--"
	summary := &entity.UpDownQsbnStat{}
	m1 = m1[:len(m1)-3]
	m1[1]["$group"].(bson.M)["_id"] = nil
	delete(m1[1]["$group"].(bson.M), "date")
	m1[1]["$group"].(bson.M)["userids"] = bson.M{"$addToSet": "$userid"}
	m1[1]["$group"].(bson.M)["strategy_players"] = bson.M{"$sum": "$strategy_players"} // 触发策略人数总汇
	m1[2]["$project"].(bson.M)["userids"] = "$userids"
	str, err := json.Marshal(m1)
	fmt.Println(err, string(str))
	var summarys []*entity.UpDownQsbnStat
	err = UpdownQsbnStats.Pipe(m1).All(&summarys)
	if err != nil {
		beego.Error("lhd summary error:", err)
		return
	} else if len(summarys) > 0 {
		summary = summarys[0]
		summary.No = fmt.Sprintf("%s-%s总汇", startDate, endDate)
		summary.Id = mark
		summary.FNickname = mark
		summary.FUserType = mark
		summary.FRegistArea = mark
		summary.FCtime = mark
		summary.FLoginTime = mark
		summary.FBackRate = mark
		UpdownQsbnStatSummaryMapping(summary)
	}
	if len(summarys) > 0 {
		stats = append([]*entity.UpDownQsbnStat{summary}, stats...)
	}
	return
}

func UpdownQsbnStatMapping(stats []*entity.UpDownQsbnStat) {
	var playerIds []string
	// var playerStats = make(map[string]*entity.CrashPlayerStat, len(stats))
	for _, stat := range stats {
		playerIds = append(playerIds, stat.Id)
		// playerStats[stat.Id] = stat
	}
	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": playerIds}}},
		{"$project": bson.M{
			"_id":          "$_id",
			"nickname":     "$nickname",
			"money":        "$money",
			"cash_out":     "$cash_out",
			"diamond":      "$diamond",
			"ctime":        "$ctime",
			"login_time":   "$login_time",
			"regist_area":  "$regist_area", //ab测试 0:A 1:B 2:C
			"state":        "$state",
			"lhd_strategy": "$lhd_strategy",
		}},
	}

	var r2 []*entity.PlayerUser
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("lhd users error:", err)
		return
	}
	var r2Map = make(map[string]*entity.PlayerUser, len(r2))
	for _, r := range r2 {
		userid := r.Userid
		r2Map[userid] = r
	}
	now := utils.BsonNow()
	for i, stat := range stats {
		r, ok := r2Map[stat.Id]
		if !ok {
			continue
		}
		// 玩家昵称	玩家类型 账号类型 注册时间 最后登陆时间 存活天数 流失天数
		nickname := r.Nickname
		money := r.Money
		ctime := r.Ctime
		login_time := r.LoginTime
		regist_area := r.RegistArea
		state := r.State

		// 将日期转成印度时间
		login_time, _ = ConvertToIndiaTime(login_time.Unix())
		ctime, _ = ConvertToIndiaTime(ctime.Unix())

		stat.No = strconv.Itoa(i + 1)
		stat.FNickname = nickname
		stat.FCtime = ctime.Format(utils.FORMAT)
		stat.FLoginTime = login_time.Format(utils.FORMAT)
		stat.FLiveDays = int(math.Ceil(login_time.Sub(ctime).Hours() / 24))
		stat.FLoseDays = int(math.Floor(now.Sub(login_time).Hours() / 24))

		switch regist_area {
		case 0:
			stat.FRegistArea = "A类"
		case 1:
			stat.FRegistArea = "B类"
		case 2:
			stat.FRegistArea = "C类"
		}

		switch state {
		case 1:
			stat.FUserType = "新手"
		case 2:
			if money == 0 {
				stat.FUserType = "零充"
			} else if money >= 20000 && money <= 99999 {
				stat.FUserType = "普充"
			} else if money >= 100000 && money <= 499999 {
				stat.FUserType = "小R"
			} else if money >= 500000 && money <= 999999 {
				stat.FUserType = "中R"
			} else if money >= 1000000 && money <= 9999999 {
				stat.FUserType = "大R"
			} else if money >= 10000000 {
				stat.FUserType = "超大R"
			}
		case 3:
			stat.FUserType = "平民"
		case 4:
			stat.FUserType = "泡沫"
		}

		if stat.StrategyPlayers > 0 {
			stat.StrategyPlayersAvg = fmt.Sprintf("%.2f", float64(stat.StrategyTimes)/float64(stat.StrategyPlayers))
		}
		if stat.AllInRounds > 0 {
			stat.StrategyRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyTimes)/float64(stat.AllInRounds)*100.0)
		}
		if stat.StrategyTimes > 0 {
			stat.StrategyMultiRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyMultiRounds)/float64(stat.StrategyTimes)*100.0)
		}
		stat.FStrategyBets = fmt.Sprintf("%.2f", float64(stat.StrategyBets)/100.0)
		if stat.StrategyTimes > 0 {
			stat.FStrategyBetsAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.StrategyBets, 100)/float64(stat.StrategyTimes))
		}
		stat.FStrategyWins = fmt.Sprintf("%.2f", float64(stat.StrategyWins)/100.0)
		stat.FStrategyLoses = fmt.Sprintf("%.2f", float64(stat.StrategyLoses)/100.0)
		stat.FStrategyCash = fmt.Sprintf("%.2f", float64(stat.StrategyWins+stat.StrategyLoses)/100.0)
		stat.FBeforeBackRate = fmt.Sprintf("%d", (stat.BeforeBackRate))
		stat.FAfterBackRate = fmt.Sprintf("%d", (stat.AfterBackRate))
		if r.Money > 0 {
			stat.FBackRate = fmt.Sprintf("%d", int((int64(r.CashOut)+r.Diamond)*10000/int64(r.Money)))
		}
	}
}

func UpdownQsbnStatSummaryMapping(stat *entity.UpDownQsbnStat) {
	stat.StrategyPlayers = int32(len(stat.Userids))
	if stat.StrategyPlayers > 0 {
		stat.StrategyPlayersAvg = fmt.Sprintf("%.2f", float64(stat.StrategyTimes)/float64(stat.StrategyPlayers))
	}
	if stat.AllInRounds > 0 {
		stat.StrategyRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyTimes)/float64(stat.AllInRounds)*100.0)
	}
	if stat.StrategyTimes > 0 {
		stat.StrategyMultiRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyMultiRounds)/float64(stat.StrategyTimes)*100.0)
	}
	stat.FStrategyBets = fmt.Sprintf("%.2f", float64(stat.StrategyBets)/100.0)
	if stat.StrategyTimes > 0 {
		stat.FStrategyBetsAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.StrategyBets, 100)/float64(stat.StrategyTimes))
	}
	stat.FStrategyWins = fmt.Sprintf("%.2f", float64(stat.StrategyWins)/100.0)
	stat.FStrategyLoses = fmt.Sprintf("%.2f", float64(stat.StrategyLoses)/100.0)
	stat.FStrategyCash = fmt.Sprintf("%.2f", float64(stat.StrategyWins+stat.StrategyLoses)/100.0)
	stat.FBeforeBackRate = fmt.Sprintf("%d", (stat.BeforeBackRate))
	stat.FAfterBackRate = fmt.Sprintf("%d", (stat.AfterBackRate))
}

func (s *gameStatsService) GetUpdownLkyhStat(page, pageSize int, m, m2, m3 bson.M, startDate, endDate string) (
	stats []*entity.UpdownLkyhStat, count int, err error,
) {
	// LHDXxscStat
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)

	// 用户列表筛选
	var filterPlayerIds []string
	if len(m2) > 0 || len(m3) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				"_id":    "$userid",
				"bets":   bson.M{"$sum": "$bets"},
				"rounds": bson.M{"$sum": "$rounds"},
				"cash":   bson.M{"$sum": "$cash"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				"cash": "$cash",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = UpdownLkyhStats.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetLHDLkyhStat error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			userid := r["_id"].(string)
			userids = append(userids, userid)
		}
		if len(userids) == 0 { // 未匹配到
			return
		}
		now := utils.BsonNow()
		m3_2 := []bson.M{
			{"$match": bson.M{"_id": bson.M{"$in": userids}}},
			{"$project": bson.M{
				"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
				"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
			}},
			{"$match": m3},
		}
		var r3_2 []bson.M
		err = PlayerUsers.Pipe(m3_2).All(&r3_2)
		if err != nil {
			beego.Error("lhd LkyhStat error32:", err)
		} else if len(r3_2) == 0 { // 未匹配到
			return
		}
		for _, r := range r3_2 {
			userid := r["_id"].(string)
			filterPlayerIds = append(filterPlayerIds, userid)
		}
	}
	if len(filterPlayerIds) > 0 {
		m["userid"] = bson.M{"$in": filterPlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":  "$userid",
			"date": bson.M{"$max": "$date"},
			// "date_str":                bson.M{"$sum": "$date_str"},
			"bstimes":               bson.M{"$max": "$bstimes"},
			"bs_days":               bson.M{"$sum": "$bs_days"},
			"trigger_times":         bson.M{"$sum": "$trigger_times"},
			"tz":                    bson.M{"$max": "$tz"},
			"nz":                    bson.M{"$max": "$nz"},
			"bs_bets":               bson.M{"$sum": "$bs_bets"},
			"yz_number":             bson.M{"$sum": "$yz_number"},
			"yz_rounds":             bson.M{"$sum": "$yz_rounds"},
			"yz_bets":               bson.M{"$sum": "$yz_bets"},
			"rz":                    bson.M{"$sum": "$rz"},
			"strategy_multi_rounds": bson.M{"$sum": "$strategy_multi_rounds"},
			"bets":                  bson.M{"$sum": "$bets"},
			"rounds":                bson.M{"$sum": "$rounds"},
			"win_bets":              bson.M{"$sum": "$win_bets"},
			"lose_bets":             bson.M{"$sum": "$lose_bets"},
			"strategy_bets":         bson.M{"$sum": "$strategy_bets"},
			"strategy_wins":         bson.M{"$sum": "$strategy_wins"},
			"strategy_loses":        bson.M{"$sum": "$strategy_loses"},
			"strategy_win_rounds":   bson.M{"$sum": "$strategy_win_rounds"},
			"strategy_lose_rounds":  bson.M{"$sum": "$strategy_lose_rounds"},
			"strategy_times":        bson.M{"$sum": "$strategy_times"},
			"win_rounds":            bson.M{"$sum": "$win_rounds"},
			"lose_rounds":           bson.M{"$sum": "$lose_rounds"},
		}},
		{"$project": bson.M{
			"_id":                   "$_id",
			"date":                  "$date",
			"bstimes":               "$bstimes",
			"bs_days":               "$bs_days",
			"trigger_times":         "$trigger_times",
			"tz":                    "$tz",
			"nz":                    "$nz",
			"bs_bets":               "$bs_bets",
			"yz_number":             "$yz_number",
			"yz_rounds":             "$yz_rounds",
			"yz_bets":               "$yz_bets",
			"rz":                    "$rz",
			"strategy_multi_rounds": "$strategy_multi_rounds",
			"bets":                  "$bets",
			"rounds":                "$rounds",
			"win_bets":              "$win_bets",
			"lose_bets":             "$lose_bets",
			"strategy_bets":         "$strategy_bets",
			"strategy_wins":         "$strategy_wins",
			"strategy_loses":        "$strategy_loses",
			"strategy_win_rounds":   "$strategy_win_rounds",
			"strategy_lose_rounds":  "$strategy_lose_rounds",
			"strategy_times":        "$strategy_times",
			"win_rounds":            "$win_rounds",
			"lose_rounds":           "$lose_rounds",
		}},
		{"$match": m2},
		{"$count": "count"},
	}
	// 查总条数
	var rCount []bson.M
	err = UpdownLkyhStats.Pipe(m1).All(&rCount)
	if err != nil {
		return
	} else if len(rCount) > 0 {
		count = rCount[0]["count"].(int)
	}

	// 查列表数据
	m1 = m1[:len(m1)-1]
	m1 = append(m1, []bson.M{
		{"$sort": bson.M{"date": -1}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}...)
	err = UpdownLkyhStats.Pipe(m1).All(&stats)
	if err != nil {
		return
	}
	updownLkyhStatMapping(stats, m)

	// 查总汇数据
	mark := "--"
	summary := &entity.UpdownLkyhStat{}
	m1 = m1[:len(m1)-3]
	m1[1]["$group"].(bson.M)["_id"] = nil
	delete(m1[1]["$group"].(bson.M), "date")
	m1[1]["$group"].(bson.M)["userids"] = bson.M{"$addToSet": "$userid"}
	m1[1]["$group"].(bson.M)["bstimes"] = bson.M{"$sum": "$bstimes"} // 触发策略人数总汇
	m1[1]["$group"].(bson.M)["tz"] = bson.M{"$sum": "$tz"}           // 触发策略人数总汇
	m1[1]["$group"].(bson.M)["nz"] = bson.M{"$sum": "$nz"}           // 触发策略人数总汇

	m1[2]["$project"].(bson.M)["userids"] = "$userids"
	str, err := json.Marshal(m1)
	fmt.Println(err, string(str))
	var summarys []*entity.UpdownLkyhStat
	err = UpdownLkyhStats.Pipe(m1).All(&summarys)
	if err != nil {
		beego.Error("lhd summary error:", err)
		return
	} else if len(summarys) > 0 {
		summary = summarys[0]
		summary.No = fmt.Sprintf("%s-%s总汇", startDate, endDate)
		summary.Id = mark
		summary.FNickname = mark
		summary.FUserType = mark
		summary.FRegistArea = mark
		summary.FCtime = mark
		summary.FLoginTime = mark
		// summary.FBackRate = mark
		updownLkyhStatSummaryMapping(summary, m)
	}
	if len(summarys) > 0 {
		stats = append([]*entity.UpdownLkyhStat{summary}, stats...)
	}
	return
}

func updownLkyhStatMapping(stats []*entity.UpdownLkyhStat, m bson.M) {
	var playerIds []string
	// var playerStats = make(map[string]*entity.CrashPlayerStat, len(stats))
	for _, stat := range stats {
		playerIds = append(playerIds, stat.Id)
		// playerStats[stat.Id] = stat
	}
	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": playerIds}}},
		{"$project": bson.M{
			"_id":          "$_id",
			"nickname":     "$nickname",
			"money":        "$money",
			"cash_out":     "$cash_out",
			"diamond":      "$diamond",
			"ctime":        "$ctime",
			"login_time":   "$login_time",
			"regist_area":  "$regist_area", //ab测试 0:A 1:B 2:C
			"state":        "$state",
			"lhd_strategy": "$lhd_strategy",
		}},
	}

	var r2 []*entity.PlayerUser
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("lhd users error:", err)
		return
	}
	var r2Map = make(map[string]*entity.PlayerUser, len(r2))
	for _, r := range r2 {
		userid := r.Userid
		r2Map[userid] = r
	}
	// 查询总数据
	var uplist []entity.UpDownPlayerStat
	n := bson.M{}
	// 手动复制 m 到 n
	for k, v := range m {
		n[k] = v
	}
	n["userid"] = bson.M{"$in": playerIds}
	m3 := []bson.M{
		{"$match": n},
		{"$group": bson.M{
			"_id":        "$userid",
			"bet_rounds": bson.M{"$sum": "$bet_rounds"},
			"bets":       bson.M{"$sum": "$bets"},
			"win_bets":   bson.M{"$sum": "$win_bets"},
			"lose_bets":  bson.M{"$sum": "$lose_bets"},
			"wins":       bson.M{"$sum": "$wins"},
			"loses":      bson.M{"$sum": "$loses"},
		}},
	}
	UpdownPlayerStats.Pipe(m3).All(&uplist)

	now := utils.BsonNow()
	for i, stat := range stats {
		r, ok := r2Map[stat.Id]
		if !ok {
			continue
		}
		// 玩家昵称	玩家类型 账号类型 注册时间 最后登陆时间 存活天数 流失天数
		nickname := r.Nickname
		money := r.Money
		ctime := r.Ctime
		login_time := r.LoginTime
		regist_area := r.RegistArea
		state := r.State

		// 将日期转成印度时间
		login_time, _ = ConvertToIndiaTime(login_time.Unix())
		ctime, _ = ConvertToIndiaTime(ctime.Unix())

		stat.No = strconv.Itoa(i + 1)
		stat.FNickname = nickname
		stat.FCtime = ctime.Format(utils.FORMAT)
		stat.FLoginTime = login_time.Format(utils.FORMAT)
		stat.FLiveDays = int(math.Ceil(login_time.Sub(ctime).Hours() / 24))
		stat.FLoseDays = int(math.Floor(now.Sub(login_time).Hours() / 24))

		switch regist_area {
		case 0:
			stat.FRegistArea = "A类"
		case 1:
			stat.FRegistArea = "B类"
		case 2:
			stat.FRegistArea = "C类"
		}

		switch state {
		case 1:
			stat.FUserType = "新手"
		case 2:
			if money == 0 {
				stat.FUserType = "零充"
			} else if money >= 20000 && money <= 99999 {
				stat.FUserType = "普充"
			} else if money >= 100000 && money <= 499999 {
				stat.FUserType = "小R"
			} else if money >= 500000 && money <= 999999 {
				stat.FUserType = "中R"
			} else if money >= 1000000 && money <= 9999999 {
				stat.FUserType = "大R"
			} else if money >= 10000000 {
				stat.FUserType = "超大R"
			}
		case 3:
			stat.FUserType = "平民"
		case 4:
			stat.FUserType = "泡沫"
		}
		if stat.BSDays != 0 {
			stat.BSNumberAvg = fmt.Sprintf("%.2f", float64(stat.BSTimes)/float64(stat.BSDays))
		}
		if stat.BSTimes != 0 {
			stat.BSRate = fmt.Sprintf("%.2f%%", float64(stat.TZ)/float64(stat.BSTimes)*100.0)
			stat.BSBetAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.BSBets, 100)/float64(stat.BSTimes))
		}
		if stat.RZ != 0 {
			stat.YZRate = fmt.Sprintf("%.2f%%", float64(stat.YZNumber)/float64(stat.RZ)*100.0)
		}
		if stat.YZRounds != 0 {
			stat.YZBetAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.YZBets, 100)/float64(stat.YZRounds))
		}
		if stat.StrategyTimes > 0 {
			stat.StrategyMultiRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyMultiRounds)/float64(stat.StrategyTimes)*100.0)
			stat.StrategyWinRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyWinRounds)/float64(stat.StrategyTimes)*100.0)
		}
		stat.FStrategyBets = fmt.Sprintf("%.2f", float64(stat.StrategyBets)/100.0)
		if stat.StrategyTimes > 0 {
			stat.FStrategyBetsAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.StrategyBets, 100)/float64(stat.StrategyTimes))
		}
		if -stat.StrategyLoses != 0 {
			stat.StrategyRewardRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.StrategyWins, 100)/ComputeFloat(-stat.StrategyLoses, 100))*100.0)
		}
		if len(uplist) > 0 {
			for _, c := range uplist {
				if c.Id == stat.Id {
					stat.Bets = c.Bets
					stat.Rounds = int64(c.BetRounds)
					stat.WinBets = c.Wins
					stat.LoseBets = c.Loses
					break
				}
			}
		}
		if stat.Rounds != 0 {
			stat.RoundBetAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.Bets, 100)/float64(stat.Rounds))
			stat.TotalWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.Rounds)*100.0)
		}
		stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
		stat.FStrategyWins = fmt.Sprintf("%.2f", float64(stat.StrategyWins)/100.0)
		stat.FStrategyLoses = fmt.Sprintf("%.2f", float64(stat.StrategyLoses)/100.0)
		stat.FStrategyCash = fmt.Sprintf("%.2f", float64(stat.StrategyWins+stat.StrategyLoses)/100.0)
		if -stat.LoseBets != 0 {
			stat.TotalRewardRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.WinBets, 100)/ComputeFloat(-stat.LoseBets, 100))*100.0)
		}
		if r.Money > 0 {
			// stat.PlayerRewardRate = fmt.Sprintf("%d", int((int64(r.CashOut)+r.Diamond)*10000/int64(r.Money)))
			stat.PlayerRewardRate = fmt.Sprintf("%.2f%%", ComputeFloat(int64(r.CashOut+int32(r.Diamond)), 100)/ComputeFloat(int64(r.Money), 100)*100.0)
		} else {
			stat.PlayerRewardRate = "未充值"
		}
	}
}

func updownLkyhStatSummaryMapping(stat *entity.UpdownLkyhStat, m bson.M) {
	// 查询总数据
	upinfo := new(entity.UpDownPlayerStat)
	n := bson.M{}
	// 手动复制 m 到 n
	for k, v := range m {
		n[k] = v
	}
	n["userid"] = bson.M{"$in": stat.Userids}
	m3 := []bson.M{
		{"$match": n},
		{"$group": bson.M{
			"_id":        nil,
			"bet_rounds": bson.M{"$sum": "$bet_rounds"},
			"bets":       bson.M{"$sum": "$bets"},
			"win_bets":   bson.M{"$sum": "$win_bets"},
			"lose_bets":  bson.M{"$sum": "$lose_bets"},
			"wins":       bson.M{"$sum": "$wins"},
			"loses":      bson.M{"$sum": "$loses"},
		}},
	}
	UpdownPlayerStats.Pipe(m3).One(&upinfo)
	if stat.BSDays != 0 {
		stat.BSNumberAvg = fmt.Sprintf("%.2f", float64(stat.BSTimes)/float64(stat.BSDays))
	}
	if stat.BSTimes != 0 {
		stat.BSRate = fmt.Sprintf("%.2f%%", float64(stat.TZ)/float64(stat.BSTimes)*100.0)
		stat.BSBetAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.BSBets, 100)/float64(stat.BSTimes))
	}
	if stat.RZ != 0 {
		stat.YZRate = fmt.Sprintf("%.2f%%", float64(stat.YZNumber)/float64(stat.RZ)*100.0)
	}
	if stat.YZRounds != 0 {
		stat.YZBetAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.YZBets, 100)/float64(stat.YZRounds))
	}
	if stat.StrategyTimes > 0 {
		stat.StrategyMultiRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyMultiRounds)/float64(stat.StrategyTimes)*100.0)
		stat.StrategyWinRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyWinRounds)/float64(stat.StrategyTimes)*100.0)
	}
	stat.FStrategyBets = fmt.Sprintf("%.2f", float64(stat.StrategyBets)/100.0)
	if stat.StrategyTimes > 0 {
		stat.FStrategyBetsAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.StrategyBets, 100)/float64(stat.StrategyTimes))
	}
	if -stat.StrategyLoses != 0 {
		stat.StrategyRewardRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.StrategyWins, 100)/ComputeFloat(-stat.StrategyLoses, 100))*100.0)
	}
	// 查询总局数
	stat.Bets = upinfo.Bets
	stat.Rounds = int64(upinfo.BetRounds)
	stat.WinBets = upinfo.Wins
	stat.LoseBets = upinfo.Loses
	if stat.Rounds != 0 {
		stat.RoundBetAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.Bets, 100)/float64(stat.Rounds))
		stat.TotalWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.Rounds)*100.0)
	}
	stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
	stat.FStrategyWins = fmt.Sprintf("%.2f", float64(stat.StrategyWins)/100.0)
	stat.FStrategyLoses = fmt.Sprintf("%.2f", float64(stat.StrategyLoses)/100.0)
	stat.FStrategyCash = fmt.Sprintf("%.2f", float64(stat.StrategyWins+stat.StrategyLoses)/100.0)
	if -stat.LoseBets != 0 {
		stat.TotalRewardRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.WinBets, 100)/ComputeFloat(-stat.LoseBets, 100))*100.0)
	}
}

/*
AB游戏策略
*/
func (s *gameStatsService) GetABStatPlayers(page, pageSize int, m, m2, m3 bson.M, startDate, endDate string) (
	stats []*entity.ABPlayerStat, count int, err error,
) {
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)

	// 用户列表筛选
	var filterPlayerIds []string
	if len(m2) > 0 || len(m3) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				"_id":        "$userid",
				"bets":       bson.M{"$sum": "$bets"},
				"bet_rounds": bson.M{"$sum": "$bet_rounds"},
				"cash":       bson.M{"$sum": "$cash"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				"cash": "$cash",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = ABPlayerStats.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetABStatPlayers error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			userid := r["_id"].(string)
			userids = append(userids, userid)
		}
		if len(userids) == 0 { // 未匹配到
			return
		}
		now := utils.BsonNow()
		m3_2 := []bson.M{
			{"$match": bson.M{"_id": bson.M{"$in": userids}}},
			{"$project": bson.M{
				"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
				"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
			}},
			{"$match": m3},
		}
		var r3_2 []bson.M
		err = PlayerUsers.Pipe(m3_2).All(&r3_2)
		if err != nil {
			beego.Error("GetABStatPlayers error32:", err)
		} else if len(r3_2) == 0 { // 未匹配到
			return
		}
		for _, r := range r3_2 {
			userid := r["_id"].(string)
			filterPlayerIds = append(filterPlayerIds, userid)
		}
	}
	if len(filterPlayerIds) > 0 {
		m["userid"] = bson.M{"$in": filterPlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":               "$userid",
			"date":              bson.M{"$max": "$date"},
			"all_rounds":        bson.M{"$sum": "$all_rounds"},
			"game_times":        bson.M{"$sum": "$game_times"},
			"bets":              bson.M{"$sum": "$bets"},
			"bet_rounds":        bson.M{"$sum": "$bet_rounds"},
			"win_rounds":        bson.M{"$sum": "$win_rounds"},
			"lose_rounds":       bson.M{"$sum": "$lose_rounds"},
			"tie_rounds":        bson.M{"$sum": "$tie_rounds"},
			"win_bets":          bson.M{"$sum": "$win_bets"},
			"lose_bets":         bson.M{"$sum": "$lose_bets"},
			"tie_bets":          bson.M{"$sum": "$tie_bets"},
			"wins":              bson.M{"$sum": "$wins"},
			"loses":             bson.M{"$sum": "$loses"},
			"cash":              bson.M{"$sum": "$cash"},
			"side_winner1":      bson.M{"$sum": "$side_winner1"},
			"side_winner2":      bson.M{"$sum": "$side_winner2"},
			"side_winner3":      bson.M{"$sum": "$side_winner3"},
			"side_winner4":      bson.M{"$sum": "$side_winner4"},
			"side_winner5":      bson.M{"$sum": "$side_winner5"},
			"side_winner6":      bson.M{"$sum": "$side_winner6"},
			"side_winner7":      bson.M{"$sum": "$side_winner7"},
			"side_winner8":      bson.M{"$sum": "$side_winner8"},
			"side_winner9":      bson.M{"$sum": "$side_winner9"},
			"side_winner10":     bson.M{"$sum": "$side_winner10"},
			"seat_bets1":        bson.M{"$sum": "$seat_bets1"},
			"seat_bets2":        bson.M{"$sum": "$seat_bets2"},
			"seat_bets3":        bson.M{"$sum": "$seat_bets3"},
			"seat_bets4":        bson.M{"$sum": "$seat_bets4"},
			"seat_bets5":        bson.M{"$sum": "$seat_bets5"},
			"seat_bets6":        bson.M{"$sum": "$seat_bets6"},
			"seat_bets7":        bson.M{"$sum": "$seat_bets7"},
			"seat_bets8":        bson.M{"$sum": "$seat_bets8"},
			"seat_bets9":        bson.M{"$sum": "$seat_bets9"},
			"seat_bets10":       bson.M{"$sum": "$seat_bets10"},
			"player_win_seat1":  bson.M{"$sum": "$player_win_seat1"},
			"player_win_seat2":  bson.M{"$sum": "$player_win_seat2"},
			"player_win_seat3":  bson.M{"$sum": "$player_win_seat3"},
			"player_win_seat4":  bson.M{"$sum": "$player_win_seat4"},
			"player_win_seat5":  bson.M{"$sum": "$player_win_seat5"},
			"player_win_seat6":  bson.M{"$sum": "$player_win_seat6"},
			"player_win_seat7":  bson.M{"$sum": "$player_win_seat7"},
			"player_win_seat8":  bson.M{"$sum": "$player_win_seat8"},
			"player_win_seat9":  bson.M{"$sum": "$player_win_seat9"},
			"player_win_seat10": bson.M{"$sum": "$player_win_seat10"},
			"player_multis":     bson.M{"$sum": "$player_multis"},
			"round_bet_avg":     bson.M{"$sum": "$round_bet_avg"},
			"money":             bson.M{"$sum": "$money"},
			"observe_rounds":    bson.M{"$sum": "$observe_rounds"},
		}},
		{"$project": bson.M{
			"_id":               "$_id",
			"all_rounds":        "$all_rounds",
			"game_times":        "$game_times",
			"bets":              "$bets",
			"bet_rounds":        "$bet_rounds",
			"win_rounds":        "$win_rounds",
			"lose_rounds":       "$lose_rounds",
			"tie_rounds":        "$tie_rounds",
			"win_bets":          "$win_bets",
			"lose_bets":         "$lose_bets",
			"tie_bets":          "$tie_bets",
			"wins":              "$wins",
			"loses":             "$loses",
			"cash":              "$cash",
			"side_winner1":      "$side_winner1",
			"side_winner2":      "$side_winner2",
			"side_winner3":      "$side_winner3",
			"side_winner4":      "$side_winner4",
			"side_winner5":      "$side_winner5",
			"side_winner6":      "$side_winner6",
			"side_winner7":      "$side_winner7",
			"side_winner8":      "$side_winner8",
			"side_winner9":      "$side_winner9",
			"side_winner10":     "$side_winner10",
			"seat_bets1":        "$seat_bets1",
			"seat_bets2":        "$seat_bets2",
			"seat_bets3":        "$seat_bets3",
			"seat_bets4":        "$seat_bets4",
			"seat_bets5":        "$seat_bets5",
			"seat_bets6":        "$seat_bets6",
			"seat_bets7":        "$seat_bets7",
			"seat_bets8":        "$seat_bets8",
			"seat_bets9":        "$seat_bets9",
			"seat_bets10":       "$seat_bets10",
			"player_win_seat1":  "$player_win_seat1",
			"player_win_seat2":  "$player_win_seat2",
			"player_win_seat3":  "$player_win_seat3",
			"player_win_seat4":  "$player_win_seat4",
			"player_win_seat5":  "$player_win_seat5",
			"player_win_seat6":  "$player_win_seat6",
			"player_win_seat7":  "$player_win_seat7",
			"player_win_seat8":  "$player_win_seat8",
			"player_win_seat9":  "$player_win_seat9",
			"player_win_seat10": "$player_win_seat10",
			"player_multis":     "$player_multis",
			"money":             "$money",
			"observe_rounds":    "$observe_rounds",
			// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}}, // "can't $divide by zero"
			"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
		}},
		{"$match": m2},
		{"$count": "count"},
	}
	// 查总条数
	var rCount []bson.M
	err = ABPlayerStats.Pipe(m1).All(&rCount)
	if err != nil {
		return
	} else if len(rCount) > 0 {
		count = rCount[0]["count"].(int)
	}

	// 查列表数据
	m1 = m1[:len(m1)-1]
	m1 = append(m1, []bson.M{
		{"$sort": bson.M{"date": -1}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}...)
	err = ABPlayerStats.Pipe(m1).All(&stats)
	if err != nil {
		return
	}
	ABStatPlayersMapping(stats)

	// 查总汇数据
	mark := "--"
	summary := &entity.ABPlayerStat{}
	m1 = m1[:len(m1)-3]
	m1[1]["$group"].(bson.M)["_id"] = nil
	delete(m1[1]["$group"].(bson.M), "date")
	m1[1]["$group"].(bson.M)["userids"] = bson.M{"$addToSet": "$userid"}
	m1[2]["$project"].(bson.M)["userids"] = "$userids"
	var summarys []*entity.ABPlayerStat
	err = ABPlayerStats.Pipe(m1).All(&summarys)
	if err != nil {
		return
	} else if len(summarys) > 0 {
		summary = summarys[0]
		summary.No = fmt.Sprintf("%s-%s总汇", startDate, endDate)
		summary.Id = mark
		summary.FNickname = mark
		summary.FUserType = mark
		summary.FRegistArea = mark
		summary.FCtime = mark
		summary.FLoginTime = mark
		// summary.FLiveDays = mark
		// summary.FLoseDays = mark
		ABStatPlayersSummaryMapping(summary)
	}
	if len(summarys) > 0 {
		stats = append([]*entity.ABPlayerStat{summary}, stats...)
	}
	return
}

func ABStatPlayersMapping(stats []*entity.ABPlayerStat) {
	var playerIds []string
	// var playerStats = make(map[string]*entity.CrashPlayerStat, len(stats))
	for _, stat := range stats {
		playerIds = append(playerIds, stat.Id)
		// playerStats[stat.Id] = stat
	}
	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": playerIds}}},
		{"$project": bson.M{
			"_id":         "$_id",
			"nickname":    "$nickname",
			"money":       "$money",
			"cash_out":    "$cash_out",
			"diamond":     "$diamond",
			"ctime":       "$ctime",
			"login_time":  "$login_time",
			"regist_area": "$regist_area", //ab测试 0:A 1:B 2:C
			"state":       "$state",
		}},
	}
	var r2 []bson.M
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("lhd users error:", err)
		return
	}
	var r2Map = make(map[string]bson.M, len(r2))
	for _, r := range r2 {
		userid := r["_id"].(string)
		r2Map[userid] = r
	}
	now := utils.BsonNow()
	for i, stat := range stats {
		r, ok := r2Map[stat.Id]
		if !ok {
			continue
		}
		// 玩家昵称	玩家类型 账号类型 注册时间 最后登陆时间 存活天数 流失天数
		// userid := r["_id"].(string)
		nickname := r["nickname"].(string)
		money := r["money"].(int)
		cash_out := r["cash_out"].(int)
		diamond := r["diamond"].(int64)
		ctime := r["ctime"].(time.Time)
		login_time := r["login_time"].(time.Time)
		regist_area := r["regist_area"].(int)
		state := r["state"].(int)

		// 将日期转成印度时间
		login_time, _ = ConvertToIndiaTime(login_time.Unix())
		ctime, _ = ConvertToIndiaTime(ctime.Unix())

		stat.No = strconv.Itoa(i + 1)
		stat.FNickname = nickname
		stat.Money = uint32(money)
		stat.FCtime = ctime.Format(utils.FORMAT)
		stat.FLoginTime = login_time.Format(utils.FORMAT)
		stat.FLiveDays = int(math.Ceil(login_time.Sub(ctime).Hours() / 24))
		stat.FLoseDays = int(math.Floor(now.Sub(login_time).Hours() / 24))

		// summary.FLiveDays += stat.FLiveDays
		// summary.FLoseDays += stat.FLoseDays
		// summary.AllRounds += stat.AllRounds
		// summary.GameTimes += stat.GameTimes
		// summary.Bets += stat.Bets
		// summary.BetRounds += stat.BetRounds
		// summary.WinRounds += stat.WinRounds
		// summary.LoseRounds += stat.LoseRounds
		// summary.TieRounds += stat.TieRounds
		// summary.WinBets += stat.WinBets
		// summary.LoseBets += stat.LoseBets
		// summary.TieBets += stat.TieBets
		// summary.Wins += stat.Wins
		// summary.Loses += stat.Loses
		// summary.Cash += stat.Cash
		// summary.MulpitleSum += stat.MulpitleSum
		// summary.WinEscapeMulpitleSum += stat.WinEscapeMulpitleSum
		// summary.WinMulpitleSum += stat.WinMulpitleSum
		// summary.LoseMulpitleSum += stat.LoseMulpitleSum
		// summary.Mulpitles = append(summary.Mulpitles, stat.Mulpitles...)
		// summary.WinEscapeMulpitles = append(summary.WinEscapeMulpitles, stat.WinEscapeMulpitles...)
		// summary.WinMulpitles = append(summary.WinMulpitles, stat.WinMulpitles...)
		// summary.LoseMulpitles = append(summary.LoseMulpitles, stat.LoseMulpitles...)
		// summary.ObserveRounds += stat.ObserveRounds
		// summary.SMoney += money
		// summary.SCashOut += cash_out
		// summary.SDiamond += diamond

		switch regist_area {
		case 0:
			stat.FRegistArea = "A类"
		case 1:
			stat.FRegistArea = "B类"
		case 2:
			stat.FRegistArea = "C类"
		}

		switch state {
		case 1:
			stat.FUserType = "新手"
		case 2:
			if money == 0 {
				stat.FUserType = "零充"
			} else if money >= 20000 && money <= 99999 {
				stat.FUserType = "普充"
			} else if money >= 100000 && money <= 499999 {
				stat.FUserType = "小R"
			} else if money >= 500000 && money <= 999999 {
				stat.FUserType = "中R"
			} else if money >= 1000000 && money <= 9999999 {
				stat.FUserType = "大R"
			} else if money >= 10000000 {
				stat.FUserType = "超大R"
			}
		case 3:
			stat.FUserType = "平民"
		case 4:
			stat.FUserType = "泡沫"
		}

		stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
		if stat.BetRounds > 0 {
			// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.BetRounds)/100.0)
			stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
		}
		stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
		stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
		stat.FTieBets = fmt.Sprintf("%.2f", float64(stat.TieBets)/100.0)
		if stat.BetRounds > 0 {
			stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
		}
		if stat.Bets > 0 {
			// stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
			stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
		}
		if stat.WinRounds > 0 {
			stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
		}
		if stat.TieRounds > 0 {
			stat.FTieBetsAvg = fmt.Sprintf("%.2f", float64(stat.TieBets)/float64(stat.TieRounds)/100.0)
		}
		stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
		stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
		stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)

		if stat.WinRounds > 0 {
			stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
		}
		// 下注开奖
		if stat.BetRounds > 0 {
			stat.SideWinnerRate1 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner1)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate2 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner2)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate3 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner3)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate4 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner4)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate5 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner5)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate6 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner6)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate7 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner7)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate8 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner8)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate9 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner9)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate10 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner10)/float64(stat.BetRounds)*100)
		}

		//玩家下注
		if stat.BetRounds > 0 {
			stat.SeatBetsRate1 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets1)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate2 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets2)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate3 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets3)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate4 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets4)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate5 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets5)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate6 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets6)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate7 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets7)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate8 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets8)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate9 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets9)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate10 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets10)/float64(stat.BetRounds)*100)
		}
		// 玩家下注中奖
		if stat.SeatBets1 > 0 {
			stat.PlayerWinSeatRate1 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat1)/float64(stat.SeatBets1)*100)
		}
		if stat.SeatBets2 > 0 {
			stat.PlayerWinSeatRate2 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat2)/float64(stat.SeatBets2)*100)
		}
		if stat.SeatBets3 > 0 {
			stat.PlayerWinSeatRate3 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat3)/float64(stat.SeatBets3)*100)
		}
		if stat.SeatBets4 > 0 {
			stat.PlayerWinSeatRate4 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat4)/float64(stat.SeatBets4)*100)
		}
		if stat.SeatBets5 > 0 {
			stat.PlayerWinSeatRate5 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat5)/float64(stat.SeatBets5)*100)
		}
		if stat.SeatBets6 > 0 {
			stat.PlayerWinSeatRate6 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat6)/float64(stat.SeatBets6)*100)
		}
		if stat.SeatBets7 > 0 {
			stat.PlayerWinSeatRate7 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat7)/float64(stat.SeatBets7)*100)
		}
		if stat.SeatBets8 > 0 {
			stat.PlayerWinSeatRate8 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat8)/float64(stat.SeatBets8)*100)
		}
		if stat.SeatBets9 > 0 {
			stat.PlayerWinSeatRate9 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat9)/float64(stat.SeatBets9)*100)
		}
		if stat.SeatBets10 > 0 {
			stat.PlayerWinSeatRate10 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat10)/float64(stat.SeatBets10)*100)
		}
		// 多门率
		if stat.BetRounds > 0 {
			stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerMultis)/float64(stat.BetRounds)*100)
		}

		// 观察局
		if stat.BetRounds > 0 {
			stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
		}
		if stat.BetRounds > 0 {
			stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
		}
		stat.FMoney = fmt.Sprintf("%.2f", float64(money)/100)
		stat.FCashOut = fmt.Sprintf("%.2f", float64(cash_out)/100)
		stat.FWinMoney = fmt.Sprintf("%.2f", float64(int(diamond)+cash_out-money)/100)
	}
}

func ABStatPlayersSummaryMapping(stat *entity.ABPlayerStat) {
	// 查用户基本信息
	// m2 := []bson.M{
	// 	{"$match": bson.M{"_id": bson.M{"$in": stat.Userids}}},
	// 	{"$group": bson.M{
	// 		"_id":      nil,
	// 		"money":    bson.M{"$sum": "$money"},
	// 		"cash_out": bson.M{"$sum": "$cash_out"},
	// 		"diamond":  bson.M{"$sum": "$diamond"},
	// 	}},
	// }
	// var r2 []bson.M
	// err := PlayerUsers.Pipe(m2).All(&r2)
	// if err != nil {
	// 	beego.Error("lhd users error:", err)
	// 	return
	// } else if len(r2) > 0 {
	// 	money := r2[0]["money"].(int)
	// 	cash_out := r2[0]["cash_out"].(int)
	// 	diamond := r2[0]["diamond"].(int64)
	// 	stat.SMoney = money
	// 	stat.SCashOut = cash_out
	// 	stat.SDiamond = diamond
	// }

	stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
	if stat.BetRounds > 0 {
		stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.BetRounds)/100.0)
		// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
	}

	stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
	stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
	stat.FTieBets = fmt.Sprintf("%.2f", float64(stat.TieBets)/100.0)
	if stat.BetRounds > 0 {
		stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
	}
	if stat.Bets > 0 {
		// stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
		stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
	}
	if stat.WinRounds > 0 {
		stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
	}
	if stat.TieRounds > 0 {
		stat.FTieBetsAvg = fmt.Sprintf("%.2f", float64(stat.TieBets)/float64(stat.TieRounds)/100.0)
	}
	stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
	stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
	stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)
	if stat.WinRounds > 0 {
		stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
	}
	// 下注开奖
	if stat.BetRounds > 0 {
		stat.SideWinnerRate1 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner1)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate2 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner2)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate3 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner3)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate4 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner4)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate5 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner5)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate6 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner6)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate7 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner7)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate8 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner8)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate9 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner9)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate10 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner10)/float64(stat.BetRounds)*100)
	}

	//玩家下注
	if stat.BetRounds > 0 {
		stat.SeatBetsRate1 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets1)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate2 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets2)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate3 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets3)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate4 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets4)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate5 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets5)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate6 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets6)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate7 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets7)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate8 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets8)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate9 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets9)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate10 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets10)/float64(stat.BetRounds)*100)
	}
	// 玩家下注中奖
	if stat.SeatBets1 > 0 {
		stat.PlayerWinSeatRate1 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat1)/float64(stat.SeatBets1)*100)
	}
	if stat.SeatBets2 > 0 {
		stat.PlayerWinSeatRate2 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat2)/float64(stat.SeatBets2)*100)
	}
	if stat.SeatBets3 > 0 {
		stat.PlayerWinSeatRate3 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat3)/float64(stat.SeatBets3)*100)
	}
	if stat.SeatBets4 > 0 {
		stat.PlayerWinSeatRate4 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat4)/float64(stat.SeatBets4)*100)
	}
	if stat.SeatBets5 > 0 {
		stat.PlayerWinSeatRate5 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat5)/float64(stat.SeatBets5)*100)
	}
	if stat.SeatBets6 > 0 {
		stat.PlayerWinSeatRate6 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat6)/float64(stat.SeatBets6)*100)
	}
	if stat.SeatBets7 > 0 {
		stat.PlayerWinSeatRate7 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat7)/float64(stat.SeatBets7)*100)
	}
	if stat.SeatBets8 > 0 {
		stat.PlayerWinSeatRate8 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat8)/float64(stat.SeatBets8)*100)
	}
	if stat.SeatBets9 > 0 {
		stat.PlayerWinSeatRate9 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat9)/float64(stat.SeatBets9)*100)
	}
	if stat.SeatBets10 > 0 {
		stat.PlayerWinSeatRate10 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat10)/float64(stat.SeatBets10)*100)
	}
	// 多门率
	if stat.BetRounds > 0 {
		stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerMultis)/float64(stat.BetRounds)*100)
	}
	// // 多门率
	// if stat.BetRounds > 0 {
	// 	stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerMultis)/float64(stat.BetRounds)*100)
	// }

	// 观察局
	if stat.BetRounds > 0 {
		stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
	}
	if stat.BetRounds > 0 {
		stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
	}
	// stat.FMoney = fmt.Sprintf("%.2f", float64(stat.SMoney)/100)
	// stat.FCashOut = fmt.Sprintf("%.2f", float64(stat.SCashOut)/100)
	// stat.FWinMoney = fmt.Sprintf("%.2f", float64(int(stat.SDiamond)+stat.SCashOut-stat.SMoney)/100)
}

// GetABStatDates AB汇总明细
func (s *gameStatsService) GetABStatDates(m, m2, m3 bson.M, startDate, endDate string) (stats []*entity.ABPlayerStat, err error) {
	// 用户列表筛选
	var filterDatePlayerIds []string
	if len(m2) > 0 || len(m3) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				// "_id":        bson.M{"userid":"$userid", "date", "$date"},
				"_id":        "$_id",
				"bets":       bson.M{"$sum": "$bets"},
				"bet_rounds": bson.M{"$sum": "$bet_rounds"},
				"cash":       bson.M{"$sum": "$cash"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				"cash": "$cash",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = ABPlayerStats.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetABStatDates error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			id := r["_id"].(string)
			filterDatePlayerIds = append(filterDatePlayerIds, id)
			ids := strings.Split(id, "-") // id=date-userid
			if len(ids) >= 2 {
				userids = append(userids, ids[1])
			}
		}
		if len(filterDatePlayerIds) == 0 { // 未匹配到
			return
		}
		if len(m3) > 0 { // 流失天数，局均打码量
			if len(userids) == 0 { // 未匹配到
				return
			}
			now := utils.BsonNow()
			m3_2 := []bson.M{
				{"$match": bson.M{"_id": bson.M{"$in": userids}}},
				{"$project": bson.M{
					"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
					"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
				}},
				{"$match": m3},
			}
			var r3_2 []bson.M
			err = PlayerUsers.Pipe(m3_2).All(&r3_2)
			if err != nil {
				beego.Error("GetCrashStatPlayers error32:", err)
			} else if len(r3_2) == 0 { // 未匹配到
				return
			}
			var filterId2 []string
			for _, id := range filterDatePlayerIds {
				for _, r := range r3_2 {
					userid := r["_id"].(string)
					if strings.HasSuffix(id, "-"+userid) {
						filterId2 = append(filterId2, id)
						break
					}
				}
			}
			if len(filterId2) == 0 { // 未匹配到
				return
			}
			filterDatePlayerIds = filterId2
		}
	}
	if len(filterDatePlayerIds) > 0 {
		m["_id"] = bson.M{"$in": filterDatePlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":               "$date_str",
			"date":              bson.M{"$min": "$date"},
			"players":           bson.M{"$sum": 1},
			"new_players":       bson.M{"$sum": bson.M{"$cond": bson.M{"if": "$new_reg", "then": 1, "else": 0}}},
			"old_players":       bson.M{"$sum": bson.M{"$cond": bson.M{"if": "$new_reg", "then": 0, "else": 1}}},
			"all_rounds":        bson.M{"$sum": "$all_rounds"},
			"game_times":        bson.M{"$sum": "$game_times"},
			"bets":              bson.M{"$sum": "$bets"},
			"bet_rounds":        bson.M{"$sum": "$bet_rounds"},
			"win_rounds":        bson.M{"$sum": "$win_rounds"},
			"lose_rounds":       bson.M{"$sum": "$lose_rounds"},
			"tie_rounds":        bson.M{"$sum": "$tie_rounds"},
			"win_bets":          bson.M{"$sum": "$win_bets"},
			"lose_bets":         bson.M{"$sum": "$lose_bets"},
			"tie_bets":          bson.M{"$sum": "$tie_bets"},
			"wins":              bson.M{"$sum": "$wins"},
			"loses":             bson.M{"$sum": "$loses"},
			"cash":              bson.M{"$sum": "$cash"},
			"side_winner1":      bson.M{"$sum": "$side_winner1"},
			"side_winner2":      bson.M{"$sum": "$side_winner2"},
			"side_winner3":      bson.M{"$sum": "$side_winner3"},
			"side_winner4":      bson.M{"$sum": "$side_winner4"},
			"side_winner5":      bson.M{"$sum": "$side_winner5"},
			"side_winner6":      bson.M{"$sum": "$side_winner6"},
			"side_winner7":      bson.M{"$sum": "$side_winner7"},
			"side_winner8":      bson.M{"$sum": "$side_winner8"},
			"side_winner9":      bson.M{"$sum": "$side_winner9"},
			"side_winner10":     bson.M{"$sum": "$side_winner10"},
			"seat_bets1":        bson.M{"$sum": "$seat_bets1"},
			"seat_bets2":        bson.M{"$sum": "$seat_bets2"},
			"seat_bets3":        bson.M{"$sum": "$seat_bets3"},
			"seat_bets4":        bson.M{"$sum": "$seat_bets4"},
			"seat_bets5":        bson.M{"$sum": "$seat_bets5"},
			"seat_bets6":        bson.M{"$sum": "$seat_bets6"},
			"seat_bets7":        bson.M{"$sum": "$seat_bets7"},
			"seat_bets8":        bson.M{"$sum": "$seat_bets8"},
			"seat_bets9":        bson.M{"$sum": "$seat_bets9"},
			"seat_bets10":       bson.M{"$sum": "$seat_bets10"},
			"player_win_seat1":  bson.M{"$sum": "$player_win_seat1"},
			"player_win_seat2":  bson.M{"$sum": "$player_win_seat2"},
			"player_win_seat3":  bson.M{"$sum": "$player_win_seat3"},
			"player_win_seat4":  bson.M{"$sum": "$player_win_seat4"},
			"player_win_seat5":  bson.M{"$sum": "$player_win_seat5"},
			"player_win_seat6":  bson.M{"$sum": "$player_win_seat6"},
			"player_win_seat7":  bson.M{"$sum": "$player_win_seat7"},
			"player_win_seat8":  bson.M{"$sum": "$player_win_seat8"},
			"player_win_seat9":  bson.M{"$sum": "$player_win_seat9"},
			"player_win_seat10": bson.M{"$sum": "$player_win_seat10"},
			"player_multis":     bson.M{"$sum": "$player_multis"},
			"round_bet_avg":     bson.M{"$sum": "$round_bet_avg"},
			"money":             bson.M{"$sum": "$money"},
			"observe_rounds":    bson.M{"$sum": "$observe_rounds"},
			"all_rounds_list":   bson.M{"$push": "$bet_rounds"},
		}},
		{"$sort": bson.M{"date": -1}},
	}
	err = ABPlayerStats.Pipe(m1).All(&stats)
	if err != nil {
		return
	}

	// 查询日活数据
	var dateLoginedUsers = make(map[int64]int)
	m9 := []bson.M{
		{"$match": bson.M{"date": m["date"]}},
		{"$project": bson.M{"date": "$date", "logined_users": "$logined_users"}},
	}
	var r9 []bson.M
	err = PddStats.Pipe(m9).All(&r9)
	if err != nil {
		beego.Error("GetCrashStatDates logined_users error: ", err)
	} else {
		for _, r := range r9 {
			date := r["date"].(int64)
			logined_users := r["logined_users"].(int)
			dateLoginedUsers[date] = logined_users
		}
	}

	mark := "--"
	summary := &entity.ABPlayerStat{
		DateStr: fmt.Sprintf("%s-%s总汇", startDate, endDate),
		Id:      mark,
	}
	ABStatDatesMapping(summary, stats, dateLoginedUsers)
	ABStatDatesSummaryMapping(summary)
	stats = append([]*entity.ABPlayerStat{summary}, stats...)
	return
}

func ABStatDatesMapping(summary *entity.ABPlayerStat, stats []*entity.ABPlayerStat, dateLoginedUsers map[int64]int) {
	for _, stat := range stats {
		dateStr := stat.Id
		date, _ := utils.Unix(dateStr)
		stat.Date = date
		stat.DateStr = dateStr
		loginedUsers := dateLoginedUsers[stat.Date] // 日活

		stat.DateStr = utils.Stamp2Time(stat.Date).Format("2006-01-02")
		stat.FLoginedUsers = loginedUsers

		summary.FLoginedUsers += stat.FLoginedUsers
		summary.FLiveDays += stat.FLiveDays
		summary.FLoseDays += stat.FLoseDays
		summary.AllRounds += stat.AllRounds
		summary.GameTimes += stat.GameTimes
		summary.Bets += stat.Bets
		summary.BetRounds += stat.BetRounds
		summary.WinRounds += stat.WinRounds
		summary.LoseRounds += stat.LoseRounds
		summary.TieRounds += stat.TieRounds
		summary.WinBets += stat.WinBets
		summary.LoseBets += stat.LoseBets
		summary.TieBets += stat.TieBets
		summary.Wins += stat.Wins
		summary.Loses += stat.Loses
		summary.Cash += stat.Cash
		summary.SideWinner1 += stat.SideWinner1
		summary.SideWinner2 += stat.SideWinner2
		summary.SideWinner3 += stat.SideWinner3
		summary.SideWinner4 += stat.SideWinner4
		summary.SideWinner5 += stat.SideWinner5
		summary.SideWinner6 += stat.SideWinner6
		summary.SideWinner7 += stat.SideWinner7
		summary.SideWinner8 += stat.SideWinner8
		summary.SideWinner9 += stat.SideWinner9
		summary.SideWinner10 += stat.SideWinner10
		summary.SeatBets1 += stat.SeatBets1
		summary.SeatBets2 += stat.SeatBets2
		summary.SeatBets3 += stat.SeatBets3
		summary.SeatBets4 += stat.SeatBets4
		summary.SeatBets5 += stat.SeatBets5
		summary.SeatBets6 += stat.SeatBets6
		summary.SeatBets7 += stat.SeatBets7
		summary.SeatBets8 += stat.SeatBets8
		summary.SeatBets9 += stat.SeatBets9
		summary.SeatBets10 += stat.SeatBets10
		summary.PlayerWinSeat1 += stat.PlayerWinSeat1
		summary.PlayerWinSeat2 += stat.PlayerWinSeat2
		summary.PlayerWinSeat3 += stat.PlayerWinSeat3
		summary.PlayerWinSeat4 += stat.PlayerWinSeat4
		summary.PlayerWinSeat5 += stat.PlayerWinSeat5
		summary.PlayerWinSeat6 += stat.PlayerWinSeat6
		summary.PlayerWinSeat7 += stat.PlayerWinSeat7
		summary.PlayerWinSeat8 += stat.PlayerWinSeat8
		summary.PlayerWinSeat9 += stat.PlayerWinSeat9
		summary.PlayerWinSeat10 += stat.PlayerWinSeat10
		summary.PlayerMultis += stat.PlayerMultis
		summary.RoundBetAvg += stat.RoundBetAvg

		summary.ObserveRounds += stat.ObserveRounds
		summary.Players += stat.Players
		summary.NewPlayers += stat.NewPlayers
		summary.OldPlayers += stat.OldPlayers
		summary.AllRoundsList = append(summary.AllRoundsList, stat.AllRoundsList...)

		// 玩 crash 人数占日活比
		if stat.FLoginedUsers > 0 {
			stat.FPlayerRate = fmt.Sprintf("%.2f", float64(stat.Players)/float64(stat.FLoginedUsers))
		}

		// 查询当天注册的新用户
		startTime := utils.Str2Time(fmt.Sprintf("%s 00:00:00", dateStr), Location())
		endTime := utils.Str2Time(fmt.Sprintf("%s 23:59:59", dateStr), Location())
		m7 := []bson.M{
			{"$match": bson.M{
				"ctime":            bson.M{"$gte": startTime, "$lte": endTime},
				"robot":            false,
				"simulation_robot": false,
			}},
			{"$group": bson.M{
				"_id":   nil,
				"total": bson.M{"$sum": 1},
			}},
		}
		var r7 []bson.M
		err := PlayerUsers.Pipe(m7).All(&r7)
		if err != nil {
			beego.Error("查询当天注册的新用户 error: ", err)
		} else if len(r7) > 0 {
			stat.FLoginedNewUsers = r7[0]["total"].(int)
		}
		stat.FLoginedOldUsers = stat.FLoginedUsers - stat.FLoginedNewUsers
		if stat.FLoginedOldUsers < 0 {
			stat.FLoginedOldUsers = 0
		}
		summary.FLoginedNewUsers += stat.FLoginedNewUsers
		summary.FLoginedOldUsers += stat.FLoginedOldUsers
		// FNewPlayerRate, FOldPlayerRate 占新老玩家日活
		if stat.FLoginedNewUsers > 0 {
			stat.FNewPlayerRate = fmt.Sprintf("%.2f", float64(stat.NewPlayers)/float64(stat.FLoginedNewUsers))
		}
		if stat.FLoginedOldUsers > 0 {
			stat.FOldPlayerRate = fmt.Sprintf("%.2f", float64(stat.OldPlayers)/float64(stat.FLoginedOldUsers))
		}

		if stat.Players > 0 {
			stat.FPlayerRoundsAvg = fmt.Sprintf("%.2f", float64(stat.BetRounds)/float64(stat.Players))
		}
		// 局数中位数
		if len(stat.AllRoundsList) > 0 {
			var rounds = stat.AllRoundsList
			sort.Slice(rounds, func(i, j int) bool {
				return rounds[i] < rounds[j]
			})
			length := len(rounds)
			if length > 0 {
				if length%2 == 1 {
					stat.FPlayerRoundsMedian = rounds[length/2]
				} else {
					// 中间两位算平均
					stat.FPlayerRoundsMedian = (rounds[length/2] + rounds[length/2-1]) / 2
				}
			}
		}
		// stat.FGameTimes = fmt.Sprintf("%.2f", float64(stat.GameTimes)/60)
		if stat.BetRounds > 0 {
			stat.FGameTimesAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.BetRounds)/60)
		}
		if stat.Players > 0 {
			stat.FGameTimesPlayerAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.Players)/60)
		}

		stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
		if stat.Players > 0 {
			stat.FPlayerBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.Players)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FRoundBetsAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.Bets, 100)/float64(stat.BetRounds))
			// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
		}
		stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
		stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
		stat.FTieBets = fmt.Sprintf("%.2f", float64(stat.TieBets)/100.0)
		if stat.BetRounds > 0 {
			stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
		}
		if stat.Bets > 0 {
			// stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
			stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
		}
		if stat.WinRounds > 0 {
			stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
		}
		if stat.TieRounds > 0 {
			stat.FTieBetsAvg = fmt.Sprintf("%.2f", float64(stat.TieBets)/float64(stat.TieRounds)/100.0)
		}
		stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
		stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
		stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)

		if stat.WinRounds > 0 {
			stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
		}
		// 下注开奖
		if stat.BetRounds > 0 {
			stat.SideWinnerRate1 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner1)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate2 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner2)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate3 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner3)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate4 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner4)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate5 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner5)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate6 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner6)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate7 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner7)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate8 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner8)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate9 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner9)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate10 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner10)/float64(stat.BetRounds)*100)
		}

		//玩家下注
		if stat.BetRounds > 0 {
			stat.SeatBetsRate1 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets1)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate2 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets2)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate3 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets3)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate4 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets4)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate5 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets5)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate6 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets6)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate7 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets7)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate8 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets8)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate9 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets9)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate10 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets10)/float64(stat.BetRounds)*100)
		}
		// 玩家下注中奖
		if stat.SeatBets1 > 0 {
			stat.PlayerWinSeatRate1 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat1)/float64(stat.SeatBets1)*100)
		}
		if stat.SeatBets2 > 0 {
			stat.PlayerWinSeatRate2 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat2)/float64(stat.SeatBets2)*100)
		}
		if stat.SeatBets3 > 0 {
			stat.PlayerWinSeatRate3 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat3)/float64(stat.SeatBets3)*100)
		}
		if stat.SeatBets4 > 0 {
			stat.PlayerWinSeatRate4 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat4)/float64(stat.SeatBets4)*100)
		}
		if stat.SeatBets5 > 0 {
			stat.PlayerWinSeatRate5 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat5)/float64(stat.SeatBets5)*100)
		}
		if stat.SeatBets6 > 0 {
			stat.PlayerWinSeatRate6 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat6)/float64(stat.SeatBets6)*100)
		}
		if stat.SeatBets7 > 0 {
			stat.PlayerWinSeatRate7 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat7)/float64(stat.SeatBets7)*100)
		}
		if stat.SeatBets8 > 0 {
			stat.PlayerWinSeatRate8 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat8)/float64(stat.SeatBets8)*100)
		}
		if stat.SeatBets9 > 0 {
			stat.PlayerWinSeatRate9 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat9)/float64(stat.SeatBets9)*100)
		}
		if stat.SeatBets10 > 0 {
			stat.PlayerWinSeatRate10 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat10)/float64(stat.SeatBets10)*100)
		}
		// 多门率
		if stat.BetRounds > 0 {
			stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerMultis)/float64(stat.BetRounds)*100)
		}

		// 观察局
		if stat.BetRounds > 0 {
			stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
		}
		if stat.BetRounds > 0 {
			stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
		}

	}
}

func ABStatDatesSummaryMapping(stat *entity.ABPlayerStat) {
	// 玩 lhd 人数占日活比
	if stat.FLoginedUsers > 0 {
		stat.FPlayerRate = fmt.Sprintf("%.2f", float64(stat.Players)/float64(stat.FLoginedUsers))
	}

	// FNewPlayerRate, FOldPlayerRate 占新老玩家日活
	if stat.FLoginedNewUsers > 0 {
		stat.FNewPlayerRate = fmt.Sprintf("%.2f", float64(stat.NewPlayers)/float64(stat.FLoginedNewUsers))
	}
	if stat.FLoginedOldUsers > 0 {
		stat.FOldPlayerRate = fmt.Sprintf("%.2f", float64(stat.OldPlayers)/float64(stat.FLoginedOldUsers))
	}

	if stat.Players > 0 {
		stat.FPlayerRoundsAvg = fmt.Sprintf("%.2f", float64(stat.BetRounds)/float64(stat.Players))
	}
	// 局数中位数
	if len(stat.AllRoundsList) > 0 {
		var rounds = stat.AllRoundsList
		sort.Slice(rounds, func(i, j int) bool {
			return rounds[i] < rounds[j]
		})
		length := len(rounds)
		if length > 0 {
			if length%2 == 1 {
				stat.FPlayerRoundsMedian = rounds[length/2]
			} else {
				// 中间两位算平均
				stat.FPlayerRoundsMedian = (rounds[length/2] + rounds[length/2-1]) / 2
			}
		}
	}
	// stat.FGameTimes = fmt.Sprintf("%.2f", float64(stat.GameTimes)/60)
	if stat.BetRounds > 0 {
		stat.FGameTimesAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.BetRounds)/60)
	}
	if stat.Players > 0 {
		stat.FGameTimesPlayerAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.Players)/60)
	}

	if stat.BetRounds > 0 {
		stat.FGameTimesAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.BetRounds)/60)
	}
	if stat.Players > 0 {
		stat.FGameTimesPlayerAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.Players)/60)
	}

	stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
	if stat.Players > 0 {
		stat.FPlayerBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.Players)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FRoundBetsAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.Bets, 100)/float64(stat.BetRounds))
		// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
	}
	stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
	stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
	stat.FTieBets = fmt.Sprintf("%.2f", float64(stat.TieBets)/100.0)

	if stat.BetRounds > 0 {
		stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
	}
	if stat.Bets > 0 {
		// stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
		stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
	}
	if stat.WinRounds > 0 {
		stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
	}
	stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
	stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
	stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)

	if stat.WinRounds > 0 {
		stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
	}
	if stat.TieRounds > 0 {
		stat.FTieBetsAvg = fmt.Sprintf("%.2f", float64(stat.TieBets)/float64(stat.TieRounds)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
	}
	// 下注开奖
	if stat.BetRounds > 0 {
		stat.SideWinnerRate1 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner1)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate2 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner2)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate3 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner3)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate4 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner4)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate5 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner5)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate6 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner6)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate7 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner7)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate8 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner8)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate9 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner9)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate10 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner10)/float64(stat.BetRounds)*100)
	}

	//玩家下注
	if stat.BetRounds > 0 {
		stat.SeatBetsRate1 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets1)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate2 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets2)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate3 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets3)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate4 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets4)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate5 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets5)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate6 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets6)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate7 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets7)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate8 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets8)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate9 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets9)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate10 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets10)/float64(stat.BetRounds)*100)
	}
	// 玩家下注中奖
	if stat.SeatBets1 > 0 {
		stat.PlayerWinSeatRate1 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat1)/float64(stat.SeatBets1)*100)
	}
	if stat.SeatBets2 > 0 {
		stat.PlayerWinSeatRate2 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat2)/float64(stat.SeatBets2)*100)
	}
	if stat.SeatBets3 > 0 {
		stat.PlayerWinSeatRate3 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat3)/float64(stat.SeatBets3)*100)
	}
	if stat.SeatBets4 > 0 {
		stat.PlayerWinSeatRate4 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat4)/float64(stat.SeatBets4)*100)
	}
	if stat.SeatBets5 > 0 {
		stat.PlayerWinSeatRate5 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat5)/float64(stat.SeatBets5)*100)
	}
	if stat.SeatBets6 > 0 {
		stat.PlayerWinSeatRate6 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat6)/float64(stat.SeatBets6)*100)
	}
	if stat.SeatBets7 > 0 {
		stat.PlayerWinSeatRate7 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat7)/float64(stat.SeatBets7)*100)
	}
	if stat.SeatBets8 > 0 {
		stat.PlayerWinSeatRate8 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat8)/float64(stat.SeatBets8)*100)
	}
	if stat.SeatBets9 > 0 {
		stat.PlayerWinSeatRate9 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat9)/float64(stat.SeatBets9)*100)
	}
	if stat.SeatBets10 > 0 {
		stat.PlayerWinSeatRate10 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat10)/float64(stat.SeatBets10)*100)
	}
	// 多门率
	if stat.BetRounds > 0 {
		stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerMultis)/float64(stat.BetRounds)*100)
	}

	// 观察局
	if stat.BetRounds > 0 {
		stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
	}
	if stat.BetRounds > 0 {
		stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
	}
}

// AB 安能求死
func (s *gameStatsService) GetABAnqsStat(page, pageSize int, m, m2, m3 bson.M, startDate, endDate string) (
	stats []*entity.ABAnqsStat, count int, err error,
) {
	// UpDownXxscStat
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)

	// 用户列表筛选
	var filterPlayerIds []string
	if len(m2) > 0 || len(m3) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				"_id":        "$userid",
				"bets":       bson.M{"$sum": "$bets"},
				"bet_rounds": bson.M{"$sum": "$rounds"},
				"cash":       bson.M{"$sum": "$cash"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				"cash": "$cash",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = ABAnqsStats.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetABAnqsStat error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			userid := r["_id"].(string)
			userids = append(userids, userid)
		}
		if len(userids) == 0 { // 未匹配到
			return
		}
		now := utils.BsonNow()
		m3_2 := []bson.M{
			{"$match": bson.M{"_id": bson.M{"$in": userids}}},
			{"$project": bson.M{
				"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
				"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
			}},
			{"$match": m3},
		}
		var r3_2 []bson.M
		err = PlayerUsers.Pipe(m3_2).All(&r3_2)
		if err != nil {
			beego.Error("lhd PlayerUsers error32:", err)
		} else if len(r3_2) == 0 { // 未匹配到
			return
		}
		for _, r := range r3_2 {
			userid := r["_id"].(string)
			filterPlayerIds = append(filterPlayerIds, userid)
		}
	}

	if len(filterPlayerIds) > 0 {
		m["userid"] = bson.M{"$in": filterPlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":  "$userid",
			"date": bson.M{"$max": "$date"},
			// "date_str":                bson.M{"$sum": "$date_str"},
			"strategy_players":      bson.M{"$max": "$strategy_players"},
			"strategy_times":        bson.M{"$sum": "$strategy_times"},
			"all_in_rounds":         bson.M{"$sum": "$all_in_rounds"},
			"strategy_rounds":       bson.M{"$sum": "$strategy_rounds"},
			"strategy_multi_rounds": bson.M{"$sum": "$strategy_multi_rounds"},
			"strategy_bets":         bson.M{"$sum": "$strategy_bets"},
			"strategy_win_rounds":   bson.M{"$sum": "$strategy_win_rounds"},
			"strategy_wins":         bson.M{"$sum": "$strategy_wins"},
			"strategy_lose_rounds":  bson.M{"$sum": "$strategy_lose_rounds"},
			"strategy_loses":        bson.M{"$sum": "$strategy_loses"},
			"rounds":                bson.M{"$sum": "$rounds"},
			"win_rounds":            bson.M{"$sum": "$win_rounds"},
			"wins":                  bson.M{"$sum": "$wins"},
			"loses":                 bson.M{"$sum": "$loses"},
		}},
		{"$project": bson.M{
			"_id":                   "$_id",
			"date":                  "$date",
			"userid":                "$userid",
			"strategy_players":      "$strategy_players",
			"all_in_rounds":         "$all_in_rounds",
			"strategy_rounds":       "$strategy_rounds",
			"strategy_times":        "$strategy_times",
			"strategy_multi_rounds": "$strategy_multi_rounds",
			"strategy_bets":         "$strategy_bets",
			"strategy_win_rounds":   "$strategy_win_rounds",
			"strategy_wins":         "$strategy_wins",
			"strategy_loses":        "$strategy_loses",
			"strategy_lose_rounds":  "$strategy_lose_rounds",
			"rounds":                "$rounds",
			"win_rounds":            "$win_rounds",
			"wins":                  "$wins",
			"loses":                 "$loses",
		}},
		{"$match": m2},
		{"$count": "count"},
	}
	// 查总条数
	var rCount []bson.M
	err = ABAnqsStats.Pipe(m1).All(&rCount)
	if err != nil {
		return
	} else if len(rCount) > 0 {
		count = rCount[0]["count"].(int)
	}

	// 查列表数据
	m1 = m1[:len(m1)-1]
	m1 = append(m1, []bson.M{
		{"$sort": bson.M{"date": -1}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}...)
	err = ABAnqsStats.Pipe(m1).All(&stats)
	if err != nil {
		return
	}
	ABAnqsStatMapping(stats, m)

	// 查总汇数据
	mark := "--"
	summary := &entity.ABAnqsStat{}
	m1 = m1[:len(m1)-3]
	m1[1]["$group"].(bson.M)["_id"] = nil
	delete(m1[1]["$group"].(bson.M), "date")
	m1[1]["$group"].(bson.M)["userids"] = bson.M{"$addToSet": "$userid"}
	m1[1]["$group"].(bson.M)["strategy_players"] = bson.M{"$max": "$strategy_players"} // 触发策略人数总汇
	m1[2]["$project"].(bson.M)["userids"] = "$userids"
	str, err := json.Marshal(m1)
	fmt.Println(err, string(str))
	var summarys []*entity.ABAnqsStat
	err = ABAnqsStats.Pipe(m1).All(&summarys)
	if err != nil {
		beego.Error("lhd summary error:", err)
		return
	} else if len(summarys) > 0 {
		summary = summarys[0]
		summary.No = fmt.Sprintf("%s-%s总汇", startDate, endDate)
		summary.Id = mark
		summary.FNickname = mark
		summary.FUserType = mark
		summary.FRegistArea = mark
		summary.FCtime = mark
		summary.FLoginTime = mark
		ABAnqsStatSummaryMapping(summary, m)
	}
	if len(summarys) > 0 {
		stats = append([]*entity.ABAnqsStat{summary}, stats...)
	}
	return
}

func ABAnqsStatMapping(stats []*entity.ABAnqsStat, m bson.M) {
	var playerIds []string
	// var playerStats = make(map[string]*entity.CrashPlayerStat, len(stats))
	for _, stat := range stats {
		playerIds = append(playerIds, stat.Id)
		// playerStats[stat.Id] = stat
	}
	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": playerIds}}},
		{"$project": bson.M{
			"_id":          "$_id",
			"nickname":     "$nickname",
			"money":        "$money",
			"cash_out":     "$cash_out",
			"diamond":      "$diamond",
			"ctime":        "$ctime",
			"login_time":   "$login_time",
			"regist_area":  "$regist_area", //ab测试 0:A 1:B 2:C
			"state":        "$state",
			"lhd_strategy": "$lhd_strategy",
		}},
	}
	var r2 []*entity.PlayerUser
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("lhd users error:", err)
		return
	}

	var r2Map = make(map[string]*entity.PlayerUser, len(r2))
	for _, r := range r2 {
		userid := r.Userid
		r2Map[userid] = r
	}
	// 查询总数据
	var ablist []entity.ABPlayerStat
	n := bson.M{}
	// 手动复制 m 到 n
	for k, v := range m {
		n[k] = v
	}
	n["userid"] = bson.M{"$in": playerIds}
	m3 := []bson.M{
		{"$match": n},
		{"$group": bson.M{
			"_id":        "$userid",
			"bet_rounds": bson.M{"$sum": "$bet_rounds"},
			"win_rounds": bson.M{"$sum": "$win_rounds"},
			"bets":       bson.M{"$sum": "$bets"},
			"win_bets":   bson.M{"$sum": "$win_bets"},
			"lose_bets":  bson.M{"$sum": "$lose_bets"},
			"wins":       bson.M{"$sum": "$wins"},
			"loses":      bson.M{"$sum": "$loses"},
		}},
	}
	ABPlayerStats.Pipe(m3).All(&ablist)
	now := utils.BsonNow()
	for i, stat := range stats {
		r, ok := r2Map[stat.Id]
		if !ok {
			continue
		}
		// 玩家昵称	玩家类型 账号类型 注册时间 最后登陆时间 存活天数 流失天数
		nickname := r.Nickname
		money := r.Money
		ctime := r.Ctime
		login_time := r.LoginTime
		regist_area := r.RegistArea
		state := r.State

		// 将日期转成印度时间
		login_time, _ = ConvertToIndiaTime(login_time.Unix())
		ctime, _ = ConvertToIndiaTime(ctime.Unix())

		stat.No = strconv.Itoa(i + 1)
		stat.FNickname = nickname
		stat.FCtime = ctime.Format(utils.FORMAT)
		stat.FLoginTime = login_time.Format(utils.FORMAT)
		stat.FLiveDays = int(math.Ceil(login_time.Sub(ctime).Hours() / 24))
		stat.FLoseDays = int(math.Floor(now.Sub(login_time).Hours() / 24))

		switch regist_area {
		case 0:
			stat.FRegistArea = "A类"
		case 1:
			stat.FRegistArea = "B类"
		case 2:
			stat.FRegistArea = "C类"
		}

		switch state {
		case 1:
			stat.FUserType = "新手"
		case 2:
			if money == 0 {
				stat.FUserType = "零充"
			} else if money >= 20000 && money <= 99999 {
				stat.FUserType = "普充"
			} else if money >= 100000 && money <= 499999 {
				stat.FUserType = "小R"
			} else if money >= 500000 && money <= 999999 {
				stat.FUserType = "中R"
			} else if money >= 1000000 && money <= 9999999 {
				stat.FUserType = "大R"
			} else if money >= 10000000 {
				stat.FUserType = "超大R"
			}
		case 3:
			stat.FUserType = "平民"
		case 4:
			stat.FUserType = "泡沫"
		}
		if stat.StrategyPlayers > 0 {
			stat.TriggerRoundsAvg = fmt.Sprintf("%.2f", float64(stat.StrategyTimes)/float64(stat.StrategyPlayers))
		}
		if stat.AllInRounds > 0 {
			stat.TiggerRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyTimes)/float64(stat.AllInRounds)*100.0)
		}
		if stat.StrategyTimes > 0 {
			stat.ExpiredRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyLoseRounds)/float64(stat.StrategyTimes)*100.0)
		}
		// 多门率
		if stat.StrategyTimes > 0 {
			stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyMultiRounds)/float64(stat.StrategyTimes)*100.0)
		}
		stat.FStrategyBets = fmt.Sprintf("%.2f", float64(stat.StrategyBets)/100.0)
		if stat.StrategyTimes > 0 {
			stat.FStrategyBetRoundsAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.StrategyBets, 100)/float64(stat.StrategyTimes))
		}
		stat.FStrategyWins = fmt.Sprintf("%.2f", float64(stat.StrategyWins)/100.0)
		stat.FStrategyLoses = fmt.Sprintf("%.2f", float64(stat.StrategyLoses)/100.0)
		stat.FStrategyCash = fmt.Sprintf("%.2f", float64(stat.StrategyWins+stat.StrategyLoses)/100.0)
		if len(ablist) > 0 {
			for _, c := range ablist {
				if c.Id == stat.Id {
					// stat.Bets = c.Bets
					stat.WinRounds = int64(c.WinRounds)
					stat.Rounds = int64(c.BetRounds)
					stat.Wins = c.Wins
					stat.Loses = c.Loses
					break
				}
			}
		}
		if stat.Rounds > 0 {
			stat.TotalWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.Rounds)*100.0)
		}
		if stat.StrategyRounds > 0 {
			stat.StrategyWinRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyWinRounds)/float64(stat.StrategyRounds)*100.0)
		}
		if -stat.StrategyLoses > 0 {
			stat.StrategyRewardRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.StrategyWins, 100)/ComputeFloat(-stat.StrategyLoses, 100))*100.0)
		}
		if -stat.Loses > 0 {
			stat.TotalRewardRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.Wins, 100)/ComputeFloat(-stat.Loses, 100))*100.0)
		}
		if r.Money > 0 {
			stat.PlayerRewardRate = fmt.Sprintf("%.2f%%", ComputeFloat(int64(r.CashOut+int32(r.Diamond)), 100)/ComputeFloat(int64(r.Money), 100)*100.0)
			//fmt.Sprintf("%d", int((int64(r.CashOut)+r.Diamond)*10000/int64(r.Money)))
		} else {
			stat.PlayerRewardRate = "未充值"
		}
	}
}

func ABAnqsStatSummaryMapping(stat *entity.ABAnqsStat, m bson.M) {
	// 查询总数据
	abinfo := new(entity.ABPlayerStat)
	n := bson.M{}
	// 手动复制 m 到 n
	for k, v := range m {
		n[k] = v
	}
	n["userid"] = bson.M{"$in": stat.Userids}
	m3 := []bson.M{
		{"$match": n},
		{"$group": bson.M{
			"_id":        nil,
			"bet_rounds": bson.M{"$sum": "$bet_rounds"},
			"win_rounds": bson.M{"$sum": "$win_rounds"},
			"bets":       bson.M{"$sum": "$bets"},
			"win_bets":   bson.M{"$sum": "$win_bets"},
			"lose_bets":  bson.M{"$sum": "$lose_bets"},
			"wins":       bson.M{"$sum": "$wins"},
			"loses":      bson.M{"$sum": "$loses"},
		}},
	}
	ABPlayerStats.Pipe(m3).One(&abinfo)
	stat.StrategyPlayers = int32(len(stat.Userids))
	if stat.StrategyPlayers > 0 {
		stat.TriggerRoundsAvg = fmt.Sprintf("%.2f", float64(stat.StrategyTimes)/float64(stat.StrategyPlayers))
	}
	if stat.AllInRounds > 0 {
		stat.TiggerRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyTimes)/float64(stat.AllInRounds)*100.0)
	}
	if stat.StrategyTimes > 0 {
		stat.ExpiredRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyLoseRounds)/float64(stat.StrategyTimes)*100.0)
	}
	// 多门率
	if stat.StrategyTimes > 0 {
		stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyMultiRounds)/float64(stat.StrategyTimes)*100.0)
	}
	stat.FStrategyBets = fmt.Sprintf("%.2f", float64(stat.StrategyBets)/100.0)
	if stat.StrategyTimes > 0 {
		stat.FStrategyBetRoundsAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.StrategyBets, 100)/float64(stat.StrategyTimes))
	}
	stat.FStrategyWins = fmt.Sprintf("%.2f", float64(stat.StrategyWins)/100.0)
	stat.FStrategyLoses = fmt.Sprintf("%.2f", float64(stat.StrategyLoses)/100.0)
	stat.FStrategyCash = fmt.Sprintf("%.2f", float64(stat.StrategyWins+stat.StrategyLoses)/100.0)
	// 查询总局数
	// stat.Bets = abinfo.Bets
	stat.WinRounds = int64(abinfo.WinRounds)
	stat.Rounds = int64(abinfo.BetRounds)
	stat.Wins = abinfo.Wins
	stat.Loses = abinfo.Loses
	if stat.Rounds > 0 {
		stat.TotalWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.Rounds)*100.0)
	}
	if stat.StrategyRounds > 0 {
		stat.StrategyWinRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyWinRounds)/float64(stat.StrategyRounds)*100.0)
	}
	if -stat.StrategyLoses > 0 {
		stat.StrategyRewardRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.StrategyWins, 100)/ComputeFloat(-stat.StrategyLoses, 100))*100.0)
	}
	if -stat.Loses > 0 {
		stat.TotalRewardRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.Wins, 100)/ComputeFloat(-stat.Loses, 100))*100.0)
	}
}

// AB 安然躺赢
func (s *gameStatsService) GetABArtyStat(page, pageSize int, m, m2, m3 bson.M, startDate, endDate string) (
	stats []*entity.ABArtyStat, count int, err error,
) {
	// UpDownXxscStat
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)

	// 用户列表筛选
	var filterPlayerIds []string
	if len(m2) > 0 || len(m3) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				"_id":        "$userid",
				"bets":       bson.M{"$sum": "$strategy_bets"},
				"bet_rounds": bson.M{"$sum": "$strategy_rounds"},
				"cash":       bson.M{"$sum": "$cash"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				"cash": "$cash",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = ABArtyStats.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetABArtyStat error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			userid := r["_id"].(string)
			userids = append(userids, userid)
		}
		if len(userids) == 0 { // 未匹配到
			return
		}
		now := utils.BsonNow()
		m3_2 := []bson.M{
			{"$match": bson.M{"_id": bson.M{"$in": userids}}},
			{"$project": bson.M{
				"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
				"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
			}},
			{"$match": m3},
		}
		var r3_2 []bson.M
		err = PlayerUsers.Pipe(m3_2).All(&r3_2)
		if err != nil {
			beego.Error("lhd PlayerUsers error32:", err)
		} else if len(r3_2) == 0 { // 未匹配到
			return
		}
		for _, r := range r3_2 {
			userid := r["_id"].(string)
			filterPlayerIds = append(filterPlayerIds, userid)
		}
	}

	if len(filterPlayerIds) > 0 {
		m["userid"] = bson.M{"$in": filterPlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":  "$userid",
			"date": bson.M{"$max": "$date"},
			// "date_str":                bson.M{"$sum": "$date_str"},
			"strategy_players":      bson.M{"$max": "$strategy_players"},
			"strategy_times":        bson.M{"$sum": "$strategy_times"},
			"strategy_rounds":       bson.M{"$sum": "$strategy_rounds"},
			"tigger_times":          bson.M{"$max": "$tigger_times"},
			"all_evo_times":         bson.M{"$sum": "$all_evo_times"},
			"strategy_multi_rounds": bson.M{"$sum": "$strategy_multi_rounds"},
			"strategy_bets":         bson.M{"$sum": "$strategy_bets"},
			"strategy_wins":         bson.M{"$sum": "$strategy_wins"},
			"strategy_win_rounds":   bson.M{"$sum": "$strategy_win_rounds"},
			"strategy_loses":        bson.M{"$sum": "$strategy_loses"},
			"strategy_lose_rounds":  bson.M{"$sum": "$strategy_lose_rounds"},
		}},
		{"$project": bson.M{
			"_id":                   "$_id",
			"date":                  "$date",
			"userid":                "$userid",
			"strategy_players":      "$strategy_players",
			"strategy_times":        "$strategy_times",
			"strategy_rounds":       "$strategy_rounds",
			"tigger_times":          "$tigger_times",
			"all_evo_times":         "$all_evo_times",
			"strategy_multi_rounds": "$strategy_multi_rounds",
			"strategy_bets":         "$strategy_bets",
			"strategy_win_rounds":   "$strategy_win_rounds",
			"strategy_wins":         "$strategy_wins",
			"strategy_loses":        "$strategy_loses",
			"strategy_lose_rounds":  "$strategy_lose_rounds",
		}},
		{"$match": m2},
		{"$count": "count"},
	}
	// 查总条数
	var rCount []bson.M
	err = ABArtyStats.Pipe(m1).All(&rCount)
	if err != nil {
		return
	} else if len(rCount) > 0 {
		count = rCount[0]["count"].(int)
	}

	// 查列表数据
	m1 = m1[:len(m1)-1]
	m1 = append(m1, []bson.M{
		{"$sort": bson.M{"date": -1}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}...)
	err = ABArtyStats.Pipe(m1).All(&stats)
	if err != nil {
		return
	}
	ABArtyStatMapping(stats)

	// 查总汇数据
	mark := "--"
	summary := &entity.ABArtyStat{}
	m1 = m1[:len(m1)-3]
	m1[1]["$group"].(bson.M)["_id"] = nil
	delete(m1[1]["$group"].(bson.M), "date")
	m1[1]["$group"].(bson.M)["userids"] = bson.M{"$addToSet": "$userid"}
	m1[1]["$group"].(bson.M)["strategy_players"] = bson.M{"$sum": "$strategy_players"} // 触发策略人数总汇
	m1[2]["$project"].(bson.M)["userids"] = "$userids"
	str, err := json.Marshal(m1)
	fmt.Println(err, string(str))
	var summarys []*entity.ABArtyStat
	err = ABArtyStats.Pipe(m1).All(&summarys)
	if err != nil {
		beego.Error("lhd summary error:", err)
		return
	} else if len(summarys) > 0 {
		summary = summarys[0]
		summary.No = fmt.Sprintf("%s-%s总汇", startDate, endDate)
		summary.Id = mark
		summary.FNickname = mark
		summary.FUserType = mark
		summary.FRegistArea = mark
		summary.FCtime = mark
		summary.FLoginTime = mark
		ABArtyStatSummaryMapping(summary)
	}
	if len(summarys) > 0 {
		stats = append([]*entity.ABArtyStat{summary}, stats...)
	}
	return
}

func ABArtyStatMapping(stats []*entity.ABArtyStat) {
	var playerIds []string
	// var playerStats = make(map[string]*entity.CrashPlayerStat, len(stats))
	for _, stat := range stats {
		playerIds = append(playerIds, stat.Id)
		// playerStats[stat.Id] = stat
	}
	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": playerIds}}},
		{"$project": bson.M{
			"_id":          "$_id",
			"nickname":     "$nickname",
			"money":        "$money",
			"cash_out":     "$cash_out",
			"diamond":      "$diamond",
			"ctime":        "$ctime",
			"login_time":   "$login_time",
			"regist_area":  "$regist_area", //ab测试 0:A 1:B 2:C
			"state":        "$state",
			"lhd_strategy": "$lhd_strategy",
		}},
	}
	var r2 []*entity.PlayerUser
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("lhd users error:", err)
		return
	}

	var r2Map = make(map[string]*entity.PlayerUser, len(r2))
	for _, r := range r2 {
		userid := r.Userid
		r2Map[userid] = r
	}
	now := utils.BsonNow()
	for i, stat := range stats {
		r, ok := r2Map[stat.Id]
		if !ok {
			continue
		}
		// 玩家昵称	玩家类型 账号类型 注册时间 最后登陆时间 存活天数 流失天数
		nickname := r.Nickname
		money := r.Money
		ctime := r.Ctime
		login_time := r.LoginTime
		regist_area := r.RegistArea
		state := r.State

		// 将日期转成印度时间
		login_time, _ = ConvertToIndiaTime(login_time.Unix())
		ctime, _ = ConvertToIndiaTime(ctime.Unix())

		stat.No = strconv.Itoa(i + 1)
		stat.FNickname = nickname
		stat.FCtime = ctime.Format(utils.FORMAT)
		stat.FLoginTime = login_time.Format(utils.FORMAT)
		stat.FLiveDays = int(math.Ceil(login_time.Sub(ctime).Hours() / 24))
		stat.FLoseDays = int(math.Floor(now.Sub(login_time).Hours() / 24))

		switch regist_area {
		case 0:
			stat.FRegistArea = "A类"
		case 1:
			stat.FRegistArea = "B类"
		case 2:
			stat.FRegistArea = "C类"
		}

		switch state {
		case 1:
			stat.FUserType = "新手"
		case 2:
			if money == 0 {
				stat.FUserType = "零充"
			} else if money >= 20000 && money <= 99999 {
				stat.FUserType = "普充"
			} else if money >= 100000 && money <= 499999 {
				stat.FUserType = "小R"
			} else if money >= 500000 && money <= 999999 {
				stat.FUserType = "中R"
			} else if money >= 1000000 && money <= 9999999 {
				stat.FUserType = "大R"
			} else if money >= 10000000 {
				stat.FUserType = "超大R"
			}
		case 3:
			stat.FUserType = "平民"
		case 4:
			stat.FUserType = "泡沫"
		}
		if stat.StrategyTimes > 0 {
			stat.TriggerRate = fmt.Sprintf("%.2f%%", float64(stat.TiggerTimes)/float64(stat.StrategyTimes)*100.0)

		}
		if stat.StrategyPlayers > 0 {
			stat.PerCapitaAvg = fmt.Sprintf("%.2f", float64(stat.TiggerTimes)/float64(stat.StrategyPlayers))
		}
		if stat.StrategyTimes > 0 {
			stat.TiggerStrategyAvg = fmt.Sprintf("%.2f", float64(stat.AllEvoTimes)/float64(stat.StrategyTimes))
		}
		if stat.TiggerTimes > 0 {
			stat.YxTiggerStrategyAvg = fmt.Sprintf("%.2f", float64(stat.AllEvoTimes)/float64(stat.TiggerTimes))
		}
		// 多门率
		if stat.AllEvoTimes > 0 {
			stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyMultiRounds)/float64(stat.AllEvoTimes)*100.0)
		}
		stat.FStrategyBets = fmt.Sprintf("%.2f", float64(stat.StrategyBets)/100.0)
		if stat.StrategyTimes > 0 {
			stat.FStrategyBetRoundsAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.StrategyBets, 100)/float64(stat.AllEvoTimes))
		}
		stat.FStrategyWins = fmt.Sprintf("%.2f", float64(stat.StrategyWins)/100.0)
		stat.FStrategyLoses = fmt.Sprintf("%.2f", float64(stat.StrategyLoses)/100.0)
		stat.FStrategyCash = fmt.Sprintf("%.2f", float64(stat.StrategyWins+stat.StrategyLoses)/100.0)

		if stat.AllEvoTimes > 0 {
			stat.FStrategyWinRate = fmt.Sprintf("%.2f%%", float64(stat.AllEvoTimes)/float64(stat.AllEvoTimes)*100.0)
		}
		if -stat.StrategyLoses > 0 {
			stat.FStrategyRebateRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.StrategyWins, 100)/ComputeFloat(-stat.StrategyLoses, 100))*100.0)
		} else {
			stat.FStrategyRebateRate = "未输过"
		}
	}
}

func ABArtyStatSummaryMapping(stat *entity.ABArtyStat) {
	stat.StrategyPlayers = int32(len(stat.Userids))
	if stat.StrategyTimes > 0 {
		stat.TriggerRate = fmt.Sprintf("%.2f%%", float64(stat.TiggerTimes)/float64(stat.StrategyTimes)*100.0)

	}
	if stat.StrategyPlayers > 0 {
		stat.PerCapitaAvg = fmt.Sprintf("%.2f", float64(stat.TiggerTimes)/float64(stat.StrategyPlayers))
	}
	if stat.StrategyTimes > 0 {
		stat.TiggerStrategyAvg = fmt.Sprintf("%.2f", float64(stat.AllEvoTimes)/float64(stat.StrategyTimes))

	}
	if stat.TiggerTimes > 0 {
		stat.YxTiggerStrategyAvg = fmt.Sprintf("%.2f", float64(stat.AllEvoTimes)/float64(stat.TiggerTimes))
	}
	// 多门率
	if stat.AllEvoTimes > 0 {
		stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyMultiRounds)/float64(stat.AllEvoTimes)*100.0)
	}
	stat.FStrategyBets = fmt.Sprintf("%.2f", float64(stat.StrategyBets)/100.0)
	if stat.StrategyTimes > 0 {
		stat.FStrategyBetRoundsAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.StrategyBets, 100)/float64(stat.AllEvoTimes))
	}
	stat.FStrategyWins = fmt.Sprintf("%.2f", float64(stat.StrategyWins)/100.0)
	stat.FStrategyLoses = fmt.Sprintf("%.2f", float64(stat.StrategyLoses)/100.0)
	stat.FStrategyCash = fmt.Sprintf("%.2f", float64(stat.StrategyWins+stat.StrategyLoses)/100.0)

	if stat.AllEvoTimes > 0 {
		stat.FStrategyWinRate = fmt.Sprintf("%.2f%%", float64(stat.AllEvoTimes)/float64(stat.AllEvoTimes)*100.0)
	}
	if -stat.StrategyLoses > 0 {
		stat.FStrategyRebateRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.StrategyWins, 100)/ComputeFloat(-stat.StrategyLoses, 100))*100.0)
	} else {
		stat.FStrategyRebateRate = "未输过"
	}
}

/*
CP游戏策略
*/
func (s *gameStatsService) GetCPStatPlayers(page, pageSize int, m, m2, m3 bson.M, startDate, endDate string) (
	stats []*entity.CPPlayerStat, count int, err error,
) {
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)

	// 用户列表筛选
	var filterPlayerIds []string
	if len(m2) > 0 || len(m3) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				"_id":        "$userid",
				"bets":       bson.M{"$sum": "$bets"},
				"bet_rounds": bson.M{"$sum": "$bet_rounds"},
				"cash":       bson.M{"$sum": "$cash"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				"cash": "$cash",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = CPPlayerStats.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetCPStatPlayers error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			userid := r["_id"].(string)
			userids = append(userids, userid)
		}
		if len(userids) == 0 { // 未匹配到
			return
		}
		now := utils.BsonNow()
		m3_2 := []bson.M{
			{"$match": bson.M{"_id": bson.M{"$in": userids}}},
			{"$project": bson.M{
				"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
				"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
			}},
			{"$match": m3},
		}
		var r3_2 []bson.M
		err = PlayerUsers.Pipe(m3_2).All(&r3_2)
		if err != nil {
			beego.Error("GetABStatPlayers error32:", err)
		} else if len(r3_2) == 0 { // 未匹配到
			return
		}
		for _, r := range r3_2 {
			userid := r["_id"].(string)
			filterPlayerIds = append(filterPlayerIds, userid)
		}
	}
	if len(filterPlayerIds) > 0 {
		m["userid"] = bson.M{"$in": filterPlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":              "$userid",
			"date":             bson.M{"$max": "$date"},
			"all_rounds":       bson.M{"$sum": "$all_rounds"},
			"game_times":       bson.M{"$sum": "$game_times"},
			"bets":             bson.M{"$sum": "$bets"},
			"bet_rounds":       bson.M{"$sum": "$bet_rounds"},
			"win_rounds":       bson.M{"$sum": "$win_rounds"},
			"lose_rounds":      bson.M{"$sum": "$lose_rounds"},
			"tie_rounds":       bson.M{"$sum": "$tie_rounds"},
			"win_bets":         bson.M{"$sum": "$win_bets"},
			"lose_bets":        bson.M{"$sum": "$lose_bets"},
			"tie_bets":         bson.M{"$sum": "$tie_bets"},
			"wins":             bson.M{"$sum": "$wins"},
			"loses":            bson.M{"$sum": "$loses"},
			"cash":             bson.M{"$sum": "$cash"},
			"side_winner1":     bson.M{"$sum": "$side_winner1"},
			"side_winner2":     bson.M{"$sum": "$side_winner2"},
			"side_winner3":     bson.M{"$sum": "$side_winner3"},
			"side_winner4":     bson.M{"$sum": "$side_winner4"},
			"side_winner5":     bson.M{"$sum": "$side_winner5"},
			"side_winner6":     bson.M{"$sum": "$side_winner6"},
			"seat_bets1":       bson.M{"$sum": "$seat_bets1"},
			"seat_bets2":       bson.M{"$sum": "$seat_bets2"},
			"seat_bets3":       bson.M{"$sum": "$seat_bets3"},
			"seat_bets4":       bson.M{"$sum": "$seat_bets4"},
			"seat_bets5":       bson.M{"$sum": "$seat_bets5"},
			"seat_bets6":       bson.M{"$sum": "$seat_bets6"},
			"player_win_seat1": bson.M{"$sum": "$player_win_seat1"},
			"player_win_seat2": bson.M{"$sum": "$player_win_seat2"},
			"player_win_seat3": bson.M{"$sum": "$player_win_seat3"},
			"player_win_seat4": bson.M{"$sum": "$player_win_seat4"},
			"player_win_seat5": bson.M{"$sum": "$player_win_seat5"},
			"player_win_seat6": bson.M{"$sum": "$player_win_seat6"},
			"player_multis":    bson.M{"$sum": "$player_multis"},
			"round_bet_avg":    bson.M{"$sum": "$round_bet_avg"},
			"money":            bson.M{"$sum": "$money"},
			"observe_rounds":   bson.M{"$sum": "$observe_rounds"},
		}},
		{"$project": bson.M{
			"_id":              "$_id",
			"all_rounds":       "$all_rounds",
			"game_times":       "$game_times",
			"bets":             "$bets",
			"bet_rounds":       "$bet_rounds",
			"win_rounds":       "$win_rounds",
			"lose_rounds":      "$lose_rounds",
			"tie_rounds":       "$tie_rounds",
			"win_bets":         "$win_bets",
			"lose_bets":        "$lose_bets",
			"tie_bets":         "$tie_bets",
			"wins":             "$wins",
			"loses":            "$loses",
			"cash":             "$cash",
			"side_winner1":     "$side_winner1",
			"side_winner2":     "$side_winner2",
			"side_winner3":     "$side_winner3",
			"side_winner4":     "$side_winner4",
			"side_winner5":     "$side_winner5",
			"side_winner6":     "$side_winner6",
			"seat_bets1":       "$seat_bets1",
			"seat_bets2":       "$seat_bets2",
			"seat_bets3":       "$seat_bets3",
			"seat_bets4":       "$seat_bets4",
			"seat_bets5":       "$seat_bets5",
			"seat_bets6":       "$seat_bets6",
			"player_win_seat1": "$player_win_seat1",
			"player_win_seat2": "$player_win_seat2",
			"player_win_seat3": "$player_win_seat3",
			"player_win_seat4": "$player_win_seat4",
			"player_win_seat5": "$player_win_seat5",
			"player_win_seat6": "$player_win_seat6",

			"player_multis":  "$player_multis",
			"money":          "$money",
			"observe_rounds": "$observe_rounds",
			// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}}, // "can't $divide by zero"
			"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
		}},
		{"$match": m2},
		{"$count": "count"},
	}
	// 查总条数
	var rCount []bson.M
	err = CPPlayerStats.Pipe(m1).All(&rCount)
	if err != nil {
		return
	} else if len(rCount) > 0 {
		count = rCount[0]["count"].(int)
	}

	// 查列表数据
	m1 = m1[:len(m1)-1]
	m1 = append(m1, []bson.M{
		{"$sort": bson.M{"date": -1}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}...)
	err = CPPlayerStats.Pipe(m1).All(&stats)
	if err != nil {
		return
	}
	CPStatPlayersMapping(stats)

	// 查总汇数据
	mark := "--"
	summary := &entity.CPPlayerStat{}
	m1 = m1[:len(m1)-3]
	m1[1]["$group"].(bson.M)["_id"] = nil
	delete(m1[1]["$group"].(bson.M), "date")
	m1[1]["$group"].(bson.M)["userids"] = bson.M{"$addToSet": "$userid"}
	m1[2]["$project"].(bson.M)["userids"] = "$userids"
	var summarys []*entity.CPPlayerStat
	err = CPPlayerStats.Pipe(m1).All(&summarys)
	if err != nil {
		return
	} else if len(summarys) > 0 {
		summary = summarys[0]
		summary.No = fmt.Sprintf("%s-%s总汇", startDate, endDate)
		summary.Id = mark
		summary.FNickname = mark
		summary.FUserType = mark
		summary.FRegistArea = mark
		summary.FCtime = mark
		summary.FLoginTime = mark
		// summary.FLiveDays = mark
		// summary.FLoseDays = mark
		CPStatPlayersSummaryMapping(summary)
	}
	if len(summarys) > 0 {
		stats = append([]*entity.CPPlayerStat{summary}, stats...)
	}
	return
}

func CPStatPlayersMapping(stats []*entity.CPPlayerStat) {
	var playerIds []string
	// var playerStats = make(map[string]*entity.CrashPlayerStat, len(stats))
	for _, stat := range stats {
		playerIds = append(playerIds, stat.Id)
		// playerStats[stat.Id] = stat
	}
	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": playerIds}}},
		{"$project": bson.M{
			"_id":         "$_id",
			"nickname":    "$nickname",
			"money":       "$money",
			"cash_out":    "$cash_out",
			"diamond":     "$diamond",
			"ctime":       "$ctime",
			"login_time":  "$login_time",
			"regist_area": "$regist_area", //ab测试 0:A 1:B 2:C
			"state":       "$state",
		}},
	}
	var r2 []bson.M
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("cp users error:", err)
		return
	}
	var r2Map = make(map[string]bson.M, len(r2))
	for _, r := range r2 {
		userid := r["_id"].(string)
		r2Map[userid] = r
	}
	now := utils.BsonNow()
	for i, stat := range stats {
		r, ok := r2Map[stat.Id]
		if !ok {
			continue
		}
		// 玩家昵称	玩家类型 账号类型 注册时间 最后登陆时间 存活天数 流失天数
		// userid := r["_id"].(string)
		nickname := r["nickname"].(string)
		money := r["money"].(int)
		cash_out := r["cash_out"].(int)
		diamond := r["diamond"].(int64)
		ctime := r["ctime"].(time.Time)
		login_time := r["login_time"].(time.Time)
		regist_area := r["regist_area"].(int)
		state := r["state"].(int)

		// 将日期转成印度时间
		login_time, _ = ConvertToIndiaTime(login_time.Unix())
		ctime, _ = ConvertToIndiaTime(ctime.Unix())

		stat.No = strconv.Itoa(i + 1)
		stat.FNickname = nickname
		stat.Money = uint32(money)
		stat.FCtime = ctime.Format(utils.FORMAT)
		stat.FLoginTime = login_time.Format(utils.FORMAT)
		stat.FLiveDays = int(math.Ceil(login_time.Sub(ctime).Hours() / 24))
		stat.FLoseDays = int(math.Floor(now.Sub(login_time).Hours() / 24))

		// summary.FLiveDays += stat.FLiveDays
		// summary.FLoseDays += stat.FLoseDays
		// summary.AllRounds += stat.AllRounds
		// summary.GameTimes += stat.GameTimes
		// summary.Bets += stat.Bets
		// summary.BetRounds += stat.BetRounds
		// summary.WinRounds += stat.WinRounds
		// summary.LoseRounds += stat.LoseRounds
		// summary.TieRounds += stat.TieRounds
		// summary.WinBets += stat.WinBets
		// summary.LoseBets += stat.LoseBets
		// summary.TieBets += stat.TieBets
		// summary.Wins += stat.Wins
		// summary.Loses += stat.Loses
		// summary.Cash += stat.Cash
		// summary.MulpitleSum += stat.MulpitleSum
		// summary.WinEscapeMulpitleSum += stat.WinEscapeMulpitleSum
		// summary.WinMulpitleSum += stat.WinMulpitleSum
		// summary.LoseMulpitleSum += stat.LoseMulpitleSum
		// summary.Mulpitles = append(summary.Mulpitles, stat.Mulpitles...)
		// summary.WinEscapeMulpitles = append(summary.WinEscapeMulpitles, stat.WinEscapeMulpitles...)
		// summary.WinMulpitles = append(summary.WinMulpitles, stat.WinMulpitles...)
		// summary.LoseMulpitles = append(summary.LoseMulpitles, stat.LoseMulpitles...)
		// summary.ObserveRounds += stat.ObserveRounds
		// summary.SMoney += money
		// summary.SCashOut += cash_out
		// summary.SDiamond += diamond

		switch regist_area {
		case 0:
			stat.FRegistArea = "A类"
		case 1:
			stat.FRegistArea = "B类"
		case 2:
			stat.FRegistArea = "C类"
		}

		switch state {
		case 1:
			stat.FUserType = "新手"
		case 2:
			if money == 0 {
				stat.FUserType = "零充"
			} else if money >= 20000 && money <= 99999 {
				stat.FUserType = "普充"
			} else if money >= 100000 && money <= 499999 {
				stat.FUserType = "小R"
			} else if money >= 500000 && money <= 999999 {
				stat.FUserType = "中R"
			} else if money >= 1000000 && money <= 9999999 {
				stat.FUserType = "大R"
			} else if money >= 10000000 {
				stat.FUserType = "超大R"
			}
		case 3:
			stat.FUserType = "平民"
		case 4:
			stat.FUserType = "泡沫"
		}

		stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
		if stat.BetRounds > 0 {
			// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.BetRounds)/100.0)
			stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
		}
		stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
		stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
		stat.FTieBets = fmt.Sprintf("%.2f", float64(stat.TieBets)/100.0)
		if stat.BetRounds > 0 {
			stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
		}
		if stat.Bets > 0 {
			// stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
			stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
		}
		if stat.WinRounds > 0 {
			stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
		}
		if stat.TieRounds > 0 {
			stat.FTieBetsAvg = fmt.Sprintf("%.2f", float64(stat.TieBets)/float64(stat.TieRounds)/100.0)
		}
		stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
		stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
		stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)

		if stat.WinRounds > 0 {
			stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
		}
		// 下注开奖
		if stat.BetRounds > 0 {
			stat.SideWinnerRate1 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner1)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate2 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner2)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate3 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner3)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate4 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner4)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate5 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner5)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate6 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner6)/float64(stat.BetRounds)*100)
		}

		//玩家下注
		if stat.BetRounds > 0 {
			stat.SeatBetsRate1 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets1)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate2 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets2)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate3 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets3)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate4 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets4)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate5 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets5)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate6 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets6)/float64(stat.BetRounds)*100)
		}
		// 玩家下注中奖
		if stat.SeatBets1 > 0 {
			stat.PlayerWinSeatRate1 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat1)/float64(stat.SeatBets1)*100)
		}
		if stat.SeatBets2 > 0 {
			stat.PlayerWinSeatRate2 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat2)/float64(stat.SeatBets2)*100)
		}
		if stat.SeatBets3 > 0 {
			stat.PlayerWinSeatRate3 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat3)/float64(stat.SeatBets3)*100)
		}
		if stat.SeatBets4 > 0 {
			stat.PlayerWinSeatRate4 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat4)/float64(stat.SeatBets4)*100)
		}
		if stat.SeatBets5 > 0 {
			stat.PlayerWinSeatRate5 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat5)/float64(stat.SeatBets5)*100)
		}
		if stat.SeatBets6 > 0 {
			stat.PlayerWinSeatRate6 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat6)/float64(stat.SeatBets6)*100)
		}
		// 多门率
		if stat.BetRounds > 0 {
			stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerMultis)/float64(stat.BetRounds)*100)
		}

		// 观察局
		if stat.BetRounds > 0 {
			stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
		}
		if stat.BetRounds > 0 {
			stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
		}
		stat.FMoney = fmt.Sprintf("%.2f", float64(money)/100)
		stat.FCashOut = fmt.Sprintf("%.2f", float64(cash_out)/100)
		stat.FWinMoney = fmt.Sprintf("%.2f", float64(int(diamond)+cash_out-money)/100)
	}
}

func CPStatPlayersSummaryMapping(stat *entity.CPPlayerStat) {
	// 查用户基本信息
	// m2 := []bson.M{
	// 	{"$match": bson.M{"_id": bson.M{"$in": stat.Userids}}},
	// 	{"$group": bson.M{
	// 		"_id":      nil,
	// 		"money":    bson.M{"$sum": "$money"},
	// 		"cash_out": bson.M{"$sum": "$cash_out"},
	// 		"diamond":  bson.M{"$sum": "$diamond"},
	// 	}},
	// }
	// var r2 []bson.M
	// err := PlayerUsers.Pipe(m2).All(&r2)
	// if err != nil {
	// 	beego.Error("lhd users error:", err)
	// 	return
	// } else if len(r2) > 0 {
	// 	money := r2[0]["money"].(int)
	// 	cash_out := r2[0]["cash_out"].(int)
	// 	diamond := r2[0]["diamond"].(int64)
	// 	stat.SMoney = money
	// 	stat.SCashOut = cash_out
	// 	stat.SDiamond = diamond
	// }

	stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
	if stat.BetRounds > 0 {
		stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.BetRounds)/100.0)
		// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
	}

	stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
	stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
	stat.FTieBets = fmt.Sprintf("%.2f", float64(stat.TieBets)/100.0)
	if stat.BetRounds > 0 {
		stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
	}
	if stat.Bets > 0 {
		stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
	}
	if stat.WinRounds > 0 {
		stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
	}
	if stat.TieRounds > 0 {
		stat.FTieBetsAvg = fmt.Sprintf("%.2f", float64(stat.TieBets)/float64(stat.TieRounds)/100.0)
	}
	stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
	stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
	stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)
	if stat.WinRounds > 0 {
		stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
	}
	// 下注开奖
	if stat.BetRounds > 0 {
		stat.SideWinnerRate1 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner1)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate2 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner2)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate3 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner3)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate4 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner4)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate5 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner5)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate6 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner6)/float64(stat.BetRounds)*100)
	}

	//玩家下注
	if stat.BetRounds > 0 {
		stat.SeatBetsRate1 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets1)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate2 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets2)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate3 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets3)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate4 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets4)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate5 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets5)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate6 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets6)/float64(stat.BetRounds)*100)
	}
	// 玩家下注中奖
	if stat.SeatBets1 > 0 {
		stat.PlayerWinSeatRate1 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat1)/float64(stat.SeatBets1)*100)
	}
	if stat.SeatBets2 > 0 {
		stat.PlayerWinSeatRate2 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat2)/float64(stat.SeatBets2)*100)
	}
	if stat.SeatBets3 > 0 {
		stat.PlayerWinSeatRate3 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat3)/float64(stat.SeatBets3)*100)
	}
	if stat.SeatBets4 > 0 {
		stat.PlayerWinSeatRate4 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat4)/float64(stat.SeatBets4)*100)
	}
	if stat.SeatBets5 > 0 {
		stat.PlayerWinSeatRate5 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat5)/float64(stat.SeatBets5)*100)
	}
	if stat.SeatBets6 > 0 {
		stat.PlayerWinSeatRate6 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat6)/float64(stat.SeatBets6)*100)
	}
	// 多门率
	if stat.BetRounds > 0 {
		stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerMultis)/float64(stat.BetRounds)*100)
	}

	// 观察局
	if stat.BetRounds > 0 {
		stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
	}
	if stat.BetRounds > 0 {
		stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
	}
	// stat.FMoney = fmt.Sprintf("%.2f", float64(stat.SMoney)/100)
	// stat.FCashOut = fmt.Sprintf("%.2f", float64(stat.SCashOut)/100)
	// stat.FWinMoney = fmt.Sprintf("%.2f", float64(int(stat.SDiamond)+stat.SCashOut-stat.SMoney)/100)
}

// GetABStatDates AB汇总明细
func (s *gameStatsService) GetCPStatDates(m, m2, m3 bson.M, startDate, endDate string) (stats []*entity.CPPlayerStat, err error) {
	// 用户列表筛选
	var filterDatePlayerIds []string
	if len(m2) > 0 || len(m3) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				// "_id":        bson.M{"userid":"$userid", "date", "$date"},
				"_id":        "$_id",
				"bets":       bson.M{"$sum": "$bets"},
				"bet_rounds": bson.M{"$sum": "$bet_rounds"},
				"cash":       bson.M{"$sum": "$cash"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				"cash": "$cash",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = CPPlayerStats.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetCPStatDates error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			id := r["_id"].(string)
			filterDatePlayerIds = append(filterDatePlayerIds, id)
			ids := strings.Split(id, "-") // id=date-userid
			if len(ids) >= 2 {
				userids = append(userids, ids[1])
			}
		}
		if len(filterDatePlayerIds) == 0 { // 未匹配到
			return
		}
		if len(m3) > 0 { // 流失天数，局均打码量
			if len(userids) == 0 { // 未匹配到
				return
			}
			now := utils.BsonNow()
			m3_2 := []bson.M{
				{"$match": bson.M{"_id": bson.M{"$in": userids}}},
				{"$project": bson.M{
					"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
					"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
				}},
				{"$match": m3},
			}
			var r3_2 []bson.M
			err = PlayerUsers.Pipe(m3_2).All(&r3_2)
			if err != nil {
				beego.Error("GetCPStatDates error32:", err)
			} else if len(r3_2) == 0 { // 未匹配到
				return
			}
			var filterId2 []string
			for _, id := range filterDatePlayerIds {
				for _, r := range r3_2 {
					userid := r["_id"].(string)
					if strings.HasSuffix(id, "-"+userid) {
						filterId2 = append(filterId2, id)
						break
					}
				}
			}
			if len(filterId2) == 0 { // 未匹配到
				return
			}
			filterDatePlayerIds = filterId2
		}
	}
	if len(filterDatePlayerIds) > 0 {
		m["_id"] = bson.M{"$in": filterDatePlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":              "$date_str",
			"date":             bson.M{"$min": "$date"},
			"players":          bson.M{"$sum": 1},
			"new_players":      bson.M{"$sum": bson.M{"$cond": bson.M{"if": "$new_reg", "then": 1, "else": 0}}},
			"old_players":      bson.M{"$sum": bson.M{"$cond": bson.M{"if": "$new_reg", "then": 0, "else": 1}}},
			"all_rounds":       bson.M{"$sum": "$all_rounds"},
			"game_times":       bson.M{"$sum": "$game_times"},
			"bets":             bson.M{"$sum": "$bets"},
			"bet_rounds":       bson.M{"$sum": "$bet_rounds"},
			"win_rounds":       bson.M{"$sum": "$win_rounds"},
			"lose_rounds":      bson.M{"$sum": "$lose_rounds"},
			"tie_rounds":       bson.M{"$sum": "$tie_rounds"},
			"win_bets":         bson.M{"$sum": "$win_bets"},
			"lose_bets":        bson.M{"$sum": "$lose_bets"},
			"tie_bets":         bson.M{"$sum": "$tie_bets"},
			"wins":             bson.M{"$sum": "$wins"},
			"loses":            bson.M{"$sum": "$loses"},
			"cash":             bson.M{"$sum": "$cash"},
			"side_winner1":     bson.M{"$sum": "$side_winner1"},
			"side_winner2":     bson.M{"$sum": "$side_winner2"},
			"side_winner3":     bson.M{"$sum": "$side_winner3"},
			"side_winner4":     bson.M{"$sum": "$side_winner4"},
			"side_winner5":     bson.M{"$sum": "$side_winner5"},
			"side_winner6":     bson.M{"$sum": "$side_winner6"},
			"seat_bets1":       bson.M{"$sum": "$seat_bets1"},
			"seat_bets2":       bson.M{"$sum": "$seat_bets2"},
			"seat_bets3":       bson.M{"$sum": "$seat_bets3"},
			"seat_bets4":       bson.M{"$sum": "$seat_bets4"},
			"seat_bets5":       bson.M{"$sum": "$seat_bets5"},
			"seat_bets6":       bson.M{"$sum": "$seat_bets6"},
			"player_win_seat1": bson.M{"$sum": "$player_win_seat1"},
			"player_win_seat2": bson.M{"$sum": "$player_win_seat2"},
			"player_win_seat3": bson.M{"$sum": "$player_win_seat3"},
			"player_win_seat4": bson.M{"$sum": "$player_win_seat4"},
			"player_win_seat5": bson.M{"$sum": "$player_win_seat5"},
			"player_win_seat6": bson.M{"$sum": "$player_win_seat6"},
			"player_multis":    bson.M{"$sum": "$player_multis"},
			"round_bet_avg":    bson.M{"$sum": "$round_bet_avg"},
			"money":            bson.M{"$sum": "$money"},
			"observe_rounds":   bson.M{"$sum": "$observe_rounds"},
			"all_rounds_list":  bson.M{"$push": "$bet_rounds"},
		}},
		{"$sort": bson.M{"date": -1}},
	}
	err = CPPlayerStats.Pipe(m1).All(&stats)
	if err != nil {
		return
	}

	// 查询日活数据
	var dateLoginedUsers = make(map[int64]int)
	m9 := []bson.M{
		{"$match": bson.M{"date": m["date"]}},
		{"$project": bson.M{"date": "$date", "logined_users": "$logined_users"}},
	}
	var r9 []bson.M
	err = PddStats.Pipe(m9).All(&r9)
	if err != nil {
		beego.Error("GetCPStatDates logined_users error: ", err)
	} else {
		for _, r := range r9 {
			date := r["date"].(int64)
			logined_users := r["logined_users"].(int)
			dateLoginedUsers[date] = logined_users
		}
	}

	mark := "--"
	summary := &entity.CPPlayerStat{
		DateStr: fmt.Sprintf("%s-%s总汇", startDate, endDate),
		Id:      mark,
	}
	CPStatDatesMapping(summary, stats, dateLoginedUsers)
	CPStatDatesSummaryMapping(summary)
	stats = append([]*entity.CPPlayerStat{summary}, stats...)
	return
}

func CPStatDatesMapping(summary *entity.CPPlayerStat, stats []*entity.CPPlayerStat, dateLoginedUsers map[int64]int) {
	for _, stat := range stats {
		dateStr := stat.Id
		date, _ := utils.Unix(dateStr)
		stat.Date = date
		stat.DateStr = dateStr
		loginedUsers := dateLoginedUsers[stat.Date] // 日活

		stat.DateStr = utils.Stamp2Time(stat.Date).Format("2006-01-02")
		stat.FLoginedUsers = loginedUsers

		summary.FLoginedUsers += stat.FLoginedUsers
		summary.FLiveDays += stat.FLiveDays
		summary.FLoseDays += stat.FLoseDays
		summary.AllRounds += stat.AllRounds
		summary.GameTimes += stat.GameTimes
		summary.Bets += stat.Bets
		summary.BetRounds += stat.BetRounds
		summary.WinRounds += stat.WinRounds
		summary.LoseRounds += stat.LoseRounds
		summary.TieRounds += stat.TieRounds
		summary.WinBets += stat.WinBets
		summary.LoseBets += stat.LoseBets
		summary.TieBets += stat.TieBets
		summary.Wins += stat.Wins
		summary.Loses += stat.Loses
		summary.Cash += stat.Cash
		summary.SideWinner1 += stat.SideWinner1
		summary.SideWinner2 += stat.SideWinner2
		summary.SideWinner3 += stat.SideWinner3
		summary.SideWinner4 += stat.SideWinner4
		summary.SideWinner5 += stat.SideWinner5
		summary.SideWinner6 += stat.SideWinner6
		summary.SeatBets1 += stat.SeatBets1
		summary.SeatBets2 += stat.SeatBets2
		summary.SeatBets3 += stat.SeatBets3
		summary.SeatBets4 += stat.SeatBets4
		summary.SeatBets5 += stat.SeatBets5
		summary.SeatBets6 += stat.SeatBets6
		summary.PlayerWinSeat1 += stat.PlayerWinSeat1
		summary.PlayerWinSeat2 += stat.PlayerWinSeat2
		summary.PlayerWinSeat3 += stat.PlayerWinSeat3
		summary.PlayerWinSeat4 += stat.PlayerWinSeat4
		summary.PlayerWinSeat5 += stat.PlayerWinSeat5
		summary.PlayerWinSeat6 += stat.PlayerWinSeat6
		summary.PlayerMultis += stat.PlayerMultis
		summary.RoundBetAvg += stat.RoundBetAvg

		summary.ObserveRounds += stat.ObserveRounds
		summary.Players += stat.Players
		summary.NewPlayers += stat.NewPlayers
		summary.OldPlayers += stat.OldPlayers
		summary.AllRoundsList = append(summary.AllRoundsList, stat.AllRoundsList...)

		// 玩 crash 人数占日活比
		if stat.FLoginedUsers > 0 {
			stat.FPlayerRate = fmt.Sprintf("%.2f", float64(stat.Players)/float64(stat.FLoginedUsers))
		}

		// 查询当天注册的新用户
		startTime := utils.Str2Time(fmt.Sprintf("%s 00:00:00", dateStr), Location())
		endTime := utils.Str2Time(fmt.Sprintf("%s 23:59:59", dateStr), Location())
		m7 := []bson.M{
			{"$match": bson.M{
				"ctime":            bson.M{"$gte": startTime, "$lte": endTime},
				"robot":            false,
				"simulation_robot": false,
			}},
			{"$group": bson.M{
				"_id":   nil,
				"total": bson.M{"$sum": 1},
			}},
		}
		var r7 []bson.M
		err := PlayerUsers.Pipe(m7).All(&r7)
		if err != nil {
			beego.Error("查询当天注册的新用户 error: ", err)
		} else if len(r7) > 0 {
			stat.FLoginedNewUsers = r7[0]["total"].(int)
		}
		stat.FLoginedOldUsers = stat.FLoginedUsers - stat.FLoginedNewUsers
		if stat.FLoginedOldUsers < 0 {
			stat.FLoginedOldUsers = 0
		}
		summary.FLoginedNewUsers += stat.FLoginedNewUsers
		summary.FLoginedOldUsers += stat.FLoginedOldUsers
		// FNewPlayerRate, FOldPlayerRate 占新老玩家日活
		if stat.FLoginedNewUsers > 0 {
			stat.FNewPlayerRate = fmt.Sprintf("%.2f", float64(stat.NewPlayers)/float64(stat.FLoginedNewUsers))
		}
		if stat.FLoginedOldUsers > 0 {
			stat.FOldPlayerRate = fmt.Sprintf("%.2f", float64(stat.OldPlayers)/float64(stat.FLoginedOldUsers))
		}

		if stat.Players > 0 {
			stat.FPlayerRoundsAvg = fmt.Sprintf("%.2f", float64(stat.BetRounds)/float64(stat.Players))
		}
		// 局数中位数
		if len(stat.AllRoundsList) > 0 {
			var rounds = stat.AllRoundsList
			sort.Slice(rounds, func(i, j int) bool {
				return rounds[i] < rounds[j]
			})
			length := len(rounds)
			if length > 0 {
				if length%2 == 1 {
					stat.FPlayerRoundsMedian = rounds[length/2]
				} else {
					// 中间两位算平均
					stat.FPlayerRoundsMedian = (rounds[length/2] + rounds[length/2-1]) / 2
				}
			}
		}
		// stat.FGameTimes = fmt.Sprintf("%.2f", float64(stat.GameTimes)/60)
		if stat.BetRounds > 0 {
			stat.FGameTimesAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.BetRounds)/60)
		}
		if stat.Players > 0 {
			stat.FGameTimesPlayerAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.Players)/60)
		}

		stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
		if stat.Players > 0 {
			stat.FPlayerBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.Players)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FRoundBetsAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.Bets, 100)/float64(stat.BetRounds))
			// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
		}
		stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
		stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
		stat.FTieBets = fmt.Sprintf("%.2f", float64(stat.TieBets)/100.0)
		if stat.BetRounds > 0 {
			stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
		}
		if stat.Bets > 0 {
			stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
		}
		if stat.WinRounds > 0 {
			stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
		}
		if stat.TieRounds > 0 {
			stat.FTieBetsAvg = fmt.Sprintf("%.2f", float64(stat.TieBets)/float64(stat.TieRounds)/100.0)
		}
		stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
		stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
		stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)

		if stat.WinRounds > 0 {
			stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
		}
		// 下注开奖
		if stat.BetRounds > 0 {
			stat.SideWinnerRate1 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner1)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate2 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner2)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate3 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner3)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate4 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner4)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate5 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner5)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate6 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner6)/float64(stat.BetRounds)*100)
		}

		//玩家下注
		if stat.BetRounds > 0 {
			stat.SeatBetsRate1 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets1)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate2 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets2)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate3 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets3)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate4 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets4)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate5 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets5)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate6 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets6)/float64(stat.BetRounds)*100)
		}
		// 玩家下注中奖
		if stat.SeatBets1 > 0 {
			stat.PlayerWinSeatRate1 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat1)/float64(stat.SeatBets1)*100)
		}
		if stat.SeatBets2 > 0 {
			stat.PlayerWinSeatRate2 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat2)/float64(stat.SeatBets2)*100)
		}
		if stat.SeatBets3 > 0 {
			stat.PlayerWinSeatRate3 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat3)/float64(stat.SeatBets3)*100)
		}
		if stat.SeatBets4 > 0 {
			stat.PlayerWinSeatRate4 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat4)/float64(stat.SeatBets4)*100)
		}
		if stat.SeatBets5 > 0 {
			stat.PlayerWinSeatRate5 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat5)/float64(stat.SeatBets5)*100)
		}
		if stat.SeatBets6 > 0 {
			stat.PlayerWinSeatRate6 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat6)/float64(stat.SeatBets6)*100)
		}
		// 多门率
		if stat.BetRounds > 0 {
			stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerMultis)/float64(stat.BetRounds)*100)
		}
		// 观察局
		if stat.BetRounds > 0 {
			stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
		}
		if stat.BetRounds > 0 {
			stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
		}

	}
}

func CPStatDatesSummaryMapping(stat *entity.CPPlayerStat) {
	// 玩 lhd 人数占日活比
	if stat.FLoginedUsers > 0 {
		stat.FPlayerRate = fmt.Sprintf("%.2f", float64(stat.Players)/float64(stat.FLoginedUsers))
	}

	// FNewPlayerRate, FOldPlayerRate 占新老玩家日活
	if stat.FLoginedNewUsers > 0 {
		stat.FNewPlayerRate = fmt.Sprintf("%.2f", float64(stat.NewPlayers)/float64(stat.FLoginedNewUsers))
	}
	if stat.FLoginedOldUsers > 0 {
		stat.FOldPlayerRate = fmt.Sprintf("%.2f", float64(stat.OldPlayers)/float64(stat.FLoginedOldUsers))
	}

	if stat.Players > 0 {
		stat.FPlayerRoundsAvg = fmt.Sprintf("%.2f", float64(stat.BetRounds)/float64(stat.Players))
	}
	// 局数中位数
	if len(stat.AllRoundsList) > 0 {
		var rounds = stat.AllRoundsList
		sort.Slice(rounds, func(i, j int) bool {
			return rounds[i] < rounds[j]
		})
		length := len(rounds)
		if length > 0 {
			if length%2 == 1 {
				stat.FPlayerRoundsMedian = rounds[length/2]
			} else {
				// 中间两位算平均
				stat.FPlayerRoundsMedian = (rounds[length/2] + rounds[length/2-1]) / 2
			}
		}
	}
	// stat.FGameTimes = fmt.Sprintf("%.2f", float64(stat.GameTimes)/60)
	if stat.BetRounds > 0 {
		stat.FGameTimesAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.BetRounds)/60)
	}
	if stat.Players > 0 {
		stat.FGameTimesPlayerAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.Players)/60)
	}

	if stat.BetRounds > 0 {
		stat.FGameTimesAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.BetRounds)/60)
	}
	if stat.Players > 0 {
		stat.FGameTimesPlayerAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.Players)/60)
	}

	stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
	if stat.Players > 0 {
		stat.FPlayerBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.Players)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FRoundBetsAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.Bets, 100)/float64(stat.BetRounds))
		// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
	}
	stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
	stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
	stat.FTieBets = fmt.Sprintf("%.2f", float64(stat.TieBets)/100.0)

	if stat.BetRounds > 0 {
		stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
	}
	if stat.Bets > 0 {
		stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
	}
	if stat.WinRounds > 0 {
		stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
	}
	stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
	stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
	stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)

	if stat.WinRounds > 0 {
		stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
	}
	if stat.TieRounds > 0 {
		stat.FTieBetsAvg = fmt.Sprintf("%.2f", float64(stat.TieBets)/float64(stat.TieRounds)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
	}
	// 下注开奖
	if stat.BetRounds > 0 {
		stat.SideWinnerRate1 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner1)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate2 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner2)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate3 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner3)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate4 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner4)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate5 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner5)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate6 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner6)/float64(stat.BetRounds)*100)
	}

	//玩家下注
	if stat.BetRounds > 0 {
		stat.SeatBetsRate1 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets1)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate2 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets2)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate3 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets3)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate4 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets4)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate5 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets5)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate6 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets6)/float64(stat.BetRounds)*100)
	}
	// 玩家下注中奖
	if stat.SeatBets1 > 0 {
		stat.PlayerWinSeatRate1 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat1)/float64(stat.SeatBets1)*100)
	}
	if stat.SeatBets2 > 0 {
		stat.PlayerWinSeatRate2 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat2)/float64(stat.SeatBets2)*100)
	}
	if stat.SeatBets3 > 0 {
		stat.PlayerWinSeatRate3 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat3)/float64(stat.SeatBets3)*100)
	}
	if stat.SeatBets4 > 0 {
		stat.PlayerWinSeatRate4 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat4)/float64(stat.SeatBets4)*100)
	}
	if stat.SeatBets5 > 0 {
		stat.PlayerWinSeatRate5 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat5)/float64(stat.SeatBets5)*100)
	}
	if stat.SeatBets6 > 0 {
		stat.PlayerWinSeatRate6 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat6)/float64(stat.SeatBets6)*100)
	}
	// 多门率
	if stat.BetRounds > 0 {
		stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerMultis)/float64(stat.BetRounds)*100)
	}

	// 观察局
	if stat.BetRounds > 0 {
		stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
	}
	if stat.BetRounds > 0 {
		stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
	}
}

// CP 来玩就赢
func (s *gameStatsService) GetCPLwjyStat(page, pageSize int, m, m2, m3 bson.M, startDate, endDate string) (
	stats []*entity.CPLwjyStat, count int, err error,
) {
	// GetCPLwjyStat
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)

	// 用户列表筛选
	var filterPlayerIds []string
	if len(m2) > 0 || len(m3) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				"_id":        "$userid",
				"bets":       bson.M{"$sum": "$strategy_bets"},
				"bet_rounds": bson.M{"$sum": "$strategy_times"},
				"cash":       bson.M{"$sum": "$cash"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				"cash": "$cash",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = CPLwjyStats.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetCPLwjyStat error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			userid := r["_id"].(string)
			userids = append(userids, userid)
		}
		if len(userids) == 0 { // 未匹配到
			return
		}
		now := utils.BsonNow()
		m3_2 := []bson.M{
			{"$match": bson.M{"_id": bson.M{"$in": userids}}},
			{"$project": bson.M{
				"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
				"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
			}},
			{"$match": m3},
		}
		var r3_2 []bson.M
		err = PlayerUsers.Pipe(m3_2).All(&r3_2)
		if err != nil {
			beego.Error("cp PlayerUsers error32:", err)
		} else if len(r3_2) == 0 { // 未匹配到
			return
		}
		for _, r := range r3_2 {
			userid := r["_id"].(string)
			filterPlayerIds = append(filterPlayerIds, userid)
		}
	}
	if len(filterPlayerIds) > 0 {
		m["userid"] = bson.M{"$in": filterPlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":  "$userid",
			"date": bson.M{"$max": "$date"},
			// "date_str":                bson.M{"$sum": "$date_str"},
			"strategy_players":      bson.M{"$max": "$strategy_players"},
			"strategy_times":        bson.M{"$max": "$strategy_times"},
			"tigger_times":          bson.M{"$max": "$tigger_times"},
			"all_evo_times":         bson.M{"$sum": "$all_evo_times"},
			"strategy_multi_rounds": bson.M{"$sum": "$strategy_multi_rounds"},
			"strategy_bets":         bson.M{"$sum": "$strategy_bets"},
			"strategy_win_rounds":   bson.M{"$sum": "$strategy_win_rounds"},
			"strategy_wins":         bson.M{"$sum": "$strategy_wins"},
			"strategy_lose_rounds":  bson.M{"$sum": "$strategy_lose_rounds"},
			"strategy_loses":        bson.M{"$sum": "$strategy_loses"},
		}},
		{"$project": bson.M{
			"_id":                   "$_id",
			"date":                  "$date",
			"userid":                "$userid",
			"strategy_players":      "$strategy_players",
			"strategy_times":        "$strategy_times",
			"tigger_times":          "$tigger_times",
			"all_evo_times":         "$all_evo_times",
			"strategy_multi_rounds": "$strategy_multi_rounds",
			"strategy_bets":         "$strategy_bets",
			"strategy_win_rounds":   "$strategy_win_rounds",
			"strategy_wins":         "$strategy_wins",
			"strategy_loses":        "$strategy_loses",
			"strategy_lose_rounds":  "$strategy_lose_rounds",
		}},
		{"$match": m2},
		{"$count": "count"},
	}
	// 查总条数
	var rCount []bson.M
	err = CPLwjyStats.Pipe(m1).All(&rCount)
	if err != nil {
		return
	} else if len(rCount) > 0 {
		count = rCount[0]["count"].(int)
	}

	// 查列表数据
	m1 = m1[:len(m1)-1]
	m1 = append(m1, []bson.M{
		{"$sort": bson.M{"date": -1}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}...)
	err = CPLwjyStats.Pipe(m1).All(&stats)
	if err != nil {
		return
	}
	CPLwjyStatMapping(stats)

	// 查总汇数据
	mark := "--"
	summary := &entity.CPLwjyStat{}
	m1 = m1[:len(m1)-3]
	m1[1]["$group"].(bson.M)["_id"] = nil
	delete(m1[1]["$group"].(bson.M), "date")
	m1[1]["$group"].(bson.M)["userids"] = bson.M{"$addToSet": "$userid"}
	m1[1]["$group"].(bson.M)["strategy_players"] = bson.M{"$sum": "$strategy_players"} // 触发策略人数总汇
	m1[1]["$group"].(bson.M)["strategy_times"] = bson.M{"$sum": "$strategy_times"}
	m1[1]["$group"].(bson.M)["tigger_times"] = bson.M{"$sum": "$tigger_times"}
	m1[2]["$project"].(bson.M)["userids"] = "$userids"
	str, err := json.Marshal(m1)
	fmt.Println(err, string(str))
	var summarys []*entity.CPLwjyStat
	err = CPLwjyStats.Pipe(m1).All(&summarys)
	if err != nil {
		beego.Error("lhd summary error:", err)
		return
	} else if len(summarys) > 0 {
		summary = summarys[0]
		summary.No = fmt.Sprintf("%s-%s总汇", startDate, endDate)
		summary.Id = mark
		summary.FNickname = mark
		summary.FUserType = mark
		summary.FRegistArea = mark
		summary.FCtime = mark
		summary.FLoginTime = mark
		CPLwjyStatSummaryMapping(summary)
	}
	if len(summarys) > 0 {
		stats = append([]*entity.CPLwjyStat{summary}, stats...)
	}
	return
}

func CPLwjyStatMapping(stats []*entity.CPLwjyStat) {
	var playerIds []string
	// var playerStats = make(map[string]*entity.CrashPlayerStat, len(stats))
	for _, stat := range stats {
		playerIds = append(playerIds, stat.Id)
		// playerStats[stat.Id] = stat
	}
	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": playerIds}}},
		{"$project": bson.M{
			"_id":         "$_id",
			"nickname":    "$nickname",
			"money":       "$money",
			"cash_out":    "$cash_out",
			"diamond":     "$diamond",
			"ctime":       "$ctime",
			"login_time":  "$login_time",
			"regist_area": "$regist_area", //ab测试 0:A 1:B 2:C
			"state":       "$state",
			"cp_strategy": "$cp_strategy",
		}},
	}
	var r2 []*entity.PlayerUser
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("cp users error:", err)
		return
	}

	var r2Map = make(map[string]*entity.PlayerUser, len(r2))
	for _, r := range r2 {
		userid := r.Userid
		r2Map[userid] = r
	}
	now := utils.BsonNow()
	for i, stat := range stats {
		r, ok := r2Map[stat.Id]
		if !ok {
			continue
		}
		// 玩家昵称	玩家类型 账号类型 注册时间 最后登陆时间 存活天数 流失天数
		nickname := r.Nickname
		money := r.Money
		ctime := r.Ctime
		login_time := r.LoginTime
		regist_area := r.RegistArea
		state := r.State

		// 将日期转成印度时间
		login_time, _ = ConvertToIndiaTime(login_time.Unix())
		ctime, _ = ConvertToIndiaTime(ctime.Unix())

		stat.No = strconv.Itoa(i + 1)
		stat.FNickname = nickname
		stat.FCtime = ctime.Format(utils.FORMAT)
		stat.FLoginTime = login_time.Format(utils.FORMAT)
		stat.FLiveDays = int(math.Ceil(login_time.Sub(ctime).Hours() / 24))
		stat.FLoseDays = int(math.Floor(now.Sub(login_time).Hours() / 24))

		switch regist_area {
		case 0:
			stat.FRegistArea = "A类"
		case 1:
			stat.FRegistArea = "B类"
		case 2:
			stat.FRegistArea = "C类"
		}

		switch state {
		case 1:
			stat.FUserType = "新手"
		case 2:
			if money == 0 {
				stat.FUserType = "零充"
			} else if money >= 20000 && money <= 99999 {
				stat.FUserType = "普充"
			} else if money >= 100000 && money <= 499999 {
				stat.FUserType = "小R"
			} else if money >= 500000 && money <= 999999 {
				stat.FUserType = "中R"
			} else if money >= 1000000 && money <= 9999999 {
				stat.FUserType = "大R"
			} else if money >= 10000000 {
				stat.FUserType = "超大R"
			}
		case 3:
			stat.FUserType = "平民"
		case 4:
			stat.FUserType = "泡沫"
		}
		if stat.StrategyTimes > 0 {
			stat.TriggerRate = fmt.Sprintf("%.2f%%", float64(stat.TiggerTimes)/float64(stat.StrategyTimes)*100.0)

		}
		if stat.StrategyPlayers > 0 {
			stat.PerCapitaAvg = fmt.Sprintf("%.2f", float64(stat.TiggerTimes)/float64(stat.StrategyPlayers))
		}
		if stat.StrategyTimes > 0 {
			stat.TiggerStrategyAvg = fmt.Sprintf("%.2f", float64(stat.AllEvoTimes)/float64(stat.StrategyTimes))

		}
		if stat.TiggerTimes > 0 {
			stat.YxTiggerStrategyAvg = fmt.Sprintf("%.2f", float64(stat.AllEvoTimes)/float64(stat.TiggerTimes))
		}
		// 多门率
		if stat.AllEvoTimes > 0 {
			stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyMultiRounds)/float64(stat.AllEvoTimes)*100.0)
		}
		stat.FStrategyBets = fmt.Sprintf("%.2f", float64(stat.StrategyBets)/100.0)
		if stat.StrategyTimes > 0 {
			stat.FStrategyBetRoundsAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.StrategyBets, 100)/float64(stat.AllEvoTimes))
		}
		stat.FStrategyWins = fmt.Sprintf("%.2f", float64(stat.StrategyWins)/100.0)
		stat.FStrategyLoses = fmt.Sprintf("%.2f", float64(stat.StrategyLoses)/100.0)
		stat.FStrategyCash = fmt.Sprintf("%.2f", float64(stat.StrategyWins+stat.StrategyLoses)/100.0)

		if stat.AllEvoTimes > 0 {
			stat.FStrategyWinRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyWinRounds)/float64(stat.AllEvoTimes)*100.0)
		}
		if -stat.StrategyLoses > 0 {
			stat.FStrategyRebateRate = fmt.Sprintf("%.2f%%", (float64(stat.StrategyWins)/float64(-stat.StrategyLoses))*100.0)
		} else {
			stat.FStrategyRebateRate = "未输过"
		}
	}
}

func CPLwjyStatSummaryMapping(stat *entity.CPLwjyStat) {
	stat.StrategyPlayers = int32(len(stat.Userids))
	if stat.StrategyTimes > 0 {
		stat.TriggerRate = fmt.Sprintf("%.2f%%", float64(stat.TiggerTimes)/float64(stat.StrategyTimes)*100.0)

	}
	if stat.StrategyPlayers > 0 {
		stat.PerCapitaAvg = fmt.Sprintf("%.2f", float64(stat.TiggerTimes)/float64(stat.StrategyPlayers))
	}
	if stat.StrategyTimes > 0 {
		stat.TiggerStrategyAvg = fmt.Sprintf("%.2f", float64(stat.AllEvoTimes)/float64(stat.StrategyTimes))
	}
	if stat.TiggerTimes > 0 {
		stat.YxTiggerStrategyAvg = fmt.Sprintf("%.2f", float64(stat.AllEvoTimes)/float64(stat.TiggerTimes))
	}
	// 多门率
	if stat.AllEvoTimes > 0 {
		stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyMultiRounds)/float64(stat.AllEvoTimes)*100.0)
	}
	stat.FStrategyBets = fmt.Sprintf("%.2f", float64(stat.StrategyBets)/100.0)
	if stat.StrategyTimes > 0 {
		stat.FStrategyBetRoundsAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.StrategyBets, 100)/float64(stat.AllEvoTimes))
	}
	stat.FStrategyWins = fmt.Sprintf("%.2f", float64(stat.StrategyWins)/100.0)
	stat.FStrategyLoses = fmt.Sprintf("%.2f", float64(stat.StrategyLoses)/100.0)
	stat.FStrategyCash = fmt.Sprintf("%.2f", float64(stat.StrategyWins+stat.StrategyLoses)/100.0)

	if stat.AllEvoTimes > 0 {
		stat.FStrategyWinRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyWinRounds)/float64(stat.AllEvoTimes)*100.0)
	}
	if -stat.StrategyLoses > 0 {
		// stat.FStrategyRebateRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.StrategyWins, 100)/ComputeFloat(stat.StrategyLoses, 100))*100.0)
		stat.FStrategyRebateRate = fmt.Sprintf("%.2f%%", (float64(stat.StrategyWins)/float64(-stat.StrategyLoses))*100.0)
	} else {
		stat.FStrategyRebateRate = "未输过"
	}
}

// CP 来易去难
func (s *gameStatsService) GetCPLyqnStat(page, pageSize int, m, m2, m3 bson.M, startDate, endDate string) (
	stats []*entity.CPLyqnStat, count int, err error,
) {
	// UpDownXxscStat
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)

	// 用户列表筛选
	var filterPlayerIds []string
	if len(m2) > 0 || len(m3) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				"_id":        "$userid",
				"bets":       bson.M{"$sum": "$strategy_bets"},
				"bet_rounds": bson.M{"$sum": "$strategy_rounds"},
				"cash":       bson.M{"$sum": "$cash"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				"cash": "$cash",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = CPLyqnStats.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetCPLyqnStat error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			userid := r["_id"].(string)
			userids = append(userids, userid)
		}
		if len(userids) == 0 { // 未匹配到
			return
		}
		now := utils.BsonNow()
		m3_2 := []bson.M{
			{"$match": bson.M{"_id": bson.M{"$in": userids}}},
			{"$project": bson.M{
				"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
				"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
			}},
			{"$match": m3},
		}
		var r3_2 []bson.M
		err = PlayerUsers.Pipe(m3_2).All(&r3_2)
		if err != nil {
			beego.Error("lhd PlayerUsers error32:", err)
		} else if len(r3_2) == 0 { // 未匹配到
			return
		}
		for _, r := range r3_2 {
			userid := r["_id"].(string)
			filterPlayerIds = append(filterPlayerIds, userid)
		}
	}
	if len(filterPlayerIds) > 0 {
		m["userid"] = bson.M{"$in": filterPlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":  "$userid",
			"date": bson.M{"$max": "$date"},
			// "date_str":                bson.M{"$sum": "$date_str"},
			"strategy_players":      bson.M{"$max": "$strategy_players"},
			"strategy_rounds":       bson.M{"$sum": "$strategy_rounds"},
			"all_in_rounds":         bson.M{"$sum": "$all_in_rounds"},
			"strategy_multi_rounds": bson.M{"$sum": "$strategy_multi_rounds"},
			"strategy_bets":         bson.M{"$sum": "$strategy_bets"},
			"strategy_wins":         bson.M{"$sum": "$strategy_wins"},
			"strategy_win_rounds":   bson.M{"$sum": "$strategy_win_rounds"},
			"strategy_loses":        bson.M{"$sum": "$strategy_loses"},
			"strategy_lose_rounds":  bson.M{"$sum": "$strategy_lose_rounds"},
			"bet_rounds":            bson.M{"$sum": "$bet_rounds"},
			"win_rounds":            bson.M{"$sum": "$win_rounds"},
			"lose_rounds":           bson.M{"$sum": "$lose_rounds"},
			"wins":                  bson.M{"$sum": "$wins"},
			"loses":                 bson.M{"$sum": "$loses"},
		}},
		{"$project": bson.M{
			"_id":                   "$_id",
			"date":                  "$date",
			"userid":                "$userid",
			"strategy_players":      "$strategy_players",
			"all_in_rounds":         "$all_in_rounds",
			"strategy_rounds":       "$strategy_rounds",
			"strategy_multi_rounds": "$strategy_multi_rounds",
			"strategy_bets":         "$strategy_bets",
			"strategy_win_rounds":   "$strategy_win_rounds",
			"strategy_wins":         "$strategy_wins",
			"strategy_loses":        "$strategy_loses",
			"strategy_lose_rounds":  "$strategy_lose_rounds",
			"bet_rounds":            "$bet_rounds",
			"win_rounds":            "$win_rounds",
			"lose_rounds":           "$lose_rounds",
			"wins":                  "$wins",
			"loses":                 "$loses",
		}},
		{"$match": m2},
		{"$count": "count"},
	}
	// 查总条数
	var rCount []bson.M
	err = CPLyqnStats.Pipe(m1).All(&rCount)
	if err != nil {
		return
	} else if len(rCount) > 0 {
		count = rCount[0]["count"].(int)
	}

	// 查列表数据
	m1 = m1[:len(m1)-1]
	m1 = append(m1, []bson.M{
		{"$sort": bson.M{"date": -1}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}...)
	err = CPLyqnStats.Pipe(m1).All(&stats)
	if err != nil {
		return
	}
	CPLyqnStatMapping(stats, m)

	// 查总汇数据
	mark := "--"
	summary := &entity.CPLyqnStat{}
	m1 = m1[:len(m1)-3]
	m1[1]["$group"].(bson.M)["_id"] = nil
	delete(m1[1]["$group"].(bson.M), "date")
	m1[1]["$group"].(bson.M)["userids"] = bson.M{"$addToSet": "$userid"}
	m1[1]["$group"].(bson.M)["strategy_players"] = bson.M{"$sum": "$strategy_players"} // 触发策略人数总汇

	m1[2]["$project"].(bson.M)["userids"] = "$userids"
	str, err := json.Marshal(m1)
	fmt.Println(err, string(str))
	var summarys []*entity.CPLyqnStat
	err = CPLyqnStats.Pipe(m1).All(&summarys)
	if err != nil {
		beego.Error("lhd summary error:", err)
		return
	} else if len(summarys) > 0 {
		summary = summarys[0]
		summary.No = fmt.Sprintf("%s-%s总汇", startDate, endDate)
		summary.Id = mark
		summary.FNickname = mark
		summary.FUserType = mark
		summary.FRegistArea = mark
		summary.FCtime = mark
		summary.FLoginTime = mark
		CPLyqnStatSummaryMapping(summary, m)
	}
	if len(summarys) > 0 {
		stats = append([]*entity.CPLyqnStat{summary}, stats...)
	}
	return
}

func CPLyqnStatMapping(stats []*entity.CPLyqnStat, m bson.M) {
	var playerIds []string
	// var playerStats = make(map[string]*entity.CrashPlayerStat, len(stats))
	for _, stat := range stats {
		playerIds = append(playerIds, stat.Id)
		// playerStats[stat.Id] = stat
	}
	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": playerIds}}},
		{"$project": bson.M{
			"_id":          "$_id",
			"nickname":     "$nickname",
			"money":        "$money",
			"cash_out":     "$cash_out",
			"diamond":      "$diamond",
			"ctime":        "$ctime",
			"login_time":   "$login_time",
			"regist_area":  "$regist_area", //ab测试 0:A 1:B 2:C
			"state":        "$state",
			"lhd_strategy": "$lhd_strategy",
		}},
	}
	var r2 []*entity.PlayerUser
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("lhd users error:", err)
		return
	}

	var r2Map = make(map[string]*entity.PlayerUser, len(r2))
	for _, r := range r2 {
		userid := r.Userid
		r2Map[userid] = r
	}
	// 查询总数据
	var cplist []entity.CPPlayerStat
	n := bson.M{}
	// 手动复制 m 到 n
	for k, v := range m {
		n[k] = v
	}
	n["userid"] = bson.M{"$in": playerIds}
	m3 := []bson.M{
		{"$match": n},
		{"$group": bson.M{
			"_id":        "$userid",
			"bet_rounds": bson.M{"$sum": "$bet_rounds"},
			"win_rounds": bson.M{"$sum": "$win_rounds"},
			"bets":       bson.M{"$sum": "$bets"},
			"win_bets":   bson.M{"$sum": "$win_bets"},
			"lose_bets":  bson.M{"$sum": "$lose_bets"},
			"wins":       bson.M{"$sum": "$wins"},
			"loses":      bson.M{"$sum": "$loses"},
		}},
	}
	CPPlayerStats.Pipe(m3).All(&cplist)

	now := utils.BsonNow()
	for i, stat := range stats {
		r, ok := r2Map[stat.Id]
		if !ok {
			continue
		}
		// 玩家昵称	玩家类型 账号类型 注册时间 最后登陆时间 存活天数 流失天数
		nickname := r.Nickname
		money := r.Money
		ctime := r.Ctime
		login_time := r.LoginTime
		regist_area := r.RegistArea
		state := r.State

		// 将日期转成印度时间
		login_time, _ = ConvertToIndiaTime(login_time.Unix())
		ctime, _ = ConvertToIndiaTime(ctime.Unix())

		stat.No = strconv.Itoa(i + 1)
		stat.FNickname = nickname
		stat.FCtime = ctime.Format(utils.FORMAT)
		stat.FLoginTime = login_time.Format(utils.FORMAT)
		stat.FLiveDays = int(math.Ceil(login_time.Sub(ctime).Hours() / 24))
		stat.FLoseDays = int(math.Floor(now.Sub(login_time).Hours() / 24))

		switch regist_area {
		case 0:
			stat.FRegistArea = "A类"
		case 1:
			stat.FRegistArea = "B类"
		case 2:
			stat.FRegistArea = "C类"
		}

		switch state {
		case 1:
			stat.FUserType = "新手"
		case 2:
			if money == 0 {
				stat.FUserType = "零充"
			} else if money >= 20000 && money <= 99999 {
				stat.FUserType = "普充"
			} else if money >= 100000 && money <= 499999 {
				stat.FUserType = "小R"
			} else if money >= 500000 && money <= 999999 {
				stat.FUserType = "中R"
			} else if money >= 1000000 && money <= 9999999 {
				stat.FUserType = "大R"
			} else if money >= 10000000 {
				stat.FUserType = "超大R"
			}
		case 3:
			stat.FUserType = "平民"
		case 4:
			stat.FUserType = "泡沫"
		}
		if stat.StrategyPlayers > 0 {
			stat.TriggerRoundsAvg = fmt.Sprintf("%.2f", float64(stat.StrategyRounds)/float64(stat.StrategyPlayers))
		}
		if stat.AllInRounds > 0 {
			stat.TiggerRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyRounds)/float64(stat.AllInRounds)*100.0)
		}
		if stat.StrategyRounds > 0 {
			stat.ExpiredRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyLoseRounds)/float64(stat.StrategyRounds)*100.0)
		}
		// 多门率
		if stat.StrategyRounds > 0 {
			stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyMultiRounds)/float64(stat.StrategyRounds)*100.0)
		}
		stat.FStrategyBets = fmt.Sprintf("%.2f", float64(stat.StrategyBets)/100.0)
		if stat.StrategyRounds > 0 {
			stat.FStrategyBetRoundsAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.StrategyBets, 100)/float64(stat.StrategyRounds))
		}
		stat.FStrategyWins = fmt.Sprintf("%.2f", float64(stat.StrategyWins)/100.0)
		stat.FStrategyLoses = fmt.Sprintf("%.2f", float64(stat.StrategyLoses)/100.0)
		stat.FStrategyCash = fmt.Sprintf("%.2f", float64(stat.StrategyWins+stat.StrategyLoses)/100.0)
		if len(cplist) > 0 {
			for _, c := range cplist {
				if c.Id == stat.Id {
					stat.WinRounds = c.WinRounds
					stat.BetRounds = c.BetRounds
					stat.Wins = c.Wins
					stat.Loses = c.Loses
					break
				}
			}
		}
		if stat.BetRounds > 0 {
			stat.TotalWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
		}
		if stat.StrategyRounds > 0 {
			stat.StrategyWinRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyWinRounds)/float64(stat.StrategyRounds)*100.0)
		}
		if -stat.StrategyLoses > 0 {
			stat.StrategyRewardRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.StrategyWins, 100)/ComputeFloat(-stat.StrategyLoses, 100))*100.0)
		}
		if -stat.Loses > 0 {
			stat.TotalRewardRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.Wins, 100)/ComputeFloat(-stat.Loses, 100))*100.0)
		}
		if r.Money > 0 {
			stat.PlayerRewardRate = fmt.Sprintf("%.2f%%", ComputeFloat(int64(r.CashOut+int32(r.Diamond)), 100)/ComputeFloat(int64(r.Money), 100)*100.0)
			//fmt.Sprintf("%d", int((int64(r.CashOut)+r.Diamond)*10000/int64(r.Money)))
		} else {
			stat.PlayerRewardRate = "未充值"
		}
	}
}

func CPLyqnStatSummaryMapping(stat *entity.CPLyqnStat, m bson.M) {
	// 查询总数据
	cpinfo := new(entity.CPPlayerStat)
	n := bson.M{}
	// 手动复制 m 到 n
	for k, v := range m {
		n[k] = v
	}
	n["userid"] = bson.M{"$in": stat.Userids}
	m3 := []bson.M{
		{"$match": n},
		{"$group": bson.M{
			"_id":        nil,
			"bet_rounds": bson.M{"$sum": "$bet_rounds"},
			"win_rounds": bson.M{"$sum": "$win_rounds"},
			"bets":       bson.M{"$sum": "$bets"},
			"win_bets":   bson.M{"$sum": "$win_bets"},
			"lose_bets":  bson.M{"$sum": "$lose_bets"},
			"wins":       bson.M{"$sum": "$wins"},
			"loses":      bson.M{"$sum": "$loses"},
		}},
	}
	CPPlayerStats.Pipe(m3).One(&cpinfo)

	stat.StrategyPlayers = int32(len(stat.Userids))
	if stat.StrategyPlayers > 0 {
		stat.TriggerRoundsAvg = fmt.Sprintf("%.2f", float64(stat.StrategyRounds)/float64(stat.StrategyPlayers))
	}
	if stat.AllInRounds > 0 {
		stat.TiggerRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyRounds)/float64(stat.AllInRounds)*100.0)
	}
	if stat.StrategyRounds > 0 {
		stat.ExpiredRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyLoseRounds)/float64(stat.StrategyRounds)*100.0)
	}
	// 多门率
	if stat.StrategyRounds > 0 {
		stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyMultiRounds)/float64(stat.StrategyRounds)*100.0)
	}
	stat.FStrategyBets = fmt.Sprintf("%.2f", float64(stat.StrategyBets)/100.0)
	if stat.StrategyRounds > 0 {
		stat.FStrategyBetRoundsAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.StrategyBets, 100)/float64(stat.StrategyRounds))
	}
	stat.FStrategyWins = fmt.Sprintf("%.2f", float64(stat.StrategyWins)/100.0)
	stat.FStrategyLoses = fmt.Sprintf("%.2f", float64(stat.StrategyLoses)/100.0)
	stat.FStrategyCash = fmt.Sprintf("%.2f", float64(stat.StrategyWins+stat.StrategyLoses)/100.0)
	// 查询总局数
	stat.WinRounds = cpinfo.WinRounds
	stat.BetRounds = cpinfo.BetRounds
	stat.Wins = cpinfo.Wins
	stat.Loses = cpinfo.Loses
	if stat.BetRounds > 0 {
		stat.TotalWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
	}
	if stat.StrategyRounds > 0 {
		stat.StrategyWinRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyWinRounds)/float64(stat.StrategyRounds)*100.0)
	}
	if -stat.StrategyLoses > 0 {
		stat.StrategyRewardRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.StrategyWins, 100)/ComputeFloat(-stat.StrategyLoses, 100))*100.0)
	}
	if -stat.Loses > 0 {
		stat.TotalRewardRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.Wins, 100)/ComputeFloat(-stat.Loses, 100))*100.0)
	}
}

/*
RB游戏策略
*/
func (s *gameStatsService) GetRBStatPlayers(page, pageSize int, m, m2, m3 bson.M, startDate, endDate string) (
	stats []*entity.RBPlayerStat, count int, err error,
) {
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)

	// 用户列表筛选
	var filterPlayerIds []string
	if len(m2) > 0 || len(m3) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				"_id":        "$userid",
				"bets":       bson.M{"$sum": "$bets"},
				"bet_rounds": bson.M{"$sum": "$bet_rounds"},
				"cash":       bson.M{"$sum": "$cash"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				"cash": "$cash",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = RBPlayerStats.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetRBStatPlayers error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			userid := r["_id"].(string)
			userids = append(userids, userid)
		}
		if len(userids) == 0 { // 未匹配到
			return
		}
		now := utils.BsonNow()
		m3_2 := []bson.M{
			{"$match": bson.M{"_id": bson.M{"$in": userids}}},
			{"$project": bson.M{
				"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
				"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
			}},
			{"$match": m3},
		}
		var r3_2 []bson.M
		err = PlayerUsers.Pipe(m3_2).All(&r3_2)
		if err != nil {
			beego.Error("GetRBStatPlayers error32:", err)
		} else if len(r3_2) == 0 { // 未匹配到
			return
		}
		for _, r := range r3_2 {
			userid := r["_id"].(string)
			filterPlayerIds = append(filterPlayerIds, userid)
		}
	}
	if len(filterPlayerIds) > 0 {
		m["userid"] = bson.M{"$in": filterPlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":              "$userid",
			"date":             bson.M{"$max": "$date"},
			"all_rounds":       bson.M{"$sum": "$all_rounds"},
			"game_times":       bson.M{"$sum": "$game_times"},
			"bets":             bson.M{"$sum": "$bets"},
			"bet_rounds":       bson.M{"$sum": "$bet_rounds"},
			"win_rounds":       bson.M{"$sum": "$win_rounds"},
			"lose_rounds":      bson.M{"$sum": "$lose_rounds"},
			"tie_rounds":       bson.M{"$sum": "$tie_rounds"},
			"win_bets":         bson.M{"$sum": "$win_bets"},
			"lose_bets":        bson.M{"$sum": "$lose_bets"},
			"tie_bets":         bson.M{"$sum": "$tie_bets"},
			"wins":             bson.M{"$sum": "$wins"},
			"loses":            bson.M{"$sum": "$loses"},
			"cash":             bson.M{"$sum": "$cash"},
			"side_winner1":     bson.M{"$sum": "$side_winner1"},
			"side_winner2":     bson.M{"$sum": "$side_winner2"},
			"side_winner3":     bson.M{"$sum": "$side_winner3"},
			"side_winner4":     bson.M{"$sum": "$side_winner4"},
			"side_winner5":     bson.M{"$sum": "$side_winner5"},
			"side_winner6":     bson.M{"$sum": "$side_winner6"},
			"side_winner7":     bson.M{"$sum": "$side_winner7"},
			"side_winner8":     bson.M{"$sum": "$side_winner8"},
			"seat_bets1":       bson.M{"$sum": "$seat_bets1"},
			"seat_bets2":       bson.M{"$sum": "$seat_bets2"},
			"seat_bets3":       bson.M{"$sum": "$seat_bets3"},
			"player_win_seat1": bson.M{"$sum": "$player_win_seat1"},
			"player_win_seat2": bson.M{"$sum": "$player_win_seat2"},
			"player_win_seat3": bson.M{"$sum": "$player_win_seat3"},
			"player_win_seat4": bson.M{"$sum": "$player_win_seat4"},
			"player_win_seat5": bson.M{"$sum": "$player_win_seat5"},
			"player_win_seat6": bson.M{"$sum": "$player_win_seat6"},
			"player_win_seat7": bson.M{"$sum": "$player_win_seat7"},
			"player_win_seat8": bson.M{"$sum": "$player_win_seat8"},
			"player_multis":    bson.M{"$sum": "$player_multis"},
			"round_bet_avg":    bson.M{"$sum": "$round_bet_avg"},
			"money":            bson.M{"$sum": "$money"},
			"observe_rounds":   bson.M{"$sum": "$observe_rounds"},
		}},
		{"$project": bson.M{
			"_id":              "$_id",
			"all_rounds":       "$all_rounds",
			"game_times":       "$game_times",
			"bets":             "$bets",
			"bet_rounds":       "$bet_rounds",
			"win_rounds":       "$win_rounds",
			"lose_rounds":      "$lose_rounds",
			"tie_rounds":       "$tie_rounds",
			"win_bets":         "$win_bets",
			"lose_bets":        "$lose_bets",
			"tie_bets":         "$tie_bets",
			"wins":             "$wins",
			"loses":            "$loses",
			"cash":             "$cash",
			"side_winner1":     "$side_winner1",
			"side_winner2":     "$side_winner2",
			"side_winner3":     "$side_winner3",
			"side_winner4":     "$side_winner4",
			"side_winner5":     "$side_winner5",
			"side_winner6":     "$side_winner6",
			"side_winner7":     "$side_winner7",
			"side_winner8":     "$side_winner8",
			"seat_bets1":       "$seat_bets1",
			"seat_bets2":       "$seat_bets2",
			"seat_bets3":       "$seat_bets3",
			"player_win_seat1": "$player_win_seat1",
			"player_win_seat2": "$player_win_seat2",
			"player_win_seat3": "$player_win_seat3",
			"player_win_seat4": "$player_win_seat4",
			"player_win_seat5": "$player_win_seat5",
			"player_win_seat6": "$player_win_seat6",
			"player_win_seat7": "$player_win_seat7",
			"player_win_seat8": "$player_win_seat8",
			"player_multis":    "$player_multis",
			"money":            "$money",
			"observe_rounds":   "$observe_rounds",
			// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}}, // "can't $divide by zero"
			"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
		}},
		{"$match": m2},
		{"$count": "count"},
	}
	// 查总条数
	var rCount []bson.M
	err = RBPlayerStats.Pipe(m1).All(&rCount)
	if err != nil {
		return
	} else if len(rCount) > 0 {
		count = rCount[0]["count"].(int)
	}

	// 查列表数据
	m1 = m1[:len(m1)-1]
	m1 = append(m1, []bson.M{
		{"$sort": bson.M{"date": -1}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}...)
	err = RBPlayerStats.Pipe(m1).All(&stats)
	if err != nil {
		return
	}
	RBStatPlayersMapping(stats)

	// 查总汇数据
	mark := "--"
	summary := &entity.RBPlayerStat{}
	m1 = m1[:len(m1)-3]
	m1[1]["$group"].(bson.M)["_id"] = nil
	delete(m1[1]["$group"].(bson.M), "date")
	m1[1]["$group"].(bson.M)["userids"] = bson.M{"$addToSet": "$userid"}
	m1[2]["$project"].(bson.M)["userids"] = "$userids"
	var summarys []*entity.RBPlayerStat
	err = RBPlayerStats.Pipe(m1).All(&summarys)
	if err != nil {
		return
	} else if len(summarys) > 0 {
		summary = summarys[0]
		summary.No = fmt.Sprintf("%s-%s总汇", startDate, endDate)
		summary.Id = mark
		summary.FNickname = mark
		summary.FUserType = mark
		summary.FRegistArea = mark
		summary.FCtime = mark
		summary.FLoginTime = mark
		// summary.FLiveDays = mark
		// summary.FLoseDays = mark
		RBStatPlayersSummaryMapping(summary)
	}
	if len(summarys) > 0 {
		stats = append([]*entity.RBPlayerStat{summary}, stats...)
	}
	return
}

func RBStatPlayersMapping(stats []*entity.RBPlayerStat) {
	var playerIds []string
	// var playerStats = make(map[string]*entity.CrashPlayerStat, len(stats))
	for _, stat := range stats {
		playerIds = append(playerIds, stat.Id)
		// playerStats[stat.Id] = stat
	}
	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": playerIds}}},
		{"$project": bson.M{
			"_id":         "$_id",
			"nickname":    "$nickname",
			"money":       "$money",
			"cash_out":    "$cash_out",
			"diamond":     "$diamond",
			"ctime":       "$ctime",
			"login_time":  "$login_time",
			"regist_area": "$regist_area", //ab测试 0:A 1:B 2:C
			"state":       "$state",
		}},
	}
	var r2 []bson.M
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("cp users error:", err)
		return
	}
	var r2Map = make(map[string]bson.M, len(r2))
	for _, r := range r2 {
		userid := r["_id"].(string)
		r2Map[userid] = r
	}
	now := utils.BsonNow()
	for i, stat := range stats {
		r, ok := r2Map[stat.Id]
		if !ok {
			continue
		}
		// 玩家昵称	玩家类型 账号类型 注册时间 最后登陆时间 存活天数 流失天数
		// userid := r["_id"].(string)
		nickname := r["nickname"].(string)
		money := r["money"].(int)
		cash_out := r["cash_out"].(int)
		diamond := r["diamond"].(int64)
		ctime := r["ctime"].(time.Time)
		login_time := r["login_time"].(time.Time)
		regist_area := r["regist_area"].(int)
		state := r["state"].(int)

		// 将日期转成印度时间
		login_time, _ = ConvertToIndiaTime(login_time.Unix())
		ctime, _ = ConvertToIndiaTime(ctime.Unix())

		stat.No = strconv.Itoa(i + 1)
		stat.FNickname = nickname
		stat.Money = uint32(money)
		stat.FCtime = ctime.Format(utils.FORMAT)
		stat.FLoginTime = login_time.Format(utils.FORMAT)
		stat.FLiveDays = int(math.Ceil(login_time.Sub(ctime).Hours() / 24))
		stat.FLoseDays = int(math.Floor(now.Sub(login_time).Hours() / 24))

		// summary.FLiveDays += stat.FLiveDays
		// summary.FLoseDays += stat.FLoseDays
		// summary.AllRounds += stat.AllRounds
		// summary.GameTimes += stat.GameTimes
		// summary.Bets += stat.Bets
		// summary.BetRounds += stat.BetRounds
		// summary.WinRounds += stat.WinRounds
		// summary.LoseRounds += stat.LoseRounds
		// summary.TieRounds += stat.TieRounds
		// summary.WinBets += stat.WinBets
		// summary.LoseBets += stat.LoseBets
		// summary.TieBets += stat.TieBets
		// summary.Wins += stat.Wins
		// summary.Loses += stat.Loses
		// summary.Cash += stat.Cash
		// summary.MulpitleSum += stat.MulpitleSum
		// summary.WinEscapeMulpitleSum += stat.WinEscapeMulpitleSum
		// summary.WinMulpitleSum += stat.WinMulpitleSum
		// summary.LoseMulpitleSum += stat.LoseMulpitleSum
		// summary.Mulpitles = append(summary.Mulpitles, stat.Mulpitles...)
		// summary.WinEscapeMulpitles = append(summary.WinEscapeMulpitles, stat.WinEscapeMulpitles...)
		// summary.WinMulpitles = append(summary.WinMulpitles, stat.WinMulpitles...)
		// summary.LoseMulpitles = append(summary.LoseMulpitles, stat.LoseMulpitles...)
		// summary.ObserveRounds += stat.ObserveRounds
		// summary.SMoney += money
		// summary.SCashOut += cash_out
		// summary.SDiamond += diamond

		switch regist_area {
		case 0:
			stat.FRegistArea = "A类"
		case 1:
			stat.FRegistArea = "B类"
		case 2:
			stat.FRegistArea = "C类"
		}

		switch state {
		case 1:
			stat.FUserType = "新手"
		case 2:
			if money == 0 {
				stat.FUserType = "零充"
			} else if money >= 20000 && money <= 99999 {
				stat.FUserType = "普充"
			} else if money >= 100000 && money <= 499999 {
				stat.FUserType = "小R"
			} else if money >= 500000 && money <= 999999 {
				stat.FUserType = "中R"
			} else if money >= 1000000 && money <= 9999999 {
				stat.FUserType = "大R"
			} else if money >= 10000000 {
				stat.FUserType = "超大R"
			}
		case 3:
			stat.FUserType = "平民"
		case 4:
			stat.FUserType = "泡沫"
		}

		stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
		if stat.BetRounds > 0 {
			// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.BetRounds)/100.0)
			stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
		}
		stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
		stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
		stat.FTieBets = fmt.Sprintf("%.2f", float64(stat.TieBets)/100.0)
		if stat.BetRounds > 0 {
			stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
		}
		if stat.Bets > 0 {
			// stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
			stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
		}
		if stat.WinRounds > 0 {
			stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
		}
		if stat.TieRounds > 0 {
			stat.FTieBetsAvg = fmt.Sprintf("%.2f", float64(stat.TieBets)/float64(stat.TieRounds)/100.0)
		}
		stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
		stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
		stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)

		if stat.WinRounds > 0 {
			stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
		}
		// 下注开奖
		if stat.BetRounds > 0 {
			stat.SideWinnerRate1 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner1)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate2 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner2)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate3 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner3)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate4 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner4)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate5 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner5)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate6 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner6)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate7 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner7)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate8 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner8)/float64(stat.BetRounds)*100)
		}

		//玩家下注
		if stat.BetRounds > 0 {
			stat.SeatBetsRate1 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets1)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate2 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets2)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate3 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets3)/float64(stat.BetRounds)*100)
		}
		// 玩家下注中奖
		if stat.SeatBets1 > 0 {
			stat.PlayerWinSeatRate1 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat1)/float64(stat.SeatBets1)*100)
		}
		if stat.SeatBets2 > 0 {
			stat.PlayerWinSeatRate2 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat2)/float64(stat.SeatBets2)*100)
		}
		if stat.SeatBets3 > 0 {
			stat.PlayerWinSeatRate3 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat3)/float64(stat.SeatBets3)*100)
			stat.PlayerWinSeatRate4 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat4)/float64(stat.SeatBets3)*100)
			stat.PlayerWinSeatRate5 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat5)/float64(stat.SeatBets3)*100)
			stat.PlayerWinSeatRate6 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat6)/float64(stat.SeatBets3)*100)
			stat.PlayerWinSeatRate7 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat7)/float64(stat.SeatBets3)*100)
			stat.PlayerWinSeatRate8 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat8)/float64(stat.SeatBets3)*100)
		}
		// 多门率
		if stat.BetRounds > 0 {
			stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerMultis)/float64(stat.BetRounds)*100)
		}

		// 观察局
		if stat.BetRounds > 0 {
			stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
		}
		if stat.BetRounds > 0 {
			stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
		}
		stat.FMoney = fmt.Sprintf("%.2f", float64(money)/100)
		stat.FCashOut = fmt.Sprintf("%.2f", float64(cash_out)/100)
		stat.FWinMoney = fmt.Sprintf("%.2f", float64(int(diamond)+cash_out-money)/100)
	}
}

func RBStatPlayersSummaryMapping(stat *entity.RBPlayerStat) {
	// 查用户基本信息
	// m2 := []bson.M{
	// 	{"$match": bson.M{"_id": bson.M{"$in": stat.Userids}}},
	// 	{"$group": bson.M{
	// 		"_id":      nil,
	// 		"money":    bson.M{"$sum": "$money"},
	// 		"cash_out": bson.M{"$sum": "$cash_out"},
	// 		"diamond":  bson.M{"$sum": "$diamond"},
	// 	}},
	// }
	// var r2 []bson.M
	// err := PlayerUsers.Pipe(m2).All(&r2)
	// if err != nil {
	// 	beego.Error("lhd users error:", err)
	// 	return
	// } else if len(r2) > 0 {
	// 	money := r2[0]["money"].(int)
	// 	cash_out := r2[0]["cash_out"].(int)
	// 	diamond := r2[0]["diamond"].(int64)
	// 	stat.SMoney = money
	// 	stat.SCashOut = cash_out
	// 	stat.SDiamond = diamond
	// }

	stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
	if stat.BetRounds > 0 {
		stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.BetRounds)/100.0)
		// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
	}

	stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
	stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
	stat.FTieBets = fmt.Sprintf("%.2f", float64(stat.TieBets)/100.0)
	if stat.BetRounds > 0 {
		stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
	}
	if stat.Bets > 0 {
		stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
	}
	if stat.WinRounds > 0 {
		stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
	}
	if stat.TieRounds > 0 {
		stat.FTieBetsAvg = fmt.Sprintf("%.2f", float64(stat.TieBets)/float64(stat.TieRounds)/100.0)
	}
	stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
	stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
	stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)
	if stat.WinRounds > 0 {
		stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
	}
	// 下注开奖
	if stat.BetRounds > 0 {
		stat.SideWinnerRate1 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner1)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate2 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner2)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate3 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner3)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate4 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner4)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate5 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner5)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate6 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner6)/float64(stat.BetRounds)*100)
	}

	// 下注开奖
	if stat.BetRounds > 0 {
		stat.SideWinnerRate1 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner1)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate2 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner2)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate3 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner3)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate4 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner4)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate5 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner5)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate6 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner6)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate7 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner7)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate8 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner8)/float64(stat.BetRounds)*100)
	}

	//玩家下注
	if stat.BetRounds > 0 {
		stat.SeatBetsRate1 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets1)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate2 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets2)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate3 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets3)/float64(stat.BetRounds)*100)
	}
	// 玩家下注中奖
	if stat.SeatBets1 > 0 {
		stat.PlayerWinSeatRate1 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat1)/float64(stat.SeatBets1)*100)
	}
	if stat.SeatBets2 > 0 {
		stat.PlayerWinSeatRate2 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat2)/float64(stat.SeatBets2)*100)
	}
	if stat.SeatBets3 > 0 {
		stat.PlayerWinSeatRate3 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat3)/float64(stat.SeatBets3)*100)
		stat.PlayerWinSeatRate4 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat4)/float64(stat.SeatBets3)*100)
		stat.PlayerWinSeatRate5 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat5)/float64(stat.SeatBets3)*100)
		stat.PlayerWinSeatRate6 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat6)/float64(stat.SeatBets3)*100)
		stat.PlayerWinSeatRate7 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat7)/float64(stat.SeatBets3)*100)
		stat.PlayerWinSeatRate8 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat8)/float64(stat.SeatBets3)*100)
	}
	// 多门率
	if stat.BetRounds > 0 {
		stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerMultis)/float64(stat.BetRounds)*100)
	}

	// 观察局
	if stat.BetRounds > 0 {
		stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
	}
	if stat.BetRounds > 0 {
		stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
	}
	// stat.FMoney = fmt.Sprintf("%.2f", float64(stat.SMoney)/100)
	// stat.FCashOut = fmt.Sprintf("%.2f", float64(stat.SCashOut)/100)
	// stat.FWinMoney = fmt.Sprintf("%.2f", float64(int(stat.SDiamond)+stat.SCashOut-stat.SMoney)/100)
}

// GetABStatDates AB汇总明细
func (s *gameStatsService) GetRBStatDates(m, m2, m3 bson.M, startDate, endDate string) (stats []*entity.RBPlayerStat, err error) {
	// 用户列表筛选
	var filterDatePlayerIds []string
	if len(m2) > 0 || len(m3) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				// "_id":        bson.M{"userid":"$userid", "date", "$date"},
				"_id":        "$_id",
				"bets":       bson.M{"$sum": "$bets"},
				"bet_rounds": bson.M{"$sum": "$bet_rounds"},
				"cash":       bson.M{"$sum": "$cash"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				"cash": "$cash",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = RBPlayerStats.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetRBStatDates error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			id := r["_id"].(string)
			filterDatePlayerIds = append(filterDatePlayerIds, id)
			ids := strings.Split(id, "-") // id=date-userid
			if len(ids) >= 2 {
				userids = append(userids, ids[1])
			}
		}
		if len(filterDatePlayerIds) == 0 { // 未匹配到
			return
		}
		if len(m3) > 0 { // 流失天数，局均打码量
			if len(userids) == 0 { // 未匹配到
				return
			}
			now := utils.BsonNow()
			m3_2 := []bson.M{
				{"$match": bson.M{"_id": bson.M{"$in": userids}}},
				{"$project": bson.M{
					"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
					"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
				}},
				{"$match": m3},
			}
			var r3_2 []bson.M
			err = PlayerUsers.Pipe(m3_2).All(&r3_2)
			if err != nil {
				beego.Error("GetRBStatDates error32:", err)
			} else if len(r3_2) == 0 { // 未匹配到
				return
			}
			var filterId2 []string
			for _, id := range filterDatePlayerIds {
				for _, r := range r3_2 {
					userid := r["_id"].(string)
					if strings.HasSuffix(id, "-"+userid) {
						filterId2 = append(filterId2, id)
						break
					}
				}
			}
			if len(filterId2) == 0 { // 未匹配到
				return
			}
			filterDatePlayerIds = filterId2
		}
	}
	if len(filterDatePlayerIds) > 0 {
		m["_id"] = bson.M{"$in": filterDatePlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":              "$date_str",
			"date":             bson.M{"$min": "$date"},
			"players":          bson.M{"$sum": 1},
			"new_players":      bson.M{"$sum": bson.M{"$cond": bson.M{"if": "$new_reg", "then": 1, "else": 0}}},
			"old_players":      bson.M{"$sum": bson.M{"$cond": bson.M{"if": "$new_reg", "then": 0, "else": 1}}},
			"all_rounds":       bson.M{"$sum": "$all_rounds"},
			"game_times":       bson.M{"$sum": "$game_times"},
			"bets":             bson.M{"$sum": "$bets"},
			"bet_rounds":       bson.M{"$sum": "$bet_rounds"},
			"win_rounds":       bson.M{"$sum": "$win_rounds"},
			"lose_rounds":      bson.M{"$sum": "$lose_rounds"},
			"tie_rounds":       bson.M{"$sum": "$tie_rounds"},
			"win_bets":         bson.M{"$sum": "$win_bets"},
			"lose_bets":        bson.M{"$sum": "$lose_bets"},
			"tie_bets":         bson.M{"$sum": "$tie_bets"},
			"wins":             bson.M{"$sum": "$wins"},
			"loses":            bson.M{"$sum": "$loses"},
			"cash":             bson.M{"$sum": "$cash"},
			"side_winner1":     bson.M{"$sum": "$side_winner1"},
			"side_winner2":     bson.M{"$sum": "$side_winner2"},
			"side_winner3":     bson.M{"$sum": "$side_winner3"},
			"side_winner4":     bson.M{"$sum": "$side_winner4"},
			"side_winner5":     bson.M{"$sum": "$side_winner5"},
			"side_winner6":     bson.M{"$sum": "$side_winner6"},
			"side_winner7":     bson.M{"$sum": "$side_winner7"},
			"side_winner8":     bson.M{"$sum": "$side_winner8"},
			"seat_bets1":       bson.M{"$sum": "$seat_bets1"},
			"seat_bets2":       bson.M{"$sum": "$seat_bets2"},
			"seat_bets3":       bson.M{"$sum": "$seat_bets3"},
			"player_win_seat1": bson.M{"$sum": "$player_win_seat1"},
			"player_win_seat2": bson.M{"$sum": "$player_win_seat2"},
			"player_win_seat3": bson.M{"$sum": "$player_win_seat3"},
			"player_win_seat4": bson.M{"$sum": "$player_win_seat4"},
			"player_win_seat5": bson.M{"$sum": "$player_win_seat5"},
			"player_win_seat6": bson.M{"$sum": "$player_win_seat6"},
			"player_win_seat7": bson.M{"$sum": "$player_win_seat7"},
			"player_win_seat8": bson.M{"$sum": "$player_win_seat8"},
			"player_multis":    bson.M{"$sum": "$player_multis"},
			"round_bet_avg":    bson.M{"$sum": "$round_bet_avg"},
			"money":            bson.M{"$sum": "$money"},
			"observe_rounds":   bson.M{"$sum": "$observe_rounds"},
			"all_rounds_list":  bson.M{"$push": "$bet_rounds"},
		}},
		{"$sort": bson.M{"date": -1}},
	}
	err = RBPlayerStats.Pipe(m1).All(&stats)
	if err != nil {
		return
	}

	// 查询日活数据
	var dateLoginedUsers = make(map[int64]int)
	m9 := []bson.M{
		{"$match": bson.M{"date": m["date"]}},
		{"$project": bson.M{"date": "$date", "logined_users": "$logined_users"}},
	}
	var r9 []bson.M
	err = PddStats.Pipe(m9).All(&r9)
	if err != nil {
		beego.Error("GetRBStatDates logined_users error: ", err)
	} else {
		for _, r := range r9 {
			date := r["date"].(int64)
			logined_users := r["logined_users"].(int)
			dateLoginedUsers[date] = logined_users
		}
	}

	mark := "--"
	summary := &entity.RBPlayerStat{
		DateStr: fmt.Sprintf("%s-%s总汇", startDate, endDate),
		Id:      mark,
	}
	RBStatDatesMapping(summary, stats, dateLoginedUsers)
	RBStatDatesSummaryMapping(summary)
	stats = append([]*entity.RBPlayerStat{summary}, stats...)
	return
}

func RBStatDatesMapping(summary *entity.RBPlayerStat, stats []*entity.RBPlayerStat, dateLoginedUsers map[int64]int) {
	for _, stat := range stats {
		dateStr := stat.Id
		date, _ := utils.Unix(dateStr)
		stat.Date = date
		stat.DateStr = dateStr
		loginedUsers := dateLoginedUsers[stat.Date] // 日活

		stat.DateStr = utils.Stamp2Time(stat.Date).Format("2006-01-02")
		stat.FLoginedUsers = loginedUsers

		summary.FLoginedUsers += stat.FLoginedUsers
		summary.FLiveDays += stat.FLiveDays
		summary.FLoseDays += stat.FLoseDays
		summary.AllRounds += stat.AllRounds
		summary.GameTimes += stat.GameTimes
		summary.Bets += stat.Bets
		summary.BetRounds += stat.BetRounds
		summary.WinRounds += stat.WinRounds
		summary.LoseRounds += stat.LoseRounds
		summary.TieRounds += stat.TieRounds
		summary.WinBets += stat.WinBets
		summary.LoseBets += stat.LoseBets
		summary.TieBets += stat.TieBets
		summary.Wins += stat.Wins
		summary.Loses += stat.Loses
		summary.Cash += stat.Cash
		summary.SideWinner1 += stat.SideWinner1
		summary.SideWinner2 += stat.SideWinner2
		summary.SideWinner3 += stat.SideWinner3
		summary.SideWinner4 += stat.SideWinner4
		summary.SideWinner5 += stat.SideWinner5
		summary.SideWinner6 += stat.SideWinner6
		summary.SideWinner7 += stat.SideWinner7
		summary.SideWinner8 += stat.SideWinner8
		summary.SeatBets1 += stat.SeatBets1
		summary.SeatBets2 += stat.SeatBets2
		summary.SeatBets3 += stat.SeatBets3
		summary.PlayerWinSeat1 += stat.PlayerWinSeat1
		summary.PlayerWinSeat2 += stat.PlayerWinSeat2
		summary.PlayerWinSeat3 += stat.PlayerWinSeat3
		summary.PlayerWinSeat4 += stat.PlayerWinSeat4
		summary.PlayerWinSeat5 += stat.PlayerWinSeat5
		summary.PlayerWinSeat6 += stat.PlayerWinSeat6
		summary.PlayerWinSeat7 += stat.PlayerWinSeat7
		summary.PlayerWinSeat8 += stat.PlayerWinSeat8
		summary.PlayerMultis += stat.PlayerMultis
		summary.RoundBetAvg += stat.RoundBetAvg

		summary.ObserveRounds += stat.ObserveRounds
		summary.Players += stat.Players
		summary.NewPlayers += stat.NewPlayers
		summary.OldPlayers += stat.OldPlayers
		summary.AllRoundsList = append(summary.AllRoundsList, stat.AllRoundsList...)

		// 玩 crash 人数占日活比
		if stat.FLoginedUsers > 0 {
			stat.FPlayerRate = fmt.Sprintf("%.2f", float64(stat.Players)/float64(stat.FLoginedUsers))
		}

		// 查询当天注册的新用户
		startTime := utils.Str2Time(fmt.Sprintf("%s 00:00:00", dateStr), Location())
		endTime := utils.Str2Time(fmt.Sprintf("%s 23:59:59", dateStr), Location())
		m7 := []bson.M{
			{"$match": bson.M{
				"ctime":            bson.M{"$gte": startTime, "$lte": endTime},
				"robot":            false,
				"simulation_robot": false,
			}},
			{"$group": bson.M{
				"_id":   nil,
				"total": bson.M{"$sum": 1},
			}},
		}
		var r7 []bson.M
		err := PlayerUsers.Pipe(m7).All(&r7)
		if err != nil {
			beego.Error("查询当天注册的新用户 error: ", err)
		} else if len(r7) > 0 {
			stat.FLoginedNewUsers = r7[0]["total"].(int)
		}
		stat.FLoginedOldUsers = stat.FLoginedUsers - stat.FLoginedNewUsers
		if stat.FLoginedOldUsers < 0 {
			stat.FLoginedOldUsers = 0
		}
		summary.FLoginedNewUsers += stat.FLoginedNewUsers
		summary.FLoginedOldUsers += stat.FLoginedOldUsers
		// FNewPlayerRate, FOldPlayerRate 占新老玩家日活
		if stat.FLoginedNewUsers > 0 {
			stat.FNewPlayerRate = fmt.Sprintf("%.2f", float64(stat.NewPlayers)/float64(stat.FLoginedNewUsers))
		}
		if stat.FLoginedOldUsers > 0 {
			stat.FOldPlayerRate = fmt.Sprintf("%.2f", float64(stat.OldPlayers)/float64(stat.FLoginedOldUsers))
		}

		if stat.Players > 0 {
			stat.FPlayerRoundsAvg = fmt.Sprintf("%.2f", float64(stat.BetRounds)/float64(stat.Players))
		}
		// 局数中位数
		if len(stat.AllRoundsList) > 0 {
			var rounds = stat.AllRoundsList
			sort.Slice(rounds, func(i, j int) bool {
				return rounds[i] < rounds[j]
			})
			length := len(rounds)
			if length > 0 {
				if length%2 == 1 {
					stat.FPlayerRoundsMedian = rounds[length/2]
				} else {
					// 中间两位算平均
					stat.FPlayerRoundsMedian = (rounds[length/2] + rounds[length/2-1]) / 2
				}
			}
		}
		// stat.FGameTimes = fmt.Sprintf("%.2f", float64(stat.GameTimes)/60)
		if stat.BetRounds > 0 {
			stat.FGameTimesAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.BetRounds)/60)
		}
		if stat.Players > 0 {
			stat.FGameTimesPlayerAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.Players)/60)
		}

		stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
		if stat.Players > 0 {
			stat.FPlayerBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.Players)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FRoundBetsAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.Bets, 100)/float64(stat.BetRounds))
			// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
		}
		stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
		stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
		stat.FTieBets = fmt.Sprintf("%.2f", float64(stat.TieBets)/100.0)
		if stat.BetRounds > 0 {
			stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
		}
		if stat.Bets > 0 {
			stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
		}
		if stat.WinRounds > 0 {
			stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
		}
		if stat.TieRounds > 0 {
			stat.FTieBetsAvg = fmt.Sprintf("%.2f", float64(stat.TieBets)/float64(stat.TieRounds)/100.0)
		}
		stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
		stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
		stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)

		if stat.WinRounds > 0 {
			stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
		}
		if stat.LoseRounds > 0 {
			stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
		}
		if stat.BetRounds > 0 {
			stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
		}
		// 下注开奖
		if stat.BetRounds > 0 {
			stat.SideWinnerRate1 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner1)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate2 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner2)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate3 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner3)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate4 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner4)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate5 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner5)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate6 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner6)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate7 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner7)/float64(stat.BetRounds)*100)
			stat.SideWinnerRate8 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner8)/float64(stat.BetRounds)*100)
		}

		//玩家下注
		if stat.BetRounds > 0 {
			stat.SeatBetsRate1 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets1)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate2 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets2)/float64(stat.BetRounds)*100)
			stat.SeatBetsRate3 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets3)/float64(stat.BetRounds)*100)
		}
		// 玩家下注中奖
		if stat.SeatBets1 > 0 {
			stat.PlayerWinSeatRate1 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat1)/float64(stat.SeatBets1)*100)
		}
		if stat.SeatBets2 > 0 {
			stat.PlayerWinSeatRate2 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat2)/float64(stat.SeatBets2)*100)
		}
		if stat.SeatBets3 > 0 {
			stat.PlayerWinSeatRate3 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat3)/float64(stat.SeatBets3)*100)
			stat.PlayerWinSeatRate4 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat4)/float64(stat.SeatBets3)*100)
			stat.PlayerWinSeatRate5 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat5)/float64(stat.SeatBets3)*100)
			stat.PlayerWinSeatRate6 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat6)/float64(stat.SeatBets3)*100)
			stat.PlayerWinSeatRate7 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat7)/float64(stat.SeatBets3)*100)
			stat.PlayerWinSeatRate8 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat8)/float64(stat.SeatBets3)*100)
		}
		// 多门率
		if stat.BetRounds > 0 {
			stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerMultis)/float64(stat.BetRounds)*100)
		}
		// 观察局
		if stat.BetRounds > 0 {
			stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
		}
		if stat.BetRounds > 0 {
			stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
		}

	}
}

func RBStatDatesSummaryMapping(stat *entity.RBPlayerStat) {
	// 玩 lhd 人数占日活比
	if stat.FLoginedUsers > 0 {
		stat.FPlayerRate = fmt.Sprintf("%.2f", float64(stat.Players)/float64(stat.FLoginedUsers))
	}

	// FNewPlayerRate, FOldPlayerRate 占新老玩家日活
	if stat.FLoginedNewUsers > 0 {
		stat.FNewPlayerRate = fmt.Sprintf("%.2f", float64(stat.NewPlayers)/float64(stat.FLoginedNewUsers))
	}
	if stat.FLoginedOldUsers > 0 {
		stat.FOldPlayerRate = fmt.Sprintf("%.2f", float64(stat.OldPlayers)/float64(stat.FLoginedOldUsers))
	}

	if stat.Players > 0 {
		stat.FPlayerRoundsAvg = fmt.Sprintf("%.2f", float64(stat.BetRounds)/float64(stat.Players))
	}
	// 局数中位数
	if len(stat.AllRoundsList) > 0 {
		var rounds = stat.AllRoundsList
		sort.Slice(rounds, func(i, j int) bool {
			return rounds[i] < rounds[j]
		})
		length := len(rounds)
		if length > 0 {
			if length%2 == 1 {
				stat.FPlayerRoundsMedian = rounds[length/2]
			} else {
				// 中间两位算平均
				stat.FPlayerRoundsMedian = (rounds[length/2] + rounds[length/2-1]) / 2
			}
		}
	}
	// stat.FGameTimes = fmt.Sprintf("%.2f", float64(stat.GameTimes)/60)
	if stat.BetRounds > 0 {
		stat.FGameTimesAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.BetRounds)/60)
	}
	if stat.Players > 0 {
		stat.FGameTimesPlayerAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.Players)/60)
	}

	if stat.BetRounds > 0 {
		stat.FGameTimesAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.BetRounds)/60)
	}
	if stat.Players > 0 {
		stat.FGameTimesPlayerAvg = fmt.Sprintf("%.2f", float64(stat.GameTimes)/float64(stat.Players)/60)
	}

	stat.FBets = fmt.Sprintf("%.2f", float64(stat.Bets)/100.0)
	if stat.Players > 0 {
		stat.FPlayerBetsAvg = fmt.Sprintf("%.2f", float64(stat.Bets)/float64(stat.Players)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FRoundBetsAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.Bets, 100)/float64(stat.BetRounds))
		// stat.FRoundBetsAvg = fmt.Sprintf("%.2f", float64(stat.RoundBetAvg)/100.0)
	}
	stat.FWinBets = fmt.Sprintf("%.2f", float64(stat.WinBets)/100.0)
	stat.FLoseBets = fmt.Sprintf("%.2f", float64(stat.LoseBets)/100.0)
	stat.FTieBets = fmt.Sprintf("%.2f", float64(stat.TieBets)/100.0)

	if stat.BetRounds > 0 {
		stat.FWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
	}
	if stat.Bets > 0 {
		stat.FRebateRate = fmt.Sprintf("%.2f%%", float64(stat.Wins+stat.Loses+stat.Bets)/float64(stat.Bets)*100)
	}
	if stat.WinRounds > 0 {
		stat.FWinBetsAvg = fmt.Sprintf("%.2f", float64(stat.WinBets)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLoseBetsAvg = fmt.Sprintf("%.2f", float64(stat.LoseBets)/float64(stat.LoseRounds)/100.0)
	}
	stat.FWins = fmt.Sprintf("%.2f", float64(stat.Wins)/100.0)
	stat.FLoses = fmt.Sprintf("%.2f", float64(stat.Loses)/100.0)
	stat.FCash = fmt.Sprintf("%.2f", float64(stat.Cash)/100.0)

	if stat.WinRounds > 0 {
		stat.FWinsAvg = fmt.Sprintf("%.2f", float64(stat.Wins)/float64(stat.WinRounds)/100.0)
	}
	if stat.LoseRounds > 0 {
		stat.FLosesAvg = fmt.Sprintf("%.2f", float64(-stat.Loses)/float64(stat.LoseRounds)/100.0)
	}
	if stat.TieRounds > 0 {
		stat.FTieBetsAvg = fmt.Sprintf("%.2f", float64(stat.TieBets)/float64(stat.TieRounds)/100.0)
	}
	if stat.BetRounds > 0 {
		stat.FCashAvg = fmt.Sprintf("%.2f", float64(stat.Cash)/float64(stat.BetRounds)/100.0)
	}
	// 下注开奖
	if stat.BetRounds > 0 {
		stat.SideWinnerRate1 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner1)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate2 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner2)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate3 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner3)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate4 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner4)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate5 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner5)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate6 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner6)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate7 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner7)/float64(stat.BetRounds)*100)
		stat.SideWinnerRate8 = fmt.Sprintf("%.2f%%", float64(stat.SideWinner8)/float64(stat.BetRounds)*100)
	}

	//玩家下注
	if stat.BetRounds > 0 {
		stat.SeatBetsRate1 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets1)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate2 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets2)/float64(stat.BetRounds)*100)
		stat.SeatBetsRate3 = fmt.Sprintf("%.2f%%", float64(stat.SeatBets3)/float64(stat.BetRounds)*100)
	}
	// 玩家下注中奖
	if stat.SeatBets1 > 0 {
		stat.PlayerWinSeatRate1 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat1)/float64(stat.SeatBets1)*100)
	}
	if stat.SeatBets2 > 0 {
		stat.PlayerWinSeatRate2 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat2)/float64(stat.SeatBets2)*100)
	}
	if stat.SeatBets3 > 0 {
		stat.PlayerWinSeatRate3 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat3)/float64(stat.SeatBets3)*100)
		stat.PlayerWinSeatRate4 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat4)/float64(stat.SeatBets3)*100)
		stat.PlayerWinSeatRate5 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat5)/float64(stat.SeatBets3)*100)
		stat.PlayerWinSeatRate6 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat6)/float64(stat.SeatBets3)*100)
		stat.PlayerWinSeatRate7 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat7)/float64(stat.SeatBets3)*100)
		stat.PlayerWinSeatRate8 = fmt.Sprintf("%.2f%%", float64(stat.PlayerWinSeat8)/float64(stat.SeatBets3)*100)
	}
	// 多门率
	if stat.BetRounds > 0 {
		stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.PlayerMultis)/float64(stat.BetRounds)*100)
	}

	// 观察局
	if stat.BetRounds > 0 {
		stat.ObserveRoundsAvg = fmt.Sprintf("%.2f", float64(stat.ObserveRounds)/float64(stat.BetRounds))
	}
	if stat.BetRounds > 0 {
		stat.ObserveRoundsRate = fmt.Sprintf("%.2f%%", float64(stat.ObserveRounds)/float64(stat.ObserveRounds+stat.BetRounds)*100)
	}
}

// RB 红运当头
func (s *gameStatsService) GetRBHydtStat(page, pageSize int, m, m2, m3 bson.M, startDate, endDate string) (
	stats []*entity.RBHydtStat, count int, err error,
) {
	// GetRBHydtStat
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)

	// 用户列表筛选
	var filterPlayerIds []string
	if len(m2) > 0 || len(m3) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				"_id":        "$userid",
				"bets":       bson.M{"$sum": "$strategy_bets"},
				"bet_rounds": bson.M{"$sum": "$strategy_times"},
				"cash":       bson.M{"$sum": "$cash"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				"cash": "$cash",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = RBHydtStats.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetRBHydtStat error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			userid := r["_id"].(string)
			userids = append(userids, userid)
		}
		if len(userids) == 0 { // 未匹配到
			return
		}
		now := utils.BsonNow()
		m3_2 := []bson.M{
			{"$match": bson.M{"_id": bson.M{"$in": userids}}},
			{"$project": bson.M{
				"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
				"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
			}},
			{"$match": m3},
		}
		var r3_2 []bson.M
		err = PlayerUsers.Pipe(m3_2).All(&r3_2)
		if err != nil {
			beego.Error("RB PlayerUsers error32:", err)
		} else if len(r3_2) == 0 { // 未匹配到
			return
		}
		for _, r := range r3_2 {
			userid := r["_id"].(string)
			filterPlayerIds = append(filterPlayerIds, userid)
		}
	}
	if len(filterPlayerIds) > 0 {
		m["userid"] = bson.M{"$in": filterPlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":  "$userid",
			"date": bson.M{"$max": "$date"},
			// "date_str":                bson.M{"$sum": "$date_str"},
			"strategy_players":      bson.M{"$max": "$strategy_players"},
			"strategy_times":        bson.M{"$max": "$strategy_times"},
			"tigger_times":          bson.M{"$max": "$tigger_times"},
			"all_evo_times":         bson.M{"$sum": "$all_evo_times"},
			"strategy_multi_rounds": bson.M{"$sum": "$strategy_multi_rounds"},
			"strategy_bets":         bson.M{"$sum": "$strategy_bets"},
			"strategy_win_rounds":   bson.M{"$sum": "$strategy_win_rounds"},
			"strategy_wins":         bson.M{"$sum": "$strategy_wins"},
			"strategy_lose_rounds":  bson.M{"$sum": "$strategy_lose_rounds"},
			"strategy_loses":        bson.M{"$sum": "$strategy_loses"},
		}},
		{"$project": bson.M{
			"_id":                   "$_id",
			"date":                  "$date",
			"userid":                "$userid",
			"strategy_players":      "$strategy_players",
			"strategy_times":        "$strategy_times",
			"tigger_times":          "$tigger_times",
			"all_evo_times":         "$all_evo_times",
			"strategy_multi_rounds": "$strategy_multi_rounds",
			"strategy_bets":         "$strategy_bets",
			"strategy_win_rounds":   "$strategy_win_rounds",
			"strategy_wins":         "$strategy_wins",
			"strategy_loses":        "$strategy_loses",
			"strategy_lose_rounds":  "$strategy_lose_rounds",
		}},
		{"$match": m2},
		{"$count": "count"},
	}
	// 查总条数
	var rCount []bson.M
	err = RBHydtStats.Pipe(m1).All(&rCount)
	if err != nil {
		return
	} else if len(rCount) > 0 {
		count = rCount[0]["count"].(int)
	}

	// 查列表数据
	m1 = m1[:len(m1)-1]
	m1 = append(m1, []bson.M{
		{"$sort": bson.M{"date": -1}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}...)
	err = RBHydtStats.Pipe(m1).All(&stats)
	if err != nil {
		return
	}
	RBHydtStatMapping(stats)

	// 查总汇数据
	mark := "--"
	summary := &entity.RBHydtStat{}
	m1 = m1[:len(m1)-3]
	m1[1]["$group"].(bson.M)["_id"] = nil
	delete(m1[1]["$group"].(bson.M), "date")
	m1[1]["$group"].(bson.M)["userids"] = bson.M{"$addToSet": "$userid"}
	m1[1]["$group"].(bson.M)["strategy_players"] = bson.M{"$sum": "$strategy_players"} // 触发策略人数总汇
	m1[1]["$group"].(bson.M)["strategy_times"] = bson.M{"$sum": "$strategy_times"}
	m1[1]["$group"].(bson.M)["tigger_times"] = bson.M{"$sum": "$tigger_times"}
	m1[2]["$project"].(bson.M)["userids"] = "$userids"
	str, err := json.Marshal(m1)
	fmt.Println(err, string(str))
	var summarys []*entity.RBHydtStat
	err = RBHydtStats.Pipe(m1).All(&summarys)
	if err != nil {
		beego.Error("RB summary error:", err)
		return
	} else if len(summarys) > 0 {
		summary = summarys[0]
		summary.No = fmt.Sprintf("%s-%s总汇", startDate, endDate)
		summary.Id = mark
		summary.FNickname = mark
		summary.FUserType = mark
		summary.FRegistArea = mark
		summary.FCtime = mark
		summary.FLoginTime = mark
		RBHydtStatSummaryMapping(summary)
	}
	if len(summarys) > 0 {
		stats = append([]*entity.RBHydtStat{summary}, stats...)
	}
	return
}

func RBHydtStatMapping(stats []*entity.RBHydtStat) {
	var playerIds []string
	// var playerStats = make(map[string]*entity.CrashPlayerStat, len(stats))
	for _, stat := range stats {
		playerIds = append(playerIds, stat.Id)
		// playerStats[stat.Id] = stat
	}
	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": playerIds}}},
		{"$project": bson.M{
			"_id":         "$_id",
			"nickname":    "$nickname",
			"money":       "$money",
			"cash_out":    "$cash_out",
			"diamond":     "$diamond",
			"ctime":       "$ctime",
			"login_time":  "$login_time",
			"regist_area": "$regist_area", //ab测试 0:A 1:B 2:C
			"state":       "$state",
			"cp_strategy": "$cp_strategy",
		}},
	}
	var r2 []*entity.PlayerUser
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("cp users error:", err)
		return
	}

	var r2Map = make(map[string]*entity.PlayerUser, len(r2))
	for _, r := range r2 {
		userid := r.Userid
		r2Map[userid] = r
	}
	now := utils.BsonNow()
	for i, stat := range stats {
		r, ok := r2Map[stat.Id]
		if !ok {
			continue
		}
		// 玩家昵称	玩家类型 账号类型 注册时间 最后登陆时间 存活天数 流失天数
		nickname := r.Nickname
		money := r.Money
		ctime := r.Ctime
		login_time := r.LoginTime
		regist_area := r.RegistArea
		state := r.State

		// 将日期转成印度时间
		login_time, _ = ConvertToIndiaTime(login_time.Unix())
		ctime, _ = ConvertToIndiaTime(ctime.Unix())

		stat.No = strconv.Itoa(i + 1)
		stat.FNickname = nickname
		stat.FCtime = ctime.Format(utils.FORMAT)
		stat.FLoginTime = login_time.Format(utils.FORMAT)
		stat.FLiveDays = int(math.Ceil(login_time.Sub(ctime).Hours() / 24))
		stat.FLoseDays = int(math.Floor(now.Sub(login_time).Hours() / 24))

		switch regist_area {
		case 0:
			stat.FRegistArea = "A类"
		case 1:
			stat.FRegistArea = "B类"
		case 2:
			stat.FRegistArea = "C类"
		}

		switch state {
		case 1:
			stat.FUserType = "新手"
		case 2:
			if money == 0 {
				stat.FUserType = "零充"
			} else if money >= 20000 && money <= 99999 {
				stat.FUserType = "普充"
			} else if money >= 100000 && money <= 499999 {
				stat.FUserType = "小R"
			} else if money >= 500000 && money <= 999999 {
				stat.FUserType = "中R"
			} else if money >= 1000000 && money <= 9999999 {
				stat.FUserType = "大R"
			} else if money >= 10000000 {
				stat.FUserType = "超大R"
			}
		case 3:
			stat.FUserType = "平民"
		case 4:
			stat.FUserType = "泡沫"
		}
		if stat.StrategyTimes > 0 {
			stat.TriggerRate = fmt.Sprintf("%.2f%%", float64(stat.TiggerTimes)/float64(stat.StrategyTimes)*100.0)

		}
		if stat.StrategyPlayers > 0 {
			stat.PerCapitaAvg = fmt.Sprintf("%.2f", float64(stat.TiggerTimes)/float64(stat.StrategyPlayers))
		}
		if stat.StrategyTimes > 0 {
			stat.TiggerStrategyAvg = fmt.Sprintf("%.2f", float64(stat.AllEvoTimes)/float64(stat.StrategyTimes))

		}
		if stat.TiggerTimes > 0 {
			stat.YxTiggerStrategyAvg = fmt.Sprintf("%.2f", float64(stat.AllEvoTimes)/float64(stat.TiggerTimes))
		}
		// 多门率
		if stat.AllEvoTimes > 0 {
			stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyMultiRounds)/float64(stat.AllEvoTimes)*100.0)
		}
		stat.FStrategyBets = fmt.Sprintf("%.2f", float64(stat.StrategyBets)/100.0)
		if stat.StrategyTimes > 0 {
			stat.FStrategyBetRoundsAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.StrategyBets, 100)/float64(stat.AllEvoTimes))
		}
		stat.FStrategyWins = fmt.Sprintf("%.2f", float64(stat.StrategyWins)/100.0)
		stat.FStrategyLoses = fmt.Sprintf("%.2f", float64(stat.StrategyLoses)/100.0)
		stat.FStrategyCash = fmt.Sprintf("%.2f", float64(stat.StrategyWins+stat.StrategyLoses)/100.0)

		if stat.AllEvoTimes > 0 {
			stat.FStrategyWinRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyWinRounds)/float64(stat.AllEvoTimes)*100.0)
		}
		if -stat.StrategyLoses > 0 {
			stat.FStrategyRebateRate = fmt.Sprintf("%.2f%%", (float64(stat.StrategyWins)/float64(-stat.StrategyLoses))*100.0)
		} else {
			stat.FStrategyRebateRate = "未输过"
		}
	}
}

func RBHydtStatSummaryMapping(stat *entity.RBHydtStat) {
	stat.StrategyPlayers = int32(len(stat.Userids))
	if stat.StrategyTimes > 0 {
		stat.TriggerRate = fmt.Sprintf("%.2f%%", float64(stat.TiggerTimes)/float64(stat.StrategyTimes)*100.0)

	}
	if stat.StrategyPlayers > 0 {
		stat.PerCapitaAvg = fmt.Sprintf("%.2f", float64(stat.TiggerTimes)/float64(stat.StrategyPlayers))
	}
	if stat.StrategyTimes > 0 {
		stat.TiggerStrategyAvg = fmt.Sprintf("%.2f", float64(stat.AllEvoTimes)/float64(stat.StrategyTimes))
	}
	if stat.TiggerTimes > 0 {
		stat.YxTiggerStrategyAvg = fmt.Sprintf("%.2f", float64(stat.AllEvoTimes)/float64(stat.TiggerTimes))
	}
	// 多门率
	if stat.AllEvoTimes > 0 {
		stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyMultiRounds)/float64(stat.AllEvoTimes)*100.0)
	}
	stat.FStrategyBets = fmt.Sprintf("%.2f", float64(stat.StrategyBets)/100.0)
	if stat.StrategyTimes > 0 {
		stat.FStrategyBetRoundsAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.StrategyBets, 100)/float64(stat.AllEvoTimes))
	}
	stat.FStrategyWins = fmt.Sprintf("%.2f", float64(stat.StrategyWins)/100.0)
	stat.FStrategyLoses = fmt.Sprintf("%.2f", float64(stat.StrategyLoses)/100.0)
	stat.FStrategyCash = fmt.Sprintf("%.2f", float64(stat.StrategyWins+stat.StrategyLoses)/100.0)

	if stat.AllEvoTimes > 0 {
		stat.FStrategyWinRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyWinRounds)/float64(stat.AllEvoTimes)*100.0)
	}
	if -stat.StrategyLoses > 0 {
		// stat.FStrategyRebateRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.StrategyWins, 100)/ComputeFloat(stat.StrategyLoses, 100))*100.0)
		stat.FStrategyRebateRate = fmt.Sprintf("%.2f%%", (float64(stat.StrategyWins)/float64(-stat.StrategyLoses))*100.0)
	} else {
		stat.FStrategyRebateRate = "未输过"
	}
}

// RB 绝处逢生
func (s *gameStatsService) GetRBJcfsStat(page, pageSize int, m, m2, m3 bson.M, startDate, endDate string) (
	stats []*entity.RBJcfsStat, count int, err error,
) {
	// UpDownXxscStat
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)

	// 用户列表筛选
	var filterPlayerIds []string
	if len(m2) > 0 || len(m3) > 0 {
		m3_1 := []bson.M{
			{"$match": m},
			{"$group": bson.M{
				"_id":        "$userid",
				"bets":       bson.M{"$sum": "$strategy_bets"},
				"bet_rounds": bson.M{"$sum": "$strategy_rounds"},
				"cash":       bson.M{"$sum": "$cash"},
			}},
			{"$project": bson.M{
				"_id":  "$_id",
				"bets": "$bets",
				"cash": "$cash",
				// "round_bet_avg": bson.M{"$divide": []string{"$bets", "$bet_rounds"}},
				"round_bet_avg": bson.M{"$cond": []any{bson.M{"$eq": []any{"$bet_rounds", 0}}, 0, bson.M{"$divide": []string{"$bets", "$bet_rounds"}}}},
			}},
			{"$match": m2},
		}
		var r3_1 []bson.M
		err = RBJcfsStats.Pipe(m3_1).All(&r3_1)
		if err != nil {
			beego.Error("GetRBJcfsStat error31:", err)
		}
		var userids []string
		for _, r := range r3_1 {
			userid := r["_id"].(string)
			userids = append(userids, userid)
		}
		if len(userids) == 0 { // 未匹配到
			return
		}
		now := utils.BsonNow()
		m3_2 := []bson.M{
			{"$match": bson.M{"_id": bson.M{"$in": userids}}},
			{"$project": bson.M{
				"_id": "$_id", "money": "$money", "state": "$state", "custom_types": "$custom_types",
				"loss_days": bson.M{"$floor": bson.M{"$divide": []any{bson.M{"$subtract": []any{now, "$login_time"}}, 1000 * 60 * 60 * 24}}},
			}},
			{"$match": m3},
		}
		var r3_2 []bson.M
		err = PlayerUsers.Pipe(m3_2).All(&r3_2)
		if err != nil {
			beego.Error("RB PlayerUsers error32:", err)
		} else if len(r3_2) == 0 { // 未匹配到
			return
		}
		for _, r := range r3_2 {
			userid := r["_id"].(string)
			filterPlayerIds = append(filterPlayerIds, userid)
		}
	}
	if len(filterPlayerIds) > 0 {
		m["userid"] = bson.M{"$in": filterPlayerIds}
	}

	m1 := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":  "$userid",
			"date": bson.M{"$max": "$date"},
			// "date_str":                bson.M{"$sum": "$date_str"},
			"strategy_players":      bson.M{"$max": "$strategy_players"},
			"strategy_rounds":       bson.M{"$sum": "$strategy_rounds"},
			"all_in_rounds":         bson.M{"$sum": "$all_in_rounds"},
			"strategy_multi_rounds": bson.M{"$sum": "$strategy_multi_rounds"},
			"strategy_bets":         bson.M{"$sum": "$strategy_bets"},
			"strategy_wins":         bson.M{"$sum": "$strategy_wins"},
			"strategy_win_rounds":   bson.M{"$sum": "$strategy_win_rounds"},
			"strategy_loses":        bson.M{"$sum": "$strategy_loses"},
			"strategy_lose_rounds":  bson.M{"$sum": "$strategy_lose_rounds"},
			"bet_rounds":            bson.M{"$sum": "$bet_rounds"},
			"win_rounds":            bson.M{"$sum": "$win_rounds"},
			"lose_rounds":           bson.M{"$sum": "$lose_rounds"},
			"wins":                  bson.M{"$sum": "$wins"},
			"loses":                 bson.M{"$sum": "$loses"},
		}},
		{"$project": bson.M{
			"_id":                   "$_id",
			"date":                  "$date",
			"userid":                "$userid",
			"strategy_players":      "$strategy_players",
			"all_in_rounds":         "$all_in_rounds",
			"strategy_rounds":       "$strategy_rounds",
			"strategy_multi_rounds": "$strategy_multi_rounds",
			"strategy_bets":         "$strategy_bets",
			"strategy_win_rounds":   "$strategy_win_rounds",
			"strategy_wins":         "$strategy_wins",
			"strategy_loses":        "$strategy_loses",
			"strategy_lose_rounds":  "$strategy_lose_rounds",
			"bet_rounds":            "$bet_rounds",
			"win_rounds":            "$win_rounds",
			"lose_rounds":           "$lose_rounds",
			"wins":                  "$wins",
			"loses":                 "$loses",
		}},
		{"$match": m2},
		{"$count": "count"},
	}
	// 查总条数
	var rCount []bson.M
	err = RBJcfsStats.Pipe(m1).All(&rCount)
	if err != nil {
		return
	} else if len(rCount) > 0 {
		count = rCount[0]["count"].(int)
	}

	// 查列表数据
	m1 = m1[:len(m1)-1]
	m1 = append(m1, []bson.M{
		{"$sort": bson.M{"date": -1}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}...)
	err = RBJcfsStats.Pipe(m1).All(&stats)
	if err != nil {
		return
	}
	RBJcfsStatMapping(stats, m)

	// 查总汇数据
	mark := "--"
	summary := &entity.RBJcfsStat{}
	m1 = m1[:len(m1)-3]
	m1[1]["$group"].(bson.M)["_id"] = nil
	delete(m1[1]["$group"].(bson.M), "date")
	m1[1]["$group"].(bson.M)["userids"] = bson.M{"$addToSet": "$userid"}
	m1[1]["$group"].(bson.M)["strategy_players"] = bson.M{"$sum": "$strategy_players"} // 触发策略人数总汇

	m1[2]["$project"].(bson.M)["userids"] = "$userids"
	str, err := json.Marshal(m1)
	fmt.Println(err, string(str))
	var summarys []*entity.RBJcfsStat
	err = RBJcfsStats.Pipe(m1).All(&summarys)
	if err != nil {
		beego.Error("RB summary error:", err)
		return
	} else if len(summarys) > 0 {
		summary = summarys[0]
		summary.No = fmt.Sprintf("%s-%s总汇", startDate, endDate)
		summary.Id = mark
		summary.FNickname = mark
		summary.FUserType = mark
		summary.FRegistArea = mark
		summary.FCtime = mark
		summary.FLoginTime = mark
		RBJcfsStatSummaryMapping(summary, m)
	}
	if len(summarys) > 0 {
		stats = append([]*entity.RBJcfsStat{summary}, stats...)
	}
	return
}

func RBJcfsStatMapping(stats []*entity.RBJcfsStat, m bson.M) {
	var playerIds []string
	// var playerStats = make(map[string]*entity.CrashPlayerStat, len(stats))
	for _, stat := range stats {
		playerIds = append(playerIds, stat.Id)
		// playerStats[stat.Id] = stat
	}
	// 查用户基本信息
	m2 := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": playerIds}}},
		{"$project": bson.M{
			"_id":          "$_id",
			"nickname":     "$nickname",
			"money":        "$money",
			"cash_out":     "$cash_out",
			"diamond":      "$diamond",
			"ctime":        "$ctime",
			"login_time":   "$login_time",
			"regist_area":  "$regist_area", //ab测试 0:A 1:B 2:C
			"state":        "$state",
			"lhd_strategy": "$lhd_strategy",
		}},
	}
	var r2 []*entity.PlayerUser
	err := PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Error("lhd users error:", err)
		return
	}

	var r2Map = make(map[string]*entity.PlayerUser, len(r2))
	for _, r := range r2 {
		userid := r.Userid
		r2Map[userid] = r
	}
	// 查询总数据
	var rblist []entity.RBPlayerStat
	n := bson.M{}
	// 手动复制 m 到 n
	for k, v := range m {
		n[k] = v
	}
	n["userid"] = bson.M{"$in": playerIds}
	m3 := []bson.M{
		{"$match": n},
		{"$group": bson.M{
			"_id":        "$userid",
			"bet_rounds": bson.M{"$sum": "$bet_rounds"},
			"win_rounds": bson.M{"$sum": "$win_rounds"},
			"bets":       bson.M{"$sum": "$bets"},
			"win_bets":   bson.M{"$sum": "$win_bets"},
			"lose_bets":  bson.M{"$sum": "$lose_bets"},
			"wins":       bson.M{"$sum": "$wins"},
			"loses":      bson.M{"$sum": "$loses"},
		}},
	}
	RBPlayerStats.Pipe(m3).All(&rblist)

	now := utils.BsonNow()
	for i, stat := range stats {
		r, ok := r2Map[stat.Id]
		if !ok {
			continue
		}
		// 玩家昵称	玩家类型 账号类型 注册时间 最后登陆时间 存活天数 流失天数
		nickname := r.Nickname
		money := r.Money
		ctime := r.Ctime
		login_time := r.LoginTime
		regist_area := r.RegistArea
		state := r.State

		// 将日期转成印度时间
		login_time, _ = ConvertToIndiaTime(login_time.Unix())
		ctime, _ = ConvertToIndiaTime(ctime.Unix())

		stat.No = strconv.Itoa(i + 1)
		stat.FNickname = nickname
		stat.FCtime = ctime.Format(utils.FORMAT)
		stat.FLoginTime = login_time.Format(utils.FORMAT)
		stat.FLiveDays = int(math.Ceil(login_time.Sub(ctime).Hours() / 24))
		stat.FLoseDays = int(math.Floor(now.Sub(login_time).Hours() / 24))

		switch regist_area {
		case 0:
			stat.FRegistArea = "A类"
		case 1:
			stat.FRegistArea = "B类"
		case 2:
			stat.FRegistArea = "C类"
		}

		switch state {
		case 1:
			stat.FUserType = "新手"
		case 2:
			if money == 0 {
				stat.FUserType = "零充"
			} else if money >= 20000 && money <= 99999 {
				stat.FUserType = "普充"
			} else if money >= 100000 && money <= 499999 {
				stat.FUserType = "小R"
			} else if money >= 500000 && money <= 999999 {
				stat.FUserType = "中R"
			} else if money >= 1000000 && money <= 9999999 {
				stat.FUserType = "大R"
			} else if money >= 10000000 {
				stat.FUserType = "超大R"
			}
		case 3:
			stat.FUserType = "平民"
		case 4:
			stat.FUserType = "泡沫"
		}
		if stat.StrategyPlayers > 0 {
			stat.TriggerRoundsAvg = fmt.Sprintf("%.2f", float64(stat.StrategyRounds)/float64(stat.StrategyPlayers))
		}
		if stat.AllInRounds > 0 {
			stat.TiggerRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyRounds)/float64(stat.AllInRounds)*100.0)
		}
		if stat.StrategyRounds > 0 {
			stat.ExpiredRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyLoseRounds)/float64(stat.StrategyRounds)*100.0)
		}
		// 多门率
		if stat.StrategyRounds > 0 {
			stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyMultiRounds)/float64(stat.StrategyRounds)*100.0)
		}
		stat.FStrategyBets = fmt.Sprintf("%.2f", float64(stat.StrategyBets)/100.0)
		if stat.StrategyRounds > 0 {
			stat.FStrategyBetRoundsAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.StrategyBets, 100)/float64(stat.StrategyRounds))
		}
		stat.FStrategyWins = fmt.Sprintf("%.2f", float64(stat.StrategyWins)/100.0)
		stat.FStrategyLoses = fmt.Sprintf("%.2f", float64(stat.StrategyLoses)/100.0)
		stat.FStrategyCash = fmt.Sprintf("%.2f", float64(stat.StrategyWins+stat.StrategyLoses)/100.0)
		if len(rblist) > 0 {
			for _, c := range rblist {
				if c.Id == stat.Id {
					stat.WinRounds = c.WinRounds
					stat.BetRounds = c.BetRounds
					stat.Wins = c.Wins
					stat.Loses = c.Loses
					break
				}
			}
		}
		if stat.BetRounds > 0 {
			stat.TotalWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
		}
		if stat.StrategyRounds > 0 {
			stat.StrategyWinRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyWinRounds)/float64(stat.StrategyRounds)*100.0)
		}
		if -stat.StrategyLoses > 0 {
			stat.StrategyRewardRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.StrategyWins, 100)/ComputeFloat(-stat.StrategyLoses, 100))*100.0)
		}
		if -stat.Loses > 0 {
			stat.TotalRewardRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.Wins, 100)/ComputeFloat(-stat.Loses, 100))*100.0)
		}
		if r.Money > 0 {
			stat.PlayerRewardRate = fmt.Sprintf("%.2f%%", ComputeFloat(int64(r.CashOut+int32(r.Diamond)), 100)/ComputeFloat(int64(r.Money), 100)*100.0)
			//fmt.Sprintf("%d", int((int64(r.CashOut)+r.Diamond)*10000/int64(r.Money)))
		} else {
			stat.PlayerRewardRate = "未充值"
		}
	}
}

func RBJcfsStatSummaryMapping(stat *entity.RBJcfsStat, m bson.M) {
	// 查询总数据
	rbinfo := new(entity.RBPlayerStat)
	n := bson.M{}
	// 手动复制 m 到 n
	for k, v := range m {
		n[k] = v
	}
	n["userid"] = bson.M{"$in": stat.Userids}
	m3 := []bson.M{
		{"$match": n},
		{"$group": bson.M{
			"_id":        nil,
			"bet_rounds": bson.M{"$sum": "$bet_rounds"},
			"win_rounds": bson.M{"$sum": "$win_rounds"},
			"bets":       bson.M{"$sum": "$bets"},
			"win_bets":   bson.M{"$sum": "$win_bets"},
			"lose_bets":  bson.M{"$sum": "$lose_bets"},
			"wins":       bson.M{"$sum": "$wins"},
			"loses":      bson.M{"$sum": "$loses"},
		}},
	}
	RBPlayerStats.Pipe(m3).One(&rbinfo)

	stat.StrategyPlayers = int32(len(stat.Userids))
	if stat.StrategyPlayers > 0 {
		stat.TriggerRoundsAvg = fmt.Sprintf("%.2f", float64(stat.StrategyRounds)/float64(stat.StrategyPlayers))
	}
	if stat.AllInRounds > 0 {
		stat.TiggerRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyRounds)/float64(stat.AllInRounds)*100.0)
	}
	if stat.StrategyRounds > 0 {
		stat.ExpiredRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyLoseRounds)/float64(stat.StrategyRounds)*100.0)
	}
	// 多门率
	if stat.StrategyRounds > 0 {
		stat.PlayerMultisRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyMultiRounds)/float64(stat.StrategyRounds)*100.0)
	}
	stat.FStrategyBets = fmt.Sprintf("%.2f", float64(stat.StrategyBets)/100.0)
	if stat.StrategyRounds > 0 {
		stat.FStrategyBetRoundsAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.StrategyBets, 100)/float64(stat.StrategyRounds))
	}
	stat.FStrategyWins = fmt.Sprintf("%.2f", float64(stat.StrategyWins)/100.0)
	stat.FStrategyLoses = fmt.Sprintf("%.2f", float64(stat.StrategyLoses)/100.0)
	stat.FStrategyCash = fmt.Sprintf("%.2f", float64(stat.StrategyWins+stat.StrategyLoses)/100.0)
	// 查询总局数
	stat.WinRounds = rbinfo.WinRounds
	stat.BetRounds = rbinfo.BetRounds
	stat.Wins = rbinfo.Wins
	stat.Loses = rbinfo.Loses
	if stat.BetRounds > 0 {
		stat.TotalWinRate = fmt.Sprintf("%.2f%%", float64(stat.WinRounds)/float64(stat.BetRounds)*100.0)
	}
	if stat.StrategyRounds > 0 {
		stat.StrategyWinRate = fmt.Sprintf("%.2f%%", float64(stat.StrategyWinRounds)/float64(stat.StrategyRounds)*100.0)
	}
	if -stat.StrategyLoses > 0 {
		stat.StrategyRewardRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.StrategyWins, 100)/ComputeFloat(-stat.StrategyLoses, 100))*100.0)
	}
	if -stat.Loses > 0 {
		stat.TotalRewardRate = fmt.Sprintf("%.2f%%", (ComputeFloat(stat.Wins, 100)/ComputeFloat(-stat.Loses, 100))*100.0)
	}
}

// TpNewStatPlayers 新tp汇总明细
func (s *gameStatsService) TpNewStatPlayers(page, pageSize int, params map[string]any) (stats []*entity.TpNewStatPlayer, total int, err error) {
	var where1, where2, where3, having4 string
	var args1, args2, args3, args4 []any

	if userid, ok := params["userid"]; ok {
		where1 += " and userid = ?"
		where2 += " and userid = ?"
		where3 += " and userid = ?"
		args1 = append(args1, userid)
		args2 = append(args2, userid)
		args3 = append(args3, userid)
	}

	// 时间
	startTime, ok1 := params["startTime"]
	endTime, ok2 := params["endTime"]
	if ok1 && ok2 {
		where1 += " and begin_time BETWEEN ? AND ?"
		where2 += " and begin_time BETWEEN ? AND ?"
		args1 = append(args1, startTime, endTime)
		args2 = append(args2, startTime, endTime)
	} else if ok1 {
		where1 += " and begin_time >= ?"
		where2 += " and begin_time >= ?"
		args1 = append(args1, startTime)
		args2 = append(args2, startTime)
	} else if ok2 {
		where1 += " and begin_time <= ?"
		where2 += " and begin_time <= ?"
		args1 = append(args1, endTime)
		args2 = append(args2, endTime)
	}

	// 打码量
	bets1, ok1 := params["bets1"]
	bets2, ok2 := params["bets2"]
	if ok1 && ok2 {
		having4 += " and bet_total between ? and ?"
		args4 = append(args4, bets1, bets2)
	}
	// 注均码
	betAvg1, ok1 := params["betAvg1"]
	betAvg2, ok2 := params["betAvg2"]
	if ok1 && ok2 {
		having4 += " and bet_avg between ? and ?"
		args4 = append(args4, betAvg1, betAvg2)
	}

	// 净赢
	cash1, ok1 := params["cash1"]
	cash2, ok2 := params["cash2"]
	if ok1 && ok2 {
		having4 += " and score_total between ? and ?"
		args4 = append(args4, cash1, cash2)
	}

	// 流失天数
	var uFilter bool
	uLossDays1, ok1 := params["uLossDays1"]
	uLossDays2, ok2 := params["uLossDays2"]
	uFilter = ok1 || ok2
	if ok1 && ok2 {
		where3 += " and date_diff('day', login_time, now()) between ? and ?"
		args3 = append(args3, uLossDays1, uLossDays2)
	}
	// money
	utypeQ, ok1 := params["utypeQ"]
	if ok1 {
		where3 += utypeQ.(string)
		uFilter = true
	}

	// 渠道
	if ad__bundle_id, ok := params["ad__bundle_id"]; ok {
		where3 += " and ad__bundle_id in ?"
		args3 = append(args3, ad__bundle_id)
		uFilter = true
	}
	// ab 类
	if regist_area, ok := params["regist_area"]; ok {
		where3 += " and regist_area in ?"
		args3 = append(args3, regist_area)
		uFilter = true
	}

	sql0 := `
		select userid, win_type, count(*) round, sum(bet_amount) bet_total, avg(bet_amount) bet_avg, sum(score) score_total, avg(score) score_avg, 
		SUM(case when is_change_table = 1 then 1 else 0 end) change_table_count,
		SUM((case when win_type != 4 and player_num = 2 then 1 else 0 end)) Rounds2,
		SUM((case when win_type != 4 and player_num = 3 then 1 else 0 end)) Rounds3,
		SUM((case when win_type != 4 and player_num = 4 then 1 else 0 end)) Rounds4,
		SUM((case when win_type != 4 and player_num = 5 then 1 else 0 end)) Rounds5,
		SUM((case when win_type = 1 and player_num == 2 then 1 else 0 end)) WinRounds2,
		SUM((case when win_type = 1 and player_num == 3 then 1 else 0 end)) WinRounds3,
		SUM((case when win_type = 1 and player_num == 4 then 1 else 0 end)) WinRounds4,
		SUM((case when win_type = 1 and player_num == 5 then 1 else 0 end)) WinRounds5,	
		SUM((case tp_hand_type_up_down when 10 then 1 else 0 end)) hua1,
			SUM((case tp_hand_type_up_down when 11 then 1 else 0 end)) hua2,
			SUM((case tp_hand_type_up_down when 20 then 1 else 0 end)) hua3,
			SUM((case tp_hand_type_up_down when 21 then 1 else 0 end)) hua4,
			SUM((case tp_hand_type_up_down when 30 then 1 else 0 end)) hua5,
			SUM((case tp_hand_type_up_down when 31 then 1 else 0 end)) hua6,
			SUM((case tp_hand_type_up_down when 40 then 1 else 0 end)) hua7,
			SUM((case tp_hand_type_up_down when 41 then 1 else 0 end)) hua8,
			SUM((case tp_hand_type_up_down when 50 then 1 else 0 end)) hua9,
			SUM((case tp_hand_type_up_down when 51 then 1 else 0 end)) hua10,
			SUM((case tp_hand_type_up_down when 60 then 1 else 0 end)) hua11,
			SUM((case tp_hand_type_up_down when 61 then 1 else 0 end)) hua12,
			SUM(tp_opponent) opponent_total, SUM(tp_hurt) hurt_total,
			SUM(case when tp_opponent = 1 then score else 0 end) opponent_score,
			SUM(case when tp_hurt = 1 then score else 0 end) hurt_score,
			SUM(charge_times) charge_times, 
			SUM(charge_launch_times) charge_launch_times, 
			SUM(charge_pay_times) charge_pay_times
		from game.col_detail final 
		where gtype = 1 and robot = 0 %s
		and userid in (
			select userid from (
				select userid, SUM(bet_amount) bet_total, avg(bet_amount) bet_avg, SUM(score) score_total, max(begin_time) lately_time 
				from game.col_detail final 
				where gtype = 1 and robot = 0 %s
				%s
				group by userid
				having 1 = 1 %s
			)
			order by lately_time desc limit ?, ?
		)
		group by userid, win_type
		order by max(begin_time) desc
	`
	uFilterSql3 := `
		and userid in (
			select userid from game.col_user final where 1 = 1  %s
		)
	`
	if !uFilter {
		uFilterSql3 = ""
		args3 = []any{}
	} else {
		uFilterSql3 = fmt.Sprintf(uFilterSql3, where3)
	}

	// 查总条数 , max(begin_time) lately_time
	err = ck.Select(&total, fmt.Sprintf(`
		select count(*) from (
			select userid, SUM(bet_amount) bet_total, avg(bet_amount) bet_avg, SUM(score) score_total
			from game.col_detail final 
			where gtype = 1 and robot = 0  %s
			%s
			group by userid
			having 1 = 1 %s
		)
	`, where2, uFilterSql3, having4), append(append(args2, args3...), args4...)...)
	if err != nil {
		return
	}

	sql0 = fmt.Sprintf(sql0, where1, where2, uFilterSql3, having4)
	var datas []map[string]any
	args := append(args1, args2...)
	args = append(args, args3...)
	args = append(args, args4...)
	offset, limit := PageCalc(page, pageSize)
	// args = append(args, offset, limit)
	err = ck.Select(&datas, sql0, append(args, offset, limit)...)
	if err != nil {
		return
	}

	var uidStats = make(map[string]*entity.TpNewStatPlayer)
	var userids []string
	var no int
	for _, data := range datas { // 输和赢的统计汇总
		userid := data["userid"].(string)
		stat, ok := uidStats[userid]
		if !ok {
			no++
			stat = &entity.TpNewStatPlayer{
				No:              strconv.Itoa(no),
				UserId:          userid,
				HuaTypeTimes:    [13]int{},
				HuaTypeWinTimes: [13]int{},
			}
			uidStats[userid] = stat
			stats = append(stats, stat)
			userids = append(userids, userid)
		}
		mappingStatPlayer1(data, stat)
	}

	// 牌型输赢
	var huaTypeDatas []map[string]any
	sqlHandType := fmt.Sprintf(`
		select userid, win_type, tp_hand_type_up_down, SUM(score) score_total, AVG(score) score_avg, count(*) round, MAX(score) score_max, MIN(score) score_min
		from game.col_detail final 
		where gtype = 1 and robot = 0 %s
			and userid in ?
		group by userid, win_type, tp_hand_type_up_down
	`, where2)
	argsHandType := append(args2, userids)
	err = ck.Select(&huaTypeDatas, sqlHandType, argsHandType...)
	if err != nil {
		return
	}
	for _, data := range huaTypeDatas {
		userid := data["userid"].(string)
		stat, ok := uidStats[userid]
		if !ok {
			continue
		}
		mappingStatPlayer2HuaType(data, stat)
	}

	for _, stat := range stats { // 汇总后二次计算
		mappingStatPlayer0(stat)
	}
	var totalHeart float64
	// 查用户信息
	var users []map[string]any
	ck.Select(&users, `
		select userid,nickname,ctime,login_time,money,state,regist_area,
			date_diff('day', ctime, now()) live_days, date_diff('day', login_time, now()) loss_days,tp_heartbeat_hr
		from game.col_user final where userid in ?
	`, userids)
	for _, user := range users {
		userid := user["userid"].(string)
		nickname := user["nickname"].(string)
		ctime := user["ctime"].(time.Time)
		login_time := user["login_time"].(time.Time)
		money := utils.ToInt64(user["money"])
		live_days := utils.ToInt64(user["live_days"])
		loss_days := utils.ToInt64(user["loss_days"])
		state := utils.ToInt64(user["state"])
		regist_area := utils.ToInt64(user["regist_area"])
		tp_heartbeat_hr := utils.ToFloat64(user["tp_heartbeat_hr"])

		stat, ok := uidStats[userid]
		if ok {
			stat.Nickname = nickname
			stat.Ctime = ctime.Format(utils.FORMAT)
			stat.LoginTime = login_time.Format(utils.FORMAT)
			stat.LiveDays = int(live_days)
			stat.LoseDays = int(loss_days)
			stat.TpHeartbeatHr = tp_heartbeat_hr
			if tp_heartbeat_hr > 0 {
				totalHeart += tp_heartbeat_hr
				stat.PlayerHeartRate = fmt.Sprintf("%.1f", tp_heartbeat_hr)
			}
			switch regist_area {
			case 0, 3:
				stat.RegistArea = "A类"
			case 1:
				stat.RegistArea = "B类"
			case 2:
				stat.RegistArea = "C类"
			}
			switch state {
			case 1:
				stat.UserType = "新手"
			case 2:
				if money == 0 {
					stat.UserType = "零充"
				} else if money >= 20000 && money <= 99999 {
					stat.UserType = "普充"
				} else if money >= 100000 && money <= 499999 {
					stat.UserType = "小R"
				} else if money >= 500000 && money <= 999999 {
					stat.UserType = "中R"
				} else if money >= 1000000 && money <= 9999999 {
					stat.UserType = "大R"
				} else if money >= 10000000 {
					stat.UserType = "超大R"
				}
			case 3:
				stat.UserType = "平民"
			case 4:
				stat.UserType = "泡沫"
			}
		}
	}

	// 查总汇数据
	mark := "--"
	summary := &entity.TpNewStatPlayer{
		No:            fmt.Sprintf("%v-%v总汇", params["startDate"], params["endDate"]),
		UserId:        mark,
		Nickname:      mark,
		Ctime:         mark,
		LoginTime:     mark,
		UserType:      mark,
		RegistArea:    mark,
		TpHeartbeatHr: totalHeart,
		TPNum:         int64(len(userids)),
	}
	sqlSummary := strings.ReplaceAll(sql0, "select userid, win_type", "select win_type")
	sqlSummary = strings.ReplaceAll(sqlSummary, "limit ?, ?", "")
	sqlSummary = strings.ReplaceAll(sqlSummary, "group by userid, win_type", "group by win_type")
	var summaryDatas []map[string]any
	err = ck.Select(&summaryDatas, sqlSummary, args...)
	if err != nil {
		return
	}
	for _, data := range summaryDatas {
		mappingStatPlayer1(data, summary)
	}
	// 牌型输赢
	var summaryHuaTypeDatas []map[string]any
	sqlSummaryHandType := strings.ReplaceAll(sqlHandType, "select userid, win_type", "select win_type")
	sqlSummaryHandType = strings.ReplaceAll(sqlSummaryHandType, "group by userid, win_type", "group by win_type")
	err = ck.Select(&summaryHuaTypeDatas, sqlSummaryHandType, argsHandType...)
	if err != nil {
		return
	}
	for _, data := range summaryHuaTypeDatas {
		mappingStatPlayer2HuaType(data, summary)
	}
	// 汇总后二次计算
	mappingStatPlayer0(summary)
	stats = append([]*entity.TpNewStatPlayer{summary}, stats...)
	return
}

func mappingStatPlayer1(src map[string]any, dst *entity.TpNewStatPlayer) {
	win_type := utils.ToInt64(src["win_type"])
	dst.AllRounds += int32(utils.ToInt64(src["round"]))
	dst.Bets += utils.ToInt64(src["bet_total"])
	dst.Cash += utils.ToInt64(src["score_total"])
	dst.OpponentCash += utils.ToFloat64(src["opponent_score"]) / 100          // 冤人局净赢
	dst.HurtCash += utils.ToFloat64(src["hurt_score"]) / 100                  // 偷与被偷净赢
	dst.ChargeTimes += int16(utils.ToInt64(src["charge_times"]))              // 充值触发次数
	dst.ChargeLaunchTimes += int16(utils.ToInt64(src["charge_launch_times"])) // 局内充值拉单次数
	dst.ChargePayTimes += int16(utils.ToInt64(src["charge_pay_times"]))       // 局内充值实付次数
	// change_table_count
	dst.ChangeTableCount += utils.ToInt64(src["change_table_count"])
	if win_type == 1 { // 赢
		dst.WinRounds = int32(utils.ToInt64(src["round"]))
		dst.WinBets = utils.ToInt64(src["bet_total"])
		dst.WinBetsAvg = utils.ToInt64(src["bet_avg"])
		dst.Wins = utils.ToInt64(src["score_total"])
		dst.FWinsAvg = fmt.Sprintf("%.2f", float64(utils.ToInt64(src["score_avg"]))/100)
		// 牌型赢次数
		for i := 1; i <= 12; i++ {
			dst.HuaTypeWinTimes[i] = int(utils.ToInt64(src[fmt.Sprintf("hua%d", i)]))
		}
		dst.OpponentWinRound = utils.ToInt64(src["opponent_total"])
		dst.HurtWinRound = utils.ToInt64(src["hurt_total"])
		dst.OpponentWins = utils.ToFloat64(src["opponent_score"]) / 100
		dst.HurtWins = utils.ToFloat64(src["hurt_score"]) / 100
	} else { // 输
		dst.LoseRounds = int32(utils.ToInt64(src["round"]))
		dst.LoseBets = utils.ToInt64(src["bet_total"])
		dst.LoseBetsAvg = utils.ToInt64(src["bet_avg"])
		dst.Loses = utils.ToInt64(src["score_total"])
		dst.FLosesAvg = fmt.Sprintf("%.2f", float64(utils.ToInt64(src["score_avg"]))/100)
		dst.OpponentLoseRound = utils.ToInt64(src["opponent_total"])
		dst.HurtLoseRound = utils.ToInt64(src["hurt_total"])
		dst.OpponentLoses = utils.ToFloat64(src["opponent_score"]) / 100
		dst.HurtLoses = utils.ToFloat64(src["hurt_score"]) / 100
	}
	dst.Rounds2 += utils.ToInt64(src["Rounds2"])
	dst.Rounds3 += utils.ToInt64(src["Rounds3"])
	dst.Rounds4 += utils.ToInt64(src["Rounds4"])
	dst.Rounds5 += utils.ToInt64(src["Rounds5"])
	dst.WinRounds2 += utils.ToInt64(src["WinRounds2"])
	dst.WinRounds3 += utils.ToInt64(src["WinRounds3"])
	dst.WinRounds4 += utils.ToInt64(src["WinRounds4"])
	dst.WinRounds5 += utils.ToInt64(src["WinRounds5"])
	if dst.TPNum > 0 {
		// TPNum
		dst.PlayerHeartRate = fmt.Sprintf("%.1f", float64(dst.TpHeartbeatHr)/float64(dst.TPNum)) // 玩家心率
	}
	// 拿牌型次数
	for i := 1; i <= 12; i++ {
		dst.HuaTypeTimes[i] += int(utils.ToInt64(src[fmt.Sprintf("hua%d", i)]))
	}
}

func mappingStatPlayer0(dst *entity.TpNewStatPlayer) {
	if dst.AllRounds > 0 {
		dst.BetsAvg = fmt.Sprintf("%.2f", float64(dst.Bets)/float64(dst.AllRounds)/100) // 局均打码量
	}
	if dst.AllRounds > 0 {
		dst.WinRate = fmt.Sprintf("%.2f%%", float64(dst.WinRounds)/float64(dst.AllRounds)*100) // 玩家胜率
	}
	if dst.Bets != 0 {
		// dst.RewardRate = fmt.Sprintf("%.2f%%", float64(dst.Wins)/float64(-dst.Loses)*100) // 总返奖率
		dst.RewardRate = fmt.Sprintf("%.2f%%", float64(dst.Wins+dst.WinBets)/float64(dst.Bets)*100) // 总返奖率
	}
	if dst.AllRounds > 0 {
		dst.FCashAvg = fmt.Sprintf("%.2f", float64(dst.Cash)/float64(dst.AllRounds)/100) // 局均净赢
	}
	if dst.AllRounds > 0 {
		dst.Hua1Rate = utils.CaseElse(dst.HuaTypeTimes[1] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[1])/float64(dst.AllRounds)*100))    // 玩家拿大豹子率
		dst.Hua2Rate = utils.CaseElse(dst.HuaTypeTimes[2] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[2])/float64(dst.AllRounds)*100))    // 玩家拿小豹子率
		dst.Hua3Rate = utils.CaseElse(dst.HuaTypeTimes[3] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[3])/float64(dst.AllRounds)*100))    // 玩家拿大顺金率
		dst.Hua4Rate = utils.CaseElse(dst.HuaTypeTimes[4] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[4])/float64(dst.AllRounds)*100))    // 玩家拿小顺金率
		dst.Hua5Rate = utils.CaseElse(dst.HuaTypeTimes[5] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[5])/float64(dst.AllRounds)*100))    // 玩家拿大顺子率
		dst.Hua6Rate = utils.CaseElse(dst.HuaTypeTimes[6] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[6])/float64(dst.AllRounds)*100))    // 玩家拿小顺子率
		dst.Hua7Rate = utils.CaseElse(dst.HuaTypeTimes[7] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[7])/float64(dst.AllRounds)*100))    // 玩家拿大同花率
		dst.Hua8Rate = utils.CaseElse(dst.HuaTypeTimes[8] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[8])/float64(dst.AllRounds)*100))    // 玩家拿小同花率
		dst.Hua9Rate = utils.CaseElse(dst.HuaTypeTimes[9] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[9])/float64(dst.AllRounds)*100))    // 玩家拿大对子率
		dst.Hua10Rate = utils.CaseElse(dst.HuaTypeTimes[10] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[10])/float64(dst.AllRounds)*100)) // 玩家拿小对子率
		dst.Hua11Rate = utils.CaseElse(dst.HuaTypeTimes[11] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[11])/float64(dst.AllRounds)*100)) // 玩家拿大高牌率
		dst.Hua12Rate = utils.CaseElse(dst.HuaTypeTimes[12] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[12])/float64(dst.AllRounds)*100)) // 玩家拿小高牌率
	}
	dst.Hua1WinRate = utils.CaseElse(dst.HuaTypeTimes[1] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[1], dst.HuaTypeTimes[1])*100))     // 玩家拿大豹子胜率
	dst.Hua2WinRate = utils.CaseElse(dst.HuaTypeTimes[2] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[2], dst.HuaTypeTimes[2])*100))     // 玩家拿小豹子胜率
	dst.Hua3WinRate = utils.CaseElse(dst.HuaTypeTimes[3] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[3], dst.HuaTypeTimes[3])*100))     // 玩家拿大顺金胜率
	dst.Hua4WinRate = utils.CaseElse(dst.HuaTypeTimes[4] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[4], dst.HuaTypeTimes[4])*100))     // 玩家拿小顺金胜率
	dst.Hua5WinRate = utils.CaseElse(dst.HuaTypeTimes[5] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[5], dst.HuaTypeTimes[5])*100))     // 玩家拿大顺子胜率
	dst.Hua6WinRate = utils.CaseElse(dst.HuaTypeTimes[6] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[6], dst.HuaTypeTimes[6])*100))     // 玩家拿小顺子胜率
	dst.Hua7WinRate = utils.CaseElse(dst.HuaTypeTimes[7] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[7], dst.HuaTypeTimes[7])*100))     // 玩家拿大同花胜率
	dst.Hua8WinRate = utils.CaseElse(dst.HuaTypeTimes[8] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[8], dst.HuaTypeTimes[8])*100))     // 玩家拿小同花胜率
	dst.Hua9WinRate = utils.CaseElse(dst.HuaTypeTimes[9] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[9], dst.HuaTypeTimes[9])*100))     // 玩家拿大对子胜率
	dst.Hua10WinRate = utils.CaseElse(dst.HuaTypeTimes[10] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[10], dst.HuaTypeTimes[10])*100)) // 玩家拿小对子胜率
	dst.Hua11WinRate = utils.CaseElse(dst.HuaTypeTimes[11] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[11], dst.HuaTypeTimes[11])*100)) // 玩家拿大高牌胜率
	dst.Hua12WinRate = utils.CaseElse(dst.HuaTypeTimes[12] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[12], dst.HuaTypeTimes[12])*100)) // 玩家拿小高牌胜率
	if dst.Rounds2 > 0 {
		dst.WinRounds2Rate = fmt.Sprintf("%.2f%%", float64(dst.WinRounds2)/float64(dst.Rounds2)*100) // 玩家2人局胜率
	}
	if dst.Rounds3 > 0 {
		dst.WinRounds3Rate = fmt.Sprintf("%.2f%%", float64(dst.WinRounds3)/float64(dst.Rounds3)*100) // 玩家3人局胜率
	}
	if dst.Rounds4 > 0 {
		dst.WinRounds4Rate = fmt.Sprintf("%.2f%%", float64(dst.WinRounds4)/float64(dst.Rounds4)*100) // 玩家4人局胜率
	}
	if dst.Rounds5 > 0 {
		dst.WinRounds5Rate = fmt.Sprintf("%.2f%%", float64(dst.WinRounds5)/float64(dst.Rounds5)*100) // 玩家5人局胜率
	}
	if dst.ChangeTableCount > 0 {
		dst.HZFrequency = fmt.Sprintf("%.2f", float64(dst.AllRounds)/float64(dst.ChangeTableCount))
	}
	dst.FBets = fmt.Sprintf("%.2f", Chip2Float(dst.Bets))
	dst.FWinBets = fmt.Sprintf("%.2f", Chip2Float(dst.WinBets))
	dst.FLoseBets = fmt.Sprintf("%.2f", Chip2Float(dst.LoseBets))
	dst.FWinBetsAvg = fmt.Sprintf("%.2f", Chip2Float(dst.WinBetsAvg))
	dst.FLoseBetsAvg = fmt.Sprintf("%.2f", Chip2Float(dst.LoseBetsAvg))
	dst.FWins = fmt.Sprintf("%.2f", Chip2Float(dst.Wins))
	dst.FLoses = fmt.Sprintf("%.2f", Chip2Float(dst.Loses))
	dst.FCash = fmt.Sprintf("%.2f", Chip2Float(dst.Cash))
	dst.FOpponentWins = fmt.Sprintf("%.2f", dst.OpponentWins)
	dst.FOpponentLoses = fmt.Sprintf("%.2f", dst.OpponentLoses)
	dst.FOpponentCash = fmt.Sprintf("%.2f", dst.OpponentCash)
	dst.FHurtWins = fmt.Sprintf("%.2f", dst.HurtWins)
	dst.FHurtLoses = fmt.Sprintf("%.2f", dst.HurtLoses)
	dst.FHurtCash = fmt.Sprintf("%.2f", dst.HurtCash)
	dst.FMaxWinHuaTypeAmount = fmt.Sprintf("%.2f", Chip2Float(dst.MaxWinHuaTypeAmount))
	dst.FMaxLoseHuaTypeAmount = fmt.Sprintf("%.2f", Chip2Float(dst.MaxLoseHuaTypeAmount))
	dst.FMaxWinHuaTypeAmountAvg = fmt.Sprintf("%.2f", Chip2Float(dst.MaxWinHuaTypeAmountAvg))
	dst.FMaxLoseHuaTypeAmountAvg = fmt.Sprintf("%.2f", Chip2Float(dst.MaxLoseHuaTypeAmountAvg))
	dst.FOnceMaxWinHuaTypeAmount = fmt.Sprintf("%.2f", Chip2Float(dst.OnceMaxWinHuaTypeAmount))
	dst.FOnceMaxLoseHuaTypeAmount = fmt.Sprintf("%.2f", Chip2Float(dst.OnceMaxLoseHuaTypeAmount))
}

var HuaTypeNames = map[int]string{
	10: "大豹子",
	11: "小豹子",
	20: "大同花顺",
	21: "小同花顺",
	30: "大顺子",
	31: "小顺子",
	40: "大同花",
	41: "小同花",
	50: "大对子",
	51: "小对子",
	60: "大高牌",
	61: "小高牌",
}

// 牌型输赢 mapping
func mappingStatPlayer2HuaType(src map[string]any, dst *entity.TpNewStatPlayer) {
	win_type := utils.ToInt64(src["win_type"])
	tp_hand_type_up_down := utils.ToInt64(src["tp_hand_type_up_down"])
	round := utils.ToInt64(src["round"])
	score_total := utils.ToInt64(src["score_total"])
	score_avg := utils.ToInt64(src["score_avg"])
	score_max := utils.ToInt64(src["score_max"])
	score_min := utils.ToInt64(src["score_min"])
	if win_type == 1 {
		if score_total > dst.MaxWinHuaTypeAmount {
			dst.MaxWinHuaTypeAmount = score_total
			dst.MaxWinHuaType = HuaTypeNames[int(tp_hand_type_up_down)]
			dst.MaxWinHuaTypeRound = round
			dst.MaxWinHuaTypeAmountAvg = score_avg
		}
		if score_max > dst.OnceMaxWinHuaTypeAmount {
			dst.OnceMaxWinHuaTypeAmount = score_max
			dst.OnceMaxWinHuaType = HuaTypeNames[int(tp_hand_type_up_down)]
		}
	} else {
		if score_total < dst.MaxLoseHuaTypeAmount {
			dst.MaxLoseHuaTypeAmount = score_total
			dst.MaxLoseHuaType = HuaTypeNames[int(tp_hand_type_up_down)]
			dst.MaxLoseHuaTypeRound = round
			dst.MaxLoseHuaTypeAmountAvg = score_avg
		}
		if score_min < dst.OnceMaxLoseHuaTypeAmount {
			dst.OnceMaxLoseHuaTypeAmount = score_min
			dst.OnceMaxLoseHuaType = HuaTypeNames[int(tp_hand_type_up_down)]
		}
	}
}

func (s *gameStatsService) TpNewStatDates(params map[string]any) (stats []*entity.TpNewStatDate, err error) {
	var where1, where2, where3, having4 string
	var args1, args2, args3, args4 []any

	if userid, ok := params["userid"]; ok {
		where1 += " and userid = ?"
		where2 += " and userid = ?"
		where3 += " and userid = ?"
		args1 = append(args1, userid)
		args2 = append(args2, userid)
		args3 = append(args3, userid)
	}

	// 时间
	startTime, ok1 := params["startTime"]
	endTime, ok2 := params["endTime"]
	if ok1 && ok2 {
		where1 += " and begin_time BETWEEN ? AND ?"
		where2 += " and begin_time BETWEEN ? AND ?"
		args1 = append(args1, startTime, endTime)
		args2 = append(args2, startTime, endTime)
	} else if ok1 {
		where1 += " and begin_time >= ?"
		where2 += " and begin_time >= ?"
		args1 = append(args1, startTime)
		args2 = append(args2, startTime)
	} else if ok2 {
		where1 += " and begin_time <= ?"
		where2 += " and begin_time <= ?"
		args1 = append(args1, endTime)
		args2 = append(args2, endTime)
	}

	// 打码量
	bets1, ok1 := params["bets1"]
	bets2, ok2 := params["bets2"]
	if ok1 && ok2 {
		having4 += " and bet_total between ? and ?"
		args4 = append(args4, bets1, bets2)
	}
	// 注均码
	betAvg1, ok1 := params["betAvg1"]
	betAvg2, ok2 := params["betAvg2"]
	if ok1 && ok2 {
		having4 += " and bet_avg between ? and ?"
		args4 = append(args4, betAvg1, betAvg2)
	}

	// 净赢
	cash1, ok1 := params["cash1"]
	cash2, ok2 := params["cash2"]
	if ok1 && ok2 {
		having4 += " and score_total between ? and ?"
		args4 = append(args4, cash1, cash2)
	}

	// 流失天数
	var uFilter bool
	uLossDays1, ok1 := params["uLossDays1"]
	uLossDays2, ok2 := params["uLossDays2"]
	uFilter = ok1 || ok2
	if ok1 && ok2 {
		where3 += " and date_diff('day', login_time, now()) between ? and ?"
		args3 = append(args3, uLossDays1, uLossDays2)
	}
	// money
	utypeQ, ok1 := params["utypeQ"]
	if ok1 {
		where3 += utypeQ.(string)
		uFilter = true
	}

	// 渠道
	if ad__bundle_id, ok := params["ad__bundle_id"]; ok {
		where3 += " and ad__bundle_id in ?"
		args3 = append(args3, ad__bundle_id)
		uFilter = true
	}
	// ab 类
	if regist_area, ok := params["regist_area"]; ok {
		where3 += " and regist_area in ?"
		args3 = append(args3, regist_area)
		uFilter = true
	}

	sql0 := `
		select toYYYYMMDD(toDateTime(begin_time)) datestr, win_type, count(*) round, sum(bet_amount) bet_total, avg(bet_amount) bet_avg, sum(score) score_total, avg(score) score_avg, 
		SUM(case when is_change_table = 1 then 1 else 0 end) change_table_count,
		SUM((case when win_type != 4 and player_num = 2 then 1 else 0 end)) Rounds2,
		SUM((case when win_type != 4 and player_num = 3 then 1 else 0 end)) Rounds3,
		SUM((case when win_type != 4 and player_num = 4 then 1 else 0 end)) Rounds4,
		SUM((case when win_type != 4 and player_num = 5 then 1 else 0 end)) Rounds5,
		SUM((case when win_type = 1 and player_num == 2 then 1 else 0 end)) WinRounds2,
		SUM((case when win_type = 1 and player_num == 3 then 1 else 0 end)) WinRounds3,
		SUM((case when win_type = 1 and player_num == 4 then 1 else 0 end)) WinRounds4,
		SUM((case when win_type = 1 and player_num == 5 then 1 else 0 end)) WinRounds5,		
			SUM((case tp_hand_type_up_down when 10 then 1 else 0 end)) hua1,
			SUM((case tp_hand_type_up_down when 11 then 1 else 0 end)) hua2,
			SUM((case tp_hand_type_up_down when 20 then 1 else 0 end)) hua3,
			SUM((case tp_hand_type_up_down when 21 then 1 else 0 end)) hua4,
			SUM((case tp_hand_type_up_down when 30 then 1 else 0 end)) hua5,
			SUM((case tp_hand_type_up_down when 31 then 1 else 0 end)) hua6,
			SUM((case tp_hand_type_up_down when 40 then 1 else 0 end)) hua7,
			SUM((case tp_hand_type_up_down when 41 then 1 else 0 end)) hua8,
			SUM((case tp_hand_type_up_down when 50 then 1 else 0 end)) hua9,
			SUM((case tp_hand_type_up_down when 51 then 1 else 0 end)) hua10,
			SUM((case tp_hand_type_up_down when 60 then 1 else 0 end)) hua11,
			SUM((case tp_hand_type_up_down when 61 then 1 else 0 end)) hua12,
			SUM(tp_opponent) opponent_total, SUM(tp_hurt) hurt_total,
			SUM(case when tp_opponent = 1 then score else 0 end) opponent_score,
			SUM(case when tp_hurt = 1 then score else 0 end) hurt_score,
			SUM(charge_times) charge_times, 
			SUM(charge_launch_times) charge_launch_times,
			SUM(charge_pay_times) charge_pay_times
		from game.col_detail final 
		where gtype = 1 and robot = 0 %s
		and userid in (
			select userid from (
				select userid, SUM(bet_amount) bet_total, avg(bet_amount) bet_avg, SUM(score) score_total
				from game.col_detail final 
				where gtype = 1 and robot = 0 %s
				%s
				group by userid
				having 1 = 1 %s
			)
		)
		group by datestr, win_type
		order by datestr desc
	`
	uFilterSql3 := `
		and userid in (
			select userid from game.col_user final where 1 = 1  %s
		)
	`
	if !uFilter {
		uFilterSql3 = ""
		args3 = []any{}
	} else {
		uFilterSql3 = fmt.Sprintf(uFilterSql3, where3)
	}

	sql0 = fmt.Sprintf(sql0, where1, where2, uFilterSql3, having4)
	var datas []map[string]any
	args := append(args1, args2...)
	args = append(args, args3...)
	args = append(args, args4...)
	err = ck.Select(&datas, sql0, args...)
	if err != nil {
		return
	}

	var dateStats = make(map[uint32]*entity.TpNewStatDate)
	for _, data := range datas { // 输和赢的统计汇总
		datestr := data["datestr"].(uint32)
		stat, ok := dateStats[datestr]
		if !ok {
			stat = &entity.TpNewStatDate{
				DateStr:         fmt.Sprint(datestr),
				HuaTypeTimes:    [13]int{},
				HuaTypeWinTimes: [13]int{},
			}
			dateStats[datestr] = stat
			stats = append(stats, stat)
		}
		mappingStatDate1(data, stat)
	}

	summary := &entity.TpNewStatDate{
		DateStr: fmt.Sprintf("%v-%v总汇", params["startDate"], params["endDate"]),
	}
	// 新老顾客玩tp数量
	sqlPlayers := `
		select s1.datestr, s1.players, s1.round_avg, s1.round_median, length(arrayIntersect(s1.userids, s2.userids)) new_players, (players - new_players) old_players
		from (
			select datestr, count(*) players, avg(round) round_avg, median(round) round_median, groupArray(userid) userids
			from (
				select toYYYYMMDD(toDateTime(begin_time)) datestr, userid, count(*) round
				from game.col_detail final 
				where gtype = 1 and robot = 0 %s
				and userid in (
					select userid from (
						select userid, SUM(bet_amount) bet_total, avg(bet_amount) bet_avg, SUM(score) score_total
						from game.col_detail final 
						where gtype = 1 and robot = 0 %s
						%s
						group by userid
						having 1 = 1 %s
					)
				)
				group by datestr, userid
			) t1 
			group by datestr
		) s1
		left join (
			select toYYYYMMDD(ctime) datestr, groupArray(userid) userids
			from game.col_user final 
			where ctime between ? AND ?
			group by datestr
		) s2 on s1.datestr = s2.datestr
	`
	sqlPlayers = fmt.Sprintf(sqlPlayers, where1, where2, uFilterSql3, having4)
	start := params["start"].(time.Time)
	end := params["end"].(time.Time)
	argsPlayers := append(args, start, end)
	var playerDatas []map[string]any
	err = ck.Select(&playerDatas, sqlPlayers, argsPlayers...)
	if err != nil {
		return
	}
	for _, data := range playerDatas {
		datestr := data["datestr"].(uint32)
		if stat, ok := dateStats[datestr]; ok {
			stat.Players = int32(utils.ToInt64(data["players"]))
			stat.PlayerRoundsAvg = fmt.Sprintf("%.2f", utils.ToFloat64(data["round_avg"]))
			stat.PlayerRoundsMedian = fmt.Sprintf("%.2f", utils.ToFloat64(data["round_median"]))
			stat.NewPlayers = int32(utils.ToInt64(data["new_players"]))
			stat.OldPlayers = int32(utils.ToInt64(data["old_players"]))

			summary.Players += stat.Players
			summary.NewPlayers += stat.NewPlayers
			summary.OldPlayers += stat.OldPlayers
		}
	}

	// 新老顾客日活
	var loginDatas []map[string]any
	err = ck.Select(&loginDatas, fmt.Sprintf(`
		select s1.datestr, s1.players, s2.new_players, (players - new_players) old_players,tp_heartbeat_hr_sum,tp_heartbeat_users
		from (
			select datestr, count(*) players
			from (
				select toYYYYMMDD(login_time) datestr
				from game.col_log_login final 
				where login_time between ? AND ?
				%s
				group by datestr, userid
			) t1 
			group by datestr
		) s1
		left join (
			select toYYYYMMDD(ctime) datestr, count(*) new_players, 
				SUM(tp_heartbeat_hr) AS tp_heartbeat_hr_sum,
				SUM(CASE WHEN tp_heartbeat_hr != 0 THEN 1 ELSE 0 END) tp_heartbeat_users
			from game.col_user final 
			where ctime between ? AND ?
			%s
			group by datestr
		) s2 on s1.datestr = s2.datestr
	`, uFilterSql3, uFilterSql3), append(append([]any{start, end}, args3...), append([]any{start, end}, args3...)...)...)
	if err != nil {
		return
	}
	for _, data := range loginDatas {
		datestr := data["datestr"].(uint32)
		if stat, ok := dateStats[datestr]; ok {
			stat.LoginedUsers = int32(utils.ToInt64(data["players"]))
			stat.NewLoginedUsers = int32(utils.ToInt64(data["new_players"]))
			stat.OldLoginedUsers = int32(utils.ToInt64(data["old_players"]))
			stat.TpHeartbeatHr = utils.Float64(data["tp_heartbeat_hr_sum"])
			stat.TPNum = utils.ToInt64(data["tp_heartbeat_users"])

			summary.LoginedUsers += stat.LoginedUsers
			summary.NewLoginedUsers += stat.NewLoginedUsers
			summary.OldLoginedUsers += stat.OldLoginedUsers
			summary.TpHeartbeatHr += stat.TpHeartbeatHr
			summary.TPNum += stat.TPNum
		}
	}
	// 查询所有玩家
	for _, stat := range stats { // 汇总后二次计算
		mappingStatDate0(stat, true)
	}

	// 查询总汇数据
	sqlSummary := strings.ReplaceAll(sql0, "select toYYYYMMDD(toDateTime(begin_time)) datestr, win_type", "select win_type")
	sqlSummary = strings.ReplaceAll(sqlSummary, "group by datestr, win_type", "group by win_type")
	sqlSummary = strings.ReplaceAll(sqlSummary, "order by datestr desc", "")
	var summaryDatas []map[string]any
	err = ck.Select(&summaryDatas, sqlSummary, args...)
	if err != nil {
		return
	}
	for _, data := range summaryDatas {
		mappingStatDate1(data, summary)
	}

	mappingStatDate0(summary, false)
	// 总汇人均局数
	summary.PlayerRoundsAvg = fmt.Sprintf("%.2f", ComputeFloat(summary.AllRounds, summary.Players))
	// 局数中位数
	var summaryMedian map[string]any
	err = ck.Select(&summaryMedian, fmt.Sprintf(`
		select median(round) round_median
		from (
			select userid, count(*) round
			from game.col_detail final 
			where gtype = 1 and robot = 0 %s
			and userid in (
				select userid from (
					select userid, SUM(bet_amount) bet_total, avg(bet_amount) bet_avg, SUM(score) score_total
					from game.col_detail final 
					where gtype = 1 and robot = 0 %s
					%s
					group by userid
					having 1 = 1 %s
				)
			)
			group by userid
		) t1
	`, where1, where2, uFilterSql3, having4), args...)
	if err != nil {
		return
	}
	summary.PlayerRoundsMedian = fmt.Sprintf("%.2f", utils.ToFloat64(summaryMedian["round_median"]))

	stats = append([]*entity.TpNewStatDate{summary}, stats...)
	return
}

func mappingStatDate1(src map[string]any, dst *entity.TpNewStatDate) {
	win_type := utils.ToInt64(src["win_type"])
	dst.AllRounds += int32(utils.ToInt64(src["round"]))
	dst.PlayerBets += utils.ToInt64(src["bet_total"])
	dst.Cash += utils.ToInt64(src["score_total"])
	dst.OpponentCash += utils.ToFloat64(src["opponent_score"]) / 100 // 冤人局净赢
	dst.HurtCash += utils.ToFloat64(src["hurt_score"]) / 100         // 偷与被偷净赢
	dst.ChargeTimes += int16(utils.ToInt64(src["charge_times"]))
	dst.ChargeLaunchTimes += int16(utils.ToInt64(src["charge_launch_times"]))
	dst.ChargePayTimes += int16(utils.ToInt64(src["charge_pay_times"]))
	dst.ChangeTableCount += utils.ToInt64(src["change_table_count"])
	if win_type == 1 { // 赢
		dst.WinRounds = int32(utils.ToInt64(src["round"]))
		dst.WinBets = utils.ToInt64(src["bet_total"])
		dst.WinBetsAvg = utils.ToInt64(src["bet_avg"])
		dst.Wins = utils.ToInt64(src["score_total"])
		dst.FWinsAvg = fmt.Sprintf("%.2f", float64(utils.ToInt64(src["score_avg"]))/100)
		// 牌型赢次数
		for i := 1; i <= 12; i++ {
			dst.HuaTypeWinTimes[i] = int(utils.ToInt64(src[fmt.Sprintf("hua%d", i)]))
		}
		dst.OpponentWinRound = utils.ToInt64(src["opponent_total"])
		dst.HurtWinRound = utils.ToInt64(src["hurt_total"])
		dst.OpponentWins = utils.ToFloat64(src["opponent_score"]) / 100
		dst.HurtWins = utils.ToFloat64(src["hurt_score"]) / 100
	} else { // 输
		dst.LoseRounds = int32(utils.ToInt64(src["round"]))
		dst.LoseBets = utils.ToInt64(src["bet_total"])
		dst.LoseBetsAvg = utils.ToInt64(src["bet_avg"])
		dst.Loses = utils.ToInt64(src["score_total"])
		dst.FLosesAvg = fmt.Sprintf("%.2f", float64(utils.ToInt64(src["score_avg"]))/100)
		dst.OpponentLoseRound = utils.ToInt64(src["opponent_total"])
		dst.HurtLoseRound = utils.ToInt64(src["hurt_total"])
		dst.OpponentLoses = utils.ToFloat64(src["opponent_score"]) / 100
		dst.HurtLoses = utils.ToFloat64(src["hurt_score"]) / 100
	}
	dst.Rounds2 += utils.ToInt64(src["Rounds2"])
	dst.Rounds3 += utils.ToInt64(src["Rounds3"])
	dst.Rounds4 += utils.ToInt64(src["Rounds4"])
	dst.Rounds5 += utils.ToInt64(src["Rounds5"])
	dst.WinRounds2 += utils.ToInt64(src["WinRounds2"])
	dst.WinRounds3 += utils.ToInt64(src["WinRounds3"])
	dst.WinRounds4 += utils.ToInt64(src["WinRounds4"])
	dst.WinRounds5 += utils.ToInt64(src["WinRounds5"])
	if dst.TPNum > 0 {
		// TPNum
		dst.PlayerHeartRate = fmt.Sprintf("%.1f", float64(dst.TpHeartbeatHr)/float64(dst.TPNum)) // 玩家心率
	}
	// 拿牌型次数
	for i := 1; i <= 12; i++ {
		dst.HuaTypeTimes[i] += int(utils.ToInt64(src[fmt.Sprintf("hua%d", i)]))
	}
}

func mappingStatDate0(dst *entity.TpNewStatDate, formatDate bool) {
	if formatDate {
		datestr := dst.DateStr
		dst.DateStr = datestr[0:4] + "-" + datestr[4:6] + "-" + datestr[6:8]
	}
	dst.PlayerRate = utils.CaseElse(dst.LoginedUsers == 0, "", fmt.Sprintf("%.2f%%", ComputeFloat(dst.Players, dst.LoginedUsers)*100))
	dst.NewPlayerRate = utils.CaseElse(dst.NewLoginedUsers == 0, "", fmt.Sprintf("%.2f%%", ComputeFloat(dst.NewPlayers, dst.NewLoginedUsers)*100))
	dst.OldPlayerRate = utils.CaseElse(dst.OldLoginedUsers == 0, "", fmt.Sprintf("%.2f%%", ComputeFloat(dst.OldPlayers, dst.OldLoginedUsers)*100))

	if dst.AllRounds > 0 {
		dst.RoundsBetsAvg = fmt.Sprintf("%.2f", float64(dst.PlayerBets)/float64(dst.AllRounds)/100) // 局均打码量
	}
	if dst.AllRounds > 0 {
		dst.WinRate = fmt.Sprintf("%.2f%%", float64(dst.WinRounds)/float64(dst.AllRounds)*100) // 玩家胜率
	}
	if dst.PlayerBets != 0 {
		// dst.RewardRate = fmt.Sprintf("%.2f%%", float64(dst.Wins)/float64(-dst.Loses)*100) // 总返奖率
		dst.RewardRate = fmt.Sprintf("%.2f%%", float64(dst.Wins+dst.WinBets)/float64(dst.PlayerBets)*100) // 总返奖率
	}
	if dst.AllRounds > 0 {
		dst.FCashAvg = fmt.Sprintf("%.2f", float64(dst.Cash)/float64(dst.AllRounds)/100) // 局均净赢
	}
	if dst.Rounds2 > 0 {
		dst.WinRounds2Rate = fmt.Sprintf("%.2f%%", float64(dst.WinRounds2)/float64(dst.Rounds2)*100) // 玩家2人局胜率
	}
	if dst.Rounds3 > 0 {
		dst.WinRounds3Rate = fmt.Sprintf("%.2f%%", float64(dst.WinRounds3)/float64(dst.Rounds3)*100) // 玩家3人局胜率
	}
	if dst.Rounds4 > 0 {
		dst.WinRounds4Rate = fmt.Sprintf("%.2f%%", float64(dst.WinRounds4)/float64(dst.Rounds4)*100) // 玩家4人局胜率
	}
	if dst.Rounds5 > 0 {
		dst.WinRounds5Rate = fmt.Sprintf("%.2f%%", float64(dst.WinRounds5)/float64(dst.Rounds5)*100) // 玩家5人局胜率
	}
	if dst.ChangeTableCount > 0 {
		dst.HZFrequency = fmt.Sprintf("%.2f", float64(dst.AllRounds)/float64(dst.ChangeTableCount))
	}
	if dst.TPNum > 0 {
		// TPNum
		dst.PlayerHeartRate = fmt.Sprintf("%.1f", float64(dst.TpHeartbeatHr)/float64(dst.TPNum)) // 玩家心率
	}
	if dst.AllRounds > 0 {
		dst.Hua1Rate = utils.CaseElse(dst.HuaTypeTimes[1] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[1])/float64(dst.AllRounds)*100))    // 玩家拿大豹子率
		dst.Hua2Rate = utils.CaseElse(dst.HuaTypeTimes[2] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[2])/float64(dst.AllRounds)*100))    // 玩家拿小豹子率
		dst.Hua3Rate = utils.CaseElse(dst.HuaTypeTimes[3] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[3])/float64(dst.AllRounds)*100))    // 玩家拿大顺金率
		dst.Hua4Rate = utils.CaseElse(dst.HuaTypeTimes[4] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[4])/float64(dst.AllRounds)*100))    // 玩家拿小顺金率
		dst.Hua5Rate = utils.CaseElse(dst.HuaTypeTimes[5] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[5])/float64(dst.AllRounds)*100))    // 玩家拿大顺子率
		dst.Hua6Rate = utils.CaseElse(dst.HuaTypeTimes[6] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[6])/float64(dst.AllRounds)*100))    // 玩家拿小顺子率
		dst.Hua7Rate = utils.CaseElse(dst.HuaTypeTimes[7] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[7])/float64(dst.AllRounds)*100))    // 玩家拿大同花率
		dst.Hua8Rate = utils.CaseElse(dst.HuaTypeTimes[8] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[8])/float64(dst.AllRounds)*100))    // 玩家拿小同花率
		dst.Hua9Rate = utils.CaseElse(dst.HuaTypeTimes[9] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[9])/float64(dst.AllRounds)*100))    // 玩家拿大对子率
		dst.Hua10Rate = utils.CaseElse(dst.HuaTypeTimes[10] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[10])/float64(dst.AllRounds)*100)) // 玩家拿小对子率
		dst.Hua11Rate = utils.CaseElse(dst.HuaTypeTimes[11] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[11])/float64(dst.AllRounds)*100)) // 玩家拿大高牌率
		dst.Hua12Rate = utils.CaseElse(dst.HuaTypeTimes[12] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[12])/float64(dst.AllRounds)*100)) // 玩家拿小高牌率
	}
	dst.Hua1WinRate = utils.CaseElse(dst.HuaTypeTimes[1] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[1], dst.HuaTypeTimes[1])*100))     // 玩家拿大豹子胜率
	dst.Hua2WinRate = utils.CaseElse(dst.HuaTypeTimes[2] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[2], dst.HuaTypeTimes[2])*100))     // 玩家拿小豹子胜率
	dst.Hua3WinRate = utils.CaseElse(dst.HuaTypeTimes[3] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[3], dst.HuaTypeTimes[3])*100))     // 玩家拿大顺金胜率
	dst.Hua4WinRate = utils.CaseElse(dst.HuaTypeTimes[4] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[4], dst.HuaTypeTimes[4])*100))     // 玩家拿小顺金胜率
	dst.Hua5WinRate = utils.CaseElse(dst.HuaTypeTimes[5] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[5], dst.HuaTypeTimes[5])*100))     // 玩家拿大顺子胜率
	dst.Hua6WinRate = utils.CaseElse(dst.HuaTypeTimes[6] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[6], dst.HuaTypeTimes[6])*100))     // 玩家拿小顺子胜率
	dst.Hua7WinRate = utils.CaseElse(dst.HuaTypeTimes[7] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[7], dst.HuaTypeTimes[7])*100))     // 玩家拿大同花胜率
	dst.Hua8WinRate = utils.CaseElse(dst.HuaTypeTimes[8] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[8], dst.HuaTypeTimes[8])*100))     // 玩家拿小同花胜率
	dst.Hua9WinRate = utils.CaseElse(dst.HuaTypeTimes[9] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[9], dst.HuaTypeTimes[9])*100))     // 玩家拿大对子胜率
	dst.Hua10WinRate = utils.CaseElse(dst.HuaTypeTimes[10] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[10], dst.HuaTypeTimes[10])*100)) // 玩家拿小对子胜率
	dst.Hua11WinRate = utils.CaseElse(dst.HuaTypeTimes[11] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[11], dst.HuaTypeTimes[11])*100)) // 玩家拿大高牌胜率
	dst.Hua12WinRate = utils.CaseElse(dst.HuaTypeTimes[12] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[12], dst.HuaTypeTimes[12])*100)) // 玩家拿小高牌胜率

	dst.FBets = fmt.Sprintf("%.2f", Chip2Float(dst.PlayerBets))
	dst.RoundsBetsAvg = fmt.Sprintf("%.2f", ComputeFloat(dst.PlayerBets, int64(dst.AllRounds))/100)
	dst.PlayerBetsAvg = fmt.Sprintf("%.2f", ComputeFloat(dst.PlayerBets, int64(dst.Players))/100)
	dst.FWinBets = fmt.Sprintf("%.2f", Chip2Float(dst.WinBets))
	dst.FLoseBets = fmt.Sprintf("%.2f", Chip2Float(dst.LoseBets))
	dst.FWinBetsAvg = fmt.Sprintf("%.2f", Chip2Float(dst.WinBetsAvg))
	dst.FLoseBetsAvg = fmt.Sprintf("%.2f", Chip2Float(dst.LoseBetsAvg))
	dst.FWins = fmt.Sprintf("%.2f", Chip2Float(dst.Wins))
	dst.FLoses = fmt.Sprintf("%.2f", Chip2Float(dst.Loses))
	dst.FCash = fmt.Sprintf("%.2f", Chip2Float(dst.Cash))
	dst.FOpponentWins = fmt.Sprintf("%.2f", dst.OpponentWins)
	dst.FOpponentLoses = fmt.Sprintf("%.2f", dst.OpponentLoses)
	dst.FOpponentCash = fmt.Sprintf("%.2f", dst.OpponentCash)
	dst.FHurtWins = fmt.Sprintf("%.2f", dst.HurtWins)
	dst.FHurtLoses = fmt.Sprintf("%.2f", dst.HurtLoses)
	dst.FHurtCash = fmt.Sprintf("%.2f", dst.HurtCash)
}

/*
TP 乐极生悲
*/
func (s *gameStatsService) TpNewStatLjsb(page, pageSize int, params map[string]any) (stats []*entity.TpNewLJSBStat, total int, err error) {
	var where1, where2, where3, having4 string
	var args1, args2, args3, args4 []any

	if userid, ok := params["userid"]; ok {
		where1 += " and userid = ?"
		where2 += " and userid = ?"
		where3 += " and userid = ?"
		args1 = append(args1, userid)
		args2 = append(args2, userid)
		args3 = append(args3, userid)
	}

	// 时间
	startTime, ok1 := params["startTime"]
	endTime, ok2 := params["endTime"]
	if ok1 && ok2 {
		where1 += " and begin_time BETWEEN ? AND ?"
		where2 += " and begin_time BETWEEN ? AND ?"
		args1 = append(args1, startTime, endTime)
		args2 = append(args2, startTime, endTime)
	} else if ok1 {
		where1 += " and begin_time >= ?"
		where2 += " and begin_time >= ?"
		args1 = append(args1, startTime)
		args2 = append(args2, startTime)
	} else if ok2 {
		where1 += " and begin_time <= ?"
		where2 += " and begin_time <= ?"
		args1 = append(args1, endTime)
		args2 = append(args2, endTime)
	}

	// 打码量
	bets1, ok1 := params["bets1"]
	bets2, ok2 := params["bets2"]
	if ok1 && ok2 {
		having4 += " and bet_total between ? and ?"
		args4 = append(args4, bets1, bets2)
	}
	// 注均码
	betAvg1, ok1 := params["betAvg1"]
	betAvg2, ok2 := params["betAvg2"]
	if ok1 && ok2 {
		having4 += " and bet_avg between ? and ?"
		args4 = append(args4, betAvg1, betAvg2)
	}

	// 净赢
	cash1, ok1 := params["cash1"]
	cash2, ok2 := params["cash2"]
	if ok1 && ok2 {
		having4 += " and score_total between ? and ?"
		args4 = append(args4, cash1, cash2)
	}

	// 流失天数
	var uFilter bool
	uLossDays1, ok1 := params["uLossDays1"]
	uLossDays2, ok2 := params["uLossDays2"]
	uFilter = ok1 || ok2
	if ok1 && ok2 {
		where3 += " and date_diff('day', login_time, now()) between ? and ?"
		args3 = append(args3, uLossDays1, uLossDays2)
	}
	// money
	utypeQ, ok1 := params["utypeQ"]
	if ok1 {
		where3 += utypeQ.(string)
		uFilter = true
	}

	// 渠道
	if ad__bundle_id, ok := params["ad__bundle_id"]; ok {
		where3 += " and ad__bundle_id in ?"
		args3 = append(args3, ad__bundle_id)
		uFilter = true
	}
	// ab 类
	if regist_area, ok := params["regist_area"]; ok {
		where3 += " and regist_area in ?"
		args3 = append(args3, regist_area)
		uFilter = true
	}

	sql0 := `
		select userid, win_type, count(*) round, sum(bet_amount) bet_total, avg(bet_amount) bet_avg, sum(score) score_total, avg(score) score_avg, 
		SUM(case when tp_ljsb_active = 1 then 1 else 0 end) s_round, 
		SUM(case when tp_ljsb_active = 1 then bet_amount else 0 end) s_bet_total,
		SUM(case when tp_ljsb_active = 1 and win_type = 1 then 1 else 0 end) s_win_round,
		SUM(case when tp_ljsb_active = 1 and win_type = 1 then score else 0 end) s_win_bets,
		SUM(case when tp_ljsb_active = 1 and win_type = 2 then score else 0 end) s_lose_bets,
		SUM(case when tp_ljsb_active = 1 and tp_ljsb_py = 1 then 1 else 0 end) py_round,
		SUM(case when tp_ljsb_active = 1 and tp_ljsb_ps = 1 then 1 else 0 end) ps_round,
		SUM(case when tp_ljsb_active = 1 and tp_ljsb_pd = 1 then 1 else 0 end) pd_round,
		SUM(case when tp_ljsb_active = 1 and tp_ljsb_pt = 1 then 1 else 0 end) pt_round,
		SUM(case when tp_hand_type_up_down = 10 and tp_ljsb_active = 1 then 1 else 0 end) hua1,
		SUM(case when tp_hand_type_up_down = 11 and tp_ljsb_active = 1 then 1 else 0 end) hua2,
		SUM(case when tp_hand_type_up_down = 20 and tp_ljsb_active = 1 then 1 else 0 end) hua3,
		SUM(case when tp_hand_type_up_down = 21 and tp_ljsb_active = 1 then 1 else 0 end) hua4,
		SUM(case when tp_hand_type_up_down = 30 and tp_ljsb_active = 1 then 1 else 0 end) hua5,
		SUM(case when tp_hand_type_up_down = 31 and tp_ljsb_active = 1 then 1 else 0 end) hua6,
		SUM(case when tp_hand_type_up_down = 40 and tp_ljsb_active = 1 then 1 else 0 end) hua7,
		SUM(case when tp_hand_type_up_down = 41 and tp_ljsb_active = 1 then 1 else 0 end) hua8,
		SUM(case when tp_hand_type_up_down = 50 and tp_ljsb_active = 1 then 1 else 0 end) hua9,
		SUM(case when tp_hand_type_up_down = 51 and tp_ljsb_active = 1 then 1 else 0 end) hua10,
		SUM(case when tp_hand_type_up_down = 60 and tp_ljsb_active = 1 then 1 else 0 end) hua11,
		SUM(case when tp_hand_type_up_down = 61 and tp_ljsb_active = 1 then 1 else 0 end) hua12,
		SUM(case when tp_ljsb_active = 1 then tp_opponent else 0 end) s_opponent_total,	
		SUM(tp_opponent) opponent_total, SUM(tp_hurt) hurt_total,
			SUM(case when tp_opponent = 1 then score else 0 end) opponent_score,
			SUM(case when tp_hurt = 1 then score else 0 end) hurt_score,
			SUM(charge_times) charge_times, 
			SUM(charge_launch_times) charge_launch_times, 
			SUM(charge_pay_times) charge_pay_times,max (tp_ljsb_jl_type) tp_ljsb_jl_type
		from game.col_detail final 
		where gtype = 1 and robot = 0 %s
		and userid in (
			select userid from (
				select userid, SUM(bet_amount) bet_total, avg(bet_amount) bet_avg, SUM(score) score_total, max(begin_time) lately_time,
					max(tp_ljsb_active) tp_ljsb_active
				from game.col_detail final 
				where gtype = 1 and robot = 0 %s
				%s
				group by userid
				having tp_ljsb_active = 1 %s
			)
			order by lately_time desc limit ?, ?
		)
		group by userid, win_type
		order by max(begin_time) desc
	`
	uFilterSql3 := `
		and userid in (
			select userid from game.col_user final where 1 = 1  %s
		)
	`
	if !uFilter {
		uFilterSql3 = ""
		args3 = []any{}
	} else {
		uFilterSql3 = fmt.Sprintf(uFilterSql3, where3)
	}

	// 查总条数 , max(begin_time) lately_time
	err = ck.Select(&total, fmt.Sprintf(`
		select count(*) from (
			select userid, SUM(bet_amount) bet_total, avg(bet_amount) bet_avg, SUM(score) score_total,
				max(tp_ljsb_active) tp_ljsb_active
			from game.col_detail final 
			where gtype = 1 and robot = 0 %s
			%s
			group by userid
			having tp_ljsb_active = 1 %s
		)
	`, where2, uFilterSql3, having4), append(append(args2, args3...), args4...)...)
	if err != nil {
		return
	}

	sql0 = fmt.Sprintf(sql0, where1, where2, uFilterSql3, having4)
	var datas []map[string]any
	args := append(args1, args2...)
	args = append(args, args3...)
	args = append(args, args4...)
	offset, limit := PageCalc(page, pageSize)
	// args = append(args, offset, limit)
	err = ck.Select(&datas, sql0, append(args, offset, limit)...)
	if err != nil {
		return
	}

	var uidStats = make(map[string]*entity.TpNewLJSBStat)
	var userids []string
	var no int
	for _, data := range datas { // 输和赢的统计汇总
		userid := data["userid"].(string)
		stat, ok := uidStats[userid]
		if !ok {
			no++
			stat = &entity.TpNewLJSBStat{
				No:              strconv.Itoa(no),
				UserId:          userid,
				HuaTypeTimes:    [13]int{},
				HuaTypeWinTimes: [13]int{},
			}
			uidStats[userid] = stat
			stats = append(stats, stat)
			userids = append(userids, userid)
		}
		mappingStatLjsb1(data, stat)
	}

	// 牌型输赢
	var huaTypeDatas []map[string]any
	sqlHandType := fmt.Sprintf(`
		select userid, win_type, tp_hand_type_up_down, SUM(score) score_total, AVG(score) score_avg, count(*) round, MAX(score) score_max, MIN(score) score_min
		from game.col_detail final 
		where gtype = 1 and robot = 0 %s
			and userid in ?
		group by userid, win_type, tp_hand_type_up_down
	`, where2)
	argsHandType := append(args2, userids)
	err = ck.Select(&huaTypeDatas, sqlHandType, argsHandType...)
	if err != nil {
		return
	}
	for _, data := range huaTypeDatas {
		userid := data["userid"].(string)
		stat, ok := uidStats[userid]
		if !ok {
			continue
		}
		mappingStatLjsb2HuaType(data, stat)
	}

	for _, stat := range stats { // 汇总后二次计算
		mappingStatLjsb0(stat)
	}
	var totalHeart float64
	// 查用户信息
	var users []map[string]any
	ck.Select(&users, `
		select userid,nickname,ctime,login_time,money,state,regist_area,
			date_diff('day', ctime, now()) live_days, date_diff('day', login_time, now()) loss_days,tp_heartbeat_hr
		from game.col_user final where userid in ?
	`, userids)
	for _, user := range users {
		userid := user["userid"].(string)
		nickname := user["nickname"].(string)
		ctime := user["ctime"].(time.Time)
		login_time := user["login_time"].(time.Time)
		money := utils.ToInt64(user["money"])
		live_days := utils.ToInt64(user["live_days"])
		loss_days := utils.ToInt64(user["loss_days"])
		state := utils.ToInt64(user["state"])
		regist_area := utils.ToInt64(user["regist_area"])
		tp_heartbeat_hr := utils.ToFloat64(user["tp_heartbeat_hr"])

		stat, ok := uidStats[userid]
		if ok {
			stat.Nickname = nickname
			stat.Ctime = ctime.Format(utils.FORMAT)
			stat.LoginTime = login_time.Format(utils.FORMAT)
			stat.LiveDays = int(live_days)
			stat.LoseDays = int(loss_days)
			stat.TpHeartbeatHr = tp_heartbeat_hr
			if tp_heartbeat_hr > 0 {
				totalHeart += tp_heartbeat_hr
				stat.PlayerHeartRate = fmt.Sprintf("%.1f", tp_heartbeat_hr)
			}
			switch regist_area {
			case 0, 3:
				stat.RegistArea = "A类"
			case 1:
				stat.RegistArea = "B类"
			case 2:
				stat.RegistArea = "C类"
			}
			switch state {
			case 1:
				stat.UserType = "新手"
			case 2:
				if money == 0 {
					stat.UserType = "零充"
				} else if money >= 20000 && money <= 99999 {
					stat.UserType = "普充"
				} else if money >= 100000 && money <= 499999 {
					stat.UserType = "小R"
				} else if money >= 500000 && money <= 999999 {
					stat.UserType = "中R"
				} else if money >= 1000000 && money <= 9999999 {
					stat.UserType = "大R"
				} else if money >= 10000000 {
					stat.UserType = "超大R"
				}
			case 3:
				stat.UserType = "平民"
			case 4:
				stat.UserType = "泡沫"
			}
		}
	}

	// 查总汇数据
	mark := "--"
	summary := &entity.TpNewLJSBStat{
		No:            fmt.Sprintf("%v-%v总汇", params["startDate"], params["endDate"]),
		UserId:        mark,
		Nickname:      mark,
		Ctime:         mark,
		LoginTime:     mark,
		UserType:      mark,
		RegistArea:    mark,
		TpHeartbeatHr: totalHeart,
		TPNum:         int64(len(userids)),
	}
	sqlSummary := strings.ReplaceAll(sql0, "select userid, win_type", "select win_type")
	sqlSummary = strings.ReplaceAll(sqlSummary, "limit ?, ?", "")
	sqlSummary = strings.ReplaceAll(sqlSummary, "group by userid, win_type", "group by win_type")
	var summaryDatas []map[string]any
	err = ck.Select(&summaryDatas, sqlSummary, args...)
	if err != nil {
		return
	}
	for _, data := range summaryDatas {
		mappingStatLjsb1(data, summary)
		// 替换汇总得极乐状态
		if summary.JlStatus != "" {
			summary.JlStatus = "--"
		}
	}

	// 牌型输赢
	var summaryHuaTypeDatas []map[string]any
	sqlSummaryHandType := strings.ReplaceAll(sqlHandType, "select userid, win_type", "select win_type")
	sqlSummaryHandType = strings.ReplaceAll(sqlSummaryHandType, "group by userid, win_type", "group by win_type")
	err = ck.Select(&summaryHuaTypeDatas, sqlSummaryHandType, argsHandType...)
	if err != nil {
		return
	}
	for _, data := range summaryHuaTypeDatas {
		mappingStatLjsb2HuaType(data, summary)
	}
	// 汇总后二次计算
	mappingStatLjsb0(summary)
	stats = append([]*entity.TpNewLJSBStat{summary}, stats...)
	return
}

func mappingStatLjsb1(src map[string]any, dst *entity.TpNewLJSBStat) {
	win_type := utils.ToInt64(src["win_type"])
	dst.AllRounds += int32(utils.ToInt64(src["round"]))
	dst.Bets += utils.ToInt64(src["bet_total"])
	dst.Cash += utils.ToInt64(src["score_total"])
	dst.OpponentCash += utils.ToFloat64(src["opponent_score"]) / 100 // 冤人局净赢
	// dst.HurtCash += utils.ToFloat64(src["hurt_score"]) / 100                  // 偷与被偷净赢
	// dst.ChargeTimes += int16(utils.ToInt64(src["charge_times"]))              // 充值触发次数
	// dst.ChargeLaunchTimes += int16(utils.ToInt64(src["charge_launch_times"])) // 局内充值拉单次数
	// dst.ChargePayTimes += int16(utils.ToInt64(src["charge_pay_times"]))       // 局内充值实付次数
	jltype := utils.ToInt64(src["tp_ljsb_jl_type"])
	switch jltype {
	case 1:
		dst.JlStatus = "超率极乐"
	case 2:
		dst.JlStatus = "超利极乐"
	case 3:
		dst.JlStatus = "利率极乐"
	}
	// 策略
	dst.StrategyRounds += utils.ToInt64(src["s_round"])
	dst.StrategyBets += utils.ToInt64(src["s_bet_total"])
	dst.ByRounds += utils.ToInt64(src["py_round"])
	dst.YsRounds += utils.ToInt64(src["ps_round"])
	dst.YdRounds += utils.ToInt64(src["pd_round"])
	dst.YtRounds += utils.ToInt64(src["pt_round"])
	dst.StrategyWinRounds += utils.ToInt64(src["s_win_round"])
	dst.StrategyWinBets += utils.ToInt64(src["s_win_bets"])
	dst.StrategyLoseBets += utils.ToInt64(src["s_lose_bets"])
	dst.OpponentRound += utils.ToInt64(src["s_opponent_total"])
	if win_type == 1 { // 赢
		dst.WinRounds = int32(utils.ToInt64(src["round"]))
		dst.WinBets = utils.ToInt64(src["bet_total"])
		dst.WinBetsAvg = utils.ToInt64(src["bet_avg"])
		dst.Wins = utils.ToInt64(src["score_total"])
		// dst.FWinsAvg = fmt.Sprintf("%.2f", float64(utils.ToInt64(src["score_avg"]))/100)
		// 牌型赢次数
		for i := 1; i <= 12; i++ {
			dst.HuaTypeWinTimes[i] = int(utils.ToInt64(src[fmt.Sprintf("hua%d", i)]))
		}
		dst.OpponentWinRound = utils.ToInt64(src["opponent_total"])
		// dst.HurtWinRound = utils.ToInt64(src["hurt_total"])
		dst.OpponentWins = utils.ToFloat64(src["opponent_score"]) / 100
		// dst.HurtWins = utils.ToFloat64(src["hurt_score"]) / 100
	} else { // 输
		dst.LoseRounds = int32(utils.ToInt64(src["round"]))
		dst.LoseBets = utils.ToInt64(src["bet_total"])
		dst.LoseBetsAvg = utils.ToInt64(src["bet_avg"])
		dst.Loses = utils.ToInt64(src["score_total"])
		// dst.FLosesAvg = fmt.Sprintf("%.2f", float64(utils.ToInt64(src["score_avg"]))/100)
		dst.OpponentLoseRound = utils.ToInt64(src["opponent_total"])
		// dst.HurtLoseRound = utils.ToInt64(src["hurt_total"])
		dst.OpponentLoses = utils.ToFloat64(src["opponent_score"]) / 100
		// dst.HurtLoses = utils.ToFloat64(src["hurt_score"]) / 100
	}
	// dst.Rounds2 += utils.ToInt64(src["Rounds2"])
	// dst.Rounds3 += utils.ToInt64(src["Rounds3"])
	// dst.Rounds4 += utils.ToInt64(src["Rounds4"])
	// dst.Rounds5 += utils.ToInt64(src["Rounds5"])
	// dst.WinRounds2 += utils.ToInt64(src["WinRounds2"])
	// dst.WinRounds3 += utils.ToInt64(src["WinRounds3"])
	// dst.WinRounds4 += utils.ToInt64(src["WinRounds4"])
	// dst.WinRounds5 += utils.ToInt64(src["WinRounds5"])
	if dst.TPNum > 0 {
		// TPNum
		dst.PlayerHeartRate = fmt.Sprintf("%.1f", float64(dst.TpHeartbeatHr)/float64(dst.TPNum)) // 玩家心率
	}
	// 拿牌型次数
	for i := 1; i <= 12; i++ {
		dst.HuaTypeTimes[i] += int(utils.ToInt64(src[fmt.Sprintf("hua%d", i)]))
	}
}

func mappingStatLjsb0(dst *entity.TpNewLJSBStat) {
	if dst.AllRounds > 0 {
		dst.BetsAvg = fmt.Sprintf("%.2f", float64(dst.Bets)/float64(dst.AllRounds)/100) // 局均打码量
	}
	if dst.AllRounds > 0 {
		dst.WinRate = fmt.Sprintf("%.2f%%", float64(dst.WinRounds)/float64(dst.AllRounds)*100) // 玩家胜率
	}
	if dst.Loses != 0 {
		dst.RewardRate = fmt.Sprintf("%.2f%%", float64(dst.Wins)/float64(-dst.Loses)*100) // 总返奖率
	}
	if dst.StrategyRounds > 0 {
		dst.StrategyBetsAvg = fmt.Sprintf("%.2f", Chip2Float(dst.StrategyBets)/float64(dst.StrategyRounds)/100)     // 策略局局均打码量
		dst.StrategyWinRate = fmt.Sprintf("%.2f%%", float64(dst.StrategyWinRounds)/float64(dst.StrategyRounds)*100) // 策略中胜率
	}
	if dst.StrategyLoseBets != 0 {
		dst.StrategyRewardRate = fmt.Sprintf("%.2f%%", float64(dst.StrategyWinBets)/float64(-dst.StrategyLoseBets)*100) // 策略中返奖率
	}
	if dst.OpponentRound > 0 {
		dst.OpponentWinRate = fmt.Sprintf("%.2f%%", float64(dst.OpponentWinRound)/float64(dst.OpponentRound)*100) // 策略中冤家局胜率
		dst.ByRate = fmt.Sprintf("%.2f%%", float64(dst.ByRounds)/float64(dst.OpponentRound)*100)
		dst.YsRate = fmt.Sprintf("%.2f%%", float64(dst.YsRounds)/float64(dst.OpponentRound)*100)
		dst.YdRate = fmt.Sprintf("%.2f%%", float64(dst.YdRounds)/float64(dst.OpponentRound)*100)
		dst.YtRate = fmt.Sprintf("%.2f%%", float64(dst.YtRounds)/float64(dst.OpponentRound)*100)
	}
	if dst.StrategyRounds > 0 {
		dst.Hua1Rate = utils.CaseElse(dst.HuaTypeTimes[1] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[1])/float64(dst.StrategyRounds)*100))    // 玩家拿大豹子率
		dst.Hua2Rate = utils.CaseElse(dst.HuaTypeTimes[2] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[2])/float64(dst.StrategyRounds)*100))    // 玩家拿小豹子率
		dst.Hua3Rate = utils.CaseElse(dst.HuaTypeTimes[3] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[3])/float64(dst.StrategyRounds)*100))    // 玩家拿大顺金率
		dst.Hua4Rate = utils.CaseElse(dst.HuaTypeTimes[4] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[4])/float64(dst.StrategyRounds)*100))    // 玩家拿小顺金率
		dst.Hua5Rate = utils.CaseElse(dst.HuaTypeTimes[5] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[5])/float64(dst.StrategyRounds)*100))    // 玩家拿大顺子率
		dst.Hua6Rate = utils.CaseElse(dst.HuaTypeTimes[6] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[6])/float64(dst.StrategyRounds)*100))    // 玩家拿小顺子率
		dst.Hua7Rate = utils.CaseElse(dst.HuaTypeTimes[7] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[7])/float64(dst.StrategyRounds)*100))    // 玩家拿大同花率
		dst.Hua8Rate = utils.CaseElse(dst.HuaTypeTimes[8] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[8])/float64(dst.StrategyRounds)*100))    // 玩家拿小同花率
		dst.Hua9Rate = utils.CaseElse(dst.HuaTypeTimes[9] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[9])/float64(dst.StrategyRounds)*100))    // 玩家拿大对子率
		dst.Hua10Rate = utils.CaseElse(dst.HuaTypeTimes[10] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[10])/float64(dst.StrategyRounds)*100)) // 玩家拿小对子率
		dst.Hua11Rate = utils.CaseElse(dst.HuaTypeTimes[11] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[11])/float64(dst.StrategyRounds)*100)) // 玩家拿大高牌率
		dst.Hua12Rate = utils.CaseElse(dst.HuaTypeTimes[12] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[12])/float64(dst.StrategyRounds)*100)) // 玩家拿小高牌率
	}
	dst.Hua1WinRate = utils.CaseElse(dst.HuaTypeTimes[1] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[1], dst.HuaTypeTimes[1])*100))     // 玩家拿大豹子胜率
	dst.Hua2WinRate = utils.CaseElse(dst.HuaTypeTimes[2] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[2], dst.HuaTypeTimes[2])*100))     // 玩家拿小豹子胜率
	dst.Hua3WinRate = utils.CaseElse(dst.HuaTypeTimes[3] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[3], dst.HuaTypeTimes[3])*100))     // 玩家拿大顺金胜率
	dst.Hua4WinRate = utils.CaseElse(dst.HuaTypeTimes[4] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[4], dst.HuaTypeTimes[4])*100))     // 玩家拿小顺金胜率
	dst.Hua5WinRate = utils.CaseElse(dst.HuaTypeTimes[5] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[5], dst.HuaTypeTimes[5])*100))     // 玩家拿大顺子胜率
	dst.Hua6WinRate = utils.CaseElse(dst.HuaTypeTimes[6] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[6], dst.HuaTypeTimes[6])*100))     // 玩家拿小顺子胜率
	dst.Hua7WinRate = utils.CaseElse(dst.HuaTypeTimes[7] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[7], dst.HuaTypeTimes[7])*100))     // 玩家拿大同花胜率
	dst.Hua8WinRate = utils.CaseElse(dst.HuaTypeTimes[8] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[8], dst.HuaTypeTimes[8])*100))     // 玩家拿小同花胜率
	dst.Hua9WinRate = utils.CaseElse(dst.HuaTypeTimes[9] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[9], dst.HuaTypeTimes[9])*100))     // 玩家拿大对子胜率
	dst.Hua10WinRate = utils.CaseElse(dst.HuaTypeTimes[10] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[10], dst.HuaTypeTimes[10])*100)) // 玩家拿小对子胜率
	dst.Hua11WinRate = utils.CaseElse(dst.HuaTypeTimes[11] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[11], dst.HuaTypeTimes[11])*100)) // 玩家拿大高牌胜率
	dst.Hua12WinRate = utils.CaseElse(dst.HuaTypeTimes[12] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[12], dst.HuaTypeTimes[12])*100)) // 玩家拿小高牌胜率

	dst.FBets = fmt.Sprintf("%.2f", Chip2Float(dst.Bets))
	dst.FWinBets = fmt.Sprintf("%.2f", Chip2Float(dst.WinBets))
	dst.FLoseBets = fmt.Sprintf("%.2f", Chip2Float(dst.LoseBets))
	dst.FWinBetsAvg = fmt.Sprintf("%.2f", Chip2Float(dst.WinBetsAvg))
	dst.FLoseBetsAvg = fmt.Sprintf("%.2f", Chip2Float(dst.LoseBetsAvg))
	dst.FWins = fmt.Sprintf("%.2f", Chip2Float(dst.Wins))
	dst.FLoses = fmt.Sprintf("%.2f", Chip2Float(dst.Loses))
	dst.FCash = fmt.Sprintf("%.2f", Chip2Float(dst.Cash))
	dst.FOpponentWins = fmt.Sprintf("%.2f", dst.OpponentWins)
	dst.FOpponentLoses = fmt.Sprintf("%.2f", dst.OpponentLoses)
	dst.FOpponentCash = fmt.Sprintf("%.2f", dst.OpponentCash)

	// 策略
	dst.FStrategyBets = fmt.Sprintf("%.2f", Chip2Float(dst.StrategyBets))
	dst.FStrategyWinBets = fmt.Sprintf("%.2f", Chip2Float(dst.StrategyWinBets))
	dst.FStrategyLoseBets = fmt.Sprintf("%.2f", Chip2Float(dst.StrategyLoseBets))
	if dst.StrategyRounds > 0 {
		dst.FStrategyWinBetAvg = fmt.Sprintf("%.2f", float64(dst.StrategyWinBets)/float64(dst.StrategyRounds)/100)
		dst.FStrategyLoseBetAvg = fmt.Sprintf("%.2f", float64(dst.StrategyLoseBets)/float64(dst.StrategyRounds)/100)
	}

	dst.FStrategyWins = fmt.Sprintf("%.2f", Chip2Float(dst.Wins))
	dst.FStrategyLoses = fmt.Sprintf("%.2f", Chip2Float(dst.Loses))
	dst.FStrategyCash = fmt.Sprintf("%.2f", Chip2Float(dst.Cash))

	// dst.FHurtWins = fmt.Sprintf("%.2f", dst.HurtWins)
	// dst.FHurtLoses = fmt.Sprintf("%.2f", dst.HurtLoses)
	// dst.FHurtCash = fmt.Sprintf("%.2f", dst.HurtCash)
	// dst.FMaxWinHuaTypeAmount = fmt.Sprintf("%.2f", Chip2Float(dst.MaxWinHuaTypeAmount))
	// dst.FMaxLoseHuaTypeAmount = fmt.Sprintf("%.2f", Chip2Float(dst.MaxLoseHuaTypeAmount))
	// dst.FMaxWinHuaTypeAmountAvg = fmt.Sprintf("%.2f", Chip2Float(dst.MaxWinHuaTypeAmountAvg))
	// dst.FMaxLoseHuaTypeAmountAvg = fmt.Sprintf("%.2f", Chip2Float(dst.MaxLoseHuaTypeAmountAvg))
	// dst.FOnceMaxWinHuaTypeAmount = fmt.Sprintf("%.2f", Chip2Float(dst.OnceMaxWinHuaTypeAmount))
	// dst.FOnceMaxLoseHuaTypeAmount = fmt.Sprintf("%.2f", Chip2Float(dst.OnceMaxLoseHuaTypeAmount))
}

// 牌型输赢 mapping
func mappingStatLjsb2HuaType(src map[string]any, dst *entity.TpNewLJSBStat) {
	win_type := utils.ToInt64(src["win_type"])
	// tp_hand_type_up_down := utils.ToInt64(src["tp_hand_type_up_down"])
	// round := utils.ToInt64(src["round"])
	// score_total := utils.ToInt64(src["score_total"])
	// score_avg := utils.ToInt64(src["score_avg"])
	// score_max := utils.ToInt64(src["score_max"])
	// score_min := utils.ToInt64(src["score_min"])
	if win_type == 1 {
		// if score_total > dst.MaxWinHuaTypeAmount {
		// 	dst.MaxWinHuaTypeAmount = score_total
		// 	dst.MaxWinHuaType = HuaTypeNames[int(tp_hand_type_up_down)]
		// 	dst.MaxWinHuaTypeRound = round
		// 	dst.MaxWinHuaTypeAmountAvg = score_avg
		// }
		// if score_max > dst.OnceMaxWinHuaTypeAmount {
		// 	dst.OnceMaxWinHuaTypeAmount = score_max
		// 	dst.OnceMaxWinHuaType = HuaTypeNames[int(tp_hand_type_up_down)]
		// }
	} else {
		// if score_total < dst.MaxLoseHuaTypeAmount {
		// 	dst.MaxLoseHuaTypeAmount = score_total
		// 	dst.MaxLoseHuaType = HuaTypeNames[int(tp_hand_type_up_down)]
		// 	dst.MaxLoseHuaTypeRound = round
		// 	dst.MaxLoseHuaTypeAmountAvg = score_avg
		// }
		// if score_min < dst.OnceMaxLoseHuaTypeAmount {
		// 	dst.OnceMaxLoseHuaTypeAmount = score_min
		// 	dst.OnceMaxLoseHuaType = HuaTypeNames[int(tp_hand_type_up_down)]
		// }
	}
}

/*
TP 高潮涌现
*/
func (s *gameStatsService) TpNewStatGCYX(page, pageSize int, params map[string]any) (stats []*entity.TpNewGCYXStat, total int, err error) {
	var where1, where2, where3, having4 string
	var args1, args2, args3, args4 []any

	if userid, ok := params["userid"]; ok {
		where1 += " and userid = ?"
		where2 += " and userid = ?"
		where3 += " and userid = ?"
		args1 = append(args1, userid)
		args2 = append(args2, userid)
		args3 = append(args3, userid)
	}

	// 时间
	startTime, ok1 := params["startTime"]
	endTime, ok2 := params["endTime"]
	if ok1 && ok2 {
		where1 += " and begin_time BETWEEN ? AND ?"
		where2 += " and begin_time BETWEEN ? AND ?"
		args1 = append(args1, startTime, endTime)
		args2 = append(args2, startTime, endTime)
	} else if ok1 {
		where1 += " and begin_time >= ?"
		where2 += " and begin_time >= ?"
		args1 = append(args1, startTime)
		args2 = append(args2, startTime)
	} else if ok2 {
		where1 += " and begin_time <= ?"
		where2 += " and begin_time <= ?"
		args1 = append(args1, endTime)
		args2 = append(args2, endTime)
	}

	// 打码量
	bets1, ok1 := params["bets1"]
	bets2, ok2 := params["bets2"]
	if ok1 && ok2 {
		having4 += " and bet_total between ? and ?"
		args4 = append(args4, bets1, bets2)
	}
	// 注均码
	betAvg1, ok1 := params["betAvg1"]
	betAvg2, ok2 := params["betAvg2"]
	if ok1 && ok2 {
		having4 += " and bet_avg between ? and ?"
		args4 = append(args4, betAvg1, betAvg2)
	}

	// 净赢
	cash1, ok1 := params["cash1"]
	cash2, ok2 := params["cash2"]
	if ok1 && ok2 {
		having4 += " and score_total between ? and ?"
		args4 = append(args4, cash1, cash2)
	}

	// 流失天数
	var uFilter bool
	uLossDays1, ok1 := params["uLossDays1"]
	uLossDays2, ok2 := params["uLossDays2"]
	uFilter = ok1 || ok2
	if ok1 && ok2 {
		where3 += " and date_diff('day', login_time, now()) between ? and ?"
		args3 = append(args3, uLossDays1, uLossDays2)
	}
	// money
	utypeQ, ok1 := params["utypeQ"]
	if ok1 {
		where3 += utypeQ.(string)
		uFilter = true
	}

	// 渠道
	if ad__bundle_id, ok := params["ad__bundle_id"]; ok {
		where3 += " and ad__bundle_id in ?"
		args3 = append(args3, ad__bundle_id)
		uFilter = true
	}
	// ab 类
	if regist_area, ok := params["regist_area"]; ok {
		where3 += " and regist_area in ?"
		args3 = append(args3, regist_area)
		uFilter = true
	}

	sql0 := `
		select userid, win_type, count(*) round, sum(bet_amount) bet_total, avg(bet_amount) bet_avg, sum(score) score_total, avg(score) score_avg, 
		SUM(case when tp_gcyx_active = 1 then 1 else 0 end) s_round, 
		SUM(case when tp_gcyx_active = 1 then bet_amount else 0 end) s_bet_total,
		SUM(case when tp_gcyx_active = 1 then score else 0 end) s_score,
		SUM(case when tp_gcyx_active = 1 and tp_gcyx_rp = 1 then 1 else 0 end) rp_round,
		SUM(case when tp_gcyx_active = 1 and tp_gcyx_bp = 1 then 1 else 0 end) bp_round,
		SUM(case when tp_hand_type_up_down = 10 and tp_gcyx_active = 1 then 1 else 0 end) hua1,
		SUM(case when tp_hand_type_up_down = 11 and tp_gcyx_active = 1 then 1 else 0 end) hua2,
		SUM(case when tp_hand_type_up_down = 20 and tp_gcyx_active = 1 then 1 else 0 end) hua3,
		SUM(case when tp_hand_type_up_down = 21 and tp_gcyx_active = 1 then 1 else 0 end) hua4,
		SUM(case when tp_hand_type_up_down = 30 and tp_gcyx_active = 1 then 1 else 0 end) hua5,
		SUM(case when tp_hand_type_up_down = 31 and tp_gcyx_active = 1 then 1 else 0 end) hua6,
		SUM(case when tp_hand_type_up_down = 40 and tp_gcyx_active = 1 then 1 else 0 end) hua7,
		SUM(case when tp_hand_type_up_down = 41 and tp_gcyx_active = 1 then 1 else 0 end) hua8,
		SUM(case when tp_hand_type_up_down = 50 and tp_gcyx_active = 1 then 1 else 0 end) hua9,
		SUM(case when tp_hand_type_up_down = 51 and tp_gcyx_active = 1 then 1 else 0 end) hua10,
		SUM(case when tp_hand_type_up_down = 60 and tp_gcyx_active = 1 then 1 else 0 end) hua11,
		SUM(case when tp_hand_type_up_down = 61 and tp_gcyx_active = 1 then 1 else 0 end) hua12,
			SUM(case when tp_gcyx_active = 1 then charge_times else 0 end) charge_times, 
			SUM(case when tp_gcyx_active = 1 then charge_launch_times else 0 end) charge_launch_times, 
			SUM(case when tp_gcyx_active = 1 then charge_pay_times else 0 end) charge_pay_times
		from game.col_detail final 
		where gtype = 1 and robot = 0 %s
		and userid in (
			select userid from (
				select userid, SUM(bet_amount) bet_total, avg(bet_amount) bet_avg, SUM(score) score_total, max(begin_time) lately_time,
					max(tp_gcyx_active) tp_gcyx_active
				from game.col_detail final 
				where gtype = 1 and robot = 0 %s
				%s
				group by userid
				having tp_gcyx_active = 1 %s
			)
			order by lately_time desc limit ?, ?
		)
		group by userid, win_type
		order by max(begin_time) desc
	`
	uFilterSql3 := `
		and userid in (
			select userid from game.col_user final where 1 = 1  %s
		)
	`
	if !uFilter {
		uFilterSql3 = ""
		args3 = []any{}
	} else {
		uFilterSql3 = fmt.Sprintf(uFilterSql3, where3)
	}

	// 查总条数 , max(begin_time) lately_time
	err = ck.Select(&total, fmt.Sprintf(`
		select count(*) from (
			select userid, SUM(bet_amount) bet_total, avg(bet_amount) bet_avg, SUM(score) score_total,
				max(tp_gcyx_active) tp_gcyx_active
			from game.col_detail final 
			where gtype = 1 and robot = 0 %s
			%s
			group by userid
			having tp_gcyx_active = 1 %s
		)
	`, where2, uFilterSql3, having4), append(append(args2, args3...), args4...)...)
	if err != nil {
		return
	}

	sql0 = fmt.Sprintf(sql0, where1, where2, uFilterSql3, having4)
	var datas []map[string]any
	args := append(args1, args2...)
	args = append(args, args3...)
	args = append(args, args4...)
	offset, limit := PageCalc(page, pageSize)
	// args = append(args, offset, limit)
	err = ck.Select(&datas, sql0, append(args, offset, limit)...)
	if err != nil {
		return
	}

	var uidStats = make(map[string]*entity.TpNewGCYXStat)
	var userids []string
	var no int
	for _, data := range datas { // 输和赢的统计汇总
		userid := data["userid"].(string)
		stat, ok := uidStats[userid]
		if !ok {
			no++
			stat = &entity.TpNewGCYXStat{
				No:              strconv.Itoa(no),
				UserId:          userid,
				HuaTypeTimes:    [13]int{},
				HuaTypeWinTimes: [13]int{},
			}
			uidStats[userid] = stat
			stats = append(stats, stat)
			userids = append(userids, userid)
		}
		mappingStatGcyx1(data, stat)
	}

	// 牌型输赢
	var huaTypeDatas []map[string]any
	sqlHandType := fmt.Sprintf(`
		select userid, win_type, tp_hand_type_up_down, SUM(score) score_total, AVG(score) score_avg, count(*) round, MAX(score) score_max, MIN(score) score_min
		from game.col_detail final 
		where gtype = 1 and robot = 0 %s
			and userid in ?
		group by userid, win_type, tp_hand_type_up_down
	`, where2)
	argsHandType := append(args2, userids)
	err = ck.Select(&huaTypeDatas, sqlHandType, argsHandType...)
	if err != nil {
		return
	}
	for _, data := range huaTypeDatas {
		userid := data["userid"].(string)
		stat, ok := uidStats[userid]
		if !ok {
			continue
		}
		mappingStatGcyx2HuaType(data, stat)
	}

	for _, stat := range stats { // 汇总后二次计算
		mappingStatGcyx0(stat)
	}

	var totalHeart float64
	// 查用户信息
	var users []map[string]any
	ck.Select(&users, `
		select userid,nickname,ctime,login_time,money,state,regist_area,
			date_diff('day', ctime, now()) live_days, date_diff('day', login_time, now()) loss_days,tp_heartbeat_hr
		from game.col_user final where userid in ?
	`, userids)
	for _, user := range users {
		userid := user["userid"].(string)
		nickname := user["nickname"].(string)
		ctime := user["ctime"].(time.Time)
		login_time := user["login_time"].(time.Time)
		money := utils.ToInt64(user["money"])
		live_days := utils.ToInt64(user["live_days"])
		loss_days := utils.ToInt64(user["loss_days"])
		state := utils.ToInt64(user["state"])
		regist_area := utils.ToInt64(user["regist_area"])
		tp_heartbeat_hr := utils.ToFloat64(user["tp_heartbeat_hr"])

		stat, ok := uidStats[userid]
		if ok {
			stat.Nickname = nickname
			stat.Ctime = ctime.Format(utils.FORMAT)
			stat.LoginTime = login_time.Format(utils.FORMAT)
			stat.LiveDays = int(live_days)
			stat.LoseDays = int(loss_days)
			stat.TpHeartbeatHr = tp_heartbeat_hr
			if tp_heartbeat_hr > 0 {
				totalHeart += tp_heartbeat_hr
				stat.PlayerHeartRate = fmt.Sprintf("%.1f", tp_heartbeat_hr)
			}

			switch regist_area {
			case 0, 3:
				stat.RegistArea = "A类"
			case 1:
				stat.RegistArea = "B类"
			case 2:
				stat.RegistArea = "C类"
			}
			switch state {
			case 1:
				stat.UserType = "新手"
			case 2:
				if money == 0 {
					stat.UserType = "零充"
				} else if money >= 20000 && money <= 99999 {
					stat.UserType = "普充"
				} else if money >= 100000 && money <= 499999 {
					stat.UserType = "小R"
				} else if money >= 500000 && money <= 999999 {
					stat.UserType = "中R"
				} else if money >= 1000000 && money <= 9999999 {
					stat.UserType = "大R"
				} else if money >= 10000000 {
					stat.UserType = "超大R"
				}
			case 3:
				stat.UserType = "平民"
			case 4:
				stat.UserType = "泡沫"
			}
		}
	}

	// 查总汇数据
	mark := "--"
	summary := &entity.TpNewGCYXStat{
		No:            fmt.Sprintf("%v-%v总汇", params["startDate"], params["endDate"]),
		UserId:        mark,
		Nickname:      mark,
		Ctime:         mark,
		LoginTime:     mark,
		TpHeartbeatHr: totalHeart,
		TPNum:         int64(len(userids)),
	}
	sqlSummary := strings.ReplaceAll(sql0, "select userid, win_type", "select win_type")
	sqlSummary = strings.ReplaceAll(sqlSummary, "limit ?, ?", "")
	sqlSummary = strings.ReplaceAll(sqlSummary, "group by userid, win_type", "group by win_type")
	var summaryDatas []map[string]any
	err = ck.Select(&summaryDatas, sqlSummary, args...)
	if err != nil {
		return
	}
	for _, data := range summaryDatas {
		mappingStatGcyx1(data, summary)
	}

	// 牌型输赢
	var summaryHuaTypeDatas []map[string]any
	sqlSummaryHandType := strings.ReplaceAll(sqlHandType, "select userid, win_type", "select win_type")
	sqlSummaryHandType = strings.ReplaceAll(sqlSummaryHandType, "group by userid, win_type", "group by win_type")
	err = ck.Select(&summaryHuaTypeDatas, sqlSummaryHandType, argsHandType...)
	if err != nil {
		return
	}
	for _, data := range summaryHuaTypeDatas {
		mappingStatGcyx2HuaType(data, summary)
	}
	// 汇总后二次计算
	mappingStatGcyx0(summary)
	stats = append([]*entity.TpNewGCYXStat{summary}, stats...)
	return
}

func mappingStatGcyx1(src map[string]any, dst *entity.TpNewGCYXStat) {
	win_type := utils.ToInt64(src["win_type"])
	dst.AllRounds += utils.ToInt64(src["round"])
	dst.Bets += utils.ToInt64(src["bet_total"])

	// 策略
	dst.StrategyRounds += utils.ToInt64(src["s_round"])
	dst.StrategyBets += utils.ToInt64(src["s_bet_total"])
	dst.HighRpRounds += utils.ToInt64(src["rp_round"])
	dst.HighBpRounds += utils.ToInt64(src["bp_round"])
	dst.HighChargeTimes += int16(utils.ToInt64(src["charge_times"]))
	dst.HighChargeLaunchTimes += int16(utils.ToInt64(src["charge_launch_times"]))
	dst.HighChargePayTimes += int16(utils.ToInt64(src["charge_pay_times"]))

	if win_type == 1 { // 赢
		dst.Wins = utils.ToInt64(src["score_total"])
		dst.WinRounds = utils.ToInt64(src["round"])

		dst.StrategyWinRounds = utils.ToInt64(src["s_round"])
		dst.StrategyWins = utils.ToInt64(src["s_score"])
		dst.FStrategyWins = fmt.Sprintf("%.2f", float64(utils.ToInt64(src["s_score"]))/100)
		dst.StrategyWinBetAvg = fmt.Sprintf("%.2f", ComputeFloat(utils.ToInt64(src["s_bet_total"]), utils.ToInt64(src["s_round"]))/100)
		// 牌型赢次数
		for i := 1; i <= 12; i++ {
			dst.HuaTypeWinTimes[i] = int(utils.ToInt64(src[fmt.Sprintf("hua%d", i)]))
		}
	} else if win_type == 2 { // 输
		dst.Loses = utils.ToInt64(src["score_total"])
		dst.StrategyLoses = utils.ToInt64(src["s_score"])
		dst.FStrategyLoses = fmt.Sprintf("%.2f", float64(utils.ToInt64(src["s_score"]))/100)
		dst.StrategyLoseBetAvg = fmt.Sprintf("%.2f", ComputeFloat(utils.ToInt64(src["s_bet_total"]), utils.ToInt64(src["s_round"]))/100)
	}

	if dst.TPNum > 0 {
		// TPNum
		dst.PlayerHeartRate = fmt.Sprintf("%.1f", float64(dst.TpHeartbeatHr)/float64(dst.TPNum)) // 玩家心率
	}
	// 拿牌型次数
	for i := 1; i <= 12; i++ {
		dst.HuaTypeTimes[i] += int(utils.ToInt64(src[fmt.Sprintf("hua%d", i)]))
	}
}

// 牌型输赢 mapping
func mappingStatGcyx2HuaType(src map[string]any, dst *entity.TpNewGCYXStat) {
	win_type := utils.ToInt64(src["win_type"])
	// tp_hand_type_up_down := utils.ToInt64(src["tp_hand_type_up_down"])
	// round := utils.ToInt64(src["round"])
	// score_total := utils.ToInt64(src["score_total"])
	// score_avg := utils.ToInt64(src["score_avg"])
	// score_max := utils.ToInt64(src["score_max"])
	// score_min := utils.ToInt64(src["score_min"])
	if win_type == 1 {
		// if score_total > dst.MaxWinHuaTypeAmount {
		// 	dst.MaxWinHuaTypeAmount = score_total
		// 	dst.MaxWinHuaType = HuaTypeNames[int(tp_hand_type_up_down)]
		// 	dst.MaxWinHuaTypeRound = round
		// 	dst.MaxWinHuaTypeAmountAvg = score_avg
		// }
		// if score_max > dst.OnceMaxWinHuaTypeAmount {
		// 	dst.OnceMaxWinHuaTypeAmount = score_max
		// 	dst.OnceMaxWinHuaType = HuaTypeNames[int(tp_hand_type_up_down)]
		// }
	} else {
		// if score_total < dst.MaxLoseHuaTypeAmount {
		// 	dst.MaxLoseHuaTypeAmount = score_total
		// 	dst.MaxLoseHuaType = HuaTypeNames[int(tp_hand_type_up_down)]
		// 	dst.MaxLoseHuaTypeRound = round
		// 	dst.MaxLoseHuaTypeAmountAvg = score_avg
		// }
		// if score_min < dst.OnceMaxLoseHuaTypeAmount {
		// 	dst.OnceMaxLoseHuaTypeAmount = score_min
		// 	dst.OnceMaxLoseHuaType = HuaTypeNames[int(tp_hand_type_up_down)]
		// }
	}
}

func mappingStatGcyx0(dst *entity.TpNewGCYXStat) {
	if dst.AllRounds > 0 {
		dst.BetsAvg = fmt.Sprintf("%.2f", float64(dst.Bets)/float64(dst.AllRounds)/100) // 局均打码量
	}
	// if dst.AllRounds > 0 {
	// 	dst.WinRate = fmt.Sprintf("%.2f%%", float64(dst.WinRounds)/float64(dst.AllRounds)*100) // 玩家胜率
	// }
	// if dst.Loses != 0 {
	// 	dst.RewardRate = fmt.Sprintf("%.2f%%", float64(dst.Wins)/float64(-dst.Loses)*100) // 总返奖率
	// }
	if dst.StrategyRounds > 0 {
		dst.StrategyBetsAvg = fmt.Sprintf("%.2f", Chip2Float(dst.StrategyBets)/float64(dst.StrategyRounds))         // 策略局局均打码量
		dst.StrategyWinRate = fmt.Sprintf("%.2f%%", float64(dst.StrategyWinRounds)/float64(dst.StrategyRounds)*100) // 策略中胜率
	}
	if dst.StrategyLoses != 0 {
		dst.StrategyRewardRate = fmt.Sprintf("%.2f%%", float64(dst.StrategyWins)/float64(-dst.StrategyLoses)*100) // 策略中返奖率
	}
	if dst.StrategyRounds > 0 {
		dst.HighRpRate = fmt.Sprintf("%.2f%%", float64(dst.HighRpRounds)/float64(dst.StrategyRounds)*100) // 高潮局压制次数占比
		dst.HighBpRate = fmt.Sprintf("%.2f%%", float64(dst.HighBpRounds)/float64(dst.StrategyRounds)*100) // 高潮局恩赐次数占比
	}

	if dst.StrategyRounds > 0 {
		dst.Hua1Rate = utils.CaseElse(dst.HuaTypeTimes[1] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[1])/float64(dst.StrategyRounds)*100))    // 玩家拿大豹子率
		dst.Hua2Rate = utils.CaseElse(dst.HuaTypeTimes[2] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[2])/float64(dst.StrategyRounds)*100))    // 玩家拿小豹子率
		dst.Hua3Rate = utils.CaseElse(dst.HuaTypeTimes[3] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[3])/float64(dst.StrategyRounds)*100))    // 玩家拿大顺金率
		dst.Hua4Rate = utils.CaseElse(dst.HuaTypeTimes[4] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[4])/float64(dst.StrategyRounds)*100))    // 玩家拿小顺金率
		dst.Hua5Rate = utils.CaseElse(dst.HuaTypeTimes[5] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[5])/float64(dst.StrategyRounds)*100))    // 玩家拿大顺子率
		dst.Hua6Rate = utils.CaseElse(dst.HuaTypeTimes[6] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[6])/float64(dst.StrategyRounds)*100))    // 玩家拿小顺子率
		dst.Hua7Rate = utils.CaseElse(dst.HuaTypeTimes[7] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[7])/float64(dst.StrategyRounds)*100))    // 玩家拿大同花率
		dst.Hua8Rate = utils.CaseElse(dst.HuaTypeTimes[8] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[8])/float64(dst.StrategyRounds)*100))    // 玩家拿小同花率
		dst.Hua9Rate = utils.CaseElse(dst.HuaTypeTimes[9] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[9])/float64(dst.StrategyRounds)*100))    // 玩家拿大对子率
		dst.Hua10Rate = utils.CaseElse(dst.HuaTypeTimes[10] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[10])/float64(dst.StrategyRounds)*100)) // 玩家拿小对子率
		dst.Hua11Rate = utils.CaseElse(dst.HuaTypeTimes[11] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[11])/float64(dst.StrategyRounds)*100)) // 玩家拿大高牌率
		dst.Hua12Rate = utils.CaseElse(dst.HuaTypeTimes[12] == 0, "", fmt.Sprintf("%.2f%%", float64(dst.HuaTypeTimes[12])/float64(dst.StrategyRounds)*100)) // 玩家拿小高牌率
	}
	dst.Hua1WinRate = utils.CaseElse(dst.HuaTypeTimes[1] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[1], dst.HuaTypeTimes[1])*100))     // 玩家拿大豹子胜率
	dst.Hua2WinRate = utils.CaseElse(dst.HuaTypeTimes[2] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[2], dst.HuaTypeTimes[2])*100))     // 玩家拿小豹子胜率
	dst.Hua3WinRate = utils.CaseElse(dst.HuaTypeTimes[3] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[3], dst.HuaTypeTimes[3])*100))     // 玩家拿大顺金胜率
	dst.Hua4WinRate = utils.CaseElse(dst.HuaTypeTimes[4] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[4], dst.HuaTypeTimes[4])*100))     // 玩家拿小顺金胜率
	dst.Hua5WinRate = utils.CaseElse(dst.HuaTypeTimes[5] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[5], dst.HuaTypeTimes[5])*100))     // 玩家拿大顺子胜率
	dst.Hua6WinRate = utils.CaseElse(dst.HuaTypeTimes[6] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[6], dst.HuaTypeTimes[6])*100))     // 玩家拿小顺子胜率
	dst.Hua7WinRate = utils.CaseElse(dst.HuaTypeTimes[7] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[7], dst.HuaTypeTimes[7])*100))     // 玩家拿大同花胜率
	dst.Hua8WinRate = utils.CaseElse(dst.HuaTypeTimes[8] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[8], dst.HuaTypeTimes[8])*100))     // 玩家拿小同花胜率
	dst.Hua9WinRate = utils.CaseElse(dst.HuaTypeTimes[9] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[9], dst.HuaTypeTimes[9])*100))     // 玩家拿大对子胜率
	dst.Hua10WinRate = utils.CaseElse(dst.HuaTypeTimes[10] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[10], dst.HuaTypeTimes[10])*100)) // 玩家拿小对子胜率
	dst.Hua11WinRate = utils.CaseElse(dst.HuaTypeTimes[11] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[11], dst.HuaTypeTimes[11])*100)) // 玩家拿大高牌胜率
	dst.Hua12WinRate = utils.CaseElse(dst.HuaTypeTimes[12] == 0, "没拿过", fmt.Sprintf("%.2f%%", ComputeFloat(dst.HuaTypeWinTimes[12], dst.HuaTypeTimes[12])*100)) // 玩家拿小高牌胜率

	dst.FBets = fmt.Sprintf("%.2f", Chip2Float(dst.Bets))

	// 策略
	dst.FStrategyBets = fmt.Sprintf("%.2f", Chip2Float(dst.StrategyBets))
	dst.StrategyRate = fmt.Sprintf("%.2f%%", ComputeFloat(dst.StrategyRounds, dst.AllRounds)*100)
	dst.TpRebateRate = fmt.Sprintf("%.2f%%", ComputeFloat(dst.Wins, -dst.Loses)*100)
	dst.TpWinRate = fmt.Sprintf("%.2f%%", ComputeFloat(dst.WinRounds, dst.AllRounds)*100)
	dst.StrategyCash = fmt.Sprintf("%.2f", Chip2Float(dst.StrategyWins+dst.StrategyLoses))
}

func (s *gameStatsService) MinesStatDates(params map[string]any) (stats []*entity.MinesStatDate, err error) {
	var where1, where2, where3, having4 string
	var args1, args2, args3, args4 []any

	if userid, ok := params["userid"]; ok {
		where1 += " and userid = ?"
		where2 += " and userid = ?"
		where3 += " and userid = ?"
		args1 = append(args1, userid)
		args2 = append(args2, userid)
		args3 = append(args3, userid)
	}

	// 时间
	startTime, ok1 := params["startTime"]
	endTime, ok2 := params["endTime"]
	if ok1 && ok2 {
		where1 += " and begin_time BETWEEN ? AND ?"
		where2 += " and begin_time BETWEEN ? AND ?"
		args1 = append(args1, startTime, endTime)
		args2 = append(args2, startTime, endTime)
	} else if ok1 {
		where1 += " and begin_time >= ?"
		where2 += " and begin_time >= ?"
		args1 = append(args1, startTime)
		args2 = append(args2, startTime)
	} else if ok2 {
		where1 += " and begin_time <= ?"
		where2 += " and begin_time <= ?"
		args1 = append(args1, endTime)
		args2 = append(args2, endTime)
	}

	// 打码量
	bets1, ok1 := params["bets1"]
	bets2, ok2 := params["bets2"]
	if ok1 && ok2 {
		having4 += " and bet_total between ? and ?"
		args4 = append(args4, bets1, bets2)
	}
	// 注均码
	betAvg1, ok1 := params["betAvg1"]
	betAvg2, ok2 := params["betAvg2"]
	if ok1 && ok2 {
		having4 += " and bet_avg between ? and ?"
		args4 = append(args4, betAvg1, betAvg2)
	}

	// 净赢
	cash1, ok1 := params["cash1"]
	cash2, ok2 := params["cash2"]
	if ok1 && ok2 {
		having4 += " and score_total between ? and ?"
		args4 = append(args4, cash1, cash2)
	}

	// 流失天数
	var uFilter bool
	uLossDays1, ok1 := params["uLossDays1"]
	uLossDays2, ok2 := params["uLossDays2"]
	uFilter = ok1 || ok2
	if ok1 && ok2 {
		where3 += " and date_diff('day', login_time, now()) between ? and ?"
		args3 = append(args3, uLossDays1, uLossDays2)
	}
	// money
	utypeQ, ok1 := params["utypeQ"]
	if ok1 {
		where3 += utypeQ.(string)
		uFilter = true
	}

	// 渠道
	if ad__bundle_id, ok := params["ad__bundle_id"]; ok {
		where3 += " and ad__bundle_id in ?"
		args3 = append(args3, ad__bundle_id)
		uFilter = true
	}
	// ab 类
	if regist_area, ok := params["regist_area"]; ok {
		where3 += " and regist_area in ?"
		args3 = append(args3, regist_area)
		uFilter = true
	}

	sql0 := `
		select toYYYYMMDD(toDateTime(begin_time)) datestr, win_type, count(*) round, sum(bet_amount) bet_total, 
			avg(bet_amount) bet_avg, sum(score) score_total, avg(score) score_avg,
			SUM(end_time - begin_time) game_times,
			MAX(mines_multiple) multiple_max, AVG(mines_multiple) multiple_avg, median(mines_multiple) multiple_median
		from game.col_detail final 
		where gtype = 14 and robot = 0 %s
		and userid in (
			select userid from (
				select userid, SUM(bet_amount) bet_total, avg(bet_amount) bet_avg, SUM(score) score_total
				from game.col_detail final 
				where gtype = 14 and robot = 0 %s
				%s
				group by userid
				having 1 = 1 %s
			)
		)
		group by datestr, win_type
		order by datestr desc
	`
	uFilterSql3 := `
		and userid in (
			select userid from game.col_user final where 1 = 1  %s
		)
	`
	if !uFilter {
		uFilterSql3 = ""
		args3 = []any{}
	} else {
		uFilterSql3 = fmt.Sprintf(uFilterSql3, where3)
	}

	sql0 = fmt.Sprintf(sql0, where1, where2, uFilterSql3, having4)
	var datas []map[string]any
	args := append(args1, args2...)
	args = append(args, args3...)
	args = append(args, args4...)
	err = ck.Select(&datas, sql0, args...)
	if err != nil {
		return
	}

	var dateStats = make(map[uint32]*entity.MinesStatDate)
	for _, data := range datas { // 输和赢的统计汇总
		datestr := data["datestr"].(uint32)
		stat, ok := dateStats[datestr]
		if !ok {
			stat = &entity.MinesStatDate{
				DateStr: fmt.Sprint(datestr),
			}
			dateStats[datestr] = stat
			stats = append(stats, stat)
		}
		minesMappingStatDate1(data, stat)
	}

	summary := &entity.MinesStatDate{
		DateStr: fmt.Sprintf("%v-%v总汇", params["startDate"], params["endDate"]),
	}
	// 新老顾客玩mines数量
	sqlPlayers := `
		select s1.datestr, s1.players, s1.round_avg, s1.round_median, length(arrayIntersect(s1.userids, s2.userids)) new_players, (players - new_players) old_players
		from (
			select datestr, count(*) players, avg(round) round_avg, median(round) round_median, groupArray(userid) userids
			from (
				select toYYYYMMDD(toDateTime(begin_time)) datestr, userid, count(*) round
				from game.col_detail final 
				where gtype = 14 and robot = 0 %s
				and userid in (
					select userid from (
						select userid, SUM(bet_amount) bet_total, avg(bet_amount) bet_avg, SUM(score) score_total
						from game.col_detail final 
						where gtype = 14 and robot = 0 %s
						%s
						group by userid
						having 1 = 1 %s
					)
				)
				group by datestr, userid
			) t1 
			group by datestr
		) s1
		left join (
			select toYYYYMMDD(ctime) datestr, groupArray(userid) userids
			from game.col_user final 
			where ctime between ? AND ?
			group by datestr
		) s2 on s1.datestr = s2.datestr
	`
	sqlPlayers = fmt.Sprintf(sqlPlayers, where1, where2, uFilterSql3, having4)
	start := params["start"].(time.Time)
	end := params["end"].(time.Time)
	argsPlayers := append(args, start, end)
	var playerDatas []map[string]any
	err = ck.Select(&playerDatas, sqlPlayers, argsPlayers...)
	if err != nil {
		return
	}
	for _, data := range playerDatas {
		datestr := data["datestr"].(uint32)
		if stat, ok := dateStats[datestr]; ok {
			stat.Players = int32(utils.ToInt64(data["players"]))
			stat.PlayerRoundsAvg = fmt.Sprintf("%.2f", utils.ToFloat64(data["round_avg"]))
			stat.PlayerRoundsMedian = fmt.Sprintf("%.2f", utils.ToFloat64(data["round_median"]))
			stat.NewPlayers = int32(utils.ToInt64(data["new_players"]))
			stat.OldPlayers = int32(utils.ToInt64(data["old_players"]))

			summary.Players += stat.Players
			summary.NewPlayers += stat.NewPlayers
			summary.OldPlayers += stat.OldPlayers
		}
	}

	// 新老顾客日活
	var loginDatas []map[string]any
	err = ck.Select(&loginDatas, fmt.Sprintf(`
		select s1.datestr, s1.players, s2.new_players, (players - new_players) old_players,tp_heartbeat_hr_sum,tp_heartbeat_users
		from (
			select datestr, count(*) players
			from (
				select toYYYYMMDD(login_time) datestr
				from game.col_log_login final 
				where login_time between ? AND ?
				%s
				group by datestr, userid
			) t1 
			group by datestr
		) s1
		left join (
			select toYYYYMMDD(ctime) datestr, count(*) new_players, 
				SUM(tp_heartbeat_hr) AS tp_heartbeat_hr_sum,
				SUM(CASE WHEN tp_heartbeat_hr != 0 THEN 1 ELSE 0 END) tp_heartbeat_users
			from game.col_user final 
			where ctime between ? AND ?
			%s
			group by datestr
		) s2 on s1.datestr = s2.datestr
	`, uFilterSql3, uFilterSql3), append(append([]any{start, end}, args3...), append([]any{start, end}, args3...)...)...)
	if err != nil {
		return
	}
	for _, data := range loginDatas {
		datestr := data["datestr"].(uint32)
		if stat, ok := dateStats[datestr]; ok {
			stat.LoginedUsers = int32(utils.ToInt64(data["players"]))
			stat.NewLoginedUsers = int32(utils.ToInt64(data["new_players"]))
			stat.OldLoginedUsers = int32(utils.ToInt64(data["old_players"]))

			summary.LoginedUsers += stat.LoginedUsers
			summary.NewLoginedUsers += stat.NewLoginedUsers
			summary.OldLoginedUsers += stat.OldLoginedUsers

		}
	}

	// mines局内充值
	sql_pay := `
		SELECT toYYYYMMDD(ctime) ctimestr,
		  SUM(case when in_gtype = 14 then 1 else 0 end) pay_times,
		  SUM(case when in_gtype = 14 then amount else 0 end) amount_sum, 
		  COUNT(*) pay_times_all
		FROM game.col_trade_record FINAL WHERE order_status = 4 AND ctime BETWEEN ? AND ?
		GROUP BY ctimestr
	`
	var payStats []map[string]any
	err = ck.Select(&payStats, sql_pay, start, end)
	if err != nil {
		return
	}
	for _, data := range payStats {
		ctimestr := data["ctimestr"].(uint32)
		if stat, ok := dateStats[ctimestr]; ok {
			pay_times := utils.ToInt64(data["pay_times"])
			amount_sum := utils.ToInt64(data["amount_sum"])
			pay_times_all := utils.ToInt64(data["pay_times_all"])
			stat.ChargeTimes = pay_times
			stat.ChargeAmounts = fmt.Sprintf("%.2f", Chip2Float(amount_sum))
			if pay_times > 0 {
				stat.ChargePayAvg = fmt.Sprintf("%.2f", Chip2Float(amount_sum)/float64(pay_times))
			}
			if pay_times_all > 0 {
				stat.ChargeRate = fmt.Sprintf("%.2f%%", ComputeFloat(pay_times, pay_times_all)*100)
			}
		}
	}

	// 汇总后二次计算
	for _, stat := range stats {
		minesMappingStatDate0(stat, true)
	}

	// 查询总汇数据
	sqlSummary := strings.ReplaceAll(sql0, "select toYYYYMMDD(toDateTime(begin_time)) datestr, win_type", "select win_type")
	sqlSummary = strings.ReplaceAll(sqlSummary, "group by datestr, win_type", "group by win_type")
	sqlSummary = strings.ReplaceAll(sqlSummary, "order by datestr desc", "")
	var summaryDatas []map[string]any
	err = ck.Select(&summaryDatas, sqlSummary, args...)
	if err != nil {
		return
	}
	for _, data := range summaryDatas {
		minesMappingStatDate1(data, summary)
	}

	minesMappingStatDate0(summary, false)
	// 总汇人均局数
	summary.PlayerRoundsAvg = fmt.Sprintf("%.2f", ComputeFloat(summary.AllRounds, summary.Players))
	// 局数中位数
	var summaryMedian map[string]any
	err = ck.Select(&summaryMedian, fmt.Sprintf(`
		select median(round) round_median
		from (
			select userid, count(*) round
			from game.col_detail final 
			where gtype = 14 and robot = 0 %s
			and userid in (
				select userid from (
					select userid, SUM(bet_amount) bet_total, avg(bet_amount) bet_avg, SUM(score) score_total
					from game.col_detail final 
					where gtype = 14 and robot = 0 %s
					%s
					group by userid
					having 1 = 1 %s
				)
			)
			group by userid
		) t1
	`, where1, where2, uFilterSql3, having4), args...)
	if err != nil {
		return
	}
	summary.PlayerRoundsMedian = fmt.Sprintf("%.2f", utils.ToFloat64(summaryMedian["round_median"]))

	stats = append([]*entity.MinesStatDate{summary}, stats...)
	return
}

func minesMappingStatDate1(src map[string]any, dst *entity.MinesStatDate) {
	win_type := utils.ToInt64(src["win_type"])
	dst.AllRounds += int32(utils.ToInt64(src["round"]))
	dst.PlayerBets += utils.ToInt64(src["bet_total"])
	dst.Cash += utils.ToInt64(src["score_total"])
	dst.GameSeconds += utils.ToInt64(src["game_times"])

	if win_type == int64(ck.WinTypeWin) { // 赢
		dst.WinRounds = int32(utils.ToInt64(src["round"]))
		dst.WinBets = utils.ToInt64(src["bet_total"])
		dst.WinBetsAvg = utils.ToInt64(src["bet_avg"])
		dst.Wins = utils.ToInt64(src["score_total"])
		dst.FWinsAvg = fmt.Sprintf("%.2f", float64(utils.ToInt64(src["score_avg"]))/100)

		dst.WinMulpitleMax = fmt.Sprintf("%.2f", utils.ToFloat64(src["multiple_max"]))
		dst.WinMulpitleAvg = fmt.Sprintf("%.2f", utils.ToFloat64(src["multiple_avg"]))
		dst.WinMulpitleMedian = fmt.Sprintf("%.2f", utils.ToFloat64(src["multiple_median"]))

	} else if win_type == int64(ck.WinTypeLose) { // 输
		dst.LoseRounds = int32(utils.ToInt64(src["round"]))
		dst.LoseBets = utils.ToInt64(src["bet_total"])
		dst.LoseBetsAvg = utils.ToInt64(src["bet_avg"])
		dst.Loses = utils.ToInt64(src["score_total"])
		dst.FLosesAvg = fmt.Sprintf("%.2f", float64(utils.ToInt64(src["score_avg"]))/100)

		dst.LoseMulpitleMax = fmt.Sprintf("%.2f", utils.ToFloat64(src["multiple_max"]))
		dst.LoseMulpitleAvg = fmt.Sprintf("%.2f", utils.ToFloat64(src["multiple_avg"]))
		dst.LoseMulpitleMedian = fmt.Sprintf("%.2f", utils.ToFloat64(src["multiple_median"]))
	}
}

func minesMappingStatDate0(dst *entity.MinesStatDate, formatDate bool) {
	if formatDate {
		datestr := dst.DateStr
		dst.DateStr = datestr[0:4] + "-" + datestr[4:6] + "-" + datestr[6:8]
	}
	dst.PlayerRate = utils.CaseElse(dst.LoginedUsers == 0, "", fmt.Sprintf("%.2f%%", ComputeFloat(dst.Players, dst.LoginedUsers)*100))
	dst.NewPlayerRate = utils.CaseElse(dst.NewLoginedUsers == 0, "", fmt.Sprintf("%.2f%%", ComputeFloat(dst.NewPlayers, dst.NewLoginedUsers)*100))
	dst.OldPlayerRate = utils.CaseElse(dst.OldLoginedUsers == 0, "", fmt.Sprintf("%.2f%%", ComputeFloat(dst.OldPlayers, dst.OldLoginedUsers)*100))

	// 游戏时长
	dst.GameTimes = fmt.Sprintf("%.2f", float64(dst.GameSeconds)/60)
	if dst.AllRounds > 0 {
		dst.GameTimesAvg = fmt.Sprintf("%.2f", float64(dst.GameSeconds)/float64(dst.AllRounds)/60)
	}
	if dst.Players > 0 {
		dst.GameTimesPlayerAvg = fmt.Sprintf("%.2f", float64(dst.GameSeconds)/float64(dst.Players)/60)
	}

	if dst.AllRounds > 0 {
		dst.RoundsBetsAvg = fmt.Sprintf("%.2f", float64(dst.PlayerBets)/float64(dst.AllRounds)/100) // 局均打码量
	}
	if dst.AllRounds > 0 {
		dst.WinRate = fmt.Sprintf("%.2f%%", float64(dst.WinRounds)/float64(dst.AllRounds)*100) // 玩家胜率
	}
	if dst.PlayerBets != 0 {
		// dst.RewardRate = fmt.Sprintf("%.2f%%", float64(dst.Wins)/float64(-dst.Loses)*100) // 总返奖率
		dst.RewardRate = fmt.Sprintf("%.2f%%", float64(dst.Wins+dst.WinBets)/float64(dst.PlayerBets)*100) // 总返奖率
	}
	if dst.AllRounds > 0 {
		dst.FCashAvg = fmt.Sprintf("%.2f", float64(dst.Cash)/float64(dst.AllRounds)/100) // 局均净赢
	}

	dst.FBets = fmt.Sprintf("%.2f", Chip2Float(dst.PlayerBets))
	dst.RoundsBetsAvg = fmt.Sprintf("%.2f", ComputeFloat(dst.PlayerBets, int64(dst.AllRounds))/100)
	dst.PlayerBetsAvg = fmt.Sprintf("%.2f", ComputeFloat(dst.PlayerBets, int64(dst.Players))/100)
	dst.FWinBets = fmt.Sprintf("%.2f", Chip2Float(dst.WinBets))
	dst.FLoseBets = fmt.Sprintf("%.2f", Chip2Float(dst.LoseBets))
	dst.FWinBetsAvg = fmt.Sprintf("%.2f", Chip2Float(dst.WinBetsAvg))
	dst.FLoseBetsAvg = fmt.Sprintf("%.2f", Chip2Float(dst.LoseBetsAvg))
	dst.FWins = fmt.Sprintf("%.2f", Chip2Float(dst.Wins))
	dst.FLoses = fmt.Sprintf("%.2f", Chip2Float(dst.Loses))
	dst.FCash = fmt.Sprintf("%.2f", Chip2Float(dst.Cash))
}

// MinesStatPlayers mines汇总明细
func (s *gameStatsService) MinesStatPlayers(page, pageSize int, params map[string]any) (stats []*entity.MinesStatPlayer, total int, err error) {
	var where1, where2, where3, having4 string
	var args1, args2, args3, args4 []any

	if userid, ok := params["userid"]; ok {
		where1 += " and userid = ?"
		where2 += " and userid = ?"
		where3 += " and userid = ?"
		args1 = append(args1, userid)
		args2 = append(args2, userid)
		args3 = append(args3, userid)
	}

	// 时间
	startTime, ok1 := params["startTime"]
	endTime, ok2 := params["endTime"]
	if ok1 && ok2 {
		where1 += " and begin_time BETWEEN ? AND ?"
		where2 += " and begin_time BETWEEN ? AND ?"
		args1 = append(args1, startTime, endTime)
		args2 = append(args2, startTime, endTime)
	} else if ok1 {
		where1 += " and begin_time >= ?"
		where2 += " and begin_time >= ?"
		args1 = append(args1, startTime)
		args2 = append(args2, startTime)
	} else if ok2 {
		where1 += " and begin_time <= ?"
		where2 += " and begin_time <= ?"
		args1 = append(args1, endTime)
		args2 = append(args2, endTime)
	}

	// 打码量
	bets1, ok1 := params["bets1"]
	bets2, ok2 := params["bets2"]
	if ok1 && ok2 {
		having4 += " and bet_total between ? and ?"
		args4 = append(args4, bets1, bets2)
	}
	// 注均码
	betAvg1, ok1 := params["betAvg1"]
	betAvg2, ok2 := params["betAvg2"]
	if ok1 && ok2 {
		having4 += " and bet_avg between ? and ?"
		args4 = append(args4, betAvg1, betAvg2)
	}

	// 净赢
	cash1, ok1 := params["cash1"]
	cash2, ok2 := params["cash2"]
	if ok1 && ok2 {
		having4 += " and score_total between ? and ?"
		args4 = append(args4, cash1, cash2)
	}

	// 流失天数
	var uFilter bool
	uLossDays1, ok1 := params["uLossDays1"]
	uLossDays2, ok2 := params["uLossDays2"]
	uFilter = ok1 || ok2
	if ok1 && ok2 {
		where3 += " and date_diff('day', login_time, now()) between ? and ?"
		args3 = append(args3, uLossDays1, uLossDays2)
	}
	// money
	utypeQ, ok1 := params["utypeQ"]
	if ok1 {
		where3 += utypeQ.(string)
		uFilter = true
	}

	// 渠道
	if ad__bundle_id, ok := params["ad__bundle_id"]; ok {
		where3 += " and ad__bundle_id in ?"
		args3 = append(args3, ad__bundle_id)
		uFilter = true
	}
	// ab 类
	if regist_area, ok := params["regist_area"]; ok {
		where3 += " and regist_area in ?"
		args3 = append(args3, regist_area)
		uFilter = true
	}

	sql0 := `
		select userid, win_type, count(*) round, sum(bet_amount) bet_total, 
			avg(bet_amount) bet_avg, sum(score) score_total, avg(score) score_avg,
			MAX(mines_multiple) multiple_max, AVG(mines_multiple) multiple_avg, median(mines_multiple) multiple_median
		from game.col_detail final 
		where gtype = 14 and robot = 0 %s
		and userid in (
			select userid from (
				select userid, SUM(bet_amount) bet_total, avg(bet_amount) bet_avg, SUM(score) score_total, max(begin_time) lately_time 
				from game.col_detail final 
				where gtype = 14 and robot = 0 %s
				%s
				group by userid
				having 1 = 1 %s
			)
			order by lately_time desc limit ?, ?
		)
		group by userid, win_type
		order by max(begin_time) desc
	`
	uFilterSql3 := `
		and userid in (
			select userid from game.col_user final where 1 = 1  %s
		)
	`
	if !uFilter {
		uFilterSql3 = ""
		args3 = []any{}
	} else {
		uFilterSql3 = fmt.Sprintf(uFilterSql3, where3)
	}

	// 查总条数 , max(begin_time) lately_time
	err = ck.Select(&total, fmt.Sprintf(`
		select count(*) from (
			select userid, SUM(bet_amount) bet_total, avg(bet_amount) bet_avg, SUM(score) score_total
			from game.col_detail final 
			where gtype = 14 and robot = 0  %s
			%s
			group by userid
			having 1 = 1 %s
		)
	`, where2, uFilterSql3, having4), append(append(args2, args3...), args4...)...)
	if err != nil {
		return
	}

	sql0 = fmt.Sprintf(sql0, where1, where2, uFilterSql3, having4)
	var datas []map[string]any
	args := append(args1, args2...)
	args = append(args, args3...)
	args = append(args, args4...)
	offset, limit := PageCalc(page, pageSize)
	// args = append(args, offset, limit)
	err = ck.Select(&datas, sql0, append(args, offset, limit)...)
	if err != nil {
		return
	}

	var uidStats = make(map[string]*entity.MinesStatPlayer)
	var userids []string
	var no int
	for _, data := range datas { // 输和赢的统计汇总
		userid := data["userid"].(string)
		stat, ok := uidStats[userid]
		if !ok {
			no++
			stat = &entity.MinesStatPlayer{
				No:     strconv.Itoa(no),
				UserId: userid,
			}
			uidStats[userid] = stat
			stats = append(stats, stat)
			userids = append(userids, userid)
		}
		minesMappingStatPlayer1(data, stat)
	}

	for _, stat := range stats { // 汇总后二次计算
		minesMappingStatPlayer0(stat)
	}

	// 查用户信息
	var users []map[string]any
	err = ck.Select(&users, `
		select userid,nickname,ctime,login_time,state,regist_area,vip_lv,money,cash_out,diamond,
			date_diff('day', ctime, now()) live_days, date_diff('day', login_time, now()) loss_days
		from game.col_user final where userid in ?
	`, userids)
	if err != nil {
		return
	}
	for _, user := range users {
		userid := user["userid"].(string)
		nickname := user["nickname"].(string)
		ctime := user["ctime"].(time.Time)
		login_time := user["login_time"].(time.Time)
		live_days := utils.ToInt64(user["live_days"])
		loss_days := utils.ToInt64(user["loss_days"])
		state := utils.ToInt64(user["state"])
		regist_area := utils.ToInt64(user["regist_area"])
		vip_lv := utils.ToInt64(user["vip_lv"])
		money := utils.ToInt64(user["money"])
		cash_out := utils.ToInt64(user["cash_out"])
		diamond := utils.ToInt64(user["diamond"])

		stat, ok := uidStats[userid]
		if ok {
			stat.Nickname = nickname
			stat.Ctime = ctime.Format(utils.FORMAT)
			stat.LoginTime = login_time.Format(utils.FORMAT)
			stat.LiveDays = int(live_days)
			stat.LoseDays = int(loss_days)
			stat.VipLv = int32(vip_lv)

			stat.FMoney = fmt.Sprintf("%.2f", Chip2Float(money))
			stat.FCashOut = fmt.Sprintf("%.2f", Chip2Float(cash_out))
			stat.FProfit = fmt.Sprintf("%.2f", Chip2Float(diamond+cash_out-money)) // 携带金额+提现金额-充值金额

			switch regist_area {
			case 0, 3:
				stat.RegistArea = "A类"
			case 1:
				stat.RegistArea = "B类"
			case 2:
				stat.RegistArea = "C类"
			}
			switch state {
			case 1:
				stat.UserType = "新手"
			case 2:
				if money == 0 {
					stat.UserType = "零充"
				} else if money >= 20000 && money <= 99999 {
					stat.UserType = "普充"
				} else if money >= 100000 && money <= 499999 {
					stat.UserType = "小R"
				} else if money >= 500000 && money <= 999999 {
					stat.UserType = "中R"
				} else if money >= 1000000 && money <= 9999999 {
					stat.UserType = "大R"
				} else if money >= 10000000 {
					stat.UserType = "超大R"
				}
			case 3:
				stat.UserType = "平民"
			case 4:
				stat.UserType = "泡沫"
			}
		}
	}

	// 查所有游戏总打码量 mines打码量占比
	sql_bets := `
		select userid, SUM(bet_amount) bet_total
		from game.col_detail final where userid in ? and robot = 0 %s
		group by userid
	`
	var betStats []map[string]any
	err = ck.Select(&betStats, fmt.Sprintf(sql_bets, where1), append([]any{userids}, args1...)...)
	if err != nil {
		return
	}
	for _, data := range betStats {
		userid := data["userid"].(string)
		if stat, ok := uidStats[userid]; ok {
			bet_total := utils.ToInt64(data["bet_total"])
			stat.BetsAll = bet_total
			stat.FBetsAll = fmt.Sprintf("%.2f", Chip2Float(stat.BetsAll))
			stat.BetsRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.Bets, stat.BetsAll)*100)
		}
	}

	// mines局内充值
	sql_pay := `
		SELECT userid,
		  SUM(case when in_gtype = 14 then 1 else 0 end) pay_times,
		  SUM(case when in_gtype = 14 then amount else 0 end) amount_sum, 
		  COUNT(*) pay_times_all
		FROM game.col_trade_record FINAL WHERE order_status = 4 AND userid IN ?
		GROUP BY userid
	`
	var payStats []map[string]any
	err = ck.Select(&payStats, sql_pay, userids)
	if err != nil {
		return
	}
	for _, data := range payStats {
		userid := data["userid"].(string)
		if stat, ok := uidStats[userid]; ok {
			pay_times := utils.ToInt64(data["pay_times"])
			amount_sum := utils.ToInt64(data["amount_sum"])
			pay_times_all := utils.ToInt64(data["pay_times_all"])
			if pay_times > 0 {
				stat.ChargeTimes = fmt.Sprint(pay_times)
				stat.ChargeAmounts = fmt.Sprintf("%.2f", Chip2Float(amount_sum))
				stat.ChargePayAvg = fmt.Sprintf("%.2f", Chip2Float(amount_sum)/float64(pay_times))
				if pay_times_all > 0 {
					stat.ChargeRate = fmt.Sprintf("%.2f%%", ComputeFloat(pay_times, pay_times_all)*100)
				}
			}
		}
	}

	// 查总汇数据
	mark := "--"
	summary := &entity.MinesStatPlayer{
		No:            fmt.Sprintf("%v-%v总汇", params["startDate"], params["endDate"]),
		UserId:        mark,
		Nickname:      mark,
		Ctime:         mark,
		LoginTime:     mark,
		UserType:      mark,
		RegistArea:    mark,
		BetsRate:      mark,
		ChargeTimes:   mark,
		ChargeAmounts: mark,
		ChargePayAvg:  mark,
		ChargeRate:    mark,
		FMoney:        mark,
		FCashOut:      mark,
		FProfit:       mark,
		FBetsAll:      mark,
	}
	sqlSummary := strings.ReplaceAll(sql0, "select userid, win_type", "select win_type")
	sqlSummary = strings.ReplaceAll(sqlSummary, "limit ?, ?", "")
	sqlSummary = strings.ReplaceAll(sqlSummary, "group by userid, win_type", "group by win_type")
	var summaryDatas []map[string]any
	err = ck.Select(&summaryDatas, sqlSummary, args...)
	if err != nil {
		return
	}
	for _, data := range summaryDatas {
		minesMappingStatPlayer1(data, summary)
	}
	minesMappingStatPlayer0(summary) // 汇总后二次计算
	stats = append([]*entity.MinesStatPlayer{summary}, stats...)
	return
}

func minesMappingStatPlayer1(src map[string]any, dst *entity.MinesStatPlayer) {
	win_type := utils.ToInt64(src["win_type"])
	dst.AllRounds += int32(utils.ToInt64(src["round"]))
	dst.Bets += utils.ToInt64(src["bet_total"])
	dst.Cash += utils.ToInt64(src["score_total"])

	if win_type == 1 { // 赢
		dst.WinRounds = int32(utils.ToInt64(src["round"]))
		dst.WinBets = utils.ToInt64(src["bet_total"])
		dst.WinBetsAvg = utils.ToInt64(src["bet_avg"])
		dst.Wins = utils.ToInt64(src["score_total"])
		dst.FWinsAvg = fmt.Sprintf("%.2f", float64(utils.ToInt64(src["score_avg"]))/100)

		dst.WinMulpitleMax = fmt.Sprintf("%.2f", utils.ToFloat64(src["multiple_max"]))
		dst.WinMulpitleAvg = fmt.Sprintf("%.2f", utils.ToFloat64(src["multiple_avg"]))
		dst.WinMulpitleMedian = fmt.Sprintf("%.2f", utils.ToFloat64(src["multiple_median"]))

	} else { // 输
		dst.LoseRounds = int32(utils.ToInt64(src["round"]))
		dst.LoseBets = utils.ToInt64(src["bet_total"])
		dst.LoseBetsAvg = utils.ToInt64(src["bet_avg"])
		dst.Loses = utils.ToInt64(src["score_total"])
		dst.FLosesAvg = fmt.Sprintf("%.2f", float64(utils.ToInt64(src["score_avg"]))/100)

		dst.LoseMulpitleMax = fmt.Sprintf("%.2f", utils.ToFloat64(src["multiple_max"]))
		dst.LoseMulpitleAvg = fmt.Sprintf("%.2f", utils.ToFloat64(src["multiple_avg"]))
		dst.LoseMulpitleMedian = fmt.Sprintf("%.2f", utils.ToFloat64(src["multiple_median"]))
	}
}

func minesMappingStatPlayer0(dst *entity.MinesStatPlayer) {
	if dst.AllRounds > 0 {
		dst.BetsAvg = fmt.Sprintf("%.2f", float64(dst.Bets)/float64(dst.AllRounds)/100) // 局均打码量
	}
	if dst.AllRounds > 0 {
		dst.WinRate = fmt.Sprintf("%.2f%%", float64(dst.WinRounds)/float64(dst.AllRounds)*100) // 玩家胜率
	}
	if dst.Bets != 0 {
		dst.RewardRate = fmt.Sprintf("%.2f%%", float64(dst.Wins+dst.WinBets)/float64(dst.Bets)*100) // 总返奖率
	}
	if dst.AllRounds > 0 {
		dst.FCashAvg = fmt.Sprintf("%.2f", float64(dst.Cash)/float64(dst.AllRounds)/100) // 局均净赢
	}
	dst.FBets = fmt.Sprintf("%.2f", Chip2Float(dst.Bets))
	dst.FWinBets = fmt.Sprintf("%.2f", Chip2Float(dst.WinBets))
	dst.FLoseBets = fmt.Sprintf("%.2f", Chip2Float(dst.LoseBets))
	dst.FWinBetsAvg = fmt.Sprintf("%.2f", Chip2Float(dst.WinBetsAvg))
	dst.FLoseBetsAvg = fmt.Sprintf("%.2f", Chip2Float(dst.LoseBetsAvg))
	dst.FWins = fmt.Sprintf("%.2f", Chip2Float(dst.Wins))
	dst.FLoses = fmt.Sprintf("%.2f", Chip2Float(dst.Loses))
	dst.FCash = fmt.Sprintf("%.2f", Chip2Float(dst.Cash))
}
