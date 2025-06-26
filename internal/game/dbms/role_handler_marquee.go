package dbms

import (
	"goserver/gen/pb"
	"goserver/pkg/game/config"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

var (
	nextMarqueeWithdrawTime   int64 // 下次触发时间戳
	nextMarqueeWithdrawAmount int   // 下次触发提现金额
)

// marqueeWithdrawTime 新跑马灯tick
func (a *RoleActor) marqueeWithdrawTime() {
	defer func() {
		if r := recover(); r != nil {
			glog.Error("marqueeWithdrawTime recover error:", r)
			debug.PrintStack()
		}
	}()

	now := time.Now().In(location)

	nowTime := now.Unix()
	if nowTime < nextMarqueeWithdrawTime {
		return
	}

	// 时间到了触发提现跑马灯
	if nextMarqueeWithdrawAmount > 0 {
		// 随机个 vip 等级
		vipLv := 0
		vrChoices := config.GetVipRobotChoices()
		if len(vrChoices) > 0 {
			vip, err := utils.WeightedChoice(vrChoices)
			if err != nil {
				glog.Errorf("robot vip weight choice error: %v, %v", err, vrChoices)
			} else {
				vipLv = vip.Item.(int)
			}
		}

		// 获取robot 信息
		req := &pb.RobotBaseGet{VipCount: map[int32]int32{int32(vipLv): 1}}
		r, err := nodePid.RequestFuture(req, 3*time.Second).Result()
		if err != nil {
			glog.Errorf("get future %d robot base err: %v", req, err)
			return
		}
		var robot *pb.RobotBase
		if rsp, ok := r.(*pb.RobotBaseGeted); ok {
			if rsp.Error != pb.OK {
				glog.Errorf("fetch robot base error: %v", rsp.Error)
				return
			}
			robot = rsp.VipRobots[int32(vipLv)].Robots[0]
		}
		if robot != nil {
			photo, _ := strconv.Atoi(robot.Photo)
			marquee := &pb.MarQueeWithdrawNtf{
				Userid:   robot.Id,
				NickName: robot.Nickname,
				Photo:    int32(photo),
				VipLv:    int32(robot.VipLv),
				Amount:   int32(nextMarqueeWithdrawAmount),
			}
			rolePid.Tell(marquee)
			glog.Debugf("marquee withdraw: %#v", marquee)
		}
	}

	// 设置下一次触发
	nowSecs := now.Second() + now.Minute()*60 + now.Hour()*60*60
	// robots := config.GetMarqueeWithdraws()
	robots := table.GetTables().WithdrawMaqueeTable.GetDataList()
	for _, robot := range robots {
		timeRanges := []int{}
		Ranges := strings.Split(robot.TimeRange, "-")
		for i, r := range Ranges {
			times := strings.Split(r, ":") // 时分秒
			if len(times) != 3 {
				glog.Error("marqueeWithdraws time range parse error:", robot.TimeRange)
				continue
			}
			hour, _ := strconv.Atoi(times[0])
			minute, _ := strconv.Atoi(times[1])
			second, _ := strconv.Atoi(times[2])
			timeRanges[i] = second + minute*60 + hour*60*60
		}
		// robot.TimeRanges = timeRanges
		if !(nowSecs >= (timeRanges[0]) && nowSecs <= int(timeRanges[1])) {
			continue
		}

		nextSecond := utils.RandMN(int(robot.RandomSec[0]), int(robot.RandomSec[1]))
		nextMarqueeWithdrawTime = nowTime + int64(nextSecond)

		var choices []utils.Choice
		for i, amount := range robot.WithdrawScope {
			choices = append(choices, utils.Choice{Weight: int(robot.WithdrawWeight[i]), Item: amount})
		}
		item, err := utils.WeightedChoice(choices)
		if err != nil {
			glog.Error("choice amount error: ", err)
			return
		}
		nextMarqueeWithdrawAmount = item.Item.(int)
		return
	}
}
