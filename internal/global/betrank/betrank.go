package betrank

import (
	"context"
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
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
	betRankEtimesKey    = "activity:betrank:etimes"          // 结束时间存储
	betRankUsersKey     = "activity:betrank:users"           // 打码榜用户信息
	betRankWindowsKey   = "activity:betrank:windows"         // 用户弹窗信息
	betRankDailyKey     = "activity:betrank:rank:daily"      // 打码量日榜
	betRankWeeklyKey    = "activity:betrank:rank:weekly"     // 打码量周榜
	betRankMonthlyKey   = "activity:betrank:rank:monthly"    // 打码量月榜
	jackpotDailyKey     = "activity:betrank:jackpot:daily"   // 日奖池
	jackpotWeeklyKey    = "activity:betrank:jackpot:weekly"  // 周奖池
	jackpotMonthlyKey   = "activity:betrank:jackpot:monthly" // 月奖池
	betRankHasPrizesKey = "activity:betrank:has_prizes"      // 有未领取奖励用户列表

	prevJackpotDailyKey       = "activity:betrank:prev_jackpot:daily"        // 上一周期日奖池
	prevJackpotWeeklyKey      = "activity:betrank:prev_jackpot:weekly"       // 上一周期周奖池
	prevJackpotMonthlyKey     = "activity:betrank:prev_jackpot:monthly"      // 上一周期月奖池
	preJackpotDailyKey        = "activity:betrank:pre_jackpot:daily"         // 日奖池已加到奖池前置展示金额
	preJackpotWeeklyKey       = "activity:betrank:pre_jackpot:weekly"        // 周奖池已加到奖池前置展示金额
	preJackpotMonthlyKey      = "activity:betrank:pre_jackpot:monthly"       // 月奖池已加到奖池前置展示金额
	repayPreJackpotDailyKey   = "activity:betrank:repay_pre_jackpot:daily"   // 日奖池前置已偿还
	repayPreJackpotWeeklyKey  = "activity:betrank:repay_pre_jackpot:weekly"  // 周奖池前置已偿还
	repayPreJackpotMonthlyKey = "activity:betrank:repay_pre_jackpot:monthly" // 月奖池前置已偿还
	repayPreJackpotSubsidyKey = "activity:betrank:repay_pre_jackpot:subsidy" // 补贴金额前置已偿还
	subsidyJackpotDailyKey    = "activity:betrank:subsidy_jackpot:daily"     // 日奖池已补贴金额
	subsidyJackpotWeeklyKey   = "activity:betrank:subsidy_jackpot:weekly"    // 周奖池已补贴金额
	subsidyJackpotMonthlyKey  = "activity:betrank:subsidy_jackpot:monthly"   // 月奖池已补贴金额

	historyDailyKey   = "activity:betrank:history:daily"   // 日榜历史查询缓存
	historyWeeklyKey  = "activity:betrank:history:weekly"  // 周榜历史查询缓存
	historyMonthlyKey = "activity:betrank:history:monthly" // 月榜历史查询缓存
)

var (
	betRankDailyEtime   int64 = 0 // 日榜结束时间
	betRankWeeklyEtime  int64 = 0 // 周榜结束时间
	betRankMonthlyEtime int64 = 0 // 月榜结束时间

	betRankDailyUpdateNtfTime   int64 = 0 // 日榜更新通知时间
	betRankWeeklyUpdateNtfTime  int64 = 0 // 周榜更新通知时间
	betRankMonthlyUpdateNtfTime int64 = 0 // 月榜更新通知时间
)

type BetRankUser struct {
	Userid                  string `json:"userid"` // 用户id
	Username                string `json:"username"`
	Photo                   string `json:"photo"`
	VipLv                   int32  `json:"vipLv"`
	RegistArea              int    `json:"registArea"` // ab测试 0:A 1:B 2:C
	Robot                   bool   `json:"robot"`
	BetsDaily               int64  `json:"betsDaily"`          // 日榜打码量
	BetsWeekly              int64  `json:"betsWeekly"`         // 周榜打码量
	BetsMonthly             int64  `json:"betsMonthly"`        // 月榜打码量
	PeriodMaxRankDaily      int64  `json:"periodMaxRankDaily"` // 本次评奖周期内最高达到排名
	PeriodMaxRankWeekly     int64  `json:"periodMaxRankWeekly"`
	PeriodMaxRankMonthly    int64  `json:"periodMaxRankMonthly"`
	PeriodResetEtimeDaily   int64  `json:"periodResetEtimeDaily"`   // 日榜周期重置时间
	PeriodResetEtimeWeekly  int64  `json:"periodResetEtimeWeekly"`  // 周榜周期重置时间
	PeriodResetEtimeMonthly int64  `json:"periodResetEtimeMonthly"` // 月榜周期重置时间

	// 同比环比记录
	PrevRankDaily       int64 `json:"prevRankDaily"` // 上次排行榜日榜位置,方便计算排名上升还是下降
	PrevRankWeekly      int64 `json:"prevRankWeekly"`
	PrevRankMonthly     int64 `json:"prevRankMonthly"`
	PrevRankDailyTime   int64 `json:"prevRankDailyTime"` // 上次排行榜日榜位置时间
	PrevRankWeeklyTime  int64 `json:"prevRankWeeklyTime"`
	PrevRankMonthlyTime int64 `json:"prevRankMonthlyTime"`
}

// GetPeriodBets 获取当前周期内打码量
func (u *BetRankUser) GetPeriodBets(rankType int32) int64 {
	periodEtime, _ := utils.CaseWhen3(rankType, 1, u.PeriodResetEtimeDaily, 2, u.PeriodResetEtimeWeekly, 3, u.PeriodResetEtimeMonthly)
	etime, _ := utils.CaseWhen3(rankType, 1, betRankDailyEtime, 2, betRankWeeklyEtime, 3, betRankMonthlyEtime)
	if periodEtime != etime {
		return 0
	}
	bets, _ := utils.CaseWhen3(rankType, 1, u.BetsDaily, 2, u.BetsWeekly, 3, u.BetsMonthly)
	return bets
}

// GetPeriodMaxRank 获取当前周期最大排名
func (u *BetRankUser) GetPeriodMaxRank(rankType int32) int64 {
	periodEtime, _ := utils.CaseWhen3(rankType, 1, u.PeriodResetEtimeDaily, 2, u.PeriodResetEtimeWeekly, 3, u.PeriodResetEtimeMonthly)
	etime, _ := utils.CaseWhen3(rankType, 1, betRankDailyEtime, 2, betRankWeeklyEtime, 3, betRankMonthlyEtime)
	if periodEtime != etime {
		return -1
	}
	maxRank, _ := utils.CaseWhen3(rankType, 1, u.PeriodMaxRankDaily, 2, u.PeriodMaxRankWeekly, 3, u.PeriodMaxRankMonthly)
	return maxRank
}

// GePrevRank 获取上一周期排名
func (u *BetRankUser) GetPrevRank(rankType int32) int64 {
	prevRankTime, _ := utils.CaseWhen3(rankType, 1, u.PrevRankDailyTime, 2, u.PrevRankWeeklyTime, 3, u.PrevRankMonthlyTime)
	if prevRankTime <= 0 {
		return 0
	}

	switch rankType {
	case 1:
		prevRankTime += int64((time.Hour * 24).Seconds())
	case 2:
		prevRankTime += int64((time.Hour * 24 * 7).Seconds())
	case 3:
		cur := time.Unix(prevRankTime, 0).In(location)
		nextMonth := time.Date(cur.Year(), cur.Month(), 1, 0, 0, 0, 0, location).AddDate(0, 2, 0).Add(-time.Second)
		prevRankTime = nextMonth.Unix()
	}

	etime, _ := utils.CaseWhen3(rankType, 1, betRankDailyEtime, 2, betRankWeeklyEtime, 3, betRankMonthlyEtime)
	if prevRankTime != etime {
		return 0
	}

	prevRank, _ := utils.CaseWhen3(rankType, 1, u.PrevRankDaily, 2, u.PrevRankWeekly, 3, u.PrevRankMonthly)
	return prevRank
}

// BetRankWindow 排行榜弹窗
type BetRankWindow struct {
	Userid                  string `json:"userid"`                  // 用户id
	PeriodMaxRankDaily      int64  `json:"periodMaxRankDaily"`      // 上一次弹窗本次评奖周期内最高达到排名
	PeriodMaxRankWeekly     int64  `json:"periodMaxRankWeekly"`     //
	PeriodMaxRankMonthly    int64  `json:"periodMaxRankMonthly"`    //
	PeriodResetEtimeDaily   int64  `json:"periodResetEtimeDaily"`   // 日榜周期重置时间
	PeriodResetEtimeWeekly  int64  `json:"periodResetEtimeWeekly"`  // 周榜周期重置时间
	PeriodResetEtimeMonthly int64  `json:"periodResetEtimeMonthly"` // 月榜周期重置时间
	RankDaily               int64  `json:"rankDaily"`               // 上次回到大厅时日榜排行
	RankWeekly              int64  `json:"rankWeekly"`              //
	RankMonthly             int64  `json:"rankMonthly"`             //

	Skip1 bool `json:"skip1"` // skip 排行榜上榜报喜弹窗
	Skip2 bool `json:"skip2"` // skip 快上榜提醒弹窗
	Skip3 bool `json:"skip3"` // skip 掉榜弹窗
}

func InitBetRank() {
	now := time.Now().In(location)
	updateBetRankTimes(now)
}

// updateBetRankTimes 更新排行榜开始时间
func updateBetRankTimes(now time.Time) {
	defer func() {
		if err := rdb.HSet(context.Background(), betRankEtimesKey, "daily", fmt.Sprint(betRankDailyEtime)).Err(); err != nil {
			glog.Error(err)
		}
		if err := rdb.HSet(context.Background(), betRankEtimesKey, "weekly", fmt.Sprint(betRankWeeklyEtime)).Err(); err != nil {
			glog.Error(err)
		}
		if err := rdb.HSet(context.Background(), betRankEtimesKey, "monthly", fmt.Sprint(betRankMonthlyEtime)).Err(); err != nil {
			glog.Error(err)
		}
	}()

	nowSec := now.Unix()
	if nowSec > betRankDailyEtime {
		if betRankDailyEtime > 0 {
			betRankDailyEtime += int64((time.Hour * 24).Seconds())
		} else {
			etime, ok := getPrevBetRankTime("daily")
			if ok {
				betRankDailyEtime = etime
			} else {
				todayStr := now.Format("2006-01-02")
				etime := utils.Str2Time(fmt.Sprintf("%s 23:59:59", todayStr), location)
				// 当天结束时间
				betRankDailyEtime = etime.Unix()
			}
		}
	}

	// 一周结束时间,周天
	if nowSec > betRankWeeklyEtime {
		if betRankWeeklyEtime > 0 {
			betRankWeeklyEtime += int64((time.Hour * 24 * 7).Seconds())
		} else {
			etime, ok := getPrevBetRankTime("weekly")
			if ok {
				betRankWeeklyEtime = etime
			} else {
				today := now.Weekday()
				offset := 7 - today
				if today == time.Sunday {
					offset = 0
				}
				sunday := now.AddDate(0, 0, int(offset))
				weekStr := sunday.Format("2006-01-02")
				etime := utils.Str2Time(fmt.Sprintf("%s 23:59:59", weekStr), location)
				betRankWeeklyEtime = etime.Unix()

			}
		}
	}

	// 本月结束时间
	if nowSec > betRankMonthlyEtime {
		if betRankMonthlyEtime > 0 {
			cur := time.Unix(betRankMonthlyEtime, 0).In(location)
			nextMonth := time.Date(cur.Year(), cur.Month(), 1, 0, 0, 0, 0, location).AddDate(0, 2, 0).Add(-time.Second)
			// nextMonth := time.Unix(betRankMonthlyEtime, 0).In(location).AddDate(0, 1, 0)
			betRankMonthlyEtime = nextMonth.Unix()
		} else {
			etime, ok := getPrevBetRankTime("monthly")
			if ok {
				betRankMonthlyEtime = etime
			} else {
				now := now.In(location)
				monthStime := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, location)
				monthEtime := monthStime.AddDate(0, 1, 0).Add(-time.Second)
				betRankMonthlyEtime = monthEtime.Unix()
			}
		}
	}
}

// getPrevBetRankTime 获取上次停机时排行榜时间
func getPrevBetRankTime(fKey string) (rankEtime int64, ok bool) {
	etimeS, err := rdb.HGet(context.Background(), betRankEtimesKey, fKey).Result()
	if err != nil {
		if err != redis.Nil {
			glog.Error(err)
		}
		return
	}
	etime, err := strconv.Atoi(etimeS)
	if err != nil {
		glog.Error(err)
		return
	}
	ok = true
	rankEtime = int64(etime)
	return
}

func getJackpot(ctx context.Context, rankType int32) (ret int64, err error) {
	// 奖池数据
	jackpotKey, _ := utils.CaseWhen3(rankType, 1, jackpotDailyKey, 2, jackpotWeeklyKey, 3, jackpotMonthlyKey)
	jackpot, e := rdb.Get(ctx, jackpotKey).Result()
	if e != nil && e != redis.Nil {
		err = e
		return
	}
	j, _ := strconv.ParseFloat(jackpot, 64)
	ret = int64(j)
	return
}

func getJackpotFloat64(ctx context.Context, key string) (ret float64, err error) {
	jackpot, e := rdb.Get(ctx, key).Result()
	if e != nil {
		if e != redis.Nil {
			err = e
		}
		return
	}
	ret, err = strconv.ParseFloat(jackpot, 64)
	return
}

func hGetJackpotFloat64(ctx context.Context, key string, field string) (ret float64, err error) {
	jackpot, e := rdb.HGet(ctx, key, field).Result()
	if e != nil {
		if e != redis.Nil {
			err = e
		}
		return
	}
	ret, err = strconv.ParseFloat(jackpot, 64)
	return
}

func getPrize(rankType int32, rank int32, jackpot int64) (prize int64, prizeRate string) {
	prizeTable := table.GetTables().BetRankPrizeRateTable
	p := prizeTable.Get(rank)
	if p != nil {
		rate, _ := utils.CaseWhen3(rankType, 1, p.Daily, 2, p.Weekly, 3, p.Monthly)
		prize = miShare(rate, jackpot)
		prizeRate = mi2pct(rate)
	}
	return
}

func getPrizeType(rankType, registArea int32) (prizeType int32) {
	betRankTable := table.GetTables().BetRankBaseTable.Get()
	prizeTypes := betRankTable.PrizeTypes[registArea]
	prizeType, _ = utils.CaseWhen3(rankType, 1, prizeTypes.Value[0], 2, prizeTypes.Value[1], 3, prizeTypes.Value[2])
	return
}

// getRank 获取实时排名 daily, weekly, monthly, err := getRank(ctx, userid)
func getRank(ctx context.Context, userid string) (daily int64, weekly int64, monthly int64, err error) {
	defer func() {
		if err == redis.Nil {
			err = nil
		}
	}()

	if daily, err = rdb.ZRevRank(ctx, betRankDailyKey, userid).Result(); err != nil {
		daily = -1
	}
	if weekly, err = rdb.ZRevRank(ctx, betRankWeeklyKey, userid).Result(); err != nil {
		weekly = -1
	}
	if monthly, err = rdb.ZRevRank(ctx, betRankMonthlyKey, userid).Result(); err != nil {
		monthly = -1
	}
	return
}

// handlePublishBets 打码上报处理
func handlePublishBets(bet *pb.PublishGameBets) (err error) {
	// fmt.Printf("handle publish bets: %#v\n", bet)
	if bet.Bets <= 0 || bet.Userid == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	betRankTable := table.GetTables().BetRankBaseTable.Get()
	if !bet.Robot && !utils.SliceIn(bet.Gtype, betRankTable.BetsExcludeGtypes...) {
		// 打码量增加总奖池
		addDailyJackpot := float64(bet.Bets) * (float64(betRankTable.ActualBetsRate[0]) / 10000)
		addWeeklyJackpot := float64(bet.Bets) * (float64(betRankTable.ActualBetsRate[1]) / 10000)
		addMonthlyJackpot := float64(bet.Bets) * (float64(betRankTable.ActualBetsRate[2]) / 10000)

		// glog.Infof("user bets: %s, %d, %d, %d, %d", bet.Userid, bet.Bets, addDailyJackpot, addWeeklyJackpot, addMonthlyJackpot)
		// 偿还奖池前置
		var addDaily, addWeekly, addMonthly float64
		if addDaily, err = calcBetPreJackpotRepay(1, addDailyJackpot); err != nil {
			return
		}
		if addWeekly, err = calcBetPreJackpotRepay(2, addWeeklyJackpot); err != nil {
			return
		}
		if addMonthly, err = calcBetPreJackpotRepay(3, addMonthlyJackpot); err != nil {
			return
		}
		// glog.Infof("user bets add jackpot: %s, %d, %d, %d, %d", bet.Userid, bet.Bets, addDaily, addWeekly, addMonthly)

		// 增加到奖池
		if err = rdb.IncrByFloat(ctx, jackpotDailyKey, addDaily).Err(); err != nil {
			return
		}
		if err = rdb.IncrByFloat(ctx, jackpotWeeklyKey, addWeekly).Err(); err != nil {
			return
		}
		if err = rdb.IncrByFloat(ctx, jackpotMonthlyKey, addMonthly).Err(); err != nil {
			return
		}
	}

	// 打码量排行处理
	if !utils.SliceIn(bet.Gtype, betRankTable.RanksExcludeGtypes...) {
		err = handlePublishBetsRanks(ctx, bet)
		if err != nil {
			return
		}
	}
	return
}

// calcBetPreJackpotRepay 奖池前置展示打码偿还
func calcBetPreJackpotRepay(rankType int32, addJackpot float64) (repayLeft float64, err error) {
	repayLeft = addJackpot
	ctx := context.Background()
	// 已经增加的前置展示金额
	var alreadyAddPreJackpot float64
	preJackpotKey, _ := utils.CaseWhen3(rankType, 1, preJackpotDailyKey, 2, preJackpotWeeklyKey, 3, preJackpotMonthlyKey)
	preJackpotS, e := rdb.Get(ctx, preJackpotKey).Result()
	if e != nil {
		if e != redis.Nil {
			err = e
			glog.Errorf("get prev jackpot error: %s, %v", preJackpotKey, err)
			return
		}
	} else {
		alreadyAddPreJackpot, err = strconv.ParseFloat(preJackpotS, 64)
		if err != nil {
			glog.Errorf("parse pre jackpot error: %s, %v", preJackpotS, err)
			return
		}
	}

	// 已经偿还的前置金额
	var repayPreJackpot float64
	repayPreJackpotKey, _ := utils.CaseWhen3(rankType, 1, repayPreJackpotDailyKey, 2, repayPreJackpotWeeklyKey, 3, repayPreJackpotMonthlyKey)
	repayPreJackpotS, e := rdb.Get(ctx, repayPreJackpotKey).Result()
	if e != nil {
		if e != redis.Nil {
			err = e
			glog.Errorf("get repay prev jackpot error: %s, %v", repayPreJackpotKey, err)
			return
		}
	} else {
		repayPreJackpot, err = strconv.ParseFloat(repayPreJackpotS, 64)
		if err != nil {
			glog.Errorf("parse repay pre jackpot error: %s, %v", repayPreJackpotS, err)
			return
		}
	}

	// 已偿还完
	if repayPreJackpot >= alreadyAddPreJackpot {
		return
	}
	// 计算偿还比例
	previewRate := table.GetTables().BetRankBaseTable.Get().PreviewRates[rankType-1]
	// 评奖周期过了前置展示比例，按升档偿还比例进行偿还
	stime := getBetRankStime(rankType)
	etime, _ := utils.CaseWhen3(rankType, 1, betRankDailyEtime, 2, betRankWeeklyEtime, 3, betRankMonthlyEtime)
	nowSec := time.Now().Unix()
	if float64(nowSec-stime)/float64(etime-stime)*10000 > float64(previewRate) {
		previewRate = table.GetTables().BetRankBaseTable.Get().UpRepayRate[rankType-1]
	}

	// 计算偿还金额
	repayNum := addJackpot * (float64(previewRate) / 10000)
	if _, err = rdb.IncrByFloat(ctx, repayPreJackpotKey, repayNum).Result(); err != nil {
		return
	}
	repayLeft = addJackpot - repayNum
	return
}

// handlePublishBetsRanks 打码量排行处理
func handlePublishBetsRanks(ctx context.Context, bet *pb.PublishGameBets) (err error) {
	var user = &BetRankUser{
		Userid:               bet.Userid,
		PeriodMaxRankDaily:   -1,
		PeriodMaxRankWeekly:  -1,
		PeriodMaxRankMonthly: -1,
	}
	r, e := rdb.HGet(ctx, betRankUsersKey, bet.Userid).Result()
	if e != nil {
		if e != redis.Nil {
			err = e
			return
		}
	} else {
		if err = sonic.Unmarshal([]byte(r), user); err != nil {
			return
		}
	}
	user.Username = bet.Username
	user.VipLv = bet.VipLv
	user.Photo = bet.Photo
	user.RegistArea = int(bet.RegistArea)
	user.Robot = bet.Robot

	// 当前分数
	var dailyScore, weeklyScore, monthlyScore float64
	if dailyScore, err = rdb.ZIncrBy(ctx, betRankDailyKey, float64(bet.Bets), bet.Userid).Result(); err != nil {
		return
	}
	if weeklyScore, err = rdb.ZIncrBy(ctx, betRankWeeklyKey, float64(bet.Bets), bet.Userid).Result(); err != nil {
		return
	}
	if monthlyScore, err = rdb.ZIncrBy(ctx, betRankMonthlyKey, float64(bet.Bets), bet.Userid).Result(); err != nil {
		return
	}
	user.BetsDaily, user.BetsWeekly, user.BetsMonthly = int64(dailyScore), int64(weeklyScore), int64(monthlyScore)

	// 当前排名
	rankDaily, rankWeekly, rankMonthly, err := getRank(ctx, bet.Userid)
	if err != nil {
		return
	}

	// 重置周期内最高记录
	if user.PeriodResetEtimeDaily != betRankDailyEtime {
		user.PeriodMaxRankDaily = -1
		user.PeriodResetEtimeDaily = betRankDailyEtime
	}
	if user.PeriodResetEtimeWeekly != betRankWeeklyEtime {
		user.PeriodMaxRankWeekly = -1
		user.PeriodResetEtimeWeekly = betRankWeeklyEtime
	}
	if user.PeriodResetEtimeMonthly != betRankMonthlyEtime {
		user.PeriodMaxRankMonthly = -1
		user.PeriodResetEtimeMonthly = betRankMonthlyEtime
	}

	// 历史新排名
	if rankDaily != -1 && (rankDaily < user.PeriodMaxRankDaily || user.PeriodMaxRankDaily == -1) {
		user.PeriodMaxRankDaily = rankDaily
	}
	if rankWeekly != -1 && (rankWeekly < user.PeriodMaxRankWeekly || user.PeriodMaxRankWeekly == -1) {
		user.PeriodMaxRankWeekly = rankWeekly
	}
	if rankMonthly != -1 && (rankMonthly < user.PeriodMaxRankMonthly || user.PeriodMaxRankMonthly == -1) {
		user.PeriodMaxRankMonthly = rankMonthly
	}

	userBytes, err := sonic.Marshal(user)
	if err == nil {
		err = rdb.HSet(ctx, betRankUsersKey, bet.Userid, string(userBytes)).Err()
		if err != nil {
			return
		}
	}
	return
}

// handleUserWindows 用户排行榜弹窗判断
func handleUserWindows(userid string) (ntfs *pb.ActivityBetRankWindowNtfs, ntf bool, err error) {
	ntfs = &pb.ActivityBetRankWindowNtfs{Userid: userid}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user := &BetRankUser{}
	r, e := rdb.HGet(ctx, betRankUsersKey, userid).Result()
	if e != nil {
		if e != redis.Nil {
			err = e
		}
		return
	}
	if err = sonic.Unmarshal([]byte(r), user); err != nil {
		return
	}

	// 获取实时排名
	rankDaily, rankWeekly, rankMonthly, err := getRank(ctx, userid)
	if err != nil {
		return
	}

	window := &BetRankWindow{}
	r, e = rdb.HGet(ctx, betRankWindowsKey, userid).Result()
	if e != nil {
		if e != redis.Nil {
			err = e
			return
		} else {
			window = &BetRankWindow{Userid: userid}
		}
	} else {
		if err = sonic.Unmarshal([]byte(r), window); err != nil {
			return
		}
	}

	// 中奖弹窗
	hasPrizes, err := rdb.HExists(ctx, betRankHasPrizesKey, userid).Result()
	if err != nil {
		glog.Error(err)
		return
	}
	// glog.Infof("betrank hasprize: %s, %d", userid, hasPrizes)
	// var ntf0 []*pb.ActivityBetRankRewordNtf
	if hasPrizes {
		// 查询中奖
		prizes := data.ListUserBetRankPrizeUnreceived(userid)
		// glog.Infof("betrank prizes: %s %#v", userid, prizes)
		for _, prize := range prizes {
			timeRange := fmt.Sprintf("%s - %s",
				time.Unix(prize.Stime, 0).In(location).Format(utils.FORMAT),
				time.Unix(prize.Etime, 0).In(location).Format(utils.FORMAT))
			ntf := &pb.ActivityBetRankRewordNtf{
				Id:        prize.Id,
				RankType:  prize.RankType,
				Jackpot:   prize.Jackpot,
				Rank:      prize.Rank + 1,
				Prize:     prize.Prize,
				PrizeRate: prize.PrizeRate,
				TimeRange: timeRange,
				PrizeType: prize.PrizeType,
			}
			ntfs.Ntf0 = append(ntfs.Ntf0, ntf)
		}
		if len(prizes) == 0 {
			// 没有奖励了
			if err := rdb.HDel(ctx, betRankHasPrizesKey, userid).Err(); err != nil {
				glog.Error(err)
			}
		}
	}

	// 重置周期内最高记录
	if window.PeriodResetEtimeDaily != betRankDailyEtime {
		window.PeriodMaxRankDaily = -1
		window.PeriodResetEtimeDaily = betRankDailyEtime
		window.RankDaily = -1
		window.Skip1 = false
		window.Skip2 = false
		window.Skip3 = false
	}
	if window.PeriodResetEtimeWeekly != betRankWeeklyEtime {
		window.PeriodMaxRankWeekly = -1
		window.PeriodResetEtimeWeekly = betRankWeeklyEtime
		window.RankWeekly = -1
	}
	if window.PeriodResetEtimeMonthly != betRankMonthlyEtime {
		window.PeriodMaxRankMonthly = -1
		window.PeriodResetEtimeMonthly = betRankMonthlyEtime
		window.RankMonthly = -1
	}

	// 弹窗与用户数据时间周期对比
	inPeriodDaily, inPeriodWeekly, inPeriodMonthly :=
		window.PeriodResetEtimeDaily == user.PeriodResetEtimeDaily,
		window.PeriodResetEtimeWeekly == user.PeriodResetEtimeWeekly,
		window.PeriodResetEtimeMonthly == user.PeriodResetEtimeMonthly

	betRankTable := table.GetTables().BetRankBaseTable.Get()
	dailyLimit, weeklyLimit, monthlyLimit :=
		int64(betRankTable.RankLimits[0]),
		int64(betRankTable.RankLimits[1]),
		int64(betRankTable.RankLimits[2])

	// jackpots init
	var jackpots = []int64{0, 0, 0}
	for i := range jackpots {
		if jackpot, e := getJackpot(ctx, int32(i+1)); e == nil {
			jackpots[i] = jackpot
		}
	}

	// 排行榜上榜报喜弹窗
	// var ntfs1 []*pb.ActivityBetRankNewRankNtf
	if !window.Skip1 {
		// 日
		if inPeriodDaily && rankDaily != -1 && rankDaily < dailyLimit &&
			(rankDaily < window.PeriodMaxRankDaily || window.PeriodMaxRankDaily == -1) {
			window.PeriodMaxRankDaily = rankDaily
			ntf := &pb.ActivityBetRankNewRankNtf{
				RankType: 1,
				Rank:     int32(rankDaily) + 1,
				Bets:     user.GetPeriodBets(1),
			}
			ntf.Jackpot = jackpots[ntf.RankType-1]
			ntf.Prize, ntf.PrizeRate = getPrize(ntf.RankType, ntf.Rank, ntf.Jackpot)
			ntf.PrizeType = getPrizeType(ntf.RankType, int32(user.RegistArea))
			ntfs.Ntf1 = append(ntfs.Ntf1, ntf)
		}
		// 周
		if inPeriodWeekly && rankWeekly != -1 && rankWeekly < weeklyLimit &&
			(rankWeekly < window.PeriodMaxRankWeekly || window.PeriodMaxRankWeekly == -1) {
			window.PeriodMaxRankWeekly = rankWeekly
			ntf := &pb.ActivityBetRankNewRankNtf{
				RankType: 2,
				Rank:     int32(rankWeekly) + 1,
				Bets:     user.GetPeriodBets(2),
			}
			ntf.Jackpot = jackpots[ntf.RankType-1]
			ntf.Prize, ntf.PrizeRate = getPrize(ntf.RankType, ntf.Rank, ntf.Jackpot)
			ntf.PrizeType = getPrizeType(ntf.RankType, int32(user.RegistArea))
			ntfs.Ntf1 = append(ntfs.Ntf1, ntf)
		}
		// 月
		if inPeriodMonthly && rankMonthly != -1 && rankMonthly < monthlyLimit &&
			(rankMonthly < window.PeriodMaxRankMonthly || window.PeriodMaxRankMonthly == -1) {
			window.PeriodMaxRankMonthly = rankMonthly
			ntf := &pb.ActivityBetRankNewRankNtf{
				RankType: 3,
				Rank:     int32(rankMonthly) + 1,
				Bets:     user.GetPeriodBets(3),
			}
			ntf.Jackpot = jackpots[ntf.RankType-1]
			ntf.Prize, ntf.PrizeRate = getPrize(ntf.RankType, ntf.Rank, ntf.Jackpot)
			ntf.PrizeType = getPrizeType(ntf.RankType, int32(user.RegistArea))
			ntfs.Ntf1 = append(ntfs.Ntf1, ntf)
		}
	}

	// 掉榜弹窗
	// var ntfs3 []*pb.ActivityBetRankLoseRankNtf
	var ntfs3Map = make(map[int32]bool, 3)
	if !window.Skip3 {
		// 日
		if inPeriodDaily && rankDaily >= dailyLimit && (window.RankDaily < dailyLimit && window.RankDaily != -1) {
			// 获取最后一名分数
			var betsGap int64 = 1
			lastone, e := rdb.ZRevRangeWithScores(ctx, betRankDailyKey, dailyLimit-1, dailyLimit-1).Result()
			if e == nil && len(lastone) > 0 {
				betsGap = int64(lastone[0].Score) - user.GetPeriodBets(1)
			}
			ntf := &pb.ActivityBetRankLoseRankNtf{
				RankType: 1,
				Rank:     int32(rankDaily) + 1,
				Bets:     user.GetPeriodBets(1),
				BetsGap:  betsGap,
			}
			ntf.Jackpot = jackpots[ntf.RankType-1]
			ntf.Prize, ntf.PrizeRate = getPrize(ntf.RankType, 1, ntf.Jackpot)
			ntf.PrizeType = getPrizeType(ntf.RankType, int32(user.RegistArea))
			ntfs.Ntf3 = append(ntfs.Ntf3, ntf)
			ntfs3Map[ntf.RankType] = true
		}
		// 周
		if inPeriodWeekly && rankWeekly >= weeklyLimit && (window.RankWeekly < weeklyLimit && window.RankWeekly != -1) {
			// 获取最后一名分数
			var betsGap int64 = 1
			lastone, e := rdb.ZRevRangeWithScores(ctx, betRankWeeklyKey, weeklyLimit-1, weeklyLimit-1).Result()
			if e == nil && len(lastone) > 0 {
				betsGap = int64(lastone[0].Score) - user.GetPeriodBets(2)
			}
			ntf := &pb.ActivityBetRankLoseRankNtf{
				RankType: 2,
				Rank:     int32(rankWeekly) + 1,
				Bets:     user.GetPeriodBets(2),
				BetsGap:  betsGap,
			}
			ntf.Jackpot = jackpots[ntf.RankType-1]
			ntf.Prize, ntf.PrizeRate = getPrize(ntf.RankType, 1, ntf.Jackpot)
			ntf.PrizeType = getPrizeType(ntf.RankType, int32(user.RegistArea))
			ntfs.Ntf3 = append(ntfs.Ntf3, ntf)
			ntfs3Map[ntf.RankType] = true
		}
		// 月
		if inPeriodMonthly && rankMonthly >= monthlyLimit && (window.RankMonthly < monthlyLimit && window.RankMonthly != -1) {
			// 获取最后一名分数
			var betsGap int64 = 1
			lastone, e := rdb.ZRevRangeWithScores(ctx, betRankMonthlyKey, monthlyLimit-1, monthlyLimit-1).Result()
			if e == nil && len(lastone) > 0 {
				betsGap = int64(lastone[0].Score) - user.GetPeriodBets(3)
			}
			ntf := &pb.ActivityBetRankLoseRankNtf{
				RankType: 3,
				Rank:     int32(rankMonthly) + 1,
				Bets:     user.GetPeriodBets(3),
				BetsGap:  betsGap,
			}
			ntf.Jackpot = jackpots[ntf.RankType-1]
			ntf.Prize, ntf.PrizeRate = getPrize(ntf.RankType, 1, ntf.Jackpot)
			ntf.PrizeType = getPrizeType(ntf.RankType, int32(user.RegistArea))
			ntfs.Ntf3 = append(ntfs.Ntf3, ntf)
			ntfs3Map[ntf.RankType] = true
		}
	}

	upTip := betRankTable.UpTips[user.RegistArea]
	// 快上榜提醒弹窗(掉榜弹窗和快上榜提醒弹窗同时弹时，只弹掉榜弹窗，快上榜弹提醒弹窗此次不能弹)
	// var ntfs2 []*pb.ActivityBetRankWillRankNtf
	if !window.Skip2 {
		// 日
		if inPeriodDaily && !ntfs3Map[1] && rankDaily >= dailyLimit && rankDaily != -1 {
			// 获取最后一名分数
			var betsGap int64 = 1
			lastone, e := rdb.ZRevRangeWithScores(ctx, betRankDailyKey, dailyLimit-1, dailyLimit-1).Result()
			if e == nil && len(lastone) > 0 {
				betsGap = int64(lastone[0].Score) - user.GetPeriodBets(1)
			}
			if betsGap <= int64(upTip.Value[0]) {
				ntf := &pb.ActivityBetRankWillRankNtf{
					RankType: 1,
					Rank:     int32(rankDaily) + 1,
					Bets:     user.GetPeriodBets(1),
					BetsGap:  betsGap,
				}
				ntf.Jackpot = jackpots[ntf.RankType-1]
				ntf.Prize, ntf.PrizeRate = getPrize(ntf.RankType, 1, ntf.Jackpot)
				ntf.PrizeType = getPrizeType(ntf.RankType, int32(user.RegistArea))
				ntf.RankLimit = int32(dailyLimit)
				ntfs.Ntf2 = append(ntfs.Ntf2, ntf)
			}
		}
		// 周
		if inPeriodWeekly && !ntfs3Map[2] && rankWeekly >= weeklyLimit && rankWeekly != -1 {
			// 获取最后一名分数
			var betsGap int64 = 1
			lastone, e := rdb.ZRevRangeWithScores(ctx, betRankWeeklyKey, weeklyLimit-1, weeklyLimit-1).Result()
			if e == nil && len(lastone) > 0 {
				betsGap = int64(lastone[0].Score) - user.GetPeriodBets(2)
			}
			if betsGap <= int64(upTip.Value[1]) {
				ntf := &pb.ActivityBetRankWillRankNtf{
					RankType: 2,
					Rank:     int32(rankWeekly) + 1,
					Bets:     user.GetPeriodBets(2),
					BetsGap:  betsGap,
				}
				ntf.Jackpot = jackpots[ntf.RankType-1]
				ntf.Prize, ntf.PrizeRate = getPrize(ntf.RankType, 1, ntf.Jackpot)
				ntf.PrizeType = getPrizeType(ntf.RankType, int32(user.RegistArea))
				ntf.RankLimit = int32(weeklyLimit)
				ntfs.Ntf2 = append(ntfs.Ntf2, ntf)
			}
		}
		// 月
		if inPeriodMonthly && !ntfs3Map[3] && rankMonthly >= monthlyLimit && rankMonthly != -1 {
			// 获取最后一名分数
			var betsGap int64 = 1
			lastone, e := rdb.ZRevRangeWithScores(ctx, betRankMonthlyKey, monthlyLimit-1, monthlyLimit-1).Result()
			if e == nil && len(lastone) > 0 {
				betsGap = int64(lastone[0].Score) - user.GetPeriodBets(3)
			}
			if betsGap <= int64(upTip.Value[2]) {
				ntf := &pb.ActivityBetRankWillRankNtf{
					RankType: 3,
					Rank:     int32(rankMonthly) + 1,
					Bets:     user.GetPeriodBets(3),
					BetsGap:  betsGap,
				}
				ntf.Jackpot = jackpots[ntf.RankType-1]
				ntf.Prize, ntf.PrizeRate = getPrize(ntf.RankType, 1, ntf.Jackpot)
				ntf.PrizeType = getPrizeType(ntf.RankType, int32(user.RegistArea))
				ntf.RankLimit = int32(monthlyLimit)
				ntfs.Ntf2 = append(ntfs.Ntf2, ntf)
			}
		}
	}

	// 更新回到大厅排名
	if inPeriodDaily {
		window.RankDaily = rankDaily
	}
	if inPeriodWeekly {
		window.RankWeekly = rankWeekly
	}
	if inPeriodMonthly {
		window.RankMonthly = rankMonthly
	}
	windowBytes, err := sonic.Marshal(window)
	if err == nil {
		err = rdb.HSet(ctx, betRankWindowsKey, userid, string(windowBytes)).Err()
		if err != nil {
			return
		}
	}

	ntf = len(ntfs.Ntf0) > 0 ||
		len(ntfs.Ntf1) > 0 ||
		len(ntfs.Ntf2) > 0 ||
		len(ntfs.Ntf3) > 0
	return
}

// handleBetRankReword 检查排行榜时间派奖和重置时间
func handleBetRankRewordTick(now time.Time) {
	var expired bool
	defer func() {
		if expired {
			updateBetRankTimes(now)
		}
	}()

	nowSec := now.Unix()
	if nowSec > betRankDailyEtime {
		expired = true
		handleBetRankReword(1, betRankDailyEtime)
	}
	if nowSec > betRankWeeklyEtime {
		expired = true
		handleBetRankReword(2, betRankWeeklyEtime)
	}
	if nowSec > betRankMonthlyEtime {
		expired = true
		handleBetRankReword(3, betRankMonthlyEtime)
	}
}

// getBetRankStime 计算排行榜开始时间
func getBetRankStime(rankType int32) (stime int64) {
	switch rankType {
	case 1:
		stime = betRankDailyEtime - int64((time.Hour * 24 * 1).Seconds()) + 1
	case 2:
		stime = betRankWeeklyEtime - int64((time.Hour * 24 * 7).Seconds()) + 1
	case 3:
		monthEtime := time.Unix(betRankMonthlyEtime, 0).In(location)
		stime = time.Date(monthEtime.Year(), monthEtime.Month(), 1, 0, 0, 0, 0, location).Unix()
		// stime = time.Unix(betRankMonthlyEtime, 0).In(location).AddDate(0, -1, 0).Unix() + 1
	}
	return
}

// handleBetRankReword 排行榜派奖、重置奖池、重置排行榜、保存排行榜记录
func handleBetRankReword(rankType int32, etime int64) {
	glog.Infof("betrank send reword: %d, %d", rankType, etime)

	betRankTable := table.GetTables().BetRankBaseTable.Get()
	dailyLimit, weeklyLimit, monthlyLimit :=
		int64(betRankTable.RankLimits[0]),
		int64(betRankTable.RankLimits[1]),
		int64(betRankTable.RankLimits[2])

	ctx := context.Background()
	rankLimit, _ := utils.CaseWhen3(rankType, 1, dailyLimit, 2, weeklyLimit, 3, monthlyLimit)
	betRankKey, _ := utils.CaseWhen3(rankType, 1, betRankDailyKey, 2, betRankWeeklyKey, 3, betRankMonthlyKey)
	rankUsers, err := rdb.ZRevRangeWithScores(ctx, betRankKey, 0, rankLimit-1).Result()
	if err != nil {
		glog.Error("betrank reword get ranklist error: ", err)
		return
	}
	prizeTable := table.GetTables().BetRankPrizeRateTable

	jackpot, err := getJackpot(ctx, rankType)
	if err != nil {
		glog.Error("betrank reword get jackpot error: ", err)
		return
	}
	jackpotPrize := float64(jackpot)

	var jackpotPre, jackpotRepayPre, jackpotSubsidy, jackpotSubsidyRepayPre int64
	// 查询前置金额
	preJackpotKey, _ := utils.CaseWhen3(rankType, 1, preJackpotDailyKey, 2, preJackpotWeeklyKey, 3, preJackpotMonthlyKey)
	pre, err := getJackpotFloat64(ctx, preJackpotKey)
	if err != nil {
		glog.Errorf("get pre jackpot error: %s, %v", preJackpotKey, err)
	} else {
		jackpotPre = int64(pre)
	}
	// 查询已偿还前置金额
	repayPreJackpotKey, _ := utils.CaseWhen3(rankType, 1, repayPreJackpotDailyKey, 2, repayPreJackpotWeeklyKey, 3, repayPreJackpotMonthlyKey)
	repayPre, err := getJackpotFloat64(ctx, repayPreJackpotKey)
	if err != nil {
		glog.Errorf("get repay pre jackpot error: %s, %v", repayPreJackpotKey, err)
	} else {
		jackpotRepayPre = int64(repayPre)
	}
	// 查询补贴金额
	subsidyJackpotKey, _ := utils.CaseWhen3(rankType, 1, subsidyJackpotDailyKey, 2, subsidyJackpotWeeklyKey, 3, subsidyJackpotMonthlyKey)
	subsidy, err := getJackpotFloat64(ctx, subsidyJackpotKey)
	if err != nil {
		glog.Errorf("get subsidy jackpot error: %s, %v", subsidyJackpotKey, err)
	} else {
		jackpotSubsidy = int64(subsidy)
	}
	subsidyRepayPre, err := hGetJackpotFloat64(ctx, repayPreJackpotSubsidyKey, fmt.Sprint(rankType))
	if err != nil {
		glog.Errorf("get subsidy repay pre error: %s, %d, %v", repayPreJackpotSubsidyKey, rankType, err)
	} else {
		jackpotSubsidyRepayPre = int64(subsidyRepayPre)
	}

	var useridMap = make(map[string]*BetRankUser)
	if len(rankUsers) > 0 {
		var userids []string
		for _, ru := range rankUsers {
			userids = append(userids, ru.Member.(string))
		}
		users, err := rdb.HMGet(ctx, betRankUsersKey, userids...).Result()
		if err != nil {
			glog.Error("betrank reword get users error: ", err)
			return
		}
		for _, u := range users {
			user := &BetRankUser{}
			if err := sonic.Unmarshal([]byte(fmt.Sprint(u)), user); err != nil {
				continue
			}
			useridMap[user.Userid] = user
		}
	}

	// 排行榜开始时间
	stime := getBetRankStime(rankType)
	now := time.Now().In(location)
	var prizes []*data.ActivityBetRankPrize

	var updateUserPrevRanks []any
	for rank, ru := range rankUsers {
		userid := ru.Member.(string)
		bets := int64(ru.Score)
		p := prizeTable.Get(int32(rank) + 1)
		if p == nil {
			continue
		}
		rate, _ := utils.CaseWhen3(rankType, 1, p.Daily, 2, p.Weekly, 3, p.Monthly)
		prizeRatePct := mi2pct(rate)
		prizeNum := miShare(rate, jackpot)
		prizeType := getPrizeType(rankType, 0)
		if prizeNum <= 0 {
			continue
		}

		var username, photo string
		var robot bool
		var vipLv int32
		// 更新上次日榜排名位置
		if user, ok := useridMap[userid]; ok {
			username = user.Username
			photo = user.Photo
			robot = user.Robot
			vipLv = user.VipLv
			switch rankType {
			case 1:
				user.PrevRankDaily = int64(rank) + 1
				user.PrevRankDailyTime = etime
			case 2:
				user.PrevRankWeekly = int64(rank) + 1
				user.PrevRankWeeklyTime = etime
			case 3:
				user.PrevRankMonthly = int64(rank) + 1
				user.PrevRankMonthlyTime = etime
			}
			ubytes, err := sonic.Marshal(user)
			if err != nil {
				glog.Error("betrank reword marshal user error: %#v", user)
			} else {
				updateUserPrevRanks = append(updateUserPrevRanks, []any{userid, string(ubytes)}...)
			}
		}

		jackpotPrize -= float64(prizeNum)
		if jackpotPrize < 0 { // 奖池不够了
			continue
		}

		prize := &data.ActivityBetRankPrize{
			Id:                     fmt.Sprintf("%d-%d-%s", etime, rankType, userid),
			Userid:                 userid,
			Username:               username,
			VipLv:                  vipLv,
			Photo:                  photo,
			Robot:                  robot,
			RankType:               rankType,
			Rank:                   int32(rank),
			Bets:                   bets,
			Prize:                  int64(prizeNum),
			PrizeRate:              prizeRatePct,
			Jackpot:                jackpot,
			JackpotPre:             jackpotPre,
			JackpotRepayPre:        jackpotRepayPre,
			JackpotSubsidyRepayPre: jackpotSubsidyRepayPre,
			JackpotSubsidy:         jackpotSubsidy,
			PrizeType:              prizeType,
			Stime:                  stime,
			Etime:                  etime,
			Ctime:                  now,
			Received:               0,
		}
		prizes = append(prizes, prize)
	}

	// 更新上次日榜排名位置
	if len(updateUserPrevRanks) > 0 {
		if err := rdb.HMSet(ctx, betRankUsersKey, updateUserPrevRanks).Err(); err != nil {
			glog.Errorf("save betrank updateUserPrevRanks error: %#v", updateUserPrevRanks)
		}
	}

	// 保存redis hmap记录有奖励, 领取时删除
	var prizeUsers []any
	for _, prize := range prizes {
		if !prize.Save() {
			glog.Errorf("save betrank prize error: %s,  %v", prize.Userid, prize)
			continue
		}
		if !prize.Robot {
			prizeUsers = append(prizeUsers, prize.Userid, "")
		}
	}
	if len(prizeUsers) > 0 {
		if err := rdb.HMSet(ctx, betRankHasPrizesKey, prizeUsers...).Err(); err != nil {
			glog.Error("save betrank prize keys error: ", err)
		}
	}

	// 记录上一期奖池
	prevJackpotKey, _ := utils.CaseWhen3(rankType, 1, prevJackpotDailyKey, 2, prevJackpotWeeklyKey, 3, prevJackpotMonthlyKey)
	if err := rdb.Set(ctx, prevJackpotKey, fmt.Sprint(jackpot), 0).Err(); err != nil {
		glog.Errorf("betrank prev jackpot set error: %d, %s, %v", rankType, prevJackpotKey, err)
	}

	// 清空奖池、排行榜、前置奖池、周期已补贴金额
	jackpotKey, _ := utils.CaseWhen3(rankType, 1, jackpotDailyKey, 2, jackpotWeeklyKey, 3, jackpotMonthlyKey)
	if err := rdb.Del(ctx, betRankKey, jackpotKey, preJackpotKey, repayPreJackpotKey, subsidyJackpotKey).Err(); err != nil {
		glog.Errorf("clean betrank prize keys error: %d, %v", rankType, err)
	}

	// 清空补贴偿前置金额
	if err := rdb.HDel(ctx, repayPreJackpotSubsidyKey, fmt.Sprint(rankType)).Err(); err != nil {
		glog.Errorf("clean repay subsidy keys error: %s, %d, %v", repayPreJackpotSubsidyKey, rankType, err)
	}

	if rankType == 1 {
		// 重置当日人机打码数据
		if err := rdb.Del(ctx, betRankHourRobotsKey).Err(); err != nil {
			glog.Errorf("clean  betrank robot hours keys error: %d, %v", rankType, err)
		}
		for i := range len(enterHourRobots) {
			enterHourRobots[i] = nil
		}
	}
}

// handleUpdateBetRankNtfsTick 排行榜更新通知
func handleUpdateBetRankNtfsTick(nowSec int64) {
	handleUpdateBetRankNtf(nowSec, 1)
	handleUpdateBetRankNtf(nowSec, 2)
	handleUpdateBetRankNtf(nowSec, 3)
}

func handleUpdateBetRankNtf(nowSec int64, rankType int32) {
	sendHz := table.GetTables().BetRankBaseTable.Get().SendHzs[rankType-1]
	sendHzSec := int64(sendHz) * 60

	prevNtfTime, _ := utils.CaseWhen3(rankType, 1, &betRankDailyUpdateNtfTime, 2, &betRankWeeklyUpdateNtfTime, 3, &betRankMonthlyUpdateNtfTime)
	// 发送时间未到
	if nowSec-*prevNtfTime < sendHzSec {
		return
	}
	*prevNtfTime = nowSec

	// 排行榜开始时间
	stime := getBetRankStime(rankType)
	if nowSec-stime > sendHzSec {
		// 排行榜奖池前置
		betRankPreviewJackpot(rankType)
		// 排行榜补贴
		betRankSubsidy(rankType)
	}

	// 奖池更新通知, gate 再批量请求排行榜更新进行推送
	if err := mq.NatsPublish(mq.TopicActivityBetRankUpdate, &pb.PublishActivityBetRankUpdate{RankType: rankType}); err != nil {
		glog.Error("publish betrank update error: ", err)
	}
}

// getPreJackpots 获取奖池前置展示量
// prevJackpot 上期奖池
// preJackpot 奖池前置展示值
// prePartNum 发送频率每次奖池前置增加值
// alreadyAddPreJackpot 已经增加的奖池前置值
func getPreJackpots(ctx context.Context, rankType int32) (prevJackpot, preJackpot, prePartNum, alreadyAddPreJackpot float64, err error) {
	previewRate := table.GetTables().BetRankBaseTable.Get().PreviewRates[rankType-1]
	previewTime := table.GetTables().BetRankBaseTable.Get().PreviewTimes[rankType-1]
	sendHz := table.GetTables().BetRankBaseTable.Get().SendHzs[rankType-1]

	prevJackpotKey, _ := utils.CaseWhen3(rankType, 1, prevJackpotDailyKey, 2, prevJackpotWeeklyKey, 3, prevJackpotMonthlyKey)
	prevJackpotS, e := rdb.Get(ctx, prevJackpotKey).Result()
	if e != nil {
		if e != redis.Nil {
			err = e
			glog.Errorf("get prev jackpot error: %s, %v", prevJackpotKey, err)
			return
		}
		return
	} else {
		prevJackpot, err = strconv.ParseFloat(prevJackpotS, 64)
		if err != nil {
			glog.Errorf("parse prev jackpot error: %s, %v", prevJackpotS, err)
			return
		}
	}

	// 奖池展示前置比例
	preJackpot = float64(prevJackpot) * (float64(previewRate) / 10000)
	parts := math.Ceil(float64(previewTime) / float64(sendHz))
	prePartNum = float64(preJackpot) / parts

	// 已经增加的前置展示比例
	preJackpotKey, _ := utils.CaseWhen3(rankType, 1, preJackpotDailyKey, 2, preJackpotWeeklyKey, 3, preJackpotMonthlyKey)
	preJackpotS, e := rdb.Get(ctx, preJackpotKey).Result()
	if e != nil {
		if e != redis.Nil {
			err = e
			glog.Errorf("get prev jackpot error: %s, %v", preJackpotKey, err)
			return
		}
	} else {
		alreadyAddPreJackpot, err = strconv.ParseFloat(preJackpotS, 64)
		if err != nil {
			glog.Errorf("parse pre jackpot error: %s, %v", preJackpotS, err)
			return
		}
	}
	return
}

// betRankPreviewJackpot 奖池展示前置 奖池更新发送前调用
func betRankPreviewJackpot(rankType int32) {
	// glog.Info("betRankPreviewJackpot: ", rankType)

	// 上周期奖池基数
	ctx := context.Background()
	prevJackpot, preJackpot, prePartNum, alreadyAddPreJackpot, err := getPreJackpots(ctx, rankType)
	if err != nil {
		glog.Errorf("get pre jackpots error: %d, %v", rankType, err)
		return
	}
	// 上期奖池为0
	if prevJackpot == 0 {
		return
	}

	// 前置奖池已全部加到奖池
	if (alreadyAddPreJackpot + prePartNum) > preJackpot {
		return
	}

	// 将前置展示比例加到奖池
	preJackpotKey, _ := utils.CaseWhen3(rankType, 1, preJackpotDailyKey, 2, preJackpotWeeklyKey, 3, preJackpotMonthlyKey)
	preJ, err := rdb.IncrByFloat(ctx, preJackpotKey, prePartNum).Result()
	if err != nil {
		glog.Errorf("pre jackpot incr error: %s, %f, %v", preJackpotKey, prePartNum, err)
		return
	}

	jackpotKey, _ := utils.CaseWhen3(rankType, 1, jackpotDailyKey, 2, jackpotWeeklyKey, 3, jackpotMonthlyKey)
	jackpot, err := rdb.IncrByFloat(ctx, jackpotKey, prePartNum).Result()
	if err != nil {
		glog.Errorf("jackpot pre incr error: %s, %f, %v", jackpotKey, prePartNum, err)
		return
	}

	glog.Infof("pre jackpot incr %s, incr=%f, preJackpot=%f, jackpot=%f, prevJackpot=%f", preJackpotKey, prePartNum, preJ, jackpot, prevJackpot)
}

// betRankSubsidy 补贴金额 奖池更新发送前调用
// 当日活少，打码少的时候，会补贴一部分钱放到奖池中
// 具体金额按照配置表中“补贴金额”执行。具体放法就是将配置的补贴金额按照“发送频率”来等分成若干份发给前端，均匀地在一天中摊如奖池中。
// 如果榜单周期内发送频率和补贴金额的配置改变了，要扣掉已经加进奖池的，然后再算余下还有多少，再按照新的发送频率和剩下的评奖周期时间来均分。
// 每日重置, 评奖周期重置？
func betRankSubsidy(rankType int32) {
	// glog.Info("betRankSubsidy: ", rankType)
	ctx := context.Background()
	// 补贴金额
	subsidyAmount := table.GetTables().BetRankBaseTable.Get().SubsidyAmounts[rankType-1]
	// 已经补贴金额
	var alreadySubsidyAmount float64
	subsidyJackpotKey, _ := utils.CaseWhen3(rankType, 1, subsidyJackpotDailyKey, 2, subsidyJackpotWeeklyKey, 3, subsidyJackpotMonthlyKey)
	alreadySubsidyAmountS, err := rdb.Get(ctx, subsidyJackpotKey).Result()
	if err != nil {
		if err != redis.Nil {
			glog.Errorf("get subsidy jackpot error: %s, %v", subsidyJackpotKey, err)
			return
		}
	} else {
		alreadySubsidyAmount, err = strconv.ParseFloat(alreadySubsidyAmountS, 64)
		if err != nil {
			glog.Errorf("parse subsidy jackpot error: %s, %v", alreadySubsidyAmountS, err)
			return
		}
	}

	// 每次补贴金额
	now := time.Now().In(location)
	etime, _ := utils.CaseWhen3(rankType, 1, betRankDailyEtime, 2, betRankWeeklyEtime, 3, betRankMonthlyEtime) // 结束时间
	nowSec := now.Unix()

	sendHz := table.GetTables().BetRankBaseTable.Get().SendHzs[rankType-1]
	parts := float64(etime-nowSec) / float64(sendHz*60)
	if parts == 0 {
		return
	}

	subsidyJackpot := (float64(subsidyAmount) - alreadySubsidyAmount) / parts

	// 超过补贴金额
	if subsidyJackpot <= 0 || alreadySubsidyAmount+subsidyJackpot > float64(subsidyAmount) {
		return
	}

	// 补贴金额偿还奖池前置
	leftSubsidyJackpot, err := calcBetPreJackpotRepay(rankType, subsidyJackpot)
	if err != nil {
		glog.Error(err)
		return
	}

	// 补贴金额偿还
	subsidyRepay := subsidyJackpot - leftSubsidyJackpot
	// 记录补贴金额偿还
	alreadySubsidyRepay, err := rdb.HIncrByFloat(ctx, repayPreJackpotSubsidyKey, fmt.Sprint(rankType), subsidyRepay).Result()
	if err != nil {
		glog.Errorf("jackpot subsidy repay incr error: %s, %d, %f, %v", repayPreJackpotSubsidyKey, rankType, alreadySubsidyRepay, err)
	}

	// 记录已补贴金额
	curSubsidyAmount, err := rdb.IncrByFloat(ctx, subsidyJackpotKey, subsidyJackpot).Result()
	if err != nil {
		glog.Errorf("subsidy jackpot incr error: %s, %f, %v", subsidyJackpotKey, subsidyJackpot, err)
		return
	}

	// 将偿还剩余的补贴金额加到奖池
	jackpotKey, _ := utils.CaseWhen3(rankType, 1, jackpotDailyKey, 2, jackpotWeeklyKey, 3, jackpotMonthlyKey)
	jackpot, err := rdb.IncrByFloat(ctx, jackpotKey, leftSubsidyJackpot).Result()
	if err != nil {
		glog.Errorf("jackpot pre incr error: %s, %f, %v", jackpotKey, leftSubsidyJackpot, err)
		return
	}

	glog.Infof("subsidy jackpot incr %s, incr=%.2f, repay=%.2f, subsidy=%.2f, jackpot=%.2f", subsidyJackpotKey, leftSubsidyJackpot, subsidyRepay, curSubsidyAmount, jackpot)
}
