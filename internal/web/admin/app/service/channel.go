package service

import (
	"errors"
	"goserver/internal/web/admin/app/entity"

	"github.com/globalsign/mgo/bson"
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

// 查询渠道数据下拉列表
func (this *channelService) GetChannelsMap() (channelsMap map[string]entity.ChannelInfo, err error) {
	channelsMap = make(map[string]entity.ChannelInfo)

	var channels []entity.ChannelInfo
	m := bson.M{"status": 1}
	err = Channels.
		Find(m).
		All(&channels)
	if err != nil {
		return
	}
	channels = this.chipList(channels)
	for _, c := range channels {
		channelsMap[c.Name] = c
	}
	return
}

// 查询渠道列表条数
func (this *channelService) GetChannelTotal(m bson.M) (int64, error) {
	return int64(Count(Channels, m)), nil
}

// 查询渠道数据下拉列表
func (this *channelService) GetChannelAll(zero ...bool) ([]*entity.PackageInfo, error) {
	var userlist []entity.ChannelInfo
	list := make([]*entity.PackageInfo, 0)
	m := bson.M{"status": 1}
	err := Channels.
		Find(m).
		All(&userlist)
	if len(userlist) > 0 {
		for _, item := range userlist {
			temp := new(entity.PackageInfo)
			temp.Key = item.Name
			temp.Name = item.Name
			list = append(list, temp)
		}
		newSlice := []*entity.PackageInfo{}
		if len(zero) == 0 || zero[0] {
			newSlice = append(newSlice, &entity.PackageInfo{Key: "0", Name: "全部"})
		}
		list = append(newSlice, list...)

		// info := new(entity.PackageInfo)
		// info.Key = "-"
		// info.Name = "其他"
		// list = append(newSlice, info)
	}
	return list, err
}

// 查询渠道数据别名下拉列表
func (this *channelService) GetChannelAliasAll(zero ...bool) ([]*entity.PackageInfo, error) {
	var userlist []entity.ChannelInfo
	list := make([]*entity.PackageInfo, 0)
	err := Channels.
		Find(bson.M{
			"$and": []bson.M{
				{"name1": bson.M{"$ne": ""}},
				{"name1": bson.M{"$ne": nil}},
				{"status": 1},
			},
		}).
		All(&userlist)
	if len(userlist) > 0 {
		for _, item := range userlist {
			temp := new(entity.PackageInfo)
			temp.Key = item.Name1
			temp.Name = item.Name1
			list = append(list, temp)
		}
		var newSlice []*entity.PackageInfo

		if len(zero) == 0 || zero[0] {
			newSlice = append(newSlice, &entity.PackageInfo{Key: "0", Name: "全部"})
		}
		list = append(newSlice, list...)
	}
	return list, err
}

// 查询渠道类下拉列表数据
func (this *channelService) GetChannelClassAll(zero ...bool) ([]*entity.PackageInfo, error) {
	var classlist []string
	list := make([]*entity.PackageInfo, 0)
	err := Channels.
		Find(bson.M{
			"$and": []bson.M{
				{"class_name": bson.M{"$ne": ""}},
				{"class_name": bson.M{"$ne": nil}},
				{"status": 1},
			},
		}).
		Distinct("class_name", &classlist)
	if len(classlist) > 0 {
		for _, item := range classlist {
			temp := new(entity.PackageInfo)
			temp.Key = item
			temp.Name = item
			list = append(list, temp)
		}
		newSlice := []*entity.PackageInfo{}
		if len(zero) == 0 || zero[0] {
			newSlice = append(newSlice, &entity.PackageInfo{Key: "0", Name: "全部"})
		}
		list = append(newSlice, list...)
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

func (this *channelService) GetChannelByIds(packageIds []string) ([]entity.ChannelInfo, error) {
	var userlist []entity.ChannelInfo
	m := bson.M{"name": bson.M{"$in": packageIds}}
	err := Channels.
		Find(m).
		All(&userlist)
	return userlist, err
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

// 根据渠道别名列表获取渠道数据
func (this *channelService) GetChannelByNameAliasList(name1List []string) ([]entity.ChannelInfo, error) {
	var list []entity.ChannelInfo
	m := bson.M{}
	m["name1"] = bson.M{"$in": name1List}
	m["status"] = 1
	err := Channels.Find(m).All(&list)
	return list, err
}

// 根据渠道别名列表获取渠道数据
func (this *channelService) GetChannelByClassList(classList []string) ([]entity.ChannelInfo, error) {
	var list []entity.ChannelInfo
	m := bson.M{}
	m["class_name"] = bson.M{"$in": classList}
	m["status"] = 1
	err := Channels.Find(m).All(&list)
	return list, err
}

// 根据渠道别名模糊查询渠道数据
func (this *channelService) GetChannelByNameAlias1(name1 string) ([]entity.ChannelInfo, error) {
	var list []entity.ChannelInfo
	m := bson.M{}
	m["name1"] = bson.M{"$regex": name1, "$options": "i"}
	m["status"] = 1
	err := Channels.Find(m).All(&list)
	return list, err
}

// 新增|修改 渠道数据
func (this *channelService) AddOrUpdateChannel(res *entity.ChannelInfo) error {
	if res.Id != "" {
		m := bson.M{"_id": res.Id}
		n := bson.M{
			"class_name": res.ClassName,
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
