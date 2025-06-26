package betrank

import (
	"context"
	"encoding/base64"
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/data/mq"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/globalsign/mgo/bson"
	"github.com/redis/go-redis/v9"
)

// pct2float 百分比转小数
func pct2float(pct string) float64 {
	if pct == "" {
		return 0
	}
	if strings.HasSuffix(pct, "%") {
		r1 := strings.ReplaceAll(pct, "%", "")
		r2, _ := strconv.ParseFloat(r1, 64)
		return r2 / 100
	}
	r2, _ := strconv.ParseFloat(pct, 64)
	return r2
}

func float2pct(val string) string {
	if strings.HasSuffix(val, "%") {
		return val
	}
	r2, _ := strconv.ParseFloat(val, 64)
	return fmt.Sprintf("%.2f%%", r2*100)
}

// 百分分比转百分比
func mi2pct(rate int32) string {
	return fmt.Sprintf("%.2f%%", float64(rate)*100/1000000)
}

// 百万分比的值
func miShare(rate int32, value int64) int64 {
	return value * int64(rate) / 1000000
}

func requestActivityBetRankListReq(req *mq.RequestActivityBetRankListArgs) (ret *pb.ActivityBetRankRsp, err error) {
	selfUserid, selfRegistArea := req.Userid, req.RegistArea
	betsRate := table.GetTables().BetRankBaseTable.Get().RepresentBetsRate
	prizeRates := table.GetTables().BetRankPrizeRateTable.GetDataList()
	ret = &pb.ActivityBetRankRsp{
		DailyBetsRate:   float2pct(fmt.Sprint(float64(betsRate[selfRegistArea].Value[0]) / 10000)),
		WeeklyBetsRate:  float2pct(fmt.Sprint(float64(betsRate[selfRegistArea].Value[1]) / 10000)),
		MonthlyBetsRate: float2pct(fmt.Sprint(float64(betsRate[selfRegistArea].Value[2]) / 10000)),
	}
	for _, rate := range prizeRates {
		ret.PrizeRules = append(ret.PrizeRules, &pb.ActivityBetRankPrizeRule{
			Rank:    rate.Rank,
			Daily:   mi2pct(rate.Daily),
			Weekly:  mi2pct(rate.Weekly),
			Monthly: mi2pct(rate.Monthly),
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ret.Daily, err = betRankList(ctx, 1, selfUserid)
	if err != nil {
		return
	}
	ret.Weekly, err = betRankList(ctx, 2, selfUserid)
	if err != nil {
		return
	}
	ret.Monthly, err = betRankList(ctx, 3, selfUserid)
	if err != nil {
		return
	}
	return
}

// betRankList 排行榜构建
// rankType 1天2周3月
func betRankList(ctx context.Context, rankType int32, selfUserid string) (ret *pb.ActivityBetRankList, err error) {
	ret = &pb.ActivityBetRankList{
		User: &pb.ActivityBetRankSelf{
			Self: &pb.ActivityBetRankUser{
				Userid: selfUserid,
			},
		},
	}

	betRankTable := table.GetTables().BetRankBaseTable.Get()

	// 奖池数据
	if jackpot, e := getJackpot(ctx, rankType); e == nil {
		ret.Jackpot = jackpot
	}

	// 排行数据
	rankLimit, _ := utils.CaseWhen3(rankType, 1, betRankTable.RankLimits[0], 2, betRankTable.RankLimits[1], 3, betRankTable.RankLimits[2])
	betRankKey, _ := utils.CaseWhen3(rankType, 1, betRankDailyKey, 2, betRankWeeklyKey, 3, betRankMonthlyKey)
	rankUserids, err := rdb.ZRevRange(ctx, betRankKey, 0, int64(rankLimit)-1).Result()
	if err != nil || len(rankUserids) == 0 {
		return
	}
	var useridsKey []string
	var useridMap = make(map[string]*BetRankUser)
	for _, userid := range rankUserids {
		if _, ok := useridMap[userid]; !ok {
			useridMap[userid] = nil
			useridsKey = append(useridsKey, userid)
		}
	}

	// 查询用户信息
	if _, ok := useridMap[selfUserid]; !ok {
		useridsKey = append(useridsKey, selfUserid)
	}
	users, err := rdb.HMGet(ctx, betRankUsersKey, useridsKey...).Result()
	if err != nil {
		return
	}

	var selfRegistArea int // 当前用户类型
	for _, u := range users {
		user := &BetRankUser{}
		if err := sonic.Unmarshal([]byte(fmt.Sprint(u)), user); err != nil {
			continue
		}
		useridMap[user.Userid] = user
		if user.Userid == selfUserid {
			selfRegistArea = user.RegistArea
		}
	}

	// 开奖倒计时
	etime, _ := utils.CaseWhen3(rankType, 1, betRankDailyEtime, 2, betRankWeeklyEtime, 3, betRankMonthlyEtime)
	ret.Time = etime

	// 表征打码贡献率
	rates := betRankTable.RepresentBetsRate[selfRegistArea]
	rate, _ := utils.CaseWhen3(rankType, 1, rates.Value[0], 2, rates.Value[1], 3, rates.Value[2])
	ret.RepresentBetsRate = fmt.Sprintf("%.2f%%", float64(rate)/10000*100)

	// 奖金类型
	ret.PrizeType = getPrizeType(rankType, int32(selfRegistArea))

	betRankListUsers(ctx, rankType, selfUserid, rankUserids, useridMap, int64(rankLimit), ret)
	return
}

// betRankListUsers 排行榜用户排名
func betRankListUsers(ctx context.Context, rankType int32, selfUserid string, rankUserids []string, useridMap map[string]*BetRankUser, rankLimit int64, ranks *pb.ActivityBetRankList) (err error) {
	// 天排名处理
	var upNextUserId string
	ranks.User = &pb.ActivityBetRankSelf{Self: &pb.ActivityBetRankUser{}}
	if self, ok := useridMap[selfUserid]; ok && self != nil {
		// 查询最新排名
		rankDaily, rankWeekly, rankMonthly, e := getRank(ctx, selfUserid)
		if e != nil {
			err = e
			return
		}

		selfRank, _ := utils.CaseWhen3(rankType, 1, rankDaily, 2, rankWeekly, 3, rankMonthly)
		// glog.Infof("bet rank: %s, %d, %d, %d, %d, %d", selfUserid, rankType, rankDaily, rankWeekly, rankMonthly, selfRank)
		// selfBets, _ := utils.CaseWhen3(rankType, 1, self.BetsDaily, 2, self.BetsWeekly, 3, self.BetsMonthly)
		selfBets := self.GetPeriodBets(rankType)

		ranks.User.Self = &pb.ActivityBetRankUser{
			Userid:   self.Userid,
			Username: self.Username,
			Photo:    self.Photo,
			Bets:     selfBets,
			VipLv:    self.VipLv,
		}

		if selfRank != -1 && selfRank < rankLimit {
			ranks.User.Self.Rank = int32(selfRank) + 1
			ranks.User.Self.Prize, ranks.User.Self.PrizeRate = getPrize(rankType, ranks.User.Self.Rank, ranks.Jackpot)

			var rankChange int32
			prevRank := self.GetPrevRank(rankType)
			if prevRank == 0 {
				rankChange = 1 // 排名上升
			} else if (selfRank + 1) != prevRank {
				if (selfRank + 1) < prevRank {
					rankChange = 1 // 排名上升
				} else {
					rankChange = 2 // 排名下降
				}
			}
			ranks.User.Self.RankChange = rankChange
		}

		// 本评奖周期内最高达到排名
		highestReached := self.GetPeriodMaxRank(rankType)
		if highestReached != -1 && highestReached < rankLimit {
			ranks.User.HighestReached = int32(highestReached) + 1
		}

		if selfRank != 0 { // 非第一名
			if selfRank > 0 && selfRank <= int64(rankLimit) {
				// 排行榜内前一名
				upNextUserId = rankUserids[int(selfRank)-1]
			} else {
				// 排行榜外 最后一名
				upNextUserId = rankUserids[len(rankUserids)-1]
			}
		}
	} else {
		// 还没有打码, 最后一名
		upNextUserId = rankUserids[len(rankUserids)-1]
	}

	// 升一名所需打码量
	if upNextUserId != "" {
		if nextUser, ok := useridMap[upNextUserId]; ok && nextUser != nil {
			// bets, _ := utils.CaseWhen3(rankType, 1, nextUser.BetsDaily, 2, nextUser.BetsWeekly, 3, nextUser.BetsMonthly)
			bets := nextUser.GetPeriodBets(rankType)
			ranks.User.UpBetsGap = bets - ranks.User.Self.Bets
		}
	}

	for i, userid := range rankUserids {
		user, ok := useridMap[userid]
		if !ok || user == nil {
			continue
		}
		// bets, _ := utils.CaseWhen3(rankType, 1, user.BetsDaily, 2, user.BetsWeekly, 3, user.BetsMonthly)
		bets := user.GetPeriodBets(rankType)
		rank := int64(i + 1)
		rankUser := &pb.ActivityBetRankUser{
			Userid:   user.Userid,
			Rank:     int32(rank),
			Bets:     bets,
			Username: user.Username,
			Photo:    user.Photo,
			VipLv:    user.VipLv,
		}
		var rankChange int32
		prevRank := user.GetPrevRank(rankType)
		if prevRank == 0 {
			rankChange = 1 // 排名上升
		} else if rank != prevRank {
			if rank < prevRank {
				rankChange = 1 // 排名上升
			} else {
				rankChange = 2 // 排名下降
			}
		}
		rankUser.RankChange = rankChange
		rankUser.Prize, rankUser.PrizeRate = getPrize(rankType, rankUser.Rank, ranks.Jackpot)
		ranks.Ranks = append(ranks.Ranks, rankUser)
	}
	return
}

// requestActivityBetRankUpdate 排行榜更新请求
func requestActivityBetRankUpdate(req *mq.RequestActivityBetRankUpdateArgs) (rsp *pb.ActivityBetRankUpdateNtfs, err error) {
	rsp = &pb.ActivityBetRankUpdateNtfs{Ntfs: make(map[string]*pb.ActivityBetRankUpdateNtf, len(req.Userids))}
	if len(req.Userids) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rankList, err := betRankList(ctx, req.RankType, req.Userids[0])
	if err != nil {
		return
	}
	rankListBytes, err := rankList.Marshal()
	if err != nil {
		return
	}
	var userRanks = make(map[string]*pb.ActivityBetRankUser)
	for _, rank := range rankList.Ranks {
		userRanks[rank.Userid] = rank
	}
	var lastRank *pb.ActivityBetRankUser
	if len(rankList.Ranks) > 0 {
		lastRank = rankList.Ranks[len(rankList.Ranks)-1]
	}

	// selfs
	var usersMap = make(map[string]*BetRankUser, len(req.Userids))
	users, err := rdb.HMGet(ctx, betRankUsersKey, req.Userids...).Result()
	if err != nil && err != redis.Nil {
		return
	}
	for _, u := range users {
		user := &BetRankUser{}
		if err := sonic.Unmarshal([]byte(fmt.Sprint(u)), user); err != nil {
			continue
		}
		usersMap[user.Userid] = user
	}

	betRankTable := table.GetTables().BetRankBaseTable.Get()
	rankLimit, _ := utils.CaseWhen3(req.RankType, 1, betRankTable.RankLimits[0], 2, betRankTable.RankLimits[1], 3, betRankTable.RankLimits[2])

	for _, userid := range req.Userids {
		list := &pb.ActivityBetRankList{}
		if err = list.Unmarshal(rankListBytes); err != nil {
			return
		}
		list.User.Self = &pb.ActivityBetRankUser{Userid: userid}
		if user, ok := usersMap[userid]; ok {
			highestReached := int32(user.GetPeriodMaxRank(req.RankType))
			if highestReached != -1 && highestReached < rankLimit {
				list.User.HighestReached = highestReached + 1
			}
			list.User.Self.Username = user.Username
			list.User.Self.VipLv = user.VipLv
			list.User.Self.Photo = user.Photo
			bets := user.GetPeriodBets(req.RankType)
			list.User.Self.Bets = bets
		}
		if rank, ok := userRanks[userid]; ok {
			list.User.Self.Rank = rank.Rank
			list.User.Self.Bets = rank.Bets
			list.User.Self.Prize = rank.Prize
			list.User.Self.PrizeRate = rank.PrizeRate
			list.User.Self.RankChange = rank.RankChange
		}

		if lastRank != nil {
			selfRank := list.User.Self.Rank
			if selfRank == 0 {
				// 排行榜开外
				list.User.UpBetsGap = lastRank.Bets - list.User.Self.Bets
			} else if selfRank > 1 && // 非第一名
				len(rankList.Ranks) >= int(selfRank-1) {
				list.User.UpBetsGap = rankList.Ranks[selfRank-2].Bets - list.User.Self.Bets
			}
		}

		rsp.Ntfs[userid] = &pb.ActivityBetRankUpdateNtf{RankType: req.RankType, List: list}
	}
	return
}

// requestActivityBetRankSkipWindow 弹窗 skip for today 设置
func requestActivityBetRankSkipWindow(req *mq.RequestActivityBetRankSkipWindowArgs) (rsp any, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	window := &BetRankWindow{}
	r, e := rdb.HGet(ctx, betRankWindowsKey, req.Userid).Result()
	if e != nil {
		if e != redis.Nil {
			err = e
		}
		return
	}
	if err = sonic.Unmarshal([]byte(r), window); err != nil {
		return
	}

	switch req.WType {
	case 1:
		window.Skip1 = true
		rsp = &pb.ActivityBetRankNewRankSkipTodayRsp{}
	case 2:
		window.Skip2 = true
		rsp = &pb.ActivityBetRankWillRankSkipTodayRsp{}
	case 3:
		window.Skip3 = true
		rsp = &pb.ActivityBetRankLoseRankSkipTodayRsp{}
	}

	windowBytes, err := sonic.Marshal(window)
	if err == nil {
		err = rdb.HSet(ctx, betRankWindowsKey, req.Userid, string(windowBytes)).Err()
		if err != nil {
			return
		}
	}
	return
}

// requestActivityBetRankReword 排行榜奖励领取
func requestActivityBetRankReword(req *mq.RequestActivityBetRankRewordArgs) (rsp *pb.ActivityBetRankRewordRsp, err error) {
	prize := &data.ActivityBetRankPrize{Id: req.Id}
	prize.GetById()
	if prize.Userid == "" || prize.Userid != req.Userid {
		err = fmt.Errorf("betrank prize not found: %s", req.Id)
		return
	}
	if prize.Received != 0 {
		err = fmt.Errorf("betrank prize already received: %s", req.Id)
		return
	}
	prize.Received = 1
	prize.ReceiveTime = time.Now().Unix()
	prize.Save()

	rsp = &pb.ActivityBetRankRewordRsp{
		Prize:     prize.Prize,
		PrizeType: prize.PrizeType,
	}
	return
}

// requestActivityBetRankMyRecord 获取个人中奖记录
func requestActivityBetRankMyRecord(req *mq.RequestUserArgs) (rsp *pb.ActivityBetRankMyRecordRsp, err error) {
	rsp = &pb.ActivityBetRankMyRecordRsp{}
	m := bson.M{"userid": req.Userid}
	prizes := data.ListActivityBetRankPrize(m)
	sort.Slice(prizes, func(i, j int) bool { return prizes[i].Etime > prizes[j].Etime })

	for _, prize := range prizes {
		rsp.Records = append(rsp.Records, &pb.ActivityBetRankMyRecord{
			RankType:  prize.RankType,
			Rank:      prize.Rank + 1,
			Bets:      prize.Bets,
			Prize:     prize.Prize,
			PrizeRate: prize.PrizeRate,
			PrizeTime: prize.Ctime.In(location).Format(utils.FORMAT2),
		})
	}
	return
}

// requestActivityBetRankMyRecord 排行榜历史请求
func requestActivityBetRankHistory(req *mq.RequestUserArgs) (rsp *pb.ActivityBetRankHistoryRsp, err error) {
	rsp = &pb.ActivityBetRankHistoryRsp{
		Daily:   &pb.ActivityBetRankList{User: &pb.ActivityBetRankSelf{}},
		Weekly:  &pb.ActivityBetRankList{User: &pb.ActivityBetRankSelf{}},
		Monthly: &pb.ActivityBetRankList{User: &pb.ActivityBetRankSelf{}},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var prevEtime1, prevEtime2, prevEtime3 int64
	prevEtime1 = betRankDailyEtime - int64((time.Hour * 24 * 1).Seconds())
	prevEtime2 = betRankWeeklyEtime - int64((time.Hour * 24 * 7).Seconds())
	// prevEtime3 = time.Unix(betRankMonthlyEtime, 0).In(location).AddDate(0, -1, 0).Unix()
	monthEtime := time.Unix(betRankMonthlyEtime, 0).In(location)
	prevEtime3 = time.Date(monthEtime.Year(), monthEtime.Month(), 1, 0, 0, 0, 0, location).Add(-time.Second).Unix()

	if err = betRankLastHistory(ctx, 1, prevEtime1, rsp.Daily); err != nil {
		return
	}
	if err = betRankLastHistory(ctx, 2, prevEtime2, rsp.Weekly); err != nil {
		return
	}
	if err = betRankLastHistory(ctx, 3, prevEtime3, rsp.Monthly); err != nil {
		return
	}
	return
}

// betRankLastHistory 排行榜上一期历史查询
func betRankLastHistory(ctx context.Context, rankType, prevEtime int64, ranks *pb.ActivityBetRankList) (err error) {
	historyKey, _ := utils.CaseWhen3(rankType, 1, historyDailyKey, 2, historyWeeklyKey, 3, historyMonthlyKey)
	cacheKey := fmt.Sprintf("%s:%d", historyKey, prevEtime)

	rankCache, err := rdb.Get(ctx, cacheKey).Result()
	if err != nil && err != redis.Nil {
		glog.Error(err)
		return
	}
	if err == nil {
		// 使用缓存
		var cacheBytes []byte
		if cacheBytes, err = base64.StdEncoding.DecodeString(rankCache); err != nil {
			return
		}
		if err = ranks.Unmarshal(cacheBytes); err != nil {
			return
		}
		return
	}

	// 查询历史并缓存
	prizes := data.ListActivityBetRankPrize(bson.M{"rank_type": rankType, "etime": prevEtime})
	for i, prize := range prizes {
		user := &pb.ActivityBetRankUser{
			Userid:    prize.Userid,
			Username:  prize.Username,
			VipLv:     prize.VipLv,
			Photo:     prize.Photo,
			Rank:      prize.Rank + 1,
			Bets:      prize.Bets,
			Prize:     prize.Prize,
			PrizeRate: prize.PrizeRate,
		}
		ranks.Ranks = append(ranks.Ranks, user)
		if i == 0 {
			ranks.Jackpot = prize.Jackpot
			ranks.PrizeType = prize.PrizeType
		}
	}
	sort.Slice(ranks.Ranks, func(i, j int) bool { return ranks.Ranks[i].Rank < ranks.Ranks[j].Rank })

	var bytes []byte
	if bytes, err = ranks.Marshal(); err != nil {
		return
	}
	cache := base64.StdEncoding.EncodeToString(bytes)
	cacheEx, _ := utils.CaseWhen3(rankType, 1, time.Hour*24, 2, time.Hour*24*7, 3, time.Hour*24*31)
	if err = rdb.SetEx(ctx, cacheKey, cache, cacheEx).Err(); err != nil {
		return
	}
	return
}
