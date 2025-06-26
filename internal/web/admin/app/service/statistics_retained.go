package service

import (
	"fmt"
	"goserver/internal/web/admin/app/entity"
	"goserver/pkg/data/ck"
	"goserver/pkg/utils"
	"sort"
	"strings"
	"time"
)

// 用户留存 实时查询
func (svc *statisticsService) GetRetained(begin, end time.Time, packageIds []string) (ret []*entity.UserRetained, err error) {
	// 所有渠道数据
	channelsMap, err := ChannelService.GetChannelsMap()
	if err != nil {
		return
	}

	var where0 string
	var args0 []any
	if len(packageIds) > 0 {
		where0 += " AND ad__bundle_id IN ?"
		args0 = append(args0, packageIds)
	}

	var dateRegistMap = make(map[uint32]int64)
	var dateChannels1Map = make(map[uint32]string)
	var dateChannelClasses1Map = make(map[uint32]string)
	var datas_regist []map[string]any
	err = ck.Select(&datas_regist, fmt.Sprintf(`
		SELECT toYYYYMMDD(ctime) regist_date, count(*) count, groupArray(DISTINCT ad__bundle_id) channels
		FROM game.col_user FINAL
		WHERE ctime BETWEEN ? AND ? %s
		GROUP BY regist_date
	`, where0), append([]any{begin, end}, args0...)...)
	if err != nil {
		return
	}
	for _, data := range datas_regist {
		regist_date := data["regist_date"].(uint32)
		count := utils.ToInt64(data["count"])
		dateRegistMap[regist_date] = count

		channels := data["channels"].([]string)
		var channelClassMap = make(map[string]bool)
		var channels1, channelsClass []string
		for _, c := range channels {
			if channel, ok := channelsMap[c]; ok {
				if channel.Name1 != "" {
					channels1 = append(channels1, channel.Name1)
				}
				if channel.ClassName != "" {
					if ok := channelClassMap[channel.ClassName]; !ok {
						channelClassMap[channel.ClassName] = true
						channelsClass = append(channelsClass, channel.ClassName)
					}
				}
			}
		}
		sort.Strings(channels1)
		sort.Strings(channelsClass)
		dateChannels1Map[regist_date] = strings.Join(channels1, ",")
		dateChannelClasses1Map[regist_date] = strings.Join(channelsClass, ",")
	}

	sql0 := fmt.Sprintf(`
		SELECT dateDiff('day', toDate(?), login_day) diff_day, count(*) c
		FROM (
			SELECT userid, toDate(login_time) login_day
			FROM game.col_log_login FINAL WHERE userid IN (
				SELECT userid FROM game.col_user FINAL 
				WHERE ctime >= ? AND ctime < ? %s
			)
			GROUP BY userid, login_day
		) t2
		GROUP BY diff_day
		HAVING diff_day IN (1,2,3,4,5,6,7,15,30,40,50,60,90)
	`, where0)

	for begin.Before(end) {
		stime := begin
		etime := begin.AddDate(0, 0, 1)
		begin = begin.AddDate(0, 0, 1)
		ymd := uint32(stime.Year()*10000) + uint32(stime.Month()*100) + uint32(stime.Day())
		item := &entity.UserRetained{
			SDate:        stime.Format(utils.FORMAT_DATE),
			Channel1:     dateChannels1Map[ymd],
			ChannelClass: dateChannelClasses1Map[ymd],
			RegistCount:  dateRegistMap[ymd],
		}
		ret = append(ret, item)

		var datas []map[string]any
		args := append([]any{stime, stime, etime}, args0...)
		err = ck.Select(&datas, sql0, args...)
		if err != nil {
			return
		}

		now := NowTime()
		nowYMD := uint32(now.Year()*10000) + uint32(now.Month()*100) + uint32(now.Day())
		for _, data := range datas {
			diff_day := utils.ToInt64(data["diff_day"])
			count := utils.ToInt64(data["c"])
			rate := fmt.Sprintf("%.2f%%", ComputeFloat(count, item.RegistCount)*100.0)
			switch diff_day {
			case 1:
				item.Day1Rate = rate
			case 2:
				item.Day2Rate = rate
			case 3:
				item.Day3Rate = rate
			case 4:
				item.Day4Rate = rate
			case 5:
				item.Day5Rate = rate
			case 6:
				item.Day6Rate = rate
			case 7:
				item.Day7Rate = rate
			case 15:
				item.Day15Rate = rate
			case 30:
				item.Day30Rate = rate
			case 40:
				item.Day40Rate = rate
			case 50:
				item.Day50Rate = rate
			case 60:
				item.Day60Rate = rate
			case 90:
				item.Day90Rate = rate
			}
		}

		var days = map[uint32]*string{
			1:  &item.Day1Rate,
			2:  &item.Day2Rate,
			3:  &item.Day3Rate,
			4:  &item.Day4Rate,
			5:  &item.Day5Rate,
			6:  &item.Day6Rate,
			7:  &item.Day7Rate,
			15: &item.Day15Rate,
			30: &item.Day30Rate,
			40: &item.Day40Rate,
			50: &item.Day50Rate,
			60: &item.Day60Rate,
			90: &item.Day90Rate,
		}
		for day, field := range days {
			if *field == "" && ymd+day <= nowYMD {
				*field = "0.00%"
			}
		}
	}
	// 表格：倒序显示
	count := len(ret)
	for i := 0; i < count/2; i++ {
		ret[i], ret[count-1-i] = ret[count-1-i], ret[i]
	}
	return
}

// 用户留存 付费留存
func (svc *statisticsService) GetRetainedPay(begin, end time.Time, packageIds []string) (ret []*entity.UserRetained, err error) {
	// 所有渠道数据
	channelsMap, err := ChannelService.GetChannelsMap()
	if err != nil {
		return
	}

	var where0 string
	var args0 []any
	if len(packageIds) > 0 {
		where0 += " AND ad__bundle_id IN ?"
		args0 = append(args0, packageIds)
	}

	// 注册人数，渠道统计
	var dateRegistMap = make(map[uint32]int64)
	var dateChannels1Map = make(map[uint32]string)
	var dateChannelClasses1Map = make(map[uint32]string)
	var datas_regist []map[string]any
	err = ck.Select(&datas_regist, fmt.Sprintf(`
		SELECT toYYYYMMDD(ctime) regist_date, count(*) count, groupArray(DISTINCT ad__bundle_id) channels
		FROM game.col_user FINAL
		WHERE ctime BETWEEN ? AND ? %s
		GROUP BY regist_date
	`, where0), append([]any{begin, end}, args0...)...)
	if err != nil {
		return
	}
	for _, data := range datas_regist {
		regist_date := data["regist_date"].(uint32)
		count := utils.ToInt64(data["count"])
		dateRegistMap[regist_date] = count

		channels := data["channels"].([]string)
		var channelClassMap = make(map[string]bool)
		var channels1, channelsClass []string
		for _, c := range channels {
			if channel, ok := channelsMap[c]; ok {
				if channel.Name1 != "" {
					channels1 = append(channels1, channel.Name1)
				}
				if channel.ClassName != "" {
					if ok := channelClassMap[channel.ClassName]; !ok {
						channelClassMap[channel.ClassName] = true
						channelsClass = append(channelsClass, channel.ClassName)
					}
				}
			}
		}
		sort.Strings(channels1)
		sort.Strings(channelsClass)
		dateChannels1Map[regist_date] = strings.Join(channels1, ",")
		dateChannelClasses1Map[regist_date] = strings.Join(channelsClass, ",")
	}

	// 支付人数统计
	var datePayCountMap = make(map[uint32]int64)
	var datas_pay []map[string]any
	err = ck.Select(&datas_pay, fmt.Sprintf(`
		SELECT regist_date, count(DISTINCT t2.userid) pay_count
		FROM game.col_trade_record t1 FINAL
		JOIN (
			SELECT userid, toYYYYMMDD(ctime) regist_date FROM game.col_user FINAL
			WHERE ctime BETWEEN ? AND ? %s
		) t2 ON t1.userid = t2.userid
		WHERE order_status = 4
		GROUP BY t2.regist_date
	`, where0), append([]any{begin, end}, args0...)...)
	if err != nil {
		return
	}
	for _, data := range datas_pay {
		regist_date := data["regist_date"].(uint32)
		pay_count := utils.ToInt64(data["pay_count"])
		datePayCountMap[regist_date] = pay_count
	}

	sql0 := fmt.Sprintf(`
		SELECT dateDiff('day', toDate(?), login_day) diff_day, count(*) c
		FROM (
			SELECT userid, toDate(login_time) login_day
			FROM game.col_log_login FINAL WHERE userid IN (
				SELECT userid FROM game.col_user FINAL 
				WHERE money > 0 AND ctime >= ? AND ctime < ? %s
			)
			GROUP BY userid, login_day
		) t2
		GROUP BY diff_day
		HAVING diff_day IN (1,2,3,4,5,6,7,15,30,40,50,60,90)
		ORDER BY diff_day
	`, where0)

	now := NowTime()
	nowYMD := uint32(now.Year()*10000) + uint32(now.Month()*100) + uint32(now.Day())
	for begin.Before(end) {
		stime := begin
		etime := begin.AddDate(0, 0, 1)
		begin = begin.AddDate(0, 0, 1)
		ymd := uint32(stime.Year()*10000) + uint32(stime.Month()*100) + uint32(stime.Day())
		item := &entity.UserRetained{
			SDate:        stime.Format(utils.FORMAT_DATE),
			Channel1:     dateChannels1Map[ymd],
			ChannelClass: dateChannelClasses1Map[ymd],
			RegistCount:  dateRegistMap[ymd],
			PayCount:     datePayCountMap[ymd],
		}
		item.PayRate = fmt.Sprintf("%.2f%%", ComputeFloat(item.PayCount, item.RegistCount)*100.0)
		ret = append(ret, item)

		var datas []map[string]any
		args := append([]any{stime, stime, etime}, args0...)
		err = ck.Select(&datas, sql0, args...)
		if err != nil {
			return
		}
		for _, data := range datas {
			diff_day := utils.ToInt64(data["diff_day"])
			count := utils.ToInt64(data["c"])
			rate := fmt.Sprintf("%.2f%%", ComputeFloat(count, item.PayCount)*100.0)
			switch diff_day {
			case 1:
				item.Day1Rate = rate
			case 2:
				item.Day2Rate = rate
			case 3:
				item.Day3Rate = rate
			case 4:
				item.Day4Rate = rate
			case 5:
				item.Day5Rate = rate
			case 6:
				item.Day6Rate = rate
			case 7:
				item.Day7Rate = rate
			case 15:
				item.Day15Rate = rate
			case 30:
				item.Day30Rate = rate
			case 40:
				item.Day40Rate = rate
			case 50:
				item.Day50Rate = rate
			case 60:
				item.Day60Rate = rate
			case 90:
				item.Day90Rate = rate
			}
		}

		var days = map[uint32]*string{
			1:  &item.Day1Rate,
			2:  &item.Day2Rate,
			3:  &item.Day3Rate,
			4:  &item.Day4Rate,
			5:  &item.Day5Rate,
			6:  &item.Day6Rate,
			7:  &item.Day7Rate,
			15: &item.Day15Rate,
			30: &item.Day30Rate,
			40: &item.Day40Rate,
			50: &item.Day50Rate,
			60: &item.Day60Rate,
			90: &item.Day90Rate,
		}
		for day, field := range days {
			if *field == "" && ymd+day <= nowYMD {
				*field = "0.00%"
			}
		}
	}
	// 表格：倒序显示
	count := len(ret)
	for i := 0; i < count/2; i++ {
		ret[i], ret[count-1-i] = ret[count-1-i], ret[i]
	}
	return
}

// 用户留存 复充留存
func (svc *statisticsService) GetRetainedRepay(begin, end time.Time, packageIds []string) (ret []*entity.UserRetained, err error) {
	// 所有渠道数据
	channelsMap, err := ChannelService.GetChannelsMap()
	if err != nil {
		return
	}

	var where0 string
	var args0 []any
	if len(packageIds) > 0 {
		where0 += " AND ad__bundle_id IN ?"
		args0 = append(args0, packageIds)
	}

	// 注册人数，渠道统计
	var dateRegistMap = make(map[uint32]int64)
	var dateChannels1Map = make(map[uint32]string)
	var dateChannelClasses1Map = make(map[uint32]string)
	var datas_regist []map[string]any
	err = ck.Select(&datas_regist, fmt.Sprintf(`
		SELECT toYYYYMMDD(ctime) regist_date, count(*) count, groupArray(DISTINCT ad__bundle_id) channels
		FROM game.col_user FINAL
		WHERE ctime BETWEEN ? AND ? %s
		GROUP BY regist_date
	`, where0), append([]any{begin, end}, args0...)...)
	if err != nil {
		return
	}
	for _, data := range datas_regist {
		regist_date := data["regist_date"].(uint32)
		count := utils.ToInt64(data["count"])
		dateRegistMap[regist_date] = count

		channels := data["channels"].([]string)
		var channelClassMap = make(map[string]bool)
		var channels1, channelsClass []string
		for _, c := range channels {
			if channel, ok := channelsMap[c]; ok {
				if channel.Name1 != "" {
					channels1 = append(channels1, channel.Name1)
				}
				if channel.ClassName != "" {
					if ok := channelClassMap[channel.ClassName]; !ok {
						channelClassMap[channel.ClassName] = true
						channelsClass = append(channelsClass, channel.ClassName)
					}
				}
			}
		}
		sort.Strings(channels1)
		sort.Strings(channelsClass)
		dateChannels1Map[regist_date] = strings.Join(channels1, ",")
		dateChannelClasses1Map[regist_date] = strings.Join(channelsClass, ",")
	}

	// 支付人数统计
	var datePayCountMap = make(map[uint32]int64)
	var datas_pay []map[string]any
	err = ck.Select(&datas_pay, fmt.Sprintf(`
		SELECT regist_date, count(DISTINCT t2.userid) pay_count
		FROM game.col_trade_record t1 FINAL
		JOIN (
			SELECT userid, toYYYYMMDD(ctime) regist_date FROM game.col_user FINAL
			WHERE ctime BETWEEN ? AND ? %s
		) t2 ON t1.userid = t2.userid
		WHERE order_status = 4
		GROUP BY t2.regist_date
	`, where0), append([]any{begin, end}, args0...)...)
	if err != nil {
		return
	}
	for _, data := range datas_pay {
		regist_date := data["regist_date"].(uint32)
		pay_count := utils.ToInt64(data["pay_count"])
		datePayCountMap[regist_date] = pay_count
	}

	// 复充人数统计
	var dateRepayCountMap = make(map[uint32]int64)
	var datas_repay []map[string]any
	err = ck.Select(&datas_repay, fmt.Sprintf(`
		SELECT regist_date, count(*) repay_count 
		FROM (
			SELECT t2.userid, t2.regist_date, count(*) pay_times
			FROM game.col_trade_record t1 FINAL
			JOIN (
				SELECT userid, toYYYYMMDD(ctime) regist_date FROM game.col_user FINAL
				WHERE ctime BETWEEN ? AND ? %s
			) t2 ON t1.userid = t2.userid
			WHERE order_status = 4
			GROUP BY t2.userid, t2.regist_date
			HAVING pay_times >= 2
		) s1
		GROUP BY regist_date
	`, where0), append([]any{begin, end}, args0...)...)
	if err != nil {
		return
	}
	for _, data := range datas_repay {
		regist_date := data["regist_date"].(uint32)
		repay_count := utils.ToInt64(data["repay_count"])
		dateRepayCountMap[regist_date] = repay_count
	}

	sql0 := fmt.Sprintf(`
		SELECT dateDiff('day', toDate(?), login_day) diff_day, count(*) c 
		FROM (
			SELECT userid, toDate(login_time) login_day
			FROM game.col_log_login FINAL WHERE userid IN (
				SELECT t2.userid
				FROM game.col_trade_record t1 FINAL
				JOIN (
					SELECT userid FROM game.col_user FINAL
					WHERE ctime >= ? AND ctime < ? %s
				) t2 ON t1.userid = t2.userid
				WHERE order_status = 4
				GROUP BY t2.userid
				HAVING count(*) >= 2
			)
			GROUP BY userid, login_day
		) t2
		GROUP BY diff_day
		HAVING diff_day IN (1,2,3,4,5,6,7,15,30,40,50,60,90)
		ORDER BY diff_day
	`, where0)

	now := NowTime()
	nowYMD := uint32(now.Year()*10000) + uint32(now.Month()*100) + uint32(now.Day())
	for begin.Before(end) {
		stime := begin
		etime := begin.AddDate(0, 0, 1)
		begin = begin.AddDate(0, 0, 1)
		ymd := uint32(stime.Year()*10000) + uint32(stime.Month()*100) + uint32(stime.Day())
		item := &entity.UserRetained{
			SDate:        stime.Format(utils.FORMAT_DATE),
			Channel1:     dateChannels1Map[ymd],
			ChannelClass: dateChannelClasses1Map[ymd],
			RegistCount:  dateRegistMap[ymd],
			PayCount:     datePayCountMap[ymd],
			RepayCount:   dateRepayCountMap[ymd],
		}
		item.PayRate = fmt.Sprintf("%.2f%%", ComputeFloat(item.PayCount, item.RegistCount)*100.0)
		item.RepayRate = fmt.Sprintf("%.2f%%", ComputeFloat(item.RepayCount, item.PayCount)*100)
		ret = append(ret, item)

		var datas []map[string]any
		args := append([]any{stime, stime, etime}, args0...)
		err = ck.Select(&datas, sql0, args...)
		if err != nil {
			return
		}
		for _, data := range datas {
			diff_day := utils.ToInt64(data["diff_day"])
			count := utils.ToInt64(data["c"])
			rate := fmt.Sprintf("%.2f%%", ComputeFloat(count, item.RepayCount)*100.0)
			switch diff_day {
			case 1:
				item.Day1Rate = rate
			case 2:
				item.Day2Rate = rate
			case 3:
				item.Day3Rate = rate
			case 4:
				item.Day4Rate = rate
			case 5:
				item.Day5Rate = rate
			case 6:
				item.Day6Rate = rate
			case 7:
				item.Day7Rate = rate
			case 15:
				item.Day15Rate = rate
			case 30:
				item.Day30Rate = rate
			case 40:
				item.Day40Rate = rate
			case 50:
				item.Day50Rate = rate
			case 60:
				item.Day60Rate = rate
			case 90:
				item.Day90Rate = rate
			}
		}

		var days = map[uint32]*string{
			1:  &item.Day1Rate,
			2:  &item.Day2Rate,
			3:  &item.Day3Rate,
			4:  &item.Day4Rate,
			5:  &item.Day5Rate,
			6:  &item.Day6Rate,
			7:  &item.Day7Rate,
			15: &item.Day15Rate,
			30: &item.Day30Rate,
			40: &item.Day40Rate,
			50: &item.Day50Rate,
			60: &item.Day60Rate,
			90: &item.Day90Rate,
		}
		for day, field := range days {
			if *field == "" && ymd+day <= nowYMD {
				*field = "0.00%"
			}
		}
	}
	// 表格：倒序显示
	count := len(ret)
	for i := 0; i < count/2; i++ {
		ret[i], ret[count-1-i] = ret[count-1-i], ret[i]
	}
	return
}

// 用户留存 付费后留存
func (svc *statisticsService) GetRetainedAfterPay(begin, end time.Time, packageIds []string) (ret []*entity.UserRetained, err error) {
	// 所有渠道数据
	channelsMap, err := ChannelService.GetChannelsMap()
	if err != nil {
		return
	}

	var where0 string
	var args0 []any
	if len(packageIds) > 0 {
		where0 += " AND ad__bundle_id IN ?"
		args0 = append(args0, packageIds)
	}

	// 注册人数，渠道统计
	var dateRegistMap = make(map[uint32]int64)
	var dateChannels1Map = make(map[uint32]string)
	var dateChannelClasses1Map = make(map[uint32]string)
	var datas_regist []map[string]any
	err = ck.Select(&datas_regist, fmt.Sprintf(`
		SELECT toYYYYMMDD(ctime) regist_date, count(*) count, groupArray(DISTINCT ad__bundle_id) channels
		FROM game.col_user FINAL
		WHERE ctime BETWEEN ? AND ? %s
		GROUP BY regist_date
	`, where0), append([]any{begin, end}, args0...)...)
	if err != nil {
		return
	}
	for _, data := range datas_regist {
		regist_date := data["regist_date"].(uint32)
		count := utils.ToInt64(data["count"])
		dateRegistMap[regist_date] = count

		channels := data["channels"].([]string)
		var channelClassMap = make(map[string]bool)
		var channels1, channelsClass []string
		for _, c := range channels {
			if channel, ok := channelsMap[c]; ok {
				if channel.Name1 != "" {
					channels1 = append(channels1, channel.Name1)
				}
				if channel.ClassName != "" {
					if ok := channelClassMap[channel.ClassName]; !ok {
						channelClassMap[channel.ClassName] = true
						channelsClass = append(channelsClass, channel.ClassName)
					}
				}
			}
		}
		sort.Strings(channels1)
		sort.Strings(channelsClass)
		dateChannels1Map[regist_date] = strings.Join(channels1, ",")
		dateChannelClasses1Map[regist_date] = strings.Join(channelsClass, ",")
	}

	// 当日成为付费的人数统计
	var dateBePayCountMap = make(map[uint32]int64)
	var datas_repay []map[string]any
	err = ck.Select(&datas_repay, fmt.Sprintf(`
		SELECT toYYYYMMDD(first_pay_time) pay_time, count(*) count
		FROM game.col_user t1 FINAL
		JOIN (
			SELECT userid, MIN(ctime) first_pay_time
			FROM game.col_trade_record t1 FINAL
			WHERE order_status = 4
			GROUP BY userid
			HAVING first_pay_time BETWEEN ? AND ?
		) t2 ON t1.userid = t2.userid
		WHERE 1 = 1 %s
		GROUP BY pay_time
	`, where0), append([]any{begin, end}, args0...)...)
	if err != nil {
		return
	}
	for _, data := range datas_repay {
		pay_date := data["pay_time"].(uint32)
		repay_count := utils.ToInt64(data["count"])
		dateBePayCountMap[pay_date] = repay_count
	}

	sql0 := fmt.Sprintf(`
		SELECT dateDiff('day', toDate(?), login_day) diff_day, count(*) c 
		FROM (
			SELECT userid, toDate(login_time) login_day
			FROM game.col_log_login FINAL WHERE userid IN (
				SELECT t1.userid 
				FROM game.col_user t1 FINAL
				JOIN (
					SELECT userid
					FROM game.col_trade_record t1 FINAL
					WHERE order_status = 4 
					GROUP BY userid
					HAVING MIN(ctime) >= ? AND MIN(ctime) < ?
				) t2 ON t1.userid = t2.userid
				WHERE 1 = 1 %s
			)
			GROUP BY userid, login_day
		) t2
		GROUP BY diff_day
		HAVING diff_day IN (1,2,3,4,5,6,7,15,30,40,50,60,90)
		ORDER BY diff_day
	`, where0)

	now := NowTime()
	nowYMD := uint32(now.Year()*10000) + uint32(now.Month()*100) + uint32(now.Day())
	for begin.Before(end) {
		stime := begin
		etime := begin.AddDate(0, 0, 1)
		begin = begin.AddDate(0, 0, 1)
		ymd := uint32(stime.Year()*10000) + uint32(stime.Month()*100) + uint32(stime.Day())
		item := &entity.UserRetained{
			SDate:        stime.Format(utils.FORMAT_DATE),
			Channel1:     dateChannels1Map[ymd],
			ChannelClass: dateChannelClasses1Map[ymd],
			RegistCount:  dateRegistMap[ymd],
			// PayCount:     datePayCountMap[ymd],
			BePayCount: dateBePayCountMap[ymd],
		}
		// item.PayRate = fmt.Sprintf("%.2f%%", ComputeFloat(item.PayCount, item.RegistCount)*100.0)
		// item.RepayRate = fmt.Sprintf("%.2f%%", ComputeFloat(item.RepayCount, item.PayCount)*100)
		ret = append(ret, item)

		var datas []map[string]any
		args := append([]any{stime, stime, etime}, args0...)
		err = ck.Select(&datas, sql0, args...)
		if err != nil {
			return
		}
		for _, data := range datas {
			diff_day := utils.ToInt64(data["diff_day"])
			count := utils.ToInt64(data["c"])
			rate := fmt.Sprintf("%.2f%%", ComputeFloat(count, item.BePayCount)*100.0)
			switch diff_day {
			case 1:
				item.Day1Rate = rate
			case 2:
				item.Day2Rate = rate
			case 3:
				item.Day3Rate = rate
			case 4:
				item.Day4Rate = rate
			case 5:
				item.Day5Rate = rate
			case 6:
				item.Day6Rate = rate
			case 7:
				item.Day7Rate = rate
			case 15:
				item.Day15Rate = rate
			case 30:
				item.Day30Rate = rate
			case 40:
				item.Day40Rate = rate
			case 50:
				item.Day50Rate = rate
			case 60:
				item.Day60Rate = rate
			case 90:
				item.Day90Rate = rate
			}
		}

		var days = map[uint32]*string{
			1:  &item.Day1Rate,
			2:  &item.Day2Rate,
			3:  &item.Day3Rate,
			4:  &item.Day4Rate,
			5:  &item.Day5Rate,
			6:  &item.Day6Rate,
			7:  &item.Day7Rate,
			15: &item.Day15Rate,
			30: &item.Day30Rate,
			40: &item.Day40Rate,
			50: &item.Day50Rate,
			60: &item.Day60Rate,
			90: &item.Day90Rate,
		}
		for day, field := range days {
			if *field == "" && ymd+day <= nowYMD {
				*field = "0.00%"
			}
		}
	}
	// 表格：倒序显示
	count := len(ret)
	for i := 0; i < count/2; i++ {
		ret[i], ret[count-1-i] = ret[count-1-i], ret[i]
	}
	return
}

// 用户留存 复充后留存
func (svc *statisticsService) GetRetainedAfterRepay(begin, end time.Time, packageIds []string) (ret []*entity.UserRetained, err error) {
	// 所有渠道数据
	channelsMap, err := ChannelService.GetChannelsMap()
	if err != nil {
		return
	}

	var where0 string
	var args0 []any
	if len(packageIds) > 0 {
		where0 += " AND ad__bundle_id IN ?"
		args0 = append(args0, packageIds)
	}

	// 注册人数，渠道统计
	var dateRegistMap = make(map[uint32]int64)
	var dateChannels1Map = make(map[uint32]string)
	var dateChannelClasses1Map = make(map[uint32]string)
	var datas_regist []map[string]any
	err = ck.Select(&datas_regist, fmt.Sprintf(`
		SELECT toYYYYMMDD(ctime) regist_date, count(*) count, groupArray(DISTINCT ad__bundle_id) channels
		FROM game.col_user FINAL
		WHERE ctime BETWEEN ? AND ? %s
		GROUP BY regist_date
	`, where0), append([]any{begin, end}, args0...)...)
	if err != nil {
		return
	}
	for _, data := range datas_regist {
		regist_date := data["regist_date"].(uint32)
		count := utils.ToInt64(data["count"])
		dateRegistMap[regist_date] = count

		channels := data["channels"].([]string)
		var channelClassMap = make(map[string]bool)
		var channels1, channelsClass []string
		for _, c := range channels {
			if channel, ok := channelsMap[c]; ok {
				if channel.Name1 != "" {
					channels1 = append(channels1, channel.Name1)
				}
				if channel.ClassName != "" {
					if ok := channelClassMap[channel.ClassName]; !ok {
						channelClassMap[channel.ClassName] = true
						channelsClass = append(channelsClass, channel.ClassName)
					}
				}
			}
		}
		sort.Strings(channels1)
		sort.Strings(channelsClass)
		dateChannels1Map[regist_date] = strings.Join(channels1, ",")
		dateChannelClasses1Map[regist_date] = strings.Join(channelsClass, ",")
	}

	// 当日成为复充的人数统计
	var dateBeRepayCountMap = make(map[uint32]int64)
	var datas_repay []map[string]any
	err = ck.Select(&datas_repay, fmt.Sprintf(`
		SELECT toYYYYMMDD(repay_time) repay_date, count(*) count
		FROM game.col_user t1 FINAL
		JOIN (
			SELECT userid, repay_time 
			FROM (
				SELECT userid, ctime repay_time
				FROM game.col_trade_record FINAL 
				WHERE order_status = 4 
				ORDER BY userid, ctime 
				LIMIT 1,1 BY userid
			) s1
			WHERE repay_time BETWEEN ? AND ?
		) t2 ON t1.userid = t2.userid
		WHERE 1 = 1 %s
		GROUP BY repay_date
	`, where0), append([]any{begin, end}, args0...)...)
	if err != nil {
		return
	}
	for _, data := range datas_repay {
		repay_time := data["repay_date"].(uint32)
		repay_count := utils.ToInt64(data["count"])
		dateBeRepayCountMap[repay_time] = repay_count
	}

	sql0 := fmt.Sprintf(`
		SELECT dateDiff('day', toDate(?), login_day) diff_day, count(*) c 
		FROM (
			SELECT userid, toDate(login_time) login_day
			FROM game.col_log_login FINAL WHERE userid IN (
				SELECT t1.userid 
				FROM game.col_user t1 FINAL
				JOIN (
					SELECT userid
					FROM (
						SELECT userid, ctime repay_time
						FROM game.col_trade_record FINAL 
						WHERE order_status = 4 
						ORDER BY userid, ctime 
						LIMIT 1,1 BY userid
					) s1
					WHERE repay_time >= ? AND repay_time < ?
				) t2 ON t1.userid = t2.userid
				WHERE 1 = 1 %s
			)
			GROUP BY userid, login_day
		) t2
		GROUP BY diff_day
		HAVING diff_day IN (1,2,3,4,5,6,7,15,30,40,50,60,90)
		ORDER BY diff_day
	`, where0)

	now := NowTime()
	nowYMD := uint32(now.Year()*10000) + uint32(now.Month()*100) + uint32(now.Day())
	for begin.Before(end) {
		stime := begin
		etime := begin.AddDate(0, 0, 1)
		begin = begin.AddDate(0, 0, 1)
		ymd := uint32(stime.Year()*10000) + uint32(stime.Month()*100) + uint32(stime.Day())
		item := &entity.UserRetained{
			SDate:        stime.Format(utils.FORMAT_DATE),
			Channel1:     dateChannels1Map[ymd],
			ChannelClass: dateChannelClasses1Map[ymd],
			RegistCount:  dateRegistMap[ymd],
			// PayCount:     datePayCountMap[ymd],
			BeRepayCount: dateBeRepayCountMap[ymd],
		}
		// item.PayRate = fmt.Sprintf("%.2f%%", ComputeFloat(item.PayCount, item.RegistCount)*100.0)
		// item.RepayRate = fmt.Sprintf("%.2f%%", ComputeFloat(item.RepayCount, item.PayCount)*100)
		ret = append(ret, item)

		var datas []map[string]any
		args := append([]any{stime, stime, etime}, args0...)
		err = ck.Select(&datas, sql0, args...)
		if err != nil {
			return
		}
		for _, data := range datas {
			diff_day := utils.ToInt64(data["diff_day"])
			count := utils.ToInt64(data["c"])
			rate := fmt.Sprintf("%.2f%%", ComputeFloat(count, item.BeRepayCount)*100.0)
			switch diff_day {
			case 1:
				item.Day1Rate = rate
			case 2:
				item.Day2Rate = rate
			case 3:
				item.Day3Rate = rate
			case 4:
				item.Day4Rate = rate
			case 5:
				item.Day5Rate = rate
			case 6:
				item.Day6Rate = rate
			case 7:
				item.Day7Rate = rate
			case 15:
				item.Day15Rate = rate
			case 30:
				item.Day30Rate = rate
			case 40:
				item.Day40Rate = rate
			case 50:
				item.Day50Rate = rate
			case 60:
				item.Day60Rate = rate
			case 90:
				item.Day90Rate = rate
			}
		}

		var days = map[uint32]*string{
			1:  &item.Day1Rate,
			2:  &item.Day2Rate,
			3:  &item.Day3Rate,
			4:  &item.Day4Rate,
			5:  &item.Day5Rate,
			6:  &item.Day6Rate,
			7:  &item.Day7Rate,
			15: &item.Day15Rate,
			30: &item.Day30Rate,
			40: &item.Day40Rate,
			50: &item.Day50Rate,
			60: &item.Day60Rate,
			90: &item.Day90Rate,
		}
		for day, field := range days {
			if *field == "" && ymd+day <= nowYMD {
				*field = "0.00%"
			}
		}
	}
	// 表格：倒序显示
	count := len(ret)
	for i := 0; i < count/2; i++ {
		ret[i], ret[count-1-i] = ret[count-1-i], ret[i]
	}
	return
}

// 用户留存 复充后留存
func (svc *statisticsService) GetRetainedR(begin, end time.Time, packageIds []string, utypeIds []int) (ret []*entity.UserRetained, err error) {
	// 所有渠道数据
	channelsMap, err := ChannelService.GetChannelsMap()
	if err != nil {
		return
	}

	var where0, where1 string
	var args0, args1 []any
	if len(packageIds) > 0 {
		where0 += " AND ad__bundle_id IN ?"
		args0 = append(args0, packageIds)
		where1 += " AND ad__bundle_id IN ?"
		args1 = append(args0, packageIds)
	}
	if len(utypeIds) > 0 {
		where1 += " AND ctype IN ?"
		args1 = append(args0, utypeIds)
	}

	// 注册人数，渠道统计
	var dateRegistMap = make(map[uint32]int64)
	var dateChannels1Map = make(map[uint32]string)
	var dateChannelClasses1Map = make(map[uint32]string)
	var datas_regist []map[string]any
	err = ck.Select(&datas_regist, fmt.Sprintf(`
		SELECT toYYYYMMDD(ctime) regist_date, count(*) count, groupArray(DISTINCT ad__bundle_id) channels
		FROM game.col_user FINAL
		WHERE ctime BETWEEN ? AND ? %s
		GROUP BY regist_date
	`, where0), append([]any{begin, end}, args0...)...)
	if err != nil {
		return
	}
	for _, data := range datas_regist {
		regist_date := data["regist_date"].(uint32)
		count := utils.ToInt64(data["count"])
		dateRegistMap[regist_date] = count

		channels := data["channels"].([]string)
		var channelClassMap = make(map[string]bool)
		var channels1, channelsClass []string
		for _, c := range channels {
			if channel, ok := channelsMap[c]; ok {
				if channel.Name1 != "" {
					channels1 = append(channels1, channel.Name1)
				}
				if channel.ClassName != "" {
					if ok := channelClassMap[channel.ClassName]; !ok {
						channelClassMap[channel.ClassName] = true
						channelsClass = append(channelsClass, channel.ClassName)
					}
				}
			}
		}
		sort.Strings(channels1)
		sort.Strings(channelsClass)
		dateChannels1Map[regist_date] = strings.Join(channels1, ",")
		dateChannelClasses1Map[regist_date] = strings.Join(channelsClass, ",")
	}

	// 当日成为R范围的人数统计
	var dateBeRCountMap = make(map[uint32]int64)
	var datas_be_r []map[string]any
	err = ck.Select(&datas_be_r, fmt.Sprintf(`
		SELECT regist_date, count(*) count
		FROM (
			SELECT toYYYYMMDD(ctime) regist_date,
				(CASE WHEN amount_sum = 0 THEN (CASE t1.state WHEN 1 THEN 0 ELSE 1 END)
					WHEN amount_sum BETWEEN 1 and 100000-1 THEN 2 
					WHEN amount_sum BETWEEN 100000 and 1000000-1 THEN 3
					WHEN amount_sum between 1000000 and 10000000-1 THEN 4
					WHEN amount_sum between 10000000 and 20000000-1 THEN 5
					WHEN amount_sum >= 20000000 THEN 6 ELSE -1 END) ctype
			FROM game.col_user t1 FINAL
			LEFT JOIN (
				SELECT userid, SUM(amount) amount_sum 
				FROM game.col_trade_record FINAL 
				WHERE order_status = 4 AND userid IN (
					SELECT userid FROM game.col_user FINAL WHERE ctime BETWEEN ? AND ?
				)
				GROUP BY userid
			) t2 ON t1.userid = t2.userid
			WHERE t1.ctime BETWEEN ? AND ? %s
		) s1
		GROUP BY regist_date
	`, where1), append([]any{begin, end, begin, end}, args1...)...)
	if err != nil {
		return
	}
	for _, data := range datas_be_r {
		regist_date := data["regist_date"].(uint32)
		r_count := utils.ToInt64(data["count"])
		dateBeRCountMap[regist_date] = r_count
	}

	sql0 := fmt.Sprintf(`
		SELECT dateDiff('day', toDate(?), login_day) diff_day, count(*) c 
		FROM (
			SELECT userid, toDate(login_time) login_day
			FROM game.col_log_login FINAL WHERE userid IN (
				SELECT userid FROM (
					SELECT t1.userid,
						(CASE WHEN amount_sum = 0 THEN (CASE t1.state WHEN 1 THEN 0 ELSE 1 END)
							WHEN amount_sum BETWEEN 1 and 100000-1 THEN 2 
							WHEN amount_sum BETWEEN 100000 and 1000000-1 THEN 3
							WHEN amount_sum between 1000000 and 10000000-1 THEN 4
							WHEN amount_sum between 10000000 and 20000000-1 THEN 5
							WHEN amount_sum >= 20000000 THEN 6 ELSE -1 END) ctype
					FROM game.col_user t1 FINAL
					LEFT JOIN (
						SELECT userid, SUM(amount) amount_sum 
						FROM game.col_trade_record FINAL 
						WHERE order_status = 4 AND userid IN (
							SELECT userid FROM game.col_user FINAL WHERE ctime >= ? AND ctime < ?
						)
						GROUP BY userid
					) t2 ON t1.userid = t2.userid
					WHERE t1.ctime BETWEEN ? AND ? %s
				) s1
			)
			GROUP BY userid, login_day
		) t2
		GROUP BY diff_day
		HAVING diff_day IN (1,2,3,4,5,6,7,15,30,40,50,60,90)
		ORDER BY diff_day
	`, where1)

	now := NowTime()
	nowYMD := uint32(now.Year()*10000) + uint32(now.Month()*100) + uint32(now.Day())
	for begin.Before(end) {
		stime := begin
		etime := begin.AddDate(0, 0, 1)
		begin = begin.AddDate(0, 0, 1)
		ymd := uint32(stime.Year()*10000) + uint32(stime.Month()*100) + uint32(stime.Day())
		item := &entity.UserRetained{
			SDate:        stime.Format(utils.FORMAT_DATE),
			Channel1:     dateChannels1Map[ymd],
			ChannelClass: dateChannelClasses1Map[ymd],
			RegistCount:  dateRegistMap[ymd],
			// PayCount:     datePayCountMap[ymd],
			BeRCount: dateBeRCountMap[ymd],
		}
		item.BeRRate = fmt.Sprintf("%.2f%%", ComputeFloat(item.BeRCount, item.RegistCount)*100.0)
		// item.PayRate = fmt.Sprintf("%.2f%%", ComputeFloat(item.PayCount, item.RegistCount)*100.0)
		// item.RepayRate = fmt.Sprintf("%.2f%%", ComputeFloat(item.RepayCount, item.PayCount)*100)
		ret = append(ret, item)

		var datas []map[string]any
		// args := append([]any{ymd, stime, etime}, args0...)
		args := append([]any{stime, stime, etime, stime, etime}, args1...)
		err = ck.Select(&datas, sql0, args...)
		if err != nil {
			return
		}
		for _, data := range datas {
			diff_day := utils.ToInt64(data["diff_day"])
			count := utils.ToInt64(data["c"])
			rate := fmt.Sprintf("%.2f%%", ComputeFloat(count, item.BeRCount)*100.0)
			switch diff_day {
			case 1:
				item.Day1Rate = rate
			case 2:
				item.Day2Rate = rate
			case 3:
				item.Day3Rate = rate
			case 4:
				item.Day4Rate = rate
			case 5:
				item.Day5Rate = rate
			case 6:
				item.Day6Rate = rate
			case 7:
				item.Day7Rate = rate
			case 15:
				item.Day15Rate = rate
			case 30:
				item.Day30Rate = rate
			case 40:
				item.Day40Rate = rate
			case 50:
				item.Day50Rate = rate
			case 60:
				item.Day60Rate = rate
			case 90:
				item.Day90Rate = rate
			}
		}

		var days = map[uint32]*string{
			1:  &item.Day1Rate,
			2:  &item.Day2Rate,
			3:  &item.Day3Rate,
			4:  &item.Day4Rate,
			5:  &item.Day5Rate,
			6:  &item.Day6Rate,
			7:  &item.Day7Rate,
			15: &item.Day15Rate,
			30: &item.Day30Rate,
			40: &item.Day40Rate,
			50: &item.Day50Rate,
			60: &item.Day60Rate,
			90: &item.Day90Rate,
		}
		for day, field := range days {
			if *field == "" && ymd+day <= nowYMD {
				*field = "0.00%"
			}
		}
	}
	// 表格：倒序显示
	count := len(ret)
	for i := 0; i < count/2; i++ {
		ret[i], ret[count-1-i] = ret[count-1-i], ret[i]
	}
	return
}
