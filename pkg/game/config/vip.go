package config

import (
	"sync"

	"goserver/pkg/data"
)

// 商城列表
var VipMap *sync.Map

// 启动初始化
func InitVip() {
	VipMap = new(sync.Map)
	l := data.GetVipList()
	for _, v := range l {
		SetVip(v)
	}
}

// 启动初始化
func InitVip2() {
	VipMap = new(sync.Map)
}

// 同步时
func GetVips2() map[int]data.Vip {
	m := make(map[int]data.Vip)
	VipMap.Range(func(k, v interface{}) bool {
		m[k.(int)] = v.(data.Vip)
		return true
	})
	return m
}

// 客户端获取商品列表
// func GetVips() []data.Vip {
// 	list := make([]data.Vip, 0)
// 	VipMap.Range(func(k, v interface{}) bool {
// 		if val, ok := v.(data.Vip); ok {
// 			list = append(list, val)
// 		}
// 		return true
// 	})
// 	return list
// }

// 删除元素
func DelVip(k interface{}) {
	VipMap.Delete(k)
}

// 添加新的商品
func SetVip(v data.Vip) {
	VipMap.Store(v.Id, v)
}

// 获取商品
// func GetVip(id int) data.Vip {
// 	if v, ok := VipMap.Load(id); ok {
// 		return v.(data.Vip)
// 	}
// 	return data.Vip{}
// }
