package service

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/internal/web/admin/app/entity"
	"goserver/pkg/data/ck"
	"goserver/pkg/utils"
	"math"
	"runtime/debug"
	"sort"
	"strconv"
	"time"

	"github.com/astaxie/beego"
	"github.com/globalsign/mgo/bson"
)

type statistics2Service struct {
}

// 分页查询Ltv统计信息
func (s *statistics2Service) GetLtvStat2List(page, pageSize int, m bson.M, start, end time.Time) (stats []bson.M, err error) {
	for i := 0; ; i++ {
		stime := start.AddDate(0, 0, i)
		if stime.After(end) {
			break
		}
		stat := s.ltvStat2ByDay(m, stime)
		stats = append(stats, stat)
	}
	return
}

func (s *statistics2Service) ltvStat2ByDay(m bson.M, today time.Time) (stat bson.M) {
	todayStr := today.Format("2006-01-02")
	stime := utils.Str2Time(fmt.Sprintf("%s 00:00:00", todayStr), Location())
	etime := utils.Str2Time(fmt.Sprintf("%s 23:59:59", todayStr), Location())
	todayStart := utils.TimestampTodayTimeZone(locationName) // 今天开始时间
	todayEnd := todayStart.AddDate(0, 0, 1)                  // 今天结束时间

	stat = bson.M{
		"date":     todayStr, // 日期
		"count":    0,        // 新注册
		"payCount": 0,        // 付费用户累积数
	}
	stat["utilToday"] = int(todayStart.Sub(stime).Hours() / 24)

	m["robot"] = false
	m["simulation_robot"] = false
	m["ctime"] = bson.M{"$gte": stime, "$lt": etime}

	// 注册用户
	list, _ := PlayerService.GetByUserList(m)
	count := len(list)
	userids := make([]string, 0, count)
	for _, user := range list {
		userids = append(userids, user.Userid)
	}

	// 新注册数
	if count > 0 {
		stat["count"] = count
	}

	// 活跃免费活跃用户
	var activeFree, activePay int
	m["login_time"] = bson.M{"$gte": todayStart}
	mActive := []bson.M{
		{"$match": m},
		{"$project": bson.M{"recharge": bson.M{"$gt": []interface{}{"$money", 0}}}},
		{"$group": bson.M{"_id": "$recharge", "count": bson.M{"$sum": 1}}},
	}
	var rActive []bson.M
	err := PlayerUsers.Pipe(mActive).All(&rActive)
	delete(m, "login_time")
	if err != nil {
		beego.Error("活跃用户统计失败", err)
	} else if len(rActive) > 0 {
		for _, ac := range rActive {
			if ac["_id"].(bool) {
				activePay = ac["count"].(int)
			} else {
				activeFree = ac["count"].(int)
			}
		}
	}
	stat["activeFree"] = activeFree
	stat["activePay"] = activePay

	// 新用户充值人数
	var payCount int
	mPay := bson.M{
		"order_status": 4,
		"userid":       bson.M{"$in": userids},
	}
	query3 := []bson.M{
		{"$match": mPay},
		{"$group": bson.M{"_id": "$userid", "num": bson.M{"$sum": 1}}},
		{"$group": bson.M{"_id": nil, "num": bson.M{"$sum": 1}}},
	}
	pipe := Pays.Pipe(query3)
	result := []bson.M{}
	if err := pipe.All(&result); err != nil {
		beego.Error(fmt.Errorf("新用户充值人数 error %v", err))
	} else if len(result) > 0 {
		payCount = result[0]["num"].(int)
	}
	stat["payCount"] = payCount
	// 付费人数占比
	var payRate = "0"
	if count > 0 {
		payRate = fmt.Sprintf("%.2f", float64(payCount)/float64(count)*100)
	}
	stat["payRate"] = payRate

	// 截止昨天昨日付费用户
	var payCountYesterday int
	mPay["ctime"] = bson.M{"$lt": todayStart}
	result2 := []bson.M{}
	if err := Pays.Pipe(query3).All(&result2); err != nil {
		beego.Error(fmt.Errorf("截止昨天昨日付费用户 error %v", err))
	} else if len(result2) > 0 {
		payCountYesterday = result2[0]["num"].(int)
	}
	stat["payCountYesterday"] = payCountYesterday

	// ltv d1-d7
	// now := time.Now()
	// statEndTime := utils.TimestampTodayTime().AddDate(0, 0, 1) // 统计结束时间

	// statEtimeStr := utils.TimestampTodayTime().Format("2006-01-02")
	// statEndTime := utils.Str2Time(fmt.Sprintf("%s 00:00:00", statEtimeStr))
	// statEndTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)

	ltvs := []int{0, 1, 3, 7, 15, 30, 60}

	for _, ltv := range ltvs {
		key := fmt.Sprintf("ltv%d", ltv)
		stat[key] = map[string]string{"ltvV1": "0", "ltvV2": "0", "ltvPayV1": "0", "ltvPayV2": "0"}

		// begin := todayStart.AddDate(0, 0, -ltv)
		// end := todayStart
		// if ltv == 0 { // 今天数据
		// 	end = todayStart.AddDate(0, 0, 1)
		// }

		begin := stime
		// end := todayEnd.AddDate(0, 0, -ltv)
		end := stime.AddDate(0, 0, ltv+1)
		if ltv == 0 { // 今天数据：截止到当前数据
			end = todayEnd
		}
		if !end.After(begin) {
			stat[key] = map[string]string{"ltvV1": "/", "ltvV2": "/", "ltvPayV1": "/", "ltvPayV2": "/"}
			continue
		}

		if payCount == 0 {
			continue
		}
		mPay["order_status"] = 4 // pay success status
		mPay["ctime"] = bson.M{"$gte": begin, "$lt": end}
		// ltv时间范围该批充值用户数
		mCharges := []bson.M{
			{"$match": mPay},
			{"$group": bson.M{"_id": "$userid", "num": bson.M{"$sum": 1}}},
			{"$group": bson.M{"_id": nil, "num": bson.M{"$sum": 1}}},
		}
		pipe := Pays.Pipe(mCharges)
		result := []bson.M{}
		err := pipe.All(&result)
		var payCount int64 = 0
		if err == nil && len(result) > 0 {
			if s, ok := result[0]["num"].(int); ok {
				payCount = int64(s)
			}
		}

		// ltv时间范围内该批充值总额
		payM := []bson.M{
			{"$match": mPay},
			{"$group": bson.M{"_id": nil, "sum": bson.M{"$sum": "$amount"}}},
		}
		pipe = Pays.Pipe(payM)
		result = []bson.M{}
		err = pipe.All(&result)
		var paySum int64 = 0
		if err == nil && len(result) > 0 {
			if s, ok := result[0]["sum"].(int); ok {
				paySum = int64(s)
			}
		}

		// ltv时间范围该批提现用户数
		// mPay["order_status"] = 2 // withdraw success status
		// mWithdraws := []bson.M{
		// 	{"$match": mPay},
		// 	{"$group": bson.M{"_id": "$userid", "num": bson.M{"$sum": 1}}},
		// 	{"$group": bson.M{"_id": nil, "num": bson.M{"$sum": 1}}},
		// }

		// pipe = Withdraws.Pipe(mWithdraws)
		// result = []bson.M{}
		// err = pipe.All(&result)
		// var ltvWithdrawCount int64 = 0
		// if err == nil && len(result) > 0 {
		// 	if s, ok := result[0]["num"].(int); ok {
		// 		ltvWithdrawCount = int64(s)
		// 	}
		// }

		// ltv时间范围内该批提现总额
		mPay["order_status"] = 2 // withdraw success status
		withdrawM := []bson.M{
			{"$match": mPay},
			{"$group": bson.M{"_id": nil, "sum": bson.M{"$sum": "$amount"}}},
		}
		pipe = Withdraws.Pipe(withdrawM)
		result = []bson.M{}
		err = pipe.All(&result)
		var withdrawSum int64 = 0
		if err == nil && len(result) > 0 {
			if s, ok := result[0]["sum"].(int); ok {
				withdrawSum = int64(s)
			}
		}
		//
		// LTV 总冲金额除当日注册人数
		// LTV2 充值减提现 除当日注册人数
		// 付费用户ltv1 总冲金额除 当日注册且付费人数
		// 付费用户ltv1 充值减提现除 当日注册且付费人数
		var ltvV1, ltvV2, ltvPayV1, ltvPayV2 float64
		ltvV1 = float64(paySum) / float64(count)
		ltvV2 = float64(paySum-withdrawSum) / float64(count)
		if payCount != 0 {
			ltvPayV1 = float64(paySum) / float64(payCount)
			ltvPayV2 = float64(paySum-withdrawSum) / float64(payCount)
		}
		stat[key] = map[string]string{
			"ltvV1":    fmt.Sprintf("%.2f", ltvV1/100.0),
			"ltvV2":    fmt.Sprintf("%.2f", ltvV2/100.0),
			"ltvPayV1": fmt.Sprintf("%.2f", ltvPayV1/100.0),
			"ltvPayV2": fmt.Sprintf("%.2f", ltvPayV2/100.0),
		}
	}

	return stat
}

// BehaviorAnalysis 玩家行为分析
func (s *statistics2Service) BehaviorAnalysis(m bson.M, startTime time.Time, charges1, range1, limit1, charges2, range2, limit2, charges3 int) (
	analysis1 map[int]map[string]int,
	analysis1Users map[int]map[string][]string,
	analysis2 map[int]map[string]int,
	analysis2Users map[int]map[string][]string,
	analysis3 map[int]map[int][]int, // n充,游戏类型,[betAmount, rounds]
	err error,
) {
	// 游戏局数 统计局数，统计区间
	m1 := []bson.M{
		{"$match": m},
		{"$project": bson.M{"_id": "$_id"}},
	}
	var r1 []bson.M
	err = PlayerUsers.Pipe(m1).All(&r1)
	if err != nil {
		beego.Error("error1: ", err)
		return
	}
	var userids []string
	for _, r := range r1 {
		userid := r["_id"].(string)
		userids = append(userids, userid)
	}
	if len(userids) == 0 {
		return
	}

	// 查询注册时间范围内玩家充值次数, userids -> 对局记录(打码量)
	payTimes := make(map[string]int, len(userids))
	m2 := []bson.M{
		{"$match": bson.M{"userid": bson.M{"$in": userids}, "order_status": 4, "ctime": bson.M{"$gt": startTime}}},
		{"$group": bson.M{"_id": "$userid", "times": bson.M{"$sum": 1}}},
	}
	var r2 []bson.M
	err = Pays.Pipe(m2).All(&r2)
	if err != nil {
		return
	}
	for _, r := range r2 {
		userid := r["_id"].(string)
		times := r["times"].(int)
		payTimes[userid] = times
	}

	// 充值次数对应玩家列表
	payTimesUser := make(map[int][]string, 16)
	for _, userid := range userids {
		times := payTimes[userid]
		// if times > pChargeTimes {
		// 	payTimesUser[pChargeTimes+1] = append(payTimesUser[pChargeTimes+1], userid)
		// 	continue
		// }
		payTimesUser[times] = append(payTimesUser[times], userid)
	}

	if charges1 > 0 && range1 > 0 && limit1 > 0 {
		analysis1, analysis1Users, err = behanviorAnalysis1Round(startTime, charges1, range1, limit1, userids, payTimesUser, payTimes)
		if err != nil {
			return
		}
	}

	if charges2 > 0 && range2 > 0 && limit2 > 0 {
		analysis2, analysis2Users, analysis3, err = behanviorAnalysis2BetAmount(startTime, charges2, range2, limit2, charges3, userids, payTimesUser, payTimes)
		if err != nil {
			return
		}
	}
	return
}

// behanviorAnalysis1Round 表一 局数分析
// pChargeTimes 充值次数参数
func behanviorAnalysis1Round(startTime time.Time, pChargeTimes, range1, limit1 int, userids []string, payTimesUser map[int][]string, payTimes map[string]int) (
	analysis1 map[int]map[string]int,
	analysis1Users map[int]map[string][]string,
	err error) {
	// 查询0充玩家游戏局数
	playerRounds := make(map[string]int, len(payTimesUser[0]))
	playerRoundsTime := make(map[string][]int) // 付费玩家第1,2,3...局时间秒

	m3 := []bson.M{
		{"$match": bson.M{"begin_time": bson.M{"$gt": startTime.Unix()}, "players": bson.M{"$ne": ""}}},
		{"$project": bson.M{"begin_time": "$begin_time", "userids": bson.M{"$split": []string{"$players", ","}}}},
		{"$match": bson.M{"userids": bson.M{"$in": userids}}},
		{"$sort": bson.M{"begin_time": 1}},
	}
	var r3 []bson.M
	err = Details.Pipe(m3).All(&r3)
	if err != nil {
		return
	}
	for _, r := range r3 {
		beginTime := r["begin_time"].(int64) // 对局开始时间
		for _, uid := range r["userids"].([]interface{}) {
			userid := uid.(string)
			if len(userid) >= 16 { // 人机id18位长度+
				continue
			}
			playerRounds[userid]++
			if payTimes[userid] > 0 {
				playerRoundsTime[userid] = append(playerRoundsTime[userid], int(beginTime))
			}
		}
	}
	// 外接记录局数
	// m4 := []bson.M{
	// 	{
	// 		"$match": bson.M{"user_id": bson.M{"$in": payTimesUser[0]}, "ctime": bson.M{"$gt": startTime.Unix()}},
	// 	},
	// 	{
	// 		"$group": bson.M{
	// 			"_id":  bson.M{"round_id": "$round_id", "user_id": "$user_id"},
	// 			"bets": bson.M{"$sum": 1},
	// 		},
	// 	},
	// 	{
	// 		"$group": bson.M{
	// 			"_id":    "$_id.user_id",
	// 			"rounds": bson.M{"$sum": 1},
	// 		},
	// 	},
	// }
	// // 外接游戏局数时间
	// r4 := []bson.M{}
	// err = NsqLogExternalBets.Pipe(m4).All(&r4)
	// if err != nil {
	// 	beego.Error("error2: ", err)
	// 	return
	// }
	// for _, ur := range r4 {
	// 	userid := ur["_id"].(string)
	// 	round := ur["rounds"].(int)
	// 	playerRounds[userid] += round
	// }

	m9 := []bson.M{
		{
			"$match": bson.M{"user_id": bson.M{"$in": userids}, "ctime": bson.M{"$gt": startTime.Unix()}, "amount": bson.M{"$ne": 0}},
		},
		{"$group": bson.M{"_id": bson.M{"round_id": "$round_id", "user_id": "$user_id"}, "ctime": bson.M{"$min": "$ctime"}}},
		{"$group": bson.M{"_id": bson.M{"user_id": "$_id.user_id", "ctime": "$ctime"}}},
	}
	// 外接游戏局数时间
	var r9 []bson.M
	err = NsqLogExternalBets.Pipe(m9).All(&r9)
	if err != nil {
		beego.Error("error2: ", err)
		return
	}
	willSortRtimeUsers := make(map[string]bool)
	for _, ur := range r9 {
		userid := ur["_id"].(bson.M)["user_id"].(string)
		ctime := ur["_id"].(bson.M)["ctime"].(int64) // 对局开始时间
		playerRounds[userid]++
		if payTimes[userid] > 0 {
			playerRoundsTime[userid] = append(playerRoundsTime[userid], int(ctime))
			willSortRtimeUsers[userid] = true
		}
	}
	// 与外接游戏时间一起排序
	for userid := range willSortRtimeUsers {
		sort.Slice(playerRoundsTime[userid], func(i, j int) bool {
			return playerRoundsTime[userid][i] < playerRoundsTime[userid][j]
		})
	}

	// 0充玩家局数分布
	range1Player := make(map[string]int) // <1-20, round>
	range1PlayerIds := make(map[string][]string)
	var ranges1 [][2]int
	var ranges1Key []string
	ranges1Key0, ranges1KeyLimit := "0", fmt.Sprintf("%d+", limit1+1)

	for i := 0; i < limit1; i += range1 {
		rr := [2]int{i + 1, i + range1}
		ranges1 = append(ranges1, rr)
		ranges1Key = append(ranges1Key, fmt.Sprintf("%d-%d", rr[0], rr[1]))
	}
	for _, userid := range payTimesUser[0] {
		round := playerRounds[userid]
		if round == 0 {
			range1Player[ranges1Key0]++
			range1PlayerIds[ranges1Key0] = append(range1PlayerIds[ranges1Key0], userid)
			continue
		} else if round > limit1 {
			range1Player[ranges1KeyLimit]++
			range1PlayerIds[ranges1KeyLimit] = append(range1PlayerIds[ranges1KeyLimit], userid)
			continue
		}
		for i, rr := range ranges1 {
			key := ranges1Key[i]
			if round >= rr[0] && round <= rr[1] {
				range1Player[key]++
				range1PlayerIds[key] = append(range1PlayerIds[key], userid)
				break
			}
		}
	}

	// 有充值记录的
	chargeTimesRoundUsers := make(map[int]map[string]int)        // n充, 1-20局, 人数
	chargeTimesRoundUserIds := make(map[int]map[string][]string) // n充, 1-20局, 用户id列表
	chargeTimesRoundUsers[0] = range1Player                      // 0充玩家局数分布
	chargeTimesRoundUserIds[0] = range1PlayerIds

	var payedUserids []string // >pChargeTimes的玩家不包含
	for times, users := range payTimesUser {
		if times > 0 {
			payedUserids = append(payedUserids, users...)
		}
	}
	// userChargeTimes := make(map[string][]int64) // 玩家第1,2,3...次充值时间
	userChargeTimes := make(map[string]int)
	m5 := []bson.M{
		{"$match": bson.M{"userid": bson.M{"$in": payedUserids}, "order_status": 4, "ctime": bson.M{"$gt": startTime}}},
		{"$project": bson.M{"_id": "$_id", "userid": "$userid", "ctime": "$ctime"}},
		{"$sort": bson.M{"ctime": 1}},
	}
	var r5 []bson.M
	err = Pays.Pipe(m5).All(&r5)
	if err != nil {
		return
	}

	for _, r := range r5 {
		userid := r["userid"].(string)
		payTime := r["ctime"].(time.Time).Unix()
		// userPayTimes[userid] = append(userPayTimes[userid], payTime)
		// chargeTimes := len(userPayTimes[userid])
		userChargeTimes[userid]++
		chargeTimes := userChargeTimes[userid] // 该user第n次充值
		if chargeTimes > pChargeTimes {
			continue
		}
		if _, ok := chargeTimesRoundUsers[chargeTimes]; !ok {
			chargeTimesRoundUsers[chargeTimes] = make(map[string]int)
			chargeTimesRoundUserIds[chargeTimes] = make(map[string][]string)
		}

		// 支付时间 匹配玩家对局时间
		if len(playerRoundsTime[userid]) == 0 { // 0局充值
			chargeTimesRoundUsers[chargeTimes][ranges1Key0]++
			chargeTimesRoundUserIds[chargeTimes][ranges1Key0] = append(chargeTimesRoundUserIds[chargeTimes][ranges1Key0], userid)
			continue
		}
	roundLoop:
		for i := len(playerRoundsTime[userid]) - 1; i >= 0; i-- {
			roundTime := playerRoundsTime[userid][i]
			round := i + 1
			if payTime > int64(roundTime) {
				if round > limit1 {
					chargeTimesRoundUsers[chargeTimes][ranges1KeyLimit]++
					chargeTimesRoundUserIds[chargeTimes][ranges1KeyLimit] = append(chargeTimesRoundUserIds[chargeTimes][ranges1KeyLimit], userid)
					break roundLoop
				}
				for i, rr := range ranges1 {
					key := ranges1Key[i]
					if round >= rr[0] && round <= rr[1] {
						chargeTimesRoundUsers[chargeTimes][key]++
						chargeTimesRoundUserIds[chargeTimes][key] = append(chargeTimesRoundUserIds[chargeTimes][key], userid)
						break roundLoop
					}
				}
			}
		}
	}

	analysis1 = chargeTimesRoundUsers
	analysis1Users = chargeTimesRoundUserIds
	return
}

// behanviorAnalysis2BetAmount 表二 统计
// charge2 充值次数参数
func behanviorAnalysis2BetAmount(startTime time.Time, charge2, range2, limit2, charge3 int, userids []string, payTimesUser map[int][]string, payTimes map[string]int) (
	analysis2 map[int]map[string]int,
	analysis2Users map[int]map[string][]string,
	analysis3 map[int]map[int][]int, // n充,游戏类型,[打码量, 局数]
	err error,
) {
	endTime := utils.Str2Time(fmt.Sprintf("%s 00:00:00", time.Now().AddDate(0, 0, 1).Format("2006-01-02")), Location())

	// 时间段查
	stepDay := 1
	start, end := startTime, startTime.AddDate(0, 0, stepDay)

	// 打码量统计
	analysis3 = make(map[int]map[int][]int) // 充值次数, 游戏类型, [打码量, 局数]
	// userGtypeBets := make(map[]) // userid, 游戏类型, [打码量, 局数]
	betAmounts := make(map[string]int64)

	for ; start.Before(endTime); start, end = end, end.AddDate(0, 0, stepDay) {
		var r1 []bson.M
		// for page, size := 1, 1000; page == 1 || len(r1) > 0; page++ {
		// 	skipNum := (page - 1) * size
		// 	if skipNum < 0 {
		// 		skipNum = 0
		// 	}
		m1 := []bson.M{
			{"$match": bson.M{"begin_time": bson.M{"$gt": start.Unix(), "$lt": end.Unix()}, "players": bson.M{"$ne": ""}}},
			{"$project": bson.M{
				"begin_time": "$begin_time",
				"userids":    bson.M{"$split": []string{"$players", ","}},
				// "gtype":      0,
				"gtype":       "$gtype",
				"tpdetail":    bson.M{"$map": bson.M{"input": "$tpdetail", "as": "det", "in": bson.M{"user_id": "$$det.user_id", "bet": "$$det.bet"}}},    //"$tpdetail",
				"jokerdetail": bson.M{"$map": bson.M{"input": "$jokerdetail", "as": "det", "in": bson.M{"user_id": "$$det.user_id", "bet": "$$det.bet"}}}, // "$jokerdetail"
				"ak47detail":  bson.M{"$map": bson.M{"input": "$ak47detail", "as": "det", "in": bson.M{"user_id": "$$det.user_id", "bet": "$$det.bet"}}},  // "$ak47detail",
				"rmdetail":    bson.M{"$map": bson.M{"input": "$rmdetail", "as": "det", "in": bson.M{"user_id": "$$det.user_id", "score": "$$det.score"}}},
				"lhdetail":    bson.M{"user_detail": bson.M{"$map": bson.M{"input": "$lhdetail.user_detail", "as": "det", "in": bson.M{"userid": "$$det.userid", "dragon": "$$det.dragon", "tiger": "$$det.tiger", "tie": "$$det.tie"}}}},
				"updetail":    bson.M{"user_detail": bson.M{"$map": bson.M{"input": "$updetail.user_detail", "as": "det", "in": bson.M{"userid": "$$det.userid", "dragon": "$$det.dragon", "tiger": "$$det.tiger", "tie": "$$det.tie"}}}},
				"crashdetail": bson.M{"user_detail": bson.M{"$map": bson.M{"input": "$crashdetail.user_detail", "as": "det", "in": bson.M{"userid": "$$det.userid", "bet": "$$det.bet"}}}},
				"cpdetail":    bson.M{"user_detail": bson.M{"$map": bson.M{"input": "$cpdetail.user_detail", "as": "det", "in": bson.M{"userid": "$$det.userid", "seat_bets": "$$det.seat_bets"}}}},
				"abdetail":    bson.M{"user_detail": bson.M{"$map": bson.M{"input": "$abdetail.user_detail", "as": "det", "in": bson.M{"userid": "$$det.userid", "seat_bets": "$$det.seat_bets"}}}},
			}},
			{"$match": bson.M{"userids": bson.M{"$in": userids}}},
			// {"$sort": bson.M{"begin_time": 1}},
			// {"$skip": skipNum},
			// {"$limit": size},
		}
		// start.Format("2006-01-02") == "2024-05-08"
		// startStr := start.Format("2006-01-02")
		// fmt.Println(startStr)
		err = Details.Pipe(m1).All(&r1)
		if err != nil {
			return
		}

		for _, r := range r1 {
			gtype := r["gtype"].(int)
			roundBets := make(map[string]int64, 3) // 回合玩家下注列表
			switch gtype {
			// TP/JOKER/AK47：玩家每局总下注
			case int(pb.HUA), int(pb.HUA2):
				for _, item := range r["tpdetail"].([]interface{}) {
					d := item.(bson.M)
					userid := d["user_id"].(string)
					amount := d["bet"].(int64)
					if len(userid) < 16 { // 机器人id 18位
						roundBets[userid] += amount
					}
				}
			case int(pb.JOKER):
				for _, item := range r["jokerdetail"].([]interface{}) {
					d := item.(bson.M)
					userid := d["user_id"].(string)
					amount := d["bet"].(int64)
					if len(userid) < 16 {
						roundBets[userid] += amount
					}
				}
			case int(pb.AK47):
				for _, item := range r["ak47detail"].([]interface{}) {
					d := item.(bson.M)
					userid := d["user_id"].(string)
					amount := d["bet"].(int64)
					if len(userid) < 16 {
						roundBets[userid] += amount
					}
				}
			// RUMMY：玩家每局输赢绝对值
			case int(pb.RUMMY):
				for _, item := range r["rmdetail"].([]interface{}) {
					d := item.(bson.M)
					userid := d["user_id"].(string)
					amount := d["score"].(int64)
					if len(userid) < 16 {
						roundBets[userid] += int64(math.Abs(float64(amount)))
					}
				}
			// LHD/7UP/红黑大战：玩家每局总下注
			case int(pb.LHD):
				detail := r["lhdetail"].(bson.M)
				for _, item := range detail["user_detail"].([]interface{}) {
					d := item.(bson.M)
					userid := d["userid"].(string)
					dragon := d["dragon"].(int64)
					tiger := d["tiger"].(int64)
					tie := d["tie"].(int64)
					if len(userid) < 16 {
						roundBets[userid] += (dragon + tiger + tie)
					}
				}
			case int(pb.SEVEN):
				detail := r["updetail"].(bson.M)
				for _, item := range detail["user_detail"].([]interface{}) {
					d := item.(bson.M)
					userid := d["userid"].(string)
					dragon := d["dragon"].(int64)
					tiger := d["tiger"].(int64)
					tie := d["tie"].(int64)
					if len(userid) < 16 {
						roundBets[userid] += (dragon + tiger + tie)
					}
				}
			case int(pb.REDBLACK):
			// CRASH/飞机：玩家每局总下注
			case int(pb.CRASH):
				fallthrough
			case int(pb.PLANE):
				detail := r["crashdetail"].(bson.M)
				for _, item := range detail["user_detail"].([]interface{}) {
					d := item.(bson.M)
					userid := d["userid"].(string)
					bet := d["bet"].(int64)
					if len(userid) < 16 {
						roundBets[userid] += bet
					}
				}
			case int(pb.LOTTERY):
				detail := r["cpdetail"].(bson.M)
				for _, item := range detail["user_detail"].([]interface{}) {
					d := item.(bson.M)
					userid := d["userid"].(string)
					seatBets := d["seat_bets"].(bson.M)
					var bet int64
					for _, bets := range seatBets {
						bet += bets.(int64)
					}
					if len(userid) < 16 {
						roundBets[userid] += bet
					}
				}
			case int(pb.ABAR):
				detail := r["abdetail"].(bson.M)
				for _, item := range detail["user_detail"].([]interface{}) {
					d := item.(bson.M)
					userid := d["userid"].(string)
					seatBets := d["seat_bets"].(bson.M)
					var bet int64
					for _, bets := range seatBets {
						bet += bets.(int64)
					}
					if len(userid) < 16 {
						roundBets[userid] += bet
					}
				}
			}

			for userid, amount := range roundBets {
				betAmounts[userid] += amount
				userPayTime := payTimes[userid]
				for payTime := 0; payTime <= userPayTime; payTime++ { // 包含 0,1,2...payTime 的数据叠加
					if payTime > charge3 {
						// payTime = charge3 + 1 // bug
						break
					}

					if _, ok := analysis3[payTime]; !ok {
						analysis3[payTime] = make(map[int][]int)
					}
					if _, ok := analysis3[payTime][gtype]; !ok {
						analysis3[payTime][gtype] = []int{0, 0}
					}
					analysis3[payTime][gtype][0] += int(amount)
					analysis3[payTime][gtype][1]++
				}
				// if payTime > charge3 {
				// 	payTime = charge3 + 1
				// }

				// if _, ok := analysis3[payTime]; !ok {
				// 	analysis3[payTime] = make(map[int][]int)
				// }
				// if _, ok := analysis3[payTime][gtype]; !ok {
				// 	analysis3[payTime][gtype] = []int{0, 0}
				// }
				// analysis3[payTime][gtype][0] += int(amount)
				// analysis3[payTime][gtype][1]++
			}
			// }
		}
	}
	// 外接游戏
	m3 := []bson.M{
		{
			"$match": bson.M{
				"user_id": bson.M{"$in": userids},
				"amount":  bson.M{"$ne": 0},
			},
		},
		{
			"$group": bson.M{
				"_id":  bson.M{"game_id": "$game_id", "round_id": "$round_id", "user_id": "$user_id"},
				"bets": bson.M{"$sum": "$amount"},
			},
		},
	}
	var r3 []bson.M
	err = NsqLogExternalBets.Pipe(m3).All(&r3)
	if err != nil {
		return
	}
	for _, r := range r3 {
		userid := r["_id"].(bson.M)["user_id"].(string)
		game_id := r["_id"].(bson.M)["game_id"].(int)
		bets := r["bets"].(int64)
		payTime := payTimes[userid]
		if payTime > charge3 {
			payTime = charge3 + 1
		}
		if _, ok := analysis3[payTime]; !ok {
			analysis3[payTime] = make(map[int][]int)
		}
		if _, ok := analysis3[payTime][game_id]; !ok {
			analysis3[payTime][game_id] = []int{0, 0}
		}
		analysis3[payTime][game_id][0] += int(bets)
		analysis3[payTime][game_id][1]++
	}

	// 计算区间
	// range1Player := make(map[string]int) // <1-20, round>
	var ranges2 [][2]int
	var ranges2Key []string
	ranges2Key0, ranges2KeyLimit := "0", fmt.Sprintf("%d+", limit2+1)

	for i := 0; i < limit2; i += range2 {
		rr := [2]int{i*100 + 1, (i + range2) * 100}
		ranges2 = append(ranges2, rr)
		ranges2Key = append(ranges2Key, fmt.Sprintf("%d-%d", i+1, i+range2))
	}

	analysis2 = make(map[int]map[string]int) // n充, 1-20局, 人数
	analysis2Users = make(map[int]map[string][]string)
	for userPayTimes, users := range payTimesUser {
		for payTimes := 0; payTimes <= userPayTimes; payTimes++ { // 包含 0,1,2...payTime 的数据叠加
			if payTimes > charge2 {
				break
			}
			if _, ok := analysis2[payTimes]; !ok {
				analysis2[payTimes] = make(map[string]int)
				analysis2Users[payTimes] = make(map[string][]string)
			}
			for _, userid := range users {
				betAmount := int(betAmounts[userid])
				if betAmount == 0 {
					analysis2[payTimes][ranges2Key0]++
					analysis2Users[payTimes][ranges2Key0] = append(analysis2Users[payTimes][ranges2Key0], userid)
					continue
				} else if betAmount > limit2*100 {
					analysis2[payTimes][ranges2KeyLimit]++
					analysis2Users[payTimes][ranges2KeyLimit] = append(analysis2Users[payTimes][ranges2KeyLimit], userid)
					continue
				}
				for i, rr := range ranges2 {
					if betAmount >= rr[0] && betAmount <= rr[1] {
						analysis2[payTimes][ranges2Key[i]]++
						analysis2Users[payTimes][ranges2Key[i]] = append(analysis2Users[payTimes][ranges2Key[i]], userid)
						break
					}
				}
			}
		}

		// if payTimes > charge2 {
		// 	continue
		// }
		// if _, ok := analysis2[payTimes]; !ok {
		// 	analysis2[payTimes] = make(map[string]int)
		// }
		// for _, userid := range users {
		// 	betAmount := int(betAmounts[userid])
		// 	if betAmount == 0 {
		// 		analysis2[payTimes][ranges2Key0]++
		// 		continue
		// 	} else if betAmount > limit2 {
		// 		analysis2[payTimes][ranges2KeyLimit]++
		// 		continue
		// 	}
		// 	for i, rr := range ranges2 {
		// 		if betAmount >= rr[0] && betAmount <= rr[1] {
		// 			analysis2[payTimes][ranges2Key[i]]++
		// 			break
		// 		}
		// 	}
		// }
	}
	return
}

// 子项目收益
func (s *statistics2Service) SubprojectIncome() ([]entity.SubprojectIncome, error) {
	var list []entity.SubprojectIncome
	err := SubprojectIncomes.Find(bson.M{}).All(&list)
	list = s.chipList(list)
	return list, err
}

func (s *statistics2Service) chipList(list []entity.SubprojectIncome) []entity.SubprojectIncome {
	gamelist := map[int32]string{
		1:  "TP",
		2:  "DRAGON TIGER",
		3:  "7UPDOWN",
		4:  "RUMMY",
		5:  "AK47",
		6:  "JOKER",
		7:  "CRASH",
		8:  "ANDARBAHAR",
		9:  "彩票",
		10: "飞机",
		11: "红黑大战",
		12: "RUMMY双人",
		13: "TP2",
		14: "MINES",
	}
	var roomlist []entity.Game
	Games.Find(bson.M{}).All(&roomlist)
	for k, v := range list {
		gtype, ok := gamelist[v.Gtype]
		if ok {
			v.GameName = gtype
		} else {
			v.GameName = "--"
		}
		for _, r := range roomlist {
			if r.Id == v.RoomId {
				v.RoomName = r.Name
			}
		}
		v.BetAvg = ComputeFloat(v.BrBets, v.BrNumber) / 100.0
		if v.BetMultiple > 0 {
			v.FBetMultiple = v.BetMultiple / float64(v.WinNumber)
		}

		v.WinRate = ComputeFloat(v.WinNumber, v.GameNumber) * 100.0
		v.FSystemWin = Chip2Float(v.PlayerBet - v.WinBet)
		v.FTotalTax = Chip2Float(v.TotalTax)
		v.FTotalRevenue = Chip2Float(v.PlayerBet - v.WinBet + v.TotalTax)

		if v.LastReset > 0 {
			l, _ := ConvertToIndiaTime(v.LastReset)
			v.SLastReset = l
		}

		if v.RefreshTime > 0 {
			r, _ := ConvertToIndiaTime(v.RefreshTime)
			v.SRefreshTime = r
		}
		list[k] = v
	}
	return list
}

// 刷新子项目收益
func (s *statistics2Service) RefreshIncome() error {
	var gameid []string
	Stocks.Find(bson.M{}).Distinct("_id", &gameid)
	endTime := time.Now()
	for _, item := range gameid {
		// 获取该房间上次重置时间，如果没有则查询前一天数据
		startTime := endTime.AddDate(0, 0, -1)
		var info entity.SubprojectIncome
		SubprojectIncomes.Find(bson.M{"room_id": item}).One(&info)
		if info.Id != "" {
			if info.LastReset > 0 {
				startTime = utils.Stamp2Time(info.LastReset)
			}

		}
		// m := bson.M{
		// 	"room_id":  item,
		// 	"players":  bson.M{"$ne": ""},
		// 	"end_time": bson.M{"$gte": startTime.Unix(), "$lte": endTime.Unix()},
		// }
		m := bson.M{
			"$or": []interface{}{
				bson.M{"room_id": item, "players": bson.M{"$ne": ""}, "end_time": bson.M{"$gte": startTime.Unix(), "$lte": endTime.Unix()}},
				bson.M{"desk_id": item, "players": bson.M{"$ne": ""}, "end_time": bson.M{"$gte": startTime.Unix(), "$lte": endTime.Unix()}},
			},
		}

		var dlist []entity.Detail
		Details.Find(m).All(&dlist)
		map_list := make(map[string]*entity.SubprojectIncomeStats, 0)
		gtype := int32(0)
		for _, item := range dlist {
			gtype = item.Gtype
			switch item.Gtype {
			case 1:
				//tp
				for _, t := range item.TPDetail {
					if len(t.UserId) >= 16 {
						// 人机id18位长度+
						continue
					}
					stat, ok := map_list[t.UserId]
					if !ok {
						stat = &entity.SubprojectIncomeStats{
							UserId: t.UserId,
						}
						// crashYhwmIds = append(crashYhwmIds, user.Userid)
					}
					stat.GameNumber++
					stat.Bets += t.Bet
					if t.Score > 0 {
						stat.WinBets += t.Score
						stat.WinNumber++
						betavg := ComputeFloat(t.Score, t.Bet)
						stat.BetMultiple += betavg
					}
					stat.TotalTax += (t.CashAnTax + t.CashMingTax)
					map_list[t.UserId] = stat
				}
			case 2:
				// lhd
				for _, t := range item.LHDetail.UserDetail {
					if len(t.Userid) >= 16 {
						// 人机id18位长度+
						continue
					}
					stat, ok := map_list[t.Userid]
					if !ok {
						stat = &entity.SubprojectIncomeStats{
							UserId: t.Userid,
						}
						// crashYhwmIds = append(crashYhwmIds, user.Userid)
					}
					stat.GameNumber++
					bets := t.Dragon + t.Tiger + t.Tie
					stat.Bets += (bets)
					stat.BrBets = bets
					stat.BrNumber++
					if t.Win > 0 {
						stat.WinBets += t.Win
						stat.WinNumber++
						betavg := ComputeFloat(t.Win, bets)
						stat.BetMultiple += betavg
					}
					stat.TotalTax += (t.CashAnTax + t.CashMingTax)
					map_list[t.Userid] = stat
				}
			case 3:
				// 7up
				for _, t := range item.UPDetail.UserDetail {
					if len(t.Userid) >= 16 {
						// 人机id18位长度+
						continue
					}
					stat, ok := map_list[t.Userid]
					if !ok {
						stat = &entity.SubprojectIncomeStats{
							UserId: t.Userid,
						}
						// crashYhwmIds = append(crashYhwmIds, user.Userid)
					}
					stat.GameNumber++
					bets := t.Dragon + t.Tiger + t.Tie
					stat.Bets += (bets)
					stat.BrBets = bets
					stat.BrNumber++
					if t.Win > 0 {
						stat.WinBets += t.Win
						stat.WinNumber++
						betavg := ComputeFloat(t.Win, bets)
						stat.BetMultiple += betavg
					}
					stat.TotalTax += (t.CashAnTax + t.CashMingTax)
					map_list[t.Userid] = stat
				}
			case 4:
				// rummy
				for _, t := range item.RMDetail {
					if len(t.UserId) >= 16 {
						// 人机id18位长度+
						continue
					}
					stat, ok := map_list[t.UserId]
					if !ok {
						stat = &entity.SubprojectIncomeStats{
							UserId: t.UserId,
						}
						// crashYhwmIds = append(crashYhwmIds, user.Userid)
					}
					stat.GameNumber++
					stat.Bets += t.Score
					if t.Score > 0 {
						stat.WinBets += t.Score
						stat.WinNumber++
						betavg := ComputeFloat(t.Score, t.Score)
						stat.BetMultiple += betavg
					}
					stat.TotalTax += (t.CashAnTax + t.CashMingTax)
					map_list[t.UserId] = stat
				}
			case 5:
				// ak47
				for _, t := range item.AK47Detail {
					if len(t.UserId) >= 16 {
						// 人机id18位长度+
						continue
					}
					stat, ok := map_list[t.UserId]
					if !ok {
						stat = &entity.SubprojectIncomeStats{
							UserId: t.UserId,
						}
						// crashYhwmIds = append(crashYhwmIds, user.Userid)
					}
					stat.GameNumber++
					stat.Bets += t.Bet
					if t.Score > 0 {
						stat.WinBets += t.Score
						stat.WinNumber++
						betavg := ComputeFloat(t.Score, t.Bet)
						stat.BetMultiple += betavg
					}
					stat.TotalTax += (t.CashAnTax + t.CashMingTax)
					map_list[t.UserId] = stat
				}
			case 6:
				// joker
				for _, t := range item.JOKERDetail {
					if len(t.UserId) >= 16 {
						// 人机id18位长度+
						continue
					}
					stat, ok := map_list[t.UserId]
					if !ok {
						stat = &entity.SubprojectIncomeStats{
							UserId: t.UserId,
						}
						// crashYhwmIds = append(crashYhwmIds, user.Userid)
					}
					stat.GameNumber++
					stat.Bets += t.Bet
					if t.Score > 0 {
						stat.WinBets += t.Score
						stat.WinNumber++
						betavg := ComputeFloat(t.Score, t.Bet)
						stat.BetMultiple += betavg
					}
					stat.TotalTax += (t.CashAnTax + t.CashMingTax)
					map_list[t.UserId] = stat
				}
			case 7:
				// crash
				for _, t := range item.CRASHDetail.UserDetail {
					if len(t.Userid) >= 16 {
						// 人机id18位长度+
						continue
					}
					stat, ok := map_list[t.Userid]
					if !ok {
						stat = &entity.SubprojectIncomeStats{
							UserId: t.Userid,
						}
						// crashYhwmIds = append(crashYhwmIds, user.Userid)
					}
					stat.GameNumber++
					stat.Bets += (t.Bet)
					stat.BrBets = t.Bet
					stat.BrNumber++
					if t.Win > 0 {
						stat.WinBets += t.Win
						stat.WinNumber++
						betavg := ComputeFloat(t.Win, t.Bet)
						stat.BetMultiple += betavg
					}
					stat.TotalTax += (t.CashAnTax + t.CashMingTax)
					map_list[t.Userid] = stat
				}
			case 8:
				// ab
				for _, t := range item.ABDetail.UserDetail {
					if len(t.Userid) >= 16 {
						// 人机id18位长度+
						continue
					}
					stat, ok := map_list[t.Userid]
					if !ok {
						stat = &entity.SubprojectIncomeStats{
							UserId: t.Userid,
						}
						// crashYhwmIds = append(crashYhwmIds, user.Userid)
					}
					stat.GameNumber++
					bets := item.ABDetail.Bets
					stat.Bets += (bets)
					stat.BrBets = bets
					stat.BrNumber++
					if t.Win > 0 {
						stat.WinBets += t.Win
						stat.WinNumber++
						betavg := ComputeFloat(t.Win, bets)
						stat.BetMultiple += betavg
					}
					stat.TotalTax += (t.CashAnTax + t.CashMingTax)
					map_list[t.Userid] = stat
				}
			case 9:
				// 彩票
				for _, t := range item.CPDetail.UserDetail {
					if len(t.Userid) >= 16 {
						// 人机id18位长度+
						continue
					}
					stat, ok := map_list[t.Userid]
					if !ok {
						stat = &entity.SubprojectIncomeStats{
							UserId: t.Userid,
						}
						// crashYhwmIds = append(crashYhwmIds, user.Userid)
					}
					stat.GameNumber++
					bets := item.CPDetail.Bets
					stat.Bets += (bets)
					stat.BrBets = bets
					stat.BrNumber++
					if t.Win > 0 {
						stat.WinBets += t.Win
						stat.WinNumber++
						betavg := ComputeFloat(t.Win, bets)
						stat.BetMultiple += betavg
					}
					stat.TotalTax += (t.CashAnTax + t.CashMingTax)
					map_list[t.Userid] = stat
				}
			case 10:
				// 飞机
				for _, t := range item.CRASHDetail.UserDetail {
					if len(t.Userid) >= 16 {
						// 人机id18位长度+
						continue
					}
					stat, ok := map_list[t.Userid]
					if !ok {
						stat = &entity.SubprojectIncomeStats{
							UserId: t.Userid,
						}
						// crashYhwmIds = append(crashYhwmIds, user.Userid)
					}
					stat.GameNumber++
					stat.Bets += (t.Bet)
					stat.BrBets = t.Bet
					stat.BrNumber++
					if t.Win > 0 {
						stat.WinBets += t.Win
						stat.WinNumber++
						betavg := ComputeFloat(t.Win, t.Bet)
						stat.BetMultiple += betavg
					}
					stat.TotalTax += (t.CashAnTax + t.CashMingTax)
					map_list[t.Userid] = stat
				}
			}

		}
		// 汇总每个玩家数据

		if len(map_list) > 0 {
			// 存在数据的时候，需要重置数据
			if info.Id != "" {
				info.GameNumber = 0
				info.WinNumber = 0
				info.PlayerBet = 0
				info.WinBet = 0
				info.BetMultiple = 0
				info.TotalTax = 0
				info.BrBets = 0
				info.BrNumber = 0
			}
			for _, v := range map_list {
				info.Gtype = gtype
				info.RoomId = item
				info.GameNumber += v.GameNumber
				info.WinNumber += v.WinNumber
				info.PlayerBet += v.Bets
				info.WinBet += v.WinBets
				info.BetMultiple += v.BetMultiple
				info.TotalTax += v.TotalTax
				info.BrBets += v.BrBets
				info.BrNumber += v.BrNumber
				// if v.BrNumber > 0 && v.BrBets > 0{
				// 	betavg :=
				// 	info.BetAvg +=
				// }
			}
			if info.Id != "" {
				n := bson.M{"_id": info.Id}
				u := bson.M{
					"game_number":  info.GameNumber,
					"win_number":   info.WinNumber,
					"lose_number":  info.LoseNumber,
					"player_bet":   info.PlayerBet,
					"win_bet":      info.WinBet,
					"lose_bet":     info.LoseBet,
					"total_tax":    info.TotalTax,
					"bet_avg":      info.BetAvg,
					"bet_multiple": info.BetMultiple,
					"win_rate":     info.WinRate,
					"system_win":   info.SystemWin,
					"br_number":    info.BrNumber,
					"br_bets":      info.BrBets,
					"refresh_time": endTime.Unix(),
				}
				Update(SubprojectIncomes, n, bson.M{"$set": u})
			} else {
				info.Id = bson.NewObjectId().Hex()
				info.RefreshTime = endTime.Unix()
				Insert(SubprojectIncomes, info)
			}
		}
		// pipeline := []bson.M{
		// 	{
		// 		"$match": bson.M{
		// 			"room_id":  item,
		// 			"end_time": bson.M{"$gte": startTime.Unix(), "$lte": endTime.Unix()},
		// 		},
		// 	},
		// 	{
		// 		"$group": bson.M{
		// 			"_id": "$pay_channel",
		// 			"ds_amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
		// 				bson.M{"$eq": []interface{}{"$p_type", 1}}, "$amount", 0,
		// 			}}},
		// 			"ds_rate": bson.M{"$last": bson.M{"$cond": []interface{}{
		// 				bson.M{"$eq": []interface{}{"$p_type", 1}}, "$rate", 0.0,
		// 			}}},
		// 			"ds_sj_amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
		// 				bson.M{"$eq": []interface{}{"$p_type", 1}}, "$actual_amount", 0,
		// 			}}},
		// 			"df_amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
		// 				bson.M{"$eq": []interface{}{"$p_type", 2}}, "$amount", 0,
		// 			}}},
		// 			"df_rate": bson.M{"$last": bson.M{"$cond": []interface{}{
		// 				bson.M{"$eq": []interface{}{"$p_type", 2}}, "$rate", 0.0,
		// 			}}},
		// 			"df_sj_amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
		// 				bson.M{"$eq": []interface{}{"$p_type", 2}}, "$actual_amount", 0,
		// 			}}},

		// 			"ds_handling": bson.M{"$sum": bson.M{"$cond": []interface{}{
		// 				bson.M{"$eq": []interface{}{"$p_type", 1}}, "$handling_charge", 0,
		// 			}}},
		// 			"df_handling": bson.M{"$sum": bson.M{"$cond": []interface{}{
		// 				bson.M{"$eq": []interface{}{"$p_type", 2}}, "$handling_charge", 0,
		// 			}}},
		// 		},
		// 	},
		// }
	}
	return nil
}

// 重置子项目收益
func (s *statistics2Service) ResetIncome(gid, roomId string) error {
	query := bson.M{}
	var err error
	if gid != "" && roomId != "" {
		// 重置单个数据
		gameid, _ := strconv.Atoi(gid)
		query["gtype"] = gameid
		query["room_id"] = roomId
	}
	var list []entity.SubprojectIncome
	err = SubprojectIncomes.Find(query).All(&list)
	for _, v := range list {
		n := bson.M{
			"game_number":  0,
			"win_number":   0,
			"lose_number":  0,
			"player_bet":   0,
			"win_bet":      0,
			"lose_bet":     0,
			"total_tax":    0,
			"bet_avg":      0,
			"bet_multiple": 0,
			"win_rate":     0,
			"system_win":   0,
			"br_number":    0,
			"br_bets":      0,
			"last_reset":   time.Now().Unix(),
		}
		m := bson.M{"_id": v.Id}
		Update(SubprojectIncomes, m, bson.M{"$set": n})
	}
	return err
}

// 游戏状态
func (s *statistics2Service) GetGameState() ([]*entity.GameStateData, error) {
	var list []*entity.GameStateData
	maplist := make(map[string]*entity.GameStateData, 0)
	result, err := GmRequest(pb.WebOnlineUser, pb.CONFIG_UPSERT, pb.NULL)
	if err != nil {
		return list, err
	}
	userlist, _ := ConvertToUserSlice(result)
	for _, item := range userlist {
		sortId, _ := strconv.Atoi(item.GameId)
		stat, ok := maplist[item.GameId]
		if !ok {
			stat = &entity.GameStateData{
				GameId: item.GameId,
				SortId: sortId,
			}
		}
		stat.Number++
		if item.Recharge > 0 {
			stat.PayNumber++
		}
		maplist[item.GameId] = stat
	}
	roomList := map[string]string{
		"-1": "全部",
		"0":  "大厅",
		"1":  "TP",
		"2":  "DRAGON TIGER",
		"3":  "7UPDOWN",
		"4":  "RUMMY",
		"5":  "TP-AK47",
		"6":  "TP-JOKER",
		"7":  "CRASH",
		"8":  "ANDARBAHAR",
		"9":  "彩票",
		"10": "飞机",
		"11": "红黑大战",
		"12": "RUMMY双人",
		"13": "TP2",
		"14": "MINES",
	}
	totalInfo := &entity.GameStateData{
		GameId:   "-1",
		GameName: "全部",
		SortId:   -1,
	}
	for idx, item := range maplist {
		for k, v := range roomList {
			if idx == k {
				item.GameName = v
			}
		}
		totalInfo.Number += item.Number
		totalInfo.PayNumber += item.PayNumber
		list = append(list, item)
	}
	// 使用 sort.Slice() 进行排序
	sort.Slice(list, func(i, j int) bool {
		return list[i].SortId < list[j].SortId
	})
	list = append([]*entity.GameStateData{totalInfo}, list...)
	return list, nil
}

// 将interface{}转成[]
func ConvertToUserSlice(data interface{}) ([]entity.OnlineUser, bool) {
	if slice, ok := data.([]entity.OnlineUser); ok {
		return slice, true
	}
	return nil, false
}

// 排行榜
func (s *statistics2Service) GetRankingList(startDate string) (map[int][]entity.RankingList, time.Time, error) {
	list := make(map[int][]entity.RankingList, 0)
	Refreshtime := time.Now()
	layout := "2006-01-02" // 指定日期的格式

	// 将字符串解析为时间
	date, err := time.Parse(layout, startDate)
	if err != nil {
		fmt.Println("日期解析失败:", err)
	}
	// 获取当天的日期
	today := time.Now().UTC().Truncate(24 * time.Hour)
	// 比较日期
	if date.Equal(today) {
		// 查询当天实时数据
		startTime := utils.Str2Time(fmt.Sprintf("%s 00:00:00", startDate), Location())
		endTime := utils.Str2Time(fmt.Sprintf("%s 23:59:59", startDate), Location())
		pay := s.GetPayRanking(startTime, endTime)
		with := s.GetWithDrawRanking(startTime, endTime)
		winlist := s.GetWinRanking(startTime, endTime)
		loselist := s.GetLoseRanking(startTime, endTime)
		paycarry := s.GetPayCarryRanking(startTime, endTime)
		fzcarry := s.GetFZCarryRanking(startTime, endTime)
		list[0] = pay
		list[1] = with
		list[2] = winlist
		list[3] = loselist
		list[4] = paycarry
		list[5] = fzcarry
		Refreshtime = time.Now()
	} else {
		// 查询往期数据
		loc, _ := time.LoadLocation(locationName)
		sdate := utils.TimestampAppiontZero(date, loc)
		edate := utils.TimestampAppiontZero(date.AddDate(0, 0, 1), loc)
		m := bson.M{}
		m["date"] = bson.M{"$gte": sdate, "$lt": edate}
		var res []entity.RankingListData
		RankingLists.Find(m).All(&res)
		pay := make([]entity.RankingList, 0)
		with := make([]entity.RankingList, 0)
		winlist := make([]entity.RankingList, 0)
		loselist := make([]entity.RankingList, 0)
		paycarry := make([]entity.RankingList, 0)
		fzcarry := make([]entity.RankingList, 0)
		if len(res) > 0 {
			c, _ := ConvertToIndiaTime(res[0].ETime)
			Refreshtime = c
		}
		for _, item := range res {
			info := new(entity.RankingList)
			info.Id = item.No
			info.Userid = item.Userid
			info.CarryAmount = item.CarryAmount
			info.PayAmount = item.PayAmount
			info.ProfitAmount = item.ProfitAmount
			info.WithdrawAmount = item.WithdrawAmount
			info.FPayAmount = Chip2Float(item.PayAmount)
			info.FWithdrawAmount = Chip2Float(item.WithdrawAmount)
			info.FCarryAmount = Chip2Float(item.CarryAmount)
			info.FProfitAmount = Chip2Float(item.ProfitAmount)
			info.RegistArea = item.RegistArea
			switch item.RType {
			case 0:
				pay = append(pay, *info)
			case 1:
				with = append(with, *info)
			case 2:
				winlist = append(winlist, *info)
			case 3:
				loselist = append(loselist, *info)
			case 4:
				paycarry = append(paycarry, *info)
			case 5:
				fzcarry = append(fzcarry, *info)
			}
		}
		sort.Slice(pay, func(i, j int) bool {
			return pay[i].Id < pay[j].Id
		})
		sort.Slice(with, func(i, j int) bool {
			return with[i].Id < with[j].Id
		})
		sort.Slice(winlist, func(i, j int) bool {
			return winlist[i].Id < winlist[j].Id
		})
		sort.Slice(loselist, func(i, j int) bool {
			return loselist[i].Id < loselist[j].Id
		})
		sort.Slice(paycarry, func(i, j int) bool {
			return paycarry[i].Id < paycarry[j].Id
		})
		sort.Slice(fzcarry, func(i, j int) bool {
			return fzcarry[i].Id < fzcarry[j].Id
		})
		list[0] = pay
		list[1] = with
		list[2] = winlist
		list[3] = loselist
		list[4] = paycarry
		list[5] = fzcarry
	}
	return list, Refreshtime, nil
}

// 排行榜-付费榜
func (s *statistics2Service) GetPayRanking(startTime, endTime time.Time) []entity.RankingList {
	pay := make([]entity.RankingList, 0)
	payList := make(map[string]entity.RankingList, 0)
	// 付费榜
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"order_status": 4,
				"ctime":        bson.M{"$gte": startTime, "$lte": endTime},
			},
		},
		{
			"$group": bson.M{
				"_id":    "$userid",
				"amount": bson.M{"$sum": "$amount"},
			},
		},
		{"$sort": bson.M{"amount": -1}},
		{"$limit": 20},
	}
	var pay_res []bson.M
	Pays.Pipe(pipeline).All(&pay_res)
	var pay_ids []string
	for i, item := range pay_res {
		pid := item["_id"].(string)
		amount := item["amount"].(int)
		info := new(entity.RankingList)
		info.Id = (i + 1)
		info.Userid = pid
		info.PayAmount = int64(amount)
		info.FPayAmount = Chip2Float(int64(amount))
		pay_ids = append(pay_ids, pid)
		payList[pid] = *info
	}
	// 获取提现
	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"order_status": 2,
				"userid":       bson.M{"$in": pay_ids},
				"ctime":        bson.M{"$gte": startTime, "$lte": endTime},
			},
		},
		{
			"$group": bson.M{
				"_id":    "$userid",
				"amount": bson.M{"$sum": "$amount"},
			},
		},
	}
	var withdraw_res []bson.M
	Withdraws.Pipe(pipeline1).All(&withdraw_res)
	for _, with := range withdraw_res {
		wid := with["_id"].(string)
		stat, ok := payList[wid]
		if ok {
			amount := with["amount"].(int)
			stat.WithdrawAmount = int64(amount)
			stat.FWithdrawAmount = Chip2Float(int64(amount))
			payList[wid] = stat
		}
	}

	// 获取携带
	pipeline2 := []bson.M{
		{
			"$match": bson.M{
				"robot":            false,
				"simulation_robot": false,
				"_id":              bson.M{"$in": pay_ids},
			},
		},
		{
			"$group": bson.M{
				"_id":            "$_id",
				"regist_area":    bson.M{"$first": "$regist_area"},
				"diamond":        bson.M{"$sum": "$diamond"},
				"shadow_diamond": bson.M{"$sum": "$shadow_diamond"},
				"coin":           bson.M{"$sum": "$coin"},
			},
		},
	}
	var xiedai_res []bson.M
	PlayerUsers.Pipe(pipeline2).All(&xiedai_res)
	for _, xd := range xiedai_res {
		uid := xd["_id"].(string)
		stat, ok := payList[uid]
		if ok {
			diamond := xd["diamond"].(int64)
			shadow_diamond := xd["shadow_diamond"].(int64)
			coin := xd["coin"].(int64)
			regist_area := 0
			dsId, _ := xd["regist_area"]
			if dsId == nil {
				regist_area = 0
			} else {
				regist_area = dsId.(int)
			}
			amount := diamond + shadow_diamond + coin
			stat.CarryAmount = int64(amount)
			stat.FCarryAmount = Chip2Float(int64(amount))
			stat.RegistArea = regist_area
			// 计算盈利
			profit := (stat.WithdrawAmount + stat.CarryAmount) - stat.PayAmount
			stat.ProfitAmount = profit
			stat.FProfitAmount = Chip2Float(profit)
			payList[uid] = stat
		}
	}

	for _, v := range payList {
		pay = append(pay, v)
	}
	sort.Slice(pay, func(i, j int) bool {
		return pay[i].Id < pay[j].Id
	})
	return pay
}

// 排行榜-提现榜
func (s *statistics2Service) GetWithDrawRanking(startTime, endTime time.Time) []entity.RankingList {
	withdraw := make([]entity.RankingList, 0)
	withdrawList := make(map[string]entity.RankingList, 0)
	// 提现榜
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"order_status": 2,
				"ctime":        bson.M{"$gte": startTime, "$lte": endTime},
			},
		},
		{
			"$group": bson.M{
				"_id":    "$userid",
				"amount": bson.M{"$sum": "$amount"},
			},
		},
		{"$sort": bson.M{"amount": -1}},
		{"$limit": 20},
	}
	var pay_res []bson.M
	Withdraws.Pipe(pipeline).All(&pay_res)
	var pay_ids []string
	for i, item := range pay_res {
		pid := item["_id"].(string)
		amount := item["amount"].(int)
		info := new(entity.RankingList)
		info.Id = (i + 1)
		info.Userid = pid
		info.WithdrawAmount = int64(amount)
		info.FWithdrawAmount = Chip2Float(int64(amount))
		pay_ids = append(pay_ids, pid)
		withdrawList[pid] = *info
	}
	// 获取充值
	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"order_status": 4,
				"userid":       bson.M{"$in": pay_ids},
				"ctime":        bson.M{"$gte": startTime, "$lte": endTime},
			},
		},
		{
			"$group": bson.M{
				"_id":    "$userid",
				"amount": bson.M{"$sum": "$amount"},
			},
		},
	}
	var withdraw_res []bson.M
	Pays.Pipe(pipeline1).All(&withdraw_res)
	for _, with := range withdraw_res {
		wid := with["_id"].(string)
		stat, ok := withdrawList[wid]
		if ok {
			amount := with["amount"].(int)
			stat.PayAmount = int64(amount)
			stat.FPayAmount = Chip2Float(int64(amount))
			withdrawList[wid] = stat
		}

	}
	// 获取携带
	pipeline2 := []bson.M{
		{
			"$match": bson.M{
				"robot":            false,
				"simulation_robot": false,
				"_id":              bson.M{"$in": pay_ids},
			},
		},
		{
			"$group": bson.M{
				"_id":            "$_id",
				"regist_area":    bson.M{"$first": "$regist_area"},
				"diamond":        bson.M{"$sum": "$diamond"},
				"shadow_diamond": bson.M{"$sum": "$shadow_diamond"},
				"coin":           bson.M{"$sum": "$coin"},
			},
		},
	}
	var xiedai_res []bson.M
	PlayerUsers.Pipe(pipeline2).All(&xiedai_res)
	for _, xd := range xiedai_res {
		uid := xd["_id"].(string)
		stat, ok := withdrawList[uid]
		if ok {
			diamond := xd["diamond"].(int64)
			shadow_diamond := xd["shadow_diamond"].(int64)
			coin := xd["coin"].(int64)
			regist_area := 0
			dsId, _ := xd["regist_area"]
			if dsId == nil {
				regist_area = 0
			} else {
				regist_area = dsId.(int)
			}
			amount := diamond + shadow_diamond + coin
			stat.CarryAmount = int64(amount)
			stat.FCarryAmount = Chip2Float(int64(amount))
			stat.RegistArea = regist_area
			// 计算盈利
			profit := (stat.WithdrawAmount + stat.CarryAmount) - stat.PayAmount
			stat.ProfitAmount = profit
			stat.FProfitAmount = Chip2Float(profit)
			withdrawList[uid] = stat
		}
	}

	for _, v := range withdrawList {
		withdraw = append(withdraw, v)
	}
	sort.Slice(withdraw, func(i, j int) bool {
		return withdraw[i].Id < withdraw[j].Id
	})
	return withdraw
}

// 排行榜-赢榜
func (s *statistics2Service) GetWinRanking(startTime, endTime time.Time) []entity.RankingList {
	list := make([]entity.RankingList, 0)
	map_list := make(map[string]entity.RankingList, 0)
	// 赢
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": startTime, "$lte": endTime},
				"ltype": bson.M{"$in": []int{110, 111, 75, 67, 69, 76, 70, 72, 116, 117, 114, 115, 112, 113, 79, 73, 86, 88, 89, 90, 91, 92, 100, 101, 93, 94, 95, 105}},
			},
		},
		{
			"$group": bson.M{
				"_id":             "$userid",
				"add_diamond":     bson.M{"$sum": "$add_diamond"},
				"add_coin":        bson.M{"$sum": "$add_coin"},
				"add_other_asset": bson.M{"$sum": "$add_other_asset"},
			},
		},
		{
			"$project": bson.M{
				"total": bson.M{"$add": []interface{}{"$add_diamond", "$add_coin", "$add_other_asset"}},
			},
		},
		{
			"$sort": bson.M{"total": -1},
		},
		{
			"$limit": 20,
		},
	}
	var xiedai_res []bson.M
	var s_ids []string
	LogWaters.Pipe(pipeline).All(&xiedai_res)
	for i, xd := range xiedai_res {
		uid := xd["_id"].(string)
		amount := xd["total"].(int64)
		info := new(entity.RankingList)
		info.Id = (i + 1)
		info.Userid = uid
		info.CarryAmount = int64(amount)
		info.FCarryAmount = Chip2Float(int64(amount))
		s_ids = append(s_ids, uid)
		map_list[uid] = *info
	}

	// 获取充值
	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"order_status": 4,
				"userid":       bson.M{"$in": s_ids},
				"ctime":        bson.M{"$gte": startTime, "$lte": endTime},
			},
		},
		{
			"$group": bson.M{
				"_id":    "$userid",
				"amount": bson.M{"$sum": "$amount"},
			},
		},
	}
	var pay_res []bson.M
	Pays.Pipe(pipeline1).All(&pay_res)
	for _, with := range pay_res {
		wid := with["_id"].(string)
		stat, ok := map_list[wid]
		if ok {
			amount := with["amount"].(int)
			stat.PayAmount = int64(amount)
			stat.FPayAmount = Chip2Float(int64(amount))
			map_list[wid] = stat
		}
	}

	// 获取提现
	pipeline2 := []bson.M{
		{
			"$match": bson.M{
				"order_status": 2,
				"userid":       bson.M{"$in": s_ids},
				"ctime":        bson.M{"$gte": startTime, "$lte": endTime},
			},
		},
		{
			"$group": bson.M{
				"_id":    "$userid",
				"amount": bson.M{"$sum": "$amount"},
			},
		},
	}
	var with_res []bson.M
	Withdraws.Pipe(pipeline2).All(&with_res)
	for _, with := range with_res {
		wid := with["_id"].(string)
		stat, ok := map_list[wid]
		if ok {
			amount := with["amount"].(int)
			stat.WithdrawAmount = int64(amount)
			stat.FWithdrawAmount = Chip2Float(int64(amount))
			map_list[wid] = stat
		}
	}
	// 获取用户类型
	pipeline3 := []bson.M{
		{
			"$match": bson.M{
				"robot":            false,
				"simulation_robot": false,
				"_id":              bson.M{"$in": s_ids},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"_id":         "$_id",
					"regist_area": "$regist_area",
				},
			},
		},
		{
			"$project": bson.M{
				"_id":         "$_id._id",
				"regist_area": "$_id.regist_area",
			},
		},
	}
	var user_res []bson.M
	PlayerUsers.Pipe(pipeline3).All(&user_res)
	for _, xd := range user_res {
		uid := xd["_id"].(string)
		stat, ok := map_list[uid]
		if ok {
			regist_area := 0
			dsId, _ := xd["regist_area"]
			if dsId == nil {
				regist_area = 0
			} else {
				regist_area = dsId.(int)
			}
			stat.RegistArea = regist_area
			map_list[uid] = stat
		}
	}
	for _, v := range map_list {
		profit := (v.WithdrawAmount + v.CarryAmount) - v.PayAmount
		v.ProfitAmount = profit
		v.FProfitAmount = Chip2Float(profit)
		list = append(list, v)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Id < list[j].Id
	})
	return list
}

// 排行榜-输榜
func (s *statistics2Service) GetLoseRanking(startTime, endTime time.Time) []entity.RankingList {
	list := make([]entity.RankingList, 0)
	map_list := make(map[string]entity.RankingList, 0)

	// 输
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"ctime": bson.M{"$gte": startTime, "$lte": endTime},
				"ltype": bson.M{"$in": []int{110, 111, 75, 67, 69, 76, 70, 72, 116, 117, 114, 115, 112, 113, 79, 73, 86, 88, 89, 90, 91, 92, 100, 101, 93, 94, 95, 105}},
			},
		},
		{
			"$group": bson.M{
				"_id":             "$userid",
				"add_diamond":     bson.M{"$sum": "$add_diamond"},
				"add_coin":        bson.M{"$sum": "$add_coin"},
				"add_other_asset": bson.M{"$sum": "$add_other_asset"},
			},
		},
		{
			"$project": bson.M{
				"total": bson.M{"$add": []interface{}{"$add_diamond", "$add_coin", "$add_other_asset"}},
			},
		},
		{
			"$sort": bson.M{"total": 1},
		},
		{
			"$limit": 20,
		},
	}
	var xiedai_res []bson.M
	var s_ids []string
	LogWaters.Pipe(pipeline).All(&xiedai_res)
	for i, xd := range xiedai_res {
		uid := xd["_id"].(string)
		amount := xd["total"].(int64)
		info := new(entity.RankingList)
		info.Id = (i + 1)
		info.Userid = uid
		info.CarryAmount = int64(amount)
		info.FCarryAmount = Chip2Float(int64(amount))
		s_ids = append(s_ids, uid)
		map_list[uid] = *info
	}
	// 获取充值
	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"order_status": 4,
				"userid":       bson.M{"$in": s_ids},
				"ctime":        bson.M{"$gte": startTime, "$lte": endTime},
			},
		},
		{
			"$group": bson.M{
				"_id":    "$userid",
				"amount": bson.M{"$sum": "$amount"},
			},
		},
	}
	var pay_res []bson.M
	Pays.Pipe(pipeline1).All(&pay_res)
	for _, with := range pay_res {
		wid := with["_id"].(string)
		stat, ok := map_list[wid]
		if ok {
			amount := with["amount"].(int)
			stat.PayAmount = int64(amount)
			stat.FPayAmount = Chip2Float(int64(amount))
			map_list[wid] = stat
		}
	}

	// 获取提现
	pipeline2 := []bson.M{
		{
			"$match": bson.M{
				"order_status": 2,
				"userid":       bson.M{"$in": s_ids},
				"ctime":        bson.M{"$gte": startTime, "$lte": endTime},
			},
		},
		{
			"$group": bson.M{
				"_id":    "$userid",
				"amount": bson.M{"$sum": "$amount"},
			},
		},
	}
	var with_res []bson.M
	Withdraws.Pipe(pipeline2).All(&with_res)
	for _, with := range with_res {
		wid := with["_id"].(string)
		stat, ok := map_list[wid]
		if ok {
			amount := with["amount"].(int)
			stat.WithdrawAmount = int64(amount)
			stat.FWithdrawAmount = Chip2Float(int64(amount))
			map_list[wid] = stat
		}
	}
	// 获取用户类型
	pipeline3 := []bson.M{
		{
			"$match": bson.M{
				"robot":            false,
				"simulation_robot": false,
				"_id":              bson.M{"$in": s_ids},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"_id":         "$_id",
					"regist_area": "$regist_area",
				},
			},
		},
		{
			"$project": bson.M{
				"_id":         "$_id._id",
				"regist_area": "$_id.regist_area",
			},
		},
	}
	var user_res []bson.M
	PlayerUsers.Pipe(pipeline3).All(&user_res)
	for _, xd := range user_res {
		uid := xd["_id"].(string)
		stat, ok := map_list[uid]
		if ok {
			regist_area := 0
			dsId, _ := xd["regist_area"]
			if dsId == nil {
				regist_area = 0
			} else {
				regist_area = dsId.(int)
			}
			stat.RegistArea = regist_area
			map_list[uid] = stat
		}
	}
	for _, v := range map_list {
		profit := (v.WithdrawAmount + v.CarryAmount) - v.PayAmount
		v.ProfitAmount = profit
		v.FProfitAmount = Chip2Float(profit)
		list = append(list, v)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Id < list[j].Id
	})
	return list
}

// 排行榜-付费携带榜
func (s *statistics2Service) GetPayCarryRanking(startTime, endTime time.Time) []entity.RankingList {
	list := make([]entity.RankingList, 0)
	map_list := make(map[string]entity.RankingList, 0)
	// 获取存在充值的玩家
	n := bson.M{}
	n["ctime"] = bson.M{"$gte": startTime, "$lte": endTime}
	n["order_status"] = 4
	var pay_ids []string
	Pays.Find(n).Distinct("userid", &pay_ids)

	// 获取携带
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"robot":            false,
				"simulation_robot": false,
				"_id":              bson.M{"$in": pay_ids},
			},
		},
		{
			"$group": bson.M{
				"_id":            "$_id",
				"regist_area":    bson.M{"$first": "$regist_area"},
				"diamond":        bson.M{"$sum": "$diamond"},
				"shadow_diamond": bson.M{"$sum": "$shadow_diamond"},
				"coin":           bson.M{"$sum": "$coin"},
			},
		},
		{
			"$project": bson.M{
				"regist_area": "$regist_area",
				"total":       bson.M{"$add": []interface{}{"$diamond", "$shadow_diamond", "$coin"}},
			},
		},
		{
			"$sort": bson.M{"total": -1},
		},
		{
			"$limit": 20,
		},
	}
	var xiedai_res []bson.M
	PlayerUsers.Pipe(pipeline).All(&xiedai_res)
	var s_ids []string
	for i, xd := range xiedai_res {
		uid := xd["_id"].(string)
		amount := xd["total"].(int64)
		regist_area := 0
		dsId, _ := xd["regist_area"]
		if dsId == nil {
			regist_area = 0
		} else {
			regist_area = dsId.(int)
		}
		info := new(entity.RankingList)
		info.Id = (i + 1)
		info.Userid = uid
		info.CarryAmount = int64(amount)
		info.FCarryAmount = Chip2Float(int64(amount))
		info.RegistArea = regist_area
		s_ids = append(s_ids, uid)
		map_list[uid] = *info
	}
	// 获取充值
	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"order_status": 4,
				"userid":       bson.M{"$in": s_ids},
				"ctime":        bson.M{"$gte": startTime, "$lte": endTime},
			},
		},
		{
			"$group": bson.M{
				"_id":    "$userid",
				"amount": bson.M{"$sum": "$amount"},
			},
		},
	}
	var pay_res []bson.M
	Pays.Pipe(pipeline1).All(&pay_res)
	for _, with := range pay_res {
		wid := with["_id"].(string)
		stat, ok := map_list[wid]
		if ok {
			amount := with["amount"].(int)
			stat.PayAmount = int64(amount)
			stat.FPayAmount = Chip2Float(int64(amount))
			map_list[wid] = stat
		}
	}

	// 获取提现
	pipeline2 := []bson.M{
		{
			"$match": bson.M{
				"order_status": 2,
				"userid":       bson.M{"$in": s_ids},
				"ctime":        bson.M{"$gte": startTime, "$lte": endTime},
			},
		},
		{
			"$group": bson.M{
				"_id":    "$userid",
				"amount": bson.M{"$sum": "$amount"},
			},
		},
	}
	var with_res []bson.M
	Withdraws.Pipe(pipeline2).All(&with_res)
	for _, with := range with_res {
		wid := with["_id"].(string)
		stat, ok := map_list[wid]
		if ok {
			amount := with["amount"].(int)
			stat.WithdrawAmount = int64(amount)
			stat.FWithdrawAmount = Chip2Float(int64(amount))
			map_list[wid] = stat
		}
	}
	for _, v := range map_list {
		profit := (v.WithdrawAmount + v.CarryAmount) - v.PayAmount
		v.ProfitAmount = profit
		v.FProfitAmount = Chip2Float(profit)
		list = append(list, v)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Id < list[j].Id
	})
	return list
}

// 排行榜-峰值携带榜
func (s *statistics2Service) GetFZCarryRanking(startTime, endTime time.Time) []entity.RankingList {
	list := make([]entity.RankingList, 0)
	map_list := make(map[string]entity.RankingList, 0)
	// 当天存在登录的玩家
	n := bson.M{}
	n["login_time"] = bson.M{"$gte": startTime, "$lte": endTime}
	var log_ids []string
	LoginLogs.Find(n).Distinct("userid", &log_ids)
	// 获取携带
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"robot":            false,
				"simulation_robot": false,
				"_id":              bson.M{"$in": log_ids},
			},
		},
		{
			"$group": bson.M{
				"_id":            "$_id",
				"regist_area":    bson.M{"$first": "$regist_area"},
				"diamond":        bson.M{"$sum": "$diamond"},
				"shadow_diamond": bson.M{"$sum": "$shadow_diamond"},
				"coin":           bson.M{"$sum": "$coin"},
			},
		},
		{
			"$project": bson.M{
				"regist_area": "$regist_area",
				"total":       bson.M{"$add": []interface{}{"$diamond", "$shadow_diamond", "$coin"}},
			},
		},
		{
			"$sort": bson.M{"total": -1},
		},
		{
			"$limit": 20,
		},
	}
	var xiedai_res []bson.M
	PlayerUsers.Pipe(pipeline).All(&xiedai_res)
	var s_ids []string
	for i, xd := range xiedai_res {
		uid := xd["_id"].(string)
		amount := xd["total"].(int64)
		regist_area := 0
		dsId, _ := xd["regist_area"]
		if dsId == nil {
			regist_area = 0
		} else {
			regist_area = dsId.(int)
		}
		info := new(entity.RankingList)
		info.Id = (i + 1)
		info.Userid = uid
		info.CarryAmount = int64(amount)
		info.FCarryAmount = Chip2Float(int64(amount))
		info.RegistArea = regist_area
		s_ids = append(s_ids, uid)
		map_list[uid] = *info
	}
	// 获取充值
	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"order_status": 4,
				"userid":       bson.M{"$in": s_ids},
				"ctime":        bson.M{"$gte": startTime, "$lte": endTime},
			},
		},
		{
			"$group": bson.M{
				"_id":    "$userid",
				"amount": bson.M{"$sum": "$amount"},
			},
		},
	}
	var pay_res []bson.M
	Pays.Pipe(pipeline1).All(&pay_res)
	for _, with := range pay_res {
		wid := with["_id"].(string)
		stat, ok := map_list[wid]
		if ok {
			amount := with["amount"].(int)
			stat.PayAmount = int64(amount)
			stat.FPayAmount = Chip2Float(int64(amount))
			map_list[wid] = stat
		}
	}

	// 获取提现
	pipeline2 := []bson.M{
		{
			"$match": bson.M{
				"order_status": 2,
				"userid":       bson.M{"$in": s_ids},
				"ctime":        bson.M{"$gte": startTime, "$lte": endTime},
			},
		},
		{
			"$group": bson.M{
				"_id":    "$userid",
				"amount": bson.M{"$sum": "$amount"},
			},
		},
	}
	var with_res []bson.M
	Withdraws.Pipe(pipeline2).All(&with_res)
	for _, with := range with_res {
		wid := with["_id"].(string)
		stat, ok := map_list[wid]
		if ok {
			amount := with["amount"].(int)
			stat.WithdrawAmount = int64(amount)
			stat.FWithdrawAmount = Chip2Float(int64(amount))
			map_list[wid] = stat
		}
	}
	for _, v := range map_list {
		profit := (v.WithdrawAmount + v.CarryAmount) - v.PayAmount
		v.ProfitAmount = profit
		v.FProfitAmount = Chip2Float(profit)
		list = append(list, v)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Id < list[j].Id
	})
	return list
}

// 玩家数据
func (s *statistics2Service) GetPlayerData(page, pageSize int, userid string) (*entity.PlayerData, error) {
	list := new(entity.PlayerData)
	m := bson.M{
		"robot":            false,
		"simulation_robot": false,
		"_id":              userid,
	}
	var Info entity.PlayerInfo
	var Player_list []entity.PlayerInfo
	PlayerUsers.Find(m).One(&Info)
	if Info.Userid != "" {
		// 获取同卡号
		w_m := []bson.M{
			{"$match": bson.M{
				"userid":       Info.Userid,
				"order_status": 2,
			}},
			{"$group": bson.M{
				"_id": "$blank_number",
			}},
		}
		var res []bson.M
		Withdraws.Pipe(w_m).All(&res)
		var card_arr []string
		for _, v := range res {
			wid := v["_id"].(string)
			card_arr = append(card_arr, wid)
		}
		w_m1 := bson.M{
			"userid":       bson.M{"$ne": Info.Userid},
			"order_status": 2,
			"blank_number": bson.M{"$in": card_arr},
		}
		var res1 []string
		Withdraws.Find(w_m1).Distinct("userid", &res1)
		Info.BankCardNumber = len(res1)
		// 同设备数
		d_m := bson.M{
			"_id":      bson.M{"$ne": Info.Userid},
			"ad__adid": Info.AD_ADID,
		}
		var res2 []string
		Withdraws.Find(d_m).Distinct("_id", &res2)
		Info.DeviceNumber = len(res2)

		t := (Info.Diamond + Info.ShadowDiamond)
		Info.Assets = Chip2Float(int64(t))
		Info.ShowAmount = Chip2Float(int64(Info.Coin + Info.Diamond))
		Info.FDiamond = Chip2Float(int64(Info.Diamond))
		Info.FCoin = Chip2Float(int64(Info.Coin))
		Info.FShadowDiamond = Chip2Float(int64(Info.ShadowDiamond))
		Info.FMoney = Chip2Float(int64(Info.Money))
		Info.FGiveDiamond = Chip2Float(int64(Info.GiveDiamond))
		Info.FOutDiamond = Chip2Float(int64(Info.OutDiamond))
		Info.FCashOut = Chip2Float(int64(Info.CashOut))
		reg, _ := ConvertToIndiaTime(Info.Ctime.Unix())
		Info.Ctime = reg
		if Info.CustomTypes != "" {
			Info.UserType = Info.CustomTypes
		} else {
			_, Info.UserType = GetChargeType(int64(Info.Money), Info.State)
			// switch Info.State {
			// case 1:
			// 	Info.UserType = "新手"
			// case 2:
			// 	// 用户类型判断
			// 	if Info.Money >= 0 && Info.Money <= 9999 {
			// 		Info.UserType = "零充"
			// 	} else if Info.Money >= 10000 && Info.Money <= 99999 {
			// 		Info.UserType = "普R"
			// 	} else if Info.Money >= 100000 && Info.Money <= 499999 {
			// 		Info.UserType = "小R"
			// 	} else if Info.Money >= 500000 && Info.Money <= 999999 {
			// 		Info.UserType = "中R"
			// 	} else if Info.Money >= 1000000 && Info.Money <= 9999999 {
			// 		Info.UserType = "大R"
			// 	} else if Info.Money >= 10000000 {
			// 		Info.UserType = "超大R"
			// 	}
			// case 3:
			// 	Info.UserType = "平民"
			// case 4:
			// 	Info.UserType = "泡沫"
			// }
		}
		Player_list = append(Player_list, Info)
	}
	list.Players = Player_list
	// 生命周期数据 注册、充值、提现
	smzq_m := bson.M{}
	smzq_m["ltype"] = 1
	smzq_m["userid"] = userid
	var log_list []entity.LogWater
	LogWaters.Find(smzq_m).All(&log_list)
	// 充值、提现
	pay_m := bson.M{}
	pay_m["order_status"] = 4
	pay_m["userid"] = userid
	var paylist []entity.PayOrder
	Pays.Find(pay_m).All(&paylist)
	// 提现
	with_m := bson.M{}
	with_m["order_status"] = 2
	with_m["userid"] = userid
	var withlist []entity.WithdrawOrder
	Withdraws.Find(with_m).All(&withlist)
	new_list := make([]entity.LifecycleInfo, 0)
	if len(log_list) > 0 {
		for _, v := range log_list {
			new_info := new(entity.LifecycleInfo)
			new_info.AddDiamond = v.AddDiamond
			new_info.FAddDiamond = Chip2Float(v.AddDiamond)
			new_info.NowDiamond = v.NowDiamond
			new_info.FNowDiamond = Chip2Float(v.NowDiamond)
			new_info.Stime = v.Ctime.Unix()
			new_info.Ytime = v.Ctime
			c, _ := ConvertToIndiaTime(new_info.Stime)
			new_info.Ctime = c
			new_info.Id = v.Id
			new_info.LType = v.LType
			if v.LType == 1 {
				new_info.EventType = 1
				new_info.LTypeName = "注册"
				new_info.ProductName = v.WaterDesc
			}
			new_list = append(new_list, *new_info)
		}
	}
	if len(paylist) > 0 {
		for _, v := range paylist {
			new_info := new(entity.LifecycleInfo)
			new_info.AddDiamond = int64(v.Score) + int64(v.OtherPresent)
			new_info.FAddDiamond = Chip2Float(int64(v.Score))
			new_info.NowDiamond = 0
			new_info.FNowDiamond = 0
			new_info.Ytime = v.Ctime
			new_info.Stime = v.Ctime.Unix()
			c, _ := ConvertToIndiaTime(new_info.Stime)
			new_info.Ctime = c
			new_info.Id = v.OrderID
			new_info.LType = 2
			new_info.EventType = 2
			new_info.LTypeName = "充值"
			new_info.ProductName = v.ShopName
			new_list = append(new_list, *new_info)
		}
	}
	if len(withlist) > 0 {
		for _, v := range withlist {
			new_info := new(entity.LifecycleInfo)
			new_info.AddDiamond = int64(v.Score)
			new_info.FAddDiamond = Chip2Float(int64(v.Score))
			new_info.NowDiamond = 0
			new_info.FNowDiamond = 0
			new_info.Ytime = v.Ctime
			new_info.Stime = v.Ctime.Unix()
			c, _ := ConvertToIndiaTime(new_info.Stime)
			new_info.Ctime = c
			new_info.Id = v.OrderID
			new_info.LType = 3
			new_info.EventType = 3
			new_info.LTypeName = "提现"
			new_info.ProductName = v.ShopName
			new_list = append(new_list, *new_info)
		}
	}
	sort.Slice(new_list, func(i, j int) bool {
		return new_list[i].Stime > new_list[j].Stime
	})

	count := len(new_list)
	startIndex := (page - 1) * pageSize
	endIndex := (page) * pageSize
	if endIndex > count {
		endIndex = count
	}
	pagelist := new_list[startIndex:endIndex]
	// 获取携带金币
	for k, v := range pagelist {
		if v.LType != 1 {
			// 根据时间获取携带金币
			xd_m := bson.M{}
			xd_m["userid"] = userid
			xd_m["ctime"] = bson.M{"$gte": v.Ytime}
			var log_info entity.LogWater
			LogWaters.Find(xd_m).Sort("ctime").One(&log_info)
			if log_info.Id != "" {
				v.NowDiamond = log_info.NowDiamond
				v.FNowDiamond = Chip2Float(log_info.NowDiamond)
				pagelist[k] = v
			}
		}
	}
	list.LifecycleCount = count
	list.Lifecycle = pagelist

	// skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	// smzq_m := bson.M{}
	// smzq_m["ltype"] = bson.M{"$in": []int{13, 63, 4, 74, 84, 86, 119, 80, 57, 58, 59, 1}}
	// smzq_m["userid"] = userid
	// smzq_m["water_desc"] = bson.M{"$not": bson.RegEx{Pattern: "每日", Options: ""}}
	// var log_list []entity.LogWater
	// LogWaters.Find(smzq_m).Sort(sortFieldR).Skip(skipNum).Limit(pageSize).All(&log_list)
	// list.LifecycleCount = Count(LogWaters, smzq_m)
	// if len(log_list) > 0 {
	// 	new_list := make([]entity.LifecycleInfo, 0)
	// 	for _, v := range log_list {
	// 		new_info := new(entity.LifecycleInfo)
	// 		new_info.AddDiamond = v.AddDiamond
	// 		new_info.FAddDiamond = Chip2Float(v.AddDiamond)
	// 		new_info.NowDiamond = v.NowDiamond
	// 		new_info.FNowDiamond = Chip2Float(v.NowDiamond)
	// 		new_info.Stime = v.Ctime.Unix()
	// 		c, _ := ConvertToIndiaTime(new_info.Stime)
	// 		new_info.Ctime = c
	// 		new_info.Id = v.Id
	// 		new_info.LType = v.LType
	// 		if v.LType == 13 || v.LType == 63 {
	// 			new_info.EventType = 3
	// 			new_info.LTypeName = v.WaterDesc
	// 		}
	// 		if v.LType == 4 || v.LType == 74 || v.LType == 84 || v.LType == 86 || v.LType == 119 || v.LType == 80 || v.LType == 57 || v.LType == 58 || v.LType == 59 {
	// 			new_info.LTypeName = "充值"
	// 			new_info.EventType = 2
	// 			new_info.ProductName = v.WaterDesc
	// 		}
	// 		if v.LType == 1 {
	// 			new_info.EventType = 1
	// 			new_info.LTypeName = "注册"
	// 		}
	// 		new_list = append(new_list, *new_info)
	// 	}
	// 	sort.Slice(new_list, func(i, j int) bool {
	// 		return new_list[i].Stime > new_list[j].Stime
	// 	})
	// 	list.Lifecycle = new_list
	// }

	var game_datas []map[string]any
	err := ck.Select(&game_datas, `
		select gtype, SUM(bets) bets, SUM(rebate) rebates, SUM(game_times) game_times, SUM(rounds) rounds, SUM(win_rounds) win_rounds
		from (
			select gtype, SUM(bet_amount) bets, SUM(bet_amount + score) rebate, SUM(end_time - begin_time) game_times, count(*) rounds,
				SUM(case when win_type = 1 then 1 else 0 end) win_rounds
			from game.col_detail final where begin_time > ? and userid = ? and robot = 0 and win_type in (1,2,3)
			group by gtype
			
				union all
			
			select t1.game_id gtype, SUM(t1.bets) bets, SUM(t1.rebate) rebate, 0 game_times, count(*) rounds, 
				SUM(case when t1.rebate > t1.bets then 1 else 0 end) win_rounds
			from (
				select s2.round_id, s2.game_id, SUM(s2.amounts) bets, SUM(s3.amounts) rebate 
				FROM (SELECT round_id, user_id, game_id, SUM(amount) amounts FROM game.col_nsq_log_external_bet FINAL WHERE amount != 0 and ctime > ? and user_id = ? GROUP BY round_id, user_id, game_id) s2
				LEFT JOIN (SELECT round_id, user_id, game_id, SUM(amount) amounts FROM game.col_nsq_log_external_reward FINAL WHERE amount != 0 and ctime > ? and user_id = ? GROUP BY round_id, user_id, game_id) s3
					ON s2.round_id = s3.round_id AND s2.user_id = s3.user_id and s2.game_id = s3.game_id
				group by s2.round_id, s2.game_id
			) t1
			group by t1.game_id
		) t0
		group by gtype
		order by bets desc
	`, Info.Ctime.Unix(), userid, Info.Ctime.Unix(), userid, Info.Ctime.Unix(), userid)
	if err != nil {
		return list, err
	}
	var game_list []entity.UserGameData
	for i, data := range game_datas {
		gtype := utils.ToInt64(data["gtype"])
		bets := utils.ToInt64(data["bets"])
		rebates := utils.ToInt64(data["rebates"])
		game_times := utils.ToInt64(data["game_times"])
		rounds := utils.ToInt64(data["rounds"])
		win_rounds := utils.ToInt64(data["win_rounds"])
		gameData := entity.UserGameData{
			Id:       int64(i + 1),
			Userid:   userid,
			Gtype:    gtype,
			Number:   rounds,
			GameTime: game_times,
			Bets:     bets,
			// WinBets:    gtype,
			// LoseBets:   gtype,
			WinNumber: win_rounds,
			// LoseNumber: gtype,
			GtypeName: GtypeNameMap[int(gtype)],
			Profit:    Chip2Float(rebates - bets),
			FBets:     Chip2Float(bets),
			FWinBets:  Chip2Float(rebates),
			FLoseBets: 0,
			WinRate:   ComputeFloat(win_rounds, rounds) * 100.0,
		}
		gameData.IsProfit = rebates > bets
		game_list = append(game_list, gameData)
	}
	list.Games = game_list

	// 对局统计
	// game_m := []bson.M{
	// 	{
	// 		"$match": bson.M{
	// 			"userid": userid,
	// 		},
	// 	},
	// 	{
	// 		"$group": bson.M{
	// 			"_id": "$gtype",
	// 			"number": bson.M{
	// 				"$sum": "$number",
	// 			},
	// 			"game_time": bson.M{
	// 				"$sum": "$game_time",
	// 			},
	// 			"bets": bson.M{
	// 				"$sum": "$bets",
	// 			},
	// 			"win_bets": bson.M{
	// 				"$sum": "$win_bets",
	// 			},
	// 			"lose_bets": bson.M{
	// 				"$sum": "$lose_bets",
	// 			},
	// 			"win_number": bson.M{
	// 				"$sum": "$win_number",
	// 			},
	// 			"lose_number": bson.M{
	// 				"$sum": "$lose_number",
	// 			},
	// 		},
	// 	},
	// }
	// var game_list []entity.UserGameData
	// UserGameDatas.Pipe(game_m).All(&game_list)
	// if len(game_list) > 0 {
	// 	// 游戏列表
	// 	recordList := map[int]string{0: "全部"}
	// 	utils.CopyMap(recordList, GtypeNameMap)
	// 	for k, v := range game_list {
	// 		v.Userid = userid
	// 		money := v.WinBets + v.LoseBets
	// 		v.Profit = Chip2Float(money)
	// 		if money >= 0 {
	// 			v.IsProfit = true
	// 		}

	// 		v.FBets = Chip2Float(v.Bets)
	// 		v.FWinBets = Chip2Float(v.WinBets)
	// 		v.FLoseBets = Chip2Float(v.LoseBets)
	// 		v.WinRate = ComputeFloat(v.WinNumber, v.Number) * 100.0
	// 		v.Gtype = v.Id
	// 		if name, ok := recordList[int(v.Gtype)]; ok {
	// 			v.GtypeName = name
	// 		}

	// 		// switch v.Id {
	// 		// case 1:
	// 		// 	v.GtypeName = "TP"
	// 		// case 2:
	// 		// 	v.GtypeName = "DRAGON TIGER"
	// 		// case 3:
	// 		// 	v.GtypeName = "7UPDOWN"
	// 		// case 4:
	// 		// 	v.GtypeName = "RUMMY"
	// 		// case 5:
	// 		// 	v.GtypeName = "AK47"
	// 		// case 6:
	// 		// 	v.GtypeName = "JOKER"
	// 		// case 7:
	// 		// 	v.GtypeName = "CRASH"
	// 		// case 8:
	// 		// 	v.GtypeName = "ANDARBAHAR"
	// 		// case 9:
	// 		// 	v.GtypeName = "彩票"
	// 		// case 10:
	// 		// 	v.GtypeName = "飞机"

	// 		// }
	// 		game_list[k] = v
	// 	}
	// 	// 使用 sort.Slice() 进行排序
	// 	sort.Slice(game_list, func(i, j int) bool {
	// 		return game_list[i].Gtype < game_list[j].Gtype
	// 	})
	// }
	// list.Games = game_list

	// 控制模型
	skipNum1, sortFieldR1 := parsePageAndSort(0, 50, "ctime", false)
	s_m := bson.M{}
	s_m["userid"] = userid
	var game_stg []entity.GameStrategy
	UserGameStrategys.Find(s_m).Sort(sortFieldR1).Skip(skipNum1).Limit(50).All(&game_stg)
	list.ControlCount = Count(UserGameStrategys, s_m)
	if len(game_stg) > 0 {

		for k, v := range game_stg {
			c, _ := ConvertToIndiaTime(v.Ctime)
			v.SDate = c

			switch v.Gtype {
			case 1:
				v.GtypeName = "TP"
			case 2:
				v.GtypeName = "DRAGON TIGER"
			case 3:
				v.GtypeName = "7UPDOWN"
			case 4:
				v.GtypeName = "RUMMY"
			case 5:
				v.GtypeName = "AK47"
			case 6:
				v.GtypeName = "JOKER"
			case 7:
				v.GtypeName = "CRASH"
			case 8:
				v.GtypeName = "ANDARBAHAR"
			case 9:
				v.GtypeName = "彩票"
			case 10:
				v.GtypeName = "飞机"
			case 11:
				v.GtypeName = "红黑大战"
			case 12:
				v.GtypeName = "RUMMY双人"
			case 13:
				v.GtypeName = "TP2"
			case 14:
				v.GtypeName = "MINES"
			}
			v.FEffectiveIncome = ComputeFloat(v.EffectiveIncome, 100)
			game_stg[k] = v
		}
		list.Controls = game_stg
	}
	return list, nil
}

// 修改剧情plus库存
func (c *statistics2Service) ModifyTpStoryStock(gameId string, stock int64) (err error) {
	info := &pb.ModifyTpStoryStock{GameId: gameId, Stock: stock}
	_, err = GmRequest(pb.WebModifyTpStoryStock, pb.CONFIG_UPSERT, info)
	return
}

// tp剧情局plus库存日志
func (c *statistics2Service) GetTpStoryStockHistoryList(m bson.M, page, pageSize int) (list []*entity.TPStoryChargeStockHistory, count int, err error) {
	count, err = TPStoryChargeStockHistoryLogs.Find(m).Count()
	if err != nil {
		return
	}

	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "timestamp", false)
	err = TPStoryChargeStockHistoryLogs.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	if err != nil {
		return
	}

	for _, log := range list {
		log.FChange = fmt.Sprintf("%.2f", float64(log.Change)/100)
		log.FBefore = fmt.Sprintf("%.2f", float64(log.Before)/100)
		log.FAfter = fmt.Sprintf("%.2f", float64(log.After)/100)
		log.FTimestamp = utils.Time2LocalStr(time.Unix(log.Timestamp, 0))
	}
	return
}

// tp剧情局plus库存
func (c *statistics2Service) GetTpStoryStockList() (stocks []*entity.TPStoryChargeStock, err error) {
	//请求服务器
	result, err := GmRequest(pb.WebTpStoryStockMin, pb.CONFIG_UPSERT, []byte{})
	if err != nil {
		debug.PrintStack()
		beego.Error("err: ", err)
		return
	}
	resp, ok := result.(map[string]int32)
	if !ok {
		beego.Error("WebTpStoryStockMin err: ", result)
		return
	}

	m := bson.M{
		"gtype":     int(pb.HUA2),
		"status":    1,
		"room_type": 0,
	}
	// skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	var tpGames []*entity.Game
	err = Games.
		Find(m).
		Sort("-ctime").
		All(&tpGames)
	if err != nil {
		return
	}

	// 查询剧情局库存
	err = TPStoryChargeStocks.Find(bson.M{}).All(&stocks)
	if err != nil {
		return
	}

	stockMap := utils.Slice2Map(stocks, func(_ int, s *entity.TPStoryChargeStock) string {
		return s.Id
	})
	var fStocks []*entity.TPStoryChargeStock
	for _, game := range tpGames {
		stock, ok := stockMap[game.Id]
		if ok {
			stock.StockMin = int64(resp[game.Id])
			stock.Name = game.Name
			stock.FStock = fmt.Sprintf("%.2f", float64(stock.Stock)/100.0)
			stock.FStockMin = fmt.Sprintf("%.2f", float64(stock.StockMin)/100.0)
		} else {
			stock = &entity.TPStoryChargeStock{
				Id:       game.Id,
				Stock:    0,
				StockMin: int64(resp[game.Id]),
				FStock:   "0.00",
				Name:     game.Name,
			}
			stock.FStockMin = fmt.Sprintf("%.2f", float64(stock.StockMin)/100.0)
		}
		fStocks = append(fStocks, stock)
	}
	stocks = fStocks
	return
}

// 日数据概览
func (c *statistics2Service) DayDataList(stime, etime time.Time, registAreas []int32, packageIds []string) (stats []*entity.DayDataList, err error) {
	where_u_sql := ""
	var where_u_args []any

	if len(registAreas) > 0 {
		where_u_sql += " and s0.regist_area in ?"
		where_u_args = append(where_u_args, registAreas)
	}
	if len(packageIds) > 0 {
		where_u_sql += " and s0.ad__bundle_id IN ?"
		where_u_args = append(where_u_args, packageIds)
	}

	var datePackageStats = make(map[uint32]map[string]*entity.DayDataList)
	var dateStats = make(map[uint32]*entity.DayDataList)

	isPackages := len(packageIds) > 0
	getStat := func(date uint32, packageId string) (stat *entity.DayDataList) {
		if !isPackages {
			stat = dateStats[date]
			if stat == nil {
				stat = &entity.DayDataList{
					Date:                  date,
					PackageId:             "全部",
					ClassId:               "全部",
					ChannelPays:           make(map[string]int64),
					ChannelWithdraws:      make(map[string]int64),
					ChannelWithdrawOrders: make(map[string]int64),
				}
				dateStats[date] = stat
				stats = append(stats, stat)
			}
		} else {
			packageStats, ok := datePackageStats[date]
			if !ok {
				packageStats = make(map[string]*entity.DayDataList)
				datePackageStats[date] = packageStats
			}
			stat, ok = packageStats[packageId]
			if !ok {
				stat = &entity.DayDataList{
					Date:                  date,
					PackageId:             packageId,
					ChannelPays:           make(map[string]int64),
					ChannelWithdraws:      make(map[string]int64),
					ChannelWithdrawOrders: make(map[string]int64),
				}
				packageStats[packageId] = stat
				stats = append(stats, stat)
			}
		}
		return
	}

	// 日活 老用户日活 付费用户日活
	var login_datas []map[string]any
	login_args := append([]any{stime, etime}, where_u_args...)
	err = ck.Select(&login_datas, fmt.Sprintf(`
		select s1.login_date, s0.ad__bundle_id, count(*) login_users,
			SUM(case when s1.login_date = toYYYYMMDD(s0.ctime) then 0 else 1 end) old_login_users,
			SUM(case when s0.money > 0 then 1 else 0 end) payed_login_users,
			(login_users - payed_login_users) unpay_login_users
		from game.col_user s0 final
		join (
			select toYYYYMMDD(login_time) login_date, userid
			from game.col_log_login final
			where login_time between ? and ?
			group by login_date, userid
		) s1 on s1.userid = s0.userid
		where 1 = 1 %s
		group by s1.login_date, s0.ad__bundle_id
		order by s1.login_date desc, s0.ad__bundle_id
	`, where_u_sql), login_args...)
	if err != nil {
		return
	}
	for _, data := range login_datas {
		login_date := data["login_date"].(uint32)
		ad__bundle_id := data["ad__bundle_id"].(string)
		login_users := utils.ToInt64(data["login_users"])
		old_login_users := utils.ToInt64(data["old_login_users"])
		payed_login_users := utils.ToInt64(data["payed_login_users"])
		unpay_login_users := utils.ToInt64(data["unpay_login_users"])

		stat := getStat(login_date, ad__bundle_id)
		stat.LoginUsers += login_users
		stat.OldLoginUsers0 += old_login_users
		stat.PayedLoginUsers += payed_login_users
		stat.UnpayLoginUsers += unpay_login_users
	}

	// 注册用户
	var reg_datas []map[string]any
	reg_args := append([]any{stime, etime}, where_u_args...)
	err = ck.Select(&reg_datas, fmt.Sprintf(`
		select t1.reg_date, t1.ad__bundle_id ad__bundle_id,
			count(distinct t1.userid) reg_users,
			count(distinct t1.ad__device_id) reg_ads,
			count(distinct (case when t0.userid != t1.userid and toYYYYMMDD(t0.ctime) < t1.reg_date then t1.ad__device_id else null end)) repeat_ads,
			(reg_ads - repeat_ads) ad_reg_users
		from game.col_user t0 final
		right join (
			select userid, toYYYYMMDD(s0.ctime) reg_date, s0.ad__bundle_id, s0.ad__device_id
			from game.col_user s0 final
			where s0.ctime between ? and ? and robot = 0 and simulation_robot = 0 %s
		) t1 on t0.ad__device_id = t1.ad__device_id
		group by t1.reg_date, t1.ad__bundle_id
	`, where_u_sql), reg_args...)
	if err != nil {
		return
	}
	for _, data := range reg_datas {
		reg_date := data["reg_date"].(uint32)
		ad__bundle_id := data["ad__bundle_id"].(string)
		reg_users := utils.ToInt64(data["reg_users"])
		ad_reg_users := utils.ToInt64(data["ad_reg_users"])

		stat := getStat(reg_date, ad__bundle_id)
		stat.RegUsers0 += reg_users
		stat.AdRegUsers += ad_reg_users
	}

	// 通道支付用户统计
	var pay_user_datas []map[string]any
	pay_user_args := append([]any{stime, etime, stime, etime}, where_u_args...)
	err = ck.Select(&pay_user_datas, fmt.Sprintf(`
		select pay_date, package_id, 
			count(distinct t1.userid) pay_users,
			count(distinct (case when new_user=1 then t1.userid else null end)) new_pay_users,
			(pay_users-new_pay_users) old_pay_users,
			SUM(orders) pay_orders,
			SUM(case when new_user=1 then orders else 0 end) new_pay_orders,
			(pay_orders - new_pay_orders) old_pay_orders,
			SUM(case when orders >= 2 then 1 else 0 end) pay2times_users,
			SUM(case when orders >= 2 and new_user=1 then 1 else 0 end) new_pay2times_users,
			(pay2times_users - new_pay2times_users) old_pay2times_users,
			count(distinct t0.withdraw_userid) withdraw_users,
			count(distinct (case when new_user=1 then t0.withdraw_userid else null end)) new_withdraw_users,
			(withdraw_users - new_withdraw_users) old_withdraw_users
		from (
			select toYYYYMMDD(ctime) withdraw_date, userid withdraw_userid
			from game.col_withdraw_record final where ctime between ? AND ? and order_status = 2
			group by withdraw_date, userid
		) t0 right join (
			select toYYYYMMDD(s1.ctime) pay_date, s1.package_id, s1.userid userid, toYYYYMMDD(s0.ctime) reg_date, SUM(s1.amount) amount_sum,
				(case when pay_date = reg_date then 1 else 0 end) new_user, count(*) orders
			from game.col_user s0 final
			join game.col_trade_record s1 final on s0.userid = s1.userid
			where s1.ctime between ? AND ? and s1.order_status = 4 %s
			group by pay_date, s1.package_id, s1.userid, reg_date
		) t1 on t0.withdraw_date = t1.pay_date and t0.withdraw_userid = t1.userid
		group by pay_date, package_id
	`, where_u_sql), pay_user_args...)
	if err != nil {
		return
	}
	for _, data := range pay_user_datas {
		pay_date := data["pay_date"].(uint32)
		package_id := data["package_id"].(string)
		pay_users := utils.ToInt64(data["pay_users"])
		new_pay_users := utils.ToInt64(data["new_pay_users"])
		old_pay_users := utils.ToInt64(data["old_pay_users"])
		pay_orders := utils.ToInt64(data["pay_orders"])
		new_pay_orders := utils.ToInt64(data["new_pay_orders"])
		old_pay_orders := utils.ToInt64(data["old_pay_orders"])
		pay2times_users := utils.ToInt64(data["pay2times_users"])
		new_pay2times_users := utils.ToInt64(data["new_pay2times_users"])
		old_pay2times_users := utils.ToInt64(data["old_pay2times_users"])
		pay_withdraw_users := utils.ToInt64(data["withdraw_users"])
		new_pay_withdraw_users := utils.ToInt64(data["new_withdraw_users"])
		old_pay_withdraw_users := utils.ToInt64(data["old_withdraw_users"])

		stat := getStat(pay_date, package_id)
		stat.PayUsers += pay_users
		stat.NewPayUsers0 += new_pay_users
		stat.OldPayUsers0 += old_pay_users

		stat.PayOrders += pay_orders
		stat.NewPayOrders += new_pay_orders
		stat.OldPayOrders += old_pay_orders

		stat.Pay2TimesUsers0 += pay2times_users
		stat.NewPay2TimesUsers0 += new_pay2times_users
		stat.OldPay2TimesUsers0 += old_pay2times_users

		stat.PayWithdrawUsers += pay_withdraw_users
		stat.NewPayWithdrawUsers += new_pay_withdraw_users
		stat.OldPayWithdrawUsers += old_pay_withdraw_users
	}

	// 通道支付
	var pay_datas []map[string]any
	pay_args := append([]any{stime, etime}, where_u_args...)
	err = ck.Select(&pay_datas, fmt.Sprintf(`
		select pay_date, package_id, channel_id, SUM(amount_sum) pay_amounts,
			SUM(case when new_user=1 then amount_sum else 0 end) new_pay_amounts,
			(pay_amounts - new_pay_amounts) old_pay_amounts
		from (
			select toYYYYMMDD(s1.ctime) pay_date, s1.package_id, s1.channel_id, s1.userid userid, toYYYYMMDD(s0.ctime) reg_date, SUM(s1.amount) amount_sum,
				(case when pay_date = reg_date then 1 else 0 end) new_user
			from game.col_user s0 final
			join game.col_trade_record s1 final on s0.userid = s1.userid
			where s1.ctime between ? AND ? and s1.order_status = 4 %s
			group by pay_date, s1.package_id, s1.channel_id, s1.userid, reg_date
		) t1
		group by pay_date, package_id, channel_id
	`, where_u_sql), pay_args...)
	if err != nil {
		return
	}
	for _, data := range pay_datas {
		pay_date := data["pay_date"].(uint32)
		package_id := data["package_id"].(string)
		pay_channel_id := fmt.Sprint(data["channel_id"])
		pay_amounts := utils.ToInt64(data["pay_amounts"])
		new_pay_amounts := utils.ToInt64(data["new_pay_amounts"])
		old_pay_amounts := utils.ToInt64(data["old_pay_amounts"])

		stat := getStat(pay_date, package_id)
		stat.PayAmounts0 += pay_amounts
		stat.ChannelPays[pay_channel_id] += pay_amounts

		stat.NewPayAmounts0 += new_pay_amounts
		stat.OldPayAmounts0 += old_pay_amounts
	}

	// 通道提现
	var withdraw_datas []map[string]any
	withdraw_args := append([]any{stime, etime}, where_u_args...)
	err = ck.Select(&withdraw_datas, fmt.Sprintf(`
		select withdraw_date, package_id, out_channel, SUM(amount_sum) withdraw_amounts,
			SUM(amount_sum) withdraw_amounts_untax, 
			count(*) withdraw_users, 
			SUM(new_user) new_withdraw_users, 
			(withdraw_users-new_withdraw_users) old_withdraw_users,
			SUM(case when new_user=1 then amount_sum else 0 end) new_withdraw_amounts,
			(withdraw_amounts - new_withdraw_amounts) old_withdraw_amounts,
			SUM(orders) orders
		from (
			select toYYYYMMDD(s1.ctime) withdraw_date, s1.package_id, s1.out_channel, s1.userid userid, toYYYYMMDD(s0.ctime) reg_date,
				SUM(s1.amount) amount_sum, SUM(s1.commission) commission_sum,
				(case when withdraw_date = reg_date then 1 else 0 end) new_user,
				count(*) orders
			from game.col_user s0 final
			join game.col_withdraw_record s1 final on s0.userid = s1.userid
			where s1.ctime between ? AND ? and s1.order_status = 2 %s
			group by withdraw_date, s1.package_id, s1.out_channel, s1.userid, reg_date
		) t0
		group by withdraw_date, package_id, out_channel
	`, where_u_sql), withdraw_args...)
	if err != nil {
		return
	}
	for _, data := range withdraw_datas {
		withdraw_date := data["withdraw_date"].(uint32)
		package_id := data["package_id"].(string)
		pay_channel_id := fmt.Sprint(data["out_channel"])
		withdraw_amounts := utils.ToInt64(data["withdraw_amounts"])
		withdraw_amounts_untax := utils.ToInt64(data["withdraw_amounts_untax"])
		withdraw_users := utils.ToInt64(data["withdraw_users"])
		new_withdraw_users := utils.ToInt64(data["new_withdraw_users"])
		old_withdraw_users := utils.ToInt64(data["old_withdraw_users"])
		new_withdraw_amounts := utils.ToInt64(data["new_withdraw_amounts"])
		old_withdraw_amounts := utils.ToInt64(data["old_withdraw_amounts"])
		orders := utils.ToInt64(data["orders"])

		stat := getStat(withdraw_date, package_id)
		stat.WithdrawAmounts0 += withdraw_amounts
		stat.WithdrawAmountsUntax += withdraw_amounts_untax
		stat.WithdrawUsers += withdraw_users
		stat.NewWithdrawUsers += new_withdraw_users
		stat.OldWithdrawUsers += old_withdraw_users
		stat.ChannelWithdraws[pay_channel_id] += withdraw_amounts_untax
		stat.ChannelWithdrawOrders[pay_channel_id] += orders

		stat.NewWithdrawAmounts0 += new_withdraw_amounts
		stat.OldWithdrawAmounts0 += old_withdraw_amounts
	}

	_, payChannels, err := PayService.PayChannelList(0)
	if err != nil {
		return
	}
	var payChannelMap = make(map[string]*entity.PayChannel)
	for _, c := range payChannels {
		payChannelMap[c.Id] = c
	}

	// 首充统计
	var first_pay_datas []map[string]any
	first_pay_args := append([]any{stime, etime}, where_u_args...)
	first_pay_args = append(first_pay_args, stime, etime)
	err = ck.Select(&first_pay_datas, fmt.Sprintf(`
		select t1.first_pay_date, t1.package_id package_id,
			count(distinct t1.userid) first_pay_users,
			count(distinct t0.userid) first_pay_withdraw_users,
			count(distinct (case when t1.first_pay_date = toYYYYMMDD(t0.ctime) then t0.userid else null end)) first_pay_day_withdraw_users,
			count(distinct (case when t1.first_pay_date_pay_times >= 2 then t1.userid else null end)) first_pay_date_pay2_users,
			count(distinct (case when t1.first_pay_next_date_pay_times >= 1 then t1.userid else null end)) first_pay_next_date_pay_users
		from game.col_withdraw_record t0 final 
		right join (
			select r1.userid userid, r1.first_pay_date, r0.package_id,
				SUM(case when toYYYYMMDD(r0.ctime) = r1.first_pay_date then 1 else 0 end) first_pay_date_pay_times,
				SUM(case when toYYYYMMDD(dateAdd(DAY, -1, r0.ctime)) = r1.first_pay_date then 1 else 0 end) first_pay_next_date_pay_times
			from game.col_trade_record r0 final
			join (
				select userid, min(ctime) first_pay_time, toYYYYMMDD(first_pay_time) first_pay_date
				from game.col_trade_record final where userid in (
					select s1.userid
					from game.col_user s0 final
					join game.col_trade_record s1 final on s1.userid = s0.userid
					where s1.ctime between ? AND ? and s1.order_status = 4 %s
					group by s1.userid
				) and order_status = 4
				group by userid
				having first_pay_time between ? AND ?
			) r1 on r0.userid = r1.userid and r0.order_status = 4
			group by r1.userid, r1.first_pay_date, r0.package_id
		) t1 on t0.userid = t1.userid and t0.order_status = 2
		group by t1.first_pay_date, t1.package_id
	`, where_u_sql), first_pay_args...)
	if err != nil {
		return
	}
	for _, data := range first_pay_datas {
		first_pay_date := data["first_pay_date"].(uint32)
		package_id := data["package_id"].(string)
		first_pay_users := utils.ToInt64(data["first_pay_users"])
		first_pay_withdraw_users := utils.ToInt64(data["first_pay_withdraw_users"])
		first_pay_day_withdraw_users := utils.ToInt64(data["first_pay_day_withdraw_users"])
		first_pay_date_pay2_users := utils.ToInt64(data["first_pay_date_pay2_users"])
		first_pay_next_date_pay_users := utils.ToInt64(data["first_pay_next_date_pay_users"])

		stat := getStat(first_pay_date, package_id)
		stat.FirstPayUsers += first_pay_users
		stat.FirstPayWithdrawUsers += first_pay_withdraw_users
		stat.FirstPayDayWithdrawUsers += first_pay_day_withdraw_users
		stat.FirstPayDatePay2Users0 += first_pay_date_pay2_users
		stat.FirstPayNextDatePayUsers += first_pay_next_date_pay_users
	}

	channelMap, err := ChannelService.GetChannelsMap()
	if err != nil {
		return
	}

	// 汇总计算
	for _, stat := range stats {
		sdate := fmt.Sprint(stat.Date)
		stat.SDate = sdate[0:4] + "-" + sdate[4:6] + "-" + sdate[6:8]

		if stat.PackageId != "全部" {
			if c, ok := channelMap[stat.PackageId]; ok {
				stat.ClassId = c.ClassName
				stat.AliasId = c.Name1
			}
		}

		stat.PaySubWithdraw0 = stat.PayAmounts0 - stat.WithdrawAmounts0
		// 代收手续费
		for channelId, pays := range stat.ChannelPays {
			if c, ok := payChannelMap[channelId]; ok {
				stat.PayTaxs0 += int64(float64(pays) * c.PayRate / 100)
			}
		}
		// 代付手续费
		for channelId, withdraws := range stat.ChannelWithdraws {
			if c, ok := payChannelMap[channelId]; ok {
				stat.WithdrawTaxs0 += int64(float64(withdraws) * c.WithdrawRate / 100)
			}
		}
		for channelId, orders := range stat.ChannelWithdrawOrders {
			if c, ok := payChannelMap[channelId]; ok {
				stat.WithdrawTaxs0 += (orders * c.WithdrawFee)
			}
		}

		stat.PayTaxs = fmt.Sprintf("%.2f", Chip2Float(stat.PayTaxs0))
		stat.WithdrawTaxs = fmt.Sprintf("%.2f", Chip2Float(stat.WithdrawTaxs0))
		stat.PayAmounts = fmt.Sprintf("%.2f", Chip2Float(stat.PayAmounts0))
		stat.WithdrawAmounts = fmt.Sprintf("%.2f", Chip2Float(stat.WithdrawAmounts0))
		stat.PaySubWithdraw = fmt.Sprintf("%.2f", Chip2Float(stat.PaySubWithdraw0))

		// 充-提-代收手续费-代付手续费
		channelSurplus := stat.PayAmounts0 - stat.WithdrawAmountsUntax - stat.PayTaxs0 - stat.WithdrawTaxs0
		stat.ChannelSurplus = fmt.Sprintf("%.2f", Chip2Float(channelSurplus))
		stat.PayWithdrawSurplusRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.PayAmounts0-stat.WithdrawAmountsUntax, stat.PayAmounts0)*100)
		stat.ChannelSurplusRate = fmt.Sprintf("%.2f%%", ComputeFloat(channelSurplus, stat.PayAmounts0)*100)

		stat.RegUsers = fmt.Sprintf("%d\n(%.2f%%)", stat.RegUsers0, ComputeFloat(stat.RegUsers0, stat.LoginUsers)*100)
		stat.EffectRegUserRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.AdRegUsers, stat.RegUsers0)*100)
		stat.PayRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.PayUsers, stat.LoginUsers)*100)
		stat.NewPayRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.NewPayUsers0, stat.RegUsers0)*100)
		stat.OldPayRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.OldPayUsers0, stat.OldLoginUsers0)*100)

		stat.WithdrawRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.WithdrawUsers, stat.LoginUsers)*100)
		stat.NewWithdrawRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.NewWithdrawUsers, stat.RegUsers0)*100)
		stat.OldWithdrawRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.OldWithdrawUsers, stat.OldLoginUsers0)*100)

		stat.NewPaySubWithdraw0 = stat.NewPayAmounts0 - stat.NewWithdrawAmounts0
		stat.OldPaySubWithdraw0 = stat.OldPayAmounts0 - stat.OldWithdrawAmounts0

		stat.PayWithdrawRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.PayWithdrawUsers, stat.PayUsers)*100)
		stat.NewPayWithdrawRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.NewPayWithdrawUsers, stat.NewPayUsers0)*100)
		stat.OldPayWithdrawRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.OldPayWithdrawUsers, stat.OldPayUsers0)*100)

		stat.NewSurplusRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.NewPaySubWithdraw0, stat.NewPayAmounts0)*100)
		stat.OldSurplusRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.OldPaySubWithdraw0, stat.OldPayAmounts0)*100)

		stat.ARPU = fmt.Sprintf("%.2f", ComputeFloat(stat.PayAmounts0, stat.LoginUsers)/100)
		stat.NewARPU = fmt.Sprintf("%.2f", ComputeFloat(stat.NewPayAmounts0, stat.RegUsers0)/100)
		stat.OldARPU = fmt.Sprintf("%.2f", ComputeFloat(stat.OldPayAmounts0, stat.OldLoginUsers0)/100)
		stat.ARPPU = fmt.Sprintf("%.2f", ComputeFloat(stat.PayAmounts0, stat.PayUsers)/100)
		stat.NewARPPU = fmt.Sprintf("%.2f", ComputeFloat(stat.NewPayAmounts0, stat.NewPayUsers0)/100)
		stat.OldARPPU = fmt.Sprintf("%.2f", ComputeFloat(stat.OldPayAmounts0, stat.OldPayUsers0)/100)

		stat.FirstPayRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.FirstPayUsers, stat.UnpayLoginUsers+stat.FirstPayUsers)*100)
		stat.FirstPayRate2 = fmt.Sprintf("%.2f%%", ComputeFloat(stat.FirstPayUsers, stat.RegUsers0)*100)
		stat.FirstPayNextDatePayRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.FirstPayNextDatePayUsers, stat.FirstPayUsers)*100)
		stat.FirstPayWithdrawRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.FirstPayDayWithdrawUsers, stat.FirstPayUsers)*100)

		stat.PayOrderAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.PayOrders, stat.PayUsers))
		stat.NewPayOrderAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.NewPayOrders, stat.NewPayUsers0))
		stat.OldPayOrderAvg = fmt.Sprintf("%.2f", ComputeFloat(stat.OldPayOrders, stat.OldPayUsers0))

		stat.OldLoginUsers = fmt.Sprintf("%d\n(%.2f%%)", stat.OldLoginUsers0, ComputeFloat(stat.OldLoginUsers0, stat.LoginUsers)*100)
		stat.NewPayUsers = fmt.Sprintf("%d\n(%.2f%%)", stat.NewPayUsers0, ComputeFloat(stat.NewPayUsers0, stat.PayUsers)*100)
		stat.OldPayUsers = fmt.Sprintf("%d\n(%.2f%%)", stat.OldPayUsers0, ComputeFloat(stat.OldPayUsers0, stat.PayUsers)*100)
		stat.NewPayAmounts = fmt.Sprintf("%.2f\n(%.2f%%)", Chip2Float(stat.NewPayAmounts0), ComputeFloat(stat.NewPayAmounts0, stat.PayAmounts0)*100)
		stat.OldPayAmounts = fmt.Sprintf("%.2f\n(%.2f%%)", Chip2Float(stat.OldPayAmounts0), ComputeFloat(stat.OldPayAmounts0, stat.PayAmounts0)*100)
		stat.NewWithdrawAmounts = fmt.Sprintf("%.2f\n(%.2f%%)", Chip2Float(stat.NewWithdrawAmounts0), ComputeFloat(stat.NewWithdrawAmounts0, stat.WithdrawAmounts0)*100)
		stat.OldWithdrawAmounts = fmt.Sprintf("%.2f\n(%.2f%%)", Chip2Float(stat.OldWithdrawAmounts0), ComputeFloat(stat.OldWithdrawAmounts0, stat.WithdrawAmounts0)*100)
		stat.NewPaySubWithdraw = fmt.Sprintf("%.2f\n(%.2f%%)", Chip2Float(stat.NewPaySubWithdraw0), ComputeFloat(stat.NewPaySubWithdraw0, stat.PaySubWithdraw0)*100)
		stat.OldPaySubWithdraw = fmt.Sprintf("%.2f\n(%.2f%%)", Chip2Float(stat.OldPaySubWithdraw0), ComputeFloat(stat.OldPaySubWithdraw0, stat.PaySubWithdraw0)*100)

		stat.Pay2TimesUsers = fmt.Sprintf("%d\n(%.2f%%)", stat.Pay2TimesUsers0, ComputeFloat(stat.Pay2TimesUsers0, stat.PayUsers)*100)
		stat.NewPay2TimesUsers = fmt.Sprintf("%d\n(%.2f%%)", stat.NewPay2TimesUsers0, ComputeFloat(stat.NewPay2TimesUsers0, stat.NewPayUsers0)*100)
		stat.OldPay2TimesUsers = fmt.Sprintf("%d\n(%.2f%%)", stat.OldPay2TimesUsers0, ComputeFloat(stat.OldPay2TimesUsers0, stat.OldPayUsers0)*100)
		stat.FirstPayDatePay2Users = fmt.Sprintf("%d\n(%.2f%%)", stat.FirstPayDatePay2Users0, ComputeFloat(stat.FirstPayDatePay2Users0, stat.FirstPayUsers)*100)
	}

	return
}

// 经济日报
func (c *statistics2Service) FinanceStats(stime, etime time.Time, registAreas []int32, packageIds []string) (stats []*entity.FinanceStat, err error) {
	where_u_sql := ""
	var where_u_args []any

	if len(registAreas) > 0 {
		where_u_sql += " and s0.regist_area in ?"
		where_u_args = append(where_u_args, registAreas)
	}
	if len(packageIds) > 0 {
		where_u_sql += " and s0.ad__bundle_id IN ?"
		where_u_args = append(where_u_args, packageIds)
	}

	var dateStatsMap = make(map[uint32]*entity.FinanceStat)

	// 日活
	var login_datas []map[string]any
	login_args := append([]any{stime, etime}, where_u_args...)
	err = ck.Select(&login_datas, fmt.Sprintf(`
		select toYYYYMMDD(s1.login_time) login_date, count(distinct s1.userid) login_users
		from game.col_user s0 final
		join game.col_log_login s1 final on s0.userid = s1.userid
		where s1.login_time between ? and ? %s
		group by login_date
		order by login_date desc
	`, where_u_sql), login_args...)
	if err != nil {
		return
	}
	for _, data := range login_datas {
		login_date := data["login_date"].(uint32)
		login_users := utils.ToInt64(data["login_users"])
		sdate := fmt.Sprint(login_date)
		stat := &entity.FinanceStat{
			SDate:      sdate[0:4] + "-" + sdate[4:6] + "-" + sdate[6:8],
			LoginUsers: login_users,
		}
		stats = append(stats, stat)
		dateStatsMap[login_date] = stat
	}

	// 当前所有玩家钱包总余额
	_ = `
		select t1.day_date,
			SUM(case when t0.ctime < t1.etime then t0.add_diamond else 0 end) cash
		from game.col_log_water t0 final 
		join (
			SELECT date_add(DAY, number, toDateTime('2025-04-20 00:00:00', 'Asia/Kolkata')) stime, date_add(DAY, 1, stime) etime, toYYYYMMDD(stime) day_date FROM system.numbers LIMIT 8
		) t1 on 1 = 1
		--where t0.userid = '7275376'
		group by t1.day_date
		order by t1.day_date desc
	`

	// 玩家钱包余额
	m := bson.M{
		"date": bson.M{
			"$gte": uint32(stime.Year()*10000 + int(stime.Month())*100 + stime.Day()),
			"$lte": uint32(etime.Year()*10000 + int(etime.Month())*100 + etime.Day()),
		},
	}
	if len(registAreas) > 0 {
		m["regist_area"] = bson.M{"$in": registAreas}
	}
	if len(packageIds) > 0 {
		m["bundle_id"] = bson.M{"$in": packageIds}
	}
	var wallet_datas []map[string]any
	err = FinanceWalletStats.Pipe([]bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":                   "$date",
			"cash":                  bson.M{"$sum": "$cash"},
			"withdrawable":          bson.M{"$sum": "$withdrawable"},
			"non_withdrawable":      bson.M{"$sum": "$non_withdrawable"},
			"live_cash":             bson.M{"$sum": "$live_cash"},
			"live_withdrawable":     bson.M{"$sum": "$live_withdrawable"},
			"live_non_withdrawable": bson.M{"$sum": "$live_non_withdrawable"},
		}},
	}).All(&wallet_datas)
	if err != nil {
		return
	}
	for _, data := range wallet_datas {
		sdate := uint32(utils.ToInt64(data["_id"]))
		cash := utils.ToInt64(data["cash"])
		withdrawable := utils.ToInt64(data["withdrawable"])
		non_withdrawable := utils.ToInt64(data["non_withdrawable"])
		live_cash := utils.ToInt64(data["live_cash"])
		live_withdrawable := utils.ToInt64(data["live_withdrawable"])
		live_non_withdrawable := utils.ToInt64(data["live_non_withdrawable"])

		if stat, ok := dateStatsMap[sdate]; ok {
			stat.Cash0 = cash
			stat.Withdrawable0 = withdrawable
			stat.NonWithdrawable0 = non_withdrawable
			stat.LiveCash0 = live_cash
			stat.LiveWithdrawable0 = live_withdrawable
			stat.LiveNonWithdrawable0 = live_non_withdrawable
		}
	}

	// 游戏收入
	var game_stats_data []map[string]any
	var game_stats_args = []any{stime.Unix(), etime.Unix()}
	game_stats_args = append(game_stats_args, where_u_args...)
	game_stats_args = append(game_stats_args, stime.Unix(), etime.Unix())
	game_stats_args = append(game_stats_args, where_u_args...)
	game_stats_args = append(game_stats_args, stime.Unix(), etime.Unix())
	game_stats_args = append(game_stats_args, where_u_args...)
	err = ck.Select(&game_stats_data, fmt.Sprintf(`
		select sdate, SUM(bets) bets, SUM(rebate) rebate from (
			SELECT toYYYYMMDD(toDateTime(s1.begin_time)) sdate, SUM(s1.bet_amount) bets, 
				SUM(s1.bet_amount+s1.score) rebate
			FROM game.col_user s0 final
			join game.col_detail s1 FINAL on s0.userid = s1.userid
			WHERE s1.begin_time between ? and ? and s1.robot = 0 and s1.win_type in (1,2,3) %s
			group by sdate

		UNION ALL
		
			select toYYYYMMDD(toDateTime(s1.ctime)) sdate, SUM(s1.amount) bets , 0 rebate
			FROM game.col_user s0 final
			JOIN game.col_nsq_log_external_bet s1 FINAL on s0.userid = s1.user_id
			WHERE s1.ctime between ? and ? and s1.amount != 0 %s
			group by sdate
			
		UNION ALL
			
			select toYYYYMMDD(toDateTime(s1.ctime)) sdate, 0 bets , SUM(s1.amount) rebate
			FROM game.col_user s0 final
			JOIN game.col_nsq_log_external_reward s1 FINAL on s0.userid = s1.user_id
			WHERE s1.ctime between ? and ? and s1.amount != 0 %s
			group by sdate
		) t0
		group by sdate
	`, where_u_sql, where_u_sql, where_u_sql), game_stats_args...)
	if err != nil {
		return
	}
	for _, data := range game_stats_data {
		sdate := data["sdate"].(uint32)
		bets := utils.ToInt64(data["bets"])
		rebate := utils.ToInt64(data["rebate"])
		if stat, ok := dateStatsMap[sdate]; ok {
			// stat.GameIncome0 = int64(float64(bets) * (1 - ComputeFloat(rebate, bets)))
			stat.GameIncome0 = bets - rebate
		}
	}

	// bonus发放总额
	var bonus_total_data []map[string]any
	var bonus_total_args = []any{stime.Unix(), etime.Unix()}
	bonus_total_args = append(bonus_total_args, where_u_args...)
	err = ck.Select(&bonus_total_data, fmt.Sprintf(`
		select toYYYYMMDD(toDateTime(s1.ctime)) sdate, SUM(s1.vb) vbs
		from game.col_user s0 final
		join game.col_log_vbdiamond s1 final on s0.userid = s1.userid
		where s1.ctime between ? and ? and s1.vb > 0 %s
		GROUP BY sdate
	`, where_u_sql), bonus_total_args...)
	if err != nil {
		return
	}
	for _, data := range bonus_total_data {
		sdate := data["sdate"].(uint32)
		vbs := utils.ToInt64(data["vbs"])
		if stat, ok := dateStatsMap[sdate]; ok {
			stat.BonusGift0 = vbs
		}
	}

	// cash发放总额
	var nonGiftLTypes []int32
	nonGiftLTypes = append(nonGiftLTypes, LTypeReason1...)
	nonGiftLTypes = append(nonGiftLTypes, LTypeReason2...)
	nonGiftLTypes = append(nonGiftLTypes, LTypeReason3...)
	nonGiftLTypes = append(nonGiftLTypes, LTypeReason4...)
	nonGiftLTypes = append(nonGiftLTypes, LTypeReason6...)
	nonGiftLTypes = append(nonGiftLTypes, LTypeReason7...)
	nonGiftLTypes = append(nonGiftLTypes, LTypeReason8...)
	var cash_gift_data []map[string]any
	var cash_gift_args = []any{stime, etime, nonGiftLTypes}
	cash_gift_args = append(cash_gift_args, where_u_args...)
	err = ck.Select(&cash_gift_data, fmt.Sprintf(`
		select toYYYYMMDD(s1.ctime) sdate, SUM(s1.add_diamond) gift_cashs,
			SUM(case when ltype = 132 then s1.add_diamond else 0 end) vb_cashs
		from game.col_user s0 final
		join game.col_log_water s1 final on s0.userid = s1.userid
		where s1.ctime between ? and ? and s1.ltype not in ? and s1.add_diamond > 0 %s
		GROUP BY sdate
	`, where_u_sql), cash_gift_args...)
	if err != nil {
		return
	}
	for _, data := range cash_gift_data {
		sdate := data["sdate"].(uint32)
		gift_cashs := utils.ToInt64(data["gift_cashs"])
		vb_cashs := utils.ToInt64(data["vb_cashs"])
		if stat, ok := dateStatsMap[sdate]; ok {
			stat.CashFlow0 = gift_cashs
			stat.CashGift0 = gift_cashs - vb_cashs
			stat.CashVb0 = vb_cashs
		}
	}

	// 充值提现
	var pay_withdraw_data []map[string]any
	var pay_withdraw_args = []any{stime, etime}
	pay_withdraw_args = append(pay_withdraw_args, where_u_args...)
	pay_withdraw_args = append(pay_withdraw_args, stime, etime)
	pay_withdraw_args = append(pay_withdraw_args, where_u_args...)
	err = ck.Select(&pay_withdraw_data, fmt.Sprintf(`
		select sdate, SUM(pays) pays, SUM(withdraws) withdraws, SUM(cash_pay_give) cash_pay_give
		from (
			select toYYYYMMDD(s1.ctime) sdate, SUM(amount) pays, 0 withdraws,
				SUM(case when s1.shop_type == 11 then score - amount else 0 end) cash_pay_give
			from game.col_user s0 final
			join game.col_trade_record s1 final on s1.userid = s0.userid
			where s1.ctime between ? and ? and s1.order_status = 4 %s
			group by sdate
			
			union all
				
			select toYYYYMMDD(s1.ctime) sdate, 0 pays, SUM(amount) withdraws, 0 cash_pay_give
			from game.col_user s0 final
			join game.col_withdraw_record s1 final on s1.userid = s0.userid
			where s1.ctime between ? and ? and s1.order_status = 2 %s
			group by sdate
		) t0
		group by sdate
	`, where_u_sql, where_u_sql), pay_withdraw_args...)
	if err != nil {
		return
	}
	for _, data := range pay_withdraw_data {
		sdate := data["sdate"].(uint32)
		pays := utils.ToInt64(data["pays"])
		withdraws := utils.ToInt64(data["withdraws"])
		cash_pay_give := utils.ToInt64(data["cash_pay_give"])
		if stat, ok := dateStatsMap[sdate]; ok {
			stat.PaySubWithdraws0 = pays - withdraws
			stat.CashPayGive0 = cash_pay_give
		}
	}

	// vip: 97,98,99
	// 转盘: 135
	// 代理: 140,141,142,144
	// 礼包码: 143
	// 波动返水: 145
	ltypes := []int32{97, 98, 99, 135, 140, 141, 142, 144, 143, 145}
	// cash 发放
	var cash_datas []map[string]any
	var cash_args = []any{stime, etime, ltypes}
	cash_args = append(cash_args, where_u_args...)
	err = ck.Select(&cash_datas, fmt.Sprintf(`
		select toYYYYMMDD(s1.ctime) sdate, s1.ltype, SUM(s1.add_diamond) cashs
		from game.col_user s0 final
		join game.col_log_water s1 final on s0.userid = s1.userid
		where s1.ctime between ? and ? and s1.ltype in ? and s1.add_diamond > 0 %s
		group by sdate, s1.ltype
	`, where_u_sql), cash_args...)
	if err != nil {
		return
	}
	for _, data := range cash_datas {
		sdate := data["sdate"].(uint32)
		ltype := utils.ToInt64(data["ltype"])
		cashs := utils.ToInt64(data["cashs"])
		if stat, ok := dateStatsMap[sdate]; ok {
			switch ltype {
			// vip
			case 97, 98, 99:
				stat.CashVip0 += cashs
				if ltype == 98 {
					stat.CashVipWeek0 = cashs
				} else if ltype == 99 {
					stat.CashVipUpgrade0 = cashs
				}
			// 代理
			case 140, 141, 142, 144:
				stat.CashShareAgent0 += cashs
				switch ltype {
				case 141:
					stat.CashShareAgentHeads0 += cashs
				case 142:
					stat.CashShareAgentBets0 += cashs
				case 144:
					stat.CashShareAgentTasks0 += cashs
				}
			// 转盘
			case 135:
				stat.CashTurn0 = cashs
			// 波动返水
			case 145:
				stat.CashSubsidy0 = cashs
			// 礼包码
			case 143:
				stat.CashGivePack0 = cashs
			}
		}
	}

	// bonus 发放
	var bonus_datas []map[string]any
	var bonus_args = []any{stime.Unix(), etime.Unix(), ltypes}
	bonus_args = append(bonus_args, where_u_args...)
	err = ck.Select(&bonus_datas, fmt.Sprintf(`
		select toYYYYMMDD(toDateTime(s1.ctime)) sdate, s1.reason, SUM(s1.vb) bonus
		from game.col_user s0 final
		join game.col_log_vbdiamond s1 final on s0.userid = s1.userid
		where s1.ctime between ? and ? and s1.reason in ? %s
		group by sdate, s1.reason
	`, where_u_sql), bonus_args...)
	if err != nil {
		return
	}
	for _, data := range bonus_datas {
		sdate := data["sdate"].(uint32)
		reason := utils.ToInt64(data["reason"])
		bonus := utils.ToInt64(data["bonus"])
		if stat, ok := dateStatsMap[sdate]; ok {
			switch reason {
			// vip
			case 97, 98, 99:
				stat.BonusVip0 += bonus
				if reason == 98 {
					stat.BonusVipWeek0 = bonus
				} else if reason == 99 {
					stat.BonusVipUpgrade0 = bonus
				}
			// 代理
			case 140, 141, 142, 144:
				stat.BonusShareAgent0 += bonus
				switch reason {
				case 141:
					stat.BonusShareAgentHeads0 += bonus
				case 142:
					stat.BonusShareAgentBets0 += bonus
				case 144:
					stat.BonusShareAgentTasks0 += bonus
				}
			// 转盘
			case 135:
				stat.BonusTurn0 = bonus
			// 波动返水
			case 145:
				stat.BonusSubsidy0 = bonus
			// 礼包码
			case 143:
				stat.BonusGivePack0 = bonus
			}
		}
	}

	// 周卡: 59, desc contails 每日
	var week_card_datas []map[string]any
	var week_card_args = []any{stime, etime}
	week_card_args = append(week_card_args, where_u_args...)
	err = ck.Select(&week_card_datas, fmt.Sprintf(`
		select toYYYYMMDD(s1.ctime) sdate, SUM(s1.add_diamond) cashs
		from game.col_user s0 final
		join game.col_log_water s1 final on s0.userid = s1.userid
		where s1.ctime between ? and ? and s1.ltype = 59 and s1.water_desc like '%%每日%%' %s
		group by sdate
	`, where_u_sql), week_card_args...)
	if err != nil {
		return
	}
	for _, data := range week_card_datas {
		sdate := data["sdate"].(uint32)
		cashs := utils.ToInt64(data["cashs"])
		if stat, ok := dateStatsMap[sdate]; ok {
			stat.CashWeekCard0 = cashs
		}
	}

	// 排行榜 日周月 cash, bonus
	var betrank_datas []map[string]any
	var betrank_args = []any{stime, etime}
	betrank_args = append(betrank_args, where_u_args...)
	err = ck.Select(&betrank_datas, fmt.Sprintf(`
		select toYYYYMMDD(s1.ctime) sdate, s1.rank_type, s1.prize_type, SUM(prize) prizes
		from game.col_user s0 final
		join game.col_activity_betrank_prize s1 final on s0.userid = s1.userid
		where s1.ctime between ? and ? and s1.robot = 0 %s
		group by sdate, s1.rank_type, s1.prize_type
	`, where_u_sql), betrank_args...)
	if err != nil {
		return
	}
	for _, data := range betrank_datas {
		sdate := data["sdate"].(uint32)
		rank_type := utils.ToInt64(data["rank_type"])
		prize_type := utils.ToInt64(data["prize_type"])
		prizes := utils.ToInt64(data["prizes"])
		if stat, ok := dateStatsMap[sdate]; ok {
			switch prize_type {
			case 1: // bonus
				stat.BonusBetRank0 += prizes
				switch rank_type {
				case 1:
					stat.BonusBetRankDaily0 = prizes
				case 2:
					stat.BonusBetRankWeekly0 = prizes
				case 3:
					stat.BonusBetRankMonthly0 = prizes
				}
			case 2: // cash
				stat.CashBetRank0 += prizes
				switch rank_type {
				case 1:
					stat.CashBetRankDaily0 = prizes
				case 2:
					stat.CashBetRankWeekly0 = prizes
				case 3:
					stat.CashBetRankMonthly0 = prizes
				}
			}
		}
	}

	// 充值订单 bonus 发放
	// 首充二充三充
	pay1ShopIds := []string{"lb1c1", "lb1c2", "lb1c3", "lb1c4", "lb1c5", "lb1c6", "lb1c7", "lb1c8"}
	pay2ShopIds := []string{"lbec1", "lbec2", "lbec3", "lbec4", "lbec5", "lbec6", "lbec7", "lbec8"}
	pay3ShopIds := []string{"lbsc1", "lbsc2", "lbsc3", "lbsc4", "lbsc5", "lbsc6", "lbsc7", "lbsc8"}
	var pay_datas []map[string]any
	var pay_args = []any{pay1ShopIds, pay2ShopIds, pay3ShopIds, stime, etime}
	pay_args = append(pay_args, where_u_args...)
	err = ck.Select(&pay_datas, fmt.Sprintf(`
		select toYYYYMMDD(s1.ctime) sdate, 
			(case when s1.shop_type = 3 then 3 
				when s1.shop_id in ? then 11 
				when s1.shop_id in ? then 12 
				when s1.shop_id in ? then 13
				else 0 end) pay_type,
			SUM(toInt64OrZero(s1.other_present)) gives
		from game.col_user s0 final
		join game.col_trade_record s1 final on s1.userid = s0.userid
		where s1.ctime between ? and ? and s1.order_status = 4 and toInt64OrZero(s1.other_present) > 0 %s
		group by sdate, pay_type
	`, where_u_sql), pay_args...)
	if err != nil {
		return
	}
	for _, data := range pay_datas {
		sdate := data["sdate"].(uint32)
		pay_type := utils.ToInt64(data["pay_type"])
		gives := utils.ToInt64(data["gives"])
		if stat, ok := dateStatsMap[sdate]; ok {
			switch pay_type {
			case 3:
				stat.BonusWeekCard0 = gives
			case 11:
				stat.BonusPay1Give0 = gives
			case 12:
				stat.BonusPay2Give0 = gives
			case 13:
				stat.BonusPay3Give0 = gives
			default:
				stat.BonusPayGive0 = gives
			}
		}
	}

	summary := &entity.FinanceStat{
		SDate: fmt.Sprintf("%s~%s汇总", stime.Format(utils.FORMAT_DATE), etime.Format(utils.FORMAT_DATE)),
	}
	for _, stat := range stats {
		summary.Cash0 += stat.Cash0
		summary.Withdrawable0 += stat.Withdrawable0
		summary.NonWithdrawable0 += stat.NonWithdrawable0
		summary.LiveCash0 += stat.LiveCash0
		summary.LiveWithdrawable0 += stat.LiveWithdrawable0
		summary.LiveNonWithdrawable0 += stat.LiveNonWithdrawable0

		summary.LoginUsers += stat.LoginUsers
		summary.GameIncome0 += stat.GameIncome0
		summary.BonusGift0 += stat.BonusGift0
		summary.CashFlow0 += stat.CashFlow0
		summary.CashGift0 += stat.CashGift0
		summary.CashVb0 += stat.CashVb0
		summary.PaySubWithdraws0 += stat.PaySubWithdraws0

		summary.BonusVip0 += stat.BonusVip0
		summary.BonusVipUpgrade0 += stat.BonusVipUpgrade0
		summary.BonusVipWeek0 += stat.BonusVipWeek0
		summary.BonusBetRank0 += stat.BonusBetRank0
		summary.BonusBetRankDaily0 += stat.BonusBetRankDaily0
		summary.BonusBetRankWeekly0 += stat.BonusBetRankWeekly0
		summary.BonusBetRankMonthly0 += stat.BonusBetRankMonthly0
		summary.BonusShareAgent0 += stat.BonusShareAgent0
		summary.BonusShareAgentBets0 += stat.BonusShareAgentBets0
		summary.BonusShareAgentHeads0 += stat.BonusShareAgentHeads0
		summary.BonusShareAgentTasks0 += stat.BonusShareAgentTasks0
		summary.BonusTurn0 += stat.BonusTurn0
		summary.BonusSubsidy0 += stat.BonusSubsidy0
		summary.BonusGivePack0 += stat.BonusGivePack0
		summary.BonusWeekCard0 += stat.BonusWeekCard0
		summary.BonusPay1Give0 += stat.BonusPay1Give0
		summary.BonusPay2Give0 += stat.BonusPay2Give0
		summary.BonusPay3Give0 += stat.BonusPay3Give0
		summary.BonusPayGive0 += stat.BonusPayGive0

		summary.CashVip0 += stat.CashVip0
		summary.CashVipUpgrade0 += stat.CashVipUpgrade0
		summary.CashVipWeek0 += stat.CashVipWeek0
		summary.CashBetRank0 += stat.CashBetRank0
		summary.CashBetRankDaily0 += stat.CashBetRankDaily0
		summary.CashBetRankWeekly0 += stat.CashBetRankWeekly0
		summary.CashBetRankMonthly0 += stat.CashBetRankMonthly0
		summary.CashShareAgent0 += stat.CashShareAgent0
		summary.CashShareAgentBets0 += stat.CashShareAgentBets0
		summary.CashShareAgentHeads0 += stat.CashShareAgentHeads0
		summary.CashShareAgentTasks0 += stat.CashShareAgentTasks0
		summary.CashTurn0 += stat.CashTurn0
		summary.CashSubsidy0 += stat.CashSubsidy0
		summary.CashGivePack0 += stat.CashGivePack0
		summary.CashWeekCard0 += stat.CashWeekCard0
		summary.CashPay1Give0 += stat.CashPay1Give0
		summary.CashPay2Give0 += stat.CashPay2Give0
		summary.CashPay3Give0 += stat.CashPay3Give0
		summary.CashPayGive0 += stat.CashPayGive0
	}

	avg := &entity.FinanceStat{
		SDate: fmt.Sprintf("%s~%s日均值", stime.Format(utils.FORMAT_DATE), etime.Format(utils.FORMAT_DATE)),
	}
	days := int64(len(stats))
	avg.Cash0 = int64(ComputeFloat(summary.Cash0, days))
	avg.Withdrawable0 = int64(ComputeFloat(summary.Withdrawable0, days))
	avg.NonWithdrawable0 = int64(ComputeFloat(summary.NonWithdrawable0, days))
	avg.LiveCash0 = int64(ComputeFloat(summary.LiveCash0, days))
	avg.LiveWithdrawable0 = int64(ComputeFloat(summary.LiveWithdrawable0, days))
	avg.LiveNonWithdrawable0 = int64(ComputeFloat(summary.LiveNonWithdrawable0, days))
	avg.LoginUsers = int64(ComputeFloat(summary.LoginUsers, days))
	avg.GameIncome0 = int64(ComputeFloat(summary.GameIncome0, days))
	avg.BonusGift0 = int64(ComputeFloat(summary.BonusGift0, days))
	avg.CashFlow0 = int64(ComputeFloat(summary.CashFlow0, days))
	avg.CashGift0 = int64(ComputeFloat(summary.CashGift0, days))
	avg.CashVb0 = int64(ComputeFloat(summary.CashVb0, days))
	avg.PaySubWithdraws0 = int64(ComputeFloat(summary.PaySubWithdraws0, days))
	avg.BonusVip0 = int64(ComputeFloat(summary.BonusVip0, days))
	avg.BonusVipUpgrade0 = int64(ComputeFloat(summary.BonusVipUpgrade0, days))
	avg.BonusVipWeek0 = int64(ComputeFloat(summary.BonusVipWeek0, days))
	avg.BonusBetRank0 = int64(ComputeFloat(summary.BonusBetRank0, days))
	avg.BonusBetRankDaily0 = int64(ComputeFloat(summary.BonusBetRankDaily0, days))
	avg.BonusBetRankWeekly0 = int64(ComputeFloat(summary.BonusBetRankWeekly0, days))
	avg.BonusBetRankMonthly0 = int64(ComputeFloat(summary.BonusBetRankMonthly0, days))
	avg.BonusShareAgent0 = int64(ComputeFloat(summary.BonusShareAgent0, days))
	avg.BonusShareAgentBets0 = int64(ComputeFloat(summary.BonusShareAgentBets0, days))
	avg.BonusShareAgentHeads0 = int64(ComputeFloat(summary.BonusShareAgentHeads0, days))
	avg.BonusShareAgentTasks0 = int64(ComputeFloat(summary.BonusShareAgentTasks0, days))
	avg.BonusTurn0 = int64(ComputeFloat(summary.BonusTurn0, days))
	avg.BonusSubsidy0 = int64(ComputeFloat(summary.BonusSubsidy0, days))
	avg.BonusGivePack0 = int64(ComputeFloat(summary.BonusGivePack0, days))
	avg.BonusWeekCard0 = int64(ComputeFloat(summary.BonusWeekCard0, days))
	avg.BonusPay1Give0 = int64(ComputeFloat(summary.BonusPay1Give0, days))
	avg.BonusPay2Give0 = int64(ComputeFloat(summary.BonusPay2Give0, days))
	avg.BonusPay3Give0 = int64(ComputeFloat(summary.BonusPay3Give0, days))
	avg.BonusPayGive0 = int64(ComputeFloat(summary.BonusPayGive0, days))
	avg.CashVip0 = int64(ComputeFloat(summary.CashVip0, days))
	avg.CashVipUpgrade0 = int64(ComputeFloat(summary.CashVipUpgrade0, days))
	avg.CashVipWeek0 = int64(ComputeFloat(summary.CashVipWeek0, days))
	avg.CashBetRank0 = int64(ComputeFloat(summary.CashBetRank0, days))
	avg.CashBetRankDaily0 = int64(ComputeFloat(summary.CashBetRankDaily0, days))
	avg.CashBetRankWeekly0 = int64(ComputeFloat(summary.CashBetRankWeekly0, days))
	avg.CashBetRankMonthly0 = int64(ComputeFloat(summary.CashBetRankMonthly0, days))
	avg.CashShareAgent0 = int64(ComputeFloat(summary.CashShareAgent0, days))
	avg.CashShareAgentBets0 = int64(ComputeFloat(summary.CashShareAgentBets0, days))
	avg.CashShareAgentHeads0 = int64(ComputeFloat(summary.CashShareAgentHeads0, days))
	avg.CashShareAgentTasks0 = int64(ComputeFloat(summary.CashShareAgentTasks0, days))
	avg.CashTurn0 = int64(ComputeFloat(summary.CashTurn0, days))
	avg.CashSubsidy0 = int64(ComputeFloat(summary.CashSubsidy0, days))
	avg.CashGivePack0 = int64(ComputeFloat(summary.CashGivePack0, days))
	avg.CashWeekCard0 = int64(ComputeFloat(summary.CashWeekCard0, days))
	avg.CashPay1Give0 = int64(ComputeFloat(summary.CashPay1Give0, days))
	avg.CashPay2Give0 = int64(ComputeFloat(summary.CashPay2Give0, days))
	avg.CashPay3Give0 = int64(ComputeFloat(summary.CashPay3Give0, days))
	avg.CashPayGive0 = int64(ComputeFloat(summary.CashPayGive0, days))

	stats = append([]*entity.FinanceStat{summary, avg}, stats...)

	for _, stat := range stats {
		stat.Cash = fmt.Sprintf("%.2f", Chip2Float(stat.Cash0))
		stat.Withdrawable = fmt.Sprintf("%.2f", Chip2Float(stat.Withdrawable0))
		stat.NonWithdrawable = fmt.Sprintf("%.2f", Chip2Float(stat.NonWithdrawable0))
		stat.LiveCash = fmt.Sprintf("%.2f", Chip2Float(stat.LiveCash0))
		stat.LiveWithdrawable = fmt.Sprintf("%.2f", Chip2Float(stat.LiveWithdrawable0))
		stat.LiveNonWithdrawable = fmt.Sprintf("%.2f", Chip2Float(stat.LiveNonWithdrawable0))
		stat.WithdrawableRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.Withdrawable0, stat.Cash0)*100)
		stat.LiveWithdrawableRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.LiveWithdrawable0, stat.LiveCash0)*100)

		stat.GameIncome = fmt.Sprintf("%.2f", Chip2Float(stat.GameIncome0))
		stat.BonusGift = fmt.Sprintf("%.2f", Chip2Float(stat.BonusGift0))
		stat.BonusIncomeRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.BonusGift0, stat.GameIncome0)*100)
		stat.CashFlow = fmt.Sprintf("%.2f", Chip2Float(stat.CashFlow0))
		stat.CashGift = fmt.Sprintf("%.2f", Chip2Float(stat.CashGift0))
		stat.CashGiftIncomeRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.CashGift0, stat.GameIncome0)*100)
		stat.CashVb = fmt.Sprintf("%.2f", Chip2Float(stat.CashVb0))
		stat.CashVbIncomeRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.CashVb0, stat.GameIncome0)*100)
		stat.PaySubWithdraws = fmt.Sprintf("%.2f", Chip2Float(stat.PaySubWithdraws0))
		stat.GameIncomeSubCashFlow = fmt.Sprintf("%.2f", Chip2Float(stat.GameIncome0-stat.CashFlow0))
		stat.SurplusMiss = fmt.Sprintf("%.2f", Chip2Float(stat.PaySubWithdraws0-stat.GameIncome0+stat.CashFlow0))

		stat.BonusVip = fmt.Sprintf("%.2f", Chip2Float(stat.BonusVip0))
		stat.BonusVipUpgrade = fmt.Sprintf("%.2f", Chip2Float(stat.BonusVipUpgrade0))
		stat.BonusVipWeek = fmt.Sprintf("%.2f", Chip2Float(stat.BonusVipWeek0))
		stat.BonusBetRank = fmt.Sprintf("%.2f", Chip2Float(stat.BonusBetRank0))
		stat.BonusBetRankDaily = fmt.Sprintf("%.2f", Chip2Float(stat.BonusBetRankDaily0))
		stat.BonusBetRankWeekly = fmt.Sprintf("%.2f", Chip2Float(stat.BonusBetRankWeekly0))
		stat.BonusBetRankMonthly = fmt.Sprintf("%.2f", Chip2Float(stat.BonusBetRankMonthly0))
		stat.BonusShareAgent = fmt.Sprintf("%.2f", Chip2Float(stat.BonusShareAgent0))
		stat.BonusShareAgentBets = fmt.Sprintf("%.2f", Chip2Float(stat.BonusShareAgentBets0))
		stat.BonusShareAgentHeads = fmt.Sprintf("%.2f", Chip2Float(stat.BonusShareAgentHeads0))
		stat.BonusShareAgentTasks = fmt.Sprintf("%.2f", Chip2Float(stat.BonusShareAgentTasks0))
		stat.BonusTurn = fmt.Sprintf("%.2f", Chip2Float(stat.BonusTurn0))
		stat.BonusSubsidy = fmt.Sprintf("%.2f", Chip2Float(stat.BonusSubsidy0))
		stat.BonusGivePack = fmt.Sprintf("%.2f", Chip2Float(stat.BonusGivePack0))
		stat.BonusWeekCard = fmt.Sprintf("%.2f", Chip2Float(stat.BonusWeekCard0))
		stat.BonusPay1Give = fmt.Sprintf("%.2f", Chip2Float(stat.BonusPay1Give0))
		stat.BonusPay2Give = fmt.Sprintf("%.2f", Chip2Float(stat.BonusPay2Give0))
		stat.BonusPay3Give = fmt.Sprintf("%.2f", Chip2Float(stat.BonusPay3Give0))
		stat.BonusPayGive = fmt.Sprintf("%.2f", Chip2Float(stat.BonusPayGive0))

		stat.CashVip = fmt.Sprintf("%.2f", Chip2Float(stat.CashVip0))
		stat.CashVipUpgrade = fmt.Sprintf("%.2f", Chip2Float(stat.CashVipUpgrade0))
		stat.CashVipWeek = fmt.Sprintf("%.2f", Chip2Float(stat.CashVipWeek0))
		stat.CashBetRank = fmt.Sprintf("%.2f", Chip2Float(stat.CashBetRank0))
		stat.CashBetRankDaily = fmt.Sprintf("%.2f", Chip2Float(stat.CashBetRankDaily0))
		stat.CashBetRankWeekly = fmt.Sprintf("%.2f", Chip2Float(stat.CashBetRankWeekly0))
		stat.CashBetRankMonthly = fmt.Sprintf("%.2f", Chip2Float(stat.CashBetRankMonthly0))
		stat.CashShareAgent = fmt.Sprintf("%.2f", Chip2Float(stat.CashShareAgent0))
		stat.CashShareAgentBets = fmt.Sprintf("%.2f", Chip2Float(stat.CashShareAgentBets0))
		stat.CashShareAgentHeads = fmt.Sprintf("%.2f", Chip2Float(stat.CashShareAgentHeads0))
		stat.CashShareAgentTasks = fmt.Sprintf("%.2f", Chip2Float(stat.CashShareAgentTasks0))
		stat.CashTurn = fmt.Sprintf("%.2f", Chip2Float(stat.CashTurn0))
		stat.CashSubsidy = fmt.Sprintf("%.2f", Chip2Float(stat.CashSubsidy0))
		stat.CashGivePack = fmt.Sprintf("%.2f", Chip2Float(stat.CashGivePack0))
		stat.CashWeekCard = fmt.Sprintf("%.2f", Chip2Float(stat.CashWeekCard0))
		stat.CashPay1Give = fmt.Sprintf("%.2f", Chip2Float(stat.CashPay1Give0))
		stat.CashPay2Give = fmt.Sprintf("%.2f", Chip2Float(stat.CashPay2Give0))
		stat.CashPay3Give = fmt.Sprintf("%.2f", Chip2Float(stat.CashPay3Give0))
		stat.CashPayGive = fmt.Sprintf("%.2f", Chip2Float(stat.CashPayGive0))

		stat.BonusVipRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.BonusVip0, stat.BonusGift0)*100)
		stat.BonusVipUpgradeRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.BonusVipUpgrade0, stat.BonusGift0)*100)
		stat.BonusVipWeekRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.BonusVipWeek0, stat.BonusGift0)*100)
		stat.BonusBetRankRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.BonusBetRank0, stat.BonusGift0)*100)
		stat.BonusBetRankDailyRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.BonusBetRankDaily0, stat.BonusGift0)*100)
		stat.BonusBetRankWeeklyRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.BonusBetRankWeekly0, stat.BonusGift0)*100)
		stat.BonusBetRankMonthlyRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.BonusBetRankMonthly0, stat.BonusGift0)*100)
		stat.BonusShareAgentRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.BonusShareAgent0, stat.BonusGift0)*100)
		stat.BonusShareAgentBetsRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.BonusShareAgentBets0, stat.BonusGift0)*100)
		stat.BonusShareAgentHeadsRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.BonusShareAgentHeads0, stat.BonusGift0)*100)
		stat.BonusShareAgentTasksRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.BonusShareAgentTasks0, stat.BonusGift0)*100)
		stat.BonusTurnRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.BonusTurn0, stat.BonusGift0)*100)
		stat.BonusSubsidyRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.BonusSubsidy0, stat.BonusGift0)*100)
		stat.BonusGivePackRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.BonusGivePack0, stat.BonusGift0)*100)
		stat.BonusWeekCardRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.BonusWeekCard0, stat.BonusGift0)*100)
		stat.BonusPay1GiveRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.BonusPay1Give0, stat.BonusGift0)*100)
		stat.BonusPay2GiveRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.BonusPay2Give0, stat.BonusGift0)*100)
		stat.BonusPay3GiveRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.BonusPay3Give0, stat.BonusGift0)*100)
		stat.BonusPayGiveRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.BonusPayGive0, stat.BonusGift0)*100)

		stat.CashVipRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.CashVip0, stat.CashGift0)*100)
		stat.CashVipUpgradeRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.CashVipUpgrade0, stat.CashGift0)*100)
		stat.CashVipWeekRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.CashVipWeek0, stat.CashGift0)*100)
		stat.CashBetRankRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.CashBetRank0, stat.CashGift0)*100)
		stat.CashBetRankDailyRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.CashBetRankDaily0, stat.CashGift0)*100)
		stat.CashBetRankWeeklyRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.CashBetRankWeekly0, stat.CashGift0)*100)
		stat.CashBetRankMonthlyRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.CashBetRankMonthly0, stat.CashGift0)*100)
		stat.CashShareAgentRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.CashShareAgent0, stat.CashGift0)*100)
		stat.CashShareAgentBetsRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.CashShareAgentBets0, stat.CashGift0)*100)
		stat.CashShareAgentHeadsRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.CashShareAgentHeads0, stat.CashGift0)*100)
		stat.CashShareAgentTasksRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.CashShareAgentTasks0, stat.CashGift0)*100)
		stat.CashTurnRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.CashTurn0, stat.CashGift0)*100)
		stat.CashSubsidyRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.CashSubsidy0, stat.CashGift0)*100)
		stat.CashGivePackRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.CashGivePack0, stat.CashGift0)*100)
		stat.CashWeekCardRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.CashWeekCard0, stat.CashGift0)*100)
		stat.CashPay1GiveRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.CashPay1Give0, stat.CashGift0)*100)
		stat.CashPay2GiveRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.CashPay2Give0, stat.CashGift0)*100)
		stat.CashPay3GiveRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.CashPay3Give0, stat.CashGift0)*100)
		stat.CashPayGiveRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.CashPayGive0, stat.CashGift0)*100)
	}

	summary.Cash = "-"
	summary.Withdrawable = "-"
	summary.NonWithdrawable = "-"
	summary.LiveCash = "-"
	summary.LiveWithdrawable = "-"
	summary.LiveNonWithdrawable = "-"
	summary.WithdrawableRate = "-"
	summary.LiveWithdrawableRate = "-"
	return
}

// 游戏行为日报
func (c *statistics2Service) GameBetStats(stime, etime time.Time, registAreas []int32, packageIds []string) (stats []*entity.GameBetStat, err error) {
	where_u_sql := ""
	var where_u_args []any

	if len(registAreas) > 0 {
		where_u_sql += " and s0.regist_area in ?"
		where_u_args = append(where_u_args, registAreas)
	}
	if len(packageIds) > 0 {
		where_u_sql += " and s0.ad__bundle_id IN ?"
		where_u_args = append(where_u_args, packageIds)
	}

	var dateStats = make(map[uint32]*entity.GameBetStat)

	// 日活 老用户日活 付费用户日活
	var login_datas []map[string]any
	login_args := append([]any{stime, etime}, where_u_args...)
	_ = `
		select login_date, count(*) login_users,
			SUM(case when login_date = regist_date then 1 else 0 end) new_login_users,
			(login_users - new_login_users) old_login_users,
			SUM(case when payed > 0 and login_date >= first_pay_date then 1 else 0 end) pay_login_users,
			SUM(case when login_date = regist_date and payed > 0 and login_date >= first_pay_date then 1 else 0 end) new_pay_login_users,
			(pay_login_users - new_pay_login_users) old_pay_login_users
		from (
			select s1.login_date, s1.userid, toYYYYMMDD(min(s0.ctime)) regist_date, toYYYYMMDD(min(s2.ctime)) first_pay_date, max(s2.amount) payed
			from game.col_trade_record s2 final 
			right join game.col_user s0 final on s0.userid = s2.userid and s2.order_status = 4
			right join (
				select toYYYYMMDD(login_time) login_date, userid
				from game.col_log_login final
				where login_time between ? and ?
				group by login_date, userid
			) s1 on s1.userid = s0.userid
			where 1 = 1 %s
			group by s1.login_date, s1.userid
		) s1
		group by login_date
		order by login_date desc
	`
	err = ck.Select(&login_datas, fmt.Sprintf(`
		select s1.login_date, count(*) login_users,
			SUM(case when s1.login_date = toYYYYMMDD(s0.ctime) then 1 else 0 end) new_login_users,
			(login_users - new_login_users) old_login_users,
			SUM(case when s0.first_charge_val > 0 and s1.login_date >= toYYYYMMDD(toDateTime(s0.first_charge_time/1000)) then 1 else 0 end) pay_login_users,
			SUM(case when s1.login_date = toYYYYMMDD(s0.ctime) and s0.first_charge_val > 0 and s1.login_date >= toYYYYMMDD(toDateTime(s0.first_charge_time/1000)) then 1 else 0 end) new_pay_login_users,
			(pay_login_users - new_pay_login_users) old_pay_login_users
		from game.col_user s0 final
		join (
			select toYYYYMMDD(login_time) login_date, userid
			from game.col_log_login final
			where login_time between ? and ?
			group by login_date, userid
		) s1 on s1.userid = s0.userid
		where 1 = 1 %s
		group by s1.login_date
		order by s1.login_date desc
	`, where_u_sql), login_args...)
	if err != nil {
		return
	}
	for _, data := range login_datas {
		login_date := data["login_date"].(uint32)
		login_users := utils.ToInt64(data["login_users"])
		new_login_users := utils.ToInt64(data["new_login_users"])
		old_login_users := utils.ToInt64(data["old_login_users"])
		pay_login_users := utils.ToInt64(data["pay_login_users"])
		new_pay_login_users := utils.ToInt64(data["new_pay_login_users"])
		old_pay_login_users := utils.ToInt64(data["old_pay_login_users"])

		sdate := fmt.Sprint(login_date)
		stat := &entity.GameBetStat{
			SDate: sdate[0:4] + "-" + sdate[4:6] + "-" + sdate[6:8],
		}
		stats = append(stats, stat)
		dateStats[login_date] = stat

		stat.LoginUsers = login_users
		stat.NewLoginUsers = new_login_users
		stat.OldLoginUsers = old_login_users
		stat.PayLoginUsers = pay_login_users
		stat.NewPayLoginUsers0 = new_pay_login_users
		stat.OldPayLoginUsers0 = old_pay_login_users
	}

	// 打码投注
	var game_bet_datas []map[string]any
	external_args := append([]any{stime.Unix(), etime.Unix(), stime.Unix(), etime.Unix(), stime.Unix(), etime.Unix()}, where_u_args...)
	err = ck.Select(&game_bet_datas, fmt.Sprintf(`
		SELECT s1.sdate sdate, SUM(s1.bets0) bets, SUM(s1.rebate0) rebate,
			SUM(case when sdate = toYYYYMMDD(s0.ctime) then s1.bets0 else 0 end) new_bets,
			(bets - new_bets) old_bets,
			SUM(case when sdate = toYYYYMMDD(s0.ctime) then s1.rebate0 else 0 end) new_rebate,
			(rebate - new_rebate) old_rebate,
			count(distinct s1.userid) bet_users,
			count(distinct (case when s1.sdate = toYYYYMMDD(s0.ctime) then s1.userid else null end)) new_bet_users,
			(bet_users - new_bet_users) old_bet_users,
			count(distinct case when s0.first_charge_val > 0 and sdate >= toYYYYMMDD(toDateTime(s0.first_charge_time/1000)) then s1.userid else null end) pay_bet_users,
			count(distinct case when sdate = toYYYYMMDD(s0.ctime) and s0.first_charge_val > 0 and sdate >= toYYYYMMDD(toDateTime(s0.first_charge_time/1000)) then s1.userid else null end) new_pay_bet_users,
			(pay_bet_users - new_pay_bet_users) old_pay_bet_users
		FROM game.col_user s0 final
		JOIN (
			select toYYYYMMDD(toDateTime(s1.begin_time)) sdate, userid, SUM(s1.bet_amount) bets0, SUM(s1.bet_amount+s1.score) rebate0
			FROM game.col_detail s1 final
			where s1.begin_time between ? and ? and s1.robot = 0 and s1.win_type in (1,2,3)
			group by sdate, userid
				UNION ALL

			select toYYYYMMDD(toDateTime(s1.ctime)) sdate, user_id userid, SUM(s1.amount) bets0, 0 rebate0
			FROM game.col_nsq_log_external_bet s1 FINAL
			WHERE s1.ctime between ? and ? and s1.amount != 0 
			group by sdate, user_id
				UNION ALL

			select toYYYYMMDD(toDateTime(s1.ctime)) sdate, user_id userid, 0 bets0, SUM(s1.amount) rebate0
			FROM game.col_nsq_log_external_reward s1 FINAL
			WHERE s1.ctime between ? and ? and s1.amount != 0 
			group by sdate, user_id
		) s1 on s0.userid = s1.userid
		WHERE 1 = 1 %s
		GROUP BY s1.sdate
	`, where_u_sql), external_args...)
	if err != nil {
		return
	}
	for _, data := range game_bet_datas {
		sdate := data["sdate"].(uint32)
		bets := utils.ToInt64(data["bets"])
		rebate := utils.ToInt64(data["rebate"])
		new_bets := utils.ToInt64(data["new_bets"])
		old_bets := utils.ToInt64(data["old_bets"])
		new_rebate := utils.ToInt64(data["new_rebate"])
		old_rebate := utils.ToInt64(data["old_rebate"])
		bet_users := utils.ToInt64(data["bet_users"])
		new_bet_users := utils.ToInt64(data["new_bet_users"])
		old_bet_users := utils.ToInt64(data["old_bet_users"])
		pay_bet_users := utils.ToInt64(data["pay_bet_users"])
		new_pay_bet_users := utils.ToInt64(data["new_pay_bet_users"])
		old_pay_bet_users := utils.ToInt64(data["old_pay_bet_users"])
		if stat, ok := dateStats[sdate]; ok {
			stat.Bets0 += bets
			stat.NewBets0 += new_bets
			stat.OldBets0 += old_bets
			stat.Rebate0 += rebate
			stat.NewRebate0 += new_rebate
			stat.OldRebate0 += old_rebate
			stat.BetUsers += bet_users
			stat.NewBetUsers += new_bet_users
			stat.OldBetUsers += old_bet_users
			stat.PayBetUsers += pay_bet_users
			stat.NewPayBetUsers += new_pay_bet_users
			stat.OldPayBetUsers += old_pay_bet_users
		}
	}

	// 充值数据
	var pay_datas []map[string]any
	pay_args := append([]any{stime, etime}, where_u_args...)
	err = ck.Select(&pay_datas, fmt.Sprintf(`
		select toYYYYMMDD(s1.ctime) sdate, SUM(s1.amount) amounts,
			SUM(case when sdate = toYYYYMMDD(s0.ctime) then s1.amount else 0 end) new_amounts,
			(amounts - new_amounts) old_amounts
		FROM game.col_user s0 final
		JOIN game.col_trade_record s1 final ON s0.userid = s1.userid
		where s1.ctime between ? and ? and s1.order_status = 4 %s
		group by sdate
	`, where_u_sql), pay_args...)
	if err != nil {
		return
	}
	for _, data := range pay_datas {
		sdate := data["sdate"].(uint32)
		amounts := utils.ToInt64(data["amounts"])
		new_amounts := utils.ToInt64(data["new_amounts"])
		old_amounts := utils.ToInt64(data["old_amounts"])
		if stat, ok := dateStats[sdate]; ok {
			stat.Pays0 = amounts
			stat.NewPays0 = new_amounts
			stat.OldPays0 = old_amounts
		}
	}

	summary := &entity.GameBetStat{
		SDate: fmt.Sprintf("%s~%s汇总", stime.Format(utils.FORMAT_DATE), etime.Format(utils.FORMAT_DATE)),
	}
	for _, stat := range stats {
		summary.LoginUsers += stat.LoginUsers
		summary.NewLoginUsers += stat.NewLoginUsers
		summary.OldLoginUsers += stat.OldLoginUsers
		summary.PayLoginUsers += stat.PayLoginUsers
		summary.NewPayLoginUsers0 += stat.NewPayLoginUsers0
		summary.OldPayLoginUsers0 += stat.OldPayLoginUsers0
		summary.Bets0 += stat.Bets0
		summary.NewBets0 += stat.NewBets0
		summary.OldBets0 += stat.OldBets0
		summary.Rebate0 += stat.Rebate0
		summary.NewRebate0 += stat.NewRebate0
		summary.OldRebate0 += stat.OldRebate0
		summary.BetUsers += stat.BetUsers
		summary.NewBetUsers += stat.NewBetUsers
		summary.OldBetUsers += stat.OldBetUsers
		summary.Pays0 += stat.Pays0
		summary.NewPays0 += stat.NewPays0
		summary.OldPays0 += stat.OldPays0
		summary.PayBetUsers += stat.PayBetUsers
		summary.NewPayBetUsers += stat.NewPayBetUsers
		summary.OldPayBetUsers += stat.OldPayBetUsers
	}
	stats = append([]*entity.GameBetStat{summary}, stats...)

	for _, stat := range stats {
		stat.BetPayRate = fmt.Sprintf("%.2f", ComputeFloat(stat.Bets0, stat.Pays0))
		stat.NewBetPayRate = fmt.Sprintf("%.2f", ComputeFloat(stat.NewBets0, stat.NewPays0))
		stat.OldBetPayRate = fmt.Sprintf("%.2f", ComputeFloat(stat.OldBets0, stat.OldPays0))
		stat.Bets = fmt.Sprintf("%d万", stat.Bets0/100/10000)
		stat.NewBets = fmt.Sprintf("%d万 (%.2f%%)", stat.NewBets0/100/10000, ComputeFloat(stat.NewBets0, stat.Bets0)*100)
		stat.OldBets = fmt.Sprintf("%d万 (%.2f%%)", stat.OldBets0/100/10000, ComputeFloat(stat.OldBets0, stat.Bets0)*100)
		stat.BetsAvg = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(stat.Bets0, stat.BetUsers)))
		stat.NewBetsAvg = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(stat.NewBets0, stat.NewBetUsers)))
		stat.OldBetsAvg = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(stat.OldBets0, stat.OldBetUsers)))
		stat.BetsRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.BetUsers, stat.LoginUsers)*100)
		stat.NewBetsRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.NewBetUsers, stat.NewLoginUsers)*100)
		stat.OldBetsRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.OldBetUsers, stat.OldLoginUsers)*100)
		stat.PayBetsRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.PayBetUsers, stat.PayLoginUsers)*100)
		stat.NewPayBetsRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.NewPayBetUsers, stat.NewPayLoginUsers0)*100)
		stat.OldPayBetsRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.OldPayBetUsers, stat.OldPayLoginUsers0)*100)
		stat.RebateRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.Rebate0, stat.Bets0)*100)
		stat.NewRebateRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.NewRebate0, stat.NewBets0)*100)
		stat.OldRebateRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.OldRebate0, stat.OldBets0)*100)
		stat.Income0 = stat.Bets0 - stat.Rebate0
		stat.NewIncome0 = stat.NewBets0 - stat.NewRebate0
		stat.OldIncome0 = stat.OldBets0 - stat.OldRebate0
		stat.Income = fmt.Sprintf("%.2f万", Chip2Float(stat.Income0)/10000)
		stat.NewIncome = fmt.Sprintf("%.2f万", Chip2Float(stat.NewIncome0)/10000)
		stat.OldIncome = fmt.Sprintf("%.2f万", Chip2Float(stat.OldIncome0)/10000)

		stat.KillRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.Income0, stat.Pays0)*100)
		stat.NewKillRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.NewIncome0, stat.NewPays0)*100)
		stat.OldKillRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.OldIncome0, stat.OldPays0)*100)
	}
	return
}
