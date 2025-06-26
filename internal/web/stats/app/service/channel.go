package service

import (
	"errors"
	"goserver/internal/web/stats/app/entity"

	"gopkg.in/mgo.v2/bson"
)

// 渠道管理
type channelService struct{}

// 查询渠道列表
func (this *channelService) GetChannelList(page, pageSize int, m bson.M) ([]entity.ChannelInfo, error) {
	var list []entity.ChannelInfo
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	err := Channels.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList(list)
	return list, err
}

func (this *channelService) chipList(list []entity.ChannelInfo) []entity.ChannelInfo {
	for k, v := range list {
		c, _ := ConvertToIndiaTime(v.Ctime.Unix())
		v.Ctime = c
		list[k] = v
	}
	return list
}

// 查询渠道列表条数
func (this *channelService) GetChannelTotal(m bson.M) (int64, error) {
	return int64(Count(Channels, m)), nil
}

// 查询渠道数据
func (this *channelService) GetChannel() map[string][]entity.ChannelInfo {
	var list []entity.ChannelInfo
	m := bson.M{"status": 1}
	ListByQ(Channels, m, &list)
	result := make(map[string][]entity.ChannelInfo)
	result["全部"] = list
	return result
}

// 查询渠道数据下拉列表
func (this *channelService) GetChannelAll(isChannel bool, channelnames []string) (map[string][]entity.PackageInfo, error) {

	list := make([]entity.PackageInfo, 0)  // 博主渠道
	alist := make([]entity.PackageInfo, 0) // 博主渠道别名
	ulist := make([]entity.PackageInfo, 0) // 博主账号
	result := make(map[string][]entity.PackageInfo)

	var clist []entity.ChannelInfo
	ChannelArr := make([]string, 0)
	UserArr := make([]string, 0)
	m := bson.M{}
	if isChannel {
		m["channel"] = bson.M{"$in": channelnames}
	}
	BloggerAccounts.Find(m).Distinct("channel", &ChannelArr)
	BloggerAccounts.Find(m).Distinct("_id", &UserArr)
	if len(ChannelArr) > 0 {
		m := bson.M{"status": 1}
		m["name"] = bson.M{"$in": ChannelArr}
		err := Channels.Find(m).All(&clist)
		if err != nil {
			return nil, err
		}
		if len(clist) > 0 {
			for _, item := range clist {
				temp := new(entity.PackageInfo)
				temp.Key = item.Name
				temp.Name = item.Name
				if item.Name1 != "" {
					temp1 := new(entity.PackageInfo)
					temp1.Key = item.Name1
					temp1.Name = item.Name1
					alist = append(alist, *temp1)
				}
				list = append(list, *temp)
			}
		}
	}
	for _, item := range UserArr {
		temp := new(entity.PackageInfo)
		temp.Key = item
		temp.Name = item
		ulist = append(ulist, *temp)
	}

	// var userlist []entity.PlayerUser
	// m := bson.M{"_id": bson.M{"$in": UserArr}}
	// err := PlayerUsers.Find(m).All(&userlist)
	// if err != nil {
	// 	return nil, err
	// }
	// if len(userlist) > 0 {
	// 	for _, item := range userlist {
	// 		temp := new(entity.PackageInfo)
	// 		temp.Key = item.Userid
	// 		temp.Name = item.Nickname
	// 		ulist = append(ulist, *temp)
	// 	}
	// }

	result["channel"] = list
	result["alias"] = alist
	result["account"] = ulist
	return result, nil
}

// 查询渠道数据别名下拉列表
func (this *channelService) GetChannelAliasAll(loginuser *entity.User) ([]*entity.PackageInfo, error) {
	channelList := make([]string, 0)
	if loginuser != nil {
		if len(loginuser.RoleList) > 0 {
			// loginuser.RoleList
			for _, item := range loginuser.RoleList {
				for _, v := range item.ChannelList {
					channelList = append(channelList, v.Name)
				}
			}
		}
	}

	var userlist []entity.ChannelInfo
	list := make([]*entity.PackageInfo, 0)
	m := bson.M{"status": 1}
	if loginuser.Id != "1" {
		m["name"] = bson.M{"$in": channelList}
	}
	m["$and"] = []bson.M{
		{"name1": bson.M{"$ne": ""}},
		{"name1": bson.M{"$ne": nil}},
	}
	err := Channels.
		Find(m).
		All(&userlist)
	if len(userlist) > 0 {
		for _, item := range userlist {
			temp := new(entity.PackageInfo)
			temp.Key = item.Name1
			temp.Name = item.Name1
			list = append(list, temp)
		}
		// newSlice := []*entity.PackageInfo{
		// 	{Key: "0", Name: "全部"},
		// }
		// list = append(newSlice, list...)
	}
	return list, err
}

// 根据id获取渠道数据
func (this *channelService) GetChannelById(id string) (*entity.ChannelInfo, error) {
	info := new(entity.ChannelInfo)
	GetByQ(Channels, id, info)
	if info.Id == "" {
		return info, errors.New("渠道数据不存在")
	}
	return info, nil
}

// 根据渠道名称获取渠道数据
func (this *channelService) GetChannelByName(id string) (*entity.ChannelInfo, error) {
	info := new(entity.ChannelInfo)
	GetByQ(Channels, bson.M{"name": id}, info)
	if info.Id == "" {
		return info, errors.New("渠道数据不存在")
	}
	return info, nil
}

// 根据渠道别名获取渠道数据
func (this *channelService) GetChannelByNameAlias(name1 string) ([]entity.ChannelInfo, error) {
	var list []entity.ChannelInfo
	m := bson.M{}
	m["name1"] = name1
	m["status"] = 1
	err := Channels.Find(m).All(&list)
	return list, err
}

// 新增|修改 渠道数据
func (this *channelService) AddOrUpdateChannel(res *entity.ChannelInfo) error {
	if res.Id != "" {
		m := bson.M{"_id": res.Id}
		n := bson.M{
			"name":       res.Name,
			"name1":      res.Name1,
			"secret_key": res.SecretKey,
			"status":     res.Status,
			"operator":   res.Operator,
			"ctime":      bson.Now(),
		}
		if Update(Channels, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + res.Id)
	} else {
		// 新增
		res.Id = bson.NewObjectId().Hex()
		res.Ctime = bson.Now()
		if !Insert(Channels, res) {
			return errors.New("写入失败:" + res.Id)
		}
		return nil
	}

}

// 删除渠道数据
func (this *channelService) DelChannel(id string) error {
	info := new(entity.ChannelInfo)
	GetByQ(Channels, bson.M{"_id": id}, info)
	if info.Id != "" {
		if Delete(Channels, bson.M{"_id": id}) {
			return nil
		}
		return errors.New("移除失败")
	} else {
		return errors.New("未查询到可以删除的渠道数据！")
	}
}

/*
博主账号设置
*/
func (this *channelService) GetBloggerId() ([]string, error) {
	info := make([]string, 0)
	BloggerAccounts.Find(nil).Distinct("_id", &info)
	return info, nil
}

// 根据id获取渠道数据
func (this *channelService) GetBloggerById(id string) (*entity.BloggerAccount, error) {
	info := new(entity.BloggerAccount)
	GetByQ(BloggerAccounts, bson.M{"_id": id}, info)
	// GetByQ(BloggerAccounts, id, info)
	if info.Userid == "" {
		return nil, errors.New("博主账号不存在")
	}
	return info, nil
}

// 新增博主账号
func (this *channelService) AddBlogger(info *entity.BloggerAccount) error {
	info.Ctime = bson.Now()
	if !Insert(BloggerAccounts, info) {
		return errors.New("写入失败:" + info.Userid)
	}
	return nil
}

// 删除博主账号
func (this *channelService) DelBlogger(id string) error {
	info := new(entity.BloggerAccount)
	GetByQ(BloggerAccounts, bson.M{"_id": id}, info)
	if info.Userid != "" {
		if Delete(BloggerAccounts, bson.M{"_id": id}) {
			return nil
		}
		return errors.New("移除失败")
	} else {
		return errors.New("未查询到可以删除的博主账号！")
	}
}
