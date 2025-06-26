package handler

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/config"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"goserver/gen/tb"

	"github.com/globalsign/mgo/bson"
)

// TaskUpdateMsg 任务变动消息
// func TaskUpdateMsg(info *data.TaskInfo) (msg *pb.TaskProgressNtf) {
// 	msg = new(pb.TaskProgressNtf)
// 	task, err := BuildTaskData(info)
// 	if err != nil {
// 		glog.Errorf("update task fail, id:%d", info.Taskid)
// 		return
// 	}
// 	msg.Task = task
// 	return
// }

// 创建新任务
func CreateTask(task data.Task) *data.TaskInfo {
	info := new(data.TaskInfo)
	info.Taskid = task.ID
	info.Prize = 1
	info.TaskType = task.Type
	return info
}

// 判断是否完成
func IsFinishTask(id int32, progress uint32) bool {
	task := config.GetTask(id)
	if task.ID == 0 {
		glog.Errorf("no find task, id:%d", id)
		return false
	}
	if task.Count2 > 0 {
		return progress >= task.Count2
	} else {
		return progress >= task.Count1
	}
}

// 构建taskData
func BuildTaskData(info *data.TaskInfo) (*pb.TaskData, error) {
	task := config.GetTask(info.Taskid)
	if task.ID == 0 {
		return nil, fmt.Errorf("no find task, id:%d", info.Taskid)
	}
	data := &pb.TaskData{
		Id:       task.ID,
		TaskType: pb.TaskType(task.Type),
		Progress: int32(info.Num),
		Status:   info.Prize,
	}
	if task.Count2 != 0 {
		data.MaxProgress = int32(task.Count2)
	} else {
		data.MaxProgress = int32(task.Count1)
	}
	item := &pb.Item{}
	if task.Diamond != 0 {
		item.Itype = 1
		item.Number = task.Diamond
	} else {
		item.Itype = 2
		item.Number = task.Coin
	}
	data.Reward = item
	return data, nil
}

// SetTaskList 配置任务数据,测试数据
func SetTaskList() {
	NewTask(1, 1, -1, "充值100元", 10000, 0, 500, 0)
	NewTask(2, 2, -1, "玩101游戏10局", 0, 20, 500, 0)
	NewTask(3, 3, -1, "赢102游戏5局", 0, 10, 500, 0)
}

// NewTask 添加新任务
func NewTask(taskid, taskType, nextid int32, name string, count1, count2 uint32,
	diamond, coin int64) {
	t := data.Task{
		//ID:      bson.NewObjectId().String(),
		ID:      taskid,
		Ctime:   bson.Now(),
		Nextid:  nextid,
		Type:    taskType,
		Name:    name,
		Count1:  count1,
		Count2:  count2,
		Diamond: diamond,
		Coin:    coin,
	}
	config.SetTask(t)
	t.Save()
}

// 创建vip bank新任务
func CreateVBTask(task *tb.VbGameTaskRecord) *data.VBTaskInfo {
	info := &data.VBTaskInfo{
		Id:         task.Id,
		TaskTypeId: task.TaskTypeId,
		RewardMode: task.RewardMode,
		Reward:     task.Rewark,
		Prize:      1,
		TaskTypes:  make(map[int32]*data.VBTaskType, len(task.TaskTypeId)),
	}
	// 任务详情
	for _, id := range task.TaskTypeId {
		taskType := table.GetTables().VBGameTaskTypeTable.Get(id)
		if taskType == nil {
			glog.Errorf("vb task type %d not exists", id)
			continue
		}
		info.TaskTypes[id] = &data.VBTaskType{
			Id:       id,
			Name:     taskType.Name,
			Progress: 0,
		}
	}
	return info
}
