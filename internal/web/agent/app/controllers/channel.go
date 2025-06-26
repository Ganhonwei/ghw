package controllers

import (
	"bytes"
	"errors"
	"fmt"
	"goserver/gen/pb"
	"goserver/internal/web/agent/app/entity"
	"goserver/internal/web/agent/app/libs"
	"goserver/internal/web/agent/app/service"
	"goserver/pkg/utils"

	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"
)

type ChannelController struct {
	BaseController
}

// 渠道数据列表
func (this *ChannelController) List() {
	page, _ := this.GetInt("page")
	name := this.GetString("name")
	if page < 1 {
		page = 1
	}
	m := bson.M{}
	if name != "" {
		m["name"] = bson.M{"$regex": name, "$options": "i"}
	}
	// m = service.FindByDate1(startDate, endDate, "date", "date")
	list, _ := service.ChannelService.GetChannelList(page, this.pageSize, m)
	count, _ := service.ChannelService.GetChannelTotal(m)
	this.Data["pageTitle"] = "渠道设置"
	this.Data["list"] = list
	this.Data["count"] = count
	this.Data["name"] = name
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("ChannelController.List", "name", name), true).ToString()
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "listopt")
	this.display()
}

// 添加渠道数据
func (this *ChannelController) AddChannel() {
	cid := this.GetString("cid")
	name := this.GetString("cname")
	aname := this.GetString("aname")
	skey := this.GetString("secret_key")
	status, _ := this.GetInt("status")
	if name == "" {
		this.checkError(errors.New("渠道名称不能为空！"))
	}
	if this.isPost() {
		// 提交

		temp, _ := service.ChannelService.GetChannelById(name)
		if temp != nil {
			if temp.Id != "" {
				this.checkError(errors.New("已存在该渠道名称，不可重复添加！"))
			}
		}

		info := new(entity.ChannelInfo)
		info.Id = cid
		info.Name = name
		info.Name1 = aname
		info.SecretKey = skey
		info.Status = status
		info.Operator = this.auth.GetUser().UserName
		// 新增
		err := service.ChannelService.AddOrUpdateChannel(info)
		if err != nil {
			this.checkError(err)
		} else {
			msg := make(map[string]entity.ChannelInfo)
			msg[cid] = *info
			result, err1 := service.GmRequest(pb.WebChannel, pb.CONFIG_UPSERT, msg)
			beego.Trace("result: ", result)
			if err1 != nil {
				this.checkError(err1)
			}
			service.ActionService.Add("add_or_update_channel", this.auth.GetUser().UserName, "", info.Id, info.Id)
			this.showMsg("操作成功！", MSG_OK, beego.URLFor("ChannelController.List"))
		}
	}
}

// 删除渠道数据
func (this *ChannelController) DelChannel() {
	id := this.GetString("id")

	if id == "" {
		this.checkError(errors.New("无法操作该渠道数据!"))
	}
	info, _ := service.ChannelService.GetChannelById(id)
	err := service.ChannelService.DelChannel(id)
	if err != nil {
		this.checkError(err)
	} else {
		msg := make(map[string]entity.ChannelInfo)
		msg[id] = *info
		result, err1 := service.GmRequest(pb.WebChannel, pb.CONFIG_DELETE, msg)
		beego.Trace("result: ", result)
		if err1 != nil {
			this.checkError(err1)
		}
		this.showMsg("删除成功！", MSG_OK, beego.URLFor("ChannelController.List"))
		service.ActionService.Add("del_channel", this.auth.GetUser().UserName, "", id, id)
	}
}

/*
博主账号设置相关接口
*/
func (this *ChannelController) BloggerAccount() {
	page, _ := this.GetInt("page")
	userid := this.GetString("userid")
	typeId, _ := this.GetInt("type_id")
	if page < 1 {
		page = 1
	}
	m := bson.M{}
	if userid != "" {
		if typeId == 0 {
			m["_id"] = userid
		} else if typeId == 1 {
			m["phone"] = userid
		} else if typeId == 2 {
			m["nickname"] = userid
		} else if typeId == 3 {
			m["login_ip"] = userid
		} else if typeId == 4 {
			m["device_id"] = userid
		}
	}
	if userid == "" {
		strArr, _ := service.ChannelService.GetBloggerId()
		if len(strArr) > 0 {
			m["_id"] = bson.M{"$in": strArr}
		} else {
			m["_id"] = ""
		}
	}
	m["robot"] = false
	m["simulation_robot"] = false
	list, _ := service.PlayerService.GetList(page, this.pageSize, m)
	if len(list) > 0 {
		for i, item := range list {
			info, _ := service.ChannelService.GetBloggerById(item.Userid)
			if info != nil {
				// item.IsBlogger = true
				list[i].IsBlogger = true
			} else {
				list[i].IsBlogger = false
				// item.IsBlogger = false
			}
		}
	}
	count, _ := service.PlayerService.GetTotal(m)
	typeList := map[int]string{
		0: "用户ID",
		1: "手机号",
		2: "用户昵称",
		3: "IP",
		4: "设备号",
	}
	this.Data["pageTitle"] = "博主账号"
	this.Data["list"] = list
	this.Data["count"] = count
	this.Data["typeList"] = typeList
	this.Data["typeId"] = typeId
	this.Data["userid"] = userid
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("ChannelController.BloggerAccount", "userid", userid, "type_id", typeId), true).ToString()
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "bloggeropt")
	this.display()
}

// 博主账号下级详情
func (this *ChannelController) BloggerDetails() {
	id := this.GetString("id")
	// page, _ := this.GetInt("page")
	// if page < 1 {
	// 	page = 1
	// }
	var buf bytes.Buffer
	info, _ := service.PlayerService.GetUser(id)
	// count := 0
	buf.WriteString("<div class=\"table-panel\">")
	buf.WriteString("<table class=\"table table-striped table-bordered table-hover\">")
	buf.WriteString("<tr>")
	buf.WriteString("<th>用户ID</th>")
	buf.WriteString("<th>昵称</th>")
	buf.WriteString("<th>充值金额</th>")
	buf.WriteString("<th>上次登录时间</th>")
	buf.WriteString("<th>总打码量</th>")
	buf.WriteString("<th>是否领取奖励</th>")
	buf.WriteString("</tr>")
	if info != nil {
		if len(info.ShareBelow) > 0 {
			// count = len(info.ShareBelow)
			for _, item := range info.ShareBelow {
				amount := service.Chip2Float(item.Recharge)
				c, _ := service.ConvertToIndiaTime(item.LastLogin)
				lastTime := c.Format("2006-01-02 15:04:05")
				strName := ""
				if item.Receive {
					strName = "已领取"
				} else {
					strName = "未领取"
				}
				buf.WriteString("<tr>")
				buf.WriteString(fmt.Sprintf("<td>%s</td>", item.UserId))
				buf.WriteString(fmt.Sprintf("<td>%s</td>", item.Name))
				buf.WriteString(fmt.Sprintf("<td>%.2f</td>", amount))
				buf.WriteString(fmt.Sprintf("<td>%s</td>", lastTime))
				buf.WriteString(fmt.Sprintf("<td>%d</td>", item.BetAmount))
				buf.WriteString(fmt.Sprintf("<td>%s</td>", strName))
				buf.WriteString("</tr>")
			}
		} else {
			buf.WriteString("<tr><td colspan=\"20\">暂无记录...</td></tr>")
		}
	}
	buf.WriteString("</table>")
	buf.WriteString("</div>")
	// if count > 10 {
	// 	// 分页
	// 	/*
	// 			<div class="row-page">
	// 			{{if gt .count 20}}
	// 			{{str2html .pageBar}}
	// 		{{end}}
	// 		</div>
	// 	*/
	// 	buf.WriteString("<div class=\"row-page\">")
	// 	strPage := libs.NewPager(page, int(count), 10, "", true).ToString()
	// 	buf.WriteString(strPage)
	// 	buf.WriteString("</div>")
	// }
	// 将查询结果渲染到模板中
	this.Ctx.WriteString(buf.String())
}

// 设置博主账号
func (this *ChannelController) BloggerSet() {
	id := this.GetString("id")
	// name := this.GetString("name")
	if this.isPost() {
		if id == "" {
			this.checkError(errors.New("用户ID不能为空!"))
		}
		name := ""
		userinfo, _ := service.PlayerService.GetUser(id)
		if userinfo != nil {
			name = userinfo.AD_BundleId
		}
		if name == "" {
			this.checkError(errors.New("博主渠道不能为空!"))
		}
		info := new(entity.BloggerAccount)
		info.Userid = id
		info.Channel = name
		info.Operator = this.auth.GetUser().UserName
		err := service.ChannelService.AddBlogger(info)
		if err != nil {
			this.checkError(err)
		}
		go func() {
			acc_id := id
			for i := 0; i <= 30; i++ {
				dateTime := utils.TimestampTodayTime(service.Location()).AddDate(0, 0, -i)
				datestamp := utils.Time2Stamp(dateTime)
				// 数据汇总统计
				service.DataStatistics(datestamp, acc_id)
				// 数据汇总统计
				service.DataStatistics(datestamp, acc_id)

				// 分享数据统计
				service.ShareStatistics(datestamp, acc_id)

				// 用户留存统计
				service.UserRetained(datestamp, acc_id)

				// 付费用户留存统计
				service.PayUserRetained(datestamp, acc_id)

				// 行为分析
				service.PointData(datestamp, acc_id)
				// 局数分析
				service.GameNumberAnalysisData(datestamp, acc_id)
				// 时间分析
				service.PlaytimeAnalysisData(datestamp, acc_id)
			}
		}()
		this.showMsg("设置成功！", MSG_OK, beego.URLFor("ChannelController.BloggerAccount"))
		service.ActionService.Add("set_blogger_account", this.auth.GetUser().UserName, "", id, id)
	}
}

func (this *ChannelController) BloggerCancel() {
	id := this.GetString("id")
	if this.isPost() {
		if id == "" {
			this.checkError(errors.New("用户ID不能为空!"))
		}
		err := service.ChannelService.DelBlogger(id)
		if err != nil {
			this.checkError(err)
		}
		this.showMsg("取消成功！", MSG_OK, beego.URLFor("ChannelController.BloggerAccount"))
		service.ActionService.Add("cancel_blogger_account", this.auth.GetUser().UserName, "", id, id)
	}
}
