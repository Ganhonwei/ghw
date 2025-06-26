package controllers

import (
	"errors"
	"fmt"
	"goserver/gen/pb"
	"goserver/internal/web/admin/app/args"
	"goserver/internal/web/admin/app/entity"
	"goserver/internal/web/admin/app/libs"
	"goserver/internal/web/admin/app/service"
	"goserver/pkg/utils"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/validation"
	"github.com/globalsign/mgo/bson"
)

type MessageController struct {
	BaseController
}

// 获取未读消息数量
func (this *MessageController) GetUnreadMsg() {
	coount := service.MessageService.GetUnreadMessage()
	strCount := fmt.Sprintf("%d", coount)
	this.showMsg(strCount, MSG_OK)
}

// 客服消息
func (this *MessageController) _Customer() {
	statusid, _ := this.GetInt("status_id")
	page, _ := this.GetInt("page")
	userid := this.GetString("userid")
	tabid, _ := this.GetInt("tabid")
	if page < 1 {
		page = 1
	}
	m := service.FindByDate("", "", "utime", "utime")
	if tabid == 1 {
		// 默认一周数据
		today := bson.Now()
		startDate := fmt.Sprintf("%s", today.AddDate(0, 0, -6).Format("2006-01-02"))
		endDate := fmt.Sprintf("%s", today.Format("2006-01-02"))
		m = service.FindByDate(startDate, endDate, "utime", "utime")
		name := this.auth.GetUser().UserName
		m["log.name"] = bson.M{"$regex": name, "$options": "i"}
		//bson.M{"$exists": true}
	} else if tabid == 0 {
		// 分配给客服的消息
		cuserid := this.auth.GetUser().Id
		m["assign_user"] = cuserid
	}

	if statusid != 0 {
		m["reply_status"] = statusid
	}
	if userid != "" {
		m["_id"] = userid
	}
	if tabid == 3 || tabid == 4 {
		// 头像
		m1 := bson.M{}
		if userid != "" {
			m1["user_id"] = userid
		}
		if tabid == 3 {
			m1["is_audit"] = 0
		}
		list, _ := service.MessageService.GetUploadHeadList(page, this.pageSize, m1)
		count, _ := service.MessageService.GetUploadHeadTotal(m1)
		this.Data["count"] = count
		this.Data["list"] = list
		this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("MessageController.Customer", "userid", userid, "tabid", tabid, "status_id", statusid), true).ToString()
	} else {
		list, _ := service.MessageService.CustomerList(page, this.pageSize, m)
		count, _ := service.MessageService.GetCustomerTotal(m)
		ids := make([]string, 0)
		for _, v := range list {
			ids = append(ids, v.Userid)
		}
		userList, _ := service.PlayerService.GetByUser(ids)
		for i := range list {
			for j := range userList {
				var user entity.PlayerUser
				if list[i].Userid == userList[j].Userid {
					user = userList[j]
					list[i].Nickname = user.Nickname
					list[i].BundleId = user.AD_BundleId
					list[i].OnlineStatus = user.OnlineStatus
					break
				}
			}
			num := len(list[i].Logs)
			last := list[i].Logs[num-1]
			list[i].LastTime = last.Ctime
			list[i].LastMsg = last.Content
		}
		this.Data["count"] = count
		this.Data["list"] = list
		this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("MessageController.Customer", "userid", userid, "tabid", tabid, "status_id", statusid), true).ToString()
	}

	statusList := map[int]string{
		0: "全部",
		1: "未查看",
		2: "未回复",
		3: "已回复",
	}
	this.Data["pageTitle"] = "客服消息"
	this.Data["statusList"] = statusList
	this.Data["statusId"] = statusid

	this.Data["userid"] = userid
	this.Data["tabid"] = tabid
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "customeropt")
	this.display()
}

// 新增聊天
func (this *MessageController) CustomerAdd() {
	if this.isPost() {
		userid := this.GetString("userid")
		if userid == "" {
			this.checkError(fmt.Errorf("用户ID不能为空"))
		}
		info, err := service.PlayerService.GetUser(userid)
		if err == nil {
			showModal := true
			this.redirect(beego.URLFor("MessageController.CustomerReply", "userid", info.Userid, "ShowModal", showModal))

			// this.Data["userid"] = info.Userid
			// this.Data["ShowModal"] = showModal

		} else {
			this.checkError(err)
		}
	}

	this.Data["pageTitle"] = "新增聊天"
	// this.TplName = "customer.html"
	this.display()
}

// 沟通
func (this *MessageController) _CustomerReply() {
	userid := this.GetString("userid")
	if userid == "" {
		this.checkError(fmt.Errorf("用户ID不能为空"))
	}
	if this.isPost() {
		// 客服发送消息
		content := this.GetString("content")
		if content == "" {
			this.checkError(fmt.Errorf("发送内容不能为空"))
		}

		msg, err := service.MessageService.GetFeedBackLog(userid)
		chatinfo := new(entity.ChatLog)
		chatinfo.Receiver = userid
		chatinfo.Content = content
		chatinfo.Ctime = bson.Now()
		chatinfo.Sender = "-1"
		chatinfo.Name = this.auth.GetUser().UserName

		chatinfo.Title = "Customer service message"
		if err != nil {
			//不存在数据
			msg = new(entity.FeedBackLog)
			chats := []entity.ChatLog{
				*chatinfo,
			}
			msg.ReplyStatus = 3
			msg.Userid = userid
			msg.Utime = bson.Now()
			msg.Logs = chats
			err := service.MessageService.AddFeedBackLog(msg)
			this.checkError(err)
		} else {
			msg.ReplyStatus = 3
			msg.Logs = append(msg.Logs, *chatinfo)
			err := service.MessageService.UpdateFeedBackLog(msg)
			this.checkError(err)
		}
		// 通知服务器
		result, err := service.GmRequest(pb.WebFeedBack, pb.CONFIG_UPSERT, chatinfo)
		beego.Trace("result: ", result)
		if err != nil {
			this.checkError(err)
		} else {
			service.ActionService.Add("send_feed_back", this.auth.GetUser().UserName,
				"", utils.String(userid), utils.String(userid), "")
		}
		this.redirect(beego.URLFor("MessageController.Customer"))
	} else {
		info, _ := service.PlayerService.GetUser(userid)
		msg, err := service.MessageService.GetFeedBackLog(userid)
		if err != nil {
			msg = new(entity.FeedBackLog)
			if info != nil {
				msg.Nickname = info.Nickname
				msg.BundleId = info.AD_BundleId
				msg.OnlineStatus = info.OnlineStatus
			}
		} else {
			if info != nil {
				msg.Nickname = info.Nickname
				msg.BundleId = info.AD_BundleId
				msg.OnlineStatus = info.OnlineStatus
			}
			le := len(msg.Logs)
			if le > 0 {
				last := msg.Logs[le-1]
				msg.LastTime = last.Ctime
				msg.LastMsg = last.Content
			}
			// 更新查看标记
			if msg.ReplyStatus == 1 {
				msg.ReplyStatus = 2
				service.MessageService.UpdateReplyStatus(msg)
			}
		}

		this.Data["info"] = msg
	}
	this.Data["pageTitle"] = "客服消息"
	this.Data["userid"] = userid
	this.display()
}

// 头像审核通过
func (this *MessageController) UploadHeadPass() {
	id := this.GetString("id")
	tabid, _ := this.GetInt("tabid")
	if id == "" {
		this.checkError(errors.New("上传头像ID不能为空"))
	}
	name := this.auth.GetUser().UserName
	msg := new(entity.UserCustomPhoto)
	msg.Id = id
	msg.Result = 1
	result, err := service.GmRequest(pb.WebUserPhoto, pb.CONFIG_UPSERT, msg)
	beego.Trace("result: ", result)
	if err != nil {
		this.checkError(err)
	}
	// 修改DB
	info := new(entity.UploadUserHead)
	info.Id = id
	info.IsAudit = 1
	info.ATime = bson.Now().Unix()
	info.Auditor = name
	err = service.MessageService.UpdateUserHead(info)
	if err != nil {
		this.checkError(err)
	}
	service.ActionService.Add("upload_head_pass", name,
		"", utils.String(id), utils.String(id), "")
	this.showMsg("操作成功！", MSG_OK, beego.URLFor("MessageController.Customer", "tabid", tabid))
}

// 头像审核拒绝
func (this *MessageController) UploadHeadRefuse() {
	id := this.GetString("id")
	tabid, _ := this.GetInt("tabid")
	if id == "" {
		this.checkError(errors.New("上传头像ID不能为空"))
	}
	name := this.auth.GetUser().UserName
	msg := new(entity.UserCustomPhoto)
	msg.Id = id
	msg.Result = 2
	result, err := service.GmRequest(pb.WebUserPhoto, pb.CONFIG_UPSERT, msg)
	beego.Trace("result: ", result)
	if err != nil {
		this.checkError(err)
	}
	// 修改DB
	info := new(entity.UploadUserHead)
	info.Id = id
	info.IsAudit = 2
	info.ATime = bson.Now().Unix()
	info.Auditor = name
	err = service.MessageService.UpdateUserHead(info)
	if err != nil {
		this.checkError(err)
	}
	service.ActionService.Add("upload_head_refuse", name,
		"", utils.String(id), utils.String(id), "")
	this.showMsg("操作成功！", MSG_OK, beego.URLFor("MessageController.Customer", "tabid", tabid))
}

// 公告广播
func (this *MessageController) NoticeList() {
	// status, _ := this.GetInt("status")
	page, _ := this.GetInt("page")
	tabid, _ := this.GetInt("tabid")
	if page < 1 {
		page = 1
	}
	typeList := map[int]string{
		0:  "无",
		1:  "TP",
		2:  "DRAGON TIGER",
		3:  "7UPDOWN",
		4:  "RUMMY",
		5:  "AK47",
		6:  "JOKER",
		7:  "CRASH",
		8:  "首充",
		9:  "入门礼包",
		10: "在线奖励",
		11: "签到",
		12: "任务",
		13: "牌型任务",
		14: "周卡",
		15: "分享",
		16: "包赔",
		17: "Bug有奖",
		18: "商城",
		19: "提现",
		20: "小米活动",
		21: "礼包码",
		22: "二充",
		23: "三充",
		24: "优惠券",
		99: "自定义",
	}
	m := service.FindByDate("", "", "ctime", "ctime")
	// 过期处理
	// if status == 0 { //未过期
	// 	m["etime"] = bson.M{"$gte": bson.Now()}
	// } else { //已过期
	// 	m["etime"] = bson.M{"$lt": bson.Now()}
	// }

	fmt.Println("tabid:", tabid)

	count := int64(0)
	if tabid == 2 {
		// Banner页
		list, _ := service.MessageService.GetBannerList(page, this.pageSize, m)
		count, _ = service.MessageService.GetBannerTotal(m)
		for i, item := range list {
			for k, t := range typeList {
				if k == item.Rtype {
					item.RtypeName = t
					list[i] = item
				}
			}
		}
		this.Data["list"] = list
	} else if tabid == 7 {
		fmt.Println("Bonus中心...", tabid)
		// Bonus中心
		list, _ := service.MessageService.GetBonusBannerList(page, this.pageSize, m)
		count, _ = service.MessageService.GetBonusBannerTotal(m)
		for i, item := range list {
			for k, t := range typeList {
				if k == item.Rtype {
					item.RtypeName = t
					list[i] = item
				}
			}
		}
		this.Data["list"] = list
	} else if tabid == 3 {
		// 联系方式
		list, _ := service.MessageService.GetContactWayList(page, this.pageSize, m)
		count, _ = service.MessageService.GetContactWayTotal(m)
		this.Data["list"] = list
	} else if tabid == 4 {
		// 客服功能
		list, _ := service.MessageService.GetCustomerUser()
		count = int64(len(list)) // service.MessageService.GetContactWayTotal(m)

		info, _ := service.MessageService.GetChatConfig()
		if info == nil {
			info = new(entity.CustomerServiceConfig)
		}
		this.Data["configInfo"] = info
		this.Data["list"] = list
	} else if tabid == 5 {
		// CDKEY
		list, _ := service.MessageService.GetCdkeyList(page, this.pageSize, m)
		count, _ = service.MessageService.GetCdkeyTotal(m)
		this.Data["list"] = list
	} else if tabid == 6 {
		// 分享配置
		list, _ := service.MessageService.GetShareWayList(page, this.pageSize, m)
		count = int64(len(list))
		this.Data["list"] = list
	} else {
		fmt.Println("Notice...", tabid)
		m["del"] = 0
		if tabid == 1 {
			// 公告消息
			m["rtype"] = bson.M{"$ne": 0}
		} else {
			// 跑马灯消息
			m["rtype"] = 0
		}
		list, _ := service.MessageService.GetNoticeList(page, this.pageSize, m)
		count, _ = service.MessageService.GetNoticeListTotal(m)
		this.Data["list"] = list
	}

	this.Data["imgtype"] = typeList

	this.Data["pageTitle"] = "运营功能"
	this.Data["tabid"] = tabid
	this.Data["count"] = count

	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("MessageController.NoticeList", "tabid", tabid), true).ToString()
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "noticelistopt")
	this.display()
}

// 新增消息
func (this *MessageController) NoticeAdd() {
	if this.isPost() {
		notice := new(entity.Notice)
		num, _ := this.GetInt("num")
		content := this.GetString("content")
		stime := this.GetString("start_date2")
		s := fmt.Sprintf("%s 00:00:00", stime)
		etime := this.GetString("end_date2")
		e := fmt.Sprintf("%s 23:59:59", etime)
		Stime := utils.Str2Time(s, service.Location())
		Etime := utils.Str2Time(e, service.Location())

		notice.Num = num
		notice.Content = content
		notice.Stime = Stime
		notice.Etime = Etime
		notice.Rtype = 0
		err := this.validNotice(notice)
		this.checkError(err)

		err = service.MessageService.AddOrUpdateNotice(notice)
		this.checkError(err)
		newId := *&notice.Id
		req := map[string]entity.Notice{
			newId: *notice,
		}
		result, err := service.GmRequest(pb.WebNotice, pb.CONFIG_UPSERT, req)
		beego.Trace("result: ", result)
		if err != nil {
			this.checkError(err)
		} else {
			service.ActionService.AddNotice(this.auth.GetUser().UserName, notice.Id)
		}
		this.redirect(beego.URLFor("MessageController.NoticeList", "tabid", 0))
	}

	this.Data["pageTitle"] = "新增消息"
	this.display()
}

func (this *MessageController) validNotice(notice *entity.Notice) error {
	valid := validation.Validation{}
	valid.Required(notice.Stime, "start_date").Message("开始时间格式不正确")
	valid.Required(notice.Etime, "end_date").Message("结束时间格式不正确")
	valid.Required(notice.Num, "num").Message("发送频率格式不正确")
	valid.Required(notice.Content, "content").Message("消息内容不能为空")
	if valid.HasErrors() {
		for _, err := range valid.Errors {
			return errors.New(err.Message)
		}
	}

	return nil
}

// 修改公告
func (this *MessageController) NoticeEdit() {
	id := this.GetString("id")
	if id == "" {
		this.checkError(fmt.Errorf("消息ID不能为空"))
	}
	info, _ := service.MessageService.GetNotice(id)
	if this.isPost() {
		num, _ := this.GetInt("num")
		content := this.GetString("content")
		stime := this.GetString("start_date2")
		etime := this.GetString("end_date2")
		var s string
		if len(stime) == 10 {
			s = fmt.Sprintf("%s 00:00:00", stime)
		} else {
			s = fmt.Sprintf("%s", stime)
		}
		var e string
		if len(etime) == 10 {
			e = fmt.Sprintf("%s 00:00:00", etime)
		} else {
			e = fmt.Sprintf("%s", etime)
		}

		Stime := utils.Str2Time(s, service.Location())
		Etime := utils.Str2Time(e, service.Location())

		info.Num = num
		info.Content = content
		info.Stime = Stime
		info.Etime = Etime
		info.Rtype = 0
		err := this.validNotice(info)
		this.checkError(err)

		err = service.MessageService.AddOrUpdateNotice(info)
		this.checkError(err)
		newId := *&info.Id
		req := map[string]entity.Notice{
			newId: *info,
		}
		result, err := service.GmRequest(pb.WebNotice, pb.CONFIG_UPSERT, req)
		beego.Trace("result: ", result)
		if err != nil {
			this.checkError(err)
		} else {
			service.ActionService.UpdateNotice(this.auth.GetUser().UserName, info.Id)
		}
		// if info.Etime.After(bson.Now()) {
		// 	status = 0
		// } else {
		// 	status = 1
		// }
		this.redirect(beego.URLFor("MessageController.NoticeList", "tabid", 0))
	}

	this.Data["pageTitle"] = "修改消息"
	this.Data["notice"] = info
	this.display()
}

// 新增、修改公告消息
func (this *MessageController) Notice() {
	id := this.GetString("id")
	info := new(entity.Notice)
	if id != "" {
		info, _ = service.MessageService.GetNotice(id)
	}
	if this.isPost() {
		msg_type, _ := this.GetInt("msg_type")
		language, _ := this.GetInt("language")
		content := this.GetString("content")
		if content == "" {
			this.checkError(errors.New("请输入公告内容！"))
		}
		info.Rtype = msg_type
		info.Language = language
		info.Content = content
		// 添加数据
		if info == nil || info.Id == "" {
			// 判断改语言是否存在数据，一个语言只能添加一条数据
			m := bson.M{}
			m["rtype"] = msg_type
			m["language"] = language
			count, _ := service.MessageService.GetNoticeListTotal(m)
			if count > 0 {
				// 已存在一条以上公告数据
				this.checkError(errors.New("该语言类型只能存在一条公告消息，请勿重复添加!"))
			}
		}
		err := service.MessageService.AddOrUpdateNotice(info)
		this.checkError(err)
		// 通知服务器
		newId := *&info.Id
		req := map[string]entity.Notice{
			newId: *info,
		}
		result, err := service.GmRequest(pb.WebNotice, pb.CONFIG_UPSERT, req)
		beego.Trace("result: ", result)
		this.checkError(err)

		service.ActionService.Notice(this.auth.GetUserName(), id)
		this.redirect(beego.URLFor("MessageController.NoticeList", "tabid", 1))
	}
}

// 移除公告广播
func (this *MessageController) NoticeDel() {
	id := this.GetString("id")
	notice, err := service.MessageService.GetNotice(id)
	this.checkError(err)

	notice.Del = 1

	err = service.MessageService.DelNotice(id)
	this.checkError(err)
	newId := *&notice.Id
	req := map[string]entity.Notice{
		newId: *notice,
	}
	result, err := service.GmRequest(pb.WebNotice, pb.CONFIG_DELETE, req)
	beego.Trace("result: ", result)
	if err != nil {
		this.checkError(err)
	} else {
		service.ActionService.DelNotice(this.auth.GetUserName(), id)
	}
	// if notice.Etime.After(bson.Now()) {
	// 	status = 0
	// } else {
	// 	status = 1
	// }
	this.redirect(beego.URLFor("MessageController.NoticeList", "tabid", 0))
}

func (this *MessageController) AddBanner() {
	id := this.GetString("bid")
	tabid, _ := this.GetInt("tabid")
	if this.isPost() {
		type_id, _ := this.GetInt("type_id")
		m := bson.M{}
		if tabid == 2 {
			count, _ := service.MessageService.GetBannerTotal(m)
			if count >= 6 {
				this.checkError(errors.New("最多添加六张Banner页图片！"))
				return
			}
		}

		f, h, err := this.GetFile("filename")
		ImgName := this.GetString("imgUrl")
		a_url := this.GetString("a_Url")
		fmt.Println("获取上传文件:", h)

		if err != nil {
			this.checkError(errors.New("请选择需要上传的图片"))
			return
		}
		// 延迟关闭文件
		defer f.Close()

		if tabid == 7 {
			// bonus中心
			info := new(entity.BonusBanner)
			if id != "" {
				info, _ = service.MessageService.GetBonusBannerById(id)
			}
			// 将图片上传服务器
			err2 := service.UploadRequest(ImgName, f)
			beego.Trace("result: ", err2)
			if err2 != nil {
				this.checkError(errors.New("上传图片到服务器失败！"))
				return
			}
			imgUrl := service.UploadUrl(ImgName) // 拼接img地址
			info.Rtype = type_id
			info.ImgName = ImgName
			info.Url = imgUrl
			info.AUrl = a_url
			err1 := service.MessageService.AddOrUpdateBonusBanner(info)
			this.checkError(err1)
		} else {
			// banner页
			info := new(entity.Banner)
			if id != "" {
				info, _ = service.MessageService.GetBannerById(id)
			}
			// 将图片上传服务器
			err2 := service.UploadRequest(ImgName, f)
			beego.Trace("result: ", err2)
			if err2 != nil {
				this.checkError(errors.New("上传图片到服务器失败！"))
				return
			}
			imgUrl := service.UploadUrl(ImgName) // 拼接img地址
			info.Rtype = type_id
			info.ImgName = ImgName
			info.Url = imgUrl
			info.AUrl = a_url
			err1 := service.MessageService.AddOrUpdateBanner(info)
			this.checkError(err1)
		}

		service.ActionService.Notice(this.auth.GetUserName(), id)
		this.redirect(beego.URLFor("MessageController.NoticeList", "tabid", tabid))
		// service.GmMsg(pb.WebUploadConfig)
	}
}

// 删除Banner数据
func (this *MessageController) DelBanner() {
	id := this.GetString("id")
	tabid, _ := this.GetInt("tabid")
	if id == "" {
		this.checkError(errors.New("无法操作该Banner页!"))
		return
	}
	if tabid == 2 {
		m := bson.M{}
		count, _ := service.MessageService.GetBannerTotal(m)
		if count == 2 {
			this.checkError(errors.New("不能删除该Banner页，最少保留两条Banner页!"))
			return
		}
	}
	var err error
	var action string
	if tabid == 7 {
		action = "del_bonus_banner"
		err = service.MessageService.DelBonusBanner(id)
	} else {
		action = "del_banner"
		// info, _ := service.MessageService.GetBannerById(id)
		err = service.MessageService.DelBanner(id)
	}

	if err != nil {
		this.checkError(err)
	} else {
		this.showMsg("删除成功！", MSG_OK, beego.URLFor("MessageController.NoticeList", "tabid", tabid))
		service.ActionService.Add(action, this.auth.GetUser().UserName, "", id, id, "")
	}
}

// Banner权重设置
func (this *MessageController) BannerSort() {
	id := this.GetString("sid")
	tabid, _ := this.GetInt("tabid")
	if id == "" {
		this.checkError(errors.New("无法操作该Banner页！"))
	}
	sortid, _ := this.GetInt("bannersort")
	if id != "" {
		if tabid == 7 {
			info := new(entity.BonusBanner)
			info, _ = service.MessageService.GetBonusBannerById(id)
			if info != nil {
				info.SortId = sortid
				err := service.MessageService.UpdateBonusBannerSort(info)
				this.checkError(err)

				service.ActionService.Notice(this.auth.GetUserName(), id)
				this.redirect(beego.URLFor("MessageController.NoticeList", "tabid", 7))
			} else {
				this.checkError(errors.New("无法操作该Banner页！"))
			}
		} else {
			info := new(entity.Banner)
			info, _ = service.MessageService.GetBannerById(id)
			if info != nil {
				info.SortId = sortid
				err := service.MessageService.UpdateBannerSort(info)
				this.checkError(err)

				service.ActionService.Notice(this.auth.GetUserName(), id)
				this.redirect(beego.URLFor("MessageController.NoticeList", "tabid", 2))
			} else {
				this.checkError(errors.New("无法操作该Banner页！"))
			}
		}
	}

}

func (this *MessageController) EditContactWay() {
	id := this.GetString("wid")
	info := new(entity.ModifyCustomer)
	if id != "" {
		info, _ = service.MessageService.GetContactWayById(id)
	}
	if this.isPost() {
		mail := this.GetString("mail")
		tg := this.GetString("telegram")
		areacode := this.GetString("areacode")
		whatsapp := this.GetString("whatsapp")

		info.Mail = mail
		info.Telegram = tg
		info.WhatsApp = whatsapp
		info.AreaCode = areacode
		err := this.validContactWay(info)
		this.checkError(err)
		info.WhatsApp = areacode + "-" + whatsapp
		err1 := service.MessageService.AddOrUpdateContactWay(info)
		this.checkError(err1)
		// 通知服务器
		newId := *&info.Id
		req := map[string]entity.ModifyCustomer{
			newId: *info,
		}
		result, err := service.GmRequest(pb.WebModifyCustomer, pb.CONFIG_UPSERT, req)
		beego.Trace("result: ", result)
		this.checkError(err)

		service.ActionService.Notice(this.auth.GetUserName(), id)
		this.redirect(beego.URLFor("MessageController.NoticeList", "tabid", 3))
	}
}

func (this *MessageController) validContactWay(info *entity.ModifyCustomer) error {
	valid := validation.Validation{}
	valid.Required(info.Mail, "name").Message("邮箱不能为空")
	valid.Required(info.Telegram, "info").Message("Telegram不能为空")
	valid.Required(info.WhatsApp, "number").Message("WhatsApp不能为空")
	valid.Required(info.AreaCode, "areacode").Message("区号不能为空")

	if valid.HasErrors() {
		for _, err := range valid.Errors {
			return errors.New(err.Message)
		}
	}

	return nil
}

func (this *MessageController) EditCdkey() {
	id := this.GetString("cid")
	info := new(entity.CdKeyConfig)
	if id != "" {
		info, _ = service.MessageService.GetCdkeyById(id)
	}
	if this.isPost() {
		cdkey := this.GetString("cdkey")
		coin, _ := this.GetInt("coin")
		limit, _ := this.GetInt("limit")
		userId := this.GetString("userId")
		remark := this.GetString("remark")
		etime := this.GetString("end_date2")
		e := fmt.Sprintf("%s 23:59:59", etime)
		if userId != "" {
			userinfo, _ := service.PlayerService.GetUser(userId)
			if userinfo == nil || userinfo.Userid == "" {
				this.checkError(errors.New("指定的用户ID不存在！"))
			}
		}
		if etime == "" {
			this.checkError(errors.New("请选择结束时间"))
		}
		if cdkey == "" {
			this.checkError(errors.New("CDKEY不能为空"))
		}
		if limit <= 0 {
			this.checkError(errors.New("额度不能小于0"))
		}
		if limit > 10000000 {
			this.checkError(errors.New("额度不能大于10000000"))
		}
		info.Key = cdkey
		info.Remark = remark
		info.UserId = userId
		diamond := int64(0)
		givediamond := int64(0)
		outdiamond := int64(0)
		bouns := int64(0)
		switch coin {
		case 0:
			diamond = int64(limit)
		case 1:
			givediamond = int64(limit)
		case 2:
			outdiamond = int64(limit)
		case 3:
			bouns = int64(limit)
		}
		info.Diamond = diamond
		info.GiveDiamond = givediamond
		info.OutDiamond = outdiamond
		info.Bouns = bouns
		info.Etime = utils.Str2Time(e, service.Location())
		info.OperatorId = this.auth.GetUserId()
		info.OperatorName = this.auth.GetUserName()
		err1 := service.MessageService.AddOrUpdateCdkey(info)
		if err1 != nil {
			this.checkError(err1)
		}
		service.ActionService.Notice(this.auth.GetUserName(), id)
		this.redirect(beego.URLFor("MessageController.NoticeList", "tabid", 5))
	}
	this.Data["info"] = info
}

func (this *MessageController) DelCdkey() {
	id := this.GetString("id")
	if id == "" {
		this.checkError(errors.New("无法操作该CDKEY!"))
		return
	}
	err := service.MessageService.DelCdkey(id)
	if err != nil {
		this.checkError(err)
	} else {
		this.showMsg("删除成功！", MSG_OK, beego.URLFor("MessageController.NoticeList", "tabid", 5))
		service.ActionService.Add("del_cdkkey", this.auth.GetUser().UserName, "", id, id, "")
	}
}

/*
客服功能相关
*/
func (this *MessageController) EditCustomerService() {
	if this.isPost() {
		cid := this.GetString("cid")
		mark, _ := this.GetInt("mark")
		count, _ := this.GetInt("count")
		Info := new(entity.CustomerServiceConfig)
		if cid != "" {
			Info.Id = cid
		}
		userlist := make([]string, 0)
		userlist2 := make([]string, 0)
		userlist3 := make([]string, 0)
		if mark != 0 {
			for i := 0; i < count; i++ {
				s := "onebox_" + strconv.Itoa(i)
				pid := this.GetString(s)
				if pid != "" {
					userlist = append(userlist, pid)
				}
			}
			if len(userlist) == 0 {
				this.checkError(errors.New("请选择10:00-19:30分配用户！"))
			}
			for i := 0; i < count; i++ {
				s := "twobox_" + strconv.Itoa(i)
				pid := this.GetString(s)
				if pid != "" {
					userlist2 = append(userlist2, pid)
				}
			}
			if len(userlist2) == 0 {
				this.checkError(errors.New("请选择19:30-02:30分配用户！"))
			}

			for i := 0; i < count; i++ {
				s := "threebox_" + strconv.Itoa(i)
				pid := this.GetString(s)
				if pid != "" {
					userlist3 = append(userlist3, pid)
				}
			}
			if len(userlist3) == 0 {
				this.checkError(errors.New("请选择02:30-10:00分配用户！"))
			}
		}
		Info.Switch = mark
		Info.Users = userlist
		Info.Users2 = userlist2
		Info.Users3 = userlist3
		Info.IsReset = true
		err := service.MessageService.SetCustomerUser(Info)
		this.checkError(err)
		this.redirect(beego.URLFor("MessageController.NoticeList", "tabid", 4))
	}
}

// EditShare 分享活动配置
func (this *MessageController) EditShare() {
	id := this.GetString("sid")
	info := new(entity.ModifyShare)
	if id != "" {
		info, _ = service.MessageService.GetShareById(id)
	}
	if this.isPost() {
		youtobe := this.GetString("youtobe")
		ins := this.GetString("ins")
		facebook := this.GetString("facebook")
		telegram := this.GetString("telegram")
		cashPrize := this.GetString("cashPrize")
		whatsApp := this.GetString("whatsApp")
		x := this.GetString("x")

		info = &entity.ModifyShare{
			Id:        info.Id,
			Youtobe:   youtobe,
			Ins:       ins,
			Facebook:  facebook,
			Telegram:  telegram,
			CashPrize: cashPrize,
			WhatsApp:  whatsApp,
			X:         x,
			Utime:     utils.LocalTime(),
		}

		err1 := service.MessageService.AddOrUpdateShare(info)
		this.checkError(err1)
		// 通知服务器
		newId := info.Id
		req := map[string]entity.ModifyShare{
			newId: *info,
		}
		result, err := service.GmRequest(pb.WebModifyShare, pb.CONFIG_UPSERT, req)
		beego.Trace("result: ", result)
		this.checkError(err)

		service.ActionService.Notice(this.auth.GetUserName(), id)
		this.redirect(beego.URLFor("MessageController.NoticeList", "tabid", 6))
	}
}

// Customer 客服界面
func (this *MessageController) Customer() {
	action := this.GetString("action")
	adminId := this.auth.GetUser().Id
	adminName := this.auth.GetUser().UserName

	// 客服回复消息
	if action == "Reply" {
		args := new(args.CustomerReplayArgs)
		this.JsonBody(args)

		msg, err := service.MessageService.CustomerReply(adminId, adminName, args.Userid, args.Ctype, args.Content)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}
		this.JsonRSuccess(msg)
		return
	}

	// 后台结束聊天
	if action == "CloseSesion" {
		userid := this.GetString("userid")
		etime, err := service.MessageService.CustomerCloseSesion(adminId, userid)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}
		this.JsonRSuccess(libs.R{
			"etime": etime,
		})
		return
	}

	// 查会话历史
	if action == "SessionHistory" {
		query := this.GetString("query")
		sessionId := this.GetString("sessionId")
		self, _ := this.GetInt("self")
		maxId, _ := this.GetInt64("maxMessageId")
		sessionHasMore, curVersion, nextSessionId, maxMessageId, sessions, err := service.MessageService.CustomerSessionHistory(adminId, sessionId, query, self == 1)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}
		maxMessageId = max(maxId, maxMessageId)
		this.JsonRSuccess(libs.R{
			"nextSessionId":  fmt.Sprint(nextSessionId),
			"maxMessageId":   fmt.Sprint(maxMessageId),
			"sessions":       sessions,
			"adminId":        adminId,
			"adminName":      adminName,
			"version":        curVersion,
			"sessionHasMore": sessionHasMore,
		})
		return
	}

	// 查消息历史
	if action == "MessageHistory" {
		userid := this.GetString("userid")
		sessionId := this.GetString("sessionId")
		prevMessageId := this.GetString("minMessageId")
		hasMore, minMessageId, messages, err := service.MessageService.CustomerMessageHistory(adminId, userid, sessionId, prevMessageId)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}
		this.JsonRSuccess(libs.R{
			"hasMore":      hasMore,
			"minMessageId": fmt.Sprint(minMessageId),
			"messages":     messages,
		})
		return
	}

	// 广播消息查询
	if action == "Subscribe" {
		id, _ := this.GetInt64("id")
		version, _ := this.GetInt64("version")
		if id == 0 {
			this.JsonRError(fmt.Sprintf("subscribe id error: %s", id))
			return
		}
		datas, maxId, refrash, err := service.MessageService.CustomerSubscribe(adminId, id, version)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}
		this.JsonRSuccess(libs.R{
			"maxId":   fmt.Sprint(maxId),
			"refrash": refrash,
			"datas":   datas,
		})
		return
	}

	// 消息撤回
	if action == "MessageRevoke" {
		messageId := this.GetString("messageId")
		msgId, err := strconv.ParseInt(messageId, 10, 64)
		if err != nil || msgId == 0 {
			this.JsonRError(fmt.Sprintf("messageId error: %s, %s", messageId, err.Error()))
			return
		}
		err = service.MessageService.CustomerMessageRevoke(adminId, msgId)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}
		this.JsonRSuccess(libs.R{})
		return
	}

	// 客服已读玩家消息
	if action == "ReplyRead" {
		sessionId, _ := this.GetInt64("sessionId")
		messageId, _ := this.GetInt64("messageId")
		err := service.MessageService.CustomerMsgReplyRead(adminId, sessionId, []int64{messageId})
		if err != nil {
			this.JsonRError(err.Error())
			return
		}
		this.JsonRSuccess(libs.R{})
		return
	}

	// 客服已读玩家消息
	if action == "Customer" {
		customer, err := service.MessageService.GetCustomerSeat(adminId)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}
		this.JsonRSuccess(libs.R{
			"customer": customer,
		})
		return
	}

	// 打开/关闭客服
	if action == "OpenMyCustomerSeat" {
		open, _ := this.GetInt("open")
		err := service.MessageService.OpenCustomerSeat(adminId, []string{adminId}, open == 1)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}
		this.JsonRSuccess(libs.R{})
		return
	}

	// 快捷语列表
	if action == "ListCustomerPhrases" {
		phrases, err := service.MessageService.ListCustomerPhrases(adminId)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}
		this.JsonRSuccess(libs.R{
			"phrases": phrases,
		})
		return
	}

	// 新增快捷语
	if action == "AddCustomerPhrase" {
		text := this.GetString("text")
		err := service.MessageService.AddCustomerPhrase(adminId, text)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}
		this.JsonRSuccess(libs.R{})
		return
	}

	// 编辑快捷语
	if action == "UpdateCustomerPhrase" {
		id, _ := this.GetInt64("id")
		text := this.GetString("text")
		err := service.MessageService.UpdateCustomerPhrase(adminId, id, text)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}
		this.JsonRSuccess(libs.R{})
		return
	}

	// 排序快捷语
	if action == "SortCustomerPhrase" {
		sorts := this.GetString("sorts")
		sorts1 := strings.Split(sorts, ",")
		var sorts0 [][]int64
		for _, sort := range sorts1 {
			sorter := strings.Split(sort, "-")
			id, _ := strconv.ParseInt(sorter[0], 10, 64)
			nextId, _ := strconv.ParseInt(sorter[1], 10, 64)
			sorts0 = append(sorts0, []int64{id, nextId})
		}
		err := service.MessageService.SortCustomerPhrase(adminId, sorts0)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}
		this.JsonRSuccess(libs.R{})
		return
	}

	// 头像
	if action == "Avatars" {
		userid := this.GetString("userid")
		status, _ := this.GetInt("status") // 0:全部 10:待审核 11：通过；12：拒绝
		page, pageSize, _, _ := this.PageArgs()

		m1 := bson.M{}
		if userid != "" {
			m1["user_id"] = userid
		}
		if status > 0 {
			m1["is_audit"] = status - 10
		}
		count, err := service.MessageService.GetUploadHeadTotal(m1)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}
		list, err := service.MessageService.GetUploadHeadList(page, pageSize, m1)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}
		this.JsonRSuccess(libs.R{
			"datas": list,
			"total": count,
		})
		return
	}

	// 头像审核
	if action == "AvatarAudit" {
		id := this.GetString("id")
		audit, _ := this.GetInt("audit")
		if !utils.SliceIn(audit, 1, 2) {
			this.JsonRError(fmt.Sprintf("审核状态有误: %d", audit))
			return
		}

		name := this.auth.GetUser().UserName
		msg := new(entity.UserCustomPhoto)
		msg.Id = id
		msg.Result = audit // 1.通过 2.拒绝
		result, err := service.GmRequest(pb.WebUserPhoto, pb.CONFIG_UPSERT, msg)
		beego.Trace("result: ", result)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}

		// 修改DB
		info := new(entity.UploadUserHead)
		info.Id = id
		info.IsAudit = int64(audit)
		info.ATime = bson.Now().Unix()
		info.Auditor = name
		err = service.MessageService.UpdateUserHead(info)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}
		service.ActionService.Add(utils.CaseElse(audit == 1, "upload_head_pass", "upload_head_refuse"), name,
			"", utils.String(id), utils.String(id), "")
		this.JsonRSuccess(libs.R{})
		return
	}

	// 全部消息
	if action == "AllSessions" {
		arg := new(args.AllSessionsArgs)
		this.JsonBody(arg)
		page, pageSize, _, _ := this.PageArgs()

		var stime, etime time.Time
		if arg.Userid == "" {
			today := service.NowTime()
			if arg.Stime == "" {
				// 默认近七天数据
				arg.Stime = today.AddDate(0, 0, -6).Format("2006-01-02")
			}
			if arg.Etime == "" {
				arg.Etime = today.Format("2006-01-02")
			}
			s, e := service.FindByDate4(arg.Stime, arg.Etime)
			stime, etime = *s, *e
		} else {
			arg.Stime = ""
			arg.Etime = ""
		}
		total, sessions, stats, err := service.MessageService.AllSessions(adminId, page, pageSize, arg.Userid, stime, etime, arg.QuestionType, arg.Resolved)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}
		this.JsonRSuccess(libs.R{
			"sessions": sessions,
			"stats":    stats,
			"total":    total,
			"stime":    arg.Stime,
			"etime":    arg.Etime,
		})
		return
	}

	if action == "SetSessionQuestionType" {
		sessionId, _ := this.GetInt64("sessionId")
		questionType, _ := this.GetInt32("questionType")
		err := service.MessageService.SetSessionQuestionType(adminId, sessionId, questionType)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}
		this.JsonRSuccess(libs.R{})
		return
	}
	if action == "SetSessionResovled" {
		sessionId, _ := this.GetInt64("sessionId")
		resolved, _ := this.GetInt32("resolved")
		err := service.MessageService.SetSessionResovled(adminId, sessionId, resolved == 1)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}
		this.JsonRSuccess(libs.R{})
		return
	}
	if action == "SetSessionRemark" {
		sessionId, _ := this.GetInt64("sessionId")
		remark := this.GetString("remark")
		err := service.MessageService.SetSessionRemark(adminId, sessionId, remark)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}
		this.JsonRSuccess(libs.R{})
		return
	}

	this.Data["pageTitle"] = "客服消息"
	this.display()
}

// CustomerManager 客服管理界面
func (this *MessageController) CustomerManager() {
	action := this.GetString("action")
	adminId := this.auth.GetUser().Id
	adminName := this.auth.GetUser().UserName

	// todo 权限

	// 客服列表查询
	if action == "ListCustomerSeat" {
		customers, uncustomerUsers, setting, err := service.MessageService.ListCustomerSeat(adminId, true)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}
		this.JsonRSuccess(libs.R{
			"customers":       customers,
			"uncustomerUsers": uncustomerUsers,
			"setting":         setting,
		})
		return
	}
	// 分单逻辑配置修改
	if action == "SeatDispatchEdit" {
		dispatch, _ := this.GetInt32("dispatch")
		err := service.MessageService.CustomerSeatDispatchUpdate(adminId, dispatch)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}
		this.JsonRSuccess(libs.R{})
		return
	}

	// 添加客服
	if action == "AddCustomerSeat" {
		customer := this.GetString("customer")
		err := service.MessageService.AddCustomerSeat(adminName, customer)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}
		this.JsonRSuccess(libs.R{})
		return
	}

	// 删除客服
	if action == "DeleteCustomerSeat" {
		customer := this.GetString("customer")
		err := service.MessageService.DeleteCustomerSeat(customer)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}
		this.JsonRSuccess(libs.R{})
		return
	}

	// 打开/关闭客服
	if action == "OpenCustomerSeat" {
		customers := this.GetString("customers")
		open, _ := this.GetInt("open")
		customers0 := strings.Split(customers, ",")
		err := service.MessageService.OpenCustomerSeat(adminId, customers0, open == 1)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}
		this.JsonRSuccess(libs.R{})
		return
	}

	// 用户信息
	if action == "UserSessions" {
		// stime, etime time.Time, userid, customer string, status int, rank int, asc bool
		arg := new(args.UserSessionsArgs)
		this.JsonBody(arg)
		page, pageSize, _, asc := this.PageArgs()
		// if sortBy != "" && sortBy != "s0.vip_lv" {
		// 	this.JsonRFail("排序字段有误")
		// 	return
		// }

		// orderBy = "s0.id"
		// orderBy = "s0.vip_lv"
		today := service.NowTime()
		if arg.Stime == "" {
			// 默认近七天数据
			arg.Stime = today.AddDate(0, 0, -6).Format("2006-01-02")
		}
		if arg.Etime == "" {
			arg.Etime = today.Format("2006-01-02")
		}
		s, e := service.FindByDate4(arg.Stime, arg.Etime)
		stime, etime := *s, *e

		customers, _, _, err := service.MessageService.ListCustomerSeat(adminId, false)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}

		total, sessions, err := service.MessageService.CustomerUserSessions(page, pageSize, stime, etime, arg.Userid, arg.Customer, int(arg.Status), 0, asc == 1)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}
		this.JsonRSuccess(libs.R{
			"total":     total,
			"sessions":  sessions,
			"customers": customers,
			"stime":     arg.Stime,
			"etime":     arg.Etime,
		})
		return
	}

	// 修改会话客服
	if action == "SetSessionsCustomer" {
		userid := this.GetString("userid")
		sessionId := this.GetString("sessionId")
		customer := this.GetString("customer")
		customerName := this.GetString("customerName")
		err := service.MessageService.CustomerSessionSet(userid, sessionId, customer, customerName)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}
		this.JsonRSuccess(libs.R{})
		return
	}

	// 客服表现
	if action == "CustomerExpStats" {
		arg := new(args.CustomerExpStatsArgs)
		this.JsonBody(arg)

		today := service.NowTime()
		if arg.Stime == "" {
			// 默认近七天数据
			arg.Stime = today.AddDate(0, 0, -6).Format("2006-01-02")
		}
		if arg.Etime == "" {
			arg.Etime = today.Format("2006-01-02")
		}
		s, e := service.FindByDate4(arg.Stime, arg.Etime)
		stime, etime := *s, *e

		stats, customers, dispatchStats, rateStats, err := service.MessageService.CustomerExpStats(adminId, stime, etime, arg.Userid, arg.Customer)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}
		this.JsonRSuccess(libs.R{
			"stats":         stats,
			"customers":     customers,
			"stime":         arg.Stime,
			"etime":         arg.Etime,
			"dispatchStats": dispatchStats,
			"rateStats":     rateStats,
		})
		return
	}

	this.Data["pageTitle"] = "客服管理"
	this.display()
}
