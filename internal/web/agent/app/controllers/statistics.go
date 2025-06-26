package controllers

import (
	"fmt"
	"goserver/internal/web/agent/app/entity"
	"goserver/internal/web/agent/app/libs"
	"goserver/internal/web/agent/app/service"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"
)

type StatisticsController struct {
	BaseController
}

// 数据汇总
func (this *StatisticsController) DataList() {
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	packageId := this.GetString("package_id")
	aliasId := this.GetString("alias_id")
	accountId := this.GetString("account_id")
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
	isChannel := false
	isfalg := true
	loginuser := this.auth.GetUser()
	channelList := make([]string, 0)
	if loginuser != nil {
		if loginuser.UserName != "admin" {
			if len(loginuser.RoleList) > 0 {
				for _, item := range loginuser.RoleList {
					for _, v := range item.ChannelList {
						channelList = append(channelList, v.Name)
					}
				}
			}
			m["channel"] = bson.M{"$in": channelList}
		} else {
			isfalg = false
		}

	}

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
			isChannel = true
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
			isChannel = true
		}
	}
	// 博主账号
	if accountId != "" && accountId != "0" {
		if strings.Contains(accountId, ",") {
			palkage_arr := strings.Split(accountId, ",")
			temp_arr := make([]string, 0)
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					temp_arr = append(temp_arr, item)
				}
			}
			m["blogger_id"] = bson.M{"$in": temp_arr}
		} else {
			m["blogger_id"] = accountId
		}
		isChannel = true
	}

	// if aliasName != "" {
	// 	if packageId == "" || packageId == "0" {
	// 		m["channel"] = bson.M{"$in": channelList}
	// 	}
	// 	m["channel1"] = bson.M{"$regex": aliasName, "$options": "i"}
	// 	isChannel = true
	// }

	count := int64(0)

	// if isChannel {
	// 	count, _ = service.StatisticsService.GetDataTotalByChannel(m)
	// 	list, _ := service.StatisticsService.GetDataListByChannerl(page, this.pageSize, m)
	// 	this.Data["list"] = list
	// } else {
	// 	m["channel"] = bson.M{"$in": channelList}
	// }
	count, _ = service.StatisticsService.GetDataTotal(m, isChannel)
	list, _ := service.StatisticsService.GetDataList(page, this.pageSize, m, isChannel)
	this.Data["list"] = list

	res, _ := service.ChannelService.GetChannelAll(isfalg, channelList)
	var packageList []entity.PackageInfo
	var aliasList []entity.PackageInfo
	var accountList []entity.PackageInfo
	if res != nil {
		packageList = res["channel"]
		aliasList = res["alias"]
		accountList = res["account"]
	}

	this.Data["pageTitle"] = "数据汇总"
	this.Data["count"] = count
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.DataList", "package_id", packageId, "alias_id", aliasId, "account_id", accountId, "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["packageId"] = packageId
	this.Data["packageList"] = packageList
	this.Data["aliasId"] = aliasId
	this.Data["packageAliasList"] = aliasList
	this.Data["accountId"] = accountId
	this.Data["accountList"] = accountList
	this.display()
}

// 分享数据
func (this *StatisticsController) Share() {
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	packageId := this.GetString("package_id")
	aliasId := this.GetString("alias_id")
	accountId := this.GetString("account_id")
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
	isChannel := false
	loginuser := this.auth.GetUser()
	channelList := make([]string, 0)
	isfalg := true
	if loginuser != nil {
		if loginuser.UserName != "admin" {
			if len(loginuser.RoleList) > 0 {
				for _, item := range loginuser.RoleList {
					for _, v := range item.ChannelList {
						channelList = append(channelList, v.Name)
					}
				}
			}
			m["channel"] = bson.M{"$in": channelList}
		} else {
			isfalg = false
		}
	}

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
			isChannel = true
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
			isChannel = true
		}
	}
	// 博主账号
	if accountId != "" && accountId != "0" {
		if strings.Contains(accountId, ",") {
			palkage_arr := strings.Split(accountId, ",")
			temp_arr := make([]string, 0)
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					temp_arr = append(temp_arr, item)
				}
			}
			m["blogger_id"] = bson.M{"$in": temp_arr}
		} else {
			m["blogger_id"] = accountId
		}
		isChannel = true
	}

	count := int64(0)
	// if isChannel {
	// 	count, _ = service.StatisticsService.GetShareByChannelTotal(m)
	// 	list, _ := service.StatisticsService.GetShareByChannelList(page, this.pageSize, m)
	// 	this.Data["list"] = list
	// } else {
	// 	m["channel"] = bson.M{"$in": channelList}

	// }

	count, _ = service.StatisticsService.GetShareTotal(m, isChannel)
	list, _ := service.StatisticsService.GetShareList(page, this.pageSize, m, isChannel)
	this.Data["list"] = list

	res, _ := service.ChannelService.GetChannelAll(isfalg, channelList)
	var packageList []entity.PackageInfo
	var aliasList []entity.PackageInfo
	var accountList []entity.PackageInfo
	if res != nil {
		packageList = res["channel"]
		aliasList = res["alias"]
		accountList = res["account"]
	}

	this.Data["pageTitle"] = "分享数据"
	this.Data["count"] = count
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.Share", "package_id", packageId, "alias_id", aliasId, "account_id", accountId, "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["packageId"] = packageId
	this.Data["packageList"] = packageList
	this.Data["aliasId"] = aliasId
	this.Data["packageAliasList"] = aliasList
	this.Data["accountId"] = accountId
	this.Data["accountList"] = accountList
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
	accountId := this.GetString("account_id")
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
	isChannel := false
	isfalg := true
	loginuser := this.auth.GetUser()
	channelList := make([]string, 0)
	if loginuser != nil {
		if loginuser.UserName != "admin" {
			if len(loginuser.RoleList) > 0 {
				for _, item := range loginuser.RoleList {
					for _, v := range item.ChannelList {
						channelList = append(channelList, v.Name)
					}
				}
			}
			m["channel"] = bson.M{"$in": channelList}
		} else {
			isfalg = false
		}

	}

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
			isChannel = true
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
			isChannel = true
		}
	}
	// 博主账号
	if accountId != "" && accountId != "0" {
		if strings.Contains(accountId, ",") {
			palkage_arr := strings.Split(accountId, ",")
			temp_arr := make([]string, 0)
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					temp_arr = append(temp_arr, item)
				}
			}
			m["blogger_id"] = bson.M{"$in": temp_arr}
		} else {
			m["blogger_id"] = accountId
		}
		isChannel = true
	}
	count := int64(0)
	if typeId == 1 {
		// 付费留存
		count, _ = service.StatisticsService.GetPayUserRetainedTotal(m, isChannel)
		list, _ := service.StatisticsService.GetPayUserRetainedList(page, this.pageSize, m, isChannel)
		this.Data["list"] = list
	} else {
		// 用户留存
		count, _ = service.StatisticsService.GetRetainedTotal(m, isChannel)
		list, _ := service.StatisticsService.GetRetainedList(page, this.pageSize, m, isChannel)
		this.Data["list"] = list
	}
	res, _ := service.ChannelService.GetChannelAll(isfalg, channelList)
	var packageList []entity.PackageInfo
	var aliasList []entity.PackageInfo
	var accountList []entity.PackageInfo
	if res != nil {
		packageList = res["channel"]
		aliasList = res["alias"]
		accountList = res["account"]
	}
	// strArr := make([]string, 0)
	// packageList, _ := service.ChannelService.GetChannelAll(false, strArr)
	// packageAliasList, _ := service.ChannelService.GetChannelAliasAll(nil)
	this.Data["pageTitle"] = "用户留存"
	this.Data["count"] = count
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.Retained", "typeId", typeId, "package_id", packageId, "alias_id", aliasId, "account_id", accountId, "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["packageId"] = packageId
	this.Data["packageList"] = packageList
	this.Data["aliasId"] = aliasId
	this.Data["packageAliasList"] = aliasList
	this.Data["accountId"] = accountId
	this.Data["accountList"] = accountList
	this.Data["typeId"] = typeId
	this.display()
}

// func (this *StatisticsController) PointList() {
// 	page, _ := strconv.Atoi(this.GetString("page"))
// 	startDate := this.GetString("start_date")
// 	endDate := this.GetString("end_date")
// 	tabid, _ := this.GetInt("tabid")
// 	packageId := this.GetString("package_id")
// 	aliasId := this.GetString("alias_id")
// 	aliasName := this.GetString("alias_name")
// 	if page < 1 {
// 		page = 1
// 	}
// 	if startDate == "" && endDate == "" {
// 		// 默认近十天数据
// 		today := bson.Now()
// 		startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -9).Format("2006-01-02"))
// 		endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
// 	}
// 	m := service.FindByDate1(startDate, endDate, "date", "date")
// 	count := int64(0)
// 	isChannel := false
// 	if packageId != "" && packageId != "0" {
// 		if packageId == "-" {
// 			m["channel"] = ""
// 		} else {
// 			if strings.Contains(packageId, ",") {
// 				palkage_arr := strings.Split(packageId, ",")
// 				temp_arr := make([]string, 0)
// 				for _, item := range palkage_arr {
// 					if item != "" && item != "0" && item != "-" {
// 						temp_arr = append(temp_arr, item)
// 					}
// 				}
// 				m["channel"] = bson.M{"$in": temp_arr}
// 			} else {
// 				m["channel"] = packageId
// 			}
// 			isChannel = true
// 		}
// 	}
// 	// 渠道别名
// 	if aliasId != "" && aliasId != "0" {
// 		if aliasId == "-" {
// 			m["channel1"] = ""
// 		} else {
// 			if strings.Contains(aliasId, ",") {
// 				palkage_arr := strings.Split(aliasId, ",")
// 				temp_arr := make([]string, 0)
// 				for _, item := range palkage_arr {
// 					if item != "" && item != "0" && item != "-" {
// 						temp_arr = append(temp_arr, item)
// 					}
// 				}
// 				m["channel1"] = bson.M{"$in": temp_arr}
// 			} else {
// 				m["channel1"] = aliasId
// 			}
// 			isChannel = true
// 		}
// 	}
// 	if aliasName != "" {
// 		m["channel1"] = bson.M{"$regex": aliasName, "$options": "i"}
// 		isChannel = true
// 	}

// 	if tabid == 1 {
// 		count, _ = service.StatisticsService.GetGameNumberAnalysisTotal(m, isChannel)
// 		list, _ := service.StatisticsService.GetGameNumberAnalysisList(page, this.pageSize, m, isChannel)
// 		this.Data["list"] = list
// 	} else if tabid == 2 {
// 		count, _ = service.StatisticsService.GetPlaytimeTotal(m, isChannel)
// 		list, _ := service.StatisticsService.GetPlaytimeList(page, this.pageSize, m, isChannel)
// 		this.Data["list"] = list
// 	} else {
// 		// 行为分析
// 		count, _ = service.StatisticsService.GetPointListTotal(m, isChannel)
// 		list, _ := service.StatisticsService.GetPointList(page, this.pageSize, m, isChannel)
// 		this.Data["list"] = list
// 	}
// 	packageList, _ := service.ChannelService.GetChannelAll()
// 	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()

// 	this.Data["pageTitle"] = "新用户埋点统计"
// 	this.Data["count"] = count
// 	this.Data["tabid"] = tabid
// 	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.PointList", "tabid", tabid, "package_id", packageId, "alias_id", aliasId, "alias_name", aliasName, "start_date", startDate, "end_date", endDate), true).ToString()
// 	this.Data["startDate"] = startDate
// 	this.Data["endDate"] = endDate
// 	this.Data["packageId"] = packageId
// 	this.Data["packageList"] = packageList
// 	this.Data["aliasId"] = aliasId
// 	this.Data["packageAliasList"] = packageAliasList
// 	this.Data["aliasName"] = aliasName
// 	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "pointlistopt")
// 	this.display()
// }

// 新用户埋点统计
func (this *StatisticsController) PointList() {
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	packageId := this.GetString("package_id")
	tabid, _ := this.GetInt("tabid")
	aliasId := this.GetString("alias_id")
	accountId := this.GetString("account_id")
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
	isChannel := false
	isfalg := true
	loginuser := this.auth.GetUser()
	channelList := make([]string, 0)
	if loginuser != nil {
		if loginuser.UserName != "admin" {
			if len(loginuser.RoleList) > 0 {
				for _, item := range loginuser.RoleList {
					for _, v := range item.ChannelList {
						channelList = append(channelList, v.Name)
					}
				}
			}
			m["channel"] = bson.M{"$in": channelList}
		} else {
			isfalg = false
		}

	}

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
			isChannel = true
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
			isChannel = true
		}
	}
	// 博主账号
	if accountId != "" && accountId != "0" {
		if strings.Contains(accountId, ",") {
			palkage_arr := strings.Split(accountId, ",")
			temp_arr := make([]string, 0)
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					temp_arr = append(temp_arr, item)
				}
			}
			m["blogger_id"] = bson.M{"$in": temp_arr}
		} else {
			m["blogger_id"] = accountId
		}
		isChannel = true
	}
	count := int64(0)
	if tabid == 1 {
		// 局数分析
		count, _ = service.StatisticsService.GetGameNumberAnalysisTotal(m, isChannel)
		list, _ := service.StatisticsService.GetGameNumberAnalysisList(page, this.pageSize, m, isChannel)
		this.Data["list"] = list
	} else if tabid == 2 {
		// 局数分析
		count, _ = service.StatisticsService.GetPlaytimeTotal(m, isChannel)
		list, _ := service.StatisticsService.GetPlaytimeList(page, this.pageSize, m, isChannel)
		this.Data["list"] = list
	} else {
		// 行为分析
		count, _ = service.StatisticsService.GetPointListTotal(m, isChannel)
		list, _ := service.StatisticsService.GetPointList(page, this.pageSize, m, isChannel)
		this.Data["list"] = list
	}
	res, _ := service.ChannelService.GetChannelAll(isfalg, channelList)
	var packageList []entity.PackageInfo
	var aliasList []entity.PackageInfo
	var accountList []entity.PackageInfo
	if res != nil {
		packageList = res["channel"]
		aliasList = res["alias"]
		accountList = res["account"]
	}

	this.Data["pageTitle"] = "新用户埋点统计"
	this.Data["count"] = count
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("StatisticsController.PointList", "tabid", tabid, "package_id", packageId, "alias_id", aliasId, "account_id", accountId, "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["packageId"] = packageId
	this.Data["packageList"] = packageList
	this.Data["aliasId"] = aliasId
	this.Data["packageAliasList"] = aliasList
	this.Data["accountId"] = accountId
	this.Data["accountList"] = accountList
	this.Data["tabid"] = tabid
	this.display()
}
