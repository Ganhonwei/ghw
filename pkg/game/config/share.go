package config

import (
	"goserver/pkg/data"
	"sync"
)

// 分享活动
var ShareMap *sync.Map
var ShareAddrMap *sync.Map

// InitShare 启动初始化
func InitShare() {
	ShareMap = new(sync.Map)
	l := data.GetShareList()
	for _, v := range l {
		SetShare(v)
	}
	ShareAddrMap = new(sync.Map)
	a := data.GetShareAddrList()
	for _, v := range a {
		SetShareAddr(v)
	}
}

// InitShare2 启动初始化
func InitShare2() {
	ShareMap = new(sync.Map)
	ShareAddrMap = new(sync.Map)
}

// SetShare 添加新的分享数据
func SetShare(v data.Share) {
	ShareMap.Store(v.Id, v)
}

// SetShareAddr 添加新的分享数据
func SetShareAddr(v data.ShareAddr) {
	ShareAddrMap.Store(v.Id, v)
}

func GetShare(id int32) data.Share {
	if s, ok := ShareMap.Load(id); ok {
		if share, ok := s.(data.Share); ok {
			return share
		}
		return data.Share{}
	}
	return data.Share{}
}

func GetShareAddr(id string) data.ShareAddr {
	if s, ok := ShareAddrMap.Load(id); ok {
		if share, ok := s.(data.ShareAddr); ok {
			return share
		}
		return data.ShareAddr{}
	}
	return data.ShareAddr{}
}

func GetShareMap() map[int32]data.Share {
	shares := make(map[int32]data.Share)
	ShareMap.Range(func(key, value any) bool {
		shares[key.(int32)] = value.(data.Share)
		return true
	})
	return shares
}

func GetShareAddrMap() map[string]data.ShareAddr {
	shares := make(map[string]data.ShareAddr)
	ShareAddrMap.Range(func(key, value any) bool {
		shares[key.(string)] = value.(data.ShareAddr)
		return true
	})
	return shares
}
