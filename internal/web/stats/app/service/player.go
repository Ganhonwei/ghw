package service

import (
	"errors"
	"fmt"
	"goserver/internal/web/stats/app/entity"

	"gopkg.in/mgo.v2/bson"
)

type playerService struct{}

// 获取
func (this *playerService) GetPlayer(userid string) (*entity.PlayerUser, error) {
	player := new(entity.PlayerUser)
	Get(PlayerUsers, userid, player)
	if player.Userid == "" {
		return nil, errors.New("用户不存在")
	}
	return player, nil
}

// 获取列表
func (this *playerService) GetList(page, pageSize int, m bson.M) ([]entity.PlayerUser, error) {
	var list []entity.PlayerUser
	if pageSize == -1 {
		pageSize = 100000
	}
	m["robot"] = false
	m["simulation_robot"] = false
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	err := PlayerUsers.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	//转换为分展示
	list = this.chipList2(list)
	return list, err
}

// 转换为分展示
func (this *playerService) chipList2(list []entity.PlayerUser) []entity.PlayerUser {
	for k, v := range list {
		t := (v.Diamond + v.ShadowDiamond)
		v.Assets = Chip2Float(int64(t))
		v.ShowAmount = Chip2Float(int64(v.Coin + v.Diamond))
		v.FDiamond = Chip2Float(int64(v.Diamond))
		v.FCoin = Chip2Float(int64(v.Coin))
		v.FShadowDiamond = Chip2Float(int64(v.ShadowDiamond))
		v.FMoney = Chip2Float(int64(v.Money))
		v.FGiveDiamond = Chip2Float(int64(v.GiveDiamond))
		v.FOutDiamond = Chip2Float(int64(v.OutDiamond))
		v.FCashOut = Chip2Float(int64(v.CashOut))
		w := (v.Diamond + v.ShadowDiamond + int64(v.CashOut)) - int64(v.Money)
		v.FWin = Chip2Float(int64(w))
		c, _ := ConvertToIndiaTime(v.Ctime.Unix())
		v.Ctime = c
		c1, _ := ConvertToIndiaTime(v.LoginTime.Unix())
		v.LoginTime = c1
		c2, _ := ConvertToIndiaTime(v.StatusTime.Unix())
		v.StatusTime = c2
		// 检测是否多账号
		m := bson.M{}
		m["userid"] = bson.M{
			"$regex":   fmt.Sprintf("\\b%s\\b", v.Userid),
			"$options": "i",
		}
		// iplist, _ := LoggerService.GetUserIpRecod(m)
		// if len(iplist) > 0 {
		// 	for _, item := range iplist {
		// 		if len(item.Userid) > 1 {
		// 			v.IsIpRepeat = true
		// 		}
		// 	}
		// }

		// if ipCount > 0 {
		// 	v.IsIpRepeat = true
		// }
		// isIp, _ := LoggerService.CheckAccountByIp(v.Userid)
		// v.IsIpRepeat = isIp

		isAppId := false
		if v.AD_ADID != "" {
			m := bson.M{
				"_id":      bson.M{"$ne": v.Userid},
				"ad__adid": v.AD_ADID,
			}
			appidCount, _ := this.GetTotal(m)
			if appidCount > 0 {
				isAppId = true
			}
		}

		v.IsAppIdRepeat = isAppId
		if v.CustomTypes != "" {
			v.UserType = v.CustomTypes
		} else {
			switch v.State {
			case 1:
				v.UserType = "新手"
			case 2:
				// 用户类型判断
				if v.Money >= 0 && v.Money <= 9999 {
					v.UserType = "零充"
				} else if v.Money >= 10000 && v.Money <= 99999 {
					v.UserType = "小R"
				} else if v.Money >= 100000 && v.Money <= 299999 {
					v.UserType = "中R"
				} else if v.Money >= 300000 && v.Money <= 1999999 {
					v.UserType = "大R"
				} else if v.Money >= 2000000 {
					v.UserType = "超大R"
				}
			case 3:
				v.UserType = "平民"
			case 4:
				v.UserType = "泡沫"
			}
		}

		list[k] = v
	}
	return list
}

// 获取总数
func (this *playerService) GetTotal(m bson.M) (int64, error) {
	m["robot"] = false
	m["simulation_robot"] = false
	return int64(Count(PlayerUsers, m)), nil
}

// 根据条件获取用户信息
func (this *playerService) GetByUserList(m bson.M) ([]entity.PlayerUser, error) {
	var list []entity.PlayerUser
	m["robot"] = false
	m["simulation_robot"] = false
	err := PlayerUsers.
		Find(m).
		All(&list)
	return list, err
}

// 根据ID获取用户信息
func (this *playerService) GetByUser(ids []string) ([]entity.PlayerUser, error) {
	var list []entity.PlayerUser
	q := bson.M{"_id": bson.M{"$in": ids}}
	err := PlayerUsers.
		Find(q).All(&list)
	return list, err
}

// 根据ID获取用户信息
func (this *playerService) GetUser(id string) (*entity.PlayerUser, error) {
	userInfo := new(entity.PlayerUser)
	Get(PlayerUsers, id, userInfo)
	if userInfo.Userid == "" {
		return nil, errors.New("用户不存在")
	}
	// userInfo = this.chipList6(userInfo)
	return userInfo, nil
}
