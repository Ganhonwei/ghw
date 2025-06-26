package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"goserver/gen/pb"
	"goserver/internal/web/admin/app/entity"
	"goserver/pkg/data"
	"goserver/pkg/data/ck"
	"goserver/pkg/utils"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/globalsign/mgo/bson"
)

type payService struct{}

// 获取充值订单列表
// Deprecated
func (this *payService) PayList_Deprecated(page, pageSize int, m bson.M, isfirst int) ([]*entity.PayOrder, error) {
	var list []*entity.PayOrder
	if pageSize == -1 {
		pageSize = 100000
	}
	var err error
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	f_id := make([]string, 0)
	// 判断是否是注册后当天充值
	n := bson.M{}
	// 手动复制 m 到 n
	for k, v := range m {
		n[k] = v
	}
	n["order_status"] = 4
	pipeline := []bson.M{
		{
			"$match": n,
		},
		{
			"$sort": bson.M{
				"ctime": 1, // 根据订单时间升序排序
			},
		},
		{
			"$group": bson.M{
				"_id":      "$userid",
				"orderID":  bson.M{"$first": "$_id"},
				"pay_time": bson.M{"$first": "$pay_time"},
			},
		},
	}
	result := []bson.M{}
	pipe := Pays.Pipe(pipeline)
	err = pipe.All(&result)

	for _, item := range result {
		uid := item["_id"].(string)
		oid := item["orderID"].(string)
		p_time := item["pay_time"].(time.Time)
		us := bson.M{}
		us["_id"] = uid
		us["robot"] = false
		us["simulation_robot"] = false
		user_info := new(entity.PlayerUser)
		PlayerUsers.Find(us).One(&user_info)

		// 判断是否是同一天（印度时区）
		if isSameDayIndia(user_info.Ctime, p_time) {
			f_id = append(f_id, oid)
		}
	}
	if isfirst == 1 {
		// 只查询首冲
		m["_id"] = bson.M{"$in": f_id}
	} else if isfirst == 2 {
		// 排除首冲
		m["_id"] = bson.M{"$nin": f_id}
	}
	Pays.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)

	// Pays.
	// 	Find(m).
	// 	Sort(sortFieldR).
	// 	Skip(skipNum).
	// 	Limit(pageSize).
	// 	All(&list)
	list = this.chipList1(list, f_id)

	payListExtra(list)

	chipUsdtPay(list)

	return list, err
}

func (this *payService) PayListCK(page, pageSize int, params map[string]any, isfirst int) (list []*entity.PayOrder, total int, err error) {
	if pageSize == -1 {
		pageSize = 100000
	}
	first_id := make([]string, 0)
	// first_map := make(map[string]any)
	sql0 := `SELECT
    t.userid
FROM (SELECT
        userid,
        pay_time
    FROM game.col_trade_record
    FINAL
    WHERE order_status = 4 %s
    AND (userid, pay_time) IN (
        SELECT
            userid,
            MIN(pay_time)
        FROM game.col_trade_record
        FINAL
        WHERE order_status = 4 %s
        GROUP BY userid
    )) t
JOIN (SELECT
        userid,
        ctime
    FROM game.col_user cu FINAL
    WHERE robot = 0 AND simulation_robot = 0) u
ON t.userid = u.userid
WHERE
    toDate(toDateTime(t.pay_time, ?)) = toDate(toDateTime(u.ctime, ?))`

	sql1 := `select %s from game.col_trade_record ctr final where 1=1 %s %s`
	// 拼接条件
	where0, where1 := "", ""
	var args0, args1 []any
	if orderid, ok := params["_id"]; ok {
		where0 += " and order_id = ?"
		where1 += " and order_id = ?"
		args0 = append(args0, orderid)
		args1 = append(args1, orderid)
	}
	if userid, ok := params["userid"]; ok {
		where0 += " and userid = ?"
		args0 = append(args0, userid)
		where1 += " and userid = ?"
		args1 = append(args1, userid)
	}
	if starttime, ok := params["starttime"]; ok {
		where0 += " and ctime >= toDateTime(?, ?) "
		args0 = append(args0, starttime)
		args0 = append(args0, locationName)
		where1 += " and ctime >= toDateTime(?, ?) "
		args1 = append(args1, starttime)
		args1 = append(args1, locationName)
	}
	if endtime, ok := params["endtime"]; ok {
		where0 += " and ctime <= toDateTime(?, ?) "
		args0 = append(args0, endtime)
		args0 = append(args0, locationName)
		where1 += " and ctime <= toDateTime(?, ?) "
		args1 = append(args1, endtime)
		args1 = append(args1, locationName)
	}
	if out_trade_no, ok := params["out_trade_no"]; ok {
		where0 += " and out_trade_no = ?"
		args0 = append(args0, out_trade_no)
		where1 += " and out_trade_no = ?"
		args1 = append(args1, out_trade_no)
	}
	if package_id, ok := params["package_id"]; ok {
		where0 += " and nickname LIKE '%?%'"
		args0 = append(args0, package_id)
		where1 += " and nickname LIKE '%?%'"
		args1 = append(args1, package_id)
	}
	if order_status, ok := params["order_status"]; ok {
		where1 += " and order_status = ?"
		args1 = append(args1, order_status)
	}
	if channel_id, ok := params["channel_id"]; ok {
		where0 += " and channel_id = ?"
		args0 = append(args0, channel_id)
		where1 += " and channel_id = ?"
		args1 = append(args1, channel_id)
	}
	if shop_type, ok := params["shop_type"]; ok {
		where0 += " and shop_type = ?"
		args0 = append(args0, shop_type)
		where1 += " and shop_type = ?"
		args1 = append(args1, shop_type)
	}
	if select_package, ok := params["select_package"]; ok {
		where0 += " and package_id in ?"
		args0 = append(args0, select_package)
		where1 += " and package_id in ?"
		args1 = append(args1, select_package)
	}
	args0 = append(args0, locationName)
	args0 = append(args0, locationName)
	// 查首充订单
	sql_first := fmt.Sprintf(sql0, where0, where0)
	err = ck.Select(&first_id, sql_first, args0...)
	if err != nil {
		return
	}
	if isfirst == 1 {
		// 只查询首冲
		where1 += " and userid in ?"
		args1 = append(args1, first_id)
	} else if isfirst == 2 {
		// 排除首冲
		where1 += " and userid nin ?"
		args1 = append(args1, first_id)
	}
	// 查总数
	sql_count := fmt.Sprintf(sql1, "count(*) c", where1, "")
	err = ck.Select(&total, sql_count, args1...)
	if err != nil || total == 0 {
		return
	}

	selects := "*"
	order := " ORDER BY ctime DESC LIMIT ?, ?"
	sql_list := fmt.Sprintf(sql1, selects, where1, order)
	args_list := args1
	offset, limit := PageCalc(page, pageSize)
	args_list = append(args_list, offset, limit)
	err = ck.Select(&list, sql_list, args_list...)
	if err != nil || len(list) == 0 {
		return
	}

	//转换为分展示
	list = this.chipList1(list, first_id)
	// 查询游戏局数

	payListExtra(list)

	chipUsdtPay(list)

	return list, total, err
}

//	func (this *payService) TotalPayOrder(params map[string]any) (*entity.TotalPayOrder, error) {
//		info := new(entity.TotalPayOrder)
//		// pipeline := []bson.M{
//		// 	{
//		// 		"$match": m,
//		// 	},
//		// 	{
//		// 		"$group": bson.M{
//		// 			"_id": nil,
//		// 			"successful_order": bson.M{"$sum": bson.M{"$cond": []interface{}{
//		// 				bson.M{"$eq": []interface{}{"$order_status", 4}}, 1, nil,
//		// 			}}},
//		// 			"successful_amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
//		// 				bson.M{"$eq": []interface{}{"$order_status", 4}}, "$amount", nil,
//		// 			}}},
//		// 			"total_order":  bson.M{"$sum": 1},
//		// 			"total_amount": bson.M{"$sum": "$amount"},
//		// 		},
//		// 	},
//		// }
//		// Pays.Pipe(pipeline).One(&info)
//		sql1 := `SELECT COUNT(1) total_order,SUM(amount) total_amount,SUM(case when order_status = 4 then 1 else 0 end) successful_order,SUM(case when order_status = 4 then amount else 0 end) successful_amount
//	    FROM game.col_trade_record ctr FINAL WHERE 1=1 %s`
//		var args1 []any
//		where1 := ""
//		if orderid, ok := params["_id"]; ok {
//			where1 += " and order_id = ?"
//			args1 = append(args1, orderid)
//		}
//		if userid, ok := params["userid"]; ok {
//			where1 += " and userid = ?"
//			args1 = append(args1, userid)
//		}
//		if starttime, ok := params["starttime"]; ok {
//			where1 += " and ctime >= toDateTime(?, ?) "
//			args1 = append(args1, starttime)
//			args1 = append(args1, timeSetting)
//		}
//		if endtime, ok := params["endtime"]; ok {
//			where1 += " and ctime <= toDateTime(?, ?) "
//			args1 = append(args1, endtime)
//			args1 = append(args1, timeSetting)
//		}
//		if out_trade_no, ok := params["out_trade_no"]; ok {
//			where1 += " and out_trade_no = ?"
//			args1 = append(args1, out_trade_no)
//		}
//		if package_id, ok := params["package_id"]; ok {
//			where1 += " and nickname LIKE '%?%'"
//			args1 = append(args1, package_id)
//		}
//		if order_status, ok := params["order_status"]; ok {
//			where1 += " and order_status = ?"
//			args1 = append(args1, order_status)
//		}
//		if channel_id, ok := params["channel_id"]; ok {
//			where1 += " and channel_id = ?"
//			args1 = append(args1, channel_id)
//		}
//		if shop_type, ok := params["shop_type"]; ok {
//			where1 += " and shop_type = ?"
//			args1 = append(args1, shop_type)
//		}
//		if select_package, ok := params["select_package"]; ok {
//			where1 += " and package_id in ?"
//			args1 = append(args1, select_package)
//		}
//		sql_select := fmt.Sprintf(sql1, where1)
//		ck.Select(&info, sql_select, args1...)
//		if info.TotalAmount != 0 {
//			info.FTotalAmount = ComputeFloat(info.TotalAmount, 100)
//		}
//		if info.SuccessfulAmount != 0 {
//			info.FSuccessfulAmount = ComputeFloat(info.SuccessfulAmount, 100)
//		}
//		return info, nil
//	}
func (this *payService) TotalPayOrder(m bson.M) (*entity.TotalPayOrder, error) {
	info := new(entity.TotalPayOrder)
	pipeline := []bson.M{
		{
			"$match": m,
		},
		{
			"$group": bson.M{
				"_id": nil,
				"successful_order": bson.M{"$sum": bson.M{"$cond": []interface{}{
					bson.M{"$eq": []interface{}{"$order_status", 4}}, 1, nil,
				}}},
				"successful_amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
					bson.M{"$eq": []interface{}{"$order_status", 4}}, "$amount", nil,
				}}},
				"total_order":  bson.M{"$sum": 1},
				"total_amount": bson.M{"$sum": "$amount"},
			},
		},
	}
	Pays.Pipe(pipeline).One(&info)
	if info.TotalAmount != 0 {
		info.FTotalAmount = ComputeFloat(info.TotalAmount, 100)
	}
	if info.SuccessfulAmount != 0 {
		info.FSuccessfulAmount = ComputeFloat(info.SuccessfulAmount, 100)
	}
	return info, nil
}

// payListExtra 在充值页面和用户列表增加成功次数和均单价。
func payListExtra(payList []*entity.PayOrder) {
	userids := utils.SliceMapping(payList, func(p *entity.PayOrder) string {
		return p.Userid
	})
	m1 := []bson.M{
		{
			"$match": bson.M{
				"userid":       bson.M{"$in": userids},
				"order_status": 4,
			},
		},
		{
			"$group": bson.M{
				"_id":   "$userid",
				"count": bson.M{"$sum": 1},
				"total": bson.M{"$sum": "$score"},
			},
		},
	}
	var list []bson.M
	err := Pays.Pipe(m1).All(&list)
	if err != nil {
		beego.Error("payListExtra error:", err)
		return
	}

	uidMap := utils.Slice2Map(list, func(_ int, item bson.M) string {
		return item["_id"].(string)
	})

	for _, pay := range payList {
		item, ok := uidMap[pay.Userid]
		if !ok {
			continue
		}
		pay.SuccessTimes = item["count"].(int)
		total := item["total"].(int)
		if pay.SuccessTimes > 0 {
			pay.AvgPay = fmt.Sprintf("%.2f", float64(total)/float64(pay.SuccessTimes)/100.0)
		}
	}
}

// // payListExtra 在充值页面和用户列表增加成功次数和均单价。
// func payListExtra(payList []*entity.PayOrder) {
// 	userids := utils.SliceMapping(payList, func(p *entity.PayOrder) string {
// 		return p.Userid
// 	})
// 	var list []map[string]any
// 	sql0 := `SELECT userid,COUNT(1) totalcount,SUM(score) totalscore
// 	FROM game.col_trade_record ctr FINAL WHERE order_status = 4 AND userid in ?
// 	GROUP BY userid`
// 	var args0 []any
// 	args0 = append(args0, userids)
// 	err := ck.Select(&list, sql0, args0...)
// 	// m1 := []bson.M{
// 	// 	{
// 	// 		"$match": bson.M{
// 	// 			"userid":       bson.M{"$in": userids},
// 	// 			"order_status": 4,
// 	// 		},
// 	// 	},
// 	// 	{
// 	// 		"$group": bson.M{
// 	// 			"_id":   "$userid",
// 	// 			"count": bson.M{"$sum": 1},
// 	// 			"total": bson.M{"$sum": "$score"},
// 	// 		},
// 	// 	},
// 	// }
// 	// var list []bson.M
// 	// err := Pays.Pipe(m1).All(&list)
// 	if err != nil {
// 		beego.Error("payListExtra error:", err)
// 		return
// 	}

// 	// uidMap := utils.Slice2Map(list, func(_ int, item bson.M) string {
// 	// 	return item["_id"].(string)
// 	// })
// 	uidMap := utils.Slice2Map(list, func(_ int, item map[string]any) string {
// 		return item["userid"].(string)
// 	})

// 	for _, pay := range payList {
// 		item, ok := uidMap[pay.Userid]
// 		if !ok {
// 			continue
// 		}
// 		// pay.SuccessTimes = item["totalcount"].(int)
// 		// total := item["totalscore"].(int)
// 		pay.SuccessTimes = int(utils.ToInt64(item["totalcount"]))
// 		total := utils.ToInt64(item["totalscore"])
// 		if pay.SuccessTimes > 0 {
// 			pay.AvgPay = fmt.Sprintf("%.2f", float64(total)/float64(pay.SuccessTimes)/100.0)
// 		}
// 	}
// }

func chipUsdtPay(payList []*entity.PayOrder) {
	var usdtOrderids []string
	var usdtOrderMap = make(map[string]*entity.PayOrder)
	for _, pay := range payList {
		if pay.ChannelId == 5002 {
			if pay.OutTradeNo != "" {
				usdtOrderids = append(usdtOrderids, pay.OutTradeNo)
				usdtOrderMap[pay.OutTradeNo] = pay
			}
		}
	}
	if len(usdtOrderids) == 0 {
		return
	}

	// usdt 回调信息
	var result []bson.M
	err := DsfPayOrder.Pipe([]bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": usdtOrderids}}},
		{"$project": bson.M{"_id": "$_id", "request_msg": "$request_msg", "call_back_msg": "$call_back_msg"}},
	}).All(&result)
	if err != nil {
		beego.Error("chipUsdtPay error: ", err)
		return
	}
	for _, r := range result {
		id := r["_id"].(string)
		request_msg, ok1 := r["request_msg"].(string)
		call_back_msg, ok2 := r["call_back_msg"].(string)
		order, ok3 := usdtOrderMap[id]
		if !ok3 {
			continue
		}
		if !ok1 {
			continue
		}
		if request_msg != "" {
			res := &entity.UsdtPayOrderRespond{}
			if err = json.Unmarshal([]byte(request_msg), &res); err == nil {
				order.USDTMsg += fmt.Sprintf("usdt:%sU", res.Data.ActualAmount)

				if res.Msg != "" && res.Msg != "success" {
					order.USDTMsg += fmt.Sprintf(",%s", res.Msg)
				}
			} else {
				order.USDTMsg = request_msg
			}
		}
		if !ok2 || call_back_msg == "" {
			continue
		}
		cb := &entity.UsdtPayRechargeCallback{}
		if err = json.Unmarshal([]byte(call_back_msg), &cb); err == nil {
			order.USDTMsg += fmt.Sprintf(",交易号:%sU", cb.BlockTransactionId)
		}
	}
}

// 获取充值订单总数
func (this *payService) GetPayTotal(m bson.M, isfirst int) (int64, error) {
	var err error
	if isfirst == 1 || isfirst == 2 {
		n := bson.M{}
		// 手动复制 m 到 n
		for k, v := range m {
			n[k] = v
		}
		n["order_status"] = 4
		pipeline := []bson.M{
			{
				"$match": n,
			},
			{
				"$sort": bson.M{
					"ctime": 1, // 根据订单时间升序排序
				},
			},
			{
				"$group": bson.M{
					"_id":      "$userid",
					"orderID":  bson.M{"$first": "$_id"},
					"pay_time": bson.M{"$first": "$pay_time"},
				},
			},
		}
		result := []bson.M{}
		pipe := Pays.Pipe(pipeline)
		err = pipe.All(&result)
		f_id := make([]string, 0)
		for _, item := range result {
			uid := item["_id"].(string)
			oid := item["orderID"].(string)
			p_time := item["pay_time"].(time.Time)
			us := bson.M{}
			us["_id"] = uid
			us["robot"] = false
			us["simulation_robot"] = false
			user_info := new(entity.PlayerUser)
			PlayerUsers.Find(us).One(&user_info)

			// 判断是否是同一天（印度时区）
			if isSameDayIndia(user_info.Ctime, p_time) {
				f_id = append(f_id, oid)
			}
		}
		if isfirst == 1 {
			// 只查询首冲
			m["_id"] = bson.M{"$in": f_id}
		} else if isfirst == 2 {
			// 排除首冲
			m["_id"] = bson.M{"$nin": f_id}
		}
	}
	return int64(Count(Pays, m)), err
}

// 根据条件查询对应的支付订单
func (this *payService) GetByPayUser(m bson.M) ([]entity.PayOrder, error) {
	var list []entity.PayOrder
	err := Pays.
		Find(m).All(&list)

	return list, err
}

// 转换为分展示
func (this *payService) chipList1(list []*entity.PayOrder, f_arr []string) []*entity.PayOrder {
	for k, v := range list {
		v.FAmount = Chip2Float(int64(v.Amount))
		if v.RealAmount != 0 {
			v.FRealAmount = fmt.Sprintf("%.2f", Chip2Float(int64(v.RealAmount)))
		}
		c, _ := ConvertToIndiaTime(v.Ctime.Unix())
		v.Ctime = c
		c, _ = ConvertToIndiaTime(v.Utime.Unix())
		v.Utime = c
		c, _ = ConvertToIndiaTime(v.PayTime.Unix())
		v.PayTime = c
		for _, b := range f_arr {
			if b == v.OrderID {
				v.IsFirst = true
			}
		}

		list[k] = v
	}
	return list
}

// 获取提现订单列表 todo 安卓评分
func (this *payService) WithdrawList(page, pageSize int, m bson.M, paylist []entity.PayChannel) ([]*entity.WithdrawOrder, error) {
	var list []*entity.WithdrawOrder
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	Withdraws.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)

	list = this.chipList2(list, paylist)
	chipUsdtWithdraw(list)

	// 统计
	err := this.withdrawStats(list)
	if err != nil {
		return nil, err
	}

	// 汇总
	summary, err := this.withdrawSummary(m)
	if err != nil {
		return nil, err
	}
	list = append([]*entity.WithdrawOrder{summary}, list...)

	return list, nil
}

func (this *payService) withdrawSummary(m bson.M) (summary *entity.WithdrawOrder, err error) {
	summary = &entity.WithdrawOrder{Userid: "总汇"}

	var where0 string
	var args0 []any
	for k, v := range m {
		switch k {
		case "ctime":
			ctime := v.(bson.M)
			if gte, ok := ctime["$gte"]; ok {
				where0 += " AND ctime >= ?"
				args0 = append(args0, gte)
			}
			if lte, ok := ctime["$lte"]; ok {
				where0 += " AND ctime <= ?"
				args0 = append(args0, lte)
			}
		case "userid":
			where0 += " AND userid = ?"
			args0 = append(args0, v)
		case "_id":
			where0 += " AND order_id = ?"
			args0 = append(args0, v)
		case "out_trade_no":
			where0 += " AND out_trade_no = ?"
			args0 = append(args0, v)
		case "blank_number":
			where0 += " AND blank_number = ?"
			args0 = append(args0, v)
		case "order_status":
			where0 += " AND order_status = ?"
			args0 = append(args0, v)
		case "examine_way":
			where0 += " AND examine_way = ?"
			args0 = append(args0, v)
		case "package_id":
			where0 += " AND package_id IN ?"
			args0 = append(args0, v.(bson.M)["$in"])
		}
	}

	var data_Withdraws map[string]any
	err = ck.Select(&data_Withdraws, fmt.Sprintf(`
		SELECT SUM(amount) amounts, SUM(commission) commissions, (amounts + commissions) total  
			FROM game.col_withdraw_record FINAL WHERE 1 = 1 %s
	`, where0), args0...)
	if err != nil {
		return
	}
	summary.FTotal = Chip2Float(utils.ToInt64(data_Withdraws["total"]))
	summary.FCommission = Chip2Float(utils.ToInt64(data_Withdraws["commissions"]))
	summary.FAmount = Chip2Float(utils.ToInt64(data_Withdraws["amounts"]))

	usersQ := `SELECT userid FROM game.col_withdraw_record FINAL WHERE 1 = 1 %s`
	usersQ = fmt.Sprintf(usersQ, where0)

	// 玩最多的游戏/局数/局均码量/返奖率
	var data_GameStats []map[string]any
	var args_GameStats []any
	args_GameStats = append(args_GameStats, args0...)
	args_GameStats = append(args_GameStats, args0...)
	args_GameStats = append(args_GameStats, args0...)
	err = ck.Select(&data_GameStats, fmt.Sprintf(`
		SELECT * FROM (
			SELECT gtype, count(*) rounds, SUM(bet_amount) bet_sum,  AVG(bet_amount) bet_avg, SUM(rebate) rebate_sum,
				row_number() OVER (ORDER BY rounds desc) AS round_no,
				row_number() OVER (ORDER BY bet_sum desc) AS bet_no
			FROM (
					SELECT id, gtype, bet_amount, win_type, score, (CASE WHEN win_type = 1 THEN score ELSE 0 END) rebate
					FROM game.col_detail t1 FINAL WHERE robot = 0 and win_type in (1,2,3) AND userid IN (%s)
						UNION ALL
					SELECT s2.round_id id, s2.game_id gtype, SUM(s2.amount_sum) bet_amount,
						(CASE WHEN score > 0 THEN 1 WHEN score == 0 THEN 3 ELSE 2 END) win_type, 
						SUM(s3.amount_sum) - SUM(s2.amount_sum) score,
						SUM(s3.amount_sum) rebate
					FROM (SELECT round_id, user_id, game_id, SUM(amount) amount_sum FROM game.col_nsq_log_external_bet FINAL WHERE amount != 0 and user_id IN (%s) GROUP BY round_id, user_id, game_id) s2
					LEFT JOIN (SELECT round_id, user_id, SUM(amount) amount_sum FROM game.col_nsq_log_external_reward FINAL WHERE amount != 0 and user_id IN (%s) GROUP BY round_id, user_id) s3
						ON s2.round_id = s3.round_id AND s2.user_id = s3.user_id
					GROUP BY s2.round_id, s2.game_id
			) s1
			GROUP BY gtype
		) tt
		WHERE tt.round_no = 1 OR tt.bet_no = 1
		ORDER BY rounds DESC
	`, usersQ, usersQ, usersQ), args_GameStats...)
	if err != nil {
		return
	}
	for _, data := range data_GameStats {
		round_no := utils.ToInt64(data["round_no"])
		bet_no := utils.ToInt64(data["bet_no"])
		gtype := utils.ToInt64(data["gtype"])
		bet_sum := utils.ToInt64(data["bet_sum"])
		rebate_sum := utils.ToInt64(data["rebate_sum"])

		gameName, ok := GtypeNameMap[int(gtype)]
		if !ok {
			gameName = fmt.Sprint(gtype)
		}
		// 游戏/返奖率
		stats := fmt.Sprintf("%s/%.2f%%", gameName, utils.CaseElse(rebate_sum == 0, 0, float64(bet_sum)/float64(rebate_sum)*100))
		if round_no == 1 {
			summary.Game1TimesStats = stats
		}
		if bet_no == 1 {
			summary.Game1BetStats = stats
		}
	}

	var RebateRate float64
	err = ck.Select(&RebateRate, fmt.Sprintf(`
		SELECT (CASE amount WHEN 0 THEN 1 ELSE ((diamond + score) / amount) END) rebate_rate FROM (
			SELECT
			(SELECT SUM(diamond + give_diamond) diamond FROM game.col_user_finance FINAL WHERE userid IN (%s)) diamond,
			(SELECT SUM(score) score FROM game.col_withdraw_record FINAL WHERE userid IN (%s) AND order_status IN (1,2)) score,
			(SELECT SUM(amount) amount FROM game.col_trade_record FINAL WHERE userid IN (%s) AND order_status = 4) amount
		) t1
	`, usersQ, usersQ, usersQ), append(args0, append(args0, args0...)...)...)
	if err != nil {
		return
	}
	summary.RebateRate = fmt.Sprintf("%.2f%%", RebateRate*100)

	// 总充值金额, 玩家充值成功率
	// 玩家提现成功率
	var data_TradeStats map[string]any
	err = ck.Select(&data_TradeStats, fmt.Sprintf(`
		SELECT SUM(t2.pay_times_try) pay_times_try, SUM(t2.pay_times) pay_times, SUM(t2.pay_amount) pay_amount, 
			SUM(t3.withdraw_times_try) withdraw_times_try, SUM(t3.withdraw_times) withdraw_times,
			SUM(t3.withdraw_amount) withdraw_amount, SUM(t3.withdraw_amount_all) withdraw_amount_all,
			SUM(t3.withdraw_amount_wait) withdraw_amount_wait, SUM(t3.withdraw_amount_freeze) withdraw_amount_freeze,
			SUM(t3.withdraw_amount_processing) withdraw_amount_processing
		FROM (SELECT userid FROM game.col_user_finance FINAL WHERE userid IN (%s)) t1
		LEFT JOIN (
			SELECT userid, count(*) pay_times_try, 
				SUM(CASE WHEN order_status = 4 THEN 1 ELSE 0 END) pay_times,
				SUM(CASE WHEN order_status = 4 THEN amount ELSE 0 END) pay_amount
			FROM game.col_trade_record FINAL
			WHERE userid IN (%s)
			GROUP BY userid
		) t2 ON t1.userid = t2.userid
		LEFT JOIN (
			SELECT userid, count(*) withdraw_times_try, 
				SUM(CASE WHEN order_status = 2 THEN 1 ELSE 0 END) withdraw_times,
				SUM(CASE WHEN order_status = 2 THEN amount ELSE 0 END) withdraw_amount,
				SUM(amount) withdraw_amount_all,
				SUM(CASE WHEN order_status = 1 THEN amount ELSE 0 END) withdraw_amount_wait,
				SUM(CASE WHEN order_status = 4 THEN amount ELSE 0 END) withdraw_amount_freeze,
				SUM(CASE WHEN order_status in (6,8) THEN amount ELSE 0 END) withdraw_amount_processing
			FROM game.col_withdraw_record FINAL
			WHERE userid IN (%s)
			GROUP BY userid
		) t3 ON t1.userid = t3.userid
	`, usersQ, usersQ, usersQ), append(args0, append(args0, args0...)...)...)
	if err != nil {
		return
	}
	pay_times_try := utils.ToInt64(data_TradeStats["pay_times_try"])
	pay_times := utils.ToInt64(data_TradeStats["pay_times"])
	pay_amount := utils.ToInt64(data_TradeStats["pay_amount"])
	withdraw_times_try := utils.ToInt64(data_TradeStats["withdraw_times_try"])
	withdraw_times := utils.ToInt64(data_TradeStats["withdraw_times"])
	withdraw_amount := utils.ToInt64(data_TradeStats["withdraw_amount"])
	withdraw_amount_all := utils.ToInt64(data_TradeStats["withdraw_amount_all"])
	withdraw_amount_wait := utils.ToInt64(data_TradeStats["withdraw_amount_wait"])
	withdraw_amount_freeze := utils.ToInt64(data_TradeStats["withdraw_amount_freeze"])
	withdraw_amount_processing := utils.ToInt64(data_TradeStats["withdraw_amount_processing"])
	withdraw_amount_other := withdraw_amount_all - withdraw_amount - withdraw_amount_wait - withdraw_amount_freeze - withdraw_amount_processing
	summary.RechargeSuccessRate = fmt.Sprintf("%.2f%%", utils.CaseElse(pay_times_try == 0, 0, float64(pay_times)/float64(pay_times_try)*100))
	summary.WithdrawSuccessRate = fmt.Sprintf("%.2f%%", utils.CaseElse(withdraw_times_try == 0, 0, float64(withdraw_times)/float64(withdraw_times_try)*100))
	summary.RechargeAmount = fmt.Sprintf("%.2f", ComputeFloat(pay_amount, 100))
	summary.RechargeAmount0 = pay_amount
	summary.WithdrawAmountAll = fmt.Sprintf("%.2f", ComputeFloat(withdraw_amount_all, 100))
	summary.WithdrawAmount = fmt.Sprintf("%.2f", ComputeFloat(withdraw_amount, 100))
	summary.WithdrawAmount0 = withdraw_amount
	summary.WithdrawWaitAmount = fmt.Sprintf("%.2f", ComputeFloat(withdraw_amount_wait, 100))
	summary.WithdrawWaitAmount0 = withdraw_amount_wait
	summary.FreezeAmount = fmt.Sprintf("%.2f", ComputeFloat(withdraw_amount_freeze, 100))
	summary.ProcessingAmount = fmt.Sprintf("%.2f", ComputeFloat(withdraw_amount_processing, 100))
	summary.WithdrawOtherAmount = fmt.Sprintf("%.2f", ComputeFloat(withdraw_amount_other, 100))
	summary.WithdrawOtherAmount0 = withdraw_amount_other

	// 玩家当前携带
	var data_finance map[string]any
	err = ck.Select(&data_finance, fmt.Sprintf(`
		SELECT SUM(diamond) diamond_sum FROM game.col_user_finance FINAL WHERE userid in (%s)
	`, usersQ), args0...)
	if err != nil {
		return
	}
	diamond := utils.ToInt64(data_finance["diamond_sum"])
	// 计算利润
	summary.FDiamond = fmt.Sprintf("%.2f", ComputeFloat(diamond, 100))
	profit := diamond + summary.WithdrawAmount0 + summary.WithdrawWaitAmount0 - summary.RechargeAmount0
	summary.FProfit = fmt.Sprintf("%.2f", ComputeFloat(profit, 100))

	// 提现失败被自动退回金额
	sql_back := fmt.Sprintf(`
		SELECT SUM(amount) back_amount
		FROM game.col_withdraw_record FINAL
		WHERE order_id IN (
			SELECT order_id
			FROM game.col_withdraw_record t1 FINAL
			JOIN game.col_withdraw_record_transfer_order_detail t2 FINAL ON t1.order_id = t2.order_id AND t1.userid = t2.userid
			WHERE t1.order_status = 3 AND t2.status = ? AND userid IN (%s)
			GROUP BY order_id
		)
	`, usersQ)
	var data_back5 map[string]any
	err = ck.Select(&data_back5, sql_back, append([]any{5}, args0...)...)
	if err != nil {
		return
	}
	// 提单失败被自动退回金额
	var data_back7 map[string]any
	err = ck.Select(&data_back7, sql_back, append([]any{7}, args0...)...)
	if err != nil {
		return
	}

	back_amount5 := utils.ToInt64(data_back5["back_amount"])
	summary.BackAmount5 = fmt.Sprintf("%.2f", Chip2Float(back_amount5))
	// 其他提现金额减提单失败回退
	summary.WithdrawOtherAmount0 -= back_amount5
	summary.WithdrawOtherAmount = fmt.Sprintf("%.2f", ComputeFloat(summary.WithdrawOtherAmount0, 100))

	back_amount7 := utils.ToInt64(data_back7["back_amount"])
	summary.BackAmount7 = fmt.Sprintf("%.2f", Chip2Float(back_amount7))
	// 其他提现金额减提现失败回退
	summary.WithdrawOtherAmount0 -= back_amount7
	summary.WithdrawOtherAmount = fmt.Sprintf("%.2f", ComputeFloat(summary.WithdrawOtherAmount0, 100))
	return
}

func (this *payService) withdrawStats(list []*entity.WithdrawOrder) (err error) {
	if len(list) == 0 {
		return
	}
	channelMap, err := ChannelService.GetChannelsMap()
	if err != nil {
		return
	}

	var userids []string
	var useridDatas = make(map[string][]*entity.WithdrawOrder)
	for _, item := range list {
		userids = append(userids, item.Userid)
		useridDatas[item.Userid] = append(useridDatas[item.Userid], item)
		if c, ok := channelMap[item.PackageId]; ok {
			item.AliasId = c.Name1
		}
		switch item.OrderStatus {
		case 1:
			item.OrderStatusName = "审核中"
		case 2:
			item.OrderStatusName = "提现成功"
		case 3:
			item.OrderStatusName = "提现退回"
		case 4:
			item.OrderStatusName = "提现冻结"
		case 5:
			item.OrderStatusName = "提现失败"
		case 6:
			item.OrderStatusName = "提单成功"
		case 7:
			item.OrderStatusName = "提单失败"
		case 8:
			item.OrderStatusName = "核单中"
		case 9:
			item.OrderStatusName = "银行信息错误"
		case 10:
			item.OrderStatusName = "商户号余额不足"
		case 11:
			item.OrderStatusName = "支付通道错误"
		case 12:
			item.OrderStatusName = "玩家主动申请退回"
		default:
			item.OrderStatusName = "--"
		}
	}

	// 玩家游戏统计
	var datas_GameStats []map[string]any
	err = ck.Select(&datas_GameStats, `
		SELECT * FROM (
			SELECT userid, gtype, count(*) rounds, SUM(bet_amount) bet_sum, AVG(bet_amount) bet_avg, SUM(rebate) rebate_sum,
				row_number() OVER (PARTITION BY userid ORDER BY rounds desc) AS round_no,
				row_number() OVER (PARTITION BY userid ORDER BY bet_sum desc) AS bet_no
			FROM (
					SELECT id, userid, gtype, bet_amount, win_type, score, (CASE WHEN win_type = 1 THEN score ELSE 0 END) rebate
					FROM game.col_detail t1 FINAL WHERE robot = 0 and win_type in (1,2,3) AND userid IN ?
						UNION ALL
					SELECT s2.round_id id, s2.user_id userid, s2.game_id gtype, SUM(s2.amount_sum) bet_amount,
						(CASE WHEN score > 0 THEN 1 WHEN score == 0 THEN 3 ELSE 2 END) win_type,
						SUM(s3.amount_sum) - SUM(s2.amount_sum) score,
						SUM(s3.amount_sum) rebate
					FROM (SELECT round_id, user_id, game_id, SUM(amount) amount_sum FROM game.col_nsq_log_external_bet FINAL WHERE amount != 0  GROUP BY round_id, user_id, game_id) s2
					LEFT JOIN (SELECT round_id, user_id, SUM(amount) amount_sum FROM game.col_nsq_log_external_reward FINAL WHERE amount != 0 GROUP BY round_id, user_id) s3
						ON s2.round_id = s3.round_id AND s2.user_id = s3.user_id
					WHERE s2.user_id IN ?
					GROUP BY s2.round_id, s2.user_id, s2.game_id
			) s1
			GROUP BY userid, gtype
		) tt
		WHERE tt.round_no = 1 OR tt.bet_no = 1
		ORDER BY userid, rounds DESC
	`, userids, userids)
	if err != nil {
		return
	}
	for _, data := range datas_GameStats {
		userid := data["userid"].(string)
		if items, ok := useridDatas[userid]; ok {
			round_no := utils.ToInt64(data["round_no"])
			bet_no := utils.ToInt64(data["bet_no"])
			gtype := utils.ToInt64(data["gtype"])
			bet_sum := utils.ToInt64(data["bet_sum"])
			rebate_sum := utils.ToInt64(data["rebate_sum"])

			gameName, ok := GtypeNameMap[int(gtype)]
			if !ok {
				gameName = fmt.Sprint(gtype)
			}
			for _, item := range items {
				// 游戏/返奖率
				stats := fmt.Sprintf("%s/%.2f%%", gameName, utils.CaseElse(rebate_sum == 0, 0, float64(bet_sum)/float64(rebate_sum)*100))
				if round_no == 1 {
					item.Game1TimesStats = stats
				}
				if bet_no == 1 {
					item.Game1BetStats = stats
				}
			}
		}
	}

	// 玩家返奖率（玩家携带金额+玩家已提现到账金额+玩家提单待审核金额）/玩家总充值成功金额
	var data_payStats []map[string]any
	err = ck.Select(&data_payStats, `
		SELECT t1.userid userid, t1.diamond, t2.score, t3.amount
			FROM (SELECT userid, (diamond + give_diamond) diamond FROM game.col_user_finance FINAL WHERE userid IN ?) t1
			LEFT JOIN (SELECT userid, SUM(score) score FROM game.col_withdraw_record FINAL WHERE userid IN ? AND order_status IN (1,2) GROUP BY userid) t2 ON t1.userid = t2.userid
			LEFT JOIN (SELECT userid, SUM(amount) amount FROM game.col_trade_record FINAL WHERE userid IN ? AND order_status = 4 GROUP BY userid) t3 ON t1.userid = t3.userid
	`, userids, userids, userids)
	if err != nil {
		return
	}
	for _, data := range data_payStats {
		userid := data["userid"].(string)
		if items, ok := useridDatas[userid]; ok {
			diamond := utils.ToInt64(data["diamond"])
			score := utils.ToInt64(data["score"])
			amount := utils.ToInt64(data["amount"])
			for _, item := range items {
				item.RebateRate = fmt.Sprintf("%.2f%%", ComputeFloat((diamond+score), amount)*100)
			}
		}
	}
	// 总充值金额, 玩家充值成功率
	// 玩家提现成功率
	var data_TradeStats []map[string]any
	err = ck.Select(&data_TradeStats, `
		SELECT t1.userid userid, t2.pay_times_try, t2.pay_times, t2.pay_amount, 
			t3.withdraw_times_try, t3.withdraw_times, t3.withdraw_amount, t3.withdraw_amount_all, t3.withdraw_amount_wait,
			t3.withdraw_amount_freeze, t3.withdraw_amount_processing,
			(t3.withdraw_amount_all - t3.withdraw_amount - t3.withdraw_amount_wait - t3.withdraw_amount_freeze - t3.withdraw_amount_processing) withdraw_amount_other
		FROM (SELECT userid FROM game.col_user_finance FINAL WHERE userid IN ?) t1
		LEFT JOIN (
			SELECT userid, count(*) pay_times_try, 
				SUM(CASE WHEN order_status = 4 THEN 1 ELSE 0 END) pay_times,
				SUM(CASE WHEN order_status = 4 THEN amount ELSE 0 END) pay_amount
			FROM game.col_trade_record FINAL
			WHERE userid IN ?
			GROUP BY userid
		) t2 ON t1.userid = t2.userid
		LEFT JOIN (
			SELECT userid, count(*) withdraw_times_try, 
				SUM(CASE WHEN order_status = 2 THEN 1 ELSE 0 END) withdraw_times,
				SUM(CASE WHEN order_status = 2 THEN amount ELSE 0 END) withdraw_amount,
				SUM(amount) withdraw_amount_all,
				SUM(CASE WHEN order_status = 1 THEN amount ELSE 0 END) withdraw_amount_wait,
				SUM(CASE WHEN order_status = 4 THEN amount ELSE 0 END) withdraw_amount_freeze,
				SUM(CASE WHEN order_status in (6,8) THEN amount ELSE 0 END) withdraw_amount_processing
			FROM game.col_withdraw_record FINAL
			WHERE userid IN ?
			GROUP BY userid
		) t3 ON t1.userid = t3.userid
	`, userids, userids, userids)
	if err != nil {
		return
	}
	for _, data := range data_TradeStats {
		userid := data["userid"].(string)
		if items, ok := useridDatas[userid]; ok {
			pay_times_try := utils.ToInt64(data["pay_times_try"])
			pay_times := utils.ToInt64(data["pay_times"])
			pay_amount := utils.ToInt64(data["pay_amount"])
			withdraw_times_try := utils.ToInt64(data["withdraw_times_try"])
			withdraw_times := utils.ToInt64(data["withdraw_times"])
			withdraw_amount := utils.ToInt64(data["withdraw_amount"])
			withdraw_amount_all := utils.ToInt64(data["withdraw_amount_all"])
			withdraw_amount_wait := utils.ToInt64(data["withdraw_amount_wait"])
			withdraw_amount_freeze := utils.ToInt64(data["withdraw_amount_freeze"])
			withdraw_amount_processing := utils.ToInt64(data["withdraw_amount_processing"])
			withdraw_amount_other := utils.ToInt64(data["withdraw_amount_other"])
			for _, item := range items {
				item.RechargeSuccessRate = fmt.Sprintf("%.2f%%", utils.CaseElse(pay_times_try == 0, 0, float64(pay_times)/float64(pay_times_try)*100))
				item.WithdrawSuccessRate = fmt.Sprintf("%.2f%%", utils.CaseElse(withdraw_times_try == 0, 0, float64(withdraw_times)/float64(withdraw_times_try)*100))
				item.RechargeAmount = fmt.Sprintf("%.2f", ComputeFloat(pay_amount, 100))
				item.RechargeAmount0 = pay_amount
				item.WithdrawAmountAll = fmt.Sprintf("%.2f", ComputeFloat(withdraw_amount_all, 100))
				item.WithdrawAmount = fmt.Sprintf("%.2f", ComputeFloat(withdraw_amount, 100))
				item.WithdrawAmount0 = withdraw_amount
				item.WithdrawWaitAmount = fmt.Sprintf("%.2f", ComputeFloat(withdraw_amount_wait, 100))
				item.WithdrawWaitAmount0 = withdraw_amount_wait
				item.FreezeAmount = fmt.Sprintf("%.2f", ComputeFloat(withdraw_amount_freeze, 100))
				item.ProcessingAmount = fmt.Sprintf("%.2f", ComputeFloat(withdraw_amount_processing, 100))
				item.WithdrawOtherAmount = fmt.Sprintf("%.2f", ComputeFloat(withdraw_amount_other, 100))
				item.WithdrawOtherAmount0 = withdraw_amount_other
			}
		}
	}

	// 玩家当前携带
	var data_finance []map[string]any
	err = ck.Select(&data_finance, `
		SELECT userid, diamond FROM game.col_user_finance FINAL WHERE userid in ?
	`, userids)
	if err != nil {
		return
	}
	for _, data := range data_finance {
		userid := data["userid"].(string)
		if items, ok := useridDatas[userid]; ok {
			diamond := utils.ToInt64(data["diamond"])
			// 计算利润
			for _, item := range items {
				item.FDiamond = fmt.Sprintf("%.2f", ComputeFloat(diamond, 100))
				profit := diamond + item.WithdrawAmount0 + item.WithdrawWaitAmount0 - item.RechargeAmount0
				item.FProfit = fmt.Sprintf("%.2f", ComputeFloat(profit, 100))
			}
		}
	}

	// 提现失败被自动退回金额
	sql_back := `
		SELECT userid, SUM(amount) back_amount
		FROM game.col_withdraw_record  FINAL
		WHERE order_id IN (
			SELECT order_id
			FROM game.col_withdraw_record t1 FINAL
			JOIN game.col_withdraw_record_transfer_order_detail t2 FINAL ON t1.order_id = t2.order_id AND t1.userid = t2.userid
			WHERE t1.order_status = 3 AND t2.status = ? AND userid IN ?
			GROUP BY order_id
		)
		GROUP BY userid
	`
	var data_back5 []map[string]any
	err = ck.Select(&data_back5, sql_back, 5, userids)
	if err != nil {
		return
	}
	// 提单失败被自动退回金额
	var data_back7 []map[string]any
	err = ck.Select(&data_back7, sql_back, 7, userids)
	if err != nil {
		return
	}
	for _, data := range data_back5 {
		userid := data["userid"].(string)
		if items, ok := useridDatas[userid]; ok {
			back_amount := utils.ToInt64(data["back_amount"])
			for _, item := range items {
				item.BackAmount5 = fmt.Sprintf("%.2f", Chip2Float(back_amount))
				// 其他提现金额
				item.WithdrawOtherAmount0 -= back_amount
				item.WithdrawOtherAmount = fmt.Sprintf("%.2f", ComputeFloat(item.WithdrawOtherAmount0, 100))
			}
		}
	}
	for _, data := range data_back7 {
		userid := data["userid"].(string)
		if items, ok := useridDatas[userid]; ok {
			back_amount := utils.ToInt64(data["back_amount"])
			for _, item := range items {
				item.BackAmount7 = fmt.Sprintf("%.2f", Chip2Float(back_amount))
				// 其他提现金额
				item.WithdrawOtherAmount0 -= back_amount
				item.WithdrawOtherAmount = fmt.Sprintf("%.2f", ComputeFloat(item.WithdrawOtherAmount0, 100))
			}
		}
	}

	return
}

// 获取提现订单列表
func (this *payService) WithdrawList0(page, pageSize int, m bson.M, paylist []entity.PayChannel) ([]*entity.WithdrawOrder, error) {
	var list []*entity.WithdrawOrder
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	Withdraws.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList2(list, paylist)
	chipUsdtWithdraw(list)
	return list, nil
}

func chipUsdtWithdraw(withdrawList []*entity.WithdrawOrder) {
	var usdtOrderids []string
	var usdtOrderMap = make(map[string]*entity.WithdrawOrder)
	for _, withdraw := range withdrawList {
		if withdraw.OutChannel == "5002" {
			if withdraw.OutTradeNo != "" {
				usdtOrderids = append(usdtOrderids, withdraw.OutTradeNo)
				usdtOrderMap[withdraw.OutTradeNo] = withdraw
			}
		}
	}
	if len(usdtOrderids) == 0 {
		return
	}

	// usdt 回调信息
	var result []bson.M
	err := DsfWithdrawOrder.Pipe([]bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": usdtOrderids}}},
		{"$project": bson.M{"_id": "$_id", "request_msg": "$request_msg", "call_back_msg": "$call_back_msg"}},
	}).All(&result)
	if err != nil {
		beego.Error("chipUsdtPay error: ", err)
		return
	}
	for _, r := range result {
		id := r["_id"].(string)
		request_msg, ok1 := r["request_msg"].(string)
		call_back_msg, ok2 := r["call_back_msg"].(string)
		order, ok3 := usdtOrderMap[id]
		if !ok3 {
			continue
		}
		if !ok1 {
			continue
		}
		if request_msg != "" {
			res := &entity.UsdtPayWithdrawRespond{}
			if err = json.Unmarshal([]byte(request_msg), &res); err == nil {
				order.USDTMsg += fmt.Sprintf("usdt:%sU", res.Data.ActualAmount)

				if res.Msg != "" && res.Msg != "success" {
					order.USDTMsg += fmt.Sprintf(",%s", res.Msg)
				}
			} else {
				order.USDTMsg = request_msg
			}
		}
		if !ok2 || call_back_msg == "" {
			continue
		}
		cb := &entity.UsdtPayRechargeCallback{}
		if err = json.Unmarshal([]byte(call_back_msg), &cb); err == nil {
			order.USDTMsg += fmt.Sprintf(",交易号:%sU", cb.BlockTransactionId)
		}
	}
}

// 获取提现订单总数
func (this *payService) GetWithdrawTotal(m bson.M) (int64, error) {
	return int64(Count(Withdraws, m)), nil
}

// 根据条件查询对应的提现订单
func (this *payService) GetByWithdrawUser(m bson.M) ([]entity.WithdrawOrder, error) {
	var list []entity.WithdrawOrder
	err := Withdraws.
		Find(m).All(&list)
	return list, err
}

func (this *payService) GetByWithdrawByOrder(orderid string) (*entity.WithdrawOrder, error) {
	orderInfo := new(entity.WithdrawOrder)
	Get(Withdraws, orderid, orderInfo)
	if orderInfo.OrderID == "" {
		return nil, errors.New("订单不存在")
	}
	return orderInfo, nil
}

func (this *payService) GetWithdrawByOrder(orderid string) (*entity.WithdrawOrder, error) {
	orderInfo := new(entity.WithdrawOrder)
	Get(Withdraws, orderid, orderInfo)
	if orderInfo.OrderID == "" {
		return nil, errors.New("订单不存在")
	}
	statusList := map[int]string{
		0:  "全部",
		1:  "审核中",
		2:  "提现成功",
		3:  "提现退回",
		4:  "提现冻结",
		5:  "未知原因付款失败",
		6:  "派单成功",
		7:  "未知原因派单失败",
		8:  "待派单",
		9:  "派单失败需转单",
		10: "派单失败需驳回",
		11: "付款超时急需转单",
		12: "玩家主动申请退款",
		13: "付款失败需转单",
		14: "付款失败需驳回",
		15: "付款超时需转单",
	}
	if orderInfo != nil {
		if len(orderInfo.TransferDetail) > 0 {
			payList, _ := GameService.GetPayChannel()
			for i, item := range orderInfo.TransferDetail {
				for _, v := range payList {
					id, _ := strconv.Atoi(v.Id)
					if item.ChannelId == uint32(id) {
						item.ChannelName = v.Name
					}
				}
				switch item.TType {
				case 1:
					item.TTypeName = "非转单"
				case 2:
					item.TTypeName = "自动转单"
				case 3:
					item.TTypeName = "手动转单"
				}

				switch item.Reason {
				case 1:
					item.ReasonName = "超时转单"
				case 2:
					item.ReasonName = "超时急转单"
				case 3:
					item.ReasonName = "失败转单"
				default:
					item.ReasonName = "-"
				}

				for k, s := range statusList {
					if k == item.Status {
						item.StatusName = s
					}
				}
				orderInfo.TransferDetail[i] = item
			}
		}
	}
	return orderInfo, nil
}

// 转换为分展示
func (this *payService) chipList2(list []*entity.WithdrawOrder, paylist []entity.PayChannel) []*entity.WithdrawOrder {
	userids := make([]string, 0)
	var with_list []entity.SetWithdraw
	SetWithdraws.Find(nil).All(&with_list)

	for k, v := range list {
		// 标记信息
		if v.ErrorMsg != "" {
			v.ErrorMsgs = strings.Split(v.ErrorMsg, ";")
		}

		if v.Diamond > 0 { // 旧数据未记录
			v.FAfterDiamond = Chip2Float(v.Diamond - int64(v.Amount))
		}

		if !v.ETime.IsZero() && v.OrderStatus != 1 {

			// 存在值
			isRecod := false
			time1 := time.Now().UTC()
			time2 := v.ETime
			if v.OrderStatus == 2 {
				time1 = v.WithdrawTime
			}
			// 冻结
			if v.OrderStatus == 4 {
				me := bson.M{}
				me["order_id"] = v.OrderID
				me["opt_id"] = bson.M{"$in": []int{1, 3}}
				var rec_info entity.WithdrawOptRecord
				WithDrawOptRecords.Find(me).Sort("-ctime").One(&rec_info)
				if rec_info.OrderId != "" {
					c := utils.Stamp2Time(rec_info.Ctime)
					time2 = c
				} else {
					// 没有通过的情况下
					isRecod = true
				}
				me["order_id"] = v.OrderID
				me["opt_id"] = 6
				var rec_info1 entity.WithdrawOptRecord
				WithDrawOptRecords.Find(me).One(&rec_info1)
				if rec_info1.OrderId != "" {
					c := utils.Stamp2Time(rec_info1.Ctime)
					time1 = c
				}
			}
			// 退回
			if v.OrderStatus == 3 {
				me := bson.M{}
				me["order_id"] = v.OrderID
				me["opt_id"] = bson.M{"$in": []int{1, 3}}
				var rec_info entity.WithdrawOptRecord
				WithDrawOptRecords.Find(me).Sort("-ctime").One(&rec_info)
				if rec_info.OrderId != "" {
					c := utils.Stamp2Time(rec_info.Ctime)
					time2 = c
				} else {
					isRecod = true
				}
				me["order_id"] = v.OrderID
				me["opt_id"] = 5
				var rec_info1 entity.WithdrawOptRecord
				WithDrawOptRecords.Find(me).One(&rec_info1)
				if rec_info1.OrderId != "" {
					c := utils.Stamp2Time(rec_info1.Ctime)
					time1 = c
				}
			}
			if !isRecod {
				diff := time1.Sub(time2)
				minutes := diff.Minutes()
				v.DelayTime = minutes
				// 判断是否超过预警
				if len(with_list) > 0 {
					for _, i := range with_list {
						if i.Id == v.ShopId {
							if v.DelayTime > float64(i.DelayTime) {
								v.IsDelay = true
							}
						}
					}
				}
			}

		}
		if v.OrderStatus == 2 {
			// 计算到账时效
			if !v.ETime.IsZero() {
				// 存在值
				diff := v.WithdrawTime.Sub(v.ETime)
				minutes := diff.Minutes()
				v.TimeOfArrival = minutes
			}
		}
		// 计算审核时长
		if !v.ETime.IsZero() {
			// 存在值
			if v.ExamineWay != 1 {
				diff := v.ETime.Sub(v.Ctime)
				minutes := diff.Minutes()
				v.AutioTime = minutes
			}
		}

		v.FTotal = Chip2Float(int64(v.Commission + v.Amount))
		v.FCommission = Chip2Float(int64(v.Commission))
		v.FAmount = Chip2Float(int64(v.Amount))
		c, _ := ConvertToIndiaTime(v.Ctime.Unix())
		v.Ctime = c
		c, _ = ConvertToIndiaTime(v.Utime.Unix())
		v.Utime = c
		c, _ = ConvertToIndiaTime(v.WithdrawTime.Unix())
		v.WithdrawTime = c
		c, _ = ConvertToIndiaTime(v.ETime.Unix())
		v.ETime = c
		//检测银行卡
		v.IsBlankNumber = this.CheckAccount(v.Userid, v.BlankNumber)
		userids = append(userids, v.Userid)
		if len(paylist) > 0 {
			for _, p := range paylist {
				if v.OutChannel == p.Id {
					v.OutChannelStr = p.Name
				}
			}
		}
		v.TransferCount = len(v.TransferDetail)
		// if len(v.TransferDetail) > 0 {
		// 	v.IsTransfer = true
		// }
		list[k] = v
	}
	if len(userids) > 0 {
		// 通过用户ID去查询用户设备号
		ulist, _ := PlayerService.GetByUser(userids)
		if len(ulist) > 0 {
			// msg := make(map[string]string, 0)
			ads := make([]string, 0)
			for _, item := range ulist {
				if item.AD_ADID != "" {
					//msg["ad_id"] = item.AD_ADID
					ads = append(ads, item.AD_ADID)
				}
			}
			// ads = append(ads, "d7af61a527db2e50f2eecbf32bd7fec8")
			// ads = append(ads, "ad0ccb9ddebca1edde88ff97b3705724")
			// ad_id=&ad_id=

			// 加载时获取安卓评分
			result, _ := AdGetRequest(ads)
			beego.Trace("result: ", result)
			for index, witem := range list {
				isTag := false
				for _, uitem := range ulist {
					if witem.Userid == uitem.Userid {
						witem.FMoney = Chip2Float(int64(uitem.Money))
						witem.FCashOut = Chip2Float(int64(uitem.CashOut))
						reg, _ := ConvertToIndiaTime(uitem.Ctime.Unix())
						witem.Regtime = reg
					}
					if result != nil {
						for _, item := range result.Data {
							if witem.Userid == uitem.Userid && uitem.AD_ADID == item.Ad_Id && item.Status == 0 {
								witem.AndroidScore = item.Score
								isTag = true
							}
						}
					}
				}
				if !isTag {
					witem.AndroidScore = -1
				}
				list[index] = witem
			}

		}

	}
	return list
}

func (this *payService) CheckAccount(userid string, account string) bool {
	Isflag := false
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"userid":       bson.M{"$ne": userid},
				"blank_number": account,
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
			},
		},
	}
	// operations := []bson.M{m, n}
	result := []bson.M{}
	pipe := Withdraws.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("CheckAccount fail err: ", err)
	}
	if len(result) > 0 {
		Isflag = true
	}
	return Isflag
}

// 更新审核人
func (this *payService) UpdateAuditor(orderinfo *entity.WithdrawOrder) error {
	m := bson.M{"_id": orderinfo.OrderID}
	n := bson.M{"$set": bson.M{"examine_account": orderinfo.ExamineAccount, "examine_way": orderinfo.ExamineWay, "e_time": orderinfo.ETime}}
	if Update(Withdraws, m, n) {
		return nil
	}
	return errors.New("更新失败")
}

// 更新标记状态
func (this *payService) UpdateTag(orderinfo *entity.WithdrawOrder) error {
	//  id string, way int, tag bool, name, msg string
	m := bson.M{"_id": orderinfo.OrderID}
	n := bson.M{"$set": bson.M{"is_tag": orderinfo.IsTag,
		"examine_way":     orderinfo.ExamineWay,
		"examine_account": orderinfo.ExamineAccount,
		"error_msg":       orderinfo.ErrorMsg,
		"e_time":          orderinfo.ETime,
	}}
	if Update(Withdraws, m, n) {
		return nil
	}
	return errors.New("更新失败")
}

// 记录操作时间
func (this *payService) AddWithdrawOptRecord(recod_info *entity.WithdrawOptRecord) error {
	recod_info.Id = bson.NewObjectId().Hex()
	if !Insert(WithDrawOptRecords, recod_info) {
		return errors.New("写入失败:" + recod_info.Id)
	}
	return nil
}

/*
Author:CC
Title:提现记录 -> 自动审核提现订单
*/
func (this *payService) AutomationAudit(timestamp int64, isNotSpecial bool) {
	//根据时间获取待审核的提现订单
	tim := utils.Stamp2Time(timestamp)
	endTime := tim                       //.Format("2006-01-02 15:04:05")
	startTime := tim.Add(-2 * time.Hour) //.Format("2006-01-02 15:04:05")
	m := bson.M{}
	m["ctime"] = bson.M{"$gte": startTime, "$lt": endTime}
	m["order_status"] = 1
	m["is_tag"] = bson.M{"$ne": true}
	list, err := this.GetByWithdrawUser(m)
	if err != nil {
		beego.Error("自动审核查询提现订单失败")
		return
	}

	// if !fixed {
	// 	m2 := bson.M{"_id": bson.M{"$in": []string{"234120604312960153", "234046292571877535"}}}
	// 	m2["order_status"] = 1
	// 	m2["is_tag"] = bson.M{"$ne": true}
	// 	list2, err2 := this.GetByWithdrawUser(m2)
	// 	if err2 != nil {
	// 		beego.Error("自动审核查询提现订单失败2")
	// 		return
	// 	}
	// 	if len(list2) > 0 {
	// 		var fixes []entity.WithdrawOrder
	// 		for _, item := range list2 {
	// 			find := false
	// 			for _, item0 := range list {
	// 				if item0.OrderID == item.OrderID {
	// 					find = true
	// 					break
	// 				}
	// 			}
	// 			if !find {
	// 				fixes = append(fixes, item)
	// 			}
	// 		}
	// 		list = append(list, fixes...)
	// 		fixed = true
	// 	}
	// }

	// fmt.Printf("============withdrawSetting: order count: %d, %#v\n", len(list), withdrawSetting)

	now := NowTime()
	nowTimeF := fmt.Sprintf("%02d:%02d:%02d", now.Hour(), now.Minute(), now.Second())

	if len(list) > 0 {
		withdrawSettings := PayService.GetWithdrawSettings()
		ctypeSettings := make(map[int]*entity.WithdrawSetting)
		for _, setting := range withdrawSettings {
			ctype, _ := strconv.Atoi(setting.Id)
			ctypeSettings[ctype] = setting
		}

		isflag := false
		for _, item := range list {
			// 查询用户类型
			var user map[string]any
			err := ck.Select(&user, `
				SELECT t1.userid userid, t1.state, ifNULL(t2.money, t1.money) money
				FROM game.col_user t1 FINAL LEFT JOIN game.col_user_finance t2 FINAL ON t1.userid = t2.userid 
				WHERE userid = ?
			`, item.Userid)
			if err != nil {
				beego.Error("用户类型查询错误: ", err)
				continue
			}
			ctype, ctypeName := GetChargeType(utils.ToInt64(user["money"]), int(utils.ToInt64(user["state"])))
			withdrawSetting, ok := ctypeSettings[ctype]
			if !ok {
				beego.Error(fmt.Sprintf("用户类型审核配置未设置:%d-%s", ctype, ctypeName))
				continue
			}

			orderid := item.OrderID
			diffDay := (now.Unix() - item.Ctime.Unix()) / 60 / 60 / 24

			// 提现黑名单直接冻结
			m0 := bson.M{"_id": item.Userid}
			wublack_count := Count(WithdrawUserBlacklists, m0)
			if wublack_count > 0 {
				rerr := OrderTag(orderid, utils.CaseElse(item.ErrorMsg != "", item.ErrorMsg+";", "")+fmt.Sprintf("%d:%s:sys:%s", diffDay, nowTimeF, "已封号"))
				if rerr != nil {
					beego.Error("OrderTag fail err: ", rerr)
				}

				msg := new(data.WithdrawOpreate)
				msg.OrderID = orderid
				msg.Op = 3 // 1：通过；2：退回；3：冻结
				result, err := GmRequest(pb.WebWithdraw, pb.CONFIG_UPSERT, msg)
				beego.Trace("result: ", result)
				if err != nil {
					beego.Error(fmt.Sprintf("冻结订单[%s]失败err: %v", orderid, err))
				}
				continue
			}

			// 提现金额是否超过机审单价上限
			if int64(item.Amount) > (withdrawSetting.JQWithdrawMax * 100) {
				rerr := OrderTag(orderid, utils.CaseElse(item.ErrorMsg != "", item.ErrorMsg+";", "")+fmt.Sprintf("%d:%s:sys:%s", diffDay, nowTimeF, "超过机审单价上限"))
				if rerr != nil {
					beego.Error("OrderTag fail err: ", rerr)
				}
				continue
			}

			// 玩家是否超过机审赢钱上限
			// 玩家赢钱=玩家携带金额+玩家提现成功金额+玩家提现待审核金额-玩家充值成功金额
			var winAmount int64
			err = ck.Select(&winAmount, `
				SELECT (
					(SELECT (diamond + give_diamond) diamond FROM game.col_user_finance FINAL WHERE userid = ?)
					+ (SELECT SUM(score) score FROM game.col_withdraw_record FINAL WHERE userid = ? AND order_status IN (1,2))
					- (SELECT SUM(amount) amount FROM game.col_trade_record FINAL WHERE userid = ? AND order_status = 4)
				) win_amount
			`, item.Userid, item.Userid, item.Userid)
			if err != nil {
				beego.Error(fmt.Errorf("玩家赢钱金额查询失败: %s, %v", item.Userid, err))
				continue
			}
			if winAmount > (withdrawSetting.JQWinMax * 100) {
				rerr := OrderTag(orderid, utils.CaseElse(item.ErrorMsg != "", item.ErrorMsg+";", "")+fmt.Sprintf("%d:%s:sys:%s", diffDay, nowTimeF, "超过机审赢钱上限"))
				if rerr != nil {
					beego.Error("OrderTag fail err: ", rerr)
				}
				continue
			}

			// 玩家是否超过机审返奖率上限
			// 玩家总返奖率=（玩家携带金额+玩家已提现到账金额+玩家提单待审核金额）/ 玩家总充值成功金额
			var winRate float64
			err = ck.Select(&winRate, `
				SELECT win_rate FROM (
					SELECT 
						(SELECT (diamond + give_diamond) diamond FROM game.col_user_finance FINAL WHERE userid = ?) diamond,
						(SELECT SUM(score) score FROM game.col_withdraw_record FINAL WHERE userid = ? AND order_status IN (1,2)) withdraw,
						(SELECT SUM(amount) amount FROM game.col_trade_record FINAL WHERE userid = ? AND order_status = 4) recharge,
						((diamond + withdraw) / (CASE recharge WHEN 0 THEN 0.001 ELSE toFloat64(recharge) END)) win_rate
				) t1
			`, item.Userid, item.Userid, item.Userid)
			if err != nil {
				beego.Error(fmt.Errorf("玩家返奖率查询失败: %s, %v", item.Userid, err))
				continue
			}
			if winRate > withdrawSetting.JQWinRateMax {
				rerr := OrderTag(orderid, utils.CaseElse(item.ErrorMsg != "", item.ErrorMsg+";", "")+fmt.Sprintf("%d:%s:sys:%s", diffDay, nowTimeF, "超过机审返奖率上限"))
				if rerr != nil {
					beego.Error("OrderTag fail err: ", rerr)
				}
				continue
			}

			// 如果是博主账号拒绝提现
			m11 := bson.M{}
			m11["_id"] = item.Userid
			account_count := Count(BloggerAccounts, m11)
			if account_count > 0 {
				rerr := OrderTag(orderid, utils.CaseElse(item.ErrorMsg != "", item.ErrorMsg+";", "")+fmt.Sprintf("%d:%s:sys:%s", diffDay, nowTimeF, "博主账号"))
				if rerr != nil {
					beego.Error("OrderTag fail err: ", rerr)
				}
				continue
			}

			// 判断银行卡是否在黑名单中
			m5 := bson.M{}
			m5["_id"] = item.BlankNumber
			card_count := Count(CardBlacklists, m5)
			if card_count > 0 {
				rerr := OrderTag(orderid, utils.CaseElse(item.ErrorMsg != "", item.ErrorMsg+";", "")+fmt.Sprintf("%d:%s:sys:%s", diffDay, nowTimeF, "银行卡黑名单"))
				if rerr != nil {
					beego.Error("OrderTag fail err: ", rerr)
				}
				continue
			}

			// 支付手机号黑名单
			m10 := bson.M{}
			m10["_id"] = item.Userid
			m10["robot"] = false
			m10["simulation_robot"] = false
			var phone_arr []string
			PlayerUsers.Find(m10).Distinct("phone", &phone_arr)
			if len(phone_arr) > 0 {
				phone_str := phone_arr[0]
				m11 := bson.M{}
				m11["_id"] = phone_str
				phone_count := Count(PayBlackLists, m11)
				if phone_count > 0 {
					rerr := OrderTag(orderid, utils.CaseElse(item.ErrorMsg != "", item.ErrorMsg+";", "")+fmt.Sprintf("%d:%s:sys:%s", diffDay, nowTimeF, "支付号码黑名单"))
					if rerr != nil {
						beego.Error("OrderTag fail err: ", rerr)
					}
					continue
				}
			}
			// 当日单个用户提现次数达到5次之后后进入人工，提现成功次数 >= 5（+当前次数）
			// 当日单个用户提现总额达到10000卢比进入人工
			// today := tim.Format("2006-01-02")
			// s := fmt.Sprintf("%s 00:00:00", today)
			// theday_startTime := utils.Str2Time(s)
			// e := fmt.Sprintf("%s 23:59:59", today)
			// theday_endTime := utils.Str2Time(e)

			// m4 := []bson.M{
			// 	{
			// 		"$match": bson.M{
			// 			"userid": item.Userid,
			// 			"ctime":  bson.M{"$gte": theday_startTime, "$lt": theday_endTime},
			// 			// "order_status": 2,
			// 		},
			// 	},
			// 	{
			// 		"$group": bson.M{
			// 			"_id": "$userid",
			// 			"Count": bson.M{
			// 				"$sum": 1,
			// 			},
			// 			"Score": bson.M{"$sum": bson.M{"$cond": []interface{}{
			// 				bson.M{"$eq": []interface{}{"$order_status", 2}}, "$score", nil,
			// 			}}},
			// 		},
			// 	},
			// }
			// // operations := []bson.M{m, n}
			// result := []bson.M{}
			// Withdraws.Pipe(m4).All(&result)
			// if len(result) > 0 {
			// 	count := 0
			// 	amount := 0
			// 	for _, v := range result {
			// 		count = v["Count"].(int)
			// 		amount = v["Score"].(int)
			// 	}
			// 	xz_amount := amount + int(item.Score)
			// 	if count >= 5 {
			// 		rerr := OrderTag(orderid, utils.CaseElse(item.ErrorMsg != "", item.ErrorMsg+";", "")+fmt.Sprintf("%d:%s:sys:%s", diffDay, nowTimeF, "提现次数当日5次"))
			// 		if rerr != nil {
			// 			beego.Error("OrderTag fail err: ", rerr)
			// 		}
			// 		continue
			// 	}
			// 	if xz_amount >= 1000000 {
			// 		rerr := OrderTag(orderid, utils.CaseElse(item.ErrorMsg != "", item.ErrorMsg+";", "")+fmt.Sprintf("%d:%s:sys:%s", diffDay, nowTimeF, "当日提现总额超过10000"))
			// 		if rerr != nil {
			// 			beego.Error("OrderTag fail err: ", rerr)
			// 		}
			// 		continue
			// 	}
			// 	// if count < 5 && xz_amount >= 1000000 {
			// 	// rerr := OrderTag(orderid, fmt.Sprintf("%d:%s:sys:%s", diffDay, nowTimeF, "当日提现总额超过10000"))
			// 	// 	if rerr != nil {
			// 	// 		beego.Error("OrderTag fail err: ", rerr)
			// 	// 	}
			// 	// 	continue
			// 	// }
			// }

			// 非特殊全部自动过审 不限制金额
			// if !isNotSpecial {
			// 	// 根据用户Id,订单ID 查询充值金额及提现金额
			// 	isPay := CheckRechargeAmount(item.Userid, int(item.Amount))
			// 	if !isPay {
			// 		// 限制
			// 		rerr := OrderTag(orderid, utils.CaseElse(item.ErrorMsg != "", item.ErrorMsg+";", "")+fmt.Sprintf("%d:%s:sys:%s", diffDay, nowTimeF, "充值金额限制"))
			// 		if rerr != nil {
			// 			beego.Error("OrderTag fail err: ", rerr)
			// 		}
			// 		continue
			// 	}
			// }
			// 检测提现玩家的银行卡号是否和其他玩家相同
			if !withdrawSetting.IsBankRepeatable {
				m3 := bson.M{}
				m3["blank_number"] = item.BlankNumber
				m3["userid"] = bson.M{"$ne": item.Userid}
				card_number := Count(Withdraws, m3)
				if card_number > 0 {
					rerr := OrderTag(orderid, utils.CaseElse(item.ErrorMsg != "", item.ErrorMsg+";", "")+fmt.Sprintf("%d:%s:sys:%s", diffDay, nowTimeF, "银行卡号重复"))
					if rerr != nil {
						beego.Error("OrderTag fail err: ", rerr)
					}
					continue
				}
			}
			// if isflag {
			// 	// 修改该订单状态、调用提现审核接口
			// 	aerr := OrderAudio(orderid)
			// 	if aerr != nil {
			// 		beego.Error("OrderAudio fail err: ", aerr)
			// 	}
			// 	continue
			// } else {
			// 	// 限制
			// 	rerr := OrderTag(orderid, "充值金额限制")
			// 	if rerr != nil {
			// 		beego.Error("OrderTag fail err: ", rerr)
			// 	}
			// 	continue
			// }

			// 充值金额未超过1000检测同场五局
			// m6 := []bson.M{
			// 	{
			// 		"$match": bson.M{
			// 			"userid":       item.Userid,
			// 			"order_status": 4,
			// 		},
			// 	},
			// 	{
			// 		"$group": bson.M{
			// 			"_id": "$userid",
			// 			"Amount": bson.M{
			// 				"$sum": "$amount",
			// 			},
			// 		},
			// 	},
			// }
			// // operations := []bson.M{m, n}
			// result6 := []bson.M{}
			// pipe6 := Pays.Pipe(m6)
			// err6 := pipe6.All(&result6)
			// if err6 != nil {
			// 	beego.Error("AutomationAudit fail err: ", err)
			// }
			// TotalPayAmount := 0
			// if len(result6) > 0 {
			// 	TotalPayAmount = result6[0]["Amount"].(int)
			// }
			// if TotalPayAmount < 100000 {

			// 检测提现玩家的设备码
			// pipeline := []bson.M{
			// 	// 根据 userid 查询用户的设备码
			// 	bson.M{"$match": bson.M{"_id": item.Userid}},
			// 	// 查找具有相同设备码的其他用户
			// 	bson.M{"$lookup": bson.M{
			// 		"from":         "col_user",    // 查询的集合名
			// 		"localField":   "ad__adid",    // 本地集合中用于关联的字段
			// 		"foreignField": "ad__adid",    // 目标集合中用于关联的字段
			// 		"as":           "other_users", // 查询结果保存到的字段名
			// 	}},
			// 	// 过滤掉本人
			// 	bson.M{"$project": bson.M{
			// 		"other_users": bson.M{
			// 			"$filter": bson.M{
			// 				"input": "$other_users",
			// 				"as":    "other_user",
			// 				"cond":  bson.M{"$ne": []interface{}{"$$other_user._id", item.Userid}},
			// 			},
			// 		},
			// 	}},
			// }
			// result := []bson.M{}
			// pipe := PlayerUsers.Pipe(pipeline)
			// err := pipe.All(&result)
			// if err != nil {
			// 	beego.Error("AutomationAudit fail err: ", err)
			// }
			// if len(result) > 0 {
			// 	temp := result[0]["other_users"].([]interface{})
			// 	if len(temp) > 0 {
			// 		rerr := OrderTag(orderid, "设备码重复")
			// 		if rerr != nil {
			// 			beego.Error("OrderTag fail err: ", rerr)
			// 		}
			// 		continue
			// 	}
			// }
			// }

			// m3 := []bson.M{
			// 	{
			// 		"$match": bson.M{
			// 			"blank_number": item.BlankNumber,
			// 			"userid":       bson.M{"$ne": item.Userid},
			// 		},
			// 	},
			// 	{
			// 		"$group": bson.M{
			// 			"_id": "$userid",
			// 			"num": bson.M{
			// 				"$sum": 1,
			// 			},
			// 		},
			// 	},
			// }
			// result3 := []bson.M{}
			// pipe3 := Withdraws.Pipe(m3)
			// err = pipe3.All(&result3)
			// if err != nil {
			// 	beego.Error("AutomationAudit fail err: ", err)
			// }

			// if len(result3) >= 1 {
			// 	tep := result3[0]["num"].(int)
			// 	if tep > 1 {
			// 		rerr := OrderTag(orderid, "银行卡号重复")
			// 		if rerr != nil {
			// 			beego.Error("OrderTag fail err: ", rerr)
			// 		}
			// 		continue
			// 	}
			// }

			// 检测提现玩家的IFCS号是否和其他玩家相同
			// m4 := []bson.M{
			// 	{
			// 		"$match": bson.M{
			// 			"ifsc":   item.IFSC,
			// 			"userid": bson.M{"$ne": item.Userid},
			// 		},
			// 	},
			// 	{
			// 		"$group": bson.M{
			// 			"_id": "$userid",
			// 			"num": bson.M{
			// 				"$sum": 1,
			// 			},
			// 		},
			// 	},
			// }
			// result4 := []bson.M{}
			// pipe4 := Withdraws.Pipe(m4)
			// err = pipe4.All(&result4)
			// if err != nil {
			// 	beego.Error("AutomationAudit fail err: ", err)
			// }
			// if len(result4) > 1 {
			// 	rerr := OrderTag(orderid, "IFSC重复")
			// 	if rerr != nil {
			// 		beego.Error("OrderTag fail err: ", rerr)
			// 	}
			// 	continue
			// }

			// // 检测提现玩家是否只玩了百人场。
			// pipeline1 := []bson.M{
			// 	{
			// 		"$match": bson.M{
			// 			// "players": bson.RegEx{Pattern: item.Userid, Options: "i"},
			// 			"players": bson.M{
			// 				"$regex":   fmt.Sprintf("\\b%s\\b", item.Userid),
			// 				"$options": "i",
			// 			},
			// 			"gtype": bson.M{"$nin": []int{2, 3, 7}},
			// 		},
			// 	},
			// 	{
			// 		"$group": bson.M{
			// 			"_id": "$gtype",
			// 			"num": bson.M{
			// 				"$sum": 1,
			// 			},
			// 		},
			// 	},
			// }
			// result1 := []bson.M{}
			// pipe1 := Details.Pipe(pipeline1)
			// err = pipe1.All(&result1)
			// if err != nil {
			// 	beego.Error("AutomationAudit fail err: ", err)
			// }
			// if len(result1) == 0 {
			// 	rerr := OrderTag(orderid, "只玩百人场")
			// 	if rerr != nil {
			// 		beego.Error("OrderTag fail err: ", rerr)
			// 	}
			// 	continue
			// }

			// 检测提现玩家设备分是否≥100分。
			// 通过用户ID去查询用户设备号
			ulist, _ := PlayerService.GetUser(item.Userid)
			if ulist != nil {
				ads := make([]string, 0)
				ads = append(ads, ulist.AD_ADID)
				// 加载时获取安卓评分
				result, _ := AdGetRequest(ads)
				beego.Trace("result: ", result)
				isAndroid := false
				if result != nil {
					if len(result.Data) > 0 {
						for _, item1 := range result.Data {
							if item.Userid == ulist.Userid && item1.Ad_Id == ulist.AD_ADID && item1.Status == 0 {
								if item1.Score >= 100 {
									// 评分大于100
									isAndroid = true
									beego.Trace("安卓评分: ", item1.Score)
								}
							}

						}
					}
				}
				if isAndroid {
					rerr := OrderTag(orderid, utils.CaseElse(item.ErrorMsg != "", item.ErrorMsg+";", "")+fmt.Sprintf("%d:%s:sys:%s", diffDay, nowTimeF, "安卓评分不符合"))
					if rerr != nil {
						beego.Error("OrderTag fail err: ", rerr)
					}
					continue
				}
			}

			// // 充值金额未超过1000检测同场五局
			// m6 := []bson.M{
			// 	{
			// 		"$match": bson.M{
			// 			"userid":       item.Userid,
			// 			"order_status": 4,
			// 		},
			// 	},
			// 	{
			// 		"$group": bson.M{
			// 			"_id": "$userid",
			// 			"Amount": bson.M{
			// 				"$sum": "$amount",
			// 			},
			// 		},
			// 	},
			// }
			// // operations := []bson.M{m, n}
			// result6 := []bson.M{}
			// pipe6 := Pays.Pipe(m6)
			// err6 := pipe6.All(&result6)
			// if err6 != nil {
			// 	beego.Error("AutomationAudit fail err: ", err)
			// }
			// TotalPayAmount := 0
			// if len(result6) > 0 {
			// 	TotalPayAmount = result6[0]["Amount"].(int)
			// }
			// if TotalPayAmount < 100000 {
			// 	// 检测提现玩家申请提现前近200局是否有相同玩家同场玩5局。（局数过多或早期对局，会让数据量太大）
			// 	pipeline2 := []bson.M{
			// 		{
			// 			"$match": bson.M{
			// 				"players": bson.M{
			// 					"$regex":   fmt.Sprintf("\\b%s\\b", item.Userid),
			// 					"$options": "i",
			// 				},
			// 				"gtype": bson.M{"$ne": 7},
			// 				// "begin_time": bson.M{"$exists": true},
			// 			},
			// 		},
			// 		{"$sort": bson.M{"begin_time": -1}},
			// 		{"$limit": 200},
			// 		{
			// 			"$group": bson.M{
			// 				"_id":     "$_id",
			// 				"players": bson.M{"$addToSet": "$players"}, // 为每个房间创建玩家列表
			// 				// "playerIds": bson.M{"$push": "$players"},     // 保存原始的玩家列表
			// 			},
			// 		},
			// 	}
			// 	result2 := []bson.M{}
			// 	pipe2 := Details.Pipe(pipeline2)
			// 	err = pipe2.All(&result2)
			// 	if err != nil {
			// 		beego.Error("AutomationAudit fail err: ", err)
			// 	}
			// 	if len(result2) > 0 {
			// 		counts := make(map[string]int)
			// 		isCheck := false
			// 		for _, game := range result2 {
			// 			temp := game["players"].([]interface{})
			// 			strid := temp[0].(string)
			// 			if strid != "" {
			// 				idarr := strings.Split(strid, ",")
			// 				for _, id := range idarr {
			// 					if id != item.Userid {
			// 						if len(id) <= 7 {
			// 							counts[id]++
			// 						}

			// 					}
			// 				}
			// 			}
			// 		}
			// 		if len(counts) > 0 {
			// 			for _, v := range counts {
			// 				if v >= 5 {
			// 					isCheck = true
			// 				}
			// 			}
			// 			if isCheck {
			// 				rerr := OrderTag(orderid, "同场五局")
			// 				if rerr != nil {
			// 					beego.Error("OrderTag fail err: ", rerr)
			// 				}
			// 				continue
			// 			}
			// 		}
			// 	}
			// }

			// 调用审核通过结果
			if !isflag {
				// 修改该订单状态、调用提现审核接口
				aerr := OrderAudio(orderid)
				if aerr != nil {
					beego.Error("OrderAudio fail err: ", aerr)
				}
			}

		}

		beego.Info(fmt.Sprintf("audit %d orders use %dms", len(list), NowTime().UnixMilli()-now.UnixMilli()))
	} else {
		beego.Info("未查询到可审核的提现订单！")
		// fmt.Println("未查询到可审核的提现订单！")
	}
}

// 根据订单ID审核提现订单
func OrderAudio(orderid string) error {
	name := "system"
	info := new(entity.WithdrawOrder)
	info.OrderID = orderid
	info.ExamineAccount = name
	info.ExamineWay = 1
	info.ETime = bson.Now()
	err := PayService.UpdateAuditor(info)
	rec_info := new(entity.WithdrawOptRecord)
	rec_info.OrderId = orderid
	rec_info.OptName = name
	rec_info.OptId = 1
	rec_info.Ctime = info.ETime.Unix()
	err = PayService.AddWithdrawOptRecord(rec_info)

	if err != nil {
		return err
	}
	msg := new(data.WithdrawOpreate)
	msg.OrderID = orderid
	msg.Op = 1 // 1：通过；2：退回；3：冻结
	result, err := GmRequest(pb.WebWithdraw, pb.CONFIG_UPSERT, msg)
	beego.Trace("result: ", result)
	if err != nil {
		return err
	} else {
		ActionService.Add("withdraw_audit", name,
			"", utils.String(orderid), utils.String(orderid), "")
		return nil
	}
}

// 根据订单ID标记订单
func OrderTag(orderid, msg string) error {
	name := "system"
	info := new(entity.WithdrawOrder)
	info.OrderID = orderid
	info.ExamineAccount = name
	info.ExamineWay = 1
	info.IsTag = true
	info.ErrorMsg = msg
	info.ETime = bson.Now()
	err := PayService.UpdateTag(info)
	rec_info := new(entity.WithdrawOptRecord)
	rec_info.OrderId = orderid
	rec_info.OptName = name
	rec_info.OptId = 8
	rec_info.Ctime = info.ETime.Unix()
	err = PayService.AddWithdrawOptRecord(rec_info)
	if err != nil {
		return err
	} else {
		ActionService.Add("withdraw_tag", name,
			"", utils.String(orderid), utils.String(orderid), "")
	}
	// 邮件通知需要手动审核订单
	sendMsg := &pb.AlertorMail{Subject: "web System Msg", Message: "提现审核：" + msg}
	SendMail(sendMsg)
	return nil
}

// 根据用户Id,订单ID 查询充值金额及提现金额
func CheckRechargeAmount(userid string, amount int) bool {
	isflag := false
	// 获取已充值金额
	m := []bson.M{
		{
			"$match": bson.M{
				"userid":       userid,
				"order_status": 4,
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Amount": bson.M{
					"$sum": "$amount",
				},
			},
		},
	}
	// operations := []bson.M{m, n}
	result := []bson.M{}
	pipe := Pays.Pipe(m)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("CheckRechargeAmount fail err: ", err)
	}
	PayAmount := 0
	if len(result) > 0 {
		PayAmount = result[0]["Amount"].(int)
	}
	// 未充值金额，禁止提现
	if PayAmount <= 0 {
		return isflag
	}
	// 获取可提现金额
	m1 := []bson.M{
		{
			"$match": bson.M{
				"userid":       userid,
				"order_status": bson.M{"$in": []int{1, 2, 6, 8}},
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Amount": bson.M{
					"$sum": "$amount",
				},
			},
		},
	}
	// operations := []bson.M{m, n}
	result1 := []bson.M{}
	pipe1 := Withdraws.Pipe(m1)
	err = pipe1.All(&result1)
	if err != nil {
		beego.Error("CheckRechargeAmount fail err: ", err)
	}
	WAmount := 0
	if len(result1) > 0 {
		WAmount = result1[0]["Amount"].(int)
	}
	total := WAmount // + amount
	if PayAmount <= 100000 {
		// 在总充值可以加十，避免105策略
		PayAmount += 1000
	} else if PayAmount > 100000 && PayAmount <= 200000 {
		// 可以超过总充值的百分之三十
		add_amount := Chip2Float(int64(PayAmount)) * 0.3
		PayAmount += (int(add_amount) * 100)
	} else if PayAmount > 200000 {
		// 可以超过总充值的百分之二十
		add_amount := Chip2Float(int64(PayAmount)) * 0.2
		PayAmount += (int(add_amount) * 100)
	}

	if PayAmount > total {
		isflag = true
	}

	return isflag
}

/*
Author:CC
Title:定时检测是否存在提现失败的订单,存在则发送邮件提醒
*/
func (this *payService) CheckOrderStatus(isReturn bool) error {
	timestamp := utils.Timestamp()
	tim := utils.Stamp2Time(timestamp)
	endTime := tim //.Format("2006-01-02 15:04:05")
	//startTime := tim.Add(-5 * time.Minute) //.Format("2006-01-02 15:04:05")
	startTime := tim.Add(-12 * time.Hour)
	m := bson.M{}
	// m["utime"] = bson.M{"$gte": startTime, "$lt": endTime}
	m["utime"] = bson.M{"$gte": startTime, "$lt": endTime}
	m["order_status"] = bson.M{"$in": []int{5, 7}}
	list, err := this.GetByWithdrawUser(m)
	if err != nil {
		return errors.New("查询提现订单失败")
	}
	// if len(list) > 0 {
	// 	sendMsg := &pb.AlertorMail{Subject: "System Msg", Message: "1001"}
	// 	SendMail(sendMsg)
	// }

	if isReturn {
		// 已开启自动退回
		for _, item := range list {
			isExecute := false
			if len(item.TransferDetail) > 0 {
				for _, v := range item.TransferDetail {
					// 存在手动转单 就不走自动退回
					if v.TType == 3 {
						isExecute = true
					}
				}
			}
			if isExecute {
				continue
			}
			name := "system"
			orderid := item.OrderID
			info := new(entity.WithdrawOrder)
			info.OrderID = orderid
			info.ExamineAccount = name
			info.ExamineWay = 1
			info.ETime = bson.Now()
			err := PayService.UpdateAuditor(info)
			if err != nil {
				beego.Error("更新提现订单失败: ", err)
				return errors.New("更新提现订单失败")
			}
			rec_info := new(entity.WithdrawOptRecord)
			rec_info.OrderId = orderid
			rec_info.OptName = name
			rec_info.OptId = 2
			rec_info.Ctime = info.ETime.Unix()
			err = PayService.AddWithdrawOptRecord(rec_info)
			if err != nil {
				beego.Error("更新提现订单失败err: ", err)
				return errors.New("更新提现订单失败")
			}
			msg := new(data.WithdrawOpreate)
			msg.OrderID = orderid
			msg.Op = 2
			result, err := GmRequest(pb.WebWithdraw, pb.CONFIG_UPSERT, msg)
			if err != nil {
				beego.Error("退回提现订单失败: ", err)
				return errors.New("退回提现订单失败")
			}
			beego.Trace("result: ", result)
			ActionService.Add("withdraw_return_system", name,
				"", utils.String(orderid), utils.String(orderid), "")
		}

	}
	return nil
}

// 获取修改金币日志列表
func (this *payService) CoinLogList(page, pageSize int, m bson.M) ([]entity.GoldGiftLog, error) {
	var list []entity.GoldGiftLog
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	GoldGiftLogs.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList3(list)
	return list, nil
}

// 获取修改金币日志总数
func (this *payService) GetCoinLogTotal(m bson.M) (int64, error) {
	return int64(Count(GoldGiftLogs, m)), nil
}

// 转换为分展示
func (this *payService) chipList3(list []entity.GoldGiftLog) []entity.GoldGiftLog {
	for k, v := range list {
		v.FScore = Chip2Float(int64(v.Score))
		v.FCurScore = Chip2Float(int64(v.CurScore))
		v.FChangeScore = Chip2Float(int64(v.ChangeScore))
		c, _ := ConvertToIndiaTime(v.Ctime.Unix())
		v.Ctime = c
		list[k] = v
	}
	return list
}

// 获取提现配置
func (this *payService) SetWithdrawList(page, pageSize int, m bson.M) ([]entity.SetWithdraw, error) {
	var list []entity.SetWithdraw
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	SetWithdraws.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList4(list)
	return list, nil
}

// 获取提现配置总数
func (this *payService) GetSetWithdrawTotal(m bson.M) (int64, error) {
	return int64(Count(SetWithdraws, m)), nil
}

// 转换为分展示
func (this *payService) chipList4(list []entity.SetWithdraw) []entity.SetWithdraw {
	for k, v := range list {
		v.FCostDiamond = Chip2Float(int64(v.CostDiamond))
		v.FGetCash = Chip2Float(int64(v.GetCash))
		v.FCommission = Chip2Float(int64(v.Commission))
		c, _ := ConvertToIndiaTime(v.Ctime.Unix())
		v.Ctime = c
		list[k] = v
	}
	return list
}

// 根据ID获取配置
func (this *payService) GetSetWithdraw(id int) (*entity.SetWithdraw, error) {
	info := new(entity.SetWithdraw)
	GetByInt(SetWithdraws, id, info)
	if info == nil {
		return info, errors.New("配置不存在")
	}
	return info, nil
}

// 添加提现配置
func (this *payService) AddSetWithdraw(with *entity.SetWithdraw) error {
	id_str, err := this.GenerateID("last_set_withdrap_id")
	if err != nil {
		return err
	}
	num, err := strconv.ParseInt(id_str, 10, 32)
	if err != nil {
		return err
	}
	intValue := int32(num)
	with.Id = intValue
	with.Ctime = bson.Now()
	if !Insert(SetWithdraws, with) {
		return errors.New("写入失败:" + id_str)
	}
	return nil
}

// 更新提现配置
func (this *payService) UpdateSetWithdraw(with *entity.SetWithdraw) error {
	m := bson.M{"_id": with.Id}
	n := bson.M{
		"name":           with.Name,
		"limit":          with.Limit,
		"recharge_limit": with.RechargeLimit,
		"flowing_limit":  with.FlowingLimit,
		"cost_diamond":   with.CostDiamond,
		"get_cash":       with.GetCash,
		"commission":     with.Commission,
		"delay_time":     with.DelayTime,
		"index":          with.Index,
		"ctime":          bson.Now()}
	if Update(SetWithdraws, m, bson.M{"$set": n}) {
		return nil
	}
	return errors.New("更新失败")
}

// 删除提现配置
func (this *payService) DelSetWithdraw(id int32) error {
	if Delete(SetWithdraws, bson.M{"_id": id}) {
		return nil
	}
	return errors.New("移除失败")
}

// 生成ID
func (this *payService) GenerateID(idName string) (string, error) {
	gen := new(entity.ShopIDGen)
	gen.Id = idName
	Get(GenIDs, gen.Id, gen)
	if gen.LastUserId == "" {
		gen.LastUserId = "1"
	}
	id := gen.LastUserId
	gen.LastUserId = utils.StringAdd(id)
	if Upsert(GenIDs, bson.M{"_id": gen.Id}, gen) {
		return id, nil
	}
	return id, errors.New("生成错误")
}

// 添加商品
func (this *payService) AddShop(shop *entity.Shop) error {
	var err error
	shop.Id, err = this.GenerateID("last_shop_id")
	if err != nil {
		return errors.New("写入失败")
	}
	if !Insert(Shops, shop) {
		return errors.New("写入失败:" + shop.Id)
	}
	return nil
}

func (this *payService) UpdateShop(shop *entity.Shop) error {
	m := bson.M{"_id": shop.Id}
	n := bson.M{
		"weight":        shop.Weight,
		"show":          shop.Show,
		"vb":            shop.VB,
		"number":        shop.Number,
		"give":          shop.Give,
		"give_type":     shop.GiveType,
		"flow_multiple": shop.FlowMultiple,
		"price":         shop.Price,
		"name":          shop.Name,
	}
	if Update(Shops, m, bson.M{"$set": n}) {
		return nil
	}
	return errors.New("更新失败")
}

// 获取商品
func (this *payService) GetShop(id string) (*entity.Shop, error) {
	shop := new(entity.Shop)
	Get(Shops, id, shop)
	if shop.Id == "" {
		return shop, errors.New("商品不存在")
	}
	return shop, nil
}

// 移除商品
func (this *payService) DelShop(id string) error {
	if Update(Shops, bson.M{"_id": id}, bson.M{"$set": bson.M{"del": 1}}) {
		return nil
	}
	return errors.New("移除失败")
}

// 获取列表
func (this *payService) GetShopList(page, pageSize int, m bson.M) ([]entity.Shop, error) {
	var list []entity.Shop
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "", false)
	err := Shops.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList5(list)
	return list, err
}

// 获取总数
func (this *payService) GetShopListTotal(m bson.M) (int64, error) {
	return int64(Count(Shops, m)), nil
}

// 转换为分展示
func (this *payService) chipList5(list []entity.Shop) []entity.Shop {
	for k, v := range list {
		v.FPrice = Chip2Float(int64(v.Price))
		v.FNumber = Chip2Float(int64(v.Number))
		v.FVB = Chip2Float(int64(v.VB))
		list[k] = v
	}
	return list
}

// 获取支付黑名单
func (this *payService) GetPayBlackList(page, pageSize int, m bson.M) ([]entity.PayBlackList, error) {
	var list []entity.PayBlackList
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	PayBlackLists.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	return list, nil
}

// 获取支付黑名单总数
func (this *payService) GetPayBlackListTotal(m bson.M) (int64, error) {
	return int64(Count(PayBlackLists, m)), nil
}

// 添加支付黑名单
func (this *payService) AddPayBlackList(info *entity.PayBlackList) error {
	if info.Phone == "" {
		return errors.New("手机号不能为空")
	}
	info.Ctime = bson.Now()
	if !Insert(PayBlackLists, info) {
		return errors.New("写入失败:" + info.Phone)
	}
	return nil
}

// 删除支付黑名单
func (this *payService) DelPayBlackList(info *entity.PayBlackList) error {
	if info.Phone == "" {
		return errors.New("手机号不能为空")
	}
	m := bson.M{"_id": info.Phone}
	if Delete(PayBlackLists, m) {
		return nil
	}
	return errors.New("更新失败")
}

// 添加提现用户黑名单
func (this *payService) AddWithdrawUserBlacklist(info *entity.WithdrawUserBlacklist) error {
	if info.Userid == "" {
		return errors.New("用户id不能为空")
	}
	info.Ctime = bson.Now()
	if !Upsert(WithdrawUserBlacklists, bson.M{"_id": info.Userid}, info) {
		return errors.New("写入失败:" + info.Userid)
	}
	return nil
}

func (this *payService) GetPayBlackListById(id string) (*entity.PayBlackList, error) {
	info := new(entity.PayBlackList)
	Get(PayBlackLists, id, info)
	if info.Phone == "" {
		return info, errors.New("手机号不存在")
	}
	return info, nil
}

func (this *payService) AddOrUpdateUserCashRecod(rate *entity.UserCashRecod) error {
	info := new(entity.UserCashRecod)
	GetByQ(UserCashRecods, bson.M{"_id": rate.Userid}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Userid}
		n := bson.M{
			"pay_amount":       rate.PayAmount,
			"pay_count":        rate.PayCount,
			"withdraw_amount":  rate.WithdrawAmount,
			"withdraw_count":   rate.WithdrawCount,
			"gain_amout":       rate.GainAmout,
			"first_pay_amount": rate.FirstPayAmount,
			"first_pay_date":   rate.FirstPayDate,
			"card_number":      rate.CardNumber,
		}
		if Update(UserCashRecods, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + rate.Userid)
	} else {
		// 新增
		if !Insert(UserCashRecods, rate) {
			return errors.New("写入失败:" + rate.Userid)
		}
		return nil
	}
}

/*
提现汇总统计接口
*/
func (this *payService) GetWithdrawStats(startdate, enddate string, startMoney, endMoney int) ([]entity.WithdrawStatsData, error) {
	list := make([]entity.WithdrawStatsData, 0)
	// 根据日期获取提现订单
	s := fmt.Sprintf("%s 00:00:00", startdate)
	sTime := utils.Str2Time(s, Location())
	e := fmt.Sprintf("%s 23:59:59", enddate)
	eTime := utils.Str2Time(e, Location())
	day := int(eTime.Sub(sTime).Hours() / 24)

	for i := 0; i <= day; i++ {
		info := new(entity.WithdrawStatsData)
		// 根据每一天计算数据
		startTime := sTime.AddDate(0, 0, i)
		end_date := startTime.Format("2006-01-02")
		e := fmt.Sprintf("%s 23:59:59", end_date)
		endTime := utils.Str2Time(e, Location())
		info.Date = startTime.Unix()
		info.SDate = startTime.Format("2006-01-02")
		n := bson.M{}
		n["ctime"] = bson.M{"$gte": startTime, "$lte": endTime}
		if startMoney != 0 && endMoney != 0 {
			n["score"] = bson.M{"$gte": (startMoney * 100), "$lte": (endMoney * 100)}
		} else if startMoney != 0 && endMoney == 0 {
			n["score"] = bson.M{"$gte": (startMoney * 100)}
		} else if startMoney == 0 && endMoney != 0 {
			n["score"] = bson.M{"$lte": (endMoney * 100)}
		}

		pipeline := []bson.M{
			{
				"$match": n,
			},
			{
				"$group": bson.M{
					"_id":          nil,
					"order_count":  bson.M{"$sum": 1},
					"total_amount": bson.M{"$sum": "$amount"},
					"pass_count": bson.M{"$sum": bson.M{"$cond": []interface{}{
						bson.M{"$in": []interface{}{"$order_status", []int{2, 6, 7, 8, 9, 10, 11}}}, 1, nil,
					}}},
					"pass_amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
						bson.M{"$in": []interface{}{"$order_status", []int{2, 6, 7, 8, 9, 10, 11}}}, "$amount", nil,
					}}},
					"rgPassCount": bson.M{"$sum": bson.M{"$cond": []interface{}{
						bson.M{"$and": []interface{}{
							bson.M{"$in": []interface{}{"$order_status", []int{2, 6, 7, 8, 9, 10, 11}}},
							bson.M{"$eq": []interface{}{"$examine_way", 2}},
						}}, 1, 0,
					}}},
					"jqPassCount": bson.M{"$sum": bson.M{"$cond": []interface{}{
						bson.M{"$and": []interface{}{
							bson.M{"$in": []interface{}{"$order_status", []int{2, 6, 7, 8, 9, 10, 11}}},
							bson.M{"$eq": []interface{}{"$examine_way", 1}},
						}}, 1, 0,
					}}},
					"jj_count": bson.M{"$sum": bson.M{"$cond": []interface{}{
						bson.M{"$eq": []interface{}{"$order_status", 3}}, 1, nil,
					}}},
					"jj_Amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
						bson.M{"$eq": []interface{}{"$order_status", 3}}, "$amount", nil,
					}}},
					"successful_count": bson.M{"$sum": bson.M{"$cond": []interface{}{
						bson.M{"$eq": []interface{}{"$order_status", 2}}, 1, nil,
					}}},
					"successful_amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
						bson.M{"$eq": []interface{}{"$order_status", 2}}, "$amount", nil,
					}}},
					"rgCount": bson.M{"$sum": bson.M{"$cond": []interface{}{
						bson.M{"$and": []interface{}{
							bson.M{"$in": []interface{}{"$order_status", []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11}}},
							bson.M{"$eq": []interface{}{"$examine_way", 2}},
						}}, 1, 0,
					}}},
					"jqCount": bson.M{"$sum": bson.M{"$cond": []interface{}{
						bson.M{"$and": []interface{}{
							bson.M{"$in": []interface{}{"$order_status", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}}},
							bson.M{"$eq": []interface{}{"$examine_way", 1}},
						}}, 1, 0,
					}}},
				},
			},
		}
		result := []bson.M{}
		Withdraws.Pipe(pipeline).All(&result)
		if len(result) > 0 {
			beego.Info("查询到提现记录数据！")
			for _, v := range result {
				total := v["order_count"].(int)
				total_amount := v["total_amount"].(int)
				pass_count := v["pass_count"].(int)
				pass_amount := v["pass_amount"].(int)
				rgPassCount := v["rgPassCount"].(int)
				jqPassCount := v["jqPassCount"].(int)
				jj_count := v["jj_count"].(int)
				jj_Amount := v["jj_Amount"].(int)
				successful_count := v["successful_count"].(int)
				successful_amount := v["successful_amount"].(int)
				rgCount := v["rgCount"].(int)
				jqCount := v["jqCount"].(int)

				info.TotalNumber = int64(total)
				info.TotalAmount = Chip2Float(int64(total_amount))
				info.PassNumber = int64(pass_count)
				info.PassAmount = Chip2Float(int64(pass_amount))
				info.RgPassNumber = int64(rgPassCount)
				info.JqPassNumber = int64(jqPassCount)
				info.RgAuditNUmber = int64(rgCount)
				info.JqAuditNUmber = int64(jqCount)
				info.RefuseNumber = int64(jj_count)
				info.RefuseAmount = Chip2Float(int64(jj_Amount))
				info.SuccessfulNumber = int64(successful_count)
				info.SuccessfulAmount = Chip2Float(int64(successful_amount))

			}
		}
		// 占比计算
		info.RgAuditRatio = ComputeFloat(info.RgAuditNUmber, info.TotalNumber) * 100
		info.JqAuditRatio = ComputeFloat(info.JqAuditNUmber, info.TotalNumber) * 100
		if info.RefuseAmount > 0 {
			info.RefuseAvg = info.RefuseAmount / float64(info.RefuseNumber)
		}
		if info.SuccessfulAmount > 0 {
			info.SuccessfulAvg = info.SuccessfulAmount / float64(info.SuccessfulNumber)
		}

		info.PassRatio = ComputeFloat(info.PassNumber, info.TotalNumber) * 100
		info.SuccessfulRatio = ComputeFloat(info.SuccessfulNumber, info.PassNumber) * 100

		// 转单数
		m := bson.M{}
		// 手动复制 n 到 m
		for k, v := range n {
			m[k] = v
		}
		m["transfer_detail"] = bson.M{"$ne": nil}
		// m["transfer_detail"] = bson.M{"$size": bson.M{"$gt": 1}}
		quy1 := []bson.M{
			{"$match": m},
			{
				"$project": bson.M{
					"count": bson.M{"$size": "$transfer_detail"},
				},
			},
			{
				"$match": bson.M{
					"count": bson.M{"$gt": 1},
				},
			},
		}
		// m["$expr"] = bson.M{"$gt": []interface{}{bson.M{"$strLenCP": "$transfer_detail"}, 1}}
		var withlist []entity.WithdrawOrder
		var res1 []bson.M
		Withdraws.Pipe(quy1).All(&res1)
		t_number := 0
		t_cg_number := 0
		oid_arr := make([]string, 0)
		for _, tid := range res1 {
			oid := tid["_id"].(string)
			oid_arr = append(oid_arr, oid)
		}
		t_query := bson.M{}
		t_query["_id"] = bson.M{"$in": oid_arr}
		Withdraws.Find(t_query).All(&withlist)
		for _, with := range withlist {
			if len(with.TransferDetail) > 0 {
				isflag := false
				for _, v := range with.TransferDetail {
					if v.Status != 1 {
						isflag = true

					}
				}
				if isflag {
					t_number += 1
					if with.OrderStatus == 2 {
						// 转单后成功记录
						t_cg_number += 1
					}
				}
			}
		}
		info.ZdNumber = int64(t_number)
		info.ZdSGNumber = int64(t_cg_number)
		if t_cg_number != 0 {
			info.ZdRatio = ComputeFloat(int64(t_cg_number), int64(t_number)) * 100
		}

		// 计算重复付款单数
		c := bson.M{}
		// 手动复制 n 到 m
		for k, v := range n {
			c[k] = v
		}
		c["repeat_times"] = bson.M{"$gte": 2}
		c["order_status"] = 2
		var cflist []entity.WithdrawOrder
		Withdraws.Find(c).All(&cflist)
		// 重复付款订单数
		cf_number := len(cflist)
		// 重复付款成功的总次数
		cf_count := 0
		// 重复付款金额
		cf_amount := 0
		for _, with := range cflist {
			cut := with.RepeatTimes
			num := (cut - 1)
			amt := num * int(with.Amount)
			cf_amount += amt
			cf_count += num
		}
		info.CfNumber = int64(cf_number)
		info.CfCount = int64(cf_count)
		info.CfAmount = int64(cf_amount)
		if cf_amount != 0 {
			info.FCfAmount = Chip2Float(int64(cf_amount))
		}
		if cf_count != 0 {
			info.CfFrequency = ComputeFloat(int64(cf_count), int64(cf_number))
		}

		// 计算平均审核时长
		query := bson.M{}
		// 手动复制 m 到 n
		for k, v := range n {
			query[k] = v
		}
		query["order_status"] = bson.M{"$in": []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11}}
		query["examine_way"] = 2
		pipeline1 := []bson.M{
			{
				"$match": query,
			},
			{
				"$group": bson.M{
					"_id": bson.M{
						"orderid": "$_id",
						"ctime":   "$ctime",
						"etime":   "$e_time",
					},
				},
			},
			{
				"$project": bson.M{
					"id":    "$_id.orderid",
					"ctime": "$_id.ctime",
					"etime": "$_id.etime",
				},
			},
		}
		result1 := []bson.M{}
		Withdraws.Pipe(pipeline1).All(&result1)
		totalAudit := float64(0.0)
		for _, a := range result1 {
			oid := a["id"].(string)
			ctime := a["ctime"].(time.Time)
			etime := a["etime"].(time.Time)
			recd := bson.M{}
			recd["order_id"] = oid
			recd["opt_id"] = bson.M{"$in": []int{1, 3}}
			var otpInfo entity.WithdrawOptRecord
			WithDrawOptRecords.Find(recd).Sort("-ctime").One(&otpInfo)
			if otpInfo.OrderId != "" {
				c := utils.Stamp2Time(otpInfo.Ctime)
				etime = c
			}

			diff := etime.Sub(ctime)
			minutes := diff.Minutes()
			totalAudit += minutes
		}
		// 计算滞单时长
		query["order_status"] = bson.M{"$in": []int{2, 4, 3, 5, 6, 7, 8, 9, 10, 11}}
		query["examine_way"] = bson.M{"$in": []int{1, 2}}
		pipeline2 := []bson.M{
			{
				"$match": query,
			},
			{
				"$group": bson.M{
					"_id": bson.M{
						"orderid":      "$_id",
						"etime":        "$e_time",
						"pay_time":     "$pay_time",
						"order_status": "$order_status",
					},
				},
			},
			{
				"$project": bson.M{
					"id":           "$_id.orderid",
					"etime":        "$_id.etime",
					"pay_time":     "$_id.pay_time",
					"order_status": "$_id.order_status",
				},
			},
		}
		result2 := []bson.M{}
		Withdraws.Pipe(pipeline2).All(&result2)
		totalDelay := float64(0.0)
		for _, a := range result2 {
			oid := a["id"].(string)
			ctime := a["etime"].(time.Time)
			etime := a["pay_time"].(time.Time)
			o_status := a["order_status"].(int)
			isRecod := false
			// 退回
			if o_status == 3 {
				me := bson.M{}
				me["order_id"] = oid
				me["opt_id"] = bson.M{"$in": []int{1, 3}}
				var rec_info entity.WithdrawOptRecord
				WithDrawOptRecords.Find(me).Sort("-ctime").One(&rec_info)
				if rec_info.OrderId != "" {
					c := utils.Stamp2Time(rec_info.Ctime)
					ctime = c
				} else {
					// 没有通过的情况下
					isRecod = true
				}
				me["order_id"] = oid
				me["opt_id"] = 5
				var rec_info1 entity.WithdrawOptRecord
				WithDrawOptRecords.Find(me).One(&rec_info1)
				if rec_info1.OrderId != "" {
					c := utils.Stamp2Time(rec_info1.Ctime)
					etime = c
				}
			}
			// 冻结
			if o_status == 4 {
				me := bson.M{}
				me["order_id"] = oid
				me["opt_id"] = bson.M{"$in": []int{1, 3}}
				var rec_info entity.WithdrawOptRecord
				WithDrawOptRecords.Find(me).Sort("-ctime").One(&rec_info)
				if rec_info.OrderId != "" {
					c := utils.Stamp2Time(rec_info.Ctime)
					ctime = c
				} else {
					// 没有通过的情况下
					isRecod = true
				}
				me["order_id"] = oid
				me["opt_id"] = 6
				var rec_info1 entity.WithdrawOptRecord
				WithDrawOptRecords.Find(me).One(&rec_info1)
				if rec_info1.OrderId != "" {
					c := utils.Stamp2Time(rec_info1.Ctime)
					etime = c
				}
			}
			if !etime.IsZero() {
				etime = time.Now().UTC()
			}
			recd := bson.M{}
			recd["order_id"] = oid
			recd["opt_id"] = bson.M{"$in": []int{1, 3}}
			var otpInfo entity.WithdrawOptRecord
			WithDrawOptRecords.Find(recd).Sort("-ctime").One(&otpInfo)
			if otpInfo.OrderId != "" {
				c := utils.Stamp2Time(otpInfo.Ctime)
				ctime = c
			}
			if !isRecod {
				diff := etime.Sub(ctime)
				minutes := diff.Minutes()
				if minutes < 0 {
					etime = time.Now().UTC()
					diff = etime.Sub(ctime)
					minutes = diff.Minutes()
				}
				totalDelay += minutes
			}
		}

		// 计算到账时长
		query["order_status"] = 2
		query["examine_way"] = bson.M{"$in": []int{1, 2}}
		pipeline3 := []bson.M{
			{
				"$match": query,
			},
			{
				"$group": bson.M{
					"_id": bson.M{
						"orderid":  "$_id",
						"etime":    "$e_time",
						"pay_time": "$pay_time",
					},
				},
			},
			{
				"$project": bson.M{
					"id":       "$_id.orderid",
					"etime":    "$_id.etime",
					"pay_time": "$_id.pay_time",
				},
			},
		}
		result3 := []bson.M{}
		Withdraws.Pipe(pipeline3).All(&result3)
		totalArrival := float64(0.0)
		for _, a := range result3 {
			oid := a["id"].(string)
			recd := bson.M{}
			recd["order_id"] = oid
			recd["opt_id"] = bson.M{"$in": []int{1, 3}}
			var otpInfo entity.WithdrawOptRecord
			WithDrawOptRecords.Find(recd).Sort("-ctime").One(&otpInfo)
			ctime := a["etime"].(time.Time)
			etime := a["pay_time"].(time.Time)
			if otpInfo.OrderId != "" {
				c := utils.Stamp2Time(otpInfo.Ctime)
				ctime = c
			}

			diff := etime.Sub(ctime)
			minutes := diff.Minutes()
			totalArrival += minutes
		}
		// 平均时长
		info.TotalAuditTime = totalAudit
		info.DelayTimeTime = totalDelay
		info.ArrivalTimeTime = totalArrival
		if totalAudit > 0 {
			info.AuditTimeAvg = totalAudit / float64(info.RgAuditNUmber)
		}
		if totalDelay > 0 {
			info.DelayTimeAvg = totalDelay / float64(info.PassNumber)
		}
		if totalArrival > 0 {
			info.ArrivalTimeAvg = totalArrival / float64(info.SuccessfulNumber)
		}

		list = append(list, *info)
	}
	// 按照时间倒序排序
	sort.Slice(list, func(i, j int) bool {
		return list[i].Date > list[j].Date
	})
	totalInfo := new(entity.WithdrawStatsData)
	isTotal := false
	for _, inf := range list {
		totalInfo.TotalNumber += inf.TotalNumber
		totalInfo.TotalAmount += inf.TotalAmount
		totalInfo.PassNumber += inf.PassNumber
		totalInfo.PassAmount += inf.PassAmount
		totalInfo.RgPassNumber += inf.RgPassNumber
		totalInfo.JqPassNumber += inf.JqPassNumber
		totalInfo.RgAuditNUmber += inf.RgAuditNUmber
		totalInfo.JqAuditNUmber += inf.JqAuditNUmber
		totalInfo.RefuseNumber += inf.RefuseNumber
		totalInfo.RefuseAmount += inf.RefuseAmount
		totalInfo.SuccessfulNumber += inf.SuccessfulNumber
		totalInfo.SuccessfulAmount += inf.SuccessfulAmount
		totalInfo.TotalAuditTime += inf.TotalAuditTime
		totalInfo.DelayTimeTime += inf.DelayTimeTime
		totalInfo.ArrivalTimeTime += inf.ArrivalTimeTime
		totalInfo.ZdNumber += inf.ZdNumber
		totalInfo.ZdSGNumber += inf.ZdSGNumber
		totalInfo.CfNumber += inf.CfNumber
		totalInfo.CfCount += inf.CfCount
		totalInfo.CfAmount += inf.CfAmount
		isTotal = true
	}
	if isTotal {
		// 计算值 且加入第一条
		totalInfo.SDate = startdate + "-" + enddate + "的汇总"
		// 占比计算
		totalInfo.RgAuditRatio = ComputeFloat(totalInfo.RgAuditNUmber, totalInfo.TotalNumber) * 100
		totalInfo.JqAuditRatio = ComputeFloat(totalInfo.JqAuditNUmber, totalInfo.TotalNumber) * 100
		if totalInfo.RefuseAmount > 0 {
			totalInfo.RefuseAvg = totalInfo.RefuseAmount / float64(totalInfo.RefuseNumber)
		}
		if totalInfo.SuccessfulAmount > 0 {
			totalInfo.SuccessfulAvg = totalInfo.SuccessfulAmount / float64(totalInfo.SuccessfulNumber)
		}

		totalInfo.PassRatio = ComputeFloat(totalInfo.PassNumber, totalInfo.TotalNumber) * 100
		totalInfo.SuccessfulRatio = ComputeFloat(totalInfo.SuccessfulNumber, totalInfo.PassNumber) * 100
		// 平均时长
		if totalInfo.TotalAuditTime > 0 {
			totalInfo.AuditTimeAvg = totalInfo.TotalAuditTime / float64(totalInfo.RgAuditNUmber)
		}
		if totalInfo.DelayTimeTime > 0 {
			totalInfo.DelayTimeAvg = totalInfo.DelayTimeTime / float64(totalInfo.PassNumber)
		}
		if totalInfo.ArrivalTimeTime > 0 {
			totalInfo.ArrivalTimeAvg = totalInfo.ArrivalTimeTime / float64(totalInfo.SuccessfulNumber)
		}
		if totalInfo.ZdSGNumber > 0 {
			totalInfo.ZdRatio = ComputeFloat(int64(totalInfo.ZdSGNumber), int64(totalInfo.ZdNumber)) * 100
		}
		if totalInfo.CfAmount > 0 {
			totalInfo.FCfAmount = Chip2Float(int64(totalInfo.CfAmount))
		}
		if totalInfo.CfCount != 0 {
			totalInfo.CfFrequency = ComputeFloat(int64(totalInfo.CfCount), int64(totalInfo.CfNumber))
		}

		newSlice := make([]entity.WithdrawStatsData, 0)
		newSlice = append(newSlice, *totalInfo)
		list = append(newSlice, list...)
	}
	return list, nil
}

/*
支付渠道统计
*/
func (this *payService) GetPayChannelStats_old(page, pageSize int, m bson.M, isChannel bool) ([]entity.PayChannelStatsData, error) {
	var list []entity.PayChannelStatsData
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)
	pipeline := []bson.M{}
	if isChannel {
		// 渠道
		pipeline = []bson.M{
			{"$match": m},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":        "$date",
						"pay_channel": "$pay_channel",
					},
					"total_pay":             bson.M{"$sum": "$total_pay"},
					"pay_rate":              bson.M{"$sum": "$pay_rate"},
					"total_pay_amount":      bson.M{"$sum": "$total_pay_amount"},
					"total_withdraw":        bson.M{"$sum": "$total_withdraw"},
					"withdraw_rate":         bson.M{"$sum": "$withdraw_rate"},
					"total_withdraw_amount": bson.M{"$sum": "$total_withdraw_amount"},
					"pay_handling_fee":      bson.M{"$sum": "$pay_handling_fee"},
					"withdraw_handling_fee": bson.M{"$sum": "$withdraw_handling_fee"},
					"yesterday_balance":     bson.M{"$sum": "$yesterday_balance"},
					"today_balance":         bson.M{"$sum": "$today_balance"},
					"balance":               bson.M{"$sum": "$balance"},
					"current_balance":       bson.M{"$sum": "$current_balance"},
					"pay_channel_balance":   bson.M{"$sum": "$pay_channel_balance"},
				},
			},
			{
				"$project": bson.M{
					"_id":                   "$_id.date",
					"date":                  "$_id.date",
					"pay_channel":           "$_id.pay_channel",
					"total_pay":             "$total_pay",
					"pay_rate":              "$pay_rate",
					"total_pay_amount":      "$total_pay_amount",
					"total_withdraw":        "$total_withdraw",
					"withdraw_rate":         "$withdraw_rate",
					"total_withdraw_amount": "$total_withdraw_amount",
					"pay_handling_fee":      "$pay_handling_fee",
					"withdraw_handling_fee": "$withdraw_handling_fee",
					"yesterday_balance":     "$yesterday_balance",
					"today_balance":         "$today_balance",
					"balance":               "$balance",
					"current_balance":       "$current_balance",
					"pay_channel_balance":   "$pay_channel_balance",
				},
			},
			{"$sort": bson.M{"date": -1}},
			{"$skip": skipNum},
			{"$limit": pageSize},
		}
	} else {
		// 全部
		pipeline = []bson.M{
			{"$match": m},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date": "$date",
					},
					"total_pay":             "$total_pay",
					"total_pay_amount":      "$total_pay_amount",
					"total_withdraw":        "$total_withdraw",
					"total_withdraw_amount": "$total_withdraw_amount",
					"pay_handling_fee":      "$pay_handling_fee",
					"withdraw_handling_fee": "$withdraw_handling_fee",
					"yesterday_balance":     "$yesterday_balance",
					"today_balance":         "$today_balance",
					"balance":               "$balance",
					"current_balance":       "$current_balance",
					"pay_channel_balance":   "$pay_channel_balance",
				},
			},
			{
				"$project": bson.M{
					"_id":                   "$_id.date",
					"date":                  "$_id.date",
					"total_pay":             "$total_pay",
					"total_pay_amount":      "$total_pay_amount",
					"total_withdraw":        "$total_withdraw",
					"total_withdraw_amount": "$total_withdraw_amount",
					"pay_handling_fee":      "$pay_handling_fee",
					"withdraw_handling_fee": "$withdraw_handling_fee",
					"yesterday_balance":     "$yesterday_balance",
					"today_balance":         "$today_balance",
					"balance":               "$balance",
					"current_balance":       "$current_balance",
					"pay_channel_balance":   "$pay_channel_balance",
				},
			},
			{"$sort": bson.M{"date": -1}},
			{"$skip": skipNum},
			{"$limit": pageSize},
		}
	}
	err := PayChannelStatss.Pipe(pipeline).All(&list)
	list = this.chipList(list)
	return list, err
}

func (this *payService) GetPayChannelStats_old_1(startdate, enddate string, paychannel int) ([]entity.PayChannelStatsData, error) {
	list := make([]entity.PayChannelStatsData, 0)
	// 根据日期获取提现订单
	s := fmt.Sprintf("%s 00:00:00", startdate)
	sTime := utils.Str2Time(s, Location())
	e := fmt.Sprintf("%s 23:59:59", enddate)
	eTime := utils.Str2Time(e, Location())
	day := int(eTime.Sub(sTime).Hours() / 24)
	var err error
	for i := 0; i <= day; i++ {
		// info := new(entity.PayChannelStatsData)
		// 根据每一天计算数据
		startTime := sTime.AddDate(0, 0, i)
		end_date := startTime.Format("2006-01-02")
		e := fmt.Sprintf("%s 23:59:59", end_date)
		endTime := utils.Str2Time(e, Location())
		startTimestamp := utils.Time2StampToMS(startTime)
		endTimestamp := utils.Time2StampToMS(endTime)
		m := bson.M{}
		m["date"] = bson.M{"$gte": startTimestamp, "$lte": endTimestamp}

		n := bson.M{}
		dsend := endTime.Unix()
		dsstart := startTime.Unix()
		n["ctime"] = bson.M{"$gte": dsstart, "$lte": dsend}
		if paychannel != 0 {
			m["pay_channel"] = paychannel
			n["pay_channel"] = paychannel
		}

		pipeline := []bson.M{}
		result := []bson.M{}
		// result1 := []bson.M{} // 每个支付渠道
		if paychannel != 0 {
			pipeline = []bson.M{
				{
					"$match": m,
				},
				{
					"$group": bson.M{
						"_id": "$pay_channel",
						"ds_amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 1}}, "$amount", 0,
						}}},
						"ds_rate": bson.M{"$last": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 1}}, "$rate", 0.0,
						}}},
						"ds_sj_amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 1}}, "$actual_amount", 0,
						}}},
						"df_amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 2}}, "$amount", 0,
						}}},
						"df_rate": bson.M{"$last": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 2}}, "$rate", 0.0,
						}}},
						"df_sj_amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 2}}, "$actual_amount", 0,
						}}},

						"ds_handling": bson.M{"$sum": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 1}}, "$handling_charge", 0,
						}}},
						"df_handling": bson.M{"$sum": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 2}}, "$handling_charge", 0,
						}}},
					},
				},
			}
		} else {
			pipeline = []bson.M{
				{
					"$match": m,
				},
				{
					"$group": bson.M{
						"_id": nil,
						"ds_amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 1}}, "$amount", 0,
						}}},
						"ds_rate": bson.M{"$max": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 1}}, "$rate", 0.0,
						}}},
						"ds_sj_amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 1}}, "$actual_amount", 0,
						}}},
						"df_amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 2}}, "$amount", 0,
						}}},
						"df_rate": bson.M{"$max": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 2}}, "$rate", 0.0,
						}}},
						"df_sj_amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 2}}, "$actual_amount", 0,
						}}},

						"ds_handling": bson.M{"$sum": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 1}}, "$handling_charge", 0,
						}}},
						"df_handling": bson.M{"$sum": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 2}}, "$handling_charge", 0,
						}}},
					},
				},
			}
			// pipeline1 := []bson.M{
			// 	{
			// 		"$match": m,
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
			// err = PayChannelLogs.Pipe(pipeline1).All(&result1)
		}

		err = PayChannelLogs.Pipe(pipeline).All(&result)
		if len(result) > 0 {
			// 查询每个渠道余额
			var tpblist []entity.ThirdPartyBalance
			ThirdPartyBalances.Find(n).All(&tpblist)

			// 查询昨日0点余额(ps：实际是查询日期当天的开始的余额)
			pipeline2 := []bson.M{
				{"$match": bson.M{
					"date": bson.M{"$lte": startTimestamp},
				}},
				{"$sort": bson.M{"pay_channel": 1, "date": 1}},
				{"$group": bson.M{
					"_id":         "$pay_channel",
					"firstRecord": bson.M{"$last": "$$ROOT"},
				}},
				{"$replaceRoot": bson.M{
					"newRoot": "$firstRecord",
				}},
			}
			var result2 []entity.PayChannelLog
			err = PayChannelLogs.Pipe(pipeline2).All(&result2)

			// 查询当日0点余额(ps：实际是查询日期当天的结束的余额，比如：查询12日 那么就是12日当天23:59:59结束时的余额)
			pipeline3 := []bson.M{
				{"$match": bson.M{
					"date": bson.M{"$gte": startTimestamp, "$lte": endTimestamp},
				}},
				{"$sort": bson.M{"pay_channel": 1, "date": 1}},
				{"$group": bson.M{
					"_id":         "$pay_channel",
					"firstRecord": bson.M{"$last": "$$ROOT"},
				}},
				{"$replaceRoot": bson.M{
					"newRoot": "$firstRecord",
				}},
			}
			var result3 []entity.PayChannelLog
			err = PayChannelLogs.Pipe(pipeline3).All(&result3)

			for _, item := range result {
				// 检查并转换字段类型
				cid := 0
				dsId, _ := item["_id"]
				if dsId == nil {
					cid = 0
				} else {
					cid = dsId.(int)
				}
				ds_amount := ConvertToInt64(item["ds_amount"])
				ds_rate := item["ds_rate"].(float64)
				ds_sj_amount := ConvertToInt64(item["ds_sj_amount"])
				df_amount := ConvertToInt64(item["df_amount"])
				df_rate := item["df_rate"].(float64)
				df_sj_amount := ConvertToInt64(item["df_sj_amount"])
				ds_handling := ConvertToInt64(item["ds_handling"])
				df_handling := ConvertToInt64(item["df_handling"])
				info := new(entity.PayChannelStatsData)
				info.Date = startTime.Unix()
				info.SDate = startTime
				info.PayChannel = uint32(cid)
				info.TotalPay = ds_amount
				info.PayRate = ds_rate
				info.TotalPayAmount = ds_sj_amount
				info.TotalWithdraw = df_amount
				info.WithdrawRate = df_rate
				info.TotalWithdrawAmount = df_sj_amount
				info.PayHandlingFee = ds_handling
				info.WithdrawHandlingFee = df_handling
				yesterdayBalance := int64(0)
				if len(result2) > 0 {
					for _, r := range result2 {
						if paychannel != 0 {
							if r.PayChannel == uint32(paychannel) {
								yesterdayBalance = r.Balance
							}
						} else {
							yesterdayBalance += r.Balance
						}
					}
				}
				todayBalance := int64(0)
				if len(result3) > 0 {
					for _, r := range result3 {
						if paychannel != 0 {
							if r.PayChannel == uint32(paychannel) {
								todayBalance = r.Balance
							}
						} else {
							todayBalance += r.Balance
						}
					}
				}
				info.YesterdayBalance = yesterdayBalance
				info.TodayBalance = todayBalance
				pcBalance := int64(0)
				if len(tpblist) > 0 {
					for _, tpitem := range tpblist {
						if paychannel != 0 {
							if tpitem.PayChannel == uint32(paychannel) {
								pcBalance = tpitem.Balance
							}
						} else {
							pcBalance += tpitem.Balance
						}
					}
				}

				info.PayChannelBalance = pcBalance
				list = append(list, *info)
			}
		}

	}
	if len(list) > 0 {
		// 获取支付渠道名称
		channellist, _ := GameService.GetPayChannel()
		for k, v := range list {
			if v.PayChannel != 0 {
				id := strconv.FormatUint(uint64(v.PayChannel), 10)
				for _, chl := range channellist {
					if id == chl.Id {
						v.PayChannelName = chl.Name
					}
				}
			}
			v.FTotalPay = Chip2Float(v.TotalPay)
			v.FTotalPayAmount = Chip2Float(v.TotalPayAmount)
			v.FTotalWithdraw = Chip2Float(v.TotalWithdraw)
			v.FTotalWithdrawAmount = Chip2Float(v.TotalWithdrawAmount)
			v.FPayHandlingFee = Chip2Float(v.PayHandlingFee)
			v.FWithdrawHandlingFee = Chip2Float(v.WithdrawHandlingFee)
			v.TotalProfit = Chip2Float(v.TotalPay - v.TotalWithdraw)
			v.ActualProfit = Chip2Float(v.TotalPayAmount - v.TotalWithdrawAmount)
			if v.TotalProfit > 0 {
				v.IsTotal = true
			}
			if v.ActualProfit > 0 {
				v.IsActual = true
			}
			v.FYesterdayBalance = Chip2Float(v.YesterdayBalance)
			v.FTodayBalance = Chip2Float(v.TodayBalance)
			v.FPayChannelBalance = Chip2Float(v.PayChannelBalance)
			v.FAmountDifference = Chip2Float(v.YesterdayBalance - v.PayChannelBalance)
			// 如果是当日获取当前余额
			now := time.Now()
			if v.SDate.Year() == now.Year() && v.SDate.Month() == now.Month() && v.SDate.Day() == now.Day() {
				totalBalance := int64(0)
				// 查询当前余额
				startTimestamp := utils.Time2StampToMS(v.SDate)
				pipeline4 := []bson.M{
					{"$match": bson.M{
						"date": bson.M{"$gte": startTimestamp},
					}},
					{"$sort": bson.M{"pay_channel": 1, "date": 1}},
					{"$group": bson.M{
						"_id":         "$pay_channel",
						"firstRecord": bson.M{"$last": "$$ROOT"},
					}},
					{"$replaceRoot": bson.M{
						"newRoot": "$firstRecord",
					}},
				}
				var result4 []entity.PayChannelLog
				err = PayChannelLogs.Pipe(pipeline4).All(&result4)
				for _, rs := range result4 {
					if paychannel != 0 {
						if rs.PayChannel == v.PayChannel {
							totalBalance = rs.Balance
						}
					} else {
						totalBalance += rs.Balance
					}
				}
				v.CurrentBalance = int64(totalBalance)
				v.FCurrentBalance = Chip2Float(int64(totalBalance))
			}
			list[k] = v
		}
	}
	// 使用 sort.Slice() 对用户信息进行排序
	sort.Slice(list, func(i, j int) bool {
		return list[i].Date > list[j].Date
	})
	return list, err
}

func (this *payService) GetPayChannelStats(startdate, enddate string, paychannel int) ([]entity.PayChannelStatsData, error) {
	list := make([]entity.PayChannelStatsData, 0)
	// 根据日期获取提现订单
	s := fmt.Sprintf("%s 00:00:00", startdate)
	sTime := utils.Str2Time(s, Location())
	e := fmt.Sprintf("%s 23:59:59", enddate)
	eTime := utils.Str2Time(e, Location())
	day := int(eTime.Sub(sTime).Hours() / 24)
	var err error
	for i := 0; i <= day; i++ {
		// info := new(entity.PayChannelStatsData)
		// 根据每一天计算数据
		startTime := sTime.AddDate(0, 0, i)
		end_date := startTime.Format("2006-01-02")
		e := fmt.Sprintf("%s 23:59:59", end_date)
		endTime := utils.Str2Time(e, Location())
		startTimestamp := utils.Time2StampToMS(startTime)
		endTimestamp := utils.Time2StampToMS(endTime)
		m := bson.M{}
		m["date"] = bson.M{"$gte": startTimestamp, "$lte": endTimestamp}

		n := bson.M{}
		dsend := endTime.Unix()
		dsstart := startTime.Unix()
		n["ctime"] = bson.M{"$gte": dsstart, "$lte": dsend}
		if paychannel != 0 {
			m["pay_channel"] = paychannel
			n["pay_channel"] = paychannel
		}

		pipeline := []bson.M{}
		result := []bson.M{}
		// result1 := []bson.M{} // 每个支付渠道
		if paychannel != 0 {
			pipeline = []bson.M{
				{
					"$match": m,
				},
				{
					"$group": bson.M{
						"_id": "$pay_channel",
						"ds_amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 1}}, "$amount", 0,
						}}},
						"ds_rate": bson.M{"$last": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 1}}, "$rate", 0.0,
						}}},
						"ds_sj_amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 1}}, "$actual_amount", 0,
						}}},
						"df_amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 2}}, "$amount", 0,
						}}},
						"df_rate": bson.M{"$last": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 2}}, "$rate", 0.0,
						}}},
						"df_sj_amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 2}}, "$actual_amount", 0,
						}}},

						"ds_handling": bson.M{"$sum": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 1}}, "$handling_charge", 0,
						}}},
						"df_handling": bson.M{"$sum": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 2}}, "$handling_charge", 0,
						}}},
					},
				},
			}
		} else {
			pipeline = []bson.M{
				{
					"$match": m,
				},
				{
					"$group": bson.M{
						"_id": nil,
						"ds_amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 1}}, "$amount", 0,
						}}},
						"ds_rate": bson.M{"$max": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 1}}, "$rate", 0.0,
						}}},
						"ds_sj_amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 1}}, "$actual_amount", 0,
						}}},
						"df_amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 2}}, "$amount", 0,
						}}},
						"df_rate": bson.M{"$max": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 2}}, "$rate", 0.0,
						}}},
						"df_sj_amount": bson.M{"$sum": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 2}}, "$actual_amount", 0,
						}}},

						"ds_handling": bson.M{"$sum": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 1}}, "$handling_charge", 0,
						}}},
						"df_handling": bson.M{"$sum": bson.M{"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$p_type", 2}}, "$handling_charge", 0,
						}}},
					},
				},
			}
		}

		err = PayChannelLogs.Pipe(pipeline).All(&result)
		if len(result) > 0 {
			// 查询每个渠道余额
			var tpblist []entity.ThirdPartyBalance
			n2 := []bson.M{
				{"$match": n},
				{"$group": bson.M{
					"_id":         "$ctime",
					"balance":     bson.M{"$sum": "$balance"},
					"own_balance": bson.M{"$sum": "$own_balance"},
				}},
			}
			if paychannel != 0 {
				n2 = []bson.M{
					{"$match": n},
					{"$group": bson.M{
						"_id":         "$pay_channel",
						"balance":     bson.M{"$sum": "$balance"},
						"own_balance": bson.M{"$sum": "$own_balance"},
					}},
				}
			}

			ThirdPartyBalances.Pipe(n2).All(&tpblist)

			// // 查询昨日0点余额(ps：实际是查询日期当天的开始的余额)
			// pipeline2 := []bson.M{
			// 	{"$match": bson.M{
			// 		"date": bson.M{"$lte": startTimestamp},
			// 	}},
			// 	{"$sort": bson.M{"pay_channel": 1, "date": 1}},
			// 	{"$group": bson.M{
			// 		"_id":         "$pay_channel",
			// 		"firstRecord": bson.M{"$last": "$$ROOT"},
			// 	}},
			// 	{"$replaceRoot": bson.M{
			// 		"newRoot": "$firstRecord",
			// 	}},
			// }
			// var result2 []entity.PayChannelLog
			// err = PayChannelLogs.Pipe(pipeline2).All(&result2)

			// // 查询当日0点余额(ps：实际是查询日期当天的结束的余额，比如：查询12日 那么就是12日当天23:59:59结束时的余额)
			// pipeline3 := []bson.M{
			// 	{"$match": bson.M{
			// 		"date": bson.M{"$gte": startTimestamp, "$lte": endTimestamp},
			// 	}},
			// 	{"$sort": bson.M{"pay_channel": 1, "date": 1}},
			// 	{"$group": bson.M{
			// 		"_id":         "$pay_channel",
			// 		"firstRecord": bson.M{"$last": "$$ROOT"},
			// 	}},
			// 	{"$replaceRoot": bson.M{
			// 		"newRoot": "$firstRecord",
			// 	}},
			// }
			// var result3 []entity.PayChannelLog
			// err = PayChannelLogs.Pipe(pipeline3).All(&result3)

			for _, item := range result {
				// 检查并转换字段类型
				cid := 0
				dsId, _ := item["_id"]
				if dsId == nil {
					cid = 0
				} else {
					cid = dsId.(int)
				}
				ds_amount := ConvertToInt64(item["ds_amount"])
				ds_rate := item["ds_rate"].(float64)
				ds_sj_amount := ConvertToInt64(item["ds_sj_amount"])
				df_amount := ConvertToInt64(item["df_amount"])
				df_rate := item["df_rate"].(float64)
				df_sj_amount := ConvertToInt64(item["df_sj_amount"])
				ds_handling := ConvertToInt64(item["ds_handling"])
				df_handling := ConvertToInt64(item["df_handling"])
				info := new(entity.PayChannelStatsData)
				info.Date = startTime.Unix()
				info.SDate = startTime
				info.PayChannel = uint32(cid)
				info.TotalPay = ds_amount
				info.PayRate = ds_rate
				info.TotalPayAmount = ds_sj_amount
				info.TotalWithdraw = df_amount
				info.WithdrawRate = df_rate
				info.TotalWithdrawAmount = df_sj_amount
				info.PayHandlingFee = ds_handling
				info.WithdrawHandlingFee = df_handling
				// yesterdayBalance := int64(0)
				// if len(result2) > 0 {
				// 	for _, r := range result2 {
				// 		if paychannel != 0 {
				// 			if r.PayChannel == uint32(paychannel) {
				// 				yesterdayBalance = r.Balance
				// 			}
				// 		} else {
				// 			yesterdayBalance += r.Balance
				// 		}
				// 	}
				// }
				// todayBalance := int64(0)
				// if len(result3) > 0 {
				// 	for _, r := range result3 {
				// 		if paychannel != 0 {
				// 			if r.PayChannel == uint32(paychannel) {
				// 				todayBalance = r.Balance
				// 			}
				// 		} else {
				// 			todayBalance += r.Balance
				// 		}
				// 	}
				// }
				// info.YesterdayBalance = yesterdayBalance
				// info.TodayBalance = todayBalance
				pcBalance := int64(0)
				myBalance := int64(0)
				if len(tpblist) > 0 {
					for _, tpitem := range tpblist {
						if paychannel != 0 {
							if tpitem.PayChannel == uint32(paychannel) {
								pcBalance = tpitem.Balance
								myBalance = tpitem.OwnBalance
							}
						} else {
							pcBalance += tpitem.Balance
							myBalance = tpitem.OwnBalance
						}
					}
				}
				info.TodayBalance = myBalance
				info.PayChannelBalance = pcBalance
				list = append(list, *info)
			}
		}

	}
	if len(list) > 0 {
		// 获取支付渠道名称
		channellist, _ := GameService.GetPayChannel()
		for k, v := range list {
			if v.PayChannel != 0 {
				id := strconv.FormatUint(uint64(v.PayChannel), 10)
				for _, chl := range channellist {
					if id == chl.Id {
						v.PayChannelName = chl.Name
						break
					}
				}
			}
			v.FTotalPay = Chip2Float(v.TotalPay)
			v.FTotalPayAmount = Chip2Float(v.TotalPayAmount)
			v.FTotalWithdraw = Chip2Float(v.TotalWithdraw)
			v.FTotalWithdrawAmount = Chip2Float(v.TotalWithdrawAmount)
			v.FPayHandlingFee = Chip2Float(v.PayHandlingFee)
			v.FWithdrawHandlingFee = Chip2Float(v.WithdrawHandlingFee)
			v.TotalProfit = Chip2Float(v.TotalPay - v.TotalWithdraw)
			v.ActualProfit = Chip2Float(v.TotalPayAmount - v.TotalWithdrawAmount)
			if v.TotalProfit > 0 {
				v.IsTotal = true
			}
			if v.ActualProfit > 0 {
				v.IsActual = true
			}
			// v.FYesterdayBalance = Chip2Float(v.YesterdayBalance)
			v.FTodayBalance = Chip2Float(v.TodayBalance)
			v.FPayChannelBalance = Chip2Float(v.PayChannelBalance)
			v.FAmountDifference = Chip2Float(v.TodayBalance - v.PayChannelBalance)
			// 如果是当日获取当前余额
			// now := time.Now()
			// if v.SDate.Year() == now.Year() && v.SDate.Month() == now.Month() && v.SDate.Day() == now.Day() {
			// 	totalBalance := int64(0)
			// 	// 查询当前余额
			// 	startTimestamp := utils.Time2StampToMS(v.SDate)
			// 	pipeline4 := []bson.M{
			// 		{"$match": bson.M{
			// 			"date": bson.M{"$gte": startTimestamp},
			// 		}},
			// 		{"$sort": bson.M{"pay_channel": 1, "date": 1}},
			// 		{"$group": bson.M{
			// 			"_id":         "$pay_channel",
			// 			"firstRecord": bson.M{"$last": "$$ROOT"},
			// 		}},
			// 		{"$replaceRoot": bson.M{
			// 			"newRoot": "$firstRecord",
			// 		}},
			// 	}
			// 	var result4 []entity.PayChannelLog
			// 	err = PayChannelLogs.Pipe(pipeline4).All(&result4)
			// 	for _, rs := range result4 {
			// 		if paychannel != 0 {
			// 			if rs.PayChannel == v.PayChannel {
			// 				totalBalance = rs.Balance
			// 			}
			// 		} else {
			// 			totalBalance += rs.Balance
			// 		}
			// 	}
			// 	v.CurrentBalance = int64(totalBalance)
			// 	v.FCurrentBalance = Chip2Float(int64(totalBalance))
			// }
			list[k] = v
		}
	}
	// 使用 sort.Slice() 对用户信息进行排序
	sort.Slice(list, func(i, j int) bool {
		return list[i].Date > list[j].Date
	})
	return list, err
}

func (this *payService) GetPayChannelStatsNew(startdate, enddate string, paychannel int) ([]entity.PayChannelStatsData, error) {
	list := make([]entity.PayChannelStatsData, 0)
	// 根据日期获取提现订单
	s := fmt.Sprintf("%s 00:00:00", startdate)
	sTime := utils.Str2Time(s, Location())
	e := fmt.Sprintf("%s 23:59:59", enddate)
	eTime := utils.Str2Time(e, Location())
	var err error
	pipeline := []bson.M{}
	if paychannel != 0 {
		pipeline = []bson.M{
			{
				"$match": bson.M{
					"date":        bson.M{"$gte": sTime.Unix(), "$lt": eTime.Unix()},
					"pay_channel": paychannel,
				},
			},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":        "$date",
						"pay_channel": "$pay_channel",
					},
					"total_pay":             bson.M{"$sum": "$total_pay"},
					"pay_rate":              bson.M{"$max": "$pay_rate"},
					"withdraw_rate":         bson.M{"$max": "$withdraw_rate"},
					"total_pay_amount":      bson.M{"$sum": "$total_pay_amount"},
					"total_withdraw":        bson.M{"$sum": "$total_withdraw"},
					"total_withdraw_amount": bson.M{"$sum": "$total_withdraw_amount"},
					"pay_handling_fee":      bson.M{"$sum": "$pay_handling_fee"},
					"withdraw_handling_fee": bson.M{"$sum": "$withdraw_handling_fee"},
				},
			},
			{
				"$project": bson.M{
					"_id":                   "$_id.date",
					"date":                  "$_id.date",
					"pay_channel":           "$_id.pay_channel",
					"total_pay":             "$total_pay",
					"pay_rate":              "$pay_rate",
					"withdraw_rate":         "$withdraw_rate",
					"total_pay_amount":      "$total_pay_amount",
					"total_withdraw":        "$total_withdraw",
					"total_withdraw_amount": "$total_withdraw_amount",
					"pay_handling_fee":      "$pay_handling_fee",
					"withdraw_handling_fee": "$withdraw_handling_fee",
				},
			},
		}
	} else {
		pipeline = []bson.M{
			{
				"$match": bson.M{
					"date": bson.M{"$gte": sTime.Unix(), "$lt": eTime.Unix()},
				},
			},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date": "$date",
					},
					"total_pay":             bson.M{"$sum": "$total_pay"},
					"total_pay_amount":      bson.M{"$sum": "$total_pay_amount"},
					"total_withdraw":        bson.M{"$sum": "$total_withdraw"},
					"total_withdraw_amount": bson.M{"$sum": "$total_withdraw_amount"},
					"pay_handling_fee":      bson.M{"$sum": "$pay_handling_fee"},
					"withdraw_handling_fee": bson.M{"$sum": "$withdraw_handling_fee"},
				},
			},
			{
				"$project": bson.M{
					"_id":                   "$_id.date",
					"date":                  "$_id.date",
					"total_pay":             "$total_pay",
					"total_pay_amount":      "$total_pay_amount",
					"total_withdraw":        "$total_withdraw",
					"total_withdraw_amount": "$total_withdraw_amount",
					"pay_handling_fee":      "$pay_handling_fee",
					"withdraw_handling_fee": "$withdraw_handling_fee",
				},
			},
		}
	}
	err = PayChannelStatss.Pipe(pipeline).All(&list)
	if len(list) > 0 {
		// 查询每个渠道余额
		n := bson.M{}
		n["ctime"] = bson.M{"$gte": sTime.Unix(), "$lte": eTime.Unix()}
		if paychannel != 0 {
			n["pay_channel"] = paychannel
		}
		var tpblist []entity.ThirdPartyBalance
		n2 := []bson.M{
			{"$match": n},
			{"$group": bson.M{
				"_id":         "$ctime",
				"balance":     bson.M{"$sum": "$balance"},
				"own_balance": bson.M{"$sum": "$own_balance"},
			}},
			{
				"$project": bson.M{
					"_id":         "$_id",
					"ctime":       "$_id",
					"balance":     "$balance",
					"own_balance": "$own_balance",
				},
			},
		}
		if paychannel != 0 {
			n2 = []bson.M{
				{"$match": n},
				{"$group": bson.M{
					"_id": bson.M{
						"ctime":       "$ctime",
						"pay_channel": "$pay_channel",
					},
					"balance":     bson.M{"$sum": "$balance"},
					"own_balance": bson.M{"$sum": "$own_balance"},
				}},
				{
					"$project": bson.M{
						"_id":         "$_id.ctime",
						"ctime":       "$_id.ctime",
						"pay_channel": "$_id.pay_channel",
						"balance":     "$balance",
						"own_balance": "$own_balance",
					},
				},
			}
		}
		err = ThirdPartyBalances.Pipe(n2).All(&tpblist)
		channellist, _ := GameService.GetPayChannel()
		for k, v := range list {
			if v.PayChannel != 0 {
				id := strconv.FormatUint(uint64(v.PayChannel), 10)
				for _, chl := range channellist {
					if id == chl.Id {
						v.PayChannelName = chl.Name
						break
					}
				}
			}
			v.SDate = time.Unix(v.Date, 0)
			v.FTotalPay = Chip2Float(v.TotalPay)
			v.FTotalPayAmount = Chip2Float(v.TotalPayAmount)
			v.FTotalWithdraw = Chip2Float(v.TotalWithdraw)
			v.FTotalWithdrawAmount = Chip2Float(v.TotalWithdrawAmount)
			v.FPayHandlingFee = Chip2Float(v.PayHandlingFee)
			v.FWithdrawHandlingFee = Chip2Float(v.WithdrawHandlingFee)
			v.TotalProfit = Chip2Float(v.TotalPay - v.TotalWithdraw)
			v.ActualProfit = Chip2Float(v.TotalPayAmount - v.TotalWithdrawAmount)
			if v.TotalProfit > 0 {
				v.IsTotal = true
			}
			if v.ActualProfit > 0 {
				v.IsActual = true
			}
			// v.FYesterdayBalance = Chip2Float(v.YesterdayBalance)
			v.FTodayBalance = Chip2Float(v.TodayBalance)
			v.FPayChannelBalance = Chip2Float(v.PayChannelBalance)
			v.FAmountDifference = Chip2Float(v.TodayBalance - v.PayChannelBalance)

			pcBalance := int64(0)
			myBalance := int64(0)
			if len(tpblist) > 0 {
				ptime := v.SDate.Format("2006-01-02")
				for _, tpitem := range tpblist {
					stime := time.Unix(tpitem.Ctime, 0).Format("2006-01-02")
					if ptime == stime {
						if paychannel != 0 {
							if tpitem.PayChannel == uint32(paychannel) {
								pcBalance = tpitem.Balance
								myBalance = tpitem.OwnBalance
							}
						} else {
							pcBalance += tpitem.Balance
							myBalance = tpitem.OwnBalance
						}
						break
					}

				}
			}
			v.TodayBalance = myBalance
			v.PayChannelBalance = pcBalance
			v.FTodayBalance = Chip2Float(v.TodayBalance)
			v.FPayChannelBalance = Chip2Float(v.PayChannelBalance)
			v.FAmountDifference = Chip2Float(v.TodayBalance - v.PayChannelBalance)
			list[k] = v
		}
		// 使用 sort.Slice() 对用户信息进行排序
		sort.Slice(list, func(i, j int) bool {
			return list[i].Date > list[j].Date
		})
	}

	return list, err
}

// 支付渠道统计(收付汇总)
func (this *payService) PayChannelCollect(begin, end time.Time, packageIds, payChannelsArg []string) (list []*entity.PayChannelCollect, err error) {
	var payChannels []entity.PayChannel
	err = PayChannels.Find(bson.M{}).All(&payChannels)
	if err != nil {
		return
	}
	payIdChannelMap := utils.Slice2Map(payChannels, func(i int, c entity.PayChannel) string { return c.Id })
	_ = payIdChannelMap

	var where_pay string
	var where_withdraw string
	var args0 = []any{begin, end}
	var allPay = len(payChannelsArg) == 0 // 是否全部渠道
	if len(packageIds) > 0 {
		where_pay += " AND package_id IN ?"
		where_withdraw += " AND package_id IN ?"
		args0 = append(args0, packageIds)
	}
	if !allPay {
		where_pay += " AND channel_id IN ?"
		where_withdraw += " AND out_channel IN ?"
		args0 = append(args0, payChannelsArg)
	}
	// 支付
	var pay_datas []map[string]any
	err = ck.Select(&pay_datas, fmt.Sprintf(`
		SELECT channel_id, toYYYYMMDD(ctime) ctimef, groupArray(DISTINCT package_id) package_ids,
			SUM(amount) amount_sum 
		FROM game.col_trade_record FINAL
		WHERE order_status = 4 AND ctime BETWEEN ? AND ? %s
		GROUP BY channel_id, ctimef
	`, where_pay), args0...)
	if err != nil {
		return
	}
	// 提现
	var withdraw_datas []map[string]any
	err = ck.Select(&withdraw_datas, fmt.Sprintf(`
		SELECT out_channel, toYYYYMMDD(ctime) ctimef, groupArray(DISTINCT package_id) package_ids, 
			SUM(amount) amount_sum, count(*) withdraw_times
		FROM game.col_withdraw_record FINAL
		WHERE order_status = 2 AND ctime BETWEEN ? AND ? %s
		GROUP BY out_channel, ctimef
	`, where_withdraw), args0...)
	if err != nil {
		return
	}

	type pay struct {
		amountPay      int64
		amountWithdraw int64
		withdrawTimes  int64
		packages       []string
	}
	var dateChannelPays = make(map[string]map[string]*pay)
	for _, data := range pay_datas {
		ctimef := fmt.Sprint(data["ctimef"])
		datestr := ctimef[0:4] + "-" + ctimef[4:6] + "-" + ctimef[6:8]
		payChannel := fmt.Sprint(data["channel_id"])
		packages := data["package_ids"].([]string)
		amount_sum := utils.ToInt64(data["amount_sum"])
		pays, ok := dateChannelPays[datestr]
		if !ok {
			pays = make(map[string]*pay)
			dateChannelPays[datestr] = pays
		}
		pays[payChannel] = &pay{amountPay: amount_sum, packages: packages}
	}
	for _, data := range withdraw_datas {
		ctimef := fmt.Sprint(data["ctimef"])
		datestr := ctimef[0:4] + "-" + ctimef[4:6] + "-" + ctimef[6:8]
		payChannel := fmt.Sprint(data["out_channel"])
		packages := data["package_ids"].([]string)
		amount_sum := utils.ToInt64(data["amount_sum"])
		withdraw_times := utils.ToInt64(data["withdraw_times"])

		pays, ok := dateChannelPays[datestr]
		if !ok {
			pays = make(map[string]*pay)
			dateChannelPays[datestr] = pays
		}
		if p, ok := pays[payChannel]; ok {
			p.amountWithdraw = amount_sum
			p.withdrawTimes = withdraw_times
			p.packages = append(p.packages, packages...)
		} else {
			pays[payChannel] = &pay{amountWithdraw: amount_sum, withdrawTimes: withdraw_times, packages: packages}
		}
	}
	mark := "--"
	summary := &entity.PayChannelCollect{
		Date:         fmt.Sprintf("%v-%v总汇", begin.Format(utils.FORMAT_DATE), end.Format(utils.FORMAT_DATE)),
		ChannelClass: mark,
		Channel1:     mark,
		PayChannel:   mark,
		PackageMap:   make(map[string]bool),
	}
	for !begin.After(end) {
		stime := begin
		begin = begin.AddDate(0, 0, 1)
		sdate := stime.Format(utils.FORMAT_DATE)
		pays, ok := dateChannelPays[sdate]
		if !ok {
			continue
		}
		if allPay {
			// 全部渠道汇总到一天
			item := &entity.PayChannelCollect{
				Date:       sdate,
				PayChannel: "全部",
				PackageMap: make(map[string]bool),
			}
			list = append(list, item)
			for channel, p := range pays {
				item.PayAmount0 += p.amountPay
				item.WithdrawAmount0 += p.amountWithdraw
				// 计算税
				if payChannel, ok := payIdChannelMap[channel]; ok {
					item.PayAmountTax0 += float64(p.amountPay) * (payChannel.PayRate / 100)
					item.WithdrawAmountTaxRate0 += float64(p.amountWithdraw) * (payChannel.WithdrawRate / 100)
					item.WithdrawAmountTaxTimes0 += p.withdrawTimes * payChannel.WithdrawFee
				}
				// 渠道渠道类
				for _, pkg := range p.packages {
					item.PackageMap[pkg] = true
				}
			}
			// 渠道余额
			for _, payChannel := range payIdChannelMap {
				item.PayChannelBalance0 += int64(payChannel.Balance)
			}
			this.PayChannelCollectItemCalc(item, summary)
		} else {
			// 每个渠道单独计算
			for channel, p := range pays {
				item := &entity.PayChannelCollect{
					Date:       sdate,
					PayChannel: channel,
					PackageMap: make(map[string]bool),
				}
				list = append(list, item)

				item.PayAmount0 += p.amountPay
				item.WithdrawAmount0 += p.amountWithdraw
				// 计算税
				if payChannel, ok := payIdChannelMap[channel]; ok {
					item.PayChannel = payChannel.Name
					item.PayChannelBalance0 += int64(payChannel.Balance)
					item.PayAmountTax0 += float64(p.amountPay) * (payChannel.PayRate / 100)
					item.WithdrawAmountTaxRate0 += float64(p.amountWithdraw) * (payChannel.WithdrawRate / 100)
					item.WithdrawAmountTaxTimes0 += p.withdrawTimes * payChannel.WithdrawFee
				}
				// 渠道渠道类
				for _, pkg := range p.packages {
					item.PackageMap[pkg] = true
				}
				this.PayChannelCollectItemCalc(item, summary)
			}
		}
	}
	// 汇总渠道余额
	if allPay {
		for _, payChannel := range payIdChannelMap {
			summary.PayChannelBalance0 += int64(payChannel.Balance)
		}
	} else {
		// 渠道余额
		for _, channel := range payChannelsArg {
			if payChannel, ok := payIdChannelMap[channel]; ok {
				summary.PayChannelBalance0 += int64(payChannel.Balance)
			}
		}
	}
	channelMap, err := ChannelService.GetChannelsMap()
	if err != nil {
		return
	}
	// 渠道类渠道别名
	for _, item := range list {
		if len(item.PackageMap) == 0 {
			continue
		}
		var channels []string
		for pkg := range item.PackageMap {
			channels = append(channels, pkg)
		}
		sort.Strings(channels)
		var aliasMap = make(map[string]bool)
		var classMap = make(map[string]bool)
		for _, c := range channels {
			if ch, ok := channelMap[c]; ok {
				if ch.Name1 != "" {
					aliasMap[ch.Name1] = true
				}
				if ch.ClassName != "" {
					classMap[ch.ClassName] = true
				}
			}
		}
		channels = make([]string, 0)
		for alias := range aliasMap {
			channels = append(channels, alias)
		}
		sort.Strings(channels)
		item.Channel1 = strings.Join(channels, ",")

		channels = make([]string, 0)
		for cls := range classMap {
			channels = append(channels, cls)
		}
		sort.Strings(channels)
		item.ChannelClass = strings.Join(channels, ",")
	}

	this.PayChannelCollectItemCalc(summary, nil)
	list = append(list, summary)
	count := len(list)
	// 表格：倒序显示
	for i := 0; i < count/2; i++ {
		list[i], list[count-1-i] = list[count-1-i], list[i]
	}
	return
}

// PayChannelCollectItemCalc 收付汇总结果计算
func (this *payService) PayChannelCollectItemCalc(item, summary *entity.PayChannelCollect) {
	item.PayAmount = fmt.Sprintf("%.2f", Chip2Float(item.PayAmount0))
	item.PayAmountTax = fmt.Sprintf("%.2f", Chip2Float(item.PayAmountTax0))
	item.WithdrawAmount = fmt.Sprintf("%.2f", Chip2Float(item.WithdrawAmount0))
	item.WithdrawAmountTaxRate = fmt.Sprintf("%.2f", Chip2Float(item.WithdrawAmountTaxRate0))
	item.WithdrawAmountTaxTimes = fmt.Sprintf("%.2f", Chip2Float(item.WithdrawAmountTaxTimes0))
	item.PayChannelBalance = fmt.Sprintf("%.2f", Chip2Float(item.PayChannelBalance0))
	earning := float64(item.PayAmount0) - item.PayAmountTax0
	expense := float64(item.WithdrawAmount0) + item.WithdrawAmountTaxRate0 + float64(item.WithdrawAmountTaxTimes0)
	item.Earning = fmt.Sprintf("%.2f", Chip2Float(earning))
	item.Expense = fmt.Sprintf("%.2f", Chip2Float(expense))
	item.Profit = fmt.Sprintf("%.2f", Chip2Float(earning-expense))

	if summary != nil {
		summary.PayAmount0 += item.PayAmount0
		summary.PayAmountTax0 += item.PayAmountTax0
		summary.WithdrawAmount0 += item.WithdrawAmount0
		summary.WithdrawAmountTaxRate0 += item.WithdrawAmountTaxRate0
		summary.WithdrawAmountTaxTimes0 += item.WithdrawAmountTaxTimes0
	}
}

func ConvertToInt64(b interface{}) int64 {
	num := int64(0)
	// if b
	num, dsAmountIsInt64 := b.(int64)
	if !dsAmountIsInt64 {
		// 进行类型转换
		dsAmountInt, dsAmountIsInt := b.(int)
		if !dsAmountIsInt {
			// 处理无法转换为int64的情况
			fmt.Println("ConvertToInt64 is not int64 or int")
		}
		num = int64(dsAmountInt)
	}
	return num
}

// 转换为分展示
func (this *payService) chipList(list []entity.PayChannelStatsData) []entity.PayChannelStatsData {
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		v.FTotalPay = Chip2Float(v.TotalPay)
		v.FTotalPayAmount = Chip2Float(v.TotalPayAmount)
		v.FTotalWithdraw = Chip2Float(v.TotalWithdraw)
		v.FTotalWithdrawAmount = Chip2Float(v.TotalWithdrawAmount)
		v.FPayHandlingFee = Chip2Float(v.PayHandlingFee)
		v.FWithdrawHandlingFee = Chip2Float(v.WithdrawHandlingFee)
		v.FYesterdayBalance = Chip2Float(v.YesterdayBalance)
		v.FTodayBalance = Chip2Float(v.TodayBalance)
		v.FCurrentBalance = Chip2Float(v.CurrentBalance)
		v.FPayChannelBalance = Chip2Float(v.PayChannelBalance)
		diffAmount := v.Balance - v.PayChannelBalance
		v.FAmountDifference = Chip2Float(diffAmount)
		actualprofit := v.TotalPayAmount - v.TotalWithdrawAmount
		v.ActualProfit = Chip2Float(actualprofit)
		totalprofit := v.TotalPay - v.TotalWithdraw
		v.TotalProfit = Chip2Float(totalprofit)
		list[k] = v
	}
	return list
}

// 查询支付渠道统计条数
func (this *payService) GetPayChannelStatTotal(m bson.M, isChannel bool) (int64, error) {
	pipeline := []bson.M{}
	if isChannel {
		pipeline = []bson.M{
			{
				"$match": m,
			},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":        "$date",
						"pay_channel": "$pay_channel",
					},
				},
			},
		}
	} else {
		pipeline = []bson.M{
			{
				"$match": m,
			},
			{
				"$group": bson.M{
					"_id": "$date",
				},
			},
		}
	}

	// operations := []bson.M{m, n}
	result := []bson.M{}
	pipe := PayChannelStatss.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("GetPayChannelStatTotal fail err: ", err)
	}
	return int64(len(result)), nil
}

/*
支付渠道日志
*/
func (this *payService) GetPayChannelLogList(page, pageSize int, m bson.M) ([]entity.PayChannelLog, error) {
	var list []entity.PayChannelLog
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "date", false)
	PayChannelLogs.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList27(list)
	return list, nil
}

func (this *payService) chipList27(list []entity.PayChannelLog) []entity.PayChannelLog {
	paylist, _ := GameService.GetPayChannel()

	for k, v := range list {
		s, _ := ConvertToIndiaTime1(v.Date)
		v.SDate = s
		v.FAmount = Chip2Float(v.Amount)
		v.FActualAmount = Chip2Float(v.ActualAmount)
		v.FHandlingCharge = Chip2Float(v.HandlingCharge)
		v.FBalance = Chip2Float(v.Balance)
		str := strconv.FormatUint(uint64(v.PayChannel), 10)
		for _, p := range paylist {
			if p.Id == str {
				v.PayChannelName = p.Name
			}
		}
		list[k] = v
	}
	return list
}

// 支付渠道日志条数
func (this *payService) GetPayChannelLogTotal(m bson.M) (int64, error) {
	return int64(Count(PayChannelLogs, m)), nil
}

// 支付对账备注列表
func (this *payService) GetPayChannelStatRecordsList(page, pageSize int, m bson.M) ([]entity.PayChannelStatRecords, error) {
	var list []entity.PayChannelStatRecords
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "date", false)
	PayChannelStatRecords.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList28(list)
	return list, nil
}

func (this *payService) chipList28(list []entity.PayChannelStatRecords) []entity.PayChannelStatRecords {
	paylist, _ := GameService.GetPayChannel()
	for k, v := range list {
		s, _ := ConvertToIndiaTime(v.Date)
		v.SDate = s
		c, _ := ConvertToIndiaTime(v.Ctime)
		v.CDate = c
		str := strconv.FormatUint(uint64(v.PayChannel), 10)
		strName := "全部"
		for _, p := range paylist {
			if p.Id == str {
				strName = p.Name
				break
			}
		}
		v.PayChannelName = strName
		list[k] = v
	}
	return list
}

// 支付对账备注
func (this *payService) AddPayChannelStatRecords(rate *entity.PayChannelStatRecords) error {
	if !Insert(PayChannelStatRecords, rate) {
		return errors.New("写入失败:" + rate.Id)
	}
	return nil
}

// 查询人机审核配置
func (this *payService) GetWithdrawSettings() (settings []*entity.WithdrawSetting) {
	ListByQ(WithdrawSetting, bson.M{}, &settings)
	var ctypes = [7]bool{}
	for _, item := range settings {
		ctype, err := strconv.Atoi(item.Id)
		if err == nil && ctype >= 0 && ctype < 7 {
			ctypes[ctype] = true
		}
	}
	for ctype, ok := range ctypes {
		if !ok {
			setting := &entity.WithdrawSetting{}
			setting.Id = strconv.Itoa(ctype)
			setting.JQWithdrawMax = 5000
			setting.JQWinRateMax = 2.0
			setting.JQWinMax = 10000
			setting.IsBankRepeatable = false
			Upsert(WithdrawSetting, bson.M{"_id": setting.Id}, setting)
			settings = append(settings, setting)
		}
	}

	return
}

// 查询人机审核配置
func (this *payService) GetWithdrawSettingById(id string) (setting *entity.WithdrawSetting) {
	setting = new(entity.WithdrawSetting)
	Get(WithdrawSetting, id, setting)
	if setting.Id != id {
		setting.Id = id
		setting.JQWithdrawMax = 5000
		setting.JQWinRateMax = 2.0
		setting.JQWinMax = 10000
		setting.IsBankRepeatable = false
		Upsert(WithdrawSetting, bson.M{"_id": setting.Id}, setting)
	}
	return
}

func (this *payService) UpdateWithdrawSetting(setting *entity.WithdrawSetting) {
	Update(WithdrawSetting, bson.M{"_id": setting.Id}, setting)
}

func (this *payService) GetPayChannelStatNoticeConfig() (c *entity.PayChannelStatNoticeConfig) {
	c = &entity.PayChannelStatNoticeConfig{}
	GetByQ(PayChannelStatNoticeConfigs, bson.M{"_id": "1"}, c)
	if c.Id == "" {
		c.Id = "1"
		c.DayTicker = 10
		c.TimeTicker = 20
		c.WithdrawTicker = 20
		c.PayUtrTicker = 5
		Upsert(PayChannelStatNoticeConfigs, bson.M{"_id": c.Id}, c)
	}
	return
}

func (this *payService) UpdatePayChannelStatNoticeConfig(dayTicker, timeTicker, withdrawTicker, payUtrTicker int64) {
	m := bson.M{"$set": bson.M{
		"day_ticker":      dayTicker,
		"time_ticker":     timeTicker,
		"withdraw_ticker": withdrawTicker,
		"pay_utr_ticker":  payUtrTicker,
	}}
	Update(PayChannelStatNoticeConfigs, bson.M{"_id": "1"}, m)
}

func (this *payService) GetPayChannelStat() (
	timeBegin, timeEnd, prevTimeBegin, prevTimeEnd, prevDayTimeBegin, prevDayTimeEnd time.Time,
	dayBegin, dayEnd, prevDayBegin, prevDayEnd time.Time,
	withdrawTimeStats, payTimeStats, prevWithdrawTimeStats, prevPayTimeStats, prevDayWithdrawTimeStats, prevDayPayTimeStats []*entity.PayChannelStat,
	withdrawDayStats, payDayStats, prevDayWithdrawStats, prevDayPayStats []*entity.PayChannelStat,
	err error,
) {
	// 统计当天和时间段的数据
	cfg := this.GetPayChannelStatNoticeConfig()

	// now := time.Date(2024, 8, 30, 15, 0, 0, 0, timeLocationSetting)
	now := NowTime()

	// ---时段统计
	timeRange := -(time.Minute * time.Duration(cfg.TimeTicker))
	timeBegin = now.Add(timeRange)
	timeEnd = now
	withdrawTimeStats, payTimeStats, err = this.payChannelStat(timeBegin, timeEnd)
	if err != nil {
		return
	}

	// 环比上一时段
	prevTimeBegin = timeBegin.Add(timeRange)
	prevTimeEnd = timeEnd.Add(timeRange)
	prevWithdrawTimeStats, prevPayTimeStats, err = this.payChannelStat(prevTimeBegin, prevTimeEnd)
	if err != nil {
		return
	}

	// 同比昨日同一时段
	prevDayTimeBegin = timeBegin.AddDate(0, 0, -1)
	prevDayTimeEnd = timeEnd.AddDate(0, 0, -1)
	prevDayWithdrawTimeStats, prevDayPayTimeStats, err = this.payChannelStat(prevDayTimeBegin, prevDayTimeEnd)
	if err != nil {
		return
	}

	// 当天统计
	dayBegin = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	dayEnd = now
	withdrawDayStats, payDayStats, err = this.payChannelStat(dayBegin, dayEnd)
	if err != nil {
		return
	}

	// 同比昨日同一时段
	prevDayBegin = dayBegin.AddDate(0, 0, -1)
	prevDayEnd = dayEnd.AddDate(0, 0, -1)
	prevDayWithdrawStats, prevDayPayStats, err = this.payChannelStat(prevDayBegin, prevDayEnd)
	if err != nil {
		return
	}
	return
}

func (this *payService) payChannelStat(begin, end time.Time) (withdrawStats []*entity.PayChannelStat, payStats []*entity.PayChannelStat, err error) {
	sqlWithdraw := `
	SELECT channel_id, status, count(*) total FROM (
		SELECT channel_id, 
			(CASE WHEN status = 2 THEN 1 WHEN status IN (5,7) THEN 2 ELSE 0 END) status
		FROM game.col_withdraw_record_transfer_order_detail FINAL WHERE ctime BETWEEN ? AND ?
	) s1
	GROUP BY channel_id, status
`
	sqlPay := `
	SELECT channel_id, status, count(*) total FROM (
		SELECT channel_id, (CASE WHEN order_status = 4 THEN 1 WHEN order_status IN (2, 5) THEN 2 ELSE 0 END) status
		FROM game.col_trade_record FINAL WHERE ctime BETWEEN ? and ? AND channel_id != 0 AND shop_type != 99
	) s1
	GROUP BY channel_id, status
`
	var datas_withdraw []map[string]any
	err = ck.Select(&datas_withdraw, sqlWithdraw, begin.Unix(), end.Unix())
	if err != nil {
		return
	}
	var datas_pay []map[string]any
	err = ck.Select(&datas_pay, sqlPay, begin, end)
	if err != nil {
		return
	}

	// 提现统计
	var payChannelIds []string
	var withdrawChannelStatsMap = make(map[string]*entity.PayChannelStat)
	var payChannelStatsMap = make(map[string]*entity.PayChannelStat)
	for _, data := range datas_withdraw {
		channel_id := fmt.Sprint(data["channel_id"])
		status := utils.ToInt64(data["status"])
		total := utils.ToInt64(data["total"])
		stat, ok := withdrawChannelStatsMap[channel_id]
		if !ok {
			stat = &entity.PayChannelStat{
				Channel: channel_id,
			}
			withdrawChannelStatsMap[channel_id] = stat
			withdrawStats = append(withdrawStats, stat)
			payChannelIds = append(payChannelIds, channel_id)
		}
		stat.TotalTimes += total
		stat.WithdrawTotalTimes += total
		switch status {
		case 1: // 成功
			stat.SuccessTimes += total
			stat.WithdrawSuccessTimes += total
		case 2: // 失败
			stat.ErrorTimes += total
			stat.WithdrawErrorTimes += total
		}
	}
	// 充值统计
	for _, data := range datas_pay {
		channel_id := fmt.Sprint(data["channel_id"])
		status := utils.ToInt64(data["status"])
		total := utils.ToInt64(data["total"])
		stat, ok := payChannelStatsMap[channel_id]
		if !ok {
			stat = &entity.PayChannelStat{
				Channel: channel_id,
			}
			payChannelStatsMap[channel_id] = stat
			payStats = append(payStats, stat)
			payChannelIds = append(payChannelIds, channel_id)
		}
		stat.TotalTimes += total
		stat.PayTotalTimes += total
		switch status {
		case 1: // 成功
			stat.SuccessTimes += total
			stat.PaySuccessTimes += total
		case 2: // 失败
			stat.ErrorTimes += total
			stat.PayErrorTimes += total
		}
	}

	if len(payChannelIds) == 0 {
		return
	}
	// 支付渠道查询
	var payChannels []entity.PayChannel
	ListByQ(PayChannels, bson.M{"_id": bson.M{"$in": payChannelIds}}, &payChannels)
	var payChannelMap = make(map[string]entity.PayChannel, len(payChannels))
	for _, c := range payChannels {
		payChannelMap[c.Id] = c
	}

	var stats []*entity.PayChannelStat
	stats = append(stats, withdrawStats...)
	stats = append(stats, payStats...)
	for _, stat := range stats {
		c, ok := payChannelMap[stat.Channel]
		if ok {
			stat.ChannelName = c.Name
		} else {
			stat.ChannelName = stat.Channel
		}
		if stat.TotalTimes > 0 {
			stat.SuccessRate = fmt.Sprintf("%.2f%%", float64(stat.SuccessTimes)/float64(stat.TotalTimes)*100)
		}
		if stat.PayTotalTimes > 0 {
			stat.PaySuccessRate = fmt.Sprintf("%.2f%%", float64(stat.PaySuccessTimes)/float64(stat.PayTotalTimes)*100)
		}
		if stat.WithdrawTotalTimes > 0 {
			stat.WithdrawSuccessRate = fmt.Sprintf("%.2f%%", float64(stat.WithdrawSuccessTimes)/float64(stat.WithdrawTotalTimes)*100)
		}
	}

	sort.Slice(withdrawStats, func(i, j int) bool { return withdrawStats[i].WithdrawTotalTimes > withdrawStats[j].WithdrawTotalTimes })
	sort.Slice(payStats, func(i, j int) bool { return payStats[i].PayTotalTimes > payStats[j].PayTotalTimes })
	return
}

// 提现标记订单统计
func (this *payService) WithdrawTaggedStat() (count int64, amountSum float64, err error) {
	list, err := GameService.GetGameSystem(4)
	if err != nil {
		return
	}
	isWithdrawAutoAudit := false

	if len(list) > 0 {
		for _, item := range list {
			if item.Name == "提现自动审核" && (item.Status == 1 || item.Status == 2) {
				isWithdrawAutoAudit = true
			}
		}
	}

	// startTime := time.Date(2024, 9, 1, 0, 0, 0, 0, timeLocationSetting)
	startTime := time.Date(2024, 11, 4, 0, 0, 0, 0, location)
	m := bson.M{}
	m["ctime"] = bson.M{"$gte": startTime} // , "$lt": endTime
	m["order_status"] = 1
	if isWithdrawAutoAudit {
		// 开启自动审核，只提示打标的数据
		m["is_tag"] = bson.M{"$eq": true}
	}
	withdraws, err := this.GetByWithdrawUser(m)
	if err != nil {
		return
	}
	if len(withdraws) == 0 {
		return
	}
	// 过滤黑名单用户
	var userids []string
	for _, w := range withdraws {
		userids = append(userids, w.Userid)
	}
	var blackUsers []map[string]any
	uPipe := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": userids}, "status": bson.M{"$eq": 3}}},
		{"$project": bson.M{"_id": 1}},
	}
	err = PlayerUsers.Pipe(uPipe).All(&blackUsers)
	if err != nil {
		return
	}
	var blackUMap = make(map[string]bool)
	for _, u := range blackUsers {
		blackUMap[fmt.Sprint(u["_id"])] = true
	}

	var amounts uint32
	for _, w := range withdraws {
		// 黑名单
		if blackUMap[w.Userid] {
			continue
		}
		count++
		amounts += (w.Commission + w.Amount)
	}
	amountSum = Chip2Float(amounts)
	return
}

// 支付订单UTR补分通知统计
func (this *payService) PayUtrStat() (count int64, amount_sum float64, err error) {
	m := bson.M{
		"order_status": data.Tradeing,
		"ctime":        bson.M{"$gt": time.Date(2025, 3, 5, 0, 0, 0, 0, location)},
		"last_utr":     bson.M{"$exists": true, "$ne": ""},
		"$or": []bson.M{
			{"error_msg": bson.M{"$exists": false}},
			{"error_msg": bson.M{"$eq": ""}},
		},
	}

	pipe := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":        nil,
			"count":      bson.M{"$sum": 1},
			"amount_sum": bson.M{"$sum": "$amount"},
		}},
	}
	stats := []bson.M{}
	err = Pays.Pipe(pipe).All(&stats)
	if err != nil {
		return
	}
	if len(stats) == 0 {
		return
	}
	count = utils.ToInt64(stats[0]["count"])
	amount_sum = Chip2Float(utils.ToInt64(stats[0]["amount_sum"]))
	return
}
