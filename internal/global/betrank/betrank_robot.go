package betrank

import (
	"context"
	"encoding/json"
	"fmt"
	"goserver/gen/pb"
	"goserver/gen/tb"
	"goserver/pkg/data/mq"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"math"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
)

const (
	// 当前进场的人机, tick打码 2小时打完, 23点1小时打完
	betRankHourRobotsKey = "activity:betrank:hour_robots" // 时间段进场人机
	betRankRobotBaseKey  = "activity:betrank:robot_base"  // 人机基本信息
)

var (
	enterHourRobots [24]*HourRobots // 24个时间段人机
)

type HourRobots struct {
	Robots []*BetRobot `json:"robots"` // 人机
}

type BetRobot struct {
	TurnTime int64  `json:"turnTime"` // 日时间戳
	No       int    `json:"no"`       // 名次
	Userid   string `json:"userid"`   // br_no (bet robot)
	Username string `json:"username"`
	Photo    string `json:"photo"`
	VipLv    int32  `json:"vipLv"`
	Bets     int64  `json:"bets"`    // 当日打码
	DayBets  int64  `json:"dayBets"` // 目标打码
	Bets10s  int64  `json:"bets10s"` // 每10秒打码量
}

// BetRobotBase 人机基本信息 生成后不变
type BetRobotBase struct {
	No       int    `json:"no"`
	VipLv    int32  `json:"vipLv"`
	Username string `json:"username"`
	Photo    string `json:"photo"`
}

// betrankRobotTick10s 排行榜人机 每10s执行
func (a *BetRankActor) betrankRobotTick10s(now time.Time) {
	nowSec := now.Unix()
	curHour := now.Hour()
	ctx := context.Background()

	var pubBets []*pb.PublishGameBets
	for hour := 0; hour <= curHour; hour++ {
		hr := enterHourRobots[hour]
		if hr == nil {
			var err error
			if hr, err = a.betrankHourRobotsInit(ctx, hour, curHour); err != nil && err != redis.Nil {
				glog.Errorf("robot hour init error: %d, %v", hour, err)
				continue
			}
			enterHourRobots[hour] = hr
		}
		if hr == nil || len(hr.Robots) == 0 {
			continue
		}

		for _, robot := range hr.Robots {
			// 打码量已达到
			if robot.Bets >= robot.DayBets {
				continue
			}
			bets := &pb.PublishGameBets{
				UserPid:    nil,
				Userid:     robot.Userid,
				Gtype:      0,
				Otype:      0,
				Bets:       robot.Bets10s,
				Robot:      true,
				Ts:         nowSec,
				Username:   robot.Username,
				Photo:      robot.Photo,
				RegistArea: 0,
				VipLv:      robot.VipLv,
			}
			pubBets = append(pubBets, bets)

			robot.Bets += bets.Bets // 人机打码量增加
		}

		// 保存缓存
		hrBytes, err := sonic.Marshal(hr)
		if err != nil {
			glog.Error(err)
			return
		}
		if err = rdb.HSet(ctx, betRankHourRobotsKey, fmt.Sprint(hour), string(hrBytes)).Err(); err != nil {
			glog.Error("hour robots cache error: %d, %v ", hour, err)
		}
	}

	// publish robot bets
	for _, bets := range pubBets {
		if err := mq.NatsPublish(mq.TopicGameBets, bets); err != nil {
			glog.Error("publish robot bets error: %#v, %v", bets, err)
		}
	}
}

func (a *BetRankActor) betrankHourRobotsInit(ctx context.Context, hour, curHour int) (hr *HourRobots, err error) {
	hr = &HourRobots{}
	// 停机恢复
	hrBytes, err := rdb.HGet(ctx, betRankHourRobotsKey, fmt.Sprint(hour)).Result()
	if err != nil {
		if err != redis.Nil {
			return
		}
	} else {
		if err = sonic.Unmarshal([]byte(hrBytes), hr); err != nil {
			return
		}
		return
	}

	// 时间段已过, 不生成人机了
	if hour != curHour {
		return
	}

	defer func() {
		if err != nil {
			return
		}
		hrBytes, err := sonic.Marshal(hr)
		if err != nil {
			return
		}
		if err = rdb.HSet(ctx, betRankHourRobotsKey, fmt.Sprint(hour), string(hrBytes)).Err(); err != nil {
			glog.Error("hour robots cache error: %d, %v ", hour, err)
		}
	}()

	robotsTable := table.GetTables().BetRankRobotTable.GetDataList()
	var rankLimit int32
	for _, limit := range table.GetTables().BetRankBaseTable.Get().RankLimits {
		if limit > rankLimit {
			rankLimit = limit
		}
	}
	robotsTable = robotsTable[0:int(math.Min(float64(len(robotsTable)), float64(rankLimit)))]

	// 计算入场人机数
	hourW, totalW := getRobotHourWeight(hour)
	if hourW == 0 || totalW == 0 {
		return
	}
	hourRobotCount := int(math.Floor(float64(len(robotsTable)) * (float64(hourW) / float64(totalW))))
	if hourRobotCount == 0 {
		return
	}

	// 已入场人机
	enteredNoRobots := make(map[int]bool, len(robotsTable))
	for i := 0; i < len(enterHourRobots); i++ {
		if enterHourRobots[i] != nil {
			hr := enterHourRobots[i]
			for _, robot := range hr.Robots {
				enteredNoRobots[robot.No] = true
			}
		}
	}
	// 过滤出空闲的机器人
	var freeRobots []*tb.ActivityBetRankRobotRecord
	for _, robot := range robotsTable {
		if !enteredNoRobots[int(robot.No)] {
			freeRobots = append(freeRobots, robot)
		}
	}
	// 最后一个小时剩余的一起出
	if hour == 23 {
		hourRobotCount = len(freeRobots)
	}

	// 随机选出要入场的人机
	var entryRobots []*BetRobot
	var entryRobotsNo []string
	freeRobotsIndex := len(freeRobots) - 1
	for i := 0; i < hourRobotCount; i++ {
		if i > freeRobotsIndex {
			// 已无空闲人机
			break
		}
		freeI := utils.RandMN(i, freeRobotsIndex)
		freeR := freeRobots[freeI]
		freeRobots[i], freeRobots[freeI] = freeRobots[freeI], freeRobots[i]

		// 组装人机信息
		dayBetsFloat := float64(freeR.DayBets) * (float64(utils.RandMN(int(freeR.BetsFloat[0]), int(freeR.BetsFloat[1]))) / 10000)
		dayBets := int64((freeR.DayBets + dayBetsFloat) * 100) // 配置打码量单位元
		bets10s := dayBets / (1 * 60 * 60 / 10)                // 23点入场的1小时打完
		if hour < 23 {
			bets10s *= 2 // 非23点入场2小时打完
		}
		entryRobots = append(entryRobots, &BetRobot{
			TurnTime: betRankDailyEtime,
			No:       int(freeR.No),
			Userid:   fmt.Sprintf("br_%d", freeR.No),
			Bets:     0,
			DayBets:  dayBets,
			Bets10s:  bets10s,
			// Username: "",
			// Photo:    "",
			// VipLv:    0,
		})
		entryRobotsNo = append(entryRobotsNo, fmt.Sprint(freeR.No))
	}

	// 没有空闲人机了
	if len(entryRobotsNo) == 0 {
		return
	}
	// 查询人机基本信息
	robotBases, err := rdb.HMGet(ctx, betRankRobotBaseKey, entryRobotsNo...).Result()
	if err != nil && err != redis.Nil {
		return
	}
	robotBaseMap := make(map[int]BetRobotBase)
	for _, r := range robotBases {
		if r == nil {
			continue
		}
		r := fmt.Sprint(r)
		rb := BetRobotBase{}
		if err = sonic.Unmarshal([]byte(r), &rb); err != nil {
			glog.Error(err)
			continue
		}
		robotBaseMap[rb.No] = rb
	}
	var noBaseRobots []*BetRobot
	for _, robot := range entryRobots {
		if base, ok := robotBaseMap[robot.No]; ok {
			robot.VipLv = base.VipLv
			robot.Username = base.Username
			robot.Photo = base.Photo
			hr.Robots = append(hr.Robots, robot)
		} else {
			noBaseRobots = append(noBaseRobots, robot)
		}
	}
	if len(noBaseRobots) == 0 {
		return
	}

	// 生成人机基本信息
	for _, robot := range noBaseRobots {
		// 随一个vip等级
		m, n := 1, 10
		if robot.No <= 10 {
			m = 3
		}
		vipLv := int32(utils.RandMN(m, n))
		robot.VipLv = vipLv
	}

	var newRobotBases []any
	for _, robot := range noBaseRobots {
		name, photo, sex := GetUserBase(int(robot.VipLv))
		_ = sex
		robot.Username = name
		robot.Photo = photo
		hr.Robots = append(hr.Robots, robot)

		// 人机基本信息保存
		baseBytes, err := sonic.Marshal(&BetRobotBase{
			No:       robot.No,
			VipLv:    robot.VipLv,
			Username: robot.Username,
			Photo:    robot.Photo,
		})
		if err != nil {
			glog.Error(err)
			continue
		}
		newRobotBases = append(newRobotBases, fmt.Sprint(robot.No), string(baseBytes))
	}
	if len(newRobotBases) >= 2 {
		if err = rdb.HMSet(ctx, betRankRobotBaseKey, newRobotBases...).Err(); err != nil {
			glog.Error(err)
		}
	}
	return
}

func getRobotHourWeight(hour int) (hourW, totolW int32) {
	t := table.GetTables().BetRankRobotHourTable.Get()
	totolW = t.Hour0 + t.Hour1 + t.Hour2 + t.Hour3 + t.Hour4 + t.Hour5 + t.Hour6 + t.Hour7 + t.Hour8 + t.Hour9 + t.Hour10 + t.Hour11 + t.Hour12 + t.Hour13 + t.Hour14 + t.Hour15 + t.Hour16 + t.Hour17 + t.Hour18 + t.Hour19 + t.Hour20 + t.Hour21 + t.Hour22 + t.Hour23
	switch hour {
	case 0:
		hourW = t.Hour0
	case 1:
		hourW = t.Hour1
	case 2:
		hourW = t.Hour2
	case 3:
		hourW = t.Hour3
	case 4:
		hourW = t.Hour4
	case 5:
		hourW = t.Hour5
	case 6:
		hourW = t.Hour6
	case 7:
		hourW = t.Hour7
	case 8:
		hourW = t.Hour8
	case 9:
		hourW = t.Hour9
	case 10:
		hourW = t.Hour10
	case 11:
		hourW = t.Hour11
	case 12:
		hourW = t.Hour12
	case 13:
		hourW = t.Hour13
	case 14:
		hourW = t.Hour14
	case 15:
		hourW = t.Hour15
	case 16:
		hourW = t.Hour16
	case 17:
		hourW = t.Hour17
	case 18:
		hourW = t.Hour18
	case 19:
		hourW = t.Hour19
	case 20:
		hourW = t.Hour20
	case 21:
		hourW = t.Hour21
	case 22:
		hourW = t.Hour22
	case 23:
		hourW = t.Hour23
	}
	return
}

type HeadInfo struct {
	Id   uint32
	Name string
	Sex  uint32
}

func GetUserBase(vipLv int) (string, string, uint32) {
	name := fmt.Sprintf("Player%d", utils.RandInt32N(8999999)+1000000)
	// photo := fmt.Sprintf("%d", utils.RandInt32N(photoRange)+1)
	var sex uint32 = uint32(utils.RandInt32N(2) + 1)
	if utils.RandWan(5000) {
		// 50%概率使用配置的名字
		india := false
		if utils.RandWan(2000) {
			// 剩下的人中20%使用印度名字
			india = true
		}
		if utils.RandWan(6000) {
			// 剩下的人中60%中是男性
			sex = 1

			manInNames := table.GetTables().ManInNameTable.GetDataList()
			if india && len(manInNames) > 0 {
				index := utils.RandInt32N(int32(len(manInNames)))
				name = manInNames[index].NameIn
			} else {
				manEnNames := table.GetTables().ManEnNameTable.GetDataList()
				index := utils.RandInt32N(int32(len(manEnNames)))
				name = manEnNames[index].NameEn
			}
		} else {
			sex = 2
			// 女性
			if india {
				womanInNames := table.GetTables().WomanInNameTable.GetDataList()
				index := utils.RandInt32N(int32(len(womanInNames)))
				name = womanInNames[index].NameIn
			} else {
				womanEnNames := table.GetTables().WomanEnNameTable.GetDataList()
				index := utils.RandInt32N(int32(len(womanEnNames)))
				name = womanEnNames[index].NameEn
			}
		}
	}

	var photo string
	// 80%概率使用自定义头像
	if utils.RandWan(8000) && vipLv > 0 {
		if utils.RandWan(9000) {
			// 剩下的人中60%中是男性
			info := getHead("man")
			if info != nil {
				name = info.Name
				photo = fmt.Sprintf("https://cdn.g2qdh.com/avatar/man/%d.jpg", info.Id)
			}
		} else {
			info := getHead("woman")
			if info != nil {
				name = info.Name
				photo = fmt.Sprintf("https://cdn.g2qdh.com/avatar/woman/%d.jpg", info.Id)
			}
		}
	}
	return name, photo, sex
}

func getHead(poolName string) (headInfo *HeadInfo) {
	var id int
	if poolName == "man" {
		id = utils.RandMN(0, 808)
	} else {
		id = utils.RandMN(0, 7210)
	}
	key := fmt.Sprintf("%s:hash", poolName)
	res, err := rdbHead.HGet(context.Background(), key, strconv.Itoa(id)).Result()
	if err != nil {
		glog.Errorf("get head error: %s, %d, %v", key, id, err)
		return
	}
	if res == "" {
		glog.Errorf("get head empty: %s, %d", key, id)
		return
	}

	headInfo = new(HeadInfo)
	err = json.Unmarshal([]byte(res), headInfo)
	if err != nil {
		glog.Error(err)
		return nil
	}
	return
}
