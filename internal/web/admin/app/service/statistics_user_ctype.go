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

// 用户分层inc

// 用户分层汇总表
func (this *statisticsService) UserLevelCtype(begin, end time.Time, packageIds, aliasIds, classIds []string) (list []*entity.UserLevelInfo, err error) {
	var where0 string
	var args0 = []any{begin, end}
	if len(packageIds) > 0 {
		where0 = " and ad__bundle_id in ?"
		args0 = append(args0, packageIds)
	}

	var datas1 []map[string]any
	err = ck.Select(&datas1, fmt.Sprintf(`
		SELECT toYYYYMMDD(ctime) regist_date, count(*) count, groupArray(DISTINCT ad__bundle_id) channels,
			(CASE WHEN money = 0 THEN (CASE state WHEN 1 THEN 0 ELSE 1 END) 
				WHEN money BETWEEN 1 and 100000-1 THEN 2 
				WHEN money BETWEEN 100000 and 1000000-1 THEN 3
				WHEN money between 1000000 and 10000000-1 THEN 4
				WHEN money between 10000000 and 20000000-1 THEN 5
				WHEN money >= 20000000 THEN 6 ELSE -1 END
			) ctype
		FROM game.col_user FINAL 
		WHERE ctime between ? AND ? %s
		GROUP BY regist_date, ctype
	`, where0), args0...)
	if err != nil {
		return
	}

	var datas2 []map[string]any
	err = ck.Select(&datas2, fmt.Sprintf(`
		SELECT t1.login_date,  count(*) count, groupArray(DISTINCT ad__bundle_id) channels,
			(CASE WHEN money = 0 THEN (CASE state WHEN 1 THEN 0 ELSE 1 END) 
				WHEN money BETWEEN 1 and 100000-1 THEN 2 
				WHEN money BETWEEN 100000 and 1000000-1 THEN 3
				WHEN money between 1000000 and 10000000-1 THEN 4
				WHEN money between 10000000 and 20000000-1 THEN 5
				WHEN money >= 20000000 THEN 6 ELSE -1 END
			) ctype
		FROM (
			SELECT toYYYYMMDD(login_time) login_date, userid
			FROM game.col_log_login FINAL
			WHERE login_time between ? AND ?
			GROUP BY login_date, userid
		) t1
		JOIN game.col_user t2 FINAL ON t1.userid = t2.userid
		WHERE 1 = 1 %s
		GROUP BY t1.login_date, ctype
	`, where0), args0...)
	if err != nil {
		return
	}

	var dateInfos = make(map[string]*entity.UserLevelInfo)
	// today := NowTime()
	for !begin.After(end) {
		stime := begin
		// etime := begin.AddDate(0, 0, 1)
		begin = begin.AddDate(0, 0, 1)
		sdate := stime.Format("2006-01-02")

		item := &entity.UserLevelInfo{
			Date:            sdate,
			ChannelMap:      make(map[string]bool),
			ChannelClassMap: make(map[string]bool),
		}
		list = append(list, item)
		dateInfos[sdate] = item
	}

	// 注册数据
	for _, data := range datas1 {
		regist_date := fmt.Sprint(data["regist_date"])
		sdate := regist_date[0:4] + "-" + regist_date[4:6] + "-" + regist_date[6:]
		count := utils.ToInt64(data["count"])
		ctype := utils.ToInt64(data["ctype"])
		channels := data["channels"].([]string)
		if item, ok := dateInfos[sdate]; ok {
			for _, c := range channels {
				if ok := item.ChannelMap[c]; !ok {
					item.ChannelMap[c] = true
					item.Channels = append(item.Channels, c)
				}
			}

			item.RegistCount += count
			switch ctype {
			case 0:
				item.Regist0Count = count
			case 1:
				item.Regist1Count = count
			case 2:
				item.Regist2Count = count
			case 3:
				item.Regist3Count = count
			case 4:
				item.Regist4Count = count
			case 5:
				item.Regist5Count = count
			case 6:
				item.Regist6Count = count
			}
		}
	}

	// 日活数据
	for _, data := range datas2 {
		login_date := fmt.Sprint(data["login_date"])
		sdate := login_date[0:4] + "-" + login_date[4:6] + "-" + login_date[6:]
		count := utils.ToInt64(data["count"])
		ctype := utils.ToInt64(data["ctype"])
		channels := data["channels"].([]string)
		if item, ok := dateInfos[sdate]; ok {
			for _, c := range channels {
				if ok := item.ChannelMap[c]; !ok {
					item.ChannelMap[c] = true
					item.Channels = append(item.Channels, c)
				}
			}

			item.ActiveCount += count
			switch ctype {
			case 0:
				item.Active0Count = count
			case 1:
				item.Active1Count = count
			case 2:
				item.Active2Count = count
			case 3:
				item.Active3Count = count
			case 4:
				item.Active4Count = count
			case 5:
				item.Active5Count = count
			case 6:
				item.Active6Count = count
			}
		}
	}

	// 所有渠道
	channelsMap, err := ChannelService.GetChannelsMap()
	if err != nil {
		return
	}

	var isAssignChannel bool = len(aliasIds) > 0 || len(classIds) > 0
	var assignAlias, assignClass string
	if isAssignChannel {
		sort.Strings(classIds)
		assignClass = strings.Join(classIds, ",")

		var aliasIdMap = make(map[string]bool)
		for _, channel := range channelsMap {
			if channel.ClassName != "" {
				for _, classId := range classIds {
					if channel.ClassName == classId {
						aliasIdMap[channel.Name1] = true
					}
				}
			}
		}
		for _, aliasId := range aliasIds {
			aliasIdMap[aliasId] = true
		}
		aliasIds = make([]string, 0)
		for aliasId := range aliasIdMap {
			aliasIds = append(aliasIds, aliasId)
		}
		sort.Strings(aliasIds)
		assignAlias = strings.Join(aliasIds, ",")
	}

	for _, item := range list {
		if item.RegistCount > 0 {
			item.Regist0Rate = fmt.Sprintf("%.2f%%", float64(item.Regist0Count)/float64(item.RegistCount)*100)
			item.Regist1Rate = fmt.Sprintf("%.2f%%", float64(item.Regist1Count)/float64(item.RegistCount)*100)
			item.Regist2Rate = fmt.Sprintf("%.2f%%", float64(item.Regist2Count)/float64(item.RegistCount)*100)
			item.Regist3Rate = fmt.Sprintf("%.2f%%", float64(item.Regist3Count)/float64(item.RegistCount)*100)
			item.Regist4Rate = fmt.Sprintf("%.2f%%", float64(item.Regist4Count)/float64(item.RegistCount)*100)
			item.Regist5Rate = fmt.Sprintf("%.2f%%", float64(item.Regist5Count)/float64(item.RegistCount)*100)
			item.Regist6Rate = fmt.Sprintf("%.2f%%", float64(item.Regist6Count)/float64(item.RegistCount)*100)
		}

		if item.ActiveCount > 0 {
			item.Active0Rate = fmt.Sprintf("%.2f%%", float64(item.Active0Count)/float64(item.ActiveCount)*100)
			item.Active1Rate = fmt.Sprintf("%.2f%%", float64(item.Active1Count)/float64(item.ActiveCount)*100)
			item.Active2Rate = fmt.Sprintf("%.2f%%", float64(item.Active2Count)/float64(item.ActiveCount)*100)
			item.Active3Rate = fmt.Sprintf("%.2f%%", float64(item.Active3Count)/float64(item.ActiveCount)*100)
			item.Active4Rate = fmt.Sprintf("%.2f%%", float64(item.Active4Count)/float64(item.ActiveCount)*100)
			item.Active5Rate = fmt.Sprintf("%.2f%%", float64(item.Active5Count)/float64(item.ActiveCount)*100)
			item.Active6Rate = fmt.Sprintf("%.2f%%", float64(item.Active6Count)/float64(item.ActiveCount)*100)
		}

		var channelClass []string // 渠道类
		var channels1 []string    // 别名
		for _, c := range item.Channels {
			if channel, ok := channelsMap[c]; ok {
				if channel.ClassName != "" {
					if ok := item.ChannelClassMap[channel.ClassName]; !ok {
						item.ChannelClassMap[channel.ClassName] = true
						channelClass = append(channelClass, channel.ClassName)
					}
				}
				if channel.Name1 != "" {
					channels1 = append(channels1, channel.Name1)
				}
			}
		}
		if isAssignChannel {
			item.ChannelClass = assignClass
			item.Channel1 = assignAlias
			item.Channels = aliasIds // 渠道别名汇总
		} else {
			item.ChannelClass = strings.Join(channelClass, ",")
			item.Channel1 = strings.Join(channels1, ",")
			item.Channels = channels1 // 渠道别名汇总
		}
	}

	// 汇总
	summary, err := this.UserLevelCtypeSummary(list, where0, args0)
	if err != nil {
		return
	}
	list = append(list, summary)

	// 表格：倒序显示
	count := len(list)
	for i := 0; i < count/2; i++ {
		list[i], list[count-1-i] = list[count-1-i], list[i]
	}
	return
}

// 用户分层汇总表汇总
func (this *statisticsService) UserLevelCtypeSummary(list []*entity.UserLevelInfo, where0 string, args0 []any) (summary *entity.UserLevelInfo, err error) {
	summary = &entity.UserLevelInfo{
		Date:            "总汇",
		ChannelMap:      make(map[string]bool),
		ChannelClassMap: make(map[string]bool),
	}
	for _, item := range list {
		for cs := range item.ChannelClassMap {
			summary.ChannelClassMap[cs] = true
			// if ok := summary.ChannelClassMap[cs]; !ok {
			// }
		}
		for _, c := range item.Channels {
			summary.ChannelMap[c] = true
			// if ok := summary.ChannelMap[c]; !ok {
			// 	summary.Channels = append(summary.Channels, c)
			// }
		}
		summary.RegistCount += item.RegistCount
		summary.Regist0Count += item.Regist0Count
		summary.Regist1Count += item.Regist1Count
		summary.Regist2Count += item.Regist2Count
		summary.Regist3Count += item.Regist3Count
		summary.Regist4Count += item.Regist4Count
		summary.Regist5Count += item.Regist5Count
		summary.Regist6Count += item.Regist6Count
	}

	if summary.RegistCount > 0 {
		summary.Regist0Rate = fmt.Sprintf("%.2f%%", float64(summary.Regist0Count)/float64(summary.RegistCount)*100)
		summary.Regist1Rate = fmt.Sprintf("%.2f%%", float64(summary.Regist1Count)/float64(summary.RegistCount)*100)
		summary.Regist2Rate = fmt.Sprintf("%.2f%%", float64(summary.Regist2Count)/float64(summary.RegistCount)*100)
		summary.Regist3Rate = fmt.Sprintf("%.2f%%", float64(summary.Regist3Count)/float64(summary.RegistCount)*100)
		summary.Regist4Rate = fmt.Sprintf("%.2f%%", float64(summary.Regist4Count)/float64(summary.RegistCount)*100)
		summary.Regist5Rate = fmt.Sprintf("%.2f%%", float64(summary.Regist5Count)/float64(summary.RegistCount)*100)
		summary.Regist6Rate = fmt.Sprintf("%.2f%%", float64(summary.Regist6Count)/float64(summary.RegistCount)*100)
	}

	var classes, channels1 []string
	for cs := range summary.ChannelClassMap {
		classes = append(classes, cs)
	}
	sort.Strings(classes)
	summary.ChannelClass = strings.Join(classes, ",")

	for c := range summary.ChannelMap {
		channels1 = append(channels1, c)
	}
	sort.Strings(channels1)
	summary.Channel1 = strings.Join(channels1, ",")

	// 日活汇总计算
	var datas2 []map[string]any
	err = ck.Select(&datas2, fmt.Sprintf(`
		SELECT count(*) count,
			(CASE WHEN money = 0 THEN (CASE state WHEN 1 THEN 0 ELSE 1 END) 
				WHEN money BETWEEN 1 and 100000-1 THEN 2 
				WHEN money BETWEEN 100000 and 1000000-1 THEN 3
				WHEN money between 1000000 and 10000000-1 THEN 4
				WHEN money between 10000000 and 20000000-1 THEN 5
				WHEN money >= 20000000 THEN 6 ELSE -1 END
			) ctype
		FROM (
			SELECT userid
			FROM game.col_log_login FINAL
			WHERE login_time between ? AND ?
			GROUP BY userid
		) t1
		JOIN game.col_user t2 FINAL ON t1.userid = t2.userid
		WHERE 1 = 1 %s
		GROUP BY ctype
	`, where0), args0...)
	if err != nil {
		return
	}

	// 日活数据总汇
	for _, data := range datas2 {
		count := utils.ToInt64(data["count"])
		ctype := utils.ToInt64(data["ctype"])
		summary.ActiveCount += count
		switch ctype {
		case 0:
			summary.Active0Count = count
		case 1:
			summary.Active1Count = count
		case 2:
			summary.Active2Count = count
		case 3:
			summary.Active3Count = count
		case 4:
			summary.Active4Count = count
		case 5:
			summary.Active5Count = count
		case 6:
			summary.Active6Count = count
		}
	}
	if summary.ActiveCount > 0 {
		summary.Active0Rate = fmt.Sprintf("%.2f%%", float64(summary.Active0Count)/float64(summary.ActiveCount)*100)
		summary.Active1Rate = fmt.Sprintf("%.2f%%", float64(summary.Active1Count)/float64(summary.ActiveCount)*100)
		summary.Active2Rate = fmt.Sprintf("%.2f%%", float64(summary.Active2Count)/float64(summary.ActiveCount)*100)
		summary.Active3Rate = fmt.Sprintf("%.2f%%", float64(summary.Active3Count)/float64(summary.ActiveCount)*100)
		summary.Active4Rate = fmt.Sprintf("%.2f%%", float64(summary.Active4Count)/float64(summary.ActiveCount)*100)
		summary.Active5Rate = fmt.Sprintf("%.2f%%", float64(summary.Active5Count)/float64(summary.ActiveCount)*100)
		summary.Active6Rate = fmt.Sprintf("%.2f%%", float64(summary.Active6Count)/float64(summary.ActiveCount)*100)
	}
	return
}

// 用户转化效率表
// Deprecated
func (this *statisticsService) UserTransEffect2(begin, end time.Time, packageIds []string) (list []*entity.UserTransEffect, err error) {
	var datas0 []map[string]any
	var where0 string
	var args0 = []any{begin, end}
	if len(packageIds) > 0 {
		where0 = " and ad__bundle_id in ?"
		args0 = append(args0, packageIds)
	}

	err = ck.Select(&datas0, fmt.Sprintf(`
		SELECT regist_date, count(*) regist_count, groupArray(DISTINCT ad__bundle_id) channels,
			SUM(CASE WHEN pay1 = 0 THEN 1 ELSE 0 END) c0_1,
			SUM(CASE WHEN pay3 = 0 THEN 1 ELSE 0 END) c0_3,
			SUM(CASE WHEN pay7 = 0 THEN 1 ELSE 0 END) c0_7,
			SUM(CASE WHEN pay10 = 0 THEN 1 ELSE 0 END) c0_10,
			SUM(CASE WHEN pay15 = 0 THEN 1 ELSE 0 END) c0_15,
			SUM(CASE WHEN pay20 = 0 THEN 1 ELSE 0 END) c0_20,
			SUM(CASE WHEN pay25 = 0 THEN 1 ELSE 0 END) c0_25,
			SUM(CASE WHEN pay30 = 0 THEN 1 ELSE 0 END) c0_30,
			SUM(CASE WHEN pay1 >= 1 THEN 1 ELSE 0 END) c2_1,
			SUM(CASE WHEN pay3 >= 1 THEN 1 ELSE 0 END) c2_3,
			SUM(CASE WHEN pay7 >= 1 THEN 1 ELSE 0 END) c2_7,
			SUM(CASE WHEN pay10 >= 1 THEN 1 ELSE 0 END) c2_10,
			SUM(CASE WHEN pay15 >= 1 THEN 1 ELSE 0 END) c2_15,
			SUM(CASE WHEN pay20 >= 1 THEN 1 ELSE 0 END) c2_20,
			SUM(CASE WHEN pay25 >= 1 THEN 1 ELSE 0 END) c2_25,
			SUM(CASE WHEN pay30 >= 1 THEN 1 ELSE 0 END) c2_30,
			SUM(CASE WHEN pay1 >= 100000 THEN 1 ELSE 0 END) c3_1,
			SUM(CASE WHEN pay3 >= 100000 THEN 1 ELSE 0 END) c3_3,
			SUM(CASE WHEN pay7 >= 100000 THEN 1 ELSE 0 END) c3_7,
			SUM(CASE WHEN pay10 >= 100000 THEN 1 ELSE 0 END) c3_10,
			SUM(CASE WHEN pay15 >= 100000 THEN 1 ELSE 0 END) c3_15,
			SUM(CASE WHEN pay20 >= 100000 THEN 1 ELSE 0 END) c3_20,
			SUM(CASE WHEN pay25 >= 100000 THEN 1 ELSE 0 END) c3_25,
			SUM(CASE WHEN pay30 >= 100000 THEN 1 ELSE 0 END) c3_30,
			SUM(CASE WHEN pay1 >= 1000000 THEN 1 ELSE 0 END) c4_1,
			SUM(CASE WHEN pay3 >= 1000000 THEN 1 ELSE 0 END) c4_3,
			SUM(CASE WHEN pay7 >= 1000000 THEN 1 ELSE 0 END) c4_7,
			SUM(CASE WHEN pay10 >= 1000000 THEN 1 ELSE 0 END) c4_10,
			SUM(CASE WHEN pay15 >= 1000000 THEN 1 ELSE 0 END) c4_15,
			SUM(CASE WHEN pay20 >= 1000000 THEN 1 ELSE 0 END) c4_20,
			SUM(CASE WHEN pay25 >= 1000000 THEN 1 ELSE 0 END) c4_25,
			SUM(CASE WHEN pay30 >= 1000000 THEN 1 ELSE 0 END) c4_30,
			SUM(CASE WHEN pay1 >= 10000000 THEN 1 ELSE 0 END) c5_1,
			SUM(CASE WHEN pay3 >= 10000000 THEN 1 ELSE 0 END) c5_3,
			SUM(CASE WHEN pay7 >= 10000000 THEN 1 ELSE 0 END) c5_7,
			SUM(CASE WHEN pay10 >= 10000000 THEN 1 ELSE 0 END) c5_10,
			SUM(CASE WHEN pay15 >= 10000000 THEN 1 ELSE 0 END) c5_15,
			SUM(CASE WHEN pay20 >= 10000000 THEN 1 ELSE 0 END) c5_20,
			SUM(CASE WHEN pay25 >= 10000000 THEN 1 ELSE 0 END) c5_25,
			SUM(CASE WHEN pay30 >= 10000000 THEN 1 ELSE 0 END) c5_30,
			SUM(CASE WHEN pay1 >= 20000000 THEN 1 ELSE 0 END) c6_1,
			SUM(CASE WHEN pay3 >= 20000000 THEN 1 ELSE 0 END) c6_3,
			SUM(CASE WHEN pay7 >= 20000000 THEN 1 ELSE 0 END) c6_7,
			SUM(CASE WHEN pay10 >= 20000000 THEN 1 ELSE 0 END) c6_10,
			SUM(CASE WHEN pay15 >= 20000000 THEN 1 ELSE 0 END) c6_15,
			SUM(CASE WHEN pay20 >= 20000000 THEN 1 ELSE 0 END) c6_20,
			SUM(CASE WHEN pay25 >= 20000000 THEN 1 ELSE 0 END) c6_25,
			SUM(CASE WHEN pay30 >= 20000000 THEN 1 ELSE 0 END) c6_30
		FROM (
			SELECT t1.regist_date, t1.userid, t1.ad__bundle_id,
				SUM(CASE WHEN t2.pay_date - t1.regist_date = 0 THEN pay_amount ELSE 0 END) pay1,
				SUM(CASE WHEN t2.pay_date - t1.regist_date BETWEEN 0 AND 2 THEN pay_amount ELSE 0 END) pay3,
				SUM(CASE WHEN t2.pay_date - t1.regist_date BETWEEN 0 AND 6 THEN pay_amount ELSE 0 END) pay7,
				SUM(CASE WHEN t2.pay_date - t1.regist_date BETWEEN 0 AND 9 THEN pay_amount ELSE 0 END) pay10,
				SUM(CASE WHEN t2.pay_date - t1.regist_date BETWEEN 0 AND 14 THEN pay_amount ELSE 0 END) pay15,
				SUM(CASE WHEN t2.pay_date - t1.regist_date BETWEEN 0 AND 19 THEN pay_amount ELSE 0 END) pay20,
				SUM(CASE WHEN t2.pay_date - t1.regist_date BETWEEN 0 AND 24 THEN pay_amount ELSE 0 END) pay25,
				SUM(CASE WHEN t2.pay_date - t1.regist_date BETWEEN 0 AND 29 THEN pay_amount ELSE 0 END) pay30
			FROM (
				SELECT userid, toYYYYMMDD(ctime) regist_date, ad__bundle_id
				FROM game.col_user FINAL
				WHERE ctime between ? AND ? %s
			) t1
			LEFT JOIN (
				SELECT toYYYYMMDD(ctime) pay_date, userid, SUM(amount) pay_amount
				FROM game.col_trade_record t2 FINAL
				WHERE order_status = 4 AND ctime > ?
				GROUP BY pay_date, userid
			) t2 ON t1.userid = t2.userid
			GROUP BY t1.regist_date, t1.userid, t1.ad__bundle_id
		) s1
		GROUP BY regist_date
	`, where0), append(args0, begin)...)
	if err != nil {
		return
	}

	var dateInfos = make(map[string]*entity.UserTransEffect)
	for !begin.After(end) {
		stime := begin
		begin = begin.AddDate(0, 0, 1)
		sdate := stime.Format("2006-01-02")

		item := &entity.UserTransEffect{
			SDate: sdate,
		}
		list = append(list, item)
		dateInfos[sdate] = item
	}

	// 所有渠道
	channelsMap, err := ChannelService.GetChannelsMap()
	if err != nil {
		return
	}

	var allChannelsClassMap = make(map[string]bool)
	var allChannels1Map = make(map[string]bool)
	var allChannels1 []string

	// 注册数据
	for _, data := range datas0 {
		regist_date := fmt.Sprint(data["regist_date"])
		sdate := regist_date[0:4] + "-" + regist_date[4:6] + "-" + regist_date[6:]
		if item, ok := dateInfos[sdate]; ok {
			channels := data["channels"].([]string)
			var channelsClassMap = make(map[string]bool)
			var channels1Map = make(map[string]bool)
			var channels1 []string
			for _, ch := range channels {
				if channel, ok := channelsMap[ch]; ok {
					if channel.ClassName != "" {
						allChannelsClassMap[channel.ClassName] = true
						channelsClassMap[channel.ClassName] = true
					}
					if channel.Name1 != "" {
						if ok := allChannels1Map[channel.Name1]; !ok {
							allChannels1Map[channel.Name1] = true
							allChannels1 = append(allChannels1, channel.Name1)
						}
						if ok := channels1Map[channel.Name1]; !ok {
							channels1Map[channel.Name1] = true
							channels1 = append(channels1, channel.Name1)
						}
					}
				}
			}
			item.Channel1 = strings.Join(channels1, ",")
			var channelClasses []string
			for c := range channelsClassMap {
				channelClasses = append(channelClasses, c)
			}
			item.ChannelClass = strings.Join(channelClasses, ",")

			item.RegistCount = utils.ToInt64(data["regist_count"])
			item.C0_1 = utils.ToInt64(data["c0_1"])
			item.C0_3 = utils.ToInt64(data["c0_3"])
			item.C0_7 = utils.ToInt64(data["c0_7"])
			item.C0_10 = utils.ToInt64(data["c0_10"])
			item.C0_15 = utils.ToInt64(data["c0_15"])
			item.C0_20 = utils.ToInt64(data["c0_20"])
			item.C0_25 = utils.ToInt64(data["c0_25"])
			item.C0_30 = utils.ToInt64(data["c0_30"])
			item.C2_1 = utils.ToInt64(data["c2_1"])
			item.C2_3 = utils.ToInt64(data["c2_3"])
			item.C2_7 = utils.ToInt64(data["c2_7"])
			item.C2_10 = utils.ToInt64(data["c2_10"])
			item.C2_15 = utils.ToInt64(data["c2_15"])
			item.C2_20 = utils.ToInt64(data["c2_20"])
			item.C2_25 = utils.ToInt64(data["c2_25"])
			item.C2_30 = utils.ToInt64(data["c2_30"])
			item.C3_1 = utils.ToInt64(data["c3_1"])
			item.C3_3 = utils.ToInt64(data["c3_3"])
			item.C3_7 = utils.ToInt64(data["c3_7"])
			item.C3_10 = utils.ToInt64(data["c3_10"])
			item.C3_15 = utils.ToInt64(data["c3_15"])
			item.C3_20 = utils.ToInt64(data["c3_20"])
			item.C3_25 = utils.ToInt64(data["c3_25"])
			item.C3_30 = utils.ToInt64(data["c3_30"])
			item.C4_1 = utils.ToInt64(data["c4_1"])
			item.C4_3 = utils.ToInt64(data["c4_3"])
			item.C4_7 = utils.ToInt64(data["c4_7"])
			item.C4_10 = utils.ToInt64(data["c4_10"])
			item.C4_15 = utils.ToInt64(data["c4_15"])
			item.C4_20 = utils.ToInt64(data["c4_20"])
			item.C4_25 = utils.ToInt64(data["c4_25"])
			item.C4_30 = utils.ToInt64(data["c4_30"])
			item.C5_1 = utils.ToInt64(data["c5_1"])
			item.C5_3 = utils.ToInt64(data["c5_3"])
			item.C5_7 = utils.ToInt64(data["c5_7"])
			item.C5_10 = utils.ToInt64(data["c5_10"])
			item.C5_15 = utils.ToInt64(data["c5_15"])
			item.C5_20 = utils.ToInt64(data["c5_20"])
			item.C5_25 = utils.ToInt64(data["c5_25"])
			item.C5_30 = utils.ToInt64(data["c5_30"])
			item.C6_1 = utils.ToInt64(data["c6_1"])
			item.C6_3 = utils.ToInt64(data["c6_3"])
			item.C6_7 = utils.ToInt64(data["c6_7"])
			item.C6_10 = utils.ToInt64(data["c6_10"])
			item.C6_15 = utils.ToInt64(data["c6_15"])
			item.C6_20 = utils.ToInt64(data["c6_20"])
			item.C6_25 = utils.ToInt64(data["c6_25"])
			item.C6_30 = utils.ToInt64(data["c6_30"])
			if item.RegistCount > 0 {
				item.C0Rate_1 = fmt.Sprintf("%.2f%%", float64(item.C0_1)/float64(item.RegistCount)*100)
				item.C0Rate_3 = fmt.Sprintf("%.2f%%", float64(item.C0_3)/float64(item.RegistCount)*100)
				item.C0Rate_7 = fmt.Sprintf("%.2f%%", float64(item.C0_7)/float64(item.RegistCount)*100)
				item.C0Rate_10 = fmt.Sprintf("%.2f%%", float64(item.C0_10)/float64(item.RegistCount)*100)
				item.C0Rate_15 = fmt.Sprintf("%.2f%%", float64(item.C0_15)/float64(item.RegistCount)*100)
				item.C0Rate_20 = fmt.Sprintf("%.2f%%", float64(item.C0_20)/float64(item.RegistCount)*100)
				item.C0Rate_25 = fmt.Sprintf("%.2f%%", float64(item.C0_25)/float64(item.RegistCount)*100)
				item.C0Rate_30 = fmt.Sprintf("%.2f%%", float64(item.C0_30)/float64(item.RegistCount)*100)
				item.C2Rate_1 = fmt.Sprintf("%.2f%%", float64(item.C2_1)/float64(item.RegistCount)*100)
				item.C2Rate_3 = fmt.Sprintf("%.2f%%", float64(item.C2_3)/float64(item.RegistCount)*100)
				item.C2Rate_7 = fmt.Sprintf("%.2f%%", float64(item.C2_7)/float64(item.RegistCount)*100)
				item.C2Rate_10 = fmt.Sprintf("%.2f%%", float64(item.C2_10)/float64(item.RegistCount)*100)
				item.C2Rate_15 = fmt.Sprintf("%.2f%%", float64(item.C2_15)/float64(item.RegistCount)*100)
				item.C2Rate_20 = fmt.Sprintf("%.2f%%", float64(item.C2_20)/float64(item.RegistCount)*100)
				item.C2Rate_25 = fmt.Sprintf("%.2f%%", float64(item.C2_25)/float64(item.RegistCount)*100)
				item.C2Rate_30 = fmt.Sprintf("%.2f%%", float64(item.C2_30)/float64(item.RegistCount)*100)
				item.C3Rate_1 = fmt.Sprintf("%.2f%%", float64(item.C3_1)/float64(item.RegistCount)*100)
				item.C3Rate_3 = fmt.Sprintf("%.2f%%", float64(item.C3_3)/float64(item.RegistCount)*100)
				item.C3Rate_7 = fmt.Sprintf("%.2f%%", float64(item.C3_7)/float64(item.RegistCount)*100)
				item.C3Rate_10 = fmt.Sprintf("%.2f%%", float64(item.C3_10)/float64(item.RegistCount)*100)
				item.C3Rate_15 = fmt.Sprintf("%.2f%%", float64(item.C3_15)/float64(item.RegistCount)*100)
				item.C3Rate_20 = fmt.Sprintf("%.2f%%", float64(item.C3_20)/float64(item.RegistCount)*100)
				item.C3Rate_25 = fmt.Sprintf("%.2f%%", float64(item.C3_25)/float64(item.RegistCount)*100)
				item.C3Rate_30 = fmt.Sprintf("%.2f%%", float64(item.C3_30)/float64(item.RegistCount)*100)
				item.C4Rate_1 = fmt.Sprintf("%.2f%%", float64(item.C4_1)/float64(item.RegistCount)*100)
				item.C4Rate_3 = fmt.Sprintf("%.2f%%", float64(item.C4_3)/float64(item.RegistCount)*100)
				item.C4Rate_7 = fmt.Sprintf("%.2f%%", float64(item.C4_7)/float64(item.RegistCount)*100)
				item.C4Rate_10 = fmt.Sprintf("%.2f%%", float64(item.C4_10)/float64(item.RegistCount)*100)
				item.C4Rate_15 = fmt.Sprintf("%.2f%%", float64(item.C4_15)/float64(item.RegistCount)*100)
				item.C4Rate_20 = fmt.Sprintf("%.2f%%", float64(item.C4_20)/float64(item.RegistCount)*100)
				item.C4Rate_25 = fmt.Sprintf("%.2f%%", float64(item.C4_25)/float64(item.RegistCount)*100)
				item.C4Rate_30 = fmt.Sprintf("%.2f%%", float64(item.C4_30)/float64(item.RegistCount)*100)
				item.C5Rate_1 = fmt.Sprintf("%.2f%%", float64(item.C5_1)/float64(item.RegistCount)*100)
				item.C5Rate_3 = fmt.Sprintf("%.2f%%", float64(item.C5_3)/float64(item.RegistCount)*100)
				item.C5Rate_7 = fmt.Sprintf("%.2f%%", float64(item.C5_7)/float64(item.RegistCount)*100)
				item.C5Rate_10 = fmt.Sprintf("%.2f%%", float64(item.C5_10)/float64(item.RegistCount)*100)
				item.C5Rate_15 = fmt.Sprintf("%.2f%%", float64(item.C5_15)/float64(item.RegistCount)*100)
				item.C5Rate_20 = fmt.Sprintf("%.2f%%", float64(item.C5_20)/float64(item.RegistCount)*100)
				item.C5Rate_25 = fmt.Sprintf("%.2f%%", float64(item.C5_25)/float64(item.RegistCount)*100)
				item.C5Rate_30 = fmt.Sprintf("%.2f%%", float64(item.C5_30)/float64(item.RegistCount)*100)
				item.C6Rate_1 = fmt.Sprintf("%.2f%%", float64(item.C6_1)/float64(item.RegistCount)*100)
				item.C6Rate_3 = fmt.Sprintf("%.2f%%", float64(item.C6_3)/float64(item.RegistCount)*100)
				item.C6Rate_7 = fmt.Sprintf("%.2f%%", float64(item.C6_7)/float64(item.RegistCount)*100)
				item.C6Rate_10 = fmt.Sprintf("%.2f%%", float64(item.C6_10)/float64(item.RegistCount)*100)
				item.C6Rate_15 = fmt.Sprintf("%.2f%%", float64(item.C6_15)/float64(item.RegistCount)*100)
				item.C6Rate_20 = fmt.Sprintf("%.2f%%", float64(item.C6_20)/float64(item.RegistCount)*100)
				item.C6Rate_25 = fmt.Sprintf("%.2f%%", float64(item.C6_25)/float64(item.RegistCount)*100)
				item.C6Rate_30 = fmt.Sprintf("%.2f%%", float64(item.C6_30)/float64(item.RegistCount)*100)
			}
		}
	}

	// 日活计算
	var datas2 []map[string]any
	err = ck.Select(&datas2, fmt.Sprintf(`
		SELECT t1.login_date, count(*) count
		FROM (
			SELECT toYYYYMMDD(login_time) login_date, userid
			FROM game.col_log_login FINAL
			WHERE login_time between ? AND ?
			GROUP BY login_date, userid
		) t1
		JOIN game.col_user t2 FINAL ON t1.userid = t2.userid
		WHERE 1 = 1 %s
		GROUP BY t1.login_date
	`, where0), args0...)
	if err != nil {
		return
	}

	// 日活数据
	for _, data := range datas2 {
		login_date := fmt.Sprint(data["login_date"])
		sdate := login_date[0:4] + "-" + login_date[4:6] + "-" + login_date[6:]
		if item, ok := dateInfos[sdate]; ok {
			item.ActiveCount = utils.ToInt64(data["count"])
		}
	}

	summary, err := this.UserTransEffectSummary(list, where0, args0)
	if err != nil {
		return
	}
	summary.Channel1 = strings.Join(allChannels1, ",")
	var allCs []string
	for c := range allChannelsClassMap {
		allCs = append(allCs, c)
	}
	summary.ChannelClass = strings.Join(allCs, ",")
	list = append(list, summary)

	// 表格：倒序显示
	count := len(list)
	for i := 0; i < count/2; i++ {
		list[i], list[count-1-i] = list[count-1-i], list[i]
	}
	return
}

// 用户转化效率表
func (this *statisticsService) UserTransEffect(begin, end time.Time, packageIds, aliasIds, classIds []string) (list []*entity.UserTransEffect, err error) {
	var datas0 []map[string]any
	var where0 string
	var args0 = []any{begin, end}
	if len(packageIds) > 0 {
		where0 = " and ad__bundle_id in ?"
		args0 = append(args0, packageIds)
	}

	err = ck.Select(&datas0, fmt.Sprintf(`
		-- SELECT regist_date, count(*) regist_count, groupArray(DISTINCT ad__bundle_id) channels,
		
		SELECT regist_date, day, ctype_day, count(*) count, groupArray(DISTINCT ad__bundle_id) channels
		FROM (
			SELECT t2.regist_date, t2.userid userid, t2.ad__bundle_id, t3.number day, SUM(t1.amount) pay_amount,
				SUM(CASE WHEN toYYYYMMDD(t1.ctime) - t2.regist_date BETWEEN 0 AND t3.number THEN t1.amount ELSE 0 END) pay_amount_day,
				(CASE WHEN pay_amount_day = 0 THEN 0 
				WHEN pay_amount_day BETWEEN 1 and 100000-1 THEN 2 
				WHEN pay_amount_day BETWEEN 100000 and 1000000-1 THEN 3
				WHEN pay_amount_day between 1000000 and 10000000-1 THEN 4
				WHEN pay_amount_day between 10000000 and 20000000-1 THEN 5
				WHEN pay_amount_day >= 20000000 THEN 6 ELSE -1 END) ctype_day
			FROM game.col_trade_record t1 FINAL
			RIGHT JOIN (
				SELECT userid, toYYYYMMDD(ctime) regist_date, ad__bundle_id
				FROM game.col_user FINAL 
				WHERE ctime BETWEEN ? AND ? %s
			) t2 ON t1.userid = t2.userid AND t1.order_status = 4
			JOIN (
				SELECT number FROM system.numbers WHERE number IN (0,2,6,9,14,19,24,29)
			) t3 ON 1 = 1
			GROUP BY t2.regist_date, t2.userid, t2.ad__bundle_id, t3.number
		) s1
		GROUP BY regist_date, day, ctype_day
		ORDER BY regist_date, day, ctype_day
	`, where0), append(args0, begin)...)
	if err != nil {
		return
	}

	var dateInfos = make(map[string]*entity.UserTransEffect)
	for !begin.After(end) {
		stime := begin
		begin = begin.AddDate(0, 0, 1)
		sdate := stime.Format("2006-01-02")

		item := &entity.UserTransEffect{
			SDate:             sdate,
			Channels1Map:      make(map[string]bool),
			ChannelClassesMap: make(map[string]bool),
		}
		list = append(list, item)
		dateInfos[sdate] = item
	}

	// 所有渠道
	channelsMap, err := ChannelService.GetChannelsMap()
	if err != nil {
		return
	}

	var isAssignChannel bool = len(aliasIds) > 0 || len(classIds) > 0
	var assignAlias, assignClass string
	if isAssignChannel {
		sort.Strings(classIds)
		assignClass = strings.Join(classIds, ",")

		var aliasIdMap = make(map[string]bool)
		for _, channel := range channelsMap {
			if channel.ClassName != "" {
				for _, classId := range classIds {
					if channel.ClassName == classId {
						aliasIdMap[channel.Name1] = true
					}
				}
			}
		}
		for _, aliasId := range aliasIds {
			aliasIdMap[aliasId] = true
		}
		aliasIds = make([]string, 0)
		for aliasId := range aliasIdMap {
			aliasIds = append(aliasIds, aliasId)
		}
		sort.Strings(aliasIds)
		assignAlias = strings.Join(aliasIds, ",")
	}

	var allChannelsClassMap = make(map[string]bool)
	var allChannels1Map = make(map[string]bool)

	// 注册数据
	for _, data := range datas0 {
		regist_date := fmt.Sprint(data["regist_date"])
		sdate := regist_date[0:4] + "-" + regist_date[4:6] + "-" + regist_date[6:]
		day := utils.ToInt64(data["day"])
		ctype_day := utils.ToInt64(data["ctype_day"])
		count := utils.ToInt64(data["count"])
		channels := data["channels"].([]string)
		if item, ok := dateInfos[sdate]; ok {
			for _, ch := range channels {
				if channel, ok := channelsMap[ch]; ok {
					if channel.ClassName != "" {
						allChannelsClassMap[channel.ClassName] = true
						item.ChannelClassesMap[channel.ClassName] = true
					}
					if channel.Name1 != "" {
						allChannels1Map[channel.Name1] = true
						item.Channels1Map[channel.Name1] = true
					}
				}
			}
			if day == 0 {
				item.RegistCount += count
			}
			switch ctype_day {
			case 0:
				switch day { // 0,2,6,9,14,19,24,29
				case 0:
					item.C0_1 = count
				case 2:
					item.C0_3 = count
				case 6:
					item.C0_7 = count
				case 9:
					item.C0_10 = count
				case 14:
					item.C0_15 = count
				case 19:
					item.C0_20 = count
				case 24:
					item.C0_25 = count
				case 29:
					item.C0_30 = count
				}
			case 2:
				switch day {
				case 0:
					item.C2_1 = count
				case 2:
					item.C2_3 = count
				case 6:
					item.C2_7 = count
				case 9:
					item.C2_10 = count
				case 14:
					item.C2_15 = count
				case 19:
					item.C2_20 = count
				case 24:
					item.C2_25 = count
				case 29:
					item.C2_30 = count
				}
			case 3:
				switch day {
				case 0:
					item.C3_1 = count
				case 2:
					item.C3_3 = count
				case 6:
					item.C3_7 = count
				case 9:
					item.C3_10 = count
				case 14:
					item.C3_15 = count
				case 19:
					item.C3_20 = count
				case 24:
					item.C3_25 = count
				case 29:
					item.C3_30 = count
				}
			case 4:

				switch day {
				case 0:
					item.C4_1 = count
				case 2:
					item.C4_3 = count
				case 6:
					item.C4_7 = count
				case 9:
					item.C4_10 = count
				case 14:
					item.C4_15 = count
				case 19:
					item.C4_20 = count
				case 24:
					item.C4_25 = count
				case 29:
					item.C4_30 = count
				}
			case 5:
				switch day {
				case 0:
					item.C5_1 = count
				case 2:
					item.C5_3 = count
				case 6:
					item.C5_7 = count
				case 9:
					item.C5_10 = count
				case 14:
					item.C5_15 = count
				case 19:
					item.C5_20 = count
				case 24:
					item.C5_25 = count
				case 29:
					item.C5_30 = count
				}
			case 6:
				switch day {
				case 0:
					item.C6_1 = count
				case 2:
					item.C6_3 = count
				case 6:
					item.C6_7 = count
				case 9:
					item.C6_10 = count
				case 14:
					item.C6_15 = count
				case 19:
					item.C6_20 = count
				case 24:
					item.C6_25 = count
				case 29:
					item.C6_30 = count
				}
			}

		}
	}

	for _, item := range dateInfos {
		var aliases, classes []string
		for alias := range item.Channels1Map {
			aliases = append(aliases, alias)
		}
		for cls := range item.ChannelClassesMap {
			classes = append(classes, cls)
		}
		if isAssignChannel {
			item.Channel1 = assignAlias
			item.ChannelClass = assignClass
		} else {
			sort.Strings(aliases)
			sort.Strings(classes)
			item.Channel1 = strings.Join(aliases, ",")
			item.ChannelClass = strings.Join(classes, ",")
		}

		if item.RegistCount > 0 {
			item.C0Rate_1 = fmt.Sprintf("%.2f%%", float64(item.C0_1)/float64(item.RegistCount)*100)
			item.C0Rate_3 = fmt.Sprintf("%.2f%%", float64(item.C0_3)/float64(item.RegistCount)*100)
			item.C0Rate_7 = fmt.Sprintf("%.2f%%", float64(item.C0_7)/float64(item.RegistCount)*100)
			item.C0Rate_10 = fmt.Sprintf("%.2f%%", float64(item.C0_10)/float64(item.RegistCount)*100)
			item.C0Rate_15 = fmt.Sprintf("%.2f%%", float64(item.C0_15)/float64(item.RegistCount)*100)
			item.C0Rate_20 = fmt.Sprintf("%.2f%%", float64(item.C0_20)/float64(item.RegistCount)*100)
			item.C0Rate_25 = fmt.Sprintf("%.2f%%", float64(item.C0_25)/float64(item.RegistCount)*100)
			item.C0Rate_30 = fmt.Sprintf("%.2f%%", float64(item.C0_30)/float64(item.RegistCount)*100)
			item.C2Rate_1 = fmt.Sprintf("%.2f%%", float64(item.C2_1)/float64(item.RegistCount)*100)
			item.C2Rate_3 = fmt.Sprintf("%.2f%%", float64(item.C2_3)/float64(item.RegistCount)*100)
			item.C2Rate_7 = fmt.Sprintf("%.2f%%", float64(item.C2_7)/float64(item.RegistCount)*100)
			item.C2Rate_10 = fmt.Sprintf("%.2f%%", float64(item.C2_10)/float64(item.RegistCount)*100)
			item.C2Rate_15 = fmt.Sprintf("%.2f%%", float64(item.C2_15)/float64(item.RegistCount)*100)
			item.C2Rate_20 = fmt.Sprintf("%.2f%%", float64(item.C2_20)/float64(item.RegistCount)*100)
			item.C2Rate_25 = fmt.Sprintf("%.2f%%", float64(item.C2_25)/float64(item.RegistCount)*100)
			item.C2Rate_30 = fmt.Sprintf("%.2f%%", float64(item.C2_30)/float64(item.RegistCount)*100)
			item.C3Rate_1 = fmt.Sprintf("%.2f%%", float64(item.C3_1)/float64(item.RegistCount)*100)
			item.C3Rate_3 = fmt.Sprintf("%.2f%%", float64(item.C3_3)/float64(item.RegistCount)*100)
			item.C3Rate_7 = fmt.Sprintf("%.2f%%", float64(item.C3_7)/float64(item.RegistCount)*100)
			item.C3Rate_10 = fmt.Sprintf("%.2f%%", float64(item.C3_10)/float64(item.RegistCount)*100)
			item.C3Rate_15 = fmt.Sprintf("%.2f%%", float64(item.C3_15)/float64(item.RegistCount)*100)
			item.C3Rate_20 = fmt.Sprintf("%.2f%%", float64(item.C3_20)/float64(item.RegistCount)*100)
			item.C3Rate_25 = fmt.Sprintf("%.2f%%", float64(item.C3_25)/float64(item.RegistCount)*100)
			item.C3Rate_30 = fmt.Sprintf("%.2f%%", float64(item.C3_30)/float64(item.RegistCount)*100)
			item.C4Rate_1 = fmt.Sprintf("%.2f%%", float64(item.C4_1)/float64(item.RegistCount)*100)
			item.C4Rate_3 = fmt.Sprintf("%.2f%%", float64(item.C4_3)/float64(item.RegistCount)*100)
			item.C4Rate_7 = fmt.Sprintf("%.2f%%", float64(item.C4_7)/float64(item.RegistCount)*100)
			item.C4Rate_10 = fmt.Sprintf("%.2f%%", float64(item.C4_10)/float64(item.RegistCount)*100)
			item.C4Rate_15 = fmt.Sprintf("%.2f%%", float64(item.C4_15)/float64(item.RegistCount)*100)
			item.C4Rate_20 = fmt.Sprintf("%.2f%%", float64(item.C4_20)/float64(item.RegistCount)*100)
			item.C4Rate_25 = fmt.Sprintf("%.2f%%", float64(item.C4_25)/float64(item.RegistCount)*100)
			item.C4Rate_30 = fmt.Sprintf("%.2f%%", float64(item.C4_30)/float64(item.RegistCount)*100)
			item.C5Rate_1 = fmt.Sprintf("%.2f%%", float64(item.C5_1)/float64(item.RegistCount)*100)
			item.C5Rate_3 = fmt.Sprintf("%.2f%%", float64(item.C5_3)/float64(item.RegistCount)*100)
			item.C5Rate_7 = fmt.Sprintf("%.2f%%", float64(item.C5_7)/float64(item.RegistCount)*100)
			item.C5Rate_10 = fmt.Sprintf("%.2f%%", float64(item.C5_10)/float64(item.RegistCount)*100)
			item.C5Rate_15 = fmt.Sprintf("%.2f%%", float64(item.C5_15)/float64(item.RegistCount)*100)
			item.C5Rate_20 = fmt.Sprintf("%.2f%%", float64(item.C5_20)/float64(item.RegistCount)*100)
			item.C5Rate_25 = fmt.Sprintf("%.2f%%", float64(item.C5_25)/float64(item.RegistCount)*100)
			item.C5Rate_30 = fmt.Sprintf("%.2f%%", float64(item.C5_30)/float64(item.RegistCount)*100)
			item.C6Rate_1 = fmt.Sprintf("%.2f%%", float64(item.C6_1)/float64(item.RegistCount)*100)
			item.C6Rate_3 = fmt.Sprintf("%.2f%%", float64(item.C6_3)/float64(item.RegistCount)*100)
			item.C6Rate_7 = fmt.Sprintf("%.2f%%", float64(item.C6_7)/float64(item.RegistCount)*100)
			item.C6Rate_10 = fmt.Sprintf("%.2f%%", float64(item.C6_10)/float64(item.RegistCount)*100)
			item.C6Rate_15 = fmt.Sprintf("%.2f%%", float64(item.C6_15)/float64(item.RegistCount)*100)
			item.C6Rate_20 = fmt.Sprintf("%.2f%%", float64(item.C6_20)/float64(item.RegistCount)*100)
			item.C6Rate_25 = fmt.Sprintf("%.2f%%", float64(item.C6_25)/float64(item.RegistCount)*100)
			item.C6Rate_30 = fmt.Sprintf("%.2f%%", float64(item.C6_30)/float64(item.RegistCount)*100)
		}
	}

	// 日活计算
	var datas2 []map[string]any
	err = ck.Select(&datas2, fmt.Sprintf(`
		SELECT t1.login_date, count(*) count
		FROM (
			SELECT toYYYYMMDD(login_time) login_date, userid
			FROM game.col_log_login FINAL
			WHERE login_time between ? AND ?
			GROUP BY login_date, userid
		) t1
		JOIN game.col_user t2 FINAL ON t1.userid = t2.userid
		WHERE 1 = 1 %s
		GROUP BY t1.login_date
	`, where0), args0...)
	if err != nil {
		return
	}

	// 日活数据
	for _, data := range datas2 {
		login_date := fmt.Sprint(data["login_date"])
		sdate := login_date[0:4] + "-" + login_date[4:6] + "-" + login_date[6:]
		if item, ok := dateInfos[sdate]; ok {
			item.ActiveCount = utils.ToInt64(data["count"])
		}
	}

	summary, err := this.UserTransEffectSummary(list, where0, args0)
	if err != nil {
		return
	}
	var aliases, classes []string
	for alias := range allChannels1Map {
		aliases = append(aliases, alias)
	}
	for cls := range allChannelsClassMap {
		classes = append(classes, cls)
	}
	if isAssignChannel {
		summary.Channel1 = assignAlias
		summary.ChannelClass = assignClass
	} else {
		sort.Strings(aliases)
		sort.Strings(classes)
		summary.Channel1 = strings.Join(aliases, ",")
		summary.ChannelClass = strings.Join(classes, ",")
	}
	list = append(list, summary)

	// 表格：倒序显示
	count := len(list)
	for i := 0; i < count/2; i++ {
		list[i], list[count-1-i] = list[count-1-i], list[i]
	}
	return
}

// 用户转化效率表
func (this *statisticsService) UserTransEffectSummary(list []*entity.UserTransEffect, where0 string, args0 []any) (summary *entity.UserTransEffect, err error) {
	summary = &entity.UserTransEffect{
		SDate: "总汇",
	}
	for _, item := range list {
		summary.RegistCount += item.RegistCount
		summary.ActiveCount += item.ActiveCount
		summary.C0_1 += item.C0_1
		summary.C0_3 += item.C0_3
		summary.C0_7 += item.C0_7
		summary.C0_10 += item.C0_10
		summary.C0_15 += item.C0_15
		summary.C0_20 += item.C0_20
		summary.C0_25 += item.C0_25
		summary.C0_30 += item.C0_30
		summary.C2_1 += item.C2_1
		summary.C2_3 += item.C2_3
		summary.C2_7 += item.C2_7
		summary.C2_10 += item.C2_10
		summary.C2_15 += item.C2_15
		summary.C2_20 += item.C2_20
		summary.C2_25 += item.C2_25
		summary.C2_30 += item.C2_30
		summary.C3_1 += item.C3_1
		summary.C3_3 += item.C3_3
		summary.C3_7 += item.C3_7
		summary.C3_10 += item.C3_10
		summary.C3_15 += item.C3_15
		summary.C3_20 += item.C3_20
		summary.C3_25 += item.C3_25
		summary.C3_30 += item.C3_30
		summary.C4_1 += item.C4_1
		summary.C4_3 += item.C4_3
		summary.C4_7 += item.C4_7
		summary.C4_10 += item.C4_10
		summary.C4_15 += item.C4_15
		summary.C4_20 += item.C4_20
		summary.C4_25 += item.C4_25
		summary.C4_30 += item.C4_30
		summary.C5_1 += item.C5_1
		summary.C5_3 += item.C5_3
		summary.C5_7 += item.C5_7
		summary.C5_10 += item.C5_10
		summary.C5_15 += item.C5_15
		summary.C5_20 += item.C5_20
		summary.C5_25 += item.C5_25
		summary.C5_30 += item.C5_30
		summary.C6_1 += item.C6_1
		summary.C6_3 += item.C6_3
		summary.C6_7 += item.C6_7
		summary.C6_10 += item.C6_10
		summary.C6_15 += item.C6_15
		summary.C6_20 += item.C6_20
		summary.C6_25 += item.C6_25
		summary.C6_30 += item.C6_30
	}

	if summary.RegistCount > 0 {
		summary.C0Rate_1 = fmt.Sprintf("%.2f%%", float64(summary.C0_1)/float64(summary.RegistCount)*100)
		summary.C0Rate_3 = fmt.Sprintf("%.2f%%", float64(summary.C0_3)/float64(summary.RegistCount)*100)
		summary.C0Rate_7 = fmt.Sprintf("%.2f%%", float64(summary.C0_7)/float64(summary.RegistCount)*100)
		summary.C0Rate_10 = fmt.Sprintf("%.2f%%", float64(summary.C0_10)/float64(summary.RegistCount)*100)
		summary.C0Rate_15 = fmt.Sprintf("%.2f%%", float64(summary.C0_15)/float64(summary.RegistCount)*100)
		summary.C0Rate_20 = fmt.Sprintf("%.2f%%", float64(summary.C0_20)/float64(summary.RegistCount)*100)
		summary.C0Rate_25 = fmt.Sprintf("%.2f%%", float64(summary.C0_25)/float64(summary.RegistCount)*100)
		summary.C0Rate_30 = fmt.Sprintf("%.2f%%", float64(summary.C0_30)/float64(summary.RegistCount)*100)
		summary.C2Rate_1 = fmt.Sprintf("%.2f%%", float64(summary.C2_1)/float64(summary.RegistCount)*100)
		summary.C2Rate_3 = fmt.Sprintf("%.2f%%", float64(summary.C2_3)/float64(summary.RegistCount)*100)
		summary.C2Rate_7 = fmt.Sprintf("%.2f%%", float64(summary.C2_7)/float64(summary.RegistCount)*100)
		summary.C2Rate_10 = fmt.Sprintf("%.2f%%", float64(summary.C2_10)/float64(summary.RegistCount)*100)
		summary.C2Rate_15 = fmt.Sprintf("%.2f%%", float64(summary.C2_15)/float64(summary.RegistCount)*100)
		summary.C2Rate_20 = fmt.Sprintf("%.2f%%", float64(summary.C2_20)/float64(summary.RegistCount)*100)
		summary.C2Rate_25 = fmt.Sprintf("%.2f%%", float64(summary.C2_25)/float64(summary.RegistCount)*100)
		summary.C2Rate_30 = fmt.Sprintf("%.2f%%", float64(summary.C2_30)/float64(summary.RegistCount)*100)
		summary.C3Rate_1 = fmt.Sprintf("%.2f%%", float64(summary.C3_1)/float64(summary.RegistCount)*100)
		summary.C3Rate_3 = fmt.Sprintf("%.2f%%", float64(summary.C3_3)/float64(summary.RegistCount)*100)
		summary.C3Rate_7 = fmt.Sprintf("%.2f%%", float64(summary.C3_7)/float64(summary.RegistCount)*100)
		summary.C3Rate_10 = fmt.Sprintf("%.2f%%", float64(summary.C3_10)/float64(summary.RegistCount)*100)
		summary.C3Rate_15 = fmt.Sprintf("%.2f%%", float64(summary.C3_15)/float64(summary.RegistCount)*100)
		summary.C3Rate_20 = fmt.Sprintf("%.2f%%", float64(summary.C3_20)/float64(summary.RegistCount)*100)
		summary.C3Rate_25 = fmt.Sprintf("%.2f%%", float64(summary.C3_25)/float64(summary.RegistCount)*100)
		summary.C3Rate_30 = fmt.Sprintf("%.2f%%", float64(summary.C3_30)/float64(summary.RegistCount)*100)
		summary.C4Rate_1 = fmt.Sprintf("%.2f%%", float64(summary.C4_1)/float64(summary.RegistCount)*100)
		summary.C4Rate_3 = fmt.Sprintf("%.2f%%", float64(summary.C4_3)/float64(summary.RegistCount)*100)
		summary.C4Rate_7 = fmt.Sprintf("%.2f%%", float64(summary.C4_7)/float64(summary.RegistCount)*100)
		summary.C4Rate_10 = fmt.Sprintf("%.2f%%", float64(summary.C4_10)/float64(summary.RegistCount)*100)
		summary.C4Rate_15 = fmt.Sprintf("%.2f%%", float64(summary.C4_15)/float64(summary.RegistCount)*100)
		summary.C4Rate_20 = fmt.Sprintf("%.2f%%", float64(summary.C4_20)/float64(summary.RegistCount)*100)
		summary.C4Rate_25 = fmt.Sprintf("%.2f%%", float64(summary.C4_25)/float64(summary.RegistCount)*100)
		summary.C4Rate_30 = fmt.Sprintf("%.2f%%", float64(summary.C4_30)/float64(summary.RegistCount)*100)
		summary.C5Rate_1 = fmt.Sprintf("%.2f%%", float64(summary.C5_1)/float64(summary.RegistCount)*100)
		summary.C5Rate_3 = fmt.Sprintf("%.2f%%", float64(summary.C5_3)/float64(summary.RegistCount)*100)
		summary.C5Rate_7 = fmt.Sprintf("%.2f%%", float64(summary.C5_7)/float64(summary.RegistCount)*100)
		summary.C5Rate_10 = fmt.Sprintf("%.2f%%", float64(summary.C5_10)/float64(summary.RegistCount)*100)
		summary.C5Rate_15 = fmt.Sprintf("%.2f%%", float64(summary.C5_15)/float64(summary.RegistCount)*100)
		summary.C5Rate_20 = fmt.Sprintf("%.2f%%", float64(summary.C5_20)/float64(summary.RegistCount)*100)
		summary.C5Rate_25 = fmt.Sprintf("%.2f%%", float64(summary.C5_25)/float64(summary.RegistCount)*100)
		summary.C5Rate_30 = fmt.Sprintf("%.2f%%", float64(summary.C5_30)/float64(summary.RegistCount)*100)
		summary.C6Rate_1 = fmt.Sprintf("%.2f%%", float64(summary.C6_1)/float64(summary.RegistCount)*100)
		summary.C6Rate_3 = fmt.Sprintf("%.2f%%", float64(summary.C6_3)/float64(summary.RegistCount)*100)
		summary.C6Rate_7 = fmt.Sprintf("%.2f%%", float64(summary.C6_7)/float64(summary.RegistCount)*100)
		summary.C6Rate_10 = fmt.Sprintf("%.2f%%", float64(summary.C6_10)/float64(summary.RegistCount)*100)
		summary.C6Rate_15 = fmt.Sprintf("%.2f%%", float64(summary.C6_15)/float64(summary.RegistCount)*100)
		summary.C6Rate_20 = fmt.Sprintf("%.2f%%", float64(summary.C6_20)/float64(summary.RegistCount)*100)
		summary.C6Rate_25 = fmt.Sprintf("%.2f%%", float64(summary.C6_25)/float64(summary.RegistCount)*100)
		summary.C6Rate_30 = fmt.Sprintf("%.2f%%", float64(summary.C6_30)/float64(summary.RegistCount)*100)
	}

	// 日活总汇
	var data map[string]any
	err = ck.Select(&data, fmt.Sprintf(`
		SELECT count(*) count
		FROM (
			SELECT userid
			FROM game.col_log_login FINAL
			WHERE login_time between ? AND ?
			GROUP BY userid
		) t1
		JOIN game.col_user t2 FINAL ON t1.userid = t2.userid
		WHERE 1 = 1 %s
	`, where0), args0...)
	if err != nil {
		return
	}
	// 日活数据
	summary.ActiveCount = utils.ToInt64(data["count"])

	return
}

// 用户转化效率表
func (this *statisticsService) UserLevelStats(begin, end time.Time, packageIds []string,
	channelClasses, aliasChannels []string, dates []string, ctypes []int,
	statsCount, statsRate bool,
) (countStats map[string][]int64, rateStats map[string][]string, allChannelClasses []string, err error) {
	channelMap, err := ChannelService.GetChannelsMap()
	if err != nil {
		return
	}
	var clsChannels = make(map[string][]entity.ChannelInfo)
	for _, c := range channelMap {
		if c.ClassName != "" {
			if _, ok := clsChannels[c.ClassName]; !ok {
				allChannelClasses = append(allChannelClasses, c.ClassName)
			}
			clsChannels[c.ClassName] = append(clsChannels[c.ClassName], c)
		}
	}
	sort.Strings(allChannelClasses)

	// 如果选择了渠道和渠道类，那么就以渠道为准来画线条。
	// 如果只选择了渠道类，那么把渠道类下的各个渠道加总来画线。
	if len(channelClasses) > 0 && len(aliasChannels) > 0 {
		for _, cls := range channelClasses {
			for _, ch := range clsChannels[cls] {
				aliasChannels = append(aliasChannels, ch.Name1)
			}
		}
		channelClasses = nil
	}

	now := NowTime()
	nowYmd := uint32(now.Year()*10000) + uint32(now.Month()*100) + uint32(now.Day())
	countStats = make(map[string][]int64)
	rateStats = make(map[string][]string)
	var channelKeyFmt, classKeyFmt = "%s,%s,渠道-%s", "%s,%s,渠道类-%s"
	for _, date := range dates {
		ctime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", date), locationName)
		ymd := uint32(ctime.Year()*10000) + uint32(ctime.Month()*100) + uint32(ctime.Day())

		for _, ctype := range ctypes {
			for _, alias := range aliasChannels {
				channelkey := fmt.Sprintf(channelKeyFmt, date, ctype2string(ctype), alias)
				// if statsCount {
				countStats["人数图:"+channelkey] = make([]int64, utils.Min(30, nowYmd-ymd+1))
				// }
				if statsRate {
					rateStats["占比图:"+channelkey] = make([]string, utils.Min(30, nowYmd-ymd+1))
				}
			}
			for _, cls := range channelClasses {
				classKey := fmt.Sprintf(classKeyFmt, date, ctype2string(ctype), cls)
				// if statsCount {
				countStats["人数图:"+classKey] = make([]int64, utils.Min(30, nowYmd-ymd+1))
				// }
				if statsRate {
					rateStats["占比图:"+classKey] = make([]string, utils.Min(30, nowYmd-ymd+1))
				}
			}
		}
	}
	// 没有统计的
	if len(countStats) == 0 && len(rateStats) == 0 || (!statsCount && !statsRate) {
		return
	}

	// channelClassesMap := utils.Slice2MapKV(channelClasses, func(_ int, item string) (string, bool) { return item, true })
	// aliasChannelsMap := utils.Slice2MapKV(aliasChannels, func(_ int, item string) (string, bool) { return item, true })
	ctypesMap := utils.Slice2MapKV(ctypes, func(_ int, item int) (int, bool) { return item, true })

	// 要统计的日期
	var ymds []uint32
	for _, date := range dates {
		ctime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", date), locationName)
		ymd := uint32(ctime.Year()*10000) + uint32(ctime.Month()*100) + uint32(ctime.Day())
		ymds = append(ymds, ymd)
	}

	var where0 string
	var args0 = []any{ymds}
	if len(packageIds) > 0 {
		where0 = " and ad__bundle_id in ?"
		args0 = append(args0, packageIds)
	}

	var datas0 []map[string]any
	err = ck.Select(&datas0, fmt.Sprintf(`
		SELECT regist_date, ad__bundle_id, day, ctype_day, count(*) count
		FROM (
			SELECT t2.regist_date, t2.userid userid, t2.ad__bundle_id, t3.number day, SUM(t1.amount) pay_amount,
				SUM(CASE WHEN toYYYYMMDD(t1.ctime) - t2.regist_date BETWEEN 0 AND t3.number THEN t1.amount ELSE 0 END) pay_amount_day,
				(CASE WHEN pay_amount_day = 0 THEN 0 
				WHEN pay_amount_day BETWEEN 1 and 100000-1 THEN 2 
				WHEN pay_amount_day BETWEEN 100000 and 1000000-1 THEN 3
				WHEN pay_amount_day between 1000000 and 10000000-1 THEN 4
				WHEN pay_amount_day between 10000000 and 20000000-1 THEN 5
				WHEN pay_amount_day >= 20000000 THEN 6 ELSE -1 END) ctype_day
			FROM game.col_trade_record t1 FINAL
			RIGHT JOIN (
				SELECT userid, toYYYYMMDD(ctime) regist_date, ad__bundle_id
				FROM game.col_user FINAL 
				WHERE regist_date IN ? %s
			) t2 ON t1.userid = t2.userid AND t1.order_status = 4
			JOIN (
				SELECT number FROM numbers(0, 30, 1)
			) t3 ON 1 = 1
			GROUP BY t2.regist_date, t2.userid, t2.ad__bundle_id, t3.number
		) s1
		GROUP BY regist_date, ad__bundle_id, day, ctype_day
		ORDER BY regist_date, ad__bundle_id, day, ctype_day
	`, where0), args0...)
	if err != nil {
		return
	}

	// 每日注册人数
	var dateCounts = make(map[string]int64)
	var dateChannelCounts = make(map[string]map[string]int64)
	var dateClassCounts = make(map[string]map[string]int64)
	var datas1 []map[string]any
	err = ck.Select(&datas1, fmt.Sprintf(`
		SELECT toYYYYMMDD(ctime) regist_date, ad__bundle_id, count(*) count
		FROM game.col_user FINAL
		WHERE regist_date IN ? %s
		GROUP BY regist_date, ad__bundle_id
	`, where0), args0...)
	if err != nil {
		return
	}
	for _, data := range datas1 {
		regist_date := fmt.Sprint(data["regist_date"])
		sdate := regist_date[0:4] + "-" + regist_date[4:6] + "-" + regist_date[6:]
		count := utils.ToInt64(data["count"])
		channel := data["ad__bundle_id"].(string)

		dateCounts[sdate] += count
		channelCounts, ok := dateChannelCounts[sdate]
		if !ok {
			channelCounts = make(map[string]int64)
			dateChannelCounts[sdate] = channelCounts
		}
		channelCounts[channel] = count

		classCounts, ok := dateClassCounts[sdate]
		if !ok {
			classCounts = make(map[string]int64)
			dateClassCounts[sdate] = classCounts
		}
		if c, ok := channelMap[channel]; ok && c.ClassName != "" {
			classCounts[c.ClassName] += count
		}
	}

	for _, data := range datas0 {
		// regist_date := data["regist_date"].(uint32)
		regist_date := fmt.Sprint(data["regist_date"])
		sdate := regist_date[0:4] + "-" + regist_date[4:6] + "-" + regist_date[6:]
		channel := data["ad__bundle_id"].(string)
		ctype_day := int(utils.ToInt64(data["ctype_day"]))
		day := utils.ToInt64(data["day"])
		count := utils.ToInt64(data["count"])

		if ok := ctypesMap[ctype_day]; !ok {
			continue
		}
		ctypeName := ctype2string(ctype_day)

		// 渠道
		c, ok := channelMap[channel]
		if !ok {
			continue
		}
		// skipChannel := !ok || c.Name1 == "" || !aliasChannelsMap[c.Name1]               // 渠道别名统计跳过
		// skipChannelClass := !ok || c.ClassName == "" || !channelClassesMap[c.ClassName] // 渠道类统计跳过
		// if skipChannel && skipChannelClass {
		// 	continue
		// }

		aliasKey := fmt.Sprintf(channelKeyFmt, sdate, ctypeName, c.Name1)
		classKey := fmt.Sprintf(classKeyFmt, sdate, ctypeName, c.ClassName)
		// 渠道别名统计
		// if statsCount {
		days, ok := countStats["人数图:"+aliasKey]
		if ok && int(day) < len(days) {
			days[day] = count
		}
		// 渠道类统计
		days, ok = countStats["人数图:"+classKey]
		if ok && int(day) < len(days) {
			days[day] += count
		}
		// }
		if statsRate {
			// 渠道占比图
			var channelCount int64
			if cc, ok := dateChannelCounts[sdate]; ok {
				channelCount = cc[channel]
			}
			rate := fmt.Sprintf("%.2f", ComputeFloat(count, channelCount)*100)
			days, ok := rateStats["占比图:"+aliasKey]
			if ok && int(day) < len(days) {
				days[day] = rate
			}
		}
	}

	if statsRate {
		// 渠道类占比图
		for key, rates := range rateStats {
			if strings.Contains(key, "渠道类-") {
				sdate := string([]rune(key)[4:14])
				clsKey2 := strings.ReplaceAll(key, "渠道类-", "cls:")
				cls := string(clsKey2[strings.LastIndex(clsKey2, "cls:")+4:])

				countKey := strings.ReplaceAll(key, "占比图:", "人数图:")
				counts, ok := countStats[countKey]
				if !ok {
					continue
				}

				var classCount int64
				if cc, ok := dateClassCounts[sdate]; ok {
					classCount = cc[cls]
				}

				for i := 0; i < len(rates); i++ {
					rates[i] = fmt.Sprintf("%.2f", ComputeFloat(counts[i], classCount)*100)
				}
			}
		}
	}

	if !statsCount {
		countStats = make(map[string][]int64)
	}
	return
}

func ctype2string(ctype int) string {
	switch ctype {
	case 0:
		return "零充"
	case 1:
		return "平民"
	case 2:
		return "普充"
	case 3:
		return "小R"
	case 4:
		return "中R"
	case 5:
		return "大R"
	case 6:
		return "超大R"
	default:
		return fmt.Sprintf("未知-%d", ctype)
	}
}

// 用户分层漏斗图
func (this *statisticsService) UserLevelFunnelStats(begin, end time.Time, packageIds []string) (list []*entity.UserLevelFunnelStats, err error) {
	var where0 string
	var args0 = []any{begin, end}
	if len(packageIds) > 0 {
		where0 = " and ad__bundle_id in ?"
		args0 = append(args0, packageIds)
	}

	var datas0 []map[string]any
	err = ck.Select(&datas0, fmt.Sprintf(`
		SELECT toYYYYMMDD(ctime) regist_date, count(*) count, ad__bundle_id,
		(CASE WHEN money = 0 THEN 0
			WHEN money BETWEEN 1 and 100000-1 THEN 2 
			WHEN money BETWEEN 100000 and 1000000-1 THEN 3
			WHEN money between 1000000 and 10000000-1 THEN 4
			WHEN money between 10000000 and 20000000-1 THEN 5
			WHEN money >= 20000000 THEN 6 ELSE -1 END
		) ctype
		FROM game.col_user FINAL 
		WHERE ctime BETWEEN ? AND ? %s
		GROUP BY regist_date, ctype, ad__bundle_id
	`, where0), args0...)
	if err != nil {
		return
	}

	channelMap, err := ChannelService.GetChannelsMap()
	if err != nil {
		return
	}

	// 每日注册人数
	var dateCounts = make(map[string]int64)
	var dateChannelCounts = make(map[string]map[string]int64)
	var datas1 []map[string]any
	err = ck.Select(&datas1, fmt.Sprintf(`
		SELECT toYYYYMMDD(ctime) regist_date, ad__bundle_id, count(*) count
		FROM game.col_user FINAL
		WHERE ctime BETWEEN ? AND ? %s
		GROUP BY regist_date, ad__bundle_id
	`, where0), args0...)
	if err != nil {
		return
	}
	for _, data := range datas1 {
		regist_date := fmt.Sprint(data["regist_date"])
		sdate := regist_date[0:4] + "-" + regist_date[4:6] + "-" + regist_date[6:]
		count := utils.ToInt64(data["count"])
		channel := data["ad__bundle_id"].(string)

		dateCounts[sdate] += count

		channelCounts, ok := dateChannelCounts[sdate]
		if !ok {
			channelCounts = make(map[string]int64)
			dateChannelCounts[sdate] = channelCounts
		}
		channelCounts[channel] = count
	}

	var summarys = make(map[string]*entity.UserLevelFunnelStat)         // date,stats
	var stats = make(map[string]map[string]*entity.UserLevelFunnelStat) // channel,date,stats
	// 汇总数据
	for _, data := range datas0 {
		regist_date := fmt.Sprint(data["regist_date"])
		sdate := regist_date[0:4] + "-" + regist_date[4:6] + "-" + regist_date[6:]
		ctype := utils.ToInt64(data["ctype"])
		ad__bundle_id := data["ad__bundle_id"].(string)
		count := utils.ToInt64(data["count"])
		channel, ok := channelMap[ad__bundle_id]
		if !ok {
			fmt.Printf("channel not exists: %s\n", ad__bundle_id)
			continue
		}
		stat, ok := stats[channel.Name]
		if !ok {
			stat = make(map[string]*entity.UserLevelFunnelStat)
			stats[channel.Name] = stat
		}
		info, ok := stat[sdate]
		if !ok {
			info = &entity.UserLevelFunnelStat{
				SDate:        sdate,
				Channel:      channel.Name,
				Channel1:     utils.CaseElse(channel.Name1 != "", channel.Name1, channel.Name),
				ChannelClass: channel.ClassName,
			}
			stat[sdate] = info
		}
		summary, ok := summarys[sdate]
		if !ok {
			summary = &entity.UserLevelFunnelStat{
				SDate: sdate,
			}
			summarys[sdate] = summary
		}

		// 当日该渠道注册人数
		var dateChannelCount int64
		if cc, ok := dateChannelCounts[sdate]; ok {
			dateChannelCount = cc[channel.Name]
		}
		rate := fmt.Sprintf("%.2f", ComputeFloat(count, dateChannelCount)*100)
		switch ctype {
		case 0:
			info.C0 = count
			info.C0Rate = rate
			summary.C0 += count
		case 2:
			info.C2 = count
			info.C2Rate = rate
			summary.C2 += count
		case 3:
			info.C3 = count
			info.C3Rate = rate
			summary.C3 += count
		case 4:
			info.C4 = count
			info.C4Rate = rate
			summary.C4 += count
		case 5:
			info.C5 = count
			info.C5Rate = rate
			summary.C5 += count
		case 6:
			info.C6 = count
			info.C6Rate = rate
			summary.C6 += count
		}
	}

	// list = make([]*entity.UserLevelFunnelStats, 0, len(stats)+1)

	suammryStats := &entity.UserLevelFunnelStats{
		Channel1: "渠道加总",
	}
	stime := begin
	for !stime.After(end) {
		sdate := stime.Format("2006-01-02")
		stime = stime.AddDate(0, 0, 1)
		summary, ok := summarys[sdate]
		if ok {
			summary.C0Rate = fmt.Sprintf("%.2f", ComputeFloat(summary.C0, dateCounts[sdate])*100)
			summary.C2Rate = fmt.Sprintf("%.2f", ComputeFloat(summary.C2, dateCounts[sdate])*100)
			summary.C3Rate = fmt.Sprintf("%.2f", ComputeFloat(summary.C3, dateCounts[sdate])*100)
			summary.C4Rate = fmt.Sprintf("%.2f", ComputeFloat(summary.C4, dateCounts[sdate])*100)
			summary.C5Rate = fmt.Sprintf("%.2f", ComputeFloat(summary.C5, dateCounts[sdate])*100)
			summary.C6Rate = fmt.Sprintf("%.2f", ComputeFloat(summary.C6, dateCounts[sdate])*100)
		} else {
			summary = &entity.UserLevelFunnelStat{
				SDate: sdate,
			}
		}
		suammryStats.Stats = append(suammryStats.Stats, summary)
	}
	list = append(list, suammryStats)

	var channels []string
	for ch := range stats {
		channels = append(channels, ch)
	}
	sort.Strings(channels) // 渠道排一下序
	for _, ch := range channels {
		dates := stats[ch]
		stime := begin
		channel := channelMap[ch]

		stats := &entity.UserLevelFunnelStats{
			Channel:      channel.Name,
			Channel1:     utils.CaseElse(channel.Name1 != "", channel.Name1, channel.Name),
			ChannelClass: channel.ClassName,
		}
		for !stime.After(end) {
			sdate := stime.Format("2006-01-02")
			stime = stime.AddDate(0, 0, 1)
			stat, ok := dates[sdate]
			if !ok {
				stat = &entity.UserLevelFunnelStat{
					SDate: sdate,
				}
			}
			stats.Stats = append(stats.Stats, stat)
		}
		list = append(list, stats)
	}
	return
}

// 进量分布图
func (this *statisticsService) UserRegChannelStats(begin, end time.Time, packageIds []string) (list []*entity.UserRegChannelStat, err error) {
	var where0 string
	var where0_args []any
	if len(packageIds) > 0 {
		where0 = " and ad__bundle_id in ?"
		where0_args = append(where0_args, packageIds)
	}

	var datas []map[string]any
	var args = []any{begin, end}
	args = append(args, where0_args...)
	err = ck.Select(&datas, fmt.Sprintf(`
		select toYYYYMMDD(ctime) reg_date, ad__bundle_id, count(*) reg_users
		from game.col_user final
		where ctime between ? and ? and robot = 0 %s
		group by reg_date, ad__bundle_id
		order by reg_date desc, ad__bundle_id
	`, where0), args...)
	if err != nil {
		return
	}

	channelMap, err := ChannelService.GetChannelsMap()
	if err != nil {
		return
	}

	var dateStats = make(map[uint32]*entity.UserRegChannelStat)

	for _, data := range datas {
		reg_date0 := data["reg_date"].(uint32)
		reg_date := fmt.Sprint(reg_date0)
		sdate := reg_date[0:4] + "-" + reg_date[4:6] + "-" + reg_date[6:]
		ad__bundle_id := data["ad__bundle_id"].(string)
		reg_users := utils.ToInt64(data["reg_users"])

		stat, ok := dateStats[reg_date0]
		if !ok {
			stat = &entity.UserRegChannelStat{
				SDate: sdate,
			}
			list = append(list, stat)
			dateStats[reg_date0] = stat
		}
		stat.RegUsers += reg_users

		channel, ok := channelMap[ad__bundle_id]
		if !ok {
			channel = entity.ChannelInfo{
				Name:      ad__bundle_id,
				Name1:     ad__bundle_id,
				ClassName: ad__bundle_id,
			}
		}
		stat.Channels = append(stat.Channels, &entity.UserRegsChannel{
			Channel:      channel.Name,
			Channel1:     utils.CaseElse(channel.Name1 != "", channel.Name1, channel.Name),
			ChannelClass: utils.CaseElse(channel.ClassName != "", channel.ClassName, "-"),
			RegUsers:     reg_users,
		})
	}
	return
}
