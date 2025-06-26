package controllers

import (
	"goserver/internal/web/admin/app/libs"
	"goserver/internal/web/admin/app/service"
	"strconv"

	"github.com/astaxie/beego"
)

type ActionController struct {
	BaseController
}

func (this *ActionController) List() {
	page, _ := strconv.Atoi(this.GetString("page"))
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	userid := this.GetString("userid")
	if page < 1 {
		page = 1
	}
	m := service.FindByDate(startDate, endDate, "create_time", "create_time")
	if userid != "" {
		m["actor"] = userid
	}
	count, _ := service.ActionService.GetTotal(m)
	users, _ := service.ActionService.GetList(page, this.pageSize, m)

	this.Data["pageTitle"] = "管理员操作日志"
	this.Data["count"] = count
	this.Data["list"] = users
	this.Data["userid"] = userid
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("ActionController.List", "userid", userid, "start_date", startDate, "end_date", endDate), true).ToString()
	this.display()
}
