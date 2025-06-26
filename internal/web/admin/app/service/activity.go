package service

import (
	"errors"
	"fmt"
	"goserver/internal/web/admin/app/args"
	"goserver/internal/web/admin/app/entity"
	"goserver/pkg/data/ck"
	"goserver/pkg/utils"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/globalsign/mgo/bson"
)

// 活动管理
type activityService struct{}

// 基础付费活动列表
func (this *activityService) GetBasicPayList(page, pageSize int, m bson.M) ([]entity.BasicPayActivity, error) {
	var list []entity.BasicPayActivity
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "date", false)
	err := BasicPays.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList(list)
	return list, err
}

// 转换为分展示
func (this *activityService) chipList(list []entity.BasicPayActivity) []entity.BasicPayActivity {
	for k, v := range list {
		c, _ := ConvertToIndiaTime(v.Date)
		v.SDate = c
		list[k] = v
	}
	return list
}

// 查询基础付费活动条数
func (this *activityService) GetBasicPayTotal(m bson.M) (int64, error) {
	return int64(Count(BasicPays, m)), nil
}

// 新增基础付费活动
func (this *activityService) AddBasicPay(res *entity.BasicPayActivity) error {
	info := new(entity.BasicPayActivity)
	GetByQ(BasicPays, bson.M{"date": res.Date}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"shop_number":  res.ShopNumber,
			"shop_diamond": res.ShopDiamond,
			"shop_coin":    res.ShopCoin,
			"rm_number":    res.RmNumber,
			"rm_diamond":   res.RmDiamond,
			"rm_coin":      res.RmCoin,
			"jy_number":    res.JyNumber,
			"jy_diamond":   res.JyDiamond,
			"jy_coin":      res.JyCoin,
			"sc_number":    res.ScNumber,
			"sc_diamond":   res.ScDiamond,
			"sc_coin":      res.ScCoin,
		}
		if Update(BasicPays, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + res.Id)
	} else {
		// 新增
		res.Id = bson.NewObjectId().Hex()
		if !Insert(BasicPays, res) {
			return errors.New("写入失败:" + res.Id)
		}
		return nil
	}
}

// 查询基础免费活动列表
func (this *activityService) GetBasicFreeList(page, pageSize int, m bson.M) ([]entity.BasicFreeActivity, error) {
	var list []entity.BasicFreeActivity
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "date", false)
	err := BasicFrees.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList1(list)
	return list, err
}

// 转换为分展示
func (this *activityService) chipList1(list []entity.BasicFreeActivity) []entity.BasicFreeActivity {
	for k, v := range list {
		c, _ := ConvertToIndiaTime(v.Date)
		v.SDate = c
		list[k] = v
	}
	return list
}

// 查询基础免费活动条数
func (this *activityService) GetBasicFreeTotal(m bson.M) (int64, error) {
	return int64(Count(BasicFrees, m)), nil
}

// 新增基础免费活动
func (this *activityService) AddBasicFree(res *entity.BasicFreeActivity) error {
	info := new(entity.BasicFreeActivity)
	GetByQ(BasicFrees, bson.M{"date": res.Date}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"sign_in_number":       res.SignInNumber,
			"sign_in_diamond":      res.SignInDiamond,
			"sign_in_coin":         res.SignInCoin,
			"online_number":        res.OnlineNumber,
			"online_diamond":       res.OnlineDiamond,
			"online_coin":          res.OnlineCoin,
			"task_number":          res.TaskNumber,
			"task_diamond":         res.TaskDiamond,
			"task_coin":            res.TaskCoin,
			"trail_or_set_number":  res.TrailOrSetNumber,
			"trail_or_set_diamond": res.TrailOrSetDiamond,
			"trail_or_set_coin":    res.TrailOrSetCoin,
		}
		if Update(BasicFrees, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + res.Id)
	} else {
		// 新增
		res.Id = bson.NewObjectId().Hex()
		if !Insert(BasicFrees, res) {
			return errors.New("写入失败:" + res.Id)
		}
		return nil
	}
}

// 查询分享活动数据列表
func (this *activityService) GetShareDataList(page, pageSize int, m bson.M) ([]entity.ShareActivityData, error) {
	var list []entity.ShareActivityData
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	err := ShareDatas.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList5(list)
	return list, err
}

// 转换为分展示
func (this *activityService) chipList5(list []entity.ShareActivityData) []entity.ShareActivityData {
	for k, v := range list {
		// v.FWaterOutPut = Chip2Float(int64(v.WaterOutPut))
		// v.FFirstOrderOutPut = Chip2Float(int64(v.FirstOrderOutPut))
		// v.FGrossOutput = Chip2Float(int64(v.GrossOutput))
		logtime, _ := ConvertToIndiaTime(v.Ctime)
		v.Rtime = logtime
		list[k] = v
	}
	return list
}

// 查询分享活动数据列表条数
func (this *activityService) GetShareDataTotal(m bson.M) (int64, error) {
	return int64(Count(ShareDatas, m)), nil
}

// 添加分享活动数据
func (this *activityService) AddShareData(rate *entity.ShareActivityData) error {
	info := new(entity.ShareActivityData)
	GetByQ(ShareDatas, bson.M{"ctime": rate.Ctime}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"register_number":     rate.RegisterNumber,
			"water_out_put":       rate.WaterOutPut,
			"first_order_out_put": rate.FirstOrderOutPut,
			"gross_output":        rate.GrossOutput,
		}
		if Update(ShareDatas, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + rate.Id)
	} else {
		// 新增
		rate.Id = bson.NewObjectId().Hex()
		if !Insert(ShareDatas, rate) {
			return errors.New("写入失败:" + rate.Id)
		}
		return nil
	}
}

// 查询分享活动领取记录列表
func (this *activityService) GetShareRecordList(page, pageSize int, m bson.M) ([]entity.LogShareWithDraw, error) {
	var list []entity.LogShareWithDraw
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	err := LogShareWaters.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList4(list)
	return list, err
}

// 转换为分展示
func (this *activityService) chipList4(list []entity.LogShareWithDraw) []entity.LogShareWithDraw {
	for k, v := range list {
		switch v.Gtype {
		case 1:
			v.GtypeName = "流水奖励"
		case 2:
			v.GtypeName = "首单奖励"
		}
		v.FCash = Chip2Float(int64(v.Cash))
		v.FCoin = Chip2Float(int64(v.Coin))
		logtime, _ := ConvertToIndiaTime(v.Ctime)
		v.Rtime = logtime
		list[k] = v
	}
	return list
}

// 查询分享活动领取记录列表条数
func (this *activityService) GetShareRecordTotal(m bson.M) (int64, error) {
	return int64(Count(LogShareWaters, m)), nil
}

// 查询分享排名列表
func (this *activityService) GetShareRankingList(page, pageSize int, m bson.M) ([]entity.ShareRanking, error) {
	var list []entity.ShareRanking
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, _ := parsePageAndSort(page, pageSize, "_id", false)
	pipeline := []bson.M{
		{"$match": m},
		{"$project": bson.M{
			"_id":           1,
			"nickname":      1,
			"ad__bundle_id": 1,
			"ctime":         1,
			"login_time":    1,
			"share_below":   1,
			"ShareCount":    bson.M{"$size": bson.M{"$ifNull": []interface{}{bson.M{"$objectToArray": "$share_below"}, []interface{}{}}}},
		}},
		{"$sort": bson.M{"ShareCount": -1}},
		{"$addFields": bson.M{
			"ad__bundle_id": "$ad__bundle_id",
			"ctime":         "$ctime",
			"login_time":    "$login_time",
			"share_below":   "$share_below",
			"ShareCount":    "$ShareCount",
		}},
		{"$skip": skipNum},
		{"$limit": pageSize},
	}
	err := PlayerUsers.Pipe(pipeline).All(&list)
	list = this.chipList3(list)
	return list, err
}

// 转换为分展示
func (this *activityService) chipList3(list []entity.ShareRanking) []entity.ShareRanking {
	for k, v := range list {
		v.ShareCount = len(v.ShareBelow)
		// v.FRecharge = Chip2Float(int64(v.Recharge))
		// v.FFirstRecharge = Chip2Float(int64(v.FirstRecharge))
		// // s := int64(v.FirstReward) / int64(100)
		// v.FFirstReward = Chip2Float(int64(v.FirstReward)) / Chip2Float(int64(10000))
		// v.FSecondReward = Chip2Float(int64(v.SecondReward)) / Chip2Float(int64(10000))
		list[k] = v
	}
	return list
}

// 查询分享排名列表条数
func (this *activityService) GetShareRankingTotal(m bson.M) (int64, error) {
	return int64(Count(PlayerUsers, m)), nil
}

// 查询分享活动配置列表
func (this *activityService) GetShareConfigList(page, pageSize int, m bson.M) ([]entity.Share, error) {
	var list []entity.Share
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "id", false)
	err := Shares.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList2(list)
	return list, err
}

// 转换为分展示
func (this *activityService) chipList2(list []entity.Share) []entity.Share {
	for k, v := range list {
		v.FRecharge = Chip2Float(int64(v.Recharge))
		v.FFirstRecharge = Chip2Float(int64(v.FirstRecharge))
		// s := int64(v.FirstReward) / int64(100)
		v.FFirstReward = Chip2Float(int64(v.FirstReward)) / Chip2Float(int64(10000))
		v.FSecondReward = Chip2Float(int64(v.SecondReward)) / Chip2Float(int64(10000))
		list[k] = v
	}
	return list
}

// 查询分享活动配置列表条数
func (this *activityService) GetShareConfigTotal(m bson.M) (int64, error) {
	return int64(Count(Shares, m)), nil
}

/*
	小米活动相关接口
*/
// 查询小米活动列表
func (this *activityService) GetMiuiList(page, pageSize int, m bson.M) ([]entity.LuckyDrawGive, error) {
	var list []entity.LuckyDrawGive
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "round", false)
	err := LuckyDrawGives.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList6(list)
	return list, err
}

// 转换为分展示
func (this *activityService) chipList6(list []entity.LuckyDrawGive) []entity.LuckyDrawGive {
	for k, v := range list {
		v.FReword = Chip2Float(int64(v.Reword))
		c, _ := ConvertToIndiaTime(v.Ctime.Unix())
		v.Ctime = c
		list[k] = v
	}
	return list
}

// 小米活动条数
func (this *activityService) GetMiuiTotal(m bson.M) (int64, error) {
	return int64(Count(LuckyDrawGives, m)), nil
}

/*
	彩票活动相关接口
*/
// 查询彩票活动列表
func (this *activityService) GetLotteryList(page, pageSize int, m bson.M) ([]entity.LotteryActivity, error) {
	var list []entity.LotteryActivity
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "date", false)
	err := LotteryActivitys.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList7(list)
	return list, err
}

// 转换为分展示
func (this *activityService) chipList7(list []entity.LotteryActivity) []entity.LotteryActivity {
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		v.FFirstPrize = Chip2Float(v.FirstPrize)
		v.FSecondPrize = Chip2Float(v.SecondPrize)
		v.FThirdPrize = Chip2Float(v.ThirdPrize)
		v.FFourthPrize = Chip2Float(v.FourthPrize)
		v.FFifthPrize = Chip2Float(v.FifthPrize)
		v.FExpendBouns = Chip2Float(v.ExpendBouns)

		list[k] = v
	}
	return list
}

// 彩票活动条数
func (this *activityService) GetLotteryTotal(m bson.M) (int64, error) {
	return int64(Count(LotteryActivitys, m)), nil
}

// Sharedraw 玩游戏分享抽大奖活动
func (s *activityService) Sharedraw(m bson.M, startDate, endDate string) (r []*entity.PddStat, err error) {
	// var r1 []*entity.PddStat
	err = PddStats.Find(m).Sort("-date").All(&r)
	if err != nil {
		return
	}
	// 日期范围内总汇
	r = sharedDrawTotal(r, startDate, endDate)
	// 计算结果
	sharedrawMapping(r)
	return
}

func sharedDrawTotal(stats []*entity.PddStat, begin, end string) (total []*entity.PddStat) {
	t := &entity.PddStat{}
	t.SDate = fmt.Sprintf("%s~%s总汇", begin, end)
	for _, stat := range stats {
		t.LoginedUsers += stat.LoginedUsers
		t.PddUsers += stat.PddUsers
		t.BeInvoteUsers += stat.BeInvoteUsers
		t.BeInvoteLoginedUsers += stat.BeInvoteLoginedUsers
		t.DrawUsers += stat.DrawUsers
		t.BeInvoteUserDraws += stat.BeInvoteUserDraws
		t.UserDraws += stat.UserDraws
		t.AwardUsers += stat.AwardUsers
		t.AwardJackpot += stat.AwardJackpot
		t.BeInvoteRegUsers += stat.BeInvoteRegUsers
		t.BeInvoteDevices += stat.BeInvoteDevices
		t.Share2Users += stat.Share2Users
		t.Share2DrawUsers += stat.Share2DrawUsers
		t.Share2Devices += stat.Share2Devices
		t.PlayedReg24 += stat.PlayedReg24
		t.RegUserChages += stat.RegUserChages
		t.RegUserChargeAmount += stat.RegUserChargeAmount
		t.RegUserWithdrawAmount += stat.RegUserWithdrawAmount
		t.ChargeAmount += stat.ChargeAmount
		t.WithdrawAmount += stat.WithdrawAmount
		t.BeInvoteChargeAmount += stat.BeInvoteChargeAmount
		t.BeInvoteWithdrawAmount += stat.BeInvoteWithdrawAmount
	}
	total = append([]*entity.PddStat{t}, stats...)
	return
}

func sharedrawMapping(stats []*entity.PddStat) {
	for _, v := range stats {
		if v.SDate == "" {
			v.SDate = time.Unix(v.Date, 0).Format("2006-01-02")
		}
		if v.LoginedUsers > 0 {
			v.PddRate = fmt.Sprintf("%.2f", float64(v.PddUsers)/float64(v.LoginedUsers))
		}
		// 分享用户参与率
		if v.BeInvoteLoginedUsers > 0 {
			v.BeInvoteRate = fmt.Sprintf("%.2f", float64(v.BeInvoteUsers)/float64(v.BeInvoteLoginedUsers))
		}
		// 总人均抽奖次数
		if v.PddUsers > 0 {
			v.ShareDrawRate = fmt.Sprintf("%.2f", float64(v.UserDraws)/float64(v.PddUsers))
		}
		// 分享用户人均抽奖次数
		if v.BeInvoteUsers > 0 {
			v.ShareDrawAvg = fmt.Sprintf("%.2f", float64(v.BeInvoteUserDraws)/float64(v.BeInvoteUsers))
		}
		v.SAwardJackpot = fmt.Sprintf("%.2f", float64(v.AwardJackpot)/100)
		// 总有效分享率
		if v.BeInvoteRegUsers > 0 {
			v.YxShareRate = fmt.Sprintf("%.2f", float64(v.BeInvoteDevices)/float64(v.BeInvoteRegUsers))
		}
		// 分享用户有效分享率 分享用户有效分享数/分享用户分享来人数
		if v.Share2Users > 0 {
			v.YxShare2Rate = fmt.Sprintf("%.2f", float64(v.Share2DrawUsers)/float64(v.Share2Users))
		}
		// 总CPR
		if v.BeInvoteRegUsers > 0 {
			v.ZCrp = fmt.Sprintf("%.4f", float64(v.AwardJackpot)*0.012/100/float64(v.BeInvoteRegUsers))
		}
		// 有效CPR
		if v.AwardJackpot > 0 {
			v.YXCrp = fmt.Sprintf("%.2f", float64(v.BeInvoteUsers)/(float64(v.AwardJackpot)*0.012/100))
		}
		// SCCpp
		if v.RegUserChages > 0 {
			v.SCCpp = fmt.Sprintf("%.2f", float64(v.AwardJackpot)*0.012/100/float64(v.RegUserChages))
		}
		// 活动利润
		v.Profit = fmt.Sprintf("%.2f", float64(v.BeInvoteChargeAmount-v.BeInvoteWithdrawAmount-v.AwardJackpot)/100.0)

		// D0净ROI 新用户的（总充-总提）/获奖金额
		if v.AwardJackpot > 0 {
			v.Roi = fmt.Sprintf("%.2f", float64(v.RegUserChargeAmount-v.WithdrawAmount)/float64(v.AwardJackpot))
		}

		if v.AwardJackpot > 0 {
			// 所有注册用户的（总充-总提）/获奖金额
			v.LjRoi = fmt.Sprintf("%.2f", float64(v.ChargeAmount-v.WithdrawAmount)/(float64(v.AwardJackpot)))
		}
		// D0ROI=分享来的新用户充值金额/获奖金额，因为是看比值，不用换算成美金。
		if v.AwardJackpot > 0 {
			v.DoRoi = fmt.Sprintf("%.2f", float64(v.RegUserChargeAmount)/float64(v.AwardJackpot))
		}
		// 累计毛ROI=总充值金额/获奖金额
		if v.AwardJackpot > 0 {
			v.LjMRoi = fmt.Sprintf("%.2f", float64(v.ChargeAmount)/float64(v.AwardJackpot))
		}
		// 日活占比 分享用户日活数/所有渠道的日活数
		if v.LoginedUsers > 0 {
			v.LoginedRate = fmt.Sprintf("%.2f", float64(v.BeInvoteLoginedUsers)/float64(v.LoginedUsers))
		}
		// 活动总充额占比 分享用户总充值金额/所有渠道总充值金额
		if v.ChargeAmount > 0 {
			v.PddChargeRate = fmt.Sprintf("%.2f", float64(v.BeInvoteChargeAmount)/float64(v.ChargeAmount))
		}
	}
}

// BetRank 排行榜活动
func (s *activityService) BetRank(page, pageSize int, rankType int32, begin, end int64) (total int64, stats []*entity.BetRankStat, err error) {
	sql0 := `
		SELECT s0.rank_type rank_type, s0.etime etime, s0.jackpot, s1.user_count, s1.prize_sum, s1.prize_avg, 
			s2.rank rank_high, s2.userid userid_high, s2.prize prize_high, s2.prize_type,
		FROM (
			SELECT rank_type, etime, MAX(jackpot) jackpot
			FROM game.col_activity_betrank_prize FINAL
			WHERE 1 = 1 %s
			GROUP BY rank_type, etime
		) s0
		LEFT JOIN (
			SELECT rank_type, etime, COUNT(*) user_count, SUM(prize) prize_sum, AVG(prize) prize_avg
			FROM game.col_activity_betrank_prize FINAL
			WHERE robot = 0 %s
			GROUP BY rank_type, etime
		) s1 ON s0.rank_type = s1.rank_type AND s0.etime = s1.etime
		LEFT  JOIN (
			SELECT rank_type, etime, rank, userid, prize, prize_type
			FROM game.col_activity_betrank_prize FINAL
			WHERE (rank_type, etime, rank) IN (
				SELECT rank_type, etime, MIN(rank) rank_min 
				FROM game.col_activity_betrank_prize FINAL
				WHERE robot = 0 %s
				GROUP BY rank_type, etime
			)
		) s2 ON s0.rank_type = s2.rank_type AND s0.etime = s2.etime
		ORDER BY s0.etime DESC, s0.rank_type LIMIT ?, ?
	`
	sql_total := `
		SELECT COUNT(*) total
		FROM (
			SELECT rank_type, etime
			FROM game.col_activity_betrank_prize FINAL
			WHERE 1 = 1 %s
			GROUP BY rank_type, etime
		) s0
	`
	var args []any
	var cond string
	if rankType != 0 {
		cond += " AND rank_type = ?"
		args = append(args, rankType)
	}

	var ts, te string
	if begin > 0 {
		cond += " AND etime >= ?"
		args = append(args, begin-1)
		ts = time.Unix(begin, 0).In(location).Format(utils.FORMAT_DATE)
	}
	if end > 0 {
		cond += " AND etime < ?"
		args = append(args, end)
		te = time.Unix(end, 0).In(location).Format(utils.FORMAT_DATE)
	}

	sql_total = fmt.Sprintf(sql_total, cond)
	args_total := args
	if err = ck.Select(&total, sql_total, args_total...); err != nil {
		return
	}
	if total == 0 {
		return
	}

	offset, limit := PageCalc(page, pageSize)
	sql0 = fmt.Sprintf(sql0, cond, cond, cond)
	args0 := append(args, args...)
	args0 = append(args0, args...)
	args0 = append(args0, offset, limit)
	var datas []map[string]any
	if err = ck.Select(&datas, sql0, args0...); err != nil {
		return
	}
	for _, data := range datas {
		rank_type := utils.ToInt64(data["rank_type"])
		etime := utils.ToInt64(data["etime"])
		user_count := utils.ToInt64(data["user_count"])
		prize_sum := utils.ToInt64(data["prize_sum"])
		prize_avg := utils.ToFloat64(data["prize_avg"])
		rank_high := utils.ToInt64(data["rank_high"])
		userid_high := data["userid_high"].(string)
		prize_high := utils.ToInt64(data["prize_high"])
		prize_type := utils.ToInt64(data["prize_type"])
		jackpot := utils.ToInt64(data["jackpot"])

		etime2 := time.Unix(etime+1, 0).In(location)
		rankType, _ := utils.CaseWhen3(rank_type, 1, "日榜", 2, "周榜", 3, "月榜")
		prizeType, _ := utils.CaseWhen3(prize_type, 1, "bonus", 2, "cash", 3, "withdrawable")

		stat := &entity.BetRankStat{
			SDate:        etime2.Format(utils.FORMAT_DATE2),
			Week:         utils.WeekdayZh(etime2.Weekday()),
			RankType:     rankType,
			PrizeUsers:   user_count,
			Jackpot:      fmt.Sprintf("%.2f", Chip2Float(jackpot)),
			PrizeSum:     fmt.Sprintf("%.2f", Chip2Float(prize_sum)),
			PrizeAvg:     fmt.Sprintf("%.2f", Chip2Float(prize_avg)),
			PrizeType:    prizeType,
			UserMaxRank:  fmt.Sprint(rank_high + 1),
			MaxUserid:    userid_high,
			MaxUserPrize: fmt.Sprintf("%.2f", Chip2Float(prize_high)),
		}
		if userid_high == "" {
			stat.UserMaxRank = ""
			stat.MaxUserPrize = ""
		}
		stats = append(stats, stat)
	}

	// 汇总统计
	summary, err := s.betRankSummary(cond, args)
	if err != nil {
		return
	}
	summary.SDate = fmt.Sprintf("%s~%s汇总", ts, te)
	stats = append([]*entity.BetRankStat{summary}, stats...)
	return
}

func (s *activityService) betRankSummary(cond string, args []any) (summary *entity.BetRankStat, err error) {
	summary = &entity.BetRankStat{
		SDate:        "",
		Week:         "-",
		RankType:     "-",
		Jackpot:      "-",
		PrizeSum:     "-",
		PrizeAvg:     "-",
		PrizeType:    "-",
		UserMaxRank:  "-",
		MaxUserid:    "-",
		MaxUserPrize: "-",
	}

	sql_summary := `
		SELECT COUNT(*) user_count, SUM(prize) prize_sum, AVG(prize) prize_avg
		FROM game.col_activity_betrank_prize FINAL
		WHERE robot = 0 %s
	`
	sql_summary_jackpot := `
		SELECT SUM(jackpot) jackpot_sum FROM (
			SELECT rank_type, etime, MAX(jackpot) jackpot 
			FROM game.col_activity_betrank_prize FINAL
			WHERE 1 = 1 %s
			GROUP BY rank_type, etime
		) s1
	`

	sql_summary = fmt.Sprintf(sql_summary, cond)
	var data = make(map[string]any)
	if err = ck.Select(&data, sql_summary, args...); err != nil {
		return
	}
	summary.PrizeUsers = utils.ToInt64(data["user_count"])
	summary.PrizeSum = fmt.Sprintf("%.2f", Chip2Float(utils.ToInt64(data["prize_sum"])))
	summary.PrizeAvg = fmt.Sprintf("%.2f", Chip2Float(utils.ToFloat64(data["prize_avg"])))

	sql_summary_jackpot = fmt.Sprintf(sql_summary_jackpot, cond)
	var jackpot_sum int64
	if err = ck.Select(&jackpot_sum, sql_summary_jackpot, args...); err != nil {
		return
	}
	summary.Jackpot = fmt.Sprintf("%.2f", Chip2Float(jackpot_sum))

	return
}

type BetRankUsersArgs struct {
	Page     int
	PageSize int
	Userid   string
	RankType int32
	Begin    int64
	End      int64
	Begin2   int64
	End2     int64
	RankSort int32 // 0.asc,1.desc
}

// BetRankUsers 排行榜活动明细
func (s *activityService) BetRankUsers(arg BetRankUsersArgs) (total int64, stats []*entity.BetRankUserStat, err error) {
	sql := `
		SELECT rank_type, stime, etime, rank, userid, prize, prize_type, jackpot, received, receive_time, bets
		FROM game.col_activity_betrank_prize FINAL
		WHERE robot = 0 %s
		ORDER BY %s
		LIMIT ?, ?
	`
	sql_total := `
		SELECT count(*)
		FROM game.col_activity_betrank_prize FINAL
		WHERE robot = 0 %s
	`
	var args []any
	var cond string
	var orderBy string
	var ts, te string

	if arg.Userid != "" {
		cond += " AND userid = ?"
		args = append(args, arg.Userid)
	}
	if arg.RankType != 0 {
		cond += " AND rank_type = ?"
		args = append(args, arg.RankType)
	}
	if arg.Begin != 0 {
		cond += " AND etime >= ?"
		args = append(args, arg.Begin-1)
	}
	if arg.End != 0 {
		cond += " AND etime < ?"
		args = append(args, arg.End)
	}
	if arg.Begin2 != 0 {
		cond += " AND stime >= ?"
		args = append(args, arg.Begin2)
		ts = time.Unix(arg.Begin2, 0).In(location).Format(utils.FORMAT_DATE)
	}
	if arg.End2 != 0 {
		cond += " AND stime <= ?"
		args = append(args, arg.End2)
		te = time.Unix(arg.End2, 0).In(location).Format(utils.FORMAT_DATE)
	}

	if arg.RankSort == 1 {
		orderBy = "rank ASC"
	} else if arg.RankSort == 2 {
		orderBy = "rank DESC"
	} else {
		orderBy = "etime DESC, rank DESC"
	}

	sql_total = fmt.Sprintf(sql_total, cond)
	args_total := args
	if err = ck.Select(&total, sql_total, args_total...); err != nil {
		return
	}
	if total == 0 {
		return
	}

	sql0 := fmt.Sprintf(sql, cond, orderBy)
	offset, limit := PageCalc(arg.Page, arg.PageSize)
	args0 := append(args, offset, limit)
	var datas []map[string]any
	if err = ck.Select(&datas, sql0, args0...); err != nil {
		return
	}
	if len(datas) == 0 {
		return
	}

	var dateUserids = make(map[[2]int64][]string) // 发放日用户列表
	var userStats = make(map[string]*entity.BetRankUserStat)
	for _, data := range datas {
		// rank_type, stime, etime, rank, userid, prize, prize_type, jackpot, received, receive_time
		rank_type := utils.ToInt64(data["rank_type"])
		stime := utils.ToInt64(data["stime"])
		etime := utils.ToInt64(data["etime"])
		userid := data["userid"].(string)
		rank := utils.ToInt64(data["rank"])
		prize := utils.ToInt64(data["prize"])
		prize_type := utils.ToInt64(data["prize_type"])
		received := utils.ToInt64(data["received"])
		receive_time := utils.ToInt64(data["receive_time"])
		bets := utils.ToInt64(data["bets"])

		stime1 := time.Unix(stime, 0).In(location)
		etime1 := time.Unix(etime, 0).In(location)
		etime2 := time.Unix(etime+1, 0).In(location)
		rankType, _ := utils.CaseWhen3(rank_type, 1, "日榜", 2, "周榜", 3, "月榜")
		prizeType, _ := utils.CaseWhen3(prize_type, 1, "bonus", 2, "cash", 3, "withdrawable")

		// 榜单期打码
		// rankDate2 := [2]int64{stime, etime}
		// rankDateUserids[rankDate2] = append(rankDateUserids[rankDate2], userid)
		// 发放日打码
		sdate2 := [2]int64{etime + 1, etime + (24 * 60 * 60)}
		dateUserids[sdate2] = append(dateUserids[sdate2], userid)

		stat := &entity.BetRankUserStat{
			RankDate:        fmt.Sprintf("%s~%s", stime1.Format(utils.FORMAT_DATE2), etime1.Format(utils.FORMAT_DATE2)),
			SDate:           etime2.Format(utils.FORMAT_DATE2),
			Userid:          userid,
			RankType:        rankType,
			Etime:           etime1.Format(utils.FORMAT_TIME),
			Stime:           etime2.Format(utils.FORMAT_TIME),
			Pays:            "",
			Withdraws:       "",
			Contribute:      "",
			StimePays:       "",
			StimeWithdraws:  "",
			StimeContribute: "",
			RankBets:        fmt.Sprintf("%.2f", Chip2Float(bets)),
			Rank:            fmt.Sprint(rank + 1),
			Prize:           fmt.Sprintf("%.2f", Chip2Float(prize)),
			PrizeType:       prizeType,
			Phone:           "",
			SDate2:          sdate2,
		}
		if received == 1 {
			rtime := time.Unix(receive_time, 0).In(location)
			stat.ReceiveTime = rtime.Format(utils.FORMAT2)
			stat.ReceiveHourGap = fmt.Sprint((rtime.Unix() - etime2.Unix()) / 60 / 60)
		}
		stats = append(stats, stat)
		userStats[userid] = stat
	}

	// 玩家打码量查询
	// var datesUserBets = make(map[[2]int64]map[string]int64)
	for dates, userids := range dateUserids {
		// 发放日打码量查询
		sql_bets := `
			SELECT userid, SUM(bets) bets FROM (
				SELECT userid, SUM(bet_amount) bets FROM game.col_detail FINAL
				WHERE robot = 0 AND userid IN ? AND end_time BETWEEN ? AND ?
				GROUP BY userid 
				
				UNION ALL
				
				SELECT s1.user_id userid, SUM(s1.amount_sum) bets
				FROM (
					SELECT round_id, user_id, game_id, SUM(amount) amount_sum, MIN(ctime) begin_time 
					FROM game.col_nsq_log_external_bet FINAL 
					WHERE amount != 0 AND user_id IN ? AND ctime BETWEEN ? AND ?
					GROUP BY round_id, user_id, game_id
				) s1
				GROUP BY s1.user_id
			) t1
			GROUP BY userid
		`
		args_bets := []any{userids, dates[0], dates[1], userids, dates[0], dates[1]}

		var bets_datas []map[string]any
		if err = ck.Select(&bets_datas, sql_bets, args_bets...); err != nil {
			return
		}
		for _, data := range bets_datas {
			userid := data["userid"].(string)
			bets := utils.ToInt64(data["bets"])

			if stat, ok := userStats[userid]; ok {
				stat.StimeBets = fmt.Sprintf("%.2f", Chip2Float(bets))
			}
		}

		// 总充值 总提现 玩家总贡献 当日充值 当日提现 当日总贡献 玩家手机号
		sql_trade := `
			SELECT s0.userid userid, s0.phone, s1.pays, s2.day_pays, s3.withdraws, s4.day_withdraws FROM (
				SELECT userid, phone FROM game.col_user FINAL WHERE userid IN ?
			) s0
			LEFT JOIN (
				SELECT userid, SUM(amount) pays
				FROM game.col_trade_record FINAL
				WHERE order_status = 4 AND userid IN ?
				GROUP BY userid
			) s1 ON s0.userid = s1.userid
			LEFT JOIN (
				SELECT userid, SUM(amount) day_pays
				FROM game.col_trade_record FINAL
				WHERE order_status = 4 AND userid IN ? AND ctime BETWEEN ? AND ?
				GROUP BY userid
			) s2 ON s0.userid = s2.userid
			LEFT JOIN (
				SELECT userid, SUM(amount + commission) withdraws
				FROM game.col_withdraw_record FINAL
				WHERE order_status = 2 AND userid IN ?
				GROUP BY userid
			) s3 ON s0.userid = s3.userid
			LEFT JOIN (
				SELECT userid, SUM(amount + commission) day_withdraws
				FROM game.col_withdraw_record FINAL
				WHERE order_status = 2 AND userid IN ? AND ctime BETWEEN ? AND ?
				GROUP BY userid
			) s4 ON s0.userid = s4.userid
		`
		stime, etime := time.Unix(dates[0], 0).In(location), time.Unix(dates[1], 0).In(location)
		args_trade := []any{userids, userids, userids, stime, etime, userids, userids, stime, etime}

		var trade_datas []map[string]any
		if err = ck.Select(&trade_datas, sql_trade, args_trade...); err != nil {
			return
		}
		for _, data := range trade_datas {
			userid := data["userid"].(string)
			pays := utils.ToInt64(data["pays"])
			day_pays := utils.ToInt64(data["day_pays"])
			withdraws := utils.ToInt64(data["withdraws"])
			day_withdraws := utils.ToInt64(data["day_withdraws"])

			if stat, ok := userStats[userid]; ok {
				stat.Pays = fmt.Sprintf("%.2f", Chip2Float(pays))
				stat.StimePays = fmt.Sprintf("%.2f", Chip2Float(day_pays))
				stat.Withdraws = fmt.Sprintf("%.2f", Chip2Float(withdraws))
				stat.StimeWithdraws = fmt.Sprintf("%.2f", Chip2Float(day_withdraws))
				stat.Contribute = fmt.Sprintf("%.2f", Chip2Float(pays-withdraws))
				stat.StimeContribute = fmt.Sprintf("%.2f", Chip2Float(day_pays-day_withdraws))
				if phone, ok := data["phone"].(string); ok {
					if len(phone) <= 6 {
						if len(phone) > 1 {
							stat.Phone = phone[0:1] + "***"
						} else {
							stat.Phone = phone
						}
					} else {
						stat.Phone = phone[0:3] + "****" + phone[len(phone)-3:]
					}
				}
			}
		}
	}

	// 汇总查询
	summary, err := s.BetRankUsersSummary(cond, args)
	if err != nil {
		return
	}
	if ts != "" && te != "" {
		summary.RankDate = fmt.Sprintf("%s~%s汇总", ts, te)
	}
	stats = append([]*entity.BetRankUserStat{summary}, stats...)
	return
}

// BetRankUsersSummary 排行榜活动明细汇总
func (s *activityService) BetRankUsersSummary(cond string, args []any) (summary *entity.BetRankUserStat, err error) {
	summary = &entity.BetRankUserStat{
		RankDate: "汇总",
		SDate:    "-", Userid: "-", RankType: "-", Etime: "-", Stime: "-", ReceiveTime: "-", ReceiveHourGap: "-", Pays: "-", Withdraws: "-", Contribute: "-", StimePays: "-", StimeWithdraws: "-", StimeContribute: "-", RankBets: "-", StimeBets: "-", Rank: "-", Prize: "-", PrizeType: "-", Phone: "-",
	}
	// 总充值	总提现	玩家总贡献	当日充值	当日提现	当日总贡献	榜单期打码	发放日打码 分到金额
	sql_pays := `
		SELECT s1.pays, s2.withdraws FROM (
			SELECT SUM(amount) pays
			FROM game.col_trade_record FINAL
			WHERE order_status = 4 AND userid IN (
				SELECT DISTINCT userid
				FROM game.col_activity_betrank_prize FINAL
				WHERE robot = 0 %s
			)
		) s1 JOIN (
			SELECT SUM(amount + commission) withdraws
			FROM game.col_withdraw_record FINAL
			WHERE order_status = 2 AND userid IN (
				SELECT DISTINCT userid
				FROM game.col_activity_betrank_prize FINAL
				WHERE robot = 0 %s
			)
		) s2 ON 1 = 1
	`
	sql_pays = fmt.Sprintf(sql_pays, cond, cond)
	args_pays := append(args, args...)
	var data_pays = make(map[string]any)
	if err = ck.Select(&data_pays, sql_pays, args_pays...); err != nil {
		return
	}
	pays := utils.ToInt64(data_pays["pays"])
	withdraws := utils.ToInt64(data_pays["withdraws"])
	summary.Pays = fmt.Sprintf("%.2f", Chip2Float(pays))
	summary.Withdraws = fmt.Sprintf("%.2f", Chip2Float(withdraws))
	summary.Contribute = fmt.Sprintf("%.2f", Chip2Float(pays-withdraws))

	sql_bets := `
		SELECT SUM(bets) bets, SUM(prize) prize
		FROM game.col_activity_betrank_prize FINAL
		WHERE robot = 0 %s
	`
	sql_bets = fmt.Sprintf(sql_bets, cond)
	var data_bets = make(map[string]any)
	if err = ck.Select(&data_bets, sql_bets, args...); err != nil {
		return
	}
	summary.RankBets = fmt.Sprintf("%.2f", Chip2Float(utils.ToInt64(data_bets["bets"])))
	summary.Prize = fmt.Sprintf("%.2f", Chip2Float(utils.ToInt64(data_bets["prize"])))
	return
}

// GetActivityTurnDates 转盘活动汇总
func (s *activityService) GetActivityTurnDates(begin, end time.Time, registArea int, utypes []int, channelIds []string) (stats []*entity.ActivityTurnStatDate, err error) {
	sql_u_filter := `
		AND userid IN (
			SELECT userid FROM game.col_user final where 1 = 1 %s
		)
	`
	var where_u string     // = " AND ctime BETWEEN ? AND ?"
	var where_u_args []any // = []any{begin, end}
	if registArea > 0 {
		where_u += " AND regist_area = ?"
		where_u_args = append(where_u_args, registArea-1)
	}

	var utypeFilters []string
	if len(utypes) > 0 {
		for _, utype := range utypes {
			// state 1.新手 2.正常 3.平民 4.泡沫
			state, sMoney, eMoney := 2, -1, -1
			switch utype {
			case 0: // 新手
				state = 1
				sMoney, eMoney = 0, 0
			case 1: // 平民
				state = -1
				sMoney, eMoney = 0, 0
			case 2: // 普充
				sMoney, eMoney = 1, 99999
			case 3: // 小R
				sMoney, eMoney = 100000, 499999
			case 4: // 中R
				sMoney, eMoney = 500000, 999999
			case 5: // 大R
				sMoney, eMoney = 1000000, 19999999
			case 6: // 超大R
				sMoney, eMoney = 20000000, -1
			}

			var utypeQs []string
			if state != -1 {
				utypeQs = append(utypeQs, fmt.Sprintf("state = %d", state))
			}
			if sMoney != -1 {
				utypeQs = append(utypeQs, fmt.Sprintf("money >= %d", sMoney))
			}
			if eMoney != -1 {
				utypeQs = append(utypeQs, fmt.Sprintf("money <= %d", eMoney))
			}
			if len(utypeQs) > 0 {
				utypeFilters = append(utypeFilters, fmt.Sprintf("(%s)", strings.Join(utypeQs, " AND ")))
			}
		}
	}
	if len(utypeFilters) > 0 {
		utypeQFmt := " AND (%s)"
		if len(utypeFilters) == 1 {
			utypeQFmt = " AND %s"
		}
		where_u += fmt.Sprintf(utypeQFmt, strings.Join(utypeFilters, " OR "))
	}
	if len(channelIds) > 0 {
		where_u += " AND ad__bundle_id in ?"
		where_u_args = append(where_u_args, channelIds)
	}
	// 无用户筛选条件
	if where_u == "" {
		sql_u_filter = ""
		where_u_args = []any{}
	} else {
		sql_u_filter = fmt.Sprintf(sql_u_filter, where_u)
	}

	var dateStats = make(map[uint32]*entity.ActivityTurnStatDate)

	// 新老顾客日活
	var loginDatas []map[string]any
	login_args := append([]any{begin, end}, where_u_args...)
	login_args = append(login_args, login_args...)
	err = ck.Select(&loginDatas, fmt.Sprintf(`
		select s1.datestr, s1.players, s2.new_players, (players - new_players) old_players
		from (
			select datestr, count(*) players
			from (
				select toYYYYMMDD(login_time) datestr
				from game.col_log_login final
				where login_time between ? AND ? %s
				group by datestr, userid
			) t1
			group by datestr
		) s1
		left join (
			select toYYYYMMDD(ctime) datestr, count(*) new_players
			from game.col_user final
			where ctime between ? AND ? %s
			group by datestr
		) s2 on s1.datestr = s2.datestr
		order by s1.datestr desc
	`, sql_u_filter, where_u), login_args...)
	if err != nil {
		return
	}
	for _, data := range loginDatas {
		date := data["datestr"].(uint32)
		players := utils.ToInt64(data["players"])
		new_players := utils.ToInt64(data["new_players"])
		old_players := utils.ToInt64(data["old_players"])
		datestr := fmt.Sprint(date)
		stat := &entity.ActivityTurnStatDate{
			SDate:           datestr[0:4] + "/" + datestr[4:6] + "/" + datestr[6:8],
			LoginedUsers:    players,
			NewLoginedUsers: new_players,
			OldLoginedUsers: old_players,
		}
		stats = append(stats, stat)
		dateStats[date] = stat
	}

	// 转盘参与统计 总、新、老、首次参与
	sql_turn_users := `
		SELECT t1.ctimestr, t1.draw_users, t1.new_draw_users, t1.old_draw_users, t1.prize_users, t2.first_draw_users, t2.first_draw_new_users, t2.first_draw_old_users FROM (
			SELECT ctimestr, count(*) draw_users, SUM(prized) prize_users, SUM(new_user_draw) new_draw_users, (draw_users - new_draw_users) old_draw_users FROM (
				SELECT userid, toYYYYMMDD(toDateTime(ctime)) ctimestr, toYYYYMMDD(toDateTime(MAX(rtime))) rtimestr, (CASE WHEN ctimestr = rtimestr THEN 1 ELSE 0 END) new_user_draw,
					MAX(CASE WHEN score >= score_target THEN 1 ELSE 0 END) prized
				FROM game.col_activity_turn_draw_log FINAL
				WHERE ctime BETWEEN ? AND ? %s
				GROUP BY userid, ctimestr
			) s1
			GROUP BY ctimestr
		) t1 LEFT JOIN (
			SELECT ctimestr_min, count(*) first_draw_users, SUM(first_draw_new_user) first_draw_new_users, (first_draw_users - first_draw_new_users) first_draw_old_users  FROM (
				SELECT userid, MIN(ctime) ctime_min, toYYYYMMDD(toDateTime(ctime_min)) ctimestr_min,  (CASE WHEN ctimestr_min = toYYYYMMDD(toDateTime(MAX(rtime))) THEN 1 ELSE 0 END) first_draw_new_user 
				FROM game.col_activity_turn_draw_log FINAL
				WHERE 1 = 1 %s
				GROUP BY userid
				HAVING ctime_min BETWEEN ? AND ?
			) s1
			GROUP BY ctimestr_min
		) t2 ON t1.ctimestr = t2.ctimestr_min
	`
	sql_turn_users = fmt.Sprintf(sql_turn_users, sql_u_filter, sql_u_filter)
	turn_users_args := append([]any{begin.Unix(), end.Unix()}, where_u_args...)
	turn_users_args = append(turn_users_args, where_u_args...)
	turn_users_args = append(turn_users_args, begin.Unix(), end.Unix())
	var turn_users_datas []map[string]any
	err = ck.Select(&turn_users_datas, sql_turn_users, turn_users_args...)
	if err != nil {
		return
	}
	for _, data := range turn_users_datas {
		ctimestr := data["ctimestr"].(uint32)
		if stat, ok := dateStats[ctimestr]; ok {
			stat.TurnUsers = utils.ToInt64(data["draw_users"])
			stat.NewTurnUsers = utils.ToInt64(data["new_draw_users"])
			stat.OldTurnUsers = utils.ToInt64(data["old_draw_users"])
			stat.FirstTurnUsers = utils.ToInt64(data["first_draw_users"])
			stat.FirstNewTurnUsers = utils.ToInt64(data["first_draw_new_users"])
			stat.FirstOldTurnUsers = utils.ToInt64(data["first_draw_old_users"])
			stat.PrizeUsers = utils.ToInt64(data["prize_users"])
		}
	}

	// 申请领奖统计
	sql_tack_prizes := `
		SELECT toYYYYMMDD(toDateTime(ctime)) ctimestr, count(distinct userid) tack_users, SUM(prize) prize_sum, 
			count(distinct (CASE WHEN state = 1 THEN userid ELSE null END)) tack_pass_users,
			SUM(CASE WHEN state = 1 THEN prize ELSE 0 END) tack_pass_prizes 
		FROM game.col_activity_turn_prize_log FINAL
		WHERE ctime BETWEEN ? AND ? %s
		GROUP BY ctimestr
	`
	sql_tack_prizes = fmt.Sprintf(sql_tack_prizes, sql_u_filter)
	tack_prizes_args := append([]any{begin.Unix(), end.Unix()}, where_u_args...)
	var tack_prizes_datas []map[string]any
	err = ck.Select(&tack_prizes_datas, sql_tack_prizes, tack_prizes_args...)
	if err != nil {
		return
	}
	for _, data := range tack_prizes_datas {
		ctimestr := data["ctimestr"].(uint32)
		if stat, ok := dateStats[ctimestr]; ok {
			stat.TackPrizeUsers = utils.ToInt64(data["tack_users"])
			stat.TackPrizes0 = utils.ToInt64(data["prize_sum"])
			stat.TackPrizes = fmt.Sprintf("%.2f", Chip2Float(stat.TackPrizes0))
			stat.TackPassPrizeUsers = utils.ToInt64(data["tack_pass_users"])
			stat.TackPassPrizes0 = utils.ToInt64(data["tack_pass_prizes"])
			stat.TackPassPrizes = fmt.Sprintf("%.2f", Chip2Float(stat.TackPassPrizes0))
		}
	}

	// 审核通过统计
	sql_pass_prizes := `
		SELECT toYYYYMMDD(toDateTime(stime)) stimestr, count(distinct userid) pass_users, SUM(prize) pass_prizes
		FROM game.col_activity_turn_prize_log FINAL
		WHERE state = 1 AND ctime BETWEEN ? AND ? %s
		GROUP BY stimestr
	`
	sql_pass_prizes = fmt.Sprintf(sql_pass_prizes, sql_u_filter)
	pass_prizes_args := append([]any{begin.Unix(), end.Unix()}, where_u_args...)
	var pass_prizes_datas []map[string]any
	err = ck.Select(&pass_prizes_datas, sql_pass_prizes, pass_prizes_args...)
	if err != nil {
		return
	}
	for _, data := range pass_prizes_datas {
		stimestr := data["stimestr"].(uint32)
		if stat, ok := dateStats[stimestr]; ok {
			stat.PassPrizeUsers = utils.ToInt64(data["pass_users"])
			stat.PassPrizes0 = utils.ToInt64(data["pass_prizes"])
			stat.PassPrizes = fmt.Sprintf("%.2f", Chip2Float(stat.PassPrizes0))
		}
	}

	// 邀请统计
	sql_invites := `
			SELECT s1.ctimestr ctimestr, count(*) inviters, SUM(s1.new_inviter) new_inviters, (inviters - new_inviters) old_inviters,
				SUM(s2.invited) invited_inviters, SUM(CASE WHEN s2.invited=1 AND s1.new_inviter=1 THEN 1 ELSE 0 END) invited_new_inviters, 
				(invited_inviters - invited_new_inviters) invited_old_inviters,
				SUM(CASE WHEN s3.ctimestr >= s1.ctimestr AND s3.invited=1 THEN 1 ELSE 0 END) now_invited_inviters,
				SUM(CASE WHEN s3.ctimestr >= s1.ctimestr AND s3.invited=1 AND s1.new_inviter=1 THEN 1 ELSE 0 END) now_invited_new_inviters,
				(now_invited_inviters - now_invited_new_inviters) now_invited_old_inviters,
				SUM(s2.invited_count) share_users, SUM(CASE WHEN s1.new_inviter=1 THEN s2.invited_count ELSE 0 END) new_inviter_share_users,
				(share_users - new_inviter_share_users) old_inviter_share_users
			FROM (
				SELECT toYYYYMMDD(toDateTime(ctime/1000)) ctimestr, userid, 
					(CASE WHEN ctimestr = toYYYYMMDD(toDateTime(MAX(rtime)/1000)) THEN 1 ELSE 0 END) new_inviter
				FROM game.col_launch_invite_log FINAL
				WHERE ctime BETWEEN ? AND ? AND ltype = 2 %s
				GROUP BY ctimestr, userid
			) s1 LEFT JOIN (
				SELECT toYYYYMMDD(ctime) ctimestr, share_superior, count(*) invited_count, 1 invited
				FROM game.col_user FINAL
				WHERE share_superior != '' AND share_source = 1 AND ctime BETWEEN ? AND ? %s
				GROUP BY ctimestr, share_superior
			) s2 ON s1.ctimestr = s2.ctimestr AND s1.userid = s2.share_superior
			LEFT JOIN (
				SELECT share_superior, toYYYYMMDD(MAX(ctime)) ctimestr, 1 invited
				FROM game.col_user FINAL
				WHERE share_superior != '' AND share_source = 1 AND ctime >= ? %s
				GROUP BY share_superior
			) s3 ON s1.userid = s3.share_superior
			GROUP BY s1.ctimestr
	`
	sql_invites = fmt.Sprintf(sql_invites, sql_u_filter, where_u, where_u)
	invites_args := append([]any{begin.UnixMilli(), end.UnixMilli()}, where_u_args...)
	invites_args = append(append(invites_args, begin, end), where_u_args...)
	invites_args = append(append(invites_args, begin), where_u_args...)
	var invites_datas []map[string]any
	err = ck.Select(&invites_datas, sql_invites, invites_args...)
	if err != nil {
		return
	}
	for _, data := range invites_datas {
		ctimestr := data["ctimestr"].(uint32)
		if stat, ok := dateStats[ctimestr]; ok {
			stat.LaunchInviters = utils.ToInt64(data["inviters"])
			stat.LaunchNewInviters = utils.ToInt64(data["new_inviters"])
			stat.LaunchOldInviters = utils.ToInt64(data["old_inviters"])
			stat.InvitedInviters = utils.ToInt64(data["invited_inviters"])
			stat.InvitedNewInviters = utils.ToInt64(data["invited_new_inviters"])
			stat.InvitedOldInviters = utils.ToInt64(data["invited_old_inviters"])
			stat.NowInvitedInviters = utils.ToInt64(data["now_invited_inviters"])
			stat.NowInvitedNewInviters = utils.ToInt64(data["now_invited_new_inviters"])
			stat.NowInvitedOldInviters = utils.ToInt64(data["now_invited_old_inviters"])
			stat.ShareUsers = utils.ToInt64(data["share_users"])
			stat.NewInviterShareUsers = utils.ToInt64(data["new_inviter_share_users"])
			stat.OldInviterShareUsers = utils.ToInt64(data["old_inviter_share_users"])
		}
	}

	// 裂变总人数
	var share_user_datas []map[string]any
	var share_user_args = []any{begin, end}
	share_user_args = append(share_user_args, where_u_args...)
	err = ck.Select(&share_user_datas, fmt.Sprintf(`
		SELECT toYYYYMMDD(ctime) ctimestr, count(*) share_users
		FROM game.col_user FINAL
		WHERE ctime BETWEEN ? and ? AND share_superior != '' AND share_source = 1 %s
		GROUP BY ctimestr
		order by ctimestr
	`, where_u), share_user_args...)
	if err != nil {
		return
	}
	for _, data := range share_user_datas {
		ctimestr := data["ctimestr"].(uint32)
		if stat, ok := dateStats[ctimestr]; ok {
			stat.ShareUsers = utils.ToInt64(data["share_users"])
			stat.OldInviterShareUsers = stat.ShareUsers - stat.NewInviterShareUsers
		}
	}

	// 汇总 summary
	summary := s.activityTurnDatesSummary(stats)
	// summary.SDate = fmt.Sprintf("%s-%s汇总", begin.Format(utils.FORMAT_DATE2), end.Format(utils.FORMAT_DATE2))
	stats = append([]*entity.ActivityTurnStatDate{summary}, stats...)

	for _, stat := range stats {
		stat.LoginTurnRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.TurnUsers, stat.LoginedUsers)*100)
		stat.NewUserTurnRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.NewTurnUsers, stat.NewLoginedUsers)*100)
		stat.OldUserTurnRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.OldTurnUsers, stat.OldLoginedUsers)*100)
		stat.LoginFirstTurnRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.FirstTurnUsers, stat.LoginedUsers)*100)
		stat.NewUserFirstTurnRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.FirstNewTurnUsers, stat.NewLoginedUsers)*100)
		stat.OldUserFirstTurnRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.FirstOldTurnUsers, stat.OldLoginedUsers)*100)
		stat.PrizeTurnRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.PrizeUsers, stat.TurnUsers)*100)
		stat.TackPrizeRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.TackPrizeUsers, stat.PrizeUsers)*100)
		stat.TackPrizePassRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.TackPassPrizeUsers, stat.TackPrizeUsers)*100)

		stat.TurnUserInviteRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.LaunchInviters, stat.TurnUsers)*100)
		stat.TurnNewUserInviteRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.LaunchNewInviters, stat.TurnUsers)*100)
		stat.TurnOldUserInviteRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.LaunchOldInviters, stat.TurnUsers)*100)
		stat.InviterInvitedRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.InvitedInviters, stat.LaunchInviters)*100)
		stat.NewInviterInvitedRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.InvitedNewInviters, stat.LaunchNewInviters)*100)
		stat.OldInviterInvitedRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.InvitedOldInviters, stat.LaunchOldInviters)*100)
		stat.InviterNowInvitedRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.NowInvitedInviters, stat.LaunchInviters)*100)
		stat.NewInviterNowInvitedRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.NowInvitedNewInviters, stat.LaunchNewInviters)*100)
		stat.OldInviterNowInvitedRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.NowInvitedOldInviters, stat.LaunchOldInviters)*100)
		stat.ShareUsersRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.ShareUsers, stat.TurnUsers)*100)
		stat.NewInviterShareUsersRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.NewInviterShareUsers, stat.NewTurnUsers)*100)
		stat.OldInviterShareUsersRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.OldInviterShareUsers, stat.OldTurnUsers)*100)
		stat.ShareInviterRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.ShareUsers, stat.LaunchInviters)*100)
		stat.NewShareInviterRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.NewInviterShareUsers, stat.LaunchNewInviters)*100)
		stat.OldShareInviterRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.OldInviterShareUsers, stat.LaunchOldInviters)*100)
	}
	return
}

// activityTurnDatesSummary 转盘活动汇总 汇总
func (s *activityService) activityTurnDatesSummary(stats []*entity.ActivityTurnStatDate) (summary *entity.ActivityTurnStatDate) {
	summary = &entity.ActivityTurnStatDate{SDate: "汇总"}
	for _, stat := range stats {
		summary.LoginedUsers += stat.LoginedUsers
		summary.NewLoginedUsers += stat.NewLoginedUsers
		summary.OldLoginedUsers += stat.OldLoginedUsers
		summary.TurnUsers += stat.TurnUsers
		summary.NewTurnUsers += stat.NewTurnUsers
		summary.OldTurnUsers += stat.OldTurnUsers
		summary.FirstTurnUsers += stat.FirstTurnUsers
		summary.FirstNewTurnUsers += stat.FirstNewTurnUsers
		summary.FirstOldTurnUsers += stat.FirstOldTurnUsers
		summary.PrizeUsers += stat.PrizeUsers
		summary.TackPrizeUsers += stat.TackPrizeUsers
		summary.TackPrizes0 += stat.TackPrizes0
		summary.TackPassPrizeUsers += stat.TackPassPrizeUsers
		summary.TackPassPrizes0 += stat.TackPassPrizes0
		summary.PassPrizeUsers += stat.PassPrizeUsers
		summary.PassPrizes0 += stat.PassPrizes0

		summary.LaunchInviters += stat.LaunchInviters
		summary.LaunchNewInviters += stat.LaunchNewInviters
		summary.LaunchOldInviters += stat.LaunchOldInviters
		summary.InvitedInviters += stat.InvitedInviters
		summary.InvitedNewInviters += stat.InvitedNewInviters
		summary.InvitedOldInviters += stat.InvitedOldInviters
		summary.NowInvitedInviters += stat.NowInvitedInviters
		summary.NowInvitedNewInviters += stat.NowInvitedNewInviters
		summary.NowInvitedOldInviters += stat.NowInvitedOldInviters
		summary.ShareUsers += stat.ShareUsers
		summary.NewInviterShareUsers += stat.NewInviterShareUsers
		summary.OldInviterShareUsers += stat.OldInviterShareUsers
	}

	summary.TackPrizes = fmt.Sprintf("%.2f", Chip2Float(summary.TackPrizes0))
	summary.TackPassPrizes = fmt.Sprintf("%.2f", Chip2Float(summary.TackPassPrizes0))
	summary.PassPrizes = fmt.Sprintf("%.2f", Chip2Float(summary.PassPrizes0))
	return
}

// GetActivityTurnDates 转盘活动明细
func (s *activityService) GetActivityTurnUsers(page, pageSize int, userid string, begin, end time.Time, begin2, end2 *time.Time, registArea int, utypes []int, channelIds []string) (total int64, stats []*entity.ActivityTurnStatUser, err error) {
	sql_u_filter := `
		AND userid IN (
			SELECT userid FROM game.col_user final where 1 = 1 %s
		)
	`
	var where_u string
	var where_u_args []any

	if begin2 != nil {
		where_u += " AND ctime >= ?"
		where_u_args = append(where_u_args, *begin2)
	}
	if end2 != nil {
		where_u += " AND ctime <= ?"
		where_u_args = append(where_u_args, *end2)
	}

	if registArea > 0 {
		where_u += " AND regist_area = ?"
		where_u_args = append(where_u_args, registArea-1)
	}

	var utypeFilters []string
	if len(utypes) > 0 {
		for _, utype := range utypes {
			// state 1.新手 2.正常 3.平民 4.泡沫
			state, sMoney, eMoney := 2, -1, -1
			switch utype {
			case 0: // 新手
				state = 1
				sMoney, eMoney = 0, 0
			case 1: // 平民
				state = -1
				sMoney, eMoney = 0, 0
			case 2: // 普充
				sMoney, eMoney = 1, 99999
			case 3: // 小R
				sMoney, eMoney = 100000, 499999
			case 4: // 中R
				sMoney, eMoney = 500000, 999999
			case 5: // 大R
				sMoney, eMoney = 1000000, 19999999
			case 6: // 超大R
				sMoney, eMoney = 20000000, -1
			}
			var utypeQs []string
			if state != -1 {
				utypeQs = append(utypeQs, fmt.Sprintf("state = %d", state))
			}
			if sMoney != -1 {
				utypeQs = append(utypeQs, fmt.Sprintf("money >= %d", sMoney))
			}
			if eMoney != -1 {
				utypeQs = append(utypeQs, fmt.Sprintf("money <= %d", eMoney))
			}
			if len(utypeQs) > 0 {
				utypeFilters = append(utypeFilters, fmt.Sprintf("(%s)", strings.Join(utypeQs, " AND ")))
			}
		}
	}
	if len(utypeFilters) > 0 {
		utypeQFmt := " AND (%s)"
		if len(utypeFilters) == 1 {
			utypeQFmt = " AND %s"
		}
		where_u += fmt.Sprintf(utypeQFmt, strings.Join(utypeFilters, " OR "))
	}
	if len(channelIds) > 0 {
		where_u += " AND ad__bundle_id in ?"
		where_u_args = append(where_u_args, channelIds)
	}
	// 无用户筛选条件
	if where_u == "" {
		sql_u_filter = ""
		where_u_args = []any{}
	} else {
		sql_u_filter = fmt.Sprintf(sql_u_filter, where_u)
	}

	sql_draw_users_count := `
		SELECT count(*) total FROM (
			SELECT userid
			FROM game.col_activity_turn_draw_log FINAL
			WHERE ctime BETWEEN ? AND ? %s
			GROUP BY userid
		) s1
	`
	sql_draw_users_count = fmt.Sprintf(sql_draw_users_count, sql_u_filter)
	draw_users_count_args := append([]any{begin.Unix(), end.Unix()}, where_u_args...)
	err = ck.Select(&total, sql_draw_users_count, draw_users_count_args...)
	if err != nil {
		return
	}
	if total == 0 {
		return
	}

	channelsMap, err := ChannelService.GetChannelsMap()
	if err != nil {
		return
	}

	var useridStats = make(map[string]*entity.ActivityTurnStatUser)
	var userids []string

	offset, limit := PageCalc(page, pageSize)
	sql_draw_users := `
		SELECT s0.userid userid, s0.ctime_min first_draw_time, s0.give_times,
			s1.ad__bundle_id, s1.regist_area, s1.ctime, s1.login_time, s1.vip_lv, s1.state, s1.money
		FROM game.col_user s1 FINAL
		JOIN (
			SELECT userid, ctime_max, ctime_min, give_times FROM (
				SELECT userid, MAX(ctime) ctime_max, MIN(ctime) ctime_min, SUM(CASE WHEN reason = 0 THEN 1 ELSE 0 END) give_times
				FROM game.col_activity_turn_draw_log FINAL
				WHERE ctime BETWEEN ? AND ? %s
				GROUP BY userid
			) t1
			ORDER BY ctime_max DESC
			LIMIT ?, ?
		) s0 ON s1.userid = s0.userid
	`
	sql_draw_users = fmt.Sprintf(sql_draw_users, sql_u_filter)
	draw_users_args := append([]any{begin.Unix(), end.Unix()}, where_u_args...)
	draw_users_args = append(draw_users_args, offset, limit)
	var draw_users_datas []map[string]any
	err = ck.Select(&draw_users_datas, sql_draw_users, draw_users_args...)
	if err != nil {
		return
	}
	now := NowTime()
	for _, data := range draw_users_datas {
		userid := data["userid"].(string)
		ad__bundle_id := data["ad__bundle_id"].(string)
		ctime := data["ctime"].(time.Time)
		login_time := data["login_time"].(time.Time)
		regist_area := utils.ToInt64(data["regist_area"])
		vip_lv := utils.ToInt64(data["vip_lv"])
		state := utils.ToInt64(data["state"])
		money := utils.ToInt64(data["money"])
		first_draw_time := utils.ToInt64(data["first_draw_time"])
		give_times := utils.ToInt64(data["give_times"])

		registAreaF, _ := utils.CaseWhen3(regist_area, 0, "A类", 1, "B类", 2, "C类")
		ctimeF := ctime.Format(utils.FORMAT)
		loginTimeF := login_time.Format(utils.FORMAT)
		liveDays := int64(math.Ceil(login_time.Sub(ctime).Hours() / 24))
		loseDays := int64(math.Floor(now.Sub(login_time).Hours() / 24))
		var channelClass string
		if c, ok := channelsMap[ad__bundle_id]; ok {
			channelClass = c.ClassName
		}
		var turns int64
		if first_draw_time > 0 {
			turns = int64(math.Ceil((float64(now.Unix()-first_draw_time) / 60 / 60 / 72)))
		}

		var userType string
		if money == 0 {
			if state == 1 {
				userType = "新手"
			} else {
				userType = "平民"
			}
		} else if money >= 1 && money <= 99999 {
			userType = "普充"
		} else if money >= 100000 && money <= 499999 {
			userType = "小R"
		} else if money >= 500000 && money <= 999999 {
			userType = "中R"
		} else if money >= 1000000 && money <= 19999999 {
			userType = "大R"
		} else if money >= 20000000 {
			userType = "超大R"
		}

		stat := &entity.ActivityTurnStatUser{
			Userid:       userid,
			ChannelClass: channelClass,
			ChannelId:    ad__bundle_id,
			RegistArea:   registAreaF,
			Ctime:        ctimeF,
			LoginTime:    loginTimeF,
			LiveDays:     fmt.Sprint(liveDays),
			LoseDays:     fmt.Sprint(loseDays),
			VipLv:        fmt.Sprint(vip_lv),
			Turns:        turns,
			DrawTurns:    give_times,
			FUserType:    userType,
		}
		stats = append(stats, stat)
		useridStats[userid] = stat
		userids = append(userids, userid)
	}

	if len(userids) == 0 {
		return
	}

	// 充值提现查询
	sql_pays := `
		SELECT userid, SUM(amount) pays
		FROM game.col_trade_record FINAL
		WHERE order_status = 4 AND userid IN ?
		GROUP BY userid
	`
	var pays_datas []map[string]any
	err = ck.Select(&pays_datas, sql_pays, userids)
	if err != nil {
		return
	}
	for _, data := range pays_datas {
		userid := data["userid"].(string)
		if stat, ok := useridStats[userid]; ok {
			stat.Pays = fmt.Sprintf("%.2f", Chip2Float(utils.ToInt64(data["pays"])))
		}
	}

	sql_withdraws := `
		SELECT userid, SUM(amount + commission) withdraws
		FROM game.col_withdraw_record FINAL
		WHERE order_status = 2 AND userid IN ?
		GROUP BY userid
	`
	var withdraws_datas []map[string]any
	err = ck.Select(&withdraws_datas, sql_withdraws, userids)
	if err != nil {
		return
	}
	for _, data := range withdraws_datas {
		userid := data["userid"].(string)
		if stat, ok := useridStats[userid]; ok {
			stat.Withdraws = fmt.Sprintf("%.2f", Chip2Float(utils.ToInt64(data["withdraws"])))
		}
	}

	// 邀请统计
	sql_invites := `
		SELECT share_superior, count(*) invite_users,
			SUM(CASE WHEN money > 0 THEN 1 ELSE 0 END) invite_pay_users,
			SUM(money) invite_users_pays,
			SUM(cash_out) invite_users_withdraws,
			(invite_users_pays - invite_users_withdraws) invite_users_contribute
		FROM game.col_user FINAL
		WHERE share_superior IN ? AND share_source = 1
		GROUP BY share_superior
	`
	var invites_datas []map[string]any
	err = ck.Select(&invites_datas, sql_invites, userids)
	if err != nil {
		return
	}
	for _, data := range invites_datas {
		share_superior := data["share_superior"].(string)
		if stat, ok := useridStats[share_superior]; ok {
			stat.InviteUsers = utils.ToInt64(data["invite_users"])
			stat.InviteUsersPayed = utils.ToInt64(data["invite_pay_users"])
			stat.InviteUsersPays = fmt.Sprintf("%.2f", Chip2Float(utils.ToInt64(data["invite_users_pays"])))
			stat.InviteUsersWithdraws = fmt.Sprintf("%.2f", Chip2Float(utils.ToInt64(data["invite_users_withdraws"])))
			stat.InviteUsersContribute0 = utils.ToInt64(data["invite_users_contribute"])
			stat.InviteUsersContribute = fmt.Sprintf("%.2f", Chip2Float(stat.InviteUsersContribute0))
		}
	}

	// 参与轮均邀请统计
	sql_draw_invites := `
		SELECT userid, AVG(turn_invite_users) turn_invite_users_avg, AVG(turn_invite_pay_users) turn_invite_pay_users_avg,
			SUM(bonus_contribute) bonus_contribute,
			SUM(cash_contribute) cash_contribute,
			SUM(withdrawable_contribute) withdrawable_contribute
		FROM (
			SELECT s0.userid, s0.stime, s0.etime, SUM(s1.invited) turn_invite_users, SUM(s1.invited_pay) turn_invite_pay_users,
				SUM(CASE WHEN s0.prize_type=1 THEN s1.contribute ELSE 0 END) bonus_contribute,
				SUM(CASE WHEN s0.prize_type=2 THEN s1.contribute ELSE 0 END) cash_contribute,
				SUM(CASE WHEN s0.prize_type=3 THEN s1.contribute ELSE 0 END) withdrawable_contribute
			FROM (
				SELECT userid, toDateTime(turn_stime) stime, toDateTime(turn_etime) etime, prize_type
				FROM game.col_activity_turn_draw_log FINAL
				WHERE userid IN ?
				GROUP BY turn_stime, turn_etime, userid, prize_type
			) s0 LEFT JOIN (
				SELECT userid, share_superior, ctime, 1 invited, 
					(CASE WHEN money > 0 THEN 1 ELSE 0 END) invited_pay,
					(money - cash_out) contribute
				FROM game.col_user FINAL
				WHERE share_superior IN ? AND share_source = 1
			) s1 ON s1.share_superior = s0.userid AND s1.ctime BETWEEN s0.stime AND s0.etime
			GROUP BY s0.userid, s0.stime, s0.etime
			SETTINGS allow_experimental_join_condition = 1
		) t1
		GROUP BY userid
	`
	var draw_invites_datas []map[string]any
	err = ck.Select(&draw_invites_datas, sql_draw_invites, userids, userids)
	if err != nil {
		return
	}
	for _, data := range draw_invites_datas {
		userid := data["userid"].(string)
		if stat, ok := useridStats[userid]; ok {
			stat.DrawTurnInviteUsersAvg = fmt.Sprintf("%.2f", utils.ToFloat64(data["turn_invite_users_avg"]))
			stat.DrawTurnInviteUsersPayedAvg = fmt.Sprintf("%.2f", utils.ToFloat64(data["turn_invite_pay_users_avg"]))

			stat.BonusContribute = utils.ToInt64(data["bonus_contribute"])
			stat.CashContribute = utils.ToInt64(data["cash_contribute"])
			stat.WithdrawableContribute = utils.ToInt64(data["withdrawable_contribute"])
		}
	}

	sql_prize_invites := `
		SELECT userid, AVG(turn_invite_users) turn_invite_users_avg, AVG(turn_invite_pay_users) turn_invite_pay_users_avg FROM (
			SELECT s0.userid, s0.stime, s0.etime, SUM(s1.invited) turn_invite_users, SUM(s1.invited_pay) turn_invite_pay_users
			FROM (
				SELECT userid, toDateTime(turn_stime) stime, toDateTime(turn_etime) etime
				FROM game.col_activity_turn_prize_log FINAL
				WHERE userid IN ?
				GROUP BY userid, turn_stime, turn_etime
			) s0 LEFT JOIN (
				SELECT userid, share_superior, ctime, 1 invited, (CASE WHEN money > 0 THEN 1 ELSE 0 END) invited_pay
				FROM game.col_user FINAL
				WHERE share_superior IN ? AND share_source = 1
			) s1 ON s1.share_superior = s0.userid AND s1.ctime BETWEEN s0.stime AND s0.etime
			GROUP BY s0.userid, s0.stime, s0.etime
			SETTINGS allow_experimental_join_condition = 1
		) t1
		GROUP BY userid
	`
	var prize_invites_datas []map[string]any
	err = ck.Select(&prize_invites_datas, sql_prize_invites, userids, userids)
	if err != nil {
		return
	}
	for _, data := range prize_invites_datas {
		userid := data["userid"].(string)
		if stat, ok := useridStats[userid]; ok {
			stat.TackPrizeTurnInviteUsersAvg = fmt.Sprintf("%.2f", utils.ToFloat64(data["turn_invite_users_avg"]))
			stat.TackPrizeTurnInviteUsersPayedAvg = fmt.Sprintf("%.2f", utils.ToFloat64(data["turn_invite_pay_users_avg"]))
		}
	}

	// 转盘领奖统计
	sql_prizes := `
		SELECT userid, count(*) prize_turns, 
			SUM(CASE WHEN state=1 THEN 1 ELSE 0 END) pass_prize_turns, 
			SUM(CASE WHEN state=1 THEN prize ELSE 0 END) pass_prizes,
			SUM(CASE WHEN state=1 AND prize_type=1 THEN prize ELSE 0 END) pass_prize_bouns,
			SUM(CASE WHEN state=1 AND prize_type=2 THEN prize ELSE 0 END) pass_prize_cash,
			SUM(CASE WHEN state=1 AND prize_type=3 THEN prize ELSE 0 END) pass_prize_withdrawalble
		FROM game.col_activity_turn_prize_log FINAL
		WHERE userid IN ?
		GROUP BY userid
	`
	var prizes_datas []map[string]any
	err = ck.Select(&prizes_datas, sql_prizes, userids)
	if err != nil {
		return
	}
	for _, data := range prizes_datas {
		userid := data["userid"].(string)
		if stat, ok := useridStats[userid]; ok {
			prize_turns := utils.ToInt64(data["prize_turns"])
			pass_prize_turns := utils.ToInt64(data["pass_prize_turns"])
			pass_prizes := utils.ToInt64(data["pass_prizes"])
			pass_prize_bouns := utils.ToInt64(data["pass_prize_bouns"])
			pass_prize_cash := utils.ToInt64(data["pass_prize_cash"])
			pass_prize_withdrawalble := utils.ToInt64(data["pass_prize_withdrawalble"])

			stat.TackPrizeTurns = prize_turns
			stat.TackPrizePassTurns = pass_prize_turns
			stat.TackPrizePassScore = fmt.Sprintf("%.2f", Chip2Float(pass_prizes))
			stat.TackPrizePassBonus = fmt.Sprintf("%.2f(%.2f%%)", Chip2Float(pass_prize_bouns), ComputeFloat(pass_prize_bouns, pass_prizes)*100)
			stat.TackPrizePassCash = fmt.Sprintf("%.2f(%.2f%%)", Chip2Float(pass_prize_cash), ComputeFloat(pass_prize_cash, pass_prizes)*100)
			stat.TackPrizePassWithdrawable = fmt.Sprintf("%.2f(%.2f%%)", Chip2Float(pass_prize_withdrawalble), ComputeFloat(pass_prize_withdrawalble, pass_prizes)*100)
			stat.ScoreRoi = fmt.Sprintf("%.2f", ComputeFloat(stat.InviteUsersContribute0, pass_prizes))
			stat.BonusRoi = fmt.Sprintf("%.2f", ComputeFloat(stat.BonusContribute, pass_prize_bouns))
			stat.CashRoi = fmt.Sprintf("%.2f", ComputeFloat(stat.CashContribute, pass_prize_cash))
			stat.WithdrawableRoi = fmt.Sprintf("%.2f", ComputeFloat(stat.WithdrawableContribute, pass_prize_withdrawalble))
		}
	}

	for _, stat := range stats {
		stat.TurnRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.DrawTurns, stat.Turns)*100)
		stat.TackPrizeRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.TackPrizeTurns, stat.DrawTurns)*100)
		stat.TackPrizePassRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.TackPrizePassTurns, stat.TackPrizeTurns)*100)
		stat.InviteUsersPayedRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.InviteUsersPayed, stat.InviteUsers)*100)
	}

	summary, err := s.activityTurnUsersSummary(begin, end, sql_u_filter, where_u_args)
	if err != nil {
		return
	}
	stats = append([]*entity.ActivityTurnStatUser{summary}, stats...)
	return
}

func (s *activityService) activityTurnUsersSummary(begin, end time.Time, sql_u_filter string, where_u_args []any) (summary *entity.ActivityTurnStatUser, err error) {
	summary = &entity.ActivityTurnStatUser{
		Userid:       "汇总",
		ChannelClass: "--",
		ChannelId:    "--",
		RegistArea:   "--",
		Ctime:        "--",
		LoginTime:    "--",
		LiveDays:     "--",
		LoseDays:     "--",
		VipLv:        "--",
	}

	userids_filter := `
		SELECT userid
		FROM game.col_activity_turn_draw_log FINAL
		WHERE ctime BETWEEN ? AND ? %s
		GROUP BY userid
	`
	userids_filter = fmt.Sprintf(userids_filter, sql_u_filter)
	userids_args := append([]any{begin.Unix(), end.Unix()}, where_u_args...)

	sql_draw_users := `
		SELECT MIN(ctime) first_draw_time,
			SUM(CASE WHEN reason = 0 THEN 1 ELSE 0 END) give_times
		FROM game.col_activity_turn_draw_log FINAL
		WHERE ctime BETWEEN ? AND ? %s
	`
	sql_draw_users = fmt.Sprintf(sql_draw_users, sql_u_filter)
	draw_users_args := append([]any{begin.Unix(), end.Unix()}, where_u_args...)
	var draw_users_datas = make(map[string]any)
	err = ck.Select(&draw_users_datas, sql_draw_users, draw_users_args...)
	if err != nil {
		return
	}
	now := NowTime()
	first_draw_time := utils.ToInt64(draw_users_datas["first_draw_time"])
	give_times := utils.ToInt64(draw_users_datas["give_times"])
	summary.Turns = int64(math.Ceil((float64(now.Unix()-first_draw_time) / 60 / 60 / 72)))
	summary.DrawTurns = give_times

	// 充值提现查询
	sql_pays := `
		SELECT SUM(amount) pays
		FROM game.col_trade_record FINAL
		WHERE order_status = 4 AND userid IN (%s)
	`
	sql_pays = fmt.Sprintf(sql_pays, userids_filter)
	var pays_datas = make(map[string]any)
	err = ck.Select(&pays_datas, sql_pays, userids_args...)
	if err != nil {
		return
	}
	summary.Pays = fmt.Sprintf("%.2f", Chip2Float(utils.ToInt64(pays_datas["pays"])))

	sql_withdraws := `
		SELECT SUM(amount + commission) withdraws
		FROM game.col_withdraw_record FINAL
		WHERE order_status = 2 AND userid IN (%s)
	`
	sql_withdraws = fmt.Sprintf(sql_withdraws, userids_filter)
	var withdraws_datas = make(map[string]any)
	err = ck.Select(&withdraws_datas, sql_withdraws, userids_args...)
	if err != nil {
		return
	}
	summary.Withdraws = fmt.Sprintf("%.2f", Chip2Float(utils.ToInt64(withdraws_datas["withdraws"])))

	// 邀请统计
	sql_invites := `
		SELECT count(*) invite_users,
			SUM(CASE WHEN money > 0 THEN 1 ELSE 0 END) invite_pay_users,
			SUM(money) invite_users_pays,
			SUM(cash_out) invite_users_withdraws,
			(invite_users_pays - invite_users_withdraws) invite_users_contribute
		FROM game.col_user FINAL
		WHERE share_superior IN (%s) AND share_source = 1
	`
	sql_invites = fmt.Sprintf(sql_invites, userids_filter)
	var invites_datas = make(map[string]any)
	err = ck.Select(&invites_datas, sql_invites, userids_args...)
	if err != nil {
		return
	}
	summary.InviteUsers = utils.ToInt64(invites_datas["invite_users"])
	summary.InviteUsersPayed = utils.ToInt64(invites_datas["invite_pay_users"])
	summary.InviteUsersPays = fmt.Sprintf("%.2f", Chip2Float(utils.ToInt64(invites_datas["invite_users_pays"])))
	summary.InviteUsersWithdraws = fmt.Sprintf("%.2f", Chip2Float(utils.ToInt64(invites_datas["invite_users_withdraws"])))
	summary.InviteUsersContribute0 = utils.ToInt64(invites_datas["invite_users_contribute"])
	summary.InviteUsersContribute = fmt.Sprintf("%.2f", Chip2Float(summary.InviteUsersContribute0))

	// 参与轮均邀请统计
	sql_draw_invites := `
		SELECT AVG(turn_invite_users) turn_invite_users_avg, AVG(turn_invite_pay_users) turn_invite_pay_users_avg,
			SUM(bonus_contribute) bonus_contribute,
			SUM(cash_contribute) cash_contribute,
			SUM(withdrawable_contribute) withdrawable_contribute
		FROM (
			SELECT s0.userid, s0.stime, s0.etime, SUM(s1.invited) turn_invite_users, SUM(s1.invited_pay) turn_invite_pay_users,
				SUM(CASE WHEN s0.prize_type=1 THEN s1.contribute ELSE 0 END) bonus_contribute,
				SUM(CASE WHEN s0.prize_type=2 THEN s1.contribute ELSE 0 END) cash_contribute,
				SUM(CASE WHEN s0.prize_type=3 THEN s1.contribute ELSE 0 END) withdrawable_contribute
			FROM (
				SELECT userid, toDateTime(turn_stime) stime, toDateTime(turn_etime) etime, prize_type
				FROM game.col_activity_turn_draw_log FINAL
				WHERE userid IN (%s)
				GROUP BY turn_stime, turn_etime, userid, prize_type
			) s0 LEFT JOIN (
				SELECT userid, share_superior, ctime, 1 invited, 
					(CASE WHEN money > 0 THEN 1 ELSE 0 END) invited_pay,
					(money - cash_out) contribute
				FROM game.col_user FINAL
				WHERE share_superior IN (%s) AND share_source = 1
			) s1 ON s1.share_superior = s0.userid AND s1.ctime BETWEEN s0.stime AND s0.etime
			GROUP BY s0.userid, s0.stime, s0.etime
			SETTINGS allow_experimental_join_condition = 1
		) t1
	`
	sql_draw_invites = fmt.Sprintf(sql_draw_invites, userids_filter, userids_filter)
	draw_invites_args := append(userids_args, userids_args...)
	var draw_invites_datas = make(map[string]any)
	err = ck.Select(&draw_invites_datas, sql_draw_invites, draw_invites_args...)
	if err != nil {
		return
	}
	summary.DrawTurnInviteUsersAvg = fmt.Sprintf("%.2f", utils.ToFloat64(draw_invites_datas["turn_invite_users_avg"]))
	summary.DrawTurnInviteUsersPayedAvg = fmt.Sprintf("%.2f", utils.ToFloat64(draw_invites_datas["turn_invite_pay_users_avg"]))
	summary.BonusContribute = utils.ToInt64(draw_invites_datas["bonus_contribute"])
	summary.CashContribute = utils.ToInt64(draw_invites_datas["cash_contribute"])
	summary.WithdrawableContribute = utils.ToInt64(draw_invites_datas["withdrawable_contribute"])

	sql_prize_invites := `
		SELECT AVG(turn_invite_users) turn_invite_users_avg, AVG(turn_invite_pay_users) turn_invite_pay_users_avg FROM (
			SELECT s0.userid, s0.stime, s0.etime, SUM(s1.invited) turn_invite_users, SUM(s1.invited_pay) turn_invite_pay_users
			FROM (
				SELECT userid, toDateTime(turn_stime) stime, toDateTime(turn_etime) etime
				FROM game.col_activity_turn_prize_log FINAL
				WHERE userid IN (%s)
				GROUP BY userid, turn_stime, turn_etime
			) s0 LEFT JOIN (
				SELECT userid, share_superior, ctime, 1 invited, (CASE WHEN money > 0 THEN 1 ELSE 0 END) invited_pay
				FROM game.col_user FINAL
				WHERE share_superior IN (%s) AND share_source = 1
			) s1 ON s1.share_superior = s0.userid AND s1.ctime BETWEEN s0.stime AND s0.etime
			GROUP BY s0.userid, s0.stime, s0.etime
			SETTINGS allow_experimental_join_condition = 1
		) t1
	`
	sql_prize_invites = fmt.Sprintf(sql_prize_invites, userids_filter, userids_filter)
	prize_invites_args := append(userids_args, userids_args...)
	var prize_invites_datas = make(map[string]any)
	err = ck.Select(&prize_invites_datas, sql_prize_invites, prize_invites_args...)
	if err != nil {
		return
	}
	summary.TackPrizeTurnInviteUsersAvg = fmt.Sprintf("%.2f", utils.ToFloat64(prize_invites_datas["turn_invite_users_avg"]))
	summary.TackPrizeTurnInviteUsersPayedAvg = fmt.Sprintf("%.2f", utils.ToFloat64(prize_invites_datas["turn_invite_pay_users_avg"]))

	// 转盘领奖统计
	sql_prizes := `
		SELECT count(*) prize_turns, 
			SUM(CASE WHEN state=1 THEN 1 ELSE 0 END) pass_prize_turns, 
			SUM(CASE WHEN state=1 THEN prize ELSE 0 END) pass_prizes,
			SUM(CASE WHEN state=1 AND prize_type=1 THEN prize ELSE 0 END) pass_prize_bouns,
			SUM(CASE WHEN state=1 AND prize_type=2 THEN prize ELSE 0 END) pass_prize_cash,
			SUM(CASE WHEN state=1 AND prize_type=3 THEN prize ELSE 0 END) pass_prize_withdrawalble
		FROM game.col_activity_turn_prize_log FINAL
		WHERE userid IN (%s)
	`
	sql_prizes = fmt.Sprintf(sql_prizes, userids_filter)
	var prizes_datas = make(map[string]any)
	err = ck.Select(&prizes_datas, sql_prizes, userids_args...)
	if err != nil {
		return
	}
	{
		prize_turns := utils.ToInt64(prizes_datas["prize_turns"])
		pass_prize_turns := utils.ToInt64(prizes_datas["pass_prize_turns"])
		pass_prizes := utils.ToInt64(prizes_datas["pass_prizes"])
		pass_prize_bouns := utils.ToInt64(prizes_datas["pass_prize_bouns"])
		pass_prize_cash := utils.ToInt64(prizes_datas["pass_prize_cash"])
		pass_prize_withdrawalble := utils.ToInt64(prizes_datas["pass_prize_withdrawalble"])

		summary.TackPrizeTurns = prize_turns
		summary.TackPrizePassTurns = pass_prize_turns
		summary.TackPrizePassScore = fmt.Sprintf("%.2f", Chip2Float(pass_prizes))
		summary.TackPrizePassBonus = fmt.Sprintf("%.2f(%.2f%%)", Chip2Float(pass_prize_bouns), ComputeFloat(pass_prize_bouns, pass_prizes))
		summary.TackPrizePassCash = fmt.Sprintf("%.2f(%.2f%%)", Chip2Float(pass_prize_cash), ComputeFloat(pass_prize_cash, pass_prizes))
		summary.TackPrizePassWithdrawable = fmt.Sprintf("%.2f(%.2f%%)", Chip2Float(pass_prize_withdrawalble), ComputeFloat(pass_prize_withdrawalble, pass_prizes))
		summary.ScoreRoi = fmt.Sprintf("%.2f", ComputeFloat(summary.InviteUsersContribute0, pass_prizes))
		summary.BonusRoi = fmt.Sprintf("%.2f", ComputeFloat(summary.BonusContribute, pass_prize_bouns))
		summary.CashRoi = fmt.Sprintf("%.2f", ComputeFloat(summary.CashContribute, pass_prize_cash))
		summary.WithdrawableRoi = fmt.Sprintf("%.2f", ComputeFloat(summary.WithdrawableContribute, pass_prize_withdrawalble))
	}

	summary.TurnRate = fmt.Sprintf("%.2f%%", ComputeFloat(summary.DrawTurns, summary.Turns)*100)
	summary.TackPrizeRate = fmt.Sprintf("%.2f%%", ComputeFloat(summary.TackPrizeTurns, summary.DrawTurns)*100)
	summary.TackPrizePassRate = fmt.Sprintf("%.2f%%", ComputeFloat(summary.TackPrizePassTurns, summary.TackPrizeTurns)*100)
	summary.InviteUsersPayedRate = fmt.Sprintf("%.2f%%", ComputeFloat(summary.InviteUsersPayed, summary.InviteUsers)*100)
	return
}

// GetActivityTurnAudits 转盘活动审核
func (s *activityService) GetActivityTurnAudits(page, pageSize int, userid string, begin, end time.Time, begin2, end2 *time.Time, utypes []int, channelIds []string, auditStatus, prizeTypes []int, auditUserIds []string) (total int64, stats []*entity.ActivityTurnStatAudit, err error) {
	var sql_prizes_filter string
	var sql_prizes_args = []any{begin.Unix(), end.Unix()}
	if begin2 != nil {
		sql_prizes_filter += " AND t1.stime >= ?"
		sql_prizes_args = append(sql_prizes_args, (*begin2).Unix())
	}
	if end2 != nil {
		sql_prizes_filter += " AND t1.stime <= ?"
		sql_prizes_args = append(sql_prizes_args, (*end2).Unix())
	}
	if len(auditStatus) > 0 {
		var audit_conds []string
		for _, status := range auditStatus {
			switch status {
			case 1: // 等待人审
				audit_conds = append(audit_conds, "(t1.state = 0)")
			case 2: // 人审拒绝
				audit_conds = append(audit_conds, "(t1.state = 2)")
			case 3: // 人审通过
				audit_conds = append(audit_conds, "(t1.state = 1 AND t1.stype = 2)")
			case 4: // 机审通过
				audit_conds = append(audit_conds, "(t1.state = 1 AND t1.stype = 1)")
			}
		}
		if len(audit_conds) > 0 {
			cond_fmt := " AND (%s)"
			if len(audit_conds) == 1 {
				cond_fmt = " AND %s"
			}
			sql_prizes_filter += fmt.Sprintf(cond_fmt, strings.Join(audit_conds, " OR "))
		}
	}
	if len(prizeTypes) > 0 {
		sql_prizes_filter += " AND t1.prize_type IN ?"
		sql_prizes_args = append(sql_prizes_args, prizeTypes)
	}
	if len(auditUserIds) > 0 {
		sql_prizes_filter += " AND t1.audit_userid IN ?"
		sql_prizes_args = append(sql_prizes_args, auditUserIds)
	}
	if userid != "" {
		sql_prizes_filter += " AND t0.userid = ?"
		sql_prizes_args = append(sql_prizes_args, userid)
	}

	var utypeFilters []string
	if len(utypes) > 0 {
		for _, utype := range utypes {
			// state 1.新手 2.正常 3.平民 4.泡沫
			state, sMoney, eMoney := 2, -1, -1
			switch utype {
			case 0: // 新手
				state = 1
				sMoney, eMoney = 0, 0
			case 1: // 平民
				state = -1
				sMoney, eMoney = 0, 0
			case 2: // 普充
				sMoney, eMoney = 1, 99999
			case 3: // 小R
				sMoney, eMoney = 100000, 499999
			case 4: // 中R
				sMoney, eMoney = 500000, 999999
			case 5: // 大R
				sMoney, eMoney = 1000000, 19999999
			case 6: // 超大R
				sMoney, eMoney = 20000000, -1
			}
			var utypeQs []string
			if state != -1 {
				utypeQs = append(utypeQs, fmt.Sprintf("t0.state = %d", state))
			}
			if sMoney != -1 {
				utypeQs = append(utypeQs, fmt.Sprintf("t0.money >= %d", sMoney))
			}
			if eMoney != -1 {
				utypeQs = append(utypeQs, fmt.Sprintf("t0.money <= %d", eMoney))
			}
			if len(utypeQs) > 0 {
				utypeFilters = append(utypeFilters, fmt.Sprintf("(%s)", strings.Join(utypeQs, " AND ")))
			}
		}
	}
	if len(utypeFilters) > 0 {
		utypeQFmt := " AND (%s)"
		if len(utypeFilters) == 1 {
			utypeQFmt = " AND %s"
		}
		sql_prizes_filter += fmt.Sprintf(utypeQFmt, strings.Join(utypeFilters, " OR "))
	}
	if len(channelIds) > 0 {
		sql_prizes_filter += " AND t0.ad__bundle_id in ?"
		sql_prizes_args = append(sql_prizes_args, channelIds)
	}

	sql_prizes_count := `
		SELECT count(*) total
		FROM game.col_user t0 FINAL
		JOIN game.col_activity_turn_prize_log t1 FINAL ON t1.userid = t0.userid
		WHERE t1.ctime BETWEEN ? AND ? %s
	`
	sql_prizes_count = fmt.Sprintf(sql_prizes_count, sql_prizes_filter)
	err = ck.Select(&total, sql_prizes_count, sql_prizes_args...)
	if err != nil {
		return
	}
	if total == 0 {
		return
	}

	sql_prizes := `
		SELECT t0.userid userid, t0.state state, t0.ctime ctime, t0.login_time, t0.vip_lv, t0.ad__bundle_id, t0.regist_area, t0.money,
			t1.id, t1.prize, t1.prize_type, t1.state audit_state, t1.stype, t1.turn_stime, t1.turn_etime, t1.stime, t1.remark, t1.ctime submit_time, t1.audit_user
		FROM game.col_user t0 FINAL
		JOIN game.col_activity_turn_prize_log t1 FINAL ON t1.userid = t0.userid
		WHERE t1.ctime BETWEEN ? AND ? %s
		ORDER BY t1.ctime DESC
		LIMIT ?, ?
	`
	offset, limit := PageCalc(page, pageSize)
	var prize_datas []map[string]any
	sql_prizes = fmt.Sprintf(sql_prizes, sql_prizes_filter)
	prizes_args := append(sql_prizes_args, offset, limit)
	err = ck.Select(&prize_datas, sql_prizes, prizes_args...)
	if err != nil {
		return
	}

	channelsMap, err := ChannelService.GetChannelsMap()
	if err != nil {
		return
	}
	now := NowTime()

	var idStats = make(map[string]*entity.ActivityTurnStatAudit)
	var useridStats = make(map[string][]*entity.ActivityTurnStatAudit)
	var userids []string
	var auditIds []string
	for _, data := range prize_datas {
		_ = data
		id := data["id"].(string)
		userid := data["userid"].(string)
		ad__bundle_id := data["ad__bundle_id"].(string)
		ctime := data["ctime"].(time.Time)
		login_time := data["login_time"].(time.Time)
		regist_area := utils.ToInt64(data["regist_area"])
		vip_lv := utils.ToInt64(data["vip_lv"])
		state := utils.ToInt64(data["state"])
		money := utils.ToInt64(data["money"])
		// first_draw_time := utils.ToInt64(data["first_draw_time"])
		// give_times := utils.ToInt64(data["give_times"])

		registAreaF, _ := utils.CaseWhen3(regist_area, 0, "A类", 1, "B类", 2, "C类")
		ctimeF := ctime.Format(utils.FORMAT)
		loginTimeF := login_time.Format(utils.FORMAT)
		liveDays := int64(math.Ceil(login_time.Sub(ctime).Hours() / 24))
		loseDays := int64(math.Floor(now.Sub(login_time).Hours() / 24))
		var channelClass string
		if c, ok := channelsMap[ad__bundle_id]; ok {
			channelClass = c.ClassName
		}
		// var turns int64
		// if first_draw_time > 0 {
		// 	turns = int64(math.Ceil((float64(now.Unix()-first_draw_time) / 60 / 60 / 72)))
		// }

		var userType string
		{
			if money == 0 {
				if state == 1 {
					userType = "新手"
				} else {
					userType = "平民"
				}
			} else if money >= 1 && money <= 99999 {
				userType = "普充"
			} else if money >= 100000 && money <= 499999 {
				userType = "小R"
			} else if money >= 500000 && money <= 999999 {
				userType = "中R"
			} else if money >= 1000000 && money <= 19999999 {
				userType = "大R"
			} else if money >= 20000000 {
				userType = "超大R"
			}
		}

		prize := utils.ToInt64(data["prize"])
		prize_type := utils.ToInt64(data["prize_type"])
		audit_state := utils.ToInt64(data["audit_state"])
		stype := utils.ToInt64(data["stype"])
		turn_stime := utils.ToInt64(data["turn_stime"])
		// turn_etime := utils.ToInt64(data["turn_etime"])
		stime := utils.ToInt64(data["stime"])
		submit_time := utils.ToInt64(data["submit_time"])
		remark := data["remark"].(string)
		audit_user := data["audit_user"].(string)
		prizeTypeF, _ := utils.CaseWhen3(prize_type, 1, "bonus", 2, "cash", 3, "withdrawable")
		var auditStateF string
		{
			// "有4种状态：1.机审直接通过 2.机审拒绝后，等待人审 3.机审拒绝后，人审拒绝 4.机审拒绝后，人审通过"
			if stype == 0 {
				auditStateF = "等待机审"
			} else {
				if stype == 1 { // 机审
					if audit_state == 1 {
						auditStateF = "机审直接通过"
					} else {
						auditStateF = "机审拒绝后，等待人审"
					}
					audit_user = "机审"
				} else if stype == 2 {
					if audit_state == 1 {
						auditStateF = "机审拒绝后，人审通过"
					} else {
						auditStateF = "机审拒绝后，人审拒绝"
					}
				}
			}
		}
		submitTime := time.Unix(submit_time, 0).In(location).Format(utils.FORMAT)
		var auditTime string
		if stime > 0 {
			auditTime = time.Unix(stime, 0).In(location).Format(utils.FORMAT)
		}

		stat := &entity.ActivityTurnStatAudit{
			Id:           id,
			Userid:       userid,
			ChannelClass: channelClass,
			ChannelId:    ad__bundle_id,
			RegistArea:   registAreaF,
			Ctime:        ctimeF,
			LoginTime:    loginTimeF,
			LiveDays:     fmt.Sprint(liveDays),
			LoseDays:     fmt.Sprint(loseDays),
			VipLv:        fmt.Sprint(vip_lv),
			UserType:     userType,
			Prize:        fmt.Sprintf("%.2f", Chip2Float(prize)),
			PrizeType:    prizeTypeF,
			Status:       auditStateF,
			State:        audit_state,
			Remark:       remark,
			SubmitTime:   submitTime,
			AuditTime:    auditTime,
			AuditUser:    audit_user,
			TurnStime:    turn_stime,
		}

		// 标记信息
		if stat.Remark != "" {
			stat.Remarks = strings.Split(stat.Remark, ";")
		}
		stats = append(stats, stat)

		idStats[stat.Id] = stat
		useridStats[stat.Userid] = append(useridStats[stat.Userid], stat)
		userids = append(userids, userid)
		auditIds = append(auditIds, stat.Id)
	}
	if len(auditIds) == 0 {
		return
	}

	sql_invites := `
		SELECT t1.id, t1.userid, groupArray(t0.userid) share_userids, count(*) share_users, 
			SUM(CASE WHEN t0.money > 0 THEN 1 ELSE 0 END) share_pay_users,
			SUM(t0.money) share_users_pays
		FROM game.col_user t0 FINAL
		JOIN (
			SELECT id, userid, toDateTime(turn_stime) stime, toDateTime(ctime) etime
			FROM game.col_activity_turn_prize_log t1 FINAL
			WHERE id IN ?
		) t1
		ON t0.share_superior = t1.userid AND t0.ctime BETWEEN t1.stime AND t1.etime
		GROUP BY t1.id, t1.userid
		SETTINGS allow_experimental_join_condition = 1
	`
	var invites_datas []map[string]any
	if err = ck.Select(&invites_datas, sql_invites, auditIds); err != nil {
		return
	}
	for _, data := range invites_datas {
		id := data["id"].(string)
		if stat, ok := idStats[id]; ok {
			share_users := utils.ToInt64(data["share_users"])
			share_pay_users := utils.ToInt64(data["share_pay_users"])
			share_users_pays := utils.ToInt64(data["share_users_pays"])
			stat.TurnInvites = share_users
			stat.TurnInvitesPayed = share_pay_users
			stat.TurnInvitesPays = fmt.Sprintf("%.2f", Chip2Float(share_users_pays))

			if share_userids, ok := data["share_userids"].([]string); ok {
				stat.TurnUserids = share_userids
			}
		}
	}
	// 总打码,充值,提现
	sql_bets := `
		SELECT userid, SUM(bets) bets FROM (
			SELECT userid, SUM(bet_amount) bets FROM game.col_detail FINAL
			WHERE robot = 0 AND userid IN ?
			GROUP BY userid 
			
			UNION ALL
			
			SELECT s1.user_id userid, SUM(s1.amount_sum) bets
			FROM (
				SELECT round_id, user_id, game_id, SUM(amount) amount_sum, MIN(ctime) begin_time 
				FROM game.col_nsq_log_external_bet FINAL 
				WHERE amount != 0 AND user_id IN ?
				GROUP BY round_id, user_id, game_id
			) s1
			GROUP BY s1.user_id
		) t1
		GROUP BY userid
	`
	var bets_datas []map[string]any
	if err = ck.Select(&bets_datas, sql_bets, userids, userids); err != nil {
		return
	}
	for _, data := range bets_datas {
		userid := data["userid"].(string)
		if stats, ok := useridStats[userid]; ok {
			for _, stat := range stats {
				stat.UserBets = fmt.Sprintf("%.2f", Chip2Float(utils.ToInt64(data["bets"])))
			}
		}
	}

	// 充值提现查询
	sql_pays := `
		SELECT userid, SUM(amount) pays
		FROM game.col_trade_record FINAL
		WHERE order_status = 4 AND userid IN ?
		GROUP BY userid
	`
	var pays_datas []map[string]any
	err = ck.Select(&pays_datas, sql_pays, userids)
	if err != nil {
		return
	}
	for _, data := range pays_datas {
		userid := data["userid"].(string)
		if stats, ok := useridStats[userid]; ok {
			for _, stat := range stats {
				stat.Pays = fmt.Sprintf("%.2f", Chip2Float(utils.ToInt64(data["pays"])))
			}
		}
	}

	sql_withdraws := `
		SELECT userid, SUM(amount + commission) withdraws
		FROM game.col_withdraw_record FINAL
		WHERE order_status = 2 AND userid IN ?
		GROUP BY userid
	`
	var withdraws_datas []map[string]any
	err = ck.Select(&withdraws_datas, sql_withdraws, userids)
	if err != nil {
		return
	}
	for _, data := range withdraws_datas {
		userid := data["userid"].(string)
		if stats, ok := useridStats[userid]; ok {
			for _, stat := range stats {
				stat.Withdraws = fmt.Sprintf("%.2f", Chip2Float(utils.ToInt64(data["withdraws"])))
			}
		}
	}

	// 当前是总的第几轮
	sql_turn_no := `
		SELECT userid, turn_stime, ROW_NUMBER() OVER (partition by userid ORDER BY turn_stime ASC) AS turn_no FROM (
			SELECT userid, turn_stime
			FROM game.col_activity_turn_draw_log FINAL
			WHERE userid IN ?
			group by userid, turn_stime
		) t1 
	`
	var turn_no_datas []map[string]any
	err = ck.Select(&turn_no_datas, sql_turn_no, userids)
	if err != nil {
		return
	}
	var uidStimeMap = make(map[string]*entity.ActivityTurnStatAudit)
	for _, stat := range stats {
		uidStimeMap[fmt.Sprintf("%s-%d", stat.Userid, stat.TurnStime)] = stat
	}
	for _, data := range turn_no_datas {
		userid := data["userid"].(string)
		turn_stime := utils.ToInt64(data["turn_stime"])
		turn_no := fmt.Sprint(data["turn_no"])
		key := fmt.Sprintf("%s-%d", userid, turn_stime)
		if stat, ok := uidStimeMap[key]; ok {
			stat.TurnNo = turn_no
		}
	}

	// 当前是第几次申请
	sql_prize_no := `
		SELECT id, userid, turn_no FROM (
			SELECT id, userid, ROW_NUMBER() OVER (partition by userid ORDER BY ctime ASC) AS turn_no
			FROM game.col_activity_turn_prize_log FINAL
			WHERE userid IN ?
		) t1 
		WHERE id IN ?
	`
	var prize_no_datas []map[string]any
	err = ck.Select(&prize_no_datas, sql_prize_no, userids, auditIds)
	if err != nil {
		return
	}
	for _, data := range prize_no_datas {
		id := data["id"].(string)
		if stat, ok := idStats[id]; ok {
			stat.TackPrizeNo = fmt.Sprint(utils.ToInt64(data["turn_no"]))
		}
	}

	summary, err := s.activityTurnAuditsSummary(sql_prizes_filter, sql_prizes_args)
	if err != nil {
		return
	}
	stats = append([]*entity.ActivityTurnStatAudit{summary}, stats...)

	return
}

func (s *activityService) activityTurnAuditsSummary(sql_prizes_filter string, sql_prizes_args []any) (summary *entity.ActivityTurnStatAudit, err error) {
	summary = &entity.ActivityTurnStatAudit{
		Userid:       "汇总",
		ChannelClass: "--",
		ChannelId:    "--",
		RegistArea:   "--",
		Ctime:        "--",
		LoginTime:    "--",
		LiveDays:     "--",
		LoseDays:     "--",
		VipLv:        "--",
		UserType:     "--",
		TurnNo:       "--",
		TackPrizeNo:  "--",
		Status:       "--",
		Remark:       "--",
		SubmitTime:   "--",
		AuditTime:    "--",
		UserBets:     "--",
	}

	where_userids := `
		SELECT DISTINCT t1.userid
		FROM game.col_user t0 FINAL
		JOIN game.col_activity_turn_prize_log t1 FINAL ON t1.userid = t0.userid
		WHERE t1.ctime BETWEEN ? AND ? %s
	`
	where_userids = fmt.Sprintf(where_userids, sql_prizes_filter)

	where_auditIds := `
		SELECT t1.id
		FROM game.col_user t0 FINAL
		JOIN game.col_activity_turn_prize_log t1 FINAL ON t1.userid = t0.userid
		WHERE t1.ctime BETWEEN ? AND ? %s
	`
	where_auditIds = fmt.Sprintf(where_auditIds, sql_prizes_filter)

	sql_prizes := `
		SELECT SUM(t1.prize) prize_sum, groupArray(DISTINCT t1.prize_type) prize_types, groupArray(DISTINCT t1.audit_user) audit_users
		FROM game.col_user t0 FINAL
		JOIN game.col_activity_turn_prize_log t1 FINAL ON t1.userid = t0.userid
		WHERE t1.ctime BETWEEN ? AND ? %s
	`
	sql_prizes = fmt.Sprintf(sql_prizes, sql_prizes_filter)
	var prizes_data = make(map[string]any)
	if err = ck.Select(&prizes_data, sql_prizes, sql_prizes_args...); err != nil {
		return
	}
	{
		summary.Prize = fmt.Sprintf("%.2f", Chip2Float(utils.ToInt64(prizes_data["prize_sum"])))
		if prize_types, ok := prizes_data["prize_types"].([]int32); ok {
			var typeFs []string
			for _, prize_type := range prize_types {
				prizeTypeF, _ := utils.CaseWhen3(prize_type, 1, "bonus", 2, "cash", 3, "withdrawable")
				typeFs = append(typeFs, prizeTypeF)
			}
			summary.PrizeType = strings.Join(typeFs, ",")
		}
		if audit_users, ok := prizes_data["audit_users"].([]string); ok {
			var users []string
			for _, user := range audit_users {
				if user != "" {
					users = append(users, user)
				}
			}
			summary.AuditUser = strings.Join(users, ",")
		}
	}

	sql_invites := `
		SELECT count(*) share_users, 
			SUM(CASE WHEN t0.money > 0 THEN 1 ELSE 0 END) share_pay_users,
			SUM(t0.money) share_users_pays
		FROM game.col_user t0 FINAL
		JOIN (
			SELECT id, userid, toDateTime(turn_stime) stime, toDateTime(turn_etime) etime
			FROM game.col_activity_turn_prize_log t1 FINAL
			WHERE id IN (%s)
		) t1
		ON t0.share_superior = t1.userid AND t0.ctime BETWEEN t1.stime AND t1.etime
		SETTINGS allow_experimental_join_condition = 1
	`
	sql_invites = fmt.Sprintf(sql_invites, where_auditIds)
	var invites_datas = make(map[string]any)
	if err = ck.Select(&invites_datas, sql_invites, sql_prizes_args...); err != nil {
		return
	}
	{
		share_users := utils.ToInt64(invites_datas["share_users"])
		share_pay_users := utils.ToInt64(invites_datas["share_pay_users"])
		share_users_pays := utils.ToInt64(invites_datas["share_users_pays"])
		summary.TurnInvites = share_users
		summary.TurnInvitesPayed = share_pay_users
		summary.TurnInvitesPays = fmt.Sprintf("%.2f", Chip2Float(share_users_pays))
	}

	// 总打码,充值,提现
	sql_bets := `
		SELECT SUM(bets) bets FROM (
			SELECT SUM(bet_amount) bets FROM game.col_detail FINAL
			WHERE robot = 0 AND userid IN (%s)

			UNION ALL

			SELECT SUM(amount) bets
			FROM game.col_nsq_log_external_bet FINAL
			WHERE amount != 0 AND user_id IN (%s)
		) t1
	`
	sql_bets = fmt.Sprintf(sql_bets, where_userids, where_userids)
	sql_bets_args := append(sql_prizes_args, sql_prizes_args...)
	var bets_data = make(map[string]any)
	if err = ck.Select(&bets_data, sql_bets, sql_bets_args...); err != nil {
		return
	}
	summary.UserBets = fmt.Sprintf("%.2f", Chip2Float(utils.ToInt64(bets_data["bets"])))

	// 充值提现查询
	sql_pays := `
		SELECT SUM(amount) pays
		FROM game.col_trade_record FINAL
		WHERE order_status = 4 AND userid IN (%s)
	`
	sql_pays = fmt.Sprintf(sql_pays, where_userids)
	var pays_datas = make(map[string]any)
	err = ck.Select(&pays_datas, sql_pays, sql_prizes_args...)
	if err != nil {
		return
	}
	summary.Pays = fmt.Sprintf("%.2f", Chip2Float(utils.ToInt64(pays_datas["pays"])))

	sql_withdraws := `
		SELECT SUM(amount + commission) withdraws
		FROM game.col_withdraw_record FINAL
		WHERE order_status = 2 AND userid IN (%s)
	`
	sql_withdraws = fmt.Sprintf(sql_withdraws, where_userids)
	var withdraws_datas = make(map[string]any)
	err = ck.Select(&withdraws_datas, sql_withdraws, sql_prizes_args...)
	if err != nil {
		return
	}
	summary.Withdraws = fmt.Sprintf("%.2f", Chip2Float(utils.ToInt64(withdraws_datas["withdraws"])))

	return
}

func (s *activityService) TurnPrizeTag(name, id, tag string) (err error) {
	prize := new(entity.ActivityTurnPrizeLog)
	GetByQ(ActivityTurnPrizeLogs, bson.M{"_id": id}, prize)
	if prize.Id == "" {
		return errors.New("奖励不存在")
	}

	now := NowTime()
	tag = fmt.Sprintf("%s %s: %s", now.Format(utils.FORMAT2), name, tag)
	if prize.Remark != "" {
		prize.Remark += ";" + tag
	} else {
		prize.Remark = tag
	}
	if !Update(ActivityTurnPrizeLogs, bson.M{"_id": id}, bson.M{"$set": bson.M{"remark": prize.Remark}}) {
		return errors.New("remark 更新失败")
	}
	err = ck.DB().Model(&ck.ActivityTurnPrizeLog{}).Where("id", prize.Id).Update("remark", prize.Remark).Error
	return
}

// 礼包码
func (s *activityService) GiftPackCodeList(page, pageSize int, m bson.M, status int, ctimeAsc bool) (
	total int, gifts []*entity.GiftPackCode,
	scoreSum, scoreReceivedSum string, err error,
) {
	if status != 0 {
		if utils.SliceIn(status, 1, 2) {
			m["status"] = status - 1
		} else {
			if status == 3 {
				nowSec := time.Now().Unix()
				m["time_close"] = bson.M{"$lt": nowSec}
			}
		}
	}

	total = Count(GiftPackCodes, m)
	if total == 0 {
		return
	}

	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", ctimeAsc)
	err = GiftPackCodes.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&gifts)
	if err != nil {
		return
	}
	s.mappingGiftPackCodes(gifts)

	pipe := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":                nil,
			"score_sum":          bson.M{"$sum": "$score"},
			"score_received_sum": bson.M{"$sum": "$score_received"},
		}},
	}
	stats := []bson.M{}
	err = GiftPackCodes.Pipe(pipe).All(&stats)
	if err != nil {
		return
	}
	if len(stats) > 0 {
		score_sum := utils.ToInt64(stats[0]["score_sum"])
		score_received_sum := utils.ToInt64(stats[0]["score_received_sum"])
		scoreSum = fmt.Sprintf("%.2f", Chip2Float(score_sum))
		scoreReceivedSum = fmt.Sprintf("%.2f", Chip2Float(score_received_sum))
	}
	return
}

func (s *activityService) mappingGiftPackCodes(gifts []*entity.GiftPackCode) {
	for _, gift := range gifts {

		gift.FCtime = time.Unix(gift.Ctime, 0).In(Location()).Format(utils.FORMAT)
		gift.FStatus = utils.CaseElse(gift.Status == 1, "开", "关")
		gift.FCode = gift.Code
		gift.FGiftType = utils.CaseElse(gift.GiftType == 1, "随机码", "等额码")
		gift.FScoreType, _ = utils.CaseWhen3(gift.ScoreType, 1, "bonus", 2, "cash", 3, "withdrawable")
		gift.FScore = fmt.Sprintf("%.2f", Chip2Float(gift.Score))
		gift.FScoreReceived = fmt.Sprintf("%.2f", Chip2Float(gift.ScoreReceived))
		gift.FFinishTime = utils.CaseElse(gift.FinishTime == 0, "", time.Unix(gift.FinishTime, 0).In(Location()).Format(utils.FORMAT))
		gift.FPieceMin = fmt.Sprintf("%.2f", Chip2Float(gift.PieceMin))
		gift.FPieceMax = fmt.Sprintf("%.2f", Chip2Float(gift.PieceMax))
		// gift.FTimeOpenClose = fmt.Sprintf("%s-%s", ,)
		gift.FTimeOpen = time.Unix(gift.TimeOpen, 0).In(Location()).Format(utils.FORMAT2)
		gift.FTimeClose = time.Unix(gift.TimeClose, 0).In(Location()).Format(utils.FORMAT2)
		nowSec := time.Now().Unix()
		if nowSec < gift.TimeOpen {
			gift.FStatus2 = "未开始"
		} else if nowSec > gift.TimeClose {
			gift.FStatus2 = "已过期"
		}
		gift.FCuser = gift.Cuser

		var fUtypes, fUtypes2 []string
		for _, utype := range gift.Utypes {
			var fUtype string
			switch utype {
			case 0:
				fUtype = "A类"
			case 1:
				fUtype = "B类"
			case 2:
				fUtype = "C类"
			case 10:
				fUtype = "新手"
			case 11:
				fUtype = "平民"
			case 12:
				fUtype = "普R"
			case 13:
				fUtype = "小R"
			case 14:
				fUtype = "中R"
			case 15:
				fUtype = "大R"
			case 16:
				fUtype = "超大R"
			}
			fUtypes = append(fUtypes, fUtype)
			fUtypes2 = append(fUtypes2, strconv.Itoa(int(utype)))
		}
		gift.FUtypes = strings.Join(fUtypes, ",")
		gift.FUtypes2 = strings.Join(fUtypes2, ",")
	}
}

// 礼包码
func (s *activityService) GetGiftPackCode(code string) (gift *entity.GiftPackCode) {
	gift = new(entity.GiftPackCode)
	GetByQ(GiftPackCodes, bson.M{"_id": code}, gift)
	return
}

// 添加或修改支付渠道
func (s *activityService) AddOrUpdateGiftPackCode(gift *entity.GiftPackCode, add bool) error {
	// info := new(entity.GiftPackCode)
	// GetByQ(GiftPackCodes, bson.M{"_id": gift.Code}, info)

	if add {
		gift.Ctime = time.Now().Unix()
		if !Insert(GiftPackCodes, gift) {
			return errors.New("写入失败:" + gift.Code)
		}
		return nil
	} else {
		m := bson.M{
			"status":     gift.Status,
			"score_type": gift.ScoreType,
			"score":      gift.Score,
			"pieces":     gift.Pieces,
			"piece_min":  gift.PieceMin,
			"piece_max":  gift.PieceMax,
			"time_open":  gift.TimeOpen,
			"time_close": gift.TimeClose,
			"utypes":     gift.Utypes,
			"remark":     gift.Remark,
		}
		if Update(GiftPackCodes, bson.M{"_id": gift.Code}, bson.M{"$set": m}) {
			return nil
		}
		return errors.New("更新失败")
	}
}

// GetActivityTurnDates 波动返水汇总
// registArea: 1.A 2.B 3.C
func (s *activityService) GetActivityVolatilitySubsidyDates(begin, end time.Time, registArea int, packageIds []string) (stats []*entity.VolatilitySubsidyDate, err error) {
	where_u_sql := ""
	var where_u_args []any

	if registArea > 0 {
		where_u_sql += " and s0.regist_area = ?"
		where_u_args = append(where_u_args, registArea-1)
	}

	if len(packageIds) > 0 {
		where_u_sql += " and s0.ad__bundle_id IN ?"
		where_u_args = append(where_u_args, packageIds)
	}

	// 新老顾客日活
	var loginDatas []map[string]any
	login_args := append([]any{begin, end}, where_u_args...)
	err = ck.Select(&loginDatas, fmt.Sprintf(`
		select datestr, count(*) players
		from (
			select toYYYYMMDD(s1.login_time) datestr
			from game.col_log_login s1 final
			join game.col_user s0 final on s1.userid = s0.userid
			where s1.login_time between ? and ? %s
			group by datestr, s1.userid
		) t0
		group by datestr
		order by datestr desc
	`, where_u_sql), login_args...)
	if err != nil {
		return
	}
	var dateStats = make(map[string]*entity.VolatilitySubsidyDate)
	for _, data := range loginDatas {
		date := data["datestr"].(uint32)
		players := utils.ToInt64(data["players"])
		datestr := fmt.Sprint(date)

		stat := &entity.VolatilitySubsidyDate{
			SDate:        datestr[0:4] + "-" + datestr[4:6] + "-" + datestr[6:8],
			LoginedUsers: players,
			SubsidyPools: make(map[int32]int64),
		}
		stats = append(stats, stat)
		dateStats[stat.SDate] = stat
	}

	// 首冲用户数
	sql_first_pay := `
		select first_pay_date, count(*) first_pay_users
		from (
			select userid, min(ctime) first_pay_time, toYYYYMMDD(min(ctime)) first_pay_date
			from game.col_trade_record s1 final where userid in (
				select s1.userid
				from game.col_user s0 final
				join game.col_trade_record s1 final on s1.userid = s0.userid
				where s1.ctime between ? AND ? and s1.order_status = 4 %s
			) and order_status = 4
			group by userid
			having first_pay_time between ? AND ?
		) t0
		group by first_pay_date
	`
	var firstPayDatas []map[string]any
	first_pay_args := append([]any{begin, end}, where_u_args...)
	first_pay_args = append(first_pay_args, begin, end)
	err = ck.Select(&firstPayDatas, fmt.Sprintf(sql_first_pay, where_u_sql), first_pay_args...)
	if err != nil {
		return
	}
	for _, data := range firstPayDatas {
		first_pay_date := data["first_pay_date"].(uint32)
		first_pay_users := utils.ToInt64(data["first_pay_users"])

		datestr := fmt.Sprint(first_pay_date)
		sdate := datestr[0:4] + "-" + datestr[4:6] + "-" + datestr[6:8]

		if stat, ok := dateStats[sdate]; ok {
			stat.FirstPayUsers = first_pay_users
		}
	}

	// 领补贴统计
	sql_subsidy := `
		select cdate, count(*) subsidy_users,	
			SUM(case when first then 1 else 0 end) first_subsidy_users,
			SUM(subsidy_times) subsidy_times,
			SUM(case when cdate = toYYYYMMDD(toDateTime(first_charge_time/1000)) then 1 else 0 end) first_pay_subsidy_users,
			SUM(subsidys) subsidy_sum,
			groupArray((rank,subsidys,userid)) top_subsidy_users
		from (
			select toYYYYMMDD(toDateTime(s1.ctime/1000)) cdate, s1.userid userid, max(first) first,
				max(s0.first_charge_time) first_charge_time, count(*) subsidy_times, SUM(subsidy) subsidys,
				rank() over (PARTITION BY cdate ORDER BY subsidys desc) as rank
			from game.col_user s0 final
			join game.col_activity_volatility_subsidy s1 final on s1.userid = s0.userid
  			where s1.ctime between ? and ? %s
			group by cdate, s1.userid
			order by cdate desc, rank asc
		) t0
		group by cdate
		order by cdate desc
	`
	_ = `
		select sdate, count(*) subsidy_users,	
			SUM(case when first then 1 else 0 end) first_subsidy_users,
			SUM(subsidy_times) subsidy_times,
			SUM(subsidys) subsidy_sum,
			topK(3)((subsidys, userid)) top_subsidy_users
		from (
			select s1.sdate, s1.userid userid, max(first) first,
				 count(*) subsidy_times, SUM(subsidy) subsidys
			from game.col_activity_volatility_subsidy s1 final
			where s1.ctime between ? and ? %s
			group by s1.sdate, s1.userid
		) t0
		group by sdate
		order by sdate desc
	`
	var subsidyDatas []map[string]any
	subsidy_args := append([]any{begin.UnixMilli(), end.UnixMilli()}, where_u_args...)
	err = ck.Select(&subsidyDatas, fmt.Sprintf(sql_subsidy, where_u_sql), subsidy_args...)
	if err != nil {
		return
	}
	for _, data := range subsidyDatas {
		cdate := data["cdate"].(uint32)
		datestr := fmt.Sprint(cdate)
		subsidy_users := utils.ToInt64(data["subsidy_users"])
		first_subsidy_users := utils.ToInt64(data["first_subsidy_users"])
		subsidy_times := utils.ToInt64(data["subsidy_times"])
		first_pay_subsidy_users := utils.ToInt64(data["first_pay_subsidy_users"])
		subsidy_sum := utils.ToInt64(data["subsidy_sum"])
		top_subsidy_users := data["top_subsidy_users"].([][]interface{})

		sdate := datestr[0:4] + "-" + datestr[4:6] + "-" + datestr[6:8]
		stat, ok := dateStats[sdate]
		// if !ok {
		// 	ok = true
		// 	stat = &entity.VolatilitySubsidyDate{SDate: sdate, SubsidyPools: make(map[int32]int64)}
		// 	stats = append(stats, stat)
		// 	dateStats[stat.SDate] = stat
		// }
		if ok {
			stat.SubsidyUsers = subsidy_users
			stat.FirstSubsidyUsers = first_subsidy_users
			stat.SubsidyTimes = subsidy_times
			stat.FirstPaySubsidyUsers = first_pay_subsidy_users
			stat.SubsidyAmount = fmt.Sprintf("%.2f", Chip2Float(subsidy_sum))

			for _, top_user := range top_subsidy_users {
				rank := utils.ToInt64(top_user[0])
				subsidys := utils.ToInt64(top_user[1])
				userid := top_user[2].(string)
				if rank > 3 {
					continue
				}
				stat.SubsidyTop3 = append(stat.SubsidyTop3, fmt.Sprintf("%s: %.2f", userid, Chip2Float(subsidys)))
				stat.TopSubsidy3 = append(stat.TopSubsidy3, subsidys)
			}
			stat.SubsidyAmount0 = subsidy_sum
			stat.FirstSubsidyRate = fmt.Sprintf("%.2f%%", ComputeFloat(first_subsidy_users, subsidy_users)*100)
			stat.SubsidyAvg = fmt.Sprintf("%.2f", ComputeFloat(subsidy_times, subsidy_users))
			stat.FirstPaySubsidyRate = fmt.Sprintf("%.2f%%", ComputeFloat(first_pay_subsidy_users, stat.FirstPayUsers)*100)
			stat.FirstPaySubsidyUsersRate = fmt.Sprintf("%.2f%%", ComputeFloat(first_pay_subsidy_users, subsidy_users)*100)
			stat.SubsidyAmountAvg = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(subsidy_sum, subsidy_users)))
		}
	}

	// 首次补贴用户次日留存率
	var scbt_lcl_args = []any{begin.AddDate(0, 0, 1), end.AddDate(0, 0, 7), begin.UnixMilli(), end.UnixMilli()}
	scbt_lcl_args = append(scbt_lcl_args, where_u_args...)
	var scbt_lcl_stats []map[string]any
	err = ck.Select(&scbt_lcl_stats, fmt.Sprintf(`
		select t2.cdate cdate, 
			count(case when t1.cdate = t2.cdate then t1.userid end) logins,
			count(case when t1.cdate3 = t2.cdate then t1.userid end) logins3,
			count(case when t1.cdate5 = t2.cdate then t1.userid end) logins5,
			count(case when t1.cdate7 = t2.cdate then t1.userid end) logins7
		from (
			select userid, toYYYYMMDD(dateAdd(DAY, -1, login_time)) cdate,
			  toYYYYMMDD(dateAdd(DAY, -3, login_time)) cdate3,
			  toYYYYMMDD(dateAdd(DAY, -5, login_time)) cdate5,
			  toYYYYMMDD(dateAdd(DAY, -7, login_time)) cdate7
			from game.col_log_login final
			where login_time between ? AND ?
			group by userid, cdate, cdate3, cdate5, cdate7
		) t1 join (
			select toYYYYMMDD(toDateTime(s1.ctime/1000)) cdate, s1.userid userid
			from game.col_user s0 final
			join game.col_activity_volatility_subsidy s1 final on s1.userid = s0.userid
			where first = 1 and s1.ctime between ? and ? %s
			group by cdate, s1.userid
		) t2 on t1.userid = t2.userid
		group by t2.cdate
	`, where_u_sql), scbt_lcl_args...)
	if err != nil {
		return
	}
	now := time.Now().In(Location())
	for _, data := range scbt_lcl_stats {
		cdate := fmt.Sprint(data["cdate"].(uint32))
		sdate := cdate[0:4] + "-" + cdate[4:6] + "-" + cdate[6:8]
		logins := utils.ToInt64(data["logins"])
		logins3 := utils.ToInt64(data["logins3"])
		logins5 := utils.ToInt64(data["logins5"])
		logins7 := utils.ToInt64(data["logins7"])

		ctime, e := time.ParseInLocation(utils.FORMAT_DATE, sdate, Location())
		if e != nil {
			err = e
			return
		}

		if stat, ok := dateStats[sdate]; ok {
			stat.FirstSubsidyLoginDay1 = logins
			stat.FirstSubsidyLoginDay3 = logins3
			stat.FirstSubsidyLoginDay5 = logins5
			stat.FirstSubsidyLoginDay7 = logins7
			if ctime.AddDate(0, 0, 1).Before(now) {
				stat.FirstSubsidyRetentionDay1 = fmt.Sprintf("%.2f%%", ComputeFloat(logins, stat.FirstSubsidyUsers)*100)
			}
			if ctime.AddDate(0, 0, 3).Before(now) {
				stat.FirstSubsidyRetentionDay3 = fmt.Sprintf("%.2f%%", ComputeFloat(logins3, stat.FirstSubsidyUsers)*100)
			}
			if ctime.AddDate(0, 0, 5).Before(now) {
				stat.FirstSubsidyRetentionDay5 = fmt.Sprintf("%.2f%%", ComputeFloat(logins5, stat.FirstSubsidyUsers)*100)
			}
			if ctime.AddDate(0, 0, 7).Before(now) {
				stat.FirstSubsidyRetentionDay7 = fmt.Sprintf("%.2f%%", ComputeFloat(logins7, stat.FirstSubsidyUsers)*100)
			}
		}
	}

	// 首充用户次日留存率
	var sc_lcl_args = []any{begin.AddDate(0, 0, 1), end.AddDate(0, 0, 7), begin, end}
	sc_lcl_args = append(sc_lcl_args, where_u_args...)
	sc_lcl_args = append(sc_lcl_args, begin, end)
	var sc_lcl_datas []map[string]any
	err = ck.Select(&sc_lcl_datas, fmt.Sprintf(`
		select t2.first_pay_date cdate, 
			count(case when t1.cdate = t2.first_pay_date then t1.userid end) logins,
			count(case when t1.cdate3 = t2.first_pay_date then t1.userid end) logins3,
			count(case when t1.cdate5 = t2.first_pay_date then t1.userid end) logins5,
			count(case when t1.cdate7 = t2.first_pay_date then t1.userid end) logins7
		from (
			select userid,
				toYYYYMMDD(dateAdd(DAY, -1, login_time)) cdate, 
			    toYYYYMMDD(dateAdd(DAY, -3, login_time)) cdate3,
			    toYYYYMMDD(dateAdd(DAY, -5, login_time)) cdate5,
			    toYYYYMMDD(dateAdd(DAY, -7, login_time)) cdate7
			from game.col_log_login final
			where login_time between ? AND ?
			group by userid, cdate, cdate3, cdate5, cdate7
		) t1 join (
			select first_pay_date, userid
			from (
				select userid, min(ctime) first_pay_time, toYYYYMMDD(min(ctime)) first_pay_date
				from game.col_trade_record s1 final where userid in (
					select userid
					from game.col_user s0 final
					join game.col_trade_record s1 final on s1.userid = s0.userid
					where ctime between ? AND ? and order_status = 4 %s
				) and order_status = 4
				group by userid
				having first_pay_time between ? AND ?
			) t0
			group by first_pay_date, userid
		) t2 on t1.userid = t2.userid
		group by t2.first_pay_date
	`, where_u_sql), sc_lcl_args...)
	if err != nil {
		return
	}
	for _, data := range sc_lcl_datas {
		cdate := fmt.Sprint(data["cdate"].(uint32))
		sdate := cdate[0:4] + "-" + cdate[4:6] + "-" + cdate[6:8]
		logins := utils.ToInt64(data["logins"])
		logins3 := utils.ToInt64(data["logins3"])
		logins5 := utils.ToInt64(data["logins5"])
		logins7 := utils.ToInt64(data["logins7"])

		ctime, e := time.ParseInLocation(utils.FORMAT_DATE, sdate, Location())
		if e != nil {
			err = e
			return
		}

		if stat, ok := dateStats[sdate]; ok {
			stat.FirstPayLoginDay1 = logins
			stat.FirstPayLoginDay3 = logins3
			stat.FirstPayLoginDay5 = logins5
			stat.FirstPayLoginDay7 = logins7
			if ctime.AddDate(0, 0, 1).Before(now) {
				stat.FirstPayRetentionDay1 = fmt.Sprintf("%.2f%%", ComputeFloat(logins, stat.FirstPayUsers)*100)
			}
			if ctime.AddDate(0, 0, 3).Before(now) {
				stat.FirstPayRetentionDay3 = fmt.Sprintf("%.2f%%", ComputeFloat(logins3, stat.FirstPayUsers)*100)
			}
			if ctime.AddDate(0, 0, 5).Before(now) {
				stat.FirstPayRetentionDay5 = fmt.Sprintf("%.2f%%", ComputeFloat(logins5, stat.FirstPayUsers)*100)
			}
			if ctime.AddDate(0, 0, 7).Before(now) {
				stat.FirstPayRetentionDay7 = fmt.Sprintf("%.2f%%", ComputeFloat(logins7, stat.FirstPayUsers)*100)
			}
		}
	}

	// 首次补贴用户7日人均付费总额
	var scbt_pay_args = []any{begin, end.AddDate(0, 0, 7), begin.UnixMilli(), end.UnixMilli()}
	scbt_pay_args = append(scbt_pay_args, where_u_args...)
	var scbt_pay_datas []map[string]any
	err = ck.Select(&scbt_pay_datas, fmt.Sprintf(`
		select t2.cdate cdate, SUM(t1.amounts) amount_sum, count(distinct t1.userid) users
		from (
			select toYYYYMMDD(ctime) cdate, userid, SUM(amount) amounts
			from game.col_trade_record final
			where order_status = 4 and ctime between ? and ?
			group by cdate, userid 
		) t1 join (
			select s1.userid userid, toYYYYMMDD(toDateTime(s1.ctime/1000)) cdate, toYYYYMMDD(dateAdd(DAY, 6, toDateTime(s1.ctime/1000))) edate
			from game.col_user s0 final
			join game.col_activity_volatility_subsidy s1 final on s1.userid = s0.userid
			where first = 1 and s1.ctime between ? and ? %s
			group by s1.userid, cdate, edate
		) t2 on t1.userid = t2.userid and t1.cdate between t2.cdate and t2.edate
		group by t2.cdate
		SETTINGS allow_experimental_join_condition = 1
	`, where_u_sql), scbt_pay_args...)
	if err != nil {
		return
	}
	for _, data := range scbt_pay_datas {
		cdate := fmt.Sprint(data["cdate"].(uint32))
		sdate := cdate[0:4] + "-" + cdate[4:6] + "-" + cdate[6:8]
		amount_sum := utils.ToInt64(data["amount_sum"])
		users := utils.ToInt64(data["users"])

		if stat, ok := dateStats[sdate]; ok {
			stat.PayAmount1 = amount_sum
			stat.PayUsers1 = users
			// stat.FirstSubsidyPayAmountDay7 = fmt.Sprintf("%.2f, %.2f", Chip2Float(amount_sum), Chip2Float(ComputeFloat(amount_sum, users)))
			stat.FirstSubsidyPayAmountDay7 = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(amount_sum, users)))
		}
	}

	// 首次补贴用户7日人均付费总额
	var sc_pay7_args = []any{begin, end.AddDate(0, 0, 7), begin, end}
	sc_pay7_args = append(sc_pay7_args, where_u_args...)
	sc_pay7_args = append(sc_pay7_args, begin, end)
	var sc_pay7_datas []map[string]any
	err = ck.Select(&sc_pay7_datas, fmt.Sprintf(`
		select t2.first_pay_date cdate, SUM(t1.amounts) amount_sum, count(distinct t1.userid) users
		from (
			select toYYYYMMDD(ctime) cdate, userid, SUM(amount) amounts
			from game.col_trade_record final
			where order_status = 4 and ctime between ? and ?
			group by cdate, userid
		) t1 join (
			select first_pay_date, edate, userid
			from (
				select userid, min(ctime) first_pay_time, 
					toYYYYMMDD(first_pay_time) first_pay_date, toYYYYMMDD(dateAdd(DAY, 6, first_pay_time)) edate
				from game.col_trade_record s1 final where userid in (
					select userid
					from game.col_user s0 final
					join game.col_trade_record s1 final on s1.userid = s0.userid
					where ctime between ? AND ? and order_status = 4 %s
				) and order_status = 4
				group by userid
				having first_pay_time between ? AND ?
			) t0
			group by first_pay_date, edate, userid
		) t2 on t1.userid = t2.userid and t1.cdate between t2.first_pay_date and t2.edate
		group by t2.first_pay_date
		SETTINGS allow_experimental_join_condition = 1
	`, where_u_sql), sc_pay7_args...)
	if err != nil {
		return
	}
	for _, data := range sc_pay7_datas {
		cdate := fmt.Sprint(data["cdate"].(uint32))
		sdate := cdate[0:4] + "-" + cdate[4:6] + "-" + cdate[6:8]
		amount_sum := utils.ToInt64(data["amount_sum"])
		users := utils.ToInt64(data["users"])

		if stat, ok := dateStats[sdate]; ok {
			stat.PayAmount2 = amount_sum
			stat.PayUsers2 = users
			// stat.FirstPayAmountDay7Avg = fmt.Sprintf("%.2f, %.2f", Chip2Float(amount_sum), Chip2Float(ComputeFloat(amount_sum, users)))
			stat.FirstPayAmountDay7Avg = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(amount_sum, users)))
		}
	}

	// 水池数据
	m := bson.M{"ptype": bson.M{"$in": []int{1, 2, 3}}}
	if registArea > 0 {
		m["regist_area"] = registArea - 1
	}
	if len(stats) > 0 {
		var dates []string
		for _, stat := range stats {
			dates = append(dates, stat.SDate)
		}
		m["date"] = bson.M{"$in": dates}
		var datas []entity.VolatilitySubsidyPool
		err = VolatilitySubsidyPools.Find(m).All(&datas)
		if err != nil {
			return
		}
		for _, data := range datas {
			if stat, ok := dateStats[data.Date]; ok {
				stat.SubsidyPools[data.Ptype] += data.Amount
				// stat.SubsidyPools[data.RegistArea] = max(data.Amount, stat.SubsidyPools[data.RegistArea])
			}
		}
	}
	for _, stat := range stats {
		subsidyPool := max(stat.SubsidyPools[1], stat.SubsidyPools[2], stat.SubsidyPools[3])
		stat.SubsidyPool0 = subsidyPool
		stat.SubsidyPool = fmt.Sprintf("%.2f", Chip2Float(stat.SubsidyPool0))
		stat.SubsidyConsumRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.SubsidyAmount0, stat.SubsidyPool0)*100)
	}

	// 汇总
	summary := &entity.VolatilitySubsidyDate{
		SDate:        fmt.Sprintf("%s~%s汇总", begin.Format(utils.FORMAT_DATE), end.Format(utils.FORMAT_DATE)),
		SubsidyPools: make(map[int32]int64),
	}
	stats = append([]*entity.VolatilitySubsidyDate{summary}, stats...)

	var topSubsidy3 = []int64{0, 0, 0}
	var subsidyTop3 = []string{"", "", ""}
	for _, stat := range stats {
		summary.LoginedUsers += stat.LoginedUsers
		summary.FirstPayUsers += stat.FirstPayUsers
		summary.SubsidyUsers += stat.SubsidyUsers
		summary.FirstSubsidyUsers += stat.FirstSubsidyUsers
		summary.SubsidyTimes += stat.SubsidyTimes
		summary.FirstPaySubsidyUsers += stat.FirstPaySubsidyUsers

		summary.SubsidyPool0 += stat.SubsidyPool0
		summary.SubsidyAmount0 += stat.SubsidyAmount0
		summary.FirstSubsidyLoginDay1 += stat.FirstSubsidyLoginDay1
		summary.FirstSubsidyLoginDay3 += stat.FirstSubsidyLoginDay3
		summary.FirstSubsidyLoginDay5 += stat.FirstSubsidyLoginDay5
		summary.FirstSubsidyLoginDay7 += stat.FirstSubsidyLoginDay7
		summary.FirstPayLoginDay1 += stat.FirstPayLoginDay1
		summary.FirstPayLoginDay3 += stat.FirstPayLoginDay3
		summary.FirstPayLoginDay5 += stat.FirstPayLoginDay5
		summary.FirstPayLoginDay7 += stat.FirstPayLoginDay7
		summary.PayAmount1 += stat.PayAmount1
		summary.PayUsers1 += stat.PayUsers1
		summary.PayAmount2 += stat.PayAmount2
		summary.PayUsers2 += stat.PayUsers2

		for i, subsidy := range stat.TopSubsidy3 {
			overhit := -1
			for j, topSubsidy := range topSubsidy3 {
				if subsidy > topSubsidy {
					overhit = j
					break
				}
			}
			if overhit >= 0 {
				topSubsidy3[overhit] = subsidy
				subsidyTop3[overhit] = stat.SubsidyTop3[i]
			}
		}
	}
	summary.TopSubsidy3 = topSubsidy3
	summary.SubsidyTop3 = subsidyTop3

	summary.FirstSubsidyRate = fmt.Sprintf("%.2f%%", ComputeFloat(summary.FirstSubsidyUsers, summary.SubsidyUsers)*100)
	summary.SubsidyAvg = fmt.Sprintf("%.2f", ComputeFloat(summary.SubsidyTimes, summary.SubsidyUsers))

	summary.FirstPaySubsidyRate = fmt.Sprintf("%.2f%%", ComputeFloat(summary.FirstPaySubsidyUsers, summary.FirstPayUsers)*100)
	summary.FirstPaySubsidyUsersRate = fmt.Sprintf("%.2f%%", ComputeFloat(summary.FirstPaySubsidyUsers, summary.SubsidyUsers)*100)
	summary.SubsidyAmount = fmt.Sprintf("%.2f", Chip2Float(summary.SubsidyAmount0))
	summary.SubsidyAmountAvg = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(summary.SubsidyAmount0, summary.SubsidyUsers)))

	summary.FirstSubsidyRetentionDay1 = fmt.Sprintf("%.2f%%", ComputeFloat(summary.FirstSubsidyLoginDay1, summary.FirstSubsidyUsers)*100)
	summary.FirstSubsidyRetentionDay3 = fmt.Sprintf("%.2f%%", ComputeFloat(summary.FirstSubsidyLoginDay3, summary.FirstSubsidyUsers)*100)
	summary.FirstSubsidyRetentionDay5 = fmt.Sprintf("%.2f%%", ComputeFloat(summary.FirstSubsidyLoginDay5, summary.FirstSubsidyUsers)*100)
	summary.FirstSubsidyRetentionDay7 = fmt.Sprintf("%.2f%%", ComputeFloat(summary.FirstSubsidyLoginDay7, summary.FirstSubsidyUsers)*100)
	summary.FirstPayRetentionDay1 = fmt.Sprintf("%.2f%%", ComputeFloat(summary.FirstPayLoginDay1, summary.FirstPayUsers)*100)
	summary.FirstPayRetentionDay3 = fmt.Sprintf("%.2f%%", ComputeFloat(summary.FirstPayLoginDay3, summary.FirstPayUsers)*100)
	summary.FirstPayRetentionDay5 = fmt.Sprintf("%.2f%%", ComputeFloat(summary.FirstPayLoginDay5, summary.FirstPayUsers)*100)
	summary.FirstPayRetentionDay7 = fmt.Sprintf("%.2f%%", ComputeFloat(summary.FirstPayLoginDay7, summary.FirstPayUsers)*100)
	summary.FirstSubsidyPayAmountDay7 = fmt.Sprintf(" %.2f", Chip2Float(ComputeFloat(summary.PayAmount1, summary.PayUsers1)))
	summary.FirstPayAmountDay7Avg = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(summary.PayAmount2, summary.PayUsers2)))

	summary.SubsidyPool = fmt.Sprintf("%.2f", Chip2Float(summary.SubsidyPool0))
	summary.SubsidyConsumRate = fmt.Sprintf("%.2f%%", ComputeFloat(summary.SubsidyAmount0, summary.SubsidyPool0)*100)
	return
}

// GetActivityVolatilitySubsidyUsers 波动返水明细
// registArea: 1.A 2.B 3.C
func (s *activityService) GetActivityVolatilitySubsidyUsers(page, pageSize int, sortBy, asc string, begin, end time.Time, registArea int, packageIds []string) (total int64, stats []*entity.VolatilitySubsidyUser, err error) {
	where_u_sql := ""
	var where_u_args []any
	if registArea > 0 {
		where_u_sql += " and s0.regist_area = ?"
		where_u_args = append(where_u_args, registArea-1)
	}
	if len(packageIds) > 0 {
		where_u_sql += " and s0.ad__bundle_id IN ?"
		where_u_args = append(where_u_args, packageIds)
	}

	var sql_total_args = []any{begin.UnixMilli(), end.UnixMilli()}
	sql_total_args = append(sql_total_args, where_u_args...)
	sql_total := `
		select count(*) total
		from game.col_user s0 final
		join (
			select userid, MIN(ctime) first_subsidy_time, SUM(subsidy) subsidys, count(*) subsidy_times, MAX(ctime) last_subsidy_ctime
			from game.col_activity_volatility_subsidy final
			group by userid
			having first_subsidy_time between ? and ?
		) s1 on s0.userid = s1.userid
		where 1 = 1 %s
	`
	err = ck.Select(&total, fmt.Sprintf(sql_total, where_u_sql), sql_total_args...)
	if err != nil {
		return
	}
	if total == 0 {
		return
	}

	orderBy := "s1.first_subsidy_time desc"
	orderField := ""
	switch sortBy {
	case "FirstPayAmount":
		orderField = "s0.first_charge_val"
	case "Pays":
		orderField = "s0.money"
	case "PaySubWithdraws":
		orderField = "profit"
	case "Subsidys":
		orderField = "s1.subsidys"
	}
	if orderField != "" {
		sort := "DESC"
		if asc == "1" {
			sort = "ASC"
		}
		orderBy = fmt.Sprintf("%s %s", orderField, sort)
	}

	offset, limit := PageCalc(page, pageSize)
	var sql_args = append(sql_total_args, offset, limit)
	sql := `
		select s1.userid userid, s1.first_subsidy_time, s1.last_subsidy_ctime, s1.subsidys, s1.subsidy_times,
			s0.regist_area, s0.ctime, s0.login_time, s0.vip_lv, s0.state, s0.money, s0.cash_out, (s0.money - s0.cash_out) profit, s0.first_charge_time, s0.first_charge_val
		from game.col_user s0 final
		join (
			select userid, MIN(ctime) first_subsidy_time, SUM(subsidy) subsidys, count(*) subsidy_times, MAX(ctime) last_subsidy_ctime
			from game.col_activity_volatility_subsidy final
			group by userid
			having first_subsidy_time between ? and ?
		) s1 on s0.userid = s1.userid
		where 1 = 1 %s
		order by %s
		limit ?, ?
	`
	var datas []map[string]any
	err = ck.Select(&datas, fmt.Sprintf(sql, where_u_sql, orderBy), sql_args...)
	if err != nil {
		return
	}

	now := time.Now().In(location)
	for _, data := range datas {
		userid := data["userid"].(string)
		first_subsidy_time := utils.ToInt64(data["first_subsidy_time"])
		subsidys := utils.ToInt64(data["subsidys"])
		subsidy_times := utils.ToInt64(data["subsidy_times"])
		// regist_area := utils.ToInt64(data["regist_area"])
		ctime := data["ctime"].(time.Time).In(location)
		vip_lv := utils.ToInt64(data["vip_lv"])
		state := utils.ToInt64(data["state"])
		money := utils.ToInt64(data["money"])
		cash_out := utils.ToInt64(data["cash_out"])
		profit := utils.ToInt64(data["profit"])
		first_charge_val := utils.ToInt64(data["first_charge_val"])
		first_charge_time := utils.ToInt64(data["first_charge_time"])

		user := &entity.VolatilitySubsidyUser{
			Userid:               userid,
			VipLv:                fmt.Sprint(vip_lv),
			RegistDate:           ctime.Format(utils.FORMAT_DATE),
			FirstPayTime:         utils.CaseElse(first_charge_time == 0, "-", time.UnixMilli(first_charge_time).In(location).Format(utils.FORMAT)),
			FirstSubsidyTime:     time.UnixMilli(first_subsidy_time).In(location).Format(utils.FORMAT),
			FirstPay2SubsidyDays: fmt.Sprint(int32((first_subsidy_time - first_charge_time) / 1000 / 60 / 60 / 24)),
			FirstPayAmount:       utils.CaseElse(first_charge_val == 0, "-", fmt.Sprintf("%.2f", Chip2Float(first_charge_val))),
			Pays:                 fmt.Sprintf("%.2f", Chip2Float(money)),
			Withdraws:            fmt.Sprintf("%.2f", Chip2Float(cash_out)),
			PaySubWithdraws:      fmt.Sprintf("%.2f", Chip2Float(profit)),
			Subsidys:             fmt.Sprintf("%.2f", Chip2Float(subsidys)),
			SubsidyTimes:         subsidy_times,
		}

		if login_time, ok := data["login_time"].(time.Time); ok {
			login_time = login_time.In(location)
			liveDays := int64(math.Ceil(login_time.Sub(ctime).Hours() / 24))
			loseDays := int64(math.Floor(now.Sub(login_time).Hours() / 24))
			user.LiveDays = fmt.Sprint(liveDays)
			user.LoseDays = fmt.Sprint(loseDays)
		}

		var chargeType string
		{
			if money == 0 {
				if state == 1 {
					chargeType = "新手"
				} else {
					chargeType = "平民"
				}
			} else if money >= 1 && money <= 99999 {
				chargeType = "普充"
			} else if money >= 100000 && money <= 499999 {
				chargeType = "小R"
			} else if money >= 500000 && money <= 999999 {
				chargeType = "中R"
			} else if money >= 1000000 && money <= 19999999 {
				chargeType = "大R"
			} else if money >= 20000000 {
				chargeType = "超大R"
			}
		}
		user.ChargeType = chargeType

		stats = append(stats, user)
	}

	// 汇总
	mark := "-"
	summary := &entity.VolatilitySubsidyUser{
		Userid:               mark,
		VipLv:                mark,
		ChargeType:           mark,
		RegistDate:           mark,
		LiveDays:             mark,
		LoseDays:             mark,
		FirstPayTime:         mark,
		FirstSubsidyTime:     mark,
		FirstPay2SubsidyDays: mark,
		FirstPayAmount:       mark,
	}
	stats = append([]*entity.VolatilitySubsidyUser{summary}, stats...)

	var summary_data = make(map[string]any)
	sql_summary := `
		select SUM(s1.subsidys) subsidy_sum, SUM(s1.subsidy_times) subsidy_times_sum, 
			SUM(s0.money) money_sum, SUM(s0.cash_out) cash_out_sum
		from game.col_user s0 final
		join (
			select userid, MIN(ctime) first_subsidy_time, SUM(subsidy) subsidys, count(*) subsidy_times, MAX(ctime) last_subsidy_ctime
			from game.col_activity_volatility_subsidy final
			group by userid
			having first_subsidy_time between ? and ?
		) s1 on s0.userid = s1.userid
		where 1 = 1 %s
	`
	err = ck.Select(&summary_data, fmt.Sprintf(sql_summary, where_u_sql), sql_total_args...)
	if err != nil {
		return
	}
	subsidy_sum := utils.ToInt64(summary_data["subsidy_sum"])
	subsidy_times_sum := utils.ToInt64(summary_data["subsidy_times_sum"])
	money_sum := utils.ToInt64(summary_data["money_sum"])
	cash_out_sum := utils.ToInt64(summary_data["cash_out_sum"])

	summary.Pays = fmt.Sprintf("%.2f", Chip2Float(money_sum))
	summary.Withdraws = fmt.Sprintf("%.2f", Chip2Float(cash_out_sum))
	summary.PaySubWithdraws = fmt.Sprintf("%.2f", Chip2Float(money_sum-cash_out_sum))
	summary.Subsidys = fmt.Sprintf("%.2f", Chip2Float(subsidy_sum))
	summary.SubsidyTimes = subsidy_times_sum
	return
}

// 代理后台-当日产出/实付总成本
func (s *activityService) shareAgentCost(stime, etime time.Time, turn bool) (dateStats map[uint32]*entity.ShareAgentCost, err error) {
	dateStats = make(map[uint32]*entity.ShareAgentCost)

	var getDateStat = func(date uint32) *entity.ShareAgentCost {
		stat, ok := dateStats[date]
		if !ok {
			stat = &entity.ShareAgentCost{
				Date: date,
			}
			dateStats[date] = stat
		}
		return stat
	}

	// 当日产出总成本/当日实付总成本
	// 代理发放
	var share_income_datas []map[string]any
	err = ck.Select(&share_income_datas, `
		select toYYYYMMDD(toDateTime(ctime/1000)) cdate, (SUM(amount) / 100) amounts,
			(SUM(case when itype = 4 then amount else 0 end) / 100) child_amounts
		from game.col_activity_share_income_record final
		where ctime between ? and ? and itype in (1,2,3,4)
		group by cdate
		order by cdate desc
	`, stime.UnixMilli(), etime.UnixMilli())
	if err != nil {
		return
	}
	for _, data := range share_income_datas {
		cdate := data["cdate"].(uint32)
		amounts := utils.ToInt64(data["amounts"])
		child_amounts := utils.ToInt64(data["child_amounts"])
		stat := getDateStat(cdate)
		stat.DayProductCost += amounts
		stat.DayProductCostHeadChild += child_amounts
	}

	// 代理领取
	var share_tack_datas []map[string]any
	err = ck.Select(&share_tack_datas, `
		select toYYYYMMDD(toDateTime(ctime/1000)) cdate, SUM(amount) amounts
		from game.col_activity_share_income_record_tack_log final
		where ctime between ? and ? and itype in (1,2,3)
		group by cdate
		order by cdate desc
	`, stime.UnixMilli(), etime.UnixMilli())
	if err != nil {
		return
	}
	for _, data := range share_tack_datas {
		cdate := data["cdate"].(uint32)
		amounts := utils.ToInt64(data["amounts"])
		stat := getDateStat(cdate)
		stat.DayPayOutCost += amounts
		stat.DayPayOutCost += stat.DayProductCostHeadChild
	}

	// 转盘领取/发放
	if turn {
		var turn_datas []map[string]any
		err = ck.Select(&turn_datas, `
			select toYYYYMMDD(toDateTime(ctime)) cdate, sum(prize) prizes,
				sum(case when state = 1 then prize else 0 end) tack_prizes
			from game.col_activity_turn_prize_log final
			where ctime between ? and ?
			group by cdate
			order by cdate desc
		`, stime.Unix(), etime.Unix())
		if err != nil {
			return
		}
		for _, data := range turn_datas {
			cdate := data["cdate"].(uint32)
			prizes := utils.ToInt64(data["prizes"])
			tack_prizes := utils.ToInt64(data["tack_prizes"])
			stat := getDateStat(cdate)
			stat.DayProductCost += prizes
			stat.DayPayOutCost += tack_prizes
		}
	}
	return
}

// 代理后台-裂变体系日报
func (s *activityService) GetShareAgentDates(stime, etime time.Time) (stats []*entity.ShareAgentDate, err error) {
	dateStats := make(map[uint32]*entity.ShareAgentDate)
	for begin := stime; begin.Before(etime); begin = begin.AddDate(0, 0, 1) {
		day_time := uint32(begin.Year()*10000 + int(begin.Month())*100 + begin.Day())
		sday := fmt.Sprint(day_time)
		sdate := sday[0:4] + "-" + sday[4:6] + "-" + sday[6:8]
		stat := &entity.ShareAgentDate{SDate: sdate}
		stats = append(stats, stat)
		dateStats[day_time] = stat
	}
	utils.SliceReverse(stats)

	// 产出/实付成本
	costMap, err := s.shareAgentCost(stime, etime, true)
	if err != nil {
		return
	}
	for date, stat := range dateStats {
		if cost, ok := costMap[date]; ok {
			stat.DayProductCost0 = cost.DayProductCost
			stat.DayPayOutCost0 = cost.DayPayOutCost
		}
	}

	var pay_datas []map[string]any
	err = ck.Select(&pay_datas, `
		select cdate, SUM(pays) pays, SUM(withdraws) withdraws from (
			select toYYYYMMDD(ctime) cdate, SUM(amount) pays, 0 withdraws
			from game.col_trade_record final 
			where ctime between ? and ? and order_status = 4
			group by cdate
			
			union all
			
			select toYYYYMMDD(ctime) cdate, 0 pays, SUM(amount) withdraws
			from game.col_withdraw_record final 
			where ctime between ? and ? and order_status = 2
			group by cdate
		) t0
		group by cdate
	`, stime, etime, stime, etime)
	if err != nil {
		return
	}
	for _, data := range pay_datas {
		cdate := data["cdate"].(uint32)
		pays := utils.ToInt64(data["pays"])
		withdraws := utils.ToInt64(data["withdraws"])
		if stat, ok := dateStats[cdate]; ok {
			stat.DayROI = fmt.Sprintf("%.2f", ComputeFloat(pays-withdraws, stat.DayPayOutCost0))
		}
	}

	var share_login_datas []map[string]any
	err = ck.Select(&share_login_datas, `
		select s1.login_date, count(distinct s1.userid) login_users,
			count(distinct (case when s1.ad__bundle_id = 'net.playtouch.classictictactoe' then s1.userid else null end)) apk_users,
			count(distinct (case when s1.cdate = s1.login_date then s1.userid else null end)) new_login_users,
			(login_users - new_login_users) old_login_users,
			count(distinct (case when payed = 1 then s1.userid else null end)) login_pay_users
		from game.col_user s0 final
		join (
			select toYYYYMMDD(s1.login_time) login_date, s1.userid userid, s0.ad__bundle_id, toYYYYMMDD(s0.ctime) cdate, 
				(case when s0.money > 0 then 1 else 0 end) payed
			from game.col_user s0 final
			join game.col_log_login s1 final on s1.userid = s0.userid
			where s1.login_time between ? and ?
			group by login_date, s1.userid, s0.ad__bundle_id, cdate, payed
		) s1 on (s1.userid = s0.userid and s0.share_superior != '') or s1.userid = s0.share_superior
		WHERE s0.share_superior != ''
		group by s1.login_date
	`, stime, etime)
	if err != nil {
		return
	}
	for _, data := range share_login_datas {
		login_date := data["login_date"].(uint32)
		login_users := utils.ToInt64(data["login_users"])
		apk_users := utils.ToInt64(data["apk_users"])
		new_login_users := utils.ToInt64(data["new_login_users"])
		old_login_users := utils.ToInt64(data["old_login_users"])
		login_pay_users := utils.ToInt64(data["login_pay_users"])
		if stat, ok := dateStats[login_date]; ok {
			stat.ShareLoginUsers = login_users
			stat.ShareApkLoginUsers = apk_users
			stat.ShareApkLoginRate = fmt.Sprintf("%.2f%%", ComputeFloat(apk_users, login_users)*100)
			stat.NewShareLoginUsers = new_login_users
			stat.OldShareLoginUsers = old_login_users
			stat.ShareLoginPayUsers = login_pay_users
		}
	}

	// 新增代理头目
	var new_inviter_datas []map[string]any
	err = ck.Select(&new_inviter_datas, `
		select first_be_super_date, count(*) first_be_supers from (
			select share_superior, min(ctime) first_be_super_time, toYYYYMMDD(first_be_super_time) first_be_super_date
			from game.col_user s0 final
			where share_superior in (
				select distinct share_superior
				from game.col_user s0 final
				where s0.ctime between ? and ? and share_superior != ''
			)
			group by share_superior
			having first_be_super_time between ? and ? 
		) t0
		group by first_be_super_date
	`, stime, etime, stime, etime)
	if err != nil {
		return
	}
	for _, data := range new_inviter_datas {
		first_be_super_date := data["first_be_super_date"].(uint32)
		first_be_supers := utils.ToInt64(data["first_be_supers"])
		if stat, ok := dateStats[first_be_super_date]; ok {
			stat.NewInviterUsers = first_be_supers
		}
	}

	// 体系充值数据
	var agent_pay_datas []map[string]any
	err = ck.Select(&agent_pay_datas, `
		select s1.sdate, count(*) pay_users, SUM(s1.pays) pay_amounts,
			SUM(case when s1.sdate = toYYYYMMDD(s0.ctime) then 1 else 0 end) new_pay_users,
			(pay_users - new_pay_users) old_pay_users,
			SUM(case when s1.sdate = toYYYYMMDD(s0.ctime) then s1.pays else 0 end) new_pay_amounts,
			(pay_amounts - new_pay_amounts) old_pay_amounts
		from game.col_user s0 final
		join (
			select distinct s1.sdate, s1.userid userid, s1.pays
			from game.col_user s0 final
			join (
				select toYYYYMMDD(ctime) sdate, userid, SUM(amount) pays
				from game.col_trade_record final 
				where ctime between ? and ? and order_status = 4
				group by sdate, userid
			) s1 on (s1.userid = s0.userid and s0.share_superior != '') or s1.userid = s0.share_superior
		) s1 on s1.userid = s0.userid
		group by s1.sdate
	`, stime, etime)
	if err != nil {
		return
	}
	for _, data := range agent_pay_datas {
		sdate := data["sdate"].(uint32)
		pay_users := utils.ToInt64(data["pay_users"])
		pay_amounts := utils.ToInt64(data["pay_amounts"])
		new_pay_users := utils.ToInt64(data["new_pay_users"])
		old_pay_users := utils.ToInt64(data["old_pay_users"])
		new_pay_amounts := utils.ToInt64(data["new_pay_amounts"])
		old_pay_amounts := utils.ToInt64(data["old_pay_amounts"])
		if stat, ok := dateStats[sdate]; ok {
			stat.SharePayAmounts0 = pay_amounts
			stat.SharePayUsers = pay_users

			stat.NewSharePayAmounts0 = new_pay_amounts
			stat.NewSharePayUsers = new_pay_users
			stat.OldSharePayAmounts0 = old_pay_amounts
			stat.OldSharePayUsers = old_pay_users
		}
	}
	// 体系提现
	var agent_withdraw_datas []map[string]any
	err = ck.Select(&agent_withdraw_datas, `
		select s1.sdate, count(*) withdraw_users, SUM(s1.withdraws) withdraw_amounts,
			SUM(case when s1.sdate = toYYYYMMDD(s0.ctime) then 1 else 0 end) new_withdraw_users,
			(withdraw_users - new_withdraw_users) old_withdraw_users,
			SUM(case when s1.sdate = toYYYYMMDD(s0.ctime) then s1.withdraws else 0 end) new_withdraw_amounts,
			(withdraw_amounts - new_withdraw_amounts) old_withdraw_amounts
		from game.col_user s0 final
		join (
			select distinct s1.sdate, s1.userid userid, s1.withdraws
			from game.col_user s0 final
			join (
				select toYYYYMMDD(ctime) sdate, userid, SUM(amount) withdraws
				from game.col_withdraw_record final
				where ctime between ? and ? and order_status = 2
				group by sdate, userid
			) s1 on (s1.userid = s0.userid and s0.share_superior != '') or s1.userid = s0.share_superior
		) s1 on s1.userid = s0.userid
		group by s1.sdate
	`, stime, etime)
	if err != nil {
		return
	}
	for _, data := range agent_withdraw_datas {
		sdate := data["sdate"].(uint32)
		withdraw_users := utils.ToInt64(data["withdraw_users"])
		withdraw_amounts := utils.ToInt64(data["withdraw_amounts"])
		new_withdraw_users := utils.ToInt64(data["new_withdraw_users"])
		old_withdraw_users := utils.ToInt64(data["old_withdraw_users"])
		new_withdraw_amounts := utils.ToInt64(data["new_withdraw_amounts"])
		old_withdraw_amounts := utils.ToInt64(data["old_withdraw_amounts"])
		if stat, ok := dateStats[sdate]; ok {
			stat.ShareWithdrawAmounts0 = withdraw_amounts
			stat.ShareWithdrawUsers = withdraw_users

			stat.NewShareWithdrawAmounts0 = new_withdraw_amounts
			stat.NewShareWithdrawUsers = new_withdraw_users
			stat.OldShareWithdrawAmounts0 = old_withdraw_amounts
			stat.OldShareWithdrawUsers = old_withdraw_users
		}
	}

	// 体系首充
	var agent_first_pay_datas []map[string]any
	err = ck.Select(&agent_first_pay_datas, `
		select s1.sdate, count(*) first_pay_users, SUM(s1.pays) first_pay_amounts
		from (
			select distinct s1.sdate, s1.userid userid, s1.pays
			from game.col_user s0 final
			join (
				select toYYYYMMDD(ctime) sdate, userid, SUM(amount) pays
				from game.col_trade_record final
				where ctime between ? and ? and order_status = 4 and first_pay = 1
				group by sdate, userid
			) s1 on (s1.userid = s0.userid and s0.share_superior != '') or s1.userid = s0.share_superior
			order by s1.sdate desc
		) s1
		group by s1.sdate
	`, stime, etime)
	if err != nil {
		return
	}
	for _, data := range agent_first_pay_datas {
		sdate := data["sdate"].(uint32)
		first_pay_users := utils.ToInt64(data["first_pay_users"])
		first_pay_amounts := utils.ToInt64(data["first_pay_amounts"])

		if stat, ok := dateStats[sdate]; ok {
			stat.ShareFirstPayUsers = first_pay_users
			stat.ShareFirstPayAmounts0 = first_pay_amounts
		}
	}

	// 体系首充提现
	var agent_first_pay_withdraw_datas []map[string]any
	err = ck.Select(&agent_first_pay_withdraw_datas, `
		select sdate, count(*) withdraw_users, sum(withdraws) withdraws from (
			select toYYYYMMDD(ctime) sdate, userid, SUM(amount) withdraws
			from game.col_withdraw_record final
			where ctime between ? and ? and order_status = 2
			group by sdate, userid
			having (sdate, userid) in (
				select toYYYYMMDD(ctime) sdate, userid
				from game.col_trade_record final
				where ctime between ? and ? and order_status = 4 and first_pay = 1
				group by sdate, userid
			)
		) s1
		group by sdate
	`, stime, etime, stime, etime)
	if err != nil {
		return
	}
	for _, data := range agent_first_pay_withdraw_datas {
		sdate := data["sdate"].(uint32)
		withdraw_users := utils.ToInt64(data["withdraw_users"])
		withdraws := utils.ToInt64(data["withdraws"])

		if stat, ok := dateStats[sdate]; ok {
			stat.ShareFirstPayWithdrawUsers = withdraw_users
			stat.ShareFirstPayWithdrawAmounts0 = withdraws
		}
	}

	for _, stat := range stats {
		stat.DayProductCost = fmt.Sprintf("%.2f", Chip2Float(stat.DayProductCost0))
		stat.DayPayOutCost = fmt.Sprintf("%.2f", Chip2Float(stat.DayPayOutCost0))
		stat.SharePayAmounts = fmt.Sprintf("%.2f", Chip2Float(stat.SharePayAmounts0))
		stat.ShareWithdrawAmounts = fmt.Sprintf("%.2f", Chip2Float(stat.ShareWithdrawAmounts0))
		stat.ShareFirstPayAmounts = fmt.Sprintf("%.2f", Chip2Float(stat.ShareFirstPayAmounts0))

		stat.ShareFirstPaySubWithdraw0 = stat.ShareFirstPayAmounts0 - stat.ShareFirstPayWithdrawAmounts0
		stat.ShareFirstPayWithdrawAmounts = fmt.Sprintf("%.2f", Chip2Float(stat.ShareFirstPayWithdrawAmounts0))
		stat.ShareFirstPaySubWithdraw = fmt.Sprintf("%.2f", Chip2Float(stat.ShareFirstPaySubWithdraw0))
		stat.ShareFirstPayProfit = fmt.Sprintf("%.2f%%", ComputeFloat(stat.ShareFirstPaySubWithdraw0, stat.ShareFirstPayAmounts0)*100)
		stat.NewSharePayAmounts = fmt.Sprintf("%.2f", Chip2Float(stat.NewSharePayAmounts0))
		stat.NewShareWithdrawAmounts = fmt.Sprintf("%.2f", Chip2Float(stat.NewShareWithdrawAmounts0))
		stat.OldSharePayAmounts = fmt.Sprintf("%.2f", Chip2Float(stat.OldSharePayAmounts0))
		stat.OldShareWithdrawAmounts = fmt.Sprintf("%.2f", Chip2Float(stat.OldShareWithdrawAmounts0))

		stat.SharePaySubWithdraw = fmt.Sprintf("%.2f", Chip2Float(stat.SharePayAmounts0-stat.ShareWithdrawAmounts0))
		stat.ShareProfitRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.SharePayAmounts0-stat.ShareWithdrawAmounts0, stat.SharePayAmounts0)*100)
		stat.SharePayRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.SharePayUsers, stat.ShareLoginUsers)*100)
		stat.ShareWithdrawRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.ShareWithdrawUsers, stat.ShareLoginUsers)*100)
		stat.ShareARPU = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(stat.SharePayAmounts0, stat.ShareLoginUsers)))
		stat.ShareARPPU = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(stat.SharePayAmounts0, stat.SharePayUsers)))

		// 1.体系首充人数/（体系日活人数-日活中已经充值过的人数（无论啥时候充的）+体系首充人数）
		// 2.因为当日首充的人肯定是日活里面以前没有充值过的人。因此他的分母就是以前没有充值过的人。
		stat.ShareFirstPayUnpayRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.ShareFirstPayUsers,
			stat.ShareLoginUsers-stat.ShareLoginPayUsers+stat.ShareFirstPayUsers)*100)
		stat.ShareFirstPayWithdrawRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.ShareFirstPayWithdrawUsers, stat.ShareFirstPayUsers)*100)

		// 体系新用户充值
		stat.NewSharePaySubWithdraw = fmt.Sprintf("%.2f", Chip2Float(stat.NewSharePayAmounts0-stat.NewShareWithdrawAmounts0))
		stat.NewShareProfitRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.NewSharePayAmounts0-stat.NewShareWithdrawAmounts0, stat.NewSharePayAmounts0)*100)
		stat.NewSharePayRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.NewSharePayUsers, stat.NewShareLoginUsers)*100)
		stat.NewShareWithdrawRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.NewShareWithdrawUsers, stat.NewShareLoginUsers)*100)
		stat.NewShareARPU = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(stat.NewSharePayAmounts0, stat.NewShareLoginUsers)))
		stat.NewShareARPPU = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(stat.NewSharePayAmounts0, stat.NewSharePayUsers)))

		// 体系老用户充值
		stat.OldSharePaySubWithdraw = fmt.Sprintf("%.2f", Chip2Float(stat.OldSharePayAmounts0-stat.OldShareWithdrawAmounts0))
		stat.OldShareProfitRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.OldSharePayAmounts0-stat.OldShareWithdrawAmounts0, stat.OldSharePayAmounts0)*100)
		stat.OldSharePayRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.OldSharePayUsers, stat.OldShareLoginUsers)*100)
		stat.OldShareWithdrawRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.OldShareWithdrawUsers, stat.OldShareLoginUsers)*100)
		stat.OldShareARPU = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(stat.OldSharePayAmounts0, stat.OldShareLoginUsers)))
		stat.OldShareARPPU = fmt.Sprintf("%.2f", Chip2Float(ComputeFloat(stat.OldSharePayAmounts0, stat.OldSharePayUsers)))
	}

	return
}

// 代理后台-代理活动经济日览
func (s *activityService) ShareAgentFinanceDates(stime, etime time.Time) (stats []*entity.ShareAgentFinanceDate, err error) {
	dateStats := make(map[uint32]*entity.ShareAgentFinanceDate)
	for begin := stime; begin.Before(etime); begin = begin.AddDate(0, 0, 1) {
		day_time := uint32(begin.Year()*10000 + int(begin.Month())*100 + begin.Day())
		sday := fmt.Sprint(day_time)
		sdate := sday[0:4] + "-" + sday[4:6] + "-" + sday[6:8]
		stat := &entity.ShareAgentFinanceDate{SDate: sdate}
		stats = append(stats, stat)
		dateStats[day_time] = stat
	}
	utils.SliceReverse(stats)

	// 产出/实付成本
	costMap, err := s.shareAgentCost(stime, etime, false)
	if err != nil {
		return
	}
	for date, stat := range dateStats {
		if cost, ok := costMap[date]; ok {
			stat.DayProductCost0 = cost.DayProductCost
			stat.DayPayOutCost0 = cost.DayPayOutCost
		}
	}

	var share_user_datas []map[string]any
	err = ck.Select(&share_user_datas, `
		select toYYYYMMDD(s0.ctime) sdate, s1.team_lv, count(*) users
		from game.col_user s1 final
		join game.col_user s0 final on s0.share_superior = s1.userid
		where s0.ctime between ? and ? and s0.share_superior != ''
		group by sdate, s1.team_lv
	`, stime, etime)
	if err != nil {
		return
	}
	for _, data := range share_user_datas {
		sdate := data["sdate"].(uint32)
		team_lv := utils.ToInt64(data["team_lv"])
		users := utils.ToInt64(data["users"])
		if stat, ok := dateStats[sdate]; ok {
			stat.ShareUsers += users
			switch team_lv {
			case 1:
				stat.ShareTeamV1Users += users
			case 2:
				stat.ShareTeamV2Users += users
			case 3:
				stat.ShareTeamV3Users += users
			case 4:
				stat.ShareTeamV4Users += users
			}
		}
	}

	// 当日领走奖励
	var income_take_datas []map[string]any
	err = ck.Select(&income_take_datas, `
		select toYYYYMMDD(toDateTime(ctime/1000)) sdate, itype, SUM(amount) amounts
		from game.col_activity_share_income_record_tack_log final
		where ctime between ? and ? and itype in (1,2,3)
		group by sdate, itype
		order by sdate, itype
	`, stime.UnixMilli(), etime.UnixMilli())
	if err != nil {
		return
	}
	for _, data := range income_take_datas {
		sdate := data["sdate"].(uint32)
		itype := utils.ToInt64(data["itype"])
		amounts := utils.ToInt64(data["amounts"])
		if stat, ok := dateStats[sdate]; ok {
			// 1.打码奖励,2人数人头奖励,3.累计任务人头奖励
			switch itype {
			case 1:
				stat.ShareBetsRewardTake0 += amounts
			case 2:
				stat.ShareHeadRewardTake0 += amounts
			case 3:
				stat.ShareTaskRewardTake0 += amounts
			}
		}
	}

	// 当日产生奖励
	var income_datas []map[string]any
	err = ck.Select(&income_datas, `
		select toYYYYMMDD(toDateTime(ctime/1000)) sdate, itype, (case when team_lv = 0 then 1 else team_lv end) team_lv0, count(*) records, (SUM(amount) / 100) amounts
		from game.col_activity_share_income_record final
		where ctime between ? and ? and itype in (1,2,3,4)
		group by sdate, itype, team_lv0
		order by sdate desc, itype, team_lv0
	`, stime.UnixMilli(), etime.UnixMilli())
	if err != nil {
		return
	}
	for _, data := range income_datas {
		sdate := data["sdate"].(uint32)
		itype := utils.ToInt64(data["itype"])
		team_lv := utils.ToInt64(data["team_lv0"])
		amounts := utils.ToInt64(data["amounts"])
		records := utils.ToInt64(data["records"])
		if stat, ok := dateStats[sdate]; ok {
			//  1.打码奖励,2.人数人头奖励,3.累计任务人头奖励,4.受邀者人头奖
			switch itype {
			case 1:
				stat.ShareBetsReward0 += amounts
				stat.ShareBetsReward20 += amounts
				switch team_lv {
				case 1:
					stat.ShareTeamV1BetsReward0 += amounts
				case 2:
					stat.ShareTeamV2BetsReward0 += amounts
				case 3:
					stat.ShareTeamV3BetsReward0 += amounts
				case 4:
					stat.ShareTeamV4BetsReward0 += amounts
				}
			case 2:
				stat.ShareHeadReward0 += amounts
				stat.ShareHeadReward20 += amounts
				stat.ShareHeadRewardSuper0 += amounts
				stat.ShareInvalidUsers += records
				switch team_lv {
				case 1:
					stat.ShareTeamV1HeadReward0 += amounts
					stat.ShareTeamV1InvalidUsers += records
				case 2:
					stat.ShareTeamV2HeadReward0 += amounts
					stat.ShareTeamV2InvalidUsers += records
				case 3:
					stat.ShareTeamV3HeadReward0 += amounts
					stat.ShareTeamV3InvalidUsers += records
				case 4:
					stat.ShareTeamV4HeadReward0 += amounts
					stat.ShareTeamV4InvalidUsers += records
				}
			case 4:
				stat.ShareHeadReward0 += amounts
				stat.ShareHeadReward20 += amounts
				stat.ShareHeadRewardChild0 += amounts
				stat.ShareHeadRewardTake0 += amounts // 受邀者奖金直接发放
				switch team_lv {
				case 1:
					stat.ShareTeamV1HeadReward0 += amounts
				case 2:
					stat.ShareTeamV2HeadReward0 += amounts
				case 3:
					stat.ShareTeamV3HeadReward0 += amounts
				case 4:
					stat.ShareTeamV4HeadReward0 += amounts
				}
			case 3:
				stat.ShareTaskReward0 += amounts
				stat.ShareTaskReward20 += amounts
				switch team_lv {
				case 1:
					stat.ShareTeamV1TaskReward0 += amounts
				case 2:
					stat.ShareTeamV2TaskReward0 += amounts
				case 3:
					stat.ShareTeamV3TaskReward0 += amounts
				case 4:
					stat.ShareTeamV4TaskReward0 += amounts
				}
			}
		}
	}

	for _, stat := range stats {
		stat.DayProductCost = fmt.Sprintf("%.2f", Chip2Float(stat.DayProductCost0))
		stat.DayPayOutCost = fmt.Sprintf("%.2f", Chip2Float(stat.DayPayOutCost0))
		stat.ShareBetsReward = fmt.Sprintf("%.2f", Chip2Float(stat.ShareBetsReward0))
		stat.ShareBetsRewardTake = fmt.Sprintf("%.2f", Chip2Float(stat.ShareBetsRewardTake0))
		stat.ShareHeadReward = fmt.Sprintf("%.2f", Chip2Float(stat.ShareHeadReward0))
		stat.ShareHeadRewardTake = fmt.Sprintf("%.2f", Chip2Float(stat.ShareHeadRewardTake0))
		stat.ShareTaskReward = fmt.Sprintf("%.2f", Chip2Float(stat.ShareTaskReward0))
		stat.ShareTaskRewardTake = fmt.Sprintf("%.2f", Chip2Float(stat.ShareTaskRewardTake0))

		stat.ShareHeadReward2 = fmt.Sprintf("%.2f", Chip2Float(stat.ShareHeadReward20))
		stat.ShareHeadRewardSuper = fmt.Sprintf("%.2f", Chip2Float(stat.ShareHeadRewardSuper0))
		stat.ShareHeadRewardChild = fmt.Sprintf("%.2f", Chip2Float(stat.ShareHeadRewardChild0))
		stat.ShareTeamV1HeadReward = fmt.Sprintf("%.2f", Chip2Float(stat.ShareTeamV1HeadReward0))
		stat.ShareTeamV2HeadReward = fmt.Sprintf("%.2f", Chip2Float(stat.ShareTeamV2HeadReward0))
		stat.ShareTeamV3HeadReward = fmt.Sprintf("%.2f", Chip2Float(stat.ShareTeamV3HeadReward0))
		stat.ShareTeamV4HeadReward = fmt.Sprintf("%.2f", Chip2Float(stat.ShareTeamV4HeadReward0))

		stat.ShareTaskReward2 = fmt.Sprintf("%.2f", Chip2Float(stat.ShareTaskReward20))
		stat.ShareTeamV1TaskReward = fmt.Sprintf("%.2f", Chip2Float(stat.ShareTeamV1TaskReward0))
		stat.ShareTeamV2TaskReward = fmt.Sprintf("%.2f", Chip2Float(stat.ShareTeamV2TaskReward0))
		stat.ShareTeamV3TaskReward = fmt.Sprintf("%.2f", Chip2Float(stat.ShareTeamV3TaskReward0))
		stat.ShareTeamV4TaskReward = fmt.Sprintf("%.2f", Chip2Float(stat.ShareTeamV4TaskReward0))

		stat.ShareBetsReward2 = fmt.Sprintf("%.2f", Chip2Float(stat.ShareBetsReward20))
		stat.ShareTeamV1BetsReward = fmt.Sprintf("%.2f", Chip2Float(stat.ShareTeamV1BetsReward0))
		stat.ShareTeamV2BetsReward = fmt.Sprintf("%.2f", Chip2Float(stat.ShareTeamV2BetsReward0))
		stat.ShareTeamV3BetsReward = fmt.Sprintf("%.2f", Chip2Float(stat.ShareTeamV3BetsReward0))
		stat.ShareTeamV4BetsReward = fmt.Sprintf("%.2f", Chip2Float(stat.ShareTeamV4BetsReward0))
	}
	return
}

// 代理后台-代理头目明细
func (s *activityService) ShareAgentSuperStats(statTotal bool, page, pageSize int, stime, etime time.Time, superStime, superEtime *time.Time, rank string, asc bool) (total int64, list []*entity.ShareAgentSuperStat, err error) {
	var having_filter_args []any
	var having_filter string
	if superStime != nil && superEtime != nil {
		having_filter_args = append(having_filter_args, *superStime, *superEtime)
		having_filter = `
			having first_be_super_time between ? and ?
		`
	}

	sql_select_count := `select count(*) total`
	sql_select := `
		select s1.super_id super_id, super_ctime, super_pays, super_cash_out, team_lv, first_be_super_time, grow_days, last_login_days, 
			group_users, group_pay_users, group_pays, group_cash_out,
			prod_amounts, take_amounts, 
			prod_amounts_bet, prod_amounts_head, prod_amounts_task, prod_amounts_child,
			take_amounts_bet, take_amounts_head, take_amounts_task, prod_amounts_child take_amounts_child,
			turn_prize, take_turn_prize, (take_amounts + take_turn_prize) team_take_prizes,
			(case when team_take_prizes = 0 then 0 else (((super_pays - super_cash_out) + (group_pays - group_cash_out)) / team_take_prizes) end) super_roi,
			(case when team_take_prizes = 0 then 0 else ((group_pays - group_cash_out) / team_take_prizes) end) team_roi,
			(super_pays - super_cash_out) super_profit, (case when super_pays = 0 then 0 else (super_profit / super_pays) end) super_profit_rate,
			(group_pay_users / group_users) group_pay_rate, 
			(group_pays - group_cash_out) group_profit, (case when group_pays = 0 then 0 else (group_profit / group_pays) end) group_profit_rate,
			(group_pays / group_users) group_arpu, (case when group_pay_users = 0 then 0 else (group_pays / group_pay_users) end) group_arppu,
			super_bets, group_bets,
			(case when super_pays = 0 then 0 else (super_bets / super_pays) end) super_bets_pay_rate,
			(case when group_pays = 0 then 0 else (group_bets / group_pays) end) group_bets_pay_rate
	`

	sql0 := "%s" + fmt.Sprintf(`
		from (
			select s0.userid super_id, s0.ctime super_ctime, s0.money super_pays, s0.cash_out super_cash_out, s0.team_lv, s1.group_bets,
				s1.first_be_super_time, dateDiff('hour', s0.ctime, s1.first_be_super_time)/24 grow_days,
				dateDiff('day', s1.group_last_login_time, toStartOfDay(now())) group_last_login_days,
				dateDiff('day', s0.login_time, toStartOfDay(now())) super_last_login_days,
				(case when group_last_login_days < super_last_login_days then group_last_login_days else super_last_login_days end) last_login_days,
				s1.group_users, s1.group_pays, s1.group_cash_out, s1.group_pay_users
			from game.col_user s0 final
			join (
				WITH RECURSIVE share_cte AS (
					select s0.userid super0, s0.userid super_id, s1.userid userid, 1 lv
					from game.col_user s0 final
					join game.col_user s1 final on s0.userid = s1.share_superior and s1.share_superior != ''
					where s0.ctime between ? and ?
					
					union all
					
					select cte.super0, cte.userid super_id, s3.userid userid, cte.lv+1 lv
					FROM game.col_user s3 FINAL
					JOIN share_cte cte ON s3.share_superior = cte.userid AND cte.lv < 3
				)
				SELECT s1.super0, count(*) group_users, SUM(s0.money) group_pays, SUM(s0.cash_out) group_cash_out,
					SUM(case when s0.money > 0 then 1 else 0 end) group_pay_users,
					MIN(s0.ctime) first_be_super_time,
					max(s0.login_time) group_last_login_time,
					SUM(s3.bets0) group_bets
				FROM game.col_user s0 final
				join share_cte s1 on s0.userid = s1.userid
				left join (
					select userid, SUM(bets) bets0 from (
						select user_id userid, SUM(amount) bets from game.col_nsq_log_external_bet final 
						where ctime > ?
						group by user_id
							union all
						select userid, SUM(bet_amount) bets from game.col_detail final 
						where begin_time > ? and robot = 0
						group by userid
					) t0
					group by userid
				) s3 on s1.userid = s3.userid
				group by s1.super0
				%s
			) s1 on s0.userid = s1.super0
		) s1 left join (
			select super_id, SUM(prod_amounts) prod_amounts, SUM(take_amounts + t0.prod_amounts_child) take_amounts,
				SUM(prod_amounts_bet) prod_amounts_bet, SUM(prod_amounts_head) prod_amounts_head, SUM(prod_amounts_task) prod_amounts_task, SUM(prod_amounts_child) prod_amounts_child,
				SUM(take_amounts_bet) take_amounts_bet, SUM(take_amounts_head) take_amounts_head, SUM(take_amounts_task) take_amounts_task,
				SUM(turn_prize) turn_prize, SUM(take_turn_prize) take_turn_prize
			from (
				select super_id, 0 prod_amounts, SUM(amount) take_amounts,
					0 prod_amounts_bet, 0 prod_amounts_head, 0 prod_amounts_task, 0 prod_amounts_child,
					SUM(case when itype = 1 then amount else 0 end) take_amounts_bet,
					SUM(case when itype = 2 then amount else 0 end) take_amounts_head,
					SUM(case when itype = 3 then amount else 0 end) take_amounts_task,
					0 turn_prize, 0 take_turn_prize
				from game.col_activity_share_income_record_tack_log final
				where ctime > ?
				group by super_id
					union all
				select super_id, SUM(amount)/100 prod_amounts, 0 take_amounts,
					SUM(case when itype = 1 then amount else 0 end)/100 prod_amounts_bet,
					SUM(case when itype = 2 then amount else 0 end)/100 prod_amounts_head,
					SUM(case when itype = 3 then amount else 0 end)/100 prod_amounts_task,
					SUM(case when itype = 4 then amount else 0 end)/100 prod_amounts_child,
					0 take_amounts_bet, 0 take_amounts_head, 0 take_amounts_task,
					0 turn_prize, 0 take_turn_prize
				from game.col_activity_share_income_record final
				where ctime > ?
				group by super_id
					union all
				select userid super_id, 0 prod_amounts, 0 take_amounts,
					0 prod_amounts_bet, 0 prod_amounts_head, 0 prod_amounts_task, 0 prod_amounts_child,
					0 take_amounts_bet, 0 take_amounts_head, 0 take_amounts_task,
					SUM(prize) turn_prize, SUM(case when state = 1 then prize else 0 end) take_turn_prize
				from game.col_activity_turn_prize_log final
				where ctime > ?
				group by userid
			) t0
			group by super_id
		) s2 on s1.super_id = s2.super_id
		left join (
			select userid, SUM(bets) super_bets from (
				select user_id userid, SUM(amount) bets from game.col_nsq_log_external_bet final 
				where ctime > ?
				group by user_id
					union all
				select userid, SUM(bet_amount) bets from game.col_detail final 
				where begin_time > ? and robot = 0 and win_type in (1,2,3)
				group by userid
			) t0
			group by userid
		) s3 on s1.super_id = s3.userid
	`, having_filter)
	args0 := []any{stime, etime, stime.Unix(), stime.Unix()}
	args0 = append(args0, having_filter_args...)
	args0 = append(args0, stime.UnixMilli(), stime.UnixMilli(), stime.Unix(), stime.Unix(), stime.Unix())

	if statTotal {
		sql_total := fmt.Sprintf(sql0, sql_select_count)
		err = ck.Select(&total, sql_total, args0...)
		if err != nil {
			return
		}
		if total == 0 {
			return
		}
	}

	offset, limit := PageCalc(page, pageSize)
	var orderBy = "super_ctime"
	var sort = "desc"
	if rank != "" {
		orderBy = rank
	}
	if asc {
		sort = "asc"
	}
	sql_query := fmt.Sprintf(sql0, sql_select) + fmt.Sprintf(`
		order by %s %s
		limit %d, %d
	`, orderBy, sort, offset, limit)

	var datas []map[string]any
	err = ck.Select(&datas, sql_query, args0...)
	if err != nil {
		return
	}

	for _, data := range datas {
		super_id := data["super_id"].(string)
		super_ctime := data["super_ctime"].(time.Time)
		super_pays := utils.ToInt64(data["super_pays"])
		super_cash_out := utils.ToInt64(data["super_cash_out"])
		team_lv := utils.ToInt64(data["team_lv"])
		first_be_super_time := data["first_be_super_time"].(time.Time)
		grow_days := utils.ToFloat64(data["grow_days"])
		last_login_days := utils.ToInt64(data["last_login_days"])
		group_users := utils.ToInt64(data["group_users"])
		group_pay_users := utils.ToInt64(data["group_pay_users"])
		group_pays := utils.ToInt64(data["group_pays"])
		group_cash_out := utils.ToInt64(data["group_cash_out"])

		prod_amounts_bet := utils.ToInt64(data["prod_amounts_bet"])
		prod_amounts_head := utils.ToInt64(data["prod_amounts_head"])
		prod_amounts_task := utils.ToInt64(data["prod_amounts_task"])
		prod_amounts_child := utils.ToInt64(data["prod_amounts_child"])
		take_amounts_bet := utils.ToInt64(data["take_amounts_bet"])
		take_amounts_head := utils.ToInt64(data["take_amounts_head"])
		take_amounts_task := utils.ToInt64(data["take_amounts_task"])
		take_amounts_child := utils.ToInt64(data["take_amounts_child"])

		turn_prize := utils.ToInt64(data["turn_prize"])
		take_turn_prize := utils.ToInt64(data["take_turn_prize"])
		// team_take_prizes := utils.ToInt64(data["team_take_prizes"])
		super_roi := utils.ToFloat64(data["super_roi"])
		team_roi := utils.ToFloat64(data["team_roi"])
		super_profit := utils.ToInt64(data["super_profit"])
		super_profit_rate := utils.ToFloat64(data["super_profit_rate"])
		group_pay_rate := utils.ToFloat64(data["group_pay_rate"])
		group_profit := utils.ToInt64(data["group_profit"])
		group_profit_rate := utils.ToFloat64(data["group_profit_rate"])
		group_arpu := utils.ToFloat64(data["group_arpu"])
		group_arppu := utils.ToFloat64(data["group_arppu"])
		super_bets := utils.ToInt64(data["super_bets"])
		group_bets := utils.ToInt64(data["group_bets"])
		super_bets_pay_rate := utils.ToFloat64(data["super_bets_pay_rate"])
		group_bets_pay_rate := utils.ToFloat64(data["group_bets_pay_rate"])

		stat := &entity.ShareAgentSuperStat{
			SuperId: super_id,
		}
		stat.SuperCtime = super_ctime.In(Location()).Format(utils.FORMAT2)
		stat.SuperPays = fmt.Sprintf("%.2f", Chip2Float(super_pays))
		stat.SuperCashOut = fmt.Sprintf("%.2f", Chip2Float(super_cash_out))
		stat.TeamLv = team_lv
		stat.BeSuperTime = first_be_super_time.In(Location()).Format(utils.FORMAT2)
		stat.GrowDays = fmt.Sprintf("%.1f", grow_days)
		stat.LastLoginDays = last_login_days
		stat.GroupUsers = group_users
		stat.GroupPayUsers = group_pay_users
		stat.GroupPays = fmt.Sprintf("%.2f", Chip2Float(group_pays))
		stat.GroupCashOut = fmt.Sprintf("%.2f", Chip2Float(group_cash_out))

		stat.ProdAmountsBet = fmt.Sprintf("%.2f", Chip2Float(prod_amounts_bet))
		stat.ProdAmountsHead = fmt.Sprintf("%.2f", Chip2Float(prod_amounts_head))
		stat.ProdAmountsTask = fmt.Sprintf("%.2f", Chip2Float(prod_amounts_task))
		stat.ProdAmountsChild = fmt.Sprintf("%.2f", Chip2Float(prod_amounts_child))
		stat.TakeAmountsBet = fmt.Sprintf("%.2f", Chip2Float(take_amounts_bet))
		stat.TakeAmountsHead = fmt.Sprintf("%.2f", Chip2Float(take_amounts_head))
		stat.TakeAmountsTask = fmt.Sprintf("%.2f", Chip2Float(take_amounts_task))
		stat.TakeAmountsChild = fmt.Sprintf("%.2f", Chip2Float(take_amounts_child))

		stat.TurnPrize = fmt.Sprintf("%.2f", Chip2Float(turn_prize))
		stat.TakeTurnPrize = fmt.Sprintf("%.2f", Chip2Float(take_turn_prize))
		// stat.TeamTakePrizes = fmt.Sprintf("%.2f", Chip2Float(team_take_prizes))
		stat.SuperROI = fmt.Sprintf("%.2f", Chip2Float(super_roi))
		stat.TeamROI = fmt.Sprintf("%.2f", Chip2Float(team_roi))
		stat.SuperProfit = fmt.Sprintf("%.2f", Chip2Float(super_profit))
		stat.SuperProfitRate = fmt.Sprintf("%.2f%%", super_profit_rate*100)
		stat.GroupPayRate = fmt.Sprintf("%.2f%%", group_pay_rate*100)
		stat.GroupProfit = fmt.Sprintf("%.2f", Chip2Float(group_profit))
		stat.GroupProfitRate = fmt.Sprintf("%.2f%%", group_profit_rate*100)
		stat.GroupArpu = fmt.Sprintf("%.2f", Chip2Float(group_arpu))
		stat.GroupArppu = fmt.Sprintf("%.2f", Chip2Float(group_arppu))

		stat.SuperBets = fmt.Sprintf("%.2f", Chip2Float(super_bets))
		stat.GroupBets = fmt.Sprintf("%.2f", Chip2Float(group_bets))

		stat.SuperBetsPayRate = fmt.Sprintf("%.2f", super_bets_pay_rate)
		stat.GroupBetsPayRate = fmt.Sprintf("%.2f", group_bets_pay_rate)

		list = append(list, stat)
	}
	return
}

// 代理后台-代理头目奖金日排名
func (s *activityService) ShareAgentSuperBonusStats(statTotal bool, page, pageSize int, stime, etime time.Time, rank string, asc bool) (total int64, list []*entity.ShareAgentSuperBonusStats, err error) {
	sql_select_count := `
		select count(*)
	`
	sql_select := `
		select s1.sdate sdate, s1.userid super_id, s0.pays group_pays, s0.withdraws group_withdraws, 
			s1.pays pays, s1.withdraws withdraws,
			s1.prod_amounts0 prod_amounts, s1.take_amounts0 take_amounts,
			s1.prod_amounts_bet, s1.prod_amounts_head, s1.prod_amounts_task, s1.prod_amounts_child,
			s1.take_amounts_bet, s1.take_amounts_head, s1.take_amounts_task, s1.prod_amounts_child take_amounts_child,
			s1.turn_prize, s1.take_turn_prize,
			(case when take_amounts = 0 then 0 else (((s1.pays-s1.withdraws)+(s0.pays-s0.withdraws)) / take_amounts) end) super_roi,
			(case when take_amounts = 0 then 0 else ((s0.pays-s0.withdraws) / take_amounts) end) team_roi
	`
	sql0 := `
		%s
		from (
			WITH RECURSIVE share_cte AS (
				select s1.userid super0, s1.userid super_id, s0.userid userid, 1 lv
				from game.col_user s0 final
				join (
					select userid from (
						select super_id userid
						from game.col_activity_share_income_record final
						where ctime between ? and ?
						group by super_id
							union all
						select super_id userid
						from game.col_activity_share_income_record_tack_log final
						where ctime between ? and ?
						group by super_id
							union all
						select userid
						from game.col_activity_turn_prize_log final
						where ctime between ? and ?
						group by userid
					) t0 
					group by userid
				) s1 on s0.share_superior = s1.userid and s0.share_superior != ''
				
				union all
				
				select cte.super0, cte.userid super_id, s3.userid userid, cte.lv+1 lv
				FROM game.col_user s3 FINAL
				JOIN share_cte cte ON s3.share_superior = cte.userid AND cte.lv < 3
			)
			select s0.super0, s1.sdate, SUM(s1.pays) pays, SUM(s1.withdraws) withdraws
			from share_cte s0
			join (
				select sdate, userid, SUM(pays) pays, SUM(withdraws) withdraws
				from (
					select toYYYYMMDD(ctime) sdate, userid, SUM(amount) pays, 0 withdraws
					from game.col_trade_record final
					where ctime between ? and ? and order_status = 4
					group by sdate, userid
						union all
					select toYYYYMMDD(ctime) sdate, userid, 0 pays, SUM(amount) withdraws
					from game.col_withdraw_record final
					where ctime between ? and ? and order_status = 2
					group by sdate, userid
				) s1
				group by sdate, userid
			) s1 on s0.userid = s1.userid
			group by s0.super0, s1.sdate
			order by s1.sdate, s0.super0
		) s0 right join (
			select s1.sdate, s1.userid, SUM(s1.pays) pays, SUM(s1.withdraws) withdraws,
				SUM(s1.prod_amounts) prod_amounts0, SUM(s1.take_amounts + s1.prod_amounts_child) take_amounts0,
				SUM(prod_amounts_bet) prod_amounts_bet, SUM(prod_amounts_head) prod_amounts_head, SUM(prod_amounts_task) prod_amounts_task, SUM(prod_amounts_child) prod_amounts_child,
				SUM(take_amounts_bet) take_amounts_bet, SUM(take_amounts_head) take_amounts_head, SUM(take_amounts_task) take_amounts_task,
				SUM(s1.turn_prize) turn_prize, SUM(s1.take_turn_prize) take_turn_prize
			from (
				select toYYYYMMDD(ctime) sdate, userid, SUM(amount) pays, 0 withdraws,
					0 prod_amounts, 0 take_amounts,
					0 prod_amounts_bet, 0 prod_amounts_head, 0 prod_amounts_task, 0 prod_amounts_child,
					0 take_amounts_bet, 0 take_amounts_head, 0 take_amounts_task,
					0 turn_prize, 0 take_turn_prize
				from game.col_trade_record final
				where ctime between ? and ? and order_status = 4
				group by sdate, userid
					union all
				select toYYYYMMDD(ctime) sdate, userid, 0 pays, SUM(amount) withdraws,
					0 prod_amounts, 0 take_amounts,
					0 prod_amounts_bet, 0 prod_amounts_head, 0 prod_amounts_task, 0 prod_amounts_child,
					0 take_amounts_bet, 0 take_amounts_head, 0 take_amounts_task,
					0 turn_prize, 0 take_turn_prize
				from game.col_withdraw_record final
				where ctime between ? and ? and order_status = 2
				group by sdate, userid
					union all
				select toYYYYMMDD(toDateTime(ctime/1000)) sdate, super_id userid, 0 pays, 0 withdraws,
					SUM(amount)/100 prod_amounts, 0 take_amounts,
					SUM(case when itype = 1 then amount else 0 end)/100 prod_amounts_bet,
					SUM(case when itype = 2 then amount else 0 end)/100 prod_amounts_head,
					SUM(case when itype = 3 then amount else 0 end)/100 prod_amounts_task,
					SUM(case when itype = 4 then amount else 0 end)/100 prod_amounts_child,
					0 take_amounts_bet, 0 take_amounts_head, 0 take_amounts_task,
					0 turn_prize, 0 take_turn_prize
				from game.col_activity_share_income_record final
				where ctime between ? and ?
				group by sdate, super_id
					union all
				select toYYYYMMDD(toDateTime(ctime/1000)) sdate, super_id userid, 0 pays, 0 withdraws,
						0 prod_amounts, SUM(amount) take_amounts,
						0 prod_amounts_bet, 0 prod_amounts_head, 0 prod_amounts_task, 0 prod_amounts_child,
						SUM(case when itype = 1 then amount else 0 end) take_amounts_bet,
						SUM(case when itype = 2 then amount else 0 end) take_amounts_head,
						SUM(case when itype = 3 then amount else 0 end) take_amounts_task,
						0 turn_prize, 0 take_turn_prize
				from game.col_activity_share_income_record_tack_log final
				where ctime between ? and ?
				group by sdate, super_id
					union all
				select toYYYYMMDD(toDateTime(ctime)) sdate, userid, 0 pays, 0 withdraws,
					0 prod_amounts, 0 take_amounts,
					0 prod_amounts_bet, 0 prod_amounts_head, 0 prod_amounts_task, 0 prod_amounts_child,
					0 take_amounts_bet, 0 take_amounts_head, 0 take_amounts_task,
					SUM(prize) turn_prize, SUM(case when state = 1 then prize else 0 end) take_turn_prize
				from game.col_activity_turn_prize_log final
				where ctime between ? and ?
				group by sdate, userid
			) s1
			group by s1.sdate, s1.userid
			having prod_amounts0 > 0 or take_amounts0 > 0
		) s1 on s0.sdate = s1.sdate and s0.super0 = s1.userid
	`
	args0 := []any{
		stime.UnixMilli(), etime.UnixMilli(), stime.UnixMilli(), etime.UnixMilli(),
		stime.Unix(), etime.Unix(),
		stime, etime, stime, etime,
		stime, etime, stime, etime,
		stime.UnixMilli(), etime.UnixMilli(), stime.UnixMilli(), etime.UnixMilli(),
		stime.Unix(), etime.Unix(),
	}

	if statTotal {
		sql_total := fmt.Sprintf(sql0, sql_select_count)
		err = ck.Select(&total, sql_total, args0...)
		if err != nil {
			return
		}
		if total == 0 {
			return
		}
	}

	offset, limit := PageCalc(page, pageSize)
	var orderBy = "prod_amounts"
	var sort = "desc"
	if rank != "" {
		orderBy = rank
	}
	if asc {
		sort = "asc"
	}
	sql_query := fmt.Sprintf(sql0, sql_select) + fmt.Sprintf(`
		order by %s %s
		limit %d, %d
	`, orderBy, sort, offset, limit)

	var datas []map[string]any
	err = ck.Select(&datas, sql_query, args0...)
	if err != nil {
		return
	}

	for _, data := range datas {
		super_id := data["super_id"].(string)
		sdate := data["sdate"].(uint32)
		// group_pays := utils.ToInt64(data["group_pays"])
		// group_withdraws := utils.ToInt64(data["group_withdraws"])
		// pays := utils.ToInt64(data["pays"])
		// withdraws := utils.ToInt64(data["withdraws"])
		prod_amounts := utils.ToInt64(data["prod_amounts"])
		take_amounts := utils.ToInt64(data["take_amounts"])
		prod_amounts_bet := utils.ToInt64(data["prod_amounts_bet"])
		prod_amounts_head := utils.ToInt64(data["prod_amounts_head"])
		prod_amounts_task := utils.ToInt64(data["prod_amounts_task"])
		prod_amounts_child := utils.ToInt64(data["prod_amounts_child"])
		take_amounts_bet := utils.ToInt64(data["take_amounts_bet"])
		take_amounts_head := utils.ToInt64(data["take_amounts_head"])
		take_amounts_task := utils.ToInt64(data["take_amounts_task"])
		take_amounts_child := utils.ToInt64(data["take_amounts_child"])
		turn_prize := utils.ToInt64(data["turn_prize"])
		take_turn_prize := utils.ToInt64(data["take_turn_prize"])
		super_roi := utils.ToFloat64(data["super_roi"])
		team_roi := utils.ToFloat64(data["team_roi"])

		sday := fmt.Sprint(sdate)
		fdate := sday[0:4] + "-" + sday[4:6] + "-" + sday[6:8]

		stat := &entity.ShareAgentSuperBonusStats{
			SDate:   fdate,
			SuperId: super_id,
		}
		list = append(list, stat)
		stat.SuperROI = fmt.Sprintf("%.2f", super_roi)
		stat.TeamROI = fmt.Sprintf("%.2f", team_roi)
		stat.ProdAmounts = fmt.Sprintf("%.2f", Chip2Float(prod_amounts))
		stat.TakeAmounts = fmt.Sprintf("%.2f", Chip2Float(take_amounts))
		stat.ProdAmountsBet = fmt.Sprintf("%.2f", Chip2Float(prod_amounts_bet))
		stat.TurnPrize = fmt.Sprintf("%.2f", Chip2Float(turn_prize))
		stat.ProdAmountsHead = fmt.Sprintf("%.2f", Chip2Float(prod_amounts_head))
		stat.ProdAmountsChild = fmt.Sprintf("%.2f", Chip2Float(prod_amounts_child))
		stat.ProdAmountsTask = fmt.Sprintf("%.2f", Chip2Float(prod_amounts_task))
		stat.TakeAmountsBet = fmt.Sprintf("%.2f", Chip2Float(take_amounts_bet))
		stat.TakeTurnPrize = fmt.Sprintf("%.2f", Chip2Float(take_turn_prize))
		stat.TakeAmountsHead = fmt.Sprintf("%.2f", Chip2Float(take_amounts_head))
		stat.TakeAmountsChild = fmt.Sprintf("%.2f", Chip2Float(take_amounts_child))
		stat.TakeAmountsTask = fmt.Sprintf("%.2f", Chip2Float(take_amounts_task))
	}
	return
}

// 代理后台-代理团队每日数量变化
func (s *activityService) ShareAgentGroupVary(stime, etime time.Time) (stats []*entity.ShareAgentGroupVary, err error) {
	days := int(etime.Sub(stime).Hours() / 24)
	startTime, _ := time.ParseInLocation(utils.FORMAT, "2025-03-06 00:00:00", Location()) // 代理活动开始时间
	sql0 := `
		select sdate,
			SUM(case when group_lv = 4 then 1 else 0 end) lv4_groups,
			SUM(case when group_lv = 3 then 1 else 0 end) lv3_groups,
			SUM(case when group_lv = 2 then 1 else 0 end) lv2_groups,
			SUM(case when group_lv = 1 then 1 else 0 end) lv1_groups,
			SUM(case when group_lv = 4 then group_users else 0 end) lv4_group_users,
			SUM(case when group_lv = 3 then group_users else 0 end) lv3_group_users,
			SUM(case when group_lv = 2 then group_users else 0 end) lv2_group_users,
			SUM(case when group_lv = 1 then group_users else 0 end) lv1_group_users,
			SUM(case when group_last_login_days <= 7 and group_lv = 4 then 1 else 0 end) live_lv4_groups,
			SUM(case when group_last_login_days <= 7 and group_lv = 3 then 1 else 0 end) live_lv3_groups,
			SUM(case when group_last_login_days <= 7 and group_lv = 2 then 1 else 0 end) live_lv2_groups,
			SUM(case when group_last_login_days <= 7 and group_lv = 1 then 1 else 0 end) live_lv1_groups,
			SUM(case when group_last_login_days <= 7 and group_lv = 4 then group_users else 0 end) live_lv4_group_users,
			SUM(case when group_last_login_days <= 7 and group_lv = 3 then group_users else 0 end) live_lv3_group_users,
			SUM(case when group_last_login_days <= 7 and group_lv = 2 then group_users else 0 end) live_lv2_group_users,
			SUM(case when group_last_login_days <= 7 and group_lv = 1 then group_users else 0 end) live_lv1_group_users
		from (
			WITH RECURSIVE share_cte AS (
				select s0.userid super0, s0.userid super_id, s1.userid userid, toYYYYMMDD(s1.ctime) user_cdate, 1 lv, s0.login_time super_login_time, s1.login_time user_login_time
				from game.col_user s0 final
				join game.col_user s1 final on s0.userid = s1.share_superior and s1.share_superior != ''
				where s1.ctime > ?
				
				union all
				
				select cte.super0, cte.userid super_id, s3.userid userid, toYYYYMMDD(s3.ctime) user_cdate, cte.lv+1 lv, cte.super_login_time, s3.login_time user_login_time
				FROM game.col_user s3 FINAL
				JOIN share_cte cte ON s3.share_superior = cte.userid AND cte.lv < 3
			)
			select s0.stime, s0.sdate sdate, s1.super0, count(*) group_users, SUM(s2.bets) bets, 
				max((case when super_login_time > user_login_time then super_login_time else user_login_time end)) group_login_time,
				dateDiff('day', group_login_time, toStartOfDay(now())) group_last_login_days,
				(case when group_users >= 100 and bets >= 2000000000 then 1 else 0 end) lv4,
				(case when group_users >= 30 and bets >= 200000000 then 1 else 0 end) lv3,
				(case when group_users >= 5 and bets >= 20000000 then 1 else 0 end) lv2,
				(lv4 + lv3 + lv2 + 1) group_lv
			from share_cte s1
			left join (
				select sdate, userid, SUM(bets) bets from (
					select toYYYYMMDD(toDateTime(begin_time)) sdate, userid, SUM(bet_amount) bets from game.col_detail s2 final where begin_time > ?
					group by sdate, userid
					union all
					select toYYYYMMDD(toDateTime(ctime)) sdate, user_id userid, SUM(amount) bets from game.col_nsq_log_external_bet s2 final where ctime > ?
					group by sdate, user_id
				) t0
				group by sdate, userid
			) s2 on s1.userid = s2.userid
			join (
				SELECT dateAdd(DAY, generate_series, ?) stime, toYYYYMMDD(stime) sdate
				FROM generate_series(0, ?)
			) s0 on 1 = 1
			where s1.user_cdate <= s0.sdate and s2.sdate <= s0.sdate
			group by s0.stime, s0.sdate, s1.super0
		) t0
		group by sdate
		order by sdate desc
	`
	var args0 = []any{startTime, startTime.Unix(), startTime.Unix(), stime, days}

	var datas []map[string]any
	err = ck.Select(&datas, sql0, args0...)
	if err != nil {
		return
	}

	for _, data := range datas {
		sdate := data["sdate"].(uint32)
		lv4_groups := utils.ToInt64(data["lv4_groups"])
		lv3_groups := utils.ToInt64(data["lv3_groups"])
		lv2_groups := utils.ToInt64(data["lv2_groups"])
		lv1_groups := utils.ToInt64(data["lv1_groups"])
		lv4_group_users := utils.ToInt64(data["lv4_group_users"])
		lv3_group_users := utils.ToInt64(data["lv3_group_users"])
		lv2_group_users := utils.ToInt64(data["lv2_group_users"])
		lv1_group_users := utils.ToInt64(data["lv1_group_users"])
		live_lv4_groups := utils.ToInt64(data["live_lv4_groups"])
		live_lv3_groups := utils.ToInt64(data["live_lv3_groups"])
		live_lv2_groups := utils.ToInt64(data["live_lv2_groups"])
		live_lv1_groups := utils.ToInt64(data["live_lv1_groups"])
		live_lv4_group_users := utils.ToInt64(data["live_lv4_group_users"])
		live_lv3_group_users := utils.ToInt64(data["live_lv3_group_users"])
		live_lv2_group_users := utils.ToInt64(data["live_lv2_group_users"])
		live_lv1_group_users := utils.ToInt64(data["live_lv1_group_users"])
		sday := fmt.Sprint(sdate)
		fdate := sday[0:4] + "-" + sday[4:6] + "-" + sday[6:8]

		stat := &entity.ShareAgentGroupVary{
			SDate: fdate,
		}
		stats = append(stats, stat)
		stat.Lv1Groups = lv1_groups
		stat.Lv1DeadGroups = lv1_groups - live_lv1_groups
		stat.Lv1LiveGroups = live_lv1_groups
		stat.Lv1GroupUsers = lv1_group_users
		stat.Lv1GroupAvgUsers = fmt.Sprintf("%.2f", ComputeFloat(lv1_group_users, lv1_groups))
		stat.Lv1DeadGroupUsers = lv1_group_users - live_lv1_group_users
		stat.Lv1LiveGroupUsers = live_lv1_group_users
		stat.Lv1GroupDeadRate = fmt.Sprintf("%.2f%%", ComputeFloat(lv1_groups-live_lv1_groups, lv1_groups)*100)
		stat.Lv1GroupLossRate = fmt.Sprintf("%.2f%%", ComputeFloat(lv1_group_users-live_lv1_group_users, lv1_group_users)*100)

		stat.Lv2Groups = lv2_groups
		stat.Lv2DeadGroups = lv2_groups - live_lv2_groups
		stat.Lv2LiveGroups = live_lv2_groups
		stat.Lv2GroupUsers = lv2_group_users
		stat.Lv2GroupAvgUsers = fmt.Sprintf("%.2f", ComputeFloat(lv2_group_users, lv2_groups))
		stat.Lv2DeadGroupUsers = lv2_group_users - live_lv2_group_users
		stat.Lv2LiveGroupUsers = live_lv2_group_users
		stat.Lv2GroupDeadRate = fmt.Sprintf("%.2f%%", ComputeFloat(lv2_groups-live_lv2_groups, lv2_groups)*100)
		stat.Lv2GroupLossRate = fmt.Sprintf("%.2f%%", ComputeFloat(lv2_group_users-live_lv2_group_users, lv2_group_users)*100)

		stat.Lv3Groups = lv3_groups
		stat.Lv3DeadGroups = lv3_groups - live_lv3_groups
		stat.Lv3LiveGroups = live_lv3_groups
		stat.Lv3GroupUsers = lv3_group_users
		stat.Lv3GroupAvgUsers = fmt.Sprintf("%.2f", ComputeFloat(lv3_group_users, lv3_groups))
		stat.Lv3DeadGroupUsers = lv3_group_users - live_lv3_group_users
		stat.Lv3LiveGroupUsers = live_lv3_group_users
		stat.Lv3GroupDeadRate = fmt.Sprintf("%.2f%%", ComputeFloat(lv3_groups-live_lv3_groups, lv3_groups)*100)
		stat.Lv3GroupLossRate = fmt.Sprintf("%.2f%%", ComputeFloat(lv3_group_users-live_lv3_group_users, lv3_group_users)*100)

		stat.Lv4Groups = lv4_groups
		stat.Lv4DeadGroups = lv4_groups - live_lv4_groups
		stat.Lv4LiveGroups = live_lv4_groups
		stat.Lv4GroupUsers = lv4_group_users
		stat.Lv4GroupAvgUsers = fmt.Sprintf("%.2f", ComputeFloat(lv4_group_users, lv4_groups))
		stat.Lv4DeadGroupUsers = lv4_group_users - live_lv4_group_users
		stat.Lv4LiveGroupUsers = live_lv4_group_users
		stat.Lv4GroupDeadRate = fmt.Sprintf("%.2f%%", ComputeFloat(lv4_groups-live_lv4_groups, lv4_groups)*100)
		stat.Lv4GroupLossRate = fmt.Sprintf("%.2f%%", ComputeFloat(lv4_group_users-live_lv4_group_users, lv4_group_users)*100)
	}
	return
}

// 满减券-满减券汇总
func (s *activityService) UserCouponDates(stime, etime time.Time, registAreas, utypes []int32, packageIds []string) (stats []*entity.UserCouponDate, err error) {
	dateStats := make(map[uint32]*entity.UserCouponDate)
	for begin := stime; begin.Before(etime); begin = begin.AddDate(0, 0, 1) {
		day_time := uint32(begin.Year()*10000 + int(begin.Month())*100 + begin.Day())
		sday := fmt.Sprint(day_time)
		sdate := sday[0:4] + "-" + sday[4:6] + "-" + sday[6:8]
		stat := &entity.UserCouponDate{SDate: sdate}
		stats = append(stats, stat)
		dateStats[day_time] = stat
	}
	utils.SliceReverse(stats)

	where_u_sql, where_u_args := s0ColUserFilter(registAreas, utypes, packageIds)

	var datas []map[string]any
	args := []any{stime.Unix(), etime.Unix()}
	args = append(args, where_u_args...)
	err = ck.Select(&datas, fmt.Sprintf(`
		select toYYYYMMDD(toDateTime(s1.ctime)) sdate, count(*) give_coupons, SUM(s1.min_recharge) min_recharge_sum, SUM(s1.amount) derate_amounts,
            SUM(case when s1.status = 1 and sdate = toYYYYMMDD(toDateTime(s1.utime)) then 1 else 0 end) current_use_coupons,
            SUM(case when startsWith(lower(s1.coupon_id), 'a') then 1 else 0 end) auto_give_coupons,
            SUM(case when startsWith(lower(s1.coupon_id), 'a') then s1.min_recharge else 0 end) auto_min_recharge_sum, 
            SUM(case when startsWith(lower(s1.coupon_id), 'a') then s1.amount else 0 end) auto_derate_amounts,
            SUM(case when startsWith(lower(s1.coupon_id), 'a') and s1.status = 1 and sdate = toYYYYMMDD(toDateTime(s1.utime)) then 1 else 0 end) auto_current_use_coupons,
            (give_coupons - auto_give_coupons) hand_give_coupons,
            (min_recharge_sum - auto_min_recharge_sum) hand_min_recharge_sum,
            (derate_amounts - auto_derate_amounts) hand_derate_amounts,
            (current_use_coupons - auto_current_use_coupons) hand_current_use_coupons
        from game.col_user s0 final
        join game.col_user_coupon s1 final on s0.userid = s1.user_id
        where s1.ctime between ? and ? %s
        group by sdate
	`, where_u_sql), args...)
	if err != nil {
		return
	}
	for _, data := range datas {
		sdate := data["sdate"].(uint32)
		give_coupons := utils.ToInt64(data["give_coupons"])
		min_recharge_sum := utils.ToInt64(data["min_recharge_sum"])
		derate_amounts := utils.ToInt64(data["derate_amounts"])
		current_use_coupons := utils.ToInt64(data["current_use_coupons"])
		auto_give_coupons := utils.ToInt64(data["auto_give_coupons"])
		auto_min_recharge_sum := utils.ToInt64(data["auto_min_recharge_sum"])
		auto_derate_amounts := utils.ToInt64(data["auto_derate_amounts"])
		auto_current_use_coupons := utils.ToInt64(data["auto_current_use_coupons"])
		hand_give_coupons := utils.ToInt64(data["hand_give_coupons"])
		hand_min_recharge_sum := utils.ToInt64(data["hand_min_recharge_sum"])
		hand_derate_amounts := utils.ToInt64(data["hand_derate_amounts"])
		hand_current_use_coupons := utils.ToInt64(data["hand_current_use_coupons"])
		if stat, ok := dateStats[sdate]; ok {
			stat.GiveCoupons = give_coupons
			stat.MinRecharge0 = min_recharge_sum
			stat.DerateAmounts0 = derate_amounts
			stat.AutoGiveCoupons = auto_give_coupons
			stat.AutoMinRecharge0 = auto_min_recharge_sum
			stat.AutoDerateAmounts0 = auto_derate_amounts
			stat.HandGiveCoupons = hand_give_coupons
			stat.HandMinRecharge0 = hand_min_recharge_sum
			stat.HandDerateAmounts0 = hand_derate_amounts

			stat.CurrentUseCoupons = current_use_coupons
			stat.AutoCurrentUseCoupons = auto_current_use_coupons
			stat.HandCurrentUseCoupons = hand_current_use_coupons
		}
	}

	var pay_datas []map[string]any
	pay_args := []any{stime, etime}
	pay_args = append(pay_args, where_u_args...)
	err = ck.Select(&pay_datas, fmt.Sprintf(`
		select toYYYYMMDD(s1.ctime) sdate, COUNT(*) use_coupons, SUM(s1.amount) pay_amounts, SUM(s1.score - s1.amount) derate_amounts,
                SUM(case when startsWith(lower(s1.shop_id), 'a') then 1 else 0 end) auto_use_coupons,
                SUM(case when startsWith(lower(s1.shop_id), 'a') then s1.amount else 0 end) auto_pay_amounts,
                SUM(case when startsWith(lower(s1.shop_id), 'a') then s1.score - s1.amount else 0 end) auto_derate_amounts,
                (use_coupons - auto_use_coupons) hand_use_coupons,
                (pay_amounts - auto_pay_amounts) hand_pay_amounts,
                (derate_amounts - auto_derate_amounts) hand_derate_amounts
        from game.col_user s0 final
        join game.col_trade_record s1 final on s1.userid = s0.userid
        where s1.ctime between ? and ? and s1.order_status = 4 and s1.shop_type = 11 %s
        group by sdate
	`, where_u_sql), pay_args...)
	if err != nil {
		return
	}
	for _, data := range pay_datas {
		sdate := data["sdate"].(uint32)
		use_coupons := utils.ToInt64(data["use_coupons"])
		pay_amounts := utils.ToInt64(data["pay_amounts"])
		derate_amounts := utils.ToInt64(data["derate_amounts"])
		auto_use_coupons := utils.ToInt64(data["auto_use_coupons"])
		auto_pay_amounts := utils.ToInt64(data["auto_pay_amounts"])
		auto_derate_amounts := utils.ToInt64(data["auto_derate_amounts"])
		hand_use_coupons := utils.ToInt64(data["hand_use_coupons"])
		hand_pay_amounts := utils.ToInt64(data["hand_pay_amounts"])
		hand_derate_amounts := utils.ToInt64(data["hand_derate_amounts"])
		if stat, ok := dateStats[sdate]; ok {
			stat.UseCouponOrders = use_coupons
			stat.UseCouponOrderAmounts0 = pay_amounts
			stat.UseCouponDerateAmounts0 = derate_amounts
			stat.AutoUseCouponOrders = auto_use_coupons
			stat.AutoUseCouponOrderAmounts0 = auto_pay_amounts
			stat.AutoUseCouponDerateAmounts0 = auto_derate_amounts
			stat.HandUseCouponOrders = hand_use_coupons
			stat.HandUseCouponOrderAmounts0 = hand_pay_amounts
			stat.HandUseCouponDerateAmounts0 = hand_derate_amounts
		}
	}

	summary := &entity.UserCouponDate{
		SDate: fmt.Sprintf("%s~%s汇总", stime.Format(utils.FORMAT_DATE), etime.Format(utils.FORMAT_DATE)),
	}
	stats = append([]*entity.UserCouponDate{summary}, stats...)
	for _, stat := range stats {
		summary.GiveCoupons += stat.GiveCoupons
		summary.MinRecharge0 += stat.MinRecharge0
		summary.DerateAmounts0 += stat.DerateAmounts0
		summary.UseCouponOrders += stat.UseCouponOrders
		summary.UseCouponOrderAmounts0 += stat.UseCouponOrderAmounts0
		summary.UseCouponDerateAmounts0 += stat.UseCouponDerateAmounts0

		summary.AutoGiveCoupons += stat.AutoGiveCoupons
		summary.AutoMinRecharge0 += stat.AutoMinRecharge0
		summary.AutoDerateAmounts0 += stat.AutoDerateAmounts0
		summary.AutoUseCouponOrders += stat.AutoUseCouponOrders
		summary.AutoUseCouponOrderAmounts0 += stat.AutoUseCouponOrderAmounts0
		summary.AutoUseCouponDerateAmounts0 += stat.AutoUseCouponDerateAmounts0

		summary.HandGiveCoupons += stat.HandGiveCoupons
		summary.HandMinRecharge0 += stat.HandMinRecharge0
		summary.HandDerateAmounts0 += stat.HandDerateAmounts0
		summary.HandUseCouponOrders += stat.HandUseCouponOrders
		summary.HandUseCouponOrderAmounts0 += stat.HandUseCouponOrderAmounts0
		summary.HandUseCouponDerateAmounts0 += stat.HandUseCouponDerateAmounts0

		summary.CurrentUseCoupons += stat.CurrentUseCoupons
	}

	for _, stat := range stats {
		stat.UseRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.UseCouponOrders, stat.GiveCoupons)*100)
		stat.CurrentUseRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.CurrentUseCoupons, stat.GiveCoupons)*100)

		stat.AutoUseRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.AutoUseCouponOrders, stat.AutoGiveCoupons)*100)
		stat.AutoCurrentUseRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.AutoCurrentUseCoupons, stat.AutoGiveCoupons)*100)

		stat.HandUseRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.HandUseCouponOrders, stat.HandGiveCoupons)*100)
		stat.HandCurrentUseRate = fmt.Sprintf("%.2f%%", ComputeFloat(stat.HandCurrentUseCoupons, stat.HandGiveCoupons)*100)

		stat.MinRecharge = fmt.Sprintf("%.2f", Chip2Float(stat.MinRecharge0))
		stat.DerateAmounts = fmt.Sprintf("%.2f", Chip2Float(stat.DerateAmounts0))
		stat.UseCouponOrderAmounts = fmt.Sprintf("%.2f", Chip2Float(stat.UseCouponOrderAmounts0))
		stat.UseCouponDerateAmounts = fmt.Sprintf("%.2f", Chip2Float(stat.UseCouponDerateAmounts0))
		stat.AutoMinRecharge = fmt.Sprintf("%.2f", Chip2Float(stat.AutoMinRecharge0))
		stat.AutoDerateAmounts = fmt.Sprintf("%.2f", Chip2Float(stat.AutoDerateAmounts0))
		stat.AutoUseCouponOrderAmounts = fmt.Sprintf("%.2f", Chip2Float(stat.AutoUseCouponOrderAmounts0))
		stat.AutoUseCouponDerateAmounts = fmt.Sprintf("%.2f", Chip2Float(stat.AutoUseCouponDerateAmounts0))
		stat.HandMinRecharge = fmt.Sprintf("%.2f", Chip2Float(stat.HandMinRecharge0))
		stat.HandDerateAmounts = fmt.Sprintf("%.2f", Chip2Float(stat.HandDerateAmounts0))
		stat.HandUseCouponOrderAmounts = fmt.Sprintf("%.2f", Chip2Float(stat.HandUseCouponOrderAmounts0))
		stat.HandUseCouponDerateAmounts = fmt.Sprintf("%.2f", Chip2Float(stat.HandUseCouponDerateAmounts0))
	}
	return
}

func (s *activityService) UserCouponHand(arg *args.CoupnHandArgs) (coupons []*entity.ManualCoupon, err error) {
	q := FindByDate5(arg.CtimeS, arg.CtimeE, "ctime", "ctime")
	st := FindByDate5(arg.StimeS, arg.StimeE, "send_time", "send_time")
	ut := FindByDate5(arg.UtimeS, arg.UtimeE, "used_time", "used_time")
	for k, v := range st {
		q[k] = v
	}
	for k, v := range ut {
		q[k] = v
	}

	if len(arg.RegistArea) > 0 {
		q["acct_type"] = bson.M{
			"$in": arg.RegistArea,
		}
	}

	if len(arg.UserTypes) > 0 {
		q["rechar_classify"] = bson.M{
			"$in": arg.UserTypes,
		}
	}
	err = ManualCoupons.Find(q).All(&coupons)
	if err != nil {
		return nil, err
	}

	classifyMap := map[int]string{
		0: "新手",
		1: "平民",
		2: "普R",
		3: "小R",
		4: "中R",
		5: "大R",
		6: "超大R",
	}
	RegistMap := map[int]string{
		0: "A类",
		1: "B类",
		2: "C类",
	}
	TagMap := map[int]string{
		1: "自定义",
	}
	var pay_datas []map[string]any
	err1 := ck.Select(&pay_datas, `
	select coupon_id, count(DISTINCT user_id) as count
	from game.col_user_coupon final
	where status = 1
	group by coupon_id
	`)
	if err1 != nil {
		return nil, err1
	}

	for _, coupon := range coupons {
		// var useUsers int64
		for _, p := range pay_datas {
			if p["coupon_id"] == coupon.Id {
				coupon.UsedUserCount = int64(p["count"].(uint64))
			}
		}
		coupon.MinRecharge = coupon.MinRecharge / 100
		coupon.Discount = coupon.Discount / 100
		coupon.UserIdStr = strings.Join(coupon.UserIds, ",")
		coupon.CtimeStr = utils.Stamp2Time(coupon.Ctime).In(location).Format(utils.FORMAT)
		coupon.SendTimeStr = utils.Stamp2Time(coupon.SendTime).In(location).Format(utils.FORMAT)
		if coupon.UsedTime > 0 {
			coupon.UsedTimeStr = utils.Stamp2Time(coupon.UsedTime).In(location).Format(utils.FORMAT)
		}
		coupon.SeeCount = coupon.SeeUserCount * int64(coupon.Num)
		coupon.PayAvg = fmt.Sprintf("[%d,%d)", coupon.OrderAvgRange[0], coupon.OrderAvgRange[1])
		coupon.ProfitabilityStr = fmt.Sprintf("[%d,%d)", coupon.ProfitabilityRatio[0], coupon.ProfitabilityRatio[1])
		if len(coupon.UserTag) > 0 {
			coupon.Tag = coupon.UserTag[0]
			coupon.UserTagStr = TagMap[coupon.UserTag[0]]
		}
		for _, classify := range coupon.RecharClassify {
			coupon.RecharClassifyStr += classifyMap[classify] + ","
		}
		for _, regist := range coupon.AcctType {
			coupon.AcctTypeStr += RegistMap[regist] + ","
		}

		if coupon.SendCount > 0 {
			coupon.UsedRatio = fmt.Sprintf("%.2f%%", float64(coupon.UsedCount*10000/coupon.SendCount)/100)
		}
		if coupon.SeeCount > 0 {
			coupon.SeeRatio = fmt.Sprintf("%.2f%%", float64(coupon.UsedCount*10000/coupon.SeeCount)/100)
		}
		if coupon.SendUserCount > 0 {
			coupon.UsedUserRatio = fmt.Sprintf("%.2f%%", float64(coupon.UsedUserCount*10000/coupon.SendUserCount)/100)
		}
		if coupon.SeeUserCount > 0 {
			coupon.SeeUserRatio = fmt.Sprintf("%.2f%%", float64(coupon.UsedUserCount*10000/coupon.SeeUserCount)/100)
		}
		if coupon.MinRecharge > 0 {
			coupon.GiveRatio = fmt.Sprintf("%.2f%%", float64(coupon.Discount*10000/coupon.MinRecharge)/100)
		}
	}

	return
}

func (s *activityService) EditSwitch(couponId string, open bool) (err error) {
	err = ManualCoupons.Update(bson.M{"_id": couponId}, bson.M{"$set": bson.M{"send_switch": open}})
	return
}

func (s *activityService) SaveManualCoupon(arg *entity.ManualCoupon) (err error) {
	return ManualCoupons.Insert(arg)
}

func (s *activityService) EditCoupon(coupon *entity.ManualCoupon) (err error) {
	sendTime, _ := time.ParseInLocation(utils.FORMAT, coupon.SendTimeStr, location)
	coupon.UserTag[0] = coupon.Tag
	var ids []string
	if coupon.UserIdStr != "" {
		ids = strings.Split(coupon.UserIdStr, ",")
	}
	err = ManualCoupons.Update(bson.M{"_id": coupon.Id}, bson.M{
		"$set": bson.M{
			"min_recharge":        coupon.MinRecharge * 100,
			"discount":            coupon.Discount * 100,
			"validity_time":       coupon.ValidityTime,
			"num":                 coupon.Num,
			"order_avg_range":     []int64{int64(coupon.OrderAvgRange[0]), int64(coupon.OrderAvgRange[1])},
			"profitability_ratio": []int64{coupon.ProfitabilityRatio[0], coupon.ProfitabilityRatio[1]},
			"rechar_classify":     coupon.RecharClassify,
			"acct_type":           coupon.AcctType,
			"user_tag":            coupon.UserTag,
			"user_ids":            ids,
			"send_time":           sendTime.Unix(),
		},
	})
	return
}

// 生成ID
func (s *activityService) CouponGenerateID() (string, error) {
	gen := new(entity.ShopIDGen)
	gen.Id = "last_manual_coupon"
	Get(GenIDs, gen.Id, gen)
	if gen.LastUserId == "" {
		gen.LastUserId = "m000001"
	} else {
		// 提取数字部分
		numStr := gen.LastUserId[1:]
		// 将数字部分加1
		newNum := utils.StringAdd(numStr)
		// 确保数字部分保持6位，不足补0
		newNum = fmt.Sprintf("%06s", newNum)
		// 组合新的ID
		gen.LastUserId = "m" + newNum
	}

	if Upsert(GenIDs, bson.M{"_id": gen.Id}, gen) {
		return gen.LastUserId, nil
	}
	return gen.LastUserId, errors.New("生成错误")
}
