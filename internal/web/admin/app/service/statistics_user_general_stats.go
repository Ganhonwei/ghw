package service

import (
	"fmt"
	"goserver/internal/web/admin/app/entity"
	"goserver/pkg/data/ck"
	"goserver/pkg/utils"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/globalsign/mgo/bson"
)

func (this *statisticsService) UserGeneralStatsUFilters(params map[string]any) (uFilter string, uArgs []any) {
	uFilter = `
		SELECT t1.userid userid, 
			t3.withdraw_amount + t3.withdraw_amount_wait + t3.withdraw_amount_processing + t1.diamond - t2.pay_amount uprofit,
			t3.withdraw_amount - t2.pay_amount uprofit_cash
		FROM game.col_user t0 FINAL 
		JOIN game.col_user_finance t1 FINAL ON t0.userid = t1.userid
		LEFT JOIN (
			SELECT userid, SUM(CASE WHEN order_status = 4 THEN amount ELSE 0 END) pay_amount
			FROM game.col_trade_record FINAL
			GROUP BY userid
		) t2 ON t0.userid = t2.userid
		LEFT JOIN (
			SELECT userid,
				SUM(CASE WHEN order_status = 2 THEN amount ELSE 0 END) withdraw_amount,
				SUM(CASE WHEN order_status = 1 THEN amount ELSE 0 END) withdraw_amount_wait,
				SUM(CASE WHEN order_status in (6,8) THEN amount ELSE 0 END) withdraw_amount_processing
			FROM game.col_withdraw_record FINAL
			GROUP BY userid
		) t3 ON t0.userid = t3.userid
	`
	// 其他条件
	var tables []string
	var where string

	defer func() {
		uFilter += strings.Join(tables, " ")
		uFilter = fmt.Sprintf(`
			SELECT userid FROM (%s)
		`, uFilter+" WHERE 1 = 1 "+where)
	}()

	if userid, ok := params["userid"]; ok {
		where += " and t0.userid = ?"
		uArgs = append(uArgs, userid)
		return
	}

	// 时间
	startTime, ok1 := params["startTime"]
	endTime, ok2 := params["endTime"]
	if ok1 && ok2 {
		where += " and t0.ctime BETWEEN ? AND ?"
		uArgs = append(uArgs, startTime, endTime)
	} else if ok1 {
		where += " and t0.ctime >= ?"
		uArgs = append(uArgs, startTime)
	} else if ok2 {
		where += " and t0.ctime <= ?"
		uArgs = append(uArgs, endTime)
	}

	// params
	utypesV, ok := params["utypes"]
	if ok {
		utypes := utypesV.([]map[string]any)
		var utypeF []string
		for _, p := range utypes {
			// state money1 money2
			var f, sf string
			money1, ok1 := p["money1"]
			money2, ok2 := p["money2"]
			state, ok3 := p["state"]
			if ok3 {
				sf = fmt.Sprintf("and t0.state = %d", state)
				f = fmt.Sprintf("(t0.state = %d)", state)
			}
			if ok1 && ok2 {
				f = fmt.Sprintf("(t1.money between %d and %d %s)", money1, money2, sf)
			} else if ok1 {
				if money1 == 0 {
					f = fmt.Sprintf("(t1.money = 0 %s)", sf)
				} else {
					f = fmt.Sprintf("(t1.money >= %d %s)", money1, sf)
				}
			}
			if f != "" {
				utypeF = append(utypeF, f)
			}
		}
		if len(utypeF) > 0 {
			where += fmt.Sprintf(" and (%s)", strings.Join(utypeF, " or "))
		}
	}

	// 存活天数
	liveDays1, ok1 := params["liveDays1"]
	liveDays2, ok2 := params["liveDays2"]
	if ok1 && ok2 {
		t4 := `
			LEFT JOIN (
				SELECT userid, count(DISTINCT toYYYYMMDD(login_time)) login_days FROM game.col_log_login FINAL GROUP BY userid
			) t4 ON t0.userid = t4.userid
		`
		tables = append(tables, t4)
		where += ` and t4.login_days between ? and ?`
		uArgs = append(uArgs, liveDays1, liveDays2)
	}

	// 流失天数
	lossDays1, ok1 := params["lossDays1"]
	lossDays2, ok2 := params["lossDays2"]
	if ok1 && ok2 {
		now := NowTime()
		where += " and date_diff('day', t0.login_time, ?) between ? and ?"
		uArgs = append(uArgs, now, lossDays1, lossDays2)
	}

	// profit1
	profit1, ok1 := params["profit1"]
	profit2, ok2 := params["profit2"]
	if ok1 && ok2 {
		where += " and uprofit between ? and ?"
		uArgs = append(uArgs, profit1, profit2)
	}

	// profitCash1
	profitCash1, ok1 := params["profitCash1"]
	profitCash2, ok2 := params["profitCash2"]
	if ok1 && ok2 {
		where += " and uprofit_cash between ? and ?"
		uArgs = append(uArgs, profitCash1, profitCash2)
	}

	// money1
	money1, ok1 := params["money1"]
	money2, ok2 := params["money2"]
	if ok1 && ok2 {
		where += " and t2.pay_amount between ? and ?"
		uArgs = append(uArgs, money1, money2)
	}

	t5 := `
	LEFT JOIN (
		SELECT userid, SUM(bet_amount) bets_sum, SUM(bet_amount) bets_avg, SUM(score) score_sum 
		FROM (
			SELECT userid, bet_amount, score, win_type
			FROM game.col_detail t1 FINAL WHERE robot = 0 and win_type in (1,2,3)
			UNION ALL
			SELECT s2.user_id userid, SUM(s2.amount) bet_amount, SUM(s3.amount) - SUM(s2.amount) score, (CASE WHEN score > 0 THEN 1 WHEN score == 0 THEN 3 ELSE 2 END) win_type
			FROM (SELECT t2.round_id, t2.user_id, SUM(t2.amount) amount FROM game.col_nsq_log_external_bet t2 FINAL WHERE t2.amount != 0 GROUP BY round_id, user_id) s2
			LEFT JOIN (SELECT t2.round_id, t2.user_id, SUM(t2.amount) amount FROM game.col_nsq_log_external_reward t2 FINAL WHERE t2.amount != 0 GROUP BY round_id, user_id) s3
				ON s2.round_id = s3.round_id AND s2.user_id = s3.user_id
			GROUP BY s2.round_id, s2.user_id
		) s1
		 GROUP BY s1.userid
	) t5 on t0.userid = t5.userid
	`
	var t5Added bool
	// 打码量
	bets1, ok1 := params["bets1"]
	bets2, ok2 := params["bets2"]
	if ok1 && ok2 {
		if !t5Added {
			tables = append(tables, t5)
			t5Added = true
		}
		where += " and t5.bets_sum between ? and ?"
		uArgs = append(uArgs, bets1, bets2)
	}
	// 注均码
	betAvg1, ok1 := params["betAvg1"]
	betAvg2, ok2 := params["betAvg2"]
	if ok1 && ok2 {
		if !t5Added {
			tables = append(tables, t5)
			t5Added = true
		}
		where += " and t5.bets_avg between ? and ?"
		uArgs = append(uArgs, betAvg1, betAvg2)
	}

	// 净赢
	cash1, ok1 := params["cash1"]
	cash2, ok2 := params["cash2"]
	if ok1 && ok2 {
		if !t5Added {
			tables = append(tables, t5)
			t5Added = true
		}
		where += " and t5.score_sum between ? and ?"
		uArgs = append(uArgs, cash1, cash2)
	}

	// 渠道
	if ad__bundle_id, ok := params["ad__bundle_id"]; ok {
		where += " and t0.ad__bundle_id in ?"
		uArgs = append(uArgs, ad__bundle_id)
	}
	// ab 类
	if regist_area, ok := params["regist_area"]; ok {
		where += " and t0.regist_area in ?"
		uArgs = append(uArgs, regist_area)
	}

	return
}

// 玩家综合情况（经济总况）
func (this *statisticsService) UserGeneralStatsFinance(page, pageSize int, params map[string]any) (total int, list []*entity.UserGeneralStatsFinance, err error) {
	uFilter, uArgs := this.UserGeneralStatsUFilters(params)
	sqlTotal := `
		SELECT count(*) total
		FROM game.col_user s0 FINAL
		JOIN game.col_user_finance s1 FINAL ON s0.userid = s1.userid
		WHERE s0.userid IN (%s)
	`
	sql0 := `
		SELECT s0.userid userid,nickname,s0.ctime ctime,login_time,ad__bundle_id,state,regist_area,
			date_diff('day', s0.ctime, now()) live_days, date_diff('day', login_time, now()) loss_days,
			s0.vbbank vbbank, s0.unlock_bonus, s0.cashout_bonus,
			s1.money money,s1.diamond diamond
		FROM game.col_user s0 FINAL
		JOIN game.col_user_finance s1 FINAL ON s0.userid = s1.userid
		WHERE s0.userid IN (%s)
		ORDER BY s0.ctime DESC
		LIMIT ?, ?
	`
	var args0 []any
	sqlTotal = fmt.Sprintf(sqlTotal, uFilter)
	sql0 = fmt.Sprintf(sql0, uFilter)
	args0 = append(args0, uArgs...)

	// 查总数
	err = ck.Select(&total, sqlTotal, args0...)
	if err != nil {
		return
	}
	if total == 0 {
		return
	}

	offset, limit := PageCalc(page, pageSize)
	var datas0 []map[string]any
	err = ck.Select(&datas0, sql0, append(args0, offset, limit)...)
	if err != nil {
		return
	}
	if len(datas0) == 0 {
		return
	}
	channelMap, err := ChannelService.GetChannelsMap()
	if err != nil {
		return
	}

	var userids []string
	var useridStats = make(map[string]*entity.UserGeneralStatsFinance)
	for i, data := range datas0 {
		userid := data["userid"].(string)
		nickname := data["nickname"].(string)
		ctime := data["ctime"].(time.Time)
		login_time := data["login_time"].(time.Time)
		channelId := data["ad__bundle_id"].(string)
		money := utils.ToInt64(data["money"])
		// pay_amount := utils.ToInt64(data["pay_amount"])
		state := utils.ToInt64(data["state"])
		regist_area := utils.ToInt64(data["regist_area"])
		loss_days := utils.ToInt64(data["loss_days"])
		diamond := utils.ToInt64(data["diamond"])
		vbbank := utils.ToInt64(data["vbbank"])
		unlock_bonus := utils.ToInt64(data["unlock_bonus"])
		cashout_bonus := utils.ToInt64(data["cashout_bonus"])

		userids = append(userids, userid)
		channel := channelMap[channelId]
		_, ctypeName := GetChargeType(money, int(state))
		stat := &entity.UserGeneralStatsFinance{
			No:           strconv.Itoa(i + 1),
			Userid:       userid,
			Nickname:     nickname,
			ChannelClass: channel.ClassName,
			ChannelAlias: channel.Name1,
			UserType:     ctypeName,
			RegistArea:   "",
			Ctime:        ctime.Format(utils.FORMAT),
			LoginTime:    login_time.Format(utils.FORMAT),
			LiveDays:     "",
			LoseDays:     fmt.Sprint(loss_days),
			Diamond:      fmt.Sprintf("%.2f", Chip2Float(diamond)),
			// PayAmount0:   pay_amount,
			// PayAmount:    fmt.Sprintf("%.2f", Chip2Float(pay_amount)),
			GtypeWins: make(map[int64][]int64),
		}
		stat.VipBankUnlockBonus = fmt.Sprintf("%.2f", Chip2Float(vbbank-unlock_bonus))
		stat.VipBankClaimedBonus0 = cashout_bonus
		stat.VipBankClaimedBonus = fmt.Sprintf("%.2f", Chip2Float(cashout_bonus))
		stat.VipBankUnclaimedBonus = fmt.Sprintf("%.2f", Chip2Float(unlock_bonus))
		switch regist_area {
		case 0, 3:
			stat.RegistArea = "A类"
		case 1:
			stat.RegistArea = "B类"
		case 2:
			stat.RegistArea = "C类"
		}
		list = append(list, stat)
		useridStats[userid] = stat
	}

	// 存活天数
	var datas_lives []map[string]any
	err = ck.Select(&datas_lives, `
		SELECT userid, count(DISTINCT toYYYYMMDD(login_time)) login_days 
		FROM game.col_log_login FINAL WHERE userid IN ?
		GROUP BY userid
	`, userids)
	if err != nil {
		return
	}
	for _, data := range datas_lives {
		userid := data["userid"].(string)
		if stat, ok := useridStats[userid]; ok {
			login_days := utils.ToInt64(data["login_days"])
			stat.LiveDays = fmt.Sprint(login_days)
		}
	}

	// 充值提现
	var datas_pays []map[string]any
	err = ck.Select(&datas_pays, `
		SELECT t1.userid userid, t1.diamond, t2.pay_amount, t3.withdraw_amount, t3.withdraw_amount_wait,t3.withdraw_amount_processing,t3.withdraw_amount_freeze,
			t3.withdraw_amount + t3.withdraw_amount_wait + t3.withdraw_amount_processing + t1.diamond - t2.pay_amount profit,
			t3.withdraw_amount - t2.pay_amount profit_cash
		FROM game.col_user_finance t1 FINAL
		LEFT JOIN (
			SELECT userid, SUM(CASE WHEN order_status = 4 THEN amount ELSE 0 END) pay_amount
			FROM game.col_trade_record FINAL
			GROUP BY userid
		) t2 ON t1.userid = t2.userid
		LEFT JOIN (
			SELECT userid,
				SUM(CASE WHEN order_status = 2 THEN amount ELSE 0 END) withdraw_amount,
				SUM(CASE WHEN order_status = 1 THEN amount ELSE 0 END) withdraw_amount_wait,
				SUM(CASE WHEN order_status = 4 THEN amount ELSE 0 END) withdraw_amount_freeze,
				SUM(CASE WHEN order_status in (6,8) THEN amount ELSE 0 END) withdraw_amount_processing
			FROM game.col_withdraw_record FINAL
			GROUP BY userid
		) t3 ON t1.userid = t3.userid
		WHERE t1.userid IN ?
	`, userids)
	if err != nil {
		return
	}
	for _, data := range datas_pays {
		userid := data["userid"].(string)
		if stat, ok := useridStats[userid]; ok {
			diamond := utils.ToInt64(data["diamond"])
			pay_amount := utils.ToInt64(data["pay_amount"])
			withdraw_amount := utils.ToInt64(data["withdraw_amount"])
			withdraw_amount_wait := utils.ToInt64(data["withdraw_amount_wait"])
			withdraw_amount_freeze := utils.ToInt64(data["withdraw_amount_freeze"])
			withdraw_amount_processing := utils.ToInt64(data["withdraw_amount_processing"])
			profit := utils.ToInt64(data["profit"])
			profit_cash := utils.ToInt64(data["profit_cash"])

			stat.PayAmount0 = pay_amount
			stat.PayAmount = fmt.Sprintf("%.2f", Chip2Float(pay_amount))
			stat.WithdrawAmount = fmt.Sprintf("%.2f", Chip2Float(withdraw_amount))
			stat.WithdrawAmountWait = fmt.Sprintf("%.2f", Chip2Float(withdraw_amount_wait))
			stat.WithdrawAmountFreeze = fmt.Sprintf("%.2f", Chip2Float(withdraw_amount_freeze))
			stat.WithdrawAmountProcess = fmt.Sprintf("%.2f", Chip2Float(withdraw_amount_processing))

			// 账面返奖率 （提现成功金额+待审核金额+审核通过未到账金额+携带金额）/总充值金额。百分比，分子保留2位小数
			stat.RebateRate = fmt.Sprintf("%.2f%%", ComputeFloat((withdraw_amount+withdraw_amount_wait+withdraw_amount_processing+diamond), pay_amount)*100)
			stat.Profit0 = profit
			stat.Profit = fmt.Sprintf("%.2f", Chip2Float(profit))
			stat.ProfitCash = fmt.Sprintf("%.2f", Chip2Float(profit_cash))
		}
	}

	// 自研游戏盈利, 返奖率
	var datas_games []map[string]any
	err = ck.Select(&datas_games, `
		SELECT userid, gtype, win_type, SUM(score) score_sum
		FROM game.col_detail t1 FINAL WHERE robot = 0 and win_type in (1,2)
			AND userid IN ?
		GROUP BY userid, gtype, win_type
	
	`, userids)
	if err != nil {
		return
	}
	for _, data := range datas_games {
		userid := data["userid"].(string)
		if stat, ok := useridStats[userid]; ok {
			gtype := utils.ToInt64(data["gtype"])
			win_type := utils.ToInt64(data["win_type"])
			score_sum := utils.ToInt64(data["score_sum"])
			if _, ok := stat.GtypeWins[gtype]; !ok {
				stat.GtypeWins[gtype] = []int64{0, 0}
			}
			stat.ProfitGames0 += score_sum
			switch win_type {
			case 1: // 赢
				stat.GtypeWins[gtype][0] += score_sum
			case 2: // 输
				stat.GtypeWins[gtype][1] += score_sum
			}
		}
	}
	for _, stat := range useridStats {
		stat.ProfitGames = fmt.Sprintf("%.2f", Chip2Float(stat.ProfitGames0))
		for gtype, score := range stat.GtypeWins {
			game := fmt.Sprintf("%.2f", Chip2Float(score[0]+score[1])) + "/" + fmt.Sprintf("%.2f%%", ComputeFloat(score[0], -score[1])*100)
			switch gtype {
			case 1:
				stat.TpScoreStats = game
			case 2:
				stat.LhdScoreStats = game
			case 3:
				stat.UpScoreStats = game
			// case 4:
			// 	if stat.RummyScoreStats == "" {
			// 		stat.RummyScoreStats = game
			// 	}
			case 5:
				stat.Ak47ScoreStats = game
			case 6:
				stat.JokerScoreStats = game
			case 7:
				stat.CrashScoreStats = game
			case 8:
				stat.AbScoreStats = game
			case 9:
				stat.L3ScoreStats = game
			case 10:
				stat.AviatorScoreStats = game
			case 11:
				stat.KingvsqueenScoreStats = game
			case 12:
				stat.RummyScoreStats = game
			}
		}
	}

	// slots 真人盈利
	var datas_slots []map[string]any
	err = ck.Select(&datas_slots, `
		SELECT userid, SUM(score) score_sum,
			(CASE WHEN game_id IN ? THEN 1 ELSE 2 END) gtype
		FROM (
			SELECT s2.user_id userid, s2.game_id, SUM(s3.amount) - SUM(s2.amount) score, (CASE WHEN score > 0 THEN 1 WHEN score == 0 THEN 3 ELSE 2 END) win_type
			FROM (SELECT t2.round_id, t2.user_id, t2.game_id, SUM(t2.amount) amount FROM game.col_nsq_log_external_bet t2 FINAL WHERE t2.amount != 0 GROUP BY round_id, user_id, game_id) s2
			LEFT JOIN (SELECT t2.round_id, t2.user_id, SUM(t2.amount) amount FROM game.col_nsq_log_external_reward t2 FINAL WHERE t2.amount != 0 GROUP BY round_id, user_id) s3
				ON s2.round_id = s3.round_id AND s2.user_id = s3.user_id
			WHERE userid IN ?
			GROUP BY s2.round_id, s2.user_id, s2.game_id
		) s0
		GROUP BY userid, gtype
	`, SlotsGameids, userids)
	if err != nil {
		return
	}
	for _, data := range datas_slots {
		userid := data["userid"].(string)
		if stat, ok := useridStats[userid]; ok {
			gtype := utils.ToInt64(data["gtype"])
			score_sum := utils.ToInt64(data["score_sum"])
			switch gtype {
			case 1:
				stat.ProfitExternalSlots0 = score_sum
				stat.ProfitExternalSlots = fmt.Sprintf("%.2f", Chip2Float(score_sum))
			case 2:
				stat.ProfitExternalZrsx0 = score_sum
				stat.ProfitExternalZrsx = fmt.Sprintf("%.2f", Chip2Float(score_sum))
			default:
				beego.Trace("UserGeneralStatsFinance unknown external score_sum: ", score_sum)
			}
		}
	}

	// 其他领取的cash
	var datas_other_cash []map[string]any
	err = ck.Select(&datas_other_cash, `
		SELECT userid, SUM(add_diamond) add_diamond_sum FROM game.col_log_water FINAL 
		WHERE ltype IN (82,120,104,118) AND userid IN ? 
		GROUP BY userid
	`, userids)
	if err != nil {
		return
	}
	for _, data := range datas_other_cash {
		userid := data["userid"].(string)
		if stat, ok := useridStats[userid]; ok {
			add_diamond_sum := utils.ToInt64(data["add_diamond_sum"])
			stat.OtherCash0 = add_diamond_sum
			stat.OtherCash = fmt.Sprintf("%.2f", Chip2Float(add_diamond_sum))
		}
	}

	pipe_vb := []bson.M{
		{"$match": bson.M{
			"userid": bson.M{"$in": userids},
			"reason": 103,
		}},
		{"$project": bson.M{
			"userid": "$userid",
			"vb":     bson.M{"$subtract": []string{"$after_out_cash", "$before_out_cash"}},
		}},
		{"$group": bson.M{
			"_id": "$userid",
			"vb":  bson.M{"$sum": "$vb"},
		}},
	}
	var r_vb []bson.M
	err = LogVBDiamonds.Pipe(pipe_vb).All(&r_vb)
	if err != nil {
		return
	}
	for _, r := range r_vb {
		userid := r["_id"].(string)
		if stat, ok := useridStats[userid]; ok {
			stat.TryBonus0 = utils.ToInt64(r["vb"])
			stat.TryBonus = fmt.Sprintf("%.2f", Chip2Float(stat.TryBonus0))
		}
	}

	for _, stat := range useridStats {
		// "误差额=账面净盈利-（自研游戏净盈利+外接slots净盈利+外接真人净盈利+打码解锁并领走的金额+其他领取的cash）负数要带符号进去运算。"
		faultCash := stat.Profit0 + stat.PayAmount0 - (stat.ProfitGames0 + stat.ProfitExternalSlots0 + stat.ProfitExternalZrsx0 + stat.VipBankClaimedBonus0 + stat.OtherCash0 - stat.TryBonus0)
		stat.FaultCash = fmt.Sprintf("%.2f", Chip2Float(faultCash))
		stat.VipCashPayRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.VipBankClaimedBonus0, stat.PayAmount0)*100)
	}
	// 汇总
	summary, err := this.UserGeneralStatsFinanceSummary(params, uFilter, uArgs)
	if err != nil {
		return
	}
	list = append([]*entity.UserGeneralStatsFinance{summary}, list...)
	return
}

func (this *statisticsService) UserGeneralStatsFinanceSummary(params map[string]any, uFilter string, uArgs []any) (summary *entity.UserGeneralStatsFinance, err error) {
	mark := "--"
	summary = &entity.UserGeneralStatsFinance{
		No:           fmt.Sprintf("%v-%v总汇", params["startDate"], params["endDate"]),
		Userid:       mark,
		Nickname:     mark,
		ChannelClass: mark,
		ChannelAlias: mark,
		UserType:     mark,
		RegistArea:   mark,
		Ctime:        mark,
		LoginTime:    mark,
		LiveDays:     mark,
		LoseDays:     mark,
		GtypeWins:    make(map[int64][]int64),
	}

	sql0 := `
		SELECT SUM(s0.vbbank) vbbank, SUM(s0.unlock_bonus) unlock_bonus, SUM(s0.cashout_bonus) cashout_bonus, 
			SUM(s1.diamond) diamond
		FROM game.col_user s0 FINAL
		JOIN game.col_user_finance s1 FINAL ON s0.userid = s1.userid
		WHERE s0.userid IN (%s)
	`
	var args0 []any
	sql0 = fmt.Sprintf(sql0, uFilter)
	args0 = append(args0, uArgs...)

	var data map[string]any
	err = ck.Select(&data, sql0, args0...)
	if err != nil {
		return
	}

	diamond := utils.ToInt64(data["diamond"])
	vbbank := utils.ToInt64(data["vbbank"])
	unlock_bonus := utils.ToInt64(data["unlock_bonus"])
	cashout_bonus := utils.ToInt64(data["cashout_bonus"])

	summary.Diamond = fmt.Sprintf("%.2f", Chip2Float(diamond))
	summary.VipBankUnlockBonus = fmt.Sprintf("%.2f", Chip2Float(vbbank-unlock_bonus))
	summary.VipBankClaimedBonus0 = cashout_bonus
	summary.VipBankClaimedBonus = fmt.Sprintf("%.2f", Chip2Float(cashout_bonus))
	summary.VipBankUnclaimedBonus = fmt.Sprintf("%.2f", Chip2Float(unlock_bonus))

	// 充值提现
	var datas_pays map[string]any
	err = ck.Select(&datas_pays, fmt.Sprintf(`
		SELECT SUM(t1.diamond) diamond, SUM(t2.pay_amount) pay_amount, SUM(t3.withdraw_amount) withdraw_amount, SUM(t3.withdraw_amount_wait) withdraw_amount_wait,
			SUM(t3.withdraw_amount_processing) withdraw_amount_processing, SUM(t3.withdraw_amount_freeze) withdraw_amount_freeze,
			SUM(t3.withdraw_amount) + SUM(t3.withdraw_amount_wait) + SUM(t3.withdraw_amount_processing) + SUM(t1.diamond) - SUM(t2.pay_amount) profit,
			SUM(t3.withdraw_amount) - SUM(t2.pay_amount) profit_cash
		FROM game.col_user_finance t1 FINAL
		LEFT JOIN (
			SELECT userid, SUM(CASE WHEN order_status = 4 THEN amount ELSE 0 END) pay_amount
			FROM game.col_trade_record FINAL
			GROUP BY userid
		) t2 ON t1.userid = t2.userid
		LEFT JOIN (
			SELECT userid,
				SUM(CASE WHEN order_status = 2 THEN amount ELSE 0 END) withdraw_amount,
				SUM(CASE WHEN order_status = 1 THEN amount ELSE 0 END) withdraw_amount_wait,
				SUM(CASE WHEN order_status = 4 THEN amount ELSE 0 END) withdraw_amount_freeze,
				SUM(CASE WHEN order_status in (6,8) THEN amount ELSE 0 END) withdraw_amount_processing
			FROM game.col_withdraw_record FINAL
			GROUP BY userid
		) t3 ON t1.userid = t3.userid
		WHERE t1.userid IN (%s)
	`, uFilter), uArgs...)
	if err != nil {
		return
	}
	data = datas_pays
	diamond = utils.ToInt64(data["diamond"])
	pay_amount := utils.ToInt64(data["pay_amount"])
	withdraw_amount := utils.ToInt64(data["withdraw_amount"])
	withdraw_amount_wait := utils.ToInt64(data["withdraw_amount_wait"])
	withdraw_amount_freeze := utils.ToInt64(data["withdraw_amount_freeze"])
	withdraw_amount_processing := utils.ToInt64(data["withdraw_amount_processing"])
	profit := utils.ToInt64(data["profit"])
	profit_cash := utils.ToInt64(data["profit_cash"])

	summary.PayAmount0 = pay_amount
	summary.PayAmount = fmt.Sprintf("%.2f", Chip2Float(pay_amount))
	summary.WithdrawAmount = fmt.Sprintf("%.2f", Chip2Float(withdraw_amount))
	summary.WithdrawAmountWait = fmt.Sprintf("%.2f", Chip2Float(withdraw_amount_wait))
	summary.WithdrawAmountFreeze = fmt.Sprintf("%.2f", Chip2Float(withdraw_amount_freeze))
	summary.WithdrawAmountProcess = fmt.Sprintf("%.2f", Chip2Float(withdraw_amount_processing))

	// 账面返奖率 （提现成功金额+待审核金额+审核通过未到账金额+携带金额）/总充值金额。百分比，分子保留2位小数
	summary.RebateRate = fmt.Sprintf("%.2f%%", ComputeFloat((withdraw_amount+withdraw_amount_wait+withdraw_amount_processing+diamond), pay_amount)*100)
	summary.Profit0 = profit
	summary.Profit = fmt.Sprintf("%.2f", Chip2Float(profit))
	summary.ProfitCash = fmt.Sprintf("%.2f", Chip2Float(profit_cash))

	// 自研游戏盈利, 返奖率
	var datas_games []map[string]any
	err = ck.Select(&datas_games, fmt.Sprintf(`
		SELECT gtype, win_type, SUM(score) score_sum
		FROM game.col_detail t1 FINAL WHERE robot = 0 and win_type in (1,2)
			AND userid IN (%s)
		GROUP BY gtype, win_type
	`, uFilter), uArgs...)
	if err != nil {
		return
	}
	for _, data := range datas_games {
		gtype := utils.ToInt64(data["gtype"])
		win_type := utils.ToInt64(data["win_type"])
		score_sum := utils.ToInt64(data["score_sum"])
		if _, ok := summary.GtypeWins[gtype]; !ok {
			summary.GtypeWins[gtype] = []int64{0, 0}
		}
		summary.ProfitGames0 += score_sum
		switch win_type {
		case 1: // 赢
			summary.GtypeWins[gtype][0] += score_sum
		case 2: // 输
			summary.GtypeWins[gtype][1] += score_sum
		}
	}
	summary.ProfitGames = fmt.Sprintf("%.2f", Chip2Float(summary.ProfitGames0))
	for gtype, score := range summary.GtypeWins {
		game := fmt.Sprintf("%.2f", Chip2Float(score[0]+score[1])) + "/" + fmt.Sprintf("%.2f%%", ComputeFloat(score[0], -score[1])*100)
		switch gtype {
		case 1:
			summary.TpScoreStats = game
		case 2:
			summary.LhdScoreStats = game
		case 3:
			summary.UpScoreStats = game
		case 5:
			summary.Ak47ScoreStats = game
		case 6:
			summary.JokerScoreStats = game
		case 7:
			summary.CrashScoreStats = game
		case 8:
			summary.AbScoreStats = game
		case 9:
			summary.L3ScoreStats = game
		case 10:
			summary.AviatorScoreStats = game
		case 11:
			summary.KingvsqueenScoreStats = game
		case 12:
			summary.RummyScoreStats = game
		}
	}

	// slots 真人盈利
	var datas_slots []map[string]any
	err = ck.Select(&datas_slots, fmt.Sprintf(`
		SELECT SUM(score) score_sum,
			(CASE WHEN game_id IN ? THEN 1 ELSE 2 END) gtype
		FROM (
			SELECT s2.user_id userid, s2.game_id, SUM(s3.amount) - SUM(s2.amount) score, (CASE WHEN score > 0 THEN 1 WHEN score == 0 THEN 3 ELSE 2 END) win_type
			FROM (SELECT t2.round_id, t2.user_id, t2.game_id, SUM(t2.amount) amount FROM game.col_nsq_log_external_bet t2 FINAL WHERE t2.amount != 0 GROUP BY round_id, user_id, game_id) s2
			LEFT JOIN (SELECT t2.round_id, t2.user_id, SUM(t2.amount) amount FROM game.col_nsq_log_external_reward t2 FINAL WHERE t2.amount != 0 GROUP BY round_id, user_id) s3
				ON s2.round_id = s3.round_id AND s2.user_id = s3.user_id
			WHERE userid IN (%s)
			GROUP BY s2.round_id, s2.user_id, s2.game_id
		) s0
		GROUP BY gtype
	`, uFilter), append([]any{SlotsGameids}, uArgs...)...)
	if err != nil {
		return
	}
	for _, data := range datas_slots {
		gtype := utils.ToInt64(data["gtype"])
		score_sum := utils.ToInt64(data["score_sum"])
		switch gtype {
		case 1:
			summary.ProfitExternalSlots0 = score_sum
			summary.ProfitExternalSlots = fmt.Sprintf("%.2f", Chip2Float(score_sum))
		case 2:
			summary.ProfitExternalZrsx0 = score_sum
			summary.ProfitExternalZrsx = fmt.Sprintf("%.2f", Chip2Float(score_sum))
		default:
			beego.Trace("UserGeneralStatsFinance unknown external score_sum: ", score_sum)
		}
	}

	// 其他领取的cash
	var datas_other_cash []map[string]any
	err = ck.Select(&datas_other_cash, fmt.Sprintf(`
		SELECT SUM(add_diamond) add_diamond_sum FROM game.col_log_water FINAL 
		WHERE ltype IN (82,120,104,118) AND userid IN (%s) 
	`, uFilter), uArgs...)
	if err != nil {
		return
	}
	for _, data := range datas_other_cash {
		add_diamond_sum := utils.ToInt64(data["add_diamond_sum"])
		summary.OtherCash0 = add_diamond_sum
		summary.OtherCash = fmt.Sprintf("%.2f", Chip2Float(add_diamond_sum))
	}

	var payUserids []string
	payUFilter := uFilter
	payUFilter, found := strings.CutSuffix(uFilter, ")")
	if found {
		payUFilter += " and t1.money > 0)"
	}
	err = ck.Select(&payUserids, payUFilter, uArgs...)
	if err != nil {
		return
	}
	if len(payUserids) > 0 {
		pipe_vb := []bson.M{
			{"$match": bson.M{
				"userid": bson.M{"$in": payUserids},
				"reason": 103,
			}},
			{"$project": bson.M{
				"vb": bson.M{"$subtract": []string{"$after_out_cash", "$before_out_cash"}},
			}},
			{"$group": bson.M{
				"_id": nil,
				"vb":  bson.M{"$sum": "$vb"},
			}},
		}
		var r_vb []bson.M
		err = LogVBDiamonds.Pipe(pipe_vb).All(&r_vb)
		if err != nil {
			return
		}
		for _, r := range r_vb {
			summary.TryBonus0 = utils.ToInt64(r["vb"])
			summary.TryBonus = fmt.Sprintf("%.2f", Chip2Float(summary.TryBonus0))
		}
	}

	// "误差额=账面净盈利-（自研游戏净盈利+外接slots净盈利+外接真人净盈利+打码解锁并领走的金额+其他领取的cash）负数要带符号进去运算。"
	faultCash := summary.Profit0 + summary.PayAmount0 - (summary.ProfitGames0 + summary.ProfitExternalSlots0 + summary.ProfitExternalZrsx0 + summary.VipBankClaimedBonus0 + summary.OtherCash0 - summary.TryBonus0)
	summary.FaultCash = fmt.Sprintf("%.2f", Chip2Float(faultCash))
	summary.VipCashPayRate = fmt.Sprintf("%.2f%%", ComputeFloat(summary.VipBankClaimedBonus0, summary.PayAmount0)*100)

	return
}

// 玩家综合情况（游戏分布）
func (this *statisticsService) UserGeneralStatsGames(page, pageSize int, params map[string]any) (total int, list []*entity.UserGeneralStatsGames, err error) {
	uFilter, uArgs := this.UserGeneralStatsUFilters(params)
	sqlTotal := `
		SELECT count(*) total
		FROM game.col_user s0 FINAL
		WHERE s0.userid IN (%s)
	`
	sql0 := `
		SELECT s0.userid,nickname,ctime,login_time,ad__bundle_id,state,regist_area,
			date_diff('day', ctime, now()) live_days, date_diff('day', login_time, now()) loss_days
		FROM game.col_user s0 FINAL
		WHERE s0.userid IN (%s)
		ORDER BY s0.ctime DESC
		LIMIT ?, ?
	`
	var args0 []any
	sqlTotal = fmt.Sprintf(sqlTotal, uFilter)
	sql0 = fmt.Sprintf(sql0, uFilter)
	args0 = append(args0, uArgs...)

	// 查总数
	err = ck.Select(&total, sqlTotal, args0...)
	if err != nil {
		return
	}
	if total == 0 {
		return
	}

	offset, limit := PageCalc(page, pageSize)
	var datas0 []map[string]any
	err = ck.Select(&datas0, sql0, append(args0, offset, limit)...)
	if err != nil {
		return
	}
	if len(datas0) == 0 {
		return
	}
	channelMap, err := ChannelService.GetChannelsMap()
	if err != nil {
		return
	}

	var userids []string
	var useridStats = make(map[string]*entity.UserGeneralStatsGames)
	for i, data := range datas0 {
		userid := data["userid"].(string)
		nickname := data["nickname"].(string)
		ctime := data["ctime"].(time.Time)
		login_time := data["login_time"].(time.Time)
		channelId := data["ad__bundle_id"].(string)
		money := utils.ToInt64(data["money"])
		state := utils.ToInt64(data["state"])
		regist_area := utils.ToInt64(data["regist_area"])
		loss_days := utils.ToInt64(data["loss_days"])

		userids = append(userids, userid)
		channel := channelMap[channelId]
		_, ctypeName := GetChargeType(money, int(state))
		stat := &entity.UserGeneralStatsGames{
			No:           strconv.Itoa(i + 1),
			Userid:       userid,
			Nickname:     nickname,
			ChannelClass: channel.ClassName,
			ChannelAlias: channel.Name1,
			UserType:     ctypeName,
			RegistArea:   "",
			Ctime:        ctime.Format(utils.FORMAT),
			LoginTime:    login_time.Format(utils.FORMAT),
			LiveDays:     "",
			LoseDays:     fmt.Sprint(loss_days),
			GtypeStats:   make(map[int64][]int64),
		}
		switch regist_area {
		case 0, 3:
			stat.RegistArea = "A类"
		case 1:
			stat.RegistArea = "B类"
		case 2:
			stat.RegistArea = "C类"
		}
		list = append(list, stat)
		useridStats[userid] = stat
	}

	// 存活天数
	var datas_lives []map[string]any
	err = ck.Select(&datas_lives, `
		SELECT userid, count(DISTINCT toYYYYMMDD(login_time)) login_days 
		FROM game.col_log_login FINAL WHERE userid IN ?
		GROUP BY userid
	`, userids)
	if err != nil {
		return
	}
	for _, data := range datas_lives {
		userid := data["userid"].(string)
		if stat, ok := useridStats[userid]; ok {
			login_days := utils.ToInt64(data["login_days"])
			stat.LiveDays = fmt.Sprint(login_days)
		}
	}

	// 自研游戏
	var datas_games []map[string]any
	err = ck.Select(&datas_games, `
		SELECT userid, gtype, count(*) rounds, SUM(score) score_sum, SUM(bet_amount) bets
		FROM game.col_detail t1 FINAL WHERE robot = 0 and win_type in (1,2,3)
			AND userid IN ?
		GROUP BY userid, gtype
	`, userids)
	if err != nil {
		return
	}
	for _, data := range datas_games {
		userid := data["userid"].(string)
		if stat, ok := useridStats[userid]; ok {
			gtype := utils.ToInt64(data["gtype"])
			rounds := utils.ToInt64(data["rounds"])
			score_sum := utils.ToInt64(data["score_sum"])
			bets := utils.ToInt64(data["bets"])

			stat.AllRounds += rounds
			stat.Bets0 += bets
			stat.Cash0 += score_sum
			stat.RoundsGame += rounds
			stat.BetsGames0 += bets

			if _, ok := stat.GtypeStats[gtype]; !ok {
				stat.GtypeStats[gtype] = []int64{0, 0}
			}
			stat.GtypeStats[gtype][0] = rounds
			stat.GtypeStats[gtype][1] = bets

		}
	}

	// 外接游戏统计
	var datas_slots []map[string]any
	err = ck.Select(&datas_slots, `
		SELECT userid, count(*) rounds, SUM(score) score_sum, SUM(bets) bets,
			(CASE WHEN game_id IN ? THEN 1 ELSE 2 END) gtype
		FROM (
			SELECT s2.user_id userid, s2.game_id, SUM(s2.amount) bets, SUM(s3.amount) - SUM(s2.amount) score, (CASE WHEN score > 0 THEN 1 WHEN score == 0 THEN 3 ELSE 2 END) win_type
			FROM (SELECT t2.round_id, t2.user_id, t2.game_id, SUM(t2.amount) amount FROM game.col_nsq_log_external_bet t2 FINAL WHERE t2.amount != 0 GROUP BY round_id, user_id, game_id) s2
			LEFT JOIN (SELECT t2.round_id, t2.user_id, SUM(t2.amount) amount FROM game.col_nsq_log_external_reward t2 FINAL WHERE t2.amount != 0 GROUP BY round_id, user_id) s3
				ON s2.round_id = s3.round_id AND s2.user_id = s3.user_id
			WHERE userid IN ?
			GROUP BY s2.round_id, s2.user_id, s2.game_id
		) s0
		GROUP BY userid, gtype
	`, SlotsGameids, userids)
	if err != nil {
		return
	}
	for _, data := range datas_slots {
		userid := data["userid"].(string)
		if stat, ok := useridStats[userid]; ok {
			gtype := utils.ToInt64(data["gtype"])
			rounds := utils.ToInt64(data["rounds"])
			score_sum := utils.ToInt64(data["score_sum"])
			bets := utils.ToInt64(data["bets"])

			stat.AllRounds += rounds
			stat.Bets0 += bets
			stat.Cash0 += score_sum

			switch gtype {
			case 1:
				stat.RoundsSlots = rounds
				stat.BetsSlots0 = bets
				stat.BetsSlots = fmt.Sprintf("%.2f", Chip2Float(bets))
			case 2:
				stat.RoundsZrsx = rounds
				stat.BetsZrsx0 = bets
				stat.BetsZrsx = fmt.Sprintf("%.2f", Chip2Float(bets))
			}
		}
	}

	for _, stat := range useridStats {
		stat.Bets = fmt.Sprintf("%.2f", Chip2Float(stat.Bets0))
		stat.BetsAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.Bets0, stat.AllRounds)/100)
		stat.Cash = fmt.Sprintf("%.2f", Chip2Float(stat.Cash0))
		stat.RoundsRateGame = fmt.Sprintf("%.2f%%", ComputeFloat(stat.RoundsGame, stat.AllRounds)*100)
		stat.RoundsRateSlots = fmt.Sprintf("%.2f%%", ComputeFloat(stat.RoundsSlots, stat.AllRounds)*100)
		stat.RoundsRateZrsx = fmt.Sprintf("%.2f%%", ComputeFloat(stat.RoundsZrsx, stat.AllRounds)*100)
		stat.BetsGames = fmt.Sprintf("%.2f", Chip2Float(stat.BetsGames0))
		stat.BetsRateGames = fmt.Sprintf("%.2f%%", ComputeFloat(stat.BetsGames0, stat.Bets0)*100)
		stat.BetsRateSlots = fmt.Sprintf("%.2f%%", ComputeFloat(stat.BetsSlots0, stat.Bets0)*100)
		stat.BetsRateZrsx = fmt.Sprintf("%.2f%%", ComputeFloat(stat.BetsZrsx0, stat.Bets0)*100)

		for gtype, rb := range stat.GtypeStats {
			roundsRate := fmt.Sprint(rb[0]) + "/" + fmt.Sprintf("%.2f%%", ComputeFloat(rb[0], stat.AllRounds)*100)
			betsRate := fmt.Sprintf("%.2f", Chip2Float(rb[1])) + "/" + fmt.Sprintf("%.2f%%", ComputeFloat(rb[1], stat.Bets0)*100)
			switch gtype {
			case 1:
				stat.TpRoundStats = roundsRate
				stat.TpBetsStats = betsRate
			case 2:
				stat.LhdRoundStats = roundsRate
				stat.LhdBetsStats = betsRate
			case 3:
				stat.UpRoundStats = roundsRate
				stat.UpBetsStats = betsRate
			case 5:
				stat.Ak47RoundStats = roundsRate
				stat.Ak47BetsStats = betsRate
			case 6:
				stat.JokerRoundStats = roundsRate
				stat.JokerBetsStats = betsRate
			case 7:
				stat.CrashRoundStats = roundsRate
				stat.CrashBetsStats = betsRate
			case 8:
				stat.AbRoundStats = roundsRate
				stat.AbBetsStats = betsRate
			case 9:
				stat.L3RoundStats = roundsRate
				stat.L3BetsStats = betsRate
			case 10:
				stat.AviatorRoundStats = roundsRate
				stat.AviatorBetsStats = betsRate
			case 11:
				stat.KingvsqueenRoundStats = roundsRate
				stat.KingvsqueenBetsStats = betsRate
			case 12:
				stat.RummyRoundStats = roundsRate
				stat.RummyBetsStats = betsRate
			}
		}
	}

	// 汇总
	summary, err := this.UserGeneralStatsGamesSummary(params, uFilter, uArgs)
	if err != nil {
		return
	}
	list = append([]*entity.UserGeneralStatsGames{summary}, list...)
	return
}

func (this *statisticsService) UserGeneralStatsGamesSummary(params map[string]any, uFilter string, uArgs []any) (summary *entity.UserGeneralStatsGames, err error) {
	mark := "--"
	summary = &entity.UserGeneralStatsGames{
		No:           fmt.Sprintf("%v-%v总汇", params["startDate"], params["endDate"]),
		Userid:       mark,
		Nickname:     mark,
		ChannelClass: mark,
		ChannelAlias: mark,
		UserType:     mark,
		RegistArea:   mark,
		Ctime:        mark,
		LoginTime:    mark,
		LiveDays:     mark,
		LoseDays:     mark,
		GtypeStats:   make(map[int64][]int64),
	}

	// 自研游戏
	var datas_games []map[string]any
	err = ck.Select(&datas_games, fmt.Sprintf(`
		SELECT gtype, count(*) rounds, SUM(score) score_sum, SUM(bet_amount) bets
		FROM game.col_detail t1 FINAL WHERE robot = 0 and win_type in (1,2,3)
			AND userid IN (%s)
		GROUP BY gtype
	`, uFilter), uArgs...)
	if err != nil {
		return
	}
	for _, data := range datas_games {
		gtype := utils.ToInt64(data["gtype"])
		rounds := utils.ToInt64(data["rounds"])
		score_sum := utils.ToInt64(data["score_sum"])
		bets := utils.ToInt64(data["bets"])

		summary.AllRounds += rounds
		summary.Bets0 += bets
		summary.Cash0 += score_sum
		summary.RoundsGame += rounds
		summary.BetsGames0 += bets

		if _, ok := summary.GtypeStats[gtype]; !ok {
			summary.GtypeStats[gtype] = []int64{0, 0}
		}
		summary.GtypeStats[gtype][0] = rounds
		summary.GtypeStats[gtype][1] = bets
	}

	// 外接游戏统计
	var datas_slots []map[string]any
	err = ck.Select(&datas_slots, fmt.Sprintf(`
		SELECT count(*) rounds, SUM(score) score_sum, SUM(bets) bets,
			(CASE WHEN game_id IN ? THEN 1 ELSE 2 END) gtype
		FROM (
			SELECT s2.user_id userid, s2.game_id, SUM(s2.amount) bets, SUM(s3.amount) - SUM(s2.amount) score, (CASE WHEN score > 0 THEN 1 WHEN score == 0 THEN 3 ELSE 2 END) win_type
			FROM (SELECT t2.round_id, t2.user_id, t2.game_id, SUM(t2.amount) amount FROM game.col_nsq_log_external_bet t2 FINAL WHERE t2.amount != 0 GROUP BY round_id, user_id, game_id) s2
			LEFT JOIN (SELECT t2.round_id, t2.user_id, SUM(t2.amount) amount FROM game.col_nsq_log_external_reward t2 FINAL WHERE t2.amount != 0 GROUP BY round_id, user_id) s3
				ON s2.round_id = s3.round_id AND s2.user_id = s3.user_id
			WHERE userid IN (%s)
			GROUP BY s2.round_id, s2.user_id, s2.game_id
		) s0
		GROUP BY gtype
	`, uFilter), append([]any{SlotsGameids}, uArgs...)...)
	if err != nil {
		return
	}
	for _, data := range datas_slots {
		gtype := utils.ToInt64(data["gtype"])
		rounds := utils.ToInt64(data["rounds"])
		score_sum := utils.ToInt64(data["score_sum"])
		bets := utils.ToInt64(data["bets"])

		summary.AllRounds += rounds
		summary.Bets0 += bets
		summary.Cash0 += score_sum

		switch gtype {
		case 1:
			summary.RoundsSlots = rounds
			summary.BetsSlots0 = bets
			summary.BetsSlots = fmt.Sprintf("%.2f", Chip2Float(bets))
		case 2:
			summary.RoundsZrsx = rounds
			summary.BetsZrsx0 = bets
			summary.BetsZrsx = fmt.Sprintf("%.2f", Chip2Float(bets))
		}
	}

	summary.Bets = fmt.Sprintf("%.2f", Chip2Float(summary.Bets0))
	summary.BetsAvg = fmt.Sprintf("%.2f", ComputeFloat(summary.Bets0, summary.AllRounds)/100)
	summary.Cash = fmt.Sprintf("%.2f", Chip2Float(summary.Cash0))
	summary.RoundsRateGame = fmt.Sprintf("%.2f%%", ComputeFloat(summary.RoundsGame, summary.AllRounds)*100)
	summary.RoundsRateSlots = fmt.Sprintf("%.2f%%", ComputeFloat(summary.RoundsSlots, summary.AllRounds)*100)
	summary.RoundsRateZrsx = fmt.Sprintf("%.2f%%", ComputeFloat(summary.RoundsZrsx, summary.AllRounds)*100)
	summary.BetsGames = fmt.Sprintf("%.2f", Chip2Float(summary.BetsGames0))
	summary.BetsRateGames = fmt.Sprintf("%.2f%%", ComputeFloat(summary.BetsGames0, summary.Bets0)*100)
	summary.BetsRateSlots = fmt.Sprintf("%.2f%%", ComputeFloat(summary.BetsSlots0, summary.Bets0)*100)
	summary.BetsRateZrsx = fmt.Sprintf("%.2f%%", ComputeFloat(summary.BetsZrsx0, summary.Bets0)*100)

	for gtype, rb := range summary.GtypeStats {
		roundsRate := fmt.Sprint(rb[0]) + "/" + fmt.Sprintf("%.2f%%", ComputeFloat(rb[0], summary.AllRounds)*100)
		betsRate := fmt.Sprintf("%.2f", Chip2Float(rb[1])) + "/" + fmt.Sprintf("%.2f%%", ComputeFloat(rb[1], summary.Bets0)*100)
		switch gtype {
		case 1:
			summary.TpRoundStats = roundsRate
			summary.TpBetsStats = betsRate
		case 2:
			summary.LhdRoundStats = roundsRate
			summary.LhdBetsStats = betsRate
		case 3:
			summary.UpRoundStats = roundsRate
			summary.UpBetsStats = betsRate
		case 5:
			summary.Ak47RoundStats = roundsRate
			summary.Ak47BetsStats = betsRate
		case 6:
			summary.JokerRoundStats = roundsRate
			summary.JokerBetsStats = betsRate
		case 7:
			summary.CrashRoundStats = roundsRate
			summary.CrashBetsStats = betsRate
		case 8:
			summary.AbRoundStats = roundsRate
			summary.AbBetsStats = betsRate
		case 9:
			summary.L3RoundStats = roundsRate
			summary.L3BetsStats = betsRate
		case 10:
			summary.AviatorRoundStats = roundsRate
			summary.AviatorBetsStats = betsRate
		case 11:
			summary.KingvsqueenRoundStats = roundsRate
			summary.KingvsqueenBetsStats = betsRate
		case 12:
			summary.RummyRoundStats = roundsRate
			summary.RummyBetsStats = betsRate
		}
	}

	return
}

// 玩家综合情况（VBBank）
func (this *statisticsService) UserGeneralStatsVBBank(page, pageSize int, params map[string]any) (total int, list []*entity.UserGeneralStatsVBBank, err error) {
	uFilter, uArgs := this.UserGeneralStatsUFilters(params)
	sqlTotal := `
		SELECT count(*) total
		FROM game.col_user s0 FINAL
		JOIN game.col_user_finance s1 FINAL ON s0.userid = s1.userid
		WHERE s0.userid IN (%s)
	`
	sql0 := `
		SELECT s0.userid userid,nickname,s0.ctime ctime,login_time,ad__bundle_id,state,regist_area,
			date_diff('day', s0.ctime, now()) live_days, date_diff('day', login_time, now()) loss_days,
			s0.vbbank vbbank, s0.unlock_bonus, s0.cashout_bonus, s1.money money,s2.pay_amount pay_amount,
			s0.vip_lv, s0.vip_max_lv
		FROM game.col_user s0 FINAL
		JOIN game.col_user_finance s1 FINAL ON s0.userid = s1.userid
		LEFT JOIN (
			SELECT userid, SUM(CASE WHEN order_status = 4 THEN amount ELSE 0 END) pay_amount
			FROM game.col_trade_record FINAL
			GROUP BY userid
		) s2 on s0.userid = s2.userid
		WHERE s0.userid IN (%s)
		ORDER BY s0.ctime DESC
		LIMIT ?, ?
	`
	var args0 []any
	sqlTotal = fmt.Sprintf(sqlTotal, uFilter)
	sql0 = fmt.Sprintf(sql0, uFilter)
	args0 = append(args0, uArgs...)

	// 查总数
	err = ck.Select(&total, sqlTotal, args0...)
	if err != nil {
		return
	}
	if total == 0 {
		return
	}

	offset, limit := PageCalc(page, pageSize)
	var datas0 []map[string]any
	err = ck.Select(&datas0, sql0, append(args0, offset, limit)...)
	if err != nil {
		return
	}
	if len(datas0) == 0 {
		return
	}
	channelMap, err := ChannelService.GetChannelsMap()
	if err != nil {
		return
	}

	var userids []string
	var useridStats = make(map[string]*entity.UserGeneralStatsVBBank)
	for i, data := range datas0 {
		userid := data["userid"].(string)
		nickname := data["nickname"].(string)
		ctime := data["ctime"].(time.Time)
		login_time := data["login_time"].(time.Time)
		channelId := data["ad__bundle_id"].(string)
		money := utils.ToInt64(data["money"])
		pay_amount := utils.ToInt64(data["pay_amount"])
		state := utils.ToInt64(data["state"])
		regist_area := utils.ToInt64(data["regist_area"])
		loss_days := utils.ToInt64(data["loss_days"])
		// diamond := utils.ToInt64(data["diamond"])
		vbbank := utils.ToInt64(data["vbbank"])
		unlock_bonus := utils.ToInt64(data["unlock_bonus"])
		cashout_bonus := utils.ToInt64(data["cashout_bonus"])
		vip_lv := utils.ToInt64(data["vip_lv"])
		vip_max_lv := utils.ToInt64(data["vip_max_lv"])
		if vip_max_lv < vip_lv {
			vip_max_lv = vip_lv
		}

		userids = append(userids, userid)
		channel := channelMap[channelId]
		_, ctypeName := GetChargeType(money, int(state))
		stat := &entity.UserGeneralStatsVBBank{
			No:           strconv.Itoa(i + 1),
			Userid:       userid,
			Nickname:     nickname,
			ChannelClass: channel.ClassName,
			ChannelAlias: channel.Name1,
			UserType:     ctypeName,
			RegistArea:   "",
			Ctime:        ctime.Format(utils.FORMAT),
			LoginTime:    login_time.Format(utils.FORMAT),
			LiveDays:     "",
			LoseDays:     fmt.Sprint(loss_days),
			PayAmount0:   pay_amount,
			PayAmount:    fmt.Sprintf("%.2f", Chip2Float(pay_amount)),
		}
		stat.VipLevelMax = fmt.Sprint(vip_max_lv)
		stat.VipLevel = fmt.Sprint(vip_lv)
		stat.VipBankBonus0 = vbbank - unlock_bonus
		stat.CashoutBonus0 = cashout_bonus
		stat.UnlockBonus0 = unlock_bonus
		switch regist_area {
		case 0, 3:
			stat.RegistArea = "A类"
		case 1:
			stat.RegistArea = "B类"
		case 2:
			stat.RegistArea = "C类"
		}
		list = append(list, stat)
		useridStats[userid] = stat
	}

	// 存活天数
	var datas_lives []map[string]any
	err = ck.Select(&datas_lives, `
		SELECT userid, count(DISTINCT toYYYYMMDD(login_time)) login_days 
		FROM game.col_log_login FINAL WHERE userid IN ?
		GROUP BY userid
	`, userids)
	if err != nil {
		return
	}
	for _, data := range datas_lives {
		userid := data["userid"].(string)
		if stat, ok := useridStats[userid]; ok {
			login_days := utils.ToInt64(data["login_days"])
			stat.LiveDays = fmt.Sprint(login_days)
		}
	}

	// 打码量
	var datas_bets []map[string]any
	err = ck.Select(&datas_bets, `
		SELECT userid, SUM(bet_amount) bet_sum
		FROM (
			SELECT userid, bet_amount
			FROM game.col_detail t1 FINAL WHERE robot = 0 and win_type in (1,2,3) AND userid IN ?
				UNION ALL
			SELECT user_id userid, SUM(amount) bet_amount
			FROM game.col_nsq_log_external_bet t2 FINAL WHERE amount != 0 AND userid IN ?
			GROUP BY id, user_id
		) s1
		GROUP BY userid
	`, userids, userids)
	if err != nil {
		return
	}
	for _, data := range datas_bets {
		userid := data["userid"].(string)
		if stat, ok := useridStats[userid]; ok {
			bet_sum := utils.ToInt64(data["bet_sum"])
			stat.Bets0 = bet_sum
			stat.Bets = fmt.Sprintf("%.2f", Chip2Float(bet_sum))
			stat.BetsPayRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.Bets0, stat.PayAmount0)*100)
		}
	}

	pipe_vb := []bson.M{
		{"$match": bson.M{
			"userid": bson.M{"$in": userids},
		}},
		{"$project": bson.M{
			"userid": "$userid",
			"reason": "$reason",
			"vb":     bson.M{"$subtract": []string{"$after_out_cash", "$before_out_cash"}},
		}},
		{"$group": bson.M{
			"_id": bson.M{"userid": "$userid", "reason": "$reason"},
			"vb":  bson.M{"$sum": "$vb"},
		}},
	}
	var r_vb []bson.M
	err = LogVBDiamonds.Pipe(pipe_vb).All(&r_vb)
	if err != nil {
		return
	}
	for _, r := range r_vb {
		id := r["_id"].(bson.M)
		userid := id["userid"].(string)
		reason := utils.ToInt64(id["reason"])
		if stat, ok := useridStats[userid]; ok {
			vb := utils.ToInt64(r["vb"])
			vbf := fmt.Sprintf("%.2f", Chip2Float(vb))
			if vb > 0 {
				stat.AllBonus += vb
			}
			switch reason {
			case 103: // B类转正常
				stat.TryBonus0 = vb
				stat.TryBonus = vbf
			case 99: // vip等级奖励
				stat.VipUpgradeBonus0 = vb
				stat.VipUpgradeBonus = vbf
			case 98: // vip每周
				stat.VipWeekBonus0 = vb
				stat.VipWeekBonus = vbf
			case 74, 57, 58, 59, 84, 86, 129, 80, 119: // 充值
				stat.ChargeBonus0 += vb
				// stat.ChargeBonus = vbf
			case 62: // 在线奖励
				stat.OnlineBonus0 = vb
				stat.OnlineBonus = vbf
			case 123: // 任务
				stat.TaskBonus0 = vb
				stat.TaskBonus = vbf
			case 121: // 利息
				stat.InterestBonus0 = vb
				stat.InterestBonus = vbf
			}
		}
	}

	// 分享领取
	pipe_share := []bson.M{
		{"$match": bson.M{
			"userid": bson.M{"$in": userids},
			"ctime":  bson.M{"$gt": 1731979800}, // 2024-11-19 09:30:00 后分享领取转bonus
		}},
		{"$group": bson.M{
			"_id":  bson.M{"userid": "$userid", "gtype": "$gtype"},
			"cash": bson.M{"$sum": "$cash"},
		}},
	}
	var r_share []bson.M
	err = LogShareWaters.Pipe(pipe_share).All(&r_share)
	if err != nil {
		return
	}
	for _, r := range r_share {
		id := r["_id"].(bson.M)
		userid := id["userid"].(string)
		gtype := utils.ToInt64(id["gtype"])
		if stat, ok := useridStats[userid]; ok {
			cash := utils.ToInt64(r["cash"])
			cashf := fmt.Sprintf("%.2f", Chip2Float(cash))
			switch gtype {
			case 1:
				stat.ShareBetsBonus0 = cash
				stat.ShareBetsBonus = cashf
			case 2:
				stat.ShareBonus0 = cash
				stat.ShareBonus = cashf
			}
		}
	}

	for _, stat := range useridStats {
		stat.BonusBetsRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.AllBonus, stat.Bets0)*100)
		stat.BonusPayRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.AllBonus, stat.PayAmount0)*100)
		stat.BetsBonusRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.CashoutBonus0+stat.UnlockBonus0, stat.Bets0)*100)
		stat.VipCashPayRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.CashoutBonus0, stat.PayAmount0)*100)
		stat.VipBonus = fmt.Sprintf("%.2f", Chip2Float(stat.AllBonus))
		stat.VipBankBonus = fmt.Sprintf("%.2f", Chip2Float(stat.VipBankBonus0))
		stat.CashoutBonus = fmt.Sprintf("%.2f", Chip2Float(stat.CashoutBonus0))
		stat.UnlockBonus = fmt.Sprintf("%.2f", Chip2Float(stat.UnlockBonus0))
		stat.FaultBank = fmt.Sprintf("%.2f", Chip2Float(stat.AllBonus-stat.VipBankBonus0-stat.CashoutBonus0-stat.UnlockBonus0))

		stat.ChargeBonus = fmt.Sprintf("%.2f", Chip2Float(stat.ChargeBonus0))
		stat.TryBonusRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.TryBonus0, stat.AllBonus)*100)
		stat.VipUpgradeBonusRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.VipUpgradeBonus0, stat.AllBonus)*100)
		stat.VipWeekBonusRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.VipWeekBonus0, stat.AllBonus)*100)
		stat.ChargeBonusRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.ChargeBonus0, stat.AllBonus)*100)
		stat.OnlineBonusRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.OnlineBonus0, stat.AllBonus)*100)
		stat.TaskBonusRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.TaskBonus0, stat.AllBonus)*100)
		stat.ShareBetsBonusRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.ShareBetsBonus0, stat.AllBonus)*100)
		stat.ShareBonusRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.ShareBonus0, stat.AllBonus)*100)
		stat.FaultBonus = fmt.Sprintf("%.2f", Chip2Float(stat.AllBonus-stat.TryBonus0-stat.VipUpgradeBonus0-stat.VipWeekBonus0-stat.ChargeBonus0-stat.OnlineBonus0-stat.TaskBonus0-stat.ShareBetsBonus0-stat.ShareBonus0-stat.InterestBonus0))
	}

	// 汇总
	summary, err := this.UserGeneralStatsVBBankSummary(params, uFilter, uArgs)
	if err != nil {
		return
	}
	list = append([]*entity.UserGeneralStatsVBBank{summary}, list...)
	return
}

func (this *statisticsService) UserGeneralStatsVBBankSummary(params map[string]any, uFilter string, uArgs []any) (summary *entity.UserGeneralStatsVBBank, err error) {
	mark := "--"
	summary = &entity.UserGeneralStatsVBBank{
		No:           fmt.Sprintf("%v-%v总汇", params["startDate"], params["endDate"]),
		Userid:       mark,
		Nickname:     mark,
		ChannelClass: mark,
		ChannelAlias: mark,
		UserType:     mark,
		RegistArea:   mark,
		Ctime:        mark,
		LoginTime:    mark,
		LiveDays:     mark,
		LoseDays:     mark,
		VipLevelMax:  mark,
		VipLevel:     mark,
	}

	var data map[string]any
	err = ck.Select(&data, fmt.Sprintf(`
		SELECT SUM(s0.vbbank) vbbank, SUM(s0.unlock_bonus) unlock_bonus, SUM(s0.cashout_bonus) cashout_bonus, SUM(s2.pay_amount) pay_amount
		FROM game.col_user s0 FINAL 
		LEFT JOIN (
			SELECT userid, SUM(CASE WHEN order_status = 4 THEN amount ELSE 0 END) pay_amount
			FROM game.col_trade_record FINAL
			GROUP BY userid
		) s2 on s0.userid = s2.userid
		WHERE s0.userid IN (%s)
	`, uFilter), uArgs...)
	if err != nil {
		return
	}

	pay_amount := utils.ToInt64(data["pay_amount"])
	vbbank := utils.ToInt64(data["vbbank"])
	unlock_bonus := utils.ToInt64(data["unlock_bonus"])
	cashout_bonus := utils.ToInt64(data["cashout_bonus"])
	summary.VipBankBonus0 = vbbank - unlock_bonus
	summary.CashoutBonus0 = cashout_bonus
	summary.UnlockBonus0 = unlock_bonus
	summary.PayAmount0 = pay_amount
	summary.PayAmount = fmt.Sprintf("%.2f", Chip2Float(pay_amount))

	// 打码量
	var datas_bets []map[string]any
	err = ck.Select(&datas_bets, fmt.Sprintf(`
		SELECT SUM(bet_amount) bet_sum
		FROM (
			SELECT userid, bet_amount
			FROM game.col_detail t1 FINAL WHERE robot = 0 and win_type in (1,2,3) AND userid IN (%s)
				UNION ALL
			SELECT user_id userid, SUM(amount) bet_amount
			FROM game.col_nsq_log_external_bet t2 FINAL WHERE amount != 0 AND userid IN (%s)
			GROUP BY id, user_id
		) s1
	`, uFilter, uFilter), append(uArgs, uArgs...)...)
	if err != nil {
		return
	}
	for _, data := range datas_bets {
		bet_sum := utils.ToInt64(data["bet_sum"])
		summary.Bets0 = bet_sum
		summary.Bets = fmt.Sprintf("%.2f", Chip2Float(bet_sum))
		summary.BetsPayRate = fmt.Sprintf("%.2f%%", ComputeFloat(summary.Bets0, summary.PayAmount0)*100)
	}

	var allUserids []string
	err = ck.Select(&allUserids, uFilter, uArgs...)
	if err != nil {
		return
	}
	if len(allUserids) == 0 {
		return
	}

	pipe_vb := []bson.M{
		{"$match": bson.M{
			"userid": bson.M{"$in": allUserids},
		}},
		{"$project": bson.M{
			"reason": "$reason",
			"vb":     bson.M{"$subtract": []string{"$after_out_cash", "$before_out_cash"}},
		}},
		{"$group": bson.M{
			"_id": "$reason",
			"vb":  bson.M{"$sum": "$vb"},
		}},
	}
	var r_vb []bson.M
	err = LogVBDiamonds.Pipe(pipe_vb).All(&r_vb)
	if err != nil {
		return
	}
	for _, r := range r_vb {
		reason := utils.ToInt64(r["_id"])
		vb := utils.ToInt64(r["vb"])
		vbf := fmt.Sprintf("%.2f", Chip2Float(vb))
		if vb > 0 {
			summary.AllBonus += vb
		}
		switch reason {
		case 103: // B类转正常
			summary.TryBonus0 = vb
			summary.TryBonus = vbf
		case 99: // vip等级奖励
			summary.VipUpgradeBonus0 = vb
			summary.VipUpgradeBonus = vbf
		case 98: // vip每周
			summary.VipWeekBonus0 = vb
			summary.VipWeekBonus = vbf
		case 74, 57, 58, 59, 84, 86, 129, 80, 119: // 充值
			summary.ChargeBonus0 += vb
			// stat.ChargeBonus = vbf
		case 62: // 在线奖励
			summary.OnlineBonus0 = vb
			summary.OnlineBonus = vbf
		case 123: // 任务
			summary.TaskBonus0 = vb
			summary.TaskBonus = vbf
		case 121: // 利息
			summary.InterestBonus0 = vb
			summary.InterestBonus = vbf
		}
	}

	// 分享领取
	pipe_share := []bson.M{
		{"$match": bson.M{
			"userid": bson.M{"$in": allUserids},
			"ctime":  bson.M{"$gt": 1731979800}, // 2024-11-19 09:30:00 后分享领取转bonus
		}},
		{"$group": bson.M{
			"_id":  "$gtype",
			"cash": bson.M{"$sum": "$cash"},
		}},
	}
	var r_share []bson.M
	err = LogShareWaters.Pipe(pipe_share).All(&r_share)
	if err != nil {
		return
	}
	for _, r := range r_share {
		gtype := utils.ToInt64(r["_id"])
		cash := utils.ToInt64(r["cash"])
		cashf := fmt.Sprintf("%.2f", Chip2Float(cash))
		switch gtype {
		case 1:
			summary.ShareBetsBonus0 = cash
			summary.ShareBetsBonus = cashf
		case 2:
			summary.ShareBonus0 = cash
			summary.ShareBonus = cashf
		}
	}

	summary.BonusBetsRate = fmt.Sprintf("%.2f%%", ComputeFloat(summary.AllBonus, summary.Bets0)*100)
	summary.BonusPayRate = fmt.Sprintf("%.2f%%", ComputeFloat(summary.AllBonus, summary.PayAmount0)*100)
	summary.BetsBonusRate = fmt.Sprintf("%.2f%%", ComputeFloat(summary.CashoutBonus0+summary.UnlockBonus0, summary.Bets0)*100)
	summary.VipCashPayRate = fmt.Sprintf("%.2f%%", ComputeFloat(summary.CashoutBonus0, summary.PayAmount0)*100)
	summary.VipBonus = fmt.Sprintf("%.2f", Chip2Float(summary.AllBonus))
	summary.VipBankBonus = fmt.Sprintf("%.2f", Chip2Float(summary.VipBankBonus0))
	summary.CashoutBonus = fmt.Sprintf("%.2f", Chip2Float(summary.CashoutBonus0))
	summary.UnlockBonus = fmt.Sprintf("%.2f", Chip2Float(summary.UnlockBonus0))
	summary.FaultBank = fmt.Sprintf("%.2f", Chip2Float(summary.AllBonus-summary.VipBankBonus0-summary.CashoutBonus0-summary.UnlockBonus0))

	summary.ChargeBonus = fmt.Sprintf("%.2f", Chip2Float(summary.ChargeBonus0))
	summary.TryBonusRate = fmt.Sprintf("%.2f%%", ComputeFloat(summary.TryBonus0, summary.AllBonus)*100)
	summary.VipUpgradeBonusRate = fmt.Sprintf("%.2f%%", ComputeFloat(summary.VipUpgradeBonus0, summary.AllBonus)*100)
	summary.VipWeekBonusRate = fmt.Sprintf("%.2f%%", ComputeFloat(summary.VipWeekBonus0, summary.AllBonus)*100)
	summary.ChargeBonusRate = fmt.Sprintf("%.2f%%", ComputeFloat(summary.ChargeBonus0, summary.AllBonus)*100)
	summary.OnlineBonusRate = fmt.Sprintf("%.2f%%", ComputeFloat(summary.OnlineBonus0, summary.AllBonus)*100)
	summary.TaskBonusRate = fmt.Sprintf("%.2f%%", ComputeFloat(summary.TaskBonus0, summary.AllBonus)*100)
	summary.ShareBetsBonusRate = fmt.Sprintf("%.2f%%", ComputeFloat(summary.ShareBetsBonus0, summary.AllBonus)*100)
	summary.ShareBonusRate = fmt.Sprintf("%.2f%%", ComputeFloat(summary.ShareBonus0, summary.AllBonus)*100)
	summary.FaultBonus = fmt.Sprintf("%.2f", Chip2Float(summary.AllBonus-summary.TryBonus0-summary.VipUpgradeBonus0-summary.VipWeekBonus0-summary.ChargeBonus0-summary.OnlineBonus0-summary.TaskBonus0-summary.ShareBetsBonus0-summary.ShareBonus0-summary.InterestBonus0))
	return
}
