package gate

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/config"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/table"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// 牌型活动
// func (rs *RoleActor) PokerHandsTaskReq(ctx actor.Context) {
// 	// 请求牌型活动
// 	glog.Debugf("PokerHandsTaskReq %#v", ctx.Message())
// 	rsp := new(pb.PokerHandsTaskRsp)
// 	user := rs.User
// 	if user.PhRewardId == 0 {
// 		user.PhRewardId = 1
// 	}
// 	task := config.GetPhTask(user.PhRewardId)
// 	if task.Id == 0 {
// 		rsp.Error = pb.NoFoundPhData
// 		rs.Send(rsp)
// 		return
// 	}
// 	rsp.Progress = rs.User.PhProgress
// 	rsp.MaxProgress = task.Progress
// 	rs.Send(rsp)
// }

// // 牌型任务领奖
// func (rs *RoleActor) PokerHandsRewardReq(ctx actor.Context) {
// 	glog.Debugf("PokerHandsTaskReq %#v", ctx.Message())
// 	rsp := new(pb.PokerHandsTaskGetRsp)
// 	task := config.GetPhTask(rs.User.PhRewardId)
// 	if task.Id == 0 {
// 		glog.Errorf("no found ph task, user:%s", rs.Userid)
// 		rsp.Error = pb.NoFoundPhData
// 		rs.Send(rsp)
// 		return
// 	}
// 	if rs.User.PhProgress < task.Progress {
// 		glog.Errorf("task progress exception, user:%s", rs.Userid)
// 		rsp.Error = pb.AwardFaild
// 		rs.Send(rsp)
// 		return
// 	}
// 	// random reward
// 	index, err := handler.RandomIndex(task.Rewards)
// 	if err != nil {
// 		glog.Errorf("[phtask] random reward fail, user:%s", rs.Userid)
// 		rsp.Error = pb.AwardFaild
// 		return
// 	}
// 	if task.NextId == -1 {
// 		// 最后一个的时候就减去进度再重复当前任务
// 		rs.User.PhProgress -= task.Progress
// 	} else {
// 		rs.User.PhRewardId = task.NextId
// 	}
// 	// send reward
// 	reward := task.Rewards[index]
// 	rs.sendGoodByType(uint32(reward[0]), int64(reward[1]), 0, int32(pb.LOG_TYPE66), "牌型任务")
// 	nextTask := config.GetPhTask(rs.User.PhRewardId)
// 	rsp.Progress = rs.User.PhProgress
// 	rsp.MaxProgress = nextTask.Progress
// 	rs.Send(rsp)
// }

// 获取任务
func (rs *RoleActor) GetTaskReq(ctx actor.Context) {
	glog.Debugf("GetTaskReq %#v", ctx.Message())
	rsp := new(pb.GetTaskRsp)
	user := rs.User
	if user.Task == nil {
		user.Task = make(map[int32]*data.TaskInfo)
	}
	tasks := config.GetTasks()
	for _, t := range tasks {
		info := user.Task[t.ID]
		if info == nil {
			// 没有，新建任务
			info = handler.CreateTask(t)
			user.Task[t.ID] = info
		}
		bean, err := handler.BuildTaskData(info)
		if err != nil {
			continue
		}
		rsp.Tasks = append(rsp.Tasks, bean)
	}
	rs.Send(rsp)
}

// 任务领奖
func (rs *RoleActor) GetTaskRewardReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.GetTaskRewardReq)
	glog.Debugf("GetTaskRewardReq %#v", ctx.Message())
	rs.getTaskReward(arg.TaskId)
}

func (rs *RoleActor) getTaskReward(id int32) {
	infoMap := rs.User.Task
	info := infoMap[id]
	rsp := new(pb.GetTaskRewardRsp)
	if info == nil {
		glog.Errorf("user not hava task, %s, task:%d", rs.User.Userid, id)
		rsp.Error = pb.NoFoundTaskId
		rs.Send(rsp)
		return
	}
	task := config.GetTask(id)
	if task.ID == 0 {
		glog.Errorf("config not hava task, %s, task:%d", rs.User.Userid, id)
		rsp.Error = pb.NoFoundTaskId
		rs.Send(rsp)
		return
	}
	if info.Prize != 2 {
		// 状态不对
		glog.Errorf("task can't receive reward, %s, task:%d", rs.User.Userid, id)
		rsp.Error = pb.NoFoundTaskId
		rs.Send(rsp)
		return
	}
	rs.sendGood(0, task.Coin, task.Diamond, 0, 0, 0, int32(pb.LOG_TYPE46), fmt.Sprintf("任务%d奖励", id))
	info.Prize = 3
	rs.status = true
	rsp.Task, err = handler.BuildTaskData(info)
	if err != nil {
		rsp.Error = pb.NoFoundTaskId
	}
	rs.Send(rsp)
}

// 跨天重置任务
func (rs *RoleActor) taskReset() {
	// 每日任务
	taskMap := rs.User.Task
	if taskMap == nil {
		taskMap = make(map[int32]*data.TaskInfo)
	}
	for k := range taskMap {
		delete(taskMap, k)
	}
	// todo 任务配置
	tasks := config.GetTasks()
	for _, t := range tasks {
		taskMap[t.ID] = handler.CreateTask(t)
	}
	// 牌型任务重置
	rs.User.PhRewardId = 1
	rs.User.PhProgress = 0
	rs.User.Task = taskMap
}

// 如果为空设置vip bank任务
func (rs *RoleActor) setVbTask() {
	if rs.User.VBTask == nil {
		rs.vbTaskReset()
	}
}

// vbTaskReset vip bank 跨天重置任务
func (rs *RoleActor) vbTaskReset() {
	// 每日任务
	taskMap := rs.User.VBTask
	if taskMap == nil {
		taskMap = make(map[int32]*data.VBTaskInfo)
	}
	for k := range taskMap {
		delete(taskMap, k)
	}
	// 读取当前任务配置
	for _, task := range table.GetTables().VBGameTaskTable.GetDataList() {
		taskMap[task.Id] = handler.CreateVBTask(task)
	}
	rs.User.VBTask = taskMap
	// 通知任务重置
	msg := &pb.ActivityDataNtf{
		VbTask: handler.BuildVBTaskDataMsg(rs.User),
	}
	rs.Send(msg)
}

// 计算已提现vb值(新加字段)
func (rs *RoleActor) setCashOutBounds() {
	if rs.User.CashOutBonus == 0 && len(rs.User.VBankLog) > 0 {
		var cashOutBonus int64
		for _, vb := range rs.User.VBankLog {
			if vb.LogType == int32(pb.LOG_TYPE132) {
				cashOutBonus += (-vb.Amount)
			}
		}
		if cashOutBonus > 0 {
			rs.User.CashOutBonus = cashOutBonus
			rs.status = true
		}
	}
}
