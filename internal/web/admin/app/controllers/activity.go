package controllers

import (
	"errors"
	"fmt"
	"goserver/gen/pb"
	"goserver/internal/web/admin/app/args"
	"goserver/internal/web/admin/app/entity"
	"goserver/internal/web/admin/app/libs"
	"goserver/internal/web/admin/app/service"
	"goserver/pkg/data"
	"goserver/pkg/data/mq"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/globalsign/mgo/bson"
)

type ActivityController struct {
	BaseController
}

// 基础付费活动
func (this *ActivityController) BasicPay() {
	status, _ := this.GetInt("status")
	page, _ := this.GetInt("page")
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	userid := this.GetString("userid")
	packageId := this.GetString("package_id")
	typeId, _ := this.GetInt("type_id")
	tabid, _ := this.GetInt("tabid")
	if page < 1 {
		page = 1
	}
	count := int64(0)
	packageList, _ := service.ChannelService.GetChannelAll()
	typeList := map[int]string{
		0: "用户ID",
		1: "手机号",
		2: "用户昵称",
		3: "IP",
		4: "设备号",
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
	if tabid == 1 {
		// 领取记录
		if userid != "" {
			n := bson.M{}
			if typeId == 0 {
				m["userid"] = userid
			} else if typeId == 1 {
				n["phone"] = userid
			} else if typeId == 2 {
				n["nickname"] = bson.M{"$regex": userid, "$options": "i"}
			} else if typeId == 3 {
				n["login_ip"] = userid
			} else if typeId == 4 {
				n["device_id"] = userid
			}
			if len(n) > 0 {
				userlist, _ := service.PlayerService.GetByUserList(n)
				ids := make([]string, 0)
				for _, v := range userlist {
					ids = append(ids, v.Userid)
				}
				if len(ids) > 0 {
					m["userid"] = bson.M{"$in": ids}
				}
			}
		}
		if packageId != "" && packageId != "0" && packageId != "-" {
			if strings.Contains(packageId, ",") {
				palkage_arr := strings.Split(packageId, ",")
				temp_arr := make([]string, 0)
				for _, item := range palkage_arr {
					if item != "" && item != "0" && item != "-" {
						temp_arr = append(temp_arr, item)
					}
				}
				m["media_source"] = bson.M{"$in": temp_arr}
			} else {
				m["media_source"] = packageId
			}
			// 存储缓存
			this.SetSession("my_select_pakeageid", packageId)
		} else {
			// 清除缓存
			this.DelSession("my_select_pakeageid")

		}
		types := []int{58, 59, 57}
		m["ltype"] = bson.M{"$in": types}
		list, _ := service.PlayerService.GetGoldList(page, this.pageSize, m)
		count, _ = service.PlayerService.GetGoldListTotal(m)

		this.Data["list"] = list
		this.Data["count"] = count
	} else {
		// 活动数据
		if startDate == "" && endDate == "" {
			// 默认近十天数据
			today := bson.Now()
			startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -10).Format("2006-01-02"))
			endDate = fmt.Sprintf("%s", today.AddDate(0, 0, -1).Format("2006-01-02"))
		}
		m = service.FindByDate1(startDate, endDate, "date", "date")
		list, _ := service.ActivityService.GetBasicPayList(page, this.pageSize, m)
		count, _ = service.ActivityService.GetBasicPayTotal(m)
		this.Data["list"] = list
		this.Data["count"] = count
	}

	this.Data["pageTitle"] = "基础付费活动"
	this.Data["packageList"] = packageList
	this.Data["packageId"] = packageId
	this.Data["userid"] = userid
	this.Data["typeList"] = typeList
	this.Data["typeId"] = typeId
	if tabid == 1 {
		this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("ActivityController.BasicPay", "status", status, "tabid", tabid, "userid", userid, "type_id", typeId, "package_id", packageId), true).ToString()
	} else {
		this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("ActivityController.BasicPay", "tabid", tabid, "start_date", startDate, "end_date", endDate), true).ToString()
	}
	this.Data["tabid"] = tabid
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["status"] = status
	this.display()
}

// 基础免费活动
func (this *ActivityController) BasicFree() {
	status, _ := this.GetInt("status")
	page, _ := this.GetInt("page")
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	userid := this.GetString("userid")
	packageId := this.GetString("package_id")
	typeId, _ := this.GetInt("type_id")
	tabid, _ := this.GetInt("tabid")
	if page < 1 {
		page = 1
	}
	packageList, _ := service.ChannelService.GetChannelAll()
	typeList := map[int]string{
		0: "用户ID",
		1: "手机号",
		2: "用户昵称",
		3: "IP",
		4: "设备号",
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
	count := int64(0)
	m := bson.M{}
	if tabid == 1 {
		n := bson.M{}
		if userid != "" {
			if typeId == 0 {
				m["userid"] = userid
			} else if typeId == 1 {
				n["phone"] = userid
			} else if typeId == 2 {
				n["nickname"] = bson.M{"$regex": userid, "$options": "i"}
			} else if typeId == 3 {
				n["login_ip"] = userid
			} else if typeId == 4 {
				n["device_id"] = userid
			}
			if len(n) > 0 {
				userlist, _ := service.PlayerService.GetByUserList(n)
				ids := make([]string, 0)
				for _, v := range userlist {
					ids = append(ids, v.Userid)
				}
				if len(ids) > 0 {
					m["userid"] = bson.M{"$in": ids}
				}
			}
		}
		if packageId != "" && packageId != "0" && packageId != "-" {
			if strings.Contains(packageId, ",") {
				palkage_arr := strings.Split(packageId, ",")
				temp_arr := make([]string, 0)
				for _, item := range palkage_arr {
					if item != "" && item != "0" && item != "-" {
						temp_arr = append(temp_arr, item)
					}
				}
				m["media_source"] = bson.M{"$in": temp_arr}
			} else {
				m["media_source"] = packageId
			}
			// 存储缓存
			this.SetSession("my_select_pakeageid", packageId)
		} else {
			// 清除缓存
			this.DelSession("my_select_pakeageid")

		}
		types := []int{61, 62, 46, 66}
		m["ltype"] = bson.M{"$in": types}
		list, _ := service.PlayerService.GetGoldList(page, this.pageSize, m)
		count, _ = service.PlayerService.GetGoldListTotal(m)

		this.Data["list"] = list
		this.Data["count"] = count
	} else {
		// 活动数据
		if startDate == "" && endDate == "" {
			// 默认近十天数据
			today := bson.Now()
			startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -10).Format("2006-01-02"))
			endDate = fmt.Sprintf("%s", today.AddDate(0, 0, -1).Format("2006-01-02"))
		}
		m = service.FindByDate1(startDate, endDate, "date", "date")
		list, _ := service.ActivityService.GetBasicFreeList(page, this.pageSize, m)
		count, _ = service.ActivityService.GetBasicFreeTotal(m)
		this.Data["list"] = list
		this.Data["count"] = count
	}

	this.Data["pageTitle"] = "基础免费活动"
	this.Data["packageList"] = packageList
	this.Data["packageId"] = packageId
	this.Data["userid"] = userid
	this.Data["typeList"] = typeList
	this.Data["typeId"] = typeId
	if tabid == 1 {
		this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("ActivityController.BasicPay", "status", status, "tabid", tabid, "userid", userid, "type_id", typeId, "package_id", packageId), true).ToString()
	} else {
		this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("ActivityController.BasicPay", "tabid", tabid, "start_date", startDate, "end_date", endDate), true).ToString()
	}
	this.Data["status"] = status
	this.Data["tabid"] = tabid
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.display()
}

// 分享活动
func (this *ActivityController) Sharing() {
	status, _ := this.GetInt("status")
	page, _ := this.GetInt("page")
	userid := this.GetString("userid")
	packageId := this.GetString("package_id")
	typeId, _ := this.GetInt("type_id")
	tabid, _ := this.GetInt("tabid")
	if page < 1 {
		page = 1
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
	if userid != "" {
		if typeId == 0 {
			m["_id"] = userid
		} else if typeId == 1 {
			m["phone"] = userid
		} else if typeId == 2 {
			m["nickname"] = bson.M{"$regex": userid, "$options": "i"}
		} else if typeId == 3 {
			m["login_ip"] = userid
		} else if typeId == 4 {
			m["device_id"] = userid
		}
	}
	if packageId != "" && packageId != "0" && packageId != "-" {
		if strings.Contains(packageId, ",") {
			palkage_arr := strings.Split(packageId, ",")
			temp_arr := make([]string, 0)
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					temp_arr = append(temp_arr, item)
				}
			}
			m["ad__bundle_id"] = bson.M{"$in": temp_arr}
		} else {
			m["ad__bundle_id"] = packageId
		}
		// 存储缓存
		this.SetSession("my_select_pakeageid", packageId)
	} else {
		// 清除缓存
		this.DelSession("my_select_pakeageid")
	}
	if tabid == 0 {
		// 活动数据
		startDate := this.GetString("start_date")
		endDate := this.GetString("end_date")
		if startDate == "" && endDate == "" {
			// 默认近十天数据
			today := bson.Now()
			startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -9).Format("2006-01-02"))
			endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
		}
		m1 := service.FindByDate1(startDate, endDate, "ctime", "ctime")
		list, _ := service.ActivityService.GetShareDataList(page, this.pageSize, m1)
		count, _ := service.ActivityService.GetShareDataTotal(m1)
		this.Data["count"] = count
		this.Data["list"] = list
		this.Data["start_date"] = startDate
		this.Data["end_date"] = endDate
		this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("ActivityController.Sharing", "status", status, "tabid", tabid, "start_date", startDate, "end_date", endDate), true).ToString()
	} else if tabid == 1 {
		// 领取记录
		m = bson.M{}
		if userid != "" {
			if typeId == 0 {
				m["userid"] = userid
			} else if typeId == 1 {
				n := bson.M{}
				n["phone"] = userid
				userlist, _ := service.PlayerService.GetByUserList(n)
				ids := make([]string, 0)
				for _, v := range userlist {
					ids = append(ids, v.Userid)
				}
				m["userid"] = bson.M{"$in": ids}
			} else if typeId == 2 {
				n := bson.M{}
				n["nickname"] = bson.M{"$regex": userid, "$options": "i"}
				userlist, _ := service.PlayerService.GetByUserList(n)
				ids := make([]string, 0)
				for _, v := range userlist {
					ids = append(ids, v.Userid)
				}
				m["userid"] = bson.M{"$in": ids}
			} else if typeId == 3 {
				m["ip"] = userid
			} else if typeId == 4 {
				n := bson.M{}
				n["device_id"] = userid
				userlist, _ := service.PlayerService.GetByUserList(n)
				ids := make([]string, 0)
				for _, v := range userlist {
					ids = append(ids, v.Userid)
				}
				m["userid"] = bson.M{"$in": ids}
			}
		}
		// bundle_id
		if packageId != "" {
			if packageId != "0" {
				if packageId == "-" {
					m["bundle_id"] = ""
				} else {
					m["bundle_id"] = packageId
				}
			}
		}

		if packageId != "" && packageId != "0" && packageId != "-" {
			if strings.Contains(packageId, ",") {
				palkage_arr := strings.Split(packageId, ",")
				temp_arr := make([]string, 0)
				for _, item := range palkage_arr {
					if item != "" && item != "0" && item != "-" {
						temp_arr = append(temp_arr, item)
					}
				}
				m["bundle_id"] = bson.M{"$in": temp_arr}
			} else {
				m["bundle_id"] = packageId
			}
		}
		list, _ := service.ActivityService.GetShareRecordList(page, this.pageSize, m)
		count, _ := service.ActivityService.GetShareRecordTotal(m)
		this.Data["count"] = count
		this.Data["list"] = list
		this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("ActivityController.Sharing", "status", status, "tabid", tabid, "userid", userid, "type_id", typeId, "package_id", packageId), true).ToString()
	} else if tabid == 2 {
		// 分享排名
		m["share_below"] = bson.M{"$ne": bson.M{}, "$exists": true}
		m["robot"] = false
		m["simulation_robot"] = false

		list, _ := service.ActivityService.GetShareRankingList(page, this.pageSize, m)
		count, _ := service.ActivityService.GetShareRankingTotal(m)
		this.Data["count"] = count
		this.Data["list"] = list
		this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("ActivityController.Sharing", "status", status, "tabid", tabid, "userid", userid, "type_id", typeId, "package_id", packageId), true).ToString()
	} else if tabid == 3 {
		// 上下级设置
		if userid == "" {
			m["_id"] = ""
		}
		m["robot"] = false
		m["simulation_robot"] = false
		list, _ := service.PlayerService.GetList(page, this.pageSize, m)
		count, _ := service.PlayerService.GetTotal(m)
		if typeId == 0 && userid != "" {
			if len(list) > 0 {
				ids := make([]string, 0)
				for _, item := range list {
					if item.ShareSuperior != "" {
						ids = append(ids, item.ShareSuperior)
					}
				}
				if len(ids) > 0 {
					m1 := bson.M{}
					m1["_id"] = bson.M{"$in": ids}
					list1, _ := service.PlayerService.GetList(page, this.pageSize, m1)
					count1, _ := service.PlayerService.GetTotal(m1)
					list = append(list, list1...)
					count += count1
				}

			}
		}
		this.Data["count"] = count
		this.Data["list"] = list
		this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("ActivityController.Sharing", "status", status, "tabid", tabid, "userid", userid, "type_id", typeId, "package_id", packageId), true).ToString()
	} else if tabid == 4 {
		// 活动配置
		m := bson.M{}
		list, _ := service.ActivityService.GetShareConfigList(page, this.pageSize, m)
		count, _ := service.ActivityService.GetShareConfigTotal(m)
		this.Data["count"] = count
		this.Data["list"] = list
		this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("ActivityController.Sharing", "tabid", tabid), true).ToString()
	}
	packageList, _ := service.ChannelService.GetChannelAll()
	typeList := map[int]string{
		0: "用户ID",
		1: "手机号",
		2: "用户昵称",
		3: "IP",
		4: "设备号",
	}
	this.Data["pageTitle"] = "分享活动"
	this.Data["userid"] = userid
	this.Data["packageList"] = packageList
	this.Data["packageId"] = packageId
	this.Data["typeList"] = typeList
	this.Data["typeId"] = typeId
	this.Data["tabid"] = tabid
	this.Data["status"] = status
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "sharingopt")
	this.display()
}

// 修改上级
func (this *ActivityController) SuperiorEdit() {
	userid := this.GetString("curuser")
	supuserid := this.GetString("superioruser")
	if userid == "" {
		this.checkError(errors.New("用户ID不能为空"))
	}
	if this.isPost() {
		// 提交
		info, _ := service.PlayerService.GetPlayer(userid)
		if info != nil {
			if info.Userid == "" {
				this.checkError(errors.New("未查询到设置的用户信息！"))
			}
		}
		if supuserid != "" {
			supinfo, _ := service.PlayerService.GetPlayer(supuserid)
			if supinfo != nil {
				if supinfo.Userid == "" {
					this.checkError(errors.New("设置的上级ID不存在，请填写正确的用户ID！"))
				}
			}
		}
		// 通知服务器
		b := new(pb.ChangeShareSuperior)
		b.Userid = info.Userid
		b.Superior = supuserid
		b.Name = info.Nickname
		result, err := service.GmRequest(pb.WebSuperior, pb.CONFIG_UPSERT, b)
		beego.Trace("result: ", result)
		if err != nil {
			this.checkError(err)
		} else {
			service.ActionService.Add("edit_share_superior", this.auth.GetUserName(),
				"", utils.String(info.Userid), utils.String(info.Userid), "")
		}

		// if !validateIndianPhoneNumber(phone) {
		// 	this.checkError(errors.New("手机号码格式不正确"))
		// }

		// temp, _ := service.PayService.GetPayBlackListById(phone)
		// if temp != nil {
		// 	if temp.Phone != "" {
		// 		this.checkError(errors.New("已存在该支付手机号码，不可重复添加！"))
		// 	}
		// }

		// name := this.auth.GetUser().UserName
		// info := new(entity.PayBlackList)
		// info.Phone = phone
		// info.Operator = name
		// // 新增
		// err := service.PayService.AddPayBlackList(info)
		// this.checkError(err)

		// // 通知服务器
		// // b := make(map[string]string)
		// // b[phone] = ""
		// b := make([]entity.PayBlackList, 0)
		// b = append(b, *info)
		// result, err := service.GmRequest(pb.WebRechareBlack, pb.CONFIG_UPSERT, b)
		// beego.Trace("result: ", result)
		// if err != nil {
		// 	this.checkError(err)
		// } else {
		// 	service.ActionService.Add("add_pay_blacklist", name,
		// 		"", utils.String(info.Phone), utils.String(info.Phone))
		// }
		this.redirect(beego.URLFor("ActivityController.Sharing", "tabid", 3, "userid", userid))
	}
}

// 小米活动
func (this *ActivityController) Miui() {
	status, _ := this.GetInt("status")
	page, _ := this.GetInt("page")
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	userid := this.GetString("userid")
	takedId, _ := this.GetInt("taked_id")
	robotId, _ := this.GetInt("robot_id")
	if page < 1 {
		page = 1
	}
	count := int64(0)
	// 活动数据
	if startDate == "" && endDate == "" {
		// 默认近十天数据
		today := bson.Now()
		startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -10).Format("2006-01-02"))
		endDate = fmt.Sprintf("%s", today.AddDate(0, 0, -1).Format("2006-01-02"))
	}
	m := service.FindByDate(startDate, endDate, "ctime", "ctime")
	if userid != "" {
		m["userid"] = userid
	}
	if takedId != 0 {
		if takedId == 1 {
			m["taked"] = true
		} else {
			m["taked"] = false
		}
	}
	if robotId != 0 {
		if robotId == 1 {
			m["robot"] = false
		} else {
			m["robot"] = true
		}
	}
	list, _ := service.ActivityService.GetMiuiList(page, this.pageSize, m)
	count, _ = service.ActivityService.GetMiuiTotal(m)
	robotList := map[int]string{
		0: "全部",
		1: "玩家",
		2: "机器人",
	}
	takedList := map[int]string{
		0: "全部",
		1: "已领取",
		2: "未领取",
	}
	this.Data["list"] = list
	this.Data["count"] = count
	this.Data["pageTitle"] = "小米活动"
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("ActivityController.Miui", "userid", userid, "taked_id", takedId, "robot_id", robotId, "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["status"] = status
	this.Data["robotList"] = robotList
	this.Data["robotId"] = robotId
	this.Data["takedList"] = takedList
	this.Data["takedId"] = takedId
	this.display()
}

// 彩票活动
func (this *ActivityController) Lottery() {
	status, _ := this.GetInt("status")
	page, _ := this.GetInt("page")
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	if page < 1 {
		page = 1
	}
	count := int64(0)
	// 活动数据
	if startDate == "" && endDate == "" {
		// 默认近十天数据
		today := bson.Now()
		startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -9).Format("2006-01-02"))
		endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
	}
	m := service.FindByDate1(startDate, endDate, "date", "date")

	list, _ := service.ActivityService.GetLotteryList(page, this.pageSize, m)
	count, _ = service.ActivityService.GetLotteryTotal(m)
	this.Data["list"] = list
	this.Data["count"] = count
	this.Data["pageTitle"] = "彩票活动"
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("ActivityController.Lottery", "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["status"] = status
	this.display()
}

// 包赔活动
func (this *ActivityController) Compensation() {
	page, _ := this.GetInt("page")
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	userid := this.GetString("userid")
	packageId := this.GetString("package_id")
	typeId, _ := this.GetInt("type_id")
	tabid, _ := this.GetInt("tabid")
	if page < 1 {
		page = 1
	}
	m := bson.M{}
	if tabid == 1 {
		if userid != "" {
			if typeId == 0 {
				m["_id"] = userid
			} else if typeId == 1 {
				m["phone"] = userid
			} else if typeId == 2 {
				m["nickname"] = bson.M{"$regex": userid, "$options": "i"}
			} else if typeId == 3 {
				m["login_ip"] = userid
			} else if typeId == 4 {
				m["device_id"] = userid
			}
		}
		if packageId != "" {
			if packageId != "0" {
				if packageId == "-" {
					m["bundle_id"] = ""
				} else {
					m["bundle_id"] = packageId
				}
			}
		}
	} else {
		m = service.FindByDate(startDate, endDate, "ctime", "ctime")
	}

	list, _ := service.GameService.GetStockList(page, this.pageSize, m)
	count, _ := service.GameService.GetStockTotal(m)

	ids := make([]string, 0)
	for _, v := range list {
		ids = append(ids, v.Id)
	}

	gameList, _ := service.GameService.GetByGame(ids)
	for i := range list {
		for j := range gameList {
			var user entity.Game
			if list[i].Id == gameList[j].Id {
				user = gameList[j]
				list[i].GameInfo = user
				break
			}
			list[i].GameInfo = user
		}
	}
	packageList, _ := service.ChannelService.GetChannelAll()
	typeList := map[int]string{
		0: "用户ID",
		1: "手机号",
		2: "用户昵称",
		3: "IP",
		4: "设备号",
	}
	this.Data["pageTitle"] = "包赔活动"
	this.Data["packageList"] = packageList
	this.Data["packageId"] = packageId
	this.Data["userid"] = userid
	this.Data["typeList"] = typeList
	this.Data["typeId"] = typeId
	this.Data["count"] = count
	this.Data["list"] = list
	if tabid == 1 {
		this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("ActivityController.Compensation", "tabid", tabid, "userid", userid, "type_id", typeId, "package_id", packageId), true).ToString()
	} else {
		this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("ActivityController.Compensation", "tabid", tabid, "start_date", startDate, "end_date", endDate), true).ToString()
	}
	this.Data["tabid"] = tabid
	this.display()
}

// Sharedraw 拼多多抽奖活动
func (c *ActivityController) Sharedraw() {
	startDate := c.GetString("start_date")
	endDate := c.GetString("end_date")
	if startDate == "" && endDate == "" {
		// 默认近十天数据
		today := bson.Now()
		startDate = today.AddDate(0, 0, -10).Format("2006-01-02")
		endDate = today.Format("2006-01-02")
	}
	m := service.FindByDate1(startDate, endDate, "date", "date")
	stats, err := service.ActivityService.Sharedraw(m, startDate, endDate)
	if err != nil {
		beego.Error(err)
	}

	c.Data["list"] = stats
	c.Data["count"] = len(stats)
	c.Data["pageTitle"] = "玩游戏分享抽大奖活动"
	c.Data["startDate"] = startDate
	c.Data["endDate"] = endDate
	c.Data["isOperation"] = c.auth.HasAccessPerm(c.controllerName, "sharedraw")
	c.display()
}

// BetRank 排行榜活动
func (c *ActivityController) BetRank() {
	tabid, _ := c.GetInt("tabid")
	page, _ := c.GetInt("page")
	startDate := c.GetString("start_date")
	endDate := c.GetString("end_date")
	rank_type, _ := c.GetInt("rank_type")
	userid := c.GetString("userid")
	startDate2 := c.GetString("start_date2")
	endDate2 := c.GetString("end_date2")
	rankSort, _ := c.GetInt("rankSort")

	if page < 1 {
		page = 1
	}
	// 活动数据
	if startDate == "" && endDate == "" {
		// 默认近七天数据
		today := service.NowTime()
		startDate = today.AddDate(0, 0, -6).Format("2006-01-02")
		endDate = today.Format("2006-01-02")
	}
	s, e := service.FindByDate4(startDate, endDate)
	var start, end int64
	if s != nil {
		start = (*s).Unix()
	}
	if e != nil {
		end = (*e).Unix()
	}

	// if startDate2 == "" && endDate2 == "" {
	// 	// 默认近七天数据
	// 	today := service.NowTime()
	// 	startDate2 = today.AddDate(0, 0, -1).Format("2006-01-02")
	// 	endDate2 = today.Format("2006-01-02")
	// }
	s2, e2 := service.FindByDate4(startDate2, endDate2)
	var start2, end2 int64
	if s2 != nil {
		start2 = (*s2).Unix()
	}
	if e2 != nil {
		end2 = (*e2).Unix()
	}

	var err error
	var total int64
	var list any
	if tabid == 0 {
		total, list, err = service.ActivityService.BetRank(page, c.pageSize, int32(rank_type), start, end)
	} else if tabid == 1 {
		arg := &service.BetRankUsersArgs{
			Page:     page,
			PageSize: c.pageSize,
			Userid:   userid,
			RankType: int32(rank_type),
			Begin:    start,
			End:      end,
			Begin2:   start2,
			End2:     end2,
			RankSort: int32(rankSort),
		}
		total, list, err = service.ActivityService.BetRankUsers(*arg)
	}
	if err != nil {
		beego.Error(err)
		c.showMsg(err.Error(), MSG_ERR)
	}

	c.Data["list"] = list
	c.Data["count"] = total
	c.Data["pageTitle"] = "排行榜"
	c.Data["pageBar"] = libs.NewPager(page, int(total), c.pageSize, beego.URLFor("ActivityController.BetRank", "tabid", tabid, "startDate", startDate, "endDate", endDate, "start_date", startDate, "end_date", endDate, "rank_type", rank_type, "userid", userid, "startDate2", startDate2, "endDate2", endDate2, "rankSort", rankSort), true).ToString()
	c.Data["startDate"] = startDate
	c.Data["endDate"] = endDate
	c.Data["tabid"] = tabid
	c.Data["rank_type"] = rank_type
	c.Data["userid"] = userid
	c.Data["startDate2"] = startDate2
	c.Data["endDate2"] = endDate2
	c.Data["rankSort"] = rankSort
	c.display()
}

// BetRankExport 排行榜活动导出
func (c *ActivityController) BetRankExport() {
	tabid, _ := c.GetInt("tabid")
	page, _ := c.GetInt("page")
	startDate := c.GetString("start_date")
	endDate := c.GetString("end_date")
	rank_type, _ := c.GetInt("rank_type")
	userid := c.GetString("userid")
	startDate2 := c.GetString("start_date2")
	endDate2 := c.GetString("end_date2")
	rankSort, _ := c.GetInt("rankSort")
	if page < 1 {
		page = 1
	}

	// 活动数据
	if startDate == "" && endDate == "" {
		// 默认近七天数据
		today := service.NowTime()
		startDate = today.AddDate(0, 0, -6).Format("2006-01-02")
		endDate = today.Format("2006-01-02")
	}
	s, e := service.FindByDate4(startDate, endDate)
	var start, end int64
	if s != nil {
		start = (*s).Unix()
	}
	if e != nil {
		end = (*e).Unix()
	}

	// if startDate2 == "" && endDate2 == "" {
	// 	// 默认近七天数据
	// 	today := service.NowTime()
	// 	startDate2 = today.AddDate(0, 0, -1).Format("2006-01-02")
	// 	endDate2 = today.Format("2006-01-02")
	// }
	s2, e2 := service.FindByDate4(startDate2, endDate2)
	var start2, end2 int64
	if s2 != nil {
		start2 = (*s2).Unix()
	}
	if e2 != nil {
		end2 = (*e2).Unix()
	}

	var err error
	var headers []string
	var rows [][]any
	if tabid == 0 {
		_, list, e := service.ActivityService.BetRank(1, 100000, int32(rank_type), start, end)
		if e != nil {
			err = e
		} else {
			headers = []string{"发放日期", "星期", "榜单类型", "获奖人数", "奖池金额", "分到金额", "人均奖金", "奖金类型", "真人最高名次", "最高名次真人的id", "最高名次真人分到金额"}
			for _, v := range list {
				row := []any{
					v.SDate,
					v.Week,
					v.RankType,
					v.PrizeUsers,
					v.Jackpot,
					v.PrizeSum,
					v.PrizeAvg,
					v.PrizeType,
					v.UserMaxRank,
					v.MaxUserid,
					v.MaxUserPrize,
				}
				rows = append(rows, row)
			}
		}
	} else if tabid == 1 {
		arg := &service.BetRankUsersArgs{
			Page:     1,
			PageSize: 100000,
			Userid:   userid,
			RankType: int32(rank_type),
			Begin:    start,
			End:      end,
			Begin2:   start2,
			End2:     end2,
			RankSort: int32(rankSort),
		}
		_, list, e := service.ActivityService.BetRankUsers(*arg)
		if e != nil {
			err = e
		} else {
			headers = []string{"榜单日期", "发放日期", "玩家ID", "榜单类型", "结算时间", "发放时间", "领取时间", "发领间隔", "总充值", "总提现", "玩家总贡献", "当日充值", "当日提现", "当日总贡献", "榜单期打码", "发放日打码", "名次", "分到金额", "奖金类型", "玩家手机号"}
			for _, v := range list {
				row := []any{
					v.RankDate,
					v.SDate,
					v.Userid,
					v.RankType,
					v.Etime,
					v.Stime,
					v.ReceiveTime,
					v.ReceiveHourGap,
					v.Pays,
					v.Withdraws,
					v.Contribute,
					v.StimePays,
					v.StimeWithdraws,
					v.StimeContribute,
					v.RankBets,
					v.StimeBets,
					v.Rank,
					v.Prize,
					v.PrizeType,
					v.Phone,
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
	switch tabid {
	case 0:
		title += "排行榜汇总"
	case 1:
		title += "排行榜明细"
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

// Turn 转盘活动
func (this *ActivityController) Turn() {
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	startDate2 := this.GetString("start_date2")
	endDate2 := this.GetString("end_date2")
	packageId := this.GetString("package_id")
	typeId, _ := this.GetInt("typeId")
	aliasId := this.GetString("alias_id")
	userid := this.GetString("userid")
	classId := this.GetString("class_id")
	utypeId := this.GetString("utype_id")
	tabId, _ := this.GetInt("tab_id")
	auditStatus := this.GetString("auditStatus")
	prizeTypes := this.GetString("prizeTypes")
	auditUsers := this.GetString("auditUsers")

	if page < 1 {
		page = 1
	}
	// 活动数据
	today := service.NowTime()
	if startDate == "" {
		// 默认近七天数据
		startDate = today.AddDate(0, 0, -6).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = today.Format("2006-01-02")
	}
	s, e := service.FindByDate4(startDate, endDate)
	begin, end := *s, *e

	begin2, end2 := service.FindByDate4(startDate2, endDate2)

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

	var auditStatuss []int
	if auditStatus != "" && auditStatus != "0" {
		ids := strings.Split(auditStatus, ",")
		for _, _id := range ids {
			id, err := strconv.Atoi(_id)
			if err != nil || id == 0 {
				continue
			}
			auditStatuss = append(auditStatuss, id)
		}
	}
	var prizeTypess []int
	if prizeTypes != "" && prizeTypes != "0" {
		ids := strings.Split(prizeTypes, ",")
		for _, _id := range ids {
			id, err := strconv.Atoi(_id)
			if err != nil || id == 0 {
				continue
			}
			prizeTypess = append(prizeTypess, id)
		}
	}
	var auditUserss []string
	if auditUsers != "" && auditUsers != "-" {
		users := strings.Split(auditUsers, ",")
		for _, user := range users {
			if user != "" && user != "-" {
				auditUserss = append(auditUserss, user)
			}
		}
	} else {
		auditUsers = "-"
	}

	var err error
	var total int64
	var list any
	if typeId == 0 {
		// 转盘活动汇总
		list, err = service.ActivityService.GetActivityTurnDates(begin, end, tabId, utypeIds, packageIds)
	} else if typeId == 1 {
		// 转盘活动明细
		total, list, err = service.ActivityService.GetActivityTurnUsers(page, this.pageSize, userid, begin, end, begin2, end2, tabId, utypeIds, packageIds)
	} else if typeId == 2 {
		// 转盘活动审核
		total, list, err = service.ActivityService.GetActivityTurnAudits(page, this.pageSize, userid, begin, end, begin2, end2, utypeIds, packageIds, auditStatuss, prizeTypess, auditUserss)
	}
	if err != nil {
		beego.Error(err)
		this.showMsg(err.Error(), MSG_ERR)
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
	// packageList, _ := service.ChannelService.GetChannelAll()
	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()
	packageClassAll, _ := service.ChannelService.GetChannelClassAll()

	// 用户列表
	var usernames []string
	users, _ := service.UserService.GetUserList(1, 1000, true)
	for _, user := range users {
		usernames = append(usernames, user.UserName)
	}

	this.Data["list"] = list
	this.Data["pageTitle"] = "转盘汇总"
	this.Data["count"] = total
	this.Data["pageBar"] = libs.NewPager(page, int(total), this.pageSize, beego.URLFor("ActivityController.Turn", "typeId", typeId, "startDate", startDate, "endDate", endDate, "startDate2", startDate2, "endDate2", endDate2, "packageId", packageId, "typeId", typeId, "aliasId", aliasId, "classId", classId, "usertypeId", utypeId, "usertype", usertype, "tabId", tabId, "userid", userid, "auditStatus", auditStatus, "prizeTypes", prizeTypes, "auditUsers", auditUsers), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["startDate2"] = startDate2
	this.Data["endDate2"] = endDate2
	this.Data["packageId"] = packageId
	// this.Data["packageList"] = packageList
	this.Data["typeId"] = typeId
	this.Data["aliasId"] = aliasId
	this.Data["packageAliasList"] = packageAliasList
	this.Data["packageClassAll"] = packageClassAll
	this.Data["classId"] = classId
	this.Data["usertypeId"] = utypeId
	this.Data["usertype"] = usertype
	this.Data["tabId"] = tabId
	this.Data["userid"] = userid
	this.Data["auditStatus"] = auditStatus
	this.Data["prizeTypes"] = prizeTypes
	this.Data["auditUsers"] = auditUsers
	this.Data["usernames"] = usernames
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "turnaudit")
	this.display()
}

// 转盘奖励后台审核
func (this *ActivityController) TurnAudit() {
	orderids := this.GetString("id")
	state, _ := this.GetInt("state")
	if orderids == "" {
		this.checkError(errors.New("奖励ID不能为空"))
	}
	if state != 1 && state != 2 {
		this.checkError(errors.New("奖励审核状态有误"))
	}
	name := this.auth.GetUser().UserName

	orderidList := strings.Split(orderids, ",")
	msg := &data.ActivityTurnPrizeOperate{
		PrizeIds: orderidList,
		Op:       state,
		UserName: name,
	}
	result, err := service.GmRequest(pb.WebTurnAudit, pb.CONFIG_UPSERT, msg)
	beego.Trace("TurnPrizeAudit result: ", result)
	if err != nil {
		this.checkError(err)
	}
	service.ActionService.Add("turn_prize_audit", name,
		"", utils.String(orderids), utils.String(orderids), "")

	this.showMsg("操作成功！", MSG_OK, beego.URLFor("ActivityController.Turn", "typeId", 2))
}

func (this *ActivityController) TurnTag() {
	prizeId := this.GetString("id")
	tag := this.GetString("tag")
	if prizeId == "" {
		this.checkError(errors.New("奖励ID不能为空"))
	}
	if tag == "" {
		this.checkError(errors.New("标记信息不能为空"))
	}
	name := this.auth.GetUser().UserName

	err := service.ActivityService.TurnPrizeTag(name, prizeId, tag)
	if err != nil {
		this.checkError(err)
	}
	this.showMsg("操作成功！", MSG_OK, beego.URLFor("ActivityController.Turn", "typeId", 2))
	this.display()
}

// GiftPackCode 礼包码活动
func (this *ActivityController) GiftPackCode() {
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	status, _ := this.GetInt("status")
	ctimeSort := 1 // 默认降序
	if ctimeSortS := this.GetString("ctimeSort"); ctimeSortS != "" {
		ctimeSort, _ = strconv.Atoi(ctimeSortS)
	}

	if page < 1 {
		page = 1
	}
	// 活动数据
	today := service.NowTime()
	if startDate == "" && endDate == "" {
		// 默认近七天数据
		// startDate = today.AddDate(0, 0, -6).Format("2006-01-02")
		endDate = today.Format("2006-01-02")
	}
	m := service.FindByDate1(startDate, endDate, "ctime", "ctime")
	total, gifts, scoreSum, scoreReceivedSum, err := service.ActivityService.GiftPackCodeList(page, this.pageSize, m, status, ctimeSort == 0)
	if err != nil {
		beego.Error(err)
		this.showMsg(err.Error(), MSG_ERR)
	}

	this.Data["list"] = gifts
	this.Data["pageTitle"] = "礼包码"
	this.Data["count"] = total
	this.Data["pageBar"] = libs.NewPager(page, int(total), this.pageSize, beego.URLFor("ActivityController.GiftPackCode", "status", status, "startDate", startDate, "endDate", endDate, "ctimeSort", ctimeSort), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["status"] = status
	this.Data["ctimeSort"] = ctimeSort
	this.Data["scoreSum"] = scoreSum
	this.Data["scoreReceivedSum"] = scoreReceivedSum

	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "giftpackcodeedit")
	this.display()
}

// GiftPackCode 礼包码编辑
func (this *ActivityController) GiftPackCodeEdit() {
	editType := this.GetString("editType") // 2开关编辑

	// 开启关闭礼包码
	if editType == "2" {
		rsp := make(map[string]any)
		defer this.jsonResult(rsp)

		code := this.GetString("code")
		open, _ := this.GetInt("open")
		if !utils.SliceIn(open, 0, 1) {
			rsp["msg"] = "请求参数有误"
			return
		}

		gift := service.ActivityService.GetGiftPackCode(code)
		if gift.Code == "" {
			rsp["msg"] = "礼包码不存在"
			return
		}
		gift.Status = int32(open)

		if err := service.ActivityService.AddOrUpdateGiftPackCode(gift, false); err != nil {
			rsp["msg"] = "操作失败:" + err.Error()
			return
		}

		service.ActionService.Add("set_giftpackcode", this.auth.GetUser().UserName,
			"", code, "2", fmt.Sprintf("%s-%d", code, open))

		rsp["success"] = true
		return
	}

	if !this.isPost() {
		return
	}

	code := this.GetString("code")
	status, _ := this.GetInt("status")
	gift_type, _ := this.GetInt("gift_type")
	score_type, _ := this.GetInt("score_type")
	score, _ := this.GetInt("score")
	pieces, _ := this.GetInt("pieces")
	piece_min, _ := this.GetInt("piece_min")
	piece_max, _ := this.GetInt("piece_max")
	time_open := this.GetString("time_open")
	time_close := this.GetString("time_close")
	utypes := this.GetString("utypes")
	remark := this.GetString("remark")

	if code == "" {
		this.showMsg("码号不能为空", MSG_ERR)
	}

	regex := regexp.MustCompile(`^[a-zA-Z0-9]{4,8}$`)
	if !regex.MatchString(code) {
		this.showMsg("码号只能包含4-8位大小写字母数字", MSG_ERR)
	}
	if score <= 0 || pieces <= 0 || score < pieces {
		this.showMsg("总金额或总份数有误", MSG_ERR)
	}
	if gift_type == 1 && (piece_min <= 0 || piece_max <= 0 || piece_min > piece_max) {
		this.showMsg("随机码单份最大最小金额有误", MSG_ERR)
	}

	var isAdd bool
	gift := service.ActivityService.GetGiftPackCode(code)
	if editType == "1" {
		// 编辑
		if gift.Code == "" {
			this.showMsg("礼包码不存在", MSG_ERR)
		}
	} else {
		isAdd = true
		// 新增
		if gift.Code != "" {
			this.showMsg("礼包码已存在", MSG_ERR)
		}
		gift = new(entity.GiftPackCode)
	}

	gift.Code = code
	gift.Status = int32(status)
	gift.GiftType = int32(gift_type)
	gift.ScoreType = int32(score_type)
	gift.Score = int64(score)
	gift.Pieces = int32(pieces)
	gift.PieceMin = int64(piece_min)
	gift.PieceMax = int64(piece_max)
	timeOpen, err := time.ParseInLocation(utils.FORMAT2, time_open, service.Location())
	if err != nil {
		this.showMsg("有效期开始时间格式有误", MSG_ERR)
	}
	timeClose, err := time.ParseInLocation(utils.FORMAT2, time_close, service.Location())
	if err != nil {
		this.showMsg("有效期结束时间格式有误", MSG_ERR)
	}
	if timeClose.Before(timeOpen) {
		this.showMsg("有效期开始时间大于结束时间", MSG_ERR)
	}

	gift.TimeOpen = timeOpen.Unix()
	gift.TimeClose = timeClose.Unix()
	utypesArr := strings.Split(utypes, ",")
	var utypesi []int32
	for _, utype := range utypesArr {
		if utypei, err := strconv.Atoi(utype); err == nil {
			utypesi = append(utypesi, int32(utypei))
		}
	}
	gift.Utypes = utypesi
	gift.Cuser = this.auth.GetUser().UserName
	gift.Remark = remark

	if err := service.ActivityService.AddOrUpdateGiftPackCode(gift, isAdd); err != nil {
		this.showMsg("操作失败:"+err.Error(), MSG_ERR)
		return
	}

	service.ActionService.Add("set_giftpackcode", this.auth.GetUser().UserName,
		"", code, editType, code)

	this.redirect(beego.URLFor("ActivityController.GiftPackCode"))
}

// GiftPackCodeExport 礼包码导出
func (c *ActivityController) GiftPackCodeExport() {
	startDate := c.GetString("start_date")
	endDate := c.GetString("end_date")
	status, _ := c.GetInt("status")
	ctimeSort := 1 // 默认降序
	if ctimeSortS := c.GetString("ctimeSort"); ctimeSortS != "" {
		ctimeSort, _ = strconv.Atoi(ctimeSortS)
	}

	// 活动数据
	today := service.NowTime()
	if startDate == "" && endDate == "" {
		// 默认近七天数据
		// startDate = today.AddDate(0, 0, -6).Format("2006-01-02")
		endDate = today.Format("2006-01-02")
	}
	m := service.FindByDate1(startDate, endDate, "ctime", "ctime")
	_, gifts, _, _, err := service.ActivityService.GiftPackCodeList(1, 100000, m, status, ctimeSort == 0)
	if err != nil {
		beego.Error(err)
		c.showMsg(err.Error(), MSG_ERR)
	}

	headers := []string{"创建时间", "状态", "码号", "码类型", "奖金类型", "总金额", "总份数", "已领取金额", "已领取份数", "领完时间", "单份最小金额", "单份最大金额", "有效期", "可领取用户范围", "创建人", "礼包码标签"}
	var rows [][]any
	for _, v := range gifts {
		row := []any{
			v.FCtime,
			v.FStatus + utils.CaseElse(v.FStatus2 == "", "", "/"+v.FStatus2),
			v.FCode,
			v.FGiftType,
			v.FScoreType,
			v.FScore,
			v.Pieces,
			v.FScoreReceived,
			v.PiecesReceived,
			v.FFinishTime,
			v.FPieceMin,
			v.FPieceMax,
			fmt.Sprintf("%s-%s", v.FTimeOpen, v.FTimeClose),
			v.FUtypes,
			v.FCuser,
			v.Remark,
		}
		rows = append(rows, row)
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

	// 设置表格头
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

	// 构造文件名称
	title := "礼包码"
	fileName := title + "_" + time.Now().Format("20060102150405") + ".xlsx"
	// 解决文件名中文乱码问题
	fileName = path.Base(fileName)
	fileName = url.QueryEscape(fileName)
	c.Ctx.Output.Header("Content-type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Ctx.Output.Header("Content-Disposition", "attachment;filename="+fileName)
	c.Ctx.Output.Header("Pragma", "No-cache")
	c.Ctx.Output.Header("Cache-Control", "No-cache")
	c.Ctx.Output.Header("Expires", "0")
	if err := file.Write(c.Ctx.ResponseWriter); err != nil {
		logs.Error(err)
	}
}

// VolatilitySubsidy 波动返水
func (c *ActivityController) VolatilitySubsidy() {
	page, _ := strconv.Atoi(c.GetString("page"))
	tabid, _ := c.GetInt("tabid")
	startDate := c.GetString("start_date")
	endDate := c.GetString("end_date")
	// userid := this.GetString("userid")
	registArea, _ := c.GetInt("registArea")
	aliasId := c.GetString("alias_id")
	classId := c.GetString("class_id")
	userid := c.GetString("userid")
	sortBy := c.GetString("sortBy")
	asc := c.GetString("asc")

	if page < 1 {
		page = 1
	}

	// 活动数据
	today := service.NowTime()
	if startDate == "" {
		// 默认近七天数据
		startDate = today.AddDate(0, 0, -6).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = today.Format("2006-01-02")
	}
	s, e := service.FindByDate4(startDate, endDate)
	begin, end := *s, *e

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

	var err error
	var total int64
	var list any
	if tabid == 0 {
		stats, e := service.ActivityService.GetActivityVolatilitySubsidyDates(begin, end, registArea, packageIds)
		err = e
		list = stats
		total = int64(len(stats))
	} else if tabid == 1 {
		count, users, e := service.ActivityService.GetActivityVolatilitySubsidyUsers(page, c.pageSize, sortBy, asc, begin, end, registArea, packageIds)
		err = e
		list = users
		total = count
	}
	if err != nil {
		beego.Error(err)
		c.showMsg(err.Error(), MSG_ERR)
	}

	packageList, _ := service.ChannelService.GetChannelAll()
	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()
	packageClassAll, _ := service.ChannelService.GetChannelClassAll()

	c.Data["list"] = list
	c.Data["count"] = total
	c.Data["pageTitle"] = "波动返水"
	c.Data["pageBar"] = libs.NewPager(page, int(total), c.pageSize, beego.URLFor("ActivityController.VolatilitySubsidy", "tabid", tabid, "startDate", startDate, "endDate", endDate, "aliasId", aliasId, "classId", classId, "userid", userid, "registArea", registArea), true).ToString()
	c.Data["startDate"] = startDate
	c.Data["endDate"] = endDate
	c.Data["tabid"] = tabid
	c.Data["registArea"] = registArea
	c.Data["aliasId"] = aliasId
	c.Data["classId"] = classId
	c.Data["userid"] = userid
	c.Data["sortBy"] = sortBy
	c.Data["asc"] = asc
	c.Data["packageList"] = packageList
	c.Data["packageAliasList"] = packageAliasList
	c.Data["packageClassAll"] = packageClassAll
	c.Data["isOperation"] = c.auth.HasAccessPerm(c.controllerName, "giftpackcodeedit")
	c.display()
}

// VolatilitySubsidyExport 波动返水导出
func (c *ActivityController) VolatilitySubsidyExport() {
	tabid, _ := c.GetInt("tabid")
	startDate := c.GetString("start_date")
	endDate := c.GetString("end_date")
	registArea, _ := c.GetInt("registArea")
	aliasId := c.GetString("alias_id")
	classId := c.GetString("class_id")
	// userid := c.GetString("userid")
	sortBy := c.GetString("sortBy")
	asc := c.GetString("asc")

	// 活动数据
	today := service.NowTime()
	if startDate == "" {
		// 默认近七天数据
		startDate = today.AddDate(0, 0, -6).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = today.Format("2006-01-02")
	}
	s, e := service.FindByDate4(startDate, endDate)
	begin, end := *s, *e

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

	var err error
	var headers []string
	var rows [][]any
	if tabid == 0 {
		// _, list, e := service.ActivityService.BetRank(1, 100000, int32(rank_type), start, end)
		list, e := service.ActivityService.GetActivityVolatilitySubsidyDates(begin, end, registArea, packageIds)
		if e != nil {
			err = e
		} else {
			headers = []string{"日期", "日活用户数", "首充用户数", "领到补贴总人数", "首次领到补贴总人数", "首领占比", "领到补贴总次数", "人均领到次数", "当日首充用户领到人数", "当日首充用户被补贴率", "当日首充用户领到人数占总领人数比", "水池总金额", "补贴总金额", "水池消耗率", "人均补贴金额", "首次补贴用户次日留存率", "首次补贴用户3日留存率", "首次补贴用户5日留存率", "首次补贴用户7日留存率", "首充用户次日留存率", "首充用户3日留存率", "首充用户5日留存率", "首充用户7日留存率", "首次补贴用户7日人均付费总额", "首充7日人均付费总额", "当日补贴金额top123"}
			for _, v := range list {
				row := []any{
					v.SDate,
					v.LoginedUsers,
					v.FirstPayUsers,
					v.SubsidyUsers,
					v.FirstSubsidyUsers,
					v.FirstSubsidyRate,
					v.SubsidyTimes,
					v.SubsidyAvg,
					v.FirstPaySubsidyUsers,
					v.FirstPaySubsidyRate,
					v.FirstPaySubsidyUsersRate,
					v.SubsidyPool,
					v.SubsidyAmount,
					v.SubsidyConsumRate,
					v.SubsidyAmountAvg,
					v.FirstSubsidyRetentionDay1,
					v.FirstSubsidyRetentionDay3,
					v.FirstSubsidyRetentionDay5,
					v.FirstSubsidyRetentionDay7,
					v.FirstPayRetentionDay1,
					v.FirstPayRetentionDay3,
					v.FirstPayRetentionDay5,
					v.FirstPayRetentionDay7,
					v.FirstSubsidyPayAmountDay7,
					v.FirstPayAmountDay7Avg,
					strings.Join(v.SubsidyTop3, ";\r\n"),
				}
				rows = append(rows, row)
			}
		}
	} else if tabid == 1 {
		page, pageSize := 1, 100000
		_, list, e := service.ActivityService.GetActivityVolatilitySubsidyUsers(page, pageSize, sortBy, asc, begin, end, registArea, packageIds)

		if e != nil {
			err = e
		} else {
			headers = []string{"UID", "VIP等级", "用户类型", "注册日期", "存活天数", "流失天数", "首充时间", "首次补贴时间", "首次补贴距首充天数", "首充金额", "总充值金额", "总提现金额", "充-提", "总补贴金额", "总补贴次数"}
			for _, v := range list {
				row := []any{
					v.Userid,
					v.VipLv,
					v.ChargeType,
					v.RegistDate,
					v.LiveDays,
					v.LoseDays,
					v.FirstPayTime,
					v.FirstSubsidyTime,
					v.FirstPay2SubsidyDays,
					v.FirstPayAmount,
					v.Pays,
					v.Withdraws,
					v.PaySubWithdraws,
					v.Subsidys,
					v.SubsidyTimes,
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
	switch tabid {
	case 0:
		title += "波动返水汇总"
	case 1:
		title += "波动返水明细"
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

// ShareAgent 代理活动
func (this *ActivityController) ShareAgent() {
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	startDate2 := this.GetString("start_date2")
	endDate2 := this.GetString("end_date2")
	typeId, _ := this.GetInt("typeId")
	rank := this.GetString("rank")
	asc := this.GetString("asc")
	if page < 1 {
		page = 1
	}
	// 活动数据
	today := service.NowTime()
	if typeId == 3 {
		// 默认当前日期前一日
		if startDate == "" {
			startDate = today.AddDate(0, 0, -1).Format("2006-01-02")
		}
		if endDate == "" {
			endDate = today.AddDate(0, 0, -1).Format("2006-01-02")
		}
	} else {
		// 默认近七天数据
		if startDate == "" {
			startDate = today.AddDate(0, 0, -6).Format("2006-01-02")
		}
		if endDate == "" {
			endDate = today.Format("2006-01-02")
		}
	}
	s, e := service.FindByDate4(startDate, endDate)
	stime, etime := *s, *e

	stime2, etime2 := service.FindByDate4(startDate2, endDate2)

	var err error
	var count int
	var list any
	if typeId == 0 {
		// 裂变体系日报
		stats, e := service.ActivityService.GetShareAgentDates(stime, etime)
		err = e
		list = stats
		count = len(stats)
	} else if typeId == 1 {
		// 代理活动经济日览
		stats, e := service.ActivityService.ShareAgentFinanceDates(stime, etime)
		err = e
		list = stats
		count = len(stats)
	} else if typeId == 2 {
		// 代理头目明细
		if rank == "" {
			rank = "super_ctime"
		}
		if asc == "" {
			asc = "0"
		}
		total, stats, e := service.ActivityService.ShareAgentSuperStats(true, page, this.pageSize, stime, etime, stime2, etime2, rank, asc == "1")
		err = e
		list = stats
		count = int(total)
	} else if typeId == 3 {
		// 代理头目明细
		if rank == "" {
			rank = "prod_amounts"
		}
		if asc == "" {
			asc = "0"
		}
		total, stats, e := service.ActivityService.ShareAgentSuperBonusStats(true, page, this.pageSize, stime, etime, rank, asc == "1")
		err = e
		list = stats
		count = int(total)
	} else if typeId == 4 {
		// 代理团队每日数量变化
		stats, e := service.ActivityService.ShareAgentGroupVary(stime, etime)
		err = e
		list = stats
		count = len(stats)
	}

	if err != nil {
		beego.Error(err)
		this.showMsg(err.Error(), MSG_ERR)
	}

	if utils.SliceIn(typeId, 2, 3) {
		this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("ActivityController.ShareAgent", "typeId", typeId, "start_date", startDate, "end_date", endDate, "start_date2", startDate2, "end_date2", endDate2), true).ToString()
	}

	this.Data["list"] = list
	this.Data["pageTitle"] = "代理活动"
	this.Data["count"] = count
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["startDate2"] = startDate2
	this.Data["endDate2"] = endDate2
	this.Data["typeId"] = typeId
	this.Data["rank"] = rank
	this.Data["asc"] = asc
	this.Data["isExport"] = this.auth.HasAccessPerm(this.controllerName, "shareagentexport")
	this.display()
}

// ShareAgentExport 代理活动导出
func (this *ActivityController) ShareAgentExport() {
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	startDate2 := this.GetString("start_date2")
	endDate2 := this.GetString("end_date2")
	typeId, _ := this.GetInt("typeId")
	rank := this.GetString("rank")
	asc := this.GetString("asc")

	if page < 1 {
		page = 1
	}
	// 活动数据
	today := service.NowTime()
	if typeId == 3 {
		// 默认当前日期前一日
		if startDate == "" {
			startDate = today.AddDate(0, 0, -1).Format("2006-01-02")
		}
		if endDate == "" {
			endDate = today.AddDate(0, 0, -1).Format("2006-01-02")
		}
	} else {
		// 默认近七天数据
		if startDate == "" {
			startDate = today.AddDate(0, 0, -6).Format("2006-01-02")
		}
		if endDate == "" {
			endDate = today.Format("2006-01-02")
		}
	}
	s, e := service.FindByDate4(startDate, endDate)
	stime, etime := *s, *e

	stime2, etime2 := service.FindByDate4(startDate2, endDate2)

	var err error
	var headers []string
	var rows [][]any
	if typeId == 0 {
		// 裂变体系日报
		list, e := service.ActivityService.GetShareAgentDates(stime, etime)
		if e != nil {
			err = e
		} else {
			headers = []string{
				"日期", "当日产出总成本", "当日实付总成本", "当日ROI", "", "体系日活人数", "分享包日活人数", "体系日活的分享包占比", "新用户日活", "老用户日活", "新增代理头目人数", "",
				"体系充值总额", "体系提现金额", "充减提", "体系盈利率", "体系充值人数", "体系提现人数", "体系付费率", "体系提现人数占比", "体系ARPU", "体系ARPPU", "",
				"体系首充人数", "首充人充值总额", "首充人的提现人数", "首充人提现总额", "首充人的充减提", "首充人盈余率", "未充人首充率", "首充人提现占比", "",
				"体系新用户充值人数", "新用户充值金额", "新用户提现人数", "新用户提现总额", "新用户的充减提", "新用户盈余率", "新用户付费率", "新用户提现率", "新用户ARPU", "新用户ARPPU", "",
				"体系老用户充值人数", "老用户充值金额", "老用户提现人数", "老用户提现总额", "老用户的充减提", "老用户盈余率", "老用户付费率", "老用户提现率", "老用户ARPU", "老用户ARPPU",
			}
			for _, v := range list {
				rows = append(rows, []any{
					v.SDate,
					v.DayProductCost,
					v.DayPayOutCost,
					v.DayROI,
					"",
					v.ShareLoginUsers,
					v.ShareApkLoginUsers,
					v.ShareApkLoginRate,
					v.NewShareLoginUsers,
					v.OldShareLoginUsers,
					v.NewInviterUsers,
					"",
					v.SharePayAmounts,
					v.ShareWithdrawAmounts,
					v.SharePaySubWithdraw,
					v.ShareProfitRate,
					v.SharePayUsers,
					v.ShareWithdrawUsers,
					v.SharePayRate,
					v.ShareWithdrawRate,
					v.ShareARPU,
					v.ShareARPPU,
					"",
					v.ShareFirstPayUsers,
					v.ShareFirstPayAmounts,
					v.ShareFirstPayWithdrawUsers,
					v.ShareFirstPayWithdrawAmounts,
					v.ShareFirstPaySubWithdraw,
					v.ShareFirstPayProfit,
					v.ShareFirstPayUnpayRate,
					v.ShareFirstPayWithdrawRate,
					"",
					v.NewSharePayUsers,
					v.NewSharePayAmounts,
					v.NewShareWithdrawUsers,
					v.NewShareWithdrawAmounts,
					v.NewSharePaySubWithdraw,
					v.NewShareProfitRate,
					v.NewSharePayRate,
					v.NewShareWithdrawRate,
					v.NewShareARPU,
					v.NewShareARPPU,
					"",
					v.OldSharePayUsers,
					v.OldSharePayAmounts,
					v.OldShareWithdrawUsers,
					v.OldShareWithdrawAmounts,
					v.OldSharePaySubWithdraw,
					v.OldShareProfitRate,
					v.OldSharePayRate,
					v.OldShareWithdrawRate,
					v.OldShareARPU,
					v.OldShareARPPU,
				})
			}
		}
	} else if typeId == 1 {
		// 代理活动经济日览
		stats, e := service.ActivityService.ShareAgentFinanceDates(stime, etime)
		if e != nil {
			err = e
		} else {
			headers = []string{
				"日期", "当日产出总成本", "当日实付总成本", "当日产生打码返佣额", "当日领走打码返佣额", "当日产生单个人头奖金额", "当日领走单个人头奖金额", "当日产生累计人头奖金额", "当日领走累计人头奖金额",
				"", "当日裂变总人数", "属于1级团队的人数", "属于2级团队的人数", "属于3级团队的人数", "属于4级团队的人数",
				"", "当日裂变算人头奖人数", "属于1级团队的人数", "属于2级团队的人数", "属于3级团队的人数", "属于4级团队的人数",
				"", "单个人头奖励总金额", "当日邀请者单个人头奖励总金额", "当日受邀者单个人头奖励总金额", "属于1级团队的金额", "属于2级团队的金额", "属于3级团队的金额", "属于4级团队的金额",
				"", "累计人头奖励总金额", "属于1级团队的金额", "属于2级团队的金额", "属于3级团队的金额", "属于4级团队的金额",
				"", "打码返佣总额", "属于1级团队的金额", "属于2级团队的金额", "属于3级团队的金额", "属于4级团队的金额",
			}
			for _, v := range stats {
				rows = append(rows, []any{
					v.SDate,
					v.DayProductCost,
					v.DayPayOutCost,
					v.ShareBetsReward,
					v.ShareBetsRewardTake,
					v.ShareHeadReward,
					v.ShareHeadRewardTake,
					v.ShareTaskReward,
					v.ShareTaskRewardTake,
					"",
					v.ShareUsers,
					v.ShareTeamV1Users,
					v.ShareTeamV2Users,
					v.ShareTeamV3Users,
					v.ShareTeamV4Users,
					"",
					v.ShareInvalidUsers,
					v.ShareTeamV1InvalidUsers,
					v.ShareTeamV2InvalidUsers,
					v.ShareTeamV3InvalidUsers,
					v.ShareTeamV4InvalidUsers,
					"",
					v.ShareHeadReward2,
					v.ShareHeadRewardSuper,
					v.ShareHeadRewardChild,
					v.ShareTeamV1HeadReward,
					v.ShareTeamV2HeadReward,
					v.ShareTeamV3HeadReward,
					v.ShareTeamV4HeadReward,
					"",
					v.ShareTaskReward2,
					v.ShareTeamV1TaskReward,
					v.ShareTeamV2TaskReward,
					v.ShareTeamV3TaskReward,
					v.ShareTeamV4TaskReward,
					"",
					v.ShareBetsReward2,
					v.ShareTeamV1BetsReward,
					v.ShareTeamV2BetsReward,
					v.ShareTeamV3BetsReward,
					v.ShareTeamV4BetsReward,
				})
			}
		}
		// total, list, err = service.ActivityService.GetActivityTurnUsers(page, this.pageSize, userid, begin, end, begin2, end2, tabId, utypeIds, packageIds)
	} else if typeId == 2 {
		// 代理头目明细
		if rank == "" {
			rank = "super_ctime"
		}
		if asc == "" {
			asc = "0"
		}
		_, stats, e := service.ActivityService.ShareAgentSuperStats(false, page, 100000, stime, etime, stime2, etime2, rank, asc == "1")
		if e != nil {
			err = e
		} else {
			headers = []string{
				"头目UID", "注册日期", "成为头目的日期", "发育效率", "团队死亡天数", "团头总ROI", "团队ROI", "",
				"产生打码返佣总额", "产生转盘奖金总额", "产生单个人头奖励总额（给头目）", "产生单个人头奖励总额（给团队）", "产生累计人头奖励总额", "领走打码返佣总额", "领走转盘奖金总额", "领走单个人头奖励总额（给头目）", "领走单个人头奖励总额（给团队）", "领走累计人头奖励总额", "",
				"头目总充值", "头目总提现", "头目充-提", "头目盈余率", "头目总打码", "头目充投比", "",
				"团队等级", "团队人数", "团队付费人数", "团队付费率", "团队总充值", "团队总提现", "团队充-提", "团队盈余率", "团队总打码", "团队充投比", "团队arpu", "团队arppu",
			}
			for _, v := range stats {
				rows = append(rows, []any{
					v.SuperId,
					v.SuperCtime,
					v.BeSuperTime,
					v.GrowDays,
					v.LastLoginDays,
					v.SuperROI,
					v.TeamROI,
					"",
					v.ProdAmountsBet,
					v.TurnPrize,
					v.ProdAmountsHead,
					v.ProdAmountsChild,
					v.ProdAmountsTask,
					v.TakeAmountsBet,
					v.TakeTurnPrize,
					v.TakeAmountsHead,
					v.TakeAmountsChild,
					v.TakeAmountsTask,
					"",
					v.SuperPays,
					v.SuperCashOut,
					v.SuperProfit,
					v.SuperProfitRate,
					v.SuperBets,
					v.SuperBetsPayRate,
					"",
					v.TeamLv,
					v.GroupUsers,
					v.GroupPayUsers,
					v.GroupPayRate,
					v.GroupPays,
					v.GroupCashOut,
					v.GroupProfit,
					v.GroupProfitRate,
					v.GroupBets,
					v.GroupBetsPayRate,
					v.GroupArpu,
					v.GroupArppu,
				})
			}
		}
	} else if typeId == 3 {
		// 代理头目奖金日排名
		if rank == "" {
			rank = "prod_amounts"
		}
		if asc == "" {
			asc = "0"
		}
		_, stats, e := service.ActivityService.ShareAgentSuperBonusStats(false, page, 100000, stime, etime, rank, asc == "1")
		err = e
		headers = []string{
			"日期", "头目UID", "团头总ROI", "团队ROI", "当日总产生奖金", "当日总领走奖金", "产生打码返佣总额", "产生转盘奖金总额", "产生单个人头奖励总额（给头目）", "产生单个人头奖励总额（给团队）", "产生累计人头奖励总额", "领走打码返佣总额", "领走转盘奖金总额", "领走单个人头奖励总额（给头目）", "领走单个人头奖励总额（给团队）", "领走累计人头奖励总额",
		}
		for _, v := range stats {
			rows = append(rows, []any{
				v.SDate,
				v.SuperId,
				v.SuperROI,
				v.TeamROI,
				v.ProdAmounts,
				v.TakeAmounts,
				v.ProdAmountsBet,
				v.TurnPrize,
				v.ProdAmountsHead,
				v.ProdAmountsChild,
				v.ProdAmountsTask,
				v.TakeAmountsBet,
				v.TakeTurnPrize,
				v.TakeAmountsHead,
				v.TakeAmountsChild,
				v.TakeAmountsTask,
			})
		}
	} else if typeId == 4 {
		// 代理团队每日数量变化
		stats, e := service.ActivityService.ShareAgentGroupVary(stime, etime)
		err = e
		if err == nil {
			headers = []string{
				"日期", "1级团队总个数", "1级死团队个数", "1级活团队个数", "1级团队总人数", "1级团队队均人数", "1级死团队人数", "1级活团队人数", "1级团队死亡率", "1级团队人数流失率",
				"", "2级团队总个数", "2级死团队个数", "2级活团队个数", "2级团队总人数", "2级团队队均人数", "2级死团队人数", "2级活团队人数", "2级团队死亡率", "2级团队人数流失率",
				"", "3级团队总个数", "3级死团队个数", "3级活团队个数", "3级团队总人数", "3级团队队均人数", "3级死团队人数", "3级活团队人数", "3级团队死亡率", "3级团队人数流失率",
				"", "4级团队总个数", "4级死团队个数", "4级活团队个数", "4级团队总人数", "4级团队队均人数", "4级死团队人数", "4级活团队人数", "4级团队死亡率", "4级团队人数流失率",
			}
			for _, v := range stats {
				rows = append(rows, []any{
					v.SDate,
					v.Lv1Groups,
					v.Lv1DeadGroups,
					v.Lv1LiveGroups,
					v.Lv1GroupUsers,
					v.Lv1GroupAvgUsers,
					v.Lv1DeadGroupUsers,
					v.Lv1LiveGroupUsers,
					v.Lv1GroupDeadRate,
					v.Lv1GroupLossRate,
					"",
					v.Lv2Groups,
					v.Lv2DeadGroups,
					v.Lv2LiveGroups,
					v.Lv2GroupUsers,
					v.Lv2GroupAvgUsers,
					v.Lv2DeadGroupUsers,
					v.Lv2LiveGroupUsers,
					v.Lv2GroupDeadRate,
					v.Lv2GroupLossRate,
					"",
					v.Lv3Groups,
					v.Lv3DeadGroups,
					v.Lv3LiveGroups,
					v.Lv3GroupUsers,
					v.Lv3GroupAvgUsers,
					v.Lv3DeadGroupUsers,
					v.Lv3LiveGroupUsers,
					v.Lv3GroupDeadRate,
					v.Lv3GroupLossRate,
					"",
					v.Lv4Groups,
					v.Lv4DeadGroups,
					v.Lv4LiveGroups,
					v.Lv4GroupUsers,
					v.Lv4GroupAvgUsers,
					v.Lv4DeadGroupUsers,
					v.Lv4LiveGroupUsers,
					v.Lv4GroupDeadRate,
					v.Lv4GroupLossRate,
				})
			}
		}
	}

	if err != nil {
		beego.Error(err)
		this.showMsg(err.Error(), MSG_ERR)
	}

	// 检查是否有数据可导出
	if len(rows) == 0 {
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
	var title string
	switch typeId {
	case 0:
		title = "裂变体系日报"
	case 1:
		title = "代理活动经济日览"
	case 2:
		title = "代理头目明细"
	case 3:
		title = "代理头目奖金日排名"
	case 4:
		title = "代理团队每日数量变化"
	}
	fileName := title + "_" + time.Now().Format("20060102150405") + ".xlsx"
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

// UserCoupon 满减券
func (c *ActivityController) UserCoupon() {
	action := c.GetString("action")

	if action == "Aside" {
		// 渠道、渠道类
		packageAliases, _ := service.ChannelService.GetChannelAliasAll(false)
		packageClasses, _ := service.ChannelService.GetChannelClassAll(false)

		usertype := []map[string]any{
			// {"key": 0, "name": "全部"},
			{"key": 0, "name": "新手"},
			{"key": 1, "name": "平民"},
			{"key": 2, "name": "普R"},
			{"key": 3, "name": "小R"},
			{"key": 4, "name": "中R"},
			{"key": 5, "name": "大R"},
			{"key": 6, "name": "超大R"},
		}

		// 默认七天数据
		etime := service.NowTime()
		stime := etime.AddDate(0, 0, -6)

		c.JsonRSuccess(libs.R{
			"userTypes":      usertype,
			"packageAliases": packageAliases,
			"packageClasses": packageClasses,
			"isOperation":    c.auth.HasAccessPerm(c.controllerName, "usercouponopt"),
			"isExport":       c.auth.HasAccessPerm(c.controllerName, "usercouponexport"),
			"version":        service.GetVersions(),
			"etime":          etime.Format(utils.FORMAT_DATE),
			"stime":          stime.Format(utils.FORMAT_DATE),
		})
		return
	} else if action == "AddCoupon" {
		// 新建优惠券
		var form = struct {
			MinRecharge   int64   `json:"minRecharge"`
			DerateAmounts int64   `json:"derateAmounts"`
			ValidHours    int32   `json:"validHours"`
			Count         int     `json:"count"`
			ClassifyType  []int   `json:"classifyType"`
			RegistArea    []int   `json:"ra"`
			SendTime      string  `json:"sendTime"`
			PayAvgMin     int64   `json:"payAvgMin"`
			PayAvgMax     int64   `json:"payAvgMax"`
			ProfitRateMin float64 `json:"profitRateMin"`
			ProfitRateMax float64 `json:"profitRateMax"`
			UserIds       string  `json:"userIds"`
			UserTag       []int   `json:"userTag"`
		}{}
		c.JsonBody(&form)

		if form.MinRecharge < 100 {
			c.JsonRError("最低充值不能小于100")
			return
		}

		if form.DerateAmounts <= 0 {
			c.JsonRError("减免金额不能小于0")
			return
		}

		if form.SendTime == "" {
			c.JsonRError("发放时间不能为空")
			return
		}

		// 生成优惠券ID
		couponId, err := service.ActivityService.CouponGenerateID()
		if err != nil {
			c.JsonRError("生成优惠券ID失败")
			return
		}

		// 创建优惠券
		var ids []string
		if form.UserIds != "" {
			ids = strings.Split(form.UserIds, ",")
		}
		sendTime, _ := time.ParseInLocation(utils.FORMAT, form.SendTime, service.Location())
		coupon := &entity.ManualCoupon{
			Id:                 couponId,
			MinRecharge:        int64(form.MinRecharge) * 100,   // 转换为分
			Discount:           int64(form.DerateAmounts) * 100, // 转换为分
			Ctime:              time.Now().Unix(),
			ValidityTime:       int32(form.ValidHours),
			Num:                int32(form.Count),
			OrderAvgRange:      []int64{int64(form.PayAvgMin), int64(form.PayAvgMax)},
			ProfitabilityRatio: []int64{int64(form.ProfitRateMin), int64(form.ProfitRateMax)},
			SendTime:           sendTime.Unix(),
			RecharClassify:     form.ClassifyType,
			AcctType:           form.RegistArea,
			UserTag:            form.UserTag,
			UserIds:            ids,
		}

		err = service.ActivityService.SaveManualCoupon(coupon)
		if err != nil {
			c.JsonRError(err.Error())
			return
		}
		// 同步到游戏
		if err = mq.NatsPublish(mq.TopicActivityCouponUpdate, &pb.ManualCouponUpdate{
			CouponId: couponId,
		}); err != nil {
			glog.Errorf("publish user share regist error: %s, %v", couponId, err)
			c.JsonRError(err.Error())
			return
		}
		c.JsonRSuccess(nil)
		return
	}

	// 满减券汇总
	if action == "CoupnStats" {
		arg := new(args.CoupnStatsArgs)
		c.JsonBody(arg)

		var stime, etime time.Time
		var err error
		if stime, err = time.ParseInLocation(utils.FORMAT, fmt.Sprintf("%s 00:00:00", arg.CtimeS), service.Location()); err != nil {
			c.JsonRError(err.Error())
			return
		}
		if etime, err = time.ParseInLocation(utils.FORMAT, fmt.Sprintf("%s 23:59:59", arg.CtimeE), service.Location()); err != nil {
			c.JsonRError(err.Error())
			return
		}
		var packageIds []string
		classPkgs, err := service.ChannelService.GetChannelByClassList(arg.ChannelClasses)
		if err != nil {
			c.JsonRError(err.Error())
			return
		}
		aliasPkgs, err := service.ChannelService.GetChannelByNameAliasList(arg.AliasIds)
		if err != nil {
			c.JsonRError(err.Error())
			return
		}
		for _, pkg := range classPkgs {
			packageIds = append(packageIds, pkg.Name)
		}
		for _, pkg := range aliasPkgs {
			packageIds = append(packageIds, pkg.Name)
		}
		var registAreas []int32
		if arg.RegistArea > 0 {
			registAreas = append(registAreas, arg.RegistArea-1)
		}
		stats, err := service.ActivityService.UserCouponDates(stime, etime, registAreas, arg.UserTypes, packageIds)
		if err != nil {
			c.JsonRError(err.Error())
			return
		}
		c.JsonRSuccess(stats)
		return
	}

	// 手动发放优惠券
	if action == "CoupnHand" {
		arg := new(args.CoupnHandArgs)
		c.JsonBody(arg)

		coupons, err := service.ActivityService.UserCouponHand(arg)
		if err != nil {
			c.JsonRError(err.Error())
			return
		}
		c.JsonRSuccess(coupons)
		return
	}
	// 开关
	if action == "EditSwitch" {
		couponId := c.GetString("couponId")
		open, _ := c.GetBool("open")
		err := service.ActivityService.EditSwitch(couponId, open)
		if err != nil {
			c.JsonRError(err.Error())
			return
		}
		// 同步到游戏
		if err = mq.NatsPublish(mq.TopicActivityCouponUpdate, &pb.ManualCouponUpdate{
			CouponId: couponId,
		}); err != nil {
			glog.Errorf("publish user share regist error: %s, %v", couponId, err)
			c.JsonRError(err.Error())
			return
		}
		c.JsonRSuccess(nil)
		return
	}

	// 编辑
	if action == "EditCoupon" {
		coupon := new(entity.ManualCoupon)
		c.JsonBody(coupon)
		err := service.ActivityService.EditCoupon(coupon)
		if err != nil {
			c.JsonRError(err.Error())
			return
		}

		// 同步到游戏
		if err = mq.NatsPublish(mq.TopicActivityCouponUpdate, &pb.ManualCouponUpdate{
			CouponId: coupon.Id,
		}); err != nil {
			glog.Errorf("publish user share regist error: %s, %v", coupon.Id, err)
			c.JsonRError(err.Error())
			return
		}
		c.JsonRSuccess(nil)
		return
	}

	if action == "CoupnHandUsersUpload" {
		file, header, err := c.GetFile("file")
		beego.Info("CoupnHandUsersUpload:", header.Filename)
		if err != nil {
			beego.Error(err)
			c.JsonRError("请选择需要上传的图片")
			return
		}
		defer file.Close()

		excel, err := excelize.OpenReader(file)
		if err != nil {
			beego.Error(err)
			c.JsonRError(fmt.Sprintf("文件解析失败: %s", err.Error()))
			return
		}
		defer excel.Close()

		sheets := excel.GetSheetList()
		if len(sheets) == 0 {
			c.JsonRError("空excel")
			return
		}
		sheet := sheets[0]
		rows, err := excel.GetRows(sheet)
		if err != nil {
			c.JsonRError(fmt.Sprintf("读取行数据失败: %v", err))
			return
		}
		var userids []string
		for _, row := range rows {
			if len(row) > 0 && row[0] != "" {
				userids = append(userids, row[0])
			}
		}
		if len(userids) == 0 {
			c.JsonRError("未找到用户id数据")
			return
		}
		c.JsonRSuccess(userids)
		return
	}

	c.Data["pageTitle"] = "满减券"
	c.display()
}

// UserCouponExport 满减券导出
func (c *ActivityController) UserCouponExport() {
	action := c.GetString("action")

	var rows [][]any
	var headers []string

	// 满减券汇总
	if action == "CoupnStats" {
		arg := new(args.CoupnStatsArgs)
		c.JsonBody(arg)

		var stime, etime time.Time
		var err error
		if stime, err = time.ParseInLocation(utils.FORMAT, fmt.Sprintf("%s 00:00:00", arg.CtimeS), service.Location()); err != nil {
			c.JsonRError(err.Error())
			return
		}
		if etime, err = time.ParseInLocation(utils.FORMAT, fmt.Sprintf("%s 23:59:59", arg.CtimeE), service.Location()); err != nil {
			c.JsonRError(err.Error())
			return
		}
		var packageIds []string
		classPkgs, err := service.ChannelService.GetChannelByClassList(arg.ChannelClasses)
		if err != nil {
			c.JsonRError(err.Error())
			return
		}
		aliasPkgs, err := service.ChannelService.GetChannelByNameAliasList(arg.AliasIds)
		if err != nil {
			c.JsonRError(err.Error())
			return
		}
		for _, pkg := range classPkgs {
			packageIds = append(packageIds, pkg.Name)
		}
		for _, pkg := range aliasPkgs {
			packageIds = append(packageIds, pkg.Name)
		}
		var registAreas []int32
		if arg.RegistArea > 0 {
			registAreas = append(registAreas, arg.RegistArea-1)
		}
		stats, err := service.ActivityService.UserCouponDates(stime, etime, registAreas, arg.UserTypes, packageIds)
		if err != nil {
			c.JsonRError(err.Error())
			return
		}

		headers = []string{"日期", "当日发放总张数", "当日发放总最低要求金额", "当日发放总减免金额", "当日用券支付总张数", "当日用券支付总金额", "当日用券减免总金额", "总发券使用率", "当日总发当日即用率", "", "当日自动发放总张数", "当日自动发放总最低要求金额", "当日自动发放总减免金额", "当日用自动券支付总张数", "当日用自发券支付总金额", "当日用自发券减免总金额", "自发券使用率", "当日自发当日即用率", "", "当日手动发放总张数", "当日手动发放总最低要求金额", "当日手动发放总减免金额", "当日用手发券支付总张数", "当日用手发券支付总金额", "当日用手发券减免总金额", "手发券使用率", "当日手发当日即用率"}
		for _, v := range stats {
			rows = append(rows, []any{
				v.SDate,
				v.GiveCoupons,
				v.MinRecharge,
				v.DerateAmounts,
				v.UseCouponOrders,
				v.UseCouponOrderAmounts,
				v.UseCouponDerateAmounts,
				v.UseRate,
				v.CurrentUseRate,
				"",
				v.AutoGiveCoupons,
				v.AutoMinRecharge,
				v.AutoDerateAmounts,
				v.AutoUseCouponOrders,
				v.AutoUseCouponOrderAmounts,
				v.AutoUseCouponDerateAmounts,
				v.AutoUseRate,
				v.AutoCurrentUseRate,
				"",
				v.HandGiveCoupons,
				v.HandMinRecharge,
				v.HandDerateAmounts,
				v.HandUseCouponOrders,
				v.HandUseCouponOrderAmounts,
				v.HandUseCouponDerateAmounts,
				v.HandUseRate,
				v.HandCurrentUseRate,
			})
		}
	}

	// 检查是否有数据可导出
	if len(rows) == 0 {
		c.Ctx.Output.SetStatus(http.StatusRequestedRangeNotSatisfiable) // 416请求范围无效
		c.JsonRFail("未查询到可以导出的数据")
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
	title := "充值记录"
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
		beego.Error(err)
	}
}
