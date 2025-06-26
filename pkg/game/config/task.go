package config

import (
	"sort"
	"sync"

	"goserver/pkg/data"
)

// TaskMap 任务列表
var TaskMap *sync.Map

// TaskList 任务列表
var TaskList []data.Task

// 牌型任务
var PokerHandsMap *sync.Map

// InitTask 启动初始化
func InitTask() {
	TaskMap = new(sync.Map)
	l := data.GetTaskList()
	for _, v := range l {
		SetTask(v)
	}
	PokerHandsMap = new(sync.Map)
	p := data.GetPokerHandsList()
	for _, v := range p {
		SetPhTask(v)
	}
}

// InitTask2 启动初始化
func InitTask2() {
	TaskMap = new(sync.Map)
	PokerHandsMap = new(sync.Map)
}

// GetTasks2 同步时获取列表
func GetTasks2() map[int32]data.Task {
	m := make(map[int32]data.Task)
	TaskMap.Range(func(k, v interface{}) bool {
		m[k.(int32)] = v.(data.Task)
		return true
	})
	return m
}

// GetTasks 获取任务列表
func GetTasks() []data.Task {
	list := make([]data.Task, 0)
	TaskMap.Range(func(k, v interface{}) bool {
		if val, ok := v.(data.Task); ok {
			if val.Del > 0 {
				return false
			}
			list = append(list, val)
		}
		return true
	})
	return list
}

// GetOrderTasks 获取有序任务列表
func GetOrderTasks() []data.Task {
	return TaskList
}

// DelTask 删除元素
func DelTask(k interface{}) {
	TaskMap.Delete(k)
	sortTask()
}

// SetTask 添加或更新任务,类型做唯一key
func SetTask(v data.Task) {
	if v.Del > 0 {
		TaskMap.Delete(v.ID)
	} else {
		TaskMap.Store(v.ID, v)
	}
	sortTask()
}

// GetTask 获取指定任务
func GetTask(taskid int32) data.Task {
	if v, ok := TaskMap.Load(taskid); ok {
		return v.(data.Task)
	}
	return data.Task{}
}

func sortTask() {
	list := GetTasks()
	sort.Slice(list, func(i, j int) bool {
		return list[i].ID < list[j].ID
	})
	TaskList = list
}

// GetPhTask 添加或更新任务,类型做唯一key
func SetPhTask(v data.PokerHands) {
	if v.Del > 0 {
		PokerHandsMap.Delete(v.Id)
	} else {
		PokerHandsMap.Store(v.Id, v)
	}
	sortTask()
}

// GetPhTask 获取指定任务
func GetPhTask(taskid int32) data.PokerHands {
	if v, ok := PokerHandsMap.Load(taskid); ok {
		return v.(data.PokerHands)
	}
	return data.PokerHands{}
}

// GetPhTasks2 同步时获取列表
func GetPhTasks2() map[int32]data.PokerHands {
	m := make(map[int32]data.PokerHands)
	PokerHandsMap.Range(func(k, v interface{}) bool {
		m[k.(int32)] = v.(data.PokerHands)
		return true
	})
	return m
}

// DelPhTask 删除元素
func DelPhTask(k interface{}) {
	PokerHandsMap.Delete(k)
}
