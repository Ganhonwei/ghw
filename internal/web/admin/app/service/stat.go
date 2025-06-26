package service

import (
	"fmt"
	"math"
	"time"

	"goserver/gen/pb"
	"goserver/internal/web/admin/app/entity"
	"goserver/pkg/utils"

	"github.com/astaxie/beego"
	"github.com/globalsign/mgo/bson"
)

// IP检测，从11月1号-执行时间，只需要执行一次
func IPTesting(timestamp int64) {
	beego.Info("开始执行IPTesting函数！")
	tim := utils.Stamp2Time(timestamp)
	// tim = tim.AddDate(0, 0, -4)
	endTime := tim //.Format("2006-01-02 15:04:05")
	loc, err := time.LoadLocation(locationName)
	if err != nil {
		panic(err)
	}
	startTime := time.Date(2023, time.November, 1, 0, 0, 0, 0, loc)
	m := bson.M{}
	m["login_time"] = bson.M{"$gte": startTime, "$lt": endTime}
	var userlist []entity.UserIpRecords // IP检测记录
	var userips []string                // 去重后的ip
	// 声明一个 map 变量用于去重
	// usermap := make(map[string]bool)
	LoginLogs.Find(m).Distinct("ip", &userips)
	if len(userips) > 0 {
		pipeline := []bson.M{
			{
				"$match": bson.M{
					"ip": bson.M{"$in": userips},
				},
			},
			{
				"$group": bson.M{
					"_id":   "$ip",
					"users": bson.M{"$addToSet": "$userid"},
				},
			},
		}
		result := []bson.M{}
		pipe := LoginLogs.Pipe(pipeline)
		err := pipe.All(&result)
		if err != nil {
			beego.Error("IPTesting fail err: ", err)
		}
		if len(result) > 0 {
			for _, item := range result {
				info := new(entity.UserIpRecords)
				// info.Userid = item
				ip := item["_id"].(string)
				arrUser := item["users"].([]interface{})
				users := make([]string, 0)
				for _, v := range arrUser {
					users = append(users, v.(string))
				}
				// users :=
				info.Ip = ip
				info.Userid = users
				userlist = append(userlist, *info)
			}
		}
	}
	if len(userlist) > 0 {
		for _, info := range userlist {
			AddErr := LoggerService.AddUserIpRecod(&info)
			if AddErr != nil {
				beego.Error("IPTesting fail err: ", AddErr.Error())
			}
		}
	}
	beego.Info("IPTesting执行函数完成！")
}

func checkSMS() {
	tim := utils.Stamp2Time(utils.Timestamp())
	today := tim.Format("2006-01-02")
	date1, _ := utils.Unix(fmt.Sprintf("%s", today))
	m := bson.M{}
	m["date"] = date1
	info, _ := StatisticsService.GetSMSByDate(m)
	if info != nil {
		if info.Date != 0 {
			sendRatio := float64(0.00)
			sendRatio = ComputeFloat(info.SendCount, info.TotalSendSMS) * 100
			if sendRatio < 95.00 {
				// 预警
				sendMsg := &pb.AlertorMail{Subject: "Web System Msg", Message: "短信成功率低于95%"}
				SendMail(sendMsg)
				beego.Info("SMSData send time: ", tim)
			}
		}
	}
}

// 汇总充值提现信息
func ComputePayAndWithdraw() {
	timestamp := utils.TimestampYesterday(location)
	tim := utils.Stamp2Time(timestamp)
	today := tim.Format("2006-01-02")
	start_today := tim.AddDate(0, 0, -1).Format("2006-01-02")
	s := fmt.Sprintf("%s 00:00:00", start_today)
	startTime := utils.Str2Time(s, Location())
	e := fmt.Sprintf("%s 23:59:59", today)
	endTime := utils.Str2Time(e, Location())

	query := bson.M{}
	query["order_status"] = 4
	query["ctime"] = bson.M{"$gte": startTime, "$lt": endTime}
	var payids []string
	Pays.Find(query).Distinct("userid", &payids)

	list := make(map[string]entity.UserCashRecod, 0)
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"order_status": 4,
				"userid":       bson.M{"$in": payids},
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Amount": bson.M{
					"$sum": "$amount",
				},
				"Count": bson.M{
					"$sum": 1,
				},
				"FirstRechargeTime": bson.M{
					"$min": "$ctime",
				},
				"FirstRechargeAmount": bson.M{
					"$first": "$amount",
				},
			},
		},
	}
	result := []bson.M{}
	pipe := Pays.Pipe(pipeline)
	pipe.All(&result)
	for _, pay := range result {
		uid := pay["_id"].(string)
		amount := pay["Amount"].(int)
		count := pay["Count"].(int)
		f_time := pay["FirstRechargeTime"].(time.Time)
		f_amount := pay["FirstRechargeAmount"].(int)
		temp_info := new(entity.UserCashRecod)
		temp_info.Userid = uid
		temp_info.PayAmount = int64(amount)
		temp_info.PayCount = int64(count)
		temp_info.FirstPayDate = f_time
		temp_info.FirstPayAmount = int64(f_amount)
		list[uid] = *temp_info
	}
	query1 := bson.M{}
	query1["order_status"] = 2
	query1["ctime"] = bson.M{"$gte": startTime, "$lt": endTime}
	var withids []string
	Withdraws.Find(query1).Distinct("userid", &withids)
	// 提现
	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"order_status": 2,
				"userid":       bson.M{"$in": withids},
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Amount": bson.M{
					"$sum": "$amount",
				},
				"Count": bson.M{
					"$sum": 1,
				},
				"UsedCards": bson.M{
					"$addToSet": "$blank_number",
				},
			},
		},
	}
	result1 := []bson.M{}
	pipe1 := Withdraws.Pipe(pipeline1)
	pipe1.All(&result1)
	for _, with := range result1 {
		uid := with["_id"].(string)
		amount := with["Amount"].(int)
		count := with["Count"].(int)
		arrCards := with["UsedCards"].([]interface{})
		cards := make([]string, 0)
		for _, v := range arrCards {
			cards = append(cards, v.(string))
		}

		temp_info := new(entity.UserCashRecod)
		temp_info.Userid = uid
		if record, ok := list[uid]; ok {
			temp_info = &record
		}
		temp_info.WithdrawAmount = int64(amount)
		temp_info.WithdrawCount = int64(count)
		temp_info.CardNumber = cards
		list[uid] = *temp_info
	}
	var ids []string
	for _, v := range list {
		ids = append(ids, v.Userid)
	}
	pipeline2 := []bson.M{
		{
			"$match": bson.M{
				"_id": bson.M{"$in": ids},
			},
		},
		{
			"$group": bson.M{
				"_id": "$_id",
				"Diamond": bson.M{
					"$sum": "$diamond",
				},
				"ShadowDiamond": bson.M{
					"$sum": "$shadow_diamond",
				},
			},
		},
	}
	result2 := []bson.M{}
	pipe2 := PlayerUsers.Pipe(pipeline2)
	pipe2.All(&result2)
	userArr := make(map[string]int64)
	for _, user := range result2 {
		id := user["_id"].(string)
		diamond := user["Diamond"].(int64)
		sdiamond := user["ShadowDiamond"].(int64)
		userArr[id] = diamond + sdiamond
	}
	for _, item := range list {
		// 获取账上余额
		balance := int64(0)
		value, exists := userArr[item.Userid]
		if exists {
			balance = int64(value)
		}
		item.GainAmout = (item.WithdrawAmount + balance) - item.PayAmount
		addErr := PayService.AddOrUpdateUserCashRecod(&item)
		if addErr != nil {
			beego.Error("BattleRoomData fail err: ", addErr)
		}
	}
}

// 重新计算所有的充提
func ComputePayAndWithdrawByAll() {
	list := make(map[string]entity.UserCashRecod, 0)
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"order_status": 4,
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Amount": bson.M{
					"$sum": "$amount",
				},
				"Count": bson.M{
					"$sum": 1,
				},
				"FirstRechargeTime": bson.M{
					"$min": "$ctime",
				},
				"FirstRechargeAmount": bson.M{
					"$first": "$amount",
				},
			},
		},
	}
	result := []bson.M{}
	pipe := Pays.Pipe(pipeline)
	pipe.All(&result)
	for _, pay := range result {
		uid := pay["_id"].(string)
		amount := pay["Amount"].(int)
		count := pay["Count"].(int)
		f_time := pay["FirstRechargeTime"].(time.Time)
		f_amount := pay["FirstRechargeAmount"].(int)
		temp_info := new(entity.UserCashRecod)
		temp_info.Userid = uid
		temp_info.PayAmount = int64(amount)
		temp_info.PayCount = int64(count)
		temp_info.FirstPayDate = f_time
		temp_info.FirstPayAmount = int64(f_amount)
		list[uid] = *temp_info
	}
	// 提现
	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"order_status": 2,
			},
		},
		{
			"$group": bson.M{
				"_id": "$userid",
				"Amount": bson.M{
					"$sum": "$amount",
				},
				"Count": bson.M{
					"$sum": 1,
				},
				"UsedCards": bson.M{
					"$addToSet": "$blank_number",
				},
			},
		},
	}
	result1 := []bson.M{}
	pipe1 := Withdraws.Pipe(pipeline1)
	pipe1.All(&result1)
	for _, with := range result1 {
		uid := with["_id"].(string)
		amount := with["Amount"].(int)
		count := with["Count"].(int)
		arrCards := with["UsedCards"].([]interface{})
		cards := make([]string, 0)
		for _, v := range arrCards {
			cards = append(cards, v.(string))
		}

		temp_info := new(entity.UserCashRecod)
		temp_info.Userid = uid
		if record, ok := list[uid]; ok {
			temp_info = &record
		}
		temp_info.WithdrawAmount = int64(amount)
		temp_info.WithdrawCount = int64(count)
		temp_info.CardNumber = cards
		list[uid] = *temp_info
	}
	var ids []string
	for _, v := range list {
		ids = append(ids, v.Userid)
	}
	pipeline2 := []bson.M{
		{
			"$match": bson.M{
				"_id": bson.M{"$in": ids},
			},
		},
		{
			"$group": bson.M{
				"_id": "$_id",
				"Diamond": bson.M{
					"$sum": "$diamond",
				},
				"ShadowDiamond": bson.M{
					"$sum": "$shadow_diamond",
				},
			},
		},
	}
	result2 := []bson.M{}
	pipe2 := PlayerUsers.Pipe(pipeline2)
	pipe2.All(&result2)
	userArr := make(map[string]int64)
	for _, user := range result2 {
		id := user["_id"].(string)
		diamond := user["Diamond"].(int64)
		sdiamond := user["ShadowDiamond"].(int64)
		userArr[id] = diamond + sdiamond
	}
	for _, item := range list {
		// 获取账上余额
		balance := int64(0)
		value, exists := userArr[item.Userid]
		if exists {
			balance = int64(value)
		}
		item.GainAmout = (item.WithdrawAmount + balance) - item.PayAmount
		addErr := PayService.AddOrUpdateUserCashRecod(&item)
		if addErr != nil {
			beego.Error("BattleRoomData fail err: ", addErr)
		}
	}
}

// 计算值
func ComputeFloat[T int | int64 | int32 | uint32 | float64](val1, val2 T) float64 {
	if val2 != 0 {
		result := float64(val1) / float64(val2)
		if math.IsNaN(result) {
			result = 0
		}
		return result
	}
	return 0
}
