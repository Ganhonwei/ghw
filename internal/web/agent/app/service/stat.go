package service

import (
	"fmt"
	"goserver/internal/web/agent/app/entity"
	"goserver/pkg/utils"
	"math"
	"strings"

	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"
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
	startTime := utils.Str2Time(s, location)
	e := fmt.Sprintf("%s 23:59:59", today)
	endTime := utils.Str2Time(e, location)
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
Title：数据总汇统计
Effect：统计用户分享
*/
func DataStatistics(timestamp int64, accountId string) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	Twotoday := tim.AddDate(0, 0, -1).Format("2006-01-02")
	s := fmt.Sprintf("%s 00:00:00", today)
	startTime := utils.Str2Time(s, location)
	e := fmt.Sprintf("%s 23:59:59", today)
	endTime := utils.Str2Time(e, location)
	var res []bson.M
	if accountId != "" {
		m := BloggerAccounts.Pipe([]bson.M{
			{
				"$match": bson.M{
					"_id": accountId,
				},
			},
			{
				"$group": bson.M{
					"_id": "$channel",
					"ids": bson.M{"$push": "$_id"},
				},
			},
		})

		err := m.All(&res)
		if err != nil {
			beego.Error("DataStatistics fail err: ", err)
		}
	} else {
		m := BloggerAccounts.Pipe([]bson.M{
			{
				"$group": bson.M{
					"_id": "$channel",
					"ids": bson.M{"$push": "$_id"},
				},
			},
		})
		err := m.All(&res)
		if err != nil {
			beego.Error("DataStatistics fail err: ", err)
		}
	}

	date1, _ := utils.Unix(fmt.Sprintf("%s", today))
	list := make([]entity.DataStatistics, 0)
	if len(res) > 0 {
		for _, item := range res {
			name := item["_id"].(string)
			ids := item["ids"].([]interface{})
			// 根据渠道名称查询渠道别名
			m := bson.M{"name": name}
			var channelinfo entity.ChannelInfo
			Channels.Find(m).One(&channelinfo)
			name1 := channelinfo.Name1

			for _, v := range ids {
				info := new(entity.DataStatistics)
				// 单个博主
				uid := v.(string)
				m1 := bson.M{
					"robot":            false,
					"simulation_robot": false,
					"ctime":            bson.M{"$gte": startTime, "$lt": endTime},
					"share_superior":   uid,
				}
				cusers := make([]string, 0)
				PlayerUsers.Find(m1).Distinct("_id", &cusers)
				rcount := int64(len(cusers))

				// 查询登录日志登录用户
				logids := make([]string, 0)
				LoginLogs.Find(bson.M{
					"login_time": bson.M{"$gte": startTime, "$lt": endTime},
				}).Distinct("userid", &logids)

				// 查询这个博主账号邀请得玩家且今日登录
				m2 := bson.M{
					"robot":            false,
					"simulation_robot": false,
					"_id":              bson.M{"$in": logids},
					"share_superior":   uid,
				}
				result2 := make([]string, 0)
				PlayerUsers.Find(m2).Distinct("_id", &result2)
				count1 := int64(len(result2))

				// 新注册用户设备数
				m3 := bson.M{
					"robot":            false,
					"simulation_robot": false,
					"ctime":            bson.M{"$gte": startTime, "$lt": endTime},
					"share_superior":   uid,
				}
				equipments := make([]string, 0)
				PlayerUsers.Find(m3).Distinct("ad__adid", &equipments)
				ecount := int64(len(equipments))

				// 昨日注册用户数
				m4 := FindByDate(Twotoday, Twotoday, "ctime", "ctime")
				m4["robot"] = false
				m4["simulation_robot"] = false
				m4["share_superior"] = uid
				count2, _ := PlayerService.GetTotal(m4)

				m5 := FindByDateBy2(Twotoday, Twotoday, today, today, "ctime", "login_time", "ctime", "login_time")
				m5["robot"] = false
				m5["simulation_robot"] = false
				m5["share_superior"] = uid
				count3, _ := PlayerService.GetTotal(m5)

				// 获取今日登录昨日充值玩家
				m11 := bson.M{}
				m11["robot"] = false
				m11["simulation_robot"] = false
				m11["share_superior"] = uid
				m11["_id"] = bson.M{"$in": logids}

				zrpayCount := int64(0)
				zrCount := int64(0)
				var yids []string
				PlayerUsers.Find(m11).Distinct("_id", &yids)
				m8 := FindByDate(Twotoday, Twotoday, "ctime", "ctime")
				m8["userid"] = bson.M{"$in": yids}
				m8["order_status"] = 4
				// m8["package_id"] = name
				var cids []string
				Pays.Find(m8).Distinct("userid", &cids)
				zrpayCount = int64(len(cids))

				//昨日充值的玩家人数
				m12 := bson.M{}
				m12["robot"] = false
				m12["simulation_robot"] = false
				m12["share_superior"] = uid
				var tids []string
				PlayerUsers.Find(m11).Distinct("_id", &tids)
				m9 := FindByDate(Twotoday, Twotoday, "ctime", "ctime")
				m9["userid"] = bson.M{"$in": tids}
				m9["order_status"] = 4
				// m9["package_id"] = name
				var zrids []string
				Pays.Find(m9).Distinct("userid", &zrids)
				zrCount = int64(len(zrids))

				// 新用户充值
				m6 := FindByDate(today, today, "ctime", "ctime")
				m6["userid"] = bson.M{"$in": cusers}
				m6["order_status"] = 4
				// m6["package_id"] = name
				list1, _ := PayService.GetByPayUser(m6)
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
				m7 := FindByDate(today, today, "ctime", "ctime")
				m7["order_status"] = 4
				m7["userid"] = bson.M{"$in": tids}
				// m7["package_id"] = name
				list2, _ := PayService.GetByPayUser(m7)
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
				m10 := FindByDate(today, today, "ctime", "ctime")
				m10["userid"] = bson.M{"$in": tids}
				// m10["package_id"] = name
				plist, _ := PayService.GetByPayUser(m10)
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
				tlist, _ := PayService.GetByWithdrawUser(m10)
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
				info.Channel = name
				info.Channel1 = name1
				info.BloggerId = uid
				info.NewRegister = rcount
				info.NewEquipment = ecount
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
				list = append(list, *info)
			}
		}
	}
	for _, item := range list {
		err := StatisticsService.AddDataStatistics(&item)
		if err != nil {
			beego.Error("DataStatistics fail err: ", err)
		}
	}

}

/*
Author：CC
Title：分享数据统计
Effect：统计用户分享
*/
func ShareStatistics(timestamp int64, accountId string) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")

	s := fmt.Sprintf("%s 00:00:00", today)
	startTime := utils.Str2Time(s, location)
	e := fmt.Sprintf("%s 23:59:59", today)
	endTime := utils.Str2Time(e, location)
	var res []bson.M
	var err error
	if accountId != "" {
		m := BloggerAccounts.Pipe([]bson.M{
			{
				"$match": bson.M{
					"_id": accountId,
				},
			},
			{
				"$group": bson.M{
					"_id": "$channel",
					"ids": bson.M{"$push": "$_id"},
				},
			},
		})
		err = m.All(&res)
		if err != nil {
			beego.Error("ShareStatistics fail err: ", err)
		}
	} else {
		m := BloggerAccounts.Pipe([]bson.M{
			{
				"$group": bson.M{
					"_id": "$channel",
					"ids": bson.M{"$push": "$_id"},
				},
			},
		})
		err = m.All(&res)
		if err != nil {
			beego.Error("ShareStatistics fail err: ", err)
		}
	}
	date1, _ := utils.Unix(fmt.Sprintf("%s", today))
	list := make([]entity.ShareDataStatistics, 0)
	if len(res) > 0 {
		for _, item := range res {
			name := item["_id"].(string)
			ids := item["ids"].([]interface{})
			users := make([]string, 0)
			for _, v := range ids {
				users = append(users, v.(string))
			}
			// 获取邀请注册人数
			pipeline := []bson.M{
				{
					"$match": bson.M{
						"robot":            false,
						"simulation_robot": false,
						"ctime":            bson.M{"$gte": startTime, "$lt": endTime},
						// "ad__bundle_id":    name,
						"share_superior": bson.M{"$in": users},
					},
				},
				{
					"$group": bson.M{
						"_id":    "$share_superior",
						"userid": bson.M{"$push": "$_id"},
						"num": bson.M{
							"$sum": 1,
						},
					},
				},
			}
			// operations := []bson.M{m, n}
			result := []bson.M{}
			pipe := PlayerUsers.Pipe(pipeline)
			err = pipe.All(&result)

			// regNumber := len(result)
			uids := make([]string, 0)

			// 根据渠道名称查询渠道别名
			m := bson.M{"name": name}
			var channelinfo entity.ChannelInfo
			Channels.Find(m).One(&channelinfo)
			name1 := channelinfo.Name1
			for _, v := range users {
				bid := v
				num := 0
				payCount := 0
				TotalAmount := 0
				for _, item := range result {
					bid = item["_id"].(string)
					if v == bid {
						num = item["num"].(int)
						userarr := item["userid"].([]interface{})
						for _, v := range userarr {
							uids = append(uids, v.(string))
						}
						// 获取被邀请的用户充值
						pipeline1 := []bson.M{
							{
								"$match": bson.M{
									"ctime":        bson.M{"$gte": startTime, "$lt": endTime},
									"userid":       bson.M{"$in": uids},
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
						// operations := []bson.M{m, n}
						result1 := []bson.M{}
						pipe1 := Pays.Pipe(pipeline1)
						err = pipe1.All(&result1)
						for _, item := range result1 {
							TotalAmount += item["Amount"].(int)
						}
						payCount = len(result1)
					}
				}
				info := new(entity.ShareDataStatistics)
				info.Date = date1
				info.Channel = name
				info.Channel1 = name1
				info.BloggerId = bid
				info.RegisterNumber = int64(num)
				info.PayNumber = int64(payCount)
				info.PayMoney = int64(TotalAmount)
				list = append(list, *info)
			}
		}
	}
	for _, item := range list {
		err := StatisticsService.AddOrUpdateShare(&item)
		if err != nil {
			beego.Error("ShareStatistics fail err: ", err)
		}
	}
}

/*
Author：CC
Title：用户留存
Effect：统计用户留存
*/
func UserRetained(timestamp int64, accountId string) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	today1 := tim.AddDate(0, 0, -1).Format("2006-01-02")
	today2 := tim.AddDate(0, 0, -2).Format("2006-01-02")
	today3 := tim.AddDate(0, 0, -3).Format("2006-01-02")
	today4 := tim.AddDate(0, 0, -4).Format("2006-01-02")
	today5 := tim.AddDate(0, 0, -5).Format("2006-01-02")
	today6 := tim.AddDate(0, 0, -6).Format("2006-01-02")
	today7 := tim.AddDate(0, 0, -7).Format("2006-01-02")
	today15 := tim.AddDate(0, 0, -15).Format("2006-01-02")
	today30 := tim.AddDate(0, 0, -30).Format("2006-01-02")
	today60 := tim.AddDate(0, 0, -60).Format("2006-01-02")

	var res []bson.M
	if accountId != "" {
		m := BloggerAccounts.Pipe([]bson.M{
			{
				"$match": bson.M{
					"_id": accountId,
				},
			},
			{
				"$group": bson.M{
					"_id": "$channel",
					"ids": bson.M{"$push": "$_id"},
				},
			},
		})
		err := m.All(&res)
		if err != nil {
			beego.Error("DataStatistics fail err: ", err)
		}
	} else {
		m := BloggerAccounts.Pipe([]bson.M{
			{
				"$group": bson.M{
					"_id": "$channel",
					"ids": bson.M{"$push": "$_id"},
				},
			},
		})
		err := m.All(&res)
		if err != nil {
			beego.Error("DataStatistics fail err: ", err)
		}
	}
	date1, _ := utils.Unix(fmt.Sprintf("%s", today))
	list := make([]entity.UserRetained, 0)
	if len(res) > 0 {
		for _, item := range res {
			name := item["_id"].(string)
			ids := item["ids"].([]interface{})
			// 根据渠道名称查询渠道别名
			m := bson.M{"name": name}
			var channelinfo entity.ChannelInfo
			Channels.Find(m).One(&channelinfo)
			name1 := channelinfo.Name1
			for _, v := range ids {
				info := new(entity.UserRetained)
				uid := v.(string)
				m := FindByDate(today, today, "ctime", "ctime")
				m["robot"] = false
				m["simulation_robot"] = false
				m["share_superior"] = uid
				count, _ := PlayerService.GetTotal(m)
				// 1日留存
				regcount1, count1 := RetainedCompute(today, today1, uid)
				// 2日留存
				regcount2, count2 := RetainedCompute(today, today2, uid)
				// 3日留存
				regcount3, count3 := RetainedCompute(today, today3, uid)
				// 4日留存
				regcount4, count4 := RetainedCompute(today, today4, uid)
				// 5日留存
				regcount5, count5 := RetainedCompute(today, today5, uid)
				// 6日留存
				regcount6, count6 := RetainedCompute(today, today6, uid)
				// 7日留存
				regcount7, count7 := RetainedCompute(today, today7, uid)
				// 15日留存
				regcount15, count15 := RetainedCompute(today, today15, uid)
				// 30日留存
				regcount30, count30 := RetainedCompute(today, today30, uid)
				// 60日留存
				regcount60, count60 := RetainedCompute(today, today60, uid)

				// 赋值
				info.Date = date1
				info.Channel = name
				info.Channel1 = name1
				info.BloggerId = uid
				info.NewNumber = count
				info.Login1 = count1
				info.Register1 = regcount1
				info.Login2 = count2
				info.Register2 = regcount2
				info.Login3 = count3
				info.Register3 = regcount3
				info.Login4 = count4
				info.Register4 = regcount4
				info.Login5 = count5
				info.Register5 = regcount5
				info.Login6 = count6
				info.Register6 = regcount6
				info.Login7 = count7
				info.Register7 = regcount7
				info.Login15 = count15
				info.Register15 = regcount15
				info.Login30 = count30
				info.Register30 = regcount30
				info.Login60 = count60
				info.Register60 = regcount60
				// info.Day1 = ComputeFloat(count1, regcount1) * 100
				// info.Day2 = ComputeFloat(count2, regcount2) * 100
				// info.Day3 = ComputeFloat(count3, regcount3) * 100
				// info.Day4 = ComputeFloat(count4, regcount4) * 100
				// info.Day5 = ComputeFloat(count5, regcount5) * 100
				// info.Day6 = ComputeFloat(count6, regcount6) * 100
				// info.Day7 = ComputeFloat(count7, regcount7) * 100
				// info.Day15 = ComputeFloat(count15, regcount15) * 100
				// info.Day30 = ComputeFloat(count30, regcount30) * 100
				// info.Day60 = ComputeFloat(count60, regcount60) * 100
				list = append(list, *info)
			}
		}
	}
	for _, item := range list {
		AddErr := StatisticsService.AddUserRetained(&item)
		if AddErr != nil {
			beego.Error("UserRetained fail err: ", AddErr)
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
			"$gte": utils.Str2Time(fmt.Sprintf("%s 00:00:00", today1), location),
			"$lt":  utils.Str2Time(fmt.Sprintf("%s 23:59:59", today1), location)},
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
			"$gte": utils.Str2Time(fmt.Sprintf("%s 00:00:00", today), location),
			"$lt":  utils.Str2Time(fmt.Sprintf("%s 23:59:59", today), location)},
		"userid": bson.M{"$in": shareIds},
	}).Distinct("userid", &logids)
	count1 = int64(len(logids))
	return regcount1, count1
}

/*
Author：CC
Title：用户付费留存
Effect：统计用户付费留存
*/
func PayUserRetained(timestamp int64, accountId string) {
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

	var res []bson.M
	if accountId != "" {
		m := BloggerAccounts.Pipe([]bson.M{
			{
				"$match": bson.M{
					"_id": accountId,
				},
			},
			{
				"$group": bson.M{
					"_id": "$channel",
					"ids": bson.M{"$push": "$_id"},
				},
			},
		})
		err := m.All(&res)
		if err != nil {
			beego.Error("DataStatistics fail err: ", err)
		}
	} else {
		m := BloggerAccounts.Pipe([]bson.M{
			{
				"$group": bson.M{
					"_id": "$channel",
					"ids": bson.M{"$push": "$_id"},
				},
			},
		})
		err := m.All(&res)
		if err != nil {
			beego.Error("DataStatistics fail err: ", err)
		}
	}
	date1, _ := utils.Unix(fmt.Sprintf("%s", today))
	list := make([]entity.PayUserRetained, 0)
	if len(res) > 0 {
		for _, item := range res {
			name := item["_id"].(string)
			ids := item["ids"].([]interface{})
			// 根据渠道名称查询渠道别名
			m := bson.M{"name": name}
			var channelinfo entity.ChannelInfo
			Channels.Find(m).One(&channelinfo)
			name1 := channelinfo.Name1
			for _, v := range ids {
				info := new(entity.PayUserRetained)
				uid := v.(string)
				// 查询该博主下得所有用户
				m := bson.M{
					"robot":            false,
					"simulation_robot": false,
					"share_superior":   uid,
				}
				shareIds := make([]string, 0)
				PlayerUsers.Find(m).Distinct("_id", &shareIds)
				var payids []string // 注册且支付ID
				m1 := FindByDate(today, today, "ctime", "ctime")
				m1["userid"] = bson.M{"$in": shareIds}
				m1["order_status"] = 4 // 充值成功
				Pays.Find(m1).Distinct("userid", &payids)
				count := len(payids)
				// 1日留存
				regcount1, count1 := PayRetainedCompute(today, today1, uid, shareIds)
				// 2日留存
				regcount2, count2 := PayRetainedCompute(today, today2, uid, shareIds)
				// 3日留存
				regcount3, count3 := PayRetainedCompute(today, today3, uid, shareIds)
				// 4日留存
				regcount4, count4 := PayRetainedCompute(today, today4, uid, shareIds)
				// 5日留存
				regcount5, count5 := PayRetainedCompute(today, today5, uid, shareIds)
				// 6日留存
				regcount6, count6 := PayRetainedCompute(today, today6, uid, shareIds)
				// 7日留存
				regcount7, count7 := PayRetainedCompute(today, today7, uid, shareIds)
				// 15日留存
				regcount15, count15 := PayRetainedCompute(today, today15, uid, shareIds)
				// 30日留存
				regcount30, count30 := PayRetainedCompute(today, today30, uid, shareIds)
				// 60日留存
				regcount60, count60 := PayRetainedCompute(today, today60, uid, shareIds)

				// 赋值
				info.Date = date1
				info.Channel = name
				info.Channel1 = name1
				info.BloggerId = uid
				info.NewNumber = int64(count)
				info.Login1 = count1
				info.Register1 = regcount1
				info.Login2 = count2
				info.Register2 = regcount2
				info.Login3 = count3
				info.Register3 = regcount3
				info.Login4 = count4
				info.Register4 = regcount4
				info.Login5 = count5
				info.Register5 = regcount5
				info.Login6 = count6
				info.Register6 = regcount6
				info.Login7 = count7
				info.Register7 = regcount7
				info.Login15 = count15
				info.Register15 = regcount15
				info.Login30 = count30
				info.Register30 = regcount30
				info.Login60 = count60
				info.Register60 = regcount60
				// info.Day1 = ComputeFloat(count1, regcount1) * 100
				// info.Day2 = ComputeFloat(count2, regcount2) * 100
				// info.Day3 = ComputeFloat(count3, regcount3) * 100
				// info.Day4 = ComputeFloat(count4, regcount4) * 100
				// info.Day5 = ComputeFloat(count5, regcount5) * 100
				// info.Day6 = ComputeFloat(count6, regcount6) * 100
				// info.Day7 = ComputeFloat(count7, regcount7) * 100
				// info.Day15 = ComputeFloat(count15, regcount15) * 100
				// info.Day30 = ComputeFloat(count30, regcount30) * 100
				// info.Day60 = ComputeFloat(count60, regcount60) * 100
				list = append(list, *info)
			}
		}
	}
	for _, item := range list {
		AddErr := StatisticsService.AddPayUserRetained(&item)
		if AddErr != nil {
			beego.Error("PayUserRetained fail err: ", AddErr)
		}
	}
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
			"$gte": utils.Str2Time(fmt.Sprintf("%s 00:00:00", today1), location),
			"$lt":  utils.Str2Time(fmt.Sprintf("%s 23:59:59", today1), location)},
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

// 统计埋点数据
func PointData(timestamp int64, accountId string) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	// s := fmt.Sprintf("%s 00:00:00", today)
	// startTime := utils.Str2Time(s)
	// e := fmt.Sprintf("%s 23:59:59", today)
	// endTime := utils.Str2Time(e)
	var res []bson.M
	if accountId != "" {
		m := BloggerAccounts.Pipe([]bson.M{
			{
				"$match": bson.M{
					"_id": accountId,
				},
			},
			{
				"$group": bson.M{
					"_id": "$channel",
					"ids": bson.M{"$push": "$_id"},
				},
			},
		})
		err := m.All(&res)
		if err != nil {
			beego.Error("DataStatistics fail err: ", err)
		}
	} else {
		m := BloggerAccounts.Pipe([]bson.M{
			{
				"$group": bson.M{
					"_id": "$channel",
					"ids": bson.M{"$push": "$_id"},
				},
			},
		})
		err := m.All(&res)
		if err != nil {
			beego.Error("DataStatistics fail err: ", err)
		}
	}
	list := make([]entity.PointData, 0)

	if len(res) > 0 {
		date1, _ := utils.Unix(fmt.Sprintf("%s", today))
		for _, item := range res {
			name := item["_id"].(string)
			ids := item["ids"].([]interface{})
			// 根据渠道名称查询渠道别名
			m := bson.M{"name": name}
			var channelinfo entity.ChannelInfo
			Channels.Find(m).One(&channelinfo)
			name1 := channelinfo.Name1
			for _, v := range ids {
				info := new(entity.PointData)
				// 单个博主
				uid := v.(string)
				var userids []string
				m := FindByDate(today, today, "ctime", "ctime")
				m["robot"] = false
				m["simulation_robot"] = false
				m["share_superior"] = uid
				PlayerUsers.Find(m).Distinct("_id", &userids)
				count := len(userids)
				// 游客注册
				m["tourist"] = bson.M{"$ne": ""}
				ykcount, _ := PlayerService.GetTotal(m)
				// 手机注册
				m["tourist"] = ""
				sjcount, _ := PlayerService.GetTotal(m)

				// 领取金币
				m3 := FindByDate1(today, today, "ctime", "ctime")
				m3["typ"] = 1
				m3["userid"] = bson.M{"$in": userids}
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

				info.Date = date1
				info.Channel = name
				info.Channel1 = name1
				info.BloggerId = uid
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
				list = append(list, *info)
			}
		}
	}
	for _, item := range list {
		err1 := StatisticsService.AddPointData(&item)
		if err1 != nil {
			beego.Error("PointData fail err: ", err1)
		}
	}
}

// 统计埋点数据-局数分析
func GameNumberAnalysisData(timestamp int64, accountId string) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	s := fmt.Sprintf("%s 00:00:00", today)
	startTime := utils.Str2Time(s, location)
	e := fmt.Sprintf("%s 23:59:59", today)
	endTime := utils.Str2Time(e, location)
	var res []bson.M
	if accountId != "" {
		m := BloggerAccounts.Pipe([]bson.M{
			{
				"$match": bson.M{
					"_id": accountId,
				},
			},
			{
				"$group": bson.M{
					"_id": "$channel",
					"ids": bson.M{"$push": "$_id"},
				},
			},
		})
		err := m.All(&res)
		if err != nil {
			beego.Error("DataStatistics fail err: ", err)
		}
	} else {
		m := BloggerAccounts.Pipe([]bson.M{
			{
				"$group": bson.M{
					"_id": "$channel",
					"ids": bson.M{"$push": "$_id"},
				},
			},
		})
		err := m.All(&res)
		if err != nil {
			beego.Error("DataStatistics fail err: ", err)
		}
	}
	list := make([]entity.GameNumberAnalysis, 0)
	if len(res) > 0 {
		date1, _ := utils.Unix(fmt.Sprintf("%s", today))
		for _, item := range res {
			name := item["_id"].(string)
			ids := item["ids"].([]interface{})
			// 根据渠道名称查询渠道别名
			m := bson.M{"name": name}
			var channelinfo entity.ChannelInfo
			Channels.Find(m).One(&channelinfo)
			name1 := channelinfo.Name1
			for _, v := range ids {
				info := new(entity.GameNumberAnalysis)
				// 单个博主
				uid := v.(string)
				pipeline := []bson.M{
					{
						"$match": bson.M{
							"ctime":            bson.M{"$gte": startTime, "$lt": endTime},
							"robot":            false,
							"simulation_robot": false,
							"share_superior":   uid,
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
				info.Date = date1
				info.Channel = name
				info.Channel1 = name1
				info.BloggerId = uid
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
				list = append(list, *info)
			}
		}
	}

	for _, item := range list {
		err1 := StatisticsService.AddGameNumberAnalysis(&item)
		if err1 != nil {
			beego.Error("GameNumberAnalysisData fail err: ", err1)
		}
	}
}

func GameByUserCount(startTime, endTime int64, userid string) int {
	count := 0
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
				"Count": bson.M{
					"$sum": 1,
				},
			},
		},
	}
	result1 := []bson.M{}
	pipe1 := LogGameTimes.Pipe(pipeline1)
	err1 := pipe1.All(&result1)
	if err1 != nil {
		beego.Error("GameByUserCount fail err: ", err1)
	}
	if len(result1) > 0 {
		for _, item := range result1 {
			// ids = append(ids, item["_id"].(string))
			count += item["Count"].(int)
		}
	}
	return count
}

// time analysis
func PlaytimeAnalysisData(timestamp int64, accountId string) {
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	s := fmt.Sprintf("%s 00:00:00", today)
	startTime := utils.Str2Time(s, location)
	e := fmt.Sprintf("%s 23:59:59", today)
	endTime := utils.Str2Time(e, location)
	var res []bson.M
	if accountId != "" {
		m := BloggerAccounts.Pipe([]bson.M{
			{
				"$match": bson.M{
					"_id": accountId,
				},
			},
			{
				"$group": bson.M{
					"_id": "$channel",
					"ids": bson.M{"$push": "$_id"},
				},
			},
		})
		err := m.All(&res)
		if err != nil {
			beego.Error("DataStatistics fail err: ", err)
		}
	} else {
		m := BloggerAccounts.Pipe([]bson.M{
			{
				"$group": bson.M{
					"_id": "$channel",
					"ids": bson.M{"$push": "$_id"},
				},
			},
		})
		err := m.All(&res)
		if err != nil {
			beego.Error("DataStatistics fail err: ", err)
		}
	}
	list := make([]entity.PlaytimeAnalysis, 0)
	if len(res) > 0 {
		date1, _ := utils.Unix(fmt.Sprintf("%s", today))
		for _, item := range res {
			name := item["_id"].(string)
			ids := item["ids"].([]interface{})
			// 根据渠道名称查询渠道别名
			m := bson.M{"name": name}
			var channelinfo entity.ChannelInfo
			Channels.Find(m).One(&channelinfo)
			name1 := channelinfo.Name1
			for _, v := range ids {
				info := new(entity.PlaytimeAnalysis)
				// 单个博主
				uid := v.(string)
				var userids []string
				m := FindByDate(today, today, "ctime", "ctime")
				m["robot"] = false
				m["simulation_robot"] = false
				m["share_superior"] = uid
				PlayerUsers.Find(m).Distinct("_id", &userids)
				rcount := len(userids)
				palytime1 := int64(0)
				palytime2 := int64(0)
				palytime6 := int64(0)
				palytime11 := int64(0)
				palytime21 := int64(0)
				palytime31 := int64(0)
				statrTimestamp := startTime.Unix()
				endTimestamp := endTime.Unix()
				for _, id := range userids {
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
				list = append(list, *info)
			}
		}
	}
	for _, item := range list {
		err1 := StatisticsService.AddPlaytimeAnalysis(&item)
		if err1 != nil {
			beego.Error("PlaytimeAnalysisData fail err: ", err1)
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
