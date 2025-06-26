package service

import (
	"fmt"
	"goserver/internal/web/admin/app/entity"
	"goserver/pkg/data/ck"
	"goserver/pkg/utils"
	"strconv"
	"strings"
	"time"
)

// GetLHDLkysStat lhd高潮涌现
func (s *gameStatsService) GetLHDGcyxStat(page, pageSize int, params map[string]any) (
	stats []*entity.LHDGcyxStat, total int, err error,
) {
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
		SELECT userid, win_type, count(*) rounds, SUM(bet_amount) bet_sum,
		SUM(CASE WHEN lhd_strategy_id = 4 THEN 1 ELSE 0 END) strategy_round,
		SUM(CASE WHEN lhd_strategy_id = 4 THEN bet_amount ELSE 0 END) strategy_bet_sum,
		SUM(CASE WHEN lhd_strategy_id = 4 THEN score ELSE 0 END) strategy_score_sum,
		SUM(score) score_sum
		FROM game.col_detail FINAL
		WHERE gtype = 2 AND robot = 0 AND win_type IN (1,2,3) %s
		AND userid IN (
			SELECT userid FROM (
				SELECT userid, SUM(bet_amount) bet_total, avg(bet_amount) bet_avg, SUM(score) score_total, max(begin_time) lately_time,
					SUM(CASE WHEN lhd_strategy_id = 4 THEN 1 ELSE 0 END) lhd_gcyx
				FROM game.col_detail FINAL 
				WHERE gtype = 2 AND robot = 0 AND win_type IN (1,2,3) %s
				%s
				GROUP BY userid
				HAVING lhd_gcyx > 0 %s
			) s0 ORDER BY lately_time desc limit ?, ?
		)
		GROUP BY userid, win_type
		ORDER BY max(begin_time) desc
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
			SELECT userid, SUM(bet_amount) bet_total, avg(bet_amount) bet_avg, SUM(score) score_total, max(begin_time) lately_time,
				SUM(CASE WHEN lhd_strategy_id = 4 THEN 1 ELSE 0 END) lhd_gcyx
			FROM game.col_detail FINAL 
			WHERE gtype = 2 AND robot = 0 AND win_type IN (1,2,3) %s
			%s
			GROUP BY userid
			HAVING lhd_gcyx > 0 %s
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
	err = ck.Select(&datas, sql0, append(args, offset, limit)...)
	if err != nil {
		return
	}

	var uidStats = make(map[string]*entity.LHDGcyxStat)
	var userids []string
	var no int
	for _, data := range datas { // 输和赢的统计汇总
		userid := data["userid"].(string)
		stat, ok := uidStats[userid]
		if !ok {
			no++
			stat = &entity.LHDGcyxStat{
				No:     strconv.Itoa(no),
				UserId: userid,
			}
			uidStats[userid] = stat
			stats = append(stats, stat)
			userids = append(userids, userid)
		}
		mappingLHDStatGcyxByWinType1(data, stat)
	}

	// 查用户信息
	var users []map[string]any
	ck.Select(&users, `
		select userid,nickname,ctime,login_time,money,state,regist_area,
			date_diff('day', ctime, now()) live_days, date_diff('day', login_time, now()) loss_days,lhd_gcyx_m
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
		lhd_gcyx_m := utils.ToInt64(user["lhd_gcyx_m"])

		stat, ok := uidStats[userid]
		if ok {
			stat.Nickname = nickname
			stat.Ctime = ctime.Format(utils.FORMAT)
			stat.LoginTime = login_time.Format(utils.FORMAT)
			stat.LiveDays = int(live_days)
			stat.LoseDays = int(loss_days)
			stat.StrategyM = lhd_gcyx_m

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

	// 汇总后二次计算
	for _, stat := range stats {
		mappingLHDStatGcyx0(stat)
	}
	// 平均单次高潮持续局数
	var users_high_avg []map[string]any
	ck.Select(&users_high_avg, `
		SELECT userid, arrayAvg(arrayFilter(x -> x != 0, arrayMap(x -> length(x)-1, arraySplit(x -> x, arrayPushFront(groupArray(gcyx), 1))))) as gcyx_avg
		FROM (
			SELECT userid, (CASE WHEN lhd_strategy_id = 4 THEN 0 ELSE 1 END) gcyx
			FROM game.col_detail FINAL
			WHERE gtype = 2 AND robot = 0 AND win_type IN (1,2,3) AND begin_time > 1730419200
				AND userid IN ?
			ORDER BY begin_time
		)
		GROUP BY userid
	`, userids)
	for _, avg := range users_high_avg {
		userid := avg["userid"].(string)
		if stat, ok := uidStats[userid]; ok {
			stat.StrategyRoundsAvg = fmt.Sprintf("%.2f", utils.ToFloat64(avg["gcyx_avg"]))
		}
	}

	// 查总汇数据
	mark := "--"
	summary := &entity.LHDGcyxStat{
		No:        fmt.Sprintf("%v-%v总汇", params["startDate"], params["endDate"]),
		UserId:    mark,
		Nickname:  mark,
		Ctime:     mark,
		LoginTime: mark,
		StrategyM: 0,
	}
	// 高潮局总M
	err = ck.Select(&(summary.StrategyM), fmt.Sprintf(`
		select SUM(lhd_gcyx_m)
		from game.col_user final where userid in (
			select userid from (
				SELECT userid, SUM(bet_amount) bet_total, avg(bet_amount) bet_avg, SUM(score) score_total, max(begin_time) lately_time,
					SUM(CASE WHEN lhd_strategy_id = 4 THEN 1 ELSE 0 END) lhd_gcyx
				FROM game.col_detail FINAL 
				WHERE gtype = 2 AND robot = 0 AND win_type IN (1,2,3) %s
				%s
				GROUP BY userid
				HAVING lhd_gcyx > 0 %s
			)
		)
	`, where2, uFilterSql3, having4), append(append(args2, args3...), args4...)...)
	if err != nil {
		return
	}
	// 平均单次高潮持续局数
	var summaryStrategyRoundsAvg float64
	err = ck.Select(&summaryStrategyRoundsAvg, fmt.Sprintf(`
		SELECT AVG(arrayJoin(gcyx_arr)) gcyx_avg FROM (
			SELECT userid, arrayFilter(x -> x != 0, arrayMap(x -> length(x)-1, arraySplit(x -> x, arrayPushFront(groupArray(gcyx), 1)))) as gcyx_arr
			FROM (
				SELECT userid, (CASE WHEN lhd_strategy_id = 4 THEN 0 ELSE 1 END) gcyx
				FROM game.col_detail FINAL
				WHERE gtype = 2 AND robot = 0 AND win_type IN (1,2,3) AND begin_time > 1730419200
					AND userid IN (
						select userid from (
							SELECT userid, SUM(bet_amount) bet_total, avg(bet_amount) bet_avg, SUM(score) score_total, max(begin_time) lately_time,
								SUM(CASE WHEN lhd_strategy_id = 4 THEN 1 ELSE 0 END) lhd_gcyx
							FROM game.col_detail FINAL 
							WHERE gtype = 2 AND robot = 0 AND win_type IN (1,2,3) %s
							%s
							GROUP BY userid
							HAVING lhd_gcyx > 0 %s
						)
					)
				ORDER BY begin_time
			)
			GROUP BY userid
		) s1
	`, where2, uFilterSql3, having4), append(append(args2, args3...), args4...)...)
	if err != nil {
		return
	}
	summary.StrategyRoundsAvg = fmt.Sprintf("%.2f", summaryStrategyRoundsAvg)

	sqlSummary := strings.ReplaceAll(sql0, "SELECT userid, win_type", "SELECT win_type")
	sqlSummary = strings.ReplaceAll(sqlSummary, "limit ?, ?", "")
	sqlSummary = strings.ReplaceAll(sqlSummary, "GROUP BY userid, win_type", "GROUP BY win_type")
	var summaryDatas []map[string]any
	err = ck.Select(&summaryDatas, sqlSummary, args...)
	if err != nil {
		return
	}
	for _, data := range summaryDatas {
		mappingLHDStatGcyxByWinType1(data, summary)
	}
	// 汇总后二次计算
	mappingLHDStatGcyx0(summary)
	stats = append([]*entity.LHDGcyxStat{summary}, stats...)
	return
}

func mappingLHDStatGcyxByWinType1(src map[string]any, dst *entity.LHDGcyxStat) {
	win_type := utils.ToInt64(src["win_type"])
	dst.AllRounds += utils.ToInt64(src["rounds"])
	dst.Bets += utils.ToInt64(src["bet_sum"])

	// 策略
	dst.StrategyRounds += utils.ToInt64(src["strategy_round"])
	dst.StrategyBets += utils.ToInt64(src["strategy_bet_sum"])

	if win_type == 1 { // 赢
		dst.WinRounds = utils.ToInt64(src["rounds"])
		dst.StrategyWinRounds = utils.ToInt64(src["strategy_round"])
		dst.StrategyWins = utils.ToInt64(src["strategy_score_sum"])
		dst.FStrategyWins = fmt.Sprintf("%.2f", float64(utils.ToInt64(src["strategy_score_sum"]))/100)
		dst.StrategyWinBetAvg = fmt.Sprintf("%.2f", ComputeFloat(utils.ToInt64(src["strategy_bet_sum"]), utils.ToInt64(src["strategy_round"]))/100)
		dst.Wins = utils.ToInt64(src["score_sum"])
	} else if win_type == 2 { // 输
		dst.StrategyLoses = utils.ToInt64(src["strategy_score_sum"])
		dst.FStrategyLoses = fmt.Sprintf("%.2f", float64(utils.ToInt64(src["strategy_score_sum"]))/100)
		dst.StrategyLoseBetAvg = fmt.Sprintf("%.2f", ComputeFloat(utils.ToInt64(src["strategy_bet_sum"]), utils.ToInt64(src["strategy_round"]))/100)
		dst.Loses = utils.ToInt64(src["score_sum"])
	}
}

func mappingLHDStatGcyx0(dst *entity.LHDGcyxStat) {
	dst.FBets = fmt.Sprintf("%.2f", Chip2Float(dst.Bets))
	dst.FStrategyBets = fmt.Sprintf("%.2f", Chip2Float(dst.StrategyBets))
	if dst.AllRounds > 0 {
		dst.BetsAvg = fmt.Sprintf("%.2f", float64(dst.Bets)/float64(dst.AllRounds)/100) // 局均打码量
	}
	if dst.StrategyRounds > 0 {
		dst.StrategyBetsAvg = fmt.Sprintf("%.2f", ComputeFloat(dst.StrategyBets, dst.StrategyRounds)/100)
	}

	dst.StrategyRate = fmt.Sprintf("%.1f", ComputeFloat(dst.AllRounds, dst.StrategyM))
	dst.StrategyWinRate = fmt.Sprintf("%.2f%%", ComputeFloat(dst.StrategyWinRounds, dst.StrategyRounds)*100)
	if dst.StrategyLoses == 0 {
		dst.StrategyRewardRate = "没输过"
	} else {
		dst.StrategyRewardRate = fmt.Sprintf("%.2f%%", ComputeFloat(dst.StrategyWins, -dst.StrategyLoses)*100)
	}
	dst.LHDRebateRate = fmt.Sprintf("%.2f%%", ComputeFloat(dst.Wins, -dst.Loses)*100)
	dst.LHDWinRate = fmt.Sprintf("%.2f%%", ComputeFloat(dst.WinRounds, dst.AllRounds)*100)
	dst.StrategyCash = fmt.Sprintf("%.2f", Chip2Float(dst.StrategyWins+dst.StrategyLoses))
}
