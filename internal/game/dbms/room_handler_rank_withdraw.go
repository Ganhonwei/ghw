package dbms

// import (
// 	"context"
// 	"encoding/base64"
// 	"fmt"
// 	"goserver/gen/pb"
// 	"goserver/pkg/data"
// 	"goserver/pkg/game/config"
// 	"goserver/pkg/glog"
// 	"goserver/pkg/utils"
// 	"io"
// 	"strings"
// 	"time"

// 	"github.com/gogo/protobuf/proto"
// 	"github.com/redis/go-redis/v9"
// )

// // 提现排行榜时钟
// func (a *RoomActor) RankWithdrawTime() {
// 	now := utils.LocalTime()
// 	nowSec := now.Unix()
// 	if a.rankWithdrawTodayEndTime == 0 {
// 		// 更新当天结束时间戳秒
// 		a.rankWithdrawTodayEndTimeUpdate(now)
// 	}
// 	if a.rankWithdrawWeekEndTime == 0 {
// 		// 更新本周结束时间戳秒
// 		a.rankWithdrawWeekEndTimeUpdate(now)
// 	}

// 	// 一天结束
// 	if nowSec > a.rankWithdrawTodayEndTime {
// 		a.rankWithdrawTodayEndTimeUpdate(now)
// 		// 保存排行榜到昨天
// 		go rankWithdrawTodayExpire()
// 	}

// 	// 一周结束
// 	if nowSec > a.rankWithdrawWeekEndTime {
// 		a.rankWithdrawWeekEndTimeUpdate(now)
// 		// 保存排行榜到上周
// 		go rankWithdrawWeekExpire()
// 	}

// 	// 排行榜机器人
// 	go a.rankWithdrawRobotTime()
// }

// // 更新提现排行榜当天结束时间
// func (a *RoomActor) rankWithdrawTodayEndTimeUpdate(now time.Time) {
// 	if a.rankWithdrawTodayEndTime > 0 {
// 		// 下一天结束时间
// 		a.rankWithdrawTodayEndTime += int64((time.Hour * 24).Seconds())
// 	} else {
// 		todayStr := now.Format("2006-01-02")
// 		etime := utils.Str2Time(fmt.Sprintf("%s 23:59:59", todayStr))
// 		// etime := utils.Str2Time(fmt.Sprintf("%s 09:30:00", todayStr))
// 		// fmt.Printf("rank end date: %v", etime)
// 		// 当天结束时间
// 		a.rankWithdrawTodayEndTime = etime.Unix()
// 	}

// }

// // 更新提现排行榜本周六结束时间
// func (a *RoomActor) rankWithdrawWeekEndTimeUpdate(now time.Time) {
// 	// 一周结束时间,周六
// 	if a.rankWithdrawWeekEndTime > 0 {
// 		a.rankWithdrawWeekEndTime += int64((time.Hour * 24 * 7).Seconds())
// 	} else {
// 		today := now.Weekday()
// 		offset := 7 - today
// 		if today == time.Sunday {
// 			offset = 0
// 		}
// 		sunday := now.AddDate(0, 0, int(offset))
// 		// today := now.Weekday()
// 		// untilSaturday := 6 - today
// 		// saturday := now.AddDate(0, 0, int(untilSaturday))
// 		weekStr := sunday.Format("2006-01-02")
// 		etime := utils.Str2Time(fmt.Sprintf("%s 23:59:59", weekStr))
// 		a.rankWithdrawWeekEndTime = etime.Unix()
// 	}
// }

// // 获取实时提现排行榜
// func getRankWithdrawCurrent(ctx context.Context, key string, limit int64) (users []*pb.EventRankWithdraw, err error) {
// 	ranks, err := client.ZRevRangeWithScores(ctx, key, 0, limit-1).Result()
// 	if err != nil {
// 		if err == redis.Nil {
// 			glog.Warningf("rank withdraw %s not exists", key)
// 			return nil, nil
// 		}
// 		glog.Error("rank withdraw rev range err", err)
// 		return nil, err
// 	}

// 	for _, rank := range ranks {
// 		if userid, ok := rank.Member.(string); ok {
// 			user, err := client.HGet(ctx, rankWithdrawUsersKey, userid).Result()
// 			if err != nil {
// 				glog.Error("rank withdraw hget user err", err)
// 				if err == redis.Nil {
// 					glog.Warningf("rank withdraw user %s not exists", userid)
// 					continue
// 				}
// 				return nil, err
// 			}
// 			r := base64.NewDecoder(base64.StdEncoding, strings.NewReader(user))
// 			body, err := io.ReadAll(r)
// 			if err != nil {
// 				glog.Error("rank withdraw decode user err", err)
// 				return nil, err
// 			}
// 			msg := &pb.EventRankWithdraw{}
// 			err = msg.Unmarshal(body)
// 			if err != nil {
// 				glog.Error("rank withdraw unmarshal user err", err)
// 				return nil, err
// 			}
// 			msg.Amount = int32(rank.Score)
// 			users = append(users, msg)
// 		}
// 	}
// 	return
// }

// func getRankWithdrawExpired(ctx context.Context, key string) (users []*pb.EventRankWithdraw, err error) {
// 	val, err := client.Get(ctx, key).Result()
// 	if err != nil {
// 		if err != redis.Nil {
// 			glog.Error("rank withdraw passed unmarshal user err", err)
// 			return nil, err
// 		}
// 		return nil, nil
// 	}

// 	err = json.Unmarshal([]byte(val), &users)
// 	if err != nil {
// 		glog.Errorf("unmarshal yesterday err: %s", err)
// 		return
// 	}
// 	return
// }

// // 保存排行榜到昨天
// func rankWithdrawTodayExpire() {
// 	// get ranks
// 	ctx := context.Background()
// 	ranks, err := getRankWithdrawCurrent(ctx, rankWithdrawTodayKey, rankWithdrawLimit)
// 	if err != nil {
// 		return
// 	}
// 	glog.Infof("today ranks: %v", ranks)
// 	if len(ranks) == 0 {
// 		glog.Warning("today ranks empty...")
// 		return
// 	}

// 	// 重置今天排行榜
// 	_, err = client.Del(ctx, rankWithdrawTodayKey).Result()
// 	if err != nil {
// 		glog.Error("rank withdraw delete today err", err)
// 		return
// 	}

// 	// save to yesterday ranks
// 	ranksToday, err := json.Marshal(ranks)
// 	if err != nil {
// 		glog.Error("rank withdraw yesterday marshal err", err)
// 		return
// 	}
// 	_, err = client.Set(ctx, rankWithdrawYesterdayKey, string(ranksToday), 0).Result()
// 	if err != nil {
// 		glog.Error("rank withdraw hmset yesterday err", err)
// 		return
// 	}

// 	glog.Infof("save today rank withday to yesterday success: %d", len(ranks))

// 	// 清理排行榜20+后的随机人机信息
// 	rankWithdrawRandomRobotsClean()
// }

// // 保存排行榜到上周
// func rankWithdrawWeekExpire() {
// 	// get ranks
// 	ctx := context.Background()
// 	ranks, err := getRankWithdrawCurrent(ctx, rankWithdrawWeekKey, rankWithdrawLimit)
// 	if err != nil {
// 		return
// 	}
// 	glog.Infof("week ranks: %v", ranks)
// 	if len(ranks) == 0 {
// 		glog.Warning("week ranks empty...")
// 		return
// 	}

// 	// 重置本周排行榜
// 	_, err = client.Del(ctx, rankWithdrawWeekKey).Result()
// 	if err != nil {
// 		glog.Error("rank withdraw delete week err", err)
// 		return
// 	}

// 	// save to last week ranks
// 	ranksWeek, err := json.Marshal(ranks)
// 	if err != nil {
// 		glog.Error("rank withdraw week marshal err", err)
// 		return
// 	}
// 	_, err = client.Set(ctx, rankWithdrawLastWeekdayKey, string(ranksWeek), 0).Result()
// 	if err != nil {
// 		glog.Error("rank withdraw hmset last week err", err)
// 		return
// 	}
// 	glog.Infof("save week rank withday to last week success: %d", len(ranks))
// }

// // 提现排行榜机器人
// func (a *RoomActor) rankWithdrawRobotTime() {
// 	robotsTimes := a.rankWithdrawRobotTimes
// 	// robots
// 	now := utils.Timestamp()

// 	robots := config.GetRankWithdraw()
// 	var events []*pb.EventRankWithdraw
// 	vipRobotBases := make(map[int32]int32)
// 	for _, robot := range robots {
// 		// 上次提现榜触发
// 		prevTime, ok := robotsTimes[robot.Id]
// 		if !ok || prevTime == 0 {
// 			robotsTimes[robot.Id] = now
// 			continue
// 		}
// 		seconds := int64((time.Duration(robot.RankMinute) * time.Minute).Seconds())
// 		if prevTime+seconds > now {
// 			continue
// 		}
// 		// 触发提现判断
// 		robotsTimes[robot.Id] = now
// 		if !utils.RandWan(robot.WithdrawRate) {
// 			continue
// 		}
// 		idx := utils.RandIntN(len(robot.Choices))
// 		amount := robot.Choices[idx]
// 		event := &pb.EventRankWithdraw{
// 			Userid:    robot.Id,
// 			Robot:     true,
// 			RobotType: robot.RobotType,
// 			Amount:    int32(amount),
// 			VipLv:     1,
// 		}

// 		events = append(events, event)

// 		// 随机个vip等级
// 		if event.Username == "" && len(robot.VipChoices) > 0 {
// 			var choices []utils.Choice
// 			for i, lv := range robot.VipChoices {
// 				choices = append(choices, utils.Choice{Weight: robot.VipWeights[i], Item: lv})
// 			}
// 			hitLv, err := utils.WeightedChoice(choices)
// 			if err != nil {
// 				glog.Errorf("rank withdraw robot vip error: %v, %v, %v", err, robot.VipChoices, robot.VipWeights)
// 			} else {
// 				event.VipLv = int32(hitLv.Item.(int))
// 			}
// 		}
// 		vipRobotBases[event.VipLv]++

// 		// 查询固定人机信息
// 		if robot.RobotType == 0 {
// 			user, err := client.HGet(context.Background(), rankWithdrawUsersKey, robot.Id).Result()
// 			if err != nil {
// 				if err == redis.Nil {
// 					continue
// 				}
// 				glog.Error("rank withdraw hget user err", err)
// 				continue
// 			}
// 			r := base64.NewDecoder(base64.StdEncoding, strings.NewReader(user))
// 			body, err := io.ReadAll(r)
// 			if err != nil {
// 				glog.Error("rank withdraw decode user err", err)
// 				continue
// 			}
// 			msg := &pb.EventRankWithdraw{}
// 			err = msg.Unmarshal(body)
// 			if err != nil {
// 				glog.Error("rank withdraw unmarshal user err", err)
// 				continue
// 			}
// 			vipRobotBases[event.VipLv]--
// 			event.Username = msg.Username
// 			event.Avatar = msg.Avatar
// 			event.VipLv = msg.VipLv
// 		}
// 	}

// 	if len(events) == 0 {
// 		return
// 	}

// 	// gen userinfo
// 	// 查询n条人机基本信息 request to robot node
// 	req := &pb.RobotBaseGet{VipCount: vipRobotBases}
// 	r, err := nodePid.RequestFuture(req, 3*time.Second).Result()
// 	if err != nil {
// 		glog.Errorf("get future %d robot base err: %v", vipRobotBases, err)
// 		return
// 	}
// 	var vipBases map[int32]*pb.RobotBases
// 	if rsp, ok := r.(*pb.RobotBaseGeted); ok {
// 		if rsp.Error != pb.OK {
// 			glog.Errorf("fetch robot base error: %v", rsp.Error)
// 			return
// 		}
// 		vipBases = rsp.VipRobots
// 	}

// 	// 补全人机基本信息
// 	var bodys [][]byte
// 	for _, event := range events {
// 		if event.Username == "" {
// 			// && len(vipBases) > 0
// 			if bases, ok := vipBases[event.VipLv]; ok {
// 				if len(bases.Robots) <= 0 {
// 					continue
// 				}
// 				base := bases.Robots[0]
// 				bases.Robots = bases.Robots[1:]
// 				event.Username = base.Nickname
// 				event.Avatar = base.Photo
// 				// 随机的人机
// 				if event.RobotType == 1 {
// 					event.Userid = base.Id
// 				}
// 			}
// 		}
// 		body, _ := proto.Marshal(event)
// 		bodys = append(bodys, body)
// 	}

// 	producer.MultiPublish(data.TopicRankWithdraw, bodys)
// }

// // 清理除了排行榜上和固定人机外的用户信息，节约空间
// func rankWithdrawRandomRobotsClean() {
// 	// 排行榜1, 排行榜2，固定人机
// 	ctx := context.Background()
// 	yesterday, err := getRankWithdrawExpired(ctx, rankWithdrawYesterdayKey)
// 	if err != nil {
// 		glog.Error("get rank withdraw yesterday for clean err:", err)
// 		return
// 	}
// 	lastweek, err := getRankWithdrawExpired(ctx, rankWithdrawLastWeekdayKey)
// 	if err != nil {
// 		glog.Error("get rank withdraw last week for clean err:", err)
// 		return
// 	}
// 	todayRanks, err := getRankWithdrawCurrent(context.Background(), rankWithdrawTodayKey, rankWithdrawLimit+10)
// 	if err != nil {
// 		glog.Error("get rank withdraw today for clean err:", err)
// 		return
// 	}
// 	weekRanks, err := getRankWithdrawCurrent(context.Background(), rankWithdrawWeekKey, rankWithdrawLimit+10)
// 	if err != nil {
// 		glog.Error("get rank withdraw today for clean err:", err)
// 		return
// 	}
// 	keepUserids := make([]string, 0, len(yesterday)+len(lastweek)+len(todayRanks)+len(weekRanks))
// 	keepUserids = append(keepUserids, utils.SliceMapping(yesterday, func(r *pb.EventRankWithdraw) string { return r.Userid })...)
// 	keepUserids = append(keepUserids, utils.SliceMapping(lastweek, func(r *pb.EventRankWithdraw) string { return r.Userid })...)
// 	keepUserids = append(keepUserids, utils.SliceMapping(todayRanks, func(r *pb.EventRankWithdraw) string { return r.Userid })...)
// 	keepUserids = append(keepUserids, utils.SliceMapping(weekRanks, func(r *pb.EventRankWithdraw) string { return r.Userid })...)
// 	if len(keepUserids) == 0 {
// 		return
// 	}

// 	// 固定人机
// 	for _, rank := range config.GetRankWithdraw() {
// 		if rank.RobotType == 0 {
// 			keepUserids = append(keepUserids, rank.Id)
// 		}
// 	}

// 	result, err := client.HMGet(context.Background(), rankWithdrawUsersKey, keepUserids...).Result()
// 	if err != nil {
// 		if err == redis.Nil {
// 			return
// 		}
// 		glog.Error("get rank withdraw last week for clean err:", err)
// 		return
// 	}
// 	if len(result) == 0 {
// 		return
// 	}
// 	userTempKey := rankWithdrawUsersKey + "_temp"
// 	for _, item := range result {
// 		user, ok := item.(string)
// 		if !ok {
// 			glog.Error("rank withdraw clean user type err:", item)
// 			return
// 		}
// 		r := base64.NewDecoder(base64.StdEncoding, strings.NewReader(user))
// 		body, err := io.ReadAll(r)
// 		if err != nil {
// 			glog.Error("rank withdraw decode user err", err)
// 			return
// 		}
// 		msg := &pb.EventRankWithdraw{}
// 		err = msg.Unmarshal(body)
// 		if err != nil {
// 			glog.Error("rank withdraw unmarshal user err", err)
// 			return
// 		}
// 		_, err = client.HSet(ctx, userTempKey, msg.Userid, item).Result()
// 		if err != nil {
// 			glog.Error("rank withdraw clean temp set user err", err)
// 			return
// 		}
// 	}
// 	if _, err = client.Del(ctx, rankWithdrawUsersKey).Result(); err != nil {
// 		glog.Error("rank withdraw clean del set user err", err)
// 		return
// 	}
// 	if _, err = client.Rename(ctx, userTempKey, rankWithdrawUsersKey).Result(); err != nil {
// 		glog.Error("rank withdraw clean rename user set err", err)
// 		return
// 	}
// 	glog.Infof("clean rank withdraw success %d", len(result))
// }
