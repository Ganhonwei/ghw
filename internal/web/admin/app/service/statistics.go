package service

import (
	"bytes"
	"errors"
	"fmt"
	"goserver/internal/web/admin/app/entity"
	"goserver/pkg/data/ck"
	"goserver/pkg/utils"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/globalsign/mgo/bson"
)

type statisticsService struct{}

// 查询数据汇总信息
func (this *statisticsService) GetList(page, pageSize int, m bson.M) ([]entity.DataSummary, error) {
	var list []entity.DataSummary
	if pageSize == -1 {
		pageSize = 100000
	}
	// skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "date", false)
	err := DataSummarys.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList(list)
	return list, err
}

// 转换为分展示
func (this *statisticsService) chipList(list []entity.DataSummary) []entity.DataSummary {
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		// v.FNewRechargeAmount = Chip2Float(v.NewRechargeAmount)
		// v.FTotalAmount = Chip2Float(v.TotalAmount)
		// v.FTotalWithdrawAmount = Chip2Float(v.TotalWithdrawAmount)
		list[k] = v
	}
	return list
}

// 查询数据汇总条数
func (this *statisticsService) GetTotal(m bson.M) (int64, error) {
	return int64(Count(DataSummarys, m)), nil
}

// 数据汇总
func (this *statisticsService) AddDataSummary(summary *entity.DataSummary) error {
	info := new(entity.DataSummary)
	GetByQ(DataSummarys, bson.M{"date": summary.Date}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"new_register":          summary.NewRegister,
			"new_equipment":         summary.NewEquipment,
			"login_num":             summary.LoginNum,
			"next_day_retention":    summary.NextDayRetention,
			"pay_retention":         summary.PayRetention,
			"new_recharge_num":      summary.NewRechargeNum,
			"new_recharge_amount":   summary.NewRechargeAmount,
			"new_recharge_arpu":     summary.NewRechargeArpu,
			"new_recharge_arppu":    summary.NewRechargeArppu,
			"new_user_payment_rate": summary.NewUserPaymentRate,
			"total_recharge":        summary.TotalRecharge,
			"total_amount":          summary.TotalAmount,
			"total_recharge_arpu":   summary.TotalRechargeArpu,
			"total_recharge_arppu":  summary.TotalRechargeArppu,
			"total_payment_rate":    summary.TotalPaymentRate,
			"total_withdraw_num":    summary.TotalWithdrawNum,
			"total_withdraw_amount": summary.TotalWithdrawAmount,
			"handling_charge":       summary.HandlingCharge,
			"cost_ratio":            summary.CostRatio,
		}
		if Update(DataSummarys, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + summary.Id)
	} else {
		// 新增
		summary.Id = bson.NewObjectId().Hex()
		if !Insert(DataSummarys, summary) {
			return errors.New("写入失败:" + summary.Id)
		}
		return nil
	}
}

// 用户留存 实时查询
func (this *statisticsService) GetRetainedListCK(begin, end time.Time, packageIds []string) (ret []*entity.UserRetainedOld, err error) {
	if len(packageIds) == 0 {
		return this.GetRetainedListCK0(begin, end)
	}
	return this.GetRetainedListCK1(begin, end, packageIds)
}

// 用户留存 实时查询
func (this *statisticsService) GetRetainedListCK0(begin, end time.Time) (ret []*entity.UserRetainedOld, err error) {
	sql0 := `
		SELECT (login_day - ?) diff_day, count(*) c FROM (
			SELECT userid, toYYYYMMDD(login_time) login_day
			FROM game.col_log_login FINAL WHERE userid IN (
				SELECT userid FROM game.col_user FINAL 
				WHERE ctime >= ? AND ctime < ?
			)
			GROUP BY userid, login_day
		) t1
		GROUP BY diff_day
		HAVING diff_day IN (0,1,2,3,4,5,6,7,15,30,40,50,60)
	`

	for begin.Before(end) {
		stime := begin
		etime := begin.AddDate(0, 0, 1)
		begin = begin.AddDate(0, 0, 1)
		ymd := uint32(stime.Year()*10000) + uint32(stime.Month()*100) + uint32(stime.Day())
		item := &entity.UserRetainedOld{
			SDate:   stime,
			Channel: "全部",
		}
		ret = append(ret, item)

		var datas []map[string]any
		args := []any{ymd, stime, etime}
		// fmt.Printf("========= stime=%v, etime=%v\n", stime, etime)
		err = ck.Select(&datas, sql0, args...)
		if err != nil {
			return
		}
		for _, data := range datas {
			diff_day := utils.ToInt64(data["diff_day"])
			c := utils.ToInt64(data["c"])
			// 0,1,2,3,4,5,6,7,15,30,60
			switch diff_day {
			case 0:
				item.NewNumber = c
			case 1:
				item.Login1 = c
			case 2:
				item.Login2 = c
			case 3:
				item.Login3 = c
			case 4:
				item.Login4 = c
			case 5:
				item.Login5 = c
			case 6:
				item.Login6 = c
			case 7:
				item.Login7 = c
			case 15:
				item.Login15 = c
			case 30:
				item.Login30 = c
			case 40:
				item.Login40 = c
			case 50:
				item.Login50 = c
			case 60:
				item.Login60 = c
			}
		}
	}
	this.chipList1(ret)
	return
}

func (this *statisticsService) GetRetainedListCK1(begin, end time.Time, packageIds []string) (ret []*entity.UserRetainedOld, err error) {
	sql0 := `
		SELECT ad__bundle_id, (login_day - ?) diff_day, count(*) c
		FROM (
			SELECT t1.userid, t1.login_day, t2.ad__bundle_id 
			FROM (
				SELECT userid, toYYYYMMDD(login_time) login_day
				FROM game.col_log_login FINAL WHERE userid IN (
					SELECT userid FROM game.col_user FINAL 
					WHERE ctime >= ? AND ctime < ? and ad__bundle_id in ?
				)
				GROUP BY userid, login_day
			) t1
			JOIN game.col_user t2 FINAL ON t1.userid = t2.userid
		) s1
		GROUP BY diff_day, ad__bundle_id
		HAVING diff_day IN (0,1,2,3,4,5,6,7,15,30,40,50,60)
		ORDER BY diff_day, ad__bundle_id desc
	`

	var channelIdMap = make(map[string]bool)
	for begin.Before(end) {
		stime := begin
		etime := begin.AddDate(0, 0, 1)
		begin = begin.AddDate(0, 0, 1)
		ymd := uint32(stime.Year()*10000) + uint32(stime.Month()*100) + uint32(stime.Day())

		var datas []map[string]any
		args := []any{ymd, stime, etime, packageIds}
		err = ck.Select(&datas, sql0, args...)
		if err != nil {
			return
		}
		var channelItem = make(map[string]*entity.UserRetainedOld)
		for _, data := range datas {
			ad__bundle_id := data["ad__bundle_id"].(string)
			diff_day := utils.ToInt64(data["diff_day"])
			c := utils.ToInt64(data["c"])
			item, ok := channelItem[ad__bundle_id]
			if !ok {
				channelIdMap[ad__bundle_id] = true
				item = &entity.UserRetainedOld{
					SDate:   stime,
					Channel: ad__bundle_id,
				}
				channelItem[ad__bundle_id] = item
				ret = append(ret, item)
			}
			// 0,1,2,3,4,5,6,7,15,30,60
			switch diff_day {
			case 0:
				item.NewNumber = c
			case 1:
				item.Login1 = c
			case 2:
				item.Login2 = c
			case 3:
				item.Login3 = c
			case 4:
				item.Login4 = c
			case 5:
				item.Login5 = c
			case 6:
				item.Login6 = c
			case 7:
				item.Login7 = c
			case 15:
				item.Login15 = c
			case 30:
				item.Login15 = c
			case 60:
				item.Login15 = c
			}
		}
	}
	this.chipList1(ret)

	// 渠道别名
	if len(channelIdMap) > 0 {
		var channelIds []string
		for channel := range channelIdMap {
			channelIds = append(channelIds, channel)
		}
		channels, err1 := ChannelService.GetChannelByIds(channelIds)
		if err1 != nil {
			err = err1
			return
		}
		var channelMap = make(map[string]entity.ChannelInfo)
		for _, ch := range channels {
			channelMap[ch.Name] = ch
		}
		for _, item := range ret {
			if channel, ok := channelMap[item.Channel]; ok {
				item.Channel1 = channel.Name1
			}
		}
	}
	return
}

// func (this *statisticsService) GetRetainedListCK1(begin, end time.Time, packageIds []string) (ret []*entity.UserRetained, err error) {
// 	sql0 := `
// 		SELECT (login_day - ?) diff_day, count(*) c FROM (
// 			SELECT userid, toYYYYMMDD(login_time) login_day
// 			FROM game.col_log_login FINAL WHERE userid IN (
// 				SELECT userid FROM game.col_user FINAL
// 				WHERE ctime >= ? AND ctime < ?
// 			)
// 			GROUP BY userid, login_day
// 		) t1
// 		GROUP BY diff_day
// 		HAVING diff_day IN (0,1,2,3,4,5,6,7,15,30,60)
// 	`
// 	// 渠道查询
// 	// sql1 := `
// 	// 	SELECT DISTINCT ad__bundle_id FROM game.col_user FINAL WHERE userid IN ?
// 	// `

// 	for begin.Before(end) {
// 		stime := begin
// 		etime := begin.AddDate(0, 0, 1)
// 		begin = begin.AddDate(0, 0, 1)
// 		ymd := uint32(stime.Year()*10000) + uint32(stime.Month()*100) + uint32(stime.Day())
// 		item := &entity.UserRetained{
// 			SDate:   stime,
// 			Channel: "全部",
// 		}
// 		ret = append(ret, item)

// 		var datas []map[string]any
// 		args := []any{ymd, stime, etime}
// 		if len(packageIds) > 0 {
// 			args = append(args, packageIds)
// 			sql0 = fmt.Sprintf(sql0, ", groupArray(t1.userid) userids", " and ad__bundle_id in ?")
// 		} else {
// 			sql0 = fmt.Sprintf(sql0, ", [''] userids", "")
// 		}
// 		err = ck.Select(&datas, sql0, args...)
// 		if err != nil {
// 			return
// 		}
// 		var dayUserids []string
// 		for _, data := range datas {
// 			diff_day := utils.ToInt64(data["diff_day"])
// 			c := utils.ToInt64(data["c"])
// 			userids := data["userids"].([]string)
// 			// 0,1,2,3,4,5,6,7,15,30,60
// 			switch diff_day {
// 			case 0:
// 				item.NewNumber = c
// 				dayUserids = userids
// 			case 1:
// 				item.Login1 = c
// 			case 2:
// 				item.Login2 = c
// 			case 3:
// 				item.Login3 = c
// 			case 4:
// 				item.Login4 = c
// 			case 5:
// 				item.Login5 = c
// 			case 6:
// 				item.Login6 = c
// 			case 7:
// 				item.Login7 = c
// 			case 15:
// 				item.Login15 = c
// 			case 30:
// 				item.Login15 = c
// 			case 60:
// 				item.Login15 = c
// 			}
// 		}

// 		if len(packageIds) > 0 && len(dayUserids) > 0 {
// 			var channels []map[string]any
// 			err = ck.Select(&channels, sql1, dayUserids)
// 			if err != nil {
// 				return
// 			}
// 			var channelIds []string
// 			for _, ch := range channels {
// 				bundle_id := ch["ad__bundle_id"].(string)
// 				channelIds = append(channelIds, bundle_id)
// 			}
// 			item.Channel = strings.Join(channelIds, ",")
// 			if len(channelIds) > 0 {
// 				var aliasIds []string
// 				channels, err1 := ChannelService.GetChannelByIds(channelIds)
// 				if err1 != nil {
// 					err = err1
// 					return
// 				}
// 				for _, ch := range channels {
// 					aliasIds = append(aliasIds, ch.Name1)
// 				}
// 				item.Channel1 = strings.Join(aliasIds, ",")
// 			}
// 		}
// 	}
// 	this.chipList1(ret)
// 	return
// }

/*
	用户留存
*/
// Deprecated: 查询用户留存信息
func (this *statisticsService) GetRetainedList(page, pageSize int, m bson.M, isChannel bool) ([]*entity.UserRetainedOld, error) {
	var list []*entity.UserRetainedOld
	if pageSize == -1 {
		pageSize = 100000
	}
	// skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)
	pipeline := []bson.M{}
	if isChannel {
		// 渠道
		pipeline = []bson.M{
			{"$match": m},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":     "$date",
						"channel":  "$channel",
						"channel1": "$channel1",
					},
					"new_number": bson.M{"$sum": "$new_number"},
					"login1":     bson.M{"$sum": "$login1"},
					"register1":  bson.M{"$sum": "$register1"},
					"login2":     bson.M{"$sum": "$login2"},
					"register2":  bson.M{"$sum": "$register2"},
					"login3":     bson.M{"$sum": "$login3"},
					"register3":  bson.M{"$sum": "$register3"},
					"login4":     bson.M{"$sum": "$login4"},
					"register4":  bson.M{"$sum": "$register4"},
					"login5":     bson.M{"$sum": "$login5"},
					"register5":  bson.M{"$sum": "$register5"},
					"login6":     bson.M{"$sum": "$login6"},
					"register6":  bson.M{"$sum": "$register6"},
					"login7":     bson.M{"$sum": "$login7"},
					"register7":  bson.M{"$sum": "$register7"},
					"login15":    bson.M{"$sum": "$login15"},
					"register15": bson.M{"$sum": "$register15"},
					"login30":    bson.M{"$sum": "$login30"},
					"register30": bson.M{"$sum": "$register30"},
					"login60":    bson.M{"$sum": "$login60"},
					"register60": bson.M{"$sum": "$register60"},
				},
			},
			{
				"$project": bson.M{
					"_id":        "$_id.date",
					"date":       "$_id.date",
					"channel":    "$_id.channel",
					"channel1":   "$_id.channel1",
					"new_number": "$new_number",
					"login1":     "$login1",
					"register1":  "$register1",
					"login2":     "$login2",
					"register2":  "$register2",
					"login3":     "$login3",
					"register3":  "$register3",
					"login4":     "$login4",
					"register4":  "$register4",
					"login5":     "$login5",
					"register5":  "$register5",
					"login6":     "$login6",
					"register6":  "$register6",
					"login7":     "$login7",
					"register7":  "$register7",
					"login15":    "$login15",
					"register15": "$register15",
					"login30":    "$login30",
					"register30": "$register30",
					"login60":    "$login60",
					"register60": "$register60",
				},
			},
			{"$sort": bson.M{"date": -1}},
			{"$skip": skipNum},
			{"$limit": pageSize},
		}
	} else {
		// 全部
		pipeline = []bson.M{
			{"$match": m},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date": "$date",
					},
					"new_number": bson.M{"$sum": "$new_number"},
					"login1":     bson.M{"$sum": "$login1"},
					"register1":  bson.M{"$sum": "$register1"},
					"login2":     bson.M{"$sum": "$login2"},
					"register2":  bson.M{"$sum": "$register2"},
					"login3":     bson.M{"$sum": "$login3"},
					"register3":  bson.M{"$sum": "$register3"},
					"login4":     bson.M{"$sum": "$login4"},
					"register4":  bson.M{"$sum": "$register4"},
					"login5":     bson.M{"$sum": "$login5"},
					"register5":  bson.M{"$sum": "$register5"},
					"login6":     bson.M{"$sum": "$login6"},
					"register6":  bson.M{"$sum": "$register6"},
					"login7":     bson.M{"$sum": "$login7"},
					"register7":  bson.M{"$sum": "$register7"},
					"login15":    bson.M{"$sum": "$login15"},
					"register15": bson.M{"$sum": "$register15"},
					"login30":    bson.M{"$sum": "$login30"},
					"register30": bson.M{"$sum": "$register30"},
					"login60":    bson.M{"$sum": "$login60"},
					"register60": bson.M{"$sum": "$register60"},
				},
			},
			{
				"$project": bson.M{
					"_id":        "$_id.date",
					"date":       "$_id.date",
					"new_number": "$new_number",
					"login1":     "$login1",
					"register1":  "$register1",
					"login2":     "$login2",
					"register2":  "$register2",
					"login3":     "$login3",
					"register3":  "$register3",
					"login4":     "$login4",
					"register4":  "$register4",
					"login5":     "$login5",
					"register5":  "$register5",
					"login6":     "$login6",
					"register6":  "$register6",
					"login7":     "$login7",
					"register7":  "$register7",
					"login15":    "$login15",
					"register15": "$register15",
					"login30":    "$login30",
					"register30": "$register30",
					"login60":    "$login60",
					"register60": "$register60",
				},
			},
			{"$sort": bson.M{"date": -1}},
			{"$skip": skipNum},
			{"$limit": pageSize},
		}
	}
	err := UserRetaineds.Pipe(pipeline).All(&list)
	list = this.chipList1(list)
	return list, err
}

// 转换为分展示
func (this *statisticsService) chipList1(list []*entity.UserRetainedOld) []*entity.UserRetainedOld {
	for k, v := range list {
		if v.Date > 0 {
			v.SDate = time.Unix(v.Date, 0)
		}
		// v.Day1 = ComputeFloat(v.Login1, v.Register1) * 100.0
		// v.Day2 = ComputeFloat(v.Login2, v.Register2) * 100.0
		// v.Day3 = ComputeFloat(v.Login3, v.Register3) * 100.0
		// v.Day4 = ComputeFloat(v.Login4, v.Register4) * 100.0
		// v.Day5 = ComputeFloat(v.Login5, v.Register5) * 100.0
		// v.Day6 = ComputeFloat(v.Login6, v.Register6) * 100.0
		// v.Day7 = ComputeFloat(v.Login7, v.Register7) * 100.0
		// v.Day15 = ComputeFloat(v.Login15, v.Register15) * 100.0
		// v.Day30 = ComputeFloat(v.Login30, v.Register30) * 100.0
		// v.Day60 = ComputeFloat(v.Login60, v.Register60) * 100.0
		v.Day1 = ComputeFloat(v.Login1, v.NewNumber) * 100.0
		v.Day2 = ComputeFloat(v.Login2, v.NewNumber) * 100.0
		v.Day3 = ComputeFloat(v.Login3, v.NewNumber) * 100.0
		v.Day4 = ComputeFloat(v.Login4, v.NewNumber) * 100.0
		v.Day5 = ComputeFloat(v.Login5, v.NewNumber) * 100.0
		v.Day6 = ComputeFloat(v.Login6, v.NewNumber) * 100.0
		v.Day7 = ComputeFloat(v.Login7, v.NewNumber) * 100.0
		v.Day15 = ComputeFloat(v.Login15, v.NewNumber) * 100.0
		v.Day30 = ComputeFloat(v.Login30, v.NewNumber) * 100.0
		v.Day40 = ComputeFloat(v.Login40, v.NewNumber) * 100.0
		v.Day50 = ComputeFloat(v.Login50, v.NewNumber) * 100.0
		v.Day60 = ComputeFloat(v.Login60, v.NewNumber) * 100.0
		list[k] = v
	}
	return list
}

// 查询用户留存条数
func (this *statisticsService) GetRetainedTotal(m bson.M, isChannel bool) (int64, error) {
	pipeline := []bson.M{}
	if isChannel {
		pipeline = []bson.M{
			{
				"$match": m,
			},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":     "$date",
						"channel":  "$channel",
						"channel1": "$channel1",
					},
				},
			},
		}
	} else {
		pipeline = []bson.M{
			{
				"$match": m,
			},
			{
				"$group": bson.M{
					"_id": "$date",
				},
			},
		}
	}
	result := []bson.M{}
	pipe := UserRetaineds.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("GetRetainedTotal fail err: ", err)
	}
	return int64(len(result)), nil
}

// 新增或更新用户留存数据
func (this *statisticsService) AddUserRetained(user *entity.UserRetainedOld) error {
	info := new(entity.UserRetainedOld)
	if user.SType == 0 {
		GetByQ(UserRetaineds, bson.M{"date": user.Date, "s_type": user.SType}, info)
	} else {
		GetByQ(UserRetaineds, bson.M{"date": user.Date, "s_type": user.SType, "channel": user.Channel}, info)
	}
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"channel1":   user.Channel1,
			"new_number": user.NewNumber,
			"s_type":     user.SType,
			"day1":       user.Day1,
			"day2":       user.Day2,
			"day3":       user.Day3,
			"day4":       user.Day4,
			"day5":       user.Day5,
			"day6":       user.Day6,
			"day7":       user.Day7,
			"day15":      user.Day15,
			"day30":      user.Day30,
			"day60":      user.Day60,
		}
		if Update(UserRetaineds, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + user.Id)
	} else {
		// 新增
		user.Id = bson.NewObjectId().Hex()
		if !Insert(UserRetaineds, user) {
			return errors.New("写入失败:" + user.Id)
		}
		return nil
	}
}

/*
	付费用户留存
*/
// 分页查询付费用户留存
func (this *statisticsService) GetPayUserRetainedList(page, pageSize int, m bson.M, isChannel bool) ([]entity.PayUserRetained, error) {
	var list []entity.PayUserRetained
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)
	pipeline := []bson.M{}
	if isChannel {
		// 渠道
		pipeline = []bson.M{
			{"$match": m},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":     "$date",
						"channel":  "$channel",
						"channel1": "$channel1",
					},
					"new_number": bson.M{"$sum": "$new_number"},
					"pay1":       bson.M{"$sum": "$pay1"},
					"register1":  bson.M{"$sum": "$register1"},
					"pay2":       bson.M{"$sum": "$pay2"},
					"register2":  bson.M{"$sum": "$register2"},
					"pay3":       bson.M{"$sum": "$pay3"},
					"register3":  bson.M{"$sum": "$register3"},
					"pay4":       bson.M{"$sum": "$pay4"},
					"register4":  bson.M{"$sum": "$register4"},
					"pay5":       bson.M{"$sum": "$pay5"},
					"register5":  bson.M{"$sum": "$register5"},
					"pay6":       bson.M{"$sum": "$pay6"},
					"register6":  bson.M{"$sum": "$register6"},
					"pay7":       bson.M{"$sum": "$pay7"},
					"register7":  bson.M{"$sum": "$register7"},
					"pay15":      bson.M{"$sum": "$pay15"},
					"register15": bson.M{"$sum": "$register15"},
					"pay30":      bson.M{"$sum": "$pay30"},
					"register30": bson.M{"$sum": "$register30"},
					"pay60":      bson.M{"$sum": "$pay60"},
					"register60": bson.M{"$sum": "$register60"},
				},
			},
			{
				"$project": bson.M{
					"_id":        "$_id.date",
					"date":       "$_id.date",
					"channel":    "$_id.channel",
					"channel1":   "$_id.channel1",
					"new_number": "$new_number",
					"pay1":       "$pay1",
					"register1":  "$register1",
					"pay2":       "$pay2",
					"register2":  "$register2",
					"pay3":       "$pay3",
					"register3":  "$register3",
					"pay4":       "$pay4",
					"register4":  "$register4",
					"pay5":       "$pay5",
					"register5":  "$register5",
					"pay6":       "$pay6",
					"register6":  "$register6",
					"pay7":       "$pay7",
					"register7":  "$register7",
					"pay15":      "$pay15",
					"register15": "$register15",
					"pay30":      "$pay30",
					"register30": "$register30",
					"pay60":      "$pay60",
					"register60": "$register60",
				},
			},
			{"$sort": bson.M{"date": -1}},
			{"$skip": skipNum},
			{"$limit": pageSize},
		}
	} else {
		// 全部
		pipeline = []bson.M{
			{"$match": m},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date": "$date",
					},
					"new_number": bson.M{"$sum": "$new_number"},
					"pay1":       bson.M{"$sum": "$pay1"},
					"register1":  bson.M{"$sum": "$register1"},
					"pay2":       bson.M{"$sum": "$pay2"},
					"register2":  bson.M{"$sum": "$register2"},
					"pay3":       bson.M{"$sum": "$pay3"},
					"register3":  bson.M{"$sum": "$register3"},
					"pay4":       bson.M{"$sum": "$pay4"},
					"register4":  bson.M{"$sum": "$register4"},
					"pay5":       bson.M{"$sum": "$pay5"},
					"register5":  bson.M{"$sum": "$register5"},
					"pay6":       bson.M{"$sum": "$pay6"},
					"register6":  bson.M{"$sum": "$register6"},
					"pay7":       bson.M{"$sum": "$pay7"},
					"register7":  bson.M{"$sum": "$register7"},
					"pay15":      bson.M{"$sum": "$pay15"},
					"register15": bson.M{"$sum": "$register15"},
					"pay30":      bson.M{"$sum": "$pay30"},
					"register30": bson.M{"$sum": "$register30"},
					"pay60":      bson.M{"$sum": "$pay60"},
					"register60": bson.M{"$sum": "$register60"},
				},
			},
			{
				"$project": bson.M{
					"_id":        "$_id.date",
					"date":       "$_id.date",
					"new_number": "$new_number",
					"pay1":       "$pay1",
					"register1":  "$register1",
					"pay2":       "$pay2",
					"register2":  "$register2",
					"pay3":       "$pay3",
					"register3":  "$register3",
					"pay4":       "$pay4",
					"register4":  "$register4",
					"pay5":       "$pay5",
					"register5":  "$register5",
					"pay6":       "$pay6",
					"register6":  "$register6",
					"pay7":       "$pay7",
					"register7":  "$register7",
					"pay15":      "$pay15",
					"register15": "$register15",
					"pay30":      "$pay30",
					"register30": "$register30",
					"pay60":      "$pay60",
					"register60": "$register60",
				},
			},
			{"$sort": bson.M{"date": -1}},
			{"$skip": skipNum},
			{"$limit": pageSize},
		}
	}
	err := PayUserRetaineds.Pipe(pipeline).All(&list)
	list = this.chipList17(list)
	// err := PayUserRetaineds.
	// 	Find(m).
	// 	Sort(sortFieldR).
	// 	Skip(skipNum).
	// 	Limit(pageSize).
	// 	All(&list)
	// list = this.chipList17(list)
	return list, err
}

// 转换为分展示
func (this *statisticsService) chipList17(list []entity.PayUserRetained) []entity.PayUserRetained {
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		v.Day1 = ComputeFloat(v.Pay1, v.NewNumber) * 100.0
		v.Day2 = ComputeFloat(v.Pay2, v.NewNumber) * 100.0
		v.Day3 = ComputeFloat(v.Pay3, v.NewNumber) * 100.0
		v.Day4 = ComputeFloat(v.Pay4, v.NewNumber) * 100.0
		v.Day5 = ComputeFloat(v.Pay5, v.NewNumber) * 100.0
		v.Day6 = ComputeFloat(v.Pay6, v.NewNumber) * 100.0
		v.Day7 = ComputeFloat(v.Pay7, v.NewNumber) * 100.0
		v.Day15 = ComputeFloat(v.Pay15, v.NewNumber) * 100.0
		v.Day30 = ComputeFloat(v.Pay30, v.NewNumber) * 100.0
		v.Day60 = ComputeFloat(v.Pay60, v.NewNumber) * 100.0
		list[k] = v
	}
	return list
}

// 查询付费用户留存条数
func (this *statisticsService) GetPayUserRetainedTotal(m bson.M, isChannel bool) (int64, error) {
	pipeline := []bson.M{}
	if isChannel {
		pipeline = []bson.M{
			{
				"$match": m,
			},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":     "$date",
						"channel":  "$channel",
						"channel1": "$channel1",
					},
				},
			},
		}
	} else {
		pipeline = []bson.M{
			{
				"$match": m,
			},
			{
				"$group": bson.M{
					"_id": "$date",
				},
			},
		}
	}
	result := []bson.M{}
	pipe := PayUserRetaineds.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("GetPayUserRetainedTotal fail err: ", err)
	}
	return int64(len(result)), nil
}

// 分页查询Ltv统计信息
func (this *statisticsService) LtvStatCK(stime, etime time.Time, registAreas []int32, packageIds []string, betLtv bool) (stats []*entity.LtvStat, err error) {
	where_u_sql := ""
	var where_u_args []any

	if len(registAreas) > 0 {
		where_u_sql += " and s0.regist_area in ?"
		where_u_args = append(where_u_args, registAreas)
	}
	if len(packageIds) > 0 {
		where_u_sql += " and s0.ad__bundle_id IN ?"
		where_u_args = append(where_u_args, packageIds)
	}

	dateStats := make(map[uint32]*entity.LtvStat)

	// 注册用户
	var reg_datas []map[string]any
	reg_args := append([]any{stime, etime}, where_u_args...)
	err = ck.Select(&reg_datas, fmt.Sprintf(`
		select t1.reg_date,
			count(distinct t1.userid) reg_users,
			count(distinct t1.ad__device_id) reg_ads,
			count(distinct (case when t0.userid != t1.userid and toYYYYMMDD(t0.ctime) < t1.reg_date then t1.ad__device_id else null end)) repeat_ads,
			(reg_ads - repeat_ads) ad_reg_users
		from game.col_user t0 final
		right join (
			select userid, toYYYYMMDD(s0.ctime) reg_date, s0.ad__device_id
			from game.col_user s0 final
			where s0.ctime between ? and ? and robot = 0 and simulation_robot = 0 %s
		) t1 on t0.ad__device_id = t1.ad__device_id
		group by t1.reg_date
		order by t1.reg_date desc
	`, where_u_sql), reg_args...)
	if err != nil {
		return
	}
	now := NowTime()
	today := now.Year()*10000 + int(now.Month())*100 + now.Day()
	for _, data := range reg_datas {
		reg_date := data["reg_date"].(uint32)
		reg_users := utils.ToInt64(data["reg_users"])
		ad_reg_users := utils.ToInt64(data["ad_reg_users"])

		datestr := fmt.Sprint(reg_date)
		sdate := datestr[0:4] + "-" + datestr[4:6] + "-" + datestr[6:]
		stat := &entity.LtvStat{
			SDate: sdate,
		}
		stats = append(stats, stat)
		dateStats[reg_date] = stat

		stat.RegUsers = reg_users
		stat.RegDevices = ad_reg_users

		var stime time.Time
		stime, err = time.ParseInLocation(utils.FORMAT_DATE, sdate, Location())
		if err != nil {
			return
		}

		for i := range 90 {
			ltvTime := stime.AddDate(0, 0, i)
			ltvDay := ltvTime.Year()*10000 + int(ltvTime.Month())*100 + ltvTime.Day()
			future := ltvDay >= today
			ltv := &entity.LtvStatDay{
				Future:                       future,
				Day:                          int64(i),
				PayAvg:                       "/",
				WithdrawAvg:                  "/",
				ProfitAvg:                    "/",
				ProfitAvgLtv1Rate:            "/",
				Pays:                         "/",
				Withdraws:                    "/",
				Profit:                       "/",
				ProfitLtv1Rate:               "/",
				PayUsersMap:                  make(map[string]bool),
				BetsAvg:                      "/",
				RebateAvg:                    "/",
				GameIncomeAvg:                "/",
				GameIncomeAvgLtv1Rate:        "/",
				Bets:                         "/",
				Rebates:                      "/",
				GameIncome:                   "/",
				GameIncomeLtv1Rate:           "/",
				GameIncomeSubGiveAvg:         "/",
				GameIncomeSubGiveAvgLtv1Rate: "/",
				GameIncomeSubGive:            "/",
				GameIncomeSubGiveLtv1Rate:    "/",
				NVProfitAvg:                  "/",
				NVProfitAvgLtv1Rate:          "/",
				NVProfit:                     "/",
				NVProfitLtv1Rate:             "/",
			}
			stat.Ltvs = append(stat.Ltvs, ltv)
		}
	}

	// 总充值单数, 复购
	var pay_datas []map[string]any
	err = ck.Select(&pay_datas, `
		select reg_date, count(*) pay_users, SUM(case when pay_orders0 >= 2 then 1 else 0 end) pay2_users, SUM(pay_orders0) pay_orders
		from (
			select toYYYYMMDD(s0.ctime) reg_date, s0.userid userid, count(*) pay_orders0
			from game.col_trade_record s1 final
			join game.col_user s0 final on s1.userid = s0.userid
			where s0.ctime between ? and ?
				and s1.ctime > ? and s1.order_status = 4
			group by reg_date, s0.userid
		) t0
		group by reg_date
		order by reg_date desc
	`, stime, etime, stime)
	if err != nil {
		return
	}
	for _, data := range pay_datas {
		reg_date := data["reg_date"].(uint32)
		pay_users := utils.ToInt64(data["pay_users"])
		pay2_users := utils.ToInt64(data["pay2_users"])
		pay_orders := utils.ToInt64(data["pay_orders"])
		if stat, ok := dateStats[reg_date]; ok {
			stat.PayUsers = pay_users
			stat.Pay2Users = pay2_users
			stat.PayOrders = pay_orders
			// 复购率
			stat.Pay2UsersRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.Pay2Users, stat.PayUsers)*100)
			// 人均付费统计: 付费成功订单数 / 付费人数
			stat.PayOrdersAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.PayOrders, stat.PayUsers))
		}
	}

	// 支付通道信息
	_, payChannels, err := PayService.PayChannelList(0)
	if err != nil {
		return
	}
	var payChannelMap = make(map[string]*entity.PayChannel)
	for _, c := range payChannels {
		payChannelMap[c.Id] = c
	}

	sql_ltv := `
		select reg_date, diff_days, channel_id, groupArray(pay_users) pay_users, SUM(pay_amounts) pay_amounts, 
			SUM(withdraw_amounts) withdraw_amounts, SUM(withdraw_orders) withdraw_orders
		from (
			select reg_date, diff_days, channel_id, groupArray(distinct userid) pay_users, SUM(amount) pay_amounts, 0 withdraw_amounts, 0 withdraw_orders
			from (
				select s0.userid userid, toString(s1.channel_id) channel_id, toYYYYMMDD(s0.ctime) reg_date, amount,
					dateDiff('day', toStartOfDay(s0.ctime), toStartOfDay(s1.ctime)) diff_days
				from game.col_trade_record s1 final
				join game.col_user s0 final on s1.userid = s0.userid
				where s0.ctime between ? and ? and s1.ctime > ? and s1.order_status = 4 %s
			) t0
			where diff_days < 90
			group by reg_date, diff_days, channel_id
			order by reg_date desc, diff_days
			
			union all
			
			select reg_date, diff_days, channel_id, [] pay_users, 0 pay_amounts, SUM(amount) withdraw_amounts, count(*) withdraw_orders
			from (
				select s0.userid userid, s1.out_channel channel_id, toYYYYMMDD(s0.ctime) reg_date, amount,
					dateDiff('day', toStartOfDay(s0.ctime), toStartOfDay(s1.ctime)) diff_days
				from game.col_withdraw_record s1 final
				join game.col_user s0 final on s1.userid = s0.userid
				where s0.ctime between ? and ? and s1.ctime > ? and s1.order_status = 2 %s
			) t0
			where diff_days < 90
			group by reg_date, diff_days, channel_id
		) t0
		group by reg_date, diff_days, channel_id
		order by reg_date desc, diff_days
	`
	var ltv_datas []map[string]any
	ltv_args := []any{stime, etime, stime}
	ltv_args = append(ltv_args, where_u_args...)
	ltv_args = append(ltv_args, stime, etime, stime)
	ltv_args = append(ltv_args, where_u_args...)
	err = ck.Select(&ltv_datas, fmt.Sprintf(sql_ltv, where_u_sql, where_u_sql), ltv_args...)
	if err != nil {
		return
	}
	for _, data := range ltv_datas {
		reg_date := data["reg_date"].(uint32)
		diff_days := utils.ToInt64(data["diff_days"])
		pay_users := data["pay_users"].([][]string)
		pay_amounts := utils.ToInt64(data["pay_amounts"])
		withdraw_amounts := utils.ToInt64(data["withdraw_amounts"])
		withdraw_orders := utils.ToInt64(data["withdraw_orders"])
		channel_id := data["channel_id"].(string)

		if stat, ok := dateStats[reg_date]; ok {
			for i, ltv := range stat.Ltvs {
				if int64(i) >= diff_days {
					ltv.Withdraws0 += withdraw_amounts
					ltv.Pays0 += pay_amounts

					for _, pay_users0 := range pay_users {
						for _, userid := range pay_users0 {
							ltv.PayUsersMap[userid] = true
						}
					}

					if c, ok := payChannelMap[channel_id]; ok {
						// 代收手续费
						ltv.PaysTax0 += int64(float64(pay_amounts) * c.PayRate / 100)

						// 代付手续费
						ltv.WithdrawsTax0 += int64(float64(withdraw_amounts) * c.WithdrawRate / 100)
						ltv.WithdrawsTax0 += (withdraw_orders * c.WithdrawFee)
					}
				}
			}
		}
	}

	if betLtv {
		sql_bet_ltv := `
			select toYYYYMMDD(s0.ctime) reg_date, dateDiff('day', toStartOfDay(s0.ctime), toStartOfDay(toDateTime(begin_time))) diff_days, sum(bet_amount) bets, sum(bet_amount+score) rebates
			from game.col_detail s1 final 
			join game.col_user s0 final	on s1.userid = s0.userid
			where s0.ctime between ? and ? and s1.begin_time > ? and s1.robot = 0 and s1.bet_amount != 0 %s
			group by reg_date, diff_days
			
			union all
			
			select t1.reg_date, t1.diff_days, sum(t1.amounts) bets, sum(t2.amounts) rebates 
			from (
				select toYYYYMMDD(s0.ctime) reg_date, dateDiff('day', toStartOfDay(s0.ctime), toStartOfDay(toDateTime(s1.ctime))) diff_days, sum(s1.amount) amounts
				from game.col_nsq_log_external_bet s1 final 
				join game.col_user s0 final	on s1.user_id = s0.userid
				where s0.ctime between ? and ? and s1.ctime > ? and s1.amount != 0 %s
				group by reg_date, diff_days
			) t1 left join (
				select toYYYYMMDD(s0.ctime) reg_date, dateDiff('day', toStartOfDay(s0.ctime), toStartOfDay(toDateTime(s1.ctime))) diff_days, sum(s1.amount) amounts
				from game.col_nsq_log_external_reward s1 final 
				join game.col_user s0 final	on s1.user_id = s0.userid
				where s0.ctime between ? and ? and s1.ctime > ? and s1.amount != 0 %s
				group by reg_date, diff_days
			) t2 on t1.reg_date = t2.reg_date and t1.diff_days = t2.diff_days
			where diff_days < 90
			group by t1.reg_date, t1.diff_days
		`
		var bet_ltv_datas []map[string]any
		bet_ltv_args := []any{stime, etime, stime.Unix()}
		bet_ltv_args = append(bet_ltv_args, where_u_args...)
		bet_ltv_args = append(bet_ltv_args, stime, etime, stime.Unix())
		bet_ltv_args = append(bet_ltv_args, where_u_args...)
		bet_ltv_args = append(bet_ltv_args, stime, etime, stime.Unix())
		bet_ltv_args = append(bet_ltv_args, where_u_args...)
		err = ck.Select(&bet_ltv_datas, fmt.Sprintf(sql_bet_ltv, where_u_sql, where_u_sql, where_u_sql), bet_ltv_args...)
		if err != nil {
			return
		}
		for _, data := range bet_ltv_datas {
			reg_date := data["reg_date"].(uint32)
			diff_days := utils.ToInt64(data["diff_days"])
			bets := utils.ToInt64(data["bets"])
			rebates := utils.ToInt64(data["rebates"])
			if stat, ok := dateStats[reg_date]; ok {
				for i, ltv := range stat.Ltvs {
					if int64(i) >= diff_days {
						ltv.Bets0 += bets
						ltv.Rebate0 += rebates
					}
				}
			}
		}

		// 发放统计
		var nonGiftLTypes []int32
		nonGiftLTypes = append(nonGiftLTypes, LTypeReason1...)
		nonGiftLTypes = append(nonGiftLTypes, LTypeReason2...)
		nonGiftLTypes = append(nonGiftLTypes, LTypeReason3...)
		nonGiftLTypes = append(nonGiftLTypes, LTypeReason4...)
		nonGiftLTypes = append(nonGiftLTypes, LTypeReason6...)
		nonGiftLTypes = append(nonGiftLTypes, LTypeReason7...)
		nonGiftLTypes = append(nonGiftLTypes, LTypeReason8...)
		give_cash_sql := `
			select sdate, diff_days, SUM(gift_cashs) gift_cashs from (
				select toYYYYMMDD(s0.ctime) sdate, dateDiff('day', toStartOfDay(s0.ctime), toStartOfDay(s1.ctime)) diff_days, 
					SUM(s1.add_diamond) gift_cashs
				from game.col_log_water s1 final
				join game.col_user s0 final on s0.userid = s1.userid
				where s0.ctime between ? AND ? and s1.ltype not in ? and s1.add_diamond > 0 %s
				GROUP BY sdate, diff_days
					union all
				select toYYYYMMDD(s0.ctime) sdate, dateDiff('day', toStartOfDay(s0.ctime), toStartOfDay(s1.ctime)) diff_days, 
					SUM(s1.score - s1.amount) gift_cashs
				from game.col_trade_record s1 final
				join game.col_user s0 final on s0.userid = s1.userid
				where s0.ctime between ? AND ? and s1.order_status = 4 and s1.shop_type = 11 %s
				GROUP BY sdate, diff_days
				ORDER BY sdate, diff_days
			) t0
			GROUP BY sdate, diff_days
			ORDER BY sdate, diff_days
		`
		var give_cash_datas []map[string]any
		var give_cash_args = []any{stime, etime, nonGiftLTypes}
		give_cash_args = append(give_cash_args, where_u_args...)
		give_cash_args = append(give_cash_args, stime, etime)
		give_cash_args = append(give_cash_args, where_u_args...)

		err = ck.Select(&give_cash_datas, fmt.Sprintf(give_cash_sql, where_u_sql, where_u_sql), give_cash_args...)
		if err != nil {
			return
		}
		for _, data := range give_cash_datas {
			sdate := data["sdate"].(uint32)
			diff_days := utils.ToInt64(data["diff_days"])
			gift_cashs := utils.ToInt64(data["gift_cashs"])
			if stat, ok := dateStats[sdate]; ok {
				for i, ltv := range stat.Ltvs {
					if int64(i) >= diff_days {
						ltv.GiveCash0 += gift_cashs
					}
				}
			}
		}
	}

	for _, stat := range stats {
		for _, ltv := range stat.Ltvs {
			if ltv.Future {
				break
			}
			ltv.PayUsers = int64(len(ltv.PayUsersMap))
			clear(ltv.PayUsersMap)
			ltv.PayAvg = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(ltv.Pays0, ltv.PayUsers)))
			ltv.WithdrawAvg = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(ltv.Withdraws0, ltv.PayUsers)))
			ltv.ProfitAvg0 = ComputeFloat(ltv.Pays0-ltv.Withdraws0, ltv.PayUsers)
			ltv.ProfitAvg = fmt.Sprintf("%.2f", Chip2Float(ltv.ProfitAvg0))
			ltv.NVProfitAvg0 = ComputeFloat(ltv.Pays0-ltv.Withdraws0-ltv.PaysTax0-ltv.WithdrawsTax0, ltv.PayUsers)
			ltv.NVProfitAvg = fmt.Sprintf("%.2f", Chip2Float(ltv.NVProfitAvg0))

			ltv.Pays = fmt.Sprintf("%.2f", Chip2Float(ltv.Pays0))
			ltv.Withdraws = fmt.Sprintf("%.2f", Chip2Float(ltv.Withdraws0))
			ltv.Profit0 = ltv.Pays0 - ltv.Withdraws0
			ltv.Profit = fmt.Sprintf("%.2f", Chip2Float(ltv.Profit0))
			ltv.NVProfit0 = ltv.Pays0 - ltv.Withdraws0 - ltv.PaysTax0 - ltv.WithdrawsTax0
			ltv.NVProfit = fmt.Sprintf("%.2f", Chip2Float(ltv.NVProfit0))

			ltv0 := stat.Ltvs[0]
			ltv.ProfitAvgLtv1Rate = fmt.Sprintf("%.2f%%", ComputeFloat(ltv.ProfitAvg0, ltv0.ProfitAvg0)*100)
			ltv.ProfitLtv1Rate = fmt.Sprintf("%.2f%%", ComputeFloat(ltv.Profit0, ltv0.Profit0)*100)
			ltv.NVProfitAvgLtv1Rate = fmt.Sprintf("%.2f%%", ComputeFloat(ltv.NVProfitAvg0, ltv0.NVProfitAvg0)*100)
			ltv.NVProfitLtv1Rate = fmt.Sprintf("%.2f%%", ComputeFloat(ltv.NVProfit0, ltv0.NVProfit0)*100)

			if betLtv {
				ltv.BetsAvg = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(ltv.Bets0, ltv.PayUsers)))
				ltv.RebateAvg = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(ltv.Rebate0, ltv.PayUsers)))
				ltv.GameIncomeAvg = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(ltv.Bets0-ltv.Rebate0, ltv.PayUsers)))
				ltv.GameIncomeAvgLtv1Rate = fmt.Sprintf("%.2f%%", ComputeFloat(
					ComputeFloat(ltv.Bets0-ltv.Rebate0, ltv.PayUsers),
					ComputeFloat(ltv0.Bets0-ltv0.Rebate0, ltv0.PayUsers),
				)*100)
				ltv.Bets = fmt.Sprintf("%.2f", Chip2Float(ltv.Bets0))
				ltv.Rebates = fmt.Sprintf("%.2f", Chip2Float(ltv.Rebate0))
				ltv.GameIncome = fmt.Sprintf("%.2f", Chip2Float(ltv.Bets0-ltv.Rebate0))
				ltv.GameIncomeLtv1Rate = fmt.Sprintf("%.2f%%", ComputeFloat(
					ltv.Bets0-ltv.Rebate0,
					ltv0.Bets0-ltv0.Rebate0,
				)*100)

				ltv.GameIncomeSubGiveAvg = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(ltv.Bets0-ltv.Rebate0-ltv.GiveCash0, ltv.PayUsers)))
				ltv.GameIncomeSubGiveAvgLtv1Rate = fmt.Sprintf("%.2f%%", ComputeFloat(
					ComputeFloat(ltv.Bets0-ltv.Rebate0-ltv.GiveCash0, ltv.PayUsers),
					ComputeFloat(ltv0.Bets0-ltv0.Rebate0-ltv0.GiveCash0, ltv0.PayUsers),
				)*100)
				ltv.GameIncomeSubGive = fmt.Sprintf("%.2f", Chip2Float(ltv.Bets0-ltv.Rebate0-ltv.GiveCash0))
				ltv.GameIncomeSubGiveLtv1Rate = fmt.Sprintf("%.2f%%", ComputeFloat(
					ltv.Bets0-ltv.Rebate0-ltv.GiveCash0,
					ltv0.Bets0-ltv0.Rebate0-ltv0.GiveCash0,
				)*100)
			}
		}
	}
	return
}

// 首充口径ltv统计
func (this *statisticsService) LtvStatFirstPay(stime, etime time.Time, registAreas []int32, packageIds []string, betLtv bool) (stats []*entity.LtvStat, err error) {
	where_u_sql := ""
	var where_u_args []any

	if len(registAreas) > 0 {
		where_u_sql += " and s0.regist_area in ?"
		where_u_args = append(where_u_args, registAreas)
	}
	if len(packageIds) > 0 {
		where_u_sql += " and s0.ad__bundle_id IN ?"
		where_u_args = append(where_u_args, packageIds)
	}

	dateStats := make(map[uint32]*entity.LtvStat)

	// 首充用户
	var first_pay_datas []map[string]any
	first_pay_args := []any{stime, etime}
	first_pay_args = append(first_pay_args, where_u_args...)
	first_pay_args = append(first_pay_args, stime)
	err = ck.Select(&first_pay_datas, fmt.Sprintf(`
		select first_pay_date, count(*) first_pay_users, SUM(case when pay_orders0 >= 2 then 1 else 0 end) pay2_users, SUM(pay_orders0) pay_orders
		from (
			select s0.first_pay_date, s0.userid userid, count(*) pay_orders0
			from game.col_trade_record s1 final
			join (
				select toYYYYMMDD(s1.ctime) first_pay_date, s1.userid userid
				from game.col_user s0 final
				join game.col_trade_record s1 final on s0.userid = s1.userid
				where s1.ctime between ? AND ? and s1.order_status = 4 and s1.first_pay = 1 %s
				group by first_pay_date, s1.userid
			) s0 on s1.userid = s0.userid
			where s1.ctime > ? and s1.order_status = 4
			group by s0.first_pay_date, s0.userid
		) t0
		group by first_pay_date
		order by first_pay_date desc
	`, where_u_sql), first_pay_args...)
	if err != nil {
		return
	}
	now := NowTime()
	today := now.Year()*10000 + int(now.Month())*100 + now.Day()
	for _, data := range first_pay_datas {
		first_pay_date := data["first_pay_date"].(uint32)
		first_pay_users := utils.ToInt64(data["first_pay_users"])
		pay2_users := utils.ToInt64(data["pay2_users"])
		pay_orders := utils.ToInt64(data["pay_orders"])

		datestr := fmt.Sprint(first_pay_date)
		sdate := datestr[0:4] + "-" + datestr[4:6] + "-" + datestr[6:]
		stat := &entity.LtvStat{
			SDate: sdate,
		}
		stats = append(stats, stat)
		dateStats[first_pay_date] = stat

		stat.RegUsers = first_pay_users
		stat.PayUsers = first_pay_users
		stat.Pay2Users = pay2_users
		stat.PayOrders = pay_orders
		stat.Pay2UsersRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.Pay2Users, stat.PayUsers)*100)
		stat.PayOrdersAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.PayOrders, stat.PayUsers))

		var stime time.Time
		stime, err = time.ParseInLocation(utils.FORMAT_DATE, sdate, Location())
		if err != nil {
			return
		}

		for i := range 90 {
			ltvTime := stime.AddDate(0, 0, i)
			ltvDay := ltvTime.Year()*10000 + int(ltvTime.Month())*100 + ltvTime.Day()
			future := ltvDay >= today
			ltv := &entity.LtvStatDay{
				Future:                future,
				Day:                   int64(i),
				PayAvg:                "/",
				WithdrawAvg:           "/",
				ProfitAvg:             "/",
				ProfitAvgLtv1Rate:     "/",
				Pays:                  "/",
				Withdraws:             "/",
				Profit:                "/",
				ProfitLtv1Rate:        "/",
				PayUsersMap:           make(map[string]bool),
				BetsAvg:               "/",
				RebateAvg:             "/",
				GameIncomeAvg:         "/",
				GameIncomeAvgLtv1Rate: "/",
				Bets:                  "/",
				Rebates:               "/",
				GameIncome:            "/",
				GameIncomeLtv1Rate:    "/",
				NVProfitAvg:           "/",
				NVProfitAvgLtv1Rate:   "/",
				NVProfit:              "/",
				NVProfitLtv1Rate:      "/",
			}
			stat.Ltvs = append(stat.Ltvs, ltv)
		}
	}

	// 支付通道信息
	_, payChannels, err := PayService.PayChannelList(0)
	if err != nil {
		return
	}
	var payChannelMap = make(map[string]*entity.PayChannel)
	for _, c := range payChannels {
		payChannelMap[c.Id] = c
	}

	sql_ltv := `
		select reg_date, diff_days, channel_id, groupArray(pay_users) pay_users, SUM(pay_amounts) pay_amounts, 
			SUM(withdraw_amounts) withdraw_amounts, SUM(withdraw_orders) withdraw_orders
		from (
			select reg_date, diff_days, channel_id, groupArray(distinct userid) pay_users, SUM(amount) pay_amounts, 0 withdraw_amounts, 0 withdraw_orders
			from (
				select s0.userid userid, toString(s1.channel_id) channel_id, s0.first_pay_date reg_date, amount,
					dateDiff('day', toStartOfDay(s0.ctime), toStartOfDay(s1.ctime)) diff_days
				from game.col_trade_record s1 final
				join (
					select s1.userid userid, s1.ctime ctime, toYYYYMMDD(s1.ctime) first_pay_date 
					from game.col_user s0 final
					join game.col_trade_record s1 final on s0.userid = s1.userid
					where s1.ctime between ? AND ? and s1.order_status = 4 and s1.first_pay = 1 %s
					group by s1.userid, s1.ctime
				) s0 on s1.userid = s0.userid
				where s1.ctime > ? and s1.order_status = 4
			) t0
			where diff_days < 90
			group by reg_date, diff_days, channel_id
			order by reg_date desc, diff_days
			
			union all
			
			select reg_date, diff_days, channel_id, [] pay_users, 0 pay_amounts, SUM(amount) withdraw_amounts, count(*) withdraw_orders
			from (
				select s0.userid userid, s1.out_channel channel_id, s0.first_pay_date reg_date, amount,
					dateDiff('day', toStartOfDay(s0.ctime), toStartOfDay(s1.ctime)) diff_days
				from game.col_withdraw_record s1 final
				join (
					select s1.userid userid, s1.ctime ctime, toYYYYMMDD(s1.ctime) first_pay_date 
					from game.col_user s0 final
					join game.col_trade_record s1 final on s0.userid = s1.userid
					where s1.ctime between ? AND ? and s1.order_status = 4 and s1.first_pay = 1 %s
					group by s1.userid, s1.ctime
				) s0 on s1.userid = s0.userid
				where s1.order_status = 2
			) t0
			where diff_days < 90
			group by reg_date, diff_days, channel_id
		) t0
		group by reg_date, diff_days, channel_id
		order by reg_date desc, diff_days
	`
	var ltv_datas []map[string]any
	ltv_args := []any{stime, etime}
	ltv_args = append(ltv_args, where_u_args...)
	ltv_args = append(ltv_args, stime)
	ltv_args = append(ltv_args, stime, etime)
	ltv_args = append(ltv_args, where_u_args...)
	err = ck.Select(&ltv_datas, fmt.Sprintf(sql_ltv, where_u_sql, where_u_sql), ltv_args...)
	if err != nil {
		return
	}
	for _, data := range ltv_datas {
		reg_date := data["reg_date"].(uint32)
		diff_days := utils.ToInt64(data["diff_days"])
		pay_users := data["pay_users"].([][]string)
		pay_amounts := utils.ToInt64(data["pay_amounts"])
		withdraw_amounts := utils.ToInt64(data["withdraw_amounts"])
		withdraw_orders := utils.ToInt64(data["withdraw_orders"])
		channel_id := data["channel_id"].(string)
		if stat, ok := dateStats[reg_date]; ok {
			for i, ltv := range stat.Ltvs {
				if int64(i) >= diff_days {
					ltv.Withdraws0 += withdraw_amounts
					ltv.Pays0 += pay_amounts

					for _, pay_users0 := range pay_users {
						for _, userid := range pay_users0 {
							ltv.PayUsersMap[userid] = true
						}
					}

					if c, ok := payChannelMap[channel_id]; ok {
						// 代收手续费
						ltv.PaysTax0 += int64(float64(pay_amounts) * c.PayRate / 100)

						// 代付手续费
						ltv.WithdrawsTax0 += int64(float64(withdraw_amounts) * c.WithdrawRate / 100)
						ltv.WithdrawsTax0 += (withdraw_orders * c.WithdrawFee)
					}
				}
			}
		}
	}

	for _, stat := range stats {
		for _, ltv := range stat.Ltvs {
			if ltv.Future {
				break
			}
			ltv.PayUsers = int64(len(ltv.PayUsersMap))
			clear(ltv.PayUsersMap)
			ltv.PayAvg = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(ltv.Pays0, ltv.PayUsers)))
			ltv.WithdrawAvg = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(ltv.Withdraws0, ltv.PayUsers)))
			ltv.ProfitAvg0 = ComputeFloat(ltv.Pays0-ltv.Withdraws0, ltv.PayUsers)
			ltv.ProfitAvg = fmt.Sprintf("%.2f", Chip2Float(ltv.ProfitAvg0))
			ltv.NVProfitAvg0 = ComputeFloat(ltv.Pays0-ltv.Withdraws0-ltv.PaysTax0-ltv.WithdrawsTax0, ltv.PayUsers)
			ltv.NVProfitAvg = fmt.Sprintf("%.2f", Chip2Float(ltv.NVProfitAvg0))

			ltv.Pays = fmt.Sprintf("%.2f", Chip2Float(ltv.Pays0))
			ltv.Withdraws = fmt.Sprintf("%.2f", Chip2Float(ltv.Withdraws0))
			ltv.Profit0 = ltv.Pays0 - ltv.Withdraws0
			ltv.Profit = fmt.Sprintf("%.2f", Chip2Float(ltv.Profit0))
			ltv.NVProfit0 = ltv.Pays0 - ltv.Withdraws0 - ltv.PaysTax0 - ltv.WithdrawsTax0
			ltv.NVProfit = fmt.Sprintf("%.2f", Chip2Float(ltv.NVProfit0))

			ltv0 := stat.Ltvs[0]
			ltv.ProfitAvgLtv1Rate = fmt.Sprintf("%.2f%%", ComputeFloat(ltv.ProfitAvg0, ltv0.ProfitAvg0)*100)
			ltv.ProfitLtv1Rate = fmt.Sprintf("%.2f%%", ComputeFloat(ltv.Profit0, ltv0.Profit0)*100)
			ltv.NVProfitAvgLtv1Rate = fmt.Sprintf("%.2f%%", ComputeFloat(ltv.NVProfitAvg0, ltv0.NVProfitAvg0)*100)
			ltv.NVProfitLtv1Rate = fmt.Sprintf("%.2f%%", ComputeFloat(ltv.NVProfit0, ltv0.NVProfit0)*100)
		}
	}

	return
}

// 分页查询Ltv统计信息
func (this *statisticsService) GetLtvStatList(page, pageSize int, m bson.M, start, end time.Time) (stats []bson.M, err error) {
	for i := 0; ; i++ {
		stime := start.AddDate(0, 0, i)
		if stime.After(end) {
			break
		}
		stat := this.ltvStateByDay(m, stime)
		stats = append(stats, stat)
	}
	return
}

func (this *statisticsService) ltvStateByDay(m bson.M, today time.Time) (stat bson.M) {
	todayStr := today.Format("2006-01-02")
	stime := utils.Str2Time(fmt.Sprintf("%s 00:00:00", todayStr), Location())
	etime := utils.Str2Time(fmt.Sprintf("%s 23:59:59", todayStr), Location())

	stat = bson.M{
		"date":     todayStr, // 日期
		"count":    0,        // 新注册
		"scount":   0,        // 新设备
		"payCount": 0,        // 付费用户累积数
	}

	m["robot"] = false
	m["simulation_robot"] = false
	m["ctime"] = bson.M{"$gte": stime, "$lt": etime}

	// 新设备数
	query2 := []bson.M{
		{"$match": m},
		{
			"$group": bson.M{
				"_id": "$ad__device_id",
				"num": bson.M{"$sum": 1},
			},
		},
		{"$group": bson.M{"_id": nil, "num": bson.M{"$sum": 1}}},
	}
	pipe2 := PlayerUsers.Pipe(query2)
	result2 := []bson.M{}
	err2 := pipe2.All(&result2)
	if err2 == nil && len(result2) > 0 {
		stat["scount"] = result2[0]["num"]
	}

	// 注册用户
	list, _ := PlayerService.GetByUserList(m)
	count := len(list)
	userids := make([]string, 0, count)
	for _, user := range list {
		userids = append(userids, user.Userid)
	}

	// 新注册数
	if count > 0 {
		stat["count"] = count
	}

	// 新用户充值人数
	mPay := bson.M{
		"order_status": 4,
		"userid":       bson.M{"$in": userids},
	}
	query3 := []bson.M{
		{"$match": mPay},
		{"$group": bson.M{"_id": "$userid", "num": bson.M{"$sum": 1}}},
		// {"$group": bson.M{"_id": nil, "num": bson.M{"$sum": 1}}},
	}
	result := []bson.M{}
	if err := Pays.Pipe(query3).All(&result); err != nil {
		beego.Error(fmt.Errorf("新用户充值人数 error %v", err))
	}

	var payCount = int64(len(result)) // 付费人数
	stat["payCount"] = payCount

	// 复购率
	var repayRate = "0.00%"
	var pay2times int64
	var payOrders int // 付费成功订单数
	if payCount > 0 {
		for _, pay := range result {
			if s, ok := pay["num"].(int); ok {
				payOrders += s
				if s >= 2 {
					pay2times++
				}
			}
		}

		repayRate = fmt.Sprintf("%.2f%%", ComputeFloat(pay2times, payCount)*100)
	}
	stat["repayRate"] = repayRate
	stat["pay2times"] = pay2times

	// 人均付费统计: 付费成功订单数 / 付费人数
	var payRate = "0"
	if payCount > 0 {
		payRate = fmt.Sprintf("%.2f", float64(payOrders)/float64(payCount))
	}
	stat["payRate"] = payRate
	stat["payOrders"] = payOrders

	// ltv d1-d7
	// now := time.Now()
	// statEndTime := utils.TimestampTodayTime().AddDate(0, 0, 1) // 统计结束时间

	statEtimeStr := utils.TimestampTodayTimeZone(locationName).Format("2006-01-02")
	statEndTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", statEtimeStr), locationName)
	// statEndTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)

	// ltvs := []int{1, 2, 3, 4, 5, 6, 7, 15, 30}
	// for _, ltv := range ltvs {
	for ltv := 1; ltv <= 30; ltv++ {
		key := fmt.Sprintf("ltv%d", ltv)
		stat[key] = map[string]string{"pay": "0", "withdraw": "0", "sub": "0"}

		endTime := etime.Add(time.Duration(ltv-1) * (time.Hour * 24))
		if endTime.After(statEndTime) {
			stat[key] = map[string]string{"pay": "/", "withdraw": "/", "sub": "/"}
			continue
		}
		if payCount == 0 {
			continue
		}
		mPay["order_status"] = 4 // pay success status
		mPay["ctime"] = bson.M{"$gte": stime, "$lt": endTime}
		// ltv时间范围该批充值用户数
		mCharges := []bson.M{
			{"$match": mPay},
			{"$group": bson.M{"_id": "$userid", "num": bson.M{"$sum": 1}}},
			{"$group": bson.M{"_id": nil, "num": bson.M{"$sum": 1}}},
		}
		pipe := Pays.Pipe(mCharges)
		result := []bson.M{}
		err := pipe.All(&result)
		var ltvPayCount int64 = 0
		if err == nil && len(result) > 0 {
			if s, ok := result[0]["num"].(int); ok {
				ltvPayCount = int64(s)
			}
		}

		// ltv时间范围内该批充值总额
		payM := []bson.M{
			{"$match": mPay},
			{"$group": bson.M{"_id": nil, "sum": bson.M{"$sum": "$amount"}}},
		}
		pipe = Pays.Pipe(payM)
		result = []bson.M{}
		err = pipe.All(&result)
		var paySum int64 = 0
		if err == nil && len(result) > 0 {
			if s, ok := result[0]["sum"].(int); ok {
				paySum = int64(s)
			}
		}

		// ltv时间范围该批提现用户数
		// mPay["order_status"] = 2 // withdraw success status
		// mWithdraws := []bson.M{
		// 	{"$match": mPay},
		// 	{"$group": bson.M{"_id": "$userid", "num": bson.M{"$sum": 1}}},
		// 	{"$group": bson.M{"_id": nil, "num": bson.M{"$sum": 1}}},
		// }

		// pipe = Withdraws.Pipe(mWithdraws)
		// result = []bson.M{}
		// err = pipe.All(&result)
		// var ltvWithdrawCount int64 = 0
		// if err == nil && len(result) > 0 {
		// 	if s, ok := result[0]["num"].(int); ok {
		// 		ltvWithdrawCount = int64(s)
		// 	}
		// }

		// ltv时间范围内该批提现总额
		mPay["order_status"] = 2 // withdraw success status
		withdrawM := []bson.M{
			{"$match": mPay},
			{"$group": bson.M{"_id": nil, "sum": bson.M{"$sum": "$amount"}}},
		}
		pipe = Withdraws.Pipe(withdrawM)
		result = []bson.M{}
		err = pipe.All(&result)
		var withdrawSum int64 = 0
		if err == nil && len(result) > 0 {
			if s, ok := result[0]["sum"].(int); ok {
				withdrawSum = int64(s)
			}
		}
		var payLtv, withdrawLtv float64
		if ltvPayCount != 0 {
			payLtv = float64(paySum) / float64(ltvPayCount)
			withdrawLtv = float64(withdrawSum) / float64(ltvPayCount)
		}
		// if ltvWithdrawCount != 0 {
		// 	withdrawLtv = float64(withdrawSum) / float64(ltvWithdrawCount)
		// }

		stat[key] = map[string]string{
			"pay":      fmt.Sprintf("%.2f", payLtv/100.0),
			"withdraw": fmt.Sprintf("%.2f", withdrawLtv/100.0),
			"sub":      fmt.Sprintf("%.2f", ComputeFloat(paySum-withdrawSum, ltvPayCount)/100.0),
		}
	}

	return stat
}

// 分页查询Ltv统计信息
func (this *statisticsService) GetLtvStatSummaryList(page, pageSize int, m bson.M, start, end time.Time) (stats []bson.M, err error) {
	for i := 0; ; i++ {
		stime := start.AddDate(0, 0, i)
		if stime.After(end) {
			break
		}
		stat := this.ltvStateSummaryByDay(m, stime)
		stats = append(stats, stat)
	}
	return
}

func (this *statisticsService) ltvStateSummaryByDay(m bson.M, today time.Time) (stat bson.M) {
	todayStr := today.Format("2006-01-02")
	stime := utils.Str2Time(fmt.Sprintf("%s 00:00:00", todayStr), Location())
	etime := utils.Str2Time(fmt.Sprintf("%s 23:59:59", todayStr), Location())

	stat = bson.M{
		"date":     todayStr, // 日期
		"count":    0,        // 新注册
		"scount":   0,        // 新设备
		"payCount": 0,        // 付费用户累积数
	}

	m["robot"] = false
	m["simulation_robot"] = false
	m["ctime"] = bson.M{"$gte": stime, "$lt": etime}

	// 新设备数
	query2 := []bson.M{
		{"$match": m},
		{
			"$group": bson.M{
				"_id": "$ad__adid",
				"num": bson.M{"$sum": 1},
			},
		},
		{"$group": bson.M{"_id": nil, "num": bson.M{"$sum": 1}}},
	}
	pipe2 := PlayerUsers.Pipe(query2)
	result2 := []bson.M{}
	err2 := pipe2.All(&result2)
	if err2 == nil && len(result2) > 0 {
		stat["scount"] = result2[0]["num"]
	}

	// 注册用户
	list, _ := PlayerService.GetByUserList(m)
	count := len(list)
	userids := make([]string, 0, count)
	for _, user := range list {
		userids = append(userids, user.Userid)
	}

	// 新注册数
	if count > 0 {
		stat["count"] = count
	}

	// 新用户充值人数
	mPay := bson.M{
		"order_status": 4,
		"userid":       bson.M{"$in": userids},
	}

	query3 := []bson.M{
		{"$match": mPay},
		{"$group": bson.M{"_id": "$userid", "num": bson.M{"$sum": 1}}},
		// {"$group": bson.M{"_id": nil, "num": bson.M{"$sum": 1}}},
	}
	pipe := Pays.Pipe(query3)
	result := []bson.M{}
	err := pipe.All(&result)
	if err != nil {
		beego.Error(fmt.Errorf("新用户充值人数 error %v", err))
	}

	var payCount = int64(len(result))
	stat["payCount"] = payCount

	// 复购率
	var repayRate = "0.00%"
	var payOrders int // 付费成功订单数
	if payCount > 0 {
		var pay2times int64
		for _, pay := range result {
			if s, ok := pay["num"].(int); ok {
				payOrders += s
				if s >= 2 {
					pay2times++
				}
			}
		}
		repayRate = fmt.Sprintf("%.2f%%", ComputeFloat(pay2times, payCount)*100)
		stat["pay2times"] = pay2times
	}
	stat["repayRate"] = repayRate

	// 人均付费统计: 付费成功订单数 / 付费人数
	var payRate = "0"
	if payCount > 0 {
		payRate = fmt.Sprintf("%.2f", float64(payOrders)/float64(payCount))
	}
	stat["payRate"] = payRate
	stat["payOrders"] = payOrders

	// ltv d1-d7
	// now := time.Now()
	// statEndTime := utils.TimestampTodayTime().AddDate(0, 0, 1) // 统计结束时间

	statEtimeStr := utils.TimestampTodayTimeZone(locationName).Format("2006-01-02")
	statEndTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", statEtimeStr), locationName)
	// statEndTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)

	// ltvs := []int{1, 2, 3, 4, 5, 6, 7, 15, 30}
	// for _, ltv := range ltvs {
	for ltv := 1; ltv <= 30; ltv++ {
		key := fmt.Sprintf("ltv%d", ltv)
		stat[key] = map[string]string{"pay": "0", "withdraw": "0", "sub": "0"}

		endTime := etime.Add(time.Duration(ltv-1) * (time.Hour * 24))
		if endTime.After(statEndTime) {
			stat[key] = map[string]string{"pay": "/", "withdraw": "/", "sub": "/"}
			continue
		}
		if payCount == 0 {
			continue
		}
		mPay["order_status"] = 4 // pay success status
		mPay["ctime"] = bson.M{"$gte": stime, "$lt": endTime}
		// ltv时间范围该批充值用户数
		// mCharges := []bson.M{
		// 	{"$match": mPay},
		// 	{"$group": bson.M{"_id": "$userid", "num": bson.M{"$sum": 1}}},
		// 	{"$group": bson.M{"_id": nil, "num": bson.M{"$sum": 1}}},
		// }
		// pipe := Pays.Pipe(mCharges)
		// result := []bson.M{}
		// err := pipe.All(&result)
		// var ltvPayCount int64 = 0
		// if err == nil && len(result) > 0 {
		// 	if s, ok := result[0]["num"].(int); ok {
		// 		ltvPayCount = int64(s)
		// 	}
		// }

		// ltv时间范围内该批充值总额
		payM := []bson.M{
			{"$match": mPay},
			{"$group": bson.M{"_id": nil, "sum": bson.M{"$sum": "$amount"}}},
		}
		pipe = Pays.Pipe(payM)
		result = []bson.M{}
		err = pipe.All(&result)
		var paySum int64 = 0
		if err == nil && len(result) > 0 {
			if s, ok := result[0]["sum"].(int); ok {
				paySum = int64(s)
			}
		}

		// ltv时间范围该批提现用户数
		// mPay["order_status"] = 2 // withdraw success status
		// mWithdraws := []bson.M{
		// 	{"$match": mPay},
		// 	{"$group": bson.M{"_id": "$userid", "num": bson.M{"$sum": 1}}},
		// 	{"$group": bson.M{"_id": nil, "num": bson.M{"$sum": 1}}},
		// }

		// pipe = Withdraws.Pipe(mWithdraws)
		// result = []bson.M{}
		// err = pipe.All(&result)
		// var ltvWithdrawCount int64 = 0
		// if err == nil && len(result) > 0 {
		// 	if s, ok := result[0]["num"].(int); ok {
		// 		ltvWithdrawCount = int64(s)
		// 	}
		// }

		// ltv时间范围内该批提现总额
		mPay["order_status"] = 2 // withdraw success status
		withdrawM := []bson.M{
			{"$match": mPay},
			{"$group": bson.M{"_id": nil, "sum": bson.M{"$sum": "$amount"}}},
		}
		pipe = Withdraws.Pipe(withdrawM)
		result = []bson.M{}
		err = pipe.All(&result)
		var withdrawSum int64 = 0
		if err == nil && len(result) > 0 {
			if s, ok := result[0]["sum"].(int); ok {
				withdrawSum = int64(s)
			}
		}
		// var payLtv, withdrawLtv float64
		// if ltvPayCount != 0 {
		// 	payLtv = float64(paySum) / float64(ltvPayCount)
		// 	withdrawLtv = float64(withdrawSum) / float64(ltvPayCount)
		// }
		// if ltvWithdrawCount != 0 {
		// 	withdrawLtv = float64(withdrawSum) / float64(ltvWithdrawCount)
		// }

		stat[key] = map[string]string{
			"pay":      fmt.Sprintf("%.2f", Chip2Float(paySum)),
			"withdraw": fmt.Sprintf("%.2f", Chip2Float(withdrawSum)),
			"sub":      fmt.Sprintf("%.2f", Chip2Float(paySum-withdrawSum)),
		}
	}

	return stat
}

// 曲线
func (this *statisticsService) RecoveryCycleStats(days int, begin, end time.Time, exchangeRate float64, dateConsume map[string]float64, packageIds []string, username string, ptype int, pss string) (datas []*entity.RecoveryCycle, resultMap map[string][]float64, err error) {
	resultMap = make(map[string][]float64)

	// 提现
	sql2 := `
	SELECT dateDiff('day', ?, t1.ctime) df, SUM(amount + commission) w
	FROM game.col_withdraw_record t1 FINAL
	WHERE order_status = 2 AND userid IN (
	  SELECT userid FROM game.col_user FINAL WHERE ctime >= ? AND ctime < ? %s
	)
	AND t1.ctime >= ?
	GROUP BY df
	HAVING df <= ?
	`
	// 充值
	sql3 := `
	SELECT dateDiff('day', ?, t1.ctime) df, SUM(amount) p
	  FROM game.col_trade_record t1 FINAL
	WHERE order_status = 4 AND userid IN (
	  SELECT userid FROM game.col_user FINAL WHERE ctime >= ? AND ctime < ? %s
	)
	AND t1.ctime >= ?
	GROUP BY df
	HAVING df <= ?
	`

	var packageQuery string
	if len(packageIds) > 0 {
		packageQuery = " and ad__bundle_id in ?"
	}
	sql2 = fmt.Sprintf(sql2, packageQuery)
	sql3 = fmt.Sprintf(sql3, packageQuery)

	this.saveDateConsume(dateConsume, username, ptype, pss)
	var cdates []string
	cbegin := begin
	for !cbegin.After(end) {
		stime := cbegin
		cbegin = cbegin.AddDate(0, 0, 1)
		sdate := stime.Format("2006-01-02")
		cdates = append(cdates, sdate)
	}
	dateConsume = this.getDateConsume(cdates, ptype, pss)

	today := bson.Now()
	for !begin.After(end) {
		stime := begin
		etime := begin.AddDate(0, 0, 1)
		begin = begin.AddDate(0, 0, 1)
		sdate := stime.Format("2006-01-02")

		item := &entity.RecoveryCycle{
			SDate:   sdate,
			Consume: fmt.Sprintf("%.2f", dateConsume[sdate]),
		}
		datas = append(datas, item)

		consume := dateConsume[sdate]

		var r2, r3 []map[string]any
		var args2 []any
		if len(packageIds) == 0 {
			args2 = []any{stime, stime, etime, stime, days}
		} else {
			args2 = []any{stime, stime, etime, packageIds, stime, days}
		}
		err = ck.Select(&r2, sql2, args2...)
		if err != nil {
			return
		}
		err = ck.Select(&r3, sql3, args2...)
		if err != nil {
			return
		}

		var withdraws, pays = make(map[int64]int64), make(map[int64]int64)
		for _, r := range r2 {
			df := utils.ToInt64(r["df"])
			w := utils.ToInt64(r["w"])
			withdraws[df] = w
		}
		for _, r := range r3 {
			df := utils.ToInt64(r["df"])
			p := utils.ToInt64(r["p"])
			pays[df] = p
		}
		// 计算roi
		var profits int64
		var roiDays []float64
		for i := 0; i < days; i++ {
			if begin.AddDate(0, 0, i-1).After(today) {
				break
			}
			day := int64(i)
			profit := pays[day] - withdraws[day]
			profits += profit
			value := fmt.Sprintf("%.2f", ComputeFloat(ComputeFloat(float64(profits)/100, exchangeRate), consume)*100)
			roi, _ := strconv.ParseFloat(value, 64)
			roiDays = append(roiDays, roi)
		}
		resultMap[sdate] = roiDays
	}
	return
}

func (this *statisticsService) RecoveryCycle(begin, end time.Time, exchangeRate float64, dateConsume map[string]float64, packageIds []string, username string, ptype int, pss string) (results []*entity.RecoveryCycle, err error) {
	sql1 := `
	SELECT ctimestr, count(*) new_regs, SUM(CASE WHEN toYYYYMMDD(first_pay_time) = ctimestr THEN 1 ELSE 0 END) first_charges,
		SUM(CASE WHEN pay_amount > 0 THEN 1 ELSE 0 END) charges,
		(charges / new_regs) charge_rate
	FROM (
		SELECT t2.ctimestr, t2.userid, t2.ctimestr, SUM(t1.amount) pay_amount, MIN(t1.ctime) first_pay_time
		FROM game.col_trade_record t1 FINAL
		RIGHT JOIN (
			SELECT userid, toYYYYMMDD(ctime) ctimestr
			FROM game.col_user FINAL 
			WHERE ctime BETWEEN ? AND ? %s
		) t2 ON t1.userid = t2.userid AND t1.order_status = 4
		GROUP BY t2.ctimestr, t2.userid, t2.ctimestr
	) s1
	GROUP BY ctimestr
	ORDER BY ctimestr DESC
	`
	// ,SUM(money - cash_out) profit
	// 提现
	sql2 := `
	SELECT 
		SUM(amount + commission) w0,
		SUM(CASE WHEN dateDiff('day', ?, t1.ctime) = 0 THEN amount + commission ELSE 0 END) w1,
		SUM(CASE WHEN dateDiff('day', ?, t1.ctime) BETWEEN 0 AND 2 THEN amount + commission ELSE 0 END) w3,
		SUM(CASE WHEN dateDiff('day', ?, t1.ctime) BETWEEN 0 AND 6 THEN amount + commission ELSE 0 END) w7,
		SUM(CASE WHEN dateDiff('day', ?, t1.ctime) BETWEEN 0 AND 9 THEN amount + commission ELSE 0 END) w10,
		SUM(CASE WHEN dateDiff('day', ?, t1.ctime) BETWEEN 0 AND 14 THEN amount + commission ELSE 0 END) w15,
		SUM(CASE WHEN dateDiff('day', ?, t1.ctime) BETWEEN 0 AND 19 THEN amount + commission ELSE 0 END) w20,
		SUM(CASE WHEN dateDiff('day', ?, t1.ctime) BETWEEN 0 AND 29 THEN amount + commission ELSE 0 END) w30,
		SUM(CASE WHEN dateDiff('day', ?, t1.ctime) BETWEEN 0 AND 44 THEN amount + commission ELSE 0 END) w45,
		SUM(CASE WHEN dateDiff('day', ?, t1.ctime) BETWEEN 0 AND 59 THEN amount + commission ELSE 0 END) w60
	FROM game.col_withdraw_record t1 FINAL
	WHERE order_status = 2 AND userid IN (
		SELECT userid FROM game.col_user FINAL WHERE ctime >= ? AND ctime < ? %s
	)
	AND t1.ctime >= ?
	`
	// 充值
	sql3 := `
	SELECT 
		SUM(amount) p0,
		SUM(CASE WHEN dateDiff('day', ?, t1.ctime) = 0 THEN amount  ELSE 0 END) p1,
		SUM(CASE WHEN dateDiff('day', ?, t1.ctime) BETWEEN 0 AND 2 THEN amount ELSE 0 END) p3,
		SUM(CASE WHEN dateDiff('day', ?, t1.ctime) BETWEEN 0 AND 6 THEN amount ELSE 0 END) p7,
		SUM(CASE WHEN dateDiff('day', ?, t1.ctime) BETWEEN 0 AND 9 THEN amount ELSE 0 END) p10,
		SUM(CASE WHEN dateDiff('day', ?, t1.ctime) BETWEEN 0 AND 14 THEN amount ELSE 0 END) p15,
		SUM(CASE WHEN dateDiff('day', ?, t1.ctime) BETWEEN 0 AND 19 THEN amount ELSE 0 END) p20,
		SUM(CASE WHEN dateDiff('day', ?, t1.ctime) BETWEEN 0 AND 29 THEN amount ELSE 0 END) p30,
		SUM(CASE WHEN dateDiff('day', ?, t1.ctime) BETWEEN 0 AND 44 THEN amount ELSE 0 END) p45,
		SUM(CASE WHEN dateDiff('day', ?, t1.ctime) BETWEEN 0 AND 59 THEN amount ELSE 0 END) p60
	FROM game.col_trade_record t1 FINAL
	WHERE order_status = 4 AND userid IN (
		SELECT userid FROM game.col_user FINAL WHERE ctime >= ? AND ctime < ? %s
	)
	AND t1.ctime >= ?
	`
	var packageQuery string
	if len(packageIds) > 0 {
		packageQuery = " and ad__bundle_id in ?"
	}
	sql1 = fmt.Sprintf(sql1, packageQuery)
	sql2 = fmt.Sprintf(sql2, packageQuery)
	sql3 = fmt.Sprintf(sql3, packageQuery)

	var r1 []map[string]any
	args1 := []any{begin, end}
	if len(packageIds) > 0 {
		args1 = append(args1, packageIds)
	}
	err = ck.Select(&r1, sql1, args1...)
	if err != nil {
		return
	}

	this.saveDateConsume(dateConsume, username, ptype, pss)
	var cdates []string
	cbegin := begin
	for cbegin.Before(end) {
		stime := cbegin
		cbegin = cbegin.AddDate(0, 0, 1)
		sdate := stime.Format("2006-01-02")
		cdates = append(cdates, sdate)
	}
	dateConsume = this.getDateConsume(cdates, ptype, pss)

	var resultMap = make(map[string]*entity.RecoveryCycle)
	var dates []string
	for _, r := range r1 {
		ctimestr := fmt.Sprint(r["ctimestr"])
		sdate := ctimestr[0:4] + "-" + ctimestr[4:6] + "-" + ctimestr[6:]
		item := &entity.RecoveryCycle{
			SDate:        sdate,
			NewRegs:      utils.ToInt64(r["new_regs"]),
			FirstCharges: utils.ToInt64(r["first_charges"]),
			Charges:      utils.ToInt64(r["charges"]),
			Consume:      fmt.Sprintf("%.2f", dateConsume[sdate]),
		}
		consume := dateConsume[sdate]
		item.PayRates = fmt.Sprintf("%.2f%%", ComputeFloat(item.Charges, item.NewRegs)*100)
		item.EachRegCost = fmt.Sprintf("%.2f", ComputeFloat(consume, float64(item.NewRegs)))
		item.EachChargeCost = fmt.Sprintf("%.2f", ComputeFloat(consume, float64(item.FirstCharges)))
		item.EachPayCost = fmt.Sprintf("%.2f", ComputeFloat(consume, float64(item.Charges)))

		results = append(results, item)
		resultMap[sdate] = item
		dates = append(dates, sdate)
	}

	for begin.Before(end) {
		stime := begin
		etime := begin.AddDate(0, 0, 1)
		begin = begin.AddDate(0, 0, 1)
		// endOfDay := time.Date(begin.Year(), begin.Month(), begin.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), begin.Location())
		sdate := stime.Format("2006-01-02")
		item, ok := resultMap[sdate]
		if !ok {
			fmt.Printf("no data: %v, %v\n", sdate, dates)
			continue
		}
		consume := dateConsume[sdate]

		var r2 []map[string]any
		var args2 []any
		if len(packageIds) == 0 {
			args2 = []any{stime, stime, stime, stime, stime, stime, stime, stime, stime, stime, etime, stime}
		} else {
			args2 = []any{stime, stime, stime, stime, stime, stime, stime, stime, stime, stime, etime, packageIds, stime}
		}
		err = ck.Select(&r2, sql2, args2...)
		if err != nil {
			return
		}
		var withdraws = [9]float64{}
		var pays = [9]float64{}
		if len(r2) > 0 {
			withdraws[0] = ComputeFloat(utils.ToFloat64(r2[0]["w1"])/100, exchangeRate)
			withdraws[1] = ComputeFloat(utils.ToFloat64(r2[0]["w3"])/100, exchangeRate)
			withdraws[2] = ComputeFloat(utils.ToFloat64(r2[0]["w7"])/100, exchangeRate)
			withdraws[3] = ComputeFloat(utils.ToFloat64(r2[0]["w10"])/100, exchangeRate)
			withdraws[4] = ComputeFloat(utils.ToFloat64(r2[0]["w15"])/100, exchangeRate)
			withdraws[5] = ComputeFloat(utils.ToFloat64(r2[0]["w20"])/100, exchangeRate)
			withdraws[6] = ComputeFloat(utils.ToFloat64(r2[0]["w30"])/100, exchangeRate)
			withdraws[7] = ComputeFloat(utils.ToFloat64(r2[0]["w45"])/100, exchangeRate)
			withdraws[8] = ComputeFloat(utils.ToFloat64(r2[0]["w60"])/100, exchangeRate)
		}

		var r3 []map[string]any
		var args3 []any
		if len(packageIds) == 0 {
			args3 = []any{stime, stime, stime, stime, stime, stime, stime, stime, stime, stime, etime, stime}
		} else {
			args3 = []any{stime, stime, stime, stime, stime, stime, stime, stime, stime, stime, etime, packageIds, stime}
		}
		err = ck.Select(&r3, sql3, args3...)
		if err != nil {
			return
		}
		if len(r3) > 0 {
			pays[0] = ComputeFloat(utils.ToFloat64(r3[0]["p1"])/100, exchangeRate)
			pays[1] = ComputeFloat(utils.ToFloat64(r3[0]["p3"])/100, exchangeRate)
			pays[2] = ComputeFloat(utils.ToFloat64(r3[0]["p7"])/100, exchangeRate)
			pays[3] = ComputeFloat(utils.ToFloat64(r3[0]["p10"])/100, exchangeRate)
			pays[4] = ComputeFloat(utils.ToFloat64(r3[0]["p15"])/100, exchangeRate)
			pays[5] = ComputeFloat(utils.ToFloat64(r3[0]["p20"])/100, exchangeRate)
			pays[6] = ComputeFloat(utils.ToFloat64(r3[0]["p30"])/100, exchangeRate)
			pays[7] = ComputeFloat(utils.ToFloat64(r3[0]["p45"])/100, exchangeRate)
			pays[8] = ComputeFloat(utils.ToFloat64(r3[0]["p60"])/100, exchangeRate)
		}
		item.Profit = fmt.Sprintf("%.2f", ComputeFloat((utils.ToFloat64(r3[0]["p0"])-utils.ToFloat64(r2[0]["w0"]))/100, exchangeRate))
		item.ROI1 = fmt.Sprintf("%.2f%%", ComputeFloat(pays[0]-withdraws[0], consume)*100)
		item.ROI3 = fmt.Sprintf("%.2f%%", ComputeFloat(pays[1]-withdraws[1], consume)*100)
		item.ROI7 = fmt.Sprintf("%.2f%%", ComputeFloat(pays[2]-withdraws[2], consume)*100)
		item.ROI10 = fmt.Sprintf("%.2f%%", ComputeFloat(pays[3]-withdraws[3], consume)*100)
		item.ROI15 = fmt.Sprintf("%.2f%%", ComputeFloat(pays[4]-withdraws[4], consume)*100)
		item.ROI20 = fmt.Sprintf("%.2f%%", ComputeFloat(pays[5]-withdraws[5], consume)*100)
		item.ROI30 = fmt.Sprintf("%.2f%%", ComputeFloat(pays[6]-withdraws[6], consume)*100)
		item.ROI45 = fmt.Sprintf("%.2f%%", ComputeFloat(pays[7]-withdraws[7], consume)*100)
		item.ROI60 = fmt.Sprintf("%.2f%%", ComputeFloat(pays[8]-withdraws[8], consume)*100)
		item.ROI1Profit = fmt.Sprintf("%.2f", pays[0]-withdraws[0])
		item.ROI3Profit = fmt.Sprintf("%.2f", pays[1]-withdraws[1])
		item.ROI7Profit = fmt.Sprintf("%.2f", pays[2]-withdraws[2])
		item.ROI10Profit = fmt.Sprintf("%.2f", pays[3]-withdraws[3])
		item.ROI15Profit = fmt.Sprintf("%.2f", pays[4]-withdraws[4])
		item.ROI20Profit = fmt.Sprintf("%.2f", pays[5]-withdraws[5])
		item.ROI30Profit = fmt.Sprintf("%.2f", pays[6]-withdraws[6])
		item.ROI45Profit = fmt.Sprintf("%.2f", pays[7]-withdraws[7])
		item.ROI60Profit = fmt.Sprintf("%.2f", pays[8]-withdraws[8])
	}
	return
}

func (this *statisticsService) saveDateConsume(dateConsume map[string]float64, username string, ptype int, pss string) {
	now := bson.Now()
	for date, consume := range dateConsume {
		id := fmt.Sprintf("%s|%d|%s", date, ptype, pss)
		data := &entity.RecoverycycleDateConsume{
			Id:       id,
			Date:     date,
			Consume:  consume,
			Ptype:    ptype,
			Packages: pss,
			Ctime:    now,
			OptName:  username,
		}
		Upsert(RecoverycycleDateConsumes, bson.M{"_id": data.Id}, data)
	}
}

func (this *statisticsService) getDateConsume(cdates []string, ptype int, pss string) (dateConsumes map[string]float64) {
	var ids []string
	for _, date := range cdates {
		id := fmt.Sprintf("%s|%d|%s", date, ptype, pss)
		ids = append(ids, id)
	}
	var consumes []entity.RecoverycycleDateConsume
	ListByQ(RecoverycycleDateConsumes, bson.M{"_id": bson.M{"$in": ids}}, &consumes)
	dateConsumes = make(map[string]float64)
	for _, c := range consumes {
		dateConsumes[c.Date] = c.Consume
	}
	return
}

func (this *statisticsService) LosingPlayersCount(page, pageSize int, pipe []bson.M) (count int64, err error) {
	pipe = append(pipe, bson.M{
		"$group": bson.M{"_id": nil, "total": bson.M{"$sum": 1}},
	})
	var results []bson.M
	err = PlayerUsers.Pipe(pipe).All(&results)
	if err != nil {
		return
	}
	if len(results) > 0 {
		total, ok := results[0]["total"].(int)
		if ok {
			count = int64(total)
		}
	}
	// return int64(Count(PlayerUsers, m)), nil
	return
}

func (this *statisticsService) LosingPlayersList(page, pageSize int, pipe []bson.M) (results []map[string]any, err error) {
	if pageSize == -1 {
		pageSize = 100000
	}

	skipNum, _ := parsePageAndSort(page, pageSize, "login_time", false)
	pipe = append(pipe, bson.M{"$sort": bson.M{"login_time": -1}})
	pipe = append(pipe, bson.M{"$skip": skipNum})
	pipe = append(pipe, bson.M{"$limit": pageSize})

	var list []*entity.PlayerUser
	err = PlayerUsers.Pipe(pipe).All(&list)
	// err = PlayerUsers.
	// 	Find(m).
	// 	Sort(sortFieldR).
	// 	Skip(skipNum).
	// 	Limit(pageSize).
	// 	All(&list)
	if err != nil {
		return
	}
	results, err = this.losingPlayersMapping(list)
	return
}

func (this *statisticsService) losingPlayersMapping(players []*entity.PlayerUser) (list []map[string]any, err error) {
	// 充值数据
	var userids []string
	var minCtime = time.Now() // 最小注册时间,筛选对局数据
	for _, player := range players {
		userids = append(userids, player.Userid)
		if minCtime.After(player.Ctime) {
			minCtime = player.Ctime
		}
	}
	var paylist []*entity.PayOrder
	err = Pays.Find(bson.M{
		"userid":       bson.M{"$in": userids},
		"order_status": 4,
	}).All(&paylist)
	if err != nil {
		return
	}
	userPays := make(map[string][]*entity.PayOrder)
	for _, pay := range paylist {
		userPays[pay.Userid] = append(userPays[pay.Userid], pay)
	}
	// 从UserGameData中获取玩家游戏记录
	m2 := bson.M{}
	m2["userid"] = bson.M{"$in": userids}
	details := make([]entity.UserGameData, 0)
	UserGameDatas.Find(m2).All(&details)
	playerGtypeTimes := make(map[string]map[int]int32) // 对局游戏类型次数
	playerGameTimes := make(map[string]int32)          // 对局次数
	playerPlayTimes := make(map[string]int64)          // 对局时长
	if err != nil {
		beego.Error("details errors", err)
	} else {
		// gameTimes playTime gameTimesMax1 gameTimesMax2
		for _, detail := range details {
			uid := detail.Userid
			playerGameTimes[uid] += int32(detail.Number)
			playerPlayTimes[uid] += detail.GameTime
			gtypeTimes, ok := playerGtypeTimes[uid]
			if !ok {
				gtypeTimes = make(map[int]int32)
				playerGtypeTimes[uid] = gtypeTimes
			}
			gtypeTimes[int(detail.Gtype)] += int32(detail.Number)
		}
	}

	// // 对局数据
	// m2 := []bson.M{
	// 	{
	// 		"$match": bson.M{"begin_time": bson.M{"$gt": minCtime.Unix()}},
	// 	},
	// 	{
	// 		"$project": bson.M{
	// 			"room_id":    "$room_id",
	// 			"begin_time": "$begin_time",
	// 			"end_time":   "$end_time",
	// 			"gtype":      "$gtype",
	// 			"rtype":      "$rtype",
	// 			"players":    "$players",
	// 			"userids":    bson.M{"$split": []string{"$players", ","}},
	// 		},
	// 	},
	// 	{
	// 		"$match": bson.M{
	// 			"userids": bson.M{"$in": userids},
	// 		},
	// 	},
	// }
	// details := []bson.M{}
	// err = Details.Pipe(m2).All(&details)

	// playerGtypeTimes := make(map[string]map[int]int32) // 对局游戏类型次数
	// playerGameTimes := make(map[string]int32)          // 对局次数
	// playerPlayTimes := make(map[string]int64)          // 对局时长
	// if err != nil {
	// 	beego.Error("details errors", err)
	// } else {
	// 	// gameTimes playTime gameTimesMax1 gameTimesMax2
	// 	for _, detail := range details {
	// 		times := detail["end_time"].(int64) - detail["begin_time"].(int64)
	// 		for _, userid := range detail["userids"].([]interface{}) {
	// 			uid := userid.(string)
	// 			playerGameTimes[uid]++
	// 			playerPlayTimes[uid] += times

	// 			gtypeTimes, ok := playerGtypeTimes[uid]
	// 			if !ok {
	// 				gtypeTimes = make(map[int]int32)
	// 				playerGtypeTimes[uid] = gtypeTimes
	// 			}
	// 			gtypeTimes[detail["gtype"].(int)]++
	// 		}
	// 	}
	// }

	for _, player := range players {
		// 充值数据
		var (
			rechargeTimes int
			avgRecharge   string = "0"
		)
		if pays, ok := userPays[player.Userid]; ok {
			rechargeTimes = len(pays)
			if len(pays) > 0 {
				var payTotal uint32
				for _, p := range pays {
					payTotal += p.Amount
				}
				avgRecharge = fmt.Sprintf("%.2f", float64(payTotal)/float64(len(pays))/100.0)
			}
		}
		// 对局数据
		var gtype1, gtype2 string
		if gtypeTimes, ok := playerGtypeTimes[player.Userid]; ok {
			var t1, t2 int32
			var g1, g2 int
			for gtype, times := range gtypeTimes {
				if times > t1 {
					if t1 > 0 {
						t2, g2 = t1, g1
					}
					t1, g1 = times, gtype
				} else {
					if times > t2 {
						t2, g2 = times, gtype
					}
				}
			}
			if t, ok := GtypeNameMap[g1]; ok {
				gtype1 = t
			} else if g1 > 0 {
				gtype1 = strconv.Itoa(g1)
			}
			if t, ok := GtypeNameMap[g2]; ok {
				gtype2 = t
			} else if g2 > 0 {
				gtype2 = strconv.Itoa(g2)
			}
			// gtype1 = GameByName(g1)
			// gtype2 = GameByName(g2)
		}

		item := map[string]any{
			"userid":        player.Userid,
			"ctime":         player.Ctime.Format(utils.FORMAT),
			"loginTime":     player.LoginTime.Format(utils.FORMAT),
			"diamond":       fmt.Sprintf("%.2f", float64(player.Diamond)/100.0),
			"money":         fmt.Sprintf("%.2f", float64(player.Money)/100.0),
			"cashOut":       fmt.Sprintf("%.2f", float64(player.CashOut)/100.0),
			"winScore":      fmt.Sprintf("%.2f", float64(player.Diamond+player.ShadowDiamond+int64(player.CashOut)-int64(player.Money))/100.0),
			"rechargeTimes": rechargeTimes,
			"avgRecharge":   avgRecharge,
			"gameTimes":     playerGameTimes[player.Userid],
			"playTime":      fmt.Sprintf("%d分%d秒", playerPlayTimes[player.Userid]/60, playerPlayTimes[player.Userid]%60),
			"gameTimesMax1": gtype1,
			"gameTimesMax2": gtype2,
		}
		list = append(list, item)
	}
	return
}

// // 新增或更新付费用户留存
// func (this *statisticsService) AddPayUserRetained(user *entity.PayUserRetained) error {
// 	info := new(entity.PayUserRetained)
// 	if user.SType == 0 {
// 		GetByQ(PayUserRetaineds, bson.M{"date": user.Date, "s_type": user.SType}, info)
// 	} else {
// 		GetByQ(PayUserRetaineds, bson.M{"date": user.Date, "s_type": user.SType, "channel": user.Channel}, info)
// 	}
// 	if info.Id != "" {
// 		m := bson.M{"_id": info.Id}
// 		n := bson.M{
// 			"channel1":   user.Channel1,
// 			"new_number": user.NewNumber,
// 			"s_type":     user.SType,
// 			"day1":       user.Day1,
// 			"day2":       user.Day2,
// 			"day3":       user.Day3,
// 			"day4":       user.Day4,
// 			"day5":       user.Day5,
// 			"day6":       user.Day6,
// 			"day7":       user.Day7,
// 			"day15":      user.Day15,
// 			"day30":      user.Day30,
// 			"day60":      user.Day60,
// 		}
// 		if Update(PayUserRetaineds, m, bson.M{"$set": n}) {
// 			return nil
// 		}
// 		return errors.New("更新失败:" + user.Id)
// 	} else {
// 		// 新增
// 		user.Id = bson.NewObjectId().Hex()
// 		if !Insert(PayUserRetaineds, user) {
// 			return errors.New("写入失败:" + user.Id)
// 		}
// 		return nil
// 	}
// }

/*
充提排名
*/
func (this *statisticsService) GetPayRanking(typeid int, m bson.M) ([]entity.PayRanking, error) {
	var list []entity.PayRanking
	var paylist []entity.PayOrder
	var wlist []entity.WithdrawOrder
	ids := make([]string, 0)
	var err error
	switch typeid {
	case 0:
		err = Pays.
			Find(m).All(&paylist)
		if len(paylist) > 0 {
			groups := make(map[string][]entity.PayOrder)
			for _, pay := range paylist {
				Userid := pay.Userid
				groups[Userid] = append(groups[Userid], pay)
			}
			if len(groups) > 0 {
				for id, order := range groups {
					info := new(entity.PayRanking)
					amount := 0
					for _, a := range order {
						amount += int(a.Amount)
					}
					info.PayAmount = float64(amount)
					info.UserId = id
					list = append(list, *info)
					ids = append(ids, id)
				}
				w := m
				w["order_status"] = 2
				w["userid"] = bson.M{"$in": ids}
				wlist, w1err := PayService.GetByWithdrawUser(w)
				if w1err == nil {
					for i, v := range list {
						amount1 := 0
						for _, w := range wlist {
							if v.UserId == w.Userid {
								amount1 += int(w.Amount)
							}
						}
						list[i].WithdrawAmount = float64(amount1)
					}
				}
			}
		}
	case 1:
		// 提现排名
		err = Withdraws.Find(m).All(&wlist)
		if err == nil {
			groups := make(map[string][]entity.WithdrawOrder)
			for _, with := range wlist {
				Userid := with.Userid
				groups[Userid] = append(groups[Userid], with)
			}
			if len(groups) > 0 {
				for id, order := range groups {
					info := new(entity.PayRanking)
					amount := 0
					for _, a := range order {
						amount += int(a.Amount)
					}
					info.WithdrawAmount = float64(amount)
					info.UserId = id
					list = append(list, *info)
					ids = append(ids, id)
				}
				p := m
				p["order_status"] = 4
				p["userid"] = bson.M{"$in": ids}
				p1err := Pays.Find(p).All(&paylist)
				if p1err == nil {
					for i, v := range list {
						amount1 := 0
						for _, w := range paylist {
							if v.UserId == w.Userid {
								amount1 += int(w.Amount)
							}
						}
						list[i].PayAmount = float64(amount1)
					}
				}
			}
		}
	case 2:
		// 盈利
		err = Pays.
			Find(m).All(&paylist)
		w := m
		w["order_status"] = 2
		wlist, w1err := PayService.GetByWithdrawUser(w)
		if err == nil && w1err == nil {
			groups := make(map[string]string)
			for _, pay := range paylist {
				Userid := pay.Userid
				groups[Userid] = Userid
			}
			for _, with := range wlist {
				Userid := with.Userid
				groups[Userid] = Userid

			}
			if len(groups) > 0 {
				for _, id := range groups {
					info := new(entity.PayRanking)
					info.UserId = id
					amount := 0
					amount1 := 0
					for _, item := range paylist {
						if item.Userid == id {
							amount += int(item.Amount)
						}
					}
					for _, item := range wlist {
						if item.Userid == id {
							amount1 += int(item.Amount)
						}
					}
					info.PayAmount = float64(amount)
					info.WithdrawAmount = float64(amount1)
					list = append(list, *info)
					ids = append(ids, id)
				}
			}
		}
	}

	// 获取用户最后登录时间
	ulist, u1err := PlayerService.GetByUser(ids)
	if u1err == nil {
		for i, v := range list {
			for _, w := range ulist {
				if v.UserId == w.Userid {
					list[i].LoginTime = w.LoginTime
					list[i].RegistArea = w.RegistArea
				}
			}
			list[i].PayAmount = v.PayAmount / 100
			list[i].WithdrawAmount = v.WithdrawAmount / 100
			list[i].ProfitAmount = (v.WithdrawAmount - v.PayAmount) / 100
		}
	}
	switch typeid {
	case 0:
		// 使用 sort.Slice() 对用户信息进行排序
		sort.Slice(list, func(i, j int) bool {
			return list[i].PayAmount > list[j].PayAmount
		})
	case 1:
		// 使用 sort.Slice() 对用户信息进行排序
		sort.Slice(list, func(i, j int) bool {
			return list[i].WithdrawAmount > list[j].WithdrawAmount
		})
	case 2:
		// 使用 sort.Slice() 对用户信息进行排序
		sort.Slice(list, func(i, j int) bool {
			return list[i].ProfitAmount > list[j].ProfitAmount
		})
	}
	for i, _ := range list {
		list[i].Ranking = int64(i + 1)
	}
	list = this.chipList10(list)
	return list, err
}

func (this *statisticsService) chipList10(list []entity.PayRanking) []entity.PayRanking {
	for k, v := range list {
		c, _ := ConvertToIndiaTime(v.LoginTime.Unix())
		v.LoginTime = c
		list[k] = v
	}
	return list
}

/*
	渠道数据
*/
// 查询渠道数据信息
func (this *statisticsService) GetChannelList(page, pageSize int, m bson.M) ([]entity.ChannelData, error) {
	var list []entity.ChannelData
	if pageSize == -1 {
		pageSize = 100000
	}
	// skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "date", false)
	err := ChannelDatas.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList2(list)
	return list, err
}

// 转换为分展示
func (this *statisticsService) chipList2(list []entity.ChannelData) []entity.ChannelData {
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		// v.FNewRechargeAmount = Chip2Float(v.NewRechargeAmount)
		// v.FTotalAmount = Chip2Float(v.TotalAmount)
		// v.FTotalWithdrawAmount = Chip2Float(v.TotalWithdrawAmount)
		list[k] = v
	}
	return list
}

// 查询渠道数据条数
func (this *statisticsService) GetChannelTotal(m bson.M) (int64, error) {
	return int64(Count(ChannelDatas, m)), nil
}

// 渠道数据
func (this *statisticsService) AddChannelData(channel *entity.ChannelData) error {
	info := new(entity.ChannelData)
	GetByQ(ChannelDatas, bson.M{"date": channel.Date, "channel": channel.Channel}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"channel1":               channel.Channel1,
			"new_register":           channel.NewRegister,
			"new_equipment":          channel.NewEquipment,
			"login_num":              channel.LoginNum,
			"next_day_retention":     channel.NextDayRetention,
			"new_recharge_num":       channel.NewRechargeNum,
			"new_recharge_amount":    channel.NewRechargeAmount,
			"new_recharge_arpu":      channel.NewRechargeArpu,
			"new_recharge_arppu":     channel.NewRechargeArppu,
			"new_user_payment_rate":  channel.NewUserPaymentRate,
			"total_recharge":         channel.TotalRecharge,
			"total_amount":           channel.TotalAmount,
			"total_recharge_arpu":    channel.TotalRechargeArpu,
			"total_recharge_arppu":   channel.TotalRechargeArppu,
			"total_payment_rate":     channel.TotalPaymentRate,
			"pay_request":            channel.PayRequest,
			"pay_request_order":      channel.PayRequestOrder,
			"pay_success_order":      channel.PaySuccessOrder,
			"pay_success_rate":       channel.PaySuccessRate,
			"withdraw_request":       channel.WithdrawRequest,
			"withdraw_request_order": channel.WithdrawRequestOrder,
			"withdraw_success_order": channel.WithdrawSuccessOrder,
			"withdraw_success_rate":  channel.WithdrawSuccessRate,
			"total_withdraw_num":     channel.TotalWithdrawNum,
			"total_withdraw_amount":  channel.TotalWithdrawAmount,
			"handling_charge":        channel.HandlingCharge,
			"cost_ratio":             channel.CostRatio,
		}
		if Update(ChannelDatas, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + channel.Id)
	} else {
		// 新增
		channel.Id = bson.NewObjectId().Hex()
		if !Insert(ChannelDatas, channel) {
			return errors.New("写入失败:" + channel.Id)
		}
		return nil
	}
}

/*
用户资源
*/
// 查询用户资源信息
func (this *statisticsService) GetUserResourceList(page, pageSize int, m bson.M) ([]entity.UserResource, error) {
	var list []entity.UserResource
	if pageSize == -1 {
		pageSize = 100000
	}
	// skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "date", false)
	err := UserResources.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList3(list)
	return list, err
}

// 转换为分展示
func (this *statisticsService) chipList3(list []entity.UserResource) []entity.UserResource {
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		list[k] = v
	}
	return list
}

// 查询用户资源条数
func (this *statisticsService) GetUserResourceTotal(m bson.M) (int64, error) {
	return int64(Count(UserResources, m)), nil
}

// 用户资源
func (this *statisticsService) AddUserResource(user *entity.UserResource) error {
	info := new(entity.UserResource)
	GetByQ(UserResources, bson.M{"date": user.Date, "u_type": user.UType}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"user_number": user.UserNumber,
			"diamond":     user.Diamond,
			"coin":        user.Coin,
		}
		if Update(UserResources, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + user.Id)
	} else {
		// 新增
		user.Id = bson.NewObjectId().Hex()
		if !Insert(UserResources, user) {
			return errors.New("写入失败:" + user.Id)
		}
		return nil
	}
}

/*
	资源流动
*/
// 查询资源流动->全局信息
func (this *statisticsService) GetGlobalList(page, pageSize int, m bson.M) ([]entity.GlobalResource, error) {
	var list []entity.GlobalResource
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "date", false)
	err := GlobalResources.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList4(list)
	return list, err
}

// 转换为分展示
func (this *statisticsService) chipList4(list []entity.GlobalResource) []entity.GlobalResource {
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		list[k] = v
	}
	return list
}

// 查询资源流动->全局条数
func (this *statisticsService) GetGlobalTotal(m bson.M) (int64, error) {
	return int64(Count(GlobalResources, m)), nil
}

// 新增资源流动->全局
func (this *statisticsService) AddGlobalResource(res *entity.GlobalResource) error {
	info := new(entity.GlobalResource)
	GetByQ(GlobalResources, bson.M{"date": res.Date}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"login_num":    res.LoginNum,
			"new_register": res.NewRegister,
			// "total_recharge":  info.TotalRecharge,
			// "total_amount":    info.TotalAmount,
			// "withdraw_num":    info.WithdrawNum,
			// "withdraw_amount": info.WithdrawAmount,
			"put_diamond":    res.PutDiamond,
			"put_coin":       res.PutCoin,
			"expend_diamond": res.ExpendDiamond,
			"expend_coin":    res.ExpendCoin,
			"gift_diamond":   res.GiftDiamond,
			"gift_coin":      res.GiftCoin,
			// "pay_put_diamond": info.PayPutDiamond,
			// "pay_put_coin":    info.PayPutCoin,
		}
		if Update(GlobalResources, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + res.Id)
	} else {
		// 新增
		res.Id = bson.NewObjectId().Hex()
		if !Insert(GlobalResources, res) {
			return errors.New("写入失败:" + res.Id)
		}
		return nil
	}
}

// 查询资源流动->游戏信息
func (this *statisticsService) GetGamesList(page, pageSize int, m bson.M) ([]entity.GameResource, error) {
	var list []entity.GameResource
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "date", false)
	err := GameResources.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList5(list)
	return list, err
}

// 转换为分展示
func (this *statisticsService) chipList5(list []entity.GameResource) []entity.GameResource {
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		list[k] = v
	}
	return list
}

// 查询资源流动->游戏信息条数
func (this *statisticsService) GetGamesTotal(m bson.M) (int64, error) {
	return int64(Count(GameResources, m)), nil
}

// 新增资源流动->游戏信息
func (this *statisticsService) AddGamesResource(res *entity.GameResource) error {
	info := new(entity.GameResource)
	GetByQ(GameResources, bson.M{"date": res.Date}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"tp_number":            res.TpNumber,
			"tp_diamond":           res.TpDiamond,
			"tp_coin":              res.TpCoin,
			"tp_expend_diamond":    res.TpExpendDiamond,
			"tp_expend_coin":       res.TpExpendCoin,
			"rm_number":            res.RmNumber,
			"rm_diamond":           res.RmDiamond,
			"rm_coin":              res.RmCoin,
			"rm_expend_diamond":    res.RmExpendDiamond,
			"rm_expend_coin":       res.RmExpendCoin,
			"lhd_number":           res.LhdNumber,
			"lhd_diamond":          res.LhdDiamond,
			"lhd_coin":             res.LhdCoin,
			"lhd_expend_diamond":   res.LhdExpendDiamond,
			"lhd_expend_coin":      res.LhdExpendCoin,
			"up_number":            res.UpNumber,
			"up_diamond":           res.UpDiamond,
			"up_coin":              res.UpCoin,
			"up_expend_diamond":    res.UpExpendDiamond,
			"up_expend_coin":       res.UpExpendCoin,
			"crash_number":         res.CRASHNumber,
			"crash_diamond":        res.CRASHDiamond,
			"crash_coin":           res.CRASHCoin,
			"crash_expend_diamond": res.CRASHExpendDiamond,
			"crash_expend_coin":    res.CRASHExpendCoin,
		}
		if Update(GameResources, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + res.Id)
	} else {
		// 新增
		res.Id = bson.NewObjectId().Hex()
		if !Insert(GameResources, res) {
			return errors.New("写入失败:" + res.Id)
		}
		return nil
	}
}

// 查询实时数据
func (this *statisticsService) GetRealTimeData(m bson.M) ([]entity.RealTimeData, error) {
	var list []entity.RealTimeData
	// skipNum, sortFieldR := parsePageAndSort(page, pageSize, "date", false)
	err := RealTimes.
		Find(m).
		All(&list)
	list = this.chipList8(list)
	return list, err
}

// 转换为分展示
func (this *statisticsService) chipList8(list []entity.RealTimeData) []entity.RealTimeData {
	for k, v := range list {
		c, _ := ConvertToIndiaTime(v.Date.Unix())
		v.Date = c
		v.SDate = v.Date.Format("15:04")
		list[k] = v
	}
	return list
}

// 查询实时数据
// func (this *statisticsService) GetRealTimeTotal(m bson.M) (int64, error) {
// 	return int64(Count(RealTimes, m)), nil
// }

// 实时数据
func (this *statisticsService) AddRealTimeData(res *entity.RealTimeData) error {
	info := new(entity.RealTimeData)
	// date := res.Date.Format("2006-01-02 15:04:05")
	GetByQ(RealTimes, bson.M{"date": bson.M{"$eq": res.Date}}, info)
	// targetTime := date
	// filter :=
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"login_number":    info.LoginNumber,
			"online_number":   info.OnlineNumber,
			"pay_number":      info.PayNumber,
			"pay_amount":      info.PayAmount,
			"withdraw_number": info.WithdrawNumber,
			"withdraw_amount": info.WithdrawAmount,
		}
		if Update(RealTimes, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + res.Id)
	} else {
		// 新增
		// 解析时间字符串
		layout := "2006-01-02 15:04:05.000"
		deleteDateStr := res.Date.Format("2006-01-02 15:04:05.000")
		deleteDate, err := time.Parse(layout, deleteDateStr)
		if err != nil {
			return errors.New("解析时间字符串错误")
		}

		// 计算删除日期
		deleteThreshold := deleteDate.Add(-3 * 24 * time.Hour)

		// 构建删除条件
		filter := bson.D{{"date", bson.D{{"$lt", deleteThreshold}}}}

		// 执行删除操作
		if !DeleteAll(RealTimes, filter) {
			return errors.New("删除之前数据失败")
		}

		res.Id = bson.NewObjectId().Hex()
		if !Insert(RealTimes, res) {
			return errors.New("写入失败:" + res.Id)
		}
		return nil
	}
}

// 查询渠道成功率列表
func (this *statisticsService) GetSuccessRateList(page, pageSize int, m bson.M, isChannel bool) ([]entity.ChannelSuccessRate, error) {
	var list []entity.ChannelSuccessRate
	if pageSize == -1 {
		pageSize = 100000
	}
	// skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)
	pipeline := []bson.M{}
	if isChannel {
		// 按渠道
		pipeline = []bson.M{
			{"$match": m},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":         "$date",
						"package_id":   "$package_id",
						"package_name": "$package_name",
						"channel":      "$channel",
					},
					"login_number":             bson.M{"$sum": "$login_number"},
					"pay_request":              bson.M{"$sum": "$pay_request"},
					"pay_request_order":        bson.M{"$sum": "$pay_request_order"},
					"pay_success_order":        bson.M{"$sum": "$pay_success_order"},
					"pay_success_money":        bson.M{"$sum": "$pay_success_money"},
					"pay_fail_money":           bson.M{"$sum": "$pay_fail_money"},
					"withdraw_request":         bson.M{"$sum": "$withdraw_request"},
					"withdraw_request_order":   bson.M{"$sum": "$withdraw_request_order"},
					"withdraw_success_order":   bson.M{"$sum": "$withdraw_success_order"},
					"withdraw_success_money":   bson.M{"$sum": "$withdraw_success_money"},
					"withdraw_fail_money":      bson.M{"$sum": "$withdraw_fail_money"},
					"thirdparty_order":         bson.M{"$sum": "$thirdparty_order"},
					"thirdparty_success_order": bson.M{"$sum": "$thirdparty_success_order"},
				},
			},
			{
				"$project": bson.M{
					"_id":                      "$_id.date",
					"date":                     "$_id.date",
					"package_id":               "$_id.package_id",
					"package_name":             "$_id.package_name",
					"channel":                  "$_id.channel",
					"login_number":             "$login_number",
					"pay_request":              "$pay_request",
					"pay_request_order":        "$pay_request_order",
					"pay_success_order":        "$pay_success_order",
					"pay_success_money":        "$pay_success_money",
					"pay_fail_money":           "$pay_fail_money",
					"withdraw_request":         "$withdraw_request",
					"withdraw_request_order":   "$withdraw_request_order",
					"withdraw_success_order":   "$withdraw_success_order",
					"withdraw_success_money":   "$withdraw_success_money",
					"withdraw_fail_money":      "$withdraw_fail_money",
					"thirdparty_order":         "$thirdparty_order",
					"thirdparty_success_order": "$thirdparty_success_order",
				},
			},
			{"$sort": bson.M{"date": -1}},
			{"$skip": skipNum},
			{"$limit": pageSize},
		}
	} else {
		// 全部
		pipeline = []bson.M{
			{"$match": m},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date": "$date",
					},
					"login_number":             bson.M{"$sum": "$login_number"},
					"pay_request":              bson.M{"$sum": "$pay_request"},
					"pay_request_order":        bson.M{"$sum": "$pay_request_order"},
					"pay_success_order":        bson.M{"$sum": "$pay_success_order"},
					"pay_success_money":        bson.M{"$sum": "$pay_success_money"},
					"pay_fail_money":           bson.M{"$sum": "$pay_fail_money"},
					"withdraw_request":         bson.M{"$sum": "$withdraw_request"},
					"withdraw_request_order":   bson.M{"$sum": "$withdraw_request_order"},
					"withdraw_success_order":   bson.M{"$sum": "$withdraw_success_order"},
					"withdraw_success_money":   bson.M{"$sum": "$withdraw_success_money"},
					"withdraw_fail_money":      bson.M{"$sum": "$withdraw_fail_money"},
					"thirdparty_order":         bson.M{"$sum": "$thirdparty_order"},
					"thirdparty_success_order": bson.M{"$sum": "$thirdparty_success_order"},
				},
			},
			{
				"$project": bson.M{
					"_id":                      "$_id.date",
					"date":                     "$_id.date",
					"login_number":             "$login_number",
					"pay_request":              "$pay_request",
					"pay_request_order":        "$pay_request_order",
					"pay_success_order":        "$pay_success_order",
					"pay_success_money":        "$pay_success_money",
					"pay_fail_money":           "$pay_fail_money",
					"withdraw_request":         "$withdraw_request",
					"withdraw_request_order":   "$withdraw_request_order",
					"withdraw_success_order":   "$withdraw_success_order",
					"withdraw_success_money":   "$withdraw_success_money",
					"withdraw_fail_money":      "$withdraw_fail_money",
					"thirdparty_order":         "$thirdparty_order",
					"thirdparty_success_order": "$thirdparty_success_order",
				},
			},
			{"$sort": bson.M{"date": -1}},
			{"$skip": skipNum},
			{"$limit": pageSize},
		}
	}
	err := ChannelRates.Pipe(pipeline).All(&list)
	list = this.chipList9(list)
	return list, err
	// err := ChannelRates.
	// 	Find(m).
	// 	Sort(sortFieldR).
	// 	Skip(skipNum).
	// 	Limit(pageSize).
	// 	All(&list)
	// list = this.chipList9(list)
	// return list, err
}

// 渠道成功率折线图数据
func (this *statisticsService) GetSuccessRateChart(m bson.M) ([]entity.ChannelSuccessRate, error) {
	var list []entity.ChannelSuccessRate
	pipeline := []bson.M{
		{"$match": m},
		{
			"$group": bson.M{
				"_id": bson.M{
					"date":    "$date",
					"channel": "$channel",
				},
				"pay_request":              bson.M{"$sum": "$pay_request"},
				"pay_request_order":        bson.M{"$sum": "$pay_request_order"},
				"pay_success_order":        bson.M{"$sum": "$pay_success_order"},
				"pay_success_money":        bson.M{"$sum": "$pay_success_money"},
				"pay_fail_money":           bson.M{"$sum": "$pay_fail_money"},
				"withdraw_request":         bson.M{"$sum": "$withdraw_request"},
				"withdraw_request_order":   bson.M{"$sum": "$withdraw_request_order"},
				"withdraw_success_order":   bson.M{"$sum": "$withdraw_success_order"},
				"withdraw_success_money":   bson.M{"$sum": "$withdraw_success_money"},
				"withdraw_fail_money":      bson.M{"$sum": "$withdraw_fail_money"},
				"thirdparty_order":         bson.M{"$sum": "$thirdparty_order"},
				"thirdparty_success_order": bson.M{"$sum": "$thirdparty_success_order"},
			},
		},
		{
			"$project": bson.M{
				"_id":                      "$_id.date",
				"date":                     "$_id.date",
				"channel":                  "$_id.channel",
				"pay_request":              "$pay_request",
				"pay_request_order":        "$pay_request_order",
				"pay_success_order":        "$pay_success_order",
				"pay_success_money":        "$pay_success_money",
				"pay_fail_money":           "$pay_fail_money",
				"withdraw_request":         "$withdraw_request",
				"withdraw_request_order":   "$withdraw_request_order",
				"withdraw_success_order":   "$withdraw_success_order",
				"withdraw_success_money":   "$withdraw_success_money",
				"withdraw_fail_money":      "$withdraw_fail_money",
				"thirdparty_order":         "$thirdparty_order",
				"thirdparty_success_order": "$thirdparty_success_order",
			},
		},
		{"$sort": bson.M{"date": -1}},
	}
	err := ChannelRates.Pipe(pipeline).All(&list)
	list = this.chipList9(list)
	return list, err
	// err := ChannelRates.
	// 	Find(m).
	// 	Sort(sortFieldR).
	// 	Skip(skipNum).
	// 	Limit(pageSize).
	// 	All(&list)
	// list = this.chipList9(list)
	// return list, err
}

func getPayChannelByLoginNumber(info *entity.ChannelSuccessRate) int {
	number := 0
	var list []entity.ChannelSuccessRate
	// 获取
	m := bson.M{}
	m["date"] = info.Date
	if info.Channel != 0 {
		m["channel"] = info.Channel
	} else {
		// 默认值
		m["channel"] = 3010
	}
	if info.PackageId != "" {
		m["package_id"] = info.PackageId
	}
	ChannelRates.Find(m).All(&list)
	for _, item := range list {
		number += int(item.LoginNumber)
	}
	return number
}

// 转换为分展示
func (this *statisticsService) chipList9(list []entity.ChannelSuccessRate) []entity.ChannelSuccessRate {
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		// loginNumber := getPayChannelByLoginNumber(&v)
		// v.LoginNumber = int64(loginNumber)
		v.PullRate = ComputeFloat(v.PayRequest, v.LoginNumber) * 100
		v.PaySuccessRate = ComputeFloat(v.PaySuccessOrder, v.PayRequestOrder) * 100
		v.WithdrawSuccessRate = ComputeFloat(v.WithdrawSuccessOrder, v.WithdrawRequestOrder) * 100
		v.ThirdpartyOrderRate = ComputeFloat(v.ThirdpartyOrder, v.ThirdpartyOrder) * 100
		// v.FNewRechargeAmount = Chip2Float(v.NewRechargeAmount)
		// v.FTotalAmount = Chip2Float(v.TotalAmount)
		// v.FTotalWithdrawAmount = Chip2Float(v.TotalWithdrawAmount)
		list[k] = v
	}
	return list
}

// 查询渠道成功率条数
func (this *statisticsService) GetSuccessRateTotal(m bson.M, isChannel bool) (int64, error) {
	// return int64(Count(ChannelRates, m)), nil

	pipeline := []bson.M{}
	if isChannel {
		pipeline = []bson.M{
			{
				"$match": m,
			},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":         "$date",
						"package_id":   "$package_id",
						"package_name": "$package_name",
						"channel":      "$channel",
						// "login_number": "$login_number",
					},
				},
			},
		}
	} else {
		pipeline = []bson.M{
			{
				"$match": m,
			},
			{
				"$group": bson.M{
					"_id": "$date",
					// "login_number": "$login_number",
				},
			},
		}
	}
	result := []bson.M{}
	pipe := ChannelRates.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("GetSuccessRateTotal fail err: ", err)
	}
	return int64(len(result)), nil
}

// 渠道成功率
func (this *statisticsService) AddChannelRate(rate *entity.ChannelSuccessRate) error {
	info := new(entity.ChannelSuccessRate)
	GetByQ(ChannelRates, bson.M{"date": rate.Date, "package_id": rate.PackageId, "channel": rate.Channel}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"package_name":             rate.PackageName,
			"login_number":             rate.LoginNumber,
			"pay_request":              rate.PayRequest,
			"pay_request_order":        rate.PayRequestOrder,
			"pay_success_order":        rate.PaySuccessOrder,
			"pay_success_rate":         rate.PaySuccessRate,
			"pay_success_money":        rate.PaySuccessMoney,
			"pay_fail_money":           rate.PayFailMoney,
			"withdraw_request":         rate.WithdrawRequest,
			"withdraw_request_order":   rate.WithdrawRequestOrder,
			"withdraw_success_order":   rate.WithdrawSuccessOrder,
			"withdraw_success_rate":    rate.WithdrawSuccessRate,
			"withdraw_success_money":   rate.WithdrawSuccessMoney,
			"withdraw_fail_money":      rate.WithdrawFailMoney,
			"thirdparty_order":         rate.ThirdpartyOrder,
			"thirdparty_success_order": rate.ThirdpartySuccessOrder,
			"thirdparty_order_rate":    rate.ThirdpartyOrderRate,
		}
		if Update(ChannelRates, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + rate.Id)
	} else {
		// 新增
		rate.Id = bson.NewObjectId().Hex()
		if !Insert(ChannelRates, rate) {
			return errors.New("写入失败:" + rate.Id)
		}
		return nil
	}
}

func (this *statisticsService) AddStrategyData(res *entity.StrategyData) error {
	res.Id = bson.NewObjectId().Hex()
	if !Insert(StrategyDatas, res) {
		return errors.New("写入失败:" + res.Id)
	}
	return nil
}

/*
	埋点统计相关接口
*/
// 查询埋点数据列表
func (this *statisticsService) GetPointList(page, pageSize int, m bson.M, isChannel bool) ([]entity.PointData, error) {
	var list []entity.PointData
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)
	pipeline := []bson.M{}
	if isChannel {
		// 按渠道
		pipeline = []bson.M{
			{"$match": m},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":     "$date",
						"channel":  "$channel",
						"channel1": "$channel1",
					},
					"total_register":    bson.M{"$sum": "$total_register"},
					"total_tourist":     bson.M{"$sum": "$total_tourist"},
					"total_mobile":      bson.M{"$sum": "$total_mobile"},
					"guidance_binding":  bson.M{"$sum": "$guidance_binding"},
					"guidance_get_gold": bson.M{"$sum": "$guidance_get_gold"},
					"tp_number1":        bson.M{"$sum": "$tp_number1"},
					"tp_number2":        bson.M{"$sum": "$tp_number2"},
					"tp_exit_manually":  bson.M{"$sum": "$tp_exit_manually"},
					"gold100":           bson.M{"$sum": "$gold100"},
					"gold200":           bson.M{"$sum": "$gold200"},
					"first_games":       bson.M{"$sum": "$first_games"},
					"second_games":      bson.M{"$sum": "$second_games"},
				},
			},
			{
				"$project": bson.M{
					"_id":               "$_id.date",
					"date":              "$_id.date",
					"channel":           "$_id.channel",
					"channel1":          "$_id.channel1",
					"total_register":    "$total_register",
					"total_tourist":     "$total_tourist",
					"total_mobile":      "$total_mobile",
					"guidance_binding":  "$guidance_binding",
					"guidance_get_gold": "$guidance_get_gold",
					"tp_number1":        "$tp_number1",
					"tp_number2":        "$tp_number2",
					"tp_exit_manually":  "$tp_exit_manually",
					"gold100":           "$gold100",
					"gold200":           "$gold200",
					"first_games":       "$first_games",
					"second_games":      "$second_games",
				},
			},
			{"$sort": bson.M{"date": -1}},
			{"$skip": skipNum},
			{"$limit": pageSize},
		}
	} else {
		// 全部
		pipeline = []bson.M{
			{"$match": m},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date": "$date",
					},
					"total_register":    bson.M{"$sum": "$total_register"},
					"total_tourist":     bson.M{"$sum": "$total_tourist"},
					"total_mobile":      bson.M{"$sum": "$total_mobile"},
					"guidance_binding":  bson.M{"$sum": "$guidance_binding"},
					"guidance_get_gold": bson.M{"$sum": "$guidance_get_gold"},
					"tp_number1":        bson.M{"$sum": "$tp_number1"},
					"tp_number2":        bson.M{"$sum": "$tp_number2"},
					"tp_exit_manually":  bson.M{"$sum": "$tp_exit_manually"},
					"gold100":           bson.M{"$sum": "$gold100"},
					"gold200":           bson.M{"$sum": "$gold200"},
					"first_games":       bson.M{"$sum": "$first_games"},
					"second_games":      bson.M{"$sum": "$second_games"},
				},
			},
			{
				"$project": bson.M{
					"_id":               "$_id.date",
					"date":              "$_id.date",
					"total_register":    "$total_register",
					"total_tourist":     "$total_tourist",
					"total_mobile":      "$total_mobile",
					"guidance_binding":  "$guidance_binding",
					"guidance_get_gold": "$guidance_get_gold",
					"tp_number1":        "$tp_number1",
					"tp_number2":        "$tp_number2",
					"tp_exit_manually":  "$tp_exit_manually",
					"gold100":           "$gold100",
					"gold200":           "$gold200",
					"first_games":       "$first_games",
					"second_games":      "$second_games",
				},
			},
			{"$sort": bson.M{"date": -1}},
			{"$skip": skipNum},
			{"$limit": pageSize},
		}
	}
	err := PointDatas.Pipe(pipeline).All(&list)
	list = this.chipList11(list)
	return list, err
}

// 转换为分展示
func (this *statisticsService) chipList11(list []entity.PointData) []entity.PointData {
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		v.TotalTouristRatio = ComputeFloat(v.TotalTourist, v.TotalRegister) * 100
		v.TotalMobileRatio = ComputeFloat(v.TotalMobile, v.TotalRegister) * 100
		v.GuidanceGetGoldRatio = ComputeFloat(v.GuidanceGetGold, v.TotalRegister) * 100
		v.TPNumber1Ratio = ComputeFloat(v.TPNumber1, v.TotalRegister) * 100
		v.TPNumber2Ratio = ComputeFloat(v.TPNumber2, v.TotalRegister) * 100
		v.TPExitManuallyRatio = ComputeFloat(v.TPExitManually, v.TotalRegister) * 100
		v.Gold100Ratio = ComputeFloat(v.Gold100, v.TotalRegister) * 100
		v.Gold200Ratio = ComputeFloat(v.Gold200, v.TotalRegister) * 100
		list[k] = v
	}
	return list
}

// 查询埋点数据条数
func (this *statisticsService) GetPointListTotal(m bson.M, isChannel bool) (int64, error) {
	pipeline := []bson.M{}
	if isChannel {
		pipeline = []bson.M{
			{
				"$match": m,
			},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":     "$date",
						"channel":  "$channel",
						"channel1": "$channel1",
					},
				},
			},
		}
	} else {
		pipeline = []bson.M{
			{
				"$match": m,
			},
			{
				"$group": bson.M{
					"_id": "$date",
				},
			},
		}
	}
	result := []bson.M{}
	pipe := PointDatas.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("GetPointListTotal fail err: ", err)
	}
	return int64(len(result)), nil
}

// 新增埋点数据
func (this *statisticsService) AddPointData(rate *entity.PointData) error {
	info := new(entity.PointData)
	GetByQ(PointDatas, bson.M{"date": rate.Date, "channel": rate.Channel}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"channel1":          rate.Channel1,
			"total_register":    rate.TotalRegister,
			"total_tourist":     rate.TotalTourist,
			"total_mobile":      rate.TotalMobile,
			"guidance_binding":  rate.GuidanceBinding,
			"guidance_get_gold": rate.GuidanceGetGold,
			"tp_number1":        rate.TPNumber1,
			"tp_number2":        rate.TPNumber2,
			"tp_exit_manually":  rate.TPExitManually,
			"gold100":           rate.Gold100,
			"gold200":           rate.Gold200,
			"first_games":       rate.FirstGames,
			"second_games":      rate.SecondGames,
		}
		if Update(PointDatas, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + rate.Id)
	} else {
		// 新增
		rate.Id = bson.NewObjectId().Hex()
		if !Insert(PointDatas, rate) {
			return errors.New("写入失败:" + rate.Id)
		}
		return nil
	}
}

// 根据ID查询埋点数据信息
func (this *statisticsService) GetPointById(id int64, channelname string) (*entity.PointData, error) {
	info := new(entity.PointData)
	// pipeline := []bson.M{}
	// if channelname != "" {
	// 	// 按渠道
	// 	m := bson.M{"date": id, "channel": channelname}
	// 	pipeline = []bson.M{
	// 		{"$match": m},
	// 		{
	// 			"$group": bson.M{
	// 				"_id": bson.M{
	// 					"date":     "$date",
	// 					"channel":  "$channel",
	// 					"channel1": "$channel1",
	// 				},
	// 				"total_register":    bson.M{"$sum": "$total_register"},
	// 				"total_tourist":     bson.M{"$sum": "$total_tourist"},
	// 				"total_mobile":      bson.M{"$sum": "$total_mobile"},
	// 				"guidance_binding":  bson.M{"$sum": "$guidance_binding"},
	// 				"guidance_get_gold": bson.M{"$sum": "$guidance_get_gold"},
	// 				"tp_number1":        bson.M{"$sum": "$tp_number1"},
	// 				"tp_number2":        bson.M{"$sum": "$tp_number2"},
	// 				"tp_exit_manually":  bson.M{"$sum": "$tp_exit_manually"},
	// 				"gold100":           bson.M{"$sum": "$gold100"},
	// 				"gold200":           bson.M{"$sum": "$gold200"},
	// 				"first_games":       bson.M{"$sum": "$first_games"},
	// 				"second_games":      bson.M{"$sum": "$second_games"},
	// 			},
	// 		},
	// 		{
	// 			"$project": bson.M{
	// 				"_id":               "$_id.date",
	// 				"date":              "$_id.date",
	// 				"channel":           "$_id.channel",
	// 				"channel1":          "$_id.channel1",
	// 				"total_register":    "$total_register",
	// 				"total_tourist":     "$total_tourist",
	// 				"total_mobile":      "$total_mobile",
	// 				"guidance_binding":  "$guidance_binding",
	// 				"guidance_get_gold": "$guidance_get_gold",
	// 				"tp_number1":        "$tp_number1",
	// 				"tp_number2":        "$tp_number2",
	// 				"tp_exit_manually":  "$tp_exit_manually",
	// 				"gold100":           "$gold100",
	// 				"gold200":           "$gold200",
	// 				"first_games":       "$first_games",
	// 				"second_games":      "$second_games",
	// 			},
	// 		},
	// 		{"$sort": bson.M{"date": -1}},
	// 	}
	// } else {
	// 	// 全部
	// 	m := bson.M{"date": id}
	// 	pipeline = []bson.M{
	// 		{"$match": m},
	// 		{
	// 			"$group": bson.M{
	// 				"_id": bson.M{
	// 					"date": "$date",
	// 				},
	// 				"total_register":    bson.M{"$sum": "$total_register"},
	// 				"total_tourist":     bson.M{"$sum": "$total_tourist"},
	// 				"total_mobile":      bson.M{"$sum": "$total_mobile"},
	// 				"guidance_binding":  bson.M{"$sum": "$guidance_binding"},
	// 				"guidance_get_gold": bson.M{"$sum": "$guidance_get_gold"},
	// 				"tp_number1":        bson.M{"$sum": "$tp_number1"},
	// 				"tp_number2":        bson.M{"$sum": "$tp_number2"},
	// 				"tp_exit_manually":  bson.M{"$sum": "$tp_exit_manually"},
	// 				"gold100":           bson.M{"$sum": "$gold100"},
	// 				"gold200":           bson.M{"$sum": "$gold200"},
	// 				"first_games":       bson.M{"$sum": "$first_games"},
	// 				"second_games":      bson.M{"$sum": "$second_games"},
	// 			},
	// 		},
	// 		{
	// 			"$project": bson.M{
	// 				"_id":               "$_id.date",
	// 				"date":              "$_id.date",
	// 				"total_register":    "$total_register",
	// 				"total_tourist":     "$total_tourist",
	// 				"total_mobile":      "$total_mobile",
	// 				"guidance_binding":  "$guidance_binding",
	// 				"guidance_get_gold": "$guidance_get_gold",
	// 				"tp_number1":        "$tp_number1",
	// 				"tp_number2":        "$tp_number2",
	// 				"tp_exit_manually":  "$tp_exit_manually",
	// 				"gold100":           "$gold100",
	// 				"gold200":           "$gold200",
	// 				"first_games":       "$first_games",
	// 				"second_games":      "$second_games",
	// 			},
	// 		},
	// 		{"$sort": bson.M{"date": -1}},
	// 	}
	// }
	// err := PointDatas.Pipe(pipeline).One(&info)
	GetByQ(PointDatas, bson.M{"date": id, "channel": channelname}, info)
	if info.Id == "" {
		return info, errors.New("未查询数据")
	}
	return info, nil
}

func (this *statisticsService) GetLogEventTracksTotal(m bson.M) (int64, error) {
	return int64(Count(LogEventTracks, m)), nil
}

// 分页查询局数分析数据列表
func (this *statisticsService) GetGameNumberAnalysisList(page, pageSize int, m bson.M, isChannel bool) ([]entity.GameNumberAnalysis, error) {
	var list []entity.GameNumberAnalysis
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)
	pipeline := []bson.M{}
	if isChannel {
		// 按渠道
		pipeline = []bson.M{
			{"$match": m},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":     "$date",
						"channel":  "$channel",
						"channel1": "$channel1",
					},
					"total_register": bson.M{"$sum": "$total_register"},
					"g_number":       bson.M{"$sum": "$g_number"},
					"g_number1":      bson.M{"$sum": "$g_number1"},
					"g_number2":      bson.M{"$sum": "$g_number2"},
					"g_number3":      bson.M{"$sum": "$g_number3"},
					"g_number4":      bson.M{"$sum": "$g_number4"},
					"g_number5":      bson.M{"$sum": "$g_number5"},
					"g_number6":      bson.M{"$sum": "$g_number6"},
					"g_number11":     bson.M{"$sum": "$g_number11"},
					"g_number21":     bson.M{"$sum": "$g_number21"},
					"g_number31":     bson.M{"$sum": "$g_number31"},
				},
			},
			{
				"$project": bson.M{
					"_id":            "$_id.date",
					"date":           "$_id.date",
					"channel":        "$_id.channel",
					"channel1":       "$_id.channel1",
					"total_register": "$total_register",
					"g_number":       "$g_number",
					"g_number1":      "$g_number1",
					"g_number2":      "$g_number2",
					"g_number3":      "$g_number3",
					"g_number4":      "$g_number4",
					"g_number5":      "$g_number5",
					"g_number6":      "$g_number6",
					"g_number11":     "$g_number11",
					"g_number21":     "$g_number21",
					"g_number31":     "$g_number31",
				},
			},
			{"$sort": bson.M{"date": -1}},
			{"$skip": skipNum},
			{"$limit": pageSize},
		}
	} else {
		// 全部
		pipeline = []bson.M{
			{"$match": m},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date": "$date",
					},
					"total_register": bson.M{"$sum": "$total_register"},
					"g_number":       bson.M{"$sum": "$g_number"},
					"g_number1":      bson.M{"$sum": "$g_number1"},
					"g_number2":      bson.M{"$sum": "$g_number2"},
					"g_number3":      bson.M{"$sum": "$g_number3"},
					"g_number4":      bson.M{"$sum": "$g_number4"},
					"g_number5":      bson.M{"$sum": "$g_number5"},
					"g_number6":      bson.M{"$sum": "$g_number6"},
					"g_number11":     bson.M{"$sum": "$g_number11"},
					"g_number21":     bson.M{"$sum": "$g_number21"},
					"g_number31":     bson.M{"$sum": "$g_number31"},
				},
			},
			{
				"$project": bson.M{
					"_id":            "$_id.date",
					"date":           "$_id.date",
					"total_register": "$total_register",
					"g_number":       "$g_number",
					"g_number1":      "$g_number1",
					"g_number2":      "$g_number2",
					"g_number3":      "$g_number3",
					"g_number4":      "$g_number4",
					"g_number5":      "$g_number5",
					"g_number6":      "$g_number6",
					"g_number11":     "$g_number11",
					"g_number21":     "$g_number21",
					"g_number31":     "$g_number31",
				},
			},
			{"$sort": bson.M{"date": -1}},
			{"$skip": skipNum},
			{"$limit": pageSize},
		}
	}
	err := GameNumberAnalysiss.Pipe(pipeline).All(&list)
	list = this.chipList14(list)
	return list, err
}

// 转换为分展示
func (this *statisticsService) chipList14(list []entity.GameNumberAnalysis) []entity.GameNumberAnalysis {
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		v.GNumberRatio = ComputeFloat(v.GNumber, v.TotalRegister) * 100
		v.GNumberRatio1 = ComputeFloat(v.GNumber1, v.TotalRegister) * 100
		v.GNumberRatio2 = ComputeFloat(v.GNumber2, v.TotalRegister) * 100
		v.GNumberRatio3 = ComputeFloat(v.GNumber3, v.TotalRegister) * 100
		v.GNumberRatio4 = ComputeFloat(v.GNumber4, v.TotalRegister) * 100
		v.GNumberRatio5 = ComputeFloat(v.GNumber5, v.TotalRegister) * 100
		v.GNumberRatio6 = ComputeFloat(v.GNumber6, v.TotalRegister) * 100
		v.GNumberRatio11 = ComputeFloat(v.GNumber11, v.TotalRegister) * 100
		v.GNumberRatio21 = ComputeFloat(v.GNumber21, v.TotalRegister) * 100
		v.GNumberRatio31 = ComputeFloat(v.GNumber31, v.TotalRegister) * 100
		// v.FNewRechargeAmount = Chip2Float(v.NewRechargeAmount)
		// v.FTotalAmount = Chip2Float(v.TotalAmount)
		// v.FTotalWithdrawAmount = Chip2Float(v.TotalWithdrawAmount)
		list[k] = v
	}
	return list
}

// 查询局数分析数据条数
func (this *statisticsService) GetGameNumberAnalysisTotal(m bson.M, isChannel bool) (int64, error) {
	pipeline := []bson.M{}
	if isChannel {
		pipeline = []bson.M{
			{
				"$match": m,
			},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":     "$date",
						"channel":  "$channel",
						"channel1": "$channel1",
					},
				},
			},
		}
	} else {
		pipeline = []bson.M{
			{
				"$match": m,
			},
			{
				"$group": bson.M{
					"_id": "$date",
				},
			},
		}
	}

	// operations := []bson.M{m, n}
	result := []bson.M{}
	pipe := GameNumberAnalysiss.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("GetGameNumberAnalysisTotal fail err: ", err)
	}
	return int64(len(result)), nil
	// return int64(Count(GameNumberAnalysiss, m)), nil
}

// 新增局数分析数据
func (this *statisticsService) AddGameNumberAnalysis(rate *entity.GameNumberAnalysis) error {
	info := new(entity.GameNumberAnalysis)
	GetByQ(GameNumberAnalysiss, bson.M{"date": rate.Date, "channel": rate.Channel}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"channel1":       rate.Channel1,
			"total_register": rate.TotalRegister,
			"g_number":       rate.GNumber,
			"g_number1":      rate.GNumber1,
			"g_number2":      rate.GNumber2,
			"g_number3":      rate.GNumber3,
			"g_number4":      rate.GNumber4,
			"g_number5":      rate.GNumber5,
			"g_number6":      rate.GNumber6,
			"g_number11":     rate.GNumber11,
			"g_number21":     rate.GNumber21,
			"g_number31":     rate.GNumber31,
		}
		if Update(GameNumberAnalysiss, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + rate.Id)
	} else {
		// 新增
		rate.Id = bson.NewObjectId().Hex()
		if !Insert(GameNumberAnalysiss, rate) {
			return errors.New("写入失败:" + rate.Id)
		}
		return nil
	}
}

/*
Bug统计相关接口
*/
func (this *statisticsService) GetBugList(page, pageSize int, m bson.M) ([]entity.BugStatistics, error) {
	var list []entity.BugStatistics
	if pageSize == -1 {
		pageSize = 100000
	}
	// skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "date", false)
	err := BugDatas.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList12(list)
	return list, err
}

// 转换为分展示
func (this *statisticsService) chipList12(list []entity.BugStatistics) []entity.BugStatistics {
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		// v.FNewRechargeAmount = Chip2Float(v.NewRechargeAmount)
		// v.FTotalAmount = Chip2Float(v.TotalAmount)
		// v.FTotalWithdrawAmount = Chip2Float(v.TotalWithdrawAmount)
		list[k] = v
	}
	return list
}

// 查询Bug数据条数
func (this *statisticsService) GetBugListTotal(m bson.M) (int64, error) {
	return int64(Count(BugDatas, m)), nil
}

// 新增Bug数据
func (this *statisticsService) AddBugData(rate *entity.BugStatistics) error {
	info := new(entity.BugStatistics)
	GetByQ(BugDatas, bson.M{"date": rate.Date}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"bug_type1": rate.BugType1,
			"bug_type2": rate.BugType2,
			"bug_type3": rate.BugType3,
			"bug_type4": rate.BugType4,
			"bug_type5": rate.BugType5,
		}
		if Update(BugDatas, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + rate.Id)
	} else {
		// 新增
		rate.Id = bson.NewObjectId().Hex()
		if !Insert(BugDatas, rate) {
			return errors.New("写入失败:" + rate.Id)
		}
		return nil
	}
}

// Bug领取记录
func (this *statisticsService) GetBugRecordList(page, pageSize int, m bson.M) ([]entity.BugFeedbackLog, error) {
	var list []entity.BugFeedbackLog
	if pageSize == -1 {
		pageSize = 100000
	}
	// skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	err := LogBugFeedbacks.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList13(list)
	return list, err
}

// 转换为分展示
func (this *statisticsService) chipList13(list []entity.BugFeedbackLog) []entity.BugFeedbackLog {
	for k, v := range list {
		s, _ := ConvertToIndiaTime1(v.Ctime)
		v.SDate = s
		cur := v.SDate
		// 获取当天的开始时间
		startOfDay := time.Date(cur.Year(), cur.Month(), cur.Day(), 0, 0, 0, 0, cur.Location())
		// 获取当天的结束时间
		endOfDay := time.Date(cur.Year(), cur.Month(), cur.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), cur.Location())
		start1 := startOfDay.UnixNano() / int64(time.Millisecond)
		end1 := endOfDay.UnixNano() / int64(time.Millisecond)
		m1 := bson.M{}
		m1["ctime"] = bson.M{"$gte": start1, "$lte": end1}
		m1["user_id"] = v.UserId

		daycount, _ := StatisticsService.GetBugRecordTotal(m1)
		v.DaySubmit = daycount
		list[k] = v
	}
	return list
}

// 查询Bug数据条数
func (this *statisticsService) GetBugRecordTotal(m bson.M) (int64, error) {
	return int64(Count(LogBugFeedbacks, m)), nil
}

func (this *statisticsService) GetLogBugFeedbackLog(m bson.M) (int64, error) {
	return int64(Count(LogBugFeedbacks, m)), nil
}

/*
 时间分析相关接口 Playtimes
*/
// 时间分析分页接口
func (this *statisticsService) GetPlaytimeList(page, pageSize int, m bson.M, isChannel bool) ([]entity.PlaytimeAnalysis, error) {
	var list []entity.PlaytimeAnalysis
	if pageSize == -1 {
		pageSize = 100000
	}
	// skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	// skipNum, sortFieldR := parsePageAndSort(page, pageSize, "date", false)
	// err := Playtimes.
	// 	Find(m).
	// 	Sort(sortFieldR).
	// 	Skip(skipNum).
	// 	Limit(pageSize).
	// 	All(&list)
	// list = this.chipList15(list)
	// return list, err
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)
	pipeline := []bson.M{}
	if isChannel {
		// 按渠道
		pipeline = []bson.M{
			{"$match": m},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":     "$date",
						"channel":  "$channel",
						"channel1": "$channel1",
					},
					"total_register": bson.M{"$sum": "$total_register"},
					"playtime1":      bson.M{"$sum": "$playtime1"},
					"playtime2":      bson.M{"$sum": "$playtime2"},
					"playtime6":      bson.M{"$sum": "$playtime6"},
					"playtime11":     bson.M{"$sum": "$playtime11"},
					"playtime21":     bson.M{"$sum": "$playtime21"},
					"playtime31":     bson.M{"$sum": "$playtime31"},
				},
			},
			{
				"$project": bson.M{
					"_id":            "$_id.date",
					"date":           "$_id.date",
					"channel":        "$_id.channel",
					"channel1":       "$_id.channel1",
					"total_register": "$total_register",
					"playtime1":      "$playtime1",
					"playtime2":      "$playtime2",
					"playtime6":      "$playtime6",
					"playtime11":     "$playtime11",
					"playtime21":     "$playtime21",
					"playtime31":     "$playtime31",
				},
			},
			{"$sort": bson.M{"date": -1}},
			{"$skip": skipNum},
			{"$limit": pageSize},
		}
	} else {
		// 全部
		pipeline = []bson.M{
			{"$match": m},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date": "$date",
					},
					"total_register": bson.M{"$sum": "$total_register"},
					"playtime1":      bson.M{"$sum": "$playtime1"},
					"playtime2":      bson.M{"$sum": "$playtime2"},
					"playtime6":      bson.M{"$sum": "$playtime6"},
					"playtime11":     bson.M{"$sum": "$playtime11"},
					"playtime21":     bson.M{"$sum": "$playtime21"},
					"playtime31":     bson.M{"$sum": "$playtime31"},
				},
			},
			{
				"$project": bson.M{
					"_id":            "$_id.date",
					"date":           "$_id.date",
					"total_register": "$total_register",
					"playtime1":      "$playtime1",
					"playtime2":      "$playtime2",
					"playtime6":      "$playtime6",
					"playtime11":     "$playtime11",
					"playtime21":     "$playtime21",
					"playtime31":     "$playtime31",
				},
			},
			{"$sort": bson.M{"date": -1}},
			{"$skip": skipNum},
			{"$limit": pageSize},
		}
	}
	err := Playtimes.Pipe(pipeline).All(&list)
	list = this.chipList15(list)
	return list, err
}

// 转换为分展示
func (this *statisticsService) chipList15(list []entity.PlaytimeAnalysis) []entity.PlaytimeAnalysis {
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		v.PlaytimeRatio1 = ComputeFloat(v.Playtime1, v.TotalRegister) * 100
		v.PlaytimeRatio2 = ComputeFloat(v.Playtime2, v.TotalRegister) * 100
		v.PlaytimeRatio6 = ComputeFloat(v.Playtime6, v.TotalRegister) * 100
		v.PlaytimeRatio11 = ComputeFloat(v.Playtime11, v.TotalRegister) * 100
		v.PlaytimeRatio21 = ComputeFloat(v.Playtime21, v.TotalRegister) * 100
		v.PlaytimeRatio31 = ComputeFloat(v.Playtime31, v.TotalRegister) * 100
		list[k] = v
	}
	return list
}

// 查询Bug数据条数
func (this *statisticsService) GetPlaytimeTotal(m bson.M, isChannel bool) (int64, error) {
	// return int64(Count(Playtimes, m)), nil
	pipeline := []bson.M{}
	if isChannel {
		pipeline = []bson.M{
			{
				"$match": m,
			},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":     "$date",
						"channel":  "$channel",
						"channel1": "$channel1",
					},
				},
			},
		}
	} else {
		pipeline = []bson.M{
			{
				"$match": m,
			},
			{
				"$group": bson.M{
					"_id": "$date",
				},
			},
		}
	}

	// operations := []bson.M{m, n}
	result := []bson.M{}
	pipe := Playtimes.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("GetPlaytimeTotal fail err: ", err)
	}
	return int64(len(result)), nil
}

// 新增Bug数据
func (this *statisticsService) AddPlaytimeAnalysis(rate *entity.PlaytimeAnalysis) error {
	info := new(entity.PlaytimeAnalysis)
	GetByQ(Playtimes, bson.M{"date": rate.Date, "channel": rate.Channel}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"channel1":       rate.Channel1,
			"total_register": rate.TotalRegister,
			"playtime1":      rate.Playtime1,
			"playtime2":      rate.Playtime2,
			"playtime6":      rate.Playtime6,
			"playtime11":     rate.Playtime11,
			"playtime21":     rate.Playtime21,
			"playtime31":     rate.Playtime31,
		}
		if Update(Playtimes, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + rate.Id)
	} else {
		// 新增
		rate.Id = bson.NewObjectId().Hex()
		if !Insert(Playtimes, rate) {
			return errors.New("写入失败:" + rate.Id)
		}
		return nil
	}
}

func (this *statisticsService) GetLogGameTimes(m bson.M) (int64, error) {
	return int64(Count(LogGameTimes, m)), nil
}

/*
	房间数据相关接口
*/
// 分页查询房间数据统计
func (this *statisticsService) GetRoomDataList(page, pageSize int, m bson.M, isChannel bool) ([]entity.RoomDataList, error) {
	var list []entity.RoomDataList
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)
	// err := RoomDatas.
	// 	Find(m).
	// 	Sort(sortFieldR).
	// 	Skip(skipNum).
	// 	Limit(pageSize).
	// 	All(&list)
	// list = this.chipList16(list)
	pipeline := []bson.M{}
	if isChannel {
		// 按渠道
		pipeline = []bson.M{
			{"$match": m},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":     "$date",
						"channel":  "$channel",
						"channel1": "$channel1",
					},
					"player_total":           bson.M{"$sum": "$player_total"},
					"tp_game_number":         bson.M{"$sum": "$tp_game_number"},
					"tp_game_real_number":    bson.M{"$sum": "$tp_game_real_number"},
					"tp_number":              bson.M{"$sum": "$tp_number"},
					"tp_game_time":           bson.M{"$sum": "$tp_game_time"},
					"rummy_game_number":      bson.M{"$sum": "$rummy_game_number"},
					"rm_game_real_number":    bson.M{"$sum": "$rm_game_real_number"},
					"rummy_number":           bson.M{"$sum": "$rummy_number"},
					"rummy_game_time":        bson.M{"$sum": "$rummy_game_time"},
					"lhd_game_number":        bson.M{"$sum": "$lhd_game_number"},
					"lhd_number":             bson.M{"$sum": "$lhd_number"},
					"lhd_game_time":          bson.M{"$sum": "$lhd_game_time"},
					"up_game_number":         bson.M{"$sum": "$up_game_number"},
					"up_number":              bson.M{"$sum": "$up_number"},
					"up_game_time":           bson.M{"$sum": "$up_game_time"},
					"ak_game_number":         bson.M{"$sum": "$ak_game_number"},
					"ak_game_real_number":    bson.M{"$sum": "$ak_game_real_number"},
					"ak_number":              bson.M{"$sum": "$ak_number"},
					"ak_game_time":           bson.M{"$sum": "$ak_game_time"},
					"joker_game_number":      bson.M{"$sum": "$joker_game_number"},
					"joker_game_real_number": bson.M{"$sum": "$joker_game_real_number"},
					"joker_number":           bson.M{"$sum": "$joker_number"},
					"joker_game_time":        bson.M{"$sum": "$joker_game_time"},
					"crash_game_number":      bson.M{"$sum": "$crash_game_number"},
					"crash_number":           bson.M{"$sum": "$crash_number"},
					"crash_game_time":        bson.M{"$sum": "$crash_game_time"},
					"ab_game_number":         bson.M{"$sum": "$ab_game_number"},
					"ab_number":              bson.M{"$sum": "$ab_number"},
					"ab_game_time":           bson.M{"$sum": "$ab_game_time"},
					"cp_game_number":         bson.M{"$sum": "$cp_game_number"},
					"cp_number":              bson.M{"$sum": "$cp_number"},
					"cp_game_time":           bson.M{"$sum": "$cp_game_time"},
					"fj_game_number":         bson.M{"$sum": "$fj_game_number"},
					"fj_number":              bson.M{"$sum": "$fj_number"},
					"fj_game_time":           bson.M{"$sum": "$fj_game_time"},
					"rb_game_number":         bson.M{"$sum": "$rb_game_number"},
					"rb_number":              bson.M{"$sum": "$rb_number"},
					"rb_game_time":           bson.M{"$sum": "$rb_game_time"},
					"rm_two_game_number":     bson.M{"$sum": "$rm_two_game_number"},
					"rm_two_number":          bson.M{"$sum": "$rm_two_number"},
					"rm_two_game_time":       bson.M{"$sum": "$rm_two_game_time"},
					"tp2_game_number":        bson.M{"$sum": "$tp2_game_number"},
					"tp2_number":             bson.M{"$sum": "$tp2_number"},
					"tp2_game_time":          bson.M{"$sum": "$tp2_game_time"},
					"slots_game_number":      bson.M{"$sum": "$slots_game_number"},
					"slots_number":           bson.M{"$sum": "$slots_number"},
					"zrsx_game_number":       bson.M{"$sum": "$zrsx_game_number"},
					"zrsx_number":            bson.M{"$sum": "$zrsx_number"},
					"tp_battle_game":         bson.M{"$sum": "$tp_battle_game"},
					"rm_battle_game":         bson.M{"$sum": "$rm_battle_game"},
					"ab_battle_game":         bson.M{"$sum": "$ab_battle_game"},
				},
			},
			{
				"$project": bson.M{
					"_id":                    "$_id.date",
					"date":                   "$_id.date",
					"channel":                "$_id.channel",
					"channel1":               "$_id.channel1",
					"player_total":           "$player_total",
					"tp_game_number":         "$tp_game_number",
					"tp_game_real_number":    "$tp_game_real_number",
					"tp_number":              "$tp_number",
					"tp_game_time":           "$tp_game_time",
					"rummy_game_number":      "$rummy_game_number",
					"rm_game_real_number":    "$rm_game_real_number",
					"rummy_number":           "$rummy_number",
					"rummy_game_time":        "$rummy_game_time",
					"lhd_game_number":        "$lhd_game_number",
					"lhd_number":             "$lhd_number",
					"lhd_game_time":          "$lhd_game_time",
					"up_game_number":         "$up_game_number",
					"up_number":              "$up_number",
					"up_game_time":           "$up_game_time",
					"ak_game_number":         "$ak_game_number",
					"ak_game_real_number":    "$ak_game_real_number",
					"ak_number":              "$ak_number",
					"ak_game_time":           "$ak_game_time",
					"joker_game_number":      "$joker_game_number",
					"joker_game_real_number": "$joker_game_real_number",
					"joker_number":           "$joker_number",
					"joker_game_time":        "$joker_game_time",
					"crash_game_number":      "$crash_game_number",
					"crash_number":           "$crash_number",
					"crash_game_time":        "$crash_game_time",
					"ab_game_number":         "$ab_game_number",
					"ab_number":              "$ab_number",
					"ab_game_time":           "$ab_game_time",
					"cp_game_number":         "$cp_game_number",
					"cp_number":              "$cp_number",
					"cp_game_time":           "$cp_game_time",
					"fj_game_number":         "$fj_game_number",
					"fj_number":              "$fj_number",
					"fj_game_time":           "$fj_game_time",
					"rb_game_number":         "$rb_game_number",
					"rb_number":              "$rb_number",
					"rb_game_time":           "$rb_game_time",
					"rm_two_game_number":     "$rm_two_game_number",
					"rm_two_number":          "$rm_two_number",
					"rm_two_game_time":       "$rm_two_game_time",
					"tp2_game_number":        "$tp2_game_number",
					"tp2_number":             "$tp2_number",
					"tp2_game_time":          "$tp2_game_time",
					"slots_game_number":      "$slots_game_number",
					"slots_number":           "$slots_number",
					"zrsx_game_number":       "$zrsx_game_number",
					"zrsx_number":            "$zrsx_number",
					"tp_battle_game":         "$tp_battle_game",
					"rm_battle_game":         "$rm_battle_game",
					"ab_battle_game":         "$ab_battle_game",
				},
			},
			{"$sort": bson.M{"_id": -1}},
			{"$skip": skipNum},
			{"$limit": pageSize},
		}
	} else {
		// 全部
		pipeline = []bson.M{
			{"$match": m},
			{
				"$group": bson.M{
					"_id":                    "$date",
					"player_total":           bson.M{"$sum": "$player_total"},
					"tp_game_number":         bson.M{"$sum": "$tp_game_number"},
					"tp_game_real_number":    bson.M{"$sum": "$tp_game_real_number"},
					"tp_number":              bson.M{"$sum": "$tp_number"},
					"tp_game_time":           bson.M{"$sum": "$tp_game_time"},
					"rummy_game_number":      bson.M{"$sum": "$rummy_game_number"},
					"rm_game_real_number":    bson.M{"$sum": "$rm_game_real_number"},
					"rummy_number":           bson.M{"$sum": "$rummy_number"},
					"rummy_game_time":        bson.M{"$sum": "$rummy_game_time"},
					"lhd_game_number":        bson.M{"$sum": "$lhd_game_number"},
					"lhd_number":             bson.M{"$sum": "$lhd_number"},
					"lhd_game_time":          bson.M{"$sum": "$lhd_game_time"},
					"up_game_number":         bson.M{"$sum": "$up_game_number"},
					"up_number":              bson.M{"$sum": "$up_number"},
					"up_game_time":           bson.M{"$sum": "$up_game_time"},
					"ak_game_number":         bson.M{"$sum": "$ak_game_number"},
					"ak_game_real_number":    bson.M{"$sum": "$ak_game_real_number"},
					"ak_number":              bson.M{"$sum": "$ak_number"},
					"ak_game_time":           bson.M{"$sum": "$ak_game_time"},
					"joker_game_number":      bson.M{"$sum": "$joker_game_number"},
					"joker_game_real_number": bson.M{"$sum": "$joker_game_real_number"},
					"joker_number":           bson.M{"$sum": "$joker_number"},
					"joker_game_time":        bson.M{"$sum": "$joker_game_time"},
					"crash_game_number":      bson.M{"$sum": "$crash_game_number"},
					"crash_number":           bson.M{"$sum": "$crash_number"},
					"crash_game_time":        bson.M{"$sum": "$crash_game_time"},
					"ab_game_number":         bson.M{"$sum": "$ab_game_number"},
					"ab_number":              bson.M{"$sum": "$ab_number"},
					"ab_game_time":           bson.M{"$sum": "$ab_game_time"},
					"cp_game_number":         bson.M{"$sum": "$cp_game_number"},
					"cp_number":              bson.M{"$sum": "$cp_number"},
					"cp_game_time":           bson.M{"$sum": "$cp_game_time"},
					"fj_game_number":         bson.M{"$sum": "$fj_game_number"},
					"fj_number":              bson.M{"$sum": "$fj_number"},
					"fj_game_time":           bson.M{"$sum": "$fj_game_time"},
					"rb_game_number":         bson.M{"$sum": "$rb_game_number"},
					"rb_number":              bson.M{"$sum": "$rb_number"},
					"rb_game_time":           bson.M{"$sum": "$rb_game_time"},
					"rm_two_game_number":     bson.M{"$sum": "$rm_two_game_number"},
					"rm_two_number":          bson.M{"$sum": "$rm_two_number"},
					"rm_two_game_time":       bson.M{"$sum": "$rm_two_game_time"},
					"tp2_game_number":        bson.M{"$sum": "$tp2_game_number"},
					"tp2_number":             bson.M{"$sum": "$tp2_number"},
					"tp2_game_time":          bson.M{"$sum": "$tp2_game_time"},
					"slots_game_number":      bson.M{"$sum": "$slots_game_number"},
					"slots_number":           bson.M{"$sum": "$slots_number"},
					"zrsx_game_number":       bson.M{"$sum": "$zrsx_game_number"},
					"zrsx_number":            bson.M{"$sum": "$zrsx_number"},
					"tp_battle_game":         bson.M{"$sum": "$tp_battle_game"},
					"rm_battle_game":         bson.M{"$sum": "$rm_battle_game"},
					"ab_battle_game":         bson.M{"$sum": "$ab_battle_game"},
				},
			},
			{
				"$project": bson.M{
					"_id":                    "$_id",
					"date":                   "$_id",
					"player_total":           "$player_total",
					"tp_game_number":         "$tp_game_number",
					"tp_game_real_number":    "$tp_game_real_number",
					"tp_number":              "$tp_number",
					"tp_game_time":           "$tp_game_time",
					"rummy_game_number":      "$rummy_game_number",
					"rm_game_real_number":    "$rm_game_real_number",
					"rummy_number":           "$rummy_number",
					"rummy_game_time":        "$rummy_game_time",
					"lhd_game_number":        "$lhd_game_number",
					"lhd_number":             "$lhd_number",
					"lhd_game_time":          "$lhd_game_time",
					"up_game_number":         "$up_game_number",
					"up_number":              "$up_number",
					"up_game_time":           "$up_game_time",
					"ak_game_number":         "$ak_game_number",
					"ak_game_real_number":    "$ak_game_real_number",
					"ak_number":              "$ak_number",
					"ak_game_time":           "$ak_game_time",
					"joker_game_number":      "$joker_game_number",
					"joker_game_real_number": "$joker_game_real_number",
					"joker_number":           "$joker_number",
					"joker_game_time":        "$joker_game_time",
					"crash_game_number":      "$crash_game_number",
					"crash_number":           "$crash_number",
					"crash_game_time":        "$crash_game_time",
					"ab_game_number":         "$ab_game_number",
					"ab_number":              "$ab_number",
					"ab_game_time":           "$ab_game_time",
					"cp_game_number":         "$cp_game_number",
					"cp_number":              "$cp_number",
					"cp_game_time":           "$cp_game_time",
					"fj_game_number":         "$fj_game_number",
					"fj_number":              "$fj_number",
					"fj_game_time":           "$fj_game_time",
					"rb_game_number":         "$rb_game_number",
					"rb_number":              "$rb_number",
					"rb_game_time":           "$rb_game_time",
					"rm_two_game_number":     "$rm_two_game_number",
					"rm_two_number":          "$rm_two_number",
					"rm_two_game_time":       "$rm_two_game_time",
					"tp2_game_number":        "$tp2_game_number",
					"tp2_number":             "$tp2_number",
					"tp2_game_time":          "$tp2_game_time",
					"slots_game_number":      "$slots_game_number",
					"slots_number":           "$slots_number",
					"zrsx_game_number":       "$zrsx_game_number",
					"zrsx_number":            "$zrsx_number",
					"tp_battle_game":         "$tp_battle_game",
					"rm_battle_game":         "$rm_battle_game",
					"ab_battle_game":         "$ab_battle_game",
				},
			},
			{"$sort": bson.M{"_id": -1}},
			{"$skip": skipNum},
			{"$limit": pageSize},
		}
	}

	err := RoomDatas.Pipe(pipeline).All(&list)
	list = this.chipList16(list)
	return list, err
}

// 转换为分展示
func (this *statisticsService) chipList16(list []entity.RoomDataList) []entity.RoomDataList {
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		// 计算单局时长、人均局数
		v.TPPerCapita = ComputeFloat(v.TPGameNumber, v.TPNumber)
		v.TPTime = ComputeFloat(v.TPGameTime, v.TPGameNumber)
		v.LHDPerCapita = ComputeFloat(v.LHDGameNumber, v.LHDNumber)
		v.LHDTime = ComputeFloat(v.LHDGameTime, v.LHDGameNumber)
		v.UPPerCapita = ComputeFloat(v.UPGameNumber, v.UPNumber)
		v.UPTime = ComputeFloat(v.UPGameTime, v.UPGameNumber)
		v.RummyPerCapita = ComputeFloat(v.RummyGameNumber, v.RummyNumber)
		v.RummyTime = ComputeFloat(v.RummyGameTime, v.RummyGameNumber)
		v.AKPerCapita = ComputeFloat(v.AKGameNumber, v.AKNumber)
		v.AKTime = ComputeFloat(v.AKGameTime, v.AKGameNumber)
		v.JokerPerCapita = ComputeFloat(v.JokerGameNumber, v.JokerNumber)
		v.JokerTime = ComputeFloat(v.JokerGameTime, v.JokerGameNumber)
		v.CrashPerCapita = ComputeFloat(v.CrashGameNumber, v.CrashNumber)
		v.CrashTime = ComputeFloat(v.CrashGameTime, v.CrashGameNumber)
		v.ABPerCapita = ComputeFloat(v.ABGameNumber, v.ABNumber)
		v.ABTime = ComputeFloat(v.ABGameTime, v.ABGameNumber)
		v.CPPerCapita = ComputeFloat(v.CPGameNumber, v.CPNumber)
		v.CPTime = ComputeFloat(v.CPGameTime, v.CPGameNumber)
		v.FJPerCapita = ComputeFloat(v.FJGameNumber, v.FJNumber)
		v.FJTime = ComputeFloat(v.FJGameTime, v.FJGameNumber)
		v.RBPerCapita = ComputeFloat(v.RBGameNumber, v.RBNumber)
		v.RBTime = ComputeFloat(v.RBGameTime, v.RBGameNumber)
		v.RMTwoPerCapita = ComputeFloat(v.RMTwoGameNumber, v.RMTwoNumber)
		v.RMTwoTime = ComputeFloat(v.RMTwoGameTime, v.RMTwoGameNumber)
		v.TP2PerCapita = ComputeFloat(v.TP2GameNumber, v.TP2Number)
		v.TP2Time = ComputeFloat(v.TP2GameTime, v.TP2GameNumber)
		v.SlotsPerCapita = ComputeFloat(v.SlotsGameNumber, v.SlotsNumber)
		v.ZRSXPerCapita = ComputeFloat(v.ZRSXGameNumber, v.ZRSXNumber)
		list[k] = v
	}
	return list
}

// 查询房间数据条数
func (this *statisticsService) GetRoomDataTotal(m bson.M, isChannel bool) (int64, error) {
	pipeline := []bson.M{}
	if isChannel {
		// 按渠道
		pipeline = []bson.M{
			{
				"$match": m,
			},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":     "$date",
						"channel":  "$channel",
						"channel1": "$channel1",
					},
				},
			},
		}
	} else {
		// 全部
		pipeline = []bson.M{
			{
				"$match": m,
			},
			{
				"$group": bson.M{
					"_id": "$date",
				},
			},
		}
	}
	result := []bson.M{}
	pipe := RoomDatas.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("RoomData fail err: ", err)
	}
	return int64(len(result)), nil
	// return int64(Count(RoomDatas, m)), nil
}

// 新增房间数据
func (this *statisticsService) AddRoomData(rate *entity.RoomData) error {
	info := new(entity.RoomData)
	GetByQ(RoomDatas, bson.M{"date": rate.Date, "channel": rate.Channel, "player_types": rate.PlayerTypes, "number_types": rate.NumberTypes}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"channel1":               rate.Channel1,
			"player_total":           rate.PlayerTotal,
			"tp_game_number":         rate.TPGameNumber,
			"tp_game_real_number":    rate.TPGameRealNumber,
			"tp_number":              rate.TPNumber,
			"tp_game_time":           rate.TPGameTime,
			"rummy_game_number":      rate.RummyGameNumber,
			"rm_game_real_number":    rate.RMGameRealNumber,
			"rummy_number":           rate.RummyNumber,
			"rummy_game_time":        rate.RummyGameTime,
			"lhd_game_number":        rate.LHDGameNumber,
			"lhd_number":             rate.LHDNumber,
			"lhd_game_time":          rate.LHDGameTime,
			"up_game_number":         rate.UPGameNumber,
			"up_number":              rate.UPNumber,
			"up_game_time":           rate.UPGameTime,
			"ak_game_number":         rate.AKGameNumber,
			"ak_game_real_number":    rate.AKGameRealNumber,
			"ak_number":              rate.AKNumber,
			"ak_game_time":           rate.AKGameTime,
			"joker_game_number":      rate.JokerGameNumber,
			"joker_game_real_number": rate.JokerGameRealNumber,
			"joker_number":           rate.JokerNumber,
			"joker_game_time":        rate.JokerGameTime,
			"crash_game_number":      rate.CrashGameNumber,
			"crash_number":           rate.CrashNumber,
			"crash_game_time":        rate.CrashGameTime,
			"ab_game_number":         rate.ABGameNumber,
			"ab_number":              rate.ABNumber,
			"ab_game_time":           rate.ABGameTime,
			"cp_game_number":         rate.CPGameNumber,
			"cp_number":              rate.CPNumber,
			"cp_game_time":           rate.CPGameTime,
			"fj_game_number":         rate.FJGameNumber,
			"fj_number":              rate.FJNumber,
			"fj_game_time":           rate.FJGameTime,
			"slots_game_number":      rate.SlotsGameNumber,
			"slots_number":           rate.SlotsNumber,
			"zrsx_game_number":       rate.ZRSXGameNumber,
			"zrsx_number":            rate.ZRSXNumber,
			"tp_battle_game":         rate.TPBattleGame,
			"rm_battle_game":         rate.RMBattleGame,
			"ab_battle_game":         rate.ABBattleGame,
		}
		if Update(RoomDatas, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + rate.Id)
	} else {
		// 新增
		rate.Id = bson.NewObjectId().Hex()
		if !Insert(RoomDatas, rate) {
			return errors.New("写入失败:" + rate.Id)
		}
		return nil
	}
}

/*
商品购买统计相关接口
*/
func (this *statisticsService) GetGoodsBuyList(page, pageSize int, m bson.M) ([]entity.GoodsBuyData, error) {
	var list []entity.GoodsBuyData
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)
	pipeline := []bson.M{
		{"$match": m},
		{
			"$group": bson.M{
				"_id": bson.M{
					"date":      "$date",
					"shop_type": "$shop_type",
					"shop_name": "$shop_name",
					"amount":    "$amount",
				},
				"total_pull_order": bson.M{"$sum": "$total_pull_order"},
				"successful_order": bson.M{"$sum": "$successful_order"},
				"pull_number":      bson.M{"$sum": "$pull_number"},
				"buy_number":       bson.M{"$sum": "$buy_number"},
			},
		},
		{
			"$project": bson.M{
				"_id":              "$_id.date",
				"date":             "$_id.date",
				"shop_type":        "$_id.shop_type",
				"shop_name":        "$_id.shop_name",
				"amount":           "$_id.amount",
				"total_pull_order": "$total_pull_order",
				"successful_order": "$successful_order",
				"pull_number":      "$pull_number",
				"buy_number":       "$buy_number",
			},
		},
		{"$sort": bson.M{"date": -1, "shop_type": 1, "amount": 1}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}
	err := GoodsBuys.Pipe(pipeline).All(&list)
	list = this.chipList18(list)
	return list, err
}

// 转换为分展示
func (this *statisticsService) chipList18(list []entity.GoodsBuyData) []entity.GoodsBuyData {
	// 初始化总和变量
	DataTotals := make(map[int64]int64)
	Datas := make([]int64, 0)
	// 循环遍历切片并累加字段值
	for _, item := range list {
		Datas = append(Datas, item.Date)
	}
	pipeline := []bson.M{
		{"$match": bson.M{
			"date": bson.M{"$in": Datas},
		},
		},
		{
			"$group": bson.M{
				"_id":              "$date",
				"successful_order": bson.M{"$sum": "$successful_order"},
			},
		},
	}
	var result []bson.M
	GoodsBuys.Pipe(pipeline).All(&result)
	if len(result) > 0 {
		for _, g := range result {
			gtime := g["_id"].(int64)
			s_order := g["successful_order"].(int64)
			DataTotals[gtime] = s_order
		}
	}
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		v.SuccessfulOrderRatio = ComputeFloat(v.SuccessfulOrder, v.TotalPullOrder) * 100
		v.BuyNumberRatio = ComputeFloat(v.BuyNumber, v.PullNumber) * 100

		for day, item := range DataTotals {
			if day == v.Date {
				v.ZBRatio = ComputeFloat(v.SuccessfulOrder, item) * 100
			}
		}
		v.BuyNumberRatio = ComputeFloat(v.BuyNumber, v.PullNumber) * 100
		v.FAmount = ComputeFloat(int64(v.Amount), 100)
		list[k] = v
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Date > list[j].Date
	})
	return list
}

// 查询商品购买条数
func (this *statisticsService) GetGoodsBuyTotal(m bson.M) (int64, error) {
	pipeline := []bson.M{
		{
			"$match": m,
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"date":      "$date",
					"shop_type": "$shop_type",
					"shop_name": "$shop_name",
					"amount":    "$amount",
				},
			},
		},
	}
	result := []bson.M{}
	pipe := GoodsBuys.Pipe(pipeline)
	err := pipe.All(&result)
	return int64(len(result)), err
	// return int64(Count(GoodsBuys, m)), nil
}

// 新增商品购买
func (this *statisticsService) AddGoodsBuy(rate *entity.GoodsBuyData) error {
	info := new(entity.GoodsBuyData)
	GetByQ(GoodsBuys, bson.M{"date": rate.Date, "shop_type": rate.ShopType, "player_types": rate.PlayerTypes, "shop_name": rate.ShopName, "amount": rate.Amount}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"total_pull_order": rate.TotalPullOrder,
			"successful_order": rate.SuccessfulOrder,
			"pull_number":      rate.PullNumber,
			"buy_number":       rate.BuyNumber,
		}
		if Update(GoodsBuys, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + rate.Id)
	} else {
		// 新增
		rate.Id = bson.NewObjectId().Hex()
		if !Insert(GoodsBuys, rate) {
			return errors.New("写入失败:" + rate.Id)
		}
		return nil
	}
}

/*
数据汇总（新）
*/
func (this *statisticsService) GetDataList(page, pageSize int, m bson.M, isChannel bool) ([]*entity.DataStatistics, error) {
	var list []*entity.DataStatistics
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)
	pipeline := []bson.M{}
	if isChannel {
		// 渠道
		pipeline = []bson.M{
			{"$match": m},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":     "$date",
						"channel":  "$channel",
						"channel1": "$channel1",
					},
					"old_recharge_num":       bson.M{"$sum": "$old_recharge_num"},
					"old_recharge_amount":    bson.M{"$sum": "$old_recharge_amount"},
					"old_login_num":          bson.M{"$sum": "$old_login_num"},
					"new_register":           bson.M{"$sum": "$new_register"},
					"new_equipment":          bson.M{"$sum": "$new_equipment"},
					"login_num":              bson.M{"$sum": "$login_num"},
					"next_new_register":      bson.M{"$sum": "$next_new_register"},
					"next_login":             bson.M{"$sum": "$next_login"},
					"jr_login_next_pay":      bson.M{"$sum": "$jr_login_next_pay"},
					"zr_pay_count":           bson.M{"$sum": "$zr_pay_count"},
					"new_recharge_num":       bson.M{"$sum": "$new_recharge_num"},
					"new_recharge_amount":    bson.M{"$sum": "$new_recharge_amount"},
					"total_recharge":         bson.M{"$sum": "$total_recharge"},
					"total_amount":           bson.M{"$sum": "$total_amount"},
					"pay_request":            bson.M{"$sum": "$pay_request"},
					"pay_request_order":      bson.M{"$sum": "$pay_request_order"},
					"pay_success_order":      bson.M{"$sum": "$pay_success_order"},
					"withdraw_request":       bson.M{"$sum": "$withdraw_request"},
					"withdraw_request_order": bson.M{"$sum": "$withdraw_request_order"},
					"withdraw_success_order": bson.M{"$sum": "$withdraw_success_order"},
					"total_withdraw_num":     bson.M{"$sum": "$total_withdraw_num"},
					"total_withdraw_amount":  bson.M{"$sum": "$total_withdraw_amount"},
					"handling_charge":        bson.M{"$sum": "$handling_charge"},
					"old_recharge_num2":      bson.M{"$sum": "$old_recharge_num2"},
					"new_recharge_num2":      bson.M{"$sum": "$new_recharge_num2"},
					"total_recharge_num2":    bson.M{"$sum": "$total_recharge_num2"},
					"old_pay_success_order":  bson.M{"$sum": "$old_pay_success_order"},
					"new_pay_success_order":  bson.M{"$sum": "$new_pay_success_order"},
					"pay_login_number":       bson.M{"$sum": "$pay_login_number"},
					"pay_minutes30":          bson.M{"$sum": "$pay_minutes30"},
					"pay_minutes60":          bson.M{"$sum": "$pay_minutes60"},
					"pay_minutes120":         bson.M{"$sum": "$pay_minutes120"},
					"trust_user_pay":         bson.M{"$sum": "$trust_user_pay"},
					"trust_user_withdraw":    bson.M{"$sum": "$trust_user_withdraw"},
				},
			},
			{
				"$project": bson.M{
					"_id":                    "$_id.date",
					"date":                   "$_id.date",
					"channel":                "$_id.channel",
					"channel1":               "$_id.channel1",
					"old_recharge_num":       "$old_recharge_num",
					"old_recharge_amount":    "$old_recharge_amount",
					"old_login_num":          "$old_login_num",
					"new_register":           "$new_register",
					"new_equipment":          "$new_equipment",
					"login_num":              "$login_num",
					"next_new_register":      "$next_new_register",
					"next_login":             "$next_login",
					"jr_login_next_pay":      "$jr_login_next_pay",
					"zr_pay_count":           "$zr_pay_count",
					"new_recharge_num":       "$new_recharge_num",
					"new_recharge_amount":    "$new_recharge_amount",
					"total_recharge":         "$total_recharge",
					"total_amount":           "$total_amount",
					"pay_request":            "$pay_request",
					"pay_request_order":      "$pay_request_order",
					"pay_success_order":      "$pay_success_order",
					"withdraw_request":       "$withdraw_request",
					"withdraw_request_order": "$withdraw_request_order",
					"withdraw_success_order": "$withdraw_success_order",
					"total_withdraw_num":     "$total_withdraw_num",
					"total_withdraw_amount":  "$total_withdraw_amount",
					"handling_charge":        "$handling_charge",
					"old_recharge_num2":      "$old_recharge_num2",
					"new_recharge_num2":      "$new_recharge_num2",
					"total_recharge_num2":    "$total_recharge_num2",
					"old_pay_success_order":  "$old_pay_success_order",
					"new_pay_success_order":  "$new_pay_success_order",
					"pay_login_number":       "$pay_login_number",
					"pay_minutes30":          "$pay_minutes30",
					"pay_minutes60":          "$pay_minutes60",
					"pay_minutes120":         "$pay_minutes120",
					"trust_user_pay":         "$trust_user_pay",
					"trust_user_withdraw":    "$trust_user_withdraw",
				},
			},
			{"$sort": bson.M{"date": -1}},
			{"$skip": skipNum},
			{"$limit": pageSize},
		}
	} else {
		// 全部
		pipeline = []bson.M{
			{"$match": m},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date": "$date",
					},
					"old_recharge_num":       bson.M{"$sum": "$old_recharge_num"},
					"old_recharge_amount":    bson.M{"$sum": "$old_recharge_amount"},
					"old_login_num":          bson.M{"$sum": "$old_login_num"},
					"new_register":           bson.M{"$sum": "$new_register"},
					"new_equipment":          bson.M{"$sum": "$new_equipment"},
					"login_num":              bson.M{"$sum": "$login_num"},
					"next_new_register":      bson.M{"$sum": "$next_new_register"},
					"next_login":             bson.M{"$sum": "$next_login"},
					"jr_login_next_pay":      bson.M{"$sum": "$jr_login_next_pay"},
					"zr_pay_count":           bson.M{"$sum": "$zr_pay_count"},
					"new_recharge_num":       bson.M{"$sum": "$new_recharge_num"},
					"new_recharge_amount":    bson.M{"$sum": "$new_recharge_amount"},
					"total_recharge":         bson.M{"$sum": "$total_recharge"},
					"total_amount":           bson.M{"$sum": "$total_amount"},
					"pay_request":            bson.M{"$sum": "$pay_request"},
					"pay_request_order":      bson.M{"$sum": "$pay_request_order"},
					"pay_success_order":      bson.M{"$sum": "$pay_success_order"},
					"withdraw_request":       bson.M{"$sum": "$withdraw_request"},
					"withdraw_request_order": bson.M{"$sum": "$withdraw_request_order"},
					"withdraw_success_order": bson.M{"$sum": "$withdraw_success_order"},
					"total_withdraw_num":     bson.M{"$sum": "$total_withdraw_num"},
					"total_withdraw_amount":  bson.M{"$sum": "$total_withdraw_amount"},
					"handling_charge":        bson.M{"$sum": "$handling_charge"},
					"old_recharge_num2":      bson.M{"$sum": "$old_recharge_num2"},
					"new_recharge_num2":      bson.M{"$sum": "$new_recharge_num2"},
					"total_recharge_num2":    bson.M{"$sum": "$total_recharge_num2"},
					"old_pay_success_order":  bson.M{"$sum": "$old_pay_success_order"},
					"new_pay_success_order":  bson.M{"$sum": "$new_pay_success_order"},
					"pay_login_number":       bson.M{"$sum": "$pay_login_number"},
					"pay_minutes30":          bson.M{"$sum": "$pay_minutes30"},
					"pay_minutes60":          bson.M{"$sum": "$pay_minutes60"},
					"pay_minutes120":         bson.M{"$sum": "$pay_minutes120"},
					"trust_user_pay":         bson.M{"$sum": "$trust_user_pay"},
					"trust_user_withdraw":    bson.M{"$sum": "$trust_user_withdraw"},
				},
			},
			{
				"$project": bson.M{
					"_id":                    "$_id.date",
					"date":                   "$_id.date",
					"old_recharge_num":       "$old_recharge_num",
					"old_recharge_amount":    "$old_recharge_amount",
					"old_login_num":          "$old_login_num",
					"new_register":           "$new_register",
					"new_equipment":          "$new_equipment",
					"login_num":              "$login_num",
					"next_new_register":      "$next_new_register",
					"next_login":             "$next_login",
					"jr_login_next_pay":      "$jr_login_next_pay",
					"zr_pay_count":           "$zr_pay_count",
					"new_recharge_num":       "$new_recharge_num",
					"new_recharge_amount":    "$new_recharge_amount",
					"total_recharge":         "$total_recharge",
					"total_amount":           "$total_amount",
					"pay_request":            "$pay_request",
					"pay_request_order":      "$pay_request_order",
					"pay_success_order":      "$pay_success_order",
					"withdraw_request":       "$withdraw_request",
					"withdraw_request_order": "$withdraw_request_order",
					"withdraw_success_order": "$withdraw_success_order",
					"total_withdraw_num":     "$total_withdraw_num",
					"total_withdraw_amount":  "$total_withdraw_amount",
					"handling_charge":        "$handling_charge",
					"old_recharge_num2":      "$old_recharge_num2",
					"new_recharge_num2":      "$new_recharge_num2",
					"total_recharge_num2":    "$total_recharge_num2",
					"old_pay_success_order":  "$old_pay_success_order",
					"new_pay_success_order":  "$new_pay_success_order",
					"pay_login_number":       "$pay_login_number",
					"pay_minutes30":          "$pay_minutes30",
					"pay_minutes60":          "$pay_minutes60",
					"pay_minutes120":         "$pay_minutes120",
					"trust_user_pay":         "$trust_user_pay",
					"trust_user_withdraw":    "$trust_user_withdraw",
				},
			},
			{"$sort": bson.M{"date": -1}},
			{"$skip": skipNum},
			{"$limit": pageSize},
		}
	}

	err := DataStatisticss.Pipe(pipeline).All(&list)
	list = this.chipList19(list)
	// 查询实时数据
	list, err = this.dataListCurrent(m, list, isChannel)
	return list, err
}

// 统计实时数据
func (this *statisticsService) dataListCurrent(m bson.M, list []*entity.DataStatistics, isChannel bool) ([]*entity.DataStatistics, error) {
	var dateStatistics = make(map[string]*entity.DataStatistics)
	var dateChannelStatistics = make(map[string]map[string]*entity.DataStatistics)
	for _, data := range list {
		data.FTotalAmount = 0
		data.FTotalWithdrawAmount = 0
		data.TotalWithdrawAmount = 0
		data.FTotalWithdrawAmount = 0
		data.HandlingCharge = 0
		data.FHandlingCharge = 0
		data.CostRatio = 0
		data.TotalProfit = 0
		data.FTotalProfit = 0
		dateStatistics[data.FDate] = data

		if isChannel {
			cs, ok := dateChannelStatistics[data.FDate]
			if !ok {
				cs = make(map[string]*entity.DataStatistics)
				dateChannelStatistics[data.FDate] = cs
			}
			cs[data.Channel] = data
		}
	}

	loc, _ := time.LoadLocation(locationName)
	var where0 string
	var args0 []any
	var regist_area []int32

	if date, ok := m["date"]; ok {
		dateM := date.(bson.M)
		if start, ok := dateM["$gte"]; ok {
			startTime := time.Unix(start.(int64), 0).In(loc)
			where0 += " and ctime >= ?"
			args0 = append(args0, startTime)
		}
		if end, ok := dateM["$lte"]; ok {
			endTime := time.Unix(end.(int64), 0).In(loc)
			where0 += " and ctime <= ?"
			args0 = append(args0, endTime)
		}
	}
	if data_types, ok := m["data_types"]; ok {
		data_types, ok := data_types.(bson.M)
		if ok {
			regist_area = []int32{0, 3}
		} else {
			regist_area = []int32{int32(utils.ToInt64(data_types))}
		}
	}
	var packageIds []string
	// 渠道
	if channel, ok := m["channel"]; ok {
		if channelM, ok := channel.(bson.M); ok {
			if channels, ok := channelM["$in"]; ok {
				packageIds = append(packageIds, channels.([]string)...)
			}
		} else {
			packageIds = append(packageIds, channel.(string))
		}
	}
	// 渠道别名
	if channel1, ok := m["channel1"]; ok {
		var channelAlias []string
		if channel1M, ok := channel1.(bson.M); ok {
			if channels1, ok := channel1M["$in"]; ok {
				channelAlias = append(channelAlias, channels1.([]string)...)
			}
		} else {
			channelAlias = append(channelAlias, channel1.(string))
		}
		if len(channelAlias) > 0 {
			channels, err := ChannelService.GetChannelByNameAliasList(channelAlias)
			if err == nil {
				for _, c := range channels {
					packageIds = append(packageIds, c.Name)
				}
			}
		}
	}
	if len(packageIds) > 0 {
		where0 += " and package_id in ?"
		args0 = append(args0, packageIds)
	}

	var uFilterPay string
	var uFilterPayArgs []any
	if len(regist_area) > 0 {
		uFilterPay = fmt.Sprintf(`
			and userid in (
				select userid from game.col_user final where regist_area in ? and userid in (
					SELECT DISTINCT userid FROM game.col_trade_record FINAL 
					WHERE order_status = 4 %s
				)
			)
		`, where0)
		uFilterPayArgs = append(uFilterPayArgs, regist_area)
		uFilterPayArgs = append(uFilterPayArgs, args0...)
	}
	sqlPay := `
		SELECT toYYYYMMDD(ctime) cday, SUM(amount) amount_sum
		FROM game.col_trade_record FINAL WHERE order_status = 4 %s %s
		GROUP BY cday
	`
	if isChannel {
		sqlPay = `
			SELECT toYYYYMMDD(ctime) cday, package_id, SUM(amount) amount_sum
			FROM game.col_trade_record FINAL WHERE order_status = 4 %s %s
			GROUP BY cday, package_id
		`
	}
	var datasPay []map[string]any
	err := ck.Select(&datasPay, fmt.Sprintf(sqlPay, where0, uFilterPay), append(args0, uFilterPayArgs...)...)
	if err != nil {
		return list, err
	}
	for _, pay := range datasPay {
		cday := fmt.Sprint(pay["cday"])
		sdate := cday[0:4] + "-" + cday[4:6] + "-" + cday[6:]
		if !isChannel {
			if data, ok := dateStatistics[sdate]; ok {
				data.TotalAmount = utils.ToInt64(pay["amount_sum"])
				data.FTotalAmount = Chip2Float(data.TotalAmount)
			}
		} else {
			if cs, ok := dateChannelStatistics[sdate]; ok {
				package_id := fmt.Sprint(pay["package_id"])
				if data, ok := cs[package_id]; ok {
					data.TotalAmount = utils.ToInt64(pay["amount_sum"])
					data.FTotalAmount = Chip2Float(data.TotalAmount)
				}
			}
		}
	}

	// 提现实时计算
	var uFilterW string
	var uFilterWArgs []any
	if len(regist_area) > 0 {
		uFilterW = fmt.Sprintf(`
			and userid in (
				select userid from game.col_user final where regist_area in ? and userid in (
					SELECT DISTINCT userid FROM game.col_withdraw_record FINAL 
					WHERE order_status = 2 %s
				)
			)
		`, where0)
		uFilterWArgs = append(uFilterWArgs, regist_area)
		uFilterWArgs = append(uFilterWArgs, args0...)
	}
	sqlW := `
		SELECT toYYYYMMDD(ctime) cday, SUM(amount) amount_sum, SUM(commission) commission_sum
		FROM game.col_withdraw_record FINAL WHERE order_status = 2 %s %s
		GROUP BY cday
	`
	if isChannel {
		sqlW = `
			SELECT toYYYYMMDD(ctime) cday, package_id, SUM(amount) amount_sum, SUM(commission) commission_sum
			FROM game.col_withdraw_record FINAL WHERE order_status = 2 %s %s
			GROUP BY cday, package_id
		`
	}
	var datasW []map[string]any
	err = ck.Select(&datasW, fmt.Sprintf(sqlW, where0, uFilterW), append(args0, uFilterWArgs...)...)
	if err != nil {
		return list, err
	}
	for _, w := range datasW {
		cday := fmt.Sprint(w["cday"])
		sdate := cday[0:4] + "-" + cday[4:6] + "-" + cday[6:]
		var data *entity.DataStatistics
		if !isChannel {
			data = dateStatistics[sdate]
		} else {
			if cs, ok := dateChannelStatistics[sdate]; ok {
				package_id := fmt.Sprint(w["package_id"])
				data = cs[package_id]
			}
		}
		if data != nil {
			v := data
			v.TotalWithdrawAmount = utils.ToInt64(w["amount_sum"])
			v.FTotalWithdrawAmount = Chip2Float(v.TotalWithdrawAmount)
			v.HandlingCharge = utils.ToInt64(w["commission_sum"])
			v.FHandlingCharge = Chip2Float(v.HandlingCharge)
		}
	}
	// 计算结果
	for _, v := range list {
		v.CostRatio = ComputeFloat(v.TotalWithdrawAmount, v.TotalAmount) * 100
		v.TotalProfit = v.TotalAmount - v.TotalWithdrawAmount - v.HandlingCharge
		v.FTotalProfit = ComputeFloat(v.TotalProfit, 100)
	}
	return list, nil
}

// 转换为分展示
func (this *statisticsService) chipList19(list []*entity.DataStatistics) []*entity.DataStatistics {
	// 计算渠道总充额占比、新充额、老充额、日活占比
	select_date := make([]int64, 0)
	for _, d := range list {
		select_date = append(select_date, d.Date)
	}
	// 根据日期查询当日所有
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"date": bson.M{"$in": select_date},
			},
		},
		{
			"$group": bson.M{
				"_id":                 "$date",
				"login_num":           bson.M{"$sum": "$login_num"},
				"new_recharge_amount": bson.M{"$sum": "$new_recharge_amount"},
				"old_recharge_amount": bson.M{"$sum": "$old_recharge_amount"},
				"total_amount":        bson.M{"$sum": "$total_amount"},
			},
		},
	}
	result := []bson.M{}
	err := DataStatisticss.Pipe(pipeline).All(&result)
	fmt.Print(err)
	date_map := make(map[int64]entity.DateTotalInfo, 0)
	if len(result) > 0 {
		for _, r := range result {
			dt := r["_id"].(int64)
			l_num := r["login_num"].(int64)
			var n_amount int64
			if v, ok := r["new_recharge_amount"].(int64); ok {
				n_amount = v
			} else if v64, ok := r["new_recharge_amount"].(int); ok {
				n_amount = int64(v64)
			}
			var o_amount int64
			if v, ok := r["old_recharge_amount"].(int64); ok {
				o_amount = v
			} else if v64, ok := r["old_recharge_amount"].(int); ok {
				o_amount = int64(v64)
			}
			t_amount := r["total_amount"].(int64)
			info := new(entity.DateTotalInfo)
			info.LoginNum = l_num
			info.NewAmount = n_amount
			info.TotalAmount = t_amount
			info.OldAmount = o_amount
			info.Date = dt
			date_map[dt] = *info
		}
	}

	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		v.FDate = v.SDate.Format(utils.FORMAT_DATE)
		v.NextDayRetention = ComputeFloat(v.NextLogin, v.NextNewRegister) * 100
		v.PayRetention = ComputeFloat(v.JRLoginNextPay, v.ZRPayCount) * 100
		FRechargeAmount := ComputeFloat(v.NewRechargeAmount, 100)
		v.NewRechargeArpu = ComputeFloat(int64(FRechargeAmount), v.NewRegister)     //ComputeFloat(FTotalAmount, v.NewRegister) * 100
		v.NewRechargeArppu = ComputeFloat(int64(FRechargeAmount), v.NewRechargeNum) // ComputeFloat(v.NewRechargeAmount, v.NewRechargeNum) * 100
		v.NewUserPaymentRate = ComputeFloat(v.NewRechargeNum, v.NewRegister) * 100
		FOldRechargeAmount := ComputeFloat(v.OldRechargeAmount, 100)
		v.OldRechargeArpu = ComputeFloat(int64(FOldRechargeAmount), v.OldLoginNum)     //ComputeFloat(FTotalAmount, v.NewRegister) * 100
		v.OldRechargeArppu = ComputeFloat(int64(FOldRechargeAmount), v.OldRechargeNum) // ComputeFloat(v.NewRechargeAmount, v.NewRechargeNum) * 100
		v.OldUserPaymentRate = ComputeFloat(v.OldRechargeNum, v.OldLoginNum) * 100
		FTotalAmount := ComputeFloat(v.TotalAmount, 100)
		v.TotalRechargeArpu = ComputeFloat(int64(FTotalAmount), v.LoginNum)       // ComputeFloat(v.TotalAmount, v.LoginNum) * 100
		v.TotalRechargeArppu = ComputeFloat(int64(FTotalAmount), v.TotalRecharge) // ComputeFloat(v.TotalAmount, v.TotalRecharge) * 100
		v.TotalPaymentRate = ComputeFloat(v.TotalRecharge, v.LoginNum) * 100
		v.PaySuccessRate = ComputeFloat(v.PaySuccessOrder, v.PayRequestOrder) * 100
		v.WithdrawSuccessRate = ComputeFloat(v.WithdrawSuccessOrder, v.WithdrawRequestOrder) * 100
		v.CostRatio = ComputeFloat(v.TotalWithdrawAmount, v.TotalAmount) * 100
		v.FNewRechargeAmount = FRechargeAmount
		v.FOldRechargeAmount = FOldRechargeAmount
		v.FTotalAmount = FTotalAmount
		v.FTotalWithdrawAmount = ComputeFloat(v.TotalWithdrawAmount, 100)
		v.FHandlingCharge = ComputeFloat(v.HandlingCharge, 100)
		v.OldRepayRate = ComputeFloat(v.OldRechargeNum2, v.OldRechargeNum) * 100.0
		v.NewRepayRate = ComputeFloat(v.NewRechargeNum2, v.NewRechargeNum) * 100.0
		v.TotalRepayRate = ComputeFloat(v.TotalRechargeNum2, v.TotalRecharge) * 100.0
		v.OldPayRate = ComputeFloat(v.OldPaySuccessOrder, v.OldRechargeNum)
		v.NewPayRate = ComputeFloat(v.NewPaySuccessOrder, v.NewRechargeNum)
		v.TotalPayRate = ComputeFloat(v.PaySuccessOrder, v.TotalRecharge)
		v.TotalProfit = v.TotalAmount - v.TotalWithdrawAmount - v.HandlingCharge
		v.FTotalProfit = ComputeFloat(v.TotalProfit, 100)
		// v.DynamicProfitAvg = ComputeFloat(v.TotalProfit, v.LoginNum)
		// v.PayDynamicProfitAvg = ComputeFloat(v.TotalProfit, v.PayLoginNumber)
		if v.LoginNum > 0 {
			v.DynamicProfitAvg = Chip2Float(v.TotalProfit) / float64(v.LoginNum)
		}
		if v.PayLoginNumber > 0 {
			v.PayDynamicProfitAvg = Chip2Float(v.TotalProfit) / float64(v.PayLoginNumber)
		}

		v.PayMinutesRate30 = ComputeFloat(v.PayMinutes30, v.NewRegister) * 100.0
		v.PayMinutesRate60 = ComputeFloat(v.PayMinutes60, v.NewRegister) * 100.0
		v.PayMinutesRate120 = ComputeFloat(v.PayMinutes120, v.NewRegister) * 100.0
		v.PayLoginNumberRate = ComputeFloat(v.PayLoginNumber, v.LoginNum) * 100.0
		v.PayRequestRate = ComputeFloat(v.PayRequest, v.LoginNum) * 100.0
		v.PaySuccessOrderRate = ComputeFloat(v.PaySuccessOrder, v.PayRequestOrder) * 100.0

		v.FTrustUserPay = ComputeFloat(v.TrustUserPay, 100)
		v.FTrustUserWithdraw = ComputeFloat(v.TrustUserWithdraw, 100)
		v.TrustRate = ComputeFloat(v.TrustUserWithdraw, v.TrustUserPay) * 100.0
		// 计算渠道总充额占比、新充额、老充额、日活占比
		for n, m := range date_map {
			if n == v.Date {
				v.ChannelNewPayRate = ComputeFloat(v.NewRechargeAmount, m.NewAmount) * 100
				v.ChannelOldPayRate = ComputeFloat(v.OldRechargeAmount, m.OldAmount) * 100
				v.ChannelZCRate = ComputeFloat(v.TotalAmount, m.TotalAmount) * 100
				v.ChannelRHRate = ComputeFloat(v.LoginNum, m.LoginNum) * 100
			}
		}

		list[k] = v
	}
	return list
}

// 查询数据汇总条数
func (this *statisticsService) GetDataTotal(m bson.M, isChannel bool) (int64, error) {
	pipeline := []bson.M{}
	if isChannel {
		pipeline = []bson.M{
			{
				"$match": m,
			},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":     "$date",
						"channel":  "$channel",
						"channel1": "$channel1",
					},
				},
			},
		}
	} else {
		pipeline = []bson.M{
			{
				"$match": m,
			},
			{
				"$group": bson.M{
					"_id": "$date",
				},
			},
		}
	}

	// operations := []bson.M{m, n}
	result := []bson.M{}
	pipe := DataStatisticss.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("GetDataTotal fail err: ", err)
	}
	return int64(len(result)), nil
}

// 新增|修改 数据汇总（新）
func (this *statisticsService) AddDataStatistics(channel *entity.DataStatistics) error {
	info := new(entity.DataStatistics)
	GetByQ(DataStatisticss, bson.M{"date": channel.Date, "data_types": channel.DataTypes, "channel": channel.Channel}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"channel1":               channel.Channel1,
			"old_recharge_num":       channel.OldRechargeNum,
			"old_recharge_amount":    channel.OldRechargeAmount,
			"old_login_num":          channel.OldLoginNum,
			"new_register":           channel.NewRegister,
			"new_equipment":          channel.NewEquipment,
			"login_num":              channel.LoginNum,
			"next_new_register":      channel.NextNewRegister,
			"next_login":             channel.NextLogin,
			"jr_login_next_pay":      channel.JRLoginNextPay,
			"zr_pay_count":           channel.ZRPayCount,
			"new_recharge_num":       channel.NewRechargeNum,
			"new_recharge_amount":    channel.NewRechargeAmount,
			"total_recharge":         channel.TotalRecharge,
			"total_amount":           channel.TotalAmount,
			"pay_request":            channel.PayRequest,
			"pay_request_order":      channel.PayRequestOrder,
			"pay_success_order":      channel.PaySuccessOrder,
			"withdraw_request":       channel.WithdrawRequest,
			"withdraw_request_order": channel.WithdrawRequestOrder,
			"withdraw_success_order": channel.WithdrawSuccessOrder,
			"total_withdraw_num":     channel.TotalWithdrawNum,
			"total_withdraw_amount":  channel.TotalWithdrawAmount,
			"handling_charge":        channel.HandlingCharge,
		}
		if Update(DataStatisticss, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + channel.Id)
	} else {
		// 新增
		channel.Id = bson.NewObjectId().Hex()
		if !Insert(DataStatisticss, channel) {
			return errors.New("写入失败:" + channel.Id)
		}
		return nil
	}
}

/*
数据汇总（按时间）
*/
// 查询数据汇总（按时间计算）
func (this *statisticsService) GetDtList(channelArr, aliasArr []string, strName string, stime, etime string) ([]*entity.DataStatistics, error) {
	var list []*entity.DataStatistics
	// skipNum, _ := parsePageAndSort(page, pageSize, "date", false)
	var channellist []entity.ChannelInfo
	isChannel := true
	if len(channelArr) <= 0 && len(aliasArr) <= 0 && strName == "" {
		isChannel = false
	}
	if isChannel {
		m := bson.M{}
		m["status"] = 1
		if len(channelArr) > 0 {
			m["name"] = bson.M{"$in": channelArr}
		}
		if len(aliasArr) > 0 {
			m["name1"] = bson.M{"$in": aliasArr}
		}
		if strName != "" {
			m["name1"] = bson.M{"$regex": strName, "$options": "i"}
		}
		Channels.Find(m).All(&channellist)
	}
	s := fmt.Sprintf("%s 00:00:00", stime)
	startTime := utils.Str2Time(s, Location())
	e := fmt.Sprintf("%s 23:59:59", etime)
	endTime := utils.Str2Time(e, Location())
	if len(channellist) > 0 {
		// 按渠道
		for _, item := range channellist {
			info := new(entity.DataStatistics)
			id := item.Name // 渠道名称
			name1 := item.Name1
			m := bson.M{}
			m["robot"] = false
			m["simulation_robot"] = false
			m["ctime"] = bson.M{"$gte": startTime, "$lt": endTime}
			m["ad__bundle_id"] = id
			var userid []string //新注册用户ID
			var adid []string
			PlayerUsers.Find(m).Distinct("_id", &userid)
			m["ad__adid"] = bson.M{"$ne": ""}
			PlayerUsers.Find(m).Distinct("ad__adid", &adid) // 新设备数
			var logId []string                              // 登录用户ID
			m1 := bson.M{}
			m1["login_time"] = bson.M{"$gte": startTime, "$lt": endTime}
			LoginLogs.Find(m1).Distinct("userid", &logId)
			var lUserid []string
			m2 := bson.M{}
			m2["robot"] = false
			m2["simulation_robot"] = false
			m2["ad__bundle_id"] = id
			m2["_id"] = bson.M{"$in": logId}
			// 查询该渠道登录人数
			PlayerUsers.Find(m2).Distinct("_id", &lUserid)
			zrRegister := 0 // 昨日注册
			jrLogin := 0    // 昨日注册且今日登录
			zrPayNum := 0   // 昨日充值
			jrPayLogin := 0 // 昨日充值且今日登录
			// 获取开始日期-结束日期相差天数
			duration := endTime.Sub(startTime)
			days := int(duration.Hours() / 24)
			for i := 0; i <= days; i++ {
				dy := i - 1

				Twotoday := startTime.AddDate(0, 0, dy).Format("2006-01-02")
				today := startTime.AddDate(0, 0, i).Format("2006-01-02")

				var zrUserid []string
				n1 := FindByDate(Twotoday, Twotoday, "ctime", "ctime")
				n1["robot"] = false
				n1["simulation_robot"] = false
				n1["ad__bundle_id"] = id
				PlayerUsers.Find(n1).Distinct("_id", &zrUserid)
				zrRegister += len(zrUserid)
				// 昨日注册且今天登录
				var jrLoginUser []string
				n2 := FindByDate(today, today, "login_time", "login_time")
				n2["userid"] = bson.M{"$in": zrUserid}
				LoginLogs.Find(n2).Distinct("userid", &jrLoginUser)
				jrLogin += len(jrLoginUser)

				// 获取今日登录昨日充值玩家
				var zrPay []string
				var jrPayUser []string
				n3 := FindByDate(Twotoday, Twotoday, "ctime", "ctime")
				n3["order_status"] = 4
				n3["package_id"] = id
				Pays.Find(n3).Distinct("userid", &zrPay)
				zrPayNum += len(zrPay)

				n2["userid"] = bson.M{"$in": zrPay}
				LoginLogs.Find(n2).Distinct("userid", &jrPayUser)
				jrPayLogin += len(jrPayUser)
			}
			// 新用户充值
			pipeline2 := []bson.M{
				{
					"$match": bson.M{
						"order_status": 4,
						"package_id":   id,
						"ctime":        bson.M{"$gte": startTime, "$lt": endTime},
						"userid":       bson.M{"$in": userid},
					},
				},
				{
					"$group": bson.M{
						"_id": "$userid",
						"amount": bson.M{
							"$sum": "$amount",
						},
					},
				},
			}
			result2 := []bson.M{}
			pipe2 := Pays.Pipe(pipeline2)
			pipe2.All(&result2)
			newAmount := 0
			newNum := len(result2)
			for _, item := range result2 {
				newAmount += item["amount"].(int)
			}

			// 充值
			pipeline := []bson.M{
				{
					"$match": bson.M{
						"order_status": 4,
						"package_id":   id,
						"ctime":        bson.M{"$gte": startTime, "$lt": endTime},
					},
				},
				{
					"$group": bson.M{
						"_id": "$userid",
						"amount": bson.M{
							"$sum": "$amount",
						},
					},
				},
			}
			result := []bson.M{}
			pipe := Pays.Pipe(pipeline)
			pipe.All(&result)
			totalAmount := 0
			totalNum := len(result)
			for _, item := range result {
				totalAmount += item["amount"].(int)
			}
			// 该时间段中所有充值
			pipeline1 := []bson.M{
				{
					"$match": bson.M{
						"package_id": id,
						"ctime":      bson.M{"$gte": startTime, "$lt": endTime},
					},
				},
				{
					"$group": bson.M{
						"_id": "$order_status",
						"num": bson.M{
							"$sum": 1,
						},
						"amount": bson.M{
							"$sum": "$amount",
						},
					},
				},
			}
			result1 := []bson.M{}
			pipe1 := Pays.Pipe(pipeline1)
			pipe1.All(&result1)
			// cgAmount := 0
			cgCount := 0
			// sbAmount := 0
			// sbCount := 0
			reqOrder := 0
			for _, item := range result1 {
				status := item["_id"].(int)
				// 成功
				if status == 4 {
					cgCount = item["num"].(int)
					// cgAmount = item["amount"].(int)
				}
				// 失败
				if status == 2 {
					// sbCount = item["num"].(int)
					// sbAmount = item["amount"].(int)
				}
				reqOrder += item["num"].(int)
			}
			var reqUser []string
			m3 := bson.M{}
			m3["ctime"] = bson.M{"$gte": startTime, "$lt": endTime}
			m3["package_id"] = id
			Pays.Find(m3).Distinct("userid", &reqUser)
			reqNum := len(reqUser)
			// 提现
			tlist, _ := PayService.GetByWithdrawUser(m3)
			cgNumber1 := int64(0) // 提现成功订单数
			cgAmount1 := int64(0) // 提现成功金额
			HAmount1 := int64(0)  // 提现手续费
			uniqueUserIDs1 := make(map[string]bool)
			sgUserIDs := make(map[string]bool)
			for _, v := range tlist {
				if v.OrderStatus == 2 {
					// 成功
					cgNumber1 += 1
					cgAmount1 += int64(v.Amount)
					HAmount1 += int64(v.Commission)
					sgUserIDs[v.Userid] = true
				}
				uniqueUserIDs1[v.Userid] = true
			}
			var tlistUsers []string
			for id := range uniqueUserIDs1 {
				tlistUsers = append(tlistUsers, id)
			}
			var cgUsers []string
			for id := range sgUserIDs {
				cgUsers = append(cgUsers, id)
			}
			reqOrder1 := int64(len(tlist))       // 请求订单数
			reqNumber1 := int64(len(tlistUsers)) // 请求订单人数
			sgNumber1 := int64(len(cgUsers))     // 成功人数

			info.DateStr = stime + "-" + etime
			info.Channel = id
			info.Channel1 = name1
			info.NewRegister = int64(len(userid))
			info.NewEquipment = int64(len(adid))
			info.LoginNum = int64(len(lUserid))
			info.NextNewRegister = int64(zrRegister)
			info.NextLogin = int64(jrLogin)
			info.ZRPayCount = int64(zrPayNum)
			info.JRLoginNextPay = int64(jrPayLogin)
			// 新用户
			info.NewRechargeNum = int64(newNum)
			info.NewRechargeAmount = int64(newAmount)
			//
			info.TotalRecharge = int64(totalNum)
			info.TotalAmount = int64(totalAmount)

			// 充值成功率
			info.PayRequest = int64(reqNum)
			info.PayRequestOrder = int64(reqOrder)
			info.PaySuccessOrder = int64(cgCount)

			// 提现成功率
			info.WithdrawRequest = reqNumber1
			info.WithdrawRequestOrder = reqOrder1
			info.WithdrawSuccessOrder = cgNumber1

			info.TotalWithdrawNum = sgNumber1
			info.TotalWithdrawAmount = cgAmount1
			info.HandlingCharge = HAmount1
			list = append(list, info)

		}

	} else {
		// 全部
		info := new(entity.DataStatistics)
		m := bson.M{}
		m["robot"] = false
		m["simulation_robot"] = false
		m["ctime"] = bson.M{"$gte": startTime, "$lt": endTime}
		var userid []string //新注册用户ID
		var adid []string
		PlayerUsers.Find(m).Distinct("_id", &userid)
		m["ad__adid"] = bson.M{"$ne": ""}
		PlayerUsers.Find(m).Distinct("ad__adid", &adid) // 新设备数
		var logId []string                              // 登录用户ID
		m1 := bson.M{}
		m1["login_time"] = bson.M{"$gte": startTime, "$lt": endTime}
		LoginLogs.Find(m1).Distinct("userid", &logId)
		var lUserid []string
		m2 := bson.M{}
		m2["robot"] = false
		m2["simulation_robot"] = false
		m2["_id"] = bson.M{"$in": logId}
		// 查询该渠道登录人数
		PlayerUsers.Find(m2).Distinct("_id", &lUserid)

		zrRegister := 0 // 昨日注册
		jrLogin := 0    // 昨日注册且今日登录
		zrPayNum := 0   // 昨日充值
		jrPayLogin := 0 // 昨日充值且今日登录
		newAmount := 0  // 新用户充值
		newNum := 0     // 新用户
		// 获取开始日期-结束日期相差天数
		duration := endTime.Sub(startTime)
		days := int(duration.Hours() / 24)
		for i := 0; i <= days; i++ {
			dy := i - 1

			Twotoday := startTime.AddDate(0, 0, dy).Format("2006-01-02")
			today := startTime.AddDate(0, 0, i).Format("2006-01-02")

			var zrUserid []string
			n1 := FindByDate(Twotoday, Twotoday, "ctime", "ctime")
			n1["robot"] = false
			n1["simulation_robot"] = false
			PlayerUsers.Find(n1).Distinct("_id", &zrUserid)
			zrRegister += len(zrUserid)
			// 昨日注册且今天登录
			var jrLoginUser []string
			n2 := FindByDate(today, today, "login_time", "login_time")
			n2["userid"] = bson.M{"$in": zrUserid}
			LoginLogs.Find(n2).Distinct("userid", &jrLoginUser)
			jrLogin += len(jrLoginUser)

			// 获取今日登录昨日充值玩家
			var zrPay []string
			var jrPayUser []string
			n3 := FindByDate(Twotoday, Twotoday, "ctime", "ctime")
			n3["order_status"] = 4
			Pays.Find(n3).Distinct("userid", &zrPay)
			zrPayNum += len(zrPay)

			n2["userid"] = bson.M{"$in": zrPay}
			LoginLogs.Find(n2).Distinct("userid", &jrPayUser)
			jrPayLogin += len(jrPayUser)

			s1 := fmt.Sprintf("%s 00:00:00", today)
			// startTime1 := utils.Str2Time(s1)
			e1 := fmt.Sprintf("%s 23:59:59", today)
			// endTime1 := utils.Str2Time(e1)
			// var new_userid []string
			// if days == 0 {
			// 	// 只查询了一天
			// 	new_userid = userid
			// } else {
			// 	// 获取当天注册用户id
			// 	new_m := bson.M{}
			// 	new_m["ctime"] = bson.M{"$gte": startTime1, "$lt": endTime1}
			// 	new_m["robot"] = false
			// 	new_m["simulation_robot"] = false
			// 	PlayerUsers.Find(new_m).Distinct("_id", &new_userid)
			// }

			// 获取当天注册用户id
			var res []string
			sql1 := `select userid  from game.col_user cu FINAL 
			where robot = 0 and simulation_robot = 0 and ctime between toDateTime(?, ?) and toDateTime(?, ?) `
			var args1 []any
			// args1 = append(args1, startTime1.Unix())
			// args1 = append(args1, endTime1.Unix())
			args1 = append(args1, s1)
			args1 = append(args1, locationName)
			args1 = append(args1, e1)
			args1 = append(args1, locationName)
			ck.Select(&res, sql1, args1...)

			// 新用户充值
			var new_pay_res []map[string]any
			sql3 := `SELECT SUM(amount) total_amount,COUNT(DISTINCT userid) number from game.col_trade_record ctr FINAL 
			where order_status = 4 and package_id != '' and ctime between toDateTime(?, ?) and toDateTime(?, ?)  and userid in ?`
			var args3 []any
			args3 = append(args3, s1)
			args3 = append(args3, locationName)
			args3 = append(args3, e1)
			args3 = append(args3, locationName)
			args3 = append(args3, res)
			ck.Select(&new_pay_res, sql3, args3...)
			for _, v := range new_pay_res {
				number := v["number"].(uint64)
				amount := v["total_amount"].(uint64)
				if number != 0 {
					newNum += int(number)
				}
				if amount != 0 {
					newAmount += int(amount)
				}
			}
		}

		// 充值
		pipeline := []bson.M{
			{
				"$match": bson.M{
					"order_status": 4,
					"ctime":        bson.M{"$gte": startTime, "$lt": endTime},
				},
			},
			{
				"$group": bson.M{
					"_id": "$userid",
					"amount": bson.M{
						"$sum": "$amount",
					},
				},
			},
		}
		result := []bson.M{}
		pipe := Pays.Pipe(pipeline)
		pipe.All(&result)
		totalAmount := 0
		totalNum := len(result)
		for _, item := range result {
			totalAmount += item["amount"].(int)
		}
		// 该时间段中所有充值
		pipeline1 := []bson.M{
			{
				"$match": bson.M{
					"ctime": bson.M{"$gte": startTime, "$lt": endTime},
				},
			},
			{
				"$group": bson.M{
					"_id": "$order_status",
					"num": bson.M{
						"$sum": 1,
					},
					"amount": bson.M{
						"$sum": "$amount",
					},
				},
			},
		}
		result1 := []bson.M{}
		pipe1 := Pays.Pipe(pipeline1)
		pipe1.All(&result1)
		// cgAmount := 0
		cgCount := 0
		// sbAmount := 0
		// sbCount := 0
		reqOrder := 0
		for _, item := range result1 {
			status := item["_id"].(int)
			// 成功
			if status == 4 {
				cgCount = item["num"].(int)
				// cgAmount = item["amount"].(int)
			}
			// 失败
			if status == 2 {
				// sbCount = item["num"].(int)
				// sbAmount = item["amount"].(int)
			}
			reqOrder += item["num"].(int)
		}
		var reqUser []string
		m3 := bson.M{}
		m3["ctime"] = bson.M{"$gte": startTime, "$lt": endTime}
		Pays.Find(m3).Distinct("userid", &reqUser)
		reqNum := len(reqUser)
		// 提现
		tlist, _ := PayService.GetByWithdrawUser(m3)
		cgNumber1 := int64(0) // 提现成功订单数
		cgAmount1 := int64(0) // 提现成功金额
		HAmount1 := int64(0)  // 提现手续费
		uniqueUserIDs1 := make(map[string]bool)
		sgUserIDs := make(map[string]bool)
		for _, v := range tlist {
			if v.OrderStatus == 2 {
				// 成功
				cgNumber1 += 1
				cgAmount1 += int64(v.Amount)
				HAmount1 += int64(v.Commission)
				sgUserIDs[v.Userid] = true
			}
			uniqueUserIDs1[v.Userid] = true
		}
		var tlistUsers []string
		for id := range uniqueUserIDs1 {
			tlistUsers = append(tlistUsers, id)
		}
		var cgUsers []string
		for id := range sgUserIDs {
			cgUsers = append(cgUsers, id)
		}
		reqOrder1 := int64(len(tlist))       // 请求订单数
		reqNumber1 := int64(len(tlistUsers)) // 请求订单人数
		sgNumber1 := int64(len(cgUsers))     // 成功人数

		info.DateStr = stime + "-" + etime
		info.Channel = ""
		info.Channel1 = ""
		info.NewRegister = int64(len(userid))
		info.NewEquipment = int64(len(adid))
		info.LoginNum = int64(len(lUserid))
		info.NextNewRegister = int64(zrRegister)
		info.NextLogin = int64(jrLogin)
		info.ZRPayCount = int64(zrPayNum)
		info.JRLoginNextPay = int64(jrPayLogin)
		// 新用户
		info.NewRechargeNum = int64(newNum)
		info.NewRechargeAmount = int64(newAmount)
		info.TotalRecharge = int64(totalNum)
		info.TotalAmount = int64(totalAmount)
		// 充值成功率
		info.PayRequest = int64(reqNum)
		info.PayRequestOrder = int64(reqOrder)
		info.PaySuccessOrder = int64(cgCount)

		// 提现成功率
		info.WithdrawRequest = reqNumber1
		info.WithdrawRequestOrder = reqOrder1
		info.WithdrawSuccessOrder = cgNumber1

		info.TotalWithdrawNum = sgNumber1
		info.TotalWithdrawAmount = cgAmount1
		info.HandlingCharge = HAmount1
		list = append(list, info)

	}
	if len(list) > 1 {
		list = this.chipList25(list)

	} else {
		list = this.chipList19(list)
	}

	return list, nil
}

// 转换为分展示
func (this *statisticsService) chipList25(list []*entity.DataStatistics) []*entity.DataStatistics {
	totalInfo := new(entity.DataStatistics)
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		v.FDate = v.SDate.Format(utils.FORMAT_DATE)
		v.NextDayRetention = ComputeFloat(v.NextLogin, v.NextNewRegister) * 100
		v.PayRetention = ComputeFloat(v.JRLoginNextPay, v.ZRPayCount) * 100
		FRechargeAmount := ComputeFloat(v.NewRechargeAmount, 100)
		v.NewRechargeArpu = ComputeFloat(int64(FRechargeAmount), v.NewRegister)     //ComputeFloat(FTotalAmount, v.NewRegister) * 100
		v.NewRechargeArppu = ComputeFloat(int64(FRechargeAmount), v.NewRechargeNum) // ComputeFloat(v.NewRechargeAmount, v.NewRechargeNum) * 100
		v.NewUserPaymentRate = ComputeFloat(v.NewRechargeNum, v.NewRegister) * 100

		FOldRechargeAmount := ComputeFloat(v.OldRechargeAmount, 100)
		v.OldRechargeArpu = ComputeFloat(int64(FOldRechargeAmount), v.OldLoginNum)     //ComputeFloat(FTotalAmount, v.NewRegister) * 100
		v.OldRechargeArppu = ComputeFloat(int64(FOldRechargeAmount), v.OldRechargeNum) // ComputeFloat(v.NewRechargeAmount, v.NewRechargeNum) * 100
		v.OldUserPaymentRate = ComputeFloat(v.OldRechargeNum, v.OldLoginNum) * 100
		FTotalAmount := ComputeFloat(v.TotalAmount, 100)
		v.TotalRechargeArpu = ComputeFloat(int64(FTotalAmount), v.LoginNum)       // ComputeFloat(v.TotalAmount, v.LoginNum) * 100
		v.TotalRechargeArppu = ComputeFloat(int64(FTotalAmount), v.TotalRecharge) // ComputeFloat(v.TotalAmount, v.TotalRecharge) * 100
		v.TotalPaymentRate = ComputeFloat(v.TotalRecharge, v.LoginNum) * 100
		v.PaySuccessRate = ComputeFloat(v.PaySuccessOrder, v.PayRequestOrder) * 100
		v.WithdrawSuccessRate = ComputeFloat(v.WithdrawSuccessOrder, v.WithdrawRequestOrder) * 100
		v.CostRatio = ComputeFloat(v.TotalWithdrawAmount, v.TotalAmount) * 100
		v.FNewRechargeAmount = FRechargeAmount
		v.FOldRechargeAmount = FOldRechargeAmount
		v.FTotalAmount = FTotalAmount
		v.FTotalWithdrawAmount = ComputeFloat(v.TotalWithdrawAmount, 100)
		v.FHandlingCharge = ComputeFloat(v.HandlingCharge, 100)

		// 汇总
		totalInfo.NewRegister += v.NewRegister
		totalInfo.NewEquipment += v.NewEquipment
		totalInfo.LoginNum += v.LoginNum
		totalInfo.OldLoginNum += v.OldLoginNum
		totalInfo.NextNewRegister += v.NextNewRegister
		totalInfo.NextLogin += v.NextLogin
		totalInfo.JRLoginNextPay += v.JRLoginNextPay
		totalInfo.ZRPayCount += v.ZRPayCount
		totalInfo.OldRechargeNum += v.OldRechargeNum
		totalInfo.OldRechargeAmount += v.OldRechargeAmount
		totalInfo.NewRechargeNum += v.NewRechargeNum
		totalInfo.NewRechargeAmount += v.NewRechargeAmount
		totalInfo.TotalRecharge += v.TotalRecharge
		totalInfo.TotalAmount += v.TotalAmount
		totalInfo.PayRequest += v.PayRequest
		totalInfo.PayRequestOrder += v.PayRequestOrder
		totalInfo.PaySuccessOrder += v.PaySuccessOrder
		totalInfo.WithdrawRequest += v.WithdrawRequest
		totalInfo.WithdrawRequestOrder += v.WithdrawRequestOrder
		totalInfo.WithdrawSuccessOrder += v.WithdrawSuccessOrder
		totalInfo.TotalWithdrawNum += v.TotalWithdrawNum
		totalInfo.TotalWithdrawAmount += v.TotalWithdrawAmount
		totalInfo.HandlingCharge += v.HandlingCharge
		list[k] = v
	}
	totalInfo.Channel = "汇总"
	newList := make([]*entity.DataStatistics, 0)
	newList = append(newList, totalInfo)
	newList = this.chipList19(newList)
	list = append(list, newList...)
	return list

}

/*
AD上报统计相关接口
*/
func (this *statisticsService) GetAdReportList(page, pageSize int, m bson.M) ([]entity.AdReportData, error) {
	var list []entity.AdReportData
	if pageSize == -1 {
		pageSize = 100000
	}

	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "date", false)
	err := AdReportDatas.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList20(list)
	return list, err
}

func (this *statisticsService) chipList20(list []entity.AdReportData) []entity.AdReportData {
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		list[k] = v
	}
	return list
}

func (this *statisticsService) GetAdReportTotal(m bson.M) (int64, error) {
	// pipeline := []bson.M{
	// 	{
	// 		"$match": m,
	// 	},
	// 	{
	// 		"$group": bson.M{
	// 			"_id": "$date",
	// 		},
	// 	},
	// }
	// // operations := []bson.M{m, n}
	// result := []bson.M{}
	// pipe := AdReportDatas.Pipe(pipeline)
	// err := pipe.All(&result)
	// if err != nil {
	// 	beego.Error("GetDataTotal fail err: ", err)
	// }
	// return int64(len(result)), nil
	return int64(Count(AdReportDatas, m)), nil
}

// 新增|修改 AD上报统计
func (this *statisticsService) AddOrUpdateAdReport(rate *entity.AdReportData) error {
	info := new(entity.AdReportData)
	GetByQ(AdReportDatas, bson.M{"date": rate.Date, "channel": rate.Channel}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"channel1":            rate.Channel1,
			"register_count":      rate.RegisterCount,
			"register_count_day":  rate.RegisterCountDay,
			"register_count_day1": rate.RegisterCountDay1,
			"fail_register_count": rate.FailRegisterCount,
			"login_count":         rate.LoginCount,
			"login_count_day":     rate.LoginCountDay,
			"login_count_day1":    rate.LoginCountDay1,
			"fail_login_count":    rate.FailLoginCount,
			"deposit_count":       rate.DepositCount,
			"deposit_count_day":   rate.DepositCountDay,
			"deposit_count_day1":  rate.DepositCountDay1,
			"fail_deposit_count":  rate.FailDepositCount,
		}
		if Update(AdReportDatas, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + rate.Id)
	} else {
		// 新增
		rate.Id = bson.NewObjectId().Hex()
		if !Insert(AdReportDatas, rate) {
			return errors.New("写入失败:" + rate.Id)
		}
		return nil
	}
}

/*
分享数据相关接口
*/
func (this *statisticsService) GetShareList(page, pageSize int, m bson.M) ([]entity.ShareDataStatistics, error) {
	var list []entity.ShareDataStatistics
	if pageSize == -1 {
		pageSize = 100000
	}

	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "date", false)
	err := ShareStatistcs.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList21(list)
	return list, err
}

func (this *statisticsService) chipList21(list []entity.ShareDataStatistics) []entity.ShareDataStatistics {
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)

		list[k] = v
	}
	return list
}

func (this *statisticsService) GetShareTotal(m bson.M) (int64, error) {
	return int64(Count(ShareStatistcs, m)), nil
}

// 新增|修改 分享数据
func (this *statisticsService) AddOrUpdateShare(rate *entity.ShareDataStatistics) error {
	info := new(entity.ShareDataStatistics)
	GetByQ(ShareStatistcs, bson.M{"date": rate.Date, "channel": rate.Channel}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"channel1":        rate.Channel1,
			"register_number": rate.RegisterNumber,
			"pay_number":      rate.PayNumber,
			"pay_money":       rate.PayMoney,
		}
		if Update(ShareStatistcs, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + rate.Id)
	} else {
		// 新增
		rate.Id = bson.NewObjectId().Hex()
		if !Insert(ShareStatistcs, rate) {
			return errors.New("写入失败:" + rate.Id)
		}
		return nil
	}
}

/*
房间资源相关接口
*/
func (this *statisticsService) GetRoomResourcesList(page, pageSize int, m bson.M) ([]entity.CashFlow, error) {
	var list []entity.CashFlow
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)

	pipeline := []bson.M{
		{"$match": m},
		{
			"$group": bson.M{
				"_id": bson.M{
					"date": "$date",
				},
				"win_score":  bson.M{"$sum": "$win_score"},
				"lose_score": bson.M{"$sum": "$lose_score"},
			},
		},
		{
			"$project": bson.M{
				"_id":        "$_id.date",
				"date":       "$_id.date",
				"win_score":  "$win_score",
				"lose_score": "$lose_score",
			},
		},
		{"$sort": bson.M{"date": -1}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}
	err := CashFlows.Pipe(pipeline).All(&list)
	list = this.chipList22(list)
	return list, err
}

func (this *statisticsService) chipList22(list []entity.CashFlow) []entity.CashFlow {
	for k, v := range list {
		// v.SDate = time.Unix(v.Date, 0)
		v.FLoseScore = ComputeFloat(v.LoseScore, 100)
		v.FWinScore = ComputeFloat(v.WinScore, 100)
		total := v.WinScore - v.LoseScore
		v.ProfitAndLoss = ComputeFloat(total, 100)
		if total > 0 {
			v.IsWin = true
		} else {
			v.IsWin = false
		}
		list[k] = v
	}
	return list
}

func (this *statisticsService) GetRoomResourcesTotal(m bson.M) (int64, error) {
	pipeline := []bson.M{
		{
			"$match": m,
		},
		{
			"$group": bson.M{
				"_id": "$date",
			},
		},
	}
	// operations := []bson.M{m, n}
	result := []bson.M{}
	pipe := CashFlows.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("GetRoomResourcesTotal fail err: ", err)
	}
	return int64(len(result)), nil
}

/*
短信统计相关接口
*/
func (this *statisticsService) GetSMSList(page, pageSize int, m bson.M) ([]entity.SmsData, error) {
	var list []entity.SmsData
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "date", false)
	err := SmsDatas.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList23(list)
	return list, err

}

func (this *statisticsService) chipList23(list []entity.SmsData) []entity.SmsData {
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		v.SendSMSRatio = ComputeFloat(v.SendCount, v.TotalSendSMS) * 100
		v.UseSMSRatio = ComputeFloat(v.UseSMS, v.SendCount) * 100
		list[k] = v
	}
	return list
}

func (this *statisticsService) GetSMSTotal(m bson.M) (int64, error) {
	return int64(Count(SmsDatas, m)), nil
}

func (this *statisticsService) GetSMSByDate(m bson.M) (*entity.SmsData, error) {
	info := new(entity.SmsData)
	GetByQ(SmsDatas, m, info)
	return info, nil
}

// 新增或更新短信统计
func (this *statisticsService) AddOrUpdateSMS(user *entity.SmsData) error {
	info := new(entity.SmsData)
	GetByQ(SmsDatas, bson.M{"date": user.Date}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"total_send_sms": user.TotalSendSMS,
			"send_count":     user.SendCount,
			"use_sms":        user.UseSMS,
		}
		if Update(SmsDatas, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + user.Id)
	} else {
		// 新增
		user.Id = bson.NewObjectId().Hex()
		if !Insert(SmsDatas, user) {
			return errors.New("写入失败:" + user.Id)
		}
		return nil
	}
}

/*
充值来源相关接口
*/
func (this *statisticsService) GetPaySourceList(page, pageSize int, m bson.M) ([]entity.PaySourceData, error) {
	var list []entity.PaySourceData
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "date", false)
	err := PaySources.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList24(list)
	return list, err

}

func (this *statisticsService) chipList24(list []entity.PaySourceData) []entity.PaySourceData {
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		v.FTPAmount = ComputeFloat(v.TPAmount, 100)
		v.TPAvg = ComputeFloat(int64(v.FTPAmount), v.TPNumber)

		v.FRMAmount = ComputeFloat(v.RMAmount, 100)
		v.RMAvg = ComputeFloat(int64(v.FRMAmount), v.RMNumber)

		v.FLHDAmount = ComputeFloat(v.LHDAmount, 100)
		v.LHDAvg = ComputeFloat(int64(v.FLHDAmount), v.LHDNumber)

		v.FUPAmount = ComputeFloat(v.UPAmount, 100)
		v.UPAvg = ComputeFloat(int64(v.FUPAmount), v.UPNumber)

		v.FAKAmount = ComputeFloat(v.AKAmount, 100)
		v.AKAvg = ComputeFloat(int64(v.FAKAmount), v.AKNumber)

		v.FJOKERAmount = ComputeFloat(v.JOKERAmount, 100)
		v.JOKERAvg = ComputeFloat(int64(v.FJOKERAmount), v.JOKERNumber)

		v.FCRASHAmount = ComputeFloat(v.CRASHAmount, 100)
		v.CRASHAvg = ComputeFloat(int64(v.FCRASHAmount), v.CRASHNumber)

		v.FABAmount = ComputeFloat(v.ABAmount, 100)
		v.ABAvg = ComputeFloat(int64(v.FABAmount), v.ABNumber)

		v.FRMTwoAmount = ComputeFloat(v.RMTwoAmount, 100)
		v.RMTwoAvg = ComputeFloat(int64(v.FRMTwoAmount), v.RMTwoNumber)
		list[k] = v
	}
	return list
}

func (this *statisticsService) GetPaySourceTotal(m bson.M) (int64, error) {
	return int64(Count(PaySources, m)), nil
}

// 新增或更新充值来源
func (this *statisticsService) AddOrUpdatePaySource(user *entity.PaySourceData) error {
	info := new(entity.PaySourceData)
	GetByQ(PaySources, bson.M{"date": user.Date}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"tp_amount":    user.TPAmount,
			"tp_number":    user.TPNumber,
			"lhd_amount":   user.LHDAmount,
			"lhd_number":   user.LHDNumber,
			"rm_amount":    user.RMAmount,
			"rm_number":    user.RMNumber,
			"up_amount":    user.UPAmount,
			"up_number":    user.UPNumber,
			"ak_amount":    user.AKAmount,
			"ak_number":    user.AKNumber,
			"joker_amount": user.JOKERAmount,
			"joker_number": user.JOKERNumber,
			"crash_amount": user.CRASHAmount,
			"crash_number": user.CRASHNumber,
			"ab_amount":    user.ABAmount,
			"ab_number":    user.ABNumber,
			"cp_amount":    user.CPAmount,
			"cp_number":    user.CPNumber,
			"fj_amount":    user.FJAmount,
			"fj_number":    user.FJNumber,
		}
		if Update(PaySources, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + user.Id)
	} else {
		// 新增
		user.Id = bson.NewObjectId().Hex()
		if !Insert(PaySources, user) {
			return errors.New("写入失败:" + user.Id)
		}
		return nil
	}
}

/*
对战房埋点统计相关接口
*/
func (this *statisticsService) GetBattleRoomList(page, pageSize int, m bson.M) ([]entity.BattleRoomData, error) {
	var list []entity.BattleRoomData
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "date", false)
	err := BattleRooms.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList26(list)
	return list, err

}

func (this *statisticsService) chipList26(list []entity.BattleRoomData) []entity.BattleRoomData {
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		list[k] = v
	}
	return list
}

func (this *statisticsService) GetBattleRoomTotal(m bson.M) (int64, error) {
	return int64(Count(BattleRooms, m)), nil
}

// 新增或更新对战房埋点
func (this *statisticsService) AddOrUpdateBattleRoom(user *entity.BattleRoomData) error {
	info := new(entity.BattleRoomData)
	GetByQ(BattleRooms, bson.M{"date": user.Date}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"clicks_count":         user.ClicksCount,
			"total_number":         user.TotalNumber,
			"create_room_count":    user.CreateRoomCount,
			"create_room_number":   user.CreateRoomNumber,
			"success_enter_count":  user.SuccessEnterCount,
			"success_enter_number": user.SuccessEnterNumber,
			"tp_enter_count":       user.TPEnterCount,
			"tp_enter_number":      user.TPEnterNumber,
			"rm_enter_count":       user.RMEnterCount,
			"rm_enter_number":      user.RMEnterNumber,
			"ab_enter_count":       user.ABEnterCount,
			"ab_enter_number":      user.ABEnterNumber,
			"join_room_count":      user.JoinRoomCount,
			"join_room_number":     user.JoinRoomNumber,
			"success_join_count":   user.SuccessJoinCount,
			"success_join_number":  user.SuccessJoinNumber,
			"tp_join_count":        user.TPJoinCount,
			"tp_join_number":       user.TPJoinNumber,
			"rm_join_count":        user.RMJoinCount,
			"rm_join_number":       user.RMJoinNumber,
			"ab_join_count":        user.ABJoinCount,
			"ab_join_number":       user.ABJoinNumber,
		}
		if Update(BattleRooms, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + user.Id)
	} else {
		// 新增
		user.Id = bson.NewObjectId().Hex()
		if !Insert(BattleRooms, user) {
			return errors.New("写入失败:" + user.Id)
		}
		return nil
	}
}

/*
大R分析相关接口
*/
func (this *statisticsService) GetUserAnalysisCk(page, pageSize int, pay_statr, pay_end, withdraw_statr, withdraw_end, gain_statr, gain_end int) (total int64, results []*entity.UserAnalysisInfo, err error) {
	query := ck.DB().Model(&ck.UserFinance{}).Table("col_user_finance final")
	if pay_statr != 0 && pay_end != 0 {
		// 开始结束范围
		query.Where("money >= ? and money < ?", pay_statr*100, pay_end*100)
	} else if pay_statr != 0 {
		// 开始 从pay_statr - 无限
		query.Where("money >= ?", pay_statr*100)
	} else if pay_end != 0 {
		// 开始 0- pay_end
		query.Where("money < ?", pay_end*100)
	}

	// 提现存在审核或者未到账金额，需要从提现记录中查询符合条件得用户
	if withdraw_statr != 0 || withdraw_end != 0 {
		var withdraws_ids []string
		var w_args []any
		str_sql := `
		SELECT userid from (SELECT userid, COUNT(*) withdraw_times, SUM(amount) withdraw_amount
		FROM game.col_withdraw_record t1 FINAL WHERE order_status = 2
		GROUP BY userid) tab
		`
		if withdraw_statr != 0 && withdraw_end != 0 {
			// 开始结束范围
			str_sql += " WHERE withdraw_amount >= ? and withdraw_amount < ?"
			w_args = append(w_args, withdraw_statr*100)
			w_args = append(w_args, withdraw_end*100)
			// query.Where("withdraw_amount >= ? and withdraw_amount < ?", withdraw_statr*100, withdraw_end*100)
		} else if withdraw_statr != 0 {
			// 开始 从pay_statr - 无限
			str_sql += " WHERE withdraw_amount >= ?"
			w_args = append(w_args, withdraw_statr*100)
			// query.Where("withdraw_amount >= ?", withdraw_statr*100)
		} else if withdraw_end != 0 {
			// 开始 0- pay_end
			str_sql += " WHERE withdraw_amount < ?"
			w_args = append(w_args, withdraw_end*100)
			// query.Where("withdraw_amount < ?", withdraw_end*100)
		}
		err = ck.Select(&withdraws_ids, str_sql, w_args...)
		if err != nil {
			return
		}
		if len(withdraws_ids) > 0 {
			query.Where("userid in ?", withdraws_ids)
		}
	}

	if gain_statr != 0 && gain_end != 0 {
		// 开始结束范围
		query.Where("profit >= ? and profit < ?", gain_statr*100, gain_end*100)
	} else if gain_statr != 0 {
		// 开始 从pay_statr - 无限
		query.Where("profit >= ?", gain_statr*100)
	} else if gain_end != 0 {
		// 开始 0- pay_end
		query.Where("profit < ?", gain_end*100)
	}

	err = query.Count(&total).Error
	if err != nil || total == 0 {
		return
	}

	query.Order("ctime desc")
	var userFinances []ck.UserFinance
	err = query.Scopes(Paginate(page, pageSize)).Find(&userFinances).Error
	if err != nil || len(userFinances) == 0 {
		return
	}

	var userResults = make(map[string]*entity.UserAnalysisInfo, len(userFinances))
	var userids = make([]string, 0, len(userFinances))
	for _, uf := range userFinances {
		info := &entity.UserAnalysisInfo{
			UserId:    uf.Userid,
			GainAmout: Chip2Float(int64(uf.Profit)),
			// PayAmount: Chip2Float(int64(uf.Money)),
			// WithdrawAmount: Chip2Float(int64(uf.CashOut)),
		}
		results = append(results, info)
		userResults[uf.Userid] = info
		userids = append(userids, uf.Userid)
	}

	// 渠道数据
	channel_list := make([]entity.ChannelInfo, 0)
	Channels.Find(nil).All(&channel_list)
	nameChannels := utils.Slice2Map(channel_list, func(i int, v entity.ChannelInfo) string { return v.Name })

	// 用户信息
	var users []ck.User
	err = ck.Select(&users, `select userid,ctime,login_time,ad__bundle_id,ad__adid,first_charge_time,first_charge_val from col_user final where userid in ?`, userids)
	if err != nil {
		return
	}
	var adids []string
	for _, user := range users {
		info, ok := userResults[user.Userid]
		if !ok {
			continue
		}
		// 渠道、别名
		info.Channel = user.AD_BundleId
		if ch, ok := nameChannels[info.Channel]; ok {
			info.Channel1 = ch.Name1
		}

		info.RegistTime = user.Ctime
		info.LastLoginTime = user.LoginTime

		time1 := info.RegistTime
		time2 := info.LastLoginTime
		// 将日期时间转换为整数天数
		date1 := time.Date(time1.Year(), time1.Month(), time1.Day(), 0, 0, 0, 0, time1.Location())
		date2 := time.Date(time2.Year(), time2.Month(), time2.Day(), 0, 0, 0, 0, time2.Location())

		// 计算相差的天数
		days := int(date2.Sub(date1).Hours() / 24)
		info.ActiveTime = int64(days + 1)

		c, _ := ConvertToIndiaTime(user.FirstChargeTime / 1000)
		info.FirstPayDate = c
		info.FirstPayAmount = Chip2Float(int64(user.FirstChargeVal))

		if user.AD_ADID != "" {
			adids = append(adids, user.AD_ADID)
		}
	}
	// 活跃天数
	var loginDays []map[string]any
	err = ck.Select(&loginDays, `
		SELECT userid, count(*) login_days FROM (
			SELECT userid, toDate(login_time) login_date FROM col_log_login cll FINAL WHERE userid in ? GROUP BY userid, login_date
		) s1 GROUP BY userid
	`, userids)
	if err != nil {
		return
	}
	for _, item := range loginDays {
		userid := item["userid"].(string)
		loginDays := utils.ToInt64(item["login_days"])
		if r, ok := userResults[userid]; ok {
			r.ActiveDay = loginDays
		}
	}

	// 关联账号,
	// 关联提现银行卡号
	var blankNumbers []map[string]any
	err = ck.Select(&blankNumbers, `
		SELECT userid, blank_number FROM col_withdraw_record final WHERE blank_number IN (
			SELECT blank_number FROM col_withdraw_record final WHERE userid IN ? AND order_status = 2 AND blank_number <> ''
		) GROUP BY userid, blank_number
	`, userids)
	if err != nil {
		return
	}
	var bankUsers = make(map[string][]string)
	var userBanks = make(map[string][]string)
	for _, item := range blankNumbers {
		userid := item["userid"].(string)
		bank := item["blank_number"].(string)
		bankUsers[bank] = append(bankUsers[bank], userid)
		userBanks[userid] = append(userBanks[userid], bank)
	}
	// 关联设备号
	var devices []map[string]any
	if len(adids) > 0 {
		err = ck.Select(&devices, `
			SELECT userid, ad__adid FROM col_user final WHERE ad__adid IN ?
		`, adids)
		if err != nil {
			return
		}
	}
	var deviceUsers = make(map[string][]string)
	for _, item := range devices {
		userid := item["userid"].(string)
		adid := item["ad__adid"].(string)
		deviceUsers[adid] = append(deviceUsers[adid], userid)
	}

	// 计算关联用户
	refUsers := make(map[string]map[string]bool)
	for _, user := range users {
		ref, ok := refUsers[user.Userid]
		if !ok {
			ref = make(map[string]bool)
			refUsers[user.Userid] = ref
		}
		// 银行卡号关联
		banks := userBanks[user.Userid]
		if len(banks) > 0 {
			for _, bank := range banks {
				for _, userid := range bankUsers[bank] {
					if userid != user.Userid {
						ref[userid] = true
					}
				}
			}
		}

		// 设备id关联
		for _, userid := range deviceUsers[user.AD_ADID] {
			if userid != user.Userid {
				ref[userid] = true
			}
		}
	}

	for _, result := range results {
		if ref, ok := refUsers[result.UserId]; ok {
			for userid := range ref {
				result.AssociatedUsers = append(result.AssociatedUsers, userid)
			}
			result.AssociatedCount = int64(len(ref))
		}
	}

	// 首充
	var firstPays []map[string]any
	err = ck.Select(&firstPays, `
		SELECT userid, ctime, amount FROM col_trade_record FINAL WHERE (userid, ctime) IN (
			SELECT userid, min(ctime) ctime FROM col_trade_record FINAL WHERE userid IN ? AND order_status = 4 GROUP BY userid
		)
	`, userids)
	if err != nil {
		return
	}
	for _, item := range firstPays {
		userid := item["userid"].(string)
		if r, ok := userResults[userid]; ok {
			r.FirstPayDate = item["ctime"].(time.Time)
			r.FirstPayAmount = Chip2Float(utils.ToInt64(item["amount"]))
		}
	}

	// 支付订单
	var pays []map[string]any
	err = ck.Select(&pays, `
		SELECT userid, COUNT(*) pay_times, SUM(amount) pay_amount, AVG(amount) pay_avg
		FROM col_trade_record t1 FINAL WHERE userid IN ? AND order_status = 4
		GROUP BY userid
	`, userids)
	if err != nil {
		return
	}
	for _, item := range pays {
		userid := item["userid"].(string)
		if r, ok := userResults[userid]; ok {
			r.PayCount = utils.ToInt64(item["pay_times"])
			r.PayAmount = Chip2Float(utils.ToInt64(item["pay_amount"]))
			r.PayAvg = Chip2Float(utils.ToInt64(item["pay_avg"]))
		}
	}

	// 提现订单
	var withdraws []map[string]any
	err = ck.Select(&withdraws, `
		SELECT userid, COUNT(*) withdraw_times, SUM(amount) withdraw_amount, AVG(amount) withdraw_avg
		FROM col_withdraw_record t1 FINAL WHERE userid IN ? AND order_status = 2
		GROUP BY userid
	`, userids)
	if err != nil {
		return
	}
	for _, item := range withdraws {
		userid := item["userid"].(string)
		if r, ok := userResults[userid]; ok {
			r.WithdrawCount = utils.ToInt64(item["withdraw_times"])
			r.WithdrawAmount = Chip2Float(utils.ToInt64(item["withdraw_amount"]))
			r.WithdrawAvg = Chip2Float(utils.ToInt64(item["withdraw_avg"]))
		}
	}
	// 游戏局数统计
	m4 := []bson.M{
		{"$match": bson.M{"userid": bson.M{"$in": userids}}},
		{
			"$group": bson.M{
				"_id":   bson.M{"userid": "$userid", "gtype": "$gtype"},
				"count": bson.M{"$sum": "$number"},
			},
		},
		{"$sort": bson.M{"count": -1}},
		{"$group": bson.M{
			"_id": "$_id.userid",
			"topGames": bson.M{
				"$push": bson.M{
					"gtype": "$_id.gtype",
					"count": "$count",
				},
			},
		}},
		{"$project": bson.M{
			"topGames": bson.M{"$slice": []interface{}{"$topGames", 2}},
		}},
	}
	var res []bson.M
	UserGameDatas.Pipe(m4).All(&res)
	for _, doc := range res {
		// 获取 userid
		userid := doc["_id"].(string) // 假设 _id 是字符串类型，根据实际情况调整类型断言
		// 获取 topGames 数组
		topGames := doc["topGames"].([]interface{})
		info, ok := userResults[userid]
		if !ok {
			continue
		}
		// 遍历 topGames 数组
		for i, game := range topGames {
			gameMap, _ := game.(bson.M)

			// 获取游戏类型和计数
			gtype := gameMap["gtype"].(int64)
			count := gameMap["count"].(int64)
			if i == 0 {
				info.GameCount = count
				info.GameName = GameByName(int(gtype))
			}
			if i == 1 {
				info.TwoGameCount = count
				info.TwoGameName = GameByName(int(gtype))
			}
		}
	}

	// // 游戏局数统计
	// m4 := []bson.M{
	// 	{"$match": bson.M{"userid": bson.M{"$in": userids}}},
	// 	{
	// 		"$group": bson.M{
	// 			"_id":   bson.M{"userid": "$userid", "gtype": "$gtype"},
	// 			"count": bson.M{"$sum": 1},
	// 		},
	// 	},
	// }
	// r4 := []bson.M{}
	// err = LogGameTimes.Pipe(m4).All(&r4)
	// if err != nil {
	// 	return
	// }
	// for _, r := range r4 {
	// 	userid := r["_id"].(bson.M)["userid"].(string)
	// 	gtype := r["_id"].(bson.M)["gtype"].(int)
	// 	count := int64(r["count"].(int))
	// 	info, ok := userResults[userid]
	// 	if !ok {
	// 		continue
	// 	}
	// 	if info.GameCount < count {
	// 		info.TwoGameCount = info.GameCount
	// 		info.TwoGameName = info.GameName
	// 		info.GameCount = count
	// 		info.GameName = GameByName(gtype)
	// 	}
	// }

	return
}

/*
大R分析相关接口
*/
func (this *statisticsService) GetUserAnalysis(pay_statr, pay_end, withdraw_statr, withdraw_end, gain_statr, gain_end int) ([]*entity.UserAnalysisInfo, error) {
	var list []*entity.UserAnalysisInfo
	n := bson.M{}
	if pay_statr != 0 && pay_end != 0 {
		// 开始结束范围
		n["pay_amount"] = bson.M{"$gte": pay_statr * 100, "$lt": pay_end * 100}
	} else if pay_statr != 0 {
		// 开始 从pay_statr - 无限
		n["pay_amount"] = bson.M{"$gte": pay_statr * 100}
	} else if pay_end != 0 {
		// 开始 0- pay_end
		n["pay_amount"] = bson.M{"$lt": pay_end * 100}
	}

	if withdraw_statr != 0 && withdraw_end != 0 {
		// 开始结束范围
		n["withdraw_amount"] = bson.M{"$gte": withdraw_statr * 100, "$lt": withdraw_end * 100}
	} else if withdraw_statr != 0 {
		// 开始 从pay_statr - 无限
		n["withdraw_amount"] = bson.M{"$gte": withdraw_statr * 100}
	} else if withdraw_end != 0 {
		// 开始 0- pay_end
		n["withdraw_amount"] = bson.M{"$lt": withdraw_end * 100}
	}

	if gain_statr != 0 && gain_end != 0 {
		// 开始结束范围
		n["gain_amout"] = bson.M{"$gte": gain_statr * 100, "$lt": gain_end * 100}
	} else if gain_statr != 0 {
		// 开始 从pay_statr - 无限
		n["gain_amout"] = bson.M{"$gte": gain_statr * 100}
	} else if gain_end != 0 {
		// 开始 0- pay_end
		n["gain_amout"] = bson.M{"$lt": gain_end * 100}
	}
	var u_list []entity.UserCashRecod
	UserCashRecods.Find(n).All(&u_list)
	// 渠道数据
	channel_list := make([]entity.ChannelInfo, 0)
	Channels.Find(nil).All(&channel_list)

	// 金额范围查询出来的信息
	for _, item := range u_list {
		new_info := new(entity.UserAnalysisInfo)
		userid := item.Userid

		user_info := new(entity.PlayerUser)
		PlayerUsers.FindId(userid).One(user_info)
		if user_info.Userid != "" {
			new_info.UserId = user_info.Userid
			c, _ := ConvertToIndiaTime(user_info.Ctime.Unix())
			new_info.RegistTime = c
			last, _ := ConvertToIndiaTime(user_info.LoginTime.Unix())
			new_info.LastLoginTime = last
			new_info.Channel = user_info.AD_BundleId
			// 存活时长
			time1 := new_info.RegistTime
			time2 := new_info.LastLoginTime
			// 将日期时间转换为整数天数
			date1 := time.Date(time1.Year(), time1.Month(), time1.Day(), 0, 0, 0, 0, time1.Location())
			date2 := time.Date(time2.Year(), time2.Month(), time2.Day(), 0, 0, 0, 0, time2.Location())

			// 计算相差的天数
			days := int(date2.Sub(date1).Hours() / 24)

			new_info.ActiveTime = int64(days + 1)

			// 登录天数
			var loginLogs []entity.LogLogin
			LoginLogs.Find(bson.M{"userid": userid}).Sort("login_time").All(&loginLogs)
			// 统计不同日期的登录次数
			loginDates := make(map[string]bool)
			for _, log := range loginLogs {
				loginDate := log.LoginTime.Format("2006-01-02") // 格式化为年-月-日
				loginDates[loginDate] = true
			}
			new_info.ActiveDay = int64(len(loginDates))
			// 获取渠道别名
			for _, channel := range channel_list {
				if channel.Name == user_info.AD_BundleId {
					new_info.Channel1 = channel.Name1
				}
			}
			// 关联账号查询
			// 查询设备号
			g_user_arr := make([]string, 0)
			m := bson.M{}
			m["ad__adid"] = user_info.AD_ADID
			m["_id"] = bson.M{"$ne": item}
			PlayerUsers.Find(m).Distinct("_id", &g_user_arr)
			// 查询银行卡号
			with_m := bson.M{}
			w_user_arr := make([]string, 0)
			with_m["order_status"] = 2
			with_m["userid"] = bson.M{"$ne": userid}
			with_m["blank_number"] = bson.M{"$in": item.CardNumber}
			Withdraws.Find(with_m).Distinct("userid", &w_user_arr)

			relevance_temp := make(map[string]bool, 0)
			for _, v := range g_user_arr {
				relevance_temp[v] = true
			}
			for _, v := range w_user_arr {
				relevance_temp[v] = true
			}
			relevance_arr := make([]string, 0)
			for k, _ := range relevance_temp {
				relevance_arr = append(relevance_arr, k)
			}
			new_info.AssociatedUsers = relevance_arr
			new_info.AssociatedCount = int64(len(relevance_arr))

			c, _ = ConvertToIndiaTime(item.FirstPayDate.Unix())
			new_info.FirstPayDate = c
			new_info.FirstPayAmount = Chip2Float(int64(item.FirstPayAmount))

			new_info.PayAmount = Chip2Float(int64(item.PayAmount))
			new_info.PayCount = int64(item.PayCount)
			new_info.PayAvg = Chip2Float(int64(item.PayAmount)) / float64(item.PayCount)
			// 赋值提现信息
			new_info.WithdrawAmount = Chip2Float(int64(item.WithdrawAmount))
			new_info.WithdrawCount = int64(item.WithdrawCount)
			if item.WithdrawAmount != 0 && item.WithdrawCount != 0 {
				new_info.WithdrawAvg = Chip2Float(int64(item.WithdrawAmount)) / float64(item.WithdrawCount)
			}

			// 游戏局数统计
			pipeline4 := []bson.M{
				{
					"$match": bson.M{
						"userid": userid,
					},
				},
				{
					"$group": bson.M{
						"_id": "$gtype",
						"Count": bson.M{
							"$sum": 1,
						},
					},
				},
				{
					"$sort": bson.M{
						"Count": -1, // 按 Count 字段降序排序
					},
				},
				{
					"$limit": 2, // 获取前两条数据
				},
			}
			result4 := []bson.M{}
			pipe4 := LogGameTimes.Pipe(pipeline4)
			pipe4.All(&result4)
			if len(result4) >= 1 {
				g_count := result4[0]["Count"].(int)
				g_type := result4[0]["_id"].(int)
				new_info.GameName = GameByName(g_type)
				new_info.GameCount = int64(g_count)
			}
			if len(result4) >= 2 {
				g_count := result4[1]["Count"].(int)
				g_type := result4[1]["_id"].(int)
				new_info.TwoGameName = GameByName(g_type)
				new_info.TwoGameCount = int64(g_count)
			}
			new_info.GainAmout = Chip2Float(item.GainAmout)
			list = append(list, new_info)
		}
	}
	// 按照注册时间倒序排序
	sort.Slice(list, func(i, j int) bool {
		return list[i].RegistTime.After(list[j].RegistTime)
	})
	return list, nil
}

func (this *statisticsService) GetUserAnalysis_old(pay_statr, pay_end, withdraw_statr, withdraw_end, gain_statr, gain_end int) ([]entity.UserAnalysisInfo, error) {
	var list []entity.UserAnalysisInfo
	user_arr := make([]string, 0)
	pay_list := []bson.M{}
	withdraw_list := []bson.M{}
	// 充值
	// query := []bson.M{
	// 	{
	// 		"$match": bson.M{
	// 			"order_status": 4,
	// 		},
	// 	},
	// 	{
	// 		"$group": bson.M{
	// 			"_id": "$userid",
	// 			"Amount": bson.M{
	// 				"$sum": "$amount",
	// 			},
	// 			"Count": bson.M{
	// 				"$sum": 1,
	// 			},
	// 		},
	// 	},
	// 	{"$group": bson.M{"_id": "$userid", "num": bson.M{"$sum": 1}}},
	// }
	// pipe2 := PlayerUsers.Pipe(query2)
	// result2 := []bson.M{}
	// err2 := pipe2.All(&result2)
	// if err2 == nil && len(result2) > 0 {
	// 	stat["scount"] = result2[0]["num"]
	// }
	// 充值金额
	n := bson.M{}
	if pay_statr != 0 && pay_end != 0 {
		// 开始结束范围
		n["Amount"] = bson.M{"$gte": pay_statr * 100, "$lt": pay_end * 100}
	} else if pay_statr != 0 {
		// 开始 从pay_statr - 无限
		s := pay_statr * 100
		n["Amount"] = bson.M{"$gte": s}
	} else if pay_end != 0 {
		// 开始 0- pay_end
		n["Amount"] = bson.M{"$lt": pay_end * 100}
	}
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"order_status": 4,
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Amount": bson.M{
					"$sum": "$amount",
				},
				"Count": bson.M{
					"$sum": 1,
				},
			},
		},
		{
			"$match": n,
		},
	}
	result := []bson.M{}
	pipe := Pays.Pipe(pipeline)
	pipe.All(&result)
	for _, item := range result {
		uid := item["_id"].(string)
		user_arr = append(user_arr, uid)
		pay_list = append(pay_list, item)
	}

	// 提现金额
	n1 := bson.M{}
	if withdraw_statr != 0 && withdraw_end != 0 {
		// 开始结束范围
		n1["Amount"] = bson.M{"$gte": withdraw_statr * 100, "$lt": withdraw_end * 100}
	} else if withdraw_statr != 0 {
		// 开始 从pay_statr - 无限
		n1["Amount"] = bson.M{"$gte": withdraw_statr * 100}
	} else if withdraw_end != 0 {
		// 开始 0- pay_end
		n1["Amount"] = bson.M{"$lt": withdraw_end * 100}
	}
	with_n := bson.M{}
	with_n["order_status"] = 2
	if pay_statr != 0 || pay_end != 0 {
		with_n["userid"] = bson.M{"$in": user_arr}
	}

	pipeline1 := []bson.M{
		{
			"$match": with_n,
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Amount": bson.M{
					"$sum": "$amount",
				},
				"Count": bson.M{
					"$sum": 1,
				},
			},
		},
		{
			"$match": n1,
		},
	}
	result1 := []bson.M{}
	pipe1 := Withdraws.Pipe(pipeline1)
	pipe1.All(&result1)
	if (pay_statr != 0 || pay_end != 0) && (withdraw_statr != 0 || withdraw_end != 0) {
		user_arr = make([]string, 0)
	}
	for _, item := range result1 {
		uid := item["_id"].(string)
		user_arr = append(user_arr, uid)
		withdraw_list = append(withdraw_list, item)
	}
	encountered := map[string]bool{} // 使用 map 存储已经遇到的元素
	user_dist := []string{}          // 存储去重后的结果
	// 去掉重复的userid
	for _, v := range user_arr {
		if !encountered[v] { // 如果当前元素在 map 中不存在，则添加到结果切片中，并在 map 中标记为已遇到
			encountered[v] = true
			user_dist = append(user_dist, v)
		}
	}

	// 充值金额
	// pipeline := []bson.M{
	// 	{
	// 		"$match": bson.M{
	// 			"order_status": 4,
	// 		},
	// 	},
	// 	{
	// 		"$group": bson.M{
	// 			"_id": "$userid",
	// 			"Amount": bson.M{
	// 				"$sum": "$amount",
	// 			},
	// 			"Count": bson.M{
	// 				"$sum": 1,
	// 			},
	// 		},
	// 	},
	// }
	// result := []bson.M{}
	// pipe := Pays.Pipe(pipeline)
	// pipe.All(&result)
	// for _, item := range result {
	// 	uid := item["_id"].(string)
	// 	amount := item["Amount"].(int)
	// 	amount_tmp := Chip2Float(utils.Int64(amount))
	// 	if pay_statr != 0 || pay_end != 0 {
	// 		if pay_statr != 0 && pay_end != 0 {
	// 			// 开始结束范围
	// 			if amount_tmp >= float64(pay_statr) && amount_tmp <= float64(pay_end) {
	// 				user_arr = append(user_arr, uid)
	// 				pay_list = append(pay_list, item)
	// 			}
	// 		} else if pay_statr != 0 {
	// 			// 开始 从pay_statr - 无限
	// 			if amount_tmp >= float64(pay_statr) {
	// 				user_arr = append(user_arr, uid)
	// 				pay_list = append(pay_list, item)
	// 			}
	// 		} else if pay_end != 0 {
	// 			// 开始 0- pay_end
	// 			if amount_tmp <= float64(pay_end) {
	// 				user_arr = append(user_arr, uid)
	// 				pay_list = append(pay_list, item)
	// 			}
	// 		}
	// 	}
	// }
	// // 提现金额
	// pipeline1 := []bson.M{
	// 	{
	// 		"$match": bson.M{
	// 			"order_status": 2,
	// 		},
	// 	},
	// 	{
	// 		"$group": bson.M{
	// 			"_id": "$userid",
	// 			"Amount": bson.M{
	// 				"$sum": "$amount",
	// 			},
	// 			"Count": bson.M{
	// 				"$sum": 1,
	// 			},
	// 		},
	// 	},
	// }
	// result1 := []bson.M{}
	// pipe1 := Withdraws.Pipe(pipeline1)
	// pipe1.All(&result1)
	// for _, item := range result1 {
	// 	uid := item["_id"].(string)
	// 	amount := item["Amount"].(int)
	// 	amount_tmp := Chip2Float(utils.Int64(amount))
	// 	if withdraw_statr != 0 || withdraw_end != 0 {
	// 		if withdraw_statr != 0 && withdraw_end != 0 {
	// 			// 开始结束范围
	// 			if amount_tmp >= float64(withdraw_statr) && amount_tmp <= float64(withdraw_end) {
	// 				user_arr = append(user_arr, uid)
	// 				withdraw_list = append(withdraw_list, item)
	// 			}
	// 		} else if withdraw_statr != 0 {
	// 			// 开始 withdraw_statr - 无限
	// 			if amount_tmp >= float64(withdraw_statr) {
	// 				user_arr = append(user_arr, uid)
	// 				withdraw_list = append(withdraw_list, item)
	// 			}
	// 		} else if withdraw_end != 0 {
	// 			// 开始 0- withdraw_end
	// 			if amount_tmp <= float64(withdraw_end) {
	// 				user_arr = append(user_arr, uid)
	// 				withdraw_list = append(withdraw_list, item)
	// 			}
	// 		}
	// 	}
	// }

	if len(user_dist) > 0 {
		channel_list := make([]entity.ChannelInfo, 0)
		Channels.Find(nil).All(&channel_list)
		// 根据筛选金额获得的userid数组查询相关用户信息
		for _, item := range user_dist {
			new_info := new(entity.UserAnalysisInfo)
			userid := item

			user_info := new(entity.PlayerUser)
			PlayerUsers.FindId(userid).One(user_info)
			if user_info != nil && user_info.Userid != "" {
				new_info.UserId = user_info.Userid
				c, _ := ConvertToIndiaTime(user_info.Ctime.Unix())
				new_info.RegistTime = c
				last, _ := ConvertToIndiaTime(user_info.LoginTime.Unix())
				new_info.LastLoginTime = last
				new_info.Channel = user_info.AD_BundleId
				// 存活时长
				time1 := new_info.RegistTime
				time2 := new_info.LastLoginTime
				// 将日期时间转换为整数天数
				date1 := time.Date(time1.Year(), time1.Month(), time1.Day(), 0, 0, 0, 0, time1.Location())
				date2 := time.Date(time2.Year(), time2.Month(), time2.Day(), 0, 0, 0, 0, time2.Location())

				// 计算相差的天数
				days := int(date2.Sub(date1).Hours() / 24)

				new_info.ActiveTime = int64(days)

				// 登录天数
				var loginLogs []entity.LogLogin
				LoginLogs.Find(bson.M{"userid": user_info.Userid}).Sort("login_time").All(&loginLogs)
				// 统计不同日期的登录次数
				loginDates := make(map[string]bool)
				for _, log := range loginLogs {
					loginDate := log.LoginTime.Format("2006-01-02") // 格式化为年-月-日
					loginDates[loginDate] = true
				}
				new_info.ActiveDay = int64(len(loginDates))
				// 获取渠道别名
				for _, channel := range channel_list {
					if channel.Name == user_info.AD_BundleId {
						new_info.Channel1 = channel.Name1
					}
				}
				// 关联账号查询
				// 查询设备号
				g_user_arr := make([]string, 0)
				m := bson.M{}
				m["ad__adid"] = user_info.AD_ADID
				m["_id"] = bson.M{"$ne": item}
				PlayerUsers.Find(m).Distinct("_id", &g_user_arr)
				// 查询提现银行卡
				num_arr := make([]string, 0)
				with_m := bson.M{}
				with_m["order_status"] = 2
				with_m["userid"] = item
				Withdraws.Find(with_m).Distinct("blank_number", &num_arr)

				w_user_arr := make([]string, 0)
				with_m["order_status"] = 2
				with_m["userid"] = bson.M{"$ne": item}
				with_m["blank_number"] = bson.M{"$in": num_arr}
				Withdraws.Find(with_m).Distinct("userid", &w_user_arr)

				relevance_temp := make(map[string]bool, 0)
				for _, v := range g_user_arr {
					relevance_temp[v] = true
				}
				for _, v := range w_user_arr {
					relevance_temp[v] = true
				}
				relevance_arr := make([]string, 0)
				for k, _ := range relevance_temp {
					relevance_arr = append(relevance_arr, k)
				}
				new_info.AssociatedUsers = relevance_arr
				new_info.AssociatedCount = int64(len(relevance_arr))
				// 获取首充日期及金额
				fpay_info := make([]entity.PayOrder, 0)
				first_pay_m := bson.M{}
				first_pay_m["userid"] = item
				first_pay_m["order_status"] = 4
				Pays.Find(first_pay_m).Sort("ctime").Limit(1).All(&fpay_info)
				if len(fpay_info) > 0 {
					tmp := fpay_info[0]
					if tmp.OrderID != "" {
						c, _ := ConvertToIndiaTime(tmp.Ctime.Unix())
						new_info.FirstPayDate = c
						new_info.FirstPayAmount = Chip2Float(int64(tmp.Amount))
					}
				}

				// 赋值充值信息
				for _, res := range pay_list {
					uid := res["_id"].(string)
					if uid == user_info.Userid {
						amount := res["Amount"].(int)
						count := res["Count"].(int)
						new_info.PayAmount = Chip2Float(int64(amount))
						new_info.PayCount = int64(count)
						new_info.PayAvg = Chip2Float(int64(amount)) / float64(count)
					}
				}
				// 赋值提现信息
				for _, res := range withdraw_list {
					uid := res["_id"].(string)
					if uid == user_info.Userid {
						amount := res["Amount"].(int)
						count := res["Count"].(int)
						new_info.WithdrawAmount = Chip2Float(int64(amount))
						new_info.WithdrawCount = int64(count)
						new_info.WithdrawAvg = Chip2Float(int64(amount)) / float64(count)
					}
				}

				// 游戏局数统计
				pipeline4 := []bson.M{
					{
						"$match": bson.M{
							"userid": item,
						},
					},
					{
						"$group": bson.M{
							"_id": "$gtype",
							"Count": bson.M{
								"$sum": 1,
							},
						},
					},
					{
						"$sort": bson.M{
							"Count": -1, // 按 Count 字段降序排序
						},
					},
					{
						"$limit": 2, // 获取前两条数据
					},
				}
				result4 := []bson.M{}
				pipe4 := LogGameTimes.Pipe(pipeline4)
				pipe4.All(&result4)
				if len(result4) >= 1 {
					g_count := result4[0]["Count"].(int)
					g_type := result4[0]["_id"].(int)
					new_info.GameName = GameByName(g_type)
					new_info.GameCount = int64(g_count)
				}
				if len(result4) >= 2 {
					g_count := result4[1]["Count"].(int)
					g_type := result4[1]["_id"].(int)
					new_info.TwoGameName = GameByName(g_type)
					new_info.TwoGameCount = int64(g_count)
				}

				list = append(list, *new_info)
			}

		}

	}
	// 按照注册时间倒序排序
	sort.Slice(list, func(i, j int) bool {
		return list[i].RegistTime.After(list[j].RegistTime)
	})
	return list, nil
}

func GameByName(gtype int) string {
	name := GtypeNameMap[gtype]
	return name
}

func (this *statisticsService) PlayerAnalysisCondition(
	begin, end *time.Time,
	userid string, utypeIds []int, moneyRange []int64, withdrawRange []int64, profitRange []int64,
	maxRoundsGtype, maxBetGtype int32,
) (where0 string, args0 []any) {
	if begin != nil {
		where0 += ` and t1.ctime >= ?`
		args0 = append(args0, *begin)
	}
	if end != nil {
		where0 += ` and t1.ctime <= ?`
		args0 = append(args0, *end)
	}
	if userid != "" {
		where0 += ` and t1.userid = ?`
		args0 = append(args0, userid)
	}
	if maxRoundsGtype != 0 {
		where0 += ` and t2.gtype_rounds = ?`
		args0 = append(args0, maxRoundsGtype)
	}
	if maxBetGtype != 0 {
		where0 += ` and t2.gtype_bets = ?`
		args0 = append(args0, maxBetGtype)
	}
	if len(moneyRange) == 2 {
		if moneyRange[0] > 0 {
			where0 += ` and t3.pay_amount >= ?`
			args0 = append(args0, moneyRange[0]*100)
		}
		if moneyRange[1] > 0 {
			where0 += ` and t3.pay_amount <= ?`
			args0 = append(args0, moneyRange[1]*100)
		}
	}
	if len(withdrawRange) == 2 {
		if withdrawRange[0] > 0 {
			where0 += ` and t4.withdraw_amount >= ?`
			args0 = append(args0, withdrawRange[0]*100)
		}
		if withdrawRange[1] > 0 {
			where0 += ` and t4.withdraw_amount <= ?`
			args0 = append(args0, withdrawRange[1]*100)
		}
	}
	if len(profitRange) == 2 {
		if profitRange[0] != 0 {
			where0 += ` and (t1.diamond + t4.withdraw_amount - t3.pay_amount) >= ?`
			args0 = append(args0, profitRange[0]*100)
		}
		if profitRange[1] != 0 {
			where0 += ` and (t1.diamond + t4.withdraw_amount - t3.pay_amount) <= ?`
			args0 = append(args0, profitRange[1]*100)
		}
	}

	if len(utypeIds) > 0 {
		var utypeFilter []string
		for _, utype := range utypeIds {
			switch utype {
			case 10: // 新手
				utypeFilter = append(utypeFilter, `(t3.pay_amount = 0 and t0.state = 1)`)
			case 1: // 平民
				utypeFilter = append(utypeFilter, `(t3.pay_amount = 0 and t0.state != 1)`)
			case 2: // 普R
				utypeFilter = append(utypeFilter, `t3.pay_amount between 1 and 100000-1`)
			case 3: // 小R
				utypeFilter = append(utypeFilter, `t3.pay_amount between 100000 and 1000000-1`)
			case 4: // 中R
				utypeFilter = append(utypeFilter, `t3.pay_amount between 1000000 and 10000000-1`)
			case 5: // 大R
				utypeFilter = append(utypeFilter, `t3.pay_amount between 10000000 and 20000000-1`)
			case 6: // 超大R
				utypeFilter = append(utypeFilter, `t3.pay_amount >= 20000000`)
			}
		}
		if len(utypeFilter) > 0 {
			where0 += fmt.Sprintf(" and (%s)", strings.Join(utypeFilter, " OR "))
		}
	}
	return
}

// 玩家分析总表
func (this *statisticsService) GetUserRegistryTime(userid string) (registryTime time.Time, err error) {
	var user map[string]any
	err = ck.Select(&user, `SELECT ctime FROM game.col_user FINAL WHERE userid = ?`, userid)
	if err != nil {
		return
	}
	registryTime = user["ctime"].(time.Time)
	return
}

// 玩家分析总表
func (this *statisticsService) PlayerAnalysis(
	page, pageSize int,
	begin, end *time.Time,
	userid string, utypeIds []int, moneyRange []int64, withdrawRange []int64, profitRange []int64,
	maxRoundsGtype, maxBetGtype int32,
) (count int, list []*entity.PlayerAnalysisInfo, err error) {
	sql0 := `
		SELECT t1.userid userid, t1.ctime ctime, t1.diamond diamond, t2.gtype_rounds, t2.gtype_bets, 
			t3.pay_times_try, t3.pay_times, t3.pay_amount,
			t4.withdraw_times_try, t4.withdraw_times, t4.withdraw_amount,
			t5.ctime first_pay_time, t5.amount first_pay_amount,
			t6.ctime first_withdraw_time, t6.amount first_withdraw_amount,
			t1.diamond + t4.withdraw_amount - t3.pay_amount profit
		FROM game.col_user_finance t1 FINAL 
		JOIN game.col_user t0 FINAL ON t1.userid = t0.userid
		LEFT JOIN (
			SELECT userid, gtype, count(*) rounds, sum(bet_amount) bet_sum,
			first_value(gtype) OVER (PARTITION BY userid ORDER BY rounds DESC) AS gtype_rounds,
			first_value(gtype) OVER (PARTITION BY userid ORDER BY bet_sum DESC) AS gtype_bets
			FROM (
				SELECT id,userid,gtype,bet_amount
				FROM game.col_detail t1 FINAL WHERE robot = 0 and win_type in (1,2,3) 
					UNION ALL
				SELECT round_id id,user_id userid, game_id gtype,SUM(amount) bet_amount
				FROM game.col_nsq_log_external_bet t2 FINAL WHERE amount != 0 
				GROUP BY id, user_id, gtype
			) s1
			GROUP BY userid, gtype
			LIMIT 1 BY userid
		) t2 ON t1.userid = t2.userid
		LEFT JOIN (
			SELECT userid, count(*) pay_times_try, 
				SUM(CASE WHEN order_status = 4 THEN 1 ELSE 0 END) pay_times,
				SUM(CASE WHEN order_status = 4 THEN amount ELSE 0 END) pay_amount
			FROM game.col_trade_record FINAL
			GROUP BY userid
		) t3 ON t1.userid = t3.userid
		LEFT JOIN (
			SELECT userid, count(*) withdraw_times_try, 
				SUM(CASE WHEN order_status = 2 THEN 1 ELSE 0 END) withdraw_times,
				SUM(CASE WHEN order_status = 2 THEN amount ELSE 0 END) withdraw_amount
			FROM game.col_withdraw_record FINAL
			GROUP BY userid
		) t4 ON t1.userid = t4.userid
		LEFT JOIN (
			SELECT userid, ctime, amount FROM game.col_trade_record FINAL 
			WHERE order_status = 4 ORDER BY userid, ctime LIMIT 1 BY userid
		) t5 ON t1.userid = t5.userid
		LEFT JOIN (
			SELECT userid, ctime, amount FROM game.col_withdraw_record FINAL 
			WHERE order_status = 2 ORDER BY userid, ctime LIMIT 1 BY userid
		) t6 ON t1.userid = t6.userid
		WHERE 1 = 1 %s
		ORDER BY t1.ctime DESC
		LIMIT ?, ?
	`
	sql_count := `
		SELECT count(*) count
		FROM game.col_user_finance t1 FINAL 
		JOIN game.col_user t0 FINAL ON t1.userid = t0.userid
		LEFT JOIN (
			SELECT userid, gtype, count(*) rounds, sum(bet_amount) bet_sum,
			first_value(gtype) OVER (PARTITION BY userid ORDER BY rounds DESC) AS gtype_rounds,
			first_value(gtype) OVER (PARTITION BY userid ORDER BY bet_sum DESC) AS gtype_bets
			FROM (
				SELECT id,userid,gtype,bet_amount
				FROM game.col_detail t1 FINAL WHERE robot = 0 and win_type in (1,2,3) 
					UNION ALL
				SELECT round_id id,user_id userid, game_id gtype,SUM(amount) bet_amount
				FROM game.col_nsq_log_external_bet t2 FINAL WHERE amount != 0 
				GROUP BY id, user_id, gtype
			) s1
			GROUP BY userid, gtype
			LIMIT 1 BY userid
		) t2 ON t1.userid = t2.userid
		LEFT JOIN (
			SELECT userid, count(*) pay_times_try, 
				SUM(CASE WHEN order_status = 4 THEN 1 ELSE 0 END) pay_times,
				SUM(CASE WHEN order_status = 4 THEN amount ELSE 0 END) pay_amount
			FROM game.col_trade_record FINAL
			GROUP BY userid
		) t3 ON t1.userid = t3.userid
		LEFT JOIN (
			SELECT userid, count(*) withdraw_times_try, 
				SUM(CASE WHEN order_status = 2 THEN 1 ELSE 0 END) withdraw_times,
				SUM(CASE WHEN order_status = 2 THEN amount ELSE 0 END) withdraw_amount
			FROM game.col_withdraw_record FINAL
			GROUP BY userid
		) t4 ON t1.userid = t4.userid
		WHERE 1 = 1 %s
	`

	where0, args0 := this.PlayerAnalysisCondition(begin, end, userid, utypeIds, moneyRange, withdrawRange, profitRange, maxRoundsGtype, maxBetGtype)

	err = ck.Select(&count, fmt.Sprintf(sql_count, where0), args0...)
	if err != nil {
		return
	}
	if count == 0 {
		return
	}

	// 汇总查询
	var summaryChan = make(chan *entity.PlayerAnalysisInfo)
	go func() {
		start := time.Now()
		summary, err := this.PlayerAnalysisSummary(where0, args0)
		fmt.Printf("=======summary use %dms\n", time.Now().UnixMilli()-start.UnixMilli())
		if err != nil {
			beego.Error("player analysis summary error:", err)
			return
		}
		summaryChan <- summary
	}()

	offset, limit := PageCalc(page, pageSize)
	var datas0 []map[string]any
	err = ck.Select(&datas0, fmt.Sprintf(sql0, where0), append(args0, offset, limit)...)
	if err != nil {
		return
	}

	var userids []string
	var useridDatas = make(map[string]*entity.PlayerAnalysisInfo)
	for _, data := range datas0 {
		userid := data["userid"].(string)
		ctime := data["ctime"].(time.Time)

		diamond := utils.ToInt64(data["diamond"])
		// gtype_rounds := utils.ToInt64(data["gtype_rounds"])
		// gtype_bets := utils.ToInt64(data["gtype_bets"])
		pay_times_try := utils.ToInt64(data["pay_times_try"])
		pay_times := utils.ToInt64(data["pay_times"])
		pay_amount := utils.ToInt64(data["pay_amount"])
		withdraw_times_try := utils.ToInt64(data["withdraw_times_try"])
		withdraw_times := utils.ToInt64(data["withdraw_times"])
		withdraw_amount := utils.ToInt64(data["withdraw_amount"])
		profit := utils.ToInt64(data["profit"])
		first_pay_time := data["first_pay_time"].(time.Time)
		first_pay_amount := utils.ToInt64(data["first_pay_amount"])
		first_withdraw_time := data["first_withdraw_time"].(time.Time)
		first_withdraw_amount := utils.ToInt64(data["first_withdraw_amount"])

		userids = append(userids, userid)
		item := &entity.PlayerAnalysisInfo{
			UserId:              userid,
			Money:               pay_amount,
			RegistTime:          ctime.Format(utils.FORMAT),
			RewardRate:          fmt.Sprintf("%.2f%%", utils.CaseElse(pay_amount == 0, 0, float64(withdraw_amount+diamond)/float64(pay_amount)*100)),
			Profit:              fmt.Sprintf("%.2f", float64(profit)/100),
			FirstPayDate:        utils.CaseElse(first_pay_amount == 0, "未充值", first_pay_time.Format(utils.FORMAT)),
			FirstPayAmount:      fmt.Sprintf("%.2f", float64(first_pay_amount)/100),
			PayAmount:           fmt.Sprintf("%.2f", float64(pay_amount)/100),
			PayCount:            pay_times,
			PayAvg:              fmt.Sprintf("%.2f", utils.CaseElse(pay_times == 0, 0, float64(pay_amount)/float64(pay_times)/100)),
			PaySuccessRate:      fmt.Sprintf("%.2f%%", utils.CaseElse(pay_times_try == 0, 0, float64(pay_times)/float64(pay_times_try)*100)),
			FirstWithdrawDate:   utils.CaseElse(first_withdraw_amount == 0, "未提现", first_withdraw_time.Format(utils.FORMAT)),
			FirstWithdrawAmount: fmt.Sprintf("%.2f", float64(first_withdraw_amount)/100),
			WithdrawAmount:      fmt.Sprintf("%.2f", float64(withdraw_amount)/100),
			WithdrawCount:       withdraw_times,
			WithdrawAvg:         fmt.Sprintf("%.2f", utils.CaseElse(withdraw_times == 0, 0, float64(withdraw_amount)/float64(withdraw_times)/100)),
			WithdrawSuccessRate: fmt.Sprintf("%.2f%%", utils.CaseElse(withdraw_times_try == 0, 0, float64(withdraw_times)/float64(withdraw_times_try)*100)),
		}
		item.FirstPayEfficience = utils.CaseElse(first_pay_amount == 0, "未充值", fmt.Sprint((first_pay_time.Unix()-ctime.Unix())/60))
		item.PayWithdrawDiffDay = utils.CaseElse(first_pay_amount == 0, "未充值", utils.CaseElse(first_withdraw_amount == 0, "未提现", fmt.Sprint((first_withdraw_time.Unix()-first_pay_time.Unix())/60)))
		item.PayWithdrawTimes = fmt.Sprintf("%.2f", utils.CaseElse(withdraw_times == 0, 0, float64(pay_times)/float64(withdraw_times)))

		list = append(list, item)
		useridDatas[userid] = item
	}

	// 每次充值前携带金额均值
	if len(userids) > 0 {
		var datas_PayBeforeCarry []map[string]any
		err = ck.Select(&datas_PayBeforeCarry, `
			SELECT userid, avg(old_diamond) old_diamond_avg FROM game.col_log_water FINAL
			WHERE ltype IN (4, 74, 84, 86, 119, 80, 57, 58, 59, 129) AND userid IN ?
			GROUP BY userid
		`, userids)
		if err != nil {
			return
		}
		for _, data := range datas_PayBeforeCarry {
			userid := data["userid"].(string)
			if item, ok := useridDatas[userid]; ok {
				old_diamond_avg := utils.ToFloat64(data["old_diamond_avg"])
				item.PayBeforeCarry = fmt.Sprintf("%.2f", old_diamond_avg/100)
			}
		}

		// 用户基本信息
		err = this.PlayerAnalysis_UserInfo(userids, useridDatas)
		if err != nil {
			return
		}
		// 单笔最大充值提现金额/日期
		err = this.PlayerAnalysis_MaxPayDate(userids, useridDatas)
		if err != nil {
			return
		}
		// 单日累计最大充值金额
		err = this.PlayerAnalysis_MaxPayDateSuccessRate(userids, useridDatas)
		if err != nil {
			return
		}
		// 单日最高充值次数/日期/成功率
		err = this.PlayerAnalysis_MaxPayTimesDateSuccessRate(userids, useridDatas)
		if err != nil {
			return
		}
		// 玩家分析：玩最多的游戏...
		err = this.PlayerAnalysis_GameStats(userids, useridDatas)
		if err != nil {
			return
		}
		// 最大单局赢输钱金额/所在游戏/日期
		err = this.PlayerAnalysis_MaxWinStats(userids, useridDatas)
		if err != nil {
			return
		}

		// 活跃天数 日均在线时长
		err = this.PlayerAnalysis_ActiveDays(userids, useridDatas)
		if err != nil {
			return
		}
	}

	summary := <-summaryChan
	list = append([]*entity.PlayerAnalysisInfo{summary}, list...)
	return
}

func GetChargeType(money int64, state int) (int, string) {
	if money == 0 {
		if state == 1 { // state=1 新手
			return 10, "新手"
		}
		return 1, "平民"
	} else if money >= 1 && money < 100000 {
		return 2, "普R"
	} else if money >= 100000 && money < 1000000 {
		return 3, "小R"
	} else if money >= 1000000 && money < 10000000 {
		return 4, "中R"
	} else if money >= 10000000 && money < 20000000 {
		return 5, "大R"
	} else {
		return 6, "超大R"
	}
}

// 玩家分析：基本信息
func (this *statisticsService) PlayerAnalysis_UserInfo(userids []string, useridDatas map[string]*entity.PlayerAnalysisInfo) (err error) {
	var datas_UserInfo []map[string]any
	err = ck.Select(&datas_UserInfo, `
		SELECT t1.userid userid, t1.ad__bundle_id, t1.state, t1.ctime, t1.login_time, t1.ad__adid, t2.ips, t3.banks
		FROM game.col_user t1 FINAL
		LEFT JOIN (
			SELECT userid, groupArray(DISTINCT ip) ips
			FROM game.col_log_login FINAL
			WHERE userid IN ?
			GROUP BY userid
		) t2 ON t1.userid = t2.userid
		LEFT JOIN (
			SELECT userid, groupArray(DISTINCT blank_number) banks
			FROM game.col_withdraw_record FINAL 
			WHERE order_status = 2 AND userid IN ?
			GROUP BY userid
		) t3 ON t1.userid = t3.userid
		WHERE t1.userid IN ?
		`, userids, userids, userids)
	if err != nil {
		return
	}

	var now = NowTime()
	var allDevicesIds, allIps, allBanks, channelIds []string
	for _, data := range datas_UserInfo {
		userid := data["userid"].(string)
		if item, ok := useridDatas[userid]; ok {
			ad__bundle_id := data["ad__bundle_id"].(string)
			ctime := data["ctime"].(time.Time)
			login_time := data["login_time"].(time.Time)
			ad__adid := data["ad__adid"].(string)
			ips := data["ips"].([]string)
			banks := data["banks"].([]string)
			state := utils.ToInt64(data["state"])

			_, chargeType := GetChargeType(item.Money, int(state))
			item.UserType = chargeType
			item.Channel = ad__bundle_id
			item.RegistTime = ctime.Format(utils.FORMAT)
			item.LastLoginTime = login_time.Format(utils.FORMAT)
			if login_time.After(ctime) {
				item.LoseDays = fmt.Sprint((now.Unix() - login_time.Unix()) / 60 / 60 / 24)
				item.ActiveTime = fmt.Sprint((login_time.Unix() - ctime.Unix()) / 60 / 60 / 24)
			}
			item.AD__ADID = ad__adid
			item.IPs = ips
			item.Banks = banks

			allDevicesIds = append(allDevicesIds, ad__adid)
			allIps = append(allIps, ips...)
			allBanks = append(allBanks, banks...)
			channelIds = append(channelIds, ad__bundle_id)
		}
	}

	// 查询渠道
	if len(channelIds) > 0 {
		channels, err1 := ChannelService.GetChannelByIds(channelIds)
		if err1 != nil {
			err = err1
			return
		}
		var channelsMap = make(map[string]entity.ChannelInfo)
		for _, ch := range channels {
			channelsMap[ch.Name] = ch
		}
		for _, item := range useridDatas {
			if ch, ok := channelsMap[item.Channel]; ok {
				item.ChannelClass = ch.ClassName
				item.Channel1 = ch.Name1
			}
		}
	}

	// 关联账号...
	// 设备号
	var deviceIdUsers = make(map[string][]string)
	if len(allDevicesIds) > 0 {
		var datas_devices []map[string]any
		err = ck.Select(&datas_devices, `
			SELECT DISTINCT userid, ad__adid
			FROM game.col_user t1 FINAL
			WHERE ad__adid IN ?
		`, allDevicesIds)
		if err != nil {
			return
		}
		for _, item := range datas_devices {
			userid := item["userid"].(string)
			ad__adid := item["ad__adid"].(string)
			if ad__adid != "" {
				deviceIdUsers[ad__adid] = append(deviceIdUsers[ad__adid], userid)
			}
		}
	}
	// ip
	var ipUsers = make(map[string][]string)
	if len(allIps) > 0 {
		var datas_ips []map[string]any
		err = ck.Select(&datas_ips, `
			SELECT DISTINCT userid, ip
			FROM game.col_log_login FINAL
			WHERE ip IN ?
		`, allIps)
		if err != nil {
			return
		}
		for _, item := range datas_ips {
			userid := item["userid"].(string)
			ip := item["ip"].(string)
			if ip != "" {
				ipUsers[ip] = append(ipUsers[ip], userid)
			}
		}
	}
	// 银行卡号
	var bankUsers = make(map[string][]string)
	if len(allBanks) > 0 {
		var datas_banks []map[string]any
		err = ck.Select(&datas_banks, `
		SELECT DISTINCT userid, blank_number
		FROM game.col_withdraw_record FINAL 
		WHERE order_status = 2 AND blank_number IN ?
	`, allBanks)
		if err != nil {
			return
		}
		for _, item := range datas_banks {
			userid := item["userid"].(string)
			blank := item["blank_number"].(string)
			if blank != "" {
				bankUsers[blank] = append(bankUsers[blank], userid)
			}
		}
	}

	for userid, item := range useridDatas {
		var refUserids = make(map[string]bool)
		if userids, ok := deviceIdUsers[item.AD__ADID]; ok && item.AD__ADID != "" {
			for _, userid := range userids {
				refUserids[userid] = true
			}
		}
		for _, ip := range item.IPs {
			if userids, ok := ipUsers[ip]; ok && ip != "" {
				for _, userid := range userids {
					refUserids[userid] = true
				}
			}
		}
		for _, bank := range item.Banks {
			if userids, ok := bankUsers[bank]; ok && bank != "" {
				for _, userid := range userids {
					refUserids[userid] = true
				}
			}
		}
		var refs []string
		for refUserid := range refUserids {
			if refUserid != userid {
				refs = append(refs, refUserid)
			}
		}
		item.RefUsers = strings.Join(refs, ",")
	}
	return
}

// 玩家分析：单笔最大充值提现金额/日期
func (this *statisticsService) PlayerAnalysis_MaxPayDate(userids []string, useridDatas map[string]*entity.PlayerAnalysisInfo) (err error) {
	var datas_MaxPayDate []map[string]any
	err = ck.Select(&datas_MaxPayDate, `
			SELECT userid, amount, arraySort(groupArray(DISTINCT toYYYYMMDD(ctime))) dates
				FROM game.col_trade_record FINAL 
			WHERE order_status = 4 AND userid IN ?
			GROUP BY userid, amount
			ORDER BY userid, amount desc
			LIMIT 1 BY userid
		`, userids)
	if err != nil {
		return
	}
	for _, data := range datas_MaxPayDate {
		userid := data["userid"].(string)
		if item, ok := useridDatas[userid]; ok {
			amount := utils.ToInt64(data["amount"])
			if amount > 0 {
				dates := data["dates"].([]uint32)
				var fmtDates []string
				for _, date := range dates {
					datestr := fmt.Sprint(date)
					sdate := datestr[0:4] + "年" + datestr[4:6] + "月" + datestr[6:] + "日"
					fmtDates = append(fmtDates, sdate)
				}
				item.MaxPayDate = fmt.Sprintf("%.0f/", float64(amount)/100.0) + strings.Join(fmtDates, "、")
			}
		}
	}

	var datas_MaxWithdrawDate []map[string]any
	err = ck.Select(&datas_MaxWithdrawDate, `
		SELECT userid, amount, arraySort(groupArray(DISTINCT toYYYYMMDD(ctime))) dates
			FROM game.col_withdraw_record FINAL 
		WHERE order_status = 2 AND userid IN ?
		GROUP BY userid, amount
		ORDER BY userid, amount desc
		LIMIT 1 BY userid
	`, userids)
	if err != nil {
		return
	}
	for _, data := range datas_MaxWithdrawDate {
		userid := data["userid"].(string)
		if item, ok := useridDatas[userid]; ok {
			amount := utils.ToInt64(data["amount"])
			if amount > 0 {
				dates := data["dates"].([]uint32)
				var fmtDates []string
				for _, date := range dates {
					datestr := fmt.Sprint(date)
					sdate := datestr[0:4] + "年" + datestr[4:6] + "月" + datestr[6:] + "日"
					fmtDates = append(fmtDates, sdate)
				}
				item.MaxWithdrawDate = fmt.Sprintf("%.0f/", float64(amount)/100.0) + strings.Join(fmtDates, "、")
			}
		}
	}
	return
}

// 玩家分析：单日累计最大充值提现金额/日期/成功率
func (this *statisticsService) PlayerAnalysis_MaxPayDateSuccessRate(userids []string, useridDatas map[string]*entity.PlayerAnalysisInfo) (err error) {
	var datas_MaxPayDateSuccessRate []map[string]any
	err = ck.Select(&datas_MaxPayDateSuccessRate, `
		SELECT userid, amount_day, groupArray(pay_day) pay_dates, groupArray(pay_success_rate_day) pay_success_rates 
		FROM (
			SELECT userid, toYYYYMMDD(ctime) pay_day, 
				SUM(CASE WHEN order_status = 4 THEN amount ELSE 0 END) amount_day,  
				SUM(CASE WHEN order_status = 4 THEN 1 ELSE 0 END) / count(*) pay_success_rate_day
			FROM game.col_trade_record FINAL 
			WHERE userid IN ?
			GROUP BY userid, pay_day
			HAVING amount_day > 0
		) t1
		GROUP BY userid, amount_day
		ORDER BY userid, amount_day desc
		LIMIT 1 BY userid
		`, userids)
	if err != nil {
		return
	}
	for _, data := range datas_MaxPayDateSuccessRate {
		userid := data["userid"].(string)
		if item, ok := useridDatas[userid]; ok {
			amount_day := utils.ToInt64(data["amount_day"])
			if amount_day > 0 {
				pay_dates := data["pay_dates"].([]uint32)
				pay_success_rates := data["pay_success_rates"].([]float64)
				var fmtDates, fmtRates []string
				for i, pay_date := range pay_dates {
					pay_success_rate := pay_success_rates[i]
					datestr := fmt.Sprint(pay_date)
					sdate := datestr[0:4] + "年" + datestr[4:6] + "月" + datestr[6:] + "日"
					fmtDates = append(fmtDates, sdate)
					fmtRates = append(fmtRates, fmt.Sprintf("%.2f%%", pay_success_rate*100))
				}
				item.MaxPayDateSuccessRate = fmt.Sprintf("%.0f/", float64(amount_day)/100.0) + strings.Join(fmtDates, "、") + "/" + strings.Join(fmtRates, "、")
			}
		}
	}

	// 单日累计最大提现金额/日期/成功率
	var datas_MaxWithdrawDateSuccessRate []map[string]any
	err = ck.Select(&datas_MaxWithdrawDateSuccessRate, `
		SELECT userid, amount_day, groupArray(pay_day) pay_dates, groupArray(pay_success_rate_day) pay_success_rates 
		FROM (
			SELECT userid, toYYYYMMDD(ctime) pay_day, 
				SUM(CASE WHEN order_status = 2 THEN amount ELSE 0 END) amount_day,  
				SUM(CASE WHEN order_status = 2 THEN 1 ELSE 0 END) / count(*) pay_success_rate_day
			FROM game.col_withdraw_record FINAL 
			WHERE userid IN ?
			GROUP BY userid, pay_day
			HAVING amount_day > 0
		) t1
		GROUP BY userid, amount_day
		ORDER BY userid, amount_day desc
		LIMIT 1 BY userid
		`, userids)
	if err != nil {
		return
	}
	for _, data := range datas_MaxWithdrawDateSuccessRate {
		userid := data["userid"].(string)
		if item, ok := useridDatas[userid]; ok {
			amount_day := utils.ToInt64(data["amount_day"])
			if amount_day > 0 {
				pay_dates := data["pay_dates"].([]uint32)
				pay_success_rates := data["pay_success_rates"].([]float64)
				var fmtDates, fmtRates []string
				for i, pay_date := range pay_dates {
					pay_success_rate := pay_success_rates[i]
					datestr := fmt.Sprint(pay_date)
					sdate := datestr[0:4] + "年" + datestr[4:6] + "月" + datestr[6:] + "日"
					fmtDates = append(fmtDates, sdate)
					fmtRates = append(fmtRates, fmt.Sprintf("%.2f%%", pay_success_rate*100))
				}
				item.MaxWithdrawDateSuccessRate = fmt.Sprintf("%.0f/", float64(amount_day)/100.0) + strings.Join(fmtDates, "、") + "/" + strings.Join(fmtRates, "、")
			}
		}
	}
	return
}

// 玩家分析：单日最高充值提现次数/日期/成功率
func (this *statisticsService) PlayerAnalysis_MaxPayTimesDateSuccessRate(userids []string, useridDatas map[string]*entity.PlayerAnalysisInfo) (err error) {
	var datas_MaxPayTimesDateSuccessRate []map[string]any
	err = ck.Select(&datas_MaxPayTimesDateSuccessRate, `
		SELECT userid, pay_times_day, groupArray(pay_day) pay_dates, groupArray(pay_success_rate_day) pay_success_rates
		FROM (
			SELECT userid, toYYYYMMDD(ctime) pay_day, 
				SUM(CASE WHEN order_status = 4 THEN 1 ELSE 0 END) pay_times_day,
				pay_times_day / count(*) pay_success_rate_day
			FROM game.col_trade_record FINAL 
			WHERE userid IN ?
			GROUP BY userid, pay_day
			HAVING pay_times_day > 0
		) t1
		GROUP BY userid, pay_times_day
		ORDER BY userid, pay_times_day desc
		LIMIT 1 BY userid
	`, userids)
	if err != nil {
		return
	}
	for _, data := range datas_MaxPayTimesDateSuccessRate {
		userid := data["userid"].(string)
		if item, ok := useridDatas[userid]; ok {
			pay_times_day := utils.ToInt64(data["pay_times_day"])
			pay_dates := data["pay_dates"].([]uint32)
			pay_success_rates := data["pay_success_rates"].([]float64)
			var fmtDates, fmtRates []string
			for i, pay_date := range pay_dates {
				pay_success_rate := pay_success_rates[i]
				datestr := fmt.Sprint(pay_date)
				sdate := datestr[0:4] + "年" + datestr[4:6] + "月" + datestr[6:] + "日"
				fmtDates = append(fmtDates, sdate)
				fmtRates = append(fmtRates, fmt.Sprintf("%.2f%%", pay_success_rate*100))
			}
			item.MaxPayTimesDateSuccessRate = fmt.Sprintf("%d/", pay_times_day) + strings.Join(fmtDates, "、") + "/" + strings.Join(fmtRates, "、")
		}
	}

	// 单日最高提现次数/日期/成功率
	var datas_MaxWithdrawTimesDateSuccessRate []map[string]any
	err = ck.Select(&datas_MaxWithdrawTimesDateSuccessRate, `
		SELECT userid, pay_times_day, groupArray(pay_day) pay_dates, groupArray(pay_success_rate_day) pay_success_rates
		FROM (
			SELECT userid, toYYYYMMDD(ctime) pay_day, 
				SUM(CASE WHEN order_status = 2 THEN 1 ELSE 0 END) pay_times_day,
				pay_times_day / count(*) pay_success_rate_day
			FROM game.col_withdraw_record FINAL 
			WHERE userid IN ?
			GROUP BY userid, pay_day
			HAVING pay_times_day > 0
		) t1
		GROUP BY userid, pay_times_day
		ORDER BY userid, pay_times_day desc
		LIMIT 1 BY userid
	`, userids)
	if err != nil {
		return
	}
	for _, data := range datas_MaxWithdrawTimesDateSuccessRate {
		userid := data["userid"].(string)
		if item, ok := useridDatas[userid]; ok {
			pay_times_day := utils.ToInt64(data["pay_times_day"])
			pay_dates := data["pay_dates"].([]uint32)
			pay_success_rates := data["pay_success_rates"].([]float64)
			var fmtDates, fmtRates []string
			for i, pay_date := range pay_dates {
				pay_success_rate := pay_success_rates[i]
				datestr := fmt.Sprint(pay_date)
				sdate := datestr[0:4] + "年" + datestr[4:6] + "月" + datestr[6:] + "日"
				fmtDates = append(fmtDates, sdate)
				fmtRates = append(fmtRates, fmt.Sprintf("%.2f%%", pay_success_rate*100))
			}
			item.MaxWithdrawTimesDateSuccessRate = fmt.Sprintf("%d/", pay_times_day) + strings.Join(fmtDates, "、") + "/" + strings.Join(fmtRates, "、")
		}
	}
	return
}

// 玩家分析：玩最多的游戏...
func (this *statisticsService) PlayerAnalysis_GameStats(userids []string, useridDatas map[string]*entity.PlayerAnalysisInfo) (err error) {
	var datas_GameStats []map[string]any
	err = ck.Select(&datas_GameStats, `
		SELECT * FROM (
			SELECT userid, gtype, count(*) rounds, SUM(bet_amount) bet_sum,  AVG(bet_amount) bet_avg, 
				SUM(CASE WHEN win_type = 1 THEN score ELSE 0 END) win_amount,
				SUM(CASE WHEN win_type = 2 THEN score ELSE 0 END) lose_amount,
				SUM(CASE WHEN win_type = 1 THEN 1 ELSE 0 END) win_rounds,
				SUM(CASE WHEN win_type = 2 THEN 1 ELSE 0 END) lose_rounds,
				row_number() OVER (PARTITION BY userid ORDER BY rounds desc) AS round_no,
				row_number() OVER (PARTITION BY userid ORDER BY bet_sum desc) AS bet_no
			FROM (
					SELECT id, userid, gtype, bet_amount, win_type, score
					FROM game.col_detail t1 FINAL WHERE robot = 0 and win_type in (1,2,3) AND userid IN ?
						UNION ALL
					SELECT s2.round_id id, s2.user_id userid, s2.game_id gtype, SUM(s2.amount_sum) bet_amount,
						(CASE WHEN score > 0 THEN 1 WHEN score == 0 THEN 3 ELSE 2 END) win_type, 
						SUM(s3.amount_sum) - SUM(s2.amount_sum) score
					FROM (SELECT round_id, user_id, game_id, SUM(amount) amount_sum FROM game.col_nsq_log_external_bet FINAL WHERE amount != 0  GROUP BY round_id, user_id, game_id) s2
					LEFT JOIN (SELECT round_id, user_id, SUM(amount) amount_sum FROM game.col_nsq_log_external_reward FINAL WHERE amount != 0 GROUP BY round_id, user_id) s3
						ON s2.round_id = s3.round_id AND s2.user_id = s3.user_id
					WHERE s2.user_id IN ?
					GROUP BY s2.round_id, s2.user_id, s2.game_id
			) s1
			GROUP BY userid, gtype
		) tt
		WHERE tt.round_no <= 3 OR tt.bet_no <= 3
		ORDER BY userid, rounds DESC
	`, userids, userids)
	if err != nil {
		return
	}
	for _, data := range datas_GameStats {
		userid := data["userid"].(string)
		if item, ok := useridDatas[userid]; ok {
			round_no := utils.ToInt64(data["round_no"])
			bet_no := utils.ToInt64(data["bet_no"])
			gtype := utils.ToInt64(data["gtype"])
			rounds := utils.ToInt64(data["rounds"])
			// bet_sum := utils.ToInt64(data["bet_sum"])
			bet_avg := utils.ToFloat64(data["bet_avg"])
			win_amount := utils.ToInt64(data["win_amount"])
			lose_amount := utils.ToInt64(data["lose_amount"])
			win_rounds := utils.ToInt64(data["win_rounds"])
			lose_rounds := utils.ToInt64(data["lose_rounds"])

			gameName, ok := GtypeNameMap[int(gtype)]
			if !ok {
				gameName = fmt.Sprint(gtype)
			}
			// 游戏/局数/局均码量/返奖率
			stats := fmt.Sprintf("%s/%d/%.2f/%.2f%%", gameName, rounds, bet_avg/100, utils.CaseElse(lose_amount == 0, 0, float64(win_amount)/float64(-lose_amount)*100))

			switch round_no {
			case 1:
				if win_rounds > 0 {
					item.Game1PlayWinsAvg = fmt.Sprintf("%.2f", float64(win_amount)/float64(win_rounds)/100)
				}
				if lose_rounds > 0 {
					item.Game1PlayLosesAvg = fmt.Sprintf("%.2f", float64(lose_amount)/float64(lose_rounds)/100)
				}
				item.Game1TimesStats = stats
			case 2:
				item.Game2TimesStats = stats
			case 3:
				item.Game3TimesStats = stats
			}
			switch bet_no {
			case 1:
				if win_rounds > 0 {
					item.Game1BetWinsAvg = fmt.Sprintf("%.2f", float64(win_amount)/float64(win_rounds)/100)
				}
				if lose_rounds > 0 {
					item.Game1BetLosesAvg = fmt.Sprintf("%.2f", float64(lose_amount)/float64(lose_rounds)/100)
				}
				item.Game1BetStats = stats
			case 2:
				item.Game2BetStats = stats
			case 3:
				item.Game3BetStats = stats
			}
		}
	}

	return
}

// 玩家分析： 最大单局赢钱输钱金额/所在游戏/日期
func (this *statisticsService) PlayerAnalysis_MaxWinStats(userids []string, useridDatas map[string]*entity.PlayerAnalysisInfo) (err error) {
	sql0 := `
		SELECT id, userid, score, gtype, toYYYYMMDD(toDateTime(begin_time)) gdate
		FROM (
			SELECT id, userid, gtype, bet_amount, win_type, score, begin_time
			FROM game.col_detail t1 FINAL WHERE robot = 0 and win_type in (1,2,3) AND userid IN ?
				UNION ALL
			SELECT s2.round_id id, s2.user_id userid, s2.game_id gtype, SUM(s2.amount_sum) bet_amount,
				(CASE WHEN score > 0 THEN 1 WHEN score == 0 THEN 3 ELSE 2 END) win_type, 
				SUM(s3.amount_sum) - SUM(s2.amount_sum) score, s2.begin_time
			FROM (SELECT round_id, user_id, game_id, SUM(amount) amount_sum, MIN(ctime) begin_time FROM game.col_nsq_log_external_bet FINAL WHERE amount != 0  GROUP BY round_id, user_id, game_id) s2
			LEFT JOIN (SELECT round_id, user_id, SUM(amount) amount_sum FROM game.col_nsq_log_external_reward FINAL WHERE amount != 0 GROUP BY round_id, user_id) s3
				ON s2.round_id = s3.round_id AND s2.user_id = s3.user_id
			WHERE s2.user_id IN ?
			GROUP BY s2.round_id, s2.user_id, s2.game_id, s2.begin_time
		) s1
		ORDER BY userid, score %s
		LIMIT 1 BY userid
	`
	var datas_MaxWinStats []map[string]any
	err = ck.Select(&datas_MaxWinStats, fmt.Sprintf(sql0, "desc"), userids, userids)
	if err != nil {
		return
	}
	for _, data := range datas_MaxWinStats {
		userid := data["userid"].(string)
		if item, ok := useridDatas[userid]; ok {
			gtype := utils.ToInt64(data["gtype"])
			score := utils.ToInt64(data["score"])
			gdatestr := fmt.Sprint(utils.ToInt64(data["gdate"]))
			sdate := gdatestr[0:4] + "年" + gdatestr[4:6] + "月" + gdatestr[6:] + "日"
			gameName, ok := GtypeNameMap[int(gtype)]
			if !ok {
				gameName = fmt.Sprint(gtype)
			}
			item.MaxWinStats = fmt.Sprintf("%.2f/%s/%s", float64(score)/100, gameName, sdate)
		}
	}
	// 输钱
	var datas_MaxLosesStats []map[string]any
	err = ck.Select(&datas_MaxLosesStats, fmt.Sprintf(sql0, "asc"), userids, userids)
	if err != nil {
		return
	}
	for _, data := range datas_MaxLosesStats {
		userid := data["userid"].(string)
		if item, ok := useridDatas[userid]; ok {
			gtype := utils.ToInt64(data["gtype"])
			score := utils.ToInt64(data["score"])
			gdatestr := fmt.Sprint(utils.ToInt64(data["gdate"]))
			sdate := gdatestr[0:4] + "年" + gdatestr[4:6] + "月" + gdatestr[6:] + "日"
			gameName, ok := GtypeNameMap[int(gtype)]
			if !ok {
				gameName = fmt.Sprint(gtype)
			}
			item.MaxLosesStats = fmt.Sprintf("%.2f/%s/%s", float64(score)/100, gameName, sdate)
		}
	}
	return
}

// 玩家分析：活跃天数
func (this *statisticsService) PlayerAnalysis_ActiveDays(userids []string, useridDatas map[string]*entity.PlayerAnalysisInfo) (err error) {
	var datas_ActiveDays []map[string]any
	err = ck.Select(&datas_ActiveDays, `
		SELECT userid, count(*) login_days FROM (
			SELECT userid, toYYYYMMDD(toDateTime(login_time)) login_day
			FROM game.col_log_login FINAL
			WHERE userid IN ?
			GROUP BY userid, login_day
		) s1
		GROUP BY userid
	`, userids, userids)
	if err != nil {
		return
	}
	for _, data := range datas_ActiveDays {
		userid := data["userid"].(string)
		if item, ok := useridDatas[userid]; ok {
			login_days := utils.ToInt64(data["login_days"])
			item.ActiveDay = login_days
			item.FActiveDay = fmt.Sprint(login_days)
		}
	}

	// 游戏时长
	var datas_GameOnline []map[string]any
	err = ck.Select(&datas_GameOnline, `
		SELECT userid, SUM(end_time - begin_time) online_sec
		FROM (
			SELECT id, userid, begin_time, end_time
			FROM game.col_detail t1 FINAL WHERE robot = 0 and win_type in (1,2,3) AND userid IN ?
				UNION ALL
			SELECT s2.round_id id, s2.user_id userid,
				s2.begin_time, IF(s3.end_time==0, s2.end_time, s3.end_time) end_time
			FROM (SELECT round_id, user_id, MIN(ctime) begin_time, MAX(ctime) end_time FROM game.col_nsq_log_external_bet FINAL WHERE amount != 0  GROUP BY round_id, user_id, game_id) s2
			LEFT JOIN (SELECT round_id, user_id, MAX(ctime) end_time FROM game.col_nsq_log_external_reward FINAL WHERE amount != 0 GROUP BY round_id, user_id) s3
				ON s2.round_id = s3.round_id AND s2.user_id = s3.user_id
			WHERE s2.user_id IN ?
			GROUP BY s2.round_id, s2.user_id, s2.begin_time, s2.end_time, s3.end_time
		) s1
		GROUP BY userid
	`, userids, userids)
	if err != nil {
		return
	}
	for _, data := range datas_GameOnline {
		userid := data["userid"].(string)
		if item, ok := useridDatas[userid]; ok && item.ActiveDay > 0 {
			online_sec := utils.ToFloat64(data["online_sec"])
			item.ActiveDayAvg = fmt.Sprintf("%.2f", online_sec/60/float64(item.ActiveDay))
		}
	}
	return
}

// 玩家分析：总汇
func (this *statisticsService) PlayerAnalysisSummary(where0 string, args0 []any) (summary *entity.PlayerAnalysisInfo, err error) {
	mark := "--"
	summary = &entity.PlayerAnalysisInfo{
		UserId: "汇总", UserType: mark, ChannelClass: mark, Channel1: mark, RegistTime: mark, LastLoginTime: mark, LoseDays: mark, ActiveTime: mark, FActiveDay: mark, RefUsers: mark,
		FirstPayDate: mark, FirstWithdrawDate: mark, FirstPayEfficience: "", PayBeforeCarry: mark, PayWithdrawTimes: mark,
		MaxPayDate: mark, MaxWithdrawDate: mark, MaxPayDateSuccessRate: mark, MaxWithdrawDateSuccessRate: mark, MaxPayTimesDateSuccessRate: mark, MaxWithdrawTimesDateSuccessRate: mark,
		Game1TimesStats: mark, Game1BetStats: mark, Game2TimesStats: mark, Game2BetStats: mark, Game3TimesStats: mark, Game3BetStats: mark, Game1PlayWinsAvg: mark, Game1PlayLosesAvg: mark, Game1BetWinsAvg: mark, Game1BetLosesAvg: mark, MaxWinStats: mark, MaxLosesStats: mark, ActiveDayAvg: mark,
		RewardRate: mark, Profit: mark,
		FirstPayAmount: mark, PayAmount: mark, PayCount: 0, PayAvg: mark, PaySuccessRate: mark, FirstWithdrawAmount: mark, WithdrawAmount: mark, WithdrawCount: 0, WithdrawAvg: mark, WithdrawSuccessRate: mark, PayWithdrawDiffDay: mark,
	}
	sql0 := `
		SELECT SUM(diamond) diamond_sum,
			SUM(pay_times_try) pay_times_try_sum,
			SUM(pay_times) pay_times_sum,
			SUM(pay_amount) pay_amount_sum,
			SUM(withdraw_times_try) withdraw_times_try_sum,
			SUM(withdraw_times) withdraw_times_sum,
			SUM(withdraw_amount) withdraw_amount_sum,
			SUM(profit) profit_sum
		FROM (
			SELECT t1.diamond diamond, t2.gtype_rounds, t2.gtype_bets,
				t3.pay_times_try, t3.pay_times, t3.pay_amount,
				t4.withdraw_times_try, t4.withdraw_times, t4.withdraw_amount,
				t1.diamond + t4.withdraw_amount - t3.pay_amount profit
			FROM game.col_user_finance t1 FINAL 
			JOIN game.col_user t0 FINAL ON t1.userid = t0.userid
			LEFT JOIN (
				SELECT userid, gtype, count(*) rounds, sum(bet_amount) bet_sum,
				first_value(gtype) OVER (PARTITION BY userid ORDER BY rounds DESC) AS gtype_rounds,
				first_value(gtype) OVER (PARTITION BY userid ORDER BY bet_sum DESC) AS gtype_bets
				FROM (
				SELECT id,userid,gtype,bet_amount
				FROM game.col_detail t1 FINAL WHERE robot = 0 and win_type in (1,2,3) 
					UNION ALL
				SELECT round_id id,user_id userid, game_id gtype,SUM(amount) bet_amount
				FROM game.col_nsq_log_external_bet t2 FINAL WHERE amount != 0 
				GROUP BY id, user_id, gtype
				) s1
				GROUP BY userid, gtype
				LIMIT 1 BY userid
			) t2 ON t1.userid = t2.userid
			LEFT JOIN (
				SELECT userid, count(*) pay_times_try, 
				SUM(CASE WHEN order_status = 4 THEN 1 ELSE 0 END) pay_times,
				SUM(CASE WHEN order_status = 4 THEN amount ELSE 0 END) pay_amount
				FROM game.col_trade_record FINAL
				GROUP BY userid
			) t3 ON t1.userid = t3.userid
			LEFT JOIN (
				SELECT userid, count(*) withdraw_times_try, 
				SUM(CASE WHEN order_status = 2 THEN 1 ELSE 0 END) withdraw_times,
				SUM(CASE WHEN order_status = 2 THEN amount ELSE 0 END) withdraw_amount
				FROM game.col_withdraw_record FINAL
				GROUP BY userid
			) t4 ON t1.userid = t4.userid
			WHERE 1 = 1 %s
		) s1
	`
	var datas0 []map[string]any
	err = ck.Select(&datas0, fmt.Sprintf(sql0, where0), args0...)
	if err != nil {
		return
	}
	if len(datas0) == 0 {
		return
	}
	data := datas0[0]
	// diamond := utils.ToInt64(data["diamond"])
	// pay_times_try := utils.ToInt64(data["pay_times_try"])
	diamond := utils.ToInt64(data["diamond_sum"])
	pay_times_try := utils.ToInt64(data["pay_times_try_sum"])
	pay_times := utils.ToInt64(data["pay_times_sum"])
	pay_amount := utils.ToInt64(data["pay_amount_sum"])
	withdraw_times_try := utils.ToInt64(data["withdraw_times_try_sum"])
	withdraw_times := utils.ToInt64(data["withdraw_times_sum"])
	withdraw_amount := utils.ToInt64(data["withdraw_amount_sum"])
	profit := utils.ToInt64(data["profit_sum"])

	summary.RewardRate = fmt.Sprintf("%.2f%%", utils.CaseElse(pay_amount == 0, 0, float64(withdraw_amount+diamond)/float64(pay_amount)*100))
	summary.Profit = fmt.Sprintf("%.2f", float64(profit)/100)
	summary.PayAmount = fmt.Sprintf("%.2f", float64(pay_amount)/100)
	summary.PayCount = pay_times
	summary.PayAvg = fmt.Sprintf("%.2f", utils.CaseElse(pay_times == 0, 0, float64(pay_amount)/float64(pay_times)/100))
	summary.PaySuccessRate = fmt.Sprintf("%.2f%%", utils.CaseElse(pay_times_try == 0, 0, float64(pay_times)/float64(pay_times_try)*100))
	summary.WithdrawAmount = fmt.Sprintf("%.2f", float64(withdraw_amount)/100)
	summary.WithdrawCount = withdraw_times
	summary.WithdrawAvg = fmt.Sprintf("%.2f", utils.CaseElse(withdraw_times == 0, 0, float64(withdraw_amount)/float64(withdraw_times)/100))
	summary.WithdrawSuccessRate = fmt.Sprintf("%.2f%%", utils.CaseElse(withdraw_times_try == 0, 0, float64(withdraw_times)/float64(withdraw_times_try)*100))
	summary.PayWithdrawTimes = fmt.Sprintf("%.2f", utils.CaseElse(withdraw_times == 0, 0, float64(pay_times)/float64(withdraw_times)))

	err = this.PlayerAnalysisSummary_Extra(summary, where0, args0)
	if err != nil {
		return
	}
	return
}

// 玩家分析：总汇额外统计
func (this *statisticsService) PlayerAnalysisSummary_Extra(summary *entity.PlayerAnalysisInfo, where0 string, args0 []any) (err error) {
	usersQ := `
		SELECT t1.userid
		FROM game.col_user_finance t1 FINAL 
		JOIN game.col_user t0 FINAL ON t1.userid = t0.userid
		LEFT JOIN (
			SELECT userid, count(*) rounds, sum(bet_amount) bet_sum,
				first_value(gtype) OVER (PARTITION BY userid ORDER BY rounds DESC) AS gtype_rounds,
				first_value(gtype) OVER (PARTITION BY userid ORDER BY bet_sum DESC) AS gtype_bets
			FROM (
				SELECT userid, gtype, bet_amount
				FROM game.col_detail t1 FINAL WHERE robot = 0 and win_type in (1,2,3) 
					UNION ALL
				SELECT user_id userid, game_id gtype, SUM(amount) bet_amount
				FROM game.col_nsq_log_external_bet t2 FINAL WHERE amount != 0 
				GROUP BY id, user_id, gtype
			) s1
			GROUP BY userid, gtype
			LIMIT 1 BY userid
		) t2 ON t1.userid = t2.userid
		LEFT JOIN (
			SELECT userid, SUM(amount) pay_amount FROM game.col_trade_record FINAL WHERE order_status = 4 GROUP BY userid
		) t3 ON t1.userid = t3.userid
		LEFT JOIN (
			SELECT userid, SUM(amount) withdraw_amount FROM game.col_withdraw_record FINAL WHERE order_status = 2 GROUP BY userid
		) t4 ON t1.userid = t4.userid
		WHERE 1 = 1 %s
	`
	usersQ = fmt.Sprintf(usersQ, where0)

	var wg = &sync.WaitGroup{}
	// 首充效率	首提距首充间隔	每次充值前携带金额均值	充值次数/提现次数
	wg.Add(1)
	go func() {
		defer wg.Done()
		var data_firstPay = make(map[string]any)
		err = ck.Select(&data_firstPay, fmt.Sprintf(`SELECT 
			SUM(case when t5.amount = 0 then 0 else DateDiff('minute', t1.ctime, t5.ctime) end) / SUM(case when t5.amount = 0 then 0 else 1 end) first_pay_time_avg,
			SUM(case when t6.amount = 0 then 0 else DateDiff('minute', t5.ctime, t6.ctime) end) / SUM(case when t6.amount = 0 then 0 else 1 end) pay_withdraw_time_avg
		FROM game.col_user_finance t1 FINAL 
		LEFT JOIN (
			SELECT userid, ctime, amount FROM game.col_trade_record FINAL 
			WHERE order_status = 4 ORDER BY userid, ctime LIMIT 1 BY userid
		) t5 ON t1.userid = t5.userid
		LEFT JOIN (
			SELECT userid, ctime, amount FROM game.col_withdraw_record FINAL 
			WHERE order_status = 2 ORDER BY userid, ctime LIMIT 1 BY userid
		) t6 ON t1.userid = t6.userid
		WHERE t1.userid IN (%s)`, usersQ), args0...)
		if err != nil {
			return
		}
		// first_withdraw_time_avg
		summary.FirstPayEfficience = fmt.Sprintf("%.2f", utils.ToFloat64(data_firstPay["first_pay_time_avg"]))
		summary.PayWithdrawDiffDay = fmt.Sprintf("%.2f", utils.ToFloat64(data_firstPay["pay_withdraw_time_avg"]))
	}()

	// 充值前携带均值
	wg.Add(1)
	go func() {
		defer wg.Done()
		var data_Carry = make(map[string]any)
		err = ck.Select(&data_Carry, fmt.Sprintf(`SELECT AVG(old_diamond) old_diamond_avg FROM game.col_log_water FINAL
				WHERE ltype IN (4, 74, 84, 86, 119, 80, 57, 58, 59, 129) AND userid IN (%s)`, usersQ), args0...)
		if err != nil {
			return
		}
		summary.PayBeforeCarry = fmt.Sprintf("%.2f", utils.ToFloat64(data_Carry["old_diamond_avg"]))
	}()

	// 单笔最大充值金额/日期
	wg.Add(1)
	go func() {
		defer wg.Done()
		var data_MaxPayDate = make(map[string]any)
		err = ck.Select(&data_MaxPayDate, fmt.Sprintf(`
			SELECT amount, arraySort(groupArray(DISTINCT toYYYYMMDD(ctime))) dates
				FROM game.col_trade_record FINAL 
			WHERE order_status = 4 AND userid IN (%s)
			GROUP BY amount
			ORDER BY amount desc
			LIMIT 1 
		`, usersQ), args0...)
		if err != nil {
			return
		}
		amount := utils.ToInt64(data_MaxPayDate["amount"])
		if amount > 0 {
			dates := data_MaxPayDate["dates"].([]uint32)
			var fmtDates []string
			for _, date := range dates {
				datestr := fmt.Sprint(date)
				sdate := datestr[0:4] + "年" + datestr[4:6] + "月" + datestr[6:] + "日"
				fmtDates = append(fmtDates, sdate)
			}
			summary.MaxPayDate = fmt.Sprintf("%.0f/", float64(amount)/100.0) + strings.Join(fmtDates, "、")
		}
	}()

	// 单笔最大提现金额/日期
	wg.Add(1)
	go func() {
		defer wg.Done()
		var data_MaxWithdrawDate = make(map[string]any)
		err = ck.Select(&data_MaxWithdrawDate, fmt.Sprintf(`
			SELECT amount, arraySort(groupArray(DISTINCT toYYYYMMDD(ctime))) dates
				FROM game.col_trade_record FINAL 
			WHERE order_status = 4 AND userid IN (%s)
			GROUP BY amount
			ORDER BY amount desc
			LIMIT 1 
		`, usersQ), args0...)
		if err != nil {
			return
		}
		amount := utils.ToInt64(data_MaxWithdrawDate["amount"])
		if amount > 0 {
			dates := data_MaxWithdrawDate["dates"].([]uint32)
			var fmtDates []string
			for _, date := range dates {
				datestr := fmt.Sprint(date)
				sdate := datestr[0:4] + "年" + datestr[4:6] + "月" + datestr[6:] + "日"
				fmtDates = append(fmtDates, sdate)
			}
			summary.MaxWithdrawDate = fmt.Sprintf("%.0f/", float64(amount)/100.0) + strings.Join(fmtDates, "、")
		}
	}()

	// 单日累计最大充值金额/日期/成功率
	wg.Add(1)
	go func() {
		defer wg.Done()
		var data_MaxPayDateSuccessRate = make(map[string]any)
		err = ck.Select(&data_MaxPayDateSuccessRate, fmt.Sprintf(`
			SELECT amount_day, groupArray(pay_day) pay_dates, groupArray(pay_success_rate_day) pay_success_rates 
			FROM (
				SELECT toYYYYMMDD(ctime) pay_day, 
					SUM(CASE WHEN order_status = 4 THEN amount ELSE 0 END) amount_day,  
					SUM(CASE WHEN order_status = 4 THEN 1 ELSE 0 END) / count(*) pay_success_rate_day
				FROM game.col_trade_record FINAL 
				WHERE userid IN (%s)
				GROUP BY pay_day
				HAVING amount_day > 0
			) t1
			GROUP BY amount_day
			ORDER BY amount_day desc
			LIMIT 1
		`, usersQ), args0...)
		if err != nil {
			return
		}
		amount_day := utils.ToInt64(data_MaxPayDateSuccessRate["amount_day"])
		if amount_day > 0 {
			pay_dates := data_MaxPayDateSuccessRate["pay_dates"].([]uint32)
			pay_success_rates := data_MaxPayDateSuccessRate["pay_success_rates"].([]float64)
			var fmtDates, fmtRates []string
			for i, pay_date := range pay_dates {
				pay_success_rate := pay_success_rates[i]
				datestr := fmt.Sprint(pay_date)
				sdate := datestr[0:4] + "年" + datestr[4:6] + "月" + datestr[6:] + "日"
				fmtDates = append(fmtDates, sdate)
				fmtRates = append(fmtRates, fmt.Sprintf("%.2f%%", pay_success_rate*100))
			}
			summary.MaxPayDateSuccessRate = fmt.Sprintf("%.0f/", float64(amount_day)/100.0) + strings.Join(fmtDates, "、") + "/" + strings.Join(fmtRates, "、")
		}
	}()

	// 单日累计最大提现金额/日期/成功率
	wg.Add(1)
	go func() {
		defer wg.Done()
		var data_MaxWithdrawDateSuccessRate = make(map[string]any)
		err = ck.Select(&data_MaxWithdrawDateSuccessRate, fmt.Sprintf(`
			SELECT amount_day, groupArray(pay_day) pay_dates, groupArray(pay_success_rate_day) pay_success_rates 
			FROM (
				SELECT toYYYYMMDD(ctime) pay_day, 
					SUM(CASE WHEN order_status = 2 THEN amount ELSE 0 END) amount_day,  
					SUM(CASE WHEN order_status = 2 THEN 1 ELSE 0 END) / count(*) pay_success_rate_day
				FROM game.col_withdraw_record FINAL 
				WHERE userid IN (%s)
				GROUP BY pay_day
				HAVING amount_day > 0
			) t1
			GROUP BY amount_day
			ORDER BY amount_day desc
			LIMIT 1
		`, usersQ), args0...)
		if err != nil {
			return
		}
		amount_day := utils.ToInt64(data_MaxWithdrawDateSuccessRate["amount_day"])
		if amount_day > 0 {
			pay_dates := data_MaxWithdrawDateSuccessRate["pay_dates"].([]uint32)
			pay_success_rates := data_MaxWithdrawDateSuccessRate["pay_success_rates"].([]float64)
			var fmtDates, fmtRates []string
			for i, pay_date := range pay_dates {
				pay_success_rate := pay_success_rates[i]
				datestr := fmt.Sprint(pay_date)
				sdate := datestr[0:4] + "年" + datestr[4:6] + "月" + datestr[6:] + "日"
				fmtDates = append(fmtDates, sdate)
				fmtRates = append(fmtRates, fmt.Sprintf("%.2f%%", pay_success_rate*100))
			}
			summary.MaxWithdrawDateSuccessRate = fmt.Sprintf("%.0f/", float64(amount_day)/100.0) + strings.Join(fmtDates, "、") + "/" + strings.Join(fmtRates, "、")
		}
	}()

	// 单日最高充值次数/日期/成功率
	wg.Add(1)
	go func() {
		defer wg.Done()
		var data_MaxPayTimesDateSuccessRate = make(map[string]any)
		err = ck.Select(&data_MaxPayTimesDateSuccessRate, fmt.Sprintf(`
			SELECT pay_times_day, groupArray(pay_day) pay_dates, groupArray(pay_success_rate_day) pay_success_rates
			FROM (
				SELECT toYYYYMMDD(ctime) pay_day, 
					SUM(CASE WHEN order_status = 4 THEN 1 ELSE 0 END) pay_times_day,
					pay_times_day / count(*) pay_success_rate_day
				FROM game.col_trade_record FINAL 
				WHERE userid IN (%s)
				GROUP BY pay_day
				HAVING pay_times_day > 0
			) t1
			GROUP BY pay_times_day
			ORDER BY pay_times_day desc
			LIMIT 1
		`, usersQ), args0...)
		if err != nil {
			return
		}
		pay_times_day := utils.ToInt64(data_MaxPayTimesDateSuccessRate["pay_times_day"])
		if pay_times_day != 0 {
			pay_dates := data_MaxPayTimesDateSuccessRate["pay_dates"].([]uint32)
			pay_success_rates := data_MaxPayTimesDateSuccessRate["pay_success_rates"].([]float64)
			var fmtDates, fmtRates []string
			for i, pay_date := range pay_dates {
				pay_success_rate := pay_success_rates[i]
				datestr := fmt.Sprint(pay_date)
				sdate := datestr[0:4] + "年" + datestr[4:6] + "月" + datestr[6:] + "日"
				fmtDates = append(fmtDates, sdate)
				fmtRates = append(fmtRates, fmt.Sprintf("%.2f%%", pay_success_rate*100))
			}
			summary.MaxPayTimesDateSuccessRate = fmt.Sprintf("%d/", pay_times_day) + strings.Join(fmtDates, "、") + "/" + strings.Join(fmtRates, "、")
		}
	}()

	// 单日最高提现次数/日期/成功率
	wg.Add(1)
	go func() {
		defer wg.Done()
		var data_MaxWithdrawTimesDateSuccessRate = make(map[string]any)
		err = ck.Select(&data_MaxWithdrawTimesDateSuccessRate, fmt.Sprintf(`
			SELECT pay_times_day, groupArray(pay_day) pay_dates, groupArray(pay_success_rate_day) pay_success_rates
			FROM (
				SELECT toYYYYMMDD(ctime) pay_day, 
					SUM(CASE WHEN order_status = 2 THEN 1 ELSE 0 END) pay_times_day,
					pay_times_day / count(*) pay_success_rate_day
				FROM game.col_withdraw_record FINAL 
				WHERE userid IN (%s)
				GROUP BY pay_day
				HAVING pay_times_day > 0
			) t1
			GROUP BY pay_times_day
			ORDER BY pay_times_day desc
			LIMIT 1
		`, usersQ), args0...)
		if err != nil {
			return
		}
		pay_times_day := utils.ToInt64(data_MaxWithdrawTimesDateSuccessRate["pay_times_day"])
		if pay_times_day > 0 {
			pay_dates := data_MaxWithdrawTimesDateSuccessRate["pay_dates"].([]uint32)
			pay_success_rates := data_MaxWithdrawTimesDateSuccessRate["pay_success_rates"].([]float64)
			fmtDates, fmtRates := []string{}, []string{}
			for i, pay_date := range pay_dates {
				pay_success_rate := pay_success_rates[i]
				datestr := fmt.Sprint(pay_date)
				sdate := datestr[0:4] + "年" + datestr[4:6] + "月" + datestr[6:] + "日"
				fmtDates = append(fmtDates, sdate)
				fmtRates = append(fmtRates, fmt.Sprintf("%.2f%%", pay_success_rate*100))
			}
			summary.MaxWithdrawTimesDateSuccessRate = fmt.Sprintf("%d/", pay_times_day) + strings.Join(fmtDates, "、") + "/" + strings.Join(fmtRates, "、")
		}
	}()

	// 玩最多的游戏/局数/局均码量/返奖率
	wg.Add(1)
	go func() {
		defer wg.Done()
		var data_GameStats []map[string]any
		err = ck.Select(&data_GameStats, fmt.Sprintf(`
			SELECT * FROM (
				SELECT gtype, count(*) rounds, SUM(bet_amount) bet_sum,  AVG(bet_amount) bet_avg, 
					SUM(CASE WHEN win_type = 1 THEN score ELSE 0 END) win_amount,
					SUM(CASE WHEN win_type = 2 THEN score ELSE 0 END) lose_amount,
					SUM(CASE WHEN win_type = 1 THEN 1 ELSE 0 END) win_rounds,
					SUM(CASE WHEN win_type = 2 THEN 1 ELSE 0 END) lose_rounds,
					row_number() OVER (ORDER BY rounds desc) AS round_no,
					row_number() OVER (ORDER BY bet_sum desc) AS bet_no
				FROM (
						SELECT id, gtype, bet_amount, win_type, score
						FROM game.col_detail t1 FINAL WHERE robot = 0 and win_type in (1,2,3) AND userid IN (%s)
							UNION ALL
						SELECT s2.round_id id, s2.game_id gtype, SUM(s2.amount_sum) bet_amount,
							(CASE WHEN score > 0 THEN 1 WHEN score == 0 THEN 3 ELSE 2 END) win_type, 
							SUM(s3.amount_sum) - SUM(s2.amount_sum) score
						FROM (SELECT round_id, user_id, game_id, SUM(amount) amount_sum FROM game.col_nsq_log_external_bet FINAL WHERE amount != 0  GROUP BY round_id, user_id, game_id) s2
						LEFT JOIN (SELECT round_id, user_id, SUM(amount) amount_sum FROM game.col_nsq_log_external_reward FINAL WHERE amount != 0 GROUP BY round_id, user_id) s3
							ON s2.round_id = s3.round_id AND s2.user_id = s3.user_id
						WHERE s2.user_id IN (%s)
						GROUP BY s2.round_id, s2.game_id
				) s1
				GROUP BY gtype
			) tt
			WHERE tt.round_no <= 3 OR tt.bet_no <= 3
			ORDER BY rounds DESC
		`, usersQ, usersQ), append(args0, args0...)...)
		if err != nil {
			return
		}
		for _, data := range data_GameStats {
			round_no := utils.ToInt64(data["round_no"])
			bet_no := utils.ToInt64(data["bet_no"])
			gtype := utils.ToInt64(data["gtype"])
			rounds := utils.ToInt64(data["rounds"])
			// bet_sum := utils.ToInt64(data["bet_sum"])
			bet_avg := utils.ToFloat64(data["bet_avg"])
			win_amount := utils.ToInt64(data["win_amount"])
			lose_amount := utils.ToInt64(data["lose_amount"])
			win_rounds := utils.ToInt64(data["win_rounds"])
			lose_rounds := utils.ToInt64(data["lose_rounds"])

			gameName, ok := GtypeNameMap[int(gtype)]
			if !ok {
				gameName = fmt.Sprint(gtype)
			}
			// 游戏/局数/局均码量/返奖率
			stats := fmt.Sprintf("%s/%d/%.2f/%.2f%%", gameName, rounds, bet_avg/100, utils.CaseElse(lose_amount == 0, 0, float64(win_amount)/float64(-lose_amount)*100))
			switch round_no {
			case 1:
				if win_rounds > 0 {
					summary.Game1PlayWinsAvg = fmt.Sprintf("%.2f", float64(win_amount)/float64(win_rounds)/100)
				}
				if lose_rounds > 0 {
					summary.Game1PlayLosesAvg = fmt.Sprintf("%.2f", float64(lose_amount)/float64(lose_rounds)/100)
				}
				summary.Game1TimesStats = stats
			case 2:
				summary.Game2TimesStats = stats
			case 3:
				summary.Game3TimesStats = stats
			}
			switch bet_no {
			case 1:
				if win_rounds > 0 {
					summary.Game1BetWinsAvg = fmt.Sprintf("%.2f", float64(win_amount)/float64(win_rounds)/100)
				}
				if lose_rounds > 0 {
					summary.Game1BetLosesAvg = fmt.Sprintf("%.2f", float64(lose_amount)/float64(lose_rounds)/100)
				}
				summary.Game1BetStats = stats
			case 2:
				summary.Game2BetStats = stats
			case 3:
				summary.Game3BetStats = stats
			}
		}
	}()

	// 最大单局赢钱金额/所在游戏/日期
	wg.Add(1)
	go func() {
		defer wg.Done()
		var data_MaxWinStats = make(map[string]any)
		sql_MaxWinStats := `
			SELECT id, score, gtype, toYYYYMMDD(toDateTime(begin_time)) gdate
			FROM (
				SELECT id, gtype, win_type, score, begin_time
				FROM game.col_detail t1 FINAL WHERE robot = 0 and win_type in (1,2,3) AND userid IN (%s)
					UNION ALL
				SELECT s2.round_id id, s2.game_id gtype,
					(CASE WHEN score > 0 THEN 1 WHEN score == 0 THEN 3 ELSE 2 END) win_type, 
					SUM(s3.amount_sum) - SUM(s2.amount_sum) score, s2.begin_time
				FROM (SELECT round_id, user_id, game_id, SUM(amount) amount_sum, MIN(ctime) begin_time FROM game.col_nsq_log_external_bet FINAL WHERE amount != 0  GROUP BY round_id, user_id, game_id) s2
				LEFT JOIN (SELECT round_id, user_id, SUM(amount) amount_sum FROM game.col_nsq_log_external_reward FINAL WHERE amount != 0 GROUP BY round_id, user_id) s3
					ON s2.round_id = s3.round_id AND s2.user_id = s3.user_id
				WHERE s2.user_id IN (%s)
				GROUP BY s2.round_id, s2.game_id, s2.begin_time
			) s1
			ORDER BY score %s
			LIMIT 1
		`
		err = ck.Select(&data_MaxWinStats, fmt.Sprintf(sql_MaxWinStats, usersQ, usersQ, "desc"), append(args0, args0...)...)
		if err != nil {
			return
		}
		gtype := utils.ToInt64(data_MaxWinStats["gtype"])
		if gtype != 0 {
			score := utils.ToInt64(data_MaxWinStats["score"])
			gdatestr := fmt.Sprint(utils.ToInt64(data_MaxWinStats["gdate"]))
			sdate := gdatestr[0:4] + "年" + gdatestr[4:6] + "月" + gdatestr[6:] + "日"
			gameName, ok := GtypeNameMap[int(gtype)]
			if !ok {
				gameName = fmt.Sprint(gtype)
			}
			summary.MaxWinStats = fmt.Sprintf("%.2f/%s/%s", float64(score)/100, gameName, sdate)
		}

		// 单局输钱
		err = ck.Select(&data_MaxWinStats, fmt.Sprintf(sql_MaxWinStats, usersQ, usersQ, "asc"), append(args0, args0...)...)
		if err != nil {
			return
		}
		gtype = utils.ToInt64(data_MaxWinStats["gtype"])
		if gtype != 0 {
			score := utils.ToInt64(data_MaxWinStats["score"])
			gdatestr := fmt.Sprint(utils.ToInt64(data_MaxWinStats["gdate"]))
			sdate := gdatestr[0:4] + "年" + gdatestr[4:6] + "月" + gdatestr[6:] + "日"
			gameName, ok := GtypeNameMap[int(gtype)]
			if !ok {
				gameName = fmt.Sprint(gtype)
			}
			summary.MaxLosesStats = fmt.Sprintf("%.2f/%s/%s", float64(score)/100, gameName, sdate)
		}
	}()

	// 日均在线时长
	var data_LoginDays = make(map[string]any)
	err = ck.Select(&data_LoginDays, fmt.Sprintf(`
		SELECT count(*) login_days FROM (
			SELECT userid, toYYYYMMDD(toDateTime(login_time)) login_day
			FROM game.col_log_login FINAL
			WHERE userid IN (%s)
			GROUP BY userid, login_day
		) s1
	`, usersQ), args0...)
	if err != nil {
		return
	}
	login_days := utils.ToInt64(data_LoginDays["login_days"])
	var data_Onlines = make(map[string]any)
	err = ck.Select(&data_Onlines, fmt.Sprintf(`
		SELECT SUM(end_time - begin_time) online_sec
		FROM (
			SELECT id, begin_time, end_time
			FROM game.col_detail t1 FINAL WHERE robot = 0 and win_type in (1,2,3) AND userid IN (%s)
				UNION ALL
			SELECT s2.round_id id,
				s2.begin_time, IF(s3.end_time==0, s2.end_time, s3.end_time) end_time
			FROM (SELECT round_id, user_id, MIN(ctime) begin_time, MAX(ctime) end_time FROM game.col_nsq_log_external_bet FINAL WHERE amount != 0  GROUP BY round_id, user_id, game_id) s2
			LEFT JOIN (SELECT round_id, user_id, MAX(ctime) end_time FROM game.col_nsq_log_external_reward FINAL WHERE amount != 0 GROUP BY round_id, user_id) s3
				ON s2.round_id = s3.round_id AND s2.user_id = s3.user_id
			WHERE s2.user_id IN (%s)
			GROUP BY s2.round_id, s2.begin_time, s2.end_time, s3.end_time
		) s1
	`, usersQ, usersQ), append(args0, args0...)...)
	if err != nil {
		return
	}
	online_sec := utils.ToInt64(data_Onlines["online_sec"])
	summary.ActiveDayAvg = utils.CaseElse(login_days == 0, "", fmt.Sprintf("%.2f", float64(online_sec)/float64(login_days)/60))

	wg.Wait()
	return
}

// 玩家分析：玩家活跃时段分布图
func (this *statisticsService) PlayerAnalysis_ActiveChart(
	begin, end *time.Time,
	userid string, utypeIds []int, moneyRange []int64, withdrawRange []int64, profitRange []int64,
	maxRoundsGtype, maxBetGtype int32,
	rangeStartTime, rangeEndTime *time.Time, density int32,
) (labels []string, datasets []int64, err error) {
	where0, args0 := this.PlayerAnalysisCondition(begin, end, userid, utypeIds, moneyRange, withdrawRange, profitRange, maxRoundsGtype, maxBetGtype)
	usersQ := `
		SELECT t1.userid
		FROM game.col_user_finance t1 FINAL 
		JOIN game.col_user t0 FINAL ON t1.userid = t0.userid
		LEFT JOIN (
			SELECT userid, count(*) rounds, sum(bet_amount) bet_sum,
				first_value(gtype) OVER (PARTITION BY userid ORDER BY rounds DESC) AS gtype_rounds,
				first_value(gtype) OVER (PARTITION BY userid ORDER BY bet_sum DESC) AS gtype_bets
			FROM (
				SELECT userid, gtype, bet_amount
				FROM game.col_detail t1 FINAL WHERE robot = 0 and win_type in (1,2,3) 
					UNION ALL
				SELECT user_id userid, game_id gtype, SUM(amount) bet_amount
				FROM game.col_nsq_log_external_bet t2 FINAL WHERE amount != 0 
				GROUP BY id, user_id, gtype
			) s1
			GROUP BY userid, gtype
			LIMIT 1 BY userid
		) t2 ON t1.userid = t2.userid
		LEFT JOIN (
			SELECT userid, SUM(amount) pay_amount FROM game.col_trade_record FINAL WHERE order_status = 4 GROUP BY userid
		) t3 ON t1.userid = t3.userid
		LEFT JOIN (
			SELECT userid, SUM(amount) withdraw_amount FROM game.col_withdraw_record FINAL WHERE order_status = 2 GROUP BY userid
		) t4 ON t1.userid = t4.userid
		WHERE 1 = 1 %s
	`
	usersQ = fmt.Sprintf(usersQ, where0)

	var where2 string
	var args2 []any
	if rangeStartTime != nil {
		where2 += " and ctime >= ?"
		args2 = append(args2, *rangeStartTime)
	}
	if rangeEndTime != nil {
		where2 += " and ctime <= ?"
		args2 = append(args2, *rangeEndTime)
	}
	var args = []any{density}
	args = append(args, args0...)
	args = append(args, args2...)
	args = append(args, args0...)
	args = append(args, args2...)

	var datas []map[string]any
	err = ck.Select(&datas, fmt.Sprintf(`
		SELECT in_minute, count(*) hots FROM (
			SELECT userid, intDiv((toHour(ctime) * 60 + toMinute(ctime)), ?) in_minute, toYYYYMMDD(ctime) in_day 
			FROM (
				SELECT id, userid, toDateTime(begin_time) ctime
				FROM game.col_detail t1 FINAL WHERE robot = 0 and win_type in (1,2,3) AND userid IN (%s) %s
					UNION ALL
				SELECT s2.round_id id, s2.user_id userid, toDateTime(s2.begin_time) ctime
				FROM (SELECT round_id, user_id, MIN(ctime) begin_time, MAX(ctime) end_time FROM game.col_nsq_log_external_bet FINAL WHERE amount != 0 GROUP BY round_id, user_id, game_id) s2
				WHERE s2.user_id IN (%s) %s
				GROUP BY s2.round_id, s2.user_id, s2.begin_time, s2.end_time
			) s1
			GROUP BY userid, in_minute, in_day
			ORDER BY userid, in_minute, in_day
		) s2
		GROUP BY in_minute
		ORDER BY in_minute
	`, usersQ, where2, usersQ, where2), args...)
	if err != nil {
		return
	}

	var inMinuteHots = make(map[int64]int64, len(datas))
	for _, data := range datas {
		in_minute := utils.ToInt64(data["in_minute"])
		hots := utils.ToInt64(data["hots"])
		inMinuteHots[in_minute] = hots
	}

	var i int64
	for ; i < 24*60/int64(density); i++ {
		datasets = append(datasets, inMinuteHots[i])

		minutes := i * int64(density)
		labels = append(labels, fmt.Sprintf("%02d:%02d", minutes/60, minutes%60))
	}
	return
}

// 玩家分析：在线时长走势图
func (this *statisticsService) PlayerAnalysis_OnlineChart(
	begin, end *time.Time,
	userid string, utypeIds []int, moneyRange []int64, withdrawRange []int64, profitRange []int64,
	maxRoundsGtype, maxBetGtype int32,
	rangeStartTime, rangeEndTime *time.Time,
) (labels []string, datasets []string, err error) {
	where0, args0 := this.PlayerAnalysisCondition(begin, end, userid, utypeIds, moneyRange, withdrawRange, profitRange, maxRoundsGtype, maxBetGtype)
	usersQ := `
		SELECT t1.userid
		FROM game.col_user_finance t1 FINAL 
		JOIN game.col_user t0 FINAL ON t1.userid = t0.userid
		LEFT JOIN (
			SELECT userid, count(*) rounds, sum(bet_amount) bet_sum,
				first_value(gtype) OVER (PARTITION BY userid ORDER BY rounds DESC) AS gtype_rounds,
				first_value(gtype) OVER (PARTITION BY userid ORDER BY bet_sum DESC) AS gtype_bets
			FROM (
				SELECT userid, gtype, bet_amount
				FROM game.col_detail t1 FINAL WHERE robot = 0 and win_type in (1,2,3) 
					UNION ALL
				SELECT user_id userid, game_id gtype, SUM(amount) bet_amount
				FROM game.col_nsq_log_external_bet t2 FINAL WHERE amount != 0 
				GROUP BY id, user_id, gtype
			) s1
			GROUP BY userid, gtype
			LIMIT 1 BY userid
		) t2 ON t1.userid = t2.userid
		LEFT JOIN (
			SELECT userid, SUM(amount) pay_amount FROM game.col_trade_record FINAL WHERE order_status = 4 GROUP BY userid
		) t3 ON t1.userid = t3.userid
		LEFT JOIN (
			SELECT userid, SUM(amount) withdraw_amount FROM game.col_withdraw_record FINAL WHERE order_status = 2 GROUP BY userid
		) t4 ON t1.userid = t4.userid
		WHERE 1 = 1 %s
	`
	usersQ = fmt.Sprintf(usersQ, where0)

	var where2 string
	var args2 []any
	if rangeStartTime != nil {
		where2 += " and begin_time >= ?"
		args2 = append(args2, rangeStartTime.Unix())
	}
	if rangeEndTime != nil {
		where2 += " and begin_time <= ?"
		args2 = append(args2, rangeEndTime.Unix())
	}
	var args []any
	args = append(args, args0...)
	args = append(args, args2...)
	args = append(args, args0...)
	args = append(args, args2...)

	var datas []map[string]any
	err = ck.Select(&datas, fmt.Sprintf(`
		SELECT days, SUM(end_time - begin_time) / 60 online_minutes
		FROM (
			SELECT id, userid, begin_time, end_time, toYYYYMMDD(toDateTime(begin_time)) days
			FROM game.col_detail t1 FINAL WHERE robot = 0 and win_type in (1,2,3) AND userid IN (%s) %s
				UNION ALL
			SELECT s2.round_id id, s2.user_id userid,
				s2.begin_time, IF(s3.end_time==0, s2.end_time, s3.end_time) end_time,
				toYYYYMMDD(toDateTime(begin_time)) days
			FROM (SELECT round_id, user_id, MIN(ctime) begin_time, MAX(ctime) end_time FROM game.col_nsq_log_external_bet FINAL WHERE amount != 0  GROUP BY round_id, user_id, game_id) s2
			LEFT JOIN (SELECT round_id, user_id, MAX(ctime) end_time FROM game.col_nsq_log_external_reward FINAL WHERE amount != 0 GROUP BY round_id, user_id) s3
				ON s2.round_id = s3.round_id AND s2.user_id = s3.user_id
			WHERE s2.user_id IN (%s) %s
			GROUP BY s2.round_id, s2.user_id, s2.begin_time, s2.end_time, s3.end_time
		) s1
		GROUP BY days
		ORDER BY days
	`, usersQ, where2, usersQ, where2), args...)
	if err != nil {
		return
	}

	var daysOnlines = make(map[string]float64, len(datas))
	for _, data := range datas {
		online_minutes := utils.ToFloat64(data["online_minutes"])
		days := fmt.Sprint(utils.ToInt64(data["days"]))
		sdate := days[0:4] + "-" + days[4:6] + "-" + days[6:]
		daysOnlines[sdate] = online_minutes
	}

	stime, etime := *rangeStartTime, *rangeEndTime

	for !stime.After(etime) {
		date := stime.Format(utils.FORMAT_DATE)
		datasets = append(datasets, fmt.Sprintf("%.2f", daysOnlines[date]))
		labels = append(labels, date)

		stime = stime.AddDate(0, 0, 1)
	}

	return
}

// 玩家分析：充提金额和次数走势图
func (this *statisticsService) PlayerAnalysis_TradeChart(
	begin, end *time.Time,
	userid string, utypeIds []int, moneyRange []int64, withdrawRange []int64, profitRange []int64,
	maxRoundsGtype, maxBetGtype int32,
	rangeStartTime, rangeEndTime *time.Time,
) (labels []string, payCounts, withdrawCounts []int64, payAmounts, withdrawAmounts []string, err error) {
	where0, args0 := this.PlayerAnalysisCondition(begin, end, userid, utypeIds, moneyRange, withdrawRange, profitRange, maxRoundsGtype, maxBetGtype)
	usersQ := `
		SELECT t1.userid
		FROM game.col_user_finance t1 FINAL 
		JOIN game.col_user t0 FINAL ON t1.userid = t0.userid
		LEFT JOIN (
			SELECT userid, count(*) rounds, sum(bet_amount) bet_sum,
				first_value(gtype) OVER (PARTITION BY userid ORDER BY rounds DESC) AS gtype_rounds,
				first_value(gtype) OVER (PARTITION BY userid ORDER BY bet_sum DESC) AS gtype_bets
			FROM (
				SELECT userid, gtype, bet_amount
				FROM game.col_detail t1 FINAL WHERE robot = 0 and win_type in (1,2,3) 
					UNION ALL
				SELECT user_id userid, game_id gtype, SUM(amount) bet_amount
				FROM game.col_nsq_log_external_bet t2 FINAL WHERE amount != 0 
				GROUP BY id, user_id, gtype
			) s1
			GROUP BY userid, gtype
			LIMIT 1 BY userid
		) t2 ON t1.userid = t2.userid
		LEFT JOIN (
			SELECT userid, SUM(amount) pay_amount FROM game.col_trade_record FINAL WHERE order_status = 4 GROUP BY userid
		) t3 ON t1.userid = t3.userid
		LEFT JOIN (
			SELECT userid, SUM(amount) withdraw_amount FROM game.col_withdraw_record FINAL WHERE order_status = 2 GROUP BY userid
		) t4 ON t1.userid = t4.userid
		WHERE 1 = 1 %s
	`
	usersQ = fmt.Sprintf(usersQ, where0)

	var where2 string
	var args2 []any
	if rangeStartTime != nil {
		where2 += " and ctime >= ?"
		args2 = append(args2, *rangeStartTime)
	}
	if rangeEndTime != nil {
		where2 += " and ctime <= ?"
		args2 = append(args2, *rangeEndTime)
	}
	var args []any
	args = append(args, args0...)
	args = append(args, args2...)

	var datas_pay, datas_withdraw []map[string]any
	err = ck.Select(&datas_pay, fmt.Sprintf(`
		SELECT toYYYYMMDD(ctime) days, count(*) count, SUM(amount) amount 
		FROM game.col_trade_record FINAL 
		WHERE order_status = 4 AND userid IN (%s) %s
		GROUP BY days
		ORDER BY days
	`, usersQ, where2), args...)
	if err != nil {
		return
	}
	err = ck.Select(&datas_withdraw, fmt.Sprintf(`
		SELECT toYYYYMMDD(ctime) days, count(*) count, SUM(amount) amount 
		FROM game.col_withdraw_record FINAL 
		WHERE order_status = 2 AND userid IN (%s) %s
		GROUP BY days
		ORDER BY days
	`, usersQ, where2), args...)
	if err != nil {
		return
	}

	var daysPay = make(map[string]float64, len(datas_pay))
	var daysPayCount = make(map[string]int64, len(datas_pay))
	var daysWithdraw = make(map[string]float64, len(datas_withdraw))
	var daysWithdrawCount = make(map[string]int64, len(datas_withdraw))
	for _, data := range datas_pay {
		amount := utils.ToFloat64(data["amount"])
		count := utils.ToInt64(data["count"])
		days := fmt.Sprint(utils.ToInt64(data["days"]))
		sdate := days[0:4] + "-" + days[4:6] + "-" + days[6:]
		daysPay[sdate] = amount
		daysPayCount[sdate] = count
	}
	for _, data := range datas_withdraw {
		amount := utils.ToFloat64(data["amount"])
		count := utils.ToInt64(data["count"])
		days := fmt.Sprint(utils.ToInt64(data["days"]))
		sdate := days[0:4] + "-" + days[4:6] + "-" + days[6:]
		daysWithdraw[sdate] = amount
		daysWithdrawCount[sdate] = count
	}

	stime, etime := *rangeStartTime, *rangeEndTime
	for !stime.After(etime) {
		date := stime.Format(utils.FORMAT_DATE)
		payCounts = append(payCounts, daysPayCount[date])
		payAmounts = append(payAmounts, fmt.Sprintf("%.2f", daysPay[date]/100))
		withdrawCounts = append(withdrawCounts, daysWithdrawCount[date])
		withdrawAmounts = append(withdrawAmounts, fmt.Sprintf("%.2f", daysWithdraw[date]/100))

		labels = append(labels, date)
		stime = stime.AddDate(0, 0, 1)
	}
	return
}

/*
金币流水曲线
*/
func (this *statisticsService) GetGoldFlowCurveInfo_old(userid, start_date, end_date string, game_start, game_end, latelygame int) ([]interface{}, []interface{}, []interface{}, []interface{}, []interface{}) {
	dates := make([]interface{}, 0) // 日期
	data := make([]interface{}, 0)  // 携带
	data1 := make([]interface{}, 0) // 盈亏
	data2 := make([]interface{}, 0) // 充值
	cause := make([]interface{}, 0) // 携带原因

	var list []entity.Detail
	n := FindByDate1(start_date, end_date, "end_time", "end_time")
	n["players"] = bson.M{
		"$regex":   fmt.Sprintf("\\b%s\\b", userid),
		"$options": "i",
	}
	total := Count(Details, n)
	tem_limit := 0
	if game_start == 0 && game_end == 0 && latelygame == 0 {
		Details.Find(n).Sort("-end_time").All(&list)
	} else if game_start != 0 && game_end != 0 {
		// tem_limit = total - game_end - game_start
		tem_limit = game_start - game_end
		//Details.Find(n).Sort("-end_time").Skip(game_start).Limit(tem_limit).All(&list)
		Details.Find(n).Sort("-end_time").Skip(game_end).Limit(tem_limit).All(&list)
	} else if game_start != 0 {
		Details.Find(n).Sort("-end_time").Skip(game_end).Limit(game_start).All(&list)
	} else if game_end != 0 {
		tem_limit = total - game_end
		// tem_limit = game_end - game_start
		Details.Find(n).Sort("-end_time").Skip(game_end).Limit(tem_limit).All(&list)
	} else if latelygame != 0 {
		tem_limit = total - latelygame
		Details.Find(n).Sort("-end_time").Limit(latelygame).All(&list)
	}
	// 反转数据列表
	sort.Slice(list, func(i, j int) bool {
		return list[i].EndTime < list[j].EndTime
	})
	if len(list) > 0 {
		// last_time := int64(0)
		for index, item := range list {
			index += 1
			dates = append(dates, index)
			curee_coin := int64(0)
			cause_str := ""
			// 判断是那个游戏
			switch item.Gtype {
			case 1:
				if item.TPDetail != nil {
					for _, v := range item.TPDetail {
						if v.UserId == userid {
							curee_coin = v.AfterCash
							js := Chip2Float(v.Score)
							cause_str = "TP结算：" + fmt.Sprintf("%.2f", js)
						}
					}
				}
			case 2:
				// lhd
				if item.LHDetail.UserDetail != nil {
					for _, v := range item.LHDetail.UserDetail {
						if v.Userid == userid {
							curee_coin = v.AfterCash
							js := Chip2Float(v.Win)
							cause_str = "DRAGON TIGER结算：" + fmt.Sprintf("%.2f", js)
						}
					}
				}
			case 3:
				// 7updown
				if item.UPDetail.UserDetail != nil {
					for _, v := range item.UPDetail.UserDetail {
						if v.Userid == userid {
							curee_coin = v.AfterCash
							js := Chip2Float(v.Win)
							cause_str = "7UPDOWN结算：" + fmt.Sprintf("%.2f", js)
						}
					}
				}
			case 4:
				// rummy
				if item.RMDetail != nil {
					for _, v := range item.RMDetail {
						if v.UserId == userid {
							curee_coin = v.AfterCash
							js := Chip2Float(v.Score)
							cause_str = "RUMMY结算：" + fmt.Sprintf("%.2f", js)
						}
					}
				}
			case 5:
				// ak47
				if item.AK47Detail != nil {
					for _, v := range item.AK47Detail {
						if v.UserId == userid {
							curee_coin = v.AfterCash
							js := Chip2Float(v.Score)
							cause_str = "AK47结算：" + fmt.Sprintf("%.2f", js)
						}
					}
				}
			case 6:
				// joker
				if item.JOKERDetail != nil {
					for _, v := range item.JOKERDetail {
						if v.UserId == userid {
							curee_coin = v.AfterCash
							js := Chip2Float(v.Score)
							cause_str = "JOKER结算：" + fmt.Sprintf("%.2f", js)
						}
					}
				}
			case 7:
				// crash
				if item.CRASHDetail.UserDetail != nil {
					for _, v := range item.CRASHDetail.UserDetail {
						if v.Userid == userid {
							curee_coin = v.AfterCash
							js := Chip2Float(v.Win)
							cause_str = "CRASH结算：" + fmt.Sprintf("%.2f", js)
						}
					}
				}
			case 8:
				// ANDARBAHAR
				if item.ABDetail.UserDetail != nil {
					for _, v := range item.ABDetail.UserDetail {
						if v.Userid == userid {
							curee_coin = v.AfterCash
							js := Chip2Float(v.Win)
							cause_str = "ANDARBAHAR结算：" + fmt.Sprintf("%.2f", js)
						}
					}
				}
			case 9:
				// cp
				if item.CPDetail.UserDetail != nil {
					for _, v := range item.CPDetail.UserDetail {
						if v.Userid == userid {
							curee_coin = v.AfterCash
							js := Chip2Float(v.Win)
							cause_str = "彩票结算：" + fmt.Sprintf("%.2f", js)
						}
					}
				}
			case 10:
				// 飞机
				if item.CRASHDetail.UserDetail != nil {
					for _, v := range item.CRASHDetail.UserDetail {
						if v.Userid == userid {
							curee_coin = v.AfterCash
							js := Chip2Float(v.Win)
							cause_str = "飞机结算：" + fmt.Sprintf("%.2f", js)
						}
					}
				}
			case 11:
				if item.RBDetail.UserDetail != nil {
					for _, v := range item.RBDetail.UserDetail {
						if v.Userid == userid {
							curee_coin = v.AfterCash
							js := Chip2Float(v.Win)
							cause_str = "红黑大战结算：" + fmt.Sprintf("%.2f", js)
						}
					}
				}
			case 12:
				// RUMMY双人
				if item.RMDetail != nil {
					for _, v := range item.RMDetail {
						if v.UserId == userid {
							curee_coin = v.AfterCash
							js := Chip2Float(v.Score)
							cause_str = "RUMMY双人结算：" + fmt.Sprintf("%.2f", js)
						}
					}
				}
			case 13:
				// TP2
				if item.TPDetail != nil {
					for _, v := range item.TPDetail {
						if v.UserId == userid {
							curee_coin = v.AfterCash
							js := Chip2Float(v.Score)
							cause_str = "TP2结算：" + fmt.Sprintf("%.2f", js)
						}
					}
				}
			}
			data = append(data, Chip2Float(curee_coin))

			// 总盈亏
			pay_amount := 0
			withdraw_amount := 0
			// 获取当前提现
			end_time := time.Unix(item.EndTime, 0).UTC()
			pipeline := []bson.M{
				{
					"$match": bson.M{
						"userid":       userid,
						"order_status": 2,
						"pay_time":     bson.M{"$lt": end_time},
					},
				},
				{
					"$group": bson.M{
						"_id": "$userid",
						"Amount": bson.M{
							"$sum": "$amount",
						},
					},
				},
			}
			result := []bson.M{}
			pipe := Withdraws.Pipe(pipeline)
			pipe.All(&result)
			if len(result) > 0 {
				withdraw_amount = result[0]["Amount"].(int)
			}

			// 总充值
			pipeline1 := []bson.M{
				{
					"$match": bson.M{
						"userid":       userid,
						"order_status": 4,
						"pay_time":     bson.M{"$lt": end_time},
					},
				},
				{
					"$group": bson.M{
						"_id": "$userid",
						"Amount": bson.M{
							"$sum": "$amount",
						},
					},
				},
			}
			result1 := []bson.M{}
			pipe1 := Pays.Pipe(pipeline1)
			pipe1.All(&result1)
			if len(result1) > 0 {
				pay_amount = result1[0]["Amount"].(int)
			}
			data2 = append(data2, Chip2Float(int64(pay_amount)))
			gain_amout := (int64(withdraw_amount) + curee_coin) - int64(pay_amount)
			data1 = append(data1, Chip2Float(int64(gain_amout)))
			pay_float := fmt.Sprintf("%.2f", Chip2Float(int64(pay_amount)))
			cause = append(cause, bson.M{"id": index, "wid": item.WaterId, "msg": cause_str, "pay": pay_float})
		}
	}
	return dates, data, data1, data2, cause
}

// 金币流水曲线
func (this *statisticsService) GetGoldFlowCurveInfo(userid, start_date, end_date string, game_start, game_end, latelygame int) ([]interface{}, []interface{}, []interface{}, []interface{}, []interface{}, []interface{}, error) {
	dates := make([]interface{}, 0) // 日期
	data := make([]interface{}, 0)  // 携带
	data1 := make([]interface{}, 0) // 盈亏
	data2 := make([]interface{}, 0) // 充值
	data3 := make([]interface{}, 0) // 提现
	cause := make([]interface{}, 0) // 携带原因
	var list []entity.Detail
	var d_list []entity.Detail
	var nsq_list []entity.NsqBet
	n := FindByDate1(start_date, end_date, "ctime", "ctime")
	n["user_id"] = userid
	n["amount"] = bson.M{"$ne": 0}
	// 外接数据
	pipeline := []bson.M{
		{
			"$match": n,
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"round_id": "$round_id",
					"ctime":    "$ctime",
				},
			},
		},
		{
			"$project": bson.M{
				"_id":   "$_id.round_id",
				"ctime": "$_id.ctime",
			},
		},
	}
	result := make([]entity.MergeDetail, 0)
	pipe := NsqLogExternalBets.Pipe(pipeline)
	err := pipe.All(&result)
	// 对局数据
	n = FindByDate1(start_date, end_date, "begin_time", "begin_time")
	n["players"] = bson.M{
		"$regex":   fmt.Sprintf("\\b%s\\b", userid),
		"$options": "i",
	}
	pipeline1 := []bson.M{
		{
			"$match": n,
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"WaterId": "$_id",
					"ctime":   "$begin_time",
				},
			},
		},
		{
			"$project": bson.M{
				"_id":   "$_id.WaterId",
				"ctime": "$_id.ctime",
			},
		},
	}
	result1 := make([]entity.MergeDetail, 0)
	pipe1 := Details.Pipe(pipeline1)
	err = pipe1.All(&result1)
	det_map := make([]entity.MergeDetail, 0)
	det_map = append(det_map, result...)
	det_map = append(det_map, result1...)

	// 对det_map按照ctime字段进行倒序排序
	sort.Slice(det_map, func(i, j int) bool {
		return det_map[i].Ctime > det_map[j].Ctime
	})
	total := len(det_map)
	tem_limit := 0
	var paginatedList []entity.MergeDetail
	if game_start == 0 && game_end == 0 && latelygame == 0 {
		paginatedList = det_map
	} else if game_start != 0 && game_end != 0 {
		tem_limit = game_start - game_end
		paginatedList = det_map[game_end:tem_limit]
	} else if game_start != 0 {
		paginatedList = det_map[game_end:game_start]
	} else if game_end != 0 {
		tem_limit = total - game_end
		paginatedList = det_map[game_end:tem_limit]
	} else if latelygame != 0 {
		// tem_limit = total - latelygame
		if total < latelygame {
			tem_limit = total
		} else {
			tem_limit = latelygame
		}
		paginatedList = det_map[0:tem_limit]
	}
	var ids []string
	for _, v := range paginatedList {
		ids = append(ids, v.Id)
	}
	query := bson.M{}
	query["round_id"] = bson.M{"$in": ids}
	query["amount"] = bson.M{"$ne": 0}
	NsqLogExternalBets.Find(query).Sort("-ctime").All(&nsq_list)
	query = bson.M{}
	query["_id"] = bson.M{"$in": ids}
	Details.Find(query).Sort("-begin_time").All(&d_list)
	list = append(list, d_list...)
	var round_arr []string
	for _, item := range nsq_list {
		info := new(entity.Detail)
		info.WaterId = item.RoundId
		info.BeginTime = item.Ctime
		info.EndTime = item.Ctime
		info.Gtype = item.GameId
		info.RoomId = strconv.FormatInt(int64(item.GameId), 10)
		info.DeskId = item.MerchantOrderNo
		info.Players = item.UserId
		round_arr = append(round_arr, item.RoundId)
		list = append(list, *info)
	}
	// 外接返奖
	pipeline2 := []bson.M{
		{
			"$match": bson.M{
				"round_id": bson.M{"$in": round_arr},
				"user_id":  userid,
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"RoundId": "$round_id",
					"UserId":  "$user_id",
				},
				"Amount": bson.M{"$sum": "$amount"},
			},
		},
		{
			"$project": bson.M{
				"_id":     "$_id.RoundId",
				"user_id": "$_id.UserId",
				"amount":  "$Amount",
			},
		},
	}
	reward_list := make([]entity.ExternalReward, 0)
	pipe2 := NsqLogExternalRewards.Pipe(pipeline2)
	err = pipe2.All(&reward_list)

	sort.Slice(list, func(i, j int) bool {
		return list[i].BeginTime < list[j].BeginTime
	})

	// 游戏列表
	recordList := map[int]string{0: "全部"}
	utils.CopyMap(recordList, GtypeNameMap)
	if len(list) > 0 {
		last_time := list[0].BeginTime
		number := len(list)
		for index, item := range list {
			index += 1
			end_time := item.EndTime
			if number < index {
				end_time = list[index].BeginTime
			}
			dates = append(dates, index)
			curee_coin := int64(0)
			cause_str := ""
			isNsq := false
			// 判断是那个游戏
			switch item.Gtype {
			case 1:
				if item.TPDetail != nil {
					for _, v := range item.TPDetail {
						if v.UserId == userid {
							curee_coin = v.AfterCash
							// js := Chip2Float(v.Score)
							// cause_str = "TP结算：" + fmt.Sprintf("%.2f", js)
							cause_str = "TP"
						}
					}
				}
			case 2:
				// lhd
				if item.LHDetail.UserDetail != nil {
					for _, v := range item.LHDetail.UserDetail {
						if v.Userid == userid {
							curee_coin = v.AfterCash
							// js := Chip2Float(v.Win)
							// cause_str = "DRAGON TIGER结算：" + fmt.Sprintf("%.2f", js)
							cause_str = "DRAGON TIGER"
						}
					}
				}
			case 3:
				// 7updown
				if item.UPDetail.UserDetail != nil {
					for _, v := range item.UPDetail.UserDetail {
						if v.Userid == userid {
							curee_coin = v.AfterCash
							// js := Chip2Float(v.Win)
							// cause_str = "7UPDOWN结算：" + fmt.Sprintf("%.2f", js)
							cause_str = "7UPDOWN"
						}
					}
				}
			case 4:
				// rummy
				if item.RMDetail != nil {
					for _, v := range item.RMDetail {
						if v.UserId == userid {
							curee_coin = v.AfterCash
							// js := Chip2Float(v.Score)
							// cause_str = "RUMMY结算：" + fmt.Sprintf("%.2f", js)
							cause_str = "RUMMY"
						}
					}
				}
			case 5:
				// ak47
				if item.AK47Detail != nil {
					for _, v := range item.AK47Detail {
						if v.UserId == userid {
							curee_coin = v.AfterCash
							// js := Chip2Float(v.Score)
							// cause_str = "AK47结算：" + fmt.Sprintf("%.2f", js)
							cause_str = "AK47"
						}
					}
				}
			case 6:
				// joker
				if item.JOKERDetail != nil {
					for _, v := range item.JOKERDetail {
						if v.UserId == userid {
							curee_coin = v.AfterCash
							// js := Chip2Float(v.Score)
							// cause_str = "JOKER结算：" + fmt.Sprintf("%.2f", js)
							cause_str = "JOKER"
						}
					}
				}
			case 7:
				// crash
				if item.CRASHDetail.UserDetail != nil {
					for _, v := range item.CRASHDetail.UserDetail {
						if v.Userid == userid {
							curee_coin = v.AfterCash
							// js := Chip2Float(v.Win)
							// cause_str = "CRASH结算：" + fmt.Sprintf("%.2f", js)
							cause_str = "CRASH"
						}
					}
				}
			case 8:
				// ANDARBAHAR
				if item.ABDetail.UserDetail != nil {
					for _, v := range item.ABDetail.UserDetail {
						if v.Userid == userid {
							curee_coin = v.AfterCash
							// js := Chip2Float(v.Win)
							// cause_str = "ANDARBAHAR结算：" + fmt.Sprintf("%.2f", js)
							cause_str = "ANDARBAHAR"
						}
					}
				}
			case 9:
				// cp
				if item.CPDetail.UserDetail != nil {
					for _, v := range item.CPDetail.UserDetail {
						if v.Userid == userid {
							curee_coin = v.AfterCash
							// js := Chip2Float(v.Win)
							// cause_str = "彩票结算：" + fmt.Sprintf("%.2f", js)
							cause_str = "彩票"
						}
					}
				}
			case 10:
				// 飞机
				if item.CRASHDetail.UserDetail != nil {
					for _, v := range item.CRASHDetail.UserDetail {
						if v.Userid == userid {
							curee_coin = v.AfterCash
							// js := Chip2Float(v.Win)
							// cause_str = "飞机结算：" + fmt.Sprintf("%.2f", js)
							cause_str = "飞机"
						}
					}
				}
			case 11:
				// 飞机
				if item.RBDetail.UserDetail != nil {
					for _, v := range item.RBDetail.UserDetail {
						if v.Userid == userid {
							curee_coin = v.AfterCash
							// js := Chip2Float(v.Win)
							// cause_str = "飞机结算：" + fmt.Sprintf("%.2f", js)
							cause_str = "红黑大战"
						}
					}
				}
			case 12:
				// rummy双人
				if item.RMDetail != nil {
					for _, v := range item.RMDetail {
						if v.UserId == userid {
							curee_coin = v.AfterCash
							// js := Chip2Float(v.Score)
							// cause_str = "RUMMY结算：" + fmt.Sprintf("%.2f", js)
							cause_str = "RUMMY双人"
						}
					}
				}
			case 13:
				// TP2
				if item.TPDetail != nil {
					for _, v := range item.TPDetail {
						if v.UserId == userid {
							curee_coin = v.AfterCash
							js := Chip2Float(v.Score)
							cause_str = "TP2结算：" + fmt.Sprintf("%.2f", js)
						}
					}
				}
			default:
				if len(reward_list) > 0 {
					isNsq = true
					win := int64(0)
					for _, r := range reward_list {
						if item.WaterId == r.Id {
							win += r.Amount
						}
					}
					for k, v := range recordList {
						if k == int(item.Gtype) {
							cause_str = v
						}
					}
					// js := Chip2Float(win)
					// cause_str = "外接测试结算：" + fmt.Sprintf("%.2f", js)
				}

			}
			// 总盈亏
			pay_amount := 0
			withdraw_amount := 0
			// 获取当前提现
			end_date := time.Unix(end_time, 0).UTC().Add(time.Second)
			start_date := time.Unix(last_time, 0).UTC()

			// 外接携带金币计算
			if isNsq {
				// 金币流水
				query := bson.M{}
				query["userid"] = userid
				//query["ctime"] = bson.M{"$gte": start_time, "$lte": end_time}
				query["water_id"] = item.WaterId
				query["water_desc"] = bson.M{"$in": []string{"外接结算", "外接下注"}}
				nsq_water := new(entity.LogWater)
				LogWaters.Find(query).Sort("-ctime").One(&nsq_water)
				if nsq_water.Id != "" {
					curee_coin = nsq_water.NowDiamond
				}
			}

			pipeline := []bson.M{
				{
					"$match": bson.M{
						"userid":       userid,
						"order_status": 2,
						"pay_time":     bson.M{"$lte": end_date},
					},
				},
				{
					"$group": bson.M{
						"_id": "$userid",
						"Amount": bson.M{
							"$sum": "$amount",
						},
					},
				},
			}
			result := []bson.M{}
			pipe := Withdraws.Pipe(pipeline)
			pipe.All(&result)
			if len(result) > 0 {
				withdraw_amount = result[0]["Amount"].(int)
			}

			// 总充值
			pipeline1 := []bson.M{
				{
					"$match": bson.M{
						"userid":       userid,
						"order_status": 4,
						"pay_time":     bson.M{"$lt": end_date},
					},
				},
				{
					"$group": bson.M{
						"_id": "$userid",
						"Amount": bson.M{
							"$sum": "$amount",
						},
					},
				},
			}
			result1 := []bson.M{}
			pipe1 := Pays.Pipe(pipeline1)
			pipe1.All(&result1)
			if len(result1) > 0 {
				pay_amount = result1[0]["Amount"].(int)
			}
			user_amount := 0
			// 充值金额
			pipeline3 := []bson.M{
				{
					"$match": bson.M{
						"userid":       userid,
						"order_status": 4,
						"pay_time":     bson.M{"$gte": start_date, "$lte": end_date},
					},
				},
				{
					"$group": bson.M{
						"_id": "$userid",
						"Amount": bson.M{
							"$sum": "$amount",
						},
					},
				},
			}
			result3 := []bson.M{}
			pipe3 := Pays.Pipe(pipeline3)
			pipe3.All(&result3)
			if len(result3) > 0 {
				user_amount = result3[0]["Amount"].(int)
			}
			user_withdraw_amount := 0
			// 提现金额
			pipeline4 := []bson.M{
				{
					"$match": bson.M{
						"userid":       userid,
						"order_status": 2,
						"pay_time":     bson.M{"$gte": start_date, "$lte": end_date},
					},
				},
				{
					"$group": bson.M{
						"_id": "$userid",
						"Amount": bson.M{
							"$sum": "$amount",
						},
					},
				},
			}
			result4 := []bson.M{}
			pipe4 := Withdraws.Pipe(pipeline4)
			pipe4.All(&result4)
			if len(result4) > 0 {
				user_withdraw_amount = result4[0]["Amount"].(int)
			}
			// 金币流水
			query := bson.M{}
			query["userid"] = userid
			query["$or"] = []bson.M{
				{"water_id": item.WaterId},
				{"water_id": ""},
			}
			if number == index {
				query["ctime"] = bson.M{"$gte": start_date}
			} else {
				query["ctime"] = bson.M{"$gte": start_date, "$lte": end_date}
			}
			pipeline2 := []bson.M{
				{
					"$match": query,
				},
				{
					"$project": bson.M{
						"_id":         "$_id",
						"user_id":     "$userid",
						"add_diamond": "$add_diamond",
						"ltype":       "$ltype",
						"water_desc":  "$water_desc",
					},
				},
				{
					"$sort": bson.M{"ctime": 1},
				},
			}
			result2 := []bson.M{}
			pipe2 := LogWaters.Pipe(pipeline2)
			pipe2.All(&result2)
			var buf bytes.Buffer
			if len(result2) > 0 {
				for _, b := range result2 {
					adddiamond := b["add_diamond"].(int64)
					water_desc := b["water_desc"].(string)
					ltype := b["ltype"].(int)
					stropt := ConvertToTypeMsg(ltype)
					if stropt == "" {
						stropt = water_desc
					}
					f_diamond := Chip2Float(adddiamond)
					buf.WriteString(stropt + ":" + fmt.Sprintf("%.2f", f_diamond))
					buf.WriteString("\r\n")
				}
			}
			data = append(data, Chip2Float(curee_coin))
			data2 = append(data2, Chip2Float(int64(user_amount)))
			data3 = append(data3, Chip2Float(int64(user_withdraw_amount)))

			gain_amout := (int64(withdraw_amount) + curee_coin) - int64(pay_amount)
			data1 = append(data1, Chip2Float(int64(gain_amout)))
			pay_float := fmt.Sprintf("%.2f", Chip2Float(int64(pay_amount)))
			cause = append(cause, bson.M{"id": index, "wid": item.WaterId, "game": cause_str, "msg": buf.String(), "pay": pay_float})

			// 修改上次最后时间
			last_time = item.EndTime
		}
	}
	return dates, data, data1, data2, data3, cause, err
}

func (this *statisticsService) GetGoldFlowCurveInfoCK(userid, start_date, end_date string, game_start, game_end, latelygame int) ([]interface{}, []interface{}, []interface{}, []interface{}, []interface{}, []interface{}, error) {
	dates := make([]interface{}, 0) // 日期
	data := make([]interface{}, 0)  // 携带
	data1 := make([]interface{}, 0) // 盈亏
	data2 := make([]interface{}, 0) // 充值
	data3 := make([]interface{}, 0) // 提现
	cause := make([]interface{}, 0) // 携带原因
	var list []entity.UserDetail
	// var d_list []entity.Detail
	// var nsq_list []entity.NsqBet
	// 	sql1 := `
	// 	sselect %s from (
	// 		select id,userid,begin_time ctime from game.col_detail t1 FINAL WHERE robot = 0 and win_type in (1,2,3) %s
	// 	UNION ALL
	// 	select round_id id,user_id userid,ctime from game.col_nsq_log_external_bet t2 FINAL WHERE amount != 0 and cancel = 0 %s ) tab %s
	// `
	sql1 := `SELECT %s FROM (SELECT id,userid,begin_time,end_time,gtype,score_untax,after_score FROM game.col_detail t1 FINAL WHERE robot=0 AND win_type IN (1,2,3) %s order by begin_time desc
UNION ALL
SELECT round_id AS id,user_id AS userid,ctime AS begin_time,ctime AS end_time,game_id AS gtype,res_amount AS score_untax,0 AS after_score FROM game.col_nsq_log_external_bet t2 FINAL 
LEFT JOIN (SELECT round_id,user_id,SUM(amount) res_amount FROM game.col_nsq_log_external_reward cnler FINAL WHERE amount != 0 and cancel = 0 GROUP BY round_id,user_id) red 
ON t2.round_id = red.round_id AND t2.user_id = red.user_id
WHERE amount != 0 and cancel = 0 %s ) tab %s`
	where1, where2 := "", ""
	var args1, args2 []any
	if userid != "" {
		where1 += " and t1.userid = ?"
		args1 = append(args1, userid)
		where2 += " and t2.user_id = ?"
		args2 = append(args2, userid)
	}

	if start_date != "" && end_date != "" {
		s := fmt.Sprintf("%s 00:00:00", start_date)
		s1 := utils.Str2Time(s, Location())
		startTime := utils.Time2Stamp(s1)
		e := fmt.Sprintf("%s 23:59:59", end_date)
		s2 := utils.Str2Time(e, Location())
		endTime := utils.Time2Stamp(s2)
		where1 += " and t1.begin_time between ? and ?"
		args1 = append(args1, startTime, endTime)
		where2 += " and t2.ctime between ? and ?"
		args2 = append(args2, startTime, endTime)
	} else if start_date != "" && end_date == "" {
		s := fmt.Sprintf("%s 23:59:59", start_date)
		s1 := utils.Str2Time(s, Location())
		startTime := utils.Time2Stamp(s1)
		where1 += " and t1.begin_time >= ?"
		args1 = append(args1, startTime)
		where2 += " and t2.ctime >= ?"
		args2 = append(args2, startTime)
	} else if end_date != "" && start_date == "" {
		e := fmt.Sprintf("%s 23:59:59", end_date)
		s2 := utils.Str2Time(e, Location())
		endTime := utils.Time2Stamp(s2)
		where1 += " and t1.begin_time <= ?"
		args1 = append(args1, endTime)
		where2 += " and t2.ctime <= ?"
		args2 = append(args2, endTime)
	}
	total := 0
	sql_count := fmt.Sprintf(sql1, "count(*) c", where1, where2, "")
	args_count := append(args1, args2...)
	err := ck.Select(&total, sql_count, args_count...)
	// if err != nil || total == 0 {
	// 	return
	// }
	page := 0
	pageSize := 0
	tem_limit := 0
	// var paginatedList []entity.MergeDetail
	if game_start == 0 && game_end == 0 && latelygame == 0 {
		// 默认100条
		pageSize = 100
	} else if game_start != 0 && game_end != 0 {
		tem_limit = game_start - game_end
		page = game_end
		pageSize = tem_limit
		// paginatedList = det_map[game_end:tem_limit]
	} else if game_start != 0 {
		page = game_end
		pageSize = game_start
		// paginatedList = det_map[game_end:game_start]
	} else if game_end != 0 {
		tem_limit = total - game_end
		page = game_end
		pageSize = tem_limit
		// paginatedList = det_map[game_end:tem_limit]
	} else if latelygame != 0 {
		tem_limit = total - latelygame
		if total < latelygame {
			tem_limit = total
		} else {
			tem_limit = latelygame
		}
		// paginatedList = det_map[0:tem_limit]
		page = 0
		pageSize = tem_limit
	}

	selects := "*"
	order := " ORDER BY begin_time DESC LIMIT ?, ?"
	sql_list := fmt.Sprintf(sql1, selects, where1, where2, order)
	args_list := append(args1, args2...)
	offset, limit := PageCalc(page, pageSize)
	args_list = append(args_list, offset, limit)
	err = ck.Select(&list, sql_list, args_list...)

	// 游戏列表
	recordList := map[int]string{0: "全部"}
	utils.CopyMap(recordList, GtypeNameMap)
	if len(list) > 0 {
		// 对det_map按照ctime字段进行倒序排序
		sort.Slice(list, func(i, j int) bool {
			return list[i].BeginTime < list[j].BeginTime
		})
		last_time := list[0].BeginTime
		for index, item := range list {
			index += 1
			end_time := item.EndTime
			if total < index {
				end_time = list[index].BeginTime
			}
			dates = append(dates, index)
			curee_coin := int64(0)
			cause_str := ""
			isNsq := false
			curee_coin = item.AfterScore
			gtypename, ok := recordList[int(item.Gtype)]
			if ok {
				cause_str = gtypename
			}
			if item.Gtype > 1000 {
				isNsq = true
			}

			// 外接携带金币计算
			if isNsq {
				// 金币流水
				query := bson.M{}
				query["userid"] = userid
				//query["ctime"] = bson.M{"$gte": start_time, "$lte": end_time}
				query["water_id"] = item.WaterId
				query["water_desc"] = bson.M{"$in": []string{"外接结算", "外接下注"}}
				nsq_water := new(entity.LogWater)
				LogWaters.Find(query).Sort("-ctime").One(&nsq_water)
				if nsq_water.Id != "" {
					curee_coin = nsq_water.NowDiamond
				}
			}
			pay_amount := 0
			user_amount := 0
			withdraw_amount := 0
			user_withdraw_amount := 0
			// 获取当前提现
			end_date := time.Unix(end_time, 0).UTC().Add(time.Second)
			start_date := time.Unix(last_time, 0).UTC()
			// 获取总提现金额
			var with []map[string]any
			var args3 []any
			args3 = append(args3, end_date)
			args3 = append(args3, locationName)
			args3 = append(args3, start_date)
			args3 = append(args3, locationName)
			args3 = append(args3, end_date)
			args3 = append(args3, locationName)
			args3 = append(args3, userid)
			err = ck.Select(&with, `
			SELECT SUM(CASE WHEN ctime <= toDateTime(?, ?) THEN amount ELSE 0 END) AS withdraw_amount,
 SUM(CASE 
        WHEN ctime >= toDateTime(?, ?) 
             AND ctime <= toDateTime(?, ?) 
        THEN amount 
        ELSE 0 
    END) AS user_amount 
FROM game.col_withdraw_record cwr FINAL WHERE order_status = 2 AND userid = ? 
		`, args3...)
			for _, w := range with {
				c1 := w["withdraw_amount"].(uint64)
				if c1 > 0 {
					withdraw_amount = int(c1)
				}
				c2 := w["user_amount"].(uint64)
				if c2 > 0 {
					user_withdraw_amount = int(c2)
				}
			}

			// 获取总充值
			var pay []map[string]any
			var args4 []any
			args4 = append(args4, end_date)
			args4 = append(args4, locationName)
			args4 = append(args4, start_date)
			args4 = append(args4, locationName)
			args4 = append(args4, end_date)
			args4 = append(args4, locationName)
			args4 = append(args4, userid)
			err = ck.Select(&pay, `
			SELECT SUM(CASE WHEN ctime <= toDateTime(?, ?) THEN amount ELSE 0 END) AS pay_amount,
 SUM(CASE 
        WHEN ctime >= toDateTime(?, ?) 
             AND ctime <= toDateTime(?, ?) 
        THEN amount 
        ELSE 0 
    END) AS user_amount 
FROM game.col_trade_record ctr FINAL WHERE order_status = 4 AND userid = ? 
		`, args4...)
			for _, w := range pay {
				c1 := w["pay_amount"].(uint64)
				if c1 > 0 {
					pay_amount = int(c1)
				}
				c2 := w["user_amount"].(uint64)
				if c2 > 0 {
					user_amount = int(c2)
				}
			}

			// 金币流水
			var gold []map[string]any
			var args5 []any
			args5 = append(args5, userid)
			args5 = append(args5, item.WaterId)
			err = ck.Select(&gold, `
			SELECT  userid,add_diamond,ltype,water_desc FROM game.col_log_water clw FINAL WHERE userid = ? AND water_id != '' AND water_id = ? ORDER BY ctime asc 
		`, args5...)
			var buf bytes.Buffer
			if len(gold) > 0 {
				for _, b := range gold {
					adddiamond := b["add_diamond"].(int64)
					water_desc := b["water_desc"].(string)
					ltype := b["ltype"].(int32)
					stropt := ConvertToTypeMsg(int(ltype))
					if stropt == "" {
						stropt = water_desc
					}
					f_diamond := Chip2Float(adddiamond)
					buf.WriteString(stropt + ":" + fmt.Sprintf("%.2f", f_diamond))
					buf.WriteString("\r\n")
				}
			}
			// query := bson.M{}
			// query["userid"] = userid
			// // query["$or"] = []bson.M{
			// // 	{"water_id": item.WaterId},
			// // 	{"water_id": ""},
			// // }
			// query["water_id"] = item.WaterId
			// // if total == index {
			// // 	query["ctime"] = bson.M{"$gte": start_date}
			// // } else {
			// // 	query["ctime"] = bson.M{"$gte": start_date, "$lte": end_date}
			// // }
			// pipeline2 := []bson.M{
			// 	{
			// 		"$match": query,
			// 	},
			// 	{
			// 		"$project": bson.M{
			// 			"_id":         "$_id",
			// 			"user_id":     "$userid",
			// 			"add_diamond": "$add_diamond",
			// 			"ltype":       "$ltype",
			// 			"water_desc":  "$water_desc",
			// 		},
			// 	},
			// 	{
			// 		"$sort": bson.M{"ctime": 1},
			// 	},
			// }
			// result2 := []bson.M{}
			// pipe2 := LogWaters.Pipe(pipeline2)
			// pipe2.All(&result2)
			// var buf bytes.Buffer
			// if len(result2) > 0 {
			// 	for _, b := range result2 {
			// 		adddiamond := b["add_diamond"].(int64)
			// 		water_desc := b["water_desc"].(string)
			// 		ltype := b["ltype"].(int)
			// 		stropt := ConvertToTypeMsg(ltype)
			// 		if stropt == "" {
			// 			stropt = water_desc
			// 		}
			// 		f_diamond := Chip2Float(adddiamond)
			// 		buf.WriteString(stropt + ":" + fmt.Sprintf("%.2f", f_diamond))
			// 		buf.WriteString("\r\n")
			// 	}
			// }
			data = append(data, Chip2Float(curee_coin))
			data2 = append(data2, Chip2Float(int64(user_amount)))
			data3 = append(data3, Chip2Float(int64(user_withdraw_amount)))

			gain_amout := (int64(withdraw_amount) + curee_coin) - int64(pay_amount)
			data1 = append(data1, Chip2Float(int64(gain_amout)))
			pay_float := fmt.Sprintf("%.2f", Chip2Float(int64(pay_amount)))
			cause = append(cause, bson.M{"id": index, "wid": item.WaterId, "game": cause_str, "msg": buf.String(), "pay": pay_float})

			// 修改上次最后时间
			last_time = item.EndTime

		}
	}

	return dates, data, data1, data2, data3, cause, err
}

func ConvertToTypeMsg(ltype int) string {
	strOtp := ""
	if ltype == 106 || ltype == 127 || ltype == 130 || ltype == 110 || ltype == 112 || ltype == 114 || ltype == 116 || ltype == 87 || ltype == 90 || ltype == 100 || ltype == 75 || ltype == 76 || ltype == 79 {
		strOtp = "下注"
	}
	if ltype == 107 || ltype == 108 || ltype == 128 || ltype == 131 || ltype == 111 || ltype == 113 || ltype == 115 || ltype == 117 || ltype == 67 || ltype == 70 || ltype == 73 || ltype == 88 || ltype == 91 || ltype == 101 {
		strOtp = "结算"
	}
	if ltype == 89 || ltype == 92 || ltype == 69 || ltype == 72 || ltype == 109 {
		strOtp = "扣税"
	}
	if ltype == 61 {
		strOtp = "签到"
	}
	return strOtp
}

// 玩家充值统计
func (this *statisticsService) GetPlayerRecharge(uid []string, num int) ([]entity.PlayerRechargeInfo, error) {
	var list []entity.PlayerRechargeInfo
	shopList := map[int]string{
		1: "商场直充",
		2: "入门礼包",
		3: "金银铜卡",
		4: "首充",
		5: "局内充值",
		6: "限时礼包",
	}

	// 游戏列表
	recordList := map[int]string{0: "全部"}
	utils.CopyMap(recordList, GtypeNameMap)

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"userid":       bson.M{"$in": uid},
				"order_status": 4,
			},
		},
		{
			"$group": bson.M{
				"_id":    "$userid",
				"amount": bson.M{"$sum": "$amount"},
				"num":    bson.M{"$sum": 1},
			},
		},
	}
	var res []bson.M
	err := Pays.Pipe(pipeline).All(&res)

	for _, id := range uid {
		new_info := new(entity.PlayerRechargeInfo)
		isPay := false
		if len(res) > 0 {
			for _, rs := range res {
				pid := rs["_id"].(string)
				if pid == id {
					num := rs["num"].(int)
					amount := rs["amount"].(int)
					new_info.Userid = pid
					new_info.TotalCount = int64(num)
					new_info.TotalAmount = int64(amount)
					famount := Chip2Float(int64(amount))
					new_info.TotalAvg = famount / float64(num)
					isPay = true
				}
			}
		}
		if !isPay {
			new_info.Userid = id
			new_info.TotalCount = 0
			new_info.TotalAvg = 0
		}
		pipeline1 := []bson.M{
			{
				"$match": bson.M{
					"userid":       id,
					"order_status": 4,
				},
			},
			{
				"$project": bson.M{
					"_id":       "$_id",
					"userid":    "$userid",
					"amount":    "$amount",
					"shop_type": "$shop_type",
				},
			},
			{"$sort": bson.M{"ctime": 1}},
			{"$limit": num},
		}

		var result1 []bson.M
		err = Pays.Pipe(pipeline1).All(&result1)

		d_info := make([]entity.PlayerRechargeDetails, 0)
		// 获取充值前打码量最多的游戏
		query := bson.M{}
		query["_id"] = id
		var water_info entity.LogGameFlowWater
		err = LogGameFlowWaters.Find(query).One(&water_info)
		if len(result1) > 0 {
			for k, item := range result1 {
				shop_type := item["shop_type"].(int)
				amount := item["amount"].(int)
				temp_dt := new(entity.PlayerRechargeDetails)
				temp_dt.ShopType = int64(shop_type)
				temp_dt.Amount = int64(amount)
				temp_dt.FAmount = Chip2Float(int64(amount))
				name, ok := shopList[shop_type]
				if ok {
					temp_dt.ShopName = name
				}
				if water_info.Userid != "" {
					if water_info.Detail != nil {
						for j, v := range water_info.Detail {
							if k == (j) {
								if v != nil {
									sort.Slice(v, func(i, j int) bool {
										return v[i].Value > v[j].Value
									})
									//
									top_info := v[0]
									if top_info != nil {
										top_gt := top_info.Gtype
										gname, ok := recordList[int(top_gt)]
										if ok {
											temp_dt.Gtype = int64(top_gt)
											temp_dt.GameName = gname
										}
									}
								}
							}
						}
					}
				}
				d_info = append(d_info, *temp_dt)
			}

		}
		if len(d_info) < num {
			// 少多少个次数,补多少个
			temp_count := num - len(d_info)
			for i := 0; i < temp_count; i++ {
				temp_dt := new(entity.PlayerRechargeDetails)
				d_info = append(d_info, *temp_dt)
			}
		}
		new_info.Details = &d_info
		list = append(list, *new_info)
	}

	return list, err
}

/*
新手库存
*/
func (this *statisticsService) GetNoviceStockList(m bson.M) ([]entity.NewbieStock, error) {
	var list []entity.NewbieStock
	err := NewbieStocks.
		Find(bson.M{}).
		All(&list)
	list = this.chipList27(list)
	return list, err
}

func (this *statisticsService) chipList27(list []entity.NewbieStock) []entity.NewbieStock {
	for k, v := range list {
		switch v.Id {
		case 1:
			v.GName = "TP"
		case 2:
			v.GName = "DRAGON TIGER"
		case 3:
			v.GName = "7UPDOWN"
		case 4:
			v.GName = "RUMMY"
		case 5:
			v.GName = "AK47"
		case 6:
			v.GName = "JOKER"
		case 7:
			v.GName = "CRASH"
		case 8:
			v.GName = "ANDARBAHAR"
		case 9:
			v.GName = "彩票"
		case 10:
			v.GName = "飞机"
		case 11:
			v.GName = "红黑大战"
		case 12:
			v.GName = "RUMMY双人"
		case 13:
			v.GName = "TP2"
		case 14:
			v.GName = "MINES"
		}
		v.FCashStock = ComputeFloat(v.CashStock, 100)
		v.RoundsAvg = ComputeFloat(v.PlayTimes, v.PartIn)
		v.GTimeAvg = ComputeFloat(v.GameTime, v.PartIn)
		list[k] = v
	}
	return list
}

/*
首充分析 FirstCharges
*/
func (this *statisticsService) GetFirstChargeList(page, pageSize int, m bson.M, isChannel bool) ([]entity.FirstChargeAnalysis, error) {
	var list []entity.FirstChargeAnalysis

	if pageSize == -1 {
		pageSize = 100000
	}
	// skipNum, _ := parsePageAndSort(page, pageSize, "date", false)
	// pipeline := []bson.M{}
	// if isChannel {
	// 	// 渠道
	// 	pipeline = []bson.M{
	// 		{"$match": m},
	// 		{
	// 			"$group": bson.M{
	// 				"_id": bson.M{
	// 					"date":     "$date",
	// 					"channel":  "$channel",
	// 					"channel1": "$channel1",
	// 				},
	// 				"reg_number":   bson.M{"$sum": "$reg_number"},
	// 				"first_number": bson.M{"$sum": "$first_number"},
	// 				"tp_charge":    bson.M{"$sum": "$tp_charge"},
	// 				"rm_charge":    bson.M{"$sum": "$rm_charge"},
	// 				"lhd_charge":   bson.M{"$sum": "$lhd_charge"},
	// 				"up_charge":    bson.M{"$sum": "$up_charge"},
	// 				"ak47_charge":  bson.M{"$sum": "$ak47_charge"},
	// 				"joker_charge": bson.M{"$sum": "$joker_charge"},
	// 				"crash_charge": bson.M{"$sum": "$crash_charge"},
	// 				"cp_charge":    bson.M{"$sum": "$cp_charge"},
	// 				"ab_charge":    bson.M{"$sum": "$ab_charge"},
	// 				"fj_charge":    bson.M{"$sum": "$fj_charge"},
	// 				"qt_number":    bson.M{"$sum": "$qt_number"},
	// 			},
	// 		},
	// 		{
	// 			"$project": bson.M{
	// 				"_id":          "$_id.date",
	// 				"date":         "$_id.date",
	// 				"channel":      "$_id.channel",
	// 				"channel1":     "$_id.channel1",
	// 				"reg_number":   "$reg_number",
	// 				"first_number": "$first_number",
	// 				"tp_charge":    "$tp_charge",
	// 				"rm_charge":    "$rm_charge",
	// 				"lhd_charge":   "$lhd_charge",
	// 				"up_charge":    "$up_charge",
	// 				"ak47_charge":  "$ak47_charge",
	// 				"joker_charge": "$joker_charge",
	// 				"crash_charge": "$crash_charge",
	// 				"cp_charge":    "$cp_charge",
	// 				"ab_charge":    "$ab_charge",
	// 				"fj_charge":    "$fj_charge",
	// 				"qt_number":    "$qt_number",
	// 			},
	// 		},
	// 		{"$sort": bson.M{"date": -1}},
	// 		{"$skip": skipNum},
	// 		{"$limit": pageSize},
	// 	}
	// } else {
	// 	// 全部
	// 	pipeline = []bson.M{
	// 		{"$match": m},
	// 		{
	// 			"$group": bson.M{
	// 				"_id": bson.M{
	// 					"date": "$date",
	// 				},
	// 				"reg_number":   bson.M{"$sum": "$reg_number"},
	// 				"first_number": bson.M{"$sum": "$first_number"},
	// 				"tp_charge":    bson.M{"$push": "$tp_charge"},
	// 				"rm_charge":    bson.M{"$push": "$rm_charge"},
	// 				"lhd_charge":   bson.M{"$push": "$lhd_charge"},
	// 				"up_charge":    bson.M{"$push": "$up_charge"},
	// 				"ak47_charge":  bson.M{"$push": "$ak47_charge"},
	// 				"joker_charge": bson.M{"$push": "$joker_charge"},
	// 				"crash_charge": bson.M{"$push": "$crash_charge"},
	// 				"cp_charge":    bson.M{"$push": "$cp_charge"},
	// 				"ab_charge":    bson.M{"$push": "$ab_charge"},
	// 				"fj_charge":    bson.M{"$push": "$fj_charge"},
	// 				"qt_number":    bson.M{"$sum": "$qt_number"},
	// 			},
	// 		},
	// 		{
	// 			"$project": bson.M{
	// 				"_id":          "$_id.date",
	// 				"date":         "$_id.date",
	// 				"reg_number":   "$reg_number",
	// 				"first_number": "$first_number",
	// 				"tp_charge":    "$tp_charge",
	// 				"rm_charge":    "$rm_charge",
	// 				"lhd_charge":   "$lhd_charge",
	// 				"up_charge":    "$up_charge",
	// 				"ak47_charge":  "$ak47_charge",
	// 				"joker_charge": "$joker_charge",
	// 				"crash_charge": "$crash_charge",
	// 				"cp_charge":    "$cp_charge",
	// 				"ab_charge":    "$ab_charge",
	// 				"fj_charge":    "$fj_charge",
	// 				"qt_number":    "$qt_number",
	// 			},
	// 		},
	// 		{"$sort": bson.M{"date": -1}},
	// 		{"$skip": skipNum},
	// 		{"$limit": pageSize},
	// 	}
	// }
	// err := FirstCharges.Find(m).All(&temp)
	var err error
	if isChannel {
		// 渠道
		skipNum, sortFieldR := parsePageAndSort(page, pageSize, "date", false)
		err = FirstCharges.Find(m).Sort(sortFieldR).
			Skip(skipNum).
			Limit(pageSize).
			All(&list)
	} else {
		// 全部
		var temp []entity.FirstChargeAnalysis
		err = FirstCharges.Find(m).Sort("-date").All(&temp)
		map_first := make(map[int64]*entity.FirstChargeAnalysis, 0)
		for _, item := range temp {
			stat, ok := map_first[item.Date]
			if !ok {
				stat = &entity.FirstChargeAnalysis{
					Date: item.Date,
				}
			}
			stat.RegNumber += item.RegNumber
			stat.FirstNumber += item.FirstNumber
			stat.QTNumber += item.QTNumber
			if len(item.TPCharge) > 0 {
				tp_map := make(map[int64]entity.GameCharge, 0)
				if len(stat.TPCharge) > 0 {
					for _, r := range stat.TPCharge {
						tp_map[r.Amount] = r
					}
				}
				for _, r := range item.TPCharge {
					char, ok := tp_map[r.Amount]
					if !ok {
						char = entity.GameCharge{
							GType:  1,
							Amount: r.Amount,
						}
					}
					char.Number += r.Number
					tp_map[r.Amount] = char
				}
				tplist := make([]entity.GameCharge, 0)
				for _, p := range tp_map {
					tplist = append(tplist, p)
				}
				stat.TPCharge = tplist
			}
			//lhd
			if len(item.LHDCharge) > 0 {
				tp_map := make(map[int64]entity.GameCharge, 0)
				if len(stat.LHDCharge) > 0 {
					for _, r := range stat.LHDCharge {
						tp_map[r.Amount] = r
					}
				}
				for _, r := range item.LHDCharge {
					char, ok := tp_map[r.Amount]
					if !ok {
						char = entity.GameCharge{
							GType:  2,
							Amount: r.Amount,
						}
					}
					char.Number += r.Number
					tp_map[r.Amount] = char
				}
				tplist := make([]entity.GameCharge, 0)
				for _, p := range tp_map {
					tplist = append(tplist, p)
				}
				stat.LHDCharge = tplist
			}
			//7UP
			if len(item.UPCharge) > 0 {
				tp_map := make(map[int64]entity.GameCharge, 0)
				if len(stat.UPCharge) > 0 {
					for _, r := range stat.UPCharge {
						tp_map[r.Amount] = r
					}
				}
				for _, r := range item.UPCharge {
					char, ok := tp_map[r.Amount]
					if !ok {
						char = entity.GameCharge{
							GType:  3,
							Amount: r.Amount,
						}
					}
					char.Number += r.Number
					tp_map[r.Amount] = char
				}
				tplist := make([]entity.GameCharge, 0)
				for _, p := range tp_map {
					tplist = append(tplist, p)
				}
				stat.UPCharge = tplist
			}
			// RM
			if len(item.RMCharge) > 0 {
				tp_map := make(map[int64]entity.GameCharge, 0)
				if len(stat.RMCharge) > 0 {
					for _, r := range stat.RMCharge {
						tp_map[r.Amount] = r
					}
				}
				for _, r := range item.RMCharge {
					char, ok := tp_map[r.Amount]
					if !ok {
						char = entity.GameCharge{
							GType:  4,
							Amount: r.Amount,
						}
					}
					char.Number += r.Number
					tp_map[r.Amount] = char
				}
				tplist := make([]entity.GameCharge, 0)
				for _, p := range tp_map {
					tplist = append(tplist, p)
				}
				stat.RMCharge = tplist
			}
			// ak
			if len(item.AK47Charge) > 0 {
				tp_map := make(map[int64]entity.GameCharge, 0)
				if len(stat.AK47Charge) > 0 {
					for _, r := range stat.AK47Charge {
						tp_map[r.Amount] = r
					}
				}
				for _, r := range item.AK47Charge {
					char, ok := tp_map[r.Amount]
					if !ok {
						char = entity.GameCharge{
							GType:  5,
							Amount: r.Amount,
						}
					}
					char.Number += r.Number
					tp_map[r.Amount] = char
				}
				tplist := make([]entity.GameCharge, 0)
				for _, p := range tp_map {
					tplist = append(tplist, p)
				}
				stat.AK47Charge = tplist
			}
			// joker
			if len(item.JokerCharge) > 0 {
				tp_map := make(map[int64]entity.GameCharge, 0)
				if len(stat.JokerCharge) > 0 {
					for _, r := range stat.JokerCharge {
						tp_map[r.Amount] = r
					}
				}
				for _, r := range item.JokerCharge {
					char, ok := tp_map[r.Amount]
					if !ok {
						char = entity.GameCharge{
							GType:  6,
							Amount: r.Amount,
						}
					}
					char.Number += r.Number
					tp_map[r.Amount] = char
				}
				tplist := make([]entity.GameCharge, 0)
				for _, p := range tp_map {
					tplist = append(tplist, p)
				}
				stat.JokerCharge = tplist
			}
			// crash
			if len(item.CrashCharge) > 0 {
				tp_map := make(map[int64]entity.GameCharge, 0)
				if len(stat.CrashCharge) > 0 {
					for _, r := range stat.CrashCharge {
						tp_map[r.Amount] = r
					}
				}
				for _, r := range item.CrashCharge {
					char, ok := tp_map[r.Amount]
					if !ok {
						char = entity.GameCharge{
							GType:  7,
							Amount: r.Amount,
						}
					}
					char.Number += r.Number
					tp_map[r.Amount] = char
				}
				tplist := make([]entity.GameCharge, 0)
				for _, p := range tp_map {
					tplist = append(tplist, p)
				}
				stat.CrashCharge = tplist
			}
			// AB
			if len(item.ABCharge) > 0 {
				tp_map := make(map[int64]entity.GameCharge, 0)
				if len(stat.ABCharge) > 0 {
					for _, r := range stat.ABCharge {
						tp_map[r.Amount] = r
					}
				}
				for _, r := range item.ABCharge {
					char, ok := tp_map[r.Amount]
					if !ok {
						char = entity.GameCharge{
							GType:  8,
							Amount: r.Amount,
						}
					}
					char.Number += r.Number
					tp_map[r.Amount] = char
				}
				tplist := make([]entity.GameCharge, 0)
				for _, p := range tp_map {
					tplist = append(tplist, p)
				}
				stat.ABCharge = tplist
			}
			// CP
			if len(item.CPCharge) > 0 {
				tp_map := make(map[int64]entity.GameCharge, 0)
				if len(stat.CPCharge) > 0 {
					for _, r := range stat.CPCharge {
						tp_map[r.Amount] = r
					}
				}
				for _, r := range item.CPCharge {
					char, ok := tp_map[r.Amount]
					if !ok {
						char = entity.GameCharge{
							GType:  9,
							Amount: r.Amount,
						}
					}
					char.Number += r.Number
					tp_map[r.Amount] = char
				}
				tplist := make([]entity.GameCharge, 0)
				for _, p := range tp_map {
					tplist = append(tplist, p)
				}
				stat.CPCharge = tplist
			}
			// 飞机
			if len(item.FJCharge) > 0 {
				tp_map := make(map[int64]entity.GameCharge, 0)
				if len(stat.FJCharge) > 0 {
					for _, r := range stat.FJCharge {
						tp_map[r.Amount] = r
					}
				}
				for _, r := range item.FJCharge {
					char, ok := tp_map[r.Amount]
					if !ok {
						char = entity.GameCharge{
							GType:  10,
							Amount: r.Amount,
						}
					}
					char.Number += r.Number
					tp_map[r.Amount] = char
				}
				tplist := make([]entity.GameCharge, 0)
				for _, p := range tp_map {
					tplist = append(tplist, p)
				}
				stat.FJCharge = tplist
			}

			// Rummy双人
			if len(item.RMTwoCharge) > 0 {
				tp_map := make(map[int64]entity.GameCharge, 0)
				if len(stat.RMTwoCharge) > 0 {
					for _, r := range stat.RMTwoCharge {
						tp_map[r.Amount] = r
					}
				}
				for _, r := range item.RMTwoCharge {
					char, ok := tp_map[r.Amount]
					if !ok {
						char = entity.GameCharge{
							GType:  10,
							Amount: r.Amount,
						}
					}
					char.Number += r.Number
					tp_map[r.Amount] = char
				}
				tplist := make([]entity.GameCharge, 0)
				for _, p := range tp_map {
					tplist = append(tplist, p)
				}
				stat.RMTwoCharge = tplist
			}
			map_first[item.Date] = stat
		}
		for _, b := range map_first {
			list = append(list, *b)
		}
		// 按照日期排序
		sort.Slice(list, func(i, j int) bool {
			return list[i].Date > list[j].Date
		})

	}
	list = this.chipList28(list)

	return list, err
}

func (this *statisticsService) chipList28(list []entity.FirstChargeAnalysis) []entity.FirstChargeAnalysis {
	var shoplist []entity.Shop
	Shops.Find(bson.M{}).All(&shoplist)
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		tplist := make([]entity.GameCharge, 0)
		lhdlist := make([]entity.GameCharge, 0)
		uplist := make([]entity.GameCharge, 0)
		rmlist := make([]entity.GameCharge, 0)
		ak47list := make([]entity.GameCharge, 0)
		jokerlist := make([]entity.GameCharge, 0)
		crashlist := make([]entity.GameCharge, 0)
		cplist := make([]entity.GameCharge, 0)
		ablist := make([]entity.GameCharge, 0)
		fjlist := make([]entity.GameCharge, 0)
		rmTwolist := make([]entity.GameCharge, 0)
		for _, o := range shoplist {
			ShowAmount := ComputeFloat(o.Number, 100)
			new_tp := &entity.GameCharge{
				GType:  1,
				Amount: int64(ShowAmount),
			}
			new_lhd := &entity.GameCharge{
				GType:  2,
				Amount: int64(ShowAmount),
			}
			new_up := &entity.GameCharge{
				GType:  3,
				Amount: int64(ShowAmount),
			}
			new_rm := &entity.GameCharge{
				GType:  4,
				Amount: int64(ShowAmount),
			}
			new_ak47 := &entity.GameCharge{
				GType:  5,
				Amount: int64(ShowAmount),
			}
			new_joker := &entity.GameCharge{
				GType:  6,
				Amount: int64(ShowAmount),
			}
			new_crash := &entity.GameCharge{
				GType:  7,
				Amount: int64(ShowAmount),
			}
			new_cp := &entity.GameCharge{
				GType:  9,
				Amount: int64(ShowAmount),
			}
			new_ab := &entity.GameCharge{
				GType:  8,
				Amount: int64(ShowAmount),
			}
			new_fj := &entity.GameCharge{
				GType:  10,
				Amount: int64(ShowAmount),
			}
			new_rm_two := &entity.GameCharge{
				GType:  12,
				Amount: int64(ShowAmount),
			}
			for _, r := range v.TPCharge {
				if o.Number == r.Amount {
					new_tp.Number = r.Number
					new_tp.Rate = ComputeFloat(r.Number, v.FirstNumber) * 100.0
				}
			}
			for _, r := range v.LHDCharge {
				if o.Number == r.Amount {
					new_lhd.Number = r.Number
					new_lhd.Rate = ComputeFloat(r.Number, v.FirstNumber) * 100.0
				}
			}
			for _, r := range v.UPCharge {
				if o.Number == r.Amount {
					new_up.Number = r.Number
					new_up.Rate = ComputeFloat(r.Number, v.FirstNumber) * 100.0
				}
			}
			for _, r := range v.RMCharge {
				if o.Number == r.Amount {
					new_rm.Number = r.Number
					new_rm.Rate = ComputeFloat(r.Number, v.FirstNumber) * 100.0
				}
			}
			for _, r := range v.AK47Charge {
				if o.Number == r.Amount {
					new_ak47.Number = r.Number
					new_ak47.Rate = ComputeFloat(r.Number, v.FirstNumber) * 100.0
				}
			}
			for _, r := range v.JokerCharge {
				if o.Number == r.Amount {
					new_joker.Number = r.Number
					new_joker.Rate = ComputeFloat(r.Number, v.FirstNumber) * 100.0
				}
			}
			for _, r := range v.CrashCharge {
				if o.Number == r.Amount {
					new_crash.Number = r.Number
					new_crash.Rate = ComputeFloat(r.Number, v.FirstNumber) * 100.0
				}
			}
			for _, r := range v.CPCharge {
				if o.Number == r.Amount {
					new_cp.Number = r.Number
					new_cp.Rate = ComputeFloat(r.Number, v.FirstNumber) * 100.0
				}
			}
			for _, r := range v.ABCharge {
				if o.Number == r.Amount {
					new_ab.Number = r.Number
					new_ab.Rate = ComputeFloat(r.Number, v.FirstNumber) * 100.0
				}
			}
			for _, r := range v.FJCharge {
				if o.Number == r.Amount {
					new_fj.Number = r.Number
					new_fj.Rate = ComputeFloat(r.Number, v.FirstNumber) * 100.0
				}
			}
			for _, r := range v.RMTwoCharge {
				if o.Number == r.Amount {
					new_rm_two.Number = r.Number
					new_rm_two.Rate = ComputeFloat(r.Number, v.FirstNumber) * 100.0
				}
			}
			tplist = append(tplist, *new_tp)
			lhdlist = append(lhdlist, *new_lhd)
			uplist = append(uplist, *new_up)
			rmlist = append(rmlist, *new_rm)
			ak47list = append(ak47list, *new_ak47)
			jokerlist = append(jokerlist, *new_joker)
			crashlist = append(crashlist, *new_crash)
			cplist = append(cplist, *new_cp)
			ablist = append(ablist, *new_ab)
			fjlist = append(fjlist, *new_fj)
			rmTwolist = append(rmTwolist, *new_rm_two)
		}
		v.TPCharge = tplist
		v.LHDCharge = lhdlist
		v.UPCharge = uplist
		v.RMCharge = rmlist
		v.AK47Charge = ak47list
		v.JokerCharge = jokerlist
		v.CrashCharge = crashlist
		v.ABCharge = ablist
		v.CPCharge = cplist
		v.FJCharge = fjlist
		v.RMTwoCharge = rmTwolist
		v.QTRate = ComputeFloat(v.QTNumber, v.FirstNumber) * 100.0
		list[k] = v
	}
	return list
}

// 查询首充分析条数
func (this *statisticsService) GetFirstChargeTotal(m bson.M, isChannel bool) (int64, error) {
	pipeline := []bson.M{}
	if isChannel {
		pipeline = []bson.M{
			{
				"$match": m,
			},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":     "$date",
						"channel":  "$channel",
						"channel1": "$channel1",
					},
				},
			},
		}
	} else {
		pipeline = []bson.M{
			{
				"$match": m,
			},
			{
				"$group": bson.M{
					"_id": "$date",
				},
			},
		}
	}

	result := []bson.M{}
	pipe := FirstCharges.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("GetFirstChargeTotal fail err: ", err)
	}
	return int64(len(result)), nil
}

// 玩家打码分层 筛选参数
func s0ColUserFilter(registAreas, utypes []int32, packageIds []string) (
	where_u_sql string, where_u_args []any,
) {
	if len(registAreas) > 0 {
		where_u_sql += " and s0.regist_area in ?"
		where_u_args = append(where_u_args, registAreas)
	}
	if len(packageIds) > 0 {
		where_u_sql += " and s0.ad__bundle_id IN ?"
		where_u_args = append(where_u_args, packageIds)
	}

	var utypeFilters []string
	if len(utypes) > 0 {
		for _, utype := range utypes {
			// state 1.新手 2.正常 3.平民 4.泡沫
			state, sMoney, eMoney := 2, -1, -1
			switch utype {
			case 0: // 新手
				state = 1
				sMoney, eMoney = 0, 0
			case 1: // 平民
				state = -1
				sMoney, eMoney = 0, 0
			case 2: // 普充
				sMoney, eMoney = 1, 99999
			case 3: // 小R
				sMoney, eMoney = 100000, 499999
			case 4: // 中R
				sMoney, eMoney = 500000, 999999
			case 5: // 大R
				sMoney, eMoney = 1000000, 19999999
			case 6: // 超大R
				sMoney, eMoney = 20000000, -1
			}

			var utypeQs []string
			if state != -1 {
				utypeQs = append(utypeQs, fmt.Sprintf("s0.state = %d", state))
			}
			if sMoney != -1 {
				utypeQs = append(utypeQs, fmt.Sprintf("s0.money >= %d", sMoney))
			}
			if eMoney != -1 {
				utypeQs = append(utypeQs, fmt.Sprintf("s0.money <= %d", eMoney))
			}
			if len(utypeQs) > 0 {
				utypeFilters = append(utypeFilters, fmt.Sprintf("(%s)", strings.Join(utypeQs, " AND ")))
			}
		}
	}
	if len(utypeFilters) > 0 {
		utypeQFmt := " AND (%s)"
		if len(utypeFilters) == 1 {
			utypeQFmt = " AND %s"
		}
		where_u_sql += fmt.Sprintf(utypeQFmt, strings.Join(utypeFilters, " OR "))
	}
	return
}

// 玩家打码分层
func (c *statisticsService) PlayerBetsLevel(stime, etime time.Time, registAreas, utypes []int32, packageIds []string) (stats []*entity.PlayerBetsLevel, err error) {
	where_u_sql, where_u_args := s0ColUserFilter(registAreas, utypes, packageIds)

	var bets_datas []map[string]any
	var bets_args []any
	bets_args = append(bets_args, stime.Unix(), etime.Unix(), stime.Unix(), etime.Unix())
	bets_args = append(bets_args, where_u_args...)
	bets_args = append(bets_args, stime.Unix(), etime.Unix(), stime.Unix(), etime.Unix())
	bets_args = append(bets_args, where_u_args...)
	err = ck.Select(&bets_datas, fmt.Sprintf(`
		select t0.sdate sdate, sum(t0.bets) bets,
			SUM(case when t0.ranking < t1.top05_rank then 1 else 0 end) top05_users,
			SUM(case when t0.ranking < t1.top1_rank then 1 else 0 end) top1_users,
			SUM(case when t0.ranking < t1.top3_rank then 1 else 0 end) top3_users,
			SUM(case when t0.ranking < t1.top5_rank then 1 else 0 end) top5_users,
			SUM(case when t0.ranking < t1.top10_rank then 1 else 0 end) top10_users,
			SUM(case when t0.ranking < t1.top05_rank then t0.bets else 0 end) top05_bets,
			SUM(case when t0.ranking < t1.top1_rank then t0.bets else 0 end) top1_bets,
			SUM(case when t0.ranking < t1.top3_rank then t0.bets else 0 end) top3_bets,
			SUM(case when t0.ranking < t1.top5_rank then t0.bets else 0 end) top5_bets,
			SUM(case when t0.ranking < t1.top10_rank then t0.bets else 0 end) top10_bets
		from (
			select sdate, userid, sum(bets) bets,
				rank() over (PARTITION BY sdate ORDER BY bets desc) as ranking
			from game.col_user s0 final
			join (
				select toYYYYMMDD(toDateTime(begin_time)) sdate, userid, sum(bet_amount) bets 
				from game.col_detail final where robot = 0 and begin_time >= ? and begin_time < ? and bet_amount != 0
				group by sdate, userid
				
				union all
				
				select toYYYYMMDD(toDateTime(ctime)) sdate, user_id userid, sum(amount) bets 
				from game.col_nsq_log_external_bet final where ctime > ? and ctime < ? and amount != 0
				group by sdate, user_id
			) s1 on s1.userid = s0.userid
			where 1 = 1 %s
			group by sdate, userid
			order by sdate, bets desc
		) t0 join (
			select sdate, count(distinct userid) bet_users,
				toUInt64(ceil(bet_users * 0.005)) top05_rank,
				toUInt64(ceil(bet_users * 0.01)) top1_rank,
				toUInt64(ceil(bet_users * 0.03)) top3_rank,
				toUInt64(ceil(bet_users * 0.05)) top5_rank,
				toUInt64(ceil(bet_users * 0.1)) top10_rank
			from game.col_user s0 final
			join (
				select toYYYYMMDD(toDateTime(begin_time)) sdate, userid
				from game.col_detail final where robot = 0 and begin_time >= ? and begin_time < ? and bet_amount != 0
				group by sdate, userid
				
				union all
				
				select toYYYYMMDD(toDateTime(ctime)) sdate, user_id userid
				from game.col_nsq_log_external_bet final where ctime > ? and ctime < ? and amount != 0
				group by sdate, user_id
			) s1 on s1.userid = s0.userid
			where 1 = 1 %s
			group by sdate
		) t1 on t0.sdate = t1.sdate
		group by t0.sdate
		order by t0.sdate desc
	`, where_u_sql, where_u_sql), bets_args...)
	if err != nil {
		return
	}
	for _, data := range bets_datas {
		sdate := data["sdate"].(uint32)
		datestr := fmt.Sprint(sdate)
		fdate := datestr[0:4] + "-" + datestr[4:6] + "-" + datestr[6:8]

		stat := &entity.PlayerBetsLevel{SDate: fdate}
		stats = append(stats, stat)

		bets := utils.ToInt64(data["bets"])
		top05_users := utils.ToInt64(data["top05_users"])
		top1_users := utils.ToInt64(data["top1_users"])
		top3_users := utils.ToInt64(data["top3_users"])
		top5_users := utils.ToInt64(data["top5_users"])
		top10_users := utils.ToInt64(data["top10_users"])
		top05_bets := utils.ToInt64(data["top05_bets"])
		top1_bets := utils.ToInt64(data["top1_bets"])
		top3_bets := utils.ToInt64(data["top3_bets"])
		top5_bets := utils.ToInt64(data["top5_bets"])
		top10_bets := utils.ToInt64(data["top10_bets"])

		stat.Bets0 = bets
		stat.Top05Users = top05_users
		stat.Top05UsersBets0 = top05_bets
		stat.Top1Users = top1_users
		stat.Top1UsersBets0 = top1_bets
		stat.Top3Users = top3_users
		stat.Top3UsersBets0 = top3_bets
		stat.Top5Users = top5_users
		stat.Top5UsersBets0 = top5_bets
		stat.Top10Users = top10_users
		stat.Top10UsersBets0 = top10_bets
	}

	summary := &entity.PlayerBetsLevel{
		SDate: fmt.Sprintf("%s~%s汇总", stime.Format(utils.FORMAT_DATE), etime.Format(utils.FORMAT_DATE)),
	}
	stats = append([]*entity.PlayerBetsLevel{summary}, stats...)
	for _, stat := range stats {
		summary.Bets0 += stat.Bets0
		summary.Top05Users += stat.Top05Users
		summary.Top05UsersBets0 += stat.Top05UsersBets0
		summary.Top1Users += stat.Top1Users
		summary.Top1UsersBets0 += stat.Top1UsersBets0
		summary.Top3Users += stat.Top3Users
		summary.Top3UsersBets0 += stat.Top3UsersBets0
		summary.Top5Users += stat.Top5Users
		summary.Top5UsersBets0 += stat.Top5UsersBets0
		summary.Top10Users += stat.Top10Users
		summary.Top10UsersBets0 += stat.Top10UsersBets0
	}

	for _, stat := range stats {
		stat.Bets = fmt.Sprintf("%.2f", Chip2Float(stat.Bets0))
		stat.Top05UsersBets = fmt.Sprintf("%.2f", Chip2Float(stat.Top05UsersBets0))
		stat.Top1UsersBets = fmt.Sprintf("%.2f", Chip2Float(stat.Top1UsersBets0))
		stat.Top3UsersBets = fmt.Sprintf("%.2f", Chip2Float(stat.Top3UsersBets0))
		stat.Top5UsersBets = fmt.Sprintf("%.2f", Chip2Float(stat.Top5UsersBets0))
		stat.Top10UsersBets = fmt.Sprintf("%.2f", Chip2Float(stat.Top10UsersBets0))

		stat.Top05UsersBetsAvg = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(stat.Top05UsersBets0, stat.Top05Users)))
		stat.Top05UsersBetsRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.Top05UsersBets0, stat.Bets0)*100)
		stat.Top1UsersBetsAvg = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(stat.Top1UsersBets0, stat.Top1Users)))
		stat.Top1UsersBetsRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.Top1UsersBets0, stat.Bets0)*100)
		stat.Top3UsersBetsAvg = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(stat.Top3UsersBets0, stat.Top3Users)))
		stat.Top3UsersBetsRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.Top3UsersBets0, stat.Bets0)*100)
		stat.Top5UsersBetsAvg = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(stat.Top5UsersBets0, stat.Top5Users)))
		stat.Top5UsersBetsRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.Top5UsersBets0, stat.Bets0)*100)
		stat.Top10UsersBetsAvg = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(stat.Top10UsersBets0, stat.Top10Users)))
		stat.Top10UsersBetsRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.Top10UsersBets0, stat.Bets0)*100)
	}
	return
}

// 玩家打码分层
func (c *statisticsService) PlayerBetsUsers(calcTotal bool, page, pageSize int, stime, etime time.Time, registAreas, utypes []int32, packageIds []string) (total int64, stats []*entity.PlayerBetsUser, err error) {
	where_u_sql, where_u_args := s0ColUserFilter(registAreas, utypes, packageIds)

	var bet_user_args = []any{stime.Unix(), etime.Unix(), stime.Unix(), etime.Unix(), stime.Unix(), etime.Unix()}
	bet_user_args = append(bet_user_args, where_u_args...)

	if calcTotal {
		err = ck.Select(&total, fmt.Sprintf(`
			select count(*) total from (
				select s1.sdate, s1.userid userid
				from game.col_user s0 final
				join (
					select userid, gtype, sdate, sum(rounds) rounds0, sum(bets) bets0, sum(rebates) rebates0
					from (
						select userid, gtype, toYYYYMMDD(toDateTime(begin_time)) sdate, count(*) rounds, sum(bet_amount) bets, sum(bet_amount+score) rebates
						from game.col_detail final where robot = 0 and begin_time >= ? and begin_time < ? and bet_amount != 0
						group by userid, gtype, sdate
						
						union all
						
						select t1.userid, t1.gtype, t1.sdate, SUM(t1.rounds) rounds, sum(t1.amounts) bets, sum(t2.amounts) rebates 
						from (
							select user_id userid, game_id gtype, toYYYYMMDD(toDateTime(ctime)) sdate, count(distinct round_id) rounds, sum(amount) amounts
							from game.col_nsq_log_external_bet final where ctime > ? and ctime < ? and amount != 0
							group by user_id, game_id, sdate
						) t1 left join (
							select user_id userid, game_id gtype, toYYYYMMDD(toDateTime(ctime)) sdate, sum(amount) amounts
							from game.col_nsq_log_external_reward final where ctime > ? and ctime < ? and amount != 0
							group by user_id, game_id, sdate
						) t2 on t1.userid = t2.userid and t1.gtype = t2.gtype and t1.sdate = t2.sdate
						group by t1.userid, t1.gtype, t1.sdate
						
					) s1
					group by userid, gtype, sdate
				) s1 on s1.userid = s0.userid
				where 1 = 1 %s
				group by s1.sdate, s1.userid
			) t0
		`, where_u_sql), bet_user_args...)
		if err != nil {
			return
		}
		if total == 0 {
			return
		}
	}

	offset, limit := PageCalc(page, pageSize)
	bet_user_args1 := append(bet_user_args, offset, limit)
	var bet_user_datas []map[string]any
	err = ck.Select(&bet_user_datas, fmt.Sprintf(`
		select s1.sdate, s1.userid userid, sum(s1.bets0) bets_sum, sum(s1.rebates0) rebates_sum,
			rank() over (PARTITION BY s1.sdate ORDER BY bets_sum desc) as ranking,
			groupArray((s1.ranking, s1.gtype, s1.rounds0, s1.bets0)) top_bet_games,
			s0.ctime, s0.login_time, s0.state, s0.money, s0.cash_out, s0.vip_lv
		from game.col_user s0 final
		join (
			select userid, gtype, sdate, sum(rounds) rounds0, sum(bets) bets0, sum(rebates) rebates0,
				rank() over (PARTITION BY userid, sdate ORDER BY bets0 desc) as ranking
			from (
				select userid, gtype, toYYYYMMDD(toDateTime(begin_time)) sdate, count(*) rounds, sum(bet_amount) bets, sum(bet_amount+score) rebates
				from game.col_detail final where robot = 0 and begin_time >= ? and begin_time < ? and bet_amount != 0
				group by userid, gtype, sdate
				
				union all
				
				select t1.userid, t1.gtype, t1.sdate, SUM(t1.rounds) rounds, sum(t1.amounts) bets, sum(t2.amounts) rebates 
				from (
					select user_id userid, game_id gtype, toYYYYMMDD(toDateTime(ctime)) sdate, count(distinct round_id) rounds, sum(amount) amounts
					from game.col_nsq_log_external_bet final where ctime > ? and ctime < ? and amount != 0
					group by user_id, game_id, sdate
				) t1 left join (
					select user_id userid, game_id gtype, toYYYYMMDD(toDateTime(ctime)) sdate, sum(amount) amounts
					from game.col_nsq_log_external_reward final where ctime > ? and ctime < ? and amount != 0
					group by user_id, game_id, sdate
				) t2 on t1.userid = t2.userid and t1.gtype = t2.gtype and t1.sdate = t2.sdate
				group by t1.userid, t1.gtype, t1.sdate
			) s1
			group by userid, gtype, sdate
		) s1 on s1.userid = s0.userid
		where 1 = 1 %s
		group by s1.sdate, s1.userid, s0.ctime, s0.login_time, s0.state, s0.money, s0.cash_out, s0.vip_lv
		order by s1.sdate desc, bets_sum desc 
		limit ?, ?
	`, where_u_sql), bet_user_args1...)
	if err != nil {
		return
	}

	for _, data := range bet_user_datas {
		sdate := data["sdate"].(uint32)
		userid := data["userid"].(string)
		bets_sum := utils.ToInt64(data["bets_sum"])
		rebates_sum := utils.ToInt64(data["rebates_sum"])
		ranking := utils.ToInt64(data["ranking"])
		state := utils.ToInt64(data["state"])
		money := utils.ToInt64(data["money"])
		cash_out := utils.ToInt64(data["cash_out"])
		vip_lv := utils.ToInt64(data["vip_lv"])
		ctime := data["ctime"].(time.Time)
		login_time := data["login_time"].(time.Time)
		top_bet_games := data["top_bet_games"].([][]interface{})

		datestr := fmt.Sprint(sdate)
		fdate := datestr[0:4] + "-" + datestr[4:6] + "-" + datestr[6:8]

		_, chargeTypeName := GetChargeType(money, int(state))
		stat := &entity.PlayerBetsUser{
			SDate:            fdate,
			Rank:             fmt.Sprint(ranking),
			Userid:           userid,
			VipLv:            fmt.Sprint(vip_lv),
			UserType:         chargeTypeName,
			Pays:             fmt.Sprintf("%.2f", Chip2Float(money)),
			Withdraws:        fmt.Sprintf("%.2f", Chip2Float(cash_out)),
			PaysSubWithdraws: fmt.Sprintf("%.2f", Chip2Float(money-cash_out)),
			Bets:             fmt.Sprintf("%.2f", Chip2Float(bets_sum)),
			RabateRate:       fmt.Sprintf("%.2f%%", ComputeFloat(rebates_sum, bets_sum)*100),
			Wins:             fmt.Sprintf("%.2f", Chip2Float(rebates_sum-bets_sum)),
		}
		stats = append(stats, stat)

		now := time.Now().In(location)
		stat.Ctime = ctime.Format(utils.FORMAT)
		stat.LastLoginTime = login_time.Format(utils.FORMAT)
		stat.LiveDays = fmt.Sprint(int64(math.Ceil(login_time.Sub(ctime).Hours() / 24)))
		stat.LoseDays = fmt.Sprint(int64(math.Floor(now.Sub(login_time).Hours() / 24)))

		for _, top_bet := range top_bet_games {
			// (s1.ranking, s1.gtype, s1.rounds0, s1.bets0)
			rank := utils.ToInt64(top_bet[0])
			gtype := int(utils.ToInt64(top_bet[1]))
			rounds := utils.ToInt64(top_bet[2])
			bets := utils.ToInt64(top_bet[3])
			if rank > 3 {
				continue
			}
			switch rank {
			case 1:
				stat.Top1Game = GtypeNameMap[gtype]
				stat.Top1GameBets = fmt.Sprintf("%.2f", Chip2Float(bets))
				stat.Top1GameRounds = fmt.Sprint(rounds)
				stat.Top1GameBetsAvg = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(bets, rounds)))
			case 2:
				stat.Top2Game = GtypeNameMap[gtype]
				stat.Top2GameBets = fmt.Sprintf("%.2f", Chip2Float(bets))
				stat.Top2GameRounds = fmt.Sprint(rounds)
				stat.Top2GameBetsAvg = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(bets, rounds)))
			case 3:
				stat.Top3Game = GtypeNameMap[gtype]
				stat.Top3GameBets = fmt.Sprintf("%.2f", Chip2Float(bets))
				stat.Top3GameRounds = fmt.Sprint(rounds)
				stat.Top3GameBetsAvg = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(bets, rounds)))
			}
		}
	}

	// 汇总查询
	var bet_summary_datas []map[string]any
	err = ck.Select(&bet_summary_datas, fmt.Sprintf(`
		select s1.gtype, sum(rounds0) rounds_sum, sum(s1.bets0) bets_sum, sum(s1.rebates0) rebates_sum,
			rank() over (ORDER BY bets_sum desc) as ranking
		from game.col_user s0 final
		join (
			select userid, gtype, sum(rounds) rounds0, sum(bets) bets0, sum(rebates) rebates0
			from (
				select userid, gtype, count(*) rounds, sum(bet_amount) bets, sum(bet_amount+score) rebates
				from game.col_detail final where robot = 0 and begin_time >= ? and begin_time < ? and bet_amount != 0
				group by userid, gtype
				
				union all
				
				select t1.userid, t1.gtype, SUM(t1.rounds) rounds, sum(t1.amounts) bets, sum(t2.amounts) rebates 
				from (
					select user_id userid, game_id gtype, count(distinct round_id) rounds, sum(amount) amounts
					from game.col_nsq_log_external_bet final where ctime > ? and ctime < ? and amount != 0
					group by user_id, game_id
				) t1 left join (
					select user_id userid, game_id gtype, sum(amount) amounts
					from game.col_nsq_log_external_reward final where ctime > ? and ctime < ? and amount != 0
					group by user_id, game_id
				) t2 on t1.userid = t2.userid and t1.gtype = t2.gtype
				group by t1.userid, t1.gtype
			) s1
			group by userid, gtype
		) s1 on s1.userid = s0.userid
		where 1 = 1 %s
		group by s1.gtype
	`, where_u_sql), bet_user_args...)
	if err != nil {
		return
	}
	summary := &entity.PlayerBetsUser{
		SDate:            fmt.Sprintf("%s~%s汇总", stime.Format(utils.FORMAT_DATE), etime.Format(utils.FORMAT_DATE)),
		Rank:             "--",
		Userid:           "--",
		Ctime:            "--",
		LastLoginTime:    "--",
		LiveDays:         "--",
		LoseDays:         "--",
		VipLv:            "--",
		UserType:         "--",
		Pays:             "--",
		Withdraws:        "--",
		PaysSubWithdraws: "--",
	}
	stats = append([]*entity.PlayerBetsUser{summary}, stats...)
	for _, data := range bet_summary_datas {
		gtype := int(utils.ToInt64(data["gtype"]))
		rounds := utils.ToInt64(data["rounds_sum"])
		bets := utils.ToInt64(data["bets_sum"])
		rebates := utils.ToInt64(data["rebates_sum"])
		rank := utils.ToInt64(data["ranking"])

		summary.Bets0 += bets
		summary.Rabate0 += rebates
		switch rank {
		case 1:
			summary.Top1Game = GtypeNameMap[gtype]
			summary.Top1GameBets = fmt.Sprintf("%.2f", Chip2Float(bets))
			summary.Top1GameRounds = fmt.Sprint(rounds)
			summary.Top1GameBetsAvg = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(bets, rounds)))
		case 2:
			summary.Top2Game = GtypeNameMap[gtype]
			summary.Top2GameBets = fmt.Sprintf("%.2f", Chip2Float(bets))
			summary.Top2GameRounds = fmt.Sprint(rounds)
			summary.Top2GameBetsAvg = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(bets, rounds)))
		case 3:
			summary.Top3Game = GtypeNameMap[gtype]
			summary.Top3GameBets = fmt.Sprintf("%.2f", Chip2Float(bets))
			summary.Top3GameRounds = fmt.Sprint(rounds)
			summary.Top3GameBetsAvg = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(bets, rounds)))
		}
	}

	summary.Bets = fmt.Sprintf("%.2f", Chip2Float(summary.Bets0))
	summary.RabateRate = fmt.Sprintf("%.2f%%", ComputeFloat(summary.Rabate0, summary.Bets0)*100)
	summary.Wins = fmt.Sprintf("%.2f", Chip2Float(summary.Rabate0-summary.Bets0))
	return
}

func (c *statisticsService) PlayerBetsGtypeDates(calcTotal bool, page, pageSize int, stime, etime time.Time, registAreas, utypes []int32, packageIds []string) (total int64, stats []*entity.PlayerBetsGtypeDate, err error) {
	where_u_sql, where_u_args := s0ColUserFilter(registAreas, utypes, packageIds)

	var bet_user_args = []any{stime.Unix(), etime.Unix(), stime.Unix(), etime.Unix(), stime.Unix(), etime.Unix()}
	bet_user_args = append(bet_user_args, where_u_args...)

	// 查总条数
	if calcTotal {
		err = ck.Select(&total, fmt.Sprintf(`
		select count(*) total from (
			select s1.sdate, s1.gtype
			from game.col_user s0 final
			join (
				select userid, gtype, sdate, sum(bets) bets0, sum(rebates) rebates0
				from (
					select userid, gtype, toYYYYMMDD(toDateTime(begin_time)) sdate, sum(bet_amount) bets, sum(bet_amount+score) rebates
					from game.col_detail final where robot = 0 and begin_time >= ? and begin_time < ? and bet_amount != 0
					group by userid, gtype, sdate
					
					union all
					
					select t1.userid, t1.gtype, t1.sdate, sum(t1.amounts) bets, sum(t2.amounts) rebates 
					from (
						select user_id userid, game_id gtype, toYYYYMMDD(toDateTime(ctime)) sdate, count(distinct round_id) rounds, sum(amount) amounts
						from game.col_nsq_log_external_bet final where ctime > ? and ctime < ? and amount != 0
						group by user_id, game_id, sdate
					) t1 left join (
						select user_id userid, game_id gtype, toYYYYMMDD(toDateTime(ctime)) sdate, sum(amount) amounts
						from game.col_nsq_log_external_reward final where ctime > ? and ctime < ? and amount != 0
						group by user_id, game_id, sdate
					) t2 on t1.userid = t2.userid and t1.gtype = t2.gtype and t1.sdate = t2.sdate
					group by t1.userid, t1.gtype, t1.sdate
				) s1
				group by userid, gtype, sdate
			) s1 on s1.userid = s0.userid
			where 1 = 1 %s
			group by s1.sdate, s1.gtype
		) t0
		`, where_u_sql), bet_user_args...)
		if err != nil {
			return
		}
		if total == 0 {
			return
		}
	}

	// 汇总查询
	offset, limit := PageCalc(page, pageSize)
	bet_user_page_args := append(bet_user_args, offset, limit)
	var game_bet_datas []map[string]any
	err = ck.Select(&game_bet_datas, fmt.Sprintf(`
		select s1.sdate, s1.gtype, count(distinct s1.userid) bet_users, sum(s1.bets0) bets_sum, sum(s1.rebates0) rebates_sum
			-- , rank() over (PARTITION BY s1.sdate ORDER BY bets_sum desc) as ranking
		from game.col_user s0 final
		join (
			select userid, gtype, sdate, sum(bets) bets0, sum(rebates) rebates0
			from (
				select userid, gtype, toYYYYMMDD(toDateTime(begin_time)) sdate, sum(bet_amount) bets, sum(bet_amount+score) rebates
				from game.col_detail final where robot = 0 and begin_time >= ? and begin_time < ? and bet_amount != 0
				group by userid, gtype, sdate
				
				union all
				
				select t1.userid, t1.gtype, t1.sdate, sum(t1.amounts) bets, sum(t2.amounts) rebates 
				from (
					select user_id userid, game_id gtype, toYYYYMMDD(toDateTime(ctime)) sdate, count(distinct round_id) rounds, sum(amount) amounts
					from game.col_nsq_log_external_bet final where ctime > ? and ctime < ? and amount != 0
					group by user_id, game_id, sdate
				) t1 left join (
					select user_id userid, game_id gtype, toYYYYMMDD(toDateTime(ctime)) sdate, sum(amount) amounts
					from game.col_nsq_log_external_reward final where ctime > ? and ctime < ? and amount != 0
					group by user_id, game_id, sdate
				) t2 on t1.userid = t2.userid and t1.gtype = t2.gtype and t1.sdate = t2.sdate
				group by t1.userid, t1.gtype, t1.sdate
			) s1
			group by userid, gtype, sdate
		) s1 on s1.userid = s0.userid
		where 1 = 1 %s
		group by s1.sdate, s1.gtype
		order by s1.sdate desc, bets_sum desc
		limit ?, ?
	`, where_u_sql), bet_user_page_args...)
	if err != nil {
		return
	}

	for _, data := range game_bet_datas {
		_ = data
		sdate := data["sdate"].(uint32)
		gtype := int(utils.ToInt64(data["gtype"]))
		bet_users := utils.ToInt64(data["bet_users"])
		bets := utils.ToInt64(data["bets_sum"])
		rebates := utils.ToInt64(data["rebates_sum"])
		datestr := fmt.Sprint(sdate)
		fdate := datestr[0:4] + "-" + datestr[4:6] + "-" + datestr[6:8]

		stat := &entity.PlayerBetsGtypeDate{
			SDate: fdate,
		}
		stats = append(stats, stat)
		stat.GameId = fmt.Sprint(gtype)
		stat.GameName = GtypeNameMap[gtype]
		stat.BetUsers = bet_users
		stat.Bets = fmt.Sprintf("%.2f", Chip2Float(bets))
		stat.BetsAvg = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(bets, bet_users)))
		stat.Rebates = fmt.Sprintf("%.2f", Chip2Float(rebates))
		stat.BetsRollback = "0.00"
		stat.Wins = fmt.Sprintf("%.2f", Chip2Float(rebates-bets))
		stat.RTP = fmt.Sprintf("%.2f%%", ComputeFloat(rebates, bets)*100)

		// live和电子两类，live就是真人类的，电子就是非真人的。
		if utils.SliceIn(gtype, ZrsxGameids...) {
			stat.GtypeCategory1 = "live"
		} else {
			stat.GtypeCategory1 = "电子"
		}

		// pve百人/pvp对战/pve单机
		switch true {
		default:
			stat.GtypeCategory2 = "--"
		case utils.SliceIn(gtype, PVEBaiRenGameids...):
			stat.GtypeCategory2 = "pve百人"
		case utils.SliceIn(gtype, PVPDuiZhanGameids...):
			stat.GtypeCategory2 = "pvp对战"
		case utils.SliceIn(gtype, PVEDanJiGameids...):
			stat.GtypeCategory2 = "pve单机"
		}
		stat.Factory = GtypeFactoryMap[gtype]
	}

	// 汇总查询
	summary := &entity.PlayerBetsGtypeDate{
		SDate:          fmt.Sprintf("%s~%s汇总", stime.Format(utils.FORMAT_DATE), etime.Format(utils.FORMAT_DATE)),
		GtypeCategory1: "--",
		GtypeCategory2: "--",
		Factory:        "--",
		GameId:         "--",
		GameName:       "--",
	}
	stats = append([]*entity.PlayerBetsGtypeDate{summary}, stats...)

	var summary_game_bet_datas = make(map[string]any)
	err = ck.Select(&summary_game_bet_datas, fmt.Sprintf(`
		select count(distinct s1.userid) bet_users, sum(s1.bets0) bets_sum, sum(s1.rebates0) rebates_sum
		from game.col_user s0 final
		join (
			select userid, gtype, sdate, sum(bets) bets0, sum(rebates) rebates0
			from (
				select userid, gtype, toYYYYMMDD(toDateTime(begin_time)) sdate, sum(bet_amount) bets, sum(bet_amount+score) rebates
				from game.col_detail final where robot = 0 and begin_time >= ? and begin_time < ? and bet_amount != 0
				group by userid, gtype, sdate
				
				union all
				
				select t1.userid, t1.gtype, t1.sdate, sum(t1.amounts) bets, sum(t2.amounts) rebates 
				from (
					select user_id userid, game_id gtype, toYYYYMMDD(toDateTime(ctime)) sdate, count(distinct round_id) rounds, sum(amount) amounts
					from game.col_nsq_log_external_bet final where ctime > ? and ctime < ? and amount != 0
					group by user_id, game_id, sdate
				) t1 left join (
					select user_id userid, game_id gtype, toYYYYMMDD(toDateTime(ctime)) sdate, sum(amount) amounts
					from game.col_nsq_log_external_reward final where ctime > ? and ctime < ? and amount != 0
					group by user_id, game_id, sdate
				) t2 on t1.userid = t2.userid and t1.gtype = t2.gtype and t1.sdate = t2.sdate
				group by t1.userid, t1.gtype, t1.sdate
			) s1
			group by userid, gtype, sdate
		) s1 on s1.userid = s0.userid
		where 1 = 1 %s
	`, where_u_sql), bet_user_args...)
	if err != nil {
		return
	}
	data := summary_game_bet_datas
	bet_users := utils.ToInt64(data["bet_users"])
	bets := utils.ToInt64(data["bets_sum"])
	rebates := utils.ToInt64(data["rebates_sum"])

	summary.BetUsers = bet_users
	summary.Bets = fmt.Sprintf("%.2f", Chip2Float(bets))
	summary.BetsAvg = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(bets, bet_users)))
	summary.Rebates = fmt.Sprintf("%.2f", Chip2Float(rebates))
	summary.BetsRollback = "0.00"
	summary.Wins = fmt.Sprintf("%.2f", Chip2Float(rebates-bets))
	summary.RTP = fmt.Sprintf("%.2f%%", ComputeFloat(rebates, bets)*100)
	return
}
