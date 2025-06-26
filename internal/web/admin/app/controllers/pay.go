package controllers

import (
	"bytes"
	"errors"
	"fmt"
	"goserver/gen/pb"
	"goserver/internal/web/admin/app/entity"
	"goserver/internal/web/admin/app/libs"
	"goserver/internal/web/admin/app/service"
	"goserver/pkg/data"
	"goserver/pkg/utils"
	"io/ioutil"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/validation"
	"github.com/dhushon/decxls"
	"github.com/globalsign/mgo/bson"
)

type PayController struct {
	BaseController
}

// 充值列表
// Deprecated
func (this *PayController) PayList_deprecated() {
	status, _ := this.GetInt("status")
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	userid := this.GetString("userid")
	type_id, _ := this.GetInt("type_id")
	status_id, _ := this.GetInt("status_id")
	pay_id := this.GetString("pay_id")
	shop_id, _ := this.GetInt("shop_id")
	packageId := this.GetString("package_id")
	isfirst, _ := this.GetInt("isfirst")
	if page < 1 {
		page = 1
	}
	// 默认当月数据
	if startDate == "" && endDate == "" {
		if userid == "" {
			today := bson.Now()
			start := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
			startDate = start.Format("2006-01-02")
			endDate = today.Format("2006-01-02")
		}
	}
	m := service.FindByDate(startDate, endDate, "ctime", "ctime")
	m1 := service.FindByDate(startDate, endDate, "ctime", "ctime")
	if status == 0 {
		my_data := this.GetSession("my_select_pakeageid")
		if my_data != nil && packageId == "" {
			if packid, ok := my_data.(string); ok {
				packageId = packid
			}
		}
		status = 1
	}
	if type_id == 0 {
		if userid != "" {
			m["userid"] = userid
			m1["userid"] = userid
		}
	} else if type_id == 1 {
		if userid != "" {
			m["_id"] = userid
			m1["_id"] = userid
		}
	} else if type_id == 2 {
		if userid != "" {
			m["out_trade_no"] = userid
			m1["out_trade_no"] = userid
		}
	} else if type_id == 3 {
		// 用户渠道
		m["package_id"] = bson.M{"$regex": userid, "$options": "i"}
		m1["package_id"] = bson.M{"$regex": userid, "$options": "i"}
	}
	if status_id != 0 {
		m["order_status"] = status_id
	}
	if pay_id != "0" && pay_id != "" {
		payId, _ := strconv.Atoi(pay_id)
		m["channel_id"] = int(payId)
		m1["channel_id"] = int(payId)
	}
	if shop_id != 0 {
		m["shop_type"] = shop_id
		m1["shop_type"] = shop_id
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
			m["package_id"] = bson.M{"$in": temp_arr}
			m1["package_id"] = bson.M{"$in": temp_arr}
		} else {
			m["package_id"] = packageId
			m1["package_id"] = packageId
		}
		// 存储缓存
		this.SetSession("my_select_pakeageid", packageId)
	} else {
		// 清除缓存
		this.DelSession("my_select_pakeageid")

	}
	count, _ := service.PayService.GetPayTotal(m, isfirst)
	list, _ := service.PayService.PayList_Deprecated(page, this.pageSize, m, isfirst)
	totalinfo, _ := service.PayService.TotalPayOrder(m1)
	typeList := map[int]string{
		0: "用户ID",
		1: "订单号",
		2: "第三方订单号",
		3: "用户渠道",
	}
	statusList := map[int]string{
		0:  "全部",
		1:  "审核中",
		2:  "提现成功",
		3:  "提现退回",
		4:  "提现冻结",
		5:  "未知原因付款失败",
		6:  "派单成功",
		7:  "未知原因派单失败",
		8:  "待派单",
		9:  "派单失败需转单",
		10: "派单失败需驳回",
		11: "付款超时急需转单",
		12: "玩家主动申请退款",
		13: "付款失败需转单",
		14: "付款失败需驳回",
		15: "付款超时需转单",
	}
	firstList := map[int]string{
		0: "全部",
		1: "是",
		2: "否",
	}

	channellist, _ := service.GameService.GetPayChannel()
	info := new(entity.PayChannel)
	info.Id = "0"
	info.Name = "全部"
	newSlice := make([]entity.PayChannel, 0)
	newSlice = append(newSlice, *info)
	payList := append(newSlice, channellist...)

	shopList := map[int]string{
		0: "全部",
		1: "商城直充",
		2: "入门礼包",
		3: "金银铜卡",
		4: "首充",
		5: "局内充值",
	}

	le := len(list)
	for i := 0; i < le; i++ {
		// list[i].Amount = list[i].Amount / 100 //转换为元
		for _, v := range payList {
			id, _ := strconv.Atoi(v.Id)
			if id == int(list[i].ChannelId) {
				list[i].ChannelName = v.Name
			}
			// list[i].ChannelName =
		}
	}
	packageList, _ := service.ChannelService.GetChannelAll()
	this.Data["pageTitle"] = "充值记录"
	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["status"] = status
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("PayController.PayList", "status", status, "package_id", packageId, "userid", userid, "type_id", type_id, "status_id", status_id, "pay_id", pay_id, "shop_id", shop_id, "start_date", startDate, "end_date", endDate, "isfirst", isfirst), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["typeList"] = typeList
	this.Data["typeId"] = type_id
	this.Data["statusList"] = statusList
	this.Data["statusId"] = status_id
	this.Data["payList"] = payList
	this.Data["payId"] = pay_id
	this.Data["shopList"] = shopList
	this.Data["shopId"] = shop_id
	this.Data["userid"] = userid
	this.Data["packageId"] = packageId
	this.Data["packageList"] = packageList
	this.Data["firstList"] = firstList
	this.Data["isfirst"] = isfirst
	this.Data["totalinfo"] = totalinfo
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "paylistopt")
	this.Data["versions"] = fmt.Sprint(service.GetVersions())
	this.display()
}

// 支付手动回调
func (this *PayController) PayCallback() {
	orderid := this.GetString("id")
	if orderid == "" {
		this.checkError(errors.New("订单ID不能为空"))
	}
	result, err := service.GmRequest(pb.WebPayCallback, pb.CONFIG_UPSERT, orderid)
	beego.Trace("result: ", result)
	name := this.auth.GetUser().UserName
	service.ActionService.Add("pay_callback", name,
		"", utils.String(orderid), utils.String(orderid), "")
	if err != nil {
		this.checkError(err)
	} else {
		this.showMsg("操作成功！", MSG_OK, beego.URLFor("PayController.PayList"))
	}
}

// 提现列表
func (this *PayController) WithdrawList0() {
	status, _ := this.GetInt("status")
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	userid := this.GetString("userid")
	type_id, _ := this.GetInt("type_id")
	status_id, _ := this.GetInt("status_id")
	audit_id, _ := this.GetInt("audit_id")
	packageId := this.GetString("package_id")
	if page < 1 {
		page = 1
	}
	// 默认当月数据
	if startDate == "" && endDate == "" {
		if userid == "" {
			today := bson.Now()
			start := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
			startDate = start.Format("2006-01-02")
			endDate = today.Format("2006-01-02")
		}
	}
	payList, _ := service.GameService.GetPayChannel()
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
	if type_id == 0 {
		if userid != "" {
			m["userid"] = userid
		}
	} else if type_id == 1 {
		if userid != "" {
			m["_id"] = userid
		}
	} else if type_id == 2 {
		if userid != "" {
			m["out_trade_no"] = userid
		}
	} else if type_id == 3 {
		if userid != "" {
			m["blank_number"] = userid
		}
	}
	if status_id != 0 {
		m["order_status"] = status_id
	}
	if audit_id != 0 {
		m["examine_way"] = audit_id
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
			m["package_id"] = bson.M{"$in": temp_arr}
		} else {
			m["package_id"] = packageId
		}
		// 存储缓存
		this.SetSession("my_select_pakeageid", packageId)
	} else {
		// 清除缓存
		this.DelSession("my_select_pakeageid")
	}

	count, _ := service.PayService.GetWithdrawTotal(m)
	list, _ := service.PayService.WithdrawList(page, this.pageSize, m, payList)

	typeList := map[int]string{
		0: "用户ID",
		1: "订单号",
		2: "第三方订单号",
		3: "用户银行卡",
	}

	statusList := map[int]string{
		0:  "全部",
		1:  "审核中",
		2:  "提现成功",
		3:  "提现退回",
		4:  "提现冻结",
		5:  "未知原因付款失败",
		6:  "派单成功",
		7:  "未知原因派单失败",
		8:  "待派单",
		9:  "派单失败需转单",
		10: "派单失败需驳回",
		11: "付款超时急需转单",
		12: "玩家主动申请退款",
		13: "付款失败需转单",
		14: "付款失败需驳回",
		15: "付款超时需转单",
	}

	auditList := map[int]string{
		0: "全部",
		1: "自动",
		2: "手动",
	}

	// le := len(list)
	// for i := 0; i < le; i++ {
	// 	for _, v := range payList {
	// 		if v.Id == list[i].OutChannel {
	// 			list[i].OutChannelStr = v.Name
	// 		}
	// 	}
	// }
	packageList, _ := service.ChannelService.GetChannelAll()
	this.Data["pageTitle"] = "提现记录"
	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["status"] = status
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("PayController.WithdrawList", "status", status, "package_id", packageId, "userid", userid, "type_id", type_id, "status_id", status_id, "audit_id", audit_id, "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["userid"] = userid
	this.Data["typeList"] = typeList
	this.Data["typeId"] = type_id
	this.Data["statusList"] = statusList
	this.Data["statusId"] = status_id
	this.Data["auditList"] = auditList
	this.Data["auditId"] = audit_id
	this.Data["packageId"] = packageId
	this.Data["packageList"] = packageList
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "withdrawopt")
	this.Data["isOperationTransferOrder"] = this.auth.HasAccessPerm(this.controllerName, "withdrawopttransferorder")
	this.display()
}

// 提现自动转单/派单设置查询
func (this *PayController) WithdrawSwitchEdit() {
	action := this.GetString("action")

	if action == "EditAutoSwitch" {
		rtype, _ := this.GetInt("rtype") // 6自动转单 7自动派单
		open, _ := this.GetInt("open")
		if !utils.SliceIn(rtype, 6, 7) || !utils.SliceIn(open, 0, 1) {
			this.JsonRFail("开关类型有误")
			return
		}
		var autoSwitch *entity.GameSystem
		tlist, err := service.GameService.GetGameSystemByM(bson.M{"stype": 4, "rtype": rtype})
		if err != nil {
			this.JsonRFail(err.Error())
			return
		}
		if len(tlist) > 0 {
			autoSwitch = &(tlist[0])
		} else {
			autoSwitch = &entity.GameSystem{
				Name:   utils.CaseElse(rtype == 6, "提现自动转单", "提现自动派单"),
				Status: open,
				Stype:  4,
				Rtype:  rtype,
			}
		}
		autoSwitch.Status = open

		err = service.GameService.AddOrUpdateGameSystem(autoSwitch)
		if err != nil {
			this.JsonRFail(err.Error())
			return
		}

		// 通知服务器
		_, err = service.GmRequest(pb.WebSwitch, pb.CONFIG_UPSERT, map[string]entity.GameSystem{
			autoSwitch.Id: *autoSwitch,
		})
		if err != nil {
			this.JsonRFail(err.Error())
			return
		}

		service.ActionService.Add("set_game_icon_config", this.auth.GetUser().UserName,
			"", autoSwitch.Id, autoSwitch.Id, "")
		this.JsonRSuccess(autoSwitch)
		return
	}

}

// 提现列表
func (this *PayController) WithdrawList() {
	status, _ := this.GetInt("status")
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	userid := this.GetString("userid")
	type_id, _ := this.GetInt("type_id")
	status_id, _ := this.GetInt("status_id")
	audit_id, _ := this.GetInt("audit_id")
	// packageId := this.GetString("package_id")
	aliasId := this.GetString("alias_id")
	action := this.GetString("action")

	// 自动转单/派单设置查询
	if action == "AutoSwitch" {
		var autoSwitch = map[string]bool{
			"autoTransfer": false,
			"autoDispatch": false,
			"editable":     this.auth.HasAccessPerm(this.controllerName, "withdrawswitchedit"),
		}
		tlist, err := service.GameService.GetGameSystemByM(bson.M{"stype": 4, "rtype": bson.M{"$in": []int{6, 7}}})
		if err != nil {
			beego.Error(err)
			this.showMsg(err.Error(), MSG_ERR)
		}
		for _, item := range tlist {
			if item.Stype == 4 {
				switch item.Rtype {
				case 6:
					autoSwitch["autoTransfer"] = item.Status == 1
				case 7:
					autoSwitch["autoDispatch"] = item.Status == 1
				}
			}
		}
		this.JsonRSuccess(autoSwitch)
		return
	}
	if page < 1 {
		page = 1
	}
	// 默认七天数据
	if startDate == "" && endDate == "" {
		if userid == "" {
			today := service.NowTime()
			start := today.AddDate(0, 0, -7)
			startDate = start.Format("2006-01-02")
			endDate = today.Format("2006-01-02")
		}
	}
	payList, _ := service.GameService.GetPayChannel()
	m := service.FindByDate(startDate, endDate, "ctime", "ctime")
	if status == 0 {
		// my_data := this.GetSession("my_select_pakeageid")
		// if my_data != nil && packageId == "" {
		// 	if packid, ok := my_data.(string); ok {
		// 		packageId = packid
		// 	}
		// }
		status = 1
	}
	if type_id == 0 {
		if userid != "" {
			m["userid"] = userid
		}
	} else if type_id == 1 {
		if userid != "" {
			m["_id"] = userid
		}
	} else if type_id == 2 {
		if userid != "" {
			m["out_trade_no"] = userid
		}
	} else if type_id == 3 {
		if userid != "" {
			m["blank_number"] = userid
		}
	}
	if status_id != 0 {
		m["order_status"] = status_id
	}
	if audit_id != 0 {
		m["examine_way"] = audit_id
	}
	// if packageId != "" && packageId != "0" && packageId != "-" {
	// 	if strings.Contains(packageId, ",") {
	// 		palkage_arr := strings.Split(packageId, ",")
	// 		temp_arr := make([]string, 0)
	// 		for _, item := range palkage_arr {
	// 			if item != "" && item != "0" && item != "-" {
	// 				temp_arr = append(temp_arr, item)
	// 			}
	// 		}
	// 		m["package_id"] = bson.M{"$in": temp_arr}
	// 	} else {
	// 		m["package_id"] = packageId
	// 	}
	// 	// 存储缓存
	// 	this.SetSession("my_select_pakeageid", packageId)
	// } else {
	// 	// 清除缓存
	// 	this.DelSession("my_select_pakeageid")
	// }

	var packageIds []string
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
		m["package_id"] = bson.M{"$in": packageIds}
	}

	count, err := service.PayService.GetWithdrawTotal(m)
	if err != nil {
		beego.Error(err)
		this.showMsg(err.Error(), MSG_ERR)
	}
	list, err := service.PayService.WithdrawList(page, this.pageSize, m, payList)
	if err != nil {
		beego.Error(err)
		this.showMsg(err.Error(), MSG_ERR)
	}

	// 提现设置
	withdrawSettings := service.PayService.GetWithdrawSettings()
	this.Data["withdrawSettings"] = withdrawSettings

	typeList := map[int]string{
		0: "用户ID",
		1: "订单号",
		2: "第三方订单号",
		3: "用户银行卡",
	}

	statusList := map[int]string{
		0:  "全部",
		1:  "审核中",
		2:  "提现成功",
		3:  "提现退回",
		4:  "提现冻结",
		5:  "未知原因付款失败",
		6:  "派单成功",
		7:  "未知原因派单失败",
		8:  "待派单",
		9:  "派单失败需转单",
		10: "派单失败需驳回",
		11: "付款超时急需转单",
		12: "玩家主动申请退款",
		13: "付款失败需转单",
		14: "付款失败需驳回",
		15: "付款超时需转单",
	}

	auditList := map[int]string{
		0: "全部",
		1: "自动",
		2: "手动",
	}

	// le := len(list)
	// for i := 0; i < le; i++ {
	// 	for _, v := range payList {
	// 		if v.Id == list[i].OutChannel {
	// 			list[i].OutChannelStr = v.Name
	// 		}
	// 	}
	// }

	// packageList, _ := service.ChannelService.GetChannelAll()
	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()

	this.Data["pageTitle"] = "提现记录与审核"
	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["status"] = status
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("PayController.WithdrawList", "status", status, "alias_id", aliasId, "userid", userid, "type_id", type_id, "status_id", status_id, "audit_id", audit_id, "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["userid"] = userid
	this.Data["typeList"] = typeList
	this.Data["typeId"] = type_id
	this.Data["statusList"] = statusList
	this.Data["statusId"] = status_id
	this.Data["auditList"] = auditList
	this.Data["auditId"] = audit_id
	// this.Data["packageId"] = packageId
	// this.Data["packageList"] = packageList
	this.Data["aliasId"] = aliasId
	this.Data["packageAliasList"] = packageAliasList
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "withdrawopt")
	this.Data["isOperationTransferOrder"] = this.auth.HasAccessPerm(this.controllerName, "withdrawopttransferorder")
	this.Data["isOperationEdit"] = this.auth.HasAccessPerm(this.controllerName, "withdrawedit")
	this.Data["isExport"] = this.auth.HasAccessPerm(this.controllerName, "withdrawlistexport")
	this.display()
}

// 提现列表导出
func (this *PayController) WithdrawListExport() {
	status, _ := this.GetInt("status")
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	userid := this.GetString("userid")
	type_id, _ := this.GetInt("type_id")
	status_id, _ := this.GetInt("status_id")
	audit_id, _ := this.GetInt("audit_id")
	// packageId := this.GetString("package_id")
	aliasId := this.GetString("alias_id")
	if page < 1 {
		page = 1
	}
	// 默认七天数据
	if startDate == "" && endDate == "" {
		if userid == "" {
			today := service.NowTime()
			start := today.AddDate(0, 0, -7)
			startDate = start.Format("2006-01-02")
			endDate = today.Format("2006-01-02")
		}
	}
	payList, _ := service.GameService.GetPayChannel()
	m := service.FindByDate(startDate, endDate, "ctime", "ctime")
	if status == 0 {
		// my_data := this.GetSession("my_select_pakeageid")
		// if my_data != nil && packageId == "" {
		// 	if packid, ok := my_data.(string); ok {
		// 		packageId = packid
		// 	}
		// }
		status = 1
	}
	if type_id == 0 {
		if userid != "" {
			m["userid"] = userid
		}
	} else if type_id == 1 {
		if userid != "" {
			m["_id"] = userid
		}
	} else if type_id == 2 {
		if userid != "" {
			m["out_trade_no"] = userid
		}
	} else if type_id == 3 {
		if userid != "" {
			m["blank_number"] = userid
		}
	}
	if status_id != 0 {
		m["order_status"] = status_id
	}
	if audit_id != 0 {
		m["examine_way"] = audit_id
	}

	var packageIds []string
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
		m["package_id"] = bson.M{"$in": packageIds}
	}

	// count, _ := service.PayService.GetWithdrawTotal(m)
	list, err := service.PayService.WithdrawList(1, 100000, m, payList)
	if err != nil {
		this.showMsg(err.Error(), MSG_ERR)
	}

	headers := []string{"用户ID", "昵称", "注册时间", "渠道别名", "打码最多的游戏/返奖率", "玩局数最多的游戏/返奖率", "玩家总返奖率", "玩家当前净盈利", "本单提现金额", "手续费", "实到金额", "总充值金额", "总提单金额", "已提现到账金额", "提单待审核金额", "审核通过未到账金额", "提单失败被自动退回金额", "提现失败被自动退回金额", "提现被冻结金额", "其他提现金额", "玩家提单后携带金额", "玩家当前携带金额", "玩家充值成功率", "玩家提现成功率", "订单状态", "用户银行", "用户银行卡号", "IFSC", "转出渠道", "订单生成时间", "订单号", "第三方订单号", "订单完成时间", "审核方式", "审核人", "操作时间", "成功付款次数", "滞单时长", "到账实效", "审核时长", "转单详情", "金币流水", "评分", "标记信息"}
	var rows [][]any
	for _, v := range list {
		row := []any{
			v.Userid,
			utils.CaseElse(v.Userid == "总汇", "--", v.NickName),
			utils.CaseElse(v.Userid == "总汇", "--", v.Regtime.In(service.Location()).Format(utils.FORMAT)),
			utils.CaseElse(v.Userid == "总汇", "--", v.AliasId),
			utils.CaseElse(v.Userid == "总汇", "--", v.Game1BetStats),
			v.Game1TimesStats,
			v.RebateRate,
			v.FProfit,
			fmt.Sprintf("%.2f", v.FTotal),
			fmt.Sprintf("%.2f", v.FCommission),
			fmt.Sprintf("%.2f", v.FAmount),
			v.RechargeAmount,
			v.WithdrawAmountAll,
			v.WithdrawAmount,
			v.WithdrawWaitAmount,
			v.ProcessingAmount,
			v.BackAmount5,
			v.BackAmount7,
			v.FreezeAmount,
			v.WithdrawOtherAmount,
			utils.CaseElse(v.Userid == "总汇", "--", fmt.Sprintf("%.2f", v.FAfterDiamond)),
			v.FDiamond,
			v.RechargeSuccessRate,
			v.WithdrawSuccessRate,
			utils.CaseElse(v.Userid == "总汇", "--", v.OrderStatusName),
			utils.CaseElse(v.Userid == "总汇", "--", v.Blank),
			utils.CaseElse(v.Userid == "总汇", "--", v.BlankNumber),
			utils.CaseElse(v.Userid == "总汇", "--", v.IFSC),
			utils.CaseElse(v.Userid == "总汇", "--", v.OutChannelStr),
			utils.CaseElse(v.Userid == "总汇", "--", v.Ctime.In(service.Location()).Format(utils.FORMAT)),
			utils.CaseElse(v.Userid == "总汇", "--", v.OrderID),
			utils.CaseElse(v.Userid == "总汇", "--", v.OutTradeNo),
			utils.CaseElse(v.WithdrawTime.IsZero(), "--", v.WithdrawTime.In(service.Location()).Format(utils.FORMAT)),
			utils.CaseElse(v.Userid == "总汇", "--", utils.CaseElse(v.ExamineWay == 1, "自动", "手动")),
			utils.CaseElse(v.Userid == "总汇", "--", v.ExamineAccount),
			utils.CaseElse(v.Userid == "总汇", "--", utils.CaseElse(v.ETime.IsZero(), "-", v.ETime.In(service.Location()).Format(utils.FORMAT))),
			utils.CaseElse(v.Userid == "总汇", "--", fmt.Sprint(v.RepeatTimes)),
			utils.CaseElse(v.Userid == "总汇", "--", fmt.Sprintf("%.1f", v.DelayTime)),
			utils.CaseElse(v.Userid == "总汇", "--", fmt.Sprintf("%.1f", v.TimeOfArrival)),
			utils.CaseElse(v.Userid == "总汇", "--", fmt.Sprintf("%.1f", v.AutioTime)),
			"--",
			"--",
			utils.CaseElse(v.Userid == "总汇", "--", utils.CaseElse(v.AndroidScore == -1, "-", fmt.Sprint(v.AndroidScore))),
			utils.CaseElse(v.Userid == "总汇", "--", strings.Join(v.ErrorMsgs, ";\r\n")),
		}
		rows = append(rows, row)
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
	title := "提现记录与审核"
	fileName := title + "_" + time.Now().Format("20060102150405") + ".xlsx"
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

func (this *PayController) WithdrawSettingEdit() {
	if this.isPost() {
		ctype := this.GetString("ctype")
		setting := &entity.WithdrawSetting{Id: ctype}
		setting.JQWithdrawMax = utils.ToInt64(this.GetString("jq_withdraw_max"))
		setting.JQWinMax = utils.ToInt64(this.GetString("jq_win_max"))
		setting.JQWinRateMax = utils.ToFloat64(this.GetString("jq_win_rate_max"))
		setting.IsBankRepeatable = this.GetString("is_bank_repeatable") == "1"

		service.PayService.UpdateWithdrawSetting(setting)
		this.redirect(beego.URLFor("PayController.WithdrawList"))
		return
	}

	ctype := this.GetString("ctype")
	this.Data["ctype"] = ctype
	this.Data["withdrawSetting"] = service.PayService.GetWithdrawSettingById(ctype)
	this.Data["pageTitle"] = "提现机审配置修改"
	this.display()
}

// 提现审核通过
func (this *PayController) WithdrawAudit() {
	orderids := this.GetString("id")
	if orderids == "" {
		this.checkError(errors.New("订单ID不能为空"))
	}
	name := this.auth.GetUser().UserName

	orderidList := strings.Split(orderids, ",")
	for _, orderid := range orderidList {
		info := new(entity.WithdrawOrder)
		info.OrderID = orderid
		info.ExamineAccount = name
		info.ExamineWay = 2
		info.ETime = bson.Now()
		err := service.PayService.UpdateAuditor(info)
		rec_info := new(entity.WithdrawOptRecord)
		rec_info.OrderId = orderid
		rec_info.OptName = name
		rec_info.OptId = 3
		rec_info.Ctime = info.ETime.Unix()
		err = service.PayService.AddWithdrawOptRecord(rec_info)
		this.checkError(err)
		msg := new(data.WithdrawOpreate)
		msg.OrderID = orderid
		msg.Op = 1 // 1：通过；2：退回；3：冻结
		result, err := service.GmRequest(pb.WebWithdraw, pb.CONFIG_UPSERT, msg)
		beego.Trace("result: ", result)
		if err != nil {
			this.checkError(err)
		}
		service.ActionService.Add("withdraw_audit", name,
			"", utils.String(orderid), utils.String(orderid), "")
	}
	this.showMsg("操作成功！", MSG_OK, beego.URLFor("PayController.WithdrawList"))
}

// 提现审核派单
func (this *PayController) WithdrawDispatch() {
	orderids := this.GetString("id")
	if orderids == "" {
		this.checkError(errors.New("订单ID不能为空"))
	}
	name := this.auth.GetUser().UserName

	orderidList := strings.Split(orderids, ",")
	for _, orderid := range orderidList {
		info := new(entity.WithdrawOrder)
		info.OrderID = orderid
		info.ExamineAccount = name
		info.ExamineWay = 2
		info.ETime = bson.Now()
		err := service.PayService.UpdateAuditor(info)
		rec_info := new(entity.WithdrawOptRecord)
		rec_info.OrderId = orderid
		rec_info.OptName = name
		rec_info.OptId = 3
		rec_info.Ctime = info.ETime.Unix()
		err = service.PayService.AddWithdrawOptRecord(rec_info)
		this.checkError(err)
		msg := new(data.WithdrawOpreate)
		msg.OrderID = orderid
		msg.Op = 5 // 1：通过；2：退回；3：冻结 5:派单
		result, err := service.GmRequest(pb.WebWithdraw, pb.CONFIG_UPSERT, msg)
		beego.Trace("result: ", result)
		if err != nil {
			this.checkError(err)
		}
		service.ActionService.Add("withdraw_dispatch", name,
			"", utils.String(orderid), utils.String(orderid), "")
	}
	this.showMsg("操作成功！", MSG_OK, beego.URLFor("PayController.WithdrawList"))
}

// 提现冻结
func (this *PayController) WithdrawFrozen() {
	orderids := this.GetString("id")
	if orderids == "" {
		this.checkError(errors.New("订单ID不能为空"))
	}
	name := this.auth.GetUser().UserName

	orderidList := strings.Split(orderids, ",")
	for _, orderid := range orderidList {
		info := new(entity.WithdrawOrder)
		info.OrderID = orderid
		info.ExamineAccount = name
		info.ExamineWay = 2
		info.ETime = bson.Now()
		err := service.PayService.UpdateAuditor(info)
		rec_info := new(entity.WithdrawOptRecord)
		rec_info.OrderId = orderid
		rec_info.OptName = name
		rec_info.OptId = 6
		rec_info.Ctime = info.ETime.Unix()
		err = service.PayService.AddWithdrawOptRecord(rec_info)
		this.checkError(err)
		msg := new(data.WithdrawOpreate)
		msg.OrderID = orderid
		msg.Op = 3 // 1：通过；2：退回；3：冻结
		result, err := service.GmRequest(pb.WebWithdraw, pb.CONFIG_UPSERT, msg)
		beego.Trace("result: ", result)
		if err != nil {
			this.checkError(err)
		}
		service.ActionService.Add("withdraw_frozen", name,
			"", utils.String(orderid), utils.String(orderid), "")
	}
	this.showMsg("操作成功！", MSG_OK, beego.URLFor("PayController.WithdrawList"))
	this.display()
}

// 提现退回
func (this *PayController) WithdrawReturn() {
	orderids := this.GetString("id")
	if orderids == "" {
		this.checkError(errors.New("订单ID不能为空"))
	}
	name := this.auth.GetUser().UserName

	orderidList := strings.Split(orderids, ",")
	for _, orderid := range orderidList {
		info := new(entity.WithdrawOrder)
		info.OrderID = orderid
		info.ExamineAccount = name
		info.ExamineWay = 2
		info.ETime = bson.Now()
		err := service.PayService.UpdateAuditor(info)
		rec_info := new(entity.WithdrawOptRecord)
		rec_info.OrderId = orderid
		rec_info.OptName = name
		rec_info.OptId = 5
		rec_info.Ctime = info.ETime.Unix()
		err = service.PayService.AddWithdrawOptRecord(rec_info)
		this.checkError(err)

		msg := new(data.WithdrawOpreate)
		msg.OrderID = orderid
		msg.Op = 2
		result, err := service.GmRequest(pb.WebWithdraw, pb.CONFIG_UPSERT, msg)
		beego.Trace("result: ", result)
		if err != nil {
			beego.Error(err)
			this.checkError(err)
		}

		service.ActionService.Add("withdraw_return", name,
			"", utils.String(orderid), utils.String(orderid), "")
	}
	this.showMsg("操作成功！", MSG_OK, beego.URLFor("PayController.WithdrawList"))
	this.display()
}

// 提现封号
func (this *PayController) WithdrawBan() {
	orderids := this.GetString("id")
	if orderids == "" {
		this.checkError(errors.New("订单ID不能为空"))
	}
	name := this.auth.GetUser().UserName

	orderidList := strings.Split(orderids, ",")
	for _, orderid := range orderidList {
		order, err := service.PayService.GetWithdrawByOrder(orderid)
		this.checkError(err)

		// 添加黑名单
		info := &entity.WithdrawUserBlacklist{
			Userid:   order.Userid,
			Operator: name,
		}
		err = service.PayService.AddWithdrawUserBlacklist(info)
		this.checkError(err)

		// 提现黑名单
		user := &entity.PlayerUser{Userid: order.Userid, Status: 3}
		err = service.PlayerService.UpdateUserStatus(user)
		this.checkError(err)
		// 登录黑名单
		reqMsg := &entity.BlackList{
			Userid: user.Userid,
			Status: user.Status,
		}
		result, err := service.GmRequest(pb.WebBlack, pb.CONFIG_UPSERT, reqMsg)
		beego.Trace("result: ", result)
		this.checkError(err)

		rec_info := new(entity.WithdrawOptRecord)
		rec_info.OrderId = orderid
		rec_info.OptName = name
		rec_info.OptId = 9
		rec_info.Ctime = bson.Now().Unix()
		err = service.PayService.AddWithdrawOptRecord(rec_info)
		this.checkError(err)

		// 所有订单添加标记 已封号
		m := bson.M{"userid": order.Userid}
		withdraws, err := service.PayService.GetByWithdrawUser(m)
		this.checkError(err)

		now := service.NowTime()
		nowTimeF := fmt.Sprintf("%02d:%02d:%02d", now.Hour(), now.Minute(), now.Second())
		for _, w := range withdraws {
			diffDay := (now.Unix() - w.Ctime.Unix()) / 60 / 60 / 24
			if strings.Contains(w.ErrorMsg, "已封号") {
				continue
			}

			msg := fmt.Sprintf("%d:%s:%s:%s", diffDay, nowTimeF, name, "已封号")
			if w.ErrorMsg != "" {
				msg = w.ErrorMsg + ";" + msg
			}
			info := new(entity.WithdrawOrder)
			info.OrderID = w.OrderID
			info.ExamineAccount = name
			info.ExamineWay = 1
			info.IsTag = true
			info.ErrorMsg = msg
			info.ETime = bson.Now()
			err := service.PayService.UpdateTag(info)
			this.checkError(err)

			// 待审核订单全部执行冻结操作(后续的由自动审核操作)
			if w.OrderStatus == 1 {
				msg := new(data.WithdrawOpreate)
				msg.OrderID = w.OrderID
				msg.Op = 3 // 1：通过；2：退回；3：冻结
				result, err := service.GmRequest(pb.WebWithdraw, pb.CONFIG_UPSERT, msg)
				beego.Trace("result: ", result)
				if err != nil {
					beego.Error(fmt.Sprintf("冻结订单[%s]失败err: %v", orderid, err))
				}
			}
		}

		service.ActionService.Add("withdraw_ban", name,
			"", utils.String(orderid), utils.String(orderid), "")
	}
	this.showMsg("操作成功！", MSG_OK, beego.URLFor("PayController.WithdrawList"))
	this.display()
}

// 提现标记
func (this *PayController) WithdrawTag() {
	page, _ := strconv.Atoi(this.GetString("page"))
	if page < 1 {
		page = 1
	}
	orderid := this.GetString("id")
	tag := this.GetString("tag")
	if orderid == "" {
		this.checkError(errors.New("订单ID不能为空"))
	}
	if tag == "" {
		this.checkError(errors.New("标记信息不能为空"))
	}
	orderinfo, err := service.PayService.GetByWithdrawByOrder(orderid)
	var tagVal = false
	if orderinfo != nil {
		if orderinfo.IsTag {
			tagVal = false
		} else {
			tagVal = true
		}
	}
	name := this.auth.GetUser().UserName
	orderinfo.OrderID = orderid
	orderinfo.ExamineWay = 2
	orderinfo.IsTag = tagVal
	orderinfo.ExamineAccount = name
	orderinfo.ETime = bson.Now()

	// 标记信息
	now := service.NowTime()
	nowTimeF := fmt.Sprintf("%02d:%02d:%02d", now.Hour(), now.Minute(), now.Second())
	diffDay := (now.Unix() - orderinfo.Ctime.Unix()) / 60 / 60 / 24
	tag = fmt.Sprintf("%d:%s:%s:%s", diffDay, nowTimeF, name, tag)
	if orderinfo.ErrorMsg != "" {
		orderinfo.ErrorMsg += ";" + tag
	} else {
		orderinfo.ErrorMsg = tag
	}
	err = service.PayService.UpdateTag(orderinfo)
	rec_info := new(entity.WithdrawOptRecord)
	rec_info.OrderId = orderid
	rec_info.OptName = name
	rec_info.OptId = 4
	rec_info.Ctime = orderinfo.ETime.Unix()
	err = service.PayService.AddWithdrawOptRecord(rec_info)
	if err != nil {
		this.checkError(err)
	}
	service.ActionService.Add("withdraw_tag", name,
		"", utils.String(orderid), utils.String(orderid), "")

	this.showMsg("操作成功！", MSG_OK, beego.URLFor("PayController.WithdrawList"))
	this.display()
}

// 提现订单转单
func (this *PayController) TransferOrder() {
	orderid := this.GetString("id")
	if orderid == "" {
		this.checkError(errors.New("订单ID不能为空"))
	}
	name := this.auth.GetUser().UserName
	info := new(entity.WithdrawOrder)
	info.OrderID = orderid
	info.ExamineAccount = name
	info.ExamineWay = 2
	info.ETime = bson.Now()
	err := service.PayService.UpdateAuditor(info)
	rec_info := new(entity.WithdrawOptRecord)
	rec_info.OrderId = orderid
	rec_info.OptName = name
	rec_info.OptId = 7
	rec_info.Ctime = info.ETime.Unix()
	err = service.PayService.AddWithdrawOptRecord(rec_info)
	this.checkError(err)

	msg := new(data.WithdrawOpreate)
	msg.OrderID = orderid
	msg.Op = 4
	msg.UserName = name
	result, err := service.GmRequest(pb.WebWithdraw, pb.CONFIG_UPSERT, msg)
	beego.Trace("result: ", result)
	if err != nil {
		this.checkError(err)
	}
	service.ActionService.Add("withdraw_transfer_order", name,
		"", utils.String(orderid), utils.String(orderid), "")
	this.showMsg("操作成功！", MSG_OK, beego.URLFor("PayController.WithdrawList"))
	this.display()
}

// 转单订单详情
func (this *PayController) TransferDetail() {
	orderid := this.GetString("id")
	if orderid == "" {
		this.checkError(errors.New("订单ID不能为空"))
	}
	orderinfo, _ := service.PayService.GetWithdrawByOrder(orderid)
	var buf bytes.Buffer
	buf.WriteString("<div class=\"table-panel\">")
	buf.WriteString("<table class=\"table table-striped table-bordered table-hover\">")
	buf.WriteString("<tr>")
	buf.WriteString("<th>订单ID</th>")
	buf.WriteString("<th>第三方订单ID</th>")
	buf.WriteString("<th>支付渠道</th>")
	buf.WriteString("<th>转单类型</th>")
	buf.WriteString("<th>转单原因</th>")
	buf.WriteString("<th>订单状态</th>")
	buf.WriteString("<th>操作人</th>")
	buf.WriteString("<th>创建时间</th>")
	buf.WriteString("<th>修改时间</th>")
	buf.WriteString("</tr>")
	isNull := true
	if orderinfo != nil {
		if len(orderinfo.TransferDetail) > 0 {
			for i := len(orderinfo.TransferDetail) - 1; i >= 0; i-- {
				item := orderinfo.TransferDetail[i]
				ctime := ""
				utime := ""
				buf.WriteString("<tr>")
				buf.WriteString("<td>" + orderid + "</td>")
				buf.WriteString("<td>" + orderinfo.OutTradeNo + "</td>")
				buf.WriteString("<td>" + item.ChannelName + "</td>")
				buf.WriteString("<td>" + item.TTypeName + "</td>")
				buf.WriteString("<td>" + item.ReasonName + "</td>")
				buf.WriteString("<td>" + item.StatusName + "</td>")
				buf.WriteString("<td>" + item.UserName + "</td>")
				if item.CTime != 0 {
					c, _ := service.ConvertToIndiaTime(item.CTime)
					ctime = c.Format("2006-01-02 15:04:05")
				}
				if item.UTime != 0 {
					u, _ := service.ConvertToIndiaTime(item.UTime)
					utime = u.Format("2006-01-02 15:04:05")
				}
				buf.WriteString("<td>" + ctime + "</td>")
				buf.WriteString("<td>" + utime + "</td>")
				buf.WriteString("</tr>")
			}
			isNull = false
		}
	}
	if isNull {
		// 没有数据 默认一条
		buf.WriteString("<tr>")
		buf.WriteString("<td colspan=\"20\">暂无转单数据</td>")
		buf.WriteString("</tr>")
	}

	buf.WriteString("</table>")
	buf.WriteString("</div>")
	// 将查询结果渲染到模板中
	this.Ctx.WriteString(buf.String())
}

// 渠道成功率
func (this *PayController) Channel() {
	status, _ := this.GetInt("status")
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	pay_id := this.GetString("pay_id")
	tabid, _ := this.GetInt("tabid")
	packageId := this.GetString("package_id")
	aliasId := this.GetString("alias_id")
	if page < 1 {
		page = 1
	}
	channellist, _ := service.GameService.GetPayChannel()
	info := new(entity.PayChannel)
	info.Id = "0"
	info.Name = "全部"
	newSlice := make([]entity.PayChannel, 0)
	newSlice = append(newSlice, *info)
	payList := append(newSlice, channellist...)
	isChannel := false
	if startDate == "" && endDate == "" {
		// 默认当天
		today := bson.Now()
		startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -9).Format("2006-01-02"))
		endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
	}
	m := service.FindByDate1(startDate, endDate, "date", "date")
	if status == 0 {
		my_data := this.GetSession("my_select_pakeageid")
		if my_data != nil && packageId == "" {
			if packid, ok := my_data.(string); ok {
				packageId = packid
			}
		}
		status = 1
	}
	if pay_id != "" && pay_id != "0" {
		id, _ := strconv.Atoi(pay_id)
		m["channel"] = id
		isChannel = true
	} else {
		m["channel"] = bson.M{"$ne": 0}
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
			m["package_id"] = bson.M{"$in": temp_arr}
		} else {
			m["package_id"] = packageId
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
			m["package_name"] = ""
		} else {
			if strings.Contains(aliasId, ",") {
				palkage_arr := strings.Split(aliasId, ",")
				temp_arr := make([]string, 0)
				for _, item := range palkage_arr {
					if item != "" && item != "0" && item != "-" {
						temp_arr = append(temp_arr, item)
					}
				}
				m["package_name"] = bson.M{"$in": temp_arr}
			} else {
				m["package_name"] = aliasId
			}
			isChannel = true
		}
	}
	if tabid == 0 {
		dates := make([]interface{}, 0)    // 日期
		data := make([]interface{}, 0)     // XDPAY
		data1 := make([]interface{}, 0)    // MLPAY
		data2 := make([]interface{}, 0)    // 1916pay
		data3 := make([]interface{}, 0)    // LestPay
		data4 := make([]interface{}, 0)    // KingPay
		data5 := make([]interface{}, 0)    // XFPAY
		data6 := make([]interface{}, 0)    // SailsPay
		data7 := make([]interface{}, 0)    // FLYPAY
		data8 := make([]interface{}, 0)    // Uwinpay
		data20 := make([]interface{}, 0)   // RamaPay
		data21 := make([]interface{}, 0)   // IcePay
		data22 := make([]interface{}, 0)   // Wepay
		data23 := make([]interface{}, 0)   // OePay
		data24 := make([]interface{}, 0)   // 9sPay
		data25 := make([]interface{}, 0)   // BLIZZARDPY
		data26 := make([]interface{}, 0)   // metagopay
		data27 := make([]interface{}, 0)   // universalpay
		data5002 := make([]interface{}, 0) // usdt
		list, _ := service.StatisticsService.GetSuccessRateChart(m)
		if len(list) > 0 {
			dataTotals := make(map[int64]entity.PaySuccessRate)
			for _, item := range list {
				if _, ok := dataTotals[item.Date]; ok {
					// 如果日期在 map 中，则累加字段值
					var tempInfo = dataTotals[item.Date]
					tempInfo.Date = item.Date
					switch item.Channel {
					case 3010:
						// XDPAY
						tempInfo.XdPayRate = item.PaySuccessRate
					case 3011:
						// MLPAY
						tempInfo.MlPayRate = item.PaySuccessRate
					case 3012:
						// 1916Pay
						tempInfo.Pay1916Rate = item.PaySuccessRate
					case 3013:
						// LetsPay
						tempInfo.LetsPayRate = item.PaySuccessRate
					case 3014:
						// KingPay
						tempInfo.KingPayRate = item.PaySuccessRate
					case 3015:
						// XFPAY
						tempInfo.XfPayRate = item.PaySuccessRate
					case 3016:
						// SailsPAY
						tempInfo.SailsPayRate = item.PaySuccessRate
					case 3017:
						// FLYPAY
						tempInfo.FlyPayRate = item.PaySuccessRate
					case 3018:
						// UwinPay
						tempInfo.UwinPayRate = item.PaySuccessRate
					case 3020:
						// RamaPay
						tempInfo.RamaPayRate = item.PaySuccessRate
					case 3021:
						// IcePay
						tempInfo.IcePayRate = item.PaySuccessRate
					case 3022:
						// WePay
						tempInfo.WePayRate = item.PaySuccessRate
					case 3023:
						// OePay
						tempInfo.OePayRate = item.PaySuccessRate
					case 3024:
						// 9sPay
						tempInfo.Pay9sRate = item.PaySuccessRate
					case 3025:
						// Blizzardpy
						tempInfo.BlizzardpyRate = item.PaySuccessRate
					case 3026:
						// METAGOPAY
						tempInfo.MetagopayRate = item.PaySuccessRate
					case 3027:
						// UNIVERSALPAY
						tempInfo.UniversalpayRate = item.PaySuccessRate
					case 5002:
						// USDT
						tempInfo.USDTRate = item.PaySuccessRate
					}
					dataTotals[item.Date] = tempInfo
				} else {
					// 如果日期不在 map 中，则添加新的日期和字段值
					tempInfo := new(entity.PaySuccessRate)
					tempInfo.Date = item.Date
					switch item.Channel {
					case 3010:
						// XDPAY
						tempInfo.XdPayRate = item.PaySuccessRate
					case 3011:
						// MLPAY
						tempInfo.MlPayRate = item.PaySuccessRate
					case 3012:
						// 1916Pay
						tempInfo.Pay1916Rate = item.PaySuccessRate
					case 3013:
						// LetsPay
						tempInfo.LetsPayRate = item.PaySuccessRate
					case 3014:
						// KingPay
						tempInfo.KingPayRate = item.PaySuccessRate
					case 3015:
						// XFPAY
						tempInfo.XfPayRate = item.PaySuccessRate
					case 3016:
						// SailsPAY
						tempInfo.SailsPayRate = item.PaySuccessRate
					case 3017:
						// FLYPAY
						tempInfo.FlyPayRate = item.PaySuccessRate
					case 3018:
						// UwinPay
						tempInfo.UwinPayRate = item.PaySuccessRate
					case 3020:
						// RamaPay
						tempInfo.RamaPayRate = item.PaySuccessRate
					case 3021:
						// IcePay
						tempInfo.IcePayRate = item.PaySuccessRate
					case 3022:
						// WePay
						tempInfo.WePayRate = item.PaySuccessRate
					case 3023:
						// OePay
						tempInfo.OePayRate = item.PaySuccessRate
					case 3024:
						// 9sPay
						tempInfo.Pay9sRate = item.PaySuccessRate
					case 3025:
						// Blizzardpy
						tempInfo.BlizzardpyRate = item.PaySuccessRate
					case 3026:
						// METAGOPAY
						tempInfo.MetagopayRate = item.PaySuccessRate
					case 3027:
						// UNIVERSALPAY
						tempInfo.UniversalpayRate = item.PaySuccessRate
					case 5002:
						// usdt
						tempInfo.USDTRate = item.PaySuccessRate
					}
					dataTotals[item.Date] = *tempInfo
				}

			}
			// 提取键到切片
			keys := make([]int64, 0, len(dataTotals))
			for key := range dataTotals {
				keys = append(keys, key)
			}

			// 对键进行排序
			sort.Slice(keys, func(i, j int) bool {
				return keys[i] < keys[j]
			})
			for _, key := range keys {
				tempInfo := dataTotals[key]
				t := time.Unix(tempInfo.Date, 0)
				strTime := t.Format("2006-01-02")
				dates = append(dates, strTime)
				data = append(data, fmt.Sprintf("%.2f", tempInfo.XdPayRate))
				data1 = append(data1, fmt.Sprintf("%.2f", tempInfo.MlPayRate))
				data2 = append(data2, fmt.Sprintf("%.2f", tempInfo.Pay1916Rate))
				data3 = append(data3, fmt.Sprintf("%.2f", tempInfo.LetsPayRate))
				data4 = append(data4, fmt.Sprintf("%.2f", tempInfo.KingPayRate))
				data5 = append(data5, fmt.Sprintf("%.2f", tempInfo.XfPayRate))
				data6 = append(data6, fmt.Sprintf("%.2f", tempInfo.SailsPayRate))
				data7 = append(data7, fmt.Sprintf("%.2f", tempInfo.FlyPayRate))
				data8 = append(data8, fmt.Sprintf("%.2f", tempInfo.UwinPayRate))
				data20 = append(data20, fmt.Sprintf("%.2f", tempInfo.RamaPayRate))
				data21 = append(data21, fmt.Sprintf("%.2f", tempInfo.IcePayRate))
				data22 = append(data22, fmt.Sprintf("%.2f", tempInfo.WePayRate))
				data23 = append(data23, fmt.Sprintf("%.2f", tempInfo.OePayRate))
				data24 = append(data24, fmt.Sprintf("%.2f", tempInfo.Pay9sRate))
				data25 = append(data25, fmt.Sprintf("%.2f", tempInfo.BlizzardpyRate))
				data26 = append(data26, fmt.Sprintf("%.2f", tempInfo.MetagopayRate))
				data27 = append(data27, fmt.Sprintf("%.2f", tempInfo.UniversalpayRate))
				data5002 = append(data5002, fmt.Sprintf("%.2f", tempInfo.USDTRate))
			}
		}
		this.Data["TimedLabel"] = dates
		this.Data["xdpay"] = data
		this.Data["mlpay"] = data1
		this.Data["pay1916"] = data2
		this.Data["letspay"] = data3
		this.Data["kingpay"] = data4
		this.Data["xfpay"] = data5
		this.Data["sailspay"] = data6
		this.Data["flypay"] = data7
		this.Data["uwinpay"] = data8
		this.Data["ramapay"] = data20
		this.Data["icepay"] = data21
		this.Data["wepay"] = data22
		this.Data["oepay"] = data23
		this.Data["pay9s"] = data24
		this.Data["blizzardpy"] = data25
		this.Data["metagopay"] = data26
		this.Data["universalpay"] = data27
		this.Data["usdt"] = data5002
	} else {
		count, _ := service.StatisticsService.GetSuccessRateTotal(m, isChannel)
		list, _ := service.StatisticsService.GetSuccessRateList(page, this.pageSize, m, isChannel)
		for i, item := range list {
			if item.Channel == 0 {
				item.ChannelName = "全部"
			} else {
				for _, v := range payList {
					i, err := strconv.ParseInt(v.Id, 10, 64)
					if err != nil {
						i = 0
					}
					if item.Channel == i {
						item.ChannelName = v.Name
					}
				}
			}
			list[i] = item
		}
		this.Data["count"] = count
		this.Data["list"] = list
		this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("PayController.Channel", "tabid", tabid, "status", status, "pay_id", pay_id, "start_date", startDate, "end_date", endDate), true).ToString()
	}

	packageList, _ := service.ChannelService.GetChannelAll()
	packageAliasList, _ := service.ChannelService.GetChannelAliasAll()
	this.Data["pageTitle"] = "渠道成功率"
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["payList"] = payList
	this.Data["payId"] = pay_id
	this.Data["packageId"] = packageId
	this.Data["packageList"] = packageList
	this.Data["aliasId"] = aliasId
	this.Data["packageAliasList"] = packageAliasList
	this.Data["status"] = status
	this.Data["tabid"] = tabid
	this.display()
}

// 修改金币日志列表
func (this *PayController) CoinLog() {
	status, _ := this.GetInt("status")
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	userid := this.GetString("userid")
	packageId := this.GetString("package_id")
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
	m := service.FindByDate(startDate, endDate, "ctime", "ctime")
	if userid != "" {
		m["user_id"] = userid
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
		// 存储缓存
		this.SetSession("my_select_pakeageid", packageId)
	} else {
		// 清除缓存
		this.DelSession("my_select_pakeageid")

	}
	count, _ := service.PayService.GetCoinLogTotal(m)
	list, _ := service.PayService.CoinLogList(page, this.pageSize, m)

	// le := len(list)
	// for i := 0; i < le; i++ {
	// 	//转换为元
	// 	list[i].Score = list[i].Score / 100
	// 	list[i].CurScore = list[i].CurScore / 100
	// 	list[i].ChangeScore = list[i].ChangeScore / 100
	// }
	packageList, _ := service.ChannelService.GetChannelAll()
	this.Data["pageTitle"] = "修改金币日志"
	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["userid"] = userid
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("PayController.CoinLog", "status", status, "userid", userid, "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["packageId"] = packageId
	this.Data["packageList"] = packageList
	this.Data["status"] = status
	this.display()
}

// 提现配置
func (this *PayController) SetWithdraw() {
	page, _ := strconv.Atoi(this.GetString("page"))
	userid := this.GetString("userid")
	if page < 1 {
		page = 1
	}

	m := bson.M{}
	if userid != "" {
		m["userid"] = userid
	}
	count, _ := service.PayService.GetSetWithdrawTotal(m)
	list, _ := service.PayService.SetWithdrawList(page, this.pageSize, m)

	this.Data["pageTitle"] = "提现配置"
	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("PayController.SetWithdraw", "userid", userid), true).ToString()
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "setwithdrawopt")
	this.display()
}

// 添加 提现配置
func (this *PayController) SetWithdrawAdd() {
	if this.isPost() {
		// 提交
		name := this.GetString("name")
		limit, _ := strconv.Atoi(this.GetString("limit"))
		recharge_limit, _ := strconv.Atoi(this.GetString("recharge_limit"))
		flowing_limit, _ := strconv.Atoi(this.GetString("flowing_limit"))
		cost_diamond, _ := strconv.Atoi(this.GetString("cost_diamond"))
		get_cash, _ := strconv.Atoi(this.GetString("get_cash"))
		commission, _ := strconv.Atoi(this.GetString("commission"))
		sort_id, _ := strconv.Atoi(this.GetString("sort_id"))
		delay_time, _ := strconv.Atoi(this.GetString("delay_time"))

		with := new(entity.SetWithdraw)
		with.Name = name
		with.Limit = int32(limit)
		with.RechargeLimit = int64(recharge_limit)
		with.FlowingLimit = uint32(flowing_limit)
		with.CostDiamond = uint32(cost_diamond)
		with.GetCash = uint32(get_cash)
		with.Commission = uint32(commission)
		with.Index = int32(sort_id)
		with.DelayTime = int64(delay_time)
		err := this.validWithdraw(with)
		this.checkError(err)
		// remark := this.GetString("remark")
		// coin, err1 := this.GetInt("coin")
		// 新增
		err = service.PayService.AddSetWithdraw(with)
		this.checkError(err)

		// 通知服务器
		b := make(map[int32]entity.SetWithdraw)
		b[with.Id] = *with
		result, err := service.GmRequest(pb.WebSetWithDraw, pb.CONFIG_UPSERT, b)
		beego.Trace("result: ", result)
		if err != nil {
			this.checkError(err)
		} else {
			service.ActionService.Add("add_set_withdraw", this.auth.GetUser().UserName,
				"", utils.String(with.Id), utils.String(with.Id), "")
		}

		this.redirect(beego.URLFor("PayController.SetWithdraw"))
	} else {
		this.Data["pageTitle"] = "提现配置 -> 新增"
		this.display()
	}
}

// 修改 提现配置
func (this *PayController) SetWithdrawEdit() {
	id, _ := strconv.Atoi(this.GetString("id"))
	if this.isPost() {
		// 提交
		// id, _ := strconv.Atoi(this.GetString("id"))
		name := this.GetString("name")
		limit, _ := strconv.Atoi(this.GetString("limit"))
		recharge_limit, _ := strconv.Atoi(this.GetString("recharge_limit"))
		flowing_limit, _ := strconv.Atoi(this.GetString("flowing_limit"))
		cost_diamond, _ := strconv.Atoi(this.GetString("cost_diamond"))
		get_cash, _ := strconv.Atoi(this.GetString("get_cash"))
		commission, _ := strconv.Atoi(this.GetString("commission"))
		sort_id, _ := strconv.Atoi(this.GetString("sort_id"))
		delay_time, _ := strconv.Atoi(this.GetString("delay_time"))

		with := new(entity.SetWithdraw)
		with.Name = name
		with.Limit = int32(limit)
		with.RechargeLimit = int64(recharge_limit)
		with.FlowingLimit = uint32(flowing_limit)
		with.CostDiamond = uint32(cost_diamond)
		with.GetCash = uint32(get_cash)
		with.Commission = uint32(commission)
		with.Index = int32(sort_id)
		with.DelayTime = int64(delay_time)
		err := this.validWithdraw(with)
		this.checkError(err)

		if id != 0 {
			// 修改
			with.Id = int32(id)
			err = service.PayService.UpdateSetWithdraw(with)
			this.checkError(err)

			// 通知服务器
			b := make(map[int32]entity.SetWithdraw)
			b[with.Id] = *with
			result, err := service.GmRequest(pb.WebSetWithDraw, pb.CONFIG_UPSERT, b)
			beego.Trace("result: ", result)
			if err != nil {
				this.checkError(err)
			} else {
				service.ActionService.Add("update_set_withdraw", this.auth.GetUser().UserName,
					"", utils.String(with.Id), utils.String(with.Id), "")
			}
		}
		this.redirect(beego.URLFor("PayController.SetWithdraw"))
	} else {

		if id != 0 {
			info, err := service.PayService.GetSetWithdraw(id)
			this.checkError(err)
			this.Data["with"] = info
		}
		this.Data["pageTitle"] = "提现配置 -> 编辑"
		this.display()
	}
}

func (this *PayController) validWithdraw(with *entity.SetWithdraw) error {
	valid := validation.Validation{}
	valid.Required(with.Name, "name").Message("商品名称不能为空")
	// valid.Numeric(with.Limit, "limit").Message("提现次数限制不能低于0")
	// valid.Numeric(with.RechargeLimit, "recharge_limit").Message("充值金额限制不能低于0")
	// valid.Min(with.FlowingLimit, -1, "flowing_limit").Message("流水限制格式不正确")
	valid.Required(with.CostDiamond, "cost_diamond").Message("消耗彩金不能为空")
	valid.Required(with.GetCash, "get_cash").Message("提现金额不能为空")
	//valid.MinSize(int(with.Commission), -1, "commission").Message("手续费不能不能低于0")
	valid.Required(with.Index, "sort_id").Message("展示权重不能为空")
	if with.Limit < 0 {
		valid.Alpha(with.Limit, "limit").Message("提现次数限制不能低于0")
	}
	if with.RechargeLimit < 0 {
		valid.Alpha(with.RechargeLimit, "recharge_limit").Message("充值金额限制不能低于0")
	}
	if int32(with.Commission) < 0 {
		valid.Alpha(with.Commission, "commission").Message("手续费不能不能低于0")
	}
	if valid.HasErrors() {
		for _, err := range valid.Errors {
			return errors.New(err.Message)
		}
	}
	return nil
}

// 删除提现配置
func (this *PayController) SetWithdrawDel() {
	id, _ := strconv.Atoi(this.GetString("id"))
	info, err := service.PayService.GetSetWithdraw(id)
	this.checkError(err)

	err = service.PayService.DelSetWithdraw(info.Id)
	this.checkError(err)

	b := make(map[int32]entity.SetWithdraw)
	b[info.Id] = *info
	result, err := service.GmRequest(pb.WebSetWithDraw, pb.CONFIG_DELETE, b)
	beego.Trace("result: ", result)
	if err != nil {
		this.checkError(err)
	} else {
		service.ActionService.Add("del_set_withdraw", this.auth.GetUser().UserName,
			"", utils.String(info.Id), utils.String(info.Id), "")
	}
	this.TplName = "setwithdrawdel.tpl"
	this.redirect(beego.URLFor("PayController.SetWithdraw"))
}

// 充值配置列表
func (this *PayController) ShopList() {
	status, _ := this.GetInt("status")
	page, _ := this.GetInt("page")
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	if page < 1 {
		page = 1
	}

	// m := service.FindByDate(startDate, endDate, "ctime", "ctime")
	m := bson.M{}
	list, _ := service.PayService.GetShopList(page, this.pageSize, m)
	count, _ := service.PayService.GetShopListTotal(m)

	this.Data["pageTitle"] = "充值配置"
	this.Data["status"] = status
	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("PayController.ShopList", "status", status, "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["isOperation"] = false // this.auth.HasAccessPerm(this.controllerName, "shopopt")
	this.display()
}

// 添加
func (this *PayController) ShopAdd() {
	if this.isPost() {
		// shop := new(entity.Shop)
		// id := this.GetString("id")
		// status, _ := this.GetInt("status")
		// propid, _ := this.GetInt("propid")
		// payway, _ := this.GetInt("payway")
		// give, _ := this.GetInt("give")
		// givetype, _ := this.GetInt("givetype")
		// flow, _ := this.GetInt("flow")
		// number, _ := this.GetInt("number")
		// price, _ := this.GetInt("price")
		// name := this.GetString("name")
		// info := this.GetString("info")
		// shop.Id = string(id)
		// shop.Give = uint32(give) // 赠送
		// shop.GiveType = givetype
		// shop.FlowMultiple = flow
		// shop.Status = status
		// shop.Propid = propid
		// shop.Payway = payway
		// shop.Number = uint32(number)
		// shop.Price = uint32(price)
		// shop.Name = name
		// shop.Info = info
		// err := this.validShop(shop)
		// this.checkError(err)

		// err = service.PayService.AddShop(shop)
		// this.checkError(err)
		// newId := *&shop.Id
		// req := map[string]entity.Shop{
		// 	newId: *shop,
		// }
		// result, err := service.GmRequest(pb.WebShop, pb.CONFIG_UPSERT, req)
		// beego.Trace("result: ", result)
		// if err != nil {
		// 	this.checkError(err)
		// } else {
		// 	service.ActionService.AddShop(this.auth.GetUser().UserName, shop.Id, "")
		// }
		// this.redirect(beego.URLFor("PayController.ShopList"))
	}
	typeList := map[int]string{
		1: "直接赠送",
		2: "打码量",
	}
	this.Data["pageTitle"] = "添加配置"
	this.Data["typeList"] = typeList
	this.display()
}

// 修改
func (this *PayController) ShopEdit() {
	id := this.GetString("id")
	shop, err := service.PayService.GetShop(id)
	this.checkError(err)
	if this.isPost() {
		// // shop := new(entity.Shop)
		// status, _ := this.GetInt("status")
		// propid, _ := this.GetInt("propid")
		// payway, _ := this.GetInt("payway")
		// give, _ := this.GetInt("give")
		// givetype, _ := this.GetInt("givetype")
		// flow, _ := this.GetInt("flow")
		// number, _ := this.GetInt("number")
		// price, _ := this.GetInt("price")
		// name := this.GetString("name")
		// info := this.GetString("info")

		// shop.Id = string(id)
		// shop.Give = uint32(give) // 赠送
		// shop.GiveType = givetype
		// shop.FlowMultiple = flow
		// shop.Status = status
		// shop.Propid = propid
		// shop.Payway = payway
		// shop.Number = uint32(number)
		// shop.Price = uint32(price)
		// shop.Name = name
		// shop.Info = info
		// err := this.validShop(shop)
		// this.checkError(err)

		// err = service.PayService.UpdateShop(shop)
		// this.checkError(err)
		// newId := *&shop.Id
		// req := map[string]entity.Shop{
		// 	newId: *shop,
		// }
		// result, err := service.GmRequest(pb.WebShop, pb.CONFIG_UPSERT, req)
		// beego.Trace("result: ", result)
		// if err != nil {
		// 	this.checkError(err)
		// } else {
		// 	service.ActionService.AddShop(this.auth.GetUser().UserName, shop.Id, "")
		// }
		// this.redirect(beego.URLFor("PayController.ShopList"))
	}
	typeList := map[int]string{
		1: "直接赠送",
		2: "打码量",
	}
	this.Data["pageTitle"] = "编辑配置"
	this.Data["typeList"] = typeList
	this.Data["shop"] = shop
	this.display()
}

func (this *PayController) validShop(shop *entity.Shop) error {
	valid := validation.Validation{}
	valid.Required(shop.Name, "name").Message("物品名称不能为空")
	// valid.Required(shop.Info, "info").Message("物品描述不能为空")
	valid.Required(shop.Number, "number").Message("购买数量不能为空")
	valid.Required(shop.Price, "price").Message("购买价格不能为空")

	if valid.HasErrors() {
		for _, err := range valid.Errors {
			return errors.New(err.Message)
		}
	}

	return nil
}

// 商品
func (this *PayController) Shop() {
	// id := this.GetString("id")

	// shop, err := service.PayService.GetShop(id)
	// this.checkError(err)

	// reqMsg := &entity.ReqShopMsg{
	// 	Id:     shop.Id,     //购买ID
	// 	Status: shop.Status, //物品状态,1=热卖
	// 	Propid: shop.Propid, //兑换的物品,1=钻石
	// 	Payway: shop.Payway, //支付方式,1=RMB
	// 	Number: shop.Number, //兑换的数量
	// 	Give:   shop.Give,   // 赠送数量
	// 	Price:  shop.Price,  //支付价格
	// 	Name:   shop.Name,   //物品名字
	// 	Info:   shop.Info,   //物品信息
	// 	Del:    shop.Del,    //是否移除
	// 	Etime:  shop.Ctime,  //过期时间
	// 	Ctime:  shop.Ctime,  //创建时间
	// }
	// data, err1 := json.Marshal(reqMsg)
	// this.checkError(err1)
	// _, err2 := service.Gm("ReqShopMsg", string(data))
	// this.checkError(err2)

	// service.ActionService.Shop(this.auth.GetUserName(), id, "")

	// this.redirect(beego.URLFor("PayController.ShopList"))
}

// 移除商品
func (this *PayController) ShopDel() {
	// id := this.GetString("id")

	// shop, err := service.PayService.GetShop(id)
	// this.checkError(err)

	// shop.Del = 1
	// err = service.PayService.DelShop(shop.Id)
	// this.checkError(err)

	// newId := *&shop.Id
	// req := map[string]entity.Shop{
	// 	newId: *shop,
	// }
	// result, err := service.GmRequest(pb.WebShop, pb.CONFIG_DELETE, req)
	// beego.Trace("result: ", result)
	// if err != nil {
	// 	this.checkError(err)
	// } else {
	// 	service.ActionService.DelShop(this.auth.GetUserName(), id, "")
	// }
	// this.redirect(beego.URLFor("PayController.ShopList"))
}

// 支付黑名单列表
func (this *PayController) PayBlackList() {
	page, _ := strconv.Atoi(this.GetString("page"))
	if page < 1 {
		page = 1
	}

	m := bson.M{}
	count, _ := service.PayService.GetPayBlackListTotal(m)
	list, _ := service.PayService.GetPayBlackList(page, this.pageSize, m)

	this.Data["pageTitle"] = "支付黑名单"
	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("PayController.PayBlackList"), true).ToString()
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "payblacklistopt")
	this.display()
}

// 添加支付黑名单
func (this *PayController) PayBlackListAdd() {
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
		this.redirect(beego.URLFor("PayController.PayBlackList"))
	}
}

// 删除支付黑名单
func (this *PayController) PayBlackListDel() {
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
	this.redirect(beego.URLFor("PayController.PayBlackList"))
}

// 支付黑名单批量上传
func (this *PayController) PayBlackListUpload() {
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
	this.redirect(beego.URLFor("PayController.PayBlackList"))
}

// // 判断是否是印度电话号码
// func validateIndianPhoneNumber(phoneNumber string) bool {
// 	regex := `^[6789]\d{9}$`
// 	match, _ := regexp.MatchString(regex, phoneNumber)
// 	return match
// }

// 提现汇总统计
func (this *PayController) WithdrawStats() {
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	money_statr := this.GetString("money_statr")
	money_end := this.GetString("money_end")
	if page < 1 {
		page = 1
	}
	if startDate == "" && endDate == "" {
		// 默认近七天数据
		today := bson.Now()
		startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -6).Format("2006-01-02"))
		endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
	}
	if startDate == "" && endDate != "" {
		date, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			this.checkError(errors.New("开始时间不能为空"))
		}
		startDate = fmt.Sprintf("%s", date.AddDate(0, 0, -6).Format("2006-01-02"))
		// this.checkError(errors.New("开始时间不能为空"))
	}
	if endDate == "" && startDate != "" {
		date, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			this.checkError(errors.New("结束时间不能为空"))
		}
		endDate = fmt.Sprintf("%s", date.AddDate(0, 0, 6).Format("2006-01-02"))
	}
	moneyStatr, _ := strconv.Atoi(money_statr)
	moneyEnd, _ := strconv.Atoi(money_end)
	list, _ := service.PayService.GetWithdrawStats(startDate, endDate, moneyStatr, moneyEnd)
	count := len(list)
	this.Data["pageTitle"] = "提现汇总统计"
	this.Data["list"] = list
	this.Data["count"] = count
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["money_statr"] = money_statr
	this.Data["money_end"] = money_end
	this.display()
}

// 支付渠道统计
func (this *PayController) PayChannelStats() {
	// page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	channelId := this.GetString("channel_id")
	typeId, _ := this.GetInt("typeId")
	aliasId := this.GetString("alias_id")
	classId := this.GetString("class_id")
	money_statr := this.GetString("money_statr")
	money_end := this.GetString("money_end")

	// var params = make(map[string]any)
	if startDate == "" || endDate == "" {
		// 默认近七天数据
		today := bson.Now()
		endDate = today.Format("2006-01-02")
		startDate = today.AddDate(0, 0, -6).Format("2006-01-02")
	}
	startTime, endTime := service.FindByDate4(startDate, endDate)
	m := service.FindByDate1(startDate, endDate, "date", "date")
	// 支付渠道
	var payChannels []string
	if channelId != "" && channelId != "0" && channelId != "-" {
		palkage_arr := strings.Split(channelId, ",")
		for _, item := range palkage_arr {
			if item != "" && item != "0" && item != "-" {
				payChannels = append(payChannels, item)
			}
		}
	}
	if len(payChannels) > 0 {
		var chs []int
		for _, ch := range payChannels {
			ch2, _ := strconv.Atoi(ch)
			chs = append(chs, ch2)
		}
		m["channel"] = bson.M{"$in": chs}
	}

	var packageIds []string
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
	if len(packageIds) > 0 {
		m["package_id"] = bson.M{"$in": packageIds}
	}

	if typeId == 0 {
		// 收付汇总
		list, err := service.PayService.PayChannelCollect(*startTime, *endTime, packageIds, payChannels)
		if err != nil {
			beego.Error("PayChannelCollect error: ", err)
		}
		this.Data["list"] = list
		this.Data["count"] = len(list)
	} else if typeId == 1 {
		// 成功率走势图
		dates := make([]interface{}, 0)    // 日期
		data := make([]interface{}, 0)     // XDPAY
		data1 := make([]interface{}, 0)    // MLPAY
		data2 := make([]interface{}, 0)    // 1916pay
		data3 := make([]interface{}, 0)    // LestPay
		data4 := make([]interface{}, 0)    // KingPay
		data5 := make([]interface{}, 0)    // XFPAY
		data6 := make([]interface{}, 0)    // SailsPay
		data7 := make([]interface{}, 0)    // FLYPAY
		data8 := make([]interface{}, 0)    // Uwinpay
		data20 := make([]interface{}, 0)   // RamaPay
		data21 := make([]interface{}, 0)   // IcePay
		data22 := make([]interface{}, 0)   // Wepay
		data23 := make([]interface{}, 0)   // OePay
		data24 := make([]interface{}, 0)   // 9sPay
		data25 := make([]interface{}, 0)   // BLIZZARDPY
		data26 := make([]interface{}, 0)   // metagopay
		data27 := make([]interface{}, 0)   // universalpay
		data5002 := make([]interface{}, 0) // usdt
		list, _ := service.StatisticsService.GetSuccessRateChart(m)
		if len(list) > 0 {
			dataTotals := make(map[int64]entity.PaySuccessRate)
			for _, item := range list {
				if _, ok := dataTotals[item.Date]; ok {
					// 如果日期在 map 中，则累加字段值
					var tempInfo = dataTotals[item.Date]
					tempInfo.Date = item.Date
					switch item.Channel {
					case 3010:
						// XDPAY
						tempInfo.XdPayRate = item.PaySuccessRate
					case 3011:
						// MLPAY
						tempInfo.MlPayRate = item.PaySuccessRate
					case 3012:
						// 1916Pay
						tempInfo.Pay1916Rate = item.PaySuccessRate
					case 3013:
						// LetsPay
						tempInfo.LetsPayRate = item.PaySuccessRate
					case 3014:
						// KingPay
						tempInfo.KingPayRate = item.PaySuccessRate
					case 3015:
						// XFPAY
						tempInfo.XfPayRate = item.PaySuccessRate
					case 3016:
						// SailsPAY
						tempInfo.SailsPayRate = item.PaySuccessRate
					case 3017:
						// FLYPAY
						tempInfo.FlyPayRate = item.PaySuccessRate
					case 3018:
						// UwinPay
						tempInfo.UwinPayRate = item.PaySuccessRate
					case 3020:
						// RamaPay
						tempInfo.RamaPayRate = item.PaySuccessRate
					case 3021:
						// IcePay
						tempInfo.IcePayRate = item.PaySuccessRate
					case 3022:
						// WePay
						tempInfo.WePayRate = item.PaySuccessRate
					case 3023:
						// OePay
						tempInfo.OePayRate = item.PaySuccessRate
					case 3024:
						// 9sPay
						tempInfo.Pay9sRate = item.PaySuccessRate
					case 3025:
						// Blizzardpy
						tempInfo.BlizzardpyRate = item.PaySuccessRate
					case 3026:
						// METAGOPAY
						tempInfo.MetagopayRate = item.PaySuccessRate
					case 3027:
						// UNIVERSALPAY
						tempInfo.UniversalpayRate = item.PaySuccessRate
					case 5002:
						// USDT
						tempInfo.USDTRate = item.PaySuccessRate
					}
					dataTotals[item.Date] = tempInfo
				} else {
					// 如果日期不在 map 中，则添加新的日期和字段值
					tempInfo := new(entity.PaySuccessRate)
					tempInfo.Date = item.Date
					switch item.Channel {
					case 3010:
						// XDPAY
						tempInfo.XdPayRate = item.PaySuccessRate
					case 3011:
						// MLPAY
						tempInfo.MlPayRate = item.PaySuccessRate
					case 3012:
						// 1916Pay
						tempInfo.Pay1916Rate = item.PaySuccessRate
					case 3013:
						// LetsPay
						tempInfo.LetsPayRate = item.PaySuccessRate
					case 3014:
						// KingPay
						tempInfo.KingPayRate = item.PaySuccessRate
					case 3015:
						// XFPAY
						tempInfo.XfPayRate = item.PaySuccessRate
					case 3016:
						// SailsPAY
						tempInfo.SailsPayRate = item.PaySuccessRate
					case 3017:
						// FLYPAY
						tempInfo.FlyPayRate = item.PaySuccessRate
					case 3018:
						// UwinPay
						tempInfo.UwinPayRate = item.PaySuccessRate
					case 3020:
						// RamaPay
						tempInfo.RamaPayRate = item.PaySuccessRate
					case 3021:
						// IcePay
						tempInfo.IcePayRate = item.PaySuccessRate
					case 3022:
						// WePay
						tempInfo.WePayRate = item.PaySuccessRate
					case 3023:
						// OePay
						tempInfo.OePayRate = item.PaySuccessRate
					case 3024:
						// 9sPay
						tempInfo.Pay9sRate = item.PaySuccessRate
					case 3025:
						// Blizzardpy
						tempInfo.BlizzardpyRate = item.PaySuccessRate
					case 3026:
						// METAGOPAY
						tempInfo.MetagopayRate = item.PaySuccessRate
					case 3027:
						// UNIVERSALPAY
						tempInfo.UniversalpayRate = item.PaySuccessRate
					case 5002:
						// usdt
						tempInfo.USDTRate = item.PaySuccessRate
					}
					dataTotals[item.Date] = *tempInfo
				}

			}
			// 提取键到切片
			keys := make([]int64, 0, len(dataTotals))
			for key := range dataTotals {
				keys = append(keys, key)
			}

			// 对键进行排序
			sort.Slice(keys, func(i, j int) bool {
				return keys[i] < keys[j]
			})
			for _, key := range keys {
				tempInfo := dataTotals[key]
				t := time.Unix(tempInfo.Date, 0)
				strTime := t.Format("2006-01-02")
				dates = append(dates, strTime)
				data = append(data, fmt.Sprintf("%.2f", tempInfo.XdPayRate))
				data1 = append(data1, fmt.Sprintf("%.2f", tempInfo.MlPayRate))
				data2 = append(data2, fmt.Sprintf("%.2f", tempInfo.Pay1916Rate))
				data3 = append(data3, fmt.Sprintf("%.2f", tempInfo.LetsPayRate))
				data4 = append(data4, fmt.Sprintf("%.2f", tempInfo.KingPayRate))
				data5 = append(data5, fmt.Sprintf("%.2f", tempInfo.XfPayRate))
				data6 = append(data6, fmt.Sprintf("%.2f", tempInfo.SailsPayRate))
				data7 = append(data7, fmt.Sprintf("%.2f", tempInfo.FlyPayRate))
				data8 = append(data8, fmt.Sprintf("%.2f", tempInfo.UwinPayRate))
				data20 = append(data20, fmt.Sprintf("%.2f", tempInfo.RamaPayRate))
				data21 = append(data21, fmt.Sprintf("%.2f", tempInfo.IcePayRate))
				data22 = append(data22, fmt.Sprintf("%.2f", tempInfo.WePayRate))
				data23 = append(data23, fmt.Sprintf("%.2f", tempInfo.OePayRate))
				data24 = append(data24, fmt.Sprintf("%.2f", tempInfo.Pay9sRate))
				data25 = append(data25, fmt.Sprintf("%.2f", tempInfo.BlizzardpyRate))
				data26 = append(data26, fmt.Sprintf("%.2f", tempInfo.MetagopayRate))
				data27 = append(data27, fmt.Sprintf("%.2f", tempInfo.UniversalpayRate))
				data5002 = append(data5002, fmt.Sprintf("%.2f", tempInfo.USDTRate))
			}
		}
		this.Data["TimedLabel"] = dates
		this.Data["xdpay"] = data
		this.Data["mlpay"] = data1
		this.Data["pay1916"] = data2
		this.Data["letspay"] = data3
		this.Data["kingpay"] = data4
		this.Data["xfpay"] = data5
		this.Data["sailspay"] = data6
		this.Data["flypay"] = data7
		this.Data["uwinpay"] = data8
		this.Data["ramapay"] = data20
		this.Data["icepay"] = data21
		this.Data["wepay"] = data22
		this.Data["oepay"] = data23
		this.Data["pay9s"] = data24
		this.Data["blizzardpy"] = data25
		this.Data["metagopay"] = data26
		this.Data["universalpay"] = data27
		this.Data["usdt"] = data5002
	} else if typeId == 2 {
		// 提现汇总
		moneyStatr, _ := strconv.Atoi(money_statr)
		moneyEnd, _ := strconv.Atoi(money_end)
		list, _ := service.PayService.GetWithdrawStats(startDate, endDate, moneyStatr, moneyEnd)
		count := len(list)
		this.Data["list"] = list
		this.Data["count"] = count
	}

	// if page < 1 {
	// 	page = 1
	// }
	// if startDate == "" && endDate == "" {
	// 	// 默认近3天数据
	// 	today := bson.Now()
	// 	startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -2).Format("2006-01-02"))
	// 	endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
	// }
	// m := service.FindByDate1(startDate, endDate, "date", "date")
	// if channelId != "0" {
	// 	m["paychannel"] = channelId
	// }
	// paychannel, _ := strconv.Atoi(channelId)
	// // count, _ := service.PayService.GetPayChannelStatTotal(m, true)
	// // list, _ := service.PayService.GetPayChannelStats(page, this.pageSize, m, true)
	// list, _ := service.PayService.GetPayChannelStatsNew(startDate, endDate, paychannel)
	// count := len(list)

	channellist, _ := service.GameService.GetPayChannel()
	newSlice := []entity.PayChannel{
		{Id: "0", Name: "全部"},
	}
	channellist = append(newSlice, channellist...)

	this.Data["pageTitle"] = "支付渠道统计"
	this.Data["typeId"] = typeId
	this.Data["aliasId"] = aliasId
	this.Data["classId"] = classId
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["channelId"] = channelId
	this.Data["channelList"] = channellist
	this.Data["money_statr"] = money_statr
	this.Data["money_end"] = money_end
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "paychannelstatsopt")
	// this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("PayController.PayChannelStats", "start_date", startDate, "end_date", endDate, "channel_id", channelId), true).ToString()
	this.display()
}

// 支付渠道统计
func (this *PayController) PayChannelStatsExport() {
	// page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	channelId := this.GetString("channel_id")
	typeId, _ := this.GetInt("typeId")
	aliasId := this.GetString("alias_id")
	classId := this.GetString("class_id")
	// money_statr := this.GetString("money_statr")
	// money_end := this.GetString("money_end")

	// var params = make(map[string]any)
	if startDate == "" || endDate == "" {
		// 默认近七天数据
		today := bson.Now()
		endDate = today.Format("2006-01-02")
		startDate = today.AddDate(0, 0, -6).Format("2006-01-02")
	}
	startTime, endTime := service.FindByDate4(startDate, endDate)
	m := service.FindByDate1(startDate, endDate, "date", "date")
	// 支付渠道
	var payChannels []string
	if channelId != "" && channelId != "0" && channelId != "-" {
		palkage_arr := strings.Split(channelId, ",")
		for _, item := range palkage_arr {
			if item != "" && item != "0" && item != "-" {
				payChannels = append(payChannels, item)
			}
		}
	}
	if len(payChannels) > 0 {
		var chs []int
		for _, ch := range payChannels {
			ch2, _ := strconv.Atoi(ch)
			chs = append(chs, ch2)
		}
		m["channel"] = bson.M{"$in": chs}
	}

	var packageIds []string
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
	if len(packageIds) > 0 {
		m["package_id"] = bson.M{"$in": packageIds}
	}

	var headers []string
	var rows [][]any
	if typeId == 0 {
		// 收付汇总
		list, err := service.PayService.PayChannelCollect(*startTime, *endTime, packageIds, payChannels)
		if err != nil {
			beego.Error("PayChannelCollect error: ", err)
		}
		headers = []string{"日期", "渠道类", "渠道别名", "支付渠道", "代收金额（玩家实付）", "代收税费", "代付金额（渠道实付）", "代付税费（按率）", "代付税费（按单）", "理论实收入", "理论实付出", "理论净收益", "渠道实收入（已扣税）", "渠道实付出（已加税）", "渠道收-支", "理论与实际的差额", "渠道实时余额（接口值）"}
		for _, v := range list {
			row := []any{
				v.Date,
				v.ChannelClass,
				v.Channel1,
				v.PayChannel,
				v.PayAmount,
				v.PayAmountTax,
				v.WithdrawAmount,
				v.WithdrawAmountTaxRate,
				v.WithdrawAmountTaxTimes,
				v.Earning,
				v.Expense,
				v.Profit,
				v.EarningActual,
				v.ExpenseActual,
				v.ProfitActual,
				v.ProfitDiff,
				v.PayChannelBalance,
			}
			rows = append(rows, row)
		}
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
	title := "支付渠道统计_"
	switch typeId {
	case 0:
		title += "收付汇总"
	case 2:
		title += "提现汇总"
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

// 支付渠道统计
func (this *PayController) PayChannelStatsBak() {
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	channelId := this.GetString("channel_id")
	if page < 1 {
		page = 1
	}
	if startDate == "" && endDate == "" {
		// 默认近3天数据
		today := bson.Now()
		startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -2).Format("2006-01-02"))
		endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
	}
	m := service.FindByDate1(startDate, endDate, "date", "date")
	if channelId != "0" {
		m["paychannel"] = channelId
	}
	paychannel, _ := strconv.Atoi(channelId)
	// count, _ := service.PayService.GetPayChannelStatTotal(m, true)
	// list, _ := service.PayService.GetPayChannelStats(page, this.pageSize, m, true)
	list, _ := service.PayService.GetPayChannelStatsNew(startDate, endDate, paychannel)
	count := len(list)
	channellist, _ := service.GameService.GetPayChannel()
	newSlice := []entity.PayChannel{
		{Id: "0", Name: "全部"},
	}
	channellist = append(newSlice, channellist...)

	this.Data["pageTitle"] = "支付渠道统计"
	this.Data["list"] = list
	this.Data["count"] = count
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["channelId"] = channelId
	this.Data["channelList"] = channellist
	// this.Data["reccount"] = 0
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "paychannelstatsopt")
	// this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("PayController.PayChannelStats", "start_date", startDate, "end_date", endDate, "channel_id", channelId), true).ToString()
	this.display()
}
func (this *PayController) AddPayStatsRemark() {
	id := this.GetString("id")
	if id == "" {
		this.checkError(errors.New("ID不能为空！"))
	}
	channelid, _ := this.GetInt("channel")
	if this.isPost() {
		// 提交

		remark := this.GetString("remark")
		if remark == "" {
			this.checkError(errors.New("备注不能为空不能为空！"))
		}
		// if channelid == 0 {

		// }
		name := this.auth.GetUser().UserName
		num, _ := strconv.ParseInt(id, 10, 64)
		info := new(entity.PayChannelStatRecords)
		info.Id = bson.NewObjectId().Hex()
		info.Date = int64(num)
		info.PayChannel = uint32(channelid)
		info.Remark = remark
		info.OptName = name
		info.Ctime = bson.Now().Unix()
		err := service.PayService.AddPayChannelStatRecords(info)
		if err != nil {
			this.checkError(err)
		}
		service.ActionService.Add("add_pay_channel_stats_remark", name,
			utils.String(info.Date), utils.String(info.PayChannel), utils.String(info.Remark), "")

		this.redirect(beego.URLFor("PayController.PayChannelStats"))
	} else {
		// 查询批注记录
		var buf bytes.Buffer
		num, _ := strconv.ParseInt(id, 10, 64)
		m := bson.M{}
		m["date"] = num
		m["pay_channel"] = channelid
		list, _ := service.PayService.GetPayChannelStatRecordsList(1, -1, m)
		count := len(list)
		// this.Data["reclist"] = list
		// this.Data["reccount"] = count
		if count > 0 {
			buf.WriteString("<div class=\"table-panel\">")
			buf.WriteString("<p>备注记录</p>")
			buf.WriteString("<table class=\"table table-striped table-bordered table-hover\">")
			buf.WriteString("<tr>")
			buf.WriteString("<th>支付渠道</th>")
			buf.WriteString("<th>添加人</th>")
			buf.WriteString("<th>添加时间</th>")
			buf.WriteString("<th>备注</th>")
			buf.WriteString("</tr>")

			for _, v := range list {
				buf.WriteString("<tr>")
				buf.WriteString(fmt.Sprintf("<td>%s</td>", v.PayChannelName))
				buf.WriteString(fmt.Sprintf("<td>%s</td>", v.OptName))
				buf.WriteString(fmt.Sprintf("<td>%s</td>", v.CDate.Format("2006-01-02 15:04:05")))
				buf.WriteString(fmt.Sprintf("<td>%s</td>", v.Remark))
				buf.WriteString("</tr>")
			}

			buf.WriteString("</table>")
			buf.WriteString("</div>")
		}

		// 将查询结果渲染到模板中
		this.Ctx.WriteString(buf.String())
	}
}

// 支付渠道日志
func (this *PayController) PayChannelLog() {
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	channelId, _ := this.GetInt("channel_id")
	typeId, _ := this.GetInt("type_id")
	if page < 1 {
		page = 1
	}
	if startDate == "" && endDate == "" {
		// 默认近3天数据
		today := bson.Now()
		startDate = fmt.Sprintf("%s", today.AddDate(0, 0, -2).Format("2006-01-02"))
		endDate = fmt.Sprintf("%s", today.Format("2006-01-02"))
	}
	m := service.FindByDate3(startDate, endDate, "date", "date")
	if channelId != 0 {
		m["pay_channel"] = channelId
	}
	if typeId != 0 {
		m["p_type"] = typeId
	}
	count, _ := service.PayService.GetPayChannelLogTotal(m)
	list, _ := service.PayService.GetPayChannelLogList(page, this.pageSize, m)

	channellist, _ := service.GameService.GetPayChannel()
	newSlice := []entity.PayChannel{
		{Id: "0", Name: "全部"},
	}
	channellist = append(newSlice, channellist...)

	payTypeList := map[int]string{
		0: "全部",
		1: "代收",
		2: "代付",
	}

	this.Data["pageTitle"] = "支付渠道日志"
	this.Data["list"] = list
	this.Data["count"] = count
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["channelId"] = strconv.Itoa(channelId)
	this.Data["channelList"] = channellist
	this.Data["paytypeList"] = payTypeList
	this.Data["paytypeId"] = typeId
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("PayController.PayChannelLog", "start_date", startDate, "end_date", endDate, "channel_id", channelId, "type_id", typeId), true).ToString()
	this.display()
}

// 支付渠道统计通知tg
func (this *PayController) PayChannelStatNotice() {
	if this.isPost() {
		dayTicker, _ := this.GetInt("dayTicker")
		timeTicker, _ := this.GetInt("timeTicker")
		withdrawTicker, _ := this.GetInt("withdrawTicker")
		payUtrTicker, _ := this.GetInt("payUtrTicker")
		service.PayService.UpdatePayChannelStatNoticeConfig(int64(dayTicker), int64(timeTicker), int64(withdrawTicker), int64(payUtrTicker))

		this.redirect(beego.URLFor("PayController.PayChannelStatNotice"))
	}

	config := service.PayService.GetPayChannelStatNoticeConfig()

	this.Data["pageTitle"] = "支付渠道统计通知"
	this.Data["dayTicker"] = config.DayTicker
	this.Data["timeTicker"] = config.TimeTicker
	this.Data["withdrawTicker"] = config.WithdrawTicker
	this.Data["payUtrTicker"] = config.PayUtrTicker
	loc, _ := time.LoadLocation("Asia/Shanghai")
	dayNextTime := time.Unix(config.DayNextTime, 0).In(loc).Format(utils.FORMAT)
	timeNextTime := time.Unix(config.TimeNextTime, 0).In(loc).Format(utils.FORMAT)
	withdrawNextTime := time.Unix(config.WithdrawNextTime, 0).In(loc).Format(utils.FORMAT)
	payUtrNextTime := time.Unix(config.PayUtrNextTime, 0).In(loc).Format(utils.FORMAT)
	this.Data["dayNextTime"] = dayNextTime
	this.Data["timeNextTime"] = timeNextTime
	this.Data["withdrawNextTime"] = withdrawNextTime
	this.Data["payUtrNextTime"] = payUtrNextTime
	this.display()
}
