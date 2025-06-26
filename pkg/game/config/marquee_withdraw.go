package config

import (
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"strconv"
	"strings"
)

// 新跑马灯配置 t_marquee_withdraw
var MarqueeWithdraws []*data.MarqueeWithdraw

func InitMarqueeWithdraws() {
	marqueeWithdraws := data.GetMarqueeWithdrawList()
	SetMarqueeWithdraws(marqueeWithdraws)
}

func SetMarqueeWithdraws(marqueeWithdraws []*data.MarqueeWithdraw) {
	MarqueeWithdraws = marqueeWithdraws

	// 解析时间范围
	for _, robot := range marqueeWithdraws {
		ranges := strings.Split(robot.TimeRange, "-")
		if len(ranges) != 2 {
			glog.Error("marqueeWithdraws time range parse error:", robot.TimeRange)
			continue
		}
		timeRanges := [2]int{}
		for i, r := range ranges {
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
		robot.TimeRanges = timeRanges
	}
}

func GetMarqueeWithdraws() []*data.MarqueeWithdraw {
	return MarqueeWithdraws
}
