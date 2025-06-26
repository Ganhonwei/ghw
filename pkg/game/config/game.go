package config

import (
	"sync"

	"goserver/pkg/data"
)

// 游戏列表
var GameMap *sync.Map

// 人数配置
var PeopleMap *sync.Map

// 启动初始化
func InitGame() {
	GameMap = new(sync.Map)
	PeopleMap = new(sync.Map)
	l := data.GetGameList()
	for _, v := range l {
		SetGame(v)
	}
	p := data.GetPeopleList()
	for _, v := range p {
		SetRoomPeople(v)
	}
}

// 启动初始化
func InitGame2() {
	GameMap = new(sync.Map)
	PeopleMap = new(sync.Map)
}

// 同步时获取列表
func GetGames2() map[string]data.Game {
	m := make(map[string]data.Game)
	GameMap.Range(func(k, v interface{}) bool {
		m[k.(string)] = v.(data.Game)
		return true
	})
	return m
}

// 获取游戏列表
func GetGames() []data.Game {
	list := make([]data.Game, 0)
	GameMap.Range(func(k, v interface{}) bool {
		if val, ok := v.(data.Game); ok {
			if val.Del > 0 {
				return false
			}
			list = append(list, val)
		}
		return true
	})
	return list
}

// 删除元素
func DelGame(k interface{}) {
	GameMap.Delete(k)
}

// 添加新的公告
func SetGame(v data.Game) {
	if v.Del > 0 {
		GameMap.Delete(v.Id)
	} else {
		GameMap.Store(v.Id, v)
	}
}

// 获取
func GetGame(id string) data.Game {
	if v, ok := GameMap.Load(id); ok {
		return v.(data.Game)
	}
	return data.Game{}
}

func SetRoomPeople(v data.RoomPeople) {
	PeopleMap.Store(v.ID, v)
}

// 获取游戏人数配置
func GetRoomPeoples() []data.RoomPeople {
	list := make([]data.RoomPeople, 0)
	PeopleMap.Range(func(k, v interface{}) bool {
		if val, ok := v.(data.RoomPeople); ok {
			list = append(list, val)
		}
		return true
	})
	return list
}
