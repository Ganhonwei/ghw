package controllers

import (
	"fmt"
	"goserver/internal/web/admin/app/libs"
	"goserver/internal/web/admin/app/service"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	"github.com/globalsign/mgo/bson"
)

type GameStatsController struct {
	BaseController
}

// Crash Crash统计
func (c *GameStatsController) Crash() {
	startDate := c.GetString("start_date")
	endDate := c.GetString("end_date")

	status, _ := c.GetInt("status")
	page, _ := strconv.Atoi(c.GetString("page"))
	packageId := c.GetString("package_id")
	typeId, _ := c.GetInt("typeId")
	aliasId := c.GetString("alias_id")
	aliasName := c.GetString("alias_name")
	utypeId := c.GetString("utype_id")
	tabId, _ := c.GetInt("tab_id")
	if page < 1 {
		page = 1
	}
	userid := c.GetString("userid")
	lossDays1, _ := c.GetInt("lossDays1")
	lossDays2, _ := c.GetInt("lossDays2")
	bets1, _ := c.GetInt("bets1")
	bets2, _ := c.GetInt("bets2")
	betAvg1, _ := c.GetInt("betAvg1")
	betAvg2, _ := c.GetInt("betAvg2")
	win1, _ := c.GetInt("win1")
	win2, _ := c.GetInt("win2")
	winAvg1, _ := c.GetInt("winAvg1")
	winAvg2, _ := c.GetInt("winAvg2")

	if startDate == "" || endDate == "" {
		// 默认近七天数据
		today := bson.Now()
		// tim := utils.Stamp2Time(timestamp)
		endDate = today.Format("2006-01-02")
		startDate = today.AddDate(0, 0, -6).Format("2006-01-02")
	}
	// sql 筛选条件
	params := getLHDStatsParam(userid, startDate, endDate,
		bets1, bets2, betAvg1, betAvg2, win1, win2, lossDays1, lossDays2)

	// 玩家类型 账号类型 流失天数筛选 打码量筛选 局均码筛选
	m := service.FindByDate1(startDate, endDate, "date", "date")
	m2 := bson.M{}
	// 打码量, 注均码
	if bets1 > 0 && bets2 > 0 {
		m2["bets"] = bson.M{"$gte": bets1 * 100, "$lte": bets2 * 100}
	} else {
		bets1, bets2 = 0, 0
	}
	if betAvg1 > 0 && betAvg2 > 0 {
		m2["round_bet_avg"] = bson.M{"$gte": betAvg1 * 100, "$lte": betAvg2 * 100}
	} else {
		betAvg1, betAvg2 = 0, 0
	}

	if win1 != 0 && win2 != 0 {
		if typeId == 0 || typeId == 1 { // 汇总/明细, 净赢筛选
			m2["cash"] = bson.M{"$gte": win1 * 100, "$lte": win2 * 100}
		} else {
			m2["wins"] = bson.M{"$gte": win1 * 100, "$lte": win2 * 100}
		}
	} else {
		win1, win2 = 0, 0
	}
	if winAvg1 > 0 && winAvg2 > 0 {
		m2["win_escape_multiple"] = bson.M{"$gte": winAvg1, "$lte": winAvg2}
	} else {
		winAvg1, winAvg2 = 0, 0
	}
	m3 := bson.M{}
	if lossDays1 > 0 && lossDays2 > 0 {
		m3["loss_days"] = bson.M{
			"$gte": lossDays1,
			"$lte": lossDays2,
		}
	}

	// 明细用户类型筛选多选
	if userid != "" {
		m["userid"] = userid
	}
	var utypeIds []string
	if utypeId != "" && utypeId != "0" && utypeId != "-" {
		utypeIds = strings.Split(utypeId, ",")
	}
	var utypesM []bson.M
	var utypeFilters []string
	for _, utypeId := range utypeIds {
		usertypeId, err := strconv.Atoi(utypeId)
		if err != nil {
			beego.Error("usertypeId error: ", err)
			continue
		}
		if usertypeId == 0 {
			continue
		}

		// 新玩家，小R，中R，大R，超大R
		/*
			新用户：充值0.00卢-99.99卢
			充值
			普充：100.00卢-999.99卢
			小R：1000.00-4999.99
			中R：充值5000.00卢-9999.99卢
			大R：充值10000.00卢-99999.99卢
			超大R：充值100000.00卢以上
		*/
		startMoney := 0
		endMoney := 0
		if usertypeId >= 5 && usertypeId <= 10 {
			switch usertypeId {
			case 5:
				// 零充
				utypeFilters = append(utypeFilters, "(money=0)")
				startMoney = 0
				endMoney = 0
			case 6:
				// 普充
				startMoney = 20000.00
				endMoney = 99999
				utypeFilters = append(utypeFilters, fmt.Sprintf("(money between %d and %d)", startMoney, endMoney))
			case 7:
				// 小R
				startMoney = 100000
				endMoney = 499999
				utypeFilters = append(utypeFilters, fmt.Sprintf("(money between %d and %d)", startMoney, endMoney))
			case 8:
				// 中R
				startMoney = 500000
				endMoney = 999999
				utypeFilters = append(utypeFilters, fmt.Sprintf("(money between %d and %d)", startMoney, endMoney))
			case 9:
				// 大R
				startMoney = 1000000
				endMoney = 9999999
				utypeFilters = append(utypeFilters, fmt.Sprintf("(money between %d and %d)", startMoney, endMoney))
			case 10:
				// 超大R
				startMoney = 10000000
				endMoney = -1
			}
			if endMoney == -1 {
				utypesM = append(utypesM, bson.M{
					"money":        bson.M{"$gte": startMoney},
					"state":        2,
					"custom_types": bson.M{"$eq": nil},
				})
				utypeFilters = append(utypeFilters, fmt.Sprintf("(state = 2 and money >= %d)", startMoney))
			} else {
				utypesM = append(utypesM, bson.M{
					"money":        bson.M{"$gte": startMoney, "$lte": endMoney},
					"state":        2,
					"custom_types": bson.M{"$eq": nil},
				})
				utypeFilters = append(utypeFilters, fmt.Sprintf("(state = 2 and money between %d and %d)", startMoney, endMoney))
			}
		} else {
			if usertypeId == 11 {
				utypesM = append(utypesM, bson.M{
					"custom_types": bson.M{"$ne": nil},
				})
				// m3["custom_types"] = bson.M{"$ne": ""}
				// m3["custom_types"] = bson.M{"$ne": nil}

				// m["custom_types"] = bson.M{"$ne": bson.M{}, "$exists": true}
				// m["custom_types"] = bson.M{"$ne": pb.NULL}
			} else {
				utypesM = append(utypesM, bson.M{
					"state":        usertypeId,
					"custom_types": bson.M{"$eq": nil},
				})
				utypeFilters = append(utypeFilters, fmt.Sprintf("(state = %d)", usertypeId))
			}
		}
	}
	if len(utypesM) > 0 {
		m3["$or"] = utypesM
	}
	if len(utypeFilters) > 0 {
		params["utypeQ"] = fmt.Sprintf(" and (%s)", strings.Join(utypeFilters, " or "))
	}

	if status == 0 {
		my_data := c.GetSession("my_select_pakeageid")
		if my_data != nil && packageId == "" {
			if packid, ok := my_data.(string); ok {
				packageId = packid
			}
		}
		status = 1
	}
	var packageIds []string
	if packageId != "" && packageId != "0" && packageId != "-" {
		if strings.Contains(packageId, ",") {
			palkage_arr := strings.Split(packageId, ",")
			temp_arr := make([]string, 0)
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					temp_arr = append(temp_arr, item)
				}
			}
			packageIds = append(packageIds, temp_arr...)
			// m["ad__bundle_id"] = bson.M{"$in": temp_arr}
		} else {
			packageIds = append(packageIds, packageId)
			// m["ad__bundle_id"] = packageId
		}
		// 存储缓存
		c.SetSession("my_select_pakeageid", packageId)
	} else {
		// 清除缓存
		c.DelSession("my_select_pakeageid")
	}
	// 渠道别名
	if aliasId != "" && aliasId != "0" && aliasId != "-" {
		var aliasIds []string
		if strings.Contains(aliasId, ",") {
			palkage_arr := strings.Split(aliasId, ",")
			temp_arr := make([]string, 0)
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					temp_arr = append(temp_arr, item)
				}
			}
			aliasIds = append(aliasIds, temp_arr...)
		} else {
			aliasIds = append(aliasIds, aliasId)
		}

		channels, err := service.ChannelService.GetChannelByNameAliasList(aliasIds)
		if err != nil {
			beego.Error("channel error1: ", err)
		} else {
			for _, c := range channels {
				packageIds = append(packageIds, c.Name)
			}
		}
	}
	// if aliasName != "" {
	// 	m["ad__bundle_id"] = bson.M{"$regex": aliasName, "$options": "i"}
	// }
	if len(packageIds) > 0 {
		m["ad__bundle_id"] = bson.M{"$in": packageIds}
		params["ad__bundle_id"] = packageIds
	}

	if tabId == 1 {
		m["regist_area"] = bson.M{"$ne": 1}
		params["regist_area"] = []int32{0, 3}
	} else if tabId == 2 {
		m["regist_area"] = 1
		params["regist_area"] = []int32{1}
	} else if tabId == 3 {
		m["regist_area"] = 2
		params["regist_area"] = []int32{2}
	}

	var count int
	if typeId == 1 { // 明细
		players, total, err := service.GameStatsService.GetCrashStatPlayers(page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetCrashStatPlayers error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 2 {
		// Crash 扶摇直上
		players, total, err := service.GameStatsService.GetCrashFYZSStat(service.CrashFYZSStats, page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetCrashFYZSStat error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 3 {
		// Crash欲薅无门
		players, total, err := service.GameStatsService.GetCrashYHWMStat(service.CrashYHWMStats, page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetCrashYHWMStat error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 4 {
		// 起死回生
		players, total, err := service.GameStatsService.GetCrashQSHSStat(service.CrashQSHSStats, page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetCrashQSHSStat error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 5 {
		// 奖池风控
		players, total, err := service.GameStatsService.GetCrashJCFKStat(service.CrashJCFKStats, page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetCrashJCFKStat error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 6 {
		// 冒险奖励
		players, total, err := service.GameStatsService.GetCrashMXJLStat(service.CrashMXJLStats, page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetCrashMXJLStat error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 7 {
		// 人狂有祸
		players, total, err := service.GameStatsService.GetCrashRKYSStat(service.CrashRKYSStats, page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetCrashRKYSStat error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 8 {
		// 高潮涌现
		players, total, err := service.GameStatsService.GetCrashGCYXStat(page, c.pageSize, params)
		if err != nil {
			beego.Error("GetCrashRKYSStat error: ", err)
		}
		c.Data["list"] = players
		count = total

	} else if typeId == 0 { // 汇总
		dates, err := service.GameStatsService.GetCrashStatDates(m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetCrashStatDates error: ", err)
		}
		c.Data["list"] = dates
		count = len(dates)
	}

	usertype := map[int]string{
		0:  "全部",
		1:  "新手",
		3:  "平民",
		4:  "泡沫",
		5:  "零充",
		6:  "普充",
		7:  "小R",
		8:  "中R",
		9:  "大R",
		10: "超大R",
		11: "手动",
	}
	packageList, _ := service.ChannelService.GetChannelAll()
	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()
	c.Data["pageTitle"] = "Crash统计"
	c.Data["count"] = count
	if typeId != 0 {
		c.Data["pageBar"] = libs.NewPager(page, int(count), c.pageSize, beego.URLFor("GameStatsController.Crash", "start_date", startDate, "end_date", endDate, "status", status, "package_id", packageId, "typeId", typeId, "alias_id", aliasId, "alias_name", aliasName, "usertypeId", utypeId, "tab_id", tabId, "utype_id", utypeId, "userid", userid, "lossDays1", lossDays1, "lossDays2", lossDays2, "bets1", bets1, "bets2", bets2, "betAvg1", betAvg1, "betAvg2", betAvg2, "win1", win1, "win2", win2, "winAvg1", winAvg1, "winAvg2", winAvg2), true).ToString()
	}
	c.Data["usertype"] = usertype
	c.Data["usertypeId"] = utypeId
	c.Data["startDate"] = startDate
	c.Data["endDate"] = endDate
	c.Data["packageId"] = packageId
	c.Data["packageList"] = packageList
	c.Data["typeId"] = typeId
	c.Data["aliasId"] = aliasId
	c.Data["packageAliasList"] = packageAliasList
	c.Data["aliasName"] = aliasName
	c.Data["status"] = status
	c.Data["isOperation"] = c.auth.HasAccessPerm(c.controllerName, "crash")
	c.Data["tabId"] = tabId
	c.Data["page"] = page
	c.Data["userid"] = userid
	c.Data["lossDays1"] = lossDays1
	c.Data["lossDays2"] = lossDays2
	c.Data["bets1"] = bets1
	c.Data["bets2"] = bets2
	c.Data["betAvg1"] = betAvg1
	c.Data["betAvg2"] = betAvg2
	c.Data["win1"] = win1
	c.Data["win2"] = win2
	c.Data["winAvg1"] = winAvg1
	c.Data["winAvg2"] = winAvg2
	c.display()
}

// Plane Plane统计
func (c *GameStatsController) Plane() {
	startDate := c.GetString("start_date")
	endDate := c.GetString("end_date")

	status, _ := c.GetInt("status")
	page, _ := strconv.Atoi(c.GetString("page"))
	packageId := c.GetString("package_id")
	typeId, _ := c.GetInt("typeId")
	aliasId := c.GetString("alias_id")
	aliasName := c.GetString("alias_name")
	utypeId := c.GetString("utype_id")
	tabId, _ := c.GetInt("tab_id")
	if page < 1 {
		page = 1
	}
	userid := c.GetString("userid")
	lossDays1, _ := c.GetInt("lossDays1")
	lossDays2, _ := c.GetInt("lossDays2")
	bets1, _ := c.GetInt("bets1")
	bets2, _ := c.GetInt("bets2")
	betAvg1, _ := c.GetInt("betAvg1")
	betAvg2, _ := c.GetInt("betAvg2")
	win1, _ := c.GetInt("win1")
	win2, _ := c.GetInt("win2")
	winAvg1, _ := c.GetInt("winAvg1")
	winAvg2, _ := c.GetInt("winAvg2")

	if startDate == "" || endDate == "" {
		// 默认近七天数据
		today := bson.Now()
		// tim := utils.Stamp2Time(timestamp)
		endDate = today.Format("2006-01-02")
		startDate = today.AddDate(0, 0, -6).Format("2006-01-02")
	}
	// sql 筛选条件
	params := getLHDStatsParam(userid, startDate, endDate,
		bets1, bets2, betAvg1, betAvg2, win1, win2, lossDays1, lossDays2)

	// 玩家类型 账号类型 流失天数筛选 打码量筛选 局均码筛选
	m := service.FindByDate1(startDate, endDate, "date", "date")
	m2 := bson.M{}
	// 打码量, 注均码
	if bets1 > 0 && bets2 > 0 {
		m2["bets"] = bson.M{"$gte": bets1 * 100, "$lte": bets2 * 100}
	} else {
		bets1, bets2 = 0, 0
	}
	if betAvg1 > 0 && betAvg2 > 0 {
		m2["round_bet_avg"] = bson.M{"$gte": betAvg1 * 100, "$lte": betAvg2 * 100}
	} else {
		betAvg1, betAvg2 = 0, 0
	}
	m3 := bson.M{}
	if lossDays1 > 0 && lossDays2 > 0 {
		m3["loss_days"] = bson.M{
			"$gte": lossDays1,
			"$lte": lossDays2,
		}
	}

	// 明细用户类型筛选多选
	if userid != "" {
		m["userid"] = userid
	}
	var utypeIds []string
	if utypeId != "" && utypeId != "0" && utypeId != "-" {
		utypeIds = strings.Split(utypeId, ",")
	}
	var utypesM []bson.M
	var utypeFilters []string
	for _, utypeId := range utypeIds {
		usertypeId, err := strconv.Atoi(utypeId)
		if err != nil {
			beego.Error("usertypeId error: ", err)
			continue
		}
		if usertypeId == 0 {
			continue
		}

		// 新玩家，小R，中R，大R，超大R
		/*
			新用户：充值0.00卢-99.99卢
			充值
			普充：100.00卢-999.99卢
			小R：1000.00-4999.99
			中R：充值5000.00卢-9999.99卢
			大R：充值10000.00卢-99999.99卢
			超大R：充值100000.00卢以上
		*/
		startMoney := 0
		endMoney := 0
		if usertypeId >= 5 && usertypeId <= 10 {
			switch usertypeId {
			case 5:
				// 零充
				utypeFilters = append(utypeFilters, "(money=0)")
				startMoney = 0
				endMoney = 0
			case 6:
				// 普充
				startMoney = 20000.00
				endMoney = 99999
				utypeFilters = append(utypeFilters, fmt.Sprintf("(money between %d and %d)", startMoney, endMoney))
			case 7:
				// 小R
				startMoney = 100000
				endMoney = 499999
				utypeFilters = append(utypeFilters, fmt.Sprintf("(money between %d and %d)", startMoney, endMoney))
			case 8:
				// 中R
				startMoney = 500000
				endMoney = 999999
				utypeFilters = append(utypeFilters, fmt.Sprintf("(money between %d and %d)", startMoney, endMoney))
			case 9:
				// 大R
				startMoney = 1000000
				endMoney = 9999999
				utypeFilters = append(utypeFilters, fmt.Sprintf("(money between %d and %d)", startMoney, endMoney))
			case 10:
				// 超大R
				startMoney = 10000000
				endMoney = -1
			}
			if endMoney == -1 {
				utypesM = append(utypesM, bson.M{
					"money":        bson.M{"$gte": startMoney},
					"state":        2,
					"custom_types": bson.M{"$eq": nil},
				})
				utypeFilters = append(utypeFilters, fmt.Sprintf("(state = 2 and money >= %d)", startMoney))
			} else {
				utypesM = append(utypesM, bson.M{
					"money":        bson.M{"$gte": startMoney, "$lte": endMoney},
					"state":        2,
					"custom_types": bson.M{"$eq": nil},
				})
				utypeFilters = append(utypeFilters, fmt.Sprintf("(state = 2 and money between %d and %d)", startMoney, endMoney))
			}
		} else {
			if usertypeId == 11 {
				utypesM = append(utypesM, bson.M{
					"custom_types": bson.M{"$ne": nil},
				})
				// m3["custom_types"] = bson.M{"$ne": ""}
				// m3["custom_types"] = bson.M{"$ne": nil}

				// m["custom_types"] = bson.M{"$ne": bson.M{}, "$exists": true}
				// m["custom_types"] = bson.M{"$ne": pb.NULL}
			} else {
				utypesM = append(utypesM, bson.M{
					"state":        usertypeId,
					"custom_types": bson.M{"$eq": nil},
				})
				utypeFilters = append(utypeFilters, fmt.Sprintf("(state = %d)", usertypeId))
			}
		}
	}
	if len(utypesM) > 0 {
		m3["$or"] = utypesM
	}
	if len(utypeFilters) > 0 {
		params["utypeQ"] = fmt.Sprintf(" and (%s)", strings.Join(utypeFilters, " or "))
	}

	if status == 0 {
		my_data := c.GetSession("my_select_pakeageid")
		if my_data != nil && packageId == "" {
			if packid, ok := my_data.(string); ok {
				packageId = packid
			}
		}
		status = 1
	}

	var packageIds []string
	if packageId != "" && packageId != "0" && packageId != "-" {
		if strings.Contains(packageId, ",") {
			palkage_arr := strings.Split(packageId, ",")
			temp_arr := make([]string, 0)
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					temp_arr = append(temp_arr, item)
				}
			}
			packageIds = append(packageIds, temp_arr...)
			// m["ad__bundle_id"] = bson.M{"$in": temp_arr}
		} else {
			packageIds = append(packageIds, packageId)
			// m["ad__bundle_id"] = packageId
		}
		// 存储缓存
		c.SetSession("my_select_pakeageid", packageId)
	} else {
		// 清除缓存
		c.DelSession("my_select_pakeageid")
	}
	// 渠道别名
	if aliasId != "" && aliasId != "0" && aliasId != "-" {
		var aliasIds []string
		if strings.Contains(aliasId, ",") {
			palkage_arr := strings.Split(aliasId, ",")
			temp_arr := make([]string, 0)
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					temp_arr = append(temp_arr, item)
				}
			}
			aliasIds = append(aliasIds, temp_arr...)
		} else {
			aliasIds = append(aliasIds, aliasId)
		}

		channels, err := service.ChannelService.GetChannelByNameAliasList(aliasIds)
		if err != nil {
			beego.Error("channel error1: ", err)
		} else {
			for _, c := range channels {
				packageIds = append(packageIds, c.Name)
			}
		}
	}
	// if aliasName != "" {
	// 	m["ad__bundle_id"] = bson.M{"$regex": aliasName, "$options": "i"}
	// }
	if len(packageIds) > 0 {
		m["ad__bundle_id"] = bson.M{"$in": packageIds}
		params["ad__bundle_id"] = packageIds
	}

	if tabId == 1 {
		m["regist_area"] = bson.M{"$ne": 1}
		params["regist_area"] = []int32{0, 3}
	} else if tabId == 2 {
		m["regist_area"] = 1
		params["regist_area"] = []int32{1}
	} else if tabId == 3 {
		m["regist_area"] = 2
		params["regist_area"] = []int32{2}
	}

	var count int
	if typeId == 1 { // 明细
		players, total, err := service.GameStatsService.GetPlaneStatPlayers(page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetPlaneStatPlayers error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 2 {
		// Crash 扶摇直上
		players, total, err := service.GameStatsService.GetCrashFYZSStat(service.PlaneFYZSStats, page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetCrashFYZSStat error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 3 {
		// Crash欲薅无门
		players, total, err := service.GameStatsService.GetCrashYHWMStat(service.PlaneYHWMStats, page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetCrashYHWMStat error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 4 {
		// 起死回生
		players, total, err := service.GameStatsService.GetCrashQSHSStat(service.PlaneQSHSStats, page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetCrashQSHSStat error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 5 {
		// 奖池风控
		players, total, err := service.GameStatsService.GetCrashJCFKStat(service.PlaneJCFKStats, page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetCrashJCFKStat error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 6 {
		// 冒险奖励
		players, total, err := service.GameStatsService.GetCrashMXJLStat(service.PlaneMXJLStats, page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetCrashMXJLStat error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 7 {
		// 人狂有祸
		players, total, err := service.GameStatsService.GetCrashRKYSStat(service.PlaneRKYSStats, page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetCrashRKYSStat error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 8 {
		// 高潮涌现
		players, total, err := service.GameStatsService.GetPlaneGCYXStat(page, c.pageSize, params)
		if err != nil {
			beego.Error("GetPlaneGCYXStat error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 0 { // 汇总
		dates, err := service.GameStatsService.GetPlaneStatDates(m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetPlaneStatDates error: ", err)
		}
		c.Data["list"] = dates
		count = len(dates)
	}

	usertype := map[int]string{
		0:  "全部",
		1:  "新手",
		3:  "平民",
		4:  "泡沫",
		5:  "零充",
		6:  "普充",
		7:  "小R",
		8:  "中R",
		9:  "大R",
		10: "超大R",
		11: "手动",
	}
	packageList, _ := service.ChannelService.GetChannelAll()
	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()
	c.Data["pageTitle"] = "飞机统计"
	c.Data["count"] = count
	if typeId != 0 {
		c.Data["pageBar"] = libs.NewPager(page, int(count), c.pageSize, beego.URLFor("GameStatsController.Plane", "start_date", startDate, "end_date", endDate, "status", status, "package_id", packageId, "typeId", typeId, "alias_id", aliasId, "alias_name", aliasName, "usertypeId", utypeId, "tab_id", tabId, "utype_id", utypeId, "userid", userid, "lossDays1", lossDays1, "lossDays2", lossDays2, "bets1", bets1, "bets2", bets2, "betAvg1", betAvg1, "betAvg2", betAvg2, "win1", win1, "win2", win2, "winAvg1", winAvg1, "winAvg2", winAvg2), true).ToString()
		// c.Data["pageBar"] = libs.NewPager(page, int(count), c.pageSize, beego.URLFor("GameStatsController.Plane", "startDate", startDate, "endDate", endDate, "status", status, "packageId", packageId, "typeId", typeId, "aliasId", aliasId, "aliasName", aliasName, "usertypeId", usertypeId, "tabId", tabId, "userid", userid, "lossDays1", lossDays1, "lossDays2", lossDays2, "bets1", bets1, "bets2", bets2, "betAvg1", betAvg1, "betAvg2", betAvg2), true).ToString()
	}
	c.Data["usertype"] = usertype
	c.Data["usertypeId"] = utypeId
	c.Data["startDate"] = startDate
	c.Data["endDate"] = endDate
	c.Data["packageId"] = packageId
	c.Data["packageList"] = packageList
	c.Data["typeId"] = typeId
	c.Data["aliasId"] = aliasId
	c.Data["packageAliasList"] = packageAliasList
	c.Data["aliasName"] = aliasName
	c.Data["status"] = status
	c.Data["isOperation"] = c.auth.HasAccessPerm(c.controllerName, "plane")
	c.Data["tabId"] = tabId
	c.Data["page"] = page
	c.Data["userid"] = userid
	c.Data["lossDays1"] = lossDays1
	c.Data["lossDays2"] = lossDays2
	c.Data["bets1"] = bets1
	c.Data["bets2"] = bets2
	c.Data["betAvg1"] = betAvg1
	c.Data["betAvg2"] = betAvg2
	c.display()
}

// TP统计
func (c *GameStatsController) TP() {
	startDate := c.GetString("start_date")
	endDate := c.GetString("end_date")

	status, _ := c.GetInt("status")
	page, _ := strconv.Atoi(c.GetString("page"))
	packageId := c.GetString("package_id")
	typeId, _ := c.GetInt("typeId")
	aliasId := c.GetString("alias_id")
	aliasName := c.GetString("alias_name")
	utypeId := c.GetString("utype_id")
	tabId, _ := c.GetInt("tab_id")
	if page < 1 {
		page = 1
	}
	userid := c.GetString("userid")
	lossDays1, _ := c.GetInt("lossDays1")
	lossDays2, _ := c.GetInt("lossDays2")
	bets1, _ := c.GetInt("bets1")
	bets2, _ := c.GetInt("bets2")
	betAvg1, _ := c.GetInt("betAvg1")
	betAvg2, _ := c.GetInt("betAvg2")

	if startDate == "" || endDate == "" {
		// 默认近七天数据
		today := bson.Now()
		// tim := utils.Stamp2Time(timestamp)
		endDate = today.Format("2006-01-02")
		startDate = today.AddDate(0, 0, -6).Format("2006-01-02")
	}

	// 玩家类型 账号类型 流失天数筛选 打码量筛选 局均码筛选
	m := service.FindByDate1(startDate, endDate, "date", "date")
	m2 := bson.M{}
	// 打码量, 注均码
	if bets1 > 0 && bets2 > 0 {
		m2["bets"] = bson.M{"$gte": bets1 * 100, "$lte": bets2 * 100}
	} else {
		bets1, bets2 = 0, 0
	}
	if betAvg1 > 0 && betAvg2 > 0 {
		m2["round_bet_avg"] = bson.M{"$gte": betAvg1 * 100, "$lte": betAvg2 * 100}
	} else {
		betAvg1, betAvg2 = 0, 0
	}

	m3 := bson.M{}
	if lossDays1 > 0 && lossDays2 > 0 {
		m3["loss_days"] = bson.M{
			"$gte": lossDays1,
			"$lte": lossDays2,
		}
	}

	// 明细用户类型筛选多选
	if userid != "" {
		m["userid"] = userid
	}
	var utypeIds []string
	if utypeId != "" && utypeId != "0" && utypeId != "-" {
		utypeIds = strings.Split(utypeId, ",")
	}
	var utypesM []bson.M
	for _, utypeId := range utypeIds {
		usertypeId, err := strconv.Atoi(utypeId)
		if err != nil {
			beego.Error("usertypeId error: ", err)
			continue
		}
		if usertypeId == 0 {
			continue
		}

		// 新玩家，小R，中R，大R，超大R
		/*
			新用户：充值0.00卢-99.99卢
			充值
			普充：100.00卢-999.99卢
			小R：1000.00-4999.99
			中R：充值5000.00卢-9999.99卢
			大R：充值10000.00卢-99999.99卢
			超大R：充值100000.00卢以上
		*/
		startMoney := 0
		endMoney := 0
		if usertypeId >= 5 && usertypeId <= 10 {
			switch usertypeId {
			case 5:
				// 零充
				startMoney = 0
				endMoney = 0
			case 6:
				// 普充
				startMoney = 20000.00
				endMoney = 99999
			case 7:
				// 小R
				startMoney = 100000
				endMoney = 499999
			case 8:
				// 中R
				startMoney = 500000
				endMoney = 999999
			case 9:
				// 大R
				startMoney = 1000000
				endMoney = 9999999
			case 10:
				// 超大R
				startMoney = 10000000
				endMoney = -1
			}
			if endMoney == -1 {
				utypesM = append(utypesM, bson.M{
					"money":        bson.M{"$gte": startMoney},
					"state":        2,
					"custom_types": bson.M{"$eq": nil},
				})
				// m3["money"] = bson.M{"$gte": startMoney}
				// m3["state"] = 2
				// m3["custom_types"] = bson.M{"$eq": ""}
				// m3["custom_types"] = bson.M{"$eq": nil}
			} else {
				utypesM = append(utypesM, bson.M{
					"money":        bson.M{"$gte": startMoney, "$lte": endMoney},
					"state":        2,
					"custom_types": bson.M{"$eq": nil},
				})
				// m3["money"] = bson.M{"$gte": startMoney, "$lte": endMoney}
				// m3["state"] = 2
				// m3["custom_types"] = bson.M{"$eq": ""}
				// m3["custom_types"] = bson.M{"$eq": nil}
			}
		} else {
			if usertypeId == 11 {
				utypesM = append(utypesM, bson.M{
					"custom_types": bson.M{"$ne": nil},
				})
				// m3["custom_types"] = bson.M{"$ne": ""}
				// m3["custom_types"] = bson.M{"$ne": nil}

				// m["custom_types"] = bson.M{"$ne": bson.M{}, "$exists": true}
				// m["custom_types"] = bson.M{"$ne": pb.NULL}
			} else {
				utypesM = append(utypesM, bson.M{
					"state":        usertypeId,
					"custom_types": bson.M{"$eq": nil},
				})
				// m3["state"] = usertypeId
				// m3["custom_types"] = bson.M{"$eq": ""}
				// m3["custom_types"] = bson.M{"$eq": nil}
			}
		}
	}
	if len(utypesM) > 0 {
		m3["$or"] = utypesM
	}

	if status == 0 {
		my_data := c.GetSession("my_select_pakeageid")
		if my_data != nil && packageId == "" {
			if packid, ok := my_data.(string); ok {
				packageId = packid
			}
		}
		status = 1
	}
	var packageIds []string
	if packageId != "" && packageId != "0" && packageId != "-" {
		if strings.Contains(packageId, ",") {
			palkage_arr := strings.Split(packageId, ",")
			temp_arr := make([]string, 0)
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					temp_arr = append(temp_arr, item)
				}
			}
			packageIds = append(packageIds, temp_arr...)
			// m["ad__bundle_id"] = bson.M{"$in": temp_arr}
		} else {
			packageIds = append(packageIds, packageId)
			// m["ad__bundle_id"] = packageId
		}
		// 存储缓存
		c.SetSession("my_select_pakeageid", packageId)
	} else {
		// 清除缓存
		c.DelSession("my_select_pakeageid")
	}
	// 渠道别名
	if aliasId != "" && aliasId != "0" && aliasId != "-" {
		var aliasIds []string
		if strings.Contains(aliasId, ",") {
			palkage_arr := strings.Split(aliasId, ",")
			temp_arr := make([]string, 0)
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					temp_arr = append(temp_arr, item)
				}
			}
			aliasIds = append(aliasIds, temp_arr...)
		} else {
			aliasIds = append(aliasIds, aliasId)
		}

		channels, err := service.ChannelService.GetChannelByNameAliasList(aliasIds)
		if err != nil {
			beego.Error("channel error1: ", err)
		} else {
			for _, c := range channels {
				packageIds = append(packageIds, c.Name)
			}
		}
	}
	// if aliasName != "" {
	// 	m["ad__bundle_id"] = bson.M{"$regex": aliasName, "$options": "i"}
	// }
	if len(packageIds) > 0 {
		m["ad__bundle_id"] = bson.M{"$in": packageIds}
	}

	if tabId == 1 {
		m["regist_area"] = bson.M{"$ne": 1}
	} else if tabId == 2 {
		m["regist_area"] = 1
	} else if tabId == 3 {
		m["regist_area"] = 2
	}

	var count int
	if typeId == 1 { // 明细
		players, total, err := service.GameStatsService.GetTPStatPlayers(page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetTPStatPlayers error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 2 {
		// tp房间统计
		dates, err := service.GameStatsService.GetTpPlayerRoomStat(m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetTpPlayerRoomStat error: ", err)
		}
		c.Data["list"] = dates
		count = 1
	} else if typeId == 3 {
		// tp玩家行为
		// dates := new(entity.TPPlayerRoomShow)
		// line := new(entity.TPProfitShow)
		if userid != "" {
			dates, line, _ := service.GameStatsService.GetTpPlayerStat(userid)
			count = 1
			c.Data["line"] = line
			c.Data["list"] = dates
		}

	} else if typeId == 0 { // 汇总
		dates, err := service.GameStatsService.GetTPStatDates(m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetTPStatDates error: ", err)
		}
		c.Data["list"] = dates
		count = len(dates)
	}

	usertype := map[int]string{
		0:  "全部",
		1:  "新手",
		3:  "平民",
		4:  "泡沫",
		5:  "零充",
		6:  "普充",
		7:  "小R",
		8:  "中R",
		9:  "大R",
		10: "超大R",
		11: "手动",
	}
	packageList, _ := service.ChannelService.GetChannelAll()
	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()
	c.Data["count"] = count
	if typeId == 1 {
		c.Data["pageBar"] = libs.NewPager(page, int(count), c.pageSize, beego.URLFor("GameStatsController.TP", "startDate", startDate, "endDate", endDate, "status", status, "packageId", packageId, "typeId", typeId, "aliasId", aliasId, "aliasName", aliasName, "usertypeId", utypeId, "utype_id", utypeId, "tabId", tabId, "userid", userid, "lossDays1", lossDays1, "lossDays2", lossDays2, "bets1", bets1, "bets2", bets2, "betAvg1", betAvg1, "betAvg2", betAvg2), true).ToString()
	}
	c.Data["usertype"] = usertype
	c.Data["usertypeId"] = utypeId
	c.Data["startDate"] = startDate
	c.Data["endDate"] = endDate
	c.Data["packageId"] = packageId
	c.Data["packageList"] = packageList
	c.Data["typeId"] = typeId
	c.Data["aliasId"] = aliasId
	c.Data["packageAliasList"] = packageAliasList
	c.Data["aliasName"] = aliasName
	c.Data["status"] = status
	c.Data["isOperation"] = c.auth.HasAccessPerm(c.controllerName, "crash")
	c.Data["tabId"] = tabId
	c.Data["page"] = page
	c.Data["userid"] = userid
	c.Data["lossDays1"] = lossDays1
	c.Data["lossDays2"] = lossDays2
	c.Data["bets1"] = bets1
	c.Data["bets2"] = bets2
	c.Data["betAvg1"] = betAvg1
	c.Data["betAvg2"] = betAvg2
	c.Data["pageTitle"] = "TP"
	c.display()
}

func getLHDStatsParam(userid, startDate, endDate string,
	bets1, bets2, betAvg1, betAvg2, cash1, cash2, lossDays1, lossDays2 int,
) (params map[string]any) {
	params = make(map[string]any)
	startTime, endTime := service.FindByDate4(startDate, endDate)
	if startTime != nil {
		params["startTime"] = startTime.Unix()
		params["startDate"] = startDate
		params["start"] = *startTime
	}
	if endTime != nil {
		params["endTime"] = endTime.Unix()
		params["endDate"] = endDate
		params["end"] = *endTime
	}

	// 打码量, 注均码, 净赢
	if bets1 > 0 && bets2 > 0 {
		params["bets1"] = bets1 * 100
		params["bets2"] = bets2 * 100
	} else {
		bets1, bets2 = 0, 0
	}
	if betAvg1 > 0 && betAvg2 > 0 {
		params["betAvg1"] = betAvg1 * 100
		params["betAvg2"] = betAvg2 * 100
	} else {
		betAvg1, betAvg2 = 0, 0
	}
	if cash1 != 0 && cash2 != 0 {
		params["cash1"] = cash1 * 100
		params["cash2"] = cash2 * 100
	} else {
		cash1, cash2 = 0, 0
	}

	if lossDays1 > 0 && lossDays2 > 0 {
		params["uLossDays1"] = lossDays1
		params["uLossDays2"] = lossDays2
	}

	// 明细用户类型筛选多选
	if userid != "" {
		params["userid"] = userid
	}

	return
}

// LHD统计
func (c *GameStatsController) LHD() {
	startDate := c.GetString("start_date")
	endDate := c.GetString("end_date")

	status, _ := c.GetInt("status")
	page, _ := strconv.Atoi(c.GetString("page"))
	packageId := c.GetString("package_id")
	typeId, _ := c.GetInt("typeId")
	aliasId := c.GetString("alias_id")
	aliasName := c.GetString("alias_name")
	utypeId := c.GetString("utype_id")
	tabId, _ := c.GetInt("tab_id")
	if page < 1 {
		page = 1
	}
	userid := c.GetString("userid")
	lossDays1, _ := c.GetInt("lossDays1")
	lossDays2, _ := c.GetInt("lossDays2")
	bets1, _ := c.GetInt("bets1")
	bets2, _ := c.GetInt("bets2")
	betAvg1, _ := c.GetInt("betAvg1")
	betAvg2, _ := c.GetInt("betAvg2")
	win1, _ := c.GetInt("win1")
	win2, _ := c.GetInt("win2")

	if startDate == "" || endDate == "" {
		// 默认近七天数据
		today := bson.Now()
		// tim := utils.Stamp2Time(timestamp)
		endDate = today.Format("2006-01-02")
		startDate = today.AddDate(0, 0, -6).Format("2006-01-02")
	}
	// sql 筛选条件
	params := getLHDStatsParam(userid, startDate, endDate,
		bets1, bets2, betAvg1, betAvg2, win1, win2, lossDays1, lossDays2)

	// 玩家类型 账号类型 流失天数筛选 打码量筛选 局均码筛选
	m := service.FindByDate1(startDate, endDate, "date", "date")
	m2 := bson.M{}
	// 打码量, 注均码
	if bets1 > 0 && bets2 > 0 {
		m2["bets"] = bson.M{"$gte": bets1 * 100, "$lte": bets2 * 100}
	} else {
		bets1, bets2 = 0, 0
	}
	if betAvg1 > 0 && betAvg2 > 0 {
		m2["round_bet_avg"] = bson.M{"$gte": betAvg1 * 100, "$lte": betAvg2 * 100}
	} else {
		betAvg1, betAvg2 = 0, 0
	}

	if win1 != 0 && win2 != 0 {
		if typeId == 0 || typeId == 1 { // 汇总/明细, 净赢筛选
			m2["cash"] = bson.M{"$gte": win1 * 100, "$lte": win2 * 100}
		} else {
			m2["wins"] = bson.M{"$gte": win1 * 100, "$lte": win2 * 100}
		}
	} else {
		win1, win2 = 0, 0
	}
	m3 := bson.M{}
	if lossDays1 > 0 && lossDays2 > 0 {
		m3["loss_days"] = bson.M{
			"$gte": lossDays1,
			"$lte": lossDays2,
		}
	}

	// 明细用户类型筛选多选
	if userid != "" {
		m["userid"] = userid
	}
	var utypeIds []string
	if utypeId != "" && utypeId != "0" && utypeId != "-" {
		utypeIds = strings.Split(utypeId, ",")
	}
	var utypesM []bson.M
	var utypeFilters []string
	for _, utypeId := range utypeIds {
		usertypeId, err := strconv.Atoi(utypeId)
		if err != nil {
			beego.Error("usertypeId error: ", err)
			continue
		}
		if usertypeId == 0 {
			continue
		}

		// 新玩家，小R，中R，大R，超大R
		/*
			新用户：充值0.00卢-99.99卢
			充值
			普充：100.00卢-999.99卢
			小R：1000.00-4999.99
			中R：充值5000.00卢-9999.99卢
			大R：充值10000.00卢-99999.99卢
			超大R：充值100000.00卢以上
		*/
		startMoney := 0
		endMoney := 0
		if usertypeId >= 5 && usertypeId <= 10 {
			switch usertypeId {
			case 5:
				// 零充
				utypeFilters = append(utypeFilters, "(money=0)")
				startMoney = 0
				endMoney = 0
			case 6:
				// 普充
				startMoney = 20000.00
				endMoney = 99999
				utypeFilters = append(utypeFilters, fmt.Sprintf("(money between %d and %d)", startMoney, endMoney))
			case 7:
				// 小R
				startMoney = 100000
				endMoney = 499999
				utypeFilters = append(utypeFilters, fmt.Sprintf("(money between %d and %d)", startMoney, endMoney))
			case 8:
				// 中R
				startMoney = 500000
				endMoney = 999999
				utypeFilters = append(utypeFilters, fmt.Sprintf("(money between %d and %d)", startMoney, endMoney))
			case 9:
				// 大R
				startMoney = 1000000
				endMoney = 9999999
				utypeFilters = append(utypeFilters, fmt.Sprintf("(money between %d and %d)", startMoney, endMoney))
			case 10:
				// 超大R
				startMoney = 10000000
				endMoney = -1
			}
			if endMoney == -1 {
				utypesM = append(utypesM, bson.M{
					"money":        bson.M{"$gte": startMoney},
					"state":        2,
					"custom_types": bson.M{"$eq": nil},
				})
				utypeFilters = append(utypeFilters, fmt.Sprintf("(state = 2 and money >= %d)", startMoney))
			} else {
				utypesM = append(utypesM, bson.M{
					"money":        bson.M{"$gte": startMoney, "$lte": endMoney},
					"state":        2,
					"custom_types": bson.M{"$eq": nil},
				})
				utypeFilters = append(utypeFilters, fmt.Sprintf("(state = 2 and money between %d and %d)", startMoney, endMoney))
			}
		} else {
			if usertypeId == 11 {
				utypesM = append(utypesM, bson.M{
					"custom_types": bson.M{"$ne": nil},
				})
				// m3["custom_types"] = bson.M{"$ne": ""}
				// m3["custom_types"] = bson.M{"$ne": nil}

				// m["custom_types"] = bson.M{"$ne": bson.M{}, "$exists": true}
				// m["custom_types"] = bson.M{"$ne": pb.NULL}
			} else {
				utypesM = append(utypesM, bson.M{
					"state":        usertypeId,
					"custom_types": bson.M{"$eq": nil},
				})
				utypeFilters = append(utypeFilters, fmt.Sprintf("(state = %d)", usertypeId))
			}
		}
	}
	if len(utypesM) > 0 {
		m3["$or"] = utypesM
	}
	if len(utypeFilters) > 0 {
		params["utypeQ"] = fmt.Sprintf(" and (%s)", strings.Join(utypeFilters, " or "))
	}

	if status == 0 {
		my_data := c.GetSession("my_select_pakeageid")
		if my_data != nil && packageId == "" {
			if packid, ok := my_data.(string); ok {
				packageId = packid
			}
		}
		status = 1
	}
	var packageIds []string
	if packageId != "" && packageId != "0" && packageId != "-" {
		if strings.Contains(packageId, ",") {
			palkage_arr := strings.Split(packageId, ",")
			temp_arr := make([]string, 0)
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					temp_arr = append(temp_arr, item)
				}
			}
			packageIds = append(packageIds, temp_arr...)
			// m["ad__bundle_id"] = bson.M{"$in": temp_arr}
		} else {
			packageIds = append(packageIds, packageId)
			// m["ad__bundle_id"] = packageId
		}
		// 存储缓存
		c.SetSession("my_select_pakeageid", packageId)
	} else {
		// 清除缓存
		c.DelSession("my_select_pakeageid")
	}
	// 渠道别名
	if aliasId != "" && aliasId != "0" && aliasId != "-" {
		var aliasIds []string
		if strings.Contains(aliasId, ",") {
			palkage_arr := strings.Split(aliasId, ",")
			temp_arr := make([]string, 0)
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					temp_arr = append(temp_arr, item)
				}
			}
			aliasIds = append(aliasIds, temp_arr...)
		} else {
			aliasIds = append(aliasIds, aliasId)
		}

		channels, err := service.ChannelService.GetChannelByNameAliasList(aliasIds)
		if err != nil {
			beego.Error("channel error1: ", err)
		} else {
			for _, c := range channels {
				packageIds = append(packageIds, c.Name)
			}
		}
	}
	// if aliasName != "" {
	// 	m["ad__bundle_id"] = bson.M{"$regex": aliasName, "$options": "i"}
	// }
	if len(packageIds) > 0 {
		m["ad__bundle_id"] = bson.M{"$in": packageIds}
		params["ad__bundle_id"] = packageIds
	}

	if tabId == 1 {
		m["regist_area"] = bson.M{"$ne": 1}
		params["regist_area"] = []int32{0, 3}
	} else if tabId == 2 {
		m["regist_area"] = 1
		params["regist_area"] = []int32{1}
	} else if tabId == 3 {
		m["regist_area"] = 2
		params["regist_area"] = []int32{2}
	}

	var count int
	if typeId == 1 { // 明细
		players, total, err := service.GameStatsService.GetLHDStatPlayers(page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetLHDStatPlayers error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 2 {
		// 龙虎新想事成
		players, total, err := service.GameStatsService.GetLHDXxscStat(page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetLHDXxscStat error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 3 {
		// 龙虎求死不能
		players, total, err := service.GameStatsService.GetLHDQsbnStat(page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetLHDQsbnStat error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 4 {
		// 龙虎龙狂有祸
		players, total, err := service.GameStatsService.GetLHDLkyhStat(page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetLHDLkyhStat error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 5 {
		// 龙虎高潮涌现
		players, total, err := service.GameStatsService.GetLHDGcyxStat(page, c.pageSize, params)
		if err != nil {
			beego.Error("GetLHDGcyxStat error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 0 { // 汇总
		dates, err := service.GameStatsService.GetLHDStatDates(m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetLHDStatDates error: ", err)
		}
		c.Data["list"] = dates
		count = len(dates)
	}

	usertype := map[int]string{
		0:  "全部",
		1:  "新手",
		3:  "平民",
		4:  "泡沫",
		5:  "零充",
		6:  "普充",
		7:  "小R",
		8:  "中R",
		9:  "大R",
		10: "超大R",
		11: "手动",
	}
	packageList, _ := service.ChannelService.GetChannelAll()
	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()
	c.Data["pageTitle"] = "LHD统计"
	c.Data["count"] = count
	if typeId != 0 {
		c.Data["pageBar"] = libs.NewPager(page, int(count), c.pageSize, beego.URLFor("GameStatsController.LHD", "start_date", startDate, "end_date", endDate, "status", status, "package_id", packageId, "typeId", typeId, "alias_id", aliasId, "alias_name", aliasName, "usertypeId", utypeId, "utype_id", utypeId, "tab_id", tabId, "userid", userid, "lossDays1", lossDays1, "lossDays2", lossDays2, "bets1", bets1, "bets2", bets2, "betAvg1", betAvg1, "betAvg2", betAvg2, "win1", win1, "win2", win2), true).ToString()
	}
	c.Data["usertype"] = usertype
	c.Data["usertypeId"] = utypeId
	c.Data["startDate"] = startDate
	c.Data["endDate"] = endDate
	c.Data["packageId"] = packageId
	c.Data["packageList"] = packageList
	c.Data["typeId"] = typeId
	c.Data["aliasId"] = aliasId
	c.Data["packageAliasList"] = packageAliasList
	c.Data["aliasName"] = aliasName
	c.Data["status"] = status
	c.Data["isOperation"] = c.auth.HasAccessPerm(c.controllerName, "crash")
	c.Data["tabId"] = tabId
	c.Data["page"] = page
	c.Data["userid"] = userid
	c.Data["lossDays1"] = lossDays1
	c.Data["lossDays2"] = lossDays2
	c.Data["bets1"] = bets1
	c.Data["bets2"] = bets2
	c.Data["betAvg1"] = betAvg1
	c.Data["betAvg2"] = betAvg2
	c.Data["win1"] = win1
	c.Data["win2"] = win2
	c.display()
}

// 7updown统计
func (c *GameStatsController) Updown() {
	startDate := c.GetString("start_date")
	endDate := c.GetString("end_date")

	status, _ := c.GetInt("status")
	page, _ := strconv.Atoi(c.GetString("page"))
	packageId := c.GetString("package_id")
	typeId, _ := c.GetInt("typeId")
	aliasId := c.GetString("alias_id")
	aliasName := c.GetString("alias_name")
	utypeId := c.GetString("utype_id")
	tabId, _ := c.GetInt("tab_id")
	if page < 1 {
		page = 1
	}
	userid := c.GetString("userid")
	lossDays1, _ := c.GetInt("lossDays1")
	lossDays2, _ := c.GetInt("lossDays2")
	bets1, _ := c.GetInt("bets1")
	bets2, _ := c.GetInt("bets2")
	betAvg1, _ := c.GetInt("betAvg1")
	betAvg2, _ := c.GetInt("betAvg2")
	win1, _ := c.GetInt("win1")
	win2, _ := c.GetInt("win2")

	if startDate == "" || endDate == "" {
		// 默认近七天数据
		today := bson.Now()
		// tim := utils.Stamp2Time(timestamp)
		endDate = today.Format("2006-01-02")
		startDate = today.AddDate(0, 0, -6).Format("2006-01-02")
	}

	// sql 筛选条件
	params := getLHDStatsParam(userid, startDate, endDate,
		bets1, bets2, betAvg1, betAvg2, win1, win2, lossDays1, lossDays2)

	// 玩家类型 账号类型 流失天数筛选 打码量筛选 局均码筛选
	m := service.FindByDate1(startDate, endDate, "date", "date")
	m2 := bson.M{}
	// 打码量, 注均码
	if bets1 > 0 && bets2 > 0 {
		m2["bets"] = bson.M{"$gte": bets1 * 100, "$lte": bets2 * 100}
	} else {
		bets1, bets2 = 0, 0
	}
	if betAvg1 > 0 && betAvg2 > 0 {
		m2["round_bet_avg"] = bson.M{"$gte": betAvg1 * 100, "$lte": betAvg2 * 100}
	} else {
		betAvg1, betAvg2 = 0, 0
	}

	if win1 != 0 && win2 != 0 {
		if typeId == 0 || typeId == 1 { // 汇总/明细, 净赢筛选
			m2["cash"] = bson.M{"$gte": win1 * 100, "$lte": win2 * 100}
		} else {
			m2["wins"] = bson.M{"$gte": win1 * 100, "$lte": win2 * 100}
		}
	} else {
		win1, win2 = 0, 0
	}

	m3 := bson.M{}
	if lossDays1 > 0 && lossDays2 > 0 {
		m3["loss_days"] = bson.M{
			"$gte": lossDays1,
			"$lte": lossDays2,
		}
	}

	// 明细用户类型筛选多选
	if userid != "" {
		m["userid"] = userid
	}
	var utypeIds []string
	if utypeId != "" && utypeId != "0" && utypeId != "-" {
		utypeIds = strings.Split(utypeId, ",")
	}
	var utypesM []bson.M
	var utypeFilters []string
	for _, utypeId := range utypeIds {
		usertypeId, err := strconv.Atoi(utypeId)
		if err != nil {
			beego.Error("usertypeId error: ", err)
			continue
		}
		if usertypeId == 0 {
			continue
		}

		// 新玩家，小R，中R，大R，超大R
		/*
			新用户：充值0.00卢-99.99卢
			充值
			普充：100.00卢-999.99卢
			小R：1000.00-4999.99
			中R：充值5000.00卢-9999.99卢
			大R：充值10000.00卢-99999.99卢
			超大R：充值100000.00卢以上
		*/
		startMoney := 0
		endMoney := 0
		if usertypeId >= 5 && usertypeId <= 10 {
			switch usertypeId {
			case 5:
				// 零充
				utypeFilters = append(utypeFilters, "(money=0)")
				startMoney = 0
				endMoney = 0
			case 6:
				// 普充
				startMoney = 20000.00
				endMoney = 99999
				utypeFilters = append(utypeFilters, fmt.Sprintf("(money between %d and %d)", startMoney, endMoney))
			case 7:
				// 小R
				startMoney = 100000
				endMoney = 499999
				utypeFilters = append(utypeFilters, fmt.Sprintf("(money between %d and %d)", startMoney, endMoney))
			case 8:
				// 中R
				startMoney = 500000
				endMoney = 999999
				utypeFilters = append(utypeFilters, fmt.Sprintf("(money between %d and %d)", startMoney, endMoney))
			case 9:
				// 大R
				startMoney = 1000000
				endMoney = 9999999
				utypeFilters = append(utypeFilters, fmt.Sprintf("(money between %d and %d)", startMoney, endMoney))
			case 10:
				// 超大R
				startMoney = 10000000
				endMoney = -1
			}
			if endMoney == -1 {
				utypesM = append(utypesM, bson.M{
					"money":        bson.M{"$gte": startMoney},
					"state":        2,
					"custom_types": bson.M{"$eq": nil},
				})
				utypeFilters = append(utypeFilters, fmt.Sprintf("(state = 2 and money >= %d)", startMoney))
			} else {
				utypesM = append(utypesM, bson.M{
					"money":        bson.M{"$gte": startMoney, "$lte": endMoney},
					"state":        2,
					"custom_types": bson.M{"$eq": nil},
				})
				utypeFilters = append(utypeFilters, fmt.Sprintf("(state = 2 and money between %d and %d)", startMoney, endMoney))
			}
		} else {
			if usertypeId == 11 {
				utypesM = append(utypesM, bson.M{
					"custom_types": bson.M{"$ne": nil},
				})
			} else {
				utypesM = append(utypesM, bson.M{
					"state":        usertypeId,
					"custom_types": bson.M{"$eq": nil},
				})
				utypeFilters = append(utypeFilters, fmt.Sprintf("(state = %d)", usertypeId))
			}
		}
	}
	if len(utypesM) > 0 {
		m3["$or"] = utypesM
	}
	if len(utypeFilters) > 0 {
		params["utypeQ"] = fmt.Sprintf(" and (%s)", strings.Join(utypeFilters, " or "))
	}

	if status == 0 {
		my_data := c.GetSession("my_select_pakeageid")
		if my_data != nil && packageId == "" {
			if packid, ok := my_data.(string); ok {
				packageId = packid
			}
		}
		status = 1
	}
	var packageIds []string
	if packageId != "" && packageId != "0" && packageId != "-" {
		if strings.Contains(packageId, ",") {
			palkage_arr := strings.Split(packageId, ",")
			temp_arr := make([]string, 0)
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					temp_arr = append(temp_arr, item)
				}
			}
			packageIds = append(packageIds, temp_arr...)
			// m["ad__bundle_id"] = bson.M{"$in": temp_arr}
		} else {
			packageIds = append(packageIds, packageId)
			// m["ad__bundle_id"] = packageId
		}
		// 存储缓存
		c.SetSession("my_select_pakeageid", packageId)
	} else {
		// 清除缓存
		c.DelSession("my_select_pakeageid")
	}
	// 渠道别名
	if aliasId != "" && aliasId != "0" && aliasId != "-" {
		var aliasIds []string
		if strings.Contains(aliasId, ",") {
			palkage_arr := strings.Split(aliasId, ",")
			temp_arr := make([]string, 0)
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					temp_arr = append(temp_arr, item)
				}
			}
			aliasIds = append(aliasIds, temp_arr...)
		} else {
			aliasIds = append(aliasIds, aliasId)
		}

		channels, err := service.ChannelService.GetChannelByNameAliasList(aliasIds)
		if err != nil {
			beego.Error("channel error1: ", err)
		} else {
			for _, c := range channels {
				packageIds = append(packageIds, c.Name)
			}
		}
	}
	// if aliasName != "" {
	// 	m["ad__bundle_id"] = bson.M{"$regex": aliasName, "$options": "i"}
	// }
	if len(packageIds) > 0 {
		m["ad__bundle_id"] = bson.M{"$in": packageIds}
		params["ad__bundle_id"] = packageIds
	}

	if tabId == 1 {
		m["regist_area"] = bson.M{"$ne": 1}
		params["regist_area"] = []int32{0, 3}
	} else if tabId == 2 {
		m["regist_area"] = 1
		params["regist_area"] = []int32{1}
	} else if tabId == 3 {
		m["regist_area"] = 2
		params["regist_area"] = []int32{2}
	}

	var count int
	if typeId == 1 { // 明细
		players, total, err := service.GameStatsService.GetUpdownStatPlayers(page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetUpdownStatPlayers error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 2 {
		// 7updown新想事成
		players, total, err := service.GameStatsService.GetUpdownXxscStat(page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetUpdownXxscStat error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 3 {
		// 7updown求死不能
		players, total, err := service.GameStatsService.GetUpdownQsbnStat(page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetUpdownQsbnStat error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 4 {
		// 7updown龙狂有祸
		players, total, err := service.GameStatsService.GetUpdownLkyhStat(page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetUpdownLkyhStat error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 5 {
		// 7updown高潮涌现
		players, total, err := service.GameStatsService.GetUpdownGcyxStat(page, c.pageSize, params)
		if err != nil {
			beego.Error("GetUpdownGcyxStat error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 0 { // 汇总
		dates, err := service.GameStatsService.GetUpDownStatDates(m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetUpDownStatDates error: ", err)
		}
		c.Data["list"] = dates
		count = len(dates)
	}

	usertype := map[int]string{
		0:  "全部",
		1:  "新手",
		3:  "平民",
		4:  "泡沫",
		5:  "零充",
		6:  "普充",
		7:  "小R",
		8:  "中R",
		9:  "大R",
		10: "超大R",
		11: "手动",
	}
	packageList, _ := service.ChannelService.GetChannelAll()
	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()
	c.Data["pageTitle"] = "7Updown统计"
	c.Data["count"] = count
	if typeId != 0 {
		c.Data["pageBar"] = libs.NewPager(page, int(count), c.pageSize, beego.URLFor("GameStatsController.Updown", "start_date", startDate, "end_date", endDate, "status", status, "package_id", packageId, "typeId", typeId, "alias_id", aliasId, "alias_name", aliasName, "usertypeId", utypeId, "tab_id", tabId, "utype_id", utypeId, "userid", userid, "lossDays1", lossDays1, "lossDays2", lossDays2, "bets1", bets1, "bets2", bets2, "betAvg1", betAvg1, "betAvg2", betAvg2, "win1", win1, "win2", win2), true).ToString()
	}
	c.Data["usertype"] = usertype
	c.Data["usertypeId"] = utypeId
	c.Data["startDate"] = startDate
	c.Data["endDate"] = endDate
	c.Data["packageId"] = packageId
	c.Data["packageList"] = packageList
	c.Data["typeId"] = typeId
	c.Data["aliasId"] = aliasId
	c.Data["packageAliasList"] = packageAliasList
	c.Data["aliasName"] = aliasName
	c.Data["status"] = status
	c.Data["isOperation"] = c.auth.HasAccessPerm(c.controllerName, "crash")
	c.Data["tabId"] = tabId
	c.Data["page"] = page
	c.Data["userid"] = userid
	c.Data["lossDays1"] = lossDays1
	c.Data["lossDays2"] = lossDays2
	c.Data["bets1"] = bets1
	c.Data["bets2"] = bets2
	c.Data["betAvg1"] = betAvg1
	c.Data["betAvg2"] = betAvg2
	c.Data["win1"] = win1
	c.Data["win2"] = win2
	c.display()
}

// AndarBahar统计
func (c *GameStatsController) AB() {
	startDate := c.GetString("start_date")
	endDate := c.GetString("end_date")

	status, _ := c.GetInt("status")
	page, _ := strconv.Atoi(c.GetString("page"))
	packageId := c.GetString("package_id")
	typeId, _ := c.GetInt("typeId")
	aliasId := c.GetString("alias_id")
	aliasName := c.GetString("alias_name")
	utypeId := c.GetString("utype_id")
	tabId, _ := c.GetInt("tab_id")
	if page < 1 {
		page = 1
	}
	userid := c.GetString("userid")
	lossDays1, _ := c.GetInt("lossDays1")
	lossDays2, _ := c.GetInt("lossDays2")
	bets1, _ := c.GetInt("bets1")
	bets2, _ := c.GetInt("bets2")
	betAvg1, _ := c.GetInt("betAvg1")
	betAvg2, _ := c.GetInt("betAvg2")
	win1, _ := c.GetInt("win1")
	win2, _ := c.GetInt("win2")

	if startDate == "" || endDate == "" {
		// 默认近七天数据
		today := bson.Now()
		// tim := utils.Stamp2Time(timestamp)
		endDate = today.Format("2006-01-02")
		startDate = today.AddDate(0, 0, -6).Format("2006-01-02")
	}

	// 玩家类型 账号类型 流失天数筛选 打码量筛选 局均码筛选
	m := service.FindByDate1(startDate, endDate, "date", "date")
	m2 := bson.M{}
	// 打码量, 注均码
	if bets1 > 0 && bets2 > 0 {
		m2["bets"] = bson.M{"$gte": bets1 * 100, "$lte": bets2 * 100}
	} else {
		bets1, bets2 = 0, 0
	}
	if betAvg1 > 0 && betAvg2 > 0 {
		m2["round_bet_avg"] = bson.M{"$gte": betAvg1 * 100, "$lte": betAvg2 * 100}
	} else {
		betAvg1, betAvg2 = 0, 0
	}

	if win1 != 0 && win2 != 0 {
		if typeId == 0 || typeId == 1 { // 汇总/明细, 净赢筛选
			m2["cash"] = bson.M{"$gte": win1 * 100, "$lte": win2 * 100}
		} else {
			m2["wins"] = bson.M{"$gte": win1 * 100, "$lte": win2 * 100}
		}
	} else {
		win1, win2 = 0, 0
	}
	m3 := bson.M{}
	if lossDays1 > 0 && lossDays2 > 0 {
		m3["loss_days"] = bson.M{
			"$gte": lossDays1,
			"$lte": lossDays2,
		}
	}

	// 明细用户类型筛选多选
	if userid != "" {
		m["userid"] = userid
	}
	var utypeIds []string
	if utypeId != "" && utypeId != "0" && utypeId != "-" {
		utypeIds = strings.Split(utypeId, ",")
	}
	var utypesM []bson.M
	for _, utypeId := range utypeIds {
		usertypeId, err := strconv.Atoi(utypeId)
		if err != nil {
			beego.Error("usertypeId error: ", err)
			continue
		}
		if usertypeId == 0 {
			continue
		}

		// 新玩家，小R，中R，大R，超大R
		/*
			新用户：充值0.00卢-99.99卢
			充值
			普充：100.00卢-999.99卢
			小R：1000.00-4999.99
			中R：充值5000.00卢-9999.99卢
			大R：充值10000.00卢-99999.99卢
			超大R：充值100000.00卢以上
		*/
		startMoney := 0
		endMoney := 0
		if usertypeId >= 5 && usertypeId <= 10 {
			switch usertypeId {
			case 5:
				// 零充
				startMoney = 0
				endMoney = 0
			case 6:
				// 普充
				startMoney = 20000.00
				endMoney = 99999
			case 7:
				// 小R
				startMoney = 100000
				endMoney = 499999
			case 8:
				// 中R
				startMoney = 500000
				endMoney = 999999
			case 9:
				// 大R
				startMoney = 1000000
				endMoney = 9999999
			case 10:
				// 超大R
				startMoney = 10000000
				endMoney = -1
			}
			if endMoney == -1 {
				utypesM = append(utypesM, bson.M{
					"money":        bson.M{"$gte": startMoney},
					"state":        2,
					"custom_types": bson.M{"$eq": nil},
				})
				// m3["money"] = bson.M{"$gte": startMoney}
				// m3["state"] = 2
				// m3["custom_types"] = bson.M{"$eq": ""}
				// m3["custom_types"] = bson.M{"$eq": nil}
			} else {
				utypesM = append(utypesM, bson.M{
					"money":        bson.M{"$gte": startMoney, "$lte": endMoney},
					"state":        2,
					"custom_types": bson.M{"$eq": nil},
				})
				// m3["money"] = bson.M{"$gte": startMoney, "$lte": endMoney}
				// m3["state"] = 2
				// m3["custom_types"] = bson.M{"$eq": ""}
				// m3["custom_types"] = bson.M{"$eq": nil}
			}
		} else {
			if usertypeId == 11 {
				utypesM = append(utypesM, bson.M{
					"custom_types": bson.M{"$ne": nil},
				})
				// m3["custom_types"] = bson.M{"$ne": ""}
				// m3["custom_types"] = bson.M{"$ne": nil}

				// m["custom_types"] = bson.M{"$ne": bson.M{}, "$exists": true}
				// m["custom_types"] = bson.M{"$ne": pb.NULL}
			} else {
				utypesM = append(utypesM, bson.M{
					"state":        usertypeId,
					"custom_types": bson.M{"$eq": nil},
				})
				// m3["state"] = usertypeId
				// m3["custom_types"] = bson.M{"$eq": ""}
				// m3["custom_types"] = bson.M{"$eq": nil}
			}
		}
	}
	if len(utypesM) > 0 {
		m3["$or"] = utypesM
	}

	if status == 0 {
		my_data := c.GetSession("my_select_pakeageid")
		if my_data != nil && packageId == "" {
			if packid, ok := my_data.(string); ok {
				packageId = packid
			}
		}
		status = 1
	}
	var packageIds []string
	if packageId != "" && packageId != "0" && packageId != "-" {
		if strings.Contains(packageId, ",") {
			palkage_arr := strings.Split(packageId, ",")
			temp_arr := make([]string, 0)
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					temp_arr = append(temp_arr, item)
				}
			}
			packageIds = append(packageIds, temp_arr...)
			// m["ad__bundle_id"] = bson.M{"$in": temp_arr}
		} else {
			packageIds = append(packageIds, packageId)
			// m["ad__bundle_id"] = packageId
		}
		// 存储缓存
		c.SetSession("my_select_pakeageid", packageId)
	} else {
		// 清除缓存
		c.DelSession("my_select_pakeageid")
	}
	// 渠道别名
	if aliasId != "" && aliasId != "0" && aliasId != "-" {
		var aliasIds []string
		if strings.Contains(aliasId, ",") {
			palkage_arr := strings.Split(aliasId, ",")
			temp_arr := make([]string, 0)
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					temp_arr = append(temp_arr, item)
				}
			}
			aliasIds = append(aliasIds, temp_arr...)
		} else {
			aliasIds = append(aliasIds, aliasId)
		}

		channels, err := service.ChannelService.GetChannelByNameAliasList(aliasIds)
		if err != nil {
			beego.Error("channel error1: ", err)
		} else {
			for _, c := range channels {
				packageIds = append(packageIds, c.Name)
			}
		}
	}
	// if aliasName != "" {
	// 	m["ad__bundle_id"] = bson.M{"$regex": aliasName, "$options": "i"}
	// }
	if len(packageIds) > 0 {
		m["ad__bundle_id"] = bson.M{"$in": packageIds}
	}

	if tabId == 1 {
		m["regist_area"] = bson.M{"$ne": 1}
	} else if tabId == 2 {
		m["regist_area"] = 1
	} else if tabId == 3 {
		m["regist_area"] = 2
	}

	var count int
	if typeId == 1 { // 明细
		players, total, err := service.GameStatsService.GetABStatPlayers(page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetABStatPlayers error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 2 {
		// AB 安能求死
		players, total, err := service.GameStatsService.GetABAnqsStat(page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetABAnqsStat error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 3 {
		// AB 安然躺赢
		players, total, err := service.GameStatsService.GetABArtyStat(page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetABArtyStat error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 0 { // 汇总
		dates, err := service.GameStatsService.GetABStatDates(m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetABStatDates error: ", err)
		}
		c.Data["list"] = dates
		count = len(dates)
	}

	usertype := map[int]string{
		0:  "全部",
		1:  "新手",
		3:  "平民",
		4:  "泡沫",
		5:  "零充",
		6:  "普充",
		7:  "小R",
		8:  "中R",
		9:  "大R",
		10: "超大R",
		11: "手动",
	}
	packageList, _ := service.ChannelService.GetChannelAll()
	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()
	c.Data["pageTitle"] = "AndarBahar统计"
	c.Data["count"] = count
	if typeId != 0 {
		c.Data["pageBar"] = libs.NewPager(page, int(count), c.pageSize, beego.URLFor("GameStatsController.AB", "start_date", startDate, "end_date", endDate, "status", status, "package_id", packageId, "typeId", typeId, "alias_id", aliasId, "alias_name", aliasName, "usertypeId", utypeId, "tab_id", tabId, "utype_id", utypeId, "userid", userid, "lossDays1", lossDays1, "lossDays2", lossDays2, "bets1", bets1, "bets2", bets2, "betAvg1", betAvg1, "betAvg2", betAvg2, "win1", win1, "win2", win2), true).ToString()
	}
	c.Data["usertype"] = usertype
	c.Data["usertypeId"] = utypeId
	c.Data["startDate"] = startDate
	c.Data["endDate"] = endDate
	c.Data["packageId"] = packageId
	c.Data["packageList"] = packageList
	c.Data["typeId"] = typeId
	c.Data["aliasId"] = aliasId
	c.Data["packageAliasList"] = packageAliasList
	c.Data["aliasName"] = aliasName
	c.Data["status"] = status
	c.Data["tabId"] = tabId
	c.Data["page"] = page
	c.Data["userid"] = userid
	c.Data["lossDays1"] = lossDays1
	c.Data["lossDays2"] = lossDays2
	c.Data["bets1"] = bets1
	c.Data["bets2"] = bets2
	c.Data["betAvg1"] = betAvg1
	c.Data["betAvg2"] = betAvg2
	c.Data["win1"] = win1
	c.Data["win2"] = win2
	c.display()
}

// CP 统计
func (c *GameStatsController) CP() {
	startDate := c.GetString("start_date")
	endDate := c.GetString("end_date")

	status, _ := c.GetInt("status")
	page, _ := strconv.Atoi(c.GetString("page"))
	packageId := c.GetString("package_id")
	typeId, _ := c.GetInt("typeId")
	aliasId := c.GetString("alias_id")
	aliasName := c.GetString("alias_name")
	utypeId := c.GetString("utype_id")
	tabId, _ := c.GetInt("tab_id")
	if page < 1 {
		page = 1
	}
	userid := c.GetString("userid")
	lossDays1, _ := c.GetInt("lossDays1")
	lossDays2, _ := c.GetInt("lossDays2")
	bets1, _ := c.GetInt("bets1")
	bets2, _ := c.GetInt("bets2")
	betAvg1, _ := c.GetInt("betAvg1")
	betAvg2, _ := c.GetInt("betAvg2")
	win1, _ := c.GetInt("win1")
	win2, _ := c.GetInt("win2")

	if startDate == "" || endDate == "" {
		// 默认近七天数据
		today := bson.Now()
		// tim := utils.Stamp2Time(timestamp)
		endDate = today.Format("2006-01-02")
		startDate = today.AddDate(0, 0, -6).Format("2006-01-02")
	}

	// 玩家类型 账号类型 流失天数筛选 打码量筛选 局均码筛选
	m := service.FindByDate1(startDate, endDate, "date", "date")
	m2 := bson.M{}
	// 打码量, 注均码
	if bets1 > 0 && bets2 > 0 {
		m2["bets"] = bson.M{"$gte": bets1 * 100, "$lte": bets2 * 100}
	} else {
		bets1, bets2 = 0, 0
	}
	if betAvg1 > 0 && betAvg2 > 0 {
		m2["round_bet_avg"] = bson.M{"$gte": betAvg1 * 100, "$lte": betAvg2 * 100}
	} else {
		betAvg1, betAvg2 = 0, 0
	}

	if win1 != 0 && win2 != 0 {
		if typeId == 0 || typeId == 1 { // 汇总/明细, 净赢筛选
			m2["cash"] = bson.M{"$gte": win1 * 100, "$lte": win2 * 100}
		} else {
			m2["wins"] = bson.M{"$gte": win1 * 100, "$lte": win2 * 100}
		}
	} else {
		win1, win2 = 0, 0
	}
	// if winAvg1 > 0 && winAvg2 > 0 {
	// 	m2["win_escape_multiple"] = bson.M{"$gte": winAvg1, "$lte": winAvg2}
	// } else {
	// 	winAvg1, winAvg2 = 0, 0
	// }
	m3 := bson.M{}
	if lossDays1 > 0 && lossDays2 > 0 {
		m3["loss_days"] = bson.M{
			"$gte": lossDays1,
			"$lte": lossDays2,
		}
	}

	// 明细用户类型筛选多选
	if userid != "" {
		m["userid"] = userid
	}
	var utypeIds []string
	if utypeId != "" && utypeId != "0" && utypeId != "-" {
		utypeIds = strings.Split(utypeId, ",")
	}
	var utypesM []bson.M
	for _, utypeId := range utypeIds {
		usertypeId, err := strconv.Atoi(utypeId)
		if err != nil {
			beego.Error("usertypeId error: ", err)
			continue
		}
		if usertypeId == 0 {
			continue
		}

		// 新玩家，小R，中R，大R，超大R
		/*
			新用户：充值0.00卢-99.99卢
			充值
			普充：100.00卢-999.99卢
			小R：1000.00-4999.99
			中R：充值5000.00卢-9999.99卢
			大R：充值10000.00卢-99999.99卢
			超大R：充值100000.00卢以上
		*/
		startMoney := 0
		endMoney := 0
		if usertypeId >= 5 && usertypeId <= 10 {
			switch usertypeId {
			case 5:
				// 零充
				startMoney = 0
				endMoney = 0
			case 6:
				// 普充
				startMoney = 20000.00
				endMoney = 99999
			case 7:
				// 小R
				startMoney = 100000
				endMoney = 499999
			case 8:
				// 中R
				startMoney = 500000
				endMoney = 999999
			case 9:
				// 大R
				startMoney = 1000000
				endMoney = 9999999
			case 10:
				// 超大R
				startMoney = 10000000
				endMoney = -1
			}
			if endMoney == -1 {
				utypesM = append(utypesM, bson.M{
					"money":        bson.M{"$gte": startMoney},
					"state":        2,
					"custom_types": bson.M{"$eq": nil},
				})
				// m3["money"] = bson.M{"$gte": startMoney}
				// m3["state"] = 2
				// m3["custom_types"] = bson.M{"$eq": ""}
				// m3["custom_types"] = bson.M{"$eq": nil}
			} else {
				utypesM = append(utypesM, bson.M{
					"money":        bson.M{"$gte": startMoney, "$lte": endMoney},
					"state":        2,
					"custom_types": bson.M{"$eq": nil},
				})
				// m3["money"] = bson.M{"$gte": startMoney, "$lte": endMoney}
				// m3["state"] = 2
				// m3["custom_types"] = bson.M{"$eq": ""}
				// m3["custom_types"] = bson.M{"$eq": nil}
			}
		} else {
			if usertypeId == 11 {
				utypesM = append(utypesM, bson.M{
					"custom_types": bson.M{"$ne": nil},
				})
				// m3["custom_types"] = bson.M{"$ne": ""}
				// m3["custom_types"] = bson.M{"$ne": nil}

				// m["custom_types"] = bson.M{"$ne": bson.M{}, "$exists": true}
				// m["custom_types"] = bson.M{"$ne": pb.NULL}
			} else {
				utypesM = append(utypesM, bson.M{
					"state":        usertypeId,
					"custom_types": bson.M{"$eq": nil},
				})
				// m3["state"] = usertypeId
				// m3["custom_types"] = bson.M{"$eq": ""}
				// m3["custom_types"] = bson.M{"$eq": nil}
			}
		}
	}
	if len(utypesM) > 0 {
		m3["$or"] = utypesM
	}

	if status == 0 {
		my_data := c.GetSession("my_select_pakeageid")
		if my_data != nil && packageId == "" {
			if packid, ok := my_data.(string); ok {
				packageId = packid
			}
		}
		status = 1
	}
	var packageIds []string
	if packageId != "" && packageId != "0" && packageId != "-" {
		if strings.Contains(packageId, ",") {
			palkage_arr := strings.Split(packageId, ",")
			temp_arr := make([]string, 0)
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					temp_arr = append(temp_arr, item)
				}
			}
			packageIds = append(packageIds, temp_arr...)
			// m["ad__bundle_id"] = bson.M{"$in": temp_arr}
		} else {
			packageIds = append(packageIds, packageId)
			// m["ad__bundle_id"] = packageId
		}
		// 存储缓存
		c.SetSession("my_select_pakeageid", packageId)
	} else {
		// 清除缓存
		c.DelSession("my_select_pakeageid")
	}
	// 渠道别名
	if aliasId != "" && aliasId != "0" && aliasId != "-" {
		var aliasIds []string
		if strings.Contains(aliasId, ",") {
			palkage_arr := strings.Split(aliasId, ",")
			temp_arr := make([]string, 0)
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					temp_arr = append(temp_arr, item)
				}
			}
			aliasIds = append(aliasIds, temp_arr...)
		} else {
			aliasIds = append(aliasIds, aliasId)
		}

		channels, err := service.ChannelService.GetChannelByNameAliasList(aliasIds)
		if err != nil {
			beego.Error("channel error1: ", err)
		} else {
			for _, c := range channels {
				packageIds = append(packageIds, c.Name)
			}
		}
	}
	// if aliasName != "" {
	// 	m["ad__bundle_id"] = bson.M{"$regex": aliasName, "$options": "i"}
	// }
	if len(packageIds) > 0 {
		m["ad__bundle_id"] = bson.M{"$in": packageIds}
	}

	if tabId == 1 {
		m["regist_area"] = bson.M{"$ne": 1}
	} else if tabId == 2 {
		m["regist_area"] = 1
	} else if tabId == 3 {
		m["regist_area"] = 2
	}

	var count int
	if typeId == 1 { // 明细
		players, total, err := service.GameStatsService.GetCPStatPlayers(page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetCPStatPlayers error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 2 {
		// CP 来玩就赢
		players, total, err := service.GameStatsService.GetCPLwjyStat(page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetCPLwjyStat error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 3 {
		// CP 来易去难
		players, total, err := service.GameStatsService.GetCPLyqnStat(page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetCPLyqnStat error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 0 { // 汇总
		dates, err := service.GameStatsService.GetCPStatDates(m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetCPStatDates error: ", err)
		}
		c.Data["list"] = dates
		count = len(dates)
	}

	usertype := map[int]string{
		0:  "全部",
		1:  "新手",
		3:  "平民",
		4:  "泡沫",
		5:  "零充",
		6:  "普充",
		7:  "小R",
		8:  "中R",
		9:  "大R",
		10: "超大R",
		11: "手动",
	}
	packageList, _ := service.ChannelService.GetChannelAll()
	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()
	c.Data["pageTitle"] = "CP统计"
	c.Data["count"] = count
	if typeId != 0 {
		c.Data["pageBar"] = libs.NewPager(page, int(count), c.pageSize, beego.URLFor("GameStatsController.CP", "start_date", startDate, "end_date", endDate, "status", status, "package_id", packageId, "typeId", typeId, "alias_id", aliasId, "alias_name", aliasName, "usertypeId", utypeId, "tab_id", tabId, "utype_id", utypeId, "userid", userid, "lossDays1", lossDays1, "lossDays2", lossDays2, "bets1", bets1, "bets2", bets2, "betAvg1", betAvg1, "betAvg2", betAvg2, "win1", win1, "win2", win2), true).ToString()
	}
	c.Data["usertype"] = usertype
	c.Data["usertypeId"] = utypeId
	c.Data["startDate"] = startDate
	c.Data["endDate"] = endDate
	c.Data["packageId"] = packageId
	c.Data["packageList"] = packageList
	c.Data["typeId"] = typeId
	c.Data["aliasId"] = aliasId
	c.Data["packageAliasList"] = packageAliasList
	c.Data["aliasName"] = aliasName
	c.Data["status"] = status
	c.Data["tabId"] = tabId
	c.Data["page"] = page
	c.Data["userid"] = userid
	c.Data["lossDays1"] = lossDays1
	c.Data["lossDays2"] = lossDays2
	c.Data["bets1"] = bets1
	c.Data["bets2"] = bets2
	c.Data["betAvg1"] = betAvg1
	c.Data["betAvg2"] = betAvg2
	c.Data["win1"] = win1
	c.Data["win2"] = win2
	c.display()
}

// RB 统计
func (c *GameStatsController) RB() {
	startDate := c.GetString("start_date")
	endDate := c.GetString("end_date")

	status, _ := c.GetInt("status")
	page, _ := strconv.Atoi(c.GetString("page"))
	packageId := c.GetString("package_id")
	typeId, _ := c.GetInt("typeId")
	aliasId := c.GetString("alias_id")
	aliasName := c.GetString("alias_name")
	utypeId := c.GetString("utype_id")
	tabId, _ := c.GetInt("tab_id")
	if page < 1 {
		page = 1
	}
	userid := c.GetString("userid")
	lossDays1, _ := c.GetInt("lossDays1")
	lossDays2, _ := c.GetInt("lossDays2")
	bets1, _ := c.GetInt("bets1")
	bets2, _ := c.GetInt("bets2")
	betAvg1, _ := c.GetInt("betAvg1")
	betAvg2, _ := c.GetInt("betAvg2")
	win1, _ := c.GetInt("win1")
	win2, _ := c.GetInt("win2")

	if startDate == "" || endDate == "" {
		// 默认近七天数据
		today := bson.Now()
		// tim := utils.Stamp2Time(timestamp)
		endDate = today.Format("2006-01-02")
		startDate = today.AddDate(0, 0, -6).Format("2006-01-02")
	}

	// 玩家类型 账号类型 流失天数筛选 打码量筛选 局均码筛选
	m := service.FindByDate1(startDate, endDate, "date", "date")
	m2 := bson.M{}
	// 打码量, 注均码
	if bets1 > 0 && bets2 > 0 {
		m2["bets"] = bson.M{"$gte": bets1 * 100, "$lte": bets2 * 100}
	} else {
		bets1, bets2 = 0, 0
	}
	if betAvg1 > 0 && betAvg2 > 0 {
		m2["round_bet_avg"] = bson.M{"$gte": betAvg1 * 100, "$lte": betAvg2 * 100}
	} else {
		betAvg1, betAvg2 = 0, 0
	}

	if win1 != 0 && win2 != 0 {
		if typeId == 0 || typeId == 1 { // 汇总/明细, 净赢筛选
			m2["cash"] = bson.M{"$gte": win1 * 100, "$lte": win2 * 100}
		} else {
			m2["wins"] = bson.M{"$gte": win1 * 100, "$lte": win2 * 100}
		}
	} else {
		win1, win2 = 0, 0
	}
	// if winAvg1 > 0 && winAvg2 > 0 {
	// 	m2["win_escape_multiple"] = bson.M{"$gte": winAvg1, "$lte": winAvg2}
	// } else {
	// 	winAvg1, winAvg2 = 0, 0
	// }
	m3 := bson.M{}
	if lossDays1 > 0 && lossDays2 > 0 {
		m3["loss_days"] = bson.M{
			"$gte": lossDays1,
			"$lte": lossDays2,
		}
	}

	// 明细用户类型筛选多选
	if userid != "" {
		m["userid"] = userid
	}
	var utypeIds []string
	if utypeId != "" && utypeId != "0" && utypeId != "-" {
		utypeIds = strings.Split(utypeId, ",")
	}
	var utypesM []bson.M
	for _, utypeId := range utypeIds {
		usertypeId, err := strconv.Atoi(utypeId)
		if err != nil {
			beego.Error("usertypeId error: ", err)
			continue
		}
		if usertypeId == 0 {
			continue
		}

		// 新玩家，小R，中R，大R，超大R
		/*
			新用户：充值0.00卢-99.99卢
			充值
			普充：100.00卢-999.99卢
			小R：1000.00-4999.99
			中R：充值5000.00卢-9999.99卢
			大R：充值10000.00卢-99999.99卢
			超大R：充值100000.00卢以上
		*/
		startMoney := 0
		endMoney := 0
		if usertypeId >= 5 && usertypeId <= 10 {
			switch usertypeId {
			case 5:
				// 零充
				startMoney = 0
				endMoney = 0
			case 6:
				// 普充
				startMoney = 20000.00
				endMoney = 99999
			case 7:
				// 小R
				startMoney = 100000
				endMoney = 499999
			case 8:
				// 中R
				startMoney = 500000
				endMoney = 999999
			case 9:
				// 大R
				startMoney = 1000000
				endMoney = 9999999
			case 10:
				// 超大R
				startMoney = 10000000
				endMoney = -1
			}
			if endMoney == -1 {
				utypesM = append(utypesM, bson.M{
					"money":        bson.M{"$gte": startMoney},
					"state":        2,
					"custom_types": bson.M{"$eq": nil},
				})
				// m3["money"] = bson.M{"$gte": startMoney}
				// m3["state"] = 2
				// m3["custom_types"] = bson.M{"$eq": ""}
				// m3["custom_types"] = bson.M{"$eq": nil}
			} else {
				utypesM = append(utypesM, bson.M{
					"money":        bson.M{"$gte": startMoney, "$lte": endMoney},
					"state":        2,
					"custom_types": bson.M{"$eq": nil},
				})
				// m3["money"] = bson.M{"$gte": startMoney, "$lte": endMoney}
				// m3["state"] = 2
				// m3["custom_types"] = bson.M{"$eq": ""}
				// m3["custom_types"] = bson.M{"$eq": nil}
			}
		} else {
			if usertypeId == 11 {
				utypesM = append(utypesM, bson.M{
					"custom_types": bson.M{"$ne": nil},
				})
				// m3["custom_types"] = bson.M{"$ne": ""}
				// m3["custom_types"] = bson.M{"$ne": nil}

				// m["custom_types"] = bson.M{"$ne": bson.M{}, "$exists": true}
				// m["custom_types"] = bson.M{"$ne": pb.NULL}
			} else {
				utypesM = append(utypesM, bson.M{
					"state":        usertypeId,
					"custom_types": bson.M{"$eq": nil},
				})
				// m3["state"] = usertypeId
				// m3["custom_types"] = bson.M{"$eq": ""}
				// m3["custom_types"] = bson.M{"$eq": nil}
			}
		}
	}
	if len(utypesM) > 0 {
		m3["$or"] = utypesM
	}

	if status == 0 {
		my_data := c.GetSession("my_select_pakeageid")
		if my_data != nil && packageId == "" {
			if packid, ok := my_data.(string); ok {
				packageId = packid
			}
		}
		status = 1
	}
	var packageIds []string
	if packageId != "" && packageId != "0" && packageId != "-" {
		if strings.Contains(packageId, ",") {
			palkage_arr := strings.Split(packageId, ",")
			temp_arr := make([]string, 0)
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					temp_arr = append(temp_arr, item)
				}
			}
			packageIds = append(packageIds, temp_arr...)
			// m["ad__bundle_id"] = bson.M{"$in": temp_arr}
		} else {
			packageIds = append(packageIds, packageId)
			// m["ad__bundle_id"] = packageId
		}
		// 存储缓存
		c.SetSession("my_select_pakeageid", packageId)
	} else {
		// 清除缓存
		c.DelSession("my_select_pakeageid")
	}
	// 渠道别名
	if aliasId != "" && aliasId != "0" && aliasId != "-" {
		var aliasIds []string
		if strings.Contains(aliasId, ",") {
			palkage_arr := strings.Split(aliasId, ",")
			temp_arr := make([]string, 0)
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					temp_arr = append(temp_arr, item)
				}
			}
			aliasIds = append(aliasIds, temp_arr...)
		} else {
			aliasIds = append(aliasIds, aliasId)
		}

		channels, err := service.ChannelService.GetChannelByNameAliasList(aliasIds)
		if err != nil {
			beego.Error("channel error1: ", err)
		} else {
			for _, c := range channels {
				packageIds = append(packageIds, c.Name)
			}
		}
	}
	// if aliasName != "" {
	// 	m["ad__bundle_id"] = bson.M{"$regex": aliasName, "$options": "i"}
	// }
	if len(packageIds) > 0 {
		m["ad__bundle_id"] = bson.M{"$in": packageIds}
	}

	if tabId == 1 {
		m["regist_area"] = bson.M{"$ne": 1}
	} else if tabId == 2 {
		m["regist_area"] = 1
	} else if tabId == 3 {
		m["regist_area"] = 2
	}

	var count int
	if typeId == 1 { // 明细
		players, total, err := service.GameStatsService.GetRBStatPlayers(page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetRBStatPlayers error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 2 {
		// RB 红运当头
		players, total, err := service.GameStatsService.GetRBHydtStat(page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetRBHydtStat error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 3 {
		// RB 绝处逢生
		players, total, err := service.GameStatsService.GetRBJcfsStat(page, c.pageSize, m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetRBJcfsStat error: ", err)
		}
		c.Data["list"] = players
		count = total
	} else if typeId == 0 { // 汇总
		dates, err := service.GameStatsService.GetRBStatDates(m, m2, m3, startDate, endDate)
		if err != nil {
			beego.Error("GetRBStatDates error: ", err)
		}
		c.Data["list"] = dates
		count = len(dates)
	}

	usertype := map[int]string{
		0:  "全部",
		1:  "新手",
		3:  "平民",
		4:  "泡沫",
		5:  "零充",
		6:  "普充",
		7:  "小R",
		8:  "中R",
		9:  "大R",
		10: "超大R",
		11: "手动",
	}
	packageList, _ := service.ChannelService.GetChannelAll()
	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()
	c.Data["pageTitle"] = "红黑统计"
	c.Data["count"] = count
	if typeId != 0 {
		c.Data["pageBar"] = libs.NewPager(page, int(count), c.pageSize, beego.URLFor("GameStatsController.RB", "start_date", startDate, "end_date", endDate, "status", status, "package_id", packageId, "typeId", typeId, "alias_id", aliasId, "alias_name", aliasName, "usertypeId", utypeId, "tab_id", tabId, "utype_id", utypeId, "userid", userid, "lossDays1", lossDays1, "lossDays2", lossDays2, "bets1", bets1, "bets2", bets2, "betAvg1", betAvg1, "betAvg2", betAvg2, "win1", win1, "win2", win2), true).ToString()
	}
	c.Data["usertype"] = usertype
	c.Data["usertypeId"] = utypeId
	c.Data["startDate"] = startDate
	c.Data["endDate"] = endDate
	c.Data["packageId"] = packageId
	c.Data["packageList"] = packageList
	c.Data["typeId"] = typeId
	c.Data["aliasId"] = aliasId
	c.Data["packageAliasList"] = packageAliasList
	c.Data["aliasName"] = aliasName
	c.Data["status"] = status
	c.Data["tabId"] = tabId
	c.Data["page"] = page
	c.Data["userid"] = userid
	c.Data["lossDays1"] = lossDays1
	c.Data["lossDays2"] = lossDays2
	c.Data["bets1"] = bets1
	c.Data["bets2"] = bets2
	c.Data["betAvg1"] = betAvg1
	c.Data["betAvg2"] = betAvg2
	c.Data["win1"] = win1
	c.Data["win2"] = win2
	c.display()
}

// TP统计
func (c *GameStatsController) TPNew() {
	startDate := c.GetString("start_date")
	endDate := c.GetString("end_date")

	status, _ := c.GetInt("status")
	page, _ := strconv.Atoi(c.GetString("page"))
	packageId := c.GetString("package_id")
	typeId, _ := c.GetInt("typeId")
	aliasId := c.GetString("alias_id")
	aliasName := c.GetString("alias_name")
	utypeId := c.GetString("utype_id")
	tabId, _ := c.GetInt("tab_id")
	if page < 1 {
		page = 1
	}
	userid := c.GetString("userid")
	lossDays1, _ := c.GetInt("lossDays1")
	lossDays2, _ := c.GetInt("lossDays2")
	bets1, _ := c.GetInt("bets1")
	bets2, _ := c.GetInt("bets2")
	betAvg1, _ := c.GetInt("betAvg1")
	betAvg2, _ := c.GetInt("betAvg2")
	cash1, _ := c.GetInt("cash1")
	cash2, _ := c.GetInt("cash2")

	var params = make(map[string]any)
	if startDate == "" || endDate == "" {
		// 默认近七天数据
		today := bson.Now()
		endDate = today.Format("2006-01-02")
		startDate = today.AddDate(0, 0, -6).Format("2006-01-02")
	}

	// 玩家类型 账号类型 流失天数筛选 打码量筛选 局均码筛选
	startTime, endTime := service.FindByDate4(startDate, endDate)
	if startTime != nil {
		params["startTime"] = startTime.Unix()
		params["startDate"] = startDate
		params["start"] = *startTime
	}
	if endTime != nil {
		params["endTime"] = endTime.Unix()
		params["endDate"] = endDate
		params["end"] = *endTime
	}

	// 打码量, 注均码, 净赢
	if bets1 > 0 && bets2 > 0 {
		params["bets1"] = bets1 * 100
		params["bets2"] = bets2 * 100
	} else {
		bets1, bets2 = 0, 0
	}
	if betAvg1 > 0 && betAvg2 > 0 {
		params["betAvg1"] = betAvg1 * 100
		params["betAvg2"] = betAvg2 * 100
	} else {
		betAvg1, betAvg2 = 0, 0
	}
	if cash1 != 0 && cash2 != 0 {
		params["cash1"] = cash1 * 100
		params["cash2"] = cash2 * 100
	} else {
		cash1, cash2 = 0, 0
	}

	if lossDays1 > 0 && lossDays2 > 0 {
		params["uLossDays1"] = lossDays1
		params["uLossDays2"] = lossDays2
	}

	// 明细用户类型筛选多选
	if userid != "" {
		params["userid"] = userid
	}

	var utypeIds []string
	if utypeId != "" && utypeId != "0" && utypeId != "-" {
		utypeIds = strings.Split(utypeId, ",")
	}

	var utypeFilters []string
	for _, utypeId := range utypeIds {
		usertypeId, err := strconv.Atoi(utypeId)
		if err != nil {
			beego.Error("usertypeId error: ", err)
			continue
		}
		if usertypeId == 0 {
			continue
		}

		// 新玩家，小R，中R，大R，超大R
		/*
			新用户：充值0.00卢-99.99卢
			充值
			普充：100.00卢-999.99卢
			小R：1000.00-4999.99
			中R：充值5000.00卢-9999.99卢
			大R：充值10000.00卢-99999.99卢
			超大R：充值100000.00卢以上
		*/
		startMoney := 0
		endMoney := 0
		if usertypeId >= 5 && usertypeId <= 10 {
			switch usertypeId {
			case 5:
				// 零充
				utypeFilters = append(utypeFilters, "(money=0)")
				startMoney = 0
				endMoney = 0
			case 6:
				// 普充
				startMoney = 20000.00
				endMoney = 99999
				utypeFilters = append(utypeFilters, fmt.Sprintf("(money between %d and %d)", startMoney, endMoney))
			case 7:
				// 小R
				startMoney = 100000
				endMoney = 499999
				utypeFilters = append(utypeFilters, fmt.Sprintf("(money between %d and %d)", startMoney, endMoney))
			case 8:
				// 中R
				startMoney = 500000
				endMoney = 999999
				utypeFilters = append(utypeFilters, fmt.Sprintf("(money between %d and %d)", startMoney, endMoney))
			case 9:
				// 大R
				startMoney = 1000000
				endMoney = 9999999
				utypeFilters = append(utypeFilters, fmt.Sprintf("(money between %d and %d)", startMoney, endMoney))
			case 10:
				// 超大R
				startMoney = 10000000
				endMoney = -1
			}
			if endMoney == -1 {
				utypeFilters = append(utypeFilters, fmt.Sprintf("(state = 2 and money >= %d)", startMoney))
			} else {
				utypeFilters = append(utypeFilters, fmt.Sprintf("(state = 2 and money between %d and %d)", startMoney, endMoney))
			}
		} else {
			if usertypeId == 11 {
			} else {
				// params["uState"] = usertypeId
				utypeFilters = append(utypeFilters, fmt.Sprintf("(state = %d)", usertypeId))
			}
		}
	}
	if len(utypeFilters) > 0 {
		params["utypeQ"] = fmt.Sprintf(" and (%s)", strings.Join(utypeFilters, " or "))
	}

	if status == 0 {
		my_data := c.GetSession("my_select_pakeageid")
		if my_data != nil && packageId == "" {
			if packid, ok := my_data.(string); ok {
				packageId = packid
			}
		}
		status = 1
	}
	var packageIds []string
	if packageId != "" && packageId != "0" && packageId != "-" {
		if strings.Contains(packageId, ",") {
			palkage_arr := strings.Split(packageId, ",")
			temp_arr := make([]string, 0)
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					temp_arr = append(temp_arr, item)
				}
			}
			packageIds = append(packageIds, temp_arr...)
			// m["ad__bundle_id"] = bson.M{"$in": temp_arr}
		} else {
			packageIds = append(packageIds, packageId)
			// m["ad__bundle_id"] = packageId
		}
		// 存储缓存
		c.SetSession("my_select_pakeageid", packageId)
	} else {
		// 清除缓存
		c.DelSession("my_select_pakeageid")
	}
	// 渠道别名
	if aliasId != "" && aliasId != "0" && aliasId != "-" {
		var aliasIds []string
		if strings.Contains(aliasId, ",") {
			palkage_arr := strings.Split(aliasId, ",")
			temp_arr := make([]string, 0)
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					temp_arr = append(temp_arr, item)
				}
			}
			aliasIds = append(aliasIds, temp_arr...)
		} else {
			aliasIds = append(aliasIds, aliasId)
		}

		channels, err := service.ChannelService.GetChannelByNameAliasList(aliasIds)
		if err != nil {
			beego.Error("channel error1: ", err)
		} else {
			for _, c := range channels {
				packageIds = append(packageIds, c.Name)
			}
		}
	}
	// if aliasName != "" {
	// 	m["ad__bundle_id"] = bson.M{"$regex": aliasName, "$options": "i"}
	// }
	if len(packageIds) > 0 {
		params["ad__bundle_id"] = packageIds
	}

	if tabId == 1 {
		params["regist_area"] = []int32{0, 3}
	} else if tabId == 2 {
		params["regist_area"] = []int32{1}
	} else if tabId == 3 {
		params["regist_area"] = []int32{2}
	}

	var count int
	if typeId == 1 { // 明细
		stats, total, err := service.GameStatsService.TpNewStatPlayers(page, c.pageSize, params)
		if err != nil {
			beego.Error("GetTPStatPlayers error: ", err)
		}
		c.Data["list"] = stats
		count = total
	} else if typeId == 2 {
		// 乐极生悲
		stats, total, err := service.GameStatsService.TpNewStatLjsb(page, c.pageSize, params)
		if err != nil {
			beego.Error("TpNewStatLjsb error: ", err)
		}
		c.Data["list"] = stats
		count = total
	} else if typeId == 3 {
		// 高潮涌现
		stats, total, err := service.GameStatsService.TpNewStatGCYX(page, c.pageSize, params)
		if err != nil {
			beego.Error("TpNewStatLjsb error: ", err)
		}
		c.Data["list"] = stats
		count = total
	} else if typeId == 0 { // 汇总
		dates, err := service.GameStatsService.TpNewStatDates(params)
		if err != nil {
			beego.Error("TpNewStatDates error: ", err)
		}
		c.Data["list"] = dates
		count = len(dates)
	}

	usertype := map[int]string{
		0:  "全部",
		1:  "新手",
		3:  "平民",
		4:  "泡沫",
		5:  "零充",
		6:  "普充",
		7:  "小R",
		8:  "中R",
		9:  "大R",
		10: "超大R",
		// 11: "手动",
	}
	packageList, _ := service.ChannelService.GetChannelAll()
	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()
	c.Data["count"] = count
	if typeId != 0 {
		c.Data["pageBar"] = libs.NewPager(page, int(count), c.pageSize, beego.URLFor("GameStatsController.TPNew", "startDate", startDate, "endDate", endDate, "status", status, "packageId", packageId, "typeId", typeId, "aliasId", aliasId, "aliasName", aliasName, "usertypeId", utypeId, "utype_id", utypeId, "tabId", tabId, "userid", userid, "lossDays1", lossDays1, "lossDays2", lossDays2, "bets1", bets1, "bets2", bets2, "betAvg1", betAvg1, "betAvg2", betAvg2, "cash1", cash1, "cash2", cash2), true).ToString()
	}
	c.Data["usertype"] = usertype
	c.Data["usertypeId"] = utypeId
	c.Data["startDate"] = startDate
	c.Data["endDate"] = endDate
	c.Data["packageId"] = packageId
	c.Data["packageList"] = packageList
	c.Data["typeId"] = typeId
	c.Data["aliasId"] = aliasId
	c.Data["packageAliasList"] = packageAliasList
	c.Data["aliasName"] = aliasName
	c.Data["status"] = status
	c.Data["isOperation"] = c.auth.HasAccessPerm(c.controllerName, "crash")
	c.Data["tabId"] = tabId
	c.Data["page"] = page
	c.Data["userid"] = userid
	c.Data["lossDays1"] = lossDays1
	c.Data["lossDays2"] = lossDays2
	c.Data["bets1"] = bets1
	c.Data["bets2"] = bets2
	c.Data["betAvg1"] = betAvg1
	c.Data["betAvg2"] = betAvg2
	c.Data["cash1"] = cash1
	c.Data["cash2"] = cash2
	c.Data["pageTitle"] = "TP"
	c.display()
}

// Mines统计
func (c *GameStatsController) Mines() {
	startDate := c.GetString("start_date")
	endDate := c.GetString("end_date")

	status, _ := c.GetInt("status")
	page, _ := strconv.Atoi(c.GetString("page"))
	packageId := c.GetString("package_id")
	typeId, _ := c.GetInt("typeId")
	aliasId := c.GetString("alias_id")
	aliasName := c.GetString("alias_name")
	classId := c.GetString("class_id")
	utypeId := c.GetString("utype_id")
	tabId, _ := c.GetInt("tab_id")
	if page < 1 {
		page = 1
	}
	userid := c.GetString("userid")
	lossDays1, _ := c.GetInt("lossDays1")
	lossDays2, _ := c.GetInt("lossDays2")
	bets1, _ := c.GetInt("bets1")
	bets2, _ := c.GetInt("bets2")
	betAvg1, _ := c.GetInt("betAvg1")
	betAvg2, _ := c.GetInt("betAvg2")
	cash1, _ := c.GetInt("cash1")
	cash2, _ := c.GetInt("cash2")

	var params = make(map[string]any)
	if startDate == "" || endDate == "" {
		// 默认近七天数据
		today := bson.Now()
		endDate = today.Format("2006-01-02")
		startDate = today.AddDate(0, 0, -6).Format("2006-01-02")
	}

	// 玩家类型 账号类型 流失天数筛选 打码量筛选 局均码筛选
	startTime, endTime := service.FindByDate4(startDate, endDate)
	if startTime != nil {
		params["startTime"] = startTime.Unix()
		params["startDate"] = startDate
		params["start"] = *startTime
	}
	if endTime != nil {
		params["endTime"] = endTime.Unix()
		params["endDate"] = endDate
		params["end"] = *endTime
	}

	// 打码量, 注均码, 净赢
	if bets1 > 0 && bets2 > 0 {
		params["bets1"] = bets1 * 100
		params["bets2"] = bets2 * 100
	} else {
		bets1, bets2 = 0, 0
	}
	if betAvg1 > 0 && betAvg2 > 0 {
		params["betAvg1"] = betAvg1 * 100
		params["betAvg2"] = betAvg2 * 100
	} else {
		betAvg1, betAvg2 = 0, 0
	}
	if cash1 != 0 && cash2 != 0 {
		params["cash1"] = cash1 * 100
		params["cash2"] = cash2 * 100
	} else {
		cash1, cash2 = 0, 0
	}

	if lossDays1 > 0 && lossDays2 > 0 {
		params["uLossDays1"] = lossDays1
		params["uLossDays2"] = lossDays2
	}

	// 明细用户类型筛选多选
	if userid != "" {
		params["userid"] = userid
	}

	var utypeIds []string
	if utypeId != "" && utypeId != "0" && utypeId != "-" {
		utypeIds = strings.Split(utypeId, ",")
	}

	var utypeFilters []string
	for _, utypeId := range utypeIds {
		usertypeId, err := strconv.Atoi(utypeId)
		if err != nil {
			beego.Error("usertypeId error: ", err)
			continue
		}
		if usertypeId == 0 {
			continue
		}

		// 新玩家，小R，中R，大R，超大R
		/*
			新用户：充值0.00卢-99.99卢
			充值
			普充：100.00卢-999.99卢
			小R：1000.00-4999.99
			中R：充值5000.00卢-9999.99卢
			大R：充值10000.00卢-99999.99卢
			超大R：充值100000.00卢以上
		*/
		startMoney := 0
		endMoney := 0
		if usertypeId >= 5 && usertypeId <= 10 {
			switch usertypeId {
			case 5:
				// 零充
				utypeFilters = append(utypeFilters, "(money=0)")
				startMoney = 0
				endMoney = 0
			case 6:
				// 普充
				startMoney = 20000.00
				endMoney = 99999
				utypeFilters = append(utypeFilters, fmt.Sprintf("(money between %d and %d)", startMoney, endMoney))
			case 7:
				// 小R
				startMoney = 100000
				endMoney = 499999
				utypeFilters = append(utypeFilters, fmt.Sprintf("(money between %d and %d)", startMoney, endMoney))
			case 8:
				// 中R
				startMoney = 500000
				endMoney = 999999
				utypeFilters = append(utypeFilters, fmt.Sprintf("(money between %d and %d)", startMoney, endMoney))
			case 9:
				// 大R
				startMoney = 1000000
				endMoney = 9999999
				utypeFilters = append(utypeFilters, fmt.Sprintf("(money between %d and %d)", startMoney, endMoney))
			case 10:
				// 超大R
				startMoney = 10000000
				endMoney = -1
			}
			if endMoney == -1 {
				utypeFilters = append(utypeFilters, fmt.Sprintf("(state = 2 and money >= %d)", startMoney))
			} else {
				utypeFilters = append(utypeFilters, fmt.Sprintf("(state = 2 and money between %d and %d)", startMoney, endMoney))
			}
		} else {
			if usertypeId == 11 {
			} else {
				// params["uState"] = usertypeId
				utypeFilters = append(utypeFilters, fmt.Sprintf("(state = %d)", usertypeId))
			}
		}
	}
	if len(utypeFilters) > 0 {
		params["utypeQ"] = fmt.Sprintf(" and (%s)", strings.Join(utypeFilters, " or "))
	}

	if status == 0 {
		my_data := c.GetSession("my_select_pakeageid")
		if my_data != nil && packageId == "" {
			if packid, ok := my_data.(string); ok {
				packageId = packid
			}
		}
		status = 1
	}
	var packageIds []string
	if packageId != "" && packageId != "0" && packageId != "-" {
		if strings.Contains(packageId, ",") {
			palkage_arr := strings.Split(packageId, ",")
			temp_arr := make([]string, 0)
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					temp_arr = append(temp_arr, item)
				}
			}
			packageIds = append(packageIds, temp_arr...)
			// m["ad__bundle_id"] = bson.M{"$in": temp_arr}
		} else {
			packageIds = append(packageIds, packageId)
			// m["ad__bundle_id"] = packageId
		}
		// 存储缓存
		c.SetSession("my_select_pakeageid", packageId)
	} else {
		// 清除缓存
		c.DelSession("my_select_pakeageid")
	}
	// 渠道别名
	if aliasId != "" && aliasId != "0" && aliasId != "-" {
		var aliasIds []string
		if strings.Contains(aliasId, ",") {
			palkage_arr := strings.Split(aliasId, ",")
			temp_arr := make([]string, 0)
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					temp_arr = append(temp_arr, item)
				}
			}
			aliasIds = append(aliasIds, temp_arr...)
		} else {
			aliasIds = append(aliasIds, aliasId)
		}

		channels, err := service.ChannelService.GetChannelByNameAliasList(aliasIds)
		if err != nil {
			beego.Error("channel error1: ", err)
		} else {
			for _, c := range channels {
				packageIds = append(packageIds, c.Name)
			}
		}
	}

	// 渠道类
	if classId != "" && classId != "0" && classId != "-" {
		var classIds []string
		if strings.Contains(classId, ",") {
			palkage_arr := strings.Split(classId, ",")
			temp_arr := make([]string, 0)
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					temp_arr = append(temp_arr, item)
				}
			}
			classIds = append(classIds, temp_arr...)
		} else {
			classIds = append(classIds, classId)
		}

		channels, err := service.ChannelService.GetChannelByClassList(classIds)
		if err != nil {
			beego.Error("channel error1: ", err)
		} else {
			for _, c := range channels {
				packageIds = append(packageIds, c.Name)
			}
		}
	}

	// if aliasName != "" {
	// 	m["ad__bundle_id"] = bson.M{"$regex": aliasName, "$options": "i"}
	// }
	if len(packageIds) > 0 {
		params["ad__bundle_id"] = packageIds
	}

	if tabId == 1 {
		params["regist_area"] = []int32{0, 3}
	} else if tabId == 2 {
		params["regist_area"] = []int32{1}
	} else if tabId == 3 {
		params["regist_area"] = []int32{2}
	}

	var count int
	if typeId == 1 { // 明细
		stats, total, err := service.GameStatsService.MinesStatPlayers(page, c.pageSize, params)
		if err != nil {
			beego.Error("MinesStatPlayers error: ", err)
		}
		c.Data["list"] = stats
		count = total
	} else if typeId == 0 { // 汇总
		dates, err := service.GameStatsService.MinesStatDates(params)
		if err != nil {
			beego.Error("MinesStatDates error: ", err)
		}
		c.Data["list"] = dates
		count = len(dates)
	}

	usertype := map[int]string{
		0:  "全部",
		1:  "新手",
		3:  "平民",
		4:  "泡沫",
		5:  "零充",
		6:  "普充",
		7:  "小R",
		8:  "中R",
		9:  "大R",
		10: "超大R",
		// 11: "手动",
	}
	packageList, _ := service.ChannelService.GetChannelAll()
	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()
	c.Data["count"] = count
	if typeId != 0 {
		c.Data["pageBar"] = libs.NewPager(page, int(count), c.pageSize, beego.URLFor("GameStatsController.Mines", "startDate", startDate, "endDate", endDate, "status", status, "packageId", packageId, "typeId", typeId, "aliasId", aliasId, "aliasName", aliasName, "usertypeId", utypeId, "utype_id", utypeId, "tabId", tabId, "userid", userid, "lossDays1", lossDays1, "lossDays2", lossDays2, "bets1", bets1, "bets2", bets2, "betAvg1", betAvg1, "betAvg2", betAvg2, "cash1", cash1, "cash2", cash2), true).ToString()
	}
	c.Data["usertype"] = usertype
	c.Data["usertypeId"] = utypeId
	c.Data["startDate"] = startDate
	c.Data["endDate"] = endDate
	c.Data["packageId"] = packageId
	c.Data["packageList"] = packageList
	c.Data["typeId"] = typeId
	c.Data["aliasId"] = aliasId
	c.Data["classId"] = classId
	c.Data["packageAliasList"] = packageAliasList
	c.Data["aliasName"] = aliasName
	c.Data["status"] = status
	c.Data["isOperation"] = c.auth.HasAccessPerm(c.controllerName, "crash")
	c.Data["tabId"] = tabId
	c.Data["page"] = page
	c.Data["userid"] = userid
	c.Data["lossDays1"] = lossDays1
	c.Data["lossDays2"] = lossDays2
	c.Data["bets1"] = bets1
	c.Data["bets2"] = bets2
	c.Data["betAvg1"] = betAvg1
	c.Data["betAvg2"] = betAvg2
	c.Data["cash1"] = cash1
	c.Data["cash2"] = cash2
	c.Data["pageTitle"] = "Mines"
	c.display()
}
