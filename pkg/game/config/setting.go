package config

import (
	"goserver/pkg/data"
	"strings"
	"sync"

	"gopkg.in/mgo.v2/bson"
)

var CustomerMap *sync.Map
var settingMap *sync.Map
var ShareWayMap *sync.Map

func InitSetting() {
	settingMap = new(sync.Map)
	l := data.GetSettings()
	for _, v := range l {
		SetSystemSwitch(v)
	}
	CustomerMap = new(sync.Map)
	c := data.GetCustomerAddressList()
	for _, v := range c {
		SetCustomerAddress(v)
	}
	ShareWayMap = new(sync.Map)
	s := data.GetModifyShareWayList()
	for _, v := range s {
		SetShareWay(v)
	}
}

func InitSetting2() {
	settingMap = new(sync.Map)
	CustomerMap = new(sync.Map)
	ShareWayMap = new(sync.Map)
}

// 保存功能开关
func SetSystemSwitch(s data.SystemSwitch) {
	settingMap.Store(s.Id, s)
}

func DelSystemSwitch(key string) {
	settingMap.Delete(key)
}

func GetSettingMap() map[string]data.SystemSwitch {
	res := make(map[string]data.SystemSwitch)
	settingMap.Range(func(key, value any) bool {
		res[key.(string)] = value.(data.SystemSwitch)
		return true
	})
	return res
}

func SettingIsOpen(stype, rtype int) bool {
	status := 0
	settingMap.Range(func(key, value any) bool {
		if val, ok := value.(data.SystemSwitch); ok {
			if val.Stype == stype && val.Rtype == rtype {
				status = val.Status
				return false
			}
		}
		return true
	})
	return status == 1
}

func SettingIsOpenByName(stype int, name string) bool {
	status := 0
	settingMap.Range(func(key, value any) bool {
		if val, ok := value.(data.SystemSwitch); ok {
			if val.Stype == stype && val.Name == name {
				status = val.Status
				return false
			}
		}
		return true
	})
	return status == 1
}

// 保存功能开关
func SetCustomerAddress(s data.ModifyCustomer) {
	CustomerMap.Store(s.Id, s)
}

func DelCustomerAddress(key string) {
	CustomerMap.Delete(key)
}

func GetCustomerAddressMap() map[string]data.ModifyCustomer {
	res := make(map[string]data.ModifyCustomer)
	CustomerMap.Range(func(key, value any) bool {
		res[key.(string)] = value.(data.ModifyCustomer)
		return true
	})
	return res
}

// 获取报警邮件地址
func GetalertorMailAddr(stype int32, rtype int32) []string {
	addrs := make([]string, 0)
	s := new(data.SystemSwitch)

	data.GetByQ(data.SystemSwitchs, bson.M{"stype": stype, "rtype": rtype}, &s)
	if s != nil && s.Id != "" {
		return strings.Split(s.Value, ";")
	}
	return addrs
}

// 分享配置
func SetShareWay(s data.ModifyShare) {
	ShareWayMap.Store(s.Id, s)
}

func GetShareWayMap() map[string]data.ModifyShare {
	res := make(map[string]data.ModifyShare)
	ShareWayMap.Range(func(key, value any) bool {
		res[key.(string)] = value.(data.ModifyShare)
		return true
	})
	return res
}

func DelShareWay(key string) {
	ShareWayMap.Delete(key)
}
