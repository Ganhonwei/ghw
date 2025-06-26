package controllers

import (
	"strings"

	"goserver/internal/web/admin/app/entity"
	"goserver/internal/web/admin/app/service"

	"github.com/astaxie/beego"
	"github.com/globalsign/mgo/bson"
)

type RoleController struct {
	BaseController
}

func (this *RoleController) List() {
	roleList, err := service.RoleService.GetAllRoles()
	this.checkError(err)
	for k, role := range roleList {
		roleList[k].UserList, _ = service.UserService.GetUserListByRoleId(role.Id)
	}
	this.Data["pageTitle"] = "角色管理"
	this.Data["list"] = roleList
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "listopt")
	this.display()
}

func (this *RoleController) Add() {
	if this.isPost() {
		role := &entity.Role{}
		role.RoleName = this.GetString("role_name")
		role.Description = this.GetString("description")
		if role.RoleName == "" {
			this.showMsg("角色名不能为空", MSG_ERR)
		}
		err := service.RoleService.AddRole(role)
		if err == nil {
			service.ActionService.AddRole(this.auth.GetUser().UserName, role.RoleName)
		}
		this.checkError(err)
		this.redirect(beego.URLFor("RoleController.List"))
	}
	this.Data["pageTitle"] = "创建角色"
	this.display()
}

func (this *RoleController) Edit() {
	id := this.GetString("id")
	role, err := service.RoleService.GetRole(id)
	this.checkError(err)

	if this.isPost() {
		role.RoleName = this.GetString("role_name")
		role.Description = this.GetString("description")
		fileds := bson.M{"RoleName": role.RoleName,
			"Description": role.Description}
		err := service.RoleService.UpdateRole(role, fileds)
		if err == nil {
			service.ActionService.UpdateRole(this.auth.GetUser().UserName, role.RoleName)
		}
		this.checkError(err)
		this.redirect(beego.URLFor("RoleController.List"))
	}

	this.Data["pageTitle"] = "编辑角色"
	this.Data["role"] = role
	this.display()
}

func (this *RoleController) Del() {
	id := this.GetString("id")

	err := service.RoleService.DeleteRole(id)
	if err == nil {
		service.ActionService.DelRole(this.auth.GetUser().UserName, id)
	}
	this.checkError(err)

	this.redirect(beego.URLFor("RoleController.List"))
}

func (this *RoleController) Perm() {
	id := this.GetString("id")
	role, err := service.RoleService.GetRole(id)
	this.checkError(err)

	if this.isPost() {
		pids := this.GetStrings("pids")
		perms := this.GetStrings("perms")
		if len(pids) == 0 {
			role.ProjectIds = ""
		} else {
			role.ProjectIds = strings.Join(pids, ",")
		}
		newPerms := make([]string, 0)
		newPerms = append(newPerms, "main.setpagesize")
		for _, item := range perms {
			suffix := "opt"
			if strings.HasSuffix(item, suffix) {
				newPerms = append(newPerms, item)
				// 如果是黑白名单，默认权限 player.blacklistopt
				if item == "player.blacklistopt" {
					newPerms = append(newPerms, "player.blacklistadd")
					newPerms = append(newPerms, "player.blacklistdel")
				}
				// 如果是opt结尾，根据名称查询对应的操作
				list := service.SystemService.GetPermChild(item)
				for _, v1 := range list {
					newPerms = append(newPerms, v1.Id)
				}
			} else {
				newPerms = append(newPerms, item)
			}
		}
		fileds := bson.M{"ProjectIds": role.ProjectIds}
		err := service.RoleService.UpdateRole(role, fileds)
		this.checkError(err)
		err = service.RoleService.SetPerm(role.Id, newPerms)
		if err == nil {
			service.ActionService.PermRole(this.auth.GetUser().UserName, role.RoleName)
		}
		this.checkError(err)
		this.redirect(beego.URLFor("RoleController.List"))
	}

	//projectList, _ := service.ProjectService.GetAllProject()
	permList := service.SystemService.GetPermList(false)

	chkmap := make(map[string]string)
	for _, v := range role.PermList {
		chkmap[v.Key] = "checked"
	}
	if role.ProjectIds != "" {
		pids := strings.Split(role.ProjectIds, ",")
		for _, pid := range pids {
			chkmap[pid] = "checked"
		}
	}

	this.Data["pageTitle"] = "编辑权限"
	this.Data["permList"] = permList
	//this.Data["projectList"] = projectList
	this.Data["role"] = role
	this.Data["chkmap"] = chkmap
	this.display()
}

// 前一日统计
func (this *RoleController) TJ() {
	go func() {
		beego.Info("执行用户充值汇总~")
		service.ComputePayAndWithdrawByAll()
		// 数据汇总统计
		// service.DataStatistics(utils.TimestampYesterday())

		// // 数据汇总(旧)
		// service.DataSummary(utils.TimestampYesterday())

		// // 渠道数据(旧)
		// service.ChannelData(utils.TimestampYesterday())

		// service.RealTimeData(utils.TimestampYesterday())

		// service.BasicPayActivity(utils.TimestampYesterday())

		// service.BasicFreeActivity(utils.TimestampYesterday())

		// service.UserRetained(utils.TimestampYesterday())

		// service.UserRetainedByChannel(utils.TimestampYesterday())

		// service.UserResource(utils.TimestampYesterday())

		// service.GlobalResource(utils.TimestampYesterday())

		// service.GameResource(utils.TimestampYesterday())

		// service.PayChannelRate(utils.TimestampYesterday())
		// // service.ChannelSuccessRate(utils.TimestampYesterday())

		// // service.ChannelSuccessRate1(utils.TimestampYesterday())

		// // 房间数据
		// service.RoomData(utils.TimestampYesterday())

		// service.CheckIPDuplication(utils.TimestampYesterday())

		// service.PayUserRetained(utils.TimestampYesterday())

		// service.PayUserRetainedByChannel(utils.TimestampYesterday())

		// // 商品购买统计
		// service.GoodsBuyData(utils.TimestampYesterday())

		// service.AdReportStatistics(utils.TimestampYesterday())

		// // 充值来源
		// service.PaySourceData(utils.TimestampYesterday())
		beego.Info("执行用户充值汇总完成~")
	}()
	this.redirect(beego.URLFor("RoleController.List"))
	this.display()
}
