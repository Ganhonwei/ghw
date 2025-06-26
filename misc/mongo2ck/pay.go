package main

import (
	"goserver/pkg/data"
	"goserver/pkg/utils"
	"sync"
	"time"

	"github.com/globalsign/mgo/bson"
)

// 全量同步 user
func SyncFullUser(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	// m := []bson.M{
	// 	{"$match": bson.M{}},
	// 	// {"$sort": bson.M{"ctime": 1}},
	// }
	// var records []*data.User
	// SyncMongo2Clickhouse("User", data.PlayerUsers, m, records, 20000)

	begin, _ := time.ParseInLocation(utils.FORMAT, "2025-06-10 00:00:00", location)
	end, _ := time.ParseInLocation(utils.FORMAT, "2025-06-12 00:00:00", location)
	// begin, _ := time.Parse(utils.FORMAT, "2025-04-01 00:00:00")
	// end, _ := time.Parse(utils.FORMAT, "2025-04-18 00:00:00")
	var onceTime, concurrent int = 1, 2
	var newRecords = func() (r []*data.User) { return }
	var timeParse = func(t time.Time) any { return t }
	filter := bson.M{}
	// Syncmongo2ClickhouseByTime("User", data.PlayerUsers, newRecords, "ctime", begin, end, timeParse, onceTime, time.Hour*12, concurrent, filter)
	Syncmongo2ClickhouseByTime("User", data.PlayerUsers, newRecords, "login_time", begin, end, timeParse, onceTime, time.Hour, concurrent, filter)
}

// 全量同步 trade_record
func SyncFullTradeRecord(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	// m := []bson.M{
	// 	{"$match": bson.M{}},
	// 	// {"$sort": bson.M{"ctime": 1}},
	// }
	// var records []*data.TradeRecord
	// SyncMongo2Clickhouse("TradeRecords", data.TradeRecords, m, records, 50000)

	begin, _ := time.ParseInLocation(utils.FORMAT, "2025-06-10 00:00:00", location)
	end, _ := time.ParseInLocation(utils.FORMAT, "2025-06-12 00:00:00", location)
	// begin, _ := time.Parse(utils.FORMAT, "2024-01-01 00:00:00")
	// end, _ := time.Parse(utils.FORMAT, "2025-04-18 00:00:00")
	var onceTime, concurrent int = 1, 2
	var newRecords = func() (r []*data.TradeRecord) { return }
	var timeParse = func(t time.Time) any { return t }
	filter := bson.M{}
	Syncmongo2ClickhouseByTime("TradeRecords", data.TradeRecords, newRecords, "ctime", begin, end, timeParse, onceTime, time.Hour*24, concurrent, filter)
}

// 全量同步 trade_record
func SyncFullWithdrawRecord(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	// m := []bson.M{
	// 	{"$match": bson.M{}},
	// 	// {"$sort": bson.M{"ctime": 1}},
	// }
	// var records []*data.WithdrawRecord
	// SyncMongo2Clickhouse("WithdrawRecord", data.WithdrawRecords, m, records, 50000)

	begin, _ := time.ParseInLocation(utils.FORMAT, "2025-06-13 00:00:00", location)
	end, _ := time.ParseInLocation(utils.FORMAT, "2025-06-17 00:00:00", location)
	// begin, _ := time.Parse(utils.FORMAT, "2024-01-01 00:00:00")
	// end, _ := time.Parse(utils.FORMAT, "2025-04-18 00:00:00")
	var onceTime, concurrent int = 1, 2
	var newRecords = func() (r []*data.WithdrawRecord) { return }
	var timeParse = func(t time.Time) any { return t }
	filter := bson.M{}
	Syncmongo2ClickhouseByTime("WithdrawRecords", data.WithdrawRecords, newRecords, "ctime", begin, end, timeParse, onceTime, time.Hour*12, concurrent, filter)
}

// 全量同步 LogWater
func SyncFullLogWater(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	// m := []bson.M{
	// 	{"$match": bson.M{}},
	// 	// {"$sort": bson.M{"ctime": 1}},
	// }
	// var records []*data.LogWater
	// SyncMongo2Clickhouse("LogWater", data.LogWaters, m, records, 50000)

	// 按 ctime 数据分片
	begin, _ := time.ParseInLocation(utils.FORMAT, "2025-06-10 00:00:00", location)
	end, _ := time.ParseInLocation(utils.FORMAT, "2025-06-12 00:00:00", location)
	// begin, _ := time.Parse(utils.FORMAT, "2024-01-01 00:00:00")
	// end, _ := time.Parse(utils.FORMAT, "2025-04-18 00:00:00")
	// begin, _ := time.Parse(utils.FORMAT, "2024-07-22 00:00:00")
	// end, _ := time.Parse(utils.FORMAT, "2024-07-30 00:00:00")
	var onceTime, concurrent int = 1, 2
	var newRecords = func() (r []*data.LogWater) { return }
	var timeParse = func(t time.Time) any { return t }
	filter := bson.M{}
	Syncmongo2ClickhouseByTime("LogWaters", data.LogWaters, newRecords, "ctime", begin, end, timeParse, onceTime, time.Hour*24, concurrent, filter)
}

// 全量同步 detail
func SyncFullDetail(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	// 按 begin_time 数据分片
	// begin, _ := time.ParseInLocation(utils.FORMAT, "2025-04-24 00:00:00", location)
	// end, _ := time.ParseInLocation(utils.FORMAT, "2025-04-25 00:00:00", location)
	begin, _ := time.ParseInLocation(utils.FORMAT, "2025-06-10 00:00:00", location)
	end, _ := time.ParseInLocation(utils.FORMAT, "2025-06-12 00:00:00", location)
	var onceTime, concurrent int = 1, 2
	var newRecords = func() (r []*data.Detail) { return }
	var timeParse = func(t time.Time) any { return t.Unix() }
	// filter := bson.M{}
	filter := bson.M{"gtype": bson.M{"$in": []int32{4, 12, 1, 13}}} // rm, tp
	// filter := bson.M{"gtype": bson.M{"$in": []int32{1, 3, 7, 10, 9}}}
	// filter := bson.M{"gtype": bson.M{"$in": []int32{4, 12, 2, 3, 7, 10, 8, 9}}}
	Syncmongo2ClickhouseByTime("Details", data.Details, newRecords, "begin_time", begin, end, timeParse, onceTime, time.Hour, concurrent, filter)
}

// 全量同步 LogLogin
func SyncFullLogLogin(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	// m := []bson.M{
	// 	{"$match": bson.M{}},
	// 	// {"$sort": bson.M{"ctime": 1}},
	// }
	// var records []*data.LogLogin
	// SyncMongo2Clickhouse("LogLogin", data.LogLogins, m, records, 50000)

	begin, _ := time.ParseInLocation(utils.FORMAT, "2025-06-10 00:00:00", location)
	end, _ := time.ParseInLocation(utils.FORMAT, "2025-06-12 00:00:00", location)
	// begin, _ := time.Parse(utils.FORMAT, "2024-01-01 00:00:00")
	// end, _ := time.Parse(utils.FORMAT, "2025-04-18 00:00:00")
	// begin, _ := time.Parse(utils.FORMAT, "2024-07-22 00:00:00")
	// end, _ := time.Parse(utils.FORMAT, "2024-07-30 00:00:00")
	var onceTime, concurrent int = 1, 2
	var newRecords = func() (r []*data.LogLogin) { return }
	var timeParse = func(t time.Time) any { return t }
	filter := bson.M{}
	Syncmongo2ClickhouseByTime("LogLogins", data.LogLogins, newRecords, "login_time", begin, end, timeParse, onceTime, time.Hour*24, concurrent, filter)
}

// 全量同步 外接下注
func SyncFullExternalBet(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	begin, _ := time.ParseInLocation(utils.FORMAT, "2025-06-10 00:00:00", location)
	end, _ := time.ParseInLocation(utils.FORMAT, "2025-06-12 00:00:00", location)
	// begin, _ := time.Parse(utils.FORMAT, "2024-01-01 00:00:00")
	// end, _ := time.Parse(utils.FORMAT, "2025-04-18 00:00:00")
	var onceTime, concurrent int = 1, 2
	var newRecords = func() (r []*data.NsqLogExternalBet) { return }
	var timeParse = func(t time.Time) any { return t.Unix() }
	filter := bson.M{}
	Syncmongo2ClickhouseByTime("NsqLogExternalBets", data.NsqLogExternalBets, newRecords, "ctime", begin, end, timeParse, onceTime, time.Hour*24, concurrent, filter)
}

// 全量同步 外接返奖
func SyncFullExternalRewards(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	// 按 begin_time 数据分片
	begin, _ := time.ParseInLocation(utils.FORMAT, "2025-06-10 00:00:00", location)
	end, _ := time.ParseInLocation(utils.FORMAT, "2025-06-12 00:00:00", location)
	// begin, _ := time.Parse(utils.FORMAT, "2024-01-01 00:00:00")
	// end, _ := time.Parse(utils.FORMAT, "2025-04-18 00:00:00")
	var onceTime, concurrent int = 1, 2
	var newRecords = func() (r []*data.NsqLogExternalReward) { return }
	var timeParse = func(t time.Time) any { return t.Unix() }
	filter := bson.M{}
	Syncmongo2ClickhouseByTime("NsqLogExternalRewards", data.NsqLogExternalRewards, newRecords, "ctime", begin, end, timeParse, onceTime, time.Hour*24, concurrent, filter)
}

// 全量同步 外接取消
func SyncFullExternalCancel(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	begin, _ := time.ParseInLocation(utils.FORMAT, "2025-06-10 00:00:00", location)
	end, _ := time.ParseInLocation(utils.FORMAT, "2025-06-12 00:00:00", location)
	// begin, _ := time.Parse(utils.FORMAT, "2024-01-01 00:00:00")
	// end, _ := time.Parse(utils.FORMAT, "2025-04-18 00:00:00")
	var onceTime, concurrent int = 1, 2
	var newRecords = func() (r []*data.NsqLogExternalCancel) { return }
	var timeParse = func(t time.Time) any { return t.Unix() }
	filter := bson.M{}
	Syncmongo2ClickhouseByTime("NsqLogExternalCancels", data.NsqLogExternalCancels, newRecords, "ctime", begin, end, timeParse, onceTime, time.Hour*24, concurrent, filter)
}

// 全量同步 vbbank变化日志
func SyncFullLogVbDiamond(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	// 按 ctime 数据分片
	begin, _ := time.ParseInLocation(utils.FORMAT, "2025-06-10 00:00:00", location)
	end, _ := time.ParseInLocation(utils.FORMAT, "2025-06-12 00:00:00", location)
	// begin, _ := time.Parse(utils.FORMAT, "2024-01-01 00:00:00")
	// end, _ := time.Parse(utils.FORMAT, "2025-04-18 00:00:00")

	var onceTime, concurrent int = 1, 2
	var newRecords = func() (r []*data.LogVbDiamond) { return }
	var timeParse = func(t time.Time) any { return t.Unix() }
	filter := bson.M{}
	Syncmongo2ClickhouseByTime("LogVbDiamond", data.LogVBDiamonds, newRecords, "ctime", begin, end, timeParse, onceTime, time.Hour*24, concurrent, filter)
}

// 全量同步 vbbank变化日志
func SyncFullLogOutDiamond(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	// 按 ctime 数据分片
	begin, _ := time.ParseInLocation(utils.FORMAT, "2025-06-10 00:00:00", location)
	end, _ := time.ParseInLocation(utils.FORMAT, "2025-06-12 00:00:00", location)
	// begin, _ := time.Parse(utils.FORMAT, "2024-01-01 00:00:00")
	// end, _ := time.Parse(utils.FORMAT, "2025-04-18 00:00:00")

	var onceTime, concurrent int = 1, 2
	var newRecords = func() (r []*data.LogOutDiamond) { return }
	var timeParse = func(t time.Time) any { return t.Unix() }
	filter := bson.M{}
	Syncmongo2ClickhouseByTime("LogOutDiamond", data.LogOutDiamonds, newRecords, "ctime", begin, end, timeParse, onceTime, time.Hour*12, concurrent, filter)
}

// 全量同步 ActivityBetRankPrize
func SyncFullActivityBetRankPrize(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	// 按 ctime 数据分片
	begin, _ := time.ParseInLocation(utils.FORMAT, "2025-06-10 00:00:00", location)
	end, _ := time.ParseInLocation(utils.FORMAT, "2025-06-12 00:00:00", location)

	var onceTime, concurrent int = 1, 2
	var newRecords = func() (r []*data.ActivityBetRankPrize) { return }
	var timeParse = func(t time.Time) any { return t }
	filter := bson.M{}
	Syncmongo2ClickhouseByTime("ActivityBetRankPrize", data.ActivityBetRankPrizes, newRecords, "ctime", begin, end, timeParse, onceTime, time.Hour*12, concurrent, filter)
}

// 全量同步 ActivityTurnDrawLog
func SyncFullActivityTurnDrawLog(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	// 按 ctime 数据分片
	begin, _ := time.ParseInLocation(utils.FORMAT, "2025-06-10 00:00:00", location)
	end, _ := time.ParseInLocation(utils.FORMAT, "2025-06-12 00:00:00", location)

	var onceTime, concurrent int = 1, 2
	var newRecords = func() (r []*data.ActivityTurnDrawLog) { return }
	var timeParse = func(t time.Time) any { return t.Unix() }
	filter := bson.M{}
	Syncmongo2ClickhouseByTime("ActivityTurnDrawLog", data.ActivityTurnDrawLogs, newRecords, "ctime", begin, end, timeParse, onceTime, time.Hour*12, concurrent, filter)
}

// 全量同步 ActivityTurnPrizeLog
func SyncFullActivityTurnPrizeLog(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	// 按 ctime 数据分片
	begin, _ := time.ParseInLocation(utils.FORMAT, "2025-06-10 00:00:00", location)
	end, _ := time.ParseInLocation(utils.FORMAT, "2025-06-12 00:00:00", location)

	var onceTime, concurrent int = 1, 2
	var newRecords = func() (r []*data.ActivityTurnPrizeLog) { return }
	var timeParse = func(t time.Time) any { return t.Unix() }
	filter := bson.M{}
	Syncmongo2ClickhouseByTime("ActivityTurnDrawLog", data.ActivityTurnPrizeLogs, newRecords, "ctime", begin, end, timeParse, onceTime, time.Hour*12, concurrent, filter)
}

// 全量同步 ShareAgentIncomeRecord
func SyncFullShareAgentIncomeRecord(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	// 按 ctime 数据分片
	begin, _ := time.ParseInLocation(utils.FORMAT, "2025-06-10 00:00:00", location)
	end, _ := time.ParseInLocation(utils.FORMAT, "2025-06-12 00:00:00", location)

	var onceTime, concurrent int = 1, 2
	var newRecords = func() (r []*data.ShareAgentIncomeRecord) { return }
	var timeParse = func(t time.Time) any { return t.UnixMilli() }
	filter := bson.M{}
	Syncmongo2ClickhouseByTime("ShareAgentIncomeRecord", data.ShareAgentIncomeRecords, newRecords, "ctime", begin, end, timeParse, onceTime, time.Hour*12, concurrent, filter)
}

// 全量同步 ShareAgentIncomeRecordTackLog
func SyncFullShareAgentIncomeRecordTackLog(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	// 按 ctime 数据分片
	begin, _ := time.ParseInLocation(utils.FORMAT, "2025-06-10 00:00:00", location)
	end, _ := time.ParseInLocation(utils.FORMAT, "2025-06-12 00:00:00", location)

	var onceTime, concurrent int = 1, 2
	var newRecords = func() (r []*data.ShareAgentIncomeRecord) { return }
	var timeParse = func(t time.Time) any { return t.UnixMilli() }
	filter := bson.M{}
	Syncmongo2ClickhouseByTime("ShareAgentIncomeRecordTackLog", data.ShareAgentIncomeRecordTackLogs, newRecords, "ctime", begin, end, timeParse, onceTime, time.Hour*12, concurrent, filter)
}

// 全量同步 VolatilitySubsidy
func SyncFullVolatilitySubsidy(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	// 按 ctime 数据分片
	begin, _ := time.ParseInLocation(utils.FORMAT, "2025-06-10 00:00:00", location)
	end, _ := time.ParseInLocation(utils.FORMAT, "2025-06-12 00:00:00", location)

	var onceTime, concurrent int = 1, 2
	var newRecords = func() (r []*data.VolatilitySubsidy) { return }
	var timeParse = func(t time.Time) any { return t.UnixMilli() }
	filter := bson.M{}
	Syncmongo2ClickhouseByTime("VolatilitySubsidy", data.VolatilitySubsidys, newRecords, "ctime", begin, end, timeParse, onceTime, time.Hour*12, concurrent, filter)
}

// 全量同步 OnlineUsersLog
func SyncFullOnlineUsersLog(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	// 按 ctime 数据分片
	begin, _ := time.ParseInLocation(utils.FORMAT, "2025-06-10 00:00:00", location)
	end, _ := time.ParseInLocation(utils.FORMAT, "2025-06-12 00:00:00", location)

	var onceTime, concurrent int = 1, 2
	var newRecords = func() (r []*data.OnlineUsersLog) { return }
	var timeParse = func(t time.Time) any { return t.Unix() }
	filter := bson.M{}
	Syncmongo2ClickhouseByTime("OnlineUsersLog", data.OnlineUsersLogs, newRecords, "ctime", begin, end, timeParse, onceTime, time.Hour*12, concurrent, filter)
}

// 全量同步 UserCoupon
func SyncFullUserCoupon(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	// 按 ctime 数据分片
	begin, _ := time.ParseInLocation(utils.FORMAT, "2025-06-10 00:00:00", location)
	end, _ := time.ParseInLocation(utils.FORMAT, "2025-06-12 00:00:00", location)

	var onceTime, concurrent int = 1, 2
	var newRecords = func() (r []*data.UserCoupon) { return }
	var timeParse = func(t time.Time) any { return t.Unix() }
	filter := bson.M{}
	Syncmongo2ClickhouseByTime("UserCoupon", data.UserCoupons, newRecords, "ctime", begin, end, timeParse, onceTime, time.Hour*12, concurrent, filter)
}
