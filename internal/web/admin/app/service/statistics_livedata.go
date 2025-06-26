package service

import (
	"context"
	"fmt"
	"goserver/internal/web/admin/app/entity"
	"goserver/pkg/data/ck"
	"goserver/pkg/utils"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// 实时数据监控
func (s *statisticsService) LiveDataHistoryStats() (users, bets int64, err error) {
	ctx := context.Background()
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return ck.Select(&users, `
			select count(*) users from game.col_user final where robot = 0 and simulation_robot = 0
		`)
	})

	g.Go(func() error {
		return ck.Select(&bets, `
			SELECT SUM(amounts) bets from (
				select sum(bet_amount) amounts from game.col_detail final where robot = 0
				union all
				SELECT SUM(amount) amounts FROM game.col_nsq_log_external_bet FINAL WHERE amount != 0
			) t1
		`)
	})
	if err = g.Wait(); err != nil {
		return
	}
	return
}

// 实时数据监控
func (s *statisticsService) LiveDataTodayStats() (stats *entity.LiveDataTodayStats, err error) {
	stats = &entity.LiveDataTodayStats{}

	now := time.Now().In(location)
	stime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	etime := stime.AddDate(0, 0, 1)

	ctx := context.Background()
	g, ctx := errgroup.WithContext(ctx)
	// 累计新注册
	g.Go(func() error {
		return ck.Select(&stats.Users, `
			select count(*) users from game.col_user final 
			where ctime >= ? and ctime < ? and robot = 0 and simulation_robot = 0
		`, stime, etime)
	})

	// 累计充值金额
	g.Go(func() error {
		var pay_datas = make(map[string]any)
		err := ck.Select(&pay_datas, `
			select SUM(amount) pays, count(distinct userid) pay_users
			from game.col_trade_record where ctime >= ? and ctime < ? and order_status = 4
		`, stime, etime)
		if err != nil {
			return err
		}
		stats.Pays = Chip2Float(utils.ToInt64(pay_datas["pays"]))
		stats.PayUsers = utils.ToInt64(pay_datas["pay_users"])
		return nil
	})

	// 累计提现金额
	g.Go(func() error {
		var withdraw_datas = make(map[string]any)
		err := ck.Select(&withdraw_datas, `
			select SUM(amount + commission) withdraws, count(distinct userid) withdraw_users
			from game.col_withdraw_record final where ctime >= ? and ctime < ? and order_status = 2
		`, stime, etime)
		if err != nil {
			return err
		}
		withdraws := utils.ToInt64(withdraw_datas["withdraws"])
		withdraw_users := utils.ToInt64(withdraw_datas["withdraw_users"])
		stats.Withdraws = Chip2Float(withdraws)
		stats.WithdrawUsers = withdraw_users
		return nil
	})

	// 累计打码量
	g.Go(func() error {
		var game_datas = make(map[string]any)
		err := ck.Select(&game_datas, `
			SELECT SUM(bets) bets, SUM(rewards) rewards from (
				select sum(bet_amount) bets, SUM(score+bet_amount) rewards from game.col_detail final where robot = 0 and begin_time >= ? and begin_time < ?
				
				UNION ALL

				select SUM(s2.amount_sum) bets, SUM(s3.amount_sum) rewards
				from (
					SELECT round_id, user_id, SUM(amount) amount_sum
					FROM game.col_nsq_log_external_reward FINAL WHERE amount != 0 GROUP BY round_id, user_id
				) s3 right join (
					SELECT round_id, user_id, game_id, SUM(amount) amount_sum, MIN(ctime) begin_time 
					FROM game.col_nsq_log_external_bet FINAL WHERE ctime > ? and ctime < ? and amount != 0
					GROUP BY round_id, user_id, game_id
				) s2 ON s2.round_id = s3.round_id AND s2.user_id = s3.user_id
			) t1
		`, stime.Unix(), etime.Unix(), stime.Unix(), etime.Unix())
		if err != nil {
			return err
		}
		bets := utils.ToInt64(game_datas["bets"])
		rewards := utils.ToInt64(game_datas["rewards"])
		stats.Bets = Chip2Float(bets)
		stats.RabateRate = fmt.Sprintf("%.2f%%", ComputeFloat(rewards, bets)*100)
		return nil
	})

	// 实时在线人数
	g.Go(func() error {
		var online_data = make(map[string]any)
		err := ck.Select(&online_data, `
			select ctime, online_users from game.col_log_online_users final where id = (
				select max(id) from game.col_log_online_users final
			)
		`)
		if err != nil {
			return err
		}
		ctime := utils.ToInt64(online_data["ctime"])
		online_users := utils.ToInt64(online_data["online_users"])
		// 先2分钟内数据有效
		if time.Now().Unix()-ctime < 120 {
			stats.OnlineUsers = online_users
		}
		return nil
	})

	if err = g.Wait(); err != nil {
		return
	}

	return
}

// 实时数据监控-走势曲线 (主要是从注册、在线、充值（以成功算）、提现（以成功算)、打码三个维度来监控的额)
// trends:
// 1.实时注册人数、2.实时在线人数、3.实时充值人数、4.实时充值单数、5.实时充值金额、6.实时提现人数、7.实时提现单数、8.实时提现金额
// 9.实时打码人数、10.实时打码量、11.累计注册人数、12.累计充值金额、13.累计提现金额、14.累计打码人数、15.累计打码量、15.累计充投比
func (s *statisticsService) LiveDataTrendStats(stime, etime time.Time, trends []int32) (trendStats []*entity.LiveDataTrendStats, err error) {
	// 去重避免重复计算
	calcedTrend := make(map[int32]bool)
	for _, trend := range trends {
		calcedTrend[trend] = true
	}
	trends = trends[:0]
	for trend := range calcedTrend {
		trends = append(trends, trend)
	}

	lock := &sync.Mutex{}
	var appendTrendStats = func(stats []*entity.LiveDataTrendStats) {
		lock.Lock()
		defer lock.Unlock()
		trendStats = append(trendStats, stats...)
	}

	ctx := context.Background()
	g, _ := errgroup.WithContext(ctx)
	var calcTrendStat = func(calc func(stime, etime time.Time) (stats []*entity.LiveDataTrendStats, err error)) {
		g.Go(func() error {
			stats, err := calc(stime, etime)
			if err != nil {
				return err
			}
			appendTrendStats(stats)
			return nil
		})
	}

	for _, trend := range trends {
		switch trend {
		case 1: // 实时注册人数
			calcTrendStat(s.liveDataTrendStat_1)
		case 2: // 实时在线人数
			calcTrendStat(s.liveDataTrendStat_2)
		case 3: // 实时充值人数
			calcTrendStat(s.liveDataTrendStat_3)
		case 4: // 实时充值单数
			calcTrendStat(s.liveDataTrendStat_4)
		case 5: // 实时充值金额
			calcTrendStat(s.liveDataTrendStat_5)
		case 6: // 实时提现人数
			calcTrendStat(s.liveDataTrendStat_6)
		case 7: // 实时提现单数
			calcTrendStat(s.liveDataTrendStat_7)
		case 8: // 实时提现金额
			calcTrendStat(s.liveDataTrendStat_8)
		case 9: // 实时打码人数
			calcTrendStat(s.liveDataTrendStat_9)
		case 10: // 实时打码量
			calcTrendStat(s.liveDataTrendStat_10)
		case 11: // 累计注册人数
			calcTrendStat(s.liveDataTrendStat_11)
		case 12: // 累计充值金额
			calcTrendStat(s.liveDataTrendStat_12)
		case 13: // 累计提现金额
			calcTrendStat(s.liveDataTrendStat_13)
		case 14: // 累计打码人数
			calcTrendStat(s.liveDataTrendStat_14)
		case 15: // 累计打码量
			calcTrendStat(s.liveDataTrendStat_15)
		case 16: // 累计充投比
			calcTrendStat(s.liveDataTrendStat_16)
		case 17: // 新用户累计充投比
			calcTrendStat(s.liveDataTrendStat_17)
		case 18: // 老用户累计充投比
			calcTrendStat(s.liveDataTrendStat_18)
		}
	}
	if err = g.Wait(); err != nil {
		return
	}
	return
}

func calcMinites(s_hhmm, e_hhmm uint32) (minutes int) {
	if s_hhmm > e_hhmm {
		return
	}
	shour, ehour := s_hhmm/100, e_hhmm/100
	sminute, eminute := s_hhmm%100, e_hhmm%100

	minutes += int(60 * (ehour - shour))
	minutes += int(eminute - sminute)
	return
}

func liveDataTrendDatas2Stats(trend int32, stime, etime time.Time, datas []map[string]any, key string) (stats []*entity.LiveDataTrendStats) {
	datasMap := make(map[uint32]map[uint32]float64)

	for _, data := range datas {
		day_time := data["day_time"].(uint32)
		minute_time := data["minute_time"].(uint32)
		value := utils.ToFloat64(data[key])
		stat, ok := datasMap[day_time]
		if !ok {
			stat = make(map[uint32]float64)
			datasMap[day_time] = stat
		}
		stat[minute_time] = value
	}

	// 是否累加指标
	addUp := utils.SliceIn(trend, 11, 12, 13, 14, 15)

	now := NowTime()
	cur_day_time := uint32(now.Year()*10000 + int(now.Month())*100 + now.Day())
	cur_minute_time := uint32(now.Hour()*100 + now.Minute())

	for ; stime.Before(etime); stime = stime.AddDate(0, 0, 1) {
		day_time := uint32(stime.Year()*10000 + int(stime.Month())*100 + stime.Day())
		stat := &entity.LiveDataTrendStats{Trend: trend, DayTime: day_time}
		sday := fmt.Sprint(day_time)
		sdate := sday[0:4] + "-" + sday[4:6] + "-" + sday[6:8]
		var name string
		switch trend {
		case 1:
			name = "实时注册人数"
		case 2:
			name = "实时在线人数"
		case 3:
			name = "实时充值人数"
		case 4:
			name = "实时充值单数"
		case 5:
			name = "实时充值金额"
		case 6:
			name = "实时提现人数"
		case 7:
			name = "实时提现单数"
		case 8:
			name = "实时提现金额"
		case 9:
			name = "实时打码人数"
		case 10:
			name = "实时打码量"
		case 11:
			name = "累计注册人数"
		case 12:
			name = "累计充值金额"
		case 13:
			name = "累计提现金额"
		case 14:
			name = "累计打码人数"
		case 15:
			name = "累计打码量"
		}
		stat.Name = sdate + "-" + name

		stats = append(stats, stat)
		dataMap, ok := datasMap[day_time]
		if !ok {
			dataMap = make(map[uint32]float64)
		}

		var add float64
		var hour, minute uint32
		for ; hour < 24; hour++ {
			for minute = range 60 {
				minute_time := hour*100 + minute
				// 时间未到
				if day_time > cur_day_time || (day_time == cur_day_time && minute_time > cur_minute_time) {
					continue
				}

				data := dataMap[minute_time]
				if addUp {
					add += data
					data = add
				}
				stat.Datas = append(stat.Datas, data)
			}
		}
	}
	return
}

// 1.实时注册人数
func (s *statisticsService) liveDataTrendStat_1(stime, etime time.Time) (stats []*entity.LiveDataTrendStats, err error) {
	var datas []map[string]any
	err = ck.Select(&datas, `
		select toYYYYMMDD(ctime) day_time, (toHour(ctime)*100 + toMinute(ctime)) minute_time, count(*) users
		from game.col_user final
		where ctime between ? AND ?
		group by day_time, minute_time
	`, stime, etime)
	if err != nil {
		return
	}

	stats = liveDataTrendDatas2Stats(1, stime, etime, datas, "users")
	return
}

// 2.实时在线人数
func (s *statisticsService) liveDataTrendStat_2(stime, etime time.Time) (stats []*entity.LiveDataTrendStats, err error) {
	var datas []map[string]any
	err = ck.Select(&datas, `
		select day_time, minute_time, MAX(online_users) users
		from game.col_log_online_users final
		where ctime between ? and ?
		group by day_time, minute_time
	`, stime.Unix(), etime.Unix())
	if err != nil {
		return
	}
	stats = liveDataTrendDatas2Stats(2, stime, etime, datas, "users")
	return
}

// 3.实时充值人数
func (s *statisticsService) liveDataTrendStat_3(stime, etime time.Time) (stats []*entity.LiveDataTrendStats, err error) {
	var datas []map[string]any
	err = ck.Select(&datas, `
		select toYYYYMMDD(ctime) day_time, (toHour(ctime)*100 + toMinute(ctime)) minute_time, count(distinct userid) users
		from game.col_trade_record final
		where ctime between ? and ? and order_status = 4
		group by day_time, minute_time
	`, stime, etime)
	if err != nil {
		return
	}

	stats = liveDataTrendDatas2Stats(3, stime, etime, datas, "users")
	return
}

// 4.实时充值单数
func (s *statisticsService) liveDataTrendStat_4(stime, etime time.Time) (stats []*entity.LiveDataTrendStats, err error) {
	var datas []map[string]any
	err = ck.Select(&datas, `
		select toYYYYMMDD(ctime) day_time, (toHour(ctime)*100 + toMinute(ctime)) minute_time, count(*) orders
		from game.col_trade_record final
		where ctime between ? and ? and order_status = 4
		group by day_time, minute_time
	`, stime, etime)
	if err != nil {
		return
	}

	stats = liveDataTrendDatas2Stats(4, stime, etime, datas, "orders")
	return
}

// 5.实时充值金额
func (s *statisticsService) liveDataTrendStat_5(stime, etime time.Time) (stats []*entity.LiveDataTrendStats, err error) {
	var datas []map[string]any
	err = ck.Select(&datas, `
		select toYYYYMMDD(ctime) day_time, (toHour(ctime)*100 + toMinute(ctime)) minute_time, SUM(amount)/100 amounts
		from game.col_trade_record final
		where ctime between ? and ? and order_status = 4
		group by day_time, minute_time
	`, stime, etime)
	if err != nil {
		return
	}
	stats = liveDataTrendDatas2Stats(5, stime, etime, datas, "amounts")
	return
}

// 6.实时提现人数
func (s *statisticsService) liveDataTrendStat_6(stime, etime time.Time) (stats []*entity.LiveDataTrendStats, err error) {
	var datas []map[string]any
	err = ck.Select(&datas, `
		select toYYYYMMDD(ctime) day_time, (toHour(ctime)*100 + toMinute(ctime)) minute_time, count(distinct userid) users
		from game.col_withdraw_record final
		where ctime between ? and ? and order_status = 2
		group by day_time, minute_time
	`, stime, etime)
	if err != nil {
		return
	}
	stats = liveDataTrendDatas2Stats(6, stime, etime, datas, "users")
	return
}

// 7.实时提现单数
func (s *statisticsService) liveDataTrendStat_7(stime, etime time.Time) (stats []*entity.LiveDataTrendStats, err error) {
	var datas []map[string]any
	err = ck.Select(&datas, `
		select toYYYYMMDD(ctime) day_time, (toHour(ctime)*100 + toMinute(ctime)) minute_time, count(*) orders
		from game.col_withdraw_record final
		where ctime between ? and ? and order_status = 2
		group by day_time, minute_time
	`, stime, etime)
	if err != nil {
		return
	}
	stats = liveDataTrendDatas2Stats(7, stime, etime, datas, "orders")
	return
}

// 8.实时提现金额
func (s *statisticsService) liveDataTrendStat_8(stime, etime time.Time) (stats []*entity.LiveDataTrendStats, err error) {
	var datas []map[string]any
	err = ck.Select(&datas, `
		select toYYYYMMDD(ctime) day_time, (toHour(ctime)*100 + toMinute(ctime)) minute_time, SUM(amount + commission)/100 amounts
		from game.col_withdraw_record final
		where ctime between ? and ? and order_status = 2
		group by day_time, minute_time
	`, stime, etime)
	if err != nil {
		return
	}
	stats = liveDataTrendDatas2Stats(8, stime, etime, datas, "amounts")
	return
}

// 9.实时打码人数
func (s *statisticsService) liveDataTrendStat_9(stime, etime time.Time) (stats []*entity.LiveDataTrendStats, err error) {
	var datas []map[string]any
	err = ck.Select(&datas, `
		select day_time, minute_time, sum(users) users from (
			select toYYYYMMDD(toDateTime(begin_time)) day_time, (toHour(toDateTime(begin_time))*100 + toMinute(toDateTime(begin_time))) minute_time, count(distinct userid) users 
			from game.col_detail final where robot = 0 and begin_time >= ? and begin_time < ?
			group by day_time, minute_time
			
			union all
			
			select toYYYYMMDD(toDateTime(ctime)) day_time, (toHour(toDateTime(ctime))*100 + toMinute(toDateTime(ctime))) minute_time, count(distinct user_id) users
			from game.col_nsq_log_external_bet final where ctime > ? and ctime < ? and amount != 0
			group by day_time, minute_time
		) s0
		group by day_time, minute_time
	`, stime.Unix(), etime.Unix(), stime.Unix(), etime.Unix())
	if err != nil {
		return
	}
	stats = liveDataTrendDatas2Stats(9, stime, etime, datas, "users")
	return
}

// 10.实时打码量
func (s *statisticsService) liveDataTrendStat_10(stime, etime time.Time) (stats []*entity.LiveDataTrendStats, err error) {
	var datas []map[string]any
	err = ck.Select(&datas, `
		select day_time, minute_time, sum(bets)/100 bets from (
			select toYYYYMMDD(toDateTime(begin_time)) day_time, (toHour(toDateTime(begin_time))*100 + toMinute(toDateTime(begin_time))) minute_time, sum(bet_amount) bets 
			from game.col_detail final where robot = 0 and begin_time >= ? and begin_time < ?
			group by day_time, minute_time
			
			union all
			
			select toYYYYMMDD(toDateTime(ctime)) day_time, (toHour(toDateTime(ctime))*100 + toMinute(toDateTime(ctime))) minute_time, sum(amount) bets 
			from game.col_nsq_log_external_bet final where ctime > ? and ctime < ? and amount != 0
			group by day_time, minute_time
		) s0
		group by day_time, minute_time
	`, stime.Unix(), etime.Unix(), stime.Unix(), etime.Unix())
	if err != nil {
		return
	}
	stats = liveDataTrendDatas2Stats(10, stime, etime, datas, "bets")
	return
}

// 11.累计注册人数
func (s *statisticsService) liveDataTrendStat_11(stime, etime time.Time) (stats []*entity.LiveDataTrendStats, err error) {
	var datas []map[string]any
	err = ck.Select(&datas, `
		select toYYYYMMDD(ctime) day_time, (toHour(ctime)*100 + toMinute(ctime)) minute_time, count(*) users
		from game.col_user final
		where ctime between ? AND ?
		group by day_time, minute_time
	`, stime, etime)
	if err != nil {
		return
	}

	stats = liveDataTrendDatas2Stats(11, stime, etime, datas, "users")
	return
}

// 12.累计充值金额
func (s *statisticsService) liveDataTrendStat_12(stime, etime time.Time) (stats []*entity.LiveDataTrendStats, err error) {
	var datas []map[string]any
	err = ck.Select(&datas, `
		select toYYYYMMDD(ctime) day_time, (toHour(ctime)*100 + toMinute(ctime)) minute_time, SUM(amount)/100 amounts
		from game.col_trade_record final
		where ctime between ? and ? and order_status = 4
		group by day_time, minute_time
	`, stime, etime)
	if err != nil {
		return
	}
	stats = liveDataTrendDatas2Stats(12, stime, etime, datas, "amounts")
	return
}

// 13.累计提现金额
func (s *statisticsService) liveDataTrendStat_13(stime, etime time.Time) (stats []*entity.LiveDataTrendStats, err error) {
	var datas []map[string]any
	err = ck.Select(&datas, `
		select toYYYYMMDD(ctime) day_time, (toHour(ctime)*100 + toMinute(ctime)) minute_time, SUM(amount + commission)/100 amounts
		from game.col_withdraw_record final
		where ctime between ? and ? and order_status = 2
		group by day_time, minute_time
	`, stime, etime)
	if err != nil {
		return
	}
	stats = liveDataTrendDatas2Stats(13, stime, etime, datas, "amounts")
	return
}

// 14.累计打码人数
func (s *statisticsService) liveDataTrendStat_14(stime, etime time.Time) (stats []*entity.LiveDataTrendStats, err error) {
	var datas []map[string]any
	err = ck.Select(&datas, `
		select toYYYYMMDD(ctime) day_time, (toHour(ctime)*100 + toMinute(ctime)) minute_time, groupArray(userid) userids from (
			select toDateTime(begin_time) ctime, userid from game.col_detail final where robot = 0 and begin_time >= ? and begin_time < ?
			union all
			select toDateTime(ctime) ctime, user_id userid from game.col_nsq_log_external_bet final where ctime > ? and ctime < ? and amount != 0
		) s0
		group by day_time, minute_time
	`, stime.Unix(), etime.Unix(), stime.Unix(), etime.Unix())
	if err != nil {
		return
	}

	var datasMap = make(map[uint32]map[uint32][]string)
	for _, data := range datas {
		day_time := data["day_time"].(uint32)
		minute_time := data["minute_time"].(uint32)
		userids := data["userids"].([]string)

		stat, ok := datasMap[day_time]
		if !ok {
			stat = make(map[uint32][]string)
			datasMap[day_time] = stat
		}
		stat[minute_time] = userids
	}

	now := NowTime()
	cur_day_time := uint32(now.Year()*10000 + int(now.Month())*100 + now.Day())
	cur_minute_time := uint32(now.Hour()*100 + now.Minute())

	for ; stime.Before(etime); stime = stime.AddDate(0, 0, 1) {
		day_time := uint32(stime.Year()*10000 + int(stime.Month())*100 + stime.Day())
		stat := &entity.LiveDataTrendStats{Trend: 14, DayTime: day_time}
		sday := fmt.Sprint(day_time)
		sdate := sday[0:4] + "-" + sday[4:6] + "-" + sday[6:8]
		stat.Name = sdate + "-" + "累计打码人数"
		stats = append(stats, stat)

		dataMap, ok := datasMap[day_time]
		if !ok {
			dataMap = make(map[uint32][]string)
		}

		var useridMap = make(map[string]bool)
		var hour, minute uint32
		for ; hour < 24; hour++ {
			for minute = range 60 {
				minute_time := hour*100 + minute

				// 时间未到
				if day_time > cur_day_time || (day_time == cur_day_time && minute_time > cur_minute_time) {
					continue
				}

				userids := dataMap[minute_time]
				for _, userid := range userids {
					useridMap[userid] = true
				}

				stat.Datas = append(stat.Datas, float64(len(useridMap)))
			}
		}
	}
	return
}

// 15.累计打码量
func (s *statisticsService) liveDataTrendStat_15(stime, etime time.Time) (stats []*entity.LiveDataTrendStats, err error) {
	var datas []map[string]any
	err = ck.Select(&datas, `
		select day_time, minute_time, sum(bets)/100 bets from (
			select toYYYYMMDD(toDateTime(begin_time)) day_time, (toHour(toDateTime(begin_time))*100 + toMinute(toDateTime(begin_time))) minute_time, sum(bet_amount) bets 
			from game.col_detail final where robot = 0 and begin_time >= ? and begin_time < ?
			group by day_time, minute_time
			
			union all
			
			select toYYYYMMDD(toDateTime(ctime)) day_time, (toHour(toDateTime(ctime))*100 + toMinute(toDateTime(ctime))) minute_time, sum(amount) bets 
			from game.col_nsq_log_external_bet final where ctime > ? and ctime < ? and amount != 0
			group by day_time, minute_time
		) s0
		group by day_time, minute_time
	`, stime.Unix(), etime.Unix(), stime.Unix(), etime.Unix())
	if err != nil {
		return
	}
	stats = liveDataTrendDatas2Stats(15, stime, etime, datas, "bets")
	return
}

// 16.累计充投比 累计充投比（下注 / 充）10.12
func (s *statisticsService) liveDataTrendStat_16(stime, etime time.Time) (stats []*entity.LiveDataTrendStats, err error) {
	// 12.累计充值金额
	payStats, err := s.liveDataTrendStat_12(stime, etime)
	if err != nil {
		return
	}
	dayPayStats := make(map[uint32]*entity.LiveDataTrendStats)
	for _, pay := range payStats {
		dayPayStats[pay.DayTime] = pay
	}

	// 15.累计打码量
	betStats, err := s.liveDataTrendStat_15(stime, etime)
	if err != nil {
		return
	}

	for _, betStat := range betStats {
		sday := fmt.Sprint(betStat.DayTime)
		sdate := sday[0:4] + "-" + sday[4:6] + "-" + sday[6:8]
		stat := &entity.LiveDataTrendStats{
			Trend:   16,
			Name:    sdate + "-累计充投比",
			DayTime: betStat.DayTime,
		}
		stats = append(stats, stat)
		if payStat, ok := dayPayStats[betStat.DayTime]; ok {
			for i, bet := range betStat.Datas {
				var rate float64
				if i < len(payStat.Datas) && payStat.Datas[i] != 0 {
					rate = bet / payStat.Datas[i]
				}
				stat.Datas = append(stat.Datas, rate)
			}
		}
	}
	return
}

// 17.新用户累计充投比 累计充投比（下注 / 充）10.12
func (s *statisticsService) liveDataTrendStat_17(stime, etime time.Time) (stats []*entity.LiveDataTrendStats, err error) {
	return s.liveDataTrendStatBetPayRate(stime, etime, 1, "新用户累计充投比")
}

// 18.老用户累计充投比
func (s *statisticsService) liveDataTrendStat_18(stime, etime time.Time) (stats []*entity.LiveDataTrendStats, err error) {
	return s.liveDataTrendStatBetPayRate(stime, etime, 0, "老用户累计充投比")
}

// 新老用户累计充投比
// utype 1新 0老
func (s *statisticsService) liveDataTrendStatBetPayRate(stime, etime time.Time, utype int32, name string) (stats []*entity.LiveDataTrendStats, err error) {
	utypeCompare := utils.CaseElse(utype == 1, "=", ">")
	// 累计充值
	var pay_datas []map[string]any
	err = ck.Select(&pay_datas, fmt.Sprintf(`
		select s1.day_time, s1.minute_time, SUM(amounts)/100 amounts
		from game.col_user s0 final
		join (
			select userid, toYYYYMMDD(ctime) day_time, (toHour(ctime)*100 + toMinute(ctime)) minute_time, SUM(amount) amounts
			from game.col_trade_record final
			where ctime between ? and ? and order_status = 4
			group by userid, day_time, minute_time
		) s1 on s0.userid = s1.userid
		where s1.day_time %s toYYYYMMDD(s0.ctime)
		group by s1.day_time, s1.minute_time
	`, utypeCompare), stime, etime)
	if err != nil {
		return
	}
	payStats := liveDataTrendDatas2Stats(12, stime, etime, pay_datas, "amounts")
	dayPayStats := make(map[uint32]*entity.LiveDataTrendStats)
	for _, pay := range payStats {
		dayPayStats[pay.DayTime] = pay
	}

	// 15.累计打码量
	var bet_datas []map[string]any
	err = ck.Select(&bet_datas, fmt.Sprintf(`
		select day_time, minute_time, sum(bets)/100 bets from (
			select s1.day_time, s1.minute_time, SUM(bets) bets
			from game.col_user s0 final
			join (
				select userid, toYYYYMMDD(toDateTime(begin_time)) day_time, (toHour(toDateTime(begin_time))*100 + toMinute(toDateTime(begin_time))) minute_time, sum(bet_amount) bets 
				from game.col_detail final where robot = 0 and begin_time >= ? and begin_time < ?
				group by userid, day_time, minute_time
			) s1 on s0.userid = s1.userid
			where s1.day_time %s toYYYYMMDD(s0.ctime)
			group by s1.day_time, s1.minute_time
			
			union all
			
			select s1.day_time, s1.minute_time, SUM(bets) bets
			from game.col_user s0 final
			join (
				select user_id, toYYYYMMDD(toDateTime(ctime)) day_time, (toHour(toDateTime(ctime))*100 + toMinute(toDateTime(ctime))) minute_time, sum(amount) bets 
				from game.col_nsq_log_external_bet final where ctime > ? and ctime < ? and amount != 0
				group by user_id, day_time, minute_time
			) s1 on s0.userid = s1.user_id
			where s1.day_time %s toYYYYMMDD(s0.ctime)
			group by s1.day_time, s1.minute_time
		) t0
		group by day_time, minute_time
	`, utypeCompare, utypeCompare), stime.Unix(), etime.Unix(), stime.Unix(), etime.Unix())
	if err != nil {
		return
	}
	betStats := liveDataTrendDatas2Stats(15, stime, etime, bet_datas, "bets")

	for _, betStat := range betStats {
		sday := fmt.Sprint(betStat.DayTime)
		sdate := sday[0:4] + "-" + sday[4:6] + "-" + sday[6:8]
		stat := &entity.LiveDataTrendStats{
			Trend:   16,
			Name:    sdate + "-" + name,
			DayTime: betStat.DayTime,
		}
		stats = append(stats, stat)
		if payStat, ok := dayPayStats[betStat.DayTime]; ok {
			for i, bet := range betStat.Datas {
				var rate float64
				if i < len(payStat.Datas) && payStat.Datas[i] != 0 {
					rate = bet / payStat.Datas[i]
				}
				stat.Datas = append(stat.Datas, rate)
			}
		}
	}
	return
}
