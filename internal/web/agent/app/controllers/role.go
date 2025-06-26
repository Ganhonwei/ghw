package controllers

import (
	"strings"

	"goserver/internal/web/agent/app/entity"
	"goserver/internal/web/agent/app/service"
	"goserver/pkg/utils"

	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"
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
		perms := this.GetStrings("channels")

		err = service.RoleService.SetRoleChannel(role.Id, perms)
		if err == nil {
			result := strings.Join(perms, ",")
			service.ActionService.Add("set_role_channel", this.auth.GetUser().UserName, "", result, result)
			// service.ActionService.Add(this.auth.GetUser().UserName, role.RoleName)
			service.ActionService.PermRole(this.auth.GetUser().UserName, role.RoleName)
		}
		this.checkError(err)
		this.redirect(beego.URLFor("RoleController.List"))
		// pids := this.GetStrings("pids")
		// perms := this.GetStrings("perms")
		// if len(pids) == 0 {
		// 	role.ProjectIds = ""
		// } else {
		// 	role.ProjectIds = strings.Join(pids, ",")
		// }
		// newPerms := make([]string, 0)
		// for _, item := range perms {
		// 	suffix := "opt"
		// 	if strings.HasSuffix(item, suffix) {
		// 		newPerms = append(newPerms, item)
		// 		// 如果是黑白名单，默认权限 player.blacklistopt
		// 		if item == "player.blacklistopt" {
		// 			newPerms = append(newPerms, "player.blacklistadd")
		// 			newPerms = append(newPerms, "player.blacklistdel")
		// 		}
		// 		// 如果是opt结尾，根据名称查询对应的操作
		// 		list := service.SystemService.GetPermChild(item)
		// 		for _, v1 := range list {
		// 			newPerms = append(newPerms, v1.Id)
		// 		}
		// 	} else {
		// 		newPerms = append(newPerms, item)
		// 	}
		// }
		// fileds := bson.M{"ProjectIds": role.ProjectIds}
		// err := service.RoleService.UpdateRole(role, fileds)
		// this.checkError(err)
		// err = service.RoleService.SetPerm(role.Id, newPerms)
		// if err == nil {
		// 	service.ActionService.PermRole(this.auth.GetUser().UserName, role.RoleName)
		// }
		// this.checkError(err)
		// this.redirect(beego.URLFor("RoleController.List"))
	}

	//projectList, _ := service.ProjectService.GetAllProject()
	channelList := service.ChannelService.GetChannel()
	chkmap := make(map[string]string)
	for _, v := range role.ChannelList {
		chkmap[v.Name] = "checked"
	}
	// permList := service.SystemService.GetAgentList(false)

	// chkmap := make(map[string]string)
	// for _, v := range role.PermList {
	// 	chkmap[v.Key] = "checked"
	// }
	// if role.ProjectIds != "" {
	// 	pids := strings.Split(role.ProjectIds, ",")
	// 	for _, pid := range pids {
	// 		chkmap[pid] = "checked"
	// 	}
	// }

	this.Data["pageTitle"] = "编辑权限"
	this.Data["permList"] = channelList
	//this.Data["projectList"] = projectList
	this.Data["role"] = role
	this.Data["chkmap"] = chkmap
	this.display()
}

// 前一日统计
func (this *RoleController) TJ() {
	go func() {
		beego.Info("执行前三十天~")
		for i := 1; i <= 30; i++ {
			dateTime := utils.TimestampTodayTime(service.Location()).AddDate(0, 0, -i)
			datestamp := utils.Time2Stamp(dateTime)
			// 数据汇总统计
			service.DataStatistics(datestamp, "")
			// 数据汇总统计
			service.DataStatistics(datestamp, "")

			// 分享数据统计
			service.ShareStatistics(datestamp, "")

			// 用户留存统计
			service.UserRetained(datestamp, "")

			// 付费用户留存统计
			service.PayUserRetained(datestamp, "")

			// 行为分析
			service.PointData(datestamp, "")
			// 局数分析
			service.GameNumberAnalysisData(datestamp, "")
			// 时间分析
			service.PlaytimeAnalysisData(datestamp, "")
		}
		beego.Info("执行完成~")
	}()
	this.redirect(beego.URLFor("RoleController.List"))
	this.display()
}
