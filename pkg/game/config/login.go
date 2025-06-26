package config

import (
	"sync"

	"goserver/pkg/data"
)

// LoginMap 游戏列表
var LoginMap *sync.Map
var IpWhiteMap *sync.Map
var ChannelMap *sync.Map

// 停服状态下白名单
var ServerWhiteMap *sync.Map

// 设备码黑名单
var DeviceBlackMap *sync.Map

// ab分类
var RoleAB *data.RoleAB

// InitLogin 启动初始化
func InitLogin() {
	LoginMap = new(sync.Map)
	IpWhiteMap = new(sync.Map)
	ChannelMap = new(sync.Map)
	DeviceBlackMap = new(sync.Map)
	ServerWhiteMap = new(sync.Map)
	RoleAB = new(data.RoleAB)
	l := data.GetLoginPrizeList()
	for _, v := range l {
		SetLogin(v)
	}
	ip := data.GetIpWhiteList()
	for _, v := range ip {
		SetIpWhite(v.IP)
	}
	serverIp := data.GetServerWhiteList()
	for _, v := range serverIp {
		SetServerWhite(v.IP)
	}
	c := data.GetChannelList()
	for _, v := range c {
		SetChannelPackage(v)
	}
	d := data.GetDeviceIdBlackList()
	for _, v := range d {
		SetDeviceBlack(v)
	}
}

// InitLogin2 启动初始化
func InitLogin2() {
	LoginMap = new(sync.Map)
	IpWhiteMap = new(sync.Map)
	ChannelMap = new(sync.Map)
	ServerWhiteMap = new(sync.Map)
	DeviceBlackMap = new(sync.Map)
}

// GetLogins2 同步时获取列表
func GetLogins2() map[uint32]data.LoginPrize {
	m := make(map[uint32]data.LoginPrize)
	LoginMap.Range(func(k, v interface{}) bool {
		m[k.(uint32)] = v.(data.LoginPrize)
		return true
	})
	return m
}

// GetLogins 获取任务列表
func GetLogins() []data.LoginPrize {
	list := make([]data.LoginPrize, 0)
	LoginMap.Range(func(k, v interface{}) bool {
		if val, ok := v.(data.LoginPrize); ok {
			if val.Del > 0 {
				return false
			}
			list = append(list, val)
		}
		return true
	})
	return list
}

// DelLogin 删除元素
func DelLogin(k interface{}) {
	LoginMap.Delete(k)
}

// SetLogin 添加或更新任务,类型做唯一key
func SetLogin(v data.LoginPrize) {
	if v.Del > 0 {
		LoginMap.Delete(v.Day)
	} else {
		LoginMap.Store(v.Day, v)
	}
}

// GetLogin 获取指定任务
func GetLogin(day uint32) data.LoginPrize {
	if v, ok := LoginMap.Load(day); ok {
		return v.(data.LoginPrize)
	}
	return data.LoginPrize{}
}

// SetIpWhite
func SetIpWhite(ip string) {
	IpWhiteMap.Store(ip, "")
}

// DeleteIpWhite
func DeleteIpWhite(ip string) {
	IpWhiteMap.Delete(ip)
}

// GetIpWhiteByKey 同步时获取列表
func GetIpWhiteByKey(key string) bool {
	if _, ok := IpWhiteMap.Load(key); ok {
		return true
	}
	return false
}

// GetIpWhiteMap 同步时获取列表
func GetIpWhiteMap() map[string]string {
	m := make(map[string]string)
	IpWhiteMap.Range(func(k, v interface{}) bool {
		m[k.(string)] = v.(string)
		return true
	})
	return m
}

// SetChannelPackage
func SetChannelPackage(c data.ChannelPackage) {
	ChannelMap.Store(c.Id, c)
}

// DeleteIpWhite
func DeleteChannelPackage(c data.ChannelPackage) {
	ChannelMap.Delete(c.Id)
}

// GetChannelPackageByKey 同步时获取列表
func GetChannelPackageByKey(key string) data.ChannelPackage {
	c := data.ChannelPackage{}
	if c, ok := ChannelMap.Load(key); ok {
		if d, ok := c.(data.ChannelPackage); ok {
			return d
		}
	}
	return c
}

// GetLogins2 同步时获取列表
func GetChannelPackageMap() map[string]data.ChannelPackage {
	m := make(map[string]data.ChannelPackage)
	ChannelMap.Range(func(k, v interface{}) bool {
		m[k.(string)] = v.(data.ChannelPackage)
		return true
	})
	return m
}

// SetServerWhite
func SetServerWhite(ip string) {
	ServerWhiteMap.Store(ip, "")
}

// DeleteServerWhite
func DeleteServerWhite(ip string) {
	ServerWhiteMap.Delete(ip)
}

// GetServerWhiteByKey 同步时获取列表
func GetServerWhiteByKey(key string) bool {
	if _, ok := ServerWhiteMap.Load(key); ok {
		return true
	}
	return false
}

// GetServerWhiteMap 同步时获取列表
func GetServerWhiteMap() map[string]string {
	m := make(map[string]string)
	ServerWhiteMap.Range(func(k, v interface{}) bool {
		m[k.(string)] = v.(string)
		return true
	})
	return m
}

// SetDeviceBlack
func SetDeviceBlack(device data.DeviceIdLimit) {
	DeviceBlackMap.Store(device.Id, device)
}

// DeleteDeviceBlack
func DeleteDeviceBlack(Id string) {
	DeviceBlackMap.Delete(Id)
}

// GetDeviceBlackByKey 同步时获取列表
func GetDeviceBlackByKey(key string) bool {
	if _, ok := DeviceBlackMap.Load(key); ok {
		return true
	}
	return false
}

// GetDeviceBlackMap 同步时获取列表
func GetDeviceBlackMap() map[string]data.DeviceIdLimit {
	m := make(map[string]data.DeviceIdLimit)
	DeviceBlackMap.Range(func(k, v interface{}) bool {
		m[k.(string)] = v.(data.DeviceIdLimit)
		return true
	})
	return m
}
