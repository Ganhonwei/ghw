package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"goserver/gen/pb"
	"goserver/internal/web/admin/app/args"
	"goserver/internal/web/admin/app/entity"
	"goserver/internal/web/admin/app/libs"
	"goserver/internal/web/admin/app/service"
	"goserver/pkg/utils"
	"log"
	"math"
	"net/url"
	"path"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/globalsign/mgo/bson"
)

type StatisticsController struct {
	BaseController
}

// 数据汇总
func (this *StatisticsController) DataList() {
	status, _ := this.GetInt("status")
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	packageId := this.GetString("package_id")
	aliasId := this.GetString("alias_id")
	aliasName := this.GetString("alias_name")
	tabId, _ := this.GetInt("tabId")
	if page < 1 {
		page = 1
	}
	if startDate == "" && endDate == "" {
		// 默认近十天数据
		today := bson.Now()
		startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -9).Format("2006-01-02"))
		endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
	}
	if status == 0 {
		my_data := this.GetSession("my_select_pakeageid")
		if my_data != nil && packageId == "" {
			if packid, ok := my_data.(string); ok {
				packageId = packid
			}
		}
		status = 1
	}
	m := service.FindByDate1(startDate, endDate, "date", "date")
	if tabId == 1 {
		// m["$or"] = []bson.M{
		// 	{"regist_area": 0},
		// 	{"regist_area": nil},
		// }
		m["data_types"] = bson.M{"$ne": 1}
	} else if tabId == 2 {
		m["data_types"] = 1
	} else if tabId == 3 {
		m["data_types"] = 2
	}
	isChannel := false
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
		isChannel = true
		// 存储缓存
		this.SetSession("my_select_pakeageid", packageId)
	} else {
		// 清除缓存
		this.DelSession("my_select_pakeageid")
	}
	// 渠道别名
	if aliasId != "" && aliasId != "0" {
		if aliasId == "-" {
			m["channel1"] = ""
		} else {
			if strings.Contains(aliasId, ",") {
				palkage_arr := strings.Split(aliasId, ",")
				temp_arr := make([]string, 0)
				for _, item := range palkage_arr {
					if item != "" && item != "0" && item != "-" {
						temp_arr = append(temp_arr, item)
					}
				}
				m["channel1"] = bson.M{"$in": temp_arr}
			} else {
				m["channel1"] = aliasId
			}
			isChannel = true
		}
	}
	if aliasName != "" {
		m["channel1"] = bson.M{"$regex": aliasName, "$options": "i"}
		isChannel = true
	}
	count, _ := service.StatisticsService.GetDataTotal(m, isChannel)
	list, err := service.StatisticsService.GetDataList(page, this.pageSize, m, isChannel)
	if err != nil {
		beego.Error(err)
	}
	this.Data["list"] = list

	packageList, _ := service.ChannelService.GetChannelAll()
	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()

	this.Data["pageTitle"] = "数据汇总"
	this.Data["count"] = count
	this.Data["status"] = status
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.DataList", "status", status, "tabId", tabId, "package_id", packageId, "alias_id", aliasId, "alias_name", aliasName, "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["packageId"] = packageId
	this.Data["packageList"] = packageList
	this.Data["aliasId"] = aliasId
	this.Data["packageAliasList"] = packageAliasList
	this.Data["aliasName"] = aliasName
	this.Data["tabId"] = tabId
	this.display()
}

// 数据汇总（按时间）
func (this *StatisticsController) DtList() {
	status, _ := this.GetInt("status")
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	packageId := this.GetString("package_id")
	aliasId := this.GetString("alias_id")
	aliasName := this.GetString("alias_name")
	if page < 1 {
		page = 1
	}
	if startDate == "" && endDate == "" {
		// 默认一周数据
		today := bson.Now()
		startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -6).Format("2006-01-02"))
		endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
	}
	if status == 0 {
		my_data := this.GetSession("my_select_pakeageid")
		if my_data != nil && packageId == "" {
			if packid, ok := my_data.(string); ok {
				packageId = packid
			}
		}
		status = 1
	}
	channelArr := make([]string, 0)
	aliasArr := make([]string, 0)
	if packageId != "" && packageId != "0" && packageId != "-" {
		if strings.Contains(packageId, ",") {
			palkage_arr := strings.Split(packageId, ",")
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					channelArr = append(channelArr, item)
				}
			}
		} else {
			channelArr = append(channelArr, packageId)
		}
	}
	// 渠道别名
	if aliasId != "" && aliasId != "0" && aliasId != "-" {
		if strings.Contains(aliasId, ",") {
			palkage_arr := strings.Split(aliasId, ",")
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					aliasArr = append(aliasArr, item)
				}
			}
		} else {
			aliasArr = append(aliasArr, aliasId)
		}
	}
	// count, _ := service.StatisticsService.GetDtList(m, isChannel)
	list, _ := service.StatisticsService.GetDtList(channelArr, aliasArr, aliasName, startDate, endDate)
	count := len(list)
	this.Data["list"] = list

	packageList, _ := service.ChannelService.GetChannelAll()
	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()

	this.Data["pageTitle"] = "数据汇总(按时间)"
	this.Data["count"] = count
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.DataList", "status", status, "package_id", packageId, "alias_id", aliasId, "alias_name", aliasName, "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["packageId"] = packageId
	this.Data["packageList"] = packageList
	this.Data["aliasId"] = aliasId
	this.Data["packageAliasList"] = packageAliasList
	this.Data["aliasName"] = aliasName
	this.Data["status"] = status
	this.display()
}

// 数据汇总
func (this *StatisticsController) List() {
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")

	if page < 1 {
		page = 1
	}
	if startDate == "" && endDate == "" {
		// 默认近十天数据
		today := bson.Now()
		startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -9).Format("2006-01-02"))
		endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
	}
	m := service.FindByDate1(startDate, endDate, "date", "date")

	count, _ := service.StatisticsService.GetTotal(m)
	list, _ := service.StatisticsService.GetList(page, this.pageSize, m)

	this.Data["pageTitle"] = "数据汇总(旧)"
	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.List", "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate

	this.display()
}

// 用户留存
func (this *StatisticsController) Retained() {
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	packageId := this.GetString("package_id")
	typeId, _ := this.GetInt("typeId")
	aliasId := this.GetString("alias_id")
	// aliasName := this.GetString("alias_name")
	classId := this.GetString("class_id")
	utypeId := this.GetString("utype_id")

	if page < 1 {
		page = 1
	}

	today := service.NowTime()
	if startDate == "" {
		// 默认近十天数据
		startDate = today.AddDate(0, 0, -9).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = today.Format("2006-01-02")
	}
	s, e := service.FindByDate4(startDate, endDate)

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

	var utypeIds []int
	if utypeId != "" && utypeId != "0" && utypeId != "-" && utypeId != "<nil>" {
		ids := strings.Split(utypeId, ",")
		for _, _id := range ids {
			id, err := strconv.Atoi(_id)
			if err != nil || id == 0 {
				continue
			}
			if id == 10 { // 新手
				id = 0
			}
			utypeIds = append(utypeIds, id)
		}
	} else {
		utypeId = ""
	}

	var count int
	if typeId == 1 {
		// 付费留存
		// count, _ = service.StatisticsService.GetPayUserRetainedTotal(m, isChannel)
		// list, _ := service.StatisticsService.GetPayUserRetainedList(page, this.pageSize, m, isChannel)
		list, err := service.StatisticsService.GetRetainedPay(*s, *e, packageIds)
		if err != nil {
			this.showMsg(err.Error(), MSG_ERR)
		}
		count = len(list)
		this.Data["list"] = list
	} else if typeId == 2 {
		// 复充留存
		s, e := service.FindByDate4(startDate, endDate)
		list, err := service.StatisticsService.GetRetainedRepay(*s, *e, packageIds)
		if err != nil {
			this.showMsg(err.Error(), MSG_ERR)
		}
		count = len(list)
		this.Data["list"] = list
	} else if typeId == 3 {
		// 付费后留存
		s, e := service.FindByDate4(startDate, endDate)
		list, err := service.StatisticsService.GetRetainedAfterPay(*s, *e, packageIds)
		if err != nil {
			this.showMsg(err.Error(), MSG_ERR)
		}
		count = len(list)
		this.Data["list"] = list
	} else if typeId == 4 {
		// 复充后留存
		s, e := service.FindByDate4(startDate, endDate)
		list, err := service.StatisticsService.GetRetainedAfterRepay(*s, *e, packageIds)
		if err != nil {
			this.showMsg(err.Error(), MSG_ERR)
		}
		count = len(list)
		this.Data["list"] = list
	} else if typeId == 5 {
		// R留存
		s, e := service.FindByDate4(startDate, endDate)
		list, err := service.StatisticsService.GetRetainedR(*s, *e, packageIds, utypeIds)
		if err != nil {
			this.showMsg(err.Error(), MSG_ERR)
		}
		count = len(list)
		this.Data["list"] = list
	} else {
		// 用户留存
		s, e := service.FindByDate4(startDate, endDate)
		list, err := service.StatisticsService.GetRetained(*s, *e, packageIds)
		if err != nil {
			this.showMsg(err.Error(), MSG_ERR)
		}
		count = len(list)
		this.Data["list"] = list
	}

	usertype := []map[int]string{
		{0: "全部"},
		{10: "新手"},
		{1: "平民"},
		{2: "普R"},
		{3: "小R"},
		{4: "中R"},
		{5: "大R"},
		{6: "超大R"},
	}
	packageList, _ := service.ChannelService.GetChannelAll()
	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()
	packageClassAll, _ := service.ChannelService.GetChannelClassAll()

	this.Data["pageTitle"] = "用户留存"
	this.Data["count"] = count
	// this.Data["pageBar"] = libs.NewPager(page, count, this.pageSize, beego.URLFor("StatisticsController.Retained", "typeId", typeId, "package_id", packageId, "alias_id", aliasId, "alias_name", aliasName, "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["packageId"] = packageId
	this.Data["packageList"] = packageList
	this.Data["typeId"] = typeId
	this.Data["aliasId"] = aliasId
	this.Data["packageAliasList"] = packageAliasList
	this.Data["packageClassAll"] = packageClassAll
	// this.Data["aliasName"] = aliasName
	this.Data["classId"] = classId
	this.Data["usertypeId"] = utypeId
	this.Data["usertype"] = usertype
	this.display()
}

// Ltv统计
func (this *StatisticsController) LtvStat() {
	status, _ := this.GetInt("status")
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	packageId := this.GetString("package_id")
	typeId, _ := this.GetInt("typeId")
	aliasId := this.GetString("alias_id")
	classId := this.GetString("class_id")
	tabId, _ := this.GetInt("tab_id")

	var gameLtvStat, firstPayLtvStat bool
	pageTitle, ok := this.Data["pageTitle"].(string)
	if ok {
		gameLtvStat = pageTitle == "游戏营收LTV"
		firstPayLtvStat = pageTitle == "首充口径LTV"
	}

	if page < 1 {
		page = 1
	}

	today := service.NowTime()
	var stime, etime time.Time
	if endDate == "" {
		etime = today
		endDate = etime.Format("2006-01-02")
	}
	if startDate == "" {
		// 默认近十天数据
		stime = etime.AddDate(0, 0, -9)
		startDate = stime.Format("2006-01-02")
	}
	s, e := service.FindByDate4(startDate, endDate)
	if s == nil || e == nil {
		beego.Error("date range error", startDate, endDate)
		this.showMsg(fmt.Sprintf("date range error: %s, %s", startDate, endDate), MSG_ERR)
	}
	stime, etime = *s, *e

	if status == 0 {
		my_data := this.GetSession("my_select_pakeageid")
		if my_data != nil && packageId == "" {
			if packid, ok := my_data.(string); ok {
				packageId = packid
			}
		}
		status = 1
	}

	var packageIds []string
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

	var registAreas []int32
	if tabId == 1 {
		registAreas = append(registAreas, 0)
	} else if tabId == 2 {
		registAreas = append(registAreas, 1)
	} else if tabId == 3 {
		registAreas = append(registAreas, 2)
	}
	var list []*entity.LtvStat
	var err error
	if firstPayLtvStat {
		// 首充ltv
		list, err = service.StatisticsService.LtvStatFirstPay(stime, etime, registAreas, packageIds, gameLtvStat)
	} else {
		list, err = service.StatisticsService.LtvStatCK(stime, etime, registAreas, packageIds, gameLtvStat)
	}
	if err != nil {
		this.showMsg(err.Error(), MSG_ERR)
	}
	this.Data["list"] = list
	count := len(list)
	if typeId == 1 {
		this.Data["ltvStats"] = list
	}
	slice := make([]int, 0, 90)
	for i := range 90 {
		slice = append(slice, i+1)
	}
	this.Data["Slice"] = slice

	// packageList, _ := service.ChannelService.GetChannelAll()
	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()
	packageClassAll, _ := service.ChannelService.GetChannelClassAll()
	if !gameLtvStat {
		this.Data["pageTitle"] = "LTV统计"
	}
	this.Data["count"] = count
	// this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.LtvStat", "typeId", typeId, "package_id", packageId, "alias_id", aliasId, "alias_name", aliasName, "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["packageId"] = packageId
	// this.Data["packageList"] = packageList
	this.Data["typeId"] = typeId
	this.Data["aliasId"] = aliasId
	this.Data["classId"] = classId
	this.Data["packageAliasList"] = packageAliasList
	this.Data["packageClassAll"] = packageClassAll
	this.Data["status"] = status
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "ltvstatopt")
	this.Data["tabId"] = tabId
	this.display()
}

// 首充口径LTV
func (this *StatisticsController) FirstPayLtvStat() {
	this.Data["pageTitle"] = "首充口径LTV"
	this.LtvStat()
}

// 游戏营收LTV
func (this *StatisticsController) GameLtvStat() {
	this.Data["pageTitle"] = "游戏营收LTV"
	this.LtvStat()
}

// Ltv统计
func (this *StatisticsController) _LtvStat() {
	status, _ := this.GetInt("status")
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	packageId := this.GetString("package_id")
	typeId, _ := this.GetInt("typeId")
	aliasId := this.GetString("alias_id")
	classId := this.GetString("class_id")
	tabId, _ := this.GetInt("tab_id")
	chatid, _ := this.GetInt("chat_id")

	if page < 1 {
		page = 1
	}
	var start, end time.Time
	if startDate == "" || endDate == "" {
		// 默认近十天数据
		today := bson.Now()
		start, end = today.AddDate(0, 0, -9), today
	} else {
		start, _ = time.Parse("2006-01-02", startDate)
		end, _ = time.Parse("2006-01-02", endDate)
	}
	if status == 0 {
		my_data := this.GetSession("my_select_pakeageid")
		if my_data != nil && packageId == "" {
			if packid, ok := my_data.(string); ok {
				packageId = packid
			}
		}
		status = 1
	}
	m := bson.M{}
	var packageIds []string
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

	if len(packageIds) > 0 {
		m["ad__bundle_id"] = bson.M{"$in": packageIds}
	}
	// // 渠道别名
	// if aliasId != "" && aliasId != "0" {
	// 	if aliasId == "-" {
	// 		m["channel1"] = ""
	// 	} else {
	// 		if strings.Contains(aliasId, ",") {
	// 			palkage_arr := strings.Split(aliasId, ",")
	// 			temp_arr := make([]string, 0)
	// 			for _, item := range palkage_arr {
	// 				if item != "" && item != "0" && item != "-" {
	// 					temp_arr = append(temp_arr, item)
	// 				}
	// 			}
	// 			m["channel1"] = bson.M{"$in": temp_arr}
	// 		} else {
	// 			m["channel1"] = aliasId
	// 		}
	// 	}
	// }
	// if aliasName != "" {
	// 	m["channel1"] = bson.M{"$regex": aliasName, "$options": "i"}
	// }
	if tabId == 1 {
		m["regist_area"] = bson.M{"$ne": 1}
	} else if tabId == 2 {
		m["regist_area"] = 1
	} else if tabId == 3 {
		m["regist_area"] = 2
	}
	list, _ := service.StatisticsService.GetLtvStatList(page, this.pageSize, m, start, end)
	this.Data["list"] = list
	count := len(list)
	if typeId == 1 {
		// 图表
		var lables []string
		var datasets []any

		if count > 0 {
			for k := range list[0] {
				if strings.HasPrefix(k, "ltv") {
					lables = append(lables, k)
				}
			}
			sort.Slice(lables, func(i, j int) bool {
				var a, b int
				fmt.Sscanf(lables[i], "ltv%d", &a)
				fmt.Sscanf(lables[j], "ltv%d", &b)
				return a < b
			})

			for i, stat := range list {
				day := stat["date"].(string)
				data := make([]any, 0, 30)
				for _, ltvK := range lables {
					ltv := stat[ltvK].(map[string]string)
					var v string
					switch chatid {
					case 0:
						v = ltv["pay"]
					case 1:
						v = ltv["withdraw"]
					case 2:
						v = ltv["sub"]
					}
					data = append(data, v)
				}
				r := 255 / count * i
				g := 255 - 255/count*i
				b := int(math.Abs(float64(r - g + 255/count)))
				datasets = append(datasets, map[string]any{
					"label":           day,
					"data":            data,
					"borderWidth":     1,
					"backgroundColor": "rgba(0, 0, 0, 0)",
					"borderColor":     fmt.Sprintf("rgba(%d, %d, %d, 1)", r, g, b), // line color
				})
			}
			lables = utils.SliceMapping(lables, func(label string) string {
				return strings.Replace(label, "ltv", "D", 1)
			})
			this.Data["lables"] = lables
			this.Data["datasets"] = datasets
		}

		// var times, ltv1, ltv7, ltv15, ltv30 []any
		// for _, stat := range list {
		// 	times = append(times, stat["date"])
		// 	ltv1 = append(ltv1, stat["ltv1"])
		// 	ltv7 = append(ltv7, stat["ltv7"])
		// 	ltv15 = append(ltv15, stat["ltv15"])
		// 	ltv30 = append(ltv30, stat["ltv30"])
		// }
		// this.Data["times"] = times
		// this.Data["ltv1"] = ltv1
		// this.Data["ltv7"] = ltv7
		// this.Data["ltv15"] = ltv15
		// this.Data["ltv30"] = ltv30
	} else if typeId == 2 {
		// 数据总汇
		list, _ := service.StatisticsService.GetLtvStatSummaryList(page, this.pageSize, m, start, end)
		for i := 0; i < count/2; i++ {
			list[i], list[count-1-i] = list[count-1-i], list[i]
		}
		this.Data["list"] = list
		count = len(list)
	} else {
		// 表格：倒序显示
		for i := 0; i < count/2; i++ {
			list[i], list[count-1-i] = list[count-1-i], list[i]
		}
	}
	slice := make([]struct{}, 30)
	this.Data["Slice"] = slice

	chatType := map[int]string{
		0: "充值平均值",
		1: "提现平均值",
		2: "盈利",
	}
	this.Data["chatType"] = chatType
	this.Data["chatid"] = chatid
	// packageList, _ := service.ChannelService.GetChannelAll()
	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()
	packageClassAll, _ := service.ChannelService.GetChannelClassAll()
	this.Data["pageTitle"] = "LTV统计"
	this.Data["count"] = count
	// this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.LtvStat", "typeId", typeId, "package_id", packageId, "alias_id", aliasId, "alias_name", aliasName, "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["packageId"] = packageId
	// this.Data["packageList"] = packageList
	this.Data["typeId"] = typeId
	this.Data["aliasId"] = aliasId
	this.Data["classId"] = classId
	this.Data["packageAliasList"] = packageAliasList
	this.Data["packageClassAll"] = packageClassAll
	this.Data["status"] = status
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "ltvstatopt")
	this.Data["tabId"] = tabId
	this.display()
}

func (this *StatisticsController) LtvStatExport() { // 创建一个新的 Excel 文件
	status, _ := this.GetInt("status")
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	packageId := this.GetString("package_id")
	typeId, _ := this.GetInt("typeId")
	aliasId := this.GetString("alias_id")
	classId := this.GetString("class_id")
	tabId, _ := this.GetInt("tab_id")

	if page < 1 {
		page = 1
	}
	var start, end time.Time
	if startDate == "" || endDate == "" {
		// 默认近十天数据
		today := bson.Now()
		start, end = today.AddDate(0, 0, -9), today
	} else {
		start, _ = time.Parse("2006-01-02", startDate)
		end, _ = time.Parse("2006-01-02", endDate)
	}
	if status == 0 {
		my_data := this.GetSession("my_select_pakeageid")
		if my_data != nil && packageId == "" {
			if packid, ok := my_data.(string); ok {
				packageId = packid
			}
		}
		status = 1
	}
	m := bson.M{}
	var packageIds []string
	// if packageId != "" && packageId != "0" && packageId != "-" {
	// 	if strings.Contains(packageId, ",") {
	// 		palkage_arr := strings.Split(packageId, ",")
	// 		temp_arr := make([]string, 0)
	// 		for _, item := range palkage_arr {
	// 			if item != "" && item != "0" && item != "-" {
	// 				temp_arr = append(temp_arr, item)
	// 			}
	// 		}
	// 		m["channel"] = bson.M{"$in": temp_arr}
	// 	} else {
	// 		m["channel"] = packageId
	// 	}
	// 	// 存储缓存
	// 	this.SetSession("my_select_pakeageid", packageId)
	// } else {
	// 	// 清除缓存
	// 	this.DelSession("my_select_pakeageid")
	// }

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

	if len(packageIds) > 0 {
		m["ad__bundle_id"] = bson.M{"$in": packageIds}
	}
	// // 渠道别名
	// if aliasId != "" && aliasId != "0" {
	// 	if aliasId == "-" {
	// 		m["channel1"] = ""
	// 	} else {
	// 		if strings.Contains(aliasId, ",") {
	// 			palkage_arr := strings.Split(aliasId, ",")
	// 			temp_arr := make([]string, 0)
	// 			for _, item := range palkage_arr {
	// 				if item != "" && item != "0" && item != "-" {
	// 					temp_arr = append(temp_arr, item)
	// 				}
	// 			}
	// 			m["channel1"] = bson.M{"$in": temp_arr}
	// 		} else {
	// 			m["channel1"] = aliasId
	// 		}
	// 	}
	// }
	// if aliasName != "" {
	// 	m["channel1"] = bson.M{"$regex": aliasName, "$options": "i"}
	// }
	if tabId == 1 {
		m["regist_area"] = bson.M{"$ne": 1}
	} else if tabId == 2 {
		m["regist_area"] = 1
	} else if tabId == 3 {
		m["regist_area"] = 2
	}

	file := excelize.NewFile()
	sheet := "Sheet1"
	list := make([]bson.M, 0)
	if typeId == 0 {
		// 平均值
		list, _ = service.StatisticsService.GetLtvStatList(page, this.pageSize, m, start, end)
		count := len(list)
		// 表格：倒序显示
		for i := 0; i < count/2; i++ {
			list[i], list[count-1-i] = list[count-1-i], list[i]
		}
	} else {
		// 总汇
		list, _ = service.StatisticsService.GetLtvStatSummaryList(page, this.pageSize, m, start, end)
		count := len(list)
		// 表格：倒序显示
		for i := 0; i < count/2; i++ {
			list[i], list[count-1-i] = list[count-1-i], list[i]
		}

	}
	//设置表格头
	file.SetCellValue(sheet, "A1", "时间（复购率,人均付费在最后列）")
	file.SetCellValue(sheet, "B1", "新注册数")
	file.SetCellValue(sheet, "C1", "新设备数")
	file.SetCellValue(sheet, "D1", "付费用户数")

	file.SetCellValue(sheet, "E1", "D1 LTV")
	file.SetCellValue(sheet, "E2", "充值平均值")
	file.SetCellValue(sheet, "F2", "提现平均值")
	file.SetCellValue(sheet, "G2", "盈利")
	file.SetCellValue(sheet, "H1", "D2 LTV")
	file.SetCellValue(sheet, "H2", "充值平均值")
	file.SetCellValue(sheet, "I2", "提现平均值")
	file.SetCellValue(sheet, "J2", "盈利")
	file.SetCellValue(sheet, "K1", "D3 LTV")
	file.SetCellValue(sheet, "K2", "充值平均值")
	file.SetCellValue(sheet, "L2", "提现平均值")
	file.SetCellValue(sheet, "M2", "盈利")
	file.SetCellValue(sheet, "N1", "D4 LTV")
	file.SetCellValue(sheet, "N2", "充值平均值")
	file.SetCellValue(sheet, "O2", "提现平均值")
	file.SetCellValue(sheet, "P2", "盈利")
	file.SetCellValue(sheet, "Q1", "D5 LTV")
	file.SetCellValue(sheet, "Q2", "充值平均值")
	file.SetCellValue(sheet, "R2", "提现平均值")
	file.SetCellValue(sheet, "S2", "盈利")
	file.SetCellValue(sheet, "T1", "D6 LTV")
	file.SetCellValue(sheet, "T2", "充值平均值")
	file.SetCellValue(sheet, "U2", "提现平均值")
	file.SetCellValue(sheet, "V2", "盈利")
	file.SetCellValue(sheet, "W1", "D7 LTV")
	file.SetCellValue(sheet, "W2", "充值平均值")
	file.SetCellValue(sheet, "X2", "提现平均值")
	file.SetCellValue(sheet, "Y2", "盈利")
	file.SetCellValue(sheet, "Z1", "D8 LTV")
	file.SetCellValue(sheet, "Z2", "充值平均值")
	file.SetCellValue(sheet, "AA2", "提现平均值")
	file.SetCellValue(sheet, "AB2", "盈利")
	file.SetCellValue(sheet, "AC1", "D9 LTV")
	file.SetCellValue(sheet, "AC2", "充值平均值")
	file.SetCellValue(sheet, "AD2", "提现平均值")
	file.SetCellValue(sheet, "AE2", "盈利")
	file.SetCellValue(sheet, "AF1", "D10 LTV")
	file.SetCellValue(sheet, "AF2", "充值平均值")
	file.SetCellValue(sheet, "AG2", "提现平均值")
	file.SetCellValue(sheet, "AH2", "盈利")
	file.SetCellValue(sheet, "AI1", "D11 LTV")
	file.SetCellValue(sheet, "AI2", "充值平均值")
	file.SetCellValue(sheet, "AJ2", "提现平均值")
	file.SetCellValue(sheet, "AK2", "盈利")
	file.SetCellValue(sheet, "AL1", "D12 LTV")
	file.SetCellValue(sheet, "AL2", "充值平均值")
	file.SetCellValue(sheet, "AM2", "提现平均值")
	file.SetCellValue(sheet, "AN2", "盈利")
	file.SetCellValue(sheet, "AO1", "D13 LTV")
	file.SetCellValue(sheet, "AO2", "充值平均值")
	file.SetCellValue(sheet, "AP2", "提现平均值")
	file.SetCellValue(sheet, "AQ2", "盈利")
	file.SetCellValue(sheet, "AR1", "D14 LTV")
	file.SetCellValue(sheet, "AR2", "充值平均值")
	file.SetCellValue(sheet, "AS2", "提现平均值")
	file.SetCellValue(sheet, "AT2", "盈利")
	file.SetCellValue(sheet, "AU1", "D15 LTV")
	file.SetCellValue(sheet, "AU2", "充值平均值")
	file.SetCellValue(sheet, "AV2", "提现平均值")
	file.SetCellValue(sheet, "AW2", "盈利")
	file.SetCellValue(sheet, "AX1", "D16 LTV")
	file.SetCellValue(sheet, "AX2", "充值平均值")
	file.SetCellValue(sheet, "AY2", "提现平均值")
	file.SetCellValue(sheet, "AZ2", "盈利")
	file.SetCellValue(sheet, "BA1", "D17 LTV")
	file.SetCellValue(sheet, "BA2", "充值平均值")
	file.SetCellValue(sheet, "BB2", "提现平均值")
	file.SetCellValue(sheet, "BC2", "盈利")
	file.SetCellValue(sheet, "BD1", "D18 LTV")
	file.SetCellValue(sheet, "BD2", "充值平均值")
	file.SetCellValue(sheet, "BE2", "提现平均值")
	file.SetCellValue(sheet, "BF2", "盈利")
	file.SetCellValue(sheet, "BG1", "D19 LTV")
	file.SetCellValue(sheet, "BG2", "充值平均值")
	file.SetCellValue(sheet, "BH2", "提现平均值")
	file.SetCellValue(sheet, "BI2", "盈利")
	file.SetCellValue(sheet, "BJ1", "D20 LTV")
	file.SetCellValue(sheet, "BJ2", "充值平均值")
	file.SetCellValue(sheet, "BK2", "提现平均值")
	file.SetCellValue(sheet, "BL2", "盈利")
	file.SetCellValue(sheet, "BM1", "D21 LTV")
	file.SetCellValue(sheet, "BM2", "充值平均值")
	file.SetCellValue(sheet, "BN2", "提现平均值")
	file.SetCellValue(sheet, "BO2", "盈利")
	file.SetCellValue(sheet, "BP1", "D22 LTV")
	file.SetCellValue(sheet, "BP2", "充值平均值")
	file.SetCellValue(sheet, "BQ2", "提现平均值")
	file.SetCellValue(sheet, "BR2", "盈利")
	file.SetCellValue(sheet, "BS1", "D23 LTV")
	file.SetCellValue(sheet, "BS2", "充值平均值")
	file.SetCellValue(sheet, "BT2", "提现平均值")
	file.SetCellValue(sheet, "BU2", "盈利")
	file.SetCellValue(sheet, "BV1", "D24 LTV")
	file.SetCellValue(sheet, "BV2", "充值平均值")
	file.SetCellValue(sheet, "BW2", "提现平均值")
	file.SetCellValue(sheet, "BX2", "盈利")
	file.SetCellValue(sheet, "BY1", "D25 LTV")
	file.SetCellValue(sheet, "BY2", "充值平均值")
	file.SetCellValue(sheet, "BZ2", "提现平均值")
	file.SetCellValue(sheet, "CA2", "盈利")
	file.SetCellValue(sheet, "CB1", "D26 LTV")
	file.SetCellValue(sheet, "CB2", "充值平均值")
	file.SetCellValue(sheet, "CC2", "提现平均值")
	file.SetCellValue(sheet, "CD2", "盈利")
	file.SetCellValue(sheet, "CE1", "D27 LTV")
	file.SetCellValue(sheet, "CE2", "充值平均值")
	file.SetCellValue(sheet, "CF2", "提现平均值")
	file.SetCellValue(sheet, "CG2", "盈利")
	file.SetCellValue(sheet, "CH1", "D28 LTV")
	file.SetCellValue(sheet, "CH2", "充值平均值")
	file.SetCellValue(sheet, "CI2", "提现平均值")
	file.SetCellValue(sheet, "CJ2", "盈利")
	file.SetCellValue(sheet, "CK1", "D29 LTV")
	file.SetCellValue(sheet, "CK2", "充值平均值")
	file.SetCellValue(sheet, "CL2", "提现平均值")
	file.SetCellValue(sheet, "CM2", "盈利")
	file.SetCellValue(sheet, "CN1", "D30 LTV")
	file.SetCellValue(sheet, "CN2", "充值平均值")
	file.SetCellValue(sheet, "CO2", "提现平均值")
	file.SetCellValue(sheet, "CP2", "盈利")
	file.SetCellValue(sheet, "CQ1", "复购率")
	file.SetCellValue(sheet, "CR1", "人均付费次数")
	file.SetCellValue(sheet, "CS1", "复购人数")
	file.SetCellValue(sheet, "CT1", "总充值单数")

	// 合并表头单元格
	file.MergeCell("Sheet1", "A1", "A2")
	file.MergeCell("Sheet1", "B1", "B2")
	file.MergeCell("Sheet1", "C1", "C2")
	file.MergeCell("Sheet1", "D1", "D2")
	file.MergeCell("Sheet1", "E1", "G1")
	file.MergeCell("Sheet1", "H1", "J1")
	file.MergeCell("Sheet1", "K1", "M1")
	file.MergeCell("Sheet1", "N1", "P1")
	file.MergeCell("Sheet1", "Q1", "S1")
	file.MergeCell("Sheet1", "T1", "V1")
	file.MergeCell("Sheet1", "W1", "Y1")
	file.MergeCell("Sheet1", "Z1", "AB1")
	file.MergeCell("Sheet1", "AC1", "AE1")
	file.MergeCell("Sheet1", "AF1", "AH1")
	file.MergeCell("Sheet1", "AI1", "AK1")
	file.MergeCell("Sheet1", "AL1", "AN1")
	file.MergeCell("Sheet1", "AO1", "AQ1")
	file.MergeCell("Sheet1", "AR1", "AT1")
	file.MergeCell("Sheet1", "AU1", "AW1")
	file.MergeCell("Sheet1", "AX1", "AZ1")
	file.MergeCell("Sheet1", "BA1", "BC1")
	file.MergeCell("Sheet1", "BD1", "BF1")
	file.MergeCell("Sheet1", "BG1", "BI1")
	file.MergeCell("Sheet1", "BJ1", "BL1")
	file.MergeCell("Sheet1", "BM1", "BO1")
	file.MergeCell("Sheet1", "BP1", "BR1")
	file.MergeCell("Sheet1", "BS1", "BU1")
	file.MergeCell("Sheet1", "BV1", "BX1")
	file.MergeCell("Sheet1", "BY1", "CA1")
	file.MergeCell("Sheet1", "CB1", "CD1")
	file.MergeCell("Sheet1", "CE1", "CG1")
	file.MergeCell("Sheet1", "CH1", "CJ1")
	file.MergeCell("Sheet1", "CK1", "CM1")
	file.MergeCell("Sheet1", "CN1", "CP1")
	file.MergeCell("Sheet1", "CQ1", "CQ2")
	file.MergeCell("Sheet1", "CR1", "CR2")
	file.MergeCell("Sheet1", "CS1", "CS2")
	file.MergeCell("Sheet1", "CT1", "CT2")
	// 写入数据
	for idx, info := range list {
		row := idx + 3 // Excel 行号从 1 开始，因此要加 2
		file.SetCellValue(sheet, "A"+strconv.Itoa(row), info["date"])
		file.SetCellValue(sheet, "B"+strconv.Itoa(row), info["count"])
		file.SetCellValue(sheet, "C"+strconv.Itoa(row), info["scount"])
		file.SetCellValue(sheet, "D"+strconv.Itoa(row), info["payCount"])
		file.SetCellValue(sheet, "CQ"+strconv.Itoa(row), info["repayRate"])
		file.SetCellValue(sheet, "CR"+strconv.Itoa(row), info["payRate"])
		file.SetCellValue(sheet, "CS"+strconv.Itoa(row), info["pay2times"])
		file.SetCellValue(sheet, "CT"+strconv.Itoa(row), info["payOrders"])

		if info["ltv1"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "E"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "F"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "G"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "E"+strconv.Itoa(row), info["ltv1"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "F"+strconv.Itoa(row), info["ltv1"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "G"+strconv.Itoa(row), info["ltv1"].(map[string]string)["sub"])
		}
		if info["ltv2"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "H"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "I"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "J"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "H"+strconv.Itoa(row), info["ltv2"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "I"+strconv.Itoa(row), info["ltv2"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "J"+strconv.Itoa(row), info["ltv2"].(map[string]string)["sub"])
		}
		if info["ltv3"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "K"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "L"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "M"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "K"+strconv.Itoa(row), info["ltv3"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "L"+strconv.Itoa(row), info["ltv3"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "M"+strconv.Itoa(row), info["ltv3"].(map[string]string)["sub"])
		}
		if info["ltv4"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "N"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "O"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "P"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "N"+strconv.Itoa(row), info["ltv4"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "O"+strconv.Itoa(row), info["ltv4"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "P"+strconv.Itoa(row), info["ltv4"].(map[string]string)["sub"])
		}
		if info["ltv5"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "Q"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "R"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "S"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "Q"+strconv.Itoa(row), info["ltv5"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "R"+strconv.Itoa(row), info["ltv5"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "S"+strconv.Itoa(row), info["ltv5"].(map[string]string)["sub"])
		}
		if info["ltv6"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "T"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "U"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "V"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "T"+strconv.Itoa(row), info["ltv6"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "U"+strconv.Itoa(row), info["ltv6"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "V"+strconv.Itoa(row), info["ltv6"].(map[string]string)["sub"])
		}
		if info["ltv7"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "W"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "X"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "Y"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "W"+strconv.Itoa(row), info["ltv7"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "X"+strconv.Itoa(row), info["ltv7"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "Y"+strconv.Itoa(row), info["ltv7"].(map[string]string)["sub"])
		}
		if info["ltv8"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "Z"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "AA"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "AB"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "Z"+strconv.Itoa(row), info["ltv8"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "AA"+strconv.Itoa(row), info["ltv8"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "AB"+strconv.Itoa(row), info["ltv8"].(map[string]string)["sub"])
		}
		if info["ltv9"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "AC"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "AD"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "AE"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "AC"+strconv.Itoa(row), info["ltv9"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "AD"+strconv.Itoa(row), info["ltv9"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "AE"+strconv.Itoa(row), info["ltv9"].(map[string]string)["sub"])
		}
		if info["ltv10"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "AF"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "AG"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "AH"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "AF"+strconv.Itoa(row), info["ltv10"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "AG"+strconv.Itoa(row), info["ltv10"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "AH"+strconv.Itoa(row), info["ltv10"].(map[string]string)["sub"])
		}
		if info["ltv11"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "AI"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "AJ"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "AK"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "AI"+strconv.Itoa(row), info["ltv11"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "AJ"+strconv.Itoa(row), info["ltv11"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "AK"+strconv.Itoa(row), info["ltv11"].(map[string]string)["sub"])
		}
		if info["ltv12"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "AL"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "AM"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "AN"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "AL"+strconv.Itoa(row), info["ltv12"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "AM"+strconv.Itoa(row), info["ltv12"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "AN"+strconv.Itoa(row), info["ltv12"].(map[string]string)["sub"])
		}
		if info["ltv13"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "AO"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "AP"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "AQ"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "AO"+strconv.Itoa(row), info["ltv13"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "AP"+strconv.Itoa(row), info["ltv13"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "AQ"+strconv.Itoa(row), info["ltv13"].(map[string]string)["sub"])
		}
		if info["ltv14"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "AR"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "AS"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "AT"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "AP"+strconv.Itoa(row), info["ltv14"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "AS"+strconv.Itoa(row), info["ltv14"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "AT"+strconv.Itoa(row), info["ltv14"].(map[string]string)["sub"])
		}
		if info["ltv15"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "AU"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "AV"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "AW"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "AU"+strconv.Itoa(row), info["ltv15"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "AV"+strconv.Itoa(row), info["ltv15"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "AW"+strconv.Itoa(row), info["ltv15"].(map[string]string)["sub"])
		}
		if info["ltv16"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "AX"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "AY"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "AZ"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "AX"+strconv.Itoa(row), info["ltv16"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "AY"+strconv.Itoa(row), info["ltv16"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "AZ"+strconv.Itoa(row), info["ltv16"].(map[string]string)["sub"])
		}
		if info["ltv17"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "BA"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "BB"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "BC"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "BA"+strconv.Itoa(row), info["ltv17"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "BB"+strconv.Itoa(row), info["ltv17"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "BC"+strconv.Itoa(row), info["ltv17"].(map[string]string)["sub"])
		}
		if info["ltv18"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "BD"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "BE"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "BF"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "BD"+strconv.Itoa(row), info["ltv18"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "BE"+strconv.Itoa(row), info["ltv18"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "BF"+strconv.Itoa(row), info["ltv18"].(map[string]string)["sub"])
		}
		if info["ltv19"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "BG"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "BH"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "BI"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "BG"+strconv.Itoa(row), info["ltv19"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "BH"+strconv.Itoa(row), info["ltv19"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "BI"+strconv.Itoa(row), info["ltv19"].(map[string]string)["sub"])
		}
		if info["ltv20"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "BJ"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "BK"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "BL"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "BJ"+strconv.Itoa(row), info["ltv20"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "BK"+strconv.Itoa(row), info["ltv20"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "BL"+strconv.Itoa(row), info["ltv20"].(map[string]string)["sub"])
		}
		if info["ltv21"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "BM"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "BN"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "BO"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "BM"+strconv.Itoa(row), info["ltv21"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "BN"+strconv.Itoa(row), info["ltv21"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "BO"+strconv.Itoa(row), info["ltv21"].(map[string]string)["sub"])
		}
		if info["ltv22"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "BP"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "BQ"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "BR"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "BP"+strconv.Itoa(row), info["ltv22"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "BQ"+strconv.Itoa(row), info["ltv22"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "BR"+strconv.Itoa(row), info["ltv22"].(map[string]string)["sub"])
		}
		if info["ltv23"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "BS"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "BT"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "BU"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "BS"+strconv.Itoa(row), info["ltv23"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "BT"+strconv.Itoa(row), info["ltv23"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "BU"+strconv.Itoa(row), info["ltv23"].(map[string]string)["sub"])
		}

		if info["ltv24"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "BV"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "BW"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "BX"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "BV"+strconv.Itoa(row), info["ltv24"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "BW"+strconv.Itoa(row), info["ltv24"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "BX"+strconv.Itoa(row), info["ltv24"].(map[string]string)["sub"])
		}
		if info["ltv25"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "BY"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "BZ"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "CA"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "BY"+strconv.Itoa(row), info["ltv25"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "BZ"+strconv.Itoa(row), info["ltv25"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "CA"+strconv.Itoa(row), info["ltv25"].(map[string]string)["sub"])
		}
		if info["ltv26"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "CB"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "CC"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "CD"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "CB"+strconv.Itoa(row), info["ltv26"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "CC"+strconv.Itoa(row), info["ltv26"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "CD"+strconv.Itoa(row), info["ltv26"].(map[string]string)["sub"])
		}
		if info["ltv27"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "CE"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "CF"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "CG"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "CE"+strconv.Itoa(row), info["ltv27"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "CF"+strconv.Itoa(row), info["ltv27"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "CG"+strconv.Itoa(row), info["ltv27"].(map[string]string)["sub"])
		}
		if info["ltv28"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "CH"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "CI"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "CJ"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "CH"+strconv.Itoa(row), info["ltv28"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "CI"+strconv.Itoa(row), info["ltv28"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "CJ"+strconv.Itoa(row), info["ltv28"].(map[string]string)["sub"])
		}
		if info["ltv29"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "CK"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "CL"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "CM"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "CK"+strconv.Itoa(row), info["ltv29"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "CL"+strconv.Itoa(row), info["ltv29"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "CM"+strconv.Itoa(row), info["ltv29"].(map[string]string)["sub"])
		}
		if info["ltv30"].(map[string]string)["pay"] == "/" {
			file.SetCellValue(sheet, "CN"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "CO"+strconv.Itoa(row), "/")
			file.SetCellValue(sheet, "CP"+strconv.Itoa(row), "/")
		} else {
			file.SetCellValue(sheet, "CN"+strconv.Itoa(row), info["ltv30"].(map[string]string)["pay"])
			file.SetCellValue(sheet, "CO"+strconv.Itoa(row), info["ltv30"].(map[string]string)["withdraw"])
			file.SetCellValue(sheet, "CP"+strconv.Itoa(row), info["ltv30"].(map[string]string)["sub"])
		}
	}

	// err := file.SaveAs("output-ltv.xlsx")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	//构造文件名称
	fileName := "LTV统计_" + time.Now().Format("20060102150405") + ".xlsx"
	// 解决文件名中文乱码问题
	fileName = path.Base(fileName)
	fileName = url.QueryEscape(fileName)
	this.Ctx.Output.Header("Content-type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	this.Ctx.Output.Header("Content-Disposition", "attachment;filename="+fileName)
	this.Ctx.Output.Header("Pragma", "No-cache")
	this.Ctx.Output.Header("Cache-Control", "No-cache")
	this.Ctx.Output.Header("Expires", "0")
	// var buffer bytes.Buffer
	if err := file.Write(this.Ctx.ResponseWriter); err != nil {
		logs.Error(err)
	}
}

// 回收周期计算
func (this *StatisticsController) RecoveryCycle() {
	startDate := this.GetString("start_date") // 注册时间1
	endDate := this.GetString("end_date")     // 注册时间2
	packageId := this.GetString("package_id")
	aliasId := this.GetString("alias_id")
	classId := this.GetString("class_id")
	exchangeRate, _ := this.GetFloat("exchangeRate") // 未登录天数
	typeId, _ := this.GetInt("typeId")
	days, _ := this.GetInt("days")
	consume := this.GetString("consume")
	_ = consume
	username := this.auth.GetUser().UserName

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			debug.PrintStack()
		}
	}()
	vers := service.GetVersions()
	if exchangeRate == 0 {
		if vers == 1 || vers == 0 {
			exchangeRate = 96
		} else if vers == 2 {
			exchangeRate = 120
		}
	}

	today := bson.Now()
	if startDate == "" {
		// 默认近七天数据
		if typeId == 1 {
			startDate = today.AddDate(0, 0, -7).Format("2006-01-02")
		} else {
			startDate = today.AddDate(0, 0, -30).Format("2006-01-02")
		}
	}
	if endDate == "" {
		endDate = today.Format("2006-01-02")
	}

	var start, end time.Time
	s, e := service.FindByDate4(startDate, endDate)
	if s != nil {
		start = *s
	} else {
		startDate = start.Format("2006-01-02")
	}
	if e != nil {
		end = *e
	} else {
		endDate = end.Format("2006-01-02")
	}

	if days == 0 {
		days = 30
	}
	dfDays := int(today.Sub(start).Hours()/24) + 1
	if dfDays < days {
		days = dfDays
	}
	if days < 0 {
		days = 0
	}

	dateConsume := map[string]float64{
		// "2024-08-28": 200,
	}
	if consume != "" {
		var consumes = make(map[string]any)
		if err := json.Unmarshal([]byte(consume), &consumes); err != nil {
			beego.Error("consume parse error:", consume, err)
		} else {
			for sdate, value := range consumes {
				dateConsume[sdate] = utils.ToFloat64(value)
			}
		}
	}

	var packages, aliases, classes []string
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
			packages = append(packages, temp_arr...)
		} else {
			packageIds = append(packageIds, packageId)
			packages = append(packages, packageId)
		}
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
			aliases = append(aliases, temp_arr...)
		} else {
			aliasIds = append(aliasIds, aliasId)
			aliases = append(aliases, aliasId)
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
			classes = append(classes, temp_arr...)
		} else {
			classIds = append(classIds, classId)
			classes = append(classes, classId)
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

	// 选择的渠道类,渠道别名,渠道
	// packages, aliases, classes
	var ptype int // 0全部,1渠道类,2别名,3渠道,4混合
	sort.Strings(classes)
	sort.Strings(aliases)
	sort.Strings(packages)
	var ps []string
	for _, cls := range classes {
		ps = append(ps, cls)
		ptype = 1
	}
	for _, alias := range aliases {
		ps = append(ps, alias)
		ptype = 2
		if ptype != 0 && ptype != 2 {
			ptype = 4
		}
	}
	for _, pkg := range packages {
		ps = append(ps, pkg)
		ptype = 3
		if ptype != 0 && ptype != 3 {
			ptype = 4
		}
	}
	pss := strings.Join(ps, ";")

	var err error
	if typeId == 0 {
		data, e := service.StatisticsService.RecoveryCycle(start, end, exchangeRate, dateConsume, packageIds, username, ptype, pss)
		if e != nil {
			err = e
		} else {
			this.Data["count"] = len(data)
			this.Data["list"] = data
		}
	} else {
		datas, results, e := service.StatisticsService.RecoveryCycleStats(days, start, end, exchangeRate, dateConsume, packageIds, username, ptype, pss)
		if e != nil {
			err = e
		} else {
			var labels = []string{"首日"}
			for i := 1; i < days; i++ {
				labels = append(labels, fmt.Sprintf("%d日", i))
			}
			// 表格：倒序显示
			count := len(datas)
			for i := 0; i < count/2; i++ {
				datas[i], datas[count-1-i] = datas[count-1-i], datas[i]
			}
			this.Data["count"] = count
			this.Data["list"] = datas
			this.Data["lables"] = labels
			var datasets []any
			for i, item := range datas {
				r := 255 / count * i
				g := 255 - 255/count*i
				b := int(math.Abs(float64(r - g + 255/count)))
				datasets = append(datasets, map[string]any{
					"label":           item.SDate,
					"data":            results[item.SDate],
					"borderWidth":     1,
					"backgroundColor": "rgba(0, 0, 0, 0)",
					"borderColor":     fmt.Sprintf("rgba(%d, %d, %d, 1)", r, g, b), // line color
				})
			}
			this.Data["datasets"] = datasets
		}
	}

	if err != nil {
		fmt.Println(err)
		this.showMsg(err.Error(), MSG_ERR, "")
		return
	}

	// 显示渠道别名
	var aliasIds string
	if len(packageIds) > 0 {
		channels, err := service.ChannelService.GetChannelByIds(packageIds)
		if err != nil {
			beego.Error(err)
		} else {
			for _, channel := range channels {
				alias := channel.Name1
				if alias == "" {
					alias = channel.Name
				}
				aliasIds += alias + "， "
			}
		}
	}

	packageList, _ := service.ChannelService.GetChannelAll()
	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()
	packageClassAll, _ := service.ChannelService.GetChannelClassAll()

	this.Data["pageTitle"] = "回收周期计算"
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["packageId"] = packageId
	this.Data["aliasId"] = aliasId
	this.Data["classId"] = classId
	this.Data["packageList"] = packageList
	this.Data["packageAliasList"] = packageAliasList
	this.Data["packageClassAll"] = packageClassAll
	this.Data["exchangeRate"] = exchangeRate
	this.Data["aliasIds"] = aliasIds
	this.Data["consume"] = consume
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "recoverycycleopt")
	this.Data["typeId"] = typeId
	this.Data["days"] = days

	this.display()
}

// 回收周期计算导出
func (this *StatisticsController) RecoveryCycleExport() {
	startDate := this.GetString("start_date") // 注册时间1
	endDate := this.GetString("end_date")     // 注册时间2
	packageId := this.GetString("package_id")
	aliasId := this.GetString("alias_id")
	classId := this.GetString("class_id")
	exchangeRate, _ := this.GetFloat("exchangeRate") // 未登录天数
	consume := this.GetString("consume")
	_ = consume
	username := this.auth.GetUser().UserName
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			debug.PrintStack()
		}
	}()

	if exchangeRate == 0 {
		exchangeRate = 96
	}
	today := bson.Now()
	var start, end = today.AddDate(0, 0, -30), today
	s, e := service.FindByDate4(startDate, endDate)
	if s != nil {
		start = *s
	}
	if e != nil {
		end = *e
	}

	dateConsume := map[string]float64{
		// "2024-08-28": 200,
	}
	if consume != "" {
		var consumes = make(map[string]any)
		if err := json.Unmarshal([]byte(consume), &consumes); err != nil {
			beego.Error("consume parse error:", consume, err)
		} else {
			for sdate, value := range consumes {
				dateConsume[sdate] = utils.ToFloat64(value)
			}
		}
	}
	var packages, aliases, classes []string
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
			packages = append(packages, temp_arr...)
		} else {
			packageIds = append(packageIds, packageId)
			packages = append(packages, packageId)
		}
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
			aliases = append(aliases, temp_arr...)
		} else {
			aliasIds = append(aliasIds, aliasId)
			aliases = append(aliases, aliasId)
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
			classes = append(classes, temp_arr...)
		} else {
			classIds = append(classIds, classId)
			classes = append(classes, classId)
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

	// 选择的渠道类,渠道别名,渠道
	// packages, aliases, classes
	var ptype int // 0全部,1渠道类,2别名,3渠道,4混合
	sort.Strings(classes)
	sort.Strings(aliases)
	sort.Strings(packages)
	var ps []string
	for _, cls := range classes {
		ps = append(ps, cls)
		ptype = 1
	}
	for _, alias := range aliases {
		ps = append(ps, alias)
		ptype = 2
		if ptype != 0 && ptype != 2 {
			ptype = 4
		}
	}
	for _, pkg := range packages {
		ps = append(ps, pkg)
		ptype = 3
		if ptype != 0 && ptype != 3 {
			ptype = 4
		}
	}
	pss := strings.Join(ps, ";")

	data, err := service.StatisticsService.RecoveryCycle(start, end, exchangeRate, dateConsume, packageIds, username, ptype, pss)
	if err != nil {
		fmt.Println(err)
		this.showMsg(err.Error(), MSG_ERR, "")
		return
	}

	// 显示渠道别名
	var aliasIds string
	if len(packageIds) > 0 {
		channels, err := service.ChannelService.GetChannelByIds(packageIds)
		if err != nil {
			beego.Error(err)
		} else {
			for _, channel := range channels {
				alias := channel.Name1
				if alias == "" {
					alias = channel.Name
				}
				aliasIds += alias + "， "
			}
		}
	}

	// 检查是否有数据可导出
	if len(data) == 0 {
		this.showMsg("未查询到可以导出的数据！", MSG_ERR, "")
		return
	}

	file := excelize.NewFile()
	sheet := "Sheet1"
	//设置表格头
	file.SetCellValue(sheet, "A1", "渠道别名")
	file.SetCellValue(sheet, "B1", "日期")
	file.SetCellValue(sheet, "C1", "消耗")
	file.SetCellValue(sheet, "D1", "新增用户数")
	file.SetCellValue(sheet, "E1", "首充人数")
	file.SetCellValue(sheet, "F1", "累计付费人数")
	file.SetCellValue(sheet, "G1", "累计付费率")
	file.SetCellValue(sheet, "H1", "单个注成本")
	file.SetCellValue(sheet, "I1", "单个首充成本")
	file.SetCellValue(sheet, "J1", "单个付费成本")
	file.SetCellValue(sheet, "K1", "累计利润")
	file.SetCellValue(sheet, "L1", "首日利润")
	file.SetCellValue(sheet, "M1", "首日ROI")
	file.SetCellValue(sheet, "N1", "3日利润")
	file.SetCellValue(sheet, "O1", "3日ROI")
	file.SetCellValue(sheet, "P1", "7日利润")
	file.SetCellValue(sheet, "Q1", "7日ROI")
	file.SetCellValue(sheet, "R1", "10日利润")
	file.SetCellValue(sheet, "S1", "10日ROI")
	file.SetCellValue(sheet, "T1", "15日利润")
	file.SetCellValue(sheet, "U1", "15日ROI")
	file.SetCellValue(sheet, "V1", "20日利润")
	file.SetCellValue(sheet, "W1", "20日ROI")
	file.SetCellValue(sheet, "X1", "30日利润")
	file.SetCellValue(sheet, "Y1", "30日ROI")
	file.SetCellValue(sheet, "Z1", "45日利润")
	file.SetCellValue(sheet, "AA1", "45日ROI")
	file.SetCellValue(sheet, "AB1", "60日利润")
	file.SetCellValue(sheet, "AC1", "60日ROI")
	for i, info := range data {
		row := i + 2 // Excel 行号从 1 开始，因此要加 2
		if i == 0 {
			file.SetCellValue(sheet, "A"+strconv.Itoa(row), aliasIds)
		}
		file.SetCellValue(sheet, "B"+strconv.Itoa(row), info.SDate)
		file.SetCellValue(sheet, "C"+strconv.Itoa(row), info.Consume)
		file.SetCellValue(sheet, "D"+strconv.Itoa(row), info.NewRegs)
		file.SetCellValue(sheet, "E"+strconv.Itoa(row), info.FirstCharges)
		file.SetCellValue(sheet, "F"+strconv.Itoa(row), info.Charges)
		file.SetCellValue(sheet, "G"+strconv.Itoa(row), info.PayRates)
		file.SetCellValue(sheet, "H"+strconv.Itoa(row), info.EachRegCost)
		file.SetCellValue(sheet, "I"+strconv.Itoa(row), info.EachChargeCost)
		file.SetCellValue(sheet, "J"+strconv.Itoa(row), info.EachPayCost)
		file.SetCellValue(sheet, "K"+strconv.Itoa(row), info.Profit)
		file.SetCellValue(sheet, "L"+strconv.Itoa(row), info.ROI1Profit)
		file.SetCellValue(sheet, "M"+strconv.Itoa(row), info.ROI1)
		file.SetCellValue(sheet, "N"+strconv.Itoa(row), info.ROI3Profit)
		file.SetCellValue(sheet, "O"+strconv.Itoa(row), info.ROI3)
		file.SetCellValue(sheet, "P"+strconv.Itoa(row), info.ROI7Profit)
		file.SetCellValue(sheet, "Q"+strconv.Itoa(row), info.ROI7)
		file.SetCellValue(sheet, "R"+strconv.Itoa(row), info.ROI10Profit)
		file.SetCellValue(sheet, "S"+strconv.Itoa(row), info.ROI10)
		file.SetCellValue(sheet, "T"+strconv.Itoa(row), info.ROI15Profit)
		file.SetCellValue(sheet, "U"+strconv.Itoa(row), info.ROI15)
		file.SetCellValue(sheet, "V"+strconv.Itoa(row), info.ROI20Profit)
		file.SetCellValue(sheet, "W"+strconv.Itoa(row), info.ROI20)
		file.SetCellValue(sheet, "X"+strconv.Itoa(row), info.ROI30Profit)
		file.SetCellValue(sheet, "Y"+strconv.Itoa(row), info.ROI30)
		file.SetCellValue(sheet, "Z"+strconv.Itoa(row), info.ROI45Profit)
		file.SetCellValue(sheet, "AA"+strconv.Itoa(row), info.ROI45)
		file.SetCellValue(sheet, "AB"+strconv.Itoa(row), info.ROI60Profit)
		file.SetCellValue(sheet, "AC"+strconv.Itoa(row), info.ROI60)
	}
	file.MergeCell("Sheet1", "A2", fmt.Sprintf("A%d", len(data)+1))

	//构造文件名称
	fileName := "回收周期计算_" + time.Now().Format("20060102150405") + ".xlsx"
	// 解决文件名中文乱码问题
	fileName = path.Base(fileName)
	fileName = url.QueryEscape(fileName)
	this.Ctx.Output.Header("Content-type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	this.Ctx.Output.Header("Content-Disposition", "attachment;filename="+fileName)
	this.Ctx.Output.Header("Pragma", "No-cache")
	this.Ctx.Output.Header("Cache-Control", "No-cache")
	this.Ctx.Output.Header("Expires", "0")
	if err := file.Write(this.Ctx.ResponseWriter); err != nil {
		logs.Error(err)
	}
}

// 流失玩家统计
func (this *StatisticsController) LosingPlayers() {
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")      // 注册时间1
	endDate := this.GetString("end_date")          // 注册时间2
	notloginDays, _ := this.GetInt("notloginDays") // 未登录天数

	winScore1 := this.GetString("winScore1") // 当前赢分1
	winScore2 := this.GetString("winScore2") // 当前赢分2
	recharge1 := this.GetString("recharge1") // 充值金额1
	recharge2 := this.GetString("recharge2") // 充值金额2

	if page < 1 {
		page = 1
	}
	// if notloginDays < 1 {
	// 	notloginDays = 1
	// }

	var (
		count int64
		list  any
	)
	// 未登录天数默认空不搜索
	if notloginDays > 0 {
		// 注册时间-距今时间
		today := bson.Now()
		if startDate == "" {
			// 默认近十天数据
			startDate = today.AddDate(0, 0, -9).Format("2006-01-02")
		}
		if endDate == "" {
			endDate = today.AddDate(0, 0, -1).Format("2006-01-02")
		}
		m := service.FindByDate(startDate, endDate, "ctime", "ctime")
		m["robot"] = false
		m["simulation_robot"] = false

		// 未登录天数
		loginTimeLt := today.AddDate(0, 0, -notloginDays).Format(utils.FORMAT)
		loginTimeLt2 := utils.Str2Time(loginTimeLt, service.Location())
		// loginTimeLt3 := utils.Time2Stamp(loginTimeLt2)
		m["login_time"] = bson.M{"$lt": loginTimeLt2}

		// 充值金额
		if recharge1 != "" || recharge2 != "" {
			r1, _ := strconv.Atoi(recharge1)
			r2, _ := strconv.Atoi(recharge2)
			r1, r2 = r1*100, r2*100
			if recharge1 != "" && recharge2 != "" {
				m["money"] = bson.M{"$gte": r1, "$lte": r2}
			} else if recharge1 != "" {
				m["money"] = bson.M{"$gte": r1}
			} else if recharge2 != "" {
				m["money"] = bson.M{"$lte": r2}
			}
		}

		pipe := []bson.M{{"$match": m}}

		// 当前赢分
		var m2 bson.M
		if winScore1 != "" || winScore2 != "" {
			w1, _ := strconv.Atoi(winScore1)
			w2, _ := strconv.Atoi(winScore2)
			w1, w2 = w1*100, w2*100
			if winScore1 != "" && winScore2 != "" {
				m2 = bson.M{"$match": bson.M{"fwin": bson.M{"$gte": w1, "$lte": w2}}}
			} else if winScore1 != "" {
				m2 = bson.M{"$match": bson.M{"fwin": bson.M{"$gte": w1}}}
			} else if winScore2 != "" {
				m2 = bson.M{"$match": bson.M{"fwin": bson.M{"$lte": w2}}}
			}
		}
		if m2 != nil {
			pipe = append(pipe, bson.M{"$project": bson.M{
				"ctime": "$ctime", "login_time": "$login_time", "diamond": "$diamond", "money": "$money", "cash_out": "$cash_out", "shadow_diamond": "$shadow_diamond",
				"fwin": bson.M{"$subtract": []any{bson.M{"$add": []string{"$diamond", "$shadow_diamond", "$cash_out"}}, "$money"}},
			}})
			pipe = append(pipe, m2)
		}

		count, _ = service.StatisticsService.LosingPlayersCount(page, this.pageSize, pipe)
		list, _ = service.StatisticsService.LosingPlayersList(page, this.pageSize, pipe)
	}
	this.Data["list"] = list

	this.Data["pageTitle"] = "流失玩家分析页"
	this.Data["count"] = count
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.LosingPlayers", "start_date", startDate, "notloginDays", notloginDays, "recharge1", recharge1, "recharge2", recharge2), true).ToString()
	this.Data["page"] = page
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	if notloginDays != 0 {
		this.Data["notloginDays"] = notloginDays
	}
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "losingplayers")
	this.Data["recharge1"] = recharge1
	this.Data["recharge2"] = recharge2
	this.Data["winScore1"] = winScore1
	this.Data["winScore2"] = winScore2
	this.display()
}

// 资源流动
func (this *StatisticsController) Resource() {
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	typeId, _ := this.GetInt("typeId")
	if page < 1 {
		page = 1
	}
	if startDate == "" && endDate == "" {
		// 默认近十天数据
		today := bson.Now()
		startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -9).Format("2006-01-02"))
		endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
	}
	m := service.FindByDate1(startDate, endDate, "date", "date")
	count := int64(0)
	switch typeId {
	case 0:
		// 全局
		count, _ = service.StatisticsService.GetGlobalTotal(m)
		list, _ := service.StatisticsService.GetGlobalList(page, this.pageSize, m)
		this.Data["list"] = list
	case 1:
		// 游戏
		count, _ = service.StatisticsService.GetGamesTotal(m)
		list, _ := service.StatisticsService.GetGamesList(page, this.pageSize, m)
		this.Data["list"] = list
		// case 2:
		// 	// 充值活动
		// 	count, _ = service.StatisticsService.GetRechargeTotal(m)
		// 	list, _ := service.StatisticsService.GetRechargeList(page, this.pageSize, m)
		// 	this.Data["list"] = list
		// case 3:
		// 	// 免费活动
		// 	count, _ = service.StatisticsService.GetFreeTotal(m)
		// 	list, _ := service.StatisticsService.GetFreeList(page, this.pageSize, m)
		// 	this.Data["list"] = list
	}

	this.Data["pageTitle"] = "资源流动"
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.Resource", "typeId", typeId, "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["count"] = count
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["typeId"] = typeId
	this.display()
}

// 用户资源
func (this *StatisticsController) User() {
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	status_id, _ := this.GetInt("status_id")
	if page < 1 {
		page = 1
	}
	if startDate == "" && endDate == "" {
		// 默认近十天数据
		today := bson.Now()
		startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -9).Format("2006-01-02"))
		endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
		status_id = 1
	}
	m := service.FindByDate1(startDate, endDate, "date", "date")
	m["u_type"] = status_id
	count, _ := service.StatisticsService.GetUserResourceTotal(m)
	list, _ := service.StatisticsService.GetUserResourceList(page, this.pageSize, m)

	statusList := map[int]string{
		0: "全部用户",
		1: "活跃用户",
	}
	this.Data["pageTitle"] = "用户资源"
	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.User", "status_id", status_id, "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["statusId"] = status_id
	this.Data["statusList"] = statusList
	this.display()
}

// 充提排名
func (this *StatisticsController) Pay() {
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	typeId, _ := strconv.Atoi(this.GetString("typeId"))
	if page < 1 {
		page = 1
	}

	list := make([]entity.PayRanking, 0)
	count := 0
	m := service.FindByDate(startDate, endDate, "pay_time", "pay_time")
	m["userid"] = bson.M{"$not": bson.RegEx{Pattern: "^10", Options: ""}}
	if typeId == 0 {
		// 充值
		m["order_status"] = 4
	} else if typeId == 1 {
		// 提现
		m["order_status"] = 2
	} else if typeId == 2 {
		// 盈利
		m["order_status"] = 4
	}
	paylist, err := service.StatisticsService.GetPayRanking(typeId, m)
	this.checkError(err)
	count = len(paylist)
	// 根据分页参数对用户数据进行切片
	startIndex := (page - 1) * this.pageSize
	endIndex := (page) * this.pageSize
	if endIndex > count {
		endIndex = count
	}
	list = paylist[startIndex:endIndex]

	this.Data["pageTitle"] = "充提排名"
	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.Pay", "typeId", typeId, "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["typeId"] = typeId
	this.display()
}

// 实时数据对比
func (this *StatisticsController) Realtime() {
	typeId, _ := this.GetInt("type_id")
	startDate := this.GetString("start_date")
	beforeDate := ""
	if startDate == "" {
		// 默认当天
		today := bson.Now()
		startDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
		beforeDate = fmt.Sprintf("%s", today.AddDate(0, 0, -1).Format("2006-01-02"))

		// startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -10).Format("2006-01-02"))
	} else {
		// 解析日期字符串
		date, err := time.Parse("2006-01-02", startDate)
		this.checkError(err)
		beforeDate = fmt.Sprintf("%s", date.AddDate(0, 0, -1).Format("2006-01-02"))
	}

	dates := []string{"00:00", "01:00", "02:00", "03:00", "04:00", "05:00", "06:00", "07:00", "08:00", "09:00", "10:00", "11:00", "12:00", "13:00", "14:00", "15:00", "16:00", "17:00", "18:00", "19:00", "20:00", "21:00", "22:00", "23:00"}
	data := make([]interface{}, 0)  // 要展示当前的数据
	data1 := make([]interface{}, 0) // 要展示前一日的数据
	m := service.FindByDate(startDate, startDate, "date", "date")
	list, err := service.StatisticsService.GetRealTimeData(m)
	this.checkError(err)
	// 前一日
	m1 := service.FindByDate(beforeDate, beforeDate, "date", "date")
	list1, err1 := service.StatisticsService.GetRealTimeData(m1)
	this.checkError(err1)
	for _, t := range dates {
		// isExist := false
		var val interface{}
		var val1 interface{}
		for _, item := range list {
			if item.SDate == t {
				switch typeId {
				case 0:
					val = item.PayAmount
				case 1:
					val = item.WithdrawAmount
				case 2:
					val = item.PayNumber
				case 3:
					val = item.WithdrawNumber
				case 4:
					val = item.LoginNumber
				case 5:
					val = item.OnlineNumber
				}
			}
		}
		for _, item := range list1 {
			if item.SDate == t {
				switch typeId {
				case 0:
					val1 = item.PayAmount
				case 1:
					val1 = item.WithdrawAmount
				case 2:
					val1 = item.PayNumber
				case 3:
					val1 = item.WithdrawNumber
				case 4:
					val1 = item.LoginNumber
				case 5:
					val1 = item.OnlineNumber
				}
			}
		}
		data = append(data, val)
		data1 = append(data1, val1)
	}

	typeList := map[int]string{
		0: "充值金额",
		1: "提现金额",
		2: "充值人数",
		3: "提现人数",
		4: "登录",
		5: "在线",
	}
	this.Data["pageTitle"] = "实时数据对比"
	this.Data["TimeLabel"] = dates
	this.Data["CurrentData"] = data
	this.Data["BeforeData"] = data1
	this.Data["typeList"] = typeList
	this.Data["typeId"] = typeId
	this.Data["startDate"] = startDate
	this.display()
}

// 渠道数据
func (this *StatisticsController) Channel() {
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	packageId := this.GetString("package_id")
	aliasId := this.GetString("alias_id")
	aliasName := this.GetString("alias_name")
	if page < 1 {
		page = 1
	}
	if startDate == "" && endDate == "" {
		// 默认近十天数据
		today := bson.Now()
		startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -9).Format("2006-01-02"))
		endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
	}
	m := service.FindByDate1(startDate, endDate, "date", "date")
	if packageId != "" && packageId != "0" {
		if packageId == "-" {
			m["channel"] = ""
		} else {
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
		}
	}
	// 渠道别名
	if aliasId != "" && aliasId != "0" {
		if aliasId == "-" {
			m["channel1"] = ""
		} else {
			if strings.Contains(aliasId, ",") {
				palkage_arr := strings.Split(aliasId, ",")
				temp_arr := make([]string, 0)
				for _, item := range palkage_arr {
					if item != "" && item != "0" && item != "-" {
						temp_arr = append(temp_arr, item)
					}
				}
				m["channel1"] = bson.M{"$in": temp_arr}
			} else {
				m["channel1"] = aliasId
			}
		}
	}
	if aliasName != "" {
		m["channel1"] = bson.M{"$regex": aliasName, "$options": "i"}
	}
	count, _ := service.StatisticsService.GetChannelTotal(m)
	list, _ := service.StatisticsService.GetChannelList(page, this.pageSize, m)

	packageList, _ := service.ChannelService.GetChannelAll()
	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()

	this.Data["pageTitle"] = "渠道数据(旧)"
	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.Channel", "package_id", packageId, "alias_id", aliasId, "alias_name", aliasName, "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["packageId"] = packageId
	this.Data["packageList"] = packageList
	this.Data["aliasId"] = aliasId
	this.Data["packageAliasList"] = packageAliasList
	this.Data["aliasName"] = aliasName
	this.display()
}

// 房间数据
func (this *StatisticsController) Room() {
	status, _ := this.GetInt("status")
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	playerid, _ := this.GetInt("player_id")
	tabid, _ := this.GetInt("tabid")
	chatid, _ := this.GetInt("chat_id")
	packageId := this.GetString("package_id")
	aliasId := this.GetString("alias_id")
	aliasName := this.GetString("alias_name")
	numberid, _ := this.GetInt("numbergame_id")
	if page < 1 {
		page = 1
	}
	if startDate == "" && endDate == "" {
		// 默认当天
		today := bson.Now()
		startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -29).Format("2006-01-02"))
		endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
	}
	if status == 0 {
		my_data := this.GetSession("my_select_pakeageid")
		if my_data != nil && packageId == "" {
			if packid, ok := my_data.(string); ok {
				packageId = packid
			}
		}
		status = 1
	}
	playerList := map[int]string{
		0: "全部玩家",
		1: "新玩家",
		2: "老玩家",
	}

	chatType := map[int]string{
		0: "总局数",
		1: "玩家数量",
		4: "真人数量",
		2: "人均局数",
		3: "人均时长",
		5: "对战房局数",
	}
	m := service.FindByDate1(startDate, endDate, "date", "date")
	isChannel := false
	// if packageId != "" && packageId != "0" {
	// 	if packageId == "-" {
	// 		m["channel"] = ""
	// 	} else {
	// 		if strings.Contains(packageId, ",") {
	// 			palkage_arr := strings.Split(packageId, ",")
	// 			temp_arr := make([]string, 0)
	// 			for _, item := range palkage_arr {
	// 				if item != "" && item != "0" && item != "-" {
	// 					temp_arr = append(temp_arr, item)
	// 				}
	// 			}
	// 			m["channel"] = bson.M{"$in": temp_arr}
	// 		} else {
	// 			m["channel"] = packageId
	// 		}
	// 		isChannel = true
	// 	}
	// }
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
		isChannel = true
		// 存储缓存
		this.SetSession("my_select_pakeageid", packageId)
	} else {
		// 清除缓存
		this.DelSession("my_select_pakeageid")
	}
	// 渠道别名
	if aliasId != "" && aliasId != "0" {
		if aliasId == "-" {
			m["channel1"] = ""
		} else {
			if strings.Contains(aliasId, ",") {
				palkage_arr := strings.Split(aliasId, ",")
				temp_arr := make([]string, 0)
				for _, item := range palkage_arr {
					if item != "" && item != "0" && item != "-" {
						temp_arr = append(temp_arr, item)
					}
				}
				m["channel1"] = bson.M{"$in": temp_arr}
			} else {
				m["channel1"] = aliasId
			}
			isChannel = true
		}
	}
	if aliasName != "" {
		m["channel1"] = bson.M{"$regex": aliasName, "$options": "i"}
		isChannel = true
	}
	if playerid != 0 {
		m["player_types"] = playerid
	}
	if numberid != 0 {
		m["number_types"] = numberid
	}
	if tabid == 0 {
		// 折线图
		dates := make([]interface{}, 0)  // 日期
		data := make([]interface{}, 0)   // TP
		data1 := make([]interface{}, 0)  // Rummy
		data2 := make([]interface{}, 0)  // lhd
		data3 := make([]interface{}, 0)  // 7up
		data4 := make([]interface{}, 0)  // AK47
		data5 := make([]interface{}, 0)  // joker
		data6 := make([]interface{}, 0)  // crash
		data7 := make([]interface{}, 0)  // ab
		data8 := make([]interface{}, 0)  // 彩票
		data9 := make([]interface{}, 0)  // 飞机
		data10 := make([]interface{}, 0) // Slots
		data11 := make([]interface{}, 0) // 真人视讯
		data12 := make([]interface{}, 0) // Rummy双人
		data13 := make([]interface{}, 0) // 红黑大战
		data14 := make([]interface{}, 0) // TP2
		list, _ := service.StatisticsService.GetRoomDataList(page, -1, m, isChannel)
		if len(list) > 0 {
			sort.Slice(list, func(i, j int) bool {
				return list[i].Id < list[j].Id
			})
			for _, item := range list {
				// 格式化时间并输出字符串
				strTime := item.SDate.Format("2006-01-02")
				dates = append(dates, strTime)
				if chatid == 1 {
					// 玩家数量
					data = append(data, item.TPNumber)
					data1 = append(data1, item.RummyNumber)
					data2 = append(data2, item.LHDNumber)
					data3 = append(data3, item.UPNumber)
					data4 = append(data4, item.AKNumber)
					data5 = append(data5, item.JokerNumber)
					data6 = append(data6, item.CrashNumber)
					data7 = append(data7, item.ABNumber)
					data8 = append(data8, item.CPNumber)
					data9 = append(data9, item.FJNumber)
					data10 = append(data10, item.SlotsNumber)
					data11 = append(data11, item.ZRSXNumber)
					data12 = append(data12, item.RMTwoNumber)
					data13 = append(data13, item.RBNumber)
					data14 = append(data14, item.TP2Number)
				} else if chatid == 2 {
					// 人均局数
					data = append(data, fmt.Sprintf("%.2f", item.TPPerCapita))
					data1 = append(data1, fmt.Sprintf("%.2f", item.RummyPerCapita))
					data2 = append(data2, fmt.Sprintf("%.2f", item.LHDPerCapita))
					data3 = append(data3, fmt.Sprintf("%.2f", item.UPPerCapita))
					data4 = append(data4, fmt.Sprintf("%.2f", item.AKPerCapita))
					data5 = append(data5, fmt.Sprintf("%.2f", item.JokerPerCapita))
					data6 = append(data6, fmt.Sprintf("%.2f", item.CrashPerCapita))
					data7 = append(data7, fmt.Sprintf("%.2f", item.ABPerCapita))
					data8 = append(data8, fmt.Sprintf("%.2f", item.CPPerCapita))
					data9 = append(data9, fmt.Sprintf("%.2f", item.FJPerCapita))
					data10 = append(data10, fmt.Sprintf("%.2f", item.SlotsPerCapita))
					data11 = append(data11, fmt.Sprintf("%.2f", item.ZRSXPerCapita))
					data12 = append(data12, fmt.Sprintf("%.2f", item.RMTwoPerCapita))
					data13 = append(data13, fmt.Sprintf("%.2f", item.RBPerCapita))
					data14 = append(data14, fmt.Sprintf("%.2f", item.TP2PerCapita))
				} else if chatid == 3 {
					// 人均时长
					data = append(data, fmt.Sprintf("%.2f", item.TPTime))
					data1 = append(data1, fmt.Sprintf("%.2f", item.RummyTime))
					data2 = append(data2, fmt.Sprintf("%.2f", item.LHDTime))
					data3 = append(data3, fmt.Sprintf("%.2f", item.UPTime))
					data4 = append(data4, fmt.Sprintf("%.2f", item.AKTime))
					data5 = append(data5, fmt.Sprintf("%.2f", item.JokerTime))
					data6 = append(data6, fmt.Sprintf("%.2f", item.CrashTime))
					data7 = append(data7, fmt.Sprintf("%.2f", item.ABTime))
					data8 = append(data8, fmt.Sprintf("%.2f", item.CPTime))
					data9 = append(data9, fmt.Sprintf("%.2f", item.FJTime))
					data12 = append(data12, fmt.Sprintf("%.2f", item.RMTwoTime))
					data13 = append(data13, fmt.Sprintf("%.2f", item.RBTime))
					data14 = append(data14, fmt.Sprintf("%.2f", item.TP2Time))
				} else if chatid == 4 {
					// 真人玩家
					data = append(data, item.TPGameRealNumber)
					data1 = append(data1, item.RMGameRealNumber)
					// data2 = append(data2, fmt.Sprintf("%.2f", item.LHDTime))
					// data3 = append(data3, fmt.Sprintf("%.2f", item.UPTime))
					data4 = append(data4, item.AKGameRealNumber)
					data5 = append(data5, item.JokerGameRealNumber)
					// data6 = append(data6, fmt.Sprintf("%.2f", item.CrashTime))
				} else if chatid == 5 {
					// 对战房局数
					data = append(data, item.TPBattleGame)
					data1 = append(data1, item.RMBattleGame)
					data7 = append(data7, item.ABBattleGame)
				} else {
					// 总局数
					data = append(data, item.TPGameNumber)
					data1 = append(data1, item.RummyGameNumber)
					data2 = append(data2, item.LHDGameNumber)
					data3 = append(data3, item.UPGameNumber)
					data4 = append(data4, item.AKGameNumber)
					data5 = append(data5, item.JokerGameNumber)
					data6 = append(data6, item.CrashGameNumber)
					data7 = append(data7, item.ABGameNumber)
					data8 = append(data8, item.CPGameNumber)
					data9 = append(data9, item.FJGameNumber)
					data10 = append(data10, item.SlotsGameNumber)
					data11 = append(data11, item.ZRSXGameNumber)
					data12 = append(data12, item.RMTwoGameNumber)
					data13 = append(data13, item.RBGameNumber)
					data14 = append(data14, item.TP2GameNumber)
				}

			}
		}
		this.Data["TimedLabel"] = dates
		this.Data["TP"] = data
		this.Data["Rummy"] = data1
		this.Data["LHD"] = data2
		this.Data["UP"] = data3
		this.Data["AK47"] = data4
		this.Data["Joker"] = data5
		this.Data["Crash"] = data6
		this.Data["AB"] = data7
		this.Data["CP"] = data8
		this.Data["FJ"] = data9
		this.Data["Slots"] = data10
		this.Data["ZRSX"] = data11
		this.Data["RB"] = data13
		this.Data["RMTWO"] = data12
		this.Data["TPTWO"] = data14
		this.Data["tabid"] = tabid
	} else {
		// 表格图
		count, _ := service.StatisticsService.GetRoomDataTotal(m, isChannel)
		list, _ := service.StatisticsService.GetRoomDataList(page, this.pageSize, m, isChannel)

		this.Data["count"] = count
		this.Data["list"] = list
		this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.Room", "status", status, "package_id", packageId, "alias_id", aliasId, "alias_name", aliasName, "numbergame_id", numberid, "player_id", playerid, "tabid", tabid, "start_date", startDate, "end_date", endDate), true).ToString()

	}
	packageList, _ := service.ChannelService.GetChannelAll()
	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()
	numberGameList := map[int]string{
		0: "全部",
		1: "1局",
		2: "2-5局",
		3: "6-10局",
		4: "11-20局",
		5: "21-30局",
		6: "31-50局",
		7: "51局及以上",
	}
	this.Data["pageTitle"] = "房间数据"
	this.Data["status"] = status
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["playerid"] = playerid
	this.Data["playerList"] = playerList
	this.Data["chatType"] = chatType
	this.Data["chatid"] = chatid
	this.Data["tabid"] = tabid
	this.Data["packageId"] = packageId
	this.Data["packageList"] = packageList
	this.Data["aliasId"] = aliasId
	this.Data["packageAliasList"] = packageAliasList
	this.Data["aliasName"] = aliasName
	this.Data["numberGameList"] = numberGameList
	this.Data["numberid"] = numberid
	this.display()
}

func (this *StatisticsController) PointList() {
	status, _ := this.GetInt("status")
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	tabid, _ := this.GetInt("tabid")
	packageId := this.GetString("package_id")
	aliasId := this.GetString("alias_id")
	aliasName := this.GetString("alias_name")
	if page < 1 {
		page = 1
	}
	if startDate == "" && endDate == "" {
		// 默认近十天数据
		today := bson.Now()
		startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -9).Format("2006-01-02"))
		endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
	}
	if status == 0 {
		my_data := this.GetSession("my_select_pakeageid")
		if my_data != nil && packageId == "" {
			if packid, ok := my_data.(string); ok {
				packageId = packid
			}
		}
		status = 1
	}
	m := service.FindByDate1(startDate, endDate, "date", "date")
	count := int64(0)
	isChannel := false
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
		isChannel = true
		// 存储缓存
		this.SetSession("my_select_pakeageid", packageId)
	} else {
		// 清除缓存
		this.DelSession("my_select_pakeageid")
	}
	// 渠道别名
	if aliasId != "" && aliasId != "0" {
		if aliasId == "-" {
			m["channel1"] = ""
		} else {
			if strings.Contains(aliasId, ",") {
				palkage_arr := strings.Split(aliasId, ",")
				temp_arr := make([]string, 0)
				for _, item := range palkage_arr {
					if item != "" && item != "0" && item != "-" {
						temp_arr = append(temp_arr, item)
					}
				}
				m["channel1"] = bson.M{"$in": temp_arr}
			} else {
				m["channel1"] = aliasId
			}
			isChannel = true
		}
	}
	if aliasName != "" {
		m["channel1"] = bson.M{"$regex": aliasName, "$options": "i"}
		isChannel = true
	}

	if tabid == 1 {
		count, _ = service.StatisticsService.GetGameNumberAnalysisTotal(m, isChannel)
		list, _ := service.StatisticsService.GetGameNumberAnalysisList(page, this.pageSize, m, isChannel)
		this.Data["list"] = list
	} else if tabid == 2 {
		count, _ = service.StatisticsService.GetPlaytimeTotal(m, isChannel)
		list, _ := service.StatisticsService.GetPlaytimeList(page, this.pageSize, m, isChannel)
		this.Data["list"] = list
	} else {
		// 行为分析
		count, _ = service.StatisticsService.GetPointListTotal(m, isChannel)
		list, _ := service.StatisticsService.GetPointList(page, this.pageSize, m, isChannel)
		this.Data["list"] = list
	}
	packageList, _ := service.ChannelService.GetChannelAll()
	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()

	this.Data["pageTitle"] = "新用户埋点统计"
	this.Data["count"] = count
	this.Data["tabid"] = tabid
	this.Data["status"] = status
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.PointList", "status", status, "tabid", tabid, "package_id", packageId, "alias_id", aliasId, "alias_name", aliasName, "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["packageId"] = packageId
	this.Data["packageList"] = packageList
	this.Data["aliasId"] = aliasId
	this.Data["packageAliasList"] = packageAliasList
	this.Data["aliasName"] = aliasName
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "pointlistopt")
	this.display()
}

func (this *StatisticsController) GameCharts() {
	pid, _ := this.GetInt("id")
	channel := this.GetString("channel")
	dtypeId, _ := this.GetInt("dtype")
	if pid == 0 {
		this.checkError(errors.New("ID不能为空"))
	}
	if channel == "" {
		this.checkError(errors.New("暂不支持全部查看"))
	}
	info, _ := service.StatisticsService.GetPointById(int64(pid), channel)
	dtype := map[int]string{
		0: "第一局玩的游戏",
		1: "第二局玩的游戏",
		2: "人均局数",
		3: "退出率",
		4: "充值率",
		5: "拉起订单率",
		6: "击败率",
		7: "大赢率",
	}
	data := make([]int64, 0)
	switch dtypeId {
	case 0:
		if info.FirstGames != nil {
			data = []int64{
				info.FirstGames.TpNumber,
				info.FirstGames.RmNumber,
				info.FirstGames.AkNumber,
				info.FirstGames.JokerNumber,
				info.FirstGames.LHDNumber,
				info.FirstGames.UPNumber,
				info.FirstGames.CrashNumber,
			}
		}
	case 1:
		if info.SecondGames != nil {
			data = []int64{
				info.SecondGames.TpNumber,
				info.SecondGames.RmNumber,
				info.SecondGames.AkNumber,
				info.SecondGames.JokerNumber,
				info.SecondGames.LHDNumber,
				info.SecondGames.UPNumber,
				info.SecondGames.CrashNumber,
			}
		}
	default:
	}
	this.Data["pageTitle"] = "游戏图表数据"
	this.Data["gamedata"] = data
	this.Data["pid"] = pid
	this.Data["dtype"] = dtype
	this.Data["dtypeId"] = dtypeId
	this.display()
}

func (this *StatisticsController) BugList() {
	page, _ := this.GetInt("page")
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	userid := this.GetString("userid")
	tabid, _ := this.GetInt("tabid")
	if page < 1 {
		page = 1
	}
	count := int64(0)
	m := bson.M{}
	if tabid == 1 {
		// 领取记录
		if userid != "" {
			m["userid"] = userid
		}
		list, _ := service.StatisticsService.GetBugRecordList(page, this.pageSize, m)
		count, _ = service.StatisticsService.GetBugRecordTotal(m)
		this.Data["list"] = list
		this.Data["count"] = count
	} else {
		// 活动数据
		if startDate == "" && endDate == "" {
			// 默认近十天数据
			today := bson.Now()
			startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -9).Format("2006-01-02"))
			endDate = fmt.Sprintf("%s", today.AddDate(0, 0, 0).Format("2006-01-02"))
		}
		m = service.FindByDate1(startDate, endDate, "date", "date")
		list, _ := service.StatisticsService.GetBugList(page, this.pageSize, m)
		count, _ = service.StatisticsService.GetBugListTotal(m)
		this.Data["list"] = list
		this.Data["count"] = count
	}

	this.Data["pageTitle"] = "Bug统计"
	this.Data["userid"] = userid
	if tabid == 1 {
		this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.BugList", "tabid", tabid, "userid", userid), true).ToString()
	} else {
		this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.BugList", "tabid", tabid, "start_date", startDate, "end_date", endDate), true).ToString()
	}
	this.Data["tabid"] = tabid
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.display()
}

// 收银台统计
func (this *StatisticsController) Checkstand() {
	page, _ := this.GetInt("page")
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	if page < 1 {
		page = 1
	}
	if startDate == "" && endDate == "" {
		// 默认近二十天数据
		today := bson.Now()
		startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -19).Format("2006-01-02"))
		endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
	}
	if startDate == "" || endDate == "" {
		this.checkError(errors.New("开始时间或结束时间不能为空！"))
	}
	count := 0
	b := make([]string, 0)
	if startDate != "" && endDate != "" {
		b = append(b, startDate)
		b = append(b, endDate)
	}

	// 加载时获取安卓评分
	result, _ := service.AdGetCheckstand(b)
	beego.Trace("result: ", result)
	var list []entity.StatDay
	if result != nil {
		if len(result.Data) > 0 {
			filtered := result.Data
			count = len(filtered)
			startIndex := (page - 1) * this.pageSize
			endIndex := (page) * this.pageSize
			if endIndex > count {
				endIndex = count
			}
			list = filtered[startIndex:endIndex]
		}
		if len(list) > 0 {
			// 查询充值请求人数

			for i, item := range list {
				uniqueUserIDs := make(map[string]bool)
				m := service.FindByDate(item.Day, item.Day, "ctime", "ctime")
				orderlist, _ := service.PayService.GetByPayUser(m)
				for _, v := range orderlist {
					uniqueUserIDs[v.Userid] = true
				}
				var result []string
				for id := range uniqueUserIDs {
					result = append(result, id)
				}
				reqCount := int64(len(result))
				item.PayRequest = reqCount
				item.DeskUserCountRatio = service.ComputeFloat(int64(item.DeskUserCount), reqCount) * 100
				item.DeskTimeCountRatio = service.ComputeFloat(int64(item.DeskTimeCount), reqCount) * 100
				item.PayTotalCountRatio = service.ComputeFloat(int64(item.PayTotalCount), reqCount) * 100
				item.PaySuccessTimeCountRatio = service.ComputeFloat(int64(item.PaySuccessTimeCount), reqCount) * 100
				item.PaySuccessUserCountRatio = service.ComputeFloat(int64(item.PaySuccessUserCount), reqCount) * 100
				list[i] = item
			}

		}
	}
	this.Data["pageTitle"] = "收银台统计"
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.Checkstand", "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["count"] = count
	this.Data["list"] = list
	this.display()
}

/*
商品购买统计相关
*/
func (this *StatisticsController) GoodsBuy() {
	page, _ := this.GetInt("page")
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	playerid, _ := this.GetInt("player_id")
	if page < 1 {
		page = 1
	}
	if startDate == "" && endDate == "" {
		// 默认近十天数据
		today := bson.Now()
		startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -9).Format("2006-01-02"))
		endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
	}
	m := service.FindByDate1(startDate, endDate, "date", "date")
	if playerid != 0 {
		m["player_types"] = playerid
	}
	list, _ := service.StatisticsService.GetGoodsBuyList(page, this.pageSize, m)
	count, _ := service.StatisticsService.GetGoodsBuyTotal(m)
	playerList := map[int]string{
		0: "全部玩家",
		1: "新玩家",
		2: "老玩家",
	}
	this.Data["pageTitle"] = "商品购买统计"
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.GoodsBuy", "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["playerid"] = playerid
	this.Data["playerList"] = playerList
	this.display()
}

/*
AD上报统计
*/
func (this *StatisticsController) AdReport() {
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	packageId := this.GetString("package_id")
	aliasId := this.GetString("alias_id")
	aliasName := this.GetString("alias_name")
	if page < 1 {
		page = 1
	}
	if startDate == "" && endDate == "" {
		// 默认近十天数据
		today := bson.Now()
		startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -9).Format("2006-01-02"))
		endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
	}
	m := service.FindByDate1(startDate, endDate, "date", "date")
	if packageId != "" && packageId != "0" {
		if packageId == "-" {
			m["channel"] = ""
		} else {
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
		}
	}
	// 渠道别名
	if aliasId != "" && aliasId != "0" {
		if aliasId == "-" {
			m["channel1"] = ""
		} else {
			if strings.Contains(aliasId, ",") {
				palkage_arr := strings.Split(aliasId, ",")
				temp_arr := make([]string, 0)
				for _, item := range palkage_arr {
					if item != "" && item != "0" && item != "-" {
						temp_arr = append(temp_arr, item)
					}
				}
				m["channel1"] = bson.M{"$in": temp_arr}
			} else {
				m["channel1"] = aliasId
			}
		}
	}
	if aliasName != "" {
		m["channel1"] = bson.M{"$regex": aliasName, "$options": "i"}
	}

	list, _ := service.StatisticsService.GetAdReportList(page, this.pageSize, m)
	count, _ := service.StatisticsService.GetAdReportTotal(m)
	packageList, _ := service.ChannelService.GetChannelAll()
	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()

	this.Data["pageTitle"] = "AD上报统计"
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.AdReport", "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["packageId"] = packageId
	this.Data["packageList"] = packageList
	this.Data["aliasId"] = aliasId
	this.Data["packageAliasList"] = packageAliasList
	this.Data["aliasName"] = aliasName
	this.display()
}

// 房间资源
func (this *StatisticsController) RoomResources() {
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	tabid, _ := this.GetInt("tabid")
	if page < 1 {
		page = 1
	}
	if startDate == "" && endDate == "" {
		// 默认近十天数据
		today := bson.Now()
		startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -9).Format("2006-01-02"))
		endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
	}
	m := service.FindByDate1(startDate, endDate, "ctime", "ctime")
	// recordList := map[int]string{
	// 	1: "TP",
	// 	2: "DRAGON TIGER",
	// 	3: "7UPDOWN",
	// 	4: "RUMMY",
	// 	5: "AK47",
	// 	6: "JOKER",
	// 	7: "CRASH",
	// }
	switch tabid {
	case 1:
		m["gtype"] = 1
	case 2:
		m["gtype"] = 4
	case 3:
		m["gtype"] = 5
	case 4:
		m["gtype"] = 6
	case 5:
		m["gtype"] = 2
	case 6:
		m["gtype"] = 3
	case 7:
		m["gtype"] = 7
	case 8:
		m["gtype"] = 8
	case 9:
		m["gtype"] = 9
	case 10:
		m["gtype"] = 10
	case 11:
		m["gtype"] = 11
	case 12:
		m["gtype"] = 12
	case 13:
		m["gtype"] = 13
	}
	// m := bson.M{}
	list, _ := service.StatisticsService.GetRoomResourcesList(page, this.pageSize, m)
	count, _ := service.StatisticsService.GetRoomResourcesTotal(m)

	this.Data["pageTitle"] = "正常库存"
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.RoomResources", "tabid", tabid, "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["tabid"] = tabid
	this.display()
}

// 短信统计
func (this *StatisticsController) SMS() {
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	if startDate == "" && endDate == "" {
		// 默认近十天数据
		today := bson.Now()
		startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -9).Format("2006-01-02"))
		endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
	}
	m := service.FindByDate1(startDate, endDate, "date", "date")

	list, _ := service.StatisticsService.GetSMSList(page, this.pageSize, m)
	count, _ := service.StatisticsService.GetSMSTotal(m)

	this.Data["pageTitle"] = "短信统计"
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.SMS", "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["count"] = count
	this.Data["list"] = list
	this.display()
}

// 充值来源
func (this *StatisticsController) PaySource() {
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	if startDate == "" && endDate == "" {
		// 默认近十天数据
		today := bson.Now()
		startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -9).Format("2006-01-02"))
		endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
	}
	m := service.FindByDate1(startDate, endDate, "date", "date")

	list, _ := service.StatisticsService.GetPaySourceList(page, this.pageSize, m)
	count, _ := service.StatisticsService.GetPaySourceTotal(m)

	this.Data["pageTitle"] = "充值来源"
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.PaySource", "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["count"] = count
	this.Data["list"] = list
	this.display()
}

// 对战房埋点记录
func (this *StatisticsController) BattleRoom() {
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	if startDate == "" && endDate == "" {
		// 默认近十天数据
		today := bson.Now()
		startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -9).Format("2006-01-02"))
		endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
	}
	m := service.FindByDate1(startDate, endDate, "date", "date")

	list, _ := service.StatisticsService.GetBattleRoomList(page, this.pageSize, m)
	count, _ := service.StatisticsService.GetBattleRoomTotal(m)

	this.Data["pageTitle"] = "对战房埋点统计"
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.BattleRoom", "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["count"] = count
	this.Data["list"] = list
	this.display()
}

// 大R分析
func (this *StatisticsController) UserAnalysis() {
	// page, _ := strconv.Atoi(this.GetString("page"))
	paystatr := this.GetString("pay_statr")
	payend := this.GetString("pay_end")
	withdrawstatr := this.GetString("withdraw_statr")
	withdrawend := this.GetString("withdraw_end")
	gainstatr := this.GetString("gain_statr")
	gainend := this.GetString("gain_end")

	list := make([]*entity.UserAnalysisInfo, 0)
	count := 0
	pay_statr, _ := strconv.Atoi(paystatr)
	pay_end, _ := strconv.Atoi(payend)
	withdraw_statr, _ := strconv.Atoi(withdrawstatr)
	withdraw_end, _ := strconv.Atoi(withdrawend)
	gain_statr, _ := strconv.Atoi(gainstatr)
	gain_end, _ := strconv.Atoi(gainend)

	if pay_statr != 0 || pay_end != 0 || withdraw_statr != 0 || withdraw_end != 0 || gain_statr != 0 || gain_end != 0 {
		list, _ = service.StatisticsService.GetUserAnalysis(pay_statr, pay_end, withdraw_statr, withdraw_end, gain_statr, gain_end)
		count = len(list)
	}

	this.Data["pageTitle"] = "大R分析"
	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["pay_statr"] = paystatr
	this.Data["pay_end"] = payend
	this.Data["withdraw_statr"] = withdrawstatr
	this.Data["withdraw_end"] = withdrawend
	this.Data["gain_statr"] = gainstatr
	this.Data["gain_end"] = gainend
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "useranalysisopt")
	this.display()
}

func (this *StatisticsController) UserAnalysisExport() {
	paystatr := this.GetString("pay_statr")
	payend := this.GetString("pay_end")
	withdrawstatr := this.GetString("withdraw_statr")
	withdrawend := this.GetString("withdraw_end")
	gainstatr := this.GetString("gain_statr")
	gainend := this.GetString("gain_end")

	// 解析查询参数
	pay_statr, _ := strconv.Atoi(paystatr)
	pay_end, _ := strconv.Atoi(payend)
	withdraw_statr, _ := strconv.Atoi(withdrawstatr)
	withdraw_end, _ := strconv.Atoi(withdrawend)
	gain_statr, _ := strconv.Atoi(gainstatr)
	gain_end, _ := strconv.Atoi(gainend)

	// 获取用户分析数据
	list, _ := service.StatisticsService.GetUserAnalysis(pay_statr, pay_end, withdraw_statr, withdraw_end, gain_statr, gain_end)
	count := len(list)

	// 检查是否有数据可导出
	if count <= 0 {
		this.showMsg("未查询到可以导出的数据！", MSG_ERR, "")
		return
	}

	file := excelize.NewFile()
	sheet := "Sheet1"
	//设置表格头
	file.SetCellValue(sheet, "A1", "用户ID")
	file.SetCellValue(sheet, "B1", "渠道")
	file.SetCellValue(sheet, "C1", "渠道别名")
	file.SetCellValue(sheet, "D1", "注册日期")
	file.SetCellValue(sheet, "E1", "最后登录日期")
	file.SetCellValue(sheet, "F1", "存活时长")
	file.SetCellValue(sheet, "G1", "活跃天数")
	file.SetCellValue(sheet, "H1", "关联账号")
	file.SetCellValue(sheet, "I1", "首充日期")
	file.SetCellValue(sheet, "J1", "首充金额")
	file.SetCellValue(sheet, "K1", "总充值金额")
	file.SetCellValue(sheet, "L1", "总充单数")
	file.SetCellValue(sheet, "M1", "充值单均价")
	file.SetCellValue(sheet, "N1", "总提现金额")
	file.SetCellValue(sheet, "O1", "总提现单数")
	file.SetCellValue(sheet, "P1", "提现单均价")
	file.SetCellValue(sheet, "Q1", "玩的最多的游戏")
	file.SetCellValue(sheet, "R1", "玩最多游戏的局数")
	file.SetCellValue(sheet, "S1", "玩第二多的游戏")
	file.SetCellValue(sheet, "T1", "玩第二多游戏的局数")

	// 写入数据
	for idx, info := range list {
		roundedNum := fmt.Sprintf("%.2f", info.PayAvg)
		roundedNum1 := fmt.Sprintf("%.2f", info.WithdrawAvg)
		row := idx + 2 // Excel 行号从 1 开始，因此要加 2
		file.SetCellValue(sheet, "A"+strconv.Itoa(row), info.UserId)
		file.SetCellValue(sheet, "B"+strconv.Itoa(row), info.Channel)
		file.SetCellValue(sheet, "C"+strconv.Itoa(row), info.Channel1)
		file.SetCellValue(sheet, "D"+strconv.Itoa(row), info.RegistTime)
		file.SetCellValue(sheet, "E"+strconv.Itoa(row), info.LastLoginTime)
		file.SetCellValue(sheet, "F"+strconv.Itoa(row), info.ActiveTime)
		file.SetCellValue(sheet, "G"+strconv.Itoa(row), info.ActiveDay)
		file.SetCellValue(sheet, "H"+strconv.Itoa(row), info.AssociatedCount)
		file.SetCellValue(sheet, "I"+strconv.Itoa(row), info.FirstPayDate)
		file.SetCellValue(sheet, "J"+strconv.Itoa(row), fmt.Sprintf("%.2f", info.FirstPayAmount))
		file.SetCellValue(sheet, "K"+strconv.Itoa(row), fmt.Sprintf("%.2f", info.PayAmount))
		file.SetCellValue(sheet, "L"+strconv.Itoa(row), info.PayCount)
		file.SetCellValue(sheet, "M"+strconv.Itoa(row), roundedNum)
		file.SetCellValue(sheet, "N"+strconv.Itoa(row), fmt.Sprintf("%.2f", info.WithdrawAmount))
		file.SetCellValue(sheet, "O"+strconv.Itoa(row), info.WithdrawCount)
		file.SetCellValue(sheet, "P"+strconv.Itoa(row), roundedNum1)
		file.SetCellValue(sheet, "Q"+strconv.Itoa(row), info.GameName)
		file.SetCellValue(sheet, "R"+strconv.Itoa(row), info.GameCount)
		file.SetCellValue(sheet, "S"+strconv.Itoa(row), info.TwoGameName)
		file.SetCellValue(sheet, "T"+strconv.Itoa(row), info.TwoGameCount)
	}
	err := file.SaveAs("output.xlsx")
	if err != nil {
		log.Fatal(err)
	}

	//构造文件名称
	fileName := "大R分析_" + time.Now().Format("20060102150405") + ".xlsx"
	// 解决文件名中文乱码问题
	fileName = path.Base(fileName)
	fileName = url.QueryEscape(fileName)
	this.Ctx.Output.Header("Content-type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	this.Ctx.Output.Header("Content-Disposition", "attachment;filename="+fileName)
	this.Ctx.Output.Header("Pragma", "No-cache")
	this.Ctx.Output.Header("Cache-Control", "No-cache")
	this.Ctx.Output.Header("Expires", "0")
	// var buffer bytes.Buffer
	if err := file.Write(this.Ctx.ResponseWriter); err != nil {
		logs.Error(err)
	}
	// r := bytes.NewReader(buffer.Bytes())

	// http.ServeContent(this.Ctx.ResponseWriter, this.Ctx.Request, fileName, time.Now(), r)
}

// 大R分析(新)
func (this *StatisticsController) UserAnalysis2_Deprecate() {
	page, _ := strconv.Atoi(this.GetString("page"))
	paystatr := this.GetString("pay_statr")
	payend := this.GetString("pay_end")
	withdrawstatr := this.GetString("withdraw_statr")
	withdrawend := this.GetString("withdraw_end")
	gainstatr := this.GetString("gain_statr")
	gainend := this.GetString("gain_end")

	list := make([]*entity.UserAnalysisInfo, 0)
	var count int64
	pay_statr, _ := strconv.Atoi(paystatr)
	pay_end, _ := strconv.Atoi(payend)
	withdraw_statr, _ := strconv.Atoi(withdrawstatr)
	withdraw_end, _ := strconv.Atoi(withdrawend)
	gain_statr, _ := strconv.Atoi(gainstatr)
	gain_end, _ := strconv.Atoi(gainend)

	if pay_statr != 0 || pay_end != 0 || withdraw_statr != 0 || withdraw_end != 0 || gain_statr != 0 || gain_end != 0 {
		var err error
		count, list, err = service.StatisticsService.GetUserAnalysisCk(page, this.pageSize, pay_statr, pay_end, withdraw_statr, withdraw_end, gain_statr, gain_end)
		if err != nil {
			beego.Error(err)
		}
	}

	this.Data["pageTitle"] = "大R分析"
	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["pay_statr"] = paystatr
	this.Data["pay_end"] = payend
	this.Data["withdraw_statr"] = withdrawstatr
	this.Data["withdraw_end"] = withdrawend
	this.Data["gain_statr"] = gainstatr
	this.Data["gain_end"] = gainend
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.UserAnalysis2", "pay_statr", paystatr, "pay_end", payend, "withdraw_statr", withdrawstatr, "withdraw_end", withdrawend, "gain_statr", gainstatr, "gain_end", gainend), true).ToString()
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "useranalysis2opt")
	this.display()
}

func getTitle() *[]interface{} {
	var titles = []interface{}{
		"序号", "名称",
	}
	return &titles
}

func (this *StatisticsController) UserLevelCtype() {
	startDate := this.GetString("start_date") // 注册时间1
	endDate := this.GetString("end_date")     // 注册时间2
	packageId := this.GetString("package_id")
	aliasId := this.GetString("alias_id")
	classId := this.GetString("class_id")
	typeId, _ := this.GetInt("typeId")

	today := service.NowTime()
	if startDate == "" {
		// 默认近七天数据
		startDate = today.AddDate(0, 0, -7).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = today.Format("2006-01-02")
	}

	var start, end time.Time
	s, e := service.FindByDate4(startDate, endDate)
	start = *s
	end = *e
	startDate = start.Format("2006-01-02")
	endDate = end.Format("2006-01-02")

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
		} else {
			packageIds = append(packageIds, packageId)
		}
	}

	// 渠道别名
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

	if typeId == 0 {
		list, err := service.StatisticsService.UserLevelCtype(start, end, packageIds, aliasIds, classIds)
		if err != nil {
			this.showMsg(err.Error(), MSG_ERR, "")
		}
		this.Data["list"] = list
	} else if typeId == 1 {
		list, err := service.StatisticsService.UserTransEffect(start, end, packageIds, aliasIds, classIds)
		if err != nil {
			this.showMsg(err.Error(), MSG_ERR, "")
		}
		this.Data["list"] = list
	} else if typeId == 2 {
		list, err := service.StatisticsService.UserLevelFunnelStats(start, end, packageIds)
		if err != nil {
			this.showMsg(err.Error(), MSG_ERR, "")
		}
		this.Data["list"] = list
	} else if typeId == 3 {
		c_channelClass := this.GetStrings("c_channelClass")
		c_channelAlias := this.GetStrings("c_channelAlias")
		c_dates := this.GetStrings("c_dates")
		c_ctype := this.GetStrings("c_ctype")
		c_stats := this.GetStrings("c_stats")
		// 默认选择第一个渠道
		// if len(c_channelAlias) == 0 && len(aliasIds) > 0 {
		// 	c_channelAlias = []string{aliasIds[0]}
		// }

		var ctypes []int
		for _, ctype := range c_ctype {
			// var ctypeI int
			// if _, err := fmt.Sscanf(ctype, "ctype_%d", &ctypeI); err != nil {
			// 	continue
			// }
			ctypeI, err := strconv.Atoi(ctype)
			if err != nil {
				continue
			}
			ctypes = append(ctypes, ctypeI)
		}

		var statsCount, statsRate bool
		for _, stats := range c_stats {
			statsCount = statsCount || stats == "0"
			statsRate = statsRate || stats == "1"
		}

		countStats, rateStats, cls, err := service.StatisticsService.UserLevelStats(start, end, packageIds, c_channelClass, c_channelAlias, c_dates, ctypes, statsCount, statsRate)
		if err != nil {
			this.showMsg(err.Error(), MSG_ERR, "")
		}
		this.Data["countStats"] = countStats
		this.Data["rateStats"] = rateStats
		this.Data["classIds"] = cls
		this.Data["aliasIds"] = aliasIds
		var dates []string
		var begin = start
		for !begin.After(end) {
			dates = append(dates, begin.Format(utils.FORMAT_DATE))
			begin = begin.AddDate(0, 0, 1)
		}
		this.Data["dates"] = dates

		this.Data["c_channelClass"] = c_channelClass
		this.Data["c_channelAlias"] = c_channelAlias
		this.Data["c_dates"] = c_dates
		this.Data["c_ctype"] = c_ctype
		this.Data["c_stats"] = c_stats
	} else if typeId == 4 {
		list, err := service.StatisticsService.UserRegChannelStats(start, end, packageIds)
		if err != nil {
			this.showMsg(err.Error(), MSG_ERR, "")
		}
		this.Data["list"] = list
	}

	packageList, _ := service.ChannelService.GetChannelAll()
	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()
	packageClassAll, _ := service.ChannelService.GetChannelClassAll()

	this.Data["pageTitle"] = "用户分层"
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["packageId"] = packageId
	this.Data["aliasId"] = aliasId
	this.Data["classId"] = classId
	this.Data["packageList"] = packageList
	this.Data["packageAliasList"] = packageAliasList
	this.Data["packageClassAll"] = packageClassAll
	// this.Data["aliasIds"] = aliasIds
	// this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "recoverycycleopt")
	this.Data["typeId"] = typeId

	this.display()
}

func (this *StatisticsController) UserAnalysisExport2() {
	userid := this.GetString("userid")
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	paystatr := this.GetString("pay_statr")
	payend := this.GetString("pay_end")
	withdrawstatr := this.GetString("withdraw_statr")
	withdrawend := this.GetString("withdraw_end")
	gainstatr := this.GetString("gain_statr")
	gainend := this.GetString("gain_end")
	utypeId := this.GetString("utype_id")
	maxRoundsGtype, _ := strconv.Atoi(this.GetString("maxRoundsGtype"))
	maxBetGtype, _ := strconv.Atoi(this.GetString("maxBetGtype"))
	typeId, _ := strconv.Atoi(this.GetString("typeId"))
	density, _ := strconv.Atoi(this.GetString("density")) // typeId=2横轴密度(分钟)
	if density <= 0 {
		density = 10
	}

	today := service.NowTime()
	if startDate == "" && endDate == "" {
		// 默认近七天数据
		endDate = today.Format("2006-01-02")
		startDate = today.AddDate(0, 0, -6).Format("2006-01-02")
	}

	startTime, endTime := service.FindByDate4(startDate, endDate)

	pay_statr, _ := strconv.Atoi(paystatr)
	pay_end, _ := strconv.Atoi(payend)
	moneyRange := []int64{int64(pay_statr), int64(pay_end)}
	withdraw_statr, _ := strconv.Atoi(withdrawstatr)
	withdraw_end, _ := strconv.Atoi(withdrawend)
	withdrawRange := []int64{int64(withdraw_statr), int64(withdraw_end)}
	gain_statr, _ := strconv.Atoi(gainstatr)
	gain_end, _ := strconv.Atoi(gainend)
	profitRange := []int64{int64(gain_statr), int64(gain_end)}

	var utypeIds []int
	if utypeId != "" && utypeId != "0" && utypeId != "-" && utypeId != "<nil>" {
		ids := strings.Split(utypeId, ",")
		for _, _id := range ids {
			id, err := strconv.Atoi(_id)
			if err != nil || id == 0 {
				continue
			}
			utypeIds = append(utypeIds, id)
		}
	} else {
		utypeId = ""
	}

	var count int
	var headers []string
	var rows [][]any
	if typeId == 0 {
		// 获取用户分析数据
		total, datas, err := service.StatisticsService.GetUserAnalysisCk(1, 100000, pay_statr, pay_end, withdraw_statr, withdraw_end, gain_statr, gain_end)
		if err != nil {
			this.showMsg(err.Error(), MSG_ERR, "")
		}
		count = int(total)
		headers = []string{"用户ID", "渠道", "渠道别名", "注册日期", "最后登录日期", "存活时长", "活跃天数", "关联账号", "首充日期", "首充金额", "总充值金额", "总充单数", "充值单均价", "总提现金额", "总提现单数", "提现单均价", "玩的最多的游戏", "玩最多游戏的局数", "玩第二多的游戏", "玩第二多游戏的局数"}
		for _, info := range datas {
			row := []any{
				info.UserId,
				info.Channel,
				info.Channel1,
				info.RegistTime,
				info.LastLoginTime,
				info.ActiveTime,
				info.ActiveDay,
				info.AssociatedCount,
				info.FirstPayDate,
				fmt.Sprintf("%.2f", info.FirstPayAmount),
				fmt.Sprintf("%.2f", info.PayAmount),
				info.PayCount,
				fmt.Sprintf("%.2f", info.PayAvg),
				fmt.Sprintf("%.2f", info.WithdrawAmount),
				info.WithdrawCount,
				fmt.Sprintf("%.2f", info.WithdrawAvg),
				info.GameName,
				info.GameCount,
				info.TwoGameName,
				info.TwoGameCount,
			}
			rows = append(rows, row)
		}
	} else if typeId == 1 {
		// 玩家分析总表
		total, datas, err := service.StatisticsService.PlayerAnalysis(1, 100000, startTime, endTime, userid, utypeIds, moneyRange, withdrawRange, profitRange, int32(maxRoundsGtype), int32(maxBetGtype))
		if err != nil {
			this.showMsg(err.Error(), MSG_ERR, "")
		}
		count = total
		headers = []string{"用户ID", "用户类型", "渠道类", "渠道别名", "注册日期", "最后登录日期", "流失天数", "存活时长", "活跃天数", "总返奖率", "盈利金额", "关联账号", "", "首充日期", "首充金额", "总充金额", "总充单数", "充值单均价", "充值成功率", "首提日期", "首提金额", "总提金额", "总提单数", "提现单均价", "提现成功率", "首充效率", "首提距首充间隔", "每次充值前携带金额均值", "充值次数/提现次数", "", "单笔最大充值金额/日期", "单笔最大提现金额/日期", "单日累计最大充值金额/日期/成功率", "单日累计最大提现金额/日期/成功率", "单日最高充值次数/日期/成功率", "单日最高提现次数/日期/成功率", "", "玩最多的游戏/局数/局均码量/返奖率", "打码最多的游戏/局数/局均码量/返奖率", "玩第二多的游戏/局数/局均码量/返奖率", "打码第二多的游戏/局数/局均码量/返奖率", "玩第三多的游戏/局数/局均码量/返奖率", "打码第三多的游戏/局数/局均码量/返奖率", "玩的多的游戏赢局局均赢钱金额", "玩的多的游戏输局局均输钱金额", "打码多的游戏赢局局均赢钱金额", "打码多的游戏输局局均输钱金额", "最大单局赢钱金额/所在游戏/日期", "最大单局输钱金额/所在游戏/日期", "日均游戏在线时长(分钟)"}
		for _, info := range datas {
			row := []any{
				info.UserId,
				info.UserType,
				info.ChannelClass,
				info.Channel1,
				info.RegistTime,
				info.LastLoginTime,
				info.LoseDays,
				info.ActiveTime,
				info.FActiveDay,
				info.RewardRate,
				info.Profit,
				info.RefUsers,
				"",
				info.FirstPayDate,
				info.FirstPayAmount,
				info.PayAmount,
				info.PayCount,
				info.PayAvg,
				info.PaySuccessRate,
				info.FirstWithdrawDate,
				info.FirstWithdrawAmount,
				info.WithdrawAmount,
				info.WithdrawCount,
				info.WithdrawAvg,
				info.WithdrawSuccessRate,
				info.FirstPayEfficience,
				info.PayWithdrawDiffDay,
				info.PayBeforeCarry,
				info.PayWithdrawTimes,
				"",
				info.MaxPayDate,
				info.MaxWithdrawDate,
				info.MaxPayDateSuccessRate,
				info.MaxWithdrawDateSuccessRate,
				info.MaxPayTimesDateSuccessRate,
				info.MaxWithdrawTimesDateSuccessRate,
				"",
				info.Game1TimesStats,
				info.Game1BetStats,
				info.Game2TimesStats,
				info.Game2BetStats,
				info.Game3TimesStats,
				info.Game3BetStats,
				info.Game1PlayWinsAvg,
				info.Game1PlayLosesAvg,
				info.Game1BetWinsAvg,
				info.Game1BetLosesAvg,
				info.MaxWinStats,
				info.MaxLosesStats,
				info.ActiveDayAvg,
			}
			rows = append(rows, row)
		}
	}

	// 检查是否有数据可导出
	if count <= 0 {
		this.showMsg("未查询到可以导出的数据！", MSG_ERR, "")
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
	title := "大R分析_"
	if typeId == 1 {
		title = "玩家分析总表"
	}
	fileName := title + time.Now().Format("20060102150405") + ".xlsx"
	// 解决文件名中文乱码问题
	fileName = path.Base(fileName)
	fileName = url.QueryEscape(fileName)
	this.Ctx.Output.Header("Content-type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	this.Ctx.Output.Header("Content-Disposition", "attachment;filename="+fileName)
	this.Ctx.Output.Header("Pragma", "No-cache")
	this.Ctx.Output.Header("Cache-Control", "No-cache")
	this.Ctx.Output.Header("Expires", "0")
	// var buffer bytes.Buffer
	if err := file.Write(this.Ctx.ResponseWriter); err != nil {
		logs.Error(err)
	}
	// r := bytes.NewReader(buffer.Bytes())

	// http.ServeContent(this.Ctx.ResponseWriter, this.Ctx.Request, fileName, time.Now(), r)
}

// 将 Excel 列的数字编号转换为相应的列标字母
// num: 1=A, 2=B, 27=AA, 703=AAA
func NumberToExcelColumn(num int) string {
	column := ""
	for num > 0 {
		remainder := (num - 1) % 26
		column = string(rune(65+remainder)) + column // 65 是 'A' 的 ASCII 值
		num = (num - 1) / 26
	}
	return column
}

func (this *StatisticsController) UserAnalysis2() {
	page, _ := strconv.Atoi(this.GetString("page"))
	if page <= 0 {
		page = 1
	}
	userid := this.GetString("userid")
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	start_date1 := this.GetString("start_date1")
	end_date1 := this.GetString("end_date1")
	paystatr := this.GetString("pay_statr")
	payend := this.GetString("pay_end")
	withdrawstatr := this.GetString("withdraw_statr")
	withdrawend := this.GetString("withdraw_end")
	gainstatr := this.GetString("gain_statr")
	gainend := this.GetString("gain_end")
	utypeId := this.GetString("utype_id")
	maxRoundsGtype, _ := strconv.Atoi(this.GetString("maxRoundsGtype"))
	maxBetGtype, _ := strconv.Atoi(this.GetString("maxBetGtype"))
	typeId, _ := strconv.Atoi(this.GetString("typeId"))
	density, _ := strconv.Atoi(this.GetString("density")) // typeId=2横轴密度(分钟)
	if density <= 0 {
		density = 10
	}
	_hand := this.GetString("_hand")

	today := service.NowTime()
	if startDate == "" && endDate == "" {
		// 默认近七天数据
		endDate = today.Format("2006-01-02")
		startDate = today.AddDate(0, 0, -6).Format("2006-01-02")
	}
	if (start_date1 == "" || start_date1 == "<nil>") && (end_date1 == "" || end_date1 == "<nil>") {
		// 默认近七天数据
		end_date1 = today.Format("2006-01-02")
		start_date1 = today.AddDate(0, 0, -6).Format("2006-01-02")
	}

	startTime, endTime := service.FindByDate4(startDate, endDate)
	rangeStartTime, rangeEndTime := service.FindByDate4(start_date1, end_date1)
	// 如果只是输入了玩家ID，则注册日期默认填入该玩家的注册日期
	if userid != "" && utils.SliceIn(typeId, 2, 3, 4) {
		registryTime, err := service.StatisticsService.GetUserRegistryTime(userid)
		if err != nil {
			this.showMsg(err.Error(), MSG_ERR)
		}
		rangeStartTime = &registryTime
		start_date1 = registryTime.Format(utils.FORMAT_DATE)
		rangeEndTime, _ = service.FindByDate4(today.Format("2006-01-02"), "")
		end_date1 = rangeEndTime.Format(utils.FORMAT_DATE)
	}
	if typeId == 3 || typeId == 4 {
		if rangeStartTime == nil && rangeEndTime != nil {
			t := rangeEndTime.AddDate(0, 0, -6)
			rangeStartTime = &t
			start_date1 = t.Format(utils.FORMAT_DATE)
		}
		if rangeStartTime != nil && rangeEndTime == nil {
			t := rangeStartTime.AddDate(0, 0, 6)
			rangeEndTime = &t
			end_date1 = t.Format(utils.FORMAT_DATE)
		}
	}

	pay_statr, _ := strconv.Atoi(paystatr)
	pay_end, _ := strconv.Atoi(payend)
	moneyRange := []int64{int64(pay_statr), int64(pay_end)}
	withdraw_statr, _ := strconv.Atoi(withdrawstatr)
	withdraw_end, _ := strconv.Atoi(withdrawend)
	withdrawRange := []int64{int64(withdraw_statr), int64(withdraw_end)}
	gain_statr, _ := strconv.Atoi(gainstatr)
	gain_end, _ := strconv.Atoi(gainend)
	profitRange := []int64{int64(gain_statr), int64(gain_end)}

	var utypeIds []int
	if utypeId != "" && utypeId != "0" && utypeId != "-" && utypeId != "<nil>" {
		ids := strings.Split(utypeId, ",")
		for _, _id := range ids {
			id, err := strconv.Atoi(_id)
			if err != nil || id == 0 {
				continue
			}
			utypeIds = append(utypeIds, id)
		}
	} else {
		utypeId = ""
	}

	var list any
	var count int
	if typeId == 1 {
		// 玩家分析总表
		if this.pageSize == 20 {
			this.pageSize = 40
		}
		if _hand == "1" { // 非手动点击页面不查询
			total, datas, err := service.StatisticsService.PlayerAnalysis(page, this.pageSize, startTime, endTime, userid, utypeIds, moneyRange, withdrawRange, profitRange, int32(maxRoundsGtype), int32(maxBetGtype))
			if err != nil {
				debug.PrintStack()
				beego.Error(err)
				this.showMsg(err.Error(), MSG_ERR)
			}
			list = datas
			count = total
		}
	} else if typeId == 2 {
		// 玩家活跃时段分布图
		labels, datasets, err := service.StatisticsService.PlayerAnalysis_ActiveChart(startTime, endTime, userid, utypeIds, moneyRange, withdrawRange, profitRange, int32(maxRoundsGtype), int32(maxBetGtype), rangeStartTime, rangeEndTime, int32(density))
		if err != nil {
			beego.Error(err)
			this.showMsg(err.Error(), MSG_ERR)
		}
		this.Data["labels"] = labels
		this.Data["datasets"] = datasets
	} else if typeId == 3 {
		// 在线时长走势图
		labels, datasets, err := service.StatisticsService.PlayerAnalysis_OnlineChart(startTime, endTime, userid, utypeIds, moneyRange, withdrawRange, profitRange, int32(maxRoundsGtype), int32(maxBetGtype), rangeStartTime, rangeEndTime)
		if err != nil {
			beego.Error(err)
			this.showMsg(err.Error(), MSG_ERR)
		}
		this.Data["labels"] = labels
		this.Data["datasets"] = datasets
	} else if typeId == 4 {
		labels, payCounts, withdrawCounts, payAmounts, withdrawAmounts, err := service.StatisticsService.PlayerAnalysis_TradeChart(startTime, endTime, userid, utypeIds, moneyRange, withdrawRange, profitRange, int32(maxRoundsGtype), int32(maxBetGtype), rangeStartTime, rangeEndTime)
		if err != nil {
			beego.Error(err)
			this.showMsg(err.Error(), MSG_ERR)
		}
		this.Data["labels"] = labels
		this.Data["payCounts"] = payCounts
		this.Data["withdrawCounts"] = withdrawCounts
		this.Data["payAmounts"] = payAmounts
		this.Data["withdrawAmounts"] = withdrawAmounts
	} else {
		// 大R分析
		if pay_statr != 0 || pay_end != 0 || withdraw_statr != 0 || withdraw_end != 0 || gain_statr != 0 || gain_end != 0 {
			total, datas, err := service.StatisticsService.GetUserAnalysisCk(page, this.pageSize, pay_statr, pay_end, withdraw_statr, withdraw_end, gain_statr, gain_end)
			if err != nil {
				beego.Error(err)
			}
			list = datas
			count = int(total)
		}
	}

	usertype := []map[int]string{
		{0: "全部"},
		{10: "新手"},
		{1: "平民"},
		{2: "普R"},
		{3: "小R"},
		{4: "中R"},
		{5: "大R"},
		{6: "超大R"},
	}

	// 游戏列表
	recordList := map[int]string{0: "全部"}
	utils.CopyMap(recordList, service.GtypeNameMap)

	this.Data["pageTitle"] = "玩家分析"
	this.Data["typeId"] = typeId
	this.Data["list"] = list
	this.Data["count"] = count

	this.Data["usertype"] = usertype
	this.Data["recordList"] = recordList

	this.Data["userid"] = userid
	this.Data["usertypeId"] = utypeId
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["pay_statr"] = paystatr
	this.Data["pay_end"] = payend
	this.Data["withdraw_statr"] = withdrawstatr
	this.Data["withdraw_end"] = withdrawend
	this.Data["gain_statr"] = gainstatr
	this.Data["gain_end"] = gainend
	this.Data["maxRoundsGtype"] = maxRoundsGtype
	this.Data["maxBetGtype"] = maxBetGtype
	this.Data["start_date1"] = start_date1
	this.Data["end_date1"] = end_date1
	this.Data["density"] = density

	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.UserAnalysis2", "_hand", "1", "typeId", typeId, "userid", userid, "start_date", startDate, "end_date", endDate, "pay_statr", pay_statr, "pay_end", pay_end, "withdraw_statr", withdraw_statr, "withdraw_end", withdraw_end, "gain_statr", gain_statr, "gain_end", gain_end, "utype_id", utypeId, "maxRoundsGtype", maxRoundsGtype, "maxBetGtype", maxBetGtype, "start_date1", start_date1, "end_date1", end_date1, "density", density), true).ToString()
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "useranalysis2opt")
	this.display()
}

func (this *StatisticsController) PlayerAnalysis() {
	page, _ := strconv.Atoi(this.GetString("page"))
	userid := this.GetString("userid")
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	rangeStartDate := this.GetString("range_start_date")
	rangeEndDate := this.GetString("range_end_date")
	paystatr := this.GetString("pay_statr")
	payend := this.GetString("pay_end")
	withdrawstatr := this.GetString("withdraw_statr")
	withdrawend := this.GetString("withdraw_end")
	gainstatr := this.GetString("gain_statr")
	gainend := this.GetString("gain_end")
	utypeId := this.GetString("utype_id")
	maxRoundsGtype, _ := strconv.Atoi(this.GetString("maxRoundsGtype"))
	maxBetGtype, _ := strconv.Atoi(this.GetString("maxBetGtype"))
	typeId, _ := strconv.Atoi(this.GetString("typeId"))
	density, _ := strconv.Atoi(this.GetString("density")) // typeId=2横轴密度(分钟)
	if density == 0 {
		density = 10
	}

	if startDate == "" && endDate == "" {
		// 默认近七天数据
		today := bson.Now()
		endDate = today.Format("2006-01-02")
		startDate = today.AddDate(0, 0, -6).Format("2006-01-02")
	}

	// 玩家类型 账号类型 流失天数筛选 打码量筛选 局均码筛选
	startTime, endTime := service.FindByDate4(startDate, endDate)

	pay_statr, _ := strconv.Atoi(paystatr)
	pay_end, _ := strconv.Atoi(payend)
	moneyRange := []int64{int64(pay_statr), int64(pay_end)}
	withdraw_statr, _ := strconv.Atoi(withdrawstatr)
	withdraw_end, _ := strconv.Atoi(withdrawend)
	withdrawRange := []int64{int64(withdraw_statr), int64(withdraw_end)}
	gain_statr, _ := strconv.Atoi(gainstatr)
	gain_end, _ := strconv.Atoi(gainend)
	profitRange := []int64{int64(gain_statr), int64(gain_end)}

	var utypeIds []int
	if utypeId != "" && utypeId != "0" && utypeId != "-" {
		ids := strings.Split(utypeId, ",")
		for _, _id := range ids {
			id, err := strconv.Atoi(_id)
			if err != nil || id == 0 {
				continue
			}
			utypeIds = append(utypeIds, id)
		}
	}

	var count int
	if typeId == 1 {
		// 玩家活跃时段分布图
		// service.StatisticsService.PlayerAnalysis_ActiveChart(page, this.pageSize, startTime, endTime, userid, utypeIds, moneyRange, withdrawRange, profitRange, int32(maxRoundsGtype), int32(maxBetGtype))

	} else {
		c, list, err := service.StatisticsService.PlayerAnalysis(page, this.pageSize, startTime, endTime, userid, utypeIds, moneyRange, withdrawRange, profitRange, int32(maxRoundsGtype), int32(maxBetGtype))
		if err != nil {
			debug.PrintStack()
			beego.Error(err)
			this.showMsg(err.Error(), MSG_ERR)
		}
		count = c
		this.Data["count"] = c
		this.Data["list"] = list
	}

	usertype := []map[int]string{
		{0: "全部"},
		{10: "新手"},
		{1: "平民"},
		{2: "普R"},
		{3: "小R"},
		{4: "中R"},
		{5: "大R"},
		{6: "超大R"},
	}

	// 游戏列表
	recordList := map[int]string{0: "全部"}
	utils.CopyMap(recordList, service.GtypeNameMap)

	this.Data["pageTitle"] = "玩家分析总表"
	this.Data["typeId"] = typeId

	this.Data["usertype"] = usertype
	this.Data["recordList"] = recordList

	this.Data["userid"] = userid
	this.Data["usertypeId"] = utypeId
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["pay_statr"] = paystatr
	this.Data["pay_end"] = payend
	this.Data["withdraw_statr"] = withdrawstatr
	this.Data["withdraw_end"] = withdrawend
	this.Data["gain_statr"] = gainstatr
	this.Data["gain_end"] = gainend
	this.Data["maxRoundsGtype"] = maxRoundsGtype
	this.Data["maxBetGtype"] = maxBetGtype
	this.Data["range_start_date"] = rangeStartDate
	this.Data["range_end_date"] = rangeEndDate
	this.Data["density"] = density

	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.PlayerAnalysis", "userid", userid, "start_date", startDate, "end_date", endDate, "pay_statr", pay_statr, "pay_end", pay_end, "withdraw_statr", withdraw_statr, "withdraw_end", withdraw_end, "gain_statr", gain_statr, "gain_end", gain_end, "utype_id", utypeId, "maxRoundsGtype", maxRoundsGtype, "maxBetGtype", maxBetGtype, "range_start_date", rangeStartDate, "range_end_date", rangeEndDate, "density", density), true).ToString()
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "playeranalysisopt")
	this.display()
}

// 金币流水曲线
func (this *StatisticsController) GoldFlowCurve() {
	userid := this.GetString("userid")
	start_date := this.GetString("start_date")
	end_date := this.GetString("end_date")
	gamestatr := this.GetString("game_statr")
	gameend := this.GetString("game_end")
	latelygame := this.GetString("lately_game")

	dates := make([]interface{}, 0) // 日期
	data := make([]interface{}, 0)  // 携带
	data1 := make([]interface{}, 0) // 盈亏
	data2 := make([]interface{}, 0) // 充值
	data3 := make([]interface{}, 0) // 提现
	cause := make([]interface{}, 0) // 携带原因
	var err error
	if userid != "" {
		game_statr, _ := strconv.Atoi(gamestatr)
		game_end, _ := strconv.Atoi(gameend)
		lately_game, _ := strconv.Atoi(latelygame)
		dates, data, data1, data2, data3, cause, err = service.StatisticsService.GetGoldFlowCurveInfoCK(userid, start_date, end_date, game_statr, game_end, lately_game)
	}
	if err != nil {
		beego.Error(err)
	}
	this.Data["pageTitle"] = "金币流水曲线"
	this.Data["TimeLabel"] = dates
	this.Data["XD"] = data
	this.Data["ZYK"] = data1
	this.Data["PAY"] = data2
	this.Data["WithDraw"] = data3
	this.Data["Cause"] = cause
	this.Data["userid"] = userid
	this.Data["startDate"] = start_date
	this.Data["endDate"] = end_date
	this.Data["game_statr"] = gamestatr
	this.Data["game_end"] = gameend
	this.Data["lately_game"] = latelygame
	this.display()
}

// 玩家充值统计
func (this *StatisticsController) PlayerRecharge() {
	userid := this.GetString("userid")
	paycount := this.GetString("paycount")
	sid := make([]string, 0)
	data := make([]int, 0)
	num1 := 0
	if paycount != "" {
		num1, _ = strconv.Atoi(paycount)
		if num1 > 10 {
			// 超过10次  默认最高10次
			paycount = "10"
			num1 = 10
		}
	} else {
		// 默认三
		num1 = 3
		paycount = "3"
	}
	if userid != "" {
		if strings.Contains(userid, ",") {
			id_arr := strings.Split(userid, ",")
			sid = append(sid, id_arr...)
		} else {
			sid = append(sid, userid)
		}
	}
	list, _ := service.StatisticsService.GetPlayerRecharge(sid, num1)
	count := len(list)
	for i := 0; i < num1; i++ {
		data = append(data, i+1)
	}
	this.Data["pageTitle"] = "玩家充值统计"
	this.Data["userid"] = userid
	this.Data["paycount"] = paycount
	this.Data["list"] = list
	this.Data["count"] = count
	this.Data["totalCount"] = data
	this.display()
}

// 新手库存
func (this *StatisticsController) NoviceStock() {
	list, _ := service.StatisticsService.GetNoviceStockList(bson.M{})
	count := len(list)
	this.Data["pageTitle"] = "新手库存"
	this.Data["list"] = list
	this.Data["count"] = count
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "novicestockopt")
	this.display()
}

// 修改新手库存
func (this *StatisticsController) EditNoviceStock() {
	if this.isPost() {
		gtype, _ := this.GetInt("gtype")
		stock, _ := this.GetInt("stock")
		if gtype == 0 {
			this.checkError(fmt.Errorf("游戏ID不正确"))
		}
		info := new(entity.ModifyNewbieStock)
		info.Gtype = int32(gtype)
		info.Stock = int64(stock)

		result, err := service.GmRequest(pb.WebNewbieStock, pb.CONFIG_UPSERT, info)
		beego.Trace("result: ", result)
		if err != nil {
			this.checkError(err)
		}
		service.ActionService.Add("NoviceStock_cash_edit", this.auth.GetUser().UserName, "", utils.String(gtype), utils.String(gtype), "")
		this.redirect(beego.URLFor("StatisticsController.NoviceStock"))
	}
	this.display()
}

func (this *StatisticsController) FirstCharge() {
	status, _ := this.GetInt("status")
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	packageId := this.GetString("package_id")
	aliasId := this.GetString("alias_id")
	aliasName := this.GetString("alias_name")
	if page < 1 {
		page = 1
	}
	if startDate == "" && endDate == "" {
		// 默认近十天数据
		today := bson.Now()
		startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -9).Format("2006-01-02"))
		endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
	}
	if status == 0 {
		my_data := this.GetSession("my_select_pakeageid")
		if my_data != nil && packageId == "" {
			if packid, ok := my_data.(string); ok {
				packageId = packid
			}
		}
		status = 1
	}
	m := service.FindByDate1(startDate, endDate, "date", "date")
	count := int64(0)
	isChannel := false
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
		isChannel = true
		// 存储缓存
		this.SetSession("my_select_pakeageid", packageId)
	} else {
		// 清除缓存
		this.DelSession("my_select_pakeageid")
	}
	// 渠道别名
	if aliasId != "" && aliasId != "0" {
		if aliasId == "-" {
			m["channel1"] = ""
		} else {
			if strings.Contains(aliasId, ",") {
				palkage_arr := strings.Split(aliasId, ",")
				temp_arr := make([]string, 0)
				for _, item := range palkage_arr {
					if item != "" && item != "0" && item != "-" {
						temp_arr = append(temp_arr, item)
					}
				}
				m["channel1"] = bson.M{"$in": temp_arr}
			} else {
				m["channel1"] = aliasId
			}
			isChannel = true
		}
	}
	if aliasName != "" {
		m["channel1"] = bson.M{"$regex": aliasName, "$options": "i"}
		isChannel = true
	}

	count, _ = service.StatisticsService.GetFirstChargeTotal(m, isChannel)
	list, _ := service.StatisticsService.GetFirstChargeList(page, this.pageSize, m, isChannel)
	this.Data["list"] = list

	packageList, _ := service.ChannelService.GetChannelAll()
	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()

	this.Data["pageTitle"] = "首充分析"
	this.Data["count"] = count
	this.Data["status"] = status
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.FirstCharge", "status", status, "package_id", packageId, "alias_id", aliasId, "alias_name", aliasName, "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["packageId"] = packageId
	this.Data["packageList"] = packageList
	this.Data["aliasId"] = aliasId
	this.Data["packageAliasList"] = packageAliasList
	this.Data["aliasName"] = aliasName
	this.display()
}

// 玩家综合情况总览
func (c *StatisticsController) UserGeneralStats() {
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

	liveDays1, _ := c.GetInt("liveDays1")
	liveDays2, _ := c.GetInt("liveDays2")
	lossDays1, _ := c.GetInt("lossDays1")
	lossDays2, _ := c.GetInt("lossDays2")
	profit1, _ := c.GetInt("profit1")
	profit2, _ := c.GetInt("profit2")
	profitCash1, _ := c.GetInt("profitCash1")
	profitCash2, _ := c.GetInt("profitCash2")
	bets1, _ := c.GetInt("bets1")
	bets2, _ := c.GetInt("bets2")
	betAvg1, _ := c.GetInt("betAvg1")
	betAvg2, _ := c.GetInt("betAvg2")
	cash1, _ := c.GetInt("cash1")
	cash2, _ := c.GetInt("cash2")
	money1, _ := c.GetInt("money1")
	money2, _ := c.GetInt("money2")

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

	// 流失天数筛选 账面净赢利筛选 实得净赢利筛选
	if liveDays1 > 0 && liveDays2 > 0 {
		params["liveDays1"] = liveDays1
		params["liveDays2"] = liveDays2
	}
	if lossDays1 > 0 && lossDays2 > 0 {
		params["lossDays1"] = lossDays1
		params["lossDays2"] = lossDays2
	}
	if profit1 != 0 && profit2 != 0 {
		params["profit1"] = profit1 * 100
		params["profit2"] = profit2 * 100
	} else {
		profit1, profit2 = 0, 0
	}
	if profitCash1 != 0 && profitCash2 != 0 {
		params["profitCash1"] = profitCash1 * 100
		params["profitCash2"] = profitCash2 * 100
	} else {
		profitCash1, profitCash2 = 0, 0
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
	if money1 != 0 && money2 != 0 {
		params["money1"] = money1 * 100
		params["money2"] = money2 * 100
	} else {
		money1, money2 = 0, 0
	}

	// 明细用户类型筛选多选
	if userid != "" {
		params["userid"] = userid
	}

	var utypeIds []string
	if utypeId != "" && utypeId != "0" && utypeId != "-" {
		utypeIds = strings.Split(utypeId, ",")
	}

	var utypeFilters []map[string]any
	for _, utypeId := range utypeIds {
		usertypeId, err := strconv.Atoi(utypeId)
		if err != nil {
			beego.Error("usertypeId error: ", err)
			continue
		}
		if usertypeId == 0 {
			continue
		}
		var uParam = make(map[string]any)
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
			uParam["state"] = 2
			uParam["money1"] = startMoney
			if endMoney != -1 {
				uParam["money2"] = endMoney
			}
		} else {
			if usertypeId != 11 {
				uParam["state"] = usertypeId
			}
		}
		if len(uParam) > 0 {
			utypeFilters = append(utypeFilters, uParam)
		}
	}
	if len(utypeFilters) > 0 {
		params["utypes"] = utypeFilters
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
	if typeId == 0 { // 经济总况
		total, list, err := service.StatisticsService.UserGeneralStatsFinance(page, c.pageSize, params)
		if err != nil {
			beego.Error("UserGeneralStatsFinance error: ", err)
		}
		c.Data["list"] = list
		count = total
	} else if typeId == 1 { // 游戏分布
		total, stats, err := service.StatisticsService.UserGeneralStatsGames(page, c.pageSize, params)
		if err != nil {
			beego.Error("UserGeneralStatsGames error: ", err)
		}
		c.Data["list"] = stats
		count = total
	} else if typeId == 2 {
		// VIPbank明细
		total, stats, err := service.StatisticsService.UserGeneralStatsVBBank(page, c.pageSize, params)
		if err != nil {
			beego.Error("UserGeneralStatsVBBank error: ", err)
		}
		c.Data["list"] = stats
		count = total
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
	packageClassAll, _ := service.ChannelService.GetChannelClassAll()
	c.Data["count"] = count
	c.Data["pageBar"] = libs.NewPager(page, int(count), c.pageSize, beego.URLFor("StatisticsController.UserGeneralStats", "start_date", startDate, "end_date", endDate, "status", status, "package_id", packageId, "typeId", typeId, "alias_id", aliasId, "class_id", classId, "alias_name", aliasName, "utype_id", utypeId, "tab_id", tabId, "userid", userid, "liveDays1", liveDays1, "liveDays2", liveDays2, "lossDays1", lossDays1, "lossDays2", lossDays2, "profit1", profit1, "profit2", profit2, "profitCash1", profitCash1, "profitCash2", profitCash2, "bets1", bets1, "bets2", bets2, "betAvg1", betAvg1, "betAvg2", betAvg2, "cash1", cash1, "cash2", cash2, "money1", money1, "money2", money2), true).ToString()
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
	c.Data["packageClassAll"] = packageClassAll
	c.Data["aliasName"] = aliasName
	c.Data["status"] = status
	c.Data["tabId"] = tabId
	c.Data["page"] = page
	c.Data["userid"] = userid
	c.Data["liveDays1"] = liveDays1
	c.Data["liveDays2"] = liveDays2
	c.Data["lossDays1"] = lossDays1
	c.Data["lossDays2"] = lossDays2
	c.Data["profit1"] = profit1
	c.Data["profit2"] = profit2
	c.Data["profitCash1"] = profitCash1
	c.Data["profitCash2"] = profitCash2
	c.Data["bets1"] = bets1
	c.Data["bets2"] = bets2
	c.Data["betAvg1"] = betAvg1
	c.Data["betAvg2"] = betAvg2
	c.Data["cash1"] = cash1
	c.Data["cash2"] = cash2
	c.Data["money1"] = money1
	c.Data["money2"] = money2
	c.Data["pageTitle"] = "玩家综合情况总览"
	c.display()
}

// 玩家综合情况总览导出
func (c *StatisticsController) UserGeneralStatsExport() {
	startDate := c.GetString("start_date")
	endDate := c.GetString("end_date")

	status, _ := c.GetInt("status")
	page, _ := strconv.Atoi(c.GetString("page"))
	packageId := c.GetString("package_id")
	typeId, _ := c.GetInt("typeId")
	aliasId := c.GetString("alias_id")
	classId := c.GetString("class_id")
	utypeId := c.GetString("utype_id")
	tabId, _ := c.GetInt("tab_id")
	if page < 1 {
		page = 1
	}
	userid := c.GetString("userid")

	liveDays1, _ := c.GetInt("liveDays1")
	liveDays2, _ := c.GetInt("liveDays2")
	lossDays1, _ := c.GetInt("lossDays1")
	lossDays2, _ := c.GetInt("lossDays2")
	profit1, _ := c.GetInt("profit1")
	profit2, _ := c.GetInt("profit2")
	profitCash1, _ := c.GetInt("profitCash1")
	profitCash2, _ := c.GetInt("profitCash2")
	bets1, _ := c.GetInt("bets1")
	bets2, _ := c.GetInt("bets2")
	betAvg1, _ := c.GetInt("betAvg1")
	betAvg2, _ := c.GetInt("betAvg2")
	cash1, _ := c.GetInt("cash1")
	cash2, _ := c.GetInt("cash2")
	money1, _ := c.GetInt("money1")
	money2, _ := c.GetInt("money2")

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

	// 流失天数筛选 账面净赢利筛选 实得净赢利筛选
	if liveDays1 > 0 && liveDays2 > 0 {
		params["liveDays1"] = liveDays1
		params["liveDays2"] = liveDays2
	}
	if lossDays1 > 0 && lossDays2 > 0 {
		params["lossDays1"] = lossDays1
		params["lossDays2"] = lossDays2
	}
	if profit1 != 0 && profit2 != 0 {
		params["profit1"] = profit1 * 100
		params["profit2"] = profit2 * 100
	} else {
		profit1, profit2 = 0, 0
	}
	if profitCash1 != 0 && profitCash2 != 0 {
		params["profitCash1"] = profitCash1 * 100
		params["profitCash2"] = profitCash2 * 100
	} else {
		profitCash1, profitCash2 = 0, 0
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
	if money1 != 0 && money2 != 0 {
		params["money1"] = money1 * 100
		params["money2"] = money2 * 100
	} else {
		money1, money2 = 0, 0
	}

	// 明细用户类型筛选多选
	if userid != "" {
		params["userid"] = userid
	}

	var utypeIds []string
	if utypeId != "" && utypeId != "0" && utypeId != "-" {
		utypeIds = strings.Split(utypeId, ",")
	}

	var utypeFilters []map[string]any
	for _, utypeId := range utypeIds {
		usertypeId, err := strconv.Atoi(utypeId)
		if err != nil {
			beego.Error("usertypeId error: ", err)
			continue
		}
		if usertypeId == 0 {
			continue
		}
		var uParam = make(map[string]any)
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
				// utypeFilters = append(utypeFilters, "(t1.money=0)")
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
			uParam["state"] = 2
			uParam["money1"] = startMoney
			if endMoney != -1 {
				uParam["money2"] = endMoney
				// utypeFilters = append(utypeFilters, fmt.Sprintf("(t0.state = 2 and money >= %d)", startMoney))
			} else {
				// utypeFilters = append(utypeFilters, fmt.Sprintf("(state = 2 and money between %d and %d)", startMoney, endMoney))
			}
		} else {
			if usertypeId != 11 {
				uParam["state"] = usertypeId
			}
		}
		if len(uParam) > 0 {
			utypeFilters = append(utypeFilters, uParam)
		}
	}
	// if len(utypeFilters) > 0 {
	// 	params["utypeQ"] = fmt.Sprintf(" and (%s)", strings.Join(utypeFilters, " or "))
	// }

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
	var headers []string
	var rows [][]any
	pageSize := 100000
	if typeId == 0 { // 经济总况
		total, list, err := service.StatisticsService.UserGeneralStatsFinance(1, pageSize, params)
		if err != nil {
			beego.Error("UserGeneralStatsFinance error: ", err)
		}
		count = total
		headers = []string{"序号", "玩家id", "玩家昵称", "渠道类", "渠道别名", "玩家类型", "账号类型", "注册时间", "最后登陆时间", "存活天数", "流失天数", "携带金额", "总充值金额", "总提走金额", "待审核金额", "审核通过未到账金额", "被冻结金额", "账面返奖率", "账面净盈利", "实得净盈利", "自研游戏净盈利", "外接slots净盈利", "外接真人净盈利", "VIPbank未解锁bonus", "打码解锁并领走的金额", "打码已经解锁未领走的金额", "其他领取的cash", "误差额", "VIP中产出cash占总充值比例", "", "tp净盈利/返奖率", "rummy净盈利/返奖率", "crash净盈利/返奖率", "aviator净盈利/返奖率", "lhd净盈利/返奖率", "7up净盈利/返奖率", "kingvsqueen净盈利/返奖率", "ab净盈利/返奖率", "l3净盈利/返奖率", "joker净盈利/返奖率", "ak47净盈利/返奖率"}
		for _, v := range list {
			row := []any{
				v.No,
				v.Userid,
				v.Nickname,
				v.ChannelClass,
				v.ChannelAlias,
				v.UserType,
				v.RegistArea,
				v.Ctime,
				v.LoginTime,
				v.LiveDays,
				v.LoseDays,
				v.Diamond,
				v.PayAmount,
				v.WithdrawAmount,
				v.WithdrawAmountWait,
				v.WithdrawAmountProcess,
				v.WithdrawAmountFreeze,
				v.RebateRate,
				v.Profit,
				v.ProfitCash,
				v.ProfitGames,
				v.ProfitExternalSlots,
				v.ProfitExternalZrsx,
				v.VipBankUnlockBonus,
				v.VipBankClaimedBonus,
				v.VipBankUnclaimedBonus,
				v.OtherCash,
				v.FaultCash,
				v.VipCashPayRate,
				"",
				v.TpScoreStats,
				v.RummyScoreStats,
				v.CrashScoreStats,
				v.AviatorScoreStats,
				v.LhdScoreStats,
				v.UpScoreStats,
				v.KingvsqueenScoreStats,
				v.AbScoreStats,
				v.L3ScoreStats,
				v.JokerScoreStats,
				v.Ak47ScoreStats,
			}
			rows = append(rows, row)
		}
	} else if typeId == 1 { // 游戏分布
		total, stats, err := service.StatisticsService.UserGeneralStatsGames(1, pageSize, params)
		if err != nil {
			beego.Error("UserGeneralStatsGames error: ", err)
		}
		count = total
		headers = []string{"序号", "玩家id", "玩家昵称", "渠道类", "渠道别名", "玩家类型", "账号类型", "注册时间", "最后登陆时间", "存活天数", "流失天数", "总游戏局数", "总打码量", "局均打码", "净赢", "自研局数合计", "外接slots局数合计", "外接真人局数合计", "自研局数占比", "外接slots局数占比", "外接真人局数占比", "自研打码量合计", "外接slots打码量合计", "外接真人打码量合计", "自研打码量占比", "外接slots打码量占比", "外接真人打码量占比", "", "tp局数/占比", "rummy局数/占比", "crash局数/占比", "aviator局数/占比", "lhd局数/占比", "7up局数/占比", "kingvsqueen局数/占比", "ab局数/占比", "l3局数/占比", "joker局数/占比", "ak47局数/占比", "", "tp打码量/占比", "rummy打码量/占比", "crash打码量/占比", "aviator打码量/占比", "lhd打码量/占比", "7up打码量/占比", "kingvsqueen打码量/占比", "ab打码量/占比", "l3打码量/占比", "joker打码量/占比", "ak47打码量/占比"}
		for _, v := range stats {
			row := []any{
				v.No,
				v.Userid,
				v.Nickname,
				v.ChannelClass,
				v.ChannelAlias,
				v.UserType,
				v.RegistArea,
				v.Ctime,
				v.LoginTime,
				v.LiveDays,
				v.LoseDays,
				v.AllRounds,
				v.Bets,
				v.BetsAvg,
				v.Cash,
				v.RoundsGame,
				v.RoundsSlots,
				v.RoundsZrsx,
				v.RoundsRateGame,
				v.RoundsRateSlots,
				v.RoundsRateZrsx,
				v.BetsGames,
				v.BetsSlots,
				v.BetsZrsx,
				v.BetsRateGames,
				v.BetsRateSlots,
				v.BetsRateZrsx,
				"",
				v.TpRoundStats,
				v.RummyRoundStats,
				v.CrashRoundStats,
				v.AviatorRoundStats,
				v.LhdRoundStats,
				v.UpRoundStats,
				v.KingvsqueenRoundStats,
				v.AbRoundStats,
				v.L3RoundStats,
				v.JokerRoundStats,
				v.Ak47RoundStats,
				"",
				v.TpBetsStats,
				v.RummyBetsStats,
				v.CrashBetsStats,
				v.AviatorBetsStats,
				v.LhdBetsStats,
				v.UpBetsStats,
				v.KingvsqueenBetsStats,
				v.AbBetsStats,
				v.L3BetsStats,
				v.JokerBetsStats,
				v.Ak47BetsStats,
			}
			rows = append(rows, row)
		}
	} else if typeId == 2 {
		// VIPbank明细
		total, stats, err := service.StatisticsService.UserGeneralStatsVBBank(1, pageSize, params)
		if err != nil {
			beego.Error("UserGeneralStatsVBBank error: ", err)
		}
		count = total
		headers = []string{"序号", "玩家id", "玩家昵称", "渠道类", "渠道别名", "玩家类型", "账号类型", "注册时间", "最后登陆时间", "存活天数", "流失天数", "VIP最高等级", "VIP当前等级", "总充值金额", "总打码量", "打码量与充值金额比", "打码解锁比例", "bonus与打码量比", "bonus与总充值金额比", "VIP中产出cash占总充值比例", "VIPbank累计bonus总额", "VIPbank未解锁bonus", "打码解锁并领走的金额", "打码已经解锁未领走的金额", "bank总流水误差额", "", "试玩金转化成bonus的金额", "VIP升级奖金", "VIP周奖金", "充值时赠送的bonus", "在线时长抽奖获得的bonus", "完成任务获得的bonus", "分享下家的打码返bonus", "分享下家首充返bonus", "利息", "bonus来源误差额", "", "试玩金转化的bonus占比", "VIP升级奖金的bonus占比", "VIP周奖金的bonus占比", "充值时赠送的的bonus占比", "在线时长抽奖获得的bonus占比", "完成任务获得的bonus占比", "分享下家的打码返bonus占比", "分享下家首充返bonus占比"}
		for _, v := range stats {
			row := []any{
				v.No,
				v.Userid,
				v.Nickname,
				v.ChannelClass,
				v.ChannelAlias,
				v.UserType,
				v.RegistArea,
				v.Ctime,
				v.LoginTime,
				v.LiveDays,
				v.LoseDays,
				v.VipLevelMax,
				v.VipLevel,
				v.PayAmount,
				v.Bets,
				v.BetsPayRate,
				v.BetsBonusRate,
				v.BonusBetsRate,
				v.BonusPayRate,
				v.VipCashPayRate,
				v.VipBonus,
				v.VipBankBonus,
				v.CashoutBonus,
				v.UnlockBonus,
				v.FaultBank,
				"",
				v.TryBonus,
				v.VipUpgradeBonus,
				v.VipWeekBonus,
				v.ChargeBonus,
				v.OnlineBonus,
				v.TaskBonus,
				v.ShareBetsBonus,
				v.ShareBonus,
				v.InterestBonus,
				v.FaultBonus,
				"",
				v.TryBonusRate,
				v.VipUpgradeBonusRate,
				v.VipWeekBonusRate,
				v.ChargeBonusRate,
				v.OnlineBonusRate,
				v.TaskBonusRate,
				v.ShareBetsBonusRate,
				v.ShareBonusRate,
			}
			rows = append(rows, row)
		}
	}
	_ = count

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
	title := "玩家综合情况总览_"
	switch typeId {
	case 0:
		title += "经济总况"
	case 1:
		title += "游戏分布"
	case 2:
		title += "VIPbank明细"
	}
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

// 实时数据监控
func (c *StatisticsController) LiveData() {
	action := c.GetString("action")

	// 大盘历史累计
	if action == "History" {
		users, bets, err := service.StatisticsService.LiveDataHistoryStats()
		if err != nil {
			c.JsonRError(err.Error())
			return
		}
		c.JsonRSuccess(libs.R{
			"users": users,
			"bets":  service.Chip2Float(bets),
		})
		return
	}

	// 今日实时数据
	if action == "Today" {
		stats, err := service.StatisticsService.LiveDataTodayStats()
		if err != nil {
			c.JsonRError(err.Error())
			return
		}
		c.JsonRSuccess(stats)
		return
	}

	// 走势曲线
	if action == "Trend" {
		arg := &args.LiveDataTrendArg{}
		c.JsonBody(arg)

		stime, etime := service.FindByDate4(arg.Stime, arg.Etime)
		if stime == nil || etime == nil {
			c.jsonResult(libs.RFail("走势曲线时间范围有误"))
			return
		}
		if len(arg.Trends) == 0 {
			c.jsonResult(libs.RFail("选择指标不能为空"))
			return
		}

		stats, err := service.StatisticsService.LiveDataTrendStats(*stime, *etime, arg.Trends)
		if err != nil {
			c.JsonRError(err.Error())
			return
		}
		c.JsonRSuccess(stats)
		return
	}

	c.Data["pageTitle"] = "实时数据监控"
	c.display()
}

// 打码统计
func (this *StatisticsController) PlayerBetStats() {
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	aliasId := this.GetString("alias_id")
	classId := this.GetString("class_id")
	registArea, _ := this.GetInt("registArea")
	typeId, _ := this.GetInt("typeId")
	utypeId := this.GetString("utype_id")
	page, _ := strconv.Atoi(this.GetString("page"))
	if page < 1 {
		page = 1
	}

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

	var utypeIds []int32
	if utypeId != "" && utypeId != "0" && utypeId != "-" && utypeId != "<nil>" {
		ids := strings.Split(utypeId, ",")
		for _, _id := range ids {
			id, err := strconv.Atoi(_id)
			if err != nil || id == 0 {
				continue
			}
			if id == 10 { // 新手
				id = 0
			}
			utypeIds = append(utypeIds, int32(id))
		}
	} else {
		utypeId = ""
	}

	var list any
	var count int64
	var err error
	if typeId == 1 {
		total, datas, e := service.StatisticsService.PlayerBetsUsers(true, page, this.pageSize, stime, etime, registAreas, utypeIds, packageIds)
		list = datas
		count = total
		err = e
	} else if typeId == 2 {
		total, datas, e := service.StatisticsService.PlayerBetsGtypeDates(true, page, this.pageSize, stime, etime, registAreas, utypeIds, packageIds)
		list = datas
		count = total
		err = e
	} else {
		datas, e := service.StatisticsService.PlayerBetsLevel(stime, etime, registAreas, utypeIds, packageIds)
		list = datas
		count = int64(len(datas))
		err = e
	}
	if err != nil {
		beego.Error(err)
		this.showMsg(err.Error(), MSG_ERR)
	}
	this.Data["list"] = list
	this.Data["count"] = count

	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()
	packageClassAll, _ := service.ChannelService.GetChannelClassAll()

	usertype := []map[int]string{
		{0: "全部"},
		{10: "新手"},
		{1: "平民"},
		{2: "普R"},
		{3: "小R"},
		{4: "中R"},
		{5: "大R"},
		{6: "超大R"},
	}

	this.Data["pageTitle"] = "打码统计"
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["aliasId"] = aliasId
	this.Data["classId"] = classId
	this.Data["registArea"] = registArea
	this.Data["packageAliasList"] = packageAliasList
	this.Data["packageClassAll"] = packageClassAll
	this.Data["typeId"] = typeId
	this.Data["usertypeId"] = utypeId
	this.Data["usertype"] = usertype
	this.Data["isExport"] = this.auth.HasAccessPerm(this.controllerName, "playerbetstatsexport")
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.PlayerBetStats", "start_date", startDate, "end_date", endDate, "alias_id", aliasId, "class_id", classId, "registArea", registArea, "typeId", typeId, "utype_id", utypeId), true).ToString()
	this.display()
}

// 打码统计(导出)
func (c *StatisticsController) PlayerBetStatsExport() {
	startDate := c.GetString("start_date")
	endDate := c.GetString("end_date")
	aliasId := c.GetString("alias_id")
	classId := c.GetString("class_id")
	registArea, _ := c.GetInt("registArea")
	typeId, _ := c.GetInt("typeId")
	utypeId := c.GetString("utype_id")

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

	var utypeIds []int32
	if utypeId != "" && utypeId != "0" && utypeId != "-" && utypeId != "<nil>" {
		ids := strings.Split(utypeId, ",")
		for _, _id := range ids {
			id, err := strconv.Atoi(_id)
			if err != nil || id == 0 {
				continue
			}
			if id == 10 { // 新手
				id = 0
			}
			utypeIds = append(utypeIds, int32(id))
		}
	} else {
		utypeId = ""
	}

	var rows [][]any
	var headers []string
	var err error
	if typeId == 1 {
		_, list, e := service.StatisticsService.PlayerBetsUsers(false, 1, 100000, stime, etime, registAreas, utypeIds, packageIds)
		err = e
		if err == nil {
			headers = []string{"日期", "名次", "UID", "注册时间", "最后登录时间", "存活天数", "流失天数", "VIP等级", "玩家类型", "玩家总充值", "玩家总提现", "玩家总充-提", "玩家总打码量", "玩家总返奖率", "玩家总输赢", "打码第1多的游戏", "打码第1多的游戏的码量", "打码第1多的游戏的局数", "打码第1多的游戏的局均码", "打码第2多的游戏", "打码第2多的游戏的码量", "打码第2多的游戏的局数", "打码第2多的游戏的局均码", "打码第3多的游戏", "打码第3多的游戏的码量", "打码第3多的游戏的局数", "打码第3多的游戏的局均码"}
			for _, v := range list {
				row := []any{
					v.SDate,
					v.Rank,
					v.Userid,
					v.Ctime,
					v.LastLoginTime,
					v.LiveDays,
					v.LoseDays,
					v.VipLv,
					v.UserType,
					v.Pays,
					v.Withdraws,
					v.PaysSubWithdraws,
					v.Bets,
					v.RabateRate,
					v.Wins,
					v.Top1Game,
					v.Top1GameBets,
					v.Top1GameRounds,
					v.Top1GameBetsAvg,
					v.Top2Game,
					v.Top2GameBets,
					v.Top2GameRounds,
					v.Top2GameBetsAvg,
					v.Top3Game,
					v.Top3GameBets,
					v.Top3GameRounds,
					v.Top3GameBetsAvg,
				}
				rows = append(rows, row)
			}
		}
	} else if typeId == 2 {
		_, list, e := service.StatisticsService.PlayerBetsGtypeDates(false, 1, 100000, stime, etime, registAreas, utypeIds, packageIds)
		err = e
		if err == nil {
			headers = []string{"日期", "游戏大类", "游戏小类", "厂商名称", "游戏ID", "游戏名称", "投注人数", "投注金额", "人均投注", "厂商回退金额", "厂商回滚金额", "玩家总输赢", "RTP"}
			for _, v := range list {
				row := []any{
					v.SDate,
					v.GtypeCategory1,
					v.GtypeCategory2,
					v.Factory,
					v.GameId,
					v.GameName,
					v.BetUsers,
					v.Bets,
					v.BetsAvg,
					v.Rebates,
					v.BetsRollback,
					v.Wins,
					v.RTP,
				}
				rows = append(rows, row)
			}
		}
	} else {
		list, e := service.StatisticsService.PlayerBetsLevel(stime, etime, registAreas, utypeIds, packageIds)
		err = e
		if err == nil {
			headers = []string{"日期", "总打码量", "打码前0.5%玩家人数", "打码前0.5%玩家码量总和", "打码前0.5%玩家人均码量", "打码前0.5%玩家码量占总码量比", "打码前1%玩家人数", "打码前1%玩家码量总和", "打码前1%玩家人均码量", "打码前1%玩家码量占总码量比", "打码前3%玩家人数", "打码前3%玩家码量总和", "打码前3%玩家人均码量", "打码前3%玩家码量占总码量比", "打码前5%玩家人数", "打码前5%玩家码量总和", "打码前5%玩家人均码量", "打码前5%玩家码量占总码量比", "打码前10%玩家人数", "打码前10%玩家码量总和", "打码前10%玩家人均码量", "打码前10%玩家码量占总码量比"}
			for _, v := range list {
				row := []any{
					v.SDate,
					v.Bets,
					v.Top05Users,
					v.Top05UsersBets,
					v.Top05UsersBetsAvg,
					v.Top05UsersBetsRate,
					v.Top1Users,
					v.Top1UsersBets,
					v.Top1UsersBetsAvg,
					v.Top1UsersBetsRate,
					v.Top3Users,
					v.Top3UsersBets,
					v.Top3UsersBetsAvg,
					v.Top3UsersBetsRate,
					v.Top5Users,
					v.Top5UsersBets,
					v.Top5UsersBetsAvg,
					v.Top5UsersBetsRate,
					v.Top10Users,
					v.Top10UsersBets,
					v.Top10UsersBetsAvg,
					v.Top10UsersBetsRate,
				}
				rows = append(rows, row)
			}
		}
	}
	if err != nil {
		beego.Error(err)
		c.showMsg(err.Error(), MSG_ERR)
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
	var title string
	switch typeId {
	default:
		title = "玩家打码分层"
	case 1:
		title = "玩家打码明细"
	case 2:
		title = "子游戏打码统计"
	}
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
