package service

import (
	"errors"
	"goserver/internal/web/stats/app/entity"

	"gopkg.in/mgo.v2/bson"
)

type loggerService struct{}

// 添加IP重复记录表数据
func (this *loggerService) AddUserIpRecod(addinfo *entity.UserIpRecords) error {
	info := new(entity.UserIpRecords)
	GetByQ(IpRecords, bson.M{"ip": addinfo.Ip}, info)
	if info.Id != "" {
		// 修改
		addinfo.Ctime = bson.Now()
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"userid": addinfo.Userid,
			"ctime":  addinfo.Ctime,
		}
		if Update(IpRecords, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + addinfo.Id)
	} else {
		// 新增
		addinfo.Id = bson.NewObjectId().Hex()
		addinfo.Ctime = bson.Now()
		if !Insert(IpRecords, addinfo) {
			return errors.New("写入失败:" + addinfo.Id)
		}
		return nil
	}
}
