package config

import (
	"goserver/pkg/data"
	"sync"
)

var Beginners *sync.Map

// 充值分类
var ChargeClassify data.ChargeClassify

// 启动初始化
func InitBeginner() {
	Beginners = new(sync.Map)
	l := data.GetBeginnerList()
	for _, v := range l {
		SetBeginner(v)
	}

	ChargeClassify = data.GetChargeClassify()
}

// 启动初始化
func InitBeginner2() {
	Beginners = new(sync.Map)
}

func SetBeginner(data data.Beginner) {
	Beginners.Store(data.Id, data)
}

func GetBeginner(id int32) data.Beginner {
	if v, ok := Beginners.Load(id); ok {
		return v.(data.Beginner)
	}
	return data.Beginner{}
}

func GetBeginnerMap() map[int32]data.Beginner {
	bs := make(map[int32]data.Beginner)
	Beginners.Range(func(key, value any) bool {
		bs[key.(int32)] = value.(data.Beginner)
		return true
	})
	return bs
}

func SetChargeClassify(d data.ChargeClassify) {
	ChargeClassify = d
}

func GetChargeClassify() data.ChargeClassify {
	return ChargeClassify
}
