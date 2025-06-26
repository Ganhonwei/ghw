package service

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/internal/web/stats/app/entity"
	"goserver/pkg/data/ck"
	"goserver/pkg/game/algo"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/astaxie/beego"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2/bson"
)

var (
	cfg *ini.File
	sec *ini.Section
	err error

	ExportDir = "./temp_stats_export"
)

// 数据汇总统计 (原数据汇总及渠道数据合并统计)
func DataStatistics1(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	Twotoday := tim.AddDate(0, 0, -1).Format("2006-01-02")

	m := bson.M{"status": 1}
	var channellist []entity.ChannelInfo
	err := Channels.
		Find(m).
		All(&channellist)
	if err != nil {
		beego.Error("DataStatistics fail err: ", err)
	}
	s := fmt.Sprintf("%s 00:00:00", today)
	startTime := utils.Str2TimeZone(s, locationName)
	e := fmt.Sprintf("%s 23:59:59", today)
	endTime := utils.Str2TimeZone(e, locationName)
	if len(channellist) > 0 {
		date1, _ := utils.Unix(fmt.Sprintf("%s", today))
		for _, item := range channellist {
			info := new(entity.DataStatistics)
			id := item.Name // 渠道名称
			name1 := item.Name1
			pipeline1 := []bson.M{
				{
					"$match": bson.M{
						"robot":            false,
						"simulation_robot": false,
						"ctime":            bson.M{"$gte": startTime, "$lt": endTime},
						"ad__bundle_id":    id,
					},
				},
				{
					"$group": bson.M{
						"_id": "$ad__adid",
						"num": bson.M{
							"$sum": 1,
						},
					},
				},
			}
			// operations := []bson.M{m, n}
			result1 := []bson.M{}
			pipe1 := PlayerUsers.Pipe(pipeline1)
			err := pipe1.All(&result1)
			if err != nil {
				beego.Error("DataStatistics fail err: ", err)
			}
			scount := int64(len(result1))

			// 查询登录日志登录用户
			pipeline := []bson.M{
				{
					"$match": bson.M{
						"login_time": bson.M{"$gte": startTime, "$lt": endTime},
					},
				},
				{
					"$group": bson.M{
						"_id": "$userid",
						"num": bson.M{
							"$sum": 1,
						},
					},
				},
			}
			result := []bson.M{}
			pipe := LoginLogs.Pipe(pipeline)
			err = pipe.All(&result)
			if err != nil {
				beego.Error("DataStatistics fail err: ", err)
			}
			logids := make([]string, 0)
			for _, item := range result {
				logids = append(logids, item["_id"].(string))
			}

			pipeline2 := []bson.M{
				{
					"$match": bson.M{
						"robot":            false,
						"simulation_robot": false,
						"_id":              bson.M{"$in": logids},
						"ad__bundle_id":    id,
					},
				},
				{
					"$group": bson.M{
						"_id": "$_id",
						"num": bson.M{
							"$sum": 1,
						},
					},
				},
			}
			result2 := []bson.M{}
			pipe2 := PlayerUsers.Pipe(pipeline2)
			err = pipe2.All(&result2)
			if err != nil {
				beego.Error("DataStatistics fail err: ", err)
			}
			count1 := int64(len(result2))

			m1 := FindByDate(today, today, "ctime", "ctime")
			m1["robot"] = false
			m1["simulation_robot"] = false
			m1["ad__bundle_id"] = id
			userlist, _ := PlayerService.GetByUserList(m1)
			count := int64(len(userlist))

			m2 := FindByDate(Twotoday, Twotoday, "ctime", "ctime")
			m2["robot"] = false
			m2["simulation_robot"] = false
			m2["ad__bundle_id"] = id
			count2, _ := PlayerService.GetTotal(m2)

			m3 := FindByDateBy2(Twotoday, Twotoday, today, today, "ctime", "login_time", "ctime", "login_time")
			m3["robot"] = false
			m3["simulation_robot"] = false
			m3["ad__bundle_id"] = id
			count3, _ := PlayerService.GetTotal(m3)

			// 获取今日登录昨日充值玩家
			zrpayCount := int64(0)
			zrCount := int64(0)
			var yids []string
			PlayerUsers.Find(m3).Distinct("_id", &yids)
			m8 := FindByDate(Twotoday, Twotoday, "ctime", "ctime")
			m8["userid"] = bson.M{"$in": logids}
			m8["order_status"] = 4
			m8["package_id"] = id
			var cids []string
			Pays.Find(m8).Distinct("userid", &cids)
			zrpayCount = int64(len(cids))
			//昨日充值的玩家人数
			m9 := FindByDate(Twotoday, Twotoday, "ctime", "ctime")
			m9["order_status"] = 4
			m9["package_id"] = id
			var zrids []string
			Pays.Find(m9).Distinct("userid", &zrids)
			zrCount = int64(len(zrids))

			// 新用户充值
			ids := make([]string, 0)
			for _, v := range userlist {
				ids = append(ids, v.Userid)
			}
			m4 := FindByDate(today, today, "ctime", "ctime")
			m4["userid"] = bson.M{"$in": ids}
			m4["order_status"] = 4
			m4["package_id"] = id
			list1, _ := PayService.GetByPayUser(m4)
			strid := ""
			Amount := int64(0)
			for _, v := range list1 {
				if strid == "" {
					strid = v.Userid
				} else {
					if !strings.Contains(strid, v.Userid) {
						strid += ("," + v.Userid)
					}
				}
				Amount += int64(v.Amount)
			}
			users := make([]string, 0)
			if strid != "" {
				users = strings.Split(strid, ",")
			}
			count4 := int64(len(users))
			// 总充值
			m5 := FindByDate(today, today, "ctime", "ctime")
			m5["order_status"] = 4
			m5["package_id"] = id
			list2, _ := PayService.GetByPayUser(m5)
			strid1 := ""
			Amount1 := int64(0)
			for _, v := range list2 {
				if strid1 == "" {
					strid1 = v.Userid
				} else {
					if !strings.Contains(strid1, v.Userid) {
						strid1 += ("," + v.Userid)
					}
				}
				Amount1 += int64(v.Amount)
			}
			users1 := make([]string, 0)
			if strid1 != "" {
				users1 = strings.Split(strid1, ",")
			}
			count5 := int64(len(users1))

			// 充值成功率
			cgCount := len(list2) // 充值成功订单数
			m7 := FindByDate(today, today, "ctime", "ctime")
			m7["package_id"] = id
			plist, _ := PayService.GetByPayUser(m7)
			soCount := int64(0)
			cgAmount := int64(0)
			sbAmount := int64(0)
			uniqueUserIDs := make(map[string]bool)
			for _, v := range plist {
				if v.OrderStatus == 4 {
					soCount += 1
					cgAmount += int64(v.Amount)
				}
				if v.OrderStatus == 2 {
					sbAmount += int64(v.Amount)
				}
				//ids = append(ids, v1.Userid)
				uniqueUserIDs[v.Userid] = true
			}
			var reqUser []string
			for id := range uniqueUserIDs {
				reqUser = append(reqUser, id)
			}
			reqNumber := int64(len(reqUser))
			reqOrder := int64(len(plist))

			// 提现
			tlist, _ := PayService.GetByWithdrawUser(m7)
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
				//ids = append(ids, v1.Userid)
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

			info.Date = date1
			info.Channel = id
			info.Channel1 = name1
			info.NewRegister = count
			info.NewEquipment = scount
			info.LoginNum = count1
			info.NextNewRegister = count2
			info.NextLogin = count3
			info.ZRPayCount = zrCount
			info.JRLoginNextPay = zrpayCount
			info.NewRechargeNum = count4
			info.NewRechargeAmount = Amount
			info.TotalRecharge = count5
			info.TotalAmount = Amount1
			// 充值成功率
			info.PayRequest = reqNumber
			info.PayRequestOrder = int64(reqOrder)
			info.PaySuccessOrder = int64(cgCount)
			// 提现成功率
			info.WithdrawRequest = reqNumber1
			info.WithdrawRequestOrder = reqOrder1
			info.WithdrawSuccessOrder = cgNumber1

			info.TotalWithdrawNum = sgNumber1
			info.TotalWithdrawAmount = cgAmount1
			info.HandlingCharge = HAmount1

			AddErr := StatisticsService.AddDataStatistics(info)
			if AddErr != nil {
				beego.Error("DataStatistics fail err: ", AddErr)
			}
		}
	}
}

/*
Author：CC
Title：用户留存
Effect：根据日期查询登录人数、注册人数
*/
func RetainedCompute(today string, today1 string, bid string) (rcount int64, count int64) {
	regcount1 := int64(0)
	count1 := int64(0)
	m1 := bson.M{
		"robot":            false,
		"simulation_robot": false,
		"ctime": bson.M{
			"$gte": utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today1), locationName),
			"$lt":  utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today1), locationName)},
		"share_superior": bid,
	}
	userIds := make([]string, 0)
	PlayerUsers.Find(m1).Distinct("_id", &userIds)
	regcount1 = int64(len(userIds))

	// 查询该博主下得所有用户
	m := bson.M{
		"robot":            false,
		"simulation_robot": false,
		"share_superior":   bid,
	}
	shareIds := make([]string, 0)
	PlayerUsers.Find(m).Distinct("_id", &shareIds)

	// 查询登录日志登录用户
	logids := make([]string, 0)
	LoginLogs.Find(bson.M{
		"login_time": bson.M{
			"$gte": utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName),
			"$lt":  utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)},
		"userid": bson.M{"$in": shareIds},
	}).Distinct("userid", &logids)
	count1 = int64(len(logids))
	return regcount1, count1
}

func PayRetainedCompute(today, today1, uid string, userArr []string) (rcount int64, count int64) {
	// 获取昨日注册，且充值玩家 今日登录
	zrCount := int64(0)  // 之前充值总人数
	zrCount1 := int64(0) // 之前充值，昨日登录人数

	// 获取总充值人数
	m := FindByDate(today1, today1, "ctime", "ctime")
	m["order_status"] = 4
	m["userid"] = bson.M{"$in": userArr}
	var cids []string
	Pays.Find(m).Distinct("userid", &cids)
	zrCount = int64(len(cids))

	// 昨日注册充值的玩家人数
	// 昨日注册
	m1 := bson.M{
		"ctime": bson.M{
			"$gte": utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today1), locationName),
			"$lt":  utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today1), locationName)},
		"robot":            false,
		"simulation_robot": false,
		"_id":              bson.M{"$in": cids},
	}
	zrzcId := make([]string, 0)
	PlayerUsers.Find(m1).Distinct("_id", &zrzcId)

	m2 := FindByDate(today, today, "login_time", "login_time")
	m2["userid"] = bson.M{"$in": zrzcId}
	var zrids []string
	LoginLogs.Find(m2).Distinct("userid", &zrids)
	zrCount1 = int64(len(zrids))

	return zrCount, zrCount1
}

func GameByUserCount(startTime, endTime int64, userid string) int {
	count := 0
	// 游戏对局
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"ctime":  bson.M{"$gte": startTime, "$lt": endTime},
				"userid": userid,
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Count": bson.M{
					"$sum": 1,
				},
			},
		},
	}
	result := []bson.M{}
	pipe := LogGameTimes.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("GameByUserCount fail err: ", err)
	}
	if len(result) > 0 {
		for _, item := range result {
			// ids = append(ids, item["_id"].(string))
			count += item["Count"].(int)
		}
	}

	// 外接游戏对局
	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"ctime":   bson.M{"$gte": startTime, "$lt": endTime},
				"user_id": userid,
			},
		},
		{
			"$group": bson.M{
				"_id":   "$user_id",
				"count": bson.M{"$addToSet": "$round_id"},
			},
		},
	}
	result1 := []bson.M{}
	pipe1 := NsqLogExternalBets.Pipe(pipeline1)
	err = pipe1.All(&result1)
	if len(result1) > 0 {
		map_num := make(map[string]bool, 0)
		for _, item := range result1 {
			arr_num := item["count"].([]interface{})
			for _, v := range arr_num {
				rid := v.(string)
				map_num[rid] = true
			}
		}
		count += len(map_num)
	}
	return count
}

// time analysis
func PlaytimeAnalysisData(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	s := fmt.Sprintf("%s 00:00:00", today)
	startTime := utils.Str2TimeZone(s, locationName)
	e := fmt.Sprintf("%s 23:59:59", today)
	endTime := utils.Str2TimeZone(e, locationName)

	m := bson.M{"status": 1}
	var channellist []entity.ChannelInfo
	err := Channels.
		Find(m).
		All(&channellist)
	if err != nil {
		beego.Error("PlaytimeAnalysisData fail err: ", err)
	}

	if len(channellist) > 0 {
		date1, _ := utils.Unix(fmt.Sprintf("%s", today))
		for _, item := range channellist {
			name := item.Name // 渠道名称
			name1 := item.Name1

			pipeline := []bson.M{
				{
					"$match": bson.M{
						"ctime":            bson.M{"$gte": startTime, "$lt": endTime},
						"robot":            false,
						"simulation_robot": false,
						"ad__bundle_id":    name,
					},
				},
				{
					"$group": bson.M{
						"_id": "$_id",
					},
				},
			}
			// operations := []bson.M{m, n}
			result := []bson.M{}
			pipe := PlayerUsers.Pipe(pipeline)
			err := pipe.All(&result)
			if err != nil {
				beego.Error("PlaytimeAnalysisData fail err: ", err)
			}
			// ids := make([]string, 0)
			palytime1 := int64(0)
			palytime2 := int64(0)
			palytime6 := int64(0)
			palytime11 := int64(0)
			palytime21 := int64(0)
			palytime31 := int64(0)
			statrTimestamp := startTime.Unix()
			endTimestamp := endTime.Unix()
			for _, item := range result {
				id := item["_id"].(string)
				userCount := PlaytimeByUserCount(statrTimestamp, endTimestamp, id)
				if userCount <= 119 {
					// 一分钟
					palytime1 += 1
				}
				if userCount >= 120 && userCount <= 365 {
					// 2-5 包含5分59秒
					palytime2 += 1
				}
				if userCount >= 360 && userCount <= 659 {
					// 6-10
					palytime6 += 1
				}
				if userCount >= 660 && userCount <= 1259 {
					palytime11 += 1
				}
				if userCount >= 1260 && userCount <= 1859 {
					palytime21 += 1
				}
				if userCount >= 1860 {
					palytime31 += 1
				}
			}

			rcount := len(result)
			info := new(entity.PlaytimeAnalysis)
			info.Date = date1
			info.Channel = name
			info.Channel1 = name1
			info.TotalRegister = int64(rcount)
			info.Playtime1 = palytime1
			info.Playtime2 = palytime2
			info.Playtime6 = palytime6
			info.Playtime11 = palytime11
			info.Playtime21 = palytime21
			info.Playtime31 = palytime31
			AddErr := StatisticsService.AddPlaytimeAnalysis(info)
			if AddErr != nil {
				beego.Error("PlaytimeAnalysisData fail err: ", AddErr.Error())
			}
		}
	}

}

func PlaytimeByUserCount(startTime, endTime int64, userid string) int64 {
	count := int64(0)
	// 游戏对局
	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"ctime":  bson.M{"$gte": startTime, "$lt": endTime},
				"userid": userid,
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Time": bson.M{
					"$sum": "$time",
				},
			},
		},
	}
	result1 := []bson.M{}
	pipe1 := LogGameTimes.Pipe(pipeline1)
	err1 := pipe1.All(&result1)
	if err1 != nil {
		beego.Error("PlaytimeByUserCount fail err: ", err1)
	}
	if len(result1) > 0 {
		for _, item := range result1 {
			// ids = append(ids, item["_id"].(string))
			count += item["Time"].(int64)
		}
	}
	return count
}

// 计算值
func ComputeFloat(val1, val2 int64) float64 {
	if val2 != 0 {
		result := float64(val1) / float64(val2)
		if math.IsNaN(result) {
			result = 0
		}
		return result
	}
	return 0
}

// 数据汇总统计 (原数据汇总及渠道数据合并统计)
func DataStatistics4Web(timestamp int64) {
	// loc, _ := time.LoadLocation(timeSetting)
	tim := utils.Stamp2Time(timestamp) //.In(loc)
	today := tim.Format("2006-01-02")
	Twotoday := tim.AddDate(0, 0, -1).Format("2006-01-02")
	beego.Info("DataStatistics4Web start: ", today)

	m := bson.M{"status": 1}
	var channellist []entity.ChannelInfo
	err := Channels.
		Find(m).
		All(&channellist)
	if err != nil {
		beego.Error("DataStatistics fail err: ", err)
	}
	s := fmt.Sprintf("%s 00:00:00", today)
	startTime := utils.Str2TimeZone(s, locationName)
	e := fmt.Sprintf("%s 23:59:59", today)
	endTime := utils.Str2TimeZone(e, locationName)
	addlist := make([]entity.DataStatistics4Web, 0)
	if len(channellist) > 0 {
		date1, _ := utils.Unix(fmt.Sprintf("%s", today))
		TypesList := []int{0, 1, 2}
		for _, item := range channellist {
			info := new(entity.DataStatistics4Web)
			id := item.Name // 渠道名称
			name1 := item.Name1
			for _, tsitem := range TypesList {
				// 用户列表
				or_m := bson.M{"$eq": tsitem}
				// 支付列表
				pay_or_m := bson.M{"$eq": tsitem}
				if tsitem == 0 {
					or_m = bson.M{"$nin": []int{1, 2}}
					pay_or_m = bson.M{"$nin": []int{1, 2}}
					// pay_or_m = bson.M{"user_type": bson.M{"$ne": 1}}
				}
				// 获取设备数
				ad_user := make([]string, 0)
				pipeline1 := FindByDate(today, today, "ctime", "ctime")
				pipeline1["robot"] = false
				pipeline1["simulation_robot"] = false
				pipeline1["ad__bundle_id"] = id
				pipeline1["regist_area"] = or_m
				pipeline1["ad__adid"] = bson.M{"$ne": ""}
				PlayerUsers.Find(pipeline1).Distinct("ad__adid", &ad_user)
				scount := int64(len(ad_user))
				// 查询登录日志登录用户
				logids := make([]string, 0)
				l_m := FindByDate(today, today, "login_time", "login_time")
				LoginLogs.Find(l_m).Distinct("userid", &logids)

				// login number
				reg_users := make([]string, 0)
				pipeline2 := bson.M{}
				pipeline2["robot"] = false
				pipeline2["simulation_robot"] = false
				pipeline2["ad__bundle_id"] = id
				pipeline2["_id"] = bson.M{"$in": logids}
				pipeline2["regist_area"] = or_m
				PlayerUsers.Find(pipeline2).Distinct("_id", &reg_users)
				count1 := int64(len(reg_users))

				// 付费用户登录数
				pay_reg := make([]string, 0)
				pay_reg_m := bson.M{}
				pay_reg_m["userid"] = bson.M{"$in": reg_users}
				pay_reg_m["order_status"] = 4
				pay_reg_m["package_id"] = id
				Pays.Find(pay_reg_m).Distinct("userid", &pay_reg)
				Paylogin := int64(len(pay_reg))

				// 新用户创建
				ids := make([]string, 0)
				// m1 := FindByDate(today, today, "ctime", "ctime")
				// m1["robot"] = false
				// m1["simulation_robot"] = false
				// m1["ad__bundle_id"] = id
				// m1["regist_area"] = or_m
				// PlayerUsers.Find(m1).Distinct("_id", &ids)
				// count := int64(len(ids))
				m1 := []bson.M{
					{
						"$match": bson.M{
							"ctime":            bson.M{"$gte": startTime, "$lt": endTime},
							"ad__bundle_id":    id,
							"simulation_robot": false,
							"robot":            false,
							"regist_area":      or_m,
						},
					},
					{
						"$group": bson.M{
							"_id": bson.M{
								"userid": "$_id",
								"ctime":  "$ctime",
							},
						},
					},
					{
						"$project": bson.M{
							"_id":   "$_id.userid",
							"ctime": "$_id.ctime",
						},
					},
				}
				var user_res []bson.M
				PlayerUsers.Pipe(m1).All(&user_res)
				count := int64(len(user_res))
				Minutes30 := 0
				Minutes60 := 0
				Minutes120 := 0
				if len(user_res) > 0 {
					for _, v := range user_res {
						uid := v["_id"].(string)
						ids = append(ids, uid)
						regtime := v["ctime"].(time.Time)
						pay_time_query := []bson.M{
							{
								"$match": bson.M{
									"order_status": 4,
									"userid":       uid,
								},
							},
							{
								"$sort": bson.M{
									"ctime": 1, // 根据订单时间升序排序
								},
							},
							{
								"$group": bson.M{
									"_id":      "$userid",
									"orderID":  bson.M{"$first": "$_id"},
									"pay_time": bson.M{"$first": "$pay_time"},
								},
							},
						}
						var pay_first_res []bson.M
						Pays.Pipe(pay_time_query).All(&pay_first_res)
						if len(pay_first_res) > 0 {
							// 有充值记录，判断第一条充值记录是否和注册时间相差
							for _, p := range pay_first_res {
								p_time := p["pay_time"].(time.Time)
								if p_time.After(regtime) && p_time.Before(regtime.Add(30*time.Minute)) {
									Minutes30++
								}
								if p_time.After(regtime) && p_time.Before(regtime.Add(60*time.Minute)) {
									Minutes60++
								}
								if p_time.After(regtime) && p_time.Before(regtime.Add(120*time.Minute)) {
									Minutes120++
								}
							}
						}
					}
				}

				m2 := FindByDate(Twotoday, Twotoday, "ctime", "ctime")
				m2["robot"] = false
				m2["simulation_robot"] = false
				m2["ad__bundle_id"] = id
				m2["regist_area"] = or_m
				count2, _ := PlayerService.GetTotal(m2)

				m3 := FindByDate(Twotoday, Twotoday, "ctime", "ctime")
				m3["robot"] = false
				m3["simulation_robot"] = false
				m3["ad__bundle_id"] = id
				m3["regist_area"] = or_m
				m3["_id"] = bson.M{"$in": logids}
				count3, _ := PlayerService.GetTotal(m3)

				// 获取今日登录昨日充值玩家
				zrpayCount := int64(0)
				zrCount := int64(0)
				var yids []string
				PlayerUsers.Find(m3).Distinct("_id", &yids)
				m8 := FindByDate(Twotoday, Twotoday, "ctime", "ctime")
				m8["userid"] = bson.M{"$in": logids}
				m8["order_status"] = 4
				m8["package_id"] = id
				m8["user_type"] = pay_or_m
				var cids []string
				Pays.Find(m8).Distinct("userid", &cids)
				zrpayCount = int64(len(cids))
				//昨日充值的玩家人数
				m9 := FindByDate(Twotoday, Twotoday, "ctime", "ctime")
				m9["order_status"] = 4
				m9["package_id"] = id
				m9["user_type"] = pay_or_m
				var zrids []string
				Pays.Find(m9).Distinct("userid", &zrids)
				zrCount = int64(len(zrids))

				// 新用户充值
				m4 := FindByDate(today, today, "ctime", "ctime")
				m4["userid"] = bson.M{"$in": ids}
				m4["order_status"] = 4
				m4["package_id"] = id
				m4["user_type"] = pay_or_m
				list1, _ := PayService.GetByPayUser(m4)
				strid := ""
				Amount := int64(0)
				newUserPayTimes := make(map[string]int, len(list1)) // 新用户付费次数
				for _, v := range list1 {
					if strid == "" {
						strid = v.Userid
					} else {
						if !strings.Contains(strid, v.Userid) {
							strid += ("," + v.Userid)
						}
					}
					Amount += int64(v.Amount)
					newUserPayTimes[v.Userid]++
				}
				users := make([]string, 0)
				if strid != "" {
					users = strings.Split(strid, ",")
				}
				count4 := int64(len(users))
				var newRechargeNum2 int64
				var newPaySuccessOrder int64 = int64(len(list1))
				for _, times := range newUserPayTimes {
					if times >= 2 {
						newRechargeNum2++
					}
				}

				// 总充值
				m5 := FindByDate(today, today, "ctime", "ctime")
				m5["order_status"] = 4
				m5["package_id"] = id
				m5["user_type"] = pay_or_m
				list2, _ := PayService.GetByPayUser(m5)
				strid1 := ""
				Amount1 := int64(0)
				userPayTimes := make(map[string]int, len(list1)) // 用户付费次数
				for _, v := range list2 {
					if strid1 == "" {
						strid1 = v.Userid
					} else {
						if !strings.Contains(strid1, v.Userid) {
							strid1 += ("," + v.Userid)
						}
					}
					Amount1 += int64(v.Amount)
					userPayTimes[v.Userid]++
				}
				users1 := make([]string, 0)
				if strid1 != "" {
					users1 = strings.Split(strid1, ",")
				}
				count5 := int64(len(users1))
				var totalRechargeNum2 int64
				for _, times := range userPayTimes {
					if times >= 2 {
						totalRechargeNum2++
					}
				}
				// 充值成功率
				cgCount := len(list2) // 充值成功订单数
				m7 := FindByDate(today, today, "ctime", "ctime")
				m7["package_id"] = id
				m7["user_type"] = pay_or_m
				plist, _ := PayService.GetByPayUser(m7)
				soCount := int64(0)
				cgAmount := int64(0)
				sbAmount := int64(0)
				uniqueUserIDs := make(map[string]bool)
				for _, v := range plist {
					if v.OrderStatus == 4 {
						soCount += 1
						cgAmount += int64(v.Amount)
					}
					if v.OrderStatus == 2 {
						sbAmount += int64(v.Amount)
					}
					//ids = append(ids, v1.Userid)
					uniqueUserIDs[v.Userid] = true
				}
				var reqUser []string
				for id := range uniqueUserIDs {
					reqUser = append(reqUser, id)
				}
				reqNumber := int64(len(reqUser))
				reqOrder := int64(len(plist))

				// 提现
				tlist, _ := PayService.GetByWithdrawUser(m7)
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
					//ids = append(ids, v1.Userid)
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

				// 老用户
				var oldUserid []string
				m10 := bson.M{}
				m10["ctime"] = bson.M{"$lt": startTime}
				m10["robot"] = false
				m10["simulation_robot"] = false
				m10["ad__bundle_id"] = id
				m10["_id"] = bson.M{"$in": logids}
				m10["regist_area"] = or_m
				PlayerUsers.Find(m10).Distinct("_id", &oldUserid)
				oldNumber := len(oldUserid)

				// 老用户充值
				pipeline3 := []bson.M{
					{
						"$match": bson.M{
							"ctime":        bson.M{"$gte": startTime, "$lt": endTime},
							"order_status": 4,
							"userid":       bson.M{"$in": oldUserid},
							"package_id":   id,
							"user_type":    pay_or_m,
						},
					},
					{
						"$group": bson.M{
							"_id": "$userid",
							"Amount": bson.M{
								"$sum": "$amount",
							},
							"Times": bson.M{ // 用户充值次数
								"$sum": 1,
							},
						},
					},
				}
				result3 := []bson.M{}
				pipe3 := Pays.Pipe(pipeline3)
				err = pipe3.All(&result3)
				if err != nil {
					beego.Error("DataStatistics fail err: ", err)
				}
				oldAmount := 0
				oldCount := 0
				var oldRechargeNum2 int64
				var oldPaySuccessOrder int64
				for _, item := range result3 {
					oldAmount += item["Amount"].(int)

					times := item["Times"].(int)
					oldPaySuccessOrder += int64(times)
					if times >= 2 {
						oldRechargeNum2++
					}
				}
				oldCount = len(result3)

				// 信任用户充值、提现
				xr_m := []bson.M{
					{
						"$match": bson.M{
							"order_status": 4,
							"package_id":   id,
						},
					},
					{
						"$group": bson.M{
							"_id":   "$userid",
							"count": bson.M{"$sum": 1}, // 添加一个额外的字段以便后续过滤
						},
					},
					{
						"$match": bson.M{
							"count": bson.M{"$gte": 2}, // 过滤count大于2
						},
					},
				}
				var xr_user_res []bson.M
				Pays.Pipe(xr_m).All(&xr_user_res)
				var xrids []string
				var temp_ids []string

				for _, res := range xr_user_res {
					xid := res["_id"].(string)
					temp_ids = append(temp_ids, xid)
				}
				// 获取对应的用户类型
				if len(temp_ids) > 0 {
					xr_user_m := bson.M{}
					xr_user_m["robot"] = false
					xr_user_m["simulation_robot"] = false
					xr_user_m["ad__bundle_id"] = id
					xr_user_m["regist_area"] = or_m
					xr_user_m["_id"] = bson.M{"$in": temp_ids}
					PlayerUsers.Find(xr_user_m).Distinct("_id", &xrids)
				}

				// 获取信任用户当天充值金额
				xr_pay_m := []bson.M{
					{
						"$match": bson.M{
							"order_status": 4,
							"package_id":   id,
							"ctime":        bson.M{"$gte": startTime, "$lt": endTime},
							"userid":       bson.M{"$in": xrids},
						},
					},
					{
						"$group": bson.M{
							"_id": nil,
							"Amount": bson.M{
								"$sum": "$amount",
							},
						},
					},
				}
				var xr_pay_res []bson.M
				Pays.Pipe(xr_pay_m).All(&xr_pay_res)
				xrpayamount := 0
				for _, res := range xr_pay_res {
					xr_amount := res["Amount"].(int)
					xrpayamount += xr_amount
				}
				// 获取信任用户当天提现金额
				xr_with_m := []bson.M{
					{
						"$match": bson.M{
							"order_status": 2,
							"package_id":   id,
							"ctime":        bson.M{"$gte": startTime, "$lt": endTime},
							"userid":       bson.M{"$in": xrids},
						},
					},
					{
						"$group": bson.M{
							"_id": nil,
							"Amount": bson.M{
								"$sum": "$amount",
							},
						},
					},
				}
				var xr_with_res []bson.M
				Withdraws.Pipe(xr_with_m).All(&xr_with_res)
				xrwithamount := 0
				for _, res := range xr_with_res {
					xr_amount := res["Amount"].(int)
					xrwithamount += xr_amount
				}

				info.Date = date1
				info.Channel = id
				info.Channel1 = name1
				info.DataTypes = tsitem // 用户类型
				info.NewRegister = count
				info.NewEquipment = scount
				info.LoginNum = count1
				info.PayLoginNumber = Paylogin
				info.OldLoginNum = int64(oldNumber)
				info.OldRechargeAmount = int64(oldAmount)
				info.OldRechargeNum = int64(oldCount)

				info.NextNewRegister = count2
				info.NextLogin = count3
				info.ZRPayCount = zrCount
				info.JRLoginNextPay = zrpayCount
				info.NewRechargeNum = count4
				info.NewRechargeAmount = Amount
				info.TotalRecharge = count5
				info.TotalAmount = Amount1
				// 充值成功率
				info.PayRequest = reqNumber
				info.PayRequestOrder = int64(reqOrder)
				info.PaySuccessOrder = int64(cgCount)
				// 提现成功率
				info.WithdrawRequest = reqNumber1
				info.WithdrawRequestOrder = reqOrder1
				info.WithdrawSuccessOrder = cgNumber1

				info.TotalWithdrawNum = sgNumber1
				info.TotalWithdrawAmount = cgAmount1
				info.HandlingCharge = HAmount1

				info.OldRechargeNum2 = oldRechargeNum2
				info.NewRechargeNum2 = newRechargeNum2
				info.TotalRechargeNum2 = totalRechargeNum2
				info.OldPaySuccessOrder = oldPaySuccessOrder
				info.NewPaySuccessOrder = newPaySuccessOrder
				info.PayMinutes30 = int64(Minutes30)
				info.PayMinutes60 = int64(Minutes60)
				info.PayMinutes120 = int64(Minutes120)
				info.TrustUserPay = int64(xrpayamount)
				info.TrustUserWithdraw = int64(xrwithamount)

				addlist = append(addlist, *info)
			}
		}
	}

	beego.Info(fmt.Sprintf("DataStatistics4Web exec: %s~%s, datas=%d", s, e, len(addlist)))
	for _, item := range addlist {
		AddErr := StatisticsService.AddDataStatistics4Web(&item)
		if AddErr != nil {
			beego.Error("DataStatistics fail err: ", AddErr)
		}
	}
}

// func DataStatistics4WebCK(timestamp int64) {
// 	tim := utils.Stamp2Time(timestamp)
// 	today := tim.Format("2006-01-02")
// 	Twotoday := tim.AddDate(0, 0, -1).Format("2006-01-02")
// 	s := fmt.Sprintf("%s 00:00:00", today)
// 	startTime := utils.Str2Time(s)
// 	e := fmt.Sprintf("%s 23:59:59", today)
// 	endTime := utils.Str2Time(e)
// 	date1, _ := utils.Unix(fmt.Sprintf("%s", today))
// 	m := bson.M{"status": 1}
// 	var channellist []entity.ChannelInfo
// 	err := Channels.
// 		Find(m).
// 		All(&channellist)
// 	if err != nil {
// 		beego.Error("DataStatistics4WebCK fail err: ", err)
// 	}
// 	addlist := make(map[string]*entity.DataStatistics4Web, 0)
// 	var reg_info []map[string]any
// 	err = ck.Select(&reg_info, `
// 	select ad__bundle_id,regist_area,COUNT(1) number, COUNT(DISTINCT CASE WHEN ad__adid != '' THEN ad__adid END ) ad_number from game.col_user cu FINAL
// where robot = 0 and simulation_robot = 0 and ctime between ? and ? group by ad__bundle_id,regist_area
// `, startTime, endTime)
// 	if err != nil {
// 		beego.Error("DataStatistics4WebCK error:", err)
// 		return
// 	}

// 	for _, item := range reg_info {
// 		adid := item["ad__bundle_id"].(string)
// 		loginDays := utils.ToInt64(item["login_days"])
// 		key := fmt.Sprintf("%s-%d", pid, cid)
// 		stat, ok := addlist[key]
// 		if !ok {
// 			stat = &entity.DataStatistics4Web{
// 				Date: date1,
// 			}
// 			addlist[key] = stat
// 		}
// 	}

// 	// 充值
// 	sql1 := `SELECT
//     channel_id,
//     package_id,
//     COUNT(DISTINCT userid) AS total_number,
//     COUNT(1)  AS total_count,
//     SUM(amount) AS total_amount,
//     COUNT(DISTINCT CASE WHEN order_status = 4 THEN userid END) AS success_number,
//     SUM(CASE WHEN order_status = 4 THEN 1 ELSE 0 END) AS success_count,
//     SUM(CASE WHEN order_status = 4 THEN amount ELSE 0 END) AS success_amount,
//     SUM(CASE WHEN order_status = 2 THEN amount ELSE 0 END) AS fail_amount
// FROM game.col_trade_record ctr FINAL
// WHERE package_id != '' and ctime BETWEEN ? AND ?
// GROUP BY channel_id, package_id`
// 	// ctime between ? and ? and
// 	var paylist []map[string]any
// 	var args1 []any
// 	args1 = append(args1, startTime)
// 	args1 = append(args1, endTime)
// 	err = ck.Select(&paylist, sql1, args1...)
// 	if err != nil {
// 		beego.Error("DataStatistics4WebCK error:", err)
// 	}

// 	for _, item := range addlist {
// 		AddErr := StatisticsService.AddDataStatistics4Web(&item)
// 		if AddErr != nil {
// 			beego.Error("DataStatistics4WebCK fail err: ", AddErr)
// 		}
// 	}
// }

/*
Author:CC
Title:数据汇总
*/
func DataSummary(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	Twotoday := tim.AddDate(0, 0, -1).Format("2006-01-02")
	m := FindByDate(today, today, "ctime", "ctime")
	m["robot"] = false
	list, _ := PlayerService.GetByUserList(m)
	count := int64(len(list))
	s := fmt.Sprintf("%s 00:00:00", today)
	startTime := utils.Str2TimeZone(s, locationName)
	e := fmt.Sprintf("%s 23:59:59", today)
	endTime := utils.Str2TimeZone(e, locationName)

	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"ctime":            bson.M{"$gte": startTime, "$lt": endTime},
				"robot":            false,
				"simulation_robot": false,
			},
		},
		{
			"$group": bson.M{
				"_id": "$ad__adid",
				"num": bson.M{
					"$sum": 1,
				},
			},
		},
	}
	// operations := []bson.M{m, n}
	result1 := []bson.M{}
	pipe1 := PlayerUsers.Pipe(pipeline1)
	err := pipe1.All(&result1)
	if err != nil {
		beego.Error("DataSummary fail err: ", err)
	}
	scount := int64(len(result1))

	// pipeline := []bson.M{
	// 	{
	// 		"$match": bson.M{
	// 			"login_time": bson.M{"$gte": startTime, "$lt": endTime},
	// 		},
	// 	},
	// 	{
	// 		"$group": bson.M{
	// 			"_id": "$userid",
	// 			"num": bson.M{
	// 				"$sum": 1,
	// 			},
	// 		},
	// 	},
	// }
	// // operations := []bson.M{m, n}
	// result := []bson.M{}
	// pipe := LoginLogs.Pipe(pipeline)
	// err = pipe.All(&result)
	// if err != nil {
	// 	beego.Error("DataSummary fail err: ", err)
	// }
	m7 := FindByDate(today, today, "login_time", "login_time")
	var logids []string
	LoginLogs.Find(m7).Distinct("userid", &logids)

	count1 := int64(len(logids))
	// if len(result) > 0 {
	// 	temp := result[0]["num"].(int)
	// 	count1 = int64(temp)
	// }

	m2 := FindByDate(Twotoday, Twotoday, "ctime", "ctime")
	m2["robot"] = false
	count2, _ := PlayerService.GetTotal(m2)

	m3 := FindByDateBy2(Twotoday, Twotoday, today, today, "ctime", "login_time", "ctime", "login_time")
	m3["robot"] = false
	m["simulation_robot"] = false
	count3, _ := PlayerService.GetTotal(m3)

	// 获取今日登录昨日充值玩家
	zrpayCount := int64(0)
	zrCount := int64(0)
	var yids []string
	PlayerUsers.Find(m3).Distinct("_id", &yids)
	m8 := FindByDate(Twotoday, Twotoday, "ctime", "ctime")
	m8["userid"] = bson.M{"$in": logids}
	m8["order_status"] = 4
	var cids []string
	Pays.Find(m8).Distinct("userid", &cids)
	zrpayCount = int64(len(cids))
	//昨日充值的玩家人数
	m9 := FindByDate(Twotoday, Twotoday, "ctime", "ctime")
	m9["order_status"] = 4
	var zrids []string
	Pays.Find(m9).Distinct("userid", &zrids)
	zrCount = int64(len(zrids))

	// 注册
	ids := make([]string, 0)
	for _, v := range list {
		ids = append(ids, v.Userid)
	}
	m4 := FindByDate(today, today, "ctime", "ctime")
	m4["userid"] = bson.M{"$in": ids}
	m4["order_status"] = 4
	list1, _ := PayService.GetByPayUser(m4)
	strid := ""
	Amount := int64(0)
	// ids1 := make([]string, 0)
	for _, v := range list1 {
		if strid == "" {
			strid = v.Userid
		} else {
			if !strings.Contains(strid, v.Userid) {
				strid += ("," + v.Userid)
			}
		}
		// ids1 = append(ids1, v.Userid)
		Amount += int64(v.Amount)
	}
	users := make([]string, 0)
	if strid != "" {
		users = strings.Split(strid, ",")
	}
	count4 := int64(len(users))

	m5 := FindByDate(today, today, "ctime", "ctime")
	m5["order_status"] = 4
	list2, _ := PayService.GetByPayUser(m5)
	strid1 := ""
	Amount1 := int64(0)
	for _, v := range list2 {
		if strid1 == "" {
			strid1 = v.Userid
		} else {
			if !strings.Contains(strid1, v.Userid) {
				strid1 += ("," + v.Userid)
			}
		}
		Amount1 += int64(v.Amount)
	}
	users1 := make([]string, 0)
	if strid1 != "" {
		users1 = strings.Split(strid1, ",")
	}
	count5 := int64(len(users1))

	m6 := FindByDate(today, today, "ctime", "ctime")
	m6["order_status"] = 2
	list3, _ := PayService.GetByWithdrawUser(m6)
	strid2 := ""
	wAmount1 := int64(0)
	HAmount := int64(0)
	for _, v := range list3 {
		if strid2 == "" {
			strid2 = v.Userid
		} else {
			if !strings.Contains(strid2, v.Userid) {
				strid2 += ("," + v.Userid)
			}
		}
		wAmount1 += int64(v.Amount)
		HAmount += int64(v.Commission)
	}
	wid := make([]string, 0)
	if strid2 != "" {
		wid = strings.Split(strid2, ",")
	}
	count6 := int64(len(wid))
	date1, _ := utils.Unix(fmt.Sprintf("%s", today))
	FAmount := ComputeFloat(Amount, 100)
	FAmount1 := ComputeFloat(Amount1, 100)
	FAmount2 := ComputeFloat(wAmount1, 100)
	FAmount3 := ComputeFloat(HAmount, 100)
	// 数据汇总
	info := new(entity.DataSummary)
	info.Date = date1
	info.NewRegister = count
	info.NewEquipment = scount
	info.LoginNum = count1
	info.NextDayRetention = ComputeFloat(count3, count2) * 100
	info.PayRetention = ComputeFloat(zrpayCount, zrCount) * 100
	info.NewRechargeNum = count4
	info.NewRechargeAmount = FAmount
	info.NewRechargeArpu = ComputeFloat(int64(FAmount), count)
	info.NewRechargeArppu = ComputeFloat(int64(FAmount), count4)
	info.NewUserPaymentRate = ComputeFloat(count4, count) * 100
	info.TotalRecharge = count5
	info.TotalAmount = FAmount1
	info.TotalRechargeArpu = ComputeFloat(int64(FAmount1), count1)
	info.TotalRechargeArppu = ComputeFloat(int64(FAmount1), count5)
	info.TotalPaymentRate = ComputeFloat(count5, count1) * 100
	info.TotalWithdrawNum = count6
	info.TotalWithdrawAmount = FAmount2
	info.HandlingCharge = FAmount3
	info.CostRatio = ComputeFloat(int64(FAmount2), int64(FAmount1)) * 100

	AddErr := StatisticsService.AddDataSummary(info)
	if AddErr != nil {
		beego.Error("DataSummary fail err: ", AddErr)
	}
}

/*
Author:CC
Title:渠道数据
*/
func ChannelData(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	Twotoday := tim.AddDate(0, 0, -1).Format("2006-01-02")

	m := bson.M{"status": 1}
	var channellist []entity.ChannelInfo
	err := Channels.
		Find(m).
		All(&channellist)
	if err != nil {
		beego.Error("ChannelData fail err: ", err)
	}
	s := fmt.Sprintf("%s 00:00:00", today)
	startTime := utils.Str2TimeZone(s, locationName)
	e := fmt.Sprintf("%s 23:59:59", today)
	endTime := utils.Str2TimeZone(e, locationName)
	if len(channellist) > 0 {
		date1, _ := utils.Unix(fmt.Sprintf("%s", today))
		for _, item := range channellist {
			info := new(entity.ChannelData)
			id := item.Name // 渠道名称
			name1 := item.Name1
			pipeline1 := []bson.M{
				{
					"$match": bson.M{
						"robot":            false,
						"simulation_robot": false,
						"ctime":            bson.M{"$gte": startTime, "$lt": endTime},
						"ad__bundle_id":    id,
					},
				},
				{
					"$group": bson.M{
						"_id": "$ad__adid",
						"num": bson.M{
							"$sum": 1,
						},
					},
				},
			}
			// operations := []bson.M{m, n}
			result1 := []bson.M{}
			pipe1 := PlayerUsers.Pipe(pipeline1)
			err := pipe1.All(&result1)
			if err != nil {
				beego.Error("ChannelData fail err: ", err)
			}
			scount := int64(len(result1))

			// 查询登录日志登录用户
			pipeline := []bson.M{
				{
					"$match": bson.M{
						"login_time": bson.M{"$gte": startTime, "$lt": endTime},
					},
				},
				{
					"$group": bson.M{
						"_id": "$userid",
						"num": bson.M{
							"$sum": 1,
						},
					},
				},
			}
			result := []bson.M{}
			pipe := LoginLogs.Pipe(pipeline)
			err = pipe.All(&result)
			if err != nil {
				beego.Error("ChannelData fail err: ", err)
			}
			logids := make([]string, 0)
			for _, item := range result {
				logids = append(logids, item["_id"].(string))
			}

			pipeline2 := []bson.M{
				{
					"$match": bson.M{
						"robot":            false,
						"simulation_robot": false,
						"_id":              bson.M{"$in": logids},
						"ad__bundle_id":    id,
					},
				},
				{
					"$group": bson.M{
						"_id": "$_id",
						"num": bson.M{
							"$sum": 1,
						},
					},
				},
			}
			// operations := []bson.M{m, n}
			result2 := []bson.M{}
			pipe2 := PlayerUsers.Pipe(pipeline2)
			err = pipe2.All(&result2)
			if err != nil {
				beego.Error("ChannelData fail err: ", err)
			}
			count1 := int64(len(result2))

			m1 := FindByDate(today, today, "ctime", "ctime")
			m1["robot"] = false
			m1["simulation_robot"] = false
			m1["ad__bundle_id"] = id
			userlist, _ := PlayerService.GetByUserList(m1)
			count := int64(len(userlist))

			m2 := FindByDate(Twotoday, Twotoday, "ctime", "ctime")
			m2["robot"] = false
			m2["simulation_robot"] = false
			m2["ad__bundle_id"] = id
			count2, _ := PlayerService.GetTotal(m2)

			m3 := FindByDateBy2(Twotoday, Twotoday, today, today, "ctime", "login_time", "ctime", "login_time")
			m3["robot"] = false
			m3["simulation_robot"] = false
			m3["ad__bundle_id"] = id
			count3, _ := PlayerService.GetTotal(m3)

			// 新用户充值
			ids := make([]string, 0)
			for _, v := range userlist {
				ids = append(ids, v.Userid)
			}
			m4 := FindByDate(today, today, "ctime", "ctime")
			m4["userid"] = bson.M{"$in": ids}
			m4["order_status"] = 4
			m4["package_id"] = id
			list1, _ := PayService.GetByPayUser(m4)
			strid := ""
			Amount := int64(0)
			for _, v := range list1 {
				if strid == "" {
					strid = v.Userid
				} else {
					if !strings.Contains(strid, v.Userid) {
						strid += ("," + v.Userid)
					}
				}
				Amount += int64(v.Amount)
			}
			users := make([]string, 0)
			if strid != "" {
				users = strings.Split(strid, ",")
			}
			count4 := int64(len(users))
			// 总充值
			m5 := FindByDate(today, today, "ctime", "ctime")
			m5["order_status"] = 4
			m5["package_id"] = id
			list2, _ := PayService.GetByPayUser(m5)
			strid1 := ""
			Amount1 := int64(0)
			for _, v := range list2 {
				if strid1 == "" {
					strid1 = v.Userid
				} else {
					if !strings.Contains(strid1, v.Userid) {
						strid1 += ("," + v.Userid)
					}
				}
				Amount1 += int64(v.Amount)
			}
			users1 := make([]string, 0)
			if strid1 != "" {
				users1 = strings.Split(strid1, ",")
			}
			count5 := int64(len(users1))

			// // 提现
			// m6 := FindByDate(today, today, "ctime", "ctime")
			// m6["order_status"] = 2
			// m6["package_id"] = id
			// list3, _ := PayService.GetByWithdrawUser(m6)
			// strid2 := ""
			// wAmount1 := int64(0)
			// HAmount := int64(0)
			// for _, v := range list3 {
			// 	if strid2 == "" {
			// 		strid2 = v.Userid
			// 	} else {
			// 		if !strings.Contains(strid2, v.Userid) {
			// 			strid2 += ("," + v.Userid)
			// 		}
			// 	}
			// 	wAmount1 += int64(v.Amount)
			// 	HAmount += int64(v.Commission)
			// }
			// wid := make([]string, 0)
			// if strid2 != "" {
			// 	wid = strings.Split(strid2, ",")
			// }
			// count6 := int64(len(wid))

			// 充值成功率
			cgCount := len(list2) // 充值成功订单数
			m7 := FindByDate(today, today, "ctime", "ctime")
			m7["package_id"] = id
			plist, _ := PayService.GetByPayUser(m7)
			soCount := int64(0)
			cgAmount := int64(0)
			sbAmount := int64(0)
			uniqueUserIDs := make(map[string]bool)
			for _, v := range plist {
				if v.OrderStatus == 4 {
					soCount += 1
					cgAmount += int64(v.Amount)
				}
				if v.OrderStatus == 2 {
					sbAmount += int64(v.Amount)
				}
				//ids = append(ids, v1.Userid)
				uniqueUserIDs[v.Userid] = true
			}
			var reqUser []string
			for id := range uniqueUserIDs {
				reqUser = append(reqUser, id)
			}
			reqNumber := int64(len(reqUser))
			reqOrder := int64(len(plist))

			// 提现
			tlist, _ := PayService.GetByWithdrawUser(m7)
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
				//ids = append(ids, v1.Userid)
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

			FAmount := ComputeFloat(Amount, 100)
			FAmount1 := ComputeFloat(Amount1, 100)
			FAmount2 := ComputeFloat(cgAmount1, 100) // 提现总金额
			// FAmount2 := ComputeFloat(wAmount1, 100)
			// FAmount3 := ComputeFloat(HAmount, 100)

			info.Date = date1
			info.Channel = id
			info.Channel1 = name1
			info.NewRegister = count
			info.NewEquipment = scount
			info.LoginNum = count1
			info.NextDayRetention = ComputeFloat(count3, count2) * 100
			info.NewRechargeNum = count4
			info.NewRechargeAmount = FAmount
			info.NewRechargeArpu = ComputeFloat(int64(FAmount), count)
			info.NewRechargeArppu = ComputeFloat(int64(FAmount), count4)
			info.NewUserPaymentRate = ComputeFloat(count4, count) * 100
			info.TotalRecharge = count5
			info.TotalAmount = FAmount1
			info.TotalRechargeArpu = ComputeFloat(int64(FAmount1), count1)
			info.TotalRechargeArppu = ComputeFloat(int64(FAmount1), count5)
			info.TotalPaymentRate = ComputeFloat(count5, count1) * 100
			// 充值成功率
			info.PayRequest = reqNumber
			info.PayRequestOrder = int64(reqOrder)
			info.PaySuccessOrder = int64(cgCount)
			info.PaySuccessRate = ComputeFloat(int64(cgCount), int64(reqOrder)) * 100
			// 提现成功率
			info.WithdrawRequest = reqNumber1
			info.WithdrawRequestOrder = reqOrder1
			info.WithdrawSuccessOrder = cgNumber1
			info.WithdrawSuccessRate = ComputeFloat(cgNumber1, reqOrder1) * 100

			info.TotalWithdrawNum = sgNumber1
			info.TotalWithdrawAmount = FAmount2
			info.HandlingCharge = ComputeFloat(HAmount1, 100)
			info.CostRatio = ComputeFloat(int64(FAmount2), int64(FAmount1)) * 100

			AddErr := StatisticsService.AddChannelData(info)
			if AddErr != nil {
				beego.Error("ChannelData fail err: ", AddErr)
			}
		}
	}
}

/*
Author:CC
Title: 实时数据统计 -> 充值金额
Time: 每个小时统计一次
*/
func RealTimeData(timestamp int64) {
	now := utils.Stamp2Time(timestamp) // time.Now() //
	// oneHourAgo := now.Add(-time.Hour)
	startTime := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
	endTime := startTime.Add(time.Hour - time.Nanosecond)

	m := bson.M{}
	m["login_time"] = bson.M{"$gte": startTime, "$lte": endTime}
	m["robot"] = false
	lCount, _ := PlayerService.GetTotal(m)
	m1 := bson.M{}
	m1["login_time"] = bson.M{"$gte": startTime, "$lte": endTime}
	m1["robot"] = false
	m1["online_status"] = true
	sCount, _ := PlayerService.GetTotal(m1)
	// 充值
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"ctime":        bson.M{"$gte": startTime, "$lt": endTime},
				"order_status": 4,
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
	pipe := Pays.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("RealTimeData fail err: ", err)
	}
	Amount := 0
	for _, item := range result {
		// ids = append(ids, item["_id"].(string))
		Amount += item["Amount"].(int)
	}
	count := len(result)

	// 提现
	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"ctime":        bson.M{"$gte": startTime, "$lt": endTime},
				"order_status": 2,
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
	pipe1 := Withdraws.Pipe(pipeline1)
	err = pipe1.All(&result1)
	if err != nil {
		beego.Error("RealTimeData fail err: ", err)
	}
	Amount1 := 0
	for _, item := range result1 {
		// ids = append(ids, item["_id"].(string))
		Amount1 += item["Amount"].(int)
	}
	count1 := len(result1)

	//拼装数据
	info := new(entity.RealTimeData)
	info.Date = startTime
	info.LoginNumber = int64(lCount)
	info.OnlineNumber = int64(sCount)
	info.PayNumber = int64(count)
	info.PayAmount = ComputeFloat(int64(Amount), 100)
	info.WithdrawNumber = int64(count1)
	info.WithdrawAmount = ComputeFloat(int64(Amount1), 100)
	err1 := StatisticsService.AddRealTimeData(info)
	if err1 != nil {
		beego.Error("RealTimeData fail err: ", err1)
	}
}

/*
Author:CC
Title:活动管理 -> 基础付费活动
*/
func BasicPayActivity(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	s := fmt.Sprintf("%s 00:00:00", today)
	startTime := utils.Str2TimeZone(s, locationName)
	e := fmt.Sprintf("%s 23:59:59", today)
	endTime := utils.Str2TimeZone(e, locationName)
	// 入门礼包
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": startTime, "$lt": endTime},
				"ltype": 58,
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Diamond": bson.M{
					"$sum": "$add_diamond",
				},
				"Coin": bson.M{
					"$sum": "$add_coin",
				},
			},
		},
	}
	result := []bson.M{}
	pipe := LogWaters.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("BasicPayActivity fail err: ", err)
	}
	// ids := make([]string, 0)
	Diamond := int64(0)
	Coin := int64(0)
	for _, item := range result {
		// ids = append(ids, item["_id"].(string))
		Diamond += item["Diamond"].(int64)
		Coin += item["Coin"].(int64)
	}
	count := len(result)

	// 金银铜卡
	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": startTime, "$lt": endTime},
				"ltype": 59,
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Diamond": bson.M{
					"$sum": "$add_diamond",
				},
				"Coin": bson.M{
					"$sum": "$add_coin",
				},
			},
		},
	}
	result1 := []bson.M{}
	pipe1 := LogWaters.Pipe(pipeline1)
	err = pipe1.All(&result1)
	if err != nil {
		beego.Error("BasicPayActivity fail err: ", err)
	}
	// ids := make([]string, 0)
	Diamond1 := int64(0)
	Coin1 := int64(0)
	for _, item := range result1 {
		// ids = append(ids, item["_id"].(string))
		Diamond1 += item["Diamond"].(int64)
		Coin1 += item["Coin"].(int64)
	}
	count1 := len(result1)

	// 首充礼包
	pipeline2 := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": startTime, "$lt": endTime},
				"ltype": 57,
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Diamond": bson.M{
					"$sum": "$add_diamond",
				},
				"Coin": bson.M{
					"$sum": "$add_coin",
				},
			},
		},
	}
	result2 := []bson.M{}
	pipe2 := LogWaters.Pipe(pipeline2)
	err = pipe2.All(&result2)
	if err != nil {
		beego.Error("BasicPayActivity fail err: ", err)
	}
	// ids := make([]string, 0)
	Diamond2 := int64(0)
	Coin2 := int64(0)
	for _, item := range result2 {
		// ids = append(ids, item["_id"].(string))
		Diamond2 += item["Diamond"].(int64)
		Coin2 += item["Coin"].(int64)
	}
	count2 := len(result2)
	// today := startTime.Format("2006-01-02")
	date1, _ := utils.Unix(fmt.Sprintf("%s", today))
	info := new(entity.BasicPayActivity)
	info.Date = date1
	info.ShopNumber = 0
	info.ShopDiamond = 0
	info.ShopCoin = 0
	info.RmNumber = int64(count)
	info.RmDiamond = ComputeFloat(Diamond, 100)
	info.RmCoin = ComputeFloat(Coin, 100)
	info.JyNumber = int64(count1)
	info.JyDiamond = ComputeFloat(Diamond1, 100)
	info.JyCoin = ComputeFloat(Coin1, 100)
	info.ScNumber = int64(count2)
	info.ScDiamond = ComputeFloat(Diamond2, 100)
	info.ScCoin = ComputeFloat(Coin2, 100)
	err1 := ActivityService.AddBasicPay(info)
	if err1 != nil {
		beego.Error("BasicPayActivity fail err: ", err1)
	}
}

/*
Author:CC
Title:活动管理 -> 基础付费活动
*/
func BasicPayActivityCK(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	s := fmt.Sprintf("%s 00:00:00", today)
	startTime := utils.Str2TimeZone(s, locationName)
	e := fmt.Sprintf("%s 23:59:59", today)
	endTime := utils.Str2TimeZone(e, locationName)

	var res []map[string]any
	sql1 := `select userid,shop_type,COUNT(1) number
	from game.col_trade_record ctr FINAL WHERE order_status = 4 and shop_type in (2,3,4) and ctime between ? AND ?
	GROUP BY userid,shop_type`
	var args1 []any
	args1 = append(args1, startTime)
	args1 = append(args1, endTime)
	err = ck.Select(&res, sql1, args1...)
	if err != nil {
		beego.Warning("BasicPayActivityCK error3:", err)
		return
	}

	// 人数
	count := 0
	count1 := 0
	count2 := 0
	// 彩金
	Diamond := int64(0)
	Diamond1 := int64(0)
	Diamond2 := int64(0)
	// 奖励金
	Coin := int64(0)
	Coin1 := int64(0)
	Coin2 := int64(0)

	if len(res) > 0 {
		for _, v := range res {
			// uid := v["userid"].(string)
			shop_type := v["shop_type"].(int32)
			switch shop_type {
			case 2:
				count++
			case 3:
				count1++
			case 4:
				count2++
			}
		}
	}
	// 产出彩金、奖励金
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": startTime, "$lt": endTime},
				"ltype": bson.M{"$in": []int{58, 59, 57}},
			},
		},
		{
			"$group": bson.M{
				"_id": "$ltype",
				"Diamond": bson.M{
					"$sum": "$add_diamond",
				},
				"Coin": bson.M{
					"$sum": "$add_coin",
				},
			},
		},
	}
	result := []bson.M{}
	pipe := LogWaters.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("BasicPayActivity fail err: ", err)
	}
	for _, item := range result {
		tpe := item["_id"].(int)
		switch tpe {
		case 57:
			// 首充
			count++
			Diamond2 += item["Diamond"].(int64)
			Coin2 += item["Coin"].(int64)
		case 58:
			// 入门礼包
			Diamond += item["Diamond"].(int64)
			Coin += item["Coin"].(int64)
		case 59:
			// 金银铜卡
			Diamond1 += item["Diamond"].(int64)
			Coin1 += item["Coin"].(int64)
		}
	}
	// today := startTime.Format("2006-01-02")
	date1, _ := utils.Unix(fmt.Sprintf("%s", today))
	info := new(entity.BasicPayActivity)
	info.Date = date1
	info.ShopNumber = 0
	info.ShopDiamond = 0
	info.ShopCoin = 0
	info.RmNumber = int64(count)
	info.RmDiamond = ComputeFloat(Diamond, 100)
	info.RmCoin = ComputeFloat(Coin, 100)
	info.JyNumber = int64(count1)
	info.JyDiamond = ComputeFloat(Diamond1, 100)
	info.JyCoin = ComputeFloat(Coin1, 100)
	info.ScNumber = int64(count2)
	info.ScDiamond = ComputeFloat(Diamond2, 100)
	info.ScCoin = ComputeFloat(Coin2, 100)
	err1 := ActivityService.AddBasicPay(info)
	if err1 != nil {
		beego.Error("BasicPayActivity fail err: ", err1)
	}
}

/*
Author:CC
Title:活动管理 -> 基础免费活动
*/
func BasicFreeActivity(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	s := fmt.Sprintf("%s 00:00:00", today)
	startTime := utils.Str2TimeZone(s, locationName)
	e := fmt.Sprintf("%s 23:59:59", today)
	endTime := utils.Str2TimeZone(e, locationName)
	// startTime := utils.Stamp2Time(utils.TimestampYesterday())
	// endTime := utils.Stamp2Time(utils.TimestampToday())
	// 每日签到
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": startTime, "$lt": endTime},
				"ltype": 61,
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Diamond": bson.M{
					"$sum": "$add_diamond",
				},
				"Coin": bson.M{
					"$sum": "$add_coin",
				},
			},
		},
	}
	result := []bson.M{}
	pipe := LogWaters.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("BasicFreeActivity fail err: ", err)
	}
	// ids := make([]string, 0)
	Diamond := int64(0)
	Coin := int64(0)
	for _, item := range result {
		// ids = append(ids, item["_id"].(string))
		Diamond += item["Diamond"].(int64)
		Coin += item["Coin"].(int64)
	}
	count := len(result)

	// 在线奖励
	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": startTime, "$lt": endTime},
				"ltype": 62,
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Diamond": bson.M{
					"$sum": "$add_diamond",
				},
				"Coin": bson.M{
					"$sum": "$add_coin",
				},
			},
		},
	}
	result1 := []bson.M{}
	pipe1 := LogWaters.Pipe(pipeline1)
	err = pipe1.All(&result1)
	if err != nil {
		beego.Error("BasicFreeActivity fail err: ", err)
	}
	// ids := make([]string, 0)
	Diamond1 := int64(0)
	Coin1 := int64(0)
	for _, item := range result1 {
		// ids = append(ids, item["_id"].(string))
		Diamond1 += item["Diamond"].(int64)
		Coin1 += item["Coin"].(int64)
	}
	count1 := len(result1)

	// 每日任务
	pipeline2 := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": startTime, "$lt": endTime},
				"ltype": 46,
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Diamond": bson.M{
					"$sum": "$add_diamond",
				},
				"Coin": bson.M{
					"$sum": "$add_coin",
				},
			},
		},
	}
	result2 := []bson.M{}
	pipe2 := LogWaters.Pipe(pipeline2)
	err = pipe2.All(&result2)
	if err != nil {
		beego.Error("BasicFreeActivity fail err: ", err)
	}
	// ids := make([]string, 0)
	Diamond2 := int64(0)
	Coin2 := int64(0)
	for _, item := range result2 {
		// ids = append(ids, item["_id"].(string))
		Diamond2 += item["Diamond"].(int64)
		Coin2 += item["Coin"].(int64)
	}
	count2 := len(result2)

	// 豹子牌型活动
	pipeline3 := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": startTime, "$lt": endTime},
				"ltype": 66,
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Diamond": bson.M{
					"$sum": "$add_diamond",
				},
				"Coin": bson.M{
					"$sum": "$add_coin",
				},
			},
		},
	}
	result3 := []bson.M{}
	pipe3 := LogWaters.Pipe(pipeline3)
	err = pipe3.All(&result3)
	if err != nil {
		beego.Error("BasicFreeActivity fail err: ", err)
	}
	// ids := make([]string, 0)
	Diamond3 := int64(0)
	Coin3 := int64(0)
	for _, item := range result3 {
		// ids = append(ids, item["_id"].(string))
		Diamond3 += item["Diamond"].(int64)
		Coin3 += item["Coin"].(int64)
	}
	count3 := len(result3)

	// today := startTime.Format("2006-01-02")
	date1, _ := utils.Unix(fmt.Sprintf("%s", today))
	info := new(entity.BasicFreeActivity)
	info.Date = date1
	info.SignInNumber = int64(count)
	info.SignInDiamond = ComputeFloat(Diamond, 100)
	info.SignInCoin = ComputeFloat(Coin, 100)
	info.OnlineNumber = int64(count1)
	info.OnlineDiamond = ComputeFloat(Diamond1, 100)
	info.OnlineCoin = ComputeFloat(Coin1, 100)
	info.TaskNumber = int64(count2)
	info.TaskDiamond = ComputeFloat(Diamond2, 100)
	info.TaskCoin = ComputeFloat(Coin2, 100)
	info.TrailOrSetNumber = int64(count3)
	info.TrailOrSetDiamond = ComputeFloat(Diamond3, 100)
	info.TrailOrSetCoin = ComputeFloat(Coin3, 100)

	err1 := ActivityService.AddBasicFree(info)
	if err1 != nil {
		beego.Error("BasicFreeActivity fail err: ", err1)
	}
}

/*
Title：用户留存
Effect：统计用户留存
*/
// Deprecated: 实时查询
func UserRetained(timestamp int64) {
	nowday := utils.Stamp2Time(timestamp)
	today := nowday.Format("2006-01-02")
	today1 := nowday.AddDate(0, 0, -1).Format("2006-01-02")
	today2 := nowday.AddDate(0, 0, -2).Format("2006-01-02")
	today3 := nowday.AddDate(0, 0, -3).Format("2006-01-02")
	today4 := nowday.AddDate(0, 0, -4).Format("2006-01-02")
	today5 := nowday.AddDate(0, 0, -5).Format("2006-01-02")
	today6 := nowday.AddDate(0, 0, -6).Format("2006-01-02")
	today7 := nowday.AddDate(0, 0, -7).Format("2006-01-02")
	today15 := nowday.AddDate(0, 0, -15).Format("2006-01-02")
	today30 := nowday.AddDate(0, 0, -30).Format("2006-01-02")
	today60 := nowday.AddDate(0, 0, -60).Format("2006-01-02")
	m := FindByDate(today, today, "ctime", "ctime")
	m["robot"] = false
	count, _ := PlayerService.GetTotal(m)

	// 1日留存
	regcount1, count1 := RetainedCompute4Web(today, today1)
	// 2日留存
	regcount2, count2 := RetainedCompute4Web(today, today2)
	// 3日留存
	regcount3, count3 := RetainedCompute4Web(today, today3)
	// 4日留存
	regcount4, count4 := RetainedCompute4Web(today, today4)
	// 5日留存
	regcount5, count5 := RetainedCompute4Web(today, today5)
	// 6日留存
	regcount6, count6 := RetainedCompute4Web(today, today6)
	// 7日留存
	regcount7, count7 := RetainedCompute4Web(today, today7)
	// 15日留存
	regcount15, count15 := RetainedCompute4Web(today, today15)
	// 30日留存
	regcount30, count30 := RetainedCompute4Web(today, today30)
	// 60日留存
	regcount60, count60 := RetainedCompute4Web(today, today60)

	day1 := ComputeFloat(count1, regcount1)
	day2 := ComputeFloat(count2, regcount2)
	day3 := ComputeFloat(count3, regcount3)
	day4 := ComputeFloat(count4, regcount4)
	day5 := ComputeFloat(count5, regcount5)
	day6 := ComputeFloat(count6, regcount6)
	day7 := ComputeFloat(count7, regcount7)
	day15 := ComputeFloat(count15, regcount15)
	day30 := ComputeFloat(count30, regcount30)
	day60 := ComputeFloat(count60, regcount60)
	info := new(entity.UserRetained4Web)
	date1, _ := utils.Unix(fmt.Sprintf("%s", today))
	info.Date = date1
	info.Channel = ""
	info.SType = 0
	info.NewNumber = int64(count)
	info.Day1 = day1 * 100
	info.Day2 = day2 * 100
	info.Day3 = day3 * 100
	info.Day4 = day4 * 100
	info.Day5 = day5 * 100
	info.Day6 = day6 * 100
	info.Day7 = day7 * 100
	info.Day15 = day15 * 100
	info.Day30 = day30 * 100
	info.Day60 = day60 * 100
	AddErr := StatisticsService.AddUserRetained(info)
	if AddErr != nil {
		beego.Error("UserRetained fail err: ", AddErr)
	}
}

/*
Title：用户留存
Effect：根据日期查询登录人数、注册人数
*/
func RetainedCompute4Web(today string, today1 string) (rcount int64, count int64) {
	regcount1 := int64(0)
	count1 := int64(0)
	m1 := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{
					"$gte": utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today1), locationName),
					"$lt":  utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today1), locationName)},
				"robot":            false,
				"simulation_robot": false,
			},
		},
		{
			"$group": bson.M{
				"_id": "$_id",
			},
		},
	}
	// operations := []bson.M{m, n}
	result := []bson.M{}
	pipe := PlayerUsers.Pipe(m1)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("RetainedCompute fail err: ", err)
	}
	ids := make([]string, 0)
	if len(result) > 0 {
		for _, item := range result {
			ids = append(ids, item["_id"].(string))
		}
		regcount1 = int64(len(result))
	}
	m1 = []bson.M{
		{
			"$match": bson.M{
				"login_time": bson.M{
					"$gte": utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName),
					"$lt":  utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)},
				"userid": bson.M{"$in": ids},
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
			},
		},
	}
	// operations := []bson.M{m, n}
	result = []bson.M{}
	pipe = LoginLogs.Pipe(m1)
	err = pipe.All(&result)
	if err != nil {
		beego.Error("RetainedCompute fail err: ", err)
	}
	if len(result) > 0 {
		uniqueData := make(map[string]bool)
		for _, item := range result {
			id := item["_id"].(string)
			uniqueData[id] = true
		}

		var list []string
		for key := range uniqueData {
			list = append(list, key)
		}
		count1 = int64(len(list))
	}

	return regcount1, count1
}

/*
Title：用户留存
Effect：按照渠道统计用户留存
*/
func UserRetainedByChannel(timestamp int64) {
	nowday := utils.Stamp2Time(timestamp)
	today := nowday.Format("2006-01-02")
	today1 := nowday.AddDate(0, 0, -1).Format("2006-01-02")
	today2 := nowday.AddDate(0, 0, -2).Format("2006-01-02")
	today3 := nowday.AddDate(0, 0, -3).Format("2006-01-02")
	today4 := nowday.AddDate(0, 0, -4).Format("2006-01-02")
	today5 := nowday.AddDate(0, 0, -5).Format("2006-01-02")
	today6 := nowday.AddDate(0, 0, -6).Format("2006-01-02")
	today7 := nowday.AddDate(0, 0, -7).Format("2006-01-02")
	today15 := nowday.AddDate(0, 0, -15).Format("2006-01-02")
	today30 := nowday.AddDate(0, 0, -30).Format("2006-01-02")
	today60 := nowday.AddDate(0, 0, -60).Format("2006-01-02")
	m := bson.M{"status": 1}
	var channellist []entity.ChannelInfo
	err := Channels.
		Find(m).
		All(&channellist)
	if err != nil {
		beego.Error("UserRetainedByChannel fail err: ", err)
	}
	if len(channellist) > 0 {
		date1, _ := utils.Unix(fmt.Sprintf("%s", today))
		retlist := make([]entity.UserRetained4Web, 0)
		for _, item := range channellist {
			channel := item.Name
			name1 := item.Name1
			pipeline := []bson.M{
				{
					"$match": bson.M{
						"ctime":            bson.M{"$gte": utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName), "$lt": utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)},
						"robot":            false,
						"simulation_robot": false,
						"ad__bundle_id":    channel,
					},
				},
				{
					"$group": bson.M{
						"_id": "$_id",
						"num": bson.M{
							"$sum": 1,
						},
					},
				},
			}
			result := []bson.M{}
			pipe := PlayerUsers.Pipe(pipeline)
			err := pipe.All(&result)
			if err != nil {
				beego.Error("UserRetainedByChannel fail err: ", err)
			}
			count := len(result)

			// 1日留存
			regcount1, count1 := RetainedComputeByChannel(today, today1, channel)
			// 2日留存
			regcount2, count2 := RetainedComputeByChannel(today, today2, channel)
			// 3日留存
			regcount3, count3 := RetainedComputeByChannel(today, today3, channel)
			// 4日留存
			regcount4, count4 := RetainedComputeByChannel(today, today4, channel)
			// 5日留存
			regcount5, count5 := RetainedComputeByChannel(today, today5, channel)
			// 6日留存
			regcount6, count6 := RetainedComputeByChannel(today, today6, channel)
			// 7日留存
			regcount7, count7 := RetainedComputeByChannel(today, today7, channel)
			// 15日留存
			regcount15, count15 := RetainedComputeByChannel(today, today15, channel)
			// 30日留存
			regcount30, count30 := RetainedComputeByChannel(today, today30, channel)
			// 60日留存
			regcount60, count60 := RetainedComputeByChannel(today, today60, channel)

			day1 := ComputeFloat(count1, regcount1)
			day2 := ComputeFloat(count2, regcount2)
			day3 := ComputeFloat(count3, regcount3)
			day4 := ComputeFloat(count4, regcount4)
			day5 := ComputeFloat(count5, regcount5)
			day6 := ComputeFloat(count6, regcount6)
			day7 := ComputeFloat(count7, regcount7)
			day15 := ComputeFloat(count15, regcount15)
			day30 := ComputeFloat(count30, regcount30)
			day60 := ComputeFloat(count60, regcount60)
			info := new(entity.UserRetained4Web)
			info.Date = date1
			info.SType = 1
			info.Channel = channel
			info.Channel1 = name1
			info.NewNumber = int64(count)
			info.Day1 = day1 * 100
			info.Day2 = day2 * 100
			info.Day3 = day3 * 100
			info.Day4 = day4 * 100
			info.Day5 = day5 * 100
			info.Day6 = day6 * 100
			info.Day7 = day7 * 100
			info.Day15 = day15 * 100
			info.Day30 = day30 * 100
			info.Day60 = day60 * 100
			retlist = append(retlist, *info)
		}
		for _, user := range retlist {
			AddErr := StatisticsService.AddUserRetained(&user)
			if AddErr != nil {
				beego.Error("UserRetainedByChannel fail err: ", AddErr)
			}
		}
	}
}

/*
Title：用户留存
Effect：根据日期、渠道查询登录人数、注册人数
*/
func RetainedComputeByChannel(today string, today1 string, channelId string) (rcount int64, count int64) {
	regcount1 := int64(0)
	count1 := int64(0)
	m1 := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{
					"$gte": utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today1), locationName),
					"$lt":  utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today1), locationName)},
				"ad__bundle_id":    channelId,
				"simulation_robot": false,
				"robot":            false,
			},
		},
		{
			"$group": bson.M{
				"_id": "$_id",
			},
		},
	}
	// operations := []bson.M{m, n}
	result := []bson.M{}
	pipe := PlayerUsers.Pipe(m1)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("RetainedComputeByChannel fail err: ", err)
	}
	ids := make([]string, 0)
	if len(result) > 0 {
		for _, item := range result {
			ids = append(ids, item["_id"].(string))
		}
		regcount1 = int64(len(result))
	}
	m1 = []bson.M{
		{
			"$match": bson.M{
				"login_time": bson.M{
					"$gte": utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName),
					"$lt":  utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)},
				"userid": bson.M{"$in": ids},
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
			},
		},
	}
	// operations := []bson.M{m, n}
	result = []bson.M{}
	pipe = LoginLogs.Pipe(m1)
	err = pipe.All(&result)
	if err != nil {
		beego.Error("RetainedComputeByChannel fail err: ", err)
	}
	if len(result) > 0 {
		uniqueData := make(map[string]bool)
		for _, item := range result {
			id := item["_id"].(string)
			uniqueData[id] = true
		}

		var list []string
		for key := range uniqueData {
			list = append(list, key)
		}
		count1 = int64(len(list))
	}

	return regcount1, count1
}

// 用户留存数据-新版本
func UserRetainedNew(timestamp int64) {
	nowday := utils.Stamp2Time(timestamp)
	today := nowday.Format("2006-01-02")
	today1 := nowday.AddDate(0, 0, -1).Format("2006-01-02")
	today2 := nowday.AddDate(0, 0, -2).Format("2006-01-02")
	today3 := nowday.AddDate(0, 0, -3).Format("2006-01-02")
	today4 := nowday.AddDate(0, 0, -4).Format("2006-01-02")
	today5 := nowday.AddDate(0, 0, -5).Format("2006-01-02")
	today6 := nowday.AddDate(0, 0, -6).Format("2006-01-02")
	today7 := nowday.AddDate(0, 0, -7).Format("2006-01-02")
	// today8 := nowday.AddDate(0, 0, -8).Format("2006-01-02")
	// today15 := nowday.AddDate(0, 0, -15).Format("2006-01-02")
	today15 := nowday.AddDate(0, 0, -15).Format("2006-01-02")
	// today30 := nowday.AddDate(0, 0, -30).Format("2006-01-02")
	today30 := nowday.AddDate(0, 0, -30).Format("2006-01-02")
	// today60 := nowday.AddDate(0, 0, -60).Format("2006-01-02")
	today60 := nowday.AddDate(0, 0, -60).Format("2006-01-02")

	m := bson.M{"status": 1}
	var channellist []entity.ChannelInfo
	err := Channels.
		Find(m).
		All(&channellist)
	if err != nil {
		beego.Error("UserRetainedNew fail err: ", err)
	}
	if len(channellist) > 0 {
		date1, _ := utils.Unix(fmt.Sprintf("%s", today))
		// date2, _ := utils.Unix(fmt.Sprintf("%s", today1))
		retlist := make([]entity.UserRetainedData, 0)
		for _, item := range channellist {
			channel := item.Name
			name1 := item.Name1
			// 当天注册
			var regIds []string
			reg_m := bson.M{
				"ctime":            bson.M{"$gte": utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName), "$lt": utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)},
				"robot":            false,
				"simulation_robot": false,
				"ad__bundle_id":    channel,
			}
			PlayerUsers.Find(reg_m).Distinct("_id", &regIds)
			// 计算当天新增人数
			info := new(entity.UserRetainedData)
			info.Date = date1
			info.Channel = channel
			info.Channel1 = name1
			info.NewNumber = int64(len(regIds))
			retlist = append(retlist, *info)

			// 获取当天登录用户ID
			l_m := bson.M{
				"login_time": bson.M{"$gte": utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName), "$lt": utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)},
			}
			var loginIds []string
			LoginLogs.Find(l_m).Distinct("userid", &loginIds)
			// 计算往天留存数据
			count1 := RetainedComputeNew(today1, loginIds, channel)
			count2 := RetainedComputeNew(today2, loginIds, channel)
			count3 := RetainedComputeNew(today3, loginIds, channel)
			count4 := RetainedComputeNew(today4, loginIds, channel)
			count5 := RetainedComputeNew(today5, loginIds, channel)
			count6 := RetainedComputeNew(today6, loginIds, channel)
			count7 := RetainedComputeNew(today7, loginIds, channel)
			count15 := RetainedComputeNew(today15, loginIds, channel)
			count30 := RetainedComputeNew(today30, loginIds, channel)
			count60 := RetainedComputeNew(today60, loginIds, channel)
			if count1 != 0 {
				// 往期存在待计算的数据
				datestamp, _ := utils.Unix(fmt.Sprintf("%s", today1))
				oldinfo := new(entity.UserRetainedData)
				oldinfo.Date = datestamp
				oldinfo.Channel = channel
				info1 := StatisticsService.GetUserRetainedNew(oldinfo)
				if info1.Id != "" {
					// 存在登录数据
					info1.Login1 = count1
				}
				retlist = append(retlist, info1)
			}
			if count2 != 0 {
				// 往期存在待计算的数据
				datestamp, _ := utils.Unix(fmt.Sprintf("%s", today2))
				oldinfo := new(entity.UserRetainedData)
				oldinfo.Date = datestamp
				oldinfo.Channel = channel
				info1 := StatisticsService.GetUserRetainedNew(oldinfo)
				if info1.Id != "" {
					// 存在登录数据
					info1.Login2 = count2
				}
				retlist = append(retlist, info1)
			}
			if count3 != 0 {
				// 往期存在待计算的数据
				datestamp, _ := utils.Unix(fmt.Sprintf("%s", today3))
				oldinfo := new(entity.UserRetainedData)
				oldinfo.Date = datestamp
				oldinfo.Channel = channel
				info1 := StatisticsService.GetUserRetainedNew(oldinfo)
				if info1.Id != "" {
					// 存在登录数据
					info1.Login3 = count3
				}
				retlist = append(retlist, info1)
			}
			if count4 != 0 {
				// 往期存在待计算的数据
				datestamp, _ := utils.Unix(fmt.Sprintf("%s", today4))
				oldinfo := new(entity.UserRetainedData)
				oldinfo.Date = datestamp
				oldinfo.Channel = channel
				info1 := StatisticsService.GetUserRetainedNew(oldinfo)
				if info1.Id != "" {
					// 存在登录数据
					info1.Login4 = count4
				}
				retlist = append(retlist, info1)
			}
			if count5 != 0 {
				// 往期存在待计算的数据
				datestamp, _ := utils.Unix(fmt.Sprintf("%s", today5))
				oldinfo := new(entity.UserRetainedData)
				oldinfo.Date = datestamp
				oldinfo.Channel = channel
				info1 := StatisticsService.GetUserRetainedNew(oldinfo)
				if info1.Id != "" {
					// 存在登录数据
					info1.Login5 = count5
				}
				retlist = append(retlist, info1)
			}
			if count6 != 0 {
				// 往期存在待计算的数据
				datestamp, _ := utils.Unix(fmt.Sprintf("%s", today6))
				oldinfo := new(entity.UserRetainedData)
				oldinfo.Date = datestamp
				oldinfo.Channel = channel
				info1 := StatisticsService.GetUserRetainedNew(oldinfo)
				if info1.Id != "" {
					// 存在登录数据
					info1.Login6 = count6
				}
				retlist = append(retlist, info1)
			}
			if count7 != 0 {
				// 往期存在待计算的数据
				datestamp, _ := utils.Unix(fmt.Sprintf("%s", today7))
				oldinfo := new(entity.UserRetainedData)
				oldinfo.Date = datestamp
				oldinfo.Channel = channel
				info1 := StatisticsService.GetUserRetainedNew(oldinfo)
				if info1.Id != "" {
					// 存在登录数据
					info1.Login7 = count7
				}
				retlist = append(retlist, info1)
			}
			if count15 != 0 {
				// 往期存在待计算的数据
				datestamp, _ := utils.Unix(fmt.Sprintf("%s", today15))
				oldinfo := new(entity.UserRetainedData)
				oldinfo.Date = datestamp
				oldinfo.Channel = channel
				info1 := StatisticsService.GetUserRetainedNew(oldinfo)
				if info1.Id != "" {
					// 存在登录数据
					info1.Login15 = count15
				}
				retlist = append(retlist, info1)
			}
			if count30 != 0 {
				// 往期存在待计算的数据
				datestamp, _ := utils.Unix(fmt.Sprintf("%s", today30))
				oldinfo := new(entity.UserRetainedData)
				oldinfo.Date = datestamp
				oldinfo.Channel = channel
				info1 := StatisticsService.GetUserRetainedNew(oldinfo)
				if info1.Id != "" {
					// 存在登录数据
					info1.Login30 = count30
				}
				retlist = append(retlist, info1)
			}
			if count60 != 0 {
				// 往期存在待计算的数据
				datestamp, _ := utils.Unix(fmt.Sprintf("%s", today60))
				oldinfo := new(entity.UserRetainedData)
				oldinfo.Date = datestamp
				oldinfo.Channel = channel
				info1 := StatisticsService.GetUserRetainedNew(oldinfo)
				if info1.Id != "" {
					// 存在登录数据
					info1.Login60 = count60
				}
				retlist = append(retlist, info1)
			}

			// // 前一天注册
			// var regIds1 []string
			// reg_m1 := bson.M{
			// 	"ctime":            bson.M{"$gte": utils.Str2Time(fmt.Sprintf("%s 00:00:00", today1)), "$lt": utils.Str2Time(fmt.Sprintf("%s 23:59:59", today1))},
			// 	"robot":            false,
			// 	"simulation_robot": false,
			// 	"ad__bundle_id":    channel,
			// }

			// PlayerUsers.Find(reg_m1).Distinct("_id", &regIds1)
			// // 计算前一天，用户留存

			// // 1日留存
			// regcount1, count1 := RetainedComputeByChannel(today1, today2, channel)
			// // 2日留存
			// regcount2, count2 := RetainedComputeByChannel(today1, today3, channel)
			// // 3日留存
			// regcount3, count3 := RetainedComputeByChannel(today1, today4, channel)
			// // 4日留存
			// regcount4, count4 := RetainedComputeByChannel(today1, today5, channel)
			// // 5日留存
			// regcount5, count5 := RetainedComputeByChannel(today1, today6, channel)
			// // 6日留存
			// regcount6, count6 := RetainedComputeByChannel(today1, today7, channel)
			// // 7日留存
			// regcount7, count7 := RetainedComputeByChannel(today1, today8, channel)
			// // 15日留存
			// regcount15, count15 := RetainedComputeByChannel(today1, today16, channel)
			// // 30日留存
			// regcount30, count30 := RetainedComputeByChannel(today1, today31, channel)
			// // 60日留存
			// regcount60, count60 := RetainedComputeByChannel(today1, today61, channel)

			// info1 := new(entity.UserRetainedData)
			// info1.Date = date2
			// info1.Channel = channel
			// info1.Channel1 = name1
			// info1.NewNumber = int64(len(regIds1))
			// info1.Login1 = count1
			// info1.Register1 = regcount1
			// info1.Login2 = count2
			// info1.Register2 = regcount2
			// info1.Login3 = count3
			// info1.Register3 = regcount3
			// info1.Login4 = count4
			// info1.Register4 = regcount4
			// info1.Login5 = count5
			// info1.Register5 = regcount5
			// info1.Login6 = count6
			// info1.Register6 = regcount6
			// info1.Login7 = count7
			// info1.Register7 = regcount7
			// info1.Login15 = count15
			// info1.Register15 = regcount15
			// info1.Login30 = count30
			// info1.Register30 = regcount30
			// info1.Login60 = count60
			// info1.Register60 = regcount60
			// retlist = append(retlist, *info1)
		}
		for _, user := range retlist {
			AddErr := StatisticsService.AddUserRetainedNew(&user)
			if AddErr != nil {
				beego.Error("UserRetainedNew fail err: ", AddErr)
			}
		}
	}
}

func RetainedComputeNew(today string, logids []string, channelId string) (rcount int64) {
	regcount := 0
	var regIds []string
	reg_m := bson.M{
		"ctime":            bson.M{"$gte": utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName), "$lt": utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)},
		"robot":            false,
		"simulation_robot": false,
		"ad__bundle_id":    channelId,
		"_id":              bson.M{"$in": logids},
	}
	PlayerUsers.Find(reg_m).Distinct("_id", &regIds)
	regcount = len(regIds)
	return int64(regcount)
}

// Deprecated: 实时查询
func UserRetainedCK(timestamp int64) {
	nowday := utils.Stamp2Time(timestamp)
	today := nowday.Format("2006-01-02")
	today1 := nowday.AddDate(0, 0, -1).Format("2006-01-02")
	today2 := nowday.AddDate(0, 0, -2).Format("2006-01-02")
	today3 := nowday.AddDate(0, 0, -3).Format("2006-01-02")
	today4 := nowday.AddDate(0, 0, -4).Format("2006-01-02")
	today5 := nowday.AddDate(0, 0, -5).Format("2006-01-02")
	today6 := nowday.AddDate(0, 0, -6).Format("2006-01-02")
	today7 := nowday.AddDate(0, 0, -7).Format("2006-01-02")
	today15 := nowday.AddDate(0, 0, -15).Format("2006-01-02")
	today30 := nowday.AddDate(0, 0, -30).Format("2006-01-02")
	today60 := nowday.AddDate(0, 0, -60).Format("2006-01-02")

	m := bson.M{"status": 1}
	var channellist []entity.ChannelInfo
	err := Channels.
		Find(m).
		All(&channellist)
	if err != nil {
		beego.Error("UserRetainedNew fail err: ", err)
	}
	s1 := fmt.Sprintf("%s 00:00:00", today)
	e1 := fmt.Sprintf("%s 23:59:59", today)
	var login_ids []string
	var args0 []any
	sql0 := `SELECT DISTINCT userid FROM game.col_log_login cll FINAL 
	WHERE login_time >= toDateTime(?, ?) AND login_time <= toDateTime(?, ?)
	`
	args0 = append(args0, s1)
	args0 = append(args0, locationName)
	args0 = append(args0, e1)
	args0 = append(args0, locationName)
	err = ck.Select(&login_ids, sql0, args0...)
	if err != nil {
		beego.Error("UserRetainedCK fail err: ", err)
	}
	retlist := make([]entity.UserRetainedData, 0)
	// 当天用户注册数统计
	var new_list []map[string]any
	var args1 []any
	sql1 := `SELECT ad__bundle_id,COUNT(1) num FROM game.col_user cu FINAL 
	WHERE robot = 0 AND simulation_robot = 0 AND ctime >= toDateTime(?, ?) AND ctime <= toDateTime(?, ?)
	GROUP BY ad__bundle_id
	`
	args1 = append(args1, s1)
	args1 = append(args1, locationName)
	args1 = append(args1, e1)
	args1 = append(args1, locationName)
	err = ck.Select(&new_list, sql1, args1...)
	if err != nil {
		beego.Error("UserRetainedCK fail err: ", err)
	}
	date1, _ := utils.Unix(fmt.Sprintf("%s", today))
	for _, v := range new_list {
		channel := v["ad__bundle_id"].(string)
		if channel != "" {
			num := v["num"].(uint64)
			info := new(entity.UserRetainedData)
			info.Date = date1
			info.Channel = channel
			info.NewNumber = int64(num)
			retlist = append(retlist, *info)
		}
	}
	for k, item := range retlist {
		for _, p := range channellist {
			if item.Channel == p.Name {
				item.Channel1 = p.Name1
				continue
			}
		}
		retlist[k] = item
	}
	tab1 := RetainedComputeCK(today1, login_ids)
	if len(tab1) > 0 {
		for _, p := range tab1 {
			channel := p["ad__bundle_id"].(string)
			if channel != "" {
				num := utils.ToInt64(p["num"])
				lognum := utils.ToInt64(p["log_num"])
				// 往期存在待计算的数据
				datestamp, _ := utils.Unix(fmt.Sprintf("%s", today1))
				oldinfo := new(entity.UserRetainedData)
				oldinfo.Date = datestamp
				oldinfo.Channel = channel
				info1 := StatisticsService.GetUserRetainedNew(oldinfo)
				if info1.Id != "" {
					// 存在登录数据
					info1.Login1 = lognum
					info1.NewNumber = num
					retlist = append(retlist, info1)
				}
			}
		}
	}
	tab2 := RetainedComputeCK(today2, login_ids)
	if len(tab2) > 0 {
		for _, p := range tab2 {
			channel := p["ad__bundle_id"].(string)
			if channel != "" {
				num := utils.ToInt64(p["num"])
				lognum := utils.ToInt64(p["log_num"])
				// 往期存在待计算的数据
				datestamp, _ := utils.Unix(fmt.Sprintf("%s", today2))
				oldinfo := new(entity.UserRetainedData)
				oldinfo.Date = datestamp
				oldinfo.Channel = channel
				info1 := StatisticsService.GetUserRetainedNew(oldinfo)
				if info1.Id != "" {
					// 存在登录数据
					info1.Login2 = lognum
					info1.NewNumber = num
					retlist = append(retlist, info1)
				}

			}
		}
	}
	tab3 := RetainedComputeCK(today3, login_ids)
	if len(tab3) > 0 {
		for _, p := range tab3 {
			channel := p["ad__bundle_id"].(string)
			if channel != "" {
				num := utils.ToInt64(p["num"])
				lognum := utils.ToInt64(p["log_num"])
				// 往期存在待计算的数据
				datestamp, _ := utils.Unix(fmt.Sprintf("%s", today3))
				oldinfo := new(entity.UserRetainedData)
				oldinfo.Date = datestamp
				oldinfo.Channel = channel
				info1 := StatisticsService.GetUserRetainedNew(oldinfo)
				if info1.Id != "" {
					// 存在登录数据
					info1.Login3 = lognum
					info1.NewNumber = num
					retlist = append(retlist, info1)
				}

			}
		}
	}
	tab4 := RetainedComputeCK(today4, login_ids)
	if len(tab4) > 0 {
		for _, p := range tab4 {
			channel := p["ad__bundle_id"].(string)
			if channel != "" {
				num := utils.ToInt64(p["num"])
				lognum := utils.ToInt64(p["log_num"])
				// 往期存在待计算的数据
				datestamp, _ := utils.Unix(fmt.Sprintf("%s", today4))
				oldinfo := new(entity.UserRetainedData)
				oldinfo.Date = datestamp
				oldinfo.Channel = channel
				info1 := StatisticsService.GetUserRetainedNew(oldinfo)
				if info1.Id != "" {
					// 存在登录数据
					info1.Login4 = lognum
					info1.NewNumber = num
					retlist = append(retlist, info1)
				}

			}
		}
	}
	tab5 := RetainedComputeCK(today5, login_ids)
	if len(tab5) > 0 {
		for _, p := range tab5 {
			channel := p["ad__bundle_id"].(string)
			if channel != "" {
				num := utils.ToInt64(p["num"])
				lognum := utils.ToInt64(p["log_num"])
				// 往期存在待计算的数据
				datestamp, _ := utils.Unix(fmt.Sprintf("%s", today5))
				oldinfo := new(entity.UserRetainedData)
				oldinfo.Date = datestamp
				oldinfo.Channel = channel
				info1 := StatisticsService.GetUserRetainedNew(oldinfo)
				if info1.Id != "" {
					// 存在登录数据
					info1.Login5 = lognum
					info1.NewNumber = num
					retlist = append(retlist, info1)
				}

			}
		}
	}
	tab6 := RetainedComputeCK(today6, login_ids)
	if len(tab6) > 0 {
		for _, p := range tab6 {
			channel := p["ad__bundle_id"].(string)
			if channel != "" {
				num := utils.ToInt64(p["num"])
				lognum := utils.ToInt64(p["log_num"])
				// 往期存在待计算的数据
				datestamp, _ := utils.Unix(fmt.Sprintf("%s", today6))
				oldinfo := new(entity.UserRetainedData)
				oldinfo.Date = datestamp
				oldinfo.Channel = channel
				info1 := StatisticsService.GetUserRetainedNew(oldinfo)
				if info1.Id != "" {
					// 存在登录数据
					info1.Login6 = lognum
					info1.NewNumber = num
					retlist = append(retlist, info1)
				}

			}
		}
	}
	tab7 := RetainedComputeCK(today7, login_ids)
	if len(tab7) > 0 {
		for _, p := range tab7 {
			channel := p["ad__bundle_id"].(string)
			if channel != "" {
				num := utils.ToInt64(p["num"])
				lognum := utils.ToInt64(p["log_num"])
				// 往期存在待计算的数据
				datestamp, _ := utils.Unix(fmt.Sprintf("%s", today7))
				oldinfo := new(entity.UserRetainedData)
				oldinfo.Date = datestamp
				oldinfo.Channel = channel
				info1 := StatisticsService.GetUserRetainedNew(oldinfo)
				if info1.Id != "" {
					// 存在登录数据
					info1.Login7 = lognum
					info1.NewNumber = num
					retlist = append(retlist, info1)
				}

			}
		}
	}
	tab15 := RetainedComputeCK(today15, login_ids)
	if len(tab15) > 0 {
		for _, p := range tab15 {
			channel := p["ad__bundle_id"].(string)
			if channel != "" {
				num := utils.ToInt64(p["num"])
				lognum := utils.ToInt64(p["log_num"])
				// 往期存在待计算的数据
				datestamp, _ := utils.Unix(fmt.Sprintf("%s", today15))
				oldinfo := new(entity.UserRetainedData)
				oldinfo.Date = datestamp
				oldinfo.Channel = channel
				info1 := StatisticsService.GetUserRetainedNew(oldinfo)
				if info1.Id != "" {
					// 存在登录数据
					info1.Login15 = lognum
					info1.NewNumber = num
					retlist = append(retlist, info1)
				}

			}
		}
	}
	tab30 := RetainedComputeCK(today30, login_ids)
	if len(tab30) > 0 {
		for _, p := range tab30 {
			channel := p["ad__bundle_id"].(string)
			if channel != "" {
				num := utils.ToInt64(p["num"])
				lognum := utils.ToInt64(p["log_num"])
				// 往期存在待计算的数据
				datestamp, _ := utils.Unix(fmt.Sprintf("%s", today30))
				oldinfo := new(entity.UserRetainedData)
				oldinfo.Date = datestamp
				oldinfo.Channel = channel
				info1 := StatisticsService.GetUserRetainedNew(oldinfo)
				if info1.Id != "" {
					// 存在登录数据
					info1.Login30 = lognum
					info1.NewNumber = num
					retlist = append(retlist, info1)
				}

			}
		}
	}
	tab60 := RetainedComputeCK(today60, login_ids)
	if len(tab60) > 0 {
		for _, p := range tab60 {
			channel := p["ad__bundle_id"].(string)
			if channel != "" {
				num := utils.ToInt64(p["num"])
				lognum := utils.ToInt64(p["log_num"])
				// 往期存在待计算的数据
				datestamp, _ := utils.Unix(fmt.Sprintf("%s", today60))
				oldinfo := new(entity.UserRetainedData)
				oldinfo.Date = datestamp
				oldinfo.Channel = channel
				info1 := StatisticsService.GetUserRetainedNew(oldinfo)
				if info1.Id != "" {
					// 存在登录数据
					info1.Login60 = lognum
					info1.NewNumber = num
					retlist = append(retlist, info1)
				}
			}
		}
	}
	for _, user := range retlist {
		AddErr := StatisticsService.AddUserRetainedNew(&user)
		if AddErr != nil {
			beego.Error("UserRetainedNew fail err: ", AddErr)
		}
	}
}

func RetainedComputeCK(today string, logids []string) []map[string]any {
	s1 := fmt.Sprintf("%s 00:00:00", today)
	e1 := fmt.Sprintf("%s 23:59:59", today)
	sql0 := `SELECT s1.ad__bundle_id,s1.num,s2.log_num FROM (SELECT ad__bundle_id,COUNT(1) num FROM  game.col_user cu FINAL 
	WHERE robot = 0 AND simulation_robot = 0 AND ctime >= toDateTime(?, ?) AND ctime <= toDateTime(?, ?)
	GROUP BY ad__bundle_id) s1
	LEFT JOIN 
	(SELECT ad__bundle_id,COUNT(1) log_num FROM game.col_user cu FINAL 
	WHERE robot = 0 AND simulation_robot = 0 AND ctime >= toDateTime(?, ?) AND ctime <= toDateTime(?, ?) AND userid IN ?
	GROUP BY ad__bundle_id) s2 ON s1.ad__bundle_id = s2.ad__bundle_id
	`
	var list []map[string]any
	var args0 []any
	args0 = append(args0, s1)
	args0 = append(args0, locationName)
	args0 = append(args0, e1)
	args0 = append(args0, locationName)
	args0 = append(args0, s1)
	args0 = append(args0, locationName)
	args0 = append(args0, e1)
	args0 = append(args0, locationName)
	args0 = append(args0, logids)
	err := ck.Select(&list, sql0, args0...)
	if err != nil {
		beego.Error("RetainedComputeCK fail err: ", err)
	}
	return list
}

func PayUserRetainedNew(timestamp int64) {
	nowday := utils.Stamp2Time(timestamp)
	today := nowday.Format("2006-01-02")
	today1 := nowday.AddDate(0, 0, -1).Format("2006-01-02")
	today2 := nowday.AddDate(0, 0, -2).Format("2006-01-02")
	today3 := nowday.AddDate(0, 0, -3).Format("2006-01-02")
	today4 := nowday.AddDate(0, 0, -4).Format("2006-01-02")
	today5 := nowday.AddDate(0, 0, -5).Format("2006-01-02")
	today6 := nowday.AddDate(0, 0, -6).Format("2006-01-02")
	today7 := nowday.AddDate(0, 0, -7).Format("2006-01-02")
	// today8 := nowday.AddDate(0, 0, -8).Format("2006-01-02")
	// today15 := nowday.AddDate(0, 0, -15).Format("2006-01-02")
	today15 := nowday.AddDate(0, 0, -15).Format("2006-01-02")
	// today30 := nowday.AddDate(0, 0, -30).Format("2006-01-02")
	today30 := nowday.AddDate(0, 0, -30).Format("2006-01-02")
	// today60 := nowday.AddDate(0, 0, -60).Format("2006-01-02")
	today60 := nowday.AddDate(0, 0, -60).Format("2006-01-02")
	m := bson.M{"status": 1}
	var channellist []entity.ChannelInfo
	err := Channels.
		Find(m).
		All(&channellist)
	if err != nil {
		beego.Error("PayUserRetainedNew fail err: ", err)
	}
	// 获取复充用户ID
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"order_status": 4,
			},
		},
		{
			"$group": bson.M{
				"_id":   "$userid",
				"count": bson.M{"$sum": 1}, // 添加一个额外的字段以便后续过滤
			},
		},
		{
			"$match": bson.M{
				"count": bson.M{"$gte": 2}, // 过滤num值在1文档
			},
		},
	}
	var res []bson.M
	Pays.Pipe(pipeline).All(&res)
	fc_pay_id := make([]string, 0)
	for _, item := range res {
		pid := item["_id"].(string)
		fc_pay_id = append(fc_pay_id, pid)
	}
	if len(channellist) > 0 {
		date1, _ := utils.Unix(fmt.Sprintf("%s", today))
		// date2, _ := utils.Unix(fmt.Sprintf("%s", today1))
		retlist := make([]entity.PayUserRetainedData, 0)
		pay_type := []int{0, 1}
		for _, item := range channellist {
			channel := item.Name
			name1 := item.Name1

			for _, v := range pay_type {
				var ids []string // 充值ID
				m := FindByDate(today, today, "ctime", "ctime")
				m["order_status"] = 4
				if v == 1 {
					m["userid"] = bson.M{"$in": fc_pay_id}
				}
				m["package_id"] = channel
				Pays.Find(m).Distinct("userid", &ids)
				if len(ids) <= 0 {
					continue
				}
				info := new(entity.PayUserRetainedData)
				info.Date = date1
				info.PayUserType = v
				info.Channel = channel
				info.Channel1 = name1
				info.NewNumber = int64(len(ids))
				retlist = append(retlist, *info)

				// 获取当天登陆用户
				m1 := FindByDate(today, today, "login_time", "login_time")
				if v == 1 {
					m1["userid"] = bson.M{"$in": fc_pay_id}
				}
				// if v == 0 {
				// 	m1["userid"] = bson.M{"$in": first_pay_id}
				// } else {
				// 	m1["userid"] = bson.M{"$nin": first_pay_id}
				// }
				var logids []string
				LoginLogs.Find(m1).Distinct("userid", &logids)
				// lCount = int64(len(logids))

				// // 前一天
				// var ids1 []string // 充值ID
				// m1 := FindByDate(today1, today1, "ctime", "ctime")
				// m1["order_status"] = 4
				// if v == 0 {
				// 	m1["userid"] = bson.M{"$in": first_pay_id}
				// } else {
				// 	m1["userid"] = bson.M{"$nin": first_pay_id}
				// }
				// m1["package_id"] = channel
				// Pays.Find(m1).Distinct("_id", &ids1)
				// // 1日留存
				count1 := PayRetainedComputeNew(today1, logids, channel)
				// 2日留存
				count2 := PayRetainedComputeNew(today2, logids, channel)
				// 3日留存
				count3 := PayRetainedComputeNew(today3, logids, channel)
				// 4日留存
				count4 := PayRetainedComputeNew(today4, logids, channel)
				// 5日留存
				count5 := PayRetainedComputeNew(today5, logids, channel)
				// 6日留存
				count6 := PayRetainedComputeNew(today6, logids, channel)
				// 7日留存
				count7 := PayRetainedComputeNew(today7, logids, channel)
				// 15日留存
				count15 := PayRetainedComputeNew(today15, logids, channel)
				// 30日留存
				count30 := PayRetainedComputeNew(today30, logids, channel)
				// 60日留存
				count60 := PayRetainedComputeNew(today60, logids, channel)
				if count1 != 0 {
					// 往期存在待计算的数据
					datestamp, _ := utils.Unix(fmt.Sprintf("%s", today1))
					oldinfo := new(entity.PayUserRetainedData)
					oldinfo.Date = datestamp
					oldinfo.Channel = channel
					oldinfo.PayUserType = v
					info1 := StatisticsService.GetPayUserRetainedNew(oldinfo)
					if info1.Id != "" {
						// 存在登录数据
						info1.Pay1 = count1
					}
					retlist = append(retlist, info1)
				}
				if count2 != 0 {
					// 往期存在待计算的数据
					datestamp, _ := utils.Unix(fmt.Sprintf("%s", today2))
					oldinfo := new(entity.PayUserRetainedData)
					oldinfo.Date = datestamp
					oldinfo.Channel = channel
					oldinfo.PayUserType = v
					info1 := StatisticsService.GetPayUserRetainedNew(oldinfo)
					if info1.Id != "" {
						// 存在登录数据
						info1.Pay2 = count2
					}
					retlist = append(retlist, info1)
				}
				if count3 != 0 {
					// 往期存在待计算的数据
					datestamp, _ := utils.Unix(fmt.Sprintf("%s", today3))
					oldinfo := new(entity.PayUserRetainedData)
					oldinfo.Date = datestamp
					oldinfo.Channel = channel
					oldinfo.PayUserType = v
					info1 := StatisticsService.GetPayUserRetainedNew(oldinfo)
					if info1.Id != "" {
						// 存在登录数据
						info1.Pay3 = count3
					}
					retlist = append(retlist, info1)
				}
				if count4 != 0 {
					// 往期存在待计算的数据
					datestamp, _ := utils.Unix(fmt.Sprintf("%s", today4))
					oldinfo := new(entity.PayUserRetainedData)
					oldinfo.Date = datestamp
					oldinfo.Channel = channel
					oldinfo.PayUserType = v
					info1 := StatisticsService.GetPayUserRetainedNew(oldinfo)
					if info1.Id != "" {
						// 存在登录数据
						info1.Pay4 = count4
					}
					retlist = append(retlist, info1)
				}
				if count5 != 0 {
					// 往期存在待计算的数据
					datestamp, _ := utils.Unix(fmt.Sprintf("%s", today5))
					oldinfo := new(entity.PayUserRetainedData)
					oldinfo.Date = datestamp
					oldinfo.Channel = channel
					oldinfo.PayUserType = v
					info1 := StatisticsService.GetPayUserRetainedNew(oldinfo)
					if info1.Id != "" {
						// 存在登录数据
						info1.Pay5 = count5
					}
					retlist = append(retlist, info1)
				}
				if count6 != 0 {
					// 往期存在待计算的数据
					datestamp, _ := utils.Unix(fmt.Sprintf("%s", today6))
					oldinfo := new(entity.PayUserRetainedData)
					oldinfo.Date = datestamp
					oldinfo.Channel = channel
					oldinfo.PayUserType = v
					info1 := StatisticsService.GetPayUserRetainedNew(oldinfo)
					if info1.Id != "" {
						// 存在登录数据
						info1.Pay6 = count6
					}
					retlist = append(retlist, info1)
				}
				if count7 != 0 {
					// 往期存在待计算的数据
					datestamp, _ := utils.Unix(fmt.Sprintf("%s", today7))
					oldinfo := new(entity.PayUserRetainedData)
					oldinfo.Date = datestamp
					oldinfo.Channel = channel
					oldinfo.PayUserType = v
					info1 := StatisticsService.GetPayUserRetainedNew(oldinfo)
					if info1.Id != "" {
						// 存在登录数据
						info1.Pay7 = count7
					}
					retlist = append(retlist, info1)
				}
				if count15 != 0 {
					// 往期存在待计算的数据
					datestamp, _ := utils.Unix(fmt.Sprintf("%s", today15))
					oldinfo := new(entity.PayUserRetainedData)
					oldinfo.Date = datestamp
					oldinfo.Channel = channel
					oldinfo.PayUserType = v
					info1 := StatisticsService.GetPayUserRetainedNew(oldinfo)
					if info1.Id != "" {
						// 存在登录数据
						info1.Pay15 = count15
					}
					retlist = append(retlist, info1)
				}
				if count30 != 0 {
					// 往期存在待计算的数据
					datestamp, _ := utils.Unix(fmt.Sprintf("%s", today30))
					oldinfo := new(entity.PayUserRetainedData)
					oldinfo.Date = datestamp
					oldinfo.Channel = channel
					info1 := StatisticsService.GetPayUserRetainedNew(oldinfo)
					if info1.Id != "" {
						// 存在登录数据
						info1.Pay30 = count30
					}
					retlist = append(retlist, info1)
				}
				if count60 != 0 {
					// 往期存在待计算的数据
					datestamp, _ := utils.Unix(fmt.Sprintf("%s", today60))
					oldinfo := new(entity.PayUserRetainedData)
					oldinfo.Date = datestamp
					oldinfo.Channel = channel
					info1 := StatisticsService.GetPayUserRetainedNew(oldinfo)
					if info1.Id != "" {
						// 存在登录数据
						info1.Pay60 = count60
					}
					retlist = append(retlist, info1)
				}

				// info1 := new(entity.PayUserRetainedData)
				// info1.Date = date2
				// info1.PayUserType = v
				// info1.Channel = channel
				// info1.Channel1 = name1
				// info1.NewNumber = int64(len(ids1))
				// info1.Pay1 = count1
				// info1.Register1 = regcount1
				// info1.Pay2 = count2
				// info1.Register2 = regcount2
				// info1.Pay3 = count3
				// info1.Register3 = regcount3
				// info1.Pay4 = count4
				// info1.Register4 = regcount4
				// info1.Pay5 = count5
				// info1.Register5 = regcount5
				// info1.Pay6 = count6
				// info1.Register6 = regcount6
				// info1.Pay7 = count7
				// info1.Register7 = regcount7
				// info1.Pay15 = count15
				// info1.Register15 = regcount15
				// info1.Pay30 = count30
				// info1.Register30 = regcount30
				// info1.Pay60 = count60
				// info1.Register60 = regcount60
				// retlist = append(retlist, *info1)
			}
		}
		for _, user := range retlist {
			AddErr := StatisticsService.AddPayUserRetainedNew(&user)
			if AddErr != nil {
				beego.Error("PayUserRetainedNew fail err: ", AddErr)
			}
		}
	}
}

func PayRetainedComputeNew(today string, logids []string, channelId string) (count int64) {
	// 获取昨日注册，且充值玩家 今日登录
	zrCount1 := 0 // 之前充值，昨日登录人数

	// 获取总充值人数
	m := FindByDate(today, today, "ctime", "ctime")
	m["order_status"] = 4
	m["userid"] = bson.M{"$in": logids}
	if channelId != "" {
		m["package_id"] = channelId
	}
	var cids []string
	Pays.Find(m).Distinct("userid", &cids)
	zrCount1 = len(cids)
	return int64(zrCount1)
}

func PayUserRetainedCK(timestamp int64) {
	nowday := utils.Stamp2Time(timestamp)
	today := nowday.Format("2006-01-02")
	today1 := nowday.AddDate(0, 0, -1).Format("2006-01-02")
	today2 := nowday.AddDate(0, 0, -2).Format("2006-01-02")
	today3 := nowday.AddDate(0, 0, -3).Format("2006-01-02")
	today4 := nowday.AddDate(0, 0, -4).Format("2006-01-02")
	today5 := nowday.AddDate(0, 0, -5).Format("2006-01-02")
	today6 := nowday.AddDate(0, 0, -6).Format("2006-01-02")
	today7 := nowday.AddDate(0, 0, -7).Format("2006-01-02")
	today15 := nowday.AddDate(0, 0, -15).Format("2006-01-02")
	today30 := nowday.AddDate(0, 0, -30).Format("2006-01-02")
	today60 := nowday.AddDate(0, 0, -60).Format("2006-01-02")
	m := bson.M{"status": 1}
	var channellist []entity.ChannelInfo
	err := Channels.
		Find(m).
		All(&channellist)
	if err != nil {
		beego.Error("PayUserRetainedCK fail err: ", err)
	}
	s1 := fmt.Sprintf("%s 00:00:00", today)
	e1 := fmt.Sprintf("%s 23:59:59", today)
	// 获取当天登陆用户
	sql3 := `SELECT DISTINCT userid FROM game.col_log_login cll FINAL 
	WHERE login_time >= toDateTime(?, ?) AND login_time <= toDateTime(?, ?) 
	`
	var login_ids []string
	var args3 []any
	args3 = append(args3, s1)
	args3 = append(args3, locationName)
	args3 = append(args3, e1)
	args3 = append(args3, locationName)
	err = ck.Select(&login_ids, sql3, args3...)
	if err != nil {
		beego.Error("PayUserRetainedCK fail err: ", err)
	}
	var fc_user []string
	var args0 []any
	sql0 := `SELECT DISTINCT userid FROM (SELECT userid,COUNT(1) AS num FROM game.col_trade_record ctr FINAL 
	WHERE order_status = 4 AND userid IN ?
	GROUP BY userid) tab WHERE num >= 2
	`
	args0 = append(args0, login_ids)
	err = ck.Select(&fc_user, sql0, args0...)
	if err != nil {
		beego.Error("PayUserRetainedCK fail err: ", err)
	}
	if len(channellist) > 0 {
		date1, _ := utils.Unix(fmt.Sprintf("%s", today))
		retlist := make([]entity.PayUserRetainedData, 0)
		pay_type := []int{0, 1}
		for _, item := range channellist {
			channel := item.Name
			name1 := item.Name1
			for _, v := range pay_type {
				var ids []string // 充值ID
				sql1 := `SELECT DISTINCT userid from game.col_trade_record ctr FINAL 
				WHERE order_status = 4 AND package_id = ? AND ctime >= toDateTime(?, ?) AND ctime <= toDateTime(?, ?) 
		`
				var args1 []any
				args1 = append(args1, channel)
				args1 = append(args1, s1)
				args1 = append(args1, locationName)
				args1 = append(args1, e1)
				args1 = append(args1, locationName)
				if v == 1 {
					// 复充用户
					sql1 = `SELECT DISTINCT userid from game.col_trade_record ctr FINAL 
					WHERE order_status = 4 AND package_id = ? AND ctime >= toDateTime(?, ?) AND ctime <= toDateTime(?, ?) and userid in ?
			`
					args1 = append(args1, fc_user)
				}
				err = ck.Select(&ids, sql1, args1...)
				if err != nil {
					beego.Error("PayUserRetainedCK fail err: ", err)
				}
				if len(ids) <= 0 {
					continue
				}
				info := new(entity.PayUserRetainedData)
				info.Date = date1
				info.PayUserType = v
				info.Channel = channel
				info.Channel1 = name1
				info.NewNumber = int64(len(ids))
				retlist = append(retlist, *info)

				// 获取当天登陆用户
				sql2 := `SELECT DISTINCT userid FROM game.col_log_login cll FINAL 
		WHERE login_time >= toDateTime(?, ?) AND login_time <= toDateTime(?, ?) 
		`
				var logids []string
				var args2 []any
				args2 = append(args2, s1)
				args2 = append(args2, locationName)
				args2 = append(args2, e1)
				args2 = append(args2, locationName)
				if v == 1 {
					// 复充用户
					sql2 = `SELECT DISTINCT userid FROM game.col_log_login cll FINAL 
		WHERE login_time >= toDateTime(?, ?) AND login_time <= toDateTime(?, ?) and userid in ? 
		`
					args2 = append(args2, fc_user)
				}
				err = ck.Select(&logids, sql2, args2...)
				if err != nil {
					beego.Error("PayUserRetainedCK fail err: ", err)
				}

				// 1日留存
				paycount1, count1 := PayRetainedComputeidsCK(today1, logids, channel)
				// 2日留存
				paycount2, count2 := PayRetainedComputeidsCK(today2, logids, channel)
				// 3日留存
				paycount3, count3 := PayRetainedComputeidsCK(today3, logids, channel)
				// 4日留存
				paycount4, count4 := PayRetainedComputeidsCK(today4, logids, channel)
				// 5日留存
				paycount5, count5 := PayRetainedComputeidsCK(today5, logids, channel)
				// 6日留存
				paycount6, count6 := PayRetainedComputeidsCK(today6, logids, channel)
				// 7日留存
				paycount7, count7 := PayRetainedComputeidsCK(today7, logids, channel)
				// 15日留存
				paycount15, count15 := PayRetainedComputeidsCK(today15, logids, channel)
				// 30日留存
				paycount30, count30 := PayRetainedComputeidsCK(today30, logids, channel)
				// 60日留存
				paycount60, count60 := PayRetainedComputeidsCK(today60, logids, channel)
				if count1 != 0 {
					// 往期存在待计算的数据
					datestamp, _ := utils.Unix(fmt.Sprintf("%s", today1))
					oldinfo := new(entity.PayUserRetainedData)
					oldinfo.Date = datestamp
					oldinfo.Channel = channel
					oldinfo.PayUserType = v
					info1 := StatisticsService.GetPayUserRetainedNew(oldinfo)
					if info1.Id != "" {
						// 存在登录数据
						info1.Pay1 = count1
						info1.NewNumber = paycount1
					}
					retlist = append(retlist, info1)
				}
				if count2 != 0 {
					// 往期存在待计算的数据
					datestamp, _ := utils.Unix(fmt.Sprintf("%s", today2))
					oldinfo := new(entity.PayUserRetainedData)
					oldinfo.Date = datestamp
					oldinfo.Channel = channel
					oldinfo.PayUserType = v
					info1 := StatisticsService.GetPayUserRetainedNew(oldinfo)
					if info1.Id != "" {
						// 存在登录数据
						info1.Pay2 = count2
						info1.NewNumber = paycount2
					}
					retlist = append(retlist, info1)
				}
				if count3 != 0 {
					// 往期存在待计算的数据
					datestamp, _ := utils.Unix(fmt.Sprintf("%s", today3))
					oldinfo := new(entity.PayUserRetainedData)
					oldinfo.Date = datestamp
					oldinfo.Channel = channel
					oldinfo.PayUserType = v
					info1 := StatisticsService.GetPayUserRetainedNew(oldinfo)
					if info1.Id != "" {
						// 存在登录数据
						info1.Pay3 = count3
						info1.NewNumber = paycount3
					}
					retlist = append(retlist, info1)
				}
				if count4 != 0 {
					// 往期存在待计算的数据
					datestamp, _ := utils.Unix(fmt.Sprintf("%s", today4))
					oldinfo := new(entity.PayUserRetainedData)
					oldinfo.Date = datestamp
					oldinfo.Channel = channel
					oldinfo.PayUserType = v
					info1 := StatisticsService.GetPayUserRetainedNew(oldinfo)
					if info1.Id != "" {
						// 存在登录数据
						info1.Pay4 = count4
						info1.NewNumber = paycount4
					}
					retlist = append(retlist, info1)
				}
				if count5 != 0 {
					// 往期存在待计算的数据
					datestamp, _ := utils.Unix(fmt.Sprintf("%s", today5))
					oldinfo := new(entity.PayUserRetainedData)
					oldinfo.Date = datestamp
					oldinfo.Channel = channel
					oldinfo.PayUserType = v
					info1 := StatisticsService.GetPayUserRetainedNew(oldinfo)
					if info1.Id != "" {
						// 存在登录数据
						info1.Pay5 = count5
						info1.NewNumber = paycount5
					}
					retlist = append(retlist, info1)
				}
				if count6 != 0 {
					// 往期存在待计算的数据
					datestamp, _ := utils.Unix(fmt.Sprintf("%s", today6))
					oldinfo := new(entity.PayUserRetainedData)
					oldinfo.Date = datestamp
					oldinfo.Channel = channel
					oldinfo.PayUserType = v
					info1 := StatisticsService.GetPayUserRetainedNew(oldinfo)
					if info1.Id != "" {
						// 存在登录数据
						info1.Pay6 = count6
						info1.NewNumber = paycount6
					}
					retlist = append(retlist, info1)
				}
				if count7 != 0 {
					// 往期存在待计算的数据
					datestamp, _ := utils.Unix(fmt.Sprintf("%s", today7))
					oldinfo := new(entity.PayUserRetainedData)
					oldinfo.Date = datestamp
					oldinfo.Channel = channel
					oldinfo.PayUserType = v
					info1 := StatisticsService.GetPayUserRetainedNew(oldinfo)
					if info1.Id != "" {
						// 存在登录数据
						info1.Pay7 = count7
						info1.NewNumber = paycount7
					}
					retlist = append(retlist, info1)
				}
				if count15 != 0 {
					// 往期存在待计算的数据
					datestamp, _ := utils.Unix(fmt.Sprintf("%s", today15))
					oldinfo := new(entity.PayUserRetainedData)
					oldinfo.Date = datestamp
					oldinfo.Channel = channel
					oldinfo.PayUserType = v
					info1 := StatisticsService.GetPayUserRetainedNew(oldinfo)
					if info1.Id != "" {
						// 存在登录数据
						info1.Pay15 = count15
						info1.NewNumber = paycount15
					}
					retlist = append(retlist, info1)
				}
				if count30 != 0 {
					// 往期存在待计算的数据
					datestamp, _ := utils.Unix(fmt.Sprintf("%s", today30))
					oldinfo := new(entity.PayUserRetainedData)
					oldinfo.Date = datestamp
					oldinfo.Channel = channel
					info1 := StatisticsService.GetPayUserRetainedNew(oldinfo)
					if info1.Id != "" {
						// 存在登录数据
						info1.Pay30 = count30
						info1.NewNumber = paycount30
					}
					retlist = append(retlist, info1)
				}
				if count60 != 0 {
					// 往期存在待计算的数据
					datestamp, _ := utils.Unix(fmt.Sprintf("%s", today60))
					oldinfo := new(entity.PayUserRetainedData)
					oldinfo.Date = datestamp
					oldinfo.Channel = channel
					info1 := StatisticsService.GetPayUserRetainedNew(oldinfo)
					if info1.Id != "" {
						// 存在登录数据
						info1.Pay60 = count60
						info1.NewNumber = paycount60
					}
					retlist = append(retlist, info1)
				}
			}
		}
		for _, user := range retlist {
			AddErr := StatisticsService.AddPayUserRetainedNew(&user)
			if AddErr != nil {
				beego.Error("PayUserRetainedNew fail err: ", AddErr)
			}
		}
	}
}

func PayRetainedComputeidsCK(today string, logids []string, channelId string) (paycount, lcount int64) {
	// 获取昨日注册，且充值玩家 今日登录
	paycount = 0
	s1 := fmt.Sprintf("%s 00:00:00", today)
	e1 := fmt.Sprintf("%s 23:59:59", today)
	sql0 := `SELECT COUNT(DISTINCT userid) AS PayUser from game.col_trade_record ctr FINAL 
	WHERE order_status = 4 AND package_id = ? AND ctime >= toDateTime(?, ?) AND ctime <= toDateTime(?, ?)
	`
	var args0 []any
	args0 = append(args0, channelId)
	args0 = append(args0, s1)
	args0 = append(args0, locationName)
	args0 = append(args0, e1)
	args0 = append(args0, locationName)
	err := ck.Select(&paycount, sql0, args0...)
	if err != nil {
		beego.Error("PayRetainedComputeidsCK fail err: ", err)
	}
	// 之前充值
	lcount = 0
	sql1 := `SELECT COUNT(DISTINCT userid) AS PayUser from game.col_trade_record ctr FINAL 
	WHERE order_status = 4 AND package_id = ? AND userid IN (?) AND ctime >= toDateTime(?, ?) AND ctime <= toDateTime(?, ?)
	`
	var args1 []any
	args1 = append(args1, channelId)
	args1 = append(args1, logids)
	args1 = append(args1, s1)
	args1 = append(args1, locationName)
	args1 = append(args1, e1)
	args1 = append(args1, locationName)
	err = ck.Select(&lcount, sql1, args1...)
	if err != nil {
		beego.Error("PayRetainedComputeidsCK fail err: ", err)
	}
	return paycount, lcount
}

/*
Title:用户资源
*/
func UserResource(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	m := FindByDate("", today, "", "login_time")
	m["robot"] = false
	// 全部用户
	list, _ := PlayerService.GetByUserList(m)
	Diamond := int64(0)
	Coin := int64(0)
	for _, user := range list {
		if user.Diamond > 0 {
			Diamond += user.Diamond
		} else {
			Diamond += 0
		}
		if user.Coin > 0 {
			Coin += user.Coin
		} else {
			Coin += 0
		}
	}
	count := len(list)
	date1, _ := utils.Unix(fmt.Sprintf("%s", today))
	info := new(entity.UserResource)
	info.Date = date1
	info.UType = 0
	info.UserNumber = int64(count)
	info.Diamond = ComputeFloat(Diamond, 100)
	info.Coin = ComputeFloat(Coin, 100)
	err := StatisticsService.AddUserResource(info)
	if err != nil {
		beego.Error("UserResource fail err: ", err)
	}

	// 计算活跃用户
	m1 := FindByDate(today, today, "login_time", "login_time")
	m1["robot"] = false
	list1, _ := PlayerService.GetByUserList(m1)
	Diamond1 := int64(0)
	Coin1 := int64(0)
	for _, user := range list1 {
		if user.Diamond > 0 {
			Diamond1 += user.Diamond
		} else {
			Diamond1 += 0
		}
		if user.Coin > 0 {
			Coin1 += user.Coin
		} else {
			Coin1 += 0
		}
	}
	count1 := len(list1)
	info1 := new(entity.UserResource)
	info1.Date = date1
	info1.UType = 1
	info1.UserNumber = int64(count1)
	info1.Diamond = ComputeFloat(Diamond1, 100)
	info1.Coin = ComputeFloat(Coin1, 100)
	err1 := StatisticsService.AddUserResource(info1)
	if err1 != nil {
		beego.Error("UserResource fail err: ", err1)
	}
}

/*
Author:CC
Title:资源流动 -> 全局
*/
func GlobalResource(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	s := fmt.Sprintf("%s 00:00:00", today)
	startTime := utils.Str2TimeZone(s, locationName)
	e := fmt.Sprintf("%s 23:59:59", today)
	endTime := utils.Str2TimeZone(e, locationName)

	// m := bson.M{
	// 	"$match": bson.M{
	// 		"ctime": bson.M{"$gte": startTime, "$lt": endTime},
	// 		"robot": false,
	// 	},
	// }
	// n := bson.M{
	// 	"$group": bson.M{
	// 		"_id": bson.M{"userid": "$userid", "robot": "$robot"},
	// 		"num": bson.M{
	// 			"$sum": 1,
	// 		},
	// 	},
	// }
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": startTime, "$lt": endTime},
				"robot": false,
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{"userid": "$userid", "robot": "$robot"},
				"num": bson.M{
					"$sum": 1,
				},
			},
		},
	}
	// operations := []bson.M{m, n}
	result := []bson.M{}
	pipe := PlayerUsers.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("GlobalResource fail err: ", err)
	}
	// 登录人数
	// m1 := bson.M{
	// 	"$match": bson.M{
	// 		"login_time": bson.M{"$gte": startTime, "$lt": endTime},
	// 		"robot":      false,
	// 	},
	// }
	// n1 := bson.M{
	// 	"$group": bson.M{
	// 		"_id": bson.M{"userid": "$userid", "robot": "$robot"},
	// 		"num": bson.M{
	// 			"$sum": 1,
	// 		},
	// 	},
	// }

	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"login_time": bson.M{"$gte": startTime, "$lt": endTime},
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"num": bson.M{
					"$sum": 1,
				},
			},
		},
	}
	// operations1 := []bson.M{m1, n1}
	result1 := []bson.M{}
	pipe1 := LoginLogs.Pipe(pipeline1)
	err = pipe1.All(&result1)
	if err != nil {
		beego.Error("GlobalResource fail err: ", err)
	}
	// // 充值
	// pipeline2 := []bson.M{
	// 	{
	// 		"$match": bson.M{
	// 			"login_time":   bson.M{"$gte": startTime, "$lt": endTime},
	// 			"order_status": 4,
	// 		},
	// 	},
	// 	{
	// 		"$group": bson.M{
	// 			"_id":         "$userid",
	// 			"totalAmount": bson.M{"$sum": "$amount"},
	// 			"present":     bson.M{"$addToSet": "$other_present"},
	// 		},
	// 	},
	// }
	// // operations2 := []bson.M{m2, n2}
	// result2 := []bson.M{}
	// pipe2 := Pays.Pipe(pipeline2)
	// err = pipe2.All(&result2)
	// if err != nil {
	// 	beego.Error("GlobalResource fail err: ", err)
	// }
	// // 提现
	// pipeline3 := []bson.M{
	// 	{
	// 		"$match": bson.M{
	// 			"login_time":   bson.M{"$gte": startTime, "$lt": endTime},
	// 			"order_status": 3,
	// 		},
	// 	},
	// 	{
	// 		"$group": bson.M{
	// 			"_id":         "$userid",
	// 			"totalAmount": bson.M{"$sum": "$amount"},
	// 		},
	// 	},
	// }
	// result3 := []bson.M{}
	// pipe3 := Withdraws.Pipe(pipeline3)
	// err = pipe3.All(&result3)
	// if err != nil {
	// 	beego.Error("GlobalResource fail err: ", err)
	// }

	// 注册
	pipeline4 := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": startTime, "$lt": endTime},
				"ltype": 1,
			},
		},
		{
			"$group": bson.M{
				"_id": nil,
				"Diamond": bson.M{
					"$sum": "$add_diamond",
				},
				"Coin": bson.M{
					"$sum": "$add_coin",
				},
			},
		},
	}
	result4 := []bson.M{}
	pipe4 := LogWaters.Pipe(pipeline4)
	err = pipe4.All(&result4)
	if err != nil {
		beego.Error("GlobalResource fail err: ", err)
	}

	// 游戏产出
	types := []int{45, 67, 70}
	pipeline5 := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": startTime, "$lt": endTime},
				"ltype": bson.M{"$in": types},
			},
		},
		{
			"$group": bson.M{
				"_id": nil,
				"Diamond": bson.M{
					"$sum": "$add_diamond",
				},
				"Coin": bson.M{
					"$sum": "$add_coin",
				},
			},
		},
	}
	// operations5 := []bson.M{m5, n5}
	result5 := []bson.M{}
	pipe5 := LogWaters.Pipe(pipeline5)
	err = pipe5.All(&result5)
	if err != nil {
		beego.Error("GlobalResource fail err: ", err)
	}
	// 游戏消耗 TP、龙虎、7updown
	types1 := []int{5, 75, 76}
	pipeline6 := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": startTime, "$lt": endTime},
				"ltype": bson.M{"$in": types1},
			},
		},
		{
			"$group": bson.M{
				"_id": nil,
				"Diamond": bson.M{
					"$sum": "$add_diamond",
				},
				"Coin": bson.M{
					"$sum": "$add_coin",
				},
			},
		},
	}
	// operations6 := []bson.M{m6, n6}
	result6 := []bson.M{}
	pipe6 := LogWaters.Pipe(pipeline6)
	err = pipe6.All(&result6)
	if err != nil {
		beego.Error("GlobalResource fail err: ", err)
	}
	// today = startTime.Format("2006-01-02")
	date1, _ := utils.Unix(fmt.Sprintf("%s", today))
	rCount := 0
	if len(result) > 0 {
		rCount = result[0]["num"].(int)
	}
	lCount := len(result1)
	// if len(result1) > 0 {
	// 	lCount = result1[0]["num"].(int)
	// }
	// count := 0
	// Amount := int64(0)
	// Present := int64(0)
	// if len(result2) > 0 {
	// 	for i, _ := range result2 {
	// 		t := result2[i]["totalAmount"].(int)
	// 		Amount += int64(t)
	// 		presentList, ok := result2[i]["present"].([]interface{})
	// 		if !ok {
	// 			fmt.Println("Failed to convert 'present' field to []interface{}")
	// 			return
	// 		}
	// 		// 使用类型断言将 value 转换为 string 类型
	// 		strValue, ok := presentList[0].(string)
	// 		if !ok {
	// 			fmt.Println("Failed to convert value to string")
	// 			return
	// 		}

	// 		// 使用 strconv.ParseInt() 函数将字符串转换为 int64
	// 		intValue, err := strconv.ParseInt(strValue, 10, 64)
	// 		if err != nil {
	// 			fmt.Println("Failed to convert string to int64")
	// 			return
	// 		}
	// 		Present += int64(intValue)

	// 	}
	// 	count = len(result2)
	// }
	// count3 := 0
	// Amount3 := 0
	// if len(result3) > 0 {
	// 	for _, v := range result3 {
	// 		Amount3 += v["totalAmount"].(int)
	// 	}
	// 	count3 = len(result3)
	// }
	Diamond4 := int64(0)
	Coin4 := int64(0)
	if len(result4) > 0 {
		Diamond4 = result4[0]["Diamond"].(int64)
		Coin4 = result4[0]["Coin"].(int64)
	}

	Diamond5 := int64(0)
	Coin5 := int64(0)
	if len(result5) > 0 {
		Diamond5 = result5[0]["Diamond"].(int64)
		Coin5 = result5[0]["Coin"].(int64)
	}

	Diamond6 := int64(0)
	Coin6 := int64(0)
	if len(result6) > 0 {
		Diamond6 = result6[0]["Diamond"].(int64)
		Coin6 = result6[0]["Coin"].(int64)
	}
	info := new(entity.GlobalResource)
	info.Date = date1
	info.LoginNum = int64(lCount)
	info.NewRegister = int64(rCount)
	// info.TotalRecharge = int64(count)
	// info.TotalAmount = ComputeFloat(Amount, 100)
	// info.WithdrawNum = int64(count3)
	// info.WithdrawAmount = ComputeFloat(int64(Amount3), 100)
	info.PutDiamond = ComputeFloat(int64(Diamond5), 100)
	info.PutCoin = ComputeFloat(int64(Coin5), 100)
	info.ExpendDiamond = ComputeFloat(int64(Diamond6), 100)
	info.ExpendCoin = ComputeFloat(int64(Coin6), 100)
	info.GiftDiamond = ComputeFloat(int64(Diamond4), 100)
	info.GiftCoin = ComputeFloat(int64(Coin4), 100)
	// info.PayPutDiamond = ComputeFloat(Amount, 100)
	// info.PayPutCoin = ComputeFloat(Present, 100)
	// 罐子BONUS产生和消耗

	m11 := []bson.M{
		{"$match": bson.M{"ctime": bson.M{"$gte": startTime.Unix(), "$lt": endTime.Unix()}}},
		{"$project": bson.M{
			"btype": bson.M{"$cond": bson.M{"if": bson.M{"$lt": []interface{}{"$bonus", 0}}, "then": 0, "else": 1}},
			"bonus": "$bonus",
		}},
		{"$group": bson.M{"_id": "$btype", "bonus": bson.M{"$sum": "$bonus"}}},
	}
	var r11 []bson.M
	err = BounsLogs.Pipe(m11).All(&r11)
	if err != nil {
		beego.Error("BounsLogs error", err)
	} else {
		for _, r := range r11 {
			btype := r["_id"].(int)
			bonus := r["bonus"].(int64)
			if btype == 0 {
				info.ExpendBonus = ComputeFloat(int64(bonus), 100)
			} else {
				info.PutBonus = ComputeFloat(int64(bonus), 100)
			}
		}
	}

	err1 := StatisticsService.AddGlobalResource(info)
	if err1 != nil {
		beego.Error("GlobalResource fail err: ", err1)
	}
}

/*
Author:CC
Title:资源流动 -> 游戏
*/
func GameResource(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")

	s := fmt.Sprintf("%s 00:00:00", today)
	startTime := utils.Str2TimeZone(s, locationName)
	e := fmt.Sprintf("%s 23:59:59", today)
	endTime := utils.Str2TimeZone(e, locationName)

	// TP 产出彩金(结算)
	ltypes := 111
	result := ComputeWater(startTime, endTime, ltypes)

	ids := make([]string, 0)
	Diamond := int64(0)
	Coin := int64(0)
	for _, item := range result {
		ids = append(ids, item["_id"].(string))
		Diamond += item["Diamond"].(int64)
		Coin += item["Coin"].(int64)
	}
	// TP 消耗彩金(下注、加注、sideshow)
	ltypes = 110
	result1 := ComputeWater(startTime, endTime, ltypes)
	EDiamond := int64(0)
	ECoin := int64(0)
	for _, item := range result1 {
		ids = append(ids, item["_id"].(string))
		EDiamond += item["Diamond"].(int64)
		ECoin += item["Coin"].(int64)
	}
	// 去掉重复ID
	uniqueIDs := make(map[string]bool)
	var resultIDs []string
	for _, id := range ids {
		if !uniqueIDs[id] {
			uniqueIDs[id] = true
			resultIDs = append(resultIDs, id)
		}
	}
	count := len(resultIDs)

	// 龙虎斗 产出彩金(结算)
	ltypes = 67
	result2 := ComputeWater(startTime, endTime, ltypes)

	ids1 := make([]string, 0)
	Diamond1 := int64(0)
	Coin1 := int64(0)
	for _, item := range result2 {
		ids1 = append(ids1, item["_id"].(string))
		Diamond1 += item["Diamond"].(int64)
		Coin1 += item["Coin"].(int64)
	}
	// 龙虎斗 消耗彩金(下注)
	ltypes = 75
	result3 := ComputeWater(startTime, endTime, ltypes)
	EDiamond1 := int64(0)
	ECoin1 := int64(0)
	for _, item := range result3 {
		ids1 = append(ids1, item["_id"].(string))
		EDiamond1 += item["Diamond"].(int64)
		ECoin1 += item["Coin"].(int64)
	}
	// 去掉重复ID
	uniqueIDs1 := make(map[string]bool)
	var resultIDs1 []string
	for _, id := range ids1 {
		if !uniqueIDs1[id] {
			uniqueIDs1[id] = true
			resultIDs1 = append(resultIDs1, id)
		}
	}
	count1 := len(resultIDs1)

	// 7updown 产出彩金(结算)
	ltypes = 70
	result4 := ComputeWater(startTime, endTime, ltypes)

	ids2 := make([]string, 0)
	Diamond2 := int64(0)
	Coin2 := int64(0)
	for _, item := range result4 {
		ids2 = append(ids2, item["_id"].(string))
		Diamond2 += item["Diamond"].(int64)
		Coin2 += item["Coin"].(int64)
	}
	// 7updown 消耗彩金(下注)
	ltypes = 76
	result5 := ComputeWater(startTime, endTime, ltypes)
	EDiamond2 := int64(0)
	ECoin2 := int64(0)
	for _, item := range result5 {
		ids2 = append(ids2, item["_id"].(string))
		EDiamond2 += item["Diamond"].(int64)
		ECoin2 += item["Coin"].(int64)
	}
	// 去掉重复ID
	uniqueIDs2 := make(map[string]bool)
	var resultIDs2 []string
	for _, id := range ids2 {
		if !uniqueIDs2[id] {
			uniqueIDs2[id] = true
			resultIDs2 = append(resultIDs2, id)
		}
	}
	count2 := len(resultIDs2)

	// CRASH 产出彩金(结算)
	ltypes = 73
	result6 := ComputeWater(startTime, endTime, ltypes)

	ids3 := make([]string, 0)
	Diamond3 := int64(0)
	Coin3 := int64(0)
	for _, item := range result6 {
		ids3 = append(ids, item["_id"].(string))
		Diamond3 += item["Diamond"].(int64)
		Coin3 += item["Coin"].(int64)
	}
	// CRASH 消耗彩金(下注)
	ltypes = 79
	result7 := ComputeWater(startTime, endTime, ltypes)
	EDiamond3 := int64(0)
	ECoin3 := int64(0)
	for _, item := range result7 {
		ids3 = append(ids, item["_id"].(string))
		EDiamond3 += item["Diamond"].(int64)
		ECoin3 += item["Coin"].(int64)
	}
	// 去掉重复ID
	uniqueIDs3 := make(map[string]bool)
	var resultIDs3 []string
	for _, id := range ids3 {
		if !uniqueIDs3[id] {
			uniqueIDs3[id] = true
			resultIDs3 = append(resultIDs3, id)
		}
	}
	count3 := len(resultIDs3)

	// RM 产出彩金(结算)
	ltypes = 117
	result8 := ComputeWater(startTime, endTime, ltypes)

	ids4 := make([]string, 0)
	Diamond4 := int64(0)
	Coin4 := int64(0)
	for _, item := range result8 {
		ids4 = append(ids4, item["_id"].(string))
		Diamond4 += item["Diamond"].(int64)
		Coin4 += item["Coin"].(int64)
	}

	// RM 消耗彩金
	ltypes = 116
	result9 := ComputeWater(startTime, endTime, ltypes)
	EDiamond4 := int64(0)
	ECoin4 := int64(0)
	for _, item := range result9 {
		ids4 = append(ids4, item["_id"].(string))
		EDiamond4 += item["Diamond"].(int64)
		ECoin4 += item["Coin"].(int64)
	}
	// 去掉重复ID
	uniqueIDs4 := make(map[string]bool)
	var resultIDs4 []string
	for _, id := range ids4 {
		if !uniqueIDs4[id] {
			uniqueIDs4[id] = true
			resultIDs4 = append(resultIDs4, id)
		}
	}
	count4 := len(resultIDs4)

	date1, _ := utils.Unix(fmt.Sprintf("%s", today))
	info := new(entity.GameResource)
	info.Date = date1
	info.TpNumber = int64(count)
	info.TpDiamond = ComputeFloat(int64(Diamond), 100)
	info.TpCoin = ComputeFloat(int64(Coin), 100)
	info.TpExpendDiamond = ComputeFloat(int64(EDiamond), 100)
	info.TpExpendCoin = ComputeFloat(int64(ECoin), 100)
	info.RmNumber = int64(count4)                              // int64(count)
	info.RmDiamond = ComputeFloat(int64(Diamond4), 100)        // ComputeFloat(Diamond, 100)
	info.RmCoin = ComputeFloat(int64(Coin4), 100)              // ComputeFloat(Coin, 100)
	info.RmExpendDiamond = ComputeFloat(int64(EDiamond4), 100) // ComputeFloat(EDiamond, 100)
	info.RmExpendCoin = ComputeFloat(int64(ECoin4), 100)       // ComputeFloat(ECoin, 100)
	info.LhdNumber = int64(count1)
	info.LhdDiamond = ComputeFloat(int64(Diamond1), 100)
	info.LhdCoin = ComputeFloat(int64(Coin1), 100)
	info.LhdExpendDiamond = ComputeFloat(int64(EDiamond1), 100)
	info.LhdExpendCoin = ComputeFloat(int64(ECoin1), 100)
	info.UpNumber = int64(count2)
	info.UpDiamond = ComputeFloat(int64(Diamond2), 100)
	info.UpCoin = ComputeFloat(int64(Coin2), 100)
	info.UpExpendDiamond = ComputeFloat(int64(EDiamond2), 100)
	info.UpExpendCoin = ComputeFloat(int64(ECoin2), 100)
	info.CRASHNumber = int64(count3)
	info.CRASHDiamond = ComputeFloat(int64(Diamond3), 100)
	info.CRASHCoin = ComputeFloat(int64(Coin3), 100)
	info.CRASHExpendDiamond = ComputeFloat(int64(EDiamond3), 100)
	info.CRASHExpendCoin = ComputeFloat(int64(ECoin3), 100)
	err1 := StatisticsService.AddGamesResource(info)
	if err1 != nil {
		beego.Error("GameResource fail err: ", err1)
	}
}

// 根据类型汇总金币流水数据
func ComputeWater(startTime time.Time, endTime time.Time, ltype int) []bson.M {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": startTime, "$lt": endTime},
				"ltype": ltype,
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Diamond": bson.M{
					"$sum": "$add_diamond",
				},
				"Coin": bson.M{
					"$sum": "$add_coin",
				},
			},
		},
	}
	result := []bson.M{}
	pipe := LogWaters.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("ComputeWater fail err: ", err)
	}
	return result
}

/*
Author:CC
Title:财务管理 -> 支付渠道成功率
*/
func PayChannelRate(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")

	m := bson.M{"status": 1}
	var channellist []entity.ChannelInfo
	err := Channels.
		Find(m).
		All(&channellist)
	if err != nil {
		beego.Error("PayChannelRate fail err: ", err)
	}
	list := make([]entity.ChannelSuccessRate, 0)
	if len(channellist) > 0 {
		date1, _ := utils.Unix(fmt.Sprintf("%s", today))
		PayChannels, _ := GameService.GetPayChannelList(-1, -1, bson.M{})
		for _, item := range channellist {
			id := item.Name // 渠道名称
			name1 := item.Name1
			for _, payitem := range PayChannels {
				info := new(entity.ChannelSuccessRate)
				payid, _ := strconv.ParseInt(payitem.Id, 10, 64)
				m := FindByDate(today, today, "ctime", "ctime")
				m["channel_id"] = payid
				m["package_id"] = id
				paylist, _ := PayService.GetByPayUser(m)
				// 登录人数
				var userids []string
				m1 := bson.M{
					"robot":            false,
					"simulation_robot": false,
					"ad__bundle_id":    id,
				}
				PlayerUsers.Find(m1).Distinct("_id", &userids)

				m2 := FindByDate(today, today, "login_time", "login_time")
				m2["userid"] = bson.M{"$in": userids}
				var logids []string
				LoginLogs.Find(m2).Distinct("userid", &logids)

				dlcount := int64(len(logids))
				soCount := int64(0)
				cgAmount := int64(0)
				sbAmount := int64(0)
				uniqueUserIDs := make(map[string]bool)
				for _, v := range paylist {
					if v.OrderStatus == 4 {
						soCount += 1
						cgAmount += int64(v.Amount)
					}
					if v.OrderStatus == 2 {
						sbAmount += int64(v.Amount)
					}
					//ids = append(ids, v1.Userid)
					uniqueUserIDs[v.Userid] = true
				}
				var result []string
				for id := range uniqueUserIDs {
					result = append(result, id)
				}
				cgOrder := int64(len(paylist))
				pOrder := int64(len(result))

				m3 := FindByDate(today, today, "ctime", "ctime")
				m3["out_channel"] = strconv.FormatInt(payid, 10)
				m3["package_id"] = id
				withlist, _ := PayService.GetByWithdrawUser(m3)
				soCount1 := int64(0)
				cgAmount1 := int64(0)
				sbAmount1 := int64(0)
				dsfCount := int64(0)   // 第三方提单数量
				dsfcgCount := int64(0) // 第三方提单成功数量
				uniqueUserIDs1 := make(map[string]bool)
				for _, v := range withlist {
					if v.OrderStatus == 2 {
						soCount1 += 1
						cgAmount1 += int64(v.Amount)
					}
					if v.OrderStatus == 4 {
						sbAmount1 += int64(v.Amount)
					}
					if v.OrderStatus == 2 || v.OrderStatus == 5 || v.OrderStatus == 6 || v.OrderStatus == 7 {
						dsfCount++
					}
					if v.OrderStatus == 2 || v.OrderStatus == 5 || v.OrderStatus == 6 {
						dsfcgCount++
					}
					//ids = append(ids, v1.Userid)
					uniqueUserIDs1[v.Userid] = true
				}
				var result1 []string
				for id := range uniqueUserIDs1 {
					result1 = append(result1, id)
				}

				cgOrder1 := int64(len(withlist))
				pOrder1 := int64(len(result1))
				info.Date = date1
				info.PackageId = id
				info.PackageName = name1
				info.Channel = payid
				info.PayRequest = pOrder
				info.LoginNumber = dlcount // 今日登录人数
				info.PayRequestOrder = cgOrder
				info.PaySuccessOrder = soCount
				info.PaySuccessRate = ComputeFloat(soCount, cgOrder) * 100
				info.PaySuccessMoney = ComputeFloat(cgAmount, 100)
				info.PayFailMoney = ComputeFloat(sbAmount, 100)
				info.WithdrawRequest = pOrder1
				info.WithdrawRequestOrder = cgOrder1
				info.WithdrawSuccessOrder = soCount1
				info.WithdrawSuccessRate = ComputeFloat(soCount1, cgOrder1) * 100
				info.WithdrawSuccessMoney = ComputeFloat(cgAmount1, 100)
				info.WithdrawFailMoney = ComputeFloat(sbAmount1, 100)
				info.ThirdpartyOrder = dsfCount
				info.ThirdpartyOrderRate = ComputeFloat(dsfCount, dsfCount) * 100
				info.ThirdpartySuccessOrder = dsfcgCount
				list = append(list, *info)
			}

		}
	}
	for _, user := range list {
		AddErr := StatisticsService.AddChannelRate(&user)
		if AddErr != nil {
			beego.Error("ChannelSuccessRate1 fail err: ", AddErr)
		}
	}
}

/*
Author:CC
Title:财务管理 -> 支付渠道成功率
*/
func PayChannelRateCK(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
	date1, _ := utils.Unix(fmt.Sprintf("%s", today))
	// 充值
	sql1 := `SELECT
    channel_id,
    package_id,
    COUNT(DISTINCT userid) AS total_number,
    COUNT(1)  AS total_count,
    SUM(amount) AS total_amount,
    COUNT(DISTINCT CASE WHEN order_status = 4 THEN userid END) AS success_number,
    SUM(CASE WHEN order_status = 4 THEN 1 ELSE 0 END) AS success_count,
    SUM(CASE WHEN order_status = 4 THEN amount ELSE 0 END) AS success_amount,
    SUM(CASE WHEN order_status = 2 THEN amount ELSE 0 END) AS fail_amount
FROM game.col_trade_record ctr FINAL
WHERE package_id != '' and ctime BETWEEN ? AND ?
GROUP BY channel_id, package_id`
	// ctime between ? and ? and
	var paylist []map[string]any
	var args1 []any
	args1 = append(args1, startTime)
	args1 = append(args1, endTime)
	err = ck.Select(&paylist, sql1, args1...)
	if err != nil {
		beego.Error("PayChannelRateCK error:", err)
	}
	// 提现
	sql2 := `SELECT
    out_channel,
    package_id,
    COUNT(DISTINCT userid) AS total_number,
    COUNT(1)  AS total_count,
    SUM(amount) AS total_amount,
    COUNT(DISTINCT CASE WHEN order_status = 2 THEN userid END) AS success_number,
    SUM(CASE WHEN order_status = 2 THEN 1 ELSE 0 END) AS success_count,
    SUM(CASE WHEN order_status = 2 THEN amount ELSE 0 END) AS success_amount,
    SUM(CASE WHEN order_status = 5 THEN amount ELSE 0 END) AS fail_amount,
    COUNT(CASE WHEN order_status in (2,5,6,7) THEN 1 ELSE 0 END) AS dsf_count,
    COUNT(CASE WHEN order_status in (2,5,6) THEN 1 ELSE 0 END) AS dsfcg_count
FROM game.col_withdraw_record ctr FINAL
WHERE package_id != '' and ctime BETWEEN ? AND ?
GROUP BY out_channel, package_id`
	// ctime between ? and ? and
	var withlist []map[string]any
	var args2 []any
	args2 = append(args2, startTime)
	args2 = append(args2, endTime)
	err = ck.Select(&withlist, sql2, args2...)
	if err != nil {
		beego.Error("PayChannelRateCK error:", err)
	}
	maplist := make(map[string]*entity.ChannelSuccessRate, 0)
	// 渠道
	packagelist := make([]string, 0)
	for _, p := range paylist {
		cid := p["channel_id"].(uint32)
		pid := p["package_id"].(string)
		total_number := p["total_number"].(uint64)
		total_count := p["total_count"].(uint64)
		// total_amount := p["total_amount"].(uint64)
		// success_number := p["success_number"].(uint64)
		success_count := p["success_count"].(uint64)
		success_amount := p["success_amount"].(uint64)
		fail_amount := p["fail_amount"].(uint64)

		key := fmt.Sprintf("%s-%d", pid, cid)
		stat, ok := maplist[key]
		if !ok {
			stat = &entity.ChannelSuccessRate{
				Date: date1,
			}
			maplist[key] = stat
		}
		stat.Channel = int64(cid)
		stat.PackageId = pid
		stat.PayRequest = int64(total_number)
		stat.PayRequestOrder = int64(total_count)
		stat.PaySuccessOrder = int64(success_count)
		stat.PaySuccessRate = ComputeFloat(int64(success_count), int64(total_count)) * 100
		stat.PaySuccessMoney = ComputeFloat(int64(success_amount), 100)
		stat.PayFailMoney = ComputeFloat(int64(fail_amount), 100)
		packagelist = append(packagelist, pid)
	}
	for _, p := range withlist {
		cel := p["out_channel"].(string)
		cid := 0
		if cel != "" {
			cid, _ = strconv.Atoi(cel)
		}
		pid := p["package_id"].(string)
		total_number := p["total_number"].(uint64)
		total_count := p["total_count"].(uint64)
		// total_amount := p["total_amount"].(uint64)
		// success_number := p["success_number"].(uint64)
		success_count := p["success_count"].(uint64)
		success_amount := p["success_amount"].(uint64)
		fail_amount := p["fail_amount"].(uint64)
		dsf_count := p["dsf_count"].(uint64)
		dsfcg_count := p["dsfcg_count"].(uint64)

		key := fmt.Sprintf("%s-%d", pid, cid)
		stat, ok := maplist[key]
		if !ok {
			stat = &entity.ChannelSuccessRate{
				Date: date1,
			}
			maplist[key] = stat
		}
		stat.Channel = int64(cid)
		stat.PackageId = pid
		stat.WithdrawRequest = int64(total_number)
		stat.WithdrawRequestOrder = int64(total_count)
		stat.WithdrawSuccessOrder = int64(success_count)
		stat.WithdrawSuccessRate = ComputeFloat(int64(success_count), int64(total_count)) * 100
		stat.WithdrawSuccessMoney = ComputeFloat(int64(success_amount), 100)
		stat.WithdrawFailMoney = ComputeFloat(int64(fail_amount), 100)
		stat.ThirdpartyOrder = int64(dsf_count)
		stat.ThirdpartyOrderRate = ComputeFloat(int64(dsfcg_count), int64(dsf_count)) * 100
		stat.ThirdpartySuccessOrder = int64(dsfcg_count)
		packagelist = append(packagelist, pid)
	}
	m := bson.M{"status": 1}
	var channellist []entity.ChannelInfo
	err := Channels.
		Find(m).
		All(&channellist)
	if err != nil {
		beego.Error("PayChannelRateCK fail err: ", err)
	}
	// 获取登陆用户id
	m2 := FindByDate(today, today, "login_time", "login_time")
	var logids []string
	LoginLogs.Find(m2).Distinct("userid", &logids)
	// 登陆用户渠道信息
	sql3 := `select ad__bundle_id,COUNT(1) as count from game.col_user cu FINAL where robot = 0 and simulation_robot = 0 and userid in ? 
	group by ad__bundle_id`
	// ctime between ? and ? and
	var ulist []map[string]any
	var args3 []any
	args3 = append(args3, logids)
	err = ck.Select(&ulist, sql3, args3...)
	if err != nil {
		beego.Error("PayChannelRateCK error:", err)
	}
	loginlist := make(map[string]int64, 0)
	for _, u := range ulist {
		name := u["ad__bundle_id"].(string)
		count := u["count"].(uint64)
		loginlist[name] = int64(count)
	}
	for _, v := range maplist {
		for _, p := range channellist {
			if p.Name == v.PackageId {
				v.PackageName = p.Name1
				break
			}
		}
		lcount, ok := loginlist[v.PackageId]
		if ok {
			v.LoginNumber = lcount
		}
		AddErr := StatisticsService.AddChannelRate(v)
		if AddErr != nil {
			beego.Error("PayChannelRateCK add fail err: ", AddErr)
		}
	}
}

// 每日0点获取一次支付渠道余额
func PayChannelStats(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
	date1, _ := utils.Unix(fmt.Sprintf("%s", today))
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"date": bson.M{"$gte": startTime.UnixMilli(), "$lt": endTime.UnixMilli()},
			},
		},
		{
			"$group": bson.M{
				"_id": "$pay_channel",
				"ds_amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
					bson.M{"$eq": []interface{}{"$p_type", 1}}, "$amount", 0,
				}}},
				"ds_rate": bson.M{"$max": bson.M{"$cond": []interface{}{
					bson.M{"$eq": []interface{}{"$p_type", 1}}, "$rate", 0.0,
				}}},
				"ds_sj_amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
					bson.M{"$eq": []interface{}{"$p_type", 1}}, "$actual_amount", 0,
				}}},
				"df_amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
					bson.M{"$eq": []interface{}{"$p_type", 2}}, "$amount", 0,
				}}},
				"df_rate": bson.M{"$max": bson.M{"$cond": []interface{}{
					bson.M{"$eq": []interface{}{"$p_type", 2}}, "$rate", 0.0,
				}}},
				"df_sj_amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
					bson.M{"$eq": []interface{}{"$p_type", 2}}, "$actual_amount", 0,
				}}},

				"ds_handling": bson.M{"$sum": bson.M{"$cond": []interface{}{
					bson.M{"$eq": []interface{}{"$p_type", 1}}, "$handling_charge", 0,
				}}},
				"df_handling": bson.M{"$sum": bson.M{"$cond": []interface{}{
					bson.M{"$eq": []interface{}{"$p_type", 2}}, "$handling_charge", 0,
				}}},
			},
		},
	}
	result := []bson.M{}
	err = PayChannelLogs.Pipe(pipeline).All(&result)
	list := make([]entity.PayChannelStatsData, 0)
	if len(result) > 0 {
		for _, item := range result {
			cid := 0
			dsId, _ := item["_id"]
			if dsId == nil {
				cid = 0
			} else {
				cid = dsId.(int)
			}
			ds_amount := ConvertToInt64(item["ds_amount"])
			ds_rate := item["ds_rate"].(float64)
			ds_sj_amount := ConvertToInt64(item["ds_sj_amount"])
			df_amount := ConvertToInt64(item["df_amount"])
			df_rate := item["df_rate"].(float64)
			df_sj_amount := ConvertToInt64(item["df_sj_amount"])
			ds_handling := ConvertToInt64(item["ds_handling"])
			df_handling := ConvertToInt64(item["df_handling"])
			info := new(entity.PayChannelStatsData)
			info.Date = date1
			info.PayChannel = uint32(cid)
			info.TotalPay = ds_amount
			info.PayRate = ds_rate
			info.TotalPayAmount = ds_sj_amount
			info.TotalWithdraw = df_amount
			info.WithdrawRate = df_rate
			info.TotalWithdrawAmount = df_sj_amount
			info.PayHandlingFee = ds_handling
			info.WithdrawHandlingFee = df_handling
			list = append(list, *info)
		}
	}
	for _, user := range list {
		AddErr := PayService.AddOrUpdatePayChannelStat(&user)
		if AddErr != nil {
			beego.Error("PayChannelStats fail err: ", AddErr)
		}
	}
}

// 分享活动数据
func ShareActivityData(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	s := fmt.Sprintf("%s 00:00:00", today)
	startTime := utils.Str2TimeZone(s, locationName)
	e := fmt.Sprintf("%s 23:59:59", today)
	endTime := utils.Str2TimeZone(e, locationName)
	// 查询邀请注册人数
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"ctime":          bson.M{"$gte": startTime, "$lt": endTime},
				"share_superior": bson.M{"$ne": ""},
			},
		},
		{
			"$group": bson.M{
				"_id": "$_id",
				"num": bson.M{
					"$sum": 1,
				},
			},
		},
	}
	// operations := []bson.M{m, n}
	result := []bson.M{}
	pipe := PlayerUsers.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("DataSummary fail err: ", err)
	}
	Rcount := len(result)
	// 注册 用户表
	//产出  LogShareWaters
	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": startTime.Unix(), "$lt": endTime.Unix()},
			},
		},
		{
			"$group": bson.M{
				"_id": "$gtype",
				"Cash": bson.M{
					"$sum": "$cash",
				},
				"Coin": bson.M{
					"$sum": "$coin",
				},
			},
		},
	}
	// operations := []bson.M{m, n}
	result1 := []bson.M{}
	pipe1 := LogShareWaters.Pipe(pipeline1)
	err = pipe1.All(&result1)
	if err != nil {
		beego.Error("DataSummary fail err: ", err)
	}
	WaterCash := int64(0)
	WaterCoin := int64(0)
	FirstCash := int64(0)
	FirstCoin := int64(0)
	for _, item := range result1 {
		id := item["_id"].(int)
		switch id {
		case 1:
			WaterCash += item["Cash"].(int64)
			WaterCoin += item["Coin"].(int64)
		case 2:
			FirstCash += item["Cash"].(int64)
			FirstCoin += item["Coin"].(int64)
		}
	}
	WaterTotal := WaterCash + WaterCoin
	FirstTotal := FirstCash + FirstCoin
	Total := WaterTotal + FirstTotal
	info := new(entity.ShareActivityData)
	date1, _ := utils.Unix(fmt.Sprintf("%s", today))
	info.Ctime = date1
	info.RegisterNumber = int64(Rcount)
	info.WaterOutPut = ComputeFloat(WaterTotal, 100)
	info.FirstOrderOutPut = ComputeFloat(FirstTotal, 100)
	info.GrossOutput = ComputeFloat(Total, 100)
	AddErr := ActivityService.AddShareData(info)
	if AddErr != nil {
		beego.Error("ShareActivityData fail err: ", AddErr)
	}
}

// 统计埋点数据
func PointData(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	m := bson.M{"status": 1}
	var channellist []entity.ChannelInfo
	err := Channels.
		Find(m).
		All(&channellist)
	if err != nil {
		beego.Error("PointData fail err: ", err)
	}

	if len(channellist) > 0 {
		date1, _ := utils.Unix(fmt.Sprintf("%s", today))
		for _, item := range channellist {
			id := item.Name // 渠道名称
			name1 := item.Name1
			m := FindByDate(today, today, "ctime", "ctime")
			m["ad__bundle_id"] = id
			list, _ := PlayerService.GetByUserList(m)
			count := len(list)
			// 游客注册
			m1 := FindByDate(today, today, "ctime", "ctime")
			m1["tourist"] = bson.M{"$ne": ""}
			m1["ad__bundle_id"] = id
			ykcount, _ := PlayerService.GetTotal(m1)
			// 手机注册
			m2 := FindByDate(today, today, "ctime", "ctime")
			m2["tourist"] = ""
			m2["ad__bundle_id"] = id
			sjcount, _ := PlayerService.GetTotal(m2)

			userid := make([]string, 0)
			for _, item := range list {
				userid = append(userid, item.Userid)
			}

			m3 := FindByDate1(today, today, "ctime", "ctime")
			m3["typ"] = 1
			m3["userid"] = bson.M{"$in": userid}
			lqCount, _ := StatisticsService.GetLogEventTracksTotal(m3)

			m3["typ"] = 2
			d1TPCount, _ := StatisticsService.GetLogEventTracksTotal(m3)

			m3["typ"] = 3
			d2TPCount, _ := StatisticsService.GetLogEventTracksTotal(m3)

			m3["typ"] = 4
			tcTPCount, _ := StatisticsService.GetLogEventTracksTotal(m3)

			m3["typ"] = 5
			// ydlist, _ := StatisticsService.GetLogEventTracksTotal(m3)
			var d1list []entity.LogEventTrack // 第一局玩的游戏
			d1Game := new(entity.PlayersNumber)
			LogEventTracks.Find(m3).All(&d1list)
			if len(d1list) > 0 {
				tpCount := 0
				rmCount := 0
				aKCount := 0
				jokerCount := 0
				lhdCount := 0
				upCount := 0
				carshCount := 0
				for _, item := range d1list {
					switch item.Gtype {
					case 1:
						tpCount += 1
					case 2:
						lhdCount += 1
					case 3:
						upCount += 1
					case 4:
						rmCount += 1
					case 5:
						aKCount += 1
					case 6:
						jokerCount += 1
					case 7:
						carshCount += 1
					}
				}
				d1Game.TpNumber = int64(tpCount)
				d1Game.LHDNumber = int64(lhdCount)
				d1Game.UPNumber = int64(upCount)
				d1Game.RmNumber = int64(rmCount)
				d1Game.AkNumber = int64(aKCount)
				d1Game.JokerNumber = int64(jokerCount)
				d1Game.CrashNumber = int64(carshCount)
			}

			m3["typ"] = 6
			// ydlist, _ := StatisticsService.GetLogEventTracksTotal(m3)
			var d2list []entity.LogEventTrack // 第二局玩的游戏
			d2Game := new(entity.PlayersNumber)
			LogEventTracks.Find(m3).All(&d2list)
			if len(d2list) > 0 {
				tpCount := 0
				rmCount := 0
				aKCount := 0
				jokerCount := 0
				lhdCount := 0
				upCount := 0
				carshCount := 0
				for _, item := range d2list {
					switch item.Gtype {
					case 1:
						tpCount += 1
					case 2:
						lhdCount += 1
					case 3:
						upCount += 1
					case 4:
						rmCount += 1
					case 5:
						aKCount += 1
					case 6:
						jokerCount += 1
					case 7:
						carshCount += 1
					}
				}
				d2Game.TpNumber = int64(tpCount)
				d2Game.LHDNumber = int64(lhdCount)
				d2Game.UPNumber = int64(upCount)
				d2Game.RmNumber = int64(rmCount)
				d2Game.AkNumber = int64(aKCount)
				d2Game.JokerNumber = int64(jokerCount)
				d2Game.CrashNumber = int64(carshCount)
			}

			m3["typ"] = 7
			newGold, _ := StatisticsService.GetLogEventTracksTotal(m3)

			m3["typ"] = 8
			newGold1, _ := StatisticsService.GetLogEventTracksTotal(m3)

			info := new(entity.PointData)
			info.Date = date1
			info.Channel = id
			info.Channel1 = name1
			info.TotalRegister = int64(count)
			info.TotalTourist = int64(ykcount)
			info.TotalMobile = int64(sjcount)
			info.GuidanceGetGold = lqCount
			info.TPNumber1 = d1TPCount
			info.TPNumber2 = d2TPCount
			info.TPExitManually = tcTPCount
			info.Gold100 = newGold
			info.Gold200 = newGold1
			info.FirstGames = d1Game
			info.SecondGames = d2Game
			err1 := StatisticsService.AddPointData(info)
			if err1 != nil {
				beego.Error("PointData fail err: ", err1)
			}
		}
	}
}

// 统计埋点数据-局数分析
func GameNumberAnalysisData(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	s := fmt.Sprintf("%s 00:00:00", today)
	startTime := utils.Str2TimeZone(s, locationName)
	e := fmt.Sprintf("%s 23:59:59", today)
	endTime := utils.Str2TimeZone(e, locationName)
	m := bson.M{"status": 1}
	var channellist []entity.ChannelInfo
	err := Channels.
		Find(m).
		All(&channellist)
	if err != nil {
		beego.Error("GameNumberAnalysisData fail err: ", err)
	}
	date1, _ := utils.Unix(fmt.Sprintf("%s", today))
	if len(channellist) > 0 {
		for _, item := range channellist {
			name := item.Name // 渠道名称
			name1 := item.Name1

			pipeline := []bson.M{
				{
					"$match": bson.M{
						"ctime":            bson.M{"$gte": startTime, "$lt": endTime},
						"robot":            false,
						"simulation_robot": false,
						"ad__bundle_id":    name,
					},
				},
				{
					"$group": bson.M{
						"_id": "$_id",
					},
				},
			}
			// operations := []bson.M{m, n}
			result := []bson.M{}
			pipe := PlayerUsers.Pipe(pipeline)
			err := pipe.All(&result)
			if err != nil {
				beego.Error("GameNumberAnalysisData fail err: ", err)
			}
			// ids := make([]string, 0)
			count := int64(0)
			count1 := int64(0)
			count2 := int64(0)
			count3 := int64(0)
			count4 := int64(0)
			count5 := int64(0)
			count6 := int64(0)
			count11 := int64(0)
			count21 := int64(0)
			count31 := int64(0)
			statrTimestamp := startTime.Unix()
			endTimestamp := endTime.Unix()
			for _, item := range result {
				id := item["_id"].(string)
				userCount := GameByUserCount(statrTimestamp, endTimestamp, id)
				if userCount == 0 {
					count += 1
				}
				if userCount == 1 {
					count1 += 1
				}
				if userCount == 2 {
					count2 += 1
				}
				if userCount == 3 {
					count3 += 1
				}
				if userCount == 4 {
					count4 += 1
				}
				if userCount == 5 {
					count5 += 1
				}
				if userCount >= 6 && userCount <= 10 {
					count6 += 1
				}
				if userCount >= 11 && userCount <= 20 {
					count11 += 1
				}
				if userCount >= 21 && userCount <= 30 {
					count21 += 1
				}
				if userCount >= 31 {
					count31 += 1
				}
			}

			rcount := len(result)
			info := new(entity.GameNumberAnalysis)
			info.Date = date1
			info.Channel = name
			info.Channel1 = name1
			info.TotalRegister = int64(rcount)
			info.GNumber = count
			info.GNumber1 = count1
			info.GNumber2 = count2
			info.GNumber3 = count3
			info.GNumber4 = count4
			info.GNumber5 = count5
			info.GNumber6 = count6
			info.GNumber11 = count11
			info.GNumber21 = count21
			info.GNumber31 = count31
			// beego.Info("GameNumberAnalysisData ADD :", info)
			AddErr := StatisticsService.AddGameNumberAnalysis(info)
			if AddErr != nil {
				beego.Error("GameNumberAnalysisData fail err: ", AddErr.Error())
			}
		}
	}

}

// 统计埋点数据-局数分析
func BugCommitData(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	// s := fmt.Sprintf("%s 00:00:00", today)
	// startTime := utils.Str2Time(s)
	// e := fmt.Sprintf("%s 23:59:59", today)
	// endTime := utils.Str2Time(e)

	m := FindByDate3(today, today, "ctime", "ctime")
	m["bug_id"] = bson.M{
		"$elemMatch": bson.M{
			"$eq": 1,
		},
	}
	bugCount1, _ := StatisticsService.GetLogBugFeedbackLog(m)
	m["bug_id"] = bson.M{
		"$elemMatch": bson.M{
			"$eq": 2,
		},
	}
	bugCount2, _ := StatisticsService.GetLogBugFeedbackLog(m)
	m["bug_id"] = bson.M{
		"$elemMatch": bson.M{
			"$eq": 3,
		},
	}
	bugCount3, _ := StatisticsService.GetLogBugFeedbackLog(m)
	m["bug_id"] = bson.M{
		"$elemMatch": bson.M{
			"$eq": 4,
		},
	}
	bugCount4, _ := StatisticsService.GetLogBugFeedbackLog(m)
	m["bug_id"] = bson.M{
		"$elemMatch": bson.M{
			"$eq": 5,
		},
	}
	bugCount5, _ := StatisticsService.GetLogBugFeedbackLog(m)

	date1, _ := utils.Unix(fmt.Sprintf("%s", today))
	info := new(entity.BugStatistics)
	info.Date = date1
	info.BugType1 = bugCount1
	info.BugType2 = bugCount2
	info.BugType3 = bugCount3
	info.BugType4 = bugCount4
	info.BugType5 = bugCount5
	AddErr := StatisticsService.AddBugData(info)
	if AddErr != nil {
		beego.Error("BugCommitData fail err: ", AddErr.Error())
	}
}

// 房间数据
func RoomData(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")

	// s := fmt.Sprintf("%s 00:00:00", today)
	// startTime := utils.Str2Time(s)
	// e := fmt.Sprintf("%s 23:59:59", today)
	// endTime := utils.Str2Time(e)

	m := bson.M{"status": 1}
	var channellist []entity.ChannelInfo
	err := Channels.
		Find(m).
		All(&channellist)
	if err != nil {
		beego.Error("RoomData fail err: ", err)
	}
	// tempinfo := new(entity.ChannelInfo)
	// tempinfo.Name = "net.gameduo.tbd"
	// tempinfo.Name1 = "fb"
	// channellist = append(channellist, *tempinfo)
	list := make([]entity.RoomData, 0)
	if len(channellist) > 0 {
		sTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
		d := sTime.AddDate(0, 0, -1)
		// 当天玩过游戏的userid
		startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
		endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
		statrTimestamp := startTime.Unix()
		endTimestamp := endTime.Unix()

		var palyer_arr []string
		n := bson.M{}
		n["ctime"] = bson.M{"$gte": statrTimestamp, "$lt": endTimestamp}
		LogGameTimes.Find(n).Distinct("userid", &palyer_arr)
		for _, item := range channellist {
			id := item.Name // 渠道名称
			name1 := item.Name1
			// 老玩家
			var user_old []string
			m := bson.M{}
			m["ctime"] = bson.M{"$lt": d}
			m["robot"] = false
			m["simulation_robot"] = false
			m["ad__bundle_id"] = id
			m["_id"] = bson.M{"$in": palyer_arr}
			PlayerUsers.Find(m).Distinct("_id", &user_old)
			list_old := RoomDataCompute(2, today, id, name1, user_old)
			list = append(list, list_old...)

			// 新玩家
			user_new := make([]string, 0)
			m1 := FindByDate(today, today, "ctime", "ctime")
			m1["robot"] = false
			m1["simulation_robot"] = false
			m1["ad__bundle_id"] = id
			m1["_id"] = bson.M{"$in": palyer_arr}
			PlayerUsers.Find(m1).Distinct("_id", &user_new)
			list_new := RoomDataCompute(1, today, id, name1, user_new)
			list = append(list, list_new...)
		}
	}
	for _, item := range list {
		AddErr := StatisticsService.AddRoomData(&item)
		if AddErr != nil {
			beego.Error("RoomData fail err: ", AddErr.Error())
		}
	}
	// beego.Info("RoomData 完成")
}
func RoomDataCompute(typ int64, today, channel, channelname string, uids []string) []entity.RoomData {
	var list []entity.RoomData
	s := fmt.Sprintf("%s 00:00:00", today)
	startTime := utils.Str2TimeZone(s, locationName)
	e := fmt.Sprintf("%s 23:59:59", today)
	endTime := utils.Str2TimeZone(e, locationName)
	date1, _ := utils.Unix(fmt.Sprintf("%s", today))
	statrTimestamp := startTime.Unix()
	endTimestamp := endTime.Unix()
	// gtype := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	gTotalMap := make(map[int][]string, 0) // 用户对局数

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"ctime":  bson.M{"$gte": statrTimestamp, "$lt": endTimestamp},
				"userid": bson.M{"$in": uids},
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Count": bson.M{
					"$sum": 1,
				},
			},
		},
	}
	result := []bson.M{}
	pipe := LogGameTimes.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("RoomDataCompute fail err: ", err)
	}
	// 外接局数获取
	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"ctime":   bson.M{"$gte": statrTimestamp, "$lt": endTimestamp},
				"user_id": bson.M{"$in": uids},
				"amount":  bson.M{"$ne": 0},
			},
		},
		{
			"$group": bson.M{
				"_id": "$user_id",
				"Count": bson.M{
					"$sum": 1,
				},
			},
		},
	}
	result1 := []bson.M{}
	pipe1 := NsqLogExternalBets.Pipe(pipeline1)
	err = pipe1.All(&result1)
	if err != nil {
		beego.Error("RoomDataCompute fail err: ", err)
	}
	// 创建一个 map 来存储每个 ID 对应的总 count 值
	countMap := make(map[string]int)

	for _, item := range result {
		id := item["_id"].(string)
		count := item["Count"].(int)
		countMap[id] += count
	}

	for _, item := range result1 {
		id := item["_id"].(string)
		count := item["Count"].(int)
		countMap[id] += count
	}
	for key, item := range countMap {
		// ids = append(ids, item["_id"].(string))
		id := key
		number := item
		if number == 1 {
			ids := make([]string, 0)
			if _, ok := gTotalMap[1]; ok {
				ids = gTotalMap[1]
			}
			ids = append(ids, id)
			gTotalMap[1] = ids
		}
		// 2-5局
		if number >= 2 && number <= 5 {
			ids := make([]string, 0)
			if _, ok := gTotalMap[2]; ok {
				ids = gTotalMap[2]
			}
			ids = append(ids, id)
			gTotalMap[2] = ids
		}
		// 6-10局
		if number >= 6 && number <= 10 {
			ids := make([]string, 0)
			if _, ok := gTotalMap[3]; ok {
				ids = gTotalMap[3]
			}
			ids = append(ids, id)
			gTotalMap[3] = ids
		}
		// 11-20局
		if number >= 11 && number <= 20 {
			ids := make([]string, 0)
			if _, ok := gTotalMap[4]; ok {
				ids = gTotalMap[4]
			}
			ids = append(ids, id)
			gTotalMap[4] = ids
		}
		// 21-30局
		if number >= 21 && number <= 30 {
			ids := make([]string, 0)
			if _, ok := gTotalMap[5]; ok {
				ids = gTotalMap[5]
			}
			ids = append(ids, id)
			gTotalMap[5] = ids
		}
		// 31-50局
		if number >= 31 && number <= 50 {
			ids := make([]string, 0)
			if _, ok := gTotalMap[6]; ok {
				ids = gTotalMap[6]
			}
			ids = append(ids, id)
			gTotalMap[6] = ids
		}
		// 51局及以上
		if number >= 51 {
			ids := make([]string, 0)
			if _, ok := gTotalMap[7]; ok {
				ids = gTotalMap[7]
			}
			ids = append(ids, id)
			gTotalMap[7] = ids
		}
	}

	for index, item := range gTotalMap {
		info := new(entity.RoomData)
		TpGame := 0
		TpRealGame := int64(0)
		TpTime := int64(0)
		TpNumber := 0
		LhdGame := 0
		LhdTime := int64(0)
		LhdNumber := 0
		UpGame := 0
		UpTime := int64(0)
		UpNumber := 0
		RummyGame := 0
		RummyRealGame := int64(0)
		RummyTime := int64(0)
		RummyNumber := 0
		AkGame := 0
		AkRealGame := int64(0)
		AkTime := int64(0)
		AkNumber := 0
		JokerGame := 0
		JokerRealGame := int64(0)
		JokerTime := int64(0)
		JokerNumber := 0
		CrashGame := 0
		CrashTime := int64(0)
		CrashNumber := 0
		ABGame := 0
		ABTime := int64(0)
		ABNumber := 0
		CPGame := 0
		CPTime := int64(0)
		CPNumber := 0
		FJGame := 0
		FJTime := int64(0)
		FJNumber := 0
		// 红黑
		RBGame := 0
		RBTime := int64(0)
		RBNumber := 0

		// Rummy双人
		RMTwoGame := 0
		RMTwoTime := int64(0)
		RMTwoNumber := 0

		// TP2
		TP2Game := 0
		TP2Time := int64(0)
		TP2Number := 0
		// Solts
		SoltsGame := 0
		SoltsNumber := 0
		ZRSXGame := 0
		ZRSXNumber := 0
		// 对战房
		TPBattleGame := 0
		RMBattleGame := 0
		ABBattleGame := 0
		// TP计算
		tpResult := RoomDataByGameType(statrTimestamp, endTimestamp, 1, item)
		if tpResult != nil {
			for _, item1 := range tpResult {
				TpGame += item1["Count"].(int)
				TpTime += item1["Time"].(int64)
			}
			TpNumber = len(tpResult)
		}
		// LHD计算
		lhdResult := RoomDataByGameType(statrTimestamp, endTimestamp, 2, item)
		if lhdResult != nil {
			for _, item1 := range lhdResult {
				LhdGame += item1["Count"].(int)
				LhdTime += item1["Time"].(int64)
			}
			LhdNumber = len(lhdResult)
		}
		// 7updown计算
		upResult := RoomDataByGameType(statrTimestamp, endTimestamp, 3, item)
		if upResult != nil {
			for _, item1 := range upResult {
				UpGame += item1["Count"].(int)
				UpTime += item1["Time"].(int64)
			}
			UpNumber = len(upResult)
		}
		// Rummy计算
		ryResult := RoomDataByGameType(statrTimestamp, endTimestamp, 4, item)
		if ryResult != nil {
			for _, item1 := range ryResult {
				RummyGame += item1["Count"].(int)
				RummyTime += item1["Time"].(int64)
			}
			RummyNumber = len(ryResult)
		}
		// AK47计算
		akResult := RoomDataByGameType(statrTimestamp, endTimestamp, 5, item)
		if akResult != nil {
			for _, item1 := range akResult {
				AkGame += item1["Count"].(int)
				AkTime += item1["Time"].(int64)
			}
			AkNumber = len(akResult)
		}
		// Joker计算
		jokerResult := RoomDataByGameType(statrTimestamp, endTimestamp, 6, item)
		if jokerResult != nil {
			for _, item1 := range jokerResult {
				JokerGame += item1["Count"].(int)
				JokerTime += item1["Time"].(int64)
			}
			JokerNumber = len(jokerResult)
		}
		// Crash计算
		crashResult := RoomDataByGameType(statrTimestamp, endTimestamp, 7, item)
		if crashResult != nil {
			for _, item1 := range crashResult {
				CrashGame += item1["Count"].(int)
				CrashTime += item1["Time"].(int64)
			}
			CrashNumber = len(crashResult)
		}

		// ANDARBAHAR计算
		abResult := RoomDataByGameType(statrTimestamp, endTimestamp, 8, item)
		if abResult != nil {
			for _, item1 := range abResult {
				ABGame += item1["Count"].(int)
				ABTime += item1["Time"].(int64)
			}
			ABNumber = len(abResult)
		}
		// 彩票计算
		cpResult := RoomDataByGameType(statrTimestamp, endTimestamp, 9, item)
		if cpResult != nil {
			for _, item1 := range cpResult {
				CPGame += item1["Count"].(int)
				CPTime += item1["Time"].(int64)
			}
			CPNumber = len(cpResult)
		}

		// 飞机计算
		fjResult := RoomDataByGameType(statrTimestamp, endTimestamp, 10, item)
		if fjResult != nil {
			for _, item1 := range fjResult {
				FJGame += item1["Count"].(int)
				FJTime += item1["Time"].(int64)
			}
			FJNumber = len(fjResult)
		}
		// 红黑计算
		rbResult := RoomDataByGameType(statrTimestamp, endTimestamp, 11, item)
		if rbResult != nil {
			for _, item1 := range rbResult {
				RBGame += item1["Count"].(int)
				RBTime += item1["Time"].(int64)
			}
			RBNumber = len(rbResult)
		}
		// Rummy双人计算
		rmTwoResult := RoomDataByGameType(statrTimestamp, endTimestamp, 12, item)
		if rmTwoResult != nil {
			for _, item1 := range rmTwoResult {
				RMTwoGame += item1["Count"].(int)
				RMTwoTime += item1["Time"].(int64)
			}
			RMTwoNumber = len(rmTwoResult)
		}
		// Rummy双人计算
		tp2Result := RoomDataByGameType(statrTimestamp, endTimestamp, 13, item)
		if tp2Result != nil {
			for _, item1 := range tp2Result {
				TP2Game += item1["Count"].(int)
				TP2Time += item1["Time"].(int64)
			}
			TP2Number = len(tp2Result)
		}
		// 真人视讯game_id
		zrsx_gameid_arr := []int{
			600607, 600572, 600526, 600513, 600594, 600583, 600642, 600536, 600631,
			602287, 604137, 602344, 600561,
			604278, // 百人非真人
			600312, // 捕鱼
			600766, // 捕鱼
		}
		slots_gameid_arr := []int{
			600101, 600002, 600073, 600022, 600039, 600120, 600054, 600104, 600037, 600028, 600041, 600098, 600025, 600093, 600004, 600108, 600009, 600012, 600119, 600110, 600029, 600081, 600086, 600099, 600102, 600117, 600103,
			600276, 600369, 600402, 600332, 600247, 602687, 600326, 600325, 600374, 600019, 604266, 600076, 600761, 600362,
		}
		// Slots计算
		pipeline4 := []bson.M{
			{
				"$match": bson.M{
					"ctime":   bson.M{"$gte": statrTimestamp, "$lt": endTimestamp},
					"user_id": bson.M{"$in": item},
					"amount":  bson.M{"$ne": 0},
					"game_id": bson.M{"$in": slots_gameid_arr},
				},
			},
			{
				"$group": bson.M{
					"_id": "$user_id",
					"Count": bson.M{
						"$sum": 1,
					},
				},
			},
		}
		result4 := []bson.M{}
		pipe4 := NsqLogExternalBets.Pipe(pipeline4)
		err = pipe4.All(&result4)
		if err != nil {
			beego.Error("RoomData fail err: ", err)
		}
		for _, temp := range result4 {
			count := temp["Count"].(int)
			SoltsGame += count
		}
		SoltsNumber = len(result4)
		// 真人视讯
		pipeline5 := []bson.M{
			{
				"$match": bson.M{
					"ctime":   bson.M{"$gte": statrTimestamp, "$lt": endTimestamp},
					"user_id": bson.M{"$in": item},
					"amount":  bson.M{"$ne": 0},
					"game_id": bson.M{"$in": zrsx_gameid_arr},
				},
			},
			{
				"$group": bson.M{
					"_id": "$user_id",
					"Count": bson.M{
						"$sum": 1,
					},
				},
			},
		}
		result5 := []bson.M{}
		pipe5 := NsqLogExternalBets.Pipe(pipeline5)
		err = pipe5.All(&result5)
		if err != nil {
			beego.Error("RoomData fail err: ", err)
		}
		for _, temp := range result5 {
			count := temp["Count"].(int)
			ZRSXGame += count
		}
		ZRSXNumber = len(result5)

		// 真人对局计算
		starttamp, _ := utils.Str2Local(s, location)
		endtamp, _ := utils.Str2Local(e, location)
		pipeline2 := []bson.M{
			{
				"$match": bson.M{
					"end_time":  bson.M{"$gte": starttamp, "$lt": endtamp},
					"gtype":     bson.M{"$in": []int{1, 4, 5, 6}},
					"real_game": true,
				},
			},
			{
				"$group": bson.M{
					"_id": bson.M{
						"jhid":    "$_id",
						"gtype":   "$gtype",
						"players": "$players",
					},
				},
			},
			{
				"$project": bson.M{
					"id":      "$_id.jhid",
					"gtype":   "$_id.gtype",
					"players": "$_id.players",
				},
			},
		}
		result2 := []bson.M{}
		pipe2 := Details.Pipe(pipeline2)
		err := pipe2.All(&result2)
		if err != nil {
			beego.Error("RoomData fail err: ", err)
		}
		newTpGame := make(map[string]bool)
		newAKGame := make(map[string]bool)
		newRMGame := make(map[string]bool)
		newJokerGame := make(map[string]bool)
		for _, v := range result2 {
			id := v["id"].(string)
			playsers := v["players"].(string)
			gtype := v["gtype"].(int)
			playerArr := strings.Split(playsers, ",")
			IsNew := false
			if len(playerArr) > 0 {
				for _, p := range playerArr {
					if len(p) <= 10 {
						// 真人
						for _, uid := range item {
							if uid == p {
								IsNew = true
							}
						}
					}
				}
				if IsNew {
					switch gtype {
					case 1:
						newTpGame[id] = true
					case 4:
						newRMGame[id] = true
					case 5:
						newAKGame[id] = true
					case 6:
						newJokerGame[id] = true
					}
				}
			}
		}
		TpRealGame = int64(len(newTpGame))
		RummyRealGame = int64(len(newRMGame))
		AkRealGame = int64(len(newAKGame))
		JokerRealGame = int64(len(newJokerGame))

		// 对战房
		for _, uid := range item {
			pipeline3 := []bson.M{
				{
					"$match": bson.M{
						"end_time": bson.M{"$gte": starttamp, "$lt": endtamp},
						"gtype":    bson.M{"$in": []int{1, 4, 8}},
						"rtype":    1,
						"players": bson.M{
							"$regex":   fmt.Sprintf("\\b%s\\b", uid),
							"$options": "i",
						},
					},
				},
				{
					"$group": bson.M{
						"_id": "$gtype",
						"num": bson.M{
							"$sum": 1,
						},
					},
				},
			}
			result3 := []bson.M{}
			pipe3 := Details.Pipe(pipeline3)
			err = pipe3.All(&result3)
			if err != nil {
				beego.Error("RoomData fail err: ", err)
			}
			for _, v := range result3 {
				gtype := v["_id"].(int)
				num := v["num"].(int)
				switch gtype {
				case 1:
					TPBattleGame += num
				case 4:
					RMBattleGame += num
				case 8:
					ABBattleGame += num
				}
			}
		}

		// RoomData 赋值
		info.Date = date1
		info.Channel = channel
		info.Channel1 = channelname
		info.NumberTypes = index
		info.PlayerTypes = typ
		info.PlayerTotal = int64(len(item))
		info.TPGameNumber = int64(TpGame)
		info.TPGameRealNumber = TpRealGame
		info.TPNumber = int64(TpNumber)
		info.TPGameTime = TpTime
		info.LHDGameNumber = int64(LhdGame)
		info.LHDNumber = int64(LhdNumber)
		info.LHDGameTime = LhdTime
		info.UPGameNumber = int64(UpGame)
		info.UPNumber = int64(UpNumber)
		info.UPGameTime = UpTime
		info.RummyGameNumber = int64(RummyGame)
		info.RMGameRealNumber = RummyRealGame
		info.RummyNumber = int64(RummyNumber)
		info.RummyGameTime = RummyTime
		info.AKGameNumber = int64(AkGame)
		info.AKGameRealNumber = AkRealGame
		info.AKNumber = int64(AkNumber)
		info.AKGameTime = AkTime
		info.JokerGameNumber = int64(JokerGame)
		info.JokerGameRealNumber = JokerRealGame
		info.JokerNumber = int64(JokerNumber)
		info.JokerGameTime = JokerTime
		info.CrashGameNumber = int64(CrashGame)
		info.CrashNumber = int64(CrashNumber)
		info.CrashGameTime = CrashTime
		info.ABGameNumber = int64(ABGame)
		info.ABNumber = int64(ABNumber)
		info.ABGameTime = ABTime
		info.CPGameNumber = int64(CPGame)
		info.CPNumber = int64(CPNumber)
		info.CPGameTime = CPTime
		info.FJGameNumber = int64(FJGame)
		info.FJNumber = int64(FJNumber)
		info.FJGameTime = FJTime
		info.RBGameNumber = int64(RBGame)
		info.RBNumber = int64(RBNumber)
		info.RBGameTime = RBTime
		info.RMTwoGameNumber = int64(RMTwoGame)
		info.RMTwoNumber = int64(RMTwoNumber)
		info.RMTwoGameTime = RMTwoTime
		info.TP2GameNumber = int64(TP2Game)
		info.TP2Number = int64(TP2Number)
		info.TP2GameTime = TP2Time
		info.SlotsGameNumber = int64(SoltsGame)
		info.SlotsNumber = int64(SoltsNumber)
		info.ZRSXGameNumber = int64(ZRSXGame)
		info.ZRSXNumber = int64(ZRSXNumber)
		// 对战房
		info.TPBattleGame = int64(TPBattleGame)
		info.RMBattleGame = int64(RMBattleGame)
		info.ABBattleGame = int64(ABBattleGame)
		list = append(list, *info)
	}
	return list
}
func RoomDataByGameType(startTime, endTime int64, gtype int, userid []string) []bson.M {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"ctime":  bson.M{"$gte": startTime, "$lt": endTime},
				"gtype":  gtype,
				"userid": bson.M{"$in": userid},
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Count": bson.M{
					"$sum": 1,
				},
				"Time": bson.M{
					"$sum": "$time",
				},
			},
		},
	}
	result := []bson.M{}
	pipe := LogGameTimes.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("RoomDataByGameType fail err: ", err)
	}
	return result
}

// 房间数据
func RoomDataCK(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")

	// s := fmt.Sprintf("%s 00:00:00", today)
	// startTime := utils.Str2Time(s)
	// e := fmt.Sprintf("%s 23:59:59", today)
	// endTime := utils.Str2Time(e)

	m := bson.M{"status": 1}
	var channellist []entity.ChannelInfo
	err := Channels.
		Find(m).
		All(&channellist)
	if err != nil {
		beego.Error("RoomData fail err: ", err)
	}
	// tempinfo := new(entity.ChannelInfo)
	// tempinfo.Name = "net.gameduo.tbd"
	// tempinfo.Name1 = "fb"
	// channellist = append(channellist, *tempinfo)
	list := make([]entity.RoomData, 0)
	if len(channellist) > 0 {
		sTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
		d := sTime.AddDate(0, 0, -1)
		// 当天玩过游戏的userid
		startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
		endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)

		var palyer_arr []string
		n := bson.M{}
		n["login_time"] = bson.M{"$gte": startTime, "$lt": endTime}
		LoginLogs.Find(n).Distinct("userid", &palyer_arr)

		for _, item := range channellist {
			id := item.Name // 渠道名称
			name1 := item.Name1
			// 老玩家
			var user_old []string
			m := bson.M{}
			m["ctime"] = bson.M{"$lt": d}
			m["robot"] = false
			m["simulation_robot"] = false
			m["ad__bundle_id"] = id
			m["_id"] = bson.M{"$in": palyer_arr}
			PlayerUsers.Find(m).Distinct("_id", &user_old)
			if len(user_old) > 0 {
				list_old := RoomDataComputeCK(2, today, id, name1, user_old)
				list = append(list, list_old...)
			}

			// 新玩家
			user_new := make([]string, 0)
			m1 := FindByDate(today, today, "ctime", "ctime")
			m1["robot"] = false
			m1["simulation_robot"] = false
			m1["ad__bundle_id"] = id
			m1["_id"] = bson.M{"$in": palyer_arr}
			PlayerUsers.Find(m1).Distinct("_id", &user_new)
			if len(user_new) > 0 {
				list_new := RoomDataComputeCK(1, today, id, name1, user_new)
				list = append(list, list_new...)
			}
		}
	}
	for _, item := range list {
		AddErr := StatisticsService.AddRoomData(&item)
		if AddErr != nil {
			beego.Error("RoomData fail err: ", AddErr.Error())
		}
	}
}
func RoomDataComputeCK(typ int64, today, channel, channelname string, uids []string) []entity.RoomData {
	var list []entity.RoomData
	s := fmt.Sprintf("%s 00:00:00", today)
	startTime := utils.Str2TimeZone(s, locationName)
	e := fmt.Sprintf("%s 23:59:59", today)
	endTime := utils.Str2TimeZone(e, locationName)
	date1, _ := utils.Unix(fmt.Sprintf("%s", today))
	statrTimestamp := startTime.Unix()
	endTimestamp := endTime.Unix()
	// gtype := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	gTotalMap := make(map[int][]string, 0) // 用户对局数
	sql1 := `select userid,COUNT(1) rounds
	from game.col_detail cd FINAL
	where robot = 0 and begin_time between ? and ? and userid in ?
	group by userid
	UNION ALL
	select  user_id userid,COUNT(1) rounds
	FROM  game.col_nsq_log_external_bet cnleb FINAL where amount != 0 and cancel = 0 and ctime between ? and ? and user_id in ?
	GROUP BY user_id`
	var glist []map[string]any
	var args1 []any
	args1 = append(args1, statrTimestamp)
	args1 = append(args1, endTimestamp)
	args1 = append(args1, uids)
	args1 = append(args1, statrTimestamp)
	args1 = append(args1, endTimestamp)
	args1 = append(args1, uids)
	err = ck.Select(&glist, sql1, args1...)
	if err != nil {
		beego.Error("RoomDataComputeCK error:", err)
	}
	for _, item := range glist {
		id := item["userid"].(string)
		number := item["rounds"].(uint64)
		// ids = append(ids, item["_id"].(string))
		if number == 1 {
			ids := make([]string, 0)
			if _, ok := gTotalMap[1]; ok {
				ids = gTotalMap[1]
			}
			ids = append(ids, id)
			gTotalMap[1] = ids
			continue
		}
		// 2-5局
		if number >= 2 && number <= 5 {
			ids := make([]string, 0)
			if _, ok := gTotalMap[2]; ok {
				ids = gTotalMap[2]
			}
			ids = append(ids, id)
			gTotalMap[2] = ids
			continue
		}
		// 6-10局
		if number >= 6 && number <= 10 {
			ids := make([]string, 0)
			if _, ok := gTotalMap[3]; ok {
				ids = gTotalMap[3]
			}
			ids = append(ids, id)
			gTotalMap[3] = ids
			continue
		}
		// 11-20局
		if number >= 11 && number <= 20 {
			ids := make([]string, 0)
			if _, ok := gTotalMap[4]; ok {
				ids = gTotalMap[4]
			}
			ids = append(ids, id)
			gTotalMap[4] = ids
			continue
		}
		// 21-30局
		if number >= 21 && number <= 30 {
			ids := make([]string, 0)
			if _, ok := gTotalMap[5]; ok {
				ids = gTotalMap[5]
			}
			ids = append(ids, id)
			gTotalMap[5] = ids
			continue
		}
		// 31-50局
		if number >= 31 && number <= 50 {
			ids := make([]string, 0)
			if _, ok := gTotalMap[6]; ok {
				ids = gTotalMap[6]
			}
			ids = append(ids, id)
			gTotalMap[6] = ids
			continue
		}
		// 51局及以上
		if number >= 51 {
			ids := make([]string, 0)
			if _, ok := gTotalMap[7]; ok {
				ids = gTotalMap[7]
			}
			ids = append(ids, id)
			gTotalMap[7] = ids
			continue
		}
	}
	// 真人视讯game_id
	wj_gameid_arr := map[int]int{
		600607: 2,
		600572: 2,
		600526: 2,
		600513: 2,
		600594: 2,
		600583: 2,
		600642: 2,
		600536: 2,
		600631: 2,
		// slots
		600101: 1,
		600002: 1,
		600073: 1,
		600022: 1,
		600039: 1,
		600120: 1,
		600054: 1,
		600104: 1,
		600037: 1,
		600028: 1,
		600041: 1,
		600098: 1,
		600025: 1,
		600093: 1,
		600004: 1,
		600108: 1,
		600009: 1,
		600012: 1,
		600119: 1,
		600110: 1,
		600029: 1,
		600081: 1,
		600086: 1,
		600099: 1,
		600102: 1,
		600117: 1,
		600103: 1,
	}
	for index, item := range gTotalMap {
		info := new(entity.RoomData)
		TpGame := int64(0)
		TpRealGame := int64(0)
		TpTime := int64(0)
		TpNumber := int64(0)
		LhdGame := int64(0)
		LhdTime := int64(0)
		LhdNumber := int64(0)
		UpGame := int64(0)
		UpTime := int64(0)
		UpNumber := int64(0)
		RummyGame := int64(0)
		RummyRealGame := int64(0)
		RummyTime := int64(0)
		RummyNumber := int64(0)
		AkGame := int64(0)
		AkRealGame := int64(0)
		AkTime := int64(0)
		AkNumber := int64(0)
		JokerGame := int64(0)
		JokerRealGame := int64(0)
		JokerTime := int64(0)
		JokerNumber := int64(0)
		CrashGame := int64(0)
		CrashTime := int64(0)
		CrashNumber := int64(0)
		ABGame := int64(0)
		ABTime := int64(0)
		ABNumber := int64(0)
		CPGame := int64(0)
		CPTime := int64(0)
		CPNumber := int64(0)
		FJGame := int64(0)
		FJTime := int64(0)
		FJNumber := int64(0)
		// 红黑
		RBGame := int64(0)
		RBTime := int64(0)
		RBNumber := int64(0)

		// Rummy双人
		RMTwoGame := int64(0)
		RMTwoTime := int64(0)
		RMTwoNumber := int64(0)

		// TP2
		TP2Game := int64(0)
		TP2Time := int64(0)
		TP2Number := int64(0)
		// Solts
		SoltsGame := int64(0)
		SoltsNumber := 0
		ZRSXGame := int64(0)
		ZRSXNumber := 0
		// 真人局
		TPGameRealNumber := int64(0)
		RMGameRealNumber := int64(0)
		AKGameRealNumber := int64(0)
		JokerGameRealNumber := int64(0)
		// 对战房
		TPBattleGame := int64(0)
		RMBattleGame := int64(0)
		ABBattleGame := int64(0)
		sql2 := `select userid,gtype,COUNT(1) rounds,SUM(end_time-begin_time) times,SUM(case when real_game = 1 then 1 else 0 end) real_rounds,SUM(case when rtype = 1 then 1 else 0 end) battle_rounds
		from game.col_detail cd FINAL
		where robot = 0 and begin_time between ? and ? and userid in ?
		group by userid,gtype
		UNION ALL
		select  user_id userid,game_id gtype,COUNT(1) rounds,0 times,0 real_rounds,0 battle_rounds
		FROM  game.col_nsq_log_external_bet cnleb FINAL where amount != 0 and cancel = 0 and ctime between ? and ? and user_id in ?
		GROUP BY user_id,game_id`
		var gamelist []map[string]any
		var args2 []any
		args2 = append(args2, statrTimestamp)
		args2 = append(args2, endTimestamp)
		args2 = append(args2, item)
		args2 = append(args2, statrTimestamp)
		args2 = append(args2, endTimestamp)
		args2 = append(args2, item)
		err = ck.Select(&gamelist, sql2, args2...)
		if err != nil {
			beego.Error("RoomDataComputeCK error1:", err)
		}
		slotsids := make(map[string]bool, 0)
		sxids := make(map[string]bool, 0)
		for _, v := range gamelist {
			id := v["userid"].(string)
			rounds := v["rounds"].(uint64)
			gtype := v["gtype"].(int32)
			times := v["times"].(int64)
			real_rounds := v["real_rounds"].(uint64)
			battle_rounds := v["battle_rounds"].(uint64)
			switch gtype {
			case 1:
				TpNumber++
				TpGame += int64(rounds)
				TpTime += int64(times)
				TPGameRealNumber += int64(real_rounds)
				TPBattleGame += int64(battle_rounds)
				continue
			case 2:
				LhdNumber++
				LhdGame += int64(rounds)
				LhdTime += int64(times)
				continue
			case 3:
				UpNumber++
				UpGame += int64(rounds)
				UpTime += int64(times)
				continue
			case 4:
				RummyNumber++
				RummyGame += int64(rounds)
				RummyTime += int64(times)
				RMGameRealNumber += int64(real_rounds)
				RMBattleGame += int64(battle_rounds)
				continue
			case 5:
				AkNumber++
				AkGame += int64(rounds)
				AkTime += int64(times)
				AKGameRealNumber += int64(real_rounds)
				continue
			case 6:
				JokerNumber++
				JokerGame += int64(rounds)
				JokerTime += int64(times)
				JokerGameRealNumber += int64(real_rounds)
				continue
			case 7:
				CrashNumber++
				CrashGame += int64(rounds)
				CrashTime += int64(times)
				continue
			case 8:
				ABNumber++
				ABGame += int64(rounds)
				ABTime += int64(times)
				ABBattleGame += int64(battle_rounds)
				continue
			case 9:
				CPNumber++
				CPGame += int64(rounds)
				CPTime += int64(times)
				continue
			case 10:
				FJNumber++
				FJGame += int64(rounds)
				FJTime += int64(times)
				continue
			case 11:
				RBNumber++
				RBGame += int64(rounds)
				RBTime += int64(times)
				continue
			case 12:
				RMTwoNumber++
				RMTwoGame += int64(rounds)
				RMTwoTime += int64(times)
				continue
			case 13:
				TP2Number++
				TP2Game += int64(rounds)
				TP2Time += int64(times)
				continue
			default:
				// 外接
				if wid, ok := wj_gameid_arr[int(gtype)]; ok {
					if wid == 1 {
						slotsids[id] = true
						SoltsGame += int64(rounds)
					} else if wid == 2 {
						sxids[id] = true
						ZRSXGame += int64(rounds)
					}
				}
				continue
			}
		}
		SoltsNumber = len(slotsids)
		ZRSXNumber = len(sxids)

		info.Date = date1
		info.Channel = channel
		info.Channel1 = channelname
		info.NumberTypes = index
		info.PlayerTypes = typ
		info.PlayerTotal = int64(len(item))
		info.TPGameNumber = int64(TpGame)
		info.TPGameRealNumber = TpRealGame
		info.TPNumber = int64(TpNumber)
		info.TPGameTime = TpTime
		info.LHDGameNumber = int64(LhdGame)
		info.LHDNumber = int64(LhdNumber)
		info.LHDGameTime = LhdTime
		info.UPGameNumber = int64(UpGame)
		info.UPNumber = int64(UpNumber)
		info.UPGameTime = UpTime
		info.RummyGameNumber = int64(RummyGame)
		info.RMGameRealNumber = RummyRealGame
		info.RummyNumber = int64(RummyNumber)
		info.RummyGameTime = RummyTime
		info.AKGameNumber = int64(AkGame)
		info.AKGameRealNumber = AkRealGame
		info.AKNumber = int64(AkNumber)
		info.AKGameTime = AkTime
		info.JokerGameNumber = int64(JokerGame)
		info.JokerGameRealNumber = JokerRealGame
		info.JokerNumber = int64(JokerNumber)
		info.JokerGameTime = JokerTime
		info.CrashGameNumber = int64(CrashGame)
		info.CrashNumber = int64(CrashNumber)
		info.CrashGameTime = CrashTime
		info.ABGameNumber = int64(ABGame)
		info.ABNumber = int64(ABNumber)
		info.ABGameTime = ABTime
		info.CPGameNumber = int64(CPGame)
		info.CPNumber = int64(CPNumber)
		info.CPGameTime = CPTime
		info.FJGameNumber = int64(FJGame)
		info.FJNumber = int64(FJNumber)
		info.FJGameTime = FJTime
		info.RBGameNumber = int64(RBGame)
		info.RBNumber = int64(RBNumber)
		info.RBGameTime = RBTime
		info.RMTwoGameNumber = int64(RMTwoGame)
		info.RMTwoNumber = int64(RMTwoNumber)
		info.RMTwoGameTime = RMTwoTime
		info.TP2GameNumber = int64(TP2Game)
		info.TP2Number = int64(TP2Number)
		info.TP2GameTime = TP2Time
		info.SlotsGameNumber = int64(SoltsGame)
		info.SlotsNumber = int64(SoltsNumber)
		info.ZRSXGameNumber = int64(ZRSXGame)
		info.ZRSXNumber = int64(ZRSXNumber)
		// 对战房
		info.TPBattleGame = int64(TPBattleGame)
		info.RMBattleGame = int64(RMBattleGame)
		info.ABBattleGame = int64(ABBattleGame)
		info.Userids = item
		list = append(list, *info)
	}
	return list
}

/*
	IP重复检测
*/
// IP重复检测 定时执行
func CheckIPDuplication(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	// tim = tim.AddDate(0, 0, -4)
	endTime := tim                       //.Format("2006-01-02 15:04:05")
	startTime := tim.Add(-1 * time.Hour) // 往前一个小时
	m := bson.M{}
	m["login_time"] = bson.M{"$gte": startTime, "$lt": endTime}
	var userlist []entity.UserIpRecords // IP检测记录
	var userips []string                // 去重后的ip
	// 声明一个 map 变量用于去重
	// usermap := make(map[string]bool)
	LoginLogs.Find(m).Distinct("ip", &userips)
	if len(userips) > 0 {
		pipeline := []bson.M{
			{
				"$match": bson.M{
					"ip": bson.M{"$in": userips},
				},
			},
			{
				"$group": bson.M{
					"_id":   "$ip",
					"users": bson.M{"$addToSet": "$userid"},
				},
			},
		}
		result := []bson.M{}
		pipe := LoginLogs.Pipe(pipeline)
		err := pipe.All(&result)
		if err != nil {
			beego.Error("CheckIPDuplication fail err: ", err)
		}
		if len(result) > 0 {
			for _, item := range result {
				info := new(entity.UserIpRecords)
				// info.Userid = item
				ip := item["_id"].(string)
				arrUser := item["users"].([]interface{})
				users := make([]string, 0)
				for _, v := range arrUser {
					users = append(users, v.(string))
				}
				// users :=
				info.Ip = ip
				info.Userid = users
				userlist = append(userlist, *info)
			}
		}
	}
	if len(userlist) > 0 {
		for _, info := range userlist {
			AddErr := LoggerService.AddUserIpRecod(&info)
			if AddErr != nil {
				beego.Error("CheckIPDuplication fail err: ", AddErr.Error())
			}
		}

	}
}

/*
Author：CC
Title：用户付费留存
Effect：统计用户付费留存
*/
func PayUserRetained(timestamp int64) {
	nowday := utils.Stamp2Time(timestamp)
	today := nowday.Format("2006-01-02")
	today1 := nowday.AddDate(0, 0, -1).Format("2006-01-02")
	today2 := nowday.AddDate(0, 0, -2).Format("2006-01-02")
	today3 := nowday.AddDate(0, 0, -3).Format("2006-01-02")
	today4 := nowday.AddDate(0, 0, -4).Format("2006-01-02")
	today5 := nowday.AddDate(0, 0, -5).Format("2006-01-02")
	today6 := nowday.AddDate(0, 0, -6).Format("2006-01-02")
	today7 := nowday.AddDate(0, 0, -7).Format("2006-01-02")
	today15 := nowday.AddDate(0, 0, -15).Format("2006-01-02")
	today30 := nowday.AddDate(0, 0, -30).Format("2006-01-02")
	today60 := nowday.AddDate(0, 0, -60).Format("2006-01-02")
	// 筛选首充用户
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"order_status": 4,
			},
		},
		{
			"$group": bson.M{
				"_id":   "$userid",
				"count": bson.M{"$sum": 1}, // 添加一个额外的字段以便后续过滤
			},
		},
		{
			"$match": bson.M{
				"count": bson.M{"$eq": 1}, // 过滤num值在1文档
			},
		},
	}
	var res []bson.M
	Pays.Pipe(pipeline).All(&res)
	first_pay_id := make([]string, 0)
	for _, item := range res {
		pid := item["_id"].(string)
		first_pay_id = append(first_pay_id, pid)
	}
	pay_type := []int{0, 1}
	addlist := make([]entity.PayUserRetained4Web, 0)
	for _, v := range pay_type {
		var payids []string // 注册且支付ID
		// 当天充值总人数
		m1 := FindByDate(today, today, "ctime", "ctime")
		if v == 0 {
			m1["userid"] = bson.M{"$in": first_pay_id}
		} else {
			m1["userid"] = bson.M{"$nin": first_pay_id}
		}
		m1["order_status"] = 4 // 充值成功
		Pays.Find(m1).Distinct("userid", &payids)
		count := len(payids)

		// 1日留存
		regcount1, count1 := PayRetainedCompute4Web(today, today1, first_pay_id, v)
		// 2日留存
		regcount2, count2 := PayRetainedCompute4Web(today, today2, first_pay_id, v)
		// 3日留存
		regcount3, count3 := PayRetainedCompute4Web(today, today3, first_pay_id, v)
		// 4日留存
		regcount4, count4 := PayRetainedCompute4Web(today, today4, first_pay_id, v)
		// 5日留存
		regcount5, count5 := PayRetainedCompute4Web(today, today5, first_pay_id, v)
		// 6日留存
		regcount6, count6 := PayRetainedCompute4Web(today, today6, first_pay_id, v)
		// 7日留存
		regcount7, count7 := PayRetainedCompute4Web(today, today7, first_pay_id, v)
		// 15日留存
		regcount15, count15 := PayRetainedCompute4Web(today, today15, first_pay_id, v)
		// 30日留存
		regcount30, count30 := PayRetainedCompute4Web(today, today30, first_pay_id, v)
		// 60日留存
		regcount60, count60 := PayRetainedCompute4Web(today, today60, first_pay_id, v)

		day1 := ComputeFloat(count1, regcount1)
		day2 := ComputeFloat(count2, regcount2)
		day3 := ComputeFloat(count3, regcount3)
		day4 := ComputeFloat(count4, regcount4)
		day5 := ComputeFloat(count5, regcount5)
		day6 := ComputeFloat(count6, regcount6)
		day7 := ComputeFloat(count7, regcount7)
		day15 := ComputeFloat(count15, regcount15)
		day30 := ComputeFloat(count30, regcount30)
		day60 := ComputeFloat(count60, regcount60)
		info := new(entity.PayUserRetained4Web)
		date1, _ := utils.Unix(fmt.Sprintf("%s", today))
		info.Date = date1
		info.Channel = ""
		info.SType = 0
		info.PayUserType = v
		info.NewNumber = int64(count)
		info.Day1 = day1 * 100
		info.Day2 = day2 * 100
		info.Day3 = day3 * 100
		info.Day4 = day4 * 100
		info.Day5 = day5 * 100
		info.Day6 = day6 * 100
		info.Day7 = day7 * 100
		info.Day15 = day15 * 100
		info.Day30 = day30 * 100
		info.Day60 = day60 * 100
		addlist = append(addlist, *info)
	}
	for _, item := range addlist {
		AddErr := StatisticsService.AddPayUserRetained(&item)
		if AddErr != nil {
			beego.Error("PayUserRetained fail err: ", AddErr)
		}
	}

}

func PayRetainedCompute4Web(today string, today1 string, payids []string, typeid int, ChannelName ...string) (rcount int64, count int64) {
	// 获取昨日注册，且充值玩家 今日登录
	zrCount := int64(0)  // 之前充值总人数
	zrCount1 := int64(0) // 之前充值，昨日登录人数

	// 获取总充值人数
	m := FindByDate(today1, today1, "ctime", "ctime")
	m["order_status"] = 4
	if typeid == 0 {
		m["userid"] = bson.M{"$in": payids}
	} else {
		m["userid"] = bson.M{"$nin": payids}
	}
	if len(ChannelName) > 0 {
		m["package_id"] = ChannelName[0]
	}
	var cids []string
	Pays.Find(m).Distinct("userid", &cids)
	zrCount = int64(len(cids))
	// 昨日注册充值的玩家人数
	m1 := FindByDate(today, today, "login_time", "login_time")
	m1["userid"] = bson.M{"$in": cids}
	var zrids []string
	LoginLogs.Find(m1).Distinct("userid", &zrids)
	zrCount1 = int64(len(zrids))

	return zrCount, zrCount1
}

// 按渠道统计付费用户留存
func PayUserRetainedByChannel(timestamp int64) {
	nowday := utils.Stamp2Time(timestamp)
	today := nowday.Format("2006-01-02")
	today1 := nowday.AddDate(0, 0, -1).Format("2006-01-02")
	today2 := nowday.AddDate(0, 0, -2).Format("2006-01-02")
	today3 := nowday.AddDate(0, 0, -3).Format("2006-01-02")
	today4 := nowday.AddDate(0, 0, -4).Format("2006-01-02")
	today5 := nowday.AddDate(0, 0, -5).Format("2006-01-02")
	today6 := nowday.AddDate(0, 0, -6).Format("2006-01-02")
	today7 := nowday.AddDate(0, 0, -7).Format("2006-01-02")
	today15 := nowday.AddDate(0, 0, -15).Format("2006-01-02")
	today30 := nowday.AddDate(0, 0, -30).Format("2006-01-02")
	today60 := nowday.AddDate(0, 0, -60).Format("2006-01-02")
	m := bson.M{"status": 1}
	var channellist []entity.ChannelInfo
	err := Channels.
		Find(m).
		All(&channellist)
	if err != nil {
		beego.Error("PayUserRetainedByChannel fail err: ", err)
	}
	// 筛选首充用户
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"order_status": 4,
			},
		},
		{
			"$group": bson.M{
				"_id":   "$userid",
				"count": bson.M{"$sum": 1}, // 添加一个额外的字段以便后续过滤
			},
		},
		{
			"$match": bson.M{
				"count": bson.M{"$eq": 1}, // 过滤num值在1文档
			},
		},
	}
	var res []bson.M
	Pays.Pipe(pipeline).All(&res)
	first_pay_id := make([]string, 0)
	for _, item := range res {
		pid := item["_id"].(string)
		first_pay_id = append(first_pay_id, pid)
	}
	if len(channellist) > 0 {
		date1, _ := utils.Unix(fmt.Sprintf("%s", today))
		retlist := make([]entity.PayUserRetained4Web, 0)
		pay_type := []int{0, 1}
		for _, item := range channellist {
			channel := item.Name
			name1 := item.Name1

			for _, v := range pay_type {
				var ids []string // 注册ID
				m := FindByDate(today, today, "ctime", "ctime")
				m["order_status"] = 4
				if v == 0 {
					m["userid"] = bson.M{"$in": first_pay_id}
				} else {
					m["userid"] = bson.M{"$nin": first_pay_id}
				}
				m["package_id"] = channel
				Pays.Find(m).Distinct("_id", &ids)
				count := len(ids)

				// 1日留存
				regcount1, count1 := PayRetainedCompute4Web(today, today1, first_pay_id, v, channel)
				// 2日留存
				regcount2, count2 := PayRetainedCompute4Web(today, today2, first_pay_id, v, channel)
				// 3日留存
				regcount3, count3 := PayRetainedCompute4Web(today, today3, first_pay_id, v, channel)
				// 4日留存
				regcount4, count4 := PayRetainedCompute4Web(today, today4, first_pay_id, v, channel)
				// 5日留存
				regcount5, count5 := PayRetainedCompute4Web(today, today5, first_pay_id, v, channel)
				// 6日留存
				regcount6, count6 := PayRetainedCompute4Web(today, today6, first_pay_id, v, channel)
				// 7日留存
				regcount7, count7 := PayRetainedCompute4Web(today, today7, first_pay_id, v, channel)
				// 15日留存
				regcount15, count15 := PayRetainedCompute4Web(today, today15, first_pay_id, v, channel)
				// 30日留存
				regcount30, count30 := PayRetainedCompute4Web(today, today30, first_pay_id, v, channel)
				// 60日留存
				regcount60, count60 := PayRetainedCompute4Web(today, today60, first_pay_id, v, channel)

				day1 := ComputeFloat(count1, regcount1)
				day2 := ComputeFloat(count2, regcount2)
				day3 := ComputeFloat(count3, regcount3)
				day4 := ComputeFloat(count4, regcount4)
				day5 := ComputeFloat(count5, regcount5)
				day6 := ComputeFloat(count6, regcount6)
				day7 := ComputeFloat(count7, regcount7)
				day15 := ComputeFloat(count15, regcount15)
				day30 := ComputeFloat(count30, regcount30)
				day60 := ComputeFloat(count60, regcount60)
				info := new(entity.PayUserRetained4Web)
				info.Date = date1
				info.SType = 1
				info.PayUserType = v
				info.Channel = channel
				info.Channel1 = name1
				info.NewNumber = int64(count)
				info.Day1 = day1 * 100
				info.Day2 = day2 * 100
				info.Day3 = day3 * 100
				info.Day4 = day4 * 100
				info.Day5 = day5 * 100
				info.Day6 = day6 * 100
				info.Day7 = day7 * 100
				info.Day15 = day15 * 100
				info.Day30 = day30 * 100
				info.Day60 = day60 * 100
				retlist = append(retlist, *info)
			}

		}
		for _, user := range retlist {
			AddErr := StatisticsService.AddPayUserRetained(&user)
			if AddErr != nil {
				beego.Error("PayUserRetainedByChannel fail err: ", AddErr)
			}
		}
	}
}

// 商品购买统计
func GoodsBuyData(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")

	s := fmt.Sprintf("%s 00:00:00", today)
	startTime := utils.Str2TimeZone(s, locationName)
	e := fmt.Sprintf("%s 23:59:59", today)
	endTime := utils.Str2TimeZone(e, locationName)
	// 获取当天注册用户
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"ctime":            bson.M{"$gte": startTime, "$lt": endTime},
				"robot":            false,
				"simulation_robot": false,
			},
		},
		{
			"$group": bson.M{
				"_id": "$_id",
			},
		},
	}
	// operations := []bson.M{m, n}
	result := []bson.M{}
	pipe := PlayerUsers.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("GoodsBuyData fail err: ", err)
	}
	userid := make([]string, 0)
	for _, item := range result {
		id := item["_id"].(string)
		userid = append(userid, id)
	}
	addlist := make([]entity.GoodsBuyData, 0)
	date1, _ := utils.Unix(fmt.Sprintf("%s", today))
	// 统计老玩家购买商品
	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"ctime":  bson.M{"$gte": startTime, "$lt": endTime},
				"userid": bson.M{"$nin": userid},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"shop_type": "$shop_type",
					"shop_name": "$shop_name",
					"amount":    "$amount",
				},
				"order_count": bson.M{"$sum": 1},
				"user_count":  bson.M{"$addToSet": "$userid"},
				"successfulOrders": bson.M{"$sum": bson.M{"$cond": []interface{}{
					bson.M{"$eq": []interface{}{"$order_status", 4}}, 1, nil,
				}}},
				"successfulUsers": bson.M{"$addToSet": bson.M{"$cond": []interface{}{
					bson.M{"$eq": []interface{}{"$order_status", 4}}, "$userid", nil,
				}}},
			},
		},
		{
			"$project": bson.M{
				"_id":              0,
				"shop_type":        "$_id.shop_type",
				"shop_name":        "$_id.shop_name",
				"amount":           "$_id.amount",
				"order_count":      1,
				"user_count":       "$user_count",
				"successfulOrders": "$successfulOrders",
				"successfulUsers":  "$successfulUsers",
			},
		},
		{"$sort": bson.M{"shop_type": -1}},
	}
	result1 := []bson.M{}
	pipe1 := Pays.Pipe(pipeline1)
	err1 := pipe1.All(&result1)
	if err1 != nil {
		beego.Error("GoodsBuyData fail err: ", err1)
	}
	if len(result1) > 0 {
		for _, item := range result1 {
			userlist := item["user_count"].([]interface{})
			ordercount := item["order_count"].(int)
			typeId := item["shop_type"].(int)
			ShopName := item["shop_name"].(string)
			amount := item["amount"].(int)
			sOrderCount := item["successfulOrders"].(int)
			sUserlist := item["successfulUsers"].([]interface{})
			info := new(entity.GoodsBuyData)
			info.Date = date1
			info.PlayerTypes = 2
			info.ShopType = int32(typeId)
			info.ShopName = ShopName
			info.Amount = uint32(amount)
			info.TotalPullOrder = int64(ordercount)
			info.SuccessfulOrder = int64(sOrderCount)
			info.PullNumber = int64(len(userlist))
			// buyCount := len(sUserlist)
			buyCount := 0
			for _, item := range sUserlist {
				if item != nil {
					buyCount++
				}
			}
			info.BuyNumber = int64(buyCount)
			addlist = append(addlist, *info)
		}
	}

	// 统计新玩家购买商品
	pipeline2 := []bson.M{
		{
			"$match": bson.M{
				"ctime":  bson.M{"$gte": startTime, "$lt": endTime},
				"userid": bson.M{"$in": userid},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"shop_type": "$shop_type",
					"shop_name": "$shop_name",
					"amount":    "$amount",
				},
				"order_count": bson.M{"$sum": 1},
				"user_count":  bson.M{"$addToSet": "$userid"},
				"successfulOrders": bson.M{"$sum": bson.M{"$cond": []interface{}{
					bson.M{"$eq": []interface{}{"$order_status", 4}}, 1, nil,
				}}},
				"successfulUsers": bson.M{"$addToSet": bson.M{"$cond": []interface{}{
					bson.M{"$eq": []interface{}{"$order_status", 4}}, "$userid", nil,
				}}},
			},
		},
		{
			"$project": bson.M{
				"_id":              0,
				"shop_type":        "$_id.shop_type",
				"shop_name":        "$_id.shop_name",
				"amount":           "$_id.amount",
				"order_count":      1,
				"user_count":       "$user_count",
				"successfulOrders": "$successfulOrders",
				"successfulUsers":  "$successfulUsers",
			},
		},
		{"$sort": bson.M{"shop_type": -1}},
	}
	result2 := []bson.M{}
	pipe2 := Pays.Pipe(pipeline2)
	err2 := pipe2.All(&result2)
	if err2 != nil {
		beego.Error("GoodsBuyData fail err: ", err2)
	}
	if len(result2) > 0 {
		for _, item := range result2 {
			userlist := item["user_count"].([]interface{})
			ordercount := item["order_count"].(int)
			typeId := item["shop_type"].(int)
			ShopName := item["shop_name"].(string)
			amount := item["amount"].(int)
			sOrderCount := item["successfulOrders"].(int)
			sUserlist := item["successfulUsers"].([]interface{})
			info := new(entity.GoodsBuyData)
			info.Date = date1
			info.PlayerTypes = 1
			info.ShopType = int32(typeId)
			info.ShopName = ShopName
			info.Amount = uint32(amount)
			info.TotalPullOrder = int64(ordercount)
			info.SuccessfulOrder = int64(sOrderCount)
			info.PullNumber = int64(len(userlist))
			buyCount := 0
			for _, item := range sUserlist {
				if item != nil {
					buyCount++
				}
			}
			info.BuyNumber = int64(buyCount)
			addlist = append(addlist, *info)
		}
	}

	for _, info := range addlist {
		AddErr := StatisticsService.AddGoodsBuy(&info)
		if AddErr != nil {
			beego.Error("GoodsBuyData fail err: ", AddErr)
		}
	}
}

// AD上报统计
func AdReportStatistics(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")

	s := fmt.Sprintf("%s 00:00:00", today)
	startTime, _ := utils.Str2Local(s, location)
	e := fmt.Sprintf("%s 23:59:59", today)
	endTime, _ := utils.Str2Local(e, location)

	m := bson.M{"status": 1}
	var channellist []entity.ChannelInfo
	err := Channels.
		Find(m).
		All(&channellist)
	if err != nil {
		beego.Error("AdReportStatistics fail err: ", err)
	}
	list := make([]entity.AdReportData, 0)
	if len(channellist) > 0 {
		date1, _ := utils.Unix(fmt.Sprintf("%s", today))
		for _, item := range channellist {
			info := new(entity.AdReportData)
			name := item.Name // 渠道名称
			name1 := item.Name1
			loginRes := AdReportByTypeCompute(startTime, endTime, name, "login")
			if len(loginRes) > 0 {
				for _, item := range loginRes {
					loginCount := item["count"].(int)
					loginToday := item["today"].(int)
					loginCross := item["cross"].(int)
					loginFailed := item["failed"].(int)
					info.LoginCount = int64(loginCount)
					info.LoginCountDay = int64(loginToday)
					info.LoginCountDay1 = int64(loginCross)
					info.FailLoginCount = int64(loginFailed)
				}
			}
			registerRes := AdReportByTypeCompute(startTime, endTime, name, "register")
			if len(registerRes) > 0 {
				for _, item := range registerRes {
					registerCount := item["count"].(int)
					registerToday := item["today"].(int)
					registerCross := item["cross"].(int)
					registerFailed := item["failed"].(int)
					info.RegisterCount = int64(registerCount)
					info.RegisterCountDay = int64(registerToday)
					info.RegisterCountDay1 = int64(registerCross)
					info.FailRegisterCount = int64(registerFailed)
				}
			}
			depositRes := AdReportByTypeCompute(startTime, endTime, name, "deposit")
			if len(depositRes) > 0 {
				for _, item := range depositRes {
					depositCount := item["count"].(int)
					depositToday := item["today"].(int)
					depositCross := item["cross"].(int)
					depositFailed := item["failed"].(int)
					info.DepositCount = int64(depositCount)
					info.DepositCountDay = int64(depositToday)
					info.DepositCountDay1 = int64(depositCross)
					info.FailDepositCount = int64(depositFailed)
				}
			}
			info.Date = date1
			info.Channel = name
			info.Channel1 = name1
			list = append(list, *info)
		}
	}
	for _, item := range list {
		err := StatisticsService.AddOrUpdateAdReport(&item)
		if err != nil {
			beego.Error("AdReportStatistics fail err: ", err)
		}
	}
}

// 根据AD上报类型计算
func AdReportByTypeCompute(startTime, endTime int64, ChannelName, TypeName string) []bson.M {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"ctime":     bson.M{"$gte": startTime, "$lt": endTime},
				"bundle_id": ChannelName,
				"type":      TypeName,
			},
		},
		{
			"$group": bson.M{
				"_id":   "$bundle_id",
				"count": bson.M{"$sum": 1},
				"today": bson.M{"$sum": bson.M{"$cond": []interface{}{
					bson.M{"$and": []interface{}{
						bson.M{"$gte": []interface{}{"$ctime", startTime}},
						bson.M{"$lt": []interface{}{"$ctime", endTime}},
						bson.M{"$gte": []interface{}{"$etime", startTime}},
						bson.M{"$lt": []interface{}{"$etime", endTime}},
						bson.M{"$eq": []interface{}{"$code", 0}},
					}}, 1, nil,
				}}},
				"cross": bson.M{"$sum": bson.M{"$cond": []interface{}{
					bson.M{"$and": []interface{}{
						bson.M{"$gte": []interface{}{"$ctime", startTime}},
						bson.M{"$lt": []interface{}{"$ctime", endTime}},
						bson.M{"$gte": []interface{}{"$etime", endTime}},
						bson.M{"$eq": []interface{}{"$code", 0}},
					}}, 1, nil,
				}}},
				"failed": bson.M{"$sum": bson.M{"$cond": []interface{}{
					bson.M{"$eq": []interface{}{"$code", 1}}, 1, nil,
				}}},
			},
		},
		{
			"$project": bson.M{
				"_id":    "$_id",
				"count":  "$count",
				"today":  "$today",
				"cross":  "$cross",
				"failed": "$failed",
			},
		},
	}
	result := []bson.M{}
	pipe := AdReports.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("AdReportByTypeCompute fail err: ", err)
	}
	return result
}

/*
Author：CC
Title：短信统计
*/
func SMSData(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")

	s := fmt.Sprintf("%s 00:00:00", today)
	startTime, _ := utils.Str2Local(s, location)
	e := fmt.Sprintf("%s 23:59:59", today)
	endTime, _ := utils.Str2Local(e, location)

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": startTime, "$lt": endTime},
			},
		},
		{
			"$group": bson.M{
				"_id":   nil,
				"count": bson.M{"$sum": 1},
				"success": bson.M{"$sum": bson.M{"$cond": []interface{}{
					bson.M{"$eq": []interface{}{"$reslut", true}}, 1, nil,
				}}},
			},
		},
		{
			"$project": bson.M{
				"_id":     nil,
				"count":   "$count",
				"success": "$success",
			},
		},
	}
	result := []bson.M{}
	pipe := LogSmsRecords.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("SMSData fail err: ", err)
	}
	total := 0
	successCount := 0
	useCount := 0
	if len(result) > 0 {
		for _, item := range result {
			total = item["count"].(int)
			successCount = item["success"].(int)
		}
	}

	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": startTime, "$lt": endTime},
			},
		},
		{
			"$group": bson.M{
				"_id":   nil,
				"count": bson.M{"$sum": 1},
			},
		},
	}
	result1 := []bson.M{}
	pipe1 := LogUserSmsRecords.Pipe(pipeline1)
	err = pipe1.All(&result1)
	if err != nil {
		beego.Error("SMSData fail err: ", err)
	}
	if len(result1) > 0 {
		for _, item := range result1 {
			useCount = item["count"].(int)
		}
	}

	date1, _ := utils.Unix(fmt.Sprintf("%s", today))
	info := new(entity.SmsData)
	info.Date = date1
	info.TotalSendSMS = int64(total)
	info.SendCount = int64(successCount)
	info.UseSMS = int64(useCount)

	addErr := StatisticsService.AddOrUpdateSMS(info)
	if addErr != nil {
		beego.Error("SMSData fail err: ", addErr)
	}
}

/*
Author：CC
Title：充值来源
*/
func PaySourceData(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")

	s := fmt.Sprintf("%s 00:00:00", today)
	startTime := utils.Str2TimeZone(s, locationName)
	e := fmt.Sprintf("%s 23:59:59", today)
	endTime := utils.Str2TimeZone(e, locationName)

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"ctime":        bson.M{"$gte": startTime, "$lt": endTime},
				"order_status": 4,
				"gtype":        bson.M{"$ne": 0},
			},
		},
		{
			"$group": bson.M{
				"_id": "$gtype",
				"Num": bson.M{
					"$sum": 1,
				},
				"Amount": bson.M{
					"$sum": "$amount",
				},
			},
		},
	}
	result := []bson.M{}
	pipe := Pays.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("PaySourceData fail err: ", err)
	}
	info := new(entity.PaySourceData)
	date1, _ := utils.Unix(fmt.Sprintf("%s", today))
	info.Date = date1
	for _, item := range result {
		id := item["_id"].(int)
		amount := item["Amount"].(int)
		num := item["Num"].(int)
		switch id {
		case 1:
			// TP
			info.TPAmount = int64(amount)
			info.TPNumber = int64(num)
		case 2:
			//LHD
			info.LHDAmount = int64(amount)
			info.LHDNumber = int64(num)
		case 3:
			// 7UP
			info.UPAmount = int64(amount)
			info.UPNumber = int64(num)
		case 4:
			// RM
			info.RMAmount = int64(amount)
			info.RMNumber = int64(num)
		case 5:
			// AK47
			info.AKAmount = int64(amount)
			info.AKNumber = int64(num)
		case 6:
			// JOKER
			info.JOKERAmount = int64(amount)
			info.JOKERNumber = int64(num)
		case 7:
			// CRASH
			info.CRASHAmount = int64(amount)
			info.CRASHNumber = int64(num)
		case 8:
			// AB
			info.ABAmount = int64(amount)
			info.ABNumber = int64(num)
		case 9:
			// CP
			info.CPAmount = int64(amount)
			info.CPNumber = int64(num)
		case 10:
			// 飞机
			info.FJAmount = int64(amount)
			info.FJNumber = int64(num)
		case 12:
			// RUMMY双人
			info.RMTwoAmount = int64(amount)
			info.RMTwoNumber = int64(num)
		}
	}
	addErr := StatisticsService.AddOrUpdatePaySource(info)
	if addErr != nil {
		beego.Error("ShareStatistics fail err: ", addErr)
	}
}

// 对战房埋点统计
func BattleRoomData(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
	statrTimestamp := startTime.Unix()
	endTimestamp := endTime.Unix()

	create_count := 0
	create_num := 0
	join_count := 0
	join_num := 0
	s_enter_count := 0
	s_enter_num := 0
	tp_enter_count := 0
	tp_enter_num := 0
	rm_enter_count := 0
	rm_enter_num := 0
	ab_enter_count := 0
	ab_enter_num := 0

	s_join_count := 0
	s_join_num := 0
	tp_join_count := 0
	tp_join_num := 0
	rm_join_count := 0
	rm_join_num := 0
	ab_join_count := 0
	ab_join_num := 0
	// 创建点击
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": statrTimestamp, "$lt": endTimestamp},
				"type":  bson.M{"$regex": "createRoom", "$options": "i"},
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Count": bson.M{
					"$sum": 1,
				},
			},
		},
	}
	result := []bson.M{}
	pipe := ButtonClicks.Pipe(pipeline)
	err := pipe.All(&result)

	uniqueData := make(map[string]bool)
	for _, item := range result {
		id := item["_id"].(string)
		number := item["Count"].(int)
		create_count += number
		uniqueData[id] = true
	}
	create_num = len(result)
	// 加入点击
	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": statrTimestamp, "$lt": endTimestamp},
				"type": bson.M{
					"$regex":   fmt.Sprintf("\\b%s\\b", "JoinRoom"),
					"$options": "i",
				},
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Count": bson.M{
					"$sum": 1,
				},
			},
		},
	}
	result1 := []bson.M{}
	pipe1 := ButtonClicks.Pipe(pipeline1)
	err = pipe1.All(&result1)
	if err != nil {
		beego.Error("BattleRoomData fail err: ", err)
	}
	for _, item := range result1 {
		id := item["_id"].(string)
		number := item["Count"].(int)
		join_count += number
		uniqueData[id] = true
	}
	join_num = len(result1)
	// 成功创建房间，进入房间
	pipeline2 := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": statrTimestamp, "$lt": endTimestamp},
				"type":  bson.M{"$regex": "creatRoom_game", "$options": "i"},
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Count": bson.M{
					"$sum": 1,
				},
			},
		},
	}
	result2 := []bson.M{}
	pipe2 := ButtonClicks.Pipe(pipeline2)
	pipe2.All(&result2)
	for _, item := range result2 {
		number := item["Count"].(int)
		s_enter_count += number
	}
	s_enter_num = len(result2)
	// TP
	pipeline3 := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": statrTimestamp, "$lt": endTimestamp},
				"type":  bson.M{"$regex": "creatRoom_game:teenpatti", "$options": "i"},
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Count": bson.M{
					"$sum": 1,
				},
			},
		},
	}
	result3 := []bson.M{}
	pipe3 := ButtonClicks.Pipe(pipeline3)
	pipe3.All(&result3)
	for _, item := range result3 {
		number := item["Count"].(int)
		tp_enter_count += number
	}
	tp_enter_num = len(result3)

	// RM
	pipeline4 := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": statrTimestamp, "$lt": endTimestamp},
				"type":  bson.M{"$regex": "creatRoom_game:rummy", "$options": "i"},
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Count": bson.M{
					"$sum": 1,
				},
			},
		},
	}
	result4 := []bson.M{}
	pipe4 := ButtonClicks.Pipe(pipeline4)
	pipe4.All(&result4)
	for _, item := range result4 {
		number := item["Count"].(int)
		rm_enter_count += number
	}
	rm_enter_num = len(result4)

	// AB
	pipeline5 := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": statrTimestamp, "$lt": endTimestamp},
				"type":  bson.M{"$regex": "creatRoom_game:ABAR", "$options": "i"},
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Count": bson.M{
					"$sum": 1,
				},
			},
		},
	}
	result5 := []bson.M{}
	pipe5 := ButtonClicks.Pipe(pipeline5)
	pipe5.All(&result5)
	for _, item := range result5 {
		number := item["Count"].(int)
		ab_enter_count += number
	}
	ab_enter_num = len(result5)

	// 成功加入房间，进入房间
	pipeline6 := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": statrTimestamp, "$lt": endTimestamp},
				"type":  bson.M{"$regex": "joinRoom_game", "$options": "i"},
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Count": bson.M{
					"$sum": 1,
				},
			},
		},
	}
	result6 := []bson.M{}
	pipe6 := ButtonClicks.Pipe(pipeline6)
	pipe6.All(&result6)
	for _, item := range result6 {
		number := item["Count"].(int)
		s_join_count += number
	}
	s_join_num = len(result6)
	// TP
	pipeline7 := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": statrTimestamp, "$lt": endTimestamp},
				"type":  bson.M{"$regex": "joinRoom_game:teenpatti", "$options": "i"},
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Count": bson.M{
					"$sum": 1,
				},
			},
		},
	}
	result7 := []bson.M{}
	pipe7 := ButtonClicks.Pipe(pipeline7)
	pipe7.All(&result7)
	for _, item := range result7 {
		number := item["Count"].(int)
		tp_join_count += number
	}
	tp_join_num = len(result7)

	// RM
	pipeline8 := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": statrTimestamp, "$lt": endTimestamp},
				"type":  bson.M{"$regex": "joinRoom_game:rummy", "$options": "i"},
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Count": bson.M{
					"$sum": 1,
				},
			},
		},
	}
	result8 := []bson.M{}
	pipe8 := ButtonClicks.Pipe(pipeline8)
	pipe8.All(&result8)
	for _, item := range result8 {
		number := item["Count"].(int)
		rm_join_count += number
	}
	rm_join_num = len(result8)

	// AB
	pipeline9 := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": statrTimestamp, "$lt": endTimestamp},
				"type":  bson.M{"$regex": "joinRoom_game:ABAR", "$options": "i"},
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Count": bson.M{
					"$sum": 1,
				},
			},
		},
	}
	result9 := []bson.M{}
	pipe9 := ButtonClicks.Pipe(pipeline9)
	pipe9.All(&result9)
	for _, item := range result9 {
		number := item["Count"].(int)
		ab_join_count += number
	}
	ab_join_num = len(result9)

	click_num := len(uniqueData)
	click_count := create_count + join_count
	date1, _ := utils.Unix(fmt.Sprintf("%s", today))
	info := new(entity.BattleRoomData)
	info.Date = date1
	info.ClicksCount = int64(click_count)
	info.TotalNumber = int64(click_num)
	info.CreateRoomCount = int64(create_count)
	info.CreateRoomNumber = int64(create_num)
	info.SuccessEnterCount = int64(s_enter_count)
	info.SuccessEnterNumber = int64(s_enter_num)
	info.TPEnterCount = int64(tp_enter_count)
	info.TPEnterNumber = int64(tp_enter_num)
	info.RMEnterCount = int64(rm_enter_count)
	info.RMEnterNumber = int64(rm_enter_num)
	info.ABEnterCount = int64(ab_enter_count)
	info.ABEnterNumber = int64(ab_enter_num)
	info.JoinRoomCount = int64(join_count)
	info.JoinRoomNumber = int64(join_num)
	info.SuccessJoinCount = int64(s_join_count)
	info.SuccessJoinNumber = int64(s_join_num)
	info.TPJoinCount = int64(tp_join_count)
	info.TPJoinNumber = int64(tp_join_num)
	info.RMJoinCount = int64(rm_join_count)
	info.RMJoinNumber = int64(rm_join_num)
	info.ABJoinCount = int64(ab_join_count)
	info.ABJoinNumber = int64(ab_join_num)

	addErr := StatisticsService.AddOrUpdateBattleRoom(info)
	if addErr != nil {
		beego.Error("BattleRoomData fail err: ", addErr)
	}
}

// 彩票活动统计
func LotteryActivityData(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
	statrTimestamp := startTime.Unix()
	endTimestamp := endTime.Unix()
	info := new(entity.LotteryActivity)
	// 获取进入人数
	query := bson.M{}
	query["otype"] = 1
	query["ctime"] = bson.M{"$gte": statrTimestamp, "$lt": endTimestamp}
	c_user_arr := make([]string, 0)
	LogActivityMonitors.Find(query).Distinct("userid", &c_user_arr)
	info.ClickNumber = int64(len(c_user_arr))
	// 刮奖人数
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": statrTimestamp, "$lt": endTimestamp},
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
			},
		},
	}
	result := []bson.M{}
	pipe := ScratchTicketss.Pipe(pipeline)
	pipe.All(&result)
	if len(result) > 0 {
		info.GJNumber = int64(len(result))
	}
	// 中奖人数
	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": statrTimestamp, "$lt": endTimestamp},
				"win":   true,
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
			},
		},
	}
	result1 := []bson.M{}
	pipe1 := ScratchTicketss.Pipe(pipeline1)
	pipe1.All(&result1)
	if len(result1) > 0 {
		info.HJNumber = int64(len(result1))
	}

	// 计算消耗
	pipeline2 := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": statrTimestamp, "$lt": endTimestamp},
			},
		},
		{
			"$group": bson.M{
				"_id":    nil,
				"total1": bson.M{"$sum": "$jackpot1"},
				"total2": bson.M{"$sum": "$jackpot2"},
				"total3": bson.M{"$sum": "$jackpot3"},
				"total4": bson.M{"$sum": "$jackpot4"},
				"total5": bson.M{"$sum": "$jackpot5"},
				"total":  bson.M{"$sum": "$cost_bonus"},
			},
		},
	}
	// result2 := []bson.M
	var result2 []bson.M
	pipe2 := ScratchTicketss.Pipe(pipeline2)
	pipe2.All(&result2)
	if len(result2) > 0 {
		info.FirstPrize = result2[0]["total1"].(int64)
		info.SecondPrize = result2[0]["total2"].(int64)
		info.ThirdPrize = result2[0]["total3"].(int64)
		info.FourthPrize = result2[0]["total4"].(int64)
		info.FifthPrize = result2[0]["total5"].(int64)
		info.ExpendBouns = result2[0]["total"].(int64)
	}
	date1, _ := utils.Unix(fmt.Sprintf("%s", today))
	info.Date = date1
	addErr := ActivityService.AddOrUpdateLottery(info)
	if addErr != nil {
		beego.Error("LotteryActivityData fail err: ", addErr)
	}
}

// 玩家游戏局数统计
func UserGameData(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)

	// // 获取有金币流水变化的用户
	// var logId []string
	// m := bson.M{}
	// m["ctime"] = bson.M{"$gte": startTime, "$lte": endTime}
	// LogWaters.Find(m).Distinct("userid", &logId)
	// 获取当天有登录的用户id
	var logId []string
	m := bson.M{}
	m["login_time"] = bson.M{"$gte": startTime, "$lte": endTime}
	LoginLogs.Find(m).Distinct("userid", &logId)
	list := make([]entity.UserGameData, 0)
	// slist := make([]entity.GameStrategy, 0)
	for _, userid := range logId {
		m := bson.M{}
		m["players"] = bson.M{
			"$regex":   fmt.Sprintf("\\b%s\\b", userid),
			"$options": "i",
		}
		m["begin_time"] = bson.M{"$gte": startTime.Unix(), "$lte": endTime.Unix()}
		var dtlist []entity.Detail
		Details.Find(m).All(&dtlist)
		map_dt := make(map[int32]*entity.UserGameData, 0)
		// stg_dt := make(map[string]*entity.GameStrategy, 0)
		if len(dtlist) > 0 {
			for _, item := range dtlist {
				stat, ok := map_dt[item.Gtype]
				if !ok {
					stat = &entity.UserGameData{
						UserId: userid,
						Gtype:  int64(item.Gtype),
					}
					// map_dt = append(crashYhwmIds, user.Userid)
				}
				switch item.Gtype {
				case 1:
					for _, v := range item.TPDetail {
						if v.UserId == userid {
							stat.Number++
							stat.Bets += v.Bet
							if v.Score > 0 {
								stat.WinNumber++
								// 包含下注,需要去掉下注金额
								amount := v.Score - v.Bet
								stat.WinBets += amount
							} else if v.Score < 0 {
								stat.LoseNumber++
								stat.LoseBets += v.Score
							}
							// if item.PlayerStageId != 0 {
							// 	str := strconv.Itoa(item.PlayerStageId)
							// 	stg, ok := stg_dt[str]
							// 	if !ok {
							// 		stg = &entity.GameStrategy{
							// 			UserId:     userid,
							// 			Gtype:      int64(item.Gtype),
							// 			StrategyId: str,
							// 		}
							// 	}
							// 	stg.EffectiveNumber++
							// 	income := v.Score
							// 	if v.Score > 0 {
							// 		income = v.Score - v.Bet
							// 	}
							// 	stg.EffectiveIncome += income
							// 	stg_dt[str] = stg
							// }
						}
					}
				case 2:
					for _, v := range item.LHDetail.UserDetail {
						if v.Userid == userid {
							stat.Number++
							stat.Bets += item.LHDetail.Bets
							if v.Win > 0 {
								stat.WinNumber++
								// 包含下注,需要去掉下注金额
								amount := v.Win - v.CashMingTax - v.BonusAnTax
								stat.WinBets += amount
							} else if v.Win < 0 {
								stat.LoseNumber++
								stat.LoseBets += v.Win
							}
							// if item.LHDetail.StrategyId != "" {
							// 	str := item.LHDetail.StrategyId
							// 	stg, ok := stg_dt[str]
							// 	if !ok {
							// 		stg = &entity.GameStrategy{
							// 			UserId:     userid,
							// 			Gtype:      int64(item.Gtype),
							// 			StrategyId: str,
							// 		}
							// 	}
							// 	stg.EnterNumber++
							// 	if item.LHDetail.StrategyTigger == true {
							// 		stg.EffectiveNumber++
							// 		income := v.Win
							// 		stg.EffectiveIncome += income
							// 	}
							// 	stg_dt[str] = stg
							// }
						}
					}
				case 3:
					for _, v := range item.UPDetail.UserDetail {
						if v.Userid == userid {
							stat.Number++
							if v.Win > 0 {
								stat.WinNumber++
								// 包含下注,需要去掉下注金额
								amount := v.Win - v.CashMingTax - v.BonusAnTax
								stat.WinBets += amount
							} else if v.Win < 0 {
								stat.LoseNumber++
								stat.LoseBets += v.Win
							}
						}
					}
				case 4:
					for _, v := range item.RMDetail {
						if v.UserId == userid {
							stat.Number++
							// stat.Bets += v.Score
							if v.Score > 0 {
								stat.WinNumber++
								// 包含下注,需要去掉下注金额
								amount := v.Score - v.CashMingTax - v.BonusAnTax
								stat.WinBets += amount
							} else if v.Score < 0 {
								num := v.Score
								absNum := int64(math.Abs(float64(num)))
								stat.Bets += absNum
								stat.LoseNumber++
								stat.LoseBets += v.Score
							}
						}
					}
				case 5:
					for _, v := range item.AK47Detail {
						if v.UserId == userid {
							stat.Number++
							stat.Bets += v.Bet
							if v.Score > 0 {
								stat.WinNumber++
								// 包含下注,需要去掉下注金额
								amount := v.Score - v.Bet
								stat.WinBets += amount
							} else if v.Score < 0 {
								stat.LoseNumber++
								stat.LoseBets += v.Score
							}
						}
					}
				case 6:
					for _, v := range item.JOKERDetail {
						if v.UserId == userid {
							stat.Number++
							stat.Bets += v.Bet
							if v.Score > 0 {
								stat.WinNumber++
								// 包含下注,需要去掉下注金额
								amount := v.Score - v.Bet
								stat.WinBets += amount
							} else if v.Score < 0 {
								stat.LoseNumber++
								stat.LoseBets += v.Score
							}
						}
					}
				case 7:
					for _, v := range item.CRASHDetail.UserDetail {
						if v.Userid == userid {
							stat.Number++
							stat.Bets += item.CRASHDetail.Bets
							if v.Win > 0 {
								stat.WinNumber++
								stat.WinBets += v.Win
							} else if v.Win < 0 {
								stat.LoseNumber++
								stat.LoseBets += v.Win
							}
						}
					}
				case 8:
					for _, v := range item.ABDetail.UserDetail {
						if v.Userid == userid {
							stat.Number++
							stat.Bets += item.ABDetail.Bets
							if v.Win > 0 {
								stat.WinNumber++
								stat.WinBets += v.Win
							} else if v.Win < 0 {
								stat.LoseNumber++
								stat.LoseBets += v.Win
							}
						}
					}
				case 9:
					for _, v := range item.CPDetail.UserDetail {
						if v.Userid == userid {
							stat.Number++
							stat.Bets += item.CPDetail.Bets
							if v.Win > 0 {
								stat.WinNumber++
								// 包含下注,需要去掉下注金额
								amount := v.Win - v.CashAnTax - v.BonusAnTax
								stat.WinBets += amount
							} else if v.Win < 0 {
								stat.LoseNumber++
								stat.LoseBets += v.Win
							}
						}
					}
				case 10:
					for _, v := range item.CRASHDetail.UserDetail {
						if v.Userid == userid {
							stat.Number++
							stat.Bets += item.CRASHDetail.Bets
							if v.Win > 0 {
								stat.WinNumber++
								stat.WinBets += v.Win
							} else if v.Win < 0 {
								stat.LoseNumber++
								stat.LoseBets += v.Win
							}
						}
					}
				case 11:
					for _, v := range item.RBDetail.UserDetail {
						if v.Userid == userid {
							stat.Number++
							stat.Bets += item.RBDetail.Bets
							if v.Win > 0 {
								stat.WinNumber++
								stat.WinBets += v.Win - v.CashAnTax - v.BonusAnTax
							} else if v.Win < 0 {
								stat.LoseNumber++
								stat.LoseBets += v.Win
							}
						}
					}
				case 12:
					for _, v := range item.RMDetail {
						if v.UserId == userid {
							stat.Number++
							// stat.Bets += v.Score
							if v.Score > 0 {
								stat.WinNumber++
								stat.WinBets += v.Score
							} else if v.Score < 0 {
								num := v.Score
								absNum := int64(math.Abs(float64(num)))
								stat.Bets += absNum
								stat.LoseNumber++
								stat.LoseBets += v.Score
							}
						}
					}
				}
				map_dt[item.Gtype] = stat
			}

			// 游戏时长统计
			pipeline2 := []bson.M{
				{
					"$match": bson.M{
						"userid": userid,
						"ctime":  bson.M{"$gte": startTime.Unix(), "$lte": endTime.Unix()},
					},
				},
				{
					"$group": bson.M{
						"_id": "$gtype",
						"time": bson.M{
							"$sum": "$time",
						},
					},
				},
			}
			result2 := []bson.M{}
			LogGameTimes.Pipe(pipeline2).All(&result2)
			for _, g := range result2 {
				gtypeid := g["_id"].(int)
				stat, ok := map_dt[int32(gtypeid)]
				if ok {
					time := g["time"].(int64)
					stat.GameTime = time
					map_dt[int32(gtypeid)] = stat
				}
			}
			if len(map_dt) > 0 {
				for _, m := range map_dt {
					m.Ctime = timestamp
					list = append(list, *m)
				}
			}
			// if len(stg_dt) > 0 {
			// 	for _, m := range stg_dt {
			// 		m.Ctime = timestamp
			// 		slist = append(slist, *m)
			// 	}
			// }
		}
		// pipeline := []bson.M{
		// 	{
		// 		"$match": bson.M{
		// 			"players": bson.M{
		// 				"$regex":   fmt.Sprintf("\\b%s\\b", userid),
		// 				"$options": "i",
		// 			},
		// 		},
		// 	},
		// 	{
		// 		"$group": bson.M{
		// 			"_id": "$gtype",
		// 			"num": bson.M{
		// 				"$sum": 1,
		// 			},
		// 		},
		// 	},
		// }
		// result := []bson.M{}
		// err := Details.Pipe(pipeline).All(&result)
		// if err != nil {
		// 	beego.Error("UserGameData fail err: ", err)
		// }
		// for _, item := range result {
		// 	id := item["_id"].(int)
		// 	num := item["num"].(int)
		// 	info := new(entity.UserGameData)
		// 	info.UserId = userid
		// 	info.Gtype = int64(id)
		// 	info.Number = int64(num)
		// 	list = append(list, *info)
		// }

		// 外接游戏局数
		pipeline1 := []bson.M{
			{
				"$match": bson.M{
					"user_id": userid,
					"amount":  bson.M{"$ne": 0},
					"ctime":   bson.M{"$gte": startTime.Unix(), "$lte": endTime.Unix()},
				},
			},
			{
				"$group": bson.M{
					"_id":   "$game_id",
					"count": bson.M{"$addToSet": "$round_id"},
				},
			},
		}
		result1 := []bson.M{}
		err := NsqLogExternalBets.Pipe(pipeline1).All(&result1)
		if err != nil {
			beego.Error("UserGameData fail err: ", err)
		}
		for _, item := range result1 {
			id := item["_id"].(int)
			arr_num := item["count"].([]interface{})
			map_num := make(map[string]bool, 0)
			for _, v := range arr_num {
				rid := v.(string)
				map_num[rid] = true
			}
			num := len(map_num)

			info := new(entity.UserGameData)
			info.UserId = userid
			info.Ctime = timestamp
			info.Gtype = int64(id)
			info.Number = int64(num)
			list = append(list, *info)
		}
	}
	for _, info := range list {
		addErr := StatisticsService.AddOrUpdateUserGame(&info)
		if addErr != nil {
			beego.Error("UserGameData fail err: ", addErr)
		}
	}
	// for _, item := range slist {
	// 	addErr := StatisticsService.AddOrUpdateUserGameStrategy(&item)
	// 	if addErr != nil {
	// 		beego.Error("UserGameData fail err: ", addErr)
	// 	}
	// }
}

// 玩家游戏局数统计
func UserGameDataCK(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)

	// 获取当天有登录的用户id
	var logId []string
	m := bson.M{}
	m["login_time"] = bson.M{"$gte": startTime, "$lte": endTime}
	LoginLogs.Find(m).Distinct("userid", &logId)
	list := make([]entity.UserGameData, 0)
	// slist := make([]entity.GameStrategy, 0)

	//		sql1 := `select userid,gtype,COUNT(1) number ,sum(CASE WHEN win_type = 2 and (gtype = 4 or gtype = 12) THEN settle_score ELSE bet_amount END) bets,SUM(end_time-begin_time) game_time,
	//		SUM(case when win_type = 1 THEN 1 else 0 end) win_number,
	//		SUM(
	//	    CASE
	//	        WHEN win_type = 1 THEN
	//	            CASE
	//		            WHEN gtype IN (1, 5, 6) THEN settle_score - bet_amount
	//		            WHEN gtype IN (7,10) THEN settle_score
	//	                WHEN gtype IN (2, 3, 4, 8,9,11,12) THEN settle_score - cash_ming_tax - cash_an_tax
	//	                ELSE 0
	//	            END
	//	        ELSE 0
	//	    END
	//
	// ) AS win_bets,
	//
	//	SUM(case when win_type = 2 THEN 1 else 0 end) lose_number,SUM(case when win_type = 2 THEN settle_score else 0 end) lose_bets
	//	from game.col_detail cd FINAL
	//	where robot = 0 and win_type in (1,2,3) and begin_time between ? and ? and userid in ?
	//	GROUP BY userid,gtype
	//	UNION ALL
	//	select user_id userid,game_id gtype,COUNT(1) number,SUM(amount) bets,0 game_time,0 win_number,0 win_bets,0 lose_number, 0 lose_bets
	//	from game.col_nsq_log_external_bet cnleb FINAL
	//	where amount != 0 and cancel = 0 and ctime between ? and ? and user_id in ?
	//	GROUP BY user_id,game_id`
	sql1 := `
select userid,gtype,COUNT(1) number ,sum(CASE WHEN win_type = 2 and (gtype = 4 or gtype = 12) THEN settle_score ELSE bet_amount END) bets,SUM(end_time-begin_time) game_time,
SUM(case when win_type = 1 THEN 1 else 0 end) win_number,
SUM(
CASE
	WHEN win_type = 1 THEN
		CASE
			WHEN gtype IN (1, 5, 6) THEN settle_score - bet_amount
			WHEN gtype IN (7,10) THEN settle_score
			WHEN gtype IN (2, 3, 4, 8,9,11,12) THEN settle_score - cash_ming_tax - cash_an_tax
			ELSE 0
		END
	ELSE 0
END
) AS win_bets,
SUM(case when win_type = 2 THEN 1 else 0 end) lose_number,SUM(case when win_type = 2 THEN settle_score else 0 end) lose_bets
from game.col_detail cd FINAL 
where robot = 0 and win_type in (1,2,3)  and userid in ?
GROUP BY userid,gtype
UNION ALL
select user_id userid,game_id gtype,COUNT(1) number,SUM(amount) bets,0 game_time,SUM(win_number) win_number,SUM(win_bets) win_bets,SUM(lose_number) lose_number,SUM(lose_bets) lose_bets
from (SELECT t1.user_id,t1.game_id,t1.amount,(CASE WHEN red.win_bets > 0 THEN 1 ELSE 0 END) win_number,red.win_bets,(CASE WHEN red.win_bets = 0 THEN 1 ELSE 0 END) lose_number,(CASE WHEN red.win_bets = 0 THEN t1.amount ELSE 0 END) lose_bets
from game.col_nsq_log_external_bet t1 FINAL 
left join (	SELECT  user_id,game_id,round_id,SUM(amount) win_bets from game.col_nsq_log_external_reward cnler FINAL where cancel = 0  and amount != 0
GROUP BY user_id,game_id,round_id) red on t1.user_id = red.user_id and t1.game_id = red.game_id and t1.round_id = red.round_id 
where  t1.amount != 0 and t1.cancel = 0) tab WHERE user_id in ?
GROUP BY user_id,game_id`

	var args1 []any
	// args1 = append(args1, startTime.UTC())
	// args1 = append(args1, endTime.UTC())
	args1 = append(args1, logId)
	// args1 = append(args1, startTime.UTC())
	// args1 = append(args1, endTime.UTC())
	args1 = append(args1, logId)
	err = ck.Select(&list, sql1, args1...)
	if err != nil {
		beego.Warning("UserGameDataCK error3:", err)
		return
	}

	for _, info := range list {
		if info.Gtype == 12 || info.Gtype == 4 {
			// RM
			num := info.Bets
			absNum := int64(math.Abs(float64(num)))
			info.Bets = absNum
		}
		if info.Gtype > 20 {
			// 外接数据
			info.LoseBets = -info.LoseBets
		}
		info.Id = fmt.Sprintf("%s-%d", info.UserId, info.Gtype)
		info.Ctime = timestamp
		addErr := StatisticsService.AddOrUpdateUserGame(&info)
		if addErr != nil {
			beego.Error("UserGameDataCK fail err: ", addErr)
		}
	}

	// for _, item := range slist {
	// 	addErr := StatisticsService.AddOrUpdateUserGameStrategy(&item)
	// 	if addErr != nil {
	// 		beego.Error("UserGameData fail err: ", addErr)
	// 	}
	// }
}

// 前端按钮全量埋点统计定时清除
func LogButton2ClickClean(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName).
		AddDate(0, 0, -3).Unix() // 清除3天前数据
	LogButton2Clicks.Remove(bson.M{"ctime": bson.M{"$lt": startTime}})
}

/*
	s := fmt.Sprintf("%s 00:00:00", startdate)
	sTime := utils.Str2Time(s)
	e := fmt.Sprintf("%s 23:59:59", enddate)
	eTime := utils.Str2Time(e)
*/

// PDDStats 玩游戏+分享抽大奖活动统计
func PDDStats(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
	date1, _ := utils.Unix(today)

	// 渠道总日活
	var loginedUsers int32
	var loginedUserids []string
	m1 := []bson.M{
		{"$match": bson.M{
			"login_time": bson.M{"$gte": startTime, "$lte": endTime},
		}},
		{"$group": bson.M{"_id": "$userid"}},
		// {"$group": bson.M{"_id": nil, "sum": bson.M{"$sum": 1}}},
	}
	var r1 []bson.M
	err := LoginLogs.Pipe(m1).All(&r1)
	if err != nil {
		beego.Error("loginedUsers error: ", err)
	} else if len(r1) > 0 {
		for _, r := range r1 {
			userid := r["_id"].(string)
			loginedUserids = append(loginedUserids, userid)
		}
		// loginedUsers = int32(r1[0]["sum"].(int))
		loginedUsers = int32(len(loginedUserids))
	}

	// 总参与人数 dtype 1：玩游戏 2：分享
	var pddUsers int32
	var pddUserids []string
	m2 := []bson.M{
		{"$match": bson.M{
			"ctime": bson.M{"$gte": startTime.Unix(), "$lte": endTime.Unix()},
			// "dtype": bson.M{"$in": []int{1, 2}},
		}},
		{"$group": bson.M{"_id": "$userid"}},
		// {"$group": bson.M{"_id": nil, "sum": bson.M{"$sum": 1}}},
	}
	var r2 []bson.M
	err = LogPlayShareDraws.Pipe(m2).All(&r2)
	fmt.Printf("r2: %d, %v, %d-%d\n", len(r2), r2, startTime.Unix(), endTime.Unix())

	if err != nil {
		beego.Error("loginedUsers error: ", err)
	} else if len(r2) > 0 {
		for _, r := range r2 {
			pddUserids = append(pddUserids, r["_id"].(string))
		}
		// pddUsers = int64(r1[0]["sum"].(int))
		pddUsers = int32(len(pddUserids))
	}

	// 查询参与活动的邀请注册的人数，有效分享数
	var beInvoteUsers int32
	var beInviteUserids []string
	if len(pddUserids) > 0 {
		m4 := []bson.M{
			{
				"$match": bson.M{
					// "ctime":          bson.M{"$gte": startTime, "$lt": endTime},
					"_id":              bson.M{"$in": pddUserids},
					"share_superior":   bson.M{"$ne": "", "$exists": true},
					"robot":            false,
					"simulation_robot": false,
				},
			},
			{
				"$group": bson.M{
					"_id": "$_id",
					// "num": bson.M{"$sum": 1},
				},
			},
		}
		var r4 []bson.M
		err = PlayerUsers.Pipe(m4).All(&r4)
		if err != nil {
			beego.Error("beInvoteUsers error: ", err)
		} else if len(r4) > 0 {
			for _, r := range r4 {
				userid := r["_id"].(string)
				beInviteUserids = append(beInviteUserids, userid)
			}
			beInvoteUsers = int32(len(beInviteUserids))
		}
	}

	// 被分享的用户登录人数
	var beInvoteLoginedUsers int32
	m5 := []bson.M{
		{
			"$match": bson.M{
				"_id":              bson.M{"$in": loginedUserids},
				"share_superior":   bson.M{"$ne": "", "$exists": true},
				"robot":            false,
				"simulation_robot": false,
			},
		},
		{"$group": bson.M{"_id": "$_id"}},
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": 1}}},
	}
	var r5 []bson.M
	err = PlayerUsers.Pipe(m5).All(&r5)
	if err != nil {
		beego.Error("beInvoteLoginedUsers error: ", err)
	} else if len(r5) > 0 {
		beInvoteLoginedUsers = int32(r5[0]["total"].(int))
	}

	// 所有渠道参与抽奖总人数
	var drawUsers int32
	m6 := []bson.M{
		{"$match": bson.M{
			"ctime": bson.M{"$gte": startTime.Unix(), "$lte": endTime.Unix()},
			"dtype": bson.M{"$in": []int{1, 2}},
		}},
		// {"$group": bson.M{"_id": "$userid"}},
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": 1}}},
	}
	var r6 []bson.M
	err = LogPlayShareDraws.Pipe(m6).All(&r6)
	if err != nil {
		beego.Error("drawUsers error: ", err)
	} else if len(r6) > 0 {
		drawUsers = int32(r6[0]["total"].(int))
	}

	// 分享用户参与抽奖总次数
	var beInvoteUserDraws int32
	m7 := []bson.M{
		{"$match": bson.M{
			"ctime":  bson.M{"$gte": startTime.Unix(), "$lte": endTime.Unix()},
			"userid": bson.M{"$in": beInviteUserids},
			"dtype":  bson.M{"$in": []int{1, 2}},
		}},
		// {"$group": bson.M{"_id": "$userid"}},
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": 1}}},
	}
	var r7 []bson.M
	err = LogPlayShareDraws.Pipe(m7).All(&r7)
	if err != nil {
		beego.Error("beInvoteUserDraws error: ", err)
	} else if len(r7) > 0 {
		beInvoteUserDraws = int32(r7[0]["total"].(int))
	}
	// 所有渠道参与抽奖总次数
	var userDraws int32
	m7_1 := []bson.M{
		{"$match": bson.M{
			"ctime": bson.M{"$gte": startTime.Unix(), "$lte": endTime.Unix()},
			// "userid": bson.M{"$in": beInviteUserids},
			"dtype": bson.M{"$in": []int{1, 2}},
		}},
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": 1}}},
	}
	var r7_1 []bson.M
	err = LogPlayShareDraws.Pipe(m7_1).All(&r7_1)
	if err != nil {
		beego.Error("userDraws error: ", err)
	} else if len(r7_1) > 0 {
		userDraws = int32(r7_1[0]["total"].(int))
	}

	// 领取到最终奖励的人数,累计发出去的金额
	var awardUsers int32
	var awardJackpot int64
	m8 := []bson.M{
		{"$match": bson.M{
			"ctime": bson.M{"$gte": startTime.Unix(), "$lte": endTime.Unix()},
			"dtype": bson.M{"$eq": 3},
		}},
		{"$group": bson.M{"_id": "$userid", "jackpot": bson.M{"$sum": "$jackpot"}}},
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": 1}, "jackpot": bson.M{"$sum": "$jackpot"}}},
	}
	var r8 []bson.M
	err = LogPlayShareDraws.Pipe(m8).All(&r8)
	if err != nil {
		beego.Error("awardUsers error: ", err)
	} else if len(r8) > 0 {
		awardUsers = int32(r8[0]["total"].(int))
		awardJackpot = int64(r8[0]["jackpot"].(int64))
	}

	// 所有渠道分享来注册的用户数
	var beInvoteRegUsers int32
	var beInvoteRegUserIds []string
	m9 := []bson.M{
		{
			"$match": bson.M{
				"ctime":            bson.M{"$gte": startTime, "$lt": endTime},
				"share_superior":   bson.M{"$ne": "", "$exists": true},
				"robot":            false,
				"simulation_robot": false,
			},
		},
		{"$group": bson.M{"_id": "$_id"}},
		// {"$group": bson.M{"_id": "$_id", "total": bson.M{"$sum": 1}}},
	}
	var r9 []bson.M
	err = PlayerUsers.Pipe(m9).All(&r9)
	if err != nil {
		beego.Error("beInvoteRegUsers error: ", err)
	} else if len(r9) > 0 {
		for _, r := range r9 {
			userid := r["_id"].(string)
			beInvoteRegUserIds = append(beInvoteRegUserIds, userid)
		}
		// beInvoteRegUsers = int32(r9[0]["total"].(int))
		beInvoteRegUsers = int32(len(beInvoteRegUserIds))
	}

	// 总分享设备数
	var beInvoteDevices int32
	m10 := []bson.M{
		{"$match": bson.M{
			"ctime":            bson.M{"$gte": startTime, "$lt": endTime},
			"share_superior":   bson.M{"$ne": "", "$exists": true},
			"robot":            false,
			"simulation_robot": false,
		}},
		{"$group": bson.M{"_id": "$ad__adid"}},
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": 1}}},
	}
	var r10 []bson.M
	err = PlayerUsers.Pipe(m10).All(&r10)
	if err != nil {
		beego.Error("beInvoteDevices error: ", err)
	} else if len(r10) > 0 {
		beInvoteDevices = int32(r10[0]["total"].(int))
	}

	// 分享用户分享来人数，仅分享用户分享后来注册的人数
	var shareSuperiorIds []string // 分享来的注册用户的所有上级
	m11_1 := []bson.M{
		{"$match": bson.M{
			"ctime":            bson.M{"$gte": startTime, "$lt": endTime},
			"share_superior":   bson.M{"$ne": "", "$exists": true},
			"robot":            false,
			"simulation_robot": false,
		}},
		{"$group": bson.M{"_id": "$share_superior"}},
		// {"$group": bson.M{"_id": nil, "total": bson.M{"$sum": 1}}},
	}
	var r11_1 []bson.M
	err = PlayerUsers.Pipe(m11_1).All(&r11_1)
	if err != nil {
		beego.Error("shareSuperiorIds error: ", err)
	} else if len(r11_1) > 0 {
		for _, r := range r11_1 {
			shareSuperiorIds = append(shareSuperiorIds, r["_id"].(string))
		}
		// beInvoteDevices = int32(r10[0]["total"].(int))
	}
	// 查询父级也是被分享来的用户
	var sharedSuperIds []string
	m11_2 := []bson.M{
		{"$match": bson.M{
			"_id":              bson.M{"$in": shareSuperiorIds},
			"share_superior":   bson.M{"$ne": "", "$exists": true},
			"robot":            false,
			"simulation_robot": false,
		}},
		{"$group": bson.M{"_id": "$_id"}},
	}
	var r11_2 []bson.M
	err = PlayerUsers.Pipe(m11_2).All(&r11_2)
	if err != nil {
		beego.Error("sharedSuperIds error: ", err)
	} else if len(r11_2) > 0 {
		for _, r := range r11_2 {
			sharedSuperIds = append(sharedSuperIds, r["_id"].(string))
		}
	}
	// 查询父级是被分享来的用户的今天注册的用户
	// 分享用户分享来用户数
	var share2Users int32
	var share2UserIds []string
	m11_13 := []bson.M{
		{"$match": bson.M{
			"ctime":            bson.M{"$gte": startTime, "$lt": endTime},
			"share_superior":   bson.M{"$in": sharedSuperIds},
			"robot":            false,
			"simulation_robot": false,
		}},
		{"$group": bson.M{"_id": "$_id"}},
		// {"$group": bson.M{"_id": nil, "total": bson.M{"$sum": 1}}},
	}
	var r11_3 []bson.M
	err = PlayerUsers.Pipe(m11_13).All(&r11_3)
	if err != nil {
		beego.Error("share2Users error: ", err)
	} else if len(r11_3) > 0 {
		for _, r := range r11_3 {
			userid := r["_id"].(string)
			share2UserIds = append(share2UserIds, userid)
		}
		// share2Users = int32(m11_13[0]["total"].(int))
		share2Users = int32(len(share2UserIds))
	}
	// 分享用户有效分享数
	var share2DrawUsers int32
	if len(share2UserIds) > 0 {
		m11_5 := []bson.M{
			{"$match": bson.M{
				"ctime": bson.M{"$gte": startTime.Unix(), "$lte": endTime.Unix()},
				// "dtype": bson.M{"$in": []int{1, 2}},
				"userid": bson.M{"$in": share2UserIds},
			}},
			{"$group": bson.M{"_id": "$userid"}},
			{"$group": bson.M{"_id": nil, "sum": bson.M{"$sum": 1}}},
		}
		var r11_5 []bson.M
		err = LogPlayShareDraws.Pipe(m11_5).All(&r11_5)
		if err != nil {
			beego.Error("m11_5 error: ", err)
		} else if len(r11_5) > 0 {
			share2DrawUsers = int32(r11_5[0]["sum"].(int))
		}
	}

	// 分享用户分享来设备数
	var share2Devices int32
	m11_14 := []bson.M{
		{"$match": bson.M{
			"ctime":            bson.M{"$gte": startTime, "$lt": endTime},
			"share_superior":   bson.M{"$in": sharedSuperIds},
			"robot":            false,
			"simulation_robot": false,
		}},
		{"$group": bson.M{"_id": "$ad__adid"}},
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": 1}}},
	}
	var r11_4 []bson.M
	err = PlayerUsers.Pipe(m11_14).All(&r11_4)
	if err != nil {
		beego.Error("share2Devices error: ", err)
	} else if len(r11_4) > 0 {
		share2Devices = int32(r11_4[0]["total"].(int))
	}

	// 24小时玩游戏人数，从该用户注册起，24小时内玩过1局以上的用户数
	var playedReg24 int32
	var regUserids []string
	var playedReg24UserCtime = make(map[string]time.Time)
	m12 := []bson.M{
		{"$match": bson.M{
			"ctime":            bson.M{"$gte": startTime, "$lt": endTime},
			"robot":            false,
			"simulation_robot": false,
		}},
		{"$group": bson.M{"_id": bson.M{"userid": "$_id", "ctime": "$ctime"}}},
		// {"$group": bson.M{"_id": nil, "total": bson.M{"$sum": 1}}},
	}
	// 今日注册用户及注册时间
	var r12 []bson.M
	err = PlayerUsers.Pipe(m12).All(&r12)
	if err != nil {
		beego.Error("playedReg24 error: ", err)
	} else if len(r12) > 0 {
		for _, r := range r12 {
			userid := r["_id"].(bson.M)["userid"].(string)
			ctime := r["_id"].(bson.M)["ctime"].(time.Time)
			regUserids = append(regUserids, userid)
			playedReg24UserCtime[userid] = ctime
		}
	}

	// 今日注册用户第一局对局时间
	var userBeginTimes = make(map[string]int64)
	m12_2 := []bson.M{
		{"$match": bson.M{
			"begin_time": bson.M{"$gte": startTime.Unix(), "$lt": endTime.AddDate(0, 0, 1).Unix()}, // 查48小时内对局信息
			"players":    bson.M{"$ne": ""},
		}},
		{"$project": bson.M{"begin_time": "$begin_time", "userids": bson.M{"$split": []string{"$players", ","}}}},
		{"$match": bson.M{
			"userids": bson.M{"$in": regUserids},
		}},
	}
	var r12_2 []bson.M
	err = Details.Pipe(m12_2).All(&r12_2)
	if err != nil {
		beego.Error("r12_2 error: ", err)
	} else if len(r12_2) > 0 {
		for _, r := range r12_2 {
			beginTime := r["begin_time"].(int64)
			for _, uid := range r["userids"].([]interface{}) {
				userid := uid.(string)
				if userBeginTimes[userid] == 0 || userBeginTimes[userid] > beginTime {
					userBeginTimes[userid] = beginTime
				}
			}
		}
	}
	// 外接游戏对局
	m12_3 := []bson.M{
		{"$match": bson.M{"amount": bson.M{"$ne": 0}, "user_id": bson.M{"$in": regUserids}}},
		{"$group": bson.M{"_id": bson.M{"round_id": "$round_id", "user_id": "$user_id"}, "ctime": bson.M{"$min": "$ctime"}}},
	}
	var r12_3 []bson.M
	err = NsqLogExternalBets.Pipe(m12_3).All(&r12_3)
	if err != nil {
		glog.Error("m12_3 error: ", err)
	} else if len(r12_3) > 0 {
		for _, r := range r12_3 {
			userid := r["_id"].(bson.M)["user_id"].(string)
			ctime := r["ctime"].(int64) // 对局开始时间
			if userBeginTimes[userid] == 0 || userBeginTimes[userid] > ctime {
				userBeginTimes[userid] = ctime
			}
		}
	}

	hour24 := int64((24 * time.Hour).Seconds())
	for userid, ctime := range playedReg24UserCtime {
		if beginTime, ok := userBeginTimes[userid]; ok {
			if beginTime-ctime.Unix() < hour24 { // 24小时内对局用户
				playedReg24++
			}
		}
	}

	// 首冲CPP, 获奖金额/新顾客充值人数
	var regUserChages int32
	var regUserChargeAmount int64
	m13 := []bson.M{
		{"$match": bson.M{
			"ctime":        bson.M{"$gte": startTime, "$lt": endTime},
			"userid":       bson.M{"$in": beInvoteRegUserIds},
			"order_status": 4,
		}},
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": 1}, "amount": bson.M{"$sum": "$amount"}}},
	}
	var r13 []bson.M
	err = Pays.Pipe(m13).All(&r13)
	if err != nil {
		beego.Error("r13 error: ", err)
	} else if len(r13) > 0 {
		regUserChages = int32(r13[0]["total"].(int))
		regUserChargeAmount = int64(r13[0]["amount"].(int))
	}

	// 新用户总提金额
	var regUserWithdrawAmount int64
	m14 := []bson.M{
		{
			"$match": bson.M{
				"ctime":        bson.M{"$gte": startTime, "$lt": endTime},
				"userid":       bson.M{"$in": beInvoteRegUserIds},
				"order_status": 2,
			},
		},
		{"$group": bson.M{"_id": nil, "amount": bson.M{"$sum": "$amount"}}},
	}

	var r14 []bson.M
	err = Withdraws.Pipe(m14).All(&r14)
	if err != nil {
		beego.Error("r14 error: ", err)
	} else if len(r14) > 0 {
		regUserWithdrawAmount = int64(r14[0]["amount"].(int))
	}

	// 总充值金额/获奖金额
	var chargeAmount int64
	m15 := []bson.M{
		{"$match": bson.M{
			"ctime": bson.M{"$gte": startTime, "$lt": endTime},
			// "userid":       bson.M{"$in": beInviteUserids},
			"order_status": 4,
		}},
		// {"$group": bson.M{"_id": bson.M{"userid": "$userid", "amount": "$amount"}}},
		{"$group": bson.M{"_id": nil, "amount": bson.M{"$sum": "$amount"}}},
	}
	var r15 []bson.M
	err = Pays.Pipe(m15).All(&r15)
	if err != nil {
		beego.Error("r15 error: ", err)
	} else if len(r15) > 0 {
		chargeAmount = int64(r15[0]["amount"].(int))
	}
	// 总提金额
	var withdrawAmount int64
	m16 := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": startTime, "$lt": endTime},
				// "userid":       bson.M{"$in": beInvoteRegUserIds},
				"order_status": 2,
			},
		},
		{"$group": bson.M{"_id": nil, "amount": bson.M{"$sum": "$amount"}}},
	}
	var r16 []bson.M
	err = Withdraws.Pipe(m16).All(&r16)
	if err != nil {
		beego.Error("r16 error: ", err)
	} else if len(r16) > 0 {
		withdrawAmount = int64(r16[0]["amount"].(int))
	}

	// 活动利润，分享用户总充-总提-获奖金额
	// BeInvoteChargeAmount BeInvoteWithdrawAmount
	var beInvoteChargeAmount int64
	var chargeUserids []string
	var chargeAmounts = make(map[string]int64)
	m17_1 := []bson.M{
		{"$match": bson.M{
			"ctime":        bson.M{"$gte": startTime, "$lt": endTime},
			"order_status": 4,
		}},
		{"$group": bson.M{"_id": "$userid", "amount": bson.M{"$sum": "$amount"}}},
	}
	var r17_1 []bson.M
	err = Pays.Pipe(m17_1).All(&r17_1)
	if err != nil {
		beego.Error("r17 error: ", err)
	} else if len(r17_1) > 0 {
		for _, r := range r17_1 {
			userid := r["_id"].(string)
			amount := r["amount"].(int)
			chargeUserids = append(chargeUserids, userid)
			chargeAmounts[userid] = int64(amount)
		}
	}
	// 查询是分享来的充值用户
	var sharedChargeUserIds []string
	m17_2 := []bson.M{
		{"$match": bson.M{
			"_id":              bson.M{"$in": chargeUserids},
			"share_superior":   bson.M{"$ne": "", "$exists": true},
			"robot":            false,
			"simulation_robot": false,
		}},
		{"$group": bson.M{"_id": "$_id"}},
	}
	var r17_2 []bson.M
	err = PlayerUsers.Pipe(m17_2).All(&r17_2)
	if err != nil {
		beego.Error("sharedChargeUserIds error: ", err)
	} else if len(r17_2) > 0 {
		for _, r := range r17_2 {
			sharedChargeUserIds = append(sharedChargeUserIds, r["_id"].(string))
		}
	}
	for _, userid := range sharedChargeUserIds {
		beInvoteChargeAmount += chargeAmounts[userid]
	}

	// 分享用户提现金额
	var beInvoteWithdrawAmount int64
	var withdrawUserids []string
	var withdrawAmounts = make(map[string]int64)
	m18_1 := []bson.M{
		{
			"$match": bson.M{
				"ctime":        bson.M{"$gte": startTime, "$lt": endTime},
				"order_status": 2,
			},
		},
		{"$group": bson.M{"_id": "$userid", "amount": bson.M{"$sum": "$amount"}}},
	}
	var r18_1 []bson.M
	err = Withdraws.Pipe(m18_1).All(&r18_1)
	if err != nil {
		beego.Error("r18_1 error: ", err)
	} else if len(r18_1) > 0 {
		for _, r := range r18_1 {
			userid := r["_id"].(string)
			amount := r["amount"].(int)
			withdrawUserids = append(withdrawUserids, userid)
			withdrawAmounts[userid] = int64(amount)
		}
	}
	// 查询是分享来的提现用户
	var sharedWithdrawUserIds []string
	m18_2 := []bson.M{
		{"$match": bson.M{
			"_id":              bson.M{"$in": withdrawUserids},
			"share_superior":   bson.M{"$ne": "", "$exists": true},
			"robot":            false,
			"simulation_robot": false,
		}},
		{"$group": bson.M{"_id": "$_id"}},
	}
	var r18_2 []bson.M
	err = PlayerUsers.Pipe(m18_2).All(&r18_2)
	if err != nil {
		beego.Error("sharedWithdrawUserIds error: ", err)
	} else if len(r18_2) > 0 {
		for _, r := range r18_2 {
			sharedWithdrawUserIds = append(sharedWithdrawUserIds, r["_id"].(string))
		}
	}
	for _, userid := range sharedWithdrawUserIds {
		beInvoteWithdrawAmount += withdrawAmounts[userid]
	}

	info := &entity.PddStat{
		Date:                   date1,
		LoginedUsers:           loginedUsers,
		PddUsers:               pddUsers,
		BeInvoteUsers:          beInvoteUsers,
		BeInvoteLoginedUsers:   beInvoteLoginedUsers,
		DrawUsers:              drawUsers,
		BeInvoteUserDraws:      beInvoteUserDraws,
		UserDraws:              userDraws,
		AwardUsers:             awardUsers,
		AwardJackpot:           awardJackpot,
		BeInvoteRegUsers:       beInvoteRegUsers,
		BeInvoteDevices:        beInvoteDevices,
		Share2Users:            share2Users,
		Share2DrawUsers:        share2DrawUsers,
		Share2Devices:          share2Devices,
		PlayedReg24:            playedReg24,
		RegUserChages:          regUserChages,
		RegUserChargeAmount:    regUserChargeAmount,
		RegUserWithdrawAmount:  regUserWithdrawAmount,
		ChargeAmount:           chargeAmount,
		WithdrawAmount:         withdrawAmount,
		BeInvoteChargeAmount:   beInvoteChargeAmount,
		BeInvoteWithdrawAmount: beInvoteWithdrawAmount,
	}
	addErr := StatisticsService.AddOrUpdatePddStat(info)
	if addErr != nil {
		beego.Error("PddStat fail err: ", addErr)
	}
}

// 每日0点获取一次支付渠道余额
// func PayChannelStats(timestamp int64) {
// 	tim := utils.Stamp2Time(timestamp)
// 	today := tim.Format("2006-01-02")
// 	s := fmt.Sprintf("%s 00:00:00", today)
// 	s1 := utils.Str2Time(s)
// 	startTime := utils.Time2StampToMS(s1)
// 	e := fmt.Sprintf("%s 23:59:59", today)
// 	s2 := utils.Str2Time(e)
// 	endTime := utils.Time2StampToMS(s2)
// 	list := make([]entity.PayChannelStatsData, 0)
// 	PayChannels, _ := GameService.GetPayChannelList(-1, -1, bson.M{})
// 	if len(PayChannels) > 0 {
// 		date1, _ := utils.Unix(fmt.Sprintf("%s", today))
// 		for _, item := range PayChannels {
// 			// fmt.Printf(item.Id)

// 			channelId, _ := strconv.Atoi(item.Id)
// 			// channelId := 3018
// 			pipeline := []bson.M{
// 				{
// 					"$match": bson.M{
// 						"pay_channel": channelId,
// 						"date":        bson.M{"$gte": startTime, "$lte": endTime},
// 					},
// 				},
// 				{
// 					"$group": bson.M{
// 						"_id":             "$p_type",
// 						"amount":          bson.M{"$sum": "$amount"},
// 						"actual_amount":   bson.M{"$sum": "$actual_amount"},
// 						"handling_charge": bson.M{"$sum": "$handling_charge"},
// 						"rate":            bson.M{"$max": "$rate"},
// 						"balance":         bson.M{"$last": "$balance"},
// 					},
// 				},
// 				{
// 					"$sort": bson.M{"date": 1},
// 				},
// 			}
// 			var res []bson.M
// 			PayChannelLogs.Pipe(pipeline).All(&res)
// 			if len(res) > 0 {
// 				info := new(entity.PayChannelStatsData)
// 				for _, v := range res {
// 					ptype := v["_id"].(int)
// 					amount := v["amount"].(int64)
// 					actual_amount := v["actual_amount"].(int64)
// 					handling_charge := v["handling_charge"].(int64)
// 					rate := v["rate"].(float64)
// 					// balance := v["balance"].(int64)
// 					switch ptype {
// 					case 1:
// 						// 支付
// 						info.TotalPay = amount
// 						info.PayRate = rate
// 						info.TotalPayAmount = actual_amount
// 						info.PayHandlingFee = handling_charge
// 					case 2:
// 						// 提现
// 						info.WithdrawRate = rate
// 						info.TotalWithdraw = amount
// 						info.TotalWithdrawAmount = actual_amount
// 						info.WithdrawHandlingFee = handling_charge
// 					}
// 				}
// 				info.Date = date1
// 				info.PayChannel = uint32(channelId)
// 				list = append(list, *info)
// 			}
// 		}
// 	}
// 	for _, user := range list {
// 		AddErr := PayService.AddOrUpdatePayChannelStat(&user)
// 		if AddErr != nil {
// 			beego.Error("PayChannelStats fail err: ", AddErr)
// 		}
// 	}
// }

/*数据导出*/
// 1到5局的用户，玩的哪些游戏和胜率 A类，B类分开
func UserGameInningsByA(begin, end string) {
	beego.Info("开始执行 UserGameInningsByA")
	s := fmt.Sprintf("%s 00:00:00", begin)
	sTime := utils.Str2TimeZone(s, locationName)
	e := fmt.Sprintf("%s 23:59:59", end)
	eTime := utils.Str2TimeZone(e, locationName)
	// A类注册用户id
	a_uid := make([]string, 0)
	// B类注册用户id
	// b_uid := make([]string, 0)
	n := bson.M{}
	n["robot"] = false
	n["simulation_robot"] = false
	n["ctime"] = bson.M{"$gte": sTime, "$lte": eTime}
	n["regist_area"] = bson.M{"$ne": 1}
	PlayerUsers.Find(n).Distinct("_id", &a_uid)
	// 传入之前将日期修改成1-30号
	s1 := fmt.Sprintf("%s 00:00:00", "2024-04-01")
	startTime := utils.Str2TimeZone(s1, locationName)
	e1 := fmt.Sprintf("%s 23:59:59", "2024-04-30")
	endTime := utils.Str2TimeZone(e1, locationName)
	statrTimestamp := utils.Time2Stamp(startTime)
	endTimestamp := utils.Time2Stamp(endTime)
	list := UserGameInningsCompute(statrTimestamp, endTimestamp, a_uid)

	// B类
	// n["regist_area"] = 1
	// PlayerUsers.Find(n).Distinct("_id", &b_uid)
	// list1 := UserGameInningsCompute(sTime, eTime, b_uid)
	beego.Info("计算完成：开始导出数据！")
	// 导出
	f := excelize.NewFile()
	sheet1 := "Sheet1"
	// 设置表头
	f.SetCellValue(sheet1, "A1", "玩家id")
	f.SetCellValue(sheet1, "B1", "局号id")
	f.SetCellValue(sheet1, "C1", "游戏名称")
	f.SetCellValue(sheet1, "D1", "输赢")
	line := 1
	for _, player := range list {
		line++
		f.SetCellValue(sheet1, fmt.Sprintf("A%d", line), player.Userid)
		f.SetCellValue(sheet1, fmt.Sprintf("B%d", line), player.Waterid)
		f.SetCellValue(sheet1, fmt.Sprintf("C%d", line), player.GtypeName)
		f.SetCellValue(sheet1, fmt.Sprintf("D%d", line), player.Result)
	}
	// sheet2 := "sheet2"
	// // 设置表头
	// f.SetCellValue(sheet2, "A1", "玩家id")
	// f.SetCellValue(sheet2, "B1", "局号id")
	// f.SetCellValue(sheet2, "C1", "游戏名称")
	// f.SetCellValue(sheet2, "D1", "输赢")
	// line = 1
	// for _, player := range list1 {
	// 	line++
	// 	f.SetCellValue(sheet2, fmt.Sprintf("A%d", line), player.Userid)
	// 	f.SetCellValue(sheet2, fmt.Sprintf("B%d", line), player.Waterid)
	// 	f.SetCellValue(sheet2, fmt.Sprintf("C%d", line), player.GtypeName)
	// 	f.SetCellValue(sheet2, fmt.Sprintf("D%d", line), player.Result)
	// }

	now := time.Now().Format("2006-01-02.15.04.05")
	filename := "A类用户局数记录" + "_" + fmt.Sprint(now) + ".xlsx"
	// 保存文件
	if err := f.SaveAs(filename); err != nil {
		glog.Error("error3: ", err)
	}
}

func UserGameInningsByB(begin, end string) {
	beego.Info("开始执行 UserGameInningsByA")
	s := fmt.Sprintf("%s 00:00:00", begin)
	sTime := utils.Str2TimeZone(s, locationName)
	e := fmt.Sprintf("%s 23:59:59", end)
	eTime := utils.Str2TimeZone(e, locationName)
	// B类注册用户id
	b_uid := make([]string, 0)
	n := bson.M{}
	n["robot"] = false
	n["simulation_robot"] = false
	n["ctime"] = bson.M{"$gte": sTime, "$lte": eTime}
	n["regist_area"] = 1
	PlayerUsers.Find(n).Distinct("_id", &b_uid)
	// 传入之前将日期修改成1-30号
	s1 := fmt.Sprintf("%s 00:00:00", "2024-04-01")
	startTime := utils.Str2TimeZone(s1, locationName)
	e1 := fmt.Sprintf("%s 23:59:59", "2024-04-30")
	endTime := utils.Str2TimeZone(e1, locationName)
	statrTimestamp := utils.Time2Stamp(startTime)
	endTimestamp := utils.Time2Stamp(endTime)
	list := UserGameInningsCompute(statrTimestamp, endTimestamp, b_uid)

	beego.Info("计算完成：开始导出数据！")
	// 导出
	f := excelize.NewFile()
	sheet1 := "Sheet1"
	// 设置表头
	f.SetCellValue(sheet1, "A1", "玩家id")
	f.SetCellValue(sheet1, "B1", "局号id")
	f.SetCellValue(sheet1, "C1", "游戏名称")
	f.SetCellValue(sheet1, "D1", "输赢")
	line := 1
	for _, player := range list {
		line++
		f.SetCellValue(sheet1, fmt.Sprintf("A%d", line), player.Userid)
		f.SetCellValue(sheet1, fmt.Sprintf("B%d", line), player.Waterid)
		f.SetCellValue(sheet1, fmt.Sprintf("C%d", line), player.GtypeName)
		f.SetCellValue(sheet1, fmt.Sprintf("D%d", line), player.Result)
	}

	now := time.Now().Format("2006-01-02.15.04.05")
	filename := "B类用户局数记录" + "_" + fmt.Sprint(now) + ".xlsx"
	// 保存文件
	if err := f.SaveAs(filename); err != nil {
		glog.Error("error3: ", err)
	}
}

func UserGameInningsCompute(statrTimestamp, endTimestamp int64, uids []string) []entity.UserGameInning {
	execl_result := make([]entity.UserGameInning, 0)
	// 游戏局数
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"userid": bson.M{"$in": uids},
				"ctime":  bson.M{"$gte": statrTimestamp, "$lte": endTimestamp},
			},
		},
		{
			"$group": bson.M{
				"_id":   "$userid",
				"count": bson.M{"$sum": 1}, // 添加一个额外的字段以便后续过滤
			},
		},
		{
			"$match": bson.M{
				"count": bson.M{"$gte": 1, "$lte": 5}, // 过滤num值在1-5之间的文档
			},
		},
	}
	result := []bson.M{}
	LogGameTimes.Pipe(pipeline).All(&result)
	// 外接游戏局数
	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"user_id": bson.M{"$in": uids},
				"amount":  bson.M{"$ne": 0},
				"ctime":   bson.M{"$gte": statrTimestamp, "$lte": endTimestamp},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"user_id":  "$user_id",
					"round_id": "$round_id",
				},
				"count": bson.M{"$sum": 1}, // 添加一个额外的字段以便后续过滤
			},
		},
		{
			"$project": bson.M{
				"id":       "$_id.user_id",
				"round_id": "$_id.round_id",
				"count":    "$count",
			},
		},
		{
			"$match": bson.M{
				"count": bson.M{"$gte": 1, "$lte": 5}, // 过滤num值在1-5之间的文档
			},
		},
	}
	result1 := []bson.M{}
	NsqLogExternalBets.Pipe(pipeline1).All(&result1)
	// 汇总两个游戏用户的局数
	t_map := make(map[string]int, 0)
	for _, item := range result {
		uid := item["_id"].(string)
		ucount := item["count"].(int)
		t_map[uid] = ucount
	}

	for _, item := range result1 {
		uid := item["id"].(string)
		ucount := item["count"].(int)
		if val, ok := t_map[uid]; ok {
			t_map[uid] = val + ucount
		} else {
			t_map[uid] = ucount
		}
	}
	num_uid := make([]string, 0)
	// 重新计算汇总后符合1-5局的数据
	for id, p := range t_map {
		if p <= 5 {
			num_uid = append(num_uid, id)
		}
	}

	// 查询符合条件的用户id的游戏对局
	for _, uid := range num_uid {
		// 对局
		pipeline2 := []bson.M{
			{
				"$match": bson.M{
					"end_time": bson.M{"$gte": statrTimestamp, "$lte": endTimestamp},
					"players": bson.M{
						"$regex":   fmt.Sprintf("\\b%s\\b", uid),
						"$options": "i",
					},
				},
			},
		}
		var result2 []entity.Detail
		Details.Pipe(pipeline2).All(&result2)
		for _, dt := range result2 {
			newinfo := new(entity.UserGameInning)
			newinfo.Userid = uid
			newinfo.Waterid = dt.WaterId
			newinfo.Gtype = dt.Gtype
			newinfo.GtypeName = GetGameNameById(true, dt.Gtype)
			switch dt.Gtype {
			case 1:
				if dt.TPDetail != nil {
					for _, e := range dt.TPDetail {
						if e.UserId == uid {
							if e.Score > 0 {
								newinfo.Result = 1
							}
						}
					}
				}
			case 2:
				if dt.LHDetail != nil {
					for _, e := range dt.LHDetail.UserDetail {
						if e.Userid == uid {
							if e.Win > 0 {
								newinfo.Result = 1
							}
						}
					}
				}
			case 3:
				if dt.UPDetail != nil {
					for _, e := range dt.UPDetail.UserDetail {
						if e.Userid == uid {
							if e.Win > 0 {
								newinfo.Result = 1
							}
						}
					}
				}
			case 4:
				if dt.RMDetail != nil {
					for _, e := range dt.RMDetail {
						if e.UserId == uid {
							if e.Score > 0 {
								newinfo.Result = 1
							}
						}
					}
				}
			case 5:
				if dt.AK47Detail != nil {
					for _, e := range dt.AK47Detail {
						if e.UserId == uid {
							if e.Score > 0 {
								newinfo.Result = 1
							}
						}
					}
				}
			case 6:
				if dt.JOKERDetail != nil {
					for _, e := range dt.JOKERDetail {
						if e.UserId == uid {
							if e.Score > 0 {
								newinfo.Result = 1
							}
						}
					}
				}
			case 10:
			case 7:
				if dt.CRASHDetail != nil {
					for _, e := range dt.CRASHDetail.UserDetail {
						if e.Userid == uid {
							if e.Win > 0 {
								newinfo.Result = 1
							}
						}
					}
				}
			case 8:
				if dt.ABDetail != nil {
					for _, e := range dt.ABDetail.UserDetail {
						if e.Userid == uid {
							if e.Win > 0 {
								newinfo.Result = 1
							}
						}
					}
				}
			case 9:
				if dt.CPDetail != nil {
					for _, e := range dt.CPDetail.UserDetail {
						if e.Userid == uid {
							if e.Win > 0 {
								newinfo.Result = 1
							}
						}
					}
				}
			}
			execl_result = append(execl_result, *newinfo)
		}

		// 外接游戏局数
		pipeline3 := []bson.M{
			{
				"$match": bson.M{
					"user_id": uid,
					"amount":  bson.M{"$ne": 0},
					"ctime":   bson.M{"$gte": statrTimestamp, "$lte": endTimestamp},
				},
			},
		}
		var result3 []entity.ExternalDetail
		NsqLogExternalBets.Pipe(pipeline3).All(&result3)
		for _, nb := range result3 {
			newinfo := new(entity.UserGameInning)
			newinfo.Userid = uid
			newinfo.Waterid = nb.RoundId
			newinfo.Gtype = nb.GameId
			newinfo.GtypeName = GetGameNameById(false, nb.GameId)
			pipeline4 := []bson.M{
				{
					"$match": bson.M{
						"user_id":  uid,
						"round_id": nb.RoundId,
						"game_id":  nb.GameId,
						"ctime":    bson.M{"$gte": statrTimestamp, "$lte": endTimestamp},
					},
				},
				{
					"$group": bson.M{
						"_id":    "$round_id",
						"amount": bson.M{"$sum": "$amount"},
					},
				},
			}
			rew := []bson.M{}
			NsqLogExternalRewards.Find(pipeline4).All(&rew)
			for _, nr := range rew {
				amount := nr["amount"].(int64)
				if amount > 0 {
					newinfo.Result = 1
				}
			}
			execl_result = append(execl_result, *newinfo)
		}

	}
	return execl_result
}

func GetGameNameById(isflag bool, gtype int32) string {
	name := ""
	var recordList map[int]string
	if isflag {
		// 自
		recordList = map[int]string{
			1:  "TP",
			2:  "DRAGON TIGER",
			3:  "7UPDOWN",
			4:  "RUMMY",
			5:  "AK47",
			6:  "JOKER",
			7:  "CRASH",
			8:  "ANDARBAHAR",
			9:  "彩票",
			10: "飞机",
			11: "红黑大战",
			12: "RUMMY双人",
			13: "TP2",
		}
	} else {
		recordList = map[int]string{
			600101: "Ganesha Fortune",
			600002: "Lucky Neko",
			600073: "Ganesha Gold",
			600022: "Fortune Ox",
			600039: "Speed Winner",
			600120: "Treasures of Aztec",
			600054: "Wild Bounty Showdown",
			600104: "Rise of Apollo",
			600037: "Fortune Tiger",
			600028: "Destiny of Sun & Moon",
			600041: "Legend of Perseus",
			600098: "CaiShen Wins",
			600025: "Songkran Splash",
			600093: "Asgardian Rising",
			600004: "Jurassic Kingdom",
			600108: "Wild Bandito",
			600009: "Supermarket Spree",
			600012: "Cocktail Nights",
			600119: "Galactic Gems",
			600110: "Ways of the Qilin",
			600029: "Honey Trap of Diao Chan",
			600081: "Dragon Hatch",
			600086: "Leprechaun Riches",
			600099: "Egypt's Book of Mystery",
			600102: "Dreams of Macau",
			600117: "Queen of Bounty",
			600103: "Candy Bonanza",
			600607: "Auto-Roulette",
			600572: "Lightning Blackjack",
			600526: "Lightning Roulette",
			600513: "Super Sic Bo",
			600594: "Dragon Tiger",
			600583: "Dream Catcher",
			600642: "Fan Tan",
			600536: "Golden Wealth Baccarat",
			600631: "Bac Bo",

			600276: "Fortune Gems",
			600369: "Fortune Gems 2",
			600402: "Fortune Gems 3",
			600332: "Money Coming",
			602287: "Crazy Time",
			600312: "Jackpot Fishing",
			600247: "Crazy777",
			602687: "Money Coming Expand Bets",
			600326: "Super Ace",
			604137: "Lightning Roulette",
			602344: "Funky Time",
			600339: "Ocean King Jackpot",
			604278: "Fortune Roulette",
			// 600101: "Ganesha Fortune",
			// 600009: "Supermarket Spree",
			600325: "Charge Buffalo",
			// 600120: "Treasures of Aztec",
			600374: "Fortune Dragon",
			600019: "Fortune Rabbit",
			604266: "Lucky Jaguar",
			600076: "Double Fortune",
			600761: "3 Coin Treasures",
			// 600583: "Dream Catcher",
			600561: "Super Andar Bahar",
			600766: "Crazy Hunter",
			600362: "Boxing King",
		}
	}

	if gname, ok := recordList[int(gtype)]; ok {
		name = gname
	}
	return name
}

func UserGameInnings_New(begin, end string) {
	beego.Info("开始执行 UserGameInnings_New")
	s := fmt.Sprintf("%s 00:00:00", begin)
	sTime := utils.Str2TimeZone(s, locationName)
	e := fmt.Sprintf("%s 23:59:59", end)
	eTime := utils.Str2TimeZone(e, locationName)

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"robot":            false,
				"simulation_robot": false,
				"ctime":            bson.M{"$gte": sTime, "$lte": eTime},
			},
		},
		{
			"$project": bson.M{
				"_id":         "$_id",
				"ctime":       "$ctime",
				"regist_area": "$regist_area",
				"diamond":     "$diamond",
			},
		},
	}
	var res []entity.UserGameInningNew
	list := make([]entity.UserGameInningNew, 0)
	PlayerUsers.Pipe(pipeline).All(&res)
	// 传入之前将日期修改成12-19号
	s1 := fmt.Sprintf("%s 00:00:00", "2024-05-12")
	startTime := utils.Str2TimeZone(s1, locationName)
	e1 := fmt.Sprintf("%s 23:59:59", "2024-05-19")
	endTime := utils.Str2TimeZone(e1, locationName)
	statrTimestamp := utils.Time2Stamp(startTime)
	endTimestamp := utils.Time2Stamp(endTime)
	for _, item := range res {
		id := item.Userid
		ucount := UserGameInningsCompute_New(statrTimestamp, endTimestamp, id)
		if ucount <= 5 {
			info := new(entity.UserGameInningNew)
			info.Userid = item.Userid
			if item.RegistArea == 1 {
				info.RegistAreaName = "B类"
			} else {
				info.RegistAreaName = "A类"
			}
			info.RegistArea = item.RegistArea
			info.Diamond = item.Diamond
			c, _ := ConvertToIndiaTime(item.Ctime.Unix())
			info.Ctime = c
			info.GameNumber = ucount
			list = append(list, *info)
		}
	}
	beego.Info("计算完成：开始导出数据！")
	// 导出
	f := excelize.NewFile()
	sheet1 := "Sheet1"
	// 设置表头
	f.SetCellValue(sheet1, "A1", "玩家id")
	f.SetCellValue(sheet1, "B1", "用户类型")
	f.SetCellValue(sheet1, "C1", "彩金")
	f.SetCellValue(sheet1, "D1", "注册时间")
	f.SetCellValue(sheet1, "E1", "游戏局数")
	line := 1
	for _, player := range list {
		line++
		f.SetCellValue(sheet1, fmt.Sprintf("A%d", line), player.Userid)
		f.SetCellValue(sheet1, fmt.Sprintf("B%d", line), player.RegistAreaName)
		f.SetCellValue(sheet1, fmt.Sprintf("C%d", line), player.Diamond)
		f.SetCellValue(sheet1, fmt.Sprintf("D%d", line), player.Ctime)
		f.SetCellValue(sheet1, fmt.Sprintf("E%d", line), player.GameNumber)
	}
	now := time.Now().Format("2006-01-02.15.04.05")
	filename := begin + "用户局数记录" + "_" + fmt.Sprint(now) + ".xlsx"
	// 保存文件
	if err := f.SaveAs(filename); err != nil {
		glog.Error("error3: ", err)
	}
}

func UserGameInningsCompute_New(statrTimestamp, endTimestamp int64, uid string) int {
	// 游戏局数
	// pipeline := []bson.M{
	// 	{
	// 		"$match": bson.M{
	// 			"end_time": bson.M{"$gte": statrTimestamp, "$lte": endTimestamp},
	// 			"players": ,
	// 		},
	// 	},
	// 	{
	// 		"$group": bson.M{
	// 			"_id":   nil,
	// 			"count": bson.M{"$sum": 1}, // 添加一个额外的字段以便后续过滤
	// 		},
	// 	},
	// }
	m := bson.M{}
	m["end_time"] = bson.M{"$gte": statrTimestamp, "$lte": endTimestamp}
	m["players"] = bson.M{
		"$regex":   fmt.Sprintf("\\b%s\\b", uid),
		"$options": "i",
	}
	zcount := Count(Details, m)

	// var result []bson.M
	// err := Details.Pipe(pipeline).All(&result)
	// fmt.Print(err)
	// 外接游戏局数
	// pipeline1 := []bson.M{
	// 	{
	// 		"$match": bson.M{
	// 			"user_id": uid,
	// 			"amount":  bson.M{"$ne": 0},
	// 			"ctime":   bson.M{"$gte": statrTimestamp, "$lte": endTimestamp},
	// 		},
	// 	},
	// 	{
	// 		"$group": bson.M{
	// 			"_id":   nil,
	// 			"count": bson.M{"$sum": 1}, // 添加一个额外的字段以便后续过滤
	// 		},
	// 	},
	// }
	// var result1 []bson.M
	// NsqLogExternalBets.Pipe(pipeline1).All(&result1)
	m1 := bson.M{}
	m1["ctime"] = bson.M{"$gte": statrTimestamp, "$lte": endTimestamp}
	m1["user_id"] = uid
	m1["amount"] = bson.M{"$ne": 0}
	ncount := Count(NsqLogExternalBets, m1)
	// 汇总两个游戏用户的局数
	gameCount := zcount + ncount
	// if len(result) > 0 {
	// 	ucount := result[0]["count"].(int)
	// 	gameCount += ucount
	// }
	// if len(result1) > 0 {
	// 	ucount := result1[0]["count"].(int)
	// 	gameCount += ucount
	// }

	return gameCount
}

// 5.11-5.20 新注册用户，游戏局数在1-5局内无充值行为的游戏数据
func UserGameDataByA1_5(begin, end string) {
	beego.Info("开始执行 UserGameDataByA1_5")
	s := fmt.Sprintf("%s 00:00:00", begin)
	sTime := utils.Str2TimeZone(s, locationName)
	e := fmt.Sprintf("%s 23:59:59", end)
	eTime := utils.Str2TimeZone(e, locationName)

	// A类注册用户id
	a_uid := make([]string, 0)
	n := bson.M{}
	n["robot"] = false
	n["simulation_robot"] = false
	n["ctime"] = bson.M{"$gte": sTime, "$lte": eTime}
	n["regist_area"] = bson.M{"$ne": 1}
	PlayerUsers.Find(n).Distinct("_id", &a_uid)
	beego.Info("UserGameDataByA1_5 查询注册用户")

	// 判断是否有充值行为
	p_uid := make([]string, 0)
	n1 := bson.M{}
	n1["order_status"] = 2
	n1["userid"] = bson.M{"$in": a_uid}
	Pays.Find(n1).Distinct("userid", &p_uid)
	d_uid := make([]string, 0)
	if len(p_uid) > 0 {
		// 将有充值行为的userid去掉
		for _, a := range a_uid {
			found := false
			for _, p := range p_uid {
				if a == p {
					found = true
					break
				}
			}
			if !found {
				d_uid = append(d_uid, a)
			}
		}
	} else {
		d_uid = a_uid
	}
	beego.Info("UserGameDataByA1_5 去掉有充值行为用户")

	s1 := fmt.Sprintf("%s 00:00:00", "2024-05-11")
	startTime := utils.Str2TimeZone(s1, locationName)
	e1 := fmt.Sprintf("%s 23:59:59", "2024-05-20")
	endTime := utils.Str2TimeZone(e1, locationName)
	statrTimestamp := utils.Time2Stamp(startTime)
	endTimestamp := utils.Time2Stamp(endTime)
	// 获取游戏局数在1-5局的用户
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"userid": bson.M{"$in": d_uid},
				"ctime":  bson.M{"$gte": statrTimestamp, "$lte": endTimestamp},
			},
		},
		{
			"$group": bson.M{
				"_id":   "$userid",
				"count": bson.M{"$sum": 1}, // 添加一个额外的字段以便后续过滤
			},
		},
		{
			"$match": bson.M{
				"count": bson.M{"$gte": 1, "$lte": 5}, // 过滤num值在1-5之间的文档
			},
		},
	}
	result := []bson.M{}
	LogGameTimes.Pipe(pipeline).All(&result)
	t_map := make(map[string]int, 0)
	for _, item := range result {
		uid := item["_id"].(string)
		ucount := item["count"].(int)
		t_map[uid] = ucount
	}
	cp_list := make([]entity.CPDetail, 0)       // 彩票数据
	crash_list := make([]entity.CRASHDetail, 0) // crash数据
	lhd_list := make([]entity.LHDetail, 0)      // 龙虎斗数据
	up_list := make([]entity.UPDetail, 0)       // 7up数据
	fj_list := make([]entity.CRASHDetail, 0)    // 飞机数据
	tp_list := make([]entity.TPDetail, 0)       // tp数据
	ak_list := make([]entity.AK47Detail, 0)     // ak47数据
	ab_list := make([]entity.ABDetail, 0)       // ab数据
	joker_list := make([]entity.JOKERDetail, 0) // joker数据
	rm_list := make([]entity.RMDetail, 0)       // rm数据
	// 查询玩家游戏结果
	for k, _ := range t_map {
		// 对局
		pipeline2 := []bson.M{
			{
				"$match": bson.M{
					"end_time": bson.M{"$gte": statrTimestamp, "$lte": endTimestamp},
					"players": bson.M{
						"$regex":   fmt.Sprintf("\\b%s\\b", k),
						"$options": "i",
					},
				},
			},
		}
		var result2 []entity.Detail
		Details.Pipe(pipeline2).All(&result2)
		// 如果查询出来大于五则不处理该数据
		if len(result2) <= 5 {
			for _, dt := range result2 {
				switch dt.Gtype {
				case 1:
					// TP
					if len(dt.TPDetail) > 0 {
						for _, v := range dt.TPDetail {
							// 只获取该用户得对局详情
							new_info := new(entity.TPDetail)
							new_info.WaterId = dt.WaterId
							new_info.UserId = v.UserId
							strtem := ""
							for _, c := range v.Cards {
								name := CardToStr(c)
								strtem += " " + name
							}
							new_info.CardStr = strtem
							new_info.FScore = Chip2Float(v.Score)
							new_info.FBeforeScore = Chip2Float(v.BeforeScore)
							new_info.FAfterScore = Chip2Float(v.AfterScore)
							for _, tn := range v.TPDetailTurn {
								num := strconv.Itoa(tn.Turn + 1)
								switch tn.Turn {
								case 0:
									new_info.TurnStr1 = "轮次" + num + ":" + tn.Operation
								case 1:
									new_info.TurnStr2 = "轮次" + num + ":" + tn.Operation
								case 2:
									new_info.TurnStr3 = "轮次" + num + ":" + tn.Operation
								case 3:
									new_info.TurnStr4 = "轮次" + num + ":" + tn.Operation
								}
							}
							tp_list = append(tp_list, *new_info)
						}
					}
				case 2:
					// LHD
					if dt.LHDetail != nil {
						for _, v := range dt.LHDetail.UserDetail {
							new_info := new(entity.LHDetail)
							new_info.WaterId = dt.WaterId
							new_info.Userid = v.Userid
							new_info.Winner = dt.LHDetail.Winner
							new_info.DragonValue = dt.LHDetail.DragonValue
							new_info.TigerValue = dt.LHDetail.TigerValue
							new_info.FBets = Chip2Float(dt.LHDetail.Bets)
							new_info.FPlayerWin = Chip2Float(dt.LHDetail.PlayerWin)
							new_info.FDragon = Chip2Float(v.Dragon)
							new_info.FTiger = Chip2Float(v.Tiger)
							new_info.FTie = Chip2Float(v.Tie)
							new_info.Result = v.Result
							new_info.FWin = Chip2Float(v.Win)
							new_info.FBeforeScore = Chip2Float(v.BeforeScore)
							new_info.FAfterScore = Chip2Float(v.AfterScore)
							lhd_list = append(lhd_list, *new_info)
						}
					}
				case 3:
					// 7UP
					if dt.UPDetail != nil {
						for _, v := range dt.UPDetail.UserDetail {
							new_info := new(entity.UPDetail)
							new_info.WaterId = dt.WaterId
							new_info.Userid = v.Userid
							new_info.Winner = dt.UPDetail.Winner
							new_info.PointValue = dt.UPDetail.PointValue
							new_info.FBets = Chip2Float(dt.UPDetail.Bets)
							new_info.FPlayerWin = Chip2Float(dt.UPDetail.PlayerWin)
							new_info.FDragon = Chip2Float(v.Dragon)
							new_info.FTiger = Chip2Float(v.Tiger)
							new_info.FTie = Chip2Float(v.Tie)
							new_info.Result = v.Result
							new_info.FWin = Chip2Float(v.Win)
							new_info.FBeforeScore = Chip2Float(v.BeforeScore)
							new_info.FAfterScore = Chip2Float(v.AfterScore)
							up_list = append(up_list, *new_info)
						}
					}
				case 4:
					// RUMMY
					if len(dt.RMDetail) > 0 {
						for _, v := range dt.RMDetail {
							new_info := new(entity.RMDetail)
							new_info.WaterId = dt.WaterId
							new_info.UserId = v.UserId
							finalcardsstr := ""
							for _, c := range v.FinalCards {
								for _, b := range c {
									name := CardToStr(b)
									finalcardsstr += " " + name
								}
							}

							initcardsstr := ""
							for _, c := range v.InitCards {
								name := CardToStr(c)
								initcardsstr += " " + name
							}
							wildcardstr := CardToStr(v.WildCard)
							new_info.FinalCardsStr = finalcardsstr
							new_info.InitCardsStr = initcardsstr
							new_info.WildCardStr = wildcardstr
							new_info.Result = v.Result
							new_info.FScore = Chip2Float(v.Score)
							new_info.FBeforeScore = Chip2Float(v.BeforeScore)
							new_info.FAfterScore = Chip2Float(v.AfterScore)
							rm_list = append(rm_list, *new_info)
						}
					}
				case 5:
					// AK47
					if len(dt.AK47Detail) > 0 {
						for _, v := range dt.AK47Detail {
							new_info := new(entity.AK47Detail)
							new_info.WaterId = dt.WaterId
							new_info.UserId = v.UserId
							cardstr := ""
							for _, c := range v.Cards {
								name := CardToStr(c)
								cardstr += " " + name
							}
							new_info.CardStr = cardstr
							new_info.FScore = Chip2Float(v.Score)
							new_info.FBeforeScore = Chip2Float(v.BeforeScore)
							new_info.FAfterScore = Chip2Float(v.AfterScore)
							for _, tn := range v.AK47DetailTurn {
								num := strconv.Itoa(tn.Turn + 1)
								switch tn.Turn {
								case 0:
									new_info.TurnStr1 = "轮次" + num + ":" + tn.Operation
								case 1:
									new_info.TurnStr2 = "轮次" + num + ":" + tn.Operation
								case 2:
									new_info.TurnStr3 = "轮次" + num + ":" + tn.Operation
								case 3:
									new_info.TurnStr4 = "轮次" + num + ":" + tn.Operation
								case 4:
									new_info.TurnStr5 = "轮次" + num + ":" + tn.Operation
								}
							}
							ak_list = append(ak_list, *new_info)
						}
					}
				case 6:
					// Joker
					if len(dt.JOKERDetail) > 0 {
						for _, v := range dt.JOKERDetail {
							new_info := new(entity.JOKERDetail)
							new_info.WaterId = dt.WaterId
							new_info.UserId = v.UserId
							cardstr := ""
							for _, c := range v.Cards {
								name := CardToStr(c)
								cardstr += " " + name
							}
							new_info.CardStr = cardstr
							new_info.FScore = Chip2Float(v.Score)
							new_info.FBeforeScore = Chip2Float(v.BeforeScore)
							new_info.FAfterScore = Chip2Float(v.AfterScore)
							for _, tn := range v.JOKERDetailTurn {
								num := strconv.Itoa(tn.Turn + 1)
								switch tn.Turn {
								case 0:
									new_info.TurnStr1 = "轮次" + num + ":" + tn.Operation
								case 1:
									new_info.TurnStr2 = "轮次" + num + ":" + tn.Operation
								case 2:
									new_info.TurnStr3 = "轮次" + num + ":" + tn.Operation
								case 3:
									new_info.TurnStr4 = "轮次" + num + ":" + tn.Operation
								}
							}
							joker_list = append(joker_list, *new_info)
						}
					}
				case 7:
					// Crash
					if dt.CRASHDetail != nil {
						for _, v := range dt.CRASHDetail.UserDetail {
							new_info := new(entity.CRASHDetail)
							new_info.WaterId = dt.WaterId
							new_info.Userid = v.Userid
							new_info.PlayerLose = dt.CRASHDetail.PlayerLose
							new_info.FBets = Chip2Float(dt.CRASHDetail.Bets)
							new_info.FPlayerWin = Chip2Float(dt.CRASHDetail.PlayerLose)
							new_info.Multiple = v.Multiple
							new_info.DResult = v.Result
							new_info.FBet = Chip2Float(v.Bet)
							new_info.Result = dt.CRASHDetail.Result
							new_info.FWin = Chip2Float(v.Win)
							new_info.FBeforeScore = Chip2Float(v.BeforeScore)
							new_info.FAfterScore = Chip2Float(v.AfterScore)
							new_info.FPlayerFactor = Chip2Float(int64(dt.PlayerFactor))
							crash_list = append(crash_list, *new_info)
						}
					}
				case 8:
					// AB
					if dt.ABDetail != nil {
						for _, v := range dt.ABDetail.UserDetail {
							new_info := new(entity.ABDetail)
							new_info.WaterId = dt.WaterId
							new_info.Userid = v.Userid
							acradstr := ""
							for _, c := range dt.ABDetail.ACards {
								name := CardToStr(c)
								acradstr += " " + name
							}
							bcradstr := ""
							for _, c := range dt.ABDetail.BCards {
								name := CardToStr(c)
								bcradstr += " " + name
							}
							new_info.JokerStr = CardToStr(dt.ABDetail.Joker) // dt.ABDetail.Joker
							new_info.JackpotValueStr = CardToStr(dt.ABDetail.JackpotValue)
							new_info.ANDAR = acradstr
							new_info.BAHAR = bcradstr
							if dt.ABDetail.Winner == 1 {
								new_info.WinnerStr = "ANDAR"
							} else if dt.ABDetail.Winner == 2 {
								new_info.WinnerStr = "BAHAR"
							}
							if dt.ABDetail.SideWinner == 3 {
								new_info.SideWinnerStr = "1-5"
							} else if dt.ABDetail.SideWinner == 4 {
								new_info.SideWinnerStr = "6-10"
							} else if dt.ABDetail.SideWinner == 5 {
								new_info.SideWinnerStr = "11-15"
							} else if dt.ABDetail.SideWinner == 6 {
								new_info.SideWinnerStr = "16-25"
							} else if dt.ABDetail.SideWinner == 7 {
								new_info.SideWinnerStr = "26-30"
							} else if dt.ABDetail.SideWinner == 8 {
								new_info.SideWinnerStr = "31-35"
							} else if dt.ABDetail.SideWinner == 9 {
								new_info.SideWinnerStr = "36-40"
							} else if dt.ABDetail.SideWinner == 10 {
								new_info.SideWinnerStr = "41以上"
							}
							new_info.FBets = Chip2Float(dt.ABDetail.Bets)
							new_info.FPlayerWin = Chip2Float(dt.ABDetail.PlayerWin)
							new_info.Result = v.Result
							new_info.FWin = Chip2Float(v.Win)
							new_info.FBeforeScore = Chip2Float(v.BeforeScore)
							new_info.FAfterScore = Chip2Float(v.AfterScore)
							for i, b := range v.SeatBets {
								switch i {
								case "1":
									new_info.SeatBets1 = Chip2Float(int64(b))
								case "2":
									new_info.SeatBets2 = Chip2Float(int64(b))
								case "3":
									new_info.SeatBets3 = Chip2Float(int64(b))
								case "4":
									new_info.SeatBets4 = Chip2Float(int64(b))
								case "5":
									new_info.SeatBets5 = Chip2Float(int64(b))
								case "6":
									new_info.SeatBets6 = Chip2Float(int64(b))
								case "7":
									new_info.SeatBets7 = Chip2Float(int64(b))
								case "8":
									new_info.SeatBets8 = Chip2Float(int64(b))
								case "9":
									new_info.SeatBets9 = Chip2Float(int64(b))
								case "10":
									new_info.SeatBets10 = Chip2Float(int64(b))
								}
							}
							ab_list = append(ab_list, *new_info)
						}
					}
				case 9:
					// 彩票
					if dt.CPDetail != nil {
						for _, v := range dt.CPDetail.UserDetail {
							new_info := new(entity.CPDetail)
							new_info.WaterId = dt.WaterId
							new_info.Userid = v.Userid
							cardstr := ""
							for _, c := range dt.CPDetail.Cards {
								name := CardToStr(c)
								cardstr += " " + name
							}
							new_info.CardStr = cardstr
							new_info.FBets = Chip2Float(dt.CPDetail.Bets)
							new_info.FPlayerWin = Chip2Float(dt.CPDetail.PlayerWin)
							for i, b := range v.SeatBets {
								switch i {
								case "1":
									new_info.SeatBets1 = Chip2Float(int64(b))
								case "2":
									new_info.SeatBets2 = Chip2Float(int64(b))
								case "3":
									new_info.SeatBets3 = Chip2Float(int64(b))
								case "4":
									new_info.SeatBets4 = Chip2Float(int64(b))
								case "5":
									new_info.SeatBets5 = Chip2Float(int64(b))
								case "6":
									new_info.SeatBets6 = Chip2Float(int64(b))
								}
							}

							new_info.Result = v.Result
							new_info.FWin = Chip2Float(v.Win)
							new_info.FBeforeScore = Chip2Float(v.BeforeScore)
							new_info.FAfterScore = Chip2Float(v.AfterScore)
							cp_list = append(cp_list, *new_info)
						}
					}
				case 10:
					// 飞机
					if dt.CRASHDetail != nil {
						for _, v := range dt.CRASHDetail.UserDetail {
							new_info := new(entity.CRASHDetail)
							new_info.WaterId = dt.WaterId
							new_info.Userid = v.Userid
							new_info.PlayerLose = dt.CRASHDetail.PlayerLose
							new_info.FBets = Chip2Float(dt.CRASHDetail.Bets)
							new_info.FPlayerWin = Chip2Float(dt.CRASHDetail.PlayerLose)
							new_info.Multiple = v.Multiple
							new_info.DResult = v.Result
							new_info.FBet = Chip2Float(v.Bet)
							new_info.Result = dt.CRASHDetail.Result
							new_info.FWin = Chip2Float(v.Win)
							new_info.FBeforeScore = Chip2Float(v.BeforeScore)
							new_info.FAfterScore = Chip2Float(v.AfterScore)
							new_info.FPlayerFactor = Chip2Float(int64(dt.PlayerFactor))
							fj_list = append(fj_list, *new_info)
						}
					}
				}
			}
		}
	}

	// TP
	if len(tp_list) > 0 {
		beego.Info("UserGameDataByA1_5 开始执行TP数据导出")
		// 导出
		f := excelize.NewFile()
		sheet1 := "Sheet1"
		// 设置表头
		f.SetCellValue(sheet1, "A1", "用户id")
		f.SetCellValue(sheet1, "B1", "局号")
		f.SetCellValue(sheet1, "C1", "牌型")
		f.SetCellValue(sheet1, "D1", "结算")
		f.SetCellValue(sheet1, "E1", "账变前分数")
		f.SetCellValue(sheet1, "F1", "账变后分数")
		f.SetCellValue(sheet1, "G1", "轮次1")
		f.SetCellValue(sheet1, "H1", "轮次2")
		f.SetCellValue(sheet1, "I1", "轮次3")
		f.SetCellValue(sheet1, "J1", "轮次4")
		line := 1
		for _, player := range tp_list {
			line++
			f.SetCellValue(sheet1, fmt.Sprintf("A%d", line), player.UserId)
			f.SetCellValue(sheet1, fmt.Sprintf("B%d", line), player.WaterId)
			f.SetCellValue(sheet1, fmt.Sprintf("C%d", line), player.CardStr)
			f.SetCellValue(sheet1, fmt.Sprintf("D%d", line), player.FScore)
			f.SetCellValue(sheet1, fmt.Sprintf("E%d", line), player.FBeforeScore)
			f.SetCellValue(sheet1, fmt.Sprintf("F%d", line), player.FAfterScore)
			f.SetCellValue(sheet1, fmt.Sprintf("G%d", line), player.TurnStr1)
			f.SetCellValue(sheet1, fmt.Sprintf("H%d", line), player.TurnStr2)
			f.SetCellValue(sheet1, fmt.Sprintf("I%d", line), player.TurnStr3)
			f.SetCellValue(sheet1, fmt.Sprintf("J%d", line), player.TurnStr4)
		}
		now := time.Now().Format("2006-01-02.15.04.05")
		filename := begin + "A类用户TP游戏" + "_" + fmt.Sprint(now) + ".xlsx"
		// 保存文件
		if err := f.SaveAs(filename); err != nil {
			glog.Error("error3: ", err)
		}
	}

	// 龙虎斗
	if len(lhd_list) > 0 {
		beego.Info("UserGameDataByA1_5 开始执行LHD数据导出")
		// 导出
		f := excelize.NewFile()
		sheet1 := "Sheet1"
		// 设置表头
		f.SetCellValue(sheet1, "A1", "局号")
		f.SetCellValue(sheet1, "B1", "开奖结果")
		f.SetCellValue(sheet1, "C1", "龙")
		f.SetCellValue(sheet1, "D1", "虎")
		f.SetCellValue(sheet1, "E1", "总下注")
		f.SetCellValue(sheet1, "F1", "玩家赢分")
		f.SetCellValue(sheet1, "G1", "下注用户ID")
		f.SetCellValue(sheet1, "H1", "龙")
		f.SetCellValue(sheet1, "I1", "虎")
		f.SetCellValue(sheet1, "J1", "和")
		f.SetCellValue(sheet1, "K1", "结果")
		f.SetCellValue(sheet1, "L1", "结算")
		f.SetCellValue(sheet1, "M1", "账变前分数")
		f.SetCellValue(sheet1, "N1", "账变前分数")
		line := 1
		for _, player := range lhd_list {
			line++
			f.SetCellValue(sheet1, fmt.Sprintf("A%d", line), player.WaterId)
			f.SetCellValue(sheet1, fmt.Sprintf("B%d", line), player.Winner)
			f.SetCellValue(sheet1, fmt.Sprintf("C%d", line), player.DragonValue)
			f.SetCellValue(sheet1, fmt.Sprintf("D%d", line), player.TigerValue)
			f.SetCellValue(sheet1, fmt.Sprintf("E%d", line), player.FBets)
			f.SetCellValue(sheet1, fmt.Sprintf("F%d", line), player.FPlayerWin)
			f.SetCellValue(sheet1, fmt.Sprintf("G%d", line), player.Userid)
			f.SetCellValue(sheet1, fmt.Sprintf("H%d", line), player.FDragon)
			f.SetCellValue(sheet1, fmt.Sprintf("I%d", line), player.FTiger)
			f.SetCellValue(sheet1, fmt.Sprintf("J%d", line), player.FTie)
			f.SetCellValue(sheet1, fmt.Sprintf("K%d", line), player.FWin)
			f.SetCellValue(sheet1, fmt.Sprintf("L%d", line), player.Result)
			f.SetCellValue(sheet1, fmt.Sprintf("M%d", line), player.FBeforeScore)
			f.SetCellValue(sheet1, fmt.Sprintf("N%d", line), player.FAfterScore)
		}
		now := time.Now().Format("2006-01-02.15.04.05")
		filename := begin + "A类用户LHD游戏" + "_" + fmt.Sprint(now) + ".xlsx"
		// 保存文件
		if err := f.SaveAs(filename); err != nil {
			glog.Error("error3: ", err)
		}
	}

	// 7updown
	if len(up_list) > 0 {
		beego.Info("UserGameDataByA1_5 开始执行7updown数据导出")
		// 导出
		f := excelize.NewFile()
		sheet1 := "Sheet1"
		// 设置表头
		f.SetCellValue(sheet1, "A1", "用户id")
		f.SetCellValue(sheet1, "B1", "局号")
		f.SetCellValue(sheet1, "C1", "玩家最终系数")
		f.SetCellValue(sheet1, "D1", "开奖结果")
		f.SetCellValue(sheet1, "E1", "点数")
		f.SetCellValue(sheet1, "F1", "总下注")
		f.SetCellValue(sheet1, "G1", "玩家赢分")
		f.SetCellValue(sheet1, "H1", "大")
		f.SetCellValue(sheet1, "I1", "小")
		f.SetCellValue(sheet1, "J1", "7")
		f.SetCellValue(sheet1, "K1", "结果")
		f.SetCellValue(sheet1, "L1", "结算")
		f.SetCellValue(sheet1, "M1", "账变前分数")
		f.SetCellValue(sheet1, "N1", "账变后分数")
		line := 1
		for _, player := range up_list {
			line++
			f.SetCellValue(sheet1, fmt.Sprintf("A%d", line), player.Userid)
			f.SetCellValue(sheet1, fmt.Sprintf("B%d", line), player.WaterId)
			f.SetCellValue(sheet1, fmt.Sprintf("C%d", line), player.FPlayerFactor)
			f.SetCellValue(sheet1, fmt.Sprintf("D%d", line), player.Winner)
			f.SetCellValue(sheet1, fmt.Sprintf("E%d", line), player.PointValue)
			f.SetCellValue(sheet1, fmt.Sprintf("F%d", line), player.FBets)
			f.SetCellValue(sheet1, fmt.Sprintf("G%d", line), player.FPlayerWin)
			f.SetCellValue(sheet1, fmt.Sprintf("H%d", line), player.FDragon)
			f.SetCellValue(sheet1, fmt.Sprintf("I%d", line), player.FTiger)
			f.SetCellValue(sheet1, fmt.Sprintf("J%d", line), player.FTie)
			f.SetCellValue(sheet1, fmt.Sprintf("K%d", line), player.Result)
			f.SetCellValue(sheet1, fmt.Sprintf("L%d", line), player.FWin)
			f.SetCellValue(sheet1, fmt.Sprintf("M%d", line), player.FBeforeScore)
			f.SetCellValue(sheet1, fmt.Sprintf("N%d", line), player.FAfterScore)

		}
		now := time.Now().Format("2006-01-02.15.04.05")
		filename := begin + "A类用户7updown游戏" + "_" + fmt.Sprint(now) + ".xlsx"
		// 保存文件
		if err := f.SaveAs(filename); err != nil {
			glog.Error("error3: ", err)
		}
	}

	// Rummy
	if len(rm_list) > 0 {
		beego.Info("UserGameDataByA1_5 开始执行Rummy数据导出")
		// 导出
		f := excelize.NewFile()
		sheet1 := "Sheet1"
		// 设置表头
		f.SetCellValue(sheet1, "A1", "用户id")
		f.SetCellValue(sheet1, "B1", "局号")
		f.SetCellValue(sheet1, "C1", "结果")
		f.SetCellValue(sheet1, "D1", "结算")
		f.SetCellValue(sheet1, "E1", "账变前分数")
		f.SetCellValue(sheet1, "F1", "账变后分数")
		f.SetCellValue(sheet1, "G1", "最终牌型")
		f.SetCellValue(sheet1, "H1", "最初牌型")
		f.SetCellValue(sheet1, "I1", "万能牌")
		line := 1
		for _, player := range rm_list {
			line++
			f.SetCellValue(sheet1, fmt.Sprintf("A%d", line), player.UserId)
			f.SetCellValue(sheet1, fmt.Sprintf("B%d", line), player.WaterId)
			f.SetCellValue(sheet1, fmt.Sprintf("C%d", line), player.Result)
			f.SetCellValue(sheet1, fmt.Sprintf("D%d", line), player.FScore)
			f.SetCellValue(sheet1, fmt.Sprintf("E%d", line), player.FBeforeScore)
			f.SetCellValue(sheet1, fmt.Sprintf("F%d", line), player.FAfterScore)
			f.SetCellValue(sheet1, fmt.Sprintf("G%d", line), player.FinalCardsStr)
			f.SetCellValue(sheet1, fmt.Sprintf("H%d", line), player.InitCardsStr)
			f.SetCellValue(sheet1, fmt.Sprintf("I%d", line), player.WildCardStr)

		}
		now := time.Now().Format("2006-01-02.15.04.05")
		filename := begin + "A类用户Rummy游戏" + "_" + fmt.Sprint(now) + ".xlsx"
		// 保存文件
		if err := f.SaveAs(filename); err != nil {
			glog.Error("error3: ", err)
		}
	}

	// AK47
	if len(ak_list) > 0 {
		beego.Info("UserGameDataByA1_5 开始执行AK47数据导出")
		// 导出
		f := excelize.NewFile()
		sheet1 := "Sheet1"
		// 设置表头
		f.SetCellValue(sheet1, "A1", "用户id")
		f.SetCellValue(sheet1, "B1", "局号")
		f.SetCellValue(sheet1, "C1", "牌型")
		f.SetCellValue(sheet1, "D1", "结算")
		f.SetCellValue(sheet1, "E1", "账变前分数")
		f.SetCellValue(sheet1, "F1", "账变后分数")
		f.SetCellValue(sheet1, "G1", "轮次1")
		f.SetCellValue(sheet1, "H1", "轮次2")
		f.SetCellValue(sheet1, "I1", "轮次3")
		f.SetCellValue(sheet1, "J1", "轮次4")
		f.SetCellValue(sheet1, "K1", "轮次5")
		line := 1
		for _, player := range ak_list {
			line++
			f.SetCellValue(sheet1, fmt.Sprintf("A%d", line), player.UserId)
			f.SetCellValue(sheet1, fmt.Sprintf("B%d", line), player.WaterId)
			f.SetCellValue(sheet1, fmt.Sprintf("C%d", line), player.CardStr)
			f.SetCellValue(sheet1, fmt.Sprintf("D%d", line), player.FScore)
			f.SetCellValue(sheet1, fmt.Sprintf("E%d", line), player.FBeforeScore)
			f.SetCellValue(sheet1, fmt.Sprintf("F%d", line), player.FAfterScore)
			f.SetCellValue(sheet1, fmt.Sprintf("G%d", line), player.TurnStr1)
			f.SetCellValue(sheet1, fmt.Sprintf("H%d", line), player.TurnStr2)
			f.SetCellValue(sheet1, fmt.Sprintf("I%d", line), player.TurnStr3)
			f.SetCellValue(sheet1, fmt.Sprintf("J%d", line), player.TurnStr4)
			f.SetCellValue(sheet1, fmt.Sprintf("K%d", line), player.TurnStr5)

		}
		now := time.Now().Format("2006-01-02.15.04.05")
		filename := begin + "A类用户AK47游戏" + "_" + fmt.Sprint(now) + ".xlsx"
		// 保存文件
		if err := f.SaveAs(filename); err != nil {
			glog.Error("error3: ", err)
		}
	}

	// joker
	if len(joker_list) > 0 {
		beego.Info("UserGameDataByA1_5 开始执行Joker数据导出")
		// 导出
		f := excelize.NewFile()
		sheet1 := "Sheet1"
		// 设置表头
		f.SetCellValue(sheet1, "A1", "用户id")
		f.SetCellValue(sheet1, "B1", "局号")
		f.SetCellValue(sheet1, "C1", "牌型")
		f.SetCellValue(sheet1, "D1", "结算")
		f.SetCellValue(sheet1, "E1", "账变前分数")
		f.SetCellValue(sheet1, "F1", "账变后分数")
		f.SetCellValue(sheet1, "G1", "轮次1")
		f.SetCellValue(sheet1, "H1", "轮次2")
		f.SetCellValue(sheet1, "I1", "轮次3")
		f.SetCellValue(sheet1, "J1", "轮次4")
		line := 1
		for _, player := range joker_list {
			line++
			f.SetCellValue(sheet1, fmt.Sprintf("A%d", line), player.UserId)
			f.SetCellValue(sheet1, fmt.Sprintf("B%d", line), player.WaterId)
			f.SetCellValue(sheet1, fmt.Sprintf("C%d", line), player.CardStr)
			f.SetCellValue(sheet1, fmt.Sprintf("D%d", line), player.FScore)
			f.SetCellValue(sheet1, fmt.Sprintf("E%d", line), player.FBeforeScore)
			f.SetCellValue(sheet1, fmt.Sprintf("F%d", line), player.FAfterScore)
			f.SetCellValue(sheet1, fmt.Sprintf("G%d", line), player.TurnStr1)
			f.SetCellValue(sheet1, fmt.Sprintf("H%d", line), player.TurnStr2)
			f.SetCellValue(sheet1, fmt.Sprintf("I%d", line), player.TurnStr3)
			f.SetCellValue(sheet1, fmt.Sprintf("J%d", line), player.TurnStr4)

		}
		now := time.Now().Format("2006-01-02.15.04.05")
		filename := begin + "A类用户Joker游戏" + "_" + fmt.Sprint(now) + ".xlsx"
		// 保存文件
		if err := f.SaveAs(filename); err != nil {
			glog.Error("error3: ", err)
		}
	}

	// Crash
	if len(crash_list) > 0 {
		beego.Info("UserGameDataByA1_5 开始执行Crash数据导出")
		// 导出
		f := excelize.NewFile()
		sheet1 := "Sheet1"
		// 设置表头
		f.SetCellValue(sheet1, "A1", "用户id")
		f.SetCellValue(sheet1, "B1", "局号")
		f.SetCellValue(sheet1, "C1", "玩家最终系数")
		f.SetCellValue(sheet1, "D1", "开奖结果")
		f.SetCellValue(sheet1, "E1", "总下注")
		f.SetCellValue(sheet1, "F1", "玩家赢分")
		f.SetCellValue(sheet1, "G1", "分数")
		f.SetCellValue(sheet1, "H1", "逃脱倍数")
		f.SetCellValue(sheet1, "I1", "结果")
		f.SetCellValue(sheet1, "J1", "结算")
		f.SetCellValue(sheet1, "K1", "账变前分数")
		f.SetCellValue(sheet1, "L1", "账变后分数")
		line := 1
		for _, player := range crash_list {
			line++
			f.SetCellValue(sheet1, fmt.Sprintf("A%d", line), player.Userid)
			f.SetCellValue(sheet1, fmt.Sprintf("B%d", line), player.WaterId)
			f.SetCellValue(sheet1, fmt.Sprintf("C%d", line), player.FPlayerFactor)
			f.SetCellValue(sheet1, fmt.Sprintf("D%d", line), player.Result)
			f.SetCellValue(sheet1, fmt.Sprintf("E%d", line), player.FBets)
			f.SetCellValue(sheet1, fmt.Sprintf("F%d", line), player.FPlayerWin)
			f.SetCellValue(sheet1, fmt.Sprintf("G%d", line), player.FBet)
			f.SetCellValue(sheet1, fmt.Sprintf("H%d", line), player.Multiple)
			f.SetCellValue(sheet1, fmt.Sprintf("I%d", line), player.DResult)
			f.SetCellValue(sheet1, fmt.Sprintf("J%d", line), player.FWin)
			f.SetCellValue(sheet1, fmt.Sprintf("K%d", line), player.FBeforeScore)
			f.SetCellValue(sheet1, fmt.Sprintf("L%d", line), player.FAfterScore)
		}
		now := time.Now().Format("2006-01-02.15.04.05")
		filename := begin + "A类用户Crash游戏" + "_" + fmt.Sprint(now) + ".xlsx"
		// 保存文件
		if err := f.SaveAs(filename); err != nil {
			glog.Error("error3: ", err)
		}
	}

	// AB
	if len(ab_list) > 0 {
		beego.Info("UserGameDataByA1_5 开始执行AB数据导出")
		// 导出
		f := excelize.NewFile()
		sheet1 := "Sheet1"
		// 设置表头
		f.SetCellValue(sheet1, "A1", "用户id")
		f.SetCellValue(sheet1, "B1", "局号")
		f.SetCellValue(sheet1, "C1", "key牌")
		f.SetCellValue(sheet1, "D1", "中奖牌")
		f.SetCellValue(sheet1, "E1", "开奖结果1")
		f.SetCellValue(sheet1, "F1", "开奖结果2")
		f.SetCellValue(sheet1, "G1", "总下注")
		f.SetCellValue(sheet1, "H1", "玩家赢分")
		f.SetCellValue(sheet1, "I1", "ANDAR")
		f.SetCellValue(sheet1, "J1", "BAHAR")
		f.SetCellValue(sheet1, "K1", "ANDAR")
		f.SetCellValue(sheet1, "L1", "BAHAR")
		f.SetCellValue(sheet1, "M1", "1-5")
		f.SetCellValue(sheet1, "N1", "6-10")
		f.SetCellValue(sheet1, "O1", "11-15")
		f.SetCellValue(sheet1, "P1", "16-25")
		f.SetCellValue(sheet1, "Q1", "26-30")
		f.SetCellValue(sheet1, "R1", "31-35")
		f.SetCellValue(sheet1, "S1", "36-40")
		f.SetCellValue(sheet1, "T1", "41以上")
		f.SetCellValue(sheet1, "U1", "结果")
		f.SetCellValue(sheet1, "V1", "结算")
		f.SetCellValue(sheet1, "W1", "账变前分数")
		f.SetCellValue(sheet1, "X1", "账变后分数")
		line := 1
		for _, player := range ab_list {
			line++
			f.SetCellValue(sheet1, fmt.Sprintf("A%d", line), player.Userid)
			f.SetCellValue(sheet1, fmt.Sprintf("B%d", line), player.WaterId)
			f.SetCellValue(sheet1, fmt.Sprintf("C%d", line), player.JokerStr)
			f.SetCellValue(sheet1, fmt.Sprintf("D%d", line), player.JackpotValueStr)
			f.SetCellValue(sheet1, fmt.Sprintf("E%d", line), player.WinnerStr)
			f.SetCellValue(sheet1, fmt.Sprintf("F%d", line), player.SideWinnerStr)
			f.SetCellValue(sheet1, fmt.Sprintf("G%d", line), player.FBets)
			f.SetCellValue(sheet1, fmt.Sprintf("H%d", line), player.FPlayerWin)
			f.SetCellValue(sheet1, fmt.Sprintf("I%d", line), player.ANDAR)
			f.SetCellValue(sheet1, fmt.Sprintf("J%d", line), player.BAHAR)
			f.SetCellValue(sheet1, fmt.Sprintf("K%d", line), player.SeatBets1)
			f.SetCellValue(sheet1, fmt.Sprintf("L%d", line), player.SeatBets2)
			f.SetCellValue(sheet1, fmt.Sprintf("M%d", line), player.SeatBets3)
			f.SetCellValue(sheet1, fmt.Sprintf("N%d", line), player.SeatBets4)
			f.SetCellValue(sheet1, fmt.Sprintf("O%d", line), player.SeatBets5)
			f.SetCellValue(sheet1, fmt.Sprintf("P%d", line), player.SeatBets6)
			f.SetCellValue(sheet1, fmt.Sprintf("Q%d", line), player.SeatBets7)
			f.SetCellValue(sheet1, fmt.Sprintf("R%d", line), player.SeatBets8)
			f.SetCellValue(sheet1, fmt.Sprintf("S%d", line), player.SeatBets9)
			f.SetCellValue(sheet1, fmt.Sprintf("T%d", line), player.SeatBets10)
			f.SetCellValue(sheet1, fmt.Sprintf("U%d", line), player.Result)
			f.SetCellValue(sheet1, fmt.Sprintf("V%d", line), player.FWin)
			f.SetCellValue(sheet1, fmt.Sprintf("W%d", line), player.FBeforeScore)
			f.SetCellValue(sheet1, fmt.Sprintf("X%d", line), player.FAfterScore)
		}
		now := time.Now().Format("2006-01-02.15.04.05")
		filename := begin + "A类用户AB游戏" + "_" + fmt.Sprint(now) + ".xlsx"
		// 保存文件
		if err := f.SaveAs(filename); err != nil {
			glog.Error("error3: ", err)
		}
	}

	// cp
	if len(cp_list) > 0 {
		beego.Info("UserGameDataByA1_5 开始执行彩票数据导出")
		// 导出
		f := excelize.NewFile()
		sheet1 := "Sheet1"
		// 设置表头
		f.SetCellValue(sheet1, "A1", "用户id")
		f.SetCellValue(sheet1, "B1", "局号")
		f.SetCellValue(sheet1, "C1", "开奖结果")
		f.SetCellValue(sheet1, "D1", "牌型")
		f.SetCellValue(sheet1, "E1", "总下注")
		f.SetCellValue(sheet1, "F1", "玩家赢分")
		f.SetCellValue(sheet1, "G1", "高牌")
		f.SetCellValue(sheet1, "H1", "对子")
		f.SetCellValue(sheet1, "I1", "同花")
		f.SetCellValue(sheet1, "J1", "顺子")
		f.SetCellValue(sheet1, "K1", "同花顺")
		f.SetCellValue(sheet1, "L1", "豹子")
		f.SetCellValue(sheet1, "M1", "结果")
		f.SetCellValue(sheet1, "N1", "结算")
		f.SetCellValue(sheet1, "O1", "账变前分数")
		f.SetCellValue(sheet1, "P1", "账变后分数")
		line := 1
		for _, player := range cp_list {
			line++
			f.SetCellValue(sheet1, fmt.Sprintf("A%d", line), player.Userid)
			f.SetCellValue(sheet1, fmt.Sprintf("B%d", line), player.WaterId)
			f.SetCellValue(sheet1, fmt.Sprintf("C%d", line), player.JackpotOutput)
			f.SetCellValue(sheet1, fmt.Sprintf("D%d", line), player.CardStr)
			f.SetCellValue(sheet1, fmt.Sprintf("E%d", line), player.FBets)
			f.SetCellValue(sheet1, fmt.Sprintf("F%d", line), player.FPlayerWin)
			f.SetCellValue(sheet1, fmt.Sprintf("G%d", line), player.SeatBets1)
			f.SetCellValue(sheet1, fmt.Sprintf("H%d", line), player.SeatBets2)
			f.SetCellValue(sheet1, fmt.Sprintf("I%d", line), player.SeatBets3)
			f.SetCellValue(sheet1, fmt.Sprintf("J%d", line), player.SeatBets4)
			f.SetCellValue(sheet1, fmt.Sprintf("K%d", line), player.SeatBets5)
			f.SetCellValue(sheet1, fmt.Sprintf("L%d", line), player.SeatBets6)
			f.SetCellValue(sheet1, fmt.Sprintf("M%d", line), player.Result)
			f.SetCellValue(sheet1, fmt.Sprintf("N%d", line), player.FWin)
			f.SetCellValue(sheet1, fmt.Sprintf("O%d", line), player.FBeforeScore)
			f.SetCellValue(sheet1, fmt.Sprintf("P%d", line), player.FAfterScore)
		}
		now := time.Now().Format("2006-01-02.15.04.05")
		filename := begin + "A类用户彩票游戏" + "_" + fmt.Sprint(now) + ".xlsx"
		// 保存文件
		if err := f.SaveAs(filename); err != nil {
			glog.Error("error3: ", err)
		}
	}

	// 飞机
	if len(fj_list) > 0 {
		beego.Info("UserGameDataByA1_5 开始执行飞机数据导出")
		// 导出
		f := excelize.NewFile()
		sheet1 := "Sheet1"
		// 设置表头
		f.SetCellValue(sheet1, "A1", "用户id")
		f.SetCellValue(sheet1, "B1", "局号")
		f.SetCellValue(sheet1, "C1", "玩家最终系数")
		f.SetCellValue(sheet1, "D1", "开奖结果")
		f.SetCellValue(sheet1, "E1", "总下注")
		f.SetCellValue(sheet1, "F1", "玩家赢分")
		f.SetCellValue(sheet1, "G1", "分数")
		f.SetCellValue(sheet1, "H1", "逃脱倍数")
		f.SetCellValue(sheet1, "I1", "结果")
		f.SetCellValue(sheet1, "J1", "结算")
		f.SetCellValue(sheet1, "K1", "账变前分数")
		f.SetCellValue(sheet1, "L1", "账变后分数")
		line := 1
		for _, player := range fj_list {
			line++
			f.SetCellValue(sheet1, fmt.Sprintf("A%d", line), player.Userid)
			f.SetCellValue(sheet1, fmt.Sprintf("B%d", line), player.WaterId)
			f.SetCellValue(sheet1, fmt.Sprintf("C%d", line), player.FPlayerFactor)
			f.SetCellValue(sheet1, fmt.Sprintf("D%d", line), player.Result)
			f.SetCellValue(sheet1, fmt.Sprintf("E%d", line), player.FBets)
			f.SetCellValue(sheet1, fmt.Sprintf("F%d", line), player.FPlayerWin)
			f.SetCellValue(sheet1, fmt.Sprintf("G%d", line), player.FBet)
			f.SetCellValue(sheet1, fmt.Sprintf("H%d", line), player.Multiple)
			f.SetCellValue(sheet1, fmt.Sprintf("I%d", line), player.DResult)
			f.SetCellValue(sheet1, fmt.Sprintf("J%d", line), player.FWin)
			f.SetCellValue(sheet1, fmt.Sprintf("K%d", line), player.FBeforeScore)
			f.SetCellValue(sheet1, fmt.Sprintf("L%d", line), player.FAfterScore)
		}
		now := time.Now().Format("2006-01-02.15.04.05")
		filename := begin + "A类用户飞机游戏" + "_" + fmt.Sprint(now) + ".xlsx"
		// 保存文件
		if err := f.SaveAs(filename); err != nil {
			glog.Error("error3: ", err)
		}
	}
}

// 5.11-5.20 新注册用户，游戏局数在1-5局内无充值行为的游戏数据
func UserGameDataByB1_5(begin, end string) {
	beego.Info("开始执行 UserGameDataByB1_5")
	s := fmt.Sprintf("%s 00:00:00", begin)
	sTime := utils.Str2TimeZone(s, locationName)
	e := fmt.Sprintf("%s 23:59:59", end)
	eTime := utils.Str2TimeZone(e, locationName)

	// A类注册用户id
	a_uid := make([]string, 0)
	n := bson.M{}
	n["robot"] = false
	n["simulation_robot"] = false
	n["ctime"] = bson.M{"$gte": sTime, "$lte": eTime}
	n["regist_area"] = bson.M{"$eq": 1}
	PlayerUsers.Find(n).Distinct("_id", &a_uid)
	beego.Info("UserGameDataByB1_5 查询注册用户")

	// 判断是否有充值行为
	p_uid := make([]string, 0)
	n1 := bson.M{}
	n1["order_status"] = 2
	n1["userid"] = bson.M{"$in": a_uid}
	Pays.Find(n1).Distinct("userid", &p_uid)
	d_uid := make([]string, 0)
	if len(p_uid) > 0 {
		// 将有充值行为的userid去掉
		for _, a := range a_uid {
			found := false
			for _, p := range p_uid {
				if a == p {
					found = true
					break
				}
			}
			if !found {
				d_uid = append(d_uid, a)
			}
		}
	} else {
		d_uid = a_uid
	}
	beego.Info("UserGameDataByB1_5 去掉有充值行为用户")

	s1 := fmt.Sprintf("%s 00:00:00", "2024-05-11")
	startTime := utils.Str2TimeZone(s1, locationName)
	e1 := fmt.Sprintf("%s 23:59:59", "2024-05-20")
	endTime := utils.Str2TimeZone(e1, locationName)
	statrTimestamp := utils.Time2Stamp(startTime)
	endTimestamp := utils.Time2Stamp(endTime)
	// 获取游戏局数在1-5局的用户
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"userid": bson.M{"$in": d_uid},
				"ctime":  bson.M{"$gte": statrTimestamp, "$lte": endTimestamp},
			},
		},
		{
			"$group": bson.M{
				"_id":   "$userid",
				"count": bson.M{"$sum": 1}, // 添加一个额外的字段以便后续过滤
			},
		},
		{
			"$match": bson.M{
				"count": bson.M{"$gte": 1, "$lte": 5}, // 过滤num值在1-5之间的文档
			},
		},
	}
	result := []bson.M{}
	LogGameTimes.Pipe(pipeline).All(&result)
	t_map := make(map[string]int, 0)
	for _, item := range result {
		uid := item["_id"].(string)
		ucount := item["count"].(int)
		t_map[uid] = ucount
	}
	cp_list := make([]entity.CPDetail, 0)       // 彩票数据
	crash_list := make([]entity.CRASHDetail, 0) // crash数据
	lhd_list := make([]entity.LHDetail, 0)      // 龙虎斗数据
	up_list := make([]entity.UPDetail, 0)       // 7up数据
	fj_list := make([]entity.CRASHDetail, 0)    // 飞机数据
	tp_list := make([]entity.TPDetail, 0)       // tp数据
	ak_list := make([]entity.AK47Detail, 0)     // ak47数据
	ab_list := make([]entity.ABDetail, 0)       // ab数据
	joker_list := make([]entity.JOKERDetail, 0) // joker数据
	rm_list := make([]entity.RMDetail, 0)       // rm数据
	// 查询玩家游戏结果
	for k, _ := range t_map {
		// 对局
		pipeline2 := []bson.M{
			{
				"$match": bson.M{
					"end_time": bson.M{"$gte": statrTimestamp, "$lte": endTimestamp},
					"players": bson.M{
						"$regex":   fmt.Sprintf("\\b%s\\b", k),
						"$options": "i",
					},
				},
			},
		}
		var result2 []entity.Detail
		Details.Pipe(pipeline2).All(&result2)
		// 如果查询出来大于五则不处理该数据
		if len(result2) <= 5 {
			for _, dt := range result2 {
				switch dt.Gtype {
				case 1:
					// TP
					if len(dt.TPDetail) > 0 {
						for _, v := range dt.TPDetail {
							// 只获取该用户得对局详情
							new_info := new(entity.TPDetail)
							new_info.WaterId = dt.WaterId
							new_info.UserId = v.UserId
							strtem := ""
							for _, c := range v.Cards {
								name := CardToStr(c)
								strtem += " " + name
							}
							new_info.CardStr = strtem
							new_info.FScore = Chip2Float(v.Score)
							new_info.FBeforeScore = Chip2Float(v.BeforeScore)
							new_info.FAfterScore = Chip2Float(v.AfterScore)
							for _, tn := range v.TPDetailTurn {
								num := strconv.Itoa(tn.Turn + 1)
								switch tn.Turn {
								case 0:
									new_info.TurnStr1 = "轮次" + num + ":" + tn.Operation
								case 1:
									new_info.TurnStr2 = "轮次" + num + ":" + tn.Operation
								case 2:
									new_info.TurnStr3 = "轮次" + num + ":" + tn.Operation
								case 3:
									new_info.TurnStr4 = "轮次" + num + ":" + tn.Operation
								}
							}
							tp_list = append(tp_list, *new_info)
						}
					}
				case 2:
					// LHD
					if dt.LHDetail != nil {
						for _, v := range dt.LHDetail.UserDetail {
							new_info := new(entity.LHDetail)
							new_info.WaterId = dt.WaterId
							new_info.Userid = v.Userid
							new_info.Winner = dt.LHDetail.Winner
							new_info.DragonValue = dt.LHDetail.DragonValue
							new_info.TigerValue = dt.LHDetail.TigerValue
							new_info.FBets = Chip2Float(dt.LHDetail.Bets)
							new_info.FPlayerWin = Chip2Float(dt.LHDetail.PlayerWin)
							new_info.FDragon = Chip2Float(v.Dragon)
							new_info.FTiger = Chip2Float(v.Tiger)
							new_info.FTie = Chip2Float(v.Tie)
							new_info.Result = v.Result
							new_info.FWin = Chip2Float(v.Win)
							new_info.FBeforeScore = Chip2Float(v.BeforeScore)
							new_info.FAfterScore = Chip2Float(v.AfterScore)
							lhd_list = append(lhd_list, *new_info)
						}
					}
				case 3:
					// 7UP
					if dt.UPDetail != nil {
						for _, v := range dt.UPDetail.UserDetail {
							new_info := new(entity.UPDetail)
							new_info.WaterId = dt.WaterId
							new_info.Userid = v.Userid
							new_info.Winner = dt.UPDetail.Winner
							new_info.PointValue = dt.UPDetail.PointValue
							new_info.FBets = Chip2Float(dt.UPDetail.Bets)
							new_info.FPlayerWin = Chip2Float(dt.UPDetail.PlayerWin)
							new_info.FDragon = Chip2Float(v.Dragon)
							new_info.FTiger = Chip2Float(v.Tiger)
							new_info.FTie = Chip2Float(v.Tie)
							new_info.Result = v.Result
							new_info.FWin = Chip2Float(v.Win)
							new_info.FBeforeScore = Chip2Float(v.BeforeScore)
							new_info.FAfterScore = Chip2Float(v.AfterScore)
							up_list = append(up_list, *new_info)
						}
					}
				case 4:
					// RUMMY
					if len(dt.RMDetail) > 0 {
						for _, v := range dt.RMDetail {
							new_info := new(entity.RMDetail)
							new_info.WaterId = dt.WaterId
							new_info.UserId = v.UserId
							finalcardsstr := ""
							for _, c := range v.FinalCards {
								for _, b := range c {
									name := CardToStr(b)
									finalcardsstr += " " + name
								}
							}

							initcardsstr := ""
							for _, c := range v.InitCards {
								name := CardToStr(c)
								initcardsstr += " " + name
							}
							wildcardstr := CardToStr(v.WildCard)
							new_info.FinalCardsStr = finalcardsstr
							new_info.InitCardsStr = initcardsstr
							new_info.WildCardStr = wildcardstr
							new_info.Result = v.Result
							new_info.FScore = Chip2Float(v.Score)
							new_info.FBeforeScore = Chip2Float(v.BeforeScore)
							new_info.FAfterScore = Chip2Float(v.AfterScore)
							rm_list = append(rm_list, *new_info)
						}
					}
				case 5:
					// AK47
					if len(dt.AK47Detail) > 0 {
						for _, v := range dt.AK47Detail {
							new_info := new(entity.AK47Detail)
							new_info.WaterId = dt.WaterId
							new_info.UserId = v.UserId
							cardstr := ""
							for _, c := range v.Cards {
								name := CardToStr(c)
								cardstr += " " + name
							}
							new_info.CardStr = cardstr
							new_info.FScore = Chip2Float(v.Score)
							new_info.FBeforeScore = Chip2Float(v.BeforeScore)
							new_info.FAfterScore = Chip2Float(v.AfterScore)
							for _, tn := range v.AK47DetailTurn {
								num := strconv.Itoa(tn.Turn + 1)
								switch tn.Turn {
								case 0:
									new_info.TurnStr1 = "轮次" + num + ":" + tn.Operation
								case 1:
									new_info.TurnStr2 = "轮次" + num + ":" + tn.Operation
								case 2:
									new_info.TurnStr3 = "轮次" + num + ":" + tn.Operation
								case 3:
									new_info.TurnStr4 = "轮次" + num + ":" + tn.Operation
								case 4:
									new_info.TurnStr5 = "轮次" + num + ":" + tn.Operation
								}
							}
							ak_list = append(ak_list, *new_info)
						}
					}
				case 6:
					// Joker
					if len(dt.JOKERDetail) > 0 {
						for _, v := range dt.JOKERDetail {
							new_info := new(entity.JOKERDetail)
							new_info.WaterId = dt.WaterId
							new_info.UserId = v.UserId
							cardstr := ""
							for _, c := range v.Cards {
								name := CardToStr(c)
								cardstr += " " + name
							}
							new_info.CardStr = cardstr
							new_info.FScore = Chip2Float(v.Score)
							new_info.FBeforeScore = Chip2Float(v.BeforeScore)
							new_info.FAfterScore = Chip2Float(v.AfterScore)
							for _, tn := range v.JOKERDetailTurn {
								num := strconv.Itoa(tn.Turn + 1)
								switch tn.Turn {
								case 0:
									new_info.TurnStr1 = "轮次" + num + ":" + tn.Operation
								case 1:
									new_info.TurnStr2 = "轮次" + num + ":" + tn.Operation
								case 2:
									new_info.TurnStr3 = "轮次" + num + ":" + tn.Operation
								case 3:
									new_info.TurnStr4 = "轮次" + num + ":" + tn.Operation
								}
							}
							joker_list = append(joker_list, *new_info)
						}
					}
				case 7:
					// Crash
					if dt.CRASHDetail != nil {
						for _, v := range dt.CRASHDetail.UserDetail {
							new_info := new(entity.CRASHDetail)
							new_info.WaterId = dt.WaterId
							new_info.Userid = v.Userid
							new_info.PlayerLose = dt.CRASHDetail.PlayerLose
							new_info.FBets = Chip2Float(dt.CRASHDetail.Bets)
							new_info.FPlayerWin = Chip2Float(dt.CRASHDetail.PlayerLose)
							new_info.Multiple = v.Multiple
							new_info.DResult = v.Result
							new_info.FBet = Chip2Float(v.Bet)
							new_info.Result = dt.CRASHDetail.Result
							new_info.FWin = Chip2Float(v.Win)
							new_info.FBeforeScore = Chip2Float(v.BeforeScore)
							new_info.FAfterScore = Chip2Float(v.AfterScore)
							new_info.FPlayerFactor = Chip2Float(int64(dt.PlayerFactor))
							crash_list = append(crash_list, *new_info)
						}
					}
				case 8:
					// AB
					if dt.ABDetail != nil {
						for _, v := range dt.ABDetail.UserDetail {
							new_info := new(entity.ABDetail)
							new_info.WaterId = dt.WaterId
							new_info.Userid = v.Userid
							acradstr := ""
							for _, c := range dt.ABDetail.ACards {
								name := CardToStr(c)
								acradstr += " " + name
							}
							bcradstr := ""
							for _, c := range dt.ABDetail.BCards {
								name := CardToStr(c)
								bcradstr += " " + name
							}
							new_info.JokerStr = CardToStr(dt.ABDetail.Joker) // dt.ABDetail.Joker
							new_info.JackpotValueStr = CardToStr(dt.ABDetail.JackpotValue)
							new_info.ANDAR = acradstr
							new_info.BAHAR = bcradstr
							if dt.ABDetail.Winner == 1 {
								new_info.WinnerStr = "ANDAR"
							} else if dt.ABDetail.Winner == 2 {
								new_info.WinnerStr = "BAHAR"
							}
							if dt.ABDetail.SideWinner == 3 {
								new_info.SideWinnerStr = "1-5"
							} else if dt.ABDetail.SideWinner == 4 {
								new_info.SideWinnerStr = "6-10"
							} else if dt.ABDetail.SideWinner == 5 {
								new_info.SideWinnerStr = "11-15"
							} else if dt.ABDetail.SideWinner == 6 {
								new_info.SideWinnerStr = "16-25"
							} else if dt.ABDetail.SideWinner == 7 {
								new_info.SideWinnerStr = "26-30"
							} else if dt.ABDetail.SideWinner == 8 {
								new_info.SideWinnerStr = "31-35"
							} else if dt.ABDetail.SideWinner == 9 {
								new_info.SideWinnerStr = "36-40"
							} else if dt.ABDetail.SideWinner == 10 {
								new_info.SideWinnerStr = "41以上"
							}
							new_info.FBets = Chip2Float(dt.ABDetail.Bets)
							new_info.FPlayerWin = Chip2Float(dt.ABDetail.PlayerWin)
							new_info.Result = v.Result
							new_info.FWin = Chip2Float(v.Win)
							new_info.FBeforeScore = Chip2Float(v.BeforeScore)
							new_info.FAfterScore = Chip2Float(v.AfterScore)
							for i, b := range v.SeatBets {
								switch i {
								case "1":
									new_info.SeatBets1 = Chip2Float(int64(b))
								case "2":
									new_info.SeatBets2 = Chip2Float(int64(b))
								case "3":
									new_info.SeatBets3 = Chip2Float(int64(b))
								case "4":
									new_info.SeatBets4 = Chip2Float(int64(b))
								case "5":
									new_info.SeatBets5 = Chip2Float(int64(b))
								case "6":
									new_info.SeatBets6 = Chip2Float(int64(b))
								case "7":
									new_info.SeatBets7 = Chip2Float(int64(b))
								case "8":
									new_info.SeatBets8 = Chip2Float(int64(b))
								case "9":
									new_info.SeatBets9 = Chip2Float(int64(b))
								case "10":
									new_info.SeatBets10 = Chip2Float(int64(b))
								}
							}
							ab_list = append(ab_list, *new_info)
						}
					}
				case 9:
					// 彩票
					if dt.CPDetail != nil {
						for _, v := range dt.CPDetail.UserDetail {
							new_info := new(entity.CPDetail)
							new_info.WaterId = dt.WaterId
							new_info.Userid = v.Userid
							cardstr := ""
							for _, c := range dt.CPDetail.Cards {
								name := CardToStr(c)
								cardstr += " " + name
							}
							new_info.CardStr = cardstr
							new_info.FBets = Chip2Float(dt.CPDetail.Bets)
							new_info.FPlayerWin = Chip2Float(dt.CPDetail.PlayerWin)
							for i, b := range v.SeatBets {
								switch i {
								case "1":
									new_info.SeatBets1 = Chip2Float(int64(b))
								case "2":
									new_info.SeatBets2 = Chip2Float(int64(b))
								case "3":
									new_info.SeatBets3 = Chip2Float(int64(b))
								case "4":
									new_info.SeatBets4 = Chip2Float(int64(b))
								case "5":
									new_info.SeatBets5 = Chip2Float(int64(b))
								case "6":
									new_info.SeatBets6 = Chip2Float(int64(b))
								}
							}

							new_info.Result = v.Result
							new_info.FWin = Chip2Float(v.Win)
							new_info.FBeforeScore = Chip2Float(v.BeforeScore)
							new_info.FAfterScore = Chip2Float(v.AfterScore)
							cp_list = append(cp_list, *new_info)
						}
					}
				case 10:
					// 飞机
					if dt.CRASHDetail != nil {
						for _, v := range dt.CRASHDetail.UserDetail {
							new_info := new(entity.CRASHDetail)
							new_info.WaterId = dt.WaterId
							new_info.Userid = v.Userid
							new_info.PlayerLose = dt.CRASHDetail.PlayerLose
							new_info.FBets = Chip2Float(dt.CRASHDetail.Bets)
							new_info.FPlayerWin = Chip2Float(dt.CRASHDetail.PlayerLose)
							new_info.Multiple = v.Multiple
							new_info.DResult = v.Result
							new_info.FBet = Chip2Float(v.Bet)
							new_info.Result = dt.CRASHDetail.Result
							new_info.FWin = Chip2Float(v.Win)
							new_info.FBeforeScore = Chip2Float(v.BeforeScore)
							new_info.FAfterScore = Chip2Float(v.AfterScore)
							new_info.FPlayerFactor = Chip2Float(int64(dt.PlayerFactor))
							fj_list = append(fj_list, *new_info)
						}
					}
				}
			}
		}
	}

	// TP
	if len(tp_list) > 0 {
		beego.Info("UserGameDataByB1_5 开始执行TP数据导出")
		// 导出
		f := excelize.NewFile()
		sheet1 := "Sheet1"
		// 设置表头
		f.SetCellValue(sheet1, "A1", "用户id")
		f.SetCellValue(sheet1, "B1", "局号")
		f.SetCellValue(sheet1, "C1", "牌型")
		f.SetCellValue(sheet1, "D1", "结算")
		f.SetCellValue(sheet1, "E1", "账变前分数")
		f.SetCellValue(sheet1, "F1", "账变后分数")
		f.SetCellValue(sheet1, "G1", "轮次1")
		f.SetCellValue(sheet1, "H1", "轮次2")
		f.SetCellValue(sheet1, "I1", "轮次3")
		f.SetCellValue(sheet1, "J1", "轮次4")
		line := 1
		for _, player := range tp_list {
			line++
			f.SetCellValue(sheet1, fmt.Sprintf("A%d", line), player.UserId)
			f.SetCellValue(sheet1, fmt.Sprintf("B%d", line), player.WaterId)
			f.SetCellValue(sheet1, fmt.Sprintf("C%d", line), player.CardStr)
			f.SetCellValue(sheet1, fmt.Sprintf("D%d", line), player.FScore)
			f.SetCellValue(sheet1, fmt.Sprintf("E%d", line), player.FBeforeScore)
			f.SetCellValue(sheet1, fmt.Sprintf("F%d", line), player.FAfterScore)
			f.SetCellValue(sheet1, fmt.Sprintf("G%d", line), player.TurnStr1)
			f.SetCellValue(sheet1, fmt.Sprintf("H%d", line), player.TurnStr2)
			f.SetCellValue(sheet1, fmt.Sprintf("I%d", line), player.TurnStr3)
			f.SetCellValue(sheet1, fmt.Sprintf("J%d", line), player.TurnStr4)
		}
		now := time.Now().Format("2006-01-02.15.04.05")
		filename := begin + "B类用户TP游戏" + "_" + fmt.Sprint(now) + ".xlsx"
		// 保存文件
		if err := f.SaveAs(filename); err != nil {
			glog.Error("error3: ", err)
		}
	}

	// 龙虎斗
	if len(lhd_list) > 0 {
		beego.Info("UserGameDataByB1_5 开始执行LHD数据导出")
		// 导出
		f := excelize.NewFile()
		sheet1 := "Sheet1"
		// 设置表头
		f.SetCellValue(sheet1, "A1", "局号")
		f.SetCellValue(sheet1, "B1", "开奖结果")
		f.SetCellValue(sheet1, "C1", "龙")
		f.SetCellValue(sheet1, "D1", "虎")
		f.SetCellValue(sheet1, "E1", "总下注")
		f.SetCellValue(sheet1, "F1", "玩家赢分")
		f.SetCellValue(sheet1, "G1", "下注用户ID")
		f.SetCellValue(sheet1, "H1", "龙")
		f.SetCellValue(sheet1, "I1", "虎")
		f.SetCellValue(sheet1, "J1", "和")
		f.SetCellValue(sheet1, "K1", "结果")
		f.SetCellValue(sheet1, "L1", "结算")
		f.SetCellValue(sheet1, "M1", "账变前分数")
		f.SetCellValue(sheet1, "N1", "账变前分数")
		line := 1
		for _, player := range lhd_list {
			line++
			f.SetCellValue(sheet1, fmt.Sprintf("A%d", line), player.WaterId)
			f.SetCellValue(sheet1, fmt.Sprintf("B%d", line), player.Winner)
			f.SetCellValue(sheet1, fmt.Sprintf("C%d", line), player.DragonValue)
			f.SetCellValue(sheet1, fmt.Sprintf("D%d", line), player.TigerValue)
			f.SetCellValue(sheet1, fmt.Sprintf("E%d", line), player.FBets)
			f.SetCellValue(sheet1, fmt.Sprintf("F%d", line), player.FPlayerWin)
			f.SetCellValue(sheet1, fmt.Sprintf("G%d", line), player.Userid)
			f.SetCellValue(sheet1, fmt.Sprintf("H%d", line), player.FDragon)
			f.SetCellValue(sheet1, fmt.Sprintf("I%d", line), player.FTiger)
			f.SetCellValue(sheet1, fmt.Sprintf("J%d", line), player.FTie)
			f.SetCellValue(sheet1, fmt.Sprintf("K%d", line), player.FWin)
			f.SetCellValue(sheet1, fmt.Sprintf("L%d", line), player.Result)
			f.SetCellValue(sheet1, fmt.Sprintf("M%d", line), player.FBeforeScore)
			f.SetCellValue(sheet1, fmt.Sprintf("N%d", line), player.FAfterScore)
		}
		now := time.Now().Format("2006-01-02.15.04.05")
		filename := begin + "B类用户LHD游戏" + "_" + fmt.Sprint(now) + ".xlsx"
		// 保存文件
		if err := f.SaveAs(filename); err != nil {
			glog.Error("error3: ", err)
		}
	}

	// 7updown
	if len(up_list) > 0 {
		beego.Info("UserGameDataByB1_5 开始执行7updown数据导出")
		// 导出
		f := excelize.NewFile()
		sheet1 := "Sheet1"
		// 设置表头
		f.SetCellValue(sheet1, "A1", "用户id")
		f.SetCellValue(sheet1, "B1", "局号")
		f.SetCellValue(sheet1, "C1", "玩家最终系数")
		f.SetCellValue(sheet1, "D1", "开奖结果")
		f.SetCellValue(sheet1, "E1", "点数")
		f.SetCellValue(sheet1, "F1", "总下注")
		f.SetCellValue(sheet1, "G1", "玩家赢分")
		f.SetCellValue(sheet1, "H1", "大")
		f.SetCellValue(sheet1, "I1", "小")
		f.SetCellValue(sheet1, "J1", "7")
		f.SetCellValue(sheet1, "K1", "结果")
		f.SetCellValue(sheet1, "L1", "结算")
		f.SetCellValue(sheet1, "M1", "账变前分数")
		f.SetCellValue(sheet1, "N1", "账变后分数")
		line := 1
		for _, player := range up_list {
			line++
			f.SetCellValue(sheet1, fmt.Sprintf("A%d", line), player.Userid)
			f.SetCellValue(sheet1, fmt.Sprintf("B%d", line), player.WaterId)
			f.SetCellValue(sheet1, fmt.Sprintf("C%d", line), player.FPlayerFactor)
			f.SetCellValue(sheet1, fmt.Sprintf("D%d", line), player.Winner)
			f.SetCellValue(sheet1, fmt.Sprintf("E%d", line), player.PointValue)
			f.SetCellValue(sheet1, fmt.Sprintf("F%d", line), player.FBets)
			f.SetCellValue(sheet1, fmt.Sprintf("G%d", line), player.FPlayerWin)
			f.SetCellValue(sheet1, fmt.Sprintf("H%d", line), player.FDragon)
			f.SetCellValue(sheet1, fmt.Sprintf("I%d", line), player.FTiger)
			f.SetCellValue(sheet1, fmt.Sprintf("J%d", line), player.FTie)
			f.SetCellValue(sheet1, fmt.Sprintf("K%d", line), player.Result)
			f.SetCellValue(sheet1, fmt.Sprintf("L%d", line), player.FWin)
			f.SetCellValue(sheet1, fmt.Sprintf("M%d", line), player.FBeforeScore)
			f.SetCellValue(sheet1, fmt.Sprintf("N%d", line), player.FAfterScore)

		}
		now := time.Now().Format("2006-01-02.15.04.05")
		filename := begin + "B类用户7updown游戏" + "_" + fmt.Sprint(now) + ".xlsx"
		// 保存文件
		if err := f.SaveAs(filename); err != nil {
			glog.Error("error3: ", err)
		}
	}

	// Rummy
	if len(rm_list) > 0 {
		beego.Info("UserGameDataByB1_5 开始执行Rummy数据导出")
		// 导出
		f := excelize.NewFile()
		sheet1 := "Sheet1"
		// 设置表头
		f.SetCellValue(sheet1, "A1", "用户id")
		f.SetCellValue(sheet1, "B1", "局号")
		f.SetCellValue(sheet1, "C1", "结果")
		f.SetCellValue(sheet1, "D1", "结算")
		f.SetCellValue(sheet1, "E1", "账变前分数")
		f.SetCellValue(sheet1, "F1", "账变后分数")
		f.SetCellValue(sheet1, "G1", "最终牌型")
		f.SetCellValue(sheet1, "H1", "最初牌型")
		f.SetCellValue(sheet1, "I1", "万能牌")
		line := 1
		for _, player := range rm_list {
			line++
			f.SetCellValue(sheet1, fmt.Sprintf("A%d", line), player.UserId)
			f.SetCellValue(sheet1, fmt.Sprintf("B%d", line), player.WaterId)
			f.SetCellValue(sheet1, fmt.Sprintf("C%d", line), player.Result)
			f.SetCellValue(sheet1, fmt.Sprintf("D%d", line), player.FScore)
			f.SetCellValue(sheet1, fmt.Sprintf("E%d", line), player.FBeforeScore)
			f.SetCellValue(sheet1, fmt.Sprintf("F%d", line), player.FAfterScore)
			f.SetCellValue(sheet1, fmt.Sprintf("G%d", line), player.FinalCardsStr)
			f.SetCellValue(sheet1, fmt.Sprintf("H%d", line), player.InitCardsStr)
			f.SetCellValue(sheet1, fmt.Sprintf("I%d", line), player.WildCardStr)

		}
		now := time.Now().Format("2006-01-02.15.04.05")
		filename := begin + "B类用户Rummy游戏" + "_" + fmt.Sprint(now) + ".xlsx"
		// 保存文件
		if err := f.SaveAs(filename); err != nil {
			glog.Error("error3: ", err)
		}
	}

	// AK47
	if len(ak_list) > 0 {
		beego.Info("UserGameDataByB1_5 开始执行AK47数据导出")
		// 导出
		f := excelize.NewFile()
		sheet1 := "Sheet1"
		// 设置表头
		f.SetCellValue(sheet1, "A1", "用户id")
		f.SetCellValue(sheet1, "B1", "局号")
		f.SetCellValue(sheet1, "C1", "牌型")
		f.SetCellValue(sheet1, "D1", "结算")
		f.SetCellValue(sheet1, "E1", "账变前分数")
		f.SetCellValue(sheet1, "F1", "账变后分数")
		f.SetCellValue(sheet1, "G1", "轮次1")
		f.SetCellValue(sheet1, "H1", "轮次2")
		f.SetCellValue(sheet1, "I1", "轮次3")
		f.SetCellValue(sheet1, "J1", "轮次4")
		f.SetCellValue(sheet1, "K1", "轮次5")
		line := 1
		for _, player := range ak_list {
			line++
			f.SetCellValue(sheet1, fmt.Sprintf("A%d", line), player.UserId)
			f.SetCellValue(sheet1, fmt.Sprintf("B%d", line), player.WaterId)
			f.SetCellValue(sheet1, fmt.Sprintf("C%d", line), player.CardStr)
			f.SetCellValue(sheet1, fmt.Sprintf("D%d", line), player.FScore)
			f.SetCellValue(sheet1, fmt.Sprintf("E%d", line), player.FBeforeScore)
			f.SetCellValue(sheet1, fmt.Sprintf("F%d", line), player.FAfterScore)
			f.SetCellValue(sheet1, fmt.Sprintf("G%d", line), player.TurnStr1)
			f.SetCellValue(sheet1, fmt.Sprintf("H%d", line), player.TurnStr2)
			f.SetCellValue(sheet1, fmt.Sprintf("I%d", line), player.TurnStr3)
			f.SetCellValue(sheet1, fmt.Sprintf("J%d", line), player.TurnStr4)
			f.SetCellValue(sheet1, fmt.Sprintf("K%d", line), player.TurnStr5)

		}
		now := time.Now().Format("2006-01-02.15.04.05")
		filename := begin + "B类用户AK47游戏" + "_" + fmt.Sprint(now) + ".xlsx"
		// 保存文件
		if err := f.SaveAs(filename); err != nil {
			glog.Error("error3: ", err)
		}
	}

	// joker
	if len(joker_list) > 0 {
		beego.Info("UserGameDataByB1_5 开始执行Joker数据导出")
		// 导出
		f := excelize.NewFile()
		sheet1 := "Sheet1"
		// 设置表头
		f.SetCellValue(sheet1, "A1", "用户id")
		f.SetCellValue(sheet1, "B1", "局号")
		f.SetCellValue(sheet1, "C1", "牌型")
		f.SetCellValue(sheet1, "D1", "结算")
		f.SetCellValue(sheet1, "E1", "账变前分数")
		f.SetCellValue(sheet1, "F1", "账变后分数")
		f.SetCellValue(sheet1, "G1", "轮次1")
		f.SetCellValue(sheet1, "H1", "轮次2")
		f.SetCellValue(sheet1, "I1", "轮次3")
		f.SetCellValue(sheet1, "J1", "轮次4")
		line := 1
		for _, player := range joker_list {
			line++
			f.SetCellValue(sheet1, fmt.Sprintf("A%d", line), player.UserId)
			f.SetCellValue(sheet1, fmt.Sprintf("B%d", line), player.WaterId)
			f.SetCellValue(sheet1, fmt.Sprintf("C%d", line), player.CardStr)
			f.SetCellValue(sheet1, fmt.Sprintf("D%d", line), player.FScore)
			f.SetCellValue(sheet1, fmt.Sprintf("E%d", line), player.FBeforeScore)
			f.SetCellValue(sheet1, fmt.Sprintf("F%d", line), player.FAfterScore)
			f.SetCellValue(sheet1, fmt.Sprintf("G%d", line), player.TurnStr1)
			f.SetCellValue(sheet1, fmt.Sprintf("H%d", line), player.TurnStr2)
			f.SetCellValue(sheet1, fmt.Sprintf("I%d", line), player.TurnStr3)
			f.SetCellValue(sheet1, fmt.Sprintf("J%d", line), player.TurnStr4)

		}
		now := time.Now().Format("2006-01-02.15.04.05")
		filename := begin + "B类用户Joker游戏" + "_" + fmt.Sprint(now) + ".xlsx"
		// 保存文件
		if err := f.SaveAs(filename); err != nil {
			glog.Error("error3: ", err)
		}
	}

	// Crash
	if len(crash_list) > 0 {
		beego.Info("UserGameDataByB1_5 开始执行Crash数据导出")
		// 导出
		f := excelize.NewFile()
		sheet1 := "Sheet1"
		// 设置表头
		f.SetCellValue(sheet1, "A1", "用户id")
		f.SetCellValue(sheet1, "B1", "局号")
		f.SetCellValue(sheet1, "C1", "玩家最终系数")
		f.SetCellValue(sheet1, "D1", "开奖结果")
		f.SetCellValue(sheet1, "E1", "总下注")
		f.SetCellValue(sheet1, "F1", "玩家赢分")
		f.SetCellValue(sheet1, "G1", "分数")
		f.SetCellValue(sheet1, "H1", "逃脱倍数")
		f.SetCellValue(sheet1, "I1", "结果")
		f.SetCellValue(sheet1, "J1", "结算")
		f.SetCellValue(sheet1, "K1", "账变前分数")
		f.SetCellValue(sheet1, "L1", "账变后分数")
		line := 1
		for _, player := range crash_list {
			line++
			f.SetCellValue(sheet1, fmt.Sprintf("A%d", line), player.Userid)
			f.SetCellValue(sheet1, fmt.Sprintf("B%d", line), player.WaterId)
			f.SetCellValue(sheet1, fmt.Sprintf("C%d", line), player.FPlayerFactor)
			f.SetCellValue(sheet1, fmt.Sprintf("D%d", line), player.Result)
			f.SetCellValue(sheet1, fmt.Sprintf("E%d", line), player.FBets)
			f.SetCellValue(sheet1, fmt.Sprintf("F%d", line), player.FPlayerWin)
			f.SetCellValue(sheet1, fmt.Sprintf("G%d", line), player.FBet)
			f.SetCellValue(sheet1, fmt.Sprintf("H%d", line), player.Multiple)
			f.SetCellValue(sheet1, fmt.Sprintf("I%d", line), player.DResult)
			f.SetCellValue(sheet1, fmt.Sprintf("J%d", line), player.FWin)
			f.SetCellValue(sheet1, fmt.Sprintf("K%d", line), player.FBeforeScore)
			f.SetCellValue(sheet1, fmt.Sprintf("L%d", line), player.FAfterScore)
		}
		now := time.Now().Format("2006-01-02.15.04.05")
		filename := begin + "B类用户Crash游戏" + "_" + fmt.Sprint(now) + ".xlsx"
		// 保存文件
		if err := f.SaveAs(filename); err != nil {
			glog.Error("error3: ", err)
		}
	}

	// AB
	if len(ab_list) > 0 {
		beego.Info("UserGameDataByB1_5 开始执行AB数据导出")
		// 导出
		f := excelize.NewFile()
		sheet1 := "Sheet1"
		// 设置表头
		f.SetCellValue(sheet1, "A1", "用户id")
		f.SetCellValue(sheet1, "B1", "局号")
		f.SetCellValue(sheet1, "C1", "key牌")
		f.SetCellValue(sheet1, "D1", "中奖牌")
		f.SetCellValue(sheet1, "E1", "开奖结果1")
		f.SetCellValue(sheet1, "F1", "开奖结果2")
		f.SetCellValue(sheet1, "G1", "总下注")
		f.SetCellValue(sheet1, "H1", "玩家赢分")
		f.SetCellValue(sheet1, "I1", "ANDAR")
		f.SetCellValue(sheet1, "J1", "BAHAR")
		f.SetCellValue(sheet1, "K1", "ANDAR")
		f.SetCellValue(sheet1, "L1", "BAHAR")
		f.SetCellValue(sheet1, "M1", "1-5")
		f.SetCellValue(sheet1, "N1", "6-10")
		f.SetCellValue(sheet1, "O1", "11-15")
		f.SetCellValue(sheet1, "P1", "16-25")
		f.SetCellValue(sheet1, "Q1", "26-30")
		f.SetCellValue(sheet1, "R1", "31-35")
		f.SetCellValue(sheet1, "S1", "36-40")
		f.SetCellValue(sheet1, "T1", "41以上")
		f.SetCellValue(sheet1, "U1", "结果")
		f.SetCellValue(sheet1, "V1", "结算")
		f.SetCellValue(sheet1, "W1", "账变前分数")
		f.SetCellValue(sheet1, "X1", "账变后分数")
		line := 1
		for _, player := range ab_list {
			line++
			f.SetCellValue(sheet1, fmt.Sprintf("A%d", line), player.Userid)
			f.SetCellValue(sheet1, fmt.Sprintf("B%d", line), player.WaterId)
			f.SetCellValue(sheet1, fmt.Sprintf("C%d", line), player.JokerStr)
			f.SetCellValue(sheet1, fmt.Sprintf("D%d", line), player.JackpotValueStr)
			f.SetCellValue(sheet1, fmt.Sprintf("E%d", line), player.WinnerStr)
			f.SetCellValue(sheet1, fmt.Sprintf("F%d", line), player.SideWinnerStr)
			f.SetCellValue(sheet1, fmt.Sprintf("G%d", line), player.FBets)
			f.SetCellValue(sheet1, fmt.Sprintf("H%d", line), player.FPlayerWin)
			f.SetCellValue(sheet1, fmt.Sprintf("I%d", line), player.ANDAR)
			f.SetCellValue(sheet1, fmt.Sprintf("J%d", line), player.BAHAR)
			f.SetCellValue(sheet1, fmt.Sprintf("K%d", line), player.SeatBets1)
			f.SetCellValue(sheet1, fmt.Sprintf("L%d", line), player.SeatBets2)
			f.SetCellValue(sheet1, fmt.Sprintf("M%d", line), player.SeatBets3)
			f.SetCellValue(sheet1, fmt.Sprintf("N%d", line), player.SeatBets4)
			f.SetCellValue(sheet1, fmt.Sprintf("O%d", line), player.SeatBets5)
			f.SetCellValue(sheet1, fmt.Sprintf("P%d", line), player.SeatBets6)
			f.SetCellValue(sheet1, fmt.Sprintf("Q%d", line), player.SeatBets7)
			f.SetCellValue(sheet1, fmt.Sprintf("R%d", line), player.SeatBets8)
			f.SetCellValue(sheet1, fmt.Sprintf("S%d", line), player.SeatBets9)
			f.SetCellValue(sheet1, fmt.Sprintf("T%d", line), player.SeatBets10)
			f.SetCellValue(sheet1, fmt.Sprintf("U%d", line), player.Result)
			f.SetCellValue(sheet1, fmt.Sprintf("V%d", line), player.FWin)
			f.SetCellValue(sheet1, fmt.Sprintf("W%d", line), player.FBeforeScore)
			f.SetCellValue(sheet1, fmt.Sprintf("X%d", line), player.FAfterScore)
		}
		now := time.Now().Format("2006-01-02.15.04.05")
		filename := begin + "B类用户AB游戏" + "_" + fmt.Sprint(now) + ".xlsx"
		// 保存文件
		if err := f.SaveAs(filename); err != nil {
			glog.Error("error3: ", err)
		}
	}

	// cp
	if len(cp_list) > 0 {
		beego.Info("UserGameDataByB1_5 开始执行彩票数据导出")
		// 导出
		f := excelize.NewFile()
		sheet1 := "Sheet1"
		// 设置表头
		f.SetCellValue(sheet1, "A1", "用户id")
		f.SetCellValue(sheet1, "B1", "局号")
		f.SetCellValue(sheet1, "C1", "开奖结果")
		f.SetCellValue(sheet1, "D1", "牌型")
		f.SetCellValue(sheet1, "E1", "总下注")
		f.SetCellValue(sheet1, "F1", "玩家赢分")
		f.SetCellValue(sheet1, "G1", "高牌")
		f.SetCellValue(sheet1, "H1", "对子")
		f.SetCellValue(sheet1, "I1", "同花")
		f.SetCellValue(sheet1, "J1", "顺子")
		f.SetCellValue(sheet1, "K1", "同花顺")
		f.SetCellValue(sheet1, "L1", "豹子")
		f.SetCellValue(sheet1, "M1", "结果")
		f.SetCellValue(sheet1, "N1", "结算")
		f.SetCellValue(sheet1, "O1", "账变前分数")
		f.SetCellValue(sheet1, "P1", "账变后分数")
		line := 1
		for _, player := range cp_list {
			line++
			f.SetCellValue(sheet1, fmt.Sprintf("A%d", line), player.Userid)
			f.SetCellValue(sheet1, fmt.Sprintf("B%d", line), player.WaterId)
			f.SetCellValue(sheet1, fmt.Sprintf("C%d", line), player.JackpotOutput)
			f.SetCellValue(sheet1, fmt.Sprintf("D%d", line), player.CardStr)
			f.SetCellValue(sheet1, fmt.Sprintf("E%d", line), player.FBets)
			f.SetCellValue(sheet1, fmt.Sprintf("F%d", line), player.FPlayerWin)
			f.SetCellValue(sheet1, fmt.Sprintf("G%d", line), player.SeatBets1)
			f.SetCellValue(sheet1, fmt.Sprintf("H%d", line), player.SeatBets2)
			f.SetCellValue(sheet1, fmt.Sprintf("I%d", line), player.SeatBets3)
			f.SetCellValue(sheet1, fmt.Sprintf("J%d", line), player.SeatBets4)
			f.SetCellValue(sheet1, fmt.Sprintf("K%d", line), player.SeatBets5)
			f.SetCellValue(sheet1, fmt.Sprintf("L%d", line), player.SeatBets6)
			f.SetCellValue(sheet1, fmt.Sprintf("M%d", line), player.Result)
			f.SetCellValue(sheet1, fmt.Sprintf("N%d", line), player.FWin)
			f.SetCellValue(sheet1, fmt.Sprintf("O%d", line), player.FBeforeScore)
			f.SetCellValue(sheet1, fmt.Sprintf("P%d", line), player.FAfterScore)
		}
		now := time.Now().Format("2006-01-02.15.04.05")
		filename := begin + "B类用户彩票游戏" + "_" + fmt.Sprint(now) + ".xlsx"
		// 保存文件
		if err := f.SaveAs(filename); err != nil {
			glog.Error("error3: ", err)
		}
	}

	// 飞机
	if len(fj_list) > 0 {
		beego.Info("UserGameDataByB1_5 开始执行飞机数据导出")
		// 导出
		f := excelize.NewFile()
		sheet1 := "Sheet1"
		// 设置表头
		f.SetCellValue(sheet1, "A1", "用户id")
		f.SetCellValue(sheet1, "B1", "局号")
		f.SetCellValue(sheet1, "C1", "玩家最终系数")
		f.SetCellValue(sheet1, "D1", "开奖结果")
		f.SetCellValue(sheet1, "E1", "总下注")
		f.SetCellValue(sheet1, "F1", "玩家赢分")
		f.SetCellValue(sheet1, "G1", "分数")
		f.SetCellValue(sheet1, "H1", "逃脱倍数")
		f.SetCellValue(sheet1, "I1", "结果")
		f.SetCellValue(sheet1, "J1", "结算")
		f.SetCellValue(sheet1, "K1", "账变前分数")
		f.SetCellValue(sheet1, "L1", "账变后分数")
		line := 1
		for _, player := range fj_list {
			line++
			f.SetCellValue(sheet1, fmt.Sprintf("A%d", line), player.Userid)
			f.SetCellValue(sheet1, fmt.Sprintf("B%d", line), player.WaterId)
			f.SetCellValue(sheet1, fmt.Sprintf("C%d", line), player.FPlayerFactor)
			f.SetCellValue(sheet1, fmt.Sprintf("D%d", line), player.Result)
			f.SetCellValue(sheet1, fmt.Sprintf("E%d", line), player.FBets)
			f.SetCellValue(sheet1, fmt.Sprintf("F%d", line), player.FPlayerWin)
			f.SetCellValue(sheet1, fmt.Sprintf("G%d", line), player.FBet)
			f.SetCellValue(sheet1, fmt.Sprintf("H%d", line), player.Multiple)
			f.SetCellValue(sheet1, fmt.Sprintf("I%d", line), player.DResult)
			f.SetCellValue(sheet1, fmt.Sprintf("J%d", line), player.FWin)
			f.SetCellValue(sheet1, fmt.Sprintf("K%d", line), player.FBeforeScore)
			f.SetCellValue(sheet1, fmt.Sprintf("L%d", line), player.FAfterScore)
		}
		now := time.Now().Format("2006-01-02.15.04.05")
		filename := begin + "B类用户飞机游戏" + "_" + fmt.Sprint(now) + ".xlsx"
		// 保存文件
		if err := f.SaveAs(filename); err != nil {
			glog.Error("error3: ", err)
		}
	}
}

func CardToStr(key uint32) string {
	name := ""
	cardlist := map[uint32]string{
		17: "方块A",
		18: "方块2",
		19: "方块3",
		20: "方块4",
		21: "方块5",
		22: "方块6",
		23: "方块7",
		24: "方块8",
		25: "方块9",
		26: "方块10",
		27: "方块J",
		28: "方块Q",
		29: "方块K",
		33: "梅花A",
		34: "梅花2",
		35: "梅花3",
		36: "梅花4",
		37: "梅花5",
		38: "梅花6",
		39: "梅花7",
		40: "梅花8",
		41: "梅花9",
		42: "梅花10",
		43: "梅花J",
		44: "梅花Q",
		45: "梅花K",
		49: "红桃A",
		50: "红桃2",
		51: "红桃3",
		52: "红桃4",
		53: "红桃5",
		54: "红桃6",
		55: "红桃7",
		56: "红桃8",
		57: "红桃9",
		58: "红桃10",
		59: "红桃J",
		60: "红桃Q",
		61: "红桃K",
		65: "黑桃A",
		66: "黑桃2",
		67: "黑桃3",
		68: "黑桃4",
		69: "黑桃5",
		70: "黑桃6",
		71: "黑桃7",
		72: "黑桃8",
		73: "黑桃9",
		74: "黑桃10",
		75: "黑桃J",
		76: "黑桃Q",
		77: "黑桃K",
		81: "小王",
		82: "大王",
	}
	if v, ok := cardlist[key]; ok {
		name = v
	}
	return name
}
func CrashStats_old(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
	date1, _ := utils.Unix(today)

	m1 := bson.M{
		"gtype":    int(pb.CRASH),
		"end_time": bson.M{"$gt": startTime.Unix(), "$lt": endTime.Unix()},
		// "players":    bson.M{"$ne": ""},
	}

	var r1 []entity.Detail
	err := Details.Find(m1).All(&r1)
	if err != nil {
		beego.Warning("CrashStats error1:", err)
		return
	}
	if len(r1) == 0 {
		beego.Warning("no crash detail data")
		return
	}

	// 查询当天注册用户
	m2 := []bson.M{
		{"$match": bson.M{
			"ctime": bson.M{"$gt": startTime, "$lt": endTime},
		}},
		{"$project": bson.M{"_id": "$_id"}},
	}
	var r2 []bson.M
	err = PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Warning("CrashStats error2:", err)
		return
	}
	var regUserids = make(map[string]bool)
	for _, r := range r2 {
		userid := r["_id"].(string)
		regUserids[userid] = true
	}

	// crash玩家列表
	var crashPlayerIds []string
	var crashPlayers = make(map[string]*entity.CrashPlayerStat)

	for _, detail := range r1 {
		if detail.CRASHDetail == nil {
			continue
		}
		// mulpitle := fmt.Sprintf("x%.2f", float32(detail.CRASHDetail.Result)/100)
		var mulpitle float64
		_, err = fmt.Sscanf(detail.CRASHDetail.Result, "x%f", &mulpitle)
		if err != nil {
			beego.Error(fmt.Scanf("crash mulpitle parse error: %s", detail.CRASHDetail.Result))
		}

		for _, user := range detail.CRASHDetail.UserDetail {
			if len(user.Userid) >= 16 {
				// 人机id18位长度+
				continue
			}
			stat, ok := crashPlayers[user.Userid]
			if !ok {
				stat = &entity.CrashPlayerStat{
					Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
					Date:    date1,
					DateStr: today,
					Userid:  user.Userid,
					NewReg:  regUserids[user.Userid], // 老顾客新顾客
				}
				crashPlayers[user.Userid] = stat
				crashPlayerIds = append(crashPlayerIds, user.Userid)
			}
			stat.AllRounds++
			stat.GameTimes += (detail.EndTime - detail.BeginTime)
			// 是否观察位
			if user.Observe {
				stat.ObserveRounds++
				continue
			}
			stat.Bets += user.Bet
			if user.Bet > 0 {
				stat.BetRounds++
			}

			var escapeMulpitle float64
			_, err = fmt.Sscanf(user.Multiple, "%f", &escapeMulpitle)
			if err != nil {
				beego.Error(fmt.Scanf("crash user mulpitle parse error: %s", user.Multiple))
			}

			switch user.Result {
			case "赢":
				stat.WinRounds++
				stat.WinBets += user.Bet
				stat.Wins += user.Win
				stat.Cash += (user.Win)
				stat.WinEscapeMulpitleSum += escapeMulpitle
				stat.WinEscapeMulpitles = append(stat.WinEscapeMulpitles, escapeMulpitle)
				stat.WinMulpitleSum += mulpitle
				stat.WinMulpitles = append(stat.WinMulpitles, mulpitle)
			case "输":
				stat.LoseRounds++
				stat.LoseBets += user.Bet
				stat.Loses += user.Win
				stat.Cash += (user.Win)
				stat.LoseMulpitleSum += mulpitle
				stat.LoseMulpitles = append(stat.LoseMulpitles, mulpitle)
			case "平":
				stat.TieRounds++
				stat.TieBets += user.Bet
			}
			stat.MulpitleSum += mulpitle
			stat.Mulpitles = append(stat.Mulpitles, mulpitle)
		}
	}

	// 查询渠道，账号类型，充值金额
	m3 := []bson.M{
		{"$match": bson.M{
			"_id": bson.M{"$in": crashPlayerIds},
		}},
		{"$project": bson.M{
			"_id":           "$_id",
			"ad__bundle_id": "$ad__bundle_id",
			"channel1":      "$channel1",
			"regist_area":   "$regist_area",
			"money":         "$money",
		}},
	}
	var r3 []bson.M
	err = PlayerUsers.Pipe(m3).All(&r3)
	if err != nil {
		beego.Warning("CrashStats error3:", err)
	}
	for _, r := range r3 {
		userid := r["_id"].(string)
		ad__bundle_id := r["ad__bundle_id"]
		channel1 := r["channel1"]

		regist_area := r["regist_area"].(int)
		money := r["money"].(int)

		if p, ok := crashPlayers[userid]; ok {
			if ad__bundle_id != nil {
				p.AD_BundleId = ad__bundle_id.(string)
			}
			if channel1 != nil {
				p.Channel1 = channel1.(string)
			}
			p.RegistArea = regist_area
			p.Money = uint32(money)
		}
	}

	for _, p := range crashPlayers {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)

		// 保存数据
		addErr := StatisticsService.AddOrUpdateCrashPlayerStat(p)
		if addErr != nil {
			beego.Error("CrashPlayerStat fail err: ", addErr)
		}
	}
}

// CrashStatsCK Crash游戏统计
func CrashStrategyStatsCK(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
	date1, _ := utils.Unix(today)
	sql := `
		select userid,win_type,bet_amount,settle_score,before_score,
			crash_strategy_type,crash_multiple2,crash_win_result2,crash_rkyh_rfge
		FROM game.col_detail FINAL
		WHERE robot = 0 AND gtype = 7 AND notEmpty(crash_strategy_type) AND begin_time between ? and ?
	`
	var datas []map[string]any
	err := ck.Select(&datas, sql, startTime.Unix(), endTime.Unix())
	if err != nil {
		beego.Warning("no tp detail data", err)
		return
	}

	// 查询用户crash数据简单统计 ======================================
	sqlStats := `
		SELECT userid, win_type, count(*) rounds, SUM(bet_amount) bet_amount, SUM(settle_score) settle_score
		FROM game.col_detail FINAL 
		WHERE robot = 0 AND gtype = 7 AND begin_time between ? and ?
		GROUP BY userid, win_type
		order by userid, win_type
	`
	var dataStats []map[string]any
	err = ck.Select(&dataStats, sqlStats, startTime.Unix(), endTime.Unix())
	if err != nil {
		beego.Warning("no crash detail data", err)
		return
	}
	var stats = make(map[string]*entity.CrashPlayerStat)
	for _, data := range dataStats {
		userid := data["userid"].(string)
		winType := int8(utils.ToInt64(data["win_type"]))
		rounds := utils.ToInt64(data["rounds"])
		bet_amount := utils.ToInt64(data["bet_amount"])
		settle_score := utils.ToInt64(data["settle_score"])
		stat, ok := stats[userid]
		if !ok {
			stat = &entity.CrashPlayerStat{
				Id:      fmt.Sprintf("%d-%s", date1, userid),
				Userid:  userid,
				Date:    date1,
				DateStr: today,
			}
			stats[userid] = stat
		}
		stat.Bets += bet_amount
		stat.AllRounds += int32(rounds)
		switch winType {
		case ck.WinTypeWin:
			stat.BetRounds += int32(rounds)
			stat.WinRounds += int32(rounds)
			stat.WinBets += bet_amount
			stat.Wins += settle_score
			stat.Cash += (settle_score)
		case ck.WinTypeLose:
			stat.BetRounds += int32(rounds)
			stat.LoseRounds += int32(rounds)
			stat.LoseBets += bet_amount
			stat.Loses += settle_score
			stat.Cash += (settle_score)
		case ck.WinTypeTie:
			stat.BetRounds += int32(rounds)
			stat.TieRounds += int32(rounds)
			stat.TieBets += bet_amount
		case ck.WinTypeObserver:
			stat.ObserveRounds += int32(rounds)
		}
	}
	// 查询用户crash数据简单统计 ======================================

	// 扶摇直上
	var crashFyzs = make(map[string]*entity.CrashFYZSStat)
	// 欲薅无门
	var crashYhwm = make(map[string]*entity.CrashYHWMStat)
	// 起死回生
	var crashQshs = make(map[string]*entity.CrashQSHSStat)
	// 奖池风控
	var crashJcfk = make(map[string]*entity.CrashJCFKStat)
	// 冒险奖励
	var crashMxjl = make(map[string]*entity.CrashMXJLStat)
	// 人狂有祸
	var crashRkys = make(map[string]*entity.CrashRKYSStat)
	// 获取策略配置
	var str_info entity.CrashStrategy
	err = CrashStrategies.Find(bson.M{}).One(&str_info)
	if err != nil {
		beego.Error("regUsers error: ", err)
		return
	}

	var userids []string
	for _, data := range datas {
		userid := data["userid"].(string)
		winType := int8(utils.ToInt64(data["win_type"]))
		crash_strategy_type := data["crash_strategy_type"].([]int32)
		bet_amount := utils.ToInt64(data["bet_amount"])
		settle_score := utils.ToInt64(data["settle_score"])
		crash_win_result2 := utils.ToInt64(data["crash_win_result2"])
		crash_multiple2 := utils.ToInt64(data["crash_multiple2"]) // 逃脱
		// before_score := utils.ToInt64(data["before_score"])
		multiple := float64(crash_win_result2) / 100
		escape_multiple := float64(crash_multiple2) / 100
		crash_rkyh_rfge := utils.ToBool(data["crash_rkyh_rfge"])

		userids = append(userids, userid)

		// 判断是否触发扶摇直上策略
		isFyzs := false
		// 判断是否触发欲薅无门策略
		isYhwm := false
		// 起死回生
		isQshs := false
		// 奖池风控
		isJCFK := false
		// 冒险奖励
		isMxjl := false
		// 人狂有祸
		isRkys := false
		for _, st := range crash_strategy_type {
			switch st {
			case 1:
				isFyzs = true
			case 2:
				isYhwm = true
			case 4:
				isQshs = true
			case 5:
				isJCFK = true
			case 6:
				isMxjl = true
			case 7:
				isRkys = true
			}
		}
		// 触发扶摇直上
		if isFyzs {
			fyzs, ok := crashFyzs[userid]
			if !ok {
				fyzs = &entity.CrashFYZSStat{
					Id:      fmt.Sprintf("%d-%s", date1, userid),
					Date:    date1,
					DateStr: today,
					Userid:  userid,
					// IsFit:          isfit,
					// IsTrigger:      istrigger,
					// TiggerTimes:    tiggerTimes,
					// AllTiggerTimes: allTiggerTimes,
					// AllEvoTimes:    allEvoTimes,
				}
				crashFyzs[userid] = fyzs
			}
			fyzs.AllRounds++
			fyzs.Bets += bet_amount
			if bet_amount > 0 {
				fyzs.BetRounds++
			}
			var escapeMulpitle = escape_multiple
			switch winType {
			case ck.WinTypeWin:
				fyzs.WinRounds++
				fyzs.SumEscapeMulpitle += escape_multiple
				if fyzs.MaximumEscapeMultiple < escapeMulpitle {
					fyzs.MaximumEscapeMultiple = escapeMulpitle
				}
				if fyzs.MinimumEscapeMultiple == 0 {
					fyzs.MinimumEscapeMultiple = escapeMulpitle
				}
				if fyzs.MinimumEscapeMultiple > escapeMulpitle {
					fyzs.MinimumEscapeMultiple = escapeMulpitle
				}
				fyzs.WinMulpitleSum += escapeMulpitle
				if fyzs.MaximumEscapeMultipleBet < bet_amount {
					fyzs.MaximumEscapeMultipleBet = bet_amount
				}
				if fyzs.MinimumEscapeMultipleBet > bet_amount {
					fyzs.MinimumEscapeMultipleBet = bet_amount
				}
				if fyzs.MinimumEscapeMultipleBet == 0 {
					fyzs.MinimumEscapeMultipleBet = bet_amount
				}
				fyzs.WinBets += bet_amount
				fyzs.Wins += settle_score
				fyzs.Cash += settle_score
			case ck.WinTypeLose:
				fyzs.LoseRounds++
				fyzs.LoseBets += bet_amount
				fyzs.LoseMulpitleSum += multiple
				fyzs.Loses += settle_score
				fyzs.Cash += settle_score
			}
		}
		// 触发欲薅无门
		if isYhwm {
			yhwm, ok := crashYhwm[userid]
			if !ok {
				yhwm = &entity.CrashYHWMStat{
					Id:      fmt.Sprintf("%d-%s", date1, userid),
					Date:    date1,
					DateStr: today,
					Userid:  userid,
				}
				crashYhwm[userid] = yhwm
			}
			// 瞬爆局数
			if multiple == 1.0 {
				yhwm.BurstNumber++
				yhwm.BurstAmount += bet_amount
			}
			yhwm.AllRounds++
			yhwm.Bets += bet_amount
			if bet_amount > 0 {
				yhwm.BetRounds++
			}
			// var escapeMulpitle float64
			// _, err = fmt.Sscanf(user.Multiple, "%f", &escapeMulpitle)
			// if err != nil {
			// 	beego.Error(fmt.Scanf("crash user mulpitle parse error: %s", user.Multiple))
			// }

			switch winType {
			case ck.WinTypeWin:
				yhwm.WinRounds++
				yhwm.WinMulpitleSum += escape_multiple
				yhwm.WinBets += bet_amount
				yhwm.Wins += settle_score
				yhwm.Cash += settle_score
			case ck.WinTypeLose:
				yhwm.LoseRounds++
				yhwm.LoseBets += bet_amount
				yhwm.Loses += settle_score
				yhwm.Cash += settle_score
			}
		}
		// 触发起死回生
		if isQshs {
			qshs, ok := crashQshs[userid]
			if !ok {
				qshs = &entity.CrashQSHSStat{
					Id:      fmt.Sprintf("%d-%s", date1, userid),
					Date:    date1,
					DateStr: today,
					Userid:  userid,
				}
				crashQshs[userid] = qshs
				// crashQshsIds = append(crashQshsIds, userid)
			}

			// 触发
			qshs.AllEvoTimes++
			qshs.TiggerBets += bet_amount
			qshs.TiggerWinMulpitleSum += escape_multiple
			if qshs.MaximumEscapeMultiple < escape_multiple {
				qshs.MaximumEscapeMultiple = escape_multiple
			}
			switch winType {
			case ck.WinTypeWin:
				qshs.WinRounds++
				qshs.WinBets += bet_amount
				qshs.Cash += (settle_score)
				qshs.Wins += settle_score
			case ck.WinTypeLose:
				qshs.LoseEscapeMultiple += multiple
				qshs.LoseRounds++
				qshs.LoseBets += bet_amount
				qshs.Cash += (settle_score)
				qshs.Loses += settle_score
				if qshs.LoseMaxEscapeMultiple < multiple {
					qshs.LoseMaxEscapeMultiple = multiple
				}
			}
		}
		// 触发奖池风控
		if isJCFK {
			// 奖池风控
			jcfk, ok := crashJcfk[userid]
			if !ok {
				jcfk = &entity.CrashJCFKStat{
					Id:      fmt.Sprintf("%d-%s", date1, userid),
					Date:    date1,
					DateStr: today,
					Userid:  userid,
				}
				crashJcfk[userid] = jcfk
				// crashJcfkIds = append(crashJcfkIds, userid)
			}
			jcfk.AllEvoTimes++
			jcfk.TiggerBets += bet_amount
			// 爆炸 = 开奖
			if jcfk.MaximumEscapeMultiple < multiple {
				jcfk.MaximumEscapeMultiple = multiple
			}
			jcfk.TiggerMulpitleSum += multiple

			jcfk.AllRounds++
			jcfk.Bets += bet_amount
			switch winType {
			case ck.WinTypeWin:
				if escape_multiple > 20 {
					jcfk.TriggerRounds++
				}
				jcfk.WinMulpitleSum += escape_multiple
				jcfk.WinRounds++
				jcfk.WinBets += bet_amount
				jcfk.Wins += settle_score
				jcfk.Cash += (settle_score)
			case ck.WinTypeLose:
				jcfk.LoseMulpitleSum += escape_multiple
				jcfk.LoseRounds++
				jcfk.LoseBets += bet_amount
				jcfk.Cash += (settle_score)
				jcfk.Loses += settle_score
			}
		}
		if isMxjl {
			// 冒险奖励
			mxjl, ok := crashMxjl[userid]
			if !ok {
				mxjl = &entity.CrashMXJLStat{
					Id:      fmt.Sprintf("%d-%s", date1, userid),
					Date:    date1,
					DateStr: today,
					Userid:  userid,
				}
				crashMxjl[userid] = mxjl
				// crashJcfkIds = append(crashJcfkIds, userid)
			}
			mxjl.AllEvoTimes++
			mxjl.TiggerBets += bet_amount
			// 爆炸 = 开奖
			if mxjl.MaximumEscapeMultiple < multiple {
				mxjl.MaximumEscapeMultiple = multiple
			}
			if mxjl.WinMaxMulpitle < escape_multiple {
				mxjl.WinMaxMulpitle = escape_multiple
			}
			mxjl.TiggerMulpitleSum += multiple

			mxjl.AllRounds++
			mxjl.Bets += bet_amount
			switch winType {
			case ck.WinTypeWin:
				mxjl.WinMulpitleSum += escape_multiple
				mxjl.WinRounds++
				mxjl.WinBets += bet_amount
				mxjl.Wins += settle_score
				mxjl.Cash += (settle_score)
			case ck.WinTypeLose:
				mxjl.LoseMulpitleSum += escape_multiple
				mxjl.LoseRounds++
				mxjl.LoseBets += bet_amount
				mxjl.Cash += (settle_score)
				mxjl.Loses += settle_score
			}
		}
		// 触发人狂有祸
		if isRkys {
			//人狂有祸
			rkys, ok := crashRkys[userid]
			if !ok {
				rkys = &entity.CrashRKYSStat{
					Id:      fmt.Sprintf("%d-%s", date1, userid),
					Date:    date1,
					DateStr: today,
					Userid:  userid,
				}
				crashRkys[userid] = rkys
				// crashJcfkIds = append(crashJcfkIds, userid)
			}
			rkys.AllEvoTimes++
			rkys.TiggerBets += bet_amount
			// 爆炸 = 开奖
			if rkys.MaximumEscapeMultiple < multiple {
				rkys.MaximumEscapeMultiple = multiple
			}
			rkys.TiggerMulpitleSum += multiple

			// rkys.AllRounds++
			// rkys.Bets += bet_amount
			if crash_rkyh_rfge {
				rkys.RFRounds++
				rkys.RFBets += bet_amount
			}
			switch winType {
			case ck.WinTypeWin:
				rkys.WinMulpitleSum += escape_multiple
				rkys.WinRounds++
				rkys.WinBets += bet_amount
				rkys.Wins += settle_score
				rkys.Cash += (settle_score)
			case ck.WinTypeLose:
				rkys.LoseMulpitleSum += escape_multiple
				rkys.LoseRounds++
				rkys.LoseBets += bet_amount
				rkys.Cash += (settle_score)
				rkys.Loses += settle_score
				if crash_rkyh_rfge {
					rkys.RFLoseRounds++
					rkys.RFReapBets += settle_score
				}
			}
		}

	}
	// 起死回生all局数
	for _, data := range datas {
		userid := data["userid"].(string)
		bet_amount := utils.ToInt64(data["bet_amount"])
		before_score := utils.ToInt64(data["before_score"])
		if qshs, ok := crashQshs[userid]; ok {
			if str_info.QSHS.D >= before_score-bet_amount {
				qshs.AllinNumber++
				qshs.AllinBets += bet_amount
			}
		}
	}
	// 查询渠道，账号类型，充值金额
	var users []ck.User
	var usersMap = make(map[string]ck.User, 0)
	err = ck.Select(&users, `
		select userid,ad__bundle_id,channel,regist_area,money,state,coin,diamond,shadow_diamond,
			crash_fyzs_play_times,crash_fyzs_tigger_times,crash_fyzs_evo_times,crash_fyzs_all_tigger_times,crash_fyzs_all_evo_times,
			crash_yhwm_tigger_times,crash_yhwm_tigger,crash_yhwm_all_recyle_score,
			crash_jcfk_trigger_times,crash_jcfk_jackpot_val
		from col_user final where userid in ?
	`, userids)
	if err != nil {
		beego.Error("regUsers error: ", err)
		return
	}

	for _, r := range users {
		usersMap[r.Userid] = r
		// stat.AD_BundleId = user.AD_BundleId
		// stat.Channel1 = user.Channel
		// stat.RegistArea = user.RegistArea
		// stat.Money = user.Money

		userid := r.Userid
		isfit := false
		istrigger := false
		tiggerTimes := 0
		allTiggerTimes := 0
		allEvoTimes := 0

		// 扶摇直上
		if p, ok := crashFyzs[r.Userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.RegistArea = r.RegistArea
			userType := 0 // 账号类型
			isCharge := false
			isUser := false
			isM := false
			userType = ConvertToUserType(int(r.Money), r.State)

			if len(str_info.FYZS.ChargeType) > 0 {
				for ide, conf := range str_info.FYZS.ChargeType {
					if conf == 1 && userType == ide {
						isCharge = true
					}
				}
			}
			if len(str_info.FYZS.UserType) > 0 {
				for ide, conf := range str_info.FYZS.UserType {
					if conf == 1 && r.RegistArea == ide {
						isUser = true
					}
				}
			}
			curCoin := r.Coin + r.Diamond
			if str_info.FYZS.M > curCoin {
				isM = true
			}
			if isM && isCharge && isUser {
				isfit = true
			}
			// if r.RegistArea == 1 {
			// 	// B类用户判断是否有充值行为
			// 	n := bson.M{
			// 		"order_status": 4,
			// 		"userid":       r.Userid,
			// 	}
			// 	payCount, _ := Pays.Find(n).Count()
			// 	if payCount <= 0 {
			// 		isfit = true
			// 	}
			// }

			if r.CrashFYZSPlayTimes > 0 || r.CrashFYZSTiggerTimes > 0 || r.CrashFYZSEvoTimes > 0 {
				istrigger = true
			}
			tiggerTimes = r.CrashFYZSTiggerTimes
			allTiggerTimes = r.CrashFYZSAllTiggerTimes
			allEvoTimes = r.CrashFYZSAllEvoTimes
			p.IsFit = true // isfit
			p.IsTrigger = istrigger
			p.TiggerTimes = tiggerTimes
			p.AllTiggerTimes = allTiggerTimes
			p.AllEvoTimes = allEvoTimes
		}
		// 欲薅无门
		if p, ok := crashYhwm[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.RegistArea = r.RegistArea
			userType := 0 // 账号类型
			isCharge := false
			isUser := false

			if len(str_info.YHWM.ChargeType) > 0 {
				for ide, conf := range str_info.YHWM.ChargeType {
					if conf == 1 && userType == ide {
						isCharge = true
					}
				}
			}
			if len(str_info.YHWM.UserType) > 0 {
				for ide, conf := range str_info.YHWM.UserType {
					if conf == 1 && r.RegistArea == ide {
						isUser = true
					}
				}
			}
			if isCharge && isUser {
				isfit = true
			}
			if r.CrashYHWMTiggerTimes > 0 {
				istrigger = true
			}
			allTiggerTimes = r.CrashYHWMTiggerTimes
			if r.CrashYHWMTigger {
				tiggerTimes = allTiggerTimes - 1
			} else {
				tiggerTimes = allTiggerTimes
			}
			p.TotalReapAmount = r.CrashYHWMAllRecyleScore
			allEvoTimes = int(p.AllRounds)
			p.IsFit = isfit
			p.IsTrigger = istrigger

			p.TiggerTimes = tiggerTimes
			p.AllTiggerTimes = allTiggerTimes
			p.AllEvoTimes = allEvoTimes
		}

		// 起死回生
		if p, ok := crashQshs[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.RegistArea = r.RegistArea
			p.PayAmount = int64(r.Money)
			p.WithdrawAmount = int64(r.CashOut)
			p.CarryAmount = r.Diamond + r.ShadowDiamond
			userType := 0 // 账号类型
			isCharge := false
			isUser := false
			userType = ConvertToUserType(int(r.Money), r.State)
			if len(str_info.QSHS.ChargeType) > 0 {
				for ide, conf := range str_info.QSHS.ChargeType {
					if conf == 1 && userType == ide {
						isCharge = true
					}
				}
			}
			if len(str_info.QSHS.UserType) > 0 {
				for ide, conf := range str_info.QSHS.UserType {
					if conf == 1 && r.RegistArea == ide {
						isUser = true
					}
				}
			}
			if isCharge && isUser {
				isfit = true
			}
			if !isfit {
				p.AllinNumber = 0
				p.AllinBets = 0
			}
			// 获取用户crash游戏总汇数据
			if t, ok := stats[userid]; ok {
				p.AllRounds = t.BetRounds
				p.Bets = t.Bets
				p.AllWinBets = t.Wins
				p.AllLoseBets = t.Loses
			}
		}
		// 奖池风控
		if p, ok := crashJcfk[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.RegistArea = r.RegistArea
			p.PayAmount = int64(r.Money)
			p.WithdrawAmount = int64(r.CashOut)
			p.CarryAmount = r.Diamond + r.ShadowDiamond
			p.TriggerTimes = r.CrashJCFKTriggerTimes
			p.TriggerAmount = r.CrashJCFKJackpotVal
			// 获取用户crash游戏总汇数据
			if t, ok := stats[userid]; ok {
				p.AllRounds = t.BetRounds
				p.Bets = t.Bets
				p.AllWinBets = t.Wins
				p.AllLoseBets = t.Loses
			}
		}
		// 冒险奖励
		if p, ok := crashMxjl[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.RegistArea = r.RegistArea
			p.PayAmount = int64(r.Money)
			p.WithdrawAmount = int64(r.CashOut)
			p.CarryAmount = r.Diamond + r.ShadowDiamond
			// 获取用户crash游戏总汇数据
			if t, ok := stats[userid]; ok {
				p.AllRounds = t.BetRounds
				p.AllWinRounds = t.WinRounds
				p.Bets = t.Bets
				p.AllWinBets = t.Wins
				p.AllLoseBets = t.Loses
			}
		}

		// 人狂有祸
		if p, ok := crashRkys[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.RegistArea = r.RegistArea
			p.PayAmount = int64(r.Money)
			p.WithdrawAmount = int64(r.CashOut)
			p.CarryAmount = r.Diamond + r.ShadowDiamond
			// 获取用户crash游戏总汇数据
			if t, ok := stats[userid]; ok {
				p.AllRounds = t.BetRounds
				p.Bets = t.Bets
				p.AllWinBets = t.Wins
				p.AllLoseBets = t.Loses
			}
		}
	}

	// 扶摇直上
	for _, p := range crashFyzs {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)
		p.WinEscapeMultiple = p.WinMulpitleSum / float64(p.WinRounds)
		// 保存数据
		addErr := StatisticsService.AddOrUpdateCrashFYZSStat(CrashFYZSStats, p)
		if addErr != nil {
			beego.Error("CrashFyzsStats fail err: ", addErr)
		}
	}

	// 欲薅无门
	for _, p := range crashYhwm {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)
		p.WinEscapeMultiple = p.WinMulpitleSum / float64(p.WinRounds)
		// 保存数据
		addErr := StatisticsService.AddOrUpdateCrashYHWMStat(CrashYHWMStats, p)
		if addErr != nil {
			beego.Error("CrashYhwmStats fail err: ", addErr)
		}
	}

	// 起死回生
	for _, p := range crashQshs {
		// 如果没有符合allin的，就不添加数据
		if p.AllinNumber == 0 {
			continue
		}
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)
		// 保存数据
		addErr := StatisticsService.AddOrUpdateCrashQSHSStat(CrashQSHSStats, p)
		if addErr != nil {
			beego.Error("CrashQshsStats fail err: ", addErr)
		}
	}
	// 奖池风控
	for _, p := range crashJcfk {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)
		addErr := StatisticsService.AddOrUpdateCrashJCFKStat(CrashJCFKStats, p)
		if addErr != nil {
			beego.Error("CrashJcfkStats fail err: ", addErr)
		}
	}

	// 冒险奖励
	for _, p := range crashMxjl {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)
		addErr := StatisticsService.AddOrUpdateCrashMXJLStat(CrashMXJLStats, p)
		if addErr != nil {
			beego.Error("CrashMxjlStats fail err: ", addErr)
		}
	}

	// 人狂有祸
	for _, p := range crashRkys {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)
		addErr := StatisticsService.AddOrUpdateCrashRKYSStat(CrashRKYSStats, p)
		if addErr != nil {
			beego.Error("CrashRkysStats fail err: ", addErr)
		}
	}
}

// CrashStatsCK Crash游戏统计
func CrashStatsCK(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
	date1, _ := utils.Unix(today)

	sql := `
		SELECT userid, win_type, count(*) rounds, SUM(end_time - begin_time) game_times,
			SUM(bet_amount) bet_amount, SUM(settle_score) settle_score, 
			SUM(crash_multiple2+crash_multiple1_2) escape_multiple_sum, groupArray(crash_multiple2) escape_multiples, groupArray(crash_multiple1_2) escape_multiples1,
			SUM(crash_win_result2) multiple_sum,  groupArray(crash_win_result2) multiples
		FROM game.col_detail FINAL 
		WHERE robot = 0 AND gtype = 7 AND begin_time between ? and ?
		GROUP BY userid, win_type
		order by userid, win_type
	`

	var datas []map[string]any
	err := ck.Select(&datas, sql, startTime.Unix(), endTime.Unix())
	if err != nil {
		beego.Warning("no tp detail data", err)
		return
	}
	if len(datas) == 0 {
		return
	}

	// 查询当天注册用户
	var regUserids = make(map[string]bool)
	var regUseridsR []string
	err = ck.Select(&regUseridsR, `select userid from game.col_user final where ctime between ? and ?`, startTime, endTime)
	if err != nil {
		beego.Error("regUsers error: ", err)
		return
	}
	for _, userid := range regUseridsR {
		regUserids[userid] = true
	}

	// 玩家列表
	var userids []string
	var stats = make(map[string]*entity.CrashPlayerStat)

	for _, data := range datas {
		userid := data["userid"].(string)
		winType := int8(utils.ToInt64(data["win_type"]))
		rounds := utils.ToInt64(data["rounds"])
		game_times := utils.ToInt64(data["game_times"])
		bet_amount := utils.ToInt64(data["bet_amount"])
		settle_score := utils.ToInt64(data["settle_score"])
		escape_multiple_sum := utils.ToInt64(data["escape_multiple_sum"])
		escape_multiples := data["escape_multiples"].([]int32)
		escape_multiples1 := data["escape_multiples1"].([]int32)
		escape_multiples = append(escape_multiples, escape_multiples1...)
		multiple_sum := utils.ToInt64(data["multiple_sum"])
		multiples := data["multiples"].([]int32)
		var multiplesF []float64
		for _, m := range multiples {
			multiplesF = append(multiplesF, float64(m)/100)
		}

		userids = append(userids, userid)
		stat, ok := stats[userid]
		if !ok {
			stat = &entity.CrashPlayerStat{
				Id:      fmt.Sprintf("%d-%s", date1, userid),
				Userid:  userid,
				Date:    date1,
				DateStr: today,
				NewReg:  regUserids[userid], // 老顾客新顾客
			}
			stats[userid] = stat
		}
		stat.Bets += bet_amount
		stat.AllRounds += int32(rounds)
		stat.GameTimes += game_times
		if winType != 4 {
			stat.MulpitleSum += float64(multiple_sum) / 100
			stat.Mulpitles = append(stat.Mulpitles, multiplesF...)
		}

		switch winType {
		case ck.WinTypeWin:
			stat.BetRounds += int32(rounds)
			stat.WinRounds += int32(rounds)
			stat.WinBets += bet_amount
			stat.Wins += settle_score
			stat.Cash += (settle_score)

			var escape_multiplesF []float64
			for _, m := range escape_multiples {
				escape_multiplesF = append(escape_multiplesF, float64(m)/100)
			}
			stat.WinEscapeMulpitleSum += float64(escape_multiple_sum) / 100
			stat.WinEscapeMulpitles = append(stat.WinEscapeMulpitles, escape_multiplesF...)
			stat.WinMulpitleSum += float64(multiple_sum) / 100
			stat.WinMulpitles = append(stat.WinMulpitles, multiplesF...)
		case ck.WinTypeLose:
			stat.BetRounds += int32(rounds)
			stat.LoseRounds += int32(rounds)
			stat.LoseBets += bet_amount
			stat.Loses += settle_score
			stat.Cash += (settle_score)
			stat.LoseMulpitleSum += float64(multiple_sum) / 100
			stat.LoseMulpitles = append(stat.LoseMulpitles, multiplesF...)
		case ck.WinTypeTie:
			stat.BetRounds += int32(rounds)
			stat.TieRounds += int32(rounds)
			stat.TieBets += bet_amount
		case ck.WinTypeObserver:
			stat.ObserveRounds += int32(rounds)
		}
	}

	// 查询渠道，账号类型，充值金额
	var users []ck.User
	err = ck.Select(&users, `select userid,ad__bundle_id,channel,regist_area,money from col_user final where userid in ?`, userids)
	if err != nil {
		beego.Error("regUsers error: ", err)
		return
	}
	for _, user := range users {
		stat, ok := stats[user.Userid]
		if !ok {
			continue
		}
		stat.AD_BundleId = user.AD_BundleId
		stat.Channel1 = user.Channel
		stat.RegistArea = user.RegistArea
		stat.Money = user.Money
	}

	// 保存统计数据
	for _, p := range stats {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)

		// 保存数据
		addErr := StatisticsService.AddOrUpdateCrashPlayerStat(p)
		if addErr != nil {
			beego.Error("CrashPlayerStat fail err: ", addErr)
		}
	}
}

// CrashStats Crash游戏统计
func CrashStats(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
	date1, _ := utils.Unix(today)

	m1 := bson.M{
		"gtype":    int(pb.CRASH),
		"end_time": bson.M{"$gt": startTime.Unix(), "$lt": endTime.Unix()},
		// "players":    bson.M{"$ne": ""},
	}

	var r1 []entity.Detail
	err := Details.Find(m1).All(&r1)
	if err != nil {
		beego.Warning("CrashStats error1:", err)
		return
	}
	if len(r1) == 0 {
		beego.Warning("no crash detail data")
		return
	}

	// 查询当天注册用户
	m2 := []bson.M{
		{"$match": bson.M{
			"ctime": bson.M{"$gt": startTime, "$lt": endTime},
		}},
		{"$project": bson.M{"_id": "$_id"}},
	}
	var r2 []bson.M
	err = PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Warning("CrashStats error2:", err)
		return
	}
	var regUserids = make(map[string]bool)
	for _, r := range r2 {
		userid := r["_id"].(string)
		regUserids[userid] = true
	}

	// crash玩家列表
	var crashPlayerIds []string
	var crashPlayers = make(map[string]*entity.CrashPlayerStat)
	// 扶摇直上
	var crashFyzs = make(map[string]*entity.CrashFYZSStat)
	// 欲薅无门
	var crashYhwm = make(map[string]*entity.CrashYHWMStat)
	// 起死回生
	var crashQshs = make(map[string]*entity.CrashQSHSStat)
	// 奖池风控
	var crashJcfk = make(map[string]*entity.CrashJCFKStat)
	// 冒险奖励
	var crashMxjl = make(map[string]*entity.CrashMXJLStat)
	// 人狂有祸
	var crashRkys = make(map[string]*entity.CrashRKYSStat)
	// 获取适配人群配置
	var str_info entity.CrashStrategy
	err = CrashStrategies.Find(bson.M{}).One(&str_info)

	for _, detail := range r1 {
		if detail.CRASHDetail == nil {
			continue
		}
		// mulpitle := fmt.Sprintf("x%.2f", float32(detail.CRASHDetail.Result)/100)
		var mulpitle float64
		_, err = fmt.Sscanf(detail.CRASHDetail.Result, "x%f", &mulpitle)
		if err != nil {
			beego.Error(fmt.Scanf("crash mulpitle parse error: %s", detail.CRASHDetail.Result))
		}

		for _, user := range detail.CRASHDetail.UserDetail {
			if len(user.Userid) >= 16 {
				// 人机id18位长度+
				continue
			}
			stat, ok := crashPlayers[user.Userid]
			if !ok {
				stat = &entity.CrashPlayerStat{
					Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
					Date:    date1,
					DateStr: today,
					Userid:  user.Userid,
					NewReg:  regUserids[user.Userid], // 老顾客新顾客
				}
				crashPlayers[user.Userid] = stat
				crashPlayerIds = append(crashPlayerIds, user.Userid)
			}
			stat.AllRounds++
			stat.GameTimes += (detail.EndTime - detail.BeginTime)
			// 是否观察位
			if user.Observe {
				stat.ObserveRounds++
				continue
			}
			stat.Bets += user.Bet
			if user.Bet > 0 {
				stat.BetRounds++
			}

			var escapeMulpitle float64
			_, err = fmt.Sscanf(user.Multiple, "%f", &escapeMulpitle)
			if err != nil {
				beego.Error(fmt.Scanf("crash user mulpitle parse error: %s", user.Multiple))
			}

			switch user.Result {
			case "赢":
				stat.WinRounds++
				stat.WinBets += user.Bet
				stat.Wins += user.Win
				stat.Cash += (user.Win)
				stat.WinEscapeMulpitleSum += escapeMulpitle
				stat.WinEscapeMulpitles = append(stat.WinEscapeMulpitles, escapeMulpitle)
				stat.WinMulpitleSum += mulpitle
				stat.WinMulpitles = append(stat.WinMulpitles, mulpitle)
			case "输":
				stat.LoseRounds++
				stat.LoseBets += user.Bet
				stat.Loses += user.Win
				stat.Cash += (user.Win)
				stat.LoseMulpitleSum += mulpitle
				stat.LoseMulpitles = append(stat.LoseMulpitles, mulpitle)
			case "平":
				stat.TieRounds++
				stat.TieBets += user.Bet
			}
			stat.MulpitleSum += mulpitle
			stat.Mulpitles = append(stat.Mulpitles, mulpitle)

			// 判断是否触发扶摇直上策略
			isFyzs := false
			// 判断是否触发欲薅无门策略
			isYhwm := false
			// 起死回生
			isQshs := false
			// 奖池风控
			isJCFK := false
			// 冒险奖励
			isMxjl := false
			// 人狂有祸
			isRkys := false
			for _, item := range detail.CRASHDetail.StrategyType {
				if item == 1 {
					isFyzs = true
				}
				if item == 2 {
					isYhwm = true
				}
				if item == 4 {
					isQshs = true
				}
				if item == 5 {
					isJCFK = true
				}
				if item == 6 {
					isMxjl = true
				}
				if item == 7 {
					isRkys = true
				}
			}
			// 触发扶摇直上
			if isFyzs {
				fyzs, ok := crashFyzs[user.Userid]
				if !ok {
					fyzs = &entity.CrashFYZSStat{
						Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
						Date:    date1,
						DateStr: today,
						Userid:  user.Userid,
						// IsFit:          isfit,
						// IsTrigger:      istrigger,
						// TiggerTimes:    tiggerTimes,
						// AllTiggerTimes: allTiggerTimes,
						// AllEvoTimes:    allEvoTimes,
					}
					crashFyzs[user.Userid] = fyzs
				}
				fyzs.AllRounds++
				fyzs.Bets += user.Bet
				if user.Bet > 0 {
					fyzs.BetRounds++
				}
				var escapeMulpitle float64
				_, err = fmt.Sscanf(user.Multiple, "%f", &escapeMulpitle)
				if err != nil {
					beego.Error(fmt.Scanf("crash user mulpitle parse error: %s", user.Multiple))
				}
				switch user.Result {
				case "赢":
					fyzs.WinRounds++
					if fyzs.MaximumEscapeMultiple < escapeMulpitle {
						fyzs.MaximumEscapeMultiple = escapeMulpitle
					}
					if fyzs.MinimumEscapeMultiple == 0 {
						fyzs.MinimumEscapeMultiple = escapeMulpitle
					}
					if fyzs.MinimumEscapeMultiple > escapeMulpitle {
						fyzs.MinimumEscapeMultiple = escapeMulpitle
					}
					fyzs.WinMulpitleSum += escapeMulpitle
					if fyzs.MaximumEscapeMultipleBet < user.Bet {
						fyzs.MaximumEscapeMultipleBet = user.Bet
					}
					if fyzs.MinimumEscapeMultipleBet > user.Bet {
						fyzs.MinimumEscapeMultipleBet = user.Bet
					}
					if fyzs.MinimumEscapeMultipleBet == 0 {
						fyzs.MinimumEscapeMultipleBet = user.Bet
					}
					fyzs.WinBets += user.Bet
					fyzs.Wins += user.Win
					fyzs.Cash += (user.Win)
				case "输":
					fyzs.LoseRounds++
					fyzs.LoseBets += user.Bet
					fyzs.LoseMulpitleSum += mulpitle
					fyzs.Loses += user.Win
					fyzs.Cash += (user.Win)
				}
			}
			//触发欲薅无门
			if isYhwm {
				yhwm, ok := crashYhwm[user.Userid]
				if !ok {
					yhwm = &entity.CrashYHWMStat{
						Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
						Date:    date1,
						DateStr: today,
						Userid:  user.Userid,
					}
					crashYhwm[user.Userid] = yhwm
				}
				// 瞬爆局数
				if mulpitle == 1.0 {
					yhwm.BurstNumber++
					yhwm.BurstAmount += user.Bet
				}
				yhwm.AllRounds++
				yhwm.Bets += user.Bet
				if user.Bet > 0 {
					yhwm.BetRounds++
				}
				// var escapeMulpitle float64
				// _, err = fmt.Sscanf(user.Multiple, "%f", &escapeMulpitle)
				// if err != nil {
				// 	beego.Error(fmt.Scanf("crash user mulpitle parse error: %s", user.Multiple))
				// }

				switch user.Result {
				case "赢":
					yhwm.WinRounds++
					yhwm.WinMulpitleSum += escapeMulpitle
					yhwm.WinBets += user.Bet
					yhwm.Wins += user.Win
					yhwm.Cash += (user.Win)
				case "输":
					yhwm.LoseRounds++
					yhwm.LoseBets += user.Bet
					yhwm.Loses += user.Win
					yhwm.Cash += (user.Win)
				}
			}
			// 触发起死回生
			if isQshs {
				qshs, ok := crashQshs[user.Userid]
				if !ok {
					qshs = &entity.CrashQSHSStat{
						Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
						Date:    date1,
						DateStr: today,
						Userid:  user.Userid,
					}
					crashQshs[user.Userid] = qshs
					// crashQshsIds = append(crashQshsIds, user.Userid)
				}
				if str_info.QSHS.D >= user.BeforeScore {
					qshs.AllinNumber++
					qshs.AllinBets += user.Bet
				}
				// 触发
				qshs.AllEvoTimes++
				qshs.TiggerBets += user.Bet
				qshs.TiggerWinMulpitleSum += escapeMulpitle
				if qshs.MaximumEscapeMultiple < escapeMulpitle {
					qshs.MaximumEscapeMultiple = escapeMulpitle
				}
				switch user.Result {
				case "赢":
					qshs.WinRounds++
					qshs.WinBets += user.Bet
					qshs.Cash += (user.Win)
					qshs.Wins += user.Win
				case "输":
					qshs.LoseEscapeMultiple += mulpitle
					qshs.LoseRounds++
					qshs.LoseBets += user.Bet
					qshs.Cash += (user.Win)
					qshs.Loses += user.Win
					if qshs.LoseMaxEscapeMultiple < mulpitle {
						qshs.LoseMaxEscapeMultiple = mulpitle
					}
				}
			}
			// 触发奖池风控
			if isJCFK {
				// 奖池风控
				jcfk, ok := crashJcfk[user.Userid]
				if !ok {
					jcfk = &entity.CrashJCFKStat{
						Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
						Date:    date1,
						DateStr: today,
						Userid:  user.Userid,
					}
					crashJcfk[user.Userid] = jcfk
					// crashJcfkIds = append(crashJcfkIds, user.Userid)
				}
				jcfk.AllEvoTimes++
				jcfk.TiggerBets += user.Bet
				// 爆炸 = 开奖
				if jcfk.MaximumEscapeMultiple < mulpitle {
					jcfk.MaximumEscapeMultiple = mulpitle
				}
				jcfk.TiggerMulpitleSum += mulpitle

				jcfk.AllRounds++
				jcfk.Bets += user.Bet
				switch user.Result {
				case "赢":
					if escapeMulpitle > 20 {
						jcfk.TriggerRounds++
					}
					jcfk.WinMulpitleSum += escapeMulpitle
					jcfk.WinRounds++
					jcfk.WinBets += user.Bet
					jcfk.Wins += user.Win
					jcfk.Cash += (user.Win)
				case "输":
					jcfk.LoseMulpitleSum += escapeMulpitle
					jcfk.LoseRounds++
					jcfk.LoseBets += user.Bet
					jcfk.Cash += (user.Win)
					jcfk.Loses += user.Win
				}
			}
			if isMxjl {
				// 冒险奖励
				mxjl, ok := crashMxjl[user.Userid]
				if !ok {
					mxjl = &entity.CrashMXJLStat{
						Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
						Date:    date1,
						DateStr: today,
						Userid:  user.Userid,
					}
					crashMxjl[user.Userid] = mxjl
					// crashJcfkIds = append(crashJcfkIds, user.Userid)
				}
				mxjl.AllEvoTimes++
				mxjl.TiggerBets += user.Bet
				// 爆炸 = 开奖
				if mxjl.MaximumEscapeMultiple < mulpitle {
					mxjl.MaximumEscapeMultiple = mulpitle
				}
				mxjl.TiggerMulpitleSum += mulpitle

				mxjl.AllRounds++
				mxjl.Bets += user.Bet
				switch user.Result {
				case "赢":
					mxjl.WinMulpitleSum += escapeMulpitle
					mxjl.WinRounds++
					mxjl.WinBets += user.Bet
					mxjl.Wins += user.Win
					mxjl.Cash += (user.Win)
				case "输":
					mxjl.LoseMulpitleSum += escapeMulpitle
					mxjl.LoseRounds++
					mxjl.LoseBets += user.Bet
					mxjl.Cash += (user.Win)
					mxjl.Loses += user.Win
				}
			}
			// 触发人狂有祸
			if isRkys {
				//人狂有祸
				rkys, ok := crashRkys[user.Userid]
				if !ok {
					rkys = &entity.CrashRKYSStat{
						Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
						Date:    date1,
						DateStr: today,
						Userid:  user.Userid,
					}
					crashRkys[user.Userid] = rkys
					// crashJcfkIds = append(crashJcfkIds, user.Userid)
				}
				rkys.AllEvoTimes++
				rkys.TiggerBets += user.Bet
				// 爆炸 = 开奖
				if rkys.MaximumEscapeMultiple < mulpitle {
					rkys.MaximumEscapeMultiple = mulpitle
				}
				rkys.TiggerMulpitleSum += mulpitle

				rkys.AllRounds++
				rkys.Bets += user.Bet
				if detail.CRASHDetail.RkyhRfge {
					rkys.RFRounds++
					rkys.RFBets += user.Bet
				}
				switch user.Result {
				case "赢":
					rkys.WinMulpitleSum += escapeMulpitle
					rkys.WinRounds++
					rkys.WinBets += user.Bet
					rkys.Wins += user.Win
					rkys.Cash += (user.Win)
				case "输":
					rkys.LoseMulpitleSum += escapeMulpitle
					rkys.LoseRounds++
					rkys.LoseBets += user.Bet
					rkys.Cash += (user.Win)
					rkys.Loses += user.Win
					if detail.CRASHDetail.RkyhRfge {
						rkys.RFLoseRounds++
						rkys.RFReapBets += user.Win
					}
				}
			}

		}
	}

	// 查询渠道，账号类型，充值金额
	// m3 := []bson.M{
	// 	{"$match": bson.M{
	// 		"_id": bson.M{"$in": crashPlayerIds},
	// 	}},
	// 	{"$project": bson.M{
	// 		"_id":           "$_id",
	// 		"ad__bundle_id": "$ad__bundle_id",
	// 		"channel1":      "$channel1",
	// 		"regist_area":   "$regist_area",
	// 		"money":         "$money",
	// 	}},
	// }
	// var r3 []bson.M
	// err = PlayerUsers.Pipe(m3).All(&r3)
	// if err != nil {
	// 	beego.Warning("CrashStats error3:", err)
	// }

	// 查询user信息
	m3 := []bson.M{
		{"$match": bson.M{
			"_id": bson.M{"$in": crashPlayerIds},
		}},
	}
	var r3 []entity.PlayerUser
	err = PlayerUsers.Pipe(m3).All(&r3)
	if err != nil {
		beego.Warning("CrashStats error3:", err)
	}
	for _, r := range r3 {
		// userid := r["_id"].(string)
		// ad__bundle_id := r["ad__bundle_id"]
		// channel1 := r["channel1"]

		// regist_area := r["regist_area"].(int)
		// money := r["money"].(int)

		userid := r.Userid
		// money := r["money"].(int)
		isfit := false
		istrigger := false
		tiggerTimes := 0
		allTiggerTimes := 0
		allEvoTimes := 0
		// 明细汇总
		if p, ok := crashPlayers[userid]; ok {
			// if ad__bundle_id != nil {
			// 	p.AD_BundleId = ad__bundle_id.(string)
			// }
			// if channel1 != nil {
			// 	p.Channel1 = channel1.(string)
			// }
			p.AD_BundleId = r.AD_BundleId
			p.Channel1 = ""
			p.RegistArea = r.RegistArea
			p.Money = uint32(r.Money)
		}
		// 扶摇直上
		if p, ok := crashFyzs[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.RegistArea = r.RegistArea
			userType := 0 // 账号类型
			isCharge := false
			isUser := false
			isM := false
			userType = ConvertToUserType(int(r.Money), r.State)

			if len(str_info.FYZS.ChargeType) > 0 {
				for ide, conf := range str_info.FYZS.ChargeType {
					if conf == 1 && userType == ide {
						isCharge = true
					}
				}
			}
			if len(str_info.FYZS.UserType) > 0 {
				for ide, conf := range str_info.FYZS.UserType {
					if conf == 1 && r.RegistArea == ide {
						isUser = true
					}
				}
			}
			curCoin := r.Coin + r.Diamond
			if str_info.FYZS.M > curCoin {
				isM = true
			}
			if isM && isCharge && isUser {
				isfit = true
			}
			// if r.RegistArea == 1 {
			// 	// B类用户判断是否有充值行为
			// 	n := bson.M{
			// 		"order_status": 4,
			// 		"userid":       r.Userid,
			// 	}
			// 	payCount, _ := Pays.Find(n).Count()
			// 	if payCount <= 0 {
			// 		isfit = true
			// 	}
			// }

			if r.CrashStrategy.FYZS.PlayTimes > 0 || r.CrashStrategy.FYZS.TiggerTimes > 0 || r.CrashStrategy.FYZS.EvoTimes > 0 {
				istrigger = true
			}
			tiggerTimes = r.CrashStrategy.FYZS.TiggerTimes
			allTiggerTimes = r.CrashStrategy.FYZS.AllTiggerTimes
			allEvoTimes = r.CrashStrategy.FYZS.AllEvoTimes
			p.IsFit = isfit
			p.IsTrigger = istrigger
			p.TiggerTimes = tiggerTimes
			p.AllTiggerTimes = allTiggerTimes
			p.AllEvoTimes = allEvoTimes
		}
		// 欲薅无门
		if p, ok := crashYhwm[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.RegistArea = r.RegistArea
			userType := 0 // 账号类型
			isCharge := false
			isUser := false

			if len(str_info.YHWM.ChargeType) > 0 {
				for ide, conf := range str_info.YHWM.ChargeType {
					if conf == 1 && userType == ide {
						isCharge = true
					}
				}
			}
			if len(str_info.YHWM.UserType) > 0 {
				for ide, conf := range str_info.YHWM.UserType {
					if conf == 1 && r.RegistArea == ide {
						isUser = true
					}
				}
			}
			if isCharge && isUser {
				isfit = true
			}
			if r.CrashStrategy.YHWM.TiggerTimes > 0 {
				istrigger = true
			}
			allTiggerTimes = r.CrashStrategy.YHWM.TiggerTimes
			if r.CrashStrategy.YHWM.Tigger {
				tiggerTimes = allTiggerTimes - 1
			} else {
				tiggerTimes = allTiggerTimes
			}
			p.TotalReapAmount = r.CrashStrategy.YHWM.AllRecyleScore
			allEvoTimes = int(p.AllRounds)
			p.IsFit = isfit
			p.IsTrigger = istrigger

			p.TiggerTimes = tiggerTimes
			p.AllTiggerTimes = allTiggerTimes
			p.AllEvoTimes = allEvoTimes
		}
		// 起死回生
		if p, ok := crashQshs[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.RegistArea = r.RegistArea
			p.PayAmount = int64(r.Money)
			p.WithdrawAmount = int64(r.CashOut)
			p.CarryAmount = r.Diamond + r.ShadowDiamond
			userType := 0 // 账号类型
			isCharge := false
			isUser := false
			userType = ConvertToUserType(int(r.Money), r.State)
			if len(str_info.QSHS.ChargeType) > 0 {
				for ide, conf := range str_info.QSHS.ChargeType {
					if conf == 1 && userType == ide {
						isCharge = true
					}
				}
			}
			if len(str_info.QSHS.UserType) > 0 {
				for ide, conf := range str_info.QSHS.UserType {
					if conf == 1 && r.RegistArea == ide {
						isUser = true
					}
				}
			}
			if isCharge && isUser {
				isfit = true
			}
			if !isfit {
				p.AllinNumber = 0
				p.AllinBets = 0
			}
			// 获取用户crash游戏总汇数据
			if t, ok := crashPlayers[userid]; ok {
				p.AllRounds = t.BetRounds
				p.Bets = t.Bets
				p.AllWinBets = t.Wins
				p.AllLoseBets = t.Loses
			}
		}
		// 奖池风控
		if p, ok := crashJcfk[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.RegistArea = r.RegistArea
			p.PayAmount = int64(r.Money)
			p.WithdrawAmount = int64(r.CashOut)
			p.CarryAmount = r.Diamond + r.ShadowDiamond
			p.TriggerTimes = r.CrashStrategy.JCFK.TriggerTimes
			p.TriggerAmount = r.CrashStrategy.JCFK.JackpotVal
			// 获取用户crash游戏总汇数据
			if t, ok := crashPlayers[userid]; ok {
				p.AllRounds = t.BetRounds
				p.Bets = t.Bets
				p.AllWinBets = t.Wins
				p.AllLoseBets = t.Loses
			}
		}

		// 冒险奖励
		if p, ok := crashMxjl[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.RegistArea = r.RegistArea
			p.PayAmount = int64(r.Money)
			p.WithdrawAmount = int64(r.CashOut)
			p.CarryAmount = r.Diamond + r.ShadowDiamond
			// 获取用户crash游戏总汇数据
			if t, ok := crashPlayers[userid]; ok {
				p.AllRounds = t.BetRounds
				p.AllWinRounds = t.WinRounds
				p.Bets = t.Bets
				p.AllWinBets = t.Wins
				p.AllLoseBets = t.Loses
			}
		}

		// 人狂有祸
		if p, ok := crashRkys[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.RegistArea = r.RegistArea
			p.PayAmount = int64(r.Money)
			p.WithdrawAmount = int64(r.CashOut)
			p.CarryAmount = r.Diamond + r.ShadowDiamond
			// 获取用户crash游戏总汇数据
			if t, ok := crashPlayers[userid]; ok {
				p.AllRounds = t.BetRounds
				p.Bets = t.Bets
				p.AllWinBets = t.Wins
				p.AllLoseBets = t.Loses
			}
		}
	}

	for _, p := range crashPlayers {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)

		// 保存数据
		addErr := StatisticsService.AddOrUpdateCrashPlayerStat(p)
		if addErr != nil {
			beego.Error("CrashPlayerStat fail err: ", addErr)
		}
	}

	// 扶摇直上
	for _, p := range crashFyzs {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)
		p.WinEscapeMultiple = p.WinMulpitleSum / float64(p.WinRounds)
		// 保存数据
		addErr := StatisticsService.AddOrUpdateCrashFYZSStat(CrashFYZSStats, p)
		if addErr != nil {
			beego.Error("CrashFyzsStats fail err: ", addErr)
		}
	}

	// 欲薅无门
	for _, p := range crashYhwm {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)
		p.WinEscapeMultiple = p.WinMulpitleSum / float64(p.WinRounds)
		// 保存数据
		addErr := StatisticsService.AddOrUpdateCrashYHWMStat(CrashYHWMStats, p)
		if addErr != nil {
			beego.Error("CrashYhwmStats fail err: ", addErr)
		}
	}

	// 起死回生
	for _, p := range crashQshs {
		// 如果没有符合allin的，就不添加数据
		if p.AllinNumber == 0 {
			continue
		}
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)
		// 保存数据
		addErr := StatisticsService.AddOrUpdateCrashQSHSStat(CrashQSHSStats, p)
		if addErr != nil {
			beego.Error("CrashQshsStats fail err: ", addErr)
		}
	}
	// 奖池风控
	for _, p := range crashJcfk {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)
		addErr := StatisticsService.AddOrUpdateCrashJCFKStat(CrashQSHSStats, p)
		if addErr != nil {
			beego.Error("CrashJcfkStats fail err: ", addErr)
		}
	}

	// 冒险奖励
	for _, p := range crashMxjl {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)
		addErr := StatisticsService.AddOrUpdateCrashMXJLStat(CrashMXJLStats, p)
		if addErr != nil {
			beego.Error("CrashMxjlStats fail err: ", addErr)
		}
	}

	// 人狂有祸
	for _, p := range crashRkys {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)
		addErr := StatisticsService.AddOrUpdateCrashRKYSStat(CrashRKYSStats, p)
		if addErr != nil {
			beego.Error("CrashRkysStats fail err: ", addErr)
		}
	}
}

// PlainStatsCK Plain游戏统计
func PlainStatsCK(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
	date1, _ := utils.Unix(today)

	sql := `
	SELECT userid, win_type, count(*) rounds, SUM(end_time - begin_time) game_times,
		SUM(bet_amount) bet_amount, SUM(settle_score) settle_score, 
		SUM(crash_multiple2+crash_multiple1_2) escape_multiple_sum, groupArray(crash_multiple2) escape_multiples, groupArray(crash_multiple1_2) escape_multiples1,
		SUM(crash_win_result2) multiple_sum,  groupArray(crash_win_result2) multiples
	FROM game.col_detail FINAL 
	WHERE robot = 0 AND gtype = 10 AND begin_time between ? and ?
	GROUP BY userid, win_type
	order by userid, win_type
	`
	var datas []map[string]any
	err := ck.Select(&datas, sql, startTime.Unix(), endTime.Unix())
	if err != nil {
		beego.Warning("no tp detail data", err)
		return
	}
	if len(datas) == 0 {
		return
	}

	// 查询当天注册用户
	var regUserids = make(map[string]bool)
	var regUseridsR []string
	err = ck.Select(&regUseridsR, `select userid from game.col_user final where ctime between ? and ?`, startTime, endTime)
	if err != nil {
		beego.Error("regUsers error: ", err)
		return
	}
	for _, userid := range regUseridsR {
		regUserids[userid] = true
	}

	// 玩家列表
	var userids []string
	var stats = make(map[string]*entity.CrashPlayerStat)
	for _, data := range datas {
		userid := data["userid"].(string)
		winType := int8(utils.ToInt64(data["win_type"]))
		rounds := utils.ToInt64(data["rounds"])
		game_times := utils.ToInt64(data["game_times"])
		bet_amount := utils.ToInt64(data["bet_amount"])
		settle_score := utils.ToInt64(data["settle_score"])
		escape_multiple_sum := utils.ToInt64(data["escape_multiple_sum"])
		escape_multiples := data["escape_multiples"].([]int32)
		escape_multiples1 := data["escape_multiples1"].([]int32)
		escape_multiples = append(escape_multiples, escape_multiples1...)
		multiple_sum := utils.ToInt64(data["multiple_sum"])
		multiples := data["multiples"].([]int32)
		var multiplesF []float64
		for _, m := range multiples {
			multiplesF = append(multiplesF, float64(m)/100)
		}

		userids = append(userids, userid)
		stat, ok := stats[userid]
		if !ok {
			stat = &entity.CrashPlayerStat{
				Id:      fmt.Sprintf("%d-%s", date1, userid),
				Userid:  userid,
				Date:    date1,
				DateStr: today,
				NewReg:  regUserids[userid], // 老顾客新顾客
			}
			stats[userid] = stat
		}
		stat.Bets += bet_amount
		stat.AllRounds += int32(rounds)
		stat.GameTimes += game_times
		if winType != 4 {
			stat.MulpitleSum += float64(multiple_sum) / 100
			stat.Mulpitles = append(stat.Mulpitles, multiplesF...)
		}

		switch winType {
		case ck.WinTypeWin:
			stat.BetRounds += int32(rounds)
			stat.WinRounds += int32(rounds)
			stat.WinBets += bet_amount
			stat.Wins += settle_score
			stat.Cash += (settle_score)

			var escape_multiplesF []float64
			for _, m := range escape_multiples {
				escape_multiplesF = append(escape_multiplesF, float64(m)/100)
			}
			stat.WinEscapeMulpitleSum += float64(escape_multiple_sum) / 100
			stat.WinEscapeMulpitles = append(stat.WinEscapeMulpitles, escape_multiplesF...)
			stat.WinMulpitleSum += float64(multiple_sum) / 100
			stat.WinMulpitles = append(stat.WinMulpitles, multiplesF...)
		case ck.WinTypeLose:
			stat.BetRounds += int32(rounds)
			stat.LoseRounds += int32(rounds)
			stat.LoseBets += bet_amount
			stat.Loses += settle_score
			stat.Cash += (settle_score)
			stat.LoseMulpitleSum += float64(multiple_sum) / 100
			stat.LoseMulpitles = append(stat.LoseMulpitles, multiplesF...)
		case ck.WinTypeTie:
			stat.BetRounds += int32(rounds)
			stat.TieRounds += int32(rounds)
			stat.TieBets += bet_amount
		case ck.WinTypeObserver:
			stat.ObserveRounds += int32(rounds)
		}
	}

	// 查询渠道，账号类型，充值金额
	var users []ck.User
	err = ck.Select(&users, `select userid,ad__bundle_id,channel,regist_area,money from col_user final where userid in ?`, userids)
	if err != nil {
		beego.Error("regUsers error: ", err)
		return
	}
	for _, user := range users {
		stat, ok := stats[user.Userid]
		if !ok {
			continue
		}
		stat.AD_BundleId = user.AD_BundleId
		stat.Channel1 = user.Channel
		stat.RegistArea = user.RegistArea
		stat.Money = user.Money
	}

	// 保存统计数据
	for _, p := range stats {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)

		// 保存数据
		addErr := StatisticsService.AddOrUpdatePlanePlayerStat(p)
		if addErr != nil {
			beego.Error("PlanePlayerStat fail err: ", addErr)
		}
	}
}

// PlaneStrategyStatsCK 飞机游戏策略统计
func PlaneStrategyStatsCK(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
	date1, _ := utils.Unix(today)
	sql := `
		select userid,win_type,bet_amount,settle_score,before_score,
			crash_strategy_type,crash_multiple2,crash_win_result2,crash_rkyh_rfge
		FROM game.col_detail FINAL
		WHERE robot = 0 AND gtype = 10 AND notEmpty(crash_strategy_type) AND begin_time between ? and ?
	`
	var datas []map[string]any
	err := ck.Select(&datas, sql, startTime.Unix(), endTime.Unix())
	if err != nil {
		beego.Warning("no plane detail data", err)
		return
	}

	// 查询用户crash数据简单统计 ======================================
	sqlStats := `
		SELECT userid, win_type, count(*) rounds, SUM(bet_amount) bet_amount, SUM(settle_score) settle_score
		FROM game.col_detail FINAL 
		WHERE robot = 0 AND gtype = 10 AND begin_time between ? and ?
		GROUP BY userid, win_type
		order by userid, win_type
	`
	var dataStats []map[string]any
	err = ck.Select(&dataStats, sqlStats, startTime.Unix(), endTime.Unix())
	if err != nil {
		beego.Warning("no crash detail data", err)
		return
	}
	var stats = make(map[string]*entity.CrashPlayerStat)
	for _, data := range dataStats {
		userid := data["userid"].(string)
		winType := int8(utils.ToInt64(data["win_type"]))
		rounds := utils.ToInt64(data["rounds"])
		bet_amount := utils.ToInt64(data["bet_amount"])
		settle_score := utils.ToInt64(data["settle_score"])
		stat, ok := stats[userid]
		if !ok {
			stat = &entity.CrashPlayerStat{
				Id:      fmt.Sprintf("%d-%s", date1, userid),
				Userid:  userid,
				Date:    date1,
				DateStr: today,
			}
			stats[userid] = stat
		}
		stat.Bets += bet_amount
		stat.AllRounds += int32(rounds)
		switch winType {
		case ck.WinTypeWin:
			stat.BetRounds += int32(rounds)
			stat.WinRounds += int32(rounds)
			stat.WinBets += bet_amount
			stat.Wins += settle_score
			stat.Cash += (settle_score)
		case ck.WinTypeLose:
			stat.BetRounds += int32(rounds)
			stat.LoseRounds += int32(rounds)
			stat.LoseBets += bet_amount
			stat.Loses += settle_score
			stat.Cash += (settle_score)
		case ck.WinTypeTie:
			stat.BetRounds += int32(rounds)
			stat.TieRounds += int32(rounds)
			stat.TieBets += bet_amount
		case ck.WinTypeObserver:
			stat.ObserveRounds += int32(rounds)
		}
	}
	// 查询用户crash数据简单统计 ======================================

	// 扶摇直上
	var crashFyzs = make(map[string]*entity.CrashFYZSStat)
	// 欲薅无门
	var crashYhwm = make(map[string]*entity.CrashYHWMStat)
	// 起死回生
	var crashQshs = make(map[string]*entity.CrashQSHSStat)
	// 奖池风控
	var crashJcfk = make(map[string]*entity.CrashJCFKStat)
	// 冒险奖励
	var crashMxjl = make(map[string]*entity.CrashMXJLStat)
	// 人狂有祸
	var crashRkys = make(map[string]*entity.CrashRKYSStat)
	// 获取策略配置
	var str_info entity.AviatorStrategy
	err = PlaneStrategies.Find(bson.M{}).One(&str_info)
	if err != nil {
		beego.Error("regUsers error: ", err)
		return
	}

	// charge classify
	classify := entity.ChargeClassify{}
	GetByQ(ChargeClassifys, bson.M{"_id": 1}, &classify)

	var userids []string
	for _, data := range datas {
		userid := data["userid"].(string)
		userids = append(userids, userid)
	}
	// 查询渠道，账号类型，充值金额
	var users []ck.User
	var usersMap = make(map[string]ck.User, 0)
	err = ck.Select(&users, `
		select userid,ad__bundle_id,channel,regist_area,money,state, coin, diamond,
			av_fyzs_tigger_times,av_fyzs_evo_times,av_fyzs_all_tigger_times,av_fyzs_all_evo_times,
			av_yhwm_tigger_times,av_yhwm_tigger,av_yhwm_all_recyle_score,
			av_jcfk_trigger_times,av_jcfk_jackpot_val
		from col_user final where userid in ?
	`, userids)
	if err != nil {
		beego.Error("regUsers error: ", err)
		return
	}
	for _, user := range users {
		usersMap[user.Userid] = user
	}

	for _, data := range datas {
		userid := data["userid"].(string)
		winType := int8(utils.ToInt64(data["win_type"]))
		crash_strategy_type := data["crash_strategy_type"].([]int32)
		bet_amount := utils.ToInt64(data["bet_amount"])
		settle_score := utils.ToInt64(data["settle_score"])
		crash_win_result2 := utils.ToInt64(data["crash_win_result2"])
		crash_multiple2 := utils.ToInt64(data["crash_multiple2"]) // 逃脱
		// before_score := utils.ToInt64(data["before_score"])
		multiple := float64(crash_win_result2) / 100
		escape_multiple := float64(crash_multiple2) / 100
		crash_rkyh_rfge := utils.ToBool(data["crash_rkyh_rfge"])

		// 判断是否触发扶摇直上策略
		isFyzs := false
		// 判断是否触发欲薅无门策略
		isYhwm := false
		// 起死回生
		isQshs := false
		// 奖池风控
		isJCFK := false
		// 冒险奖励
		isMxjl := false
		// 人狂有祸
		isRkys := false
		for _, st := range crash_strategy_type {
			switch st {
			case 1:
				isFyzs = true
			case 2:
				isYhwm = true
			case 4:
				isQshs = true
			case 5:
				isJCFK = true
			case 6:
				isMxjl = true
			case 7:
				isRkys = true
			}
		}
		// 触发扶摇直上
		if isFyzs {
			fyzs, ok := crashFyzs[userid]
			if !ok {
				fyzs = &entity.CrashFYZSStat{
					Id:      fmt.Sprintf("%d-%s", date1, userid),
					Date:    date1,
					DateStr: today,
					Userid:  userid,
					// IsFit:          isfit,
					// IsTrigger:      istrigger,
					// TiggerTimes:    tiggerTimes,
					// AllTiggerTimes: allTiggerTimes,
					// AllEvoTimes:    allEvoTimes,
				}
				crashFyzs[userid] = fyzs
			}
			fyzs.AllRounds++
			fyzs.Bets += bet_amount
			if bet_amount > 0 {
				fyzs.BetRounds++
			}
			var escapeMulpitle = escape_multiple
			switch winType {
			case ck.WinTypeWin:
				fyzs.WinRounds++
				fyzs.SumEscapeMulpitle += escape_multiple
				if fyzs.MaximumEscapeMultiple < escapeMulpitle {
					fyzs.MaximumEscapeMultiple = escapeMulpitle
				}
				if fyzs.MinimumEscapeMultiple == 0 {
					fyzs.MinimumEscapeMultiple = escapeMulpitle
				}
				if fyzs.MinimumEscapeMultiple > escapeMulpitle {
					fyzs.MinimumEscapeMultiple = escapeMulpitle
				}
				fyzs.WinMulpitleSum += escapeMulpitle
				if fyzs.MaximumEscapeMultipleBet < bet_amount {
					fyzs.MaximumEscapeMultipleBet = bet_amount
				}
				if fyzs.MinimumEscapeMultipleBet > bet_amount {
					fyzs.MinimumEscapeMultipleBet = bet_amount
				}
				if fyzs.MinimumEscapeMultipleBet == 0 {
					fyzs.MinimumEscapeMultipleBet = bet_amount
				}
				fyzs.WinBets += bet_amount
				fyzs.Wins += settle_score
				fyzs.Cash += settle_score
			case ck.WinTypeLose:
				fyzs.LoseRounds++
				fyzs.LoseBets += bet_amount
				fyzs.LoseMulpitleSum += multiple
				fyzs.Loses += settle_score
				fyzs.Cash += settle_score
			}
		}
		// 触发欲薅无门
		if isYhwm {
			yhwm, ok := crashYhwm[userid]
			if !ok {
				yhwm = &entity.CrashYHWMStat{
					Id:      fmt.Sprintf("%d-%s", date1, userid),
					Date:    date1,
					DateStr: today,
					Userid:  userid,
				}
				crashYhwm[userid] = yhwm
			}
			// 瞬爆局数
			if multiple == 1.0 {
				yhwm.BurstNumber++
				yhwm.BurstAmount += bet_amount
			}
			yhwm.AllRounds++
			yhwm.Bets += bet_amount
			if bet_amount > 0 {
				yhwm.BetRounds++
			}
			// var escapeMulpitle float64
			// _, err = fmt.Sscanf(user.Multiple, "%f", &escapeMulpitle)
			// if err != nil {
			// 	beego.Error(fmt.Scanf("crash user mulpitle parse error: %s", user.Multiple))
			// }

			switch winType {
			case ck.WinTypeWin:
				yhwm.WinRounds++
				yhwm.WinMulpitleSum += escape_multiple
				yhwm.WinBets += bet_amount
				yhwm.Wins += settle_score
				yhwm.Cash += settle_score
			case ck.WinTypeLose:
				yhwm.LoseRounds++
				yhwm.LoseBets += bet_amount
				yhwm.Loses += settle_score
				yhwm.Cash += settle_score
			}
		}
		// 触发起死回生
		if isQshs {
			qshs, ok := crashQshs[userid]
			if !ok {
				qshs = &entity.CrashQSHSStat{
					Id:      fmt.Sprintf("%d-%s", date1, userid),
					Date:    date1,
					DateStr: today,
					Userid:  userid,
				}
				crashQshs[userid] = qshs
				// crashQshsIds = append(crashQshsIds, userid)
			}

			// 触发
			qshs.AllEvoTimes++
			qshs.TiggerBets += bet_amount
			qshs.TiggerWinMulpitleSum += escape_multiple
			if qshs.MaximumEscapeMultiple < escape_multiple {
				qshs.MaximumEscapeMultiple = escape_multiple
			}
			switch winType {
			case ck.WinTypeWin:
				qshs.WinRounds++
				qshs.WinBets += bet_amount
				qshs.Cash += (settle_score)
				qshs.Wins += settle_score
			case ck.WinTypeLose:
				qshs.LoseEscapeMultiple += multiple
				qshs.LoseRounds++
				qshs.LoseBets += bet_amount
				qshs.Cash += (settle_score)
				qshs.Loses += settle_score
				if qshs.LoseMaxEscapeMultiple < multiple {
					qshs.LoseMaxEscapeMultiple = multiple
				}
			}
		}
		// 触发奖池风控
		if isJCFK {
			// 奖池风控
			jcfk, ok := crashJcfk[userid]
			if !ok {
				jcfk = &entity.CrashJCFKStat{
					Id:      fmt.Sprintf("%d-%s", date1, userid),
					Date:    date1,
					DateStr: today,
					Userid:  userid,
				}
				crashJcfk[userid] = jcfk
				// crashJcfkIds = append(crashJcfkIds, userid)
			}
			jcfk.AllEvoTimes++
			jcfk.TiggerBets += bet_amount
			// 爆炸 = 开奖
			if jcfk.MaximumEscapeMultiple < multiple {
				jcfk.MaximumEscapeMultiple = multiple
			}
			jcfk.TiggerMulpitleSum += multiple

			jcfk.AllRounds++
			jcfk.Bets += bet_amount
			switch winType {
			case ck.WinTypeWin:
				if escape_multiple > 20 {
					jcfk.TriggerRounds++
				}
				jcfk.WinMulpitleSum += escape_multiple
				jcfk.WinRounds++
				jcfk.WinBets += bet_amount
				jcfk.Wins += settle_score
				jcfk.Cash += (settle_score)
			case ck.WinTypeLose:
				jcfk.LoseMulpitleSum += escape_multiple
				jcfk.LoseRounds++
				jcfk.LoseBets += bet_amount
				jcfk.Cash += (settle_score)
				jcfk.Loses += settle_score
			}
		}
		if isMxjl {
			// 冒险奖励
			mxjl, ok := crashMxjl[userid]
			if !ok {
				mxjl = &entity.CrashMXJLStat{
					Id:      fmt.Sprintf("%d-%s", date1, userid),
					Date:    date1,
					DateStr: today,
					Userid:  userid,
				}
				crashMxjl[userid] = mxjl
				// crashJcfkIds = append(crashJcfkIds, userid)
			}
			mxjl.AllEvoTimes++
			mxjl.TiggerBets += bet_amount
			// 爆炸 = 开奖
			if mxjl.MaximumEscapeMultiple < multiple {
				mxjl.MaximumEscapeMultiple = multiple
			}
			if mxjl.WinMaxMulpitle < escape_multiple {
				mxjl.WinMaxMulpitle = escape_multiple
			}
			mxjl.TiggerMulpitleSum += multiple

			mxjl.AllRounds++
			mxjl.Bets += bet_amount
			switch winType {
			case ck.WinTypeWin:
				mxjl.WinMulpitleSum += escape_multiple
				mxjl.WinRounds++
				mxjl.WinBets += bet_amount
				mxjl.Wins += settle_score
				mxjl.Cash += (settle_score)
			case ck.WinTypeLose:
				mxjl.LoseMulpitleSum += escape_multiple
				mxjl.LoseRounds++
				mxjl.LoseBets += bet_amount
				mxjl.Cash += (settle_score)
				mxjl.Loses += settle_score
			}
		}
		// 触发人狂有祸
		if isRkys {
			//人狂有祸
			rkys, ok := crashRkys[userid]
			if !ok {
				rkys = &entity.CrashRKYSStat{
					Id:      fmt.Sprintf("%d-%s", date1, userid),
					Date:    date1,
					DateStr: today,
					Userid:  userid,
				}
				crashRkys[userid] = rkys
				// crashJcfkIds = append(crashJcfkIds, userid)
			}
			rkys.AllEvoTimes++
			rkys.TiggerBets += bet_amount
			// 爆炸 = 开奖
			if rkys.MaximumEscapeMultiple < multiple {
				rkys.MaximumEscapeMultiple = multiple
			}
			rkys.TiggerMulpitleSum += multiple

			rkys.AllRounds++
			rkys.Bets += bet_amount
			if crash_rkyh_rfge {
				rkys.RFRounds++
				rkys.RFBets += bet_amount
			}
			switch winType {
			case ck.WinTypeWin:
				rkys.WinMulpitleSum += escape_multiple
				rkys.WinRounds++
				rkys.WinBets += bet_amount
				rkys.Wins += settle_score
				rkys.Cash += (settle_score)
			case ck.WinTypeLose:
				rkys.LoseMulpitleSum += escape_multiple
				rkys.LoseRounds++
				rkys.LoseBets += bet_amount
				rkys.Cash += (settle_score)
				rkys.Loses += settle_score
				if crash_rkyh_rfge {
					rkys.RFLoseRounds++
					rkys.RFReapBets += settle_score
				}
			}
		}

	}

	// 起死回生all局数
	for _, data := range datas {
		userid := data["userid"].(string)
		bet_amount := utils.ToInt64(data["bet_amount"])
		before_score := utils.ToInt64(data["before_score"])

		if qshs, ok := crashQshs[userid]; ok {
			if user, ok := usersMap[userid]; ok {
				ctype := GetChargeType(classify, int64(user.Money), user.State)
				if ctype >= 0 && ctype < len(str_info.QSHS.D) {
					if str_info.QSHS.D[ctype] >= before_score {
						qshs.AllinNumber++
						qshs.AllinBets += bet_amount
					}
				}
			}
		}
	}

	for _, r := range users {
		// usersMap[r.Userid] = r
		// stat.AD_BundleId = user.AD_BundleId
		// stat.Channel1 = user.Channel
		// stat.RegistArea = user.RegistArea
		// stat.Money = user.Money

		userid := r.Userid
		isfit := false
		istrigger := false
		tiggerTimes := 0
		allTiggerTimes := 0
		allEvoTimes := 0

		// 扶摇直上
		if p, ok := crashFyzs[r.Userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.RegistArea = r.RegistArea
			userType := 0 // 账号类型
			isCharge := false
			isUser := false
			isM := false
			userType = ConvertToUserType(int(r.Money), r.State)

			if len(str_info.FYZS.ChargeType) > 0 {
				for ide, conf := range str_info.FYZS.ChargeType {
					if conf == 1 && userType == ide {
						isCharge = true
					}
				}
			}
			if len(str_info.FYZS.UserType) > 0 {
				for ide, conf := range str_info.FYZS.UserType {
					if conf == 1 && r.RegistArea == ide {
						isUser = true
					}
				}
			}
			curCoin := r.Coin + r.Diamond

			if user, ok := usersMap[userid]; ok {
				ctype := GetChargeType(classify, int64(user.Money), user.State)
				if ctype >= 0 && ctype < len(str_info.FYZS.M) {
					if str_info.FYZS.M[ctype] > curCoin {
						isM = true
					}
				}
			}

			if isM && isCharge && isUser {
				isfit = true
			}
			// if r.RegistArea == 1 {
			// 	// B类用户判断是否有充值行为
			// 	n := bson.M{
			// 		"order_status": 4,
			// 		"userid":       r.Userid,
			// 	}
			// 	payCount, _ := Pays.Find(n).Count()
			// 	if payCount <= 0 {
			// 		isfit = true
			// 	}
			// }

			// r.CrashFYZSPlayTimes > 0 ||
			if r.AvFYZSTiggerTimes > 0 || r.AvFYZSEvoTimes > 0 {
				istrigger = true
			}
			tiggerTimes = r.AvFYZSTiggerTimes
			allTiggerTimes = r.AvFYZSAllTiggerTimes
			allEvoTimes = r.AvFYZSAllEvoTimes
			p.IsFit = true // isfit
			p.IsTrigger = istrigger
			p.TiggerTimes = tiggerTimes
			p.AllTiggerTimes = allTiggerTimes
			p.AllEvoTimes = allEvoTimes
		}
		// 欲薅无门
		if p, ok := crashYhwm[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.RegistArea = r.RegistArea
			userType := 0 // 账号类型
			isCharge := false
			isUser := false

			if len(str_info.YHWM.ChargeType) > 0 {
				for ide, conf := range str_info.YHWM.ChargeType {
					if conf == 1 && userType == ide {
						isCharge = true
					}
				}
			}
			if len(str_info.YHWM.UserType) > 0 {
				for ide, conf := range str_info.YHWM.UserType {
					if conf == 1 && r.RegistArea == ide {
						isUser = true
					}
				}
			}
			if isCharge && isUser {
				isfit = true
			}
			if r.AvYHWMTiggerTimes > 0 {
				istrigger = true
			}
			allTiggerTimes = r.AvYHWMTiggerTimes
			if r.CrashYHWMTigger {
				tiggerTimes = allTiggerTimes - 1
			} else {
				tiggerTimes = allTiggerTimes
			}
			p.TotalReapAmount = r.AvYHWMAllRecyleScore
			allEvoTimes = int(p.AllRounds)
			p.IsFit = isfit
			p.IsTrigger = istrigger

			p.TiggerTimes = tiggerTimes
			p.AllTiggerTimes = allTiggerTimes
			p.AllEvoTimes = allEvoTimes
		}

		// 起死回生
		if p, ok := crashQshs[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.RegistArea = r.RegistArea
			p.PayAmount = int64(r.Money)
			p.WithdrawAmount = int64(r.CashOut)
			p.CarryAmount = r.Diamond + r.ShadowDiamond
			userType := 0 // 账号类型
			isCharge := false
			isUser := false
			userType = ConvertToUserType(int(r.Money), r.State)
			if len(str_info.QSHS.ChargeType) > 0 {
				for ide, conf := range str_info.QSHS.ChargeType {
					if conf == 1 && userType == ide {
						isCharge = true
					}
				}
			}
			if len(str_info.QSHS.UserType) > 0 {
				for ide, conf := range str_info.QSHS.UserType {
					if conf == 1 && r.RegistArea == ide {
						isUser = true
					}
				}
			}
			if isCharge && isUser {
				isfit = true
			}
			if !isfit {
				p.AllinNumber = 0
				p.AllinBets = 0
			}
			// 获取用户crash游戏总汇数据
			if t, ok := stats[userid]; ok {
				p.AllRounds = t.BetRounds
				p.Bets = t.Bets
				p.AllWinBets = t.Wins
				p.AllLoseBets = t.Loses
			}
		}
		// 奖池风控
		if p, ok := crashJcfk[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.RegistArea = r.RegistArea
			p.PayAmount = int64(r.Money)
			p.WithdrawAmount = int64(r.CashOut)
			p.CarryAmount = r.Diamond + r.ShadowDiamond
			p.TriggerTimes = r.AvJCFKTriggerTimes
			p.TriggerAmount = r.AvJCFKJackpotVal
			// 获取用户av游戏总汇数据
			if t, ok := stats[userid]; ok {
				p.AllRounds = t.BetRounds
				p.Bets = t.Bets
				p.AllWinBets = t.Wins
				p.AllLoseBets = t.Loses
			}
		}
		// 冒险奖励
		if p, ok := crashMxjl[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.RegistArea = r.RegistArea
			p.PayAmount = int64(r.Money)
			p.WithdrawAmount = int64(r.CashOut)
			p.CarryAmount = r.Diamond + r.ShadowDiamond
			// 获取用户crash游戏总汇数据
			if t, ok := stats[userid]; ok {
				p.AllRounds = t.BetRounds
				p.AllWinRounds = t.WinRounds
				p.Bets = t.Bets
				p.AllWinBets = t.Wins
				p.AllLoseBets = t.Loses
			}
		}

		// 人狂有祸
		if p, ok := crashRkys[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.RegistArea = r.RegistArea
			p.PayAmount = int64(r.Money)
			p.WithdrawAmount = int64(r.CashOut)
			p.CarryAmount = r.Diamond + r.ShadowDiamond
			// 获取用户crash游戏总汇数据
			if t, ok := stats[userid]; ok {
				p.AllRounds = t.BetRounds
				p.Bets = t.Bets
				p.AllWinBets = t.Wins
				p.AllLoseBets = t.Loses
			}
		}
	}

	// 扶摇直上
	for _, p := range crashFyzs {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)
		p.WinEscapeMultiple = p.WinMulpitleSum / float64(p.WinRounds)
		// 保存数据
		addErr := StatisticsService.AddOrUpdateCrashFYZSStat(PlaneFYZSStats, p)
		if addErr != nil {
			beego.Error("PlaneFYZSStats fail err: ", addErr)
		}
	}

	// 欲薅无门
	for _, p := range crashYhwm {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)
		p.WinEscapeMultiple = p.WinMulpitleSum / float64(p.WinRounds)
		// 保存数据
		addErr := StatisticsService.AddOrUpdateCrashYHWMStat(PlaneYHWMStats, p)
		if addErr != nil {
			beego.Error("PlaneYHWMStats fail err: ", addErr)
		}
	}

	// 起死回生
	for _, p := range crashQshs {
		// 如果没有符合allin的，就不添加数据
		if p.AllinNumber == 0 {
			continue
		}
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)
		// 保存数据
		addErr := StatisticsService.AddOrUpdateCrashQSHSStat(PlaneQSHSStats, p)
		if addErr != nil {
			beego.Error("PlaneQSHSStats fail err: ", addErr)
		}
	}
	// 奖池风控
	for _, p := range crashJcfk {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)
		addErr := StatisticsService.AddOrUpdateCrashJCFKStat(PlaneJCFKStats, p)
		if addErr != nil {
			beego.Error("PlaneJCFKStats fail err: ", addErr)
		}
	}

	// 冒险奖励
	for _, p := range crashMxjl {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)
		addErr := StatisticsService.AddOrUpdateCrashMXJLStat(PlaneMXJLStats, p)
		if addErr != nil {
			beego.Error("PlaneMXJLStats fail err: ", addErr)
		}
	}

	// 人狂有祸
	for _, p := range crashRkys {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)
		addErr := StatisticsService.AddOrUpdateCrashRKYSStat(PlaneRKYSStats, p)
		if addErr != nil {
			beego.Error("PlaneRKYSStats fail err: ", addErr)
		}
	}
}

// Crash扶摇直上策略统计
func CrashFyzsStats(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
	date1, _ := utils.Unix(today)

	m1 := bson.M{
		"gtype":    bson.M{"$in": []int{7, 10}},
		"end_time": bson.M{"$gt": startTime.Unix(), "$lt": endTime.Unix()},
		// "players":    bson.M{"$ne": ""},
	}

	var r1 []entity.Detail
	err := Details.Find(m1).All(&r1)
	if err != nil {
		beego.Warning("CrashFyzsStats error1:", err)
		return
	}
	if len(r1) == 0 {
		beego.Warning("no crash detail data")
		return
	}

	// // 查询当天注册用户
	// m2 := []bson.M{
	// 	{"$match": bson.M{
	// 		"ctime": bson.M{"$gt": startTime, "$lt": endTime},
	// 	}},
	// 	{"$project": bson.M{"_id": "$_id"}},
	// }
	// var r2 []bson.M
	// err = PlayerUsers.Pipe(m2).All(&r2)
	// if err != nil {
	// 	beego.Warning("CrashFyzsStats error2:", err)
	// 	return
	// }
	// var regUserids = make(map[string]bool)
	// for _, r := range r2 {
	// 	userid := r["_id"].(string)
	// 	regUserids[userid] = true
	// }

	// crash玩家列表
	var crashFyzsIds []string
	var crashFyzs = make(map[string]*entity.CrashFYZSStat)

	for _, detail := range r1 {
		if detail.CRASHDetail == nil {
			continue
		}
		// 判断是否触发扶摇直上策略
		isFyzs := false
		for _, item := range detail.CRASHDetail.StrategyType {
			if item == 1 {
				isFyzs = true
			}
		}
		if !isFyzs {
			continue
		}
		// mulpitle := fmt.Sprintf("x%.2f", float32(detail.CRASHDetail.Result)/100)
		var mulpitle float64
		_, err = fmt.Sscanf(detail.CRASHDetail.Result, "x%f", &mulpitle)
		if err != nil {
			beego.Error(fmt.Scanf("crash mulpitle parse error: %s", detail.CRASHDetail.Result))
		}

		for _, user := range detail.CRASHDetail.UserDetail {
			if len(user.Userid) >= 16 {
				// 人机id18位长度+
				continue
			}

			stat, ok := crashFyzs[user.Userid]
			if !ok {
				stat = &entity.CrashFYZSStat{
					Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
					Date:    date1,
					DateStr: today,
					Userid:  user.Userid,
					// IsFit:          isfit,
					// IsTrigger:      istrigger,
					// TiggerTimes:    tiggerTimes,
					// AllTiggerTimes: allTiggerTimes,
					// AllEvoTimes:    allEvoTimes,
				}
				crashFyzs[user.Userid] = stat
				crashFyzsIds = append(crashFyzsIds, user.Userid)
			}
			stat.AllRounds++
			stat.Bets += user.Bet
			if user.Bet > 0 {
				stat.BetRounds++
			}
			var escapeMulpitle float64
			_, err = fmt.Sscanf(user.Multiple, "%f", &escapeMulpitle)
			if err != nil {
				beego.Error(fmt.Scanf("crash user mulpitle parse error: %s", user.Multiple))
			}
			switch user.Result {
			case "赢":
				stat.WinRounds++
				if stat.MaximumEscapeMultiple < escapeMulpitle {
					stat.MaximumEscapeMultiple = escapeMulpitle
				}
				if stat.MinimumEscapeMultiple == 0 {
					stat.MinimumEscapeMultiple = escapeMulpitle
				}
				if stat.MinimumEscapeMultiple > escapeMulpitle {
					stat.MinimumEscapeMultiple = escapeMulpitle
				}
				stat.WinMulpitleSum += escapeMulpitle
				if stat.MaximumEscapeMultipleBet < user.Bet {
					stat.MaximumEscapeMultipleBet = user.Bet
				}
				if stat.MinimumEscapeMultipleBet > user.Bet {
					stat.MinimumEscapeMultipleBet = user.Bet
				}
				if stat.MinimumEscapeMultipleBet == 0 {
					stat.MinimumEscapeMultipleBet = user.Bet
				}
				stat.WinBets += user.Bet
				stat.Wins += user.Win
				stat.Cash += (user.Win)
			case "输":
				stat.LoseRounds++
				stat.LoseBets += user.Bet
				stat.LoseMulpitleSum += mulpitle
				stat.Loses += user.Win
				stat.Cash += (user.Win)
			}
		}
	}
	// 获取适配人群配置
	var str_info entity.CrashStrategy
	err = CrashStrategies.Find(bson.M{}).One(&str_info)

	// 查询user信息
	m3 := []bson.M{
		{"$match": bson.M{
			"_id": bson.M{"$in": crashFyzsIds},
		}},
	}
	var r3 []entity.PlayerUser
	err = PlayerUsers.Pipe(m3).All(&r3)
	if err != nil {
		beego.Warning("CrashStats error3:", err)
	}
	for _, r := range r3 {
		userid := r.Userid
		// money := r["money"].(int)
		isfit := false
		istrigger := false
		tiggerTimes := 0
		allTiggerTimes := 0
		allEvoTimes := 0
		if p, ok := crashFyzs[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.RegistArea = r.RegistArea
			userType := 0 // 账号类型
			isCharge := false
			isUser := false
			isM := false
			userType = ConvertToUserType(int(r.Money), r.State)
			if len(str_info.FYZS.ChargeType) > 0 {
				for ide, conf := range str_info.FYZS.ChargeType {
					if conf == 1 && userType == ide {
						isCharge = true
					}
				}
			}
			if len(str_info.FYZS.UserType) > 0 {
				for ide, conf := range str_info.FYZS.UserType {
					if conf == 1 && r.RegistArea == ide {
						isUser = true
					}
				}
			}
			curCoin := r.Coin + r.Diamond
			if str_info.FYZS.M > curCoin {
				isM = true
			}
			if isM && isCharge && isUser {
				isfit = true
			}
			// if r.RegistArea == 1 {
			// 	// B类用户判断是否有充值行为
			// 	n := bson.M{
			// 		"order_status": 4,
			// 		"userid":       r.Userid,
			// 	}
			// 	payCount, _ := Pays.Find(n).Count()
			// 	if payCount <= 0 {
			// 		isfit = true
			// 	}
			// }

			if r.CrashStrategy.FYZS.PlayTimes > 0 || r.CrashStrategy.FYZS.TiggerTimes > 0 || r.CrashStrategy.FYZS.EvoTimes > 0 {
				istrigger = true
			}
			tiggerTimes = r.CrashStrategy.FYZS.TiggerTimes
			allTiggerTimes = r.CrashStrategy.FYZS.AllTiggerTimes
			allEvoTimes = r.CrashStrategy.FYZS.AllEvoTimes
			p.IsFit = isfit
			p.IsTrigger = istrigger
			p.TiggerTimes = tiggerTimes
			p.AllTiggerTimes = allTiggerTimes
			p.AllEvoTimes = allEvoTimes
		}

	}

	for _, p := range crashFyzs {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)
		p.WinEscapeMultiple = p.WinMulpitleSum / float64(p.WinRounds)
		// 保存数据
		addErr := StatisticsService.AddOrUpdateCrashFYZSStat(CrashFYZSStats, p)
		if addErr != nil {
			beego.Error("CrashFyzsStats fail err: ", addErr)
		}
	}
}

// Crash欲薅无门策略统计
func CrashYhwmStats(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
	date1, _ := utils.Unix(today)

	m1 := bson.M{
		"gtype":    bson.M{"$in": []int{7, 10}},
		"end_time": bson.M{"$gt": startTime.Unix(), "$lt": endTime.Unix()},
		// "players":    bson.M{"$ne": ""},
	}

	var r1 []entity.Detail
	err := Details.Find(m1).All(&r1)
	if err != nil {
		beego.Warning("CrashYhwmStats error1:", err)
		return
	}
	if len(r1) == 0 {
		beego.Warning("no crash detail data")
		return
	}

	// // 查询当天注册用户
	// m2 := []bson.M{
	// 	{"$match": bson.M{
	// 		"ctime": bson.M{"$gt": startTime, "$lt": endTime},
	// 	}},
	// 	{"$project": bson.M{"_id": "$_id"}},
	// }
	// var r2 []bson.M
	// err = PlayerUsers.Pipe(m2).All(&r2)
	// if err != nil {
	// 	beego.Warning("CrashYhwmStats error2:", err)
	// 	return
	// }
	// var regUserids = make(map[string]bool)
	// for _, r := range r2 {
	// 	userid := r["_id"].(string)
	// 	regUserids[userid] = true
	// }

	// crash玩家列表
	var crashYhwmIds []string
	var crashYhwm = make(map[string]*entity.CrashYHWMStat)

	for _, detail := range r1 {
		if detail.CRASHDetail == nil {
			continue
		}
		// 判断是否触发欲薅无门策略
		isYhwm := false
		for _, item := range detail.CRASHDetail.StrategyType {
			if item == 2 {
				isYhwm = true
			}
		}
		if !isYhwm {
			continue
		}
		// mulpitle := fmt.Sprintf("x%.2f", float32(detail.CRASHDetail.Result)/100)
		var mulpitle float64
		_, err = fmt.Sscanf(detail.CRASHDetail.Result, "x%f", &mulpitle)
		if err != nil {
			beego.Error(fmt.Scanf("crash mulpitle parse error: %s", detail.CRASHDetail.Result))
		}

		for _, user := range detail.CRASHDetail.UserDetail {
			if len(user.Userid) >= 16 {
				// 人机id18位长度+
				continue
			}

			stat, ok := crashYhwm[user.Userid]
			if !ok {
				stat = &entity.CrashYHWMStat{
					Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
					Date:    date1,
					DateStr: today,
					Userid:  user.Userid,
				}
				crashYhwm[user.Userid] = stat
				crashYhwmIds = append(crashYhwmIds, user.Userid)
			}
			// 瞬爆局数
			if mulpitle == 1.0 {
				stat.BurstNumber++
				stat.BurstAmount += user.Bet
			}
			stat.AllRounds++
			stat.Bets += user.Bet
			if user.Bet > 0 {
				stat.BetRounds++
			}
			var escapeMulpitle float64
			_, err = fmt.Sscanf(user.Multiple, "%f", &escapeMulpitle)
			if err != nil {
				beego.Error(fmt.Scanf("crash user mulpitle parse error: %s", user.Multiple))
			}

			switch user.Result {
			case "赢":
				stat.WinRounds++
				stat.WinMulpitleSum += escapeMulpitle
				stat.WinBets += user.Bet
				stat.Wins += user.Win
				stat.Cash += (user.Win)
			case "输":
				stat.LoseRounds++
				stat.LoseBets += user.Bet
				stat.Loses += user.Win
				stat.Cash += (user.Win)
			}
		}
	}
	// 获取适配人群配置
	var str_info entity.CrashStrategy
	err = CrashStrategies.Find(bson.M{}).One(&str_info)

	// 查询user信息
	m3 := []bson.M{
		{"$match": bson.M{
			"_id": bson.M{"$in": crashYhwmIds},
		}},
	}
	var r3 []entity.PlayerUser
	err = PlayerUsers.Pipe(m3).All(&r3)
	if err != nil {
		beego.Warning("CrashYhwmStats error3:", err)
	}
	for _, r := range r3 {
		userid := r.Userid
		// money := r["money"].(int)
		isfit := false
		istrigger := false
		tiggerTimes := 0
		allTiggerTimes := 0
		allEvoTimes := 0
		if p, ok := crashYhwm[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.RegistArea = r.RegistArea
			// userType := 0 // 账号类型

			// if r.Money == 0 {
			// 	userType = 0
			// } else if r.Money == 0 && r.State == 3 {
			// 	userType = 1
			// } else if r.Money >= 20000 && r.Money <= 99999 {
			// 	userType = 2
			// } else if r.Money >= 100000 && r.Money <= 499999 {
			// 	userType = 3
			// } else if r.Money >= 500000 && r.Money <= 999999 {
			// 	userType = 4
			// } else if r.Money >= 1000000 && r.Money <= 9999999 {
			// 	userType = 5
			// } else if r.Money >= 10000000 {
			// 	userType = 6
			// }
			// if len(str_info.YHWM.ChargeType) > 0 {
			// 	for ide, conf := range str_info.YHWM.ChargeType {
			// 		if conf == 1 && userType == ide {
			// 			isfit = true
			// 		}
			// 	}
			// }
			userType := 0 // 账号类型
			isCharge := false
			isUser := false
			userType = ConvertToUserType(int(r.Money), r.State)
			if len(str_info.YHWM.ChargeType) > 0 {
				for ide, conf := range str_info.YHWM.ChargeType {
					if conf == 1 && userType == ide {
						isCharge = true
					}
				}
			}
			if len(str_info.YHWM.UserType) > 0 {
				for ide, conf := range str_info.YHWM.UserType {
					if conf == 1 && r.RegistArea == ide {
						isUser = true
					}
				}
			}
			if isCharge && isUser {
				isfit = true
			}
			if r.CrashStrategy.YHWM.TiggerTimes > 0 {
				istrigger = true
			}
			allTiggerTimes = r.CrashStrategy.YHWM.TiggerTimes
			if r.CrashStrategy.YHWM.Tigger {
				tiggerTimes = allTiggerTimes - 1
			} else {
				tiggerTimes = allTiggerTimes
			}
			p.TotalReapAmount = r.CrashStrategy.YHWM.AllRecyleScore
			allEvoTimes = int(p.AllRounds)
			p.IsFit = isfit
			p.IsTrigger = istrigger

			p.TiggerTimes = tiggerTimes
			p.AllTiggerTimes = allTiggerTimes
			p.AllEvoTimes = allEvoTimes
		}

	}

	for _, p := range crashYhwm {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)
		p.WinEscapeMultiple = p.WinMulpitleSum / float64(p.WinRounds)
		// 保存数据
		addErr := StatisticsService.AddOrUpdateCrashYHWMStat(CrashYHWMStats, p)
		if addErr != nil {
			beego.Error("CrashYhwmStats fail err: ", addErr)
		}
	}
}

// Crash起死回生策略统计
func CrashQshsOrJcfkStats(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
	date1, _ := utils.Unix(today)
	m1 := bson.M{
		"gtype":    bson.M{"$in": []int{7, 10}},
		"end_time": bson.M{"$gt": startTime.Unix(), "$lt": endTime.Unix()},
		// "players":    bson.M{"$ne": ""},
	}

	var r1 []entity.Detail
	err := Details.Find(m1).All(&r1)
	if err != nil {
		beego.Warning("CrashQshsStats error1:", err)
		return
	}
	if len(r1) == 0 {
		beego.Warning("no crash detail data")
		return
	}

	// crash玩家列表
	var crashQshsIds []string
	var crashQshs = make(map[string]*entity.CrashQSHSStat)

	var crashJcfkIds []string
	var crashJcfk = make(map[string]*entity.CrashJCFKStat)
	// 获取适配人群配置
	var str_info entity.CrashStrategy
	err = CrashStrategies.Find(bson.M{}).One(&str_info)

	for _, detail := range r1 {
		if detail.CRASHDetail == nil {
			continue
		}
		// 起死回生
		isQshs := false
		// 奖池风控
		isJCFK := false
		for _, item := range detail.CRASHDetail.StrategyType {
			if item == 4 {
				isQshs = true
			}
			if item == 5 {
				isJCFK = true
			}
		}
		var mulpitle float64
		_, err = fmt.Sscanf(detail.CRASHDetail.Result, "x%f", &mulpitle)
		if err != nil {
			beego.Error(fmt.Scanf("crash mulpitle parse error: %s", detail.CRASHDetail.Result))
		}
		for _, user := range detail.CRASHDetail.UserDetail {
			if len(user.Userid) >= 16 {
				// 人机id18位长度+
				continue
			}
			stat, ok := crashQshs[user.Userid]
			if !ok {
				stat = &entity.CrashQSHSStat{
					Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
					Date:    date1,
					DateStr: today,
					Userid:  user.Userid,
				}
				crashQshsIds = append(crashQshsIds, user.Userid)
			}
			stat.AllRounds++
			stat.Bets += user.Bet
			switch user.Result {
			case "赢":
				stat.AllWinBets += user.Win
			case "输":
				stat.AllLoseBets += user.Win
			}
			if str_info.QSHS.D >= user.BeforeScore {
				stat.AllinNumber++
				stat.AllinBets += user.Bet
			}
			// if user.Bet != 0 {
			// 	stat.BetRounds++
			// }
			if isQshs {
				// 触发
				stat.AllEvoTimes++
				stat.TiggerBets += user.Bet
				var escapeMulpitle float64
				_, err = fmt.Sscanf(user.Multiple, "%f", &escapeMulpitle)
				if err != nil {
					beego.Error(fmt.Scanf("crash user mulpitle parse error: %s", user.Multiple))
				}
				stat.TiggerWinMulpitleSum += escapeMulpitle
				if stat.MaximumEscapeMultiple < escapeMulpitle {
					stat.MaximumEscapeMultiple = escapeMulpitle
				}
				switch user.Result {
				case "赢":
					stat.WinRounds++
					stat.WinBets += user.Bet
					stat.Cash += (user.Win)
					stat.Wins += user.Win
				case "输":
					stat.LoseEscapeMultiple += mulpitle
					stat.LoseRounds++
					stat.LoseBets += user.Bet
					stat.Cash += (user.Win)
					stat.Loses += user.Win
					if stat.LoseMaxEscapeMultiple < mulpitle {
						stat.LoseMaxEscapeMultiple = mulpitle
					}
				}
			}
			crashQshs[user.Userid] = stat

			// 奖池风控
			jcfk, ok := crashJcfk[user.Userid]
			if !ok {
				jcfk = &entity.CrashJCFKStat{
					Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
					Date:    date1,
					DateStr: today,
					Userid:  user.Userid,
				}
				crashJcfkIds = append(crashJcfkIds, user.Userid)
			}
			jcfk.AllRounds++
			jcfk.Bets += user.Bet
			switch user.Result {
			case "赢":
				jcfk.AllWinBets += user.Win
			case "输":
				jcfk.AllLoseBets += user.Win
			}
			if isJCFK {
				// 触发奖池风控

				jcfk.AllEvoTimes++
				jcfk.TiggerBets += user.Bet
				// 爆炸 = 开奖
				if stat.MaximumEscapeMultiple < mulpitle {
					stat.MaximumEscapeMultiple = mulpitle
				}
				jcfk.TiggerMulpitleSum += mulpitle

				jcfk.AllRounds++
				jcfk.Bets += user.Bet
				var escapeMulpitle float64
				_, err = fmt.Sscanf(user.Multiple, "%f", &escapeMulpitle)
				if err != nil {
					beego.Error(fmt.Scanf("crash user mulpitle parse error: %s", user.Multiple))
				}
				switch user.Result {
				case "赢":
					if escapeMulpitle > 20 {
						jcfk.TriggerRounds++
					}
					jcfk.WinMulpitleSum += escapeMulpitle
					jcfk.WinRounds++
					jcfk.WinBets += user.Bet
					jcfk.Wins += user.Win
					jcfk.Cash += (user.Win)
				case "输":
					jcfk.LoseMulpitleSum += escapeMulpitle
					jcfk.LoseRounds++
					jcfk.LoseBets += user.Bet
					jcfk.Cash += (user.Win)
					jcfk.Loses += user.Win
				}

			}
			crashJcfk[user.Userid] = jcfk
		}

	}

	// 判断是否存在起死回生策略玩家，如果存在,查询用户相关信息
	if len(crashQshsIds) > 0 {
		// 查询user信息
		m3 := []bson.M{
			{"$match": bson.M{
				"_id": bson.M{"$in": crashQshsIds},
			}},
		}
		var r3 []entity.PlayerUser
		err = PlayerUsers.Pipe(m3).All(&r3)
		if err != nil {
			beego.Warning("CrashQshsStats error3:", err)
		}
		for _, r := range r3 {
			userid := r.Userid
			// money := r["money"].(int)
			isfit := false
			// istrigger := false
			// tiggerTimes := 0
			// allTiggerTimes := 0
			// allEvoTimes := 0
			if p, ok := crashQshs[userid]; ok {
				p.AD_BundleId = r.AD_BundleId
				p.RegistArea = r.RegistArea
				p.PayAmount = int64(r.Money)
				p.WithdrawAmount = int64(r.CashOut)
				p.CarryAmount = r.Diamond + r.ShadowDiamond
				userType := 0 // 账号类型
				isCharge := false
				isUser := false
				userType = ConvertToUserType(int(r.Money), r.State)
				if len(str_info.QSHS.ChargeType) > 0 {
					for ide, conf := range str_info.QSHS.ChargeType {
						if conf == 1 && userType == ide {
							isCharge = true
						}
					}
				}
				if len(str_info.QSHS.UserType) > 0 {
					for ide, conf := range str_info.QSHS.UserType {
						if conf == 1 && r.RegistArea == ide {
							isUser = true
						}
					}
				}
				if isCharge && isUser {
					isfit = true
				}
				if !isfit {
					p.AllinNumber = 0
					p.AllinBets = 0
				}
			}

		}
		// 起死回生
		for _, p := range crashQshs {
			// 如果没有符合allin的，就不添加数据
			if p.AllinNumber == 0 {
				continue
			}
			// 局均码
			p.RoundBetAvg = p.Bets / int64(p.AllRounds)
			// 保存数据
			addErr := StatisticsService.AddOrUpdateCrashQSHSStat(CrashQSHSStats, p)
			if addErr != nil {
				beego.Error("CrashQshsStats fail err: ", addErr)
			}
		}
	}
	// 判断是否存在奖池风控策略玩家，如果存在,查询用户相关信息
	if len(crashJcfkIds) > 0 {
		// 查询user信息
		m3 := []bson.M{
			{"$match": bson.M{
				"_id": bson.M{"$in": crashJcfkIds},
			}},
		}
		var r3 []entity.PlayerUser
		err = PlayerUsers.Pipe(m3).All(&r3)
		if err != nil {
			beego.Warning("CrashJcfkStats error3:", err)
		}
		for _, r := range r3 {
			userid := r.Userid
			if p, ok := crashJcfk[userid]; ok {
				p.AD_BundleId = r.AD_BundleId
				p.RegistArea = r.RegistArea
				p.PayAmount = int64(r.Money)
				p.WithdrawAmount = int64(r.CashOut)
				p.CarryAmount = r.Diamond + r.ShadowDiamond
				p.TriggerTimes = r.CrashStrategy.JCFK.TriggerTimes
				p.TriggerAmount = r.CrashStrategy.JCFK.JackpotVal

			}
		}
		// 奖池风控
		for _, p := range crashJcfk {
			// 局均码
			p.RoundBetAvg = p.Bets / int64(p.AllRounds)
			addErr := StatisticsService.AddOrUpdateCrashJCFKStat(CrashJCFKStats, p)
			if addErr != nil {
				beego.Error("CrashJcfkStats fail err: ", addErr)
			}
		}
	}
}

// 排行榜
func RankingList(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
	paylist := ComputePayRanking(startTime, endTime)
	for _, item := range paylist {
		item.RType = 0
		item.Date = timestamp
		item.ETime = utils.TimestampToday(location)
		addErr := StatisticsService.AddOrUpdateRankingList(&item)
		if addErr != nil {
			beego.Error("RankingList fail err: ", addErr)
		}
	}
	withdrawlist := ComputeWithDrawRanking(startTime, endTime)
	for _, item := range withdrawlist {
		item.RType = 1
		item.Date = timestamp
		item.ETime = utils.TimestampToday(location)
		addErr := StatisticsService.AddOrUpdateRankingList(&item)
		if addErr != nil {
			beego.Error("RankingList fail err: ", addErr)
		}
	}
	winlist := ComputeWinRanking(startTime, endTime)
	for _, item := range winlist {
		item.RType = 2
		item.Date = timestamp
		item.ETime = utils.TimestampToday(location)
		addErr := StatisticsService.AddOrUpdateRankingList(&item)
		if addErr != nil {
			beego.Error("RankingList fail err: ", addErr)
		}
	}
	loselist := GetLoseRanking(startTime, endTime)
	for _, item := range loselist {
		item.RType = 3
		item.Date = timestamp
		item.ETime = utils.TimestampToday(location)
		addErr := StatisticsService.AddOrUpdateRankingList(&item)
		if addErr != nil {
			beego.Error("RankingList fail err: ", addErr)
		}
	}
	paycarrylist := ComputePayCarryRanking(startTime, endTime)
	for _, item := range paycarrylist {
		item.RType = 4
		item.Date = timestamp
		item.ETime = utils.TimestampToday(location)
		addErr := StatisticsService.AddOrUpdateRankingList(&item)
		if addErr != nil {
			beego.Error("RankingList fail err: ", addErr)
		}
	}
	fzlist := ComputeFZCarryRanking(startTime, endTime)
	for _, item := range fzlist {
		item.RType = 5
		item.Date = timestamp
		item.ETime = utils.TimestampToday(location)
		addErr := StatisticsService.AddOrUpdateRankingList(&item)
		if addErr != nil {
			beego.Error("RankingList fail err: ", addErr)
		}
	}
}

// 排行榜-付费榜
func ComputePayRanking(startTime, endTime time.Time) []entity.RankingList {
	pay := make([]entity.RankingList, 0)
	payList := make(map[string]entity.RankingList, 0)
	// 付费榜
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"order_status": 4,
				"ctime":        bson.M{"$gte": startTime, "$lte": endTime},
			},
		},
		{
			"$group": bson.M{
				"_id":    "$userid",
				"amount": bson.M{"$sum": "$amount"},
			},
		},
		{"$sort": bson.M{"amount": -1}},
		{"$limit": 20},
	}
	var pay_res []bson.M
	Pays.Pipe(pipeline).All(&pay_res)
	var pay_ids []string
	for i, item := range pay_res {
		pid := item["_id"].(string)
		amount := item["amount"].(int)
		info := new(entity.RankingList)
		info.No = (i + 1)
		info.Userid = pid
		info.PayAmount = int64(amount)
		pay_ids = append(pay_ids, pid)
		payList[pid] = *info
	}
	// 获取提现
	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"order_status": 2,
				"ctime":        bson.M{"$gte": startTime, "$lte": endTime},
				"userid":       bson.M{"$in": pay_ids},
			},
		},
		{
			"$group": bson.M{
				"_id":    "$userid",
				"amount": bson.M{"$sum": "$amount"},
			},
		},
	}
	var withdraw_res []bson.M
	Withdraws.Pipe(pipeline1).All(&withdraw_res)
	for _, with := range withdraw_res {
		wid := with["_id"].(string)
		stat, ok := payList[wid]
		if ok {
			amount := with["amount"].(int)
			stat.WithdrawAmount = int64(amount)
			payList[wid] = stat
		}
	}
	// 获取用户类型
	pipeline2 := []bson.M{
		{
			"$match": bson.M{
				"robot":            false,
				"simulation_robot": false,
				"_id":              bson.M{"$in": pay_ids},
			},
		},
		{
			"$project": bson.M{
				"_id":         "$_id",
				"regist_area": "$regist_area",
			},
		},
	}
	var user_res []bson.M
	PlayerUsers.Pipe(pipeline2).All(&user_res)
	for _, xd := range user_res {
		uid := xd["_id"].(string)
		stat, ok := payList[uid]
		if ok {
			// diamond := xd["diamond"].(int64)
			// shadow_diamond := xd["shadow_diamond"].(int64)
			// coin := xd["coin"].(int64)
			regist_area := 0
			dsId, _ := xd["regist_area"]
			if dsId == nil {
				regist_area = 0
			} else {
				regist_area = dsId.(int)
			}
			// amount := diamond + shadow_diamond + coin
			// stat.CarryAmount = int64(amount)
			stat.RegistArea = regist_area
			// 计算盈利
			profit := (stat.WithdrawAmount + stat.CarryAmount) - stat.PayAmount
			stat.ProfitAmount = profit
			payList[uid] = stat
		}
	}

	// 获取用户携带金币
	pipeline3 := []bson.M{
		{
			"$match": bson.M{
				"ctime":  bson.M{"$gte": startTime, "$lte": endTime},
				"userid": bson.M{"$in": pay_ids},
			},
		},
		{
			"$sort": bson.M{"ctime": 1},
		},
		{
			"$group": bson.M{
				"_id":         "$userid",
				"now_diamond": bson.M{"$last": "$now_diamond"},
				"now_coin":    bson.M{"$last": "$now_coin"},
			},
		},
	}
	var xiedai_res []bson.M
	LogWaters.Pipe(pipeline3).All(&xiedai_res)
	for _, xd := range xiedai_res {
		uid := xd["_id"].(string)
		stat, ok := payList[uid]
		if ok {
			diamond := xd["now_diamond"].(int64)
			shadow_diamond := xd["now_coin"].(int64)
			amount := diamond + shadow_diamond
			stat.CarryAmount = int64(amount)
			// 计算盈利
			profit := (stat.WithdrawAmount + stat.CarryAmount) - stat.PayAmount
			stat.ProfitAmount = profit
			payList[uid] = stat
		}
	}
	for _, v := range payList {
		pay = append(pay, v)
	}
	sort.Slice(pay, func(i, j int) bool {
		return pay[i].Id < pay[j].Id
	})
	return pay
}

// 排行榜-提现榜
func ComputeWithDrawRanking(startTime, endTime time.Time) []entity.RankingList {
	withdraw := make([]entity.RankingList, 0)
	withdrawList := make(map[string]entity.RankingList, 0)
	// 提现榜
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"ctime":        bson.M{"$gte": startTime, "$lte": endTime},
				"order_status": 2,
			},
		},
		{
			"$group": bson.M{
				"_id":    "$userid",
				"amount": bson.M{"$sum": "$amount"},
			},
		},
		{"$sort": bson.M{"amount": -1}},
		{"$limit": 20},
	}
	var pay_res []bson.M
	Withdraws.Pipe(pipeline).All(&pay_res)
	var pay_ids []string
	for i, item := range pay_res {
		pid := item["_id"].(string)
		amount := item["amount"].(int)
		info := new(entity.RankingList)
		info.No = (i + 1)
		info.Userid = pid
		info.WithdrawAmount = int64(amount)
		pay_ids = append(pay_ids, pid)
		withdrawList[pid] = *info
	}
	// 获取充值
	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"order_status": 4,
				"ctime":        bson.M{"$gte": startTime, "$lte": endTime},
				"userid":       bson.M{"$in": pay_ids},
			},
		},
		{
			"$group": bson.M{
				"_id":    "$userid",
				"amount": bson.M{"$sum": "$amount"},
			},
		},
	}
	var withdraw_res []bson.M
	Pays.Pipe(pipeline1).All(&withdraw_res)
	for _, with := range withdraw_res {
		wid := with["_id"].(string)
		stat, ok := withdrawList[wid]
		if ok {
			amount := with["amount"].(int)
			stat.PayAmount = int64(amount)
			withdrawList[wid] = stat
		}

	}
	// 获取用户类型
	pipeline2 := []bson.M{
		{
			"$match": bson.M{
				"robot":            false,
				"simulation_robot": false,
				"_id":              bson.M{"$in": pay_ids},
			},
		},
		{
			"$project": bson.M{
				"_id":         "$_id",
				"regist_area": "$regist_area",
			},
		},
	}
	var user_res []bson.M
	PlayerUsers.Pipe(pipeline2).All(&user_res)
	for _, xd := range user_res {
		uid := xd["_id"].(string)
		stat, ok := withdrawList[uid]
		if ok {
			// diamond := xd["diamond"].(int64)
			// shadow_diamond := xd["shadow_diamond"].(int64)
			// coin := xd["coin"].(int64)
			regist_area := 0
			dsId, _ := xd["regist_area"]
			if dsId == nil {
				regist_area = 0
			} else {
				regist_area = dsId.(int)
			}
			// amount := diamond + shadow_diamond + coin
			// stat.CarryAmount = int64(amount)
			stat.RegistArea = regist_area
			// 计算盈利
			profit := (stat.WithdrawAmount + stat.CarryAmount) - stat.PayAmount
			stat.ProfitAmount = profit
			withdrawList[uid] = stat
		}
	}

	// 获取用户携带金币
	pipeline3 := []bson.M{
		{
			"$match": bson.M{
				"ctime":  bson.M{"$gte": startTime, "$lte": endTime},
				"userid": bson.M{"$in": pay_ids},
			},
		},
		{
			"$sort": bson.M{"ctime": 1},
		},
		{
			"$group": bson.M{
				"_id":         "$userid",
				"now_diamond": bson.M{"$last": "$now_diamond"},
				"now_coin":    bson.M{"$last": "$now_coin"},
			},
		},
	}
	var xiedai_res []bson.M
	LogWaters.Pipe(pipeline3).All(&xiedai_res)
	for _, xd := range xiedai_res {
		uid := xd["_id"].(string)
		stat, ok := withdrawList[uid]
		if ok {
			diamond := xd["now_diamond"].(int64)
			shadow_diamond := xd["now_coin"].(int64)
			amount := diamond + shadow_diamond
			stat.CarryAmount = int64(amount)
			// 计算盈利
			profit := (stat.WithdrawAmount + stat.CarryAmount) - stat.PayAmount
			stat.ProfitAmount = profit
			withdrawList[uid] = stat
		}
	}

	for _, v := range withdrawList {
		withdraw = append(withdraw, v)
	}
	sort.Slice(withdraw, func(i, j int) bool {
		return withdraw[i].Id < withdraw[j].Id
	})
	return withdraw
}

// 排行榜-赢榜
func ComputeWinRanking(startTime, endTime time.Time) []entity.RankingList {
	list := make([]entity.RankingList, 0)
	map_list := make(map[string]entity.RankingList, 0)

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": startTime, "$lte": endTime},
				"ltype": bson.M{"$in": []int{110, 111, 75, 67, 69, 76, 70, 72, 116, 117, 114, 115, 112, 113, 79, 73, 86, 88, 89, 90, 91, 92, 100, 101, 93, 94, 95, 105}},
			},
		},
		{
			"$group": bson.M{
				"_id":             "$userid",
				"add_diamond":     bson.M{"$sum": "$add_diamond"},
				"add_coin":        bson.M{"$sum": "$add_coin"},
				"add_other_asset": bson.M{"$sum": "$add_other_asset"},
			},
		},
		{
			"$project": bson.M{
				"total": bson.M{"$add": []interface{}{"$add_diamond", "$add_coin", "$add_other_asset"}},
			},
		},
		{
			"$sort": bson.M{"total": -1},
		},
		{
			"$limit": 20,
		},
	}
	var xiedai_res []bson.M
	var s_ids []string
	LogWaters.Pipe(pipeline).All(&xiedai_res)
	for i, xd := range xiedai_res {
		uid := xd["_id"].(string)
		amount := xd["total"].(int64)
		info := new(entity.RankingList)
		info.No = (i + 1)
		info.Userid = uid
		info.CarryAmount = int64(amount)
		s_ids = append(s_ids, uid)
		map_list[uid] = *info
	}
	// 获取充值
	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"order_status": 4,
				"ctime":        bson.M{"$gte": startTime, "$lte": endTime},
				"userid":       bson.M{"$in": s_ids},
			},
		},
		{
			"$group": bson.M{
				"_id":    "$userid",
				"amount": bson.M{"$sum": "$amount"},
			},
		},
	}
	var pay_res []bson.M
	Pays.Pipe(pipeline1).All(&pay_res)
	for _, with := range pay_res {
		wid := with["_id"].(string)
		stat, ok := map_list[wid]
		if ok {
			amount := with["amount"].(int)
			stat.PayAmount = int64(amount)
			map_list[wid] = stat
		}
	}

	// 获取提现
	pipeline2 := []bson.M{
		{
			"$match": bson.M{
				"order_status": 2,
				"ctime":        bson.M{"$gte": startTime, "$lte": endTime},
				"userid":       bson.M{"$in": s_ids},
			},
		},
		{
			"$group": bson.M{
				"_id":    "$userid",
				"amount": bson.M{"$sum": "$amount"},
			},
		},
	}
	var with_res []bson.M
	Withdraws.Pipe(pipeline2).All(&with_res)
	for _, with := range with_res {
		wid := with["_id"].(string)
		stat, ok := map_list[wid]
		if ok {
			amount := with["amount"].(int)
			stat.WithdrawAmount = int64(amount)
			map_list[wid] = stat
		}
	}
	// 获取用户类型
	pipeline3 := []bson.M{
		{
			"$match": bson.M{
				"robot":            false,
				"simulation_robot": false,
				"_id":              bson.M{"$in": s_ids},
			},
		},
		{
			"$project": bson.M{
				"_id":         "$_id",
				"regist_area": "$regist_area",
			},
		},
	}
	var user_res []bson.M
	PlayerUsers.Pipe(pipeline3).All(&user_res)
	for _, xd := range user_res {
		uid := xd["_id"].(string)
		stat, ok := map_list[uid]
		if ok {
			regist_area := 0
			dsId, _ := xd["regist_area"]
			if dsId == nil {
				regist_area = 0
			} else {
				regist_area = dsId.(int)
			}
			stat.RegistArea = regist_area
			// 计算盈利
			profit := (stat.WithdrawAmount + stat.CarryAmount) - stat.PayAmount
			stat.ProfitAmount = profit
			map_list[uid] = stat
		}
	}
	for _, v := range map_list {
		list = append(list, v)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Id < list[j].Id
	})
	return list
}

// 排行榜-输榜
func GetLoseRanking(startTime, endTime time.Time) []entity.RankingList {
	list := make([]entity.RankingList, 0)
	map_list := make(map[string]entity.RankingList, 0)
	// 输
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": startTime, "$lte": endTime},
				"ltype": bson.M{"$in": []int{110, 111, 75, 67, 69, 76, 70, 72, 116, 117, 114, 115, 112, 113, 79, 73, 86, 88, 89, 90, 91, 92, 100, 101, 93, 94, 95, 105}},
			},
		},
		{
			"$group": bson.M{
				"_id":             "$userid",
				"add_diamond":     bson.M{"$sum": "$add_diamond"},
				"add_coin":        bson.M{"$sum": "$add_coin"},
				"add_other_asset": bson.M{"$sum": "$add_other_asset"},
			},
		},
		{
			"$project": bson.M{
				"total": bson.M{"$add": []interface{}{"$add_diamond", "$add_coin", "$add_other_asset"}},
			},
		},
		{
			"$sort": bson.M{"total": 1},
		},
		{
			"$limit": 20,
		},
	}
	var xiedai_res []bson.M
	var s_ids []string
	LogWaters.Pipe(pipeline).All(&xiedai_res)
	for i, xd := range xiedai_res {
		uid := xd["_id"].(string)
		amount := xd["total"].(int64)
		info := new(entity.RankingList)
		info.No = (i + 1)
		info.Userid = uid
		info.CarryAmount = int64(amount)
		s_ids = append(s_ids, uid)
		map_list[uid] = *info
	}
	// 获取充值
	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"order_status": 4,
				"ctime":        bson.M{"$gte": startTime, "$lte": endTime},
				"userid":       bson.M{"$in": s_ids},
			},
		},
		{
			"$group": bson.M{
				"_id":    "$userid",
				"amount": bson.M{"$sum": "$amount"},
			},
		},
	}
	var pay_res []bson.M
	Pays.Pipe(pipeline1).All(&pay_res)
	for _, with := range pay_res {
		wid := with["_id"].(string)
		stat, ok := map_list[wid]
		if ok {
			amount := with["amount"].(int)
			stat.PayAmount = int64(amount)
			map_list[wid] = stat
		}
	}

	// 获取提现
	pipeline2 := []bson.M{
		{
			"$match": bson.M{
				"order_status": 2,
				"ctime":        bson.M{"$gte": startTime, "$lte": endTime},
				"userid":       bson.M{"$in": s_ids},
			},
		},
		{
			"$group": bson.M{
				"_id":    "$userid",
				"amount": bson.M{"$sum": "$amount"},
			},
		},
	}
	var with_res []bson.M
	Withdraws.Pipe(pipeline2).All(&with_res)
	for _, with := range with_res {
		wid := with["_id"].(string)
		stat, ok := map_list[wid]
		if ok {
			amount := with["amount"].(int)
			stat.WithdrawAmount = int64(amount)
			map_list[wid] = stat
		}
	}
	// 获取用户类型
	pipeline3 := []bson.M{
		{
			"$match": bson.M{
				"robot":            false,
				"simulation_robot": false,
				"_id":              bson.M{"$in": s_ids},
			},
		},
		{
			"$project": bson.M{
				"_id":         "$_id",
				"regist_area": "$regist_area",
			},
		},
	}
	var user_res []bson.M
	PlayerUsers.Pipe(pipeline3).All(&user_res)
	for _, xd := range user_res {
		uid := xd["_id"].(string)
		stat, ok := map_list[uid]
		if ok {
			regist_area := 0
			dsId, _ := xd["regist_area"]
			if dsId == nil {
				regist_area = 0
			} else {
				regist_area = dsId.(int)
			}
			stat.RegistArea = regist_area
			// 计算盈利
			profit := (stat.WithdrawAmount + stat.CarryAmount) - stat.PayAmount
			stat.ProfitAmount = profit
			map_list[uid] = stat
		}
	}
	for _, v := range map_list {
		list = append(list, v)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Id < list[j].Id
	})
	return list
}

// 排行榜-付费携带榜
func ComputePayCarryRanking(startTime, endTime time.Time) []entity.RankingList {
	list := make([]entity.RankingList, 0)
	map_list := make(map[string]entity.RankingList, 0)
	// 获取存在充值的玩家
	n := bson.M{}
	n["ctime"] = bson.M{"$gte": startTime, "$lte": endTime}
	n["order_status"] = 4
	var pay_ids []string
	Pays.Find(n).Distinct("userid", &pay_ids)

	// 获取用户携带金币
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"ctime":  bson.M{"$gte": startTime, "$lte": endTime},
				"userid": bson.M{"$in": pay_ids},
			},
		},
		{
			"$sort": bson.M{"ctime": 1},
		},
		{
			"$group": bson.M{
				"_id":         "$userid",
				"now_diamond": bson.M{"$last": "$now_diamond"},
				"now_coin":    bson.M{"$last": "$now_coin"},
			},
		},
		{
			"$project": bson.M{
				"total": bson.M{"$add": []interface{}{"$now_diamond", "$now_coin"}},
			},
		},
		{
			"$sort": bson.M{"total": -1},
		},
		{
			"$limit": 20,
		},
	}
	var xiedai_res []bson.M
	var s_ids []string
	LogWaters.Pipe(pipeline).All(&xiedai_res)
	for i, xd := range xiedai_res {
		uid := xd["_id"].(string)
		amount := xd["total"].(int64)
		info := new(entity.RankingList)
		info.No = (i + 1)
		info.Userid = uid
		info.CarryAmount = int64(amount)
		s_ids = append(s_ids, uid)
		map_list[uid] = *info
	}

	// 获取充值
	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"order_status": 4,
				"ctime":        bson.M{"$gte": startTime, "$lte": endTime},
				"userid":       bson.M{"$in": s_ids},
			},
		},
		{
			"$group": bson.M{
				"_id":    "$userid",
				"amount": bson.M{"$sum": "$amount"},
			},
		},
	}
	var pay_res []bson.M
	Pays.Pipe(pipeline1).All(&pay_res)
	for _, with := range pay_res {
		wid := with["_id"].(string)
		stat, ok := map_list[wid]
		if ok {
			amount := with["amount"].(int)
			stat.PayAmount = int64(amount)
			map_list[wid] = stat
		}
	}

	// 获取提现
	pipeline2 := []bson.M{
		{
			"$match": bson.M{
				"order_status": 2,
				"ctime":        bson.M{"$gte": startTime, "$lte": endTime},
				"userid":       bson.M{"$in": s_ids},
			},
		},
		{
			"$group": bson.M{
				"_id":    "$userid",
				"amount": bson.M{"$sum": "$amount"},
			},
		},
	}
	var with_res []bson.M
	Withdraws.Pipe(pipeline2).All(&with_res)
	for _, with := range with_res {
		wid := with["_id"].(string)
		stat, ok := map_list[wid]
		if ok {
			amount := with["amount"].(int)
			stat.WithdrawAmount = int64(amount)
			map_list[wid] = stat
		}
	}
	// 获取用户类型
	pipeline3 := []bson.M{
		{
			"$match": bson.M{
				"robot":            false,
				"simulation_robot": false,
				"_id":              bson.M{"$in": s_ids},
			},
		},
		{
			"$project": bson.M{
				"_id":         "$_id",
				"regist_area": "$regist_area",
			},
		},
	}
	var user_res []bson.M
	PlayerUsers.Pipe(pipeline3).All(&user_res)
	for _, xd := range user_res {
		uid := xd["_id"].(string)
		stat, ok := map_list[uid]
		if ok {
			regist_area := 0
			dsId, _ := xd["regist_area"]
			if dsId == nil {
				regist_area = 0
			} else {
				regist_area = dsId.(int)
			}
			stat.RegistArea = regist_area
			// 计算盈利
			profit := (stat.WithdrawAmount + stat.CarryAmount) - stat.PayAmount
			stat.ProfitAmount = profit
			map_list[uid] = stat
		}
	}
	for _, v := range map_list {
		list = append(list, v)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Id < list[j].Id
	})
	return list
}

// 排行榜-峰值携带榜
func ComputeFZCarryRanking(startTime, endTime time.Time) []entity.RankingList {
	list := make([]entity.RankingList, 0)
	map_list := make(map[string]entity.RankingList, 0)
	// 当天存在登录的玩家
	n := bson.M{}
	n["login_time"] = bson.M{"$gte": startTime, "$lte": endTime}
	var log_ids []string
	LoginLogs.Find(n).Distinct("userid", &log_ids)
	// 获取用户携带金币
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"ctime":  bson.M{"$gte": startTime, "$lte": endTime},
				"userid": bson.M{"$in": log_ids},
			},
		},
		{
			"$sort": bson.M{"ctime": 1},
		},
		{
			"$group": bson.M{
				"_id":         "$userid",
				"now_diamond": bson.M{"$last": "$now_diamond"},
				"now_coin":    bson.M{"$last": "$now_coin"},
			},
		},
		{
			"$project": bson.M{
				"total": bson.M{"$add": []interface{}{"$now_diamond", "$now_coin"}},
			},
		},
		{
			"$sort": bson.M{"total": -1},
		},
		{
			"$limit": 20,
		},
	}
	var xiedai_res []bson.M
	var s_ids []string
	LogWaters.Pipe(pipeline).All(&xiedai_res)
	for i, xd := range xiedai_res {
		uid := xd["_id"].(string)
		amount := xd["total"].(int64)
		info := new(entity.RankingList)
		info.No = (i + 1)
		info.Userid = uid
		info.CarryAmount = int64(amount)
		s_ids = append(s_ids, uid)
		map_list[uid] = *info
	}
	// 获取充值
	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"order_status": 4,
				"ctime":        bson.M{"$gte": startTime, "$lte": endTime},
				"userid":       bson.M{"$in": s_ids},
			},
		},
		{
			"$group": bson.M{
				"_id":    "$userid",
				"amount": bson.M{"$sum": "$amount"},
			},
		},
	}
	var pay_res []bson.M
	Pays.Pipe(pipeline1).All(&pay_res)
	for _, with := range pay_res {
		wid := with["_id"].(string)
		stat, ok := map_list[wid]
		if ok {
			amount := with["amount"].(int)
			stat.PayAmount = int64(amount)
			map_list[wid] = stat
		}
	}

	// 获取提现
	pipeline2 := []bson.M{
		{
			"$match": bson.M{
				"order_status": 2,
				"ctime":        bson.M{"$gte": startTime, "$lte": endTime},
				"userid":       bson.M{"$in": s_ids},
			},
		},
		{
			"$group": bson.M{
				"_id":    "$userid",
				"amount": bson.M{"$sum": "$amount"},
			},
		},
	}
	var with_res []bson.M
	Withdraws.Pipe(pipeline2).All(&with_res)
	for _, with := range with_res {
		wid := with["_id"].(string)
		stat, ok := map_list[wid]
		if ok {
			amount := with["amount"].(int)
			stat.WithdrawAmount = int64(amount)
			map_list[wid] = stat
		}
	}
	// 获取用户类型
	pipeline3 := []bson.M{
		{
			"$match": bson.M{
				"robot":            false,
				"simulation_robot": false,
				"_id":              bson.M{"$in": s_ids},
			},
		},
		{
			"$project": bson.M{
				"_id":         "$_id",
				"regist_area": "$regist_area",
			},
		},
	}
	var user_res []bson.M
	PlayerUsers.Pipe(pipeline3).All(&user_res)
	for _, xd := range user_res {
		uid := xd["_id"].(string)
		stat, ok := map_list[uid]
		if ok {
			regist_area := 0
			dsId, _ := xd["regist_area"]
			if dsId == nil {
				regist_area = 0
			} else {
				regist_area = dsId.(int)
			}
			stat.RegistArea = regist_area
			// 计算盈利
			profit := (stat.WithdrawAmount + stat.CarryAmount) - stat.PayAmount
			stat.ProfitAmount = profit
			map_list[uid] = stat
		}
	}
	for _, v := range map_list {
		list = append(list, v)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Id < list[j].Id
	})
	return list
}

// TPStatsCK TP游戏统计
func TPStatsCK(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
	date1, _ := utils.Unix(today)

	// 查询对局
	sql := `
	SELECT userid, win_type, count(*) rounds, SUM(end_time - begin_time) game_times,
		SUM(bet_amount) bet_amount, SUM(settle_score) settle_score 
	FROM game.col_detail FINAL 
	WHERE robot = 0 AND gtype = 1 AND begin_time between ? and ?
	GROUP BY userid, win_type
	order by userid, win_type
	`
	var datas []map[string]any
	err := ck.Select(&datas, sql, startTime.Unix(), endTime.Unix())
	if err != nil {
		beego.Warning("no tp detail data", err)
		return
	}
	if len(datas) == 0 {
		return
	}

	// 查询当天注册用户
	var regUserids = make(map[string]bool)
	var regUseridsR []string
	err = ck.Select(&regUseridsR, `select userid from game.col_user final where ctime between ? and ?`, startTime, endTime)
	if err != nil {
		beego.Error("regUsers error: ", err)
		return
	}
	for _, userid := range regUseridsR {
		regUserids[userid] = true
	}

	// tp玩家列表
	var tpPlayerIds []string
	var tpPlayers = make(map[string]*entity.TPPlayerStat)
	for _, data := range datas {
		userid := data["userid"].(string)
		winType := int8(utils.ToInt64(data["win_type"]))
		rounds := utils.ToInt64(data["rounds"])
		game_times := utils.ToInt64(data["game_times"])
		bet_amount := utils.ToInt64(data["bet_amount"])
		settle_score := utils.ToInt64(data["settle_score"])

		tpPlayerIds = append(tpPlayerIds, userid)
		stat, ok := tpPlayers[userid]
		if !ok {
			stat = &entity.TPPlayerStat{
				Id:      fmt.Sprintf("%d-%s", date1, userid),
				Userid:  userid,
				Date:    date1,
				DateStr: today,
				NewReg:  regUserids[userid], // 老顾客新顾客
			}
			tpPlayers[userid] = stat
		}
		stat.Bets += bet_amount
		stat.AllRounds += int32(rounds)
		stat.GameTimes += game_times

		switch winType {
		case ck.WinTypeWin:
			stat.BetRounds += int32(rounds)
			stat.WinRounds += int32(rounds)
			stat.WinBets += bet_amount
			stat.Wins += settle_score
			stat.Cash += (settle_score)
		case ck.WinTypeLose:
			stat.BetRounds += int32(rounds)
			stat.LoseRounds += int32(rounds)
			stat.LoseBets += bet_amount
			stat.Loses += settle_score
			stat.Cash += (settle_score)
		case ck.WinTypeTie:
			stat.BetRounds += int32(rounds)
			stat.TieRounds += int32(rounds)
			stat.TieBets += bet_amount
		case ck.WinTypeObserver:
			stat.ObserveRounds += int32(rounds)
		}
	}

	// 查询渠道，账号类型，充值金额
	var users []ck.User
	err = ck.Select(&users, `select userid,ad__bundle_id,channel,regist_area,money from col_user final where userid in ?`, tpPlayerIds)
	if err != nil {
		beego.Error("regUsers error: ", err)
		return
	}
	for _, user := range users {
		stat, ok := tpPlayers[user.Userid]
		if !ok {
			continue
		}
		stat.AD_BundleId = user.AD_BundleId
		stat.Channel1 = user.Channel
		stat.RegistArea = user.RegistArea
		stat.Money = user.Money
	}

	// 保存统计数据
	for _, p := range tpPlayers {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)

		// 保存数据
		addErr := StatisticsService.AddOrUpdateTpPlayerStat(p)
		if addErr != nil {
			beego.Error("TPStats fail err: ", addErr)
		}
	}
}

// TP房间统计
func TPRoomStatsCK(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
	date1, _ := utils.Unix(today)
	_ = date1

	var datas []map[string]any
	err := ck.Select(&datas, `
		SELECT room_id, userid, win_type, player_num, tp_hand_type_up_down, count(*) rounds, SUM(bet_amount) bet_amount, SUM(settle_score) settle_score,
			SUM(tp_pack) tp_pack, SUM(tp_call) tp_call, SUM(tp_room_up) tp_room_up,SUM(tp_room_down) tp_room_down,SUM(tp_big_win) tp_big_win,SUM(tp_big_lose) tp_big_lose,
			SUM(end_time - begin_time) game_times, SUM(tp_turns) tp_turns
		FROM game.col_detail FINAL 
		WHERE robot = 0 AND gtype = 1 AND begin_time between ? and ?
		GROUP BY room_id, userid, win_type, player_num, tp_hand_type_up_down 
		order by room_id, userid
	`, startTime.Unix(), endTime.Unix())
	if err != nil {
		beego.Error("TPRoomStatsCK error: ", err)
		return
	}

	// 房间数据
	var roomlist []entity.Game
	var rooms = make(map[string]entity.Game)
	m := bson.M{"status": 1}
	m["gtype"] = 1
	m["ai_status"] = 1
	Games.Find(m).All(&roomlist)
	for _, room := range roomlist {
		rooms[room.Id] = room
	}

	var userids []string
	var roomUsers = make(map[string]map[string]*entity.TPRoomStat)
	for _, data := range datas {
		room_id := data["room_id"].(string)
		userid := data["userid"].(string)
		win_type := int8(utils.ToInt64(data["win_type"]))
		player_num := int64(utils.ToInt64(data["player_num"]))
		tp_hand_type_up_down := int64(utils.ToInt64(data["tp_hand_type_up_down"]))
		rounds := utils.ToInt64(data["rounds"])
		bet_amount := utils.ToInt64(data["bet_amount"])
		settle_score := utils.ToInt64(data["settle_score"])
		tp_pack := utils.ToInt64(data["tp_pack"])
		tp_call := utils.ToInt64(data["tp_call"])
		game_times := utils.ToInt64(data["game_times"])
		tp_room_up := utils.ToInt64(data["tp_room_up"])
		tp_room_down := utils.ToInt64(data["tp_room_down"])
		tp_big_win := utils.ToInt64(data["tp_big_win"])
		tp_big_lose := utils.ToInt64(data["tp_big_lose"])
		tp_turns := utils.ToInt64(data["tp_turns"])

		userids = append(userids, userid)
		room, ok := roomUsers[room_id]
		if !ok {
			room = make(map[string]*entity.TPRoomStat)
			roomUsers[room_id] = room
		}
		stat, ok := room[userid]
		if !ok {
			stat = &entity.TPRoomStat{
				Id:      fmt.Sprintf("%d-%s-%s", date1, userid, room_id),
				Date:    date1,
				DateStr: today,
				Userid:  userid,
				Roomid:  room_id,
			}
			if room, ok := rooms[room_id]; ok {
				stat.ChipPool = int64(room.TP.Pool_Limit)
			}
			room[userid] = stat
		}
		stat.AllRounds += rounds
		stat.Bets += bet_amount
		stat.GameTime += game_times
		stat.UpNumber += tp_room_up
		stat.DownNumber += tp_room_down

		switch win_type {
		case ck.WinTypeWin:
			stat.WinRounds += int32(rounds)
			stat.WinBets += bet_amount
		case ck.WinTypeLose:
			stat.LoseRounds += int32(rounds)
			stat.LoseBets += bet_amount
		case ck.WinTypeTie:
		}

		switch tp_hand_type_up_down {
		case algo.BaoziUp:
			stat.BZUpBets += bet_amount
			stat.BZUpNumber += rounds
			stat.BZUpRounds += tp_turns
			stat.BZUpWinRounds += tp_big_win
			stat.BZUpLoseRounds += tp_big_lose
		case algo.BaoziDown:
			stat.BZDownBets += bet_amount
			stat.BZDownNumber += rounds
			stat.BZDownRounds += tp_turns
			stat.BZDownWinRounds += tp_big_win
			stat.BZDownLoseRounds += tp_big_lose
		case algo.TonghuashunUp:
			stat.THSUpBets += bet_amount
			stat.THSUpNumber += rounds
			stat.THSUpRounds += tp_turns
			stat.THSUpWinRounds += tp_big_win
			stat.THSUpLoseRounds += tp_big_lose
		case algo.TonghuashunDown:
			stat.THSDownBets += bet_amount
			stat.THSDownNumber += rounds
			stat.THSDownRounds += tp_turns
			stat.THSDownWinRounds += tp_big_win
			stat.THSDownLoseRounds += tp_big_lose
		case algo.ShunziUp:
			stat.DSZUpBets += bet_amount
			stat.DSZUpNumber += rounds
			stat.DSZUpRounds += tp_turns
			stat.DSZUpWinRounds += tp_big_win
			stat.DSZUpLoseRounds += tp_big_lose
		case algo.ShunziDown:
			stat.DSZDownBets += bet_amount
			stat.DSZDownNumber += rounds
			stat.DSZDownRounds += tp_turns
			stat.DSZDownWinRounds += tp_big_win
			stat.DSZDownLoseRounds += tp_big_lose
		case algo.TonghuaUp:
			stat.DTHUpBets += bet_amount
			stat.DTHUpNumber += rounds
			stat.DTHUpRounds += tp_turns
			stat.DTHUpWinRounds += tp_big_win
			stat.DTHUpLoseRounds += tp_big_lose
		case algo.TonghuaDown:
			stat.DTHDownBets += bet_amount
			stat.DTHDownNumber += rounds
			stat.DTHDownRounds += tp_turns
			stat.DTHDownWinRounds += tp_big_win
			stat.DTHDownLoseRounds += tp_big_lose
		case algo.DuiziUp:
			stat.DDZUpBets += bet_amount
			stat.DDZUpNumber += rounds
			stat.DDZUpRounds += tp_turns
			stat.DDZUpWinRounds += tp_big_win
			stat.DDZUpLoseRounds += tp_big_lose
		case algo.DuiziDown:
			stat.DDZDownBets += bet_amount
			stat.DDZDownNumber += rounds
			stat.DDZDownRounds += tp_turns
			stat.DDZDownWinRounds += tp_big_win
			stat.DDZDownLoseRounds += tp_big_lose
		case algo.GaopaiUp:
			stat.DGPUpBets += bet_amount
			stat.DGPUpNumber += rounds
			stat.DGPUpRounds += tp_turns
			stat.DGPUpWinRounds += tp_big_win
			stat.DGPUpLoseRounds += tp_big_lose
		case algo.GaopaiDown:
			stat.DGPDownBets += bet_amount
			stat.DGPDownNumber += rounds
			stat.DGPDownRounds += tp_turns
			stat.DGPDownWinRounds += tp_big_win
			stat.DGPDownLoseRounds += tp_big_lose
		}

		switch player_num {
		case 2: // 2人局
			stat.TwoRounds += rounds
			stat.TwoBets += bet_amount
			if win_type == ck.WinTypeWin {
				stat.WinTwoRounds += rounds
				stat.WinTwoBets += settle_score
			}
			switch tp_hand_type_up_down {
			case algo.BaoziUp:
				stat.BZUpTwoFollowNum += tp_call
				stat.BZUpTwoDiscardNum += tp_pack
			case algo.BaoziDown:
				stat.BZDownTwoFollowNum += tp_call
				stat.BZDownTwoDiscardNum += tp_pack
			case algo.TonghuashunUp:
				stat.THSUpTwoFollowNum += tp_call
				stat.THSUpTwoDiscardNum += tp_pack
			case algo.TonghuashunDown:
				stat.THSDownTwoFollowNum += tp_call
				stat.THSDownTwoDiscardNum += tp_pack
			case algo.ShunziUp:
				stat.DSZUpTwoFollowNum += tp_call
				stat.DSZUpTwoDiscardNum += tp_pack
			case algo.ShunziDown:
				stat.DSZDownTwoFollowNum += tp_call
				stat.DSZDownTwoDiscardNum += tp_pack
			case algo.TonghuaUp:
				stat.DTHUpTwoFollowNum += tp_call
				stat.DTHUpTwoDiscardNum += tp_pack
			case algo.TonghuaDown:
				stat.DTHDownTwoFollowNum += tp_call
				stat.DTHDownTwoDiscardNum += tp_pack
			case algo.DuiziUp:
				stat.DDZUpTwoFollowNum += tp_call
				stat.DDZUpTwoDiscardNum += tp_pack
			case algo.DuiziDown:
				stat.DDZDownTwoFollowNum += tp_call
				stat.DDZDownTwoDiscardNum += tp_pack
			case algo.GaopaiUp:
				stat.DGPUpTwoFollowNum += tp_call
				stat.DGPUpTwoDiscardNum += tp_pack
			case algo.GaopaiDown:
				stat.DGPDownTwoFollowNum += tp_call
				stat.DGPDownTwoDiscardNum += tp_pack
			}
		case 3:
			stat.ThreeRounds += rounds
			stat.ThreeBets += bet_amount
			if win_type == ck.WinTypeWin {
				stat.WinThreeRounds += rounds
				stat.WinThreeBets += settle_score
			}
			switch tp_hand_type_up_down {
			case algo.BaoziUp:
				stat.BZUpThreeFollowNum += tp_call
				stat.BZUpThreeDiscardNum += tp_pack
			case algo.BaoziDown:
				stat.BZDownThreeFollowNum += tp_call
				stat.BZDownThreeDiscardNum += tp_pack
			case algo.TonghuashunUp:
				stat.THSUpThreeFollowNum += tp_call
				stat.THSUpThreeDiscardNum += tp_pack
			case algo.TonghuashunDown:
				stat.THSDownThreeFollowNum += tp_call
				stat.THSDownThreeDiscardNum += tp_pack
			case algo.ShunziUp:
				stat.DSZUpThreeFollowNum += tp_call
				stat.DSZUpThreeDiscardNum += tp_pack
			case algo.ShunziDown:
				stat.DSZDownThreeFollowNum += tp_call
				stat.DSZDownThreeDiscardNum += tp_pack
			case algo.TonghuaUp:
				stat.DTHUpThreeFollowNum += tp_call
				stat.DTHUpThreeDiscardNum += tp_pack
			case algo.TonghuaDown:
				stat.DTHDownThreeFollowNum += tp_call
				stat.DTHDownThreeDiscardNum += tp_pack
			case algo.DuiziUp:
				stat.DDZUpThreeFollowNum += tp_call
				stat.DDZUpThreeDiscardNum += tp_pack
			case algo.DuiziDown:
				stat.DDZDownThreeFollowNum += tp_call
				stat.DDZDownThreeDiscardNum += tp_pack
			case algo.GaopaiUp:
				stat.DGPUpThreeFollowNum += tp_call
				stat.DGPUpThreeDiscardNum += tp_pack
			case algo.GaopaiDown:
				stat.DGPDownThreeFollowNum += tp_call
				stat.DGPDownThreeDiscardNum += tp_pack
			}
		case 4:
			stat.FourRounds += rounds
			stat.FourBets += bet_amount
			if win_type == ck.WinTypeWin {
				stat.WinFourRounds += rounds
				stat.WinFourBets += settle_score
			}
			switch tp_hand_type_up_down {
			case algo.BaoziUp:
				stat.BZUpFourFollowNum += tp_call
				stat.BZUpFourDiscardNum += tp_pack
			case algo.BaoziDown:
				stat.BZDownFourFollowNum += tp_call
				stat.BZDownFourDiscardNum += tp_pack
			case algo.TonghuashunUp:
				stat.THSUpFourFollowNum += tp_call
				stat.THSUpFourDiscardNum += tp_pack
			case algo.TonghuashunDown:
				stat.THSDownFourFollowNum += tp_call
				stat.THSDownFourDiscardNum += tp_pack
			case algo.ShunziUp:
				stat.DSZUpFourFollowNum += tp_call
				stat.DSZUpFourDiscardNum += tp_pack
			case algo.ShunziDown:
				stat.DSZDownFourFollowNum += tp_call
				stat.DSZDownFourDiscardNum += tp_pack
			case algo.TonghuaUp:
				stat.DTHUpFourFollowNum += tp_call
				stat.DTHUpFourDiscardNum += tp_pack
			case algo.TonghuaDown:
				stat.DTHDownFourFollowNum += tp_call
				stat.DTHDownFourDiscardNum += tp_pack
			case algo.DuiziUp:
				stat.DDZUpFourFollowNum += tp_call
				stat.DDZUpFourDiscardNum += tp_pack
			case algo.DuiziDown:
				stat.DDZDownFourFollowNum += tp_call
				stat.DDZDownFourDiscardNum += tp_pack
			case algo.GaopaiUp:
				stat.DGPUpFourFollowNum += tp_call
				stat.DGPUpFourDiscardNum += tp_pack
			case algo.GaopaiDown:
				stat.DGPDownFourFollowNum += tp_call
				stat.DGPDownFourDiscardNum += tp_pack
			}
		case 5:
			stat.FiveRounds += rounds
			stat.FiveBets += bet_amount
			if win_type == ck.WinTypeWin {
				stat.WinFiveRounds += rounds
				stat.WinFiveBets += settle_score
			}
			switch tp_hand_type_up_down {
			case algo.BaoziUp:
				stat.BZUpFiveFollowNum += tp_call
				stat.BZUpFiveDiscardNum += tp_pack
			case algo.BaoziDown:
				stat.BZDownFiveFollowNum += tp_call
				stat.BZDownFiveDiscardNum += tp_pack
			case algo.TonghuashunUp:
				stat.THSUpFiveFollowNum += tp_call
				stat.THSUpFiveDiscardNum += tp_pack
			case algo.TonghuashunDown:
				stat.THSDownFiveFollowNum += tp_call
				stat.THSDownFiveDiscardNum += tp_pack
			case algo.ShunziUp:
				stat.DSZUpFiveFollowNum += tp_call
				stat.DSZUpFiveDiscardNum += tp_pack
			case algo.ShunziDown:
				stat.DSZDownFiveFollowNum += tp_call
				stat.DSZDownFiveDiscardNum += tp_pack
			case algo.TonghuaUp:
				stat.DTHUpFiveFollowNum += tp_call
				stat.DTHUpFiveDiscardNum += tp_pack
			case algo.TonghuaDown:
				stat.DTHDownFiveFollowNum += tp_call
				stat.DTHDownFiveDiscardNum += tp_pack
			case algo.DuiziUp:
				stat.DDZUpFiveFollowNum += tp_call
				stat.DDZUpFiveDiscardNum += tp_pack
			case algo.DuiziDown:
				stat.DDZDownFiveFollowNum += tp_call
				stat.DDZDownFiveDiscardNum += tp_pack
			case algo.GaopaiUp:
				stat.DGPUpFiveFollowNum += tp_call
				stat.DGPUpFiveDiscardNum += tp_pack
			case algo.GaopaiDown:
				stat.DGPDownFiveFollowNum += tp_call
				stat.DGPDownFiveDiscardNum += tp_pack
			}
		}
	}

	// 查询渠道，账号类型，充值金额
	var users []ck.User
	var usersMap = make(map[string]ck.User)
	err = ck.Select(&users, `select userid,ad__bundle_id,channel,regist_area,money from col_user final where userid in ?`, userids)
	if err != nil {
		beego.Error("regUsers error: ", err)
		return
	}
	for _, user := range users {
		usersMap[user.Userid] = user
	}

	for _, stats := range roomUsers {
		for userid, stat := range stats {
			if user, ok := usersMap[userid]; ok {
				stat.AD_BundleId = user.AD_BundleId
				stat.Channel1 = user.Channel
				stat.RegistArea = user.RegistArea
				stat.Money = user.Money
			}

			// 保存数据
			addErr := StatisticsService.AddOrUpdateTpRoomStat(stat)
			if addErr != nil {
				beego.Error("TPRoomStatsCK fail err: ", addErr)
			}
		}
	}
}

// TP房间统计
func TPRoomStats(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
	date1, _ := utils.Unix(today)

	m := bson.M{"status": 1}
	var channellist []entity.ChannelInfo
	err := Channels.
		Find(m).
		All(&channellist)
	if err != nil {
		beego.Error("TPRoomStats fail err: ", err)
	}
	// 房间数据
	var roomlist []entity.Game
	m["gtype"] = 1
	m["ai_status"] = 1
	Games.Find(m).All(&roomlist)
	if len(channellist) > 0 {
		for _, item := range channellist {
			// 根据渠道ID获取用户ID
			uids := make([]string, 0)
			user_m := bson.M{}
			user_m["ad__bundle_id"] = item.Name
			user_m["robot"] = false
			user_m["simulation_robot"] = false
			PlayerUsers.Find(user_m).Distinct("_id", &uids)

			// 根据有金币流水的查询对局ID
			wids := make([]string, 0)
			log_m := bson.M{}
			log_m["userid"] = bson.M{"$in": uids}
			log_m["ltype"] = bson.M{"$in": []int{110, 111}}
			log_m["ctime"] = bson.M{"$gte": startTime, "$lte": endTime}
			LogWaters.Find(log_m).Distinct("water_id", &wids)

			// 根据查询到的对局ID 查询对局信息
			dt_m := bson.M{}
			dt_m["end_time"] = bson.M{"$gte": startTime.Unix(), "$lte": endTime.Unix()}
			dt_m["_id"] = bson.M{"$in": wids}
			dt_m["gtype"] = 1
			dt_m["players"] = bson.M{"$ne": ""}
			dtlist := make([]entity.Detail, 0)
			Details.Find(dt_m).All(&dtlist)

			var tpPlayerIds []string
			tpPlayers := make(map[string]map[string]*entity.TPRoomStat)
			if len(dtlist) > 0 {
				for _, item := range dtlist {
					SeatNumber := len(item.TPDetail)
					for _, v := range item.TPDetail {
						if len(v.UserId) >= 16 {
							// 人机id18位长度+
							continue
						}
						_, ok := tpPlayers[v.UserId]
						if !ok {
							tpPlayers[v.UserId] = make(map[string]*entity.TPRoomStat, 0)
						}
						stat, ok := tpPlayers[v.UserId][item.RoomId]
						if !ok {
							stat = &entity.TPRoomStat{
								Id:      fmt.Sprintf("%d-%s-%s", date1, v.UserId, item.RoomId),
								Date:    date1,
								DateStr: today,
								Userid:  v.UserId,
								Roomid:  item.RoomId,
							}
							// tpPlayerIds = append(tpPlayerIds, user.UserId)
						}
						stat.AllRounds++
						stat.Bets += v.Bet
						winflag := false
						if v.Score > 0 {
							// 赢局
							winflag = true
							stat.WinRounds++
							stat.WinBets += (v.Score - v.Bet)
						}
						if v.Score < 0 {
							// 输局
							winflag = false
							stat.LoseRounds++
							stat.LoseBets += v.Score
						}
						switch SeatNumber {
						case 2:
							stat.TwoRounds++
							stat.TwoBets += v.Bet
							if winflag {
								stat.WinTwoRounds++
								stat.WinTwoBets += (v.Score - v.Bet)
							}
						case 3:
							stat.ThreeRounds++
							stat.ThreeBets += v.Bet
							if winflag {
								stat.WinThreeRounds++
								stat.WinThreeBets += (v.Score - v.Bet)
							}
						case 4:
							stat.FourRounds++
							stat.FourBets += v.Bet
							if winflag {
								stat.WinFourRounds++
								stat.WinFourBets += (v.Score - v.Bet)
							}
						case 5:
							stat.FiveRounds++
							stat.FiveBets += v.Bet
							if winflag {
								stat.WinFiveRounds++
								stat.WinFiveBets += (v.Score - v.Bet)
							}
						}
						cardId := algo.HuaTypeUpOrDown(v.Cards)
						if cardId != 0 {
							switch cardId {
							case 10:
								stat.BZUpBets += v.Bet
								stat.BZUpNumber++
								tpTurn := len(v.TPDetailTurn)
								stat.BZUpRounds += int64(tpTurn)
								bigwin := false
								if winflag {
									// 赢
									win := v.Score - v.Bet
									multiple := v.Bottom * 10
									if win > multiple {
										bigwin = true
									}
								} else {
									// 输
									lose := v.Score - v.Bet
									multiple := v.Bottom * 10
									if lose > multiple {
										bigwin = true
									}
								}
								followNum := 0
								discardNum := 0
								// 跟注率、弃牌率
								switch SeatNumber {
								case 2:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.BZUpTwoFollowNum++
									}
									if discardNum > 0 {
										stat.BZUpTwoDiscardNum++
									}
								case 3:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.BZUpThreeFollowNum++
									}
									if discardNum > 0 {
										stat.BZUpThreeDiscardNum++
									}
								case 4:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.BZUpFourFollowNum++
									}
									if discardNum > 0 {
										stat.BZUpFourDiscardNum++
									}
								case 5:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.BZUpFiveFollowNum++
									}
									if discardNum > 0 {
										stat.BZUpFiveDiscardNum++
									}
								}

								if bigwin {
									// 存在大赢局或大输局
									if winflag {
										stat.BZUpWinRounds++
									} else {
										stat.BZUpLoseRounds++
									}
								}
							case 11:
								stat.BZDownBets += v.Bet
								stat.BZDownNumber++
								tpTurn := len(v.TPDetailTurn)
								stat.BZDownRounds += int64(tpTurn)
								bigwin := false
								if winflag {
									// 赢
									win := v.Score - v.Bet
									multiple := v.Bottom * 10
									if win > multiple {
										bigwin = true
									}
								} else {
									// 输
									lose := v.Score - v.Bet
									multiple := v.Bottom * 10
									if lose > multiple {
										bigwin = true
									}
								}
								followNum := 0
								discardNum := 0
								// 跟注率、弃牌率
								switch SeatNumber {
								case 2:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.BZDownTwoFollowNum++
									}
									if discardNum > 0 {
										stat.BZDownTwoDiscardNum++
									}
								case 3:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.BZDownThreeFollowNum++
									}
									if discardNum > 0 {
										stat.BZDownThreeDiscardNum++
									}
								case 4:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.BZDownFourFollowNum++
									}
									if discardNum > 0 {
										stat.BZDownFourDiscardNum++
									}
								case 5:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.BZDownFiveFollowNum++
									}
									if discardNum > 0 {
										stat.BZDownFiveDiscardNum++
									}
								}
								if bigwin {
									// 存在大赢局或大输局
									if winflag {
										stat.BZDownWinRounds++
									} else {
										stat.BZDownLoseRounds++
									}
								}
							case 20:
								stat.THSUpBets += v.Bet
								stat.THSUpNumber++
								tpTurn := len(v.TPDetailTurn)
								stat.THSUpRounds += int64(tpTurn)
								bigwin := false
								if winflag {
									// 赢
									win := v.Score - v.Bet
									multiple := v.Bottom * 10
									if win > multiple {
										bigwin = true
									}
								} else {
									// 输
									lose := v.Score - v.Bet
									multiple := v.Bottom * 10
									if lose > multiple {
										bigwin = true
									}
								}
								followNum := 0
								discardNum := 0
								// 跟注率、弃牌率
								switch SeatNumber {
								case 2:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.THSUpTwoFollowNum++
									}
									if discardNum > 0 {
										stat.THSUpTwoDiscardNum++
									}
								case 3:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.THSUpThreeFollowNum++
									}
									if discardNum > 0 {
										stat.THSUpThreeDiscardNum++
									}
								case 4:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.THSUpFourFollowNum++
									}
									if discardNum > 0 {
										stat.THSUpFourDiscardNum++
									}
								case 5:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.THSUpFiveFollowNum++
									}
									if discardNum > 0 {
										stat.THSUpFiveDiscardNum++
									}
								}
								if bigwin {
									// 存在大赢局或大输局
									if winflag {
										stat.THSUpWinRounds++
									} else {
										stat.THSUpLoseRounds++
									}
								}
							case 21:
								stat.THSDownBets += v.Bet
								stat.THSDownNumber++
								tpTurn := len(v.TPDetailTurn)
								stat.THSDownRounds += int64(tpTurn)
								bigwin := false
								if winflag {
									// 赢
									win := v.Score - v.Bet
									multiple := v.Bottom * 10
									if win > multiple {
										bigwin = true
									}
								} else {
									// 输
									lose := v.Score - v.Bet
									multiple := v.Bottom * 10
									if lose > multiple {
										bigwin = true
									}
								}
								followNum := 0
								discardNum := 0
								// 跟注率、弃牌率
								switch SeatNumber {
								case 2:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.THSDownTwoFollowNum++
									}
									if discardNum > 0 {
										stat.THSDownTwoDiscardNum++
									}
								case 3:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.THSDownThreeFollowNum++
									}
									if discardNum > 0 {
										stat.THSDownThreeDiscardNum++
									}
								case 4:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.THSDownFourFollowNum++
									}
									if discardNum > 0 {
										stat.THSDownFourDiscardNum++
									}
								case 5:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.THSDownFiveFollowNum++
									}
									if discardNum > 0 {
										stat.THSDownFiveDiscardNum++
									}
								}
								if bigwin {
									// 存在大赢局或大输局
									if winflag {
										stat.THSDownWinRounds++
									} else {
										stat.THSDownLoseRounds++
									}
								}
							case 30:
								stat.DSZUpBets += v.Bet
								stat.DSZUpNumber++
								tpTurn := len(v.TPDetailTurn)
								stat.DSZUpRounds += int64(tpTurn)
								bigwin := false
								if winflag {
									// 赢
									win := v.Score - v.Bet
									multiple := v.Bottom * 10
									if win > multiple {
										bigwin = true
									}
								} else {
									// 输
									lose := v.Score - v.Bet
									multiple := v.Bottom * 10
									if lose > multiple {
										bigwin = true
									}
								}
								followNum := 0
								discardNum := 0
								// 跟注率、弃牌率
								switch SeatNumber {
								case 2:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DSZUpTwoFollowNum++
									}
									if discardNum > 0 {
										stat.DSZUpTwoDiscardNum++
									}
								case 3:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DSZUpThreeFollowNum++
									}
									if discardNum > 0 {
										stat.DSZUpThreeDiscardNum++
									}
								case 4:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DSZUpFourFollowNum++
									}
									if discardNum > 0 {
										stat.DSZUpFourDiscardNum++
									}
								case 5:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DSZUpFiveFollowNum++
									}
									if discardNum > 0 {
										stat.DSZUpFiveDiscardNum++
									}
								}
								if bigwin {
									// 存在大赢局或大输局
									if winflag {
										stat.DSZUpWinRounds++
									} else {
										stat.DSZUpLoseRounds++
									}
								}
							case 31:
								stat.DSZDownBets += v.Bet
								stat.DSZDownNumber++
								tpTurn := len(v.TPDetailTurn)
								stat.DSZDownRounds += int64(tpTurn)
								bigwin := false
								if winflag {
									// 赢
									win := v.Score - v.Bet
									multiple := v.Bottom * 10
									if win > multiple {
										bigwin = true
									}
								} else {
									// 输
									lose := v.Score - v.Bet
									multiple := v.Bottom * 10
									if lose > multiple {
										bigwin = true
									}
								}
								followNum := 0
								discardNum := 0
								// 跟注率、弃牌率
								switch SeatNumber {
								case 2:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DSZDownTwoFollowNum++
									}
									if discardNum > 0 {
										stat.DSZDownTwoDiscardNum++
									}
								case 3:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DSZDownThreeFollowNum++
									}
									if discardNum > 0 {
										stat.DSZDownThreeDiscardNum++
									}
								case 4:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DSZDownFourFollowNum++
									}
									if discardNum > 0 {
										stat.DSZDownFourDiscardNum++
									}
								case 5:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DSZDownFiveFollowNum++
									}
									if discardNum > 0 {
										stat.DSZDownFiveDiscardNum++
									}
								}
								if bigwin {
									// 存在大赢局或大输局
									if winflag {
										stat.DSZDownWinRounds++
									} else {
										stat.DSZDownLoseRounds++
									}
								}
							case 40:
								stat.DTHUpBets += v.Bet
								stat.DTHUpNumber++
								tpTurn := len(v.TPDetailTurn)
								stat.DTHUpRounds += int64(tpTurn)
								bigwin := false
								if winflag {
									// 赢
									win := v.Score - v.Bet
									multiple := v.Bottom * 10
									if win > multiple {
										bigwin = true
									}
								} else {
									// 输
									lose := v.Score - v.Bet
									multiple := v.Bottom * 10
									if lose > multiple {
										bigwin = true
									}
								}
								followNum := 0
								discardNum := 0
								// 跟注率、弃牌率
								switch SeatNumber {
								case 2:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DTHUpTwoFollowNum++
									}
									if discardNum > 0 {
										stat.DTHUpTwoDiscardNum++
									}
								case 3:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DTHUpThreeFollowNum++
									}
									if discardNum > 0 {
										stat.DTHUpThreeDiscardNum++
									}
								case 4:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DTHUpFourFollowNum++
									}
									if discardNum > 0 {
										stat.DTHUpFourDiscardNum++
									}
								case 5:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DTHUpFiveFollowNum++
									}
									if discardNum > 0 {
										stat.DTHUpFiveDiscardNum++
									}
								}
								if bigwin {
									// 存在大赢局或大输局
									if winflag {
										stat.DTHUpWinRounds++
									} else {
										stat.DTHUpLoseRounds++
									}
								}
							case 41:
								stat.DTHDownBets += v.Bet
								stat.DTHDownNumber++
								tpTurn := len(v.TPDetailTurn)
								stat.DTHDownRounds += int64(tpTurn)
								bigwin := false
								if winflag {
									// 赢
									win := v.Score - v.Bet
									multiple := v.Bottom * 10
									if win > multiple {
										bigwin = true
									}
								} else {
									// 输
									lose := v.Score - v.Bet
									multiple := v.Bottom * 10
									if lose > multiple {
										bigwin = true
									}
								}
								followNum := 0
								discardNum := 0
								// 跟注率、弃牌率
								switch SeatNumber {
								case 2:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DTHDownTwoFollowNum++
									}
									if discardNum > 0 {
										stat.DTHDownTwoDiscardNum++
									}
								case 3:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DTHDownThreeFollowNum++
									}
									if discardNum > 0 {
										stat.DTHDownThreeDiscardNum++
									}
								case 4:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DTHDownFourFollowNum++
									}
									if discardNum > 0 {
										stat.DTHDownFourDiscardNum++
									}
								case 5:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DTHDownFiveFollowNum++
									}
									if discardNum > 0 {
										stat.DTHDownFiveDiscardNum++
									}
								}
								if bigwin {
									// 存在大赢局或大输局
									if winflag {
										stat.DTHDownWinRounds++
									} else {
										stat.DTHDownLoseRounds++
									}
								}
							case 50:
								stat.DDZUpBets += v.Bet
								stat.DDZUpNumber++
								tpTurn := len(v.TPDetailTurn)
								stat.DDZUpRounds += int64(tpTurn)
								bigwin := false
								if winflag {
									// 赢
									win := v.Score - v.Bet
									multiple := v.Bottom * 10
									if win > multiple {
										bigwin = true
									}
								} else {
									// 输
									lose := v.Score - v.Bet
									multiple := v.Bottom * 10
									if lose > multiple {
										bigwin = true
									}
								}
								followNum := 0
								discardNum := 0
								// 跟注率、弃牌率
								switch SeatNumber {
								case 2:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DDZUpTwoFollowNum++
									}
									if discardNum > 0 {
										stat.DDZUpTwoDiscardNum++
									}
								case 3:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DDZUpThreeFollowNum++
									}
									if discardNum > 0 {
										stat.DDZUpThreeDiscardNum++
									}
								case 4:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DDZUpFourFollowNum++
									}
									if discardNum > 0 {
										stat.DDZUpFourDiscardNum++
									}
								case 5:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DDZUpFiveFollowNum++
									}
									if discardNum > 0 {
										stat.DDZUpFiveDiscardNum++
									}
								}
								if bigwin {
									// 存在大赢局或大输局
									if winflag {
										stat.DDZUpWinRounds++
									} else {
										stat.DDZUpLoseRounds++
									}
								}
							case 51:
								stat.DDZDownBets += v.Bet
								stat.DDZDownNumber++
								tpTurn := len(v.TPDetailTurn)
								stat.DDZDownRounds += int64(tpTurn)
								bigwin := false
								if winflag {
									// 赢
									win := v.Score - v.Bet
									multiple := v.Bottom * 10
									if win > multiple {
										bigwin = true
									}
								} else {
									// 输
									lose := v.Score - v.Bet
									multiple := v.Bottom * 10
									if lose > multiple {
										bigwin = true
									}
								}
								followNum := 0
								discardNum := 0
								// 跟注率、弃牌率
								switch SeatNumber {
								case 2:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DDZDownTwoFollowNum++
									}
									if discardNum > 0 {
										stat.DDZDownTwoDiscardNum++
									}
								case 3:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DDZDownThreeFollowNum++
									}
									if discardNum > 0 {
										stat.DDZDownThreeDiscardNum++
									}
								case 4:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DDZDownFourFollowNum++
									}
									if discardNum > 0 {
										stat.DDZDownFourDiscardNum++
									}
								case 5:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DDZDownFiveFollowNum++
									}
									if discardNum > 0 {
										stat.DDZDownFiveDiscardNum++
									}
								}
								if bigwin {
									// 存在大赢局或大输局
									if winflag {
										stat.DDZDownWinRounds++
									} else {
										stat.DDZDownLoseRounds++
									}
								}
							case 60:
								stat.DGPUpBets += v.Bet
								stat.DGPUpNumber++
								tpTurn := len(v.TPDetailTurn)
								stat.DGPUpRounds += int64(tpTurn)
								bigwin := false
								if winflag {
									// 赢
									win := v.Score - v.Bet
									multiple := v.Bottom * 10
									if win > multiple {
										bigwin = true
									}
								} else {
									// 输
									lose := v.Score - v.Bet
									multiple := v.Bottom * 10
									if lose > multiple {
										bigwin = true
									}
								}
								followNum := 0
								discardNum := 0
								// 跟注率、弃牌率
								switch SeatNumber {
								case 2:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DGPUpTwoFollowNum++
									}
									if discardNum > 0 {
										stat.DGPUpTwoDiscardNum++
									}
								case 3:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DGPUpThreeFollowNum++
									}
									if discardNum > 0 {
										stat.DGPUpThreeDiscardNum++
									}
								case 4:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DGPUpFourFollowNum++
									}
									if discardNum > 0 {
										stat.DGPUpFourDiscardNum++
									}
								case 5:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DGPUpFiveFollowNum++
									}
									if discardNum > 0 {
										stat.DGPUpFiveDiscardNum++
									}
								}
								if bigwin {
									// 存在大赢局或大输局
									if winflag {
										stat.DGPUpWinRounds++
									} else {
										stat.DGPUpLoseRounds++
									}
								}
							case 61:
								stat.DGPDownBets += v.Bet
								stat.DGPDownNumber++
								tpTurn := len(v.TPDetailTurn)
								stat.DGPDownRounds += int64(tpTurn)
								bigwin := false
								if winflag {
									// 赢
									win := v.Score - v.Bet
									multiple := v.Bottom * 10
									if win > multiple {
										bigwin = true
									}
								} else {
									// 输
									lose := v.Score - v.Bet
									multiple := v.Bottom * 10
									if lose > multiple {
										bigwin = true
									}
								}
								followNum := 0
								discardNum := 0
								// 跟注率、弃牌率
								switch SeatNumber {
								case 2:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DGPDownTwoFollowNum++
									}
									if discardNum > 0 {
										stat.DGPDownTwoDiscardNum++
									}
								case 3:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DGPDownThreeFollowNum++
									}
									if discardNum > 0 {
										stat.DGPDownThreeDiscardNum++
									}
								case 4:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DGPDownFourFollowNum++
									}
									if discardNum > 0 {
										stat.DGPDownFourDiscardNum++
									}
								case 5:
									for _, t := range v.TPDetailTurn {
										contains := strings.Contains(t.Operation, "chaal")
										if contains {
											followNum++
										}
										contains1 := strings.Contains(t.Operation, "pack")
										if contains1 {
											discardNum++
										}
									}
									if followNum > 0 {
										stat.DGPDownFiveFollowNum++
									}
									if discardNum > 0 {
										stat.DGPDownFiveDiscardNum++
									}
								}
								if bigwin {
									// 存在大赢局或大输局
									if winflag {
										stat.DGPDownWinRounds++
									} else {
										stat.DGPDownLoseRounds++
									}
								}

							}
						}
						minRoomId := ""
						maxRoomId := ""
						minAccess := 0
						maxAccess := 0
						totalNum := 0
						pool_limit := 0
						carrycoin := int(v.BeforeScore)
						// 根据房间信息获取筹码池、次数
						for _, room := range roomlist {
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
							if room.Id == item.RoomId {
								pool_limit = room.TP.Pool_Limit
							}
						}
						if totalNum >= 2 {
							if maxRoomId == item.RoomId {
								stat.UpNumber++
							}
							if minRoomId == item.RoomId {
								stat.DownNumber++
							}
						}
						stat.ChipPool = int64(pool_limit)
						tpPlayerIds = append(tpPlayerIds, v.UserId)
						tpPlayers[v.UserId][item.RoomId] = stat
					}
				}
				// 查询渠道，账号类型，充值金额
				m3 := []bson.M{
					{"$match": bson.M{
						"_id": bson.M{"$in": tpPlayerIds},
					}},
					{"$project": bson.M{
						"_id":           "$_id",
						"ad__bundle_id": "$ad__bundle_id",
						"channel1":      "$channel1",
						"regist_area":   "$regist_area",
						"money":         "$money",
					}},
				}
				var r3 []bson.M
				err = PlayerUsers.Pipe(m3).All(&r3)
				if err != nil {
					beego.Warning("TPStats error3:", err)
				}
				for _, r := range r3 {
					userid := r["_id"].(string)
					ad__bundle_id := r["ad__bundle_id"]
					channel1 := r["channel1"]

					regist_area := r["regist_area"].(int)
					money := r["money"].(int)
					_, ok := tpPlayers[userid]
					if ok {
						if item, ok := tpPlayers[userid]; ok {
							for _, p := range item {
								if ad__bundle_id != nil {
									p.AD_BundleId = ad__bundle_id.(string)
								}
								if channel1 != nil {
									p.Channel1 = channel1.(string)
								}
								p.RegistArea = regist_area
								p.Money = uint32(money)
							}

						}
					}
				}
			}
			// 判断tpPlayers是否存在值
			if len(tpPlayers) > 0 {
				for _, tp := range tpPlayers {
					for _, v := range tp {
						// 保存数据
						addErr := StatisticsService.AddOrUpdateTpRoomStat(v)
						if addErr != nil {
							beego.Error("TPRoomStats fail err: ", addErr)
						}
					}
				}
			}
		}
	}
}

// 用户各类游戏对局统计第一次查全部
func UserGameDataFull() {
	// startTime := utils.Str2Time("2023-10-01 00:00:00")
	startTime := utils.Str2TimeZone("2023-10-01 00:00:00", locationName)
	endTime := utils.Str2TimeZone("2024-06-30 00:00:00", locationName)
	// 先清空表
	r_m := bson.M{}
	r_m["gtype"] = bson.M{"$in": []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}}
	UserGameDatas.RemoveAll(r_m)
	// 5天5天的查对局详情
	start, end := startTime, startTime.AddDate(0, 0, 5)
	for ; start.Before(endTime); start, end = end, end.AddDate(0, 0, 5) {
		gtypeUserRounds := make(map[string]map[int32]*entity.UserGameData, 16)
		m1 := []bson.M{
			{"$match": bson.M{
				"begin_time": bson.M{"$gte": start.Unix(), "$lt": end.Unix()},
				"players":    bson.M{"$ne": ""},
			}},
			// {"$project": bson.M{"gtype": "$gtype", "userids": bson.M{"$split": []string{"$players", ","}}}},
		}
		var details []entity.Detail
		err := Details.Pipe(m1).All(&details)
		if err != nil {
			glog.Errorf("", err)
			return
		}

		glog.Infof("用户对局统计: %v ~ %v, details=%d", start, end, len(details))

		for _, detail := range details {
			gtype := detail.Gtype
			switch gtype {
			case 1:
				for _, v := range detail.TPDetail {
					if len(v.UserId) >= 16 {
						continue
					}
					userid := v.UserId
					if _, ok := gtypeUserRounds[userid]; !ok {
						gtypeUserRounds[userid] = make(map[int32]*entity.UserGameData, 0)
					}
					stat, ok := gtypeUserRounds[userid][gtype]
					if !ok {
						stat = &entity.UserGameData{
							UserId: userid,
							Gtype:  int64(gtype),
						}
					}
					stat.Number++
					stat.Bets += v.Bet
					if v.Score > 0 {
						stat.WinNumber++
						stat.WinBets += v.Bet
					} else if v.Score < 0 {
						stat.LoseNumber++
						stat.LoseBets += v.Bet
					}
					gtypeUserRounds[userid][gtype] = stat
				}
			case 2:
				for _, v := range detail.LHDetail.UserDetail {
					if len(v.Userid) >= 16 {
						continue
					}
					userid := v.Userid
					if _, ok := gtypeUserRounds[userid]; !ok {
						gtypeUserRounds[userid] = make(map[int32]*entity.UserGameData, 0)
					}
					stat, ok := gtypeUserRounds[userid][gtype]
					if !ok {
						stat = &entity.UserGameData{
							UserId: userid,
							Gtype:  int64(gtype),
						}
					}
					if v.Userid == userid {
						stat.Number++
						stat.Bets += v.Win
						if v.Win > 0 {
							stat.WinNumber++
							stat.WinBets += v.Win
						} else if v.Win < 0 {
							stat.LoseNumber++
							stat.LoseBets += v.Win
						}
					}
					gtypeUserRounds[userid][gtype] = stat
				}
			case 3:
				for _, v := range detail.UPDetail.UserDetail {
					if len(v.Userid) >= 16 {
						continue
					}
					userid := v.Userid
					if _, ok := gtypeUserRounds[userid]; !ok {
						gtypeUserRounds[userid] = make(map[int32]*entity.UserGameData, 0)
					}
					stat, ok := gtypeUserRounds[userid][gtype]
					if !ok {
						stat = &entity.UserGameData{
							UserId: userid,
							Gtype:  int64(gtype),
						}
					}
					if v.Userid == userid {
						stat.Number++
						stat.Bets += v.Win
						if v.Win > 0 {
							stat.WinNumber++
							stat.WinBets += v.Win
						} else if v.Win < 0 {
							stat.LoseNumber++
							stat.LoseBets += v.Win
						}
					}
					gtypeUserRounds[userid][gtype] = stat
				}
			case 4:
				for _, v := range detail.RMDetail {
					if len(v.UserId) >= 16 {
						continue
					}
					userid := v.UserId
					if _, ok := gtypeUserRounds[userid]; !ok {
						gtypeUserRounds[userid] = make(map[int32]*entity.UserGameData, 0)
					}
					stat, ok := gtypeUserRounds[userid][gtype]
					if !ok {
						stat = &entity.UserGameData{
							UserId: userid,
							Gtype:  int64(gtype),
						}
					}
					stat.Number++
					stat.Bets += v.Score
					if v.Score > 0 {
						stat.WinNumber++
						stat.WinBets += v.Score
					} else if v.Score < 0 {
						stat.LoseNumber++
						stat.LoseBets += v.Score
					}
					gtypeUserRounds[userid][gtype] = stat
				}
			case 5:
				for _, v := range detail.AK47Detail {
					if len(v.UserId) >= 16 {
						continue
					}
					userid := v.UserId
					if _, ok := gtypeUserRounds[userid]; !ok {
						gtypeUserRounds[userid] = make(map[int32]*entity.UserGameData, 0)
					}
					stat, ok := gtypeUserRounds[userid][gtype]
					if !ok {
						stat = &entity.UserGameData{
							UserId: userid,
							Gtype:  int64(gtype),
						}
					}
					stat.Number++
					stat.Bets += v.Bet
					if v.Score > 0 {
						stat.WinNumber++
						stat.WinBets += v.Score
					} else if v.Score < 0 {
						stat.LoseNumber++
						stat.LoseBets += v.Score
					}
					gtypeUserRounds[userid][gtype] = stat
				}
			case 6:
				for _, v := range detail.JOKERDetail {
					if len(v.UserId) >= 16 {
						continue
					}
					userid := v.UserId
					if _, ok := gtypeUserRounds[userid]; !ok {
						gtypeUserRounds[userid] = make(map[int32]*entity.UserGameData, 0)
					}
					stat, ok := gtypeUserRounds[userid][gtype]
					if !ok {
						stat = &entity.UserGameData{
							UserId: userid,
							Gtype:  int64(gtype),
						}
					}
					stat.Number++
					stat.Bets += v.Bet
					if v.Score > 0 {
						stat.WinNumber++
						stat.WinBets += v.Score
					} else if v.Score < 0 {
						stat.LoseNumber++
						stat.LoseBets += v.Score
					}
					gtypeUserRounds[userid][gtype] = stat
				}
			case 7:
				for _, v := range detail.CRASHDetail.UserDetail {
					if len(v.Userid) >= 16 {
						continue
					}
					userid := v.Userid
					if _, ok := gtypeUserRounds[userid]; !ok {
						gtypeUserRounds[userid] = make(map[int32]*entity.UserGameData, 0)
					}
					stat, ok := gtypeUserRounds[userid][gtype]
					if !ok {
						stat = &entity.UserGameData{
							UserId: userid,
							Gtype:  int64(gtype),
						}
					}
					if v.Userid == userid {
						stat.Number++
						stat.Bets += v.Win
						if v.Win > 0 {
							stat.WinNumber++
							stat.WinBets += v.Win
						} else if v.Win < 0 {
							stat.LoseNumber++
							stat.LoseBets += v.Win
						}
					}
					gtypeUserRounds[userid][gtype] = stat
				}
			case 8:
				for _, v := range detail.ABDetail.UserDetail {
					if len(v.Userid) >= 16 {
						continue
					}
					userid := v.Userid
					if _, ok := gtypeUserRounds[userid]; !ok {
						gtypeUserRounds[userid] = make(map[int32]*entity.UserGameData, 0)
					}
					stat, ok := gtypeUserRounds[userid][gtype]
					if !ok {
						stat = &entity.UserGameData{
							UserId: userid,
							Gtype:  int64(gtype),
						}
					}
					if v.Userid == userid {
						stat.Number++
						stat.Bets += v.Win
						if v.Win > 0 {
							stat.WinNumber++
							stat.WinBets += v.Win
						} else if v.Win < 0 {
							stat.LoseNumber++
							stat.LoseBets += v.Win
						}
					}
					gtypeUserRounds[userid][gtype] = stat
				}
			case 9:
				for _, v := range detail.CPDetail.UserDetail {
					if len(v.Userid) >= 16 {
						continue
					}
					userid := v.Userid
					if _, ok := gtypeUserRounds[userid]; !ok {
						gtypeUserRounds[userid] = make(map[int32]*entity.UserGameData, 0)
					}
					stat, ok := gtypeUserRounds[userid][gtype]
					if !ok {
						stat = &entity.UserGameData{
							UserId: userid,
							Gtype:  int64(gtype),
						}
					}
					if v.Userid == userid {
						stat.Number++
						stat.Bets += v.Win
						if v.Win > 0 {
							stat.WinNumber++
							stat.WinBets += v.Win
						} else if v.Win < 0 {
							stat.LoseNumber++
							stat.LoseBets += v.Win
						}
					}
					gtypeUserRounds[userid][gtype] = stat
				}
			case 10:
				for _, v := range detail.CRASHDetail.UserDetail {
					if len(v.Userid) >= 16 {
						continue
					}
					userid := v.Userid
					if _, ok := gtypeUserRounds[userid]; !ok {
						gtypeUserRounds[userid] = make(map[int32]*entity.UserGameData, 0)
					}
					stat, ok := gtypeUserRounds[userid][gtype]
					if !ok {
						stat = &entity.UserGameData{
							UserId: userid,
							Gtype:  int64(gtype),
						}
					}
					if v.Userid == userid {
						stat.Number++
						stat.Bets += v.Win
						if v.Win > 0 {
							stat.WinNumber++
							stat.WinBets += v.Win
						} else if v.Win < 0 {
							stat.LoseNumber++
							stat.LoseBets += v.Win
						}
					}
					gtypeUserRounds[userid][gtype] = stat
				}
			}
		}
		dt_userid := make([]string, 0)
		for uid, _ := range gtypeUserRounds {
			dt_userid = append(dt_userid, uid)
		}
		// 游戏时长统计
		pipeline2 := []bson.M{
			{
				"$match": bson.M{
					"userid": bson.M{"$in": dt_userid},
					"ctime":  bson.M{"$gte": start.Unix(), "$lt": end.Unix()},
				},
			},
			{
				"$group": bson.M{
					"_id": bson.M{
						"gtype":  "$gtype",
						"userid": "$userid",
					},
					"time": bson.M{
						"$sum": "$time",
					},
				},
			},
			{
				"$project": bson.M{
					"_id":    "$_id.gtype",
					"userid": "$_id.userid",
					"time":   "$time",
				},
			},
		}
		result2 := []bson.M{}
		LogGameTimes.Pipe(pipeline2).All(&result2)
		for _, g := range result2 {
			gtypeid := g["_id"].(int)
			uid := g["userid"].(string)
			stat, ok := gtypeUserRounds[uid][int32(gtypeid)]
			if ok {
				time := g["time"].(int64)
				stat.GameTime = time
				gtypeUserRounds[uid][int32(gtypeid)] = stat
			}
		}
		// 将数据写入DB
		glog.Infof("写入数据: %d", len(gtypeUserRounds))
		if len(gtypeUserRounds) > 0 {
			for _, item := range gtypeUserRounds {
				for _, info := range item {
					info.Ctime = start.Unix()
					addErr := StatisticsService.AddOrUpdateUserGame(info)
					if addErr != nil {
						beego.Error("UserGameData fail err: ", addErr)
					}
				}
			}
			glog.Infof("写入数据完成: %d", len(gtypeUserRounds))
		}

	}
	glog.Infof("用户游戏数据补充完成！")
}

// 用户各类游戏对局统计第一次查全部
func UserGameDataFullCK() {
	timestamp := utils.TimestampToday(location)
	r_m := bson.M{}
	// r_m["gtype"] = bson.M{"$in": []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}}
	UserGameDatas.RemoveAll(r_m)
	sql1 := `
select userid,gtype,COUNT(1) number ,sum(CASE WHEN win_type = 2 and (gtype = 4 or gtype = 12) THEN settle_score ELSE bet_amount END) bets,SUM(end_time-begin_time) game_time,
SUM(case when win_type = 1 THEN 1 else 0 end) win_number,
SUM(
CASE
	WHEN win_type = 1 THEN
		CASE
			WHEN gtype IN (1, 5, 6) THEN settle_score - bet_amount
			WHEN gtype IN (7,10) THEN settle_score
			WHEN gtype IN (2, 3, 4, 8,9,11,12) THEN settle_score - cash_ming_tax - cash_an_tax
			ELSE 0
		END
	ELSE 0
END
) AS win_bets,
SUM(case when win_type = 2 THEN 1 else 0 end) lose_number,SUM(case when win_type = 2 THEN settle_score else 0 end) lose_bets
from game.col_detail cd FINAL 
where robot = 0 and win_type in (1,2,3) 
GROUP BY userid,gtype
UNION ALL
select user_id userid,game_id gtype,COUNT(1) number,SUM(amount) bets,0 game_time,SUM(win_number) win_number,SUM(win_bets) win_bets,SUM(lose_number) lose_number,SUM(lose_bets) lose_bets
from (SELECT t1.user_id,t1.game_id,t1.amount,(CASE WHEN red.win_bets > 0 THEN 1 ELSE 0 END) win_number,red.win_bets,(CASE WHEN red.win_bets = 0 THEN 1 ELSE 0 END) lose_number,(CASE WHEN red.win_bets = 0 THEN t1.amount ELSE 0 END) lose_bets
from game.col_nsq_log_external_bet t1 FINAL 
left join (	SELECT  user_id,game_id,round_id,SUM(amount) win_bets from game.col_nsq_log_external_reward cnler FINAL where cancel = 0  and amount != 0
GROUP BY user_id,game_id,round_id) red on t1.user_id = red.user_id and t1.game_id = red.game_id and t1.round_id = red.round_id 
where  t1.amount != 0 and t1.cancel = 0) tab 
GROUP BY user_id,game_id`
	list := make([]entity.UserGameData, 0)
	var args1 []any
	err = ck.Select(&list, sql1, args1...)
	if err != nil {
		beego.Warning("UserGameDataFullCK error3:", err)
		return
	}

	for _, info := range list {
		if info.Gtype == 12 || info.Gtype == 4 {
			// RM
			num := info.Bets
			absNum := int64(math.Abs(float64(num)))
			info.Bets = absNum
		}
		if info.Gtype > 20 {
			// 外接数据
			info.LoseBets = -info.LoseBets
		}
		info.Id = fmt.Sprintf("%s-%d", info.UserId, info.Gtype)
		info.Ctime = timestamp
		addErr := StatisticsService.AddOrUpdateUserGame(&info)
		if addErr != nil {
			beego.Error("UserGameDataFullCK fail err: ", addErr)
		}
	}
}

// LHDStats 龙虎游戏统计
func LHDStats(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
	date1, _ := utils.Unix(today)

	m1 := bson.M{
		"gtype":    int(pb.LHD),
		"end_time": bson.M{"$gt": startTime.Unix(), "$lt": endTime.Unix()},
		// "players":    bson.M{"$ne": ""},
	}

	var r1 []entity.Detail
	err := Details.Find(m1).All(&r1)
	if err != nil {
		beego.Warning("LHDStats error1:", err)
		return
	}
	if len(r1) == 0 {
		beego.Warning("no lhd detail data")
		return
	}

	// 查询当天注册用户
	m2 := []bson.M{
		{"$match": bson.M{
			"ctime": bson.M{"$gt": startTime, "$lt": endTime},
		}},
		{"$project": bson.M{"_id": "$_id"}},
	}
	var r2 []bson.M
	err = PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Warning("LHDStats error2:", err)
		return
	}
	var regUserids = make(map[string]bool)
	for _, r := range r2 {
		userid := r["_id"].(string)
		regUserids[userid] = true
	}

	// crash玩家列表
	var lhdPlayerIds []string
	var lhdPlayers = make(map[string]*entity.LHDPlayerStat)
	var lhdXxsc = make(map[string]*entity.LHDXxscStat)
	var lhdQsbns = make(map[string]*entity.LHDQsbnStat)
	var lhdUserIds []string
	var lhdLkyh = make(map[string]*entity.LHDLkyhStat)

	// 获取lhd适配人群配置
	var str_info entity.LhdStrategy
	err = LHDStrategies.Find(bson.M{}).One(&str_info)
	if err != nil {
		beego.Error("get lhd strategy err: ", err)
	}

	for _, detail := range r1 {
		if detail.LHDetail == nil {
			continue
		}

		for _, user := range detail.LHDetail.UserDetail {
			if len(user.Userid) >= 16 {
				// 人机id18位长度+
				continue
			}
			stat, ok := lhdPlayers[user.Userid]
			if !ok {
				stat = &entity.LHDPlayerStat{
					Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
					Date:    date1,
					DateStr: today,
					Userid:  user.Userid,
					NewReg:  regUserids[user.Userid], // 老顾客新顾客
				}
				lhdPlayers[user.Userid] = stat
				lhdPlayerIds = append(lhdPlayerIds, user.Userid)
			}
			stat.AllRounds++
			stat.GameTimes += (detail.EndTime - detail.BeginTime)
			// 是否观察位
			if user.Observe {
				stat.ObserveRounds++
				continue
			}
			userBet := user.Dragon + user.Tiger + user.Tie
			stat.Bets += userBet
			if userBet > 0 {
				stat.BetRounds++
			}

			switch user.Result {
			case "赢":
				stat.WinRounds++
				stat.WinBets += userBet
				stat.Wins += user.Win
				stat.Cash += user.Win

			case "输":
				stat.LoseRounds++
				stat.LoseBets += userBet
				stat.Loses += user.Win
				stat.Cash += user.Win

			case "平":
				stat.TieRounds++
				stat.TieBets += userBet
			}

			// detail.LHDetail.Winner // 1龙,2虎,else和
			if userBet > 0 {
				switch detail.LHDetail.Winner {
				case 1:
					stat.Dragons++
				case 2:
					stat.Tigers++
				default:
					stat.Ties++
				}
			}

			var betTypes int
			if user.Dragon > 0 {
				stat.PlayerDragons++
				betTypes++
				if user.Result == "赢" {
					stat.PlayerWinDragons++
				}
			}
			if user.Tiger > 0 {
				stat.PlayerTigers++
				betTypes++
				if user.Result == "赢" {
					stat.PlayerWinTigers++
				}
			}
			if user.Tie > 0 {
				stat.PlayerTies++
				betTypes++
				if user.Result == "赢" {
					stat.PlayerWinTies++
				}
			}
			if betTypes >= 2 {
				stat.PlayerMultis++
			}

			if detail.LHDetail.StrategyId != 0 {
				switch detail.LHDetail.StrategyId {
				case 1:
					xxsc, ok := lhdXxsc[user.Userid]
					if !ok {
						xxsc = &entity.LHDXxscStat{
							Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
							Date:    date1,
							DateStr: today,
							Userid:  user.Userid,
						}
						lhdXxsc[user.Userid] = xxsc
					}
					xxsc.StrategyPlayers = 1
					xxsc.StrategyTimes++
					xxsc.StrategyRounds++
					if user.Result == "输" { // 扰动局
						xxsc.StrategyDisturbRounds++
					}
					if betTypes > 1 {
						xxsc.StrategyMultiRounds++
					}
					xxsc.StrategyBets += userBet
					if user.Result == "赢" {
						xxsc.StrategyWins += user.Win
						xxsc.StrategyWinRounds++
					}
					if user.Result == "输" {
						xxsc.StrategyLoses += user.Win
					}

				case 2:
					qsbn, ok := lhdQsbns[user.Userid]

					if !ok {
						qsbn = &entity.LHDQsbnStat{
							Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
							Date:    date1,
							DateStr: today,
							Userid:  user.Userid,
						}
						lhdQsbns[user.Userid] = qsbn
					}
					qsbn.StrategyPlayers = 1
					qsbn.StrategyTimes++
					qsbn.BeforeBackRateSum += user.BeforeBackRate
					qsbn.BeforeBackRateCount++
					qsbn.AfterBackRateSum += user.AfterBackRate
					qsbn.AfterBackRateCount++

					if betTypes > 1 {
						qsbn.StrategyMultiRounds++
					}
					qsbn.StrategyBets += userBet
					if user.Result == "赢" {
						qsbn.StrategyWins += user.Win
					}
					if user.Result == "输" {
						qsbn.StrategyLoses += user.Win
					}
				case 3:
					// 龙狂有祸
					lkyh, ok := lhdLkyh[user.Userid]
					if !ok {
						lkyh = &entity.LHDLkyhStat{
							Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
							Date:    date1,
							DateStr: today,
							Userid:  user.Userid,
							BSDays:  1,
						}
						lhdLkyh[user.Userid] = lkyh
						lhdUserIds = append(lhdUserIds, user.Userid)
					}
					lkyh.StrategyTimes++
					// 判断是否存在压制局
					if detail.LHDetail.IsSuppress {
						lkyh.YZRounds++
						lkyh.YZNumber++
						lkyh.YZBets += userBet
					} else {
						lkyh.BSBets += userBet
					}

					if betTypes > 1 {
						lkyh.StrategyMultiRounds++
					}
					lkyh.StrategyBets += userBet
					if user.Result == "赢" {
						lkyh.StrategyWinRounds++
						lkyh.StrategyWins += user.Win
					}
					if user.Result == "输" {
						lkyh.StrategyLoseRounds++
						lkyh.StrategyLoses += user.Win
					}
				}
			}

			// if str_info.QSBN.D > user.BeforeScore-userBet {
			// 	qsbn, ok := lhdQsbns[user.Userid]
			// 	if !ok {
			// 		qsbn = &entity.LHDQsbnStat{
			// 			Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
			// 			Date:    date1,
			// 			DateStr: today,
			// 			Userid:  user.Userid,
			// 		}
			// 		lhdQsbns[user.Userid] = qsbn
			// 	}
			// 	qsbn.AllInRounds++
			// }
		}
	}

	// 查询渠道，账号类型，充值金额
	m3 := []bson.M{
		{"$match": bson.M{
			"_id": bson.M{"$in": lhdPlayerIds},
		}},
		{"$project": bson.M{
			"_id":           "$_id",
			"ad__bundle_id": "$ad__bundle_id",
			"channel1":      "$channel1",
			"regist_area":   "$regist_area",
			"money":         "$money",
			"lhd_strategy":  "$lhd_strategy",
		}},
	}
	var r3 []entity.PlayerUser
	err = PlayerUsers.Pipe(m3).All(&r3)
	if err != nil {
		beego.Warning("LHDStats error3:", err)
	}
	for _, r := range r3 {
		userid := r.Userid
		if p, ok := lhdPlayers[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			// p.Channel1 = r.Channel1
			p.RegistArea = r.RegistArea
			p.Money = r.Money
		}
		if p, ok := lhdXxsc[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			// p.Channel1 = r.Channel1
			p.RegistArea = r.RegistArea
			p.Money = r.Money
		}
		if p, ok := lhdQsbns[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			// p.Channel1 = r.Channel1
			p.RegistArea = r.RegistArea
			p.Money = r.Money
		}
		if p, ok := lhdLkyh[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			// p.Channel1 = r.Channel1
			p.RegistArea = r.RegistArea
			p.Money = r.Money
			// 策略
			p.BSTimes = r.LHDStrategy.LKYH.BSTimes
			p.TriggerTimes = int64(r.LHDStrategy.LKYH.TriggerTimes)
			p.TZ = int64(r.LHDStrategy.LKYH.TZ)
			p.NZ = r.LHDStrategy.LKYH.NZ
			p.RZ = r.LHDStrategy.LKYH.RZ
			// 获取用户LHD游戏总汇数据
			if t, ok := lhdPlayers[userid]; ok {
				p.Rounds = int64(t.BetRounds)
				p.Bets = t.Bets
				p.WinRounds = int64(t.WinRounds)
				p.WinBets = t.Wins
				p.LoseBets = t.Loses
				p.LoseRounds = int64(t.LoseRounds)
			}
		}
	}

	for _, p := range lhdPlayers {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)

		// 保存数据
		addErr := StatisticsService.AddOrUpdateLHDPlayerStat(p)
		if addErr != nil {
			beego.Error("LHDPlayerStat fail err: ", addErr)
		}
	}

	for _, item := range lhdXxsc {
		err := StatisticsService.AddOrUpdateLHDXxscStat(item)
		if err != nil {
			beego.Error("LHDXxscStat fail err: ", err)
		}
	}

	for _, item := range lhdQsbns {
		if item.BeforeBackRateCount > 0 {
			item.BeforeBackRate = item.BeforeBackRateSum / item.BeforeBackRateCount
		}
		if item.AfterBackRateCount > 0 {
			item.AfterBackRate = item.AfterBackRateSum / item.AfterBackRateCount
		}
		err := StatisticsService.AddOrUpdateLHDQsbnStat(item)
		if err != nil {
			beego.Error("LHDQsbnStat fail err: ", err)
		}
	}
	for _, item := range lhdLkyh {
		// 局均码
		item.RoundBetAvg = item.Bets / int64(item.Rounds)
		err := StatisticsService.AddOrUpdateLHDLkyhStat(item)
		if err != nil {
			beego.Error("LHDLkyhStat fail err: ", err)
		}
	}
}

// LHDStats 龙虎游戏统计
func LHDStatsCK(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
	date1, _ := utils.Unix(today)
	gtype := int(pb.LHD)

	sql1 := `select userid,gtype,begin_time,end_time,lhd_winner,lhd_dragon_value,lhd_tiger_value,lhd_bets,lhd_player_win,lhd_strategy_id,lhd_is_suppress,lhd_userid,lhd_dragon,lhd_tiger,lhd_tie,lhd_observe,lhd_result,lhd_win,lhd_before_score,lhd_after_score,lhd_before_cash,lhd_after_cash,lhd_before_bonus,lhd_after_bonus,lhd_cash_ming_tax,lhd_bonus_ming_tax,lhd_cash_an_tax,lhd_bonus_an_tax,lhd_before_back_rate,lhd_after_back_rate from game.col_detail cd FINAL where robot = false and gtype = ? and begin_time between ? and ?`
	var r1 []entity.DetailCK
	var args1 []any
	args1 = append(args1, gtype)
	args1 = append(args1, startTime.Unix())
	args1 = append(args1, endTime.Unix())
	err = ck.Select(&r1, sql1, args1...)
	if err != nil {
		beego.Warning("LHDStats error1:", err)
		return
	}
	if len(r1) == 0 {
		beego.Info("no lhd detail data")
		return
	}

	// 查询当天注册用户
	var r2 []string
	var args2 []any
	sql2 := `select userid FROM game.col_user cu FINAL where robot = false and simulation_robot = false and ctime BETWEEN ? and ?`
	args2 = append(args2, startTime)
	args2 = append(args2, endTime)
	err = ck.Select(&r2, sql2, args2...)
	if err != nil {
		beego.Warning("LHDStats error2:", err)
		return
	}
	var regUserids = make(map[string]bool)
	for _, r := range r2 {
		userid := r
		regUserids[userid] = true
	}

	// lhd玩家列表
	var lhdPlayerIds []string
	var lhdPlayers = make(map[string]*entity.LHDPlayerStat)
	var lhdXxsc = make(map[string]*entity.LHDXxscStat)
	var lhdQsbns = make(map[string]*entity.LHDQsbnStat)
	var lhdUserIds []string
	var lhdLkyh = make(map[string]*entity.LHDLkyhStat)

	// 玩家下注后分数
	var lhdBets = make(map[string][]*entity.ABAnqsUserBet)
	// 获取lhd适配人群配置
	var str_info entity.LhdStrategy
	err = LHDStrategies.Find(bson.M{}).One(&str_info)
	if err != nil {
		beego.Error("get lhd strategy err: ", err)
	}

	for _, user := range r1 {

		stat, ok := lhdPlayers[user.UserId]
		if !ok {
			stat = &entity.LHDPlayerStat{
				Id:      fmt.Sprintf("%d-%s", date1, user.UserId),
				Date:    date1,
				DateStr: today,
				Userid:  user.UserId,
				NewReg:  regUserids[user.UserId], // 老顾客新顾客
			}
			lhdPlayers[user.UserId] = stat
			lhdPlayerIds = append(lhdPlayerIds, user.UserId)
		}
		stat.AllRounds++
		stat.GameTimes += (user.EndTime - user.BeginTime)
		// 是否观察位
		if user.LhdObserve {
			stat.ObserveRounds++
			continue
		}
		userBet := user.LhdDragon + user.LhdTiger + user.LhdTie
		stat.Bets += userBet
		if userBet > 0 {
			stat.BetRounds++
		}

		switch user.LhdResult {
		case "赢":
			stat.WinRounds++
			stat.WinBets += userBet
			stat.Wins += user.LhdWin
			stat.Cash += user.LhdWin

		case "输":
			stat.LoseRounds++
			stat.LoseBets += userBet
			stat.Loses += user.LhdWin
			stat.Cash += user.LhdWin

		case "平":
			stat.TieRounds++
			stat.TieBets += userBet
		}

		// detail.LHDetail.Winner // 1龙,2虎,else和
		if userBet > 0 {
			switch user.LhdWinner {
			case 1:
				stat.Dragons++
			case 2:
				stat.Tigers++
			default:
				stat.Ties++
			}
		}

		var betTypes int
		if user.LhdDragon > 0 {
			stat.PlayerDragons++
			betTypes++
			if user.LhdResult == "赢" {
				stat.PlayerWinDragons++
			}
		}
		if user.LhdTiger > 0 {
			stat.PlayerTigers++
			betTypes++
			if user.LhdResult == "赢" {
				stat.PlayerWinTigers++
			}
		}
		if user.LhdTie > 0 {
			stat.PlayerTies++
			betTypes++
			if user.LhdResult == "赢" {
				stat.PlayerWinTies++
			}
		}
		if betTypes >= 2 {
			stat.PlayerMultis++
		}

		if user.LhdStrategyId != 0 {
			switch user.LhdStrategyId {
			case 1:
				xxsc, ok := lhdXxsc[user.UserId]
				if !ok {
					xxsc = &entity.LHDXxscStat{
						Id:      fmt.Sprintf("%d-%s", date1, user.UserId),
						Date:    date1,
						DateStr: today,
						Userid:  user.UserId,
					}
					lhdXxsc[user.UserId] = xxsc
				}
				xxsc.StrategyPlayers = 1
				xxsc.StrategyTimes++
				xxsc.StrategyRounds++
				if user.LhdResult == "输" { // 扰动局
					xxsc.StrategyDisturbRounds++
				}
				if betTypes > 1 {
					xxsc.StrategyMultiRounds++
				}
				xxsc.StrategyBets += userBet
				if user.LhdResult == "赢" {
					xxsc.StrategyWins += user.LhdWin
					xxsc.StrategyWinRounds++
				}
				if user.LhdResult == "输" {
					xxsc.StrategyLoses += user.LhdWin
				}

			case 2:
				qsbn, ok := lhdQsbns[user.UserId]

				if !ok {
					qsbn = &entity.LHDQsbnStat{
						Id:      fmt.Sprintf("%d-%s", date1, user.UserId),
						Date:    date1,
						DateStr: today,
						Userid:  user.UserId,
					}
					lhdQsbns[user.UserId] = qsbn
				}
				qsbn.StrategyPlayers = 1
				qsbn.StrategyTimes++
				qsbn.BeforeBackRateSum += user.LhdBeforeBackRate
				qsbn.BeforeBackRateCount++
				qsbn.AfterBackRateSum += user.LhdAfterBackRate
				qsbn.AfterBackRateCount++

				if betTypes > 1 {
					qsbn.StrategyMultiRounds++
				}
				qsbn.StrategyBets += userBet
				if user.LhdResult == "赢" {
					qsbn.StrategyWins += user.LhdWin
				}
				if user.LhdResult == "输" {
					qsbn.StrategyLoses += user.LhdWin
				}
			case 3:
				// 龙狂有祸
				lkyh, ok := lhdLkyh[user.UserId]
				if !ok {
					lkyh = &entity.LHDLkyhStat{
						Id:      fmt.Sprintf("%d-%s", date1, user.UserId),
						Date:    date1,
						DateStr: today,
						Userid:  user.UserId,
						BSDays:  1,
					}
					lhdLkyh[user.UserId] = lkyh
					lhdUserIds = append(lhdUserIds, user.UserId)
				}
				lkyh.StrategyTimes++
				// 判断是否存在压制局
				if user.LhdIsSuppress {
					lkyh.YZRounds++
					lkyh.YZNumber++
					lkyh.YZBets += userBet
				} else {
					lkyh.BSBets += userBet
				}

				if betTypes > 1 {
					lkyh.StrategyMultiRounds++
				}
				lkyh.StrategyBets += userBet
				if user.LhdResult == "赢" {
					lkyh.StrategyWinRounds++
					lkyh.StrategyWins += user.LhdWin
				}
				if user.LhdResult == "输" {
					lkyh.StrategyLoseRounds++
					lkyh.StrategyLoses += user.LhdWin
				}
			}
		}
		// 记录每个用户下注后金额
		score := user.LhdBeforeScore - userBet
		betinfo := &entity.ABAnqsUserBet{
			Userid: user.UserId,
			Score:  score,
		}
		betlist, ok := lhdBets[user.UserId]
		if !ok {
			betlist := make([]*entity.ABAnqsUserBet, 0)
			lhdBets[user.UserId] = betlist
		}
		betlist = append(betlist, betinfo)
		lhdBets[user.UserId] = betlist

		// if str_info.QSBN.D > user.LhdBeforeScore-userBet {
		// 	qsbn, ok := lhdQsbns[user.UserId]
		// 	if !ok {
		// 		qsbn = &entity.LHDQsbnStat{
		// 			Id:      fmt.Sprintf("%d-%s", date1, user.UserId),
		// 			Date:    date1,
		// 			DateStr: today,
		// 			Userid:  user.UserId,
		// 		}
		// 		lhdQsbns[user.UserId] = qsbn
		// 	}
		// 	qsbn.AllInRounds++
		// }
	}

	// 查询渠道，账号类型，充值金额
	sql3 := `select userid,ad__bundle_id,regist_area,money,state,lhd_round_bet,lhd_xxsc_disturb_cool_down,lhd_xxsc_win_length,lhd_xxsc_evo,lhd_xxsc_trigger_times,lhd_xxsc_max_trigger_times,lhd_qsbn_trigger_times,lhd_lkyh_trigger_times,lhd_lkyh_history_t,lhd_lkyh_tz,lhd_lkyh_rz,lhd_lkyh_bs_cool_down,lhd_lkyh_bs_rounds,lhd_lkyh_suppress,lhd_lkyh_nz,lhd_lkyh_bstimes,lhd_lkyh_doubleBetTimes from game.col_user cu FINAL where robot = false and simulation_robot = false and userid in ?`
	var args3 []any
	args3 = append(args3, lhdPlayerIds)
	var r3 []entity.PlayerUserCK
	err = ck.Select(&r3, sql3, args3...)
	if err != nil {
		beego.Warning("LHDStats error3:", err)
		return
	}
	for _, r := range r3 {
		userid := r.Userid
		if p, ok := lhdPlayers[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.RegistArea = r.RegistArea
			p.Money = r.Money
		}
		if p, ok := lhdXxsc[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.RegistArea = r.RegistArea
			p.Money = r.Money
		}
		if p, ok := lhdQsbns[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.RegistArea = r.RegistArea
			p.Money = r.Money
			userType := ConvertToUserType(int(r.Money), r.State)
			//计算allin局数
			allinRounds := 0
			if len(str_info.QSBN.D) > 0 {
				for t, d := range str_info.QSBN.D {
					if t == userType {
						if u, ok := lhdBets[userid]; ok {
							for _, t := range u {
								if d > t.Score {
									allinRounds++
								}
							}
						}
					}
				}
			}
			p.AllInRounds = int32(allinRounds)
		}
		if p, ok := lhdLkyh[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.RegistArea = r.RegistArea
			p.Money = r.Money
			// 策略
			p.BSTimes = r.LHDLKYHBSTimes
			p.TriggerTimes = int64(r.LHDLKYHTriggerTimes)
			p.TZ = int64(r.LHDLKYHTZ)
			p.NZ = r.LHDLKYHTZ
			p.RZ = r.LHDLKYHRZ
			// 获取用户LHD游戏总汇数据
			if t, ok := lhdPlayers[userid]; ok {
				p.Rounds = int64(t.BetRounds)
				p.Bets = t.Bets
				p.WinRounds = int64(t.WinRounds)
				p.WinBets = t.Wins
				p.LoseBets = t.Loses
				p.LoseRounds = int64(t.LoseRounds)
			}
		}
	}

	for _, p := range lhdPlayers {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)

		// 保存数据
		addErr := StatisticsService.AddOrUpdateLHDPlayerStat(p)
		if addErr != nil {
			beego.Error("LHDPlayerStat fail err: ", addErr)
		}
	}

	for _, item := range lhdXxsc {
		err := StatisticsService.AddOrUpdateLHDXxscStat(item)
		if err != nil {
			beego.Error("LHDXxscStat fail err: ", err)
		}
	}

	for _, item := range lhdQsbns {
		if item.BeforeBackRateCount > 0 {
			item.BeforeBackRate = item.BeforeBackRateSum / item.BeforeBackRateCount
		}
		if item.AfterBackRateCount > 0 {
			item.AfterBackRate = item.AfterBackRateSum / item.AfterBackRateCount
		}
		err := StatisticsService.AddOrUpdateLHDQsbnStat(item)
		if err != nil {
			beego.Error("LHDQsbnStat fail err: ", err)
		}
	}
	for _, item := range lhdLkyh {
		// 局均码
		item.RoundBetAvg = item.Bets / int64(item.Rounds)
		err := StatisticsService.AddOrUpdateLHDLkyhStat(item)
		if err != nil {
			beego.Error("LHDLkyhStat fail err: ", err)
		}
	}
}

// UpdownStats 7updown游戏统计
func UpdownStats(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
	date1, _ := utils.Unix(today)

	m1 := bson.M{
		"gtype":    int(pb.SEVEN),
		"end_time": bson.M{"$gt": startTime.Unix(), "$lt": endTime.Unix()},
		// "players":    bson.M{"$ne": ""},
	}

	var r1 []entity.Detail
	err := Details.Find(m1).All(&r1)
	if err != nil {
		beego.Warning("UpdownStats error1:", err)
		return
	}
	if len(r1) == 0 {
		beego.Warning("no 7updown detail data")
		return
	}

	// 查询当天注册用户
	m2 := []bson.M{
		{"$match": bson.M{
			"ctime": bson.M{"$gt": startTime, "$lt": endTime},
		}},
		{"$project": bson.M{"_id": "$_id"}},
	}
	var r2 []bson.M
	err = PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Warning("UpdownStats error2:", err)
		return
	}
	var regUserids = make(map[string]bool)
	for _, r := range r2 {
		userid := r["_id"].(string)
		regUserids[userid] = true
	}

	// crash玩家列表
	var upPlayerIds []string
	var upPlayers = make(map[string]*entity.UpdownPlayerStat)
	var upXxsc = make(map[string]*entity.UpdownXxscStat)
	var upQsbns = make(map[string]*entity.UpdownQsbnStat)
	var upLkyh = make(map[string]*entity.UpdownLkyhStat)

	// 获取lhd适配人群配置
	var str_info entity.SevenStrategy
	err = UPStrategies.Find(bson.M{}).One(&str_info)
	if err != nil {
		beego.Error("get 7updown strategy err: ", err)
	}

	for _, detail := range r1 {
		if detail.UPDetail == nil {
			continue
		}

		for _, user := range detail.UPDetail.UserDetail {
			if len(user.Userid) >= 16 {
				// 人机id18位长度+
				continue
			}
			stat, ok := upPlayers[user.Userid]
			if !ok {
				stat = &entity.UpdownPlayerStat{
					Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
					Date:    date1,
					DateStr: today,
					Userid:  user.Userid,
					NewReg:  regUserids[user.Userid], // 老顾客新顾客
				}
				upPlayers[user.Userid] = stat
				upPlayerIds = append(upPlayerIds, user.Userid)
			}
			stat.AllRounds++
			stat.GameTimes += (detail.EndTime - detail.BeginTime)
			// 是否观察位
			if user.Observe {
				stat.ObserveRounds++
				continue
			}
			userBet := user.Dragon + user.Tiger + user.Tie
			stat.Bets += userBet
			if userBet > 0 {
				stat.BetRounds++
			}

			switch user.Result {
			case "赢":
				stat.WinRounds++
				stat.WinBets += userBet
				stat.Wins += user.Win
				stat.Cash += user.Win

			case "输":
				stat.LoseRounds++
				stat.LoseBets += userBet
				stat.Loses += user.Win
				stat.Cash += user.Win

			case "平":
				stat.TieRounds++
				stat.TieBets += userBet
			}

			// detail.LHDetail.Winner // 1龙,2虎,else和
			if userBet > 0 {
				switch detail.UPDetail.Winner {
				case 1:
					stat.Dragons++
				case 2:
					stat.Tigers++
				default:
					stat.Ties++
				}
			}

			var betTypes int
			if user.Dragon > 0 {
				stat.PlayerDragons++
				betTypes++
				if user.Result == "赢" {
					stat.PlayerWinDragons++
				}
			}
			if user.Tiger > 0 {
				stat.PlayerTigers++
				betTypes++
				if user.Result == "赢" {
					stat.PlayerWinTigers++
				}
			}
			if user.Tie > 0 {
				stat.PlayerTies++
				betTypes++
				if user.Result == "赢" {
					stat.PlayerWinTies++
				}
			}
			if betTypes >= 2 {
				stat.PlayerMultis++
			}

			if detail.UPDetail.StrategyId != 0 {
				switch detail.UPDetail.StrategyId {
				case 1:
					xxsc, ok := upXxsc[user.Userid]
					if !ok {
						xxsc = &entity.UpdownXxscStat{
							Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
							Date:    date1,
							DateStr: today,
							Userid:  user.Userid,
						}
						upXxsc[user.Userid] = xxsc
					}
					xxsc.StrategyPlayers = 1
					xxsc.StrategyTimes++
					xxsc.StrategyRounds++
					if user.Result == "输" { // 扰动局
						xxsc.StrategyDisturbRounds++
					}
					if betTypes > 1 {
						xxsc.StrategyMultiRounds++
					}
					xxsc.StrategyBets += userBet
					if user.Result == "赢" {
						xxsc.StrategyWins += user.Win
						xxsc.StrategyWinRounds++
					}
					if user.Result == "输" {
						xxsc.StrategyLoses += user.Win
					}

				case 2:
					qsbn, ok := upQsbns[user.Userid]

					if !ok {
						qsbn = &entity.UpdownQsbnStat{
							Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
							Date:    date1,
							DateStr: today,
							Userid:  user.Userid,
						}
						upQsbns[user.Userid] = qsbn
					}
					qsbn.StrategyPlayers = 1
					qsbn.StrategyTimes++
					qsbn.BeforeBackRateSum += user.BeforeBackRate
					qsbn.BeforeBackRateCount++
					qsbn.AfterBackRateSum += user.AfterBackRate
					qsbn.AfterBackRateCount++

					if betTypes > 1 {
						qsbn.StrategyMultiRounds++
					}
					qsbn.StrategyBets += userBet
					if user.Result == "赢" {
						qsbn.StrategyWins += user.Win
					}
					if user.Result == "输" {
						qsbn.StrategyLoses += user.Win
					}
				case 3:
					// 龙狂有祸
					lkyh, ok := upLkyh[user.Userid]
					if !ok {
						lkyh = &entity.UpdownLkyhStat{
							Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
							Date:    date1,
							DateStr: today,
							Userid:  user.Userid,
							BSDays:  1,
						}
						upLkyh[user.Userid] = lkyh
					}
					lkyh.StrategyTimes++
					// 判断是否存在压制局
					if detail.UPDetail.IsSuppress {
						lkyh.YZRounds++
						lkyh.YZNumber++
						lkyh.YZBets += userBet
					} else {
						lkyh.BSBets += userBet
					}

					if betTypes > 1 {
						lkyh.StrategyMultiRounds++
					}
					lkyh.StrategyBets += userBet
					if user.Result == "赢" {
						lkyh.StrategyWinRounds++
						lkyh.StrategyWins += user.Win
					}
					if user.Result == "输" {
						lkyh.StrategyLoseRounds++
						lkyh.StrategyLoses += user.Win
					}
				}
			}

			// if str_info.QSBN.D > user.BeforeScore-userBet {
			// 	qsbn, ok := upQsbns[user.Userid]
			// 	if !ok {
			// 		qsbn = &entity.UpdownQsbnStat{
			// 			Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
			// 			Date:    date1,
			// 			DateStr: today,
			// 			Userid:  user.Userid,
			// 		}
			// 		upQsbns[user.Userid] = qsbn
			// 	}
			// 	qsbn.AllInRounds++
			// }
		}
	}

	// 查询渠道，账号类型，充值金额
	m3 := []bson.M{
		{"$match": bson.M{
			"_id": bson.M{"$in": upPlayerIds},
		}},
		{"$project": bson.M{
			"_id":            "$_id",
			"ad__bundle_id":  "$ad__bundle_id",
			"channel1":       "$channel1",
			"regist_area":    "$regist_area",
			"money":          "$money",
			"seven_strategy": "$seven_strategy",
		}},
	}
	var r3 []entity.PlayerUser
	err = PlayerUsers.Pipe(m3).All(&r3)
	if err != nil {
		beego.Warning("7updownStats error3:", err)
	}
	for _, r := range r3 {
		userid := r.Userid
		if p, ok := upPlayers[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			// p.Channel1 = r.Channel1
			p.RegistArea = r.RegistArea
			p.Money = r.Money
		}
		if p, ok := upXxsc[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			// p.Channel1 = r.Channel1
			p.RegistArea = r.RegistArea
			p.Money = r.Money
		}
		if p, ok := upQsbns[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			// p.Channel1 = r.Channel1
			p.RegistArea = r.RegistArea
			p.Money = r.Money
		}

		// 龙狂有祸
		if p, ok := upLkyh[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			// p.Channel1 = r.Channel1
			p.RegistArea = r.RegistArea
			p.Money = r.Money
			// 策略
			p.BSTimes = r.SevenStrategy.LKYH.BSTimes
			p.TriggerTimes = int64(r.SevenStrategy.LKYH.TriggerTimes)
			p.TZ = int64(r.SevenStrategy.LKYH.TZ)
			p.NZ = r.SevenStrategy.LKYH.NZ
			p.RZ = r.SevenStrategy.LKYH.RZ
			// 获取用户7UP游戏总汇数据
			if t, ok := upPlayers[userid]; ok {
				p.Rounds = int64(t.BetRounds)
				p.Bets = t.Bets
				p.WinRounds = int64(t.WinRounds)
				p.WinBets = t.Wins
				p.LoseBets = t.Loses
				p.LoseRounds = int64(t.LoseRounds)
			}
		}
	}

	for _, p := range upPlayers {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)

		// 保存数据
		addErr := StatisticsService.AddOrUpdateUpdownPlayerStat(p)
		if addErr != nil {
			beego.Error("UpdownPlayerStat fail err: ", addErr)
		}
	}

	for _, item := range upXxsc {
		err := StatisticsService.AddOrUpdateUpdownXxscStat(item)
		if err != nil {
			beego.Error("UpdownXxscStat fail err: ", err)
		}
	}

	for _, item := range upQsbns {
		if item.BeforeBackRateCount > 0 {
			item.BeforeBackRate = item.BeforeBackRateSum / item.BeforeBackRateCount
		}
		if item.AfterBackRateCount > 0 {
			item.AfterBackRate = item.AfterBackRateSum / item.AfterBackRateCount
		}
		err := StatisticsService.AddOrUpdateUpdownQsbnStat(item)
		if err != nil {
			beego.Error("UpdownQsbnStat fail err: ", err)
		}
	}

	for _, item := range upLkyh {
		// 局均码
		item.RoundBetAvg = item.Bets / int64(item.Rounds)
		err := StatisticsService.AddOrUpdateUpdownLkyhStat(item)
		if err != nil {
			beego.Error("UpdownLkyhStat fail err: ", err)
		}
	}
}

func UpdownStatsCK(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
	date1, _ := utils.Unix(today)
	gtype := int(pb.SEVEN)

	sql1 := `
		select userid,gtype,begin_time,end_time, win_type, score, bet_amount, up_seat_bets,
			up_winner,up_point_value,up_bets,up_player_win,up_strategy_id,up_is_suppress,up_userid,up_dragon,
			up_tiger,up_tie,up_observe,up_result,up_win,up_before_score,up_after_score,up_before_cash,up_after_cash,
			up_before_bonus,up_after_bonus,up_cash_ming_tax,up_bonus_ming_tax,up_cash_an_tax,up_bonus_an_tax,up_before_back_rate,up_after_back_rate 
		from game.col_detail cd FINAL where robot = false and gtype = ? and begin_time between ? and ?
	`
	var r1 []entity.DetailCK
	var args1 []any
	args1 = append(args1, gtype)
	args1 = append(args1, startTime.Unix())
	args1 = append(args1, endTime.Unix())
	err = ck.Select(&r1, sql1, args1...)
	if err != nil {
		beego.Warning("UpdownStats error1:", err)
		return
	}
	if len(r1) == 0 {
		beego.Info("no 7updown detail data")
		return
	}

	// 查询当天注册用户
	var r2 []string
	var args2 []any
	sql2 := `select userid FROM game.col_user cu FINAL where robot = false and simulation_robot = false and ctime BETWEEN ? and ?`
	args2 = append(args2, startTime)
	args2 = append(args2, endTime)
	err = ck.Select(&r2, sql2, args2...)
	if err != nil {
		beego.Warning("UpdownStats error2:", err)
		return
	}
	var regUserids = make(map[string]bool)
	for _, r := range r2 {
		userid := r
		regUserids[userid] = true
	}

	// 7up玩家列表
	var upPlayerIds []string
	var upPlayers = make(map[string]*entity.UpdownPlayerStat)
	var upXxsc = make(map[string]*entity.UpdownXxscStat)
	var upQsbns = make(map[string]*entity.UpdownQsbnStat)
	var upLkyh = make(map[string]*entity.UpdownLkyhStat)
	// 玩家下注后分数
	var upBets = make(map[string][]*entity.ABAnqsUserBet)

	// 获取7up适配人群配置
	var str_info entity.SevenStrategy
	err = UPStrategies.Find(bson.M{}).One(&str_info)
	if err != nil {
		beego.Error("get 7updown strategy err: ", err)
	}

	for _, user := range r1 {
		stat, ok := upPlayers[user.UserId]
		if !ok {
			stat = &entity.UpdownPlayerStat{
				Id:      fmt.Sprintf("%d-%s", date1, user.UserId),
				Date:    date1,
				DateStr: today,
				Userid:  user.UserId,
				NewReg:  regUserids[user.UserId], // 老顾客新顾客
			}
			upPlayers[user.UserId] = stat
			upPlayerIds = append(upPlayerIds, user.UserId)
		}
		stat.AllRounds++
		stat.GameTimes += (user.EndTime - user.BeginTime)
		// 是否观察位
		if user.WinType == ck.WinTypeObserver {
			stat.ObserveRounds++
			continue
		}
		userBet := user.BetAmount
		stat.Bets += userBet
		if userBet > 0 {
			stat.BetRounds++
		}

		switch user.WinType {
		case ck.WinTypeWin:
			stat.WinRounds++
			stat.WinBets += userBet
			stat.Wins += user.Score
			stat.Cash += user.Score

		case ck.WinTypeLose:
			stat.LoseRounds++
			stat.LoseBets += userBet
			stat.Loses += user.Score
			stat.Cash += user.Score

		case ck.WinTypeTie:
			stat.TieRounds++
			stat.TieBets += userBet
		}

		// detail.LHDetail.Winner // 1龙,2虎,else和
		if userBet > 0 {
			switch user.UpWinner {
			case 1:
				stat.Dragons++
			case 2:
				stat.Tigers++
			default:
				stat.Ties++
			}
		}

		var betTypes int
		if user.UpDragon > 0 {
			stat.PlayerDragons++
			betTypes++
			if user.WinType == ck.WinTypeWin {
				stat.PlayerWinDragons++
			}
		}
		if user.UpTiger > 0 {
			stat.PlayerTigers++
			betTypes++
			if user.WinType == ck.WinTypeWin {
				stat.PlayerWinTigers++
			}
		}
		if user.UpTie > 0 {
			stat.PlayerTies++
			betTypes++
			if user.WinType == ck.WinTypeWin {
				stat.PlayerWinTies++
			}
		}
		if len(user.UpSeatBets) >= 2 {
			stat.PlayerMultis++
		}

		// if user.UpStrategyId != 0 {
		// 	switch user.UpStrategyId {
		// 	case 1:
		// 		xxsc, ok := upXxsc[user.UserId]
		// 		if !ok {
		// 			xxsc = &entity.UpdownXxscStat{
		// 				Id:      fmt.Sprintf("%d-%s", date1, user.UserId),
		// 				Date:    date1,
		// 				DateStr: today,
		// 				Userid:  user.UserId,
		// 			}
		// 			upXxsc[user.UserId] = xxsc
		// 		}
		// 		xxsc.StrategyPlayers = 1
		// 		xxsc.StrategyTimes++
		// 		xxsc.StrategyRounds++
		// 		if user.UpResult == "输" { // 扰动局
		// 			xxsc.StrategyDisturbRounds++
		// 		}
		// 		if betTypes > 1 {
		// 			xxsc.StrategyMultiRounds++
		// 		}
		// 		xxsc.StrategyBets += userBet
		// 		if user.UpResult == "赢" {
		// 			xxsc.StrategyWins += user.UpWin
		// 			xxsc.StrategyWinRounds++
		// 		}
		// 		if user.UpResult == "输" {
		// 			xxsc.StrategyLoses += user.UpWin
		// 		}

		// 	case 2:
		// 		qsbn, ok := upQsbns[user.UserId]

		// 		if !ok {
		// 			qsbn = &entity.UpdownQsbnStat{
		// 				Id:      fmt.Sprintf("%d-%s", date1, user.UserId),
		// 				Date:    date1,
		// 				DateStr: today,
		// 				Userid:  user.UserId,
		// 			}
		// 			upQsbns[user.UserId] = qsbn
		// 		}
		// 		qsbn.StrategyPlayers = 1
		// 		qsbn.StrategyTimes++
		// 		qsbn.BeforeBackRateSum += user.UpBeforeBackRate
		// 		qsbn.BeforeBackRateCount++
		// 		qsbn.AfterBackRateSum += user.UpAfterBackRate
		// 		qsbn.AfterBackRateCount++

		// 		if betTypes > 1 {
		// 			qsbn.StrategyMultiRounds++
		// 		}
		// 		qsbn.StrategyBets += userBet
		// 		if user.UpResult == "赢" {
		// 			qsbn.StrategyWins += user.UpWin
		// 		}
		// 		if user.UpResult == "输" {
		// 			qsbn.StrategyLoses += user.UpWin
		// 		}
		// 	case 3:
		// 		// 龙狂有祸
		// 		lkyh, ok := upLkyh[user.UserId]
		// 		if !ok {
		// 			lkyh = &entity.UpdownLkyhStat{
		// 				Id:      fmt.Sprintf("%d-%s", date1, user.UserId),
		// 				Date:    date1,
		// 				DateStr: today,
		// 				Userid:  user.UserId,
		// 				BSDays:  1,
		// 			}
		// 			upLkyh[user.UserId] = lkyh
		// 		}
		// 		lkyh.StrategyTimes++
		// 		// 判断是否存在压制局
		// 		if user.UpIsSuppress {
		// 			lkyh.YZRounds++
		// 			lkyh.YZNumber++
		// 			lkyh.YZBets += userBet
		// 		} else {
		// 			lkyh.BSBets += userBet
		// 		}

		// 		if betTypes > 1 {
		// 			lkyh.StrategyMultiRounds++
		// 		}
		// 		lkyh.StrategyBets += userBet
		// 		if user.UpResult == "赢" {
		// 			lkyh.StrategyWinRounds++
		// 			lkyh.StrategyWins += user.UpWin
		// 		}
		// 		if user.UpResult == "输" {
		// 			lkyh.StrategyLoseRounds++
		// 			lkyh.StrategyLoses += user.UpWin
		// 		}
		// 	}
		// }
		// 记录每个用户下注后金额
		score := user.UpBeforeScore - userBet
		betinfo := &entity.ABAnqsUserBet{
			Userid: user.UserId,
			Score:  score,
		}
		betlist, ok := upBets[user.UserId]
		if !ok {
			betlist := make([]*entity.ABAnqsUserBet, 0)
			upBets[user.UserId] = betlist
		}
		betlist = append(betlist, betinfo)
		upBets[user.UserId] = betlist
		// if str_info.QSBN.D > user.UpBeforeScore-userBet {
		// 	qsbn, ok := upQsbns[user.UserId]
		// 	if !ok {
		// 		qsbn = &entity.UpdownQsbnStat{
		// 			Id:      fmt.Sprintf("%d-%s", date1, user.UserId),
		// 			Date:    date1,
		// 			DateStr: today,
		// 			Userid:  user.UserId,
		// 		}
		// 		upQsbns[user.UserId] = qsbn
		// 	}
		// 	qsbn.AllInRounds++
		// }
	}

	// 查询渠道，账号类型，充值金额
	sql3 := `select userid,ad__bundle_id,regist_area,money,state,seven_round_bet,seven_xxsc_disturb_cool_down,seven_xxsc_win_length,seven_xxsc_evo,seven_xxsc_trigger_times,seven_xxsc_max_trigger_times,seven_qsbn_trigger_times,seven_lkyh_trigger_times,seven_lkyh_history_t,seven_lkyh_tz,seven_lkyh_rz,seven_lkyh_bs_cool_down,seven_lkyh_bs_rounds,seven_lkyh_suppress,seven_lkyh_nz,seven_lkyh_bstimes,seven_lkyh_doubleBetTimes from game.col_user cu FINAL where robot = false and simulation_robot = false and userid in ?`
	var args3 []any
	args3 = append(args3, upPlayerIds)
	var r3 []entity.PlayerUserCK
	err = ck.Select(&r3, sql3, args3...)
	if err != nil {
		beego.Warning("7updownStats error3:", err)
		return
	}
	for _, r := range r3 {
		userid := r.Userid
		if p, ok := upPlayers[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.RegistArea = r.RegistArea
			p.Money = r.Money
		}
		if p, ok := upXxsc[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.RegistArea = r.RegistArea
			p.Money = r.Money
		}
		if p, ok := upQsbns[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.RegistArea = r.RegistArea
			p.Money = r.Money
			userType := ConvertToUserType(int(r.Money), r.State)
			//计算allin局数
			allinRounds := 0
			if len(str_info.QSBN.D) > 0 {
				for t, d := range str_info.QSBN.D {
					if t == userType {
						if u, ok := upBets[userid]; ok {
							for _, t := range u {
								if d > t.Score {
									allinRounds++
								}
							}
						}
					}
				}
			}
			p.AllInRounds = int32(allinRounds)
		}

		// 龙狂有祸
		if p, ok := upLkyh[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.RegistArea = r.RegistArea
			p.Money = r.Money
			// 策略
			p.BSTimes = r.SevenLKYHBSTimes
			p.TriggerTimes = int64(r.SevenLKYHTriggerTimes)
			p.TZ = int64(r.SevenLKYHTZ)
			p.NZ = r.SevenLKYHNZ
			p.RZ = r.SevenLKYHRZ
			// 获取用户7UP游戏总汇数据
			if t, ok := upPlayers[userid]; ok {
				p.Rounds = int64(t.BetRounds)
				p.Bets = t.Bets
				p.WinRounds = int64(t.WinRounds)
				p.WinBets = t.Wins
				p.LoseBets = t.Loses
				p.LoseRounds = int64(t.LoseRounds)
			}
		}
	}

	for _, p := range upPlayers {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)

		// 保存数据
		addErr := StatisticsService.AddOrUpdateUpdownPlayerStat(p)
		if addErr != nil {
			beego.Error("UpdownPlayerStat fail err: ", addErr)
		}
	}

	for _, item := range upXxsc {
		err := StatisticsService.AddOrUpdateUpdownXxscStat(item)
		if err != nil {
			beego.Error("UpdownXxscStat fail err: ", err)
		}
	}

	for _, item := range upQsbns {
		if item.BeforeBackRateCount > 0 {
			item.BeforeBackRate = item.BeforeBackRateSum / item.BeforeBackRateCount
		}
		if item.AfterBackRateCount > 0 {
			item.AfterBackRate = item.AfterBackRateSum / item.AfterBackRateCount
		}
		err := StatisticsService.AddOrUpdateUpdownQsbnStat(item)
		if err != nil {
			beego.Error("UpdownQsbnStat fail err: ", err)
		}
	}

	for _, item := range upLkyh {
		// 局均码
		item.RoundBetAvg = item.Bets / int64(item.Rounds)
		err := StatisticsService.AddOrUpdateUpdownLkyhStat(item)
		if err != nil {
			beego.Error("UpdownLkyhStat fail err: ", err)
		}
	}
}

// ABStats AndarBahar游戏统计
func ABStats(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
	date1, _ := utils.Unix(today)

	m1 := bson.M{
		"gtype":    int(pb.ABAR),
		"end_time": bson.M{"$gt": startTime.Unix(), "$lt": endTime.Unix()},
		// "players":    bson.M{"$ne": ""},
	}

	var r1 []entity.Detail
	err := Details.Find(m1).All(&r1)
	if err != nil {
		beego.Warning("ABStats error1:", err)
		return
	}
	if len(r1) == 0 {
		beego.Warning("no AndarBahar detail data")
		return
	}

	// 查询当天注册用户
	m2 := []bson.M{
		{"$match": bson.M{
			"ctime": bson.M{"$gt": startTime, "$lt": endTime},
		}},
		{"$project": bson.M{"_id": "$_id"}},
	}
	var r2 []bson.M
	err = PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Warning("ABStats error2:", err)
		return
	}
	var regUserids = make(map[string]bool)
	for _, r := range r2 {
		userid := r["_id"].(string)
		regUserids[userid] = true
	}

	// ab玩家列表
	var abPlayerIds []string
	var abPlayers = make(map[string]*entity.ABPlayerStat)
	var abAnqs = make(map[string]*entity.ABAnqsStat)
	var abArty = make(map[string]*entity.ABArtyStat)
	// 玩家下注后分数
	var abAnqsBets = make(map[string][]*entity.ABAnqsUserBet)
	// 获取AB适配人群配置
	var str_info entity.ABStrategy
	err = ABStrategies.Find(bson.M{}).One(&str_info)
	if err != nil {
		beego.Error("get AndarBahar strategy err: ", err)
	}

	for _, detail := range r1 {
		if detail.ABDetail == nil {
			continue
		}
		for _, user := range detail.ABDetail.UserDetail {
			if len(user.Userid) >= 16 {
				// 人机id18位长度+
				continue
			}
			stat, ok := abPlayers[user.Userid]
			if !ok {
				stat = &entity.ABPlayerStat{
					Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
					Date:    date1,
					DateStr: today,
					Userid:  user.Userid,
					NewReg:  regUserids[user.Userid], // 老顾客新顾客
				}
				abPlayers[user.Userid] = stat
				abPlayerIds = append(abPlayerIds, user.Userid)
			}
			stat.AllRounds++
			stat.GameTimes += (detail.EndTime - detail.BeginTime)
			// 是否观察位
			if user.Observe {
				stat.ObserveRounds++
				continue
			}
			userBet := int64(0)
			for _, v := range user.SeatBets {
				userBet += v
			}
			stat.Bets += userBet
			if userBet > 0 {
				stat.BetRounds++
			}

			switch user.Result {
			case "赢":
				stat.WinRounds++
				stat.WinBets += userBet
				stat.Wins += user.Win
				stat.Cash += user.Win

			case "输":
				stat.LoseRounds++
				stat.LoseBets += userBet
				stat.Loses += user.Win
				stat.Cash += user.Win

			case "平":
				stat.TieRounds++
				stat.TieBets += userBet
			}

			// detail.LHDetail.Winner // 1龙,2虎,else和
			if userBet > 0 {
				switch detail.ABDetail.Winner {
				case 1:
					stat.SideWinner1++
				case 2:
					stat.SideWinner2++
				}
				switch detail.ABDetail.SideWinner {
				case 3:
					stat.SideWinner3++
				case 4:
					stat.SideWinner4++
				case 5:
					stat.SideWinner5++
				case 6:
					stat.SideWinner6++
				case 7:
					stat.SideWinner7++
				case 8:
					stat.SideWinner8++
				case 9:
					stat.SideWinner9++
				case 10:
					stat.SideWinner10++
				}
			}
			var betTypes int
			for k, s := range user.SeatBets {
				switch k {
				case "1":
					if s > 0 {
						stat.SeatBets1++
						betTypes++
						if user.Result == "赢" {
							stat.PlayerWinSeat1++
						}
					}
				case "2":
					if s > 0 {
						stat.SeatBets2++
						betTypes++
						if user.Result == "赢" {
							stat.PlayerWinSeat2++
						}
					}
				case "3":
					if s > 0 {
						stat.SeatBets3++
						betTypes++
						if user.Result == "赢" {
							stat.PlayerWinSeat3++
						}
					}
				case "4":
					if s > 0 {
						stat.SeatBets4++
						betTypes++
						if user.Result == "赢" {
							stat.PlayerWinSeat4++
						}
					}
				case "5":
					if s > 0 {
						stat.SeatBets5++
						betTypes++
						if user.Result == "赢" {
							stat.PlayerWinSeat5++
						}
					}
				case "6":
					if s > 0 {
						stat.SeatBets6++
						betTypes++
						if user.Result == "赢" {
							stat.PlayerWinSeat6++
						}
					}
				case "7":
					if s > 0 {
						stat.SeatBets7++
						betTypes++
						if user.Result == "赢" {
							stat.PlayerWinSeat7++
						}
					}
				case "8":
					if s > 0 {
						stat.SeatBets8++
						betTypes++
						if user.Result == "赢" {
							stat.PlayerWinSeat8++
						}
					}
				case "9":
					if s > 0 {
						stat.SeatBets9++
						betTypes++
						if user.Result == "赢" {
							stat.PlayerWinSeat9++
						}
					}
				case "10":
					if s > 0 {
						stat.SeatBets10++
						betTypes++
						if user.Result == "赢" {
							stat.PlayerWinSeat10++
						}
					}

				}
			}
			if betTypes >= 2 {
				stat.PlayerMultis++
			}

			if detail.ABDetail.StrategyId != 0 {
				switch detail.ABDetail.StrategyId {
				case 1:
					// 安能求死
					anqs, ok := abAnqs[user.Userid]
					if !ok {
						anqs = &entity.ABAnqsStat{
							Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
							Date:    date1,
							DateStr: today,
							Userid:  user.Userid,
						}
						abAnqs[user.Userid] = anqs
					}
					anqs.StrategyPlayers = 1
					anqs.StrategyTimes++
					anqs.StrategyRounds++
					if betTypes >= 2 {
						anqs.StrategyMultiRounds++
					}
					anqs.StrategyBets += userBet
					if user.Result == "赢" {
						anqs.StrategyWins += user.Win
						anqs.StrategyWinRounds++
					}
					if user.Result == "输" {
						anqs.StrategyLoses += user.Win
						anqs.StrategyLoseRounds++
					}

				case 2:
					// 安然躺赢
					arty, ok := abArty[user.Userid]
					if !ok {
						arty = &entity.ABArtyStat{
							Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
							Date:    date1,
							DateStr: today,
							Userid:  user.Userid,
						}
						abArty[user.Userid] = arty
					}
					arty.StrategyPlayers = 1
					arty.StrategyRounds++
					arty.AllEvoTimes++
					arty.StrategyTimes++
					if betTypes >= 2 {
						arty.StrategyMultiRounds++
					}
					arty.StrategyBets += userBet
					if user.Result == "赢" {
						arty.StrategyWins += user.Win
						arty.StrategyWinRounds++
					}
					if user.Result == "输" {
						arty.StrategyLoses += user.Win
						arty.StrategyLoseRounds++
					}
				}
			}
			// 记录每个用户下注后金额
			score := user.BeforeScore - userBet
			betinfo := &entity.ABAnqsUserBet{
				Userid: user.Userid,
				Score:  score,
			}
			betlist, ok := abAnqsBets[user.Userid]
			if !ok {
				betlist := make([]*entity.ABAnqsUserBet, 0)
				abAnqsBets[user.Userid] = betlist
			}
			betlist = append(betlist, betinfo)
			abAnqsBets[user.Userid] = betlist

			// if len(str_info.ANQS.D) > 0{
			// 	score := user.BeforeScore - userBet
			// 	for i := 0; i < count; i++ {

			// 	}
			// }

			// if  > user.BeforeScore-userBet {
			// 	qsbn, ok := upQsbns[user.Userid]
			// 	if !ok {
			// 		qsbn = &entity.UpdownQsbnStat{
			// 			Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
			// 			Date:    date1,
			// 			DateStr: today,
			// 			Userid:  user.Userid,
			// 		}
			// 		upQsbns[user.Userid] = qsbn
			// 	}
			// 	qsbn.AllInRounds++
			// }
		}
	}

	// 查询渠道，账号类型，充值金额
	m3 := []bson.M{
		{"$match": bson.M{
			"_id": bson.M{"$in": abPlayerIds},
		}},
		{"$project": bson.M{
			"_id":           "$_id",
			"ad__bundle_id": "$ad__bundle_id",
			"channel1":      "$channel1",
			"regist_area":   "$regist_area",
			"money":         "$money",
			"ab_strategy":   "$ab_strategy",
		}},
	}
	var r3 []entity.PlayerUser
	err = PlayerUsers.Pipe(m3).All(&r3)
	if err != nil {
		beego.Warning("ABStats error3:", err)
	}
	for _, r := range r3 {
		// userid := r["_id"].(string)
		// ad__bundle_id := r["ad__bundle_id"]
		// channel1 := r["channel1"]

		// regist_area := r["regist_area"].(int)
		// money := r["money"].(int)
		userid := r.Userid
		if p, ok := abPlayers[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.Channel1 = ""
			p.RegistArea = r.RegistArea
			p.Money = uint32(r.Money)
		}
		// 安能求死
		if p, ok := abAnqs[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.Channel1 = ""
			p.RegistArea = r.RegistArea
			p.Money = uint32(r.Money)

			userType := ConvertToUserType(int(r.Money), r.State)
			// 计算allin局数
			allinRounds := 0
			if len(str_info.ANQS.D) > 0 {
				for t, d := range str_info.ANQS.D {
					if t == userType {
						if u, ok := abAnqsBets[userid]; ok {
							for _, t := range u {
								if d > t.Score {
									allinRounds++
								}
							}
						}
					}
				}
			}
			p.AllInRounds = int32(allinRounds)
			// 获取用户AB游戏总汇数据
			if t, ok := abPlayers[userid]; ok {
				p.Rounds = int64(t.BetRounds)
				p.WinRounds = int64(t.WinRounds)
				p.Wins = t.Wins
				p.Loses = t.Loses
				p.LoseRounds = int64(t.LoseRounds)
			}
		}

		// 安然躺赢
		if p, ok := abArty[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.Channel1 = ""
			p.RegistArea = r.RegistArea
			p.Money = uint32(r.Money)

			if r.ABStrategy.ARTY.UZ != 0 {
				p.StrategyTimes = int32(r.ABStrategy.ARTY.UZ)
			}
			if r.ABStrategy.ARTY.U != 0 {
				p.TiggerTimes = r.ABStrategy.ARTY.U
			}
		}

		// if p, ok := upXxsc[userid]; ok {
		// 	if ad__bundle_id != nil {
		// 		p.AD_BundleId = ad__bundle_id.(string)
		// 	}
		// 	if channel1 != nil {
		// 		p.Channel1 = channel1.(string)
		// 	}
		// 	p.RegistArea = regist_area
		// 	p.Money = uint32(money)
		// }
		// if p, ok := upQsbns[userid]; ok {
		// 	if ad__bundle_id != nil {
		// 		p.AD_BundleId = ad__bundle_id.(string)
		// 	}
		// 	if channel1 != nil {
		// 		p.Channel1 = channel1.(string)
		// 	}
		// 	p.RegistArea = regist_area
		// 	p.Money = uint32(money)
		// }
	}

	for _, p := range abPlayers {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)

		// 保存数据
		addErr := StatisticsService.AddOrUpdateABPlayerStat(p)
		if addErr != nil {
			beego.Error("ABStats fail err: ", addErr)
		}
	}
	for _, item := range abAnqs {
		err := StatisticsService.AddOrUpdateABAnqsStat(item)
		if err != nil {
			beego.Error("ABAnqsStat fail err: ", err)
		}
	}
	for _, item := range abArty {
		err := StatisticsService.AddOrUpdateABArtyStat(item)
		if err != nil {
			beego.Error("ABArtyStat fail err: ", err)
		}
	}
}

// ABStats AndarBahar游戏统计
func ABStatsCK(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
	date1, _ := utils.Unix(today)
	gtype := int(pb.ABAR)

	sql1 := `select userid,gtype,begin_time,end_time,ab_joker,ab_jackpot_value,ab_a_cards,ab_b_cards,ab_winner,ab_side_winner,ab_bets,ab_player_win,ab_strategy_id,ab_is_random,ab_final_factor,ab_can_win_score,ab_userid,ab_seat_bets,ab_observe,ab_result,ab_win,ab_before_score,ab_after_score,ab_before_cash,ab_after_cash,ab_before_bonus,ab_after_bonus,ab_cash_ming_tax,ab_bonus_ming_tax,ab_cash_an_tax,ab_bonus_an_tax from game.col_detail cd FINAL where robot = false and gtype = ? and begin_time between ? and ?`
	var r1 []entity.DetailCK
	var args1 []any
	args1 = append(args1, gtype)
	args1 = append(args1, startTime.Unix())
	args1 = append(args1, endTime.Unix())
	err = ck.Select(&r1, sql1, args1...)
	if err != nil {
		beego.Warning("ABStats error1:", err)
		return
	}
	if len(r1) == 0 {
		beego.Warning("no AndarBahar detail data")
		return
	}

	// 查询当天注册用户
	var r2 []string
	var args2 []any
	sql2 := `select userid FROM game.col_user cu FINAL where robot = false and simulation_robot = false and ctime BETWEEN ? and ?`
	args2 = append(args2, startTime)
	args2 = append(args2, endTime)
	err = ck.Select(&r2, sql2, args2...)
	if err != nil {
		beego.Warning("ABStats error2:", err)
		return
	}
	var regUserids = make(map[string]bool)
	for _, r := range r2 {
		userid := r
		regUserids[userid] = true
	}

	// ab玩家列表
	var abPlayerIds []string
	var abPlayers = make(map[string]*entity.ABPlayerStat)
	var abAnqs = make(map[string]*entity.ABAnqsStat)
	var abArty = make(map[string]*entity.ABArtyStat)
	// 玩家下注后分数
	var abAnqsBets = make(map[string][]*entity.ABAnqsUserBet)
	// 获取AB适配人群配置
	var str_info entity.ABStrategy
	err = ABStrategies.Find(bson.M{}).One(&str_info)
	if err != nil {
		beego.Error("get AndarBahar strategy err: ", err)
	}

	for _, user := range r1 {
		if len(user.UserId) >= 16 {
			// 人机id18位长度+
			continue
		}

		stat, ok := abPlayers[user.UserId]
		if !ok {
			stat = &entity.ABPlayerStat{
				Id:      fmt.Sprintf("%d-%s", date1, user.UserId),
				Date:    date1,
				DateStr: today,
				Userid:  user.UserId,
				NewReg:  regUserids[user.UserId], // 老顾客新顾客
			}
			abPlayers[user.UserId] = stat
			abPlayerIds = append(abPlayerIds, user.UserId)
		}
		stat.AllRounds++
		stat.GameTimes += (user.EndTime - user.BeginTime)
		// 是否观察位
		if user.AbObserve {
			stat.ObserveRounds++
			continue
		}
		userBet := int64(0)
		for _, v := range user.AbSeatBets {
			userBet += v
		}
		stat.Bets += userBet
		if userBet > 0 {
			stat.BetRounds++
		}

		switch user.AbResult {
		case "赢":
			stat.WinRounds++
			stat.WinBets += userBet
			stat.Wins += user.AbWin
			stat.Cash += user.AbWin

		case "输":
			stat.LoseRounds++
			stat.LoseBets += userBet
			stat.Loses += user.AbWin
			stat.Cash += user.AbWin

		case "平":
			stat.TieRounds++
			stat.TieBets += userBet
		}

		// detail.LHDetail.Winner // 1龙,2虎,else和
		if userBet > 0 {
			switch user.AbWinner {
			case 1:
				stat.SideWinner1++
			case 2:
				stat.SideWinner2++
			}
			switch user.AbSideWinner {
			case 3:
				stat.SideWinner3++
			case 4:
				stat.SideWinner4++
			case 5:
				stat.SideWinner5++
			case 6:
				stat.SideWinner6++
			case 7:
				stat.SideWinner7++
			case 8:
				stat.SideWinner8++
			case 9:
				stat.SideWinner9++
			case 10:
				stat.SideWinner10++
			}
		}
		var betTypes int
		for k, s := range user.AbSeatBets {
			switch k {
			case "1":
				if s > 0 {
					stat.SeatBets1++
					betTypes++
					if user.AbResult == "赢" {
						stat.PlayerWinSeat1++
					}
				}
			case "2":
				if s > 0 {
					stat.SeatBets2++
					betTypes++
					if user.AbResult == "赢" {
						stat.PlayerWinSeat2++
					}
				}
			case "3":
				if s > 0 {
					stat.SeatBets3++
					betTypes++
					if user.AbResult == "赢" {
						stat.PlayerWinSeat3++
					}
				}
			case "4":
				if s > 0 {
					stat.SeatBets4++
					betTypes++
					if user.AbResult == "赢" {
						stat.PlayerWinSeat4++
					}
				}
			case "5":
				if s > 0 {
					stat.SeatBets5++
					betTypes++
					if user.AbResult == "赢" {
						stat.PlayerWinSeat5++
					}
				}
			case "6":
				if s > 0 {
					stat.SeatBets6++
					betTypes++
					if user.AbResult == "赢" {
						stat.PlayerWinSeat6++
					}
				}
			case "7":
				if s > 0 {
					stat.SeatBets7++
					betTypes++
					if user.AbResult == "赢" {
						stat.PlayerWinSeat7++
					}
				}
			case "8":
				if s > 0 {
					stat.SeatBets8++
					betTypes++
					if user.AbResult == "赢" {
						stat.PlayerWinSeat8++
					}
				}
			case "9":
				if s > 0 {
					stat.SeatBets9++
					betTypes++
					if user.AbResult == "赢" {
						stat.PlayerWinSeat9++
					}
				}
			case "10":
				if s > 0 {
					stat.SeatBets10++
					betTypes++
					if user.AbResult == "赢" {
						stat.PlayerWinSeat10++
					}
				}

			}
		}
		if betTypes >= 2 {
			stat.PlayerMultis++
		}

		if user.AbStrategyId != 0 {
			switch user.AbStrategyId {
			case 1:
				// 安能求死
				anqs, ok := abAnqs[user.UserId]
				if !ok {
					anqs = &entity.ABAnqsStat{
						Id:      fmt.Sprintf("%d-%s", date1, user.UserId),
						Date:    date1,
						DateStr: today,
						Userid:  user.UserId,
					}
					abAnqs[user.UserId] = anqs
				}
				anqs.StrategyPlayers = 1
				anqs.StrategyTimes++
				anqs.StrategyRounds++
				if betTypes >= 2 {
					anqs.StrategyMultiRounds++
				}
				anqs.StrategyBets += userBet
				if user.AbResult == "赢" {
					anqs.StrategyWins += user.AbWin
					anqs.StrategyWinRounds++
				}
				if user.AbResult == "输" {
					anqs.StrategyLoses += user.AbWin
					anqs.StrategyLoseRounds++
				}

			case 2:
				// 安然躺赢
				arty, ok := abArty[user.UserId]
				if !ok {
					arty = &entity.ABArtyStat{
						Id:      fmt.Sprintf("%d-%s", date1, user.UserId),
						Date:    date1,
						DateStr: today,
						Userid:  user.UserId,
					}
					abArty[user.UserId] = arty
				}
				arty.StrategyPlayers = 1
				arty.StrategyRounds++
				arty.AllEvoTimes++
				arty.StrategyTimes++
				if betTypes >= 2 {
					arty.StrategyMultiRounds++
				}
				arty.StrategyBets += userBet
				if user.AbResult == "赢" {
					arty.StrategyWins += user.AbWin
					arty.StrategyWinRounds++
				}
				if user.AbResult == "输" {
					arty.StrategyLoses += user.AbWin
					arty.StrategyLoseRounds++
				}
			}
		}
		// 记录每个用户下注后金额
		score := user.AbBeforeScore - userBet
		betinfo := &entity.ABAnqsUserBet{
			Userid: user.UserId,
			Score:  score,
		}
		betlist, ok := abAnqsBets[user.UserId]
		if !ok {
			betlist := make([]*entity.ABAnqsUserBet, 0)
			abAnqsBets[user.UserId] = betlist
		}
		betlist = append(betlist, betinfo)
		abAnqsBets[user.UserId] = betlist

	}

	// 查询渠道，账号类型，充值金额
	sql3 := `select userid,ad__bundle_id,regist_area,money,state,ab_round_bet,ab_anqs_trigger_times,ab_arty_u,ab_arty_uz,ab_arty_evoTimes,ab_arty_win_score from game.col_user cu FINAL where robot = false and simulation_robot = false and userid in ?`
	var args3 []any
	args3 = append(args3, abPlayerIds)
	var r3 []entity.PlayerUserCK
	err = ck.Select(&r3, sql3, args3...)
	if err != nil {
		beego.Warning("ABStats error3:", err)
	}
	for _, r := range r3 {
		userid := r.Userid
		if p, ok := abPlayers[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.Channel1 = ""
			p.RegistArea = r.RegistArea
			p.Money = uint32(r.Money)
		}
		// 安能求死
		if p, ok := abAnqs[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.Channel1 = ""
			p.RegistArea = r.RegistArea
			p.Money = uint32(r.Money)

			userType := ConvertToUserType(int(r.Money), r.State)
			// 计算allin局数
			allinRounds := 0
			if len(str_info.ANQS.D) > 0 {
				for t, d := range str_info.ANQS.D {
					if t == userType {
						if u, ok := abAnqsBets[userid]; ok {
							for _, t := range u {
								if d > t.Score {
									allinRounds++
								}
							}
						}
					}
				}
			}
			p.AllInRounds = int32(allinRounds)
			// 获取用户AB游戏总汇数据
			if t, ok := abPlayers[userid]; ok {
				p.Rounds = int64(t.BetRounds)
				p.WinRounds = int64(t.WinRounds)
				p.Wins = t.Wins
				p.Loses = t.Loses
				p.LoseRounds = int64(t.LoseRounds)
			}
		}

		// 安然躺赢
		if p, ok := abArty[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.Channel1 = ""
			p.RegistArea = r.RegistArea
			p.Money = uint32(r.Money)

			if r.ABARTYUZ != 0 {
				p.StrategyTimes = int32(r.ABARTYUZ)
			}
			if r.ABARTYU != 0 {
				p.TiggerTimes = r.ABARTYU
			}
		}
	}

	for _, p := range abPlayers {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)

		// 保存数据
		addErr := StatisticsService.AddOrUpdateABPlayerStat(p)
		if addErr != nil {
			beego.Error("ABStats fail err: ", addErr)
		}
	}
	for _, item := range abAnqs {
		err := StatisticsService.AddOrUpdateABAnqsStat(item)
		if err != nil {
			beego.Error("ABAnqsStat fail err: ", err)
		}
	}
	for _, item := range abArty {
		err := StatisticsService.AddOrUpdateABArtyStat(item)
		if err != nil {
			beego.Error("ABArtyStat fail err: ", err)
		}
	}
}

// CPStats 彩票游戏统计
func CPStats(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
	date1, _ := utils.Unix(today)

	m1 := bson.M{
		"gtype":    int(pb.LOTTERY),
		"end_time": bson.M{"$gt": startTime.Unix(), "$lt": endTime.Unix()},
		// "players":    bson.M{"$ne": ""},
	}

	var r1 []entity.Detail
	err := Details.Find(m1).All(&r1)
	if err != nil {
		beego.Warning("CPStats error1:", err)
		return
	}
	if len(r1) == 0 {
		beego.Warning("no CP detail data")
		return
	}

	// 查询当天注册用户
	m2 := []bson.M{
		{"$match": bson.M{
			"ctime": bson.M{"$gt": startTime, "$lt": endTime},
		}},
		{"$project": bson.M{"_id": "$_id"}},
	}
	var r2 []bson.M
	err = PlayerUsers.Pipe(m2).All(&r2)
	if err != nil {
		beego.Warning("CPStats error2:", err)
		return
	}
	var regUserids = make(map[string]bool)
	for _, r := range r2 {
		userid := r["_id"].(string)
		regUserids[userid] = true
	}

	// cp玩家列表
	var cpPlayerIds []string
	var cpPlayers = make(map[string]*entity.CPPlayerStat)
	var cpLwjy = make(map[string]*entity.CPLwjyStat)
	var cpLyqn = make(map[string]*entity.CPLyqnStat)
	// 玩家下注后分数
	var cpBets = make(map[string][]*entity.ABAnqsUserBet)
	// 获取CP适配人群配置
	var str_info entity.CPStrategy
	err = CPStrategies.Find(bson.M{}).One(&str_info)
	if err != nil {
		beego.Error("get CP strategy err: ", err)
	}

	for _, detail := range r1 {
		if detail.CPDetail == nil {
			continue
		}
		for _, user := range detail.CPDetail.UserDetail {
			if len(user.Userid) >= 16 {
				// 人机id18位长度+
				continue
			}
			stat, ok := cpPlayers[user.Userid]
			if !ok {
				stat = &entity.CPPlayerStat{
					Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
					Date:    date1,
					DateStr: today,
					Userid:  user.Userid,
					NewReg:  regUserids[user.Userid], // 老顾客新顾客
				}
				cpPlayers[user.Userid] = stat
				cpPlayerIds = append(cpPlayerIds, user.Userid)
			}
			stat.AllRounds++
			stat.GameTimes += (detail.EndTime - detail.BeginTime)
			// 是否观察位
			if user.Observe {
				stat.ObserveRounds++
				continue
			}
			userBet := int64(0)
			for _, v := range user.SeatBets {
				userBet += v
			}
			stat.Bets += userBet
			if userBet > 0 {
				stat.BetRounds++
			}

			switch user.Result {
			case "赢":
				stat.WinRounds++
				stat.WinBets += userBet
				stat.Wins += user.Win
				stat.Cash += user.Win

			case "输":
				stat.LoseRounds++
				stat.LoseBets += userBet
				stat.Loses += user.Win
				stat.Cash += user.Win

			case "平":
				stat.TieRounds++
				stat.TieBets += userBet
			}

			if userBet > 0 {
				switch detail.CPDetail.CardType {
				case 1:
					stat.SideWinner1++
				case 2:
					stat.SideWinner2++
				case 3:
					stat.SideWinner3++
				case 4:
					stat.SideWinner4++
				case 5:
					stat.SideWinner5++
				case 6:
					stat.SideWinner6++
				}
			}
			var betTypes int
			for k, s := range user.SeatBets {
				switch k {
				case "1":
					if s > 0 {
						stat.SeatBets1++
						betTypes++
						if user.Result == "赢" {
							stat.PlayerWinSeat1++
						}
					}
				case "2":
					if s > 0 {
						stat.SeatBets2++
						betTypes++
						if user.Result == "赢" {
							stat.PlayerWinSeat2++
						}
					}
				case "3":
					if s > 0 {
						stat.SeatBets3++
						betTypes++
						if user.Result == "赢" {
							stat.PlayerWinSeat3++
						}
					}
				case "4":
					if s > 0 {
						stat.SeatBets4++
						betTypes++
						if user.Result == "赢" {
							stat.PlayerWinSeat4++
						}
					}
				case "5":
					if s > 0 {
						stat.SeatBets5++
						betTypes++
						if user.Result == "赢" {
							stat.PlayerWinSeat5++
						}
					}
				case "6":
					if s > 0 {
						stat.SeatBets6++
						betTypes++
						if user.Result == "赢" {
							stat.PlayerWinSeat6++
						}
					}

				}
			}
			if betTypes >= 2 {
				stat.PlayerMultis++
			}

			if detail.CPDetail.StrategyId != 0 {
				switch detail.CPDetail.StrategyId {
				case 1:
					// 来玩就赢
					anqs, ok := cpLwjy[user.Userid]
					if !ok {
						anqs = &entity.CPLwjyStat{
							Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
							Date:    date1,
							DateStr: today,
							Userid:  user.Userid,
						}
						cpLwjy[user.Userid] = anqs
					}
					anqs.StrategyPlayers = 1
					anqs.AllEvoTimes++
					// anqs.StrategyRounds++
					if betTypes >= 2 {
						anqs.StrategyMultiRounds++
					}
					anqs.StrategyBets += userBet
					if user.Result == "赢" {
						anqs.StrategyWins += user.Win
						anqs.StrategyWinRounds++
					}
					if user.Result == "输" {
						anqs.StrategyLoses += user.Win
						anqs.StrategyLoseRounds++
					}

				case 2:
					// 来易去难
					arty, ok := cpLyqn[user.Userid]
					if !ok {
						arty = &entity.CPLyqnStat{
							Id:      fmt.Sprintf("%d-%s", date1, user.Userid),
							Date:    date1,
							DateStr: today,
							Userid:  user.Userid,
						}
						cpLyqn[user.Userid] = arty
					}
					arty.StrategyPlayers = 1
					arty.StrategyRounds++
					// arty.AllEvoTimes++
					// arty.StrategyTimes++
					if betTypes >= 2 {
						arty.StrategyMultiRounds++
					}
					arty.StrategyBets += userBet
					if user.Result == "赢" {
						arty.StrategyWins += user.Win
						arty.StrategyWinRounds++
					}
					if user.Result == "输" {
						arty.StrategyLoses += user.Win
						arty.StrategyLoseRounds++
					}
				}
			}
			// 记录每个用户下注后金额
			score := user.BeforeScore - userBet
			betinfo := &entity.ABAnqsUserBet{
				Userid: user.Userid,
				Score:  score,
			}
			betlist, ok := cpBets[user.Userid]
			if !ok {
				betlist := make([]*entity.ABAnqsUserBet, 0)
				cpBets[user.Userid] = betlist
			}
			betlist = append(betlist, betinfo)
			cpBets[user.Userid] = betlist
		}
	}

	// 查询渠道，账号类型，充值金额
	m3 := []bson.M{
		{"$match": bson.M{
			"_id": bson.M{"$in": cpPlayerIds},
		}},
		{"$project": bson.M{
			"_id":           "$_id",
			"ad__bundle_id": "$ad__bundle_id",
			"channel1":      "$channel1",
			"regist_area":   "$regist_area",
			"money":         "$money",
			"cp_strategy":   "$cp_strategy",
		}},
	}
	var r3 []entity.PlayerUser
	err = PlayerUsers.Pipe(m3).All(&r3)
	if err != nil {
		beego.Warning("CPStats error3:", err)
	}
	for _, r := range r3 {
		// userid := r["_id"].(string)
		// ad__bundle_id := r["ad__bundle_id"]
		// channel1 := r["channel1"]

		// regist_area := r["regist_area"].(int)
		// money := r["money"].(int)
		userid := r.Userid
		if p, ok := cpPlayers[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.Channel1 = ""
			p.RegistArea = r.RegistArea
			p.Money = uint32(r.Money)
		}
		// 来玩就赢
		if p, ok := cpLwjy[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.Channel1 = ""
			p.RegistArea = r.RegistArea
			p.Money = uint32(r.Money)
			if r.CPStrategy.LWJY.U != 0 {
				p.TiggerTimes = r.CPStrategy.LWJY.U
			}
			if r.CPStrategy.LWJY.UZ != 0 {
				p.StrategyTimes = int32(r.CPStrategy.LWJY.UZ)
			}
			// userType := ConvertToUserType(int(r.Money), r.State)
			// 计算allin局数
			// allinRounds := 0
			// if len(str_info.LWJY.D) > 0 {
			// 	for t, d := range str_info.ANQS.D {
			// 		if t == userType {
			// 			if u, ok := abAnqsBets[userid]; ok {
			// 				for _, t := range u {
			// 					if d > t.Score {
			// 						allinRounds++
			// 					}
			// 				}
			// 			}
			// 		}
			// 	}
			// }
			// p.AllInRounds = int32(allinRounds)
			// // 获取用户CP游戏总汇数据
			// if t, ok := cpPlayers[userid]; ok {
			// 	p.Rounds = int64(t.BetRounds)
			// 	p.WinRounds = int64(t.WinRounds)
			// 	p.Wins = t.Wins
			// 	p.Loses = t.Loses
			// 	p.LoseRounds = int64(t.LoseRounds)
			// }
		}

		// 安然躺赢
		if p, ok := cpLyqn[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.Channel1 = ""
			p.RegistArea = r.RegistArea
			p.Money = uint32(r.Money)

			userType := ConvertToUserType(int(r.Money), r.State)
			//计算allin局数
			allinRounds := 0
			if len(str_info.LYQN.D) > 0 {
				for t, d := range str_info.LYQN.D {
					if t == userType {
						if u, ok := cpBets[userid]; ok {
							for _, t := range u {
								if d > t.Score {
									allinRounds++
								}
							}
						}
					}
				}
			}
			p.AllInRounds = int32(allinRounds)
			// 获取用户CP游戏总汇数据
			if t, ok := cpPlayers[userid]; ok {
				p.BetRounds = t.BetRounds
				p.WinRounds = t.WinRounds
				p.Wins = t.Wins
				p.Loses = t.Loses
				p.LoseRounds = t.LoseRounds
			}
		}
	}

	for _, p := range cpPlayers {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)

		// 保存数据
		addErr := StatisticsService.AddOrUpdateCPPlayerStat(p)
		if addErr != nil {
			beego.Error("CPStats fail err: ", addErr)
		}
	}
	for _, item := range cpLwjy {
		err := StatisticsService.AddOrUpdateCPLwjyStat(item)
		if err != nil {
			beego.Error("CPLwjyStat fail err: ", err)
		}
	}
	for _, item := range cpLyqn {
		err := StatisticsService.AddOrUpdateCPLyqnStat(item)
		if err != nil {
			beego.Error("CPLyqnStat fail err: ", err)
		}
	}
}

// CPStats 彩票游戏统计
func CPStatsCK(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
	date1, _ := utils.Unix(today)
	gtype := int(pb.LOTTERY)

	sql1 := `select userid,gtype,begin_time,end_time,cp_jackpot_output,cp_card_type,cp_a_cards,cp_winner,cp_bets,cp_player_win,cp_strategy_id,cp_is_random,cp_final_factor,cp_can_win_score,cp_userid,cp_seat_bets,cp_result,cp_win,cp_observe,cp_before_score,cp_after_score,cp_before_cash,cp_after_cash,cp_before_bonus,cp_after_bonus,cp_cash_ming_tax,cp_bonus_ming_tax,cp_cash_an_tax,cp_bonus_an_tax from game.col_detail cd FINAL where robot = false and gtype = ? and begin_time between ? and ?`
	var r1 []entity.DetailCK
	var args1 []any
	args1 = append(args1, gtype)
	args1 = append(args1, startTime.Unix())
	args1 = append(args1, endTime.Unix())
	err = ck.Select(&r1, sql1, args1...)
	if err != nil {
		beego.Warning("CPStats error1:", err)
		return
	}
	if len(r1) == 0 {
		beego.Info("no CP detail data")
		return
	}

	// 查询当天注册用户
	var r2 []string
	var args2 []any
	sql2 := `select userid FROM game.col_user cu FINAL where robot = false and simulation_robot = false and ctime BETWEEN ? and ?`
	args2 = append(args2, startTime)
	args2 = append(args2, endTime)
	err = ck.Select(&r2, sql2, args2...)
	if err != nil {
		beego.Warning("CPStats error2:", err)
		return
	}
	var regUserids = make(map[string]bool)
	for _, r := range r2 {
		userid := r
		regUserids[userid] = true
	}

	// cp玩家列表
	var cpPlayerIds []string
	var cpPlayers = make(map[string]*entity.CPPlayerStat)
	var cpLwjy = make(map[string]*entity.CPLwjyStat)
	var cpLyqn = make(map[string]*entity.CPLyqnStat)
	// 玩家下注后分数
	var cpBets = make(map[string][]*entity.ABAnqsUserBet)
	// 获取CP适配人群配置
	var str_info entity.CPStrategy
	err = CPStrategies.Find(bson.M{}).One(&str_info)
	if err != nil {
		beego.Error("get CP strategy err: ", err)
	}

	for _, user := range r1 {
		if len(user.UserId) >= 16 {
			// 人机id18位长度+
			continue
		}
		stat, ok := cpPlayers[user.UserId]
		if !ok {
			stat = &entity.CPPlayerStat{
				Id:      fmt.Sprintf("%d-%s", date1, user.UserId),
				Date:    date1,
				DateStr: today,
				Userid:  user.UserId,
				NewReg:  regUserids[user.UserId], // 老顾客新顾客
			}
			cpPlayers[user.UserId] = stat
			cpPlayerIds = append(cpPlayerIds, user.UserId)
		}
		stat.AllRounds++
		stat.GameTimes += (user.EndTime - user.BeginTime)
		// 是否观察位
		if user.CpObserve {
			stat.ObserveRounds++
			continue
		}
		userBet := int64(0)
		for _, v := range user.CpSeatBets {
			userBet += v
		}
		stat.Bets += userBet
		if userBet > 0 {
			stat.BetRounds++
		}

		switch user.CpResult {
		case "赢":
			stat.WinRounds++
			stat.WinBets += userBet
			stat.Wins += user.CpWin
			stat.Cash += user.CpWin

		case "输":
			stat.LoseRounds++
			stat.LoseBets += userBet
			stat.Loses += user.CpWin
			stat.Cash += user.CpWin

		case "平":
			stat.TieRounds++
			stat.TieBets += userBet
		}
		if userBet > 0 {
			switch user.CpCardType {
			case 1:
				stat.SideWinner1++
			case 2:
				stat.SideWinner2++
			case 3:
				stat.SideWinner3++
			case 4:
				stat.SideWinner4++
			case 5:
				stat.SideWinner5++
			case 6:
				stat.SideWinner6++
			}
		}
		var betTypes int
		for k, s := range user.CpSeatBets {
			switch k {
			case "1":
				if s > 0 {
					stat.SeatBets1++
					betTypes++
					if user.CpResult == "赢" && user.CpCardType == 1 {
						stat.PlayerWinSeat1++
					}
				}
			case "2":
				if s > 0 {
					stat.SeatBets2++
					betTypes++
					if user.CpResult == "赢" && user.CpCardType == 2 {
						stat.PlayerWinSeat2++
					}
				}
			case "3":
				if s > 0 {
					stat.SeatBets3++
					betTypes++
					if user.CpResult == "赢" && user.CpCardType == 3 {
						stat.PlayerWinSeat3++
					}
				}
			case "4":
				if s > 0 {
					stat.SeatBets4++
					betTypes++
					if user.CpResult == "赢" && user.CpCardType == 4 {
						stat.PlayerWinSeat4++
					}
				}
			case "5":
				if s > 0 {
					stat.SeatBets5++
					betTypes++
					if user.CpResult == "赢" && user.CpCardType == 5 {
						stat.PlayerWinSeat5++
					}
				}
			case "6":
				if s > 0 {
					stat.SeatBets6++
					betTypes++
					if user.CpResult == "赢" && user.CpCardType == 6 {
						stat.PlayerWinSeat6++
					}
				}

			}
		}
		if betTypes >= 2 {
			stat.PlayerMultis++
		}

		if user.CpStrategyId != 0 {
			switch user.CpStrategyId {
			case 1:
				// 来玩就赢
				anqs, ok := cpLwjy[user.UserId]
				if !ok {
					anqs = &entity.CPLwjyStat{
						Id:      fmt.Sprintf("%d-%s", date1, user.UserId),
						Date:    date1,
						DateStr: today,
						Userid:  user.UserId,
					}
					cpLwjy[user.UserId] = anqs
				}
				anqs.StrategyPlayers = 1
				anqs.AllEvoTimes++
				// anqs.StrategyRounds++
				if betTypes >= 2 {
					anqs.StrategyMultiRounds++
				}
				anqs.StrategyBets += userBet
				if user.CpResult == "赢" {
					anqs.StrategyWins += user.CpWin
					anqs.StrategyWinRounds++
				}
				if user.CpResult == "输" {
					anqs.StrategyLoses += user.CpWin
					anqs.StrategyLoseRounds++
				}

			case 2:
				// 来易去难
				arty, ok := cpLyqn[user.UserId]
				if !ok {
					arty = &entity.CPLyqnStat{
						Id:      fmt.Sprintf("%d-%s", date1, user.UserId),
						Date:    date1,
						DateStr: today,
						Userid:  user.UserId,
					}
					cpLyqn[user.UserId] = arty
				}
				arty.StrategyPlayers = 1
				arty.StrategyRounds++
				// arty.AllEvoTimes++
				// arty.StrategyTimes++
				if betTypes >= 2 {
					arty.StrategyMultiRounds++
				}
				arty.StrategyBets += userBet
				if user.CpResult == "赢" {
					arty.StrategyWins += user.CpWin
					arty.StrategyWinRounds++
				}
				if user.CpResult == "输" {
					arty.StrategyLoses += user.CpWin
					arty.StrategyLoseRounds++
				}
			}
		}
		// 记录每个用户下注后金额
		score := user.CpBeforeScore - userBet
		betinfo := &entity.ABAnqsUserBet{
			Userid: user.UserId,
			Score:  score,
		}
		betlist, ok := cpBets[user.UserId]
		if !ok {
			betlist := make([]*entity.ABAnqsUserBet, 0)
			cpBets[user.UserId] = betlist
		}
		betlist = append(betlist, betinfo)
		cpBets[user.UserId] = betlist
	}

	// 查询渠道，账号类型，充值金额
	sql3 := `select userid,ad__bundle_id,regist_area,money,state,cp_round_bet,trigger_times,cp_lwjy_u,cp_lwjy_uz,cp_lwjy_evoTimes,cp_lwjy_win_score from game.col_user cu FINAL where robot = false and simulation_robot = false and userid in ?`
	var args3 []any
	args3 = append(args3, cpPlayerIds)
	var r3 []entity.PlayerUserCK
	err = ck.Select(&r3, sql3, args3...)
	if err != nil {
		beego.Warning("CPStats error3:", err)
		return
	}
	for _, r := range r3 {
		userid := r.Userid
		if p, ok := cpPlayers[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.Channel1 = ""
			p.RegistArea = r.RegistArea
			p.Money = uint32(r.Money)
		}
		// 来玩就赢
		if p, ok := cpLwjy[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.Channel1 = ""
			p.RegistArea = r.RegistArea
			p.Money = uint32(r.Money)
			if r.CPLWJYU != 0 {
				p.TiggerTimes = r.CPLWJYU
			}
			if r.CPLWJYUZ != 0 {
				p.StrategyTimes = int32(r.CPLWJYUZ)
			}
		}

		// 安然躺赢
		if p, ok := cpLyqn[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.Channel1 = ""
			p.RegistArea = r.RegistArea
			p.Money = uint32(r.Money)

			userType := ConvertToUserType(int(r.Money), r.State)
			//计算allin局数
			allinRounds := 0
			if len(str_info.LYQN.D) > 0 {
				for t, d := range str_info.LYQN.D {
					if t == userType {
						if u, ok := cpBets[userid]; ok {
							for _, t := range u {
								if d > t.Score {
									allinRounds++
								}
							}
						}
					}
				}
			}
			p.AllInRounds = int32(allinRounds)
			// 获取用户CP游戏总汇数据
			if t, ok := cpPlayers[userid]; ok {
				p.BetRounds = t.BetRounds
				p.WinRounds = t.WinRounds
				p.Wins = t.Wins
				p.Loses = t.Loses
				p.LoseRounds = t.LoseRounds
			}
		}
	}

	for _, p := range cpPlayers {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)

		// 保存数据
		addErr := StatisticsService.AddOrUpdateCPPlayerStat(p)
		if addErr != nil {
			beego.Error("CPStats fail err: ", addErr)
		}
	}
	for _, item := range cpLwjy {
		err := StatisticsService.AddOrUpdateCPLwjyStat(item)
		if err != nil {
			beego.Error("CPLwjyStat fail err: ", err)
		}
	}
	for _, item := range cpLyqn {
		err := StatisticsService.AddOrUpdateCPLyqnStat(item)
		if err != nil {
			beego.Error("CPLyqnStat fail err: ", err)
		}
	}
}

// RBStats 红黑大战游戏统计
func RBStatsCK(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	startTime := utils.Str2TimeZone(fmt.Sprintf("%s 00:00:00", today), locationName)
	endTime := utils.Str2TimeZone(fmt.Sprintf("%s 23:59:59", today), locationName)
	date1, _ := utils.Unix(today)
	gtype := int(pb.REDBLACK)

	sql1 := `select userid,gtype,begin_time,end_time,rb_card_type,rb_a_cards,rb_winner,rb_bets,rb_player_win,rb_strategy_id,rb_can_win_score,rb_userid,rb_seat_bets,rb_result,rb_win,rb_observe,rb_before_score,rb_after_score,rb_before_cash,rb_after_cash,rb_before_bonus,rb_after_bonus,rb_cash_ming_tax,rb_bonus_ming_tax,rb_cash_an_tax,rb_bonus_an_tax from game.col_detail cd FINAL where robot = false and gtype = ? and begin_time between ? and ?`
	var r1 []entity.DetailCK
	var args1 []any
	args1 = append(args1, gtype)
	args1 = append(args1, startTime.Unix())
	args1 = append(args1, endTime.Unix())
	err = ck.Select(&r1, sql1, args1...)
	if err != nil {
		beego.Warning("RBStats error1:", err)
		return
	}
	if len(r1) == 0 {
		beego.Info("no RB detail data")
		return
	}
	// 查询当天注册用户
	var r2 []string
	var args2 []any
	sql2 := `select userid FROM game.col_user cu FINAL where robot = false and simulation_robot = false and ctime BETWEEN ? and ?`
	args2 = append(args2, startTime)
	args2 = append(args2, endTime)
	err = ck.Select(&r2, sql2, args2...)
	if err != nil {
		beego.Warning("RBStats error2:", err)
		return
	}
	var regUserids = make(map[string]bool)
	for _, r := range r2 {
		userid := r
		regUserids[userid] = true
	}
	// 获取RB适配人群配置
	var str_info entity.RBStrategy
	err = RBStrategies.Find(bson.M{}).One(&str_info)
	if err != nil {
		beego.Error("get RB strategy err: ", err)
	}
	// RB 玩家列表
	var rbPlayerIds []string
	var rbPlayers = make(map[string]*entity.RBPlayerStat)
	var rbHydt = make(map[string]*entity.RBHydtStat)
	var rbJcfs = make(map[string]*entity.RBJcfsStat)
	// 玩家下注后分数
	var rbBets = make(map[string][]*entity.ABAnqsUserBet)
	for _, user := range r1 {
		if len(user.UserId) >= 16 {
			// 人机id18位长度+
			continue
		}
		stat, ok := rbPlayers[user.UserId]
		if !ok {
			stat = &entity.RBPlayerStat{
				Id:      fmt.Sprintf("%d-%s", date1, user.UserId),
				Date:    date1,
				DateStr: today,
				Userid:  user.UserId,
				NewReg:  regUserids[user.UserId], // 老顾客新顾客
			}
			rbPlayers[user.UserId] = stat
			rbPlayerIds = append(rbPlayerIds, user.UserId)
		}
		stat.AllRounds++
		stat.GameTimes += (user.EndTime - user.BeginTime)
		// 是否观察位
		if user.RBObserve {
			stat.ObserveRounds++
			continue
		}
		userBet := int64(0)
		for _, v := range user.RBSeatBets {
			userBet += v
		}
		stat.Bets += userBet
		if userBet > 0 {
			stat.BetRounds++
		}

		switch user.RBResult {
		case "赢":
			stat.WinRounds++
			stat.WinBets += userBet
			stat.Wins += user.RBWin
			stat.Cash += user.RBWin

		case "输":
			stat.LoseRounds++
			stat.LoseBets += userBet
			stat.Loses += user.RBWin
			stat.Cash += user.RBWin

		case "平":
			stat.TieRounds++
			stat.TieBets += userBet
		}
		WinId := -1
		MaxType := 0
		isLuckyshot := false
		if userBet > 0 {
			for _, v := range user.RBWinner {
				switch v {
				case 0:
					// 幸运一击
					stat.SideWinner3++
					isLuckyshot = true
				case 1:
					// 红
					stat.SideWinner2++
					WinId = 0
				case 2:
					// 黑
					stat.SideWinner1++
					WinId = 1
				}
			}
			for _, c := range user.RBCardType {
				if isLuckyshot {
					switch c {
					case 2:
						stat.SideWinner4++
					case 3:
						stat.SideWinner4++
					case 4:
						stat.SideWinner5++
					case 5:
						stat.SideWinner6++
					case 6:
						stat.SideWinner7++
					case 7:
						stat.SideWinner8++
					}
					if isLuckyshot && MaxType < int(c) {
						MaxType = int(c)
					}
				}

				// if k == WinId {
				// 	// 红
				// 	switch c {
				// 	case 2:
				// 		stat.SideWinner4++
				// 	case 3:
				// 		stat.SideWinner4++
				// 	case 4:
				// 		stat.SideWinner5++
				// 	case 5:
				// 		stat.SideWinner6++
				// 	case 6:
				// 		stat.SideWinner7++
				// 	case 7:
				// 		stat.SideWinner8++
				// 	}
				// 	if isLuckyshot && MaxType < int(c) {
				// 		MaxType = int(c)
				// 	}
				// }
				// if k == WinId {
				// 	// 黑
				// 	switch c {
				// 	case 2:
				// 		stat.SideWinner4++
				// 	case 3:
				// 		stat.SideWinner4++
				// 	case 4:
				// 		stat.SideWinner5++
				// 	case 5:
				// 		stat.SideWinner6++
				// 	case 6:
				// 		stat.SideWinner7++
				// 	case 7:
				// 		stat.SideWinner8++
				// 	}
				// 	if isLuckyshot && MaxType < int(c) {
				// 		MaxType = int(c)
				// 	}
				// }
			}
		}
		var betTypes int
		for k, s := range user.RBSeatBets {
			switch k {
			case "0":
				if s > 0 {
					stat.SeatBets3++
					betTypes++
					if user.RBResult == "赢" {
						stat.PlayerWinSeat3++
						if WinId == 1 {
							switch MaxType {
							case 2:
								stat.PlayerWinSeat4++
							case 3:
								stat.PlayerWinSeat4++
							case 4:
								stat.PlayerWinSeat5++
							case 5:
								stat.PlayerWinSeat6++
							case 6:
								stat.PlayerWinSeat7++
							case 7:
								stat.PlayerWinSeat8++
							}
						}
					}
				}
			case "1":
				if s > 0 {
					stat.SeatBets2++
					betTypes++
					if user.RBResult == "赢" {
						stat.PlayerWinSeat2++
					}
				}
			case "2":
				if s > 0 {
					stat.SeatBets1++
					betTypes++
					if user.RBResult == "赢" {
						stat.PlayerWinSeat1++
					}
				}

			}
		}
		if betTypes >= 2 {
			stat.PlayerMultis++
		}
		if user.RBStrategyId != 0 {
			switch user.RBStrategyId {
			case 1:
				// RB 红运当头
				hydt, ok := rbHydt[user.UserId]
				if !ok {
					hydt = &entity.RBHydtStat{
						Id:      fmt.Sprintf("%d-%s", date1, user.UserId),
						Date:    date1,
						DateStr: today,
						Userid:  user.UserId,
					}
					rbHydt[user.UserId] = hydt
				}
				hydt.StrategyPlayers = 1
				hydt.AllEvoTimes++
				// anqs.StrategyRounds++
				if betTypes >= 2 {
					hydt.StrategyMultiRounds++
				}
				hydt.StrategyBets += userBet
				if user.RBResult == "赢" {
					hydt.StrategyWins += user.RBWin
					hydt.StrategyWinRounds++
				}
				if user.RBResult == "输" {
					hydt.StrategyLoses += user.RBWin
					hydt.StrategyLoseRounds++
				}
			case 2:
				// RB 绝处逢生
				jcfs, ok := rbJcfs[user.UserId]
				if !ok {
					jcfs = &entity.RBJcfsStat{
						Id:      fmt.Sprintf("%d-%s", date1, user.UserId),
						Date:    date1,
						DateStr: today,
						Userid:  user.UserId,
					}
					rbJcfs[user.UserId] = jcfs
				}
				jcfs.StrategyPlayers = 1
				jcfs.StrategyRounds++
				// arty.AllEvoTimes++
				// arty.StrategyTimes++
				if betTypes >= 2 {
					jcfs.StrategyMultiRounds++
				}
				jcfs.StrategyBets += userBet
				if user.RBResult == "赢" {
					jcfs.StrategyWins += user.RBWin
					jcfs.StrategyWinRounds++
				}
				if user.RBResult == "输" {
					jcfs.StrategyLoses += user.RBWin
					jcfs.StrategyLoseRounds++
				}

			}
		}
		// 记录每个用户下注后金额
		score := user.RBBeforeScore - userBet
		betinfo := &entity.ABAnqsUserBet{
			Userid: user.UserId,
			Score:  score,
		}
		betlist, ok := rbBets[user.UserId]
		if !ok {
			betlist := make([]*entity.ABAnqsUserBet, 0)
			rbBets[user.UserId] = betlist
		}
		betlist = append(betlist, betinfo)
		rbBets[user.UserId] = betlist
	}

	// 查询渠道，账号类型，充值金额
	sql3 := `select userid,ad__bundle_id,regist_area,money,state,round_bet,rb_jcfs_trigger_times,rb_hydt_u,rb_hydt_uz,rb_hydt_evoTimes,rb_hydt_win_score from game.col_user cu FINAL where robot = false and simulation_robot = false and userid in ?`
	var args3 []any
	args3 = append(args3, rbPlayerIds)
	var r3 []entity.PlayerUserCK
	err = ck.Select(&r3, sql3, args3...)
	if err != nil {
		beego.Warning("RBStats error3:", err)
		return
	}
	for _, r := range r3 {

		userid := r.Userid
		if p, ok := rbPlayers[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.Channel1 = ""
			p.RegistArea = r.RegistArea
			p.Money = uint32(r.Money)
		}
		// 红运当头
		if p, ok := rbHydt[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.Channel1 = ""
			p.RegistArea = r.RegistArea
			p.Money = uint32(r.Money)
			if r.RBHYDTU != 0 {
				p.TiggerTimes = r.RBHYDTU
			}
			if r.RBHYDTUZ != 0 {
				p.StrategyTimes = int32(r.RBHYDTUZ)
			}

		}

		// 绝处逢生
		if p, ok := rbJcfs[userid]; ok {
			p.AD_BundleId = r.AD_BundleId
			p.Channel1 = ""
			p.RegistArea = r.RegistArea
			p.Money = uint32(r.Money)

			userType := ConvertToUserType(int(r.Money), r.State)
			//计算allin局数
			allinRounds := 0
			if len(str_info.JCFS.D) > 0 {
				for t, d := range str_info.JCFS.D {
					if t == userType {
						if u, ok := rbBets[userid]; ok {
							for _, t := range u {
								if d > t.Score {
									allinRounds++
								}
							}
						}
					}
				}
			}
			p.AllInRounds = int32(allinRounds)
			// 获取用户RB游戏总汇数据
			if t, ok := rbPlayers[userid]; ok {
				p.BetRounds = t.BetRounds
				p.WinRounds = t.WinRounds
				p.Wins = t.Wins
				p.Loses = t.Loses
				p.LoseRounds = t.LoseRounds
			}
		}
	}

	for _, p := range rbPlayers {
		// 局均码
		p.RoundBetAvg = p.Bets / int64(p.AllRounds)

		// 保存数据
		addErr := StatisticsService.AddOrUpdateRBPlayerStat(p)
		if addErr != nil {
			beego.Error("RBStats fail err: ", addErr)
		}
	}
	for _, item := range rbHydt {
		err := StatisticsService.AddOrUpdateRBHydtStat(item)
		if err != nil {
			beego.Error("RBHydtStats fail err: ", err)
		}
	}
	for _, item := range rbJcfs {
		err := StatisticsService.AddOrUpdateRBJcfsStat(item)
		if err != nil {
			beego.Error("RBJcfsStats fail err: ", err)
		}
	}

}

// VB-首充分析
func FirstChargeAnalysis(timestamp int64) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")

	m := bson.M{"status": 1}
	var channellist []entity.ChannelInfo
	err := Channels.
		Find(m).
		All(&channellist)
	if err != nil {
		beego.Error("FirstChargeAnalysis fail err: ", err)
	}
	s := fmt.Sprintf("%s 00:00:00", today)
	startTime := utils.Str2TimeZone(s, locationName)
	e := fmt.Sprintf("%s 23:59:59", today)
	endTime := utils.Str2TimeZone(e, locationName)
	addlist := make([]entity.FirstChargeAnalysis, 0)
	if len(channellist) > 0 {
		date1, _ := utils.Unix(fmt.Sprintf("%s", today))
		for _, item := range channellist {
			info := new(entity.FirstChargeAnalysis)
			id := item.Name // 渠道名称
			name1 := item.Name1

			// 获取注册人数
			var uids []string
			pipeline1 := FindByDate(today, today, "ctime", "ctime")
			pipeline1["robot"] = false
			pipeline1["simulation_robot"] = false
			pipeline1["ad__bundle_id"] = id
			PlayerUsers.Find(pipeline1).Distinct("_id", &uids)

			// 充值userid
			var payIds []string
			pay_m := FindByDate(today, today, "ctime", "ctime")
			pay_m["order_status"] = 4
			pay_m["package_id"] = id
			Pays.Find(pay_m).Distinct("userid", &payIds)

			// 获取今日之前存在充值的人
			var beforeIds []string
			pay_m1 := bson.M{}
			pay_m1["ctime"] = bson.M{"$lt": startTime}
			pay_m1["userid"] = bson.M{"$in": payIds}
			pay_m1["order_status"] = 4
			pay_m1["package_id"] = id
			Pays.Find(pay_m1).Distinct("userid", &beforeIds)

			// 去掉重复userid
			map_ids := make(map[string]bool, 0)
			for _, cid := range beforeIds {
				map_ids[cid] = true
			}
			var qcids []string
			for _, ids := range payIds {
				if _, ok := map_ids[ids]; !ok {
					qcids = append(qcids, ids)
				}
			}

			// 根据userid查询第一次充值时间
			pipeline := []bson.M{
				{
					"$match": bson.M{
						"order_status": 4,
						"userid":       bson.M{"$in": qcids},
					},
				},
				{
					"$sort": bson.M{
						"ctime": 1, // 根据订单时间升序排序
					},
				},
				{
					"$group": bson.M{
						"_id":      "$userid",
						"orderID":  bson.M{"$first": "$_id"},
						"pay_time": bson.M{"$first": "$pay_time"},
						"amount":   bson.M{"$first": "$amount"},
					},
				},
			}
			result := []bson.M{}
			Pays.Pipe(pipeline).All(&result)
			// 获取用户玩的时间最久的游戏
			pipeline2 := []bson.M{
				{
					"$match": bson.M{
						"ctime":  bson.M{"$lte": endTime.Unix()},
						"userid": bson.M{"$in": qcids},
					},
				},
				{
					"$group": bson.M{
						"_id": bson.M{
							"userid": "$userid",
							"gtype":  "$gtype",
						},
						"time": bson.M{"$sum": "$time"},
					},
				},
				{
					"$sort": bson.M{"time": -1}, // 按时间倒序排序，确保每个用户时间最长的游戏类型在前面
				},
				{
					"$group": bson.M{
						"_id": "$_id.userid", // 重新按用户分组
						"topGame": bson.M{
							"$first": "$_id.gtype", // 取每个用户时间最长的游戏类型
						},
						"time": bson.M{
							"$first": "$time", // 取该游戏类型对应的时间
						},
					},
				},
				{
					"$project": bson.M{
						"_id":   1,          // 保留用户id
						"gtype": "$topGame", // 最长时间的游戏类型
						"time":  1,          // 对应时间
					},
				},
			}
			var time_res []bson.M
			LogGameTimes.Pipe(pipeline2).All(&time_res)
			qtCount := 0
			tp_map := make(map[int64]*entity.GameCharge, 0)
			rm_map := make(map[int64]*entity.GameCharge, 0)
			lhd_map := make(map[int64]*entity.GameCharge, 0)
			up_map := make(map[int64]*entity.GameCharge, 0)
			ak47_map := make(map[int64]*entity.GameCharge, 0)
			joker_map := make(map[int64]*entity.GameCharge, 0)
			crash_map := make(map[int64]*entity.GameCharge, 0)
			cp_map := make(map[int64]*entity.GameCharge, 0)
			ab_map := make(map[int64]*entity.GameCharge, 0)
			fj_map := make(map[int64]*entity.GameCharge, 0)
			rm_two_map := make(map[int64]*entity.GameCharge, 0)
			tplist := make([]entity.GameCharge, 0)
			lhdlist := make([]entity.GameCharge, 0)
			uplist := make([]entity.GameCharge, 0)
			rmlist := make([]entity.GameCharge, 0)
			ak47list := make([]entity.GameCharge, 0)
			jokerlist := make([]entity.GameCharge, 0)
			ablist := make([]entity.GameCharge, 0)
			crashlist := make([]entity.GameCharge, 0)
			cplist := make([]entity.GameCharge, 0)
			fjlist := make([]entity.GameCharge, 0)
			rm_two_list := make([]entity.GameCharge, 0)

			for _, res := range result {
				gtype := 0
				// number := 0
				uid := res["_id"].(string)
				money := res["amount"].(int)
				amount := int64(money)
				for _, t := range time_res {
					tid := t["_id"].(string)
					if uid == tid {
						gtype = t["gtype"].(int)
					}
				}
				switch gtype {
				case 1:
					stat, ok := tp_map[amount]
					if !ok {
						stat = &entity.GameCharge{
							GType:  1,
							Amount: amount,
						}
					}
					stat.Number++
					tp_map[amount] = stat
				case 2:
					stat, ok := lhd_map[amount]
					if !ok {
						stat = &entity.GameCharge{
							GType:  2,
							Amount: amount,
						}
					}
					stat.Number++
					lhd_map[amount] = stat
				case 3:
					stat, ok := up_map[amount]
					if !ok {
						stat = &entity.GameCharge{
							GType:  3,
							Amount: amount,
						}
					}
					stat.Number++
					up_map[amount] = stat
				case 4:
					stat, ok := rm_map[amount]
					if !ok {
						stat = &entity.GameCharge{
							GType:  4,
							Amount: amount,
						}
					}
					stat.Number++
					rm_map[amount] = stat
				case 5:
					stat, ok := ak47_map[amount]
					if !ok {
						stat = &entity.GameCharge{
							GType:  5,
							Amount: amount,
						}
					}
					stat.Number++
					ak47_map[amount] = stat
				case 6:
					stat, ok := joker_map[amount]
					if !ok {
						stat = &entity.GameCharge{
							GType:  6,
							Amount: amount,
						}
					}
					stat.Number++
					joker_map[amount] = stat
				case 7:
					stat, ok := crash_map[amount]
					if !ok {
						stat = &entity.GameCharge{
							GType:  7,
							Amount: amount,
						}
					}
					stat.Number++
					crash_map[amount] = stat
				case 8:
					stat, ok := ab_map[amount]
					if !ok {
						stat = &entity.GameCharge{
							GType:  8,
							Amount: amount,
						}
					}
					stat.Number++
					ab_map[amount] = stat
				case 9:
					stat, ok := cp_map[amount]
					if !ok {
						stat = &entity.GameCharge{
							GType:  9,
							Amount: amount,
						}
					}
					stat.Number++
					cp_map[amount] = stat
				case 10:
					stat, ok := fj_map[amount]
					if !ok {
						stat = &entity.GameCharge{
							GType:  10,
							Amount: amount,
						}
					}
					stat.Number++
					fj_map[amount] = stat
				case 12:
					stat, ok := rm_two_map[amount]
					if !ok {
						stat = &entity.GameCharge{
							GType:  10,
							Amount: amount,
						}
					}
					stat.Number++
					rm_two_map[amount] = stat
				case 0:
					qtCount++
				}
			}
			for _, v := range tp_map {
				tplist = append(tplist, *v)
			}
			for _, v := range lhd_map {
				lhdlist = append(lhdlist, *v)
			}
			for _, v := range up_map {
				uplist = append(uplist, *v)
			}
			for _, v := range rm_map {
				rmlist = append(rmlist, *v)
			}
			for _, v := range ak47_map {
				ak47list = append(ak47list, *v)
			}
			for _, v := range joker_map {
				jokerlist = append(jokerlist, *v)
			}
			for _, v := range crash_map {
				crashlist = append(crashlist, *v)
			}
			for _, v := range ab_map {
				ablist = append(ablist, *v)
			}
			for _, v := range cp_map {
				cplist = append(cplist, *v)
			}
			for _, v := range fj_map {
				fjlist = append(fjlist, *v)
			}
			for _, v := range rm_two_map {
				rm_two_list = append(rm_two_list, *v)
			}
			info.Date = date1
			info.Channel = id
			info.Channel1 = name1
			info.RegNumber = int64(len(uids))
			info.FirstNumber = int64(len(qcids))
			info.TPCharge = tplist
			info.LHDCharge = lhdlist
			info.UPCharge = uplist
			info.RMCharge = rmlist
			info.AK47Charge = ak47list
			info.JokerCharge = jokerlist
			info.CrashCharge = crashlist
			info.ABCharge = ablist
			info.CPCharge = cplist
			info.FJCharge = fjlist
			info.RMTwoCharge = rm_two_list
			info.QTNumber = int64(qtCount)
			addlist = append(addlist, *info)
		}
	}
	for _, item := range addlist {
		AddErr := StatisticsService.AddOrUpdateFirstCharge(&item)
		if AddErr != nil {
			beego.Error("FirstChargeAnalysis fail err: ", AddErr)
		}
	}
}

func ConvertToInt64(b interface{}) int64 {
	num := int64(0)
	// if b
	num, dsAmountIsInt64 := b.(int64)
	if !dsAmountIsInt64 {
		// 进行类型转换
		dsAmountInt, dsAmountIsInt := b.(int)
		if !dsAmountIsInt {
			// 处理无法转换为int64的情况
			fmt.Println("ConvertToInt64 is not int64 or int")
		}
		num = int64(dsAmountInt)
	}
	return num
}

func ConvertToUserType(money, state int) int {
	userType := 0 // 账号类型
	if money == 0 {
		if state == 1 {
			userType = 0
		} else {
			userType = 1
		}
	} else if money >= 20000 && money <= 99999 {
		userType = 2
	} else if money >= 100000 && money <= 499999 {
		userType = 3
	} else if money >= 500000 && money <= 999999 {
		userType = 4
	} else if money >= 1000000 && money <= 9999999 {
		userType = 5
	} else if money >= 10000000 {
		userType = 6
	}
	return userType
}
