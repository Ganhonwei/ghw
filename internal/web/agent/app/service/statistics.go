package service

import (
	"errors"
	"goserver/internal/web/agent/app/entity"
	"time"

	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"
)

type statisticsService struct{}

/*
数据汇总（新）
*/
func (this *statisticsService) GetDataList(page, pageSize int, m bson.M, isChannel bool) ([]entity.DataStatistics, error) {
	var list []entity.DataStatistics
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)
	pipeline := []bson.M{}
	if isChannel {
		// 按渠道
		pipeline = []bson.M{
			{"$match": m},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":       "$date",
						"channel":    "$channel",
						"channel1":   "$channel1",
						"blogger_id": "$blogger_id",
					},
					"new_register":           bson.M{"$sum": "$new_register"},
					"new_equipment":          bson.M{"$sum": "$new_equipment"},
					"login_num":              bson.M{"$sum": "$login_num"},
					"next_new_register":      bson.M{"$sum": "$next_new_register"},
					"next_login":             bson.M{"$sum": "$next_login"},
					"jr_login_next_pay":      bson.M{"$sum": "$jr_login_next_pay"},
					"zr_pay_count":           bson.M{"$sum": "$zr_pay_count"},
					"new_recharge_num":       bson.M{"$sum": "$new_recharge_num"},
					"new_recharge_amount":    bson.M{"$sum": "$new_recharge_amount"},
					"total_recharge":         bson.M{"$sum": "$total_recharge"},
					"total_amount":           bson.M{"$sum": "$total_amount"},
					"pay_request":            bson.M{"$sum": "$pay_request"},
					"pay_request_order":      bson.M{"$sum": "$pay_request_order"},
					"pay_success_order":      bson.M{"$sum": "$pay_success_order"},
					"withdraw_request":       bson.M{"$sum": "$withdraw_request"},
					"withdraw_request_order": bson.M{"$sum": "$withdraw_request_order"},
					"withdraw_success_order": bson.M{"$sum": "$withdraw_success_order"},
					"total_withdraw_num":     bson.M{"$sum": "$total_withdraw_num"},
					"total_withdraw_amount":  bson.M{"$sum": "$total_withdraw_amount"},
					"handling_charge":        bson.M{"$sum": "$handling_charge"},
				},
			},
			{
				"$project": bson.M{
					"_id":                    "$_id.date",
					"date":                   "$_id.date",
					"channel":                "$_id.channel",
					"channel1":               "$_id.channel1",
					"blogger_id":             "$_id.blogger_id",
					"new_register":           "$new_register",
					"new_equipment":          "$new_equipment",
					"login_num":              "$login_num",
					"next_new_register":      "$next_new_register",
					"next_login":             "$next_login",
					"jr_login_next_pay":      "$jr_login_next_pay",
					"zr_pay_count":           "$zr_pay_count",
					"new_recharge_num":       "$new_recharge_num",
					"new_recharge_amount":    "$new_recharge_amount",
					"total_recharge":         "$total_recharge",
					"total_amount":           "$total_amount",
					"pay_request":            "$pay_request",
					"pay_request_order":      "$pay_request_order",
					"pay_success_order":      "$pay_success_order",
					"withdraw_request":       "$withdraw_request",
					"withdraw_request_order": "$withdraw_request_order",
					"withdraw_success_order": "$withdraw_success_order",
					"total_withdraw_num":     "$total_withdraw_num",
					"total_withdraw_amount":  "$total_withdraw_amount",
					"handling_charge":        "$handling_charge",
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
					"new_register":           bson.M{"$sum": "$new_register"},
					"new_equipment":          bson.M{"$sum": "$new_equipment"},
					"login_num":              bson.M{"$sum": "$login_num"},
					"next_new_register":      bson.M{"$sum": "$next_new_register"},
					"next_login":             bson.M{"$sum": "$next_login"},
					"jr_login_next_pay":      bson.M{"$sum": "$jr_login_next_pay"},
					"zr_pay_count":           bson.M{"$sum": "$zr_pay_count"},
					"new_recharge_num":       bson.M{"$sum": "$new_recharge_num"},
					"new_recharge_amount":    bson.M{"$sum": "$new_recharge_amount"},
					"total_recharge":         bson.M{"$sum": "$total_recharge"},
					"total_amount":           bson.M{"$sum": "$total_amount"},
					"pay_request":            bson.M{"$sum": "$pay_request"},
					"pay_request_order":      bson.M{"$sum": "$pay_request_order"},
					"pay_success_order":      bson.M{"$sum": "$pay_success_order"},
					"withdraw_request":       bson.M{"$sum": "$withdraw_request"},
					"withdraw_request_order": bson.M{"$sum": "$withdraw_request_order"},
					"withdraw_success_order": bson.M{"$sum": "$withdraw_success_order"},
					"total_withdraw_num":     bson.M{"$sum": "$total_withdraw_num"},
					"total_withdraw_amount":  bson.M{"$sum": "$total_withdraw_amount"},
					"handling_charge":        bson.M{"$sum": "$handling_charge"},
				},
			},
			{
				"$project": bson.M{
					"_id":                    "$_id.date",
					"date":                   "$_id.date",
					"new_register":           "$new_register",
					"new_equipment":          "$new_equipment",
					"login_num":              "$login_num",
					"next_new_register":      "$next_new_register",
					"next_login":             "$next_login",
					"jr_login_next_pay":      "$jr_login_next_pay",
					"zr_pay_count":           "$zr_pay_count",
					"new_recharge_num":       "$new_recharge_num",
					"new_recharge_amount":    "$new_recharge_amount",
					"total_recharge":         "$total_recharge",
					"total_amount":           "$total_amount",
					"pay_request":            "$pay_request",
					"pay_request_order":      "$pay_request_order",
					"pay_success_order":      "$pay_success_order",
					"withdraw_request":       "$withdraw_request",
					"withdraw_request_order": "$withdraw_request_order",
					"withdraw_success_order": "$withdraw_success_order",
					"total_withdraw_num":     "$total_withdraw_num",
					"total_withdraw_amount":  "$total_withdraw_amount",
					"handling_charge":        "$handling_charge",
				},
			},
			{"$sort": bson.M{"date": -1}},
			{"$skip": skipNum},
			{"$limit": pageSize},
		}
	}

	err := DataStatisticss.Pipe(pipeline).All(&list)
	list = this.chipList(list)
	return list, err
}

// 查询数据汇总条数
func (this *statisticsService) GetDataTotal(m bson.M, isChannel bool) (int64, error) {
	pipeline := []bson.M{}
	if isChannel {
		// 按渠道
		pipeline = []bson.M{
			{
				"$match": m,
			},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":      "$date",
						"channel":   "$channel",
						"channel1":  "$channel1",
						"bloggerId": "$blogger_id",
					},
				},
			},
		}
	} else {
		// 全部
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
	result := []bson.M{}
	pipe := DataStatisticss.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("GetDataTotal fail err: ", err)
	}
	return int64(len(result)), nil
}

// 转换为分展示
func (this *statisticsService) chipList(list []entity.DataStatistics) []entity.DataStatistics {
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		v.NextDayRetention = ComputeFloat(v.NextLogin, v.NextNewRegister) * 100
		v.PayRetention = ComputeFloat(v.JRLoginNextPay, v.ZRPayCount) * 100
		FRechargeAmount := ComputeFloat(v.NewRechargeAmount, 100)
		v.NewRechargeArpu = ComputeFloat(int64(FRechargeAmount), v.NewRegister)     //ComputeFloat(FTotalAmount, v.NewRegister) * 100
		v.NewRechargeArppu = ComputeFloat(int64(FRechargeAmount), v.NewRechargeNum) // ComputeFloat(v.NewRechargeAmount, v.NewRechargeNum) * 100
		v.NewUserPaymentRate = ComputeFloat(v.NewRechargeNum, v.NewRegister) * 100
		FTotalAmount := ComputeFloat(v.TotalAmount, 100)
		v.TotalRechargeArpu = ComputeFloat(int64(FTotalAmount), v.LoginNum)       // ComputeFloat(v.TotalAmount, v.LoginNum) * 100
		v.TotalRechargeArppu = ComputeFloat(int64(FTotalAmount), v.TotalRecharge) // ComputeFloat(v.TotalAmount, v.TotalRecharge) * 100
		v.TotalPaymentRate = ComputeFloat(v.TotalRecharge, v.LoginNum) * 100
		v.PaySuccessRate = ComputeFloat(v.PaySuccessOrder, v.PayRequestOrder) * 100
		v.WithdrawSuccessRate = ComputeFloat(v.WithdrawSuccessOrder, v.WithdrawRequestOrder) * 100
		v.CostRatio = ComputeFloat(v.TotalWithdrawAmount, v.TotalAmount) * 100
		v.FNewRechargeAmount = FRechargeAmount
		v.FTotalAmount = FTotalAmount
		v.FTotalWithdrawAmount = ComputeFloat(v.TotalWithdrawAmount, 100)
		v.FHandlingCharge = ComputeFloat(v.HandlingCharge, 100)
		list[k] = v
	}
	return list
}

// 新增|修改 数据汇总（新）
func (this *statisticsService) AddDataStatistics(channel *entity.DataStatistics) error {
	info := new(entity.DataStatistics)
	GetByQ(DataStatisticss, bson.M{"date": channel.Date, "channel": channel.Channel, "blogger_id": channel.BloggerId}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"channel1":               channel.Channel1,
			"blogger_id":             channel.BloggerId,
			"new_register":           channel.NewRegister,
			"new_equipment":          channel.NewEquipment,
			"login_num":              channel.LoginNum,
			"next_new_register":      channel.NextNewRegister,
			"next_login":             channel.NextLogin,
			"new_recharge_num":       channel.NewRechargeNum,
			"new_recharge_amount":    channel.NewRechargeAmount,
			"total_recharge":         channel.TotalRecharge,
			"total_amount":           channel.TotalAmount,
			"pay_request":            channel.PayRequest,
			"pay_request_order":      channel.PayRequestOrder,
			"pay_success_order":      channel.PaySuccessOrder,
			"withdraw_request":       channel.WithdrawRequest,
			"withdraw_request_order": channel.WithdrawRequestOrder,
			"withdraw_success_order": channel.WithdrawSuccessOrder,
			"total_withdraw_num":     channel.TotalWithdrawNum,
			"total_withdraw_amount":  channel.TotalWithdrawAmount,
			"handling_charge":        channel.HandlingCharge,
		}
		if Update(DataStatisticss, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + channel.Id)
	} else {
		// 新增
		channel.Id = bson.NewObjectId().Hex()
		if !Insert(DataStatisticss, channel) {
			return errors.New("写入失败:" + channel.Id)
		}
		return nil
	}
}

/*
分享数据相关接口
*/
func (this *statisticsService) GetShareList(page, pageSize int, m bson.M, isChannel bool) ([]entity.ShareDataStatistics, error) {
	var list []entity.ShareDataStatistics
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
						"date":       "$date",
						"channel":    "$channel",
						"channel1":   "$channel1",
						"blogger_id": "$blogger_id",
					},
					"register_number": bson.M{"$sum": "$register_number"},
					"pay_number":      bson.M{"$sum": "$pay_number"},
					"pay_money":       bson.M{"$sum": "$pay_money"},
				},
			},
			{
				"$project": bson.M{
					"_id":             "$_id.date",
					"date":            "$_id.date",
					"channel":         "$_id.channel",
					"channel1":        "$_id.channel1",
					"blogger_id":      "$_id.blogger_id",
					"register_number": "$register_number",
					"pay_number":      "$pay_number",
					"pay_money":       "$pay_money",
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
					"register_number": bson.M{"$sum": "$register_number"},
					"pay_number":      bson.M{"$sum": "$pay_number"},
					"pay_money":       bson.M{"$sum": "$pay_money"},
				},
			},
			{
				"$project": bson.M{
					"_id":             "$_id.date",
					"date":            "$_id.date",
					"register_number": "$register_number",
					"pay_number":      "$pay_number",
					"pay_money":       "$pay_money",
				},
			},
			{"$sort": bson.M{"date": -1}},
			{"$skip": skipNum},
			{"$limit": pageSize},
		}
	}

	err := ShareStatistcs.Pipe(pipeline).All(&list)
	list = this.chipList1(list)
	return list, err
}

func (this *statisticsService) GetShareTotal(m bson.M, isChannel bool) (int64, error) {
	pipeline := []bson.M{}
	if isChannel {
		// 渠道
		pipeline = []bson.M{
			{
				"$match": m,
			},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":     "$date",
						"channel":  "$channel",
						"channel1": "$channel1",
					},
				},
			},
		}
	} else {
		// 全部
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
	pipe := ShareStatistcs.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("GetDataTotal fail err: ", err)
	}
	return int64(len(result)), nil
}

func (this *statisticsService) chipList1(list []entity.ShareDataStatistics) []entity.ShareDataStatistics {
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		FPayMoney := ComputeFloat(v.PayMoney, 100)
		v.FPayMoney = FPayMoney
		v.ShareArpu = ComputeFloat(int64(FPayMoney), v.RegisterNumber)
		v.ShareArppu = ComputeFloat(int64(FPayMoney), v.PayNumber)
		list[k] = v
	}
	return list
}

// 新增|修改 分享数据
func (this *statisticsService) AddOrUpdateShare(rate *entity.ShareDataStatistics) error {
	info := new(entity.ShareDataStatistics)
	GetByQ(ShareStatistcs, bson.M{"date": rate.Date, "channel": rate.Channel, "blogger_id": rate.BloggerId}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"channel1":        rate.Channel1,
			"blogger_id":      rate.BloggerId,
			"register_number": rate.RegisterNumber,
			"pay_number":      rate.PayNumber,
			"pay_money":       rate.PayMoney,
		}
		if Update(ShareStatistcs, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + rate.Id)
	} else {
		// 新增
		rate.Id = bson.NewObjectId().Hex()
		if !Insert(ShareStatistcs, rate) {
			return errors.New("写入失败:" + rate.Id)
		}
		return nil
	}
}

/*
	用户留存
*/
// 查询用户留存信息
func (this *statisticsService) GetRetainedList(page, pageSize int, m bson.M, isChannel bool) ([]entity.UserRetained, error) {
	var list []entity.UserRetained
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
						"date":       "$date",
						"channel":    "$channel",
						"channel1":   "$channel1",
						"blogger_id": "$blogger_id",
					},
					"new_number": bson.M{"$sum": "$new_number"},
					"login1":     bson.M{"$sum": "$login1"},
					"register1":  bson.M{"$sum": "$register1"},
					"login2":     bson.M{"$sum": "$login2"},
					"register2":  bson.M{"$sum": "$register2"},
					"login3":     bson.M{"$sum": "$login3"},
					"register3":  bson.M{"$sum": "$register3"},
					"login4":     bson.M{"$sum": "$login4"},
					"register4":  bson.M{"$sum": "$register4"},
					"login5":     bson.M{"$sum": "$login5"},
					"register5":  bson.M{"$sum": "$register5"},
					"login6":     bson.M{"$sum": "$login6"},
					"register6":  bson.M{"$sum": "$register6"},
					"login7":     bson.M{"$sum": "$login7"},
					"register7":  bson.M{"$sum": "$register7"},
					"login15":    bson.M{"$sum": "$login15"},
					"register15": bson.M{"$sum": "$register15"},
					"login30":    bson.M{"$sum": "$login30"},
					"register30": bson.M{"$sum": "$register30"},
					"login60":    bson.M{"$sum": "$login60"},
					"register60": bson.M{"$sum": "$register60"},
				},
			},
			{
				"$project": bson.M{
					"_id":        "$_id.date",
					"date":       "$_id.date",
					"channel":    "$_id.channel",
					"channel1":   "$_id.channel1",
					"blogger_id": "$_id.blogger_id",
					"new_number": "$new_number",
					"login1":     "$login1",
					"register1":  "$register1",
					"login2":     "$login2",
					"register2":  "$register2",
					"login3":     "$login3",
					"register3":  "$register3",
					"login4":     "$login4",
					"register4":  "$register4",
					"login5":     "$login5",
					"register5":  "$register5",
					"login6":     "$login6",
					"register6":  "$register6",
					"login7":     "$login7",
					"register7":  "$register7",
					"login15":    "$login15",
					"register15": "$register15",
					"login30":    "$login30",
					"register30": "$register30",
					"login60":    "$login60",
					"register60": "$register60",
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
					"new_number": bson.M{"$sum": "$new_number"},
					"login1":     bson.M{"$sum": "$login1"},
					"register1":  bson.M{"$sum": "$register1"},
					"login2":     bson.M{"$sum": "$login2"},
					"register2":  bson.M{"$sum": "$register2"},
					"login3":     bson.M{"$sum": "$login3"},
					"register3":  bson.M{"$sum": "$register3"},
					"login4":     bson.M{"$sum": "$login4"},
					"register4":  bson.M{"$sum": "$register4"},
					"login5":     bson.M{"$sum": "$login5"},
					"register5":  bson.M{"$sum": "$register5"},
					"login6":     bson.M{"$sum": "$login6"},
					"register6":  bson.M{"$sum": "$register6"},
					"login7":     bson.M{"$sum": "$login7"},
					"register7":  bson.M{"$sum": "$register7"},
					"login15":    bson.M{"$sum": "$login15"},
					"register15": bson.M{"$sum": "$register15"},
					"login30":    bson.M{"$sum": "$login30"},
					"register30": bson.M{"$sum": "$register30"},
					"login60":    bson.M{"$sum": "$login60"},
					"register60": bson.M{"$sum": "$register60"},
				},
			},
			{
				"$project": bson.M{
					"_id":        "$_id.date",
					"date":       "$_id.date",
					"new_number": "$new_number",
					"login1":     "$login1",
					"register1":  "$register1",
					"login2":     "$login2",
					"register2":  "$register2",
					"login3":     "$login3",
					"register3":  "$register3",
					"login4":     "$login4",
					"register4":  "$register4",
					"login5":     "$login5",
					"register5":  "$register5",
					"login6":     "$login6",
					"register6":  "$register6",
					"login7":     "$login7",
					"register7":  "$register7",
					"login15":    "$login15",
					"register15": "$register15",
					"login30":    "$login30",
					"register30": "$register30",
					"login60":    "$login60",
					"register60": "$register60",
				},
			},
			{"$sort": bson.M{"date": -1}},
			{"$skip": skipNum},
			{"$limit": pageSize},
		}
	}

	err := UserRetaineds.Pipe(pipeline).All(&list)
	list = this.chipList2(list)
	return list, err
}

// 转换为分展示
func (this *statisticsService) chipList2(list []entity.UserRetained) []entity.UserRetained {
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		v.Day1 = ComputeFloat(v.Login1, v.Register1) * 100
		v.Day2 = ComputeFloat(v.Login2, v.Register2) * 100
		v.Day3 = ComputeFloat(v.Login3, v.Register3) * 100
		v.Day4 = ComputeFloat(v.Login4, v.Register4) * 100
		v.Day5 = ComputeFloat(v.Login5, v.Register5) * 100
		v.Day6 = ComputeFloat(v.Login6, v.Register6) * 100
		v.Day7 = ComputeFloat(v.Login7, v.Register7) * 100
		v.Day15 = ComputeFloat(v.Login15, v.Register15) * 100
		v.Day30 = ComputeFloat(v.Login30, v.Register30) * 100
		v.Day60 = ComputeFloat(v.Login60, v.Register60) * 100
		list[k] = v
	}
	return list
}

// 查询用户留存条数
func (this *statisticsService) GetRetainedTotal(m bson.M, isChannel bool) (int64, error) {
	pipeline := []bson.M{}
	if isChannel {
		// 按渠道
		pipeline = []bson.M{
			{
				"$match": m,
			},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":      "$date",
						"channel":   "$channel",
						"channel1":  "$channel1",
						"bloggerId": "$blogger_id",
					},
				},
			},
		}
	} else {
		// 全部
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
	result := []bson.M{}
	pipe := UserRetaineds.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("GetRetainedTotal fail err: ", err)
	}
	return int64(len(result)), nil
}

// 新增或更新用户留存数据
func (this *statisticsService) AddUserRetained(user *entity.UserRetained) error {
	info := new(entity.UserRetained)
	GetByQ(UserRetaineds, bson.M{"date": user.Date, "channel": user.Channel, "blogger_id": user.BloggerId}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"channel1":   user.Channel1,
			"blogger_id": user.BloggerId,
			"new_number": user.NewNumber,
			"login1":     user.Login1,
			"register1":  user.Register1,
			"login2":     user.Login2,
			"register2":  user.Register2,
			"login3":     user.Login3,
			"register3":  user.Register3,
			"login4":     user.Login4,
			"register4":  user.Register4,
			"login5":     user.Login5,
			"register5":  user.Register5,
			"login6":     user.Login6,
			"register6":  user.Register6,
			"login7":     user.Login7,
			"register7":  user.Register7,
			"login15":    user.Login15,
			"register15": user.Register15,
			"login30":    user.Login30,
			"register30": user.Register30,
			"login60":    user.Login60,
			"register60": user.Register60,
		}
		if Update(UserRetaineds, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + user.Id)
	} else {
		// 新增
		user.Id = bson.NewObjectId().Hex()
		if !Insert(UserRetaineds, user) {
			return errors.New("写入失败:" + user.Id)
		}
		return nil
	}

}

/*
	付费用户留存
*/
// 分页查询付费用户留存
func (this *statisticsService) GetPayUserRetainedList(page, pageSize int, m bson.M, isChannel bool) ([]entity.PayUserRetained, error) {
	// var list []entity.PayUserRetained
	// if pageSize == -1 {
	// 	pageSize = 100000
	// }
	// // skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	// skipNum, sortFieldR := parsePageAndSort(page, pageSize, "date", false)
	// err := PayUserRetaineds.
	// 	Find(m).
	// 	Sort(sortFieldR).
	// 	Skip(skipNum).
	// 	Limit(pageSize).
	// 	All(&list)
	// list = this.chipList17(list)
	// return list, err
	var list []entity.PayUserRetained
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
						"date":       "$date",
						"channel":    "$channel",
						"channel1":   "$channel1",
						"blogger_id": "$blogger_id",
					},
					"new_number": bson.M{"$sum": "$new_number"},
					"login1":     bson.M{"$sum": "$login1"},
					"register1":  bson.M{"$sum": "$register1"},
					"login2":     bson.M{"$sum": "$login2"},
					"register2":  bson.M{"$sum": "$register2"},
					"login3":     bson.M{"$sum": "$login3"},
					"register3":  bson.M{"$sum": "$register3"},
					"login4":     bson.M{"$sum": "$login4"},
					"register4":  bson.M{"$sum": "$register4"},
					"login5":     bson.M{"$sum": "$login5"},
					"register5":  bson.M{"$sum": "$register5"},
					"login6":     bson.M{"$sum": "$login6"},
					"register6":  bson.M{"$sum": "$register6"},
					"login7":     bson.M{"$sum": "$login7"},
					"register7":  bson.M{"$sum": "$register7"},
					"login15":    bson.M{"$sum": "$login15"},
					"register15": bson.M{"$sum": "$register15"},
					"login30":    bson.M{"$sum": "$login30"},
					"register30": bson.M{"$sum": "$register30"},
					"login60":    bson.M{"$sum": "$login60"},
					"register60": bson.M{"$sum": "$register60"},
				},
			},
			{
				"$project": bson.M{
					"_id":        "$_id.date",
					"date":       "$_id.date",
					"channel":    "$_id.channel",
					"channel1":   "$_id.channel1",
					"blogger_id": "$_id.blogger_id",
					"new_number": "$new_number",
					"login1":     "$login1",
					"register1":  "$register1",
					"login2":     "$login2",
					"register2":  "$register2",
					"login3":     "$login3",
					"register3":  "$register3",
					"login4":     "$login4",
					"register4":  "$register4",
					"login5":     "$login5",
					"register5":  "$register5",
					"login6":     "$login6",
					"register6":  "$register6",
					"login7":     "$login7",
					"register7":  "$register7",
					"login15":    "$login15",
					"register15": "$register15",
					"login30":    "$login30",
					"register30": "$register30",
					"login60":    "$login60",
					"register60": "$register60",
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
					"new_number": bson.M{"$sum": "$new_number"},
					"login1":     bson.M{"$sum": "$login1"},
					"register1":  bson.M{"$sum": "$register1"},
					"login2":     bson.M{"$sum": "$login2"},
					"register2":  bson.M{"$sum": "$register2"},
					"login3":     bson.M{"$sum": "$login3"},
					"register3":  bson.M{"$sum": "$register3"},
					"login4":     bson.M{"$sum": "$login4"},
					"register4":  bson.M{"$sum": "$register4"},
					"login5":     bson.M{"$sum": "$login5"},
					"register5":  bson.M{"$sum": "$register5"},
					"login6":     bson.M{"$sum": "$login6"},
					"register6":  bson.M{"$sum": "$register6"},
					"login7":     bson.M{"$sum": "$login7"},
					"register7":  bson.M{"$sum": "$register7"},
					"login15":    bson.M{"$sum": "$login15"},
					"register15": bson.M{"$sum": "$register15"},
					"login30":    bson.M{"$sum": "$login30"},
					"register30": bson.M{"$sum": "$register30"},
					"login60":    bson.M{"$sum": "$login60"},
					"register60": bson.M{"$sum": "$register60"},
				},
			},
			{
				"$project": bson.M{
					"_id":        "$_id.date",
					"date":       "$_id.date",
					"new_number": "$new_number",
					"login1":     "$login1",
					"register1":  "$register1",
					"login2":     "$login2",
					"register2":  "$register2",
					"login3":     "$login3",
					"register3":  "$register3",
					"login4":     "$login4",
					"register4":  "$register4",
					"login5":     "$login5",
					"register5":  "$register5",
					"login6":     "$login6",
					"register6":  "$register6",
					"login7":     "$login7",
					"register7":  "$register7",
					"login15":    "$login15",
					"register15": "$register15",
					"login30":    "$login30",
					"register30": "$register30",
					"login60":    "$login60",
					"register60": "$register60",
				},
			},
			{"$sort": bson.M{"date": -1}},
			{"$skip": skipNum},
			{"$limit": pageSize},
		}
	}

	err := PayUserRetaineds.Pipe(pipeline).All(&list)
	list = this.chipList17(list)
	return list, err
}

// 转换为分展示
func (this *statisticsService) chipList17(list []entity.PayUserRetained) []entity.PayUserRetained {
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		v.Day1 = ComputeFloat(v.Login1, v.Register1) * 100
		v.Day2 = ComputeFloat(v.Login2, v.Register2) * 100
		v.Day3 = ComputeFloat(v.Login3, v.Register3) * 100
		v.Day4 = ComputeFloat(v.Login4, v.Register4) * 100
		v.Day5 = ComputeFloat(v.Login5, v.Register5) * 100
		v.Day6 = ComputeFloat(v.Login6, v.Register6) * 100
		v.Day7 = ComputeFloat(v.Login7, v.Register7) * 100
		v.Day15 = ComputeFloat(v.Login15, v.Register15) * 100
		v.Day30 = ComputeFloat(v.Login30, v.Register30) * 100
		v.Day60 = ComputeFloat(v.Login60, v.Register60) * 100
		list[k] = v
	}
	return list
}

// 查询付费用户留存条数
func (this *statisticsService) GetPayUserRetainedTotal(m bson.M, isChannel bool) (int64, error) {
	pipeline := []bson.M{}
	if isChannel {
		// 按渠道
		pipeline = []bson.M{
			{
				"$match": m,
			},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":      "$date",
						"channel":   "$channel",
						"channel1":  "$channel1",
						"bloggerId": "$blogger_id",
					},
				},
			},
		}
	} else {
		// 全部
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
	result := []bson.M{}
	pipe := PayUserRetaineds.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("GetPayUserRetainedTotal fail err: ", err)
	}
	return int64(len(result)), nil
}

// 新增或更新付费用户留存
func (this *statisticsService) AddPayUserRetained(user *entity.PayUserRetained) error {
	info := new(entity.PayUserRetained)
	GetByQ(PayUserRetaineds, bson.M{"date": user.Date, "channel": user.Channel, "blogger_id": user.BloggerId}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"channel1":   user.Channel1,
			"blogger_id": user.BloggerId,
			"new_number": user.NewNumber,
			"login1":     user.Login1,
			"register1":  user.Register1,
			"login2":     user.Login2,
			"register2":  user.Register2,
			"login3":     user.Login3,
			"register3":  user.Register3,
			"login4":     user.Login4,
			"register4":  user.Register4,
			"login5":     user.Login5,
			"register5":  user.Register5,
			"login6":     user.Login6,
			"register6":  user.Register6,
			"login7":     user.Login7,
			"register7":  user.Register7,
			"login15":    user.Login15,
			"register15": user.Register15,
			"login30":    user.Login30,
			"register30": user.Register30,
			"login60":    user.Login60,
			"register60": user.Register60,
		}
		if Update(PayUserRetaineds, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + user.Id)
	} else {
		// 新增
		user.Id = bson.NewObjectId().Hex()
		if !Insert(PayUserRetaineds, user) {
			return errors.New("写入失败:" + user.Id)
		}
		return nil
	}

}

/*
 时间分析相关接口 Playtimes
*/
// 时间分析分页接口
func (this *statisticsService) GetPlaytimeList(page, pageSize int, m bson.M, isChannel bool) ([]entity.PlaytimeAnalysis, error) {
	var list []entity.PlaytimeAnalysis
	if pageSize == -1 {
		pageSize = 100000
	}
	// skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	// skipNum, sortFieldR := parsePageAndSort(page, pageSize, "date", false)
	// err := Playtimes.
	// 	Find(m).
	// 	Sort(sortFieldR).
	// 	Skip(skipNum).
	// 	Limit(pageSize).
	// 	All(&list)
	// list = this.chipList15(list)
	// return list, err
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)
	pipeline := []bson.M{}
	if isChannel {
		// 按渠道
		pipeline = []bson.M{
			{"$match": m},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":     "$date",
						"channel":  "$channel",
						"channel1": "$channel1",
					},
					"total_register": bson.M{"$sum": "$total_register"},
					"playtime1":      bson.M{"$sum": "$playtime1"},
					"playtime2":      bson.M{"$sum": "$playtime2"},
					"playtime6":      bson.M{"$sum": "$playtime6"},
					"playtime11":     bson.M{"$sum": "$playtime11"},
					"playtime21":     bson.M{"$sum": "$playtime21"},
					"playtime31":     bson.M{"$sum": "$playtime31"},
				},
			},
			{
				"$project": bson.M{
					"_id":            "$_id.date",
					"date":           "$_id.date",
					"channel":        "$_id.channel",
					"channel1":       "$_id.channel1",
					"total_register": "$total_register",
					"playtime1":      "$playtime1",
					"playtime2":      "$playtime2",
					"playtime6":      "$playtime6",
					"playtime11":     "$playtime11",
					"playtime21":     "$playtime21",
					"playtime31":     "$playtime31",
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
					"total_register": bson.M{"$sum": "$total_register"},
					"playtime1":      bson.M{"$sum": "$playtime1"},
					"playtime2":      bson.M{"$sum": "$playtime2"},
					"playtime6":      bson.M{"$sum": "$playtime6"},
					"playtime11":     bson.M{"$sum": "$playtime11"},
					"playtime21":     bson.M{"$sum": "$playtime21"},
					"playtime31":     bson.M{"$sum": "$playtime31"},
				},
			},
			{
				"$project": bson.M{
					"_id":            "$_id.date",
					"date":           "$_id.date",
					"total_register": "$total_register",
					"playtime1":      "$playtime1",
					"playtime2":      "$playtime2",
					"playtime6":      "$playtime6",
					"playtime11":     "$playtime11",
					"playtime21":     "$playtime21",
					"playtime31":     "$playtime31",
				},
			},
			{"$sort": bson.M{"date": -1}},
			{"$skip": skipNum},
			{"$limit": pageSize},
		}
	}
	err := Playtimes.Pipe(pipeline).All(&list)
	list = this.chipList15(list)
	return list, err
}

// 转换为分展示
func (this *statisticsService) chipList15(list []entity.PlaytimeAnalysis) []entity.PlaytimeAnalysis {
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		v.PlaytimeRatio1 = ComputeFloat(v.Playtime1, v.TotalRegister) * 100
		v.PlaytimeRatio2 = ComputeFloat(v.Playtime2, v.TotalRegister) * 100
		v.PlaytimeRatio6 = ComputeFloat(v.Playtime6, v.TotalRegister) * 100
		v.PlaytimeRatio11 = ComputeFloat(v.Playtime11, v.TotalRegister) * 100
		v.PlaytimeRatio21 = ComputeFloat(v.Playtime21, v.TotalRegister) * 100
		v.PlaytimeRatio31 = ComputeFloat(v.Playtime31, v.TotalRegister) * 100
		list[k] = v
	}
	return list
}

// 查询Bug数据条数
func (this *statisticsService) GetPlaytimeTotal(m bson.M, isChannel bool) (int64, error) {
	// return int64(Count(Playtimes, m)), nil
	pipeline := []bson.M{}
	if isChannel {
		pipeline = []bson.M{
			{
				"$match": m,
			},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":     "$date",
						"channel":  "$channel",
						"channel1": "$channel1",
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
	pipe := Playtimes.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("GetPlaytimeTotal fail err: ", err)
	}
	return int64(len(result)), nil
}

// 新增Bug数据
func (this *statisticsService) AddPlaytimeAnalysis(rate *entity.PlaytimeAnalysis) error {
	info := new(entity.PlaytimeAnalysis)
	GetByQ(Playtimes, bson.M{"date": rate.Date, "channel": rate.Channel, "blogger_id": rate.BloggerId}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"channel1":       rate.Channel1,
			"total_register": rate.TotalRegister,
			"playtime1":      rate.Playtime1,
			"playtime2":      rate.Playtime2,
			"playtime6":      rate.Playtime6,
			"playtime11":     rate.Playtime11,
			"playtime21":     rate.Playtime21,
			"playtime31":     rate.Playtime31,
		}
		if Update(Playtimes, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + rate.Id)
	} else {
		// 新增
		rate.Id = bson.NewObjectId().Hex()
		if !Insert(Playtimes, rate) {
			return errors.New("写入失败:" + rate.Id)
		}
		return nil
	}
}

// func (this *statisticsService) GetLogGameTimes(m bson.M) (int64, error) {
// 	return int64(Count(LogGameTimes, m)), nil
// }

/*
	局数分析相关接口
*/

// 分页查询局数分析数据列表
func (this *statisticsService) GetGameNumberAnalysisList(page, pageSize int, m bson.M, isChannel bool) ([]entity.GameNumberAnalysis, error) {
	var list []entity.GameNumberAnalysis
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)
	pipeline := []bson.M{}
	if isChannel {
		// 按渠道
		pipeline = []bson.M{
			{"$match": m},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":     "$date",
						"channel":  "$channel",
						"channel1": "$channel1",
					},
					"total_register": bson.M{"$sum": "$total_register"},
					"g_number":       bson.M{"$sum": "$g_number"},
					"g_number1":      bson.M{"$sum": "$g_number1"},
					"g_number2":      bson.M{"$sum": "$g_number2"},
					"g_number3":      bson.M{"$sum": "$g_number3"},
					"g_number4":      bson.M{"$sum": "$g_number4"},
					"g_number5":      bson.M{"$sum": "$g_number5"},
					"g_number6":      bson.M{"$sum": "$g_number6"},
					"g_number11":     bson.M{"$sum": "$g_number11"},
					"g_number21":     bson.M{"$sum": "$g_number21"},
					"g_number31":     bson.M{"$sum": "$g_number31"},
				},
			},
			{
				"$project": bson.M{
					"_id":            "$_id.date",
					"date":           "$_id.date",
					"channel":        "$_id.channel",
					"channel1":       "$_id.channel1",
					"total_register": "$total_register",
					"g_number":       "$g_number",
					"g_number1":      "$g_number1",
					"g_number2":      "$g_number2",
					"g_number3":      "$g_number3",
					"g_number4":      "$g_number4",
					"g_number5":      "$g_number5",
					"g_number6":      "$g_number6",
					"g_number11":     "$g_number11",
					"g_number21":     "$g_number21",
					"g_number31":     "$g_number31",
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
					"total_register": bson.M{"$sum": "$total_register"},
					"g_number":       bson.M{"$sum": "$g_number"},
					"g_number1":      bson.M{"$sum": "$g_number1"},
					"g_number2":      bson.M{"$sum": "$g_number2"},
					"g_number3":      bson.M{"$sum": "$g_number3"},
					"g_number4":      bson.M{"$sum": "$g_number4"},
					"g_number5":      bson.M{"$sum": "$g_number5"},
					"g_number6":      bson.M{"$sum": "$g_number6"},
					"g_number11":     bson.M{"$sum": "$g_number11"},
					"g_number21":     bson.M{"$sum": "$g_number21"},
					"g_number31":     bson.M{"$sum": "$g_number31"},
				},
			},
			{
				"$project": bson.M{
					"_id":            "$_id.date",
					"date":           "$_id.date",
					"total_register": "$total_register",
					"g_number":       "$g_number",
					"g_number1":      "$g_number1",
					"g_number2":      "$g_number2",
					"g_number3":      "$g_number3",
					"g_number4":      "$g_number4",
					"g_number5":      "$g_number5",
					"g_number6":      "$g_number6",
					"g_number11":     "$g_number11",
					"g_number21":     "$g_number21",
					"g_number31":     "$g_number31",
				},
			},
			{"$sort": bson.M{"date": -1}},
			{"$skip": skipNum},
			{"$limit": pageSize},
		}
	}
	err := GameNumberAnalysiss.Pipe(pipeline).All(&list)
	list = this.chipList14(list)
	return list, err
}

// 转换为分展示
func (this *statisticsService) chipList14(list []entity.GameNumberAnalysis) []entity.GameNumberAnalysis {
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		v.GNumberRatio = ComputeFloat(v.GNumber, v.TotalRegister) * 100
		v.GNumberRatio1 = ComputeFloat(v.GNumber1, v.TotalRegister) * 100
		v.GNumberRatio2 = ComputeFloat(v.GNumber2, v.TotalRegister) * 100
		v.GNumberRatio3 = ComputeFloat(v.GNumber3, v.TotalRegister) * 100
		v.GNumberRatio4 = ComputeFloat(v.GNumber4, v.TotalRegister) * 100
		v.GNumberRatio5 = ComputeFloat(v.GNumber5, v.TotalRegister) * 100
		v.GNumberRatio6 = ComputeFloat(v.GNumber6, v.TotalRegister) * 100
		v.GNumberRatio11 = ComputeFloat(v.GNumber11, v.TotalRegister) * 100
		v.GNumberRatio21 = ComputeFloat(v.GNumber21, v.TotalRegister) * 100
		v.GNumberRatio31 = ComputeFloat(v.GNumber31, v.TotalRegister) * 100
		// v.FNewRechargeAmount = Chip2Float(v.NewRechargeAmount)
		// v.FTotalAmount = Chip2Float(v.TotalAmount)
		// v.FTotalWithdrawAmount = Chip2Float(v.TotalWithdrawAmount)
		list[k] = v
	}
	return list
}

// 查询局数分析数据条数
func (this *statisticsService) GetGameNumberAnalysisTotal(m bson.M, isChannel bool) (int64, error) {
	pipeline := []bson.M{}
	if isChannel {
		pipeline = []bson.M{
			{
				"$match": m,
			},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":     "$date",
						"channel":  "$channel",
						"channel1": "$channel1",
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
	pipe := GameNumberAnalysiss.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("GetGameNumberAnalysisTotal fail err: ", err)
	}
	return int64(len(result)), nil
	// return int64(Count(GameNumberAnalysiss, m)), nil
}

// 新增局数分析数据
func (this *statisticsService) AddGameNumberAnalysis(rate *entity.GameNumberAnalysis) error {
	info := new(entity.GameNumberAnalysis)
	GetByQ(GameNumberAnalysiss, bson.M{"date": rate.Date, "channel": rate.Channel, "blogger_id": rate.BloggerId}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"channel1":       rate.Channel1,
			"blogger_id":     rate.BloggerId,
			"total_register": rate.TotalRegister,
			"g_number":       rate.GNumber,
			"g_number1":      rate.GNumber1,
			"g_number2":      rate.GNumber2,
			"g_number3":      rate.GNumber3,
			"g_number4":      rate.GNumber4,
			"g_number5":      rate.GNumber5,
			"g_number6":      rate.GNumber6,
			"g_number11":     rate.GNumber11,
			"g_number21":     rate.GNumber21,
			"g_number31":     rate.GNumber31,
		}
		if Update(GameNumberAnalysiss, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + rate.Id)
	} else {
		// 新增
		rate.Id = bson.NewObjectId().Hex()
		if !Insert(GameNumberAnalysiss, rate) {
			return errors.New("写入失败:" + rate.Id)
		}
		return nil
	}
}

/*
	埋点统计相关接口
*/
// 查询埋点数据列表
func (this *statisticsService) GetPointList(page, pageSize int, m bson.M, isChannel bool) ([]entity.PointData, error) {
	var list []entity.PointData
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, _ := parsePageAndSort(page, pageSize, "date", false)
	pipeline := []bson.M{}
	if isChannel {
		// 按渠道
		pipeline = []bson.M{
			{"$match": m},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":     "$date",
						"channel":  "$channel",
						"channel1": "$channel1",
					},
					"total_register":    bson.M{"$sum": "$total_register"},
					"total_tourist":     bson.M{"$sum": "$total_tourist"},
					"total_mobile":      bson.M{"$sum": "$total_mobile"},
					"guidance_binding":  bson.M{"$sum": "$guidance_binding"},
					"guidance_get_gold": bson.M{"$sum": "$guidance_get_gold"},
					"tp_number1":        bson.M{"$sum": "$tp_number1"},
					"tp_number2":        bson.M{"$sum": "$tp_number2"},
					"tp_exit_manually":  bson.M{"$sum": "$tp_exit_manually"},
					"gold100":           bson.M{"$sum": "$gold100"},
					"gold200":           bson.M{"$sum": "$gold200"},
					"first_games":       bson.M{"$sum": "$first_games"},
					"second_games":      bson.M{"$sum": "$second_games"},
				},
			},
			{
				"$project": bson.M{
					"_id":               "$_id.date",
					"date":              "$_id.date",
					"channel":           "$_id.channel",
					"channel1":          "$_id.channel1",
					"total_register":    "$total_register",
					"total_tourist":     "$total_tourist",
					"total_mobile":      "$total_mobile",
					"guidance_binding":  "$guidance_binding",
					"guidance_get_gold": "$guidance_get_gold",
					"tp_number1":        "$tp_number1",
					"tp_number2":        "$tp_number2",
					"tp_exit_manually":  "$tp_exit_manually",
					"gold100":           "$gold100",
					"gold200":           "$gold200",
					"first_games":       "$first_games",
					"second_games":      "$second_games",
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
					"total_register":    bson.M{"$sum": "$total_register"},
					"total_tourist":     bson.M{"$sum": "$total_tourist"},
					"total_mobile":      bson.M{"$sum": "$total_mobile"},
					"guidance_binding":  bson.M{"$sum": "$guidance_binding"},
					"guidance_get_gold": bson.M{"$sum": "$guidance_get_gold"},
					"tp_number1":        bson.M{"$sum": "$tp_number1"},
					"tp_number2":        bson.M{"$sum": "$tp_number2"},
					"tp_exit_manually":  bson.M{"$sum": "$tp_exit_manually"},
					"gold100":           bson.M{"$sum": "$gold100"},
					"gold200":           bson.M{"$sum": "$gold200"},
					"first_games":       bson.M{"$sum": "$first_games"},
					"second_games":      bson.M{"$sum": "$second_games"},
				},
			},
			{
				"$project": bson.M{
					"_id":               "$_id.date",
					"date":              "$_id.date",
					"total_register":    "$total_register",
					"total_tourist":     "$total_tourist",
					"total_mobile":      "$total_mobile",
					"guidance_binding":  "$guidance_binding",
					"guidance_get_gold": "$guidance_get_gold",
					"tp_number1":        "$tp_number1",
					"tp_number2":        "$tp_number2",
					"tp_exit_manually":  "$tp_exit_manually",
					"gold100":           "$gold100",
					"gold200":           "$gold200",
					"first_games":       "$first_games",
					"second_games":      "$second_games",
				},
			},
			{"$sort": bson.M{"date": -1}},
			{"$skip": skipNum},
			{"$limit": pageSize},
		}
	}
	err := PointDatas.Pipe(pipeline).All(&list)
	list = this.chipList11(list)
	return list, err
}

// 转换为分展示
func (this *statisticsService) chipList11(list []entity.PointData) []entity.PointData {
	for k, v := range list {
		v.SDate = time.Unix(v.Date, 0)
		v.TotalTouristRatio = ComputeFloat(v.TotalTourist, v.TotalRegister) * 100
		v.TotalMobileRatio = ComputeFloat(v.TotalMobile, v.TotalRegister) * 100
		v.GuidanceGetGoldRatio = ComputeFloat(v.GuidanceGetGold, v.TotalRegister) * 100
		v.TPNumber1Ratio = ComputeFloat(v.TPNumber1, v.TotalRegister) * 100
		v.TPNumber2Ratio = ComputeFloat(v.TPNumber2, v.TotalRegister) * 100
		v.TPExitManuallyRatio = ComputeFloat(v.TPExitManually, v.TotalRegister) * 100
		v.Gold100Ratio = ComputeFloat(v.Gold100, v.TotalRegister) * 100
		v.Gold200Ratio = ComputeFloat(v.Gold200, v.TotalRegister) * 100
		list[k] = v
	}
	return list
}

// 查询埋点数据条数
func (this *statisticsService) GetPointListTotal(m bson.M, isChannel bool) (int64, error) {
	pipeline := []bson.M{}
	if isChannel {
		pipeline = []bson.M{
			{
				"$match": m,
			},
			{
				"$group": bson.M{
					"_id": bson.M{
						"date":     "$date",
						"channel":  "$channel",
						"channel1": "$channel1",
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
	result := []bson.M{}
	pipe := PointDatas.Pipe(pipeline)
	err := pipe.All(&result)
	if err != nil {
		beego.Error("GetPointListTotal fail err: ", err)
	}
	return int64(len(result)), nil
}

// 新增埋点数据
func (this *statisticsService) AddPointData(rate *entity.PointData) error {
	info := new(entity.PointData)
	GetByQ(PointDatas, bson.M{"date": rate.Date, "channel": rate.Channel, "blogger_id": rate.BloggerId}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"channel1":          rate.Channel1,
			"blogger_id":        rate.BloggerId,
			"total_register":    rate.TotalRegister,
			"total_tourist":     rate.TotalTourist,
			"total_mobile":      rate.TotalMobile,
			"guidance_binding":  rate.GuidanceBinding,
			"guidance_get_gold": rate.GuidanceGetGold,
			"tp_number1":        rate.TPNumber1,
			"tp_number2":        rate.TPNumber2,
			"tp_exit_manually":  rate.TPExitManually,
			"gold100":           rate.Gold100,
			"gold200":           rate.Gold200,
			"first_games":       rate.FirstGames,
			"second_games":      rate.SecondGames,
		}
		if Update(PointDatas, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + rate.Id)
	} else {
		// 新增
		rate.Id = bson.NewObjectId().Hex()
		if !Insert(PointDatas, rate) {
			return errors.New("写入失败:" + rate.Id)
		}
		return nil
	}
}

// 根据ID查询埋点数据信息
func (this *statisticsService) GetPointById(id int64, channelname string) (*entity.PointData, error) {
	info := new(entity.PointData)
	// pipeline := []bson.M{}
	// if channelname != "" {
	// 	// 按渠道
	// 	m := bson.M{"date": id, "channel": channelname}
	// 	pipeline = []bson.M{
	// 		{"$match": m},
	// 		{
	// 			"$group": bson.M{
	// 				"_id": bson.M{
	// 					"date":     "$date",
	// 					"channel":  "$channel",
	// 					"channel1": "$channel1",
	// 				},
	// 				"total_register":    bson.M{"$sum": "$total_register"},
	// 				"total_tourist":     bson.M{"$sum": "$total_tourist"},
	// 				"total_mobile":      bson.M{"$sum": "$total_mobile"},
	// 				"guidance_binding":  bson.M{"$sum": "$guidance_binding"},
	// 				"guidance_get_gold": bson.M{"$sum": "$guidance_get_gold"},
	// 				"tp_number1":        bson.M{"$sum": "$tp_number1"},
	// 				"tp_number2":        bson.M{"$sum": "$tp_number2"},
	// 				"tp_exit_manually":  bson.M{"$sum": "$tp_exit_manually"},
	// 				"gold100":           bson.M{"$sum": "$gold100"},
	// 				"gold200":           bson.M{"$sum": "$gold200"},
	// 				"first_games":       bson.M{"$sum": "$first_games"},
	// 				"second_games":      bson.M{"$sum": "$second_games"},
	// 			},
	// 		},
	// 		{
	// 			"$project": bson.M{
	// 				"_id":               "$_id.date",
	// 				"date":              "$_id.date",
	// 				"channel":           "$_id.channel",
	// 				"channel1":          "$_id.channel1",
	// 				"total_register":    "$total_register",
	// 				"total_tourist":     "$total_tourist",
	// 				"total_mobile":      "$total_mobile",
	// 				"guidance_binding":  "$guidance_binding",
	// 				"guidance_get_gold": "$guidance_get_gold",
	// 				"tp_number1":        "$tp_number1",
	// 				"tp_number2":        "$tp_number2",
	// 				"tp_exit_manually":  "$tp_exit_manually",
	// 				"gold100":           "$gold100",
	// 				"gold200":           "$gold200",
	// 				"first_games":       "$first_games",
	// 				"second_games":      "$second_games",
	// 			},
	// 		},
	// 		{"$sort": bson.M{"date": -1}},
	// 	}
	// } else {
	// 	// 全部
	// 	m := bson.M{"date": id}
	// 	pipeline = []bson.M{
	// 		{"$match": m},
	// 		{
	// 			"$group": bson.M{
	// 				"_id": bson.M{
	// 					"date": "$date",
	// 				},
	// 				"total_register":    bson.M{"$sum": "$total_register"},
	// 				"total_tourist":     bson.M{"$sum": "$total_tourist"},
	// 				"total_mobile":      bson.M{"$sum": "$total_mobile"},
	// 				"guidance_binding":  bson.M{"$sum": "$guidance_binding"},
	// 				"guidance_get_gold": bson.M{"$sum": "$guidance_get_gold"},
	// 				"tp_number1":        bson.M{"$sum": "$tp_number1"},
	// 				"tp_number2":        bson.M{"$sum": "$tp_number2"},
	// 				"tp_exit_manually":  bson.M{"$sum": "$tp_exit_manually"},
	// 				"gold100":           bson.M{"$sum": "$gold100"},
	// 				"gold200":           bson.M{"$sum": "$gold200"},
	// 				"first_games":       bson.M{"$sum": "$first_games"},
	// 				"second_games":      bson.M{"$sum": "$second_games"},
	// 			},
	// 		},
	// 		{
	// 			"$project": bson.M{
	// 				"_id":               "$_id.date",
	// 				"date":              "$_id.date",
	// 				"total_register":    "$total_register",
	// 				"total_tourist":     "$total_tourist",
	// 				"total_mobile":      "$total_mobile",
	// 				"guidance_binding":  "$guidance_binding",
	// 				"guidance_get_gold": "$guidance_get_gold",
	// 				"tp_number1":        "$tp_number1",
	// 				"tp_number2":        "$tp_number2",
	// 				"tp_exit_manually":  "$tp_exit_manually",
	// 				"gold100":           "$gold100",
	// 				"gold200":           "$gold200",
	// 				"first_games":       "$first_games",
	// 				"second_games":      "$second_games",
	// 			},
	// 		},
	// 		{"$sort": bson.M{"date": -1}},
	// 	}
	// }
	// err := PointDatas.Pipe(pipeline).One(&info)
	GetByQ(PointDatas, bson.M{"date": id, "channel": channelname}, info)
	if info.Id == "" {
		return info, errors.New("未查询数据")
	}
	return info, nil
}

func (this *statisticsService) GetLogEventTracksTotal(m bson.M) (int64, error) {
	return int64(Count(LogEventTracks, m)), nil
}
