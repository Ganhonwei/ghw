package controllers

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"goserver/gen/pb"
	"goserver/internal/web/admin/app/entity"
	"goserver/internal/web/admin/app/libs"
	"goserver/internal/web/admin/app/service"
	"goserver/pkg/utils"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/validation"
	"github.com/dhushon/decxls"
	"github.com/globalsign/mgo/bson"
)

type PlayerController struct {
	BaseController
}

func (this *PlayerController) List() {
	status, _ := this.GetInt("status")
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	userid := this.GetString("userid")
	packageId := this.GetString("package_id")
	typeId, _ := this.GetInt("type_id")
	usertypeId, _ := this.GetInt("utype_id")
	//最后登录时间
	startDate1 := this.GetString("start_date1")
	endDate1 := this.GetString("end_date1")
	accountId, _ := this.GetInt("account_id")
	numStatr := this.GetString("num_statr")
	numEnd := this.GetString("num_end")
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
		// 默认当天数据
		if userid == "" && startDate == "" && endDate == "" {
			today := bson.Now()
			startDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
			endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
		}
		status = 1
	}
	m := service.FindByDate(startDate, endDate, "ctime", "ctime")

	if accountId != 0 {
		if accountId == 2 {
			m["regist_area"] = 1
		} else if accountId == 3 {
			m["regist_area"] = 2
		} else {
			m["regist_area"] = bson.M{"$nin": []int{1, 2}}
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

	if startDate1 != "" || endDate1 != "" {
		s := fmt.Sprintf("%s 00:00:00", startDate1)
		startTime1 := utils.Str2Time(s, service.Location())
		e := fmt.Sprintf("%s 23:59:59", endDate1)
		endTime1 := utils.Str2Time(e, service.Location())
		if !startTime1.IsZero() && !endTime1.IsZero() && startDate1 != "" && endDate1 != "" {
			m["login_time"] = bson.M{"$gte": startTime1, "$lte": endTime1}
		} else if !startTime1.IsZero() && startDate1 != "" {
			m["login_time"] = bson.M{"$gte": startTime1}
		} else if !endTime1.IsZero() && endDate1 != "" &&
			endDate != "" {
			m["login_time"] = bson.M{"$lte": endTime1}
		}
	}
	if usertypeId != 0 {
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
				m["money"] = bson.M{"$gte": startMoney}
				m["state"] = 2
				m["custom_types"] = bson.M{"$eq": ""}
				m["custom_types"] = bson.M{"$eq": nil}
			} else {
				m["money"] = bson.M{"$gte": startMoney, "$lte": endMoney}
				m["state"] = 2
				m["custom_types"] = bson.M{"$eq": ""}
				m["custom_types"] = bson.M{"$eq": nil}
			}
		} else {
			if usertypeId == 11 {
				m["custom_types"] = bson.M{"$ne": ""}
				m["custom_types"] = bson.M{"$ne": nil}
				// m["custom_types"] = bson.M{"$ne": bson.M{}, "$exists": true}
				// m["custom_types"] = bson.M{"$ne": pb.NULL}
			} else {
				m["state"] = usertypeId
				m["custom_types"] = bson.M{"$eq": ""}
				m["custom_types"] = bson.M{"$eq": nil}
			}

		}
	}
	// 局数筛选
	if numStatr != "" || numEnd != "" {
		startNum, _ := strconv.Atoi(numStatr)
		endNum, _ := strconv.Atoi(numEnd)
		if startNum != 0 || endNum != 0 {
			Idarr, _ := service.PlayerService.GetUserGameNumber(startNum, endNum)
			m["_id"] = bson.M{"$in": Idarr}
		}

	}
	// 提高查询优先级
	if userid != "" {
		if typeId == 0 {
			m["_id"] = userid
		} else if typeId == 1 {
			m["phone"] = userid
		} else if typeId == 2 {
			m["nickname"] = bson.M{"$regex": userid, "$options": "i"}
		} else if typeId == 3 {
			// m["login_ip"] = userid
			m["$or"] = []bson.M{
				{"login_ip": bson.M{"$regex": userid, "$options": "i"}},
				{"regist_ip": bson.M{"$regex": userid, "$options": "i"}},
			}
		} else if typeId == 4 {
			m["ad__adid"] = userid
		} else if typeId == 5 {
			// 银行卡号
			m["bank_accounts"] = userid
		}
	}
	m["robot"] = false
	m["simulation_robot"] = false
	count, _ := service.PlayerService.GetTotal(m)
	list, _ := service.PlayerService.GetList(page, this.pageSize, m)
	// list, count, _ := service.PlayerService.GetListCk(page, this.pageSize, params)

	packageList, _ := service.ChannelService.GetChannelAll()
	typeList := map[int]string{
		0: "用户ID",
		1: "手机号",
		2: "用户昵称",
		3: "IP",
		4: "设备号",
		5: "银行卡号",
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

	accountType := map[int]string{
		0: "全部",
		1: "A类",
		2: "B类",
		3: "C类",
	}

	this.Data["pageTitle"] = "用户列表"
	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["packageList"] = packageList
	this.Data["packageId"] = packageId
	this.Data["userid"] = userid
	this.Data["typeList"] = typeList
	this.Data["typeId"] = typeId
	this.Data["usertype"] = usertype
	this.Data["usertypeId"] = usertypeId
	this.Data["status"] = status
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("PlayerController.List", "status", status, "userid", userid, "type_id", typeId, "package_id", packageId, "utype_id", usertypeId, "account_id", "num_statr", numStatr, "num_end", numEnd, accountId, "start_date", startDate, "end_date", endDate, "start_date1", startDate1, "end_date1", endDate1), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["startDate1"] = startDate1
	this.Data["endDate1"] = endDate1
	this.Data["accountType"] = accountType
	this.Data["accountId"] = accountId
	this.Data["num_statr"] = numStatr
	this.Data["num_end"] = numEnd
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "useropt")
	this.display()
}

// 用户列表
func (this *PlayerController) ListCK() {
	status, _ := this.GetInt("status")
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	userid := this.GetString("userid")
	packageId := this.GetString("package_id")
	typeId, _ := this.GetInt("type_id")
	usertypeId, _ := this.GetInt("utype_id")
	//最后登录时间
	startDate1 := this.GetString("start_date1")
	endDate1 := this.GetString("end_date1")
	accountId, _ := this.GetInt("account_id")
	numStatr := this.GetString("num_statr")
	numEnd := this.GetString("num_end")
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
		// 默认当天数据
		if userid == "" && startDate == "" && endDate == "" {
			today := bson.Now()
			startDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
			endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
		}
		status = 1
	}
	var params = make(map[string]any)
	if startDate != "" || endDate != "" {
		// startTime, endTime := service.FindByDate4(startDate, endDate)
		if startDate != "" {
			startTime := fmt.Sprintf("%s 00:00:00", startDate)
			params["starttime"] = startTime
		}
		if endDate != "" {
			endTime := fmt.Sprintf("%s 00:00:00", endDate)
			params["endtime"] = endTime
		}
	}
	if accountId != 0 {
		if accountId == 2 {
			params["regist_area"] = []int{1}
		} else if accountId == 3 {
			params["regist_area"] = []int{2}
		} else {
			params["regist_area"] = []int{1, 2}
		}
	}

	if packageId != "" && packageId != "0" && packageId != "-" {
		temp_arr := make([]string, 0)
		if strings.Contains(packageId, ",") {
			palkage_arr := strings.Split(packageId, ",")
			for _, item := range palkage_arr {
				if item != "" && item != "0" && item != "-" {
					temp_arr = append(temp_arr, item)
				}
			}
			params["ad__bundle_id"] = temp_arr
		} else {
			temp_arr = append(temp_arr, packageId)
			params["ad__bundle_id"] = temp_arr
		}
		// 存储缓存
		this.SetSession("my_select_pakeageid", packageId)
	} else {
		// 清除缓存
		this.DelSession("my_select_pakeageid")

	}

	if startDate1 != "" || endDate1 != "" {
		if startDate1 != "" {
			startTime1 := fmt.Sprintf("%s 00:00:00", startDate1)
			params["login_start_time"] = startTime1
		}
		if endDate1 != "" {
			endTime1 := fmt.Sprintf("%s 00:00:00", endDate1)
			params["login_end_time"] = endTime1
		}
	}
	if usertypeId != 0 {
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
				params["money_start"] = startMoney
				params["state"] = 2
				params["custom_types"] = bson.M{"$eq": ""}
				params["custom_types"] = bson.M{"$eq": nil}
			} else {
				params["money_start"] = startMoney
				params["money_end"] = endMoney
				// bson.M{"$gte": startMoney, "$lte": endMoney}
				params["state"] = 2
				params["custom_types"] = bson.M{"$eq": ""}
				params["custom_types"] = bson.M{"$eq": nil}
			}
		} else {
			if usertypeId == 11 {
				params["custom_types"] = bson.M{"$ne": ""}
				params["custom_types"] = bson.M{"$ne": nil}
			} else {
				params["state"] = usertypeId
				params["custom_types"] = bson.M{"$eq": ""}
				params["custom_types"] = bson.M{"$eq": nil}
			}

		}
	}
	// 局数筛选
	if numStatr != "" || numEnd != "" {
		startNum, _ := strconv.Atoi(numStatr)
		endNum, _ := strconv.Atoi(numEnd)
		if startNum != 0 || endNum != 0 {
			Idarr, _ := service.PlayerService.GetUserGameNumber(startNum, endNum)
			params["_id"] = bson.M{"$in": Idarr}
		}

	}
	// 提高查询优先级
	if userid != "" {
		if typeId == 0 {
			params["_id"] = userid
		} else if typeId == 1 {
			params["phone"] = userid
		} else if typeId == 2 {
			params["nickname"] = userid
		} else if typeId == 3 {
			params["ip"] = userid
		} else if typeId == 4 {
			params["ad__adid"] = userid
		} else if typeId == 5 {
			// 银行卡号
			params["bank_accounts"] = userid
		}
	}
	list, count, _ := service.PlayerService.GetListCk(page, this.pageSize, params)

	packageList, _ := service.ChannelService.GetChannelAll()
	typeList := map[int]string{
		0: "用户ID",
		1: "手机号",
		2: "用户昵称",
		3: "IP",
		4: "设备号",
		5: "银行卡号",
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

	accountType := map[int]string{
		0: "全部",
		1: "A类",
		2: "B类",
		3: "C类",
	}

	this.Data["pageTitle"] = "用户列表"
	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["packageList"] = packageList
	this.Data["packageId"] = packageId
	this.Data["userid"] = userid
	this.Data["typeList"] = typeList
	this.Data["typeId"] = typeId
	this.Data["usertype"] = usertype
	this.Data["usertypeId"] = usertypeId
	this.Data["status"] = status
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("PlayerController.List", "status", status, "userid", userid, "type_id", typeId, "package_id", packageId, "utype_id", usertypeId, "account_id", "num_statr", numStatr, "num_end", numEnd, accountId, "start_date", startDate, "end_date", endDate, "start_date1", startDate1, "end_date1", endDate1), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["startDate1"] = startDate1
	this.Data["endDate1"] = endDate1
	this.Data["accountType"] = accountType
	this.Data["accountId"] = accountId
	this.Data["num_statr"] = numStatr
	this.Data["num_end"] = numEnd
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "useropt")
	this.display()
}

// 修改玩家资料
func (this *PlayerController) Edit() {
	userid := this.GetString("id")
	p, err := service.PlayerService.GetPlayer(userid)
	this.checkError(err)
	if this.isPost() {
		// photo := this.GetString("photo")
		nickname := this.GetString("nickname")
		// phone := this.GetString("phone")
		phone2 := this.GetString("phone2")
		bankname := this.GetString("bankname")
		accounts := this.GetString("accounts")
		ifsc := this.GetString("ifsc")
		realname := this.GetString("realname")
		afid := this.GetString("afid")
		os := this.GetString("os")
		bundleId := this.GetString("bundleId")
		afKey := this.GetString("afKey")
		mediasource := this.GetString("mediasource")
		appid := this.GetString("appid")
		refgameid := this.GetString("refgameid")
		refpkgname := this.GetString("refpkgname")

		// user.Photo = photo
		p.Nickname = nickname
		// p.Phone = phone
		p.Bank = bankname
		p.BankAccounts = accounts
		p.IFSC = ifsc
		p.RealName = realname
		p.AD_ADID = afid
		p.AD_Key = afKey
		p.AD_OsVersion = os
		p.AD_BundleId = bundleId
		p.AD_Tracker_Channel = mediasource
		p.AD_AppId = appid
		p.AD_RefGameId = refgameid
		p.AD_RefPkgName = refpkgname
		p.Phone2 = phone2

		err := service.PlayerService.UpdateUser(p)
		this.checkError(err)
		msg := new(entity.ModifyUserData)
		msg.UserId = userid
		msg.Photo = p.Photo
		msg.NickName = nickname
		msg.Phone = phone2 //支付时手机号
		msg.BankName = bankname
		msg.BankNumber = accounts
		msg.Ifsc = ifsc
		msg.BundleId = bundleId
		msg.OsVersion = os
		msg.AfId = afid
		msg.MediaSource = mediasource
		msg.AfKey = afKey
		msg.AppId = appid
		msg.RefGameId = refgameid
		msg.RefPkgName = refpkgname
		result, err := service.GmRequest(pb.WebModifyUser, pb.CONFIG_UPSERT, msg)
		beego.Trace("result: ", result)
		if err != nil {
			this.checkError(err)
		}

		service.ActionService.UpdateUser(this.auth.GetUserName(), p.Userid)
		this.redirect(beego.URLFor("PlayerController.List"))
	} else {
		if p.AD_ADID != "" {
			// 加载时获取安卓评分
			msg := make([]string, 0)
			// msg["ad_id"] = p.AD_ADID
			msg = append(msg, p.AD_ADID)
			result, err := service.AdGetRequest(msg)
			beego.Trace("result: ", result)
			if err == nil {
				if result != nil {
					if result.Data != nil {
						for _, item := range result.Data {
							if item.Status == 0 {
								p.AndroidScore = item.Score
							}
						}
					} else {
						p.AndroidScore = -1
					}
				}
			}
		} else {
			p.AndroidScore = -1
		}
	}

	this.Data["pageTitle"] = "资料修改"
	this.Data["user"] = p
	this.display()
}

// 充值彩金
func (this *PlayerController) RechargeDiamond() {
	userid := this.GetString("id")
	p, err := service.PlayerService.GetPlayer(userid)
	this.checkError(err)

	if this.isPost() {
		remark := this.GetString("remark")
		diamond := this.GetString("diamond")
		if diamond == "" {
			this.checkError(errors.New("请输入充值金币!"))
		}
		if remark == "" {
			this.checkError(errors.New("请输入备注!"))
		}
		// err := this.validGift(1, diamond)
		// this.checkError(err)
		diamond1, err1 := this.GetInt("diamond")
		userid := p.Userid
		if userid == "" {
			this.checkError(errors.New("未查询到该用户信息!"))
		}
		if err1 != nil {
			this.checkError(err1)
		}

		msg := new(pb.PayCurrency)
		msg.Userid = userid
		msg.Type = entity.LogType65
		msg.Coin = 0
		if diamond1 != 0 {
			// 通知服务器
			msg.Diamond = int64(diamond1)
			result, err := service.GmRequest(pb.WebGive, pb.CONFIG_UPSERT, msg)
			beego.Trace("result: ", result)
			if err != nil {
				this.checkError(err)
			} else {
				loginUser := this.auth.GetUser()
				gift := new(entity.GoldGiftLog)
				gift.Userid = msg.Userid
				gift.Nickname = p.Nickname
				gift.BundleId = p.AD_BundleId
				gift.GiftType = 1 // 充值金币
				gift.Score = int64(diamond1)
				gift.Remark = remark
				gift.CurScore = int64(p.Diamond)
				gift.ChangeScore = int64(p.Diamond) + int64(diamond1)
				gift.Operator = loginUser.UserName
				gift.RemarkPerson = loginUser.UserName

				// 写入后台金币赠送日志中
				err = service.PlayerService.AddGoldGiftLog(gift)
				this.checkError(err)
				if err == nil {
					if diamond1 != 0 {
						service.ActionService.UpdateDiamond(this.auth.GetUser().UserName,
							utils.String(entity.LogType63), userid, utils.String(diamond1), remark)
					}
				}
			}
		} else {
			this.checkError(errors.New("请输入正确的彩金"))
		}
		this.redirect(beego.URLFor("PlayerController.List"))
	} else {
		this.Data["player"] = p
		this.Data["pageTitle"] = "充值彩金"
		this.display()
	}
}

// 修改彩金
func (this *PlayerController) EditDiamond() {
	userid := this.GetString("id")
	p, err := service.PlayerService.GetPlayer(userid)
	this.checkError(err)

	if this.isPost() {
		remark := this.GetString("remark")
		diamond := this.GetString("diamond")
		if diamond == "" {
			this.checkError(errors.New("请输入修改金币!"))
		}
		if remark == "" {
			this.checkError(errors.New("请输入备注!"))
		}
		diamond1, err1 := this.GetInt("diamond")
		userid := p.Userid
		if userid == "" {
			this.checkError(errors.New("未查询到该用户信息!"))
		}
		if err1 != nil {
			this.checkError(err1)
		}

		msg := new(pb.ModifyCurrency)
		msg.Userid = userid
		msg.Type = entity.LogType64
		msg.Coin = -1
		msg.Give = -1
		if diamond1 >= 0 {
			// 通知服务器
			msg.Diamond = int64(diamond1)
			result, err := service.GmRequest(pb.WebModifyNum, pb.CONFIG_UPSERT, msg)
			beego.Trace("result: ", result)
			if err != nil {
				this.checkError(err)
			} else {
				loginUser := this.auth.GetUser()
				gift := new(entity.GoldGiftLog)
				gift.Userid = msg.Userid
				gift.Nickname = p.Nickname
				gift.BundleId = p.AD_BundleId
				gift.GiftType = 2
				gift.Score = int64(diamond1)
				gift.Remark = remark
				gift.CurScore = int64(p.Diamond)
				gift.ChangeScore = int64(diamond1)
				gift.Operator = loginUser.UserName
				gift.RemarkPerson = loginUser.UserName

				// 写入后台金币赠送日志中
				err = service.PlayerService.AddGoldGiftLog(gift)
				this.checkError(err)
				if err == nil {
					if diamond1 != 0 {
						service.ActionService.UpdateDiamond(this.auth.GetUser().UserName,
							utils.String(entity.LogType64), userid, utils.String(diamond1), remark)
					}
				}
			}
		} else {
			this.checkError(errors.New("请输入正确的彩金"))
		}
		this.redirect(beego.URLFor("PlayerController.List"))
	} else {
		this.Data["player"] = p
		this.Data["pageTitle"] = "修改彩金"
		this.display()
	}
}

// 充值奖励金
func (this *PlayerController) RechargeCoin() {
	userid := this.GetString("id")
	p, err := service.PlayerService.GetPlayer(userid)
	this.checkError(err)

	if this.isPost() {
		remark := this.GetString("remark")
		coin := this.GetString("coin")
		if coin == "" {
			this.checkError(errors.New("请输入充值奖励金!"))
		}
		if remark == "" {
			this.checkError(errors.New("请输入备注!"))
		}
		coin1, err1 := this.GetInt("coin")
		userid := p.Userid
		if userid == "" {
			this.checkError(errors.New("未查询到该用户信息!"))
		}
		if err1 != nil {
			this.checkError(err1)
		}
		msg := new(pb.PayCurrency)
		msg.Userid = userid
		msg.Give = int64(coin1)
		msg.Diamond = 0
		msg.Type = entity.LogType65
		//  msg.Remark = remark
		if coin1 != 0 {
			result, err := service.GmRequest(pb.WebGive, pb.CONFIG_UPSERT, msg)
			beego.Trace("result: ", result)
			if err != nil {
				this.checkError(err)
			} else {
				// 写入后台金币赠送日志中
				loginUser := this.auth.GetUser()
				gift := new(entity.GoldGiftLog)
				gift.Userid = msg.Userid
				gift.Nickname = p.Nickname
				gift.BundleId = p.AD_BundleId
				gift.GiftType = 3
				gift.Score = int64(coin1)
				gift.Remark = remark
				// gift.CurScore = int64(p.Coin)
				// gift.ChangeScore = int64(p.Coin) + int64(coin1)
				gift.Operator = loginUser.UserName
				gift.RemarkPerson = loginUser.UserName
				err := service.PlayerService.AddGoldGiftLog(gift)
				this.checkError(err)
				if err == nil {
					if coin1 != 0 {
						service.ActionService.UpdateDiamond(this.auth.GetUser().UserName,
							utils.String(entity.LogType65), userid, utils.String(coin), remark)
					}
				}
			}
		} else {
			this.checkError(errors.New("请输入正确的奖励金"))
		}
		this.redirect(beego.URLFor("PlayerController.List"))
	} else {
		this.Data["player"] = p
		this.Data["pageTitle"] = "充值奖励金"
		this.display()
	}
}

// 修改奖励金
func (this *PlayerController) EditCoin() {
	userid := this.GetString("id")
	p, err := service.PlayerService.GetPlayer(userid)
	this.checkError(err)

	if this.isPost() {
		remark := this.GetString("remark")
		coin := this.GetString("coin")
		if coin == "" {
			this.checkError(errors.New("请输入修改金币!"))
		}
		if remark == "" {
			this.checkError(errors.New("请输入备注!"))
		}
		coin1, err1 := this.GetInt("coin")
		userid := p.Userid
		if userid == "" {
			this.checkError(errors.New("未查询到该用户信息!"))
		}
		if err1 != nil {
			this.checkError(err1)
		}
		msg := new(pb.ModifyCurrency)
		msg.Userid = userid
		msg.Give = int64(coin1)
		msg.Diamond = -1
		msg.Coin = -1
		msg.Type = entity.LogType64
		//  msg.Remark = remark
		if coin1 >= 0 {
			result, err := service.GmRequest(pb.WebModifyNum, pb.CONFIG_UPSERT, msg)
			beego.Trace("result: ", result)
			if err != nil {
				this.checkError(err)
			} else {
				// 写入后台金币赠送日志中
				loginUser := this.auth.GetUser()
				gift := new(entity.GoldGiftLog)
				gift.Userid = msg.Userid
				gift.Nickname = p.Nickname
				gift.BundleId = p.AD_BundleId
				gift.GiftType = 4
				gift.Score = int64(coin1)
				gift.Remark = remark
				// gift.CurScore = int64(p.Coin)
				// gift.ChangeScore = int64(coin1)
				gift.Operator = loginUser.UserName
				gift.RemarkPerson = loginUser.UserName
				err := service.PlayerService.AddGoldGiftLog(gift)
				this.checkError(err)
				if err == nil {
					if coin1 != 0 {
						service.ActionService.UpdateDiamond(this.auth.GetUser().UserName,
							utils.String(entity.LogType66), userid, utils.String(coin), "")
					}
				}
			}
		} else {
			this.checkError(errors.New("请输入正确的奖励金"))
		}
		this.redirect(beego.URLFor("PlayerController.List"))
	} else {
		this.Data["player"] = p
		this.Data["pageTitle"] = "修改奖励金"
		this.display()
	}
}

// 修改可提现金额
func (this *PlayerController) EditCashOut() {
	userid := this.GetString("id")
	p, err := service.PlayerService.GetPlayer(userid)
	this.checkError(err)
	if this.isPost() {
		remark := this.GetString("remark")
		coin, err1 := this.GetInt("coin")

		userid := p.Userid
		if userid == "" {
			this.checkError(fmt.Errorf(userid))
		}
		if err1 != nil {
			this.checkError(err1)
		}
		msg := new(entity.GiveWithdrawCash)
		msg.Userid = userid
		msg.Cash = int64(coin)
		result, err := service.GmRequest(pb.WebGiveWithdraw, pb.CONFIG_UPSERT, msg)
		beego.Trace("result: ", result)
		if err != nil {
			this.checkError(err)
		} else {
			// 写入后台金币赠送日志中
			loginUser := this.auth.GetUser()
			gift := new(entity.GoldGiftLog)
			gift.Userid = msg.Userid
			gift.Nickname = p.Nickname
			gift.BundleId = p.AD_BundleId
			gift.GiftType = 5
			gift.Score = int64(coin)
			gift.Remark = remark
			gift.CurScore = int64(p.CashOut)
			gift.ChangeScore = int64(coin)
			gift.Operator = loginUser.UserName
			gift.RemarkPerson = loginUser.UserName
			err := service.PlayerService.AddGoldGiftLog(gift)
			this.checkError(err)
			if err == nil {
				if coin != 0 {
					service.ActionService.UpdateDiamond(this.auth.GetUser().UserName,
						utils.String(entity.LogType83), userid, utils.String(coin), remark)
				}
			}
		}
		this.redirect(beego.URLFor("PlayerController.List"))
	} else {
		this.Data["player"] = p
		this.Data["pageTitle"] = "增加可提现金额"
		this.display()
	}
}

// 验证
func (this *PlayerController) validGift(Gtype int, coin string) error {
	valid := validation.Validation{}
	switch Gtype {
	case 3:
		valid.Required(coin, "coin").Message("奖励金不能为空")
		valid.Numeric(coin, "coin").Message("请输入正确的奖励金")
		// valid.Range(coin, 1, 1000000, "diamond").Message("输入值不能超过1000000")
	default:
		valid.Required(coin, "diamond").Message("彩金不能为空")
		valid.Numeric(coin, "diamond").Message("请输入正确的彩金")
		// valid.Range(coin, 1, 1000000, "coin").Message("输入值不能超过1000000")
	}

	// 处理符合要求的输入
	fmt.Println("输入正确！")

	if valid.HasErrors() {
		for _, err := range valid.Errors {
			return errors.New(err.Message)
		}
	}

	return nil
}

// 加入黑名单
func (this *PlayerController) BlackListAdd() {
	id := this.GetString("id")

	user, err := service.PlayerService.GetUser(id)
	this.checkError(err)
	user.Status = 3

	err = service.PlayerService.UpdateUserStatus(user)
	this.checkError(err)
	reqMsg := &entity.BlackList{
		Userid: user.Userid,
		Status: user.Status,
	}

	result, err := service.GmRequest(pb.WebBlack, pb.CONFIG_UPSERT, reqMsg)
	beego.Trace("result: ", result)
	if err != nil {
		this.checkError(err)
	} else {
		service.ActionService.UpdateUser(this.auth.GetUserName(), id)
	}
	this.redirect(beego.URLFor("PlayerController.List"))
}

// 移除黑名单
func (this *PlayerController) BlackListDel() {
	id := this.GetString("id")
	page := this.GetString("page")
	stype, _ := this.GetInt("stype")

	user, err := service.PlayerService.GetUser(id)
	this.checkError(err)
	status := user.Status
	user.Status = 1

	err = service.PlayerService.UpdateUserStatus(user)
	this.checkError(err)
	reqMsg := &entity.BlackList{
		Userid: user.Userid,
		Status: status,
	}

	result, err := service.GmRequest(pb.WebBlack, pb.CONFIG_DELETE, reqMsg)
	beego.Trace("result: ", result)
	if err != nil {
		this.checkError(err)
	} else {
		service.ActionService.UpdateUser(this.auth.GetUserName(), id)
	}
	if page != "" {
		if page == "blacklist" {
			this.redirect(beego.URLFor("PlayerController.BlackList", "stype", stype))
		} else {
			this.redirect(beego.URLFor("PlayerController.List"))
		}
	} else {
		this.redirect(beego.URLFor("PlayerController.List"))
	}
}

func (this *PlayerController) ExportUser() {

}

// 游戏局数
func (this *PlayerController) UserGames() {
	// 获取用户 ID
	userId := this.GetString("id")
	if userId == "" {
		this.checkError(errors.New("用户ID不能为空"))
	}
	var buf bytes.Buffer
	info, _ := service.PlayerService.GetDetailTotalByUser(userId)

	buf.WriteString(fmt.Sprintf("<p>用户ID:%s</p>", info.Id))
	buf.WriteString("<div class=\"table-panel\">")
	buf.WriteString("<table class=\"table table-striped table-bordered table-hover\">")
	buf.WriteString("<tr>")
	buf.WriteString("<th>TP</th>")
	buf.WriteString("<th>DRAGON TIGER</th>")
	buf.WriteString("<th>7UPDOWN</th>")
	buf.WriteString("<th>RUMMY</th>")
	buf.WriteString("<th>AK47</th>")
	buf.WriteString("<th>JOKER</th>")
	buf.WriteString("<th>CRASH</th>")
	buf.WriteString("<th>ANDARBAHAR</th>")
	buf.WriteString("<th>彩票</th>")
	buf.WriteString("<th>飞机</th>")
	buf.WriteString("<th>红黑大战</th>")
	buf.WriteString("<th>RUMMY双人</th>")
	buf.WriteString("<th>TP2</th>")
	buf.WriteString("</tr>")
	buf.WriteString("<tr>")
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.TPNumber))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.LHDNumber))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.UPNumber))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.RMNumber))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.AKNumber))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.JOKERNumber))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.CRASHNumber))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.ABNumber))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.CPNumber))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.FJNumber))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.RBNumber))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.RMTwoNumber))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.TP2Number))
	buf.WriteString("</tr>")
	buf.WriteString("</table>")

	buf.WriteString("<div class=\"table-top\">")
	buf.WriteString("<p>Slots</p>")
	buf.WriteString("<table class=\"table table-striped table-bordered table-hover\">")
	buf.WriteString("<tr>")
	buf.WriteString("<th>Ganesha Fortune</th>")
	buf.WriteString("<th>Lucky Neko</th>")
	buf.WriteString("<th>Ganesha Gold</th>")
	buf.WriteString("<th>Fortune Ox</th>")
	buf.WriteString("<th>Speed Winner</th>")
	buf.WriteString("<th>Treasures of Aztec</th>")
	buf.WriteString("<th>Wild Bounty Showdown</th>")
	buf.WriteString("<th>Rise of Apollo</th>")
	buf.WriteString("<th>Fortune Tiger</th>")
	buf.WriteString("<th>Destiny of Sun & Moon</th>")
	buf.WriteString("<th>Legend of Perseus</th>")
	buf.WriteString("<th>CaiShen Wins</th>")
	buf.WriteString("<th>Songkran Splash</th>")
	buf.WriteString("<th>Asgardian Rising</th>")
	buf.WriteString("<th>Jurassic Kingdom</th>")
	buf.WriteString("<th>Wild Bandito</th>")
	buf.WriteString("<th>Supermarket Spree</th>")
	buf.WriteString("<th>Cocktail Nights</th>")
	buf.WriteString("</tr>")
	buf.WriteString("<tr>")
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.G_F_Number))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.L_N_Number))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.G_G_Number))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.F_O_Number))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.S_W_Number))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.T_O_A_Number))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.W_B_S_Number))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.R_O_A_Number))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.F_T_Number))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.D_O_S_M_Number))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.L_O_P_Number))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.C_W_Number))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.S_S_Number))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.A_R_Number))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.J_K_Number))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.W_B_Number))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.S_S_Number))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.C_N_Number))
	buf.WriteString("</tr>")
	buf.WriteString("</table>")

	buf.WriteString("<div class=\"table-panel table-top\">")
	buf.WriteString("<p>真人视讯</p>")
	buf.WriteString("<table class=\"table table-striped table-bordered table-hover\">")
	buf.WriteString("<tr>")
	buf.WriteString("<th>Auto-Roulette</th>")
	buf.WriteString("<th>Lightning Blackjack</th>")
	buf.WriteString("<th>Lightning Roulette</th>")
	buf.WriteString("<th>Super Sic Bo</th>")
	buf.WriteString("<th>Dragon Tiger</th>")
	buf.WriteString("<th>Dream Catcher</th>")
	buf.WriteString("<th>Fan Tan</th>")
	buf.WriteString("<th>Golden Wealth Baccarat</th>")
	buf.WriteString("<th>Bac Bo</th>")
	buf.WriteString("</tr>")
	buf.WriteString("<tr>")
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.A_R_Number))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.L_B_Number))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.L_R_Number))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.S_S_B_Number))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.D_T_Number))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.D_C_Number))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.FAN_T_Number))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.G_W_B_Number))
	buf.WriteString(fmt.Sprintf("<td>%d</td>", info.B_B_Number))
	buf.WriteString("</tr>")
	buf.WriteString("</table>")
	buf.WriteString("</div>")
	// 将查询结果渲染到模板中
	this.Ctx.WriteString(buf.String())
}

func (this *PlayerController) EditUserType() {
	userId := this.GetString("updateuserid")
	usertypeId := this.GetString("listusertype")
	if userId == "" {
		this.checkError(errors.New("用户ID不能为空"))
	}

	if usertypeId == "6" {
		usertypeId = this.GetString("customtype")
	}
	err := service.PlayerService.UpdateUserTypes(userId, usertypeId)
	this.checkError(err)
	this.redirect(beego.URLFor("PlayerController.List"))
}

// 将interface{}转成[]
func ConvertToUserSlice(data interface{}) ([]entity.OnlineUser, bool) {
	if slice, ok := data.([]entity.OnlineUser); ok {
		return slice, true
	}
	return nil, false
}

// 在线玩家 从服务器加载数据
func (this *PlayerController) OnlinePlayer() {
	page, _ := strconv.Atoi(this.GetString("page"))
	userid := this.GetString("userid")
	roomId := this.GetString("roomId")
	packageId := this.GetString("package_id")
	if page < 1 {
		page = 1
	}
	// 获取在线用户列表
	userlist := service.PlayerService.ListOnlineUsers()

	// 定义初始筛选结果切片
	var filteredUsers []entity.OnlineUser

	// 遍历切片并根据条件筛选
	if userid != "" || roomId != "" || packageId != "" {
		temp_arr := make([]string, 0)
		if packageId != "" && packageId != "0" && packageId != "-" {
			if strings.Contains(packageId, ",") {
				// 多个
				palkage_arr := strings.Split(packageId, ",")
				for _, item := range palkage_arr {
					if item != "" && item != "0" && item != "-" {
						temp_arr = append(temp_arr, item)
					}
				}
			} else {
				// 单个
				temp_arr = append(temp_arr, packageId)
			}
		}
		for _, user := range userlist {
			if userid != "" {
				if user.Id != userid {
					continue
				}
			}
			if roomId != "" && roomId != "-1" {
				if user.GameId != roomId {
					continue
				}
			}
			if len(temp_arr) > 0 {
				isFlag := false
				for _, v := range temp_arr {
					if user.Channel == v {
						isFlag = true
						break
					}
				}
				if !isFlag {
					continue
				}
			}

			// 将符合条件的用户添加到筛选结果中
			filteredUsers = append(filteredUsers, user)
		}
	} else {
		filteredUsers = userlist
	}
	// 使用 sort.Slice() 对用户信息进行排序
	sort.Slice(filteredUsers, func(i, j int) bool {
		return filteredUsers[i].Asset > filteredUsers[j].Asset
	})
	// list := filteredUsers
	count := len(filteredUsers)
	startIndex := (page - 1) * this.pageSize
	endIndex := (page) * this.pageSize
	if endIndex > count {
		endIndex = count
	}
	list := filteredUsers[startIndex:endIndex]

	roomList := map[string]string{
		"-1": "全部",
		"0":  "大厅",
	}
	for k, v := range service.GtypeNameMap {
		roomList[utils.String(k)] = v
	}

	le := len(list)
	for i := 0; i < le; i++ {
		// list[i].Amount = list[i].Amount / 100 //转换为元
		for k, v := range roomList {
			if k == list[i].GameId {
				list[i].GameName = v
			}
		}
	}

	packageList, _ := service.ChannelService.GetChannelAll()

	this.Data["pageTitle"] = "在线玩家"
	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["roomList"] = roomList
	this.Data["roomId"] = roomId
	this.Data["packageList"] = packageList
	this.Data["packageId"] = packageId
	this.Data["userid"] = userid
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("PlayerController.OnlinePlayer", "userid", userid, "roomId", roomId, "packageId", packageId), true).ToString()
	this.display()
}

// 黑名单列表
func (this *PlayerController) BlackList() {
	status, _ := this.GetInt("status")
	page, _ := strconv.Atoi(this.GetString("page"))
	userid := this.GetString("userid")
	stype, _ := this.GetInt("stype")
	packageId := this.GetString("package_id")
	if page < 1 {
		page = 1
	}
	m := bson.M{}
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
	if stype == 2 {
		// IP限制
		if userid != "" {
			m["_id"] = userid
		}
		count, _ = service.PlayerService.GetIPwhiteTotal(m)
		list, _ := service.PlayerService.GetIPwhite(page, this.pageSize, m)

		this.Data["list"] = list
	} else if stype == 3 {
		// 开服IP限制
		if userid != "" {
			m["_id"] = userid
		}
		count, _ = service.PlayerService.GetServerWhiteTotal(m)
		list, _ := service.PlayerService.GetServerWhite(page, this.pageSize, m)
		this.Data["list"] = list
	} else if stype == 4 {
		// 银行卡黑名单
		m = bson.M{}
		count, _ = service.PlayerService.GetCardBlacklistTotal(m)
		list, _ := service.PlayerService.GetCardBlacklist(page, this.pageSize, m)
		this.Data["list"] = list
	} else if stype == 5 {
		// 设备码黑名单
		m = bson.M{}
		count, _ = service.PlayerService.GetEquipmentBlacklistTotal(m)
		list, _ := service.PlayerService.GetEquipmentBlacklist(page, this.pageSize, m)
		this.Data["list"] = list
	} else if stype == 6 {
		// 支付黑名单
		m := bson.M{}
		count, _ = service.PayService.GetPayBlackListTotal(m)
		list, _ := service.PayService.GetPayBlackList(page, this.pageSize, m)
		this.Data["list"] = list
	} else {
		if stype == 0 {
			m["status"] = 3
		} else if stype == 1 {
			m["status"] = 4
		}
		if userid != "" {
			m["_id"] = userid
		}
		// if packageId != "" {
		// 	if packageId != "0" {
		// 		if packageId == "-" {
		// 			m["ad__bundle_id"] = ""
		// 		} else {
		// 			m["ad__bundle_id"] = packageId
		// 		}
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
		count, _ = service.PlayerService.GetTotal(m)
		list, _ := service.PlayerService.GetBlackList(page, this.pageSize, m)

		this.Data["list"] = list
	}

	packageList, _ := service.ChannelService.GetChannelAll()
	this.Data["count"] = count
	this.Data["pageTitle"] = "黑白名单"
	this.Data["packageList"] = packageList
	this.Data["packageId"] = packageId
	this.Data["stype"] = stype
	this.Data["userid"] = userid
	this.Data["status"] = status
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("PlayerController.BlackList", "status", status, "userid", userid, "stype", stype, "packageId", packageId), true).ToString()
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "blacklistopt")
	this.display()
}

// 新增黑白名单
func (this *PlayerController) BlackWhiteAdd() {
	userid := this.GetString("userid")
	typeId, _ := this.GetInt("typeId")
	if userid == "" {
		this.checkError(errors.New("请输入！"))
	}
	if this.isPost() {
		if typeId == 2 || typeId == 3 {
			// IP限制 或者开服IP限制
			// parsedIP := net.ParseIP(userid)
			// if parsedIP == nil {
			// 	this.checkError(errors.New("IP地址无效"))
			// }
			temp := new(entity.IPwhite)
			if typeId == 3 {
				temp, _ = service.PlayerService.GetServerWhiteByIP(userid)
			} else {
				temp, _ = service.PlayerService.GetIPwhiteByIP(userid)
			}
			if temp != nil {
				if temp.IP != "" {
					this.checkError(errors.New("已存在该IP地址，不可重复添加！"))
				}
			}
			info := new(entity.IPwhite)
			info.IP = userid
			info.Operator = this.auth.GetUser().UserName
			ipmap := map[string]string{
				userid: "",
			}

			if typeId == 3 {
				// 开服IP限制
				err := service.PlayerService.AddServerWhite(info)
				this.checkError(err)
				// 通知服务器
				result, err := service.GmRequest(pb.WebServerWhite, pb.CONFIG_UPSERT, ipmap)
				beego.Trace("result: ", result)
				if err != nil {
					this.checkError(err)
				} else {

					service.ActionService.Add("ServerWhite_add", this.auth.GetUser().UserName, "", userid, userid, "")
				}
			} else {
				// IP限制
				err := service.PlayerService.AddIPwhite(info)
				this.checkError(err)
				// 通知服务器
				result, err := service.GmRequest(pb.WebIpWhite, pb.CONFIG_UPSERT, ipmap)
				beego.Trace("result: ", result)
				if err != nil {
					this.checkError(err)
				} else {

					service.ActionService.Add("IPwhite_add", this.auth.GetUser().UserName, "", userid, userid, "")
				}
			}

		} else {

			user, err := service.PlayerService.GetUser(userid)
			if user == nil {
				this.checkError(errors.New("未查询到该用户信息！"))
			}
			if typeId == 1 {
				// 白名单
				user.Status = 4
			} else {
				// 黑名单
				user.Status = 3
			}
			err = service.PlayerService.UpdateUserStatus(user)
			this.checkError(err)
			reqMsg := &entity.BlackList{
				Userid: user.Userid,
				Status: user.Status,
			}

			result, err := service.GmRequest(pb.WebBlack, pb.CONFIG_UPSERT, reqMsg)
			beego.Trace("result: ", result)
			if err != nil {
				this.checkError(err)
			} else {
				service.ActionService.UpdateUser(this.auth.GetUserName(), userid)
			}
		}
	}
	this.redirect(beego.URLFor("PlayerController.BlackList", "stype", typeId))
}

// 移除IP限制
func (this *PlayerController) IpWhiteDel() {
	id := this.GetString("id")

	user, err := service.PlayerService.GetIPwhiteByIP(id)
	this.checkError(err)

	err = service.PlayerService.DelIPwhite(user)
	this.checkError(err)
	ipmap := map[string]string{
		id: "",
	}

	result, err := service.GmRequest(pb.WebIpWhite, pb.CONFIG_DELETE, ipmap)
	beego.Trace("result: ", result)
	if err != nil {
		this.checkError(err)
	} else {

		service.ActionService.Add("IPwhite_del", this.auth.GetUser().UserName, "", id, id, "")
	}
	this.redirect(beego.URLFor("PlayerController.BlackList", "stype", 2))
}

// 移除开服IP限制
func (this *PlayerController) ServerWhiteDel() {
	id := this.GetString("id")

	user, err := service.PlayerService.GetServerWhiteByIP(id)
	this.checkError(err)

	err = service.PlayerService.DelServerWhite(user)
	this.checkError(err)
	ipmap := map[string]string{
		id: "",
	}

	result, err := service.GmRequest(pb.WebServerWhite, pb.CONFIG_DELETE, ipmap)
	beego.Trace("result: ", result)
	if err != nil {
		this.checkError(err)
	} else {

		service.ActionService.Add("ServerWhite_del", this.auth.GetUser().UserName, "", id, id, "")
	}
	this.redirect(beego.URLFor("PlayerController.BlackList", "stype", 3))
}

func (this *PlayerController) CardAdd() {
	id := this.GetString("cardid")
	if id == "" {
		this.checkError(fmt.Errorf("卡号不能为空"))
	}
	info, err := service.PlayerService.GetCard(id)
	if info.Card != "" {
		this.checkError(fmt.Errorf("该卡号已存在，请重新输入"))
	}
	addinfo := new(entity.CardBlacklist)
	addinfo.Card = id
	addinfo.Operator = this.auth.GetUser().UserName
	err = service.PlayerService.AddCardBlacklist(addinfo)
	if err != nil {
		this.checkError(err)
	} else {
		service.ActionService.Add("CardBlacklist_add", this.auth.GetUser().UserName, "", id, id, "")
	}
	this.redirect(beego.URLFor("PlayerController.BlackList", "stype", 4))
}

// 移除银行卡黑名单
func (this *PlayerController) CardDel() {
	id := this.GetString("id")
	if id == "" {
		this.checkError(fmt.Errorf("卡号不能为空"))
	}

	err := service.PlayerService.DelCard(id)
	if err != nil {
		this.checkError(err)
	} else {
		service.ActionService.Add("CardBlacklist_del", this.auth.GetUser().UserName, "", id, id, "")
	}
	this.redirect(beego.URLFor("PlayerController.BlackList", "stype", 4))
}

// 设备码黑名单添加
func (this *PlayerController) EquipmentAdd() {
	id := this.GetString("equipmentid")
	if id == "" {
		this.checkError(fmt.Errorf("设备码不能为空"))
	}
	info, err := service.PlayerService.GetEquipment(id)
	if info.Code != "" {
		this.checkError(fmt.Errorf("该设备码已存在，请重新输入"))
	}
	addinfo := new(entity.EquipmentBlacklist)
	addinfo.Code = id
	addinfo.Operator = this.auth.GetUser().UserName
	err = service.PlayerService.AddEquipmentBlacklist(addinfo)
	if err != nil {
		this.checkError(err)
	} else {
		u_arr := make(map[string]entity.EquipmentBlacklist)
		u_arr[info.Code] = *addinfo
		result, err := service.GmRequest(pb.WebDeviceBlack, pb.CONFIG_UPSERT, u_arr)
		beego.Trace("result: ", result)
		if err != nil {
			this.checkError(err)
		}
		service.ActionService.Add("Equipment_Blacklist_add", this.auth.GetUser().UserName, "", id, id, "")
	}
	this.redirect(beego.URLFor("PlayerController.BlackList", "stype", 5))
}

// 移除设备码黑名单
func (this *PlayerController) EquipmentDel() {
	id := this.GetString("id")
	if id == "" {
		this.checkError(fmt.Errorf("设备码不能为空"))
	}
	info, _ := service.PlayerService.GetEquipment(id)
	if info.Code != "" {
		u_arr := make(map[string]entity.EquipmentBlacklist)
		u_arr[info.Code] = *info
		result, err := service.GmRequest(pb.WebDeviceBlack, pb.CONFIG_DELETE, u_arr)
		beego.Trace("result: ", result)
		if err != nil {
			this.checkError(err)
		}
		err = service.PlayerService.DelEquipment(id)
		if err != nil {
			this.checkError(err)
		} else {
			service.ActionService.Add("Equipment_Blacklist_del", this.auth.GetUser().UserName, "", id, id, "")
		}
	}
	this.redirect(beego.URLFor("PlayerController.BlackList", "stype", 5))
}

// 添加支付黑名单
func (this *PlayerController) PayBlackListAdd() {
	phone := this.GetString("phone")
	if phone == "" {
		this.checkError(errors.New("手机号码不能为空"))
	}
	if this.isPost() {
		// 提交
		if !validateIndianPhoneNumber(phone) {
			this.checkError(errors.New("手机号码格式不正确"))
		}

		temp, _ := service.PayService.GetPayBlackListById(phone)
		if temp != nil {
			if temp.Phone != "" {
				this.checkError(errors.New("已存在该支付手机号码，不可重复添加！"))
			}
		}

		name := this.auth.GetUser().UserName
		info := new(entity.PayBlackList)
		info.Phone = phone
		info.Operator = name
		// 新增
		err := service.PayService.AddPayBlackList(info)
		this.checkError(err)

		// 通知服务器
		// b := make(map[string]string)
		// b[phone] = ""
		// b := make(map[string]entity.PayBlackList, 0)
		// b[phone] = *info
		// result, err := service.GmRequest(pb.WebRechareBlack, pb.CONFIG_UPSERT, b)
		// beego.Trace("result: ", result)
		// if err != nil {
		// 	this.checkError(err)
		// } else {
		service.ActionService.Add("add_pay_blacklist", name,
			"", utils.String(info.Phone), utils.String(info.Phone), "")
		// }
		this.redirect(beego.URLFor("PlayerController.BlackList", "stype", 6))
	}
}

// 删除支付黑名单
func (this *PlayerController) PayBlackListDel() {
	id := this.GetString("id")

	info, err := service.PayService.GetPayBlackListById(id)
	this.checkError(err)

	err = service.PayService.DelPayBlackList(info)
	this.checkError(err)

	// 通知服务器
	// b := make(map[string]string)
	// b[id] = ""
	// b := make(map[string]entity.PayBlackList, 0)
	// b[id] = *info
	// result, err := service.GmRequest(pb.WebRechareBlack, pb.CONFIG_DELETE, b)
	// beego.Trace("result: ", result)
	// if err != nil {
	// 	this.checkError(err)
	// } else {

	service.ActionService.Add("del_pay_blacklist", this.auth.GetUser().UserName, "", id, id, "")
	// }
	this.redirect(beego.URLFor("PlayerController.BlackList", "stype", 6))
}

// 支付黑名单批量上传
func (this *PlayerController) PayBlackListUpload() {
	if this.isPost() {
		// 判断是否上传文件
		file, h, err := this.GetFile("filename")
		fmt.Println("获取上传文件:", h)

		if err != nil {
			this.checkError(errors.New("请选择正确的上传文件"))
			return
		}
		// 延迟关闭文件
		defer file.Close()
		data, err := ioutil.ReadAll(file)
		if err != nil {
			this.checkError(errors.New("读取上传文件内容错误"))
		}

		// 解析文件内容
		byte_reader := bytes.NewReader(data)

		f, err := excelize.OpenReader(byte_reader)
		this.checkError(err)
		// if err != nil {
		// 	log.Fatal(err)
		// }
		ret := make([]entity.PhoneUpload, 0)
		// 解析第一个工作表中的内容
		sheet := "Sheet1"
		err = decxls.UnmarshalExcelize(f, sheet, &ret)
		this.checkError(err)

		// 输出文件内容
		if len(ret) == 0 {
			this.checkError(errors.New("配置表解析错误"))
			return
		}
		arr := make([]string, 0)
		for _, item := range ret {
			phone := item.Phone
			if phone == "" {
				continue
			}
			if !validateIndianPhoneNumber(phone) {
				//fmt.Println("导入支付黑名单：%s 手机号码格式不正确", phone)
				continue
				// this.checkError(errors.New("手机号码格式不正确"))
			}

			temp, _ := service.PayService.GetPayBlackListById(phone)
			if temp != nil {
				//fmt.Println("导入支付黑名单：%s 已存在该支付手机号码，不可重复添加！", phone)
				if temp.Phone != "" {
					continue
				}
				// this.checkError(errors.New("已存在该支付手机号码，不可重复添加！"))
			}
			arr = append(arr, phone)

		}
		name := this.auth.GetUser().UserName
		b := make(map[string]entity.PayBlackList, 0)
		strId := ""
		for _, v := range arr {
			info := new(entity.PayBlackList)
			info.Phone = v
			info.Operator = name
			// 新增
			err := service.PayService.AddPayBlackList(info)
			this.checkError(err)
			if err != nil {
				this.checkError(err)
			}
			b[v] = *info
			strId += v + ","
		}
		// // 通知服务器
		// result, err := service.GmRequest(pb.WebRechareBlack, pb.CONFIG_UPSERT, b)
		// beego.Trace("result: ", result)

		service.ActionService.Add("add_pay_blacklist_upload", this.auth.GetUser().UserName,
			"", utils.String(strId), utils.String(strId), "")
	}
	this.redirect(beego.URLFor("PlayerController.BlackList", "stype", 6))
}

// 判断是否是印度电话号码
func validateIndianPhoneNumber(phoneNumber string) bool {
	regex := `^[6789]\d{9}$`
	match, _ := regexp.MatchString(regex, phoneNumber)
	return match
}

// 金币流水
func (this *PlayerController) GoldList() {
	status, _ := this.GetInt("status")
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	userid := this.GetString("userid")
	typeId, _ := this.GetInt("type_id")
	packageId := this.GetString("package_id")
	flowId, _ := this.GetInt("flow_id")
	resultid, _ := this.GetInt("result_id")
	if page < 1 {
		page = 1
	}
	// 默认当天数据
	if startDate == "" && endDate == "" {
		today := bson.Now()
		startDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
		endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
	}
	m := service.FindByDate(startDate, endDate, "ctime", "ctime")
	if status == 0 {
		my_data := this.GetSession("my_select_pakeageid")
		if my_data != nil && packageId == "" {
			if packid, ok := my_data.(string); ok {
				packageId = packid
			}
		}
		status = 1
	}
	if userid != "" {
		if typeId == 0 {
			m["userid"] = userid
		} else {
			m["water_id"] = userid
		}
	}
	if flowId != 0 {
		switch flowId {
		case 1:
			m["ltype"] = bson.M{"$in": []int{110, 111}}
		case 2:
			m["ltype"] = bson.M{"$in": []int{75, 67, 68, 69}}
		case 3:
			m["ltype"] = bson.M{"$in": []int{76, 70, 71, 72}}
		case 4:
			m["ltype"] = bson.M{"$in": []int{116, 117}}
		case 5:
			m["ltype"] = bson.M{"$in": []int{114, 115}}
		case 6:
			m["ltype"] = bson.M{"$in": []int{112, 113}}
		case 7:
			m["ltype"] = bson.M{"$in": []int{79, 73}}
		case 8:
			m["ltype"] = bson.M{"$in": []int{87, 88, 89}}
		case 9:
			m["ltype"] = bson.M{"$in": []int{90, 91, 92}}
		case 10:
			m["ltype"] = bson.M{"$in": []int{100, 101}}
		case 11:
			m["ltype"] = bson.M{"$in": []int{106, 107, 108, 109}}
		case 12:
			m["ltype"] = bson.M{"$in": []int{127, 128}}
		case 13:
			m["ltype"] = bson.M{"$in": []int{130, 131}}
		case 99:
			m["ltype"] = bson.M{"$in": []int{93, 94, 95, 105}}
		case 100:
			m["ltype"] = bson.M{"$in": []int{4, 74, 84, 86, 119, 80, 57, 58, 59, 129}}
		case 101:
			m["ltype"] = bson.M{"$in": []int{13, 63}}
		case 102:
			m["ltype"] = bson.M{"$in": []int{61}}
		case 103:
			m["ltype"] = bson.M{"$in": []int{46, 47, 62, 66, 77, 78, 81, 82, 85, 96, 97, 98, 99, 104}}
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
	if resultid != 0 {
		// change_asset
		if resultid == 1 {
			// 赢
			m["change_asset"] = bson.M{"$gt": 0}
		} else if resultid == 2 {
			// 输
			m["change_asset"] = bson.M{"$lt": 0}
		}
	}
	count, _ := service.PlayerService.GetGoldListTotal(m)
	list, _ := service.PlayerService.GetGoldList(page, this.pageSize, m)

	packageList, _ := service.ChannelService.GetChannelAll()
	typeList := map[int]string{
		0: "用户ID",
		1: "流水号",
	}
	resultList := map[int]string{
		0: "全部",
		1: "赢",
		2: "输",
	}
	flowtypeList := map[int]string{
		0:  "全部",
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
		14: "MINES",

		99:  "外接",
		100: "充值",
		101: "提现",
		102: "签到",
		103: "任务",
	}
	this.Data["pageTitle"] = "金币流水"
	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["typeList"] = typeList
	this.Data["typeId"] = typeId
	this.Data["flowList"] = flowtypeList
	this.Data["flowId"] = flowId
	this.Data["packageList"] = packageList
	this.Data["packageId"] = packageId
	this.Data["userid"] = userid
	this.Data["status"] = status
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("PlayerController.GoldList", "status", status, "userid", userid, "type_id", typeId, "flow_id", flowId, "packageId", packageId, "start_date", startDate, "end_date", endDate, "result_id", resultid), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["resultid"] = resultid
	this.Data["resultList"] = resultList
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "playlist")
	this.display()
}

// 对局列表
func (this *PlayerController) PlayList0() {
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	id := this.GetString("id")
	typeId, _ := this.GetInt("type_id")
	recordId, _ := this.GetInt("record_id")
	roomId := this.GetString("room_id")
	if page < 1 {
		page = 1
	}
	// 默认当天数据
	if startDate == "" && endDate == "" {
		if id == "" {
			today := bson.Now()
			startDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
			endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
		}
	}
	// m := bson.M{}
	m := service.FindByDate1(startDate, endDate, "begin_time", "begin_time")
	n := service.FindByDate1(startDate, endDate, "ctime", "ctime")
	m["players"] = bson.M{"$ne": ""}
	n["amount"] = bson.M{"$ne": 0}
	if id != "" {
		if typeId == 0 {
			// 玩家ID
			// bson.RegEx{Pattern: id, Options: "i"}
			m["players"] = bson.M{
				"$regex":   fmt.Sprintf("\\b%s\\b", id),
				"$options": "i",
			}
			n["user_id"] = id
		} else if typeId == 1 {
			// 局号
			m["_id"] = id
			n["round_id"] = id
		} else if typeId == 2 {
			//桌子ID
			m["desk_id"] = id
		}
	}
	isGtype := false
	if recordId != 0 {
		m["gtype"] = recordId
		n["game_id"] = recordId
	}
	if roomId != "全部" && roomId != "" {
		isGtype = true
		if recordId == 2 || recordId == 3 || recordId == 7 || recordId == 8 || recordId == 9 || recordId == 10 {
			m["desk_id"] = roomId
		} else {
			m["room_id"] = roomId
		}
	}
	// count, _ := service.PlayerService.GetDetailTotal(m)
	list, info, count, _ := service.PlayerService.GetDetailList(page, this.pageSize, m, n)

	le := len(list)
	for i := 0; i < le; i++ {
		s, err1 := service.ConvertToIndiaTime(list[i].BeginTime)
		if err1 == nil {
			list[i].STime = s
		}
		e, err1 := service.ConvertToIndiaTime(list[i].EndTime)
		if err1 == nil {
			list[i].ETime = e
		}
		if list[i].Gtype == 0 {
			list[i].GtypeName = "全部"
		} else if name, ok := service.GtypeNameMap[int(list[i].Gtype)]; ok {
			list[i].GtypeName = name
		}
	}
	typeList := map[int]string{
		0: "玩家ID",
		1: "局号",
		2: "桌子ID",
	}
	roomlist := make([]entity.Game, 0)
	if isGtype {
		m := bson.M{}
		m["gtype"] = recordId
		m["status"] = 1
		roomlist, _ = service.GameService.GetRoomDropList(m)
	}
	allGame := []entity.Game{
		{Id: "全部", Name: "全部"},
	}
	// 游戏列表
	recordList := map[int]string{0: "全部"}
	utils.CopyMap(recordList, service.GtypeNameMap)

	roomlist = append(allGame, roomlist...)
	this.Data["pageTitle"] = "对局列表"
	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["typeList"] = typeList
	this.Data["typeId"] = typeId
	this.Data["recordList"] = recordList
	this.Data["recordId"] = recordId
	this.Data["roomList"] = roomlist
	this.Data["roomId"] = roomId
	this.Data["id"] = id
	this.Data["info"] = info
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("PlayerController.PlayList", "id", id, "type_id", typeId, "record_id", recordId, "room_id", roomId, "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "gamedetailsopt")
	this.display()
}

// 对局列表（新）
func (this *PlayerController) PlayList() {
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	id := this.GetString("id")
	typeId, _ := this.GetInt("type_id")
	recordId, _ := this.GetInt("record_id")
	roomId := this.GetString("room_id")
	usertypeId, _ := this.GetInt("utype_id")
	if page <= 0 {
		page = 1
	}

	var params = make(map[string]any)

	// 默认当天数据
	if startDate == "" && endDate == "" {
		if id == "" {
			today := bson.Now()
			startDate = today.Format("2006-01-02")
			endDate = today.Format("2006-01-02")
		}
	}
	if startDate != "" {
		s := fmt.Sprintf("%s 00:00:00", startDate)
		s1 := utils.Str2Time(s, service.Location())
		startTime := utils.Time2Stamp(s1)
		params["startTime"] = startTime
	}
	if endDate != "" {
		e := fmt.Sprintf("%s 23:59:59", endDate)
		s2 := utils.Str2Time(e, service.Location())
		endTime := utils.Time2Stamp(s2)
		params["endTime"] = endTime
	}

	if id != "" {
		if typeId == 0 {
			// 玩家ID
			params["user_id"] = id
		} else if typeId == 1 {
			// 局号
			params["round_id"] = id
		} else if typeId == 2 {
			//桌子ID
			params["desk_id"] = id
		}
	}
	isGtype := false
	if recordId != 0 {
		params["gtype"] = recordId
	}
	if roomId != "全部" && roomId != "" {
		isGtype = true
		if recordId == 2 || recordId == 3 || recordId == 7 || recordId == 8 || recordId == 9 || recordId == 10 {
			params["desk_id"] = roomId
		} else {
			params["room_id"] = roomId
		}
	}
	// n := bson.M{}
	if usertypeId != 0 {
		n := service.FindByDate(startDate, "", "logout_time", "logout_time")
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
				n["money"] = bson.M{"$gte": startMoney}
				n["state"] = 2
				n["custom_types"] = bson.M{"$eq": ""}
				n["custom_types"] = bson.M{"$eq": nil}
			} else {
				n["money"] = bson.M{"$gte": startMoney, "$lte": endMoney}
				n["state"] = 2
				n["custom_types"] = bson.M{"$eq": ""}
				n["custom_types"] = bson.M{"$eq": nil}
			}
		} else {
			if usertypeId == 11 {
				n["custom_types"] = bson.M{"$ne": ""}
				n["custom_types"] = bson.M{"$ne": nil}
			} else {
				n["state"] = usertypeId
				n["custom_types"] = bson.M{"$eq": ""}
				n["custom_types"] = bson.M{"$eq": nil}
			}
		}
		if id != "" && typeId == 0 {
			// 玩家ID
			n["_id"] = id
		}
		user_arr, _ := service.PlayerService.GetByUserIdArray(n)
		if len(user_arr) > 0 {
			params["user_id"] = user_arr
		}
	}
	list, info, count, _ := service.PlayerService.GetDetailListCk(page, this.pageSize, params)

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

	for _, dtl := range list {
		// dtl.STime = TimestampToTime(dtl.BeginTime)
		// dtl.ETime = TimestampToTime(dtl.EndTime)
		s, err1 := service.ConvertToIndiaTime(dtl.BeginTime)
		if err1 == nil {
			dtl.STime = s
		}
		e, err1 := service.ConvertToIndiaTime(dtl.EndTime)
		if err1 == nil {
			dtl.ETime = e
		}

		if dtl.Gtype == 0 {
			dtl.GtypeName = "全部"
		} else if name, ok := service.GtypeNameMap[int(dtl.Gtype)]; ok {
			dtl.GtypeName = name
		}

		dtl.FPlayerIds = dtl.UserId
		dtl.FBet = fmt.Sprintf("%.2f", service.Chip2Float(dtl.BetAmount))
		if !service.IsExternalGame(dtl.Gtype) {
			dtl.FBeforeScore = fmt.Sprintf("%.2f", service.Chip2Float(dtl.BeforeScore))
			dtl.FAfterScore = fmt.Sprintf("%.2f", service.Chip2Float(dtl.AfterScore))
		}
		if dtl.WinType == 1 {
			dtl.IsScore = true
		}

		switch dtl.Gtype {
		case 1, 13:
			var score int64
			if dtl.SettleScore > 0 {
				score = dtl.SettleScore - dtl.BetAmount
				dtl.FWin = fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore-dtl.BetAmount+dtl.CashMingTax))
			} else {
				score = dtl.SettleScore
				dtl.FWin = fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore))
			}
			dtl.FScore = fmt.Sprintf("%.2f", service.Chip2Float(score))
			// if dtl.IsCharge {
			// 	dtl.FControl += " 局内充值"
			// }

			var controls []string
			if dtl.TpPrxdActive {
				controls = append(controls, "怦然心动")
			}
			if dtl.TpLjsbActive {
				var cs []string
				switch dtl.TpLjsbJLType {
				case 1:
					cs = append(cs, "超率极乐")
				case 2:
					cs = append(cs, "超利极乐")
				case 3:
					cs = append(cs, "利率极乐")
				}
				if dtl.TpLjsbPy {
					cs = append(cs, "被冤局")
				}
				if dtl.TpLjsbPs {
					cs = append(cs, "冤杀局")
				}
				if dtl.TpLjsbPd {
					cs = append(cs, "冤大局")
				}
				if dtl.TpLjsbPt {
					cs = append(cs, "冤逃局")
				}
				controls = append(controls, fmt.Sprintf("乐极生悲(%s)", strings.Join(cs, ",")))
			}
			if dtl.TpGcyxActive {
				c := "高潮涌现"
				if dtl.TpGcyxRp {
					c += "(压制)"
				}
				if dtl.TpGcyxBp {
					c += "(恩赐)"
				}
				controls = append(controls, c)
			}
			dtl.FControl = strings.Join(controls, " ")
			// dtl.FControl = fmt.Sprintf("TP_%d", dtl.PlayerStageId)
			// // 0新手，1免费，2正常，3剧情，4控制策略
			// switch dtl.TPModel {
			// case 1:
			// 	dtl.FControl += " 新手"
			// case 2:
			// 	dtl.FControl += " 免费"
			// case 3:
			// 	dtl.FControl += " 正常"
			// case 4:
			// 	if dtl.IsTPModel3StoryPlus {
			// 		dtl.FControl += " 剧情P"
			// 	} else {
			// 		dtl.FControl += " 剧情"
			// 	}
			// case 5:
			// 	dtl.FControl += " 控制策略"
			// }
			if dtl.IsCharge {
				dtl.FControl += " 局内充值"
			}
		case 2:
			dtl.FWin = fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore))
			dtl.FScore = fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore-dtl.CashMingTax))
			switch dtl.LhdStrategyId {
			case 1:
				dtl.FControl += "新想事成 "
			case 2:
				dtl.FControl += "求死不能"
			case 3:
				dtl.FControl += "龙狂有祸"
			case 4:
				dtl.FControl += "高潮涌现"
			default:
				dtl.FControl += "--"
			}
		case 3:
			dtl.FWin += fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore))
			dtl.FScore += fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore-dtl.CashMingTax))
			switch dtl.UpStrategyId {
			case 1:
				dtl.FControl += "新想事成 "
			case 2:
				dtl.FControl += "求死不能"
			case 3:
				dtl.FControl += "龙狂有祸"
			case 4:
				dtl.FControl += "高潮涌现"
			default:
				dtl.FControl += "--"
			}
		case 4:
			if dtl.SettleScore > 0 {
				dtl.FWin += fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore+dtl.CashMingTax))
				dtl.FScore += fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore)) //+dtl.CashAnTax
			} else {
				dtl.FWin += fmt.Sprintf("%.2f", service.Chip2Float(0))
				dtl.FScore += fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore))
			}
		case 5:
			var score int64
			if dtl.SettleScore > 0 {
				score = dtl.SettleScore - dtl.BetAmount
				dtl.FWin = fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore-dtl.BetAmount+dtl.CashMingTax))
			} else {
				score = dtl.SettleScore
				dtl.FWin = fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore))
			}
			dtl.FScore = fmt.Sprintf("%.2f", service.Chip2Float(score))
		case 6:
			var score int64
			if dtl.SettleScore > 0 {
				score = dtl.SettleScore - dtl.BetAmount
				dtl.FWin = fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore-dtl.BetAmount+dtl.CashMingTax))
			} else {
				score = dtl.SettleScore
				dtl.FWin = fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore))
			}
			dtl.FScore = fmt.Sprintf("%.2f", service.Chip2Float(score))
		case 7, 10:
			dtl.FWin += fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore+dtl.CashMingTax)) // +e.CashAnTax
			dtl.FScore += fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore))
			for _, t := range dtl.CrashStrategyType {
				switch t {
				case 1:
					dtl.FControl += "扶摇直上 "
				case 2:
					dtl.FControl += "欲薅无门 "
				case 3:
					dtl.FControl += "虚假情报 "
				case 4:
					dtl.FControl += "起死回生 "
				case 5:
					dtl.FControl += "奖池风控 "
				case 6:
					dtl.FControl += "冒险奖励 "
				case 7:
					dtl.FControl += "人狂有祸 "
				case 8:
					dtl.FControl += "高潮涌现 "
				default:
					dtl.FControl += "常规 "
				}
			}
		case 8:
			dtl.FWin += fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore)) // +e.CashAnTax
			dtl.FScore += fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore-dtl.CashMingTax))
			switch dtl.AbStrategyId {
			case 1:
				dtl.FControl += "安能求死 "
			case 2:
				dtl.FControl += "安然躺赢"
			case 3:
				dtl.FControl += "爱拼才赢"
			default:
				dtl.FControl += "--"
			}
		case 9:
			dtl.FWin += fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore))                   // +e.CashAnTax
			dtl.FScore += fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore-dtl.CashMingTax)) // fmt.Sprintf("%s%.2f", pre, service.Chip2Float(e.Win))
			switch dtl.ChangeCardType {
			case 1:
				dtl.FControl = "开局换牌"
			case 2:
				dtl.FControl = "局中换牌"
				// default:
				// 	dtl.FControl = "没有"
			}
			switch dtl.ControlType {
			case 1:
				dtl.FControl += " 当前赢分"
			case 2:
				dtl.FControl += " 房间系数"
				// default:
				// 	dtl.FControl += " 无"
			}
			switch dtl.CpStrategyId {
			case 1:
				dtl.FControl += " 来玩就赢"
			case 2:
				dtl.FControl += " 来易去难"
			default:
				dtl.FControl += "--"
			}
			dtl.FControl += fmt.Sprintf(" 玩家最终系数%.2f", service.Chip2Float(int64(dtl.PlayerFactor)))
		case 11:
			switch dtl.RBStrategyId {
			case 1:
				dtl.FControl += "红运当头 "
			case 2:
				dtl.FControl += "绝处逢生"
			default:
				dtl.FControl += "--"
			}
			dtl.FWin += fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore))
			dtl.FScore += fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore-dtl.CashMingTax))
		case 12:
			if dtl.SettleScore > 0 {
				dtl.FWin += fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore+dtl.CashMingTax))
				dtl.FScore += fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore)) //+dtl.CashAnTax
			} else {
				dtl.FWin += fmt.Sprintf("%.2f", service.Chip2Float(0))
				dtl.FScore += fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore))
			}
			if dtl.RmcOk {
				var robotDraw string = "否"
				if dtl.RmcControlRobotDraw {
					robotDraw = "是"
				}

				switch dtl.RmcCtype {
				case 0:
					dtl.FControl = "随机"
					if dtl.RmcControlEffect {
						// 旧库存控制数据
						dtl.FControl = fmt.Sprintf("库存:%d,人机/玩家抽牌数:%d/%d,人机干预:%s",
							dtl.RmcHierarchy,
							dtl.RmcDrawNumRobot,
							dtl.RmcDrawNumPlayer,
							robotDraw,
						)
					}
				case 1:
					dtl.FControl = fmt.Sprintf("库存:%d,人机/玩家抽牌数:%d/%d,人机干预:%s",
						dtl.RmcHierarchy,
						dtl.RmcDrawNumRobot,
						dtl.RmcDrawNumPlayer,
						robotDraw,
					)
				case 2:
					dtl.FControl += "ROI:" + dtl.RmcRoiId
					dtl.FControl = fmt.Sprintf("ROI:%s,人机/玩家抽牌数:%d/%d,人机干预:%s",
						dtl.RmcRoiId,
						dtl.RmcDrawNumRobot,
						dtl.RmcDrawNumPlayer,
						robotDraw,
					)
				}
				if dtl.RmcControlRobotDraw {
					dtl.FControl += fmt.Sprintf(",人机干预回合:%d", dtl.RmcControlRobotDrawRound)
				}
				if dtl.RmcDropId != "" {
					dtl.FControl += fmt.Sprintf(",弃牌:%s", dtl.RmcDropId)
				}
			}
		case 14: // 地雷mines
			var score int64
			if dtl.SettleScore > 0 {
				score = dtl.SettleScore - dtl.BetAmount
				dtl.FWin = fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore-dtl.BetAmount+dtl.CashMingTax))
			} else {
				score = dtl.SettleScore
				dtl.FWin = fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore))
			}
			dtl.FScore = fmt.Sprintf("%.2f", service.Chip2Float(score))

			dtl.FControl += fmt.Sprintf("埋雷:%d", dtl.MinesMines)
			dtl.FControl += fmt.Sprintf(" 步数:%d", dtl.MinesStep)
			if dtl.WinType == 1 {
				dtl.FControl += fmt.Sprintf(" 返奖倍数:x%.2f", dtl.MinesMultiple)
			}
			if dtl.MinesAutoMines {
				dtl.FControl += " 自动对局"
			}
			if dtl.MinesForce {
				dtl.FControl += " 超时强制结算"
			}
		case 15: //宝石2
			dtl.FWin = fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore+dtl.BetAmount))
			dtl.FScore = fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore))
		case 16: //宝石
			dtl.FWin = fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore+dtl.BetAmount))
			dtl.FScore = fmt.Sprintf("%.2f", service.Chip2Float(dtl.SettleScore))
		}
	}

	// roomList := map[int]string{
	// 	0: "全部",
	// }
	typeList := map[int]string{
		0: "玩家ID",
		1: "局号",
		2: "桌子ID",
	}
	roomlist := make([]entity.Game, 0)
	if isGtype {
		m := bson.M{}
		m["gtype"] = recordId
		m["status"] = 1
		roomlist, _ = service.GameService.GetRoomDropList(m)
	}
	allGame := []entity.Game{
		{Id: "全部", Name: "全部"},
	}
	// 游戏列表
	recordList := map[int]string{0: "全部"}
	utils.CopyMap(recordList, service.GtypeNameMap)

	roomlist = append(allGame, roomlist...)
	this.Data["pageTitle"] = "对局列表"
	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["typeList"] = typeList
	this.Data["typeId"] = typeId
	this.Data["recordList"] = recordList
	this.Data["recordId"] = recordId
	this.Data["roomList"] = roomlist
	this.Data["roomId"] = roomId
	this.Data["id"] = id
	this.Data["info"] = info
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("PlayerController.PlayList", "id", id, "type_id", typeId, "record_id", recordId, "room_id", roomId, "start_date", startDate, "end_date", endDate, "utype_id", usertypeId), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["usertype"] = usertype
	this.Data["usertypeId"] = usertypeId
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "gamedetailsopt")
	this.display()
}

func IsPlayer(userid string) bool {
	return len(userid) < 16
}

// 根据游戏类型筛选房间信息
func (c *PlayerController) GetGameRoom() {
	gtype, _ := c.GetInt("gtype")
	if gtype != 0 && gtype <= 10 {
		m := bson.M{}
		m["gtype"] = gtype
		m["status"] = 1
		roomlist, _ := service.GameService.GetRoomDropList(m)
		allGame := []entity.Game{
			{Id: "全部", Name: "全部"},
		}
		roomlist = append(allGame, roomlist...)
		c.Data["json"] = roomlist
	}
	c.ServeJSON()
}

// TP游戏 对局详情
func (this *PlayerController) TpDetails() {
	id := this.GetString("id")
	if id == "" {
		this.checkError(fmt.Errorf("局号不能为空"))
	}
	uid := this.GetString("userid")
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	typeId, _ := this.GetInt("type_id")
	recordId, _ := this.GetInt("record_id")
	roomId := this.GetString("room_id")
	details, err := service.PlayerService.GetDetail(id)
	var maxLen = 0
	if details != nil {
		for _, v := range details.TPDetail {
			if len(v.TPDetailTurn) > maxLen {
				maxLen = len(v.TPDetailTurn)
			}
		}
	}
	thLen := make(map[int]int, 0)
	for i := 0; i < maxLen; i++ {
		thLen[i] = i + 1
	}

	this.checkError(err)
	this.Data["pageTitle"] = "TP对局详情"
	this.Data["info"] = details
	this.Data["maxlen"] = thLen
	this.Data["typeId"] = typeId
	this.Data["recordId"] = recordId
	this.Data["roomId"] = roomId
	this.Data["id"] = id
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["userid"] = uid
	this.display()
}

// 龙虎斗 对局详情
func (this *PlayerController) DtDetails() {
	id := this.GetString("id")
	if id == "" {
		this.checkError(fmt.Errorf("局号不能为空"))
	}
	uid := this.GetString("userid")
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	typeId, _ := this.GetInt("type_id")
	recordId, _ := this.GetInt("record_id")
	roomId := this.GetString("room_id")
	details, err := service.PlayerService.GetDetail(id)
	this.checkError(err)
	this.Data["pageTitle"] = "DRAGON TIGER对局详情"
	this.Data["info"] = details
	this.Data["typeId"] = typeId
	this.Data["recordId"] = recordId
	this.Data["roomId"] = roomId
	this.Data["id"] = id
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["userid"] = uid
	// this.Data["maxlen"] = thLen
	this.display()
}

// 7UPDOWN 对局详情
func (this *PlayerController) UpDetails() {
	id := this.GetString("id")
	if id == "" {
		this.checkError(fmt.Errorf("局号不能为空"))
	}
	uid := this.GetString("userid")
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	typeId, _ := this.GetInt("type_id")
	recordId, _ := this.GetInt("record_id")
	roomId := this.GetString("room_id")
	details, err := service.PlayerService.GetDetail(id)
	this.checkError(err)
	this.Data["pageTitle"] = "7UPDOWN对局详情"
	this.Data["info"] = details
	// this.Data["maxlen"] = thLen
	this.Data["typeId"] = typeId
	this.Data["recordId"] = recordId
	this.Data["roomId"] = roomId
	this.Data["id"] = id
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["userid"] = uid
	this.display()
}

// CRASH 对局详情
func (this *PlayerController) CrashDetails() {
	id := this.GetString("id")
	if id == "" {
		this.checkError(fmt.Errorf("局号不能为空"))
	}
	uid := this.GetString("userid")
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	typeId, _ := this.GetInt("type_id")
	recordId, _ := this.GetInt("record_id")
	roomId := this.GetString("room_id")
	details, err := service.PlayerService.GetDetail(id)
	this.checkError(err)
	title := "CRASH对局详情"
	if details.Gtype == 10 {
		title = "飞机对局详情"
	}
	this.Data["pageTitle"] = title
	this.Data["info"] = details
	// this.Data["maxlen"] = thLen
	this.Data["typeId"] = typeId
	this.Data["recordId"] = recordId
	this.Data["roomId"] = roomId
	this.Data["id"] = id
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["userid"] = uid
	this.display()
}

// Rummy 对局详情
func (this *PlayerController) RmDetails() {
	id := this.GetString("id")
	if id == "" {
		this.checkError(fmt.Errorf("局号不能为空"))
	}
	uid := this.GetString("userid")
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	typeId, _ := this.GetInt("type_id")
	recordId, _ := this.GetInt("record_id")
	roomId := this.GetString("room_id")
	details, err := service.PlayerService.GetDetail(id)
	this.checkError(err)
	this.Data["pageTitle"] = "RUMMY对局详情"
	this.Data["info"] = details
	// this.Data["maxlen"] = thLen
	this.Data["typeId"] = typeId
	this.Data["recordId"] = recordId
	this.Data["roomId"] = roomId
	this.Data["id"] = id
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["userid"] = uid
	this.display()
}

// Joker游戏 对局详情
func (this *PlayerController) JokerDetails() {
	id := this.GetString("id")
	if id == "" {
		this.checkError(fmt.Errorf("局号不能为空"))
	}
	uid := this.GetString("userid")
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	typeId, _ := this.GetInt("type_id")
	recordId, _ := this.GetInt("record_id")
	roomId := this.GetString("room_id")
	details, err := service.PlayerService.GetDetail(id)
	var maxLen = 0
	if details != nil {
		for _, v := range details.JOKERDetail {
			if len(v.JOKERDetailTurn) > maxLen {
				maxLen = len(v.JOKERDetailTurn)
			}
		}
	}
	thLen := make(map[int]int, 0)
	for i := 0; i < maxLen; i++ {
		thLen[i] = i + 1
	}

	this.checkError(err)
	this.Data["pageTitle"] = "Joker对局详情"
	this.Data["info"] = details
	this.Data["maxlen"] = thLen
	this.Data["typeId"] = typeId
	this.Data["recordId"] = recordId
	this.Data["roomId"] = roomId
	this.Data["id"] = id
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["userid"] = uid
	this.display()
}

// AK47游戏 对局详情
func (this *PlayerController) Ak47Details() {
	id := this.GetString("id")
	if id == "" {
		this.checkError(fmt.Errorf("局号不能为空"))
	}
	uid := this.GetString("userid")
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	typeId, _ := this.GetInt("type_id")
	recordId, _ := this.GetInt("record_id")
	roomId := this.GetString("room_id")
	details, err := service.PlayerService.GetDetail(id)
	var maxLen = 0
	if details != nil {
		for _, v := range details.AK47Detail {
			if len(v.AK47DetailTurn) > maxLen {
				maxLen = len(v.AK47DetailTurn)
			}
		}
	}
	thLen := make(map[int]int, 0)
	for i := 0; i < maxLen; i++ {
		thLen[i] = i + 1
	}

	this.checkError(err)
	this.Data["pageTitle"] = "AK47对局详情"
	this.Data["info"] = details
	this.Data["maxlen"] = thLen
	this.Data["typeId"] = typeId
	this.Data["recordId"] = recordId
	this.Data["roomId"] = roomId
	this.Data["id"] = id
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["userid"] = uid
	this.display()
}

// ANDARBAHAR 对局详情
func (this *PlayerController) ABDetails() {
	id := this.GetString("id")
	if id == "" {
		this.checkError(fmt.Errorf("局号不能为空"))
	}
	uid := this.GetString("userid")
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	typeId, _ := this.GetInt("type_id")
	recordId, _ := this.GetInt("record_id")
	roomId := this.GetString("room_id")
	details, err := service.PlayerService.GetDetail(id)
	this.checkError(err)
	this.Data["pageTitle"] = "ANDARBAHAR(百人版)对局详情"
	this.Data["info"] = details
	this.Data["typeId"] = typeId
	this.Data["recordId"] = recordId
	this.Data["roomId"] = roomId
	this.Data["id"] = id
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["userid"] = uid
	// this.Data["maxlen"] = thLen
	this.display()
}

// 彩票对局详情
func (this *PlayerController) CPDetails() {
	id := this.GetString("id")
	if id == "" {
		this.checkError(fmt.Errorf("局号不能为空"))
	}
	uid := this.GetString("userid")
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	typeId, _ := this.GetInt("type_id")
	recordId, _ := this.GetInt("record_id")
	roomId := this.GetString("room_id")
	details, err := service.PlayerService.GetDetail(id)
	this.checkError(err)
	this.Data["pageTitle"] = "彩票对局详情"
	this.Data["info"] = details
	// this.Data["maxlen"] = thLen
	this.Data["typeId"] = typeId
	this.Data["recordId"] = recordId
	this.Data["roomId"] = roomId
	this.Data["id"] = id
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["userid"] = uid
	this.display()
}

// 红黑大战 对局详情
func (this *PlayerController) RBDetails() {
	id := this.GetString("id")
	if id == "" {
		this.checkError(fmt.Errorf("局号不能为空"))
	}
	uid := this.GetString("userid")
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	typeId, _ := this.GetInt("type_id")
	recordId, _ := this.GetInt("record_id")
	roomId := this.GetString("room_id")
	details, err := service.PlayerService.GetDetail(id)
	this.checkError(err)
	this.Data["pageTitle"] = "红黑大战对局详情"
	this.Data["info"] = details
	this.Data["typeId"] = typeId
	this.Data["recordId"] = recordId
	this.Data["roomId"] = roomId
	this.Data["id"] = id
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["userid"] = uid
	// this.Data["maxlen"] = thLen
	this.display()
}

// MINES 对局详情
func (this *PlayerController) MINESDetails() {
	id := this.GetString("id")
	if id == "" {
		this.checkError(fmt.Errorf("局号不能为空"))
	}
	uid := this.GetString("userid")
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	typeId, _ := this.GetInt("type_id")
	recordId, _ := this.GetInt("record_id")
	roomId := this.GetString("room_id")
	details, err := service.PlayerService.GetDetail(id)
	this.checkError(err)
	this.Data["pageTitle"] = "Mines对局详情"
	this.Data["info"] = details
	this.Data["mines"] = details.MinesDetail
	this.Data["typeId"] = typeId
	this.Data["recordId"] = recordId
	this.Data["roomId"] = roomId
	this.Data["id"] = id
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["userid"] = uid
	this.display()
}

// 外接游戏对局详情
func (this *PlayerController) ExternalDetails() {
	id := this.GetString("id")
	if id == "" {
		this.checkError(fmt.Errorf("局号不能为空"))
	}
	uid := this.GetString("userid")
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	typeId, _ := this.GetInt("type_id")
	recordId, _ := this.GetInt("record_id")
	roomId := this.GetString("room_id")
	details, err := service.PlayerService.GetExternalDetail(id)
	this.checkError(err)
	this.Data["pageTitle"] = "外接游戏详情"
	this.Data["info"] = details
	this.Data["typeId"] = typeId
	this.Data["recordId"] = recordId
	this.Data["roomId"] = roomId
	this.Data["id"] = id
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["userid"] = uid
	// this.Data["maxlen"] = thLen
	this.display()
}

// 对局曲线
func (this *PlayerController) GameCurve() {
	id := this.GetString("id")
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
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
	data10 := make([]interface{}, 0) // 红黑大战
	data11 := make([]interface{}, 0) // Rummy双人
	data12 := make([]interface{}, 0) // TP2
	if id != "" {
		list, _ := service.PlayerService.GetGameCurve(id, startDate, endDate)
		if len(list) > 0 {
			sort.Slice(list, func(i, j int) bool {
				return list[i].BeginTime < list[j].BeginTime
			})
			for _, item := range list {
				c, _ := service.ConvertToIndiaTime(item.BeginTime)
				strTime := c.Format("2006-01-02 15:04:05")
				dates = append(dates, strTime)
				switch item.Gtype {
				case 1:
					// TP
					for _, v := range item.TPDetail {
						if v.UserId == id {
							res := service.Chip2Float(v.Score)
							data = append(data, res)
						}
					}
				case 2:
					// LHD
					for _, v := range item.LHDetail.UserDetail {
						if v.Userid == id {
							res := service.Chip2Float(v.Win)
							data2 = append(data2, res)
						}
					}
				case 3:
					// 7UP
					for _, v := range item.UPDetail.UserDetail {
						if v.Userid == id {
							res := service.Chip2Float(v.Win)
							data3 = append(data3, res)
						}
					}
				case 4:
					// RM
					for _, v := range item.RMDetail {
						if v.UserId == id {
							res := service.Chip2Float(v.Score)
							data1 = append(data1, res)
						}
					}
				case 5:
					// AK47
					for _, v := range item.AK47Detail {
						if v.UserId == id {
							res := service.Chip2Float(v.Score)
							data4 = append(data4, res)
						}
					}
				case 6:
					// Joker
					for _, v := range item.JOKERDetail {
						if v.UserId == id {
							res := service.Chip2Float(v.Score)
							data5 = append(data5, res)
						}
					}
				case 7:
					// Crash
					for _, v := range item.CRASHDetail.UserDetail {
						if v.Userid == id {
							res := service.Chip2Float(v.Win)
							data6 = append(data6, res)
						}
					}
				case 8:
					// AB
					for _, v := range item.ABDetail.UserDetail {
						if v.Userid == id {
							res := service.Chip2Float(v.Win)
							data7 = append(data7, res)
						}
					}
				case 9:
					// CP
					for _, v := range item.CPDetail.UserDetail {
						if v.Userid == id {
							res := service.Chip2Float(v.Win)
							data8 = append(data8, res)
						}
					}
				case 10:
					// 飞机
					for _, v := range item.CRASHDetail.UserDetail {
						if v.Userid == id {
							res := service.Chip2Float(v.Win)
							data9 = append(data9, res)
						}
					}
				case 11:
					for _, v := range item.RBDetail.UserDetail {
						if v.Userid == id {
							res := service.Chip2Float(v.Win)
							data10 = append(data10, res)
						}
					}
				case 12:
					// RM双人
					for _, v := range item.RMDetail {
						if v.UserId == id {
							res := service.Chip2Float(v.Score)
							data11 = append(data11, res)
						}
					}
				case 13:
					for _, v := range item.TPDetail {
						if v.UserId == id {
							res := service.Chip2Float(v.Score)
							data12 = append(data12, res)
						}
					}
				}
			}
		}
	}
	this.Data["pageTitle"] = "对局曲线"
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
	this.Data["RB"] = data10
	this.Data["RummyTwo"] = data11
	this.Data["TpTwo"] = data12
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["id"] = id
	this.display()
}

// VB流水
func (this *PlayerController) VbWater() {
	status, _ := this.GetInt("status")
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	userid := this.GetString("userid")
	typeId, _ := this.GetInt("type_id")
	packageId := this.GetString("package_id")
	flowId, _ := this.GetInt("flow_id")
	if page < 1 {
		page = 1
	}
	// 默认当天数据
	if startDate == "" && endDate == "" {
		today := bson.Now()
		startDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
		endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
	}
	m := service.FindByDate1(startDate, endDate, "ctime", "ctime")
	if status == 0 {
		my_data := this.GetSession("my_select_pakeageid")
		if my_data != nil && packageId == "" {
			if packid, ok := my_data.(string); ok {
				packageId = packid
			}
		}
		status = 1
	}
	if userid != "" {
		if typeId == 0 {
			m["userid"] = userid
		} else {
			// m["water_id"] = userid
		}
	}
	if flowId != 0 {
		switch flowId {
		case 1:
			m["reason"] = bson.M{"$in": []int{103}}
		case 2:
			m["reason"] = bson.M{"$in": []int{123, 124}}
		case 3:
			m["reason"] = bson.M{"$in": []int{97, 98}}
		case 4:
			m["reason"] = bson.M{"$in": []int{121, 122}}
		case 5:
			m["reason"] = bson.M{"$in": []int{4, 74, 84, 86, 119, 80, 57, 58, 59}}
		case 6:
			m["reason"] = bson.M{"$in": []int{125, 126}}
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
	count, _ := service.PlayerService.GetVbWaterTotal(m)
	list, _ := service.PlayerService.GetVbWaterList(page, this.pageSize, m)

	packageList, _ := service.ChannelService.GetChannelAll()
	typeList := map[int]string{
		0: "用户ID",
		// 1: "流水号",
	}
	flowtypeList := map[int]string{
		0: "全部",
		1: "首充转入",
		2: "任务",
		3: "签到",
		4: "利息",
		5: "充值释放",
		6: "输分转化",
	}
	this.Data["pageTitle"] = "VB流水"
	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["typeList"] = typeList
	this.Data["typeId"] = typeId
	this.Data["flowList"] = flowtypeList
	this.Data["flowId"] = flowId
	this.Data["packageList"] = packageList
	this.Data["packageId"] = packageId
	this.Data["userid"] = userid
	this.Data["status"] = status
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("PlayerController.VbWater", "status", status, "userid", userid, "type_id", typeId, "flow_id", flowId, "packageId", packageId, "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "playlist")
	this.display()
}
