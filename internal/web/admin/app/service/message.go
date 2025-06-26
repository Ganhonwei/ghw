package service

import (
	"errors"
	"fmt"
	"goserver/gen/pb"
	"goserver/internal/web/admin/app/entity"
	"goserver/pkg/data/ck"
	"goserver/pkg/data/mq"
	"goserver/pkg/utils"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/globalsign/mgo/bson"
	"golang.org/x/sync/singleflight"
)

type messageService struct{}

// 获取未读消息
func (this *messageService) GetUnreadMessage() int {
	m := bson.M{}
	m["reply_status"] = 1
	count, _ := this.GetCustomerTotal(m)

	return int(count)
}

// 获取客服消息列表
func (this *messageService) CustomerList(page, pageSize int, m bson.M) ([]*entity.FeedBackLog, error) {
	var list []*entity.FeedBackLog
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "utime", false)
	CustomerMsgs.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList(list)
	return list, nil
}

func (this *messageService) chipList(list []*entity.FeedBackLog) []*entity.FeedBackLog {
	var userIds []string

	var userlist []entity.User
	Users.Find(bson.M{}).All(&userlist)
	for i, v := range list {
		userIds = append(userIds, v.Userid)
		c, _ := ConvertToIndiaTime(v.LastTime.Unix())
		v.LastTime = c
		c1, _ := ConvertToIndiaTime(v.Utime.Unix())
		v.Utime = c1
		for j, k := range v.Logs {
			c1, _ := ConvertToIndiaTime(k.Ctime.Unix())
			k.Ctime = c1
			v.Logs[j] = k
		}
		username := ""
		if v.AssignUser != "" {
			if len(userlist) > 0 {
				for _, item := range userlist {
					if item.Id == v.AssignUser {
						username = item.UserName
						break
					}
				}
			}
		}
		v.AssignUserName = username
		list[i] = v
	}

	if len(userIds) > 0 {
		var user_datas []map[string]any
		err := ck.Select(&user_datas, `
			select userid, vip_lv, money, state from game.col_user final where userid in ?
		`, userIds)
		if err != nil {
			beego.Error(err)
			return list
		}
		for _, data := range user_datas {
			userid := data["userid"].(string)
			vip_lv := utils.ToInt64(data["vip_lv"])
			money := utils.ToInt64(data["money"])
			state := utils.ToInt64(data["state"])
			for _, v := range list {
				if v.Userid == userid {
					v.VipLv = vip_lv
					_, v.ChargeType = GetChargeType(money, int(state))
				}
			}
		}
	}
	return list
}

// 获取客服消息总数
func (this *messageService) GetCustomerTotal(m bson.M) (int64, error) {
	return int64(Count(CustomerMsgs, m)), nil
}

// 根据用户ID获取客服消息
func (this *messageService) GetFeedBackLog(id string) (*entity.FeedBackLog, error) {
	userInfo := new(entity.FeedBackLog)
	Get(CustomerMsgs, id, userInfo)
	if userInfo.Userid == "" {
		return userInfo, errors.New("用户不存在")
	}
	userInfo = this.chipList27(userInfo)
	return userInfo, nil
}

func (this *messageService) chipList27(info *entity.FeedBackLog) *entity.FeedBackLog {
	if info != nil {
		for k, v := range info.Logs {
			c, _ := ConvertToIndiaTime(v.Ctime.Unix())
			v.Ctime = c
			info.Logs[k] = v
		}

		var user_data = make(map[string]any)
		err := ck.Select(&user_data, `
			select userid, vip_lv, money, state from game.col_user final where userid = ?
		`, info.Userid)
		if err != nil {
			beego.Error(err)
			return info
		}
		userid := user_data["userid"].(string)
		vip_lv := utils.ToInt64(user_data["vip_lv"])
		money := utils.ToInt64(user_data["money"])
		state := utils.ToInt64(user_data["state"])
		_ = userid
		info.VipLv = vip_lv
		_, info.ChargeType = GetChargeType(money, int(state))
	}
	return info
}

// 新增客服消息
func (this *messageService) AddFeedBackLog(log *entity.FeedBackLog) error {
	if !Insert(CustomerMsgs, log) {
		return errors.New("写入失败:" + log.Userid)
	}
	return nil
}

// 更新客服消息
func (this *messageService) UpdateFeedBackLog(log *entity.FeedBackLog) error {
	m := bson.M{"_id": log.Userid}
	n := bson.M{
		"utime":        bson.Now(),
		"reply_status": log.ReplyStatus,
		"log":          log.Logs,
	}
	if Update(CustomerMsgs, m, bson.M{"$set": n}) {
		return nil
	}
	return errors.New("更新失败")
}

// 更新沟通状态
func (this *messageService) UpdateReplyStatus(log *entity.FeedBackLog) error {
	m := bson.M{"_id": log.Userid}
	n := bson.M{
		"reply_status": log.ReplyStatus,
	}
	if Update(CustomerMsgs, m, bson.M{"$set": n}) {
		return nil
	}
	return errors.New("更新失败")
}

func (this *messageService) UpdateAssignUser(log *entity.FeedBackLog) error {
	m := bson.M{"_id": log.Userid}
	n := bson.M{
		"assign_user": log.AssignUser,
	}
	if Update(CustomerMsgs, m, bson.M{"$set": n}) {
		return nil
	}
	return errors.New("更新失败")
}

// 添加公告
func (this *messageService) AddOrUpdateNotice(notice *entity.Notice) error {
	if notice.Id != "" {
		// 修改
		notice.Ctime = bson.Now()
		m := bson.M{"_id": notice.Id}
		n := bson.M{
			"rtype":   notice.Rtype,
			"num":     notice.Num,
			"content": notice.Content,
			"stime":   notice.Stime,
			"etime":   notice.Etime,
			"ctime":   notice.Ctime,
		}
		if Update(Notices, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败")
	} else {
		// 新增
		notice.Id = bson.NewObjectId().Hex()
		notice.Ctime = bson.Now()
		if !Insert(Notices, notice) {
			return errors.New("写入失败:" + notice.Id)
		}
		return nil
	}

}

// 获取
func (this *messageService) GetNotice(id string) (*entity.Notice, error) {
	notice := new(entity.Notice)
	Get(Notices, id, notice)
	if notice.Id == "" {
		return notice, errors.New("消息不存在")
	}
	return notice, nil
}

// 获取
func (this *messageService) DelNotice(id string) error {
	if Update(Notices, bson.M{"_id": id}, bson.M{"$set": bson.M{"del": 1}}) {
		return nil
	}
	return errors.New("移除失败")
}

// 获取列表
func (this *messageService) GetNoticeList(page, pageSize int, m bson.M) ([]entity.Notice, error) {
	var list []entity.Notice
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	err := Notices.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList1(list)
	return list, err
}
func (this *messageService) chipList1(list []entity.Notice) []entity.Notice {
	for i, v := range list {
		c, _ := ConvertToIndiaTime(v.Ctime.Unix())
		v.Ctime = c
		c1, _ := ConvertToIndiaTime(v.Etime.Unix())
		v.Etime = c1
		c1, _ = ConvertToIndiaTime(v.Stime.Unix())
		v.Stime = c1
		list[i] = v
	}
	return list
}

// 获取总数
func (this *messageService) GetNoticeListTotal(m bson.M) (int64, error) {
	return int64(Count(Notices, m)), nil
}

/*
	Bonus中心页相关接口
*/
// 分页查询Banner列表
func (this *messageService) GetBonusBannerList(page, pageSize int, m bson.M) ([]entity.BonusBanner, error) {
	var list []entity.BonusBanner
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "sort_id", true)
	err := BonusBanners.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	for i, v := range list {
		c, _ := ConvertToIndiaTime(v.Ctime.Unix())
		v.Ctime = c
		list[i] = v
	}
	return list, err
}

// 获取banner页条数
func (this *messageService) GetBonusBannerTotal(m bson.M) (int64, error) {
	return int64(Count(BonusBanners, m)), nil
}

// 新增、修改banner页数据
func (this *messageService) AddOrUpdateBonusBanner(banners *entity.BonusBanner) error {
	if banners.Id != "" {
		// 修改
		banners.Ctime = bson.Now()
		m := bson.M{"_id": banners.Id}
		n := bson.M{
			"rtype":    banners.Rtype,
			"sort_id":  banners.SortId,
			"img_name": banners.ImgName,
			"url":      banners.Url,
			"a_url":    banners.AUrl,
			"ctime":    banners.Ctime,
		}
		if Update(BonusBanners, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败")
	} else {
		// 新增
		banners.Id = bson.NewObjectId().Hex()
		banners.Ctime = bson.Now()
		if !Insert(BonusBanners, banners) {
			return errors.New("写入失败:" + banners.Id)
		}
		return nil
	}
}

// 获取
func (this *messageService) DelBonusBanner(id string) error {
	if Delete(BonusBanners, bson.M{"_id": id}) {
		return nil
	}
	return errors.New("移除失败")
}

// 根据id获取Banner页数据
func (this *messageService) GetBonusBannerById(id string) (*entity.BonusBanner, error) {
	info := new(entity.BonusBanner)
	Get(BonusBanners, id, info)
	if info.Id == "" {
		return info, errors.New("Banner页数据不存在")
	}
	return info, nil
}

func (this *messageService) UpdateBonusBannerSort(banners *entity.BonusBanner) error {
	if banners.Id != "" {
		// 修改
		banners.Ctime = bson.Now()
		m := bson.M{"_id": banners.Id}
		n := bson.M{
			"sort_id": banners.SortId,
			"ctime":   banners.Ctime,
		}
		if Update(BonusBanners, m, bson.M{"$set": n}) {
			return nil
		}
	}
	return errors.New("更新失败")
}

/*
	Banner页相关接口
*/
// 分页查询Banner列表
func (this *messageService) GetBannerList(page, pageSize int, m bson.M) ([]entity.Banner, error) {
	var list []entity.Banner
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "sort_id", true)
	err := Banners.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList2(list)
	return list, err
}

func (this *messageService) chipList2(list []entity.Banner) []entity.Banner {
	for i, v := range list {
		c, _ := ConvertToIndiaTime(v.Ctime.Unix())
		v.Ctime = c
		list[i] = v
	}
	return list
}

// 获取banner页条数
func (this *messageService) GetBannerTotal(m bson.M) (int64, error) {
	return int64(Count(Banners, m)), nil
}

// 新增、修改banner页数据
func (this *messageService) AddOrUpdateBanner(banners *entity.Banner) error {
	if banners.Id != "" {
		// 修改
		banners.Ctime = bson.Now()
		m := bson.M{"_id": banners.Id}
		n := bson.M{
			"rtype":    banners.Rtype,
			"sort_id":  banners.SortId,
			"img_name": banners.ImgName,
			"url":      banners.Url,
			"a_url":    banners.AUrl,
			"ctime":    banners.Ctime,
		}
		if Update(Banners, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败")
	} else {
		// 新增
		banners.Id = bson.NewObjectId().Hex()
		banners.Ctime = bson.Now()
		if !Insert(Banners, banners) {
			return errors.New("写入失败:" + banners.Id)
		}
		return nil
	}
}

// 获取
func (this *messageService) DelBanner(id string) error {
	if Delete(Banners, bson.M{"_id": id}) {
		return nil
	}
	return errors.New("移除失败")
}

// 根据id获取Banner页数据
func (this *messageService) GetBannerById(id string) (*entity.Banner, error) {
	info := new(entity.Banner)
	Get(Banners, id, info)
	if info.Id == "" {
		return info, errors.New("Banner页数据不存在")
	}
	return info, nil
}

func (this *messageService) UpdateBannerSort(banners *entity.Banner) error {
	if banners.Id != "" {
		// 修改
		banners.Ctime = bson.Now()
		m := bson.M{"_id": banners.Id}
		n := bson.M{
			"sort_id": banners.SortId,
			"ctime":   banners.Ctime,
		}
		if Update(Banners, m, bson.M{"$set": n}) {
			return nil
		}
	}
	return errors.New("更新失败")
}

/*
联系方式相关接口
*/
func (this *messageService) GetContactWayList(page, pageSize int, m bson.M) ([]entity.ModifyCustomer, error) {
	var list []entity.ModifyCustomer
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "sort_id", true)
	err := ContactWays.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList3(list)
	return list, err
}

func (this *messageService) chipList3(list []entity.ModifyCustomer) []entity.ModifyCustomer {
	for i, v := range list {
		c, _ := ConvertToIndiaTime(v.Utime.Unix())
		v.Utime = c
		list[i] = v
	}
	return list
}

// 获取banner页条数
func (this *messageService) GetContactWayTotal(m bson.M) (int64, error) {
	return int64(Count(ContactWays, m)), nil
}

// 新增、修改banner页数据
func (this *messageService) AddOrUpdateContactWay(infos *entity.ModifyCustomer) error {
	if infos.Id != "" {
		// 修改
		infos.Utime = bson.Now()
		m := bson.M{"_id": infos.Id}
		n := bson.M{
			"mail":      infos.Mail,
			"telegram":  infos.Telegram,
			"whats_app": infos.WhatsApp,
			"utime":     infos.Utime,
		}
		if Update(ContactWays, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败")
	} else {
		// 新增
		infos.Id = bson.NewObjectId().Hex()
		infos.Utime = bson.Now()
		if !Insert(ContactWays, infos) {
			return errors.New("写入失败:" + infos.Id)
		}
		return nil
	}
}

func (this *messageService) GetContactWayById(id string) (*entity.ModifyCustomer, error) {
	info := new(entity.ModifyCustomer)
	Get(ContactWays, id, info)
	if info.Id == "" {
		return nil, errors.New("联系方式数据不存在")
	}
	return info, nil
}

/*
CDKEY功能相关接口 Cdkeys
*/
func (this *messageService) GetCdkeyList(page, pageSize int, m bson.M) ([]entity.CdKeyConfig, error) {
	var list []entity.CdKeyConfig
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "etime", false)
	err := Cdkeys.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList4(list)
	return list, err
}

func (this *messageService) chipList4(list []entity.CdKeyConfig) []entity.CdKeyConfig {
	for i, v := range list {
		u, _ := ConvertToIndiaTime(v.Utime.Unix())
		v.Utime = u
		c, _ := ConvertToIndiaTime(v.Ctime.Unix())
		v.Ctime = c
		e, _ := ConvertToIndiaTime(v.Etime.Unix())
		v.Etime = e
		v.FDiamond = ComputeFloat(v.Diamond, 100)
		v.FGiveDiamond = ComputeFloat(v.GiveDiamond, 100)
		v.FOutDiamond = ComputeFloat(v.OutDiamond, 100)
		v.FBouns = ComputeFloat(v.Bouns, 100)
		list[i] = v
	}
	return list
}

// 获取CDKEY条数
func (this *messageService) GetCdkeyTotal(m bson.M) (int64, error) {
	return int64(Count(Cdkeys, m)), nil
}

func (this *messageService) GetCdkeyById(id string) (*entity.CdKeyConfig, error) {
	info := new(entity.CdKeyConfig)
	GetByQ(Cdkeys, bson.M{"_id": id}, info)
	if info.Key == "" {
		return nil, errors.New("CDKEY数据不存在")
	}
	return info, nil
}

// 新增、修改CDKEY数据
func (this *messageService) AddOrUpdateCdkey(infos *entity.CdKeyConfig) error {
	info := new(entity.CdKeyConfig)
	GetByQ(Cdkeys, bson.M{"_id": infos.Key}, info)
	if info.Key != "" {
		// 修改
		infos.Utime = bson.Now()
		m := bson.M{"_id": infos.Key}
		n := bson.M{
			"remark":        infos.Remark,
			"diamond":       infos.Diamond,
			"give_diamond":  infos.GiveDiamond,
			"out_diamond":   infos.OutDiamond,
			"bouns":         infos.Bouns,
			"user_id":       infos.UserId,
			"operator_id":   infos.OperatorId,
			"operator_name": infos.OperatorName,
			"ctime":         info.Ctime,
			"etime":         infos.Etime,
			"utime":         infos.Utime,
		}
		if Update(Cdkeys, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败")
	} else {
		// 新增
		infos.Utime = bson.Now()
		infos.Ctime = bson.Now()
		if !Insert(Cdkeys, infos) {
			return errors.New("写入失败:" + infos.Key)
		}
		return nil
	}
}
func (this *messageService) DelCdkey(id string) error {
	if Delete(Cdkeys, bson.M{"_id": id}) {
		return nil
	}
	return errors.New("移除失败")
}

/*
客服功能相关接口
*/
func (this *messageService) GetCustomerUser() ([]entity.User, error) {
	var list []entity.User
	var rolelist []entity.Role
	m := bson.M{}
	m["role_name"] = "客服"
	Roles.Find(m).All(&rolelist)
	if len(rolelist) > 0 {
		var rid []string
		for _, item := range rolelist {
			rid = append(rid, item.Id)
		}
		m1 := bson.M{}
		m1["role_id"] = bson.M{"$in": rid}
		var userrolelist []entity.UserRole
		UserRoles.Find(m1).All(&userrolelist)
		if len(userrolelist) > 0 {
			var uid []string
			for _, item := range userrolelist {
				uid = append(uid, item.UserId)
			}
			m2 := bson.M{}
			m2["_id"] = bson.M{"$in": uid}
			Users.Find(m2).All(&list)
		}
	}
	return list, nil
}

// 设置客服配置
func (this *messageService) SetCustomerUser(infos *entity.CustomerServiceConfig) error {
	if infos.Id != "" {
		// 修改
		infos.Ctime = bson.Now()
		m := bson.M{"_id": infos.Id}
		n := bson.M{
			"switch":   infos.Switch,
			"users":    infos.Users,
			"users2":   infos.Users2,
			"users3":   infos.Users3,
			"is_reset": infos.IsReset,
			"ctime":    infos.Ctime,
		}
		if Update(ChatConfigs, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败")
	} else {
		// 新增
		infos.Id = bson.NewObjectId().Hex()
		infos.Ctime = bson.Now()
		if !Insert(ChatConfigs, infos) {
			return errors.New("写入失败:" + infos.Id)
		}
		return nil
	}
}

// 查询客服配置
func (this *messageService) GetChatConfig() (*entity.CustomerServiceConfig, error) {
	var list []entity.CustomerServiceConfig
	m := bson.M{}
	ChatConfigs.Find(m).All(&list)
	if len(list) > 0 {
		info := list[0]
		return &info, nil
	}
	return nil, errors.New("未查询到客服配置信息")
}

// 客服聊天消息分配
func (this *messageService) ChatAssign() error {
	timestamp := utils.Timestamp()
	tim := utils.Stamp2Time(timestamp)
	endTime := tim                         //.Format("2006-01-02 15:04:05")
	startTime := tim.Add(-5 * time.Minute) //.Format("2006-01-02 15:04:05")

	info, _ := this.GetChatConfig()
	isAssign := false
	isReset := false
	var cuser []string
	if info != nil {
		if info.Switch == 1 {
			isAssign = true
			cuser = GetCustomerUserArr(info)
			// cuser2 = info.Users2
			// cuser3 = info.Users3
			isReset = info.IsReset
		}
	}

	if isAssign {
		// 开始查询需要分配的消息
		m := bson.M{}
		var list []entity.FeedBackLog
		m["reply_status"] = 1
		m["utime"] = bson.M{"$gte": startTime, "$lt": endTime}
		m["assign_user"] = bson.M{"$eq": nil}
		CustomerMsgs.Find(m).All(&list)
		if len(list) > 0 {
			for _, item := range list {
				nextValue := "" // 下一个分配的客服
				if isReset {
					// 重置
					nextValue = cuser[0]
					isReset = false
					info.IsReset = false
					this.SetCustomerUser(info)
				} else {
					//查询上一个分配客服
					if len(cuser) > 0 {
						curindex := 0
						var chatlog []entity.ChatAssignLog
						ChatAssigns.Find(bson.M{}).Sort("-ctime").Limit(1).All(&chatlog)
						if len(chatlog) > 0 {
							loginfo := chatlog[0]
							for i, v := range cuser {
								if v == loginfo.CUserid {
									curindex = i
									break
								}
							}
						} else {
							curindex = -1
						}

						if curindex != -1 && curindex+1 < len(cuser) {
							nextValue = cuser[curindex+1]
							fmt.Println(nextValue) // 输出 "x4"
						} else {
							nextValue = cuser[0]
						}
					}
				}
				// 分配用户
				item.AssignUser = nextValue
				err := this.UpdateAssignUser(&item)
				if err != nil {
					return err
				}
				// 日志记录
				info := new(entity.ChatAssignLog)
				info.Userid = item.Userid
				info.CUserid = nextValue
				info.CSUsers = cuser
				err1 := this.AddChatAssignLog(info)
				if err1 != nil {
					return err1
				}
			}
		}
	}
	//
	return nil
}

func GetCustomerUserArr(info *entity.CustomerServiceConfig) []string {
	isFlag := false
	user := make([]string, 0)
	now := bson.Now()

	startTime := time.Date(now.Year(), now.Month(), now.Day(), 7, 30, 0, 0, now.Location())
	endTime := time.Date(now.Year(), now.Month(), now.Day(), 16, 59, 0, 0, now.Location())
	if now.After(startTime) && now.Before(endTime) {
		isFlag = true
		user = info.Users
	}
	if !isFlag {
		startTime2 := time.Date(now.Year(), now.Month(), now.Day(), 17, 0, 0, 0, now.Location())
		endTime2 := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 0, 0, now.Location())
		if now.After(startTime2) && now.Before(endTime2) {
			isFlag = true
			user = info.Users2
		}
	}
	if !isFlag {
		startTime3 := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endTime3 := time.Date(now.Year(), now.Month(), now.Day(), 7, 29, 0, 0, now.Location())
		if now.After(startTime3) && now.Before(endTime3) {
			isFlag = true
			user = info.Users3
		}
	}
	return user
}

// 添加分配客服聊天消息日志
func (this *messageService) AddChatAssignLog(infos *entity.ChatAssignLog) error {
	// 新增
	infos.Id = bson.NewObjectId().Hex()
	infos.Ctime = bson.Now()
	if !Insert(ChatAssigns, infos) {
		return errors.New("写入失败:" + infos.Id)
	}
	return nil
}

func (this *messageService) GetShareById(id string) (*entity.ModifyShare, error) {
	info := new(entity.ModifyShare)
	Get(ShareWays, id, info)
	if info.Id == "" {
		return info, errors.New("分享配置数据不存在")
	}
	return info, nil
}

// 新增、修改banner页数据
func (this *messageService) AddOrUpdateShare(infos *entity.ModifyShare) error {
	if infos.Id != "" {
		// 修改
		infos.Utime = bson.Now()
		m := bson.M{"_id": infos.Id}
		n := bson.M{
			"youtobe":   infos.Youtobe,
			"ins":       infos.Ins,
			"facebook":  infos.Facebook,
			"telegram":  infos.Telegram,
			"cashPrize": infos.CashPrize,
			"whatsapp":  infos.WhatsApp,
			"x":         infos.X,
			"utime":     infos.Utime,
		}
		if Update(ShareWays, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败")
	} else {
		// 新增
		infos.Id = bson.NewObjectId().Hex()
		infos.Utime = bson.Now()
		if !Insert(ShareWays, infos) {
			return errors.New("写入失败:" + infos.Id)
		}
		return nil
	}
}

/*
CDKEY功能相关接口 Cdkeys
*/
func (this *messageService) GetShareWayList(page, pageSize int, m bson.M) ([]entity.ModifyShare, error) {
	var list []entity.ModifyShare
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "etime", false)
	err := ShareWays.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	// list = this.chipList4(list)
	return list, err
}

/*
 头像审核、列表
*/

// 获取上传头像列表
func (this *messageService) GetUploadHeadList(page, pageSize int, m bson.M) ([]entity.UploadUserHead, error) {
	var list []entity.UploadUserHead
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	UploadUserHeads.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList5(list)
	return list, nil
}

func (this *messageService) chipList5(list []entity.UploadUserHead) []entity.UploadUserHead {
	ids := make([]string, 0)
	for _, v := range list {
		ids = append(ids, v.UserId)
	}
	userList, _ := PlayerService.GetByUser(ids)
	for i, v := range list {
		if v.ATime != 0 {
			c, _ := ConvertToIndiaTime(v.ATime)
			v.S_ATime = c.Format(utils.FORMAT)
		}

		c1, _ := ConvertToIndiaTime1(v.Ctime)
		v.S_Ctime = c1.Format(utils.FORMAT)
		for _, j := range userList {
			if v.UserId == j.Userid {
				v.Nickname = j.Nickname
				v.BundleId = j.AD_BundleId
				break
			}
		}
		list[i] = v

	}
	return list
}

// 获取上传头像总数
func (this *messageService) GetUploadHeadTotal(m bson.M) (int64, error) {
	return int64(Count(UploadUserHeads, m)), nil
}

// 修改传头像数据
func (this *messageService) UpdateUserHead(infos *entity.UploadUserHead) error {
	if infos.Id != "" {
		// 修改
		infos.ATime = bson.Now().Unix()
		m := bson.M{"_id": infos.Id}
		n := bson.M{
			"auditor":  infos.Auditor,
			"a_time":   infos.ATime,
			"is_audit": infos.IsAudit,
		}
		if Update(UploadUserHeads, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败")
	}
	return nil
}

// =======================================================================================================

var (
	customerSessionCache   = &sync.Map{}
	customerSessionLoadSF  singleflight.Group
	customerBoardcastQueue []*CustomerBoardcastLog // 客服后台消息广播队列, 接受新消息/发送新消息/已读/发送
	customerBoardcastLock  sync.RWMutex
	customerVersion        int64 = time.Now().UnixMilli()
)

const (
	CustomerBoardcastNewSession        int32 = 1
	CustomerBoardcastNewMessage        int32 = 2
	CustomerBoardcastReadReplyMessage  int32 = 3 // 玩家已读客服消息
	CustomerBoardcastReadPlayerMessage int32 = 4 // 客服已读玩家消息
)

type CustomerBoardcastLog struct {
	Id      int64 `json:"id"`
	Btype   int32 `json:"btype"`
	Payload any   `json:"payload"`
}

// 后台客服聊天广播事件
func (this *messageService) customerBoardcat(bid int64, btype int32, payload any) {
	if bid == 0 {
		bid = utils.GenSnowId()
	}
	customerBoardcastLock.Lock()
	defer customerBoardcastLock.Unlock()

	customerBoardcastQueue = append(customerBoardcastQueue, &CustomerBoardcastLog{
		Id:      bid,
		Btype:   btype,
		Payload: payload,
	})
	if len(customerBoardcastQueue) > 300 {
		customerBoardcastQueue = customerBoardcastQueue[len(customerBoardcastQueue)-300:]
	}
}

// 获取用户当前会话信息
func (this *messageService) customerLoadSession(userid string, create bool, username ...string) (session *entity.CustomerChatSession, err error) {
	v, err, _ := customerSessionLoadSF.Do(userid, func() (interface{}, error) {
		// 查缓存
		if v, ok := customerSessionCache.Load(userid); ok {
			session := v.(*entity.CustomerChatSession)
			return session, nil
		}
		// 查mongo
		session := new(entity.CustomerChatSession)
		GetByQ(CustomerChatSessions, bson.M{"userid": userid, "etime": 0}, session)
		if session.SessionId != 0 {
			customerSessionCache.Store(userid, session)
			return session, nil
		}
		// 不新生成
		if !create {
			return nil, fmt.Errorf("用户 %s 会话未开启", userid)
		}
		// 生成新的session
		sessionId := utils.GenSnowId()
		session = &entity.CustomerChatSession{
			SessionId:  sessionId,
			SessionId2: fmt.Sprint(sessionId),
			Userid:     userid,
			Ctime:      time.Now().UnixMilli(),
		}
		if len(username) > 0 {
			session.Username = username[0]
		}

		// 新会话分配客服
		customerSeat := this.dispatchCustomerSeat()
		if customerSeat != nil {
			session.Customer = customerSeat.Customer
			session.CustomerName = customerSeat.CustomerName
		}
		err := this.customerSessionUpdate(session)
		if err != nil {
			return nil, err
		}

		customerSessionCache.Store(userid, session)
		// 广播会话
		this.customerSessionBroadcast(session.SessionId, session)
		return session, err
	})
	if err != nil {
		return
	}
	session, ok := v.(*entity.CustomerChatSession)
	if !ok {
		err = errors.New("unknown type error")
	}
	return
}

// 广播新会话
func (this *messageService) customerSessionBroadcast(id int64, session *entity.CustomerChatSession) {
	sessionRecord := &entity.CustomerChatSessionRecord{
		SessionId2:   session.SessionId2,
		Userid:       session.Userid,
		Username:     session.Username,
		Customer:     session.Customer,
		CustomerName: session.CustomerName,
		Ctime:        time.UnixMilli(session.Ctime).In(Location()).Format(utils.FORMAT),
		Etime:        session.Etime,
		Etime2:       time.UnixMilli(session.Etime).In(Location()).Format(utils.FORMAT),
	}
	if _, e := this.customerSessionStats(sessionRecord); e != nil {
		beego.Error(e)
	}
	this.customerBoardcat(id, CustomerBoardcastNewSession, sessionRecord)
}

// 更新对话 session 并同步ck
func (this *messageService) customerSessionUpdate(session *entity.CustomerChatSession) (err error) {
	if !Upsert(CustomerChatSessions, bson.M{"_id": session.SessionId}, session) {
		err = fmt.Errorf("customer session update error: %d", session.SessionId)
		beego.Error(err)
		return
	}
	// session同步ck
	sessiontCK := new(ck.CustomerChatSession)
	err = sessiontCK.SyncMarshal(session)
	return
}

// 更新对话消息 并同步ck
func (this *messageService) customerMessageUpdate(msg *entity.CustomerChatMessage) (err error) {
	if !Update(CustomerChatMessages, bson.M{"_id": msg.Id}, msg) {
		err = fmt.Errorf("customer message update error: %d", msg.Id)
		beego.Error(err)
		return
	}
	// message同步ck
	msgCK := new(ck.CustomerChatMessage)
	err = msgCK.SyncMarshal(msg)
	if err != nil {
		return
	}
	return
}

// 客服收到消息已读
func (this *messageService) CustomerMsgReplyRead(adminId string, sessionId int64, msgs []int64) (err error) {
	// var msgsI64 []int64
	// for _, msg := range msgs {
	// 	if msgI64, err := strconv.ParseInt(msg, 10, 64); err != nil {
	// 		msgsI64 = append(msgsI64, msgI64)
	// 	}
	// }
	// sessionIdI64, _ := strconv.ParseInt(sessionId, 10, 64)

	// sessionId
	var session = &entity.CustomerChatSession{}
	GetByQ(CustomerChatSessions, bson.M{"_id": sessionId}, session)
	if session.SessionId == 0 || session.Customer != adminId {
		beego.Error(fmt.Sprintf("not admin session: %s, %d, %v", adminId, sessionId, msgs))
		return
	}

	var messages []*entity.CustomerChatMessage
	ListByQ(CustomerChatMessages, bson.M{"_id": bson.M{"$in": msgs}, "session_id": sessionId}, &messages)

	var readMsgs []string
	for _, msg := range messages {
		if msg.Reply || msg.Read {
			continue
		}
		msg.Read = true
		msg.Rtime = time.Now().UnixMilli()
		msg.Reader = adminId
		readMsgs = append(readMsgs, msg.Id2)

		err = this.customerMessageUpdate(msg)
		if err != nil {
			return
		}
	}

	// 广播
	if len(readMsgs) > 0 {
		this.customerBoardcat(0, CustomerBoardcastReadPlayerMessage, readMsgs)
	}
	return
}

// 玩家收到消息已读
func (this *messageService) CustomerMsgRead(userid string, msgs []int64) (err error) {
	var messages []*entity.CustomerChatMessage
	ListByQ(CustomerChatMessages, bson.M{"_id": bson.M{"$in": msgs}}, &messages)

	var readMsgs []string
	for _, msg := range messages {
		if !msg.Reply || msg.Read {
			continue
		}
		msg.Read = true
		msg.Rtime = time.Now().UnixMilli()
		msg.Reader = userid
		readMsgs = append(readMsgs, msg.Id2)

		err = this.customerMessageUpdate(msg)
		if err != nil {
			return
		}
	}

	// 广播
	if len(readMsgs) > 0 {
		this.customerBoardcat(0, CustomerBoardcastReadReplyMessage, readMsgs)
	}
	return
}

// 客服会话玩家评分
func (this *messageService) CustomerSessionScore(userid, sessionId string, score int32) (err error) {
	sessionIdI64, err := strconv.ParseInt(sessionId, 10, 64)
	if err != nil {
		return
	}
	session := new(entity.CustomerChatSession)
	GetByQ(CustomerChatSessions, bson.M{"_id": sessionIdI64}, session)
	if session.SessionId == 0 {
		err = fmt.Errorf("session not exists: %s", sessionId)
		beego.Error(err)
		return
	}

	session.Score = score
	session.ScoreTime = time.Now().UnixMilli()
	this.customerSessionUpdate(session)
	return
}

// 客服撤回消息
func (this *messageService) CustomerMessageRevoke(adminId string, messageId int64) (err error) {
	// todo 广播
	var msg = &entity.CustomerChatMessage{}
	GetByQ(CustomerChatMessages, bson.M{"_id": messageId}, msg)
	if msg.Id == 0 {
		err = fmt.Errorf("message %d not found", messageId)
		return
	}
	// 只发送者可撤回
	if msg.Sender != adminId {
		err = errors.New("发送者才能撤回")
		return
	}
	msg.Revoke = true
	msg.RevokeTime = time.Now().UnixMilli()
	err = this.customerMessageUpdate(msg)
	if err != nil {
		return
	}

	// 通知客户端消息撤回
	err = mq.NatsPublish(mq.TopicGameCustomerRevoke, &pb.PublishCustomerMessageRevoke{
		Userid:    msg.Userid,
		MessageId: fmt.Sprint(msg.Id),
	})
	if err != nil {
		return
	}
	return
}

// 玩家发送消息请求
func (this *messageService) CustomerSend(userid, username string, ctype int32, content, fileName string, fileSize int64) (msg *entity.CustomerChatMessage, err error) {
	// 获取当前聊天 session
	session, err := this.customerLoadSession(userid, true, username)
	if err != nil {
		return
	}

	msgId := utils.GenSnowId()
	msg = &entity.CustomerChatMessage{
		Id:         msgId,
		Id2:        fmt.Sprint(msgId),
		SessionId:  session.SessionId,
		SessionId2: session.SessionId2,
		Userid:     userid,
		Reply:      false,
		Sender:     userid,
		SenderName: username,
		Ctype:      ctype,
		Content:    content,
		Filename:   fileName,
		Filesize:   fileSize,
		Read:       false,
		Ctime:      time.Now().UnixMilli(),
		Rtime:      0,
	}
	if !Insert(CustomerChatMessages, msg) {
		err = fmt.Errorf("insert user msg error: %s, %d, %s", userid, ctype, content)
		return
	}
	// message同步ck
	msgCK := new(ck.CustomerChatMessage)
	err = msgCK.SyncMarshal(msg)
	if err != nil {
		return
	}

	// 更新会话名字
	if username != "" && session.Username != username {
		session.Username = username
		this.customerSessionUpdate(session)
	}

	// 广播
	this.customerBoardcat(msg.Id, CustomerBoardcastNewMessage, &entity.CustomerChatMessageRecord{
		Id2:        msg.Id2,
		Userid:     msg.Userid,
		Reply:      msg.Reply,
		Sender:     msg.Sender,
		SenderName: msg.SenderName,
		Ctype:      msg.Ctype,
		Content:    msg.Content,
		Read:       msg.Read,
		Ctime:      time.UnixMilli(msg.Ctime).In(Location()).Format(utils.FORMAT),
		Filename:   msg.Filename,
		Filesize:   msg.Filesize,
	})
	return
}

// CustomerReply 后台客服回复消息
func (this *messageService) CustomerReply(customer, customerName, userid string, ctype int32, content string) (msg *entity.CustomerChatMessageRecord, err error) {
	// 获取当前聊天 session
	session, err := this.customerLoadSession(userid, false)
	if err != nil {
		return
	}

	// todo 判断是否分配的客服

	msgId := utils.GenSnowId()
	message := &entity.CustomerChatMessage{
		Id:         msgId,
		Id2:        fmt.Sprint(msgId),
		SessionId:  session.SessionId,
		SessionId2: session.SessionId2,
		Userid:     userid,
		Reply:      true,
		Sender:     customer,
		SenderName: customerName,
		Ctype:      ctype,
		Content:    content,
		Read:       false,
		Ctime:      time.Now().UnixMilli(),
		Rtime:      0,
	}
	if !Insert(CustomerChatMessages, message) {
		err = fmt.Errorf("insert user msg error: %s, %d, %s", userid, ctype, content)
		return
	}

	// 发布消息到游戏
	reply := &pb.CustomerMessage{
		MessageId: message.Id2,
		Reply:     message.Reply,
		Ctype:     message.Ctype,
		Content:   message.Content,
		Ctime:     time.UnixMilli(message.Ctime).In(Location()).Format(utils.FORMAT),
		Read:      message.Read,
	}
	err = mq.NatsPublish(mq.TopicGameCustomerReply, &pb.ConsumerReplyMessage{
		Userid: userid,
		Msg:    reply,
	})
	if err != nil {
		return
	}

	// message同步ck
	msgCK := new(ck.CustomerChatMessage)
	err = msgCK.SyncMarshal(message)
	if err != nil {
		return
	}

	msg = &entity.CustomerChatMessageRecord{
		Id2:        message.Id2,
		Userid:     message.Userid,
		Reply:      message.Reply,
		Sender:     message.Sender,
		SenderName: message.SenderName,
		Ctype:      message.Ctype,
		Content:    message.Content,
		Read:       message.Read,
		Ctime:      reply.Ctime,
		Filename:   message.Filename,
		Filesize:   message.Filesize,
	}
	// 广播
	this.customerBoardcat(message.Id, CustomerBoardcastNewMessage, msg)

	// 记录首次回复时间
	if session.ReplyTime == 0 {
		session.ReplyTime = message.Ctime
		this.customerSessionUpdate(session)
	}
	return
}

// CustomerHistory 后台历史会话查询
func (this *messageService) CustomerSessionHistory(adminId, sessionId, query string, self bool) (
	sessionHasMore bool, curVersion, nextSessionId, maxMessageId int64, sessions []*entity.CustomerChatSessionRecord, err error,
) {
	curVersion = customerVersion

	// 查 session
	var sql_session_filter string
	var sql_session_args []any

	var customer_filter string
	if query != "" {
		sql_session_filter += " and userid = ?"
		sql_session_args = append(sql_session_args, query)
	} else if sessionId != "" {
		sessionIdI64, e := strconv.ParseInt(sessionId, 10, 64)
		if e != nil {
			err = e
			return
		}
		customer_filter += " and id < ?"
		sql_session_args = append(sql_session_args, sessionIdI64)
	}
	if self {
		customer_filter += " and customer = ?"
		sql_session_args = append(sql_session_args, adminId)
	}
	size := 100
	sql_session := `
		select id session_id, userid, username, customer, customer_name, ctime, etime
		from game.col_customer_chat_sessions final
		where id in (
			select max(id) session_id
			from game.col_customer_chat_sessions final 
			where 1 = 1 %s
			group by userid
			order by session_id desc
		) %s
		order by id desc
		limit %d
	`
	var session_datas []map[string]any
	err = ck.Select(&session_datas, fmt.Sprintf(sql_session, sql_session_filter, customer_filter, size), sql_session_args...)
	if err != nil {
		return
	}
	sessionHasMore = len(session_datas) >= size
	// 当前页最小id
	var min_session_id int64
	if len(session_datas) > 0 {
		min_session_id = utils.ToInt64(session_datas[len(session_datas)-1]["session_id"])
	}
	nextSessionId = min_session_id

	for _, data := range session_datas {
		session_id := utils.ToInt64(data["session_id"])
		userid := data["userid"].(string)
		username := data["username"].(string)
		customer := data["customer"].(string)
		customer_name := data["customer_name"].(string)
		ctime := utils.ToInt64(data["ctime"])
		etime := utils.ToInt64(data["etime"])
		session := &entity.CustomerChatSessionRecord{
			SessionId2:   fmt.Sprint(session_id),
			Userid:       userid,
			Username:     username,
			Customer:     customer,
			CustomerName: customer_name,
			Ctime:        time.UnixMilli(ctime).In(Location()).Format(utils.FORMAT),
			Etime:        etime,
			Etime2:       time.UnixMilli(etime).In(Location()).Format(utils.FORMAT),
		}
		sessions = append(sessions, session)

		maxMessageId = max(session_id, maxMessageId)
	}

	maxMsgId, err := this.customerSessionStats(sessions...)
	if err != nil {
		return
	}
	maxMessageId = max(maxMessageId, maxMsgId)

	customerBoardcastLock.RLock()
	defer customerBoardcastLock.RUnlock()
	if len(customerBoardcastQueue) > 0 {
		id := customerBoardcastQueue[len(customerBoardcastQueue)-1].Id
		maxMessageId = max(maxMessageId, id)
	}
	return
}

func (this *messageService) customerSessionStats(sessions ...*entity.CustomerChatSessionRecord) (maxMessageId int64, err error) {
	var userIds []string
	var userIdMap = make(map[string]*entity.CustomerChatSessionRecord)
	for _, session := range sessions {
		userIdMap[session.Userid] = session
		userIds = append(userIds, session.Userid)
	}

	// 查询消息
	var msg_datas []map[string]any
	err = ck.Select(&msg_datas, `
		select id, userid, s1.session_id session_id, reply, sender, sender_name, ctype, content, ctime, read, filename, s2.unread
		from game.col_customer_chat_messages s1 final  
		join (
			select max(id) message_id, max(session_id) session_id
			from game.col_customer_chat_messages final where userid in ? and revoke = 0
			group by userid
		) s0 on s1.id = s0.message_id
		left join (
			select session_id, count(*) unread
			from game.col_customer_chat_messages final where userid in ? and revoke = 0 and reply = 0 and read = 0
			group by userid, session_id
		) s2 on s0.session_id = s2.session_id
		order by ctime desc
	`, userIds, userIds)
	if err != nil {
		return
	}
	var max_message_id int64
	for _, data := range msg_datas {
		id := utils.ToInt64(data["id"])
		userid := data["userid"].(string)
		// session_id := utils.ToInt64(data["session_id"])
		reply := utils.ToInt64(data["reply"]) == 1
		sender := data["sender"].(string)
		sender_name := data["sender_name"].(string)
		ctype := utils.ToInt64(data["ctype"])
		content := data["content"].(string)
		ctime := utils.ToInt64(data["ctime"])
		read := utils.ToInt64(data["read"]) == 1
		filename := data["filename"].(string)
		unread := utils.ToInt64(data["unread"])

		max_message_id = max(max_message_id, id)
		if session, ok := userIdMap[userid]; ok {
			session.Unread = int32(unread)
			msg := &entity.CustomerChatMessageRecord{
				Id2:        fmt.Sprint(id),
				Userid:     userid,
				Reply:      reply,
				Sender:     sender,
				SenderName: sender_name,
				Ctype:      int32(ctype),
				Content:    content,
				Read:       read,
				Ctime:      time.UnixMilli(ctime).In(Location()).Format(utils.FORMAT),
				Filename:   filename,
			}
			session.LatestMsg = msg
		}
	}
	maxMessageId = max_message_id

	// 是否在线
	onlines := PlayerService.IsUsersOnline(userIds)

	// 玩家基本信息
	var user_datas []map[string]any
	err = ck.Select(&user_datas, `select userid, vip_lv, photo from game.col_user final where userid in ?`, userIds)
	if err != nil {
		return
	}
	var useridDatas = make(map[string]map[string]any, len(user_datas))
	for _, data := range user_datas {
		userid := data["userid"].(string)
		useridDatas[userid] = data
	}

	for _, session := range sessions {
		session.Online = onlines[session.Userid]

		if data, ok := useridDatas[session.Userid]; ok {
			vip_lv := utils.ToInt64(data["vip_lv"])
			photo := data["photo"].(string)
			session.VipLv = int32(vip_lv)
			if photo != "" && strings.HasPrefix(photo, "https://") {
				session.Photo = photo
			}
		}
	}

	return
}

// CustomerMessageHistory 后台历史会话消息查询
func (this *messageService) CustomerMessageHistory(adminId, userid, sessionId, prevMessageId string) (
	hasMore bool, minMessageId int64, messages []*entity.CustomerChatMessageRecord, err error,
) {
	var sql_messages_filter string
	var sql_messages_args = []any{userid}
	if prevMessageId != "" {
		prevMessageIdI64, e := strconv.ParseInt(prevMessageId, 10, 64)
		if e != nil {
			err = e
			return
		}
		sql_messages_filter += " and id < ?"
		sql_messages_args = append(sql_messages_args, prevMessageIdI64)
	}

	pageSize := 10
	sql_messages := `
		select id, session_id, reply, sender, sender_name, ctype, content, ctime, read, filename, filesize
		from game.col_customer_chat_messages final
		where userid = ? and revoke = 0 %s
		order by id desc
		limit %d
	`
	var message_datas []map[string]any
	err = ck.Select(&message_datas, fmt.Sprintf(sql_messages, sql_messages_filter, pageSize), sql_messages_args...)
	if err != nil {
		return
	}
	hasMore = len(message_datas) >= pageSize

	var unreadMsgs = make(map[int64][]int64)
	for _, data := range message_datas {
		id := utils.ToInt64(data["id"])
		session_id := utils.ToInt64(data["session_id"])
		reply := utils.ToInt64(data["reply"]) == 1
		sender := data["sender"].(string)
		sender_name := data["sender_name"].(string)
		ctype := utils.ToInt64(data["ctype"])
		content := data["content"].(string)
		ctime := utils.ToInt64(data["ctime"])
		read := utils.ToInt64(data["read"]) == 1
		filename := data["filename"].(string)
		filesize := utils.ToInt64(data["filesize"])

		if minMessageId == 0 {
			minMessageId = id
		} else {
			minMessageId = min(minMessageId, id)
		}

		msg := &entity.CustomerChatMessageRecord{
			Id2:        fmt.Sprint(id),
			Userid:     userid,
			Reply:      reply,
			Sender:     sender,
			SenderName: sender_name,
			Ctype:      int32(ctype),
			Content:    content,
			Read:       read,
			Ctime:      time.UnixMilli(ctime).In(Location()).Format(utils.FORMAT),
			Filename:   filename,
			Filesize:   filesize,
		}
		messages = append(messages, msg)

		if !read {
			unreadMsgs[session_id] = append(unreadMsgs[session_id], id)
		}

	}
	utils.SliceReverse(messages)

	// 未读置为已读
	if len(unreadMsgs) > 0 {
		for sessionId, msgs := range unreadMsgs {
			this.CustomerMsgReplyRead(adminId, sessionId, msgs)
		}
	}
	return
}

// CustomerSubscribe  后台广播消息查询,
// 后台重启了让刷新
func (this *messageService) CustomerSubscribe(adminId string, id, version int64) (datas []any, maxId int64, refresh bool, err error) {
	if version != customerVersion {
		refresh = true
		return
	}
	maxId = id
	// session
	customerBoardcastLock.RLock()
	defer customerBoardcastLock.RUnlock()

	length := len(customerBoardcastQueue)
	for i := length - 1; i >= 0; i-- {
		log := customerBoardcastQueue[i]
		if log.Id <= id {
			break
		}
		if i == length-1 {
			maxId = log.Id
		}
		if len(datas) > 200 {
			refresh = true
			return // 缺失消息太多, 让直接刷新页面
		}
		datas = append(datas, log)
	}
	utils.SliceReverse(datas)
	return
}

// CustomerSubscribe 后台结束聊天
func (this *messageService) CustomerCloseSesion(adminId, userid string) (etime int64, err error) {
	session, err := this.customerLoadSession(userid, false)
	if err != nil {
		return
	}

	// 判断是否分配的客服
	if session.Customer != adminId {
		err = fmt.Errorf("客服%s才能结束会话", session.CustomerName)
		return
	}

	session.Etime = time.Now().UnixMilli()
	err = this.customerSessionUpdate(session)
	if err != nil {
		return
	}

	etime = session.Etime

	// 移除会话session缓存
	customerSessionCache.Delete(userid)

	// 广播会话
	this.customerSessionBroadcast(0, session)

	// 通知游戏客户端评分
	err = mq.NatsPublish(mq.TopicGameCustomerSessionOver, &pb.PublishCustomerSessionOver{
		Userid:    session.Userid,
		SessionId: session.SessionId2,
	})
	if err != nil {
		beego.Error("publish customer session over error:", err)
		return
	}
	return
}

// ListCustomerPhrases 快捷语列表
func (this *messageService) ListCustomerPhrases(adminId string) (phrases []*entity.CustomerChatPhrase, err error) {
	defer func() {
		for _, phrase := range phrases {
			phrase.Id2 = fmt.Sprint(phrase.Id)
			phrase.NextId2 = fmt.Sprint(phrase.NextId)
		}
	}()
	var dbPhrases []*entity.CustomerChatPhrase
	ListByQ(CustomerChatPhrases, bson.M{"customer": adminId}, &dbPhrases)
	if len(dbPhrases) < 2 {
		phrases = dbPhrases
		return
	}

	// 链表排序
	var idPhrasesMap = make(map[int64]*entity.CustomerChatPhrase)
	var tail *entity.CustomerChatPhrase
	for _, phrase := range dbPhrases {
		idPhrasesMap[phrase.NextId] = phrase
		if phrase.NextId == 0 {
			tail = phrase
		}
	}
	if tail == nil {
		phrases = dbPhrases
		return
	}
	phrases = append(phrases, tail)

	var prevId = tail.Id
	for {
		if prevPhrase, ok := idPhrasesMap[prevId]; ok {
			phrases = append(phrases, prevPhrase)
			prevId = prevPhrase.Id
		} else {
			break
		}
	}
	utils.SliceReverse(phrases)
	return
}

// AddCustomerPhrases 新增快捷语
func (this *messageService) AddCustomerPhrase(adminId, phraseText string) (err error) {
	tail := new(entity.CustomerChatPhrase)
	GetByQ(CustomerChatPhrases, bson.M{"customer": adminId, "next_id": 0}, tail)
	if tail.Id == 0 {
		tail = nil
	}

	nextPhraseId := utils.GenSnowId()
	if tail != nil {
		tail.NextId = nextPhraseId
		if !Update(CustomerChatPhrases, bson.M{"_id": tail.Id}, tail) {
			err = errors.New("新增失败1")
			return
		}
	}

	var phrase = &entity.CustomerChatPhrase{
		Id:       nextPhraseId,
		Phrase:   phraseText,
		Customer: adminId,
		NextId:   0,
		Ctime:    time.Now().UnixMilli(),
	}
	if !Insert(CustomerChatPhrases, phrase) {
		err = errors.New("新增失败2")
		return
	}
	return
}

// UpdateCustomerPhrase 编辑快捷语
func (this *messageService) UpdateCustomerPhrase(adminId string, id int64, phraseText string) (err error) {
	phrase := new(entity.CustomerChatPhrase)
	GetByQ(CustomerChatPhrases, bson.M{"_id": id, "customer": adminId}, phrase)
	if phrase.Id == 0 {
		err = errors.New("快捷语不存在")
		return
	}
	phrase.Phrase = phraseText
	Update(CustomerChatPhrases, bson.M{"_id": id}, phrase)
	return
}

// SortCustomerPhrase 排序快捷语
func (this *messageService) SortCustomerPhrase(adminId string, sorts [][]int64) (err error) {
	for _, sorter := range sorts {
		e := CustomerChatPhrases.Update(bson.M{"_id": sorter[0], "customer": adminId},
			bson.M{"$set": bson.M{"next_id": sorter[1]}})
		if e != nil && e.Error() != "not found" {
			err = e
			return
		}
	}
	return
}

var (
	customerSetting         *entity.CustomerSetting // 客服配置
	customerActiveSeats     []*entity.CustomerSeat  // 在线客服列表
	customerActiveSeatsInit bool                    // 在线客服列表初始化
	customerActiveSeatsLock sync.RWMutex
)

func (this *messageService) getCustomerSetting() (setting *entity.CustomerSetting) {
	if customerSetting != nil {
		setting = customerSetting
		return
	}
	setting = &entity.CustomerSetting{}
	GetByQ(CustomerSettings, bson.M{"_id": 1}, setting)
	if setting.Id != 0 {
		customerSetting = setting
		return
	}
	// 添加配置
	setting.Id = 1
	if !Upsert(CustomerSettings, bson.M{"_id": setting.Id}, setting) {
		beego.Error("upsert customer setting error")
	}
	return
}

// 初始化客服统计数据
func (this *messageService) initCustomerSeatStats(seats ...*entity.CustomerSeat) {
	if len(seats) == 0 {
		return
	}
	var customers []string
	var customerSeatMap = make(map[string]*entity.CustomerSeat)
	for _, seat := range seats {
		customers = append(customers, seat.Customer)
		customerSeatMap[seat.Customer] = seat
	}
	// 查询当前未完成对话数最少
	var unfinish_sessions []map[string]any
	err := ck.Select(&unfinish_sessions, `
		select customer, count(*) sessions
		from game.col_customer_chat_sessions final where customer in ? and etime = 0
		group by customer
	`, customers)
	if err != nil {
		beego.Error(err)
		return
	}
	for _, se := range unfinish_sessions {
		customer := se["customer"].(string)
		sessions := utils.ToInt64(se["sessions"])
		if seat, ok := customerSeatMap[customer]; ok {
			seat.ActiveSessions = int32(sessions)
		}
	}
	// 当日累积总对话数
	now := time.Now().In(Location())
	stime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	etime := stime.AddDate(0, 0, 1)
	var total_sessions []map[string]any
	err = ck.Select(&total_sessions, `
		select customer, count(*) sessions
		from game.col_customer_chat_sessions final where customer in ? and ctime >= ? and ctime < ?
		group by customer
	`, customers, stime.UnixMilli(), etime.UnixMilli())
	if err != nil {
		beego.Error(err)
		return
	}
	for _, se := range total_sessions {
		customer := se["customer"].(string)
		sessions := utils.ToInt64(se["sessions"])
		if seat, ok := customerSeatMap[customer]; ok {
			seat.TotalSessions = int32(sessions)
		}
	}
}

// 分配一个客服
func (this *messageService) dispatchCustomerSeat() (customerSeat *entity.CustomerSeat) {
	defer func() {
		if customerSeat != nil {
			customerSeat.ActiveSessions++
			customerSeat.TotalSessions++
		}
	}()

	if !customerActiveSeatsInit {
		ListByQ(CustomerSeats, bson.M{"open": true}, &customerActiveSeats)
		// 初始化座位
		this.initCustomerSeatStats(customerActiveSeats...)
		customerActiveSeatsInit = true
	}

	customerActiveSeatsLock.RLock()
	defer customerActiveSeatsLock.RUnlock()

	// 没有在线的客服
	if len(customerActiveSeats) == 0 {
		return
	}

	setting := this.getCustomerSetting()
	if setting.SeatDispatch == 3 {
		// 手动
		return
	}
	if len(customerActiveSeats) == 1 {
		customerSeat = customerActiveSeats[0]
		return
	}
	switch setting.SeatDispatch {
	case 0: // 自动
		seatI := utils.RandMN(0, len(customerActiveSeats)-1)
		customerSeat = customerActiveSeats[seatI]
		return
	case 1: // 当前未完成对话数最少
		var minSeat *entity.CustomerSeat
		for _, seat := range customerActiveSeats {
			if minSeat == nil {
				minSeat = seat
			} else {
				if seat.ActiveSessions < minSeat.ActiveSessions {
					minSeat = seat
				}
			}
		}
		customerSeat = minSeat
		return
	case 2: // 当日累积总对话数
		var minSeat *entity.CustomerSeat
		for _, seat := range customerActiveSeats {
			if minSeat == nil {
				minSeat = seat
			} else {
				if seat.TotalSessions < minSeat.TotalSessions {
					minSeat = seat
				}
			}
		}
		customerSeat = minSeat
		return
	}
	return
}

// 客服分单逻辑配置修改
func (this *messageService) CustomerSeatDispatchUpdate(adminId string, dispatch int32) (err error) {
	setting := this.getCustomerSetting()
	if setting.SeatDispatch == dispatch {
		return
	}
	setting.SeatDispatch = dispatch

	if !Upsert(CustomerSettings, bson.M{"_id": setting.Id}, setting) {
		beego.Error("upsert customer setting error")
	}
	return
}

// 客服列表查询
func (this *messageService) GetCustomerSeat(adminId string) (customer *entity.CustomerSeat, err error) {
	customer = &entity.CustomerSeat{}
	GetByQ(CustomerSeats, bson.M{"_id": adminId}, customer)
	if customer.Customer == "" {
		customer = nil
	}
	return
}

// 客服界面2全部消息
// resolved: 0全部 1未解决 2已解决
func (this *messageService) AllSessions(adminid string, page, size int, userid string, stime, etime time.Time, questionType, resolved int32) (
	total int64, sessions []*entity.CustomerChatSessionRecord, stats []map[string]any, err error,
) {
	var sql_sessions_filter string
	var sql_sessions_args []any
	if userid != "" {
		sql_sessions_filter += " and s1.userid = ?"
		sql_sessions_args = append(sql_sessions_args, userid)
	} else {
		sql_sessions_filter += " and s1.ctime between ? and ?"
		sql_sessions_args = append(sql_sessions_args, stime.UnixMilli(), etime.UnixMilli())
	}

	if questionType > 0 {
		sql_sessions_filter += " and s1.question_type = ?"
		sql_sessions_args = append(sql_sessions_args, questionType)
	}
	if resolved > 0 {
		sql_sessions_filter += " and s1.resolved = ?"
		sql_sessions_args = append(sql_sessions_args, resolved-1)
	}

	var stat_datas []map[string]any
	sql_sessions_stat := `
		select s1.question_type, count(*) sessions
		from game.col_customer_chat_sessions s1 final
		where 1 = 1 %s
		group by s1.question_type
	`
	err = ck.Select(&stat_datas, fmt.Sprintf(sql_sessions_stat, sql_sessions_filter), sql_sessions_args...)
	if err != nil {
		return
	}
	for _, data := range stat_datas {
		question_type := utils.ToInt64(data["question_type"])
		sessions := utils.ToInt64(data["sessions"])
		var name = fmt.Sprint(question_type)
		// 0.未分类 1.充值 2.提现 3.故障或疑问 4.其他
		switch question_type {
		case 0:
			name = "未分类"
		case 1:
			name = "充值"
		case 2:
			name = "提现"
		case 3:
			name = "故障或疑问"
		case 4:
			name = "其他"
		}
		stats = append(stats, map[string]any{
			"value": sessions,
			"name":  name,
		})
	}

	sql_sessions_total := `
		select count(*) total
		from game.col_customer_chat_sessions s1 final
		where 1 = 1 %s
	`
	err = ck.Select(&total, fmt.Sprintf(sql_sessions_total, sql_sessions_filter), sql_sessions_args...)
	if err != nil {
		return
	}

	sql_sessions := `
		select s1.id, s1.ctime ctime, s1.userid userid, s0.vip_lv, s0.nickname, s1.customer, s1.customer_name, s1.etime, s1.question_type, s1.resolved, s1.remark,
			(case when s1.customer == ? then 1 else 0 end) self_customer
		from game.col_user s0 final
		join game.col_customer_chat_sessions s1 final on s0.userid = s1.userid
		where 1 = 1 %s
		order by self_customer desc, s1.id desc
		limit ?, ?
	`
	sql_sessions_args = append([]any{adminid}, sql_sessions_args...)
	offset, limit := PageCalc(page, size)
	sql_sessions_args = append(sql_sessions_args, offset, limit)
	var datas []map[string]any
	err = ck.Select(&datas, fmt.Sprintf(sql_sessions, sql_sessions_filter), sql_sessions_args...)
	if err != nil {
		return
	}
	for _, data := range datas {
		id := utils.ToInt64(data["id"])
		ctime := utils.ToInt64(data["ctime"])
		userid := data["userid"].(string)
		vip_lv := utils.ToInt64(data["vip_lv"])
		nickname := data["nickname"].(string)
		customer := data["customer"].(string)
		customer_name := data["customer_name"].(string)
		etime := utils.ToInt64(data["etime"])
		question_type := utils.ToInt64(data["question_type"])
		resolved := utils.ToInt64(data["resolved"]) == 1
		remark := data["remark"].(string)

		session := &entity.CustomerChatSessionRecord{
			SessionId2:   fmt.Sprint(id),
			Userid:       userid,
			Username:     nickname,
			Customer:     customer,
			CustomerName: customer_name,
			Ctime:        time.UnixMilli(ctime).In(Location()).Format(utils.FORMAT),
			Etime:        etime,
			Etime2:       time.UnixMilli(etime).In(Location()).Format(utils.FORMAT),
			VipLv:        int32(vip_lv),
			QuestionType: int32(question_type),
			Resolved:     resolved,
			Remark:       remark,
		}
		sessions = append(sessions, session)
	}
	return
}

// 设置会话问题解决类型
func (this *messageService) SetSessionQuestionType(adminId string, sessionId int64, questionType int32) (err error) {
	session := &entity.CustomerChatSession{}
	GetByQ(CustomerChatSessions, bson.M{"_id": sessionId}, session)
	if session.SessionId == 0 {
		err = errors.New("会话不存在")
		return
	}
	session.QuestionType = questionType
	return this.customerSessionUpdate(session)
}

// 设置会话已解决
func (this *messageService) SetSessionResovled(adminId string, sessionId int64, resolved bool) (err error) {
	session := &entity.CustomerChatSession{}
	GetByQ(CustomerChatSessions, bson.M{"_id": sessionId}, session)
	if session.SessionId == 0 {
		err = errors.New("会话不存在")
		return
	}
	session.Resolved = resolved
	return this.customerSessionUpdate(session)
}

// 设置会话备注
func (this *messageService) SetSessionRemark(adminId string, sessionId int64, remark string) (err error) {
	session := &entity.CustomerChatSession{}
	GetByQ(CustomerChatSessions, bson.M{"_id": sessionId}, session)
	if session.SessionId == 0 {
		err = errors.New("会话不存在")
		return
	}
	session.Remark = remark
	return this.customerSessionUpdate(session)
}

// 客服列表查询
func (this *messageService) ListCustomerSeat(adminId string, needUsers bool) (
	customers []*entity.CustomerSeat, uncustomerUsers []*entity.SystemUserRecord, setting *entity.CustomerSetting, err error,
) {
	setting = this.getCustomerSetting()

	ListByQ(CustomerSeats, bson.M{}, &customers)

	var customerIds []string
	for _, seat := range customers {
		customerIds = append(customerIds, seat.Customer)
	}

	if needUsers {
		var m = bson.M{}
		if len(customerIds) > 0 {
			m["_id"] = bson.M{"$nin": customerIds}
		}
		users, e := UserService.ListUserByQ(m)
		if e != nil {
			err = e
			return
		}
		for _, user := range users {
			uncustomerUsers = append(uncustomerUsers, &entity.SystemUserRecord{
				Id:       user.Id,
				UserName: user.UserName,
			})
		}
	}
	return
}

// 添加客服
func (this *messageService) AddCustomerSeat(adminName, customerId string) (err error) {
	user, err := UserService.GetUser(customerId, false)
	if err != nil {
		return
	}

	var seat = &entity.CustomerSeat{}
	GetByQ(CustomerSeats, bson.M{"_id": customerId}, seat)
	if seat.Customer != "" {
		err = fmt.Errorf("客服已存在")
		return
	}

	seat.Customer = user.Id
	seat.CustomerName = user.UserName
	seat.Open = false
	seat.Ctime = time.Now().UnixMilli()
	seat.Creater = adminName
	if !Upsert(CustomerSeats, bson.M{"_id": seat.Customer}, seat) {
		err = fmt.Errorf("客服添加失败")
		return
	}
	return
}

func (this *messageService) DeleteCustomerSeat(customer string) (err error) {
	Delete(CustomerSeats, bson.M{"_id": customer})
	return
}

func (this *messageService) OpenCustomerSeat(adminId string, customers []string, open bool) (err error) {
	var seats []*entity.CustomerSeat
	ListByQ(CustomerSeats, bson.M{"_id": bson.M{"$in": customers}}, &seats)

	var diffSeats []*entity.CustomerSeat
	var diffSeatIds []string
	for _, seat := range seats {
		if seat.Customer == "" {
			err = errors.New("客服不存在")
			return
		}
		if seat.Open == open {
			// err = fmt.Errorf("客服已%s", utils.CaseElse(open, "打开", "关闭"))
			continue
		}
		seat.Open = open
		if !Update(CustomerSeats, bson.M{"_id": seat.Customer}, seat) {
			err = fmt.Errorf("操作失败")
			return
		}

		// 添加记录
		log := &entity.CustomerSeatLog{}
		if open {
			log.Id = utils.GenSnowId()
			log.Customer = seat.Customer
			log.OpenTime = time.Now().UnixMilli()
			log.OpenUser = adminId
			if !Insert(CustomerSeatLogs, log) {
				beego.Error("添加客服上线记录失败")
			}
		} else {
			GetByQ(CustomerSeatLogs, bson.M{"customer": seat.Customer, "complete": false}, log)
			if log.Id == 0 {
				beego.Error("客服上线记录查询失败")
			} else {
				log.CloseTime = time.Now().UnixMilli()
				log.ActiveTime = log.CloseTime - log.OpenTime
				log.CloseUser = adminId
				log.Complete = true
				if !Update(CustomerSeatLogs, bson.M{"_id": log.Id}, log) {
					beego.Error("更新客服下线记录失败")
				}
			}
		}
		if log.Id != 0 {
			// log 同步ck
			err := new(ck.CustomerSeatLog).SyncMarshal(log)
			if err != nil {
				beego.Error("同步客服坐席记录失败")
			}
		}

		diffSeats = append(diffSeats, seat)
		diffSeatIds = append(diffSeatIds, seat.Customer)
	}

	if len(diffSeats) == 0 {
		return
	}

	// 分配未分配对话
	defer func() {
		if open {
			var sessions []*entity.CustomerChatSession
			stime := time.Now().In(Location()).AddDate(0, 0, -3).UnixMilli()
			ListByQ(CustomerChatSessions, bson.M{"customer": bson.M{"$eq": ""}, "etime": 0, "ctime": bson.M{"$gt": stime}}, &sessions)
			if len(sessions) > 0 {
				for _, session := range sessions {
					customerSeat := this.dispatchCustomerSeat()
					if customerSeat != nil {
						session.Customer = customerSeat.Customer
						session.CustomerName = customerSeat.CustomerName

						e := this.customerSessionUpdate(session)
						if e != nil {
							beego.Error("update sesion error")
							return
						}
						// 广播会话
						this.customerSessionBroadcast(0, session)
					}
				}
			}
		}
	}()

	if !customerActiveSeatsInit {
		return
	}
	customerActiveSeatsLock.Lock()
	defer customerActiveSeatsLock.Unlock()
	if open {
		// 初始化座位
		this.initCustomerSeatStats(diffSeats...)
		customerActiveSeats = append(customerActiveSeats, diffSeats...)
	} else {
		// 关闭座位，移除在线
		var seats []*entity.CustomerSeat
		for _, seat := range customerActiveSeats {
			if !utils.SliceIn(seat.Customer, diffSeatIds...) {
				seats = append(seats, seat)
			}
		}
		customerActiveSeats = seats
	}
	return
}

// CustomerUserSessions 用户信息(会话)
func (this *messageService) CustomerUserSessions(page, size int, stime, etime time.Time, userid, customer string, status int, rank int, asc bool) (total int64, sessions []*entity.CustomerChatSessionRecord, err error) {
	// 查 session
	var sql_session_filter, status_filter string
	var sql_session_args = []any{stime.UnixMilli(), etime.UnixMilli()}
	if userid != "" {
		sql_session_filter += " and s0.userid = ?"
		sql_session_args = append(sql_session_args, userid)
	}
	if customer != "" {
		sql_session_filter += " and s0.customer = ?"
		sql_session_args = append(sql_session_args, customer)
	}
	sql_session_args = append(sql_session_args, stime.UnixMilli())
	if utils.SliceIn(status, 1, 2, 3) {
		status_filter += " and status = ?"
		sql_session_args = append(sql_session_args, status)
	}

	sql_session_total := `
		select count(*) total from (
			select s0.id session_id, s0.userid userid, s2.vip_lv, s0.username, s0.customer, s0.customer_name, s0.ctime session_ctime, s0.etime session_etime, 
				s1.id message_id, s1.ctime message_time, s1.reply, s1.read,
				(case when s0.etime > 0 then 3 else (case when s1.reply == 1 or s1.read == 1 then 2 else 1 end) end) status
			from game.col_user s2 final 
			join game.col_customer_chat_sessions s0 final on s0.userid = s2.userid
			join (
				select userid, toYYYYMMDD(toDateTime(s0.ctime / 1000)) cdate, max(s0.id) session_id
				from game.col_customer_chat_sessions s0 final
				where s0.ctime between ? and ? %s
				group by userid, cdate
			) s3 on s0.id = s3.session_id
			left join (
				select id, session_id, reply, read, ctime
				from game.col_customer_chat_messages s0 final
				join (
					select max(id) max_message_id
					from game.col_customer_chat_messages final
					where ctime > ?
					group by session_id
				) s1 on s0.id = s1.max_message_id
			) s1 on s0.id = s1.session_id
			where 1 = 1 %s
		) s0
	`
	sql_session := `
		select s3.cdate, s0.id session_id, s0.userid userid, s2.vip_lv, s0.username, s0.customer, s0.customer_name, s0.ctime session_ctime, s0.etime session_etime, 
			s1.id message_id, s1.ctime message_time, s1.ctype, s1.content, s1.filename, s1.reply, s1.read, s1.sender, s1.sender_name,
			(case when s0.etime > 0 then 3 else (case when s1.reply == 1 or s1.read == 1 then 2 else 1 end) end) status
		from game.col_user s2 final 
		join game.col_customer_chat_sessions s0 final on s0.userid = s2.userid
		join (
			select userid, toYYYYMMDD(toDateTime(s0.ctime / 1000)) cdate, max(s0.id) session_id
			from game.col_customer_chat_sessions s0 final
			where s0.ctime between ? and ? %s
			group by userid, cdate
		) s3 on s0.id = s3.session_id
		left join (
			select id, session_id, reply, read, sender, sender_name, ctime, ctype, content, filename
			from game.col_customer_chat_messages s0 final
			join (
				select max(id) max_message_id
				from game.col_customer_chat_messages final
				where ctime > ?
				group by session_id
			) s1 on s0.id = s1.max_message_id
		) s1 on s0.id = s1.session_id
		where 1 = 1 %s
		order by %s %s
		limit ?, ?
	`
	err = ck.Select(&total, fmt.Sprintf(sql_session_total, sql_session_filter, status_filter), sql_session_args...)
	if err != nil {
		return
	}
	if total == 0 {
		return
	}

	offset, limit := PageCalc(page, size)
	var orderBy = "s0.id"
	var sort = "desc"
	switch rank {
	case 1:
		orderBy = "s0.vip_lv"
	}
	if asc {
		sort = "asc"
	}
	sql_session_args = append(sql_session_args, offset, limit)
	var session_datas []map[string]any
	err = ck.Select(&session_datas, fmt.Sprintf(sql_session, sql_session_filter, status_filter, orderBy, sort), sql_session_args...)
	if err != nil {
		return
	}

	for _, data := range session_datas {
		datestr := fmt.Sprint(data["cdate"])
		session_id := utils.ToInt64(data["session_id"])
		userid := data["userid"].(string)
		username := data["username"].(string)
		vip_lv := utils.ToInt64(data["vip_lv"])
		customer := data["customer"].(string)
		customer_name := data["customer_name"].(string)
		session_ctime := utils.ToInt64(data["session_ctime"])
		session_etime := utils.ToInt64(data["session_etime"])
		message_id := utils.ToInt64(data["message_id"])
		message_time := utils.ToInt64(data["message_time"])
		ctype := utils.ToInt64(data["ctype"])
		content := data["content"].(string)
		filename := data["filename"].(string)
		reply := utils.ToInt64(data["reply"]) == 1
		read := utils.ToInt64(data["read"]) == 1
		sender := data["sender"].(string)
		sender_name := data["sender_name"].(string)
		status := utils.ToInt64(data["status"])

		_ = session_ctime
		sdate := datestr[0:4] + "-" + datestr[4:6] + "-" + datestr[6:8]
		session := &entity.CustomerChatSessionRecord{
			SessionId2:   fmt.Sprint(session_id),
			Userid:       userid,
			Username:     username,
			Customer:     customer,
			CustomerName: customer_name,
			Ctime:        sdate, // time.UnixMilli(session_ctime).In(Location()).Format(utils.FORMAT_DATE),
			Etime:        session_etime,
			Etime2:       time.UnixMilli(session_etime).In(Location()).Format(utils.FORMAT),
			Status:       int32(status),
			VipLv:        int32(vip_lv),
		}

		msg := &entity.CustomerChatMessageRecord{
			Id2:        fmt.Sprint(message_id),
			Userid:     userid,
			Reply:      reply,
			Sender:     sender,
			SenderName: sender_name,
			Ctype:      int32(ctype),
			Content:    content,
			Read:       read,
			Ctime:      time.UnixMilli(message_time).In(Location()).Format(utils.FORMAT),
			Filename:   filename,
		}
		session.LatestMsg = msg
		sessions = append(sessions, session)
	}
	return
}

// CustomerUserSessions 用户信息(会话)
func (this *messageService) CustomerSessionSet(userid, sessionId, customer, customerName string) (err error) {
	session, err := this.customerLoadSession(userid, false)
	if err != nil {
		err = errors.New("该会话已结束0")
		return
	}
	if session.SessionId2 != sessionId {
		err = errors.New("该会话已结束1")
		return
	}
	if session.Customer == customer {
		return
	}
	session.Customer = customer
	session.CustomerName = customerName
	err = this.customerSessionUpdate(session)
	if err != nil {
		return
	}

	// 广播会话
	this.customerSessionBroadcast(0, session)
	return
}

// CustomerExpStats 客服表现
func (this *messageService) CustomerExpStats(adminId string, stime, etime time.Time, userid, customer string) (
	stats []*entity.CustomerExpStat, customers []*entity.CustomerSeat, dispatchStats []map[string]any, rateStats map[string]any, err error,
) {
	// 获取客服列表
	customers, _, _, err = this.ListCustomerSeat(adminId, false)
	if err != nil {
		return
	}
	var customerMap = make(map[string]*entity.CustomerSeat)
	for _, customer := range customers {
		customerMap[customer.Customer] = customer
	}

	var statsMap = make(map[string]*entity.CustomerExpStat)
	var getStat = func(sdate uint32, customer string) (stat *entity.CustomerExpStat) {
		datestr := fmt.Sprint(sdate)
		key := fmt.Sprintf("%d-%s", sdate, customer)
		stat, ok := statsMap[key]
		if !ok {
			stat = &entity.CustomerExpStat{
				Id:       key,
				SDate:    datestr[0:4] + "-" + datestr[4:6] + "-" + datestr[6:8],
				Customer: customer,
			}
			if c, ok := customerMap[customer]; ok {
				stat.CustomerName = c.CustomerName
			} else {
				stat.CustomerName = fmt.Sprint(customer)
			}
			statsMap[key] = stat
			stats = append(stats, stat)
		}
		return
	}

	var sql_filter1, sql_filter2 string
	var sql_filter_args1, sql_filter_args2 []any
	if userid != "" {
		sql_filter1 += " and userid = ?"
		sql_filter_args1 = append(sql_filter_args1, userid)
	}
	if customer != "" {
		sql_filter2 += " and customer = ?"
		sql_filter_args2 = append(sql_filter_args2, customer)
	}

	sql_exp := `
		select toYYYYMMDD(toDateTime(ctime / 1000)) sdate, customer, count(*) sessions, 
			SUM(case when etime > 0 then 1 else 0 end) end_sessions,
			SUM(case when reply_time > 0 then 1 else 0 end) reply_sessions,
			(reply_sessions / sessions) reply_rate,
			(end_sessions / sessions) end_rate,
			AVG(case when etime > 0 then (etime - ctime) else null end) / 1000 sessions_time_avg,
			AVG(case when reply_time > 0 then (reply_time - ctime) else null end) / 1000 reply_time_avg,
			AVG(case when score > 0 then score else null end) score_avg,
			SUM(case when etime > 0 then (etime - ctime) else null end) sessions_time_ms,
			SUM(case when reply_time > 0 then (reply_time - ctime) else null end) reply_time_ms,
			SUM(case when score > 0 then score else null end) scores,
			SUM(case when score > 0 then 1 else null end) scores_times
		from game.col_customer_chat_sessions final
		where ctime between ? and ? %s %s
		group by sdate, customer
		order by sdate desc, customer
	`
	var sql_exp_datas []map[string]any
	var sql_exp_args = []any{stime.UnixMilli(), etime.UnixMilli()}
	sql_exp_args = append(sql_exp_args, sql_filter_args1...)
	sql_exp_args = append(sql_exp_args, sql_filter_args2...)
	err = ck.Select(&sql_exp_datas, fmt.Sprintf(sql_exp, sql_filter1, sql_filter2), sql_exp_args...)
	if err != nil {
		return
	}
	for _, data := range sql_exp_datas {
		sdate := data["sdate"].(uint32)
		customer := data["customer"].(string)
		sessions := utils.ToInt64(data["sessions"])
		end_sessions := utils.ToInt64(data["end_sessions"])
		reply_sessions := utils.ToInt64(data["reply_sessions"])
		reply_rate := utils.ToFloat64(data["reply_rate"])
		end_rate := utils.ToFloat64(data["end_rate"])
		sessions_time_avg := utils.ToFloat64(data["sessions_time_avg"])
		reply_time_avg := utils.ToFloat64(data["reply_time_avg"])
		score_avg := utils.ToFloat64(data["score_avg"])
		sessions_time_ms := utils.ToInt64(data["sessions_time_ms"])
		reply_time_ms := utils.ToInt64(data["reply_time_ms"])
		scores := utils.ToInt64(data["scores"])
		scores_times := utils.ToInt64(data["scores_times"])

		stat := getStat(sdate, customer)

		// stat.SeatTime = ""
		stat.Sessions = sessions
		stat.EndSessions = end_sessions
		stat.ReplyRate = fmt.Sprintf("%.2f%%", reply_rate*100)
		stat.EndRate = fmt.Sprintf("%.2f%%", end_rate*100)
		stat.SessionsTimeAvg = fmt.Sprintf("%d", int(sessions_time_avg))
		stat.ReplyTimeAvg = fmt.Sprintf("%d", int(reply_time_avg))
		if score_avg == 0 {
			stat.ScoreAvg = "--"
		} else {
			stat.ScoreAvg = fmt.Sprintf("%.2f%%", score_avg/5*100)
		}

		stat.ReplySessions = reply_sessions
		stat.SessionsTimeMs = sessions_time_ms
		stat.ReplyTimeMs = reply_time_ms
		stat.Scores = scores
		stat.ScoresTimes = scores_times
	}

	// 客服对话分配占比
	var sql_dispatch_datas []map[string]any
	sql_dispatch := `
		select customer, count(*) sessions
		from game.col_customer_chat_sessions final
		where ctime between ? and ? %s %s
		group by customer
		order by customer
	`
	err = ck.Select(&sql_dispatch_datas, fmt.Sprintf(sql_dispatch, sql_filter1, sql_filter2), sql_exp_args...)
	if err != nil {
		return
	}
	for _, data := range sql_dispatch_datas {
		customer := data["customer"].(string)
		sessions := utils.ToInt64(data["sessions"])
		var customerName string
		if c, ok := customerMap[customer]; ok {
			customerName = c.CustomerName
		} else {
			customerName = fmt.Sprintf("id: %s", customer)
		}
		dispatchStats = append(dispatchStats, map[string]any{
			"name":  customerName,
			"value": sessions,
		})
	}

	// 消息应答情况
	var sql_rate_datas []map[string]any
	sql_rate := `
		select customer, count(*) sessions, 
			SUM(case when etime > 0 then 1 else 0 end) end_sessions,
			SUM(case when reply_time > 0 then 1 else 0 end) reply_sessions,
			(reply_sessions / sessions) reply_rate,
			(end_sessions / sessions) end_rate
		from game.col_customer_chat_sessions final
		where ctime between ? and ? %s %s
		group by customer
		order by customer
	`
	err = ck.Select(&sql_rate_datas, fmt.Sprintf(sql_rate, sql_filter1, sql_filter2), sql_exp_args...)
	if err != nil {
		return
	}
	var rate_customers, reply_rate_datas, end_rate_datas []string
	for _, data := range sql_rate_datas {
		customer := data["customer"].(string)
		reply_rate := utils.ToFloat64(data["reply_rate"])
		end_rate := utils.ToFloat64(data["end_rate"])

		var customerName string
		if c, ok := customerMap[customer]; ok {
			customerName = c.CustomerName
		} else {
			customerName = fmt.Sprintf("id: %s", customer)
		}
		rate_customers = append(rate_customers, customerName)
		reply_rate_datas = append(reply_rate_datas, fmt.Sprintf("%.2f", reply_rate*100))
		end_rate_datas = append(end_rate_datas, fmt.Sprintf("%.2f", end_rate*100))
	}
	rateStats = make(map[string]any)
	rateStats["customers"] = rate_customers
	rateStats["replyRateDatas"] = reply_rate_datas
	rateStats["endRateDatas"] = end_rate_datas

	if userid == "" {
		var sql_seat_datas []map[string]any
		var sql_seat_args = []any{stime.UnixMilli(), etime.UnixMilli()}
		sql_seat_args = append(sql_seat_args, sql_filter_args2...)
		sql_seat_time := `
			select toYYYYMMDD(toDateTime(open_time / 1000)) sdate, customer,
				SUM(case when complete = 1 then close_time - open_time else 0 end) seat_time_ms
			from game.col_customer_seat_log final
			where open_time between ? and ? and complete = 1 %s
			group by sdate, customer
		`
		err = ck.Select(&sql_seat_datas, fmt.Sprintf(sql_seat_time, sql_filter2), sql_seat_args...)
		if err != nil {
			return
		}
		for _, data := range sql_seat_datas {
			sdate := data["sdate"].(uint32)
			customer := data["customer"].(string)
			seat_time_ms := utils.ToInt64(data["seat_time_ms"])

			stat := getStat(sdate, customer)
			var h, m, s int64
			h = seat_time_ms / 1000 / 60 / 60
			m = (seat_time_ms/1000/60 - h*60)
			s = (seat_time_ms/1000 - h*60*60 - m*60)
			stat.SeatTime = fmt.Sprintf("%dh %dm %ds", h, m, s)
		}
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Id > stats[j].Id
	})

	summary := this.CustomerExpStatsSummary(stime, etime, stats)
	stats = append([]*entity.CustomerExpStat{summary}, stats...)
	return
}

// CustomerExpStatsSummary 客服表现总汇
func (this *messageService) CustomerExpStatsSummary(stime, etime time.Time, stats []*entity.CustomerExpStat) (summary *entity.CustomerExpStat) {
	summary = &entity.CustomerExpStat{
		SDate:        fmt.Sprintf("%v-%v总汇", stime.Format(utils.FORMAT_DATE), etime.Format(utils.FORMAT_DATE)),
		CustomerName: "--",
		SeatTime:     "--",
	}

	for _, stat := range stats {
		summary.Sessions += stat.Sessions
		summary.EndSessions += stat.EndSessions
		summary.ReplySessions += stat.ReplySessions
		summary.SessionsTimeMs += stat.SessionsTimeMs
		summary.ReplyTimeMs += stat.ReplyTimeMs
		summary.Scores += stat.Scores
		summary.ScoresTimes += stat.ScoresTimes
	}

	summary.ReplyRate = fmt.Sprintf("%.2f%%", ComputeFloat(summary.ReplySessions, summary.Sessions)*100)
	summary.EndRate = fmt.Sprintf("%.2f%%", ComputeFloat(summary.EndSessions, summary.Sessions)*100)
	summary.SessionsTimeAvg = fmt.Sprintf("%d", int(ComputeFloat(summary.SessionsTimeMs/1000, summary.EndSessions)))
	summary.ReplyTimeAvg = fmt.Sprintf("%d", int(ComputeFloat(summary.ReplyTimeMs/1000, summary.ReplySessions)))
	if summary.Scores == 0 {
		summary.ScoreAvg = "--"
	} else {
		summary.ScoreAvg = fmt.Sprintf("%.2f%%", ComputeFloat(summary.Scores, summary.ScoresTimes)/5*100)
	}
	return
}
