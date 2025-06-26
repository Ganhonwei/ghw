package controllers

import (
	"errors"
	"fmt"
	"goserver/internal/web/admin/app/entity"
	"goserver/internal/web/admin/app/libs"
	"goserver/internal/web/admin/app/service"
	"goserver/pkg/utils"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/globalsign/mgo/bson"
)

type Statistics2Controller struct {
	BaseController
}

// Ltv统计
func (c *Statistics2Controller) LtvStat2() {
	status, _ := c.GetInt("status")
	page, _ := strconv.Atoi(c.GetString("page"))
	startDate := c.GetString("start_date")
	endDate := c.GetString("end_date")
	packageId := c.GetString("package_id")
	aliasId := c.GetString("alias_id")
	aliasName := c.GetString("alias_name")
	tabId, _ := c.GetInt("tab_id")
	typeId, _ := c.GetInt("typeId")
	chatid, _ := c.GetInt("chat_id")
	if page < 1 {
		page = 1
	}
	var start, end time.Time
	if startDate == "" || endDate == "" {
		// 默认近十天数据
		today := bson.Now()
		if typeId == 1 {
			if startDate == "" && endDate != "" {
				// 从后往前推
				end, _ = time.Parse("2006-01-02", endDate)
				start = end.AddDate(0, 0, -9)
				// start, end = today.AddDate(0, 0, -9), today
			} else if startDate != "" && endDate == "" {
				// 从前往后推
				start, _ = time.Parse("2006-01-02", startDate)
				end = start.AddDate(0, 0, 9)
			} else {
				start, end = today.AddDate(0, 0, -9), today
			}
		} else {
			start, end = today.AddDate(0, 0, -9), today
		}

	} else {
		start, _ = time.Parse("2006-01-02", startDate)
		end, _ = time.Parse("2006-01-02", endDate)
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
	m := bson.M{}
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
			m["channel"] = bson.M{"$in": temp_arr}
		} else {
			m["channel"] = packageId
		}
		// 存储缓存
		c.SetSession("my_select_pakeageid", packageId)
	} else {
		// 清除缓存
		c.DelSession("my_select_pakeageid")
	}
	// 渠道别名
	if aliasId != "" && aliasId != "0" && aliasId != "-" && aliasName == "" {
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
	} else {
		if aliasName != "" {
			channels, err := service.ChannelService.GetChannelByNameAlias1(aliasName)
			if err != nil {
				beego.Error("channel error1: ", err)
			} else {
				packageIds = make([]string, 0)
				for _, c := range channels {
					packageIds = append(packageIds, c.Name)
				}
			}
		}
	}
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
	if typeId == 1 {
		// 判断日期必须是十天
		diff := end.Sub(start)
		daysDiff := int(diff.Hours() / 24)
		if daysDiff < 10 {
			start = end.AddDate(0, 0, -9)
		} else {
			start = end.AddDate(0, 0, -9)
		}

		fmt.Printf("两个日期相差 %d 天\n", daysDiff)

		// 折线图
		dates := []int{0, 1, 3, 7, 15, 30, 60}
		datetimes := make([]interface{}, 0)
		data := make([]interface{}, 0)  // 1
		data1 := make([]interface{}, 0) // 2
		data2 := make([]interface{}, 0) // 3
		data3 := make([]interface{}, 0) // 4
		data4 := make([]interface{}, 0) // 5
		data5 := make([]interface{}, 0) // 6
		data6 := make([]interface{}, 0) // 7
		data7 := make([]interface{}, 0) // 8
		data8 := make([]interface{}, 0) // 9
		data9 := make([]interface{}, 0) // 10
		list, _ := service.Statistics2Service.GetLtvStat2List(page, -1, m, start, end)
		startDate = fmt.Sprintf("%s", start.Format("2006-01-02"))
		endDate = fmt.Sprintf("%s", end.Format("2006-01-02"))
		if len(list) > 0 {
			count := len(list)
			for i := 0; i < count/2; i++ {
				list[i], list[count-1-i] = list[count-1-i], list[i]
			}
			for idx, item := range list {
				datetimes = append(datetimes, item["date"])
				ltv0 := item["ltv0"].(map[string]string)
				ltv1 := item["ltv1"].(map[string]string)
				ltv3 := item["ltv3"].(map[string]string)
				ltv7 := item["ltv7"].(map[string]string)
				ltv15 := item["ltv15"].(map[string]string)
				ltv30 := item["ltv30"].(map[string]string)
				ltv60 := item["ltv60"].(map[string]string)
				var v0 string
				var v1 string
				var v3 string
				var v7 string
				var v15 string
				var v30 string
				var v60 string
				if chatid == 0 {
					v0 = ltv0["ltvV1"]
					v1 = ltv1["ltvV1"]
					v3 = ltv3["ltvV1"]
					v7 = ltv7["ltvV1"]
					v15 = ltv15["ltvV1"]
					v30 = ltv30["ltvV1"]
					v60 = ltv60["ltvV1"]
				} else if chatid == 1 {
					v0 = ltv0["ltvV2"]
					v1 = ltv1["ltvV2"]
					v3 = ltv3["ltvV2"]
					v7 = ltv7["ltvV2"]
					v15 = ltv15["ltvV2"]
					v30 = ltv30["ltvV2"]
					v60 = ltv60["ltvV2"]
				} else if chatid == 2 {
					v0 = ltv0["ltvPayV1"]
					v1 = ltv1["ltvPayV1"]
					v3 = ltv3["ltvPayV1"]
					v7 = ltv7["ltvPayV1"]
					v15 = ltv15["ltvPayV1"]
					v30 = ltv30["ltvPayV1"]
					v60 = ltv60["ltvPayV1"]

				} else if chatid == 3 {
					v0 = ltv0["ltvPayV2"]
					v1 = ltv1["ltvPayV2"]
					v3 = ltv3["ltvPayV2"]
					v7 = ltv7["ltvPayV2"]
					v15 = ltv15["ltvPayV2"]
					v30 = ltv30["ltvPayV2"]
					v60 = ltv60["ltvPayV2"]

				}
				switch idx {
				case 0:
					data = append(data, v0)
					data = append(data, v1)
					data = append(data, v3)
					data = append(data, v7)
					data = append(data, v15)
					data = append(data, v30)
					data = append(data, v60)
				case 1:
					data1 = append(data1, v0)
					data1 = append(data1, v1)
					data1 = append(data1, v3)
					data1 = append(data1, v7)
					data1 = append(data1, v15)
					data1 = append(data1, v30)
					data1 = append(data1, v60)
				case 2:
					data2 = append(data2, v0)
					data2 = append(data2, v1)
					data2 = append(data2, v3)
					data2 = append(data2, v7)
					data2 = append(data2, v15)
					data2 = append(data2, v30)
					data2 = append(data2, v60)
				case 3:
					data3 = append(data3, v0)
					data3 = append(data3, v1)
					data3 = append(data3, v3)
					data3 = append(data3, v7)
					data3 = append(data3, v15)
					data3 = append(data3, v30)
					data3 = append(data3, v60)
				case 4:
					data4 = append(data4, v0)
					data4 = append(data4, v1)
					data4 = append(data4, v3)
					data4 = append(data4, v7)
					data4 = append(data4, v15)
					data4 = append(data4, v30)
					data4 = append(data4, v60)
				case 5:
					data5 = append(data5, v0)
					data5 = append(data5, v1)
					data5 = append(data5, v3)
					data5 = append(data5, v7)
					data5 = append(data5, v15)
					data5 = append(data5, v30)
					data5 = append(data5, v60)
				case 6:
					data6 = append(data6, v0)
					data6 = append(data6, v1)
					data6 = append(data6, v3)
					data6 = append(data6, v7)
					data6 = append(data6, v15)
					data6 = append(data6, v30)
					data6 = append(data6, v60)
				case 7:
					data7 = append(data7, v0)
					data7 = append(data7, v1)
					data7 = append(data7, v3)
					data7 = append(data7, v7)
					data7 = append(data7, v15)
					data7 = append(data7, v30)
					data7 = append(data7, v60)
				case 8:
					data8 = append(data8, v0)
					data8 = append(data8, v1)
					data8 = append(data8, v3)
					data8 = append(data8, v7)
					data8 = append(data8, v15)
					data8 = append(data8, v30)
					data8 = append(data8, v60)
				case 9:
					data9 = append(data9, v0)
					data9 = append(data9, v1)
					data9 = append(data9, v3)
					data9 = append(data9, v7)
					data9 = append(data9, v15)
					data9 = append(data9, v30)
					data9 = append(data9, v60)
				}
			}
		}

		c.Data["TimedLabel"] = dates
		c.Data["Data1"] = data
		c.Data["Data2"] = data1
		c.Data["Data3"] = data2
		c.Data["Data4"] = data3
		c.Data["Data5"] = data4
		c.Data["Data6"] = data5
		c.Data["Data7"] = data6
		c.Data["Data8"] = data7
		c.Data["Data9"] = data8
		c.Data["Data10"] = data9
		c.Data["DateTimes"] = datetimes
	} else {
		// 表格图
		list, _ := service.Statistics2Service.GetLtvStat2List(page, c.pageSize, m, start, end)
		c.Data["list"] = list
		count := len(list)
		c.Data["count"] = count
		// 表格：倒序显示
		for i := 0; i < count/2; i++ {
			list[i], list[count-1-i] = list[count-1-i], list[i]
		}
	}

	slice := make([]struct{}, 7)
	c.Data["Slice"] = slice
	packageList, _ := service.ChannelService.GetChannelAll()
	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()
	c.Data["pageTitle"] = "LTV统计"
	chatType := map[int]string{
		0: "LTV",
		1: "净值LTV",
		2: "付费用户LTV",
		3: "付费用户净值",
	}
	c.Data["chatType"] = chatType
	c.Data["chatid"] = chatid
	// this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.LtvStat", "typeId", typeId, "package_id", packageId, "alias_id", aliasId, "alias_name", aliasName, "start_date", startDate, "end_date", endDate), true).ToString()
	c.Data["startDate"] = startDate
	c.Data["endDate"] = endDate
	c.Data["packageId"] = packageId
	c.Data["packageList"] = packageList
	c.Data["aliasId"] = aliasId
	c.Data["packageAliasList"] = packageAliasList
	c.Data["aliasName"] = aliasName
	c.Data["status"] = status
	c.Data["isOperation"] = c.auth.HasAccessPerm(c.controllerName, "ltvstat2")
	c.Data["tabId"] = tabId
	c.Data["typeId"] = typeId
	c.display()
}

// 玩家行为分析
func (c *Statistics2Controller) BehaviorAnalysis() {
	startDate := c.GetString("start_date")
	endDate := c.GetString("end_date")
	utype_id, _ := c.GetInt("utype_id")
	range1, _ := c.GetInt("range1")
	limit1, _ := c.GetInt("limit1")
	charges1, _ := c.GetInt("charges1")
	range2, _ := c.GetInt("range2")
	limit2, _ := c.GetInt("limit2")
	charges2, _ := c.GetInt("charges2")
	charges3, _ := c.GetInt("charges3")
	fmt.Println(utype_id)

	m := service.FindByDate(startDate, endDate, "ctime", "ctime")
	startTime := utils.Str2Time(fmt.Sprintf("%s 00:00:00", startDate), service.Location())

	if utype_id != 0 {
		startMoney := 0
		endMoney := 0
		switch utype_id {
		case 1:
			// 零充
			startMoney, endMoney = 0, 0
		case 2:
			// 小R
			startMoney, endMoney = 100000, 499999
		case 3:
			// 中R
			startMoney, endMoney = 500000, 999999
		case 4:
			// 大R
			startMoney, endMoney = 1000000, 9999999
		case 5:
			// 超大R
			startMoney, endMoney = 10000000, -1
		}
		if endMoney == -1 {
			m["money"] = bson.M{"$gte": startMoney}
		} else {
			m["money"] = bson.M{"$gte": startMoney, "$lte": endMoney}
		}
	}

	if startDate != "" {
		// pChargeTimes := 10
		// var range1, limit1 = 20, 100

		analysis1, analysis1Users, analysis2, analysis2Users, analysis3, err := service.Statistics2Service.BehaviorAnalysis(m, startTime, charges1, range1, limit1, charges2, range2, limit2, charges3)
		if err != nil {
			beego.Error("BehaviorAnalysis error", err)
		}

		if len(analysis1) > 0 {
			var ranges1Key []string // columns1
			ranges1Key0, ranges1KeyLimit := "0", fmt.Sprintf("%d+", limit1+1)
			ranges1Key = append(ranges1Key, ranges1Key0)
			for i := 0; i < limit1; i += range1 {
				rr := [2]int{i + 1, i + range1}
				ranges1Key = append(ranges1Key, fmt.Sprintf("%d-%d", rr[0], rr[1]))
			}
			ranges1Key = append(ranges1Key, ranges1KeyLimit)

			var rows1 [][]map[string]any // rows1
			// var rows1Users [][]string
			for chargeTimes := 0; chargeTimes <= charges1; chargeTimes++ {
				var row1 []map[string]any
				// var row1Users []string
				row1 = append(row1, map[string]any{
					"val": fmt.Sprintf("%d充人数", chargeTimes),
				})
				roundPlayers, ok := analysis1[chargeTimes]
				if !ok {
					continue
				}
				roundPlayerIds := analysis1Users[chargeTimes]

				for _, key := range ranges1Key {
					row1 = append(row1, map[string]any{
						"val":   roundPlayers[key],
						"users": utils.Join(roundPlayerIds[key], ","),
					})
					// row1Users = append(row1Users, utils.Join(roundPlayerIds[key], ","))
				}
				rows1 = append(rows1, row1)
				// rows1Users = append(rows1Users, row1Users)
			}
			c.Data["columns1"] = append([]string{"局数"}, ranges1Key...)
			c.Data["rows1"] = rows1
			// c.Data["rows1Users"] = rows1Users
		}

		if len(analysis2) > 0 {
			var ranges2Key []string // columns1
			ranges2Key0, ranges2KeyLimit := "0", fmt.Sprintf("%d+", limit2+1)
			ranges2Key = append(ranges2Key, ranges2Key0)
			for i := 0; i < limit2; i += range2 {
				rr := [2]int{i + 1, i + range2}
				ranges2Key = append(ranges2Key, fmt.Sprintf("%d-%d", rr[0], rr[1]))
			}
			ranges2Key = append(ranges2Key, ranges2KeyLimit)

			var rows2 [][]map[string]any
			for chargeTimes := 0; chargeTimes <= charges2; chargeTimes++ {
				var row2 []map[string]any
				row2 = append(row2, map[string]any{
					"val": fmt.Sprintf("%d充人数", chargeTimes),
				})
				roundPlayers, ok := analysis2[chargeTimes]
				if !ok {
					continue
				}
				roundPlayerIds := analysis2Users[chargeTimes]
				for _, key := range ranges2Key {
					row2 = append(row2, map[string]any{
						"val":   roundPlayers[key],
						"users": utils.Join(roundPlayerIds[key], ","),
					})
				}
				rows2 = append(rows2, row2)
			}
			c.Data["columns2"] = append([]string{"局数"}, ranges2Key...)
			c.Data["rows2"] = rows2
		}

		if len(analysis3) > 0 {
			gcount := 10
			gtypeColumn := map[int]int{
				1:  1,
				2:  2,
				3:  3,
				4:  4,
				5:  5,
				6:  6,
				7:  7,
				8:  8,
				9:  9,
				10: 10,
				// 11: 11,
			}
			columns3 := []string{"游戏", "TP", "LHD", "SEVEN", "RUMMY", "AK47", "JOKER", "CRASH", "ABAR", "LOTTERY", "PLANE"} //, "REDBLACK"}

			var rows3 [][]any
			for chargeTimes := 0; chargeTimes <= charges3; chargeTimes++ {
				var row3 = make([]any, len(columns3))
				row3[0] = fmt.Sprintf("%d次充值人打码/局", chargeTimes)
				gtypeBets, ok := analysis3[chargeTimes]
				if !ok {
					continue
				}
				for gtype, bets := range gtypeBets {
					column, ok := gtypeColumn[gtype]
					if !ok { // 外接游戏...
						gcount++
						column = gcount
						gtypeColumn[gtype] = gcount
						gname, ok := service.GtypeNameMap[gtype]
						if !ok {
							gname = strconv.Itoa(gtype)
						}
						columns3 = append(columns3, gname)
						row3 = append(row3, "0")
					}
					var bet = "0"
					if bets[1] > 0 {
						bet = fmt.Sprintf("%.2f", float64(bets[0])/100.0/float64(bets[1]))
					}
					row3[column] = bet
				}
				rows3 = append(rows3, row3)
			}
			c.Data["columns3"] = columns3
			c.Data["rows3"] = rows3
		}
	}

	c.Data["pageTitle"] = "玩家行为分析"
	c.Data["startDate"] = startDate
	c.Data["endDate"] = endDate
	c.Data["usertypeId"] = utype_id
	c.Data["range1"] = range1
	c.Data["limit1"] = limit1
	c.Data["charges1"] = charges1
	c.Data["range2"] = range2
	c.Data["limit2"] = limit2
	c.Data["charges2"] = charges2
	c.Data["charges3"] = charges3
	c.Data["isOperation"] = c.auth.HasAccessPerm(c.controllerName, "behavioranalysis")
	c.display()
}

// 子项目收益
func (c *Statistics2Controller) SubprojectIncome() {
	list, _ := service.Statistics2Service.SubprojectIncome()
	count := len(list)
	c.Data["pageTitle"] = "子项目收益"
	c.Data["list"] = list
	c.Data["count"] = count
	c.Data["isOperation"] = c.auth.HasAccessPerm(c.controllerName, "subprojectincomeopt")
	c.display()
}

// 刷新子项目收益
func (c *Statistics2Controller) RefreshIncome() {
	go func() {
		// 执行数据计算的代码
		err := service.Statistics2Service.RefreshIncome()
		if err != nil {
			c.checkError(err)
		}
	}()
	// 返回一个立即响应，告知前端请求已收到
	c.showMsg("刷新成功，后台计算中...", MSG_OK, beego.URLFor("Statistics2Controller.SubprojectIncome"))
	c.display()
}

// 重置子项目收益
func (c *Statistics2Controller) ResetIncome() {
	// "id" $v.GameName "rid" $v.RoomName
	id := c.GetString("id")
	roomId := c.GetString("rid")
	err := service.Statistics2Service.ResetIncome(id, roomId)
	if err != nil {
		c.checkError(err)
	}
	c.redirect(beego.URLFor("Statistics2Controller.SubprojectIncome"))
	c.display()
}

// 游戏状态
func (c *Statistics2Controller) GameState() {
	list, err := service.Statistics2Service.GetGameState()
	if err != nil {
		c.checkError(err)
	}
	count := len(list)
	c.Data["pageTitle"] = "游戏状态"
	c.Data["list"] = list
	c.Data["count"] = count
	c.display()
}

// 排行榜
func (c *Statistics2Controller) RankingList() {
	startDate := c.GetString("start_date")
	if startDate == "" {
		today := bson.Now()
		startDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
	}
	list, RefreshTime, _ := service.Statistics2Service.GetRankingList(startDate)
	var paylist []entity.RankingList
	var withlist []entity.RankingList
	var winlist []entity.RankingList
	var loselist []entity.RankingList
	var paycarrylist []entity.RankingList
	var fzcarrylist []entity.RankingList
	RefreshTime, _ = service.ConvertToIndiaTime(RefreshTime.Unix())
	if len(list) > 0 {
		paylist = list[0]
		withlist = list[1]
		winlist = list[2]
		loselist = list[3]
		paycarrylist = list[4]
		fzcarrylist = list[5]
	}
	c.Data["pageTitle"] = "排行榜"
	c.Data["startDate"] = startDate
	c.Data["paylist"] = paylist
	c.Data["withdrawlist"] = withlist
	c.Data["winlist"] = winlist
	c.Data["loselist"] = loselist
	c.Data["payxdlist"] = paycarrylist
	c.Data["fzxdlist"] = fzcarrylist
	c.Data["RefreshTime"] = RefreshTime
	c.display()
}

// 玩家数据
func (c *Statistics2Controller) PlayerData() {
	userid := c.GetString("userid")
	page, _ := strconv.Atoi(c.GetString("page"))
	if page < 1 {
		page = 1
	}
	count := 0
	GameCount := 0
	LifecycleCount := 0
	ControlsCount := 0
	if userid != "" {
		list, _ := service.Statistics2Service.GetPlayerData(page, c.pageSize, userid)
		c.Data["Players"] = list.Players
		c.Data["Lifecycle"] = list.Lifecycle
		c.Data["Games"] = list.Games
		c.Data["Controls"] = list.Controls
		count = len(list.Players)
		LifecycleCount = list.LifecycleCount
		GameCount = len(list.Games)
		ControlsCount = list.ControlCount
	}
	c.Data["pageTitle"] = "玩家数据"
	c.Data["userid"] = userid
	c.Data["count"] = count
	c.Data["GamesCount"] = GameCount
	c.Data["LifecycleCount"] = LifecycleCount
	c.Data["ControlsCount"] = ControlsCount
	c.Data["pageBar"] = libs.NewPager(page, int(LifecycleCount), c.pageSize, beego.URLFor("Statistics2Controller.PlayerData", "userid", userid), true).ToString()
	c.display()
}

// tp剧情局盘口
func (c *Statistics2Controller) TpStoryStock() {
	tabId, _ := c.GetInt("tabId")
	id := c.GetString("id")
	typeId, _ := c.GetInt("type_id")
	page, _ := strconv.Atoi(c.GetString("page"))
	startDate := c.GetString("start_date")
	endDate := c.GetString("end_date")

	var list any
	var count int
	if tabId == 0 {
		stocks, err := service.Statistics2Service.GetTpStoryStockList()
		if err != nil {
			beego.Error("GetTpStoryStock error: ", err)
		} else {
			list = stocks
			count = len(stocks)
		}
	} else if tabId == 1 {
		// 默认当天数据
		if startDate == "" && endDate == "" {
			if id == "" {
				today := bson.Now()
				startDate = today.Format("2006-01-02")
				endDate = today.Format("2006-01-02")
			}
		}
		m := service.FindByDate1(startDate, endDate, "timestamp", "timestamp")
		if id != "" {
			switch typeId {
			case 0:
				m["userid"] = id
			case 1:
				m["detailid"] = id
			}
		}
		logs, logCount, err := service.Statistics2Service.GetTpStoryStockHistoryList(m, page, c.pageSize)
		if err != nil {
			beego.Error("GetTpStoryStock error: ", err)
		} else {
			list = logs
			count = logCount
			c.Data["pageBar"] = libs.NewPager(page, int(count), c.pageSize, beego.URLFor("Statistics2Controller.TpStoryStock", "tabId", tabId, "id", id, "typeId", typeId, "start_date", startDate, "end_date", endDate), true).ToString()
		}
	}
	c.Data["pageTitle"] = "剧情局盘口"
	c.Data["tabId"] = tabId
	c.Data["typeId"] = typeId
	c.Data["id"] = id
	c.Data["startDate"] = startDate
	c.Data["endDate"] = endDate
	c.Data["count"] = count
	c.Data["list"] = list
	c.display()
}

// 修改库存
func (c *Statistics2Controller) TpStoryStockEdit() {
	id := c.GetString("id")
	inputStock := c.GetString("new_stock")
	remark := c.GetString("remark")
	title := ""
	if id == "" {
		c.checkError(fmt.Errorf("房间ID不能为空"))
	}
	if c.isPost() {
		if inputStock == "" {
			c.checkError(fmt.Errorf("库存值不能为空"))
		}
		newStock, err := strconv.Atoi(inputStock)
		if err != nil {
			c.checkError(fmt.Errorf("库存值只能是数字"))

		}
		if remark == "" {
			c.checkError(errors.New("请输入备注！"))
		}
		err = service.Statistics2Service.ModifyTpStoryStock(id, int64(newStock))
		if err != nil {
			c.checkError(errors.New(err.Error()))
		}

		service.ActionService.Add("tp_story_stock_edit", c.auth.GetUser().UserName, "", id, fmt.Sprint(newStock), remark)

		c.redirect(beego.URLFor("Statistics2Controller.TpStoryStock"))
	}

	title = "修改库存"
	c.Data["pageTitle"] = title
	c.Data["id"] = id
	c.display()
}

// 日数据概览
func (this *Statistics2Controller) DayDataList() {
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	aliasId := this.GetString("alias_id")
	classId := this.GetString("class_id")
	registArea, _ := this.GetInt("registArea")

	today := service.NowTime()
	var stime, etime time.Time
	if endDate == "" {
		etime = today
		endDate = etime.Format("2006-01-02")
	}
	if startDate == "" {
		// 默认近半个月数据
		stime = etime.AddDate(0, 0, -14)
		startDate = stime.Format("2006-01-02")
	}
	s, e := service.FindByDate4(startDate, endDate)
	if s == nil || e == nil {
		beego.Error("date range error", startDate, endDate)
		this.showMsg(fmt.Sprintf("date range error: %s, %s", startDate, endDate), MSG_ERR)
	}
	stime, etime = *s, *e

	var registAreas []int32
	if registArea > 0 {
		registAreas = append(registAreas, int32(registArea-1))
	}

	// 渠道别名
	var packageIds []string
	var aliasIds []string
	if aliasId != "" && aliasId != "0" && aliasId != "-" {
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
	var classIds []string
	if classId != "" && classId != "0" && classId != "-" {
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

	list, err := service.Statistics2Service.DayDataList(stime, etime, registAreas, packageIds)
	if err != nil {
		beego.Error(err)
		this.showMsg(err.Error(), MSG_ERR)
	}
	this.Data["list"] = list
	this.Data["count"] = len(list)

	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()
	packageClassAll, _ := service.ChannelService.GetChannelClassAll()

	this.Data["pageTitle"] = "日数据概览"
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["aliasId"] = aliasId
	this.Data["classId"] = classId
	this.Data["registArea"] = registArea
	this.Data["packageAliasList"] = packageAliasList
	this.Data["packageClassAll"] = packageClassAll
	this.Data["isExport"] = this.auth.HasAccessPerm(this.controllerName, "daydatalistexport")
	this.display()
}

// 日数据概览-导出
func (c *Statistics2Controller) DayDataListExport() {
	startDate := c.GetString("start_date")
	endDate := c.GetString("end_date")
	aliasId := c.GetString("alias_id")
	classId := c.GetString("class_id")
	registArea, _ := c.GetInt("registArea")

	today := service.NowTime()
	var stime, etime time.Time
	if endDate == "" {
		etime = today
		endDate = etime.Format("2006-01-02")
	}
	if startDate == "" {
		// 默认近半个月数据
		stime = etime.AddDate(0, 0, -14)
		startDate = stime.Format("2006-01-02")
	}
	s, e := service.FindByDate4(startDate, endDate)
	if s == nil || e == nil {
		beego.Error("date range error", startDate, endDate)
		c.showMsg(fmt.Sprintf("date range error: %s, %s", startDate, endDate), MSG_ERR)
	}
	stime, etime = *s, *e

	var registAreas []int32
	if registArea > 0 {
		registAreas = append(registAreas, int32(registArea-1))
	}

	// 渠道别名
	var packageIds []string
	var aliasIds []string
	if aliasId != "" && aliasId != "0" && aliasId != "-" {
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
	var classIds []string
	if classId != "" && classId != "0" && classId != "-" {
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

	list, err := service.Statistics2Service.DayDataList(stime, etime, registAreas, packageIds)
	if err != nil {
		beego.Error(err)
		c.showMsg(err.Error(), MSG_ERR)
	}

	var rows [][]any
	headers := []string{"日期", "渠道类", "渠道别名", "日活", "新注册用户数", "新增设备数", "有效新增率", "老用户日活", "付费用户日活", "总充值金额", "总提现金额", "充-提", "代收手续费", "代付手续费", "通道盈余", "充提盈余率", "通道盈余率", "", "总充值人数", "新用户充值人数", "老用户充值人数", "首充人数", "总提现人数", "新用户提现人数", "老用户提现人数", "首充提现人数", "总付费率", "新用户付费率", "老用户付费率", "首充付费率", "首充付费率(竞)", "首充次日复充率", "总提现率", "新用户提现率", "老用户提现率", "总充值者中提现者占比", "新用户充值者中提现者占比", "老用户充值者中提现者占比", "首充者中提现者占比", "", "新用户充值总金额", "老用户充值总金额", "新用户提现总金额", "老用户提现总金额", "新用户充-提", "老用户充-提", "新用户充提盈余率", "老用户充提盈余率", "总充值ARPU", "新用户ARPU", "老用户ARPU", "总充值ARPPU", "新用户ARPPU", "老用户ARPPU", "", "总当日复购人数+复购率", "新用户当日复购人数+复购率", "老用户当日复购人数+复购率", "首充用户当日复购人数", "总人均付费次数", "新用户人均付费次数", "老用户人均付费次数"}
	for _, v := range list {
		_ = v
		row := []any{
			v.SDate,
			v.ClassId,
			v.AliasId,
			parsePercent(fmt.Sprint(v.LoginUsers)),
			parsePercent(v.RegUsers),
			parsePercent(fmt.Sprint(v.AdRegUsers)),
			parsePercent(v.EffectRegUserRate),
			parsePercent(v.OldLoginUsers),
			parsePercent(fmt.Sprint(v.PayedLoginUsers)),
			parsePercent(v.PayAmounts),
			parsePercent(v.WithdrawAmounts),
			parsePercent(v.PaySubWithdraw),
			parsePercent(v.PayTaxs),
			parsePercent(v.WithdrawTaxs),
			parsePercent(v.ChannelSurplus),
			parsePercent(v.PayWithdrawSurplusRate),
			parsePercent(v.ChannelSurplusRate),
			"",
			parsePercent(fmt.Sprint(v.PayUsers)),
			parsePercent(fmt.Sprint(v.NewPayUsers)),
			parsePercent(fmt.Sprint(v.OldPayUsers)),
			parsePercent(fmt.Sprint(v.FirstPayUsers)),
			parsePercent(fmt.Sprint(v.WithdrawUsers)),
			parsePercent(fmt.Sprint(v.NewWithdrawUsers)),
			parsePercent(fmt.Sprint(v.OldWithdrawUsers)),
			parsePercent(fmt.Sprint(v.FirstPayWithdrawUsers)),
			parsePercent(fmt.Sprint(v.PayRate)),
			parsePercent(fmt.Sprint(v.NewPayRate)),
			parsePercent(fmt.Sprint(v.OldPayRate)),
			parsePercent(fmt.Sprint(v.FirstPayRate)),
			parsePercent(fmt.Sprint(v.FirstPayRate2)),
			parsePercent(fmt.Sprint(v.FirstPayNextDatePayRate)),
			parsePercent(fmt.Sprint(v.WithdrawRate)),
			parsePercent(fmt.Sprint(v.NewWithdrawRate)),
			parsePercent(fmt.Sprint(v.OldWithdrawRate)),
			parsePercent(fmt.Sprint(v.PayWithdrawRate)),
			parsePercent(fmt.Sprint(v.NewPayWithdrawRate)),
			parsePercent(fmt.Sprint(v.OldPayWithdrawRate)),
			parsePercent(fmt.Sprint(v.FirstPayWithdrawRate)),
			"",
			parsePercent(fmt.Sprint(v.NewPayAmounts)),
			parsePercent(fmt.Sprint(v.OldPayAmounts)),
			parsePercent(fmt.Sprint(v.NewWithdrawAmounts)),
			parsePercent(fmt.Sprint(v.OldWithdrawAmounts)),
			parsePercent(fmt.Sprint(v.NewPaySubWithdraw)),
			parsePercent(fmt.Sprint(v.OldPaySubWithdraw)),
			parsePercent(fmt.Sprint(v.NewSurplusRate)),
			parsePercent(fmt.Sprint(v.OldSurplusRate)),
			parsePercent(fmt.Sprint(v.ARPU)),
			parsePercent(fmt.Sprint(v.NewARPU)),
			parsePercent(fmt.Sprint(v.OldARPU)),
			parsePercent(fmt.Sprint(v.ARPPU)),
			parsePercent(fmt.Sprint(v.NewARPPU)),
			parsePercent(fmt.Sprint(v.OldARPPU)),
			"",
			parsePercent(fmt.Sprint(v.Pay2TimesUsers)),
			parsePercent(fmt.Sprint(v.NewPay2TimesUsers)),
			parsePercent(fmt.Sprint(v.OldPay2TimesUsers)),
			parsePercent(fmt.Sprint(v.FirstPayDatePay2Users)),
			parsePercent(fmt.Sprint(v.PayOrderAvg)),
			parsePercent(fmt.Sprint(v.NewPayOrderAvg)),
			parsePercent(fmt.Sprint(v.OldPayOrderAvg)),
		}
		rows = append(rows, row)
	}

	// 检查是否有数据可导出
	if len(rows) == 0 {
		c.showMsg("未查询到可以导出的数据！", MSG_ERR, "")
		return
	}

	file := excelize.NewFile()
	sheet := "Sheet1"

	//设置表格头
	for i, head := range headers {
		no := NumberToExcelColumn(i + 1)
		file.SetCellValue(sheet, fmt.Sprintf("%s1", no), head)
		if head == "" {
			file.MergeCell("Sheet1", fmt.Sprintf("%s1", no), fmt.Sprintf("%s%d", no, len(rows)+1))
		}
	}
	// 写入数据
	for rno, row := range rows {
		for cno, value := range row {
			column := NumberToExcelColumn(cno + 1)
			// Excel 行号从 1 开始，因此要加 2
			file.SetCellValue(sheet, fmt.Sprintf("%s%d", column, rno+2), value)
		}
	}

	//构造文件名称
	title := "日数据概览"
	fileName := title + "_" + time.Now().Format("20060102150405") + ".xlsx"
	// 解决文件名中文乱码问题
	fileName = path.Base(fileName)
	fileName = url.QueryEscape(fileName)
	c.Ctx.Output.Header("Content-type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Ctx.Output.Header("Content-Disposition", "attachment;filename="+fileName)
	c.Ctx.Output.Header("Pragma", "No-cache")
	c.Ctx.Output.Header("Cache-Control", "No-cache")
	c.Ctx.Output.Header("Expires", "0")
	// var buffer bytes.Buffer
	if err := file.Write(c.Ctx.ResponseWriter); err != nil {
		logs.Error(err)
	}
}

// parsePercent 解析百分比
// 12(20.2%) -> 12
// 15.5 -> 15.5
func parsePercent(v string) string {
	var r []rune
	for _, chat := range v {
		if !((chat >= rune('0') && chat <= rune('9')) || chat == rune('.') || chat == rune('-') || chat == rune('%')) {
			break
		}
		r = append(r, chat)
	}
	return string(r)
}

// 日数据概览
func (this *Statistics2Controller) FinanceStats() {
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	aliasId := this.GetString("alias_id")
	classId := this.GetString("class_id")
	registArea, _ := this.GetInt("registArea")

	today := service.NowTime()
	var stime, etime time.Time
	if endDate == "" {
		etime = today
		endDate = etime.Format("2006-01-02")
	}
	if startDate == "" {
		// 默认近半个月数据
		stime = etime.AddDate(0, 0, -14)
		startDate = stime.Format("2006-01-02")
	}
	s, e := service.FindByDate4(startDate, endDate)
	if s == nil || e == nil {
		beego.Error("date range error", startDate, endDate)
		this.showMsg(fmt.Sprintf("date range error: %s, %s", startDate, endDate), MSG_ERR)
	}
	stime, etime = *s, *e

	var registAreas []int32
	if registArea > 0 {
		registAreas = append(registAreas, int32(registArea-1))
	}

	// 渠道别名
	var packageIds []string
	var aliasIds []string
	if aliasId != "" && aliasId != "0" && aliasId != "-" {
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
	var classIds []string
	if classId != "" && classId != "0" && classId != "-" {
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

	list, err := service.Statistics2Service.FinanceStats(stime, etime, registAreas, packageIds)
	if err != nil {
		beego.Error(err)
		this.showMsg(err.Error(), MSG_ERR)
	}
	this.Data["list"] = list
	this.Data["count"] = len(list)

	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()
	packageClassAll, _ := service.ChannelService.GetChannelClassAll()

	this.Data["pageTitle"] = "经济日报"
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["aliasId"] = aliasId
	this.Data["classId"] = classId
	this.Data["registArea"] = registArea
	this.Data["packageAliasList"] = packageAliasList
	this.Data["packageClassAll"] = packageClassAll
	this.Data["isExport"] = this.auth.HasAccessPerm(this.controllerName, "financestatsexport")
	this.display()
}

// 经济日报-导出
func (c *Statistics2Controller) FinanceStatsExport() {
	startDate := c.GetString("start_date")
	endDate := c.GetString("end_date")
	aliasId := c.GetString("alias_id")
	classId := c.GetString("class_id")
	registArea, _ := c.GetInt("registArea")

	today := service.NowTime()
	var stime, etime time.Time
	if endDate == "" {
		etime = today
		endDate = etime.Format("2006-01-02")
	}
	if startDate == "" {
		// 默认近半个月数据
		stime = etime.AddDate(0, 0, -14)
		startDate = stime.Format("2006-01-02")
	}
	s, e := service.FindByDate4(startDate, endDate)
	if s == nil || e == nil {
		beego.Error("date range error", startDate, endDate)
		c.showMsg(fmt.Sprintf("date range error: %s, %s", startDate, endDate), MSG_ERR)
	}
	stime, etime = *s, *e

	var registAreas []int32
	if registArea > 0 {
		registAreas = append(registAreas, int32(registArea-1))
	}

	// 渠道别名
	var packageIds []string
	var aliasIds []string
	if aliasId != "" && aliasId != "0" && aliasId != "-" {
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
	var classIds []string
	if classId != "" && classId != "0" && classId != "-" {
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

	list, err := service.Statistics2Service.FinanceStats(stime, etime, registAreas, packageIds)
	if err != nil {
		beego.Error(err)
		c.showMsg(err.Error(), MSG_ERR)
	}

	var rows [][]any
	type head struct {
		Name    string
		Colspan int
		Tag     int
	}
	headers0 := []head{{}, {},
		{"钱包余额", 8, 0},
		{Tag: 1},
		{"游戏营收", 11, 0},
		{Tag: 1},
		{"VIP bonus发放", 6, 0},
		{"排行榜bonus发放", 8, 0},
		{"代理bonus发放", 8, 0},
		{"转盘bonus发放", 2, 0},
		{"波动返水bonus发放", 2, 0},
		{"礼包码bonus发放", 2, 0},
		{"周卡bonus发放", 2, 0},
		{"首充bonus发放", 2, 0},
		{"二充bonus发放", 2, 0},
		{"三充bonus发放", 2, 0},
		{"普充bonus发放", 2, 0},
		{Tag: 1},
		{"VIP cash发放", 6, 0},
		{"排行榜cash发放", 8, 0},
		{"代理cash发放", 8, 0},
		{"转盘cash发放", 2, 0},
		{"波动返水cash发放", 2, 0},
		{"礼包码cash发放", 2, 0},
		{"周卡cash发放", 2, 0},
		{"首充cash发放", 2, 0},
		{"二充cash发放", 2, 0},
		{"三充cash发放", 2, 0},
		{"普充cash发放", 2, 0},
	}
	headers := []string{"日期", "日活", "钱包总余额", "可提现余额", "不可提现余额", "可提现占总余额比", "日活钱包余额", "日活可提现余额", "日活不可提现余额", "日活可提现占总余额比", "", "游戏收入", "bonus发放总额", "bonus发放占收入比", "cash流入总额", "cash发放总额", "cash发放占收入比", "bonus转cash总额", "bonus转cash占收入比", "充-提收入", "游戏收入-cash流入总额", "盈余误差", "", "VIP发放总额", "VIP占比", "VIP升级奖励", "升级奖励占比", "VIP周奖励", "周奖励占比", "排行榜发放总额", "排行榜占比", "日榜", "日榜占比", "周榜", "周榜占比", "月榜", "月榜占比", "代理发放总额", "代理占比", "下注返佣总额", "下注返佣占比", "人头奖总额", "人头奖占比", "里程碑奖总额", "里程碑奖占比", "转盘发放总额", "转盘占比", "波动返水总额", "返水占比", "礼包码发放总额", "礼包码占比", "周卡发放总额", "周卡占比", "首充发放总额", "首充占比", "二充发放总额", "二充占比", "三充发放总额", "三充占比", "普充发放总额", "普充占比", "", "VIP发放总额", "VIP占比", "VIP升级奖励", "升级奖励占比", "VIP周奖励", "周奖励占比", "排行榜发放总额", "排行榜占比", "日榜", "日榜占比", "周榜", "周榜占比", "月榜", "月榜占比", "代理发放总额", "代理占比", "下注返佣总额", "下注返佣占比", "人头奖总额", "人头奖占比", "里程碑奖总额", "里程碑奖占比", "转盘发放总额", "转盘占比", "波动返水总额", "返水占比", "礼包码发放总额", "礼包码占比", "周卡发放总额", "周卡占比", "首充发放总额", "首充占比", "二充发放总额", "二充占比", "三充发放总额", "三充占比", "普充发放总额", "普充占比"}
	for _, v := range list {
		_ = v
		row := []any{
			v.SDate,
			v.LoginUsers,
			v.Cash,
			v.Withdrawable,
			v.NonWithdrawable,
			v.WithdrawableRate,
			v.LiveCash,
			v.LiveWithdrawable,
			v.LiveNonWithdrawable,
			v.LiveWithdrawableRate,
			"",
			v.GameIncome,
			v.BonusGift,
			v.BonusIncomeRate,
			v.CashFlow,
			v.CashGift,
			v.CashGiftIncomeRate,
			v.CashVb,
			v.CashVbIncomeRate,
			v.PaySubWithdraws,
			v.GameIncomeSubCashFlow,
			v.SurplusMiss,
			"",
			v.BonusVip,
			v.BonusVipRate,
			v.BonusVipUpgrade,
			v.BonusVipUpgradeRate,
			v.BonusVipWeek,
			v.BonusVipWeekRate,
			v.BonusBetRank,
			v.BonusBetRankRate,
			v.BonusBetRankDaily,
			v.BonusBetRankDailyRate,
			v.BonusBetRankWeekly,
			v.BonusBetRankWeeklyRate,
			v.BonusBetRankMonthly,
			v.BonusBetRankMonthlyRate,
			v.BonusShareAgent,
			v.BonusShareAgentRate,
			v.BonusShareAgentBets,
			v.BonusShareAgentBetsRate,
			v.BonusShareAgentHeads,
			v.BonusShareAgentHeadsRate,
			v.BonusShareAgentTasks,
			v.BonusShareAgentTasksRate,
			v.BonusTurn,
			v.BonusTurnRate,
			v.BonusSubsidy,
			v.BonusSubsidyRate,
			v.BonusGivePack,
			v.BonusGivePackRate,
			v.BonusWeekCard,
			v.BonusWeekCardRate,
			v.BonusPay1Give,
			v.BonusPay1GiveRate,
			v.BonusPay2Give,
			v.BonusPay2GiveRate,
			v.BonusPay3Give,
			v.BonusPay3GiveRate,
			v.BonusPayGive,
			v.BonusPayGiveRate,
			"",
			v.CashVip,
			v.CashVipRate,
			v.CashVipUpgrade,
			v.CashVipUpgradeRate,
			v.CashVipWeek,
			v.CashVipWeekRate,
			v.CashBetRank,
			v.CashBetRankRate,
			v.CashBetRankDaily,
			v.CashBetRankDailyRate,
			v.CashBetRankWeekly,
			v.CashBetRankWeeklyRate,
			v.CashBetRankMonthly,
			v.CashBetRankMonthlyRate,
			v.CashShareAgent,
			v.CashShareAgentRate,
			v.CashShareAgentBets,
			v.CashShareAgentBetsRate,
			v.CashShareAgentHeads,
			v.CashShareAgentHeadsRate,
			v.CashShareAgentTasks,
			v.CashShareAgentTasksRate,
			v.CashTurn,
			v.CashTurnRate,
			v.CashSubsidy,
			v.CashSubsidyRate,
			v.CashGivePack,
			v.CashGivePackRate,
			v.CashWeekCard,
			v.CashWeekCardRate,
			v.CashPay1Give,
			v.CashPay1GiveRate,
			v.CashPay2Give,
			v.CashPay2GiveRate,
			v.CashPay3Give,
			v.CashPay3GiveRate,
			v.CashPayGive,
			v.CashPayGiveRate,
		}
		rows = append(rows, row)
	}

	// 检查是否有数据可导出
	if len(rows) == 0 {
		c.showMsg("未查询到可以导出的数据！", MSG_ERR, "")
		return
	}

	file := excelize.NewFile()
	sheet := "Sheet1"

	//设置表格头
	styleCenter, err := file.NewStyle(&excelize.Style{
		// Alignment: excelize.,
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	})
	if err != nil {
		c.showMsg(err.Error(), MSG_ERR, "")
		return
	}
	var col int
	for _, head := range headers0 {
		no := NumberToExcelColumn(col + 1)
		file.SetCellValue(sheet, fmt.Sprintf("%s1", no), head.Name)
		if head.Tag == 1 {
			file.MergeCell("Sheet1", fmt.Sprintf("%s1", no), fmt.Sprintf("%s%d", no, len(rows)+2))
			col++
			continue
		}
		if head.Colspan > 0 {
			toNo := NumberToExcelColumn(col + head.Colspan)
			col += head.Colspan
			c1, c2 := fmt.Sprintf("%s1", no), fmt.Sprintf("%s1", toNo)
			file.MergeCell("Sheet1", c1, c2)
			file.SetCellStyle("Sheet1", c1, c2, styleCenter)
			// fmt.Printf("merge: %s to %s \n", fmt.Sprintf("%s1", no), fmt.Sprintf("%s1", toNo))
			continue
		}
		col++
	}
	for i, head := range headers {
		no := NumberToExcelColumn(i + 1)
		file.SetCellValue(sheet, fmt.Sprintf("%s2", no), head)
	}
	// 写入数据
	for rno, row := range rows {
		for cno, value := range row {
			column := NumberToExcelColumn(cno + 1)
			// Excel 行号从 1 开始，因此要加 2
			file.SetCellValue(sheet, fmt.Sprintf("%s%d", column, rno+3), value)
		}
	}

	//构造文件名称
	title := "经济日报"
	fileName := title + "_" + time.Now().Format("20060102150405") + ".xlsx"
	// 解决文件名中文乱码问题
	fileName = path.Base(fileName)
	fileName = url.QueryEscape(fileName)
	c.Ctx.Output.Header("Content-type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Ctx.Output.Header("Content-Disposition", "attachment;filename="+fileName)
	c.Ctx.Output.Header("Pragma", "No-cache")
	c.Ctx.Output.Header("Cache-Control", "No-cache")
	c.Ctx.Output.Header("Expires", "0")
	// var buffer bytes.Buffer
	if err := file.Write(c.Ctx.ResponseWriter); err != nil {
		logs.Error(err)
	}
}

// 游戏行为日报
func (this *Statistics2Controller) GameBetStats() {
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	aliasId := this.GetString("alias_id")
	classId := this.GetString("class_id")
	registArea, _ := this.GetInt("registArea")

	today := service.NowTime()
	var stime, etime time.Time
	if endDate == "" {
		etime = today
		endDate = etime.Format("2006-01-02")
	}
	if startDate == "" {
		// 默认近半个月数据
		stime = etime.AddDate(0, 0, -14)
		startDate = stime.Format("2006-01-02")
	}
	s, e := service.FindByDate4(startDate, endDate)
	if s == nil || e == nil {
		beego.Error("date range error", startDate, endDate)
		this.showMsg(fmt.Sprintf("date range error: %s, %s", startDate, endDate), MSG_ERR)
	}
	stime, etime = *s, *e

	var registAreas []int32
	if registArea > 0 {
		registAreas = append(registAreas, int32(registArea-1))
	}

	// 渠道别名
	var packageIds []string
	var aliasIds []string
	if aliasId != "" && aliasId != "0" && aliasId != "-" {
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
	var classIds []string
	if classId != "" && classId != "0" && classId != "-" {
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

	list, err := service.Statistics2Service.GameBetStats(stime, etime, registAreas, packageIds)
	if err != nil {
		beego.Error(err)
		this.showMsg(err.Error(), MSG_ERR)
	}
	this.Data["list"] = list
	this.Data["count"] = len(list)

	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()
	packageClassAll, _ := service.ChannelService.GetChannelClassAll()

	this.Data["pageTitle"] = "游戏行为日报"
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["aliasId"] = aliasId
	this.Data["classId"] = classId
	this.Data["registArea"] = registArea
	this.Data["packageAliasList"] = packageAliasList
	this.Data["packageClassAll"] = packageClassAll
	this.Data["isExport"] = this.auth.HasAccessPerm(this.controllerName, "gamebetstatsexport")
	this.display()
}

// 游戏行为日报-导出
func (c *Statistics2Controller) GameBetStatsExport() {
	startDate := c.GetString("start_date")
	endDate := c.GetString("end_date")
	aliasId := c.GetString("alias_id")
	classId := c.GetString("class_id")
	registArea, _ := c.GetInt("registArea")

	today := service.NowTime()
	var stime, etime time.Time
	if endDate == "" {
		etime = today
		endDate = etime.Format("2006-01-02")
	}
	if startDate == "" {
		// 默认近半个月数据
		stime = etime.AddDate(0, 0, -14)
		startDate = stime.Format("2006-01-02")
	}
	s, e := service.FindByDate4(startDate, endDate)
	if s == nil || e == nil {
		beego.Error("date range error", startDate, endDate)
		c.showMsg(fmt.Sprintf("date range error: %s, %s", startDate, endDate), MSG_ERR)
	}
	stime, etime = *s, *e

	var registAreas []int32
	if registArea > 0 {
		registAreas = append(registAreas, int32(registArea-1))
	}

	// 渠道别名
	var packageIds []string
	var aliasIds []string
	if aliasId != "" && aliasId != "0" && aliasId != "-" {
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
	var classIds []string
	if classId != "" && classId != "0" && classId != "-" {
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

	list, err := service.Statistics2Service.GameBetStats(stime, etime, registAreas, packageIds)
	if err != nil {
		beego.Error(err)
		c.showMsg(err.Error(), MSG_ERR)
	}

	var rows [][]any
	headers := []string{"日期", "日活", "新用户日活", "老用户日活", "付费用户日活", "总投注人数", "新用户投注人数", "老用户投注人数", "总充投比", "新用户总充投比", "老用户总充投比", "总投注金额", "新用户投注金额", "老用户投注金额", "总人均日投注额", "新用户人均日投注额", "老用户人均日投注额", "总日活投注率", "总付费投注率", "新用户日活投注率", "新用户付费投注率", "老用户日活投注率", "老用户付费投注率", "总返奖率", "新用户返奖率", "老用户返奖率", "总游戏收入", "新用户游戏收入", "老用户游戏收入", "总杀率", "新用户杀率", "老用户杀率"}
	for _, v := range list {
		_ = v
		row := []any{
			v.SDate,
			v.LoginUsers,
			v.NewLoginUsers,
			v.OldLoginUsers,
			v.PayLoginUsers,
			v.BetUsers,
			v.NewBetUsers,
			v.OldBetUsers,
			v.BetPayRate,
			v.NewBetPayRate,
			v.OldBetPayRate,
			v.Bets,
			v.NewBets,
			v.OldBets,
			v.BetsAvg,
			v.NewBetsAvg,
			v.OldBetsAvg,
			v.BetsRate,
			v.PayBetsRate,
			v.NewBetsRate,
			v.NewPayBetsRate,
			v.OldBetsRate,
			v.OldPayBetsRate,
			v.RebateRate,
			v.NewRebateRate,
			v.OldRebateRate,
			v.Income,
			v.NewIncome,
			v.OldIncome,
			v.KillRate,
			v.NewKillRate,
			v.OldKillRate,
		}
		rows = append(rows, row)
	}

	// 检查是否有数据可导出
	if len(rows) == 0 {
		c.showMsg("未查询到可以导出的数据！", MSG_ERR, "")
		return
	}

	file := excelize.NewFile()
	sheet := "Sheet1"

	//设置表格头
	for i, head := range headers {
		no := NumberToExcelColumn(i + 1)
		file.SetCellValue(sheet, fmt.Sprintf("%s1", no), head)
		if head == "" {
			file.MergeCell("Sheet1", fmt.Sprintf("%s1", no), fmt.Sprintf("%s%d", no, len(rows)+1))
		}
	}
	// 写入数据
	for rno, row := range rows {
		for cno, value := range row {
			column := NumberToExcelColumn(cno + 1)
			// Excel 行号从 1 开始，因此要加 2
			file.SetCellValue(sheet, fmt.Sprintf("%s%d", column, rno+2), value)
		}
	}

	//构造文件名称
	title := "游戏行为日报"
	fileName := title + "_" + time.Now().Format("20060102150405") + ".xlsx"
	// 解决文件名中文乱码问题
	fileName = path.Base(fileName)
	fileName = url.QueryEscape(fileName)
	c.Ctx.Output.Header("Content-type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Ctx.Output.Header("Content-Disposition", "attachment;filename="+fileName)
	c.Ctx.Output.Header("Pragma", "No-cache")
	c.Ctx.Output.Header("Cache-Control", "No-cache")
	c.Ctx.Output.Header("Expires", "0")
	// var buffer bytes.Buffer
	if err := file.Write(c.Ctx.ResponseWriter); err != nil {
		logs.Error(err)
	}
}
