package data

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

// Task 任务信息
type Task struct {
	ID      int32     `bson:"_id" json:"id"`          //unique ID
	Type    int32     `bson:"type" json:"type"`       //类型
	Name    string    `bson:"name" json:"name"`       //名称
	Count1  uint32    `bson:"count1" json:"count1"`   //任务数值1
	Count2  uint32    `bson:"count2" json:"count2"`   //任务数值2
	Diamond int64     `bson:"diamond" json:"diamond"` //钻石奖励
	Coin    int64     `bson:"coin" json:"coin"`       //金币奖励
	Del     int       `bson:"del" json:"del"`         //是否移除
	Nextid  int32     `bson:"nextid" json:"nextid"`   //下个任务
	Ctime   time.Time `bson:"ctime" json:"ctime"`     //创建时间
}

// 牌型任务
type PokerHands struct {
	Id       int32     `bson:"_id" json:"id"`            //taskID
	NextId   int32     `bson:"next_id" json:"next_id"`   //下一级taskID
	Progress int32     `bson:"progress" json:"progress"` //进度
	Rewards  [][]int32 `bson:"rewards" json:"rewards"`   //奖励 id+数量+权重
	Del      int       `bson:"del" json:"del"`           //是否移除
}

// GetTaskList 任务
func GetTaskList() []Task {
	var list []Task
	ListByQ(Tasks, bson.M{"del": 0}, &list)
	return list
}

// GetPokerHandsTaskList 任务
func GetPokerHandsList() []PokerHands {
	var list []PokerHands
	ListByQ(PokerHandss, bson.M{"del": 0}, &list)
	return list
}

// TaskInfo 玩家的任务信息
type TaskInfo struct {
	Taskid   int32     `bson:"taskid" json:"taskid"`       //unique
	TaskType int32     `bson:"task_type" json:"task_type"` //任务类型
	Prize    int32     `bson:"prize" json:"prize"`         //任务状态 1:未完成 2:可领奖 3:已领奖
	Num      uint32    `bson:"num" json:"num"`             //完成数值
	Utime    time.Time `bson:"utime" json:"utime"`         //更新时间
}

// Save 写入数据库
func (t *Task) Save() bool {
	//t.Ctime = bson.Now()
	return Upsert(Tasks, bson.M{"_id": t.ID}, t)
}

// Save 写入数据库
func (t *PokerHands) Save() bool {
	//t.Ctime = bson.Now()
	return Upsert(PokerHandss, bson.M{"_id": t.Id}, t)
}

// VBTaskInfo vip bank task info
type VBTaskInfo struct {
	Id         int32                 `bson:"taskid" json:"taskid"`             // 任务Id
	TaskTypeId []int32               `bson:"task_type_id" json:"task_type_id"` // 任务类型ID
	RewardMode int32                 `bson:"reward_mode" json:"reward_mode"`   // 奖励模式（1.VB增长、2.VB释放）
	Reward     int32                 `bson:"rewark" json:"rewark"`             // 奖励
	Prize      int32                 `bson:"prize" json:"prize"`               // 任务状态 1:未完成 2:可领奖 3:已领奖
	TaskTypes  map[int32]*VBTaskType `bson:"task_types" json:"task_types"`     // 任务类型map
}

type VBTaskType struct {
	Id       int32  `bson:"task_type_id" json:"task_type_id"` // 任务类型Id
	Name     string `bson:"name" json:"name"`                 // 任务名称
	Progress int32  `bson:"progress" json:"progress"`         // 任务进度
}
